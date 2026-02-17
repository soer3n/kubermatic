package provider

import (
	"fmt"

	"k8c.io/kubermatic/v2/cmd/conformance-tester/pkg/ginkgo/settings"
)

func ConvertModifiersToAny[T any](mods []settings.MachineSpecModifier[T]) []settings.MachineSpecModifier[any] {
	result := make([]settings.MachineSpecModifier[any], len(mods))

	for i, m := range mods {
		mod := m

		result[i] = settings.MachineSpecModifier[any]{
			Modify: func(a any) {
				t, ok := a.(T)
				if !ok {
					panic(fmt.Sprintf("expected %T but got %T", *new(T), a))
				}
				mod.Modify(t)
			},
			Name:  mod.Name,
			Group: mod.Group,
		}
	}

	return result
}
