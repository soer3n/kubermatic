package form

import (
	"k8c.io/machine-controller/sdk/providerconfig"
)

// getAlibabaTestSettings returns provider-specific test settings (placeholder for future use)
func (fd *FormData) getAlibabaTestSettings() []string {
	testSettings := GetTestSettingsForProvider(providerconfig.CloudProviderAlibaba)
	var result []string
	for _, ts := range testSettings {
		result = append(result, ts.Description)
	}
	return result
}

// getAlibabaSecretFields returns the secret credential fields for Alibaba provider
func (fd *FormData) getAlibabaSecretFields() []SecretField {
	return []SecretField{
		{
			Name:     "KKPDatacenter",
			Label:    "Alibaba KKPDatacenter",
			Value:    &fd.Secrets.Alibaba.KKPDatacenter,
			Required: true,
		},
		{
			Name:     "AccessKeyID",
			Label:    "Alibaba AccessKeyID",
			Value:    &fd.Secrets.Alibaba.AccessKeyID,
			Required: true,
		},
		{
			Name:     "AccessKeySecret",
			Label:    "Alibaba AccessKeySecret",
			Value:    &fd.Secrets.Alibaba.AccessKeySecret,
			Required: true,
		},
	}
}
