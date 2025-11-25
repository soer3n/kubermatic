package form

import (
	"github.com/charmbracelet/huh"
	"k8c.io/machine-controller/sdk/providerconfig"
)

func (fd *FormData) getVMwareCloudDirectorTestSettings() []*huh.Group {
	return buildProviderFormGroups(
		providerDisplayMap[providerconfig.CloudProviderVMwareCloudDirector],
		providerconfig.CloudProviderVMwareCloudDirector,
		fd,
		fd.getVDCSecretFields(),
	)
}
func (fd *FormData) getVDCSecretFields() []huh.Field {
	return []huh.Field{
		huh.NewInput().Title("VMware Cloud Director KKPDatacenter").Value(&fd.Secrets.VMwareCloudDirector.KKPDatacenter).Validate(fd.requiredIf(string(providerconfig.CloudProviderVMwareCloudDirector))),
		huh.NewInput().Title("VCD Username").Value(&fd.Secrets.VMwareCloudDirector.Username).Validate(fd.requiredIf(string(providerconfig.CloudProviderVMwareCloudDirector))),
		huh.NewInput().Title("VCD Password").Value(&fd.Secrets.VMwareCloudDirector.Password).Validate(fd.requiredIf(string(providerconfig.CloudProviderVMwareCloudDirector))),
		huh.NewInput().Title("VCD Organization").Value(&fd.Secrets.VMwareCloudDirector.Organization).Validate(fd.requiredIf(string(providerconfig.CloudProviderVMwareCloudDirector))),
		huh.NewInput().Title("VCD VDC").Value(&fd.Secrets.VMwareCloudDirector.VDC).Validate(fd.requiredIf(string(providerconfig.CloudProviderVMwareCloudDirector))),
		huh.NewInput().Title("VCD OVDC Networks (comma-separated)").Value(&fd.VMCDNetworksStr).Validate(fd.requiredIf(string(providerconfig.CloudProviderVMwareCloudDirector))),
	}
}
