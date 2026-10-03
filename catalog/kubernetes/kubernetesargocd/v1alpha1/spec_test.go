package kubernetesargocdv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesArgocd(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesArgocd Suite")
}

func int32Ptr(i int32) *int32 { return &i }
func boolPtr(b bool) *bool    { return &b }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind:      kind,
				Name:      name,
				FieldPath: fieldPath,
			},
		},
	}
}

var _ = ginkgo.Describe("KubernetesArgocd Validation Tests", func() {
	var input *KubernetesArgocd

	ginkgo.BeforeEach(func() {
		input = &KubernetesArgocd{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesArgocd",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "gitops",
			},
			Spec: &KubernetesArgocdSpec{
				Namespace: literal("argocd"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (every optional block omitted)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "argocd", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("component replica and autoscaling knobs should be valid", func() {
			input.Spec.Controller = &KubernetesArgocdComponent{Replicas: int32Ptr(2)}
			input.Spec.Server = &KubernetesArgocdServer{
				Autoscaling: &KubernetesArgocdAutoscaling{Enabled: true, MinReplicas: int32Ptr(2), MaxReplicas: int32Ptr(5)},
				Insecure:    true,
			}
			input.Spec.RepoServer = &KubernetesArgocdScalableComponent{
				Autoscaling: &KubernetesArgocdAutoscaling{Enabled: true, MinReplicas: int32Ptr(2), MaxReplicas: int32Ptr(4)},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("autoscaling with equal min and max should be valid", func() {
			input.Spec.Server = &KubernetesArgocdServer{
				Autoscaling: &KubernetesArgocdAutoscaling{Enabled: true, MinReplicas: int32Ptr(3), MaxReplicas: int32Ptr(3)},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("the bundled redis arm should be valid", func() {
			input.Spec.Redis = &KubernetesArgocdRedis{
				Arm: &KubernetesArgocdRedis_Bundled{Bundled: &KubernetesArgocdRedisBundled{}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("the redis-ha arm at the quorum floor should be valid", func() {
			input.Spec.Redis = &KubernetesArgocdRedis{
				Arm: &KubernetesArgocdRedis_Ha{Ha: &KubernetesArgocdRedisHa{Replicas: int32Ptr(3)}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an external redis with a literal host should be valid", func() {
			input.Spec.Redis = &KubernetesArgocdRedis{
				Arm: &KubernetesArgocdRedis_External{External: &KubernetesArgocdRedisExternal{
					Host:                  literal("redis.cache.svc.cluster.local"),
					Port:                  int32Ptr(6379),
					CredentialsSecretName: "argocd-redis-creds",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an external redis referencing a KubernetesValkey should be valid", func() {
			input.Spec.Redis = &KubernetesArgocdRedis{
				Arm: &KubernetesArgocdRedis_External{External: &KubernetesArgocdRedisExternal{
					Host: valueFrom(catalogkind.CatalogKind_KubernetesValkey, "cache", "status.outputs.service"),
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("direct OIDC sso with a labeled client-secret reference should be valid", func() {
			input.Spec.Sso = &KubernetesArgocdSso{
				Oidc: &KubernetesArgocdOidc{
					Name:               "Okta",
					Issuer:             "https://example.okta.com",
					ClientId:           "argocd",
					ClientSecretSecret: &KubernetesArgocdSecretKeyRef{Name: "argocd-oidc", Key: "clientSecret"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a public OIDC client (PKCE, no client secret) should be valid", func() {
			input.Spec.Sso = &KubernetesArgocdSso{
				Oidc: &KubernetesArgocdOidc{Name: "Okta", Issuer: "https://example.okta.com", ClientId: "argocd"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("dex connector configuration should be valid", func() {
			input.Spec.Sso = &KubernetesArgocdSso{
				DexConfig: "connectors:\n  - type: github\n    id: github\n    name: GitHub\n    config:\n      clientID: abc\n      clientSecret: $argocd-github-sso:clientSecret\n",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("rbac policy configuration should be valid", func() {
			input.Spec.Rbac = &KubernetesArgocdRbac{
				PolicyDefault: "role:readonly",
				PolicyCsv:     "g, my-org:platform, role:admin\n",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("public repositories of both types should be valid", func() {
			gitType := "git"
			helmType := "helm"
			input.Spec.Repositories = []*KubernetesArgocdRepository{
				{Name: "platform", Url: "https://github.com/org/platform"},
				{Name: "apps", Url: "https://github.com/org/apps", Type: &gitType},
				{Name: "charts", Url: "https://charts.example.com", Type: &helmType},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("reconciliation timeout in each unit should be valid", func() {
			for _, v := range []string{"120s", "3m", "1h"} {
				input.Spec.ReconciliationTimeout = v
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("crds lifecycle toggles should be valid", func() {
			input.Spec.Crds = &KubernetesArgocdCrds{Install: boolPtr(true), Keep: boolPtr(false)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("component toggles, image override and scheduling should be valid", func() {
			input.Spec.Notifications = &KubernetesArgocdToggleableComponent{Enabled: boolPtr(false)}
			input.Spec.Dex = &KubernetesArgocdToggleableComponent{Enabled: boolPtr(false)}
			input.Spec.CommitServer = &KubernetesArgocdToggleableComponent{Enabled: boolPtr(true)}
			input.Spec.Image = &KubernetesArgocdImage{Repository: "my.registry.com/argoproj/argocd", Tag: "v3.4.5", PullSecretName: "mirror-pull"}
			input.Spec.Scheduling = &KubernetesArgocdScheduling{
				NodeSelector:      map[string]string{"role": "platform"},
				PriorityClassName: "platform-critical",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("helm values escape hatch should be valid", func() {
			input.Spec.HelmValues = "notifications:\n  notifiers:\n    service.slack: |\n      token: $slack-token\n"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero controller replicas should fail", func() {
			input.Spec.Controller = &KubernetesArgocdComponent{Replicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("autoscaling max below min should fail the bounds rule", func() {
			input.Spec.Server = &KubernetesArgocdServer{
				Autoscaling: &KubernetesArgocdAutoscaling{Enabled: true, MinReplicas: int32Ptr(4), MaxReplicas: int32Ptr(2)},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("max_replicas"))
		})

		ginkgo.It("redis-ha below the quorum floor should fail", func() {
			input.Spec.Redis = &KubernetesArgocdRedis{
				Arm: &KubernetesArgocdRedis_Ha{Ha: &KubernetesArgocdRedisHa{Replicas: int32Ptr(1)}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an external redis without a host should fail", func() {
			input.Spec.Redis = &KubernetesArgocdRedis{
				Arm: &KubernetesArgocdRedis_External{External: &KubernetesArgocdRedisExternal{}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an external redis port out of range should fail", func() {
			input.Spec.Redis = &KubernetesArgocdRedis{
				Arm: &KubernetesArgocdRedis_External{External: &KubernetesArgocdRedisExternal{
					Host: literal("redis.cache.svc.cluster.local"),
					Port: int32Ptr(70000),
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("OIDC without an issuer should fail", func() {
			input.Spec.Sso = &KubernetesArgocdSso{
				Oidc: &KubernetesArgocdOidc{Name: "Okta", ClientId: "argocd"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("OIDC without a client id should fail", func() {
			input.Spec.Sso = &KubernetesArgocdSso{
				Oidc: &KubernetesArgocdOidc{Name: "Okta", Issuer: "https://example.okta.com"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a client-secret reference missing its key should fail", func() {
			input.Spec.Sso = &KubernetesArgocdSso{
				Oidc: &KubernetesArgocdOidc{
					Name: "Okta", Issuer: "https://example.okta.com", ClientId: "argocd",
					ClientSecretSecret: &KubernetesArgocdSecretKeyRef{Name: "argocd-oidc"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a repository without a url should fail", func() {
			input.Spec.Repositories = []*KubernetesArgocdRepository{{Name: "platform"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a repository with an unsupported type should fail", func() {
			badType := "oci"
			input.Spec.Repositories = []*KubernetesArgocdRepository{
				{Name: "images", Url: "https://registry.example.com", Type: &badType},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a malformed reconciliation timeout should fail", func() {
			for _, v := range []string{"120", "2 minutes", "5d"} {
				input.Spec.ReconciliationTimeout = v
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			}
		})
	})
})
