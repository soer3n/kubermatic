package kubevirt

import (
	"encoding/json"
	"strconv"

	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"

	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"

	"k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	"k8c.io/machine-controller/sdk/cloudprovider/kubevirt"
	"k8c.io/machine-controller/sdk/providerconfig"

	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	kubevirtv1 "kubevirt.io/api/core/v1"
	kubevirtv1alpha3 "kubevirt.io/api/instancetype/v1alpha1"

	"context"
	"fmt"
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
			spec.VirtualMachine.Template.PrimaryDisk.OsImage.Value = "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img"
			// This assumes some default URL is set elsewhere, as the model doesn't have a dedicated URL field anymore.
		},
	},
	{
		name:  "with primary disk OS image from a container",
		group: "os-image",
		modify: func(spec *kubevirt.RawConfig) {
			spec.VirtualMachine.Template.PrimaryDisk.OsImage.Value = "container"
			spec.VirtualMachine.Template.PrimaryDisk.Source.Value = "docker://quay.io/kubermatic-virt-disks/ubuntu:22.04"
			// This assumes some default image is set elsewhere.
		},
	},
	{
		name:  "with changed cluster name",
		group: "cluster-name",
		modify: func(spec *kubevirt.RawConfig) {
			spec.ClusterName.Value = "changed-cluster-name"
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
		name:  "with topology spread constraint on hostname",
		group: "topology-spread",
		modify: func(spec *kubevirt.RawConfig) {
			spec.TopologySpreadConstraints = []kubevirt.TopologySpreadConstraint{{
				TopologyKey:       providerconfig.ConfigVarString{Value: "kubernetes.io/hostname"},
				MaxSkew:           providerconfig.ConfigVarString{Value: "1"},
				WhenUnsatisfiable: providerconfig.ConfigVarString{Value: "ScheduleAnyway"},
			}}
		},
	},
	{
		name:  "with topology spread constraint set to DoNotSchedule",
		group: "topology-spread",
		modify: func(spec *kubevirt.RawConfig) {
			spec.TopologySpreadConstraints = []kubevirt.TopologySpreadConstraint{{
				TopologyKey:       providerconfig.ConfigVarString{Value: "kubernetes.io/hostname"},
				WhenUnsatisfiable: providerconfig.ConfigVarString{Value: "DoNotSchedule"},
				MaxSkew:           providerconfig.ConfigVarString{Value: "1"},
			}}
		},
	},
}

// getDefaultMachineSpec returns the default machine spec.
func getDefaultMachineSpec(infraClient ctrlruntimeclient.Client) (*v1alpha1.MachineSpec, error) {
	cfg, err := getDefaultKubevirtConfig(infraClient)
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
func getDefaultKubevirtConfig(infraClient ctrlruntimeclient.Client) (*kubevirt.RawConfig, error) {
	var scsList storagev1.StorageClassList
	if err := infraClient.List(context.TODO(), &scsList); err != nil {
		return nil, fmt.Errorf("failed to list storage classes: %w", err)
	}
	var defaultStorageClass *storagev1.StorageClass
	for _, sc := range scsList.Items {
		if v, ok := sc.ObjectMeta.Annotations["storageclass.kubernetes.io/is-default-class"]; ok && v == "true" {
			defaultStorageClass = &sc
		}
	}

	return &kubevirt.RawConfig{
		ClusterName: providerconfig.ConfigVarString{Value: "test-cluster"},
		VirtualMachine: kubevirt.VirtualMachine{
			DNSPolicy: providerconfig.ConfigVarString{Value: "None"},
			DNSConfig: &v1.PodDNSConfig{
				Nameservers: []string{"10.97.179.24"},
			},
			EnableNetworkMultiQueue: providerconfig.ConfigVarBool{Value: ptr.To(true)},
			Template: kubevirt.Template{
				CPUs:   providerconfig.ConfigVarString{Value: "2"},
				Memory: providerconfig.ConfigVarString{Value: "4096Mi"},
				PrimaryDisk: kubevirt.PrimaryDisk{
					OsImage: providerconfig.ConfigVarString{Value: "docker://quay.io/kubermatic-virt-disks/ubuntu:22.04"},
					Disk: kubevirt.Disk{
						Size:             providerconfig.ConfigVarString{Value: "20Gi"},
						StorageClassName: providerconfig.ConfigVarString{Value: defaultStorageClass.Name},
					},
				},
			},
		},
	}, nil
}

func MachineSettings(ctx context.Context, client ctrlruntimeclient.Client, namespace string, resources *k8cginkgo.ResourceSettings) []machineSpecModifier {
	discoverdSettings, err := discoverDefaultMachineSettings(ctx, client, namespace)
	if err != nil {
		panic(fmt.Errorf("failed to discover default machine settings: %w", err))
	}
	generatedMachineSettings := buildDefaultMachineSettings(discoverdSettings)
	for _, settings := range machineSettings {
		generatedMachineSettings = append(generatedMachineSettings, settings)
	}
	if resources == nil {
		return append(generatedMachineSettings, machineSpecModifier{
			name:  "with custom cpu and memory",
			group: "custom-resources",
			modify: func(spec *kubevirt.RawConfig) {
				// No-op, just a placeholder to indicate default resources.
			},
		})
	}
	if resources.Memory != nil {
		for _, settings := range memoryModifiers(resources.Memory) {
			generatedMachineSettings = append(generatedMachineSettings, settings)
		}
	}
	if resources.Cpu != nil {
		for _, settings := range cpuModifiers(resources.Cpu) {
			generatedMachineSettings = append(generatedMachineSettings, settings)
		}
	}
	if resources.DiskSize != nil {
		for _, settings := range diskModifiers(resources.DiskSize) {
			generatedMachineSettings = append(generatedMachineSettings, settings)
		}
	}
	return generatedMachineSettings
}

func memoryModifiers(memories []string) []machineSpecModifier {
	var mods []machineSpecModifier
	for _, mem := range memories {
		mods = append(mods, machineSpecModifier{
			name:  fmt.Sprintf("with %s memory", mem),
			group: "memory",
			modify: func(spec *kubevirt.RawConfig) {
				spec.VirtualMachine.Template.Memory.Value = mem
			},
		})
	}
	return mods
}

func cpuModifiers(cpus []int) []machineSpecModifier {
	var mods []machineSpecModifier
	for _, v := range cpus {
		v := v // capture range variable
		mods = append(mods, machineSpecModifier{
			name:  fmt.Sprintf("with %d vCPUs", v),
			group: "cpu",
			modify: func(spec *kubevirt.RawConfig) {
				spec.VirtualMachine.Template.VCPUs.Cores = v
			},
		})
		mods = append(mods, machineSpecModifier{
			name:  fmt.Sprintf("with %d CPUs", v),
			group: "cpu",
			modify: func(spec *kubevirt.RawConfig) {
				spec.VirtualMachine.Template.CPUs.Value = strconv.Itoa(v)
			},
		})
	}
	return mods
}

func diskModifiers(sizes []string) []machineSpecModifier {
	var mods []machineSpecModifier
	for _, size := range sizes {
		size := size // capture range variable
		mods = append(mods, machineSpecModifier{
			name:  fmt.Sprintf("with primary disk size %s", size),
			group: "primary-disk-size",
			modify: func(spec *kubevirt.RawConfig) {
				spec.VirtualMachine.Template.PrimaryDisk.Size.Value = size
			},
		})
		mods = append(mods, machineSpecModifier{
			name:  fmt.Sprintf("with secondary disk size %s", size),
			group: "secondary-disk-size",
			modify: func(spec *kubevirt.RawConfig) {
				if len(spec.VirtualMachine.Template.SecondaryDisks) == 0 {
					spec.VirtualMachine.Template.SecondaryDisks = append(spec.VirtualMachine.Template.SecondaryDisks, kubevirt.SecondaryDisks{
						Disk: kubevirt.Disk{},
					})
				}
				spec.VirtualMachine.Template.SecondaryDisks[0].Size.Value = size
			},
		})
	}
	return mods
}

type DefaultMachineSettings struct {
	StorageClasses []string
	NodeNames      []string
	InstanceTypes  []string
	Preferences    []string
}

func discoverDefaultMachineSettings(ctx context.Context, client ctrlruntimeclient.Client, namespace string) (*DefaultMachineSettings, error) {
	// Discover storage classes
	var scsList storagev1.StorageClassList
	if err := client.List(ctx, &scsList); err != nil {
		return nil, fmt.Errorf("failed to list storage classes: %w", err)
	}
	var storageClasses []string
	for _, sc := range scsList.Items {
		storageClasses = append(storageClasses, sc.Name)
	}

	// Discover node names
	var nodeList v1.NodeList
	if err := client.List(ctx, &nodeList); err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	var nodeNames []string
	for _, node := range nodeList.Items {
		nodeNames = append(nodeNames, node.Name)
	}

	// Discover KubeVirt instancetypes and preferences using ctrlruntimeclient
	var instanceTypes, preferences []string

	// Namespace-scoped instancetypes
	var nsITList kubevirtv1alpha3.VirtualMachineInstancetypeList
	if err := client.List(ctx, &nsITList, ctrlruntimeclient.InNamespace(namespace)); err == nil {
		for _, it := range nsITList.Items {
			instanceTypes = append(instanceTypes, it.Name)
		}
	}

	// Namespace-scoped preferences
	var nsPrefList kubevirtv1alpha3.VirtualMachinePreferenceList
	if err := client.List(ctx, &nsPrefList, ctrlruntimeclient.InNamespace(namespace)); err == nil {
		for _, pref := range nsPrefList.Items {
			preferences = append(preferences, pref.Name)
		}
	}

	return &DefaultMachineSettings{
		StorageClasses: storageClasses,
		NodeNames:      nodeNames,
		InstanceTypes:  instanceTypes,
		Preferences:    preferences,
	}, nil
}

func buildDefaultMachineSettings(settings *DefaultMachineSettings) []machineSpecModifier {
	var modifiers []machineSpecModifier

	// Storage classes
	for _, sc := range settings.StorageClasses {
		sc := sc // capture range variable
		modifiers = append(modifiers, machineSpecModifier{
			name:  fmt.Sprintf("with primary disk storage class set to %s", sc),
			group: "primary-disk-sc",
			modify: func(spec *kubevirt.RawConfig) {
				spec.VirtualMachine.Template.PrimaryDisk.StorageClassName.Value = sc
				spec.VirtualMachine.Template.PrimaryDisk.Size.Value = "20Gi"
			},
		})
		modifiers = append(modifiers, machineSpecModifier{
			name:  fmt.Sprintf("with secondary disk storage class set to %s", sc),
			group: "secondary-disk-sc",
			modify: func(spec *kubevirt.RawConfig) {
				if len(spec.VirtualMachine.Template.SecondaryDisks) == 0 {
					spec.VirtualMachine.Template.SecondaryDisks = append(spec.VirtualMachine.Template.SecondaryDisks, kubevirt.SecondaryDisks{
						Disk: kubevirt.Disk{Size: providerconfig.ConfigVarString{Value: "20Gi"}},
					})
				}
				spec.VirtualMachine.Template.SecondaryDisks[0].StorageClassName.Value = sc
				spec.VirtualMachine.Template.SecondaryDisks[0].Size.Value = "20Gi"
			},
		})
	}

	// Node names for affinity
	for _, nodeName := range settings.NodeNames {
		nodeName := nodeName // capture range variable
		modifiers = append(modifiers, machineSpecModifier{
			name:  fmt.Sprintf("with node affinity for hostname %s", nodeName),
			group: "node-affinity",
			modify: func(spec *kubevirt.RawConfig) {
				spec.Affinity.NodeAffinityPreset.Key.Value = "kubernetes.io/hostname"
				spec.Affinity.NodeAffinityPreset.Values = []providerconfig.ConfigVarString{{Value: nodeName}}
			},
		})
	}

	// Instance types
	for _, it := range settings.InstanceTypes {
		it := it // capture range variable
		modifiers = append(modifiers, machineSpecModifier{
			name:  fmt.Sprintf("with instancetype %s", it),
			group: "instancetype",
			modify: func(spec *kubevirt.RawConfig) {
				spec.VirtualMachine.Instancetype = &kubevirtv1.InstancetypeMatcher{Name: it}
			},
		})
	}

	// Preferences
	for _, pref := range settings.Preferences {
		pref := pref // capture range variable
		modifiers = append(modifiers, machineSpecModifier{
			name:  fmt.Sprintf("with preference %s", pref),
			group: "preference",
			modify: func(spec *kubevirt.RawConfig) {
				spec.VirtualMachine.Preference = &kubevirtv1.PreferenceMatcher{Name: pref}
			},
		})
	}

	return modifiers
}
