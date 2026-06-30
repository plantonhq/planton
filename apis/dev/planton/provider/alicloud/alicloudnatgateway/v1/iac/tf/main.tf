data "alicloud_eip_addresses" "nat" {
  ids = [var.spec.eip_id]
}

resource "alicloud_nat_gateway" "main" {
  nat_gateway_name    = var.spec.nat_gateway_name
  vpc_id              = var.spec.vpc_id
  vswitch_id          = var.spec.vswitch_id
  description         = var.spec.description != "" ? var.spec.description : null
  nat_type            = var.spec.nat_type
  payment_type        = var.spec.payment_type
  internet_charge_type = var.spec.internet_charge_type
  specification       = var.spec.specification != "" ? var.spec.specification : null
  deletion_protection = var.spec.deletion_protection
  tags                = local.final_tags
}

resource "alicloud_eip_association" "nat" {
  allocation_id = var.spec.eip_id
  instance_id   = alicloud_nat_gateway.main.id
  instance_type = "Nat"
}
