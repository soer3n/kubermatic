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
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"k8c.io/kubermatic/v2/pkg/test/e2e/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Constants for error messages.
const (
	ErrCIDRRequired             = "Network CIDR is required"
	ErrInvalidCIDRFormat        = "Invalid CIDR format (e.g., 10.244.0.0/16)"
	ErrCIDRDefaultServiceSubnet = "Invalid Range: cannot use the default service subnet"
	ErrDNSServerRequired        = "DNS Server is required"
	ErrInvalidDNSServerFormat   = "Invalid IP address format"
	ErrGatewayIPRequired        = "Gateway IP is required"
	ErrInvalidGatewayIPFormat   = "Invalid IP address format"
	ErrRegistryEndpointRequired = "Registry endpoint is required"
	// Node count validation errors.
	ErrNodeCountRequired             = "Node count is required"
	ErrNodeCountOutOfRange           = "Node count must be between 1 and %d"
	ErrControlPlaneCountRequired     = "Control plane count is required"
	ErrControlPlaneCountInvalid      = "Control plane count must be at least 1"
	ErrControlPlaneCountExceedsNodes = "Control plane count cannot exceed total node count"
)

// validateIPRange validates an IP range in the format "IP1-IP2"
// e.g., "192.168.1.100-192.168.1.150".
func (m *Model) validateIPRange(input string) bool {
	if input == "" {
		return false
	}

	// Split by hyphen
	parts := strings.Split(input, "-")
	if len(parts) != 2 {
		return false
	}

	startIP := strings.TrimSpace(parts[0])
	endIP := strings.TrimSpace(parts[1])

	// Validate both IP addresses
	if !m.isValidIP(startIP) || !m.isValidIP(endIP) {
		return false
	}

	// Optional: Check that start IP is less than or equal to end IP
	return m.isIPGreaterOrEqual(startIP, endIP)
}

// isValidIP validates a single IP address (e.g., "192.168.1.100").
func (m *Model) isValidIP(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		// Check for empty parts
		if part == "" {
			return false
		}

		// Check for leading zeros (e.g., "01" is invalid)
		if len(part) > 1 && part[0] == '0' {
			return false
		}

		// Convert to integer
		num, err := strconv.Atoi(part)
		if err != nil || num < 0 || num > 255 {
			return false
		}
	}
	return true
}

// isIPGreaterOrEqual checks if ip1 <= ip2.
func (m *Model) isIPGreaterOrEqual(ip1, ip2 string) bool {
	parts1 := strings.Split(ip1, ".")
	parts2 := strings.Split(ip2, ".")

	for i := 0; i < 4; i++ {
		num1, _ := strconv.Atoi(parts1[i])
		num2, _ := strconv.Atoi(parts2[i])

		if num1 < num2 {
			return true
		} else if num1 > num2 {
			return false
		}
		// If equal, continue to next octet
	}
	// IPs are identical
	return true
}

// updateEnvironmentFocus updates the focus state of environment input fields
func (m *Model) updateEnvironmentFocus() {
	// Blur all local environment fields
	m.localEnv.KubermaticConfigurationsPath.Blur()
	m.localEnv.HelmValuesPath.Blur()
	m.localEnv.MLAValuesPath.Blur()

	// Blur all existing environment fields
	m.existingEnv.KubeconfigPath.Blur()
	m.existingEnv.SeedName.Blur()
	m.existingEnv.PresetName.Blur()
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
			m.existingEnv.KubeconfigPath.Focus()
		case 2:
			m.existingEnv.SeedName.Focus()
		case 3:
			m.existingEnv.PresetName.Focus()
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
	m.existingEnv.Errors = EnvironmentExistingErrors{}

	// // Validate Kubeconfig Path
	// if err := m.isValidKubeConfig(m.existingEnv.KubeconfigPath.Value()); err != nil {
	// 	m.existingEnv.Errors.KubeconfigPath = fmt.Sprintf("Invalid kubeconfig: %v", err)
	// 	return false
	// }

	// // Validate Seed Name
	// if strings.TrimSpace(m.existingEnv.SeedName.Value()) == "" {
	// 	m.existingEnv.Errors.SeedName = "Seed name is required"
	// 	return false
	// }

	// // Validate Preset Name
	// if strings.TrimSpace(m.existingEnv.PresetName.Value()) == "" {
	// 	m.existingEnv.Errors.PresetName = "Preset name is required"
	// 	return false
	// }

	// // Validate Project Name
	// if strings.TrimSpace(m.existingEnv.ProjectName.Value()) == "" {
	// 	m.existingEnv.Errors.ProjectName = "Project name is required"
	// 	return false
	// }
	return true
}

func (m *Model) isValidKubeConfig(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path is empty")
	}

	// Check if file exists
	if _, err := os.Stat(strings.TrimSpace(path)); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist")
	}

	// Temporarily set KUBECONFIG environment variable
	oldKubeconfig := os.Getenv("KUBECONFIG")
	os.Setenv("KUBECONFIG", strings.TrimSpace(path))
	defer os.Setenv("KUBECONFIG", oldKubeconfig)

	// Try to create a client using the kubeconfig
	client, _, err := utils.GetClients()
	if err != nil {
		return err
	}

	// Test a simple API call to verify connectivity with a 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Get(ctx, types.NamespacedName{Name: "default"}, &corev1.Namespace{})
	if err != nil {
		return fmt.Errorf("Cluster unreachable")
	}

	return nil
}

// Validate CIDR notation (e.g., 10.244.0.0/16).
func (m *Model) isValidCIDR(cidr string) bool {
	parts := strings.Split(cidr, "/")
	if len(parts) != 2 {
		return false
	}

	ip := parts[0]
	prefix := parts[1]

	// Validate IP part
	if !m.isValidIP(ip) {
		return false
	}

	// Validate prefix
	prefixNum, err := strconv.Atoi(prefix)
	if err != nil || prefixNum < 0 || prefixNum > 32 {
		return false
	}

	return true
}

// // Validate network configuration fields.
// func (m *Model) validateNetworkConfig() bool {
// 	// Clear previous errors
// 	m.Network.Errors = NetworkErrors{}

// 	// Get values from inputs
// 	network := m.Network.CIDR.Value()
// 	dnsServer := m.Network.DNSServer.Value()
// 	gatewayIP := m.Network.GatewayIP.Value()

// 	// Validate Network CIDR
// 	if network == "" {
// 		m.Network.Errors.CIDR = ErrCIDRRequired
// 		return false
// 	}
// 	if !m.isValidCIDR(network) {
// 		m.Network.Errors.CIDR = ErrInvalidCIDRFormat
// 		return false
// 	}

// 	// if network == kubeone.DefaultServiceSubnet {
// 	// 	m.Network.Errors.CIDR = ErrCIDRDefaultServiceSubnet
// 	// 	return false
// 	// }

// 	// Validate DNS Server
// 	if dnsServer == "" {
// 		m.Network.Errors.DNSServer = ErrDNSServerRequired
// 		return false
// 	}
// 	if !m.isValidIP(dnsServer) {
// 		m.Network.Errors.DNSServer = ErrInvalidDNSServerFormat
// 		return false
// 	}

// 	// Validate Gateway IP
// 	if gatewayIP == "" {
// 		m.Network.Errors.GatewayIP = ErrGatewayIPRequired
// 		return false
// 	}
// 	if !m.isValidIP(gatewayIP) {
// 		m.Network.Errors.GatewayIP = ErrInvalidGatewayIPFormat
// 		return false
// 	}

// 	return true
// }

// // validateContainerRegistry ensures the offline registry settings are usable.
// func (m *Model) validateContainerRegistry() bool {
// 	if !m.offline {
// 		return true
// 	}

// 	m.ContainerRegistry.Error = ""

// 	if strings.TrimSpace(m.ContainerRegistry.Endpoint.Value()) == "" {
// 		m.ContainerRegistry.Error = ErrRegistryEndpointRequired
// 		return false
// 	}

// 	return true
// }

// // updateHelmRegistryFocus ensures the focused input matches the selected field for Helm registry.
// func (m *Model) updateHelmRegistryFocus() {
// 	if !m.offline {
// 		return
// 	}

// 	m.HelmRegistry.Endpoint.Blur()
// 	m.HelmRegistry.Username.Blur()
// 	m.HelmRegistry.Password.Blur()

// 	switch m.HelmRegistry.CurrentField {
// 	case 0:
// 		m.HelmRegistry.Endpoint.Focus()
// 	case 1:
// 		m.HelmRegistry.Username.Focus()
// 	case 2:
// 		m.HelmRegistry.Password.Focus()
// 	}
// }

// // validateHelmRegistry ensures the Helm registry settings are valid.
// func (m *Model) validateHelmRegistry() bool {
// 	if !m.offline {
// 		return true
// 	}

// 	m.HelmRegistry.Error = ""

// 	if strings.TrimSpace(m.HelmRegistry.Endpoint.Value()) == "" {
// 		m.HelmRegistry.Error = "Helm registry endpoint is required"
// 		return false
// 	}

// 	return true
// }

// updatePackageRepoFocus ensures the package repository address input is focused when enabled.
// func (m *Model) updatePackageRepoFocus() {
// 	if m.PackageRepo.Enabled {
// 		m.PackageRepo.Address.Focus()
// 	} else {
// 		m.PackageRepo.Address.Blur()
// 	}
// }

// normalizeRegistryBase converts any registry URL to oci:// format.
func normalizeRegistryBase(input string) string {
	// Strip all possible schemes
	clean := strings.TrimPrefix(input, "http://")
	clean = strings.TrimPrefix(clean, "https://")
	clean = strings.TrimPrefix(clean, "oci://")

	// Handle possible double slashes after scheme removal
	clean = strings.TrimPrefix(clean, "//")

	// Remove leading/trailing slashes
	clean = strings.Trim(clean, "/")

	// Enforce oci:// scheme
	return "oci://" + clean
}

func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 1
	}
	return i
}

// func (m *Model) processNodeCountInput() bool {
// 	nodeCountStr := strings.TrimSpace(m.NodeCount.NodeCountInput.Value())
// 	cpCountStr := strings.TrimSpace(m.NodeCount.ControlPlaneCountInput.Value())

// 	// Validate node count
// 	if nodeCountStr == "" {
// 		m.NodeCount.Error = ErrNodeCountRequired
// 		return false
// 	}

// 	nodeCount, err := strconv.Atoi(nodeCountStr)
// 	if err != nil || nodeCount < 1 || nodeCount > m.NodeCount.Max {
// 		m.NodeCount.Error = fmt.Sprintf(ErrNodeCountOutOfRange, m.NodeCount.Max)
// 		return false
// 	}

// 	// Validate control plane count
// 	if cpCountStr == "" {
// 		m.NodeCount.Error = ErrControlPlaneCountRequired
// 		return false
// 	}

// 	cpCount, err := strconv.Atoi(cpCountStr)
// 	if err != nil || cpCount < 1 {
// 		m.NodeCount.Error = ErrControlPlaneCountInvalid
// 		return false
// 	}

// 	// Control plane count cannot exceed node count
// 	if cpCount > nodeCount {
// 		m.NodeCount.Error = ErrControlPlaneCountExceedsNodes
// 		return false
// 	}

// 	// Initialize nodes
// 	// m.initializeNodes(nodeCount)
// 	m.NodeCount.Error = ""
// 	return true
// }
