package form

import (
	"github.com/charmbracelet/huh"
	"k8c.io/machine-controller/sdk/providerconfig"
)

func (fd *FormData) getAzureTestSettings() []*huh.Group {
	return buildProviderFormGroups(
		providerDisplayMap[providerconfig.CloudProviderAzure],
		providerconfig.CloudProviderAzure,
		fd,
		fd.getAzureSecretFields(),
	)
}
func (fd *FormData) getAzureSecretFields() []huh.Field {
	return []huh.Field{
		huh.NewInput().Title("Azure KKPDatacenter").Value(&fd.Secrets.Azure.KKPDatacenter).Validate(fd.requiredIf(string(providerconfig.CloudProviderAzure))),
		huh.NewInput().Title("Azure ClientID").Value(&fd.Secrets.Azure.ClientID).Validate(fd.requiredIf(string(providerconfig.CloudProviderAzure))),
		huh.NewInput().Title("Azure ClientSecret").Value(&fd.Secrets.Azure.ClientSecret).Validate(fd.requiredIf(string(providerconfig.CloudProviderAzure))),
		huh.NewInput().Title("Azure TenantID").Value(&fd.Secrets.Azure.TenantID).Validate(fd.requiredIf(string(providerconfig.CloudProviderAzure))),
		huh.NewInput().Title("Azure SubscriptionID").Value(&fd.Secrets.Azure.SubscriptionID).Validate(fd.requiredIf(string(providerconfig.CloudProviderAzure))),
	}
}
