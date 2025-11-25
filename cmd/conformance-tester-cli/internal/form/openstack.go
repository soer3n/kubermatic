package form

import (
	"github.com/charmbracelet/huh"
	"k8c.io/machine-controller/sdk/providerconfig"
)

func (fd *FormData) getOpenStackTestSettings() []*huh.Group {
	return buildProviderFormGroups(
		providerDisplayMap[providerconfig.CloudProviderOpenstack],
		providerconfig.CloudProviderOpenstack,
		fd,
		fd.getOpenStackSecretFields(),
	)
}
func (fd *FormData) getOpenStackSecretFields() []huh.Field {
	return []huh.Field{
		huh.NewInput().Title("OpenStack KKPDatacenter").Value(&fd.Secrets.OpenStack.KKPDatacenter).Validate(fd.requiredIf(string(providerconfig.CloudProviderOpenstack))),
		huh.NewInput().Title("OpenStack Domain").Value(&fd.Secrets.OpenStack.Domain).Validate(fd.requiredIf(string(providerconfig.CloudProviderOpenstack))),
		huh.NewInput().Title("OpenStack Project").Value(&fd.Secrets.OpenStack.Project).Validate(fd.requiredIf(string(providerconfig.CloudProviderOpenstack))),
		huh.NewInput().Title("OpenStack ProjectID").Value(&fd.Secrets.OpenStack.ProjectID).Validate(fd.requiredIf(string(providerconfig.CloudProviderOpenstack))),
		huh.NewInput().Title("OpenStack Username").Value(&fd.Secrets.OpenStack.Username).Validate(fd.requiredIf(string(providerconfig.CloudProviderOpenstack))),
		huh.NewInput().Title("OpenStack Password").Value(&fd.Secrets.OpenStack.Password).Validate(fd.requiredIf(string(providerconfig.CloudProviderOpenstack))),
	}
}
