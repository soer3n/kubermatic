package form

import (
	"github.com/charmbracelet/huh"
	"k8c.io/machine-controller/sdk/providerconfig"
)

func (fd *FormData) getNutanixTestSettings() []*huh.Group {
	return buildProviderFormGroups(
		providerDisplayMap[providerconfig.CloudProviderNutanix],
		providerconfig.CloudProviderNutanix,
		fd,
		fd.getNutanixSecretFields(),
	)
}
func (fd *FormData) getNutanixSecretFields() []huh.Field {
	return []huh.Field{
		huh.NewInput().Title("Nutanix KKPDatacenter").Value(&fd.Secrets.Nutanix.KKPDatacenter).Validate(fd.requiredIf(string(providerconfig.CloudProviderNutanix))),
		huh.NewInput().Title("Nutanix Username").Value(&fd.Secrets.Nutanix.Username).Validate(fd.requiredIf(string(providerconfig.CloudProviderNutanix))),
		huh.NewInput().Title("Nutanix Password").Value(&fd.Secrets.Nutanix.Password).Validate(fd.requiredIf(string(providerconfig.CloudProviderNutanix))),
		huh.NewInput().Title("Nutanix CSI Username").Value(&fd.Secrets.Nutanix.CSIUsername).Validate(fd.requiredIf(string(providerconfig.CloudProviderNutanix))),
		huh.NewInput().Title("Nutanix CSI Password").Value(&fd.Secrets.Nutanix.CSIPassword).Validate(fd.requiredIf(string(providerconfig.CloudProviderNutanix))),
		huh.NewInput().Title("Nutanix CSI Endpoint").Value(&fd.Secrets.Nutanix.CSIEndpoint).Validate(fd.requiredIf(string(providerconfig.CloudProviderNutanix))),
		huh.NewInput().Title("Nutanix CSI Port").Value(&fd.NutanixCSIPortStr).Validate(fd.requiredIf(string(providerconfig.CloudProviderNutanix))),
		huh.NewInput().Title("Nutanix Proxy URL").Value(&fd.Secrets.Nutanix.ProxyURL).Validate(fd.requiredIf(string(providerconfig.CloudProviderNutanix))),
		huh.NewInput().Title("Nutanix Cluster Name").Value(&fd.Secrets.Nutanix.ClusterName).Validate(fd.requiredIf(string(providerconfig.CloudProviderNutanix))),
		huh.NewInput().Title("Nutanix Project Name").Value(&fd.Secrets.Nutanix.ProjectName).Validate(fd.requiredIf(string(providerconfig.CloudProviderNutanix))),
		huh.NewInput().Title("Nutanix Subnet Name").Value(&fd.Secrets.Nutanix.SubnetName).Validate(fd.requiredIf(string(providerconfig.CloudProviderNutanix))),
	}
}
