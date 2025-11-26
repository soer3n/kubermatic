package form

import (
	"k8c.io/machine-controller/sdk/providerconfig"
)

// getDigitalOceanTestSettings returns provider-specific test settings (placeholder for future use)
func (fd *FormData) getDigitalOceanTestSettings() []string {
	testSettings := GetTestSettingsForProvider(providerconfig.CloudProviderDigitalocean)
	var result []string
	for _, ts := range testSettings {
		result = append(result, ts.Description)
	}
	return result
}

// getDigitalOceanSecretFields returns the secret credential fields for DigitalOcean provider
func (fd *FormData) getDigitalOceanSecretFields() []SecretField {
	return []SecretField{
		{
			Name:     "KKPDatacenter",
			Label:    "DigitalOcean KKPDatacenter",
			Value:    &fd.Secrets.Digitalocean.KKPDatacenter,
			Required: true,
		},
		{
			Name:     "Token",
			Label:    "DigitalOcean Token",
			Value:    &fd.Secrets.Digitalocean.Token,
			Required: true,
		},
	}
}
