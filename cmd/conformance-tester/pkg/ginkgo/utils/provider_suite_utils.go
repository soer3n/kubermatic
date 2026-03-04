package utils

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/aws/smithy-go/ptr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/clients"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/scenarios"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/util"
	controllerutil "k8c.io/kubermatic/v2/pkg/controller/util"
	"k8c.io/kubermatic/v2/pkg/version/kubermatic"
	"k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/retry"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

var MachineNameLabel = "cluster.k8s.io/machine-set-name"

func KKP(msg string) string {
	return fmt.Sprintf("[KKP] %s", msg)
}

func CloudProvider(msg string) string {
	return fmt.Sprintf("[CloudProvider] %s", msg)
}

func CommonSetup(rootCtx context.Context, log *zap.SugaredLogger, scenario scenarios.Scenario, legacyOpts *legacytypes.Options) (*kubermaticv1.Cluster, ctrlruntimeclient.Client) {
	var userClusterClient ctrlruntimeclient.Client
	var cluster *kubermaticv1.Cluster
	var err error
	// By("Creating a new cluster")
	// legacyOpts := toLegacyOptions(opts, runtimeOpts)
	By(KKP("Create Cluster"), func() {
		cluster, err = clients.NewKubeClient(legacyOpts).CreateCluster(rootCtx, log, scenario)
		Expect(err).NotTo(HaveOccurred())
	})

	By(KKP("Wait for successful reconciliation"), func() {
		// NB: It's important for this health check loop to refresh the cluster object, as
		// during reconciliation some cloud providers will fill in missing fields in the CloudSpec,
		// and later when we create MachineDeployments we potentially rely on these fields
		// being set in the cluster variable.
		versions := kubermatic.GetVersions()
		log.Info("Waiting for cluster to be successfully reconciled...")

		Eventually(func() bool {
			if err := legacyOpts.SeedClusterClient.Get(rootCtx, types.NamespacedName{Name: cluster.Name}, cluster); err != nil {
				return false
			}

			// ignore Kubermatic version in this check, to allow running against a 3rd party setup
			missingConditions, _ := controllerutil.ClusterReconciliationSuccessful(cluster, versions, true)
			if len(missingConditions) > 0 {
				return false
			}

			return true
		}, 20*time.Minute, 5*time.Second).Should(BeTrue(), "cluster was not reconciled successfully within the timeout")

		Eventually(func() bool {
			newCluster := &kubermaticv1.Cluster{}
			namespacedClusterName := types.NamespacedName{Name: cluster.Name}
			if err := legacyOpts.SeedClusterClient.Get(rootCtx, namespacedClusterName, newCluster); err != nil {
				if apierrors.IsNotFound(err) {
					return false
				}
			}

			// Check for this first, because otherwise we instantly return as the cluster-controller did not
			// create any pods yet
			if !newCluster.Status.ExtendedHealth.AllHealthy() {
				return false
			}

			controlPlanePods := &corev1.PodList{}
			if err := legacyOpts.SeedClusterClient.List(
				rootCtx,
				controlPlanePods,
				&ctrlruntimeclient.ListOptions{Namespace: newCluster.Status.NamespaceName},
			); err != nil {
				return false
			}

			unready := sets.New[string]()
			for _, pod := range controlPlanePods.Items {
				if !util.PodIsReady(&pod) {
					unready.Insert(pod.Name)
				}
			}

			if unready.Len() == 0 {
				return true
			}

			return false
		}, 20*time.Minute, 5*time.Second).Should(BeTrue(), "cluster did not become healthy within the timeout")
	})

	By(KKP("Wait for control plane"), func() {
		Eventually(func() bool {
			if err := legacyOpts.SeedClusterClient.Get(rootCtx, types.NamespacedName{Name: cluster.Name}, cluster); err != nil {
				return false
			}
			versions := kubermatic.GetVersions()
			// ignore Kubermatic version in this check, to allow running against a 3rd party setup
			missingConditions, _ := controllerutil.ClusterReconciliationSuccessful(cluster, versions, true)
			return len(missingConditions) == 0
		}, 20*time.Minute, 5*time.Second).Should(BeTrue())

		Eventually(func() bool {
			var err error
			userClusterClient, err = legacyOpts.ClusterClientProvider.GetClient(rootCtx, cluster)
			return err == nil
		}, 20*time.Minute, 5*time.Second).Should(BeTrue())
	})

	By(KKP("Add LB and PV Finalizers"), func() {
		Expect(retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			if err := legacyOpts.SeedClusterClient.Get(rootCtx, types.NamespacedName{Name: cluster.Name}, cluster); err != nil {
				return err
			}
			cluster.Finalizers = append(cluster.Finalizers,
				kubermaticv1.InClusterPVCleanupFinalizer,
				kubermaticv1.InClusterLBCleanupFinalizer,
			)
			return legacyOpts.SeedClusterClient.Update(rootCtx, cluster)
		})).NotTo(HaveOccurred(), "failed to add finalizers to the cluster")
	})

	return cluster, userClusterClient
}

func CommonCleanup(rootCtx context.Context, log *zap.SugaredLogger, client clients.Client, scenario scenarios.Scenario, userClusterClient ctrlruntimeclient.Client, cluster *kubermaticv1.Cluster, name string) {
	// By(KKP("Removing machine deployment"))
	// err := client.DeleteMachineDeployments(rootCtx, log, scenario, userClusterClient, cluster)
	// Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed to delete machine deployments with error %v", err))
	By(KKP(fmt.Sprintf("Delete cluster %s.", name)), func() {
		Eventually(func() bool {
			// deleteTimeout := 15 * time.Minute
			cluster = &kubermaticv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			}
			// return client.DeleteCluster(rootCtx, log, cluster, deleteTimeout)
			// err := userClusterClient.Delete(rootCtx, cluster)
			err := userClusterClient.Get(rootCtx, types.NamespacedName{Name: cluster.Name}, cluster)
			if apierrors.IsNotFound(err) {
				// Cluster is already deleted
				log.Infof("Cluster %s is already deleted", cluster.Name)
				return true
			}
			if cluster.DeletionTimestamp == nil {
				// Cluster is already being deleted
				log.Infof("Cluster %s is not being deleted yet", cluster.Name)
				err = userClusterClient.Delete(rootCtx, cluster)
				if err != nil {
					log.Errorf("Failed to delete cluster %s: %v", cluster.Name, err)
				}
				return false

			}
			log.Infof("Waiting for Cluster %s being deleted. Deletion timestamp: %v", cluster.Name, cluster.DeletionTimestamp)
			return false
		}).WithTimeout(25*time.Minute).WithPolling(15*time.Second).Should(BeTrue(), "cluster deletion did not finish within the timeout")
	})
	log.Info("Ending scenario test")
}

func MachineUpdate(rootCtx context.Context, log *zap.SugaredLogger, userClusterClient ctrlruntimeclient.Client, cluster *kubermaticv1.Cluster, clusterName string, scenarioName string, machineSpec *v1alpha1.MachineSpec, legacyOpts *legacytypes.Options) {
	// var err error
	currentVersion := cluster.Spec.Version.Semver()
	By(KKP("Update MachineDeployments"), func() {
		log.Infof("machinedeploymnt name: %s-%s", clusterName[:12], scenarioName[:8])
		log.Infof("machinedeploymnt label: %s=%s-%s", MachineNameLabel, clusterName, scenarioName)
		log.Infof("cluster label: %s=%s", clusterv1alpha1.MachineClusterLabelName, cluster.Name)
		existingMd := &clusterv1alpha1.MachineDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", clusterName[:12], scenarioName[:8]),
				Namespace: "kube-system",
				Labels: map[string]string{
					clusterv1alpha1.MachineClusterLabelName: cluster.Name,
					MachineNameLabel:                        fmt.Sprintf("%s-%s", clusterName, scenarioName),
				},
			},
		}
		err := userClusterClient.Get(rootCtx, types.NamespacedName{Name: existingMd.Name, Namespace: existingMd.Namespace}, existingMd)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed to get existing machine deployment %s: %v", existingMd.Name, err))

		currentMachine := existingMd.Spec.Template.Spec
		currentMachine.Versions.Kubelet = fmt.Sprintf("%s", currentVersion.String())
		existingMd.Spec.Template.Spec = currentMachine
		err = userClusterClient.Update(rootCtx, existingMd)
		Expect(err).ShouldNot(HaveOccurred(), fmt.Sprintf("failed to update existing machine deployment %s: %v", existingMd.Name, err))
	})

	var nodeNames []string
	By(KKP("Wait for machines to get a node"), func() {
		Eventually(func() bool {
			log.Infof("Waiting for machines with updated kubelet version %s to get a node", currentVersion.String())
			nodeNames = nil
			machineList := &clusterv1alpha1.MachineList{}
			s := labels.NewSelector()
			req, err := labels.NewRequirement(MachineNameLabel, selection.Equals, []string{fmt.Sprintf("%s-%s", clusterName, scenarioName)})
			Expect(err).ShouldNot(HaveOccurred())
			s = s.Add(*req)
			if err := userClusterClient.List(rootCtx, machineList, &ctrlruntimeclient.ListOptions{
				LabelSelector: s,
			}); err != nil {
				log.Errorf("Failed to list machines: %v", err)
				return false
			}

			updatedCount := 0
			for _, machine := range machineList.Items {
				if machine.DeletionTimestamp != nil {
					log.Infof("Machine %s is being deleted, skipping", machine.Name)
					continue
				}
				// Only consider machines with the updated kubelet version
				if machine.Spec.Versions.Kubelet != currentVersion.String() {
					log.Infof("Machine %s still has old kubelet version %s, skipping", machine.Name, machine.Spec.Versions.Kubelet)
					continue
				}
				updatedCount++
				if machine.Status.NodeRef == nil {
					log.Infof("Machine %s (updated) does not have a node yet", machine.Name)
					return false
				}
				if !slices.Contains(nodeNames, machine.Status.NodeRef.Name) {
					nodeNames = append(nodeNames, machine.Status.NodeRef.Name)
				}
			}

			if updatedCount < legacyOpts.NodeCount {
				log.Infof("Not all machines have the updated kubelet version yet. Expected: %d, Found: %d", legacyOpts.NodeCount, updatedCount)
				return false
			}

			log.Infof("All %d updated machines have nodes: %v", updatedCount, nodeNames)
			return true
		}).WithTimeout(15*time.Minute).WithPolling(15*time.Second).Should(BeTrue(), "not all machines got a node within the timeout")
	})
	By(KKP("Wait for nodes to be ready"), func() {
		Eventually(func() bool {
			log.Infof("Waiting for nodes to be ready")
			unready := sets.New[string]()
			for _, nodeName := range nodeNames {
				node := &corev1.Node{}
				if err := userClusterClient.Get(rootCtx, types.NamespacedName{Name: nodeName}, node); err != nil {
					return false
				}
				log.Infof("Node %s has status: %v", node.Name, node.Status)

				if !util.NodeIsReady(*node) {
					unready.Insert(node.Name)
				}

				log.Infof("Found %d nodes, %d are not ready", len(nodeNames), unready.Len())
			}
			if unready.Len() == 0 {
				return true
			}
			return false
		}).WithTimeout(15*time.Minute).WithPolling(15*time.Second).Should(BeTrue(), "not all nodes became ready within the timeout")
	})
	By(KKP("Add label to nodes"), func() {
		Eventually(func() bool {
			log.Infof("Adding label to nodes to be ready")
			for _, nodeName := range nodeNames {
				node := &corev1.Node{}
				if err := userClusterClient.Get(rootCtx, types.NamespacedName{Name: nodeName}, node); err != nil {
					return false
				}
				if _, ok := node.Labels[MachineNameLabel]; !ok {
					node.Labels[MachineNameLabel] = fmt.Sprintf("%s-%s", clusterName, scenarioName)
					if err := userClusterClient.Update(rootCtx, node); err != nil {
						return false
					}
				}
				if !util.NodeIsReady(*node) {
					continue
				}
			}
			return true
		}).WithTimeout(15*time.Minute).WithPolling(15*time.Second).Should(BeTrue(), "not all nodes became ready within the timeout")
	})
	By(KKP("Wait for Pods inside usercluster to be ready"), func() {
		Eventually(func() bool {
			log.Infof("Waiting for Pods inside usercluster to be ready")
			for _, nodeName := range nodeNames {
				podList := &corev1.PodList{}
				if err := userClusterClient.List(rootCtx, podList, &ctrlruntimeclient.ListOptions{
					FieldSelector: fields.OneTermEqualSelector("spec.nodeName", nodeName),
				}); err != nil {
					log.Errorf("Failed to list pods on node %s: %v", nodeName, err)
					return false
				}

				unready := sets.New[string]()
				for _, pod := range podList.Items {
					// Ignore pods failing kubelet admission (KKP #6185)
					if !util.PodIsReady(&pod) && !podFailedKubeletAdmissionDueToNodeAffinityPredicate(&pod, log) && !util.PodIsCompleted(&pod) {
						unready.Insert(pod.Name)
					}
				}

				log.Infof("Found %d pods on node %s, %d are not ready", len(podList.Items), nodeName, unready.Len())

				if unready.Len() > 0 {
					log.Infof("Not all pods on node %s are ready yet. Unready pods: %v", nodeName, unready.UnsortedList())
					return false
				}
			}

			return true
		}).WithTimeout(15*time.Minute).WithPolling(15*time.Second).Should(BeTrue(), "not all pods became ready within the timeout")
	})
}

func MachineSetup(rootCtx context.Context, log *zap.SugaredLogger, userClusterClient ctrlruntimeclient.Client, clusterName string, scenarioName string, machineSpec *v1alpha1.MachineSpec, legacyOpts *legacytypes.Options) {
	By(KKP("Create MachineDeployments"), func() {
		err := userClusterClient.Create(rootCtx, &clusterv1alpha1.MachineDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", clusterName[:12], scenarioName[:8]),
				Namespace: "kube-system",
				Labels: map[string]string{
					clusterv1alpha1.MachineClusterLabelName: clusterName,
					MachineNameLabel:                        fmt.Sprintf("%s-%s", clusterName, scenarioName),
				},
			},
			Spec: clusterv1alpha1.MachineDeploymentSpec{
				Replicas: ptr.Int32(int32(legacyOpts.NodeCount)),
				Selector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						MachineNameLabel: fmt.Sprintf("%s-%s", clusterName, scenarioName),
					},
				},
				Template: clusterv1alpha1.MachineTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							MachineNameLabel: fmt.Sprintf("%s-%s", clusterName, scenarioName),
						},
					},
					Spec: *machineSpec},
				Strategy: &clusterv1alpha1.MachineDeploymentStrategy{},
			},
		})
		Expect(err).ShouldNot(HaveOccurred())
	})
	var nodeNames []string
	By(KKP("Wait for machines to get a node"), func() {
		Eventually(func() bool {
			log.Infof("Waiting for machines to get a node")
			machineList := &clusterv1alpha1.MachineList{}
			nodeNames = []string{}
			s := labels.NewSelector()
			req, err := labels.NewRequirement(MachineNameLabel, selection.Equals, []string{fmt.Sprintf("%s-%s", clusterName, scenarioName)})
			Expect(err).ShouldNot(HaveOccurred())
			s = s.Add(*req)
			if err := userClusterClient.List(rootCtx, machineList, &ctrlruntimeclient.ListOptions{
				LabelSelector: s,
			}); err != nil {
				return false
			}
			log.Infof("Found %d machines. With Label: %v Names: %v", len(machineList.Items), s.String(), getMachineNames(machineList.Items))
			if len(machineList.Items) < legacyOpts.NodeCount {
				return false
			}

			for _, machine := range machineList.Items {
				log.Infof("Machine %s has status: %v", machine.Name, machine.Status)
				if machine.Status.NodeRef == nil || machine.Status.NodeRef.Name == "" {
					return false
				}
			}

			for _, machine := range machineList.Items {
				nodeNames = append(nodeNames, machine.Status.NodeRef.Name)
			}

			return true
		}).WithTimeout(10*time.Minute).WithPolling(15*time.Second).Should(BeTrue(), "not all machines got a node within the timeout")
	})
	By(KKP("Add label to nodes"), func() {
		Eventually(func() bool {
			log.Infof("Adding label to nodes to be ready")
			for _, nodeName := range nodeNames {
				node := &corev1.Node{}
				if err := userClusterClient.Get(rootCtx, types.NamespacedName{Name: nodeName}, node); err != nil {
					return false
				}
				if !util.NodeIsReady(*node) {
					continue
				}
				node.Labels[MachineNameLabel] = fmt.Sprintf("%s-%s", clusterName, scenarioName)
				if err := userClusterClient.Update(rootCtx, node); err != nil {
					return false
				}
			}
			return true
		}).WithTimeout(5*time.Minute).WithPolling(15*time.Second).Should(BeTrue(), "not all nodes became ready within the timeout")
	})
	By(KKP("Wait for nodes to be ready"), func() {
		Eventually(func() bool {
			log.Infof("Waiting for nodes to be ready")
			unready := sets.New[string]()
			for _, nodeName := range nodeNames {
				node := &corev1.Node{}
				if err := userClusterClient.Get(rootCtx, types.NamespacedName{Name: nodeName}, node); err != nil {
					return false
				}
				log.Infof("Node %s has status: %v", node.Name, node.Status)

				if !util.NodeIsReady(*node) {
					unready.Insert(node.Name)
				}

				log.Infof("Found %d nodes, %d are not ready", len(nodeNames), unready.Len())
			}
			if unready.Len() == 0 {
				return true
			}
			return false
		}).WithTimeout(5*time.Minute).WithPolling(15*time.Second).Should(BeTrue(), "not all nodes became ready within the timeout")
	})
	By(KKP("Wait for Pods inside usercluster to be ready"), func() {
		Eventually(func() bool {
			log.Infof("Waiting for Pods inside usercluster to be ready")
			for _, nodeName := range nodeNames {
				podList := &corev1.PodList{}
				if err := userClusterClient.List(rootCtx, podList, &ctrlruntimeclient.ListOptions{
					FieldSelector: fields.OneTermEqualSelector("spec.nodeName", nodeName),
				}); err != nil {
					return false
				}

				unready := sets.New[string]()
				for _, pod := range podList.Items {
					// Ignore pods failing kubelet admission (KKP #6185)
					if !util.PodIsReady(&pod) && !podFailedKubeletAdmissionDueToNodeAffinityPredicate(&pod, log) && !util.PodIsCompleted(&pod) {
						unready.Insert(pod.Name)
					}
				}

				log.Infof("Found %d pods on node %s, %d are not ready", len(podList.Items), nodeName, unready.Len())

				if unready.Len() > 0 {
					return false
				}
			}

			return true
		}).WithTimeout(5*time.Minute).WithPolling(15*time.Second).Should(BeTrue(), "not all pods became ready within the timeout")
	})
	// By(KKP("Wait for addons"), func() {
	// 	Eventually(func() bool {
	// 		addons := kubermaticv1.AddonList{}
	// 		if err := legacyOpts.SeedClusterClient.List(rootCtx, &addons, ctrlruntimeclient.InNamespace(fmt.Sprintf("cluster-%s", clusterName))); err != nil {
	// 			return false
	// 		}

	// 		unhealthyAddons := sets.New[string]()
	// 		for _, addon := range addons.Items {
	// 			if addon.Status.Conditions[kubermaticv1.AddonReconciledSuccessfully].Status != corev1.ConditionTrue {
	// 				unhealthyAddons.Insert(addon.Name)
	// 			}
	// 		}

	// 		if unhealthyAddons.Len() > 0 {
	// 			return false
	// 		}

	// 		return true
	// 	}, 2*time.Minute, 2*time.Second).Should(BeTrue(), "not all addons became healthy within the timeout")
	// })
}

// podFailedKubeletAdmissionDueToNodeAffinityPredicate detects a condition in
// which a pod is scheduled but fails kubelet admission due to a race condition
// between scheduler and kubelet.
// see: https://github.com/kubernetes/kubernetes/issues/93338
func podFailedKubeletAdmissionDueToNodeAffinityPredicate(p *corev1.Pod, log *zap.SugaredLogger) bool {
	failedAdmission := p.Status.Phase == "Failed" && p.Status.Reason == "NodeAffinity"
	if failedAdmission {
		log.Infow("pod failed kubelet admission due to NodeAffinity predicate", "pod", *p)
	}

	return failedAdmission
}

func getMachineNames(machines []v1alpha1.Machine) []string {
	names := make([]string, 0, len(machines))
	for _, machine := range machines {
		names = append(names, machine.Name)
	}
	return names
}
