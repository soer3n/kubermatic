package kubevirt

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/go-logr/zapr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubectl/pkg/util/slice"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlruntimelog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/clients"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/utils"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	kubermaticlog "k8c.io/kubermatic/v2/pkg/log"
	kkpreconciling "k8c.io/kubermatic/v2/pkg/resources/reconciling"

	"k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"

	apitypes "k8s.io/apimachinery/pkg/types"

	"k8c.io/machine-controller/sdk/cloudprovider/kubevirt"
	"k8c.io/machine-controller/sdk/providerconfig"
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
	datacenters         string
	kubeVersions        string
	skipClusterDeletion bool
	skipClusterCreation bool
	updateClusters      bool
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
	seed                     *kubermaticv1.Seed
)

func TestMain(m *testing.M) {
	var err error

	// // step 1
	// versions := utils.GetReleaseVersions()
	// log.Infof("Available Kubernetes versions: %v", versions)
	// step 2
	datacenters := GetDatacenterDescriptions()
	log.Infof("Available datacenter descriptions: %v", datacenters)
	// step 3
	clusters := utils.GetClusterDescriptions()
	log.Infof("Available cluster descriptions: %v", clusters)
	// step 4
	machines := GetMachineDescriptions()
	log.Infof("Available machine descriptions: %v", machines)
	// step 5
	_ = k8cginkgo.ResourceSettings{}

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

	flag.Parse()

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
	opts.Tests = []string{}
	opts.EnableTests = []string{}
	opts.ExcludeTests = []string{}
	for _, test := range testSlice {
		opts.EnableTests = append(opts.EnableTests, test)
		opts.Tests = append(opts.Tests, test)
	}
	legacyOpts = k8cginkgo.MergeOptions(log, opts, legacyOpts, runtimeOpts)
	legacyOpts.Providers = sets.Set[string]{"kubevirt": {}}
	if err := legacyOpts.ParseFlags(log); err != nil {
		log.Warnf("Invalid flags", zap.Error(err))
	}

	log.Infof("Included datacenter descriptions: %v", opts.Included.DatacenterDescriptions)
	log.Infof("Excluded datacenter descriptions: %v", opts.Excluded.DatacenterDescriptions)
	log.Infof("Included cluster descriptions: %v", opts.Included.ClusterDescriptions)
	log.Infof("Excluded cluster descriptions: %v", opts.Excluded.ClusterDescriptions)
	log.Infof("Included machine descriptions: %v", opts.Included.MachineDescriptions)
	log.Infof("Excluded machine descriptions: %v", opts.Excluded.MachineDescriptions)

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
		err := runtimeOpts.SeedClusterClient.Update(rootCtx, seed)
		Expect(err).NotTo(HaveOccurred(), "Failed to update seed")
	})

	suiteCfg, reporterCfg := GinkgoConfiguration()
	By(fmt.Sprintf("parallel=%d", suiteCfg.ParallelTotal))
	By(fmt.Sprintf("Reporter: %#v", reporterCfg))
	By(fmt.Sprintf("Creating clusters for datacenters and kube versions: %v", maps.Keys(newClusters)))

	var wg sync.WaitGroup
	maxConcurrent := 4 // Set your desired concurrency limit
	sem := make(chan struct{}, maxConcurrent)
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
	for i, _ := range defaultSeedSettings {
		log.Infof("defaultSeedSettings[%d]: %+v", i, datacenterNameMappings[i])
	}
	for i, _ := range newClusters {
		log.Infof("newClusters[%d]: %+v", i, finalClusterDescriptions[i])
	}
	for seedKey := range defaultSeedSettings {
		exclude := false
		if len(opts.Included.DatacenterDescriptions) > 0 {
			exclude = true
			for _, included := range opts.Included.DatacenterDescriptions {
				if strings.Contains(seedKey, included) {
					exclude = false
					break
				}
			}
		} else {
			for _, excluded := range opts.Excluded.DatacenterDescriptions {
				if strings.Contains(seedKey, excluded) {
					exclude = true
					break
				}
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
			if len(opts.Included.ClusterDescriptions) > 0 {
				exclude = true
				for _, included := range opts.Included.ClusterDescriptions {
					if slice.ContainsString(clusterDesc, included, nil) {
						exclude = false
						break
					}
				}
			} else {
				for _, excluded := range opts.Excluded.ClusterDescriptions {
					if slice.ContainsString(clusterDesc, excluded, nil) {
						exclude = true
						break
					}
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
	if len(opts.Releases) > 0 {
		for _, v := range opts.Releases {
			versionSlice = append(versionSlice, v)
		}
	} else {
		for _, scenario := range kkpConfig.Spec.Versions.Versions {
			versionSlice = append(versionSlice, scenario.String())
		}
	}

	if !skipClusterDeletion {
		for seedKey := range defaultSeedSettings {
			for dc := range newClusters {
				wg.Add(1)
				go func(dc string) {
					defer wg.Done()
					cluster := &kubermaticv1.Cluster{}

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
	}

})

var _ = ReportBeforeSuite(func(r Report) {})

var _ = ReportAfterSuite("ReportAfterSuite", func(r Report) {
	By("Reporting test results")
})
