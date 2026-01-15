package form

import (
	"k8c.io/machine-controller/sdk/providerconfig"
)

// ProviderSecrets represents credential fields for a provider
type ProviderSecrets struct {
	Label  string
	Fields []SecretField
}

// SecretField represents a single secret credential field
type SecretField struct {
	Name     string
	Label    string
	Value    *string
	Required bool
}

// getTestSettingsValue returns a pointer to the correct TestSettings field in FormData.
func getTestSettingsValue(provider providerconfig.CloudProvider, fd *FormData) *[]string {
	// The map is pre-populated in NewFormData, so we can safely assume the key exists.
	return &fd.ProviderSettings[string(provider)].TestSettings
}
