package utils

import (
	"context"
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/zap"
	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/clients"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/build"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/options"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	apitypes "k8s.io/apimachinery/pkg/types"
)

func PostProcessingSuite(
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
	skipClusterDeletion bool,
) {
	// primary node: idempotent deletion attempt for each cluster
	By(fmt.Sprintf("Deleting created clusters for e2e project %q", legacyOpts.KubermaticProject))
	var wg sync.WaitGroup

	deleteCtx := context.Background()
	if !skipClusterDeletion {
		for name, entry := range entries {
			wg.Go(func() {
				if entry.Exclude {
					By("Skipping deletion for excluded cluster " + name)
					return
				}

				cluster := &kubermaticv1.Cluster{}
				err := legacyOpts.SeedClusterClient.Get(deleteCtx, apitypes.NamespacedName{Name: name}, cluster)
				if err != nil {
					log.Errorf("Failed to get cluster '%s' for cleanup: %v", name, err)
					return
				}

				By(fmt.Sprintf("Cleaning up resources for cluster %s.", name))
				CommonCleanup(deleteCtx, log, clients.NewKubeClient(&legacyOpts), nil, legacyOpts.SeedClusterClient, cluster, name)

			})
		}
		wg.Wait()

		By("Detaching datacenters from seed")
		log.Infof("Removing datacenters with hashed names %v from seed 'kubermatic'", datacenterNameMappings)
		seed := &kubermaticv1.Seed{}
		err := legacyOpts.SeedClusterClient.Get(deleteCtx, apitypes.NamespacedName{Name: legacyOpts.KubermaticSeedName, Namespace: legacyOpts.KubermaticNamespace}, seed)
		if err != nil {
			log.Errorf("Failed to get seed 'kubermatic' for cleanup: %v", err)
		} else {
			for _, hashedName := range datacenterNameMappings {
				log.Infof("Removing datacenter with hashed name %q from seed 'kubermatic'", hashedName)
				delete(seed.Spec.Datacenters, hashedName)
			}

			if err := legacyOpts.SeedClusterClient.Update(deleteCtx, seed); err != nil {
				log.Errorf("Failed to update seed 'kubermatic' to remove datacenters: %v", err)
			}
		}

		var client clients.Client
		By(KKP("Creating a KKP client"), func() {
			client = clients.NewKubeClient(&legacyOpts)
			Expect(client.Setup(context.Background(), log)).To(Succeed())
		})

		By(KKP("Ensuring a project exists"), func() {
			if legacyOpts.KubermaticProject == "" {
				err := client.DeleteProject(context.Background(), log, projectName, time.Minute*2)
				Expect(err).NotTo(HaveOccurred())
			}
			fmt.Fprintf(GinkgoWriter, "Using project %q\n", legacyOpts.KubermaticProject)
		})
	}

}
