package kubevirt

import (
	"dario.cat/mergo"
	"github.com/aws/smithy-go/ptr"
	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/machine-controller/sdk/providerconfig"
	v1 "kubevirt.io/api/core/v1"
)

type DatacenterSetting struct {
	name     string
	group    string
	modifier func(dc *kubermaticv1.Datacenter)
}

var datacenterSettings = []DatacenterSetting{
	{
		name:  "with default control plane network policies enabled",
		group: "netpol",
		modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.EnableDefaultNetworkPolicies = ptr.Bool(true)
		},
	},
	{
		name:  "with default control plane network policies disabled",
		group: "netpol",
		modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.EnableDefaultNetworkPolicies = ptr.Bool(false)
		},
	},
	{
		name:  "with namespaced mode enabled",
		group: "namespace",
		modifier: func(dc *kubermaticv1.Datacenter) {
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
		name:  "with namespaced mode disabled",
		group: "namespace",
		modifier: func(dc *kubermaticv1.Datacenter) {
			if dc.Spec.Kubevirt == nil {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
			}
			dc.Spec.Kubevirt.NamespacedMode = &kubermaticv1.NamespacedMode{
				Enabled: false,
			}
		},
	},
	// {
	// 	name:  "with dns policy set to ClusterFirst",
	// 	group: "dns",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.DNSPolicy = "ClusterFirst"
	// 	},
	// },

	// {
	// 	name:  "with dns policy set to Default",
	// 	group: "dns",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.DNSPolicy = "Default"
	// 	},
	// },

	// {
	// 	name:  "with dns policy set to None",
	// 	group: "dns",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.DNSPolicy = "None"
	// 	},
	// },

	// {
	// 	name:  "with images from container disk",
	// 	group: "images",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.Images = kubermaticv1.KubeVirtImageSources{
	// 			HTTP: &kubermaticv1.KubeVirtHTTPSource{
	// 				OperatingSystems: map[providerconfig.OperatingSystem]kubermaticv1.OSVersions{
	// 					providerconfig.OperatingSystemUbuntu: {
	// 						"22.04": "docker://quay.io/kubermatic-virt-disks/ubuntu:22.04",
	// 					},
	// 				},
	// 			},
	// 		}
	// 	},
	// },

	// {
	// 	name:  "with images from http source",
	// 	group: "images",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.Images = kubermaticv1.KubeVirtImageSources{
	// 			HTTP: &kubermaticv1.KubeVirtHTTPSource{
	// 				OperatingSystems: map[providerconfig.OperatingSystem]kubermaticv1.OSVersions{
	// 					providerconfig.OperatingSystemUbuntu: {
	// 						"22.04": "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img",
	// 					},
	// 				},
	// 			},
	// 		}
	// 	},
	// },

	// {
	// 	name:  "with eviction strategy set to live-migrate",
	// 	group: "eviction",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.VMEvictionStrategy = v1.EvictionStrategyLiveMigrate
	// 	},
	// },

	// {
	// 	name:  "with eviction strategy set to external",
	// 	group: "eviction",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.VMEvictionStrategy = v1.EvictionStrategyExternal
	// 	},
	// },

	// {
	// 	name:  "with match subnet and storage location enabled",
	// 	group: "subnet",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.MatchSubnetAndStorageLocation = ptr.Bool(true)
	// 	},
	// },

	// {
	// 	name:  "with match subnet and storage location disabled",
	// 	group: "subnet",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.MatchSubnetAndStorageLocation = ptr.Bool(false)
	// 	},
	// },

	// {
	// 	name:  "with default instance types enabled",
	// 	group: "instancetypes",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.DisableDefaultInstanceTypes = false
	// 	},
	// },

	// {
	// 	name:  "with default instance types disabled",
	// 	group: "instancetypes",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.DisableDefaultInstanceTypes = true
	// 	},
	// },

	// {
	// 	name:  "with default preferences types enabled",
	// 	group: "preferences",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.DisableDefaultPreferences = false
	// 	},
	// },

	// {
	// 	name:  "with default preferences types disabled",
	// 	group: "preferences",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.DisableDefaultPreferences = true
	// 	},
	// },

	// {
	// 	name:  "with ccm zone and region enabled",
	// 	group: "ccm",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.CCMZoneAndRegionEnabled = ptr.Bool(true)
	// 	},
	// },

	// {
	// 	name:  "with ccm zone and region disabled",
	// 	group: "ccm",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.CCMZoneAndRegionEnabled = ptr.Bool(false)
	// 	},
	// },

	// {
	// 	name:  "with ccm load balancer enabled",
	// 	group: "ccm-lb",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.CCMLoadBalancerEnabled = ptr.Bool(true)
	// 	},
	// },

	// {
	// 	name:  "with ccm load balancer disabled",
	// 	group: "ccm-lb",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.CCMLoadBalancerEnabled = ptr.Bool(false)
	// 	},
	// },

	// {
	// 	name:  "with use pod resources cpu enabled",
	// 	group: "pod-cpu",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.UsePodResourcesCPU = true
	// 	},
	// },

	// {
	// 	name:  "with use pod resources cpu disabled",
	// 	group: "pod-cpu",
	// 	modifier: func(dc *kubermaticv1.Datacenter) {
	// 		if dc.Spec.Kubevirt == nil {
	// 			dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
	// 		}
	// 		dc.Spec.Kubevirt.UsePodResourcesCPU = false
	// 	},
	// },
}

var defaultDatacenterSettings = DatacenterSetting{
	name: "default",
	modifier: func(dc *kubermaticv1.Datacenter) {
		defaultDC := kubermaticv1.Datacenter{
			Spec: kubermaticv1.DatacenterSpec{
				Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
					EnableDefaultNetworkPolicies:  ptr.Bool(true),
					NamespacedMode:                &kubermaticv1.NamespacedMode{Enabled: false},
					DNSPolicy:                     "ClusterFirst",
					VMEvictionStrategy:            v1.EvictionStrategyLiveMigrate,
					MatchSubnetAndStorageLocation: ptr.Bool(false),
					DisableDefaultInstanceTypes:   false,
					DisableDefaultPreferences:     false,
					CCMZoneAndRegionEnabled:       ptr.Bool(false),
					CCMLoadBalancerEnabled:        ptr.Bool(false),
					UsePodResourcesCPU:            false,
					Images: kubermaticv1.KubeVirtImageSources{
						HTTP: &kubermaticv1.KubeVirtHTTPSource{
							OperatingSystems: map[providerconfig.OperatingSystem]kubermaticv1.OSVersions{
								providerconfig.OperatingSystemUbuntu: {
									"22.04": "docker://quay.io/kubermatic-virt-disks/ubuntu:22.04",
								},
							},
						},
					},
				},
			},
		}
		if err := mergo.Merge(dc, defaultDC); err != nil {
			return
		}
	},
}
