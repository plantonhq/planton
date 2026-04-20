locals {
  all_domains  = concat([var.spec.primary_domain_name], var.spec.alternate_domain_names)
  is_managed   = var.spec.certificate_type == null || var.spec.certificate_type == 0
  has_dns_zone = var.spec.cloud_dns_zone_id != null
}

#############################################
# Certificate Manager Certificate (MANAGED)
#############################################

# DNS authorizations for each domain (Certificate Manager only)
resource "google_certificate_manager_dns_authorization" "dns_auth" {
  for_each = local.is_managed ? toset(local.all_domains) : toset([])
  
  name        = "${var.metadata.name}-${replace(each.value, "*", "wildcard")}-dns-auth"
  description = "DNS authorization for ${each.value}"
  domain      = each.value
  project     = var.spec.gcp_project_id
  labels      = local.gcp_labels
}

# DNS validation records — only created when a Cloud DNS zone is provided.
# When omitted, the dns-validation-records output contains the records for manual insertion.
resource "google_dns_record_set" "validation_records" {
  for_each = local.is_managed && local.has_dns_zone ? google_certificate_manager_dns_authorization.dns_auth : {}
  
  name         = each.value.dns_resource_record[0].name
  type         = each.value.dns_resource_record[0].type
  ttl          = 300
  managed_zone = var.spec.cloud_dns_zone_id.value
  project      = var.spec.gcp_project_id

  rrdatas = [each.value.dns_resource_record[0].data]
}

# Certificate Manager certificate (MANAGED type)
resource "google_certificate_manager_certificate" "cert" {
  count = local.is_managed ? 1 : 0
  
  name        = var.metadata.name
  description = "SSL certificate for ${var.spec.primary_domain_name}"
  project     = var.spec.gcp_project_id
  labels      = local.gcp_labels

  managed {
    domains = local.all_domains
    
    dns_authorizations = [
      for auth in google_certificate_manager_dns_authorization.dns_auth : auth.id
    ]
  }

  depends_on = [google_dns_record_set.validation_records]
}

#############################################
# Google-managed SSL Certificate (LOAD_BALANCER)
#############################################

# Google-managed SSL certificate for load balancers
resource "google_compute_managed_ssl_certificate" "lb_cert" {
  count = !local.is_managed ? 1 : 0
  
  name        = var.metadata.name
  description = "SSL certificate for ${var.spec.primary_domain_name}"
  project     = var.spec.gcp_project_id

  managed {
    domains = local.all_domains
  }
}

