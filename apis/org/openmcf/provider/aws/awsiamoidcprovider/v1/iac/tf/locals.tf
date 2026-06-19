locals {
  resource_name = coalesce(try(var.metadata.name, null), "aws-iam-oidc-provider")
  tags = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # url is a StringValueOrRef; flatten it to a primitive. At deploy time the runtime
  # resolves references into `value`, so that wins; value_from.name is a degenerate fallback.
  url = coalesce(try(var.spec.url.value, null), try(var.spec.url.value_from.name, null))

  client_id_list = try(var.spec.client_id_list, [])

  # Pass thumbprints only when provided. An empty list is normalized to null so the
  # provider treats thumbprint_list as Computed and lets AWS derive it from its trusted
  # CA store -- this is the single explicit Pulumi/Terraform parity point for this module.
  thumbprint_list = length(try(var.spec.thumbprint_list, [])) > 0 ? var.spec.thumbprint_list : null
}
