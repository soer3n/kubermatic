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

package deploy

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// ConformanceNamespace is the dedicated namespace for conformance tests.
	ConformanceNamespace = "conformance-tests"
	// ConformanceImage is the Docker image used for running tests.
	ConformanceImage = "docker.io/soer3n/conformance:ginkgo-09510005032026"

	GracefulDeletionPeriod = 15 * time.Minute
)

// kubeconfigFileRegex matches the kubeconfigFile key in YAML configuration,
// used to replace the user-provided path with the container mount path.
var kubeconfigFileRegex = regexp.MustCompile(`kubeconfigFile:\s*["']?[^"'\n]+["']?`)

// JobConfig holds configuration for creating a Kubernetes Job.
type JobConfig struct {
	ProviderName      string
	ProviderLabel     string
	ConfigYAML        string
	KubeconfigContent string
	JobName           string
	ConfigMapName     string
	SecretName        string
	Namespace         string
	ReportsRoot       string
	LogDirectory      string
}

// GetKubernetesClient creates a Kubernetes clientset from the given kubeconfig path.
func GetKubernetesClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return clientset, nil
}

// EnsureNamespace creates the conformance-tests namespace if it doesn't exist.
func EnsureNamespace(ctx context.Context, clientset *kubernetes.Clientset) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ConformanceNamespace,
		},
	}

	_, err := clientset.CoreV1().Namespaces().Get(ctx, ConformanceNamespace, metav1.GetOptions{})
	if err == nil {
		// Namespace already exists
		return nil
	}

	_, err = clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create namespace %s: %w", ConformanceNamespace, err)
	}

	return nil
}

// CreateConfigMap creates a ConfigMap with the provider configuration.
func CreateConfigMap(ctx context.Context, clientset *kubernetes.Clientset, config JobConfig) error {
	// Add "client: kube" to the config if not present
	configYAML := config.ConfigYAML
	if !strings.Contains(configYAML, "client:") {
		configYAML = "client: \"kube\"\n\n" + configYAML
	}

	// Replace any kubeconfigFile path with the mounted secret path.
	configYAML = kubeconfigFileRegex.ReplaceAllString(configYAML, `kubeconfigFile: "/opt/kubeconfig"`)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.ConfigMapName,
			Namespace: config.Namespace,
			Labels: map[string]string{
				"app":      "conformance-tester",
				"provider": strings.ToLower(config.ProviderLabel),
			},
		},
		Data: map[string]string{
			"config.yaml": configYAML,
		},
	}

	_, err := clientset.CoreV1().ConfigMaps(config.Namespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create ConfigMap: %w", err)
	}

	return nil
}

// EnsureClusterRoleBinding creates or updates a ClusterRoleBinding that grants cluster-admin
// permissions to the default service account in the conformance-tests namespace.
func EnsureClusterRoleBinding(ctx context.Context, clientset *kubernetes.Clientset) error {
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "conformance-cluster-admin",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup:  "",
				Kind:      "ServiceAccount",
				Name:      "default",
				Namespace: ConformanceNamespace,
			},
		},
	}

	existing, err := clientset.RbacV1().ClusterRoleBindings().Get(ctx, "conformance-cluster-admin", metav1.GetOptions{})
	if err == nil {
		// Update the existing binding to ensure it has the correct subjects.
		crb.ResourceVersion = existing.ResourceVersion
		_, err = clientset.RbacV1().ClusterRoleBindings().Update(ctx, crb, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update ClusterRoleBinding: %w", err)
		}
		return nil
	}

	_, err = clientset.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create ClusterRoleBinding: %w", err)
	}

	return nil
}

// CreateSecret creates a Secret with the kubeconfig for provider credentials.
func CreateSecret(ctx context.Context, clientset *kubernetes.Clientset, config JobConfig) error {
	// The kubeconfig content should already be raw content (not base64 encoded for the Secret data field)
	// Kubernetes will handle the base64 encoding when storing in etcd
	kubeconfigData := []byte(config.KubeconfigContent)

	// If it's already base64 encoded, decode it first
	if decoded, err := base64.StdEncoding.DecodeString(config.KubeconfigContent); err == nil {
		kubeconfigData = decoded
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.SecretName,
			Namespace: config.Namespace,
			Labels: map[string]string{
				"app":      "conformance-tester",
				"provider": strings.ToLower(config.ProviderLabel),
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"kubeconfig": kubeconfigData,
		},
	}

	_, err := clientset.CoreV1().Secrets(config.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create Secret: %w", err)
	}

	return nil
}

// CreateJob creates a Kubernetes Job for running conformance tests.
func CreateJob(ctx context.Context, clientset *kubernetes.Clientset, config JobConfig) error {
	backoffLimit := int32(0)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.JobName,
			Namespace: config.Namespace,
			Labels: map[string]string{
				"app":      "conformance-tester",
				"provider": strings.ToLower(config.ProviderLabel),
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":      "conformance-tester",
						"provider": strings.ToLower(config.ProviderLabel),
						"job-name": config.JobName,
					},
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: func() *int64 { v := int64(GracefulDeletionPeriod.Seconds()); return &v }(),
					RestartPolicy:                 corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:            "ginkgo",
							Image:           ConformanceImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{
									Name:  "CONFORMANCE_TESTER_CONFIG_FILE",
									Value: "/opt/config.yaml",
								},
							},
							Args: []string{
								"--ginkgo.v",
								"--ginkgo.parallel.process=1",
								"--ginkgo.parallel.total=1",
								"--ginkgo.grace-period=15m",
								"--ginkgo.focus",
								config.ProviderName,
								fmt.Sprintf("--ginkgo.label-filter=%s", config.ProviderLabel),
								"./...",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config",
									MountPath: "/opt/config.yaml",
									SubPath:   "config.yaml",
									ReadOnly:  true,
								},
								{
									Name:      "kubeconfig",
									MountPath: "/opt/kubeconfig",
									SubPath:   "kubeconfig",
									ReadOnly:  true,
								},
								{
									Name:      "results",
									MountPath: "/" + config.ReportsRoot,
								},
								{
									Name:      "results",
									MountPath: "/" + config.LogDirectory,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: config.ConfigMapName,
									},
								},
							},
						},
						{
							Name: "kubeconfig",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: config.SecretName,
								},
							},
						},
						{
							Name: "results",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	_, err := clientset.BatchV1().Jobs(config.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create Job: %w", err)
	}

	return nil
}

// WaitForPodRunning waits for the Job's pod to start running.
func WaitForPodRunning(ctx context.Context, clientset *kubernetes.Clientset, jobName string) (*corev1.Pod, error) {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for pod to start")
		case <-ticker.C:
			pods, err := clientset.CoreV1().Pods(ConformanceNamespace).List(ctx, metav1.ListOptions{
				LabelSelector: fmt.Sprintf("job-name=%s", jobName),
			})
			if err != nil {
				continue
			}

			if len(pods.Items) == 0 {
				continue
			}

			pod := &pods.Items[0]
			if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
				return pod, nil
			}
		}
	}
}

// StreamPodLogs streams logs from a pod to the output channel, line by line.
func StreamPodLogs(ctx context.Context, clientset *kubernetes.Clientset, podName string, outputChan chan<- string) error {
	req := clientset.CoreV1().Pods(ConformanceNamespace).GetLogs(podName, &corev1.PodLogOptions{
		Follow: true,
	})

	stream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to open log stream: %w", err)
	}
	defer stream.Close()

	scanner := bufio.NewScanner(stream)
	// Increase buffer to handle long JSON log lines (default is 64KB).
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)
	for scanner.Scan() {
		outputChan <- scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return fmt.Errorf("error reading logs: %w", err)
	}

	return nil
}

// WaitForJobCompletion waits for a Job to complete and returns its status.
func WaitForJobCompletion(ctx context.Context, clientset *kubernetes.Clientset, jobName string) error {
	timeout := time.After(2 * time.Hour) // Conformance tests can take a long time
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for job to complete")
		case <-ticker.C:
			job, err := clientset.BatchV1().Jobs(ConformanceNamespace).Get(ctx, jobName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get job status: %w", err)
			}

			if job.Status.Succeeded > 0 {
				return nil
			}

			if job.Status.Failed > 0 {
				return fmt.Errorf("job failed")
			}
		}
	}
}
