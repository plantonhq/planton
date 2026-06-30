output "record_id" {
  description = "The Azure Resource Manager ID of the DNS record"
  value = coalesce(
    try(azurerm_dns_a_record.a_record[0].id, null),
    try(azurerm_dns_aaaa_record.aaaa_record[0].id, null),
    try(azurerm_dns_cname_record.cname_record[0].id, null),
    try(azurerm_dns_mx_record.mx_record[0].id, null),
    try(azurerm_dns_txt_record.txt_record[0].id, null),
    try(azurerm_dns_ns_record.ns_record[0].id, null),
    try(azurerm_dns_caa_record.caa_record[0].id, null),
    try(azurerm_dns_srv_record.srv_record[0].id, null),
    try(azurerm_dns_ptr_record.ptr_record[0].id, null),
    ""
  )
}

output "fqdn" {
  description = "The fully qualified domain name (FQDN) for this record"
  value = coalesce(
    try(azurerm_dns_a_record.a_record[0].fqdn, null),
    try(azurerm_dns_aaaa_record.aaaa_record[0].fqdn, null),
    try(azurerm_dns_cname_record.cname_record[0].fqdn, null),
    try(azurerm_dns_mx_record.mx_record[0].fqdn, null),
    try(azurerm_dns_txt_record.txt_record[0].fqdn, null),
    try(azurerm_dns_ns_record.ns_record[0].fqdn, null),
    try(azurerm_dns_caa_record.caa_record[0].fqdn, null),
    try(azurerm_dns_srv_record.srv_record[0].fqdn, null),
    try(azurerm_dns_ptr_record.ptr_record[0].fqdn, null),
    local.fqdn
  )
}
