package gcplogmetricv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpLogMetricSpec Suite")
}

var _ = ginkgo.Describe("GcpLogMetricSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpLogMetric {
		return &GcpLogMetric{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpLogMetric",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-log-metric",
			},
			Spec: &GcpLogMetricSpec{
				Filter: `resource.type="cloud_run_revision" AND severity>=ERROR`,
			},
		}
	}

	withDescriptor := func() *GcpLogMetric {
		m := minimal()
		m.Spec.MetricDescriptor = &GcpLogMetricDescriptor{
			MetricKind: "DELTA",
			ValueType:  "INT64",
		}
		return m
	}

	ginkgo.Context("minimal manifest", func() {
		ginkgo.It("passes validation", func() {
			gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
		})
	})

	ginkgo.Context("filter", func() {
		ginkgo.It("is required", func() {
			m := minimal()
			m.Spec.Filter = ""
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("description", func() {
		ginkgo.It("enforces the provider's 8000-character cap", func() {
			m := minimal()
			m.Spec.Description = strings.Repeat("a", 8000)
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())

			m.Spec.Description = strings.Repeat("a", 8001)
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("metric_descriptor", func() {
		ginkgo.It("requires provider-valid kind and value type", func() {
			gomega.Expect(validator.Validate(withDescriptor())).To(gomega.Succeed())

			m := withDescriptor()
			m.Spec.MetricDescriptor.MetricKind = "RATE"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "unknown kind")

			m = withDescriptor()
			m.Spec.MetricDescriptor.ValueType = "FLOAT"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "unknown value type")

			m = withDescriptor()
			m.Spec.MetricDescriptor.ValueType = "DISTRIBUTION"
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())
		})

		ginkgo.It("labels require a key and a valid value type", func() {
			m := withDescriptor()
			m.Spec.MetricDescriptor.Labels = []*GcpLogMetricLabel{{Description: "status class"}}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "missing key")

			m.Spec.MetricDescriptor.Labels[0].Key = "status"
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())

			m.Spec.MetricDescriptor.Labels[0].ValueType = "DOUBLE"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "DOUBLE is not a label value type")

			m.Spec.MetricDescriptor.Labels[0].ValueType = "INT64"
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())
		})
	})

	ginkgo.Context("bucket_options", func() {
		ginkgo.It("requires at least one layout", func() {
			m := withDescriptor()
			m.Spec.BucketOptions = &GcpLogMetricBucketOptions{}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "no layout")

			m.Spec.BucketOptions.ExponentialBuckets = &GcpLogMetricExponentialBuckets{
				NumFiniteBuckets: 64,
				GrowthFactor:     2,
				Scale:            0.01,
			}
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())
		})

		ginkgo.It("explicit buckets need at least one bound; finite bucket counts must be positive", func() {
			m := withDescriptor()
			m.Spec.BucketOptions = &GcpLogMetricBucketOptions{
				ExplicitBuckets: &GcpLogMetricExplicitBuckets{},
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "no bounds")

			m.Spec.BucketOptions.ExplicitBuckets.Bounds = []float64{0.1, 0.5, 1, 5}
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())

			m.Spec.BucketOptions = &GcpLogMetricBucketOptions{
				LinearBuckets: &GcpLogMetricLinearBuckets{NumFiniteBuckets: 0, Offset: 0, Width: 10},
			}
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "zero finite buckets")
		})
	})

	ginkgo.Context("deletion_policy", func() {
		ginkgo.It("accepts the documented values and rejects others", func() {
			for _, v := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
				m := minimal()
				m.Spec.DeletionPolicy = v
				gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "value %q", v)
			}
			m := minimal()
			m.Spec.DeletionPolicy = "KEEP"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})
})
