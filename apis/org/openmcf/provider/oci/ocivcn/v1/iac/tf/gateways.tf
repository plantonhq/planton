##############################
# Internet Gateway
##############################

resource "oci_core_internet_gateway" "this" {
  count = var.spec.is_internet_gateway_enabled ? 1 : 0

  compartment_id = var.spec.compartment_id.value
  vcn_id         = oci_core_vcn.this.id
  display_name   = "${local.display_name}-igw"
  enabled        = true
  freeform_tags  = local.freeform_tags
}

##############################
# NAT Gateway
##############################

resource "oci_core_nat_gateway" "this" {
  count = var.spec.is_nat_gateway_enabled ? 1 : 0

  compartment_id = var.spec.compartment_id.value
  vcn_id         = oci_core_vcn.this.id
  display_name   = "${local.display_name}-ngw"
  block_traffic  = false
  freeform_tags  = local.freeform_tags
}

##############################
# Service Gateway
##############################

data "oci_core_services" "all" {
  count = var.spec.is_service_gateway_enabled ? 1 : 0
}

resource "oci_core_service_gateway" "this" {
  count = var.spec.is_service_gateway_enabled ? 1 : 0

  compartment_id = var.spec.compartment_id.value
  vcn_id         = oci_core_vcn.this.id
  display_name   = "${local.display_name}-sgw"
  freeform_tags  = local.freeform_tags

  dynamic "services" {
    for_each = var.spec.is_service_gateway_enabled ? data.oci_core_services.all[0].services : []
    content {
      service_id = services.value.id
    }
  }
}
