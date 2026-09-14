package gcpcomputeinstancev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestGcpComputeInstanceSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GcpComputeInstanceSpec Validation Suite")
}

var _ = Describe("GcpComputeInstanceSpec validations", func() {

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
	boolPtr := func(v bool) *bool { return &v }

	makeValidSpec := func() *GcpComputeInstanceSpec {
		return &GcpComputeInstanceSpec{
			ProjectId:   strVal("my-gcp-project"),
			Zone:        "us-central1-a",
			MachineType: "e2-medium",
			BootDisk: &GcpComputeInstanceBootDisk{
				Image: "debian-cloud/debian-12",
			},
			NetworkInterfaces: []*GcpComputeInstanceNetworkInterface{
				{Network: strVal("default")},
			},
		}
	}

	Context("Required fields", func() {
		It("accepts a minimal valid spec", func() {
			Expect(protovalidate.Validate(makeValidSpec())).To(BeNil())
		})

		It("rejects a spec with missing zone", func() {
			spec := makeValidSpec()
			spec.Zone = ""
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a spec with missing machine_type", func() {
			spec := makeValidSpec()
			spec.MachineType = ""
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a spec with missing boot_disk", func() {
			spec := makeValidSpec()
			spec.BootDisk = nil
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a spec with no network interfaces", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces = nil
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("instance_name", func() {
		It("accepts an empty instance_name (falls back to metadata.name)", func() {
			spec := makeValidSpec()
			spec.InstanceName = ""
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts a valid instance_name", func() {
			spec := makeValidSpec()
			spec.InstanceName = "web-server-1"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects an instance_name starting with a digit", func() {
			spec := makeValidSpec()
			spec.InstanceName = "1web"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects an instance_name with uppercase", func() {
			spec := makeValidSpec()
			spec.InstanceName = "Web"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects an instance_name ending with a hyphen", func() {
			spec := makeValidSpec()
			spec.InstanceName = "web-"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("zone", func() {
		It("accepts a multi-digit region zone", func() {
			spec := makeValidSpec()
			spec.Zone = "europe-west12-a"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a region (missing zone suffix)", func() {
			spec := makeValidSpec()
			spec.Zone = "us-central1"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("hostname", func() {
		It("accepts a fully qualified hostname", func() {
			spec := makeValidSpec()
			spec.Hostname = "db-1.prod.internal"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a single-label hostname", func() {
			spec := makeValidSpec()
			spec.Hostname = "db-1"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("boot_disk source arms (CEL)", func() {
		It("accepts image as the sole source", func() {
			spec := makeValidSpec()
			spec.BootDisk = &GcpComputeInstanceBootDisk{Image: "debian-cloud/debian-12"}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts source_snapshot as the sole source", func() {
			spec := makeValidSpec()
			spec.BootDisk = &GcpComputeInstanceBootDisk{SourceSnapshot: "my-snapshot"}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts source_disk as the sole source", func() {
			spec := makeValidSpec()
			spec.BootDisk = &GcpComputeInstanceBootDisk{
				SourceDisk: refVal("golden-boot-disk", "status.outputs.self_link"),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a boot disk with no source", func() {
			spec := makeValidSpec()
			spec.BootDisk = &GcpComputeInstanceBootDisk{}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a boot disk with both image and snapshot", func() {
			spec := makeValidSpec()
			spec.BootDisk = &GcpComputeInstanceBootDisk{
				Image:          "debian-cloud/debian-12",
				SourceSnapshot: "my-snapshot",
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a boot disk with image and source_disk", func() {
			spec := makeValidSpec()
			spec.BootDisk = &GcpComputeInstanceBootDisk{
				Image:      "debian-cloud/debian-12",
				SourceDisk: strVal("projects/p/zones/us-central1-a/disks/d"),
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("boot_disk fields", func() {
		It("accepts size bounds (API floor is 1 GB)", func() {
			spec := makeValidSpec()
			spec.BootDisk.SizeGb = 1
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.BootDisk.SizeGb = 65536
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects sizes outside bounds", func() {
			spec := makeValidSpec()
			spec.BootDisk.SizeGb = -1
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
			spec.BootDisk.SizeGb = 65537
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts a CMEK kms_key reference", func() {
			spec := makeValidSpec()
			spec.BootDisk.KmsKey = refVal("boot-key", "status.outputs.key_id")
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts hyperdisk tuning fields", func() {
			spec := makeValidSpec()
			spec.BootDisk.Type = "hyperdisk-balanced"
			spec.BootDisk.ProvisionedIops = i64(3000)
			spec.BootDisk.ProvisionedThroughput = i64(140)
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a zero provisioned_iops", func() {
			spec := makeValidSpec()
			spec.BootDisk.ProvisionedIops = i64(0)
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts valid architectures and rejects invalid ones", func() {
			spec := makeValidSpec()
			spec.BootDisk.Architecture = "ARM64"
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.BootDisk.Architecture = "SPARC"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects more than one boot-disk resource policy", func() {
			spec := makeValidSpec()
			spec.BootDisk.ResourcePolicies = []string{"policy-a", "policy-b"}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("validates mode and interface values", func() {
			spec := makeValidSpec()
			spec.BootDisk.Mode = "READ_ONLY"
			spec.BootDisk.Interface = "NVME"
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.BootDisk.Mode = "WRITE_ONLY"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
			spec.BootDisk.Mode = ""
			spec.BootDisk.Interface = "IDE"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts guest OS features and force_attach", func() {
			spec := makeValidSpec()
			spec.BootDisk.GuestOsFeatures = []string{"UEFI_COMPATIBLE", "GVNIC"}
			spec.BootDisk.ForceAttach = true
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("requires exactly two replica zones when set (CEL)", func() {
			spec := makeValidSpec()
			// Regional boot disks are legal only from a snapshot source —
			// GCP rejects creating one from an image (live-verified 400).
			spec.BootDisk.Image = ""
			spec.BootDisk.SourceSnapshot = "projects/my-project/global/snapshots/base-snap"
			spec.BootDisk.ReplicaZones = []string{"us-central1-a", "us-central1-b"}
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.BootDisk.ReplicaZones = []string{"us-central1-a"}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
			spec.BootDisk.ReplicaZones = []string{"us-central1-a", "us-central1-b", "us-central1-c"}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects replica zones on an image-sourced boot disk (CEL)", func() {
			spec := makeValidSpec()
			spec.BootDisk.ReplicaZones = []string{"us-central1-a", "us-central1-b"}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("replica_zones requires a source_snapshot"))
		})

		It("accepts boot-disk resource manager tags and a kms service account", func() {
			spec := makeValidSpec()
			spec.BootDisk.KmsKey = refVal("boot-key", "status.outputs.key_id")
			spec.BootDisk.KmsKeyServiceAccount = "cmek-agent@p.iam.gserviceaccount.com"
			spec.BootDisk.ResourceManagerTags = map[string]string{"tagKeys/123": "tagValues/456"}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts source_image_encryption together with image (CEL)", func() {
			spec := makeValidSpec()
			spec.BootDisk.SourceImageEncryption = &GcpComputeInstanceSourceEncryption{
				KmsKey: refVal("image-key", "status.outputs.key_id"),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects source_image_encryption without image (CEL)", func() {
			spec := makeValidSpec()
			spec.BootDisk = &GcpComputeInstanceBootDisk{
				SourceSnapshot: "my-snapshot",
				SourceImageEncryption: &GcpComputeInstanceSourceEncryption{
					KmsKey: strVal("projects/p/locations/l/keyRings/r/cryptoKeys/k"),
				},
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects source_snapshot_encryption without source_snapshot (CEL)", func() {
			spec := makeValidSpec()
			spec.BootDisk.SourceSnapshotEncryption = &GcpComputeInstanceSourceEncryption{
				KmsKey: strVal("projects/p/locations/l/keyRings/r/cryptoKeys/k"),
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("requires kms_key inside a source encryption block", func() {
			spec := makeValidSpec()
			spec.BootDisk.SourceImageEncryption = &GcpComputeInstanceSourceEncryption{}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("attached_disks", func() {
		It("accepts a disk referenced from a GcpComputeDisk", func() {
			spec := makeValidSpec()
			spec.AttachedDisks = []*GcpComputeInstanceAttachedDisk{
				{Source: refVal("data-disk", "status.outputs.self_link")},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts a literal disk self link with mode", func() {
			spec := makeValidSpec()
			spec.AttachedDisks = []*GcpComputeInstanceAttachedDisk{
				{
					Source: strVal("projects/p/zones/us-central1-a/disks/data"),
					Mode:   "READ_ONLY",
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a disk without source", func() {
			spec := makeValidSpec()
			spec.AttachedDisks = []*GcpComputeInstanceAttachedDisk{{}}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects an invalid mode", func() {
			spec := makeValidSpec()
			spec.AttachedDisks = []*GcpComputeInstanceAttachedDisk{
				{Source: strVal("d"), Mode: "WRITE_ONLY"},
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts a CMEK service account and force_attach", func() {
			spec := makeValidSpec()
			spec.AttachedDisks = []*GcpComputeInstanceAttachedDisk{
				{
					Source:               refVal("data-disk", "status.outputs.self_link"),
					KmsKey:               refVal("data-key", "status.outputs.key_id"),
					KmsKeyServiceAccount: "cmek-agent@p.iam.gserviceaccount.com",
					ForceAttach:          true,
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	Context("scratch_disks", func() {
		It("accepts an NVME scratch disk", func() {
			spec := makeValidSpec()
			spec.ScratchDisks = []*GcpComputeInstanceScratchDisk{{Interface: "NVME"}}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts explicit valid sizes", func() {
			spec := makeValidSpec()
			spec.ScratchDisks = []*GcpComputeInstanceScratchDisk{{Interface: "NVME", SizeGb: 375}}
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.ScratchDisks[0].SizeGb = 3000
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects an arbitrary size", func() {
			spec := makeValidSpec()
			spec.ScratchDisks = []*GcpComputeInstanceScratchDisk{{Interface: "NVME", SizeGb: 500}}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a missing interface", func() {
			spec := makeValidSpec()
			spec.ScratchDisks = []*GcpComputeInstanceScratchDisk{{}}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("network_interfaces", func() {
		It("rejects an interface with no attachment point (CEL)", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces = []*GcpComputeInstanceNetworkInterface{{}}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts a PSC attachment-only interface (CEL)", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces = []*GcpComputeInstanceNetworkInterface{
				{NetworkAttachment: "projects/123/regions/us-central1/networkAttachments/psc-a"},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("enforces vlan bounds on dynamic interfaces", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces[0].Vlan = i32(2)
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.NetworkInterfaces[0].Vlan = i32(255)
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.NetworkInterfaces[0].Vlan = i32(1)
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
			spec.NetworkInterfaces[0].Vlan = i32(256)
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("validates igmp_query values", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces[0].IgmpQuery = "IGMP_QUERY_V2"
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.NetworkInterfaces[0].IgmpQuery = "IGMP_QUERY_DISABLED"
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.NetworkInterfaces[0].IgmpQuery = "IGMP_QUERY_V3"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts internal IPv6 addressing and enforces prefix bounds", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces[0].StackType = "IPV4_IPV6"
			spec.NetworkInterfaces[0].Ipv6Address = "2600:1900:4000:1::"
			spec.NetworkInterfaces[0].InternalIpv6PrefixLength = i32(96)
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.NetworkInterfaces[0].InternalIpv6PrefixLength = i32(129)
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts an interface with subnetwork only", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces = []*GcpComputeInstanceNetworkInterface{
				{Subnetwork: refVal("app-subnet", "status.outputs.subnetwork_self_link")},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts a static internal IP referencing a GcpAddress", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces[0].NetworkIp = refVal("internal-vip", "status.outputs.address")
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects two access configs on one interface", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces[0].AccessConfigs = []*GcpComputeInstanceAccessConfig{{Ephemeral: proto.Bool(true)}, {Ephemeral: proto.Bool(true)}}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts valid stack types and rejects invalid ones", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces[0].StackType = "IPV4_IPV6"
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.NetworkInterfaces[0].StackType = "IPV6_PREFERRED"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts valid nic types and rejects invalid ones", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces[0].NicType = "GVNIC"
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.NetworkInterfaces[0].NicType = "E1000"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("enforces queue_count bounds", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces[0].QueueCount = i32(16)
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.NetworkInterfaces[0].QueueCount = i32(33)
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("requires PREMIUM tier on ipv6 access configs", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces[0].Ipv6AccessConfigs = []*GcpComputeInstanceIpv6AccessConfig{
				{NetworkTier: "PREMIUM"},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.NetworkInterfaces[0].Ipv6AccessConfigs[0].NetworkTier = "STANDARD"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("requires ip_cidr_range on alias ranges", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces[0].AliasIpRanges = []*GcpComputeInstanceAliasIpRange{{}}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts a pinned static external IPv6 access config", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces[0].StackType = "IPV4_IPV6"
			spec.NetworkInterfaces[0].Ipv6AccessConfigs = []*GcpComputeInstanceIpv6AccessConfig{
				{
					NetworkTier:              "PREMIUM",
					ExternalIpv6:             "2600:1900:4000:1::",
					ExternalIpv6PrefixLength: "96",
					Name:                     "External IPv6",
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	Context("access_configs", func() {
		It("accepts a static external IP referencing a GcpAddress", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces[0].AccessConfigs = []*GcpComputeInstanceAccessConfig{
				{NatIp: refVal("static-ip", "status.outputs.address")},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts an ephemeral external IP stated out loud", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces[0].AccessConfigs = []*GcpComputeInstanceAccessConfig{
				{Ephemeral: proto.Bool(true), NetworkTier: "PREMIUM"},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects an access config that names neither a static nat_ip nor ephemeral", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces[0].AccessConfigs = []*GcpComputeInstanceAccessConfig{
				{NetworkTier: "PREMIUM"},
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("set nat_ip, or set ephemeral: true"))
		})

		It("rejects an access config that names both a static nat_ip and ephemeral", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces[0].AccessConfigs = []*GcpComputeInstanceAccessConfig{
				{NatIp: refVal("static-ip", "status.outputs.address"), Ephemeral: proto.Bool(true)},
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects an invalid network tier", func() {
			spec := makeValidSpec()
			spec.NetworkInterfaces[0].AccessConfigs = []*GcpComputeInstanceAccessConfig{
				{Ephemeral: proto.Bool(true), NetworkTier: "BASIC"},
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("service_account", func() {
		It("accepts an email reference with scopes", func() {
			spec := makeValidSpec()
			spec.ServiceAccount = &GcpComputeInstanceServiceAccount{
				Email:  refVal("app-sa", "status.outputs.email"),
				Scopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a service account block without scopes", func() {
			spec := makeValidSpec()
			spec.ServiceAccount = &GcpComputeInstanceServiceAccount{
				Email: strVal("sa@p.iam.gserviceaccount.com"),
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("scheduling", func() {
		It("accepts a Spot configuration", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				ProvisioningModel:         "SPOT",
				AutomaticRestart:          boolPtr(false),
				OnHostMaintenance:         "TERMINATE",
				InstanceTerminationAction: "DELETE",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects Spot with automatic_restart true (CEL)", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				ProvisioningModel: "SPOT",
				AutomaticRestart:  boolPtr(true),
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts Spot with automatic_restart unset", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{ProvisioningModel: "SPOT"}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a termination action on a plain standard VM (CEL)", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				InstanceTerminationAction: "STOP",
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts a termination action on a timed-run standard VM (CEL)", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				InstanceTerminationAction: "STOP",
				MaxRunDurationSeconds:     i64(7200),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects an invalid provisioning model", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{ProvisioningModel: "FLEX"}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts a FLEX_START configuration", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				ProvisioningModel:         "FLEX_START",
				InstanceTerminationAction: "DELETE",
				MaxRunDurationSeconds:     i64(3600),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects FLEX_START without max_run_duration_seconds (CEL)", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				ProvisioningModel: "FLEX_START",
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects FLEX_START with automatic_restart true (CEL)", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				ProvisioningModel:     "FLEX_START",
				MaxRunDurationSeconds: i64(3600),
				AutomaticRestart:      boolPtr(true),
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects FLEX_START with a STOP termination action (CEL)", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				ProvisioningModel:         "FLEX_START",
				MaxRunDurationSeconds:     i64(3600),
				InstanceTerminationAction: "STOP",
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts RESERVATION_BOUND with a specific reservation (CEL)", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				ProvisioningModel: "RESERVATION_BOUND",
			}
			spec.ReservationAffinity = &GcpComputeInstanceReservationAffinity{
				Type: "SPECIFIC_RESERVATION",
				SpecificReservation: &GcpComputeInstanceSpecificReservation{
					Key:    "compute.googleapis.com/reservation-name",
					Values: []string{"my-reservation"},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects RESERVATION_BOUND without a specific reservation (CEL)", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				ProvisioningModel: "RESERVATION_BOUND",
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts max_run_duration_seconds", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				ProvisioningModel:         "SPOT",
				InstanceTerminationAction: "DELETE",
				MaxRunDurationSeconds:     i64(3600),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects max_run_duration together with termination_time (CEL)", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				MaxRunDurationSeconds: i64(3600),
				TerminationTime:       "2030-01-01T00:00:00Z",
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts sole-tenant node affinities", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				MinNodeCpus: i32(4),
				NodeAffinities: []*GcpComputeInstanceNodeAffinity{
					{
						Key:      "compute.googleapis.com/node-group-name",
						Operator: "IN",
						Values:   []string{"prod-nodes"},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a node affinity with an invalid operator", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				NodeAffinities: []*GcpComputeInstanceNodeAffinity{
					{Key: "k", Operator: "EQUALS", Values: []string{"v"}},
				},
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("enforces the local-SSD recovery timeout ceiling", func() {
			spec := makeValidSpec()
			spec.Scheduling = &GcpComputeInstanceScheduling{
				LocalSsdRecoveryTimeoutSeconds: i64(604801),
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("shielded and confidential", func() {
		It("accepts a shielded config", func() {
			spec := makeValidSpec()
			spec.ShieldedInstanceConfig = &GcpComputeInstanceShieldedConfig{
				EnableSecureBoot:          boolPtr(true),
				EnableVtpm:                boolPtr(true),
				EnableIntegrityMonitoring: boolPtr(true),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts a confidential VM with TERMINATE maintenance", func() {
			spec := makeValidSpec()
			spec.MachineType = "n2d-standard-2"
			spec.ConfidentialInstanceConfig = &GcpComputeInstanceConfidentialConfig{
				ConfidentialInstanceType: "SEV",
			}
			spec.Scheduling = &GcpComputeInstanceScheduling{OnHostMaintenance: "TERMINATE"}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a confidential VM without TERMINATE maintenance (CEL)", func() {
			spec := makeValidSpec()
			spec.ConfidentialInstanceConfig = &GcpComputeInstanceConfidentialConfig{
				ConfidentialInstanceType: "SEV",
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects an invalid confidential type", func() {
			spec := makeValidSpec()
			spec.ConfidentialInstanceConfig = &GcpComputeInstanceConfidentialConfig{
				ConfidentialInstanceType: "SGX",
			}
			spec.Scheduling = &GcpComputeInstanceScheduling{OnHostMaintenance: "TERMINATE"}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("advanced_machine_features", func() {
		It("accepts SMT-off with visible cores", func() {
			spec := makeValidSpec()
			spec.AdvancedMachineFeatures = &GcpComputeInstanceAdvancedMachineFeatures{
				ThreadsPerCore:   i32(1),
				VisibleCoreCount: i32(2),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects an invalid threads_per_core", func() {
			spec := makeValidSpec()
			spec.AdvancedMachineFeatures = &GcpComputeInstanceAdvancedMachineFeatures{
				ThreadsPerCore: i32(4),
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("validates pmu and turbo_mode values", func() {
			spec := makeValidSpec()
			spec.AdvancedMachineFeatures = &GcpComputeInstanceAdvancedMachineFeatures{
				PerformanceMonitoringUnit: "ENHANCED",
				TurboMode:                 "ALL_CORE_MAX",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.AdvancedMachineFeatures.PerformanceMonitoringUnit = "FULL"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("guest_accelerators", func() {
		It("accepts a GPU with TERMINATE maintenance", func() {
			spec := makeValidSpec()
			spec.GuestAccelerators = []*GcpComputeInstanceGuestAccelerator{
				{Type: "nvidia-tesla-t4", Count: 1},
			}
			spec.Scheduling = &GcpComputeInstanceScheduling{OnHostMaintenance: "TERMINATE"}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a GPU without TERMINATE maintenance (CEL)", func() {
			spec := makeValidSpec()
			spec.GuestAccelerators = []*GcpComputeInstanceGuestAccelerator{
				{Type: "nvidia-tesla-t4", Count: 1},
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a zero-count accelerator", func() {
			spec := makeValidSpec()
			spec.GuestAccelerators = []*GcpComputeInstanceGuestAccelerator{
				{Type: "nvidia-l4", Count: 0},
			}
			spec.Scheduling = &GcpComputeInstanceScheduling{OnHostMaintenance: "TERMINATE"}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("reservation_affinity", func() {
		It("accepts ANY_RESERVATION without a specific reservation", func() {
			spec := makeValidSpec()
			spec.ReservationAffinity = &GcpComputeInstanceReservationAffinity{
				Type: "ANY_RESERVATION",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts SPECIFIC_RESERVATION with the reservation named", func() {
			spec := makeValidSpec()
			spec.ReservationAffinity = &GcpComputeInstanceReservationAffinity{
				Type: "SPECIFIC_RESERVATION",
				SpecificReservation: &GcpComputeInstanceSpecificReservation{
					Key:    "compute.googleapis.com/reservation-name",
					Values: []string{"my-reservation"},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects SPECIFIC_RESERVATION without the reservation (CEL)", func() {
			spec := makeValidSpec()
			spec.ReservationAffinity = &GcpComputeInstanceReservationAffinity{
				Type: "SPECIFIC_RESERVATION",
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a specific reservation on NO_RESERVATION (CEL)", func() {
			spec := makeValidSpec()
			spec.ReservationAffinity = &GcpComputeInstanceReservationAffinity{
				Type: "NO_RESERVATION",
				SpecificReservation: &GcpComputeInstanceSpecificReservation{
					Key:    "k",
					Values: []string{"v"},
				},
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("instance-level scalars", func() {
		It("validates desired_status values", func() {
			spec := makeValidSpec()
			spec.DesiredStatus = "TERMINATED"
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.DesiredStatus = "STOPPED"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("validates total_egress_bandwidth_tier values", func() {
			spec := makeValidSpec()
			spec.TotalEgressBandwidthTier = "TIER_1"
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.TotalEgressBandwidthTier = "TIER_2"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("validates key_revocation_action_type values", func() {
			spec := makeValidSpec()
			spec.KeyRevocationActionType = "STOP"
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.KeyRevocationActionType = "DELETE"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("validates deletion_policy values", func() {
			spec := makeValidSpec()
			for _, valid := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
				spec.DeletionPolicy = valid
				Expect(protovalidate.Validate(spec)).To(BeNil())
			}
			spec.DeletionPolicy = "RETAIN"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts an instance-level CMEK key and requires kms_key inside it", func() {
			spec := makeValidSpec()
			spec.InstanceEncryptionKey = &GcpComputeInstanceEncryptionKey{
				KmsKey:               refVal("vm-key", "status.outputs.key_id"),
				KmsKeyServiceAccount: "cmek-agent@p.iam.gserviceaccount.com",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.InstanceEncryptionKey = &GcpComputeInstanceEncryptionKey{}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects more than one instance resource policy", func() {
			spec := makeValidSpec()
			spec.ResourcePolicies = []string{"a", "b"}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts metadata, labels, tags, ssh keys, and startup script", func() {
			spec := makeValidSpec()
			spec.Metadata = map[string]string{"enable-oslogin": "TRUE"}
			spec.Labels = map[string]string{"env": "prod"}
			spec.Tags = []string{"web-server"}
			spec.SshKeys = []string{"ops:ssh-ed25519 AAAA ops"}
			spec.StartupScript = "#!/bin/bash\necho hello"
			spec.ResourceManagerTags = map[string]string{"tagKeys/1": "tagValues/2"}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	Context("production-grade composition", func() {
		It("accepts a full stateful data VM", func() {
			spec := &GcpComputeInstanceSpec{
				ProjectId:    strVal("prod-project"),
				InstanceName: "pg-primary",
				Zone:         "us-central1-a",
				MachineType:  "n2-standard-8",
				Description:  "PostgreSQL primary",
				BootDisk: &GcpComputeInstanceBootDisk{
					Image:      "ubuntu-os-cloud/ubuntu-2404-lts-amd64",
					SizeGb:     100,
					Type:       "pd-balanced",
					AutoDelete: boolPtr(true),
					KmsKey:     refVal("boot-key", "status.outputs.key_id"),
					DiskLabels: map[string]string{"role": "boot"},
				},
				AttachedDisks: []*GcpComputeInstanceAttachedDisk{
					{
						Source:     refVal("pg-data", "status.outputs.self_link"),
						DeviceName: "pg-data",
						Mode:       "READ_WRITE",
					},
				},
				NetworkInterfaces: []*GcpComputeInstanceNetworkInterface{
					{
						Subnetwork: refVal("db-subnet", "status.outputs.subnetwork_self_link"),
						NetworkIp:  refVal("pg-vip", "status.outputs.address"),
						StackType:  "IPV4_ONLY",
						NicType:    "GVNIC",
					},
				},
				ServiceAccount: &GcpComputeInstanceServiceAccount{
					Email:  refVal("pg-sa", "status.outputs.email"),
					Scopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
				},
				ShieldedInstanceConfig: &GcpComputeInstanceShieldedConfig{
					EnableSecureBoot:          boolPtr(true),
					EnableVtpm:                boolPtr(true),
					EnableIntegrityMonitoring: boolPtr(true),
				},
				Scheduling: &GcpComputeInstanceScheduling{
					ProvisioningModel: "STANDARD",
					AutomaticRestart:  boolPtr(true),
					OnHostMaintenance: "MIGRATE",
				},
				Labels:                 map[string]string{"env": "prod", "app": "postgres"},
				Tags:                   []string{"db-server"},
				DeletionProtection:     true,
				AllowStoppingForUpdate: boolPtr(true),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})
})
