package ginkgo

import (
	"fmt"

	"k8c.io/machine-controller/sdk/cloudprovider/kubevirt"
	"k8c.io/machine-controller/sdk/providerconfig"
	kubevirtcorev1 "kubevirt.io/api/core/v1"
)

// Table-driven scenario generator example for RawConfig fields
// You can move this to a more appropriate file if needed.
type FieldVariant struct {
	Name   string
	Values []interface{}
}

// Example for kubevirt.RawConfig, adapt as needed for your struct
func generateKubevirtTestSettings(defaults *kubevirt.RawConfig, variants []FieldVariant) []TestSettings {
	var settings []TestSettings
	for _, variant := range variants {
		for _, value := range variant.Values {
			cfg := *defaults // shallow copy
			desc := variant.Name + ": " + fmt.Sprintf("%v", value)
			switch variant.Name {
			// Top-level fields
			case "Instancetype":
				cfg.VirtualMachine.Instancetype = &kubevirtcorev1.InstancetypeMatcher{Name: value.(string)}
			case "Preference":
				cfg.VirtualMachine.Preference = &kubevirtcorev1.PreferenceMatcher{Name: value.(string)}
			case "DNSPolicy":
				cfg.VirtualMachine.DNSPolicy = providerconfig.ConfigVarString{Value: value.(string)}
			case "EvictionStrategy":
				cfg.VirtualMachine.EvictionStrategy = value.(string)

			// Nested: Template
			case "CPUs":
				cfg.VirtualMachine.Template.CPUs = providerconfig.ConfigVarString{Value: value.(string)}
			case "Memory":
				cfg.VirtualMachine.Template.Memory = providerconfig.ConfigVarString{Value: value.(string)}
			case "SecondaryDisks.Size":
				cfg.VirtualMachine.Template.SecondaryDisks = []kubevirt.SecondaryDisks{{
					Disk: kubevirt.Disk{Size: providerconfig.ConfigVarString{Value: value.(string)}},
				}}
			case "SecondaryDisks.StorageClassName":
				cfg.VirtualMachine.Template.SecondaryDisks = []kubevirt.SecondaryDisks{{
					Disk: kubevirt.Disk{StorageClassName: providerconfig.ConfigVarString{Value: value.(string)}},
				}}

			// Nested: Affinity.NodeAffinityPreset
			case "NodeAffinityPreset.Type":
				cfg.Affinity.NodeAffinityPreset.Type = providerconfig.ConfigVarString{Value: value.(string)}
			case "NodeAffinityPreset.Key":
				cfg.Affinity.NodeAffinityPreset.Key = providerconfig.ConfigVarString{Value: value.(string)}
			case "NodeAffinityPreset.Values":
				cfg.Affinity.NodeAffinityPreset.Values = []providerconfig.ConfigVarString{{Value: value.(string)}}

			// Nested: TopologySpreadConstraints
			case "TopologySpreadConstraints.TopologyKey":
				cfg.TopologySpreadConstraints = []kubevirt.TopologySpreadConstraint{{
					TopologyKey: providerconfig.ConfigVarString{Value: value.(string)},
				}}
			case "TopologySpreadConstraints.MaxSkew":
				cfg.TopologySpreadConstraints = []kubevirt.TopologySpreadConstraint{{
					MaxSkew: providerconfig.ConfigVarString{Value: value.(string)},
				}}
			case "TopologySpreadConstraints.WhenUnsatisfiable":
				cfg.TopologySpreadConstraints = []kubevirt.TopologySpreadConstraint{{
					WhenUnsatisfiable: providerconfig.ConfigVarString{Value: value.(string)},
				}}
				// Add more fields as needed
			}
			settings = append(settings, TestSettings{
				Description:  desc,
				ProviderSpec: &cfg,
			})
		}
	}
	return settings
}

func getKubevirtTestSettings() []TestSettings {
	var settings []TestSettings

	// Default config for other fields
	// defaultConfig := &kubevirt.RawConfig{}

	// String fields: valid, invalid, empty
	stringVariants := []providerconfig.ConfigVarString{
		{Value: "valid"},
		{Value: ""},
		{Value: "invalid"},
	}

	// Boolean fields: true, false
	// boolVariants := []bool{true, false}

	// Instancetype
	for _, v := range stringVariants {
		settings = append(settings, TestSettings{
			Description: "Instancetype: " + v.Value,
			ProviderSpec: &kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Instancetype: &kubevirtcorev1.InstancetypeMatcher{Name: v.Value},
				},
			},
		})
	}
	// Preference
	for _, v := range stringVariants {
		settings = append(settings, TestSettings{
			Description: "Preference: " + v.Value,
			ProviderSpec: &kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Preference: &kubevirtcorev1.PreferenceMatcher{Name: v.Value},
				},
			},
		})
	}
	// CPUs
	for _, v := range stringVariants {
		settings = append(settings, TestSettings{
			Description: "CPUs: " + v.Value,
			ProviderSpec: &kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Template: kubevirt.Template{
						CPUs: v,
					},
				},
			},
		})
	}
	// Memory
	for _, v := range stringVariants {
		settings = append(settings, TestSettings{
			Description: "Memory: " + v.Value,
			ProviderSpec: &kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Template: kubevirt.Template{
						Memory: v,
					},
				},
			},
		})
	}
	// SecondaryDisks.Size
	for _, v := range stringVariants {
		settings = append(settings, TestSettings{
			Description: "SecondaryDisk Size: " + v.Value,
			ProviderSpec: &kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Template: kubevirt.Template{
						SecondaryDisks: []kubevirt.SecondaryDisks{{
							Disk: kubevirt.Disk{Size: v},
						}},
					},
				},
			},
		})
	}
	// SecondaryDisks.StorageClassName
	for _, v := range stringVariants {
		settings = append(settings, TestSettings{
			Description: "SecondaryDisk StorageClassName: " + v.Value,
			ProviderSpec: &kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					Template: kubevirt.Template{
						SecondaryDisks: []kubevirt.SecondaryDisks{{
							Disk: kubevirt.Disk{StorageClassName: v},
						}},
					},
				},
			},
		})
	}
	// DNSPolicy
	for _, v := range stringVariants {
		settings = append(settings, TestSettings{
			Description: "DNSPolicy: " + v.Value,
			ProviderSpec: &kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					DNSPolicy: v,
				},
			},
		})
	}
	// EvictionStrategy (string)
	for _, v := range stringVariants {
		settings = append(settings, TestSettings{
			Description: "EvictionStrategy: " + v.Value,
			ProviderSpec: &kubevirt.RawConfig{
				VirtualMachine: kubevirt.VirtualMachine{
					EvictionStrategy: v.Value,
				},
			},
		})
	}
	// NodeAffinityPreset.Type
	for _, v := range stringVariants {
		settings = append(settings, TestSettings{
			Description: "NodeAffinityPreset.Type: " + v.Value,
			ProviderSpec: &kubevirt.RawConfig{
				Affinity: kubevirt.Affinity{
					NodeAffinityPreset: kubevirt.NodeAffinityPreset{
						Type: v,
					},
				},
			},
		})
	}
	// NodeAffinityPreset.Key
	for _, v := range stringVariants {
		settings = append(settings, TestSettings{
			Description: "NodeAffinityPreset.Key: " + v.Value,
			ProviderSpec: &kubevirt.RawConfig{
				Affinity: kubevirt.Affinity{
					NodeAffinityPreset: kubevirt.NodeAffinityPreset{
						Key: v,
					},
				},
			},
		})
	}
	// NodeAffinityPreset.Values
	for _, v := range stringVariants {
		settings = append(settings, TestSettings{
			Description: "NodeAffinityPreset.Values: " + v.Value,
			ProviderSpec: &kubevirt.RawConfig{
				Affinity: kubevirt.Affinity{
					NodeAffinityPreset: kubevirt.NodeAffinityPreset{
						Values: []providerconfig.ConfigVarString{v},
					},
				},
			},
		})
	}
	// TopologySpreadConstraints.TopologyKey
	for _, v := range stringVariants {
		settings = append(settings, TestSettings{
			Description: "TopologySpreadConstraints.TopologyKey: " + v.Value,
			ProviderSpec: &kubevirt.RawConfig{
				TopologySpreadConstraints: []kubevirt.TopologySpreadConstraint{{
					TopologyKey: v,
				}},
			},
		})
	}
	// TopologySpreadConstraints.MaxSkew
	for _, v := range stringVariants {
		settings = append(settings, TestSettings{
			Description: "TopologySpreadConstraints.MaxSkew: " + v.Value,
			ProviderSpec: &kubevirt.RawConfig{
				TopologySpreadConstraints: []kubevirt.TopologySpreadConstraint{{
					MaxSkew: v,
				}},
			},
		})
	}
	// TopologySpreadConstraints.WhenUnsatisfiable
	for _, v := range stringVariants {
		settings = append(settings, TestSettings{
			Description: "TopologySpreadConstraints.WhenUnsatisfiable: " + v.Value,
			ProviderSpec: &kubevirt.RawConfig{
				TopologySpreadConstraints: []kubevirt.TopologySpreadConstraint{{
					WhenUnsatisfiable: v,
				}},
			},
		})
	}
	// Add boolean fields if any (example: some hypothetical field)
	// for _, b := range boolVariants {
	// 	settings = append(settings, TestSettings{
	// 		Description:  fmt.Sprintf("SomeBool: %v", b),
	// 		ProviderSpec: &kubevirt.RawConfig{
	// 			SomeBool: b,
	// 		},
	// 	})
	// }
	// Add a few mixed cases to test multiple non-defaults together
	settings = append(settings, TestSettings{
		Description: "Mixed: valid instancetype, invalid CPUs, valid DNSPolicy",
		ProviderSpec: &kubevirt.RawConfig{
			VirtualMachine: kubevirt.VirtualMachine{
				Instancetype: &kubevirtcorev1.InstancetypeMatcher{Name: "valid"},
				Template: kubevirt.Template{
					CPUs: providerconfig.ConfigVarString{Value: "invalid"},
				},
				DNSPolicy: providerconfig.ConfigVarString{Value: "valid"},
			},
		},
	})
	settings = append(settings, TestSettings{
		Description: "Mixed: empty preference, valid memory, invalid eviction strategy",
		ProviderSpec: &kubevirt.RawConfig{
			VirtualMachine: kubevirt.VirtualMachine{
				Preference: &kubevirtcorev1.PreferenceMatcher{Name: ""},
				Template: kubevirt.Template{
					Memory: providerconfig.ConfigVarString{Value: "valid"},
				},
				EvictionStrategy: "invalid",
			},
		},
	})
	settings = append(settings, TestSettings{
		Description: "Mixed: invalid secondary disk size, valid affinity key, empty topology key",
		ProviderSpec: &kubevirt.RawConfig{
			VirtualMachine: kubevirt.VirtualMachine{
				Template: kubevirt.Template{
					SecondaryDisks: []kubevirt.SecondaryDisks{{
						Disk: kubevirt.Disk{Size: providerconfig.ConfigVarString{Value: "invalid"}},
					}},
				},
			},
			Affinity: kubevirt.Affinity{
				NodeAffinityPreset: kubevirt.NodeAffinityPreset{
					Key: providerconfig.ConfigVarString{Value: "valid"},
				},
			},
			TopologySpreadConstraints: []kubevirt.TopologySpreadConstraint{{
				TopologyKey: providerconfig.ConfigVarString{Value: ""},
			}},
		},
	})
	// All fields set (including nested fields)
	settings = append(settings, TestSettings{
		Description: "All fields set (valid/invalid/empty)",
		ProviderSpec: &kubevirt.RawConfig{
			VirtualMachine: kubevirt.VirtualMachine{
				Instancetype: &kubevirtcorev1.InstancetypeMatcher{Name: "valid"},
				Preference:   &kubevirtcorev1.PreferenceMatcher{Name: "invalid"},
				Template: kubevirt.Template{
					CPUs:   providerconfig.ConfigVarString{Value: "2"},
					Memory: providerconfig.ConfigVarString{Value: "4Gi"},
					SecondaryDisks: []kubevirt.SecondaryDisks{{
						Disk: kubevirt.Disk{
							Size:             providerconfig.ConfigVarString{Value: "10Gi"},
							StorageClassName: providerconfig.ConfigVarString{Value: "local-path"},
						},
					}},
				},
				DNSPolicy:        providerconfig.ConfigVarString{Value: "ClusterFirstWithHostNet"},
				EvictionStrategy: "LiveMigrate",
			},
			Affinity: kubevirt.Affinity{
				NodeAffinityPreset: kubevirt.NodeAffinityPreset{
					Type:   providerconfig.ConfigVarString{Value: "required"},
					Key:    providerconfig.ConfigVarString{Value: "kubernetes.io/hostname"},
					Values: []providerconfig.ConfigVarString{{Value: "node-01"}},
				},
			},
			TopologySpreadConstraints: []kubevirt.TopologySpreadConstraint{{
				MaxSkew:           providerconfig.ConfigVarString{Value: "1"},
				TopologyKey:       providerconfig.ConfigVarString{Value: "kubernetes.io/hostname"},
				WhenUnsatisfiable: providerconfig.ConfigVarString{Value: "DoNotSchedule"},
			}},
		},
	})

	return settings
}

// Helper: Cartesian product generator for field values
func cartesianProduct(variants []FieldVariant) [][]interface{} {
	if len(variants) == 0 {
		return [][]interface{}{{}}
	}
	result := [][]interface{}{}
	rest := cartesianProduct(variants[1:])
	for _, v := range variants[0].Values {
		for _, r := range rest {
			row := append([]interface{}{v}, r...)
			result = append(result, row)
		}
	}
	return result
}

// Apply a combination to a config
func applyCombinationToConfig(defaults *kubevirt.RawConfig, fields []string, values []interface{}) *kubevirt.RawConfig {
	cfg := *defaults
	for i, field := range fields {
		switch field {
		case "Instancetype":
			cfg.VirtualMachine.Instancetype = &kubevirtcorev1.InstancetypeMatcher{Name: values[i].(string)}
		case "Preference":
			cfg.VirtualMachine.Preference = &kubevirtcorev1.PreferenceMatcher{Name: values[i].(string)}
		case "DNSPolicy":
			cfg.VirtualMachine.DNSPolicy = providerconfig.ConfigVarString{Value: values[i].(string)}
		case "EvictionStrategy":
			cfg.VirtualMachine.EvictionStrategy = values[i].(string)
		case "CPUs":
			cfg.VirtualMachine.Template.CPUs = providerconfig.ConfigVarString{Value: values[i].(string)}
		case "Memory":
			cfg.VirtualMachine.Template.Memory = providerconfig.ConfigVarString{Value: values[i].(string)}
		case "SecondaryDisks.Size":
			cfg.VirtualMachine.Template.SecondaryDisks = []kubevirt.SecondaryDisks{{
				Disk: kubevirt.Disk{Size: providerconfig.ConfigVarString{Value: values[i].(string)}},
			}}
		case "SecondaryDisks.StorageClassName":
			cfg.VirtualMachine.Template.SecondaryDisks = []kubevirt.SecondaryDisks{{
				Disk: kubevirt.Disk{StorageClassName: providerconfig.ConfigVarString{Value: values[i].(string)}},
			}}
		case "SecondaryDisks.Disk.Name":
			// Field 'Name' does not exist in kubevirt.Disk, skip or handle as needed
			continue
		case "SecondaryDisks.Disk.Bus":
			// Field 'Bus' does not exist in kubevirt.Disk, skip or handle as needed
			continue
		case "NodeAffinityPreset.Type":
			cfg.Affinity.NodeAffinityPreset.Type = providerconfig.ConfigVarString{Value: values[i].(string)}
		case "NodeAffinityPreset.Key":
			cfg.Affinity.NodeAffinityPreset.Key = providerconfig.ConfigVarString{Value: values[i].(string)}
		case "NodeAffinityPreset.Values":
			cfg.Affinity.NodeAffinityPreset.Values = []providerconfig.ConfigVarString{{Value: values[i].(string)}}
		case "TopologySpreadConstraints.TopologyKey":
			cfg.TopologySpreadConstraints = []kubevirt.TopologySpreadConstraint{{
				TopologyKey: providerconfig.ConfigVarString{Value: values[i].(string)},
			}}
		case "TopologySpreadConstraints.MaxSkew":
			cfg.TopologySpreadConstraints = []kubevirt.TopologySpreadConstraint{{
				MaxSkew: providerconfig.ConfigVarString{Value: values[i].(string)},
			}}
		case "TopologySpreadConstraints.WhenUnsatisfiable":
			cfg.TopologySpreadConstraints = []kubevirt.TopologySpreadConstraint{{
				WhenUnsatisfiable: providerconfig.ConfigVarString{Value: values[i].(string)},
			}}
		}
	}
	return &cfg
}

// Generate all cartesian product test settings
func generateKubevirtCartesianTestSettings(defaults *kubevirt.RawConfig, variants []FieldVariant) []TestSettings {
	var settings []TestSettings
	fields := make([]string, len(variants))
	for i, v := range variants {
		fields[i] = v.Name
	}
	combos := cartesianProduct(variants)
	for _, combo := range combos {
		cfg := applyCombinationToConfig(defaults, fields, combo)
		desc := "Cartesian: "
		for i, f := range fields {
			desc += f + "=" + fmt.Sprintf("%v", combo[i]) + ", "
		}
		settings = append(settings, TestSettings{
			Description:  desc,
			ProviderSpec: cfg,
		})
	}
	return settings
}
