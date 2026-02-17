package utils

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/zap"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/clients"
	k8cginkgo "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/cluster"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/options"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	"k8s.io/kubectl/pkg/util/slice"
)

func PrepareSuite(
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
		err := runtimeOpts.SeedClusterClient.Update(rootCtx, seed)
		Expect(err).NotTo(HaveOccurred(), "Failed to update seed")
	})

	suiteCfg, reporterCfg := GinkgoConfiguration()
	By(fmt.Sprintf("parallel=%d", suiteCfg.ParallelTotal))
	By(fmt.Sprintf("Reporter: %#v", reporterCfg))
	By(fmt.Sprintf("Creating clusters for datacenters and kube versions: %v", maps.Keys(newClusters)))

	var wg sync.WaitGroup
	maxConcurrent := 4 // Set your desired concurrency limit
	sem := make(chan struct{}, maxConcurrent)
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
	for i, _ := range defaultSeedSettings {
		log.Infof("defaultSeedSettings[%d]: %+v", i, datacenterNameMappings[i])
	}
	for i, _ := range newClusters {
		log.Infof("newClusters[%d]: %+v", i, finalClusterDescriptions[i])
	}
	for seedKey := range defaultSeedSettings {
		exclude := false
		if len(opts.Included.DatacenterDescriptions) > 0 {
			exclude = true
			for _, included := range opts.Included.DatacenterDescriptions {
				if strings.Contains(seedKey, included) {
					exclude = false
					break
				}
			}
		} else {
			for _, excluded := range opts.Excluded.DatacenterDescriptions {
				if strings.Contains(seedKey, excluded) {
					exclude = true
					break
				}
			}
		}
		if exclude {
			continue
		}
		clustersToBuild := map[string]*kubermaticv1.ClusterSpec{}
		for name, clusterSpec := range newClusters {
			if !slice.ContainsString(versionSlice, clusterSpec.Version.String(), nil) {
				continue
			}
			clusterDesc, ok := finalClusterDescriptions[name]
			if !ok {
				continue
			}
			exclude = false
			if len(opts.Included.ClusterDescriptions) > 0 {
				exclude = true
				for _, included := range opts.Included.ClusterDescriptions {
					if slice.ContainsString(clusterDesc, included, nil) {
						exclude = false
						break
					}
				}
			} else {
				for _, excluded := range opts.Excluded.ClusterDescriptions {
					if slice.ContainsString(clusterDesc, excluded, nil) {
						exclude = true
						break
					}
				}
			}
			if exclude {
				continue
			}
			clustersToBuild[name] = clusterSpec
		}

		for name, clusterSpec := range clustersToBuild {
			log.Infof("Preparing creation of cluster %s for datacenter %s", name, clusterSpec.Cloud.DatacenterName)
			sem <- struct{}{} // acquire a slot
			wg.Add(1)
			go func(name string, project string, spec *kubermaticv1.ClusterSpec) {
				defer wg.Done()
				defer func() { <-sem }() // release the slot
				if !skipClusterCreation {
					cluster.Ensure(rootCtx, log, name, spec.Cloud.DatacenterName, project, spec, &legacyOpts, runtimeOpts, &opts)
				}
				if skipClusterCreation && updateClusters {
					cluster.Update(name, spec)
				}
			}(name, legacyOpts.KubermaticProject, clusterSpec)
		}
	}
	wg.Wait()
}
