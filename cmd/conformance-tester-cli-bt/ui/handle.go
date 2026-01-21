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

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	keyEsc       = "esc"
	keyEnter     = "enter"
	keyRight     = "right"
	keyLeft      = "left"
	keyUp        = "up"
	keyDown      = "down"
	keyTab       = "tab"
	keyShiftTab  = "shift+tab"
	keyYes       = "y"
	keyNo        = "n"
	keyControlC  = "ctrl+c"
	keyQuit      = "q"
	keySpace     = " "
	keySelectAll = "ctrl+a"
	digits       = "0123456789"
)

func (m Model) handleWelcomePage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyEnter:
		m.stage = stageEnvironmentSelection
		return m, nil
	}
	return m, nil
}

func (m Model) handleEnvironmentSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func (m Model) handleReleaseSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

			// Check if all minors in this major are selected
			allMinorsSelected := true
			for _, minor := range minorVersions {
				if !m.releaseSelection.SelectedMinor[minor] {
					allMinorsSelected = false
					break
				}
			}
			m.releaseSelection.SelectedMajor[currentMajor] = allMinorsSelected
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

func (m Model) handleProviderSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case keyUp:
		if m.providerFocusIndex > 0 {
			m.providerFocusIndex--
			m.providerFieldIndex = 0 // Reset field index when switching providers
		}
		m.updateProviderFocus()
		return m, nil
	case keyDown:
		if m.providerFocusIndex < len(m.providers)-1 {
			m.providerFocusIndex++
			m.providerFieldIndex = 0 // Reset field index when switching providers
		}
		m.updateProviderFocus()
		return m, nil
	case keyTab:
		// Move focus forward through fields within the selected provider
		if m.providers[m.providerFocusIndex].Selected {
			maxField := m.getMaxFieldIndexForProvider(m.providers[m.providerFocusIndex])
			if m.providerFieldIndex < maxField {
				m.providerFieldIndex++
			}
		}
		m.updateProviderFocus()
		return m, nil
	case keyShiftTab:
		// Move focus backward through fields within the selected provider
		if m.providerFieldIndex > 0 {
			m.providerFieldIndex--
		}
		m.updateProviderFocus()
		return m, nil
	case keyEnter:
		// Proceed to next stage if at least one provider is selected
		hasSelection := false
		var selectedProviders []string
		for _, provider := range m.providers {
			if provider.Selected {
				hasSelection = true
				selectedProviders = append(selectedProviders, provider.DisplayName)
			}
		}
		if hasSelection {
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
				m.providerFieldIndex = 1 // Move to first credential field
			}
		}
		m.updateProviderFocus()
		return m, nil
	case keyEsc:
		m.stage = stageEnvironmentSelection
		return m, nil
	default:
		// Update the focused text input
		if m.providers[m.providerFocusIndex].Selected && m.providerFieldIndex > 0 {
			m.providers[m.providerFocusIndex], cmd = m.updateProviderCredentialField(m.providers[m.providerFocusIndex], m.providerFieldIndex, msg)
		}
		return m, cmd
	}
}

// getMaxFieldIndexForProvider returns the maximum field index for a provider based on its credential type.
func (m Model) getMaxFieldIndexForProvider(provider Provider) int {
	switch provider.Credentials.(type) {
	case GCPCredentials, AnexiaCredentials, DigitalOceanCredentials, HetznerCredentials, KubeVirtCredentials:
		return 1
	case AlibabaCredentials, VSphereCredentials:
		return 2
	case AWSCredentials, AzureCredentials:
		return 4
	case VMwareCloudDirectorCredentials:
		return 5
	case NutanixCredentials:
		return 7
	case OpenStackCredentials:
		return 8
	default:
		return 1
	}
}

// updateProviderCredentialField updates a specific credential field for a provider.
func (m Model) updateProviderCredentialField(provider Provider, fieldIndex int, msg tea.KeyMsg) (Provider, tea.Cmd) {
	var cmd tea.Cmd

	switch creds := provider.Credentials.(type) {
	case AWSCredentials:
		switch fieldIndex {
		case 1:
			creds.AccessKeyID, cmd = creds.AccessKeyID.Update(msg)
		case 2:
			creds.SecretAccessKey, cmd = creds.SecretAccessKey.Update(msg)
		case 3:
			creds.AssumeRoleARN, cmd = creds.AssumeRoleARN.Update(msg)
		case 4:
			creds.AssumeRoleExternalID, cmd = creds.AssumeRoleExternalID.Update(msg)
		}
		provider.Credentials = creds

	case AzureCredentials:
		switch fieldIndex {
		case 1:
			creds.TenantID, cmd = creds.TenantID.Update(msg)
		case 2:
			creds.SubscriptionID, cmd = creds.SubscriptionID.Update(msg)
		case 3:
			creds.ClientID, cmd = creds.ClientID.Update(msg)
		case 4:
			creds.ClientSecret, cmd = creds.ClientSecret.Update(msg)
		}
		provider.Credentials = creds

	case GCPCredentials:
		if fieldIndex == 1 {
			creds.ServiceAccount, cmd = creds.ServiceAccount.Update(msg)
			provider.Credentials = creds
		}

	case AlibabaCredentials:
		switch fieldIndex {
		case 1:
			creds.AccessKeyID, cmd = creds.AccessKeyID.Update(msg)
		case 2:
			creds.AccessKeySecret, cmd = creds.AccessKeySecret.Update(msg)
		}
		provider.Credentials = creds

	case AnexiaCredentials:
		if fieldIndex == 1 {
			creds.Token, cmd = creds.Token.Update(msg)
			provider.Credentials = creds
		}

	case DigitalOceanCredentials:
		if fieldIndex == 1 {
			creds.Token, cmd = creds.Token.Update(msg)
			provider.Credentials = creds
		}

	case HetznerCredentials:
		if fieldIndex == 1 {
			creds.Token, cmd = creds.Token.Update(msg)
			provider.Credentials = creds
		}

	case KubeVirtCredentials:
		if fieldIndex == 1 {
			creds.Kubeconfig, cmd = creds.Kubeconfig.Update(msg)
			provider.Credentials = creds
		}

	case NutanixCredentials:
		switch fieldIndex {
		case 1:
			creds.Username, cmd = creds.Username.Update(msg)
		case 2:
			creds.Password, cmd = creds.Password.Update(msg)
		case 3:
			creds.ClusterName, cmd = creds.ClusterName.Update(msg)
		case 4:
			creds.ProxyURL, cmd = creds.ProxyURL.Update(msg)
		case 5:
			creds.CSIUsername, cmd = creds.CSIUsername.Update(msg)
		case 6:
			creds.CSIPassword, cmd = creds.CSIPassword.Update(msg)
		case 7:
			creds.CSIEndpoint, cmd = creds.CSIEndpoint.Update(msg)
		}
		provider.Credentials = creds

	case OpenStackCredentials:
		switch fieldIndex {
		case 1:
			creds.Username, cmd = creds.Username.Update(msg)
		case 2:
			creds.Password, cmd = creds.Password.Update(msg)
		case 3:
			creds.Project, cmd = creds.Project.Update(msg)
		case 4:
			creds.ProjectID, cmd = creds.ProjectID.Update(msg)
		case 5:
			creds.Domain, cmd = creds.Domain.Update(msg)
		case 6:
			creds.ApplicationCredentialID, cmd = creds.ApplicationCredentialID.Update(msg)
		case 7:
			creds.ApplicationCredentialSecret, cmd = creds.ApplicationCredentialSecret.Update(msg)
		case 8:
			creds.Token, cmd = creds.Token.Update(msg)
		}
		provider.Credentials = creds

	case VSphereCredentials:
		switch fieldIndex {
		case 1:
			creds.Username, cmd = creds.Username.Update(msg)
		case 2:
			creds.Password, cmd = creds.Password.Update(msg)
		}
		provider.Credentials = creds

	case VMwareCloudDirectorCredentials:
		switch fieldIndex {
		case 1:
			creds.Username, cmd = creds.Username.Update(msg)
		case 2:
			creds.Password, cmd = creds.Password.Update(msg)
		case 3:
			creds.APIToken, cmd = creds.APIToken.Update(msg)
		case 4:
			creds.Organization, cmd = creds.Organization.Update(msg)
		case 5:
			creds.VDC, cmd = creds.VDC.Update(msg)
		}
		provider.Credentials = creds
	}

	return provider, cmd
}

// updateProviderFocus manages focus state for provider selection.
func (m *Model) updateProviderFocus() {
	// Blur all provider text inputs
	for i := range m.providers {
		m.providers[i] = m.blurAllProviderFields(m.providers[i])
	}

	// Focus the current field in the currently focused provider
	if m.providers[m.providerFocusIndex].Selected && m.providerFieldIndex > 0 {
		m.providers[m.providerFocusIndex] = m.focusProviderField(m.providers[m.providerFocusIndex], m.providerFieldIndex)
	}
}

// blurAllProviderFields blurs all credential fields in a provider.
func (m Model) blurAllProviderFields(provider Provider) Provider {
	switch creds := provider.Credentials.(type) {
	case AWSCredentials:
		creds.AccessKeyID.Blur()
		creds.SecretAccessKey.Blur()
		creds.AssumeRoleARN.Blur()
		creds.AssumeRoleExternalID.Blur()
		provider.Credentials = creds
	case AzureCredentials:
		creds.TenantID.Blur()
		creds.SubscriptionID.Blur()
		creds.ClientID.Blur()
		creds.ClientSecret.Blur()
		provider.Credentials = creds
	case GCPCredentials:
		creds.ServiceAccount.Blur()
		provider.Credentials = creds
	case AlibabaCredentials:
		creds.AccessKeyID.Blur()
		creds.AccessKeySecret.Blur()
		provider.Credentials = creds
	case AnexiaCredentials:
		creds.Token.Blur()
		provider.Credentials = creds
	case DigitalOceanCredentials:
		creds.Token.Blur()
		provider.Credentials = creds
	case HetznerCredentials:
		creds.Token.Blur()
		provider.Credentials = creds
	case KubeVirtCredentials:
		creds.Kubeconfig.Blur()
		provider.Credentials = creds
	case NutanixCredentials:
		creds.Username.Blur()
		creds.Password.Blur()
		creds.ClusterName.Blur()
		creds.ProxyURL.Blur()
		creds.CSIUsername.Blur()
		creds.CSIPassword.Blur()
		creds.CSIEndpoint.Blur()
		provider.Credentials = creds
	case OpenStackCredentials:
		creds.Username.Blur()
		creds.Password.Blur()
		creds.Project.Blur()
		creds.ProjectID.Blur()
		creds.Domain.Blur()
		creds.ApplicationCredentialID.Blur()
		creds.ApplicationCredentialSecret.Blur()
		creds.Token.Blur()
		provider.Credentials = creds
	case VSphereCredentials:
		creds.Username.Blur()
		creds.Password.Blur()
		provider.Credentials = creds
	case VMwareCloudDirectorCredentials:
		creds.Username.Blur()
		creds.Password.Blur()
		creds.APIToken.Blur()
		creds.Organization.Blur()
		creds.VDC.Blur()
		provider.Credentials = creds
	}
	return provider
}

// focusProviderField focuses a specific field in provider credentials.
func (m Model) focusProviderField(provider Provider, fieldIndex int) Provider {
	switch creds := provider.Credentials.(type) {
	case AWSCredentials:
		switch fieldIndex {
		case 1:
			creds.AccessKeyID.Focus()
		case 2:
			creds.SecretAccessKey.Focus()
		case 3:
			creds.AssumeRoleARN.Focus()
		case 4:
			creds.AssumeRoleExternalID.Focus()
		}
		provider.Credentials = creds
	case AzureCredentials:
		switch fieldIndex {
		case 1:
			creds.TenantID.Focus()
		case 2:
			creds.SubscriptionID.Focus()
		case 3:
			creds.ClientID.Focus()
		case 4:
			creds.ClientSecret.Focus()
		}
		provider.Credentials = creds
	case GCPCredentials:
		if fieldIndex == 1 {
			creds.ServiceAccount.Focus()
			provider.Credentials = creds
		}
	case AlibabaCredentials:
		switch fieldIndex {
		case 1:
			creds.AccessKeyID.Focus()
		case 2:
			creds.AccessKeySecret.Focus()
		}
		provider.Credentials = creds
	case AnexiaCredentials:
		if fieldIndex == 1 {
			creds.Token.Focus()
			provider.Credentials = creds
		}
	case DigitalOceanCredentials:
		if fieldIndex == 1 {
			creds.Token.Focus()
			provider.Credentials = creds
		}
	case HetznerCredentials:
		if fieldIndex == 1 {
			creds.Token.Focus()
			provider.Credentials = creds
		}
	case KubeVirtCredentials:
		if fieldIndex == 1 {
			creds.Kubeconfig.Focus()
			provider.Credentials = creds
		}
	case NutanixCredentials:
		switch fieldIndex {
		case 1:
			creds.Username.Focus()
		case 2:
			creds.Password.Focus()
		case 3:
			creds.ClusterName.Focus()
		case 4:
			creds.ProxyURL.Focus()
		case 5:
			creds.CSIUsername.Focus()
		case 6:
			creds.CSIPassword.Focus()
		case 7:
			creds.CSIEndpoint.Focus()
		}
		provider.Credentials = creds
	case OpenStackCredentials:
		switch fieldIndex {
		case 1:
			creds.Username.Focus()
		case 2:
			creds.Password.Focus()
		case 3:
			creds.Project.Focus()
		case 4:
			creds.ProjectID.Focus()
		case 5:
			creds.Domain.Focus()
		case 6:
			creds.ApplicationCredentialID.Focus()
		case 7:
			creds.ApplicationCredentialSecret.Focus()
		case 8:
			creds.Token.Focus()
		}
		provider.Credentials = creds
	case VSphereCredentials:
		switch fieldIndex {
		case 1:
			creds.Username.Focus()
		case 2:
			creds.Password.Focus()
		}
		provider.Credentials = creds
	case VMwareCloudDirectorCredentials:
		switch fieldIndex {
		case 1:
			creds.Username.Focus()
		case 2:
			creds.Password.Focus()
		case 3:
			creds.APIToken.Focus()
		case 4:
			creds.Organization.Focus()
		case 5:
			creds.VDC.Focus()
		}
		provider.Credentials = creds
	}
	return provider
}

func (m Model) handleDistributionSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		// Check if at least one distribution is selected
		hasSelection := false
		for _, selected := range m.distributionSelection.Selected {
			if selected {
				hasSelection = true
				break
			}
		}

		if hasSelection {
			// Get all selected providers to initialize datacenter settings
			var selectedProviders []string
			for _, provider := range m.providers {
				if provider.Selected {
					selectedProviders = append(selectedProviders, provider.DisplayName)
				}
			}

			// Initialize datacenter settings for all selected providers
			if len(selectedProviders) > 0 {
				m.datacenterSettingsSelection = initializeDatacenterSettingsSelection(selectedProviders)

				m.stage = stageDatacenterSettingsSelection // Move to next stage
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

// handleDatacenterSettingsSelection handles input for the datacenter settings selection stage.
func (m Model) handleDatacenterSettingsSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyUp:
		if m.datacenterSettingsSelection.FocusedIndex > 0 {
			m.datacenterSettingsSelection.FocusedIndex--
		}
		return m, nil

	case keyDown:
		maxIndex := m.getDatacenterMaxFocusIndex()
		if m.datacenterSettingsSelection.FocusedIndex < maxIndex {
			m.datacenterSettingsSelection.FocusedIndex++
		}
		return m, nil

	case keyRight:
		// Expand provider
		provider, groupIdx := m.getDatacenterFocusedItem()
		if provider != "" && groupIdx == -1 {
			// Focused on provider header - expand provider
			m.datacenterSettingsSelection.ExpandedProviders[provider] = true
		}
		return m, nil

	case keyLeft:
		// Collapse provider
		provider, groupIdx := m.getDatacenterFocusedItem()
		if provider != "" && groupIdx == -1 {
			// Focused on provider header - collapse provider
			m.datacenterSettingsSelection.ExpandedProviders[provider] = false
		}
		return m, nil

	case keySpace:
		// Check if focused on provider header
		provider, groupIdx := m.getDatacenterFocusedItem()
		if provider != "" && groupIdx == -1 {
			// Focused on provider header - ignore
			return m, nil
		}

		// Check if focused on setting group or option
		provider, groupIdx, optionIdx := m.getDatacenterFocusedOption()
		if provider == "" {
			return m, nil
		}

		groups := m.datacenterSettingsSelection.SettingsByProvider[provider]
		if groupIdx >= len(groups) {
			return m, nil
		}

		group := groups[groupIdx]
		groupKey := fmt.Sprintf("%s:%s", provider, group.Key)

		if optionIdx == -1 {
			// Focused on setting group - toggle all options
			allSelected := true
			for _, option := range group.Options {
				optionKey := fmt.Sprintf("%s:%s", groupKey, option)
				if !m.datacenterSettingsSelection.Selected[optionKey] {
					allSelected = false
					break
				}
			}

			// Toggle all options
			for _, option := range group.Options {
				optionKey := fmt.Sprintf("%s:%s", groupKey, option)
				m.datacenterSettingsSelection.Selected[optionKey] = !allSelected
			}
			m.datacenterSettingsSelection.SelectedGroups[groupKey] = !allSelected
		} else {
			// Focused on individual option - toggle it
			if optionIdx >= len(group.Options) {
				return m, nil
			}
			optionKey := fmt.Sprintf("%s:%s:%s", provider, group.Key, group.Options[optionIdx])
			m.datacenterSettingsSelection.Selected[optionKey] = !m.datacenterSettingsSelection.Selected[optionKey]

			// Update group selection state
			allSelected := true
			for _, option := range group.Options {
				optKey := fmt.Sprintf("%s:%s:%s", provider, group.Key, option)
				if !m.datacenterSettingsSelection.Selected[optKey] {
					allSelected = false
					break
				}
			}
			m.datacenterSettingsSelection.SelectedGroups[groupKey] = allSelected
		}
		return m, nil

	case keySelectAll:
		// Check if all options are selected
		allSelected := true
		for _, provider := range m.datacenterSettingsSelection.Providers {
			groups := m.datacenterSettingsSelection.SettingsByProvider[provider]
			for _, group := range groups {
				for _, option := range group.Options {
					selectionKey := fmt.Sprintf("%s:%s:%s", provider, group.Key, option)
					if !m.datacenterSettingsSelection.Selected[selectionKey] {
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

		// Toggle all options and groups
		for _, provider := range m.datacenterSettingsSelection.Providers {
			groups := m.datacenterSettingsSelection.SettingsByProvider[provider]
			for _, group := range groups {
				groupKey := fmt.Sprintf("%s:%s", provider, group.Key)
				for _, option := range group.Options {
					selectionKey := fmt.Sprintf("%s:%s:%s", provider, group.Key, option)
					m.datacenterSettingsSelection.Selected[selectionKey] = !allSelected
				}
				m.datacenterSettingsSelection.SelectedGroups[groupKey] = !allSelected
			}
		}
		return m, nil

	case keyEnter:
		// Get all selected providers to initialize cluster settings
		var selectedProviders []string
		for _, provider := range m.providers {
			if provider.Selected {
				selectedProviders = append(selectedProviders, provider.DisplayName)
			}
		}

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
	count := 0
	for _, provider := range m.datacenterSettingsSelection.Providers {
		count++ // Provider header
		if m.datacenterSettingsSelection.ExpandedProviders[provider] {
			groups := m.datacenterSettingsSelection.SettingsByProvider[provider]
			for _, group := range groups {
				count++                     // Setting group header
				count += len(group.Options) // Options (always shown)
			}
		}
	}
	return count - 1
}

// getDatacenterFocusedItem returns (provider, groupIdx) for the focused item.
// Returns ("", -1) if not on provider or group, (provider, -1) if on provider header,
// (provider, groupIdx) if on setting group.
func (m Model) getDatacenterFocusedItem() (string, int) {
	currentIndex := 0
	for _, provider := range m.datacenterSettingsSelection.Providers {
		if currentIndex == m.datacenterSettingsSelection.FocusedIndex {
			return provider, -1 // On provider header
		}
		currentIndex++

		if m.datacenterSettingsSelection.ExpandedProviders[provider] {
			groups := m.datacenterSettingsSelection.SettingsByProvider[provider]
			for groupIdx, group := range groups {
				if currentIndex == m.datacenterSettingsSelection.FocusedIndex {
					return provider, groupIdx // On setting group
				}
				currentIndex++

				if group.IsExpanded {
					currentIndex += len(group.Options) // Skip options
				}
			}
		}
	}
	return "", -1
}

// getDatacenterFocusedOption returns (provider, groupIdx, optionIdx) for the focused option.
// Returns ("", -1, -1) if not focused on an option.
func (m Model) getDatacenterFocusedOption() (string, int, int) {
	currentIndex := 0
	for _, provider := range m.datacenterSettingsSelection.Providers {
		// Provider header
		if currentIndex == m.datacenterSettingsSelection.FocusedIndex {
			return provider, -1, -1 // On provider header
		}
		currentIndex++

		if m.datacenterSettingsSelection.ExpandedProviders[provider] {
			groups := m.datacenterSettingsSelection.SettingsByProvider[provider]
			for groupIdx, group := range groups {
				// Group header
				if currentIndex == m.datacenterSettingsSelection.FocusedIndex {
					return provider, groupIdx, -1 // On setting group
				}
				currentIndex++

				// Options (always shown since IsExpanded is always true)
				for optionIdx := range group.Options {
					if currentIndex == m.datacenterSettingsSelection.FocusedIndex {
						return provider, groupIdx, optionIdx
					}
					currentIndex++
				}
			}
		}
	}
	return "", -1, -1
}

// handleClusterSettingsSelection handles input for the cluster settings selection stage.
func (m Model) handleClusterSettingsSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyUp:
		if m.clusterSettingsSelection.FocusedIndex > 0 {
			m.clusterSettingsSelection.FocusedIndex--
		}
		return m, nil

	case keyDown:
		maxIndex := m.getClusterMaxFocusIndex()
		if m.clusterSettingsSelection.FocusedIndex < maxIndex {
			m.clusterSettingsSelection.FocusedIndex++
		}
		return m, nil

	case keyRight:
		// Expand provider
		provider, groupIdx := m.getClusterFocusedItem()
		if provider != "" && groupIdx == -1 {
			// Focused on provider header - expand provider
			m.clusterSettingsSelection.ExpandedProviders[provider] = true
		}
		return m, nil

	case keyLeft:
		// Collapse provider
		provider, groupIdx := m.getClusterFocusedItem()
		if provider != "" && groupIdx == -1 {
			// Focused on provider header - collapse provider
			m.clusterSettingsSelection.ExpandedProviders[provider] = false
		}
		return m, nil

	case keySpace:
		// Check if focused on provider header
		provider, groupIdx := m.getClusterFocusedItem()
		if provider != "" && groupIdx == -1 {
			// Focused on provider header - ignore
			return m, nil
		}

		// Check if focused on setting group or option
		provider, groupIdx, optionIdx := m.getClusterFocusedOption()
		if provider == "" {
			return m, nil
		}

		groups := m.clusterSettingsSelection.SettingsByProvider[provider]
		if groupIdx >= len(groups) {
			return m, nil
		}

		group := groups[groupIdx]
		groupKey := fmt.Sprintf("%s:%s", provider, group.Key)

		if optionIdx == -1 {
			// Focused on setting group - toggle all options
			allSelected := true
			for _, option := range group.Options {
				optionKey := fmt.Sprintf("%s:%s", groupKey, option)
				if !m.clusterSettingsSelection.Selected[optionKey] {
					allSelected = false
					break
				}
			}

			// Toggle all options
			for _, option := range group.Options {
				optionKey := fmt.Sprintf("%s:%s", groupKey, option)
				m.clusterSettingsSelection.Selected[optionKey] = !allSelected
			}
			m.clusterSettingsSelection.SelectedGroups[groupKey] = !allSelected
		} else {
			// Focused on individual option - toggle it
			if optionIdx >= len(group.Options) {
				return m, nil
			}
			optionKey := fmt.Sprintf("%s:%s:%s", provider, group.Key, group.Options[optionIdx])
			m.clusterSettingsSelection.Selected[optionKey] = !m.clusterSettingsSelection.Selected[optionKey]

			// Update group selection state
			allSelected := true
			for _, option := range group.Options {
				optKey := fmt.Sprintf("%s:%s:%s", provider, group.Key, option)
				if !m.clusterSettingsSelection.Selected[optKey] {
					allSelected = false
					break
				}
			}
			m.clusterSettingsSelection.SelectedGroups[groupKey] = allSelected
		}
		return m, nil

	case keySelectAll:
		// Check if all options are selected
		allSelected := true
		for _, provider := range m.clusterSettingsSelection.Providers {
			groups := m.clusterSettingsSelection.SettingsByProvider[provider]
			for _, group := range groups {
				for _, option := range group.Options {
					selectionKey := fmt.Sprintf("%s:%s:%s", provider, group.Key, option)
					if !m.clusterSettingsSelection.Selected[selectionKey] {
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

		// Toggle all options and groups
		for _, provider := range m.clusterSettingsSelection.Providers {
			groups := m.clusterSettingsSelection.SettingsByProvider[provider]
			for _, group := range groups {
				groupKey := fmt.Sprintf("%s:%s", provider, group.Key)
				for _, option := range group.Options {
					selectionKey := fmt.Sprintf("%s:%s:%s", provider, group.Key, option)
					m.clusterSettingsSelection.Selected[selectionKey] = !allSelected
				}
				m.clusterSettingsSelection.SelectedGroups[groupKey] = !allSelected
			}
		}
		return m, nil

	case keyEnter:
		// Get all selected providers to initialize machine deployment settings
		var selectedProviders []string
		for _, provider := range m.providers {
			if provider.Selected {
				selectedProviders = append(selectedProviders, provider.DisplayName)
			}
		}

		// Initialize machine deployment settings for the selected providers
		if len(selectedProviders) > 0 {
			m.machineDeploymentSettingsSelection = initializeMachineDeploymentSettingsSelection(selectedProviders)
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
	count := 0
	for _, provider := range m.clusterSettingsSelection.Providers {
		count++ // Provider header
		if m.clusterSettingsSelection.ExpandedProviders[provider] {
			groups := m.clusterSettingsSelection.SettingsByProvider[provider]
			for _, group := range groups {
				count++                     // Setting group header
				count += len(group.Options) // Options (always shown)
			}
		}
	}
	return count - 1
}

// getClusterFocusedItem returns (provider, groupIdx) for the focused item.
// Returns ("", -1) if not on provider or group, (provider, -1) if on provider header,
// (provider, groupIdx) if on setting group.
func (m Model) getClusterFocusedItem() (string, int) {
	currentIndex := 0
	for _, provider := range m.clusterSettingsSelection.Providers {
		if currentIndex == m.clusterSettingsSelection.FocusedIndex {
			return provider, -1 // On provider header
		}
		currentIndex++

		if m.clusterSettingsSelection.ExpandedProviders[provider] {
			groups := m.clusterSettingsSelection.SettingsByProvider[provider]
			for groupIdx, group := range groups {
				if currentIndex == m.clusterSettingsSelection.FocusedIndex {
					return provider, groupIdx // On setting group
				}
				currentIndex++

				currentIndex += len(group.Options) // Skip options
			}
		}
	}
	return "", -1
}

// getClusterFocusedOption returns (provider, groupIdx, optionIdx) for the focused option.
// Returns ("", -1, -1) if not focused on an option.
func (m Model) getClusterFocusedOption() (string, int, int) {
	currentIndex := 0
	for _, provider := range m.clusterSettingsSelection.Providers {
		// Provider header
		if currentIndex == m.clusterSettingsSelection.FocusedIndex {
			return provider, -1, -1 // On provider header
		}
		currentIndex++

		if m.clusterSettingsSelection.ExpandedProviders[provider] {
			groups := m.clusterSettingsSelection.SettingsByProvider[provider]
			for groupIdx, group := range groups {
				// Group header
				if currentIndex == m.clusterSettingsSelection.FocusedIndex {
					return provider, groupIdx, -1 // On setting group
				}
				currentIndex++

				// Options (always shown since IsExpanded is always true)
				for optionIdx := range group.Options {
					if currentIndex == m.clusterSettingsSelection.FocusedIndex {
						return provider, groupIdx, optionIdx
					}
					currentIndex++
				}
			}
		}
	}
	return "", -1, -1
}

// handleMachineDeploymentSettingsSelection handles input for the machine deployment settings selection stage.
func (m Model) handleMachineDeploymentSettingsSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyUp:
		if m.machineDeploymentSettingsSelection.FocusedIndex > 0 {
			m.machineDeploymentSettingsSelection.FocusedIndex--
		}
		return m, nil

	case keyDown:
		maxIndex := m.getMachineDeploymentMaxFocusIndex()
		if m.machineDeploymentSettingsSelection.FocusedIndex < maxIndex {
			m.machineDeploymentSettingsSelection.FocusedIndex++
		}
		return m, nil

	case keyRight:
		// Expand provider
		provider, groupIdx := m.getMachineDeploymentFocusedItem()
		if provider != "" && groupIdx == -1 {
			// Focused on provider header - expand provider
			m.machineDeploymentSettingsSelection.ExpandedProviders[provider] = true
		}
		return m, nil

	case keyLeft:
		// Collapse provider
		provider, groupIdx := m.getMachineDeploymentFocusedItem()
		if provider != "" && groupIdx == -1 {
			// Focused on provider header - collapse provider
			m.machineDeploymentSettingsSelection.ExpandedProviders[provider] = false
		}
		return m, nil

	case keySpace:
		// Check if focused on provider header
		provider, groupIdx := m.getMachineDeploymentFocusedItem()
		if provider != "" && groupIdx == -1 {
			// Focused on provider header - ignore
			return m, nil
		}

		// Check if focused on setting group or option
		provider, groupIdx, optionIdx := m.getMachineDeploymentFocusedOption()
		if provider == "" {
			return m, nil
		}

		groups := m.machineDeploymentSettingsSelection.SettingsByProvider[provider]
		if groupIdx >= len(groups) {
			return m, nil
		}

		group := groups[groupIdx]
		groupKey := fmt.Sprintf("%s:%s", provider, group.Key)

		if optionIdx == -1 {
			// Focused on setting group - toggle all options
			allSelected := true
			for _, option := range group.Options {
				optionKey := fmt.Sprintf("%s:%s", groupKey, option)
				if !m.machineDeploymentSettingsSelection.Selected[optionKey] {
					allSelected = false
					break
				}
			}

			// Toggle all options
			for _, option := range group.Options {
				optionKey := fmt.Sprintf("%s:%s", groupKey, option)
				m.machineDeploymentSettingsSelection.Selected[optionKey] = !allSelected
			}
			m.machineDeploymentSettingsSelection.SelectedGroups[groupKey] = !allSelected
		} else {
			// Focused on individual option - toggle it
			if optionIdx >= len(group.Options) {
				return m, nil
			}
			optionKey := fmt.Sprintf("%s:%s:%s", provider, group.Key, group.Options[optionIdx])
			m.machineDeploymentSettingsSelection.Selected[optionKey] = !m.machineDeploymentSettingsSelection.Selected[optionKey]

			// Update group selection state
			allSelected := true
			for _, option := range group.Options {
				optKey := fmt.Sprintf("%s:%s:%s", provider, group.Key, option)
				if !m.machineDeploymentSettingsSelection.Selected[optKey] {
					allSelected = false
					break
				}
			}
			m.machineDeploymentSettingsSelection.SelectedGroups[groupKey] = allSelected
		}
		return m, nil

	case keySelectAll:
		// Check if all options are selected
		allSelected := true
		for _, provider := range m.machineDeploymentSettingsSelection.Providers {
			groups := m.machineDeploymentSettingsSelection.SettingsByProvider[provider]
			for _, group := range groups {
				for _, option := range group.Options {
					selectionKey := fmt.Sprintf("%s:%s:%s", provider, group.Key, option)
					if !m.machineDeploymentSettingsSelection.Selected[selectionKey] {
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

		// Toggle all options and groups
		for _, provider := range m.machineDeploymentSettingsSelection.Providers {
			groups := m.machineDeploymentSettingsSelection.SettingsByProvider[provider]
			for _, group := range groups {
				groupKey := fmt.Sprintf("%s:%s", provider, group.Key)
				for _, option := range group.Options {
					selectionKey := fmt.Sprintf("%s:%s:%s", provider, group.Key, option)
					m.machineDeploymentSettingsSelection.Selected[selectionKey] = !allSelected
				}
				m.machineDeploymentSettingsSelection.SelectedGroups[groupKey] = !allSelected
			}
		}
		return m, nil

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
	count := 0
	for _, provider := range m.machineDeploymentSettingsSelection.Providers {
		count++ // Provider header
		if m.machineDeploymentSettingsSelection.ExpandedProviders[provider] {
			groups := m.machineDeploymentSettingsSelection.SettingsByProvider[provider]
			for _, group := range groups {
				count++                     // Setting group header
				count += len(group.Options) // Options (always shown)
			}
		}
	}
	return count - 1
}

// getMachineDeploymentFocusedItem returns (provider, groupIdx) for the focused item.
// Returns ("", -1) if not on provider or group, (provider, -1) if on provider header,
// (provider, groupIdx) if on setting group.
func (m Model) getMachineDeploymentFocusedItem() (string, int) {
	currentIndex := 0
	for _, provider := range m.machineDeploymentSettingsSelection.Providers {
		if currentIndex == m.machineDeploymentSettingsSelection.FocusedIndex {
			return provider, -1 // On provider header
		}
		currentIndex++

		if m.machineDeploymentSettingsSelection.ExpandedProviders[provider] {
			groups := m.machineDeploymentSettingsSelection.SettingsByProvider[provider]
			for groupIdx, group := range groups {
				if currentIndex == m.machineDeploymentSettingsSelection.FocusedIndex {
					return provider, groupIdx // On setting group
				}
				currentIndex++

				if group.IsExpanded {
					currentIndex += len(group.Options) // Skip options
				}
			}
		}
	}
	return "", -1
}

// getMachineDeploymentFocusedOption returns (provider, groupIdx, optionIdx) for the focused option.
// Returns ("", -1, -1) if not focused on an option.
func (m Model) getMachineDeploymentFocusedOption() (string, int, int) {
	currentIndex := 0
	for _, provider := range m.machineDeploymentSettingsSelection.Providers {
		// Provider header
		if currentIndex == m.machineDeploymentSettingsSelection.FocusedIndex {
			return provider, -1, -1 // On provider header
		}
		currentIndex++

		if m.machineDeploymentSettingsSelection.ExpandedProviders[provider] {
			groups := m.machineDeploymentSettingsSelection.SettingsByProvider[provider]
			for groupIdx, group := range groups {
				// Group header
				if currentIndex == m.machineDeploymentSettingsSelection.FocusedIndex {
					return provider, groupIdx, -1 // On setting group
				}
				currentIndex++

				// Options (always shown since IsExpanded is always true)
				for optionIdx := range group.Options {
					if currentIndex == m.machineDeploymentSettingsSelection.FocusedIndex {
						return provider, groupIdx, optionIdx
					}
					currentIndex++
				}
			}
		}
	}
	return "", -1, -1
}

// handleWindowSize manages viewport resizing.
func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) tea.Cmd {
	m.terminalWidth = msg.Width
	m.terminalHeight = msg.Height
	m.Review.Viewport.Width = msg.Width - 8
	m.Review.Viewport.Height = msg.Height - 10
	return nil
}

// handleStart initializes execution.
func (m *Model) handleStart(msg startMsg) tea.Cmd {
	m.cmdChan = msg.ch
	if m.Review.Viewport.Width == 0 {
		m.Review.Viewport = viewport.New(80, 15)
	}
	return streamLogs(m.cmdChan)
}

// handleLog processes log messages.
func (m *Model) handleLog(msg logMsg) tea.Cmd {
	m.logs = append(m.logs, msg.line)
	m.Review.Viewport.SetContent(strings.Join(m.logs, "\n"))
	m.Review.Viewport.GotoBottom()
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

// handleDone processes completion messages.
func (m *Model) handleDone(_ doneMsg) tea.Cmd {
	// Only mark success if no prior error was recorded
	if m.executionError == "" {
		m.executionDone = true
		m.logs = append(m.logs, "Successfully applied configuration!")
		m.logs = append(m.logs, "[DONE] Process completed")
		m.Review.Viewport.SetContent(strings.Join(m.logs, "\n"))
		m.Review.Viewport.GotoBottom()
		m.stage = stageDone
	}
	return nil
}

// handleExecOutput processes execution output.
func (m *Model) handleExecOutput(msg execOutputMsg) tea.Cmd {
	lines := strings.Split(msg.output, "\n")
	for _, line := range lines {
		if line != "" {
			m.logs = append(m.logs, line)
		}
	}
	if msg.success {
		m.executionDone = true
		m.executionError = msg.output
		m.logs = append(m.logs, "Successfully applied configuration!")
		m.stage = stageDone
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
			return true, tea.Quit
		}
		m.quitConfirmVisible = false
		return true, nil
	case keyEsc, keyNo:
		m.quitConfirmVisible = false
		return true, nil
	case keyYes:
		return true, tea.Quit
	}
	return false, nil
}

// handleClusterConfiguration handles input for the cluster configuration stage.
func (m Model) handleClusterConfiguration(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

	case keyLeft:
		// Collapse category if focused on a category header
		catIdx, _, itemType := m.getClusterConfigFocusedItem()
		if itemType == "category" && catIdx >= 0 && catIdx < len(m.clusterConfiguration.Categories) {
			categoryName := m.clusterConfiguration.Categories[catIdx].Name
			m.clusterConfiguration.ExpandedCategories[categoryName] = false
		}
		return m, nil

	case keyRight:
		// Expand category if focused on a category header
		catIdx, _, itemType := m.getClusterConfigFocusedItem()
		if itemType == "category" && catIdx >= 0 && catIdx < len(m.clusterConfiguration.Categories) {
			categoryName := m.clusterConfiguration.Categories[catIdx].Name
			m.clusterConfiguration.ExpandedCategories[categoryName] = true
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
func (m Model) handleClusterConfigurationEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

// getKubeconfigSectionAtIndex returns the section type ("env", "file", "custom") if the given index
// is a section header, otherwise returns empty string
func (m Model) getKubeconfigSectionAtIndex(index int) string {
	currentIndex := 0

	// Count env options
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

	// Check if index is env section header
	if len(envOptions) > 0 {
		if index == currentIndex {
			return "env"
		}
		currentIndex++
		if m.existingEnv.KubeconfigExpandedSections["env"] {
			currentIndex += len(envOptions)
		}
	}

	// Check if index is file section header
	if len(fileOptions) > 0 {
		if index == currentIndex {
			return "file"
		}
		currentIndex++
		if m.existingEnv.KubeconfigExpandedSections["file"] {
			currentIndex += len(fileOptions)
		}
	}

	// Check if index is custom section header
	if len(customOptions) > 0 {
		if index == currentIndex {
			return "custom"
		}
	}

	return ""
}

// getKubeconfigOptionIndexFromVisualIndex converts a visual index (including headers) to an actual option index
// Returns -1 if the visual index is a header
func (m Model) getKubeconfigOptionIndexFromVisualIndex(visualIndex int) int {
	currentIndex := 0
	optionIndex := 0

	// Count env options
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

	// Process env section
	if len(envOptions) > 0 {
		if visualIndex == currentIndex {
			return -1 // This is the header
		}
		currentIndex++
		if m.existingEnv.KubeconfigExpandedSections["env"] {
			for i := 0; i < len(envOptions); i++ {
				if visualIndex == currentIndex {
					return optionIndex
				}
				currentIndex++
				optionIndex++
			}
		} else {
			optionIndex += len(envOptions)
		}
	}

	// Process file section
	if len(fileOptions) > 0 {
		if visualIndex == currentIndex {
			return -1 // This is the header
		}
		currentIndex++
		if m.existingEnv.KubeconfigExpandedSections["file"] {
			for i := 0; i < len(fileOptions); i++ {
				if visualIndex == currentIndex {
					return optionIndex
				}
				currentIndex++
				optionIndex++
			}
		} else {
			optionIndex += len(fileOptions)
		}
	}

	// Process custom section
	if len(customOptions) > 0 {
		if visualIndex == currentIndex {
			return -1 // This is the header
		}
		currentIndex++
		if m.existingEnv.KubeconfigExpandedSections["custom"] {
			for i := 0; i < len(customOptions); i++ {
				if visualIndex == currentIndex {
					return optionIndex
				}
				currentIndex++
				optionIndex++
			}
		}
	}

	return -1
}

// getMaxKubeconfigVisualIndex returns the maximum visual index (including headers)
func (m Model) getMaxKubeconfigVisualIndex() int {
	count := 0

	// Count env options
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

	// Count env section
	if len(envOptions) > 0 {
		count++ // Header
		if m.existingEnv.KubeconfigExpandedSections["env"] {
			count += len(envOptions)
		}
	}

	// Count file section
	if len(fileOptions) > 0 {
		count++ // Header
		if m.existingEnv.KubeconfigExpandedSections["file"] {
			count += len(fileOptions)
		}
	}

	// Count custom section
	if len(customOptions) > 0 {
		count++ // Header
		if m.existingEnv.KubeconfigExpandedSections["custom"] {
			count += len(customOptions)
		}
	}

	return count - 1 // Convert count to max index
}
