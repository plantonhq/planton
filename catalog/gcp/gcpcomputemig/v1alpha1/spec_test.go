package gcpcomputemigv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpComputeMigSpec Suite")
}

func strVal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpComputeMigSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpComputeMig {
		return &GcpComputeMig{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpComputeMig",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-mig",
			},
			Spec: &GcpComputeMigSpec{
				Zone: "us-central1-a",
				Template: &GcpComputeMigTemplate{
					MachineType: "e2-micro",
					Disks: []*GcpComputeMigTemplateDisk{
						{
							Boot:        true,
							SourceImage: "debian-cloud/debian-12",
						},
					},
					NetworkInterfaces: []*GcpComputeMigTemplateNetworkInterface{
						{
							Network: strVal("default"),
						},
					},
				},
			},
		}
	}

	ginkgo.It("accepts the minimal zonal manifest", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("accepts the minimal regional manifest", func() {
		m := minimal()
		m.Spec.Zone = ""
		m.Spec.Region = "us-central1"
		gomega.Expect(validator.Validate(m)).To(gomega.Succeed())
	})

	ginkgo.Context("location selector", func() {
		ginkgo.It("rejects a manifest with neither zone nor region", func() {
			m := minimal()
			m.Spec.Zone = ""
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("rejects a manifest with both zone and region", func() {
			m := minimal()
			m.Spec.Region = "us-central1"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("rejects malformed zone and region values", func() {
			m := minimal()
			m.Spec.Zone = "us-central1"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "region shape in zone")

			m = minimal()
			m.Spec.Zone = ""
			m.Spec.Region = "us-central1-a"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "zone shape in region")
		})
	})

	ginkgo.Context("regional-only surfaces", func() {
		ginkgo.It("rejects distribution_policy on a zonal group", func() {
			m := minimal()
			m.Spec.DistributionPolicy = &GcpComputeMigDistributionPolicy{TargetShape: "EVEN"}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("accepts distribution_policy on a regional group and walls target_shape", func() {
			m := minimal()
			m.Spec.Zone = ""
			m.Spec.Region = "us-central1"
			for _, v := range []string{"", "EVEN", "BALANCED", "ANY", "ANY_SINGLE_ZONE"} {
				m.Spec.DistributionPolicy = &GcpComputeMigDistributionPolicy{TargetShape: v}
				gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "value %q", v)
			}
			m.Spec.DistributionPolicy = &GcpComputeMigDistributionPolicy{TargetShape: "SPREAD"}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("rejects instance_flexibility_policy on a zonal group", func() {
			m := minimal()
			m.Spec.InstanceFlexibilityPolicy = &GcpComputeMigInstanceFlexibilityPolicy{
				InstanceSelections: []*GcpComputeMigInstanceSelection{
					{Name: "primary", MachineTypes: []string{"e2-micro"}},
				},
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("rejects update_policy.instance_redistribution_type on a zonal group", func() {
			m := minimal()
			m.Spec.UpdatePolicy = &GcpComputeMigUpdatePolicy{
				MinimalAction:              "REPLACE",
				Type:                       "PROACTIVE",
				InstanceRedistributionType: "NONE",
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("sizing", func() {
		ginkgo.It("rejects target_size together with an autoscaler", func() {
			m := minimal()
			m.Spec.TargetSize = proto.Int32(2)
			m.Spec.Autoscaler = &GcpComputeMigAutoscaler{MinReplicas: 1, MaxReplicas: 3}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("accepts each sizing mode alone", func() {
			m := minimal()
			m.Spec.TargetSize = proto.Int32(2)
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "fixed size")

			m = minimal()
			m.Spec.Autoscaler = &GcpComputeMigAutoscaler{MinReplicas: 1, MaxReplicas: 3}
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "autoscaler")
		})
	})

	ginkgo.Context("template", func() {
		ginkgo.It("requires machine_type", func() {
			m := minimal()
			m.Spec.Template.MachineType = ""
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("requires at least one disk and one network interface", func() {
			m := minimal()
			m.Spec.Template.Disks = nil
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "no disks")

			m = minimal()
			m.Spec.Template.NetworkInterfaces = nil
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "no NICs")
		})

		ginkgo.It("requires exactly one boot disk", func() {
			m := minimal()
			m.Spec.Template.Disks[0].Boot = false
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "zero boot disks")

			m = minimal()
			m.Spec.Template.Disks = append(m.Spec.Template.Disks, &GcpComputeMigTemplateDisk{
				Boot:        true,
				SourceImage: "debian-cloud/debian-12",
			})
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "two boot disks")
		})

		ginkgo.It("requires the boot disk to carry an OS source", func() {
			m := minimal()
			m.Spec.Template.Disks[0].SourceImage = ""
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("allows a blank non-boot data disk and rejects multi-source disks", func() {
			m := minimal()
			m.Spec.Template.Disks = append(m.Spec.Template.Disks, &GcpComputeMigTemplateDisk{
				SizeGb: 10,
			})
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "blank data disk")

			m.Spec.Template.Disks[1].SourceImage = "debian-cloud/debian-12"
			m.Spec.Template.Disks[1].SourceSnapshot = "snap-1"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "image + snapshot")
		})

		ginkgo.It("pairs source encryption with its source", func() {
			m := minimal()
			m.Spec.Template.Disks[0].SourceSnapshotEncryption = &GcpComputeMigEncryptionKey{
				KmsKey: strVal("projects/p/locations/l/keyRings/r/cryptoKeys/k"),
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "snapshot encryption without snapshot")
		})

		ginkgo.It("requires an attachment point on every NIC", func() {
			m := minimal()
			m.Spec.Template.NetworkInterfaces[0].Network = nil
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())

			m.Spec.Template.NetworkInterfaces[0].NetworkAttachment = "projects/1/regions/us-central1/networkAttachments/a"
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "attachment-only NIC")
		})

		ginkgo.It("rejects SPOT with automatic_restart true", func() {
			m := minimal()
			m.Spec.Template.Scheduling = &GcpComputeMigScheduling{
				ProvisioningModel: "SPOT",
				AutomaticRestart:  proto.Bool(true),
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("requires TERMINATE maintenance for confidential VMs and accelerators", func() {
			m := minimal()
			m.Spec.Template.ConfidentialInstanceConfig = &GcpComputeMigConfidentialConfig{
				ConfidentialInstanceType: "SEV",
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "confidential without TERMINATE")

			m.Spec.Template.Scheduling = &GcpComputeMigScheduling{OnHostMaintenance: "TERMINATE"}
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "confidential with TERMINATE")

			m = minimal()
			m.Spec.Template.GuestAccelerators = []*GcpComputeMigGuestAccelerator{
				{Type: "nvidia-tesla-t4", Count: 1},
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "GPU without TERMINATE")
		})

		ginkgo.It("requires a specific reservation for RESERVATION_BOUND", func() {
			m := minimal()
			m.Spec.Template.Scheduling = &GcpComputeMigScheduling{ProvisioningModel: "RESERVATION_BOUND"}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())

			m.Spec.Template.ReservationAffinity = &GcpComputeMigReservationAffinity{
				Type: "SPECIFIC_RESERVATION",
				SpecificReservation: &GcpComputeMigSpecificReservation{
					Key:    "compute.googleapis.com/reservation-name",
					Values: []string{"my-reservation"},
				},
			}
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())
		})

		ginkgo.It("rejects max_run_duration together with termination_time", func() {
			m := minimal()
			m.Spec.Template.Scheduling = &GcpComputeMigScheduling{
				MaxRunDurationSeconds: proto.Int64(3600),
				TerminationTime:       "2030-01-01T00:00:00Z",
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("versions", func() {
		ginkgo.It("rejects a version with both fixed and percent target size", func() {
			m := minimal()
			m.Spec.Versions = []*GcpComputeMigVersion{
				{
					VersionName:       "canary",
					TargetSizeFixed:   proto.Int32(1),
					TargetSizePercent: proto.Int32(10),
				},
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("update_policy", func() {
		valid := func() *GcpComputeMigUpdatePolicy {
			return &GcpComputeMigUpdatePolicy{
				MinimalAction: "REPLACE",
				Type:          "PROACTIVE",
			}
		}

		ginkgo.It("requires minimal_action and type from the provider's value sets", func() {
			m := minimal()
			m.Spec.UpdatePolicy = valid()
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())

			m.Spec.UpdatePolicy.MinimalAction = "PATCH"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "bad minimal_action")

			m.Spec.UpdatePolicy = valid()
			m.Spec.UpdatePolicy.Type = "AUTOMATIC"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "bad type")
		})

		ginkgo.It("rejects fixed together with percent budgets", func() {
			m := minimal()
			m.Spec.UpdatePolicy = valid()
			m.Spec.UpdatePolicy.MaxSurgeFixed = proto.Int32(1)
			m.Spec.UpdatePolicy.MaxSurgePercent = proto.Int32(10)
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "surge fixed+percent")

			m.Spec.UpdatePolicy = valid()
			m.Spec.UpdatePolicy.MaxUnavailableFixed = proto.Int32(1)
			m.Spec.UpdatePolicy.MaxUnavailablePercent = proto.Int32(10)
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "unavailable fixed+percent")
		})

		ginkgo.It("requires an unavailability budget above 0 for RECREATE", func() {
			m := minimal()
			m.Spec.UpdatePolicy = valid()
			m.Spec.UpdatePolicy.ReplacementMethod = "RECREATE"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "no budget")

			m.Spec.UpdatePolicy.MaxUnavailableFixed = proto.Int32(0)
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "zero budget")

			m.Spec.UpdatePolicy.MaxUnavailableFixed = proto.Int32(1)
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "budget of 1")
		})
	})

	ginkgo.Context("auto_healing", func() {
		ginkgo.It("requires the health check and bounds initial_delay_sec", func() {
			m := minimal()
			m.Spec.AutoHealing = &GcpComputeMigAutoHealing{InitialDelaySec: 300}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "no health check")

			m.Spec.AutoHealing.HealthCheck = strVal("https://www.googleapis.com/compute/v1/projects/p/global/healthChecks/hc")
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())

			m.Spec.AutoHealing.InitialDelaySec = 3601
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "delay above 3600")
		})
	})

	ginkgo.Context("autoscaler", func() {
		ginkgo.It("requires max_replicas >= min_replicas", func() {
			m := minimal()
			m.Spec.Autoscaler = &GcpComputeMigAutoscaler{MinReplicas: 3, MaxReplicas: 1}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("bounds cpu_target to (0, 1]", func() {
			m := minimal()
			m.Spec.Autoscaler = &GcpComputeMigAutoscaler{
				MinReplicas: 1, MaxReplicas: 3,
				CpuTarget: proto.Float64(0.6),
			}
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())

			m.Spec.Autoscaler.CpuTarget = proto.Float64(1.5)
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("walls mode and cpu_predictive_method to the documented sets", func() {
			m := minimal()
			m.Spec.Autoscaler = &GcpComputeMigAutoscaler{MinReplicas: 1, MaxReplicas: 3}
			for _, v := range []string{"", "ON", "OFF", "ONLY_SCALE_OUT"} {
				m.Spec.Autoscaler.Mode = v
				gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "mode %q", v)
			}
			m.Spec.Autoscaler.Mode = "ONLY_UP"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "retired ONLY_UP alias")

			m.Spec.Autoscaler.Mode = ""
			m.Spec.Autoscaler.CpuPredictiveMethod = "OPTIMIZE_AVAILABILITY"
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())
			m.Spec.Autoscaler.CpuPredictiveMethod = "PREDICT"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("keeps a metric utilization-targeted XOR workload-proportional", func() {
			m := minimal()
			m.Spec.Autoscaler = &GcpComputeMigAutoscaler{
				MinReplicas: 1, MaxReplicas: 3,
				Metrics: []*GcpComputeMigAutoscalerMetric{
					{
						Name:                     "pubsub.googleapis.com/subscription/num_undelivered_messages",
						Target:                   proto.Float64(100),
						SingleInstanceAssignment: proto.Float64(50),
					},
				},
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("requires a bound on scale_in_control and keeps the cap fixed XOR percent", func() {
			m := minimal()
			m.Spec.Autoscaler = &GcpComputeMigAutoscaler{
				MinReplicas: 1, MaxReplicas: 3,
				ScaleInControl: &GcpComputeMigScaleInControl{},
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "no bound")

			m.Spec.Autoscaler.ScaleInControl.MaxScaledInReplicasFixed = proto.Int32(1)
			m.Spec.Autoscaler.ScaleInControl.MaxScaledInReplicasPercent = proto.Int32(10)
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "fixed+percent")

			m.Spec.Autoscaler.ScaleInControl.MaxScaledInReplicasPercent = nil
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())
		})

		ginkgo.It("enforces the documented 300-second schedule minimum", func() {
			m := minimal()
			m.Spec.Autoscaler = &GcpComputeMigAutoscaler{
				MinReplicas: 1, MaxReplicas: 3,
				Schedules: []*GcpComputeMigScalingSchedule{
					{
						ScheduleName:        "business-hours",
						Schedule:            "0 8 * * MON-FRI",
						DurationSec:         299,
						MinRequiredReplicas: 2,
					},
				},
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())

			m.Spec.Autoscaler.Schedules[0].DurationSec = 300
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())
		})
	})

	ginkgo.Context("per_instance_configs", func() {
		ginkgo.It("requires config_name and walls the preserved-state value sets", func() {
			m := minimal()
			m.Spec.PerInstanceConfigs = []*GcpComputeMigPerInstanceConfig{{}}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "no name")

			m.Spec.PerInstanceConfigs[0].ConfigName = "test-mig-db-0"
			m.Spec.PerInstanceConfigs[0].PreservedState = &GcpComputeMigPreservedState{
				Disks: []*GcpComputeMigPreservedDisk{
					{
						DeviceName: "data",
						Source:     strVal("projects/p/zones/us-central1-a/disks/d"),
						DeleteRule: "ON_PERMANENT_INSTANCE_DELETION",
					},
				},
			}
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())

			m.Spec.PerInstanceConfigs[0].PreservedState.Disks[0].DeleteRule = "ALWAYS"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "bad delete_rule")
		})
	})

	ginkgo.Context("resize_requests", func() {
		ginkgo.It("bounds requested_run_duration_seconds to the provider's documented 600-604800", func() {
			m := minimal()
			m.Spec.ResizeRequests = []*GcpComputeMigResizeRequest{
				{
					RequestName:                 "batch-ask",
					ResizeBy:                    2,
					RequestedRunDurationSeconds: proto.Int64(599),
				},
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "below 600")

			m.Spec.ResizeRequests[0].RequestedRunDurationSeconds = proto.Int64(600)
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())

			m.Spec.ResizeRequests[0].RequestedRunDurationSeconds = proto.Int64(604801)
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "above 7 days")
		})

		ginkgo.It("requires a positive resize_by", func() {
			m := minimal()
			m.Spec.ResizeRequests = []*GcpComputeMigResizeRequest{
				{RequestName: "batch-ask", ResizeBy: 0},
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("simple value walls", func() {
		ginkgo.It("walls list_managed_instances_results, target_size_policy_mode, wait status, and deletion_policy", func() {
			m := minimal()
			m.Spec.ListManagedInstancesResults = "PAGINATED"
			m.Spec.TargetSizePolicyMode = "BULK"
			m.Spec.WaitForInstancesStatus = "UPDATED"
			m.Spec.DeletionPolicy = "ABANDON"
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())

			bad := []func(*GcpComputeMig){
				func(x *GcpComputeMig) { x.Spec.ListManagedInstancesResults = "ALL" },
				func(x *GcpComputeMig) { x.Spec.TargetSizePolicyMode = "ATOMIC" },
				func(x *GcpComputeMig) { x.Spec.WaitForInstancesStatus = "READY" },
				func(x *GcpComputeMig) { x.Spec.DeletionPolicy = "KEEP" },
			}
			for i, mutate := range bad {
				x := minimal()
				mutate(x)
				gomega.Expect(validator.Validate(x)).ToNot(gomega.Succeed(), "case %d", i)
			}
		})
	})

	ginkgo.Context("host error timeout and workload identity", func() {
		ginkgo.It("accepts host_error_timeout_seconds on the 30-second grid", func() {
			for _, v := range []int32{90, 210, 330} {
				m := minimal()
				m.Spec.Template.Scheduling = &GcpComputeMigScheduling{HostErrorTimeoutSeconds: proto.Int32(v)}
				gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "value %d", v)
			}
		})

		ginkgo.It("rejects host_error_timeout_seconds off the grid or out of range", func() {
			for _, v := range []int32{95, 60, 360} {
				m := minimal()
				m.Spec.Template.Scheduling = &GcpComputeMigScheduling{HostErrorTimeoutSeconds: proto.Int32(v)}
				gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "value %d", v)
			}
		})

		ginkgo.It("accepts a SPIFFE workload identity on the template", func() {
			m := minimal()
			m.Spec.Template.WorkloadIdentityConfig = &GcpComputeMigWorkloadIdentityConfig{
				Identity:                   "spiffe://my-project.svc.id.goog/ns/default/sa/app",
				IdentityCertificateEnabled: true,
			}
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())
		})

		ginkgo.It("rejects a workload identity that is not a SPIFFE ID", func() {
			m := minimal()
			m.Spec.Template.WorkloadIdentityConfig = &GcpComputeMigWorkloadIdentityConfig{Identity: "app@my-project.iam.gserviceaccount.com"}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})
})
