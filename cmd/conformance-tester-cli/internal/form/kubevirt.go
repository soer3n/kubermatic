package form

import (
	"k8c.io/machine-controller/sdk/providerconfig"
)

// getKubevirtTestSettings returns provider-specific test settings (placeholder for future use)
func (fd *FormData) getKubevirtTestSettings() []string {
	testSettings := GetTestSettingsForProvider(providerconfig.CloudProviderKubeVirt)
	var result []string
	for _, ts := range testSettings {
		result = append(result, ts.Description)
	}
	return result
}

// getKubevirtSecretFields returns the secret credential fields for KubeVirt provider
func (fd *FormData) getKubevirtSecretFields() []SecretField {
	return []SecretField{
		{
			Name:     "KKPDatacenter",
			Label:    "KubeVirt KKPDatacenter",
			Value:    &fd.Secrets.Kubevirt.KKPDatacenter,
			Required: true,
		},
		{
			Name:     "Kubeconfig",
			Label:    "KubeVirt Kubeconfig",
			Value:    &fd.Secrets.Kubevirt.Kubeconfig,
			Required: true,
		},
	}
}
