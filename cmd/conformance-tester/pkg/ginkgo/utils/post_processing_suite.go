package utils

import (
	"context"
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo/v2"

	"go.uber.org/zap"
	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/clients"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/options"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/util/slice"
)

func PostProcessingSuite(
	rootCtx context.Context,
	log *zap.SugaredLogger,
	legacyOpts legacytypes.Options,
	runtimeOpts *options.RuntimeOptions,
	opts options.Options,
	seed *kubermaticv1.Seed,
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

	if !skipClusterDeletion {
		for seedKey := range defaultSeedSettings {
			for dc := range newClusters {
				wg.Add(1)
				go func(dc string) {
					defer wg.Done()
					cluster := &kubermaticv1.Cluster{}

					if !slice.ContainsString(versionSlice, cluster.Spec.Version.String(), nil) {
						return
					}
					err := runtimeOpts.SeedClusterClient.Get(rootCtx, apitypes.NamespacedName{Name: fmt.Sprintf("%s-%s", dc, seedKey)}, cluster)
					if err != nil {
						log.Errorf("Failed to get cluster %s: %v", dc, err)
						return
					}
					By("Deleting cluster for " + dc)

					userClusterClient, err := runtimeOpts.ClusterClientProvider.GetClient(rootCtx, cluster)
					if err != nil {
						log.Errorf("Failed to get user cluster client for cluster %s: %v", dc, err)
						return
					}
					By(fmt.Sprintf("Cleaning up resources for cluster %s. Name is %s", dc, cluster.Name))
					CommonCleanup(rootCtx, log, clients.NewKubeClient(&legacyOpts), nil, userClusterClient, cluster)

				}(dc)
			}
		}
		wg.Wait()

		By("Detaching datacenters from seed")
		seed := &kubermaticv1.Seed{}
		err := runtimeOpts.SeedClusterClient.Get(rootCtx, apitypes.NamespacedName{Name: "kubermatic", Namespace: "kubermatic"}, seed)
		if err != nil {
			log.Errorf("Failed to get seed 'kubermatic' for cleanup: %v", err)
		} else {
			for _, hashedName := range datacenterNameMappings {
				delete(seed.Spec.Datacenters, hashedName)
			}

			if err := runtimeOpts.SeedClusterClient.Update(rootCtx, seed); err != nil {
				log.Errorf("Failed to update seed 'kubermatic' to remove datacenters: %v", err)
			}
		}
	}

}
