package form

import (
	"k8c.io/machine-controller/sdk/providerconfig"
)

// getHetznerTestSettings returns provider-specific test settings (placeholder for future use)
func (fd *FormData) getHetznerTestSettings() []string {
	testSettings := GetTestSettingsForProvider(providerconfig.CloudProviderHetzner)
	var result []string
	for _, ts := range testSettings {
		result = append(result, ts.Description)
	}
	return result
}

// getHetznerSecretFields returns the secret credential fields for Hetzner provider
func (fd *FormData) getHetznerSecretFields() []SecretField {
	return []SecretField{
		{
			Name:     "KKPDatacenter",
			Label:    "Hetzner KKPDatacenter",
			Value:    &fd.Secrets.Hetzner.KKPDatacenter,
			Required: true,
		},
		{
			Name:     "Token",
			Label:    "Hetzner Token",
			Value:    &fd.Secrets.Hetzner.Token,
			Required: true,
		},
	}
}
