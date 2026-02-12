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
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"k8c.io/kubermatic/v2/cmd/conformance-tester-cli/deploy"
	"k8s.io/client-go/kubernetes"
)

// executeConformanceTests runs the conformance tests for all selected providers using Kubernetes Jobs.
func (m *Model) executeConformanceTests() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Get Kubernetes client
		kubeconfigPath := m.getSelectedKubeconfigPath()
		if kubeconfigPath == "" {
			return execOutputMsg{
				output: "Error: No kubeconfig selected\n",
				err:    fmt.Errorf("no kubeconfig selected"),
			}
		}

		clientset, err := deploy.GetKubernetesClient(kubeconfigPath)
		if err != nil {
			return execOutputMsg{
				output: fmt.Sprintf("Error creating Kubernetes client: %v\n", err),
				err:    err,
			}
		}

		// Ensure namespace exists
		if err := deploy.EnsureNamespace(ctx, clientset); err != nil {
			return execOutputMsg{
				output: fmt.Sprintf("Error creating namespace: %v\n", err),
				err:    err,
			}
		}

		// Send success message for namespace
		if m.program != nil {
			m.program.Send(execOutputMsg{
				output: fmt.Sprintf("✓ Namespace %s ready\n", deploy.ConformanceNamespace),
			})
		}

		// Ensure PVC exists
		if err := deploy.EnsurePVC(ctx, clientset); err != nil {
			return execOutputMsg{
				output: fmt.Sprintf("Error creating PVC: %v\n", err),
				err:    err,
			}
		}

		// Send success message for PVC
		if m.program != nil {
			m.program.Send(execOutputMsg{
				output: fmt.Sprintf("✓ PVC %s ready\n", deploy.PVCName),
			})
		}

		// Get kubeconfig content for provider credentials
		kubeconfigContent, err := os.ReadFile(kubeconfigPath)
		if err != nil {
			return execOutputMsg{
				output: fmt.Sprintf("Error reading kubeconfig: %v\n", err),
				err:    err,
			}
		}

		// Get reports and logs directories from cluster configuration
		reportsRoot := "_reports"
		logDirectory := "_logs"
		for _, category := range m.clusterConfiguration.Categories {
			if category.Name == "Output Configuration" {
				for _, setting := range category.Settings {
					if setting.Name == "Reports Directory" {
						reportsRoot = setting.Value.(string)
					} else if setting.Name == "Log Directory" {
						logDirectory = setting.Value.(string)
					}
				}
			}
		}

		// Create jobs for all selected providers
		var jobConfigs []deploy.JobConfig
		for _, provider := range m.providers {
			if !provider.Selected {
				continue
			}

			// Generate config YAML for this provider
			configYAML := m.generateCompleteYAMLForProvider(provider)

			// Generate unique names
			timestamp := time.Now().Unix()
			providerLabel := strings.ToLower(provider.DisplayName)
			providerLabel = strings.ReplaceAll(providerLabel, " ", "-")

			// For KubeVirt, process the kubeconfig if it's from a preset
			providerKubeconfigContent := string(kubeconfigContent)
			if provider.DisplayName == "KubeVirt" && provider.HasPresetCredentials {
				if creds, ok := provider.PresetCredentials.(KubeVirtCredentials); ok {
					kubeconfigValue := creds.Kubeconfig.Value()
					if kubeconfigValue != "" && kubeconfigValue != "***preset-value***" {
						// This is base64-encoded kubeconfig from preset
						decodedKubeconfig, err := m.processKubeVirtKubeconfig(kubeconfigValue)
						if err == nil {
							// Read the temp file content
							if tempContent, readErr := os.ReadFile(decodedKubeconfig); readErr == nil {
								providerKubeconfigContent = string(tempContent)
							}
						}
					}
				}
			}

			jobConfig := deploy.JobConfig{
				ProviderName:      provider.DisplayName,
				ProviderLabel:     providerLabel,
				ConfigYAML:        configYAML,
				KubeconfigContent: providerKubeconfigContent,
				JobName:           fmt.Sprintf("conformance-%s-%d", providerLabel, timestamp),
				ConfigMapName:     fmt.Sprintf("conformance-%s-config-%d", providerLabel, timestamp),
				SecretName:        fmt.Sprintf("conformance-%s-secret-%d", providerLabel, timestamp),
				Namespace:         deploy.ConformanceNamespace,
				ReportsRoot:       reportsRoot,
				LogDirectory:      logDirectory,
			}

			jobConfigs = append(jobConfigs, jobConfig)
		}

		if len(jobConfigs) == 0 {
			return execOutputMsg{
				output: "No providers selected for testing\n",
				err:    fmt.Errorf("no providers selected"),
			}
		}

		// Create all resources and start jobs in parallel
		var wg sync.WaitGroup
		errorChan := make(chan error, len(jobConfigs))

		// Track jobs for cleanup
		m.runningJobs = []string{}
		m.jobConfigMaps = make(map[string]string)
		m.jobSecrets = make(map[string]string)

		for _, config := range jobConfigs {
			// Track job resources
			m.runningJobs = append(m.runningJobs, config.JobName)
			m.jobConfigMaps[config.JobName] = config.ConfigMapName
			m.jobSecrets[config.JobName] = config.SecretName

			wg.Add(1)
			go func(cfg deploy.JobConfig) {
				defer wg.Done()

				// Create ConfigMap
				if err := deploy.CreateConfigMap(ctx, clientset, cfg); err != nil {
					errorChan <- fmt.Errorf("failed to create ConfigMap for %s: %w", cfg.ProviderName, err)
					if m.program != nil {
						m.program.Send(execOutputMsg{
							output: fmt.Sprintf("✗ Error creating ConfigMap for %s: %v\n", cfg.ProviderName, err),
							err:    err,
						})
					}
					return
				}

				// Create Secret
				if err := deploy.CreateSecret(ctx, clientset, cfg); err != nil {
					errorChan <- fmt.Errorf("failed to create Secret for %s: %w", cfg.ProviderName, err)
					if m.program != nil {
						m.program.Send(execOutputMsg{
							output: fmt.Sprintf("✗ Error creating Secret for %s: %v\n", cfg.ProviderName, err),
							err:    err,
						})
					}
					return
				}

				// Create Job
				if err := deploy.CreateJob(ctx, clientset, cfg); err != nil {
					errorChan <- fmt.Errorf("failed to create Job for %s: %w", cfg.ProviderName, err)
					if m.program != nil {
						m.program.Send(execOutputMsg{
							output: fmt.Sprintf("✗ Error creating Job for %s: %v\n", cfg.ProviderName, err),
							err:    err,
						})
					}
					return
				}

				// Send success message
				if m.program != nil {
					m.program.Send(execOutputMsg{
						output: fmt.Sprintf("✓ Job created for %s: %s\n", cfg.ProviderName, cfg.JobName),
					})
				}
			}(config)
		}

		// Wait for all jobs to be created
		wg.Wait()
		close(errorChan)

		// Check for errors during job creation
		if len(errorChan) > 0 {
			var errMsgs []string
			for err := range errorChan {
				errMsgs = append(errMsgs, err.Error())
			}
			return execOutputMsg{
				output: fmt.Sprintf("Errors during job creation:\n%s\n", strings.Join(errMsgs, "\n")),
				err:    fmt.Errorf("job creation errors"),
			}
		}

		// Start log streaming for all jobs in parallel
		for _, config := range jobConfigs {
			go m.streamJobLogsAndWaitForCompletion(ctx, clientset, config)
		}

		return execOutputMsg{
			output: fmt.Sprintf("\n%d job(s) started. Streaming logs...\n\n", len(jobConfigs)),
		}
	}
}

// streamJobLogsAndWaitForCompletion handles log streaming and completion for a single job.
func (m *Model) streamJobLogsAndWaitForCompletion(ctx context.Context, clientset *kubernetes.Clientset, config deploy.JobConfig) {
	// Wait for pod to start
	if m.program != nil {
		m.program.Send(execOutputMsg{
			output: fmt.Sprintf("[%s] Waiting for pod to start...\n", config.ProviderName),
		})
	}

	pod, err := deploy.WaitForPodRunning(ctx, clientset, config.JobName)
	if err != nil {
		if m.program != nil {
			m.program.Send(execOutputMsg{
				output: fmt.Sprintf("[%s] Error waiting for pod: %v\n", config.ProviderName, err),
				err:    err,
			})
		}
		return
	}

	if m.program != nil {
		m.program.Send(execOutputMsg{
			output: fmt.Sprintf("[%s] ✓ Pod started: %s\n", config.ProviderName, pod.Name),
		})
	}

	// Stream logs
	outputChan := make(chan string, 100)
	go func() {
		for msg := range outputChan {
			if m.program != nil {
				m.program.Send(execOutputMsg{
					output: fmt.Sprintf("[%s] %s", config.ProviderName, msg),
				})
			}
		}
	}()

	if err := deploy.StreamPodLogs(ctx, clientset, pod.Name, outputChan); err != nil {
		if m.program != nil {
			m.program.Send(execOutputMsg{
				output: fmt.Sprintf("[%s] Error streaming logs: %v\n", config.ProviderName, err),
			})
		}
	}
	close(outputChan)

	// Wait for job completion
	if err := deploy.WaitForJobCompletion(ctx, clientset, config.JobName); err != nil {
		if m.program != nil {
			m.program.Send(execOutputMsg{
				output: fmt.Sprintf("[%s] ✗ Job failed: %v\n", config.ProviderName, err),
				err:    err,
			})
		}
	} else {
		if m.program != nil {
			m.program.Send(execOutputMsg{
				output: fmt.Sprintf("[%s] ✓ Job completed successfully\n", config.ProviderName),
			})
		}
	}
}
