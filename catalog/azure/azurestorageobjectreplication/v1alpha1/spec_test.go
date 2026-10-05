package azurestorageobjectreplicationv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureStorageObjectReplicationSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureStorageObjectReplicationSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	srcAccountId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/plantonorsrc"
	dstAccountId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/plantonordst"
)

// minimal valid spec: one rule replicating a container pair.
func minimalSpec() *AzureStorageObjectReplication {
	return &AzureStorageObjectReplication{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureStorageObjectReplication",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-object-replication",
		},
		Spec: &AzureStorageObjectReplicationSpec{
			SourceStorageAccountId:      literal(srcAccountId),
			DestinationStorageAccountId: literal(dstAccountId),
			Rules: []*AzureStorageObjectReplicationRule{
				{
					SourceContainerName:      literal("invoices"),
					DestinationContainerName: literal("invoices-replica"),
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureStorageObjectReplicationSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal one-rule policy", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept account references via valueFrom", func() {
			input := minimalSpec()
			input.Spec.SourceStorageAccountId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureStorageAccount,
						Name:      "primary-storage",
						FieldPath: "status.outputs.storage_account_id",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept container references via valueFrom", func() {
			input := minimalSpec()
			input.Spec.Rules[0].SourceContainerName = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureStorageContainer,
						Name:      "invoices",
						FieldPath: "status.outputs.container_name",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept every copy_blobs_created_after form", func() {
			for _, value := range []string{"OnlyNewObjects", "Everything", "2026-01-01T00:00:00Z", "2026-01-01T00:00:00.5+05:30"} {
				input := minimalSpec()
				input.Spec.Rules[0].CopyBlobsCreatedAfter = proto.String(value)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "value %q must be accepted", value)
			}
		})

		ginkgo.It("should accept prefix filters", func() {
			input := minimalSpec()
			input.Spec.Rules[0].PrefixMatch = []string{"invoices/2026", "receipts/"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept multiple rules", func() {
			input := minimalSpec()
			input.Spec.Rules = append(input.Spec.Rules, &AzureStorageObjectReplicationRule{
				SourceContainerName:      literal("receipts"),
				DestinationContainerName: literal("receipts-replica"),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing source account reference", func() {
			input := minimalSpec()
			input.Spec.SourceStorageAccountId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing destination account reference", func() {
			input := minimalSpec()
			input.Spec.DestinationStorageAccountId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a policy with no rules", func() {
			input := minimalSpec()
			input.Spec.Rules = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a rule without a source container", func() {
			input := minimalSpec()
			input.Spec.Rules[0].SourceContainerName = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a rule without a destination container", func() {
			input := minimalSpec()
			input.Spec.Rules[0].DestinationContainerName = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject malformed copy_blobs_created_after values", func() {
			for _, value := range []string{"", "everything", "Never", "2026-01-01", "2026-01-01T00:00:00", "01/01/2026"} {
				input := minimalSpec()
				input.Spec.Rules[0].CopyBlobsCreatedAfter = proto.String(value)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "value %q must be rejected", value)
			}
		})

		ginkgo.It("should reject an empty prefix filter entry", func() {
			input := minimalSpec()
			input.Spec.Rules[0].PrefixMatch = []string{""}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
