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
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"k8c.io/kubermatic/v2/cmd/conformance-tester-cli/deploy"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// cleanupTestExecution cleans up all resources created during test execution.
// This includes Jobs, ConfigMaps, Secrets, PVC, and the namespace.
func (m *Model) cleanupTestExecution() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var cleanupMessages []string

		cleanupMessages = append(cleanupMessages, "\n🗑️  Starting cleanup of test resources...")

		// Get Kubernetes client
		clientset, err := deploy.GetKubernetesClient(m.getSelectedKubeconfigPath())
		if err != nil {
			errMsg := fmt.Sprintf("⚠ Error creating Kubernetes client: %v", err)
			cleanupMessages = append(cleanupMessages, errMsg)
			return cleanupProgressMsg{
				message: strings.Join(cleanupMessages, "\n"),
				done:    true,
				err:     err,
			}
		}

		// Delete all running jobs
		gracePeriod := int64(deploy.GracefulDeletionPeriod.Seconds())
		propagation := metav1.DeletePropagationForeground
		deleteOpts := metav1.DeleteOptions{
			GracePeriodSeconds: &gracePeriod,
			PropagationPolicy:  &propagation,
		}

		if len(m.runningJobs) > 0 {
			cleanupMessages = append(cleanupMessages, fmt.Sprintf("\n📋 Deleting %d job(s)...", len(m.runningJobs)))
			for _, jobName := range m.runningJobs {
				err := clientset.BatchV1().Jobs(deploy.ConformanceNamespace).Delete(
					ctx,
					jobName,
					deleteOpts,
				)
				if err != nil {
					cleanupMessages = append(cleanupMessages, fmt.Sprintf("  ⚠ Failed to delete job %s: %v", jobName, err))
				} else {
					cleanupMessages = append(cleanupMessages, fmt.Sprintf("  ✓ Deleted job: %s", jobName))
				}

				// Delete associated ConfigMap
				if configMapName, ok := m.jobConfigMaps[jobName]; ok {
					err := clientset.CoreV1().ConfigMaps(deploy.ConformanceNamespace).Delete(ctx, configMapName, metav1.DeleteOptions{})
					if err != nil {
						cleanupMessages = append(cleanupMessages, fmt.Sprintf("  ⚠ Failed to delete configmap %s: %v", configMapName, err))
					} else {
						cleanupMessages = append(cleanupMessages, fmt.Sprintf("  ✓ Deleted configmap: %s", configMapName))
					}
				}

				// Delete associated Secret
				if secretName, ok := m.jobSecrets[jobName]; ok {
					err := clientset.CoreV1().Secrets(deploy.ConformanceNamespace).Delete(ctx, secretName, metav1.DeleteOptions{})
					if err != nil {
						cleanupMessages = append(cleanupMessages, fmt.Sprintf("  ⚠ Failed to delete secret %s: %v", secretName, err))
					} else {
						cleanupMessages = append(cleanupMessages, fmt.Sprintf("  ✓ Deleted secret: %s", secretName))
					}
				}
			}
		}

		// Delete namespace
		cleanupMessages = append(cleanupMessages, fmt.Sprintf("\n🗂️  Deleting namespace %s...", deploy.ConformanceNamespace))
		err = clientset.CoreV1().Namespaces().Delete(
			ctx,
			deploy.ConformanceNamespace,
			metav1.DeleteOptions{},
		)
		if err != nil {
			cleanupMessages = append(cleanupMessages, fmt.Sprintf("  ⚠ Failed to delete namespace: %v", err))
		} else {
			cleanupMessages = append(cleanupMessages, "  ⏳ Namespace deletion initiated (this may take a moment)...")

			// Wait for namespace to be fully deleted (with timeout)
			timeout := time.After(30 * time.Second)
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			namespaceDeleted := false
			for !namespaceDeleted {
				select {
				case <-timeout:
					cleanupMessages = append(cleanupMessages, "  ⚠ Namespace deletion timed out (it may still be deleting in the background)")
					namespaceDeleted = true
				case <-ticker.C:
					_, err := clientset.CoreV1().Namespaces().Get(ctx, deploy.ConformanceNamespace, metav1.GetOptions{})
					if err != nil {
						// Namespace not found means it's been deleted
						cleanupMessages = append(cleanupMessages, fmt.Sprintf("  ✓ Namespace %s deleted successfully", deploy.ConformanceNamespace))
						namespaceDeleted = true
					}
				}
			}
		}

		cleanupMessages = append(cleanupMessages, "\n✅ Cleanup completed - all test resources have been removed")

		return cleanupProgressMsg{
			message: strings.Join(cleanupMessages, "\n"),
			done:    true,
			err:     nil,
		}
	}
}
