package gcpvertexaifeaturegroupv1alpha1

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
	ginkgo.RunSpecs(t, "GcpVertexAiFeatureGroupSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpVertexAiFeatureGroupSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpVertexAiFeatureGroup {
		return &GcpVertexAiFeatureGroup{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiFeatureGroup",
			Metadata:   &shared.CloudResourceMetadata{Name: "customer-features"},
			Spec: &GcpVertexAiFeatureGroupSpec{
				Location:       "us-central1",
				FeatureGroupId: "customer_features",
				BigQuery: &GcpVertexAiFeatureGroupBigQuery{
					InputUri: litRef("bq://ml-project.features.customers"),
				},
				Features: []*GcpVertexAiFeatureGroupFeature{
					{FeatureId: "age"},
					{FeatureId: "lifetime_value", VersionColumnName: "ltv_usd"},
				},
			},
		}
	}

	ginkgo.It("should accept a group with a source and features", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set, and a source without the bq:// prefix", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("ml-project")
		msg.Spec.Description = "Customer features"
		msg.Spec.Labels = map[string]string{"team": "growth"}
		msg.Spec.BigQuery.InputUri = litRef("ml-project.features.customers")
		msg.Spec.BigQuery.EntityIdColumns = []string{"customer_id"}
		msg.Spec.Features[0].Description = "Age in years"
		msg.Spec.Features[0].Labels = map[string]string{"pii": "false"}
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a feature group id Google accepts", func() {
		msg := minimal()
		msg.Spec.FeatureGroupId = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		for _, bad := range []string{"customer-features", "1customers", "Customers", strings.Repeat("a", 129)} {
			msg = minimal()
			msg.Spec.FeatureGroupId = bad
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), bad)
		}
	})

	ginkgo.It("should require the source table when a BigQuery source is declared", func() {
		msg := minimal()
		msg.Spec.BigQuery.InputUri = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a feature id Google rejects", func() {
		msg := minimal()
		msg.Spec.Features[0].FeatureId = "life-time"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a duplicate feature id", func() {
		msg := minimal()
		msg.Spec.Features[1].FeatureId = "age"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
