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

package form

import (
	"fmt"
	"strconv"
	"strings"

	"k8c.io/kubermatic/sdk/v2/semver"
	"k8c.io/kubermatic/v2/cmd/conformance-tester-cli/internal/config"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	"k8c.io/machine-controller/sdk/providerconfig"
)

var providerDisplayMap = map[providerconfig.CloudProvider]string{
	providerconfig.CloudProviderAlibaba:             "Alibaba",
	providerconfig.CloudProviderAnexia:              "Anexia",
	providerconfig.CloudProviderAWS:                 "AWS",
	providerconfig.CloudProviderAzure:               "Azure",
	providerconfig.CloudProviderDigitalocean:        "DigitalOcean",
	providerconfig.CloudProviderGoogle:              "GCP",
	providerconfig.CloudProviderHetzner:             "Hetzner",
	providerconfig.CloudProviderKubeVirt:            "KubeVirt",
	providerconfig.CloudProviderNutanix:             "Nutanix",
	providerconfig.CloudProviderOpenstack:           "OpenStack",
	providerconfig.CloudProviderVMwareCloudDirector: "VMware Cloud Director",
	providerconfig.CloudProviderVsphere:             "vSphere",
}

// ProviderFormSettings holds settings for a single provider form.
type ProviderFormSettings struct {
	TestSettings []string
}

// FormData holds all the form state and temporary variables.
type FormData struct {
	Config            *config.Config
	Secrets           *types.Secrets
	ProvidersSelected []string
	Dists             []string
	Releases          []string
	Runtimes          []string
	Excludes          []string
	EnvOpt            string
	ParallelStr       string
	NutanixCSIPortStr string
	VMCDNetworksStr   string
	NodeCountStr      string
	RunTests          bool
	// Test settings selections for each provider, keyed by provider name.
	ProviderSettings map[string]*ProviderFormSettings
}

// NewFormData creates a new FormData with initialized values.
func NewFormData() *FormData {
	fd := &FormData{
		Config:           config.NewConfig(),
		Secrets:          &types.Secrets{},
		ParallelStr:      "2",
		ProviderSettings: make(map[string]*ProviderFormSettings),
	}

	// Pre-populate the map for all known providers to ensure we can safely get
	// a pointer to the TestSettings slice for the form builder.
	for provider := range providerDisplayMap {
		fd.ProviderSettings[string(provider)] = &ProviderFormSettings{}
	}

	return fd
}

// PostProcess handles post-form processing of form data.
func (fd *FormData) PostProcess() error {
	// Build final config
	fd.Config.Providers = fd.ProvidersSelected
	fd.Config.Distributions = fd.Dists
	fd.Config.Releases = fd.Releases
	fd.Config.Environment = fd.EnvOpt
	fd.Config.Runtimes = fd.Runtimes
	fd.Config.ExcludeTests = fd.Excludes

	// Build the provider-centric settings map.
	for provider, settings := range fd.ProviderSettings {
		fd.Config.ProviderSettings[provider] = config.ProviderSettings{
			TestSettings: settings.TestSettings,
		}
	}

	// Parse parallel string
	if ps := strings.TrimSpace(fd.ParallelStr); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil {
			fd.Config.Parallel = n
		}
	}

	// Parse node count string
	if nc := strings.TrimSpace(fd.NodeCountStr); nc != "" {
		if n, err := strconv.Atoi(nc); err == nil {
			fd.Config.NodeCount = n
		}
	}

	// Post-process composite inputs
	if config.Contains(fd.ProvidersSelected, string(providerconfig.CloudProviderNutanix)) && strings.TrimSpace(fd.NutanixCSIPortStr) != "" {
		if n, err := strconv.Atoi(fd.NutanixCSIPortStr); err == nil {
			fd.Secrets.Nutanix.CSIPort = int32(n)
		}
	}

	if config.Contains(fd.ProvidersSelected, string(providerconfig.CloudProviderVMwareCloudDirector)) && strings.TrimSpace(fd.VMCDNetworksStr) != "" {
		parts := strings.Split(fd.VMCDNetworksStr, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		fd.Secrets.VMwareCloudDirector.OVDCNetworks = out
	}

	return nil
}

// requiredIf returns a validator that requires input when a provider is selected.
func (fd *FormData) requiredIf(provider string) func(string) error {
	return func(s string) error {
		if config.Contains(fd.ProvidersSelected, provider) && strings.TrimSpace(s) == "" {
			return fmt.Errorf("required")
		}
		return nil
	}
}

// requiredIfEnv returns a validator that requires input when an environment is selected.
func (fd *FormData) requiredIfEnv(env string) func(string) error {
	return func(s string) error {
		if fd.EnvOpt == env && strings.TrimSpace(s) == "" {
			return fmt.Errorf("required")
		}
		return nil
	}
}

// requireAtLeastOne returns a validator for Select/MultiSelect values
// that enforces at least one selection. The kind is used in the message.
func requireAtLeastOne(kind string) func([]string) error {
	return func(vals []string) error {
		if len(vals) == 0 {
			if strings.TrimSpace(kind) == "" {
				return fmt.Errorf("select at least one option")
			}
			return fmt.Errorf("select at least one %s", kind)
		}
		return nil
	}
}

// validateInt ensures the provided string parses as an integer.
func validateInt(str string) error {
	if _, err := fmt.Sscanf(str, "%d", new(int)); err != nil {
		return fmt.Errorf("must be a number")
	}
	return nil
}

// buildReleaseOptions builds option list from version strings
func buildReleaseOptions(versions []semver.Semver) []string {
	opts := make([]string, 0, len(versions))
	for _, v := range versions {
		opts = append(opts, v.String())
	}
	return opts
}
