package form

// getVSphereSecretFields returns the secret credential fields for vSphere provider
func (fd *FormData) getVSphereSecretFields() []SecretField {
	return []SecretField{
		{
			Name:     "KKPDatacenter",
			Label:    "vSphere KKPDatacenter",
			Value:    &fd.Secrets.VSphere.KKPDatacenter,
			Required: true,
		},
		{
			Name:     "Username",
			Label:    "vSphere Username",
			Value:    &fd.Secrets.VSphere.Username,
			Required: true,
		},
		{
			Name:     "Password",
			Label:    "vSphere Password",
			Value:    &fd.Secrets.VSphere.Password,
			Required: true,
		},
	}
}
