/*
                  Kubermatic Enterprise Read-Only License
                         Version 1.0 ("KERO-1.0")
                     Copyright © 2025 Kubermatic GmbH

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
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-logr/logr"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	ginkgoutils "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/utils"
	"k8c.io/kubermatic/v2/pkg/defaulting"
	"k8c.io/machine-controller/sdk/providerconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	ctrlruntimelog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Model holds the state of the app.
type Model struct {
	stage int

	localEnv              EnvironmentLocal
	existingEnv           EnvironmentExisting
	environmentFocusIndex int // 0 for Local Environment, 1 for Existing Cluster
	environmentFieldIndex int // Index of the field within the selected environment

	releaseSelection ReleaseSelection

	providers             []Provider
	providerFocusIndex    int // Index of currently focused provider
	providerFieldIndex    int // Index of the field within the selected provider
	expandedProviderIndex int // Index of the provider whose credentials are currently expanded (-1 if none)

	distributionSelection DistributionSelection

	datacenterSettingsSelection DatacenterSettingsSelection

	clusterSettingsSelection ClusterSettingsSelection

	machineDeploymentSettingsSelection MachineDeploymentSettingsSelection

	clusterConfiguration ClusterConfigurationSettings

	Review  Review
	cmdChan <-chan tea.Msg
	program *tea.Program // Reference to the program for sending messages from goroutines

	// Execution state
	logs                []string
	executionError      string
	executionDone       bool
	runningJobs         []string          // Track job names currently running
	jobConfigMaps       map[string]string // Map job name to configmap name
	jobSecrets          map[string]string // Map job name to secret name
	executionCancelling bool              // Track if cancellation is in progress

	// Quit confirmation modal state
	quitConfirmVisible bool
	quitConfirmIndex   int

	// View modal for showing long credential content (e.g. kubeconfig, service account)
	viewModalVisible  bool
	viewModalTitle    string
	viewModalContent  string
	viewModalViewport viewport.Model

	// Terminal dimensions for dynamic sizing
	terminalWidth  int
	terminalHeight int
}

func newTextInput(placeholder string, charLimit int) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = charLimit
	ti.Width = 50
	return ti
}

// newTextInputWithMask creates a text input with optional password masking.
func newTextInputWithMask(placeholder string, charLimit int, masked bool) textinput.Model {
	ti := newTextInput(placeholder, charLimit)
	if masked {
		ti.EchoMode = textinput.EchoPassword
		ti.EchoCharacter = '•'
	}
	return ti
}

// latestDefaultKubernetesVersion returns the highest Kubernetes version known to
// the defaulting package, used as a fallback when the user selects nothing.
func latestDefaultKubernetesVersion() string {
	sel := initializeReleaseSelection()
	if len(sel.MajorVersions) > 0 {
		major := sel.MajorVersions[0]
		if len(sel.MinorVersions[major]) > 0 {
			return sel.MinorVersions[major][0]
		}
	}
	return "1.31.0" // absolute last resort
}

// initializeReleaseSelection creates the release selection structure from defaulting package.
func initializeReleaseSelection() ReleaseSelection {
	versions := defaulting.DefaultKubernetesVersioning.Versions

	// Group versions by major.minor (e.g., "1.31")
	majorVersionMap := make(map[string][]string)
	majorVersions := []string{}
	seenMajor := make(map[string]bool)

	for _, version := range versions {
		// Parse version to extract major.minor
		// Assuming format is "1.31.1", "1.32.0", etc.
		versionStr := version.String()
		parts := strings.Split(versionStr, ".")
		if len(parts) >= 2 {
			majorMinor := parts[0] + "." + parts[1]

			if !seenMajor[majorMinor] {
				majorVersions = append(majorVersions, majorMinor)
				seenMajor[majorMinor] = true
			}

			majorVersionMap[majorMinor] = append(majorVersionMap[majorMinor], versionStr)
		}
	}

	// Sort major versions in descending order (newest first)
	sort.Slice(majorVersions, func(i, j int) bool {
		return majorVersions[i] > majorVersions[j]
	})

	// Sort minor versions within each major version in descending order
	for major := range majorVersionMap {
		sort.Slice(majorVersionMap[major], func(i, j int) bool {
			return majorVersionMap[major][i] > majorVersionMap[major][j]
		})
	}

	return ReleaseSelection{
		MajorVersions:     majorVersions,
		MinorVersions:     majorVersionMap,
		SelectedMajor:     make(map[string]bool),
		SelectedMinor:     make(map[string]bool),
		FocusedMajorIndex: 0,
		FocusedMinorIndex: 0,
		IsMinorFocused:    false,
	}
}

// providerDistributionCompatibility defines which distributions are supported by each provider.
var providerDistributionCompatibility = map[string][]providerconfig.OperatingSystem{
	"AWS": {
		providerconfig.OperatingSystemUbuntu,
		providerconfig.OperatingSystemFlatcar,
		providerconfig.OperatingSystemAmazonLinux2,
		providerconfig.OperatingSystemRHEL,
		providerconfig.OperatingSystemRockyLinux,
	},
	"Alibaba": {
		providerconfig.OperatingSystemUbuntu,
	},
	"Anexia": {
		providerconfig.OperatingSystemUbuntu,
	},
	"Azure": {
		providerconfig.OperatingSystemUbuntu,
		providerconfig.OperatingSystemFlatcar,
		providerconfig.OperatingSystemRHEL,
		providerconfig.OperatingSystemRockyLinux,
	},
	"DigitalOcean": {
		providerconfig.OperatingSystemUbuntu,
		providerconfig.OperatingSystemRockyLinux,
	},
	"GCP": {
		providerconfig.OperatingSystemUbuntu,
		providerconfig.OperatingSystemFlatcar,
	},
	"Hetzner": {
		providerconfig.OperatingSystemUbuntu,
		providerconfig.OperatingSystemRockyLinux,
	},
	"KubeVirt": {
		providerconfig.OperatingSystemUbuntu,
		providerconfig.OperatingSystemFlatcar,
		providerconfig.OperatingSystemRHEL,
		providerconfig.OperatingSystemRockyLinux,
	},
	"Nutanix": {
		providerconfig.OperatingSystemUbuntu,
	},
	"OpenStack": {
		providerconfig.OperatingSystemUbuntu,
		providerconfig.OperatingSystemFlatcar,
		providerconfig.OperatingSystemRHEL,
		providerconfig.OperatingSystemRockyLinux,
	},
	"VMware Cloud Director": {
		providerconfig.OperatingSystemUbuntu,
		providerconfig.OperatingSystemFlatcar,
	},
	"vSphere": {
		providerconfig.OperatingSystemUbuntu,
		providerconfig.OperatingSystemFlatcar,
		providerconfig.OperatingSystemRHEL,
		providerconfig.OperatingSystemRockyLinux,
	},
}

// initializeDistributionSelection creates the distribution selection structure.
func initializeDistributionSelection(providers []string) DistributionSelection {
	// Create display names map
	displayNames := map[providerconfig.OperatingSystem]string{
		providerconfig.OperatingSystemUbuntu:       "Ubuntu",
		providerconfig.OperatingSystemAmazonLinux2: "Amazon Linux 2",
		providerconfig.OperatingSystemRHEL:         "Red Hat Enterprise Linux (RHEL)",
		providerconfig.OperatingSystemFlatcar:      "Flatcar Container Linux",
		providerconfig.OperatingSystemRockyLinux:   "Rocky Linux",
	}

	distributionsByProvider := make(map[string][]providerconfig.OperatingSystem)

	// Get distributions for each selected provider
	for _, provider := range providers {
		if supportedDists, exists := providerDistributionCompatibility[provider]; exists {
			distributionsByProvider[provider] = supportedDists
		}
	}

	// Initialize all providers as expanded by default
	expandedProviders := make(map[string]bool)
	for _, provider := range providers {
		expandedProviders[provider] = true
	}

	return DistributionSelection{
		Providers:               providers,
		DistributionsByProvider: distributionsByProvider,
		DistributionNames:       displayNames,
		Selected:                make(map[string]bool),
		FocusedIndex:            0,
		ExpandedProviders:       expandedProviders,
	}
}

// initializeDatacenterSettingsSelection creates the datacenter settings selection structure.
// Settings will be fetched lazily when entering this stage.
func initializeDatacenterSettingsSelection(providers []string) DatacenterSettingsSelection {
	settingsByProvider := make(map[string][]SettingGroup)

	// Initialize empty settings for all providers - will be populated later via lazy loading
	for _, provider := range providers {
		settingsByProvider[provider] = []SettingGroup{}
	}

	// Initialize all providers as expanded by default
	expandedProviders := make(map[string]bool)
	providerSettings := make(map[string]*ProviderSettingsState)
	for _, provider := range providers {
		expandedProviders[provider] = true
		providerSettings[provider] = &ProviderSettingsState{
			LoadingSettings:    false,
			SettingsFetchError: "",
			Descriptions:       make(map[string]k8cginkgo.Description),
		}
	}

	return DatacenterSettingsSelection{
		Providers:          providers,
		SettingsByProvider: settingsByProvider,
		Selected:           make(map[string]bool),
		SelectedGroups:     make(map[string]bool),
		FocusedIndex:       0,
		ExpandedProviders:  expandedProviders,
		ProviderSettings:   providerSettings,
	}
}

// initializeMachineDeploymentSettingsSelection creates the machine deployment settings selection structure.
// Settings will be fetched lazily when entering this stage.
func initializeMachineDeploymentSettingsSelection(providers []string) MachineDeploymentSettingsSelection {
	settingsByProvider := make(map[string][]SettingGroup)

	// Initialize empty settings for all providers - will be populated later via lazy loading
	for _, provider := range providers {
		settingsByProvider[provider] = []SettingGroup{}
	}

	// Initialize all providers as expanded by default
	expandedProviders := make(map[string]bool)
	providerSettings := make(map[string]*ProviderSettingsState)
	for _, provider := range providers {
		expandedProviders[provider] = true
		providerSettings[provider] = &ProviderSettingsState{
			LoadingSettings:    false,
			SettingsFetchError: "",
			Descriptions:       make(map[string]k8cginkgo.Description),
		}
	}

	return MachineDeploymentSettingsSelection{
		Providers:          providers,
		SettingsByProvider: settingsByProvider,
		Selected:           make(map[string]bool),
		SelectedGroups:     make(map[string]bool),
		FocusedIndex:       0,
		ExpandedProviders:  expandedProviders,
		ProviderSettings:   providerSettings,
	}
}

// initializeClusterSettingsSelection creates the cluster settings selection structure.
func initializeClusterSettingsSelection(providers []string) ClusterSettingsSelection {
	descriptionsMap := ginkgoutils.GetClusterDescriptions(nil)

	// Convert map to SettingGroup slice (same for all providers)
	var groups []SettingGroup
	keys := make([]string, 0, len(descriptionsMap))
	for key := range descriptionsMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		desc := descriptionsMap[key]
		groups = append(groups, SettingGroup{
			Key:        key,
			Name:       desc.Name,
			Options:    desc.Options,
			IsExpanded: true, // Always show options
		})
	}

	// Create settings map for each provider (same settings for all)
	settingsMap := make(map[string][]SettingGroup)
	for _, provider := range providers {
		settingsMap[provider] = groups
	}

	// Initialize expanded providers map (all expanded by default)
	expandedMap := make(map[string]bool)
	for _, provider := range providers {
		expandedMap[provider] = true
	}

	return ClusterSettingsSelection{
		Providers:          providers,
		SettingsByProvider: settingsMap,
		Selected:           make(map[string]bool),
		SelectedGroups:     make(map[string]bool),
		FocusedIndex:       0,
		ExpandedProviders:  expandedMap,
	}
}

// initializeClusterConfiguration creates the cluster configuration structure.
func initializeClusterConfiguration() ClusterConfigurationSettings {
	expandedCategories := map[string]bool{
		"Cluster Naming":       true,
		"Machine Deployment":   true,
		"Resource Allocation":  true,
		"Test Options":         true,
		"Output Configuration": true,
	}

	return ClusterConfigurationSettings{
		Categories: []ConfigCategory{
			{
				Name:        "Cluster Naming",
				Description: "Configure how user clusters will be named during testing",
				Settings: []ConfigSetting{
					{
						Name:        "Name Prefix",
						Description: "Prefix for created user cluster names (e.g., 'test-cluster' creates test-cluster-1, test-cluster-2, etc.)",
						Type:        ConfigTypeString,
						Value:       "conformance-test",
					},
				},
			},
			{
				Name:        "Machine Deployment",
				Description: "Configure the worker nodes for each user cluster",
				Settings: []ConfigSetting{
					{
						Name:        "Node Count",
						Description: "Number of machine deployment replicas in each user cluster",
						Type:        ConfigTypeInt,
						Value:       3,
					},
				},
			},
			{
				Name:        "Resource Allocation",
				Description: "Configure resource requirements for worker nodes. Multiple values create separate test scenarios",
				Settings: []ConfigSetting{
					{
						Name:        "CPU Cores",
						Description: "CPU cores per worker node (e.g., 2, 4, 8). Comma-separated for multiple values",
						Type:        ConfigTypeIntArray,
						Value:       []int{2},
					},
					{
						Name:        "Memory",
						Description: "RAM per worker node (e.g., 4Gi, 8Gi, 16Gi). Comma-separated for multiple values",
						Type:        ConfigTypeStringArray,
						Value:       []string{"4Gi"},
					},
					{
						Name:        "Disk Size",
						Description: "Disk size per worker node (e.g., 25Gi, 50Gi, 100Gi). Comma-separated for multiple values",
						Type:        ConfigTypeStringArray,
						Value:       []string{"25Gi"},
					},
				},
			},
			{
				Name:        "Test Options",
				Description: "Additional testing features and configurations",
				Settings: []ConfigSetting{
					{
						Name:        "Test Cluster Update",
						Description: "Test Kubernetes version upgrades by updating clusters after initial deployment",
						Type:        ConfigTypeBool,
						Value:       false,
					},
					{
						Name:        "Enable Dual Stack",
						Description: "Enable IPv4/IPv6 dual-stack networking for created clusters",
						Type:        ConfigTypeBool,
						Value:       false,
					},
				},
			},
			{
				Name:        "Output Configuration",
				Description: "Configure paths for test results, reports, and logs",
				Settings: []ConfigSetting{
					{
						Name:        "Results File",
						Description: "File to write test results to (relative to current directory)",
						Type:        ConfigTypeString,
						Value:       "results.json",
					},
					{
						Name:        "Reports Directory",
						Description: "Path to save test reports (relative to current directory)",
						Type:        ConfigTypeString,
						Value:       "_reports",
					},
					{
						Name:        "Log Directory",
						Description: "Path to save execution logs (relative to current directory)",
						Type:        ConfigTypeString,
						Value:       "_logs",
					},
				},
			},
		},
		FocusedIndex:       0,
		EditMode:           false,
		EditingBuffer:      "",
		ExpandedCategories: expandedCategories,
	}
}

// providerDisplayMap maps cloud providers to their display names.
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

// initializeProviders creates the initial list of providers with their text inputs.
func initializeProviders() []Provider {
	providers := []Provider{}

	for cloudProvider, displayName := range providerDisplayMap {
		provider := Provider{
			CloudProvider: cloudProvider,
			DisplayName:   displayName,
			Selected:      false,
			CurrentField:  0,
			Errors: ProviderErrors{
				Fields: make(map[string]string),
			},
		}

		// Initialize credentials based on provider type
		switch cloudProvider {
		case providerconfig.CloudProviderAWS:
			provider.Credentials = AWSCredentials{
				AccessKeyID:          newTextInput("Access Key ID", 256),
				SecretAccessKey:      newTextInputWithMask("Secret Access Key", 256, true),
				AssumeRoleARN:        newTextInput("Assume Role ARN (optional)", 256),
				AssumeRoleExternalID: newTextInput("External ID (optional)", 256),
			}
		case providerconfig.CloudProviderAzure:
			provider.Credentials = AzureCredentials{
				TenantID:       newTextInput("Tenant ID", 256),
				SubscriptionID: newTextInput("Subscription ID", 256),
				ClientID:       newTextInput("Client ID", 256),
				ClientSecret:   newTextInputWithMask("Client Secret", 256, true),
			}
		case providerconfig.CloudProviderGoogle:
			provider.Credentials = GCPCredentials{
				ServiceAccount: newTextInput("Service Account JSON Path", 256),
			}
		case providerconfig.CloudProviderAlibaba:
			provider.Credentials = AlibabaCredentials{
				AccessKeyID:     newTextInput("Access Key ID", 256),
				AccessKeySecret: newTextInputWithMask("Access Key Secret", 256, true),
			}
		case providerconfig.CloudProviderAnexia:
			provider.Credentials = AnexiaCredentials{
				Token: newTextInputWithMask("API Token", 256, true),
			}
		case providerconfig.CloudProviderDigitalocean:
			provider.Credentials = DigitalOceanCredentials{
				Token: newTextInputWithMask("API Token", 256, true),
			}
		case providerconfig.CloudProviderHetzner:
			provider.Credentials = HetznerCredentials{
				Token: newTextInputWithMask("API Token", 256, true),
			}
		case providerconfig.CloudProviderKubeVirt:
			provider.Credentials = KubeVirtCredentials{
				Kubeconfig: newTextInput("Kubeconfig Path", 256),
			}
		case providerconfig.CloudProviderNutanix:
			provider.Credentials = NutanixCredentials{
				Username:    newTextInput("Username", 256),
				Password:    newTextInputWithMask("Password", 256, true),
				ClusterName: newTextInput("Cluster Name", 256),
				ProxyURL:    newTextInput("Proxy URL (optional)", 256),
				CSIUsername: newTextInput("CSI Username (optional)", 256),
				CSIPassword: newTextInputWithMask("CSI Password (optional)", 256, true),
				CSIEndpoint: newTextInput("CSI Endpoint (optional)", 256),
			}
		case providerconfig.CloudProviderOpenstack:
			provider.Credentials = OpenStackCredentials{
				Username:                    newTextInput("Username", 256),
				Password:                    newTextInputWithMask("Password", 256, true),
				Project:                     newTextInput("Project", 256),
				ProjectID:                   newTextInput("Project ID (optional)", 256),
				Domain:                      newTextInput("Domain", 256),
				ApplicationCredentialID:     newTextInput("App Credential ID (optional)", 256),
				ApplicationCredentialSecret: newTextInputWithMask("App Credential Secret (optional)", 256, true),
				Token:                       newTextInputWithMask("Token (optional)", 256, true),
			}
		case providerconfig.CloudProviderVsphere:
			provider.Credentials = VSphereCredentials{
				Username: newTextInput("Username", 256),
				Password: newTextInputWithMask("Password", 256, true),
			}
		case providerconfig.CloudProviderVMwareCloudDirector:
			provider.Credentials = VMwareCloudDirectorCredentials{
				Username:     newTextInput("Username", 256),
				Password:     newTextInputWithMask("Password", 256, true),
				APIToken:     newTextInputWithMask("API Token (optional)", 256, true),
				Organization: newTextInput("Organization", 256),
				VDC:          newTextInput("VDC", 256),
			}
		}

		providers = append(providers, provider)
	}

	// Sort providers by display name
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].DisplayName < providers[j].DisplayName
	})

	return providers
}

// discoverKubeconfigOptions discovers available kubeconfig options.
func discoverKubeconfigOptions() []KubeconfigOption {
	options := []KubeconfigOption{}

	// Option 1: Environment variable KUBECONFIG
	if envPath := os.Getenv("KUBECONFIG"); envPath != "" {
		options = append(options, KubeconfigOption{
			Type:        "env",
			DisplayName: "Environment Variable (KUBECONFIG)",
			Path:        envPath,
			Selected:    false, // Don't select by default
		})
	}

	// Option 2: Discover kubeconfigs in ~/.kube directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		kubeDir := filepath.Join(homeDir, ".kube")
		if entries, err := os.ReadDir(kubeDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					// Skip common non-kubeconfig files
					name := entry.Name()
					if name == "cache" || name == "http-cache" {
						continue
					}

					fullPath := filepath.Join(kubeDir, name)
					displayName := fmt.Sprintf("~/.kube/%s", name)

					options = append(options, KubeconfigOption{
						Type:        "file",
						DisplayName: displayName,
						Path:        fullPath,
						Selected:    false,
					})
				}
			}
		}
	}

	// Option 3: Custom path
	options = append(options, KubeconfigOption{
		Type:        "custom",
		DisplayName: "Custom Path",
		Path:        "",
		Selected:    false,
	})

	// Don't select any option by default - let user choose
	// Removed: if len(options) > 0 && options[0].Type != "env" {
	//     options[0].Selected = true
	// }

	return options
}

func initialModel() Model {
	model := Model{
		stage: stageWelcome,
		localEnv: EnvironmentLocal{
			Selected:                     false,
			CurrentField:                 0,
			KubermaticConfigurationsPath: newTextInput("e.g., /path/to/kubermatic.yaml", 256),
			HelmValuesPath:               newTextInput("e.g., /path/to/values.yaml", 256),
			MLAValuesPath:                newTextInput("e.g., /path/to/mla-values.yaml", 256),
			Errors:                       EnvironmentLocalErrors{},
		},
		existingEnv: EnvironmentExisting{
			Selected:               false,
			CurrentField:           0,
			KubeconfigOptions:      discoverKubeconfigOptions(),
			KubeconfigFocusedIndex: 0,
			KubeconfigExpandedSections: map[string]bool{
				"env":    false,
				"file":   false,
				"custom": false,
			},
			CustomKubeconfigPath: newTextInput("Enter path to Kubeconfig", 500),
			AvailableSeeds:       []string{},
			SeedFocusedIndex:     0,
			SelectedSeedIndex:    -1,
			AvailablePresets:     []string{},
			PresetFocusedIndex:   0,
			SelectedPresetIndex:  -1,
			LoadingSeeds:         false,
			LoadingPresets:       false,
			FetchError:           "",
			ProjectName:          newTextInput("Leave empty for a new project", 64),
			Errors:               EnvironmentExistingErrors{Fields: make(map[string]string)},
		},
		environmentFocusIndex: 0,
		environmentFieldIndex: 0,
		releaseSelection:      initializeReleaseSelection(),
		providers:             initializeProviders(),
		distributionSelection: initializeDistributionSelection([]string{}), // Will be reinitialized after provider selection
		providerFocusIndex:    0,
		providerFieldIndex:    0,
		expandedProviderIndex: -1,
		Review: Review{
			ExpandedProviders: make(map[string]bool),
			ExpandedSections:  make(map[string]bool),
			FocusedIndex:      0,
		},
		runningJobs:         []string{},
		jobConfigMaps:       make(map[string]string),
		jobSecrets:          make(map[string]string),
		executionCancelling: false,
	}

	return model
}

// InitViewport initializes the viewport with content and dimensions.
func (m *Model) InitViewport(content string, width, height int) {
	m.Review.Viewport = viewport.New(width, height)
	m.Review.Viewport.SetContent(content)
	// Enable mouse wheel scrolling
	m.Review.Viewport.YPosition = 0
}

func (m Model) Init() tea.Cmd {
	return nil
}

// getSelectedKubeconfigPath returns the path of the selected kubeconfig.
func (m *Model) getSelectedKubeconfigPath() string {
	for _, option := range m.existingEnv.KubeconfigOptions {
		if option.Selected {
			if option.Type == "custom" {
				return m.existingEnv.CustomKubeconfigPath.Value()
			}
			return option.Path
		}
	}
	return ""
}

// fetchSeedsAndPresets fetches Seeds and Presets from the Kubernetes cluster.
func (m *Model) fetchSeedsAndPresets() tea.Cmd {
	return func() tea.Msg {
		// Get the selected kubeconfig path
		kubeconfigPath := m.getSelectedKubeconfigPath()

		if kubeconfigPath == "" {
			return seedsPresetsLoadedMsg{
				err: fmt.Errorf("Please select a kubeconfig file"),
			}
		}

		// Expand the path if it contains ~
		if strings.HasPrefix(kubeconfigPath, "~/") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return seedsPresetsLoadedMsg{err: fmt.Errorf("Unable to access home directory")}
			}
			kubeconfigPath = filepath.Join(homeDir, kubeconfigPath[2:])
		}

		// Build config from kubeconfig file
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return seedsPresetsLoadedMsg{
				err: fmt.Errorf("Invalid kubeconfig file or cluster is not accessible"),
			}
		}

		// Create dynamic client
		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			return seedsPresetsLoadedMsg{
				err: fmt.Errorf("Unable to connect to the cluster. Please check your kubeconfig"),
			}
		}

		// Define GVRs for Seeds and Presets
		seedGVR := schema.GroupVersionResource{
			Group:    "kubermatic.k8c.io",
			Version:  "v1",
			Resource: "seeds",
		}
		presetGVR := schema.GroupVersionResource{
			Group:    "kubermatic.k8c.io",
			Version:  "v1",
			Resource: "presets",
		}

		ctx := context.Background()

		// Fetch Seeds
		seedList, err := dynamicClient.Resource(seedGVR).List(ctx, metav1.ListOptions{})
		if err != nil {
			return seedsPresetsLoadedMsg{
				err: fmt.Errorf("Unable to fetch Seeds. Ensure this is a Kubermatic cluster"),
			}
		}

		seeds := make([]string, 0, len(seedList.Items))
		for _, item := range seedList.Items {
			seeds = append(seeds, item.GetName())
		}
		sort.Strings(seeds)

		// Fetch Presets
		presetList, err := dynamicClient.Resource(presetGVR).List(ctx, metav1.ListOptions{})
		if err != nil {
			return seedsPresetsLoadedMsg{
				seeds: seeds,
				err:   fmt.Errorf("Unable to fetch Presets. Ensure this is a Kubermatic cluster"),
			}
		}

		presets := make([]string, 0, len(presetList.Items))
		for _, item := range presetList.Items {
			presets = append(presets, item.GetName())
		}
		sort.Strings(presets)

		return seedsPresetsLoadedMsg{
			seeds:   seeds,
			presets: presets,
			err:     nil,
		}
	}
}

// fetchPresetDetails fetches the full preset details from the cluster.
func (m *Model) fetchPresetDetails() tea.Cmd {
	return func() tea.Msg {
		if m.existingEnv.SelectedPresetIndex < 0 || m.existingEnv.SelectedPresetIndex >= len(m.existingEnv.AvailablePresets) {
			return presetDetailsLoadedMsg{
				err: fmt.Errorf("No preset selected"),
			}
		}

		presetName := m.existingEnv.AvailablePresets[m.existingEnv.SelectedPresetIndex]

		// Get the selected kubeconfig path
		kubeconfigPath := m.getSelectedKubeconfigPath()

		if kubeconfigPath == "" {
			return presetDetailsLoadedMsg{
				err: fmt.Errorf("No kubeconfig selected"),
			}
		}

		// Expand path if needed
		if strings.HasPrefix(kubeconfigPath, "~/") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return presetDetailsLoadedMsg{
					err: fmt.Errorf("Unable to access home directory"),
				}
			}
			kubeconfigPath = filepath.Join(homeDir, kubeconfigPath[2:])
		}

		// Build config
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return presetDetailsLoadedMsg{
				err: fmt.Errorf("Invalid kubeconfig file or cluster is not accessible"),
			}
		}

		// Create dynamic client
		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			return presetDetailsLoadedMsg{
				err: fmt.Errorf("Unable to connect to the cluster. Please check your kubeconfig"),
			}
		}

		ctx := context.Background()

		// Fetch the specific preset
		presetGVR := schema.GroupVersionResource{
			Group:    "kubermatic.k8c.io",
			Version:  "v1",
			Resource: "presets",
		}

		presetObj, err := dynamicClient.Resource(presetGVR).Get(ctx, presetName, metav1.GetOptions{})
		if err != nil {
			return presetDetailsLoadedMsg{
				err: fmt.Errorf("Unable to fetch preset details. Ensure this is a Kubermatic cluster"),
			}
		}

		// Parse the preset spec
		spec, found := presetObj.Object["spec"].(map[string]interface{})
		if !found {
			return presetDetailsLoadedMsg{
				err: fmt.Errorf("Invalid preset format"),
			}
		}

		return presetDetailsLoadedMsg{
			presetName: presetName,
			spec:       spec,
		}
	}
}

// updateProvidersWithPresetCredentials updates provider structs with credentials from the preset.
func (m *Model) updateProvidersWithPresetCredentials(presetSpec map[string]interface{}) {
	providerNameMap := map[string]string{
		"AWS":                   "aws",
		"Azure":                 "azure",
		"GCP":                   "gcp",
		"Alibaba":               "alibaba",
		"Anexia":                "anexia",
		"DigitalOcean":          "digitalocean",
		"Hetzner":               "hetzner",
		"KubeVirt":              "kubevirt",
		"Nutanix":               "nutanix",
		"OpenStack":             "openstack",
		"vSphere":               "vsphere",
		"VMware Cloud Director": "vmwareclouddirector",
	}

	for i := range m.providers {
		provider := &m.providers[i]
		presetKey, ok := providerNameMap[provider.DisplayName]
		if !ok {
			continue
		}

		// Check if preset has credentials for this provider
		if providerSpec, found := presetSpec[presetKey]; found && providerSpec != nil {
			specMap, ok := providerSpec.(map[string]interface{})
			if !ok {
				continue
			}

			// Parse credentials based on provider type
			var presetCreds interface{}
			switch provider.DisplayName {
			case "AWS":
				presetCreds = m.parseAWSPresetCredentials(specMap)
			case "Azure":
				presetCreds = m.parseAzurePresetCredentials(specMap)
			case "GCP":
				presetCreds = m.parseGCPPresetCredentials(specMap)
			case "Alibaba":
				presetCreds = m.parseAlibabaPresetCredentials(specMap)
			case "Anexia":
				presetCreds = m.parseAnexiaPresetCredentials(specMap)
			case "DigitalOcean":
				presetCreds = m.parseDigitalOceanPresetCredentials(specMap)
			case "Hetzner":
				presetCreds = m.parseHetznerPresetCredentials(specMap)
			case "KubeVirt":
				presetCreds = m.parseKubeVirtPresetCredentials(specMap)
			case "Nutanix":
				presetCreds = m.parseNutanixPresetCredentials(specMap)
			case "OpenStack":
				presetCreds = m.parseOpenStackPresetCredentials(specMap)
			case "vSphere":
				presetCreds = m.parseVSpherePresetCredentials(specMap)
			case "VMware Cloud Director":
				presetCreds = m.parseVMwareCloudDirectorPresetCredentials(specMap)
			}

			provider.HasPresetCredentials = true
			provider.PresetCredentials = presetCreds
			provider.CredentialSource = CredentialSourcePreset
		} else {
			provider.HasPresetCredentials = false
			provider.PresetCredentials = nil
			provider.CredentialSource = CredentialSourceCustom
		}
	}
}

// Parse functions for each provider
func (m *Model) parseAWSPresetCredentials(spec map[string]interface{}) AWSCredentials {
	accessKeyID := ""
	if val, ok := spec["accessKeyID"].(string); ok {
		accessKeyID = val
	}
	secretAccessKey := ""
	if val, ok := spec["secretAccessKey"].(string); ok {
		secretAccessKey = val
	}

	creds := AWSCredentials{
		AccessKeyID:          newTextInputReadOnly(accessKeyID, 200),
		SecretAccessKey:      newTextInputReadOnly(secretAccessKey, 200),
		AssumeRoleARN:        newTextInputReadOnly("", 200),
		AssumeRoleExternalID: newTextInputReadOnly("", 200),
	}

	if val, ok := spec["assumeRoleARN"].(string); ok && val != "" {
		creds.AssumeRoleARN = newTextInputReadOnly(val, 200)
	}
	if val, ok := spec["assumeRoleExternalID"].(string); ok && val != "" {
		creds.AssumeRoleExternalID = newTextInputReadOnly(val, 200)
	}

	return creds
}

func (m *Model) parseAzurePresetCredentials(spec map[string]interface{}) AzureCredentials {
	tenantID := ""
	if val, ok := spec["tenantID"].(string); ok {
		tenantID = val
	}
	subscriptionID := ""
	if val, ok := spec["subscriptionID"].(string); ok {
		subscriptionID = val
	}
	clientID := ""
	if val, ok := spec["clientID"].(string); ok {
		clientID = val
	}
	clientSecret := ""
	if val, ok := spec["clientSecret"].(string); ok {
		clientSecret = val
	}

	return AzureCredentials{
		TenantID:       newTextInputReadOnly(tenantID, 200),
		SubscriptionID: newTextInputReadOnly(subscriptionID, 200),
		ClientID:       newTextInputReadOnly(clientID, 200),
		ClientSecret:   newTextInputReadOnly(clientSecret, 200),
	}
}

func (m *Model) parseGCPPresetCredentials(spec map[string]interface{}) GCPCredentials {
	serviceAccount := ""
	if val, ok := spec["serviceAccount"].(string); ok {
		serviceAccount = val
	}

	return GCPCredentials{
		ServiceAccount: newTextInputReadOnly(serviceAccount, 500),
	}
}

func (m *Model) parseAlibabaPresetCredentials(spec map[string]interface{}) AlibabaCredentials {
	accessKeyID := ""
	if val, ok := spec["accessKeyID"].(string); ok {
		accessKeyID = val
	}
	accessKeySecret := ""
	if val, ok := spec["accessKeySecret"].(string); ok {
		accessKeySecret = val
	}

	return AlibabaCredentials{
		AccessKeyID:     newTextInputReadOnly(accessKeyID, 200),
		AccessKeySecret: newTextInputReadOnly(accessKeySecret, 200),
	}
}

func (m *Model) parseAnexiaPresetCredentials(spec map[string]interface{}) AnexiaCredentials {
	token := ""
	if val, ok := spec["token"].(string); ok {
		token = val
	}

	return AnexiaCredentials{
		Token: newTextInputReadOnly(token, 200),
	}
}

func (m *Model) parseDigitalOceanPresetCredentials(spec map[string]interface{}) DigitalOceanCredentials {
	token := ""
	if val, ok := spec["token"].(string); ok {
		token = val
	}

	return DigitalOceanCredentials{
		Token: newTextInputReadOnly(token, 200),
	}
}

func (m *Model) parseHetznerPresetCredentials(spec map[string]interface{}) HetznerCredentials {
	token := ""
	if val, ok := spec["token"].(string); ok {
		token = val
	}

	return HetznerCredentials{
		Token: newTextInputReadOnly(token, 200),
	}
}

func (m *Model) parseKubeVirtPresetCredentials(spec map[string]interface{}) KubeVirtCredentials {
	kubeconfig := ""
	if val, ok := spec["kubeconfig"].(string); ok {
		kubeconfig = val
	}

	return KubeVirtCredentials{
		Kubeconfig: newTextInputReadOnly(kubeconfig, 500),
	}
}

func (m *Model) parseNutanixPresetCredentials(spec map[string]interface{}) NutanixCredentials {
	username := ""
	if val, ok := spec["username"].(string); ok {
		username = val
	}
	password := ""
	if val, ok := spec["password"].(string); ok {
		password = val
	}
	clusterName := ""
	if val, ok := spec["clusterName"].(string); ok {
		clusterName = val
	}

	creds := NutanixCredentials{
		Username:    newTextInputReadOnly(username, 200),
		Password:    newTextInputReadOnly(password, 200),
		ClusterName: newTextInputReadOnly(clusterName, 200),
		ProxyURL:    newTextInputReadOnly("", 200),
		CSIUsername: newTextInputReadOnly("", 200),
		CSIPassword: newTextInputReadOnly("", 200),
		CSIEndpoint: newTextInputReadOnly("", 200),
	}

	if val, ok := spec["proxyURL"].(string); ok && val != "" {
		creds.ProxyURL = newTextInputReadOnly(val, 200)
	}
	if val, ok := spec["csiUsername"].(string); ok && val != "" {
		creds.CSIUsername = newTextInputReadOnly(val, 200)
	}
	if val, ok := spec["csiPassword"].(string); ok && val != "" {
		creds.CSIPassword = newTextInputReadOnly(val, 200)
	}
	if val, ok := spec["csiEndpoint"].(string); ok && val != "" {
		creds.CSIEndpoint = newTextInputReadOnly(val, 200)
	}

	return creds
}

func (m *Model) parseOpenStackPresetCredentials(spec map[string]interface{}) OpenStackCredentials {
	creds := OpenStackCredentials{
		Username:                    newTextInputReadOnly("", 200),
		Password:                    newTextInputReadOnly("", 200),
		Project:                     newTextInputReadOnly("", 200),
		ProjectID:                   newTextInputReadOnly("", 200),
		Domain:                      newTextInputReadOnly("", 200),
		ApplicationCredentialID:     newTextInputReadOnly("", 200),
		ApplicationCredentialSecret: newTextInputReadOnly("", 200),
		Token:                       newTextInputReadOnly("", 200),
	}

	if val, ok := spec["username"].(string); ok && val != "" {
		creds.Username = newTextInputReadOnly(val, 200)
	}
	if val, ok := spec["password"].(string); ok && val != "" {
		creds.Password = newTextInputReadOnly(val, 200)
	}
	if val, ok := spec["project"].(string); ok && val != "" {
		creds.Project = newTextInputReadOnly(val, 200)
	}
	if val, ok := spec["projectID"].(string); ok && val != "" {
		creds.ProjectID = newTextInputReadOnly(val, 200)
	}
	if val, ok := spec["domain"].(string); ok && val != "" {
		creds.Domain = newTextInputReadOnly(val, 200)
	}
	if val, ok := spec["applicationCredentialID"].(string); ok && val != "" {
		creds.ApplicationCredentialID = newTextInputReadOnly(val, 200)
	}
	if val, ok := spec["applicationCredentialSecret"].(string); ok && val != "" {
		creds.ApplicationCredentialSecret = newTextInputReadOnly(val, 200)
	}
	if val, ok := spec["token"].(string); ok && val != "" {
		creds.Token = newTextInputReadOnly(val, 200)
	}

	return creds
}

func (m *Model) parseVSpherePresetCredentials(spec map[string]interface{}) VSphereCredentials {
	username := ""
	if val, ok := spec["username"].(string); ok {
		username = val
	}
	password := ""
	if val, ok := spec["password"].(string); ok {
		password = val
	}

	return VSphereCredentials{
		Username: newTextInputReadOnly(username, 200),
		Password: newTextInputReadOnly(password, 200),
	}
}

func (m *Model) parseVMwareCloudDirectorPresetCredentials(spec map[string]interface{}) VMwareCloudDirectorCredentials {
	organization := ""
	if val, ok := spec["organization"].(string); ok {
		organization = val
	}
	vdc := ""
	if val, ok := spec["vdc"].(string); ok {
		vdc = val
	}

	creds := VMwareCloudDirectorCredentials{
		Username:     newTextInputReadOnly("", 200),
		Password:     newTextInputReadOnly("", 200),
		APIToken:     newTextInputReadOnly("", 200),
		Organization: newTextInputReadOnly(organization, 200),
		VDC:          newTextInputReadOnly(vdc, 200),
	}

	if val, ok := spec["username"].(string); ok && val != "" {
		creds.Username = newTextInputReadOnly(val, 200)
	}
	if val, ok := spec["password"].(string); ok && val != "" {
		creds.Password = newTextInputReadOnly(val, 200)
	}
	if val, ok := spec["apiToken"].(string); ok && val != "" {
		creds.APIToken = newTextInputReadOnly(val, 200)
	}

	return creds
}

// newTextInputReadOnly creates a read-only text input with a preset value.
// If HIDE_PRESET_CREDENTIALS environment variable is set to "true", the value is masked.
func newTextInputReadOnly(value string, width int) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Width = width

	// Check if preset credentials should be hidden
	hideCredentials := os.Getenv("HIDE_PRESET_CREDENTIALS") == "true"
	if hideCredentials && value != "" {
		ti.SetValue("***preset-value***")
	} else {
		ti.SetValue(value)
	}
	return ti
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global key events first
		if handled, globalCmd := m.handleGlobalKeys(msg); handled {
			return m, globalCmd
		}

		// Handle stage-specific key events
		switch m.stage {
		case stageWelcome:
			return m.handleWelcomePage(msg)
		case stageEnvironmentSelection:
			return m.handleEnvironmentSelection(msg)
		case stageReleaseSelection:
			return m.handleReleaseSelection(msg)
		case stageProviderSelection:
			return m.handleProviderSelection(msg)
		case stageDistributionSelection:
			return m.handleDistributionSelection(msg)
		case stageDatacenterSettingsSelection:
			return m.handleDatacenterSettingsSelection(msg)
		case stageClusterSettingsSelection:
			return m.handleClusterSettingsSelection(msg)
		case stageMachineDeploymentSettingsSelection:
			return m.handleMachineDeploymentSettingsSelection(msg)
		case stageClusterConfiguration:
			return m.handleClusterConfiguration(msg)
		case stageReviewSettings:
			return m.handleReviewSettings(msg)
		case stageExecuting:
			return m.handleExecuting(msg)
		case stageDone:
			return m.handleDone(msg)
		}

	// Handle mouse wheel events for scrolling
	case tea.MouseMsg:
		// Mouse events not used for scrolling are ignored
		return m, nil

	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height

		// Update execution viewport if it exists
		if m.stage == stageExecuting || m.stage == stageReviewSettings {
			if m.Review.Viewport.Width == 0 {
				m.Review.Viewport = viewport.New(m.getUIInnerWidth(), m.terminalHeight-12)
				m.Review.Viewport.SetContent(strings.Join(m.logs, "\n"))
			} else {
				m.Review.Viewport.Width = m.getUIInnerWidth()
				m.Review.Viewport.Height = m.terminalHeight - 12
			}
		}
		cmd = m.handleWindowSize(msg)
	case startMsg:
		cmd = m.handleStart(msg)
	case logMsg:
		cmd = m.handleLog(msg)
	// case errMsg:
	// 	cmd = m.handleError(msg)
	case doneMsg:
		cmd = m.handleDoneMessage(msg)
	case execOutputMsg:
		cmd = m.handleExecOutput(msg)
	case cleanupProgressMsg:
		// Append cleanup messages to logs
		lines := strings.Split(msg.message, "\n")
		for _, line := range lines {
			if line != "" {
				m.logs = append(m.logs, line)
			}
		}

		// Update viewport
		if m.Review.Viewport.Width > 0 {
			m.Review.Viewport.SetContent(strings.Join(m.logs, "\n"))
			m.Review.Viewport.GotoBottom()
		}

		if msg.done {
			m.executionCancelling = false
			m.executionDone = true
			if msg.err != nil {
				m.executionError = msg.err.Error()
			}
			m.stage = stageDone
		}
		return m, nil
	case datacenterSettingsLoadedMsg:
		// Update datacenter settings for the provider
		if ps, ok := m.datacenterSettingsSelection.ProviderSettings[msg.provider]; ok {
			ps.LoadingSettings = false
			if msg.err != nil {
				ps.SettingsFetchError = msg.err.Error()
			} else {
				ps.SettingsFetchError = ""
				ps.Descriptions = msg.descriptions
				// Convert descriptions to SettingGroup structures
				var settingGroups []SettingGroup
				for key, desc := range msg.descriptions {
					settingGroups = append(settingGroups, SettingGroup{
						Key:     key,
						Name:    desc.Name,
						Options: desc.Options,
					})
				}
				m.datacenterSettingsSelection.SettingsByProvider[msg.provider] = settingGroups
			}
			m.datacenterSettingsSelection.ProviderSettings[msg.provider] = ps
		}
	case machineSettingsLoadedMsg:
		// Update machine settings for the provider
		if ps, ok := m.machineDeploymentSettingsSelection.ProviderSettings[msg.provider]; ok {
			ps.LoadingSettings = false
			if msg.err != nil {
				ps.SettingsFetchError = msg.err.Error()
			} else {
				ps.SettingsFetchError = ""
				ps.Descriptions = msg.descriptions
				// Convert descriptions to SettingGroup structures
				var settingGroups []SettingGroup
				for key, desc := range msg.descriptions {
					settingGroups = append(settingGroups, SettingGroup{
						Key:     key,
						Name:    desc.Name,
						Options: desc.Options,
					})
				}
				m.machineDeploymentSettingsSelection.SettingsByProvider[msg.provider] = settingGroups
			}
			m.machineDeploymentSettingsSelection.ProviderSettings[msg.provider] = ps
		}
	case seedsPresetsLoadedMsg:
		m.existingEnv.LoadingSeeds = false
		m.existingEnv.LoadingPresets = false
		if msg.err != nil {
			m.existingEnv.FetchError = msg.err.Error()
		} else {
			m.existingEnv.AvailableSeeds = msg.seeds
			m.existingEnv.AvailablePresets = msg.presets
			m.existingEnv.FetchError = ""
			// Reset selections when new data is loaded
			m.existingEnv.SeedFocusedIndex = 0
			m.existingEnv.SelectedSeedIndex = -1
			m.existingEnv.PresetFocusedIndex = 0
			m.existingEnv.SelectedPresetIndex = -1
		}
	case presetDetailsLoadedMsg:
		if msg.err != nil {
			m.existingEnv.FetchError = msg.err.Error()
		} else {
			// Update providers with preset credentials
			m.updateProvidersWithPresetCredentials(msg.spec)
			m.existingEnv.FetchError = ""
		}
	}

	// Reset viewport scroll position when stage changes
	// Scrolling removed

	return m, cmd
}

// handleGlobalKeys handles global key events that work across all stages
func (m *Model) handleGlobalKeys(msg tea.KeyMsg) (bool, tea.Cmd) {
	// If view modal is visible, intercept keys for closing/scrolling
	if m.viewModalVisible {
		switch msg.String() {
		case keyEsc, keyEnter, keyNo, keyControlC:
			m.viewModalVisible = false
			return true, nil
		case keyUp:
			m.viewModalViewport.ScrollUp(1)
			return true, nil
		case keyDown:
			m.viewModalViewport.ScrollDown(1)
			return true, nil
		default:
			return true, nil
		}
	}

	// Global quit confirmation handler
	if m.quitConfirmVisible {
		if handled, cmd := m.handleQuitConfirmation(msg); handled {
			return true, cmd
		}
	}

	switch msg.String() {
	case keyControlC:
		// Show quit confirmation modal
		if m.stage != stageDone {
			m.quitConfirmVisible = true
			m.quitConfirmIndex = 0 // Default to "No"
		}
		return true, nil
	case keyQuit:
		// Immediate quit if we're at the done stage
		if m.stage == stageDone {
			return true, tea.Quit
		}
	}

	return false, nil
}

// getUIWidth returns the dynamic UI width based on terminal size.
// Falls back to 150 if terminal width hasn't been detected yet.
func (m Model) getUIWidth() int {
	if m.terminalWidth == 0 {
		return 150 // Default width
	}
	// Use 95% of terminal width, with minimum of 80 and maximum of 200
	width := int(float64(m.terminalWidth) * 0.95)
	if width < 80 {
		width = 80
	}
	if width > 200 {
		width = 200
	}
	return width
}

// getUIInnerWidth returns the inner width (accounting for box padding).
func (m Model) getUIInnerWidth() int {
	return m.getUIWidth() - 8
}

// getUIHeight returns the dynamic UI height based on terminal size.
// Falls back to 40 if terminal height hasn't been detected yet.
// This accounts for banner (5 lines) and help bar (3 lines) to fill the terminal.
func (m Model) getUIHeight() int {
	if m.terminalHeight == 0 {
		return 40 // Default height
	}
	// Reserve space for:
	// - Banner (~5 lines)
	// - Help bar (~3 lines)
	// - Borders, padding, and margins (~4 lines)
	// Box height = terminal height - 12
	height := m.terminalHeight - 12
	if height < 20 {
		height = 20
	}
	return height
}

// getUIInnerHeight returns the inner height (accounting for help bar and potential banner).
func (m Model) getUIInnerHeight() int {
	return m.getUIHeight() - 10 // Approximate space for banner and help bar
}

// SetProgram sets the tea.Program reference for sending messages from goroutines.
func (m *Model) SetProgram(p *tea.Program) {
	m.program = p
}

func ConformanceTester() (tea.Model, error) {
	// Set up a no-op logger for controller-runtime to suppress warnings
	// This is needed because some provider settings code tries to connect to k8s clusters
	ctrlruntimelog.SetLogger(logr.Discard())

	initialMdl := initialModel()
	// Create program with AltScreen enabled for full-screen TUI
	p := tea.NewProgram(
		&initialMdl,
		tea.WithAltScreen(),       // Enable alternate screen buffer (full-screen mode like Vim)
		tea.WithMouseCellMotion(), // Optional: enable mouse support
	)

	// Set the program reference in the model so goroutines can send messages
	initialMdl.SetProgram(p)

	m, err := p.Run()
	if err != nil {
		return nil, err
	}
	return m, nil
}

// --- View Entry Point ---
// View renders the entire UI based on the current application stage.
func (m Model) View() string {
	if m.terminalWidth == 0 || m.terminalHeight == 0 {
		return "Initializing..."
	}

	// Get dynamic UI dimensions
	uiWidth := m.getUIWidth()
	uiInnerWidth := m.getUIInnerWidth()

	helpWithBorder := m.renderCommonHelpBar(m.stage)

	var content string
	switch m.stage {
	case stageWelcome:
		content = m.renderWelcome("", uiWidth)
	case stageEnvironmentSelection:
		content = m.renderEnvironmentSelection("", uiWidth)
	case stageReleaseSelection:
		content = m.renderReleaseSelection("", uiWidth)
	case stageProviderSelection:
		content = m.renderProviderSelection("", uiWidth)
	case stageDistributionSelection:
		content = m.renderDistributionSelection("", uiWidth)
	case stageDatacenterSettingsSelection:
		content = m.renderDatacenterSettingsSelection("", uiWidth)
	case stageClusterSettingsSelection:
		content = m.renderClusterSettingsSelection("", uiWidth)
	case stageMachineDeploymentSettingsSelection:
		content = m.renderMachineDeploymentSettingsSelection("", uiWidth)
	case stageClusterConfiguration:
		content = m.renderClusterConfiguration("", uiWidth)
	case stageReviewSettings:
		content = m.renderReviewSettings("", uiWidth)
	case stageExecuting:
		content = m.renderExecuting("", uiWidth)
	case stageDone:
		content = m.renderDone("", uiWidth)
	default:
		// Render nothing for unknown stages
		os.Exit(0)
		return ""
	}

	if m.viewModalVisible {
		// Show view modal (full credential content)
		modal := m.renderViewModal(uiWidth, uiInnerWidth)
		bannerContent := styleBanner.Width(uiWidth).Render(bannerText())
		fullContent := bannerContent + "\n" + modal
		return m.renderCommonLayout(fullContent, helpWithBorder)
	}

	if m.quitConfirmVisible {
		// Show quit confirmation modal
		modal := m.renderQuitConfirm(uiWidth, uiInnerWidth)
		bannerContent := styleBanner.Width(uiWidth).Render(bannerText())
		fullContent := bannerContent + "\n" + modal
		return m.renderCommonLayout(fullContent, helpWithBorder)
	}

	// Render normal content with banner
	return m.renderCommonLayout(content, helpWithBorder)
}

// generateReviewYAML generates the YAML configuration from all selected settings, organized by provider.
func (m *Model) generateReviewYAML() []ProviderReview {
	providerReviews := []ProviderReview{}

	// Get selected providers
	selectedProviders := []Provider{}
	for _, provider := range m.providers {
		if provider.Selected {
			selectedProviders = append(selectedProviders, provider)
		}
	}

	// Generate review for each selected provider
	for _, provider := range selectedProviders {
		sections := []ReviewSection{}

		// Generate all sections for this provider
		sections = append(sections, m.generateConfigurationSectionForProvider(provider)...)
		sections = append(sections, m.generateEnvironmentSectionForProvider(provider))
		sections = append(sections, m.generateSecretsSectionForProvider(provider))
		sections = append(sections, m.generateReleasesSectionForProvider(provider))
		sections = append(sections, m.generateDistributionsSectionForProvider(provider))
		sections = append(sections, m.generateResourcesSectionForProvider(provider))
		sections = append(sections, m.generateIncludedSectionForProvider(provider))

		providerReviews = append(providerReviews, ProviderReview{
			ProviderName: provider.DisplayName,
			Sections:     sections,
		})
	}

	return providerReviews
}

// generateConfigurationSectionForProvider generates the configuration section for a specific provider.
func (m *Model) generateConfigurationSectionForProvider(provider Provider) []ReviewSection {
	var b strings.Builder

	// Get configuration values
	var (
		namePrefix           = "conformance-tester"
		resultsFile          = "results.json"
		reportsRoot          = "_reports"
		logDirectory         = "_logs"
		retryFailedScenarios = false
		keepFailedClusters   = true
		nodeCount            = 2
	)

	// Extract values from cluster configuration
	for _, category := range m.clusterConfiguration.Categories {
		for _, setting := range category.Settings {
			switch setting.Name {
			case "Name Prefix":
				if val, ok := setting.Value.(string); ok && val != "" {
					namePrefix = val
				}
			case "Results File":
				if val, ok := setting.Value.(string); ok && val != "" {
					resultsFile = val
				}
			case "Reports Directory":
				if val, ok := setting.Value.(string); ok && val != "" {
					reportsRoot = val
				}
			case "Log Directory":
				if val, ok := setting.Value.(string); ok && val != "" {
					logDirectory = val
				}
			case "Node Count":
				if val, ok := setting.Value.(int); ok {
					nodeCount = val
				}
			}
		}
	}

	// Build configuration sections
	sections := []ReviewSection{}

	// Basic configuration
	b.WriteString(fmt.Sprintf("namePrefix: \"%s\"\n", namePrefix))
	b.WriteString(fmt.Sprintf("resultsFile: \"%s\"\n", resultsFile))
	b.WriteString(fmt.Sprintf("retryFailedScenarios: %v\n", retryFailedScenarios))
	b.WriteString(fmt.Sprintf("keepFailedClustersAfterTests: %v\n", keepFailedClusters))
	b.WriteString(fmt.Sprintf("nodeCount: %d\n", nodeCount))
	b.WriteString(fmt.Sprintf("reportsRoot: \"%s\"\n", reportsRoot))
	b.WriteString(fmt.Sprintf("logDirectory: \"%s\"\n", logDirectory))

	sections = append(sections, ReviewSection{
		Name:    "Configuration",
		Content: b.String(),
	})

	return sections
}

// generateEnvironmentSectionForProvider generates the environment section for a specific provider.
func (m *Model) generateEnvironmentSectionForProvider(provider Provider) ReviewSection {
	var b strings.Builder

	if m.existingEnv.Selected {
		if m.existingEnv.SelectedSeedIndex >= 0 && m.existingEnv.SelectedSeedIndex < len(m.existingEnv.AvailableSeeds) {
			b.WriteString("kubermaticNamespace: \"kubermatic\"\n")
			b.WriteString(fmt.Sprintf("kubermaticSeedName: \"%s\"\n",
				m.existingEnv.AvailableSeeds[m.existingEnv.SelectedSeedIndex]))
		}
		if m.existingEnv.ProjectName.Value() != "" {
			b.WriteString(fmt.Sprintf("kubermaticProject: \"%s\"\n",
				m.existingEnv.ProjectName.Value()))
		} else {
			b.WriteString("kubermaticProject: \"\"\n")
		}
	} else {
		b.WriteString("kubermaticNamespace: \"kubermatic\"\n")
		b.WriteString("kubermaticSeedName: \"\"\n")
		b.WriteString("kubermaticProject: \"\"\n")
	}

	return ReviewSection{
		Name:    "Environment",
		Content: b.String(),
	}
}

// generateSecretsSectionForProvider generates the secrets section for a specific provider.
func (m *Model) generateSecretsSectionForProvider(provider Provider) ReviewSection {
	var b strings.Builder

	b.WriteString("secrets:\n")
	b.WriteString(fmt.Sprintf("  %s:\n", strings.ToLower(provider.DisplayName)))

	// Add credentials
	if provider.HasPresetCredentials && provider.CredentialSource == CredentialSourcePreset {
		// Credentials come from preset - use PresetCredentials
		if provider.PresetCredentials != nil {
			// Create a temporary provider with Credentials set to PresetCredentials
			tempProvider := Provider{
				DisplayName: provider.DisplayName,
				Credentials: provider.PresetCredentials,
			}
			m.addProviderCredentialsToSecretsYAML(&b, &tempProvider)
		}
	} else {
		m.addProviderCredentialsToSecretsYAML(&b, &provider)
	}

	return ReviewSection{
		Name:    "Secrets",
		Content: b.String(),
	}
}

// generateReleasesSectionForProvider generates the releases section for a specific provider.
func (m *Model) generateReleasesSectionForProvider(provider Provider) ReviewSection {
	var b strings.Builder

	selectedReleases := []string{}
	for major := range m.releaseSelection.SelectedMajor {
		for _, minor := range m.releaseSelection.MinorVersions[major] {
			if m.releaseSelection.SelectedMinor[minor] {
				selectedReleases = append(selectedReleases, minor)
			}
		}
	}

	if len(selectedReleases) == 0 {
		selectedReleases = []string{latestDefaultKubernetesVersion()}
	}

	b.WriteString("releases:\n")
	for _, release := range selectedReleases {
		b.WriteString(fmt.Sprintf("- \"%s\"\n", release))
	}

	return ReviewSection{
		Name:    "Releases",
		Content: b.String(),
	}
}

// generateDistributionsSectionForProvider generates the distributions section for a specific provider.
func (m *Model) generateDistributionsSectionForProvider(provider Provider) ReviewSection {
	var b strings.Builder

	selectedDistributions := []string{}
	// Only get distributions for this specific provider
	for _, dist := range m.distributionSelection.DistributionsByProvider[provider.DisplayName] {
		key := fmt.Sprintf("%s:%s", provider.DisplayName, dist)
		if m.distributionSelection.Selected[key] {
			selectedDistributions = append(selectedDistributions, string(dist))
		}
	}

	if len(selectedDistributions) == 0 {
		selectedDistributions = []string{"ubuntu"}
	}

	b.WriteString("enableDistributions:\n")
	for _, dist := range selectedDistributions {
		b.WriteString(fmt.Sprintf("- %s\n", dist))
	}

	return ReviewSection{
		Name:    "Distributions",
		Content: b.String(),
	}
}

// generateResourcesSectionForProvider generates the resources section for a specific provider.
func (m *Model) generateResourcesSectionForProvider(provider Provider) ReviewSection {
	var b strings.Builder

	var (
		cpuResources      []string
		memoryResources   []string
		diskSizeResources []string
	)

	// Extract resources from cluster configuration
	for _, category := range m.clusterConfiguration.Categories {
		for _, setting := range category.Settings {
			switch setting.Name {
			case "CPU Cores":
				if val, ok := setting.Value.([]int); ok {
					for _, cpu := range val {
						cpuResources = append(cpuResources, fmt.Sprintf("%dm", cpu*1000))
					}
				}
			case "Memory":
				if val, ok := setting.Value.([]string); ok {
					memoryResources = val
				}
			case "Disk Size":
				if val, ok := setting.Value.([]string); ok {
					diskSizeResources = val
				}
			}
		}
	}

	// Default values
	if len(cpuResources) == 0 {
		cpuResources = []string{"2"}
	}
	if len(memoryResources) == 0 {
		memoryResources = []string{"4Gi"}
	}
	if len(diskSizeResources) == 0 {
		diskSizeResources = []string{"25Gi"}
	}

	b.WriteString("resources:\n")
	b.WriteString("  cpu: [")
	for i, cpu := range cpuResources {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(cpu)
	}
	b.WriteString("]\n")

	b.WriteString("  memory: [")
	for i, mem := range memoryResources {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("\"%s\"", mem))
	}
	b.WriteString("]\n")

	b.WriteString("  diskSize: [")
	for i, disk := range diskSizeResources {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("\"%s\"", disk))
	}
	b.WriteString("]\n")

	return ReviewSection{
		Name:    "Resources",
		Content: b.String(),
	}
}

// generateIncludedSectionForProvider generates the included section for a specific provider.
func (m *Model) generateIncludedSectionForProvider(provider Provider) ReviewSection {
	var b strings.Builder

	b.WriteString("included:\n")

	// Datacenter descriptions
	datacenterDescriptions := m.generateDatacenterDescriptionsForProvider(provider)
	if len(datacenterDescriptions) > 0 {
		b.WriteString("  datacenterDescriptions:\n")
		for _, desc := range datacenterDescriptions {
			b.WriteString(fmt.Sprintf("    - \"%s\"\n", desc))
		}
	}

	// Cluster descriptions
	clusterDescriptions := m.generateClusterDescriptionsForProvider(provider)
	if len(clusterDescriptions) > 0 {
		b.WriteString("  clusterDescriptions:\n")
		for _, desc := range clusterDescriptions {
			b.WriteString(fmt.Sprintf("    - \"%s\"\n", desc))
		}
	}

	// Machine descriptions
	machineDescriptions := m.generateMachineDescriptionsForProvider(provider)
	if len(machineDescriptions) > 0 {
		b.WriteString("  machineDescriptions:\n")
		for _, desc := range machineDescriptions {
			b.WriteString(fmt.Sprintf("    - \"%s\"\n", desc))
		}
	}

	return ReviewSection{
		Name:    "Included",
		Content: b.String(),
	}
}

// processKubeVirtKubeconfig handles KubeVirt kubeconfig - if it's base64 encoded content,
// decode it and save to a temporary file, returning the file path.
// If it's already a file path, return it as-is.
func (m *Model) processKubeVirtKubeconfig(kubeconfigValue string) (string, error) {
	// Check if the value looks like a file path
	if strings.HasPrefix(kubeconfigValue, "/") || strings.HasPrefix(kubeconfigValue, "./") || strings.HasPrefix(kubeconfigValue, "~/") {
		// It's already a file path
		return kubeconfigValue, nil
	}

	// Try to decode as base64
	decodedData, err := base64.StdEncoding.DecodeString(kubeconfigValue)
	if err != nil {
		// Not base64, might be raw kubeconfig content or short value
		// If it contains typical kubeconfig content, treat it as raw YAML
		if strings.Contains(kubeconfigValue, "apiVersion") || strings.Contains(kubeconfigValue, "clusters:") {
			decodedData = []byte(kubeconfigValue)
		} else {
			// It's a short value, return as-is
			return kubeconfigValue, nil
		}
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "kubevirt-kubeconfig-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	// Write decoded content to file
	if _, err := tmpFile.Write(decodedData); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write kubeconfig to temp file: %w", err)
	}

	return tmpFile.Name(), nil
}

// addProviderCredentialsToSecretsYAML adds provider credentials to the secrets section.
func (m *Model) addProviderCredentialsToSecretsYAML(b *strings.Builder, provider *Provider) {
	switch creds := provider.Credentials.(type) {
	case AWSCredentials:
		b.WriteString(fmt.Sprintf("    accessKeyID: \"%s\"\n", creds.AccessKeyID.Value()))
		b.WriteString(fmt.Sprintf("    secretAccessKey: \"%s\"\n", creds.SecretAccessKey.Value()))
		if creds.AssumeRoleARN.Value() != "" {
			b.WriteString(fmt.Sprintf("    assumeRoleARN: \"%s\"\n", creds.AssumeRoleARN.Value()))
		}
		if creds.AssumeRoleExternalID.Value() != "" {
			b.WriteString(fmt.Sprintf("    assumeRoleExternalID: \"%s\"\n", creds.AssumeRoleExternalID.Value()))
		}
	case AzureCredentials:
		b.WriteString(fmt.Sprintf("    tenantID: \"%s\"\n", creds.TenantID.Value()))
		b.WriteString(fmt.Sprintf("    subscriptionID: \"%s\"\n", creds.SubscriptionID.Value()))
		b.WriteString(fmt.Sprintf("    clientID: \"%s\"\n", creds.ClientID.Value()))
		b.WriteString(fmt.Sprintf("    clientSecret: \"%s\"\n", creds.ClientSecret.Value()))
	case GCPCredentials:
		b.WriteString(fmt.Sprintf("    serviceAccount: \"%s\"\n", creds.ServiceAccount.Value()))
	case AlibabaCredentials:
		b.WriteString(fmt.Sprintf("    accessKeyID: \"%s\"\n", creds.AccessKeyID.Value()))
		b.WriteString(fmt.Sprintf("    accessKeySecret: \"%s\"\n", creds.AccessKeySecret.Value()))
	case AnexiaCredentials:
		b.WriteString(fmt.Sprintf("    token: \"%s\"\n", creds.Token.Value()))
	case DigitalOceanCredentials:
		b.WriteString(fmt.Sprintf("    token: \"%s\"\n", creds.Token.Value()))
	case HetznerCredentials:
		b.WriteString(fmt.Sprintf("    token: \"%s\"\n", creds.Token.Value()))
	case KubeVirtCredentials:
		// Process kubeconfig - if it's base64 encoded, decode and save to temp file
		kubeconfigPath := creds.Kubeconfig.Value()
		if processedPath, err := m.processKubeVirtKubeconfig(kubeconfigPath); err == nil {
			kubeconfigPath = processedPath
		}
		b.WriteString(fmt.Sprintf("    kubeconfigFile: \"%s\"\n", kubeconfigPath))
	case NutanixCredentials:
		b.WriteString(fmt.Sprintf("    username: \"%s\"\n", creds.Username.Value()))
		b.WriteString(fmt.Sprintf("    password: \"%s\"\n", creds.Password.Value()))
		b.WriteString(fmt.Sprintf("    clusterName: \"%s\"\n", creds.ClusterName.Value()))
		if creds.ProxyURL.Value() != "" {
			b.WriteString(fmt.Sprintf("    proxyURL: \"%s\"\n", creds.ProxyURL.Value()))
		}
		if creds.CSIUsername.Value() != "" {
			b.WriteString(fmt.Sprintf("    csiUsername: \"%s\"\n", creds.CSIUsername.Value()))
		}
		if creds.CSIPassword.Value() != "" {
			b.WriteString(fmt.Sprintf("    csiPassword: \"%s\"\n", creds.CSIPassword.Value()))
		}
		if creds.CSIEndpoint.Value() != "" {
			b.WriteString(fmt.Sprintf("    csiEndpoint: \"%s\"\n", creds.CSIEndpoint.Value()))
		}
	case OpenStackCredentials:
		if creds.Username.Value() != "" {
			b.WriteString(fmt.Sprintf("    username: \"%s\"\n", creds.Username.Value()))
		}
		if creds.Password.Value() != "" {
			b.WriteString(fmt.Sprintf("    password: \"%s\"\n", creds.Password.Value()))
		}
		if creds.Project.Value() != "" {
			b.WriteString(fmt.Sprintf("    project: \"%s\"\n", creds.Project.Value()))
		}
		if creds.ProjectID.Value() != "" {
			b.WriteString(fmt.Sprintf("    projectID: \"%s\"\n", creds.ProjectID.Value()))
		}
		if creds.Domain.Value() != "" {
			b.WriteString(fmt.Sprintf("    domain: \"%s\"\n", creds.Domain.Value()))
		}
		if creds.ApplicationCredentialID.Value() != "" {
			b.WriteString(fmt.Sprintf("    applicationCredentialID: \"%s\"\n", creds.ApplicationCredentialID.Value()))
		}
		if creds.ApplicationCredentialSecret.Value() != "" {
			b.WriteString(fmt.Sprintf("    applicationCredentialSecret: \"%s\"\n", creds.ApplicationCredentialSecret.Value()))
		}
		if creds.Token.Value() != "" {
			b.WriteString(fmt.Sprintf("    token: \"%s\"\n", creds.Token.Value()))
		}
	case VSphereCredentials:
		b.WriteString(fmt.Sprintf("    username: \"%s\"\n", creds.Username.Value()))
		b.WriteString(fmt.Sprintf("    password: \"%s\"\n", creds.Password.Value()))
	case VMwareCloudDirectorCredentials:
		if creds.Username.Value() != "" {
			b.WriteString(fmt.Sprintf("    username: \"%s\"\n", creds.Username.Value()))
		}
		if creds.Password.Value() != "" {
			b.WriteString(fmt.Sprintf("    password: \"%s\"\n", creds.Password.Value()))
		}
		if creds.APIToken.Value() != "" {
			b.WriteString(fmt.Sprintf("    apiToken: \"%s\"\n", creds.APIToken.Value()))
		}
		b.WriteString(fmt.Sprintf("    organization: \"%s\"\n", creds.Organization.Value()))
		b.WriteString(fmt.Sprintf("    vdc: \"%s\"\n", creds.VDC.Value()))
	}
}

// generateDatacenterDescriptionsForProvider generates descriptions for selected datacenter settings for a specific provider.
func (m *Model) generateDatacenterDescriptionsForProvider(provider Provider) []string {
	var descriptions []string

	// Only process settings for this specific provider
	for _, group := range m.datacenterSettingsSelection.SettingsByProvider[provider.DisplayName] {
		groupKey := fmt.Sprintf("%s:%s", provider.DisplayName, group.Key)

		// Check if this is a boolean flag group (no options)
		if len(group.Options) == 0 {
			// Check if the group itself is selected
			if m.datacenterSettingsSelection.SelectedGroups[groupKey] {
				// group.Name already includes "with" prefix; trim trailing space added by longestCommonPrefixTokens
				descriptions = append(descriptions, strings.TrimRight(group.Name, " "))
			}
		} else {
			// For each selected option in this group, generate both enabled and disabled entries
			for _, option := range group.Options {
				optionKey := fmt.Sprintf("%s:%s", groupKey, option)
				if m.datacenterSettingsSelection.Selected[optionKey] {
					// group.Name already includes "with" prefix and proper spacing
					desc := fmt.Sprintf("%s%s", group.Name, option)
					descriptions = append(descriptions, desc)
				}
			}
		}
	}

	return descriptions
}

// generateClusterDescriptionsForProvider generates descriptions for selected cluster settings for a specific provider.
func (m *Model) generateClusterDescriptionsForProvider(provider Provider) []string {
	var descriptions []string

	// Only process settings for this specific provider
	for _, group := range m.clusterSettingsSelection.SettingsByProvider[provider.DisplayName] {
		groupKey := fmt.Sprintf("%s:%s", provider.DisplayName, group.Key)

		// Check if this is a boolean flag group (no options)
		if len(group.Options) == 0 {
			// Check if the group itself is selected
			if m.clusterSettingsSelection.SelectedGroups[groupKey] {
				// group.Name already includes "with" prefix; trim trailing space added by longestCommonPrefixTokens
				descriptions = append(descriptions, strings.TrimRight(group.Name, " "))
			}
		} else {
			// For each selected option in this group
			for _, option := range group.Options {
				optionKey := fmt.Sprintf("%s:%s", groupKey, option)
				if m.clusterSettingsSelection.Selected[optionKey] {
					// group.Name already includes "with" prefix and proper spacing
					desc := fmt.Sprintf("%s%s", group.Name, option)
					descriptions = append(descriptions, desc)
				}
			}
		}
	}

	return descriptions
}

// generateMachineDescriptionsForProvider generates descriptions for selected machine deployment settings for a specific provider.
func (m *Model) generateMachineDescriptionsForProvider(provider Provider) []string {
	var descriptions []string

	// Only process settings for this specific provider
	for _, group := range m.machineDeploymentSettingsSelection.SettingsByProvider[provider.DisplayName] {
		groupKey := fmt.Sprintf("%s:%s", provider.DisplayName, group.Key)

		// Check if this is a boolean flag group (no options)
		if len(group.Options) == 0 {
			// Check if the group itself is selected
			if m.machineDeploymentSettingsSelection.SelectedGroups[groupKey] {
				// group.Name already includes "with" prefix; trim trailing space added by longestCommonPrefixTokens
				descriptions = append(descriptions, strings.TrimRight(group.Name, " "))
			}
		} else {
			// For each selected option in this group
			for _, option := range group.Options {
				optionKey := fmt.Sprintf("%s:%s", groupKey, option)
				if m.machineDeploymentSettingsSelection.Selected[optionKey] {
					// group.Name already includes "with" prefix and proper spacing
					desc := fmt.Sprintf("%s%s", group.Name, option)
					descriptions = append(descriptions, desc)
				}
			}
		}
	}

	return descriptions
}

// generateCompleteYAMLForProvider generates the complete YAML configuration for a provider in sample.yaml format.
func (m *Model) generateCompleteYAMLForProvider(provider Provider) string {
	var b strings.Builder

	// Get configuration values
	var (
		namePrefix           = "conformance-tester"
		resultsFile          = "results.json"
		reportsRoot          = "_reports"
		logDirectory         = "_logs"
		retryFailedScenarios = false
		keepFailedClusters   = true
		nodeCount            = 2
	)

	// Extract values from cluster configuration
	for _, category := range m.clusterConfiguration.Categories {
		for _, setting := range category.Settings {
			switch setting.Name {
			case "Name Prefix":
				if val, ok := setting.Value.(string); ok && val != "" {
					namePrefix = val
				}
			case "Results File":
				if val, ok := setting.Value.(string); ok && val != "" {
					resultsFile = val
				}
			case "Reports Directory":
				if val, ok := setting.Value.(string); ok && val != "" {
					reportsRoot = val
				}
			case "Log Directory":
				if val, ok := setting.Value.(string); ok && val != "" {
					logDirectory = val
				}
			case "Node Count":
				if val, ok := setting.Value.(int); ok {
					nodeCount = val
				}
			}
		}
	}

	// Header
	b.WriteString(fmt.Sprintf("namePrefix: \"%s\"\n\n", namePrefix))
	b.WriteString("# The file to write the test results to.\n")
	b.WriteString(fmt.Sprintf("resultsFile: \"%s\"\n\n", resultsFile))
	b.WriteString("# If set, only scenarios that failed in a previous run will be executed.\n")
	b.WriteString(fmt.Sprintf("retryFailedScenarios: %v\n\n", retryFailedScenarios))
	b.WriteString("# Cluster settings\n")
	b.WriteString(fmt.Sprintf("keepFailedClustersAfterTests: %v\n", keepFailedClusters))
	b.WriteString(fmt.Sprintf("nodeCount: %d\n\n", nodeCount))
	b.WriteString("# Paths\n")
	b.WriteString(fmt.Sprintf("reportsRoot: \"%s\"\n", reportsRoot))
	b.WriteString(fmt.Sprintf("logDirectory: \"%s\"\n\n", logDirectory))

	// Environment
	b.WriteString("# Kubermatic settings\n")
	if m.existingEnv.Selected {
		if m.existingEnv.SelectedSeedIndex >= 0 && m.existingEnv.SelectedSeedIndex < len(m.existingEnv.AvailableSeeds) {
			b.WriteString("kubermaticNamespace: \"kubermatic\"\n")
			b.WriteString(fmt.Sprintf("kubermaticSeedName: \"%s\"\n",
				m.existingEnv.AvailableSeeds[m.existingEnv.SelectedSeedIndex]))
		} else {
			b.WriteString("kubermaticNamespace: \"kubermatic\"\n")
			b.WriteString("kubermaticSeedName: \"\"\n")
		}
		if m.existingEnv.ProjectName.Value() != "" {
			b.WriteString(fmt.Sprintf("kubermaticProject: \"%s\"\n\n",
				m.existingEnv.ProjectName.Value()))
		} else {
			b.WriteString("kubermaticProject: \"\"\n\n")
		}
	} else {
		b.WriteString("kubermaticNamespace: \"kubermatic\"\n")
		b.WriteString("kubermaticSeedName: \"\"\n")
		b.WriteString("kubermaticProject: \"\"\n\n")
	}

	// Secrets
	b.WriteString("secrets:\n")
	b.WriteString(fmt.Sprintf("  %s:\n", strings.ToLower(provider.DisplayName)))

	// Add credentials
	if provider.HasPresetCredentials && provider.CredentialSource == CredentialSourcePreset {
		// Credentials come from preset - use PresetCredentials
		if provider.PresetCredentials != nil {
			// Create a temporary provider with Credentials set to PresetCredentials
			tempProvider := Provider{
				DisplayName: provider.DisplayName,
				Credentials: provider.PresetCredentials,
			}
			m.addProviderCredentialsToSecretsYAML(&b, &tempProvider)
		}
	} else {
		m.addProviderCredentialsToSecretsYAML(&b, &provider)
	}

	b.WriteString("\n")

	// Releases
	selectedReleases := []string{}
	for major, selected := range m.releaseSelection.SelectedMajor {
		if selected {
			for _, minor := range m.releaseSelection.MinorVersions[major] {
				if m.releaseSelection.SelectedMinor[minor] {
					selectedReleases = append(selectedReleases, minor)
				}
			}
		}
	}
	if len(selectedReleases) == 0 {
		selectedReleases = []string{latestDefaultKubernetesVersion()}
	}

	b.WriteString("# A list of Kubernetes releases to test.\n")
	b.WriteString("releases:\n")
	for _, release := range selectedReleases {
		b.WriteString(fmt.Sprintf("- \"%s\"\n", release))
	}
	b.WriteString("\n")

	// Distributions
	selectedDistributions := []string{}
	for _, dist := range m.distributionSelection.DistributionsByProvider[provider.DisplayName] {
		key := fmt.Sprintf("%s:%s", provider.DisplayName, dist)
		if m.distributionSelection.Selected[key] {
			selectedDistributions = append(selectedDistributions, string(dist))
		}
	}
	if len(selectedDistributions) == 0 {
		selectedDistributions = []string{"ubuntu"}
	}

	b.WriteString("# A list of enabled operating system distributions.\n")
	b.WriteString("enableDistributions:\n")
	for _, dist := range selectedDistributions {
		b.WriteString(fmt.Sprintf("- %s\n", dist))
	}
	b.WriteString("\n")

	// Resources
	var (
		cpuResources      []string
		memoryResources   []string
		diskSizeResources []string
	)

	for _, category := range m.clusterConfiguration.Categories {
		for _, setting := range category.Settings {
			switch setting.Name {
			case "CPU Cores":
				if val, ok := setting.Value.([]int); ok {
					for _, cpu := range val {
						cpuResources = append(cpuResources, fmt.Sprintf("%d", cpu))
					}
				}
			case "Memory":
				if val, ok := setting.Value.([]string); ok {
					memoryResources = val
				}
			case "Disk Size":
				if val, ok := setting.Value.([]string); ok {
					diskSizeResources = val
				}
			}
		}
	}

	if len(cpuResources) == 0 {
		cpuResources = []string{"2"}
	}
	if len(memoryResources) == 0 {
		memoryResources = []string{"4Gi"}
	}
	if len(diskSizeResources) == 0 {
		diskSizeResources = []string{"25Gi"}
	}

	b.WriteString("resources:\n")
	b.WriteString("  cpu: [")
	for i, cpu := range cpuResources {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(cpu)
	}
	b.WriteString("]\n")

	b.WriteString("  memory: [")
	for i, mem := range memoryResources {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("\"%s\"", mem))
	}
	b.WriteString("]\n")

	b.WriteString("  diskSize: [")
	for i, disk := range diskSizeResources {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("\"%s\"", disk))
	}
	b.WriteString("]\n\n")

	// Included
	b.WriteString("included:\n")

	datacenterDescriptions := m.generateDatacenterDescriptionsForProvider(provider)
	if len(datacenterDescriptions) > 0 {
		b.WriteString("  datacenterDescriptions:\n")
		for _, desc := range datacenterDescriptions {
			b.WriteString(fmt.Sprintf("    - \"%s\"\n", desc))
		}
	}

	clusterDescriptions := m.generateClusterDescriptionsForProvider(provider)
	if len(clusterDescriptions) > 0 {
		b.WriteString("  clusterDescriptions:\n")
		for _, desc := range clusterDescriptions {
			b.WriteString(fmt.Sprintf("    - \"%s\"\n", desc))
		}
	}

	machineDescriptions := m.generateMachineDescriptionsForProvider(provider)
	if len(machineDescriptions) > 0 {
		b.WriteString("  machineDescriptions:\n")
		for _, desc := range machineDescriptions {
			b.WriteString(fmt.Sprintf("    - \"%s\"\n", desc))
		}
	}

	return b.String()
}

// saveProviderConfigurations saves YAML configuration files for all selected providers.
func (m *Model) saveProviderConfigurations() error {
	for _, provider := range m.providers {
		if !provider.Selected {
			continue
		}

		// Generate YAML content
		yamlContent := m.generateCompleteYAMLForProvider(provider)

		// Create filename (lowercase provider name + .yaml)
		filename := strings.ToLower(provider.DisplayName) + ".yaml"
		filename = strings.ReplaceAll(filename, " ", "-") // Replace spaces with dashes

		// Save to file in current directory
		if err := os.WriteFile(filename, []byte(yamlContent), 0644); err != nil {
			return fmt.Errorf("failed to save %s: %w", filename, err)
		}
	}

	return nil
}
