package settings

import (
	"context"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/machine-controller/sdk/providerconfig"

	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/options"
	legacytypes "k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/types"
)

// machineSpecModifier is a struct that holds a name and a modify function for a machine spec.
type MachineSpecModifier[T any] struct {
	Name   string
	Group  string
	Modify func(spec T)
}

type DatacenterSetting struct {
	Name     string
	Group    string
	Modifier func(dc *kubermaticv1.Datacenter)
}

type DefaultDatacenterSettings struct {
	VPCs []VPC
}

type VPC struct {
	Name    string
	Subnets []kubermaticv1.Subnet
}

type Provider string

type ProviderInterface interface {
	CpuModifiers(cpus []int) []MachineSpecModifier[any]
	DiskModifiers(sizes []string) []MachineSpecModifier[any]
	MachineSettings(ctx context.Context, providerConfig *providerconfig.Config, namespace string, secrets legacytypes.Secrets, resources *options.ResourceSettings) []MachineSpecModifier[any]
	MemoryModifiers(memories []string) []MachineSpecModifier[any]
	DiscoverDefaultDatacenterSettings(ctx context.Context, providerConfig *providerconfig.Config, secrets legacytypes.Secrets) (*DefaultDatacenterSettings, error)
	BuildDefaultDatacenterSettings(settings *DefaultDatacenterSettings) []DatacenterSetting
}
