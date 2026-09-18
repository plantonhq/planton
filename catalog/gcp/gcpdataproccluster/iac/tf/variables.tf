variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for the GCP Dataproc cluster"
  type = object({
    # The GCP project for the cluster. The CLI's tfvars converter resolves
    # StringValueOrRef fields to their literal string before the module
    # runs, so this arrives as a plain string. If empty, the provider's
    # default project is used (see locals.tf).
    project_id = optional(string, "")

    # Region for the cluster (e.g. us-central1). Immutable (ForceNew).
    region = string

    # Cluster name. Immutable (ForceNew).
    cluster_name = string

    # YARN graceful decommission window applied when worker counts shrink.
    graceful_decommission_timeout = optional(string, "")

    # User labels merged beneath Planton platform labels (platform keys win
    # on conflict). The Dataproc API does not support labels on virtual
    # clusters — the spec validation rejects that combination pre-deploy.
    labels = optional(map(string), {})

    # Engine-side teardown behavior: DELETE, PREVENT, or ABANDON.
    deletion_policy = optional(string, "")

    # The standard Compute Engine arm. Mutually exclusive with
    # virtual_cluster_config (enforced pre-deploy by the spec validation;
    # the API enforces it server-side too).
    cluster_config = optional(object({
      # Staging/temp buckets arrive as plain bucket names (resolved refs).
      staging_bucket = optional(string, "")
      temp_bucket    = optional(string, "")

      # CLUSTER_TIER_STANDARD or CLUSTER_TIER_PREMIUM. Immutable.
      cluster_tier = optional(string, "")

      gce_config = optional(object({
        # network XOR subnetwork (resolved refs — self-links).
        network                = optional(string, "")
        subnetwork             = optional(string, "")
        service_account        = optional(string, "")
        service_account_scopes = optional(list(string), [])
        zone                   = optional(string, "")
        internal_ip_only       = optional(bool, false)
        tags                   = optional(list(string), [])
        metadata               = optional(map(string), {})
        shielded_instance_config = optional(object({
          enable_secure_boot          = optional(bool, false)
          enable_vtpm                 = optional(bool, false)
          enable_integrity_monitoring = optional(bool, false)
        }), null)
        reservation_affinity = optional(object({
          consume_reservation_type = optional(string, "")
          key                      = optional(string, "")
          values                   = optional(list(string), [])
        }), null)
        node_group_affinity = optional(object({
          node_group_uri = string
        }), null)
        confidential_instance_config = optional(object({
          enable_confidential_compute = optional(bool, false)
          # SEV, SEV_SNP, or TDX; the current shape (the boolean is the
          # provider-deprecated predecessor).
          confidential_instance_type = optional(string, "")
        }), null)
        # IAM-governed secure tags (tagKeys/... = tagValues/...).
        resource_manager_tags = optional(map(string), {})
      }), null)

      # Structural type: STANDARD, SINGLE_NODE, or ZERO_SCALE. Immutable.
      cluster_type = optional(string, "")

      # Execution engine: DEFAULT or LIGHTNING. Immutable.
      engine = optional(string, "")

      master_config = optional(object({
        num_instances    = optional(number, 0)
        machine_type     = optional(string, "")
        min_cpu_platform = optional(string, "")
        image_uri        = optional(string, "")
        disk_config = optional(object({
          boot_disk_size_gb   = optional(number, 0)
          boot_disk_type      = optional(string, "")
          num_local_ssds      = optional(number, 0)
          local_ssd_interface = optional(string, "")
          # Provisioned-performance dials (hyperdisk classes).
          boot_disk_provisioned_iops       = optional(number, null)
          boot_disk_provisioned_throughput = optional(number, null)
          # Additional persistent disks on every node (ForceNew).
          attached_disks = optional(list(object({
            disk_size_gb           = optional(number, 0)
            disk_type              = optional(string, "")
            provisioned_iops       = optional(number, null)
            provisioned_throughput = optional(number, null)
          })), [])
        }), null)
        accelerators = optional(list(object({
          accelerator_type  = string
          accelerator_count = number
        })), [])
        # Ranked machine-type fallbacks for master provisioning (no
        # provisioning mix — masters are always on-demand).
        instance_flexibility_policy = optional(object({
          instance_selection_list = optional(list(object({
            machine_types = list(string)
            rank          = optional(number, 0)
            # Per-selection disk shape overriding the role's disk_config.
            disk_config = optional(object({
              boot_disk_size_gb                = optional(number, 0)
              boot_disk_type                   = optional(string, "")
              num_local_ssds                   = optional(number, 0)
              local_ssd_interface              = optional(string, "")
              boot_disk_provisioned_iops       = optional(number, null)
              boot_disk_provisioned_throughput = optional(number, null)
              attached_disks = optional(list(object({
                disk_size_gb           = optional(number, 0)
                disk_type              = optional(string, "")
                provisioned_iops       = optional(number, null)
                provisioned_throughput = optional(number, null)
              })), [])
            }), null)
          })), [])
        }), null)
      }), null)

      worker_config = optional(object({
        # num_instances and min_num_instances are the manual-scaling levers —
        # the only node-count fields that update in place.
        num_instances     = optional(number, 0)
        machine_type      = optional(string, "")
        min_cpu_platform  = optional(string, "")
        image_uri         = optional(string, "")
        min_num_instances = optional(number, 0)
        disk_config = optional(object({
          boot_disk_size_gb   = optional(number, 0)
          boot_disk_type      = optional(string, "")
          num_local_ssds      = optional(number, 0)
          local_ssd_interface = optional(string, "")
          # Provisioned-performance dials (hyperdisk classes).
          boot_disk_provisioned_iops       = optional(number, null)
          boot_disk_provisioned_throughput = optional(number, null)
          # Additional persistent disks on every node (ForceNew).
          attached_disks = optional(list(object({
            disk_size_gb           = optional(number, 0)
            disk_type              = optional(string, "")
            provisioned_iops       = optional(number, null)
            provisioned_throughput = optional(number, null)
          })), [])
        }), null)
        accelerators = optional(list(object({
          accelerator_type  = string
          accelerator_count = number
        })), [])
        # Ranked machine-type fallbacks for primary-worker provisioning (no
        # provisioning mix — primary workers are always on-demand).
        instance_flexibility_policy = optional(object({
          instance_selection_list = optional(list(object({
            machine_types = list(string)
            rank          = optional(number, 0)
            # Per-selection disk shape overriding the role's disk_config.
            disk_config = optional(object({
              boot_disk_size_gb                = optional(number, 0)
              boot_disk_type                   = optional(string, "")
              num_local_ssds                   = optional(number, 0)
              local_ssd_interface              = optional(string, "")
              boot_disk_provisioned_iops       = optional(number, null)
              boot_disk_provisioned_throughput = optional(number, null)
              attached_disks = optional(list(object({
                disk_size_gb           = optional(number, 0)
                disk_type              = optional(string, "")
                provisioned_iops       = optional(number, null)
                provisioned_throughput = optional(number, null)
              })), [])
            }), null)
          })), [])
        }), null)
      }), null)

      secondary_worker_config = optional(object({
        num_instances  = optional(number, 0)
        preemptibility = optional(string, "")
        disk_config = optional(object({
          boot_disk_size_gb   = optional(number, 0)
          boot_disk_type      = optional(string, "")
          num_local_ssds      = optional(number, 0)
          local_ssd_interface = optional(string, "")
          # Provisioned-performance dials (hyperdisk classes).
          boot_disk_provisioned_iops       = optional(number, null)
          boot_disk_provisioned_throughput = optional(number, null)
          # Additional persistent disks on every node (ForceNew).
          attached_disks = optional(list(object({
            disk_size_gb           = optional(number, 0)
            disk_type              = optional(string, "")
            provisioned_iops       = optional(number, null)
            provisioned_throughput = optional(number, null)
          })), [])
        }), null)
        # Machine-type flexibility + standard/spot mix — only the secondary
        # group supports flexible provisioning on the released line.
        instance_flexibility_policy = optional(object({
          instance_selection_list = optional(list(object({
            machine_types = list(string)
            rank          = optional(number, 0)
            # Per-selection disk shape overriding the role's disk_config.
            disk_config = optional(object({
              boot_disk_size_gb                = optional(number, 0)
              boot_disk_type                   = optional(string, "")
              num_local_ssds                   = optional(number, 0)
              local_ssd_interface              = optional(string, "")
              boot_disk_provisioned_iops       = optional(number, null)
              boot_disk_provisioned_throughput = optional(number, null)
              attached_disks = optional(list(object({
                disk_size_gb           = optional(number, 0)
                disk_type              = optional(string, "")
                provisioned_iops       = optional(number, null)
                provisioned_throughput = optional(number, null)
              })), [])
            }), null)
          })), [])
          provisioning_model_mix = optional(object({
            standard_capacity_base               = optional(number, 0)
            standard_capacity_percent_above_base = optional(number, 0)
          }), null)
        }), null)
      }), null)

      software_config = optional(object({
        image_version       = optional(string, "")
        optional_components = optional(list(string), [])
        # Maps to the provider's override_properties (the API's writable
        # properties surface; the provider's `properties` attribute is the
        # computed resolved set).
        properties = optional(map(string), {})
      }), null)

      initialization_actions = optional(list(object({
        script      = string
        timeout_sec = optional(number, 0)
      })), [])

      # Resolved autoscaling-policy resource name
      # (projects/{p}/locations/{l}/autoscalingPolicies/{id}).
      # Attach/swap/detach updates in place.
      autoscaling_policy_uri = optional(string, "")

      # Resolved CMEK key resource ID. Empty means Google-managed keys.
      encryption_kms_key_name = optional(string, "")

      # Kerberos XOR identity mapping (exactly one, enforced pre-deploy).
      # Kerberos secret fields are GCS URIs of KMS-encrypted files — paths,
      # never inline secret material (the API's own contract).
      security_config = optional(object({
        kerberos_config = optional(object({
          enable_kerberos                       = optional(bool, false)
          root_principal_password_uri           = string
          kms_key_uri                           = string
          realm                                 = optional(string, "")
          tgt_lifetime_hours                    = optional(number, 0)
          kdc_db_key_uri                        = optional(string, "")
          keystore_uri                          = optional(string, "")
          keystore_password_uri                 = optional(string, "")
          key_password_uri                      = optional(string, "")
          truststore_uri                        = optional(string, "")
          truststore_password_uri               = optional(string, "")
          cross_realm_trust_realm               = optional(string, "")
          cross_realm_trust_kdc                 = optional(string, "")
          cross_realm_trust_admin_server        = optional(string, "")
          cross_realm_trust_shared_password_uri = optional(string, "")
        }), null)
        identity_config = optional(object({
          user_service_account_mapping = map(string)
        }), null)
      }), null)

      endpoint_config = optional(object({
        enable_http_port_access = optional(bool, false)
      }), null)

      # All four levers update in place — delete destroys the cluster,
      # stop shuts the VMs down and keeps it restartable.
      lifecycle_config = optional(object({
        idle_delete_ttl  = optional(string, "")
        auto_delete_time = optional(string, "")
        idle_stop_ttl    = optional(string, "")
        auto_stop_time   = optional(string, "")
      }), null)

      # Resolved Dataproc Metastore service resource name.
      metastore_config = optional(object({
        dataproc_metastore_service = string
      }), null)

      dataproc_metric_config = optional(object({
        metrics = list(object({
          metric_source    = string
          metric_overrides = optional(list(string), [])
        }))
      }), null)

      auxiliary_node_groups = optional(list(object({
        roles = list(string)
        node_group_config = optional(object({
          num_instances    = optional(number, 0)
          machine_type     = optional(string, "")
          min_cpu_platform = optional(string, "")
          disk_config = optional(object({
            boot_disk_size_gb   = optional(number, 0)
            boot_disk_type      = optional(string, "")
            num_local_ssds      = optional(number, 0)
            local_ssd_interface = optional(string, "")
            # Provisioned-performance dials (hyperdisk classes).
            boot_disk_provisioned_iops       = optional(number, null)
            boot_disk_provisioned_throughput = optional(number, null)
            # Present for shape parity with the shared disk message; the
            # spec rejects attached disks on auxiliary groups and the
            # provider has no argument for them here.
            attached_disks = optional(list(object({
              disk_size_gb           = optional(number, 0)
              disk_type              = optional(string, "")
              provisioned_iops       = optional(number, null)
              provisioned_throughput = optional(number, null)
            })), [])
          }), null)
          accelerators = optional(list(object({
            accelerator_type  = string
            accelerator_count = number
          })), [])
        }), null)
        node_group_id = optional(string, "")
      })), [])
    }), null)

    # The Dataproc-on-GKE arm. Fully immutable (any change replaces the
    # virtual cluster; the underlying GKE resources are untouched).
    virtual_cluster_config = optional(object({
      staging_bucket = optional(string, "")
      kubernetes_cluster_config = object({
        # Resolved Kubernetes namespace name; empty lets Dataproc derive one.
        kubernetes_namespace = optional(string, "")
        gke_cluster_config = object({
          # Resolved fully qualified GKE cluster resource name
          # (projects/{p}/locations/{l}/clusters/{c}).
          gke_cluster_target = string
          node_pool_target = optional(list(object({
            # Resolved fully qualified node-pool resource name.
            node_pool = string
            roles     = list(string)
            node_pool_config = optional(object({
              locations = list(string)
              autoscaling = optional(object({
                min_node_count = optional(number, 0)
                max_node_count = optional(number, 0)
              }), null)
              machine_type     = optional(string, "")
              local_ssd_count  = optional(number, 0)
              min_cpu_platform = optional(string, "")
              preemptible      = optional(bool, false)
              spot             = optional(bool, false)
            }), null)
          })), [])
        })
        kubernetes_software_config = object({
          component_version = map(string)
          properties        = optional(map(string), {})
        })
      })
      auxiliary_services_config = optional(object({
        metastore_config = optional(object({
          dataproc_metastore_service = string
        }), null)
        spark_history_server_config = optional(object({
          # Resolved fully qualified Dataproc cluster resource name
          # (projects/{p}/regions/{r}/clusters/{c}).
          dataproc_cluster = optional(string, "")
        }), null)
      }), null)
    }), null)
  })

  validation {
    condition     = var.spec.region != ""
    error_message = "region is required."
  }

  validation {
    condition     = var.spec.cluster_name != ""
    error_message = "cluster_name is required."
  }

  validation {
    condition     = !(var.spec.cluster_config != null && var.spec.virtual_cluster_config != null)
    error_message = "cluster_config and virtual_cluster_config are mutually exclusive."
  }
}
