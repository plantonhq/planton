package module

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/billing"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// budget builds one Cloud Billing budget: an amount for a period, the spend
// it measures, the thresholds that alert, and where the alerts go.
// Everything but the billing account changes in place.
func budget(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpBillingBudget.Spec

	// Exactly one arm (spec CEL): a fixed amount, or last period's spend.
	amount := &billing.BudgetAmountArgs{}
	if sa := spec.Amount.GetSpecifiedAmount(); sa != nil {
		saArgs := &billing.BudgetAmountSpecifiedAmountArgs{
			Units: pulumi.StringPtr(strconv.FormatInt(sa.Units, 10)),
		}
		// Optional+Computed: unset takes the billing account's currency.
		if sa.CurrencyCode != "" {
			saArgs.CurrencyCode = pulumi.StringPtr(sa.CurrencyCode)
		}
		if sa.Nanos != 0 {
			saArgs.Nanos = pulumi.IntPtr(int(sa.Nanos))
		}
		amount.SpecifiedAmount = saArgs
	}
	if spec.Amount.LastPeriodAmount {
		amount.LastPeriodAmount = pulumi.BoolPtr(true)
	}

	args := &billing.BudgetArgs{
		BillingAccount: pulumi.String(locals.BillingAccount),
		DisplayName:    pulumi.StringPtr(locals.DisplayName),
		Amount:         amount,
	}

	// Who may read the budget's data; unset lets Google apply its default.
	if spec.OwnershipScope != "" {
		args.OwnershipScope = pulumi.StringPtr(spec.OwnershipScope)
	}

	// Which spend counts. The defaulted enums are sent explicitly so the
	// manifest states them (the loader applies the proto defaults before
	// either engine runs).
	if f := spec.BudgetFilter; f != nil {
		fArgs := &billing.BudgetBudgetFilterArgs{
			CreditTypesTreatment: pulumi.StringPtr(orDefault(f.GetCreditTypesTreatment(), "INCLUDE_ALL_CREDITS")),
		}
		if len(f.Projects) > 0 {
			// The API's form is projects/{number}; a resolved reference or a
			// bare literal number gets the prefix, a prefixed literal passes.
			projects := make(pulumi.StringArray, 0, len(f.Projects))
			for _, p := range f.Projects {
				v := p.GetValue()
				if !strings.HasPrefix(v, "projects/") {
					v = "projects/" + v
				}
				projects = append(projects, pulumi.String(v))
			}
			fArgs.Projects = projects
		}
		if len(f.ResourceAncestors) > 0 {
			ancestors := make(pulumi.StringArray, 0, len(f.ResourceAncestors))
			for _, a := range f.ResourceAncestors {
				ancestors = append(ancestors, pulumi.String(a.GetValue()))
			}
			fArgs.ResourceAncestors = ancestors
		}
		if len(f.Services) > 0 {
			fArgs.Services = pulumi.ToStringArray(f.Services)
		}
		if len(f.Subaccounts) > 0 {
			fArgs.Subaccounts = pulumi.ToStringArray(f.Subaccounts)
		}
		if len(f.Labels) > 0 {
			fArgs.Labels = pulumi.ToStringMap(f.Labels)
		}
		if len(f.CreditTypes) > 0 {
			fArgs.CreditTypes = pulumi.ToStringArray(f.CreditTypes)
		}
		if f.CalendarPeriod != "" {
			fArgs.CalendarPeriod = pulumi.StringPtr(f.CalendarPeriod)
		}
		if cp := f.CustomPeriod; cp != nil {
			cpArgs := &billing.BudgetBudgetFilterCustomPeriodArgs{
				StartDate: &billing.BudgetBudgetFilterCustomPeriodStartDateArgs{
					Year:  pulumi.Int(int(cp.StartDate.Year)),
					Month: pulumi.Int(int(cp.StartDate.Month)),
					Day:   pulumi.Int(int(cp.StartDate.Day)),
				},
			}
			if cp.EndDate != nil {
				cpArgs.EndDate = &billing.BudgetBudgetFilterCustomPeriodEndDateArgs{
					Year:  pulumi.Int(int(cp.EndDate.Year)),
					Month: pulumi.Int(int(cp.EndDate.Month)),
					Day:   pulumi.Int(int(cp.EndDate.Day)),
				}
			}
			fArgs.CustomPeriod = cpArgs
		}
		args.BudgetFilter = fArgs
	}

	// The thresholds that alert; spend_basis defaults to CURRENT_SPEND and
	// is sent explicitly.
	if len(spec.ThresholdRules) > 0 {
		var rules billing.BudgetThresholdRuleArray
		for _, r := range spec.ThresholdRules {
			rules = append(rules, &billing.BudgetThresholdRuleArgs{
				ThresholdPercent: pulumi.Float64(r.ThresholdPercent),
				SpendBasis:       pulumi.StringPtr(orDefault(r.GetSpendBasis(), "CURRENT_SPEND")),
			})
		}
		args.ThresholdRules = rules
	}

	// Where alerts go beyond the default administrator emails (the
	// provider's all_updates_rule block).
	if n := spec.Notifications; n != nil {
		nArgs := &billing.BudgetAllUpdatesRuleArgs{
			DisableDefaultIamRecipients:  pulumi.BoolPtr(n.DisableDefaultIamRecipients),
			EnableProjectLevelRecipients: pulumi.BoolPtr(n.EnableProjectLevelRecipients),
			SchemaVersion:                pulumi.StringPtr(orDefault(n.GetSchemaVersion(), "1.0")),
		}
		if v := n.PubsubTopic.GetValue(); v != "" {
			nArgs.PubsubTopic = pulumi.StringPtr(v)
		}
		if len(n.MonitoringNotificationChannels) > 0 {
			channels := make(pulumi.StringArray, 0, len(n.MonitoringNotificationChannels))
			for _, c := range n.MonitoringNotificationChannels {
				channels = append(channels, pulumi.String(c.GetValue()))
			}
			nArgs.MonitoringNotificationChannels = channels
		}
		args.AllUpdatesRule = nArgs
	}

	// What destroy does to the guardrail: DELETE (default), PREVENT
	// (refuse), or ABANDON (drop from state, keep alerting).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	created, err := billing.NewBudget(ctx, "budget", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create budget")
	}

	ctx.Export(OpName, created.Name)
	// The server-assigned id is the last segment of the resource name.
	ctx.Export(OpBudgetId, created.Name.ApplyT(func(name string) string {
		parts := strings.Split(name, "/")
		return parts[len(parts)-1]
	}).(pulumi.StringOutput))
	ctx.Export(OpBillingAccount, pulumi.String("billingAccounts/"+locals.BillingAccount))
	return nil
}

// orDefault returns v, or def when v is empty -- the module-side twin of
// the proto default the manifest loader applies.
func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
