/*
                  Kubermatic Enterprise Read-Only License
                         Version 1.0 ("KERO-1.0")
                     Copyright © 2026 Kubermatic GmbH

   1.	You may only view, read and display for studying purposes the source
      code of the software licensed under this license, and, to the extent
      explicitly provided under this license, the binary code.
   2.	Any use of the software which exceeds the foregoing right, including,
      without limitation, its execution, compilation, copying, modification
      and distribution, is expressly prohibited.
   3.	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
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
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	kubevirt "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/kubevirt"
)

// fetchDatacenterSettingsForProvider fetches datacenter settings from the cluster for a specific provider.
func (m *Model) fetchDatacenterSettingsForProvider(provider string) tea.Cmd {
	return func() tea.Msg {
		// For now, we'll use environment variable to temporarily set the kubeconfig
		// This is a workaround since we can't modify the conformance-tester package
		kubeconfigPath := m.getSelectedKubeconfigPath()
		if kubeconfigPath == "" {
			return datacenterSettingsLoadedMsg{
				provider: provider,
				err:      fmt.Errorf("no kubeconfig selected"),
			}
		}

		// Expand path if needed
		if strings.HasPrefix(kubeconfigPath, "~/") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return datacenterSettingsLoadedMsg{
					provider: provider,
					err:      fmt.Errorf("unable to access home directory: %w", err),
				}
			}
			kubeconfigPath = filepath.Join(homeDir, kubeconfigPath[2:])
		}

		// Temporarily set KUBECONFIG environment variable
		// Note: This is not ideal but necessary since we can't modify the conformance-tester package
		oldKubeconfig := os.Getenv("KUBECONFIG")
		os.Setenv("KUBECONFIG", kubeconfigPath)
		defer func() {
			if oldKubeconfig != "" {
				os.Setenv("KUBECONFIG", oldKubeconfig)
			} else {
				os.Unsetenv("KUBECONFIG")
			}
		}()

		var descriptionsMap map[string]k8cginkgo.Description

		// Call provider-specific functions
		switch strings.ToLower(provider) {
		case "kubevirt":
			// Import is at package level, call the existing function
			descriptionsMap = fetchKubeVirtDatacenterSettings()
		// Add more providers as they become available
		default:
			descriptionsMap = make(map[string]k8cginkgo.Description)
		}

		if descriptionsMap == nil {
			descriptionsMap = make(map[string]k8cginkgo.Description)
		}

		return datacenterSettingsLoadedMsg{
			provider:     provider,
			descriptions: descriptionsMap,
			err:          nil,
		}
	}
}

// fetchMachineSettingsForProvider fetches machine deployment settings from the cluster for a specific provider.
func (m *Model) fetchMachineSettingsForProvider(provider string) tea.Cmd {
	return func() tea.Msg {
		// For now, we'll use environment variable to temporarily set the kubeconfig
		// This is a workaround since we can't modify the conformance-tester package
		kubeconfigPath := m.getSelectedKubeconfigPath()
		if kubeconfigPath == "" {
			return machineSettingsLoadedMsg{
				provider: provider,
				err:      fmt.Errorf("no kubeconfig selected"),
			}
		}

		// Expand path if needed
		if strings.HasPrefix(kubeconfigPath, "~/") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return machineSettingsLoadedMsg{
					provider: provider,
					err:      fmt.Errorf("unable to access home directory: %w", err),
				}
			}
			kubeconfigPath = filepath.Join(homeDir, kubeconfigPath[2:])
		}

		// Temporarily set KUBECONFIG environment variable
		oldKubeconfig := os.Getenv("KUBECONFIG")
		os.Setenv("KUBECONFIG", kubeconfigPath)
		defer func() {
			if oldKubeconfig != "" {
				os.Setenv("KUBECONFIG", oldKubeconfig)
			} else {
				os.Unsetenv("KUBECONFIG")
			}
		}()

		var descriptionsMap map[string]k8cginkgo.Description

		// Call provider-specific functions
		switch strings.ToLower(provider) {
		case "kubevirt":
			descriptionsMap = fetchKubeVirtMachineSettings()
		// Add more providers as they become available
		default:
			descriptionsMap = make(map[string]k8cginkgo.Description)
		}

		if descriptionsMap == nil {
			descriptionsMap = make(map[string]k8cginkgo.Description)
		}

		return machineSettingsLoadedMsg{
			provider:     provider,
			descriptions: descriptionsMap,
			err:          nil,
		}
	}
}

// fetchKubeVirtDatacenterSettings fetches KubeVirt datacenter settings.
// This replicates the logic from kubevirt.GetDatacenterDescriptions() but uses the environment's kubeconfig.
func fetchKubeVirtDatacenterSettings() map[string]k8cginkgo.Description {
	// Call the existing function - it will now use the KUBECONFIG we set in the environment
	// We recover from any panics that might occur
	defer func() {
		if r := recover(); r != nil {
			// Silently recover - return will be nil which is handled by caller
		}
	}()

	// Call the kubevirt package function directly
	// This will use the KUBECONFIG environment variable we temporarily set
	return kubevirt.GetDatacenterDescriptions()
}

// fetchKubeVirtMachineSettings fetches KubeVirt machine deployment settings.
// This replicates the logic from kubevirt.GetMachineDescriptions() but uses the environment's kubeconfig.
func fetchKubeVirtMachineSettings() map[string]k8cginkgo.Description {
	// Call the existing function - it will now use the KUBECONFIG we set in the environment
	// We recover from any panics that might occur
	defer func() {
		if r := recover(); r != nil {
			// Silently recover - return will be nil which is handled by caller
		}
	}()

	// Call the kubevirt package function directly
	// This will use the KUBECONFIG environment variable we temporarily set
	return kubevirt.GetMachineDescriptions()
}
