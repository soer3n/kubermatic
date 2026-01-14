package utils

import (
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
)

func GetReleaseVersions() []string {
	versions := []string{}
	kkpConfig, err := k8cginkgo.LoadKubermaticConfiguration()
	if err != nil {
		return versions
	}
	for _, scenario := range kkpConfig.Spec.Versions.Versions {
		versions = append(versions, scenario.String())
	}
	return versions
}

func GetClusterDescriptions() []string {
	descriptions := []string{}
	for _, modifier := range k8cginkgo.ClusterSettings {
		descriptions = append(descriptions, modifier.Name)
	}
	return descriptions
}
