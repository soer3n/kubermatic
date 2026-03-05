package settings

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

// ClusterSettings is now a slice of modifiers, each representing a distinct test case.
var ClusterSettings = []ClusterSpecModifier{
	// --- CNI ---
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
	// --- Expose Strategy ---
	{
		Name:  "with expose strategy set to NodePort",
		Group: "expose-strategy",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.ExposeStrategy = kubermaticv1.ExposeStrategyNodePort
		},
	},
	{
		Name:  "with expose strategy set to LoadBalancer",
		Group: "expose-strategy",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.ExposeStrategy = kubermaticv1.ExposeStrategyLoadBalancer
		},
	},
	{
		Name:  "with expose strategy set to Tunneling",
		Group: "expose-strategy",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.ExposeStrategy = kubermaticv1.ExposeStrategyTunneling
		},
	},
	// --- Proxy Mode ---
	{
		Name:  "with proxy mode set to ipvs",
		Group: "proxy-mode",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.ClusterNetwork.ProxyMode = "ipvs"
		},
	},
	{
		Name:  "with proxy mode set to iptables",
		Group: "proxy-mode",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.ClusterNetwork.ProxyMode = "iptables"
		},
	},
	{
		Name:  "with proxy mode set to ebpf",
		Group: "proxy-mode",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.ClusterNetwork.ProxyMode = "ebpf"
		},
	},
	// --- Audit Logging ---
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
		Name:  "with audit logging disabled",
		Group: "audit",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.AuditLogging = &kubermaticv1.AuditLoggingSettings{
				Enabled: false,
			}
		},
	},
	// --- SSH Key Agent ---
	{
		Name:  "with user ssh key agent enabled",
		Group: "ssh",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.EnableUserSSHKeyAgent = ptr.To(true)
		},
	},
	{
		Name:  "with user ssh key agent disabled",
		Group: "ssh",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.EnableUserSSHKeyAgent = ptr.To(false)
		},
	},
	// --- OPA Integration ---
	{
		Name:  "with opa integration enabled",
		Group: "opa",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.OPAIntegration = &kubermaticv1.OPAIntegrationSettings{
				Enabled: true,
			}
		},
	},
	{
		Name:  "with opa integration disabled",
		Group: "opa",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.OPAIntegration = &kubermaticv1.OPAIntegrationSettings{
				Enabled: false,
			}
		},
	},
	// --- MLA Monitoring ---
	{
		Name:  "with mla monitoring enabled",
		Group: "mla-monitoring",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			if spec.MLA == nil {
				spec.MLA = &kubermaticv1.MLASettings{}
			}
			spec.MLA.MonitoringEnabled = true
		},
	},
	{
		Name:  "with mla monitoring disabled",
		Group: "mla-monitoring",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			if spec.MLA == nil {
				spec.MLA = &kubermaticv1.MLASettings{}
			}
			spec.MLA.MonitoringEnabled = false
		},
	},
	// --- MLA Logging ---
	{
		Name:  "with mla logging enabled",
		Group: "mla-logging",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			if spec.MLA == nil {
				spec.MLA = &kubermaticv1.MLASettings{}
			}
			spec.MLA.LoggingEnabled = true
		},
	},
	{
		Name:  "with mla logging disabled",
		Group: "mla-logging",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			if spec.MLA == nil {
				spec.MLA = &kubermaticv1.MLASettings{}
			}
			spec.MLA.LoggingEnabled = false
		},
	},
	// --- Node Local DNS Cache ---
	{
		Name:  "with node local dns cache enabled",
		Group: "node-local-dns",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.ClusterNetwork.NodeLocalDNSCacheEnabled = ptr.To(true)
		},
	},
	{
		Name:  "with node local dns cache disabled",
		Group: "node-local-dns",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.ClusterNetwork.NodeLocalDNSCacheEnabled = ptr.To(false)
		},
	},
	// --- Kubernetes Dashboard ---
	{
		Name:  "with kubernetes dashboard enabled",
		Group: "k8s-dashboard",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.KubernetesDashboard = &kubermaticv1.KubernetesDashboard{
				Enabled: true,
			}
		},
	},
	{
		Name:  "with kubernetes dashboard disabled",
		Group: "k8s-dashboard",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.KubernetesDashboard = &kubermaticv1.KubernetesDashboard{
				Enabled: false,
			}
		},
	},
	// --- PodNodeSelector Admission Plugin ---
	{
		Name:  "with pod node selector admission plugin enabled",
		Group: "pod-node-selector",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.UsePodNodeSelectorAdmissionPlugin = true
		},
	},
	{
		Name:  "with pod node selector admission plugin disabled",
		Group: "pod-node-selector",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.UsePodNodeSelectorAdmissionPlugin = false
		},
	},
	// --- EventRateLimit Admission Plugin ---
	{
		Name:  "with event rate limit admission plugin enabled",
		Group: "event-rate-limit",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.UseEventRateLimitAdmissionPlugin = true
		},
	},
	{
		Name:  "with event rate limit admission plugin disabled",
		Group: "event-rate-limit",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.UseEventRateLimitAdmissionPlugin = false
		},
	},
	// --- CSI Driver ---
	{
		Name:  "with csi driver enabled",
		Group: "csi-driver",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.DisableCSIDriver = false
		},
	},
	{
		Name:  "with csi driver disabled",
		Group: "csi-driver",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.DisableCSIDriver = true
		},
	},
	// --- Update Window ---
	{
		Name:  "with update window configured",
		Group: "update-window",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.UpdateWindow = &kubermaticv1.UpdateWindow{
				Start:  "01:00",
				Length: "1h",
			}
		},
	},
	{
		Name:  "with no update window",
		Group: "update-window",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.UpdateWindow = nil
		},
	},
	// --- OIDC ---
	// {
	// 	Name:  "with oidc authentication enabled",
	// 	Group: "oidc",
	// 	Modify: func(spec *kubermaticv1.ClusterSpec) {
	// 		spec.OIDC = kubermaticv1.OIDCSettings{
	// 			IssuerURL: "https://dex.example.com",
	// 			ClientID:  "conformance-test",
	// 		}
	// 	},
	// },
	{
		Name:  "with oidc authentication disabled",
		Group: "oidc",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			spec.OIDC = kubermaticv1.OIDCSettings{}
		},
	},
	// --- External Cloud Provider ---
	{
		Name:  "with external cloud provider enabled",
		Group: "external-ccm",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			if spec.Features == nil {
				spec.Features = map[string]bool{}
			}
			spec.Features[kubermaticv1.ClusterFeatureExternalCloudProvider] = true
		},
	},
	{
		Name:  "with external cloud provider disabled",
		Group: "external-ccm",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			if spec.Features == nil {
				spec.Features = map[string]bool{}
			}
			spec.Features[kubermaticv1.ClusterFeatureExternalCloudProvider] = false
		},
	},
	// --- IPVS Strict ARP ---
	{
		Name:  "with ipvs strict arp enabled",
		Group: "ipvs-strict-arp",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			if spec.ClusterNetwork.IPVS == nil {
				spec.ClusterNetwork.IPVS = &kubermaticv1.IPVSConfiguration{}
			}
			spec.ClusterNetwork.IPVS.StrictArp = ptr.To(true)
		},
	},
	{
		Name:  "with ipvs strict arp disabled",
		Group: "ipvs-strict-arp",
		Modify: func(spec *kubermaticv1.ClusterSpec) {
			if spec.ClusterNetwork.IPVS == nil {
				spec.ClusterNetwork.IPVS = &kubermaticv1.IPVSConfiguration{}
			}
			spec.ClusterNetwork.IPVS.StrictArp = ptr.To(false)
		},
	},
}

var DefaultClusterSettings = kubermaticv1.Cluster{
	Spec: kubermaticv1.ClusterSpec{
		ContainerRuntime: "containerd",
		ExposeStrategy:   kubermaticv1.ExposeStrategyTunneling,
	},
}
