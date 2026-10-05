package gcpcloudcomposeruserworkloadssecretv1alpha1

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
	ginkgo.RunSpecs(t, "GcpCloudComposerUserWorkloadsSecretSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpCloudComposerUserWorkloadsSecretSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpCloudComposerUserWorkloadsSecret {
		return &GcpCloudComposerUserWorkloadsSecret{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpCloudComposerUserWorkloadsSecret",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "airflow-db-credentials",
			},
			Spec: &GcpCloudComposerUserWorkloadsSecretSpec{
				Region:      "us-central1",
				Environment: litRef("prod-airflow"),
				SecretName:  "db-credentials",
				Data: map[string]string{
					// base64 of "postgresql://airflow:pass@10.0.0.5/mydb"
					"connection": "cG9zdGdyZXNxbDovL2FpcmZsb3c6cGFzc0AxMC4wLjAuNS9teWRi",
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept an omitted project_id (ambient project)", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project_id literal", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("my-gcp-project")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a reference-shaped environment", func() {
		msg := minimal()
		msg.Spec.Environment = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
				ValueFrom: &foreignkeyv1.ValueFromRef{Name: "prod-airflow"},
			},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a multi-digit region", func() {
		msg := minimal()
		msg.Spec.Region = "europe-west12"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept multiple base64-encoded data entries", func() {
		msg := minimal()
		msg.Spec.Data = map[string]string{
			// base64 of "sk-live-4f9a2b"
			"api-token": "c2stbGl2ZS00ZjlhMmI=",
			// base64 of "hunter2"
			"password": "aHVudGVyMg==",
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an unpadded base64 value", func() {
		msg := minimal()
		// base64 of "airflow" (no padding needed for some inputs)
		msg.Spec.Data = map[string]string{"user": "YWlyZmxvdw"}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a single-character secret_name", func() {
		msg := minimal()
		msg.Spec.SecretName = "s"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing region", func() {
		msg := minimal()
		msg.Spec.Region = ""
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a malformed region", func() {
		msg := minimal()
		msg.Spec.Region = "US-Central1"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing environment", func() {
		msg := minimal()
		msg.Spec.Environment = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing secret_name", func() {
		msg := minimal()
		msg.Spec.SecretName = ""
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a secret_name with uppercase letters", func() {
		msg := minimal()
		msg.Spec.SecretName = "Db-Credentials"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a secret_name ending with a hyphen", func() {
		msg := minimal()
		msg.Spec.SecretName = "db-credentials-"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a secret_name with underscores", func() {
		msg := minimal()
		msg.Spec.SecretName = "db_credentials"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an empty data map", func() {
		msg := minimal()
		msg.Spec.Data = map[string]string{}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a raw (non-base64) data value", func() {
		msg := minimal()
		msg.Spec.Data = map[string]string{
			"connection": "postgresql://airflow:pass@10.0.0.5/mydb",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("base64"))
	})

	ginkgo.It("should reject a data value containing whitespace", func() {
		msg := minimal()
		msg.Spec.Data = map[string]string{"token": "abc def"}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing spec", func() {
		msg := minimal()
		msg.Spec = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind literal", func() {
		msg := minimal()
		msg.Kind = "GcpCloudComposerUserWorkloadsSecrets"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept deletion_policy DELETE, PREVENT, ABANDON, and empty", func() {
		for _, policy := range []string{"DELETE", "PREVENT", "ABANDON", ""} {
			msg := minimal()
			msg.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.HaveOccurred(), "policy %q", policy)
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "RETAIN"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})
})
