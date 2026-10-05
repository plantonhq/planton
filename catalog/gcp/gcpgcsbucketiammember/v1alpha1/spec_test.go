package gcpgcsbucketiammemberv1alpha1

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
	ginkgo.RunSpecs(t, "GcpGcsBucketIamMemberSpec Suite")
}

func valueOf(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s},
	}
}

func refTo(name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{
			Name:      name,
			FieldPath: fieldPath,
		}},
	}
}

var _ = ginkgo.Describe("GcpGcsBucketIamMemberSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpGcsBucketIamMember {
		return &GcpGcsBucketIamMember{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpGcsBucketIamMember",
			Metadata:   &shared.CatalogObjectMetadata{Name: "sink-writer"},
			Spec: &GcpGcsBucketIamMemberSpec{
				Bucket: valueOf("acme-audit-logs"),
				Role:   valueOf("roles/storage.objectCreator"),
				Member: valueOf("serviceAccount:service-123456789@gcp-sa-logging.iam.gserviceaccount.com"),
			},
		}
	}

	ginkgo.It("accepts an object-creator grant to a sink writer", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("accepts a dotted bucket name and references", func() {
		msg := minimal()
		msg.Spec.Bucket = valueOf("logs.example.com")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.Bucket = refTo("audit-logs", "status.outputs.bucket_id")
		msg.Spec.Member = refTo("org-audit", "status.outputs.writer_identity")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("accepts a prefix-scoped condition", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpGcsBucketIamMemberCondition{
			Title:      "logs-prefix-only",
			Expression: `resource.name.startsWith("projects/_/buckets/acme-audit-logs/objects/logs/")`,
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("refuses a gs:// URL, uppercase, and a one-character bucket", func() {
		for _, bucket := range []string{"gs://acme-audit-logs", "Acme-Logs", "a"} {
			msg := minimal()
			msg.Spec.Bucket = valueOf(bucket)
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), bucket)
		}
	})

	ginkgo.It("refuses a missing bucket, role, or member", func() {
		for _, mutate := range []func(*GcpGcsBucketIamMemberSpec){
			func(s *GcpGcsBucketIamMemberSpec) { s.Bucket = nil },
			func(s *GcpGcsBucketIamMemberSpec) { s.Role = nil },
			func(s *GcpGcsBucketIamMemberSpec) { s.Member = nil },
		} {
			msg := minimal()
			mutate(msg.Spec)
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		}
	})

	ginkgo.It("refuses a malformed role and member", func() {
		msg := minimal()
		msg.Spec.Role = valueOf("storage.objectCreator")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Member = valueOf("writer@example.com")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Member = valueOf("deleted:user:gone@example.com?uid=1")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
