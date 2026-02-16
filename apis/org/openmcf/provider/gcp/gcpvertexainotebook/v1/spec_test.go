package gcpvertexainotebookv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpVertexAiNotebookSpec Suite")
}

var _ = ginkgo.Describe("GcpVertexAiNotebookSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpVertexAiNotebook.
	minimal := func() *GcpVertexAiNotebook {
		return &GcpVertexAiNotebook{
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpVertexAiNotebook",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-notebook",
			},
			Spec: &GcpVertexAiNotebookSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				Location:    "us-central1-a",
				MachineType: "e2-standard-4",
			},
		}
	}

	// Helper to build a StringValueOrRef with a literal value.
	strRef := func(val string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: val,
			},
		}
	}

	// Suppress unused variable warning.
	_ = strRef

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with instance_name", func() {
		msg := minimal()
		msg.Spec.InstanceName = "my-notebook-instance"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with instance_name containing digits and hyphens", func() {
		msg := minimal()
		msg.Spec.InstanceName = "notebook-v2-test-01"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with desired_state ACTIVE", func() {
		msg := minimal()
		msg.Spec.DesiredState = "ACTIVE"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with desired_state STOPPED", func() {
		msg := minimal()
		msg.Spec.DesiredState = "STOPPED"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with boot_disk configuration", func() {
		msg := minimal()
		msg.Spec.BootDisk = &GcpVertexAiNotebookBootDisk{
			DiskType:   "PD_SSD",
			DiskSizeGb: 200,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept all valid boot disk types", func() {
		types := []string{"PD_STANDARD", "PD_SSD", "PD_BALANCED", "PD_EXTREME"}
		for _, dt := range types {
			msg := minimal()
			msg.Spec.BootDisk = &GcpVertexAiNotebookBootDisk{DiskType: dt}
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "disk_type %s should be valid", dt)
		}
	})

	ginkgo.It("should accept spec with data_disk configuration", func() {
		msg := minimal()
		msg.Spec.DataDisk = &GcpVertexAiNotebookDataDisk{
			DiskType:   "PD_BALANCED",
			DiskSizeGb: 500,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept boot_disk with CMEK encryption", func() {
		msg := minimal()
		msg.Spec.BootDisk = &GcpVertexAiNotebookBootDisk{
			DiskType:   "PD_SSD",
			DiskSizeGb: 200,
			KmsKey:     strRef("projects/p/locations/l/keyRings/kr/cryptoKeys/k"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with GPU accelerator", func() {
		msg := minimal()
		msg.Spec.MachineType = "n1-standard-8"
		msg.Spec.AcceleratorConfig = &GcpVertexAiNotebookAcceleratorConfig{
			Type:      "NVIDIA_TESLA_T4",
			CoreCount: 1,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept all valid accelerator types", func() {
		types := []string{
			"NVIDIA_TESLA_P100", "NVIDIA_TESLA_V100", "NVIDIA_TESLA_P4",
			"NVIDIA_TESLA_T4", "NVIDIA_TESLA_A100", "NVIDIA_A100_80GB",
			"NVIDIA_L4", "NVIDIA_TESLA_T4_VWS", "NVIDIA_TESLA_P100_VWS",
			"NVIDIA_TESLA_P4_VWS",
		}
		for _, at := range types {
			msg := minimal()
			msg.Spec.MachineType = "n1-standard-4"
			msg.Spec.AcceleratorConfig = &GcpVertexAiNotebookAcceleratorConfig{
				Type:      at,
				CoreCount: 1,
			}
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "accelerator type %s should be valid", at)
		}
	})

	ginkgo.It("should accept spec with network interface", func() {
		msg := minimal()
		msg.Spec.NetworkInterface = &GcpVertexAiNotebookNetworkInterface{
			Network: strRef("projects/p/global/networks/my-vpc"),
			Subnet:  strRef("projects/p/regions/us-central1/subnetworks/my-subnet"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with nic_type GVNIC", func() {
		msg := minimal()
		msg.Spec.NetworkInterface = &GcpVertexAiNotebookNetworkInterface{
			Network: strRef("default"),
			NicType: "GVNIC",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with disable_public_ip", func() {
		msg := minimal()
		msg.Spec.DisablePublicIp = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with service_account", func() {
		msg := minimal()
		msg.Spec.ServiceAccount = strRef("sa@my-gcp-project.iam.gserviceaccount.com")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with tags", func() {
		msg := minimal()
		msg.Spec.Tags = []string{"notebook", "ml-team"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with vm_image", func() {
		msg := minimal()
		msg.Spec.VmImage = &GcpVertexAiNotebookVmImage{
			Project: "deeplearning-platform-release",
			Family:  "common-cpu-notebooks",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with container_image", func() {
		msg := minimal()
		msg.Spec.ContainerImage = &GcpVertexAiNotebookContainerImage{
			Repository: "gcr.io/deeplearning-platform-release/base-cu113.py310",
			Tag:        "latest",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with shielded_instance_config", func() {
		msg := minimal()
		msg.Spec.ShieldedInstanceConfig = &GcpVertexAiNotebookShieldedInstanceConfig{
			EnableSecureBoot:          true,
			EnableVtpm:                true,
			EnableIntegrityMonitoring: true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with instance_owners", func() {
		msg := minimal()
		msg.Spec.InstanceOwners = []string{"user@example.com"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with metadata", func() {
		msg := minimal()
		msg.Spec.Metadata = map[string]string{
			"install-monitoring-agent": "true",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with boot and data disk boundary sizes", func() {
		msg := minimal()
		msg.Spec.BootDisk = &GcpVertexAiNotebookBootDisk{DiskSizeGb: 10}
		msg.Spec.DataDisk = &GcpVertexAiNotebookDataDisk{DiskSizeGb: 64000}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept fully configured spec", func() {
		msg := minimal()
		msg.Spec.InstanceName = "full-notebook"
		msg.Spec.DesiredState = "ACTIVE"
		msg.Spec.DisableProxyAccess = false
		msg.Spec.InstanceOwners = []string{"user@example.com"}
		msg.Spec.Metadata = map[string]string{"install-monitoring-agent": "true"}
		msg.Spec.BootDisk = &GcpVertexAiNotebookBootDisk{
			DiskType:   "PD_SSD",
			DiskSizeGb: 200,
			KmsKey:     strRef("projects/p/locations/l/keyRings/kr/cryptoKeys/k"),
		}
		msg.Spec.DataDisk = &GcpVertexAiNotebookDataDisk{
			DiskType:   "PD_BALANCED",
			DiskSizeGb: 500,
			KmsKey:     strRef("projects/p/locations/l/keyRings/kr/cryptoKeys/k"),
		}
		msg.Spec.AcceleratorConfig = &GcpVertexAiNotebookAcceleratorConfig{
			Type:      "NVIDIA_TESLA_T4",
			CoreCount: 1,
		}
		msg.Spec.NetworkInterface = &GcpVertexAiNotebookNetworkInterface{
			Network: strRef("projects/p/global/networks/my-vpc"),
			Subnet:  strRef("projects/p/regions/us-central1/subnetworks/my-subnet"),
			NicType: "GVNIC",
		}
		msg.Spec.DisablePublicIp = true
		msg.Spec.ServiceAccount = strRef("sa@p.iam.gserviceaccount.com")
		msg.Spec.Tags = []string{"notebook", "ml"}
		msg.Spec.VmImage = &GcpVertexAiNotebookVmImage{
			Project: "deeplearning-platform-release",
			Family:  "tf-latest-gpu",
		}
		msg.Spec.ShieldedInstanceConfig = &GcpVertexAiNotebookShieldedInstanceConfig{
			EnableSecureBoot:          true,
			EnableVtpm:                true,
			EnableIntegrityMonitoring: true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject spec without project_id", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec without location", func() {
		msg := minimal()
		msg.Spec.Location = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec without machine_type", func() {
		msg := minimal()
		msg.Spec.MachineType = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with invalid location (region instead of zone)", func() {
		msg := minimal()
		msg.Spec.Location = "us-central1"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with invalid location (uppercase)", func() {
		msg := minimal()
		msg.Spec.Location = "US-CENTRAL1-A"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with invalid instance_name (starts with digit)", func() {
		msg := minimal()
		msg.Spec.InstanceName = "123-notebook"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with invalid instance_name (uppercase)", func() {
		msg := minimal()
		msg.Spec.InstanceName = "My-Notebook"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with invalid instance_name (ends with hyphen)", func() {
		msg := minimal()
		msg.Spec.InstanceName = "notebook-"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with invalid desired_state", func() {
		msg := minimal()
		msg.Spec.DesiredState = "RUNNING"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with invalid boot_disk type", func() {
		msg := minimal()
		msg.Spec.BootDisk = &GcpVertexAiNotebookBootDisk{DiskType: "SSD"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with boot_disk size below minimum", func() {
		msg := minimal()
		msg.Spec.BootDisk = &GcpVertexAiNotebookBootDisk{DiskSizeGb: 5}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with boot_disk size above maximum", func() {
		msg := minimal()
		msg.Spec.BootDisk = &GcpVertexAiNotebookBootDisk{DiskSizeGb: 65000}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with invalid data_disk type", func() {
		msg := minimal()
		msg.Spec.DataDisk = &GcpVertexAiNotebookDataDisk{DiskType: "INVALID"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with data_disk size below minimum", func() {
		msg := minimal()
		msg.Spec.DataDisk = &GcpVertexAiNotebookDataDisk{DiskSizeGb: 3}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with invalid accelerator type", func() {
		msg := minimal()
		msg.Spec.AcceleratorConfig = &GcpVertexAiNotebookAcceleratorConfig{Type: "RTX_4090"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with invalid nic_type", func() {
		msg := minimal()
		msg.Spec.NetworkInterface = &GcpVertexAiNotebookNetworkInterface{NicType: "FAST"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with both vm_image and container_image set", func() {
		msg := minimal()
		msg.Spec.VmImage = &GcpVertexAiNotebookVmImage{
			Project: "deeplearning-platform-release",
			Family:  "common-cpu-notebooks",
		}
		msg.Spec.ContainerImage = &GcpVertexAiNotebookContainerImage{
			Repository: "gcr.io/my-project/my-image",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject container_image without repository", func() {
		msg := minimal()
		msg.Spec.ContainerImage = &GcpVertexAiNotebookContainerImage{
			Repository: "",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with wrong api_version", func() {
		msg := minimal()
		msg.ApiVersion = "wrong.openmcf.org/v1"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with wrong kind", func() {
		msg := minimal()
		msg.Kind = "WrongKind"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec without metadata", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec without spec", func() {
		msg := &GcpVertexAiNotebook{
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpVertexAiNotebook",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-notebook",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
