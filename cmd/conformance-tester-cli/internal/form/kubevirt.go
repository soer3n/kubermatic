package form

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
