package form

import (
	"github.com/charmbracelet/huh"
	"k8c.io/machine-controller/sdk/providerconfig"
)

func (fd *FormData) getKubevirtTestSettings() []*huh.Group {
	return buildProviderFormGroups(
		providerDisplayMap[providerconfig.CloudProviderKubeVirt],
		providerconfig.CloudProviderKubeVirt,
		fd,
		fd.getKubevirtSecretFields(),
	)
}

func (fd *FormData) getKubevirtSecretFields() []huh.Field {
	return []huh.Field{
		huh.NewInput().Title("KubeVirt KKPDatacenter").Value(&fd.Secrets.Kubevirt.KKPDatacenter).Validate(fd.requiredIf(string(providerconfig.CloudProviderKubeVirt))),
		huh.NewText().Title("KubeVirt Kubeconfig").Value(&fd.Secrets.Kubevirt.Kubeconfig).CharLimit(100000).Validate(fd.requiredIf(string(providerconfig.CloudProviderKubeVirt))).
			Placeholder("Paste as plaintext kubeconfig without any (base64) encoding"),
	}
}
