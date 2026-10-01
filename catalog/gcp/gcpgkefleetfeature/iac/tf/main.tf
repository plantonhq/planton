# The Fleet API (GKE Hub), and the feature's own API when it has one.
# disable_on_destroy is false: turning one feature off must never disable an
# API the rest of the fleet uses.
resource "google_project_service" "gkehub_api" {
  project = local.project_id
  service = "gkehub.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

resource "google_project_service" "feature_api" {
  count = local.feature_api != null ? 1 : 0

  project = local.project_id
  service = local.feature_api

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The feature. Creating a feature that is already on adopts it; destroy
# turns it off (rbacrolebindingactuation only empties its allowlist, which
# the provider does itself).
resource "google_gke_hub_feature" "this" {
  project         = local.project_id
  name            = var.spec.feature
  location        = local.location
  labels          = local.final_labels
  deletion_policy = local.deletion_policy

  dynamic "spec" {
    for_each = local.has_feature_spec ? [1] : []
    content {
      dynamic "multiclusteringress" {
        for_each = var.spec.multiclusteringress != null ? [var.spec.multiclusteringress] : []
        content {
          config_membership = multiclusteringress.value.config_membership
        }
      }

      dynamic "fleetobservability" {
        for_each = var.spec.fleetobservability != null ? [var.spec.fleetobservability] : []
        content {
          dynamic "logging_config" {
            for_each = fleetobservability.value.logging_config != null ? [fleetobservability.value.logging_config] : []
            content {
              dynamic "default_config" {
                for_each = logging_config.value.default_config != null ? [logging_config.value.default_config] : []
                content {
                  mode = default_config.value.mode != "" ? default_config.value.mode : null
                }
              }
              dynamic "fleet_scope_logs_config" {
                for_each = logging_config.value.fleet_scope_logs_config != null ? [logging_config.value.fleet_scope_logs_config] : []
                content {
                  mode = fleet_scope_logs_config.value.mode != "" ? fleet_scope_logs_config.value.mode : null
                }
              }
            }
          }
        }
      }

      dynamic "clusterupgrade" {
        for_each = var.spec.clusterupgrade != null ? [var.spec.clusterupgrade] : []
        content {
          upstream_fleets = clusterupgrade.value.upstream_fleets

          # Optional+Computed: sent only when declared, so Google's default
          # never shows as a diff.
          dynamic "post_conditions" {
            for_each = clusterupgrade.value.post_conditions != null ? [clusterupgrade.value.post_conditions] : []
            content {
              soaking = post_conditions.value.soaking
            }
          }

          dynamic "gke_upgrade_overrides" {
            for_each = clusterupgrade.value.gke_upgrade_overrides
            content {
              upgrade {
                name    = gke_upgrade_overrides.value.upgrade.name
                version = gke_upgrade_overrides.value.upgrade.version
              }
              post_conditions {
                soaking = gke_upgrade_overrides.value.post_conditions.soaking
              }
            }
          }
        }
      }

      dynamic "rbacrolebindingactuation" {
        for_each = var.spec.rbacrolebindingactuation != null ? [var.spec.rbacrolebindingactuation] : []
        content {
          allowed_custom_roles = rbacrolebindingactuation.value.allowed_custom_roles
        }
      }

      dynamic "workloadidentity" {
        for_each = var.spec.workloadidentity != null ? [var.spec.workloadidentity] : []
        content {
          scope_tenancy_pool = workloadidentity.value.scope_tenancy_pool != "" ? workloadidentity.value.scope_tenancy_pool : null
        }
      }
    }
  }

  # The fleet-wide defaults Google applies to every cluster in the fleet,
  # including clusters that join later.
  dynamic "fleet_default_member_config" {
    for_each = local.default_member_config != null ? [local.default_member_config] : []
    content {
      dynamic "configmanagement" {
        for_each = fleet_default_member_config.value.configmanagement != null ? [fleet_default_member_config.value.configmanagement] : []
        content {
          management = configmanagement.value.management != "" ? configmanagement.value.management : null
          version    = configmanagement.value.version != "" ? configmanagement.value.version : null

          dynamic "config_sync" {
            for_each = configmanagement.value.config_sync != null ? [configmanagement.value.config_sync] : []
            content {
              enabled                           = config_sync.value.enabled
              metrics_gcp_service_account_email = config_sync.value.metrics_gcp_service_account_email != "" ? config_sync.value.metrics_gcp_service_account_email : null
              prevent_drift                     = config_sync.value.prevent_drift
              source_format                     = config_sync.value.source_format != "" ? config_sync.value.source_format : null

              dynamic "git" {
                for_each = config_sync.value.git != null ? [config_sync.value.git] : []
                content {
                  secret_type               = git.value.secret_type
                  gcp_service_account_email = git.value.gcp_service_account_email != "" ? git.value.gcp_service_account_email : null
                  https_proxy               = git.value.https_proxy != "" ? git.value.https_proxy : null
                  policy_dir                = git.value.policy_dir != "" ? git.value.policy_dir : null
                  sync_branch               = git.value.sync_branch != "" ? git.value.sync_branch : null
                  sync_repo                 = git.value.sync_repo
                  sync_rev                  = git.value.sync_rev != "" ? git.value.sync_rev : null
                  # The provider takes the period as a decimal string.
                  sync_wait_secs = git.value.sync_wait_secs > 0 ? tostring(git.value.sync_wait_secs) : null
                }
              }

              dynamic "oci" {
                for_each = config_sync.value.oci != null ? [config_sync.value.oci] : []
                content {
                  secret_type               = oci.value.secret_type
                  gcp_service_account_email = oci.value.gcp_service_account_email != "" ? oci.value.gcp_service_account_email : null
                  policy_dir                = oci.value.policy_dir != "" ? oci.value.policy_dir : null
                  sync_repo                 = oci.value.sync_repo
                  sync_wait_secs            = oci.value.sync_wait_secs > 0 ? tostring(oci.value.sync_wait_secs) : null
                }
              }
            }
          }
        }
      }

      dynamic "mesh" {
        for_each = fleet_default_member_config.value.mesh != null ? [fleet_default_member_config.value.mesh] : []
        content {
          management = mesh.value.management
        }
      }

      dynamic "policycontroller" {
        for_each = fleet_default_member_config.value.policycontroller != null ? [fleet_default_member_config.value.policycontroller] : []
        content {
          version = policycontroller.value.version != "" ? policycontroller.value.version : null

          policy_controller_hub_config {
            install_spec               = policycontroller.value.policy_controller_hub_config.install_spec
            audit_interval_seconds     = policycontroller.value.policy_controller_hub_config.audit_interval_seconds
            constraint_violation_limit = policycontroller.value.policy_controller_hub_config.constraint_violation_limit
            exemptable_namespaces      = length(policycontroller.value.policy_controller_hub_config.exemptable_namespaces) > 0 ? policycontroller.value.policy_controller_hub_config.exemptable_namespaces : null
            log_denies_enabled         = policycontroller.value.policy_controller_hub_config.log_denies_enabled
            mutation_enabled           = policycontroller.value.policy_controller_hub_config.mutation_enabled
            referential_rules_enabled  = policycontroller.value.policy_controller_hub_config.referential_rules_enabled

            dynamic "monitoring" {
              for_each = policycontroller.value.policy_controller_hub_config.monitoring != null ? [policycontroller.value.policy_controller_hub_config.monitoring] : []
              content {
                backends = monitoring.value.backends
              }
            }

            dynamic "deployment_configs" {
              for_each = policycontroller.value.policy_controller_hub_config.deployment_configs
              content {
                component     = deployment_configs.value.component
                replica_count = deployment_configs.value.replica_count
                pod_affinity  = deployment_configs.value.pod_affinity != "" ? deployment_configs.value.pod_affinity : null

                dynamic "container_resources" {
                  for_each = deployment_configs.value.container_resources != null ? [deployment_configs.value.container_resources] : []
                  content {
                    dynamic "limits" {
                      for_each = container_resources.value.limits != null ? [container_resources.value.limits] : []
                      content {
                        cpu    = limits.value.cpu != "" ? limits.value.cpu : null
                        memory = limits.value.memory != "" ? limits.value.memory : null
                      }
                    }
                    dynamic "requests" {
                      for_each = container_resources.value.requests != null ? [container_resources.value.requests] : []
                      content {
                        cpu    = requests.value.cpu != "" ? requests.value.cpu : null
                        memory = requests.value.memory != "" ? requests.value.memory : null
                      }
                    }
                  }
                }

                # The fleet-default resource names the block in the singular.
                dynamic "pod_toleration" {
                  for_each = deployment_configs.value.pod_tolerations
                  content {
                    effect   = pod_toleration.value.effect != "" ? pod_toleration.value.effect : null
                    key      = pod_toleration.value.key != "" ? pod_toleration.value.key : null
                    operator = pod_toleration.value.operator != "" ? pod_toleration.value.operator : null
                    value    = pod_toleration.value.value != "" ? pod_toleration.value.value : null
                  }
                }
              }
            }

            dynamic "policy_content" {
              for_each = policycontroller.value.policy_controller_hub_config.policy_content != null ? [policycontroller.value.policy_controller_hub_config.policy_content] : []
              content {
                dynamic "bundles" {
                  for_each = policy_content.value.bundles
                  content {
                    bundle              = bundles.value.bundle
                    exempted_namespaces = length(bundles.value.exempted_namespaces) > 0 ? bundles.value.exempted_namespaces : null
                  }
                }
                dynamic "template_library" {
                  for_each = policy_content.value.template_library != null ? [policy_content.value.template_library] : []
                  content {
                    installation = template_library.value.installation != "" ? template_library.value.installation : null
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  depends_on = [google_project_service.gkehub_api, google_project_service.feature_api]
}

# Per-cluster settings, one entry of the feature's membershipSpecs map per
# spec.membership_configs[] entry. Destroy empties the entry and leaves the
# feature and the membership alone.
resource "google_gke_hub_feature_membership" "this" {
  for_each = local.membership_configs

  project             = local.project_id
  location            = local.location
  feature             = google_gke_hub_feature.this.name
  membership          = each.value.membership_id
  membership_location = each.value.membership_location
  deletion_policy     = local.deletion_policy

  dynamic "configmanagement" {
    for_each = each.value.configmanagement != null ? [each.value.configmanagement] : []
    content {
      management = configmanagement.value.management != "" ? configmanagement.value.management : null
      version    = configmanagement.value.version != "" ? configmanagement.value.version : null

      dynamic "config_sync" {
        for_each = configmanagement.value.config_sync != null ? [configmanagement.value.config_sync] : []
        content {
          enabled                           = config_sync.value.enabled
          metrics_gcp_service_account_email = config_sync.value.metrics_gcp_service_account_email != "" ? config_sync.value.metrics_gcp_service_account_email : null
          prevent_drift                     = config_sync.value.prevent_drift
          source_format                     = config_sync.value.source_format != "" ? config_sync.value.source_format : null
          stop_syncing                      = config_sync.value.stop_syncing

          dynamic "git" {
            for_each = config_sync.value.git != null ? [config_sync.value.git] : []
            content {
              secret_type               = git.value.secret_type
              gcp_service_account_email = git.value.gcp_service_account_email != "" ? git.value.gcp_service_account_email : null
              https_proxy               = git.value.https_proxy != "" ? git.value.https_proxy : null
              policy_dir                = git.value.policy_dir != "" ? git.value.policy_dir : null
              sync_branch               = git.value.sync_branch != "" ? git.value.sync_branch : null
              sync_repo                 = git.value.sync_repo
              sync_rev                  = git.value.sync_rev != "" ? git.value.sync_rev : null
              sync_wait_secs            = git.value.sync_wait_secs > 0 ? tostring(git.value.sync_wait_secs) : null
            }
          }

          dynamic "oci" {
            for_each = config_sync.value.oci != null ? [config_sync.value.oci] : []
            content {
              secret_type               = oci.value.secret_type
              gcp_service_account_email = oci.value.gcp_service_account_email != "" ? oci.value.gcp_service_account_email : null
              policy_dir                = oci.value.policy_dir != "" ? oci.value.policy_dir : null
              sync_repo                 = oci.value.sync_repo
              sync_wait_secs            = oci.value.sync_wait_secs > 0 ? tostring(oci.value.sync_wait_secs) : null
            }
          }

          dynamic "deployment_overrides" {
            for_each = config_sync.value.deployment_overrides
            content {
              deployment_name      = deployment_overrides.value.deployment_name != "" ? deployment_overrides.value.deployment_name : null
              deployment_namespace = deployment_overrides.value.deployment_namespace != "" ? deployment_overrides.value.deployment_namespace : null

              dynamic "containers" {
                for_each = deployment_overrides.value.containers
                content {
                  container_name = containers.value.container_name != "" ? containers.value.container_name : null
                  cpu_limit      = containers.value.cpu_limit != "" ? containers.value.cpu_limit : null
                  cpu_request    = containers.value.cpu_request != "" ? containers.value.cpu_request : null
                  memory_limit   = containers.value.memory_limit != "" ? containers.value.memory_limit : null
                  memory_request = containers.value.memory_request != "" ? containers.value.memory_request : null
                }
              }
            }
          }
        }
      }
    }
  }

  dynamic "mesh" {
    for_each = each.value.mesh != null ? [each.value.mesh] : []
    content {
      management = mesh.value.management
    }
  }

  dynamic "policycontroller" {
    for_each = each.value.policycontroller != null ? [each.value.policycontroller] : []
    content {
      version = policycontroller.value.version != "" ? policycontroller.value.version : null

      policy_controller_hub_config {
        install_spec               = policycontroller.value.policy_controller_hub_config.install_spec
        audit_interval_seconds     = policycontroller.value.policy_controller_hub_config.audit_interval_seconds
        constraint_violation_limit = policycontroller.value.policy_controller_hub_config.constraint_violation_limit
        exemptable_namespaces      = length(policycontroller.value.policy_controller_hub_config.exemptable_namespaces) > 0 ? policycontroller.value.policy_controller_hub_config.exemptable_namespaces : null
        log_denies_enabled         = policycontroller.value.policy_controller_hub_config.log_denies_enabled
        mutation_enabled           = policycontroller.value.policy_controller_hub_config.mutation_enabled
        referential_rules_enabled  = policycontroller.value.policy_controller_hub_config.referential_rules_enabled

        dynamic "monitoring" {
          for_each = policycontroller.value.policy_controller_hub_config.monitoring != null ? [policycontroller.value.policy_controller_hub_config.monitoring] : []
          content {
            backends = monitoring.value.backends
          }
        }

        # The membership resource names the component component_name and
        # the tolerations block in the plural.
        dynamic "deployment_configs" {
          for_each = policycontroller.value.policy_controller_hub_config.deployment_configs
          content {
            component_name = deployment_configs.value.component
            replica_count  = deployment_configs.value.replica_count
            pod_affinity   = deployment_configs.value.pod_affinity != "" ? deployment_configs.value.pod_affinity : null

            dynamic "container_resources" {
              for_each = deployment_configs.value.container_resources != null ? [deployment_configs.value.container_resources] : []
              content {
                dynamic "limits" {
                  for_each = container_resources.value.limits != null ? [container_resources.value.limits] : []
                  content {
                    cpu    = limits.value.cpu != "" ? limits.value.cpu : null
                    memory = limits.value.memory != "" ? limits.value.memory : null
                  }
                }
                dynamic "requests" {
                  for_each = container_resources.value.requests != null ? [container_resources.value.requests] : []
                  content {
                    cpu    = requests.value.cpu != "" ? requests.value.cpu : null
                    memory = requests.value.memory != "" ? requests.value.memory : null
                  }
                }
              }
            }

            dynamic "pod_tolerations" {
              for_each = deployment_configs.value.pod_tolerations
              content {
                effect   = pod_tolerations.value.effect != "" ? pod_tolerations.value.effect : null
                key      = pod_tolerations.value.key != "" ? pod_tolerations.value.key : null
                operator = pod_tolerations.value.operator != "" ? pod_tolerations.value.operator : null
                value    = pod_tolerations.value.value != "" ? pod_tolerations.value.value : null
              }
            }
          }
        }

        dynamic "policy_content" {
          for_each = policycontroller.value.policy_controller_hub_config.policy_content != null ? [policycontroller.value.policy_controller_hub_config.policy_content] : []
          content {
            # The membership resource names a bundle bundle_name.
            dynamic "bundles" {
              for_each = policy_content.value.bundles
              content {
                bundle_name         = bundles.value.bundle
                exempted_namespaces = length(bundles.value.exempted_namespaces) > 0 ? bundles.value.exempted_namespaces : null
              }
            }
            dynamic "template_library" {
              for_each = policy_content.value.template_library != null ? [policy_content.value.template_library] : []
              content {
                installation = template_library.value.installation != "" ? template_library.value.installation : null
              }
            }
          }
        }
      }
    }
  }
}
