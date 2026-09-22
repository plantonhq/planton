locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project (ambient credentials decide).
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # policy_name falls back to metadata.name — explicit conditional, so both
  # engines derive the identical cloud-side name.
  policy_name = var.spec.policy_name != "" ? var.spec.policy_name : var.metadata.name

  # The scope selector. An empty region builds the global security policy;
  # a region name builds the regional one. Exactly one of the two resources
  # in main.tf exists (count guards), and outputs.tf picks whichever was
  # created.
  is_regional = var.spec.region != null && var.spec.region != ""

  # The region's enrollment in advanced network DDoS protection is a
  # companion resource created only when the spec declares it (and the spec
  # CEL admits it only on a regional CLOUD_ARMOR_NETWORK policy with
  # ddos_protection_config). Its name defaults to the policy's.
  create_edge_service = local.is_regional && var.spec.network_edge_security_service != null
  edge_service_name = (
    var.spec.network_edge_security_service != null && var.spec.network_edge_security_service.name != ""
    ? var.spec.network_edge_security_service.name
    : local.policy_name
  )

  # What destroy does to the WAF shield: DELETE (default), PREVENT (refuse),
  # or ABANDON (drop from state, keep enforcing). Fanned out to the edge
  # service so the enrollment shares the policy's fate.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The same planton-ai_* label set the Pulumi module applies, so a resource
  # is attributable to its Planton object regardless of the engine that
  # created it. Conditional labels appear under the same conditions on both
  # sides.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = local.policy_name
    "planton-ai_kind"     = "gcpcloudarmorpolicy"
  }

  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "planton-ai_organization" = var.metadata.org } : {}

  env_label = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "planton-ai_environment" = var.metadata.env } : {}

  id_label = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "planton-ai_id" = var.metadata.id } : {}

  # User labels first: the platform labels win on key conflicts.
  final_labels = merge(var.spec.labels, local.base_labels, local.org_label, local.env_label, local.id_label)
}
