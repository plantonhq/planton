package digitaloceanuptimecheckv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
)

func TestDigitalOceanUptimeCheckSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanUptimeCheckSpec Validation Suite")
}

var _ = ginkgo.Describe("DigitalOceanUptimeCheckSpec validations", func() {

	makeValidSpec := func() *DigitalOceanUptimeCheckSpec {
		return &DigitalOceanUptimeCheckSpec{
			CheckName: "homepage",
			Target:    "https://www.example.com",
			Regions:   []string{"us_east", "eu_west"},
		}
	}

	makeValidAlert := func() *DigitalOceanUptimeCheckAlert {
		return &DigitalOceanUptimeCheckAlert{
			AlertName: "homepage-down",
			Type:      "down",
			Period:    "2m",
			Notifications: &DigitalOceanUptimeCheckNotifications{
				Emails: []string{"ops@example.com"},
			},
		}
	}

	ginkgo.Context("Required fields", func() {
		ginkgo.It("accepts a minimal valid spec", func() {
			err := protovalidate.Validate(makeValidSpec())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing check_name", func() {
			spec := makeValidSpec()
			spec.CheckName = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing target", func() {
			spec := makeValidSpec()
			spec.Target = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects spec with no regions (the defaulted-regions drift class)", func() {
			spec := makeValidSpec()
			spec.Regions = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Type and regions", func() {
		ginkgo.It("accepts each probe type", func() {
			for _, typ := range []string{"ping", "http", "https"} {
				spec := makeValidSpec()
				spec.Type = typ
				if typ == "ping" {
					spec.Target = "www.example.com"
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil(), "type %q should validate", typ)
			}
		})

		ginkgo.It("rejects an unknown probe type", func() {
			spec := makeValidSpec()
			spec.Type = "tcp"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts all four vantage regions", func() {
			spec := makeValidSpec()
			spec.Regions = []string{"us_east", "us_west", "eu_west", "se_asia"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown region", func() {
			spec := makeValidSpec()
			spec.Regions = []string{"ap_south"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Alert rules", func() {
		ginkgo.It("accepts a down alert with no threshold or comparison (the API fixes both)", func() {
			spec := makeValidSpec()
			spec.Alerts = []*DigitalOceanUptimeCheckAlert{makeValidAlert()}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a down_global alert with no threshold or comparison", func() {
			spec := makeValidSpec()
			alert := makeValidAlert()
			alert.Type = "down_global"
			spec.Alerts = []*DigitalOceanUptimeCheckAlert{alert}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects a down alert that sets threshold (the API would overwrite it with 1)", func() {
			spec := makeValidSpec()
			alert := makeValidAlert()
			alert.Threshold = proto.Int32(3)
			spec.Alerts = []*DigitalOceanUptimeCheckAlert{alert}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a down_global alert that sets comparison (the API would overwrite it with less_than)", func() {
			spec := makeValidSpec()
			alert := makeValidAlert()
			alert.Type = "down_global"
			alert.Comparison = "greater_than"
			spec.Alerts = []*DigitalOceanUptimeCheckAlert{alert}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects any alert without a period (the API requires it for every type)", func() {
			for _, typ := range []string{"down", "latency", "ssl_expiry"} {
				spec := makeValidSpec()
				alert := makeValidAlert()
				alert.Type = typ
				alert.Period = ""
				if typ != "down" {
					alert.Threshold = proto.Int32(300)
				}
				if typ == "latency" {
					alert.Comparison = "greater_than"
				}
				spec.Alerts = []*DigitalOceanUptimeCheckAlert{alert}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil(), "type %q without period should be rejected", typ)
			}
		})

		ginkgo.It("accepts a latency alert with a threshold and a comparison", func() {
			spec := makeValidSpec()
			alert := makeValidAlert()
			alert.Type = "latency"
			alert.Threshold = proto.Int32(300)
			alert.Comparison = "greater_than"
			spec.Alerts = []*DigitalOceanUptimeCheckAlert{alert}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects a latency alert without a threshold", func() {
			spec := makeValidSpec()
			alert := makeValidAlert()
			alert.Type = "latency"
			alert.Comparison = "greater_than"
			spec.Alerts = []*DigitalOceanUptimeCheckAlert{alert}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a latency alert without a comparison (DigitalOcean would pick a direction silently)", func() {
			spec := makeValidSpec()
			alert := makeValidAlert()
			alert.Type = "latency"
			alert.Threshold = proto.Int32(300)
			spec.Alerts = []*DigitalOceanUptimeCheckAlert{alert}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts an ssl_expiry alert with a day threshold and no comparison", func() {
			spec := makeValidSpec()
			alert := makeValidAlert()
			alert.Type = "ssl_expiry"
			alert.Threshold = proto.Int32(14)
			spec.Alerts = []*DigitalOceanUptimeCheckAlert{alert}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects an ssl_expiry alert without a threshold (0 days is a post-mortem)", func() {
			spec := makeValidSpec()
			alert := makeValidAlert()
			alert.Type = "ssl_expiry"
			spec.Alerts = []*DigitalOceanUptimeCheckAlert{alert}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an ssl_expiry alert that sets comparison (the API always evaluates less_than)", func() {
			spec := makeValidSpec()
			alert := makeValidAlert()
			alert.Type = "ssl_expiry"
			alert.Threshold = proto.Int32(14)
			alert.Comparison = "less_than"
			spec.Alerts = []*DigitalOceanUptimeCheckAlert{alert}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown alert type", func() {
			spec := makeValidSpec()
			alert := makeValidAlert()
			alert.Type = "response_time"
			spec.Alerts = []*DigitalOceanUptimeCheckAlert{alert}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects the CamelCase comparison spelling (that is the monitor API's)", func() {
			spec := makeValidSpec()
			alert := makeValidAlert()
			alert.Type = "latency"
			alert.Threshold = proto.Int32(300)
			alert.Comparison = "GreaterThan"
			spec.Alerts = []*DigitalOceanUptimeCheckAlert{alert}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown period", func() {
			spec := makeValidSpec()
			alert := makeValidAlert()
			alert.Period = "45m"
			spec.Alerts = []*DigitalOceanUptimeCheckAlert{alert}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an alert without notifications", func() {
			spec := makeValidSpec()
			alert := makeValidAlert()
			alert.Notifications = nil
			spec.Alerts = []*DigitalOceanUptimeCheckAlert{alert}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects notifications with no channel at all", func() {
			spec := makeValidSpec()
			alert := makeValidAlert()
			alert.Notifications = &DigitalOceanUptimeCheckNotifications{}
			spec.Alerts = []*DigitalOceanUptimeCheckAlert{alert}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts slack as the only notification channel", func() {
			spec := makeValidSpec()
			alert := makeValidAlert()
			alert.Notifications = &DigitalOceanUptimeCheckNotifications{
				Slack: []*DigitalOceanUptimeCheckSlack{{
					Channel: "#alerts",
					Url:     "https://hooks.slack.com/services/T0/B0/XXXX",
				}},
			}
			spec.Alerts = []*DigitalOceanUptimeCheckAlert{alert}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})
