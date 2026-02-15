package gcpfilestoreinstancev1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpFilestoreInstanceSpec Suite")
}

var _ = ginkgo.Describe("GcpFilestoreInstanceSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpFilestoreInstance.
	minimal := func() *GcpFilestoreInstance {
		return &GcpFilestoreInstance{
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpFilestoreInstance",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-filestore",
			},
			Spec: &GcpFilestoreInstanceSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				InstanceName: "my-nfs-server",
				Location:     "us-central1-a",
				Tier:         "BASIC_SSD",
				FileShare: &GcpFilestoreInstanceFileShare{
					Name:       "vol1",
					CapacityGb: 2560,
				},
				NetworkConfig: &GcpFilestoreInstanceNetworkConfig{
					Network: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
							Value: "default",
						},
					},
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid BASIC_SSD spec", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept STANDARD tier", func() {
		msg := minimal()
		msg.Spec.Tier = "STANDARD"
		msg.Spec.FileShare.CapacityGb = 1024
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept PREMIUM tier", func() {
		msg := minimal()
		msg.Spec.Tier = "PREMIUM"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept BASIC_HDD tier", func() {
		msg := minimal()
		msg.Spec.Tier = "BASIC_HDD"
		msg.Spec.FileShare.CapacityGb = 1024
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept HIGH_SCALE_SSD tier", func() {
		msg := minimal()
		msg.Spec.Tier = "HIGH_SCALE_SSD"
		msg.Spec.FileShare.CapacityGb = 10240
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept ZONAL tier", func() {
		msg := minimal()
		msg.Spec.Tier = "ZONAL"
		msg.Spec.FileShare.CapacityGb = 1024
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept REGIONAL tier", func() {
		msg := minimal()
		msg.Spec.Tier = "REGIONAL"
		msg.Spec.Location = "us-central1"
		msg.Spec.FileShare.CapacityGb = 1024
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept ENTERPRISE tier", func() {
		msg := minimal()
		msg.Spec.Tier = "ENTERPRISE"
		msg.Spec.Location = "us-central1"
		msg.Spec.FileShare.CapacityGb = 1024
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept instance_name at minimum boundary (2 chars)", func() {
		msg := minimal()
		msg.Spec.InstanceName = "ab"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept instance_name at maximum boundary (63 chars)", func() {
		msg := minimal()
		msg.Spec.InstanceName = "a" + strings.Repeat("b", 61) + "c"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept instance_name with hyphens and numbers", func() {
		msg := minimal()
		msg.Spec.InstanceName = "my-nfs-01"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept file share name with uppercase letters", func() {
		msg := minimal()
		msg.Spec.FileShare.Name = "MyShare"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept file share name with underscores", func() {
		msg := minimal()
		msg.Spec.FileShare.Name = "vol_data_01"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept file share name at 16 char boundary", func() {
		msg := minimal()
		msg.Spec.FileShare.Name = "a" + strings.Repeat("b", 15)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept single-char file share name", func() {
		msg := minimal()
		msg.Spec.FileShare.Name = "v"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept NFS_V3 protocol", func() {
		msg := minimal()
		msg.Spec.Protocol = "NFS_V3"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept NFS_V4_1 protocol", func() {
		msg := minimal()
		msg.Spec.Protocol = "NFS_V4_1"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty protocol (defaults to NFS_V3)", func() {
		msg := minimal()
		msg.Spec.Protocol = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept DIRECT_PEERING connect mode", func() {
		msg := minimal()
		msg.Spec.NetworkConfig.ConnectMode = "DIRECT_PEERING"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept PRIVATE_SERVICE_ACCESS connect mode", func() {
		msg := minimal()
		msg.Spec.NetworkConfig.ConnectMode = "PRIVATE_SERVICE_ACCESS"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept PRIVATE_SERVICE_CONNECT connect mode", func() {
		msg := minimal()
		msg.Spec.NetworkConfig.ConnectMode = "PRIVATE_SERVICE_CONNECT"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept NFS export options with READ_ONLY access", func() {
		msg := minimal()
		msg.Spec.FileShare.NfsExportOptions = []*GcpFilestoreInstanceNfsExportOption{
			{
				IpRanges:   []string{"10.0.0.0/24"},
				AccessMode: "READ_ONLY",
				SquashMode: "ROOT_SQUASH",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept NFS export options with READ_WRITE access", func() {
		msg := minimal()
		msg.Spec.FileShare.NfsExportOptions = []*GcpFilestoreInstanceNfsExportOption{
			{
				IpRanges:   []string{"10.0.0.0/8"},
				AccessMode: "READ_WRITE",
				SquashMode: "NO_ROOT_SQUASH",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept CMEK encryption", func() {
		msg := minimal()
		msg.Spec.KmsKeyName = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "projects/my-proj/locations/us-central1/keyRings/kr/cryptoKeys/key1",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept deletion protection enabled", func() {
		msg := minimal()
		msg.Spec.DeletionProtectionEnabled = true
		msg.Spec.DeletionProtectionReason = "production instance"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept description field", func() {
		msg := minimal()
		msg.Spec.Description = "Shared NFS storage for rendering pipeline"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept reserved IP range", func() {
		msg := minimal()
		msg.Spec.NetworkConfig.ReservedIpRange = "10.0.1.0/29"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept performance config with fixed IOPS", func() {
		msg := minimal()
		msg.Spec.Tier = "ZONAL"
		msg.Spec.FileShare.CapacityGb = 1024
		msg.Spec.PerformanceConfig = &GcpFilestoreInstancePerformanceConfig{
			FixedIops: &GcpFilestoreInstanceFixedIops{
				MaxIops: 12000,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept performance config with IOPS per TB", func() {
		msg := minimal()
		msg.Spec.Tier = "ZONAL"
		msg.Spec.FileShare.CapacityGb = 1024
		msg.Spec.PerformanceConfig = &GcpFilestoreInstancePerformanceConfig{
			IopsPerTb: &GcpFilestoreInstanceIopsPerTb{
				MaxIopsPerTb: 17000,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept full-featured spec", func() {
		anonUid := int32(1000)
		anonGid := int32(1000)
		msg := minimal()
		msg.Spec.Tier = "ENTERPRISE"
		msg.Spec.Location = "us-central1"
		msg.Spec.Description = "Enterprise NFS for production workloads"
		msg.Spec.Protocol = "NFS_V3"
		msg.Spec.DeletionProtectionEnabled = true
		msg.Spec.DeletionProtectionReason = "critical production data"
		msg.Spec.KmsKeyName = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "projects/p/locations/us-central1/keyRings/kr/cryptoKeys/k",
			},
		}
		msg.Spec.FileShare.Name = "enterprise_vol"
		msg.Spec.FileShare.CapacityGb = 10240
		msg.Spec.FileShare.NfsExportOptions = []*GcpFilestoreInstanceNfsExportOption{
			{
				IpRanges:   []string{"10.0.0.0/24"},
				AccessMode: "READ_WRITE",
				SquashMode: "ROOT_SQUASH",
				AnonUid:    &anonUid,
				AnonGid:    &anonGid,
			},
			{
				IpRanges:   []string{"10.1.0.0/24"},
				AccessMode: "READ_ONLY",
				SquashMode: "NO_ROOT_SQUASH",
			},
		}
		msg.Spec.NetworkConfig.ConnectMode = "PRIVATE_SERVICE_ACCESS"
		msg.Spec.NetworkConfig.ReservedIpRange = "10.10.0.0/29"
		msg.Spec.PerformanceConfig = &GcpFilestoreInstancePerformanceConfig{
			FixedIops: &GcpFilestoreInstanceFixedIops{
				MaxIops: 20000,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject missing project_id", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("project_id"))
	})

	ginkgo.It("should reject missing instance_name", func() {
		msg := minimal()
		msg.Spec.InstanceName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name starting with digit", func() {
		msg := minimal()
		msg.Spec.InstanceName = "1bad-name"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name with uppercase", func() {
		msg := minimal()
		msg.Spec.InstanceName = "Bad-Name"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name ending with hyphen", func() {
		msg := minimal()
		msg.Spec.InstanceName = "bad-name-"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name exceeding 63 chars", func() {
		msg := minimal()
		msg.Spec.InstanceName = "a" + strings.Repeat("b", 63)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing location", func() {
		msg := minimal()
		msg.Spec.Location = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing tier", func() {
		msg := minimal()
		msg.Spec.Tier = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid tier", func() {
		msg := minimal()
		msg.Spec.Tier = "ULTRA"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid protocol", func() {
		msg := minimal()
		msg.Spec.Protocol = "NFS_V5"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("protocol"))
	})

	ginkgo.It("should reject invalid connect_mode", func() {
		msg := minimal()
		msg.Spec.NetworkConfig.ConnectMode = "PUBLIC"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("connect_mode"))
	})

	ginkgo.It("should reject missing file_share", func() {
		msg := minimal()
		msg.Spec.FileShare = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("file_share"))
	})

	ginkgo.It("should reject missing file share name", func() {
		msg := minimal()
		msg.Spec.FileShare.Name = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject file share name starting with digit", func() {
		msg := minimal()
		msg.Spec.FileShare.Name = "1vol"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject file share name with hyphens", func() {
		msg := minimal()
		msg.Spec.FileShare.Name = "my-vol"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject file share name exceeding 16 chars", func() {
		msg := minimal()
		msg.Spec.FileShare.Name = "a" + strings.Repeat("b", 16)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject file share capacity below 1024 GiB", func() {
		msg := minimal()
		msg.Spec.FileShare.CapacityGb = 512
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject zero file share capacity", func() {
		msg := minimal()
		msg.Spec.FileShare.CapacityGb = 0
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing network_config", func() {
		msg := minimal()
		msg.Spec.NetworkConfig = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("network_config"))
	})

	ginkgo.It("should reject missing network in network_config", func() {
		msg := minimal()
		msg.Spec.NetworkConfig.Network = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid access_mode in NFS export options", func() {
		msg := minimal()
		msg.Spec.FileShare.NfsExportOptions = []*GcpFilestoreInstanceNfsExportOption{
			{
				AccessMode: "EXECUTE",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("access_mode"))
	})

	ginkgo.It("should reject invalid squash_mode in NFS export options", func() {
		msg := minimal()
		msg.Spec.FileShare.NfsExportOptions = []*GcpFilestoreInstanceNfsExportOption{
			{
				SquashMode: "ALL_SQUASH",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("squash_mode"))
	})

	ginkgo.It("should reject performance config with both fixed_iops and iops_per_tb", func() {
		msg := minimal()
		msg.Spec.PerformanceConfig = &GcpFilestoreInstancePerformanceConfig{
			FixedIops: &GcpFilestoreInstanceFixedIops{
				MaxIops: 12000,
			},
			IopsPerTb: &GcpFilestoreInstanceIopsPerTb{
				MaxIopsPerTb: 17000,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("mutually exclusive"))
	})

	ginkgo.It("should reject fixed_iops below 1000", func() {
		msg := minimal()
		msg.Spec.PerformanceConfig = &GcpFilestoreInstancePerformanceConfig{
			FixedIops: &GcpFilestoreInstanceFixedIops{
				MaxIops: 500,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject iops_per_tb below 1", func() {
		msg := minimal()
		msg.Spec.PerformanceConfig = &GcpFilestoreInstancePerformanceConfig{
			IopsPerTb: &GcpFilestoreInstanceIopsPerTb{
				MaxIopsPerTb: 0,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing spec entirely", func() {
		msg := minimal()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing metadata", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong api_version", func() {
		msg := minimal()
		msg.ApiVersion = "wrong/v1"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong kind", func() {
		msg := minimal()
		msg.Kind = "WrongKind"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})

// Ensure proto.Message interface is satisfied (compilation check).
var _ proto.Message = (*GcpFilestoreInstance)(nil)
