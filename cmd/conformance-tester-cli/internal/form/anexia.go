package form

// getAnexiaSecretFields returns the secret credential fields for Anexia provider
func (fd *FormData) getAnexiaSecretFields() []SecretField {
	return []SecretField{
		{
			Name:     "KKPDatacenter",
			Label:    "Anexia KKPDatacenter",
			Value:    &fd.Secrets.Anexia.KKPDatacenter,
			Required: true,
		},
		{
			Name:     "Token",
			Label:    "Anexia Token",
			Value:    &fd.Secrets.Anexia.Token,
			Required: true,
		},
		{
			Name:     "TemplateID",
			Label:    "Anexia TemplateID",
			Value:    &fd.Secrets.Anexia.TemplateID,
			Required: true,
		},
		{
			Name:     "VlanID",
			Label:    "Anexia VlanID",
			Value:    &fd.Secrets.Anexia.VlanID,
			Required: true,
		},
	}
}
