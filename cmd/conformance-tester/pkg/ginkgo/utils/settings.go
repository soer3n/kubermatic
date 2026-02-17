package utils

import (
	"strings"

	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var DefaultMachineResources = k8cginkgo.ResourceSettings{
	Cpu:      []int{2},
	Memory:   []string{"4Gi"},
	DiskSize: []string{"20Gi"},
}

func GetClusterDescriptions(client ctrlclient.Client) map[string]k8cginkgo.Description {
	groupedSettings := map[string]k8cginkgo.Description{}
	groupedSettingsDesc := map[string][]string{}
	for _, modifier := range k8cginkgo.ClusterSettings {
		groupedSettingsDesc[modifier.Group] = append(groupedSettingsDesc[modifier.Group], modifier.Name)
	}

	for group, descs := range groupedSettingsDesc {
		strippedDescs := stripPrefix(descs)
		if len(strippedDescs) == 1 {
			strippedDescs = nil
		}
		groupedSettings[group] = k8cginkgo.Description{
			Name:    longestCommonPrefixTokens(descs, " "),
			Options: strippedDescs,
		}
	}
	return groupedSettings
}

func stripPrefix(strs []string) []string {
	prefix := longestCommonPrefixTokens(strs, " ")
	out := make([]string, 0, len(strs))

	for _, s := range strs {
		out = append(out, strings.TrimPrefix(s, prefix))
	}
	return out
}

func longestCommonPrefixTokens(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}

	base := strings.Split(strs[0], sep)
	maxTokens := len(base)

	for _, s := range strs[1:] {
		tokens := strings.Split(s, sep)

		i := 0
		for i < maxTokens && i < len(tokens) && tokens[i] == base[i] {
			i++
		}
		maxTokens = i

		if maxTokens == 0 {
			return ""
		}
	}

	prefix := strings.Join(base[:maxTokens], sep)

	// preserve trailing separator if it existed
	return prefix + sep
}
