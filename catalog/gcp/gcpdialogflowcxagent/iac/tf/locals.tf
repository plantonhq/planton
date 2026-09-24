locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # Google requires a display name; the spec defaults it to metadata.name --
  # identical to the Pulumi module.
  display_name = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  # Empty optional strings become null so the provider omits them from the
  # API payload instead of sending values it would reject or diff on.
  description               = var.spec.description != "" ? var.spec.description : null
  avatar_uri                = var.spec.avatar_uri != "" ? var.spec.avatar_uri : null
  security_settings         = var.spec.security_settings != "" ? var.spec.security_settings : null
  default_end_user_metadata = var.spec.default_end_user_metadata != "" ? var.spec.default_end_user_metadata : null
  synthesize_speech_configs = var.spec.synthesize_speech_configs != "" ? var.spec.synthesize_speech_configs : null
  supported_language_codes  = length(var.spec.supported_language_codes) > 0 ? var.spec.supported_language_codes : null
  deletion_policy           = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Google allows only the default playbook as a start playbook, and the
  # agent's id does not exist before creation, so the path is the wildcard
  # form Google accepts ("-" for project, location, and agent) -- the form
  # Google's own provider tests send.
  default_playbook = "projects/-/locations/-/agents/-/playbooks/00000000-0000-0000-0000-000000000000"
  start_playbook   = var.spec.start_with_default_playbook ? local.default_playbook : null

  # The start flow every agent is created with; a version or an environment
  # entry with an empty flow_id addresses it.
  start_flow_id = "00000000-0000-0000-0000-000000000000"

  # The folded webhooks and tools keyed by display name (Google requires it
  # unique within the agent), so adding or removing one never renumbers the
  # others.
  webhooks = { for webhook in var.spec.webhooks : webhook.display_name => webhook }
  tools    = { for tool in var.spec.tools : tool.display_name => tool }

  # Tool versions keyed "{tool display name}/{version display name}".
  tool_versions = merge([
    for tool in var.spec.tools : {
      for version in tool.versions : "${tool.display_name}/${version.display_name}" => {
        tool_key = tool.display_name
        version  = version
      }
    }
  ]...)

  # Flow versions keyed "{flow_id}/{display name}", the empty flow_id
  # resolved to the start flow.
  versions = {
    for version in var.spec.versions : "${version.flow_id != "" ? version.flow_id : local.start_flow_id}/${version.display_name}" => merge(version, {
      flow_id = version.flow_id != "" ? version.flow_id : local.start_flow_id
    })
  }

  environments        = { for environment in var.spec.environments : environment.display_name => environment }
  generative_settings = { for settings in var.spec.generative_settings : settings.language_code => settings }
}
