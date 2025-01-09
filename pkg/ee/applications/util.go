/*
                  Kubermatic Enterprise Read-Only License
                         Version 1.0 ("KERO-1.0”)
                     Copyright © 2025 Kubermatic GmbH

   1.	You may only view, read and display for studying purposes the source
      code of the software licensed under this license, and, to the extent
      explicitly provided under this license, the binary code.
   2.	Any use of the software which exceeds the foregoing right, including,
      without limitation, its execution, compilation, copying, modification
      and distribution, is expressly prohibited.
   3.	THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND,
      EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
      MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
      IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
      CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
      TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
      SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

   END OF TERMS AND CONDITIONS
*/

package applications

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	semverlib "github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"

	kubermaticv1 "k8c.io/kubermatic/v2/pkg/apis/kubermatic/v1"

	"k8s.io/apimachinery/pkg/types"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// TemplateData is the root context injected into each application values.
type TemplateData struct {
	Cluster ClusterData
}

// ClusterData contains data related to the user cluster
// the application is rendered for.
type ClusterData struct {
	Name string
	// HumanReadableName is the user-specified cluster name.
	HumanReadableName string
	// OwnerEmail is the owner's e-mail address.
	OwnerEmail string
	// ClusterAddress stores access and address information of a cluster.
	Address kubermaticv1.ClusterAddress
	// Version is the exact current cluster version.
	Version string
	// MajorMinorVersion is a shortcut for common testing on "Major.Minor" on the
	// current cluster version.
	MajorMinorVersion string
	// AutoscalerVersion is the tag which should be used for the cluster autoscaler
	AutoscalerVersion string
}

// GetTemplateData fetches the related cluster object by the given cluster namespace, parses pre defined values to a template data struct.
func GetTemplateData(ctx context.Context, seedClient ctrlruntimeclient.Client, clusterName string) (*TemplateData, error) {
	cluster := &kubermaticv1.Cluster{}
	if err := seedClient.Get(ctx, types.NamespacedName{Name: clusterName}, cluster); err != nil {
		return nil, fmt.Errorf("failed to get cluster %q: %w", clusterName, err)
	}
	var clusterVersion *semverlib.Version
	if s := cluster.Status.Versions.ControlPlane.Semver(); s != nil {
		clusterVersion = s
	} else {
		clusterVersion = cluster.Spec.Version.Semver()
	}
	if clusterVersion != nil {
		clusterAutoscalerVersion, err := getAutoscalerImageTag(fmt.Sprintf("%d.%d", clusterVersion.Major(), clusterVersion.Minor()))
		if err != nil {
			return nil, fmt.Errorf("failed to parse autoscaler version for cluster %q", clusterName)
		}
		return &TemplateData{
			Cluster: ClusterData{
				Name:              cluster.Name,
				HumanReadableName: cluster.Spec.HumanReadableName,
				OwnerEmail:        cluster.Status.UserEmail,
				Address:           cluster.Status.Address,
				Version:           fmt.Sprintf("%d.%d.%d", clusterVersion.Major(), clusterVersion.Minor(), clusterVersion.Patch()),
				MajorMinorVersion: fmt.Sprintf("%d.%d", clusterVersion.Major(), clusterVersion.Minor()),
				AutoscalerVersion: clusterAutoscalerVersion,
			},
		}, nil
	}

	return nil, fmt.Errorf("failed to parse semver version for cluster %q", clusterName)
}

// RenderValueTemplate is rendering the given template data into the given map of values an error is returned when undefined values are used in the values map values.
func RenderValueTemplate(applicationValues map[string]interface{}, templateData *TemplateData) (map[string]interface{}, error) {
	yamlData, err := yaml.Marshal(applicationValues)
	if err != nil {
		return nil, err
	}

	parser := template.New("application-pre-defined-values")
	parsed, err := parser.Parse(string(yamlData))
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	if err := parsed.Execute(&buffer, templateData); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	parsedMap := make(map[string]interface{})
	err = yaml.Unmarshal(buffer.Bytes(), parsedMap)
	if err != nil {
		return nil, err
	}
	return parsedMap, nil
}

func getAutoscalerImageTag(majorMinorVersion string) (string, error) {
	switch majorMinorVersion {
	case "1.29":
		return "v1.29.5", nil
	case "1.30":
		return "v1.30.3", nil
	case "1.31":
		return "v1.31.1", nil
	}
	return "", fmt.Errorf("could not find cluster autoscaler tag for cluster minor-major version %v", majorMinorVersion)
}
