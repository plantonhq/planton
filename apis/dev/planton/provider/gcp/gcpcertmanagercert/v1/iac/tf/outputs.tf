output "certificate-id" {
  description = "The ID of the created certificate"
  value = local.is_managed ? (
    length(google_certificate_manager_certificate.cert) > 0 ? google_certificate_manager_certificate.cert[0].id : ""
  ) : (
    length(google_compute_managed_ssl_certificate.lb_cert) > 0 ? google_compute_managed_ssl_certificate.lb_cert[0].id : ""
  )
}

output "certificate-name" {
  description = "The name of the created certificate"
  value = local.is_managed ? (
    length(google_certificate_manager_certificate.cert) > 0 ? google_certificate_manager_certificate.cert[0].name : ""
  ) : (
    length(google_compute_managed_ssl_certificate.lb_cert) > 0 ? google_compute_managed_ssl_certificate.lb_cert[0].name : ""
  )
}

output "certificate-domain-name" {
  description = "The primary domain name of the certificate"
  value       = var.spec.primary_domain_name
}

output "certificate-status" {
  description = "The status of the certificate"
  value       = "PROVISIONING"
}

output "dns-validation-records" {
  description = "DNS validation records for manual insertion when cloud_dns_zone_id is omitted"
  value = [for k, auth in google_certificate_manager_dns_authorization.dns_auth : {
    record_name = auth.dns_resource_record[0].name
    record_type = auth.dns_resource_record[0].type
    record_data = auth.dns_resource_record[0].data
    domain      = auth.domain
  }]
}

