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
		if m.environmentFocusIndex > 0 {
			m.environmentFocusIndex--
			m.environmentFieldIndex = 0 // Reset field index when switching environments
		}
		m.updateEnvironmentFocus()
		return m, nil
	case keyDown:
		if m.environmentFocusIndex < 1 {
			m.environmentFocusIndex++
			m.environmentFieldIndex = 0 // Reset field index when switching environments
		}
		m.updateEnvironmentFocus()
		return m, nil
	case keyTab:
		// Move focus forward through fields within the selected environment
		if m.environmentFocusIndex == 0 && m.localEnv.Selected {
			if m.environmentFieldIndex < 3 { // 3 fields (1, 2, 3)
				m.environmentFieldIndex++
			}
		} else if m.environmentFocusIndex == 1 && m.existingEnv.Selected {
			if m.environmentFieldIndex < 4 { // 4 fields (1, 2, 3, 4)
				m.environmentFieldIndex++
			}
		}
		m.updateEnvironmentFocus()
		return m, nil
	case keyShiftTab:
		// Move focus backward through fields within the selected environment
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
		// Toggle selection based on focused index (only when on the checkbox)
		if m.environmentFieldIndex == 0 {
			if m.environmentFocusIndex == 0 {
				m.localEnv.Selected = !m.localEnv.Selected
				if m.localEnv.Selected {
					m.existingEnv.Selected = false
					m.environmentFieldIndex = 1 // Move to first input field
				}
			} else {
				m.existingEnv.Selected = !m.existingEnv.Selected
				if m.existingEnv.Selected {
					m.localEnv.Selected = false
					m.environmentFieldIndex = 1 // Move to first input field
				}
			}
		}
		m.updateEnvironmentFocus()
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
		} else if m.environmentFocusIndex == 1 && m.existingEnv.Selected && m.environmentFieldIndex > 0 {
			switch m.environmentFieldIndex {
			case 1:
				m.existingEnv.KubeconfigPath, cmd = m.existingEnv.KubeconfigPath.Update(msg)
			case 2:
				m.existingEnv.SeedName, cmd = m.existingEnv.SeedName.Update(msg)
			case 3:
				m.existingEnv.PresetName, cmd = m.existingEnv.PresetName.Update(msg)
			case 4:
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
		for _, provider := range m.providers {
			if provider.Selected {
				hasSelection = true
				break
			}
		}
		if hasSelection {
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
		if m.distributionSelection.FocusedIndex < len(m.distributionSelection.Distributions)-1 {
			m.distributionSelection.FocusedIndex++
		}
		return m, nil

	case keySpace:
		// Toggle selection for the focused distribution
		currentDist := m.distributionSelection.Distributions[m.distributionSelection.FocusedIndex]
		m.distributionSelection.Selected[currentDist] = !m.distributionSelection.Selected[currentDist]
		return m, nil

	case keySelectAll:
		// Toggle select/deselect all distributions
		allSelected := true
		for _, dist := range m.distributionSelection.Distributions {
			if !m.distributionSelection.Selected[dist] {
				allSelected = false
				break
			}
		}

		// If all are selected, deselect all; otherwise select all
		for _, dist := range m.distributionSelection.Distributions {
			m.distributionSelection.Selected[dist] = !allSelected
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
			m.stage = stageDatacenterSettingsSelection // Move to next stage
		}
		return m, nil

	case keyEsc:
		m.stage = stageProviderSelection
		return m, nil
	}

	return m, nil
}

// func (m Model) handleNodesConfig(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	var cmd tea.Cmd
// 	m.focusCurrent()

// 	// Node details stage handling
// 	switch msg := msg.(type) {
// 	case tea.KeyMsg:
// 		switch msg.String() {
// 		case keyEsc:
// 			m.stage = stageSelectNodeCount
// 			return m, nil

// 		case keyEnter:
// 			if m.validateAndProceed() {
// 				m.stage = stageCSIToggle
// 			}
// 			return m, nil

// 		// Navigation keys
// 		case keyUp:
// 			if m.Nodes.CurrentField > 0 {
// 				m.Nodes.CurrentField--
// 			}
// 			m.focusCurrent()
// 			return m, nil

// 		case keyDown:
// 			if m.Nodes.CurrentField < 2 {
// 				m.Nodes.CurrentField++
// 			}
// 			m.focusCurrent()
// 			return m, nil

// 		case keyLeft:
// 			if m.Nodes.Current > 0 {
// 				m.Nodes.Current--
// 				m.Nodes.CurrentField = 0 // Reset field on node change
// 				m.focusCurrent()
// 			}
// 			return m, nil

// 		case keyRight:
// 			if m.Nodes.Current < len(m.Nodes.Inputs)-1 {
// 				m.Nodes.Current++
// 				m.Nodes.CurrentField = 0 // Reset field on node change
// 				m.focusCurrent()
// 			}
// 			return m, nil

// 		case keyTab:
// 			m.Nodes.CurrentField = (m.Nodes.CurrentField + 1) % 3
// 			m.focusCurrent()
// 			return m, nil

// 		case keyShiftTab:
// 			m.Nodes.CurrentField = (m.Nodes.CurrentField + 2) % 3
// 			m.focusCurrent()
// 			return m, nil
// 		}

// 		// Handle text input
// 		currentNode := m.Nodes.Inputs[m.Nodes.Current]
// 		switch m.Nodes.CurrentField {
// 		case 0:
// 			currentNode.Address, cmd = currentNode.Address.Update(msg)
// 		case 1:
// 			currentNode.Username, cmd = currentNode.Username.Update(msg)
// 		case 2:
// 			currentNode.SSHKeyPath, cmd = currentNode.SSHKeyPath.Update(msg)
// 		}
// 		m.Nodes.Inputs[m.Nodes.Current] = currentNode
// 		return m, cmd

// 	default:
// 		// Handle non-key messages (like window resize)
// 		currentNode := m.Nodes.Inputs[m.Nodes.Current]
// 		switch m.Nodes.CurrentField {
// 		case 0:
// 			currentNode.Address, cmd = currentNode.Address.Update(msg)
// 		case 1:
// 			currentNode.Username, cmd = currentNode.Username.Update(msg)
// 		case 2:
// 			currentNode.SSHKeyPath, cmd = currentNode.SSHKeyPath.Update(msg)
// 		}
// 		m.Nodes.Inputs[m.Nodes.Current] = currentNode
// 		return m, cmd
// 	}
// }

// func (m Model) handleMetalLBConfiguration(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
// 	switch msg.String() {
// 	case keySpace:
// 		m.MetalLB.Enabled = !m.MetalLB.Enabled
// 		if m.MetalLB.Enabled {
// 			m.MetalLB.Input.Focus()
// 		} else {
// 			m.MetalLB.Input.Blur()
// 		}
// 		return m, nil
// 	case keyEnter:
// 		if m.MetalLB.Enabled {
// 			// Validate IP range input
// 			inputValue := m.MetalLB.Input.Value()
// 			if inputValue == "" {
// 				m.MetalLB.Error = "IP range is required"
// 				return m, nil
// 			}
// 			if !m.validateIPRange(inputValue) {
// 				m.MetalLB.Error = "Invalid IP range format. Use format: 192.168.1.100-192.168.1.150"
// 				return m, nil
// 			}
// 			m.MetalLB.Error = ""
// 			if m.offline {
// 				m.stage = stageContainerRegistry
// 			} else {
// 				m.stage = stageSelectNodeCount
// 			}
// 			return m, nil
// 		}
// 		if m.offline {
// 			m.stage = stageContainerRegistry
// 		} else {
// 			m.stage = stageSelectNodeCount
// 		}
// 		return m, nil
// 	case keyEsc:
// 		m.stage = stageNetworkConfig
// 		return m, nil
// 	default:
// 		// Handle text input when enabled
// 		if m.MetalLB.Enabled {
// 			var cmd tea.Cmd
// 			m.MetalLB.Input, cmd = m.MetalLB.Input.Update(msg)
// 			return m, cmd
// 		}
// 		return m, nil
// 	}
// }

// func (m Model) handleContainerRegistry(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
// 	if !m.offline {
// 		m.stage = stageSelectNodeCount
// 		return m, nil
// 	}

// 	m.updateContainerRegistryFocus()

// 	switch msg.String() {
// 	case keyEsc:
// 		m.stage = stageMetalLB
// 		return m, nil
// 	case keyLeft, keyShiftTab, keyUp:
// 		if m.ContainerRegistry.CurrentField > 0 {
// 			m.ContainerRegistry.CurrentField--
// 			m.updateContainerRegistryFocus()
// 		}
// 		return m, nil
// 	case keyRight, keyTab, keyDown:
// 		if m.ContainerRegistry.CurrentField < 3 {
// 			m.ContainerRegistry.CurrentField++
// 			m.updateContainerRegistryFocus()
// 		}
// 		return m, nil
// 	case keyEnter:
// 		if m.validateContainerRegistry() {
// 			if m.offline {
// 				m.stage = stageHelmRegistry
// 			} else {
// 				m.stage = stageSelectNodeCount
// 			}
// 		}
// 		return m, nil
// 	case keySpace:
// 		if m.ContainerRegistry.CurrentField == 3 {
// 			m.ContainerRegistry.Insecure = !m.ContainerRegistry.Insecure
// 			return m, nil
// 		}
// 	}

// 	var cmd tea.Cmd
// 	switch m.ContainerRegistry.CurrentField {
// 	case 0:
// 		m.ContainerRegistry.Endpoint, cmd = m.ContainerRegistry.Endpoint.Update(msg)
// 	case 1:
// 		m.ContainerRegistry.Username, cmd = m.ContainerRegistry.Username.Update(msg)
// 	case 2:
// 		m.ContainerRegistry.Password, cmd = m.ContainerRegistry.Password.Update(msg)
// 	}

// 	return m, cmd
// }

// // handleHelmRegistry manages user input for Helm registry configuration.
// func (m Model) handleHelmRegistry(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
// 	if !m.offline {
// 		m.stage = stagePackageRepository
// 		return m, nil
// 	}

// 	m.updateHelmRegistryFocus()

// 	switch msg.String() {
// 	case keyEsc:
// 		m.stage = stageContainerRegistry
// 		return m, nil
// 	case keyLeft, keyShiftTab, keyUp:
// 		if m.HelmRegistry.CurrentField > 0 {
// 			m.HelmRegistry.CurrentField--
// 			m.updateHelmRegistryFocus()
// 		}
// 		return m, nil
// 	case keyRight, keyTab, keyDown:
// 		if m.HelmRegistry.CurrentField < 3 {
// 			m.HelmRegistry.CurrentField++
// 			m.updateHelmRegistryFocus()
// 		}
// 		return m, nil
// 	case keyEnter:
// 		if m.validateHelmRegistry() {
// 			m.stage = stagePackageRepository
// 		}
// 		return m, nil
// 	case keySpace:
// 		if m.HelmRegistry.CurrentField == 3 {
// 			m.HelmRegistry.Insecure = !m.HelmRegistry.Insecure
// 			return m, nil
// 		}
// 	}

// 	var cmd tea.Cmd
// 	switch m.HelmRegistry.CurrentField {
// 	case 0:
// 		m.HelmRegistry.Endpoint, cmd = m.HelmRegistry.Endpoint.Update(msg)
// 	case 1:
// 		m.HelmRegistry.Username, cmd = m.HelmRegistry.Username.Update(msg)
// 	case 2:
// 		m.HelmRegistry.Password, cmd = m.HelmRegistry.Password.Update(msg)
// 	}

// 	return m, cmd
// }

// // handlePackageRepository manages user input for the package repository configuration.
// func (m Model) handlePackageRepository(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
// 	// Update focus state before handling input
// 	m.updatePackageRepoFocus()

// 	switch msg.String() {
// 	case keyEsc:
// 		m.stage = stageHelmRegistry
// 		return m, nil
// 	case keyEnter:
// 		if m.PackageRepo.Enabled && strings.TrimSpace(m.PackageRepo.Address.Value()) == "" {
// 			m.PackageRepo.Error = "Package repository address is required when enabled"
// 		} else {
// 			m.PackageRepo.Error = ""
// 			m.stage = stageSelectNodeCount
// 		}
// 		return m, nil
// 	case keySpace:
// 		m.PackageRepo.Enabled = !m.PackageRepo.Enabled
// 		m.PackageRepo.Error = ""
// 		return m, nil
// 	}

// 	// Only update the address field if the repository is enabled
// 	if m.PackageRepo.Enabled {
// 		var cmd tea.Cmd
// 		m.PackageRepo.Address, cmd = m.PackageRepo.Address.Update(msg)
// 		return m, cmd
// 	}

// 	return m, nil
// }

// func (m Model) handleCSIToggle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
// 	switch msg.String() {
// 	case keySpace:
// 		m.CSIEnabled = !m.CSIEnabled
// 		return m, nil
// 	case keyEnter:
// 		m.stage = stageReview
// 		return m, nil
// 	case keyEsc:
// 		m.stage = stageNodeDetails
// 		return m, nil
// 	}
// 	return m, nil
// }

// func (m Model) handleReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
// 	switch msg.String() {
// 	case "up":
// 		m.Review.Viewport.ScrollUp(1)
// 	case "down":
// 		m.Review.Viewport.ScrollDown(1)
// 	case "pgup":
// 		m.Review.Viewport.HalfPageUp()
// 	case "pgdown":
// 		m.Review.Viewport.HalfPageDown()
// 	case "left", keyEsc:
// 		m.stage = stageCSIToggle
// 	case keyEnter:
// 		m.stage = stageExecuting
// 		return m, m.startExecution()
// 	}
// 	return m, nil
// }

// // Replace startExecution to use BootstrapCluster.
// func (m *Model) startExecution() tea.Cmd {
// 	return func() tea.Msg {
// 		ch := make(chan string)
// 		msgCh := make(chan tea.Msg)
// 		// Forward string lines as logMsg to msgCh
// 		go func() {
// 			for line := range ch {
// 				msgCh <- logMsg{line: line}
// 			}
// 			close(msgCh)
// 		}()
// 		// go func() {
// 		// 	writer := &tuiLogWriter{ch: ch}
// 		// 	config := getKubeVConfig(m)
// 		// 	err := kubeone.BootstrapCluster(writer, config)
// 		// 	if err != nil {
// 		// 		ch <- "[ERROR] " + err.Error()
// 		// 		msgCh <- errMsg{err: err}
// 		// 	}
// 		// 	close(ch)
// 		// }()
// 		return startMsg{ch: msgCh}
// 	}
// }

// handleWindowSize manages viewport resizing.
func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) tea.Cmd {
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

// func (m Model) handleNodeCountSelection(msg tea.KeyMsg) (Model, tea.Cmd) {
// 	m.updateNodeCountFocus()

// 	switch msg.String() {
// 	case keyEsc, "left":
// 		if m.offline {
// 			m.stage = stagePackageRepository
// 		} else {
// 			m.stage = stageMetalLB
// 		}
// 		return m, nil
// 	case keyEnter:
// 		// Validate and proceed
// 		if m.processNodeCountInput() {
// 			m.stage = stageNodeDetails
// 		}
// 		return m, nil
// 	case keyUp, keyShiftTab:
// 		if m.NodeCount.CurrentField > 0 {
// 			m.NodeCount.CurrentField--
// 			m.updateNodeCountFocus()
// 		}
// 		return m, nil
// 	case keyDown, keyTab:
// 		if m.NodeCount.CurrentField < 2 {
// 			m.NodeCount.CurrentField++
// 			m.updateNodeCountFocus()
// 		}
// 		return m, nil
// 	default:
// 		// Handle text input based on current field
// 		var cmd tea.Cmd
// 		switch m.NodeCount.CurrentField {
// 		case 0:
// 			m.NodeCount.NodeCountInput, cmd = m.NodeCount.NodeCountInput.Update(msg)
// 		case 1:
// 			m.NodeCount.ControlPlaneCountInput, cmd = m.NodeCount.ControlPlaneCountInput.Update(msg)
// 		case 2:
// 			m.NodeCount.APIEndpointInput, cmd = m.NodeCount.APIEndpointInput.Update(msg)
// 		}
// 		return m, cmd
// 	}
// }

func streamLogs(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return doneMsg{success: true}
		}
		return msg
	}
}

// // Initialize nodes based on count.
// func (m *Model) initializeNodes(n int) {
// 	m.Nodes.Configs = make([]NodeConfig, n)
// 	m.Nodes.Inputs = make([]NodeInputFields, n)
// 	for i := range m.Nodes.Inputs {
// 		m.Nodes.Inputs[i] = NodeInputFields{
// 			Address:    newTextInput("Address", 64),
// 			Username:   newTextInput("Username", 32),
// 			SSHKeyPath: newTextInput("SSH Key Path", 256),
// 		}
// 	}
// 	m.Nodes.Current = 0
// }

// func (m *Model) focusCurrent() {
// 	// Safety checks
// 	if len(m.Nodes.Inputs) == 0 {
// 		return
// 	}

// 	// Clamp indexes to valid ranges
// 	m.Nodes.Current = clamp(m.Nodes.Current, 0, len(m.Nodes.Inputs)-1)
// 	m.Nodes.CurrentField = clamp(m.Nodes.CurrentField, 0, 2)

// 	currentNode := m.Nodes.Inputs[m.Nodes.Current]

// 	// Blur all fields first
// 	currentNode.Address.Blur()
// 	currentNode.Username.Blur()
// 	currentNode.SSHKeyPath.Blur()

// 	// Focus current field
// 	switch m.Nodes.CurrentField {
// 	case 0:
// 		currentNode.Address.Focus()
// 	case 1:
// 		currentNode.Username.Focus()
// 	case 2:
// 		currentNode.SSHKeyPath.Focus()
// 	}

// 	m.Nodes.Inputs[m.Nodes.Current] = currentNode
// }

// // clamp ensures a value stays within min/max bounds.
// //
// //nolint:predeclared
// func clamp(v, min, max int) int {
// 	if v < min {
// 		return min
// 	}
// 	if v > max {
// 		return max
// 	}
// 	return v
// }

// // func getKubeVConfig(m *Model) kubeone.KubeVConfig {
// // 	config := kubeone.KubeVConfig{
// // 		DefaultCSIEnabled: m.CSIEnabled,
// // 		CPCount:           mustAtoi(m.NodeCount.ControlPlaneCountInput.Value()),
// // 		APIEndpoint:       m.NodeCount.APIEndpointInput.Value(),
// // 		DefaultLBEnabled:  m.MetalLB.Enabled,
// // 		LoadBalancerRange: m.MetalLB.Input.Value(),
// // 		NetworkConfig: kubeone.NetworkConfig{
// // 			NetworkCIDR: m.Network.CIDR.Value(),
// // 			GatewayIP:   m.Network.GatewayIP.Value(),
// // 			DNSServerIP: m.Network.DNSServer.Value(),
// // 		},
// // 		OfflineSettings: kubeone.OfflineSettings{
// // 			Enabled: false,
// // 		},
// // 	}

// // 	for _, n := range m.Nodes.Configs {
// // 		node := kubeone.NodeConfig{
// // 			Address:    n.Address,
// // 			SSHKeyPath: n.SSHKeyPath,
// // 			Username:   n.Username,
// // 		}
// // 		config.Nodes = append(config.Nodes, node)
// // 	}

// // 	if m.offline {
// // 		config.OfflineSettings = kubeone.OfflineSettings{
// // 			Enabled: true,
// // 			ContainerRegistry: kubeone.OCIConfiguration{
// // 				Address:  m.ContainerRegistry.Endpoint.Value(),
// // 				Username: m.ContainerRegistry.Username.Value(),
// // 				Password: m.ContainerRegistry.Password.Value(),
// // 				Insecure: m.ContainerRegistry.Insecure,
// // 			},
// // 			HelmRegistry: kubeone.OCIConfiguration{
// // 				Address:  normalizeRegistryBase(m.HelmRegistry.Endpoint.Value()),
// // 				Username: m.HelmRegistry.Username.Value(),
// // 				Password: m.HelmRegistry.Password.Value(),
// // 				Insecure: m.HelmRegistry.Insecure,
// // 			},
// // 			PackageRepository: m.PackageRepo.Address.Value(),
// // 		}
// // 	}
// // 	return config
// // }

// func (m *Model) validateAndProceed() bool {
// 	allFilled := true

// 	// Validate all required fields are filled
// 	for _, node := range m.Nodes.Inputs {
// 		if node.Address.Value() == "" ||
// 			node.Username.Value() == "" ||
// 			node.SSHKeyPath.Value() == "" {
// 			allFilled = false
// 			break
// 		}
// 	}

// 	if allFilled {
// 		// Convert input models to config structs
// 		configs := make([]NodeConfig, len(m.Nodes.Inputs))
// 		for i, input := range m.Nodes.Inputs {
// 			configs[i] = NodeConfig{
// 				Address:    input.Address.Value(),
// 				Username:   input.Username.Value(),
// 				SSHKeyPath: input.SSHKeyPath.Value(),
// 			}
// 		}
// 		m.Nodes.Configs = configs

// 		// Generate and set review configuration
// 		yamlContent := generateReviewConfig(*m)
// 		m.Review.ConfigYAML = yamlContent

// 		// Initialize viewport with generated YAML
// 		m.InitViewport(yamlContent, 80, 15)
// 		m.stage = stageCSIToggle
// 		return true
// 	}
// 	return false
// }

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
