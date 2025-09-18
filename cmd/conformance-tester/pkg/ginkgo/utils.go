package ginkgo

import (
	"reflect"
	"slices"
	"strings"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/machine-controller/sdk/providerconfig"
	"k8s.io/apimachinery/pkg/util/sets"
)

// GetSettingsForProvider returns a list of test settings (user stories) for a given provider.
func GetSettingsForProvider(provider providerconfig.CloudProvider) []TestSettings {
	// User stories for the KubeVirt provider
	switch provider {
	case providerconfig.CloudProviderKubeVirt:
		return getKubevirtTestSettings()
	}

	// Default user story for other providers
	return []TestSettings{
		{Description: "with default settings"},
	}
}

func getAllProviders() map[string]providerconfig.CloudProvider {
	providers := make(map[string]providerconfig.CloudProvider)
	specType := reflect.TypeOf(kubermaticv1.DatacenterSpec{})

	for i := 0; i < specType.NumField(); i++ {
		field := specType.Field(i)

		// We are only interested in fields that represent a provider. We identify
		// them by checking if they are a pointer to a struct.
		if field.Type.Kind() != reflect.Ptr || field.Type.Elem().Kind() != reflect.Struct {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		providerName := strings.Split(jsonTag, ",")[0]
		// Some fields like `Fake` are not actual providers we want to test, so we skip them.
		if providerName != "" && providerName != "fake" {
			// The name in the struct is usually capitalized (e.g., "KubeVirt"),
			// but the providerconfig constants are lowercase.
			// We can convert the field name to the provider name.
			provider := providerconfig.CloudProvider(providerName)
			if slices.Contains(providerconfig.AllCloudProviders, provider) {
				providers[field.Name] = provider
			}
		}
	}
	return providers
}

// areCompatible checks if two TestSettings can be merged without conflicts.
func areCompatible(s1, s2 TestSettings) bool {
	if s1.ProviderSpec == nil || s2.ProviderSpec == nil {
		return true // Both are compatible if one has no provider-specific settings
	}

	// Ensure both specs are of the same type
	if reflect.TypeOf(s1.ProviderSpec) != reflect.TypeOf(s2.ProviderSpec) {
		return false
	}

	v1 := reflect.ValueOf(s1.ProviderSpec).Elem()
	v2 := reflect.ValueOf(s2.ProviderSpec).Elem()

	for i := 0; i < v1.NumField(); i++ {
		field1 := v1.Field(i)
		field2 := v2.Field(i)

		// A conflict exists if both fields are set (not their zero value)
		isField1Set := !reflect.DeepEqual(field1.Interface(), reflect.Zero(field1.Type()).Interface())
		isField2Set := !reflect.DeepEqual(field2.Interface(), reflect.Zero(field2.Type()).Interface())

		if isField1Set && isField2Set {
			return false
		}
	}

	return true
}

// mergeTestSettings takes a list of TestSettings and merges compatible ones.
// The returned list will be of equal or smaller size.
func mergeTestSettings(settings []TestSettings, enabledSettings sets.Set[string]) ([]TestSettings, error) {
	if len(settings) == 0 {
		return nil, nil
	}

	var mergedSettings []TestSettings
	mergedIndices := make(map[int]bool)

	for i := range settings {
		if mergedIndices[i] {
			continue
		}

		current := settings[i]
		mergedIndices[i] = true

		for j := i + 1; j < len(settings); j++ {
			if mergedIndices[j] {
				continue
			}

			if areCompatible(current, settings[j]) && enabledSettings.Has(settings[j].Description) {
				// Merge s2 into s1
				var descriptions []string
				if current.Description != "" {
					descriptions = append(descriptions, current.Description)
				}
				if settings[j].Description != "" {
					descriptions = append(descriptions, settings[j].Description)
				}
				current.Description = strings.Join(descriptions, " and ")

				if settings[j].ProviderSpec != nil {
					if current.ProviderSpec == nil {
						// Create a new instance of the same type as settings[j].ProviderSpec
						specType := reflect.TypeOf(settings[j].ProviderSpec).Elem()
						current.ProviderSpec = reflect.New(specType).Interface()
					}

					vj := reflect.ValueOf(settings[j].ProviderSpec).Elem()
					vc := reflect.ValueOf(current.ProviderSpec).Elem()
					for k := 0; k < vj.NumField(); k++ {
						fieldJ := vj.Field(k)
						if !reflect.DeepEqual(fieldJ.Interface(), reflect.Zero(fieldJ.Type()).Interface()) {
							vc.Field(k).Set(fieldJ)
						}
					}
				}
				mergedIndices[j] = true
			}
		}
		mergedSettings = append(mergedSettings, current)
	}

	return mergedSettings, nil
}
