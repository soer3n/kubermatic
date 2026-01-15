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
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// Color definitions for consistent theming.
const (
	colorMainBlue  = "#2196F3"
	colorMainWhite = "#FFFFFF"
	colorErrorRed  = "#FF5252"
)

// Application-wide style configuration.
// Grouping styles improves readability and makes theme changes easier.
var (
	// Core UI element styles.
	styleBanner = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMainBlue)).
			Bold(true).
			Padding(0, 0).
			Align(lipgloss.Center)

	// Style for focused/highlighted UI elements.
	styleFocusHighlight = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorMainWhite)).
				Background(lipgloss.Color(colorMainBlue)).
				Bold(true)
	styleHeader = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMainBlue)).
			Bold(true).
			Padding(0, 1)
	styleBox = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color(colorMainBlue)).
			Padding(1, 4) // Provides internal padding

	styleItem = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMainBlue)).
			Padding(0, 1).
			Width(40)

		// Input and label styles.
	styleInput = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMainBlue)).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorMainBlue)).
			Padding(0, 2).
			Width(40)
	styleLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMainBlue)).
			Bold(true).
			Width(20). // Unified label width for better alignment
			Align(lipgloss.Right)

		// Feedback and help styles.
	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorErrorRed)).
			Bold(true).
			Align(lipgloss.Center).
			Margin(1, 0) // Adds vertical space around errors
	styleHelpBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMainWhite)).
			Background(lipgloss.Color(colorMainBlue)).
			Padding(0, 2)
	styleHelpBarBorder = lipgloss.NewStyle().
				BorderTop(true).
				BorderForeground(lipgloss.Color(colorMainBlue)).
				Padding(1, 0) // Adds space above the help bar

	// Help text style for informational messages.
	styleHelpText = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Italic(true)

	styleButton = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colorMainBlue)).
			Foreground(lipgloss.Color(colorMainBlue)).
			Width(12).
			Align(lipgloss.Center)
	styleSelectedButton = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorMainBlue)).
				Bold(true).
				Border(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color(colorMainBlue)).
				Width(12).
				Align(lipgloss.Center)
)

// Fixed dimensions for consistent UI layout.
const (
	uiWidth        = 120 // Main content width
	uiInnerWidth   = uiWidth - 8
	uiBoxHeightPad = 2 // Adjustment for top/bottom borders in boxStyle
)

// renderQuitConfirm draws the quit confirmation dialog.
func (m Model) renderQuitConfirm() string {
	const boxHeight = 15

	// Dynamic content based on stage
	titleContent := "Confirm Quit"
	warningContent := "Are you sure you want to quit? Unsaved progress will be lost."
	if m.stage == stageExecuting {
		titleContent = "Caution: Interrupting Installation"
		warningContent = executionWarning
	}
	var b strings.Builder

	title := lipgloss.PlaceHorizontal(
		uiWidth,
		lipgloss.Center,
		styleHeader.Render(titleContent),
	)
	b.WriteString(title + "\n\n")

	// Warning text - centered within content area
	warning := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMainWhite)).
		Width(uiInnerWidth).
		Align(lipgloss.Center).
		Render(warningContent)
	b.WriteString(warning + "\n\n")

	// Create buttons with CONSISTENT width and border style
	const btnWidth = 12
	const btnSpacing = 4

	noBtn := styleButton.Width(btnWidth).Render("No")
	yesBtn := styleButton.Width(btnWidth).Render("Yes")
	if m.quitConfirmIndex == 0 {
		noBtn = styleSelectedButton.Width(btnWidth).Render("No")
	} else {
		yesBtn = styleSelectedButton.Width(btnWidth).Render("Yes")
	}

	// Center button group within full width
	btnGroup := lipgloss.JoinHorizontal(
		lipgloss.Center,
		noBtn,
		lipgloss.NewStyle().Width(btnSpacing).Render(" "),
		yesBtn,
	)

	centeredButtons := lipgloss.PlaceHorizontal(
		uiWidth,
		lipgloss.Center,
		btnGroup,
	)
	b.WriteString(centeredButtons)

	// Pad content to maintain consistent height
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")

	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody)
}

// renderWelcome displays the initial welcome screen.
func (m Model) renderWelcome(helpWithBorder string) string {
	const boxHeight = 15
	title := styleHeader.Render(welcomeTitleText)
	disclaimer := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(welcomeDisclaimerText)

	// Build the main content with extra spacing
	var b strings.Builder
	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n")
	b.WriteString(disclaimer + "\n")

	// Pad content to ensure help bar is at the bottom
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")

	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
}

func (m Model) renderEnvironmentSelection(helpWithBorder string) string {
	const boxHeight = 20
	title := styleHeader.Render("Select Deployment Environment")
	description := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(environmentSelectionText)

	var b strings.Builder
	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n")
	b.WriteString(description + "\n\n")

	// Local Environment option
	localCheckbox := "[ ]"
	if m.localEnv.Selected {
		localCheckbox = "[x]"
	}
	localOption := fmt.Sprintf("%s Local Environment", lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(localCheckbox))

	// Highlight if focused on the checkbox
	if m.environmentFocusIndex == 0 && m.environmentFieldIndex == 0 {
		localOption = styleFocusHighlight.Render(localOption)
	}
	b.WriteString(localOption + "\n\n")

	// Show Local Environment fields if selected
	if m.localEnv.Selected {
		localFields := []struct {
			Label    string
			Input    textinput.Model
			Error    string
			FieldIdx int
		}{
			{"Kubermatic Config Path:", m.localEnv.KubermaticConfigurationsPath, m.localEnv.Errors.KubermaticConfigurationsPath, 1},
			{"Helm Values Path:", m.localEnv.HelmValuesPath, m.localEnv.Errors.HelmValuesPath, 2},
			{"MLA Values Path:", m.localEnv.MLAValuesPath, m.localEnv.Errors.MLAValuesPath, 3},
		}

		for _, field := range localFields {
			line := lipgloss.JoinHorizontal(
				lipgloss.Left,
				styleLabel.Render(field.Label),
				" ",
				styleInput.Render(field.Input.View()),
			)
			b.WriteString(line + "\n")

			// Add error message if present
			if field.Error != "" {
				b.WriteString(styleError.Width(uiWidth-4).Render(field.Error) + "\n")
			}
		}
		b.WriteString("\n")
	}

	// Existing Cluster option
	existingCheckbox := "[ ]"
	if m.existingEnv.Selected {
		existingCheckbox = "[x]"
	}
	existingOption := fmt.Sprintf("%s Existing Cluster", lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(existingCheckbox))

	// Highlight if focused on the checkbox
	if m.environmentFocusIndex == 1 && m.environmentFieldIndex == 0 {
		existingOption = styleFocusHighlight.Render(existingOption)
	}
	b.WriteString(existingOption + "\n\n")

	// Show Existing Cluster fields if selected
	if m.existingEnv.Selected {
		existingFields := []struct {
			Label    string
			Input    textinput.Model
			Error    string
			FieldIdx int
		}{
			{"Kubeconfig Path:", m.existingEnv.KubeconfigPath, m.existingEnv.Errors.KubeconfigPath, 1},
			{"Seed Name:", m.existingEnv.SeedName, m.existingEnv.Errors.SeedName, 2},
			{"Preset Name:", m.existingEnv.PresetName, m.existingEnv.Errors.PresetName, 3},
			{"Project Name:", m.existingEnv.ProjectName, m.existingEnv.Errors.ProjectName, 4},
		}

		for _, field := range existingFields {
			line := lipgloss.JoinHorizontal(
				lipgloss.Left,
				styleLabel.Render(field.Label),
				" ",
				styleInput.Render(field.Input.View()),
			)
			b.WriteString(line + "\n")

			// Add error message if present
			if field.Error != "" {
				b.WriteString(styleError.Width(uiWidth-4).Render(field.Error) + "\n")
			}
		}
		b.WriteString("\n")
	}

	// Pad content to ensure help bar is at the bottom
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")
	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
}

func (m Model) renderReleaseSelection(helpWithBorder string) string {
	const boxHeight = 30
	title := styleHeader.Render("Select Kubernetes Release")
	description := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(releaseSelectionText)

	var b strings.Builder
	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n")
	b.WriteString(description + "\n\n")

	// Render hierarchical version list - always show all versions
	for majorIdx, majorVersion := range m.releaseSelection.MajorVersions {
		isMajorSelected := m.releaseSelection.SelectedMajor[majorVersion]
		isFocusedMajor := majorIdx == m.releaseSelection.FocusedMajorIndex && !m.releaseSelection.IsMinorFocused

		// Render major version with checkbox
		majorCheckbox := "[ ]"
		if isMajorSelected {
			majorCheckbox = "[x]"
		}

		majorLine := fmt.Sprintf("%s %s", lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(majorCheckbox), majorVersion)

		// Highlight if this major version is focused
		if isFocusedMajor {
			majorLine = styleFocusHighlight.Render(majorLine)
		} else if isMajorSelected {
			majorLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(majorLine)
		} else {
			majorLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(majorLine)
		}

		b.WriteString(majorLine + "\n")

		// Show minor versions (always expanded)
		minorVersions := m.releaseSelection.MinorVersions[majorVersion]
		for minorIdx, minorVersion := range minorVersions {
			isFocusedMinor := majorIdx == m.releaseSelection.FocusedMajorIndex &&
				minorIdx == m.releaseSelection.FocusedMinorIndex &&
				m.releaseSelection.IsMinorFocused

			isSelected := m.releaseSelection.SelectedMinor[minorVersion]

			// Build minor version line
			minorCheckbox := "[ ]"
			if isSelected {
				minorCheckbox = "[x]"
			}
			minorLine := fmt.Sprintf("  %s %s", lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(minorCheckbox), minorVersion)

			// Apply styling
			if isFocusedMinor {
				minorLine = styleFocusHighlight.Render(minorLine)
			} else if isSelected {
				minorLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(minorLine)
			} else {
				minorLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(minorLine)
			}

			b.WriteString(minorLine + "\n")
		}
	}

	// Pad content to ensure help bar is at the bottom
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")
	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
}

func (m Model) renderProviderSelection(helpWithBorder string) string {
	const boxHeight = 15
	title := styleHeader.Render("Select Infrastructure Provider")
	description := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(providerSelectionText)

	var b strings.Builder
	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n")
	b.WriteString(description + "\n\n")

	// Render provider checkboxes
	for i, provider := range m.providers {
		checkbox := "[ ]"

		if provider.Selected {
			checkbox = "[x]"
		}

		providerOption := fmt.Sprintf("%s %s", lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(checkbox), provider.DisplayName)

		if i == m.providerFocusIndex && m.providerFieldIndex == 0 {
			providerOption = styleFocusHighlight.Render(providerOption)
		}

		b.WriteString(providerOption + "\n")

		// If provider is selected, show credential fields
		if provider.Selected {
			b.WriteString(m.renderProviderCredentials(provider, i) + "\n")
		}
	}

	// Pad content to ensure help bar is at the bottom
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")
	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)

}

// renderProviderCredentials renders the credential fields for a specific provider.
func (m Model) renderProviderCredentials(provider Provider, providerIndex int) string {
	var b strings.Builder
	b.WriteString("\n")

	renderField := func(label string, input textinput.Model, error string, fieldIndex int) {
		if providerIndex == m.providerFocusIndex && fieldIndex == m.providerFieldIndex {
			b.WriteString(styleFocusHighlight.Render("  "+label) + " " + input.View() + "\n")
		} else {
			b.WriteString("  " + label + " " + input.View() + "\n")
		}
		if error != "" {
			b.WriteString(styleError.Width(uiWidth-4).Render("    "+error) + "\n")
		}
	}

	switch creds := provider.Credentials.(type) {
	case AWSCredentials:
		renderField("Access Key ID:", creds.AccessKeyID, provider.Errors.Fields["AccessKeyID"], 1)
		renderField("Secret Access Key:", creds.SecretAccessKey, provider.Errors.Fields["SecretAccessKey"], 2)
		renderField("Assume Role ARN:", creds.AssumeRoleARN, provider.Errors.Fields["AssumeRoleARN"], 3)
		renderField("External ID:", creds.AssumeRoleExternalID, provider.Errors.Fields["AssumeRoleExternalID"], 4)

	case AzureCredentials:
		renderField("Tenant ID:", creds.TenantID, provider.Errors.Fields["TenantID"], 1)
		renderField("Subscription ID:", creds.SubscriptionID, provider.Errors.Fields["SubscriptionID"], 2)
		renderField("Client ID:", creds.ClientID, provider.Errors.Fields["ClientID"], 3)
		renderField("Client Secret:", creds.ClientSecret, provider.Errors.Fields["ClientSecret"], 4)

	case GCPCredentials:
		renderField("Service Account JSON:", creds.ServiceAccount, provider.Errors.Fields["ServiceAccount"], 1)

	case AlibabaCredentials:
		renderField("Access Key ID:", creds.AccessKeyID, provider.Errors.Fields["AccessKeyID"], 1)
		renderField("Access Key Secret:", creds.AccessKeySecret, provider.Errors.Fields["AccessKeySecret"], 2)

	case AnexiaCredentials:
		renderField("API Token:", creds.Token, provider.Errors.Fields["Token"], 1)

	case DigitalOceanCredentials:
		renderField("API Token:", creds.Token, provider.Errors.Fields["Token"], 1)

	case HetznerCredentials:
		renderField("API Token:", creds.Token, provider.Errors.Fields["Token"], 1)

	case KubeVirtCredentials:
		renderField("Kubeconfig Path:", creds.Kubeconfig, provider.Errors.Fields["Kubeconfig"], 1)

	case NutanixCredentials:
		renderField("Username:", creds.Username, provider.Errors.Fields["Username"], 1)
		renderField("Password:", creds.Password, provider.Errors.Fields["Password"], 2)
		renderField("Cluster Name:", creds.ClusterName, provider.Errors.Fields["ClusterName"], 3)
		renderField("Proxy URL:", creds.ProxyURL, provider.Errors.Fields["ProxyURL"], 4)
		renderField("CSI Username:", creds.CSIUsername, provider.Errors.Fields["CSIUsername"], 5)
		renderField("CSI Password:", creds.CSIPassword, provider.Errors.Fields["CSIPassword"], 6)
		renderField("CSI Endpoint:", creds.CSIEndpoint, provider.Errors.Fields["CSIEndpoint"], 7)

	case OpenStackCredentials:
		renderField("Username:", creds.Username, provider.Errors.Fields["Username"], 1)
		renderField("Password:", creds.Password, provider.Errors.Fields["Password"], 2)
		renderField("Project:", creds.Project, provider.Errors.Fields["Project"], 3)
		renderField("Project ID:", creds.ProjectID, provider.Errors.Fields["ProjectID"], 4)
		renderField("Domain:", creds.Domain, provider.Errors.Fields["Domain"], 5)
		renderField("App Credential ID:", creds.ApplicationCredentialID, provider.Errors.Fields["ApplicationCredentialID"], 6)
		renderField("App Credential Secret:", creds.ApplicationCredentialSecret, provider.Errors.Fields["ApplicationCredentialSecret"], 7)
		renderField("Token:", creds.Token, provider.Errors.Fields["Token"], 8)

	case VSphereCredentials:
		renderField("Username:", creds.Username, provider.Errors.Fields["Username"], 1)
		renderField("Password:", creds.Password, provider.Errors.Fields["Password"], 2)

	case VMwareCloudDirectorCredentials:
		renderField("Username:", creds.Username, provider.Errors.Fields["Username"], 1)
		renderField("Password:", creds.Password, provider.Errors.Fields["Password"], 2)
		renderField("API Token:", creds.APIToken, provider.Errors.Fields["APIToken"], 3)
		renderField("Organization:", creds.Organization, provider.Errors.Fields["Organization"], 4)
		renderField("VDC:", creds.VDC, provider.Errors.Fields["VDC"], 5)
	}

	return b.String()
}

func (m Model) renderDistributionSelection(helpWithBorder string) string {
	const boxHeight = 20
	title := styleHeader.Render("Select Operating System Distributions")
	description := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(distributionSelectionText)

	var b strings.Builder
	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n")
	b.WriteString(description + "\n\n")

	// Render distribution checkboxes
	for i, distribution := range m.distributionSelection.Distributions {
		checkbox := "[ ]"
		if m.distributionSelection.Selected[distribution] {
			checkbox = "[x]"
		}

		displayName := m.distributionSelection.DistributionNames[distribution]
		distOption := fmt.Sprintf("%s %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(checkbox),
			displayName)

		// Highlight if focused
		if i == m.distributionSelection.FocusedIndex {
			distOption = styleFocusHighlight.Render(distOption)
		} else if m.distributionSelection.Selected[distribution] {
			distOption = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(distOption)
		} else {
			distOption = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(distOption)
		}

		b.WriteString(distOption + "\n")
	}

	// pad content to ensure help bar is at the bottom
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")
	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
}

// // renderNetworkConfig displays fields for network configuration.
// func (m Model) renderNetworkConfig(helpWithBorder string) string {
// 	const boxHeight = 22 // Increased height slightly for better spacing
// 	title := styleHeader.Render("Network configuration for Kubermatic virtualization stack")
// 	description := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(networkConfigDescription)

// 	// Build the main content with extra spacing
// 	var b strings.Builder
// 	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n") // Extra line after title
// 	b.WriteString(description + "\n\n")                                               // Extra line after description

// 	// Add form fields with consistent alignment and spacing
// 	fields := []struct {
// 		Label string
// 		Input textinput.Model
// 		Error string
// 	}{
// 		{"Network (CIDR):", m.Network.CIDR, m.Network.Errors.CIDR},
// 		{"DNS Server:", m.Network.DNSServer, m.Network.Errors.DNSServer},
// 		{"Gateway IP:", m.Network.GatewayIP, m.Network.Errors.GatewayIP},
// 	}

// 	for _, field := range fields {
// 		inputLine := lipgloss.JoinHorizontal(
// 			lipgloss.Center,
// 			styleLabel.Render(field.Label),
// 			" ", // Space between label and input
// 			styleInput.Render(field.Input.View()),
// 		)
// 		b.WriteString(inputLine + "\n")
// 		// Add error message or blank line for spacing
// 		if field.Error != "" {
// 			b.WriteString(" " + styleError.Width(uiWidth-4).Render(field.Error) + "\n\n") // Extra line after error
// 		} else {
// 			b.WriteString("\n") // Blank line if no error
// 		}
// 	}

// 	// Pad content to ensure help bar is at the bottom
// 	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
// 	contentBody := strings.Join(lines, "\n")

// 	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
// }

// // renderMetalLB displays the MetalLB configuration toggle and input.
// func (m Model) renderMetalLB(helpWithBorder string) string {
// 	const boxHeight = 22 // Increased height for better spacing
// 	title := styleHeader.Render("LoadBalancer service for Kubermatic virtualization")
// 	disclaimerLabel := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Bold(true).Render("Disclaimer:")
// 	disclaimerText := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(metalLBDisclaimerText)
// 	checkboxState := "[ ]"
// 	if m.MetalLB.Enabled {
// 		checkboxState = "[x]"
// 	}
// 	checkbox := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(checkboxState)
// 	toggleLabel := "Enable LoadBalancer service (MetalLB)"
// 	rangeLabel := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Width(uiWidth - 20).Align(lipgloss.Left).Render("Define the LoadBalancer IP range:")

// 	// Build the main content with extra spacing
// 	var b strings.Builder
// 	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n") // Extra line after title
// 	b.WriteString(disclaimerLabel + " " + disclaimerText + "\n\n")                    // Extra line after disclaimer

// 	b.WriteString(fmt.Sprintf("%s %s\n\n", checkbox, toggleLabel)) // Extra line after toggle

// 	// Conditionally render input and error if enabled
// 	if m.MetalLB.Enabled {
// 		b.WriteString(rangeLabel + "\n")
// 		b.WriteString(styleInput.Render(m.MetalLB.Input.View()) + "\n")
// 		if m.MetalLB.Error != "" {
// 			b.WriteString("\n" + styleError.Width(uiWidth).Render(m.MetalLB.Error) + "\n\n") // Extra lines around error
// 		} else {
// 			b.WriteString("\n\n") // Extra blank lines if no error
// 		}
// 	} else {
// 		// Reserve space for input and potential error when disabled, with spacing
// 		b.WriteString("\n\n")
// 	}

// 	// Pad content to ensure help bar is at the bottom
// 	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
// 	contentBody := strings.Join(lines, "\n")

// 	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
// }

// // renderContainerRegistry displays the container registry configuration for offline installations.
// func (m Model) renderContainerRegistry(helpWithBorder string) string {
// 	const boxHeight = 22

// 	title := styleHeader.Render("Container Registry Configuration")
// 	description := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(containerRegistryText)

// 	var b strings.Builder
// 	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n")
// 	b.WriteString(description + "\n\n")

// 	// Only show the form in offline mode
// 	if m.offline {
// 		fields := []struct {
// 			Label string
// 			Input textinput.Model
// 		}{
// 			{"Registry endpoint:", m.ContainerRegistry.Endpoint},
// 			{"Username (optional):", m.ContainerRegistry.Username},
// 			{"Password (optional):", m.ContainerRegistry.Password},
// 		}

// 		for _, field := range fields {
// 			line := lipgloss.JoinHorizontal(
// 				lipgloss.Center,
// 				styleLabel.Render(field.Label),
// 				" ",
// 				styleInput.Render(field.Input.View()),
// 			)
// 			b.WriteString(line + "\n\n")
// 		}

// 		// Insecure registry toggle
// 		insecureState := "[ ]"
// 		if m.ContainerRegistry.Insecure {
// 			insecureState = "[x]"
// 		}
// 		insecureLine := fmt.Sprintf("%s Allow insecure connections", insecureState)
// 		insecureContent := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(insecureLine)
// 		if m.ContainerRegistry.CurrentField == 3 { // 3 is the index for the insecure toggle
// 			insecureContent = styleFocusHighlight.Render(insecureLine)
// 		}
// 		b.WriteString(lipgloss.JoinHorizontal(
// 			lipgloss.Center,
// 			styleLabel.Render("Insecure pull:"),
// 			insecureContent,
// 		) + "\n")

// 		// Show error if any
// 		if m.ContainerRegistry.Error != "" {
// 			b.WriteString("\n" + styleError.Width(uiWidth).Render(m.ContainerRegistry.Error) + "\n")
// 		}
// 	} else {
// 		// In online mode, just show a message
// 		b.WriteString("Container registry configuration is not required for online installations.\n\n")
// 	}

// 	// Pad content to ensure help bar is at the bottom
// 	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
// 	contentBody := strings.Join(lines, "\n")

// 	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
// }

// // renderHelmRegistry displays the Helm registry configuration for offline installations.
// func (m Model) renderHelmRegistry(helpWithBorder string) string {
// 	const boxHeight = 22

// 	title := styleHeader.Render("Helm Registry Configuration")
// 	description := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(helmRegistryText)

// 	var b strings.Builder
// 	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n")
// 	b.WriteString(description + "\n\n")

// 	// Only show the form in offline mode
// 	if m.offline {
// 		fields := []struct {
// 			Label string
// 			Input textinput.Model
// 		}{
// 			{"Helm registry endpoint:", m.HelmRegistry.Endpoint},
// 			{"Username (optional):", m.HelmRegistry.Username},
// 			{"Password (optional):", m.HelmRegistry.Password},
// 		}

// 		for _, field := range fields {
// 			line := lipgloss.JoinHorizontal(
// 				lipgloss.Center,
// 				styleLabel.Render(field.Label),
// 				" ",
// 				styleInput.Render(field.Input.View()),
// 			)
// 			b.WriteString(line + "\n\n")
// 		}

// 		// Insecure registry toggle
// 		insecureState := "[ ]"
// 		if m.HelmRegistry.Insecure {
// 			insecureState = "[x]"
// 		}
// 		insecureLine := fmt.Sprintf("%s Allow insecure connections", insecureState)
// 		insecureContent := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(insecureLine)
// 		if m.HelmRegistry.CurrentField == 3 { // 3 is the index for the insecure toggle
// 			insecureContent = styleFocusHighlight.Render(insecureLine)
// 		}
// 		b.WriteString(lipgloss.JoinHorizontal(
// 			lipgloss.Center,
// 			styleLabel.Render("Insecure pull:"),
// 			insecureContent,
// 		) + "\n")

// 		// Show error if any
// 		if m.HelmRegistry.Error != "" {
// 			b.WriteString("\n" + styleError.Width(uiWidth).Render(m.HelmRegistry.Error) + "\n")
// 		}
// 	} else {
// 		// In online mode, just show a message
// 		b.WriteString("Helm registry configuration is not required for online installations.\n\n")
// 	}

// 	// Pad the content to maintain consistent height
// 	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
// 	contentBody := strings.Join(lines, "\n")

// 	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
// }

// // renderPackageRepository displays the package repository configuration for offline installations.
// func (m Model) renderPackageRepository(helpWithBorder string) string {
// 	const boxHeight = 22

// 	title := styleHeader.Render("Package Repository Configuration")
// 	description := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(packageRepositoryText)

// 	var b strings.Builder
// 	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n")
// 	b.WriteString(description + "\n\n")

// 	// Toggle for enabling/disabling the package repository
// 	toggleText := "Enable Package Repository"
// 	if m.PackageRepo.Enabled {
// 		toggleText = "✓ " + toggleText
// 	} else {
// 		toggleText = "  " + toggleText
// 	}

// 	// Highlight the toggle if it's the active field
// 	toggleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue))
// 	if m.PackageRepo.Enabled {
// 		toggleStyle = toggleStyle.Bold(true)
// 	}
// 	b.WriteString(toggleStyle.Render(toggleText) + "\n\n")

// 	// Show the address input if enabled
// 	if m.PackageRepo.Enabled {
// 		b.WriteString(styleLabel.Render("Package Repository Address:") + "\n")
// 		b.WriteString(styleInput.Render(m.PackageRepo.Address.View()) + "\n\n")

// 		// Show error if any
// 		if m.PackageRepo.Error != "" {
// 			b.WriteString(styleError.Render(m.PackageRepo.Error) + "\n\n")
// 		}

// 		b.WriteString(styleHelpText.Render("Enter the URL of your package repository (e.g., http://package-repo.local:8080"))
// 	} else {
// 		b.WriteString(styleHelpText.Render("Press SPACE to enable and configure a package repository") + "\n\n")
// 	}

// 	// Pad the content to maintain consistent height
// 	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
// 	contentBody := strings.Join(lines, "\n")

// 	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
// }

// // renderNodeCount displays the interface for selecting node counts and API endpoint.
// func (m Model) renderNodeCount(helpWithBorder string) string {
// 	const boxHeight = 22
// 	title := styleHeader.Render("Cluster Configuration")
// 	description := lipgloss.NewStyle().
// 		Foreground(lipgloss.Color(colorMainWhite)).
// 		Render("Configure your cluster topology and API endpoint")

// 	var b strings.Builder
// 	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n")
// 	b.WriteString(description + "\n\n")

// 	// Define fields with labels and inputs
// 	fields := []struct {
// 		Label string
// 		Input textinput.Model
// 	}{
// 		{"Total Node Count:", m.NodeCount.NodeCountInput},
// 		{"Control Plane Count:", m.NodeCount.ControlPlaneCountInput},
// 		{"API Endpoint:", m.NodeCount.APIEndpointInput},
// 	}

// 	// Render each field
// 	for _, field := range fields {
// 		inputLine := lipgloss.JoinHorizontal(
// 			lipgloss.Center,
// 			styleLabel.Render(field.Label),
// 			" ",
// 			styleInput.Render(field.Input.View()),
// 		)
// 		b.WriteString(inputLine + "\n\n")
// 	}

// 	// Show error if any
// 	if m.NodeCount.Error != "" {
// 		b.WriteString(styleError.Width(uiWidth).Render(m.NodeCount.Error) + "\n\n")
// 	}

// 	// Add help text
// 	helpText := styleHelpText.Render(
// 		"The first N nodes (where N = Control Plane Count) will be configured as control plane nodes.\n" +
// 			"Remaining nodes will be worker nodes. API endpoint can be a comma-separated list of DNS names.",
// 	)
// 	b.WriteString(helpText + "\n")

// 	// Pad content to ensure help bar is at the bottom
// 	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
// 	contentBody := strings.Join(lines, "\n")

// 	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
// }

// // renderNodeDetails displays input fields for configuring individual nodes.
// // This view lets the height grow dynamically based on content.
// func (m Model) renderNodeDetails(helpWithBorder string) string {
// 	header := styleHeader.Render(fmt.Sprintf("Node %d/%d", m.Nodes.Current+1, len(m.Nodes.Inputs)))

// 	currentNode := m.Nodes.Inputs[m.Nodes.Current]
// 	fields := []struct {
// 		Label string
// 		Input textinput.Model
// 	}{
// 		{"Address:", currentNode.Address},
// 		{"Username:", currentNode.Username},
// 		{"SSH Key Path:", currentNode.SSHKeyPath},
// 	}

// 	var b strings.Builder
// 	b.WriteString(header + "\n\n") // Extra line after header
// 	for i, field := range fields {
// 		fieldLine := lipgloss.JoinHorizontal(
// 			lipgloss.Center,
// 			styleLabel.Render(field.Label),
// 			styleInput.Render(field.Input.View()),
// 		)
// 		b.WriteString(fieldLine + "\n")
// 		// Add an extra line after each field, except potentially the last if you prefer
// 		// Or add extra line only between fields
// 		if i < len(fields)-1 { // Add space between fields, not after the last one
// 			b.WriteString("\n")
// 		}
// 	}

// 	return styleBox.Width(uiWidth).Render(b.String() + "\n" + helpWithBorder)
// }

// // renderCSIToggle displays the CSI driver configuration toggle.
// func (m Model) renderCSIToggle(helpWithBorder string) string {
// 	const boxHeight = 20 // Increased height for better spacing
// 	title := styleHeader.Render("Important Disclaimer: Storage CSI Driver for Kubermatic-virtualization")
// 	disclaimerLabel := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Bold(true).Render("Please be advised:")
// 	disclaimerText := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(csiDisclaimerText)
// 	checkboxState := "[ ]"
// 	if m.CSIEnabled {
// 		checkboxState = "[x]"
// 	}
// 	checkbox := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(checkboxState)
// 	toggleLabel := "Install default CSI driver (for evaluation only)"

// 	var b strings.Builder
// 	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n") // Extra line after title
// 	b.WriteString(disclaimerLabel + disclaimerText + "\n\n")                          // Extra line after disclaimer
// 	b.WriteString(fmt.Sprintf("%s %s\n", checkbox, toggleLabel))                      // Line for the toggle

// 	// Pad content to ensure help bar is at the bottom
// 	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
// 	contentBody := strings.Join(lines, "\n")

// 	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
// }

// // renderReview displays the final configuration for user confirmation.
// func (m Model) renderReview(helpWithBorder string) string {
// 	header := styleHeader.Render("Review Configuration")

// 	// Build the main content
// 	var b strings.Builder
// 	b.WriteString(header + "\n\n") // Extra line after header

// 	// Display viewport content or fallback
// 	if m.Review.Viewport.Height > 0 {
// 		b.WriteString(m.Review.Viewport.View())
// 	} else {
// 		// Fallback if viewport isn't initialized
// 		b.WriteString(styleItem.Render(m.Review.ConfigYAML))
// 	}

// 	return styleBox.Width(uiWidth).Render(b.String() + "\n" + helpWithBorder)
// }

// renderExecuting displays logs during the configuration application process.
func (m Model) renderExecuting(helpWithBorder string) string {
	header := styleHeader.Render("Applying Configuration")

	// Build the main content
	var b strings.Builder
	b.WriteString(header + "\n\n") // Extra line after header
	b.WriteString(m.Review.Viewport.View())

	if m.executionError != "" {
		b.WriteString("\n" + styleError.Render(m.executionError))
	}

	return styleBox.Width(uiWidth).Render(b.String() + "\n" + helpWithBorder)
}

// renderDone displays the final success message.
func (m Model) renderDone(helpWithBorder string) string {
	header := styleHeader.Render("Congratulations!")
	var message string
	if m.executionError != "" {
		header = styleHeader.Render("Execution Finished With Errors")
		message = styleError.Render(m.executionError)
	} else {
		message = styleItem.Render(successInstallationText)
	}

	// Build the main content
	var b strings.Builder
	b.WriteString(header + "\n\n") // Extra line after header
	b.WriteString(message)

	return styleBox.Width(uiWidth).Render(b.String() + "\n" + helpWithBorder)
}

// bannerText returns the application banner/logo.
func bannerText() string {
	logo := `
░█▀▀░█▀█░█▀█░█▀▀░█▀█░█▀▄░█▄ ▄█░█▀█░█▀█░█▀▀░█▀▀░░░▀█▀░█▀▀░█▀▀░▀█▀░█▀▀░█▀▄
░█░░░█░█░█░█░█▀▀░█░█░█▀▄░█░▀░█░█▀█░█░█░█░░░█▀▀░░░░█░░█▀▀░▀▀█░░█░░█▀▀░█▀▄
░▀▀▀░▀▀▀░▀░▀░▀░░░▀▀▀░▀░▀░▀░░░▀░▀░▀░▀░▀░▀▀▀░▀▀▀░░░░▀░░▀▀▀░▀▀▀░░▀░░▀▀▀░▀░▀
	By Kubermatic `
	return logo
}

// helpBar returns context-sensitive help text for each stage.
func helpBar(stage int) string {
	// Using a map can be slightly more efficient and clearer for static mappings.
	helpTexts := map[int]string{
		stageWelcome:                            "Press Enter to continue.",
		stageEnvironmentSelection:               "↑/↓ to navigate, Space to select, Tab/Shift+Tab to move between fields, Enter to continue, Esc to go back.",
		stageReleaseSelection:                   "↑/↓ to navigate, Space to select, CTRL+A to select/deselect all, Enter to continue, Esc to go back.",
		stageProviderSelection:                  "↑/↓ to navigate, Space to select, Tab/Shift+Tab to move between fields, Enter to continue, Esc to go back.",
		stageDistributionSelection:              "↑/↓ to navigate, Space to select, CTRL+A to select/deselect all, Enter to continue, Esc to go back.",
		stageDatacenterSettingsSelection:        "↑/↓ to navigate, Space to select, Enter to continue, Esc to go back.",
		stageClusterSettingsSelection:           "↑/↓ to navigate, Space to select, Enter to continue, Esc to go back.",
		stageMachineDeploymentSettingsSelection: "↑/↓ to navigate, Space to select, Enter to continue, Esc to go back.",
		stageClusterSettings:                    "↑/↓ to navigate, Space to select, Enter to continue, Esc to go back.",
		stageReviewSettings:                     "↑/↓ to scroll, PgUp/PgDn for faster scroll, Enter to confirm, ← to go back.",
		stageExecuting:                          "Logs will appear here. Press ctrl+c to cancel.",
		stageDone:                               "Press q to quit.",
	}
	// Return empty string for unknown stages
	return helpTexts[stage]
}

// // RenderComponentTable returns a table with the components used in Kubermatic Virtualization.
// func RenderComponentTable(components []kubeone.ComponentInfo) string {
// 	cellWidth := 35
// 	// Header with same bold treatment
// 	headerStyle := lipgloss.NewStyle().
// 		Bold(true).
// 		Background(lipgloss.Color("#005f87")).
// 		Foreground(lipgloss.Color("#ffffff")).
// 		Width(cellWidth).
// 		Align(lipgloss.Center).
// 		Padding(0, 1)

// 	cellStyle := lipgloss.NewStyle().
// 		Bold(true).
// 		Width(cellWidth).
// 		Padding(0, 1)

// 	columns := []table.Column{
// 		{Title: "Component", Width: cellWidth},
// 		{Title: "Version", Width: cellWidth},
// 	}

// 	// Convert components to table rows
// 	rows := make([]table.Row, len(components))
// 	for i, comp := range components {
// 		rows[i] = table.Row{comp.Name, comp.Version}
// 	}

// 	t := table.New(
// 		table.WithColumns(columns),
// 		table.WithRows(rows),
// 		table.WithHeight(len(rows)+1),
// 	)

// 	t.SetStyles(table.Styles{
// 		Header:   headerStyle,
// 		Cell:     cellStyle,
// 		Selected: lipgloss.NewStyle().Bold(true),
// 	})

// 	// Render the table first to get its width
// 	tableContent := lipgloss.NewStyle().
// 		Border(lipgloss.RoundedBorder()).
// 		BorderForeground(lipgloss.Color("#00afff")).
// 		Render(t.View())

// 	// Get the width of the rendered table
// 	tableWidth, _ := lipgloss.Size(tableContent)

// 	// Create banner with the same width as the table
// 	banner := lipgloss.NewStyle().
// 		Foreground(lipgloss.Color("#FFFFFF")).
// 		Bold(true).
// 		PaddingTop(1).
// 		Align(lipgloss.Center).
// 		Width(tableWidth).
// 		Render("Kubermatic Virtualization Details")

// 	return banner + "\n" + tableContent
// }

// padContentToHeight ensures content has a minimum number of lines.
// This is used to keep the help bar consistently positioned at the bottom.
func padContentToHeight(content string, minHeight int) []string {
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	for len(lines) < minHeight {
		lines = append(lines, "")
	}
	return lines
}

// --- Static Text Content ---

const welcomeTitleText = "Welcome to Kubermatic Conformance Tester"

const welcomeDisclaimerText = `Kubermatic Conformance Tester is an automated testing tool that validates Kubernetes cluster functionality and compliance across multiple cloud providers. It performs comprehensive conformance tests including storage, networking, load balancing, and security context validation to ensure your clusters meet production standards.`

const environmentSelectionText = `Choose whether you want to run conformance tester using an already existing Kubernetes cluster or set up a new KKP instance locally.`

const releaseSelectionText = `Select the supported kubernetes release version for your cluster. Kubermatic Conformance Tester ensures compatibility and performance across different Kubernetes versions by providing options for various releases.`

const providerSelectionText = `Select the infrastructure provider where your Kubernetes cluster will be deployed. Kubermatic Conformance Tester supports multiple providers to ensure compatibility and performance across different environments.`

const distributionSelectionText = `Select one or more operating system distributions for your cluster nodes. Multiple distributions can be selected to test compatibility across different OS environments.`

const containerRegistryText = `Specify the address of your offline container registry—and credentials if authentication is required. Kubermatic Conformance Tester will rely exclusively on this registry for all container images. Toggle the insecure option if the registry uses self-signed certificates or HTTP.`
const helmRegistryText = `Specify the address of your offline Helm registry—and credentials if authentication is required. Kubermatic Virtualization will rely exclusively on this registry for all Helm charts. Toggle the insecure option if the registry uses self-signed certificates or HTTP.`

const packageRepositoryText = `Kubermatic Virtualization provides a package repository for offline environments. If you are using it, specify the address or hostname of the repository server. In that case, Kubermatic Virtualization will rely exclusively on this repository for all package installations and updates.

If you are not using the repository, leave this field empty and configure your system manually to use your custom package repository.`

const networkConfigDescription = `Network configuration is necessary to configure the Kubermatic virtualization stack. Please provide the following details:`

const metalLBDisclaimerText = `MetalLB in Kubermatic Virtualization is provided as the default load balancer solution. While MetalLB is a mature and production-ready solution, its configuration, maintenance, and suitability for your specific use case are the responsibility of the user. Ensure that it aligns with your organization's operational requirements, scalability needs, and compliance policies.`

const csiDisclaimerText = ` The default Container Storage Interface (CSI) driver provided by Kubermatic-virtualization is not intended or supported for production environments.
This CSI driver is specifically offered for evaluation and staging purposes only. It serves to provide baseline storage functionality for Kubermatic-virtualization during these phases.
Kubermatic does not guarantee the ongoing maintenance, reliability, or performance of this default CSI driver. Its use in any production capacity is at your own risk and is not recommended. For production deployments, we strongly advise utilizing a fully supported and robust storage solution.`

const successInstallationText = `Kubermatic Virtualization has been successfully installed! You can now proceed with configuring and using the virtualization features. If you encounter any issues, feel free to consult the documentation or support resources.`

const executionWarning = `Exiting now will interrupt the kubev bootstrap process. Any partial setup will remain on your servers, and simply rerunning the installer will not clean it up—you'll need to reset the servers manually before trying again.`
