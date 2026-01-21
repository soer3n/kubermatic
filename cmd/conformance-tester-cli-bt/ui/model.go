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
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/kubevirt"
	ginkgoutils "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/utils"
	"k8c.io/kubermatic/v2/pkg/defaulting"
	"k8c.io/machine-controller/sdk/providerconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// Model holds the state of the app.
type Model struct {
	stage int

	localEnv              EnvironmentLocal
	existingEnv           EnvironmentExisting
	environmentFocusIndex int // 0 for Local Environment, 1 for Existing Cluster
	environmentFieldIndex int // Index of the field within the selected environment

	releaseSelection ReleaseSelection

	providers          []Provider
	providerFocusIndex int // Index of currently focused provider
	providerFieldIndex int // Index of the field within the selected provider

	distributionSelection DistributionSelection

	datacenterSettingsSelection DatacenterSettingsSelection

	clusterSettingsSelection ClusterSettingsSelection

	machineDeploymentSettingsSelection MachineDeploymentSettingsSelection

	clusterConfiguration ClusterConfigurationSettings

	Review  Review
	cmdChan <-chan tea.Msg

	// Execution state
	logs           []string
	executionError string
	executionDone  bool

	// Quit confirmation modal state
	quitConfirmVisible bool
	quitConfirmIndex   int

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
func initializeDatacenterSettingsSelection(providers []string) DatacenterSettingsSelection {
	settingsByProvider := make(map[string][]SettingGroup)

	// Gather settings from all selected providers
	for _, provider := range providers {
		var descriptionsMap map[string]k8cginkgo.Description

		// Provider-specific datacenter settings retrieval
		switch strings.ToLower(provider) {
		case "kubevirt":
			descriptionsMap = kubevirt.GetDatacenterDescriptions()
		// Add more providers here as they become available
		// case "aws":
		// 	descriptionsMap = aws.GetDatacenterDescriptions()
		default:
			descriptionsMap = make(map[string]k8cginkgo.Description)
		}

		// Convert map to SettingGroup slice
		var groups []SettingGroup
		for key, desc := range descriptionsMap {
			groups = append(groups, SettingGroup{
				Key:        key,
				Name:       desc.Name,
				Options:    desc.Options,
				IsExpanded: true, // Always show options
			})
		}

		// Sort groups by key for consistent display
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].Key < groups[j].Key
		})

		settingsByProvider[provider] = groups
	}

	// Initialize all providers as expanded by default
	expandedProviders := make(map[string]bool)
	for _, provider := range providers {
		expandedProviders[provider] = true
	}

	return DatacenterSettingsSelection{
		Providers:          providers,
		SettingsByProvider: settingsByProvider,
		Selected:           make(map[string]bool),
		SelectedGroups:     make(map[string]bool),
		FocusedIndex:       0,
		ExpandedProviders:  expandedProviders,
	}
}

// initializeMachineDeploymentSettingsSelection creates the machine deployment settings selection structure.
func initializeMachineDeploymentSettingsSelection(providers []string) MachineDeploymentSettingsSelection {
	settingsByProvider := make(map[string][]SettingGroup)

	// Gather settings from all selected providers
	for _, provider := range providers {
		var descriptionsMap map[string]k8cginkgo.Description

		// Provider-specific machine deployment settings retrieval
		switch strings.ToLower(provider) {
		case "kubevirt":
			descriptionsMap = kubevirt.GetMachineDescriptions()
		// Add more providers here as they become available
		// case "aws":
		// 	descriptionsMap = aws.GetMachineDescriptions()
		default:
			descriptionsMap = make(map[string]k8cginkgo.Description)
		}

		// Convert map to SettingGroup slice
		var groups []SettingGroup
		for key, desc := range descriptionsMap {
			groups = append(groups, SettingGroup{
				Key:        key,
				Name:       desc.Name,
				Options:    desc.Options,
				IsExpanded: true, // Always show options
			})
		}

		// Sort groups by key for consistent display
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].Key < groups[j].Key
		})

		settingsByProvider[provider] = groups
	}

	// Initialize all providers as expanded by default
	expandedProviders := make(map[string]bool)
	for _, provider := range providers {
		expandedProviders[provider] = true
	}

	return MachineDeploymentSettingsSelection{
		Providers:          providers,
		SettingsByProvider: settingsByProvider,
		Selected:           make(map[string]bool),
		SelectedGroups:     make(map[string]bool),
		FocusedIndex:       0,
		ExpandedProviders:  expandedProviders,
	}
}

// initializeClusterSettingsSelection creates the cluster settings selection structure.
func initializeClusterSettingsSelection(providers []string) ClusterSettingsSelection {
	descriptionsMap := ginkgoutils.GetClusterDescriptions()

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
		"Cluster Naming":      true,
		"Machine Deployment":  true,
		"Resource Allocation": true,
		"Test Options":        true,
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
			Selected:    true, // Default selection
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

	// If no env variable, select first available option
	if len(options) > 0 && options[0].Type != "env" {
		options[0].Selected = true
	}

	return options
}

// getSelectedKubeconfigPath returns the currently selected kubeconfig path.
func (m Model) getSelectedKubeconfigPath() string {
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
			ProjectName:          newTextInput("e.g., my-project", 64),
			Errors:               EnvironmentExistingErrors{Fields: make(map[string]string)},
		},
		environmentFocusIndex: 0,
		environmentFieldIndex: 0,
		releaseSelection:      initializeReleaseSelection(),
		providers:             initializeProviders(),
		distributionSelection: initializeDistributionSelection([]string{}), // Will be reinitialized after provider selection
		providerFocusIndex:    0,
		providerFieldIndex:    0,
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

// fetchSeedsAndPresets fetches Seeds and Presets from the Kubernetes cluster.
func (m *Model) fetchSeedsAndPresets() tea.Cmd {
	return func() tea.Msg {
		// Get the selected kubeconfig path
		kubeconfigPath := ""
		optionIndex := m.getKubeconfigOptionIndexFromVisualIndex(m.existingEnv.KubeconfigFocusedIndex)
		if optionIndex >= 0 && optionIndex < len(m.existingEnv.KubeconfigOptions) {
			selectedOption := m.existingEnv.KubeconfigOptions[optionIndex]
			if selectedOption.Type == "custom" {
				kubeconfigPath = m.existingEnv.CustomKubeconfigPath.Value()
			} else {
				kubeconfigPath = selectedOption.Path
			}
		}

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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit confirmation handler
		if m.quitConfirmVisible {
			if handled, cmd := m.handleQuitConfirmation(msg); handled {
				return m, cmd
			}
		}
		switch msg.String() {
		case keyControlC:
			// Show quit confirmation modal
			if m.stage != stageDone {
				m.quitConfirmVisible = true
				m.quitConfirmIndex = 0 // Default to "No"
			}
			return m, nil
		case keyQuit:
			// Immediate quit if we're at the done stage
			if m.stage == stageDone {
				return m, tea.Quit
			}
		}

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
		}

	case tea.WindowSizeMsg:
		cmd = m.handleWindowSize(msg)
	case startMsg:
		cmd = m.handleStart(msg)
	case logMsg:
		cmd = m.handleLog(msg)
	// case errMsg:
	// 	cmd = m.handleError(msg)
	case doneMsg:
		cmd = m.handleDone(msg)
	case execOutputMsg:
		cmd = m.handleExecOutput(msg)
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
	}

	return m, cmd
}

// getUIWidth returns the dynamic UI width based on terminal size.
// Falls back to 150 if terminal width hasn't been detected yet.
func (m Model) getUIWidth() int {
	if m.terminalWidth == 0 {
		return 150 // Default width
	}
	// Use 90% of terminal width, with minimum of 80 and maximum of 200
	width := int(float64(m.terminalWidth) * 0.9)
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

func ConformanceTester() (tea.Model, error) {
	m, err := tea.NewProgram(initialModel()).Run()
	if err != nil {
		return nil, err
	}
	// if myModel, ok := m.(Model); ok {
	// 	return myModel.Nodes.Configs, nil
	// }
	return m, nil
}

// --- View Entry Point ---
// View renders the entire UI based on the current application stage.
func (m Model) View() string {
	// Get dynamic UI dimensions
	uiWidth := m.getUIWidth()
	uiInnerWidth := m.getUIInnerWidth()

	helpText := helpBar(m.stage)
	helpContent := styleHelpBar.Width(uiInnerWidth).Render(helpText)
	helpWithBorder := styleHelpBarBorder.Width(uiInnerWidth).Render("") + "\n" + helpContent

	var content string
	switch m.stage {
	case stageWelcome:
		content = m.renderWelcome(helpWithBorder, uiWidth, uiInnerWidth)
	case stageEnvironmentSelection:
		content = m.renderEnvironmentSelection(helpWithBorder, uiWidth, uiInnerWidth)
	case stageReleaseSelection:
		content = m.renderReleaseSelection(helpWithBorder, uiWidth, uiInnerWidth)
	case stageProviderSelection:
		content = m.renderProviderSelection(helpWithBorder, uiWidth, uiInnerWidth)
	case stageDistributionSelection:
		content = m.renderDistributionSelection(helpWithBorder, uiWidth, uiInnerWidth)
	case stageDatacenterSettingsSelection:
		content = m.renderDatacenterSettingsSelection(helpWithBorder, uiWidth, uiInnerWidth)
	case stageClusterSettingsSelection:
		content = m.renderClusterSettingsSelection(helpWithBorder, uiWidth, uiInnerWidth)
	case stageMachineDeploymentSettingsSelection:
		content = m.renderMachineDeploymentSettingsSelection(helpWithBorder, uiWidth, uiInnerWidth)
	case stageClusterConfiguration:
		content = m.renderClusterConfiguration(helpWithBorder, uiWidth, uiInnerWidth)

	// case stageReview:
	// 	content = m.renderReview(helpWithBorder)
	case stageExecuting:
		content = m.renderExecuting(helpWithBorder, uiWidth, uiInnerWidth)
	case stageDone:
		content = m.renderDone(helpWithBorder, uiWidth, uiInnerWidth)
	default:
		// Render nothing for unknown stages
		os.Exit(0)
		return ""
	}

	// Combine banner and content, then center the entire layout
	bannerContent := styleBanner.Width(uiWidth).Render(bannerText())
	finalContent := lipgloss.JoinVertical(lipgloss.Center, bannerContent, content)

	// The outer width includes the box padding (4 chars left + 4 chars right)
	base := lipgloss.PlaceHorizontal(uiWidth+8, lipgloss.Center, finalContent)

	if m.quitConfirmVisible {
		// Show only the modal centered, on top of everything
		modal := m.renderQuitConfirm(uiWidth, uiInnerWidth)
		bannerContent := styleBanner.Width(uiWidth).Render(bannerText())

		return lipgloss.Place(uiWidth+8, 0, lipgloss.Center, lipgloss.Center, bannerContent+"\n"+modal)
	}

	return base
}
