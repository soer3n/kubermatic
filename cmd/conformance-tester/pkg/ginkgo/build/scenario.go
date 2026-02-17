package build

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"iter"
	"sort"
	"strings"

	"go.uber.org/zap"
	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/options"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/settings"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	"k8c.io/kubermatic/v2/pkg/version"
	"k8c.io/machine-controller/sdk/providerconfig"
	"k8c.io/machine-controller/sdk/providerconfig/configvar"
	apitypes "k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func GetTableEntries(rootCtx context.Context, log *zap.SugaredLogger, runtimeOpts *options.RuntimeOptions, legacyOpts *legacytypes.Options, opts *options.Options, infraClient ctrlclient.Client, projectName string, cloudProvider providerconfig.CloudProvider) (map[string][]ScenarioInfo, map[string]kubermaticv1.Seed, map[string]*kubermaticv1.ClusterSpec, map[string][]string, map[string]string, iter.Seq[string]) {
	kkpConfig, err := options.LoadKubermaticConfiguration()
	if err != nil {
		log.Fatalw("Failed to load KKP configuration", zap.Error(err))
	}
	log.Info("generating seeds...")
	providerConfig, err := getProviderConfig(rootCtx, log, opts.Secrets, cloudProvider)
	if err != nil {
		log.Fatalw("Failed to get default kubevirt config", zap.Error(err))
	}
	datacenterNameMappings := make(map[string]string)
	defaultSeedSettings := buildDefaultSeedSettings(GenericDatacenterSettings(rootCtx, providerConfig, opts.Secrets), kkpConfig, log, defaultDatacenterSettings, opts.Excluded.DatacenterDescriptions, opts.Included.DatacenterDescriptions)
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
	newClusters, finalClusterDescriptions := buildNewClusters(rootCtx, versions, settings.ClusterSettings, defaultSeedSettings, seed, opts, kkpConfig, log, versionManager, nil, providerConfig, opts.Excluded.ClusterDescriptions, opts.Included.ClusterDescriptions)
	resolver := configvar.NewResolver(rootCtx, runtimeOpts.SeedClusterClient)
	log.Info("generating scenarios...")
	machineSettings := MachineSettings(rootCtx, providerConfig, legacyOpts.KubermaticNamespace, opts.Secrets, &opts.Resources)
	machineSettings = append(machineSettings, ResourceMachineSettings(rootCtx, providerConfig, legacyOpts.KubermaticNamespace, opts.Secrets, &opts.Resources)...)
	newScenarios, finalMachineDescriptions, finalMachineDescriptionsSlice := buildNewScenarios(machineSettings, newClusters, opts, log, providerConfig, resolver, nil, rootCtx, legacyOpts.Distributions, opts.Excluded.MachineDescriptions, opts.Included.MachineDescriptions)

	var groupedEntries map[string][]ScenarioInfo
	var machinesPerCluster []string
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
				exclude := false
				if len(opts.Included.DatacenterDescriptions) > 0 {
					for _, included := range opts.Included.DatacenterDescriptions {
						if strings.Contains(title, included) {
							exclude = false
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
					if machinesPerCluster == nil {
						machinesPerCluster = finalMachineDescriptions[clusterName][scenario]
					}
				}

				if !exclude {
					if len(opts.Included.ClusterDescriptions) > 0 {
						for _, included := range opts.Included.ClusterDescriptions {
							if strings.Contains(title, included) {
								exclude = false
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
							if strings.Contains(title, included) {
								exclude = false
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

				if groupedEntries == nil {
					groupedEntries = make(map[string][]ScenarioInfo)
				}
				if groupedEntries[clusterName] == nil {
					groupedEntries[clusterName] = []ScenarioInfo{}
				}
				groupedEntries[clusterName] = append(groupedEntries[clusterName], ScenarioInfo{
					ClusterSpec:  clusterSpec,
					Machine:      machine,
					ScenarioName: scenario,
					ProjectName:  projectName,
					Exclude:      exclude,
					Description:  title,
				})

			}
			continue
		}
	}

	return groupedEntries, defaultSeedSettings, newClusters, finalClusterDescriptions, datacenterNameMappings, finalMachineDescriptionsSlice

}
