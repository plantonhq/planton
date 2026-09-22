package gcpbillingbudgetv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpBillingBudgetSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func nameRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: v}},
	}
}

var _ = ginkgo.Describe("GcpBillingBudgetSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// A fixed monthly budget on the whole account with one alert.
	minimal := func() *GcpBillingBudget {
		return &GcpBillingBudget{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpBillingBudget",
			Metadata:   &shared.CloudResourceMetadata{Name: "monthly-guardrail"},
			Spec: &GcpBillingBudgetSpec{
				BillingAccount: "012345-6789AB-CDEF01",
				Amount:         &GcpBillingBudgetAmount{SpecifiedAmount: &GcpBillingBudgetSpecifiedAmount{CurrencyCode: "USD", Units: 1000}},
				ThresholdRules: []*GcpBillingBudgetThresholdRule{{ThresholdPercent: 0.9}},
			},
		}
	}

	ginkgo.It("should accept the minimal budget and the billingAccounts/ prefix", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
		msg := minimal()
		msg.Spec.BillingAccount = "billingAccounts/012345-6789AB-CDEF01"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every lever together", func() {
		msg := minimal()
		msg.Spec.DisplayName = "Platform monthly guardrail"
		msg.Spec.OwnershipScope = "BILLING_ACCOUNT"
		msg.Spec.Amount.SpecifiedAmount.Nanos = 500000000
		msg.Spec.BudgetFilter = &GcpBillingBudgetFilter{
			Projects:             []*foreignkeyv1.StringValueOrRef{nameRef("platform-prod"), litRef("projects/123456789012")},
			ResourceAncestors:    []*foreignkeyv1.StringValueOrRef{litRef("organizations/1234567890")},
			Services:             []string{"services/6F81-5844-456A"},
			Subaccounts:          []string{"billingAccounts/ABCDEF-012345-6789AB"},
			Labels:               map[string]string{"env": "prod"},
			CreditTypesTreatment: proto.String("INCLUDE_SPECIFIED_CREDITS"),
			CreditTypes:          []string{"PROMOTION", "FREE_TIER"},
			CalendarPeriod:       "QUARTER",
		}
		msg.Spec.ThresholdRules = append(msg.Spec.ThresholdRules, &GcpBillingBudgetThresholdRule{ThresholdPercent: 1.0, SpendBasis: proto.String("FORECASTED_SPEND")})
		msg.Spec.Notifications = &GcpBillingBudgetNotifications{
			PubsubTopic:                    nameRef("budget-alerts"),
			MonitoringNotificationChannels: []*foreignkeyv1.StringValueOrRef{nameRef("oncall-email")},
			DisableDefaultIamRecipients:    true,
			EnableProjectLevelRecipients:   true,
			SchemaVersion:                  proto.String("1.0"),
		}
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a last-period budget over a calendar period and a custom-period budget", func() {
		msg := minimal()
		msg.Spec.Amount = &GcpBillingBudgetAmount{LastPeriodAmount: true}
		msg.Spec.BudgetFilter = &GcpBillingBudgetFilter{CalendarPeriod: "MONTH"}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg = minimal()
		msg.Spec.BudgetFilter = &GcpBillingBudgetFilter{CustomPeriod: &GcpBillingBudgetCustomPeriod{
			StartDate: &GcpBillingBudgetDate{Year: 2026, Month: 10, Day: 1},
			EndDate:   &GcpBillingBudgetDate{Year: 2027, Month: 3, Day: 31},
		}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require billing_account and amount, and reject a malformed account", func() {
		msg := minimal()
		msg.Spec.BillingAccount = ""
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.BillingAccount = "my-billing-account"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.Amount = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an amount with both or neither arm", func() {
		msg := minimal()
		msg.Spec.Amount.LastPeriodAmount = true
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.Amount = &GcpBillingBudgetAmount{}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a last-period budget over a custom period", func() {
		msg := minimal()
		msg.Spec.Amount = &GcpBillingBudgetAmount{LastPeriodAmount: true}
		msg.Spec.BudgetFilter = &GcpBillingBudgetFilter{CustomPeriod: &GcpBillingBudgetCustomPeriod{StartDate: &GcpBillingBudgetDate{Year: 2026, Month: 1, Day: 1}}}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject both a calendar and a custom period", func() {
		msg := minimal()
		msg.Spec.BudgetFilter = &GcpBillingBudgetFilter{CalendarPeriod: "MONTH", CustomPeriod: &GcpBillingBudgetCustomPeriod{StartDate: &GcpBillingBudgetDate{Year: 2026, Month: 1, Day: 1}}}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject credit_types without INCLUDE_SPECIFIED_CREDITS", func() {
		msg := minimal()
		msg.Spec.BudgetFilter = &GcpBillingBudgetFilter{CreditTypes: []string{"PROMOTION"}}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid enums, a bad currency, a zero threshold, and an out-of-range date", func() {
		msg := minimal()
		msg.Spec.OwnershipScope = "EVERYONE"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.Amount.SpecifiedAmount.CurrencyCode = "usd"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.ThresholdRules[0].ThresholdPercent = 0
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.ThresholdRules[0].SpendBasis = proto.String("ACTUAL")
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.BudgetFilter = &GcpBillingBudgetFilter{CustomPeriod: &GcpBillingBudgetCustomPeriod{StartDate: &GcpBillingBudgetDate{Year: 2026, Month: 13, Day: 1}}}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.BudgetFilter = &GcpBillingBudgetFilter{Labels: map[string]string{"a": "1", "b": "2"}}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a notifications block with neither a topic nor a channel, and a wrong schema version", func() {
		msg := minimal()
		msg.Spec.Notifications = &GcpBillingBudgetNotifications{DisableDefaultIamRecipients: true}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.Notifications = &GcpBillingBudgetNotifications{PubsubTopic: nameRef("t"), SchemaVersion: proto.String("2.0")}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})
})
