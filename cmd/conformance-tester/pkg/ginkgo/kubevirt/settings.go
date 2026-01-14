package kubevirt

import (
	"context"

	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	"k8c.io/kubermatic/v2/pkg/test/e2e/utils"
)

func GetMachineDescriptions() []string {
	descriptions := []string{}
	client, _, err := utils.GetClients()
	if err != nil {
		return nil
	}
	opts := k8cginkgo.NewDefaultOptions()
	for _, modifier := range MachineSettings(context.Background(), client, "kubermatic", &opts.Resources) {
		descriptions = append(descriptions, modifier.name)
	}
	return descriptions
}

func GetDatacenterDescriptions() []string {
	descriptions := []string{}
	client, _, err := utils.GetClients()
	if err != nil {
		return nil
	}
	for _, modifier := range GenericDatacenterSettings(context.Background(), client, "kubermatic") {
		descriptions = append(descriptions, modifier.Name)
	}
	return descriptions
}
