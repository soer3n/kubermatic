package form

import (
	"github.com/charmbracelet/huh"
	"k8c.io/kubermatic/v2/cmd/conformance-tester-cli/internal/config"
	ginkgoutils "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	"k8c.io/machine-controller/sdk/providerconfig"
)

// buildProviderFormGroups creates generic secret and scenario form groups for a provider.
func buildProviderFormGroups(
	providerName string,
	providerCloudName providerconfig.CloudProvider,
	fd *FormData,
	secretFields []huh.Field,
) []*huh.Group {
	// Hide secrets when the provider is not selected OR when running against
	// an existing KKP instance (we shouldn't ask for provider credentials).
	secretsHide := func() bool {
		return !config.Contains(fd.ProvidersSelected, string(providerCloudName)) || fd.EnvOpt == "KKP"
	}

	// Get test scenarios for this provider
	testScenarios := ginkgoutils.GetSettingsForProvider(providerCloudName)

	// Each provider needs its own scenario field function to bind to the correct FormData field.
	scenarioFields := buildScenarioFields(testScenarios, getTestSettingsValue(providerCloudName, fd))

	secretsGroup := huh.NewGroup(secretFields...).
		WithHideFunc(secretsHide).
		Title("Enter Credentials for " + providerName)

	// Scenarios should still be shown when the provider is selected even if
	// we're using an existing KKP instance — only secrets are skipped.
	scenarioGroup := huh.NewGroup(scenarioFields...).
		WithHideFunc(func() bool { return !config.Contains(fd.ProvidersSelected, string(providerCloudName)) }).
		Title("Select Test Scenarios for " + providerName)

	return []*huh.Group{secretsGroup, scenarioGroup}
}

// buildScenarioFields creates a generic scenario multi-select field.
func buildScenarioFields(
	testScenarios []ginkgoutils.TestSettings,
	value *[]string,
) []huh.Field {
	options := make([]huh.Option[string], 0, len(testScenarios))
	for _, s := range testScenarios {
		options = append(options, huh.NewOption(s.Description, s.Description))
	}

	return []huh.Field{
		huh.NewMultiSelect[string]().
			Options(options...).
			Value(value),
	}
}

// getTestSettingsValue returns a pointer to the correct TestSettings field in FormData.
func getTestSettingsValue(provider providerconfig.CloudProvider, fd *FormData) *[]string {
	// The map is pre-populated in NewFormData, so we can safely assume the key exists.
	return &fd.ProviderSettings[string(provider)].TestSettings
}
