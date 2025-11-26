package form

import (
	"k8c.io/machine-controller/sdk/providerconfig"
)

// getAzureTestSettings returns provider-specific test settings (placeholder for future use)
func (fd *FormData) getAzureTestSettings() []string {
	testSettings := GetTestSettingsForProvider(providerconfig.CloudProviderAzure)
	var result []string
	for _, ts := range testSettings {
		result = append(result, ts.Description)
	}
	return result
}

// getAzureSecretFields returns the secret credential fields for Azure provider
func (fd *FormData) getAzureSecretFields() []SecretField {
	return []SecretField{
		{
			Name:     "KKPDatacenter",
			Label:    "Azure KKPDatacenter",
			Value:    &fd.Secrets.Azure.KKPDatacenter,
			Required: true,
		},
		{
			Name:     "ClientID",
			Label:    "Azure ClientID",
			Value:    &fd.Secrets.Azure.ClientID,
			Required: true,
		},
		{
			Name:     "ClientSecret",
			Label:    "Azure ClientSecret",
			Value:    &fd.Secrets.Azure.ClientSecret,
			Required: true,
		},
		{
			Name:     "TenantID",
			Label:    "Azure TenantID",
			Value:    &fd.Secrets.Azure.TenantID,
			Required: true,
		},
		{
			Name:     "SubscriptionID",
			Label:    "Azure SubscriptionID",
			Value:    &fd.Secrets.Azure.SubscriptionID,
			Required: true,
		},
	}
}
