package azurestoragequeuev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureStorageQueueSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureStorageQueueSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const accountId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/plantonapp"

// minimal valid spec: a work queue.
func minimalSpec() *AzureStorageQueue {
	return &AzureStorageQueue{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureStorageQueue",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-queue",
		},
		Spec: &AzureStorageQueueSpec{
			StorageAccountId: literal(accountId),
			QueueName:        "work-items",
		},
	}
}

var _ = ginkgo.Describe("AzureStorageQueueSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal queue", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept hyphenated names within the length bounds", func() {
			for _, name := range []string{"app", "raw-events", "orders-poison"} {
				input := minimalSpec()
				input.Spec.QueueName = name
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "name %q must be accepted", name)
			}
		})

		ginkgo.It("should accept queue metadata", func() {
			input := minimalSpec()
			input.Spec.Metadata = map[string]string{"purpose": "work-items", "owner": "platform"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a valueFrom reference for the account", func() {
			input := minimalSpec()
			input.Spec.StorageAccountId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureStorageAccount,
						Name:      "app-storage",
						FieldPath: "status.outputs.storage_account_id",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing storage account reference", func() {
			input := minimalSpec()
			input.Spec.StorageAccountId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject names that break the queue naming rules", func() {
			for _, name := range []string{"", "ab", "UPPER", "-leading", "trailing-", "double--hyphen", "has_underscore"} {
				input := minimalSpec()
				input.Spec.QueueName = name
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "name %q must be rejected", name)
			}
		})

		ginkgo.It("should reject a name longer than 63 characters", func() {
			input := minimalSpec()
			input.Spec.QueueName = "a123456789b123456789c123456789d123456789e123456789f123456789g123"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
