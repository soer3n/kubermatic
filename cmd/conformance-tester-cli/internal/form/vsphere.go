package form

import (
	"k8c.io/machine-controller/sdk/providerconfig"
)

// getVSphereTestSettings returns provider-specific test settings (placeholder for future use)
func (fd *FormData) getVSphereTestSettings() []string {
	testSettings := GetTestSettingsForProvider(providerconfig.CloudProviderVsphere)
	var result []string
	for _, ts := range testSettings {
		result = append(result, ts.Description)
	}
	return result
}

// getVSphereSecretFields returns the secret credential fields for vSphere provider
func (fd *FormData) getVSphereSecretFields() []SecretField {
	return []SecretField{
		{
			Name:     "KKPDatacenter",
			Label:    "vSphere KKPDatacenter",
			Value:    &fd.Secrets.VSphere.KKPDatacenter,
			Required: true,
		},
		{
			Name:     "Username",
			Label:    "vSphere Username",
			Value:    &fd.Secrets.VSphere.Username,
			Required: true,
		},
		{
			Name:     "Password",
			Label:    "vSphere Password",
			Value:    &fd.Secrets.VSphere.Password,
			Required: true,
		},
	}
}
