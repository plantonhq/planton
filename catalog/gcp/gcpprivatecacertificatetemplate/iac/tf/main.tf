# Enable the Certificate Authority Service API first so a fresh project works
# on the first deploy. disable_on_destroy is false: tearing down one template
# must never disable the API for every pool, template, and certificate in the
# project.
resource "google_project_service" "privateca_api" {
  project = local.project_id
  service = "privateca.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The certificate template. Everything but the location and ID updates in
# place.
resource "google_privateca_certificate_template" "this" {
  project          = local.project_id
  location         = var.spec.location
  name             = local.template_id
  description      = local.description
  maximum_lifetime = local.maximum_lifetime
  labels           = local.final_labels

  dynamic "predefined_values" {
    for_each = var.spec.predefined_values != null ? [var.spec.predefined_values] : []
    content {
      aia_ocsp_servers = length(predefined_values.value.aia_ocsp_servers) > 0 ? predefined_values.value.aia_ocsp_servers : null

      dynamic "ca_options" {
        for_each = predefined_values.value.ca_options != null ? [predefined_values.value.ca_options] : []
        content {
          # Unset is_ca omits the CA flag; the template resource would send
          # false unless null_ca is set, so unset maps to null_ca.
          is_ca   = ca_options.value.is_ca
          null_ca = ca_options.value.is_ca == null ? true : null
          # A path length of 0 goes through zero_max_issuer_path_length: the
          # provider reads a plain 0 as unset.
          max_issuer_path_length      = coalesce(ca_options.value.max_issuer_path_length, 0) > 0 ? ca_options.value.max_issuer_path_length : null
          zero_max_issuer_path_length = ca_options.value.max_issuer_path_length == 0 ? true : null
        }
      }

      dynamic "key_usage" {
        for_each = predefined_values.value.key_usage != null ? [predefined_values.value.key_usage] : []
        content {
          dynamic "base_key_usage" {
            for_each = key_usage.value.base_key_usage != null ? [key_usage.value.base_key_usage] : []
            content {
              digital_signature  = base_key_usage.value.digital_signature
              content_commitment = base_key_usage.value.content_commitment
              key_encipherment   = base_key_usage.value.key_encipherment
              data_encipherment  = base_key_usage.value.data_encipherment
              key_agreement      = base_key_usage.value.key_agreement
              cert_sign          = base_key_usage.value.cert_sign
              crl_sign           = base_key_usage.value.crl_sign
              encipher_only      = base_key_usage.value.encipher_only
              decipher_only      = base_key_usage.value.decipher_only
            }
          }
          dynamic "extended_key_usage" {
            for_each = key_usage.value.extended_key_usage != null ? [key_usage.value.extended_key_usage] : []
            content {
              server_auth      = extended_key_usage.value.server_auth
              client_auth      = extended_key_usage.value.client_auth
              code_signing     = extended_key_usage.value.code_signing
              email_protection = extended_key_usage.value.email_protection
              time_stamping    = extended_key_usage.value.time_stamping
              ocsp_signing     = extended_key_usage.value.ocsp_signing
            }
          }
          dynamic "unknown_extended_key_usages" {
            for_each = key_usage.value.unknown_extended_key_usages
            content {
              object_id_path = unknown_extended_key_usages.value.object_id_path
            }
          }
        }
      }

      dynamic "policy_ids" {
        for_each = predefined_values.value.policy_ids
        content {
          object_id_path = policy_ids.value.object_id_path
        }
      }

      dynamic "additional_extensions" {
        for_each = predefined_values.value.additional_extensions
        content {
          critical = additional_extensions.value.critical
          value    = additional_extensions.value.value
          object_id {
            object_id_path = additional_extensions.value.object_id.object_id_path
          }
        }
      }

      dynamic "name_constraints" {
        for_each = predefined_values.value.name_constraints != null ? [predefined_values.value.name_constraints] : []
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

  dynamic "identity_constraints" {
    for_each = var.spec.identity_constraints != null ? [var.spec.identity_constraints] : []
    content {
      allow_subject_passthrough           = identity_constraints.value.allow_subject_passthrough
      allow_subject_alt_names_passthrough = identity_constraints.value.allow_subject_alt_names_passthrough
      dynamic "cel_expression" {
        for_each = identity_constraints.value.cel_expression != null ? [identity_constraints.value.cel_expression] : []
        content {
          expression  = cel_expression.value.expression != "" ? cel_expression.value.expression : null
          title       = cel_expression.value.title != "" ? cel_expression.value.title : null
          description = cel_expression.value.description != "" ? cel_expression.value.description : null
          location    = cel_expression.value.location != "" ? cel_expression.value.location : null
        }
      }
    }
  }

  dynamic "passthrough_extensions" {
    for_each = var.spec.passthrough_extensions != null ? [var.spec.passthrough_extensions] : []
    content {
      known_extensions = length(passthrough_extensions.value.known_extensions) > 0 ? passthrough_extensions.value.known_extensions : null
      dynamic "additional_extensions" {
        for_each = passthrough_extensions.value.additional_extensions
        content {
          object_id_path = additional_extensions.value.object_id_path
        }
      }
    }
  }

  # Client-side destroy behavior: DELETE (the provider's default), PREVENT, or
  # ABANDON. Sent only when set so the provider default stays in charge.
  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.privateca_api]
}
