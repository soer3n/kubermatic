package kubevirt

import (
	"fmt"
	"iter"
	"maps"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/retry"
	"k8s.io/kubectl/pkg/util/slice"

	. "github.com/onsi/gomega"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/tests"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/util"
	"k8c.io/kubermatic/v2/pkg/version/kubermatic"
	"k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"

	controllerutil "k8c.io/kubermatic/v2/pkg/controller/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

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
	versionSlice := []string{}
	for _, v := range opts.Releases {
		versionSlice = append(versionSlice, v)
	}
	for seedKey, _ := range defaultSeedSettings {
		for clusterName, machines := range newScenarios {
			clusterSpec, ok := newClusters[clusterName]
			if !ok {
				continue
			}
			clusterDesc, ok := finalClusterDescriptions[clusterName]
			if !ok {
				continue
			}

			for scenario, machine := range machines {
				desc, ok := finalMachineDescriptions[clusterName][scenario]
				if !ok {
					continue
				}
				title := fmt.Sprintf("%s and %s and %s", strings.Replace(seedKey, "-", " and ", -1), strings.Join(clusterDesc, " and "), strings.Join(desc, " and "))
				entry := Entry(title, title, clusterName, clusterSpec, &machine, scenario, Label("kubevirt"))
				if !slice.ContainsString(versionSlice, clusterSpec.Version.String(), nil) {
					entry = Entry(title, title, clusterName, clusterSpec, &machine, scenario, Label("skip"))
				}

				exclude := false
				for _, excluded := range opts.DatacenterDescriptions {
					if strings.Contains(title, excluded) {
						exclude = true
						break
					}
				}

				for _, excluded := range opts.ClusterDescriptions {
					if strings.Contains(title, excluded) {
						exclude = true
						break
					}
				}

				if exclude {
					entry = Entry(title, title, clusterName, clusterSpec, &machine, scenario, Label("skip"))
				}

				newEntries = append(newEntries, entry)
			}
			continue
		}
	}
	return newEntries
}

var _ = Describe("Scenario", func() {
	var disabledSettings map[string][]interface{}
	if len(legacyOpts.TestSettings) > 0 {
		disabledSettings = make(map[string][]interface{})
		for _, s := range legacyOpts.TestSettings {
			disabledSettings[strings.TrimSpace(s.Description)] = s.StringVariants
		}
	}
	versionSlice := []string{}
	for _, v := range opts.Releases {
		versionSlice = append(versionSlice, v)
	}
	DescribeTable("KubeVirt", func(description string, clusterName string, clusterSpec *kubermaticv1.ClusterSpec, machineSpec *v1alpha1.MachineSpec, scenarioName string) {
		runTest := true
		if disabledSettings != nil {
			runTest = runTestFunc(maps.Keys(disabledSettings), description)
			log.Infof("Considering test setting %q for provider %q: runTest=%v", description, "kubevirt", runTest)
			log.Infof("Datacenters: %v", dcClusters)
		}
		if !runTest {
			Skip("This test setting was not selected to run via the --test-settings flag.")
		}

		exclude := false
		for _, excluded := range opts.DatacenterDescriptions {
			if strings.Contains(description, excluded) {
				exclude = true
				break
			}
		}

		if exclude {
			Skip(fmt.Sprintf("Excluding test setting %q via the --exclude-machine-descriptions flag.", description))
		}

		cluster := &kubermaticv1.Cluster{}
		if err := runtimeOpts.SeedClusterClient.Get(rootCtx, types.NamespacedName{Name: fmt.Sprintf("%s-%s", clusterName, clusterSpec.Cloud.DatacenterName[:8])}, cluster); err != nil {
			log.Errorf("Failed to get cluster %s: %v", clusterName, err)
			Fail(fmt.Sprintf("Failed to get cluster %s: %v", clusterName, err))
		}

		userClusterClient, err := runtimeOpts.ClusterClientProvider.GetClient(rootCtx, cluster)
		if err != nil {
			log.Errorf("Failed to get user cluster client for cluster %s: %v", clusterName, err)
			Fail(fmt.Sprintf("Failed to get user cluster client for cluster %s: %v", clusterName, err))
		}

		By(fmt.Sprintf("Running tests for datacenter %q kubeVersion %q", clusterSpec.Cloud.DatacenterName, clusterSpec.Version.String()))
		By(fmt.Sprintf("Scenario with dc %q cluster %q", clusterSpec.Cloud.DatacenterName, clusterName))
		By(fmt.Sprintf("Setting up machine for %s %s", clusterSpec.Cloud.DatacenterName, clusterName), func() {
			k8cginkgo.MachineSetup(rootCtx, log, userClusterClient, clusterName, scenarioName, machineSpec, legacyOpts)
		})

		By(fmt.Sprintf("Machine setup done %q", clusterName))
		By(fmt.Sprintf("Running smoke tests %q", clusterName), func() {
			n := 0
			ExpectWithOffset(3, tests.TestStorage(rootCtx, log, legacyOpts, cluster, map[string]string{
				k8cginkgo.MachineNameLabel: fmt.Sprintf("machine-%s", clusterName),
			}, userClusterClient, n+1), nil)
			n = 0
			ExpectWithOffset(3, tests.TestLoadBalancer(rootCtx, log, legacyOpts, cluster, map[string]string{
				k8cginkgo.MachineNameLabel: fmt.Sprintf("machine-%s", clusterName),
			}, userClusterClient, n+1), nil)
			ExpectWithOffset(3, tests.TestUserClusterMetrics(rootCtx, log, legacyOpts, cluster, userClusterClient), nil)
			ExpectWithOffset(3, tests.TestUserclusterControllerRBAC(rootCtx, log, legacyOpts, cluster, userClusterClient, runtimeOpts.SeedClusterClient), nil)
			ExpectWithOffset(3, tests.TestUserClusterNoK8sGcrImages(rootCtx, log, legacyOpts, cluster, userClusterClient), nil)
			ExpectWithOffset(3, tests.TestUserClusterPodAndNodeMetrics(rootCtx, log, legacyOpts, cluster, map[string]string{
				k8cginkgo.MachineNameLabel: fmt.Sprintf("machine-%s", clusterName),
			}, userClusterClient), nil)
			ExpectWithOffset(3, tests.TestUserClusterSeccompProfiles(rootCtx, log, legacyOpts, cluster, userClusterClient), nil)
		})
		By(fmt.Sprintf("Smoke tests done %q", clusterName))
		time.Sleep(500 * time.Millisecond)
	}, getTableEntries())
})

func ensureCluster(name string, dcName string, projectName string, spec *kubermaticv1.ClusterSpec) {
	By(fmt.Sprintf("Ensuring cluster %s\n", name))
	currentOpts := *legacyOpts // copy
	currentOpts.Secrets.Kubevirt.KKPDatacenter = name
	spec.Cloud.DatacenterName = dcName
	cluster := &kubermaticv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s", name, dcName[:8]), Labels: map[string]string{kubermaticv1.ProjectIDLabelKey: projectName}},
		Spec:       *spec,
	}
	By(k8cginkgo.KKP("Create Cluster"), func() {
		err := runtimeOpts.SeedClusterClient.Create(rootCtx, cluster)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed to create cluster %s: %v", name, err))
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
		}, 5*time.Minute, 5*time.Second).Should(BeTrue(), "cluster was not reconciled successfully within the timeout")

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
		}, 5*time.Minute, 5*time.Second).Should(BeTrue(), "cluster did not become healthy within the timeout")
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
			_, err = legacyOpts.ClusterClientProvider.GetClient(rootCtx, cluster)
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
	// newClusterClients[name] = userClusterClient
	clustersToDelete = append(clustersToDelete, cluster.Name)
}

func updateCluster(name string, spec *kubermaticv1.ClusterSpec) {
	// st := getClusterState(name)
	By(fmt.Sprintf("Updating cluster %s\n", name))
	// simulate work
	time.Sleep(2 * time.Second)
	By(fmt.Sprintf("Cluster %s updated\n", name))
}
