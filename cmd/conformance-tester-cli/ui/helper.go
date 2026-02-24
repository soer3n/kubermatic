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
	"os"
	"strings"
)

// Constants for error messages.
const (
// Define error message constants here if needed
)

// updateEnvironmentFocus updates the focus state of environment input fields
func (m *Model) updateEnvironmentFocus() {
	// Blur all local environment fields
	m.localEnv.KubermaticConfigurationsPath.Blur()
	m.localEnv.HelmValuesPath.Blur()
	m.localEnv.MLAValuesPath.Blur()

	// Blur all existing environment text input fields
	m.existingEnv.CustomKubeconfigPath.Blur()
	m.existingEnv.ProjectName.Blur()

	// Focus the current field based on state
	if m.environmentFocusIndex == 0 && m.localEnv.Selected && m.environmentFieldIndex > 0 {
		switch m.environmentFieldIndex {
		case 1:
			m.localEnv.KubermaticConfigurationsPath.Focus()
		case 2:
			m.localEnv.HelmValuesPath.Focus()
		case 3:
			m.localEnv.MLAValuesPath.Focus()
		}
	} else if m.environmentFocusIndex == 1 && m.existingEnv.Selected && m.environmentFieldIndex > 0 {
		switch m.environmentFieldIndex {
		case 1:
			// Focus custom kubeconfig path if custom option is selected
			// Convert visual index to actual option index
			optionIndex := m.getKubeconfigOptionIndexFromVisualIndex(m.existingEnv.KubeconfigFocusedIndex)
			if optionIndex >= 0 && optionIndex < len(m.existingEnv.KubeconfigOptions) {
				selectedOption := m.existingEnv.KubeconfigOptions[optionIndex]
				if selectedOption.Type == "custom" && selectedOption.Selected {
					m.existingEnv.CustomKubeconfigPath.Focus()
				}
			}
		// case 2 and 3 are now Seeds and Presets selection lists (not text inputs)
		case 4:
			m.existingEnv.ProjectName.Focus()
		}
	}
}

// validateLocalEnvironment validates the local environment input fields.
func (m *Model) validateLocalEnvironment() bool {
	// Clear previous errors
	m.localEnv.Errors = EnvironmentLocalErrors{}

	// Validate Kubermatic Configurations Path
	if strings.TrimSpace(m.localEnv.KubermaticConfigurationsPath.Value()) == "" {
		m.localEnv.Errors.KubermaticConfigurationsPath = "Kubermatic configurations path is required"
		return false
	}

	// Validate Kubermatic Configurations Path existence
	if _, err := os.Stat(strings.TrimSpace(m.localEnv.KubermaticConfigurationsPath.Value())); os.IsNotExist(err) {
		m.localEnv.Errors.KubermaticConfigurationsPath = "Kubermatic configurations file does not exist"
		return false
	}

	// Validate Helm Values Path
	if strings.TrimSpace(m.localEnv.HelmValuesPath.Value()) == "" {
		m.localEnv.Errors.HelmValuesPath = "Helm values path is required"
		return false
	}

	// Validate Helm Values Path existence
	if _, err := os.Stat(strings.TrimSpace(m.localEnv.HelmValuesPath.Value())); os.IsNotExist(err) {
		m.localEnv.Errors.HelmValuesPath = "Helm values file does not exist"
		return false
	}

	// Validate MLA Values Path existence
	if strings.TrimSpace(m.localEnv.MLAValuesPath.Value()) != "" {
		if _, err := os.Stat(strings.TrimSpace(m.localEnv.MLAValuesPath.Value())); os.IsNotExist(err) {
			m.localEnv.Errors.MLAValuesPath = "MLA values file does not exist"
			return false
		}
	}

	return true
}

func (m *Model) validateExistingEnvironment() bool {
	// Clear previous errors
	m.existingEnv.Errors = EnvironmentExistingErrors{Fields: make(map[string]string)}

	// Validate kubeconfig
	kubeconfigPath := m.getSelectedKubeconfigPath()
	if kubeconfigPath == "" {
		m.existingEnv.Errors.KubeconfigPath = "Kubeconfig path is required"
		return false
	} else if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		m.existingEnv.Errors.KubeconfigPath = "Kubeconfig file does not exist"
		return false
	}

	// Validate Seed selection
	if m.existingEnv.SelectedSeedIndex < 0 || m.existingEnv.SelectedSeedIndex >= len(m.existingEnv.AvailableSeeds) {
		m.existingEnv.Errors.Fields["SeedName"] = "Please select a Seed"
		return false
	}

	// // Validate Preset selection
	// if m.existingEnv.SelectedPresetIndex < 0 || m.existingEnv.SelectedPresetIndex >= len(m.existingEnv.AvailablePresets) && len(m.existingEnv.AvailablePresets) > 0 {
	// 	m.existingEnv.Errors.Fields["PresetName"] = "Please select a Preset"
	// 	return false
	// }

	return true
}
