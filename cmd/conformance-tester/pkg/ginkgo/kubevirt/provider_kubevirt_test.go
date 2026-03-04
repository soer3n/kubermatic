package kubevirt

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/gomega"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/tests"
)

var _ = ReportAfterEach(func(f SpecContext, r SpecReport) {
	By("Report after smoke tests")
})

var _ = Describe("KubeVirt", func() {
	// var entry []build.Scenario
	for name, entry := range GetTableEntries(rootCtx, log, runtimeOpts, legacyOpts, opts, infraClient, projectName) {
		// name := name
		// entry := entry
		Describe(name, func() {
			BeforeEach(func() {
				// cluster.Ensure(rootCtx, log, name, entry[0].ClusterSpec.Cloud.DatacenterName, projectName, entry[0].ClusterSpec, legacyOpts, runtimeOpts, opts)
			})

			for _, v := range entry {
				label := Label("kubevirt")
				if v.Exclude {
					label = Label("skip")
				}
				clusterLabel := Label(fmt.Sprintf("cluster-%s", v.ScenarioName))
				It(v.Description, label, clusterLabel, func() {
					cluster := &kubermaticv1.Cluster{}
					if err := runtimeOpts.SeedClusterClient.Get(rootCtx, types.NamespacedName{Name: name}, cluster); err != nil {
						log.Errorf("Failed to get cluster %s: %v", name, err)
						Fail(fmt.Sprintf("Failed to get cluster %s: %v", name, err))
					}

					userClusterClient, err := runtimeOpts.ClusterClientProvider.GetClient(rootCtx, cluster)
					if err != nil {
						log.Errorf("Failed to get user cluster client for cluster %s: %v", name, err)
						Fail(fmt.Sprintf("Failed to get user cluster client for cluster %s: %v", name, err))
					}
					By(fmt.Sprintf("Running tests for datacenter %q kubeVersion %q", v.ClusterSpec.Cloud.DatacenterName, v.ClusterSpec.Version.String()))
					By(fmt.Sprintf("Scenario with dc %q cluster %q", v.ClusterSpec.Cloud.DatacenterName, name))
					By(fmt.Sprintf("Setting up machine for %s %s. Scenario: %s", v.ClusterSpec.Cloud.DatacenterName, name, v.Description), func() {
						k8cginkgo.MachineSetup(rootCtx, log, userClusterClient, name, v.ScenarioName, &v.Machine, legacyOpts)
					})

					By(fmt.Sprintf("Machine setup done %q", name))
					By(fmt.Sprintf("Running smoke tests %q (enabled: %v) (%v)", name, legacyOpts.EnableTests, opts.EnableTests), func() {
						n := 0
						ExpectWithOffset(3, tests.TestStorage(rootCtx, log, legacyOpts, cluster, map[string]string{
							k8cginkgo.MachineNameLabel: fmt.Sprintf("machine-%s", name),
						}, userClusterClient, "", n+1)).To(BeNil())
						n = 0
						ExpectWithOffset(3, tests.TestLoadBalancer(rootCtx, log, legacyOpts, cluster, map[string]string{
							k8cginkgo.MachineNameLabel: fmt.Sprintf("machine-%s", name),
						}, userClusterClient, "", n+1)).To(BeNil())
					})
					By(fmt.Sprintf("Smoke tests done %q", name))
					time.Sleep(500 * time.Millisecond)
				})
			}
		})
	}
})
