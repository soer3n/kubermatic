package kubevirt

import (
	"encoding/json"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"

	"k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	"k8c.io/machine-controller/sdk/cloudprovider/kubevirt"
	"k8c.io/machine-controller/sdk/providerconfig"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

// machineSpecModifier is a struct that holds a name and a modify function for a machine spec.
type machineSpecModifier struct {
	name   string
	group  string // Modifiers with the same group name will be merged.
	modify func(spec *kubevirt.RawConfig)
}

var machineSettings = []machineSpecModifier{
	{
		name:  "with primary disk OS image from an HTTP source",
		group: "os-image",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Template.PrimaryDisk.Source.Value = "http"
			// This assumes some default URL is set elsewhere, as the model doesn't have a dedicated URL field anymore.
		},
	},
	{
		name:  "with primary disk OS image from a container",
		group: "os-image",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Template.PrimaryDisk.OsImage.Value = "container"
			// This assumes some default image is set elsewhere.
		},
	},
	{
		name:  "with custom cpu and memory",
		group: "resources",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Instancetype = nil
			spec.VirtualMachine.Preference = nil
			spec.VirtualMachine.Template.CPUs.Value = "4"
			spec.VirtualMachine.Template.Memory.Value = "8Gi"
		},
	},
	{
		name:  "with a secondary disk",
		group: "storage",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Template.SecondaryDisks = append(spec.VirtualMachine.Template.SecondaryDisks, kubevirt.SecondaryDisks{
				Disk: kubevirt.Disk{
					Size:             providerconfig.ConfigVarString{Value: "10Gi"},
					StorageClassName: providerconfig.ConfigVarString{Value: "kubermatic-fast"},
				},
			})
		},
	},
	{
		name:  "with a different primary disk size",
		group: "storage",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Template.PrimaryDisk.Size.Value = "25Gi"
			spec.VirtualMachine.Template.PrimaryDisk.StorageClassName = providerconfig.ConfigVarString{Value: "kubermatic-fast"}
		},
	},
	// New modifiers converted from old format
	{
		name:  "with changed cluster name",
		group: "cluster-name",
		modify: func(spec *kubevirt.RawConfig) {
			spec.ClusterName.Value = "changed-cluster-name"
		},
	},
	{
		name:  "with auth kubeconfig set",
		group: "auth",
		modify: func(spec *kubevirt.RawConfig) {
			spec.Auth.Kubeconfig.Value = "valid-kubeconfig"
		},
	},
	{
		name:  "with empty instancetype",
		group: "instancetype",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Instancetype = &kubevirtv1.InstancetypeMatcher{Name: ""}
		},
	},
	{
		name:  "with empty preference",
		group: "preference",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Preference = &kubevirtv1.PreferenceMatcher{Name: ""}
		},
	},
	{
		name:  "with dns policy set to ClusterFirstWithHostNet",
		group: "dns-policy",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.DNSPolicy.Value = "ClusterFirstWithHostNet"
		},
	},
	{
		name:  "with dns policy set to ClusterFirst",
		group: "dns-policy",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.DNSPolicy.Value = "ClusterFirst"
		},
	},
	{
		name:  "with dns policy set to Default",
		group: "dns-policy",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.DNSPolicy.Value = "Default"
		},
	},
	{
		name:  "with dns policy set to None",
		group: "dns-policy",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.DNSPolicy.Value = "None"
			spec.VirtualMachine.DNSConfig = &v1.PodDNSConfig{
				Nameservers: []string{"8.8.8.8", "8.8.4.4"},
			}
		},
	},
	{
		name:  "with eviction strategy set to LiveMigrate",
		group: "eviction-strategy",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.EvictionStrategy = "LiveMigrate"
		},
	},
	{
		name:  "with eviction strategy set to External",
		group: "eviction-strategy",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.EvictionStrategy = "External"
		},
	},
	{
		name:  "with network multi-queue enabled",
		group: "multi-queue",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.EnableNetworkMultiQueue.Value = ptr.To(true)
		},
	},
	{
		name:  "with network multi-queue disabled",
		group: "multi-queue",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.EnableNetworkMultiQueue.Value = ptr.To(false)
		},
	},
	{
		name:  "with 4 CPUs",
		group: "cpu",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Template.CPUs.Value = "4"
		},
	},
	{
		name:  "with 2 vCPUs",
		group: "vcpu",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Template.VCPUs.Cores = 2
		},
	},
	{
		name:  "with 4096Mi memory",
		group: "memory",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Template.Memory.Value = "4096Mi"
		},
	},
	{
		name:  "with primary disk storage class set to standard",
		group: "primary-disk-sc",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Template.PrimaryDisk.StorageClassName.Value = "standard"
		},
	},
	{
		name:  "with primary disk storage class set to kubermatic-fast",
		group: "primary-disk-sc",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Template.PrimaryDisk.StorageClassName.Value = "kubermatic-fast"
		},
	},
	{
		name:  "with secondary disk storage class set to standard",
		group: "secondary-disk-sc",
		modify: func(spec *kubevirt.RawConfig) {
			if len(spec.VirtualMachine.Template.SecondaryDisks) == 0 {
				spec.VirtualMachine.Template.SecondaryDisks = append(spec.VirtualMachine.Template.SecondaryDisks, kubevirt.SecondaryDisks{
					Disk: kubevirt.Disk{Size: providerconfig.ConfigVarString{Value: "10Gi"}},
				})
			}
			spec.VirtualMachine.Template.SecondaryDisks[0].StorageClassName.Value = "standard"
		},
	},
	{
		name:  "with secondary disk storage class set to kubermatic-fast",
		group: "secondary-disk-sc",
		modify: func(spec *kubevirt.RawConfig) {
			if len(spec.VirtualMachine.Template.SecondaryDisks) == 0 {
				spec.VirtualMachine.Template.SecondaryDisks = append(spec.VirtualMachine.Template.SecondaryDisks, kubevirt.SecondaryDisks{
					Disk: kubevirt.Disk{Size: providerconfig.ConfigVarString{Value: "10Gi"}},
				})
			}
			spec.VirtualMachine.Template.SecondaryDisks[0].StorageClassName.Value = "kubermatic-fast"
		},
	},
	{
		name:  "with node affinity for hostname node-01",
		group: "node-affinity",
		modify: func(spec *kubevirt.RawConfig) {
			spec.Affinity.NodeAffinityPreset.Key.Value = "kubernetes.io/hostname"
			spec.Affinity.NodeAffinityPreset.Values = []providerconfig.ConfigVarString{{Value: "node-01"}}
		},
	},
	{
		name:  "with node affinity preset key",
		group: "node-affinity-key",
		modify: func(spec *kubevirt.RawConfig) {
			spec.Affinity.NodeAffinityPreset.Key.Value = "kubernetes.io/hostname"
		},
	},
	{
		name:  "with node affinity preset values",
		group: "node-affinity-values",
		modify: func(spec *kubevirt.RawConfig) {
			spec.Affinity.NodeAffinityPreset.Values = []providerconfig.ConfigVarString{{Value: "node-01"}}
		},
	},
	{
		name:  "with empty node affinity preset type",
		group: "node-affinity-type",
		modify: func(spec *kubevirt.RawConfig) {
			spec.Affinity.NodeAffinityPreset.Type.Value = ""
		},
	},
	{
		name:  "with topology spread constraint on hostname",
		group: "topology-spread",
		modify: func(spec *kubevirt.RawConfig) {
			spec.TopologySpreadConstraints = []kubevirt.TopologySpreadConstraint{{
				TopologyKey: providerconfig.ConfigVarString{Value: "kubernetes.io/hostname"},
				MaxSkew:     providerconfig.ConfigVarString{Value: "1"},
			}}
		},
	},
	{
		name:  "with topology spread constraint set to DoNotSchedule",
		group: "topology-spread",
		modify: func(spec *kubevirt.RawConfig) {
			spec.TopologySpreadConstraints = []kubevirt.TopologySpreadConstraint{{
				WhenUnsatisfiable: providerconfig.ConfigVarString{Value: "DoNotSchedule"},
				MaxSkew:           providerconfig.ConfigVarString{Value: "1"},
			}}
		},
	},
}

// getDefaultMachineSpec returns the default machine spec.
func getDefaultMachineSpec() (*v1alpha1.MachineSpec, error) {
	cfg, err := getDefaultKubevirtConfig()
	if err != nil {
		return nil, err
	}

	return &v1alpha1.MachineSpec{
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(cfg),
		},
	}, nil
}

func mustEncodeProviderSpec(cfg *kubevirt.RawConfig) *runtime.RawExtension {
	pconfig := providerconfig.Config{
		CloudProviderSpec: runtime.RawExtension{
			Raw: toJSON(cfg),
		},
	}

	raw, err := json.Marshal(pconfig)
	if err != nil {
		panic(err)
	}

	return &runtime.RawExtension{Raw: raw}
}

func toJSON(i interface{}) []byte {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return b
}

// getDefaultKubevirtConfig returns the default kubevirt config.
func getDefaultKubevirtConfig() (*kubevirt.RawConfig, error) {
	return &kubevirt.RawConfig{
		ClusterName: providerconfig.ConfigVarString{Value: "test-cluster"},
		Auth: kubevirt.Auth{
			Kubeconfig: providerconfig.ConfigVarString{Value: "valid-kubeconfig"},
		},
		VirtualMachine: kubevirt.VirtualMachine{
			Instancetype:            &kubevirtv1.InstancetypeMatcher{Name: ""},
			Preference:              &kubevirtv1.PreferenceMatcher{Name: ""},
			DNSPolicy:               providerconfig.ConfigVarString{Value: "ClusterFirstWithHostNet"},
			EnableNetworkMultiQueue: providerconfig.ConfigVarBool{Value: ptr.To(true)},
			Template: kubevirt.Template{
				PrimaryDisk: kubevirt.PrimaryDisk{
					Disk: kubevirt.Disk{
						Size:             providerconfig.ConfigVarString{Value: "20Gi"},
						StorageClassName: providerconfig.ConfigVarString{Value: "kubermatic-fast"},
					},
				},
			},
		},
	}, nil
}
