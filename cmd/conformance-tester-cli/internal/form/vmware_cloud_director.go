package form

import (
	"k8c.io/machine-controller/sdk/providerconfig"
)

// getVMwareCloudDirectorTestSettings returns provider-specific test settings (placeholder for future use)
func (fd *FormData) getVMwareCloudDirectorTestSettings() []string {
	testSettings := GetTestSettingsForProvider(providerconfig.CloudProviderVMwareCloudDirector)
	var result []string
	for _, ts := range testSettings {
		result = append(result, ts.Description)
	}
	return result
}

// getVDCSecretFields returns the secret credential fields for VMware Cloud Director provider
func (fd *FormData) getVDCSecretFields() []SecretField {
	return []SecretField{
		{
			Name:     "KKPDatacenter",
			Label:    "VMware Cloud Director KKPDatacenter",
			Value:    &fd.Secrets.VMwareCloudDirector.KKPDatacenter,
			Required: true,
		},
		{
			Name:     "Username",
			Label:    "VCD Username",
			Value:    &fd.Secrets.VMwareCloudDirector.Username,
			Required: true,
		},
		{
			Name:     "Password",
			Label:    "VCD Password",
			Value:    &fd.Secrets.VMwareCloudDirector.Password,
			Required: true,
		},
		{
			Name:     "Organization",
			Label:    "VCD Organization",
			Value:    &fd.Secrets.VMwareCloudDirector.Organization,
			Required: true,
		},
		{
			Name:     "VDC",
			Label:    "VCD VDC",
			Value:    &fd.Secrets.VMwareCloudDirector.VDC,
			Required: true,
		},
		{
			Name:     "OVDCNetworks",
			Label:    "VCD OVDC Networks (comma-separated)",
			Value:    &fd.VMCDNetworksStr,
			Required: true,
		},
	}
}
