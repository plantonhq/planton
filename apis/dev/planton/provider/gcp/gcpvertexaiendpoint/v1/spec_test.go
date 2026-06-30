package gcpvertexaiendpointv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpVertexAiEndpointSpec Suite")
}

var _ = ginkgo.Describe("GcpVertexAiEndpointSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpVertexAiEndpoint {
		return &GcpVertexAiEndpoint{
			ApiVersion: "gcp.planton.dev/v1",
			Kind:       "GcpVertexAiEndpoint",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-endpoint",
			},
			Spec: &GcpVertexAiEndpointSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				Location:    "us-central1",
				DisplayName: "My ML Endpoint",
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

	// Suppress unused variable warning.
	_ = strRef

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with description", func() {
		msg := minimal()
		msg.Spec.Description = "Endpoint for serving recommendation models"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with VPC-peered network", func() {
		msg := minimal()
		msg.Spec.Network = strRef("projects/123456789/global/networks/my-vpc")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with CMEK encryption", func() {
		msg := minimal()
		msg.Spec.KmsKeyName = strRef("projects/my-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with dedicated endpoint enabled", func() {
		msg := minimal()
		msg.Spec.DedicatedEndpointEnabled = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with dedicated endpoint and VPC network", func() {
		msg := minimal()
		msg.Spec.DedicatedEndpointEnabled = true
		msg.Spec.Network = strRef("projects/123456789/global/networks/my-vpc")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with PSC config (empty allowlist)", func() {
		msg := minimal()
		msg.Spec.PrivateServiceConnectConfig = &GcpVertexAiEndpointPrivateServiceConnectConfig{}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with PSC config and project allowlist", func() {
		msg := minimal()
		msg.Spec.PrivateServiceConnectConfig = &GcpVertexAiEndpointPrivateServiceConnectConfig{
			ProjectAllowlist: []string{"project-a", "project-b"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with valid numeric endpoint_name (1 digit)", func() {
		msg := minimal()
		msg.Spec.EndpointName = "1"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with valid numeric endpoint_name (10 digits)", func() {
		msg := minimal()
		msg.Spec.EndpointName = "1234567890"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with empty endpoint_name (auto-generated)", func() {
		msg := minimal()
		msg.Spec.EndpointName = ""
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

	ginkgo.It("should accept spec with CMEK and VPC network", func() {
		msg := minimal()
		msg.Spec.Network = strRef("projects/123456789/global/networks/my-vpc")
		msg.Spec.KmsKeyName = strRef("projects/my-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with PSC and CMEK", func() {
		msg := minimal()
		msg.Spec.PrivateServiceConnectConfig = &GcpVertexAiEndpointPrivateServiceConnectConfig{
			ProjectAllowlist: []string{"project-a"},
		}
		msg.Spec.KmsKeyName = strRef("projects/my-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept full-featured spec (VPC-peered with all options)", func() {
		msg := minimal()
		msg.Spec.Description = "Production recommendation endpoint"
		msg.Spec.Network = strRef("projects/123456789/global/networks/prod-vpc")
		msg.Spec.KmsKeyName = strRef("projects/my-project/locations/us-central1/keyRings/prod-ring/cryptoKeys/endpoint-key")
		msg.Spec.DedicatedEndpointEnabled = true
		msg.Spec.EndpointName = "9876543210"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject spec with missing project_id", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

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
		msg := &GcpVertexAiEndpoint{
			ApiVersion: "gcp.planton.dev/v1",
			Kind:       "GcpVertexAiEndpoint",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-endpoint",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject network + private_service_connect_config mutual exclusion", func() {
		msg := minimal()
		msg.Spec.Network = strRef("projects/123456789/global/networks/my-vpc")
		msg.Spec.PrivateServiceConnectConfig = &GcpVertexAiEndpointPrivateServiceConnectConfig{
			ProjectAllowlist: []string{"project-a"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject dedicated_endpoint_enabled + PSC mutual exclusion", func() {
		msg := minimal()
		msg.Spec.DedicatedEndpointEnabled = true
		msg.Spec.PrivateServiceConnectConfig = &GcpVertexAiEndpointPrivateServiceConnectConfig{}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject endpoint_name with leading zero", func() {
		msg := minimal()
		msg.Spec.EndpointName = "0123456789"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject endpoint_name with non-numeric characters", func() {
		msg := minimal()
		msg.Spec.EndpointName = "abc123"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject endpoint_name exceeding 10 digits", func() {
		msg := minimal()
		msg.Spec.EndpointName = "12345678901"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject endpoint_name that is just zero", func() {
		msg := minimal()
		msg.Spec.EndpointName = "0"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject endpoint_name with hyphens", func() {
		msg := minimal()
		msg.Spec.EndpointName = "123-456"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject endpoint_name with spaces", func() {
		msg := minimal()
		msg.Spec.EndpointName = "123 456"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
