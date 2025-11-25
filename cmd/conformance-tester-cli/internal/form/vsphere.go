package form

import (
	"github.com/charmbracelet/huh"
	"k8c.io/machine-controller/sdk/providerconfig"
)

func (fd *FormData) getVSphereTestSettings() []*huh.Group {
	return buildProviderFormGroups(
		providerDisplayMap[providerconfig.CloudProviderVsphere],
		providerconfig.CloudProviderVsphere,
		fd,
		fd.getVSphereSecretFields(),
	)
}
func (fd *FormData) getVSphereSecretFields() []huh.Field {
	return []huh.Field{
		huh.NewInput().Title("vSphere KKPDatacenter").Value(&fd.Secrets.VSphere.KKPDatacenter).Validate(fd.requiredIf(string(providerconfig.CloudProviderVsphere))),
		huh.NewInput().Title("vSphere Username").Value(&fd.Secrets.VSphere.Username).Validate(fd.requiredIf(string(providerconfig.CloudProviderVsphere))),
		huh.NewInput().Title("vSphere Password").Value(&fd.Secrets.VSphere.Password).Validate(fd.requiredIf(string(providerconfig.CloudProviderVsphere))),
	}
}
