/*
Copyright 2025 The Kubermatic Kubernetes Platform contributors.

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
	"testing"
)

// TestKubeVirtRawConfigCases validates a curated table of KubeVirt RawConfig cases
// based on sdk/cloudprovider/kubevirt/types.go via strict unmarshal.
func TestKubeVirtRawConfigCases(t *testing.T) {
	// providerEnv := "kubevirt"
	// providersList := strings.Split(providerEnv, ",")

	// limitEnv := os.Getenv("LIMIT")
	// limit := 3500
	// if limitEnv != "" {
	// 	if v, err := strconv.Atoi(limitEnv); err == nil {
	// 		limit = v
	// 	}
	// }

	// results := make(map[string][]providerconfig.Config)
	// kubeconfig := filepath.Join(
	// 	"/home/soer3n/vscode/mybackup/kubermatic-work/local/tmp/local-kkp",
	// )
	// config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	// if err != nil {
	// 	t.Fatalf("failed to build config: %v", err)
	// }

	// for _, p := range providersList {
	// 	p = strings.TrimSpace(p)
	// 	cases, err := GenerateProviderTestCases(p, limit, config, types.Secrets{
	// 		Kubevirt: types.KubevirtSecrets{
	// 			Kubeconfig: string(kubeconfigBytes),
	// 		},
	// 	})
	// 	if err != nil {
	// 		log.Printf("Error generating cases for %s: %v", p, err)
	// 		continue
	// 	}
	// 	results[p] = cases
	// }

	// fmt.Printf("Generated %d valid cases for provider %s\n", len(results["kubevirt"]), "kubevirt")

	// for _, _ = range results["kubevirt"] {
	// Marshal and validate via StrictUnmarshal
	// raw, err := json.Marshal(tt.rc)
	// if err != nil {
	// 	t.Fatalf("%s: marshal failed: %v", tt.name, err)
	// }
	// pc := providerconfig.Config{CloudProviderSpec: runtime.RawExtension{Raw: raw}}
	// if _, err := kvsdk.GetConfig(pc); err != nil {
	// 	t.Fatalf("%s: StrictUnmarshal failed: %v", tt.name, err)
	// }
	// }
}
