package build

import (
	"context"
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
)

var defaultDatacenterSettings = settings.DatacenterSetting{
	Name: "default",
	Modifier: func(dc *kubermaticv1.Datacenter) {
		dc.Location = "Default datacenter"
	},
}

func getProviderConfig(ctx context.Context, log *zap.SugaredLogger, secrets legacytypes.Secrets, cloudProvider providerconfig.CloudProvider) (providerConfig *providerconfig.Config, err error) {
	switch cloudProvider {
	case providerconfig.CloudProviderKubeVirt:
		kubevirtProvider := provider.KubeVirtProvider(providerconfig.CloudProviderKubeVirt)
		rawConfig, err := kubevirtProvider.GetDefaultConfig(secrets, log)
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

func getProviderSpec(log *zap.SugaredLogger, secrets legacytypes.Secrets, cloudProvider providerconfig.CloudProvider) (any, error) {
	switch cloudProvider {
	case providerconfig.CloudProviderKubeVirt:
		kubevirtProvider := provider.KubeVirtProvider(providerconfig.CloudProviderKubeVirt)
		rawConfig, err := kubevirtProvider.GetDefaultConfig(secrets, log)
		if err != nil {
			return nil, err
		}
		return rawConfig, nil
	}
	return nil, nil
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
