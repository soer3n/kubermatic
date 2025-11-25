package form

import (
	"github.com/charmbracelet/huh"
	"k8c.io/machine-controller/sdk/providerconfig"
)

func (fd *FormData) getAWSTestSettings() []*huh.Group {
	return buildProviderFormGroups(
		providerDisplayMap[providerconfig.CloudProviderAWS],
		providerconfig.CloudProviderAWS,
		fd,
		fd.getAWSSecretFields(),
	)
}
func (fd *FormData) getAWSSecretFields() []huh.Field {
	return []huh.Field{
		huh.NewInput().Title("AWS KKPDatacenter").Value(&fd.Secrets.AWS.KKPDatacenter).Validate(fd.requiredIf(string(providerconfig.CloudProviderAWS))),
		huh.NewInput().Title("AWS AccessKeyID").Value(&fd.Secrets.AWS.AccessKeyID).Validate(fd.requiredIf(string(providerconfig.CloudProviderAWS))),
		huh.NewInput().Title("AWS SecretAccessKey").Value(&fd.Secrets.AWS.SecretAccessKey).Validate(fd.requiredIf(string(providerconfig.CloudProviderAWS))),
	}
}
