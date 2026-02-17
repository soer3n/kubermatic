package settings

import (
	"k8c.io/machine-controller/sdk/cloudprovider/kubevirt"
	"k8c.io/machine-controller/sdk/providerconfig"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"

	kubevirtv1 "kubevirt.io/api/core/v1"
)

// func GetMachineDescriptions(client ctrlclient.Client) map[string]k8cginkgo.Description {
// 	// client, _, err := utils.GetClients()
// 	// if err != nil {
// 	// 	return nil
// 	// }

// 	settings := MachineSettings(context.Background(), client, "kubermatic", nil)

// 	groupedSettings := map[string]k8cginkgo.Description{}
// 	groupedSettingsDesc := map[string][]string{}
// 	for _, modifier := range settings {
// 		groupedSettingsDesc[modifier.group] = append(groupedSettingsDesc[modifier.group], modifier.name)
// 	}

// 	for group, descs := range groupedSettingsDesc {
// 		strippedDescs := stripPrefix(descs)
// 		if len(strippedDescs) == 1 {
// 			strippedDescs = nil
// 		}
// 		groupedSettings[group] = k8cginkgo.Description{
// 			Name:    longestCommonPrefixTokens(descs, " "),
// 			Options: strippedDescs,
// 		}
// 	}
// 	return groupedSettings
// }

// func GetDatacenterDescriptions(client ctrlclient.Client) map[string]k8cginkgo.Description {
// 	// client, _, err := utils.GetClients()
// 	// if err != nil {
// 	// 	return nil
// 	// }
// 	settings := GenericDatacenterSettings(context.Background(), client, "kubermatic")
// 	groupedSettings := map[string]k8cginkgo.Description{}
// 	groupedSettingsDesc := map[string][]string{}
// 	for _, modifier := range settings {
// 		groupedSettingsDesc[modifier.Group] = append(groupedSettingsDesc[modifier.Group], modifier.Name)
// 	}

// 	for group, descs := range groupedSettingsDesc {
// 		strippedDescs := stripPrefix(descs)
// 		if len(strippedDescs) == 1 {
// 			strippedDescs = nil
// 		}
// 		groupedSettings[group] = k8cginkgo.Description{
// 			Name:    longestCommonPrefix(descs),
// 			Options: strippedDescs,
// 		}
// 	}
// 	return groupedSettings
// }

// func longestCommonPrefix(strs []string) string {
// 	if len(strs) == 0 {
// 		return ""
// 	}

// 	prefix := strs[0]

// 	for _, s := range strs[1:] {
// 		for len(prefix) > 0 && !strings.HasPrefix(s, prefix) {
// 			prefix = prefix[:len(prefix)-1]
// 		}
// 		if prefix == "" {
// 			return ""
// 		}
// 	}
// 	return prefix
// }

// func longestCommonPrefixTokens(strs []string, sep string) string {
// 	if len(strs) == 0 {
// 		return ""
// 	}

// 	base := strings.Split(strs[0], sep)
// 	maxTokens := len(base)

// 	for _, s := range strs[1:] {
// 		tokens := strings.Split(s, sep)

// 		i := 0
// 		for i < maxTokens && i < len(tokens) && tokens[i] == base[i] {
// 			i++
// 		}
// 		maxTokens = i

// 		if maxTokens == 0 {
// 			return ""
// 		}
// 	}

// 	prefix := strings.Join(base[:maxTokens], sep)

// 	// preserve trailing separator if it existed
// 	return prefix + sep
// }

// func stripPrefix(strs []string) []string {
// 	prefix := longestCommonPrefixTokens(strs, " ")
// 	out := make([]string, 0, len(strs))

// 	for _, s := range strs {
// 		out = append(out, strings.TrimPrefix(s, prefix))
// 	}
// 	return out
// }

var MachineSettings = []MachineSpecModifier[*kubevirt.RawConfig]{
	{
		Name:  "with primary disk OS image from an HTTP source",
		Group: "os-image",
		Modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Template.PrimaryDisk.Source.Value = "http"
			spec.VirtualMachine.Template.PrimaryDisk.OsImage.Value = "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img"
			// This assumes some default URL is set elsewhere, as the model doesn't have a dedicated URL field anymore.
		},
	},
	{
		Name:  "with primary disk OS image from a container",
		Group: "os-image",
		Modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Template.PrimaryDisk.OsImage.Value = "container"
			spec.VirtualMachine.Template.PrimaryDisk.Source.Value = "docker://quay.io/kubermatic-virt-disks/ubuntu:22.04"
			// This assumes some default image is set elsewhere.
		},
	},
	{
		Name:  "with changed cluster name",
		Group: "cluster-name",
		Modify: func(spec *kubevirt.RawConfig) {
			spec.ClusterName.Value = "changed-cluster-name"
		},
	},
	{
		Name:  "with empty instancetype",
		Group: "instancetype",
		Modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Instancetype = &kubevirtv1.InstancetypeMatcher{Name: ""}
		},
	},
	{
		Name:  "with empty preference",
		Group: "preference",
		Modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Preference = &kubevirtv1.PreferenceMatcher{Name: ""}
		},
	},
	{
		Name:  "with dns policy set to ClusterFirstWithHostNet",
		Group: "dns-policy",
		Modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.DNSPolicy.Value = "ClusterFirstWithHostNet"
		},
	},
	{
		Name:  "with dns policy set to ClusterFirst",
		Group: "dns-policy",
		Modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.DNSPolicy.Value = "ClusterFirst"
		},
	},
	{
		Name:  "with dns policy set to Default",
		Group: "dns-policy",
		Modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.DNSPolicy.Value = "Default"
		},
	},
	{
		Name:  "with dns policy set to None",
		Group: "dns-policy",
		Modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.DNSPolicy.Value = "None"
			spec.VirtualMachine.DNSConfig = &v1.PodDNSConfig{
				Nameservers: []string{"8.8.8.8", "8.8.4.4"},
			}
		},
	},
	{
		Name:  "with eviction strategy set to LiveMigrate",
		Group: "eviction-strategy",
		Modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.EvictionStrategy = "LiveMigrate"
		},
	},
	{
		Name:  "with eviction strategy set to External",
		Group: "eviction-strategy",
		Modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.EvictionStrategy = "External"
		},
	},
	{
		Name:  "with network multi-queue enabled",
		Group: "multi-queue",
		Modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.EnableNetworkMultiQueue.Value = ptr.To(true)
		},
	},
	{
		Name:  "with network multi-queue disabled",
		Group: "multi-queue",
		Modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.EnableNetworkMultiQueue.Value = ptr.To(false)
		},
	},
	{
		Name:  "with topology spread constraint on hostname",
		Group: "topology-spread",
		Modify: func(spec *kubevirt.RawConfig) {
			spec.TopologySpreadConstraints = []kubevirt.TopologySpreadConstraint{{
				TopologyKey:       providerconfig.ConfigVarString{Value: "kubernetes.io/hostname"},
				MaxSkew:           providerconfig.ConfigVarString{Value: "1"},
				WhenUnsatisfiable: providerconfig.ConfigVarString{Value: "ScheduleAnyway"},
			}}
		},
	},
	{
		Name:  "with topology spread constraint set to DoNotSchedule",
		Group: "topology-spread",
		Modify: func(spec *kubevirt.RawConfig) {
			spec.TopologySpreadConstraints = []kubevirt.TopologySpreadConstraint{{
				TopologyKey:       providerconfig.ConfigVarString{Value: "kubernetes.io/hostname"},
				WhenUnsatisfiable: providerconfig.ConfigVarString{Value: "DoNotSchedule"},
				MaxSkew:           providerconfig.ConfigVarString{Value: "1"},
			}}
		},
	},
}

var DatacenterSettings = []DatacenterSetting{
	{
		Name:  "with default control plane network policies enabled",
		Group: "netpol",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.EnableDefaultNetworkPolicies = ptr.To(true)
		},
	},
	{
		Name:  "with default control plane network policies disabled",
		Group: "netpol",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.EnableDefaultNetworkPolicies = ptr.To(false)
		},
	},
	{
		Name:  "with namespaced mode enabled",
		Group: "namespace",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.NamespacedMode = &kubermaticv1.NamespacedMode{
				Enabled:   true,
				Namespace: "kkp-namespaced-mode",
			}
		},
	},
	{
		Name:  "with namespaced mode disabled",
		Group: "namespace",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.NamespacedMode = &kubermaticv1.NamespacedMode{
				Enabled: false,
			}
		},
	},
	{
		Name:  "with dns policy set to ClusterFirst",
		Group: "dns",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.DNSPolicy = "ClusterFirst"
		},
	},
	{
		Name:  "with dns policy set to Default",
		Group: "dns",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.DNSPolicy = "Default"
		},
	},
	{
		Name:  "with dns policy set to None",
		Group: "dns",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.DNSPolicy = "None"
		},
	},
	{
		Name:  "with images from container disk",
		Group: "images",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.Images = kubermaticv1.KubeVirtImageSources{
				HTTP: &kubermaticv1.KubeVirtHTTPSource{
					OperatingSystems: map[providerconfig.OperatingSystem]kubermaticv1.OSVersions{
						providerconfig.OperatingSystemUbuntu: {
							"22.04": "docker://quay.io/kubermatic-virt-disks/ubuntu:22.04",
						},
					},
				},
			}
		},
	},
	{
		Name:  "with images from http source",
		Group: "images",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.Images = kubermaticv1.KubeVirtImageSources{
				HTTP: &kubermaticv1.KubeVirtHTTPSource{
					OperatingSystems: map[providerconfig.OperatingSystem]kubermaticv1.OSVersions{
						providerconfig.OperatingSystemUbuntu: {
							"22.04": "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img",
						},
					},
				},
			}
		},
	},
	{
		Name:  "with eviction strategy set to live-migrate",
		Group: "eviction",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.VMEvictionStrategy = kubevirtv1.EvictionStrategyLiveMigrate
		},
	},
	{
		Name:  "with eviction strategy set to external",
		Group: "eviction",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.VMEvictionStrategy = kubevirtv1.EvictionStrategyExternal
		},
	},
	{
		Name:  "with match subnet and storage location enabled",
		Group: "subnet",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.MatchSubnetAndStorageLocation = ptr.To(true)
		},
	},
	{
		Name:  "with match subnet and storage location disabled",
		Group: "subnet",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.MatchSubnetAndStorageLocation = ptr.To(false)
		},
	},
	{
		Name:  "with default instance types enabled",
		Group: "instancetypes",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.DisableDefaultInstanceTypes = false
		},
	},
	{
		Name:  "with default instance types disabled",
		Group: "instancetypes",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.DisableDefaultInstanceTypes = true
		},
	},
	{
		Name:  "with default preferences types enabled",
		Group: "preferences",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.DisableDefaultPreferences = false
		},
	},
	{
		Name:  "with default preferences types disabled",
		Group: "preferences",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.DisableDefaultPreferences = true
		},
	},
	{
		Name:  "with ccm zone and region enabled",
		Group: "ccm",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.CCMZoneAndRegionEnabled = ptr.To(true)
		},
	},
	{
		Name:  "with ccm zone and region disabled",
		Group: "ccm",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.CCMZoneAndRegionEnabled = ptr.To(false)
		},
	},
	{
		Name:  "with ccm load balancer enabled",
		Group: "ccm-lb",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.CCMLoadBalancerEnabled = ptr.To(true)
		},
	},
	{
		Name:  "with ccm load balancer disabled",
		Group: "ccm-lb",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.CCMLoadBalancerEnabled = ptr.To(false)
		},
	},
	{
		Name:  "with use pod resources cpu enabled",
		Group: "pod-cpu",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.UsePodResourcesCPU = true
		},
	},

	{
		Name:  "with use pod resources cpu disabled",
		Group: "pod-cpu",
		Modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.UsePodResourcesCPU = false
		},
	},
}
