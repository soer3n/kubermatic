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
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"k8c.io/kubermatic/v2/pkg/defaulting"
	"k8c.io/machine-controller/sdk/providerconfig"
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

	offline bool
	MetalLB MetalLB
	// CSI toggle
	CSIEnabled bool
	// Network config
	Network NetworkConfig

	ContainerRegistry OCIConfiguration
	HelmRegistry      OCIConfiguration
	PackageRepo       PackageRepository

	NodeCount NodeCount

	Nodes Nodes

	Review  Review
	cmdChan <-chan tea.Msg

	// Execution state
	logs           []string
	executionError string
	executionDone  bool

	// Quit confirmation modal state
	quitConfirmVisible bool
	quitConfirmIndex   int
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

// initializeDistributionSelection creates the distribution selection structure.
func initializeDistributionSelection() DistributionSelection {
	distributions := providerconfig.AllOperatingSystems

	// Create display names map
	displayNames := map[providerconfig.OperatingSystem]string{
		providerconfig.OperatingSystemUbuntu:       "Ubuntu",
		providerconfig.OperatingSystemAmazonLinux2: "Amazon Linux 2",
		providerconfig.OperatingSystemRHEL:         "Red Hat Enterprise Linux (RHEL)",
		providerconfig.OperatingSystemFlatcar:      "Flatcar Container Linux",
		providerconfig.OperatingSystemRockyLinux:   "Rocky Linux",
	}

	return DistributionSelection{
		Distributions:     distributions,
		DistributionNames: displayNames,
		Selected:          make(map[providerconfig.OperatingSystem]bool),
		FocusedIndex:      0,
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

func initialModel(offline bool) Model {
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
			Selected:       false,
			CurrentField:   0,
			KubeconfigPath: newTextInput("e.g., ~/.kube/config", 256),
			SeedName:       newTextInput("e.g., seed-1", 64),
			PresetName:     newTextInput("e.g., my-preset", 64),
			ProjectName:    newTextInput("e.g., my-project", 64),
			Errors:         EnvironmentExistingErrors{},
		},
		environmentFocusIndex: 0,
		environmentFieldIndex: 0,
		releaseSelection:      initializeReleaseSelection(),
		providers:             initializeProviders(), distributionSelection: initializeDistributionSelection(), providerFocusIndex: 0,
		providerFieldIndex: 0,
		// MetalLB: MetalLB{
		// 	Enabled: false,
		// 	Input:   newTextInput("e.g., 192.168.1.100-192.168.1.150", 50),
		// },
		// CSIEnabled: false,
		// Network: NetworkConfig{
		// 	CIDR:         newTextInput("e.g., 10.244.0.0/16", 18),
		// 	DNSServer:    newTextInput("e.g., 8.8.8.8", 15),
		// 	GatewayIP:    newTextInput("e.g., 192.168.1.1", 15),
		// 	CurrentField: 0,
		// 	Errors:       NetworkErrors{},
		// },
		// NodeCount: NodeCount{
		// 	NodeCountInput:         newTextInput("e.g., 10", 4),
		// 	ControlPlaneCountInput: newTextInput("e.g., 3", 4),
		// 	APIEndpointInput:       newTextInput("e.g., dns1.example.com,dns2.example.com", 256),
		// 	CurrentField:           0,
		// 	Max:                    1000,
		// },
		// Nodes: Nodes{
		// 	Configs: []NodeConfig{{}},
		// 	Inputs: []NodeInputFields{{
		// 		Address:    newTextInput("Address", 64),
		// 		Username:   newTextInput("Username", 32),
		// 		SSHKeyPath: newTextInput("SSH Key Path", 256),
		// 	}},
		// 	Current:      0,
		// 	CurrentField: 0,
		// },
	}

	// Focus the first input field
	// model.NodeCount.NodeCountInput.Focus()

	if offline {
		model.ContainerRegistry = OCIConfiguration{
			Endpoint:     newTextInput("Container Registry Endpoint", 256),
			Insecure:     false,
			Username:     newTextInput("Container Registry Username", 32),
			Password:     newTextInput("Container Registry Password", 32),
			CurrentField: 0,
		}

		model.HelmRegistry = OCIConfiguration{
			Endpoint: newTextInput("Helm Registry Endpoint", 256),
			Insecure: false,
			Username: newTextInput("Helm Registry Username", 32),
			Password: newTextInput("Helm Registry Password", 32),
		}

		model.PackageRepo = PackageRepository{
			Enabled: false,
			Address: newTextInput("e.g., http://package-repo.local:8080", 256),
		}

		model.offline = true
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
	}

	return m, cmd
}

func KubeVCluster(offline bool) ([]NodeConfig, error) {
	m, err := tea.NewProgram(initialModel(offline)).Run()
	if err != nil {
		return nil, err
	}
	if myModel, ok := m.(Model); ok {
		return myModel.Nodes.Configs, nil
	}
	return nil, nil
}

// --- View Entry Point ---
// View renders the entire UI based on the current application stage.
func (m Model) View() string {
	helpText := helpBar(m.stage)
	helpContent := styleHelpBar.Width(uiInnerWidth).Render(helpText)
	helpWithBorder := styleHelpBarBorder.Width(uiInnerWidth).Render("") + "\n" + helpContent

	var content string
	switch m.stage {
	case stageWelcome:
		content = m.renderWelcome(helpWithBorder)
	case stageEnvironmentSelection:
		content = m.renderEnvironmentSelection(helpWithBorder)
	case stageReleaseSelection:
		content = m.renderReleaseSelection(helpWithBorder)
	case stageProviderSelection:
		content = m.renderProviderSelection(helpWithBorder)
	case stageDistributionSelection:
		content = m.renderDistributionSelection(helpWithBorder)

	// case stageReview:
	// 	content = m.renderReview(helpWithBorder)
	case stageExecuting:
		content = m.renderExecuting(helpWithBorder)
	case stageDone:
		content = m.renderDone(helpWithBorder)
	default:
		// Render nothing for unknown stages
		return ""
	}

	// Combine banner and content, then center the entire layout
	bannerContent := styleBanner.Width(uiWidth).Render(bannerText())
	finalContent := lipgloss.JoinVertical(lipgloss.Center, bannerContent, content)

	// The outer width includes the box padding (4 chars left + 4 chars right)
	base := lipgloss.PlaceHorizontal(uiWidth+8, lipgloss.Center, finalContent)

	if m.quitConfirmVisible {
		// Show only the modal centered, on top of everything
		modal := m.renderQuitConfirm()
		bannerContent := styleBanner.Width(uiWidth).Render(bannerText())

		return lipgloss.Place(uiWidth+8, 0, lipgloss.Center, lipgloss.Center, bannerContent+"\n"+modal)
	}

	return base
}
