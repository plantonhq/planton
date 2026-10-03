package azuremachinelearningdatastorev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMachineLearningDatastoreSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMachineLearningDatastoreSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testWorkspaceId  = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/ml-workspace"
	testContainerId  = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/mlstorage/blobServices/default/containers/training-data"
	testFilesystemId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/mldatalake/blobServices/default/containers/lake-fs"
	testShareId      = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/mlstorage/fileServices/default/shares/training-files"
)

// validResource returns a minimal valid blob-variant datastore that
// individual cases mutate into the shape under test.
func validResource() *AzureMachineLearningDatastore {
	return &AzureMachineLearningDatastore{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMachineLearningDatastore",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-ml-datastore",
		},
		Spec: &AzureMachineLearningDatastoreSpec{
			WorkspaceId: literal(testWorkspaceId),
			Name:        "training_data",
			BlobStorage: &AzureMachineLearningDatastoreBlobStorage{
				StorageContainerId: literal(testContainerId),
				AccountKey:         literal("account-key-value"),
			},
		},
	}
}

var _ = ginkgo.Describe("AzureMachineLearningDatastoreSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_machine_learning_datastore", func() {

			ginkgo.It("should not return a validation error for a blob datastore with an account key", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a blob datastore with a SAS token set as default", func() {
				input := validResource()
				input.Spec.BlobStorage.AccountKey = nil
				input.Spec.BlobStorage.SharedAccessSignature = literal("sv=2024-01-01&sig=abc")
				input.Spec.BlobStorage.IsDefault = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a credential-free blob datastore under workspace identity", func() {
				input := validResource()
				input.Spec.BlobStorage.AccountKey = nil
				input.Spec.ServiceDataIdentity = AzureMachineLearningDatastoreServiceDataIdentity_WORKSPACE_SYSTEM_ASSIGNED_IDENTITY
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a data lake datastore with the full service-principal triad", func() {
				input := validResource()
				input.Spec.BlobStorage = nil
				input.Spec.DataLakeGen2 = &AzureMachineLearningDatastoreDataLakeGen2{
					StorageContainerId: literal(testFilesystemId),
					TenantId:           "d67d43c0-6b17-4d4a-8e5e-3b8a76f1f1a1",
					ClientId:           "b6870b46-2b6d-4b0e-8bbd-7a4d81f1c1b2",
					ClientSecret:       literal("sp-secret-value"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a credential-free data lake datastore", func() {
				input := validResource()
				input.Spec.BlobStorage = nil
				input.Spec.DataLakeGen2 = &AzureMachineLearningDatastoreDataLakeGen2{
					StorageContainerId: literal(testFilesystemId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a file share datastore with an account key", func() {
				input := validResource()
				input.Spec.BlobStorage = nil
				input.Spec.FileShare = &AzureMachineLearningDatastoreFileShare{
					StorageFileshareId: literal(testShareId),
					AccountKey:         literal("account-key-value"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_machine_learning_datastore", func() {

			ginkgo.It("should reject a missing workspace reference", func() {
				input := validResource()
				input.Spec.WorkspaceId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name with a leading hyphen", func() {
				input := validResource()
				input.Spec.Name = "-training-data"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name with a dot", func() {
				input := validResource()
				input.Spec.Name = "training.data"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a datastore without any variant block", func() {
				input := validResource()
				input.Spec.BlobStorage = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a datastore with two variant blocks", func() {
				input := validResource()
				input.Spec.FileShare = &AzureMachineLearningDatastoreFileShare{
					StorageFileshareId: literal(testShareId),
					AccountKey:         literal("account-key-value"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a credential-free blob datastore when identity is NONE", func() {
				input := validResource()
				input.Spec.BlobStorage.AccountKey = nil
				input.Spec.ServiceDataIdentity = AzureMachineLearningDatastoreServiceDataIdentity_NONE
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a credential-free blob datastore when identity is unset", func() {
				input := validResource()
				input.Spec.BlobStorage.AccountKey = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a blob datastore without a container reference", func() {
				input := validResource()
				input.Spec.BlobStorage.StorageContainerId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a partial service-principal triad on the data lake variant", func() {
				input := validResource()
				input.Spec.BlobStorage = nil
				input.Spec.DataLakeGen2 = &AzureMachineLearningDatastoreDataLakeGen2{
					StorageContainerId: literal(testFilesystemId),
					TenantId:           "d67d43c0-6b17-4d4a-8e5e-3b8a76f1f1a1",
					ClientId:           "b6870b46-2b6d-4b0e-8bbd-7a4d81f1c1b2",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a non-UUID tenant id on the data lake variant", func() {
				input := validResource()
				input.Spec.BlobStorage = nil
				input.Spec.DataLakeGen2 = &AzureMachineLearningDatastoreDataLakeGen2{
					StorageContainerId: literal(testFilesystemId),
					TenantId:           "not-a-uuid",
					ClientId:           "b6870b46-2b6d-4b0e-8bbd-7a4d81f1c1b2",
					ClientSecret:       literal("sp-secret-value"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a file share datastore with both credentials", func() {
				input := validResource()
				input.Spec.BlobStorage = nil
				input.Spec.FileShare = &AzureMachineLearningDatastoreFileShare{
					StorageFileshareId:    literal(testShareId),
					AccountKey:            literal("account-key-value"),
					SharedAccessSignature: literal("sv=2024-01-01&sig=abc"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a file share datastore with no credentials", func() {
				input := validResource()
				input.Spec.BlobStorage = nil
				input.Spec.FileShare = &AzureMachineLearningDatastoreFileShare{
					StorageFileshareId: literal(testShareId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
