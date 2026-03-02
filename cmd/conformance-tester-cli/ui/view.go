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

// previewString returns a compact preview of a potentially large string.
// It limits the output to `maxLines` lines and `maxChars` characters.
// The boolean return indicates whether the original was truncated.
func previewString(s string, maxLines, maxChars int) (string, bool) {
	if s == "" {
		return "", false
	}
	lines := strings.Split(s, "\n")
	truncated := false
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		truncated = true
	}
	joined := strings.Join(lines, "\n")
	if len(joined) > maxChars {
		joined = joined[:maxChars]
		truncated = true
	}
	return joined, truncated
}

// Color definitions for consistent theming.
const (
	colorMainBlue      = "#2196F3"
	colorMainWhite     = "#FFFFFF"
	colorErrorRed      = "#FF5252"
	colorWarningYellow = "#FFC107"
	colorDimGray       = "#666666"
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

// Preview thresholds for long credential values.
const (
	previewMaxLines = 6
	previewMaxChars = 400
	// truncationCharLimit: any single-line value longer than this
	// will be considered for truncation and moved to the inspect modal.
	truncationCharLimit = 150
)

// renderQuitConfirm draws the quit confirmation dialog.
func (m Model) renderQuitConfirm(uiWidth, uiInnerWidth int) string {
	boxHeight := m.getUIHeight()

	// Dynamic content based on stage
	titleContent := "Confirm Quit"
	warningContent := "Are you sure you want to quit? Unsaved progress will be lost."
	if m.stage == stageExecuting {
		if m.executionCancelling {
			// Already cancelling - don't show confirmation
			return ""
		}
		titleContent = "Cancel Test Execution"
		warningContent = "This will stop all running tests and clean up created resources (namespace, PVC, jobs, configmaps, secrets).\n\nAre you sure you want to cancel?"
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

// renderViewModal displays the full content of a credential inside a scrollable viewport.
func (m Model) renderViewModal(uiWidth, uiInnerWidth int) string {
	boxHeight := m.getUIHeight()

	title := lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, styleHeader.Render(m.viewModalTitle))

	// If a viewport exists (initialized when opening modal) use it, otherwise render raw content
	var content string
	if m.viewModalViewport.Width > 0 {
		content = m.viewModalViewport.View()
	} else {
		content = m.viewModalContent
	}

	var b strings.Builder
	b.WriteString(title + "\n\n")

	body := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Width(uiInnerWidth).Render(content)
	b.WriteString(body + "\n\n")

	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Render("Press Esc or Enter to close. Use Up/Down to scroll.")
	b.WriteString(hint)

	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")

	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody)
}

// renderWelcome displays the initial welcome screen.
func (m Model) renderWelcome(helpWithBorder string, uiWidth int) string {
	boxHeight := m.getUIHeight()
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

func (m Model) renderEnvironmentSelection(helpWithBorder string, uiWidth int) string {
	boxHeight := m.getUIHeight()
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
	localOption := fmt.Sprintf("%s Local Environment (WIP)", lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(localCheckbox))

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

					content := fmt.Sprintf("%s %s",
						lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(radioBtn),
						option.Path)

					if isFocused {
						content = styleFocusHighlight.Render(content)
					} else if option.Selected {
						content = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(content)
					} else {
						content = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(content)
					}

					b.WriteString("\t\t\t\t" + content + "\n")
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
					var content string

					// If selected, show filename and path together with path faint
					if option.Selected {
						pathStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Faint(true).Render(fmt.Sprintf(" → %s", option.Path))
						content = fmt.Sprintf("%s %s%s", radioBtnStyled, filename, pathStyled)
					} else {
						content = fmt.Sprintf("%s %s", radioBtnStyled, filename)
					}

					if isFocused {
						content = styleFocusHighlight.Render(content)
					} else if option.Selected {
						// For selected items, only bold the filename part, keep path faint
						filenameBold := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(filename)
						pathStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Faint(true).Render(fmt.Sprintf(" → %s", option.Path))
						content = fmt.Sprintf("%s %s%s", radioBtnStyled, filenameBold, pathStyled)
					} else {
						content = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(content)
					}

					b.WriteString("\t\t\t\t" + content + "\n")
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

					content := fmt.Sprintf("%s %s",
						lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(radioBtn), "Use custom path")

					if isFocused {
						content = styleFocusHighlight.Render(content)
					} else if option.Selected {
						content = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(content)
					} else {
						content = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(content)
					}

					b.WriteString("\t\t\t\t" + content + "\n")

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

		// Seeds Selection (fetched from cluster)
		b.WriteString(styleLabel.Render("Seed Name:") + "\n")
		if m.existingEnv.LoadingSeeds {
			loadingMsg := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Faint(true).Render("Loading seeds...")
			b.WriteString("\t\t\t\t\t\t" + loadingMsg + "\n")
		} else if m.existingEnv.FetchError != "" {
			inlineError := lipgloss.NewStyle().Foreground(lipgloss.Color(colorErrorRed)).Bold(true).Width(uiWidth - 27).Render(m.existingEnv.FetchError)
			b.WriteString("\t\t\t\t\t\t" + inlineError + "\n")
		} else if len(m.existingEnv.AvailableSeeds) == 0 {
			emptyMsg := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Faint(true).Render("No seeds found. Select a kubeconfig first.")
			b.WriteString("\t\t\t\t\t\t" + emptyMsg + "\n")
		} else {
			for i, seed := range m.existingEnv.AvailableSeeds {
				radioBtn := "( )"
				if i == m.existingEnv.SelectedSeedIndex {
					radioBtn = "(•)"
				}

				content := fmt.Sprintf("%s %s", radioBtn, seed)
				isFocused := m.environmentFocusIndex == 1 && m.environmentFieldIndex == 2 && i == m.existingEnv.SeedFocusedIndex
				if isFocused {
					content = styleFocusHighlight.Render(content)
				}
				seedLine := "\t\t\t\t\t\t" + content
				b.WriteString(seedLine + "\n")
			}
		}
		if m.existingEnv.Errors.SeedName != "" {
			inlineError := lipgloss.NewStyle().Foreground(lipgloss.Color(colorErrorRed)).Bold(true).Width(uiWidth - 27).Render(m.existingEnv.Errors.SeedName)
			b.WriteString("\t\t\t\t\t\t" + inlineError + "\n")
		}
		if err, ok := m.existingEnv.Errors.Fields["SeedName"]; ok && err != "" {
			inlineError := lipgloss.NewStyle().Foreground(lipgloss.Color(colorErrorRed)).Bold(true).Width(uiWidth - 27).Render(err)
			b.WriteString("\t\t\t\t\t\t" + inlineError + "\n")
		}
		b.WriteString("\n")

		// Presets Selection (fetched from cluster)
		b.WriteString(styleLabel.Render("Preset Name:") + "\n")
		if m.existingEnv.LoadingPresets {
			loadingMsg := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Faint(true).Render("Loading presets...")
			b.WriteString("\t\t\t\t\t\t" + loadingMsg + "\n")
		} else if m.existingEnv.FetchError != "" {
			inlineError := lipgloss.NewStyle().Foreground(lipgloss.Color(colorErrorRed)).Bold(true).Width(uiWidth - 27).Render(m.existingEnv.FetchError)
			b.WriteString("\t\t\t\t\t\t" + inlineError + "\n")
		} else if len(m.existingEnv.AvailablePresets) == 0 {
			emptyMsg := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Faint(true).Render("No presets found. Select a kubeconfig first.")
			b.WriteString("\t\t\t\t\t\t" + emptyMsg + "\n")
		} else {
			for i, preset := range m.existingEnv.AvailablePresets {
				radioBtn := "( )"
				if i == m.existingEnv.SelectedPresetIndex {
					radioBtn = "(•)"
				}

				content := fmt.Sprintf("%s %s", radioBtn, preset)
				isFocused := m.environmentFocusIndex == 1 && m.environmentFieldIndex == 3 && i == m.existingEnv.PresetFocusedIndex
				if isFocused {
					content = styleFocusHighlight.Render(content)
				}
				presetLine := "\t\t\t\t\t\t" + content
				b.WriteString(presetLine + "\n")
			}
		}
		if m.existingEnv.Errors.PresetName != "" {
			inlineError := lipgloss.NewStyle().Foreground(lipgloss.Color(colorErrorRed)).Bold(true).Width(uiWidth - 27).Render(m.existingEnv.Errors.PresetName)
			b.WriteString("\t\t\t\t\t\t" + inlineError + "\n")
		}
		if err, ok := m.existingEnv.Errors.Fields["PresetName"]; ok && err != "" {
			inlineError := lipgloss.NewStyle().Foreground(lipgloss.Color(colorErrorRed)).Bold(true).Width(uiWidth - 27).Render(err)
			b.WriteString("\t\t\t\t\t\t" + inlineError + "\n")
		}
		b.WriteString("\n")

		// Project Name (text input field)
		projectLine := lipgloss.JoinHorizontal(
			lipgloss.Left,
			styleLabel.Render("Project Name:"),
			" ",
			styleInput.Render(m.existingEnv.ProjectName.View()),
		)
		b.WriteString(projectLine + "\n")
		if m.existingEnv.Errors.ProjectName != "" {
			inlineError := lipgloss.NewStyle().Foreground(lipgloss.Color(colorErrorRed)).Bold(true).Width(uiWidth - 27).Render(m.existingEnv.Errors.ProjectName)
			b.WriteString("\t\t\t\t\t\t" + inlineError + "\n")
		}
		b.WriteString("\n")
	}

	// Pad content to ensure help bar is at the bottom
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")
	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
}

func (m Model) renderReleaseSelection(helpWithBorder string, uiWidth int) string {
	boxHeight := m.getUIHeight()
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

func (m Model) renderProviderSelection(helpWithBorder string, uiWidth int) string {
	boxHeight := m.getUIHeight()
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

		// If provider is selected and expanded, show credential fields
		if i == m.expandedProviderIndex && provider.Selected {
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

	// If provider has preset credentials, show credential source selector
	if provider.HasPresetCredentials {
		// Radio button for "From Preset"
		presetRadio := "( )"
		if provider.CredentialSource == CredentialSourcePreset {
			presetRadio = "(•)"
		}

		presetLine := fmt.Sprintf("  %s From Preset",
			lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(presetRadio))

		isFocused := providerIndex == m.providerFocusIndex && m.providerFieldIndex == 1
		if isFocused {
			presetLine = styleFocusHighlight.Render(presetLine)
		} else if provider.CredentialSource == CredentialSourcePreset {
			presetLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(presetLine)
		} else {
			presetLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(presetLine)
		}
		b.WriteString(presetLine + "\n")

		// Show preset credential values (masked/read-only)
		if provider.CredentialSource == CredentialSourcePreset {
			b.WriteString(m.renderPresetCredentialsSummary(provider))
		}

		// Radio button for "Enter Custom Credentials"
		customRadio := "( )"
		if provider.CredentialSource == CredentialSourceCustom {
			customRadio = "(•)"
		}

		customLine := fmt.Sprintf("  %s Enter Custom Credentials",
			lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(customRadio))

		customFocused := providerIndex == m.providerFocusIndex && m.providerFieldIndex == 2
		if customFocused {
			customLine = styleFocusHighlight.Render(customLine)
		} else if provider.CredentialSource == CredentialSourceCustom {
			customLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(customLine)
		} else {
			customLine = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(customLine)
		}
		b.WriteString(customLine + "\n\n")
	}

	// Show custom credential fields if selected or if no preset available
	if !provider.HasPresetCredentials || provider.CredentialSource == CredentialSourceCustom {
		fieldOffset := 3 // Offset for field index if preset is available
		if !provider.HasPresetCredentials {
			fieldOffset = 1
		}

		renderField := func(label string, input textinput.Model, error string, fieldIndex int) {
			actualFieldIndex := fieldOffset + fieldIndex - 1

			val := input.Value()
			// If the value is longer than truncationCharLimit, show a truncated preview
			if len(val) > truncationCharLimit {
				preview, _ := previewString(val, previewMaxLines, truncationCharLimit)
				if providerIndex == m.providerFocusIndex && actualFieldIndex == m.providerFieldIndex {
					b.WriteString(styleFocusHighlight.Render("    "+label) + " " + lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Faint(true).Render(preview+" ...") + "  [i]\n")
				} else {
					b.WriteString("    " + label + " " + lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Faint(true).Render(preview+" ...") + "  [i]\n")
				}
			} else {
				if providerIndex == m.providerFocusIndex && actualFieldIndex == m.providerFieldIndex {
					b.WriteString(styleFocusHighlight.Render("    "+label) + " " + input.View() + "\n")
				} else {
					b.WriteString("    " + label + " " + input.View() + "\n")
				}
			}

			if error != "" {
				errorMsg := lipgloss.NewStyle().Foreground(lipgloss.Color(colorErrorRed)).Bold(true).Width(uiWidth - 27).Render(error)
				b.WriteString("      " + errorMsg + "\n")
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
	}

	return b.String()
}

// renderPresetCredentialsSummary shows a summary of preset credentials (masked).
func (m Model) renderPresetCredentialsSummary(provider Provider) string {
	var b strings.Builder

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMainWhite)).
		Faint(true)
	// helper to render a label/value, truncating long values and appending the (i) hint
	renderVal := func(label, val string) {
		if val == "" {
			return
		}
		if preview, truncated := previewString(val, previewMaxLines, truncationCharLimit); truncated {
			b.WriteString(style.Render("      "+label+" "+preview+" ...") + "  [i]\n")
		} else {
			b.WriteString(style.Render("      "+label+" "+val) + "\n")
		}
	}

	switch creds := provider.PresetCredentials.(type) {
	case AWSCredentials:
		renderVal("Access Key ID:", creds.AccessKeyID.Value())
		renderVal("Secret Access Key:", creds.SecretAccessKey.Value())
		renderVal("Assume Role ARN:", creds.AssumeRoleARN.Value())
		renderVal("External ID:", creds.AssumeRoleExternalID.Value())

	case AzureCredentials:
		renderVal("Tenant ID:", creds.TenantID.Value())
		renderVal("Subscription ID:", creds.SubscriptionID.Value())
		renderVal("Client ID:", creds.ClientID.Value())
		renderVal("Client Secret:", creds.ClientSecret.Value())

	case GCPCredentials:
		renderVal("Service Account:", creds.ServiceAccount.Value())

	case AlibabaCredentials:
		renderVal("Access Key ID:", creds.AccessKeyID.Value())
		renderVal("Access Key Secret:", creds.AccessKeySecret.Value())

	case AnexiaCredentials:
		renderVal("API Token:", creds.Token.Value())

	case DigitalOceanCredentials:
		renderVal("API Token:", creds.Token.Value())

	case HetznerCredentials:
		renderVal("API Token:", creds.Token.Value())

	case KubeVirtCredentials:
		renderVal("Kubeconfig:", creds.Kubeconfig.Value())

	case NutanixCredentials:
		renderVal("Username:", creds.Username.Value())
		renderVal("Password:", creds.Password.Value())
		renderVal("Cluster Name:", creds.ClusterName.Value())
		renderVal("Proxy URL:", creds.ProxyURL.Value())
		renderVal("CSI Username:", creds.CSIUsername.Value())

	case OpenStackCredentials:
		renderVal("Username:", creds.Username.Value())
		renderVal("Password:", creds.Password.Value())
		renderVal("Project:", creds.Project.Value())
		renderVal("Domain:", creds.Domain.Value())

	case VSphereCredentials:
		renderVal("Username:", creds.Username.Value())
		renderVal("Password:", creds.Password.Value())

	case VMwareCloudDirectorCredentials:
		renderVal("Username:", creds.Username.Value())
		renderVal("Password:", creds.Password.Value())
		renderVal("API Token:", creds.APIToken.Value())
		renderVal("Organization:", creds.Organization.Value())
		renderVal("VDC:", creds.VDC.Value())
	}

	b.WriteString("\n")
	return b.String()
}

func (m Model) renderDistributionSelection(helpWithBorder string, uiWidth int) string {
	boxHeight := m.getUIHeight()
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

// settingsTreeRow represents one row in the flattened settings tree for pagination.
type settingsTreeRow struct {
	rowType   string // "provider", "group", "option", "spacing"
	provider  string
	groupIdx  int
	optionIdx int
}

// buildSettingsTreeRows flattens the provider → group → options hierarchy into an ordered
// slice of rows so we can paginate and render only the visible window.
func buildSettingsTreeRows(
	providers []string,
	settingsByProvider map[string][]SettingGroup,
	expandedProviders map[string]bool,
	multiProvider bool,
) []settingsTreeRow {
	var rows []settingsTreeRow
	for providerIdx, provider := range providers {
		if multiProvider {
			rows = append(rows, settingsTreeRow{rowType: "provider", provider: provider, groupIdx: -1, optionIdx: -1})
		}

		isExpanded := expandedProviders[provider]
		if isExpanded || !multiProvider {
			groups := settingsByProvider[provider]
			for gi, group := range groups {
				rows = append(rows, settingsTreeRow{rowType: "group", provider: provider, groupIdx: gi, optionIdx: -1})
				for oi := range group.Options {
					rows = append(rows, settingsTreeRow{rowType: "option", provider: provider, groupIdx: gi, optionIdx: oi})
				}
			}
		}

		if multiProvider && providerIdx < len(providers)-1 {
			rows = append(rows, settingsTreeRow{rowType: "spacing"})
		}
	}
	return rows
}

// renderSettingsTreeRow renders a single row in the settings tree.
func (m Model) renderSettingsTreeRow(
	row settingsTreeRow,
	rowAbsIndex int,
	focusedIndex int,
	settingsByProvider map[string][]SettingGroup,
	expandedProviders map[string]bool,
	selected map[string]bool,
	selectedGroups map[string]bool,
	multiProvider bool,
) string {
	isFocused := rowAbsIndex == focusedIndex

	switch row.rowType {
	case "provider":
		isExpanded := expandedProviders[row.provider]
		expandIndicator := "▶"
		if isExpanded {
			expandIndicator = "▼"
		}
		providerHeader := fmt.Sprintf("%s %s", expandIndicator, row.provider)
		if isFocused {
			return styleFocusHighlight.Render(providerHeader)
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(providerHeader)

	case "group":
		groups := settingsByProvider[row.provider]
		if row.groupIdx < 0 || row.groupIdx >= len(groups) {
			return ""
		}
		group := groups[row.groupIdx]
		groupKey := fmt.Sprintf("%s:%s", row.provider, group.Key)
		isGroupSelected := selectedGroups[groupKey]

		checkbox := "[ ]"
		if isGroupSelected {
			checkbox = "[✓]"
		}

		groupHeader := fmt.Sprintf("  %s %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(checkbox),
			group.Name)
		if isFocused {
			return styleFocusHighlight.Render(groupHeader)
		} else if isGroupSelected {
			return lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(groupHeader)
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(groupHeader)

	case "option":
		groups := settingsByProvider[row.provider]
		if row.groupIdx < 0 || row.groupIdx >= len(groups) {
			return ""
		}
		group := groups[row.groupIdx]
		if row.optionIdx < 0 || row.optionIdx >= len(group.Options) {
			return ""
		}
		option := group.Options[row.optionIdx]
		selectionKey := fmt.Sprintf("%s:%s:%s", row.provider, group.Key, option)
		isSelected := selected[selectionKey]

		checkbox := "[ ]"
		if isSelected {
			checkbox = "[✓]"
		}

		optionLine := fmt.Sprintf("    %s %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(checkbox),
			option)
		if isFocused {
			return styleFocusHighlight.Render(optionLine)
		} else if isSelected {
			return lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Bold(true).Render(optionLine)
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainWhite)).Render(optionLine)

	case "spacing":
		return ""
	}
	return ""
}

// renderPaginatedSettingsBody renders the paginated settings tree body (the Provider → Group → Options section).
// It updates the viewport's page size dynamically and returns the rendered string (without title/description/help).
func (m Model) renderPaginatedSettingsBody(
	b *strings.Builder,
	providers []string,
	settingsByProvider map[string][]SettingGroup,
	expandedProviders map[string]bool,
	selected map[string]bool,
	selectedGroups map[string]bool,
	focusedIndex int,
	vp *SettingsViewport,
) {
	multiProvider := len(providers) > 1

	// Build all rows
	rows := buildSettingsTreeRows(providers, settingsByProvider, expandedProviders, multiProvider)
	totalRows := len(rows)

	if totalRows == 0 {
		return
	}

	// Reserve lines: title(1) + blank(1) + description(1) + blank(1) + scroll-up(1) + scroll-down(1) + position(1) + trailing-newline(1) + box-border(2) = 10
	const reservedLines = 10
	vp.updatePageSize(m.getUIHeight(), reservedLines)
	vp.ensureFocusVisible(focusedIndex)

	showUp, showDown := vp.scrollIndicators(totalRows)

	if showUp {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorDimGray)).Render("  ▲ more above") + "\n")
	}

	start, end := vp.visibleRange(totalRows)

	// We need to map visual row index back to the absolute index in the full tree.
	// The absolute index (used for focus) counts only provider/group/option rows, not spacing.
	// Build a map of row index → absolute focus index.
	absIndex := 0
	absIndices := make([]int, totalRows)
	for i, row := range rows {
		if row.rowType == "spacing" {
			absIndices[i] = -1 // spacing rows don't have a focus index
		} else {
			absIndices[i] = absIndex
			absIndex++
		}
	}

	for i := start; i < end; i++ {
		row := rows[i]
		if row.rowType == "spacing" {
			b.WriteString("\n")
			continue
		}
		line := m.renderSettingsTreeRow(row, absIndices[i], focusedIndex, settingsByProvider, expandedProviders, selected, selectedGroups, multiProvider)
		b.WriteString(line + "\n")
	}

	if showDown {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorDimGray)).Render("  ▼ more below") + "\n")
	}

	// Position indicator
	pos := focusedIndex + 1
	total := absIndex // total focusable items
	if total > 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorDimGray)).Render(fmt.Sprintf("  %d/%d", pos, total)) + "\n")
	}
}

// renderDatacenterSettingsSelection renders the datacenter settings selection stage.
func (m Model) renderDatacenterSettingsSelection(helpWithBorder string, uiWidth int) string {
	boxHeight := m.getUIHeight()

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

	// Check if any provider is loading or has errors
	anyLoading := false
	var loadingProviders []string
	var errorProviders []string
	for _, provider := range m.datacenterSettingsSelection.Providers {
		if ps, ok := m.datacenterSettingsSelection.ProviderSettings[provider]; ok {
			if ps.LoadingSettings {
				anyLoading = true
				loadingProviders = append(loadingProviders, provider)
			}
			if ps.SettingsFetchError != "" {
				errorProviders = append(errorProviders, provider)
			}
		}
	}

	// Show loading indicators
	if anyLoading {
		loadingMsg := "⏳ Loading settings"
		if len(loadingProviders) > 0 {
			loadingMsg += " for " + strings.Join(loadingProviders, ", ")
		}
		loadingMsg += "..."
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(loadingMsg))
		b.WriteString("\n\n")
		lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
		contentBody := strings.Join(lines, "\n")
		return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
	}

	// Show errors if any
	if len(errorProviders) > 0 {
		for _, provider := range errorProviders {
			if ps, ok := m.datacenterSettingsSelection.ProviderSettings[provider]; ok {
				errorMsg := fmt.Sprintf("⚠ Error loading settings for %s: %s", provider, ps.SettingsFetchError)
				b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorErrorRed)).Render(errorMsg))
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}

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

	// Render paginated hierarchical structure: Provider → Group → Options
	m.renderPaginatedSettingsBody(
		&b,
		m.datacenterSettingsSelection.Providers,
		m.datacenterSettingsSelection.SettingsByProvider,
		m.datacenterSettingsSelection.ExpandedProviders,
		m.datacenterSettingsSelection.Selected,
		m.datacenterSettingsSelection.SelectedGroups,
		m.datacenterSettingsSelection.FocusedIndex,
		&m.datacenterViewport,
	)

	// Pad content to ensure help bar is at the bottom
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")
	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
}

// renderClusterSettingsSelection renders the cluster settings selection stage.
func (m Model) renderClusterSettingsSelection(helpWithBorder string, uiWidth int) string {
	boxHeight := m.getUIHeight()

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

	// Render paginated hierarchical structure: Provider → Group → Options
	m.renderPaginatedSettingsBody(
		&b,
		m.clusterSettingsSelection.Providers,
		m.clusterSettingsSelection.SettingsByProvider,
		m.clusterSettingsSelection.ExpandedProviders,
		m.clusterSettingsSelection.Selected,
		m.clusterSettingsSelection.SelectedGroups,
		m.clusterSettingsSelection.FocusedIndex,
		&m.clusterViewport,
	)

	// Pad content to ensure help bar is at the bottom
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")
	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
}

// renderMachineDeploymentSettingsSelection renders the machine deployment settings selection stage.
func (m Model) renderMachineDeploymentSettingsSelection(helpWithBorder string, uiWidth int) string {
	boxHeight := m.getUIHeight()

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

	// Check if any provider is loading or has errors
	anyLoading := false
	var loadingProviders []string
	var errorProviders []string
	for _, provider := range m.machineDeploymentSettingsSelection.Providers {
		if ps, ok := m.machineDeploymentSettingsSelection.ProviderSettings[provider]; ok {
			if ps.LoadingSettings {
				anyLoading = true
				loadingProviders = append(loadingProviders, provider)
			}
			if ps.SettingsFetchError != "" {
				errorProviders = append(errorProviders, provider)
			}
		}
	}

	// Show loading indicators
	if anyLoading {
		loadingMsg := "⏳ Loading settings"
		if len(loadingProviders) > 0 {
			loadingMsg += " for " + strings.Join(loadingProviders, ", ")
		}
		loadingMsg += "..."
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorMainBlue)).Render(loadingMsg))
		b.WriteString("\n\n")
		lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
		contentBody := strings.Join(lines, "\n")
		return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
	}

	// Show errors if any
	if len(errorProviders) > 0 {
		for _, provider := range errorProviders {
			if ps, ok := m.machineDeploymentSettingsSelection.ProviderSettings[provider]; ok {
				errorMsg := fmt.Sprintf("⚠ Error loading settings for %s: %s", provider, ps.SettingsFetchError)
				b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorErrorRed)).Render(errorMsg))
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}

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

	// Render paginated hierarchical structure: Provider → Group → Options
	m.renderPaginatedSettingsBody(
		&b,
		m.machineDeploymentSettingsSelection.Providers,
		m.machineDeploymentSettingsSelection.SettingsByProvider,
		m.machineDeploymentSettingsSelection.ExpandedProviders,
		m.machineDeploymentSettingsSelection.Selected,
		m.machineDeploymentSettingsSelection.SelectedGroups,
		m.machineDeploymentSettingsSelection.FocusedIndex,
		&m.machineViewport,
	)

	// Pad content to ensure help bar is at the bottom
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")
	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody + "\n" + helpWithBorder)
}

// renderClusterConfiguration renders the cluster configuration stage.
func (m Model) renderClusterConfiguration(helpWithBorder string, uiWidth int) string {
	boxHeight := m.getUIHeight()

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

// renderReviewSettings displays the configuration review stage.
func (m Model) renderReviewSettings(helpWithBorder string, uiWidth int) string {
	boxHeight := m.getUIHeight()

	header := styleHeader.Render("Review Configuration")
	description := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMainWhite)).
		Render("Review your configuration before proceeding with test execution. Use ←/→ or Space to expand/collapse sections.")

	// Generate YAML sections per provider
	providerReviews := m.generateReviewYAML()

	var content strings.Builder
	currentIndex := 0

	for _, providerReview := range providerReviews {
		// Provider header
		isProviderFocused := currentIndex == m.Review.FocusedIndex
		isProviderExpanded := m.Review.ExpandedProviders[providerReview.ProviderName]

		providerIndicator := "▼"
		if !isProviderExpanded {
			providerIndicator = "▶"
		}

		providerHeader := fmt.Sprintf("%s %s", providerIndicator, providerReview.ProviderName)

		if isProviderFocused {
			content.WriteString(styleFocusHighlight.Render(providerHeader) + "\n")
		} else {
			content.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorMainBlue)).
				Bold(true).
				Render(providerHeader) + "\n")
		}
		currentIndex++

		// Show sections if provider is expanded
		if isProviderExpanded {
			for _, section := range providerReview.Sections {
				isSectionFocused := currentIndex == m.Review.FocusedIndex
				sectionKey := fmt.Sprintf("%s:%s", providerReview.ProviderName, section.Name)
				isSectionExpanded := m.Review.ExpandedSections[sectionKey]

				// Section header with expand/collapse indicator
				sectionIndicator := "▼"
				if !isSectionExpanded {
					sectionIndicator = "▶"
				}

				sectionHeader := fmt.Sprintf("  %s %s", sectionIndicator, section.Name)

				if isSectionFocused {
					content.WriteString(styleFocusHighlight.Render(sectionHeader) + "\n")
				} else {
					content.WriteString(styleItem.Render(sectionHeader) + "\n")
				}
				currentIndex++

				// Show content if section is expanded
				if isSectionExpanded {
					// Indent the content (4 spaces total - 2 for provider + 2 for content)
					lines := strings.Split(section.Content, "\n")
					for _, line := range lines {
						if line != "" {
							content.WriteString("    " + line + "\n")
						}
					}
				}
			}
		}

		content.WriteString("\n")
	}

	// Add save to file checkbox
	isSaveCheckboxFocused := currentIndex == m.Review.FocusedIndex
	checkbox := "[ ]"
	if m.Review.SaveToFile {
		checkbox = "[✓]"
	}
	saveOption := fmt.Sprintf("%s Save configurations to files", checkbox)
	if isSaveCheckboxFocused {
		content.WriteString(styleFocusHighlight.Render(saveOption) + "\n")
	} else {
		content.WriteString(styleItem.Render(saveOption) + "\n")
	}

	var b strings.Builder
	b.WriteString(lipgloss.PlaceHorizontal(uiWidth, lipgloss.Center, header) + "\n\n")
	b.WriteString(description + "\n\n")
	b.WriteString(content.String())

	// Pad content to ensure help bar is at the bottom
	lines := padContentToHeight(b.String(), boxHeight-uiBoxHeightPad)
	contentBody := strings.Join(lines, "\n")
	return styleBox.Width(uiWidth).Height(boxHeight).Render(contentBody)
}

// renderExecuting displays logs during the configuration application process.
func (m Model) renderExecuting(helpWithBorder string, uiWidth int) string {
	boxHeight := m.getUIHeight()
	var header string
	if m.executionCancelling {
		header = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorWarningYellow)).
			Bold(true).
			Render("⚠ Cancelling Test Execution and Cleaning Up...")
	} else {
		header = styleHeader.Render("Applying Configuration")
	}

	// Build the main content
	var b strings.Builder
	b.WriteString(header + "\n\n") // Extra line after header
	b.WriteString(m.Review.Viewport.View())

	if m.executionError != "" && !m.executionCancelling {
		b.WriteString("\n" + styleError.Render(m.executionError))
	}

	return styleBox.Width(uiWidth).Height(boxHeight).Render(b.String())
}

// renderDone displays the final success message.
func (m Model) renderDone(helpWithBorder string, uiWidth int) string {
	boxHeight := m.getUIHeight()
	header := styleHeader.Render("Congratulations!")
	var message string
	if m.executionError != "" {
		header = styleHeader.Render("Execution Finished With Errors")
		message = styleError.Render(m.executionError)
	} else {
		message = styleItem.Render(successTestsText)
	}

	// Build the main content
	var b strings.Builder
	b.WriteString(header + "\n\n") // Extra line after header
	b.WriteString(message)

	return styleBox.Width(uiWidth).Height(boxHeight).Render(b.String())
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
		stageProviderSelection:                  "↑/↓ to navigate, Space to select, I/i for Detailed View, Enter to continue, Esc to go back.",
		stageDistributionSelection:              "↑/↓ to navigate, Space to select, CTRL+A to select/deselect all, Enter to continue, Esc to go back.",
		stageDatacenterSettingsSelection:        "↑/↓ to navigate, ←/→ to collapse/expand providers, Space to select, CTRL+A to select/deselect all, Enter to continue, Esc to go back.",
		stageClusterSettingsSelection:           "↑/↓ to navigate, ←/→ to collapse/expand providers, Space to select, CTRL+A to select/deselect all, Enter to continue, Esc to go back.",
		stageMachineDeploymentSettingsSelection: "↑/↓ to navigate, ←/→ to collapse/expand providers, Space to select, CTRL+A to select/deselect all, Enter to continue, Esc to go back.",
		stageClusterConfiguration:               "↑/↓ to navigate, ←/→ to collapse/expand categories, Space to edit/toggle, Enter to continue, Esc to go back.",
		stageReviewSettings:                     "↑/↓: Navigate • Space/←/→: Expand/Collapse • Enter: Execute • Esc: Back",
		stageExecuting:                          "Logs will appear here. Press ctrl+c to cancel.",
		stageDone:                               "Press q to quit.",
	}
	// Return empty string for unknown stages
	return helpTexts[stage]
}

// --- Static Text Content ---

const welcomeTitleText = "Welcome to Kubermatic Conformance Tester"

const welcomeDisclaimerText = `Kubermatic Conformance Tester is an automated testing tool that validates Kubernetes cluster functionality and compliance across multiple cloud providers. It performs comprehensive conformance tests including storage, networking, load balancing, and security context validation to ensure your clusters meet production standards.`

const environmentSelectionText = `Choose whether you want to run conformance tester using an already existing Kubernetes cluster or set up a new KKP instance locally.`

const releaseSelectionText = `Select the supported kubernetes release version for your cluster. Kubermatic Conformance Tester ensures compatibility and performance across different Kubernetes versions by providing options for various releases.`

const providerSelectionText = `Select the infrastructure provider where your Kubernetes cluster will be deployed. Kubermatic Conformance Tester supports multiple providers to ensure compatibility and performance across different environments.`

const distributionSelectionText = `Select one or more operating system distributions for your cluster nodes. Multiple distributions can be selected to test compatibility across different OS environments.`

const successTestsText = "All tests executed successfully! Your cluster is compliant with the selected configurations."
