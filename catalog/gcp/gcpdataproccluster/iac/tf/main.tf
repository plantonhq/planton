# Enable the Dataproc API — the control plane that owns the cluster.
# disable_on_destroy is false: tearing down one cluster must never disable
# the API for everything else in the project.
resource "google_project_service" "dataproc_api" {
  project = local.project_id
  service = "dataproc.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The Dataproc cluster. Two mutually exclusive arms mirror the API:
# cluster_config provisions dedicated Compute Engine VMs; virtual_cluster_config
# runs Dataproc as pods on an existing GKE cluster. Omitting both creates a
# default GCE cluster (2 workers, default machine types).
#
# Mutability: the cluster is create-mostly-immutable. The only in-place
# updates the API supports are labels, primary/secondary worker counts
# (manual scaling), min_num_instances, the autoscaling-policy attachment,
# and the lifecycle TTLs — everything else forces recreation. The virtual
# arm has no update paths at all.
resource "google_dataproc_cluster" "cluster" {
  name    = local.cluster_name
  region  = local.region
  project = local.project_id

  # The Dataproc API rejects user labels on virtual (GKE-based) clusters —
  # labels (including the platform attribution set) are sent only for the
  # GCE arm, identically to the Pulumi module.
  labels = var.spec.virtual_cluster_config == null ? local.final_labels : null

  # Applied when worker counts shrink during an apply: YARN drains running
  # tasks for up to this window before nodes are removed.
  graceful_decommission_timeout = var.spec.graceful_decommission_timeout != "" ? var.spec.graceful_decommission_timeout : null

  # Engine-side teardown behavior (DELETE / PREVENT / ABANDON) — the
  # ABANDON lever hands the cluster to out-of-band management.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  dynamic "cluster_config" {
    for_each = var.spec.cluster_config != null ? [var.spec.cluster_config] : []
    content {
      # Bucket names arrive as resolved references. GCP auto-creates
      # staging/temp buckets when these are unset.
      staging_bucket = cluster_config.value.staging_bucket != "" ? cluster_config.value.staging_bucket : null
      temp_bucket    = cluster_config.value.temp_bucket != "" ? cluster_config.value.temp_bucket : null

      cluster_tier = cluster_config.value.cluster_tier != "" ? cluster_config.value.cluster_tier : null

      # Structural type (STANDARD / SINGLE_NODE / ZERO_SCALE) and execution
      # engine (DEFAULT / LIGHTNING). Both immutable.
      cluster_type = cluster_config.value.cluster_type != "" ? cluster_config.value.cluster_type : null
      engine       = cluster_config.value.engine != "" ? cluster_config.value.engine : null

      # GCE environment: networking, identity, hardening, placement.
      dynamic "gce_cluster_config" {
        for_each = cluster_config.value.gce_config != null ? [cluster_config.value.gce_config] : []
        content {
          network                = gce_cluster_config.value.network != "" ? gce_cluster_config.value.network : null
          subnetwork             = gce_cluster_config.value.subnetwork != "" ? gce_cluster_config.value.subnetwork : null
          service_account        = gce_cluster_config.value.service_account != "" ? gce_cluster_config.value.service_account : null
          service_account_scopes = length(gce_cluster_config.value.service_account_scopes) > 0 ? gce_cluster_config.value.service_account_scopes : null
          zone                   = gce_cluster_config.value.zone != "" ? gce_cluster_config.value.zone : null
          internal_ip_only       = gce_cluster_config.value.internal_ip_only
          tags                   = length(gce_cluster_config.value.tags) > 0 ? gce_cluster_config.value.tags : null
          metadata               = length(gce_cluster_config.value.metadata) > 0 ? gce_cluster_config.value.metadata : null
          # IAM-governed secure tags (distinct from network tags).
          resource_manager_tags = length(gce_cluster_config.value.resource_manager_tags) > 0 ? gce_cluster_config.value.resource_manager_tags : null

          dynamic "shielded_instance_config" {
            for_each = gce_cluster_config.value.shielded_instance_config != null ? [gce_cluster_config.value.shielded_instance_config] : []
            content {
              enable_secure_boot          = shielded_instance_config.value.enable_secure_boot
              enable_vtpm                 = shielded_instance_config.value.enable_vtpm
              enable_integrity_monitoring = shielded_instance_config.value.enable_integrity_monitoring
            }
          }

          dynamic "reservation_affinity" {
            for_each = gce_cluster_config.value.reservation_affinity != null ? [gce_cluster_config.value.reservation_affinity] : []
            content {
              consume_reservation_type = reservation_affinity.value.consume_reservation_type != "" ? reservation_affinity.value.consume_reservation_type : null
              key                      = reservation_affinity.value.key != "" ? reservation_affinity.value.key : null
              values                   = length(reservation_affinity.value.values) > 0 ? reservation_affinity.value.values : null
            }
          }

          dynamic "node_group_affinity" {
            for_each = gce_cluster_config.value.node_group_affinity != null ? [gce_cluster_config.value.node_group_affinity] : []
            content {
              node_group_uri = node_group_affinity.value.node_group_uri
            }
          }

          dynamic "confidential_instance_config" {
            for_each = gce_cluster_config.value.confidential_instance_config != null ? [gce_cluster_config.value.confidential_instance_config] : []
            content {
              # The provider-deprecated boolean is sent only when a manifest
              # still sets it, so a manifest on confidential_instance_type
              # alone never trips the deprecation warning.
              enable_confidential_compute = confidential_instance_config.value.enable_confidential_compute ? true : null
              confidential_instance_type  = confidential_instance_config.value.confidential_instance_type != "" ? confidential_instance_config.value.confidential_instance_type : null
            }
          }
        }
      }

      # Master node group.
      dynamic "master_config" {
        for_each = cluster_config.value.master_config != null ? [cluster_config.value.master_config] : []
        content {
          num_instances = master_config.value.num_instances > 0 ? master_config.value.num_instances : null
          # machine_type never coexists with instance_flexibility_policy
          # (spec CEL): the API drops a paired machineTypeUri from the
          # stored config, and the create-only argument then re-plans as
          # cluster replacement forever.
          machine_type     = master_config.value.machine_type != "" ? master_config.value.machine_type : null
          min_cpu_platform = master_config.value.min_cpu_platform != "" ? master_config.value.min_cpu_platform : null
          image_uri        = master_config.value.image_uri != "" ? master_config.value.image_uri : null

          dynamic "disk_config" {
            for_each = master_config.value.disk_config != null ? [master_config.value.disk_config] : []
            content {
              boot_disk_size_gb   = disk_config.value.boot_disk_size_gb > 0 ? disk_config.value.boot_disk_size_gb : null
              boot_disk_type      = disk_config.value.boot_disk_type != "" ? disk_config.value.boot_disk_type : null
              num_local_ssds      = disk_config.value.num_local_ssds > 0 ? disk_config.value.num_local_ssds : null
              local_ssd_interface = disk_config.value.local_ssd_interface != "" ? disk_config.value.local_ssd_interface : null
              # Provisioned-performance dials (hyperdisk classes).
              boot_disk_provisioned_iops       = disk_config.value.boot_disk_provisioned_iops
              boot_disk_provisioned_throughput = disk_config.value.boot_disk_provisioned_throughput

              # Additional persistent disks on every node of this role.
              dynamic "attached_disk_config" {
                for_each = disk_config.value.attached_disks
                content {
                  disk_size_gb           = attached_disk_config.value.disk_size_gb > 0 ? attached_disk_config.value.disk_size_gb : null
                  disk_type              = attached_disk_config.value.disk_type != "" ? attached_disk_config.value.disk_type : null
                  provisioned_iops       = attached_disk_config.value.provisioned_iops
                  provisioned_throughput = attached_disk_config.value.provisioned_throughput
                }
              }
            }
          }

          dynamic "accelerators" {
            for_each = master_config.value.accelerators
            content {
              accelerator_type  = accelerators.value.accelerator_type
              accelerator_count = accelerators.value.accelerator_count
            }
          }

          # Ranked machine-type fallbacks (masters carry no provisioning
          # mix — that argument exists on secondary workers only).
          dynamic "instance_flexibility_policy" {
            for_each = master_config.value.instance_flexibility_policy != null ? [master_config.value.instance_flexibility_policy] : []
            content {
              dynamic "instance_selection_list" {
                for_each = instance_flexibility_policy.value.instance_selection_list
                content {
                  machine_types = instance_selection_list.value.machine_types
                  rank          = instance_selection_list.value.rank

                  # Per-selection disk shape overriding the role's disk_config.
                  dynamic "disk_config" {
                    for_each = instance_selection_list.value.disk_config != null ? [instance_selection_list.value.disk_config] : []
                    content {
                      boot_disk_size_gb                = disk_config.value.boot_disk_size_gb > 0 ? disk_config.value.boot_disk_size_gb : null
                      boot_disk_type                   = disk_config.value.boot_disk_type != "" ? disk_config.value.boot_disk_type : null
                      num_local_ssds                   = disk_config.value.num_local_ssds > 0 ? disk_config.value.num_local_ssds : null
                      local_ssd_interface              = disk_config.value.local_ssd_interface != "" ? disk_config.value.local_ssd_interface : null
                      boot_disk_provisioned_iops       = disk_config.value.boot_disk_provisioned_iops
                      boot_disk_provisioned_throughput = disk_config.value.boot_disk_provisioned_throughput

                      dynamic "attached_disk_config" {
                        for_each = disk_config.value.attached_disks
                        content {
                          disk_size_gb           = attached_disk_config.value.disk_size_gb > 0 ? attached_disk_config.value.disk_size_gb : null
                          disk_type              = attached_disk_config.value.disk_type != "" ? attached_disk_config.value.disk_type : null
                          provisioned_iops       = attached_disk_config.value.provisioned_iops
                          provisioned_throughput = attached_disk_config.value.provisioned_throughput
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }

      # Primary worker node group. num_instances / min_num_instances are
      # the manual-scaling levers — the only node counts that update in
      # place (with graceful_decommission_timeout honored on shrink).
      dynamic "worker_config" {
        for_each = cluster_config.value.worker_config != null ? [cluster_config.value.worker_config] : []
        content {
          num_instances = worker_config.value.num_instances > 0 ? worker_config.value.num_instances : null
          # machine_type never coexists with instance_flexibility_policy
          # (spec CEL) — see the master_config note.
          machine_type      = worker_config.value.machine_type != "" ? worker_config.value.machine_type : null
          min_cpu_platform  = worker_config.value.min_cpu_platform != "" ? worker_config.value.min_cpu_platform : null
          image_uri         = worker_config.value.image_uri != "" ? worker_config.value.image_uri : null
          min_num_instances = worker_config.value.min_num_instances > 0 ? worker_config.value.min_num_instances : null

          dynamic "disk_config" {
            for_each = worker_config.value.disk_config != null ? [worker_config.value.disk_config] : []
            content {
              boot_disk_size_gb   = disk_config.value.boot_disk_size_gb > 0 ? disk_config.value.boot_disk_size_gb : null
              boot_disk_type      = disk_config.value.boot_disk_type != "" ? disk_config.value.boot_disk_type : null
              num_local_ssds      = disk_config.value.num_local_ssds > 0 ? disk_config.value.num_local_ssds : null
              local_ssd_interface = disk_config.value.local_ssd_interface != "" ? disk_config.value.local_ssd_interface : null
              # Provisioned-performance dials (hyperdisk classes).
              boot_disk_provisioned_iops       = disk_config.value.boot_disk_provisioned_iops
              boot_disk_provisioned_throughput = disk_config.value.boot_disk_provisioned_throughput

              # Additional persistent disks on every node of this role.
              dynamic "attached_disk_config" {
                for_each = disk_config.value.attached_disks
                content {
                  disk_size_gb           = attached_disk_config.value.disk_size_gb > 0 ? attached_disk_config.value.disk_size_gb : null
                  disk_type              = attached_disk_config.value.disk_type != "" ? attached_disk_config.value.disk_type : null
                  provisioned_iops       = attached_disk_config.value.provisioned_iops
                  provisioned_throughput = attached_disk_config.value.provisioned_throughput
                }
              }
            }
          }

          dynamic "accelerators" {
            for_each = worker_config.value.accelerators
            content {
              accelerator_type  = accelerators.value.accelerator_type
              accelerator_count = accelerators.value.accelerator_count
            }
          }

          # Ranked machine-type fallbacks (primary workers carry no
          # provisioning mix — that argument exists on secondary workers
          # only).
          dynamic "instance_flexibility_policy" {
            for_each = worker_config.value.instance_flexibility_policy != null ? [worker_config.value.instance_flexibility_policy] : []
            content {
              dynamic "instance_selection_list" {
                for_each = instance_flexibility_policy.value.instance_selection_list
                content {
                  machine_types = instance_selection_list.value.machine_types
                  rank          = instance_selection_list.value.rank

                  # Per-selection disk shape overriding the role's disk_config.
                  dynamic "disk_config" {
                    for_each = instance_selection_list.value.disk_config != null ? [instance_selection_list.value.disk_config] : []
                    content {
                      boot_disk_size_gb                = disk_config.value.boot_disk_size_gb > 0 ? disk_config.value.boot_disk_size_gb : null
                      boot_disk_type                   = disk_config.value.boot_disk_type != "" ? disk_config.value.boot_disk_type : null
                      num_local_ssds                   = disk_config.value.num_local_ssds > 0 ? disk_config.value.num_local_ssds : null
                      local_ssd_interface              = disk_config.value.local_ssd_interface != "" ? disk_config.value.local_ssd_interface : null
                      boot_disk_provisioned_iops       = disk_config.value.boot_disk_provisioned_iops
                      boot_disk_provisioned_throughput = disk_config.value.boot_disk_provisioned_throughput

                      dynamic "attached_disk_config" {
                        for_each = disk_config.value.attached_disks
                        content {
                          disk_size_gb           = attached_disk_config.value.disk_size_gb > 0 ? attached_disk_config.value.disk_size_gb : null
                          disk_type              = attached_disk_config.value.disk_type != "" ? attached_disk_config.value.disk_type : null
                          provisioned_iops       = attached_disk_config.value.provisioned_iops
                          provisioned_throughput = attached_disk_config.value.provisioned_throughput
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }

      # Secondary (preemptible/spot) worker group. The count updates in
      # place; preemptibility is immutable. Machine shape is inherited from
      # the primary workers unless the flexibility policy overrides it.
      dynamic "preemptible_worker_config" {
        for_each = cluster_config.value.secondary_worker_config != null ? [cluster_config.value.secondary_worker_config] : []
        content {
          num_instances  = preemptible_worker_config.value.num_instances > 0 ? preemptible_worker_config.value.num_instances : null
          preemptibility = preemptible_worker_config.value.preemptibility != "" ? preemptible_worker_config.value.preemptibility : null

          dynamic "disk_config" {
            for_each = preemptible_worker_config.value.disk_config != null ? [preemptible_worker_config.value.disk_config] : []
            content {
              boot_disk_size_gb   = disk_config.value.boot_disk_size_gb > 0 ? disk_config.value.boot_disk_size_gb : null
              boot_disk_type      = disk_config.value.boot_disk_type != "" ? disk_config.value.boot_disk_type : null
              num_local_ssds      = disk_config.value.num_local_ssds > 0 ? disk_config.value.num_local_ssds : null
              local_ssd_interface = disk_config.value.local_ssd_interface != "" ? disk_config.value.local_ssd_interface : null
              # Provisioned-performance dials (hyperdisk classes).
              boot_disk_provisioned_iops       = disk_config.value.boot_disk_provisioned_iops
              boot_disk_provisioned_throughput = disk_config.value.boot_disk_provisioned_throughput

              # Additional persistent disks on every node of this role.
              dynamic "attached_disk_config" {
                for_each = disk_config.value.attached_disks
                content {
                  disk_size_gb           = attached_disk_config.value.disk_size_gb > 0 ? attached_disk_config.value.disk_size_gb : null
                  disk_type              = attached_disk_config.value.disk_type != "" ? attached_disk_config.value.disk_type : null
                  provisioned_iops       = attached_disk_config.value.provisioned_iops
                  provisioned_throughput = attached_disk_config.value.provisioned_throughput
                }
              }
            }
          }

          dynamic "instance_flexibility_policy" {
            for_each = preemptible_worker_config.value.instance_flexibility_policy != null ? [preemptible_worker_config.value.instance_flexibility_policy] : []
            content {
              dynamic "instance_selection_list" {
                for_each = instance_flexibility_policy.value.instance_selection_list
                content {
                  machine_types = instance_selection_list.value.machine_types
                  rank          = instance_selection_list.value.rank

                  # Per-selection disk shape overriding the role's disk_config.
                  dynamic "disk_config" {
                    for_each = instance_selection_list.value.disk_config != null ? [instance_selection_list.value.disk_config] : []
                    content {
                      boot_disk_size_gb                = disk_config.value.boot_disk_size_gb > 0 ? disk_config.value.boot_disk_size_gb : null
                      boot_disk_type                   = disk_config.value.boot_disk_type != "" ? disk_config.value.boot_disk_type : null
                      num_local_ssds                   = disk_config.value.num_local_ssds > 0 ? disk_config.value.num_local_ssds : null
                      local_ssd_interface              = disk_config.value.local_ssd_interface != "" ? disk_config.value.local_ssd_interface : null
                      boot_disk_provisioned_iops       = disk_config.value.boot_disk_provisioned_iops
                      boot_disk_provisioned_throughput = disk_config.value.boot_disk_provisioned_throughput

                      dynamic "attached_disk_config" {
                        for_each = disk_config.value.attached_disks
                        content {
                          disk_size_gb           = attached_disk_config.value.disk_size_gb > 0 ? attached_disk_config.value.disk_size_gb : null
                          disk_type              = attached_disk_config.value.disk_type != "" ? attached_disk_config.value.disk_type : null
                          provisioned_iops       = attached_disk_config.value.provisioned_iops
                          provisioned_throughput = attached_disk_config.value.provisioned_throughput
                        }
                      }
                    }
                  }
                }
              }

              dynamic "provisioning_model_mix" {
                for_each = instance_flexibility_policy.value.provisioning_model_mix != null ? [instance_flexibility_policy.value.provisioning_model_mix] : []
                content {
                  standard_capacity_base               = provisioning_model_mix.value.standard_capacity_base
                  standard_capacity_percent_above_base = provisioning_model_mix.value.standard_capacity_percent_above_base
                }
              }
            }
          }
        }
      }

      # Image version, optional components, and framework properties. The
      # spec's `properties` map feeds the provider's override_properties —
      # the API's writable surface (the provider's `properties` attribute
      # is the computed resolved set).
      dynamic "software_config" {
        for_each = cluster_config.value.software_config != null ? [cluster_config.value.software_config] : []
        content {
          image_version       = software_config.value.image_version != "" ? software_config.value.image_version : null
          optional_components = length(software_config.value.optional_components) > 0 ? software_config.value.optional_components : null
          override_properties = length(software_config.value.properties) > 0 ? software_config.value.properties : null
        }
      }

      dynamic "initialization_action" {
        for_each = cluster_config.value.initialization_actions
        content {
          script      = initialization_action.value.script
          timeout_sec = initialization_action.value.timeout_sec > 0 ? initialization_action.value.timeout_sec : null
        }
      }

      # Autoscaling policy attachment (a first-class resource, referenced by
      # its full resource name). Attach/swap/detach updates in place.
      dynamic "autoscaling_config" {
        for_each = cluster_config.value.autoscaling_policy_uri != "" ? [1] : []
        content {
          policy_uri = cluster_config.value.autoscaling_policy_uri
        }
      }

      # CMEK for all persistent disks. Changing the key forces recreation.
      dynamic "encryption_config" {
        for_each = cluster_config.value.encryption_kms_key_name != "" ? [1] : []
        content {
          kms_key_name = cluster_config.value.encryption_kms_key_name
        }
      }

      # Kerberos XOR personal-cluster identity mapping. Kerberos secret
      # fields are GCS URIs of KMS-encrypted files — never inline material.
      dynamic "security_config" {
        for_each = cluster_config.value.security_config != null ? [cluster_config.value.security_config] : []
        content {
          dynamic "kerberos_config" {
            for_each = security_config.value.kerberos_config != null ? [security_config.value.kerberos_config] : []
            content {
              enable_kerberos                       = kerberos_config.value.enable_kerberos
              root_principal_password_uri           = kerberos_config.value.root_principal_password_uri
              kms_key_uri                           = kerberos_config.value.kms_key_uri
              realm                                 = kerberos_config.value.realm != "" ? kerberos_config.value.realm : null
              tgt_lifetime_hours                    = kerberos_config.value.tgt_lifetime_hours > 0 ? kerberos_config.value.tgt_lifetime_hours : null
              kdc_db_key_uri                        = kerberos_config.value.kdc_db_key_uri != "" ? kerberos_config.value.kdc_db_key_uri : null
              keystore_uri                          = kerberos_config.value.keystore_uri != "" ? kerberos_config.value.keystore_uri : null
              keystore_password_uri                 = kerberos_config.value.keystore_password_uri != "" ? kerberos_config.value.keystore_password_uri : null
              key_password_uri                      = kerberos_config.value.key_password_uri != "" ? kerberos_config.value.key_password_uri : null
              truststore_uri                        = kerberos_config.value.truststore_uri != "" ? kerberos_config.value.truststore_uri : null
              truststore_password_uri               = kerberos_config.value.truststore_password_uri != "" ? kerberos_config.value.truststore_password_uri : null
              cross_realm_trust_realm               = kerberos_config.value.cross_realm_trust_realm != "" ? kerberos_config.value.cross_realm_trust_realm : null
              cross_realm_trust_kdc                 = kerberos_config.value.cross_realm_trust_kdc != "" ? kerberos_config.value.cross_realm_trust_kdc : null
              cross_realm_trust_admin_server        = kerberos_config.value.cross_realm_trust_admin_server != "" ? kerberos_config.value.cross_realm_trust_admin_server : null
              cross_realm_trust_shared_password_uri = kerberos_config.value.cross_realm_trust_shared_password_uri != "" ? kerberos_config.value.cross_realm_trust_shared_password_uri : null
            }
          }

          dynamic "identity_config" {
            for_each = security_config.value.identity_config != null ? [security_config.value.identity_config] : []
            content {
              user_service_account_mapping = identity_config.value.user_service_account_mapping
            }
          }
        }
      }

      # Component Gateway (authenticated web UIs).
      dynamic "endpoint_config" {
        for_each = cluster_config.value.endpoint_config != null ? [cluster_config.value.endpoint_config] : []
        content {
          enable_http_port_access = endpoint_config.value.enable_http_port_access
        }
      }

      # Cost-control levers — all update in place. Delete destroys the
      # cluster; stop shuts the VMs down and keeps it restartable.
      dynamic "lifecycle_config" {
        for_each = cluster_config.value.lifecycle_config != null ? [cluster_config.value.lifecycle_config] : []
        content {
          idle_delete_ttl  = lifecycle_config.value.idle_delete_ttl != "" ? lifecycle_config.value.idle_delete_ttl : null
          auto_delete_time = lifecycle_config.value.auto_delete_time != "" ? lifecycle_config.value.auto_delete_time : null
          idle_stop_ttl    = lifecycle_config.value.idle_stop_ttl != "" ? lifecycle_config.value.idle_stop_ttl : null
          auto_stop_time   = lifecycle_config.value.auto_stop_time != "" ? lifecycle_config.value.auto_stop_time : null
        }
      }

      # Persistent shared Hive metastore.
      dynamic "metastore_config" {
        for_each = cluster_config.value.metastore_config != null ? [cluster_config.value.metastore_config] : []
        content {
          dataproc_metastore_service = metastore_config.value.dataproc_metastore_service
        }
      }

      # OSS metric collection into Cloud Monitoring.
      dynamic "dataproc_metric_config" {
        for_each = cluster_config.value.dataproc_metric_config != null ? [cluster_config.value.dataproc_metric_config] : []
        content {
          dynamic "metrics" {
            for_each = dataproc_metric_config.value.metrics
            content {
              metric_source    = metrics.value.metric_source
              metric_overrides = length(metrics.value.metric_overrides) > 0 ? metrics.value.metric_overrides : null
            }
          }
        }
      }

      # Dedicated DRIVER node groups.
      dynamic "auxiliary_node_groups" {
        for_each = cluster_config.value.auxiliary_node_groups
        content {
          node_group_id = auxiliary_node_groups.value.node_group_id != "" ? auxiliary_node_groups.value.node_group_id : null

          node_group {
            roles = auxiliary_node_groups.value.roles

            dynamic "node_group_config" {
              for_each = auxiliary_node_groups.value.node_group_config != null ? [auxiliary_node_groups.value.node_group_config] : []
              content {
                num_instances    = node_group_config.value.num_instances > 0 ? node_group_config.value.num_instances : null
                machine_type     = node_group_config.value.machine_type != "" ? node_group_config.value.machine_type : null
                min_cpu_platform = node_group_config.value.min_cpu_platform != "" ? node_group_config.value.min_cpu_platform : null

                dynamic "disk_config" {
                  for_each = node_group_config.value.disk_config != null ? [node_group_config.value.disk_config] : []
                  content {
                    boot_disk_size_gb   = disk_config.value.boot_disk_size_gb > 0 ? disk_config.value.boot_disk_size_gb : null
                    boot_disk_type      = disk_config.value.boot_disk_type != "" ? disk_config.value.boot_disk_type : null
                    num_local_ssds      = disk_config.value.num_local_ssds > 0 ? disk_config.value.num_local_ssds : null
                    local_ssd_interface = disk_config.value.local_ssd_interface != "" ? disk_config.value.local_ssd_interface : null
                    # Provisioned-performance dials (hyperdisk classes).
                    boot_disk_provisioned_iops       = disk_config.value.boot_disk_provisioned_iops
                    boot_disk_provisioned_throughput = disk_config.value.boot_disk_provisioned_throughput
                  }
                }

                dynamic "accelerators" {
                  for_each = node_group_config.value.accelerators
                  content {
                    accelerator_type  = accelerators.value.accelerator_type
                    accelerator_count = accelerators.value.accelerator_count
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  # Dataproc-on-GKE: Spark runs as pods on an existing GKE cluster. All
  # references arrive resolved to the fully qualified resource names the
  # Dataproc API requires. The whole arm is immutable — changes replace
  # the virtual cluster without touching the underlying GKE resources.
  dynamic "virtual_cluster_config" {
    for_each = var.spec.virtual_cluster_config != null ? [var.spec.virtual_cluster_config] : []
    content {
      staging_bucket = virtual_cluster_config.value.staging_bucket != "" ? virtual_cluster_config.value.staging_bucket : null

      kubernetes_cluster_config {
        kubernetes_namespace = virtual_cluster_config.value.kubernetes_cluster_config.kubernetes_namespace != "" ? virtual_cluster_config.value.kubernetes_cluster_config.kubernetes_namespace : null

        gke_cluster_config {
          gke_cluster_target = virtual_cluster_config.value.kubernetes_cluster_config.gke_cluster_config.gke_cluster_target

          dynamic "node_pool_target" {
            for_each = virtual_cluster_config.value.kubernetes_cluster_config.gke_cluster_config.node_pool_target
            content {
              node_pool = node_pool_target.value.node_pool
              roles     = node_pool_target.value.roles

              dynamic "node_pool_config" {
                for_each = node_pool_target.value.node_pool_config != null ? [node_pool_target.value.node_pool_config] : []
                content {
                  locations = node_pool_config.value.locations

                  dynamic "autoscaling" {
                    for_each = node_pool_config.value.autoscaling != null ? [node_pool_config.value.autoscaling] : []
                    content {
                      min_node_count = node_pool_config.value.autoscaling.min_node_count
                      max_node_count = node_pool_config.value.autoscaling.max_node_count
                    }
                  }

                  dynamic "config" {
                    for_each = (
                      node_pool_config.value.machine_type != "" ||
                      node_pool_config.value.local_ssd_count > 0 ||
                      node_pool_config.value.min_cpu_platform != "" ||
                      node_pool_config.value.preemptible ||
                      node_pool_config.value.spot
                    ) ? [1] : []
                    content {
                      machine_type     = node_pool_config.value.machine_type != "" ? node_pool_config.value.machine_type : null
                      local_ssd_count  = node_pool_config.value.local_ssd_count > 0 ? node_pool_config.value.local_ssd_count : null
                      min_cpu_platform = node_pool_config.value.min_cpu_platform != "" ? node_pool_config.value.min_cpu_platform : null
                      preemptible      = node_pool_config.value.preemptible ? true : null
                      spot             = node_pool_config.value.spot ? true : null
                    }
                  }
                }
              }
            }
          }
        }

        kubernetes_software_config {
          component_version = virtual_cluster_config.value.kubernetes_cluster_config.kubernetes_software_config.component_version
          properties        = length(virtual_cluster_config.value.kubernetes_cluster_config.kubernetes_software_config.properties) > 0 ? virtual_cluster_config.value.kubernetes_cluster_config.kubernetes_software_config.properties : null
        }
      }

      dynamic "auxiliary_services_config" {
        for_each = virtual_cluster_config.value.auxiliary_services_config != null ? [virtual_cluster_config.value.auxiliary_services_config] : []
        content {
          dynamic "metastore_config" {
            for_each = auxiliary_services_config.value.metastore_config != null ? [auxiliary_services_config.value.metastore_config] : []
            content {
              dataproc_metastore_service = metastore_config.value.dataproc_metastore_service
            }
          }

          dynamic "spark_history_server_config" {
            for_each = auxiliary_services_config.value.spark_history_server_config != null ? [auxiliary_services_config.value.spark_history_server_config] : []
            content {
              dataproc_cluster = spark_history_server_config.value.dataproc_cluster != "" ? spark_history_server_config.value.dataproc_cluster : null
            }
          }
        }
      }
    }
  }

  depends_on = [
    google_project_service.dataproc_api,
  ]
}
