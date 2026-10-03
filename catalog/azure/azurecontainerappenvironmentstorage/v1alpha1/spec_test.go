package azurecontainerappenvironmentstoragev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureContainerAppEnvironmentStorageSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureContainerAppEnvironmentStorageSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimalSpec returns a valid SMB storage registration.
func minimalSpec() *AzureContainerAppEnvironmentStorage {
	return &AzureContainerAppEnvironmentStorage{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureContainerAppEnvironmentStorage",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-storage",
		},
		Spec: &AzureContainerAppEnvironmentStorageSpec{
			ContainerAppEnvironmentId: literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.App/managedEnvironments/env"),
			StorageName:               "app-data",
			ShareName:                 literal("shared-files"),
			AccessMode:                AzureContainerAppEnvironmentStorageAccessMode_READ_WRITE,
			AccountName:               literal("mystorageaccount"),
			AccessKey:                 literal("base64key=="),
		},
	}
}

var _ = ginkgo.Describe("AzureContainerAppEnvironmentStorageSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts an SMB registration (account name + access key)", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("accepts an NFS registration (server URL alone)", func() {
			input := minimalSpec()
			input.Spec.AccountName = nil
			input.Spec.AccessKey = nil
			input.Spec.NfsServerUrl = "mystorageaccount.file.core.windows.net"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a read-only access mode", func() {
			input := minimalSpec()
			input.Spec.AccessMode = AzureContainerAppEnvironmentStorageAccessMode_READ_ONLY
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a single-letter storage name", func() {
			input := minimalSpec()
			input.Spec.StorageName = "s"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("rejects a missing environment reference", func() {
			input := minimalSpec()
			input.Spec.ContainerAppEnvironmentId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing storage name", func() {
			input := minimalSpec()
			input.Spec.StorageName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a storage name starting with a digit", func() {
			input := minimalSpec()
			input.Spec.StorageName = "1storage"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a storage name with consecutive hyphens", func() {
			input := minimalSpec()
			input.Spec.StorageName = "app--data"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a storage name longer than 32 characters", func() {
			input := minimalSpec()
			input.Spec.StorageName = "a1234567890123456789012345678901x"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing share name", func() {
			input := minimalSpec()
			input.Spec.ShareName = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an unspecified access mode", func() {
			input := minimalSpec()
			input.Spec.AccessMode = AzureContainerAppEnvironmentStorageAccessMode_azure_container_app_environment_storage_access_mode_unspecified
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects mixing the SMB and NFS paths", func() {
			input := minimalSpec()
			input.Spec.NfsServerUrl = "mystorageaccount.file.core.windows.net"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an SMB registration missing the access key", func() {
			input := minimalSpec()
			input.Spec.AccessKey = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a registration with neither SMB nor NFS", func() {
			input := minimalSpec()
			input.Spec.AccountName = nil
			input.Spec.AccessKey = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
