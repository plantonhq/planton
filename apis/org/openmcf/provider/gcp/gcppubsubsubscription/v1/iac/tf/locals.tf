locals {
  project_id        = var.spec.project_id.value
  subscription_name = var.spec.subscription_name
  topic             = var.spec.topic.value

  ack_deadline_seconds       = var.spec.ack_deadline_seconds > 0 ? var.spec.ack_deadline_seconds : null
  message_retention_duration = var.spec.message_retention_duration != "" ? var.spec.message_retention_duration : null
  filter                     = var.spec.filter != "" ? var.spec.filter : null
}
