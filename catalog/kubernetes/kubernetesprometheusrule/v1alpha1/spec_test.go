package kubernetesprometheusrulev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// The validation suite pins what the spec refuses before a rule file reaches
// Prometheus: the record-xor-alert shape, the alert-only fields a recording
// rule may not carry (Prometheus drops the whole PrometheusRule for those),
// unique group names (the list's key upstream), Prometheus duration and metric
// name grammar, and the Kubernetes label grammar on the object's own labels.

func TestKubernetesPrometheusRule(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesPrometheusRule Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind cloudresourcekind.CloudResourceKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name, FieldPath: fieldPath},
		},
	}
}

// ptr returns a pointer to v, for setting proto3 `optional` scalar fields.
func ptr[T any](v T) *T { return &v }

func alertRule() *KubernetesPrometheusRuleRule {
	return &KubernetesPrometheusRuleRule{
		Alert:       ptr("ApiErrorBudgetBurn"),
		Expr:        "job:slo_errors_per_request:ratio_rate1h > (14.4 * 0.001)",
		ForDuration: ptr("2m"),
		Labels:      map[string]string{"severity": "page"},
		Annotations: map[string]string{"summary": "The API is burning its error budget fast"},
	}
}

func recordRule() *KubernetesPrometheusRuleRule {
	return &KubernetesPrometheusRuleRule{
		Record: ptr("job:slo_errors_per_request:ratio_rate1h"),
		Expr:   "sum(rate(errors_total[1h])) / sum(rate(requests_total[1h]))",
	}
}

func expectViolation(err error, fragment string) {
	gomega.Expect(err).NotTo(gomega.BeNil())
	gomega.Expect(strings.Contains(err.Error(), fragment)).To(gomega.BeTrue(),
		"expected a violation mentioning %q, got: %v", fragment, err)
}

var _ = ginkgo.Describe("KubernetesPrometheusRule Validation Tests", func() {
	var input *KubernetesPrometheusRule

	ginkgo.BeforeEach(func() {
		input = &KubernetesPrometheusRule{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesPrometheusRule",
			Metadata:   &shared.CloudResourceMetadata{Name: "api-slo"},
			Spec: &KubernetesPrometheusRuleSpec{
				Namespace: literal("monitoring"),
				Groups: []*KubernetesPrometheusRuleGroup{{
					Name:     "api-availability",
					Interval: ptr("30s"),
					Rules:    []*KubernetesPrometheusRuleRule{recordRule(), alertRule()},
				}},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a group with a recording rule and an alerting rule", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a namespace given as a reference", func() {
			input.Spec.Namespace = valueFrom(cloudresourcekind.CloudResourceKind_KubernetesNamespace, "monitoring-ns", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts every group setting at its upstream shape", func() {
			g := input.Spec.Groups[0]
			g.QueryOffset = ptr("1m")
			g.Limit = ptr(int32(100))
			g.PartialResponseStrategy = ptr("WARN")
			g.Labels = map[string]string{"component": "api"}
			input.Spec.Groups[0].Rules[1].KeepFiringFor = ptr("10m")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an empty partial-response strategy, which upstream allows", func() {
			input.Spec.Groups[0].PartialResponseStrategy = ptr("")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts the object labels a fenced Prometheus selects by", func() {
			input.Spec.Labels = map[string]string{"release": "management-hub-metrics", "app.kubernetes.io/part-of": "observability"}
			input.Spec.Annotations = map[string]string{"planton.ai/owner": "platform"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a spec with no groups", func() {
			input.Spec.Groups = nil
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When the envelope is invalid", func() {
		ginkgo.It("refuses a missing namespace", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an object label key that is not a Kubernetes qualified name", func() {
			input.Spec.Labels = map[string]string{"not a key": "x"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an object label value longer than 63 characters", func() {
			input.Spec.Labels = map[string]string{"release": strings.Repeat("a", 64)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When groups collide or are malformed", func() {
		ginkgo.It("refuses two groups with the same name", func() {
			input.Spec.Groups = append(input.Spec.Groups, &KubernetesPrometheusRuleGroup{Name: "api-availability"})
			expectViolation(protovalidate.Validate(input), "unique name")
		})

		ginkgo.It("refuses a group without a name", func() {
			input.Spec.Groups[0].Name = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an interval that is not a Prometheus duration", func() {
			input.Spec.Groups[0].Interval = ptr("30 seconds")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a query offset that is not a Prometheus duration", func() {
			input.Spec.Groups[0].QueryOffset = ptr("1 minute")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a negative limit", func() {
			input.Spec.Groups[0].Limit = ptr(int32(-1))
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a partial-response strategy other than abort or warn", func() {
			input.Spec.Groups[0].PartialResponseStrategy = ptr("retry")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When a rule's shape is wrong", func() {
		ginkgo.It("refuses a rule that is neither recording nor alerting", func() {
			input.Spec.Groups[0].Rules[0].Record = nil
			expectViolation(protovalidate.Validate(input), "set exactly one")
		})

		ginkgo.It("refuses a rule that is both recording and alerting", func() {
			input.Spec.Groups[0].Rules[0].Alert = ptr("Both")
			expectViolation(protovalidate.Validate(input), "set exactly one")
		})

		ginkgo.It("refuses a rule with no expression", func() {
			input.Spec.Groups[0].Rules[1].Expr = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a recording rule name that is not a metric name", func() {
			input.Spec.Groups[0].Rules[0].Record = ptr("errors-per-request")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an empty alert name", func() {
			input.Spec.Groups[0].Rules[1].Alert = ptr("")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a for duration that is not a Prometheus duration", func() {
			input.Spec.Groups[0].Rules[1].ForDuration = ptr("five minutes")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an empty keep_firing_for, which upstream requires to be non-empty when set", func() {
			input.Spec.Groups[0].Rules[1].KeepFiringFor = ptr("")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When a recording rule carries alerting-rule fields", func() {
		ginkgo.It("refuses for", func() {
			input.Spec.Groups[0].Rules[0].ForDuration = ptr("5m")
			expectViolation(protovalidate.Validate(input), "recording rule")
		})

		ginkgo.It("refuses keep_firing_for", func() {
			input.Spec.Groups[0].Rules[0].KeepFiringFor = ptr("5m")
			expectViolation(protovalidate.Validate(input), "recording rule")
		})

		ginkgo.It("refuses annotations", func() {
			input.Spec.Groups[0].Rules[0].Annotations = map[string]string{"summary": "x"}
			expectViolation(protovalidate.Validate(input), "recording rule")
		})

		ginkgo.It("still accepts labels on a recording rule", func() {
			input.Spec.Groups[0].Rules[0].Labels = map[string]string{"component": "api"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})
})
