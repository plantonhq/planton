package gcpcomputeinstancev1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestGcpComputeInstanceSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GcpComputeInstanceSpec Validation Suite")
}

var _ = Describe("GcpComputeInstanceSpec validations", func() {

	// Helper function to create a StringValueOrRef with a literal value
	strVal := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
		}
	}

	// Helper function to create a minimal valid spec
	makeValidSpec := func() *GcpComputeInstanceSpec {
		return &GcpComputeInstanceSpec{
			ProjectId:   strVal("my-gcp-project"),
			Zone:        "us-central1-a",
			MachineType: "e2-medium",
			BootDisk: &GcpComputeInstanceBootDisk{
				Image: "debian-cloud/debian-11",
			},
			NetworkInterfaces: []*GcpComputeInstanceNetworkInterface{
				{
					Network: strVal("default"),
				},
			},
		}
	}

	Context("Required fields", func() {
		It("accepts a minimal valid spec", func() {
			spec := makeValidSpec()
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects spec with missing project_id", func() {
			spec := makeValidSpec()
			spec.ProjectId = nil
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("rejects spec with missing zone", func() {
			spec := makeValidSpec()
			spec.Zone = ""
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("rejects spec with missing machine_type", func() {
			spec := makeValidSpec()
			spec.MachineType = ""
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("rejects spec with missing boot_disk", func() {
			spec := makeValidSpec()
			spec.BootDisk = nil
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("rejects spec with empty network_interfaces", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces = nil
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("rejects spec with empty network_interfaces list", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces = []*GcpComputeInstanceNetworkInterface{}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Project ID validation (StringValueOrRef)", func() {
		It("accepts project_id with literal value", func() {
			spec := makeValidSpec()
			spec.ProjectId = strVal("my-project-123")
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts project_id with value_from reference", func() {
			spec := makeValidSpec()
			spec.ProjectId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Name:      "my-gcp-project",
						FieldPath: "status.outputs.project_id",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects empty project_id (nil)", func() {
			spec := makeValidSpec()
			spec.ProjectId = nil
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Zone validation", func() {
		It("accepts valid zone format", func() {
			spec := makeValidSpec()
			spec.Zone = "us-west1-b"
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts valid multi-region zone format", func() {
			spec := makeValidSpec()
			spec.Zone = "europe-west2-c"
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects invalid zone format without suffix", func() {
			spec := makeValidSpec()
			spec.Zone = "us-central1"
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("rejects invalid zone format with uppercase", func() {
			spec := makeValidSpec()
			spec.Zone = "US-CENTRAL1-A"
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Machine type validation", func() {
		It("accepts e2-medium machine type", func() {
			spec := makeValidSpec()
			spec.MachineType = "e2-medium"
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts n1-standard-1 machine type", func() {
			spec := makeValidSpec()
			spec.MachineType = "n1-standard-1"
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts n2-highmem-4 machine type", func() {
			spec := makeValidSpec()
			spec.MachineType = "n2-highmem-4"
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects empty machine type", func() {
			spec := makeValidSpec()
			spec.MachineType = ""
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Boot disk validation", func() {
		It("accepts boot disk with valid image", func() {
			spec := makeValidSpec()
			spec.BootDisk = &GcpComputeInstanceBootDisk{
				Image: "ubuntu-os-cloud/ubuntu-2204-lts",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects boot disk with empty image", func() {
			spec := makeValidSpec()
			spec.BootDisk = &GcpComputeInstanceBootDisk{
				Image: "",
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("accepts boot disk with valid size", func() {
			spec := makeValidSpec()
			spec.BootDisk = &GcpComputeInstanceBootDisk{
				Image:  "debian-cloud/debian-11",
				SizeGb: 50,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts boot disk at minimum size (10 GB)", func() {
			spec := makeValidSpec()
			spec.BootDisk = &GcpComputeInstanceBootDisk{
				Image:  "debian-cloud/debian-11",
				SizeGb: 10,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts boot disk at maximum size (65536 GB)", func() {
			spec := makeValidSpec()
			spec.BootDisk = &GcpComputeInstanceBootDisk{
				Image:  "debian-cloud/debian-11",
				SizeGb: 65536,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects boot disk below minimum size (9 GB)", func() {
			spec := makeValidSpec()
			spec.BootDisk = &GcpComputeInstanceBootDisk{
				Image:  "debian-cloud/debian-11",
				SizeGb: 9,
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("rejects boot disk above maximum size (65537 GB)", func() {
			spec := makeValidSpec()
			spec.BootDisk = &GcpComputeInstanceBootDisk{
				Image:  "debian-cloud/debian-11",
				SizeGb: 65537,
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("accepts boot disk with type pd-ssd", func() {
			spec := makeValidSpec()
			spec.BootDisk = &GcpComputeInstanceBootDisk{
				Image:  "debian-cloud/debian-11",
				SizeGb: 20,
				Type:   "pd-ssd",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})

	Context("Network interface validation - CEL", func() {
		It("accepts network interface with network only", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces = []*GcpComputeInstanceNetworkInterface{
				{
					Network: strVal("projects/my-project/global/networks/my-vpc"),
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts network interface with subnetwork only", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces = []*GcpComputeInstanceNetworkInterface{
				{
					Subnetwork: strVal("projects/my-project/regions/us-central1/subnetworks/my-subnet"),
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts network interface with both network and subnetwork", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces = []*GcpComputeInstanceNetworkInterface{
				{
					Network:    strVal("projects/my-project/global/networks/my-vpc"),
					Subnetwork: strVal("projects/my-project/regions/us-central1/subnetworks/my-subnet"),
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects network interface without network or subnetwork (CEL rule)", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces = []*GcpComputeInstanceNetworkInterface{
				{
					// Neither network nor subnetwork specified
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("accepts network interface with value_from reference", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces = []*GcpComputeInstanceNetworkInterface{
				{
					Network: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
							ValueFrom: &foreignkeyv1.ValueFromRef{
								Name:      "my-vpc",
								FieldPath: "status.outputs.network_self_link",
							},
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts multiple network interfaces", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces = []*GcpComputeInstanceNetworkInterface{
				{
					Network: strVal("projects/my-project/global/networks/vpc1"),
				},
				{
					Subnetwork: strVal("projects/my-project/regions/us-central1/subnetworks/subnet2"),
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})

	Context("Access config validation", func() {
		It("accepts access config with PREMIUM network tier", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces = []*GcpComputeInstanceNetworkInterface{
				{
					Network: strVal("default"),
					AccessConfigs: []*GcpComputeInstanceAccessConfig{
						{
							NetworkTier: "PREMIUM",
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts access config with STANDARD network tier", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces = []*GcpComputeInstanceNetworkInterface{
				{
					Network: strVal("default"),
					AccessConfigs: []*GcpComputeInstanceAccessConfig{
						{
							NetworkTier: "STANDARD",
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts access config with empty network tier (defaults)", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces = []*GcpComputeInstanceNetworkInterface{
				{
					Network: strVal("default"),
					AccessConfigs: []*GcpComputeInstanceAccessConfig{
						{
							NetworkTier: "",
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects access config with invalid network tier", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces = []*GcpComputeInstanceNetworkInterface{
				{
					Network: strVal("default"),
					AccessConfigs: []*GcpComputeInstanceAccessConfig{
						{
							NetworkTier: "INVALID",
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Attached disk validation", func() {
		It("accepts attached disk with source", func() {
			spec := makeValidSpec()
			spec.AttachedDisks = []*GcpComputeInstanceAttachedDisk{
				{
					Source: "projects/my-project/zones/us-central1-a/disks/my-disk",
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects attached disk without source", func() {
			spec := makeValidSpec()
			spec.AttachedDisks = []*GcpComputeInstanceAttachedDisk{
				{
					Source: "",
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("accepts attached disk with READ_WRITE mode", func() {
			spec := makeValidSpec()
			spec.AttachedDisks = []*GcpComputeInstanceAttachedDisk{
				{
					Source: "projects/my-project/zones/us-central1-a/disks/my-disk",
					Mode:   "READ_WRITE",
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts attached disk with READ_ONLY mode", func() {
			spec := makeValidSpec()
			spec.AttachedDisks = []*GcpComputeInstanceAttachedDisk{
				{
					Source: "projects/my-project/zones/us-central1-a/disks/my-disk",
					Mode:   "READ_ONLY",
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects attached disk with invalid mode", func() {
			spec := makeValidSpec()
			spec.AttachedDisks = []*GcpComputeInstanceAttachedDisk{
				{
					Source: "projects/my-project/zones/us-central1-a/disks/my-disk",
					Mode:   "INVALID",
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Scheduling validation", func() {
		It("accepts scheduling with MIGRATE on_host_maintenance", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				OnHostMaintenance: "MIGRATE",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts scheduling with TERMINATE on_host_maintenance", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				OnHostMaintenance: "TERMINATE",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects scheduling with invalid on_host_maintenance", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				OnHostMaintenance: "INVALID",
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("accepts scheduling with STANDARD provisioning_model", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				ProvisioningModel: "STANDARD",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts scheduling with SPOT provisioning_model", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				ProvisioningModel: "SPOT",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects scheduling with invalid provisioning_model", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				ProvisioningModel: "INVALID",
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("accepts scheduling with STOP instance_termination_action", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				InstanceTerminationAction: "STOP",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts scheduling with DELETE instance_termination_action", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				InstanceTerminationAction: "DELETE",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects scheduling with invalid instance_termination_action", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				InstanceTerminationAction: "INVALID",
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Service account validation", func() {
		It("accepts service account with email (literal) and scopes", func() {
			spec := makeValidSpec()
			spec.ServiceAccount = &GcpComputeInstanceServiceAccount{
				Email:  strVal("my-sa@my-project.iam.gserviceaccount.com"),
				Scopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts service account with email (value_from reference)", func() {
			spec := makeValidSpec()
			spec.ServiceAccount = &GcpComputeInstanceServiceAccount{
				Email: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name:      "my-service-account",
							FieldPath: "status.outputs.email",
						},
					},
				},
				Scopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts service account with email only", func() {
			spec := makeValidSpec()
			spec.ServiceAccount = &GcpComputeInstanceServiceAccount{
				Email: strVal("my-sa@my-project.iam.gserviceaccount.com"),
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts spec without service account", func() {
			spec := makeValidSpec()
			spec.ServiceAccount = nil
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})

	Context("Optional fields", func() {
		It("accepts spec with metadata", func() {
			spec := makeValidSpec()
			spec.Metadata = map[string]string{
				"foo": "bar",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts spec with labels", func() {
			spec := makeValidSpec()
			spec.Labels = map[string]string{
				"env": "production",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts spec with tags", func() {
			spec := makeValidSpec()
			spec.Tags = []string{"web-server", "http-server"}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts spec with ssh_keys", func() {
			spec := makeValidSpec()
			spec.SshKeys = []string{"user:ssh-rsa AAAAB..."}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts spec with startup_script", func() {
			spec := makeValidSpec()
			spec.StartupScript = "#!/bin/bash\necho 'Hello World'"
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts spec with deletion_protection enabled", func() {
			spec := makeValidSpec()
			spec.DeletionProtection = true
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts spec with preemptible enabled", func() {
			spec := makeValidSpec()
			spec.Preemptible = true
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts spec with spot enabled", func() {
			spec := makeValidSpec()
			spec.Spot = true
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})

	Context("Production-grade configuration example", func() {
		It("accepts a complete production spec with all options (literal values)", func() {
			spec := &GcpComputeInstanceSpec{
				ProjectId:   strVal("prod-project"),
				Zone:        "us-central1-a",
				MachineType: "n2-standard-4",
				BootDisk: &GcpComputeInstanceBootDisk{
					Image:      "debian-cloud/debian-11",
					SizeGb:     100,
					Type:       "pd-ssd",
					AutoDelete: true,
				},
				NetworkInterfaces: []*GcpComputeInstanceNetworkInterface{
					{
						Network:    strVal("projects/prod-project/global/networks/prod-vpc"),
						Subnetwork: strVal("projects/prod-project/regions/us-central1/subnetworks/prod-subnet"),
						AccessConfigs: []*GcpComputeInstanceAccessConfig{
							{
								NetworkTier: "PREMIUM",
							},
						},
					},
				},
				AttachedDisks: []*GcpComputeInstanceAttachedDisk{
					{
						Source: "projects/prod-project/zones/us-central1-a/disks/data-disk",
						Mode:   "READ_WRITE",
					},
				},
				ServiceAccount: &GcpComputeInstanceServiceAccount{
					Email:  strVal("prod-sa@prod-project.iam.gserviceaccount.com"),
					Scopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
				},
				DeletionProtection:     true,
				AllowStoppingForUpdate: true,
				Metadata: map[string]string{
					"enable-oslogin": "TRUE",
				},
				Labels: map[string]string{
					"env":  "production",
					"app":  "webserver",
					"team": "platform",
				},
				Tags:          []string{"web-server", "https-server"},
				StartupScript: "#!/bin/bash\necho 'Starting production server'",
				Scheduling: &GcpComputeInstanceScheduling{
					AutomaticRestart:  true,
					OnHostMaintenance: "MIGRATE",
					ProvisioningModel: "STANDARD",
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts a Spot VM configuration", func() {
			spec := &GcpComputeInstanceSpec{
				ProjectId:   strVal("dev-project"),
				Zone:        "us-central1-a",
				MachineType: "e2-medium",
				BootDisk: &GcpComputeInstanceBootDisk{
					Image:  "debian-cloud/debian-11",
					SizeGb: 20,
				},
				NetworkInterfaces: []*GcpComputeInstanceNetworkInterface{
					{
						Network: strVal("default"),
					},
				},
				Spot: true,
				Scheduling: &GcpComputeInstanceScheduling{
					Preemptible:               true,
					AutomaticRestart:          false,
					OnHostMaintenance:         "TERMINATE",
					ProvisioningModel:         "SPOT",
					InstanceTerminationAction: "STOP",
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})
})
