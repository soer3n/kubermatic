package form

import (
	"github.com/charmbracelet/huh"
	"k8c.io/machine-controller/sdk/providerconfig"
)

func (fd *FormData) getAnexiaTestSettings() []*huh.Group {
	return buildProviderFormGroups(
		providerDisplayMap[providerconfig.CloudProviderAnexia],
		providerconfig.CloudProviderAnexia,
		fd,
		fd.getAnexiaSecretFields(),
	)
}
func (fd *FormData) getAnexiaSecretFields() []huh.Field {
	return []huh.Field{
		huh.NewInput().Title("Anexia KKPDatacenter").Value(&fd.Secrets.Anexia.KKPDatacenter).Validate(fd.requiredIf(string(providerconfig.CloudProviderAnexia))),
		huh.NewInput().Title("Anexia Token").Value(&fd.Secrets.Anexia.Token).Validate(fd.requiredIf(string(providerconfig.CloudProviderAnexia))),
		huh.NewInput().Title("Anexia TemplateID").Value(&fd.Secrets.Anexia.TemplateID).Validate(fd.requiredIf(string(providerconfig.CloudProviderAnexia))),
		huh.NewInput().Title("Anexia VlanID").Value(&fd.Secrets.Anexia.VlanID).Validate(fd.requiredIf(string(providerconfig.CloudProviderAnexia))),
	}
}
