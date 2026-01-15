package form

// getAzureSecretFields returns the secret credential fields for Azure provider
func (fd *FormData) getAzureSecretFields() []SecretField {
	return []SecretField{
		{
			Name:     "KKPDatacenter",
			Label:    "Azure KKPDatacenter",
			Value:    &fd.Secrets.Azure.KKPDatacenter,
			Required: true,
		},
		{
			Name:     "ClientID",
			Label:    "Azure ClientID",
			Value:    &fd.Secrets.Azure.ClientID,
			Required: true,
		},
		{
			Name:     "ClientSecret",
			Label:    "Azure ClientSecret",
			Value:    &fd.Secrets.Azure.ClientSecret,
			Required: true,
		},
		{
			Name:     "TenantID",
			Label:    "Azure TenantID",
			Value:    &fd.Secrets.Azure.TenantID,
			Required: true,
		},
		{
			Name:     "SubscriptionID",
			Label:    "Azure SubscriptionID",
			Value:    &fd.Secrets.Azure.SubscriptionID,
			Required: true,
		},
	}
}
