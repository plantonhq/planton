# Computed values for the AzurePlantonRunner module. Every resolution here
# has an exact twin in the Pulumi module's locals.go — keep them in
# lockstep.

locals {
  runner_name = var.metadata.name

  # The name the runner registers itself under when it joins the control
  # plane: "<env>-<metadata.name>" (metadata.name outside an environment)
  # — the SAME derivation the platform uses for records that reference
  # this runner (its minted token, its managed destroy); changing this
  # formula breaks arrival attribution and managed teardown.
  registration_name = try(var.metadata.env, "") != "" ? "${var.metadata.env}-${var.metadata.name}" : var.metadata.name

  # The Container App secret holding the runner token — the app's own
  # secret store, referenced by the env's secret_name. A fixed name: the
  # app carries exactly one secret, and the name is part of the
  # cross-engine contract (and an output).
  token_secret_name = "runner-token"

  # The runner's Consumption-plan sizing — the spec's documented defaults
  # applied when unset, so the tfvars path (which prunes unset fields)
  # lands on the same values the Pulumi module resolves.
  cpu    = try(var.spec.cpu, null) != null ? var.spec.cpu : 0.5
  memory = try(var.spec.memory, "") != "" ? var.spec.memory : "1Gi"

  # Resource-identity tags, matching the Pulumi module key-for-key.
  azure_tags = merge(
    {
      "resource"      = "true"
      "resource_name" = var.metadata.name
      "resource_kind" = "azureplantonrunner"
    },
    try(var.metadata.id, "") != "" ? { "resource_id" = var.metadata.id } : {},
    try(var.metadata.org, "") != "" ? { "organization" = var.metadata.org } : {},
    try(var.metadata.env, "") != "" ? { "environment" = var.metadata.env } : {}
  )
}
