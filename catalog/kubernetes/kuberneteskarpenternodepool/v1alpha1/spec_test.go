package kuberneteskarpenternodepoolv1alpha1

import (
	"fmt"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesKarpenterNodePool(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKarpenterNodePool Suite")
}

func stringPtr(s string) *string { return &s }
func int32Ptr(i int32) *int32    { return &i }
func int64Ptr(i int64) *int64    { return &i }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind:      kind,
				Name:      name,
				FieldPath: fieldPath,
			},
		},
	}
}

// capacityTypeRequirement returns the practical-minimum requirement every
// fixture starts from.
func capacityTypeRequirement() *KubernetesKarpenterNodePoolRequirement {
	return &KubernetesKarpenterNodePoolRequirement{
		Key:      "karpenter.sh/capacity-type",
		Operator: "In",
		Values:   []string{"spot"},
	}
}

var _ = ginkgo.Describe("KubernetesKarpenterNodePool Validation Tests", func() {
	var input *KubernetesKarpenterNodePool

	ginkgo.BeforeEach(func() {
		input = &KubernetesKarpenterNodePool{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesKarpenterNodePool",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-node-pool",
			},
			Spec: &KubernetesKarpenterNodePoolSpec{
				Template: &KubernetesKarpenterNodePoolTemplate{
					NodeClassRef: &KubernetesKarpenterNodePoolNodeClassRef{
						Name: literal("default-ec2"),
					},
					Requirements: []*KubernetesKarpenterNodePoolRequirement{
						capacityTypeRequirement(),
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal pool (nodeClassRef + one requirement) should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface should be valid", func() {
			input.Spec.Template.Labels = map[string]string{
				"team":                       "platform",
				"karpenter.sh/capacity-type": "spot",
			}
			input.Spec.Template.Annotations = map[string]string{
				"example.com/owner": "platform-team",
			}
			input.Spec.Template.NodeClassRef = &KubernetesKarpenterNodePoolNodeClassRef{
				Group: stringPtr("karpenter.k8s.aws"),
				Kind:  stringPtr("EC2NodeClass"),
				Name:  literal("gp-nodeclass"),
			}
			input.Spec.Template.Requirements = []*KubernetesKarpenterNodePoolRequirement{
				capacityTypeRequirement(),
				{Key: "kubernetes.io/arch", Operator: "In", Values: []string{"amd64", "arm64"}},
				{Key: "karpenter.k8s.aws/instance-category", Operator: "NotIn", Values: []string{"t"}},
				{Key: "karpenter.k8s.aws/instance-generation", Operator: "Gt", Values: []string{"2"}},
				{Key: "topology.kubernetes.io/zone", Operator: "Exists"},
			}
			input.Spec.Template.Taints = []*KubernetesKarpenterNodePoolTaint{
				{Key: "dedicated", Value: "batch", Effect: "NoSchedule"},
			}
			input.Spec.Template.StartupTaints = []*KubernetesKarpenterNodePoolTaint{
				{Key: "example.com/cni-not-ready", Effect: "NoExecute"},
			}
			input.Spec.Template.ExpireAfter = stringPtr("720h")
			input.Spec.Template.TerminationGracePeriod = stringPtr("48h")
			input.Spec.Disruption = &KubernetesKarpenterNodePoolDisruption{
				ConsolidationPolicy: stringPtr("WhenEmptyOrUnderutilized"),
				ConsolidateAfter:    stringPtr("5m"),
				Budgets: []*KubernetesKarpenterNodePoolDisruptionBudget{
					{Nodes: "10%"},
					{
						Nodes:    "0",
						Schedule: stringPtr("0 9 * * 1-5"),
						Duration: stringPtr("8h"),
						Reasons:  []string{"Underutilized", "Empty", "Drifted"},
					},
				},
			}
			input.Spec.Limits = map[string]string{
				"cpu":    "1000",
				"memory": "1000Gi",
			}
			input.Spec.Weight = int32Ptr(50)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("nodeClassRef name as a foreign-key reference should be valid", func() {
			input.Spec.Template.NodeClassRef.Name = valueFrom(
				catalogkind.CatalogKind_KubernetesKarpenterEc2NodeClass,
				"gp-nodeclass", "status.outputs.node_class_name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("static pool (replicas with only the nodes limit, no weight) should be valid", func() {
			input.Spec.Replicas = int64Ptr(3)
			input.Spec.Limits = map[string]string{"nodes": "10"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("static pool with no limits at all should be valid", func() {
			input.Spec.Replicas = int64Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("expire_after 'Never' and compound durations should be valid", func() {
			input.Spec.Template.ExpireAfter = stringPtr("Never")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.Template.ExpireAfter = stringPtr("1h30m")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("consolidate_after 'Never' (consolidation disabled) should be valid", func() {
			input.Spec.Disruption = &KubernetesKarpenterNodePoolDisruption{
				ConsolidateAfter: stringPtr("Never"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("budget with an @daily macro schedule should be valid", func() {
			input.Spec.Disruption = &KubernetesKarpenterNodePoolDisruption{
				Budgets: []*KubernetesKarpenterNodePoolDisruptionBudget{
					{Nodes: "2", Schedule: stringPtr("@daily"), Duration: stringPtr("1h30m")},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("requirement with min_values satisfied by the values list should be valid", func() {
			input.Spec.Template.Requirements = []*KubernetesKarpenterNodePoolRequirement{
				{
					Key:       "node.kubernetes.io/instance-type",
					Operator:  "In",
					Values:    []string{"m5.large", "m5.xlarge", "m6i.large"},
					MinValues: int32Ptr(2),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("numeric operators with a single non-negative integer value should be valid", func() {
			for _, operator := range []string{"Gt", "Lt", "Gte", "Lte"} {
				input.Spec.Template.Requirements = []*KubernetesKarpenterNodePoolRequirement{
					{Key: "karpenter.k8s.aws/instance-cpu", Operator: operator, Values: []string{"4"}},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("labels outside the karpenter.sh domain should be valid", func() {
			input.Spec.Template.Labels = map[string]string{
				"karpenter.k8s.aws/instance-family": "m5",
				"example.com/pool-tier":             "general",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Describe("static-mode rules", func() {
			ginkgo.It("replicas with a non-nodes limit should fail", func() {
				input.Spec.Replicas = int64Ptr(3)
				input.Spec.Limits = map[string]string{"cpu": "100"}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("replicas with nodes plus another limit should fail", func() {
				input.Spec.Replicas = int64Ptr(3)
				input.Spec.Limits = map[string]string{"nodes": "10", "cpu": "100"}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("replicas together with weight should fail", func() {
				input.Spec.Replicas = int64Ptr(3)
				input.Spec.Weight = int32Ptr(10)
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("negative replicas should fail", func() {
				input.Spec.Replicas = int64Ptr(-1)
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Describe("top-level spec fields", func() {
			ginkgo.It("missing template should fail", func() {
				input.Spec.Template = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("limit value that is not a Kubernetes quantity should fail", func() {
				input.Spec.Limits = map[string]string{"cpu": "lots"}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("weight of 0 should fail (gte=1)", func() {
				input.Spec.Weight = int32Ptr(0)
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("weight above 100 should fail (lte=100)", func() {
				input.Spec.Weight = int32Ptr(101)
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Describe("template label domain restrictions", func() {
			ginkgo.It("kubernetes.io/hostname label should fail", func() {
				input.Spec.Template.Labels = map[string]string{"kubernetes.io/hostname": "n1"}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("karpenter.sh/nodepool label should fail (controller-owned)", func() {
				input.Spec.Template.Labels = map[string]string{"karpenter.sh/nodepool": "mine"}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("arbitrary karpenter.sh-domain label should fail", func() {
				input.Spec.Template.Labels = map[string]string{"karpenter.sh/custom": "x"}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("karpenter.sh subdomain label should fail", func() {
				input.Spec.Template.Labels = map[string]string{"compute.karpenter.sh/custom": "x"}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("label value violating the label-value pattern should fail", func() {
				input.Spec.Template.Labels = map[string]string{"team": "-bad-"}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("more than 100 labels should fail (max_pairs=100)", func() {
				labels := map[string]string{}
				for i := 0; i < 101; i++ {
					labels[fmt.Sprintf("example.com/label-%d", i)] = "x"
				}
				input.Spec.Template.Labels = labels
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Describe("nodeClassRef", func() {
			ginkgo.It("missing nodeClassRef should fail", func() {
				input.Spec.Template.NodeClassRef = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("missing nodeClassRef name should fail", func() {
				input.Spec.Template.NodeClassRef.Name = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("explicitly empty group should fail", func() {
				input.Spec.Template.NodeClassRef.Group = stringPtr("")
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("group containing a slash should fail (pattern ^[^/]*$)", func() {
				input.Spec.Template.NodeClassRef.Group = stringPtr("karpenter.k8s.aws/v1")
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("explicitly empty kind should fail", func() {
				input.Spec.Template.NodeClassRef.Kind = stringPtr("")
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Describe("requirements", func() {
			ginkgo.It("zero requirements should fail (min_items=1)", func() {
				input.Spec.Template.Requirements = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("more than 100 requirements should fail (max_items=100)", func() {
				requirements := make([]*KubernetesKarpenterNodePoolRequirement, 0, 101)
				for i := 0; i < 101; i++ {
					requirements = append(requirements, capacityTypeRequirement())
				}
				input.Spec.Template.Requirements = requirements
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("operator 'In' with no values should fail", func() {
				input.Spec.Template.Requirements = []*KubernetesKarpenterNodePoolRequirement{
					{Key: "kubernetes.io/arch", Operator: "In"},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("numeric operator with two values should fail", func() {
				input.Spec.Template.Requirements = []*KubernetesKarpenterNodePoolRequirement{
					{Key: "karpenter.k8s.aws/instance-cpu", Operator: "Gt", Values: []string{"2", "4"}},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("numeric operator with a non-integer value should fail", func() {
				input.Spec.Template.Requirements = []*KubernetesKarpenterNodePoolRequirement{
					{Key: "karpenter.k8s.aws/instance-cpu", Operator: "Lt", Values: []string{"four"}},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("min_values greater than the number of values should fail", func() {
				input.Spec.Template.Requirements = []*KubernetesKarpenterNodePoolRequirement{
					{
						Key:       "node.kubernetes.io/instance-type",
						Operator:  "In",
						Values:    []string{"m5.large"},
						MinValues: int32Ptr(2),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("min_values of 0 should fail (gte=1)", func() {
				requirement := capacityTypeRequirement()
				requirement.MinValues = int32Ptr(0)
				input.Spec.Template.Requirements = []*KubernetesKarpenterNodePoolRequirement{requirement}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("min_values above 50 should fail (lte=50)", func() {
				requirement := capacityTypeRequirement()
				requirement.MinValues = int32Ptr(51)
				input.Spec.Template.Requirements = []*KubernetesKarpenterNodePoolRequirement{requirement}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("unknown operator should fail (closed enum)", func() {
				input.Spec.Template.Requirements = []*KubernetesKarpenterNodePoolRequirement{
					{Key: "kubernetes.io/arch", Operator: "Equals", Values: []string{"amd64"}},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("constraining kubernetes.io/hostname should fail", func() {
				input.Spec.Template.Requirements = []*KubernetesKarpenterNodePoolRequirement{
					{Key: "kubernetes.io/hostname", Operator: "In", Values: []string{"n1"}},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("constraining karpenter.sh/nodepool should fail", func() {
				input.Spec.Template.Requirements = []*KubernetesKarpenterNodePoolRequirement{
					{Key: "karpenter.sh/nodepool", Operator: "In", Values: []string{"other"}},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("arbitrary karpenter.sh-domain key should fail", func() {
				input.Spec.Template.Requirements = []*KubernetesKarpenterNodePoolRequirement{
					{Key: "karpenter.sh/custom", Operator: "Exists"},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("requirement value violating the label-value pattern should fail", func() {
				input.Spec.Template.Requirements = []*KubernetesKarpenterNodePoolRequirement{
					{Key: "kubernetes.io/arch", Operator: "In", Values: []string{"-bad-"}},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Describe("taints", func() {
			ginkgo.It("taint without a key should fail", func() {
				input.Spec.Template.Taints = []*KubernetesKarpenterNodePoolTaint{
					{Effect: "NoSchedule"},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("taint without an effect should fail", func() {
				input.Spec.Template.Taints = []*KubernetesKarpenterNodePoolTaint{
					{Key: "dedicated"},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("taint with an unknown effect should fail (closed enum)", func() {
				input.Spec.Template.Taints = []*KubernetesKarpenterNodePoolTaint{
					{Key: "dedicated", Effect: "Evict"},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("startup taint with an unknown effect should fail", func() {
				input.Spec.Template.StartupTaints = []*KubernetesKarpenterNodePoolTaint{
					{Key: "example.com/not-ready", Effect: "Sometimes"},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("taint value violating the value pattern should fail", func() {
				input.Spec.Template.Taints = []*KubernetesKarpenterNodePoolTaint{
					{Key: "dedicated", Value: "-bad-", Effect: "NoSchedule"},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Describe("node lifetime durations", func() {
			ginkgo.It("expire_after using unsupported units should fail", func() {
				input.Spec.Template.ExpireAfter = stringPtr("30d")
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("termination_grace_period 'Never' should fail (durations only)", func() {
				input.Spec.Template.TerminationGracePeriod = stringPtr("Never")
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("termination_grace_period using unsupported units should fail", func() {
				input.Spec.Template.TerminationGracePeriod = stringPtr("2d")
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Describe("disruption", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.Disruption = &KubernetesKarpenterNodePoolDisruption{}
			})

			ginkgo.It("unknown consolidation_policy should fail (closed enum)", func() {
				input.Spec.Disruption.ConsolidationPolicy = stringPtr("WhenIdle")
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("consolidate_after using unsupported units should fail", func() {
				input.Spec.Disruption.ConsolidateAfter = stringPtr("1d")
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("budget without nodes should fail (required)", func() {
				input.Spec.Disruption.Budgets = []*KubernetesKarpenterNodePoolDisruptionBudget{{}}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("budget nodes above 100% should fail the pattern", func() {
				input.Spec.Disruption.Budgets = []*KubernetesKarpenterNodePoolDisruptionBudget{
					{Nodes: "150%"},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("non-numeric budget nodes should fail the pattern", func() {
				input.Spec.Disruption.Budgets = []*KubernetesKarpenterNodePoolDisruptionBudget{
					{Nodes: "ten"},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("budget schedule without duration should fail (pairing)", func() {
				input.Spec.Disruption.Budgets = []*KubernetesKarpenterNodePoolDisruptionBudget{
					{Nodes: "0", Schedule: stringPtr("@daily")},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("budget duration without schedule should fail (pairing)", func() {
				input.Spec.Disruption.Budgets = []*KubernetesKarpenterNodePoolDisruptionBudget{
					{Nodes: "0", Duration: stringPtr("1h")},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("budget schedule with too few cron fields should fail the pattern", func() {
				input.Spec.Disruption.Budgets = []*KubernetesKarpenterNodePoolDisruptionBudget{
					{Nodes: "0", Schedule: stringPtr("* * *"), Duration: stringPtr("1h")},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("budget duration in seconds should fail (minutes/hours only)", func() {
				input.Spec.Disruption.Budgets = []*KubernetesKarpenterNodePoolDisruptionBudget{
					{Nodes: "0", Schedule: stringPtr("@daily"), Duration: stringPtr("90s")},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("unknown budget reason should fail (closed enum)", func() {
				input.Spec.Disruption.Budgets = []*KubernetesKarpenterNodePoolDisruptionBudget{
					{Nodes: "10%", Reasons: []string{"Expired"}},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("more than 50 budgets should fail (max_items=50)", func() {
				budgets := make([]*KubernetesKarpenterNodePoolDisruptionBudget, 0, 51)
				for i := 0; i < 51; i++ {
					budgets = append(budgets, &KubernetesKarpenterNodePoolDisruptionBudget{Nodes: "10%"})
				}
				input.Spec.Disruption.Budgets = budgets
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
