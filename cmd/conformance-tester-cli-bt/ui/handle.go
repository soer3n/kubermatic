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
		// Initialize cluster settings for the next stage
		m.clusterSettingsSelection = initializeClusterSettingsSelection()

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

	case keySpace:
		// Check if focused on setting group or option
		groupIdx, optionIdx := m.getClusterFocusedOption()
		if groupIdx < 0 || groupIdx >= len(m.clusterSettingsSelection.SettingGroups) {
			return m, nil
		}

		group := m.clusterSettingsSelection.SettingGroups[groupIdx]
		groupKey := group.Key

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
			optionKey := fmt.Sprintf("%s:%s", groupKey, group.Options[optionIdx])
			m.clusterSettingsSelection.Selected[optionKey] = !m.clusterSettingsSelection.Selected[optionKey]

			// Update group selection state
			allSelected := true
			for _, option := range group.Options {
				optKey := fmt.Sprintf("%s:%s", groupKey, option)
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
		for _, group := range m.clusterSettingsSelection.SettingGroups {
			for _, option := range group.Options {
				selectionKey := fmt.Sprintf("%s:%s", group.Key, option)
				if !m.clusterSettingsSelection.Selected[selectionKey] {
					allSelected = false
					break
				}
			}
			if !allSelected {
				break
			}
		}

		// Toggle all options and groups
		for _, group := range m.clusterSettingsSelection.SettingGroups {
			groupKey := group.Key
			for _, option := range group.Options {
				selectionKey := fmt.Sprintf("%s:%s", groupKey, option)
				m.clusterSettingsSelection.Selected[selectionKey] = !allSelected
			}
			m.clusterSettingsSelection.SelectedGroups[groupKey] = !allSelected
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
	for _, group := range m.clusterSettingsSelection.SettingGroups {
		count++                     // Setting group header
		count += len(group.Options) // Options (always shown)
	}
	return count - 1
}

// getClusterFocusedOption returns (groupIdx, optionIdx) for the focused item.
// Returns (groupIdx, -1) if focused on group header.
func (m Model) getClusterFocusedOption() (int, int) {
	currentIndex := 0
	for groupIdx, group := range m.clusterSettingsSelection.SettingGroups {
		// Group header
		if currentIndex == m.clusterSettingsSelection.FocusedIndex {
			return groupIdx, -1 // On setting group
		}
		currentIndex++

		// Options (always shown since IsExpanded is always true)
		for optionIdx := range group.Options {
			if currentIndex == m.clusterSettingsSelection.FocusedIndex {
				return groupIdx, optionIdx
			}
			currentIndex++
		}
	}
	return -1, -1
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
		// Move to next stage
		m.stage++
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
