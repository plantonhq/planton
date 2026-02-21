resource "oci_containerengine_cluster" "this" {
  compartment_id     = var.spec.compartment_id.value
  vcn_id             = var.spec.vcn_id.value
  kubernetes_version = var.spec.kubernetes_version
  name               = local.display_name
  freeform_tags      = local.freeform_tags

  type = var.spec.type != "" ? lookup(local.cluster_type_map, var.spec.type, var.spec.type) : null

  kms_key_id = var.spec.kms_key_id != null ? var.spec.kms_key_id.value : null

  dynamic "cluster_pod_network_options" {
    for_each = var.spec.cni_type != "" ? [var.spec.cni_type] : []
    content {
      cni_type = lookup(local.cni_type_map, cluster_pod_network_options.value, cluster_pod_network_options.value)
    }
  }

  dynamic "endpoint_config" {
    for_each = var.spec.endpoint_config != null ? [var.spec.endpoint_config] : []
    content {
      subnet_id            = endpoint_config.value.subnet_id.value
      is_public_ip_enabled = endpoint_config.value.is_public_ip_enabled
      nsg_ids              = local.endpoint_nsg_ids
    }
  }

  dynamic "image_policy_config" {
    for_each = var.spec.image_policy_config != null ? [var.spec.image_policy_config] : []
    content {
      is_policy_enabled = image_policy_config.value.is_policy_enabled

      dynamic "key_details" {
        for_each = image_policy_config.value.key_details
        content {
          kms_key_id = key_details.value.kms_key_id.value
        }
      }
    }
  }

  dynamic "options" {
    for_each = var.spec.options != null ? [var.spec.options] : []
    content {
      service_lb_subnet_ids = length(local.service_lb_subnet_ids) > 0 ? local.service_lb_subnet_ids : null
      ip_families           = length(local.ip_families) > 0 ? local.ip_families : null

      dynamic "kubernetes_network_config" {
        for_each = options.value.kubernetes_network_config != null ? [options.value.kubernetes_network_config] : []
        content {
          pods_cidr     = kubernetes_network_config.value.pods_cidr != "" ? kubernetes_network_config.value.pods_cidr : null
          services_cidr = kubernetes_network_config.value.services_cidr != "" ? kubernetes_network_config.value.services_cidr : null
        }
      }

      dynamic "service_lb_config" {
        for_each = options.value.service_lb_config != null ? [options.value.service_lb_config] : []
        content {
          backend_nsg_ids = length(local.service_lb_backend_nsg_ids) > 0 ? local.service_lb_backend_nsg_ids : null
          freeform_tags   = length(service_lb_config.value.freeform_tags) > 0 ? service_lb_config.value.freeform_tags : null
          defined_tags    = length(service_lb_config.value.defined_tags) > 0 ? service_lb_config.value.defined_tags : null
        }
      }

      dynamic "persistent_volume_config" {
        for_each = options.value.persistent_volume_config != null ? [options.value.persistent_volume_config] : []
        content {
          freeform_tags = length(persistent_volume_config.value.freeform_tags) > 0 ? persistent_volume_config.value.freeform_tags : null
          defined_tags  = length(persistent_volume_config.value.defined_tags) > 0 ? persistent_volume_config.value.defined_tags : null
        }
      }

      dynamic "open_id_connect_token_authentication_config" {
        for_each = options.value.open_id_connect_token_authentication_config != null ? [options.value.open_id_connect_token_authentication_config] : []
        content {
          is_open_id_connect_auth_enabled = open_id_connect_token_authentication_config.value.is_open_id_connect_auth_enabled
          configuration_file              = open_id_connect_token_authentication_config.value.configuration_file != "" ? open_id_connect_token_authentication_config.value.configuration_file : null
          issuer_url                      = open_id_connect_token_authentication_config.value.issuer_url != "" ? open_id_connect_token_authentication_config.value.issuer_url : null
          client_id                       = open_id_connect_token_authentication_config.value.client_id != "" ? open_id_connect_token_authentication_config.value.client_id : null
          ca_certificate                  = open_id_connect_token_authentication_config.value.ca_certificate != "" ? open_id_connect_token_authentication_config.value.ca_certificate : null
          username_claim                  = open_id_connect_token_authentication_config.value.username_claim != "" ? open_id_connect_token_authentication_config.value.username_claim : null
          username_prefix                 = open_id_connect_token_authentication_config.value.username_prefix != "" ? open_id_connect_token_authentication_config.value.username_prefix : null
          groups_claim                    = open_id_connect_token_authentication_config.value.groups_claim != "" ? open_id_connect_token_authentication_config.value.groups_claim : null
          groups_prefix                   = open_id_connect_token_authentication_config.value.groups_prefix != "" ? open_id_connect_token_authentication_config.value.groups_prefix : null
          signing_algorithms              = length(open_id_connect_token_authentication_config.value.signing_algorithms) > 0 ? open_id_connect_token_authentication_config.value.signing_algorithms : null

          dynamic "required_claims" {
            for_each = open_id_connect_token_authentication_config.value.required_claims
            content {
              key   = required_claims.value.key
              value = required_claims.value.value
            }
          }
        }
      }

      dynamic "open_id_connect_discovery" {
        for_each = options.value.is_open_id_connect_discovery_enabled ? [true] : []
        content {
          is_open_id_connect_discovery_enabled = true
        }
      }
    }
  }
}
