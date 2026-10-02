# The certificate, issued from a CSR or from structured config. Everything
# but labels is immutable; destroy revokes it.
resource "google_privateca_certificate" "this" {
  project               = local.project_id
  location              = var.spec.location
  pool                  = local.pool
  name                  = local.certificate_id
  certificate_authority = local.certificate_authority
  certificate_template  = local.certificate_template
  lifetime              = local.lifetime
  pem_csr               = local.pem_csr
  labels                = local.final_labels

  dynamic "config" {
    for_each = var.spec.config != null ? [var.spec.config] : []
    content {
      subject_config {
        subject {
          common_name         = config.value.subject_config.subject.common_name
          country_code        = config.value.subject_config.subject.country_code != "" ? config.value.subject_config.subject.country_code : null
          organization        = config.value.subject_config.subject.organization != "" ? config.value.subject_config.subject.organization : null
          organizational_unit = config.value.subject_config.subject.organizational_unit != "" ? config.value.subject_config.subject.organizational_unit : null
          locality            = config.value.subject_config.subject.locality != "" ? config.value.subject_config.subject.locality : null
          province            = config.value.subject_config.subject.province != "" ? config.value.subject_config.subject.province : null
          street_address      = config.value.subject_config.subject.street_address != "" ? config.value.subject_config.subject.street_address : null
          postal_code         = config.value.subject_config.subject.postal_code != "" ? config.value.subject_config.subject.postal_code : null
        }
        dynamic "subject_alt_name" {
          for_each = config.value.subject_config.subject_alt_name != null ? [config.value.subject_config.subject_alt_name] : []
          content {
            dns_names       = length(subject_alt_name.value.dns_names) > 0 ? subject_alt_name.value.dns_names : null
            uris            = length(subject_alt_name.value.uris) > 0 ? subject_alt_name.value.uris : null
            email_addresses = length(subject_alt_name.value.email_addresses) > 0 ? subject_alt_name.value.email_addresses : null
            ip_addresses    = length(subject_alt_name.value.ip_addresses) > 0 ? subject_alt_name.value.ip_addresses : null
          }
        }
      }

      dynamic "subject_key_id" {
        for_each = config.value.subject_key_id != "" ? [config.value.subject_key_id] : []
        content {
          key_id = subject_key_id.value
        }
      }

      x509_config {
        aia_ocsp_servers = length(config.value.x509_config.aia_ocsp_servers) > 0 ? config.value.x509_config.aia_ocsp_servers : null

        dynamic "ca_options" {
          for_each = config.value.x509_config.ca_options != null ? [config.value.x509_config.ca_options] : []
          content {
            # Unset is_ca omits the CA flag; false is sent with non_ca, the
            # provider's way to state CA:FALSE rather than leave it out.
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
            digital_signature  = try(config.value.x509_config.key_usage.base_key_usage.digital_signature, false)
            content_commitment = try(config.value.x509_config.key_usage.base_key_usage.content_commitment, false)
            key_encipherment   = try(config.value.x509_config.key_usage.base_key_usage.key_encipherment, false)
            data_encipherment  = try(config.value.x509_config.key_usage.base_key_usage.data_encipherment, false)
            key_agreement      = try(config.value.x509_config.key_usage.base_key_usage.key_agreement, false)
            cert_sign          = try(config.value.x509_config.key_usage.base_key_usage.cert_sign, false)
            crl_sign           = try(config.value.x509_config.key_usage.base_key_usage.crl_sign, false)
            encipher_only      = try(config.value.x509_config.key_usage.base_key_usage.encipher_only, false)
            decipher_only      = try(config.value.x509_config.key_usage.base_key_usage.decipher_only, false)
          }
          extended_key_usage {
            server_auth      = try(config.value.x509_config.key_usage.extended_key_usage.server_auth, false)
            client_auth      = try(config.value.x509_config.key_usage.extended_key_usage.client_auth, false)
            code_signing     = try(config.value.x509_config.key_usage.extended_key_usage.code_signing, false)
            email_protection = try(config.value.x509_config.key_usage.extended_key_usage.email_protection, false)
            time_stamping    = try(config.value.x509_config.key_usage.extended_key_usage.time_stamping, false)
            ocsp_signing     = try(config.value.x509_config.key_usage.extended_key_usage.ocsp_signing, false)
          }
          dynamic "unknown_extended_key_usages" {
            for_each = try(config.value.x509_config.key_usage.unknown_extended_key_usages, [])
            content {
              object_id_path = unknown_extended_key_usages.value.object_id_path
            }
          }
        }

        dynamic "policy_ids" {
          for_each = config.value.x509_config.policy_ids
          content {
            object_id_path = policy_ids.value.object_id_path
          }
        }

        dynamic "additional_extensions" {
          for_each = config.value.x509_config.additional_extensions
          content {
            critical = additional_extensions.value.critical
            value    = additional_extensions.value.value
            object_id {
              object_id_path = additional_extensions.value.object_id.object_id_path
            }
          }
        }

        dynamic "name_constraints" {
          for_each = config.value.x509_config.name_constraints != null ? [config.value.x509_config.name_constraints] : []
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

      public_key {
        # PEM is the only key format Google accepts; an empty format sends it.
        format = config.value.public_key.format != "" ? config.value.public_key.format : "PEM"
        key    = config.value.public_key.key
      }
    }
  }

  # Client-side destroy behavior: DELETE (the provider's default, which
  # revokes), PREVENT, or ABANDON. Sent only when set so the provider default
  # stays in charge.
  deletion_policy = local.deletion_policy
}
