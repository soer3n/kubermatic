/*
                  Kubermatic Enterprise Read-Only License
                         Version 1.0 ("KERO-1.0")
                     Copyright © 2026 Kubermatic GmbH

   1.	You may only view, read and display for studying purposes the source
      code of the software licensed under this license, and, to the extent
      explicitly provided under this license, the binary code.
   2.	Any use of the software which exceeds the foregoing right, including,
      without limitation, its execution, compilation, copying, modification
      and distribution, is expressly prohibited.
   3.	THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND,
      EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
      MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
      IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
      CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
      TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
      SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

   END OF TERMS AND CONDITIONS
*/

package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	"k8c.io/machine-controller/sdk/providerconfig"
)

// Message types for execution.
type (
	logMsg        struct{ line string }
	errMsg        struct{ err error }
	doneMsg       struct{ success bool }
	execOutputMsg struct {
		output  string
		success bool
		err     error
	}
	startMsg              struct{ ch <-chan tea.Msg }
	seedsPresetsLoadedMsg struct {
		seeds   []string
		presets []string
		err     error
	}
	presetDetailsLoadedMsg struct {
		presetName string
		spec       map[string]interface{}
		err        error
	}
	datacenterSettingsLoadedMsg struct {
		provider     string
		descriptions map[string]k8cginkgo.Description
		err          error
	}
	machineSettingsLoadedMsg struct {
		provider     string
		descriptions map[string]k8cginkgo.Description
		err          error
	}
	cleanupProgressMsg struct {
		message string
		done    bool
		err     error
	}
)

// CredentialSource indicates whether credentials come from a preset or are custom.
type CredentialSource string

const (
	CredentialSourcePreset CredentialSource = "preset"
	CredentialSourceCustom CredentialSource = "custom"
)

// App stages.
const (
	stageWelcome = iota
	stageEnvironmentSelection
	stageReleaseSelection
	stageProviderSelection
	stageDistributionSelection
	stageDatacenterSettingsSelection
	stageClusterSettingsSelection
	stageMachineDeploymentSettingsSelection
	stageClusterConfiguration
	stageReviewSettings
	stageExecuting
	stageDone
)

type EnvironmentLocal struct {
	Selected                     bool
	CurrentField                 int
	KubermaticConfigurationsPath textinput.Model
	HelmValuesPath               textinput.Model
	MLAValuesPath                textinput.Model
	Errors                       EnvironmentLocalErrors
}

type EnvironmentLocalErrors struct {
	KubermaticConfigurationsPath string
	HelmValuesPath               string
	MLAValuesPath                string
}

// KubeconfigOption represents a kubeconfig source option.
type KubeconfigOption struct {
	Type        string // "env", "file", "custom"
	DisplayName string
	Path        string
	Selected    bool
}

type EnvironmentExisting struct {
	Selected                   bool
	CurrentField               int
	KubeconfigOptions          []KubeconfigOption
	KubeconfigFocusedIndex     int
	KubeconfigExpandedSections map[string]bool // Keys: "env", "file", "custom"
	CustomKubeconfigPath       textinput.Model
	// Seeds and Presets fetched from cluster
	AvailableSeeds      []string
	SeedFocusedIndex    int
	SelectedSeedIndex   int
	AvailablePresets    []string
	PresetFocusedIndex  int
	SelectedPresetIndex int
	LoadingSeeds        bool
	LoadingPresets      bool
	FetchError          string
	ProjectName         textinput.Model
	Errors              EnvironmentExistingErrors
}

type EnvironmentExistingErrors struct {
	KubeconfigPath string
	SeedName       string // Now validates selection rather than text input
	PresetName     string // Now validates selection rather than text input
	ProjectName    string
	Fields         map[string]string // Additional validation errors
}

// Provider holds configuration for a single cloud provider.
type Provider struct {
	CloudProvider        providerconfig.CloudProvider
	DisplayName          string
	Selected             bool
	CurrentField         int
	CredentialSource     CredentialSource // "preset" or "custom"
	PresetCredentials    interface{}      // Credentials from preset (read-only)
	HasPresetCredentials bool             // Whether preset has credentials for this provider
	Credentials          interface{}      // Custom credentials or working copy
	Errors               ProviderErrors
}

type ProviderErrors struct {
	Fields map[string]string // Dynamic field errors
}

// AWS credentials
type AWSCredentials struct {
	AccessKeyID          textinput.Model
	SecretAccessKey      textinput.Model
	AssumeRoleARN        textinput.Model
	AssumeRoleExternalID textinput.Model
}

// Azure credentials
type AzureCredentials struct {
	TenantID       textinput.Model
	SubscriptionID textinput.Model
	ClientID       textinput.Model
	ClientSecret   textinput.Model
}

// GCP credentials
type GCPCredentials struct {
	ServiceAccount textinput.Model // Path to service account JSON
}

// Alibaba credentials
type AlibabaCredentials struct {
	AccessKeyID     textinput.Model
	AccessKeySecret textinput.Model
}

// Anexia credentials
type AnexiaCredentials struct {
	Token textinput.Model
}

// DigitalOcean credentials
type DigitalOceanCredentials struct {
	Token textinput.Model
}

// Hetzner credentials
type HetznerCredentials struct {
	Token textinput.Model
}

// KubeVirt credentials
type KubeVirtCredentials struct {
	Kubeconfig textinput.Model
}

// Nutanix credentials
type NutanixCredentials struct {
	Username    textinput.Model
	Password    textinput.Model
	ClusterName textinput.Model
	ProxyURL    textinput.Model
	CSIUsername textinput.Model
	CSIPassword textinput.Model
	CSIEndpoint textinput.Model
}

// OpenStack credentials
type OpenStackCredentials struct {
	Username                    textinput.Model
	Password                    textinput.Model
	Project                     textinput.Model
	ProjectID                   textinput.Model
	Domain                      textinput.Model
	ApplicationCredentialID     textinput.Model
	ApplicationCredentialSecret textinput.Model
	Token                       textinput.Model
}

// vSphere credentials
type VSphereCredentials struct {
	Username textinput.Model
	Password textinput.Model
}

// VMware Cloud Director credentials
type VMwareCloudDirectorCredentials struct {
	Username     textinput.Model
	Password     textinput.Model
	APIToken     textinput.Model
	Organization textinput.Model
	VDC          textinput.Model
}

// ReleaseSelection holds the state for Kubernetes release selection.
type ReleaseSelection struct {
	MajorVersions     []string            // e.g., ["1.31", "1.32"]
	MinorVersions     map[string][]string // e.g., {"1.31": ["1.31.1", "1.31.2"]}
	SelectedMajor     map[string]bool     // Tracks which major versions are selected (selects all minors)
	SelectedMinor     map[string]bool     // Tracks which minor versions are selected
	FocusedMajorIndex int                 // Currently focused major version
	FocusedMinorIndex int                 // Currently focused minor version within major
	IsMinorFocused    bool                // Whether focus is on a minor version or major version
}

// SettingsViewport tracks scroll state for paginated settings stages.
type SettingsViewport struct {
	ScrollOffset int // First visible row index
	PageSize     int // Number of visible rows (updated dynamically from UI height)
}

// SettingGroup represents a group of related settings with a parent name and child options.
type SettingGroup struct {
	Key        string   // Unique key from map (e.g., "cpuTopology")
	Name       string   // Display name (e.g., "CPU Topology")
	Options    []string // Child options (e.g., ["threads", "cores", "sockets"])
	IsExpanded bool     // Whether options are visible
}

// ProviderSettingsState holds loading state and data for a provider's settings.
type ProviderSettingsState struct {
	LoadingSettings    bool                             // Tracks if settings are being loaded
	SettingsFetchError string                           // Tracks fetch errors
	Descriptions       map[string]k8cginkgo.Description // Loaded descriptions
}

// DatacenterSettingsSelection holds the state for datacenter settings selection stage.
type DatacenterSettingsSelection struct {
	Providers          []string                          // List of provider names
	SettingsByProvider map[string][]SettingGroup         // Settings grouped by provider
	Selected           map[string]bool                   // Selected options (key: "provider:groupKey:option")
	SelectedGroups     map[string]bool                   // Selected groups (key: "provider:groupKey")
	FocusedIndex       int                               // Currently focused item index
	ExpandedProviders  map[string]bool                   // Tracks which providers are expanded
	ProviderSettings   map[string]*ProviderSettingsState // Per-provider loading state and data
}

// MachineDeploymentSettingsSelection holds the state for machine deployment settings selection stage.
type MachineDeploymentSettingsSelection struct {
	Providers          []string                          // List of provider names
	SettingsByProvider map[string][]SettingGroup         // Settings grouped by provider
	Selected           map[string]bool                   // Selected options (key: "provider:groupKey:option")
	SelectedGroups     map[string]bool                   // Selected groups (key: "provider:groupKey")
	FocusedIndex       int                               // Currently focused item index
	ExpandedProviders  map[string]bool                   // Tracks which providers are expanded
	ProviderSettings   map[string]*ProviderSettingsState // Per-provider loading state and data
}

// ClusterSettingsSelection holds the state for cluster settings selection stage.
type ClusterSettingsSelection struct {
	Providers          []string                  // List of provider names
	SettingsByProvider map[string][]SettingGroup // Settings grouped by provider
	Selected           map[string]bool           // Selected options (key: "provider:groupKey:option")
	SelectedGroups     map[string]bool           // Selected groups (key: "provider:groupKey")
	FocusedIndex       int                       // Currently focused item index
	ExpandedProviders  map[string]bool           // Tracks which providers are expanded
}

// DistributionSelection holds the state for OS distribution selection.
type DistributionSelection struct {
	Providers               []string                                    // List of provider names
	DistributionsByProvider map[string][]providerconfig.OperatingSystem // Distributions grouped by provider
	DistributionNames       map[providerconfig.OperatingSystem]string   // Display names
	Selected                map[string]bool                             // Selected distributions (key: "provider:distribution")
	FocusedIndex            int                                         // Currently focused item index
	ExpandedProviders       map[string]bool                             // Tracks which providers are expanded
}

// ClusterConfigurationSettings holds the state for cluster configuration stage.
type ClusterConfigurationSettings struct {
	Categories         []ConfigCategory
	FocusedIndex       int             // Currently focused item index
	EditMode           bool            // Whether we're editing a field
	EditingBuffer      string          // Buffer for editing values
	ExpandedCategories map[string]bool // Map of category names to their expanded state
}

// ConfigCategory represents a group of related configuration settings.
type ConfigCategory struct {
	Name        string
	Description string
	Settings    []ConfigSetting
}

// ConfigSetting represents a single configuration setting.
type ConfigSetting struct {
	Name        string
	Description string
	Type        ConfigSettingType
	Value       interface{} // Can be string, int, []int, []string, or bool
}

// ConfigSettingType defines the type of configuration setting.
type ConfigSettingType int

const (
	ConfigTypeString ConfigSettingType = iota
	ConfigTypeInt
	ConfigTypeIntArray
	ConfigTypeStringArray
	ConfigTypeBool
)

type Review struct {
	ConfigYAML        string
	Viewport          viewport.Model
	ProviderReviews   []ProviderReview
	ExpandedProviders map[string]bool // Tracks which providers are expanded
	ExpandedSections  map[string]bool // Tracks which sections are expanded (key: "provider:section")
	FocusedIndex      int             // Global index for navigation
	SaveToFile        bool            // Whether to save configurations to files
}

type ProviderReview struct {
	ProviderName string
	Sections     []ReviewSection
}

type ReviewSection struct {
	Name    string
	Content string
}

// Add tuiLogWriter type.
type tuiLogWriter struct {
	ch chan<- string
}

func (w *tuiLogWriter) Write(p []byte) (n int, err error) {
	lines := strings.Split(string(p), "\n")
	for _, line := range lines {
		if line != "" {
			w.ch <- line
		}
	}
	return len(p), nil
}
