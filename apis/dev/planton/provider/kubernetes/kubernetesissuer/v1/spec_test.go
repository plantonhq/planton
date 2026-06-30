package kubernetesissuerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesIssuer(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesIssuer Suite")
}

var _ = ginkgo.Describe("KubernetesIssuer Validation Tests", func() {
	var input *KubernetesIssuer

	ginkgo.BeforeEach(func() {
		input = &KubernetesIssuer{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesIssuer",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-issuer",
			},
			Spec: &KubernetesIssuerSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "cert-manager",
					},
				},
				IssuerType: &KubernetesIssuerSpec_Ca{
					Ca: &CaIssuerConfig{
						CaSecretName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "root-ca-secret",
							},
						},
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with CA issuer", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with SelfSigned issuer", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.IssuerType = &KubernetesIssuerSpec_SelfSigned{
					SelfSigned: &SelfSignedIssuerConfig{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing issuer_type", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.IssuerType = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing namespace", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Namespace = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("empty namespace value", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Namespace = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing ca_secret_name for CA issuer", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.IssuerType = &KubernetesIssuerSpec_Ca{
					Ca: &CaIssuerConfig{
						CaSecretName: nil,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("empty ca_secret_name value for CA issuer", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.IssuerType = &KubernetesIssuerSpec_Ca{
					Ca: &CaIssuerConfig{
						CaSecretName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing metadata", func() {
			ginkgo.It("should return a validation error", func() {
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
