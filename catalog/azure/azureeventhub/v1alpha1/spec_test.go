package azureeventhubv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureEventHubSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureEventHubSpec Validation Tests")
}

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

// helper to create a minimal valid hub (simple day-count retention)
func minimalHub() *AzureEventHub {
	retention := int32(1)
	return &AzureEventHub{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureEventHub",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-hub",
		},
		Spec: &AzureEventHubSpec{
			NamespaceId:      literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.EventHub/namespaces/my-ehns"),
			EventHubName:     "telemetry",
			PartitionCount:   4,
			MessageRetention: &retention,
		},
	}
}

// helper for a valid capture destination
func boolPtr(v bool) *bool { return &v }

func captureDestination() *AzureEventHubCaptureDestination {
	return &AzureEventHubCaptureDestination{
		ArchiveNameFormat: "{Namespace}/{EventHub}/{PartitionId}/{Year}/{Month}/{Day}/{Hour}/{Minute}/{Second}",
		BlobContainerName: literal("capture-archives"),
		StorageAccountId:  literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/mystorageacct"),
	}
}

var _ = ginkgo.Describe("AzureEventHubSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_event_hub", func() {

			ginkgo.It("should accept a minimal hub", func() {
				gomega.Expect(protovalidate.Validate(minimalHub())).To(gomega.BeNil())
			})

			ginkgo.It("should accept retention at the dedicated-tier ceiling", func() {
				retention := int32(90)
				input := minimalHub()
				input.Spec.MessageRetention = &retention
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept the dedicated-tier partition ceiling", func() {
				input := minimalHub()
				input.Spec.PartitionCount = 1024
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a DELETE retention description with a window", func() {
				window := int32(168)
				input := minimalHub()
				input.Spec.MessageRetention = nil
				input.Spec.RetentionDescription = &AzureEventHubRetentionDescription{
					CleanupPolicy:        AzureEventHubCleanupPolicy_DELETE,
					RetentionTimeInHours: &window,
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a COMPACT retention description with a tombstone window", func() {
				window := int32(24)
				input := minimalHub()
				input.Spec.MessageRetention = nil
				input.Spec.RetentionDescription = &AzureEventHubRetentionDescription{
					CleanupPolicy:                 AzureEventHubCleanupPolicy_COMPACT,
					TombstoneRetentionTimeInHours: &window,
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept every gate state", func() {
				for _, s := range []AzureEventHubEntityStatus{
					AzureEventHubEntityStatus_ACTIVE,
					AzureEventHubEntityStatus_DISABLED,
					AzureEventHubEntityStatus_SEND_DISABLED,
				} {
					input := minimalHub()
					input.Spec.Status = s
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept capture with the SAS default", func() {
				input := minimalHub()
				input.Spec.CaptureDescription = &AzureEventHubCaptureDescription{
					Enabled:     boolPtr(true),
					Encoding:    AzureEventHubCaptureEncoding_AVRO,
					Destination: captureDestination(),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept capture with cadence dials and deflate encoding", func() {
				interval := int32(600)
				size := int32(104857600)
				skip := true
				input := minimalHub()
				input.Spec.CaptureDescription = &AzureEventHubCaptureDescription{
					Enabled:           boolPtr(true),
					Encoding:          AzureEventHubCaptureEncoding_AVRO_DEFLATE,
					IntervalInSeconds: &interval,
					SizeLimitInBytes:  &size,
					SkipEmptyArchives: &skip,
					Destination:       captureDestination(),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept capture with user-assigned identity auth", func() {
				input := minimalHub()
				destination := captureDestination()
				destination.StorageAuthenticationType = AzureEventHubCaptureStorageAuthenticationType_USER_ASSIGNED
				destination.StorageAuthenticationId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai")
				input.Spec.CaptureDescription = &AzureEventHubCaptureDescription{
					Enabled:     boolPtr(true),
					Encoding:    AzureEventHubCaptureEncoding_AVRO,
					Destination: destination,
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept capture with system-assigned identity auth", func() {
				input := minimalHub()
				destination := captureDestination()
				destination.StorageAuthenticationType = AzureEventHubCaptureStorageAuthenticationType_SYSTEM_ASSIGNED
				input.Spec.CaptureDescription = &AzureEventHubCaptureDescription{
					Enabled:     boolPtr(true),
					Encoding:    AzureEventHubCaptureEncoding_AVRO,
					Destination: destination,
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a paused capture block (enabled false)", func() {
				input := minimalHub()
				input.Spec.CaptureDescription = &AzureEventHubCaptureDescription{
					Enabled:     boolPtr(false),
					Encoding:    AzureEventHubCaptureEncoding_AVRO,
					Destination: captureDestination(),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.Context("event_hub_name", func() {

			ginkgo.It("should reject a missing name", func() {
				input := minimalHub()
				input.Spec.EventHubName = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name ending with a period", func() {
				input := minimalHub()
				input.Spec.EventHubName = "telemetry."
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name with invalid characters", func() {
				input := minimalHub()
				input.Spec.EventHubName = "tele metry!"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("required references", func() {

			ginkgo.It("should reject a missing namespace_id", func() {
				input := minimalHub()
				input.Spec.NamespaceId = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("partition_count", func() {

			ginkgo.It("should reject zero partitions", func() {
				input := minimalHub()
				input.Spec.PartitionCount = 0
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject more than 1024 partitions", func() {
				input := minimalHub()
				input.Spec.PartitionCount = 1025
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("retention model", func() {

			ginkgo.It("should reject NEITHER retention model set", func() {
				input := minimalHub()
				input.Spec.MessageRetention = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject BOTH retention models set", func() {
				window := int32(24)
				input := minimalHub()
				input.Spec.RetentionDescription = &AzureEventHubRetentionDescription{
					CleanupPolicy:        AzureEventHubCleanupPolicy_DELETE,
					RetentionTimeInHours: &window,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject message_retention beyond 90 days", func() {
				retention := int32(91)
				input := minimalHub()
				input.Spec.MessageRetention = &retention
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unspecified cleanup policy", func() {
				window := int32(24)
				input := minimalHub()
				input.Spec.MessageRetention = nil
				input.Spec.RetentionDescription = &AzureEventHubRetentionDescription{
					RetentionTimeInHours: &window,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject DELETE without a retention window", func() {
				input := minimalHub()
				input.Spec.MessageRetention = nil
				input.Spec.RetentionDescription = &AzureEventHubRetentionDescription{
					CleanupPolicy: AzureEventHubCleanupPolicy_DELETE,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject DELETE with a tombstone window", func() {
				window := int32(24)
				tombstone := int32(24)
				input := minimalHub()
				input.Spec.MessageRetention = nil
				input.Spec.RetentionDescription = &AzureEventHubRetentionDescription{
					CleanupPolicy:                 AzureEventHubCleanupPolicy_DELETE,
					RetentionTimeInHours:          &window,
					TombstoneRetentionTimeInHours: &tombstone,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject COMPACT with a retention window", func() {
				window := int32(24)
				input := minimalHub()
				input.Spec.MessageRetention = nil
				input.Spec.RetentionDescription = &AzureEventHubRetentionDescription{
					CleanupPolicy:        AzureEventHubCleanupPolicy_COMPACT,
					RetentionTimeInHours: &window,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("capture_description", func() {

			ginkgo.It("should reject capture without a destination", func() {
				input := minimalHub()
				input.Spec.CaptureDescription = &AzureEventHubCaptureDescription{
					Enabled:  boolPtr(true),
					Encoding: AzureEventHubCaptureEncoding_AVRO,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject capture with an unspecified encoding", func() {
				input := minimalHub()
				input.Spec.CaptureDescription = &AzureEventHubCaptureDescription{
					Enabled:     boolPtr(true),
					Destination: captureDestination(),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an interval outside 60-900 seconds", func() {
				interval := int32(30)
				input := minimalHub()
				input.Spec.CaptureDescription = &AzureEventHubCaptureDescription{
					Enabled:           boolPtr(true),
					Encoding:          AzureEventHubCaptureEncoding_AVRO,
					IntervalInSeconds: &interval,
					Destination:       captureDestination(),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a size limit below 10 MB", func() {
				size := int32(1048576)
				input := minimalHub()
				input.Spec.CaptureDescription = &AzureEventHubCaptureDescription{
					Enabled:          boolPtr(true),
					Encoding:         AzureEventHubCaptureEncoding_AVRO,
					SizeLimitInBytes: &size,
					Destination:      captureDestination(),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an archive_name_format missing tokens", func() {
				input := minimalHub()
				destination := captureDestination()
				destination.ArchiveNameFormat = "{Namespace}/{EventHub}/{Year}"
				input.Spec.CaptureDescription = &AzureEventHubCaptureDescription{
					Enabled:     boolPtr(true),
					Encoding:    AzureEventHubCaptureEncoding_AVRO,
					Destination: destination,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject USER_ASSIGNED auth without an identity", func() {
				input := minimalHub()
				destination := captureDestination()
				destination.StorageAuthenticationType = AzureEventHubCaptureStorageAuthenticationType_USER_ASSIGNED
				input.Spec.CaptureDescription = &AzureEventHubCaptureDescription{
					Enabled:     boolPtr(true),
					Encoding:    AzureEventHubCaptureEncoding_AVRO,
					Destination: destination,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an identity with SAS auth", func() {
				input := minimalHub()
				destination := captureDestination()
				destination.StorageAuthenticationType = AzureEventHubCaptureStorageAuthenticationType_STORAGE_SAS
				destination.StorageAuthenticationId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai")
				input.Spec.CaptureDescription = &AzureEventHubCaptureDescription{
					Enabled:     boolPtr(true),
					Encoding:    AzureEventHubCaptureEncoding_AVRO,
					Destination: destination,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing storage account", func() {
				input := minimalHub()
				destination := captureDestination()
				destination.StorageAccountId = nil
				input.Spec.CaptureDescription = &AzureEventHubCaptureDescription{
					Enabled:     boolPtr(true),
					Encoding:    AzureEventHubCaptureEncoding_AVRO,
					Destination: destination,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
