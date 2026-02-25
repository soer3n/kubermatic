package cluster

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/retry"

	. "github.com/onsi/gomega"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/options"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/tests"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/util"
	"k8c.io/kubermatic/v2/pkg/version/kubermatic"

	controllerutil "k8c.io/kubermatic/v2/pkg/controller/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func Ensure(rootCtx context.Context,
	log *zap.SugaredLogger,
	name string,
	dcName string,
	projectName string,
	spec *kubermaticv1.ClusterSpec,
	legacyOpts *legacytypes.Options,
	runtimeOpts *options.RuntimeOptions,
	opts *options.Options,
) {
	By(fmt.Sprintf("Ensuring cluster %s\n", name))
	currentOpts := *legacyOpts // copy
	currentOpts.Secrets.Kubevirt.KKPDatacenter = name
	spec.Cloud.DatacenterName = dcName
	cluster := &kubermaticv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: map[string]string{kubermaticv1.ProjectIDLabelKey: projectName}},
		Spec:       *spec,
	}

	By(k8cginkgo.KKP("Create Cluster"), func() {
		err := runtimeOpts.SeedClusterClient.Create(rootCtx, cluster)
		if !errors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed to create cluster %s: %v", name, err))
		}
	})

	By(k8cginkgo.KKP("Wait for successful reconciliation"), func() {
		versions := kubermatic.GetVersions()
		log.Info("Waiting for cluster to be successfully reconciled...")

		Eventually(func() bool {
			if err := runtimeOpts.SeedClusterClient.Get(rootCtx, types.NamespacedName{Name: cluster.Name}, cluster); err != nil {
				return false
			}

			missingConditions, _ := controllerutil.ClusterReconciliationSuccessful(cluster, versions, true)
			if len(missingConditions) > 0 {
				return false
			}

			return true
		}).WithTimeout(5*time.Minute).WithPolling(15*time.Second).Should(BeTrue(), "cluster was not reconciled successfully within the timeout")

		Eventually(func() bool {
			newCluster := &kubermaticv1.Cluster{}
			namespacedClusterName := types.NamespacedName{Name: cluster.Name}
			if err := runtimeOpts.SeedClusterClient.Get(rootCtx, namespacedClusterName, newCluster); err != nil {
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
		}).WithTimeout(10*time.Minute).WithPolling(15*time.Second).Should(BeTrue(), "cluster did not become healthy within the timeout")
	})

	By(k8cginkgo.KKP("Wait for control plane"), func() {
		Eventually(func() bool {
			if err := runtimeOpts.SeedClusterClient.Get(rootCtx, types.NamespacedName{Name: cluster.Name}, cluster); err != nil {
				return false
			}
			versions := kubermatic.GetVersions()
			// ignore Kubermatic version in this check, to allow running against a 3rd party setup
			missingConditions, _ := controllerutil.ClusterReconciliationSuccessful(cluster, versions, true)
			return len(missingConditions) == 0
		}, 10*time.Minute, 5*time.Second).Should(BeTrue())

		Eventually(func() bool {
			var err error
			_, err = runtimeOpts.ClusterClientProvider.GetClient(rootCtx, cluster)
			return err == nil
		}).WithTimeout(10 * time.Minute).WithPolling(15 * time.Second).Should(BeTrue())
	})

	By(k8cginkgo.KKP("Add LB and PV Finalizers"), func() {
		Expect(retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			if err := runtimeOpts.SeedClusterClient.Get(rootCtx, types.NamespacedName{Name: cluster.Name}, cluster); err != nil {
				return err
			}
			cluster.Finalizers = append(cluster.Finalizers,
				kubermaticv1.InClusterPVCleanupFinalizer,
				kubermaticv1.InClusterLBCleanupFinalizer,
			)
			return legacyOpts.SeedClusterClient.Update(rootCtx, cluster)
		})).NotTo(HaveOccurred(), "failed to add finalizers to the cluster")
	})

	By(fmt.Sprintf("Running smoke cluster tests %q (enabled: %v) (%v)", cluster.Name, legacyOpts.EnableTests, opts.EnableTests), func() {
		userClusterClient, err := runtimeOpts.ClusterClientProvider.GetClient(rootCtx, cluster)
		if err != nil {
			log.Errorf("Failed to get user cluster client for cluster %s: %v", cluster.Name, err)
			Fail(fmt.Sprintf("Failed to get user cluster client for cluster %s: %v", cluster.Name, err))
		}
		// ExpectWithOffset(3, tests.TestUserClusterMetrics(rootCtx, log, legacyOpts, cluster, userClusterClient)).To(BeNil())
		ExpectWithOffset(3, tests.TestUserclusterControllerRBAC(rootCtx, log, legacyOpts, cluster, userClusterClient, runtimeOpts.SeedClusterClient)).To(BeNil())
		ExpectWithOffset(3, tests.TestUserClusterNoK8sGcrImages(rootCtx, log, legacyOpts, cluster, userClusterClient)).To(BeNil())
		ExpectWithOffset(3, tests.TestUserClusterSeccompProfiles(rootCtx, log, legacyOpts, cluster, userClusterClient)).To(BeNil())
	})
}
