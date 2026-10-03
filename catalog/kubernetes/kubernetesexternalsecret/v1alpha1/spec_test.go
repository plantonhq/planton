package kubernetesexternalsecretv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesExternalSecret(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesExternalSecret Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func strPtr(v string) *string { return &v }

var _ = ginkgo.Describe("KubernetesExternalSecret Validation Tests", func() {
	var input *KubernetesExternalSecret

	ginkgo.BeforeEach(func() {
		input = &KubernetesExternalSecret{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesExternalSecret",
			Metadata:   &shared.CatalogObjectMetadata{Name: "app-db-credentials"},
			Spec: &KubernetesExternalSecretSpec{
				Namespace: literal("team-a"),
				StoreRef: &KubernetesExternalSecretStoreRef{
					StoreType: &KubernetesExternalSecretStoreRef_SecretStore{
						SecretStore: &KubernetesExternalSecretSecretStoreRef{Name: literal("team-a-gcp")},
					},
				},
				Data: []*KubernetesExternalSecretData{
					{
						SecretKey: "password",
						RemoteRef: &KubernetesExternalSecretRemoteRef{Key: "prod/app/db-password"},
					},
				},
			},
		}
	})

	ginkgo.Describe("valid configurations", func() {
		ginkgo.It("accepts a minimal explicit sync", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a cluster store reference", func() {
			input.Spec.StoreRef.StoreType = &KubernetesExternalSecretStoreRef_ClusterSecretStore{
				ClusterSecretStore: &KubernetesExternalSecretClusterSecretStoreRef{Name: literal("platform-gcp")},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a property extraction with version and decoding", func() {
			decoding := "Base64"
			input.Spec.Data[0].RemoteRef.Property = "password"
			input.Spec.Data[0].RemoteRef.Version = "2"
			input.Spec.Data[0].RemoteRef.DecodingStrategy = &decoding
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a dataFrom extract with rewrites", func() {
			input.Spec.Data = nil
			input.Spec.DataFrom = []*KubernetesExternalSecretDataFrom{
				{
					Source: &KubernetesExternalSecretDataFrom_Extract{
						Extract: &KubernetesExternalSecretRemoteRef{Key: "prod/app/all"},
					},
					Rewrite: []*KubernetesExternalSecretRewrite{
						{Source: "^prod/app/(.*)$", Target: "$1"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a dataFrom find by pattern and tags", func() {
			input.Spec.DataFrom = []*KubernetesExternalSecretDataFrom{
				{
					Source: &KubernetesExternalSecretDataFrom_Find{
						Find: &KubernetesExternalSecretFind{
							NameRegexp: "^app-.*",
							Tags:       map[string]string{"env": "prod"},
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a templated docker-registry target", func() {
			input.Spec.Target = &KubernetesExternalSecretTarget{
				Name: "regcred",
				Template: &KubernetesExternalSecretTemplate{
					Type: "kubernetes.io/dockerconfigjson",
					Data: map[string]string{
						".dockerconfigjson": `{"auths":{"ghcr.io":{"auth":"{{ .token }}"}}}`,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts an immutable bootstrap secret posture", func() {
			policy := "CreatedOnce"
			input.Spec.RefreshPolicy = &policy
			input.Spec.Target = &KubernetesExternalSecretTarget{Immutable: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("required fields and contracts", func() {
		ginkgo.It("rejects a missing namespace", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a missing store reference", func() {
			input.Spec.StoreRef = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a store reference that names no store", func() {
			input.Spec.StoreRef = &KubernetesExternalSecretStoreRef{}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("A store is required"))
		})

		ginkgo.It("rejects a store arm without its name", func() {
			input.Spec.StoreRef.StoreType = &KubernetesExternalSecretStoreRef_ClusterSecretStore{
				ClusterSecretStore: &KubernetesExternalSecretClusterSecretStoreRef{},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a sync that declares nothing", func() {
			input.Spec.Data = nil
			input.Spec.DataFrom = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a data entry without a secret key", func() {
			input.Spec.Data[0].SecretKey = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a data entry without a remote key", func() {
			input.Spec.Data[0].RemoteRef = &KubernetesExternalSecretRemoteRef{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a dataFrom pull without a source", func() {
			input.Spec.DataFrom = []*KubernetesExternalSecretDataFrom{{}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a find with no criteria", func() {
			input.Spec.DataFrom = []*KubernetesExternalSecretDataFrom{
				{
					Source: &KubernetesExternalSecretDataFrom_Find{
						Find: &KubernetesExternalSecretFind{Path: "prod/"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a rewrite without source or target", func() {
			input.Spec.DataFrom = []*KubernetesExternalSecretDataFrom{
				{
					Source: &KubernetesExternalSecretDataFrom_Extract{
						Extract: &KubernetesExternalSecretRemoteRef{Key: "prod/app/all"},
					},
					Rewrite: []*KubernetesExternalSecretRewrite{{Source: "^x$"}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an unknown creation policy", func() {
			policy := "Adopt"
			input.Spec.Target = &KubernetesExternalSecretTarget{CreationPolicy: &policy}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an unknown deletion policy", func() {
			policy := "Cascade"
			input.Spec.Target = &KubernetesExternalSecretTarget{DeletionPolicy: &policy}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a malformed refresh interval", func() {
			input.Spec.RefreshInterval = strPtr("hourly")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an unknown refresh policy", func() {
			input.Spec.RefreshPolicy = strPtr("Always")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})
	})
})
