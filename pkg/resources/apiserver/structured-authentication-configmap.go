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

package apiserver

import (
	"k8c.io/kubermatic/v2/pkg/resources"
	"k8c.io/reconciler/pkg/reconciling"
	corev1 "k8s.io/api/core/v1"
	apiserverv1beta1 "k8s.io/apiserver/pkg/apis/apiserver/v1beta1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

const (
	OidcUsernamePrefix = "oidc:"
	OidcGroupsPrefix   = "oidc:"
)

// StructuredAuthenticationConfigReconciler returns function to create cm that contains structured authentication configuration for apiserver
// to work with oidc providers.
func StructuredAuthenticationConfigReconciler(data *resources.TemplateData, enableOIDCAuthentication bool) reconciling.NamedConfigMapReconcilerFactory {
	return func() (string, reconciling.ConfigMapReconciler) {
		return resources.StructuredAuthenticationConfig, func(c *corev1.ConfigMap) (*corev1.ConfigMap, error) {
			oidcSettings := data.Cluster().Spec.OIDC
			var issuerURL, clientID, usernameClaim, groupsClaim, groupsPrefix, usernamePrefix string
			// var requiredClaim string
			if oidcSettings.IssuerURL != "" && oidcSettings.ClientID != "" {
				issuerURL = oidcSettings.IssuerURL
				clientID = oidcSettings.ClientID

				if oidcSettings.UsernameClaim != "" {
					usernameClaim = oidcSettings.UsernameClaim
				}
				if oidcSettings.GroupsClaim != "" {
					groupsClaim = oidcSettings.GroupsClaim
				}
				// if oidcSettings.RequiredClaim != "" {
				// 	requiredClaim = oidcSettings.RequiredClaim
				// }
				if oidcSettings.GroupsPrefix != "" {
					groupsPrefix = oidcSettings.GroupsPrefix
				}
				if oidcSettings.UsernamePrefix != "" {
					usernamePrefix = oidcSettings.UsernamePrefix
				}
			} else if enableOIDCAuthentication {
				issuerURL = data.OIDCIssuerURL()
				clientID = data.OIDCIssuerClientID()
				usernameClaim = "email"
				groupsPrefix = "oidc:"
				groupsClaim = "groups"
			}

			defaultAuthenticator := apiserverv1beta1.JWTAuthenticator{
				Issuer: apiserverv1beta1.Issuer{
					URL: issuerURL,
					Audiences: []string{
						clientID,
					},
					CertificateAuthority: data.CABundle().String(),
				},
				ClaimMappings: apiserverv1beta1.ClaimMappings{
					Username: apiserverv1beta1.PrefixedClaimOrExpression{
						Claim:  usernameClaim,
						Prefix: ptr.To(usernamePrefix),
					},
					Groups: apiserverv1beta1.PrefixedClaimOrExpression{
						Claim:  groupsClaim,
						Prefix: ptr.To(groupsPrefix),
					},
				},
			}
			authConfig := apiserverv1beta1.AuthenticationConfiguration{
				JWT: []apiserverv1beta1.JWTAuthenticator{
					defaultAuthenticator,
				},
			}
			authConfigData, err := yaml.Marshal(authConfig)
			if err != nil {
				return nil, err
			}

			c.Data = map[string]string{
				"structured-authentication-config.yaml": string(authConfigData),
			}
			return c, nil
		}
	}
}
