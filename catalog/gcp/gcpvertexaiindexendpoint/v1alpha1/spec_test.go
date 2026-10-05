package gcpvertexaiindexendpointv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpVertexAiIndexEndpointSpec Suite")
}

var _ = ginkgo.Describe("GcpVertexAiIndexEndpointSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpVertexAiIndexEndpoint {
		return &GcpVertexAiIndexEndpoint{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiIndexEndpoint",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-index-endpoint",
			},
			Spec: &GcpVertexAiIndexEndpointSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				Location:    "us-central1",
				DisplayName: "My Index Endpoint",
			},
		}
	}

	strRef := func(val string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: val,
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with description", func() {
		msg := minimal()
		msg.Spec.Description = "Serving surface for the product embeddings index"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec without project_id (ambient project)", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a public endpoint", func() {
		msg := minimal()
		msg.Spec.PublicEndpointEnabled = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a VPC-peered endpoint", func() {
		msg := minimal()
		msg.Spec.Network = strRef("projects/123456789/global/networks/my-vpc")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a network reference by self-link", func() {
		msg := minimal()
		msg.Spec.Network = strRef("https://www.googleapis.com/compute/v1/projects/my-project/global/networks/my-vpc")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a PSC endpoint (empty allowlist)", func() {
		msg := minimal()
		msg.Spec.PrivateServiceConnectConfig = &GcpVertexAiIndexEndpointPrivateServiceConnectConfig{
			EnablePrivateServiceConnect: true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a PSC endpoint with project allowlist", func() {
		msg := minimal()
		msg.Spec.PrivateServiceConnectConfig = &GcpVertexAiIndexEndpointPrivateServiceConnectConfig{
			EnablePrivateServiceConnect: true,
			ProjectAllowlist:            []string{"consumer-a", "consumer-b"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with user labels", func() {
		msg := minimal()
		msg.Spec.Labels = map[string]string{"team": "ml-platform", "cost-center": "research"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with display_name at max length (128 chars)", func() {
		msg := minimal()
		msg.Spec.DisplayName = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
			"abcdefghijklmnopqrstuvwxyz" +
			"0123456789012345678901234567890123456789" +
			"abcdefghijklmnopqrstuvwxyz0123456789"
		gomega.Expect(len(msg.Spec.DisplayName)).To(gomega.Equal(128))
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a private-by-default endpoint (no connectivity arm)", func() {
		msg := minimal()
		msg.Spec.PublicEndpointEnabled = false
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept full-featured public spec", func() {
		msg := minimal()
		msg.Spec.Description = "Production vector-search serving surface"
		msg.Spec.PublicEndpointEnabled = true
		msg.Spec.Labels = map[string]string{"env": "prod"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject spec with missing location", func() {
		msg := minimal()
		msg.Spec.Location = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with missing display_name", func() {
		msg := minimal()
		msg.Spec.DisplayName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with display_name exceeding 128 chars", func() {
		msg := minimal()
		msg.Spec.DisplayName = "A very long display name that exceeds the maximum allowed length of one hundred and twenty-eight characters and should be rejected by validation"
		gomega.Expect(len(msg.Spec.DisplayName)).To(gomega.BeNumerically(">", 128))
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject network + PSC mutual exclusion", func() {
		msg := minimal()
		msg.Spec.Network = strRef("projects/123456789/global/networks/my-vpc")
		msg.Spec.PrivateServiceConnectConfig = &GcpVertexAiIndexEndpointPrivateServiceConnectConfig{
			EnablePrivateServiceConnect: true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject public + network mutual exclusion", func() {
		msg := minimal()
		msg.Spec.PublicEndpointEnabled = true
		msg.Spec.Network = strRef("projects/123456789/global/networks/my-vpc")
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject public + PSC mutual exclusion", func() {
		msg := minimal()
		msg.Spec.PublicEndpointEnabled = true
		msg.Spec.PrivateServiceConnectConfig = &GcpVertexAiIndexEndpointPrivateServiceConnectConfig{
			EnablePrivateServiceConnect: true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject PSC block with enable_private_service_connect false", func() {
		msg := minimal()
		msg.Spec.PrivateServiceConnectConfig = &GcpVertexAiIndexEndpointPrivateServiceConnectConfig{
			EnablePrivateServiceConnect: false,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with wrong api_version", func() {
		msg := minimal()
		msg.ApiVersion = "wrong/v1"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with wrong kind", func() {
		msg := minimal()
		msg.Kind = "WrongKind"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with missing metadata", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with missing spec", func() {
		msg := &GcpVertexAiIndexEndpoint{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiIndexEndpoint",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-index-endpoint",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// ──────────────── CMEK, PSC automation, deletion policy ────────────────

	ginkgo.It("should accept a CMEK key", func() {
		msg := minimal()
		msg.Spec.KmsKeyName = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "projects/p/locations/us-central1/keyRings/kr/cryptoKeys/k",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// The Vertex AI API accepts psc_automation_configs on an index
	// endpoint create but silently drops them (nothing stored, no
	// consumer-side endpoint provisioned — API-verified live), and the
	// immutable PSC block would turn that drop into a perpetual
	// replacement diff. The spec therefore refuses the field on this
	// kind; project_allowlist is the supported PSC consumption path.
	ginkgo.It("should reject PSC automation configs (API silently drops them on index endpoints)", func() {
		msg := minimal()
		msg.Spec.PublicEndpointEnabled = false
		msg.Spec.PrivateServiceConnectConfig = &GcpVertexAiIndexEndpointPrivateServiceConnectConfig{
			EnablePrivateServiceConnect: true,
			ProjectAllowlist:            []string{"consumer-project"},
			PscAutomationConfigs: []*GcpVertexAiIndexEndpointPscAutomationConfig{
				{
					Network: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
							Value: "projects/consumer-project/global/networks/prod-vpc",
						},
					},
					ProjectId: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "consumer-project"},
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("psc_automation_configs is not honored"))
	})

	ginkgo.It("should accept all deletion_policy values", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			msg := minimal()
			msg.Spec.DeletionPolicy = policy
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "RETAIN"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("deletion_policy"))
	})
})
