# Azure DNS Record Terraform Module
# Creates individual DNS records in an existing Azure DNS Zone.

# Create A record if record_type is A
resource "azurerm_dns_a_record" "a_record" {
  count = local.is_a_record ? 1 : 0

  name                = var.spec.name
  zone_name           = var.spec.zone_name
  resource_group_name = var.spec.resource_group
  ttl                 = var.spec.ttl_seconds
  records             = var.spec.values
  tags                = local.final_tags
}

# Create AAAA record if record_type is AAAA
resource "azurerm_dns_aaaa_record" "aaaa_record" {
  count = local.is_aaaa_record ? 1 : 0

  name                = var.spec.name
  zone_name           = var.spec.zone_name
  resource_group_name = var.spec.resource_group
  ttl                 = var.spec.ttl_seconds
  records             = var.spec.values
  tags                = local.final_tags
}

# Create CNAME record if record_type is CNAME
resource "azurerm_dns_cname_record" "cname_record" {
  count = local.is_cname_record ? 1 : 0

  name                = var.spec.name
  zone_name           = var.spec.zone_name
  resource_group_name = var.spec.resource_group
  ttl                 = var.spec.ttl_seconds
  record              = var.spec.values[0]
  tags                = local.final_tags
}

# Create MX record if record_type is MX
resource "azurerm_dns_mx_record" "mx_record" {
  count = local.is_mx_record ? 1 : 0

  name                = var.spec.name
  zone_name           = var.spec.zone_name
  resource_group_name = var.spec.resource_group
  ttl                 = var.spec.ttl_seconds
  tags                = local.final_tags

  dynamic "record" {
    for_each = var.spec.values
    content {
      preference = var.spec.mx_priority
      exchange   = record.value
    }
  }
}

# Create TXT record if record_type is TXT
resource "azurerm_dns_txt_record" "txt_record" {
  count = local.is_txt_record ? 1 : 0

  name                = var.spec.name
  zone_name           = var.spec.zone_name
  resource_group_name = var.spec.resource_group
  ttl                 = var.spec.ttl_seconds
  tags                = local.final_tags

  dynamic "record" {
    for_each = var.spec.values
    content {
      value = record.value
    }
  }
}

# Create NS record if record_type is NS
resource "azurerm_dns_ns_record" "ns_record" {
  count = local.is_ns_record ? 1 : 0

  name                = var.spec.name
  zone_name           = var.spec.zone_name
  resource_group_name = var.spec.resource_group
  ttl                 = var.spec.ttl_seconds
  records             = var.spec.values
  tags                = local.final_tags
}

# Create CAA record if record_type is CAA
resource "azurerm_dns_caa_record" "caa_record" {
  count = local.is_caa_record ? 1 : 0

  name                = var.spec.name
  zone_name           = var.spec.zone_name
  resource_group_name = var.spec.resource_group
  ttl                 = var.spec.ttl_seconds
  tags                = local.final_tags

  dynamic "record" {
    for_each = var.spec.values
    content {
      flags = 0
      tag   = "issue"
      value = record.value
    }
  }
}

# Create SRV record if record_type is SRV
resource "azurerm_dns_srv_record" "srv_record" {
  count = local.is_srv_record ? 1 : 0

  name                = var.spec.name
  zone_name           = var.spec.zone_name
  resource_group_name = var.spec.resource_group
  ttl                 = var.spec.ttl_seconds
  tags                = local.final_tags

  dynamic "record" {
    for_each = var.spec.values
    content {
      priority = 10
      weight   = 10
      port     = 80
      target   = record.value
    }
  }
}

# Create PTR record if record_type is PTR
resource "azurerm_dns_ptr_record" "ptr_record" {
  count = local.is_ptr_record ? 1 : 0

  name                = var.spec.name
  zone_name           = var.spec.zone_name
  resource_group_name = var.spec.resource_group
  ttl                 = var.spec.ttl_seconds
  records             = var.spec.values
  tags                = local.final_tags
}
