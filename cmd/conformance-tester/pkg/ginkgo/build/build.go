package build

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"os"
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
	"k8c.io/machine-controller/sdk/providerconfig"
	"k8c.io/machine-controller/sdk/providerconfig/configvar"
	"k8c.io/machine-controller/sdk/userdata"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/options"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/settings"
)

func toJSON(i any) []byte {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return b
}

func buildDefaultSeedSettings(datacenterSettings []settings.DatacenterSetting, kkpConfig *kubermaticv1.KubermaticConfiguration, log *zap.SugaredLogger, defaultDatacenterSettings settings.DatacenterSetting, excludedDatacenterDescriptions, includedDatacenterDescriptions []string) map[string]kubermaticv1.Seed {
	seeds := make(map[string]kubermaticv1.Seed)
	const maxCombinedSettings = 6

	// Build a set for fast lookup
	excludedDescSet := make(map[string]struct{}, len(excludedDatacenterDescriptions))
	for _, desc := range excludedDatacenterDescriptions {
		excludedDescSet[desc] = struct{}{}
	}

	includedDescSet := make(map[string]struct{}, len(includedDatacenterDescriptions))
	for _, desc := range includedDatacenterDescriptions {
		includedDescSet[desc] = struct{}{}
	}

	// Separate settings into included and excluded
	var included, excluded []settings.DatacenterSetting
	for _, setting := range datacenterSettings {
		if len(includedDescSet) > 0 {
			if _, ok := includedDescSet[setting.Name]; ok {
				included = append(included, setting)
			} else {
				excluded = append(excluded, setting)
			}
		} else {
			if _, ok := excludedDescSet[setting.Name]; !ok {
				included = append(included, setting)
			} else {
				excluded = append(excluded, setting)
			}
		}
	}

	// Helper to group/combine a set of settings
	combineSettings := func(datacenterSettings []settings.DatacenterSetting, groupLabel string) map[string]kubermaticv1.Seed {
		groupedSettings := make(map[string][]settings.DatacenterSetting)
		var ungroupedSettings []settings.DatacenterSetting
		for _, setting := range datacenterSettings {
			if setting.Group != "" {
				groupedSettings[setting.Group] = append(groupedSettings[setting.Group], setting)
			} else if setting.Name != "default" {
				ungroupedSettings = append(ungroupedSettings, setting)
			}
		}
		// Process individual settings (from groups and ungrouped)
		allIndividualSettings := ungroupedSettings
		for _, group := range groupedSettings {
			allIndividualSettings = append(allIndividualSettings, group...)
		}
		descriptions := make(map[string][]string)
		parentKeys := []string{"default"}
		seeds := make(map[string]kubermaticv1.Seed)
		// Create a base "default" seed
		defaultDst := kubermaticv1.Datacenter{}
		if defaultDatacenterSettings.Modifier != nil {
			defaultDatacenterSettings.Modifier(&defaultDst)
		}
		seeds["default"] = kubermaticv1.Seed{
			Spec: kubermaticv1.SeedSpec{
				Datacenters: map[string]kubermaticv1.Datacenter{"default": defaultDst},
			},
		}
		descriptions["default"] = []string{"default"}
		for _, setting := range allIndividualSettings {
			merged := false
			for _, pKey := range parentKeys {
				canMerge := true
				if setting.Group != "" {
					for _, desc := range descriptions[pKey] {
						for _, s := range datacenterSettings {
							if s.Name == desc && s.Group == setting.Group {
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
					seed := seeds[pKey]
					dc := seed.Spec.Datacenters[pKey]
					if setting.Modifier != nil {
						setting.Modifier(&dc)
					}
					seed.Spec.Datacenters[pKey] = dc
					seeds[pKey] = seed
					descriptions[pKey] = append(descriptions[pKey], setting.Name)
					merged = true
					break
				}
			}
			if !merged {
				newKey := groupLabel + "-" + setting.Name
				dst := kubermaticv1.Datacenter{}
				if defaultDatacenterSettings.Modifier != nil {
					defaultDatacenterSettings.Modifier(&dst)
				}
				if setting.Modifier != nil {
					setting.Modifier(&dst)
				}
				seeds[newKey] = kubermaticv1.Seed{
					Spec: kubermaticv1.SeedSpec{
						Datacenters: map[string]kubermaticv1.Datacenter{newKey: dst},
					},
				}
				descriptions[newKey] = []string{setting.Name}
				parentKeys = append(parentKeys, newKey)
			}
		}
		// Rebuild the final map with combined names
		finalSeeds := make(map[string]kubermaticv1.Seed)
		for key, descs := range descriptions {
			combinedName := strings.Join(descs, " and ")
			seed := seeds[key]
			if dc, exists := seed.Spec.Datacenters[key]; exists {
				delete(seed.Spec.Datacenters, key)
				seed.Spec.Datacenters[combinedName] = dc
				finalSeeds[combinedName] = seed
			}
		}
		return finalSeeds
	}

	includedSeeds := combineSettings(included, "included")
	excludedSeeds := combineSettings(excluded, "excluded")

	// Merge both maps
	maps.Copy(seeds, includedSeeds)
	maps.Copy(seeds, excludedSeeds)

	return seeds
}

func clusterWorker(jobs <-chan clusterJob, results chan<- clusterResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		// Create a descriptive name from the combination.
		var modifierNames []string
		for _, modifier := range job.combination {
			modifierNames = append(modifierNames, modifier.Name)
		}
		clusterName := strings.Join(modifierNames, " & ")
		if clusterName == "" {
			clusterName = "default"
		}

		// Create and modify the base spec.
		baseSpec := k8cginkgo.DefaultClusterSettings.Spec.DeepCopy()
		for _, modifier := range job.combination {
			modifier.Modify(baseSpec)
		}

		// Create a sanitized spec for deduplication, ignoring certain modifier groups.
		sanitizedSpec := k8cginkgo.DefaultClusterSettings.Spec.DeepCopy()
		ignoredGroups := map[string]bool{
			"update-window": true,
			"oidc":          true,
		}
		for _, modifier := range job.combination {
			if !ignoredGroups[modifier.Group] {
				modifier.Modify(sanitizedSpec)
			}
		}

		dcName := ""

		for k, v := range job.seed.Spec.Datacenters {
			if v.Location == job.dcKey {
				dcName = k
			}

		}

		// Now, continue with the full baseSpec for the actual cluster creation,
		// but use the sanitizedSpec for generating the dedup key.
		clusterSettingSpec := baseSpec
		currentSeedDatacenter := job.seed.Spec.Datacenters[dcName]
		p, c, err := getClusterProvider(job.providerConfig.CloudProvider, dcName, &currentSeedDatacenter, job.opts.Secrets)
		if err != nil {
			results <- clusterResult{err: fmt.Errorf("failed to get cluster provider for %s: %w", dcName, err)}
			continue
		}
		clusterSettingSpec.Cloud = c
		clusterSettingSpec.HumanReadableName = clusterName
		clusterSettingSpec.ContainerRuntime = "containerd"
		clusterSettingSpec.Version = semver.Semver(job.kubeVersion.Version.String())

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
		f := fmt.Sprintf("%x", sha256.Sum256(specBytes))[:6]
		dedupKey := fmt.Sprintf("k8c-%s-%s", f, strings.ReplaceAll(job.kubeVersion.Version.String(), ".", "-"))

		results <- clusterResult{
			clusterName: clusterName,
			dedupKey:    dedupKey,
			clusterSpec: clusterSettingSpec,
			err:         nil,
		}
	}
}

func buildNewClusters(
	rootCtx context.Context,
	versions []*version.Version,
	clusterModifiers []settings.ClusterSpecModifier,
	defaultSeedSettings map[string]kubermaticv1.Seed,
	seed *kubermaticv1.Seed,
	opts *options.Options,
	kkpConfig *kubermaticv1.KubermaticConfiguration,
	log *zap.SugaredLogger,
	versionManager *version.Manager,
	file *os.File,
	providerConfig *providerconfig.Config,
	clusterDescriptions []string, // NEW: descriptions to include
	includedDescriptions []string, // NEW: descriptions to include
) (map[string]*kubermaticv1.ClusterSpec, map[string][]string) {
	finalClusters := make(map[string]*kubermaticv1.ClusterSpec)
	finalClusterDescriptions := make(map[string][]string)

	// Build a set for fast lookup
	descSet := make(map[string]struct{}, len(clusterDescriptions))
	for _, desc := range clusterDescriptions {
		descSet[desc] = struct{}{}
	}

	// Build a set for fast lookup
	includedDescSet := make(map[string]struct{}, len(includedDescriptions))
	for _, desc := range includedDescriptions {
		includedDescSet[desc] = struct{}{}
	}

	// Separate modifiers into included and excluded
	var included, excluded []settings.ClusterSpecModifier
	for _, m := range clusterModifiers {
		if len(includedDescSet) > 0 {
			if _, ok := includedDescSet[m.Name]; ok {
				included = append(included, m)
			} else {
				excluded = append(excluded, m)
			}
		} else {
			if _, ok := descSet[m.Name]; !ok {
				included = append(included, m)
			} else {
				excluded = append(excluded, m)
			}
		}
	}

	// Helper to group/combine a set of modifiers
	combineModifiers := func(modifiers []settings.ClusterSpecModifier, groupLabel string) (map[string]*kubermaticv1.ClusterSpec, map[string][]string) {
		// Group modifiers by their group name.
		groupedModifiers := make(map[string][]settings.ClusterSpecModifier)
		for _, m := range modifiers {
			groupedModifiers[m.Group] = append(groupedModifiers[m.Group], m)
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
		if maxLen == 0 {
			return map[string]*kubermaticv1.ClusterSpec{}, map[string][]string{}
		}
		// Combine modifiers and descriptions by index
		combinedModifiers := make([][]settings.ClusterSpecModifier, len(groupedModifiers[longestKey]))
		combinedDescriptions := make([][]string, len(groupedModifiers[longestKey]))
		for _, modifiers := range groupedModifiers {
			for idx, modifier := range modifiers {
				combinedModifiers[idx] = append(combinedModifiers[idx], modifier)
				combinedDescriptions[idx] = append(combinedDescriptions[idx], modifier.Name)
			}
		}

		const numWorkers = 100
		jobs := make(chan clusterJob)
		results := make(chan clusterResult)

		// Start workers
		var workerWg sync.WaitGroup
		for range numWorkers {
			workerWg.Add(1)
			go clusterWorker(jobs, results, &workerWg)
		}

		// Start a goroutine to generate combinations and send all jobs
		go func(seed *kubermaticv1.Seed) {
			defer close(jobs)
			for _, mods := range combinedModifiers {
				for _, kubeVersion := range versions {
					for dcKey := range defaultSeedSettings {
						jobCombination := make([]settings.ClusterSpecModifier, len(mods))
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
							providerConfig: providerConfig,
						}
					}
				}
			}
		}(seed)

		go func() {
			workerWg.Wait()
			close(results)
		}()

		localClusters := make(map[string]*kubermaticv1.ClusterSpec)
		localClusterDescriptions := make(map[string][]string)
		for result := range results {
			if result.err != nil {
				log.Errorw("Cluster generation worker failed", "error", result.err)
				continue
			}
			if result.clusterSpec == nil {
				continue
			}
			if _, exists := localClusters[result.dedupKey]; exists {
				isUnique := true
				for _, name := range localClusterDescriptions[result.dedupKey] {
					if name == result.clusterName {
						isUnique = false
						break
					}
				}
				if isUnique {
					localClusterDescriptions[result.dedupKey] = append(localClusterDescriptions[result.dedupKey], result.clusterName)
				}
			} else {
				localClusters[result.dedupKey] = result.clusterSpec
				localClusterDescriptions[result.dedupKey] = []string{result.clusterName}
			}
		}
		// Post-process to remove duplicates from descriptions
		for key, descs := range localClusterDescriptions {
			allPartsStr := strings.Join(descs, " and ")
			normalizedStr := strings.ReplaceAll(allPartsStr, " & ", " and ")
			normalizedStr = strings.ReplaceAll(normalizedStr, ", ", " and ")
			parts := strings.Split(normalizedStr, " and ")
			uniqueParts := make(map[string]bool)
			var finalParts []string
			for _, part := range parts {
				trimmedPart := strings.TrimSpace(part)
				if trimmedPart != "" && !uniqueParts[trimmedPart] {
					uniqueParts[trimmedPart] = true
					finalParts = append(finalParts, trimmedPart)
				}
			}
			localClusterDescriptions[key] = finalParts
		}
		return localClusters, localClusterDescriptions
	}

	includedClusters, combinedIncludedDescriptions := combineModifiers(included, "included")
	excludedClusters, excludedDescriptions := combineModifiers(excluded, "excluded")

	// Merge both maps
	maps.Copy(finalClusters, includedClusters)
	maps.Copy(finalClusters, excludedClusters)
	maps.Copy(finalClusterDescriptions, combinedIncludedDescriptions)
	maps.Copy(finalClusterDescriptions, excludedDescriptions)
	return finalClusters, finalClusterDescriptions
}

func scenarioWorker(jobs <-chan scenarioJob, results chan<- scenarioResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		// Create a descriptive name from the combination.
		var modifierNames []string
		for _, modifier := range job.combination {
			modifierNames = append(modifierNames, modifier.Name)
		}
		machineName := strings.Join(modifierNames, " & ")
		if machineName == "" {
			machineName = "default"
		}

		// Create a sanitized config for deduplication, ignoring certain modifier groups.
		sanitizedRawConfig := job.providerConfig.CloudProviderSpec.Raw
		ignoredGroups := map[string]bool{
			"affinity": true,
		}
		ps, err := getProviderSpec(job.log, job.opts.Secrets, job.providerConfig.CloudProvider)
		if err != nil {
			results <- scenarioResult{err: fmt.Errorf("failed to get provider spec for %s: %w", machineName, err)}
			continue
		}
		for _, modifier := range job.combination {
			if !ignoredGroups[modifier.Group] {
				modifier.Modify(ps)
			}
		}

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
		machine, err := getDefaultMachineSpec(job.rootCtx, job.log, job.providerConfig, job.opts.Secrets)
		if err != nil {
			job.log.Errorw("Failed to get default machine spec", "machine", machineName, zap.Error(err))
			results <- scenarioResult{err: err}
			continue
		}

		psb, err := json.Marshal(ps)
		if err != nil {
			job.log.Errorw("Failed to marshal provider spec", "machine", machineName, zap.Error(err))
			results <- scenarioResult{err: err}
			continue
		}
		pconfig := providerconfig.Config{
			CloudProvider: job.providerConfig.CloudProvider,
			CloudProviderSpec: runtime.RawExtension{
				Raw: psb,
			},
		}

		osspec, err := userdata.DefaultOperatingSystemSpec(job.distribution, runtime.RawExtension{})
		if err != nil {
			job.log.Errorw("Failed to get default OS spec", "machine", machineName, zap.Error(err))
			results <- scenarioResult{err: err}
			continue
		}
		pconfig.CloudProvider = job.providerConfig.CloudProvider
		pconfig.OperatingSystemSpec = osspec
		pconfig.OperatingSystem = job.distribution
		reencodedPConfig, err := json.Marshal(pconfig)
		if err != nil {
			err = fmt.Errorf("failed to re-marshal provider config: %w", err)
			job.log.Errorw(err.Error(), "machine", machineName)
			results <- scenarioResult{err: err}
			continue
		}
		machine.ProviderSpec.Value.Raw = reencodedPConfig
		machine.Versions.Kubelet = job.version.String()

		p := getProvider(job.providerConfig.CloudProvider, job.resolver)
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

func buildNewScenarios(
	machineModifiers []settings.MachineSpecModifier[any],
	newClusters map[string]*kubermaticv1.ClusterSpec,
	opts *options.Options,
	log *zap.SugaredLogger,
	providerConfig *providerconfig.Config,
	resolver *configvar.Resolver,
	file *os.File,
	rootCtx context.Context,
	operatingSystems sets.Set[string],
	machineDescriptions []string, // NEW: descriptions to exlclude
	includedMachineDescription []string, // NEW: descriptions to include

) (map[string]map[string]v1alpha1.MachineSpec, map[string]map[string][]string, iter.Seq[string]) {
	finalScenarios := make(map[string]map[string]v1alpha1.MachineSpec)
	finalMachineDescriptions := make(map[string]map[string][]string)

	// Build a set for fast lookup
	descSet := make(map[string]struct{}, len(machineDescriptions))
	for _, desc := range machineDescriptions {
		descSet[desc] = struct{}{}
	}

	// Build a set for fast lookup
	includedDescSet := make(map[string]struct{}, len(includedMachineDescription))
	for _, desc := range includedMachineDescription {
		includedDescSet[desc] = struct{}{}
	}

	// Separate modifiers into included and excluded
	var included, excluded []settings.MachineSpecModifier[any]
	for _, m := range machineModifiers {
		if len(includedDescSet) > 0 {
			if _, ok := includedDescSet[m.Name]; ok {
				included = append(included, m)
			} else {
				excluded = append(excluded, m)
			}
		} else {
			if _, ok := descSet[m.Name]; !ok {
				included = append(included, m)
			} else {
				excluded = append(excluded, m)
			}
		}
	}

	// Helper to group/combine a set of modifiers
	combineModifiers := func(modifiers []settings.MachineSpecModifier[any], groupLabel string) (map[string]map[string]v1alpha1.MachineSpec, map[string]map[string][]string) {
		localScenarios := make(map[string]map[string]v1alpha1.MachineSpec)
		localMachineDescriptions := make(map[string]map[string][]string)
		// Group modifiers by their group name.
		groupedModifiers := make(map[string][]settings.MachineSpecModifier[any])
		for _, m := range modifiers {
			groupedModifiers[m.Group] = append(groupedModifiers[m.Group], m)
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
		if maxLen == 0 {
			return map[string]map[string]v1alpha1.MachineSpec{}, map[string]map[string][]string{}
		}
		// Combine modifiers and descriptions by index
		combinedModifiers := make([][]settings.MachineSpecModifier[any], len(groupedModifiers[longestKey]))
		combinedDescriptions := make([][]string, len(groupedModifiers[longestKey]))
		for _, modifiers := range groupedModifiers {
			for idx, modifier := range modifiers {
				combinedModifiers[idx] = append(combinedModifiers[idx], modifier)
				combinedDescriptions[idx] = append(combinedDescriptions[idx], modifier.Name)
			}
		}

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
			for _, mods := range combinedModifiers {
				for clusterKey, clusterSpec := range newClusters {
					for distro, _ := range operatingSystems {
						jobCombination := make([]settings.MachineSpecModifier[any], len(mods))
						copy(jobCombination, mods)
						d := providerconfig.OperatingSystem(distro)
						jobs <- scenarioJob{
							combination:    jobCombination,
							clusterKey:     clusterKey,
							version:        clusterSpec.Version,
							log:            log,
							rootCtx:        rootCtx,
							resolver:       resolver,
							opts:           opts,
							providerConfig: providerConfig,
							distribution:   d,
						}
					}
				}
			}
		}()

		go func() {
			workerWg.Wait()
			close(results)
		}()

		for result := range results {
			if result.err != nil {
				log.Errorw("Scenario generation worker failed", "error", result.err)
				continue
			}
			if result.machineSpec.ProviderSpec.Value == nil {
				continue
			}
			if _, ok := localScenarios[result.clusterKey]; !ok {
				localScenarios[result.clusterKey] = make(map[string]v1alpha1.MachineSpec)
				localMachineDescriptions[result.clusterKey] = make(map[string][]string)
			}
			if existing, exists := localScenarios[result.clusterKey][result.dedupKey]; exists {
				merged := existing
				if err := mergo.Merge(&merged, result.machineSpec, mergo.WithOverride); err == nil {
					localScenarios[result.clusterKey][result.dedupKey] = merged
					localMachineDescriptions[result.clusterKey][result.dedupKey] = append(localMachineDescriptions[result.clusterKey][result.dedupKey], result.machineName)
				}
			} else {
				localScenarios[result.clusterKey][result.dedupKey] = result.machineSpec
				localMachineDescriptions[result.clusterKey][result.dedupKey] = []string{result.machineName}
			}
		}
		return localScenarios, localMachineDescriptions
	}

	includedScenarios, includedDescriptions := combineModifiers(included, "included")
	excludedScenarios, excludedDescriptions := combineModifiers(excluded, "excluded")

	// Merge both maps
	for clusterKey, deduped := range includedScenarios {
		if _, ok := finalScenarios[clusterKey]; !ok {
			finalScenarios[clusterKey] = make(map[string]v1alpha1.MachineSpec)
		}
		for dedupKey, spec := range deduped {
			finalScenarios[clusterKey][dedupKey] = spec
		}
	}
	for clusterKey, deduped := range excludedScenarios {
		if _, ok := finalScenarios[clusterKey]; !ok {
			finalScenarios[clusterKey] = make(map[string]v1alpha1.MachineSpec)
		}
		for dedupKey, spec := range deduped {
			finalScenarios[clusterKey][dedupKey] = spec
		}
	}
	for clusterKey, descMap := range includedDescriptions {
		if _, ok := finalMachineDescriptions[clusterKey]; !ok {
			finalMachineDescriptions[clusterKey] = make(map[string][]string)
		}
		for dedupKey, descs := range descMap {
			finalMachineDescriptions[clusterKey][dedupKey] = descs
		}
	}
	for clusterKey, descMap := range excludedDescriptions {
		if _, ok := finalMachineDescriptions[clusterKey]; !ok {
			finalMachineDescriptions[clusterKey] = make(map[string][]string)
		}
		for dedupKey, descs := range descMap {
			finalMachineDescriptions[clusterKey][dedupKey] = descs
		}
	}

	log.Infof("Finished generating scenarios with included/excluded grouping.")
	var finalMachineDescriptionsSlice iter.Seq[string]
	for _, v := range includedScenarios {
		finalMachineDescriptionsSlice = maps.Keys(v)
		break
	}
	return finalScenarios, finalMachineDescriptions, finalMachineDescriptionsSlice
}
