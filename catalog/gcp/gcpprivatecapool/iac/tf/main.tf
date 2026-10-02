# Enable the Certificate Authority Service API first so a fresh project works
# on the first deploy. disable_on_destroy is false: tearing down one pool
# must never disable the API for every pool, template, and certificate in the
# project.
resource "google_project_service" "privateca_api" {
  project = local.project_id
  service = "privateca.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The CA pool. The tier is immutable; the issuance policy, publishing options,
# encryption key, and labels update in place.
resource "google_privateca_ca_pool" "this" {
  project  = local.project_id
  location = var.spec.location
  name     = local.ca_pool_id
  tier     = var.spec.tier
  labels   = local.final_labels

  dynamic "issuance_policy" {
    for_each = var.spec.issuance_policy != null ? [var.spec.issuance_policy] : []
    content {
      maximum_lifetime  = issuance_policy.value.maximum_lifetime != "" ? issuance_policy.value.maximum_lifetime : null
      backdate_duration = issuance_policy.value.backdate_duration != "" ? issuance_policy.value.backdate_duration : null

      dynamic "allowed_key_types" {
        for_each = issuance_policy.value.allowed_key_types
        content {
          dynamic "rsa" {
            for_each = allowed_key_types.value.rsa != null ? [allowed_key_types.value.rsa] : []
            content {
              # RSA bounds go as the decimal strings the provider takes; 0 is
              # Google's "no explicit bound" and stays unset.
              min_modulus_size = rsa.value.min_modulus_size > 0 ? tostring(rsa.value.min_modulus_size) : null
              max_modulus_size = rsa.value.max_modulus_size > 0 ? tostring(rsa.value.max_modulus_size) : null
            }
          }
          dynamic "elliptic_curve" {
            for_each = allowed_key_types.value.elliptic_curve_signature_algorithm != "" ? [allowed_key_types.value.elliptic_curve_signature_algorithm] : []
            content {
              signature_algorithm = elliptic_curve.value
            }
          }
        }
      }

      dynamic "allowed_issuance_modes" {
        for_each = issuance_policy.value.allowed_issuance_modes != null ? [issuance_policy.value.allowed_issuance_modes] : []
        content {
          allow_csr_based_issuance    = allowed_issuance_modes.value.allow_csr_based_issuance
          allow_config_based_issuance = allowed_issuance_modes.value.allow_config_based_issuance
        }
      }

      dynamic "identity_constraints" {
        for_each = issuance_policy.value.identity_constraints != null ? [issuance_policy.value.identity_constraints] : []
        content {
          allow_subject_passthrough           = identity_constraints.value.allow_subject_passthrough
          allow_subject_alt_names_passthrough = identity_constraints.value.allow_subject_alt_names_passthrough
          dynamic "cel_expression" {
            for_each = identity_constraints.value.cel_expression != null ? [identity_constraints.value.cel_expression] : []
            content {
              expression  = cel_expression.value.expression
              title       = cel_expression.value.title != "" ? cel_expression.value.title : null
              description = cel_expression.value.description != "" ? cel_expression.value.description : null
              location    = cel_expression.value.location != "" ? cel_expression.value.location : null
            }
          }
        }
      }

      dynamic "baseline_values" {
        for_each = issuance_policy.value.baseline_values != null ? [issuance_policy.value.baseline_values] : []
        content {
          aia_ocsp_servers = length(baseline_values.value.aia_ocsp_servers) > 0 ? baseline_values.value.aia_ocsp_servers : null

          # The provider requires ca_options whenever the block is sent; an
          # omitted one is sent empty, which states nothing.
          dynamic "ca_options" {
            for_each = [baseline_values.value.ca_options != null ? baseline_values.value.ca_options : { is_ca = null, max_issuer_path_length = null }]
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
              digital_signature  = try(baseline_values.value.key_usage.base_key_usage.digital_signature, false)
              content_commitment = try(baseline_values.value.key_usage.base_key_usage.content_commitment, false)
              key_encipherment   = try(baseline_values.value.key_usage.base_key_usage.key_encipherment, false)
              data_encipherment  = try(baseline_values.value.key_usage.base_key_usage.data_encipherment, false)
              key_agreement      = try(baseline_values.value.key_usage.base_key_usage.key_agreement, false)
              cert_sign          = try(baseline_values.value.key_usage.base_key_usage.cert_sign, false)
              crl_sign           = try(baseline_values.value.key_usage.base_key_usage.crl_sign, false)
              encipher_only      = try(baseline_values.value.key_usage.base_key_usage.encipher_only, false)
              decipher_only      = try(baseline_values.value.key_usage.base_key_usage.decipher_only, false)
            }
            extended_key_usage {
              server_auth      = try(baseline_values.value.key_usage.extended_key_usage.server_auth, false)
              client_auth      = try(baseline_values.value.key_usage.extended_key_usage.client_auth, false)
              code_signing     = try(baseline_values.value.key_usage.extended_key_usage.code_signing, false)
              email_protection = try(baseline_values.value.key_usage.extended_key_usage.email_protection, false)
              time_stamping    = try(baseline_values.value.key_usage.extended_key_usage.time_stamping, false)
              ocsp_signing     = try(baseline_values.value.key_usage.extended_key_usage.ocsp_signing, false)
            }
            dynamic "unknown_extended_key_usages" {
              for_each = try(baseline_values.value.key_usage.unknown_extended_key_usages, [])
              content {
                object_id_path = unknown_extended_key_usages.value.object_id_path
              }
            }
          }

          dynamic "policy_ids" {
            for_each = baseline_values.value.policy_ids
            content {
              object_id_path = policy_ids.value.object_id_path
            }
          }

          dynamic "additional_extensions" {
            for_each = baseline_values.value.additional_extensions
            content {
              critical = additional_extensions.value.critical
              value    = additional_extensions.value.value
              object_id {
                object_id_path = additional_extensions.value.object_id.object_id_path
              }
            }
          }

          dynamic "name_constraints" {
            for_each = baseline_values.value.name_constraints != null ? [baseline_values.value.name_constraints] : []
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
    }
  }

  dynamic "publishing_options" {
    for_each = var.spec.publishing_options != null ? [var.spec.publishing_options] : []
    content {
      publish_ca_cert = publishing_options.value.publish_ca_cert
      publish_crl     = publishing_options.value.publish_crl
      encoding_format = publishing_options.value.encoding_format != "" ? publishing_options.value.encoding_format : null
    }
  }

  dynamic "encryption_spec" {
    for_each = local.kms_key_name != null ? [local.kms_key_name] : []
    content {
      cloud_kms_key = encryption_spec.value
    }
  }

  # Client-side destroy behavior: DELETE (the provider's default), PREVENT, or
  # ABANDON. Sent only when set so the provider default stays in charge.
  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.privateca_api]
}
