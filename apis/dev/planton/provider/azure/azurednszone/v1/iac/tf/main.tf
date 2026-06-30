# Azure DNS Zone Terraform Module
# Auto-release test: Multi-provider Terraform change (Azure component).

# Create the Azure DNS Zone
# This is a public DNS zone that will be authoritative for the specified domain
resource "azurerm_dns_zone" "dns_zone" {
  name                = var.spec.zone_name
  resource_group_name = var.spec.resource_group
  tags                = local.final_tags
}

# Create A records
# A records map hostnames to IPv4 addresses
resource "azurerm_dns_a_record" "a_records" {
  for_each = { for idx, rec in local.a_records : idx => rec }

  name                = each.value.name
  zone_name           = azurerm_dns_zone.dns_zone.name
  resource_group_name = var.spec.resource_group
  ttl                 = each.value.ttl_seconds
  records             = each.value.values
  tags                = local.final_tags
}

# Create AAAA records
# AAAA records map hostnames to IPv6 addresses
resource "azurerm_dns_aaaa_record" "aaaa_records" {
  for_each = { for idx, rec in local.aaaa_records : idx => rec }

  name                = each.value.name
  zone_name           = azurerm_dns_zone.dns_zone.name
  resource_group_name = var.spec.resource_group
  ttl                 = each.value.ttl_seconds
  records             = each.value.values
  tags                = local.final_tags
}

# Create CNAME records
# CNAME records create aliases from one domain name to another
resource "azurerm_dns_cname_record" "cname_records" {
  for_each = { for idx, rec in local.cname_records : idx => rec }

  name                = each.value.name
  zone_name           = azurerm_dns_zone.dns_zone.name
  resource_group_name = var.spec.resource_group
  ttl                 = each.value.ttl_seconds
  record              = each.value.values[0]
  tags                = local.final_tags
}

# Create MX records
# MX records specify mail servers for the domain
resource "azurerm_dns_mx_record" "mx_records" {
  for_each = { for idx, rec in local.mx_records : idx => rec }

  name                = each.value.name
  zone_name           = azurerm_dns_zone.dns_zone.name
  resource_group_name = var.spec.resource_group
  ttl                 = each.value.ttl_seconds
  tags                = local.final_tags

  dynamic "record" {
    for_each = each.value.values
    content {
      preference = 10
      exchange   = record.value
    }
  }
}

# Create TXT records
# TXT records store arbitrary text data, commonly used for SPF, DKIM, DMARC, and domain verification
resource "azurerm_dns_txt_record" "txt_records" {
  for_each = { for idx, rec in local.txt_records : idx => rec }

  name                = each.value.name
  zone_name           = azurerm_dns_zone.dns_zone.name
  resource_group_name = var.spec.resource_group
  ttl                 = each.value.ttl_seconds
  tags                = local.final_tags

  dynamic "record" {
    for_each = each.value.values
    content {
      value = record.value
    }
  }
}

# Create NS records
# NS records delegate a subdomain to different nameservers
resource "azurerm_dns_ns_record" "ns_records" {
  for_each = { for idx, rec in local.ns_records : idx => rec }

  name                = each.value.name
  zone_name           = azurerm_dns_zone.dns_zone.name
  resource_group_name = var.spec.resource_group
  ttl                 = each.value.ttl_seconds
  records             = each.value.values
  tags                = local.final_tags
}

# Create CAA records
# CAA records control which certificate authorities can issue certificates for the domain
resource "azurerm_dns_caa_record" "caa_records" {
  for_each = { for idx, rec in local.caa_records : idx => rec }

  name                = each.value.name
  zone_name           = azurerm_dns_zone.dns_zone.name
  resource_group_name = var.spec.resource_group
  ttl                 = each.value.ttl_seconds
  tags                = local.final_tags

  dynamic "record" {
    for_each = each.value.values
    content {
      flags = 0
      tag   = "issue"
      value = record.value
    }
  }
}

# Create SRV records
# SRV records specify the location of services (used for LDAP, SIP, XMPP, etc.)
resource "azurerm_dns_srv_record" "srv_records" {
  for_each = { for idx, rec in local.srv_records : idx => rec }

  name                = each.value.name
  zone_name           = azurerm_dns_zone.dns_zone.name
  resource_group_name = var.spec.resource_group
  ttl                 = each.value.ttl_seconds
  tags                = local.final_tags

  dynamic "record" {
    for_each = each.value.values
    content {
      priority = 10
      weight   = 10
      port     = 80
      target   = record.value
    }
  }
}

# Create PTR records
# PTR records are used for reverse DNS lookups (IP to hostname)
resource "azurerm_dns_ptr_record" "ptr_records" {
  for_each = { for idx, rec in local.ptr_records : idx => rec }

  name                = each.value.name
  zone_name           = azurerm_dns_zone.dns_zone.name
  resource_group_name = var.spec.resource_group
  ttl                 = each.value.ttl_seconds
  records             = each.value.values
  tags                = local.final_tags
}

