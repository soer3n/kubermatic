package build

import (
	"context"

	"go.uber.org/zap"
	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/sdk/v2/semver"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/options"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/settings"
	"k8c.io/kubermatic/v2/pkg/version"
	"k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	"k8c.io/machine-controller/sdk/providerconfig"
	"k8c.io/machine-controller/sdk/providerconfig/configvar"
)

// clusterResult is used to pass data from a producer to a consumer.
type clusterResult struct {
	clusterName string
	dedupKey    string
	clusterSpec *kubermaticv1.ClusterSpec
	err         error
}

type clusterJob struct {
	combination    []settings.ClusterSpecModifier
	dcKey          string
	dcName         string                  // hashed datacenter name in the seed
	datacenter     kubermaticv1.Datacenter // resolved datacenter config
	seed           kubermaticv1.Seed
	kubeVersion    *version.Version
	log            *zap.SugaredLogger
	rootCtx        context.Context
	opts           *options.Options
	kkpConfig      *kubermaticv1.KubermaticConfiguration
	versionManager *version.Manager
	providerConfig *providerconfig.Config
}

// scenarioResult is used to pass data from a producer to a consumer.
type scenarioResult struct {
	clusterKey  string
	machineName string
	dedupKey    string
	machineSpec v1alpha1.MachineSpec
	err         error
}

type Scenario struct {
	ClusterName  string
	ScenarioName string
	Distribution string
	Description  string
	ProjectName  string
	Exclude      bool
	ClusterSpec  *kubermaticv1.ClusterSpec
	Machines     map[string]v1alpha1.MachineSpec
}

type scenarioJob struct {
	combination             []settings.MachineSpecModifier[any]
	clusterKey              string
	version                 semver.Semver
	log                     *zap.SugaredLogger
	rootCtx                 context.Context
	resolver                *configvar.Resolver
	opts                    *options.Options
	providerConfig          *providerconfig.Config
	distribution            providerconfig.OperatingSystem
	cachedProviderSpecBytes []byte // pre-computed provider spec JSON to avoid repeated API calls
}
