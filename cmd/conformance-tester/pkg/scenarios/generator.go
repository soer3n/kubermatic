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
	"math/rand"
	"reflect"
	"time"

	"go.uber.org/zap"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/sdk/v2/semver"
	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
	"k8c.io/machine-controller/sdk/providerconfig"

	"k8s.io/apimachinery/pkg/util/sets"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	b64 "encoding/base64"
	"encoding/json"

	kvprov "k8c.io/machine-controller/pkg/cloudprovider/provider/kubevirt"
	ctypes "k8c.io/machine-controller/pkg/cloudprovider/types"
	"k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	"k8c.io/machine-controller/sdk/cloudprovider/aws"
	"k8c.io/machine-controller/sdk/cloudprovider/azure"
	kvsdk "k8c.io/machine-controller/sdk/cloudprovider/kubevirt"
	"k8c.io/machine-controller/sdk/cloudprovider/openstack"
	"k8c.io/machine-controller/sdk/providerconfig/configvar"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var candidateValues = map[string]CandidateValues{
	"aws": {
		"Region":       []string{"us-east-1", "eu-central-1"},
		"InstanceType": []string{"t3.medium", "m5.large"},
		"NetworkConfig": CandidateValues{
			"SubnetID":        []string{"subnet-aaa", "subnet-bbb"},
			"SecurityGroupID": []string{"sg-1111", "sg-2222"},
		},
	},
	"azure": {
		"Location":      []string{"eastus", "westeurope"},
		"VMSize":        []string{"Standard_B2s", "Standard_D2s_v3"},
		"ResourceGroup": []string{"rg-test", "rg-prod"},
	},
	"openstack": {
		"Region":     []string{"RegionOne", "RegionTwo"},
		"Flavor":     []string{"m1.small", "m1.medium"},
		"Image":      []string{"ubuntu-20.04", "centos-8"},
		"Network":    []string{"public", "private"},
		"FloatingIP": []string{"true", "false"},
	},
	"kubevirt": {
		"ClusterName": []string{"cluster-a", "cluster-b"},
		"VirtualMachine": CandidateValues{
			// Use Flavor as deprecated-but-supported knob (optional)
			"Flavor": CandidateValues{
				"Name":    []string{"", "small"},
				"Profile": []string{"", "base"},
			},
			"Template": CandidateValues{
				"CPUs":   []string{"2", "4"},
				"Memory": []string{"2Gi", "4Gi"},
				"PrimaryDisk": CandidateValues{
					"Size":              []string{"10Gi", "20Gi"},
					"StorageClassName":  []string{"longhorn", "local-path"},
					"StorageAccessType": []string{"ReadWriteMany"},
					"StorageTarget":     []string{"pvc"},
					// "Template": CandidateValues{
					"CPUs":   []string{"2", "4"},
					"Memory": []string{"2Gi", "4Gi"},
					"PrimaryDisk": CandidateValues{
						"Size":              []string{"10Gi", "20Gi"},
						"StorageClassName":  []string{"longhorn", "local-path"},
						"StorageAccessType": []string{"ReadWriteMany"},
						"StorageTarget":     []string{"pvc"},
						"Source":            []string{"http", "registry", "pvc"},
						"OsImage":           []string{"http://example/vm.img", "docker://repo/image:tag", "ns/dvname"},
						"PullMethod":        []string{"node", "pod"},
					},
				}, "Source": []string{"http", "registry", "pvc"},
				"OsImage":    []string{"http://example/vm.img", "docker://repo/image:tag", "ns/dvname"},
				"PullMethod": []string{"node", "pod"},
			},
		},
	},
}

type Generator struct {
	cloudProviders   sets.Set[string]
	operatingSystems sets.Set[string]
	versions         sets.Set[string]
	enableDualstack  bool
}

func NewGenerator() *Generator {
	return &Generator{
		cloudProviders:   sets.New[string](),
		operatingSystems: sets.New[string](),
		versions:         sets.New[string](),
	}
}

func (g *Generator) WithCloudProviders(providerNames ...string) *Generator {
	for _, provider := range providerNames {
		g.cloudProviders.Insert(provider)
	}
	return g
}

func (g *Generator) WithOperatingSystems(operatingSystems ...string) *Generator {
	for _, os := range operatingSystems {
		g.operatingSystems.Insert(os)
	}
	return g
}

func (g *Generator) WithVersions(versions ...*semver.Semver) *Generator {
	for _, version := range versions {
		g.versions.Insert(version.String())
	}
	return g
}

func (g *Generator) WithDualstack(enable bool) *Generator {
	g.enableDualstack = enable
	return g
}

func (g *Generator) Scenarios(ctx context.Context, opts *types.Options, log *zap.SugaredLogger) (map[kubermaticv1.ProviderType][][]Scenario, error) {
	scenarios := make(map[kubermaticv1.ProviderType][][]Scenario)
	for _, providerName := range sets.List(g.cloudProviders) {
		datacenters, err := g.datacenterMatrix(ctx, opts, kubermaticv1.ProviderType(providerName))
		if err != nil {
			return nil, fmt.Errorf("failed to determine target datacenter for provider %q: %w", providerName, err)
		}
		scenarios[kubermaticv1.ProviderType(providerName)] = make([][]Scenario, len(datacenters))
		for _, version := range sets.List(g.versions) {
			s, err := semver.NewSemver(version)
			if err != nil {
				return nil, fmt.Errorf("invalid version %q: %w", version, err)
			}

			for _, operatingSystem := range sets.List(g.operatingSystems) {

				for i, datacenter := range datacenters {
					scenario, err := providerScenario(opts, kubermaticv1.ProviderType(providerName), providerconfig.OperatingSystem(operatingSystem), *s, datacenter)
					if err != nil {
						return nil, err
					}

					scenarios[kubermaticv1.ProviderType(providerName)][i] = append(scenarios[kubermaticv1.ProviderType(providerName)][i], scenario)
				}
			}
		}
	}

	return shuffleScenarios(scenarios), nil
}

func (g *Generator) datacenter(ctx context.Context, client ctrlruntimeclient.Client, secrets types.Secrets, provider kubermaticv1.ProviderType) (*kubermaticv1.Datacenter, error) {
	var datacenterName string

	switch provider {
	case kubermaticv1.AlibabaCloudProvider:
		datacenterName = secrets.Alibaba.KKPDatacenter
	case kubermaticv1.AnexiaCloudProvider:
		datacenterName = secrets.Anexia.KKPDatacenter
	case kubermaticv1.AWSCloudProvider:
		datacenterName = secrets.AWS.KKPDatacenter
	case kubermaticv1.AzureCloudProvider:
		datacenterName = secrets.Azure.KKPDatacenter
	case kubermaticv1.DigitaloceanCloudProvider:
		datacenterName = secrets.Digitalocean.KKPDatacenter
	case kubermaticv1.GCPCloudProvider:
		datacenterName = secrets.GCP.KKPDatacenter
	case kubermaticv1.HetznerCloudProvider:
		datacenterName = secrets.Hetzner.KKPDatacenter
	// case kubermaticv1.KubevirtCloudProvider:
	// 	datacenterName = secrets.Kubevirt.KKPDatacenter
	case kubermaticv1.NutanixCloudProvider:
		datacenterName = secrets.Nutanix.KKPDatacenter
	case kubermaticv1.OpenstackCloudProvider:
		datacenterName = secrets.OpenStack.KKPDatacenter
	case kubermaticv1.VMwareCloudDirectorCloudProvider:
		datacenterName = secrets.VMwareCloudDirector.KKPDatacenter
	case kubermaticv1.VSphereCloudProvider:
		datacenterName = secrets.VSphere.KKPDatacenter
	default:
		return nil, fmt.Errorf("cloud provider %q is not supported yet in conformance-tester", provider)
	}

	return getDatacenter(ctx, client, datacenterName)
}

func (g *Generator) datacenterMatrix(ctx context.Context, opts *types.Options, provider kubermaticv1.ProviderType) ([]*kubermaticv1.Datacenter, error) {
	base := baseScenario{
		cloudProvider: provider,
		// operatingSystem:  os,
		// clusterVersion:   version,
		dualstackEnabled: opts.DualStackEnabled,
	}
	switch provider {
	case kubermaticv1.KubevirtCloudProvider:
		kubevirtBaseScenario := &kubevirtScenario{baseScenario: base}
		dcs, err := kubevirtBaseScenario.DatacenterMatrix()
		if err != nil {
			return nil, err
		}
		return dcs, nil
	}
	dc, err := g.datacenter(ctx, opts.SeedClusterClient, opts.Secrets, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get datacenter %q: %w", dc.Location, err)
	}
	return []*kubermaticv1.Datacenter{dc}, nil
}

func providerScenario(
	opts *types.Options,
	provider kubermaticv1.ProviderType,
	os providerconfig.OperatingSystem,
	version semver.Semver,
	datacenter *kubermaticv1.Datacenter,
) (Scenario, error) {
	base := baseScenario{
		cloudProvider:    provider,
		operatingSystem:  os,
		clusterVersion:   version,
		datacenter:       datacenter,
		dualstackEnabled: opts.DualStackEnabled,
	}

	switch provider {
	case kubermaticv1.AlibabaCloudProvider:
		return &alibabaScenario{baseScenario: base}, nil
	case kubermaticv1.AnexiaCloudProvider:
		return &anexiaScenario{baseScenario: base}, nil
	case kubermaticv1.AWSCloudProvider:
		return &awsScenario{baseScenario: base}, nil
	case kubermaticv1.AzureCloudProvider:
		return &azureScenario{baseScenario: base}, nil
	case kubermaticv1.DigitaloceanCloudProvider:
		return &digitaloceanScenario{baseScenario: base}, nil
	case kubermaticv1.GCPCloudProvider:
		return &googleScenario{baseScenario: base}, nil
	case kubermaticv1.HetznerCloudProvider:
		return &hetznerScenario{baseScenario: base}, nil
	case kubermaticv1.KubevirtCloudProvider:
		return &kubevirtScenario{baseScenario: base}, nil
	case kubermaticv1.NutanixCloudProvider:
		return &nutanixScenario{baseScenario: base}, nil
	case kubermaticv1.OpenstackCloudProvider:
		return &openStackScenario{baseScenario: base}, nil
	case kubermaticv1.VMwareCloudDirectorCloudProvider:
		return &vmwareCloudDirectorScenario{baseScenario: base}, nil
	case kubermaticv1.VSphereCloudProvider:
		scenario := &vSphereScenario{baseScenario: base}
		scenario.customFolder = opts.ScenarioOptions.Has("custom-folder")
		scenario.basePath = opts.ScenarioOptions.Has("basepath")
		scenario.datastoreCluster = opts.ScenarioOptions.Has("datastore-cluster")

		if scenario.customFolder && scenario.basePath {
			return nil, fmt.Errorf("cannot run mutually exclusive %q scenarios 'custom-folder' and 'basepath' together", provider)
		}

		return scenario, nil
	default:
		return nil, fmt.Errorf("cloud provider %q is not supported yet in conformance-tester", provider)
	}
}

func shuffleScenarios(scenarios map[kubermaticv1.ProviderType][][]Scenario) map[kubermaticv1.ProviderType][][]Scenario {
	for provider, scenarioLists := range scenarios {
		for i, scenarioList := range scenarioLists {
			scenarios[provider][i] = shuffle(scenarioList)
		}
	}
	return scenarios
}

func shuffle(vals []Scenario) []Scenario {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ret := make([]Scenario, len(vals))
	n := len(vals)
	for i := range n {
		randIndex := r.Intn(len(vals))
		ret[i] = vals[randIndex]
		vals = append(vals[:randIndex], vals[randIndex+1:]...)
	}
	return ret
}

func getDatacenter(ctx context.Context, client ctrlruntimeclient.Client, datacenter string) (*kubermaticv1.Datacenter, error) {
	seeds := &kubermaticv1.SeedList{}
	if err := client.List(ctx, seeds); err != nil {
		return nil, fmt.Errorf("failed to list seeds: %w", err)
	}

	for _, seed := range seeds.Items {
		for name, dc := range seed.Spec.Datacenters {
			if name == datacenter {
				return &dc, nil
			}
		}
	}

	return nil, fmt.Errorf("no Seed contains datacenter %q", datacenter)
}

type CandidateValues map[string]interface{}
type TestCase map[string]interface{}

func flattenCandidates(prefix string, data CandidateValues, acc map[string][]string) {
	for key, val := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		switch v := val.(type) {
		case []string:
			acc[fullKey] = v
		case CandidateValues:
			flattenCandidates(fullKey, v, acc)
		default:
			panic(fmt.Sprintf("unsupported type for %s", fullKey))
		}
	}
}

func cartesian(dimensions [][]string) [][]string {
	result := [][]string{{}}
	for _, dim := range dimensions {
		var tmp [][]string
		for _, r := range result {
			for _, v := range dim {
				tmp = append(tmp, append(append([]string{}, r...), v))
			}
		}
		result = tmp
	}
	return result
}

func setNestedField(obj reflect.Value, path []string, value string) {
	if obj.Kind() == reflect.Ptr {
		if obj.IsNil() {
			obj.Set(reflect.New(obj.Type().Elem()))
		}
		obj = obj.Elem()
	}
	field := obj.FieldByName(path[0])
	if !field.IsValid() {
		return
	}
	if len(path) == 1 {
		if field.Kind() == reflect.Ptr {
			ptr := reflect.New(field.Type().Elem())
			vField := ptr.Elem().FieldByName("Value")
			if vField.IsValid() && vField.Kind() == reflect.String {
				vField.SetString(value)
			}
			field.Set(ptr)
		} else if field.Kind() == reflect.Struct {
			vf := field.FieldByName("Value")
			if vf.IsValid() && vf.Kind() == reflect.String {
				vf.SetString(value)
			}
		}
		return
	}
	setNestedField(field, path[1:], value)
}

func buildConfigStruct(provider string, values map[string]string) (interface{}, error) {
	switch provider {
	case "aws":
		cfg := aws.RawConfig{}
		for k, v := range values {
			setNestedField(reflect.ValueOf(&cfg), splitPath(k), v)
		}
		return cfg, nil
	case "azure":
		cfg := azure.RawConfig{}
		for k, v := range values {
			setNestedField(reflect.ValueOf(&cfg), splitPath(k), v)
		}
		return cfg, nil
	case "openstack":
		cfg := openstack.RawConfig{}
		for k, v := range values {
			setNestedField(reflect.ValueOf(&cfg), splitPath(k), v)
		}
		return cfg, nil
	case "kubevirt":
		cfg := kvsdk.RawConfig{}
		for k, v := range values {
			setNestedField(reflect.ValueOf(&cfg), splitPath(k), v)
		}
		return cfg, nil
	}
	return nil, fmt.Errorf("unsupported provider: %s", provider)
}

func splitPath(path string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '.' {
			parts = append(parts, path[start:i])
			start = i + 1
		}
	}
	parts = append(parts, path[start:])
	return parts
}

func validateConfig(providerName string, cfg interface{}, restConfig *rest.Config, secrets types.Secrets) (providerconfig.Config, error) {
	var cloudSpecRaw []byte
	var prov ctypes.Provider
	kubeClient, err := client.New(restConfig, client.Options{})
	if err != nil {
		panic(err)
	}

	switch providerName {
	case "aws":
		cfg, _ = cfg.(aws.RawConfig)
	case "azure":
		cfg, _ = cfg.(azure.RawConfig)
	case "openstack":
		cfg, _ = cfg.(openstack.RawConfig)
	case "kubevirt":
		kvConfig, _ := cfg.(kvsdk.RawConfig)
		config, err := clientcmd.NewClientConfigFromBytes([]byte(secrets.Kubevirt.Kubeconfig))
		if err != nil {
			return providerconfig.Config{}, fmt.Errorf("failed to create client config from kubevirt kubeconfig: %w", err)
		}
		restConfig, _ := config.ClientConfig()
		kubeconfigBytes, err := RestConfigToKubeconfigBytes(restConfig)
		if err != nil {
			return providerconfig.Config{}, fmt.Errorf("convert rest config to kubeconfig: %w", err)
		}
		// fmt.Println("Using kubeconfig file: ", kubeconfig)
		kvConfig.Auth.Kubeconfig.Value = string(b64.StdEncoding.EncodeToString(kubeconfigBytes))
		cfg = kvConfig
		// fmt.Printf("Using KubeVirt cloud provider spec: %v\n", kvConfig)
		cloudSpecRaw, err = json.Marshal(kvConfig)
		if err != nil {
			return providerconfig.Config{}, fmt.Errorf("marshal cloud spec: %w", err)
		}
		prov = kvprov.New(configvar.NewResolver(context.Background(), kubeClient))
	default:
		return providerconfig.Config{}, fmt.Errorf("unsupported provider: %s", providerName)
	}

	// JSON-encode provider-specific RawConfig and wrap into providerconfig.Config

	pc := providerconfig.Config{
		CloudProviderSpec: runtime.RawExtension{Raw: cloudSpecRaw},
		// Provide minimal OS to satisfy providers that require it
		OperatingSystem:     providerconfig.OperatingSystemUbuntu,
		OperatingSystemSpec: runtime.RawExtension{Raw: []byte("{}")},
	}
	pcBytes, err := json.Marshal(pc)
	if err != nil {
		return providerconfig.Config{}, fmt.Errorf("marshal providerconfig: %w", err)
	}

	return pc, prov.Validate(
		context.Background(),
		zap.NewNop().Sugar(),
		v1alpha1.MachineSpec{ProviderSpec: v1alpha1.ProviderSpec{Value: &runtime.RawExtension{Raw: pcBytes}}},
	)
}

func RestConfigToKubeconfigBytes(cfg *rest.Config) ([]byte, error) {
	apiConfig := clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"cluster": {
				Server:                   cfg.Host,
				CertificateAuthority:     cfg.CAFile,
				CertificateAuthorityData: cfg.CAData,
				InsecureSkipTLSVerify:    cfg.Insecure,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"user": {
				ClientCertificate:     cfg.CertFile,
				ClientCertificateData: cfg.CertData,
				ClientKey:             cfg.KeyFile,
				ClientKeyData:         cfg.KeyData,
				Token:                 cfg.BearerToken,
				Username:              cfg.Username,
				Password:              cfg.Password,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"context": {
				Cluster:  "cluster",
				AuthInfo: "user",
			},
		},
		CurrentContext: "context",
	}

	return clientcmd.Write(apiConfig)
}

func GenerateProviderTestCases(provider string, limit int, restConfig *rest.Config, secrets types.Secrets) ([]runtime.RawExtension, error) {
	valuesMap, ok := candidateValues[provider]
	if !ok {
		return nil, fmt.Errorf("no candidate values for provider %s", provider)
	}

	flat := make(map[string][]string)
	flattenCandidates("", valuesMap, flat)

	var dims [][]string
	var fieldOrder []string
	for field, vals := range flat {
		fieldOrder = append(fieldOrder, field)
		dims = append(dims, vals)
	}

	combos := cartesian(dims)
	var validCases []runtime.RawExtension
	for _, combo := range combos {
		// log.Printf("Testing combo: %v", combo)
		caseMap := make(map[string]string)
		for i, val := range combo {
			caseMap[fieldOrder[i]] = val
		}
		cfg, err := buildConfigStruct(provider, caseMap)
		if err != nil {
			continue
		}
		pc, err := validateConfig(provider, cfg, restConfig, secrets)
		if err == nil {
			// log.Printf("Valid case: %v", caseMap)
			validCases = append(validCases, pc.CloudProviderSpec)
			if limit > 0 && len(validCases) >= limit {
				break
			}
		} else {
			log.Printf("Invalid case: %v, error: %v", caseMap, err)
		}
	}
	return validCases, nil
}
