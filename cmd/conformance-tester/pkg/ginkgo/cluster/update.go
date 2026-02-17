package cluster

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"

	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
)

func Update(name string, spec *kubermaticv1.ClusterSpec) {
	// st := getClusterState(name)
	By(fmt.Sprintf("Updating cluster %s\n", name))
	// simulate work
	time.Sleep(2 * time.Second)
	By(fmt.Sprintf("Cluster %s updated\n", name))
}
