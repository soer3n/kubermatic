package ginkgo

import (
	"iter"
	"maps"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	// . "github.com/onsi/gomega"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/clients"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/scenarios"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/tests"
)

// The ReportAfterEach function has been removed in favor of using AfterEach
// with CurrentSpecReport(). This is a Ginkgo best practice as it allows the
// reporting logic to live alongside the test and have access to its closure
// (e.g., the `cluster` variable) for more detailed, context-aware reporting on failure.

// getAllProviders dynamically discovers all supported cloud providers by inspecting
// the fields of the `DatacenterSpec` struct from the Kubermatic API. This approach
// ensures that the test suite automatically adapts to include new providers as they
// are added to the API, reducing manual maintenance.

// Following Ginkgo best practices, we iterate over the providers to create a `Describe`
// block for each, rather than using a custom `DescribeTableSubtree` helper. This improves
// readability and aligns with standard Ginkgo patterns.
var runTestFunc = func(k iter.Seq[string], v string) bool {
	for item := range k {
		if strings.Contains(v, item) {
			return true
		}
	}
	return false
}

var _ = Describe("[provider]", func() {
	// Dynamically discover all providers from the DatacenterSpec struct. This makes
	// the test suite more robust and easier to maintain.
	allProviders := getAllProviders()

	for description, provider := range allProviders {
		// Capture range variables to ensure they have the correct value in the closure.
		provider := provider
		description := description
		settingsList := GetSettingsForProvider(provider)
		settingsList, _ = mergeTestSettings(settingsList, legacyOpts.TestSettings)

		// Expect(err).NotTo(HaveOccurred())
		// Assume `opts.TestSettings` is a comma-separated string of descriptions from a flag.
		// If the flag is empty, all tests will run.
		var enabledSettings map[string]struct{}
		if len(legacyOpts.TestSettings) > 0 {
			enabledSettings = make(map[string]struct{})
			for s, _ := range legacyOpts.TestSettings {
				enabledSettings[strings.TrimSpace(s)] = struct{}{}
			}
		}

		Describe(description, func() {

			for _, settings := range settingsList {
				settings := settings // capture range variable

				// Determine if this setting should be run
				runTest := true
				if enabledSettings != nil {
					// _, runTest = enabledSettings[settings.Description]
					runTest = runTestFunc(maps.Keys(enabledSettings), settings.Description)
					log.Infof("Considering test setting %q for provider %q: runTest=%v", settings.Description, description, runTest)
				}

				Context("using user story", func() {
					// A `DescribeTable` is used here for the scenarios, which is the idiomatic
					// way to run the same test logic against multiple data-driven inputs.
					DescribeTableSubtree(settings.Description,
						func(scenario scenarios.Scenario) {
							// The function passed to DescribeTable serves as the complete test body for each entry.
							// Unlike Describe/Context blocks, it's not a container for other nodes like
							// `BeforeEach`, `AfterEach`, or `It`. Therefore, setup is performed directly
							// at the beginning of the function, and teardown is managed with a `defer`
							// statement to ensure it executes reliably after the test logic.

							var (
								cluster           *kubermaticv1.Cluster
								userClusterClient ctrlruntimeclient.Client
							)

							BeforeEach(func() {
								if !runTest {
									Skip("This test setting was not selected to run via the --test-settings flag.")
								}
								cluster, userClusterClient = commonSetup(rootCtx, log, scenario, legacyOpts)
							})

							AfterEach(func() {
								if runTest {
									commonCleanup(rootCtx, log, client, scenario, userClusterClient, cluster)
									currentSpecReport := CurrentSpecReport()
									if currentSpecReport.Failed() {
										By("Capturing diagnostics for failed test")
										// e.g., AddReportEntry("Cluster Events", captureClusterEvents(cluster))
									}
									r := NewJUnitReporter(opts.ReportsRoot)
									r.AfterSpec(currentSpecReport)
								}
							})

							It("should succeed", func() {
								// This is the actual test logic.
								machineSetup(rootCtx, log, clients.NewKubeClient(legacyOpts), scenario, userClusterClient, cluster, legacyOpts)

								var failures []string
								// Individual smoke tests are wrapped in `By` to clearly delineate them in the test report.
								By(KKP(CloudProvider("Test PersistentVolumes")), func() {
									if err := tests.TestStorage(rootCtx, log, legacyOpts, cluster, userClusterClient, 1); err != nil {
										failures = append(failures, "PersistentVolumes test failed: "+err.Error())
									}
								})

								By(KKP(CloudProvider("Test LoadBalancers")), func() {
									if err := tests.TestLoadBalancer(rootCtx, log, legacyOpts, cluster, userClusterClient, 1); err != nil {
										failures = append(failures, "LoadBalancers test failed: "+err.Error())
									}
								})

								By(KKP("Test user cluster RBAC controller"), func() {
									if err := tests.TestUserclusterControllerRBAC(rootCtx, log, legacyOpts, cluster, userClusterClient, legacyOpts.SeedClusterClient); err != nil {
										failures = append(failures, "User cluster RBAC controller test failed: "+err.Error())
									}
								})

								By(KKP("Test prometheus metrics availability"), func() {
									if err := tests.TestUserClusterMetrics(rootCtx, log, legacyOpts, cluster, legacyOpts.SeedClusterClient); err != nil {
										failures = append(failures, "Prometheus metrics availability test failed: "+err.Error())
									}
								})

								By(KKP("Test pod and node metrics availability"), func() {
									if err := tests.TestUserClusterPodAndNodeMetrics(rootCtx, log, legacyOpts, cluster, userClusterClient); err != nil {
										failures = append(failures, "Pod and node metrics availability test failed: "+err.Error())
									}
								})

								By(KKP("Test pod seccomp profiles on user cluster"), func() {
									if err := tests.TestUserClusterSeccompProfiles(rootCtx, log, legacyOpts, cluster, userClusterClient); err != nil {
										failures = append(failures, "Pod seccomp profiles on user cluster test failed: "+err.Error())
									}
								})

								By(KKP("Test container images not containing k8s.gcr.io on user cluster"), func() {
									if err := tests.TestUserClusterNoK8sGcrImages(rootCtx, log, legacyOpts, cluster, userClusterClient); err != nil {
										failures = append(failures, "Container images on user cluster test failed: "+err.Error())
									}
								})

								By(KKP("Test pod security context on seed cluster"), func() {
									if err := tests.TestUserClusterControlPlaneSecurityContext(rootCtx, log, legacyOpts, cluster); err != nil {
										failures = append(failures, "Pod security context on seed cluster test failed: "+err.Error())
									}
								})

								By(KKP("Test container images not containing k8s.gcr.io on seed cluster"), func() {
									if err := tests.TestNoK8sGcrImages(rootCtx, log, legacyOpts, cluster); err != nil {
										failures = append(failures, "Container images on seed cluster test failed: "+err.Error())
									}
								})

								By(KKP("Test telemetry"), func() {
									if err := tests.TestTelemetry(rootCtx, log, legacyOpts); err != nil {
										failures = append(failures, "Telemetry test failed: "+err.Error())
									}
								})

								if len(failures) > 0 {
									Fail(strings.Join(failures, "\n"))
								}
							})
						},
						// Scenarios are generated dynamically for the current provider in the loop.
						scenarioEntriesByProvider(testSuiteScenarios, provider))
				})
			}
		})
	}
})
