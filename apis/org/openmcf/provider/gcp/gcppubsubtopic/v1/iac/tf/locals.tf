locals {
  project_id                 = var.spec.project_id.value
  topic_name                 = var.spec.topic_name
  kms_key_name               = var.spec.kms_key_name != null ? var.spec.kms_key_name.value : null
  message_retention_duration = var.spec.message_retention_duration != "" ? var.spec.message_retention_duration : null
}
