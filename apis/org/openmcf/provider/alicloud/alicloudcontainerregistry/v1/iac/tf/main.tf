resource "alicloud_cr_ee_instance" "main" {
  instance_name     = var.spec.instance_name
  instance_type     = var.spec.instance_type
  payment_type      = var.spec.payment_type
  period            = var.spec.period > 0 ? var.spec.period : null
  password          = var.spec.password != "" ? var.spec.password : null
  resource_group_id = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
}
