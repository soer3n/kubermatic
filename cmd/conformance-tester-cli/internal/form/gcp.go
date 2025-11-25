package form

import (
	"github.com/charmbracelet/huh"
	"k8c.io/machine-controller/sdk/providerconfig"
)

func (fd *FormData) getGCPTestSettings() []*huh.Group {
	return buildProviderFormGroups(
		providerDisplayMap[providerconfig.CloudProviderGoogle],
		providerconfig.CloudProviderGoogle,
		fd,
		fd.getGCPSecretFields(),
	)
}
func (fd *FormData) getGCPSecretFields() []huh.Field {
	return []huh.Field{
		huh.NewInput().Title("GCP KKPDatacenter").Value(&fd.Secrets.GCP.KKPDatacenter).Validate(fd.requiredIf(string(providerconfig.CloudProviderGoogle))),
		huh.NewText().Title("GCP ServiceAccount (JSON)").Value(&fd.Secrets.GCP.ServiceAccount).CharLimit(10000).Validate(fd.requiredIf(string(providerconfig.CloudProviderGoogle))).
			Placeholder("ServiceAccount is the plaintext Service account (as JSON) without any (base64) encoding"),
		huh.NewInput().Title("GCP Network").Value(&fd.Secrets.GCP.Network).Validate(fd.requiredIf(string(providerconfig.CloudProviderGoogle))),
		huh.NewInput().Title("GCP Subnetwork").Value(&fd.Secrets.GCP.Subnetwork).Validate(fd.requiredIf(string(providerconfig.CloudProviderGoogle))),
	}
}
