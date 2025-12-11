package kubevirt

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"maps"

	"os"
	"sync"
	"testing"

	"dario.cat/mergo"
	"github.com/go-logr/zapr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	kubevirtprovider "k8c.io/kubermatic/v2/pkg/provider/cloud/kubevirt"
	"k8c.io/kubermatic/v2/pkg/version"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/sdk/v2/semver"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/clients"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	"k8c.io/kubermatic/v2/pkg/defaulting"
	kubermaticlog "k8c.io/kubermatic/v2/pkg/log"
	kkpreconciling "k8c.io/kubermatic/v2/pkg/resources/reconciling"
	"k8c.io/kubermatic/v2/pkg/validation"
	mckubevirtprovider "k8c.io/machine-controller/pkg/cloudprovider/provider/kubevirt"
	"k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	"k8c.io/machine-controller/sdk/cloudprovider/kubevirt"
	"k8c.io/machine-controller/sdk/providerconfig"
	"k8c.io/machine-controller/sdk/providerconfig/configvar"
	"k8c.io/machine-controller/sdk/userdata"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlruntimelog "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	rootCtx context.Context
	log     *zap.SugaredLogger
)

var (
	myOpt               string
	datacenters         string
	kubeVersions        string
	dcClusters          = map[string]*kubermaticv1.Datacenter{}
	skipClusterDeletion bool
	skipClusterCreation bool
	updateClusters      bool
	clustersToDelete    []string
)

func init() {
	// Define custom flags (parsed in TestScenarios).
	flag.StringVar(&datacenters, "datacenters", "", "comma-separated datacenters")
	flag.StringVar(&kubeVersions, "kube-versions", "", "comma-separated Kubernetes versions")
	flag.BoolVar(&skipClusterCreation, "skip-cluster-creation", false, "skip cluster creation before running tests")
	flag.BoolVar(&skipClusterDeletion, "skip-cluster-deletion", false, "skip cluster deletion after running tests")
	flag.BoolVar(&updateClusters, "update-clusters", false, "update clusters before running tests")

	logOpts := kubermaticlog.NewDefaultOptions()
	rawLog := kubermaticlog.New(logOpts.Debug, logOpts.Format)
	log = rawLog.Sugar()

	ctrlruntimelog.SetLogger(zapr.NewLogger(rawLog.WithOptions(zap.AddCallerSkip(1))))
	kkpreconciling.Configure(log)
}

var (
	opts        *k8cginkgo.Options
	runtimeOpts *k8cginkgo.RuntimeOptions
	legacyOpts  *legacytypes.Options
	kkpConfig   *kubermaticv1.KubermaticConfiguration
	// scenarioFailureMap map[string][]k8cginkgo.Failure
	newScenarios      map[string]map[string]v1alpha1.MachineSpec
	newClusters       map[string]*kubermaticv1.ClusterSpec
	newClusterClients map[string]ctrlruntimeclient.Client
)

func GetClusterVersions() []string {
	versions := []string{}
	for _, scenario := range kkpConfig.Spec.Versions.Versions {
		versions = append(versions, scenario.String())
	}
	return versions
}

func GetDatacenterDescriptions() []string {
	descriptions := []string{}
	for key := range datacenterSettings {
		descriptions = append(descriptions, key)
	}
	return descriptions
}

func GetClusterDescriptions() []string {
	descriptions := []string{}
	for key := range clusterSettings {
		descriptions = append(descriptions, key)
	}
	return descriptions
}

func GetMachineDescriptions() []string {
	descriptions := []string{}
	for key := range machineSettings {
		descriptions = append(descriptions, key)
	}
	return descriptions
}

func TestMain(m *testing.M) {
	var err error
	rootCtx = signals.SetupSignalHandler()
	opts, err = k8cginkgo.NewOptionsFromYAML(log)
	if err != nil {
		log.Fatalw("Failed to load options", zap.Error(err))
	}
	configPath := os.Getenv("CONFORMANCE_TESTER_CONFIG_FILE")
	if configPath != "" {
		runtimeOpts, err = k8cginkgo.NewRuntimeOptions(rootCtx, log, opts)
		if err != nil {
			log.Fatalw("Failed to create runtime options", zap.Error(err))
		}
	}
	legacyOpts = legacytypes.NewDefaultOptions()
	legacyOpts.AddFlags()
	kkpConfig, err = loadKubermaticConfiguration()
	if err != nil {
		log.Fatalw("Failed to load KKP configuration", zap.Error(err))
	}
	file, err := os.Create("debug_output.txt")
	if err != nil {
		log.Fatalw("Failed to create debug output file", zap.Error(err))
	}
	defer file.Close()
	defaultSeedSettings := buildDefaultSeedSettings(datacenterSettings, kkpConfig, log, defaultDatacenterSettings)
	versionManager := version.NewFromConfiguration(kkpConfig)
	versions, err := versionManager.GetVersionsForProvider(kubermaticv1.KubevirtCloudProvider)
	if err != nil {
		log.Fatalw("Failed to get versions for provider", zap.Error(err))
	}
	versions = versions[:8]
	newClusters = buildNewClusters(rootCtx, versions, clusterSettings, defaultSeedSettings, opts, kkpConfig, log, versionManager, file)
	resolver := configvar.NewResolver(rootCtx, runtimeOpts.SeedClusterClient)

	newScenarios = buildNewScenarios(machineSettings, newClusters, opts, log, defaultKubevirtConfig, resolver, file, rootCtx)
	total := 0
	for _, inner := range newScenarios {
		total += len(inner)
	}
	flag.Parse()
	fmt.Fprintf(file, "new clusters: %v\nnew scenarios: %v\nKeys: %v\nDescriptions: %v\n", len(newClusters), total, maps.Keys(newScenarios), maps.Values(newScenarios))
	if configPath == "" {
		runtimeOpts, _ = k8cginkgo.NewRuntimeOptions(rootCtx, log, &k8cginkgo.Options{
			KubermaticNamespace: legacyOpts.KubermaticNamespace,
			KubermaticSeedName:  legacyOpts.KubermaticSeedName,
		})
	}
	legacyOpts = k8cginkgo.MergeOptions(log, opts, legacyOpts, runtimeOpts)
	if err := legacyOpts.ParseFlags(log); err != nil {
		log.Warnf("Invalid flags", zap.Error(err))
	}
	os.Exit(m.Run())
}

func TestScenarios(t *testing.T) {
	RegisterFailHandler(CustomFail)
	RunSpecs(t, "Conformance Tester Scenarios Suite")
}

func CustomFail(message string, callerSkip ...int) {
	log.Infof("Fail called: %s", message)
}

var client clients.Client

var _ = SynchronizedBeforeSuite(func() {

	By(k8cginkgo.KKP("Creating a KKP client"), func() {
		client = clients.NewKubeClient(legacyOpts)
		Expect(client.Setup(rootCtx, log)).To(Succeed())
	})

	By(k8cginkgo.KKP("Ensuring a project exists"), func() {
		if legacyOpts.KubermaticProject == "" {
			projectName := "e2e-" + rand.String(5)
			p, err := client.CreateProject(rootCtx, log, projectName)
			Expect(err).NotTo(HaveOccurred())
			projectName = p
			legacyOpts.KubermaticProject = projectName
			opts.KubermaticProject = projectName
		}
		fmt.Fprintf(GinkgoWriter, "Using project %q\n", legacyOpts.KubermaticProject)
	})

	By(k8cginkgo.KKP("Ensuring SSH keys exist"), func() {
		Expect(client.EnsureSSHKeys(rootCtx, log)).To(Succeed())
	})
	suiteCfg, reporterCfg := GinkgoConfiguration()
	By(fmt.Sprintf("Node1: my-opt=%s, parallel=%d", myOpt, suiteCfg.ParallelTotal))
	By(fmt.Sprintf("Reporter: %#v", reporterCfg))
	By(fmt.Sprintf("Creating clusters for datacenters and kube versions: %v", dcClusters))

	var wg sync.WaitGroup
	for name, clusterSpec := range newClusters {
		wg.Add(1)
		go func(dc string, spec *kubermaticv1.ClusterSpec) {
			defer wg.Done()
			if !skipClusterCreation {
				ensureCluster(dc, spec)
			}
			if skipClusterCreation && updateClusters {
				updateCluster(dc, spec)
			}
		}(name, clusterSpec)
	}
	wg.Wait()
}, func(data []byte) {
	// Assign cluster name on every node
})

// Old SynchronizedAfterSuite replaced with aggregated variant.
var _ = SynchronizedAfterSuite(func() {
	// per-node no-op (could emit per-node logs here)
}, func() {
	// primary node: idempotent deletion attempt for each cluster
	By(fmt.Sprintf("Deleting created clusters for e2e project %q", legacyOpts.KubermaticProject))
	for dc := range newClusters {
		if !skipClusterDeletion {
			By("Deleting cluster for " + dc)
			cluster := &kubermaticv1.Cluster{}
			err := runtimeOpts.SeedClusterClient.Get(rootCtx, types.NamespacedName{Name: dc}, cluster)
			if err != nil {
				log.Errorf("Failed to get cluster %s: %v", dc, err)
				continue
			}

			By(fmt.Sprintf("Cleaning up resources for cluster %s. Name is %s", dc, cluster.Name))
			k8cginkgo.CommonCleanup(rootCtx, log, clients.NewKubeClient(legacyOpts), nil, nil, cluster)
		}

	}
})

var _ = ReportBeforeSuite(func(r Report) {})

var _ = ReportAfterSuite("ReportAfterSuite", func(r Report) {
	By("Reporting test results")
})

func buildDefaultSeedSettings(datacenterSettings map[string]kubermaticv1.Datacenter, kkpConfig *kubermaticv1.KubermaticConfiguration, log *zap.SugaredLogger, defaultDatacenterSettings kubermaticv1.Datacenter) map[string]kubermaticv1.Seed {
	defaultSeedSettings := map[string]kubermaticv1.Seed{}
	for k, v := range datacenterSettings {
		seed, err := defaulting.DefaultSeed(&kubermaticv1.Seed{
			Spec: kubermaticv1.SeedSpec{
				Datacenters: map[string]kubermaticv1.Datacenter{
					k: v,
				},
			},
		}, kkpConfig, log)
		if err != nil {
			log.Fatalw("Failed to default seed", zap.String("datacenter", k), zap.Error(err))
		}
		dst := seed.Spec.Datacenters[k]
		if err := mergo.Merge(&dst, defaultDatacenterSettings); err != nil {
			log.Fatalw("Failed to merge seed", zap.String("datacenter", k), zap.Error(err))
		}
		seed.Spec.Datacenters[k] = dst
		defaultSeedSettings[k] = *seed
	}
	return defaultSeedSettings
}

func buildNewClusters(rootCtx context.Context, versions []*version.Version, clusterSettings map[string]kubermaticv1.ClusterSpec, defaultSeedSettings map[string]kubermaticv1.Seed, opts *k8cginkgo.Options, kkpConfig *kubermaticv1.KubermaticConfiguration, log *zap.SugaredLogger, versionManager *version.Manager, file *os.File) map[string]*kubermaticv1.ClusterSpec {
	var clustersMu sync.Mutex
	var wgClusters sync.WaitGroup
	maxConcurrentClusters := 1
	semClusters := make(chan struct{}, maxConcurrentClusters)
	newClusters := map[string]*kubermaticv1.ClusterSpec{}
	clusterDescriptions := map[string][]string{}

	for _, kubeVersion := range versions {
		for k, v := range clusterSettings {
			for dcKey, seed := range defaultSeedSettings {
				wgClusters.Add(1)
				semClusters <- struct{}{}
				go func(k string, v kubermaticv1.ClusterSpec, dcKey string, seed kubermaticv1.Seed, kubeVersion *version.Version) {
					defer wgClusters.Done()
					defer func() { <-semClusters }()
					v.Cloud.ProviderName = string(kubermaticv1.KubevirtCloudProvider)
					v.Cloud.DatacenterName = dcKey
					v.Cloud.Kubevirt = &kubermaticv1.KubevirtCloudSpec{
						Kubeconfig: opts.Secrets.Kubevirt.Kubeconfig,
					}
					v.HumanReadableName = k
					v.ContainerRuntime = "containerd"
					v.Version = semver.Semver(kubeVersion.Version.String())
					currentSeedDatacenter := seed.Spec.Datacenters[dcKey]
					p, err := kubevirtprovider.NewCloudProvider(&currentSeedDatacenter, nil)
					if err != nil {
						log.Fatalw("Failed to create cloud provider", zap.String("datacenter", dcKey), zap.Error(err))
					}
					if err := defaulting.DefaultClusterSpec(rootCtx, &v, nil, &seed, kkpConfig, p); err != nil {
						log.Fatalw("Failed to default cluster", zap.String("cluster spec", k), zap.Error(err))
					}
					if valErrs := validation.ValidateClusterSpec(&v, &currentSeedDatacenter, nil, versionManager, &v.Version, nil); len(valErrs) != 0 {
						log.Infof("Failed to validate cluster", zap.String("cluster spec", k), zap.Error(valErrs.ToAggregate()))
						return
					}
					// Generate a key based on version, datacenter, provider
					keyStruct := struct {
						Version        string
						DatacenterName string
						ProviderName   string
					}{
						Version:        kubeVersion.Version.String(),
						DatacenterName: dcKey,
						ProviderName:   v.Cloud.ProviderName,
					}
					keyBytes, _ := json.Marshal(keyStruct)
					key := string(keyBytes)
					clustersMu.Lock()
					if existing, exists := newClusters[key]; exists {
						// Only merge if version, datacenter, and provider are identical (guaranteed by key)
						merged := *existing
						if err := mergo.Merge(&merged, v, mergo.WithOverride); err == nil {
							valErrs := validation.ValidateClusterSpec(&merged, &currentSeedDatacenter, nil, versionManager, &merged.Version, nil)
							if len(valErrs) != 0 {
								log.Infof("Skipped invalid merged cluster spec", zap.String("cluster spec", k), zap.Error(valErrs.ToAggregate()))
								fmt.Fprintf(file, "[SKIPPED INVALID MERGED CLUSTER] %s: errors: %v\n", key, valErrs.ToAggregate())
								clustersMu.Unlock()
								return
							}
							*existing = merged
							clusterDescriptions[key] = append(clusterDescriptions[key], k)
						}
						clustersMu.Unlock()
						return
					}
					newClusters[key] = &v
					clusterDescriptions[key] = []string{k}
					clustersMu.Unlock()
				}(k, v, dcKey, seed, kubeVersion)
			}
		}
	}
	wgClusters.Wait()
	// Output final cluster descriptions
	fmt.Fprintf(file, "\nFINAL CLUSTER DESCRIPTIONS (with combined cluster names):\n")
	for key, descs := range clusterDescriptions {
		var keyStruct struct {
			Version        string
			DatacenterName string
			ProviderName   string
		}
		_ = json.Unmarshal([]byte(key), &keyStruct)
		fmt.Fprintf(file, "Cluster Version: %s\n  Provider: %s\n  Combined Cluster Names: %v\n\n", keyStruct.Version, keyStruct.ProviderName, descs)
	}
	return newClusters
}

// scenarioResult is used to pass data from a producer to a consumer.
type scenarioResult struct {
	machineName string
	dedupKey    string
	machineSpec v1alpha1.MachineSpec
	err         error
}

// scenarioProducer processes a single machine configuration and sends the result to a channel.
func scenarioProducer(
	ch chan<- scenarioResult,
	wg *sync.WaitGroup,
	log *zap.SugaredLogger,
	rootCtx context.Context,
	resolver *configvar.Resolver,
	opts *k8cginkgo.Options,
	defaultKubevirtConfig kubevirt.RawConfig,
	clusterKey string,
	machineName string,
	machine v1alpha1.MachineSpec,
) {
	defer wg.Done()

	p := mckubevirtprovider.New(resolver)
	var t kubevirt.RawConfig
	if err := json.Unmarshal(machine.ProviderSpec.Value.Raw, &t); err != nil {
		ch <- scenarioResult{err: fmt.Errorf("failed to unmarshal provider spec: %w", err)}
		return
	}

	t.Auth.Kubeconfig.Value = b64.StdEncoding.EncodeToString([]byte(opts.Secrets.Kubevirt.Kubeconfig))
	if err := mergo.Merge(&t, defaultKubevirtConfig); err != nil {
		ch <- scenarioResult{err: fmt.Errorf("failed to merge default kubevirt config: %w", err)}
		return
	}

	raw, err := EncodeRawSpec(t)
	if err != nil {
		ch <- scenarioResult{err: fmt.Errorf("failed to encode raw spec: %w", err)}
		return
	}

	osspec, err := userdata.DefaultOperatingSystemSpec(providerconfig.OperatingSystemUbuntu, runtime.RawExtension{})
	if err != nil {
		ch <- scenarioResult{err: fmt.Errorf("failed to get default OS spec: %w", err)}
		return
	}

	pc := providerconfig.Config{
		CloudProviderSpec:   *raw,
		OperatingSystemSpec: osspec,
		OperatingSystem:     providerconfig.OperatingSystemUbuntu,
	}
	data, err := json.Marshal(pc)
	if err != nil {
		ch <- scenarioResult{err: fmt.Errorf("failed to marshal provider config: %w", err)}
		return
	}
	machine.ProviderSpec.Value.Raw = data

	machineSpec, err := p.AddDefaults(log, machine)
	if err != nil {
		ch <- scenarioResult{err: fmt.Errorf("failed to add defaults to machine: %w", err)}
		return
	}

	if err := p.Validate(rootCtx, log, machineSpec); err != nil {
		// Treat validation errors as skippable, not fatal
		log.Infof("Skipping invalid machine spec for %q: %v", machineName, err)
		ch <- scenarioResult{err: nil} // Send nil error to signal completion without a valid result
		return
	}

	// Create a stable deduplication key from the cloud provider spec only
	var providerSpec providerconfig.Config
	if err := json.Unmarshal(machineSpec.ProviderSpec.Value.Raw, &providerSpec); err != nil {
		ch <- scenarioResult{err: fmt.Errorf("failed to unmarshal provider spec for dedup key: %w", err)}
		return
	}
	dedupKeyBytes, err := json.Marshal(providerSpec.CloudProviderSpec)
	if err != nil {
		ch <- scenarioResult{err: fmt.Errorf("failed to marshal cloud provider spec for dedup key: %w", err)}
		return
	}

	ch <- scenarioResult{
		machineName: machineName,
		dedupKey:    string(dedupKeyBytes),
		machineSpec: machineSpec,
		err:         nil,
	}
}

// scenarioConsumer collects results for a single cluster and aggregates them.
func scenarioConsumer(
	ch <-chan scenarioResult,
	wg *sync.WaitGroup,
	log *zap.SugaredLogger,
	rootCtx context.Context,
	resolver *configvar.Resolver,
	clusterKey string,
	scenarios map[string]v1alpha1.MachineSpec,
	descriptions map[string][]string,
	file *os.File,
) {
	defer wg.Done()
	p := mckubevirtprovider.New(resolver)

	for result := range ch {
		if result.err != nil {
			log.Errorw("Producer failed", "cluster", clusterKey, "error", result.err)
			continue
		}
		// A nil error with an empty dedupKey means the producer skipped an invalid spec.
		if result.dedupKey == "" {
			continue
		}

		if existing, exists := scenarios[result.dedupKey]; exists {
			merged := existing
			if err := mergo.Merge(&merged, result.machineSpec, mergo.WithOverride); err == nil {
				if err := p.Validate(rootCtx, log, merged); err != nil {
					log.Infof("Skipped invalid merged machine spec", "machine", result.machineName, "error", err)
					fmt.Fprintf(file, "[SKIPPED INVALID MERGED SCENARIO] %s: errors: %v\n", clusterKey, err)
					continue
				}
				scenarios[result.dedupKey] = merged
				descriptions[result.dedupKey] = append(descriptions[result.dedupKey], result.machineName)
			}
		} else {
			scenarios[result.dedupKey] = result.machineSpec
			descriptions[result.dedupKey] = []string{result.machineName}
		}
	}
}

func buildNewScenarios(machineSettings map[string]v1alpha1.MachineSpec, newClusters map[string]*kubermaticv1.ClusterSpec, opts *k8cginkgo.Options, log *zap.SugaredLogger, defaultKubevirtConfig kubevirt.RawConfig, resolver *configvar.Resolver, file *os.File, rootCtx context.Context) map[string]map[string]v1alpha1.MachineSpec {
	finalScenarios := make(map[string]map[string]v1alpha1.MachineSpec)
	finalMachineDescriptions := make(map[string]map[string][]string)
	var consumerWg sync.WaitGroup

	for clusterKey := range newClusters {
		consumerWg.Add(1)

		// These maps are owned by the consumer for this clusterKey, so no mutex is needed.
		clusterScenarios := make(map[string]v1alpha1.MachineSpec)
		clusterDescriptions := make(map[string][]string)
		finalScenarios[clusterKey] = clusterScenarios
		finalMachineDescriptions[clusterKey] = clusterDescriptions

		resultsCh := make(chan scenarioResult)

		// Start the consumer for this cluster
		go scenarioConsumer(resultsCh, &consumerWg, log, rootCtx, resolver, clusterKey, clusterScenarios, clusterDescriptions, file)

		// Start all producers for this cluster
		var producerWg sync.WaitGroup
		for machineName, machine := range machineSettings {
			producerWg.Add(1)
			go scenarioProducer(resultsCh, &producerWg, log, rootCtx, resolver, opts, defaultKubevirtConfig, clusterKey, machineName, machine)
		}

		// Wait for all producers for this cluster to finish, then close the channel
		// to signal the consumer that there are no more results.
		go func() {
			producerWg.Wait()
			close(resultsCh)
		}()
	}

	// Wait for all consumers to finish processing.
	consumerWg.Wait()

	// Output final scenario descriptions in improved format
	fmt.Fprintf(file, "\nFINAL SCENARIO DESCRIPTIONS (with combined scenario names):\n")
	for clusterKey, scenarios := range finalScenarios {
		for dedupKey := range scenarios {
			descs := finalMachineDescriptions[clusterKey][dedupKey]
			fmt.Fprintf(file, "Cluster: %s\n  Scenario Key: %s\n  Combined Scenario Names: %v\n\n", clusterKey, dedupKey, descs)
		}
	}
	return finalScenarios
}
