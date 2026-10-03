package gcpkmskeyhandlev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	gcpgcsbucketv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpgcsbucket/v1alpha1"
	"github.com/plantonhq/planton/pkg/refannotations"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpKmsKeyHandleSpec Suite")
}

var _ = ginkgo.Describe("GcpKmsKeyHandleSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpKmsKeyHandle {
		return &GcpKmsKeyHandle{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpKmsKeyHandle",
			Metadata:   &shared.CatalogObjectMetadata{Name: "orders-bucket-key"},
			Spec: &GcpKmsKeyHandleSpec{
				Location:             "us-central1",
				ResourceTypeSelector: "storage.googleapis.com/Bucket",
			},
		}
	}

	ginkgo.It("should accept a bucket key in a region", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set, a multi-region, and a dashed service", func() {
		msg := minimal()
		msg.Spec.ProjectId = &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "orders-prod"}}
		msg.Spec.Location = "us"
		msg.Spec.ResourceTypeSelector = "secretmanager.googleapis.com/Secret"
		msg.Spec.KeyHandleName = "orders-secrets"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.ResourceTypeSelector = "backup-dr.googleapis.com/BackupVault"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a location and a resource type", func() {
		msg := minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.ResourceTypeSelector = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a resource type that is not service.googleapis.com/Type", func() {
		for _, selector := range []string{"Bucket", "storage.googleapis.com", "storage.googleapis.com/bucket", "storage/Bucket"} {
			msg := minimal()
			msg.Spec.ResourceTypeSelector = selector
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), selector)
		}
	})

	ginkgo.It("should reject a malformed location", func() {
		msg := minimal()
		msg.Spec.Location = "US-Central1"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})

// The key fields of the kinds Autokey serves accept a key handle beside a
// GcpKmsKey: a valueFrom naming GcpKmsKeyHandle composes from its kms_key
// output, and one naming no kind still reads as a GcpKmsKey's key_id.
func TestAutokeyConsumerAcceptsAKeyHandle(t *testing.T) {
	field := refannotations.Of((&gcpgcsbucketv1alpha1.GcpGcsBucketSpec{}).ProtoReflect().Descriptor().Fields().ByName("kms_key_name"))
	if path, ok := field.DefaultPath(catalogkind.CatalogKind_GcpKmsKeyHandle); !ok || path != "status.outputs.kms_key" {
		t.Fatalf("GcpGcsBucket.kms_key_name composes from a key handle at %q (ok=%t), want status.outputs.kms_key", path, ok)
	}
	if kind := field.EffectiveKind(catalogkind.CatalogKind_unspecified); kind != catalogkind.CatalogKind_GcpKmsKey {
		t.Fatalf("a kindless valueFrom on GcpGcsBucket.kms_key_name reads as %s, want GcpKmsKey", kind)
	}
}
