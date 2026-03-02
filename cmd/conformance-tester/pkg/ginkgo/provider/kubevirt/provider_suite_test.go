package kubevirt

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/go-logr/zapr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/clientcmd"
	ctrlruntimelog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/build"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/options"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/utils"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	kubermaticlog "k8c.io/kubermatic/v2/pkg/log"
	kkpreconciling "k8c.io/kubermatic/v2/pkg/resources/reconciling"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
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
	opts                     *options.Options
	runtimeOpts              *options.RuntimeOptions
	legacyOpts               *legacytypes.Options
	infraClient              ctrlclient.Client
	kkpConfig                *kubermaticv1.KubermaticConfiguration
	newClusters              map[string]*kubermaticv1.ClusterSpec
	defaultSeedSettings      map[string]kubermaticv1.Seed
	finalClusterDescriptions map[string][]string
	datacenterNameMappings   map[string]string
	seed                     *kubermaticv1.Seed
	entries                  map[string]*build.Scenario
	projectName              string
)

func TestMain(m *testing.M) {
	var err error
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	opts, err = options.NewOptionsFromYAML(log)
	if err != nil {
		log.Fatalw("Failed to load options", zap.Error(err))
	}

	config, err := clientcmd.BuildConfigFromFlags("", opts.Secrets.Kubevirt.KubeconfigFile)
	if err != nil {
		panic(err)
	}

	client, err := ctrlclient.New(config, ctrlclient.Options{
		Scheme: scheme,
	})
	if err != nil {
		panic(err)
	}

	infraClient = client
	rootCtx = signals.SetupSignalHandler()

	configPath := os.Getenv("CONFORMANCE_TESTER_CONFIG_FILE")
	if configPath != "" {
		runtimeOpts, err = options.NewRuntimeOptions(rootCtx, log, opts)
		if err != nil {
			log.Fatalw("Failed to create runtime options", zap.Error(err))
		}
	}
	legacyOpts = legacytypes.NewDefaultOptions()
	legacyOpts.AddFlags()

	flag.Parse()

	if configPath == "" {
		runtimeOpts, _ = options.NewRuntimeOptions(rootCtx, log, &options.Options{
			KubermaticNamespace: legacyOpts.KubermaticNamespace,
			KubermaticSeedName:  legacyOpts.KubermaticSeedName,
		})
	}
	testSlice := []string{
		legacytypes.StorageTests, legacytypes.LoadbalancerTests, legacytypes.MetricsTests,
		legacytypes.UserClusterRBACTests, legacytypes.K8sGcrImageTests, legacytypes.MetricsTests,
		legacytypes.SecurityContextTests, legacytypes.UserClusterSeccompTests, legacytypes.UserClusterK8sGcrImageTests,
	}
	// enable all tests
	opts.Tests = []string{}
	opts.EnableTests = []string{}
	opts.ExcludeTests = []string{}
	for _, test := range testSlice {
		opts.EnableTests = append(opts.EnableTests, test)
		opts.Tests = append(opts.Tests, test)
	}
	legacyOpts = options.MergeOptions(log, opts, legacyOpts, runtimeOpts)
	legacyOpts.Providers = sets.Set[string]{"kubevirt": {}}
	if err := legacyOpts.ParseFlags(log); err != nil {
		log.Warnf("Invalid flags", zap.Error(err))
	}

	projectName = legacyOpts.KubermaticProject
	if legacyOpts.KubermaticProject == "" {
		projectName = "e2e-" + rand.String(5)
	}

	log.Infof("Included datacenter descriptions: %v", opts.Included.DatacenterDescriptions)
	log.Infof("Excluded datacenter descriptions: %v", opts.Excluded.DatacenterDescriptions)
	log.Infof("Included cluster descriptions: %v", opts.Included.ClusterDescriptions)
	log.Infof("Excluded cluster descriptions: %v", opts.Excluded.ClusterDescriptions)
	log.Infof("Included machine descriptions: %v", opts.Included.MachineDescriptions)
	log.Infof("Excluded machine descriptions: %v", opts.Excluded.MachineDescriptions)

	os.Exit(m.Run())
}

func TestScenarios(t *testing.T) {
	RegisterFailHandler(CustomFail)
	RunSpecs(t, "Conformance Tester Scenarios Suite KubeVirt")
}

func CustomFail(message string, callerSkip ...int) {
	log.Infof("Fail called: %+v", message)
	Fail(message, callerSkip...)
}

var _ = SynchronizedBeforeSuite(func() {
	defer GinkgoRecover()
	By("Preparing test environment and creating clusters")
	utils.PrepareSuite(rootCtx, log, *legacyOpts, runtimeOpts, *opts, seed, entries, kkpConfig, projectName, defaultSeedSettings, newClusters, finalClusterDescriptions, datacenterNameMappings, skipClusterCreation, updateClusters)
}, func(data []byte) {
	// Assign cluster name on every node
})

// Old SynchronizedAfterSuite replaced with aggregated variant.
var _ = SynchronizedAfterSuite(func() {
	// per-node no-op (could emit per-node logs here)
}, func() {
	defer GinkgoRecover()
	By("Cleaning up created clusters")
	utils.PostProcessingSuite(rootCtx, log, *legacyOpts, runtimeOpts, *opts, seed, entries, kkpConfig, projectName, defaultSeedSettings, newClusters, finalClusterDescriptions, datacenterNameMappings, skipClusterCreation, updateClusters, skipClusterDeletion)
})

var _ = ReportBeforeSuite(func(r Report) {})

var _ = ReportAfterSuite("ReportAfterSuite", func(r Report) {
})
