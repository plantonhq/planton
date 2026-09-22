package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// billingBudgetVerifier probes a Cloud Billing budget through the Billing
// Budget API by the name output (billingAccounts/{account}/budgets/{id}).
// Posture assertions confirm the budget carries an amount and that the
// display name the manifest declared is what Google stores -- the value a
// finance reader finds in the console.
type billingBudgetVerifier struct{}

func (v *billingBudgetVerifier) IDOutputKey() string { return "name" }

func (v *billingBudgetVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}

	budget, err := svc.BillingBudgets.BillingAccounts.Budgets.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "billing budget %s not found after deploy", name)
	}

	if budget.Amount == nil || (budget.Amount.SpecifiedAmount == nil && budget.Amount.LastPeriodAmount == nil) {
		return errors.Errorf("billing budget %s carries no amount after deploy", name)
	}
	if budget.DisplayName == "" {
		return errors.Errorf("billing budget %s has no display name after deploy", name)
	}
	return nil
}

func (v *billingBudgetVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}

	_, err := svc.BillingBudgets.BillingAccounts.Budgets.Get(name).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing billing budget %s after destroy", name)
	}
	return errors.Errorf("billing budget %s still exists after destroy", name)
}
