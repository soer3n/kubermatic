/*
Copyright 2022 The Kubermatic Kubernetes Platform contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scenarios

import (
	"context"
	"fmt"
	"log"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	"k8c.io/machine-controller/sdk/providerconfig"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/ptr"
)

const (
	kubevirtImageHTTPServerSvc = "http://image-repo.kube-system.svc/images"
	kubevirtVCPUs              = 2
	kubevirtMemory             = "4Gi"
	kubevirtDiskSize           = "25Gi"
	kubevirtStorageClassName   = "standard"
)

type kubevirtScenario struct {
	baseScenario
}

func (s *kubevirtScenario) compatibleOperatingSystems() sets.Set[providerconfig.OperatingSystem] {
	return sets.New(
		providerconfig.OperatingSystemUbuntu,
		providerconfig.OperatingSystemRHEL,
		providerconfig.OperatingSystemFlatcar,
		providerconfig.OperatingSystemRockyLinux,
	)
}

func (s *kubevirtScenario) IsValid() error {
	if err := s.baseScenario.IsValid(); err != nil {
		return err
	}

	if compat := s.compatibleOperatingSystems(); !compat.Has(s.operatingSystem) {
		return fmt.Errorf("provider supports only %v", sets.List(compat))
	}

	return nil
}

func (s *kubevirtScenario) Cluster(secrets types.Secrets) *kubermaticv1.ClusterSpec {
	return &kubermaticv1.ClusterSpec{
		Cloud: kubermaticv1.CloudSpec{
			DatacenterName: secrets.Kubevirt.KKPDatacenter,
			Kubevirt: &kubermaticv1.KubevirtCloudSpec{
				Kubeconfig: secrets.Kubevirt.Kubeconfig,
				StorageClasses: []kubermaticv1.KubeVirtInfraStorageClass{{
					Name:           kubevirtStorageClassName,
					IsDefaultClass: ptr.To(true),
				}},
			},
		},
		Version: s.clusterVersion,
	}
}

func (s *kubevirtScenario) MachineDeployments(_ context.Context, num int, secrets types.Secrets, cluster *kubermaticv1.Cluster, sshPubKeys []string) ([]clusterv1alpha1.MachineDeployment, error) {
	// image, err := s.getOSImage()
	// if err != nil {
	// 	return nil, err
	// }

	// baseCloudProviderSpec := provider.NewKubevirtConfig().
	// 	WithVCPUs(kubevirtVCPUs).
	// 	WithMemory(kubevirtMemory).
	// 	WithPrimaryDiskOSImage(image).
	// 	WithPrimaryDiskSize(kubevirtDiskSize).
	// 	WithPrimaryDiskStorageClassName(kubevirtStorageClassName).
	// 	Build()

	// log.Printf("Using base KubeVirt cloud provider spec: %v", baseCloudProviderSpec)

	config, err := clientcmd.NewClientConfigFromBytes([]byte(secrets.Kubevirt.Kubeconfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create client config from kubevirt kubeconfig: %w", err)
	}

	results := make(map[string][]runtime.RawExtension)
	restConfig, _ := config.ClientConfig()
	limit := 10

	cases, err := GenerateProviderTestCases("kubevirt", limit, restConfig, secrets)
	if err != nil {
		log.Printf("Error generating cases for %s: %v", "kubevirt", err)
		return nil, fmt.Errorf("failed to generate test cases for kubevirt: %w", err)
	}
	results["kubevirt"] = cases
	mds := make([]clusterv1alpha1.MachineDeployment, 0, len(results["kubevirt"]))
	for _, providerConfig := range results["kubevirt"] {
		md, err := s.createMachineDeployment(cluster, num, providerConfig, sshPubKeys, secrets)
		if err != nil {
			return nil, err
		}
		mds = append(mds, md)
	}

	return mds, nil
}

func (s *kubevirtScenario) getOSImage() (string, error) {
	switch s.operatingSystem {
	case providerconfig.OperatingSystemUbuntu:
		return "docker://quay.io/kubermatic-virt-disks/ubuntu:22.04", nil
	default:
		return "", fmt.Errorf("unsupported OS %q selected", s.operatingSystem)
	}
}

func (s *kubevirtScenario) DatacenterMatrix() ([]*kubermaticv1.Datacenter, error) {
	datacenters := []*kubermaticv1.Datacenter{{
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				EnableDefaultNetworkPolicies: ptr.To(false),
				DNSPolicy:                    "ClusterFirst",
				InfraStorageClasses: []kubermaticv1.KubeVirtInfraStorageClass{{
					IsDefaultClass: ptr.To(true),
					Name:           kubevirtStorageClassName,
				}},
				Images: kubermaticv1.KubeVirtImageSources{
					HTTP: &kubermaticv1.KubeVirtHTTPSource{},
				},
			},
		},
	}, {
		Spec: kubermaticv1.DatacenterSpec{
			Kubevirt: &kubermaticv1.DatacenterSpecKubevirt{
				EnableDefaultNetworkPolicies: ptr.To(true),
				DNSPolicy:                    "ClusterFirst",
				InfraStorageClasses: []kubermaticv1.KubeVirtInfraStorageClass{{
					IsDefaultClass: ptr.To(true),
					Name:           kubevirtStorageClassName,
				}},
				Images: kubermaticv1.KubeVirtImageSources{
					HTTP: &kubermaticv1.KubeVirtHTTPSource{
						OperatingSystems: map[providerconfig.OperatingSystem]kubermaticv1.OSVersions{
							providerconfig.OperatingSystemUbuntu: {
								"22.04": "docker://quay.io/kubermatic-virt-disks/ubuntu:22.04",
							},
						},
					},
				},
			},
		},
	}}

	return datacenters, nil
}
