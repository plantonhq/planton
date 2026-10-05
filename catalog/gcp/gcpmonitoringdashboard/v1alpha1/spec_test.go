package gcpmonitoringdashboardv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpMonitoringDashboardSpec Suite")
}

var _ = ginkgo.Describe("GcpMonitoringDashboardSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpMonitoringDashboard {
		return &GcpMonitoringDashboard{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpMonitoringDashboard",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-dashboard",
			},
			Spec: &GcpMonitoringDashboardSpec{
				DashboardJson: `{"displayName":"API health","gridLayout":{"widgets":[{"title":"CPU","xyChart":{"dataSets":[{"timeSeriesQuery":{"timeSeriesFilter":{"filter":"metric.type=\"compute.googleapis.com/instance/cpu/utilization\""}}}]}}]}}`,
			},
		}
	}

	ginkgo.Context("minimal manifest", func() {
		ginkgo.It("passes validation", func() {
			gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
		})
	})

	ginkgo.Context("dashboard_json", func() {
		ginkgo.It("is required", func() {
			d := minimal()
			d.Spec.DashboardJson = ""
			gomega.Expect(validator.Validate(d)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("deletion_policy", func() {
		ginkgo.It("accepts empty and the three documented values", func() {
			for _, v := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
				d := minimal()
				d.Spec.DeletionPolicy = v
				gomega.Expect(validator.Validate(d)).To(gomega.Succeed(), "value %q", v)
			}
		})

		ginkgo.It("rejects unknown values", func() {
			d := minimal()
			d.Spec.DeletionPolicy = "RETAIN"
			gomega.Expect(validator.Validate(d)).ToNot(gomega.Succeed())
		})
	})
})
