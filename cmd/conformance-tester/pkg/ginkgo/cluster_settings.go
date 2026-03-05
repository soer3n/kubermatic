package ginkgo

import (
	"k8s.io/utils/ptr"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
)

// ClusterSpecModifier is a struct that holds a name and a modify function for a cluster spec.
type ClusterSpecModifier struct {
	Name   string
	Group  string // Modifiers with the same group name will be merged.
	Modify func(spec *kubermaticv1.ClusterSpec)
}

// CloudSpecModifier is a struct that holds a name and a modify function for a cloud spec.
type CloudSpecModifier struct {
	Name   string
	Group  string
	Modify func(spec *kubermaticv1.CloudSpec)
}

// ClusterSettings is now a slice of modifiers, each representing a distinct test case.
var ClusterSettings = []ClusterSpecModifier{
	{
		Name:  "with audit logging enabled",
		Group: "audit",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.AuditLogging = &kubermaticv1.AuditLoggingSettings{
				Enabled: true,
			}
		},
	},
	{
		Name:  "with user ssh key agent enabled",
		Group: "ssh",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.EnableUserSSHKeyAgent = ptr.To(true)
		},
	},
	// {
	// 	name:  "with oidc provider configured",
	// 	group: "oidc",
	// 	modify: func(spec *kubermaticv1.ClusterSpec) {
	// 		spec.OIDC = kubermaticv1.OIDCSettings{
	// 			IssuerURL: "https://test.com",
	// 			ClientID:  "test",
	// 		}
	// 	},
	// },
	{
		Name:  "with cni plugin set to canal",
		Group: "cni",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.CNIPlugin = &kubermaticv1.CNIPluginSettings{
				Type:    kubermaticv1.CNIPluginTypeCanal,
				Version: "v3.29",
			}
		},
	},
	{
		Name:  "with cni plugin set to cilium",
		Group: "cni",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.CNIPlugin = &kubermaticv1.CNIPluginSettings{
				Type:    kubermaticv1.CNIPluginTypeCilium,
				Version: "1.18.2",
			}
		},
	},
	{
		Name:  "with different update window",
		Group: "update-window",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.UpdateWindow = &kubermaticv1.UpdateWindow{
				Start:  "01:00",
				Length: "1h",
			}
		},
	},
}

var DefaultClusterSettings = kubermaticv1.Cluster{
	Spec: kubermaticv1.ClusterSpec{
		ContainerRuntime: "containerd",
		ExposeStrategy:   kubermaticv1.ExposeStrategyTunneling,
		ClusterNetwork: kubermaticv1.ClusterNetworkingConfig{
			ProxyMode: "iptables",
		},
	},
}
