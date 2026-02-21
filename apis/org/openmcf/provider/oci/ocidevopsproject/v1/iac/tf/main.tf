resource "oci_devops_project" "this" {
  compartment_id = var.spec.compartment_id.value
  name           = var.metadata.name
  freeform_tags  = local.freeform_tags

  notification_config {
    topic_id = var.spec.notification_topic_id.value
  }

  description = var.spec.description != "" ? var.spec.description : null
}
