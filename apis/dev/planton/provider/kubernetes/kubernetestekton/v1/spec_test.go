package kubernetestektonv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestKubernetesTekton(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesTekton Suite")
}

var _ = ginkgo.Describe("KubernetesTekton Custom Validation Tests", func() {
	var input *KubernetesTekton

	ginkgo.BeforeEach(func() {
		input = &KubernetesTekton{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesTekton",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-tekton",
			},
			Spec: &KubernetesTektonSpec{
				PipelineVersion: "latest",
				Dashboard: &KubernetesTektonDashboard{
					Enabled: true,
					Version: "latest",
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("basic configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with cloud events configured", func() {
			ginkgo.It("should not return a validation error for valid HTTP URL", func() {
				input.Spec.CloudEvents = &KubernetesTektonCloudEvents{
					SinkUrl: "http://my-service.my-namespace.svc.cluster.local/events",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for valid HTTPS URL", func() {
				input.Spec.CloudEvents = &KubernetesTektonCloudEvents{
					SinkUrl: "https://events.example.com/tekton",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for empty sink URL", func() {
				input.Spec.CloudEvents = &KubernetesTektonCloudEvents{
					SinkUrl: "",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with dashboard ingress enabled", func() {
			ginkgo.It("should not return a validation error when hostname is provided", func() {
				input.Spec.Dashboard.Ingress = &KubernetesTektonDashboardIngress{
					Enabled:  true,
					Hostname: "tekton-dashboard.example.com",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error when disabled without hostname", func() {
				input.Spec.Dashboard.Ingress = &KubernetesTektonDashboardIngress{
					Enabled:  false,
					Hostname: "",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("cloud events with invalid URL", func() {
			ginkgo.It("should return a validation error for non-HTTP URL", func() {
				input.Spec.CloudEvents = &KubernetesTektonCloudEvents{
					SinkUrl: "ftp://invalid-url.com/events",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("dashboard ingress enabled without hostname", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Dashboard.Ingress = &KubernetesTektonDashboardIngress{
					Enabled:  true,
					Hostname: "",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
