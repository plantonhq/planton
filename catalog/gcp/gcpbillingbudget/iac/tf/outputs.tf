output "name" {
  description = "The budget's resource name (billingAccounts/{account}/budgets/{id})"
  value       = google_billing_budget.this.name
}

# The server-assigned id is the last segment of the resource name.
output "budget_id" {
  description = "The server-assigned budget id"
  value       = element(split("/", google_billing_budget.this.name), length(split("/", google_billing_budget.this.name)) - 1)
}

output "billing_account" {
  description = "The billing account the budget belongs to (billingAccounts/{id})"
  value       = "billingAccounts/${local.billing_account}"
}
