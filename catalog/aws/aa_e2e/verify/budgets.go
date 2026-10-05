// AWS Budgets verifiers. Budgets is an account-global service: every
// call is scoped by the account ID (resolved from the credentials via
// STS), and the region parameter is ignored. The budget's identity at
// AWS is (account_id, budget_name); each folded action's identity is
// its AWS-generated action ID, exported in the action_ids map keyed by
// the action's spec name. The absent signal is the service's typed
// NotFoundException, never a string match.
package verify

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/budgets"
	budgetstypes "github.com/aws/aws-sdk-go-v2/service/budgets/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	pkgerrors "github.com/pkg/errors"
)

// budgetVerifier verifies an AwsBudget (and its folded actions through
// the outputs path), keyed on the budget name.
type budgetVerifier struct{}

func (*budgetVerifier) IDOutputKey() string { return "budget_name" }

func (*budgetVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, _ string) error {
	exists, err := budgetExists(ctx, cfg, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsbudget verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awsbudget %q not found after deploy", id)
	}
	return nil
}

func (*budgetVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, _ string) error {
	exists, err := budgetExists(ctx, cfg, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsbudget verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awsbudget %q still exists after destroy", id)
	}
	return nil
}

// VerifyExistsFromOutputs additionally walks the action_ids map: each
// folded action must answer DescribeBudgetAction.
func (v *budgetVerifier) VerifyExistsFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	budgetName, _ := outputs["budget_name"].(string)
	if err := v.VerifyExists(ctx, cfg, budgetName, region); err != nil {
		return err
	}
	accountID, err := callerAccountID(ctx, cfg)
	if err != nil {
		return pkgerrors.Wrap(err, "awsbudget verify-exists: resolve account id")
	}
	client := budgets.NewFromConfig(cfg)
	for actionName, actionID := range stringMapOutput(outputs["action_ids"]) {
		_, err := client.DescribeBudgetAction(ctx, &budgets.DescribeBudgetActionInput{
			AccountId:  aws.String(accountID),
			ActionId:   aws.String(actionID),
			BudgetName: aws.String(budgetName),
		})
		if err != nil {
			return pkgerrors.Wrapf(err, "awsbudget action %q (%s) not found after deploy", actionName, actionID)
		}
	}
	return nil
}

// VerifyAbsentFromOutputs asserts the budget and every folded action
// are gone.
func (v *budgetVerifier) VerifyAbsentFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	budgetName, _ := outputs["budget_name"].(string)
	if err := v.VerifyAbsent(ctx, cfg, budgetName, region); err != nil {
		return err
	}
	accountID, err := callerAccountID(ctx, cfg)
	if err != nil {
		return pkgerrors.Wrap(err, "awsbudget verify-absent: resolve account id")
	}
	client := budgets.NewFromConfig(cfg)
	for actionName, actionID := range stringMapOutput(outputs["action_ids"]) {
		_, err := client.DescribeBudgetAction(ctx, &budgets.DescribeBudgetActionInput{
			AccountId:  aws.String(accountID),
			ActionId:   aws.String(actionID),
			BudgetName: aws.String(budgetName),
		})
		if err == nil {
			return pkgerrors.Errorf("awsbudget action %q (%s) still exists after destroy", actionName, actionID)
		}
		var notFound *budgetstypes.NotFoundException
		if !errors.As(err, &notFound) {
			return pkgerrors.Wrapf(err, "awsbudget action %q verify-absent failed", actionName)
		}
	}
	return nil
}

func budgetExists(ctx context.Context, cfg aws.Config, budgetName string) (bool, error) {
	accountID, err := callerAccountID(ctx, cfg)
	if err != nil {
		return false, err
	}
	_, err = budgets.NewFromConfig(cfg).DescribeBudget(ctx, &budgets.DescribeBudgetInput{
		AccountId:  aws.String(accountID),
		BudgetName: aws.String(budgetName),
	})
	if err != nil {
		var notFound *budgetstypes.NotFoundException
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// callerAccountID resolves the deploying account from the credentials
// -- the scope every Budgets call requires.
func callerAccountID(ctx context.Context, cfg aws.Config) (string, error) {
	identity, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}
	return aws.ToString(identity.Account), nil
}

// stringMapOutput coerces a map-typed output (decoded JSON) to
// map[string]string, tolerating the empty/missing map.
func stringMapOutput(raw interface{}) map[string]string {
	out := map[string]string{}
	m, ok := raw.(map[string]interface{})
	if !ok {
		return out
	}
	for k, v := range m {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out
}
