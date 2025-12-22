package kubevirt

import (
	"context"
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"sort"
	"strings"
	"sync"

	"dario.cat/mergo"
	"go.uber.org/zap"
	"k8c.io/kubermatic/sdk/v2/semver"
	"k8c.io/kubermatic/v2/pkg/defaulting"
	"k8c.io/kubermatic/v2/pkg/validation"
	"k8c.io/kubermatic/v2/pkg/version"
	"k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	"k8c.io/machine-controller/sdk/cloudprovider/kubevirt"
	"k8c.io/machine-controller/sdk/providerconfig"
	"k8c.io/machine-controller/sdk/providerconfig/configvar"
	"k8c.io/machine-controller/sdk/userdata"
	"k8s.io/apimachinery/pkg/runtime"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	kubevirtprovider "k8c.io/kubermatic/v2/pkg/provider/cloud/kubevirt"
	mckubevirtprovider "k8c.io/machine-controller/pkg/cloudprovider/provider/kubevirt"
)

func buildDefaultSeedSettings(datacenterSettings []DatacenterSetting, kkpConfig *kubermaticv1.KubermaticConfiguration, log *zap.SugaredLogger, defaultDatacenterSettings DatacenterSetting) map[string]kubermaticv1.Seed {
	seeds := make(map[string]kubermaticv1.Seed)
	descriptions := make(map[string][]string)
	const maxCombinedSettings = 4 // Limit how many settings to combine into one seed

	// Create a base "default" seed
	defaultDst := kubermaticv1.Datacenter{}
	if defaultDatacenterSettings.modifier != nil {
		defaultDatacenterSettings.modifier(&defaultDst)
	}
	seeds["default"] = kubermaticv1.Seed{
		Spec: kubermaticv1.SeedSpec{
			Datacenters: map[string]kubermaticv1.Datacenter{"default": defaultDst},
		},
	}
	descriptions["default"] = []string{"default"}
	parentKeys := []string{"default"}

	// Separate ungrouped and grouped settings
	groupedSettings := make(map[string][]DatacenterSetting)
	var ungroupedSettings []DatacenterSetting
	for _, setting := range datacenterSettings {
		if setting.group != "" {
			groupedSettings[setting.group] = append(groupedSettings[setting.group], setting)
		} else if setting.name != "default" {
			ungroupedSettings = append(ungroupedSettings, setting)
		}
	}

	// Process individual settings (from groups and ungrouped)
	allIndividualSettings := ungroupedSettings
	for _, group := range groupedSettings {
		allIndividualSettings = append(allIndividualSettings, group...)
	}

	for _, setting := range allIndividualSettings {
		// Attempt to merge this setting into an existing parent seed
		merged := false
		for _, pKey := range parentKeys {
			// Check if the parent already has a setting from the same group
			canMerge := true
			if setting.group != "" {
				for _, desc := range descriptions[pKey] {
					for _, s := range datacenterSettings {
						if s.name == desc && s.group == setting.group {
							canMerge = false
							break
						}
					}
					if !canMerge {
						break
					}
				}
			}

			if canMerge && len(descriptions[pKey]) < maxCombinedSettings {
				// Merge into this parent
				seed := seeds[pKey]
				dc := seed.Spec.Datacenters[pKey]
				if setting.modifier != nil {
					setting.modifier(&dc)
				}
				seed.Spec.Datacenters[pKey] = dc
				seeds[pKey] = seed

				descriptions[pKey] = append(descriptions[pKey], setting.name)
				merged = true
				break
			}
		}

		if !merged {
			// Could not merge, create a new seed for this setting
			newKey := setting.name
			dst := kubermaticv1.Datacenter{}
			if defaultDatacenterSettings.modifier != nil {
				defaultDatacenterSettings.modifier(&dst)
			}
			if setting.modifier != nil {
				setting.modifier(&dst)
			}
			seeds[newKey] = kubermaticv1.Seed{
				Spec: kubermaticv1.SeedSpec{
					Datacenters: map[string]kubermaticv1.Datacenter{newKey: dst},
				},
			}
			descriptions[newKey] = []string{setting.name}
			parentKeys = append(parentKeys, newKey)
		}
	}

	// Rebuild the final map with combined names
	finalSeeds := make(map[string]kubermaticv1.Seed)
	for key, descs := range descriptions {
		combinedName := strings.Join(descs, "-")
		seed := seeds[key]
		// Update the inner datacenter key to the new combined name
		if dc, exists := seed.Spec.Datacenters[key]; exists {
			delete(seed.Spec.Datacenters, key)
			seed.Spec.Datacenters[combinedName] = dc
			finalSeeds[combinedName] = seed
		}
	}

	return finalSeeds
}

// clusterResult is used to pass data from a producer to a consumer.
type clusterResult struct {
	clusterName string
	dedupKey    string
	clusterSpec *kubermaticv1.ClusterSpec
	err         error
}

type clusterJob struct {
	combination    []clusterSpecModifier
	dcKey          string
	seed           kubermaticv1.Seed
	kubeVersion    *version.Version
	log            *zap.SugaredLogger
	rootCtx        context.Context
	opts           *k8cginkgo.Options
	kkpConfig      *kubermaticv1.KubermaticConfiguration
	versionManager *version.Manager
}

func clusterWorker(jobs <-chan clusterJob, results chan<- clusterResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		// Create a descriptive name from the combination.
		var modifierNames []string
		for _, modifier := range job.combination {
			modifierNames = append(modifierNames, modifier.name)
		}
		clusterName := strings.Join(modifierNames, " & ")
		if clusterName == "" {
			clusterName = "default"
		}

		// Create and modify the base spec.
		baseSpec := defaultClusterSettings.Spec.DeepCopy()
		for _, modifier := range job.combination {
			modifier.modify(baseSpec)
		}

		// Create a sanitized spec for deduplication, ignoring certain modifier groups.
		sanitizedSpec := defaultClusterSettings.Spec.DeepCopy()
		ignoredGroups := map[string]bool{
			"update-window": true,
			"oidc":          true,
		}
		for _, modifier := range job.combination {
			if !ignoredGroups[modifier.group] {
				modifier.modify(sanitizedSpec)
			}
		}

		dcName := "kubevirt"

		for k, v := range job.seed.Spec.Datacenters {
			if v.Location == job.dcKey {
				dcName = k
			}

		}

		// Now, continue with the full baseSpec for the actual cluster creation,
		// but use the sanitizedSpec for generating the dedup key.
		clusterSettingSpec := baseSpec

		clusterSettingSpec.Cloud.ProviderName = string(kubermaticv1.KubevirtCloudProvider)
		clusterSettingSpec.Cloud.DatacenterName = dcName
		clusterSettingSpec.Cloud.Kubevirt = &kubermaticv1.KubevirtCloudSpec{
			Kubeconfig: job.opts.Secrets.Kubevirt.Kubeconfig,
		}
		clusterSettingSpec.HumanReadableName = clusterName
		clusterSettingSpec.ContainerRuntime = "containerd"
		clusterSettingSpec.Version = semver.Semver(job.kubeVersion.Version.String())

		currentSeedDatacenter := job.seed.Spec.Datacenters[dcName]
		p, err := kubevirtprovider.NewCloudProvider(&currentSeedDatacenter, nil)
		if err != nil {
			results <- clusterResult{err: fmt.Errorf("failed to create cloud provider for %s: %w", dcName, err)}
			continue
		}

		if err := defaulting.DefaultClusterSpec(job.rootCtx, clusterSettingSpec, nil, &job.seed, job.kkpConfig, p); err != nil {
			results <- clusterResult{err: fmt.Errorf("failed to default cluster spec %s: %w", clusterName, err)}
			continue
		}

		if valErrs := validation.ValidateClusterSpec(clusterSettingSpec, &currentSeedDatacenter, nil, job.versionManager, &clusterSettingSpec.Version, nil); len(valErrs) != 0 {
			job.log.Infof("Skipping invalid cluster spec %q: %v", clusterName, valErrs.ToAggregate())
			results <- clusterResult{err: nil} // Skippable error
			continue
		}

		// Generate a stable hash of the sanitized spec for true deduplication.
		specBytes, err := json.Marshal(sanitizedSpec)
		if err != nil {
			results <- clusterResult{err: fmt.Errorf("failed to marshal spec for hashing: %w", err)}
			continue
		}
		dedupKey := "ginkgo-" + fmt.Sprintf("%x", sha256.Sum256(specBytes))[:25]

		results <- clusterResult{
			clusterName: clusterName,
			dedupKey:    dedupKey,
			clusterSpec: clusterSettingSpec,
			err:         nil,
		}
	}
}

func buildNewClusters(rootCtx context.Context, versions []*version.Version, clusterModifiers []clusterSpecModifier, defaultSeedSettings map[string]kubermaticv1.Seed, seed *kubermaticv1.Seed, opts *k8cginkgo.Options, kkpConfig *kubermaticv1.KubermaticConfiguration, log *zap.SugaredLogger, versionManager *version.Manager, file *os.File) (map[string]*kubermaticv1.ClusterSpec, map[string][]string) {
	finalClusters := make(map[string]*kubermaticv1.ClusterSpec)
	finalClusterDescriptions := make(map[string][]string)
	var finalMu sync.Mutex

	// Group modifiers by their group name.
	groupedModifiers := make(map[string][]clusterSpecModifier)
	for _, m := range clusterModifiers {
		groupedModifiers[m.group] = append(groupedModifiers[m.group], m)
	}

	var groupNames []string
	for name := range groupedModifiers {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	var longestKey string
	maxLen := 0
	for k, s := range groupedModifiers {
		if len(s) > maxLen {
			maxLen = len(s)
			longestKey = k
		}
	}
	// longestKey is the key of the longest slice, maxLen is its length
	// Combine modifiers and descriptions by index
	combinedModifiers := make([][]clusterSpecModifier, len(groupedModifiers[longestKey]))
	combinedDescriptions := make([][]string, len(groupedModifiers[longestKey]))
	for _, modifiers := range groupedModifiers {
		for idx, modifier := range modifiers {
			combinedModifiers[idx] = append(combinedModifiers[idx], modifier)
			combinedDescriptions[idx] = append(combinedDescriptions[idx], modifier.name)
		}
	}

	const numWorkers = 100
	jobs := make(chan clusterJob)
	results := make(chan clusterResult)

	// Start workers
	var workerWg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		workerWg.Add(1)
		go clusterWorker(jobs, results, &workerWg)
	}

	// Start a goroutine to generate combinations and send all jobs
	go func(seed *kubermaticv1.Seed) {
		defer close(jobs)

		// Generate jobs for each combined set of modifiers
		for _, mods := range combinedModifiers {
			for _, kubeVersion := range versions {
				for dcKey := range defaultSeedSettings {
					// Copy to ensure a fresh slice
					jobCombination := make([]clusterSpecModifier, len(mods))
					copy(jobCombination, mods)
					jobs <- clusterJob{
						combination:    jobCombination,
						dcKey:          dcKey,
						seed:           *seed,
						kubeVersion:    kubeVersion,
						log:            log,
						rootCtx:        rootCtx,
						opts:           opts,
						kkpConfig:      kkpConfig,
						versionManager: versionManager,
					}
				}
			}
		}
	}(seed)

	// Start a goroutine to close the results channel when all workers are done
	go func() {
		workerWg.Wait()
		close(results)
	}()

	// Collect results
	for result := range results {
		if result.err != nil {
			log.Errorw("Cluster generation worker failed", "error", result.err)
			continue
		}
		if result.clusterSpec == nil {
			continue // Worker skipped an invalid spec
		}

		// Use finalMu for safe concurrent access to the final maps
		finalMu.Lock()
		if _, exists := finalClusters[result.dedupKey]; exists {
			// This spec is a duplicate of one we've already stored.
			// Just add the descriptive name and discard the new spec object, saving memory.
			// Check if we already have this description to avoid duplicates in the output.
			isUnique := true
			for _, name := range finalClusterDescriptions[result.dedupKey] {
				if name == result.clusterName {
					isUnique = false
					break
				}
			}
			if isUnique {
				finalClusterDescriptions[result.dedupKey] = append(finalClusterDescriptions[result.dedupKey], result.clusterName)
			}
		} else {
			// This is a new, unique spec. Store it.
			finalClusters[result.dedupKey] = result.clusterSpec
			finalClusterDescriptions[result.dedupKey] = []string{result.clusterName}
		}
		finalMu.Unlock()
	}

	// Post-process to remove duplicates from descriptions
	for key, descs := range finalClusterDescriptions {
		// Combine all parts into a single string, using a consistent separator.
		allPartsStr := strings.Join(descs, " and ")

		// Normalize separators (e.g., handle " & ", ", ") and split into individual components.
		normalizedStr := strings.ReplaceAll(allPartsStr, " & ", " and ")
		normalizedStr = strings.ReplaceAll(normalizedStr, ", ", " and ")
		parts := strings.Split(normalizedStr, " and ")

		// Filter for unique parts.
		uniqueParts := make(map[string]bool)
		var finalParts []string
		for _, part := range parts {
			trimmedPart := strings.TrimSpace(part)
			if trimmedPart != "" && !uniqueParts[trimmedPart] {
				uniqueParts[trimmedPart] = true
				finalParts = append(finalParts, trimmedPart)
			}
		}
		finalClusterDescriptions[key] = finalParts
	}

	// Output final cluster descriptions.
	fmt.Fprintf(file, "\nFINAL CLUSTER DESCRIPTIONS (with combined cluster names):\n")
	for key, descs := range finalClusterDescriptions {
		fmt.Fprintf(file, "Cluster (dedup key: %s...):\n  Generated from combinations: %v\n\n", key[:12], descs)
	}

	return finalClusters, finalClusterDescriptions
}

// scenarioResult is used to pass data from a producer to a consumer.
type scenarioResult struct {
	clusterKey  string
	machineName string
	dedupKey    string
	machineSpec v1alpha1.MachineSpec
	err         error
}

type scenarioJob struct {
	combination           []machineSpecModifier
	clusterKey            string
	version               semver.Semver
	log                   *zap.SugaredLogger
	rootCtx               context.Context
	resolver              *configvar.Resolver
	opts                  *k8cginkgo.Options
	defaultKubevirtConfig kubevirt.RawConfig
}

func scenarioWorker(jobs <-chan scenarioJob, results chan<- scenarioResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		// Create a descriptive name from the combination.
		var modifierNames []string
		for _, modifier := range job.combination {
			modifierNames = append(modifierNames, modifier.name)
		}
		machineName := strings.Join(modifierNames, " & ")
		if machineName == "" {
			machineName = "default"
		}

		// Create a sanitized config for deduplication, ignoring certain modifier groups.
		sanitizedRawConfig := job.defaultKubevirtConfig
		ignoredGroups := map[string]bool{
			"affinity": true,
		}
		for _, modifier := range job.combination {
			if !ignoredGroups[modifier.group] {
				modifier.modify(&sanitizedRawConfig)
			}
		}
		sanitizedRawConfig.Auth.Kubeconfig.Value = b64.StdEncoding.EncodeToString([]byte(job.opts.Secrets.Kubevirt.Kubeconfig))

		// Create a full config for the actual machine spec.
		rawConfig := job.defaultKubevirtConfig
		for _, modifier := range job.combination {
			modifier.modify(&rawConfig)
		}
		rawConfig.Auth.Kubeconfig.Value = b64.StdEncoding.EncodeToString([]byte(job.opts.Secrets.Kubevirt.Kubeconfig))

		// Generate the dedup key from the sanitized config AND the machine name to ensure uniqueness for the "default" case.
		sanitizedSpecBytes, err := json.Marshal(sanitizedRawConfig)
		if err != nil {
			results <- scenarioResult{err: fmt.Errorf("failed to marshal sanitized spec for hashing: %w", err)}
			continue
		}
		h := sha256.New()
		h.Write(sanitizedSpecBytes)
		h.Write([]byte(machineName))
		dedupKey := fmt.Sprintf("%x", h.Sum(nil))

		// Create a base machine spec for this group.
		machine, err := getDefaultMachineSpec()
		if err != nil {
			job.log.Errorw("Failed to get default machine spec", "machine", machineName, zap.Error(err))
			results <- scenarioResult{err: err}
			continue
		}

		// Unmarshal, modify with the FULL config, and re-marshal the provider spec.
		var pconfig providerconfig.Config
		if err := json.Unmarshal(machine.ProviderSpec.Value.Raw, &pconfig); err != nil {
			err = fmt.Errorf("failed to unmarshal provider config: %w", err)
			job.log.Errorw(err.Error(), "machine", machineName)
			results <- scenarioResult{err: err}
			continue
		}

		pconfig.CloudProviderSpec.Raw = toJSON(rawConfig)
		osspec, err := userdata.DefaultOperatingSystemSpec(providerconfig.OperatingSystemUbuntu, runtime.RawExtension{})
		if err != nil {
			job.log.Errorw("Failed to get default OS spec", "machine", machineName, zap.Error(err))
			results <- scenarioResult{err: err}
			continue
		}
		pconfig.CloudProvider = providerconfig.CloudProviderKubeVirt
		pconfig.OperatingSystemSpec = osspec
		pconfig.OperatingSystem = providerconfig.OperatingSystemUbuntu
		reencodedPConfig, err := json.Marshal(pconfig)
		if err != nil {
			err = fmt.Errorf("failed to re-marshal provider config: %w", err)
			job.log.Errorw(err.Error(), "machine", machineName)
			results <- scenarioResult{err: err}
			continue
		}
		machine.ProviderSpec.Value.Raw = reencodedPConfig
		machine.Versions.Kubelet = job.version.String()
		p := mckubevirtprovider.New(job.resolver)
		machineSpec, err := p.AddDefaults(job.log, *machine)
		if err != nil {
			results <- scenarioResult{err: fmt.Errorf("failed to add defaults to machine: %w", err)}
			continue
		}

		if err := p.Validate(job.rootCtx, job.log, machineSpec); err != nil {
			job.log.Infof("Skipping invalid machine spec for %q: %v", machineName, err)
			results <- scenarioResult{err: nil} // Skippable
			continue
		}

		results <- scenarioResult{
			clusterKey:  job.clusterKey,
			machineName: machineName,
			dedupKey:    dedupKey,
			machineSpec: machineSpec,
			err:         nil,
		}
	}
}

func buildNewScenarios(machineModifiers []machineSpecModifier, newClusters map[string]*kubermaticv1.ClusterSpec, opts *k8cginkgo.Options, log *zap.SugaredLogger, defaultKubevirtConfig kubevirt.RawConfig, resolver *configvar.Resolver, file *os.File, rootCtx context.Context) (map[string]map[string]v1alpha1.MachineSpec, map[string]map[string][]string) {
	finalScenarios := make(map[string]map[string]v1alpha1.MachineSpec)
	finalMachineDescriptions := make(map[string]map[string][]string)
	var finalMu sync.Mutex

	// Group modifiers by their group name.
	groupedModifiers := make(map[string][]machineSpecModifier)
	for _, m := range machineModifiers {
		groupedModifiers[m.group] = append(groupedModifiers[m.group], m)
	}

	var groupNames []string
	for name := range groupedModifiers {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	var longestKey string
	maxLen := 0
	for k, s := range groupedModifiers {
		if len(s) > maxLen {
			maxLen = len(s)
			longestKey = k
		}
	}
	// longestKey is the key of the longest slice, maxLen is its length

	// Combine modifiers and descriptions by index
	combinedModifiers := make([][]machineSpecModifier, len(groupedModifiers[longestKey]))
	combinedDescriptions := make([][]string, len(groupedModifiers[longestKey]))
	for _, modifiers := range groupedModifiers {
		for idx, modifier := range modifiers {
			combinedModifiers[idx] = append(combinedModifiers[idx], modifier)
			combinedDescriptions[idx] = append(combinedDescriptions[idx], modifier.name)
		}
	}

	log.Infof("Starting scenario generation...")

	const numWorkers = 100
	jobs := make(chan scenarioJob)
	results := make(chan scenarioResult)

	// Start workers
	var workerWg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		workerWg.Add(1)
		go scenarioWorker(jobs, results, &workerWg)
	}

	// Start a goroutine to generate combinations and send all jobs
	go func() {
		defer close(jobs)
		// Generate jobs for each combined set of modifiers
		for _, mods := range combinedModifiers {
			for clusterKey, clusterSpec := range newClusters {
				// Copy to ensure a fresh slice
				jobCombination := make([]machineSpecModifier, len(mods))
				copy(jobCombination, mods)
				jobs <- scenarioJob{
					combination:           jobCombination,
					clusterKey:            clusterKey,
					version:               clusterSpec.Version,
					log:                   log,
					rootCtx:               rootCtx,
					resolver:              resolver,
					opts:                  opts,
					defaultKubevirtConfig: defaultKubevirtConfig,
				}
			}
		}
	}()

	// Start a goroutine to close the results channel when all workers are done
	go func() {
		workerWg.Wait()
		close(results)
	}()

	// Collect results
	processedCount := 0
	// Buffer for scenarios to be written to the file periodically.
	scenarioBuffer := []scenarioResult{}
	var bufferMu sync.Mutex

	for result := range results {
		processedCount++
		if processedCount%100 == 0 {
			log.Infof("Generated %d scenarios...", processedCount)

			finalMu.Lock()
			uniqueCount := 0
			for _, clusterScenarios := range finalScenarios {
				uniqueCount += len(clusterScenarios)
			}
			finalMu.Unlock()
			log.Infof("Resulting in %d unique scenarios so far.", uniqueCount)

			// flushBuffer()
		}

		if result.err != nil {
			log.Errorw("Scenario generation worker failed", "error", result.err)
			continue
		}
		if result.machineSpec.ProviderSpec.Value == nil {
			continue // Worker skipped an invalid spec
		}

		bufferMu.Lock()
		scenarioBuffer = append(scenarioBuffer, result)
		bufferMu.Unlock()

		finalMu.Lock()
		// Ensure nested maps are initialized
		if _, ok := finalScenarios[result.clusterKey]; !ok {
			finalScenarios[result.clusterKey] = make(map[string]v1alpha1.MachineSpec)
			finalMachineDescriptions[result.clusterKey] = make(map[string][]string)
		}

		if existing, exists := finalScenarios[result.clusterKey][result.dedupKey]; exists {
			merged := existing
			if err := mergo.Merge(&merged, result.machineSpec, mergo.WithOverride); err == nil {
				finalScenarios[result.clusterKey][result.dedupKey] = merged
				finalMachineDescriptions[result.clusterKey][result.dedupKey] = append(finalMachineDescriptions[result.clusterKey][result.dedupKey], result.machineName)
			}
		} else {
			finalScenarios[result.clusterKey][result.dedupKey] = result.machineSpec
			finalMachineDescriptions[result.clusterKey][result.dedupKey] = []string{result.machineName}
		}
		finalMu.Unlock()
	}

	log.Infof("Finished generating a total of %d scenarios.", processedCount)
	return finalScenarios, finalMachineDescriptions
}

func postProcessScenarios(
	newScenarios map[string]map[string]v1alpha1.MachineSpec,
	finalMachineDescriptions map[string]map[string][]string,
	file *os.File,
	log *zap.SugaredLogger,
) map[string]map[string][]string {
	processedDescriptions := make(map[string]map[string][]string)
	const maxCombinedScenarios = 3

	fmt.Fprintf(file, "\n--- POST-PROCESSED SCENARIO DESCRIPTIONS ---\n")

	for clusterKey, scenarios := range finalMachineDescriptions {
		var defaultKey string
		otherKeys := []string{}
		log.Infof("Generated Scenarios: %q\n", strings.Join(slices.Collect(maps.Keys(newScenarios[clusterKey])), ", "))
		log.Infof("Scenario Descriptions: %q\n", strings.Join(slices.Collect(maps.Keys(scenarios)), ", "))
		for key, descs := range scenarios {
			if len(descs) == 1 && descs[0] == "default" {
				defaultKey = key
			} else {
				otherKeys = append(otherKeys, key)
			}
		}

		if defaultKey == "" {
			processedDescriptions[clusterKey] = scenarios
			fmt.Fprintf(file, "Cluster: %s (no default scenario found, skipping post-processing)\n", clusterKey)
			for key, descs := range scenarios {
				fmt.Fprintf(file, "  Scenario Key: %s\n    Combined Scenario Names: %v\n\n", key, descs)
			}
			continue
		}

		// Initialize the processed map with the default scenario.
		processedScenarios := make(map[string][]string)
		processedScenarios[defaultKey] = scenarios[defaultKey]
		parents := []string{defaultKey}

		// Keep track of which keys have been merged away.
		mergedKeys := make(map[string]bool)

		for _, childKey := range otherKeys {
			wasMerged := false
			for _, parentKey := range parents {
				if len(processedScenarios[parentKey]) < maxCombinedScenarios {
					processedScenarios[parentKey] = append(processedScenarios[parentKey], scenarios[childKey]...)
					mergedKeys[childKey] = true
					wasMerged = true
					break
				}
			}

			// If it couldn't be merged, it becomes a new parent, if it's not already merged.
			if !wasMerged && !mergedKeys[childKey] {
				parents = append(parents, childKey)
				processedScenarios[childKey] = scenarios[childKey]
			}
		}

		// Final map should only contain the parents (which now contain the children).
		finalProcessedScenarios := make(map[string][]string)
		for _, parentKey := range parents {
			finalProcessedScenarios[parentKey] = processedScenarios[parentKey]
		}

		processedDescriptions[clusterKey] = finalProcessedScenarios

		fmt.Fprintf(file, "Cluster: %s\n", clusterKey)
		for key, descs := range finalProcessedScenarios {
			// Combine all parts into a single string, using a consistent separator.
			allPartsStr := strings.Join(descs, " and ")

			// Normalize separators (e.g., handle " & ", ", ") and split into individual components.
			normalizedStr := strings.ReplaceAll(allPartsStr, " & ", " and ")
			normalizedStr = strings.ReplaceAll(normalizedStr, ", ", " and ")
			parts := strings.Split(normalizedStr, " and ")

			// Filter for unique parts.
			uniqueParts := make(map[string]bool)
			var finalParts []string
			for _, part := range parts {
				trimmedPart := strings.TrimSpace(part)
				if trimmedPart != "" && !uniqueParts[trimmedPart] {
					uniqueParts[trimmedPart] = true
					finalParts = append(finalParts, trimmedPart)
				}
			}

			// Update the descriptions in the map with the cleaned-up version.
			finalProcessedScenarios[key] = finalParts

			fmt.Fprintf(file, "  Scenario Key: %s\n    Combined Scenario Names: %v\n\n", key, finalParts)
		}
	}

	return processedDescriptions
}
