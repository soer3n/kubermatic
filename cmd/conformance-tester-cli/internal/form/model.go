/*
Copyright 2025 The Kubermatic Kubernetes Platform contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package form

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"k8c.io/kubermatic/v2/cmd/conformance-tester-cli/internal/config"
	"k8c.io/kubermatic/v2/pkg/defaulting"
	"k8c.io/machine-controller/sdk/providerconfig"
)

// PageType represents the current page in the form flow
type PageType int

const (
	PageSplash PageType = iota
	PageEnvironment
	PageKKPCredentials
	PageProviders
	PageDistributions
	PageReleases
	PageNamePrefix
	PageClusterSettings
	PageExclusions
	PageConfirmation
)

// FormModel is the main bubbletea model for the conformance tester CLI
type FormModel struct {
	// Form state
	FormData *FormData

	// Navigation
	CurrentPage  PageType
	PageSequence []PageType

	// UI state for current page
	Width        int
	Height       int
	Err          error
	InputBuffer  string
	CursorIdx    int
	PressedEnter bool

	// Page-specific state
	pageEnvOptions     []string
	pageProvidersList  []providerconfig.CloudProvider
	pageDistOptions    []string
	pageExcludeOptions []string

	// Multi-select state
	selectedMap map[string]bool // For multi-select pages
}

// NewFormModel creates a new FormModel with initialized values
func NewFormModel() *FormModel {
	fm := &FormModel{
		FormData:    NewFormData(),
		CurrentPage: PageSplash,
		Width:       80,
		Height:      24,
		selectedMap: make(map[string]bool),
	}

	fm.pageEnvOptions = []string{"Local", "Existing KKP Instance"}

	// Build provider list
	for name := range providerDisplayMap {
		fm.pageProvidersList = append(fm.pageProvidersList, name)
	}
	sort.Slice(fm.pageProvidersList, func(i, j int) bool {
		return providerDisplayMap[fm.pageProvidersList[i]] < providerDisplayMap[fm.pageProvidersList[j]]
	})

	fm.pageDistOptions = []string{"ubuntu", "flatcar", "rhel", "rockylinux"}
	fm.pageExcludeOptions = []string{
		"conformance",
		"storage",
		"loadbalancer",
		"usercluster-controller",
		"usercluster-metrics",
		"pod-and-node-metrics",
		"seccomp-profiles",
		"no-k8s-gcr-images",
		"control-plane-security-context",
		"telemetry",
		"images",
	}

	fm.buildPageSequence()
	return fm
}

// buildPageSequence constructs the flow of pages based on the form data
func (m *FormModel) buildPageSequence() {
	m.PageSequence = []PageType{
		PageSplash,
		PageEnvironment,
	}

	// KKP credentials only shown if KKP environment is selected
	if m.FormData.EnvOpt == "KKP" {
		m.PageSequence = append(m.PageSequence, PageKKPCredentials)
	}

	m.PageSequence = append(m.PageSequence,
		PageProviders,
		PageDistributions,
		PageReleases,
		PageNamePrefix,
		PageClusterSettings,
		PageExclusions,
		PageConfirmation,
	)
}

// Init initializes the model
func (m *FormModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and state updates
func (m *FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyInput(msg)
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

// handleKeyInput processes keyboard input
func (m *FormModel) handleKeyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		return m.prevPage()
	case "enter":
		// Validate current page before moving on
		if m.validateCurrentPage() {
			return m.nextPage()
		}
	default:
		// Handle page-specific input
		m.handlePageInput(msg)
	}
	return m, nil
}

// validateCurrentPage checks if the current page input is valid
func (m *FormModel) validateCurrentPage() bool {
	switch m.CurrentPage {
	case PageEnvironment:
		return m.FormData.EnvOpt != ""
	case PageProviders:
		return len(m.FormData.ProvidersSelected) > 0
	case PageDistributions:
		return len(m.FormData.Dists) > 0
	case PageReleases:
		return len(m.FormData.Releases) > 0
	case PageNamePrefix:
		return strings.TrimSpace(m.FormData.Config.NamePrefix) != ""
	case PageClusterSettings:
		return m.FormData.NodeCountStr != "" && validateInt(m.FormData.NodeCountStr) == nil
	case PageKKPCredentials:
		if m.FormData.EnvOpt == "KKP" {
			return strings.TrimSpace(m.FormData.Config.Seed) != "" &&
				strings.TrimSpace(m.FormData.Config.Preset) != "" &&
				strings.TrimSpace(m.FormData.Config.Project) != ""
		}
		return true
	default:
		return true
	}
}

// handlePageInput handles input specific to each page
func (m *FormModel) handlePageInput(msg tea.KeyMsg) {
	switch m.CurrentPage {
	case PageEnvironment:
		m.handleEnvironmentInput(msg)
	case PageKKPCredentials:
		m.handleKKPCredentialsInput(msg)
	case PageProviders:
		m.handleMultiSelectInput(msg, m.pageProvidersList, func(idx int) string {
			return string(m.pageProvidersList[idx])
		}, &m.FormData.ProvidersSelected)
	case PageDistributions:
		m.handleDistributionsInput(msg)
	case PageReleases:
		m.handleReleasesInput(msg)
	case PageNamePrefix:
		m.handleNamePrefixInput(msg)
	case PageClusterSettings:
		m.handleClusterSettingsInput(msg)
	case PageExclusions:
		m.handleExclusionsInput(msg)
	}
}

// handleEnvironmentInput processes input for environment selection page
func (m *FormModel) handleEnvironmentInput(msg tea.KeyMsg) {
	switch msg.String() {
	case "up", "left":
		if m.CursorIdx > 0 {
			m.CursorIdx--
		}
	case "down", "right":
		if m.CursorIdx < len(m.pageEnvOptions)-1 {
			m.CursorIdx++
		}
	}
	if m.CursorIdx >= 0 && m.CursorIdx < len(m.pageEnvOptions) {
		m.FormData.EnvOpt = m.pageEnvOptions[m.CursorIdx]
	}
}

// handleKKPCredentialsInput processes input for KKP credentials page
func (m *FormModel) handleKKPCredentialsInput(msg tea.KeyMsg) {
	// Simple text input handler
	switch msg.Type {
	case tea.KeyRunes:
		m.InputBuffer += string(msg.Runes)
	case tea.KeyBackspace:
		if len(m.InputBuffer) > 0 {
			m.InputBuffer = m.InputBuffer[:len(m.InputBuffer)-1]
		}
	case tea.KeyTab:
		// Move to next field
		m.CursorIdx = (m.CursorIdx + 1) % 3
		m.saveKKPField()
	}
	// Save current input to the appropriate field
	m.saveKKPField()
}

// saveKKPField saves the current input buffer to the appropriate KKP field
func (m *FormModel) saveKKPField() {
	switch m.CursorIdx {
	case 0:
		m.FormData.Config.Seed = m.InputBuffer
	case 1:
		m.FormData.Config.Preset = m.InputBuffer
	case 2:
		m.FormData.Config.Project = m.InputBuffer
	}
}

// handleMultiSelectInput processes input for multi-select pages
func (m *FormModel) handleMultiSelectInput(msg tea.KeyMsg, options []providerconfig.CloudProvider, getValue func(int) string, target *[]string) {
	switch msg.String() {
	case "up":
		if m.CursorIdx > 0 {
			m.CursorIdx--
		}
	case "down":
		if m.CursorIdx < len(options)-1 {
			m.CursorIdx++
		}
	case " ":
		value := getValue(m.CursorIdx)
		if config.Contains(*target, value) {
			// Remove
			for i, v := range *target {
				if v == value {
					*target = append((*target)[:i], (*target)[i+1:]...)
					break
				}
			}
		} else {
			// Add
			*target = append(*target, value)
		}
	}
}

// handleDistributionsInput processes input for distributions page
func (m *FormModel) handleDistributionsInput(msg tea.KeyMsg) {
	switch msg.String() {
	case "up":
		if m.CursorIdx > 0 {
			m.CursorIdx--
		}
	case "down":
		if m.CursorIdx < len(m.pageDistOptions)-1 {
			m.CursorIdx++
		}
	case " ":
		value := m.pageDistOptions[m.CursorIdx]
		if config.Contains(m.FormData.Dists, value) {
			for i, v := range m.FormData.Dists {
				if v == value {
					m.FormData.Dists = append(m.FormData.Dists[:i], m.FormData.Dists[i+1:]...)
					break
				}
			}
		} else {
			m.FormData.Dists = append(m.FormData.Dists, value)
		}
	}
}

// handleReleasesInput processes input for releases page
func (m *FormModel) handleReleasesInput(msg tea.KeyMsg) {
	releases := defaulting.DefaultKubernetesVersioning.Versions
	switch msg.String() {
	case "up":
		if m.CursorIdx > 0 {
			m.CursorIdx--
		}
	case "down":
		if m.CursorIdx < len(releases)-1 {
			m.CursorIdx++
		}
	case " ":
		value := releases[m.CursorIdx].String()
		if config.Contains(m.FormData.Releases, value) {
			for i, v := range m.FormData.Releases {
				if v == value {
					m.FormData.Releases = append(m.FormData.Releases[:i], m.FormData.Releases[i+1:]...)
					break
				}
			}
		} else {
			m.FormData.Releases = append(m.FormData.Releases, value)
		}
	}
}

// handleNamePrefixInput processes input for name prefix page
func (m *FormModel) handleNamePrefixInput(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyRunes:
		m.FormData.Config.NamePrefix += string(msg.Runes)
	case tea.KeyBackspace:
		if len(m.FormData.Config.NamePrefix) > 0 {
			m.FormData.Config.NamePrefix = m.FormData.Config.NamePrefix[:len(m.FormData.Config.NamePrefix)-1]
		}
	}
}

// handleClusterSettingsInput processes input for cluster settings page
func (m *FormModel) handleClusterSettingsInput(msg tea.KeyMsg) {
	if m.CursorIdx == 0 {
		// Handling node count input
		switch msg.Type {
		case tea.KeyRunes:
			m.FormData.NodeCountStr += string(msg.Runes)
		case tea.KeyBackspace:
			if len(m.FormData.NodeCountStr) > 0 {
				m.FormData.NodeCountStr = m.FormData.NodeCountStr[:len(m.FormData.NodeCountStr)-1]
			}
		case tea.KeyTab:
			m.CursorIdx = 1
		}
	} else {
		// Handling delete confirmation
		switch msg.String() {
		case "up", "left":
			m.FormData.Config.DeleteClusterAfterTests = true
		case "down", "right":
			m.FormData.Config.DeleteClusterAfterTests = false
		case "tab":
			m.CursorIdx = 0
		}
	}
}

// handleExclusionsInput processes input for exclusions page
func (m *FormModel) handleExclusionsInput(msg tea.KeyMsg) {
	switch msg.String() {
	case "up":
		if m.CursorIdx > 0 {
			m.CursorIdx--
		}
	case "down":
		if m.CursorIdx < len(m.pageExcludeOptions)-1 {
			m.CursorIdx++
		}
	case " ":
		value := m.pageExcludeOptions[m.CursorIdx]
		if config.Contains(m.FormData.Excludes, value) {
			for i, v := range m.FormData.Excludes {
				if v == value {
					m.FormData.Excludes = append(m.FormData.Excludes[:i], m.FormData.Excludes[i+1:]...)
					break
				}
			}
		} else {
			m.FormData.Excludes = append(m.FormData.Excludes, value)
		}
	}
}

// nextPage moves to the next page in the sequence
func (m *FormModel) nextPage() (tea.Model, tea.Cmd) {
	idx := -1
	for i, p := range m.PageSequence {
		if p == m.CurrentPage {
			idx = i
			break
		}
	}

	if idx+1 >= len(m.PageSequence) {
		// We're at the last page - finish
		return m, tea.Quit
	}

	m.CursorIdx = 0
	m.InputBuffer = ""
	m.CurrentPage = m.PageSequence[idx+1]

	// Rebuild sequence in case conditionals changed
	oldSeq := m.PageSequence
	m.buildPageSequence()

	// Make sure new page is in the sequence
	found := false
	for _, p := range m.PageSequence {
		if p == m.CurrentPage {
			found = true
			break
		}
	}
	if !found && len(oldSeq) > idx+1 {
		m.CurrentPage = m.PageSequence[idx+1]
	}

	return m, nil
}

// prevPage moves to the previous page
func (m *FormModel) prevPage() (tea.Model, tea.Cmd) {
	idx := -1
	for i, p := range m.PageSequence {
		if p == m.CurrentPage {
			idx = i
			break
		}
	}

	if idx <= 0 {
		return m, nil // Can't go back from first page
	}

	m.CursorIdx = 0
	m.InputBuffer = ""
	m.CurrentPage = m.PageSequence[idx-1]
	return m, nil
}

// View renders the current page
func (m *FormModel) View() string {
	switch m.CurrentPage {
	case PageSplash:
		return m.renderSplashPage()
	case PageEnvironment:
		return m.renderEnvironmentPage()
	case PageKKPCredentials:
		return m.renderKKPCredentialsPage()
	case PageProviders:
		return m.renderProvidersPage()
	case PageDistributions:
		return m.renderDistributionsPage()
	case PageReleases:
		return m.renderReleasesPage()
	case PageNamePrefix:
		return m.renderNamePrefixPage()
	case PageClusterSettings:
		return m.renderClusterSettingsPage()
	case PageExclusions:
		return m.renderExclusionsPage()
	case PageConfirmation:
		return m.renderConfirmationPage()
	default:
		return "Unknown page"
	}
}

// Render methods for each page

func (m *FormModel) renderSplashPage() string {
	return `
╔════════════════════════════════════════════════════════════════╗
║                                                                ║
║                    CONFORMANCE TESTER                          ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝

Press Enter to continue...
(Press Ctrl+C to quit)
`
}

func (m *FormModel) renderEnvironmentPage() string {
	var output strings.Builder
	output.WriteString("\n")
	output.WriteString("Environment Selection\n")
	output.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	output.WriteString(fmt.Sprintf("Page %d of %d\n\n", 2, len(m.PageSequence)))

	for i, opt := range m.pageEnvOptions {
		cursor := " "
		if i == m.CursorIdx {
			cursor = "›"
		}
		output.WriteString(fmt.Sprintf("%s %s\n", cursor, opt))
	}

	output.WriteString("\n(Use ↑↓ to select, Enter to confirm)\n")
	return output.String()
}

func (m *FormModel) renderKKPCredentialsPage() string {
	fields := []string{"Seed", "Preset", "Project"}
	values := []string{m.FormData.Config.Seed, m.FormData.Config.Preset, m.FormData.Config.Project}

	var output strings.Builder
	output.WriteString("\n")
	output.WriteString("KKP Instance Credentials\n")
	output.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	output.WriteString(fmt.Sprintf("Page %d of %d\n\n", m.getCurrentPageNum(), len(m.PageSequence)))

	for i, field := range fields {
		marker := " "
		if i == m.CursorIdx {
			marker = "›"
		}
		output.WriteString(fmt.Sprintf("%s %s: %s\n", marker, field, values[i]))
	}

	output.WriteString("\n(Tab to next field, Enter to confirm, Esc to go back)\n")
	return output.String()
}

func (m *FormModel) renderProvidersPage() string {
	var output strings.Builder
	output.WriteString("\n")
	output.WriteString("Providers Selection\n")
	output.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	output.WriteString(fmt.Sprintf("Page %d of %d\n\n", m.getCurrentPageNum(), len(m.PageSequence)))

	for i, p := range m.pageProvidersList {
		cursor := " "
		if i == m.CursorIdx {
			cursor = "›"
		}

		checked := " "
		if config.Contains(m.FormData.ProvidersSelected, string(p)) {
			checked = "✓"
		}

		output.WriteString(fmt.Sprintf("  %s [%s] %s\n", cursor, checked, providerDisplayMap[p]))
	}

	output.WriteString("\n(Space to select/deselect, Enter to confirm, Esc to go back)\n")
	return output.String()
}

func (m *FormModel) renderDistributionsPage() string {
	distNames := []string{"Ubuntu", "Flatcar", "RHEL", "RockyLinux"}

	var output strings.Builder
	output.WriteString("\n")
	output.WriteString("Distributions Selection\n")
	output.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	output.WriteString(fmt.Sprintf("Page %d of %d\n\n", m.getCurrentPageNum(), len(m.PageSequence)))

	for i, name := range distNames {
		cursor := " "
		if i == m.CursorIdx {
			cursor = "›"
		}

		checked := " "
		if config.Contains(m.FormData.Dists, m.pageDistOptions[i]) {
			checked = "✓"
		}

		output.WriteString(fmt.Sprintf("  %s [%s] %s\n", cursor, checked, name))
	}

	output.WriteString("\n(Space to select/deselect, Enter to confirm, Esc to go back)\n")
	return output.String()
}

func (m *FormModel) renderReleasesPage() string {
	releases := defaulting.DefaultKubernetesVersioning.Versions

	var output strings.Builder
	output.WriteString("\n")
	output.WriteString("Releases Selection\n")
	output.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	output.WriteString(fmt.Sprintf("Page %d of %d\n\n", m.getCurrentPageNum(), len(m.PageSequence)))

	for i, rel := range releases {
		cursor := " "
		if i == m.CursorIdx {
			cursor = "›"
		}

		checked := " "
		if config.Contains(m.FormData.Releases, rel.String()) {
			checked = "✓"
		}

		output.WriteString(fmt.Sprintf("  %s [%s] %s\n", cursor, checked, rel.String()))
	}

	output.WriteString("\n(Space to select/deselect, Enter to confirm, Esc to go back)\n")
	return output.String()
}

func (m *FormModel) renderNamePrefixPage() string {
	return fmt.Sprintf(`
Name Prefix
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Page %d of %d

Name Prefix: %s_

(Enter to confirm, Esc to go back)
`, m.getCurrentPageNum(), len(m.PageSequence), m.FormData.Config.NamePrefix)
}

func (m *FormModel) renderClusterSettingsPage() string {
	nodeCountIndicator := " "
	deleteIndicator := " "

	if m.CursorIdx == 0 {
		nodeCountIndicator = "›"
	} else {
		deleteIndicator = "›"
	}

	deleteStr := "No"
	if m.FormData.Config.DeleteClusterAfterTests {
		deleteStr = "Yes"
	}

	return fmt.Sprintf(`
Cluster Settings
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Page %d of %d

%s Node Count: %s_
%s Delete After Tests: %s

(Tab to next field, Enter to confirm, Esc to go back)
`, m.getCurrentPageNum(), len(m.PageSequence), nodeCountIndicator, m.FormData.NodeCountStr, deleteIndicator, deleteStr)
}

func (m *FormModel) renderExclusionsPage() string {
	exclusionNames := []string{
		"Conformance",
		"Storage",
		"Load Balancer",
		"Usercluster Controller (RBAC)",
		"Usercluster Metrics",
		"Pod & Node Metrics",
		"Seccomp Profiles",
		"No K8s GCR Images",
		"Control Plane Security Context",
		"Telemetry",
		"Images (general)",
	}

	var output strings.Builder
	output.WriteString("\n")
	output.WriteString("Exclude Tests\n")
	output.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	output.WriteString(fmt.Sprintf("Page %d of %d\n\n", m.getCurrentPageNum(), len(m.PageSequence)))

	for i, name := range exclusionNames {
		cursor := " "
		if i == m.CursorIdx {
			cursor = "›"
		}

		checked := " "
		if config.Contains(m.FormData.Excludes, m.pageExcludeOptions[i]) {
			checked = "✓"
		}

		output.WriteString(fmt.Sprintf("  %s [%s] %s\n", cursor, checked, name))
	}

	output.WriteString("\n(Space to select/deselect, Enter to confirm, Esc to go back)\n")
	return output.String()
}

func (m *FormModel) renderConfirmationPage() string {
	confirmStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	denyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	runTestsYes := " Yes "
	runTestsNo := " No  "

	if m.FormData.RunTests {
		runTestsYes = confirmStyle.Render("> Yes <")
	} else {
		runTestsNo = denyStyle.Render("> No  <")
	}

	return fmt.Sprintf(`
Review Configuration
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Page %d of %d

Environment:       %s
Providers:         %d selected
Distributions:     %d selected
Releases:          %d selected
Name Prefix:       %s
Node Count:        %s
Delete Cluster:    %v
Exclude Tests:     %d selected

Run tests after configuration? %s %s

(Enter to finish)
`, m.getCurrentPageNum(), len(m.PageSequence),
		m.FormData.EnvOpt,
		len(m.FormData.ProvidersSelected),
		len(m.FormData.Dists),
		len(m.FormData.Releases),
		m.FormData.Config.NamePrefix,
		m.FormData.NodeCountStr,
		m.FormData.Config.DeleteClusterAfterTests,
		len(m.FormData.Excludes),
		runTestsYes, runTestsNo)
}

// getCurrentPageNum returns the current page number in the sequence
func (m *FormModel) getCurrentPageNum() int {
	for i, p := range m.PageSequence {
		if p == m.CurrentPage {
			return i + 1
		}
	}
	return 1
}
