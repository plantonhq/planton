variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "GcpGkeFleetFeature specification"
  type = object({
    # The fleet host project: a literal project ID or a GcpGkeFleet
    # reference (the fleet this feature configures, which orders it after
    # the fleet in a chart). Empty means the provider's default project.
    # Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Which feature, by Google's feature name (see the list above): one of
    # the documented names, or another lowercase name Google adds. Immutable.
    feature = string

    # Where the feature lives. Fleet features are "global" (the default).
    # Immutable.
    location = optional(string, "")

    # Labels on the feature. The platform attribution labels are added on
    # top and win on key conflicts.
    labels = optional(map(string), {})

    # Settings for "multiclusteringress".
    multiclusteringress = optional(object({
      # The config cluster's fleet membership, by full name: a
      # GcpGkeFleetMembership reference, a GcpGkeCluster reference (its
      # fleet_membership output), or the literal name. Required.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      config_membership = string
    }))

    # Settings for "fleetobservability".
    fleetobservability = optional(object({
      # The routing configuration.
      logging_config = optional(object({
        # Routing for logs no other route covers.
        default_config = optional(object({
          # How the logs reach the fleet host project:
          #   "COPY" -- copied there and kept in the cluster's project too
          #   "MOVE" -- moved there only
          # Empty leaves the route off.
          mode = optional(string, "")
        }))

        # Routing for all logs of every fleet scope.
        fleet_scope_logs_config = optional(object({
          # How the logs reach the fleet host project:
          #   "COPY" -- copied there and kept in the cluster's project too
          #   "MOVE" -- moved there only
          # Empty leaves the route off.
          mode = optional(string, "")
        }))
      }))
    }))

    # Settings for "clusterupgrade".
    clusterupgrade = optional(object({
      # The upstream fleet whose completed upgrades this fleet consumes: its
      # host project's ID or number, as a literal or a GcpGkeFleet reference.
      # Google accepts at most one today; the list leaves room for more.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      upstream_fleets = list(string)

      # When an upgrade counts as complete in this fleet. Empty takes Google's
      # default.
      post_conditions = optional(object({
        # How long to soak after the rollout finishes before marking it
        # complete, as seconds with an "s" suffix ("604800s" is seven days); at
        # most 30 days. Required.
        soaking = string
      }))

      # Different completion conditions for individual upgrades.
      gke_upgrade_overrides = optional(list(object({
        # Which upgrade. Required.
        upgrade = object({
          # The upgrade's name, e.g. "k8s_control_plane" or "k8s_node". At most 99
          # characters. Required.
          name = string

          # The upgrade's version, e.g. "1.31.1-gke.1146000". At most 99
          # characters. Required.
          version = string
        })

        # Its completion conditions. Required.
        post_conditions = object({
          # How long to soak after the rollout finishes before marking it
          # complete, as seconds with an "s" suffix ("604800s" is seven days); at
          # most 30 days. Required.
          soaking = string
        })
      })), [])
    }))

    # Settings for "rbacrolebindingactuation".
    rbacrolebindingactuation = optional(object({
      # The Kubernetes ClusterRoles a scope role binding may grant through
      # custom_role. A role in use cannot leave the list until the bindings
      # that use it are gone.
      allowed_custom_roles = optional(list(string), [])
    }))

    # Settings for "workloadidentity".
    workloadidentity = optional(object({
      # The workload identity pool, in trust-domain mode, that fleet tenancy
      # uses so identities stay the same across the clusters of a scope: a
      # GcpWorkloadIdentityPool reference (its name) or the literal pool name.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      scope_tenancy_pool = optional(string, "")
    }))

    # The fleet-wide defaults for "configmanagement", "servicemesh", or
    # "policycontroller": applied to every cluster in the fleet, including
    # clusters that join later, unless membership_configs overrides one.
    fleet_default_member_config = optional(object({
      # Config Sync for every cluster (feature "configmanagement").
      configmanagement = optional(object({
        # Upgrades: "MANAGEMENT_AUTOMATIC" lets Google keep Config Sync current;
        # "MANAGEMENT_MANUAL" pins it to `version`. Empty takes Google's default.
        management = optional(string, "")

        # The Config Sync version to install, e.g. "1.22.0"; meaningful with
        # MANAGEMENT_MANUAL. Empty takes the current release.
        version = optional(string, "")

        # What to sync and how.
        config_sync = optional(object({
          # true installs Config Sync and applies these settings; false removes it
          # and ignores the rest. Unset installs it exactly when git or oci is set.
          enabled = optional(bool)

          # Sync from a Git repository.
          git = optional(object({
            # How Config Sync authenticates to the repository:
            #   "none" (public repository), "ssh", "cookiefile", "token",
            #   "gcenode" (the node's service account), "gcpserviceaccount" (a
            #   Google service account through Workload Identity, see
            #   gcp_service_account_email), "githubapp".
            # Required; Google matches it case-sensitively.
            secret_type = string

            # The Google service account Config Sync reads as when secret_type is
            # "gcpserviceaccount": a GcpServiceAccount reference or a literal email.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            gcp_service_account_email = optional(string, "")

            # An HTTPS proxy for reaching the repository.
            https_proxy = optional(string, "")

            # The directory within the repository to sync. Empty is the root.
            policy_dir = optional(string, "")

            # The branch to sync. Empty is Google's default ("master").
            sync_branch = optional(string, "")

            # The repository URL. Required.
            sync_repo = string

            # A tag or commit to check out instead of the branch head.
            sync_rev = optional(string, "")

            # Seconds between syncs. 0 takes Google's default (15).
            sync_wait_secs = optional(number, 0)
          }))

          # Sync from an OCI image in Artifact Registry.
          oci = optional(object({
            # How Config Sync authenticates to the registry:
            #   "none", "gcenode" (the node's service account), "gcpserviceaccount"
            #   (a Google service account through Workload Identity, see
            #   gcp_service_account_email), "k8sserviceaccount".
            # Required; Google matches it case-sensitively.
            secret_type = string

            # The Google service account Config Sync pulls as when secret_type is
            # "gcpserviceaccount": a GcpServiceAccount reference or a literal email.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            gcp_service_account_email = optional(string, "")

            # The directory within the image to sync. Empty is the image root.
            policy_dir = optional(string, "")

            # The image repository, e.g.
            # "us-docker.pkg.dev/my-project/configs/platform". Required.
            sync_repo = string

            # Seconds between syncs. 0 takes Google's default (15).
            sync_wait_secs = optional(number, 0)
          }))

          # The service account Config Sync exports metrics to Cloud Monitoring
          # as (it needs roles/monitoring.metricWriter): a GcpServiceAccount
          # reference or a literal email.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          metrics_gcp_service_account_email = optional(string, "")

          # true turns on Config Sync's admission webhook, which refuses manual
          # changes to synced objects; false turns it off. Unset takes Google's
          # default.
          prevent_drift = optional(bool)

          # The repository's layout: "unstructured" (any layout; the recommended
          # choice) or "hierarchy" (the namespace-directory layout). Empty takes
          # Google's default.
          source_format = optional(string, "")
        }))
      }))

      # Cloud Service Mesh for every cluster (feature "servicemesh").
      mesh = optional(object({
        # "MANAGEMENT_AUTOMATIC" lets Google provision and upgrade the managed
        # mesh; "MANAGEMENT_MANUAL" leaves mesh components to you. Required: it
        # is the one mesh setting Google offers here.
        management = string
      }))

      # Policy Controller for every cluster (feature "policycontroller").
      policycontroller = optional(object({
        # The Policy Controller version to install. Empty takes the current
        # release.
        version = optional(string, "")

        # How Policy Controller is installed and what it enforces. Required.
        policy_controller_hub_config = object({
          # The installation state:
          #   "INSTALL_SPEC_ENABLED"       -- installed and enforcing
          #   "INSTALL_SPEC_SUSPENDED"     -- installed with its webhooks off
          #   "INSTALL_SPEC_NOT_INSTALLED" -- uninstalled
          #   "INSTALL_SPEC_DETACHED"      -- Google stops reconciling (break-glass)
          # Required.
          install_spec = string

          # Seconds between audit scans; 0 turns auditing off. Unset takes
          # Google's default.
          audit_interval_seconds = optional(number)

          # How many violations a constraint stores. Unset takes Google's default
          # (20).
          constraint_violation_limit = optional(number)

          # Namespaces Policy Controller never checks; they need not exist yet.
          exemptable_namespaces = optional(list(string), [])

          # Log every denial and dry-run failure.
          log_denies_enabled = optional(bool, false)

          # Allow mutation policies, which change objects at admission.
          mutation_enabled = optional(bool, false)

          # Allow constraint templates that read objects other than the one being
          # checked (referential rules).
          referential_rules_enabled = optional(bool, false)

          # Where Policy Controller exports metrics. Unset takes Google's default;
          # set with no backends to turn metrics export off.
          monitoring = optional(object({
            # "PROMETHEUS" and/or "CLOUD_MONITORING". Empty turns export off.
            backends = optional(list(string), [])
          }))

          # Sizing and placement for Policy Controller's components.
          deployment_configs = optional(list(object({
            # The component: "admission", "audit", or "mutation". Required.
            component = string

            # Pod replicas. Unset takes Google's default.
            replica_count = optional(number)

            # "ANTI_AFFINITY" spreads replicas across nodes; "NO_AFFINITY" does not.
            # Empty takes Google's default.
            pod_affinity = optional(string, "")

            # Container resources.
            container_resources = optional(object({
              # The most the container may use.
              limits = optional(object({
                # CPU.
                cpu = optional(string, "")

                # Memory.
                memory = optional(string, "")
              }))

              # What the scheduler reserves for the container.
              requests = optional(object({
                # CPU.
                cpu = optional(string, "")

                # Memory.
                memory = optional(string, "")
              }))
            }))

            # Tolerations, so the component can run on tainted nodes.
            pod_tolerations = optional(list(object({
              # The taint effect matched, e.g. "NoSchedule".
              effect = optional(string, "")

              # The taint key matched.
              key = optional(string, "")

              # "Equal" or "Exists".
              operator = optional(string, "")

              # The taint value matched (with "Equal").
              value = optional(string, "")
            })), [])
          })), [])

          # The policies installed.
          policy_content = optional(object({
            # Google's policy bundles to install, e.g. "cis-k8s-v1.5.1",
            # "pss-baseline-v2022", "policy-essentials-v2022".
            bundles = optional(list(object({
              # The bundle's name. Required.
              bundle = string

              # Namespaces the bundle's constraints skip.
              exempted_namespaces = optional(list(string), [])
            })), [])

            # The constraint template library.
            template_library = optional(object({
              # "ALL" installs Google's whole template library; "NOT_INSTALLED"
              # installs none. Empty takes Google's default.
              installation = optional(string, "")
            }))
          }))
        })
      }))
    }))

    # Per-cluster settings for "configmanagement", "servicemesh", or
    # "policycontroller", each overriding the fleet-wide default for one
    # cluster.
    membership_configs = optional(list(object({
      # The cluster's fleet membership, by full name: a GcpGkeFleetMembership
      # reference, a GcpGkeCluster reference (its fleet_membership output), or
      # the literal name. The membership must be in this feature's fleet
      # project; both modules derive the membership ID and location from it.
      # Immutable.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      membership = string

      # Config Sync for this cluster (feature "configmanagement").
      configmanagement = optional(object({
        # Upgrades: "MANAGEMENT_AUTOMATIC" or "MANAGEMENT_MANUAL" (pinned to
        # `version`). Empty takes Google's default.
        management = optional(string, "")

        # The Config Sync version to install on this cluster.
        version = optional(string, "")

        # What this cluster syncs and how.
        config_sync = optional(object({
          # true installs Config Sync and applies these settings; false removes it
          # and ignores the rest. Unset installs it exactly when git or oci is set.
          enabled = optional(bool)

          # Sync from a Git repository.
          git = optional(object({
            # How Config Sync authenticates to the repository:
            #   "none" (public repository), "ssh", "cookiefile", "token",
            #   "gcenode" (the node's service account), "gcpserviceaccount" (a
            #   Google service account through Workload Identity, see
            #   gcp_service_account_email), "githubapp".
            # Required; Google matches it case-sensitively.
            secret_type = string

            # The Google service account Config Sync reads as when secret_type is
            # "gcpserviceaccount": a GcpServiceAccount reference or a literal email.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            gcp_service_account_email = optional(string, "")

            # An HTTPS proxy for reaching the repository.
            https_proxy = optional(string, "")

            # The directory within the repository to sync. Empty is the root.
            policy_dir = optional(string, "")

            # The branch to sync. Empty is Google's default ("master").
            sync_branch = optional(string, "")

            # The repository URL. Required.
            sync_repo = string

            # A tag or commit to check out instead of the branch head.
            sync_rev = optional(string, "")

            # Seconds between syncs. 0 takes Google's default (15).
            sync_wait_secs = optional(number, 0)
          }))

          # Sync from an OCI image in Artifact Registry.
          oci = optional(object({
            # How Config Sync authenticates to the registry:
            #   "none", "gcenode" (the node's service account), "gcpserviceaccount"
            #   (a Google service account through Workload Identity, see
            #   gcp_service_account_email), "k8sserviceaccount".
            # Required; Google matches it case-sensitively.
            secret_type = string

            # The Google service account Config Sync pulls as when secret_type is
            # "gcpserviceaccount": a GcpServiceAccount reference or a literal email.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            gcp_service_account_email = optional(string, "")

            # The directory within the image to sync. Empty is the image root.
            policy_dir = optional(string, "")

            # The image repository, e.g.
            # "us-docker.pkg.dev/my-project/configs/platform". Required.
            sync_repo = string

            # Seconds between syncs. 0 takes Google's default (15).
            sync_wait_secs = optional(number, 0)
          }))

          # The service account Config Sync exports metrics as: a
          # GcpServiceAccount reference or a literal email.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          metrics_gcp_service_account_email = optional(string, "")

          # true turns on the admission webhook that refuses manual changes to
          # synced objects. Unset takes Google's default.
          prevent_drift = optional(bool)

          # The repository's layout: "unstructured" or "hierarchy".
          source_format = optional(string, "")

          # true pauses syncing on this cluster (a break-glass lever during an
          # incident); Config Sync stays installed. Unset or false keeps syncing.
          stop_syncing = optional(bool)

          # Resource overrides for Config Sync's own Deployments on this cluster,
          # e.g. more memory for the reconciler of a large repository.
          deployment_overrides = optional(list(object({
            # The Deployment's name, e.g. "root-reconciler".
            deployment_name = optional(string, "")

            # The Deployment's namespace, e.g. "config-management-system".
            deployment_namespace = optional(string, "")

            # Per-container resource overrides.
            containers = optional(list(object({
              # The container's name, e.g. "reconciler".
              container_name = optional(string, "")

              # CPU limit.
              cpu_limit = optional(string, "")

              # CPU request.
              cpu_request = optional(string, "")

              # Memory limit.
              memory_limit = optional(string, "")

              # Memory request.
              memory_request = optional(string, "")
            })), [])
          })), [])
        }))
      }))

      # Cloud Service Mesh for this cluster (feature "servicemesh").
      mesh = optional(object({
        # "MANAGEMENT_AUTOMATIC" lets Google provision and upgrade the managed
        # mesh; "MANAGEMENT_MANUAL" leaves mesh components to you. Required: it
        # is the one mesh setting Google offers here.
        management = string
      }))

      # Policy Controller for this cluster (feature "policycontroller").
      policycontroller = optional(object({
        # The Policy Controller version to install. Empty takes the current
        # release.
        version = optional(string, "")

        # How Policy Controller is installed and what it enforces. Required.
        policy_controller_hub_config = object({
          # The installation state:
          #   "INSTALL_SPEC_ENABLED"       -- installed and enforcing
          #   "INSTALL_SPEC_SUSPENDED"     -- installed with its webhooks off
          #   "INSTALL_SPEC_NOT_INSTALLED" -- uninstalled
          #   "INSTALL_SPEC_DETACHED"      -- Google stops reconciling (break-glass)
          # Required.
          install_spec = string

          # Seconds between audit scans; 0 turns auditing off. Unset takes
          # Google's default.
          audit_interval_seconds = optional(number)

          # How many violations a constraint stores. Unset takes Google's default
          # (20).
          constraint_violation_limit = optional(number)

          # Namespaces Policy Controller never checks; they need not exist yet.
          exemptable_namespaces = optional(list(string), [])

          # Log every denial and dry-run failure.
          log_denies_enabled = optional(bool, false)

          # Allow mutation policies, which change objects at admission.
          mutation_enabled = optional(bool, false)

          # Allow constraint templates that read objects other than the one being
          # checked (referential rules).
          referential_rules_enabled = optional(bool, false)

          # Where Policy Controller exports metrics. Unset takes Google's default;
          # set with no backends to turn metrics export off.
          monitoring = optional(object({
            # "PROMETHEUS" and/or "CLOUD_MONITORING". Empty turns export off.
            backends = optional(list(string), [])
          }))

          # Sizing and placement for Policy Controller's components.
          deployment_configs = optional(list(object({
            # The component: "admission", "audit", or "mutation". Required.
            component = string

            # Pod replicas. Unset takes Google's default.
            replica_count = optional(number)

            # "ANTI_AFFINITY" spreads replicas across nodes; "NO_AFFINITY" does not.
            # Empty takes Google's default.
            pod_affinity = optional(string, "")

            # Container resources.
            container_resources = optional(object({
              # The most the container may use.
              limits = optional(object({
                # CPU.
                cpu = optional(string, "")

                # Memory.
                memory = optional(string, "")
              }))

              # What the scheduler reserves for the container.
              requests = optional(object({
                # CPU.
                cpu = optional(string, "")

                # Memory.
                memory = optional(string, "")
              }))
            }))

            # Tolerations, so the component can run on tainted nodes.
            pod_tolerations = optional(list(object({
              # The taint effect matched, e.g. "NoSchedule".
              effect = optional(string, "")

              # The taint key matched.
              key = optional(string, "")

              # "Equal" or "Exists".
              operator = optional(string, "")

              # The taint value matched (with "Equal").
              value = optional(string, "")
            })), [])
          })), [])

          # The policies installed.
          policy_content = optional(object({
            # Google's policy bundles to install, e.g. "cis-k8s-v1.5.1",
            # "pss-baseline-v2022", "policy-essentials-v2022".
            bundles = optional(list(object({
              # The bundle's name. Required.
              bundle = string

              # Namespaces the bundle's constraints skip.
              exempted_namespaces = optional(list(string), [])
            })), [])

            # The constraint template library.
            template_library = optional(object({
              # "ALL" installs Google's whole template library; "NOT_INSTALLED"
              # installs none. Empty takes Google's default.
              installation = optional(string, "")
            }))
          }))
        })
      }))
    })), [])

    # What destroy does, for the feature and its per-cluster settings:
    #   "" / "DELETE" -- the feature is turned off and the per-cluster
    #                    settings removed ("rbacrolebindingactuation" only
    #                    empties its allowlist)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- everything leaves management and stays on in Google
    deletion_policy = optional(string, "")
  })
}
