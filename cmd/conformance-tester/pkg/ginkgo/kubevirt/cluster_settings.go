package kubevirt

import (
	"fmt"

	"github.com/aws/smithy-go/ptr"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

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

var clusterSettings = map[string]kubermaticv1.ClusterSpec{
	"with default settings": {},
	"with human readable name": {
		HumanReadableName: "my-cluster",
	},
	"with human readable name alt": {
		HumanReadableName: "alt-cluster",
	},
	"with container runtime set to containerd": {
		ContainerRuntime: "containerd",
	},
	"with image pull secret": {
		ImagePullSecret: &v1.SecretReference{Name: "my-secret"},
	},
	"with image pull secret alt": {
		ImagePullSecret: &v1.SecretReference{Name: "alt-secret"},
	},
	"with cni plugin set to canal": {
		CNIPlugin: &kubermaticv1.CNIPluginSettings{Type: "canal", Version: "v3.29"},
	},
	"with cni plugin set to cilium": {
		CNIPlugin: &kubermaticv1.CNIPluginSettings{Type: "cilium", Version: "1.16.9"},
	},
	"with cluster network IPv4": {
		ClusterNetwork: kubermaticv1.ClusterNetworkingConfig{IPFamily: "IPv4"},
	},
	"with cluster network dual stack": {
		ClusterNetwork: kubermaticv1.ClusterNetworkingConfig{IPFamily: "IPv4+IPv6"},
	},
	"with machine networks": {
		MachineNetworks: []kubermaticv1.MachineNetworkingConfig{{CIDR: "192.168.1.0/24"}},
	},
	"with machine networks alt": {
		MachineNetworks: []kubermaticv1.MachineNetworkingConfig{{CIDR: "10.0.0.0/16"}},
	},
	"with expose strategy set to NodePort": {
		ExposeStrategy: kubermaticv1.ExposeStrategyNodePort,
	},
	"with expose strategy set to LoadBalancer": {
		ExposeStrategy: kubermaticv1.ExposeStrategyLoadBalancer,
	},
	"with api server allowed ip ranges": {
		APIServerAllowedIPRanges: &kubermaticv1.NetworkRanges{CIDRBlocks: []string{"0.0.0.0/0"}},
	},
	"with api server allowed ip ranges alt": {
		APIServerAllowedIPRanges: &kubermaticv1.NetworkRanges{CIDRBlocks: []string{"10.0.0.0/8"}},
	},
	"with components override": {
		ComponentsOverride: kubermaticv1.ComponentSettings{},
	},
	"with oidc settings": {
		OIDC: kubermaticv1.OIDCSettings{IssuerURL: "https://issuer.example.com"},
	},
	"with features": {
		Features: map[string]bool{"externalCloudProvider": true},
	},
	"with update window": {
		UpdateWindow: &kubermaticv1.UpdateWindow{Start: "Mon 21:00", Length: "2h"},
	},
	"with use pod security policy admission plugin": {
		UsePodSecurityPolicyAdmissionPlugin: true,
	},
	"with use pod node selector admission plugin": {
		UsePodNodeSelectorAdmissionPlugin: true,
	},
	"with use event rate limit admission plugin": {
		UseEventRateLimitAdmissionPlugin: true,
	},
	"with admission plugins": {
		AdmissionPlugins: []string{"NamespaceLifecycle"},
	},
	"with pod node selector admission plugin config": {
		PodNodeSelectorAdmissionPluginConfig: map[string]string{"clusterDefaultNodeSelector": "role=worker"},
	},
	"with event rate limit config": {
		EventRateLimitConfig: &kubermaticv1.EventRateLimitConfig{},
	},
	"with enable user ssh key agent set to true": {
		EnableUserSSHKeyAgent: ptr.Bool(true),
	},
	"with kubelb": {
		KubeLB: &kubermaticv1.KubeLB{Enabled: true},
	},
	"with kubernetes dashboard": {
		KubernetesDashboard: &kubermaticv1.KubernetesDashboard{Enabled: true},
	},
	"with audit logging enabled": {
		AuditLogging: &kubermaticv1.AuditLoggingSettings{Enabled: true},
	},
	"with opa integration enabled": {
		OPAIntegration: &kubermaticv1.OPAIntegrationSettings{Enabled: true},
	},
	"with opa integration disabled": {
		OPAIntegration: &kubermaticv1.OPAIntegrationSettings{Enabled: false},
	},
	"with mla monitoring enabled": {
		MLA: &kubermaticv1.MLASettings{MonitoringEnabled: true, LoggingEnabled: false},
	},
	"with mla logging enabled": {
		MLA: &kubermaticv1.MLASettings{MonitoringEnabled: false, LoggingEnabled: true},
	},
	"with application settings cache size": {
		ApplicationSettings: &kubermaticv1.ApplicationSettings{CacheSize: func() *resource.Quantity { q := resource.MustParse("10Gi"); return &q }()},
	},
	"with application settings cache size alt": {
		ApplicationSettings: &kubermaticv1.ApplicationSettings{CacheSize: func() *resource.Quantity { q := resource.MustParse("20Gi"); return &q }()},
	},
	"with encryption configuration enabled": {
		EncryptionConfiguration: &kubermaticv1.EncryptionConfiguration{Enabled: true, Resources: []string{"secrets"}},
	},
	"with encryption configuration disabled": {
		EncryptionConfiguration: &kubermaticv1.EncryptionConfiguration{Enabled: false, Resources: []string{"configmaps"}},
	},
	"with pause true": {
		Pause: true,
	},
	"with pause false": {
		Pause: false,
	},
	"with debug log true": {
		DebugLog: true,
	},
	"with debug log false": {
		DebugLog: false,
	},
	"with disable csi driver true": {
		DisableCSIDriver: true,
	},
	"with disable csi driver false": {
		DisableCSIDriver: false,
	},
	"with backup config": {
		BackupConfig: &kubermaticv1.BackupConfig{},
	},
	"with kyverno": {
		Kyverno: &kubermaticv1.KyvernoSettings{Enabled: true},
	},
	"with authorization config": {
		AuthorizationConfig: &kubermaticv1.AuthorizationConfig{EnabledModes: []string{"Node", "RBAC"}},
	},
	"with container runtime opts": {
		ContainerRuntimeOpts: &kubermaticv1.ContainerRuntimeOpts{},
	},
}

var defaultClusterSettings = kubermaticv1.Cluster{
	Spec: kubermaticv1.ClusterSpec{
		ContainerRuntime: "containerd",
		ExposeStrategy:   kubermaticv1.ExposeStrategyLoadBalancer,
	},
}
