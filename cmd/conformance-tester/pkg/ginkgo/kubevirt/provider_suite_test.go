package kubevirt

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/go-logr/zapr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/kubectl/pkg/util/slice"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlruntimelog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"k8c.io/kubermatic/v2/pkg/version"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/clients"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	kubermaticlog "k8c.io/kubermatic/v2/pkg/log"
	kkpreconciling "k8c.io/kubermatic/v2/pkg/resources/reconciling"

	"k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"

	apitypes "k8s.io/apimachinery/pkg/types"

	"k8c.io/machine-controller/sdk/cloudprovider/kubevirt"
	"k8c.io/machine-controller/sdk/providerconfig"
	"k8c.io/machine-controller/sdk/providerconfig/configvar"
)

type MachineScenario struct {
	Names    []string
	SpecHash string
}

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
	opts                     *k8cginkgo.Options
	runtimeOpts              *k8cginkgo.RuntimeOptions
	legacyOpts               *legacytypes.Options
	kkpConfig                *kubermaticv1.KubermaticConfiguration
	newScenarios             map[string]map[string]v1alpha1.MachineSpec
	newClusters              map[string]*kubermaticv1.ClusterSpec
	defaultSeedSettings      map[string]kubermaticv1.Seed
	finalClusterDescriptions map[string][]string
	finalMachineDescriptions map[string]map[string][]string
	newClusterClients        map[string]ctrlruntimeclient.Client
	datacenterNameMappings   map[string]string
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
	for _, modifier := range datacenterSettings {
		descriptions = append(descriptions, modifier.name)
	}
	return descriptions
}

func GetClusterDescriptions() []string {
	descriptions := []string{}
	for _, modifier := range clusterSettings {
		descriptions = append(descriptions, modifier.name)
	}
	return descriptions
}

func GetMachineDescriptions() []string {
	descriptions := []string{}
	for _, modifier := range machineSettings {
		descriptions = append(descriptions, modifier.name)
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
	log.Info("generating seeds...")
	datacenterNameMappings = make(map[string]string)
	defaultSeedSettings = buildDefaultSeedSettings(datacenterSettings, kkpConfig, log, defaultDatacenterSettings, opts.DatacenterDescriptions)

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

	err = runtimeOpts.SeedClusterClient.Update(rootCtx, seed)
	if err != nil {
		log.Fatalw("Failed to update seed", zap.Error(err))
	}

	versionManager := version.NewFromConfiguration(kkpConfig)
	versions, err := versionManager.GetVersionsForProvider(kubermaticv1.KubevirtCloudProvider)
	if err != nil {
		log.Fatalw("Failed to get versions for provider", zap.Error(err))
	}
	log.Info("generating clusters...")
	newClusters, finalClusterDescriptions = buildNewClusters(rootCtx, versions, clusterSettings, defaultSeedSettings, seed, opts, kkpConfig, log, versionManager, file, opts.ClusterDescriptions)
	resolver := configvar.NewResolver(rootCtx, runtimeOpts.SeedClusterClient)
	fmt.Fprintf(file, "\nGenerated Clusters: %v\n", len(newClusters))
	defaultKubevirtConfig, err := getDefaultKubevirtConfig()
	if err != nil {
		log.Fatalw("Failed to get default kubevirt config", zap.Error(err))
	}
	fmt.Fprintf(file, "Default KubeVirt Config: %+v\n", defaultKubevirtConfig)
	fmt.Fprint(file, "\nGenerated Scenarios:\n")
	log.Info("generating scenarios...")
	newScenarios, finalMachineDescriptions = buildNewScenarios(machineSettings, newClusters, opts, log, *defaultKubevirtConfig, resolver, file, rootCtx, opts.MachineDescriptions)
	// log.Infof("Final Machine Descriptions: %v\n", finalMachineDescriptions)
	// log.Info("post-processing scenarios...")

	// // Create and write to the scenarios summary file
	// summaryFile, err := os.Create("scenarios_summary.txt")
	// if err != nil {
	// 	log.Fatalw("Failed to create scenarios summary file", zap.Error(err))
	// }
	// defer summaryFile.Close()

	// fmt.Fprintln(summaryFile, "--- FINAL SCENARIOS SUMMARY ---")
	// for seedSettings := range defaultSeedSettings {
	// 	for _, kubeVersion := range versions {
	// 		for clusterKey, scenarios := range finalMachineDescriptions {
	// 			clusterDesc := "default"
	// 			if descs, ok := finalClusterDescriptions[clusterKey]; ok {
	// 				clusterDesc = strings.Join(descs, ", ")
	// 			}

	// 			for _, names := range scenarios {
	// 				machineDesc := strings.Join(names, ", ")
	// 				fmt.Fprintf(summaryFile, "A cluster with seed settings %s and kubernetes version %s and %s with a machine %s\n", strings.Replace(seedSettings, "-", " & ", -1), kubeVersion.Version.String(), clusterDesc, machineDesc)
	// 			}
	// 		}
	// 	}
	// }

	// total := 0
	// for _, inner := range newScenarios {
	// 	total += len(inner)
	// }
	flag.Parse()

	// Improved debug output
	// fmt.Fprintf(file, "new clusters: %d\n", len(newClusters))
	// fmt.Fprintf(file, "new scenarios: %d\n", total)

	if configPath == "" {
		runtimeOpts, _ = k8cginkgo.NewRuntimeOptions(rootCtx, log, &k8cginkgo.Options{
			KubermaticNamespace: legacyOpts.KubermaticNamespace,
			KubermaticSeedName:  legacyOpts.KubermaticSeedName,
		})
	}
	testSlice := []string{
		legacytypes.StorageTests, legacytypes.LoadbalancerTests, legacytypes.MetricsTests,
		legacytypes.UserClusterRBACTests, legacytypes.K8sGcrImageTests, legacytypes.MetricsTests,
		legacytypes.SecurityContextTests,
	}
	// enable all tests
	opts.EnableTests = []string{}
	for _, test := range testSlice {
		opts.EnableTests = append(opts.EnableTests, test)
	}
	legacyOpts = k8cginkgo.MergeOptions(log, opts, legacyOpts, runtimeOpts)
	if err := legacyOpts.ParseFlags(log); err != nil {
		log.Warnf("Invalid flags", zap.Error(err))
	}

	os.Exit(m.Run())
}

func EncodeRawSpec(rawConfig kubevirt.RawConfig) (*runtime.RawExtension, error) {
	pconfig := providerconfig.Config{
		CloudProviderSpec: runtime.RawExtension{
			Raw: toJSON(rawConfig),
		},
	}

	raw, err := json.Marshal(pconfig)
	if err != nil {
		return nil, err
	}

	return &runtime.RawExtension{Raw: raw}, nil
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

	By(k8cginkgo.KKP("Attaching datacenters to seed"), func() {

	})

	suiteCfg, reporterCfg := GinkgoConfiguration()
	By(fmt.Sprintf("Node1: my-opt=%s, parallel=%d", myOpt, suiteCfg.ParallelTotal))
	By(fmt.Sprintf("Reporter: %#v", reporterCfg))
	By(fmt.Sprintf("Creating clusters for datacenters and kube versions: %v", maps.Keys(newClusters)))

	var wg sync.WaitGroup
	maxConcurrent := 4 // Set your desired concurrency limit
	sem := make(chan struct{}, maxConcurrent)
	versionSlice := []string{}
	for _, v := range opts.Releases {
		versionSlice = append(versionSlice, v)
	}
	for i, _ := range defaultSeedSettings {
		log.Infof("defaultSeedSettings[%d]: %+v", i, datacenterNameMappings[i])
	}
	for i, _ := range newClusters {
		log.Infof("newClusters[%d]: %+v", i, finalClusterDescriptions[i])
	}
	for seedKey := range defaultSeedSettings {
		exclude := false
		for _, excluded := range opts.DatacenterDescriptions {
			if strings.Contains(seedKey, excluded) {
				exclude = true
				break
			}
		}
		if exclude {
			continue
		}
		for name, clusterSpec := range newClusters {
			if !slice.ContainsString(versionSlice, clusterSpec.Version.String(), nil) {
				continue
			}
			clusterDesc, ok := finalClusterDescriptions[name]
			if !ok {
				continue
			}
			exclude = false
			for _, excluded := range opts.ClusterDescriptions {
				if slice.ContainsString(clusterDesc, excluded, nil) {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
			log.Infof("Preparing creation of cluster %s for datacenter %s", name, clusterSpec.Cloud.DatacenterName)
			sem <- struct{}{} // acquire a slot
			wg.Add(1)
			go func(name string, project string, spec *kubermaticv1.ClusterSpec) {
				defer wg.Done()
				defer func() { <-sem }() // release the slot
				if !skipClusterCreation {
					ensureCluster(name, datacenterNameMappings[seedKey], project, spec)
				}
				if skipClusterCreation && updateClusters {
					// updateCluster(dc, spec)
				}
			}(name, legacyOpts.KubermaticProject, clusterSpec)
		}
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
	var wg sync.WaitGroup
	versionSlice := []string{}
	for _, v := range opts.Releases {
		versionSlice = append(versionSlice, v)
	}

	for seedKey := range defaultSeedSettings {
		for dc := range newClusters {
			wg.Add(1)
			go func(dc string) {
				defer wg.Done()
				cluster := &kubermaticv1.Cluster{}
				if !skipClusterDeletion {
					if !slice.ContainsString(versionSlice, cluster.Spec.Version.String(), nil) {
						return
					}
					err := runtimeOpts.SeedClusterClient.Get(rootCtx, apitypes.NamespacedName{Name: fmt.Sprintf("%s-%s", dc, seedKey)}, cluster)
					if err != nil {
						log.Errorf("Failed to get cluster %s: %v", dc, err)
						return
					}
					By("Deleting cluster for " + dc)

					userClusterClient, err := runtimeOpts.ClusterClientProvider.GetClient(rootCtx, cluster)
					if err != nil {
						log.Errorf("Failed to get user cluster client for cluster %s: %v", dc, err)
						return
					}
					By(fmt.Sprintf("Cleaning up resources for cluster %s. Name is %s", dc, cluster.Name))
					k8cginkgo.CommonCleanup(rootCtx, log, clients.NewKubeClient(legacyOpts), nil, userClusterClient, cluster)
				}

			}(dc)
		}
	}
	wg.Wait()

	By("Detaching datacenters from seed")
	seed := &kubermaticv1.Seed{}
	err := runtimeOpts.SeedClusterClient.Get(rootCtx, apitypes.NamespacedName{Name: "kubermatic", Namespace: "kubermatic"}, seed)
	if err != nil {
		log.Errorf("Failed to get seed 'kubermatic' for cleanup: %v", err)
	} else {
		for _, hashedName := range datacenterNameMappings {
			delete(seed.Spec.Datacenters, hashedName)
		}

		if err := runtimeOpts.SeedClusterClient.Update(rootCtx, seed); err != nil {
			log.Errorf("Failed to update seed 'kubermatic' to remove datacenters: %v", err)
		}
	}
})

var _ = ReportBeforeSuite(func(r Report) {})

var _ = ReportAfterSuite("ReportAfterSuite", func(r Report) {
	By("Reporting test results")
})
