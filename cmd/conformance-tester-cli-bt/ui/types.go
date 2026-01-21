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
	"k8c.io/machine-controller/sdk/providerconfig"
)

// Message types for execution.
type (
	logMsg   struct{ line string }
	errMsg   struct{ err error }
	doneMsg  struct{ success bool }
	startMsg struct{ ch <-chan tea.Msg }
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
	SeedName                   textinput.Model
	PresetName                 textinput.Model
	ProjectName                textinput.Model
	Errors                     EnvironmentExistingErrors
}

type EnvironmentExistingErrors struct {
	KubeconfigPath string
	SeedName       string
	PresetName     string
	ProjectName    string
}

// Provider holds configuration for a single cloud provider.
type Provider struct {
	CloudProvider providerconfig.CloudProvider
	DisplayName   string
	Selected      bool
	CurrentField  int
	Credentials   interface{} // Will hold provider-specific credentials
	Errors        ProviderErrors
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

// SettingGroup represents a group of related settings with a parent name and child options.
type SettingGroup struct {
	Key        string   // Unique key from map (e.g., "cpuTopology")
	Name       string   // Display name (e.g., "CPU Topology")
	Options    []string // Child options (e.g., ["threads", "cores", "sockets"])
	IsExpanded bool     // Whether options are visible
}

// DatacenterSettingsSelection holds the state for datacenter settings selection stage.
type DatacenterSettingsSelection struct {
	Providers          []string                  // List of provider names
	SettingsByProvider map[string][]SettingGroup // Settings grouped by provider
	Selected           map[string]bool           // Selected options (key: "provider:groupKey:option")
	SelectedGroups     map[string]bool           // Selected groups (key: "provider:groupKey")
	FocusedIndex       int                       // Currently focused item index
	ExpandedProviders  map[string]bool           // Tracks which providers are expanded
}

// MachineDeploymentSettingsSelection holds the state for machine deployment settings selection stage.
type MachineDeploymentSettingsSelection struct {
	Providers          []string                  // List of provider names
	SettingsByProvider map[string][]SettingGroup // Settings grouped by provider
	Selected           map[string]bool           // Selected options (key: "provider:groupKey:option")
	SelectedGroups     map[string]bool           // Selected groups (key: "provider:groupKey")
	FocusedIndex       int                       // Currently focused item index
	ExpandedProviders  map[string]bool           // Tracks which providers are expanded
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
	ConfigYAML string
	Viewport   viewport.Model
}

type execOutputMsg struct {
	output  string
	success bool
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
