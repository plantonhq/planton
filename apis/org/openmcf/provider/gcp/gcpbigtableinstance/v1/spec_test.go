package gcpbigtableinstancev1

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
	ginkgo.RunSpecs(t, "GcpBigtableInstanceSpec Suite")
}

var _ = ginkgo.Describe("GcpBigtableInstanceSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpBigtableInstance.
	minimal := func() *GcpBigtableInstance {
		return &GcpBigtableInstance{
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpBigtableInstance",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-bigtable",
			},
			Spec: &GcpBigtableInstanceSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				InstanceName: "my-bigtable-instance",
				Clusters: []*GcpBigtableInstanceCluster{
					{
						ClusterId: "my-cluster-01",
						Zone:      "us-central1-a",
					},
				},
			},
		}
	}

	// Helper for StringValueOrRef with a literal value.
	svr := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
		}
	}

	// ──────────────────────────────────────────────────────────
	//  POSITIVE CASES
	// ──────────────────────────────────────────────────────────

	ginkgo.Context("valid specs", func() {

		ginkgo.It("should accept a minimal valid spec", func() {
			gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
		})

		ginkgo.It("should accept display_name set", func() {
			r := minimal()
			r.Spec.DisplayName = "My Bigtable Dev Instance"
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept deletion_protection explicitly false", func() {
			r := minimal()
			r.Spec.DeletionProtection = proto.Bool(false)
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept deletion_protection explicitly true", func() {
			r := minimal()
			r.Spec.DeletionProtection = proto.Bool(true)
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept force_destroy enabled", func() {
			r := minimal()
			r.Spec.ForceDestroy = true
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept instance_name at minimum length (6 chars)", func() {
			r := minimal()
			r.Spec.InstanceName = "ab1234" // 6 chars
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept instance_name at maximum length (33 chars)", func() {
			r := minimal()
			r.Spec.InstanceName = "a" + strings.Repeat("b", 31) + "c" // 33 chars
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept instance_name with hyphens", func() {
			r := minimal()
			r.Spec.InstanceName = "my-bt-instance-01"
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept cluster_id at minimum length (6 chars)", func() {
			r := minimal()
			r.Spec.Clusters[0].ClusterId = "cl1234" // 6 chars
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept cluster_id at maximum length (30 chars)", func() {
			r := minimal()
			r.Spec.Clusters[0].ClusterId = "a" + strings.Repeat("b", 28) + "c" // 30 chars
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept single cluster with fixed num_nodes", func() {
			r := minimal()
			r.Spec.Clusters[0].NumNodes = 3
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept single cluster with autoscaling", func() {
			r := minimal()
			r.Spec.Clusters[0].AutoscalingConfig = &GcpBigtableInstanceClusterAutoscalingConfig{
				MinNodes:  1,
				MaxNodes:  10,
				CpuTarget: 60,
			}
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept autoscaling with storage_target", func() {
			r := minimal()
			r.Spec.Clusters[0].AutoscalingConfig = &GcpBigtableInstanceClusterAutoscalingConfig{
				MinNodes:      2,
				MaxNodes:      20,
				CpuTarget:     70,
				StorageTarget: 3072,
			}
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept autoscaling cpu_target at lower bound (10)", func() {
			r := minimal()
			r.Spec.Clusters[0].AutoscalingConfig = &GcpBigtableInstanceClusterAutoscalingConfig{
				MinNodes:  1,
				MaxNodes:  5,
				CpuTarget: 10,
			}
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept autoscaling cpu_target at upper bound (80)", func() {
			r := minimal()
			r.Spec.Clusters[0].AutoscalingConfig = &GcpBigtableInstanceClusterAutoscalingConfig{
				MinNodes:  1,
				MaxNodes:  5,
				CpuTarget: 80,
			}
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept storage_type SSD", func() {
			r := minimal()
			r.Spec.Clusters[0].StorageType = proto.String("SSD")
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept storage_type HDD", func() {
			r := minimal()
			r.Spec.Clusters[0].StorageType = proto.String("HDD")
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept node_scaling_factor NodeScalingFactor1X", func() {
			r := minimal()
			r.Spec.Clusters[0].NodeScalingFactor = "NodeScalingFactor1X"
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept node_scaling_factor NodeScalingFactor2X", func() {
			r := minimal()
			r.Spec.Clusters[0].NodeScalingFactor = "NodeScalingFactor2X"
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept cluster with CMEK encryption", func() {
			r := minimal()
			r.Spec.Clusters[0].KmsKeyName = svr("projects/p/locations/us-central1/keyRings/kr/cryptoKeys/k")
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept multiple clusters in different zones", func() {
			r := minimal()
			r.Spec.Clusters = []*GcpBigtableInstanceCluster{
				{
					ClusterId: "cluster-zone-a",
					Zone:      "us-central1-a",
					NumNodes:  3,
				},
				{
					ClusterId: "cluster-zone-b",
					Zone:      "us-central1-b",
					NumNodes:  3,
				},
			}
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept full-featured spec with all optional fields", func() {
			r := minimal()
			r.Spec.DisplayName = "Production Bigtable"
			r.Spec.DeletionProtection = proto.Bool(true)
			r.Spec.ForceDestroy = false
			r.Spec.Clusters = []*GcpBigtableInstanceCluster{
				{
					ClusterId:         "prod-cluster-a",
					Zone:              "us-central1-a",
					StorageType:       proto.String("SSD"),
					NodeScalingFactor: "NodeScalingFactor1X",
					KmsKeyName:        svr("projects/p/locations/us-central1/keyRings/kr/cryptoKeys/k"),
					AutoscalingConfig: &GcpBigtableInstanceClusterAutoscalingConfig{
						MinNodes:      3,
						MaxNodes:      30,
						CpuTarget:     65,
						StorageTarget: 4096,
					},
				},
				{
					ClusterId:         "prod-cluster-b",
					Zone:              "us-central1-b",
					StorageType:       proto.String("SSD"),
					NodeScalingFactor: "NodeScalingFactor1X",
					KmsKeyName:        svr("projects/p/locations/us-central1/keyRings/kr/cryptoKeys/k"),
					AutoscalingConfig: &GcpBigtableInstanceClusterAutoscalingConfig{
						MinNodes:      3,
						MaxNodes:      30,
						CpuTarget:     65,
						StorageTarget: 4096,
					},
				},
			}
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept autoscaling where min_nodes equals max_nodes", func() {
			r := minimal()
			r.Spec.Clusters[0].AutoscalingConfig = &GcpBigtableInstanceClusterAutoscalingConfig{
				MinNodes:  5,
				MaxNodes:  5,
				CpuTarget: 50,
			}
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})

		ginkgo.It("should accept cluster with neither num_nodes nor autoscaling (auto-allocate)", func() {
			r := minimal()
			// Neither num_nodes nor autoscaling_config set; Bigtable auto-allocates.
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		})
	})

	// ──────────────────────────────────────────────────────────
	//  NEGATIVE CASES
	// ──────────────────────────────────────────────────────────

	ginkgo.Context("invalid specs", func() {

		ginkgo.It("should reject missing project_id", func() {
			r := minimal()
			r.Spec.ProjectId = nil
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("project_id"))
		})

		ginkgo.It("should reject missing instance_name", func() {
			r := minimal()
			r.Spec.InstanceName = ""
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("instance_name"))
		})

		ginkgo.It("should reject instance_name too short (5 chars)", func() {
			r := minimal()
			r.Spec.InstanceName = "ab123" // 5 chars
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("instance_name"))
		})

		ginkgo.It("should reject instance_name too long (34 chars)", func() {
			r := minimal()
			r.Spec.InstanceName = "a" + strings.Repeat("b", 32) + "c" // 34 chars
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("instance_name"))
		})

		ginkgo.It("should reject instance_name with uppercase letters", func() {
			r := minimal()
			r.Spec.InstanceName = "My-Instance-01"
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("instance_name"))
		})

		ginkgo.It("should reject instance_name starting with a number", func() {
			r := minimal()
			r.Spec.InstanceName = "1my-instance"
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("instance_name"))
		})

		ginkgo.It("should reject instance_name ending with a hyphen", func() {
			r := minimal()
			r.Spec.InstanceName = "my-instance-"
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("instance_name"))
		})

		ginkgo.It("should reject empty clusters list", func() {
			r := minimal()
			r.Spec.Clusters = []*GcpBigtableInstanceCluster{}
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("clusters"))
		})

		ginkgo.It("should reject missing cluster_id", func() {
			r := minimal()
			r.Spec.Clusters[0].ClusterId = ""
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("cluster_id"))
		})

		ginkgo.It("should reject cluster_id too short (5 chars)", func() {
			r := minimal()
			r.Spec.Clusters[0].ClusterId = "ab123" // 5 chars
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("cluster_id"))
		})

		ginkgo.It("should reject cluster_id too long (31 chars)", func() {
			r := minimal()
			r.Spec.Clusters[0].ClusterId = "a" + strings.Repeat("b", 29) + "c" // 31 chars
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("cluster_id"))
		})

		ginkgo.It("should reject cluster_id with uppercase letters", func() {
			r := minimal()
			r.Spec.Clusters[0].ClusterId = "My-Cluster-01"
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("cluster_id"))
		})

		ginkgo.It("should reject missing cluster zone", func() {
			r := minimal()
			r.Spec.Clusters[0].Zone = ""
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("zone"))
		})

		ginkgo.It("should reject invalid storage_type", func() {
			r := minimal()
			r.Spec.Clusters[0].StorageType = proto.String("NVME")
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("storage_type"))
		})

		ginkgo.It("should reject invalid node_scaling_factor", func() {
			r := minimal()
			r.Spec.Clusters[0].NodeScalingFactor = "NodeScalingFactor3X"
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("node_scaling_factor"))
		})

		ginkgo.It("should reject num_nodes and autoscaling_config both set", func() {
			r := minimal()
			r.Spec.Clusters[0].NumNodes = 3
			r.Spec.Clusters[0].AutoscalingConfig = &GcpBigtableInstanceClusterAutoscalingConfig{
				MinNodes:  1,
				MaxNodes:  10,
				CpuTarget: 60,
			}
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("num_nodes"))
		})

		ginkgo.It("should reject autoscaling with missing min_nodes", func() {
			r := minimal()
			r.Spec.Clusters[0].AutoscalingConfig = &GcpBigtableInstanceClusterAutoscalingConfig{
				MaxNodes:  10,
				CpuTarget: 60,
			}
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("min_nodes"))
		})

		ginkgo.It("should reject autoscaling with missing max_nodes", func() {
			r := minimal()
			r.Spec.Clusters[0].AutoscalingConfig = &GcpBigtableInstanceClusterAutoscalingConfig{
				MinNodes:  1,
				CpuTarget: 60,
			}
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("max_nodes"))
		})

		ginkgo.It("should reject autoscaling with missing cpu_target", func() {
			r := minimal()
			r.Spec.Clusters[0].AutoscalingConfig = &GcpBigtableInstanceClusterAutoscalingConfig{
				MinNodes: 1,
				MaxNodes: 10,
			}
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("cpu_target"))
		})

		ginkgo.It("should reject autoscaling cpu_target below 10", func() {
			r := minimal()
			r.Spec.Clusters[0].AutoscalingConfig = &GcpBigtableInstanceClusterAutoscalingConfig{
				MinNodes:  1,
				MaxNodes:  10,
				CpuTarget: 5,
			}
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("cpu_target"))
		})

		ginkgo.It("should reject autoscaling cpu_target above 80", func() {
			r := minimal()
			r.Spec.Clusters[0].AutoscalingConfig = &GcpBigtableInstanceClusterAutoscalingConfig{
				MinNodes:  1,
				MaxNodes:  10,
				CpuTarget: 95,
			}
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("cpu_target"))
		})

		ginkgo.It("should reject autoscaling min_nodes less than 1", func() {
			r := minimal()
			r.Spec.Clusters[0].AutoscalingConfig = &GcpBigtableInstanceClusterAutoscalingConfig{
				MinNodes:  0,
				MaxNodes:  10,
				CpuTarget: 60,
			}
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("min_nodes"))
		})

		ginkgo.It("should reject autoscaling max_nodes less than min_nodes", func() {
			r := minimal()
			r.Spec.Clusters[0].AutoscalingConfig = &GcpBigtableInstanceClusterAutoscalingConfig{
				MinNodes:  10,
				MaxNodes:  5,
				CpuTarget: 60,
			}
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("max_nodes"))
		})

		ginkgo.It("should reject wrong api_version", func() {
			r := minimal()
			r.ApiVersion = "wrong/v1"
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("api_version"))
		})

		ginkgo.It("should reject wrong kind", func() {
			r := minimal()
			r.Kind = "WrongKind"
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("kind"))
		})

		ginkgo.It("should reject missing metadata", func() {
			r := minimal()
			r.Metadata = nil
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("metadata"))
		})

		ginkgo.It("should reject missing spec", func() {
			r := minimal()
			r.Spec = nil
			err := validator.Validate(r)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("spec"))
		})
	})
})
