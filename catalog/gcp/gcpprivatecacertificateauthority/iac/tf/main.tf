# The certificate authority. A subordinate named by reference is signed and
# activated on create; destroy disables it and, unless skip_grace_period,
# leaves it in Google's 30-day soft delete. config, key_spec, type, lifetime,
# and gcs_bucket are immutable.
resource "google_privateca_certificate_authority" "this" {
  project                  = local.project_id
  location                 = var.spec.location
  pool                     = local.pool
  certificate_authority_id = local.certificate_authority_id
  type                     = local.type
  lifetime                 = local.lifetime
  pem_ca_certificate       = local.pem_ca_certificate
  gcs_bucket               = local.gcs_bucket
  desired_state            = local.desired_state
  labels                   = local.final_labels

  deletion_protection                    = local.deletion_protection
  skip_grace_period                      = var.spec.skip_grace_period
  ignore_active_certificates_on_deletion = var.spec.ignore_active_certificates_on_deletion

  config {
    subject_config {
      subject {
        common_name         = var.spec.config.subject_config.subject.common_name
        country_code        = var.spec.config.subject_config.subject.country_code != "" ? var.spec.config.subject_config.subject.country_code : null
        organization        = var.spec.config.subject_config.subject.organization != "" ? var.spec.config.subject_config.subject.organization : null
        organizational_unit = var.spec.config.subject_config.subject.organizational_unit != "" ? var.spec.config.subject_config.subject.organizational_unit : null
        locality            = var.spec.config.subject_config.subject.locality != "" ? var.spec.config.subject_config.subject.locality : null
        province            = var.spec.config.subject_config.subject.province != "" ? var.spec.config.subject_config.subject.province : null
        street_address      = var.spec.config.subject_config.subject.street_address != "" ? var.spec.config.subject_config.subject.street_address : null
        postal_code         = var.spec.config.subject_config.subject.postal_code != "" ? var.spec.config.subject_config.subject.postal_code : null
      }
      dynamic "subject_alt_name" {
        for_each = var.spec.config.subject_config.subject_alt_name != null ? [var.spec.config.subject_config.subject_alt_name] : []
        content {
          dns_names       = length(subject_alt_name.value.dns_names) > 0 ? subject_alt_name.value.dns_names : null
          uris            = length(subject_alt_name.value.uris) > 0 ? subject_alt_name.value.uris : null
          email_addresses = length(subject_alt_name.value.email_addresses) > 0 ? subject_alt_name.value.email_addresses : null
          ip_addresses    = length(subject_alt_name.value.ip_addresses) > 0 ? subject_alt_name.value.ip_addresses : null
        }
      }
    }

    dynamic "subject_key_id" {
      for_each = var.spec.config.subject_key_id != "" ? [var.spec.config.subject_key_id] : []
      content {
        key_id = subject_key_id.value
      }
    }

    x509_config {
      aia_ocsp_servers = length(var.spec.config.x509_config.aia_ocsp_servers) > 0 ? var.spec.config.x509_config.aia_ocsp_servers : null

      dynamic "ca_options" {
        for_each = var.spec.config.x509_config.ca_options != null ? [var.spec.config.x509_config.ca_options] : []
        content {
          # is_ca is required on an authority (the spec's rule); false is sent
          # with non_ca, the provider's way to state CA:FALSE.
          is_ca  = ca_options.value.is_ca
          non_ca = ca_options.value.is_ca == false ? true : null
          # A path length of 0 goes through zero_max_issuer_path_length: the
          # provider reads a plain 0 as unset.
          max_issuer_path_length      = coalesce(ca_options.value.max_issuer_path_length, 0) > 0 ? ca_options.value.max_issuer_path_length : null
          zero_max_issuer_path_length = ca_options.value.max_issuer_path_length == 0 ? true : null
        }
      }

      # The provider requires key_usage with both groups whenever the block
      # is sent; an omitted group is sent with every bit false.
      key_usage {
        base_key_usage {
          digital_signature  = try(var.spec.config.x509_config.key_usage.base_key_usage.digital_signature, false)
          content_commitment = try(var.spec.config.x509_config.key_usage.base_key_usage.content_commitment, false)
          key_encipherment   = try(var.spec.config.x509_config.key_usage.base_key_usage.key_encipherment, false)
          data_encipherment  = try(var.spec.config.x509_config.key_usage.base_key_usage.data_encipherment, false)
          key_agreement      = try(var.spec.config.x509_config.key_usage.base_key_usage.key_agreement, false)
          cert_sign          = try(var.spec.config.x509_config.key_usage.base_key_usage.cert_sign, false)
          crl_sign           = try(var.spec.config.x509_config.key_usage.base_key_usage.crl_sign, false)
          encipher_only      = try(var.spec.config.x509_config.key_usage.base_key_usage.encipher_only, false)
          decipher_only      = try(var.spec.config.x509_config.key_usage.base_key_usage.decipher_only, false)
        }
        extended_key_usage {
          server_auth      = try(var.spec.config.x509_config.key_usage.extended_key_usage.server_auth, false)
          client_auth      = try(var.spec.config.x509_config.key_usage.extended_key_usage.client_auth, false)
          code_signing     = try(var.spec.config.x509_config.key_usage.extended_key_usage.code_signing, false)
          email_protection = try(var.spec.config.x509_config.key_usage.extended_key_usage.email_protection, false)
          time_stamping    = try(var.spec.config.x509_config.key_usage.extended_key_usage.time_stamping, false)
          ocsp_signing     = try(var.spec.config.x509_config.key_usage.extended_key_usage.ocsp_signing, false)
        }
        dynamic "unknown_extended_key_usages" {
          for_each = try(var.spec.config.x509_config.key_usage.unknown_extended_key_usages, [])
          content {
            object_id_path = unknown_extended_key_usages.value.object_id_path
          }
        }
      }

      dynamic "policy_ids" {
        for_each = var.spec.config.x509_config.policy_ids
        content {
          object_id_path = policy_ids.value.object_id_path
        }
      }

      dynamic "additional_extensions" {
        for_each = var.spec.config.x509_config.additional_extensions
        content {
          critical = additional_extensions.value.critical
          value    = additional_extensions.value.value
          object_id {
            object_id_path = additional_extensions.value.object_id.object_id_path
          }
        }
      }

      dynamic "name_constraints" {
        for_each = var.spec.config.x509_config.name_constraints != null ? [var.spec.config.x509_config.name_constraints] : []
        content {
          critical                  = name_constraints.value.critical
          permitted_dns_names       = length(name_constraints.value.permitted_dns_names) > 0 ? name_constraints.value.permitted_dns_names : null
          excluded_dns_names        = length(name_constraints.value.excluded_dns_names) > 0 ? name_constraints.value.excluded_dns_names : null
          permitted_ip_ranges       = length(name_constraints.value.permitted_ip_ranges) > 0 ? name_constraints.value.permitted_ip_ranges : null
          excluded_ip_ranges        = length(name_constraints.value.excluded_ip_ranges) > 0 ? name_constraints.value.excluded_ip_ranges : null
          permitted_email_addresses = length(name_constraints.value.permitted_email_addresses) > 0 ? name_constraints.value.permitted_email_addresses : null
          excluded_email_addresses  = length(name_constraints.value.excluded_email_addresses) > 0 ? name_constraints.value.excluded_email_addresses : null
          permitted_uris            = length(name_constraints.value.permitted_uris) > 0 ? name_constraints.value.permitted_uris : null
          excluded_uris             = length(name_constraints.value.excluded_uris) > 0 ? name_constraints.value.excluded_uris : null
        }
      }
    }
  }

  key_spec {
    algorithm             = var.spec.key_spec.algorithm != "" ? var.spec.key_spec.algorithm : null
    cloud_kms_key_version = var.spec.key_spec.cloud_kms_key_version != "" ? var.spec.key_spec.cloud_kms_key_version : null
  }

  dynamic "subordinate_config" {
    for_each = var.spec.subordinate_config != null ? [var.spec.subordinate_config] : []
    content {
      certificate_authority = subordinate_config.value.certificate_authority != "" ? subordinate_config.value.certificate_authority : null
      dynamic "pem_issuer_chain" {
        for_each = length(subordinate_config.value.pem_issuer_chain) > 0 ? [subordinate_config.value.pem_issuer_chain] : []
        content {
          pem_certificates = pem_issuer_chain.value
        }
      }
    }
  }

  dynamic "user_defined_access_urls" {
    for_each = var.spec.user_defined_access_urls != null ? [var.spec.user_defined_access_urls] : []
    content {
      aia_issuing_certificate_urls = length(user_defined_access_urls.value.aia_issuing_certificate_urls) > 0 ? user_defined_access_urls.value.aia_issuing_certificate_urls : null
      crl_access_urls              = length(user_defined_access_urls.value.crl_access_urls) > 0 ? user_defined_access_urls.value.crl_access_urls : null
    }
  }

  # Client-side destroy behavior: DELETE (the provider's default), PREVENT, or
  # ABANDON. Sent only when set so the provider default stays in charge.
  deletion_policy = local.deletion_policy
}
