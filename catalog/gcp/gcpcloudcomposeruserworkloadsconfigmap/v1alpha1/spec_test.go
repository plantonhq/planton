package gcpcloudcomposeruserworkloadsconfigmapv1alpha1

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
	ginkgo.RunSpecs(t, "GcpCloudComposerUserWorkloadsConfigMapSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpCloudComposerUserWorkloadsConfigMapSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpCloudComposerUserWorkloadsConfigMap {
		return &GcpCloudComposerUserWorkloadsConfigMap{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpCloudComposerUserWorkloadsConfigMap",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "dag-configuration",
			},
			Spec: &GcpCloudComposerUserWorkloadsConfigMapSpec{
				Region:        "us-central1",
				Environment:   litRef("prod-airflow"),
				ConfigMapName: "dag-configuration",
				Data: map[string]string{
					"api_endpoint": "https://api.example.com/v2",
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

	ginkgo.It("should accept multiple plain data entries", func() {
		msg := minimal()
		msg.Spec.Data = map[string]string{
			"api_endpoint":    "https://api.example.com/v2",
			"batch_size":      "500",
			"enable_new_flow": "true",
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a single-character config_map_name", func() {
		msg := minimal()
		msg.Spec.ConfigMapName = "c"
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

	ginkgo.It("should reject a missing config_map_name", func() {
		msg := minimal()
		msg.Spec.ConfigMapName = ""
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a config_map_name with uppercase letters", func() {
		msg := minimal()
		msg.Spec.ConfigMapName = "Dag-Configuration"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a config_map_name ending with a hyphen", func() {
		msg := minimal()
		msg.Spec.ConfigMapName = "dag-configuration-"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a config_map_name with underscores", func() {
		msg := minimal()
		msg.Spec.ConfigMapName = "dag_configuration"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an empty data map", func() {
		msg := minimal()
		msg.Spec.Data = map[string]string{}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing spec", func() {
		msg := minimal()
		msg.Spec = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind literal", func() {
		msg := minimal()
		msg.Kind = "GcpCloudComposerUserWorkloadsConfigMaps"
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
