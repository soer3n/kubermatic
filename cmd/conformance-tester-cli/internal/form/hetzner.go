package form

import (
	"github.com/charmbracelet/huh"
	"k8c.io/machine-controller/sdk/providerconfig"
)

func (fd *FormData) getHetznerTestSettings() []*huh.Group {
	return buildProviderFormGroups(
		providerDisplayMap[providerconfig.CloudProviderHetzner],
		providerconfig.CloudProviderHetzner,
		fd,
		fd.getHetznerSecretFields(),
	)
}
func (fd *FormData) getHetznerSecretFields() []huh.Field {
	return []huh.Field{
		huh.NewInput().Title("Hetzner KKPDatacenter").Value(&fd.Secrets.Hetzner.KKPDatacenter).Validate(fd.requiredIf(string(providerconfig.CloudProviderHetzner))),
		huh.NewInput().Title("Hetzner Token").Value(&fd.Secrets.Hetzner.Token).Validate(fd.requiredIf(string(providerconfig.CloudProviderHetzner))),
	}
}
