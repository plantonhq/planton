# Computed values for the GcpPlantonRunner module. Every resolution here
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

  # The Secret Manager secret holding the runner token.
  token_secret_id = "${var.metadata.name}-token"

  # null resolves to the provider's default project — the ambient-project
  # contract every GCP kind honors.
  project_id = try(var.spec.project_id, "") != "" ? var.spec.project_id : null

  # Resource-identity labels, matching the Pulumi module key-for-key
  # (the gcplabelkeys constants; lowercased kind — GCP label values must
  # be lowercase).
  gcp_labels = merge(
    {
      "planton-ai_resource" = "true"
      "planton-ai_name"     = var.metadata.name
      "planton-ai_kind"     = lower("GcpPlantonRunner")
    },
    try(var.metadata.org, "") != "" ? { "planton-ai_organization" = var.metadata.org } : {},
    try(var.metadata.env, "") != "" ? { "planton-ai_environment" = var.metadata.env } : {},
    try(var.metadata.id, "") != "" ? { "planton-ai_id" = var.metadata.id } : {}
  )

  # The runtime identity the runner holds while executing work: the
  # referenced service_account when supplied (the module never mutates a
  # resource it merely references), else the dedicated permissionless
  # account created below. one() over the splat stays null-safe when
  # count is 0.
  create_service_account = try(var.spec.service_account, "") == ""
  service_account_email  = local.create_service_account ? one(google_service_account.runtime[*].email) : var.spec.service_account
}
