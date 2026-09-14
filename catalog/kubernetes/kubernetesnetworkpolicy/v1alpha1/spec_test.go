package kubernetesnetworkpolicyv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestKubernetesNetworkPolicySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesNetworkPolicySpec Validation Suite")
}

func protocol(p KubernetesNetworkPolicyPort_KubernetesNetworkPolicyProtocol) *KubernetesNetworkPolicyPort_KubernetesNetworkPolicyProtocol {
	return &p
}

var _ = ginkgo.Describe("KubernetesNetworkPolicySpec validations", func() {

	ginkgo.Context("When valid specs are provided", func() {

		ginkgo.It("accepts a minimal spec (all pods, inferred types)", func() {
			spec := &KubernetesNetworkPolicySpec{Name: "isolate"}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts match_all selectors at the policy and in a peer (the default-deny and same-namespace shapes)", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name:        "same-namespace",
				PodSelector: &KubernetesNetworkPolicyLabelSelector{MatchAll: proto.Bool(true)},
				IngressRules: []*KubernetesNetworkPolicyIngressRule{{
					From: []*KubernetesNetworkPolicyPeer{{
						PodSelector: &KubernetesNetworkPolicyLabelSelector{MatchAll: proto.Bool(true)},
					}},
				}},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a namespace provided as a resource reference", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "isolate",
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{Name: "team-namespace"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts the default-deny-all shape", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "default-deny-all",
				PolicyTypes: []KubernetesNetworkPolicySpec_KubernetesNetworkPolicyType{
					KubernetesNetworkPolicySpec_ingress,
					KubernetesNetworkPolicySpec_egress,
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a pod-selector ingress allow", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "allow-frontend",
				PodSelector: &KubernetesNetworkPolicyLabelSelector{
					MatchLabels: map[string]string{"app": "backend"},
				},
				IngressRules: []*KubernetesNetworkPolicyIngressRule{{
					From: []*KubernetesNetworkPolicyPeer{{
						PodSelector: &KubernetesNetworkPolicyLabelSelector{
							MatchLabels: map[string]string{"app": "frontend"},
						},
					}},
					Ports: []*KubernetesNetworkPolicyPort{{Port: "8080"}},
				}},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an AND'd pod+namespace selector peer", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "allow-monitoring",
				IngressRules: []*KubernetesNetworkPolicyIngressRule{{
					From: []*KubernetesNetworkPolicyPeer{{
						NamespaceSelector: &KubernetesNetworkPolicyLabelSelector{
							MatchLabels: map[string]string{"kubernetes.io/metadata.name": "monitoring"},
						},
						PodSelector: &KubernetesNetworkPolicyLabelSelector{
							MatchExpressions: []*KubernetesNetworkPolicyLabelSelectorRequirement{{
								Key:      "app",
								Operator: "In",
								Values:   []string{"prometheus", "grafana"},
							}},
						},
					}},
				}},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an IP block with exceptions and a port range", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "allow-egress-world",
				PolicyTypes: []KubernetesNetworkPolicySpec_KubernetesNetworkPolicyType{
					KubernetesNetworkPolicySpec_egress,
				},
				EgressRules: []*KubernetesNetworkPolicyEgressRule{{
					To: []*KubernetesNetworkPolicyPeer{{
						IpBlock: &KubernetesNetworkPolicyIpBlock{
							Cidr:   "0.0.0.0/0",
							Except: []string{"169.254.169.254/32"},
						},
					}},
					Ports: []*KubernetesNetworkPolicyPort{{
						Port:    "30000",
						EndPort: 32767,
					}},
				}},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a named port and an Exists expression", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "allow-metrics",
				PodSelector: &KubernetesNetworkPolicyLabelSelector{
					MatchExpressions: []*KubernetesNetworkPolicyLabelSelectorRequirement{{
						Key:      "app",
						Operator: "Exists",
					}},
				},
				IngressRules: []*KubernetesNetworkPolicyIngressRule{{
					Ports: []*KubernetesNetworkPolicyPort{{
						Protocol: protocol(KubernetesNetworkPolicyPort_TCP),
						Port:     "metrics",
					}},
				}},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})
	})

	ginkgo.Context("When invalid specs are provided", func() {

		ginkgo.It("rejects an empty name", func() {
			spec := &KubernetesNetworkPolicySpec{}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects duplicate policy types", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "bad",
				PolicyTypes: []KubernetesNetworkPolicySpec_KubernetesNetworkPolicyType{
					KubernetesNetworkPolicySpec_ingress,
					KubernetesNetworkPolicySpec_ingress,
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects ingress rules under an egress-only policy", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "bad",
				PolicyTypes: []KubernetesNetworkPolicySpec_KubernetesNetworkPolicyType{
					KubernetesNetworkPolicySpec_egress,
				},
				IngressRules: []*KubernetesNetworkPolicyIngressRule{{}},
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects egress rules under an ingress-only policy", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "bad",
				PolicyTypes: []KubernetesNetworkPolicySpec_KubernetesNetworkPolicyType{
					KubernetesNetworkPolicySpec_ingress,
				},
				EgressRules: []*KubernetesNetworkPolicyEgressRule{{}},
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a peer combining ip_block with a selector", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "bad",
				IngressRules: []*KubernetesNetworkPolicyIngressRule{{
					From: []*KubernetesNetworkPolicyPeer{{
						IpBlock: &KubernetesNetworkPolicyIpBlock{Cidr: "10.0.0.0/8"},
						PodSelector: &KubernetesNetworkPolicyLabelSelector{
							MatchLabels: map[string]string{"app": "x"},
						},
					}},
				}},
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a selector that names neither match_all nor a label criterion", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name:        "bad",
				PodSelector: &KubernetesNetworkPolicyLabelSelector{},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("must say what it selects"))
		})

		ginkgo.It("rejects match_all combined with labels", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "bad",
				PodSelector: &KubernetesNetworkPolicyLabelSelector{
					MatchAll:    proto.Bool(true),
					MatchLabels: map[string]string{"app": "x"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("match_all: true selects everything"))
		})

		ginkgo.It("rejects an empty peer", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "bad",
				IngressRules: []*KubernetesNetworkPolicyIngressRule{{
					From: []*KubernetesNetworkPolicyPeer{{}},
				}},
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid CIDR", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "bad",
				IngressRules: []*KubernetesNetworkPolicyIngressRule{{
					From: []*KubernetesNetworkPolicyPeer{{
						IpBlock: &KubernetesNetworkPolicyIpBlock{Cidr: "10.0.0.5"},
					}},
				}},
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid except CIDR", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "bad",
				IngressRules: []*KubernetesNetworkPolicyIngressRule{{
					From: []*KubernetesNetworkPolicyPeer{{
						IpBlock: &KubernetesNetworkPolicyIpBlock{
							Cidr:   "10.0.0.0/8",
							Except: []string{"not-a-cidr"},
						},
					}},
				}},
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a numeric port out of range", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "bad",
				IngressRules: []*KubernetesNetworkPolicyIngressRule{{
					Ports: []*KubernetesNetworkPolicyPort{{Port: "70000"}},
				}},
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an end_port anchored on a named port", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "bad",
				IngressRules: []*KubernetesNetworkPolicyIngressRule{{
					Ports: []*KubernetesNetworkPolicyPort{{
						Port:    "metrics",
						EndPort: 9999,
					}},
				}},
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an end_port below the anchor port", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "bad",
				IngressRules: []*KubernetesNetworkPolicyIngressRule{{
					Ports: []*KubernetesNetworkPolicyPort{{
						Port:    "9000",
						EndPort: 8000,
					}},
				}},
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown selector operator", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "bad",
				PodSelector: &KubernetesNetworkPolicyLabelSelector{
					MatchExpressions: []*KubernetesNetworkPolicyLabelSelectorRequirement{{
						Key:      "app",
						Operator: "Equals",
						Values:   []string{"x"},
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects In without values", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "bad",
				PodSelector: &KubernetesNetworkPolicyLabelSelector{
					MatchExpressions: []*KubernetesNetworkPolicyLabelSelectorRequirement{{
						Key:      "app",
						Operator: "In",
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects Exists with values", func() {
			spec := &KubernetesNetworkPolicySpec{
				Name: "bad",
				PodSelector: &KubernetesNetworkPolicyLabelSelector{
					MatchExpressions: []*KubernetesNetworkPolicyLabelSelectorRequirement{{
						Key:      "app",
						Operator: "Exists",
						Values:   []string{"x"},
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})
	})
})
