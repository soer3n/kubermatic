package form

// getDigitalOceanSecretFields returns the secret credential fields for DigitalOcean provider
func (fd *FormData) getDigitalOceanSecretFields() []SecretField {
	return []SecretField{
		{
			Name:     "KKPDatacenter",
			Label:    "DigitalOcean KKPDatacenter",
			Value:    &fd.Secrets.Digitalocean.KKPDatacenter,
			Required: true,
		},
		{
			Name:     "Token",
			Label:    "DigitalOcean Token",
			Value:    &fd.Secrets.Digitalocean.Token,
			Required: true,
		},
	}
}
