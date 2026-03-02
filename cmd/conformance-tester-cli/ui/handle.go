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
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	keyEsc         = "esc"
	keyEnter       = "enter"
	keyRight       = "right"
	keyLeft        = "left"
	keyUp          = "up"
	keyDown        = "down"
	keyTab         = "tab"
	keyShiftTab    = "shift+tab"
	keyYes         = "y"
	keyNo          = "n"
	keyControlC    = "ctrl+c"
	keyQuit        = "q"
	keyI           = "i"
	keySpace       = " "
	keySelectAll   = "ctrl+a"
	mouseWheelUp   = "wheel up"
	mouseWheelDown = "wheel down"
	digits         = "0123456789"
)

// ----------------------------------- Stage 0: Welcome -----------------------------------

func (m *Model) handleWelcomePage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyEnter:
		m.stage = stageEnvironmentSelection
		return m, nil
	}
	return m, nil
}

// ----------------------------------- Stage 1: Environment Selection -----------------------------------

func (m *Model) handleEnvironmentSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case keyUp:
		if m.environmentFocusIndex == 1 && m.existingEnv.Selected && m.environmentFieldIndex == 1 {
			// Navigate within kubeconfig options
			if m.existingEnv.KubeconfigFocusedIndex > 0 {
				m.existingEnv.KubeconfigFocusedIndex--
			} else {
				m.environmentFieldIndex = 0 // Move to checkbox
			}
		} else if m.environmentFocusIndex == 1 && m.existingEnv.Selected && m.environmentFieldIndex == 2 {
			// Navigate within Seeds list
			if m.existingEnv.SeedFocusedIndex > 0 {
				m.existingEnv.SeedFocusedIndex--
			} else {
				m.environmentFieldIndex-- // Move to previous field
			}
		} else if m.environmentFocusIndex == 1 && m.existingEnv.Selected && m.environmentFieldIndex == 3 {
			// Navigate within Presets list
			if m.existingEnv.PresetFocusedIndex > 0 {
				m.existingEnv.PresetFocusedIndex--
			} else {
				m.environmentFieldIndex-- // Move to previous field
			}
		} else if m.environmentFocusIndex == 1 && m.existingEnv.Selected && m.environmentFieldIndex > 0 {
			m.environmentFieldIndex--
		} else if m.environmentFocusIndex == 0 && m.localEnv.Selected && m.environmentFieldIndex > 0 {
			m.environmentFieldIndex--
		} else if m.environmentFieldIndex == 0 && m.environmentFocusIndex > 0 {
			m.environmentFocusIndex--
		}
		m.updateEnvironmentFocus()
		return m, nil

	case keyDown:
		if m.environmentFocusIndex == 0 && m.localEnv.Selected {
			if m.environmentFieldIndex < 3 {
				m.environmentFieldIndex++
			} else {
				m.environmentFocusIndex++
				m.environmentFieldIndex = 0
			}
		} else if m.environmentFocusIndex == 1 && m.existingEnv.Selected {
			if m.environmentFieldIndex == 0 {
				m.environmentFieldIndex = 1 // Move to kubeconfig selection
			} else if m.environmentFieldIndex == 1 {
				// In kubeconfig selection
				maxVisualIndex := m.getMaxKubeconfigVisualIndex()
				if m.existingEnv.KubeconfigFocusedIndex < maxVisualIndex {
					m.existingEnv.KubeconfigFocusedIndex++
				} else {
					m.environmentFieldIndex++ // Move to next field (Seeds)
				}
			} else if m.environmentFieldIndex == 2 {
				// In Seeds selection
				if len(m.existingEnv.AvailableSeeds) > 0 && m.existingEnv.SeedFocusedIndex < len(m.existingEnv.AvailableSeeds)-1 {
					m.existingEnv.SeedFocusedIndex++
				} else {
					m.environmentFieldIndex++ // Move to Presets
				}
			} else if m.environmentFieldIndex == 3 {
				// In Presets selection
				if len(m.existingEnv.AvailablePresets) > 0 && m.existingEnv.PresetFocusedIndex < len(m.existingEnv.AvailablePresets)-1 {
					m.existingEnv.PresetFocusedIndex++
				} else {
					m.environmentFieldIndex++ // Move to Project Name
				}
			} else if m.environmentFieldIndex < 4 {
				m.environmentFieldIndex++
			}
		} else if m.environmentFieldIndex == 0 {
			m.environmentFocusIndex++
		}
		m.updateEnvironmentFocus()
		return m, nil

	case keyTab:
		// Tab moves forward, same as down but without wrapping
		if m.environmentFocusIndex == 0 && m.localEnv.Selected {
			if m.environmentFieldIndex < 3 {
				m.environmentFieldIndex++
			}
		} else if m.environmentFocusIndex == 1 && m.existingEnv.Selected {
			if m.environmentFieldIndex < 4 {
				m.environmentFieldIndex++
			}
		}
		m.updateEnvironmentFocus()
		return m, nil

	case keyShiftTab:
		if m.environmentFieldIndex > 0 {
			m.environmentFieldIndex--
		}
		m.updateEnvironmentFocus()
		return m, nil

	case keyEnter:
		// Validate and proceed if an environment is selected
		if m.localEnv.Selected && m.validateLocalEnvironment() {
			m.stage = stageReleaseSelection
		} else if m.existingEnv.Selected && m.validateExistingEnvironment() {
			m.stage = stageReleaseSelection
		}
		return m, nil

	case keySpace:
		if m.environmentFieldIndex == 0 {
			// Toggle environment checkbox
			if m.environmentFocusIndex == 0 {
				m.localEnv.Selected = !m.localEnv.Selected
				if m.localEnv.Selected {
					m.existingEnv.Selected = false
					m.environmentFieldIndex = 1
				}
			} else if m.environmentFocusIndex == 1 {
				m.existingEnv.Selected = !m.existingEnv.Selected
				if m.existingEnv.Selected {
					m.localEnv.Selected = false
					m.environmentFieldIndex = 1
				}
			}
		} else if m.environmentFocusIndex == 1 && m.environmentFieldIndex == 1 {
			// Check if we're on a section header
			sectionType := m.getKubeconfigSectionAtIndex(m.existingEnv.KubeconfigFocusedIndex)
			if sectionType != "" {
				// Don't toggle selection on section headers with space
				return m, nil
			}

			// Get the actual option index from the visual index
			optionIndex := m.getKubeconfigOptionIndexFromVisualIndex(m.existingEnv.KubeconfigFocusedIndex)
			if optionIndex >= 0 && optionIndex < len(m.existingEnv.KubeconfigOptions) {
				// Toggle kubeconfig option selection
				for i := range m.existingEnv.KubeconfigOptions {
					m.existingEnv.KubeconfigOptions[i].Selected = (i == optionIndex)
				}

				// If custom is selected, focus the custom input
				if m.existingEnv.KubeconfigOptions[optionIndex].Type == "custom" {
					m.existingEnv.CustomKubeconfigPath.Focus()
				} else {
					m.existingEnv.CustomKubeconfigPath.Blur()
				}

				// Trigger Seeds and Presets fetching
				m.existingEnv.LoadingSeeds = true
				m.existingEnv.LoadingPresets = true
				m.existingEnv.FetchError = ""
				cmd = m.fetchSeedsAndPresets()
			}
		} else if m.environmentFocusIndex == 1 && m.environmentFieldIndex == 2 {
			// Select Seed
			if len(m.existingEnv.AvailableSeeds) > 0 {
				m.existingEnv.SelectedSeedIndex = m.existingEnv.SeedFocusedIndex
			}
		} else if m.environmentFocusIndex == 1 && m.environmentFieldIndex == 3 {
			// Select Preset
			if len(m.existingEnv.AvailablePresets) > 0 {
				m.existingEnv.SelectedPresetIndex = m.existingEnv.PresetFocusedIndex
				// Fetch preset details to populate provider credentials
				cmd = m.fetchPresetDetails()
			}
		}
		m.updateEnvironmentFocus()
		return m, cmd

	case keyLeft:
		// Collapse kubeconfig section if on section header
		if m.environmentFocusIndex == 1 && m.environmentFieldIndex == 1 {
			sectionType := m.getKubeconfigSectionAtIndex(m.existingEnv.KubeconfigFocusedIndex)
			if sectionType != "" {
				m.existingEnv.KubeconfigExpandedSections[sectionType] = false
			}
		}
		return m, nil

	case keyRight:
		// Expand kubeconfig section if on section header
		if m.environmentFocusIndex == 1 && m.environmentFieldIndex == 1 {
			sectionType := m.getKubeconfigSectionAtIndex(m.existingEnv.KubeconfigFocusedIndex)
			if sectionType != "" {
				m.existingEnv.KubeconfigExpandedSections[sectionType] = true
			}
		}
		return m, nil

	case keyEsc:
		m.stage = stageWelcome
		return m, nil

	default:
		// Update the focused text input
		if m.environmentFocusIndex == 0 && m.localEnv.Selected && m.environmentFieldIndex > 0 {
			switch m.environmentFieldIndex {
			case 1:
				m.localEnv.KubermaticConfigurationsPath, cmd = m.localEnv.KubermaticConfigurationsPath.Update(msg)
			case 2:
				m.localEnv.HelmValuesPath, cmd = m.localEnv.HelmValuesPath.Update(msg)
			case 3:
				m.localEnv.MLAValuesPath, cmd = m.localEnv.MLAValuesPath.Update(msg)
			}
		} else if m.environmentFocusIndex == 1 && m.existingEnv.Selected {
			if m.environmentFieldIndex == 1 {
				// Handle custom kubeconfig path input
				// Convert visual index to actual option index
				optionIndex := m.getKubeconfigOptionIndexFromVisualIndex(m.existingEnv.KubeconfigFocusedIndex)
				if optionIndex >= 0 && optionIndex < len(m.existingEnv.KubeconfigOptions) {
					selectedOption := m.existingEnv.KubeconfigOptions[optionIndex]
					if selectedOption.Type == "custom" && selectedOption.Selected {
						m.existingEnv.CustomKubeconfigPath, cmd = m.existingEnv.CustomKubeconfigPath.Update(msg)
					}
				}
			} else if m.environmentFieldIndex == 4 {
				// Only Project Name is a text input now (Seeds and Presets are selection lists)
				m.existingEnv.ProjectName, cmd = m.existingEnv.ProjectName.Update(msg)
			}
		}
		return m, cmd
	}
}

// kubeconfigSection groups options of the same type into a named section.
type kubeconfigSection struct {
	sectionType string
	options     []KubeconfigOption
}

// kubeconfigSections returns the ordered list of non-empty kubeconfig sections.
func (m Model) kubeconfigSections() []kubeconfigSection {
	buckets := map[string][]KubeconfigOption{}
	for _, opt := range m.existingEnv.KubeconfigOptions {
		buckets[opt.Type] = append(buckets[opt.Type], opt)
	}
	var sections []kubeconfigSection
	for _, t := range []string{"env", "file", "custom"} {
		if len(buckets[t]) > 0 {
			sections = append(sections, kubeconfigSection{t, buckets[t]})
		}
	}
	return sections
}

// getKubeconfigSectionAtIndex returns the section type ("env", "file", "custom") if the given index
// is a section header, otherwise returns empty string.
func (m Model) getKubeconfigSectionAtIndex(index int) string {
	currentIndex := 0
	for _, s := range m.kubeconfigSections() {
		if index == currentIndex {
			return s.sectionType
		}
		currentIndex++
		if m.existingEnv.KubeconfigExpandedSections[s.sectionType] {
			currentIndex += len(s.options)
		}
	}
	return ""
}

// getKubeconfigOptionIndexFromVisualIndex converts a visual index (including headers) to an actual option index.
// Returns -1 if the visual index is a header.
func (m Model) getKubeconfigOptionIndexFromVisualIndex(visualIndex int) int {
	currentIndex := 0
	optionIndex := 0
	for _, s := range m.kubeconfigSections() {
		if visualIndex == currentIndex {
			return -1 // Header
		}
		currentIndex++
		if m.existingEnv.KubeconfigExpandedSections[s.sectionType] {
			for range s.options {
				if visualIndex == currentIndex {
					return optionIndex
				}
				currentIndex++
				optionIndex++
			}
		} else {
			optionIndex += len(s.options)
		}
	}
	return -1
}

// getMaxKubeconfigVisualIndex returns the maximum visual index (including headers).
func (m Model) getMaxKubeconfigVisualIndex() int {
	count := 0
	for _, s := range m.kubeconfigSections() {
		count++ // Header
		if m.existingEnv.KubeconfigExpandedSections[s.sectionType] {
			count += len(s.options)
		}
	}
	return count - 1
}

// ----------------------------------- Stage 2: Release Selection -----------------------------------

func (m *Model) handleReleaseSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keySelectAll:
		// Toggle select/deselect all
		allCurrentlySelected := true
		for _, major := range m.releaseSelection.MajorVersions {
			if !m.releaseSelection.SelectedMajor[major] {
				allCurrentlySelected = false
				break
			}
		}

		if allCurrentlySelected {
			// Deselect all
			m.releaseSelection.SelectedMajor = make(map[string]bool)
			m.releaseSelection.SelectedMinor = make(map[string]bool)
		} else {
			// Select all
			for _, major := range m.releaseSelection.MajorVersions {
				m.releaseSelection.SelectedMajor[major] = true
				for _, minor := range m.releaseSelection.MinorVersions[major] {
					m.releaseSelection.SelectedMinor[minor] = true
				}
			}
		}
		return m, nil

	case keyUp:
		if m.releaseSelection.IsMinorFocused {
			// Navigate within minor versions
			if m.releaseSelection.FocusedMinorIndex > 0 {
				m.releaseSelection.FocusedMinorIndex--
			} else {
				// Move back to major version
				m.releaseSelection.IsMinorFocused = false
				m.releaseSelection.FocusedMinorIndex = 0
			}
		} else {
			// Navigate between major versions
			if m.releaseSelection.FocusedMajorIndex > 0 {
				m.releaseSelection.FocusedMajorIndex--
				m.releaseSelection.FocusedMinorIndex = 0
			}
		}
		return m, nil

	case keyDown:
		if m.releaseSelection.FocusedMajorIndex < 0 {
			m.releaseSelection.FocusedMajorIndex = 0
			return m, nil
		}

		currentMajor := m.releaseSelection.MajorVersions[m.releaseSelection.FocusedMajorIndex]

		if !m.releaseSelection.IsMinorFocused {
			// Move into minor versions
			m.releaseSelection.IsMinorFocused = true
			m.releaseSelection.FocusedMinorIndex = 0
		} else {
			// Navigate within minor versions
			minorVersions := m.releaseSelection.MinorVersions[currentMajor]
			if m.releaseSelection.FocusedMinorIndex < len(minorVersions)-1 {
				m.releaseSelection.FocusedMinorIndex++
			} else {
				// Move to next major version
				if m.releaseSelection.FocusedMajorIndex < len(m.releaseSelection.MajorVersions)-1 {
					m.releaseSelection.FocusedMajorIndex++
					m.releaseSelection.IsMinorFocused = false
					m.releaseSelection.FocusedMinorIndex = 0
				}
			}
		}
		return m, nil

	case keySpace:
		// Handle selection
		if m.releaseSelection.FocusedMajorIndex < 0 {
			return m, nil
		}

		currentMajor := m.releaseSelection.MajorVersions[m.releaseSelection.FocusedMajorIndex]

		if m.releaseSelection.IsMinorFocused {
			// Toggle selection of focused minor version
			minorVersions := m.releaseSelection.MinorVersions[currentMajor]
			minorVersion := minorVersions[m.releaseSelection.FocusedMinorIndex]
			m.releaseSelection.SelectedMinor[minorVersion] = !m.releaseSelection.SelectedMinor[minorVersion]
			m.syncMajorSelectionState(currentMajor)
		} else {
			// Toggle selection of major version (selects/deselects all minors)
			m.releaseSelection.SelectedMajor[currentMajor] = !m.releaseSelection.SelectedMajor[currentMajor]
			minorVersions := m.releaseSelection.MinorVersions[currentMajor]
			for _, minor := range minorVersions {
				m.releaseSelection.SelectedMinor[minor] = m.releaseSelection.SelectedMajor[currentMajor]
			}
		}
		return m, nil

	case keyLeft:
		// Move back to major version if on minor
		if m.releaseSelection.IsMinorFocused {
			m.releaseSelection.IsMinorFocused = false
			m.releaseSelection.FocusedMinorIndex = 0
		}
		return m, nil

	case keyEnter:
		// Proceed if at least one version is selected
		if len(m.releaseSelection.SelectedMinor) > 0 {
			m.stage = stageProviderSelection
		}
		return m, nil

	case keyEsc:
		m.stage = stageEnvironmentSelection
		return m, nil
	}

	return m, nil
}

// based on whether all its minor versions are selected.
func (m *Model) syncMajorSelectionState(major string) {
	minorVersions := m.releaseSelection.MinorVersions[major]
	allSelected := true
	for _, minor := range minorVersions {
		if !m.releaseSelection.SelectedMinor[minor] {
			allSelected = false
			break
		}
	}
	m.releaseSelection.SelectedMajor[major] = allSelected
}

// ----------------------------------- Stage 3: Provider Selection -----------------------------------

// getFocusedCredentialContent returns full content for the currently focused provider credential.
func (m *Model) getFocusedCredentialContent() string {
	if len(m.providers) == 0 {
		return ""
	}
	p := m.providers[m.providerFocusIndex]

	if p.HasPresetCredentials && p.CredentialSource == CredentialSourcePreset {
		switch creds := p.PresetCredentials.(type) {
		case KubeVirtCredentials:
			return creds.Kubeconfig.Value()
		case GCPCredentials:
			return creds.ServiceAccount.Value()
		default:
			// Fallback: return the masked summary text
			return m.renderPresetCredentialsSummary(p)
		}
	}

	switch creds := p.Credentials.(type) {
	case KubeVirtCredentials:
		return creds.Kubeconfig.Value()
	case GCPCredentials:
		return creds.ServiceAccount.Value()
	case AWSCredentials:
		return creds.AccessKeyID.Value() + "\n" + creds.SecretAccessKey.Value()
	default:
		return ""
	}
}

func (m *Model) handleProviderSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case keyUp:
		// Move up through fields
		if m.providerFieldIndex > 0 {
			// Move to previous field in current provider
			m.providerFieldIndex--
		} else if m.providerFocusIndex > 0 {
			// Move to previous provider's last field
			m.providerFocusIndex--
			if m.providers[m.providerFocusIndex].Selected {
				m.providerFieldIndex = m.getMaxFieldIndexForProvider(m.providers[m.providerFocusIndex])
			} else {
				m.providerFieldIndex = 0
			}
		}
		m.updateProviderFocus()
		return m, nil
	case keyDown:
		// Move down through fields
		maxField := 0
		if m.providers[m.providerFocusIndex].Selected {
			maxField = m.getMaxFieldIndexForProvider(m.providers[m.providerFocusIndex])
		}

		if m.providerFieldIndex < maxField {
			// Move to next field in current provider
			m.providerFieldIndex++
		} else if m.providerFocusIndex < len(m.providers)-1 {
			// Move to next provider's first field
			m.providerFocusIndex++
			m.providerFieldIndex = 0
		}
		m.updateProviderFocus()
		return m, nil
	case keyEnter:
		// Proceed to next stage if at least one provider is selected
		selectedProviders := selectedProviderNames(m.providers)
		if len(selectedProviders) > 0 {
			// Initialize distribution selection based on selected providers
			m.distributionSelection = initializeDistributionSelection(selectedProviders)
			m.stage = stageDistributionSelection
		}
		return m, nil
	case keySpace:
		// Toggle selection based on focused provider (only when on the checkbox)
		if m.providerFieldIndex == 0 {
			m.providers[m.providerFocusIndex].Selected = !m.providers[m.providerFocusIndex].Selected
			if m.providers[m.providerFocusIndex].Selected {
				m.expandedProviderIndex = m.providerFocusIndex
				m.providerFieldIndex = 1 // Move to first field after checkbox
			} else {
				// If deselecting the expanded one, collapse it
				if m.expandedProviderIndex == m.providerFocusIndex {
					m.expandedProviderIndex = -1
				}
			}
		} else if m.providerFieldIndex == 1 || m.providerFieldIndex == 2 {
			// Select the option at the current field index
			if m.providers[m.providerFocusIndex].HasPresetCredentials {
				if m.providerFieldIndex == 1 {
					// Pressing space on "From Preset" selects preset
					m.providers[m.providerFocusIndex].CredentialSource = CredentialSourcePreset
				} else if m.providerFieldIndex == 2 {
					// Pressing space on "Enter Custom Credentials" selects custom
					m.providers[m.providerFocusIndex].CredentialSource = CredentialSourceCustom
				}
			}
		}
		m.updateProviderFocus()
		return m, nil
	case keyEsc:
		m.stage = stageReleaseSelection
		return m, nil
	case keyI:
		// Open view modal for focused provider credential (if any)
		if m.providers[m.providerFocusIndex].Selected {
			content := m.getFocusedCredentialContent()
			if content != "" {
				m.showViewModal(m.providers[m.providerFocusIndex].DisplayName+" credentials", content)
			}
		}
		return m, nil
	default:
		// Update the focused text input (only for custom credentials)
		if m.providers[m.providerFocusIndex].Selected && m.providerFieldIndex > 0 {
			// Skip text input if using preset credentials
			if m.providers[m.providerFocusIndex].HasPresetCredentials &&
				m.providers[m.providerFocusIndex].CredentialSource == CredentialSourcePreset {
				return m, nil
			}
			fIdx := credentialFieldIndex(m.providers[m.providerFocusIndex], m.providerFieldIndex)
			if fIdx < 0 {
				return m, nil
			}
			m.providers[m.providerFocusIndex], cmd = m.updateProviderCredentialField(m.providers[m.providerFocusIndex], fIdx, msg)
		}
		return m, cmd
	}
}

// getMaxFieldIndexForProvider returns the maximum field index for a provider based on its credential type.
func (m Model) getMaxFieldIndexForProvider(provider Provider) int {
	baseFields := 0
	switch provider.Credentials.(type) {
	case GCPCredentials, AnexiaCredentials, DigitalOceanCredentials, HetznerCredentials, KubeVirtCredentials:
		baseFields = 1
	case AlibabaCredentials, VSphereCredentials:
		baseFields = 2
	case AWSCredentials, AzureCredentials:
		baseFields = 4
	case VMwareCloudDirectorCredentials:
		baseFields = 5
	case NutanixCredentials:
		baseFields = 7
	case OpenStackCredentials:
		baseFields = 8
	default:
		baseFields = 1
	}

	// Add fields for credential source selector if preset is available
	if provider.HasPresetCredentials {
		// Field 1 = "From Preset", Field 2 = "Enter Custom Credentials"
		if provider.CredentialSource == CredentialSourcePreset {
			// Only show preset fields (no custom credential inputs)
			return 2
		}
		// If custom is selected, add the credential source selectors + custom fields
		return baseFields + 2
	}

	return baseFields
}

type credentialFieldAction int

const (
	credentialFieldActionUpdate credentialFieldAction = iota
	credentialFieldActionBlur
	credentialFieldActionFocus
)

func applyTextInputFieldAction(fields []*textinput.Model, action credentialFieldAction, fieldIndex int, msg tea.KeyMsg) tea.Cmd {
	if len(fields) == 0 {
		return nil
	}

	switch action {
	case credentialFieldActionUpdate:
		if fieldIndex < 1 || fieldIndex > len(fields) {
			return nil
		}
		updated, cmd := fields[fieldIndex-1].Update(msg)
		*fields[fieldIndex-1] = updated
		return cmd
	case credentialFieldActionBlur:
		for _, field := range fields {
			field.Blur()
		}
	case credentialFieldActionFocus:
		if fieldIndex < 1 || fieldIndex > len(fields) {
			return nil
		}
		fields[fieldIndex-1].Focus()
	}

	return nil
}

func (m Model) applyProviderCredentialFieldAction(provider Provider, fieldIndex int, msg tea.KeyMsg, action credentialFieldAction) (Provider, tea.Cmd) {
	var cmd tea.Cmd

	switch creds := provider.Credentials.(type) {
	case AWSCredentials:
		fields := []*textinput.Model{&creds.AccessKeyID, &creds.SecretAccessKey, &creds.AssumeRoleARN, &creds.AssumeRoleExternalID}
		cmd = applyTextInputFieldAction(fields, action, fieldIndex, msg)
		provider.Credentials = creds
	case AzureCredentials:
		fields := []*textinput.Model{&creds.TenantID, &creds.SubscriptionID, &creds.ClientID, &creds.ClientSecret}
		cmd = applyTextInputFieldAction(fields, action, fieldIndex, msg)
		provider.Credentials = creds
	case GCPCredentials:
		fields := []*textinput.Model{&creds.ServiceAccount}
		cmd = applyTextInputFieldAction(fields, action, fieldIndex, msg)
		provider.Credentials = creds
	case AlibabaCredentials:
		fields := []*textinput.Model{&creds.AccessKeyID, &creds.AccessKeySecret}
		cmd = applyTextInputFieldAction(fields, action, fieldIndex, msg)
		provider.Credentials = creds
	case AnexiaCredentials:
		fields := []*textinput.Model{&creds.Token}
		cmd = applyTextInputFieldAction(fields, action, fieldIndex, msg)
		provider.Credentials = creds
	case DigitalOceanCredentials:
		fields := []*textinput.Model{&creds.Token}
		cmd = applyTextInputFieldAction(fields, action, fieldIndex, msg)
		provider.Credentials = creds
	case HetznerCredentials:
		fields := []*textinput.Model{&creds.Token}
		cmd = applyTextInputFieldAction(fields, action, fieldIndex, msg)
		provider.Credentials = creds
	case KubeVirtCredentials:
		fields := []*textinput.Model{&creds.Kubeconfig}
		cmd = applyTextInputFieldAction(fields, action, fieldIndex, msg)
		provider.Credentials = creds
	case NutanixCredentials:
		fields := []*textinput.Model{&creds.Username, &creds.Password, &creds.ClusterName, &creds.ProxyURL, &creds.CSIUsername, &creds.CSIPassword, &creds.CSIEndpoint}
		cmd = applyTextInputFieldAction(fields, action, fieldIndex, msg)
		provider.Credentials = creds
	case OpenStackCredentials:
		fields := []*textinput.Model{&creds.Username, &creds.Password, &creds.Project, &creds.ProjectID, &creds.Domain, &creds.ApplicationCredentialID, &creds.ApplicationCredentialSecret, &creds.Token}
		cmd = applyTextInputFieldAction(fields, action, fieldIndex, msg)
		provider.Credentials = creds
	case VSphereCredentials:
		fields := []*textinput.Model{&creds.Username, &creds.Password}
		cmd = applyTextInputFieldAction(fields, action, fieldIndex, msg)
		provider.Credentials = creds
	case VMwareCloudDirectorCredentials:
		fields := []*textinput.Model{&creds.Username, &creds.Password, &creds.APIToken, &creds.Organization, &creds.VDC}
		cmd = applyTextInputFieldAction(fields, action, fieldIndex, msg)
		provider.Credentials = creds
	}

	return provider, cmd
}

// updateProviderCredentialField updates a specific credential field for a provider.
func (m Model) updateProviderCredentialField(provider Provider, fieldIndex int, msg tea.KeyMsg) (Provider, tea.Cmd) {
	return m.applyProviderCredentialFieldAction(provider, fieldIndex, msg, credentialFieldActionUpdate)
}

// updateProviderFocus manages focus state for provider selection.
func (m *Model) updateProviderFocus() {
	// Blur all provider text inputs
	for i := range m.providers {
		m.providers[i] = m.blurAllProviderFields(m.providers[i])
	}

	// Focus the current field in the currently focused provider
	if m.providers[m.providerFocusIndex].Selected && m.providerFieldIndex > 0 {
		if fIdx := credentialFieldIndex(m.providers[m.providerFocusIndex], m.providerFieldIndex); fIdx > 0 {
			m.providers[m.providerFocusIndex] = m.focusProviderField(m.providers[m.providerFocusIndex], fIdx)
		}
	}
}

// blurAllProviderFields blurs all credential fields in a provider.
func (m Model) blurAllProviderFields(provider Provider) Provider {
	provider, _ = m.applyProviderCredentialFieldAction(provider, 0, tea.KeyMsg{}, credentialFieldActionBlur)
	return provider
}

// focusProviderField focuses a specific field in provider credentials.
func (m Model) focusProviderField(provider Provider, fieldIndex int) Provider {
	provider, _ = m.applyProviderCredentialFieldAction(provider, fieldIndex, tea.KeyMsg{}, credentialFieldActionFocus)
	return provider
}

// credentialFieldIndex returns the 1-based credential field index for a provider,
// adjusting for preset selector fields. Returns -1 if the field is not a credential input.
func credentialFieldIndex(provider Provider, fieldIndex int) int {
	if provider.HasPresetCredentials {
		if fieldIndex > 2 {
			return fieldIndex - 2
		}
		return -1 // On radio-button fields 1 or 2
	}
	return fieldIndex
}

// ----------------------------------- Stage 4: Distribution Selection -----------------------------------

func (m *Model) handleDistributionSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyUp:
		if m.distributionSelection.FocusedIndex > 0 {
			m.distributionSelection.FocusedIndex--
		}
		return m, nil

	case keyDown:
		maxIndex := m.getDistributionMaxFocusIndex()
		if m.distributionSelection.FocusedIndex < maxIndex {
			m.distributionSelection.FocusedIndex++
		}
		return m, nil

	case keyRight:
		// Expand provider if focused on provider header
		provider := m.getDistributionFocusedProvider()
		if provider != "" {
			m.distributionSelection.ExpandedProviders[provider] = true
		}
		return m, nil

	case keyLeft:
		// Collapse provider if focused on provider header
		provider := m.getDistributionFocusedProvider()
		if provider != "" {
			m.distributionSelection.ExpandedProviders[provider] = false
		}
		return m, nil

	case keySpace:
		// Toggle selection for the focused distribution (not provider header)
		distKey := m.getDistributionFocusedDistribution()
		if distKey != "" {
			m.distributionSelection.Selected[distKey] = !m.distributionSelection.Selected[distKey]
		}
		return m, nil

	case keySelectAll:
		// Toggle select/deselect all distributions across all providers
		allSelected := true
		for _, provider := range m.distributionSelection.Providers {
			dists := m.distributionSelection.DistributionsByProvider[provider]
			for _, dist := range dists {
				selectionKey := fmt.Sprintf("%s:%s", provider, dist)
				if !m.distributionSelection.Selected[selectionKey] {
					allSelected = false
					break
				}
			}
			if !allSelected {
				break
			}
		}

		// Toggle all distributions
		for _, provider := range m.distributionSelection.Providers {
			dists := m.distributionSelection.DistributionsByProvider[provider]
			for _, dist := range dists {
				selectionKey := fmt.Sprintf("%s:%s", provider, dist)
				m.distributionSelection.Selected[selectionKey] = !allSelected
			}
		}
		return m, nil

	case keyEnter:
		if hasAnySelected(m.distributionSelection.Selected) {
			// Get all selected providers to initialize datacenter settings
			selectedProviders := selectedProviderNames(m.providers)

			// Initialize datacenter settings for all selected providers
			if len(selectedProviders) > 0 {
				m.datacenterSettingsSelection = initializeDatacenterSettingsSelection(selectedProviders)
				cmds := m.buildDatacenterSettingsFetchCmds(selectedProviders)

				m.stage = stageDatacenterSettingsSelection // Move to next stage
				return m, tea.Batch(cmds...)
			}
		}
		return m, nil

	case keyEsc:
		m.stage = stageProviderSelection
		return m, nil
	}

	return m, nil
}

// getDistributionMaxFocusIndex returns the maximum focus index for distributions.
func (m Model) getDistributionMaxFocusIndex() int {
	count := 0
	for _, provider := range m.distributionSelection.Providers {
		count++ // Provider header
		if m.distributionSelection.ExpandedProviders[provider] {
			dists := m.distributionSelection.DistributionsByProvider[provider]
			count += len(dists) // Distributions
		}
	}
	return count - 1
}

// getDistributionFocusedProvider returns the provider name if focused on a provider header, empty string otherwise.
func (m Model) getDistributionFocusedProvider() string {
	currentIndex := 0
	for _, provider := range m.distributionSelection.Providers {
		if currentIndex == m.distributionSelection.FocusedIndex {
			return provider
		}
		currentIndex++

		if m.distributionSelection.ExpandedProviders[provider] {
			dists := m.distributionSelection.DistributionsByProvider[provider]
			currentIndex += len(dists)
		}
	}
	return ""
}

// getDistributionFocusedDistribution returns the distribution key if focused on a distribution, empty string otherwise.
func (m Model) getDistributionFocusedDistribution() string {
	currentIndex := 0
	for _, provider := range m.distributionSelection.Providers {
		currentIndex++ // Provider header

		if m.distributionSelection.ExpandedProviders[provider] {
			dists := m.distributionSelection.DistributionsByProvider[provider]
			for _, dist := range dists {
				if currentIndex == m.distributionSelection.FocusedIndex {
					return fmt.Sprintf("%s:%s", provider, dist)
				}
				currentIndex++
			}
		}
	}
	return ""
}

// ----------------------------------- Stage 5: Datacenter Settings Selection -----------------------------------

// handleDatacenterSettingsSelection handles input for the datacenter settings selection stage.
func (m *Model) handleDatacenterSettingsSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if handleSharedSettingsKeys(
		msg.String(),
		&m.datacenterSettingsSelection.FocusedIndex,
		m.getDatacenterMaxFocusIndex(),
		m.getDatacenterFocusedItem,
		m.getDatacenterFocusedOption,
		m.datacenterSettingsSelection.ExpandedProviders,
		m.datacenterSettingsSelection.Providers,
		m.datacenterSettingsSelection.SettingsByProvider,
		m.datacenterSettingsSelection.Selected,
		m.datacenterSettingsSelection.SelectedGroups,
		&m.datacenterViewport,
	) {
		return m, nil
	}

	switch msg.String() {
	case keyEnter:
		// Get all selected providers to initialize cluster settings
		selectedProviders := selectedProviderNames(m.providers)

		// Initialize cluster settings for the selected providers
		if len(selectedProviders) > 0 {
			m.clusterSettingsSelection = initializeClusterSettingsSelection(selectedProviders)
		}

		// Move to next stage
		m.stage = stageClusterSettingsSelection
		return m, nil

	case keyEsc:
		// Move to previous stage
		m.stage = stageDistributionSelection
		return m, nil
	}

	return m, nil
}

// getDatacenterMaxFocusIndex returns the maximum focus index for datacenter settings.
func (m Model) getDatacenterMaxFocusIndex() int {
	return getSettingsMaxFocusIndex(
		m.datacenterSettingsSelection.Providers,
		m.datacenterSettingsSelection.SettingsByProvider,
		m.datacenterSettingsSelection.ExpandedProviders,
	)
}

// getDatacenterFocusedItem returns (provider, groupIdx) for the focused item.
// Returns ("", -1) if not on provider or group, (provider, -1) if on provider header,
// (provider, groupIdx) if on setting group.
func (m Model) getDatacenterFocusedItem() (string, int) {
	return getFocusedSettingItem(
		m.datacenterSettingsSelection.Providers,
		m.datacenterSettingsSelection.SettingsByProvider,
		m.datacenterSettingsSelection.ExpandedProviders,
		m.datacenterSettingsSelection.FocusedIndex,
		true,
	)
}

// getDatacenterFocusedOption returns (provider, groupIdx, optionIdx) for the focused option.
// Returns ("", -1, -1) if not focused on an option.
func (m Model) getDatacenterFocusedOption() (string, int, int) {
	return getFocusedSettingOption(
		m.datacenterSettingsSelection.Providers,
		m.datacenterSettingsSelection.SettingsByProvider,
		m.datacenterSettingsSelection.ExpandedProviders,
		m.datacenterSettingsSelection.FocusedIndex,
	)
}

// ----------------------------------- Stage 6: Cluster Settings Selection -----------------------------------

// handleClusterSettingsSelection handles input for the cluster settings selection stage.
func (m *Model) handleClusterSettingsSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if handleSharedSettingsKeys(
		msg.String(),
		&m.clusterSettingsSelection.FocusedIndex,
		m.getClusterMaxFocusIndex(),
		m.getClusterFocusedItem,
		m.getClusterFocusedOption,
		m.clusterSettingsSelection.ExpandedProviders,
		m.clusterSettingsSelection.Providers,
		m.clusterSettingsSelection.SettingsByProvider,
		m.clusterSettingsSelection.Selected,
		m.clusterSettingsSelection.SelectedGroups,
		&m.clusterViewport,
	) {
		return m, nil
	}

	switch msg.String() {
	case keyEnter:
		// Get all selected providers to initialize machine deployment settings
		selectedProviders := selectedProviderNames(m.providers)

		// Initialize machine deployment settings for the selected providers
		if len(selectedProviders) > 0 {
			m.machineDeploymentSettingsSelection = initializeMachineDeploymentSettingsSelection(selectedProviders)

			cmds := m.buildMachineSettingsFetchCmds(selectedProviders)

			// Move to next stage
			m.stage = stageMachineDeploymentSettingsSelection
			return m, tea.Batch(cmds...)
		}

		// Move to next stage
		m.stage = stageMachineDeploymentSettingsSelection
		return m, nil

	case keyEsc:
		// Move to previous stage
		m.stage = stageDatacenterSettingsSelection
		return m, nil
	}

	return m, nil
}

// getClusterMaxFocusIndex returns the maximum focus index for cluster settings.
func (m Model) getClusterMaxFocusIndex() int {
	return getSettingsMaxFocusIndex(
		m.clusterSettingsSelection.Providers,
		m.clusterSettingsSelection.SettingsByProvider,
		m.clusterSettingsSelection.ExpandedProviders,
	)
}

// getClusterFocusedItem returns (provider, groupIdx) for the focused item.
// Returns ("", -1) if not on provider or group, (provider, -1) if on provider header,
// (provider, groupIdx) if on setting group.
func (m Model) getClusterFocusedItem() (string, int) {
	return getFocusedSettingItem(
		m.clusterSettingsSelection.Providers,
		m.clusterSettingsSelection.SettingsByProvider,
		m.clusterSettingsSelection.ExpandedProviders,
		m.clusterSettingsSelection.FocusedIndex,
		false,
	)
}

// getClusterFocusedOption returns (provider, groupIdx, optionIdx) for the focused option.
// Returns ("", -1, -1) if not focused on an option.
func (m Model) getClusterFocusedOption() (string, int, int) {
	return getFocusedSettingOption(
		m.clusterSettingsSelection.Providers,
		m.clusterSettingsSelection.SettingsByProvider,
		m.clusterSettingsSelection.ExpandedProviders,
		m.clusterSettingsSelection.FocusedIndex,
	)
}

// ----------------------------------- Stage 7: Machine Deployment Settings -----------------------------------

// handleMachineDeploymentSettingsSelection handles input for the machine deployment settings selection stage.
func (m *Model) handleMachineDeploymentSettingsSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if handleSharedSettingsKeys(
		msg.String(),
		&m.machineDeploymentSettingsSelection.FocusedIndex,
		m.getMachineDeploymentMaxFocusIndex(),
		m.getMachineDeploymentFocusedItem,
		m.getMachineDeploymentFocusedOption,
		m.machineDeploymentSettingsSelection.ExpandedProviders,
		m.machineDeploymentSettingsSelection.Providers,
		m.machineDeploymentSettingsSelection.SettingsByProvider,
		m.machineDeploymentSettingsSelection.Selected,
		m.machineDeploymentSettingsSelection.SelectedGroups,
		&m.machineViewport,
	) {
		return m, nil
	}

	switch msg.String() {
	case keyEnter:
		// Initialize cluster configuration for the next stage
		m.clusterConfiguration = initializeClusterConfiguration()

		// Move to next stage
		m.stage = stageClusterConfiguration
		return m, nil

	case keyEsc:
		// Move to previous stage
		m.stage = stageClusterSettingsSelection
		return m, nil
	}

	return m, nil
}

// getMachineDeploymentMaxFocusIndex returns the maximum focus index for machine deployment settings.
func (m Model) getMachineDeploymentMaxFocusIndex() int {
	return getSettingsMaxFocusIndex(
		m.machineDeploymentSettingsSelection.Providers,
		m.machineDeploymentSettingsSelection.SettingsByProvider,
		m.machineDeploymentSettingsSelection.ExpandedProviders,
	)
}

// getMachineDeploymentFocusedItem returns (provider, groupIdx) for the focused item.
// Returns ("", -1) if not on provider or group, (provider, -1) if on provider header,
// (provider, groupIdx) if on setting group.
func (m Model) getMachineDeploymentFocusedItem() (string, int) {
	return getFocusedSettingItem(
		m.machineDeploymentSettingsSelection.Providers,
		m.machineDeploymentSettingsSelection.SettingsByProvider,
		m.machineDeploymentSettingsSelection.ExpandedProviders,
		m.machineDeploymentSettingsSelection.FocusedIndex,
		true,
	)
}

// getMachineDeploymentFocusedOption returns (provider, groupIdx, optionIdx) for the focused option.
// Returns ("", -1, -1) if not focused on an option.
func (m Model) getMachineDeploymentFocusedOption() (string, int, int) {
	return getFocusedSettingOption(
		m.machineDeploymentSettingsSelection.Providers,
		m.machineDeploymentSettingsSelection.SettingsByProvider,
		m.machineDeploymentSettingsSelection.ExpandedProviders,
		m.machineDeploymentSettingsSelection.FocusedIndex,
	)
}

// ----------------------------------- Stage 8: Cluster Configuration -----------------------------------

// handleClusterConfiguration handles input for the cluster configuration stage.
func (m *Model) handleClusterConfiguration(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.clusterConfiguration.EditMode {
		return m.handleClusterConfigurationEdit(msg)
	}

	maxIndex := m.getClusterConfigMaxFocusIndex()

	switch msg.String() {
	case keyUp:
		if m.clusterConfiguration.FocusedIndex > 0 {
			m.clusterConfiguration.FocusedIndex--
		}
		return m, nil

	case keyLeft, keyRight:
		// Expand/collapse category if focused on a category header
		catIdx, _, itemType := m.getClusterConfigFocusedItem()
		if itemType == "category" && catIdx >= 0 && catIdx < len(m.clusterConfiguration.Categories) {
			categoryName := m.clusterConfiguration.Categories[catIdx].Name
			m.clusterConfiguration.ExpandedCategories[categoryName] = (msg.String() == keyRight)
		}
		return m, nil

	case keyDown:
		if m.clusterConfiguration.FocusedIndex < maxIndex {
			m.clusterConfiguration.FocusedIndex++
		}
		return m, nil

	case keySpace:
		categoryIdx, settingIdx, itemType := m.getClusterConfigFocusedItem()
		if itemType == "setting" {
			setting := &m.clusterConfiguration.Categories[categoryIdx].Settings[settingIdx]

			switch setting.Type {
			case ConfigTypeBool:
				// Toggle boolean value
				setting.Value = !setting.Value.(bool)
			case ConfigTypeString, ConfigTypeInt, ConfigTypeIntArray, ConfigTypeStringArray:
				// Enter edit mode
				m.clusterConfiguration.EditMode = true
				m.clusterConfiguration.EditingBuffer = m.formatConfigValue(setting)
			}
		}
		return m, nil

	case keyEnter:
		// Initialize review state before moving to next stage
		m.Review.ExpandedProviders = make(map[string]bool)
		m.Review.ExpandedSections = make(map[string]bool)
		m.Review.FocusedIndex = 0

		// Expand all providers by default
		for _, provider := range m.providers {
			if provider.Selected {
				m.Review.ExpandedProviders[provider.DisplayName] = true
			}
		}

		// Move to next stage
		m.stage++
		return m, nil

	case keyEsc:
		// Move to previous stage
		m.stage = stageMachineDeploymentSettingsSelection
		return m, nil
	}

	return m, nil
}

// handleClusterConfigurationEdit handles input when editing a configuration value.
func (m *Model) handleClusterConfigurationEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	categoryIdx, settingIdx, _ := m.getClusterConfigFocusedItem()
	setting := &m.clusterConfiguration.Categories[categoryIdx].Settings[settingIdx]

	switch msg.String() {
	case keyEnter:
		// Save the edited value
		if err := m.parseConfigValue(setting, m.clusterConfiguration.EditingBuffer); err == nil {
			m.clusterConfiguration.EditMode = false
			m.clusterConfiguration.EditingBuffer = ""
		}
		return m, nil

	case keyEsc:
		// Cancel editing
		m.clusterConfiguration.EditMode = false
		m.clusterConfiguration.EditingBuffer = ""
		return m, nil

	case "backspace":
		if len(m.clusterConfiguration.EditingBuffer) > 0 {
			m.clusterConfiguration.EditingBuffer = m.clusterConfiguration.EditingBuffer[:len(m.clusterConfiguration.EditingBuffer)-1]
		}
		return m, nil

	default:
		// Add character to buffer (allow alphanumeric, comma, space, and Gi/Mi suffixes)
		if len(msg.String()) == 1 {
			m.clusterConfiguration.EditingBuffer += msg.String()
		}
		return m, nil
	}
}

// getClusterConfigMaxFocusIndex returns the maximum valid focus index.
func (m Model) getClusterConfigMaxFocusIndex() int {
	count := -1
	for _, category := range m.clusterConfiguration.Categories {
		count++                         // Category header
		count += len(category.Settings) // Settings
	}
	return count
}

// getClusterConfigFocusedItem returns the currently focused item.
// Returns: category index, setting index, item type ("category" or "setting")
func (m Model) getClusterConfigFocusedItem() (int, int, string) {
	currentIndex := 0
	for catIdx, category := range m.clusterConfiguration.Categories {
		if currentIndex == m.clusterConfiguration.FocusedIndex {
			return catIdx, -1, "category"
		}
		currentIndex++

		for setIdx := range category.Settings {
			if currentIndex == m.clusterConfiguration.FocusedIndex {
				return catIdx, setIdx, "setting"
			}
			currentIndex++
		}
	}
	return 0, 0, "category"
}

// formatConfigValue formats a configuration value as a string for editing.
func (m Model) formatConfigValue(setting *ConfigSetting) string {
	switch setting.Type {
	case ConfigTypeString:
		return setting.Value.(string)
	case ConfigTypeInt:
		return fmt.Sprintf("%d", setting.Value.(int))
	case ConfigTypeIntArray:
		values := setting.Value.([]int)
		strs := make([]string, len(values))
		for i, v := range values {
			strs[i] = fmt.Sprintf("%d", v)
		}
		return strings.Join(strs, ", ")
	case ConfigTypeStringArray:
		return strings.Join(setting.Value.([]string), ", ")
	case ConfigTypeBool:
		if setting.Value.(bool) {
			return "Yes"
		}
		return "No"
	}
	return ""
}

// parseConfigValue parses a string into the appropriate type and updates the setting.
func (m Model) parseConfigValue(setting *ConfigSetting, value string) error {
	value = strings.TrimSpace(value)

	switch setting.Type {
	case ConfigTypeString:
		setting.Value = value
	case ConfigTypeInt:
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		setting.Value = parsed
	case ConfigTypeIntArray:
		parts := strings.Split(value, ",")
		ints := make([]int, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			parsed, err := strconv.Atoi(part)
			if err != nil {
				return err
			}
			ints = append(ints, parsed)
		}
		if len(ints) == 0 {
			return fmt.Errorf("at least one value required")
		}
		setting.Value = ints
	case ConfigTypeStringArray:
		parts := strings.Split(value, ",")
		strs := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				strs = append(strs, part)
			}
		}
		if len(strs) == 0 {
			return fmt.Errorf("at least one value required")
		}
		setting.Value = strs
	}
	return nil
}

// ----------------------------------- Stage 9: Review Settings -----------------------------------

// handleReviewSettings handles input for the review settings stage with nested provider structure.
func (m *Model) handleReviewSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	providerReviews := m.generateReviewYAML()
	maxIndex := m.getReviewMaxFocusIndex(providerReviews)

	switch msg.String() {
	case keyUp:
		if m.Review.FocusedIndex > 0 {
			m.Review.FocusedIndex--
		}
		return m, nil

	case keyDown:
		if m.Review.FocusedIndex < maxIndex {
			m.Review.FocusedIndex++
		}
		return m, nil

	case keyLeft, keyRight, keySpace:
		m.toggleReviewFocusedItem(providerReviews)
		return m, nil

	case keyEnter:
		// Save provider configurations if checkbox is enabled
		if m.Review.SaveToFile {
			if err := m.saveProviderConfigurations(); err != nil {
				m.executionError = fmt.Sprintf("Warning: Failed to save configuration files: %v", err)
			}
		}

		// Initialize viewport for log display
		viewportWidth := m.getUIInnerWidth()
		viewportHeight := m.getUIHeight() - 4
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		m.Review.Viewport = viewport.New(viewportWidth, viewportHeight)
		m.Review.Viewport.SetContent("")
		m.logs = []string{}

		// Move to execution stage and start execution
		m.stage = stageExecuting
		return m, m.executeConformanceTests()

	case keyEsc:
		m.stage = stageClusterConfiguration
		return m, nil
	}

	return m, nil
}

// getReviewMaxFocusIndex returns the maximum navigable index for the review stage.
func (m Model) getReviewMaxFocusIndex(providerReviews []ProviderReview) int {
	count := 0
	for _, pr := range providerReviews {
		count++ // Provider header
		if m.Review.ExpandedProviders[pr.ProviderName] {
			count += len(pr.Sections)
		}
	}
	count++ // Save checkbox
	return count - 1
}

// toggleReviewFocusedItem toggles expansion for the currently focused review item
// (provider header, section header, or save checkbox).
func (m *Model) toggleReviewFocusedItem(providerReviews []ProviderReview) {
	currentIndex := 0
	for _, pr := range providerReviews {
		if currentIndex == m.Review.FocusedIndex {
			m.Review.ExpandedProviders[pr.ProviderName] = !m.Review.ExpandedProviders[pr.ProviderName]
			return
		}
		currentIndex++

		if m.Review.ExpandedProviders[pr.ProviderName] {
			for _, section := range pr.Sections {
				if currentIndex == m.Review.FocusedIndex {
					sectionKey := fmt.Sprintf("%s:%s", pr.ProviderName, section.Name)
					m.Review.ExpandedSections[sectionKey] = !m.Review.ExpandedSections[sectionKey]
					return
				}
				currentIndex++
			}
		}
	}
	// Save checkbox
	if currentIndex == m.Review.FocusedIndex {
		m.Review.SaveToFile = !m.Review.SaveToFile
	}
}

// ----------------------------------- Stage 10: Executing & Done -----------------------------------

// handleDone processes key input in the done stage.
func (m Model) handleDone(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyQuit:
		return m, tea.Quit
	}
	return m, nil
}

// handleExecuting processes key input in the executing stage.
func (m *Model) handleExecuting(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyControlC:
		// Don't show quit confirmation if already cancelling
		if !m.executionCancelling {
			m.quitConfirmVisible = true
			m.quitConfirmIndex = 0 // Default to "No"
		}
		return m, nil
	}

	// Allow user to scroll through logs if viewport is present
	if m.Review.Viewport.TotalLineCount() > 0 {
		switch msg.String() {
		case keyUp, mouseWheelUp:
			m.Review.Viewport.ScrollUp(1)
			return m, nil
		case keyDown, mouseWheelDown:
			m.Review.Viewport.ScrollDown(1)
			return m, nil
		}
	}
	return m, nil
}

// handleDoneMessage processes completion messages from doneMsg.
func (m *Model) handleDoneMessage(_ doneMsg) tea.Cmd {
	// Only mark success if no prior error was recorded
	if m.executionError == "" {
		m.executionDone = true
		m.logs = append(m.logs, "Successfully applied configuration!", "[DONE] Process completed")
		m.refreshLogsViewport()
		m.stage = stageDone
	}
	return nil
}

// handleExecOutput processes execution output.
func (m *Model) handleExecOutput(msg execOutputMsg) tea.Cmd {
	m.appendNonEmptyOutput(msg.output)

	if msg.success {
		m.executionDone = true
		m.stage = stageDone
	}
	if msg.err != nil {
		m.executionError = msg.err.Error()
	}
	return nil
}

func streamLogs(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return doneMsg{success: true}
		}
		return msg
	}
}

// confirmQuit returns the appropriate quit command: cancellation during execution, or tea.Quit otherwise.
func (m *Model) confirmQuit() tea.Cmd {
	if m.stage == stageExecuting && !m.executionCancelling {
		return m.startExecutionCancellation()
	}
	return tea.Quit
}

func (m *Model) handleQuitConfirmation(msg tea.KeyMsg) (handled bool, cmd tea.Cmd) {
	switch msg.String() {
	case keyLeft, keyShiftTab:
		if m.quitConfirmIndex > 0 {
			m.quitConfirmIndex--
		}
		return true, nil
	case keyRight, keyTab:
		if m.quitConfirmIndex < 1 {
			m.quitConfirmIndex++
		}
		return true, nil
	case keyEnter:
		if m.quitConfirmIndex == 1 { // Yes
			return true, m.confirmQuit()
		}
		m.quitConfirmVisible = false
		return true, nil
	case keyEsc, keyNo:
		m.quitConfirmVisible = false
		return true, nil
	case keyYes, keyControlC:
		return true, m.confirmQuit()
	}
	return false, nil
}

// ----------------------------------- Common: UI Helpers -----------------------------------

// showViewModal initialises and displays the view modal with the given title and content.
func (m *Model) showViewModal(title, content string) {
	m.viewModalContent = content
	m.viewModalTitle = title
	m.viewModalVisible = true

	uiInnerWidth := m.getUIInnerWidth()
	uiInnerHeight := m.getUIInnerHeight()
	vpHeight := uiInnerHeight
	if vpHeight < 3 {
		vpHeight = 3
	}
	m.viewModalViewport = viewport.New(uiInnerWidth, vpHeight)
	wrappedContent := lipgloss.NewStyle().Width(uiInnerWidth).Render(content)
	m.viewModalViewport.SetContent(wrappedContent)
	m.viewModalViewport.GotoTop()
}

// ----------------------------------- Common: Settings Tree Helpers -----------------------------------

func settingGroupKey(provider, group string) string {
	return fmt.Sprintf("%s:%s", provider, group)
}

func settingOptionKey(provider, group, option string) string {
	return fmt.Sprintf("%s:%s:%s", provider, group, option)
}

func isProviderHeaderFocused(provider string, groupIdx int) bool {
	return provider != "" && groupIdx == -1
}

// getFocusedSettingItem resolves the currently focused settings-tree item.
// It returns (provider, -1) for provider headers and (provider, groupIdx) for group headers.
// When respectGroupExpansion is true, option rows are counted only if group.IsExpanded is true.
func getFocusedSettingItem(providers []string, settingsByProvider map[string][]SettingGroup, expandedProviders map[string]bool, focusedIndex int, respectGroupExpansion bool) (string, int) {
	currentIndex := 0
	multiProvider := len(providers) > 1
	for _, provider := range providers {
		if multiProvider {
			if currentIndex == focusedIndex {
				return provider, -1
			}
			currentIndex++
		}

		if expandedProviders[provider] || !multiProvider {
			groups := settingsByProvider[provider]
			for groupIdx, group := range groups {
				if currentIndex == focusedIndex {
					return provider, groupIdx
				}
				currentIndex++

				if !respectGroupExpansion || group.IsExpanded {
					currentIndex += len(group.Options)
				}
			}
		}
	}

	return "", -1
}

// getFocusedSettingOption resolves the currently focused item including option rows.
// It returns (provider, -1, -1) for provider headers, (provider, groupIdx, -1) for group headers,
// and (provider, groupIdx, optionIdx) for option rows.
func getFocusedSettingOption(providers []string, settingsByProvider map[string][]SettingGroup, expandedProviders map[string]bool, focusedIndex int) (string, int, int) {
	currentIndex := 0
	multiProvider := len(providers) > 1
	for _, provider := range providers {
		if multiProvider {
			if currentIndex == focusedIndex {
				return provider, -1, -1
			}
			currentIndex++
		}

		if expandedProviders[provider] || !multiProvider {
			groups := settingsByProvider[provider]
			for groupIdx, group := range groups {
				if currentIndex == focusedIndex {
					return provider, groupIdx, -1
				}
				currentIndex++

				for optionIdx := range group.Options {
					if currentIndex == focusedIndex {
						return provider, groupIdx, optionIdx
					}
					currentIndex++
				}
			}
		}
	}

	return "", -1, -1
}

// ----------------------------------- Common: Settings Stage Helpers -----------------------------------

func moveSelectionFocus(focused *int, max int, key string) {
	switch key {
	case keyUp:
		if *focused > 0 {
			*focused--
		}
	case keyDown:
		if *focused < max {
			*focused++
		}
	}
}

func toggleFocusedSetting(groups []SettingGroup, selected map[string]bool, selectedGroups map[string]bool, provider string, groupIdx, optionIdx int) {
	if groupIdx < 0 || groupIdx >= len(groups) {
		return
	}

	group := groups[groupIdx]
	groupKey := settingGroupKey(provider, group.Key)

	if optionIdx == -1 {
		if len(group.Options) == 0 {
			selectedGroups[groupKey] = !selectedGroups[groupKey]
			return
		}

		allSelected := true
		for _, option := range group.Options {
			optionKey := settingOptionKey(provider, group.Key, option)
			if !selected[optionKey] {
				allSelected = false
				break
			}
		}

		for _, option := range group.Options {
			optionKey := settingOptionKey(provider, group.Key, option)
			selected[optionKey] = !allSelected
		}
		selectedGroups[groupKey] = !allSelected
		return
	}

	if optionIdx >= len(group.Options) {
		return
	}

	optionKey := settingOptionKey(provider, group.Key, group.Options[optionIdx])
	selected[optionKey] = !selected[optionKey]

	allSelected := true
	for _, option := range group.Options {
		optKey := settingOptionKey(provider, group.Key, option)
		if !selected[optKey] {
			allSelected = false
			break
		}
	}
	selectedGroups[groupKey] = allSelected
}

func toggleAllSettings(providers []string, settingsByProvider map[string][]SettingGroup, selected map[string]bool, selectedGroups map[string]bool) {
	allSelected := true
	for _, provider := range providers {
		groups := settingsByProvider[provider]
		for _, group := range groups {
			for _, option := range group.Options {
				selectionKey := settingOptionKey(provider, group.Key, option)
				if !selected[selectionKey] {
					allSelected = false
					break
				}
			}
			if !allSelected {
				break
			}
		}
		if !allSelected {
			break
		}
	}

	for _, provider := range providers {
		groups := settingsByProvider[provider]
		for _, group := range groups {
			groupKey := settingGroupKey(provider, group.Key)
			for _, option := range group.Options {
				selectionKey := settingOptionKey(provider, group.Key, option)
				selected[selectionKey] = !allSelected
			}
			selectedGroups[groupKey] = !allSelected
		}
	}
}

func getSettingsMaxFocusIndex(providers []string, settingsByProvider map[string][]SettingGroup, expandedProviders map[string]bool) int {
	count := 0
	multiProvider := len(providers) > 1
	for _, provider := range providers {
		if multiProvider {
			count++ // Provider header (only counted when visible)
		}
		if expandedProviders[provider] || !multiProvider {
			groups := settingsByProvider[provider]
			for _, group := range groups {
				count++                     // Setting group header
				count += len(group.Options) // Options (always shown)
			}
		}
	}
	if count == 0 {
		return 0
	}
	return count - 1
}

func setProviderExpanded(provider string, groupIdx int, expandedProviders map[string]bool, expanded bool) {
	if provider != "" && groupIdx == -1 {
		expandedProviders[provider] = expanded
	}
}

func selectedProviderNames(providers []Provider) []string {
	selectedProviders := make([]string, 0)
	for _, provider := range providers {
		if provider.Selected {
			selectedProviders = append(selectedProviders, provider.DisplayName)
		}
	}
	return selectedProviders
}

func hasAnySelected(selection map[string]bool) bool {
	for _, selected := range selection {
		if selected {
			return true
		}
	}
	return false
}

// handleSharedSettingsKeys processes the common settings-stage keys:
// up/down, left/right, space, select-all, pgup/pgdown, home/end.
// It returns true when the key was handled.
func handleSharedSettingsKeys(
	key string,
	focusedIndex *int,
	maxIndex int,
	getFocusedItem func() (string, int),
	getFocusedOption func() (string, int, int),
	expandedProviders map[string]bool,
	providers []string,
	settingsByProvider map[string][]SettingGroup,
	selected map[string]bool,
	selectedGroups map[string]bool,
	vp *SettingsViewport,
) bool {
	switch key {
	case keyUp, keyDown:
		moveSelectionFocus(focusedIndex, maxIndex, key)
		vp.ensureFocusVisible(*focusedIndex)
		return true
	case keyRight:
		provider, groupIdx := getFocusedItem()
		setProviderExpanded(provider, groupIdx, expandedProviders, true)
		return true
	case keyLeft:
		provider, groupIdx := getFocusedItem()
		setProviderExpanded(provider, groupIdx, expandedProviders, false)
		return true
	case keySpace:
		provider, groupIdx := getFocusedItem()
		if isProviderHeaderFocused(provider, groupIdx) {
			return true
		}

		provider, groupIdx, optionIdx := getFocusedOption()
		if provider == "" {
			return true
		}

		groups := settingsByProvider[provider]
		toggleFocusedSetting(groups, selected, selectedGroups, provider, groupIdx, optionIdx)
		return true
	case keySelectAll:
		toggleAllSettings(providers, settingsByProvider, selected, selectedGroups)
		return true
	case "pgup":
		jump := vp.PageSize
		if jump <= 0 {
			jump = 10
		}
		*focusedIndex -= jump
		if *focusedIndex < 0 {
			*focusedIndex = 0
		}
		vp.ensureFocusVisible(*focusedIndex)
		return true
	case "pgdown":
		jump := vp.PageSize
		if jump <= 0 {
			jump = 10
		}
		*focusedIndex += jump
		if *focusedIndex > maxIndex {
			*focusedIndex = maxIndex
		}
		vp.ensureFocusVisible(*focusedIndex)
		return true
	case "home":
		*focusedIndex = 0
		vp.ensureFocusVisible(*focusedIndex)
		return true
	case "end":
		*focusedIndex = maxIndex
		vp.ensureFocusVisible(*focusedIndex)
		return true
	}

	return false
}

// ensureFocusVisible adjusts the scroll offset so the focused row is within the visible page.
func (vp *SettingsViewport) ensureFocusVisible(focusedIndex int) {
	if vp.PageSize <= 0 {
		return
	}
	if focusedIndex < vp.ScrollOffset {
		vp.ScrollOffset = focusedIndex
	}
	if focusedIndex >= vp.ScrollOffset+vp.PageSize {
		vp.ScrollOffset = focusedIndex - vp.PageSize + 1
	}
	if vp.ScrollOffset < 0 {
		vp.ScrollOffset = 0
	}
}

// visibleRange returns the start (inclusive) and end (exclusive) indices of visible rows.
func (vp *SettingsViewport) visibleRange(totalRows int) (int, int) {
	if vp.PageSize <= 0 || totalRows == 0 {
		return 0, totalRows
	}
	start := vp.ScrollOffset
	if start >= totalRows {
		start = totalRows - 1
	}
	if start < 0 {
		start = 0
	}
	end := start + vp.PageSize
	if end > totalRows {
		end = totalRows
	}
	return start, end
}

// updatePageSize recalculates the page size based on available UI height.
// reservedLines accounts for title, help bar, borders, scroll indicators, etc.
func (vp *SettingsViewport) updatePageSize(uiHeight, reservedLines int) {
	vp.PageSize = uiHeight - reservedLines
	if vp.PageSize < 3 {
		vp.PageSize = 3
	}
}

// scrollIndicators returns (showUp, showDown) booleans for scroll arrows.
func (vp *SettingsViewport) scrollIndicators(totalRows int) (bool, bool) {
	showUp := vp.ScrollOffset > 0
	showDown := vp.ScrollOffset+vp.PageSize < totalRows
	return showUp, showDown
}

// ----------------------------------- Common: Stage Transition Helpers -----------------------------------

func (m *Model) buildDatacenterSettingsFetchCmds(selectedProviders []string) []tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(selectedProviders))
	for _, provider := range selectedProviders {
		if ps, ok := m.datacenterSettingsSelection.ProviderSettings[provider]; ok {
			ps.LoadingSettings = true
			ps.SettingsFetchError = ""
			m.datacenterSettingsSelection.ProviderSettings[provider] = ps
		}
		cmds = append(cmds, m.fetchDatacenterSettingsForProvider(provider))
	}
	return cmds
}

func (m *Model) buildMachineSettingsFetchCmds(selectedProviders []string) []tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(selectedProviders))
	for _, provider := range selectedProviders {
		if ps, ok := m.machineDeploymentSettingsSelection.ProviderSettings[provider]; ok {
			ps.LoadingSettings = true
			ps.SettingsFetchError = ""
			m.machineDeploymentSettingsSelection.ProviderSettings[provider] = ps
		}
		cmds = append(cmds, m.fetchMachineSettingsForProvider(provider))
	}
	return cmds
}

// ----------------------------------- Common: Execution & Log Helpers -----------------------------------

func (m *Model) refreshLogsViewport() {
	m.Review.Viewport.SetContent(strings.Join(m.logs, "\n"))
	m.Review.Viewport.GotoBottom()
}

func (m *Model) appendLogLine(line string) {
	m.logs = append(m.logs, line)
	m.refreshLogsViewport()
}

func (m *Model) appendNonEmptyOutput(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line != "" {
			m.logs = append(m.logs, line)
		}
	}
	m.refreshLogsViewport()
}

func (m *Model) startExecutionCancellation() tea.Cmd {
	m.executionCancelling = true
	m.executionError = "Test execution cancelled by user"
	m.quitConfirmVisible = false
	m.logs = append(m.logs, "\n⚠️  Cancellation requested - cleaning up resources...")
	if m.Review.Viewport.Width > 0 {
		m.refreshLogsViewport()
	}
	return m.cleanupTestExecution()
}

// ----------------------------------- Common: Window, Messages & Quit -----------------------------------

// handleWindowSize manages viewport resizing.
func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) tea.Cmd {
	m.terminalWidth = msg.Width
	m.terminalHeight = msg.Height
	viewportWidth := m.getUIInnerWidth()
	viewportHeight := m.getUIHeight() - 4
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	m.Review.Viewport.Width = viewportWidth
	m.Review.Viewport.Height = viewportHeight

	// Recalculate settings pagination page sizes on resize
	const settingsReservedLines = 10 // title + description + borders + scroll indicators + help bar
	m.datacenterViewport.updatePageSize(m.getUIHeight(), settingsReservedLines)
	m.clusterViewport.updatePageSize(m.getUIHeight(), settingsReservedLines)
	m.machineViewport.updatePageSize(m.getUIHeight(), settingsReservedLines)
	return nil
}

// handleStart initializes execution.
func (m *Model) handleStart(msg startMsg) tea.Cmd {
	m.cmdChan = msg.ch
	viewportWidth := m.getUIInnerWidth()
	viewportHeight := m.getUIHeight() - 4
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	if m.Review.Viewport.Width == 0 {
		m.Review.Viewport = viewport.New(viewportWidth, viewportHeight)
	} else {
		m.Review.Viewport.Width = viewportWidth
		m.Review.Viewport.Height = viewportHeight
	}
	return streamLogs(m.cmdChan)
}

// handleLog processes log messages.
func (m *Model) handleLog(msg logMsg) tea.Cmd {
	m.appendLogLine(msg.line)
	return streamLogs(m.cmdChan)
}

// // handleError processes error messages.
// func (m *Model) handleError(msg errMsg) tea.Cmd {
// 	m.executionError = fmt.Sprintf("The bootstrapping process for Kubermatic Virtualization has encountered an issue. For more details, please review the log file located at /tmp/%s.", kubeone.DefaultLogFileName)
// 	m.executionDone = true
// 	m.logs = append(m.logs, fmt.Sprintf("[ERROR] %v", msg.err))
// 	m.stage = stageDone
// 	return nil
// }
