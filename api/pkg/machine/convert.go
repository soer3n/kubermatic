package machine

import (
	"fmt"

	apiv1 "github.com/kubermatic/kubermatic/api/pkg/api/v1"
	"github.com/kubermatic/machine-controller/pkg/cloudprovider/provider/aws"
	"github.com/kubermatic/machine-controller/pkg/cloudprovider/provider/azure"
	"github.com/kubermatic/machine-controller/pkg/cloudprovider/provider/digitalocean"
	"github.com/kubermatic/machine-controller/pkg/cloudprovider/provider/gce"
	"github.com/kubermatic/machine-controller/pkg/cloudprovider/provider/hetzner"
	"github.com/kubermatic/machine-controller/pkg/cloudprovider/provider/openstack"
	"github.com/kubermatic/machine-controller/pkg/cloudprovider/provider/packet"
	"github.com/kubermatic/machine-controller/pkg/cloudprovider/provider/vsphere"
	"github.com/kubermatic/machine-controller/pkg/providerconfig"
	"github.com/kubermatic/machine-controller/pkg/userdata/centos"
	"github.com/kubermatic/machine-controller/pkg/userdata/coreos"
	"github.com/kubermatic/machine-controller/pkg/userdata/ubuntu"

	"k8s.io/apimachinery/pkg/util/json"

	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// GetAPIV1OperatingSystemSpec returns the api compatible OperatingSystemSpec for the given machine
func GetAPIV1OperatingSystemSpec(machineSpec clusterv1alpha1.MachineSpec) (*apiv1.OperatingSystemSpec, error) {
	decodedProviderSpec, err := providerconfig.GetConfig(machineSpec.ProviderSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to get machine providerConfig: %v", err)
	}

	operatingSystemSpec := &apiv1.OperatingSystemSpec{}

	if decodedProviderSpec.OperatingSystem == providerconfig.OperatingSystemCoreos {
		config := &coreos.Config{}
		if err := json.Unmarshal(decodedProviderSpec.OperatingSystemSpec.Raw, &config); err != nil {
			return nil, fmt.Errorf("failed to parse coreos config: %v", err)
		}
		operatingSystemSpec.ContainerLinux = &apiv1.ContainerLinuxSpec{
			DisableAutoUpdate: config.DisableAutoUpdate,
		}
	} else if decodedProviderSpec.OperatingSystem == providerconfig.OperatingSystemUbuntu {
		config := &ubuntu.Config{}
		if err := json.Unmarshal(decodedProviderSpec.OperatingSystemSpec.Raw, &config); err != nil {
			return nil, fmt.Errorf("failed to parse ubuntu config: %v", err)
		}
		operatingSystemSpec.Ubuntu = &apiv1.UbuntuSpec{
			DistUpgradeOnBoot: config.DistUpgradeOnBoot,
		}
	} else if decodedProviderSpec.OperatingSystem == providerconfig.OperatingSystemCentOS {
		config := &centos.Config{}
		if err := json.Unmarshal(decodedProviderSpec.OperatingSystemSpec.Raw, &config); err != nil {
			return nil, fmt.Errorf("failed to parse centos config: %v", err)
		}
		operatingSystemSpec.CentOS = &apiv1.CentOSSpec{
			DistUpgradeOnBoot: config.DistUpgradeOnBoot,
		}
	}

	return operatingSystemSpec, nil
}

// GetAPIV2NodeCloudSpec returns the api compatible NodeCloudSpec for the given machine
func GetAPIV2NodeCloudSpec(machineSpec clusterv1alpha1.MachineSpec) (*apiv1.NodeCloudSpec, error) {
	decodedProviderSpec, err := providerconfig.GetConfig(machineSpec.ProviderSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to get machine providerConfig: %v", err)
	}

	cloudSpec := &apiv1.NodeCloudSpec{}

	switch decodedProviderSpec.CloudProvider {
	case providerconfig.CloudProviderAWS:
		config := &aws.RawConfig{}
		if err := json.Unmarshal(decodedProviderSpec.CloudProviderSpec.Raw, &config); err != nil {
			return nil, fmt.Errorf("failed to parse aws config: %v", err)
		}
		cloudSpec.AWS = &apiv1.AWSNodeSpec{
			Tags:             config.Tags,
			VolumeSize:       config.DiskSize,
			VolumeType:       config.DiskType.Value,
			InstanceType:     config.InstanceType.Value,
			AMI:              config.AMI.Value,
			AvailabilityZone: config.AvailabilityZone.Value,
			SubnetID:         config.SubnetID.Value,
		}
	case providerconfig.CloudProviderAzure:
		config := &azure.RawConfig{}
		if err := json.Unmarshal(decodedProviderSpec.CloudProviderSpec.Raw, &config); err != nil {
			return nil, fmt.Errorf("failed to parse Azure config: %v", err)
		}
		cloudSpec.Azure = &apiv1.AzureNodeSpec{
			Size:           config.VMSize.Value,
			AssignPublicIP: config.AssignPublicIP.Value,
			Tags:           config.Tags,
		}
	case providerconfig.CloudProviderDigitalocean:
		config := &digitalocean.RawConfig{}
		if err := json.Unmarshal(decodedProviderSpec.CloudProviderSpec.Raw, &config); err != nil {
			return nil, fmt.Errorf("failed to parse digitalocean config: %v", err)
		}
		cloudSpec.Digitalocean = &apiv1.DigitaloceanNodeSpec{
			IPv6:       config.IPv6.Value,
			Size:       config.Size.Value,
			Backups:    config.Backups.Value,
			Monitoring: config.Monitoring.Value,
		}
		for _, v := range config.Tags {
			cloudSpec.Digitalocean.Tags = append(cloudSpec.Digitalocean.Tags, v.Value)
		}
	case providerconfig.CloudProviderOpenstack:
		config := &openstack.RawConfig{}
		if err := json.Unmarshal(decodedProviderSpec.CloudProviderSpec.Raw, &config); err != nil {
			return nil, fmt.Errorf("failed to parse openstack config: %v", err)
		}
		cloudSpec.Openstack = &apiv1.OpenstackNodeSpec{
			Flavor: config.Flavor.Value,
			Image:  config.Image.Value,
			Tags:   config.Tags,
		}
	case providerconfig.CloudProviderHetzner:
		config := &hetzner.RawConfig{}
		if err := json.Unmarshal(decodedProviderSpec.CloudProviderSpec.Raw, &config); err != nil {
			return nil, fmt.Errorf("failed to parse hetzner config: %v", err)
		}
		cloudSpec.Hetzner = &apiv1.HetznerNodeSpec{
			Type: config.ServerType.Value,
		}
	case providerconfig.CloudProviderVsphere:
		config := &vsphere.RawConfig{}
		if err := json.Unmarshal(decodedProviderSpec.CloudProviderSpec.Raw, &config); err != nil {
			return nil, fmt.Errorf("failed to parse vsphere config: %v", err)
		}
		cloudSpec.VSphere = &apiv1.VSphereNodeSpec{
			CPUs:       int(config.CPUs),
			Memory:     int(config.MemoryMB),
			DiskSizeGB: config.DiskSizeGB,
			Template:   config.TemplateVMName.Value,
		}
	case providerconfig.CloudProviderPacket:
		config := &packet.RawConfig{}
		if err := json.Unmarshal(decodedProviderSpec.CloudProviderSpec.Raw, &config); err != nil {
			return nil, fmt.Errorf("failed to parse packet config: %v", err)
		}
		cloudSpec.Packet = &apiv1.PacketNodeSpec{
			InstanceType: config.InstanceType.Value,
		}
		for _, v := range config.Tags {
			cloudSpec.Packet.Tags = append(cloudSpec.Packet.Tags, v.Value)
		}
	case providerconfig.CloudProviderGoogle:
		config := &gce.CloudProviderSpec{}
		if err := json.Unmarshal(decodedProviderSpec.CloudProviderSpec.Raw, &config); err != nil {
			return nil, fmt.Errorf("failed to parse gcp config: %v", err)
		}
		cloudSpec.GCP = &apiv1.GCPNodeSpec{
			Zone:        config.Zone.Value,
			MachineType: config.MachineType.Value,
			DiskSize:    config.DiskSize,
			DiskType:    config.DiskType.Value,
			Preemptible: config.Preemptible.Value,
			Labels:      config.Labels,
			Tags:        config.Tags,
		}
	default:
		return nil, fmt.Errorf("unknown cloud provider %q", decodedProviderSpec.CloudProvider)
	}

	return cloudSpec, nil
}
