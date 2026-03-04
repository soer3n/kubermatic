package build

import (
	"context"
	"flag"
	"iter"
	"maps"
	"os"

	"go.uber.org/zap"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/options"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	"k8c.io/machine-controller/sdk/providerconfig"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

type ProviderSettings struct {
	Clusters iter.Seq[string]
	Machines iter.Seq[string]
}

func GetGinkgoLabels(log *zap.SugaredLogger, provider string) (ProviderSettings, error) {
	var err error
	var (
		rootCtx     context.Context
		opts        *options.Options
		runtimeOpts *options.RuntimeOptions
		legacyOpts  *legacytypes.Options
	)

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

	// infraClient = client
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
	returnProviderSettings := ProviderSettings{}

	scenarios, _, _, _, _, _ := GetTableEntries(context.Background(), log, runtimeOpts, legacyOpts, opts, client, "dummy", providerconfig.CloudProvider(provider))
	includedScenarios := make(map[string]*Scenario)
	includedMachineSpecs := make(map[string]string)
	for k, v := range scenarios {
		if !v.Exclude {
			includedScenarios[k] = v
			if includedMachineSpecs == nil {
				includedMachineSpecs = make(map[string]string)
			}
			if len(includedMachineSpecs) == 0 {
				for machineName, machineSpec := range v.Machines {
					includedMachineSpecs[machineName] = machineSpec.Versions.Kubelet
				}
			}
		}
	}
	returnProviderSettings = ProviderSettings{
		Clusters: maps.Keys(includedScenarios),
		Machines: maps.Keys(includedMachineSpecs),
	}
	return returnProviderSettings, nil
}
