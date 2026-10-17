package kubernetesgkecomputeclassv1alpha1

import (
	"errors"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// The validation suite pins what the spec refuses before GKE sees the class:
// the ComputeClass definition's own rules (one selector per rule, nodepools
// alone, Autopilot constraints, confidential computing families, Spot-only
// shutdown grace, reservation affinity, swap profiles, kubelet thresholds),
// the served schema's enums and ranges, and the envelope's name grammar,
// including GKE's reserved built-in class names. Every rule has one case.

func TestKubernetesGkeComputeClass(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesGkeComputeClass Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name, FieldPath: fieldPath},
		},
	}
}

// ptr returns a pointer to v, for setting proto3 `optional` scalar fields.
func ptr[T any](v T) *T { return &v }

func familyRule(family string, cores int32) *KubernetesGkeComputeClassPriority {
	return &KubernetesGkeComputeClassPriority{MachineFamily: ptr(family), MinCores: ptr(cores)}
}

// expectViolation asserts err carries a violation of the rule with ruleID.
func expectViolation(err error, ruleID string) {
	gomega.Expect(err).NotTo(gomega.BeNil())
	var validationErr *protovalidate.ValidationError
	gomega.Expect(errors.As(err, &validationErr)).To(gomega.BeTrue(), "not a validation error: %v", err)
	var ids []string
	for _, violation := range validationErr.Violations {
		ids = append(ids, violation.Proto.GetRuleId())
	}
	gomega.Expect(ids).To(gomega.ContainElement(ruleID), "violations: %v", err)
}

var _ = ginkgo.Describe("KubernetesGkeComputeClass Validation Tests", func() {
	var input *KubernetesGkeComputeClass

	ginkgo.BeforeEach(func() {
		input = &KubernetesGkeComputeClass{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesGkeComputeClass",
			Metadata:   &shared.CatalogObjectMetadata{Name: "prod-default-class"},
			Spec: &KubernetesGkeComputeClassSpec{
				Name: "default",
				Priorities: []*KubernetesGkeComputeClassPriority{
					familyRule("n2d", 8),
					familyRule("c3d", 8),
					familyRule("n2", 8),
					familyRule("t2d", 8),
				},
				WhenUnsatisfiable:    ptr("ScaleUpAnyway"),
				NodePoolAutoCreation: &KubernetesGkeComputeClassNodePoolAutoCreation{Enabled: ptr(true)},
			},
		}
	})

	validate := func() error { return protovalidate.Validate(input) }

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts the stockout-resilient default class", func() {
			gomega.Expect(validate()).To(gomega.BeNil())
		})

		ginkgo.It("accepts a class with no priorities, as the definition does", func() {
			input.Spec.Priorities = nil
			gomega.Expect(validate()).To(gomega.BeNil())
		})

		ginkgo.It("accepts a nodepools rule that references node pools and carries a score", func() {
			input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{
				{
					Nodepools: []*foreignkeyv1.StringValueOrRef{
						valueFrom(catalogkind.CatalogKind_GcpGkeNodePool, "prod-gke-system", "status.outputs.node_pool_name"),
						literal("legacy-pool"),
					},
					AllocationStrategy: ptr("lowest-cost"),
					PriorityScore:      ptr(int32(100)),
				},
				{MachineFamily: ptr("n2d"), PriorityScore: ptr(int32(50))},
			}
			gomega.Expect(validate()).To(gomega.BeNil())
		})

		ginkgo.It("accepts a Specific reservation rule followed by on-demand fallbacks", func() {
			input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{
				{
					MachineType: ptr("n2d-standard-16"),
					Reservations: &KubernetesGkeComputeClassReservations{
						Affinity: "Specific",
						Specific: []*KubernetesGkeComputeClassSpecificReservation{{
							Name:  "platform-reservation",
							Zones: []string{"us-east1-b"},
						}},
					},
				},
				familyRule("c3d", 16),
			}
			gomega.Expect(validate()).To(gomega.BeNil())
		})

		ginkgo.It("accepts an Autopilot class with pod families and no nodepools", func() {
			input.Spec.Autopilot = &KubernetesGkeComputeClassAutopilot{Enabled: ptr(true)}
			input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{PodFamily: ptr("general-purpose")}}
			gomega.Expect(validate()).To(gomega.BeNil())
		})

		ginkgo.It("accepts a confidential SEV class on supported families", func() {
			input.Spec.NodePoolConfig = &KubernetesGkeComputeClassNodePoolConfig{ConfidentialNodeType: ptr("SEV")}
			input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{
				familyRule("n2d", 4),
				{MachineType: ptr("c3d-standard-8")},
			}
			gomega.Expect(validate()).To(gomega.BeNil())
		})

		ginkgo.It("accepts a Spot rule with kubelet shutdown grace", func() {
			input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{
				MachineFamily: ptr("c3d"),
				Spot:          ptr(true),
				NodeSystemConfig: &KubernetesGkeComputeClassNodeSystemConfig{
					KubeletConfig: &KubernetesGkeComputeClassKubeletConfig{
						ShutdownGracePeriodSeconds:             ptr(int32(30)),
						ShutdownGracePeriodCriticalPodsSeconds: ptr(int32(15)),
					},
				},
			}}
			gomega.Expect(validate()).To(gomega.BeNil())
		})

		ginkgo.It("accepts sysctls under their dotted names and references for identity and CMEK", func() {
			input.Spec.NodePoolConfig = &KubernetesGkeComputeClassNodePoolConfig{
				ServiceAccount: valueFrom(catalogkind.CatalogKind_GcpServiceAccount, "gke-nodes", "status.outputs.email"),
			}
			input.Spec.Priorities[0].Storage = &KubernetesGkeComputeClassStorage{
				BootDiskKmsKey: valueFrom(catalogkind.CatalogKind_GcpKmsKey, "gke-boot", "status.outputs.key_id"),
			}
			input.Spec.PriorityDefaults = &KubernetesGkeComputeClassPriorityDefaults{
				NodeSystemConfig: &KubernetesGkeComputeClassNodeSystemConfig{
					LinuxNodeConfig: &KubernetesGkeComputeClassLinuxNodeConfig{
						Sysctls: &KubernetesGkeComputeClassSysctls{
							NetCoreSomaxconn:   ptr(int64(4096)),
							VmMaxMapCount:      ptr(int64(262144)),
							KernelShmmax:       ptr("68719476736"),
							VmOvercommitMemory: ptr(int64(1)),
						},
					},
				},
			}
			gomega.Expect(validate()).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Describe("the envelope", func() {
			ginkgo.It("refuses an empty name", func() {
				input.Spec.Name = ""
				gomega.Expect(validate()).NotTo(gomega.BeNil())
			})

			ginkgo.It("refuses a name that is not a DNS subdomain", func() {
				input.Spec.Name = "Default_Class"
				expectViolation(validate(), "spec.name.dns_subdomain")
			})

			ginkgo.It("refuses GKE's built-in class names", func() {
				for _, name := range []string{"autopilot", "autopilot-spot", "autopilot-arm-spot"} {
					input.Spec.Name = name
					expectViolation(validate(), "spec.name.not_builtin")
				}
			})

			ginkgo.It("refuses an object label key outside Kubernetes label grammar", func() {
				input.Spec.Labels = map[string]string{"bad key": "x"}
				gomega.Expect(validate()).NotTo(gomega.BeNil())
			})

			ginkgo.It("refuses an unknown when_unsatisfiable", func() {
				input.Spec.WhenUnsatisfiable = ptr("ScaleUpSometimes")
				gomega.Expect(validate()).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Describe("class-level rules", func() {
			ginkgo.It("refuses a nodepools rule on an Autopilot class", func() {
				input.Spec.Autopilot = &KubernetesGkeComputeClassAutopilot{Enabled: ptr(true)}
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{Nodepools: []*foreignkeyv1.StringValueOrRef{literal("pool-a")}}}
				expectViolation(validate(), "spec.autopilot_forbids_nodepools")
			})

			ginkgo.It("refuses disabling node pool auto-creation on an Autopilot class", func() {
				input.Spec.Autopilot = &KubernetesGkeComputeClassAutopilot{Enabled: ptr(true)}
				input.Spec.NodePoolAutoCreation = &KubernetesGkeComputeClassNodePoolAutoCreation{Enabled: ptr(false)}
				expectViolation(validate(), "spec.autopilot_requires_auto_creation")
			})

			ginkgo.It("refuses an Ubuntu image on an Autopilot class", func() {
				input.Spec.Autopilot = &KubernetesGkeComputeClassAutopilot{Enabled: ptr(true)}
				input.Spec.NodePoolConfig = &KubernetesGkeComputeClassNodePoolConfig{ImageType: ptr("ubuntu_containerd")}
				expectViolation(validate(), "spec.autopilot_image_type")
			})

			ginkgo.It("refuses the MAX_THROUGHPUT logging variant on an Autopilot class", func() {
				input.Spec.Autopilot = &KubernetesGkeComputeClassAutopilot{Enabled: ptr(true)}
				input.Spec.NodePoolConfig = &KubernetesGkeComputeClassNodePoolConfig{
					LoggingConfig: &KubernetesGkeComputeClassLoggingConfig{
						LoggingVariantConfig: &KubernetesGkeComputeClassLoggingVariantConfig{Variant: ptr("MAX_THROUGHPUT")},
					},
				}
				expectViolation(validate(), "spec.autopilot_logging_variant")
			})

			ginkgo.It("refuses HIGH_THROUGHPUT without a node pool group", func() {
				input.Spec.NodePoolConfig = &KubernetesGkeComputeClassNodePoolConfig{WorkloadType: ptr("HIGH_THROUGHPUT")}
				expectViolation(validate(), "spec.workload_type_requires_node_pool_group")
			})

			ginkgo.It("refuses a pod family on a Standard class", func() {
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{PodFamily: ptr("general-purpose")}}
				expectViolation(validate(), "spec.pod_family_requires_autopilot")
			})

			ginkgo.It("refuses confidential nodes when a rule names no machine", func() {
				input.Spec.NodePoolConfig = &KubernetesGkeComputeClassNodePoolConfig{ConfidentialNodeType: ptr("SEV")}
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{MinCores: ptr(int32(4))}}
				expectViolation(validate(), "spec.confidential_requires_machine")
			})

			ginkgo.It("refuses SEV on an unsupported family", func() {
				input.Spec.NodePoolConfig = &KubernetesGkeComputeClassNodePoolConfig{ConfidentialNodeType: ptr("SEV")}
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{MachineType: ptr("n2-standard-8")}}
				expectViolation(validate(), "spec.confidential_sev_families")
			})

			ginkgo.It("refuses SEV on the g4 family without the one supported machine type", func() {
				input.Spec.NodePoolConfig = &KubernetesGkeComputeClassNodePoolConfig{ConfidentialNodeType: ptr("SEV")}
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{MachineFamily: ptr("g4")}}
				expectViolation(validate(), "spec.confidential_sev_g4")
			})

			ginkgo.It("refuses SEV_SNP outside n2d", func() {
				input.Spec.NodePoolConfig = &KubernetesGkeComputeClassNodePoolConfig{ConfidentialNodeType: ptr("SEV_SNP")}
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{familyRule("c3d", 8)}
				expectViolation(validate(), "spec.confidential_sev_snp_families")
			})

			ginkgo.It("refuses TDX on an unsupported family", func() {
				input.Spec.NodePoolConfig = &KubernetesGkeComputeClassNodePoolConfig{ConfidentialNodeType: ptr("TDX")}
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{familyRule("n2d", 8)}
				expectViolation(validate(), "spec.confidential_tdx_families")
			})

			ginkgo.It("refuses TDX on a3 with a GPU other than the H100", func() {
				input.Spec.NodePoolConfig = &KubernetesGkeComputeClassNodePoolConfig{ConfidentialNodeType: ptr("TDX")}
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{
					MachineType: ptr("a3-highgpu-1g"),
					Gpu:         &KubernetesGkeComputeClassGpu{Type: ptr("nvidia-l4")},
				}}
				expectViolation(validate(), "spec.confidential_tdx_gpus")
			})

			ginkgo.It("refuses a priority score on some rules but not all", func() {
				input.Spec.Priorities[0].PriorityScore = ptr(int32(10))
				expectViolation(validate(), "spec.priority_score_all_or_none")
			})

			ginkgo.It("refuses a default shutdown grace unless every rule is Spot", func() {
				input.Spec.PriorityDefaults = &KubernetesGkeComputeClassPriorityDefaults{
					NodeSystemConfig: &KubernetesGkeComputeClassNodeSystemConfig{
						KubeletConfig: &KubernetesGkeComputeClassKubeletConfig{ShutdownGracePeriodSeconds: ptr(int32(30))},
					},
				}
				expectViolation(validate(), "spec.default_shutdown_grace_requires_spot")
			})

			ginkgo.It("refuses a class-level minimum capacity over rules without an exact machine", func() {
				input.Spec.MinimumCapacity = &KubernetesGkeComputeClassMinimumCapacity{TargetNodeCount: ptr(int32(2))}
				expectViolation(validate(), "spec.minimum_capacity_requires_machine")
			})

			ginkgo.It("refuses the tcp_congestion_control sysctl on an Autopilot class", func() {
				input.Spec.Autopilot = &KubernetesGkeComputeClassAutopilot{Enabled: ptr(true)}
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{
					MachineFamily: ptr("n2d"),
					NodeSystemConfig: &KubernetesGkeComputeClassNodeSystemConfig{
						LinuxNodeConfig: &KubernetesGkeComputeClassLinuxNodeConfig{
							Sysctls: &KubernetesGkeComputeClassSysctls{NetIpv4TcpCongestionControl: ptr("bbr")},
						},
					},
				}}
				expectViolation(validate(), "spec.autopilot_forbids_tcp_congestion_control")
			})
		})

		ginkgo.Describe("priority rules", func() {
			ginkgo.It("refuses a nodepools rule that carries other settings", func() {
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{
					Nodepools: []*foreignkeyv1.StringValueOrRef{literal("pool-a")},
					Spot:      ptr(true),
				}}
				expectViolation(validate(), "priority.nodepools_alone")
			})

			ginkgo.It("refuses machine_family together with machine_type", func() {
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{MachineFamily: ptr("n2d"), MachineType: ptr("n2d-standard-8")}}
				expectViolation(validate(), "priority.family_xor_type")
			})

			ginkgo.It("refuses machine_type together with minimums", func() {
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{MachineType: ptr("n2d-standard-8"), MinCores: ptr(int32(8))}}
				expectViolation(validate(), "priority.type_excludes_minimums")
			})

			ginkgo.It("refuses the reserved ek family and machine types", func() {
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{MachineFamily: ptr("ek")}}
				expectViolation(validate(), "priority.no_ek_family")
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{MachineType: ptr("ek-standard-8")}}
				expectViolation(validate(), "priority.no_ek_family")
			})

			ginkgo.It("refuses flex-start together with Spot", func() {
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{
					MachineFamily: ptr("a3"),
					Spot:          ptr(true),
					FlexStart:     &KubernetesGkeComputeClassFlexStart{Enabled: ptr(true)},
				}}
				expectViolation(validate(), "priority.flex_start_excludes_spot")
			})

			ginkgo.It("refuses a capacity wait without flex-start or a TPU", func() {
				input.Spec.Priorities[0].CapacityCheckWaitTimeSeconds = ptr(int32(600))
				expectViolation(validate(), "priority.capacity_wait_scope")
			})

			ginkgo.It("refuses a rule shutdown grace on an on-demand rule", func() {
				input.Spec.Priorities[0].NodeSystemConfig = &KubernetesGkeComputeClassNodeSystemConfig{
					KubeletConfig: &KubernetesGkeComputeClassKubeletConfig{ShutdownGracePeriodSeconds: ptr(int32(30))},
				}
				expectViolation(validate(), "priority.shutdown_grace_requires_spot")
			})

			ginkgo.It("refuses GPUDirect RDMA without an accelerator network profile", func() {
				input.Spec.Priorities[0].GpuDirect = ptr("rdma")
				expectViolation(validate(), "priority.rdma_requires_network_profile")
			})

			ginkgo.It("refuses a gpu_direct value other than rdma", func() {
				input.Spec.Priorities[0].GpuDirect = ptr("tcpx")
				gomega.Expect(validate()).NotTo(gomega.BeNil())
			})

			ginkgo.It("refuses a rule minimum capacity without an exact machine", func() {
				input.Spec.Priorities[0].MinimumCapacity = &KubernetesGkeComputeClassMinimumCapacity{TargetNodeCount: ptr(int32(1))}
				expectViolation(validate(), "priority.minimum_capacity_requires_machine")
			})

			ginkgo.It("refuses a reserved instance metadata key", func() {
				input.Spec.Priorities[0].InstanceMetadata = map[string]string{"startup-script": "echo hi"}
				expectViolation(validate(), "priority.instance_metadata.reserved_keys")
			})

			ginkgo.It("refuses a machine family longer than the definition allows", func() {
				input.Spec.Priorities[0].MachineFamily = ptr("averylongfamily")
				gomega.Expect(validate()).NotTo(gomega.BeNil())
			})

			ginkgo.It("refuses max pods per node outside 8-256", func() {
				input.Spec.Priorities[0].MaxPodsPerNode = ptr(int32(300))
				gomega.Expect(validate()).NotTo(gomega.BeNil())
			})

			ginkgo.It("refuses an unknown boot disk type", func() {
				input.Spec.Priorities[0].Storage = &KubernetesGkeComputeClassStorage{BootDiskType: ptr("hyperdisk-extreme")}
				gomega.Expect(validate()).NotTo(gomega.BeNil())
			})

			ginkgo.It("refuses an unknown zone type", func() {
				input.Spec.Priorities[0].Location = &KubernetesGkeComputeClassLocation{ZoneTypes: []string{"EDGE"}}
				gomega.Expect(validate()).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Describe("nested rules", func() {
			ginkgo.It("refuses specific reservations without Specific affinity", func() {
				input.Spec.Priorities[0].Reservations = &KubernetesGkeComputeClassReservations{
					Affinity: "AnyBestEffort",
					Specific: []*KubernetesGkeComputeClassSpecificReservation{{Name: "r1"}},
				}
				expectViolation(validate(), "reservations.specific_requires_affinity")
			})

			ginkgo.It("refuses Specific affinity without a reservation", func() {
				input.Spec.Priorities[0].Reservations = &KubernetesGkeComputeClassReservations{Affinity: "Specific"}
				expectViolation(validate(), "reservations.affinity_requires_specific")
			})

			ginkgo.It("refuses a reservation block without a name", func() {
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{
					MachineType: ptr("a3-highgpu-8g"),
					Reservations: &KubernetesGkeComputeClassReservations{
						Affinity: "Specific",
						Specific: []*KubernetesGkeComputeClassSpecificReservation{{
							Name:             "r1",
							ReservationBlock: &KubernetesGkeComputeClassReservationBlock{},
						}},
					},
				}}
				gomega.Expect(validate()).NotTo(gomega.BeNil())
			})

			ginkgo.It("refuses kubelet image GC thresholds out of order", func() {
				input.Spec.Priorities[0].NodeSystemConfig = &KubernetesGkeComputeClassNodeSystemConfig{
					KubeletConfig: &KubernetesGkeComputeClassKubeletConfig{
						ImageGcLowThresholdPercent:  ptr(int64(80)),
						ImageGcHighThresholdPercent: ptr(int64(70)),
					},
				}
				expectViolation(validate(), "kubelet_config.image_gc_thresholds_ordered")
			})

			ginkgo.It("refuses a lone image GC high threshold at or below the default low", func() {
				input.Spec.Priorities[0].NodeSystemConfig = &KubernetesGkeComputeClassNodeSystemConfig{
					KubeletConfig: &KubernetesGkeComputeClassKubeletConfig{ImageGcHighThresholdPercent: ptr(int64(75))},
				}
				expectViolation(validate(), "kubelet_config.image_gc_high_above_default_low")
			})

			ginkgo.It("refuses a critical-pods shutdown grace without or above the shutdown grace", func() {
				input.Spec.Priorities[0].Spot = ptr(true)
				input.Spec.Priorities[0].NodeSystemConfig = &KubernetesGkeComputeClassNodeSystemConfig{
					KubeletConfig: &KubernetesGkeComputeClassKubeletConfig{ShutdownGracePeriodCriticalPodsSeconds: ptr(int32(15))},
				}
				expectViolation(validate(), "kubelet_config.critical_shutdown_within_shutdown")
			})

			ginkgo.It("refuses a shutdown grace other than 0, 30 or 120", func() {
				input.Spec.Priorities[0].Spot = ptr(true)
				input.Spec.Priorities[0].NodeSystemConfig = &KubernetesGkeComputeClassNodeSystemConfig{
					KubeletConfig: &KubernetesGkeComputeClassKubeletConfig{ShutdownGracePeriodSeconds: ptr(int32(60))},
				}
				gomega.Expect(validate()).NotTo(gomega.BeNil())
			})

			ginkgo.It("refuses two swap profiles at once", func() {
				input.Spec.Priorities[0].NodeSystemConfig = &KubernetesGkeComputeClassNodeSystemConfig{
					LinuxNodeConfig: &KubernetesGkeComputeClassLinuxNodeConfig{
						SwapConfig: &KubernetesGkeComputeClassSwapConfig{
							BootDiskProfile:          &KubernetesGkeComputeClassBootDiskProfile{SwapSizeGib: ptr(int64(4))},
							DedicatedLocalSsdProfile: &KubernetesGkeComputeClassDedicatedLocalSsdProfile{DiskCount: ptr(int64(1))},
						},
					},
				}
				expectViolation(validate(), "swap_config.one_profile")
			})

			ginkgo.It("refuses a swap size in both GiB and percent", func() {
				input.Spec.Priorities[0].NodeSystemConfig = &KubernetesGkeComputeClassNodeSystemConfig{
					LinuxNodeConfig: &KubernetesGkeComputeClassLinuxNodeConfig{
						SwapConfig: &KubernetesGkeComputeClassSwapConfig{
							BootDiskProfile: &KubernetesGkeComputeClassBootDiskProfile{SwapSizeGib: ptr(int64(4)), SwapSizePercent: ptr(int32(10))},
						},
					},
				}
				expectViolation(validate(), "boot_disk_profile.one_size")
				input.Spec.Priorities[0].NodeSystemConfig.LinuxNodeConfig.SwapConfig = &KubernetesGkeComputeClassSwapConfig{
					EphemeralLocalSsdProfile: &KubernetesGkeComputeClassEphemeralLocalSsdProfile{SwapSizeGib: ptr(int64(4)), SwapSizePercent: ptr(int32(10))},
				}
				expectViolation(validate(), "ephemeral_local_ssd_profile.one_size")
			})

			ginkgo.It("refuses a sysctl outside the range GKE accepts", func() {
				input.Spec.Priorities[0].NodeSystemConfig = &KubernetesGkeComputeClassNodeSystemConfig{
					LinuxNodeConfig: &KubernetesGkeComputeClassLinuxNodeConfig{
						Sysctls: &KubernetesGkeComputeClassSysctls{NetCoreSomaxconn: ptr(int64(64))},
					},
				}
				gomega.Expect(validate()).NotTo(gomega.BeNil())
			})

			ginkgo.It("refuses a flex-start node recycling without its lead time", func() {
				input.Spec.Priorities = []*KubernetesGkeComputeClassPriority{{
					MachineFamily: ptr("a3"),
					FlexStart: &KubernetesGkeComputeClassFlexStart{
						Enabled:       ptr(true),
						NodeRecycling: &KubernetesGkeComputeClassNodeRecycling{},
					},
				}}
				gomega.Expect(validate()).NotTo(gomega.BeNil())
			})

			ginkgo.It("refuses a taint without a key or with an unknown effect", func() {
				input.Spec.NodePoolConfig = &KubernetesGkeComputeClassNodePoolConfig{
					Taints: []*KubernetesGkeComputeClassTaint{{Key: "", Effect: "NoSchedule"}},
				}
				gomega.Expect(validate()).NotTo(gomega.BeNil())
				input.Spec.NodePoolConfig.Taints = []*KubernetesGkeComputeClassTaint{{Key: "team", Effect: "Evict"}}
				gomega.Expect(validate()).NotTo(gomega.BeNil())
			})

			ginkgo.It("refuses more than five resource manager tags", func() {
				tags := make([]*KubernetesGkeComputeClassResourceManagerTag, 6)
				for i := range tags {
					tags[i] = &KubernetesGkeComputeClassResourceManagerTag{Key: "tagKeys/1", Value: "tagValues/1"}
				}
				input.Spec.NodePoolConfig = &KubernetesGkeComputeClassNodePoolConfig{ResourceManagerTags: tags}
				gomega.Expect(validate()).NotTo(gomega.BeNil())
			})

			ginkgo.It("refuses a node pool group without a name", func() {
				input.Spec.NodePoolGroup = &KubernetesGkeComputeClassNodePoolGroup{}
				gomega.Expect(validate()).NotTo(gomega.BeNil())
			})

			ginkgo.It("refuses a consolidation delay outside 1-1440 minutes", func() {
				input.Spec.AutoscalingPolicy = &KubernetesGkeComputeClassAutoscalingPolicy{ConsolidationDelayMinutes: ptr(int32(0))}
				gomega.Expect(validate()).NotTo(gomega.BeNil())
			})
		})
	})
})
