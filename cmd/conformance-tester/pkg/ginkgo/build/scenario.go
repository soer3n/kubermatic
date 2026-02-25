package build

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"sort"

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

func GetTableEntries(rootCtx context.Context, log *zap.SugaredLogger, runtimeOpts *options.RuntimeOptions, legacyOpts *legacytypes.Options, opts *options.Options, infraClient ctrlclient.Client, projectName string, cloudProvider providerconfig.CloudProvider) (map[string]*Scenario, map[string]map[string][]string, map[string]map[string][]string, *kubermaticv1.Seed) {
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
	includedSeeds, excludedSeeds := buildDefaultSeedSettings(GenericDatacenterSettings(rootCtx, providerConfig, opts.Secrets), kkpConfig, log, defaultDatacenterSettings, opts.Excluded.DatacenterDescriptions, opts.Included.DatacenterDescriptions)
	seed := &kubermaticv1.Seed{}
	err = runtimeOpts.SeedClusterClient.Get(rootCtx, apitypes.NamespacedName{Name: "kubermatic", Namespace: "kubermatic"}, seed)
	if err != nil {
		log.Fatalw("Failed to get seed", zap.Error(err))
	}

	if seed.Spec.Datacenters == nil {
		seed.Spec.Datacenters = map[string]kubermaticv1.Datacenter{}
	}

	seedKeys := make([]string, 0, len(includedSeeds))
	for k := range includedSeeds {
		seedKeys = append(seedKeys, k)
	}
	for k := range excludedSeeds {
		seedKeys = append(seedKeys, k)
	}
	sort.Strings(seedKeys)

	for _, key := range seedKeys {
		s := includedSeeds[key]
		for dcName, dc := range s.Spec.Datacenters {
			hasher := sha1.New()
			hasher.Write([]byte(dcName))
			hashedName := hex.EncodeToString(hasher.Sum(nil))[:10]
			datacenterNameMappings[dcName] = hashedName
			dc.Country = "conformance"
			dc.Location = dcName
			seed.Spec.Datacenters[hashedName] = dc
		}
		s = excludedSeeds[key]
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
	newClusters, _, scenarios := buildNewClusters(rootCtx, versions, settings.ClusterSettings, includedSeeds, excludedSeeds, seed, opts, kkpConfig, log, versionManager, nil, providerConfig, opts.Excluded.ClusterDescriptions, opts.Included.ClusterDescriptions)
	resolver := configvar.NewResolver(rootCtx, runtimeOpts.SeedClusterClient)
	log.Info("generating scenarios...")
	machineSettings := MachineSettings(rootCtx, providerConfig, legacyOpts.KubermaticNamespace, opts.Secrets, &opts.Resources)
	machineSettings = append(machineSettings, ResourceMachineSettings(rootCtx, providerConfig, legacyOpts.KubermaticNamespace, opts.Secrets, &opts.Resources)...)
	_, _, includedMachineDescriptions, excludedMachineDescriptions, scenarios, _ := buildNewScenarios(scenarios, machineSettings, newClusters, opts, log, providerConfig, resolver, nil, rootCtx, legacyOpts.Distributions, opts.Excluded.MachineDescriptions, opts.Included.MachineDescriptions)

	return scenarios, includedMachineDescriptions, excludedMachineDescriptions, seed
}
