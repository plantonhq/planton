package kubernetesgatewayclassv1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
)

func TestKubernetesGatewayClass(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesGatewayClass Suite")
}

var _ = ginkgo.Describe("KubernetesGatewayClass Validation Tests", func() {
	var input *KubernetesGatewayClass

	ginkgo.BeforeEach(func() {
		input = &KubernetesGatewayClass{
			ApiVersion: "kubernetes.openmcf.org/v1",
			Kind:       "KubernetesGatewayClass",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-gateway-class",
			},
			Spec: &KubernetesGatewayClassSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				ControllerName: "istio.io/gateway-controller",
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with only the required controller_name", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with the Envoy Gateway controller name", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ControllerName = "gateway.envoyproxy.io/gatewayclass-controller"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with an optional description set", func() {
			ginkgo.It("should not return a validation error", func() {
				description := "Primary ingress gateway class for production"
				input.Spec.Description = &description
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a namespaced parameters_ref set", func() {
			ginkgo.It("should not return a validation error", func() {
				namespace := "istio-system"
				input.Spec.ParametersRef = &kubernetes.KubernetesGatewayApiParametersReference{
					Group:     "",
					Kind:      "ConfigMap",
					Name:      "istio-gateway-config",
					Namespace: &namespace,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a cluster-scoped parameters_ref (no namespace)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ParametersRef = &kubernetes.KubernetesGatewayApiParametersReference{
					Group: "gateway.envoyproxy.io",
					Kind:  "EnvoyProxy",
					Name:  "custom-proxy-config",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a maximum-length (64 char) description", func() {
			ginkgo.It("should not return a validation error", func() {
				description := strings.Repeat("a", 64)
				input.Spec.Description = &description
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing controller_name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ControllerName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("controller_name without a domain-prefixed path (no slash)", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ControllerName = "istio"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("controller_name with an uppercase domain segment", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ControllerName = "Istio.io/gateway-controller"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("controller_name longer than 253 characters", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ControllerName = "example.io/" + strings.Repeat("a", 250)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("description longer than 64 characters", func() {
			ginkgo.It("should return a validation error", func() {
				description := strings.Repeat("a", 65)
				input.Spec.Description = &description
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("parameters_ref missing the required name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ParametersRef = &kubernetes.KubernetesGatewayApiParametersReference{
					Group: "",
					Kind:  "ConfigMap",
					Name:  "",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing spec", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing metadata", func() {
			ginkgo.It("should return a validation error", func() {
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
