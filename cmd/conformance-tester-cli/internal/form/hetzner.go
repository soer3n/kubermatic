package form

// getHetznerSecretFields returns the secret credential fields for Hetzner provider
func (fd *FormData) getHetznerSecretFields() []SecretField {
	return []SecretField{
		{
			Name:     "KKPDatacenter",
			Label:    "Hetzner KKPDatacenter",
			Value:    &fd.Secrets.Hetzner.KKPDatacenter,
			Required: true,
		},
		{
			Name:     "Token",
			Label:    "Hetzner Token",
			Value:    &fd.Secrets.Hetzner.Token,
			Required: true,
		},
	}
}
