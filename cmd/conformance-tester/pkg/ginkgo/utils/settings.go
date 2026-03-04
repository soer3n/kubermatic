package utils

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/build"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/settings"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	"k8c.io/machine-controller/sdk/providerconfig"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var DefaultMachineResources = ResourceSettings{
	Cpu:      []int{2},
	Memory:   []string{"4Gi"},
	DiskSize: []string{"20Gi"},
}

type Description struct {
	Name    string   `yaml:"name,omitempty"`
	Options []string `yaml:"options,omitempty"`
}

type ResourceSettings struct {
	Cpu      []int    `yaml:"cpu,omitempty"`
	Memory   []string `yaml:"memory,omitempty"`
	DiskSize []string `yaml:"diskSize,omitempty"`
}

func GetMachineDescriptions(provider string, secrets legacytypes.Secrets) map[string]Description {
	providerConfig, err := build.GetProviderConfig(context.Background(), &zap.SugaredLogger{}, secrets, providerconfig.OperatingSystemUbuntu, providerconfig.CloudProvider(provider))
	if err != nil {
		panic(err)
	}
	settings := build.MachineSettings(context.Background(), providerConfig, "", secrets, nil)
	groupedSettings := map[string]Description{}
	groupedSettingsDesc := map[string][]string{}
	for _, modifier := range settings {
		groupedSettingsDesc[modifier.Group] = append(groupedSettingsDesc[modifier.Group], modifier.Name)
	}

	for group, descs := range groupedSettingsDesc {
		strippedDescs := stripPrefix(descs)
		if len(strippedDescs) == 1 {
			strippedDescs = nil
		}
		groupedSettings[group] = Description{
			Name:    longestCommonPrefixTokens(descs, " "),
			Options: strippedDescs,
		}
	}
	return groupedSettings
}

func GetDatacenterDescriptions(provider string, secrets legacytypes.Secrets) map[string]Description {
	providerConfig, err := build.GetProviderConfig(context.Background(), &zap.SugaredLogger{}, secrets, providerconfig.OperatingSystemUbuntu, providerconfig.CloudProvider(provider))
	if err != nil {
		panic(err)
	}
	settings := build.GenericDatacenterSettings(context.Background(), providerConfig, secrets)
	groupedSettings := map[string]Description{}
	groupedSettingsDesc := map[string][]string{}
	for _, modifier := range settings {
		groupedSettingsDesc[modifier.Group] = append(groupedSettingsDesc[modifier.Group], modifier.Name)
	}

	for group, descs := range groupedSettingsDesc {
		strippedDescs := stripPrefix(descs)
		if len(strippedDescs) == 1 {
			strippedDescs = nil
		}
		groupedSettings[group] = Description{
			Name:    longestCommonPrefix(descs),
			Options: strippedDescs,
		}
	}
	return groupedSettings
}

func GetClusterDescriptions(client ctrlclient.Client) map[string]Description {
	groupedSettings := map[string]Description{}
	groupedSettingsDesc := map[string][]string{}
	for _, modifier := range settings.ClusterSettings {
		groupedSettingsDesc[modifier.Group] = append(groupedSettingsDesc[modifier.Group], modifier.Name)
	}

	for group, descs := range groupedSettingsDesc {
		strippedDescs := stripPrefix(descs)
		if len(strippedDescs) == 1 {
			strippedDescs = nil
		}
		groupedSettings[group] = Description{
			Name:    longestCommonPrefixTokens(descs, " "),
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
