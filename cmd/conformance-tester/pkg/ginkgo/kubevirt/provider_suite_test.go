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
	myOpt        string
	datacenters  string
	kubeVersions string
	// dcClusters used by provider_kubevirt.go
	dcClusters          = map[string]*kubermaticv1.Datacenter{}
	skipClusterDeletion bool
	skipClusterCreation bool
	updateClusters      bool
	clustersToDelete    []string
	// Ensure defaultClusterSettings is a map
	// defaultClusterSettings = map[string]kubermaticv1.ClusterSpec{}
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
	// setup context
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

	// load cli-flags
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

	fmt.Print("GetClusterVersions: ", GetClusterVersions(), "\n")
	fmt.Print("GetDatacenterDescriptions: ", GetDatacenterDescriptions(), "\n")
	fmt.Print("GetClusterDescriptions: ", GetClusterDescriptions(), "\n")
	fmt.Print("GetMachineDescriptions: ", GetMachineDescriptions(), "\n")

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
		err = mergo.Merge(&dst, defaultDatacenterSettings)
		if err != nil {
			log.Fatalw("Failed to merge seed", zap.String("datacenter", k), zap.Error(err))
		}
		seed.Spec.Datacenters[k] = dst
		defaultSeedSettings[k] = *seed
	}

	versionManager := version.NewFromConfiguration(kkpConfig)
	versions, err := versionManager.GetVersionsForProvider(kubermaticv1.KubevirtCloudProvider)
	if err != nil {
		log.Fatalw("Failed to get versions for provider", zap.Error(err))
	}
	var clustersMu sync.Mutex
	var wgClusters sync.WaitGroup
	maxConcurrentClusters := 20 // set your desired concurrency limit for clusters
	semClusters := make(chan struct{}, maxConcurrentClusters)
	newClusters = map[string]*kubermaticv1.ClusterSpec{}
	for _, kubeVersion := range versions {
		for k, v := range clusterSettings {
			for dcKey, seed := range defaultSeedSettings {
				wgClusters.Add(1)
				semClusters <- struct{}{} // acquire a slot
				go func(k string, v kubermaticv1.ClusterSpec, dcKey string, seed kubermaticv1.Seed, kubeVersion *version.Version) {
					defer wgClusters.Done()
					defer func() { <-semClusters }() // release the slot
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
					err = defaulting.DefaultClusterSpec(rootCtx, &v, nil, &seed, kkpConfig, p)
					if err != nil {
						log.Fatalw("Failed to default cluster", zap.String("cluster spec", k), zap.Error(err))
					}
					if err := validation.ValidateClusterSpec(&v, &currentSeedDatacenter, nil, versionManager, &v.Version, nil); len(err) != 0 {
						log.Infof("Failed to validate cluster", zap.String("cluster spec", k), zap.Error(err.ToAggregate()))
						return
					}
					clustersMu.Lock()
					// defaultClusterSettings["with kube version "+kubeVersion.Version.String()+" "+dcKey+" "+k] = v
					newClusters["with kube version "+kubeVersion.Version.String()+" "+dcKey+" "+k] = &v
					clustersMu.Unlock()
				}(k, v, dcKey, seed, kubeVersion)
			}
		}
	}
	wgClusters.Wait()

	defaultMachineSettings := map[string]v1alpha1.MachineSpec{}
	resolver := configvar.NewResolver(rootCtx, runtimeOpts.SeedClusterClient)
	newScenarios = map[string]map[string]v1alpha1.MachineSpec{}
	var machinesMu sync.Mutex
	var wgMachines sync.WaitGroup
	maxConcurrent := 40 // set your desired concurrency limit
	sem := make(chan struct{}, maxConcurrent)
	for machineKey, machine := range machineSettings {
		for k := range newClusters {
			wgMachines.Add(1)
			sem <- struct{}{} // acquire a slot
			go func(machineKey string, machine v1alpha1.MachineSpec, k string) {
				defer wgMachines.Done()
				defer func() { <-sem }() // release the slot
				p := mckubevirtprovider.New(resolver)
				var t kubevirt.RawConfig
				if err := json.Unmarshal(machine.ProviderSpec.Value.Raw, &t); err != nil {
					fmt.Println("Error unmarshalling JSON:", err)
					return
				}
				t.Auth.Kubeconfig.Value = b64.StdEncoding.EncodeToString([]byte(opts.Secrets.Kubevirt.Kubeconfig))
				err = mergo.Merge(&t, defaultKubevirtConfig)
				if err != nil {
					log.Fatalw("Failed to merge seed", zap.String("datacenter", k), zap.Error(err))
				}
				raw, err := EncodeRawSpec(t)
				if err != nil {
					log.Fatalw("Failed to encode raw spec", zap.String("machine", machineKey), zap.Error(err))
				}
				osspec, err := userdata.DefaultOperatingSystemSpec(providerconfig.OperatingSystemUbuntu, runtime.RawExtension{})
				if err != nil {
					log.Fatalw("Failed to get default OS spec", zap.String("machine", machineKey), zap.Error(err))
				}
				pc := providerconfig.Config{
					CloudProviderSpec:   *raw,
					OperatingSystemSpec: osspec,
					OperatingSystem:     providerconfig.OperatingSystemUbuntu,
				}
				data, err := json.Marshal(pc)
				if err != nil {
					log.Fatalw("Failed to marshal config", zap.String("machine", machineKey), zap.Error(err))
				}
				machine.ProviderSpec.Value.Raw = data
				if machine.ProviderSpec.Value != nil {
					fmt.Fprintf(file, "[DEBUG] scenario %s and %s\n", k, machineKey)
				}
				machineSpec, err := p.AddDefaults(log, machine)
				if err != nil {
					log.Fatalw("Failed to add defaults to machine", zap.String("machine", machineKey), zap.Error(err))
					return
				}
				if err := p.Validate(rootCtx, log, machineSpec); err != nil {
					log.Infof("Failed to validate machine", zap.String("machine", machineKey), zap.Error(err))
				}
				machinesMu.Lock()
				defaultMachineSettings[k+" "+machineKey] = machineSpec
				if newScenarios[k] == nil {
					newScenarios[k] = map[string]v1alpha1.MachineSpec{}
				}
				newScenarios[k][machineKey] = machineSpec
				machinesMu.Unlock()
			}(machineKey, machine, k)
		}
	}
	wgMachines.Wait()

	total := 0
	for _, inner := range newScenarios {
		total += len(inner)
	}

	// scenarioFailureMap = make(map[string][]Failure)
	flag.Parse()
	fmt.Fprintf(file, "new clusters: %v\nnew scenarios: %v\nKeys: %v\n", len(newClusters), total, maps.Keys(newScenarios))
	if configPath == "" {
		runtimeOpts, _ = k8cginkgo.NewRuntimeOptions(rootCtx, log, &k8cginkgo.Options{
			KubermaticNamespace: legacyOpts.KubermaticNamespace,
			KubermaticSeedName:  legacyOpts.KubermaticSeedName,
		})
	}

	// merge options by file and cli flags
	legacyOpts = k8cginkgo.MergeOptions(log, opts, legacyOpts, runtimeOpts)

	// parse our CLI flags
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
			err := runtimeOpts.SeedClusterClient.Get(rootCtx, types.NamespacedName{Name: clusterName}, cluster)
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
