locals {
  project_id = var.spec.project_id.value
  network    = var.spec.network.value
  rule_name  = var.spec.rule_name
  direction  = var.spec.direction
  action     = var.spec.action
  rules      = var.spec.rules
  priority   = var.spec.priority
}
