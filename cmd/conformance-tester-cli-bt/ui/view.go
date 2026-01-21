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
	uiBoxHeightPad = 2 // Adjustment for top/bottom borders in boxStyle
)

// renderQuitConfirm draws the quit confirmation dialog.
func (m Model) renderQuitConfirm(uiWidth, uiInnerWidth int) string {
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
func (m Model) renderWelcome(helpWithBorder string, uiWidth, uiInnerWidth int) string {
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

func (m Model) renderEnvironmentSelection(helpWithBorder string, uiWidth, uiInnerWidth int) string {
	const boxHeight = 30 // Increased height for kubeconfig options
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
		// Kubeconfig selection
		kubeconfigLabel := styleLabel.Render("Kubeconfig:")
		b.WriteString(kubeconfigLabel + "\n")

		// Group options by type
		var envOptions, fileOptions, customOptions []KubeconfigOption
		for _, option := range m.existingEnv.KubeconfigOptions {
			switch option.Type {
			case "env":
				envOptions = append(envOptions, option)
			case "file":
				fileOptions = append(fileOptions, option)
			case "custom":
				customOptions = append(customOptions, option)
			}
		}

		currentOptionIndex := 0

		// Section 1: Environment Variable
		if len(envOptions) > 0 {
			isExpanded := m.existingEnv.KubeconfigExpandedSections["env"]
			isSectionFocused := m.environmentFocusIndex == 1 && m.environmentFieldIndex == 1 && currentOptionIndex == m.existingEnv.KubeconfigFocusedIndex

			expandIndicator := "▶"
			if isExpanded {
				expandIndicator = "▼"
			}

			sectionHeaderText := fmt.Sprintf("\t\t\t%s Environment Variable", expandIndicator)
			if isSectionFocused {
				sectionHeaderText = styleFocusHighlight.Render(sectionHeaderText)
			} else {
				sectionHeaderText = lipgloss.NewStyle().
					Foreground(lipgloss.Color(colorMainBlue)).
					Bold(true).
					Render(sectionHeaderText)
			}
			b.WriteString(sectionHeaderText + "\n")

			sectionDesc := lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorMainWhite)).
				Faint(true).
				Render("\t\t\tKubeconfig path from KUBECONFIG environment variable")
			b.WriteString(sectionDesc + "\n")
			currentOptionIndex++

			if isExpanded {

				for i, option := range envOptions {
					optionIndex := currentOptionIndex + i
					isFocused := m.environmentFocusIndex == 1 && m.environmentFieldIndex == 1 && optionIndex == m.existingEnv.KubeconfigFocusedIndex

					radioBtn := "( )"
					if option.Selected {
						radioBtn = "(•)"
					}

					optionLine := fmt.Sprintf("\t\t\t\t%s %s",
						lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(radioBtn),
						option.Path)

					if isFocused {
						optionLine = styleFocusHighlight.Render(optionLine)
					} else if option.Selected {
						optionLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(optionLine)
					} else {
						optionLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(optionLine)
					}

					b.WriteString(optionLine + "\n")
				}
				currentOptionIndex += len(envOptions)
			}
			b.WriteString("\n")
		}

		// Section 2: Files from .kube directory
		if len(fileOptions) > 0 {
			isExpanded := m.existingEnv.KubeconfigExpandedSections["file"]
			isSectionFocused := m.environmentFocusIndex == 1 && m.environmentFieldIndex == 1 && currentOptionIndex == m.existingEnv.KubeconfigFocusedIndex

			expandIndicator := "▶"
			if isExpanded {
				expandIndicator = "▼"
			}

			sectionHeaderText := fmt.Sprintf("\t\t\t%s Kubeconfigs from ~/.kube", expandIndicator)
			if isSectionFocused {
				sectionHeaderText = styleFocusHighlight.Render(sectionHeaderText)
			} else {
				sectionHeaderText = lipgloss.NewStyle().
					Foreground(lipgloss.Color(colorMainBlue)).
					Bold(true).
					Render(sectionHeaderText)
			}
			b.WriteString(sectionHeaderText + "\n")

			sectionDesc := lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorMainWhite)).
				Faint(true).
				Render("\t\t\tDiscovered kubeconfig files in your .kube directory")
			b.WriteString(sectionDesc + "\n")
			currentOptionIndex++

			if isExpanded {

				for i, option := range fileOptions {
					optionIndex := currentOptionIndex + i
					isFocused := m.environmentFocusIndex == 1 && m.environmentFieldIndex == 1 && optionIndex == m.existingEnv.KubeconfigFocusedIndex

					radioBtn := "( )"
					if option.Selected {
						radioBtn = "(•)"
					}

					// Extract just the filename for display
					filename := option.DisplayName[len("~/.kube/"):]

					// Build the option line with radio button and filename
					radioBtnStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(radioBtn)
					var optionLine string

					// If selected, show filename and path together with path faint
					if option.Selected {
						pathStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Faint(true).Render(fmt.Sprintf(" → %s", option.Path))
						optionLine = fmt.Sprintf("\t\t\t\t%s %s%s", radioBtnStyled, filename, pathStyled)
					} else {
						optionLine = fmt.Sprintf("\t\t\t\t%s %s", radioBtnStyled, filename)
					}

					if isFocused {
						optionLine = styleFocusHighlight.Render(optionLine)
					} else if option.Selected {
						// For selected items, only bold the filename part, keep path faint
						filenameBold := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(filename)
						pathStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Faint(true).Render(fmt.Sprintf(" → %s", option.Path))
						optionLine = fmt.Sprintf("\t\t\t\t%s %s%s", radioBtnStyled, filenameBold, pathStyled)
					} else {
						optionLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(optionLine)
					}

					b.WriteString(optionLine + "\n")
				}
				currentOptionIndex += len(fileOptions)
			}
			b.WriteString("\n")
		}

		// Section 3: Custom Path
		if len(customOptions) > 0 {
			isExpanded := m.existingEnv.KubeconfigExpandedSections["custom"]
			isSectionFocused := m.environmentFocusIndex == 1 && m.environmentFieldIndex == 1 && currentOptionIndex == m.existingEnv.KubeconfigFocusedIndex

			expandIndicator := "▶"
			if isExpanded {
				expandIndicator = "▼"
			}

			sectionHeaderText := fmt.Sprintf("\t\t\t%s Custom Path", expandIndicator)
			if isSectionFocused {
				sectionHeaderText = styleFocusHighlight.Render(sectionHeaderText)
			} else {
				sectionHeaderText = lipgloss.NewStyle().
					Foreground(lipgloss.Color(colorMainBlue)).
					Bold(true).
					Render(sectionHeaderText)
			}
			b.WriteString(sectionHeaderText + "\n")

			sectionDesc := lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorMainWhite)).
				Faint(true).
				Render("\t\t\tSpecify a custom path to your kubeconfig file")
			b.WriteString(sectionDesc + "\n")
			currentOptionIndex++

			if isExpanded {

				for i, option := range customOptions {
					optionIndex := currentOptionIndex + i
					isFocused := m.environmentFocusIndex == 1 && m.environmentFieldIndex == 1 && optionIndex == m.existingEnv.KubeconfigFocusedIndex

					// Always show radio button
					radioBtn := "( )"
					if option.Selected {
						radioBtn = "(•)"
					}

					optionLine := fmt.Sprintf("\t\t\t\t%s %s",
						lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(radioBtn), "Use custom path")

					if isFocused {
						optionLine = styleFocusHighlight.Render(optionLine)
					} else if option.Selected {
						optionLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(optionLine)
					} else {
						optionLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(optionLine)
					}

					b.WriteString(optionLine + "\n")

					// Show custom path input on a new line when selected
					if option.Selected {
						line := lipgloss.JoinHorizontal(
							lipgloss.Left,
							styleLabel.Render("Custom Path:"),
							" ",
							styleInput.Render(m.existingEnv.CustomKubeconfigPath.View()),
						)
						b.WriteString(line + "\n")
					}
				}
				currentOptionIndex += len(customOptions)
			}
			b.WriteString("\n")
		}

		// Show kubeconfig error if present
		if err := m.existingEnv.Errors.KubeconfigPath; err != "" {
			b.WriteString(styleError.Width(uiWidth-4).Render(err) + "\n")
		}

		// Other existing cluster fields
		existingFields := []struct {
			Label    string
			Input    textinput.Model
			Error    string
			FieldIdx int
		}{
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

func (m Model) renderReleaseSelection(helpWithBorder string, uiWidth, uiInnerWidth int) string {
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

func (m Model) renderProviderSelection(helpWithBorder string, uiWidth, uiInnerWidth int) string {
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
			b.WriteString(m.renderProviderCredentials(provider, i, uiWidth) + "\n")
		}
	}

	// Pad content to ensure help bar is at the bottom
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")
	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)

}

// renderProviderCredentials renders the credential fields for a specific provider.
func (m Model) renderProviderCredentials(provider Provider, providerIndex int, uiWidth int) string {
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

func (m Model) renderDistributionSelection(helpWithBorder string, uiWidth, uiInnerWidth int) string {
	const boxHeight = 20
	title := styleHeader.Render("Select Operating System Distributions")
	description := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(distributionSelectionText)

	var b strings.Builder
	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n")
	b.WriteString(description + "\n\n")

	// Check if no distributions are available
	hasDistributions := false
	for _, provider := range m.distributionSelection.Providers {
		if len(m.distributionSelection.DistributionsByProvider[provider]) > 0 {
			hasDistributions = true
			break
		}
	}

	if !hasDistributions {
		var providersStr string
		if len(m.distributionSelection.Providers) == 1 {
			providersStr = m.distributionSelection.Providers[0]
		} else {
			providersStr = strings.Join(m.distributionSelection.Providers, ", ")
		}
		noDistMsg := fmt.Sprintf("No distributions available for %s.", providersStr)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorErrorRed)).Render(noDistMsg))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render("Press Enter to continue"))
		b.WriteString("\n\n")
		lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
		contentBody := strings.Join(lines, "\n")
		return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
	}

	// Render hierarchical structure: Provider → Distributions
	currentIndex := 0
	for providerIdx, provider := range m.distributionSelection.Providers {
		isProviderExpanded := m.distributionSelection.ExpandedProviders[provider]
		isProviderFocused := currentIndex == m.distributionSelection.FocusedIndex

		// Provider header with expand/collapse indicator
		expandIndicator := "▶"
		if isProviderExpanded {
			expandIndicator = "▼"
		}

		providerHeader := fmt.Sprintf("%s %s", expandIndicator, provider)
		if len(m.distributionSelection.Providers) > 1 {
			if isProviderFocused {
				providerHeader = styleFocusHighlight.Render(providerHeader)
			} else {
				providerHeader = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(providerHeader)
			}
			b.WriteString(providerHeader + "\n")
		}
		currentIndex++

		// Render distributions for this provider if expanded
		if isProviderExpanded {
			dists := m.distributionSelection.DistributionsByProvider[provider]
			for _, dist := range dists {
				isDistFocused := currentIndex == m.distributionSelection.FocusedIndex
				selectionKey := fmt.Sprintf("%s:%s", provider, dist)
				isSelected := m.distributionSelection.Selected[selectionKey]

				checkbox := "[ ]"
				if isSelected {
					checkbox = "[x]"
				}

				displayName := m.distributionSelection.DistributionNames[dist]
				distLine := fmt.Sprintf("  %s %s",
					lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(checkbox),
					displayName)

				if isDistFocused {
					distLine = styleFocusHighlight.Render(distLine)
				} else if isSelected {
					distLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(distLine)
				} else {
					distLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(distLine)
				}

				b.WriteString(distLine + "\n")
				currentIndex++
			}
		}

		// Add spacing between provider sections
		if providerIdx < len(m.distributionSelection.Providers)-1 {
			b.WriteString("\n")
		}
	}

	// pad content to ensure help bar is at the bottom
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")
	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
}

// renderDatacenterSettingsSelection renders the datacenter settings selection stage.
func (m Model) renderDatacenterSettingsSelection(helpWithBorder string, uiWidth, uiInnerWidth int) string {
	const boxHeight = 20

	// Build title based on number of providers
	var title string
	if len(m.datacenterSettingsSelection.Providers) == 1 {
		title = styleHeader.Render(fmt.Sprintf("%s Datacenter Settings Selection", m.datacenterSettingsSelection.Providers[0]))
	} else {
		title = styleHeader.Render("Datacenter Settings Selection")
	}

	description := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render("Select datacenter settings to test. Selecting none will use default values (typically false) for all settings.")

	var b strings.Builder
	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n")
	b.WriteString(description + "\n\n")

	// Check if no settings are available
	hasSettings := false
	for _, provider := range m.datacenterSettingsSelection.Providers {
		if len(m.datacenterSettingsSelection.SettingsByProvider[provider]) > 0 {
			hasSettings = true
			break
		}
	}

	if !hasSettings {
		var providersStr string
		if len(m.datacenterSettingsSelection.Providers) == 1 {
			providersStr = m.datacenterSettingsSelection.Providers[0]
		} else {
			providersStr = strings.Join(m.datacenterSettingsSelection.Providers, ", ")
		}
		noSettingsMsg := fmt.Sprintf("No datacenter settings available for %s.", providersStr)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorErrorRed)).Render(noSettingsMsg))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render("Press Enter to continue"))
		b.WriteString("\n\n")
		lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
		contentBody := strings.Join(lines, "\n")
		return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
	}

	// Render hierarchical structure: Provider → Group → Options
	currentIndex := 0
	for providerIdx, provider := range m.datacenterSettingsSelection.Providers {
		isProviderExpanded := m.datacenterSettingsSelection.ExpandedProviders[provider]
		isProviderFocused := currentIndex == m.datacenterSettingsSelection.FocusedIndex

		// Provider header with expand/collapse indicator
		expandIndicator := "▶"
		if isProviderExpanded {
			expandIndicator = "▼"
		}

		providerHeader := fmt.Sprintf("%s %s", expandIndicator, provider)
		if len(m.datacenterSettingsSelection.Providers) > 1 {
			if isProviderFocused {
				providerHeader = styleFocusHighlight.Render(providerHeader)
			} else {
				providerHeader = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(providerHeader)
			}
			b.WriteString(providerHeader + "\n")
		}
		currentIndex++

		// Render setting groups for this provider if expanded
		if isProviderExpanded {
			groups := m.datacenterSettingsSelection.SettingsByProvider[provider]
			for groupIdx, group := range groups {
				isGroupFocused := currentIndex == m.datacenterSettingsSelection.FocusedIndex
				groupKey := fmt.Sprintf("%s:%s", provider, group.Key)
				isGroupSelected := m.datacenterSettingsSelection.SelectedGroups[groupKey]

				// Setting group with checkbox
				checkbox := "[ ]"
				if isGroupSelected {
					checkbox = "[✓]"
				}

				groupHeader := fmt.Sprintf("  %s %s",
					lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(checkbox),
					group.Name)
				if isGroupFocused {
					groupHeader = styleFocusHighlight.Render(groupHeader)
				} else if isGroupSelected {
					groupHeader = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(groupHeader)
				} else {
					groupHeader = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(groupHeader)
				}
				b.WriteString(groupHeader + "\n")
				currentIndex++

				// Render options for this group (always shown)
				for optionIdx, option := range group.Options {
					isOptionFocused := currentIndex == m.datacenterSettingsSelection.FocusedIndex
					selectionKey := fmt.Sprintf("%s:%s:%s", provider, group.Key, option)
					isSelected := m.datacenterSettingsSelection.Selected[selectionKey]

					checkbox := "[ ]"
					if isSelected {
						checkbox = "[✓]"
					}

					optionLine := fmt.Sprintf("    %s %s",
						lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(checkbox),
						option)

					if isOptionFocused {
						optionLine = styleFocusHighlight.Render(optionLine)
					} else if isSelected {
						optionLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(optionLine)
					} else {
						optionLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(optionLine)
					}

					b.WriteString(optionLine + "\n")
					currentIndex++
					_ = optionIdx // Suppress unused variable warning
				}
				_ = groupIdx // Suppress unused variable warning
			}
		}

		// Add spacing between provider sections
		if providerIdx < len(m.datacenterSettingsSelection.Providers)-1 {
			b.WriteString("\n")
		}
	}

	// Pad content to ensure help bar is at the bottom
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")
	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
}

// renderClusterSettingsSelection renders the cluster settings selection stage.
func (m Model) renderClusterSettingsSelection(helpWithBorder string, uiWidth, uiInnerWidth int) string {
	const boxHeight = 20

	// Build title based on number of providers
	var title string
	if len(m.clusterSettingsSelection.Providers) == 1 {
		title = styleHeader.Render(fmt.Sprintf("%s Cluster Settings Selection", m.clusterSettingsSelection.Providers[0]))
	} else {
		title = styleHeader.Render("Cluster Settings Selection")
	}

	description := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render("Select cluster settings to test. Selecting none will use default values (typically false) for all settings.")

	var b strings.Builder
	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n")
	b.WriteString(description + "\n\n")

	// Check if no settings are available
	hasSettings := false
	for _, provider := range m.clusterSettingsSelection.Providers {
		if len(m.clusterSettingsSelection.SettingsByProvider[provider]) > 0 {
			hasSettings = true
			break
		}
	}

	if !hasSettings {
		var providersStr string
		if len(m.clusterSettingsSelection.Providers) == 1 {
			providersStr = m.clusterSettingsSelection.Providers[0]
		} else {
			providersStr = strings.Join(m.clusterSettingsSelection.Providers, ", ")
		}
		noSettingsMsg := fmt.Sprintf("No cluster settings available for %s.", providersStr)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorErrorRed)).Render(noSettingsMsg))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render("Press Enter to continue"))
		b.WriteString("\n\n")
		lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
		contentBody := strings.Join(lines, "\n")
		return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
	}

	// Render hierarchical structure: Provider → Group → Options
	currentIndex := 0
	for providerIdx, provider := range m.clusterSettingsSelection.Providers {
		isProviderExpanded := m.clusterSettingsSelection.ExpandedProviders[provider]
		isProviderFocused := currentIndex == m.clusterSettingsSelection.FocusedIndex

		// Provider header with expand/collapse indicator
		expandIndicator := "▶"
		if isProviderExpanded {
			expandIndicator = "▼"
		}

		providerHeader := fmt.Sprintf("%s %s", expandIndicator, provider)
		if len(m.clusterSettingsSelection.Providers) > 1 {
			if isProviderFocused {
				providerHeader = styleFocusHighlight.Render(providerHeader)
			} else {
				providerHeader = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(providerHeader)
			}
			b.WriteString(providerHeader + "\n")
		}
		currentIndex++

		// Render setting groups for this provider if expanded
		if isProviderExpanded {
			groups := m.clusterSettingsSelection.SettingsByProvider[provider]
			for groupIdx, group := range groups {
				isGroupFocused := currentIndex == m.clusterSettingsSelection.FocusedIndex
				groupKey := fmt.Sprintf("%s:%s", provider, group.Key)
				isGroupSelected := m.clusterSettingsSelection.SelectedGroups[groupKey]

				// Setting group with checkbox
				checkbox := "[ ]"
				if isGroupSelected {
					checkbox = "[✓]"
				}

				groupHeader := fmt.Sprintf("  %s %s",
					lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(checkbox),
					group.Name)
				if isGroupFocused {
					groupHeader = styleFocusHighlight.Render(groupHeader)
				} else if isGroupSelected {
					groupHeader = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(groupHeader)
				} else {
					groupHeader = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(groupHeader)
				}
				b.WriteString(groupHeader + "\n")
				currentIndex++

				// Render options for this group (always shown)
				for optionIdx, option := range group.Options {
					isOptionFocused := currentIndex == m.clusterSettingsSelection.FocusedIndex
					selectionKey := fmt.Sprintf("%s:%s:%s", provider, group.Key, option)
					isSelected := m.clusterSettingsSelection.Selected[selectionKey]

					checkbox := "[ ]"
					if isSelected {
						checkbox = "[✓]"
					}

					optionLine := fmt.Sprintf("    %s %s",
						lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(checkbox),
						option)

					if isOptionFocused {
						optionLine = styleFocusHighlight.Render(optionLine)
					} else if isSelected {
						optionLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(optionLine)
					} else {
						optionLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(optionLine)
					}

					b.WriteString(optionLine + "\n")
					currentIndex++
					_ = optionIdx // Suppress unused variable warning
				}
				_ = groupIdx // Suppress unused variable warning
			}
		}

		// Add spacing between provider sections
		if providerIdx < len(m.clusterSettingsSelection.Providers)-1 {
			b.WriteString("\n")
		}
	}

	// Pad content to ensure help bar is at the bottom
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")
	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
}

// renderMachineDeploymentSettingsSelection renders the machine deployment settings selection stage.
func (m Model) renderMachineDeploymentSettingsSelection(helpWithBorder string, uiWidth, uiInnerWidth int) string {
	const boxHeight = 20

	// Build title based on number of providers
	var title string
	if len(m.machineDeploymentSettingsSelection.Providers) == 1 {
		title = styleHeader.Render(fmt.Sprintf("%s Machine Deployment Settings Selection", m.machineDeploymentSettingsSelection.Providers[0]))
	} else {
		title = styleHeader.Render("Machine Deployment Settings Selection")
	}

	description := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render("Select machine deployment settings to test. Selecting none will use default values (typically false) for all settings.")

	var b strings.Builder
	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n")
	b.WriteString(description + "\n\n")

	// Check if no settings are available
	hasSettings := false
	for _, provider := range m.machineDeploymentSettingsSelection.Providers {
		if len(m.machineDeploymentSettingsSelection.SettingsByProvider[provider]) > 0 {
			hasSettings = true
			break
		}
	}

	if !hasSettings {
		var providersStr string
		if len(m.machineDeploymentSettingsSelection.Providers) == 1 {
			providersStr = m.machineDeploymentSettingsSelection.Providers[0]
		} else {
			providersStr = strings.Join(m.machineDeploymentSettingsSelection.Providers, ", ")
		}
		noSettingsMsg := fmt.Sprintf("No machine deployment settings available for %s.", providersStr)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorErrorRed)).Render(noSettingsMsg))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render("Press Enter to continue"))
		b.WriteString("\n\n")
		lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
		contentBody := strings.Join(lines, "\n")
		return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
	}

	// Render hierarchical structure: Provider → Group → Options
	currentIndex := 0
	for providerIdx, provider := range m.machineDeploymentSettingsSelection.Providers {
		isProviderExpanded := m.machineDeploymentSettingsSelection.ExpandedProviders[provider]
		isProviderFocused := currentIndex == m.machineDeploymentSettingsSelection.FocusedIndex

		// Provider header with expand/collapse indicator
		expandIndicator := "▶"
		if isProviderExpanded {
			expandIndicator = "▼"
		}

		providerHeader := fmt.Sprintf("%s %s", expandIndicator, provider)
		if len(m.machineDeploymentSettingsSelection.Providers) > 1 {
			if isProviderFocused {
				providerHeader = styleFocusHighlight.Render(providerHeader)
			} else {
				providerHeader = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(providerHeader)
			}
			b.WriteString(providerHeader + "\n")
		}
		currentIndex++

		// Render setting groups for this provider if expanded
		if isProviderExpanded {
			groups := m.machineDeploymentSettingsSelection.SettingsByProvider[provider]
			for groupIdx, group := range groups {
				isGroupFocused := currentIndex == m.machineDeploymentSettingsSelection.FocusedIndex
				groupKey := fmt.Sprintf("%s:%s", provider, group.Key)
				isGroupSelected := m.machineDeploymentSettingsSelection.SelectedGroups[groupKey]

				// Setting group with checkbox
				checkbox := "[ ]"
				if isGroupSelected {
					checkbox = "[✓]"
				}

				groupHeader := fmt.Sprintf("  %s %s",
					lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(checkbox),
					group.Name)
				if isGroupFocused {
					groupHeader = styleFocusHighlight.Render(groupHeader)
				} else if isGroupSelected {
					groupHeader = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(groupHeader)
				} else {
					groupHeader = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(groupHeader)
				}
				b.WriteString(groupHeader + "\n")
				currentIndex++

				// Render options for this group (always shown)
				for optionIdx, option := range group.Options {
					isOptionFocused := currentIndex == m.machineDeploymentSettingsSelection.FocusedIndex
					selectionKey := fmt.Sprintf("%s:%s:%s", provider, group.Key, option)
					isSelected := m.machineDeploymentSettingsSelection.Selected[selectionKey]

					checkbox := "[ ]"
					if isSelected {
						checkbox = "[✓]"
					}

					optionLine := fmt.Sprintf("    %s %s",
						lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(checkbox),
						option)

					if isOptionFocused {
						optionLine = styleFocusHighlight.Render(optionLine)
					} else if isSelected {
						optionLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(optionLine)
					} else {
						optionLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(optionLine)
					}

					b.WriteString(optionLine + "\n")
					currentIndex++
					_ = optionIdx // Suppress unused variable warning
				}
				_ = groupIdx // Suppress unused variable warning
			}
		}

		// Add spacing between provider sections
		if providerIdx < len(m.machineDeploymentSettingsSelection.Providers)-1 {
			b.WriteString("\n")
		}
	}

	// Pad content to ensure help bar is at the bottom
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")
	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
}

// renderClusterConfiguration renders the cluster configuration stage.
func (m Model) renderClusterConfiguration(helpWithBorder string, uiWidth, uiInnerWidth int) string {
	const boxHeight = 20

	title := styleHeader.Render("Cluster Configuration")
	description := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(
		"Configure resource allocation and testing options for user clusters")

	var b strings.Builder
	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, title) + "\n\n")
	b.WriteString(description + "\n\n")

	currentIndex := 0

	// Render categories and settings
	for _, category := range m.clusterConfiguration.Categories {
		// Category header
		categoryFocused := currentIndex == m.clusterConfiguration.FocusedIndex
		isCategoryExpanded := m.clusterConfiguration.ExpandedCategories[category.Name]

		// Expand/collapse indicator
		expandIndicator := "▶"
		if isCategoryExpanded {
			expandIndicator = "▼"
		}

		categoryHeader := fmt.Sprintf("%s %s", expandIndicator, category.Name)
		if categoryFocused {
			b.WriteString(styleFocusHighlight.Render(categoryHeader))
		} else {
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorMainBlue)).
				Bold(true).
				Render(categoryHeader))
		}
		b.WriteString("\n")

		// Category description
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMainWhite)).
			Faint(true).
			Render(fmt.Sprintf("  %s", category.Description)))
		b.WriteString("\n")
		currentIndex++

		// Render settings in this category only if expanded
		if isCategoryExpanded {
			for _, setting := range category.Settings {
				settingFocused := currentIndex == m.clusterConfiguration.FocusedIndex
				isEditing := settingFocused && m.clusterConfiguration.EditMode

				// Setting name and value
				var settingLine string
				if setting.Type == ConfigTypeBool {
					// Boolean settings show as toggles
					checkbox := "[ ]"
					if setting.Value.(bool) {
						checkbox = "[✓]"
					}
					settingLine = fmt.Sprintf("  %s %s", checkbox, setting.Name)
				} else {
					// Other settings show their values
					valueStr := m.formatConfigValue(&setting)
					if isEditing {
						valueStr = m.clusterConfiguration.EditingBuffer + "█" // Show cursor
					}
					settingLine = fmt.Sprintf("  %s: %s", setting.Name, valueStr)
				}

				if settingFocused {
					b.WriteString(styleFocusHighlight.Render(settingLine))
				} else {
					b.WriteString(lipgloss.NewStyle().
						Foreground(lipgloss.Color(colorMainWhite)).
						Render(settingLine))
				}
				b.WriteString("\n")

				// Setting description (smaller, dimmed)
				descLine := fmt.Sprintf("    %s", setting.Description)
				b.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color(colorMainWhite)).
					Faint(true).
					Render(descLine))
				b.WriteString("\n")
				currentIndex++
			}
		}
		b.WriteString("\n")
	}

	// Pad content to ensure help bar is at the bottom
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")
	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
}

// renderExecuting displays logs during the configuration application process.
func (m Model) renderExecuting(helpWithBorder string, uiWidth, uiInnerWidth int) string {
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
func (m Model) renderDone(helpWithBorder string, uiWidth, uiInnerWidth int) string {
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
		stageEnvironmentSelection:               "↑/↓ to navigate, ←/→ to collapse/expand, Space to select, Tab/Shift+Tab to move between fields, Enter to continue, Esc to go back.",
		stageReleaseSelection:                   "↑/↓ to navigate, Space to select, CTRL+A to select/deselect all, Enter to continue, Esc to go back.",
		stageProviderSelection:                  "↑/↓ to navigate, Space to select, Tab/Shift+Tab to move between fields, Enter to continue, Esc to go back.",
		stageDistributionSelection:              "↑/↓ to navigate, Space to select, CTRL+A to select/deselect all, Enter to continue, Esc to go back.",
		stageDatacenterSettingsSelection:        "↑/↓ to navigate, ←/→ to collapse/expand providers, Space to select, CTRL+A to select/deselect all, Enter to continue, Esc to go back.",
		stageClusterSettingsSelection:           "↑/↓ to navigate, ←/→ to collapse/expand providers, Space to select, CTRL+A to select/deselect all, Enter to continue, Esc to go back.",
		stageMachineDeploymentSettingsSelection: "↑/↓ to navigate, ←/→ to collapse/expand providers, Space to select, CTRL+A to select/deselect all, Enter to continue, Esc to go back.",
		stageClusterConfiguration:               "↑/↓ to navigate, ←/→ to collapse/expand categories, Space to edit/toggle, Enter to continue, Esc to go back.",
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

const successInstallationText = `Kubermatic Virtualization has been successfully installed! You can now proceed with configuring and using the virtualization features. If you encounter any issues, feel free to consult the documentation or support resources.`

const executionWarning = `Exiting now will interrupt the kubev bootstrap process. Any partial setup will remain on your servers, and simply rerunning the installer will not clean it up—you'll need to reset the servers manually before trying again.`
