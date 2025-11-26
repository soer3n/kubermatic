package form

import (
	"k8c.io/machine-controller/sdk/providerconfig"
)

// getGCPTestSettings returns provider-specific test settings (placeholder for future use)
func (fd *FormData) getGCPTestSettings() []string {
	testSettings := GetTestSettingsForProvider(providerconfig.CloudProviderGoogle)
	var result []string
	for _, ts := range testSettings {
		result = append(result, ts.Description)
	}
	return result
}

// getGCPSecretFields returns the secret credential fields for GCP provider
func (fd *FormData) getGCPSecretFields() []SecretField {
	return []SecretField{
		{
			Name:     "KKPDatacenter",
			Label:    "GCP KKPDatacenter",
			Value:    &fd.Secrets.GCP.KKPDatacenter,
			Required: true,
		},
		{
			Name:     "ServiceAccount",
			Label:    "GCP ServiceAccount (JSON)",
			Value:    &fd.Secrets.GCP.ServiceAccount,
			Required: true,
		},
		{
			Name:     "Network",
			Label:    "GCP Network",
			Value:    &fd.Secrets.GCP.Network,
			Required: true,
		},
		{
			Name:     "Subnetwork",
			Label:    "GCP Subnetwork",
			Value:    &fd.Secrets.GCP.Subnetwork,
			Required: true,
		},
	}
}
