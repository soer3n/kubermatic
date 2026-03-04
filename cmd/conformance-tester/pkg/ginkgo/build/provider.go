package build

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/build/provider"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/options"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/settings"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	kubermaticprovider "k8c.io/kubermatic/v2/pkg/provider"
	kubermatickubevirtprovider "k8c.io/kubermatic/v2/pkg/provider/cloud/kubevirt"
	mckubevirtprovider "k8c.io/machine-controller/pkg/cloudprovider/provider/kubevirt"
	cloudprovidertypes "k8c.io/machine-controller/pkg/cloudprovider/types"
	"k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	"k8c.io/machine-controller/sdk/cloudprovider/kubevirt"
	"k8c.io/machine-controller/sdk/providerconfig"
	"k8c.io/machine-controller/sdk/providerconfig/configvar"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

var defaultDatacenterSettings = settings.DatacenterSetting{
	Name: "default",
	Modifier: func(dc *kubermaticv1.Datacenter) {
		dc.Location = "Default datacenter"
	},
}

func getDefaultDatacenterSettings(ctx context.Context, providerConfig *providerconfig.Config, secrets legacytypes.Secrets) (settings.DatacenterSetting, error) {
	switch providerConfig.CloudProvider {
	case providerconfig.CloudProviderKubeVirt:
		kubevirtProvider := provider.KubeVirtProvider(providerconfig.CloudProviderKubeVirt)
		discoveredSettings, err := kubevirtProvider.DiscoverDefaultDatacenterSettings(ctx, providerConfig, secrets)
		if err != nil {
			return settings.DatacenterSetting{}, fmt.Errorf("failed to discover default datacenter settings: %w", err)
		}
		infraStorageClasses := []kubermaticv1.KubeVirtInfraStorageClass{}
		vpcs := []kubermaticv1.VPC{}

		for _, sc := range discoveredSettings.StorageClasses {
			newSc := kubermaticv1.KubeVirtInfraStorageClass{
				Name: sc.Name,
			}
			v, ok := sc.ObjectMeta.Annotations["storageclass.kubernetes.io/is-default-class"]
			if ok && v == "true" {
				newSc.IsDefaultClass = ptr.To(true)
			} else {
				newSc.IsDefaultClass = ptr.To(false)
			}
			infraStorageClasses = append(infraStorageClasses, newSc)
		}

		for _, vpc := range discoveredSettings.VPCs {
			vpcs = append(vpcs, kubermaticv1.VPC{
				Name:    vpc.Name,
				Subnets: vpc.Subnets,
			})
		}
		return settings.DatacenterSetting{
			Name: "default",
			Modifier: func(dc *kubermaticv1.Datacenter) {
				dc.Spec.Kubevirt = &kubermaticv1.DatacenterSpecKubevirt{
					ProviderNetwork: &kubermaticv1.ProviderNetwork{
						Name: "default",
						VPCs: vpcs,
					},
					InfraStorageClasses: infraStorageClasses,
					Images: kubermaticv1.KubeVirtImageSources{
						HTTP: &kubermaticv1.KubeVirtHTTPSource{
							OperatingSystems: map[providerconfig.OperatingSystem]kubermaticv1.OSVersions{
								providerconfig.OperatingSystemUbuntu: {
									"20.04": "docker://quay.io/kubermatic-virt-disks/ubuntu:20.04",
									"22.04": "docker://quay.io/kubermatic-virt-disks/ubuntu:22.04",
								},
								providerconfig.OperatingSystemRHEL: {
									"8": "docker://quay.io/kubermatic-virt-disks/rhel:8",
									"9": "docker://quay.io/kubermatic-virt-disks/rhel:9",
								},
								providerconfig.OperatingSystemFlatcar: {
									"3374.2.2": "docker://quay.io/kubermatic-virt-disks/flatcar:3374.2.2",
								},
								providerconfig.OperatingSystemRockyLinux: {
									"8": "docker://quay.io/kubermatic-virt-disks/rocky:8",
									"9": "docker://quay.io/kubermatic-virt-disks/rocky:9",
								},
							},
						},
					},
				}
			},
		}, nil
	}
	return settings.DatacenterSetting{}, nil
}

func getProviderConfig(ctx context.Context, log *zap.SugaredLogger, secrets legacytypes.Secrets, distribution providerconfig.OperatingSystem, cloudProvider providerconfig.CloudProvider) (providerConfig *providerconfig.Config, err error) {
	switch cloudProvider {
	case providerconfig.CloudProviderKubeVirt:
		kubevirtProvider := provider.KubeVirtProvider(providerconfig.CloudProviderKubeVirt)
		rawConfig, err := kubevirtProvider.GetDefaultConfig(secrets, distribution, log, "test-cluster")
		if err != nil {
			return nil, err
		}
		return &providerconfig.Config{
			CloudProvider:     providerconfig.CloudProviderKubeVirt,
			CloudProviderSpec: runtime.RawExtension{Raw: toJSON(rawConfig)},
		}, nil
	}
	return nil, nil
}

func getProviderSpec(log *zap.SugaredLogger, secrets legacytypes.Secrets, distribution providerconfig.OperatingSystem, cloudProvider providerconfig.CloudProvider) (any, error) {
	switch cloudProvider {
	case providerconfig.CloudProviderKubeVirt:
		kubevirtProvider := provider.KubeVirtProvider(providerconfig.CloudProviderKubeVirt)
		rawConfig, err := kubevirtProvider.GetDefaultConfig(secrets, distribution, log, "test-cluster")
		if err != nil {
			return nil, err
		}
		return rawConfig, nil
	}
	return nil, nil
}

// getProviderSpecBytes returns the JSON-serialized provider spec for caching.
func getProviderSpecBytes(log *zap.SugaredLogger, secrets legacytypes.Secrets, distribution providerconfig.OperatingSystem, cloudProvider providerconfig.CloudProvider) ([]byte, error) {
	ps, err := getProviderSpec(log, secrets, distribution, cloudProvider)
	if err != nil {
		return nil, err
	}
	return json.Marshal(ps)
}

// newProviderSpecFromCache unmarshals cached JSON bytes into a fresh provider spec instance.
func newProviderSpecFromCache(cachedBytes []byte, cloudProvider providerconfig.CloudProvider) (any, error) {
	switch cloudProvider {
	case providerconfig.CloudProviderKubeVirt:
		var config kubevirt.RawConfig
		if err := json.Unmarshal(cachedBytes, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cached kubevirt config: %w", err)
		}
		return &config, nil
	}
	return nil, fmt.Errorf("unsupported cloud provider: %s", cloudProvider)
}

func getProvider(provider providerconfig.CloudProvider, resolver *configvar.Resolver) cloudprovidertypes.Provider {
	switch provider {
	case providerconfig.CloudProviderKubeVirt:
		return mckubevirtprovider.New(resolver)
	}
	return nil
}

func getClusterProvider(provider providerconfig.CloudProvider, dcName string, dc *kubermaticv1.Datacenter, secrets legacytypes.Secrets) (kubermaticprovider.CloudProvider, kubermaticv1.CloudSpec, error) {
	cloud := kubermaticv1.CloudSpec{}
	cloud.ProviderName = string(kubermaticv1.KubevirtCloudProvider)
	cloud.DatacenterName = dcName

	switch provider {
	case providerconfig.CloudProviderKubeVirt:
		c, err := kubermatickubevirtprovider.NewCloudProvider(dc, nil)
		if err != nil {
			return nil, cloud, fmt.Errorf("failed to create kubevirt cloud provider: %w", err)
		}
		cloud.Kubevirt = &kubermaticv1.KubevirtCloudSpec{
			Kubeconfig: secrets.Kubevirt.Kubeconfig,
		}
		return c, cloud, nil
	}
	return nil, cloud, nil
}

func getDefaultMachineSpec(ctx context.Context, log *zap.SugaredLogger, providerConfig *providerconfig.Config, secrets legacytypes.Secrets) (*v1alpha1.MachineSpec, error) {
	switch providerConfig.CloudProvider {
	case providerconfig.CloudProviderKubeVirt:
		// getDefaultMachineSpec returns the default machine spec.
		return &v1alpha1.MachineSpec{
			ProviderSpec: v1alpha1.ProviderSpec{
				Value: &providerConfig.CloudProviderSpec,
			},
		}, nil
	}
	return nil, nil
}

func MachineSettings(ctx context.Context, providerConfig *providerconfig.Config, namespace string, secrets legacytypes.Secrets, resources *options.ResourceSettings) []settings.MachineSpecModifier[any] {
	switch providerConfig.CloudProvider {
	case providerconfig.CloudProviderKubeVirt:
		kubevirtProvider := provider.KubeVirtProvider(providerconfig.CloudProviderKubeVirt)
		return provider.ConvertModifiersToAny(kubevirtProvider.MachineSettings(ctx, providerConfig, namespace, secrets, resources))
	}
	return nil
}

func ResourceMachineSettings(ctx context.Context, providerConfig *providerconfig.Config, namespace string, secrets legacytypes.Secrets, resources *options.ResourceSettings) []settings.MachineSpecModifier[any] {
	var p settings.ProviderInterface
	var generatedMachineSettings []settings.MachineSpecModifier[any]
	switch providerConfig.CloudProvider {
	case providerconfig.CloudProviderKubeVirt:
		if resources == nil {
			return provider.ConvertModifiersToAny([]settings.MachineSpecModifier[*kubevirt.RawConfig]{
				{
					Name:  "with custom cpu and memory",
					Group: "custom-resources",
					Modify: func(spec *kubevirt.RawConfig) {
						// No-op, just a placeholder to indicate default resources.
					},
				},
			})
		}
		p = provider.NewKubeVirtProviderAdapter(provider.KubeVirtProvider(providerconfig.CloudProviderKubeVirt))
	}
	if resources.Memory != nil {
		for _, settings := range p.MemoryModifiers(resources.Memory) {
			generatedMachineSettings = append(generatedMachineSettings, settings)
		}
	}
	if resources.Cpu != nil {
		for _, settings := range p.CpuModifiers(resources.Cpu) {
			generatedMachineSettings = append(generatedMachineSettings, settings)
		}
	}
	if resources.DiskSize != nil {
		for _, settings := range p.DiskModifiers(resources.DiskSize) {
			generatedMachineSettings = append(generatedMachineSettings, settings)
		}
	}
	return generatedMachineSettings
}

func GenericDatacenterSettings(ctx context.Context, providerConfig *providerconfig.Config, secrets legacytypes.Secrets) []settings.DatacenterSetting {
	var p settings.ProviderInterface
	switch providerConfig.CloudProvider {
	case providerconfig.CloudProviderKubeVirt:
		p = provider.NewKubeVirtProviderAdapter(provider.KubeVirtProvider(providerconfig.CloudProviderKubeVirt))
	}
	discoverdSettings, err := p.DiscoverDefaultDatacenterSettings(ctx, providerConfig, secrets)
	if err != nil {
		panic(fmt.Errorf("failed to discover default datacenter settings: %w", err))
	}
	generatedDatacenterSettings := p.BuildDefaultDatacenterSettings(discoverdSettings)
	for _, settings := range settings.DatacenterSettings {
		generatedDatacenterSettings = append(generatedDatacenterSettings, settings)
	}
	return generatedDatacenterSettings
}
