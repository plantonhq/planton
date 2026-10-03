package gcploggingsinkv1alpha1

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
	ginkgo.RunSpecs(t, "GcpLoggingSinkSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func nameRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: v}},
	}
}

var _ = ginkgo.Describe("GcpLoggingSinkSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// The canonical project sink: errors to a GCS bucket.
	minimal := func() *GcpLoggingSink {
		return &GcpLoggingSink{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpLoggingSink",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-logging-sink",
			},
			Spec: &GcpLoggingSinkSpec{
				Destination: &GcpLoggingSinkDestination{
					GcsBucket: litRef("my-log-archive-bucket"),
				},
				Filter: "severity>=ERROR",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept the canonical project sink to GCS", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept an explicit project scope", func() {
		target := minimal()
		target.Spec.Scope = &GcpLoggingSinkScope{ProjectId: litRef("my-project")}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a BigQuery destination with partitioned tables", func() {
		target := minimal()
		target.Spec.Destination = &GcpLoggingSinkDestination{
			BigqueryDataset:      litRef("projects/my-project/datasets/logs"),
			UsePartitionedTables: true,
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a Pub/Sub destination", func() {
		target := minimal()
		target.Spec.Destination = &GcpLoggingSinkDestination{
			PubsubTopic: litRef("projects/my-project/topics/log-stream"),
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a raw destination URI", func() {
		target := minimal()
		target.Spec.Destination = &GcpLoggingSinkDestination{
			RawUri: "logging.googleapis.com/projects/my-project/locations/global/buckets/central",
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a folder sink with children flags", func() {
		target := minimal()
		target.Spec.Scope = &GcpLoggingSinkScope{FolderId: litRef("123456789")}
		target.Spec.IncludeChildren = true
		target.Spec.InterceptChildren = true
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a folder sink whose folder is a GcpFolder reference, with children flags", func() {
		target := minimal()
		target.Spec.Scope = &GcpLoggingSinkScope{FolderId: nameRef("security-folder")}
		target.Spec.IncludeChildren = true
		target.Spec.InterceptChildren = true
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an organization sink with include_children", func() {
		target := minimal()
		target.Spec.Scope = &GcpLoggingSinkScope{OrganizationId: "987654321"}
		target.Spec.IncludeChildren = true
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a billing-account sink", func() {
		target := minimal()
		target.Spec.Scope = &GcpLoggingSinkScope{BillingAccount: "012345-6789AB-CDEF01"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept exclusions", func() {
		target := minimal()
		target.Spec.Exclusions = []*GcpLoggingSinkExclusion{{
			Name:        "drop-health-checks",
			Filter:      `httpRequest.requestUrl:"/healthz"`,
			Description: "Health-check noise",
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept project-scope writer identity controls", func() {
		target := minimal()
		uw := false
		target.Spec.UniqueWriterIdentity = &uw
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		custom := minimal()
		custom.Spec.CustomWriterIdentity = "log-writer@my-project.iam.gserviceaccount.com"
		gomega.Expect(validator.Validate(custom)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing destination", func() {
		target := minimal()
		target.Spec.Destination = nil
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an empty destination and multiple destinations", func() {
		empty := minimal()
		empty.Spec.Destination = &GcpLoggingSinkDestination{}
		gomega.Expect(validator.Validate(empty)).ToNot(gomega.Succeed())

		both := minimal()
		both.Spec.Destination = &GcpLoggingSinkDestination{
			GcsBucket:   litRef("bucket"),
			PubsubTopic: litRef("projects/p/topics/t"),
		}
		err := validator.Validate(both)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "exactly one destination")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject use_partitioned_tables on a non-BigQuery destination", func() {
		target := minimal()
		target.Spec.Destination.UsePartitionedTables = true
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject two scopes at once", func() {
		target := minimal()
		target.Spec.Scope = &GcpLoggingSinkScope{
			FolderId:       litRef("123"),
			OrganizationId: "456",
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())

		target = minimal()
		target.Spec.Scope = &GcpLoggingSinkScope{
			ProjectId: litRef("my-project"),
			FolderId:  nameRef("security-folder"),
		}
		gomega.Expect(validator.Validate(target).Error()).To(gomega.ContainSubstring("set at most one of project_id, folder_id"))
	})

	ginkgo.It("should reject children flags on a project scope given by reference", func() {
		target := minimal()
		target.Spec.Scope = &GcpLoggingSinkScope{ProjectId: nameRef("app-project")}
		target.Spec.IncludeChildren = true
		gomega.Expect(validator.Validate(target).Error()).To(gomega.ContainSubstring("include_children and intercept_children apply only to folder or organization scoped sinks"))
	})

	ginkgo.It("should reject children flags on project and billing scopes", func() {
		project := minimal()
		project.Spec.IncludeChildren = true
		gomega.Expect(validator.Validate(project)).ToNot(gomega.Succeed())

		billing := minimal()
		billing.Spec.Scope = &GcpLoggingSinkScope{BillingAccount: "012345-6789AB-CDEF01"}
		billing.Spec.InterceptChildren = true
		gomega.Expect(validator.Validate(billing)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject writer identity controls on non-project scopes", func() {
		folder := minimal()
		folder.Spec.Scope = &GcpLoggingSinkScope{FolderId: litRef("123")}
		uw := false
		folder.Spec.UniqueWriterIdentity = &uw
		gomega.Expect(validator.Validate(folder)).ToNot(gomega.Succeed())

		folderRef := minimal()
		folderRef.Spec.Scope = &GcpLoggingSinkScope{FolderId: nameRef("security-folder")}
		folderRef.Spec.CustomWriterIdentity = "writer@p.iam.gserviceaccount.com"
		gomega.Expect(validator.Validate(folderRef).Error()).To(gomega.ContainSubstring("apply only to project-scoped sinks"))

		org := minimal()
		org.Spec.Scope = &GcpLoggingSinkScope{OrganizationId: "456"}
		org.Spec.CustomWriterIdentity = "writer@p.iam.gserviceaccount.com"
		gomega.Expect(validator.Validate(org)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a BigQuery destination with unique_writer_identity=false", func() {
		target := minimal()
		target.Spec.Destination = &GcpLoggingSinkDestination{
			BigqueryDataset: litRef("projects/p/datasets/d"),
		}
		uw := false
		target.Spec.UniqueWriterIdentity = &uw
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "unique_writer_identity")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a malformed exclusion name and a missing exclusion filter", func() {
		badName := minimal()
		badName.Spec.Exclusions = []*GcpLoggingSinkExclusion{{Name: "-starts-with-dash", Filter: "severity>=ERROR"}}
		gomega.Expect(validator.Validate(badName)).ToNot(gomega.Succeed())

		noFilter := minimal()
		noFilter.Spec.Exclusions = []*GcpLoggingSinkExclusion{{Name: "valid-name"}}
		gomega.Expect(validator.Validate(noFilter)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})
})
