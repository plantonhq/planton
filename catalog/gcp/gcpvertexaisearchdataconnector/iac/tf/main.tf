# Enable the Discovery Engine API first so a fresh project works on the
# first deploy. disable_on_destroy is false: tearing down one connector
# must never disable the API for everything else in the project.
resource "google_project_service" "discoveryengine_api" {
  project = local.project_id
  service = "discoveryengine.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The data connector: setting it up creates the collection and one data
# store per entity. The collection ids, the source, the location, the KMS
# key, the static-IP switch, and each entity's name are immutable; the
# schedule, the parameters, the modes, and the action side update in
# place. Optional strings, lists, and blocks are sent only when set so
# Google's defaults stay in charge (the Pulumi module's posture). The
# parameter maps carry Secret Manager resource names, never secret
# material, by Google's design.
resource "google_discovery_engine_data_connector" "this" {
  project                 = local.project_id
  location                = var.spec.location
  collection_id           = local.collection_id
  collection_display_name = local.collection_display_name
  data_source             = var.spec.data_source
  data_source_version     = var.spec.data_source_version
  params                  = local.params
  json_params             = local.json_params

  refresh_interval             = var.spec.refresh_interval
  incremental_refresh_interval = local.incremental_refresh_interval
  incremental_sync_disabled    = var.spec.incremental_sync_disabled
  auto_run_disabled            = var.spec.auto_run_disabled
  sync_mode                    = local.sync_mode
  connector_modes              = local.connector_modes
  static_ip_enabled            = var.spec.static_ip_enabled
  kms_key_name                 = local.kms_key_name

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  # Sent only when set so the provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy

  dynamic "entities" {
    for_each = var.spec.entities
    content {
      entity_name           = entities.value.entity_name
      params                = entities.value.params != "" ? entities.value.params : null
      key_property_mappings = length(entities.value.key_property_mappings) > 0 ? entities.value.key_property_mappings : null
    }
  }

  dynamic "destination_configs" {
    for_each = var.spec.destination_configs
    content {
      key    = destination_configs.value.key != "" ? destination_configs.value.key : null
      params = destination_configs.value.params != "" ? destination_configs.value.params : null

      dynamic "destinations" {
        for_each = destination_configs.value.destinations
        content {
          host = destinations.value.host != "" ? destinations.value.host : null
          port = destinations.value.port
        }
      }
    }
  }

  dynamic "action_config" {
    for_each = var.spec.action_config != null ? [var.spec.action_config] : []
    content {
      action_params         = length(action_config.value.action_params) > 0 ? action_config.value.action_params : null
      create_bap_connection = action_config.value.create_bap_connection
    }
  }

  dynamic "bap_config" {
    for_each = var.spec.bap_config != null ? [var.spec.bap_config] : []
    content {
      supported_connector_modes = length(bap_config.value.supported_connector_modes) > 0 ? bap_config.value.supported_connector_modes : null
      enabled_actions           = length(bap_config.value.enabled_actions) > 0 ? bap_config.value.enabled_actions : null
    }
  }

  depends_on = [google_project_service.discoveryengine_api]
}
