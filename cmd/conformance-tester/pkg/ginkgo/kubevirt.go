package ginkgo

import (
	"k8c.io/machine-controller/sdk/cloudprovider/kubevirt"
	"k8c.io/machine-controller/sdk/providerconfig"
	kubevirtcorev1 "kubevirt.io/api/core/v1"
)

func getKubevirtTestSettings() []TestSettings {
	return []TestSettings{
		{
			Description:  "with default settings",
			ProviderSpec: &kubevirt.RawConfig{},
		},
		{
			Description: "with a specific instancetype",
			ProviderSpec: &kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Instancetype: &kubevirtcorev1.InstancetypeMatcher{Name: "u1.small"},
				},
			},
		},
		{
			Description: "with a specific preference",
			ProviderSpec: &kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Preference: &kubevirtcorev1.PreferenceMatcher{Name: "fedora"},
				},
			},
		},
		{
			Description: "with custom CPU and Memory",
			ProviderSpec: &kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Template: kubevirt.Template{
						CPUs:   providerconfig.ConfigVarString{Value: "2"},
						Memory: providerconfig.ConfigVarString{Value: "4Gi"},
					},
				},
			},
		},
		{
			Description: "with a secondary disk",
			ProviderSpec: &kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Template: kubevirt.Template{
						SecondaryDisks: []kubevirt.SecondaryDisks{
							{
								Disk: kubevirt.Disk{
									Size:             providerconfig.ConfigVarString{Value: "10Gi"},
									StorageClassName: providerconfig.ConfigVarString{Value: "local-path"},
								},
							},
						},
					},
				},
			},
		},
		{
			Description: "with a custom DNS policy",
			ProviderSpec: &kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					DNSPolicy: providerconfig.ConfigVarString{Value: "ClusterFirstWithHostNet"},
				},
			},
		},
		{
			Description: "with a live migration eviction strategy",
			ProviderSpec: &kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					EvictionStrategy: "LiveMigrate",
				},
			},
		},
		{
			Description: "with a required node affinity",
			ProviderSpec: &kubevirt.RawConfig{
				Affinity: kubevirt.Affinity{
					NodeAffinityPreset: kubevirt.NodeAffinityPreset{
						Type: providerconfig.ConfigVarString{Value: "required"},
						Key:  providerconfig.ConfigVarString{Value: "kubernetes.io/hostname"},
						Values: []providerconfig.ConfigVarString{
							{Value: "node-01"},
						},
					},
				},
			},
		},
		{
			Description: "with topology spread constraints",
			ProviderSpec: &kubevirt.RawConfig{
				TopologySpreadConstraints: []kubevirt.TopologySpreadConstraint{
					{
						MaxSkew:           providerconfig.ConfigVarString{Value: "1"},
						TopologyKey:       providerconfig.ConfigVarString{Value: "kubernetes.io/hostname"},
						WhenUnsatisfiable: providerconfig.ConfigVarString{Value: "DoNotSchedule"},
					},
				},
			},
		},
	}
}
