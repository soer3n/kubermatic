package kubevirt

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"iter"
	"os"
	"sort"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/retry"
	"k8s.io/kubectl/pkg/util/slice"

	. "github.com/onsi/gomega"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/tests"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/util"
	"k8c.io/kubermatic/v2/pkg/version"
	"k8c.io/kubermatic/v2/pkg/version/kubermatic"
	"k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	"k8c.io/machine-controller/sdk/providerconfig/configvar"

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

	kkpConfig, err := k8cginkgo.LoadKubermaticConfiguration()
	if err != nil {
		log.Fatalw("Failed to load KKP configuration", zap.Error(err))
	}
	file, err := os.Create("debug_output.txt")
	if err != nil {
		log.Fatalw("Failed to create debug output file", zap.Error(err))
	}
	defer file.Close()
	log.Info("generating seeds...")
	datacenterNameMappings = make(map[string]string)
	defaultSeedSettings = buildDefaultSeedSettings(GenericDatacenterSettings(rootCtx, runtimeOpts.SeedClusterClient, legacyOpts.KubermaticNamespace), kkpConfig, log, defaultDatacenterSettings, opts.Excluded.DatacenterDescriptions, opts.Included.DatacenterDescriptions)
	seed = &kubermaticv1.Seed{}
	err = runtimeOpts.SeedClusterClient.Get(rootCtx, apitypes.NamespacedName{Name: "kubermatic", Namespace: "kubermatic"}, seed)
	if err != nil {
		log.Fatalw("Failed to get seed", zap.Error(err))
	}

	if seed.Spec.Datacenters == nil {
		seed.Spec.Datacenters = map[string]kubermaticv1.Datacenter{}
	}

	seedKeys := make([]string, 0, len(defaultSeedSettings))
	for k := range defaultSeedSettings {
		seedKeys = append(seedKeys, k)
	}
	sort.Strings(seedKeys)

	for _, key := range seedKeys {
		s := defaultSeedSettings[key]
		for dcName, dc := range s.Spec.Datacenters {
			hasher := sha1.New()
			hasher.Write([]byte(dcName))
			hashedName := hex.EncodeToString(hasher.Sum(nil))[:10]
			datacenterNameMappings[dcName] = hashedName
			dc.Country = "conformance"
			dc.Location = dcName
			seed.Spec.Datacenters[hashedName] = dc
		}
	}

	versionManager := version.NewFromConfiguration(kkpConfig)
	versions, err := versionManager.GetVersionsForProvider(kubermaticv1.KubevirtCloudProvider)
	versions = []*version.Version{}
	for _, v := range opts.Releases {
		versionObj, _ := versionManager.GetVersion(v)
		versions = append(versions, versionObj)
	}
	log.Info("generating clusters...")
	newClusters, finalClusterDescriptions = buildNewClusters(rootCtx, versions, k8cginkgo.ClusterSettings, defaultSeedSettings, seed, opts, kkpConfig, log, versionManager, file, opts.Excluded.ClusterDescriptions, opts.Included.ClusterDescriptions)
	resolver := configvar.NewResolver(rootCtx, runtimeOpts.SeedClusterClient)
	fmt.Fprintf(file, "\nGenerated Clusters: %v\n", len(newClusters))
	defaultKubevirtConfig, err := getDefaultKubevirtConfig()
	if err != nil {
		log.Fatalw("Failed to get default kubevirt config", zap.Error(err))
	}
	fmt.Fprintf(file, "Default KubeVirt Config: %+v\n", defaultKubevirtConfig)
	fmt.Fprint(file, "\nGenerated Scenarios:\n")
	log.Info("generating scenarios...")
	newScenarios, finalMachineDescriptions = buildNewScenarios(MachineSettings(rootCtx, runtimeOpts.SeedClusterClient, legacyOpts.KubermaticNamespace, &opts.Resources), newClusters, opts, log, *defaultKubevirtConfig, resolver, file, rootCtx, opts.Excluded.MachineDescriptions, opts.Included.MachineDescriptions)

	var newEntries []TableEntry
	versionSlice := []string{}
	if len(opts.Releases) > 0 {
		for _, v := range opts.Releases {
			versionSlice = append(versionSlice, v)
		}
	} else {
		for _, scenario := range kkpConfig.Spec.Versions.Versions {
			versionSlice = append(versionSlice, scenario.String())
		}
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
				title := fmt.Sprintf("kubernetes version %s and %s and %s and %s", clusterSpec.Version.String(), seedKey, strings.Join(clusterDesc, " and "), strings.Join(desc, " and "))
				entry := Entry(title, title, clusterName, clusterSpec, &machine, scenario, Label("kubevirt"))
				if !slice.ContainsString(versionSlice, clusterSpec.Version.String(), nil) {
					entry = Entry(title, title, clusterName, clusterSpec, &machine, scenario, Label("skip"))
					newEntries = append(newEntries, entry)
					continue
				}

				exclude := false
				if len(opts.Included.DatacenterDescriptions) > 0 {
					for _, included := range opts.Included.DatacenterDescriptions {
						if !strings.Contains(title, included) {
							exclude = true
							break
						}
					}
				} else {
					for _, excluded := range opts.Excluded.DatacenterDescriptions {
						if strings.Contains(title, excluded) {
							exclude = true
							break
						}
					}
				}

				if !exclude {
					if len(opts.Included.ClusterDescriptions) > 0 {
						for _, included := range opts.Included.ClusterDescriptions {
							if !strings.Contains(title, included) {
								exclude = true
								break
							}
						}
					} else {
						for _, excluded := range opts.Excluded.ClusterDescriptions {
							if strings.Contains(title, excluded) {
								exclude = true
								break
							}
						}
					}
				}

				if !exclude {
					if len(opts.Included.MachineDescriptions) > 0 {
						for _, included := range opts.Included.MachineDescriptions {
							if !strings.Contains(title, included) {
								exclude = true
								break
							}
						}
					} else {
						for _, excluded := range opts.Excluded.MachineDescriptions {
							if strings.Contains(title, excluded) {
								exclude = true
								break
							}
						}
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
	DescribeTable("KubeVirt", func(description string, clusterName string, clusterSpec *kubermaticv1.ClusterSpec, machineSpec *v1alpha1.MachineSpec, scenarioName string) {
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
		By(fmt.Sprintf("Running smoke tests %q (enabled: %v) (%v)", clusterName, legacyOpts.EnableTests, opts.EnableTests), func() {
			n := 0
			ExpectWithOffset(3, tests.TestStorage(rootCtx, log, legacyOpts, cluster, map[string]string{
				k8cginkgo.MachineNameLabel: fmt.Sprintf("machine-%s", clusterName),
			}, userClusterClient, n+1)).To(BeNil())
			n = 0
			ExpectWithOffset(3, tests.TestLoadBalancer(rootCtx, log, legacyOpts, cluster, map[string]string{
				k8cginkgo.MachineNameLabel: fmt.Sprintf("machine-%s", clusterName),
			}, userClusterClient, n+1)).To(BeNil())
			ExpectWithOffset(3, tests.TestUserClusterMetrics(rootCtx, log, legacyOpts, cluster, userClusterClient)).To(BeNil())
			ExpectWithOffset(3, tests.TestUserclusterControllerRBAC(rootCtx, log, legacyOpts, cluster, userClusterClient, runtimeOpts.SeedClusterClient)).To(BeNil())
			ExpectWithOffset(3, tests.TestUserClusterNoK8sGcrImages(rootCtx, log, legacyOpts, cluster, userClusterClient)).To(BeNil())
			ExpectWithOffset(3, tests.TestUserClusterPodAndNodeMetrics(rootCtx, log, legacyOpts, cluster, map[string]string{
				k8cginkgo.MachineNameLabel: fmt.Sprintf("machine-%s", clusterName),
			}, userClusterClient)).To(BeNil())
			ExpectWithOffset(3, tests.TestUserClusterSeccompProfiles(rootCtx, log, legacyOpts, cluster, userClusterClient)).To(BeNil())
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
}

func updateCluster(name string, spec *kubermaticv1.ClusterSpec) {
	// st := getClusterState(name)
	By(fmt.Sprintf("Updating cluster %s\n", name))
	// simulate work
	time.Sleep(2 * time.Second)
	By(fmt.Sprintf("Cluster %s updated\n", name))
}
