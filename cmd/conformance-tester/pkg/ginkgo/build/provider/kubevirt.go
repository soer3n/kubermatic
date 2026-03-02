package provider

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"strconv"

	"go.uber.org/zap"
	"k8c.io/machine-controller/sdk/cloudprovider/kubevirt"
	"k8c.io/machine-controller/sdk/providerconfig"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/ptr"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/options"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/settings"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"

	kubevirtv1 "kubevirt.io/api/core/v1"
	kubevirtv1alpha3 "kubevirt.io/api/instancetype/v1alpha1"
)

type KubeVirtProvider settings.Provider

type KubeVirtProviderAdapter struct {
	inner *KubeVirtProvider
}

func NewKubeVirtProviderAdapter(k KubeVirtProvider) *KubeVirtProviderAdapter {
	return &KubeVirtProviderAdapter{
		inner: &k,
	}
}

func (a KubeVirtProviderAdapter) CpuModifiers(cpu []int) []settings.MachineSpecModifier[any] {
	raw := a.inner.CpuModifiers(cpu)
	return ConvertModifiersToAny(raw)
}

// Repeat for MemoryModifiers, DiskModifiers, MachineSettings, etc.

func (a KubeVirtProviderAdapter) MemoryModifiers(mem []string) []settings.MachineSpecModifier[any] {
	raw := a.inner.MemoryModifiers(mem)
	return ConvertModifiersToAny(raw)
}

func (a KubeVirtProviderAdapter) DiskModifiers(disk []string) []settings.MachineSpecModifier[any] {
	raw := a.inner.DiskModifiers(disk)
	return ConvertModifiersToAny(raw)
}

func (a KubeVirtProviderAdapter) DiscoverDefaultDatacenterSettings(ctx context.Context, providerConfig *providerconfig.Config, secrets legacytypes.Secrets) (*settings.DefaultDatacenterSettings, error) {
	raw, err := a.inner.DiscoverDefaultDatacenterSettings(ctx, providerConfig, secrets)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (a KubeVirtProviderAdapter) BuildDefaultDatacenterSettings(settings *settings.DefaultDatacenterSettings) []settings.DatacenterSetting {
	raw := a.inner.BuildDefaultDatacenterSettings(settings)
	return raw
}

type DefaultDatacenterSettings struct {
	VPCs []VPC
}

type VPC struct {
	Name    string
	Subnets []kubermaticv1.Subnet
}

func (k *KubeVirtProvider) DiscoverDefaultDatacenterSettings(ctx context.Context, providerConfig *providerconfig.Config, secrets legacytypes.Secrets) (*settings.DefaultDatacenterSettings, error) {
	defaultSettings := &settings.DefaultDatacenterSettings{}

	var err error

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	config, err := clientcmd.BuildConfigFromFlags("", secrets.Kubevirt.KubeconfigFile)
	if err != nil {
		panic(err)
	}

	client, err := ctrlruntimeclient.New(config, ctrlruntimeclient.Options{
		Scheme: scheme,
	})
	if err != nil {
		panic(err)
	}

	// Discover VPCs (if you have a VPC CRD, adjust group/version/kind accordingly)
	vpcList := &unstructured.UnstructuredList{}
	vpcList.SetAPIVersion("kubeovn.io/v1") // adjust if needed
	vpcList.SetKind("VpcList")             // adjust if needed
	if err := client.List(ctx, vpcList); err == nil {
		for _, item := range vpcList.Items {
			vpc := settings.VPC{}
			if name, found, _ := unstructured.NestedString(item.Object, "metadata", "name"); found {
				vpc.Name = name
			}
			var subnetObjs []kubermaticv1.Subnet
			if subnets, found, _ := unstructured.NestedStringSlice(item.Object, "status", "subnets"); found {

				for _, subnetName := range subnets {
					if subnetName == "join" {
						continue
					}
					subnet := &unstructured.Unstructured{}
					subnet.SetAPIVersion("kubeovn.io/v1") // adjust if needed
					subnet.SetKind("Subnet")              // adjust if needed
					if err := client.Get(ctx, ctrlruntimeclient.ObjectKey{Name: subnetName}, subnet); err != nil {
						return nil, fmt.Errorf("failed to get subnet %s: %w", subnetName, err)
					}
					cidr, found, _ := unstructured.NestedString(subnet.Object, "spec", "cidrBlock")
					if !found {
						// ("subnet %s does not have cidrBlock", subnetName)
						continue
					}
					// gateway, found, _ := unstructured.NestedStringSlice(item.Object, "spec", "gateway")
					// if !found {
					// 	return nil, fmt.Errorf("subnet %s does not have gateway", subnetName)
					// }
					subnetObjs = append(subnetObjs, kubermaticv1.Subnet{
						Name: subnetName,
						CIDR: cidr,
						// Gateway:    gateway,
						// ProviderID: subnet.Spec.ProviderID,
					})
				}
				vpc.Subnets = subnetObjs
			}
			defaultSettings.VPCs = append(defaultSettings.VPCs, vpc)
		}
	}
	// If VPC CRD does not exist, settings.VPCs will remain empty

	var scsList storagev1.StorageClassList
	if err := client.List(ctx, &scsList); err != nil {
		return nil, fmt.Errorf("failed to list storage classes: %w", err)
	}

	for _, sc := range scsList.Items {
		defaultSettings.StorageClasses = append(defaultSettings.StorageClasses, sc)
	}

	return defaultSettings, nil
}

func (k *KubeVirtProvider) BuildDefaultDatacenterSettings(defaultSettings *settings.DefaultDatacenterSettings) []settings.DatacenterSetting {
	modifiers := []settings.DatacenterSetting{}

	for _, setting := range defaultSettings.VPCs {
		for _, subnet := range setting.Subnets {
			modifiers = append(modifiers, settings.DatacenterSetting{
				Name:  fmt.Sprintf("with subnet %s in vpc %s", subnet.Name, setting.Name),
				Group: "vpc",
				Modifier: func(dc *kubermaticv1.Datacenter) {
					if dc.Spec.Kubevirt == nil {
						dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{}
					}
					if dc.Spec.Kubevirt.ProviderNetwork == nil {
						dc.Spec.Kubevirt.ProviderNetwork = &kubermaticv1.ProviderNetwork{}
					}
					dc.Spec.Kubevirt.ProviderNetwork.VPCs = []kubermaticv1.VPC{
						{
							Name:    setting.Name,
							Subnets: []kubermaticv1.Subnet{{Name: subnet.Name, CIDR: subnet.CIDR}},
						},
					}
				},
			})
		}
	}

	return modifiers
}

func (a KubeVirtProviderAdapter) MachineSettings(ctx context.Context, providerConfig *providerconfig.Config, namespace string, secrets legacytypes.Secrets, resources *options.ResourceSettings) []settings.MachineSpecModifier[any] {
	raw := a.inner.MachineSettings(ctx, providerConfig, namespace, secrets, resources)
	return ConvertModifiersToAny(raw)
}

func (k *KubeVirtProvider) GetDefaultConfig(secrets legacytypes.Secrets, distribution providerconfig.OperatingSystem, log *zap.SugaredLogger, clusterName string) (*kubevirt.RawConfig, error) {
	var err error

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	config, err := clientcmd.BuildConfigFromFlags("", secrets.Kubevirt.KubeconfigFile)
	if err != nil {
		panic(err)
	}

	infraClient, err := ctrlruntimeclient.New(config, ctrlruntimeclient.Options{
		Scheme: scheme,
	})
	if err != nil {
		panic(err)
	}
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

	osImage := "docker://quay.io/kubermatic-virt-disks/ubuntu:22.04"

	switch expression := distribution; expression {
	case providerconfig.OperatingSystemUbuntu:
		// Ubuntu specific settings
		osImage = "docker://quay.io/kubermatic-virt-disks/ubuntu:22.04"
	case providerconfig.OperatingSystemRHEL:
		// RHEL specific settings
		osImage = "docker://quay.io/kubermatic-virt-disks/rhel:8"
	case providerconfig.OperatingSystemFlatcar:
		// Flatcar specific settings
		osImage = "docker://quay.io/kubermatic-virt-disks/flatcar:3374.2.2"
	case providerconfig.OperatingSystemRockyLinux:
		// Rocky Linux specific settings
		osImage = "docker://quay.io/kubermatic-virt-disks/rocky:8"
	}

	return &kubevirt.RawConfig{
		ClusterName: providerconfig.ConfigVarString{Value: clusterName},
		Auth: kubevirt.Auth{
			Kubeconfig: providerconfig.ConfigVarString{Value: b64.StdEncoding.EncodeToString([]byte(secrets.Kubevirt.Kubeconfig))},
		},
		VirtualMachine: kubevirt.VirtualMachine{
			EnableNetworkMultiQueue: providerconfig.ConfigVarBool{Value: ptr.To(true)},
			Template: kubevirt.Template{
				CPUs:   providerconfig.ConfigVarString{Value: "2"},
				Memory: providerconfig.ConfigVarString{Value: "4096Mi"},
				PrimaryDisk: kubevirt.PrimaryDisk{
					OsImage: providerconfig.ConfigVarString{Value: osImage},
					Disk: kubevirt.Disk{
						Size:             providerconfig.ConfigVarString{Value: "20Gi"},
						StorageClassName: providerconfig.ConfigVarString{Value: defaultStorageClass.Name},
					},
				},
			},
		},
	}, nil
}

func (k *KubeVirtProvider) MachineSettings(ctx context.Context, providerConfig *providerconfig.Config, namespace string, secrets legacytypes.Secrets, resources *options.ResourceSettings) []settings.MachineSpecModifier[*kubevirt.RawConfig] {
	var err error

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	config, err := clientcmd.BuildConfigFromFlags("", secrets.Kubevirt.KubeconfigFile)
	if err != nil {
		panic(err)
	}

	client, err := ctrlruntimeclient.New(config, ctrlruntimeclient.Options{
		Scheme: scheme,
	})
	if err != nil {
		panic(err)
	}
	discoverdSettings, err := k.discoverDefaultMachineSettings(ctx, client, namespace)
	if err != nil {
		panic(fmt.Errorf("failed to discover default machine settings: %w", err))
	}
	generatedMachineSettings := k.buildDefaultMachineSettings(discoverdSettings)
	for _, settings := range settings.MachineSettings {
		generatedMachineSettings = append(generatedMachineSettings, settings)
	}

	return generatedMachineSettings
}

func (k KubeVirtProvider) MemoryModifiers(memories []string) []settings.MachineSpecModifier[*kubevirt.RawConfig] {
	var mods []settings.MachineSpecModifier[*kubevirt.RawConfig]
	for _, mem := range memories {
		mods = append(mods, settings.MachineSpecModifier[*kubevirt.RawConfig]{
			Name:  fmt.Sprintf("with %s memory", mem),
			Group: "memory",
			Modify: func(spec *kubevirt.RawConfig) {
				spec.VirtualMachine.Template.Memory.Value = mem
			},
		})
	}
	return mods
}

func (k KubeVirtProvider) CpuModifiers(cpus []int) []settings.MachineSpecModifier[*kubevirt.RawConfig] {
	var mods []settings.MachineSpecModifier[*kubevirt.RawConfig]
	for _, v := range cpus {
		v := v // capture range variable
		mods = append(mods, settings.MachineSpecModifier[*kubevirt.RawConfig]{
			Name:  fmt.Sprintf("with %d vCPUs", v),
			Group: "cpu",
			Modify: func(spec *kubevirt.RawConfig) {
				spec.VirtualMachine.Template.VCPUs.Cores = v
			},
		})
		mods = append(mods, settings.MachineSpecModifier[*kubevirt.RawConfig]{
			Name:  fmt.Sprintf("with %d CPUs", v),
			Group: "cpu",
			Modify: func(spec *kubevirt.RawConfig) {
				spec.VirtualMachine.Template.CPUs.Value = strconv.Itoa(v)
			},
		})
	}
	return mods
}

func (k KubeVirtProvider) DiskModifiers(sizes []string) []settings.MachineSpecModifier[*kubevirt.RawConfig] {
	var mods []settings.MachineSpecModifier[*kubevirt.RawConfig]
	for _, size := range sizes {
		size := size // capture range variable
		mods = append(mods, settings.MachineSpecModifier[*kubevirt.RawConfig]{
			Name:  fmt.Sprintf("with primary disk size %s", size),
			Group: "primary-disk-size",
			Modify: func(spec *kubevirt.RawConfig) {
				spec.VirtualMachine.Template.PrimaryDisk.Size.Value = size
			},
		})
		mods = append(mods, settings.MachineSpecModifier[*kubevirt.RawConfig]{
			Name:  fmt.Sprintf("with secondary disk size %s", size),
			Group: "secondary-disk-size",
			Modify: func(spec *kubevirt.RawConfig) {
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

func (k *KubeVirtProvider) discoverDefaultMachineSettings(ctx context.Context, client ctrlruntimeclient.Client, namespace string) (*DefaultMachineSettings, error) {
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

func (k *KubeVirtProvider) buildDefaultMachineSettings(defaultSettings *DefaultMachineSettings) []settings.MachineSpecModifier[*kubevirt.RawConfig] {
	var mods []settings.MachineSpecModifier[*kubevirt.RawConfig]

	// Storage classes
	for _, sc := range defaultSettings.StorageClasses {
		sc := sc // capture range variable
		mods = append(mods, settings.MachineSpecModifier[*kubevirt.RawConfig]{
			Name:  fmt.Sprintf("with primary disk storage class set to %s", sc),
			Group: "primary-disk-sc",
			Modify: func(spec *kubevirt.RawConfig) {
				if len(spec.VirtualMachine.Template.PrimaryDisk.Size.Value) == 0 {
					spec.VirtualMachine.Template.PrimaryDisk.Size.Value = "20Gi"
				}
				spec.VirtualMachine.Template.PrimaryDisk.StorageClassName.Value = sc
			},
		})
		mods = append(mods, settings.MachineSpecModifier[*kubevirt.RawConfig]{
			Name:  fmt.Sprintf("with secondary disk storage class set to %s", sc),
			Group: "secondary-disk-sc",
			Modify: func(spec *kubevirt.RawConfig) {
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
	for _, nodeName := range defaultSettings.NodeNames {
		nodeName := nodeName // capture range variable
		mods = append(mods, settings.MachineSpecModifier[*kubevirt.RawConfig]{
			Name:  fmt.Sprintf("with node affinity for hostname %s", nodeName),
			Group: "node-affinity",
			Modify: func(spec *kubevirt.RawConfig) {
				spec.Affinity.NodeAffinityPreset.Key.Value = "kubernetes.io/hostname"
				spec.Affinity.NodeAffinityPreset.Values = []providerconfig.ConfigVarString{{Value: nodeName}}
			},
		})
	}

	// Instance types
	for _, it := range defaultSettings.InstanceTypes {
		it := it // capture range variable
		mods = append(mods, settings.MachineSpecModifier[*kubevirt.RawConfig]{
			Name:  fmt.Sprintf("with instancetype %s", it),
			Group: "instancetype",
			Modify: func(spec *kubevirt.RawConfig) {
				spec.VirtualMachine.Instancetype = &kubevirtv1.InstancetypeMatcher{Name: it}
			},
		})
	}

	// Preferences
	for _, pref := range defaultSettings.Preferences {
		pref := pref // capture range variable
		mods = append(mods, settings.MachineSpecModifier[*kubevirt.RawConfig]{
			Name:  fmt.Sprintf("with preference %s", pref),
			Group: "preference",
			Modify: func(spec *kubevirt.RawConfig) {
				spec.VirtualMachine.Preference = &kubevirtv1.PreferenceMatcher{Name: pref}
			},
		})
	}

	return mods
}
