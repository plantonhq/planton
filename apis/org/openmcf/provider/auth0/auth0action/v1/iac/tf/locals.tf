# Local values for Auth0Action module

locals {
  action_name       = var.metadata.name
  code              = var.spec.code
  runtime           = var.spec.runtime
  deploy            = var.spec.deploy
  supported_trigger = var.spec.supported_trigger
  dependencies      = var.spec.dependencies
  secrets           = var.spec.secrets
  trigger_binding   = var.spec.trigger_binding

  display_name = try(
    local.trigger_binding.display_name != null ? local.trigger_binding.display_name : local.action_name,
    local.action_name
  )
}
