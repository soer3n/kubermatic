package kubevirt

import (
	"github.com/aws/smithy-go/ptr"
	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/machine-controller/sdk/providerconfig"
	v1 "kubevirt.io/api/core/v1"
)

var datacenterSettings = map[string]kubermaticv1.Datacenter{
	"with default control plane network policies enabled": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				EnableDefaultNetworkPolicies: ptr.Bool(true),
			},
		},
	},
	"with default netpols disabled": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				EnableDefaultNetworkPolicies: ptr.Bool(false),
			},
		},
	},
	"with namespaced mode enabled": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				NamespacedMode: &kubermaticv1.NamespacedMode{
					Enabled:   true,
					Namespace: "kkp-namespaced-mode",
				},
			},
		},
	},
	"with namespaced mode disabled": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				NamespacedMode: &kubermaticv1.NamespacedMode{
					Enabled: false,
				},
			},
		},
	},
	"with dns policy set to ClusterFirst": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				DNSPolicy: "ClusterFirst",
			},
		},
	},
	"with dns policy set to Default": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				DNSPolicy: "Default",
			},
		},
	},
	"with dns policy set to None": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				DNSPolicy: "None",
			},
		},
	},
	"with images from container disk": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
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
	},
	"with images from http source": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				Images: kubermaticv1.KubeVirtImageSources{
					HTTP: &kubermaticv1.KubeVirtHTTPSource{
						OperatingSystems: map[providerconfig.OperatingSystem]kubermaticv1.OSVersions{
							providerconfig.OperatingSystemUbuntu: {
								"22.04": "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img",
							},
						},
					},
				},
			},
		},
	},
	"with eviction strategy set to live-migrate": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				VMEvictionStrategy: v1.EvictionStrategyLiveMigrate,
			},
		},
	},
	"with eviction strategy set to external": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				VMEvictionStrategy: v1.EvictionStrategyExternal,
			},
		},
	},
	"with match subnet and storage location enabled": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				MatchSubnetAndStorageLocation: ptr.Bool(true),
			},
		},
	},
	"with match subnet and storage location disabled": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				MatchSubnetAndStorageLocation: ptr.Bool(false),
			},
		},
	},
	"with default instance types enabled": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				DisableDefaultInstanceTypes: false,
			},
		},
	},
	"with default instance types disabled": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				DisableDefaultInstanceTypes: true,
			},
		},
	},
	"with default preferences types enabled": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				DisableDefaultPreferences: false,
			},
		},
	},
	"with default preferences types disabled": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				DisableDefaultPreferences: true,
			},
		},
	},
	"with ccm zone and region enabled": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				CCMZoneAndRegionEnabled: ptr.Bool(true),
			},
		},
	},
	"with ccm zone and region disabled": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				CCMZoneAndRegionEnabled: ptr.Bool(false),
			},
		},
	},
	"with ccm load balancer enabled": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				CCMLoadBalancerEnabled: ptr.Bool(true),
			},
		},
	},
	"with ccm load balancer disabled": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				CCMLoadBalancerEnabled: ptr.Bool(false),
			},
		},
	},
	"with use pod resources cpu enabled": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				UsePodResourcesCPU: true,
			},
		},
	},
	"with use pod resources cpu disabled": {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				UsePodResourcesCPU: false,
			},
		},
	},
}

var defaultDatacenterSettings = kubermaticv1.Datacenter{
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
