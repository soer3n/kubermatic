package utils

import (
	"context"
	"fmt"
	"maps"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/zap"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/clients"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/build"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/cluster"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/options"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
)

func PrepareSuite(
	rootCtx context.Context,
	log *zap.SugaredLogger,
	legacyOpts legacytypes.Options,
	runtimeOpts *options.RuntimeOptions,
	opts options.Options,
	seed *kubermaticv1.Seed,
	entries map[string]*build.Scenario,
	kkpConfig *kubermaticv1.KubermaticConfiguration,
	projectName string,
	defaultSeedSettings map[string]kubermaticv1.Seed,
	newClusters map[string]*kubermaticv1.ClusterSpec,
	finalClusterDescriptions map[string][]string,
	datacenterNameMappings map[string]string,
	skipClusterCreation bool,
	updateClusters bool,
) {
	var client clients.Client
	By(k8cginkgo.KKP("Creating a KKP client"), func() {
		client = clients.NewKubeClient(&legacyOpts)
		Expect(client.Setup(rootCtx, log)).To(Succeed())
	})

	By(k8cginkgo.KKP("Ensuring a project exists"), func() {
		if legacyOpts.KubermaticProject == "" {
			p, err := client.CreateProject(rootCtx, log, projectName)
			Expect(err).NotTo(HaveOccurred())
			projectName = p
			legacyOpts.KubermaticProject = projectName
			opts.KubermaticProject = projectName
		}
		fmt.Fprintf(GinkgoWriter, "Using project %q\n", legacyOpts.KubermaticProject)
	})

	By(k8cginkgo.KKP("Ensuring SSH keys exist"), func() {
		Expect(client.EnsureSSHKeys(rootCtx, log)).To(Succeed())
	})

	By(k8cginkgo.KKP("Attaching datacenters to seed"), func() {
		err := legacyOpts.SeedClusterClient.Update(rootCtx, seed)
		Expect(err).NotTo(HaveOccurred(), "Failed to update seed")
	})

	suiteCfg, reporterCfg := GinkgoConfiguration()
	By(fmt.Sprintf("parallel=%d", suiteCfg.ParallelTotal))
	By(fmt.Sprintf("Reporter: %#v", reporterCfg))
	By(fmt.Sprintf("Creating/updating clusters for datacenters and kube versions: %v", maps.Keys(newClusters)))

	var wg sync.WaitGroup
	maxConcurrent := 4 // Set your desired concurrency limit
	sem := make(chan struct{}, maxConcurrent)

	for name, entry := range entries {
		if entry.Exclude {
			continue
		}
		clusterSpec := entry.ClusterSpec
		log.Infof("Preparing creation of cluster %s for datacenter %s", name, clusterSpec.Cloud.DatacenterName)
		sem <- struct{}{} // acquire a slot
		wg.Add(1)
		go func(name string, project string, spec *kubermaticv1.ClusterSpec) {
			defer wg.Done()
			defer func() { <-sem }() // release the slot
			if !skipClusterCreation {
				cluster.Ensure(rootCtx, log, name, spec.Cloud.DatacenterName, project, spec, &legacyOpts, runtimeOpts, &opts, client)
			}
			if skipClusterCreation && updateClusters {
				cluster.Update(rootCtx, log, name, spec.Cloud.DatacenterName, project, spec, &legacyOpts, runtimeOpts, &opts, client)
			}
		}(name, legacyOpts.KubermaticProject, clusterSpec)
	}
	wg.Wait()
	log.Infof("Finished preparing creation of clusters")
}
