locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The arm: spec.engine_type, or SEARCH when the spec leaves it empty --
  # identical to the Pulumi module. Exactly one of the three engine
  # resources below exists.
  engine_type       = var.spec.engine_type != "" ? var.spec.engine_type : "SEARCH"
  is_search         = local.engine_type == "SEARCH"
  is_chat           = local.engine_type == "CHAT"
  is_recommendation = local.engine_type == "RECOMMENDATION"

  # The Discovery Engine solution the arm hard-codes; every folded control
  # carries the same one.
  solution_type = local.is_chat ? "SOLUTION_TYPE_CHAT" : (local.is_recommendation ? "SOLUTION_TYPE_RECOMMENDATION" : "SOLUTION_TYPE_SEARCH")

  # The engine's GCP id and display name default to metadata.name (the
  # spec-level contract); the collection defaults to Google's
  # default_collection. Every companion is addressed under the collection.
  engine_id     = var.spec.engine_id != "" ? var.spec.engine_id : var.metadata.name
  display_name  = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name
  collection_id = var.spec.collection_id != "" ? var.spec.collection_id : "default_collection"

  # Empty optional strings become null so the provider omits them from the
  # API payload instead of sending empty values it would reject or diff on.
  industry_vertical = var.spec.industry_vertical != "" ? var.spec.industry_vertical : null
  app_type          = var.spec.app_type != "" ? var.spec.app_type : null
  kms_key_name      = var.spec.kms_key_name != "" ? var.spec.kms_key_name : null
  features          = length(var.spec.features) > 0 ? var.spec.features : null
  deletion_policy   = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
  company_name      = var.spec.common_config != null && var.spec.common_config.company_name != "" ? var.spec.common_config.company_name : null

  # The engine's id as Google returns it, whichever arm built it -- the
  # handle every companion addresses the engine by.
  created_engine_id = one(concat(
    google_discovery_engine_search_engine.this[*].engine_id,
    google_discovery_engine_chat_engine.this[*].engine_id,
    google_discovery_engine_recommendation_engine.this[*].engine_id,
  ))

  # The folded controls and assistants, keyed by their ids so adding or
  # removing one never renumbers the others.
  controls   = { for control in var.spec.controls : control.control_id => control }
  assistants = { for assistant in var.spec.assistants : assistant.assistant_id => assistant }
}
