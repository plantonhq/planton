package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpName is the budget's resource name, billingAccounts/{account}/budgets/{id}.
	OpName = "name"
	// OpBudgetId is the server-assigned id, the last segment of name.
	OpBudgetId = "budget_id"
	// OpBillingAccount is the owning account as billingAccounts/{id}.
	OpBillingAccount = "billing_account"
)
