package form

import (
	"github.com/charmbracelet/huh"
	"k8c.io/machine-controller/sdk/providerconfig"
)

func (fd *FormData) getAlibabaTestSettings() []*huh.Group {
	return buildProviderFormGroups(
		providerDisplayMap[providerconfig.CloudProviderAlibaba],
		providerconfig.CloudProviderAlibaba,
		fd,
		fd.getAlibabaSecretFields(),
	)
}
func (fd *FormData) getAlibabaSecretFields() []huh.Field {
	return []huh.Field{
		huh.NewInput().Title("Alibaba KKPDatacenter").Value(&fd.Secrets.Alibaba.KKPDatacenter).Validate(fd.requiredIf(string(providerconfig.CloudProviderAlibaba))),
		huh.NewInput().Title("Alibaba AccessKeyID").Value(&fd.Secrets.Alibaba.AccessKeyID).Validate(fd.requiredIf(string(providerconfig.CloudProviderAlibaba))),
		huh.NewInput().Title("Alibaba AccessKeySecret").Value(&fd.Secrets.Alibaba.AccessKeySecret).Validate(fd.requiredIf(string(providerconfig.CloudProviderAlibaba))),
	}
}
