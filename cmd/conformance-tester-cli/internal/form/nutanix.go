package form

import (
	"k8c.io/machine-controller/sdk/providerconfig"
)

// getNutanixTestSettings returns provider-specific test settings (placeholder for future use)
func (fd *FormData) getNutanixTestSettings() []string {
	testSettings := GetTestSettingsForProvider(providerconfig.CloudProviderNutanix)
	var result []string
	for _, ts := range testSettings {
		result = append(result, ts.Description)
	}
	return result
}

// getNutanixSecretFields returns the secret credential fields for Nutanix provider
func (fd *FormData) getNutanixSecretFields() []SecretField {
	return []SecretField{
		{
			Name:     "KKPDatacenter",
			Label:    "Nutanix KKPDatacenter",
			Value:    &fd.Secrets.Nutanix.KKPDatacenter,
			Required: true,
		},
		{
			Name:     "Username",
			Label:    "Nutanix Username",
			Value:    &fd.Secrets.Nutanix.Username,
			Required: true,
		},
		{
			Name:     "Password",
			Label:    "Nutanix Password",
			Value:    &fd.Secrets.Nutanix.Password,
			Required: true,
		},
		{
			Name:     "CSIUsername",
			Label:    "Nutanix CSI Username",
			Value:    &fd.Secrets.Nutanix.CSIUsername,
			Required: true,
		},
		{
			Name:     "CSIPassword",
			Label:    "Nutanix CSI Password",
			Value:    &fd.Secrets.Nutanix.CSIPassword,
			Required: true,
		},
		{
			Name:     "CSIEndpoint",
			Label:    "Nutanix CSI Endpoint",
			Value:    &fd.Secrets.Nutanix.CSIEndpoint,
			Required: true,
		},
		{
			Name:     "CSIPort",
			Label:    "Nutanix CSI Port",
			Value:    &fd.NutanixCSIPortStr,
			Required: true,
		},
		{
			Name:     "ProxyURL",
			Label:    "Nutanix Proxy URL",
			Value:    &fd.Secrets.Nutanix.ProxyURL,
			Required: true,
		},
		{
			Name:     "ClusterName",
			Label:    "Nutanix Cluster Name",
			Value:    &fd.Secrets.Nutanix.ClusterName,
			Required: true,
		},
		{
			Name:     "ProjectName",
			Label:    "Nutanix Project Name",
			Value:    &fd.Secrets.Nutanix.ProjectName,
			Required: true,
		},
		{
			Name:     "SubnetName",
			Label:    "Nutanix Subnet Name",
			Value:    &fd.Secrets.Nutanix.SubnetName,
			Required: true,
		},
	}
}
