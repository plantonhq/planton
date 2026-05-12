# Auth0Action Main Resources

resource "auth0_action" "this" {
  name    = local.action_name
  code    = local.code
  runtime = local.runtime
  deploy  = local.deploy

  supported_triggers {
    id      = local.supported_trigger.id
    version = local.supported_trigger.version
  }

  dynamic "dependencies" {
    for_each = local.dependencies
    content {
      name    = dependencies.value.name
      version = dependencies.value.version
    }
  }

  dynamic "secrets" {
    for_each = local.secrets
    content {
      name  = secrets.value.name
      value = secrets.value.value
    }
  }
}

# Conditionally bind the action to its trigger
resource "auth0_trigger_action" "this" {
  count = local.trigger_binding != null ? 1 : 0

  trigger      = local.supported_trigger.id
  action_id    = auth0_action.this.id
  display_name = local.display_name

  depends_on = [auth0_action.this]
}
