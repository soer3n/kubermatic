package form

import (
	"k8c.io/machine-controller/sdk/providerconfig"
)

// getOpenStackTestSettings returns provider-specific test settings (placeholder for future use)
func (fd *FormData) getOpenStackTestSettings() []string {
	testSettings := GetTestSettingsForProvider(providerconfig.CloudProviderOpenstack)
	var result []string
	for _, ts := range testSettings {
		result = append(result, ts.Description)
	}
	return result
}

// getOpenStackSecretFields returns the secret credential fields for OpenStack provider
func (fd *FormData) getOpenStackSecretFields() []SecretField {
	return []SecretField{
		{
			Name:     "KKPDatacenter",
			Label:    "OpenStack KKPDatacenter",
			Value:    &fd.Secrets.OpenStack.KKPDatacenter,
			Required: true,
		},
		{
			Name:     "Domain",
			Label:    "OpenStack Domain",
			Value:    &fd.Secrets.OpenStack.Domain,
			Required: true,
		},
		{
			Name:     "Project",
			Label:    "OpenStack Project",
			Value:    &fd.Secrets.OpenStack.Project,
			Required: true,
		},
		{
			Name:     "ProjectID",
			Label:    "OpenStack ProjectID",
			Value:    &fd.Secrets.OpenStack.ProjectID,
			Required: true,
		},
		{
			Name:     "Username",
			Label:    "OpenStack Username",
			Value:    &fd.Secrets.OpenStack.Username,
			Required: true,
		},
		{
			Name:     "Password",
			Label:    "OpenStack Password",
			Value:    &fd.Secrets.OpenStack.Password,
			Required: true,
		},
	}
}
