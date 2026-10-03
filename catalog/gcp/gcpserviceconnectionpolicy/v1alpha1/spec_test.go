package gcpserviceconnectionpolicyv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpServiceConnectionPolicySpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpServiceConnectionPolicySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpServiceConnectionPolicy {
		return &GcpServiceConnectionPolicy{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpServiceConnectionPolicy",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-scp",
			},
			Spec: &GcpServiceConnectionPolicySpec{
				Location:     "us-central1",
				Network:      litRef("projects/my-proj/global/networks/my-vpc"),
				ServiceClass: "gcp-memorystore",
				PscConfig: &GcpServiceConnectionPolicyPscConfig{
					Subnetworks: []*foreignkeyv1.StringValueOrRef{
						litRef("projects/my-proj/regions/us-central1/subnetworks/my-subnet"),
					},
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal spec (location + network + service class + one subnet)", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project_id literal", func() {
		target := minimal()
		target.Spec.ProjectId = litRef("my-gcp-project-123")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an explicit policy_name", func() {
		target := minimal()
		target.Spec.PolicyName = "memorystore-policy"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a multi-digit region", func() {
		target := minimal()
		target.Spec.Location = "europe-west12"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a third-party service class", func() {
		target := minimal()
		target.Spec.ServiceClass = "test-service-a3dfcx"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a description and labels", func() {
		target := minimal()
		target.Spec.Description = "Authorizes Memorystore for Valkey in the shared VPC"
		target.Spec.Labels = map[string]string{"team": "platform"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a spec without psc_config (API-optional)", func() {
		target := minimal()
		target.Spec.PscConfig = nil
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept multiple subnets", func() {
		target := minimal()
		target.Spec.PscConfig.Subnetworks = []*foreignkeyv1.StringValueOrRef{
			litRef("subnet-a"),
			litRef("subnet-b"),
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a connection limit", func() {
		target := minimal()
		target.Spec.PscConfig.Limit = 10
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a custom producer hierarchy allowlist with the matching location mode", func() {
		target := minimal()
		target.Spec.PscConfig.ProducerInstanceLocation = "CUSTOM_RESOURCE_HIERARCHY_LEVELS"
		target.Spec.PscConfig.AllowedGoogleProducersResourceHierarchyLevels = []string{
			"projects/my-project-id",
			"folders/891",
			"organizations/123",
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept reference-shaped network and subnetworks", func() {
		target := minimal()
		target.Spec.Network = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
				ValueFrom: &foreignkeyv1.ValueFromRef{Name: "my-vpc"},
			},
		}
		target.Spec.PscConfig.Subnetworks = []*foreignkeyv1.StringValueOrRef{
			{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Name: "my-subnet"},
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing location", func() {
		target := minimal()
		target.Spec.Location = ""
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("location"))
	})

	ginkgo.It("should reject a zone where a region is expected", func() {
		target := minimal()
		target.Spec.Location = "us-central1-a"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing network", func() {
		target := minimal()
		target.Spec.Network = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing service_class", func() {
		target := minimal()
		target.Spec.ServiceClass = ""
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.ToLower(err.Error())).To(gomega.ContainSubstring("service_class"))
	})

	ginkgo.It("should reject an uppercase service_class", func() {
		target := minimal()
		target.Spec.ServiceClass = "GCP-Memorystore"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a policy_name starting with a digit", func() {
		target := minimal()
		target.Spec.PolicyName = "1-bad-name"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a policy_name ending with a hyphen", func() {
		target := minimal()
		target.Spec.PolicyName = "bad-name-"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an empty subnetworks list when psc_config is set", func() {
		target := minimal()
		target.Spec.PscConfig.Subnetworks = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("subnetworks"))
	})

	ginkgo.It("should reject a negative connection limit", func() {
		target := minimal()
		target.Spec.PscConfig.Limit = -1
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid producer_instance_location value", func() {
		target := minimal()
		target.Spec.PscConfig.ProducerInstanceLocation = "ANYWHERE"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject CUSTOM_RESOURCE_HIERARCHY_LEVELS without an allowlist", func() {
		target := minimal()
		target.Spec.PscConfig.ProducerInstanceLocation = "CUSTOM_RESOURCE_HIERARCHY_LEVELS"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("allowed_google_producers_resource_hierarchy_levels"))
	})

	ginkgo.It("should reject an allowlist without CUSTOM_RESOURCE_HIERARCHY_LEVELS", func() {
		target := minimal()
		target.Spec.PscConfig.AllowedGoogleProducersResourceHierarchyLevels = []string{"projects/my-project-id"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a malformed hierarchy allowlist entry", func() {
		target := minimal()
		target.Spec.PscConfig.ProducerInstanceLocation = "CUSTOM_RESOURCE_HIERARCHY_LEVELS"
		target.Spec.PscConfig.AllowedGoogleProducersResourceHierarchyLevels = []string{"my-project-id"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing spec", func() {
		target := minimal()
		target.Spec = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing metadata", func() {
		target := minimal()
		target.Metadata = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind literal", func() {
		target := minimal()
		target.Kind = "GcpServiceConnexionPolicy"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong api_version literal", func() {
		target := minimal()
		target.ApiVersion = "gcp.planton.dev/v2"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
