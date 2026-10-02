package gcpsccbigqueryexportv1alpha1

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
	ginkgo.RunSpecs(t, "GcpSccBigQueryExportSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpSccBigQueryExportSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpSccBigQueryExport {
		return &GcpSccBigQueryExport{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpSccBigQueryExport",
			Metadata:   &shared.CloudResourceMetadata{Name: "findings-history"},
			Spec: &GcpSccBigQueryExportSpec{
				BigQueryExportId: "findings-history",
				Dataset:          litRef("projects/sec/datasets/scc_findings"),
			},
		}
	}

	ginkgo.It("should accept a project export to a dataset", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept an organization export to a dataset self link with every field set", func() {
		msg := minimal()
		msg.Spec.Scope = &GcpSccBigQueryExportScope{OrganizationId: "123456789012"}
		msg.Spec.Dataset = litRef("https://bigquery.googleapis.com/bigquery/v2/projects/sec/datasets/scc_findings")
		msg.Spec.Filter = `state = "ACTIVE" AND NOT mute = "MUTED"`
		msg.Spec.Description = "Active, unmuted findings"
		msg.Spec.Location = "global"
		msg.Spec.DeletionPolicy = "DELETE"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a dataset", func() {
		msg := minimal()
		msg.Spec.Dataset = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed dataset", func() {
		for _, dataset := range []string{"scc_findings", "projects/sec/datasets/scc-findings", "projects/sec/tables/t"} {
			msg := minimal()
			msg.Spec.Dataset = litRef(dataset)
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), dataset)
		}
	})

	ginkgo.It("should reject an export id outside Google's rule", func() {
		msg := minimal()
		msg.Spec.BigQueryExportId = "Findings_History"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject two scope arms", func() {
		msg := minimal()
		msg.Spec.Scope = &GcpSccBigQueryExportScope{ProjectId: litRef("p"), OrganizationId: "123"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
