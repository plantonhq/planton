locals {
  # The spec accepts the billing account as its bare ID or as the resource
  # name; the provider wants the bare ID and prefixes billingAccounts/
  # itself.
  billing_account = trimprefix(var.spec.billing_account, "billingAccounts/")

  # The console name defaults to metadata.name when the spec leaves
  # display_name empty -- the same naming basis every kind uses.
  display_name = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  # The filter's project and folder references arrive flattened: a project
  # NUMBER (the API's form is projects/{number}) and a folder resource name
  # (folders/{id}); an organizations/{id} literal passes through. The
  # module adds the projects/ prefix where a bare number was given so a
  # literal number and a resolved reference render the same.
  filter_projects = var.spec.budget_filter != null ? [
    for p in var.spec.budget_filter.projects : startswith(p, "projects/") ? p : "projects/${p}"
  ] : []

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
