package gcpdeploypolicyv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpDeployPolicySpec Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func reference(kind catalogkind.CatalogKind, name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name}}}
}

var _ = ginkgo.Describe("GcpDeployPolicySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	weekendFreeze := func() *GcpDeployPolicyRule {
		return &GcpDeployPolicyRule{RolloutRestriction: &GcpDeployPolicyRolloutRestriction{
			Id: "weekend-freeze",
			TimeWindows: &GcpDeployPolicyTimeWindows{
				TimeZone:      "America/New_York",
				WeeklyWindows: []*GcpDeployPolicyWeeklyWindow{{DaysOfWeek: []string{"SATURDAY", "SUNDAY"}}},
			},
		}}
	}

	minimal := func() *GcpDeployPolicy {
		return &GcpDeployPolicy{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpDeployPolicy",
			Metadata:   &shared.CatalogObjectMetadata{Name: "prod-freeze"},
			Spec: &GcpDeployPolicySpec{
				Location:  "us-central1",
				Rules:     []*GcpDeployPolicyRule{weekendFreeze()},
				Selectors: []*GcpDeployPolicySelector{{Target: &GcpDeployPolicyTargetSelector{Id: literal("*")}}},
			},
		}
	}

	full := func() *GcpDeployPolicy {
		msg := minimal()
		msg.Spec.ProjectId = reference(catalogkind.CatalogKind_GcpProject, "delivery")
		msg.Spec.DeployPolicyId = "prod-freeze"
		msg.Spec.Description = "No production rollouts on weekends or over the year-end freeze"
		msg.Spec.Labels = map[string]string{"team": "platform"}
		msg.Spec.Annotations = map[string]string{"owner": "release-eng"}
		msg.Spec.Suspended = true
		msg.Spec.DeletionPolicy = "PREVENT"
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpDeployPolicyRule{RolloutRestriction: &GcpDeployPolicyRolloutRestriction{
			Id:       "year-end",
			Actions:  []string{"CREATE", "ROLLBACK"},
			Invokers: []string{"USER", "DEPLOY_AUTOMATION"},
			TimeWindows: &GcpDeployPolicyTimeWindows{
				TimeZone: "Europe/Berlin",
				OneTimeWindows: []*GcpDeployPolicyOneTimeWindow{{
					StartDate: &GcpDeployPolicyDate{Year: 2026, Month: 12, Day: 20},
					StartTime: &GcpDeployPolicyTimeOfDay{Hours: 17},
					EndDate:   &GcpDeployPolicyDate{Year: 2027, Month: 1, Day: 3},
					EndTime:   &GcpDeployPolicyTimeOfDay{Hours: 9, Minutes: 30},
				}},
				WeeklyWindows: []*GcpDeployPolicyWeeklyWindow{{
					DaysOfWeek: []string{"FRIDAY"},
					StartTime:  &GcpDeployPolicyTimeOfDay{Hours: 15},
					EndTime:    &GcpDeployPolicyTimeOfDay{Hours: 24},
				}},
			},
		}})
		msg.Spec.Selectors = []*GcpDeployPolicySelector{
			{
				DeliveryPipeline: &GcpDeployPolicyDeliveryPipelineSelector{Id: reference(catalogkind.CatalogKind_GcpDeliveryPipeline, "web")},
				Target:           &GcpDeployPolicyTargetSelector{Id: reference(catalogkind.CatalogKind_GcpDeployTarget, "web-prod")},
			},
			{Target: &GcpDeployPolicyTargetSelector{Labels: map[string]string{"env": "prod"}}},
			{DeliveryPipeline: &GcpDeployPolicyDeliveryPipelineSelector{Id: literal("*"), Labels: map[string]string{"tier": "critical"}}},
		}
		return msg
	}

	ginkgo.It("should accept a minimal policy and a fully declared one", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
		gomega.Expect(validator.Validate(full())).To(gomega.Succeed())
	})

	ginkgo.It("should require a location, a rule, and a selector", func() {
		msg := minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Rules = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Selectors = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a well-formed, unique restriction ID", func() {
		msg := minimal()
		msg.Spec.Rules[0].RolloutRestriction.Id = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		for _, id := range []string{"Weekend", "1freeze", "freeze-", "weekend_freeze"} {
			msg = minimal()
			msg.Spec.Rules[0].RolloutRestriction.Id = id
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), id)
		}
		msg = full()
		msg.Spec.Rules[1].RolloutRestriction.Id = "weekend-freeze"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require time windows with a time zone", func() {
		msg := minimal()
		msg.Spec.Rules[0].RolloutRestriction.TimeWindows = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Rules[0].RolloutRestriction.TimeWindows.TimeZone = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should refuse unknown actions, invokers, and days, and duplicates", func() {
		msg := minimal()
		msg.Spec.Rules[0].RolloutRestriction.Actions = []string{"DEPLOY"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Rules[0].RolloutRestriction.Actions = []string{"CREATE", "CREATE"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Rules[0].RolloutRestriction.Invokers = []string{"SERVICE_ACCOUNT"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Rules[0].RolloutRestriction.TimeWindows.WeeklyWindows[0].DaysOfWeek = []string{"SAT"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a weekly window's start and end together", func() {
		msg := minimal()
		msg.Spec.Rules[0].RolloutRestriction.TimeWindows.WeeklyWindows[0].EndTime = &GcpDeployPolicyTimeOfDay{Hours: 24}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Rules[0].RolloutRestriction.TimeWindows.WeeklyWindows[0].StartTime = &GcpDeployPolicyTimeOfDay{Hours: 15}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require every bound of a one-time window and in-range clock values", func() {
		msg := full()
		msg.Spec.Rules[1].RolloutRestriction.TimeWindows.OneTimeWindows[0].EndDate = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Rules[1].RolloutRestriction.TimeWindows.OneTimeWindows[0].StartTime = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Rules[1].RolloutRestriction.TimeWindows.OneTimeWindows[0].EndTime.Minutes = 60
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Rules[1].RolloutRestriction.TimeWindows.OneTimeWindows[0].StartDate.Month = 13
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Rules[1].RolloutRestriction.TimeWindows.WeeklyWindows[0].EndTime.Hours = 25
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should take an ID or * in selectors, never a full name", func() {
		msg := minimal()
		msg.Spec.Selectors[0].Target.Id = literal("projects/p/locations/us-central1/targets/web-prod")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Selectors = []*GcpDeployPolicySelector{{DeliveryPipeline: &GcpDeployPolicyDeliveryPipelineSelector{Id: literal("Web")}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Selectors = []*GcpDeployPolicySelector{{DeliveryPipeline: &GcpDeployPolicyDeliveryPipelineSelector{Id: literal("web")}}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed policy ID, a long description, and an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeployPolicyId = "Prod_Freeze"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Description = strings.Repeat("a", 256)
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
