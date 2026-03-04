package kubevirt

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	ginkgotypes "github.com/onsi/ginkgo/v2/types"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/gomega"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/build"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/utils"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/tests"
	"k8c.io/machine-controller/sdk/providerconfig"
)

var _ = ReportAfterEach(func(f SpecContext, r SpecReport) {
	if r.State.Is(ginkgotypes.SpecStateSkipped) {
		return
	}
	By("Report after smoke tests")
})

var _ = Describe("KubeVirt", func() {
	var includedMachineDescriptions, excludedMachineDescriptions map[string]map[string][]string
	entries, includedMachineDescriptions, excludedMachineDescriptions, datacenterNameMappings, seed, seedKeys = build.GetTableEntries(rootCtx, log, runtimeOpts, legacyOpts, opts, infraClient, projectName, providerconfig.CloudProviderKubeVirt)
	for name, entry := range entries {
		Describe(fmt.Sprintf("with kubernetes version %s", entry.ClusterSpec.Version), func() {
			for k, v := range entry.Machines {
				label := Label("kubevirt")
				if entry.Exclude {
					label = Label("skip")
				}
				clusterLabel := Label(entry.ClusterName)
				machineDescription := []string{}
				if entry.Exclude {
					machineDescription = excludedMachineDescriptions[entry.ClusterName][k]

				} else {
					machineDescription = includedMachineDescriptions[entry.ClusterName][k]

				}

				// Extract OS distribution from the machine's provider config
				osDistro := "unknown"
				if v.ProviderSpec.Value != nil {
					var pc providerconfig.Config
					if err := json.Unmarshal(v.ProviderSpec.Value.Raw, &pc); err == nil {
						osDistro = string(pc.OperatingSystem)
					}
				}

				It(fmt.Sprintf("%s and operating system set to %v and datacenter %v and machine %v", entry.Description, osDistro, seed.Spec.Datacenters[entry.ClusterSpec.Cloud.DatacenterName].Location, strings.Join(machineDescription, " and ")), label, clusterLabel, func() {
					cluster := &kubermaticv1.Cluster{}
					if err := runtimeOpts.SeedClusterClient.Get(rootCtx, types.NamespacedName{Name: entry.ClusterName}, cluster); err != nil {
						log.Errorf("Failed to get cluster %s: %v", name, err)
						Fail(fmt.Sprintf("Failed to get cluster %s: %v", name, err))
					}

					userClusterClient, err := runtimeOpts.ClusterClientProvider.GetClient(rootCtx, cluster)
					if err != nil {
						log.Errorf("Failed to get user cluster client for cluster %s: %v", name, err)
						Fail(fmt.Sprintf("Failed to get user cluster client for cluster %s: %v", name, err))
					}
					By(fmt.Sprintf("Running tests for datacenter %q kubeVersion %q", entry.ClusterSpec.Cloud.DatacenterName, entry.ClusterSpec.Version.String()))
					By(fmt.Sprintf("Scenario with dc %q cluster %q", entry.ClusterSpec.Cloud.DatacenterName, name))
					if skipClusterCreation && updateClusters {
						By(fmt.Sprintf("Updating machine %s for %s %s. Scenario: %s", k, entry.ClusterSpec.Cloud.DatacenterName, name, entry.Description), func() {
							utils.MachineUpdate(rootCtx, log, userClusterClient, cluster, name, k[:8], &v, legacyOpts)
						})
					} else {
						By(fmt.Sprintf("Setting up machine %s for %s %s. Scenario: %s", k, entry.ClusterSpec.Cloud.DatacenterName, name, entry.Description), func() {
							utils.MachineSetup(rootCtx, log, userClusterClient, name, k[:8], &v, legacyOpts)
						})
					}

					By(fmt.Sprintf("Machine setup done %q", name))
					attemp := 1
					if skipClusterCreation && updateClusters {
						attemp += attemp
					}
					By(fmt.Sprintf("Running smoke tests %q (enabled: %v) (%v)", name, legacyOpts.EnableTests, opts.EnableTests), func() {
						var err error
						legacyOpts.CustomTestTimeout = 10 * time.Minute
						machineLabel := fmt.Sprintf("%s-%s", name, k[:8])
						nodeSelector := map[string]string{
							utils.MachineNameLabel: machineLabel,
						}
						testPrefix := fmt.Sprintf("%s-%s", name[:8], k[:8])
						err = tests.TestStorage(rootCtx, log, legacyOpts, cluster, nodeSelector, userClusterClient, testPrefix, attemp)
						Expect(err).NotTo(HaveOccurred(), "PersistentVolume test failed after multiple attempts")

						// Do a simple LoadBalancer test
						err = tests.TestLoadBalancer(rootCtx, log, legacyOpts, cluster, nodeSelector, userClusterClient, testPrefix, attemp)
						Expect(err).NotTo(HaveOccurred(), "LoadBalancer test failed after multiple attempts")
					})
					By(fmt.Sprintf("Smoke tests done %q", name))
				})
			}
		})
	}
})
