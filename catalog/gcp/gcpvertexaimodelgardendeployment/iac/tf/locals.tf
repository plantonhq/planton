locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # Empty optional strings become null so the provider omits them from the
  # API payload instead of sending empty values it would reject or diff on.
  publisher_model_name  = var.spec.publisher_model_name != "" ? var.spec.publisher_model_name : null
  hugging_face_model_id = var.spec.hugging_face_model_id != "" ? var.spec.hugging_face_model_id : null
  deletion_policy       = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The three optional top-level blocks, each emitted only when the
  # manifest declares it (a missing deploy_config lets Model Garden pick the
  # model's recommended shape).
  model_config    = var.spec.model_config
  deploy_config   = var.spec.deploy_config
  endpoint_config = var.spec.endpoint_config
}
