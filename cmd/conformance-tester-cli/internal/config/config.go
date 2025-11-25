/*
Copyright 2025 The Kubermatic Kubernetes Platform contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"os"
	"reflect"
	"strings"

	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
)

// ProviderSettings holds the settings for a single provider.
type ProviderSettings struct {
	TestSettings []string `json:"testSettings,omitempty"`
}

// Config represents the internal configuration structure.
type Config struct {
	Providers               []string
	ProviderSettings        map[string]ProviderSettings
	Distributions           []string
	Releases                []string
	Environment             string
	Seed                    string
	Preset                  string
	Project                 string
	Runtimes                []string
	Parallel                int
	NamePrefix              string
	ExcludeTests            []string
	NodeCount               int
	DeleteClusterAfterTests bool
}

// OutputConfig is the final shape printed as JSON, structured by provider.
type OutputConfig struct {
	NamePrefix                 string                      `json:"namePrefix"`
	Providers                  map[string]ProviderSettings `json:"providers"`
	Secrets                    types.Secrets               `json:"secrets"`
	Releases                   []string                    `json:"releases"`
	EnableDistributions        []string                    `json:"enableDistributions"`
	DeleteClusterAfterTests    bool                        `json:"deleteClusterAfterTests"`
	NodeCount                  int                         `json:"nodeCount"`
	KubermaticSeedName         string                      `json:"kubermaticSeedName"`
	KubermaticProject          string                      `json:"kubermaticProject"`
	KubermaticParallelClusters int                         `json:"kubermaticParallelClusters"`
	ExcludeTests               []string                    `json:"excludeTests,omitempty"`
}

// GinkgoOutputConfig is the final shape printed as YAML for the ginkgo runner.
type GinkgoOutputConfig struct {
	NamePrefix          string        `yaml:"namePrefix,omitempty"`
	Providers           []string      `yaml:"providers,omitempty"`
	Releases            []string      `yaml:"releases,omitempty"`
	EnableDistributions []string      `yaml:"enableDistributions,omitempty"`
	TestSettings        []string      `yaml:"testSettings,omitempty"`
	DeleteCluster       bool          `yaml:"deleteClusterAfterTests"`
	NodeCount           int           `yaml:"nodeCount,omitempty"`
	ReportsRoot         string        `yaml:"reportsRoot,omitempty"`
	LogDirectory        string        `yaml:"logDirectory,omitempty"`
	KubermaticSeedName  string        `yaml:"kubermaticSeedName,omitempty"`
	Secrets             types.Secrets `yaml:"secrets"`
}

// NewConfig creates a new Config with sensible defaults.
func NewConfig() *Config {
	cfg := &Config{
		Parallel:         2,
		NodeCount:        1,
		ProviderSettings: make(map[string]ProviderSettings),
	}

	// Default Name Prefix to hostname
	if host, err := os.Hostname(); err == nil {
		cfg.NamePrefix = host
	}

	return cfg
}

// ToOutputConfig converts Config to OutputConfig with the provided secrets.
func (c *Config) ToOutputConfig(secrets types.Secrets) OutputConfig {
	// Filter out providers that were not selected by the user.
	selectedProviders := make(map[string]ProviderSettings)
	for _, provider := range c.Providers {
		if settings, ok := c.ProviderSettings[provider]; ok {
			selectedProviders[provider] = settings
		}
	}

	return OutputConfig{
		NamePrefix:                 c.NamePrefix,
		Providers:                  selectedProviders,
		Releases:                   c.Releases,
		EnableDistributions:        c.Distributions,
		Secrets:                    secrets,
		DeleteClusterAfterTests:    c.DeleteClusterAfterTests,
		NodeCount:                  c.NodeCount,
		KubermaticSeedName:         c.Seed,
		KubermaticProject:          c.Project,
		KubermaticParallelClusters: c.Parallel,
		ExcludeTests:               c.ExcludeTests,
	}
}

// ToGinkgoOutputConfig converts Config to GinkgoOutputConfig with the provided secrets.
func (c *Config) ToGinkgoOutputConfig(secrets types.Secrets) GinkgoOutputConfig {
	// Filter out secrets for providers that were not selected to keep the output clean.
	activeSecrets := types.Secrets{}
	v := reflect.ValueOf(&activeSecrets).Elem()
	s := reflect.ValueOf(secrets)

	for i := 0; i < v.NumField(); i++ {
		providerName := v.Type().Field(i).Name
		if Contains(c.Providers, strings.ToLower(providerName)) {
			v.Field(i).Set(s.FieldByName(providerName))
		}
	}

	// Combine test settings from all selected providers.
	var allTestSettings []string
	for _, provider := range c.Providers {
		if settings, ok := c.ProviderSettings[provider]; ok {
			allTestSettings = append(allTestSettings, settings.TestSettings...)
		}
	}

	return GinkgoOutputConfig{
		NamePrefix:          c.NamePrefix,
		Providers:           c.Providers,
		Releases:            c.Releases,
		EnableDistributions: c.Distributions,
		TestSettings:        allTestSettings,
		DeleteCluster:       c.DeleteClusterAfterTests,
		NodeCount:           c.NodeCount,
		ReportsRoot:         "_reports",
		LogDirectory:        "_logs",
		KubermaticSeedName:  c.Seed,
		Secrets:             activeSecrets,
	}
}

// Contains checks if a string exists in a slice (case-insensitive).
func Contains(list []string, s string) bool {
	for _, v := range list {
		if strings.EqualFold(v, s) {
			return true
		}
	}
	return false
}
