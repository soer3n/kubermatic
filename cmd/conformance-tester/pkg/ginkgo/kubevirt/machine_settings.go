package kubevirt

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/smithy-go/ptr"
	"k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	"k8c.io/machine-controller/sdk/cloudprovider/kubevirt"
	"k8c.io/machine-controller/sdk/providerconfig"
	kubevirtcorev1 "kubevirt.io/api/core/v1"
)

type testCase struct {
	os          string
	description string
	name        string
	variants    []string
}

type FieldVariantEntry struct {
	Description string
	TestCase    testCase
}

var machineSettings = map[string]v1alpha1.MachineSpec{
	"cluster name set to test-cluster": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				ClusterName: providerconfig.ConfigVarString{Value: "test-cluster"},
			}),
		},
	},
	"auth kubeconfig set to valid-kubeconfig": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				Auth: kubevirt.Auth{
					Kubeconfig: providerconfig.ConfigVarString{Value: "valid-kubeconfig"},
				},
			}),
		},
	},
	"virtual machine instancetype set to empty": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Instancetype: &kubevirtcorev1.InstancetypeMatcher{Name: ""},
				},
			}),
		},
	},
	"virtual machine preference set to empty": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Preference: &kubevirtcorev1.PreferenceMatcher{Name: ""},
				},
			}),
		},
	},
	"virtual machine dns policy set to ClusterFirstWithHostNet": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					DNSPolicy: providerconfig.ConfigVarString{Value: "ClusterFirstWithHostNet"},
				},
			}),
		},
	},
	"virtual machine eviction strategy set to LiveMigrate": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					EvictionStrategy: "LiveMigrate",
				},
			}),
		},
	},
	"virtual machine enable network multi queue set to true": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					EnableNetworkMultiQueue: providerconfig.ConfigVarBool{Value: ptr.Bool(true)},
				},
			}),
		},
	},
	"virtual machine enable network multi queue set to false": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					EnableNetworkMultiQueue: providerconfig.ConfigVarBool{Value: ptr.Bool(false)},
				},
			}),
		},
	},
	"virtual machine template CPUs set to 2": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Template: kubevirt.Template{
						CPUs: providerconfig.ConfigVarString{Value: "2"},
					},
				},
			}),
		},
	},
	"virtual machine template vCPUs set to 2": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Template: kubevirt.Template{
						VCPUs: kubevirt.VCPUs{Cores: 2},
					},
				},
			}),
		},
	},
	"virtual machine template memory set to 4096Mi": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Template: kubevirt.Template{
						Memory: providerconfig.ConfigVarString{Value: "4096Mi"},
					},
				},
			}),
		},
	},
	"virtual machine template primary disk size set to 20Gi": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Template: kubevirt.Template{
						PrimaryDisk: kubevirt.PrimaryDisk{
							Disk: kubevirt.Disk{
								Size: providerconfig.ConfigVarString{Value: "20Gi"},
							},
						},
					},
				},
			}),
		},
	},
	"virtual machine template primary disk storage class set to standard": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Template: kubevirt.Template{
						PrimaryDisk: kubevirt.PrimaryDisk{
							Disk: kubevirt.Disk{
								StorageClassName: providerconfig.ConfigVarString{Value: "standard"},
							},
						},
					},
				},
			}),
		},
	},
	"virtual machine template secondary disks size set to 10Gi and storage class set to standard": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Template: kubevirt.Template{
						SecondaryDisks: []kubevirt.SecondaryDisks{{
							Disk: kubevirt.Disk{
								Size:             providerconfig.ConfigVarString{Value: "10Gi"},
								StorageClassName: providerconfig.ConfigVarString{Value: "standard"},
							},
						}}},
				},
			}),
		},
	},
	"virtual machine template secondary disks size set to 10Gi and storage class set to kubermatic-fast": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Template: kubevirt.Template{
						SecondaryDisks: []kubevirt.SecondaryDisks{{
							Disk: kubevirt.Disk{
								Size:             providerconfig.ConfigVarString{Value: "10Gi"},
								StorageClassName: providerconfig.ConfigVarString{Value: "kubermatic-fast"},
							},
						}}},
				},
			}),
		},
	},
	"affinity node affinity preset type set to empty": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				Affinity: kubevirt.Affinity{
					NodeAffinityPreset: kubevirt.NodeAffinityPreset{
						Type: providerconfig.ConfigVarString{Value: ""},
					},
				},
			}),
		},
	},
	"affinity node affinity preset key set to kubernetes.io/hostname": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				Affinity: kubevirt.Affinity{
					NodeAffinityPreset: kubevirt.NodeAffinityPreset{
						Key: providerconfig.ConfigVarString{Value: "kubernetes.io/hostname"},
					},
				},
			}),
		},
	},
	"affinity node affinity preset values set to node-01": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				Affinity: kubevirt.Affinity{
					NodeAffinityPreset: kubevirt.NodeAffinityPreset{
						Values: []providerconfig.ConfigVarString{{Value: "node-01"}},
					},
				},
			}),
		},
	},
	"topology spread constraints topology key set to kubernetes.io/hostname and max skew set to 1": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				TopologySpreadConstraints: []kubevirt.TopologySpreadConstraint{{
					TopologyKey: providerconfig.ConfigVarString{Value: "kubernetes.io/hostname"},
					MaxSkew:     providerconfig.ConfigVarString{Value: "1"},
				}},
			}),
		},
	},
	"topology spread constraints when unsatisfiable set to DoNotSchedule and max skew set to 1": {
		ProviderSpec: v1alpha1.ProviderSpec{
			Value: mustEncodeProviderSpec(kubevirt.RawConfig{
				TopologySpreadConstraints: []kubevirt.TopologySpreadConstraint{{
					WhenUnsatisfiable: providerconfig.ConfigVarString{Value: "DoNotSchedule"},
					MaxSkew:           providerconfig.ConfigVarString{Value: "1"},
				}},
			}),
		},
	},
}

var DefaultFieldVariantsTableEntries = []FieldVariantEntry{
	{Description: "ClusterName=test-cluster", TestCase: testCase{os: "", description: "ClusterName", name: "test-cluster", variants: []string{"test-cluster"}}},
	{Description: "Auth.Kubeconfig=valid-kubeconfig", TestCase: testCase{os: "", description: "Auth.Kubeconfig", name: "valid-kubeconfig", variants: []string{"valid-kubeconfig"}}},
	{Description: "VirtualMachine.Instancetype=", TestCase: testCase{os: "", description: "VirtualMachine.Instancetype", name: "", variants: []string{""}}},
	{Description: "VirtualMachine.Preference=", TestCase: testCase{os: "", description: "VirtualMachine.Preference", name: "", variants: []string{""}}},
	{Description: "VirtualMachine.DNSPolicy=ClusterFirstWithHostNet", TestCase: testCase{os: "", description: "VirtualMachine.DNSPolicy", name: "ClusterFirstWithHostNet", variants: []string{"ClusterFirstWithHostNet"}}},
	{Description: "VirtualMachine.EvictionStrategy=LiveMigrate", TestCase: testCase{os: "", description: "VirtualMachine.EvictionStrategy", name: "LiveMigrate", variants: []string{"LiveMigrate"}}},
	{Description: "VirtualMachine.EnableNetworkMultiQueue=true", TestCase: testCase{os: "", description: "VirtualMachine.EnableNetworkMultiQueue", name: "true", variants: []string{"true"}}},
	{Description: "VirtualMachine.EnableNetworkMultiQueue=false", TestCase: testCase{os: "", description: "VirtualMachine.EnableNetworkMultiQueue", name: "false", variants: []string{"false"}}},
	{Description: "VirtualMachine.Template.CPUs=2", TestCase: testCase{os: "", description: "VirtualMachine.Template.CPUs", name: "2", variants: []string{"2"}}},
	{Description: "VirtualMachine.Template.Memory=4096Mi", TestCase: testCase{os: "", description: "VirtualMachine.Template.Memory", name: "4096Mi", variants: []string{"4096Mi"}}},
	{Description: "VirtualMachine.Template.PrimaryDisk.Size=20Gi", TestCase: testCase{os: "", description: "VirtualMachine.Template.PrimaryDisk.Size", name: "20Gi", variants: []string{"20Gi"}}},
	{Description: "VirtualMachine.Template.PrimaryDisk.StorageClassName=local-path", TestCase: testCase{os: "", description: "VirtualMachine.Template.PrimaryDisk.StorageClassName", name: "local-path", variants: []string{"local-path"}}},
	{Description: "VirtualMachine.Template.SecondaryDisks.Size=10Gi", TestCase: testCase{os: "", description: "VirtualMachine.Template.SecondaryDisks.Size", name: "10Gi", variants: []string{"10Gi"}}},
	{Description: "VirtualMachine.Template.SecondaryDisks.StorageClassName=fast-storage", TestCase: testCase{os: "", description: "VirtualMachine.Template.SecondaryDisks.StorageClassName", name: "fast-storage", variants: []string{"fast-storage"}}},
	{Description: "Affinity.NodeAffinityPreset.Type=", TestCase: testCase{os: "", description: "Affinity.NodeAffinityPreset.Type", name: "", variants: []string{""}}},
	{Description: "Affinity.NodeAffinityPreset.Key=kubernetes.io/hostname", TestCase: testCase{os: "", description: "Affinity.NodeAffinityPreset.Key", name: "kubernetes.io/hostname", variants: []string{"kubernetes.io/hostname"}}},
	{Description: "Affinity.NodeAffinityPreset.Values=node-01", TestCase: testCase{os: "", description: "Affinity.NodeAffinityPreset.Values", name: "node-01", variants: []string{"node-01"}}},
	{Description: "TopologySpreadConstraints.TopologyKey=kubernetes.io/hostname", TestCase: testCase{os: "", description: "TopologySpreadConstraints.TopologyKey", name: "kubernetes.io/hostname", variants: []string{"kubernetes.io/hostname"}}},
	{Description: "TopologySpreadConstraints.MaxSkew=1", TestCase: testCase{os: "", description: "TopologySpreadConstraints.MaxSkew", name: "1", variants: []string{"1"}}},
	{Description: "TopologySpreadConstraints.WhenUnsatisfiable=DoNotSchedule", TestCase: testCase{os: "", description: "TopologySpreadConstraints.WhenUnsatisfiable", name: "DoNotSchedule", variants: []string{"DoNotSchedule"}}},
}

var defaultKubevirtConfig = kubevirt.RawConfig{
	ClusterName: providerconfig.ConfigVarString{Value: "test-cluster"},
	Auth: kubevirt.Auth{
		Kubeconfig: providerconfig.ConfigVarString{Value: "valid-kubeconfig"},
	},
	VirtualMachine: kubevirt.VirtualMachine{
		Instancetype:            &kubevirtcorev1.InstancetypeMatcher{Name: ""},
		Preference:              &kubevirtcorev1.PreferenceMatcher{Name: ""},
		DNSPolicy:               providerconfig.ConfigVarString{Value: "ClusterFirstWithHostNet"},
		EvictionStrategy:        "LiveMigrate",
		EnableNetworkMultiQueue: providerconfig.ConfigVarBool{Value: ptr.Bool(true)},
		Template: kubevirt.Template{
			CPUs:   providerconfig.ConfigVarString{Value: "2"},
			Memory: providerconfig.ConfigVarString{Value: "4096Mi"},
			PrimaryDisk: kubevirt.PrimaryDisk{
				Disk: kubevirt.Disk{
					Size:             providerconfig.ConfigVarString{Value: "20Gi"},
					StorageClassName: providerconfig.ConfigVarString{Value: "standard"},
				},
			},
			SecondaryDisks: []kubevirt.SecondaryDisks{},
		},
	},
	TopologySpreadConstraints: []kubevirt.TopologySpreadConstraint{
		{
			TopologyKey:       providerconfig.ConfigVarString{Value: "kubernetes.io/hostname"},
			MaxSkew:           providerconfig.ConfigVarString{Value: "1"},
			WhenUnsatisfiable: providerconfig.ConfigVarString{Value: "DoNotSchedule"},
		},
	},
}

func EncodeRawSpec(cfg kubevirt.RawConfig) (*runtime.RawExtension, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	return &runtime.RawExtension{Raw: data}, nil
}

func mustEncodeProviderSpec(cfg kubevirt.RawConfig) *runtime.RawExtension {
	re, err := EncodeRawSpec(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to encode provider spec: %v", err))
	}
	return re
}
