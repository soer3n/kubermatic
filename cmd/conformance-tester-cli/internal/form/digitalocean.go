package form

import (
	"github.com/charmbracelet/huh"
	"k8c.io/machine-controller/sdk/providerconfig"
)

func (fd *FormData) getDigitalOceanTestSettings() []*huh.Group {
	return buildProviderFormGroups(
		providerDisplayMap[providerconfig.CloudProviderDigitalocean],
		providerconfig.CloudProviderDigitalocean,
		fd,
		fd.getDigitalOceanSecretFields(),
	)
}
func (fd *FormData) getDigitalOceanSecretFields() []huh.Field {
	return []huh.Field{
		huh.NewInput().Title("DigitalOcean KKPDatacenter").Value(&fd.Secrets.Digitalocean.KKPDatacenter).Validate(fd.requiredIf("digitalocean")),
		huh.NewInput().Title("DigitalOcean Token").Value(&fd.Secrets.Digitalocean.Token).Validate(fd.requiredIf("digitalocean")),
	}
}
