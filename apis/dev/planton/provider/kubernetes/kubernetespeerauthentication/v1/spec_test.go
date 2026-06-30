package kubernetespeerauthenticationv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesPeerAuthentication(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesPeerAuthentication Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind cloudresourcekind.CloudResourceKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
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

// mtls returns a MutualTLS block with the given mode.
func mtls(mode string) *KubernetesPeerAuthenticationMutualTls {
	return &KubernetesPeerAuthenticationMutualTls{Mode: mode}
}

var _ = ginkgo.Describe("KubernetesPeerAuthentication Validation Tests", func() {
	var input *KubernetesPeerAuthentication

	ginkgo.BeforeEach(func() {
		input = &KubernetesPeerAuthentication{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesPeerAuthentication",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-peer-authentication",
			},
			Spec: &KubernetesPeerAuthenticationSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: literal("default"),
				Mtls:      mtls("STRICT"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with a namespace-wide STRICT mtls policy (no selector)", func() {
			ginkgo.It("should not return a validation error", func() {
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with no mtls block at all (inherit from parent)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Mtls = nil
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with each valid mtls mode", func() {
			ginkgo.It("should accept UNSET, DISABLE, PERMISSIVE, and STRICT", func() {
				for _, mode := range []string{"UNSET", "DISABLE", "PERMISSIVE", "STRICT"} {
					input.Spec.Mtls = mtls(mode)
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "mode %s should be valid", mode)
				}
			})
		})

		ginkgo.Context("with a selector targeting specific workloads", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{
					MatchLabels: map[string]string{"app": "finance"},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with port_level_mtls overrides plus a selector", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{
					MatchLabels: map[string]string{"app": "finance"},
				}
				input.Spec.PortLevelMtls = map[uint32]*KubernetesPeerAuthenticationMutualTls{
					8080: mtls("DISABLE"),
					9090: mtls("STRICT"),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with namespace resolved via a valueFrom reference", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Namespace = valueFrom(cloudresourcekind.CloudResourceKind_KubernetesNamespace, "finance-ns", "spec.name")
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with boundary port numbers in port_level_mtls", func() {
			ginkgo.It("should accept ports 1 and 65535", func() {
				input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{
					MatchLabels: map[string]string{"app": "finance"},
				}
				input.Spec.PortLevelMtls = map[uint32]*KubernetesPeerAuthenticationMutualTls{
					1:     mtls("STRICT"),
					65535: mtls("PERMISSIVE"),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When the envelope is invalid", func() {
		ginkgo.Context("with the wrong api_version", func() {
			ginkgo.It("should return a validation error", func() {
				input.ApiVersion = "v1"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with the wrong kind", func() {
			ginkgo.It("should return a validation error", func() {
				input.Kind = "KubernetesGateway"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without metadata", func() {
			ginkgo.It("should return a validation error", func() {
				input.Metadata = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without a spec", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without a namespace", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Namespace = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When the mtls mode is invalid", func() {
		ginkgo.Context("with an empty mode string", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Mtls = mtls("")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an unknown mode value", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Mtls = mtls("MUTUAL")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a lowercase mode value (UPPERCASE is required)", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Mtls = mtls("strict")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an invalid mode in a port_level_mtls entry", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{
					MatchLabels: map[string]string{"app": "finance"},
				}
				input.Spec.PortLevelMtls = map[uint32]*KubernetesPeerAuthenticationMutualTls{
					8080: mtls("BOGUS"),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When port_level_mtls keys are out of range", func() {
		ginkgo.BeforeEach(func() {
			input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{
				MatchLabels: map[string]string{"app": "finance"},
			}
		})

		ginkgo.Context("with port 0", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.PortLevelMtls = map[uint32]*KubernetesPeerAuthenticationMutualTls{
					0: mtls("STRICT"),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a port above 65535", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.PortLevelMtls = map[uint32]*KubernetesPeerAuthenticationMutualTls{
					70000: mtls("STRICT"),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When port_level_mtls is set without a valid selector", func() {
		ginkgo.Context("with no selector at all", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Selector = nil
				input.Spec.PortLevelMtls = map[uint32]*KubernetesPeerAuthenticationMutualTls{
					8080: mtls("STRICT"),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a selector that has no match labels", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{}
				input.Spec.PortLevelMtls = map[uint32]*KubernetesPeerAuthenticationMutualTls{
					8080: mtls("STRICT"),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When the selector match labels are invalid", func() {
		ginkgo.Context("with an empty label key", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{
					MatchLabels: map[string]string{"": "finance"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a wildcard in a label key", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{
					MatchLabels: map[string]string{"app*": "finance"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a wildcard in a label value", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{
					MatchLabels: map[string]string{"app": "fin*"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
