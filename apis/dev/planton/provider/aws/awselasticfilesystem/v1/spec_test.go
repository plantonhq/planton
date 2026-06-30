package awselasticfilesystemv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAwsElasticFileSystemSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsElasticFileSystemSpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

var _ = ginkgo.Describe("AwsElasticFileSystemSpec validations", func() {
	var spec *AwsElasticFileSystemSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: subnet_ids required.
		spec = &AwsElasticFileSystemSpec{
			Region: "us-east-1",
			SubnetIds: []*foreignkeyv1.StringValueOrRef{
				strRef("subnet-abc123"),
			},
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal valid spec (just subnet_ids)", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts encrypted with default KMS", func() {
		spec.Encrypted = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts encrypted with custom KMS key", func() {
		spec.Encrypted = true
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/test-key")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts One Zone storage (single AZ)", func() {
		spec.AvailabilityZoneName = "us-east-1a"
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{strRef("subnet-abc123")}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts provisioned throughput mode", func() {
		spec.ThroughputMode = "provisioned"
		spec.ProvisionedThroughputInMibps = 100.0
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts elastic throughput mode", func() {
		spec.ThroughputMode = "elastic"
		spec.PerformanceMode = "generalPurpose"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts lifecycle policy (transition_to_ia only)", func() {
		spec.TransitionToIa = "AFTER_30_DAYS"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts lifecycle policy (IA + archive + primary storage class)", func() {
		spec.TransitionToIa = "AFTER_30_DAYS"
		spec.TransitionToArchive = "AFTER_90_DAYS"
		spec.TransitionToPrimaryStorageClass = "AFTER_1_ACCESS"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts backup enabled", func() {
		spec.BackupEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts single access point with POSIX user and root directory", func() {
		spec.AccessPoints = []*AwsElasticFileSystemAccessPoint{
			{
				Name: "app-data",
				PosixUser: &AwsElasticFileSystemAccessPointPosixUser{
					Uid: 1000,
					Gid: 1000,
				},
				RootDirectory: &AwsElasticFileSystemAccessPointRootDirectory{
					Path: "/app/data",
					CreationInfo: &AwsElasticFileSystemAccessPointCreationInfo{
						OwnerUid:    1000,
						OwnerGid:    1000,
						Permissions: "0755",
					},
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts multiple access points", func() {
		spec.AccessPoints = []*AwsElasticFileSystemAccessPoint{
			{
				Name: "app-data",
				PosixUser: &AwsElasticFileSystemAccessPointPosixUser{
					Uid: 1000,
					Gid: 1000,
				},
				RootDirectory: &AwsElasticFileSystemAccessPointRootDirectory{
					Path: "/app/data",
					CreationInfo: &AwsElasticFileSystemAccessPointCreationInfo{
						OwnerUid:    1000,
						OwnerGid:    1000,
						Permissions: "0755",
					},
				},
			},
			{
				Name: "logs",
				PosixUser: &AwsElasticFileSystemAccessPointPosixUser{
					Uid: 1001,
					Gid: 1001,
				},
				RootDirectory: &AwsElasticFileSystemAccessPointRootDirectory{
					Path: "/logs",
					CreationInfo: &AwsElasticFileSystemAccessPointCreationInfo{
						OwnerUid:    1001,
						OwnerGid:    1001,
						Permissions: "0750",
					},
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts production-ready configuration", func() {
		policy, _ := structpb.NewStruct(map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []interface{}{
				map[string]interface{}{
					"Effect":    "Deny",
					"Principal": map[string]interface{}{"AWS": "*"},
					"Action":    "*",
					"Condition": map[string]interface{}{
						"Bool": map[string]interface{}{
							"aws:SecureTransport": "false",
						},
					},
				},
			},
		})
		spec.Encrypted = true
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/test-key")
		spec.PerformanceMode = "generalPurpose"
		spec.ThroughputMode = "elastic"
		spec.TransitionToIa = "AFTER_30_DAYS"
		spec.TransitionToArchive = "AFTER_90_DAYS"
		spec.BackupEnabled = true
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
			strRef("subnet-abc123"),
			strRef("subnet-def456"),
		}
		spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{strRef("sg-123")}
		spec.AccessPoints = []*AwsElasticFileSystemAccessPoint{
			{
				Name: "app-data",
				PosixUser: &AwsElasticFileSystemAccessPointPosixUser{
					Uid: 1000,
					Gid: 1000,
				},
				RootDirectory: &AwsElasticFileSystemAccessPointRootDirectory{
					Path: "/app/data",
					CreationInfo: &AwsElasticFileSystemAccessPointCreationInfo{
						OwnerUid:    1000,
						OwnerGid:    1000,
						Permissions: "0755",
					},
				},
			},
		}
		spec.Policy = policy
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure cases
	// -------------------------------------------------------------------------

	ginkgo.It("fails when subnet_ids is missing", func() {
		spec.SubnetIds = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when subnet_ids is empty", func() {
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when performance_mode is invalid", func() {
		spec.PerformanceMode = "invalid-mode"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when throughput_mode is invalid", func() {
		spec.ThroughputMode = "invalid-throughput"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when provisioned throughput is set without provisioned mode", func() {
		spec.ThroughputMode = "bursting"
		spec.ProvisionedThroughputInMibps = 100.0
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when provisioned mode is set without throughput value", func() {
		spec.ThroughputMode = "provisioned"
		spec.ProvisionedThroughputInMibps = 0.0
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when transition_to_archive is set without transition_to_ia", func() {
		spec.TransitionToArchive = "AFTER_90_DAYS"
		spec.TransitionToIa = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when kms_key_id is set without encrypted", func() {
		spec.Encrypted = false
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/test-key")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when transition_to_ia has invalid value", func() {
		spec.TransitionToIa = "AFTER_2_DAYS"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when transition_to_primary_storage_class has invalid value", func() {
		spec.TransitionToPrimaryStorageClass = "AFTER_2_ACCESS"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
