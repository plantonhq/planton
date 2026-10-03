package gcpmonitoringslov1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpMonitoringSloSpec Suite")
}

// goodTotalRatioSli builds the workhorse SLI shape used across cases.
func goodTotalRatioSli() *GcpMonitoringSloSli {
	return &GcpMonitoringSloSli{
		RequestBasedSli: &GcpMonitoringSloRequestBasedSli{
			GoodTotalRatio: &GcpMonitoringSloGoodTotalRatio{
				GoodServiceFilter:  `metric.type="serviceruntime.googleapis.com/api/request_count" metric.labels.response_code_class="2xx"`,
				TotalServiceFilter: `metric.type="serviceruntime.googleapis.com/api/request_count"`,
			},
		},
	}
}

var _ = ginkgo.Describe("GcpMonitoringSloSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpMonitoringSlo {
		return &GcpMonitoringSlo{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpMonitoringSlo",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-slo",
			},
			Spec: &GcpMonitoringSloSpec{
				Service: &GcpMonitoringSloService{
					CustomService: &GcpMonitoringSloCustomService{},
				},
				Goal:              0.999,
				RollingPeriodDays: 30,
				Sli:               goodTotalRatioSli(),
			},
		}
	}

	ginkgo.Context("minimal manifest", func() {
		ginkgo.It("passes validation", func() {
			gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
		})
	})

	ginkgo.Context("goal", func() {
		ginkgo.It("rejects zero and values above 0.9999", func() {
			for _, v := range []float64{0, 1, 0.99991} {
				s := minimal()
				s.Spec.Goal = v
				gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "goal %v", v)
			}
		})

		ginkgo.It("accepts the API's upper bound 0.9999", func() {
			s := minimal()
			s.Spec.Goal = 0.9999
			gomega.Expect(validator.Validate(s)).To(gomega.Succeed())
		})
	})

	ginkgo.Context("period", func() {
		ginkgo.It("requires exactly one of calendar_period and rolling_period_days", func() {
			s := minimal()
			s.Spec.RollingPeriodDays = 0
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "neither period")

			s = minimal()
			s.Spec.CalendarPeriod = "MONTH"
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "both periods")

			s = minimal()
			s.Spec.RollingPeriodDays = 0
			s.Spec.CalendarPeriod = "WEEK"
			gomega.Expect(validator.Validate(s)).To(gomega.Succeed(), "calendar only")
		})

		ginkgo.It("rejects unknown calendar periods and out-of-range rolling windows", func() {
			s := minimal()
			s.Spec.RollingPeriodDays = 0
			s.Spec.CalendarPeriod = "QUARTER"
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed())

			s = minimal()
			s.Spec.RollingPeriodDays = 31
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("service", func() {
		ginkgo.It("requires exactly one arm", func() {
			s := minimal()
			s.Spec.Service = &GcpMonitoringSloService{}
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "no arm")

			s = minimal()
			s.Spec.Service.ServiceId = "existing-service"
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "two arms")

			s = minimal()
			s.Spec.Service = &GcpMonitoringSloService{ServiceId: "existing-service"}
			gomega.Expect(validator.Validate(s)).To(gomega.Succeed(), "existing-service arm")

			s = minimal()
			s.Spec.Service = &GcpMonitoringSloService{
				BasicService: &GcpMonitoringSloBasicService{
					ServiceType:   "CLOUD_RUN",
					ServiceLabels: map[string]string{"service_name": "checkout", "location": "us-central1"},
				},
			}
			gomega.Expect(validator.Validate(s)).To(gomega.Succeed(), "basic-service arm")
		})
	})

	ginkgo.Context("slo_id", func() {
		ginkgo.It("accepts the API's documented charset and rejects others", func() {
			s := minimal()
			s.Spec.SloId = "checkout-availability_v1:prod.slo"
			gomega.Expect(validator.Validate(s)).To(gomega.Succeed())

			s = minimal()
			s.Spec.SloId = "has spaces"
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("sli", func() {
		ginkgo.It("requires exactly one family", func() {
			s := minimal()
			s.Spec.Sli = &GcpMonitoringSloSli{}
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "no family")

			s = minimal()
			s.Spec.Sli.BasicSli = &GcpMonitoringSloBasicSli{
				Availability: &GcpMonitoringSloAvailability{},
			}
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "two families")
		})

		ginkgo.It("basic_sli requires exactly one of availability and latency", func() {
			s := minimal()
			s.Spec.Sli = &GcpMonitoringSloSli{BasicSli: &GcpMonitoringSloBasicSli{}}
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "neither arm")

			s.Spec.Sli.BasicSli.Availability = &GcpMonitoringSloAvailability{}
			s.Spec.Sli.BasicSli.Latency = &GcpMonitoringSloLatency{Threshold: "1s"}
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "both arms")

			s.Spec.Sli.BasicSli.Availability = nil
			gomega.Expect(validator.Validate(s)).To(gomega.Succeed(), "latency arm")
		})

		ginkgo.It("request_based_sli requires exactly one of distribution_cut and good_total_ratio", func() {
			s := minimal()
			s.Spec.Sli.RequestBasedSli.DistributionCut = &GcpMonitoringSloDistributionCut{
				DistributionFilter: `metric.type="loadbalancing.googleapis.com/https/total_latencies"`,
			}
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "both arms")

			s.Spec.Sli.RequestBasedSli.GoodTotalRatio = nil
			gomega.Expect(validator.Validate(s)).To(gomega.Succeed(), "distribution arm")

			s.Spec.Sli.RequestBasedSli.DistributionCut = nil
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "neither arm")
		})

		ginkgo.It("distribution_cut requires its filter", func() {
			s := minimal()
			s.Spec.Sli = &GcpMonitoringSloSli{
				RequestBasedSli: &GcpMonitoringSloRequestBasedSli{
					DistributionCut: &GcpMonitoringSloDistributionCut{},
				},
			}
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed())
		})

		ginkgo.It("good_total_ratio requires exactly two filters", func() {
			s := minimal()
			s.Spec.Sli.RequestBasedSli.GoodTotalRatio.BadServiceFilter = "bad"
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "three filters")

			s.Spec.Sli.RequestBasedSli.GoodTotalRatio.TotalServiceFilter = ""
			gomega.Expect(validator.Validate(s)).To(gomega.Succeed(), "good+bad")

			s.Spec.Sli.RequestBasedSli.GoodTotalRatio.BadServiceFilter = ""
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "one filter")
		})

		ginkgo.It("windows_based_sli requires exactly one window criterion", func() {
			s := minimal()
			s.Spec.Sli = &GcpMonitoringSloSli{WindowsBasedSli: &GcpMonitoringSloWindowsBasedSli{}}
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "no criterion")

			s.Spec.Sli.WindowsBasedSli.GoodBadMetricFilter = `metric.type="monitoring.googleapis.com/uptime_check/check_passed"`
			gomega.Expect(validator.Validate(s)).To(gomega.Succeed(), "good_bad_metric_filter")

			s.Spec.Sli.WindowsBasedSli.MetricMeanInRange = &GcpMonitoringSloMetricRange{
				TimeSeries: `metric.type="agent.googleapis.com/cpu/utilization"`,
				Range:      &GcpMonitoringSloRange{Max: proto.Float64(80)},
			}
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "two criteria")
		})

		ginkgo.It("good_total_ratio_threshold requires exactly one of basic_sli_performance and performance", func() {
			s := minimal()
			s.Spec.Sli = &GcpMonitoringSloSli{
				WindowsBasedSli: &GcpMonitoringSloWindowsBasedSli{
					GoodTotalRatioThreshold: &GcpMonitoringSloGoodTotalRatioThreshold{
						Threshold: 0.95,
					},
				},
			}
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "neither arm")

			s.Spec.Sli.WindowsBasedSli.GoodTotalRatioThreshold.Performance = &GcpMonitoringSloPerformance{
				GoodTotalRatio: &GcpMonitoringSloGoodTotalRatio{
					GoodServiceFilter:  "good",
					TotalServiceFilter: "total",
				},
			}
			gomega.Expect(validator.Validate(s)).To(gomega.Succeed(), "performance arm")

			s.Spec.Sli.WindowsBasedSli.GoodTotalRatioThreshold.Threshold = 1.5
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed(), "threshold above 1")
		})
	})

	ginkgo.Context("deletion_policy", func() {
		ginkgo.It("accepts the documented values and rejects others", func() {
			for _, v := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
				s := minimal()
				s.Spec.DeletionPolicy = v
				gomega.Expect(validator.Validate(s)).To(gomega.Succeed(), "value %q", v)
			}
			s := minimal()
			s.Spec.DeletionPolicy = "KEEP"
			gomega.Expect(validator.Validate(s)).ToNot(gomega.Succeed())
		})
	})
})
