package kubevirt

import (
	"fmt"
	"iter"
	"maps"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"k8c.io/kubermatic/sdk/v2/semver"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/retry"

	. "github.com/onsi/gomega"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/util"
	"k8c.io/kubermatic/v2/pkg/version/kubermatic"
	"k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"

	controllerutil "k8c.io/kubermatic/v2/pkg/controller/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Declare clusterName here; it is assigned a value in SynchronizedBeforeSuite in the suite file.
var clusterName string

type scenario struct {
	dc          string
	kubeVersion string
	clusterName string // e2e-<dc>-<kv>
}

func ScenarioTableEntries(_ map[string]*kubermaticv1.Datacenter) []scenario {
	var scenarios []scenario
	for name := range dcClusters { // use filtered dcs slice from suite
		dc := strings.Split(name, "-")[1]
		kv := strings.Split(name, "-")[2]
		scenarios = append(scenarios, scenario{dc: dc, kubeVersion: kv, clusterName: name})
	}
	return scenarios
}

var _ = ReportAfterEach(func(f SpecContext, r SpecReport) {
	By("Report after smoke tests")
})

var runTestFunc = func(k iter.Seq[string], v string) bool {
	for item := range k {
		if strings.Contains(v, item) {
			return false
		}
	}
	return true
}

func getTableEntries() []TableEntry {
	var newEntries []TableEntry
	for clusterName, machines := range newScenarios {
		clusterClient, ok := newClusterClients[clusterName]
		if !ok {
			continue
		}
		clusterSpec, ok := newClusters[clusterName]
		if !ok {
			continue
		}

		for scenario, machine := range machines {
			newEntries = append(newEntries, Entry(fmt.Sprintf("%s and %s", clusterName, scenario), fmt.Sprintf("%s and %s", clusterName, scenario), clusterSpec.Cloud.DatacenterName, clusterName, clusterSpec.Version, &machine, clusterClient))
		}
		continue
	}
	return newEntries
}

// clusterName is defined in the suite file (provider_suite_test.go) and broadcast via SynchronizedBeforeSuite.
// We only track whether creation happened in this process group.
var _ = Describe("Scenarios", func() {
	var disabledSettings map[string][]interface{}
	if len(legacyOpts.TestSettings) > 0 {
		disabledSettings = make(map[string][]interface{})
		for _, s := range legacyOpts.TestSettings {
			disabledSettings[strings.TrimSpace(s.Description)] = s.StringVariants
		}
	}
	DescribeTable("KubeVirt", func(description string, clusterName string, datacenterName string, clusterVersion semver.Semver, machineSpec *v1alpha1.MachineSpec, clusterClient ctrlruntimeclient.Client) {
		runTest := true
		if disabledSettings != nil {
			runTest = runTestFunc(maps.Keys(disabledSettings), description)
			log.Infof("Considering test setting %q for provider %q: runTest=%v", description, "kubevirt", runTest)
			log.Infof("Datacenters: %v", dcClusters)
		}
		if description == "VirtualMachine.EvictionStrategy" {
			log.Infof("Special case handling for VirtualMachine.EvictionStrategy")
		}
		if !runTest {
			Skip("This test setting was not selected to run via the --test-settings flag.")
		}
		if _, ok := dcClusters[fmt.Sprintf("e2e-%s-%s", datacenterName, clusterVersion.String())]; !ok {
			Skip(fmt.Sprintf("Datacenter %q with kubeVersion %q not selected to run via --datacenters or --kube-versions", datacenterName, clusterVersion.String()))
		}

		// runtimeOpts.ClusterClientProvider.GetClient()

		By(fmt.Sprintf("Running tests for datacenter %q kubeVersion %q", datacenterName, clusterVersion.String()))
		By(fmt.Sprintf("Scenario with dc %q cluster %q", datacenterName, clusterName))
		By(fmt.Sprintf("Setting up machine for %s %s", datacenterName, clusterName))
		By(fmt.Sprintf("Machine setup done %q", clusterName))
		By(fmt.Sprintf("Running smoke tests %q", clusterName))
		By(fmt.Sprintf("Smoke tests done %q", clusterName))
		time.Sleep(500 * time.Millisecond)
	}, getTableEntries())
})

func ensureCluster(name string, spec *kubermaticv1.ClusterSpec) {
	By(fmt.Sprintf("Ensuring cluster %s\n", name))
	currentOpts := *legacyOpts // copy
	currentOpts.Secrets.Kubevirt.KKPDatacenter = strings.Split(name, "-")[1]
	var userClusterClient ctrlruntimeclient.Client
	var cluster *kubermaticv1.Cluster
	By(k8cginkgo.KKP("Create Cluster"), func() {
		err := runtimeOpts.SeedClusterClient.Create(rootCtx, cluster)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed to create cluster %s: %v", name, err))
	})
	By(k8cginkgo.KKP("Wait for successful reconciliation"), func() {
		versions := kubermatic.GetVersions()
		log.Info("Waiting for cluster to be successfully reconciled...")

		Eventually(func() bool {
			if err := legacyOpts.SeedClusterClient.Get(rootCtx, types.NamespacedName{Name: cluster.Name}, cluster); err != nil {
				return false
			}

			missingConditions, _ := controllerutil.ClusterReconciliationSuccessful(cluster, versions, true)
			if len(missingConditions) > 0 {
				return false
			}

			return true
		}, legacyOpts.ControlPlaneReadyWaitTimeout, 5*time.Second).Should(BeTrue(), "cluster was not reconciled successfully within the timeout")

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
		}, legacyOpts.ControlPlaneReadyWaitTimeout, 5*time.Second).Should(BeTrue(), "cluster did not become healthy within the timeout")
	})

	By(k8cginkgo.KKP("Wait for control plane"), func() {
		Eventually(func() bool {
			if err := legacyOpts.SeedClusterClient.Get(rootCtx, types.NamespacedName{Name: cluster.Name}, cluster); err != nil {
				return false
			}
			versions := kubermatic.GetVersions()
			// ignore Kubermatic version in this check, to allow running against a 3rd party setup
			missingConditions, _ := controllerutil.ClusterReconciliationSuccessful(cluster, versions, true)
			return len(missingConditions) == 0
		}, 10*time.Minute, 5*time.Second).Should(BeTrue())

		Eventually(func() bool {
			var err error
			userClusterClient, err = legacyOpts.ClusterClientProvider.GetClient(rootCtx, cluster)
			return err == nil
		}, 10*time.Minute, 5*time.Second).Should(BeTrue())
	})

	By(k8cginkgo.KKP("Add LB and PV Finalizers"), func() {
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
	newClusterClients[name] = userClusterClient
	clustersToDelete = append(clustersToDelete, cluster.Name)
}

func updateCluster(name string, spec *kubermaticv1.ClusterSpec) {
	// st := getClusterState(name)
	By(fmt.Sprintf("Updating cluster %s\n", name))
	// simulate work
	time.Sleep(2 * time.Second)
	By(fmt.Sprintf("Cluster %s updated\n", name))
	// AddReportEntry("cluster-ensure", fmt.Sprintf("cluster=%s created-now ready=%v", name, st.ready.Load()))

}
