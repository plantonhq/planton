package gcpcomputeimagev1alpha1

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

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpComputeImageSpec Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func reference(kind catalogkind.CatalogKind, name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name}}}
}

var _ = ginkgo.Describe("GcpComputeImageSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	fromDisk := func() *GcpComputeImage {
		return &GcpComputeImage{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpComputeImage",
			Metadata:   &shared.CatalogObjectMetadata{Name: "web-base-20261001"},
			Spec: &GcpComputeImageSpec{
				SourceDisk: reference(catalogkind.CatalogKind_GcpComputeDisk, "web-build-disk"),
			},
		}
	}

	ginkgo.It("should accept an image from each single source", func() {
		gomega.Expect(validator.Validate(fromDisk())).To(gomega.Succeed())

		msg := fromDisk()
		msg.Spec.SourceDisk = nil
		msg.Spec.SourceImage = literal("projects/debian-cloud/global/images/family/debian-12")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())

		msg = fromDisk()
		msg.Spec.SourceDisk = nil
		msg.Spec.SourceSnapshot = "web-build-snapshot"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())

		msg = fromDisk()
		msg.Spec.SourceDisk = nil
		msg.Spec.RawDisk = &GcpComputeImageRawDisk{Source: "https://storage.googleapis.com/images/web-base.tar.gz", ContainerType: "TAR"}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := fromDisk()
		msg.Spec.ProjectId = literal("images-prod")
		msg.Spec.ImageName = "web-base-20261001"
		msg.Spec.Description = "Hardened web base"
		msg.Spec.Family = "web-base"
		msg.Spec.KmsKey = reference(catalogkind.CatalogKind_GcpKmsKeyHandle, "images-key")
		msg.Spec.KmsKeyServiceAccount = "builder@images-prod.iam.gserviceaccount.com"
		msg.Spec.SourceDiskEncryption = &GcpComputeImageSourceEncryption{KmsKey: literal("projects/p/locations/us/keyRings/r/cryptoKeys/k")}
		msg.Spec.DiskSizeGb = proto.Int64(20)
		msg.Spec.GuestOsFeatures = []string{"UEFI_COMPATIBLE", "SECURE_BOOT", "GVNIC"}
		msg.Spec.Licenses = []string{"https://www.googleapis.com/compute/v1/projects/debian-cloud/global/licenses/debian-12-bookworm"}
		msg.Spec.StorageLocations = []string{"us"}
		msg.Spec.ShieldedInstanceInitialState = &GcpComputeImageShieldedInstanceInitialState{
			Pk:  &GcpComputeImageFileContentBuffer{Content: "cGs=", FileType: "X509"},
			Dbs: []*GcpComputeImageFileContentBuffer{{Content: "ZGI=", FileType: "BIN"}},
		}
		msg.Spec.ResourceManagerTags = map[string]string{"tagKeys/1": "tagValues/2"}
		msg.Spec.Labels = map[string]string{"os": "debian"}
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require exactly one source", func() {
		msg := fromDisk()
		msg.Spec.SourceDisk = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = fromDisk()
		msg.Spec.SourceSnapshot = "web-build-snapshot"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = fromDisk()
		msg.Spec.RawDisk = &GcpComputeImageRawDisk{Source: "gs://images/web.tar.gz"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should accept a source key only with its source", func() {
		msg := fromDisk()
		msg.Spec.SourceImageEncryption = &GcpComputeImageSourceEncryption{KmsKey: literal("projects/p/locations/us/keyRings/r/cryptoKeys/k")}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = fromDisk()
		msg.Spec.SourceSnapshotEncryption = &GcpComputeImageSourceEncryption{KmsKey: literal("projects/p/locations/us/keyRings/r/cryptoKeys/k")}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = fromDisk()
		msg.Spec.SourceDisk = nil
		msg.Spec.SourceSnapshot = "snap"
		msg.Spec.SourceDiskEncryption = &GcpComputeImageSourceEncryption{KmsKey: literal("projects/p/locations/us/keyRings/r/cryptoKeys/k")}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = fromDisk()
		msg.Spec.SourceDiskEncryption = &GcpComputeImageSourceEncryption{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a key service account without a key", func() {
		msg := fromDisk()
		msg.Spec.KmsKeyServiceAccount = "builder@images-prod.iam.gserviceaccount.com"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject malformed names, families, raw disks, and Secure Boot entries", func() {
		for _, name := range []string{"Web-Base", "1web", "web-", "web_base"} {
			msg := fromDisk()
			msg.Spec.ImageName = name
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), name)
			msg = fromDisk()
			msg.Spec.Family = name
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), name)
		}
		msg := fromDisk()
		msg.Spec.SourceDisk = nil
		msg.Spec.RawDisk = &GcpComputeImageRawDisk{Source: "gs://images/web.tar.gz", ContainerType: "ZIP"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = fromDisk()
		msg.Spec.SourceDisk = nil
		msg.Spec.RawDisk = &GcpComputeImageRawDisk{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = fromDisk()
		msg.Spec.ShieldedInstanceInitialState = &GcpComputeImageShieldedInstanceInitialState{Keks: []*GcpComputeImageFileContentBuffer{{Content: "a2Vr", FileType: "PEM"}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = fromDisk()
		msg.Spec.ShieldedInstanceInitialState = &GcpComputeImageShieldedInstanceInitialState{Dbxs: []*GcpComputeImageFileContentBuffer{{FileType: "BIN"}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a non-positive size and an unknown deletion policy", func() {
		msg := fromDisk()
		msg.Spec.DiskSizeGb = proto.Int64(0)
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = fromDisk()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
