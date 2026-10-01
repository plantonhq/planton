package gcpcomputediskv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestGcpComputeDiskSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GcpComputeDiskSpec Validation Suite")
}

var _ = Describe("GcpComputeDiskSpec validations", func() {

	strVal := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
		}
	}

	refVal := func(name, fieldPath string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
				ValueFrom: &foreignkeyv1.ValueFromRef{Name: name, FieldPath: fieldPath},
			},
		}
	}

	i32 := func(v int32) *int32 { return &v }
	i64 := func(v int64) *int64 { return &v }

	makeValidSpec := func() *GcpComputeDiskSpec {
		return &GcpComputeDiskSpec{
			ProjectId: strVal("my-gcp-project"),
			Zone:      "us-central1-a",
			SizeGb:    100,
		}
	}

	Context("Required fields and empty-disk semantics", func() {
		It("accepts a minimal empty data disk", func() {
			Expect(protovalidate.Validate(makeValidSpec())).To(BeNil())
		})

		It("rejects a spec with missing zone", func() {
			spec := makeValidSpec()
			spec.Zone = ""
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects an empty disk without size_gb (CEL)", func() {
			spec := makeValidSpec()
			spec.SizeGb = 0
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts a sourced disk without size_gb", func() {
			spec := makeValidSpec()
			spec.SizeGb = 0
			spec.Image = strVal("debian-cloud/debian-12")
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	Context("disk_name", func() {
		It("accepts an empty disk_name (falls back to metadata.name)", func() {
			spec := makeValidSpec()
			spec.DiskName = ""
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts a valid disk_name", func() {
			spec := makeValidSpec()
			spec.DiskName = "pg-data-1"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects an invalid disk_name", func() {
			spec := makeValidSpec()
			spec.DiskName = "-bad"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("zone", func() {
		It("accepts a multi-digit region zone", func() {
			spec := makeValidSpec()
			spec.Zone = "europe-west12-b"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a region", func() {
			spec := makeValidSpec()
			spec.Zone = "us-central1"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("source arms (CEL at-most-one)", func() {
		It("accepts image alone", func() {
			spec := makeValidSpec()
			spec.Image = strVal("debian-cloud/debian-12")
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts a GcpComputeImage reference as the image (satisfies the size rule)", func() {
			spec := makeValidSpec()
			spec.SizeGb = 0
			spec.Image = refVal("web-base", "status.outputs.self_link")
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects an image reference together with source_disk", func() {
			spec := makeValidSpec()
			spec.Image = refVal("web-base", "status.outputs.self_link")
			spec.SourceDisk = refVal("golden-disk", "status.outputs.self_link")
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts source_snapshot alone", func() {
			spec := makeValidSpec()
			spec.SourceSnapshot = "nightly-snap"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts source_disk alone (self-kind clone)", func() {
			spec := makeValidSpec()
			spec.SourceDisk = refVal("golden-disk", "status.outputs.self_link")
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects image together with snapshot", func() {
			spec := makeValidSpec()
			spec.Image = strVal("debian-cloud/debian-12")
			spec.SourceSnapshot = "snap"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects snapshot together with source_disk", func() {
			spec := makeValidSpec()
			spec.SourceSnapshot = "snap"
			spec.SourceDisk = strVal("projects/p/zones/z/disks/d")
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts source_instant_snapshot alone (satisfies the size rule)", func() {
			spec := makeValidSpec()
			spec.SizeGb = 0
			spec.SourceInstantSnapshot = "fast-restore-point"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts source_storage_object alone (satisfies the size rule)", func() {
			spec := makeValidSpec()
			spec.SizeGb = 0
			spec.SourceStorageObject = "https://storage.googleapis.com/imports/disk.vmdk"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects image together with source_instant_snapshot", func() {
			spec := makeValidSpec()
			spec.Image = strVal("debian-cloud/debian-12")
			spec.SourceInstantSnapshot = "fast-restore-point"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects source_storage_object together with source_snapshot", func() {
			spec := makeValidSpec()
			spec.SourceStorageObject = "https://storage.googleapis.com/imports/disk.vmdk"
			spec.SourceSnapshot = "snap"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("size bounds", func() {
		It("accepts bounds", func() {
			spec := makeValidSpec()
			spec.SizeGb = 1
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.SizeGb = 65536
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects above the ceiling", func() {
			spec := makeValidSpec()
			spec.SizeGb = 65537
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("encryption", func() {
		It("accepts a CMEK key reference", func() {
			spec := makeValidSpec()
			spec.KmsKey = refVal("disk-key", "status.outputs.key_id")
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects confidential compute without CMEK (CEL)", func() {
			spec := makeValidSpec()
			spec.Type = "hyperdisk-balanced"
			spec.EnableConfidentialCompute = true
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts confidential compute with CMEK", func() {
			spec := makeValidSpec()
			spec.Type = "hyperdisk-balanced"
			spec.EnableConfidentialCompute = true
			spec.KmsKey = refVal("disk-key", "status.outputs.key_id")
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts CMEK with an explicit encryption service account", func() {
			spec := makeValidSpec()
			spec.KmsKey = refVal("disk-key", "status.outputs.key_id")
			spec.KmsKeyServiceAccount = "kms-agent@prod-project.iam.gserviceaccount.com"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts source_image_encryption together with image", func() {
			spec := makeValidSpec()
			spec.SizeGb = 0
			spec.Image = strVal("projects/p/global/images/encrypted-golden")
			spec.SourceImageEncryption = &GcpComputeDiskSourceEncryption{
				KmsKey: refVal("image-key", "status.outputs.key_id"),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects source_image_encryption without image (CEL)", func() {
			spec := makeValidSpec()
			spec.SourceImageEncryption = &GcpComputeDiskSourceEncryption{
				KmsKey: refVal("image-key", "status.outputs.key_id"),
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects source_snapshot_encryption without source_snapshot (CEL)", func() {
			spec := makeValidSpec()
			spec.SourceSnapshotEncryption = &GcpComputeDiskSourceEncryption{
				KmsKey: refVal("snap-key", "status.outputs.key_id"),
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a source encryption block without its kms_key", func() {
			spec := makeValidSpec()
			spec.SizeGb = 0
			spec.Image = strVal("projects/p/global/images/encrypted-golden")
			spec.SourceImageEncryption = &GcpComputeDiskSourceEncryption{}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("replication, guest features, and deletion policy", func() {
		It("accepts an async replication secondary referencing its primary", func() {
			spec := makeValidSpec()
			spec.AsyncPrimaryDisk = refVal("primary-disk", "status.outputs.self_link")
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts guest OS features and licenses", func() {
			spec := makeValidSpec()
			spec.GuestOsFeatures = []string{"UEFI_COMPATIBLE", "GVNIC"}
			spec.Licenses = []string{"https://www.googleapis.com/compute/v1/projects/windows-cloud/global/licenses/windows-server-core"}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts each valid deletion_policy value", func() {
			for _, policy := range []string{"DELETE", "PREVENT", "ABANDON"} {
				spec := makeValidSpec()
				spec.DeletionPolicy = policy
				Expect(protovalidate.Validate(spec)).To(BeNil())
			}
		})

		It("rejects an invalid deletion_policy", func() {
			spec := makeValidSpec()
			spec.DeletionPolicy = "KEEP"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("performance and layout", func() {
		It("accepts hyperdisk tuning", func() {
			spec := makeValidSpec()
			spec.Type = "hyperdisk-balanced"
			spec.ProvisionedIops = i64(5000)
			spec.ProvisionedThroughput = i64(200)
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects zero provisioned_iops", func() {
			spec := makeValidSpec()
			spec.ProvisionedIops = i64(0)
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("validates physical block sizes", func() {
			spec := makeValidSpec()
			spec.PhysicalBlockSizeBytes = i32(16384)
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.PhysicalBlockSizeBytes = i32(8192)
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("validates access_mode values", func() {
			spec := makeValidSpec()
			spec.AccessMode = "READ_ONLY_MANY"
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.AccessMode = "WRITE_MANY"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("validates architecture values", func() {
			spec := makeValidSpec()
			spec.Architecture = "X86_64"
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.Architecture = "RISCV"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("labels, tags, and destroy safety", func() {
		It("accepts user labels and resource manager tags", func() {
			spec := makeValidSpec()
			spec.Labels = map[string]string{"env": "prod"}
			spec.ResourceManagerTags = map[string]string{"tagKeys/1": "tagValues/2"}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts the snapshot-before-destroy net", func() {
			spec := makeValidSpec()
			spec.CreateSnapshotBeforeDestroy = true
			spec.SnapshotBeforeDestroyPrefix = "final"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	Context("production-grade composition", func() {
		It("accepts a CMEK data volume for a database VM", func() {
			spec := &GcpComputeDiskSpec{
				ProjectId:                   strVal("prod-project"),
				DiskName:                    "pg-data",
				Zone:                        "us-central1-a",
				Description:                 "PostgreSQL data volume",
				Type:                        "pd-ssd",
				SizeGb:                      500,
				KmsKey:                      refVal("disk-key", "status.outputs.key_id"),
				Labels:                      map[string]string{"app": "postgres"},
				CreateSnapshotBeforeDestroy: true,
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})
})
