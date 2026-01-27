package kubevirt

import (
	"context"
	"strings"

	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	"k8c.io/kubermatic/v2/pkg/test/e2e/utils"
)

func GetMachineDescriptions() map[string]k8cginkgo.Description {
	client, _, err := utils.GetClients()
	if err != nil {
		return nil
	}

	settings := MachineSettings(context.Background(), client, "kubermatic", nil)

	groupedSettings := map[string]k8cginkgo.Description{}
	groupedSettingsDesc := map[string][]string{}
	for _, modifier := range settings {
		groupedSettingsDesc[modifier.group] = append(groupedSettingsDesc[modifier.group], modifier.name)
	}

	for group, descs := range groupedSettingsDesc {
		strippedDescs := stripPrefix(descs)
		if len(strippedDescs) == len(descs) {
			strippedDescs = nil
		}
		groupedSettings[group] = k8cginkgo.Description{
			Name:    longestCommonPrefixTokens(descs, " "),
			Options: strippedDescs,
		}
	}
	return groupedSettings
}

func GetDatacenterDescriptions() map[string]k8cginkgo.Description {
	client, _, err := utils.GetClients()
	if err != nil {
		return nil
	}
	settings := GenericDatacenterSettings(context.Background(), client, "kubermatic")
	groupedSettings := map[string]k8cginkgo.Description{}
	groupedSettingsDesc := map[string][]string{}
	for _, modifier := range settings {
		groupedSettingsDesc[modifier.Group] = append(groupedSettingsDesc[modifier.Group], modifier.Name)
	}

	for group, descs := range groupedSettingsDesc {
		strippedDescs := stripPrefix(descs)
		if len(strippedDescs) == len(descs) {
			strippedDescs = nil
		}
		groupedSettings[group] = k8cginkgo.Description{
			Name:    longestCommonPrefix(descs),
			Options: strippedDescs,
		}
	}
	return groupedSettings
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	prefix := strs[0]

	for _, s := range strs[1:] {
		for len(prefix) > 0 && !strings.HasPrefix(s, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
		if prefix == "" {
			return ""
		}
	}
	return prefix
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

func stripPrefix(strs []string) []string {
	prefix := longestCommonPrefixTokens(strs, " ")
	out := make([]string, 0, len(strs))

	for _, s := range strs {
		out = append(out, strings.TrimPrefix(s, prefix))
	}
	return out
}
