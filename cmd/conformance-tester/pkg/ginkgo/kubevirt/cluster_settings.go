package kubevirt

import (
	"fmt"

	"go.uber.org/zap"
	"k8s.io/utils/ptr"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/pkg/defaulting"
)

func loadKubermaticConfiguration() (*kubermaticv1.KubermaticConfiguration, error) {
	config := &kubermaticv1.KubermaticConfiguration{}
	defaulted, err := defaulting.DefaultConfiguration(config, zap.NewNop().Sugar())
	if err != nil {
		return nil, fmt.Errorf("failed to process: %w", err)
	}

	return defaulted, nil
}

// clusterSpecModifier is a struct that holds a name and a modify function for a cluster spec.
type clusterSpecModifier struct {
	name   string
	group  string // Modifiers with the same group name will be merged.
	modify func(spec *kubermaticv1.ClusterSpec)
}

// clusterSettings is now a slice of modifiers, each representing a distinct test case.
var clusterSettings = []clusterSpecModifier{
	{
		name:  "with audit logging enabled",
		group: "audit",
		modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.AuditLogging = &kubermaticv1.AuditLoggingSettings{
				Enabled: true,
			}
		},
	},
	{
		name:  "with user ssh key agent enabled",
		group: "ssh",
		modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.EnableUserSSHKeyAgent = ptr.To(true)
		},
	},
	{
		name:  "with oidc provider configured",
		group: "oidc",
		modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.OIDC = kubermaticv1.OIDCSettings{
				IssuerURL: "https://test.com",
				ClientID:  "test",
			}
		},
	},
	{
		name:  "with cni plugin set to canal",
		group: "cni",
		modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.CNIPlugin = &kubermaticv1.CNIPluginSettings{
				Type:    kubermaticv1.CNIPluginTypeCanal,
				Version: "v3.29",
			}
		},
	},
	{
		name:  "with cni plugin set to cilium",
		group: "cni",
		modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.CNIPlugin = &kubermaticv1.CNIPluginSettings{
				Type:    kubermaticv1.CNIPluginTypeCilium,
				Version: "1.18.2",
			}
		},
	},
	{
		name:  "with different update window",
		group: "update-window",
		modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.UpdateWindow = &kubermaticv1.UpdateWindow{
				Start:  "01:00",
				Length: "1h",
			}
		},
	},
}

var defaultClusterSettings = kubermaticv1.Cluster{
	Spec: kubermaticv1.ClusterSpec{
		ContainerRuntime: "containerd",
		ExposeStrategy:   kubermaticv1.ExposeStrategyLoadBalancer,
	},
}
