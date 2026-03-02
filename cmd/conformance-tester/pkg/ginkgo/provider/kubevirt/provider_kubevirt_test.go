package kubevirt

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/gomega"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/build"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/utils"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/tests"
	"k8c.io/machine-controller/sdk/providerconfig"
)

var _ = ReportAfterEach(func(f SpecContext, r SpecReport) {
	By("Report after smoke tests")
})

var _ = Describe("KubeVirt", func() {
	var includedMachineDescriptions, excludedMachineDescriptions map[string]map[string][]string
	entries, includedMachineDescriptions, excludedMachineDescriptions, datacenterNameMappings, seed = build.GetTableEntries(rootCtx, log, runtimeOpts, legacyOpts, opts, infraClient, projectName, providerconfig.CloudProviderKubeVirt)
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

				It(fmt.Sprintf("%s and %v", entry.Description, strings.Join(machineDescription, " and ")), label, clusterLabel, func() {
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
					By(fmt.Sprintf("Setting up machine for %s %s. Scenario: %s", entry.ClusterSpec.Cloud.DatacenterName, name, entry.Description), func() {
						utils.MachineSetup(rootCtx, log, userClusterClient, name, k[:8], &v, legacyOpts)
					})

					By(fmt.Sprintf("Machine setup done %q", name))
					By(fmt.Sprintf("Running smoke tests %q (enabled: %v) (%v)", name, legacyOpts.EnableTests, opts.EnableTests), func() {

						Eventually(func() bool {
							for i := range 3 {
								err := tests.TestStorage(rootCtx, log, legacyOpts, cluster, map[string]string{
									utils.MachineNameLabel: fmt.Sprintf("machine-%s", name),
								}, userClusterClient, i+1)
								if err == nil {
									return true
								}
							}
							return false
						}).WithTimeout(5 * time.Minute).To(BeTrue())
						Eventually(func() bool {
							for i := range 3 {
								err := tests.TestLoadBalancer(rootCtx, log, legacyOpts, cluster, map[string]string{
									utils.MachineNameLabel: fmt.Sprintf("machine-%s", name),
								}, userClusterClient, i+1)
								if err == nil {
									return true
								}
							}
							return false
						}).WithTimeout(5 * time.Minute).To(BeTrue())
					})
					By(fmt.Sprintf("Smoke tests done %q", name))
				})
			}
		})
	}
})
