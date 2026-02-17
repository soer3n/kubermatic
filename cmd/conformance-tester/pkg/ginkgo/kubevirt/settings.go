package kubevirt

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"

	"go.uber.org/zap"
	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	"k8c.io/kubermatic/v2/pkg/version"
	"k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	"k8c.io/machine-controller/sdk/providerconfig/configvar"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	apitypes "k8s.io/apimachinery/pkg/types"
)

func GetMachineDescriptions(client ctrlclient.Client) map[string]k8cginkgo.Description {
	// client, _, err := utils.GetClients()
	// if err != nil {
	// 	return nil
	// }

	settings := MachineSettings(context.Background(), client, "kubermatic", nil)

	groupedSettings := map[string]k8cginkgo.Description{}
	groupedSettingsDesc := map[string][]string{}
	for _, modifier := range settings {
		groupedSettingsDesc[modifier.group] = append(groupedSettingsDesc[modifier.group], modifier.name)
	}

	for group, descs := range groupedSettingsDesc {
		strippedDescs := stripPrefix(descs)
		if len(strippedDescs) == 1 {
			strippedDescs = nil
		}
		groupedSettings[group] = k8cginkgo.Description{
			Name:    longestCommonPrefixTokens(descs, " "),
			Options: strippedDescs,
		}
	}
	return groupedSettings
}

func GetDatacenterDescriptions(client ctrlclient.Client) map[string]k8cginkgo.Description {
	// client, _, err := utils.GetClients()
	// if err != nil {
	// 	return nil
	// }
	settings := GenericDatacenterSettings(context.Background(), client, "kubermatic")
	groupedSettings := map[string]k8cginkgo.Description{}
	groupedSettingsDesc := map[string][]string{}
	for _, modifier := range settings {
		groupedSettingsDesc[modifier.Group] = append(groupedSettingsDesc[modifier.Group], modifier.Name)
	}

	for group, descs := range groupedSettingsDesc {
		strippedDescs := stripPrefix(descs)
		if len(strippedDescs) == 1 {
			strippedDescs = nil
		}
		groupedSettings[group] = k8cginkgo.Description{
			Name:    longestCommonPrefix(descs),
			Options: strippedDescs,
		}
	}
	return groupedSettings
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	prefix := strs[0]

	for _, s := range strs[1:] {
		for len(prefix) > 0 && !strings.HasPrefix(s, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
		if prefix == "" {
			return ""
		}
	}
	return prefix
}

func longestCommonPrefixTokens(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}

	base := strings.Split(strs[0], sep)
	maxTokens := len(base)

	for _, s := range strs[1:] {
		tokens := strings.Split(s, sep)

		i := 0
		for i < maxTokens && i < len(tokens) && tokens[i] == base[i] {
			i++
		}
		maxTokens = i

		if maxTokens == 0 {
			return ""
		}
	}

	prefix := strings.Join(base[:maxTokens], sep)

	// preserve trailing separator if it existed
	return prefix + sep
}

func stripPrefix(strs []string) []string {
	prefix := longestCommonPrefixTokens(strs, " ")
	out := make([]string, 0, len(strs))

	for _, s := range strs {
		out = append(out, strings.TrimPrefix(s, prefix))
	}
	return out
}

type scenarioInfo struct {
	ClusterSpec  *kubermaticv1.ClusterSpec
	Machine      v1alpha1.MachineSpec
	ScenarioName string
	ProjectName  string
	Exclude      bool
	Description  string
}

func GetTableEntries(rootCtx context.Context, log *zap.SugaredLogger, runtimeOpts *k8cginkgo.RuntimeOptions, legacyOpts *legacytypes.Options, opts *k8cginkgo.Options, infraClient ctrlclient.Client, projectName string) map[string][]scenarioInfo {

	kkpConfig, err := k8cginkgo.LoadKubermaticConfiguration()
	if err != nil {
		log.Fatalw("Failed to load KKP configuration", zap.Error(err))
	}
	file, err := os.Create("debug_output.txt")
	if err != nil {
		log.Fatalw("Failed to create debug output file", zap.Error(err))
	}
	defer file.Close()
	log.Info("generating seeds...")
	datacenterNameMappings := make(map[string]string)
	defaultSeedSettings := buildDefaultSeedSettings(GenericDatacenterSettings(rootCtx, runtimeOpts.SeedClusterClient, legacyOpts.KubermaticNamespace), kkpConfig, log, defaultDatacenterSettings, opts.Excluded.DatacenterDescriptions, opts.Included.DatacenterDescriptions)
	seed := &kubermaticv1.Seed{}
	err = runtimeOpts.SeedClusterClient.Get(rootCtx, apitypes.NamespacedName{Name: "kubermatic", Namespace: "kubermatic"}, seed)
	if err != nil {
		log.Fatalw("Failed to get seed", zap.Error(err))
	}

	if seed.Spec.Datacenters == nil {
		seed.Spec.Datacenters = map[string]kubermaticv1.Datacenter{}
	}

	seedKeys := make([]string, 0, len(defaultSeedSettings))
	for k := range defaultSeedSettings {
		seedKeys = append(seedKeys, k)
	}
	sort.Strings(seedKeys)

	for _, key := range seedKeys {
		s := defaultSeedSettings[key]
		for dcName, dc := range s.Spec.Datacenters {
			hasher := sha1.New()
			hasher.Write([]byte(dcName))
			hashedName := hex.EncodeToString(hasher.Sum(nil))[:10]
			datacenterNameMappings[dcName] = hashedName
			dc.Country = "conformance"
			dc.Location = dcName
			seed.Spec.Datacenters[hashedName] = dc
		}
	}

	versionManager := version.NewFromConfiguration(kkpConfig)
	versions, err := versionManager.GetVersionsForProvider(kubermaticv1.KubevirtCloudProvider)
	versions = []*version.Version{}
	for _, v := range opts.Releases {
		versionObj, _ := versionManager.GetVersion(v)
		versions = append(versions, versionObj)
	}
	log.Info("generating clusters...")
	newClusters, finalClusterDescriptions := buildNewClusters(rootCtx, versions, k8cginkgo.ClusterSettings, defaultSeedSettings, seed, opts, kkpConfig, log, versionManager, file, opts.Excluded.ClusterDescriptions, opts.Included.ClusterDescriptions)
	resolver := configvar.NewResolver(rootCtx, runtimeOpts.SeedClusterClient)
	fmt.Fprintf(file, "\nGenerated Clusters: %v\n", len(newClusters))
	defaultKubevirtConfig, err := getDefaultKubevirtConfig(infraClient)
	if err != nil {
		log.Fatalw("Failed to get default kubevirt config", zap.Error(err))
	}
	fmt.Fprintf(file, "Default KubeVirt Config: %+v\n", defaultKubevirtConfig)
	fmt.Fprint(file, "\nGenerated Scenarios:\n")
	log.Info("generating scenarios...")
	newScenarios, finalMachineDescriptions := buildNewScenarios(MachineSettings(rootCtx, infraClient, legacyOpts.KubermaticNamespace, &opts.Resources), newClusters, opts, log, *defaultKubevirtConfig, resolver, file, rootCtx, opts.Excluded.MachineDescriptions, opts.Included.MachineDescriptions, infraClient)

	// var newEntries []TableEntry
	var groupedEntries map[string][]scenarioInfo
	versionSlice := []string{}
	if len(opts.Releases) > 0 {
		for _, v := range opts.Releases {
			versionSlice = append(versionSlice, v)
		}
	} else {
		for _, scenario := range kkpConfig.Spec.Versions.Versions {
			versionSlice = append(versionSlice, scenario.String())
		}
	}
	for seedKey, _ := range defaultSeedSettings {
		for clusterName, machines := range newScenarios {
			clusterSpec, ok := newClusters[clusterName]
			if !ok {
				continue
			}
			clusterDesc, ok := finalClusterDescriptions[clusterName]
			if !ok {
				continue
			}

			for scenario, machine := range machines {
				desc, ok := finalMachineDescriptions[clusterName][scenario]
				if !ok {
					continue
				}
				title := fmt.Sprintf("kubernetes version %s and %s and %s and %s", clusterSpec.Version.String(), seedKey, strings.Join(clusterDesc, " and "), strings.Join(desc, " and "))
				// entry := Entry(title, title, clusterName, clusterSpec, &machine, scenario, Label("kubevirt"))
				// if !slice.ContainsString(versionSlice, clusterSpec.Version.String(), nil) {
				// 	entry = Entry(title, title, clusterName, clusterSpec, &machine, scenario, Label("skip"))
				// 	newEntries = append(newEntries, entry)
				// 	continue
				// }

				exclude := false
				if len(opts.Included.DatacenterDescriptions) > 0 {
					for _, included := range opts.Included.DatacenterDescriptions {
						if !strings.Contains(title, included) {
							exclude = true
							break
						}
					}
				} else {
					for _, excluded := range opts.Excluded.DatacenterDescriptions {
						if strings.Contains(title, excluded) {
							exclude = true
							break
						}
					}
				}

				if !exclude {
					if len(opts.Included.ClusterDescriptions) > 0 {
						for _, included := range opts.Included.ClusterDescriptions {
							if !strings.Contains(title, included) {
								exclude = true
								break
							}
						}
					} else {
						for _, excluded := range opts.Excluded.ClusterDescriptions {
							if strings.Contains(title, excluded) {
								exclude = true
								break
							}
						}
					}
				}

				if !exclude {
					if len(opts.Included.MachineDescriptions) > 0 {
						for _, included := range opts.Included.MachineDescriptions {
							if !strings.Contains(title, included) {
								exclude = true
								break
							}
						}
					} else {
						for _, excluded := range opts.Excluded.MachineDescriptions {
							if strings.Contains(title, excluded) {
								exclude = true
								break
							}
						}
					}
				}

				// if exclude {
				// 	entry = Entry(title, title, clusterName, clusterSpec, &machine, scenario, Label("skip"))
				// }
				if groupedEntries == nil {
					groupedEntries = make(map[string][]scenarioInfo)
				}
				if groupedEntries[clusterName] == nil {
					groupedEntries[clusterName] = []scenarioInfo{}
				}
				groupedEntries[clusterName] = append(groupedEntries[clusterName], scenarioInfo{
					ClusterSpec:  clusterSpec,
					Machine:      machine,
					ScenarioName: scenario,
					ProjectName:  projectName,
					Exclude:      exclude,
					Description:  title,
				})
				// newEntries = append(newEntries, entry)
			}
			continue
		}
	}
	// newEntries = []TableEntry{}
	// for clusterName, entry := range groupedEntries {
	// 	newEntries = append(newEntries, Entry(clusterName, clusterName, projectName, entry))
	// }
	return groupedEntries
}
