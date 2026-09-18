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
  description = "GcpDataprocCluster specification"
  type = object({
    # GCP project where the Dataproc cluster will be created.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # GCP region for the cluster (e.g., "us-central1", "europe-west12").
    # All cluster nodes will be placed in this region.
    # Immutable after creation.
    region = string

    # Name of the Dataproc cluster. Must start with a lowercase letter,
    # can contain lowercase letters, numbers, and hyphens, and must end
    # with a lowercase letter or number. Maximum 55 characters.
    # Immutable after creation.
    cluster_name = string

    # Standard (Compute Engine) cluster arm: nodes, software, networking,
    # encryption, security, and lifecycle. Mutually exclusive with
    # virtual_cluster_config.
    cluster_config = optional(object({
      # Cloud Storage bucket for staging job dependencies, jar files, and
      # other temporary data. If not specified, GCP auto-creates a staging
      # bucket in the cluster's project and region.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      staging_bucket = optional(string, "")

      # Cloud Storage bucket for ephemeral cluster data (shuffle, spill).
      # If not specified, GCP auto-creates a temp bucket.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      temp_bucket = optional(string, "")

      # Cluster tier controlling the Dataproc feature set and SLA.
      # CLUSTER_TIER_STANDARD: the default tier.
      # CLUSTER_TIER_PREMIUM: premium features (e.g. advanced repair and
      # faster scaling). Immutable after creation.
      cluster_tier = optional(string, "")

      # Compute Engine configuration for the cluster's nodes (networking,
      # service account, zone, tags, metadata, hardening, placement).
      gce_config = optional(object({
        # VPC network for the cluster's nodes. Mutually exclusive with subnetwork.
        # Expects a network self-link or resource reference.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        network = optional(string, "")

        # VPC subnetwork for the cluster's nodes. Mutually exclusive with network.
        # Using a subnetwork is recommended for production clusters with
        # controlled IP ranges.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        subnetwork = optional(string, "")

        # Service account for the cluster's VMs. If not specified, the default
        # Compute Engine service account is used. A custom service account with
        # minimal permissions is recommended for production.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        service_account = optional(string, "")

        # OAuth 2.0 scopes for the service account. If not specified, GCP uses
        # a default set of scopes. Override only when you need to restrict or
        # expand API access.
        service_account_scopes = optional(list(string), [])

        # GCP zone within the region for node placement. If not specified,
        # GCP auto-selects a zone within the cluster's region.
        zone = optional(string, "")

        # Whether to use only internal IP addresses for cluster nodes.
        # When true, nodes have no external IP and require Cloud NAT or
        # Private Google Access for internet connectivity.
        # Recommended for production to reduce attack surface.
        internal_ip_only = optional(bool, false)

        # GCE network tags applied to all cluster nodes. Useful for
        # firewall rule targeting.
        tags = optional(list(string), [])

        # Compute Engine metadata key-value pairs applied to all cluster nodes.
        # Common use: startup scripts, environment variables for init actions.
        metadata = optional(map(string), {})

        # Shielded VM hardening (secure boot, vTPM, integrity monitoring)
        # for all cluster nodes.
        shielded_instance_config = optional(object({
          # Verify the boot integrity of the VM using a secure boot chain.
          enable_secure_boot = optional(bool, false)

          # Enable the virtual Trusted Platform Module (vTPM).
          enable_vtpm = optional(bool, false)

          # Enable integrity monitoring comparing boot measurements against a
          # known-good baseline.
          enable_integrity_monitoring = optional(bool, false)
        }))

        # Reservation affinity for guaranteed compute capacity.
        reservation_affinity = optional(object({
          # How the cluster consumes reservations.
          # NO_RESERVATION: never consume reservations.
          # ANY_RESERVATION: consume any matching open reservation (default).
          # SPECIFIC_RESERVATION: consume only the reservation named by key/values.
          consume_reservation_type = optional(string, "")

          # Reservation label key. For SPECIFIC_RESERVATION use
          # "compute.googleapis.com/reservation-name".
          key = optional(string, "")

          # Reservation label values (the reservation name for
          # SPECIFIC_RESERVATION).
          values = optional(list(string), [])
        }))

        # Sole-tenant node group placement for compliance or licensing
        # isolation.
        node_group_affinity = optional(object({
          # URI of the sole-tenant node group the cluster will be created on.
          # Format: projects/{project}/zones/{zone}/nodeGroups/{node_group}
          # (shorter forms accepted by the API also work).
          node_group_uri = string
        }))

        # Confidential VM (data-in-use encryption) for all cluster nodes.
        # Requires an N2D machine type.
        confidential_instance_config = optional(object({
          # Enable Confidential Compute for all cluster nodes with the AMD SEV
          # technology. The provider marks this boolean deprecated in favor of
          # confidential_instance_type, which names the technology explicitly;
          # prefer that field for new clusters. Both may be set for a cluster
          # authored before the type field existed.
          enable_confidential_compute = optional(bool, false)

          # Confidential Compute technology for all cluster nodes:
          #   "SEV"     -- AMD Secure Encrypted Virtualization (N2D, C2D, C3D)
          #   "SEV_SNP" -- AMD SEV with Secure Nested Paging: adds integrity
          #                protection and attestation (N2D)
          #   "TDX"     -- Intel Trust Domain Extensions (C3)
          # The machine type must support the chosen technology. Immutable.
          confidential_instance_type = optional(string, "")
        }))

        # Resource manager (secure) tags applied to all cluster instances,
        # keyed "tagKeys/{tag_key_id}" with values "tagValues/{tag_value_id}".
        # Unlike network tags, these are IAM-governed and usable in org
        # policies and firewall rules.
        resource_manager_tags = optional(map(string), {})
      }))

      # Master node configuration. If not specified, GCP defaults to
      # 1 master with a default machine type and 500 GB pd-standard disk.
      master_config = optional(object({
        # Number of master instances. Valid values: 1 (standard) or 3 (HA).
        # If not specified, GCP defaults to 1.
        num_instances = optional(number, 0)

        # Compute Engine machine type (e.g., "n2-standard-4", "e2-standard-8").
        # If not specified, GCP selects a default machine type.
        # Mutually exclusive with instance_flexibility_policy — a flexibility
        # policy replaces the single machine type (see that field).
        machine_type = optional(string, "")

        # Boot disk and local SSD configuration.
        disk_config = optional(object({
          # Size of the boot disk in GB. Minimum 10 GB.
          # If not specified, GCP defaults to 500 GB for master and worker nodes.
          boot_disk_size_gb = optional(number, 0)

          # Boot disk type: "pd-standard" (GCP default), "pd-ssd", "pd-balanced",
          # or "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
          # dials apply). The API validates availability per image version and
          # machine family at deploy time.
          boot_disk_type = optional(string, "")

          # Number of local SSDs to attach. Each local SSD is 375 GB.
          # Default: 0 (no local SSDs).
          num_local_ssds = optional(number, 0)

          # Interface used to attach local SSDs: "scsi" (default) or "nvme".
          # NVMe offers higher throughput for shuffle-heavy Spark workloads but
          # requires an image that ships NVMe drivers (all current Dataproc
          # images do).
          local_ssd_interface = optional(string, "")

          # Provisioned I/O operations per second for the boot disk — the IOPS
          # dial decoupled from disk size, honored by disk types that support
          # provisioned performance (hyperdisks).
          boot_disk_provisioned_iops = optional(number)

          # Provisioned throughput in MB/s for the boot disk — the bandwidth
          # dial decoupled from disk size, honored by disk types that support
          # provisioned performance (hyperdisks).
          boot_disk_provisioned_throughput = optional(number)

          # Additional persistent disks attached to every node of this role,
          # beyond the boot disk — for HDFS data or shuffle spill that should not
          # share the boot volume, or for Hyperdisk performance tiers the boot
          # disk cannot use. Each entry is one disk on each node. Immutable: the
          # whole set is fixed at cluster creation (ForceNew). Not available on
          # auxiliary (driver) node groups.
          attached_disks = optional(list(object({
            # Size of the disk in GB. Leave unset for the API's default size.
            disk_size_gb = optional(number, 0)

            # Disk type: "pd-standard", "pd-ssd", "pd-balanced", or
            # "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
            # dials apply). Leave empty for the API's default.
            disk_type = optional(string, "")

            # Provisioned I/O operations per second, honored by disk types that
            # support provisioned performance (hyperdisks).
            provisioned_iops = optional(number)

            # Provisioned throughput in MB/s, honored by disk types that support
            # provisioned performance (hyperdisks).
            provisioned_throughput = optional(number)
          })), [])
        }))

        # GPU accelerators attached to master nodes.
        # Typically not needed for masters unless running single-node ML workloads.
        accelerators = optional(list(object({
          # Full accelerator type name (e.g., "nvidia-tesla-t4", "nvidia-tesla-v100").
          # See https://cloud.google.com/compute/docs/gpus for available types.
          accelerator_type = string

          # Number of accelerators to attach. Must be at least 1.
          accelerator_count = optional(number, 0)
        })), [])

        # Minimum CPU platform for the instances (e.g., "Intel Cascade Lake").
        # Forces nodes onto a specific or newer CPU generation.
        min_cpu_platform = optional(string, "")

        # Custom Dataproc image URI. If not specified, GCP uses the image
        # determined by software_config.image_version.
        image_uri = optional(string, "")

        # Ranked machine-type preferences for master provisioning — keeps the
        # cluster creatable when the preferred type's zonal capacity dries up.
        # REPLACES machine_type: with a flexibility policy present the API
        # provisions solely from the ranked selections and drops a paired
        # machineTypeUri from the stored config. Masters are on-demand
        # capacity: provisioning_model_mix does not apply here (secondary
        # workers only).
        instance_flexibility_policy = optional(object({
          # Ranked machine-type preferences. Dataproc provisions from the
          # lowest-rank entry with available capacity.
          instance_selection_list = optional(list(object({
            # Machine types to try for this selection entry (e.g.
            # ["n2-standard-8", "n2d-standard-8"]).
            machine_types = list(string)

            # Preference rank. Lower rank is preferred; Dataproc falls back to
            # higher ranks when capacity for the preferred types is unavailable.
            rank = optional(number, 0)

            # Disk shape for nodes provisioned from THIS selection entry —
            # overrides the role's disk_config so, for example, the fallback
            # machine types can carry a different boot disk or local SSD count
            # than the preferred ones. Leave unset to inherit the role's
            # disk_config. Immutable.
            disk_config = optional(object({
              # Size of the boot disk in GB. Minimum 10 GB.
              # If not specified, GCP defaults to 500 GB for master and worker nodes.
              boot_disk_size_gb = optional(number, 0)

              # Boot disk type: "pd-standard" (GCP default), "pd-ssd", "pd-balanced",
              # or "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
              # dials apply). The API validates availability per image version and
              # machine family at deploy time.
              boot_disk_type = optional(string, "")

              # Number of local SSDs to attach. Each local SSD is 375 GB.
              # Default: 0 (no local SSDs).
              num_local_ssds = optional(number, 0)

              # Interface used to attach local SSDs: "scsi" (default) or "nvme".
              # NVMe offers higher throughput for shuffle-heavy Spark workloads but
              # requires an image that ships NVMe drivers (all current Dataproc
              # images do).
              local_ssd_interface = optional(string, "")

              # Provisioned I/O operations per second for the boot disk — the IOPS
              # dial decoupled from disk size, honored by disk types that support
              # provisioned performance (hyperdisks).
              boot_disk_provisioned_iops = optional(number)

              # Provisioned throughput in MB/s for the boot disk — the bandwidth
              # dial decoupled from disk size, honored by disk types that support
              # provisioned performance (hyperdisks).
              boot_disk_provisioned_throughput = optional(number)

              # Additional persistent disks attached to every node of this role,
              # beyond the boot disk — for HDFS data or shuffle spill that should not
              # share the boot volume, or for Hyperdisk performance tiers the boot
              # disk cannot use. Each entry is one disk on each node. Immutable: the
              # whole set is fixed at cluster creation (ForceNew). Not available on
              # auxiliary (driver) node groups.
              attached_disks = optional(list(object({
                # Size of the disk in GB. Leave unset for the API's default size.
                disk_size_gb = optional(number, 0)

                # Disk type: "pd-standard", "pd-ssd", "pd-balanced", or
                # "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
                # dials apply). Leave empty for the API's default.
                disk_type = optional(string, "")

                # Provisioned I/O operations per second, honored by disk types that
                # support provisioned performance (hyperdisks).
                provisioned_iops = optional(number)

                # Provisioned throughput in MB/s, honored by disk types that support
                # provisioned performance (hyperdisks).
                provisioned_throughput = optional(number)
              })), [])
            }))
          })), [])

          # Standard/spot capacity mix for the group.
          provisioning_model_mix = optional(object({
            # Number of secondary workers that must be standard (on-demand)
            # capacity before any spot VMs are used.
            standard_capacity_base = optional(number, 0)

            # Percentage of capacity above the base that should also be standard
            # (0-100). The remainder is provisioned as spot.
            standard_capacity_percent_above_base = optional(number, 0)
          }))
        }))
      }))

      # Primary worker node configuration. If not specified, GCP defaults
      # to 2 workers. For a single-node cluster set the software property
      # "dataproc:dataproc.allow.zero.workers" = "true" instead of a
      # worker count of zero.
      worker_config = optional(object({
        # Number of primary worker instances.
        # If not specified, GCP defaults to 2. This is one of the few fields
        # that can be changed in place after creation (manual scaling).
        num_instances = optional(number, 0)

        # Compute Engine machine type (e.g., "n2-standard-4", "e2-standard-8").
        # If not specified, GCP selects a default machine type.
        # Mutually exclusive with instance_flexibility_policy — a flexibility
        # policy replaces the single machine type (see that field).
        machine_type = optional(string, "")

        # Boot disk and local SSD configuration.
        disk_config = optional(object({
          # Size of the boot disk in GB. Minimum 10 GB.
          # If not specified, GCP defaults to 500 GB for master and worker nodes.
          boot_disk_size_gb = optional(number, 0)

          # Boot disk type: "pd-standard" (GCP default), "pd-ssd", "pd-balanced",
          # or "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
          # dials apply). The API validates availability per image version and
          # machine family at deploy time.
          boot_disk_type = optional(string, "")

          # Number of local SSDs to attach. Each local SSD is 375 GB.
          # Default: 0 (no local SSDs).
          num_local_ssds = optional(number, 0)

          # Interface used to attach local SSDs: "scsi" (default) or "nvme".
          # NVMe offers higher throughput for shuffle-heavy Spark workloads but
          # requires an image that ships NVMe drivers (all current Dataproc
          # images do).
          local_ssd_interface = optional(string, "")

          # Provisioned I/O operations per second for the boot disk — the IOPS
          # dial decoupled from disk size, honored by disk types that support
          # provisioned performance (hyperdisks).
          boot_disk_provisioned_iops = optional(number)

          # Provisioned throughput in MB/s for the boot disk — the bandwidth
          # dial decoupled from disk size, honored by disk types that support
          # provisioned performance (hyperdisks).
          boot_disk_provisioned_throughput = optional(number)

          # Additional persistent disks attached to every node of this role,
          # beyond the boot disk — for HDFS data or shuffle spill that should not
          # share the boot volume, or for Hyperdisk performance tiers the boot
          # disk cannot use. Each entry is one disk on each node. Immutable: the
          # whole set is fixed at cluster creation (ForceNew). Not available on
          # auxiliary (driver) node groups.
          attached_disks = optional(list(object({
            # Size of the disk in GB. Leave unset for the API's default size.
            disk_size_gb = optional(number, 0)

            # Disk type: "pd-standard", "pd-ssd", "pd-balanced", or
            # "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
            # dials apply). Leave empty for the API's default.
            disk_type = optional(string, "")

            # Provisioned I/O operations per second, honored by disk types that
            # support provisioned performance (hyperdisks).
            provisioned_iops = optional(number)

            # Provisioned throughput in MB/s, honored by disk types that support
            # provisioned performance (hyperdisks).
            provisioned_throughput = optional(number)
          })), [])
        }))

        # GPU accelerators attached to worker nodes.
        # Common for Spark ML workloads using GPU-accelerated libraries.
        accelerators = optional(list(object({
          # Full accelerator type name (e.g., "nvidia-tesla-t4", "nvidia-tesla-v100").
          # See https://cloud.google.com/compute/docs/gpus for available types.
          accelerator_type = string

          # Number of accelerators to attach. Must be at least 1.
          accelerator_count = optional(number, 0)
        })), [])

        # Minimum CPU platform for the instances.
        min_cpu_platform = optional(string, "")

        # Custom Dataproc image URI.
        image_uri = optional(string, "")

        # Minimum number of primary worker instances when autoscaling is active.
        # The autoscaler will not scale below this threshold. Updatable in place.
        min_num_instances = optional(number, 0)

        # Ranked machine-type preferences for primary-worker provisioning —
        # keeps scale-ups schedulable when the preferred type's zonal capacity
        # dries up. REPLACES machine_type: with a flexibility policy present
        # the API provisions solely from the ranked selections and drops a
        # paired machineTypeUri from the stored config. Primary workers are
        # on-demand capacity: provisioning_model_mix does not apply here
        # (secondary workers only).
        instance_flexibility_policy = optional(object({
          # Ranked machine-type preferences. Dataproc provisions from the
          # lowest-rank entry with available capacity.
          instance_selection_list = optional(list(object({
            # Machine types to try for this selection entry (e.g.
            # ["n2-standard-8", "n2d-standard-8"]).
            machine_types = list(string)

            # Preference rank. Lower rank is preferred; Dataproc falls back to
            # higher ranks when capacity for the preferred types is unavailable.
            rank = optional(number, 0)

            # Disk shape for nodes provisioned from THIS selection entry —
            # overrides the role's disk_config so, for example, the fallback
            # machine types can carry a different boot disk or local SSD count
            # than the preferred ones. Leave unset to inherit the role's
            # disk_config. Immutable.
            disk_config = optional(object({
              # Size of the boot disk in GB. Minimum 10 GB.
              # If not specified, GCP defaults to 500 GB for master and worker nodes.
              boot_disk_size_gb = optional(number, 0)

              # Boot disk type: "pd-standard" (GCP default), "pd-ssd", "pd-balanced",
              # or "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
              # dials apply). The API validates availability per image version and
              # machine family at deploy time.
              boot_disk_type = optional(string, "")

              # Number of local SSDs to attach. Each local SSD is 375 GB.
              # Default: 0 (no local SSDs).
              num_local_ssds = optional(number, 0)

              # Interface used to attach local SSDs: "scsi" (default) or "nvme".
              # NVMe offers higher throughput for shuffle-heavy Spark workloads but
              # requires an image that ships NVMe drivers (all current Dataproc
              # images do).
              local_ssd_interface = optional(string, "")

              # Provisioned I/O operations per second for the boot disk — the IOPS
              # dial decoupled from disk size, honored by disk types that support
              # provisioned performance (hyperdisks).
              boot_disk_provisioned_iops = optional(number)

              # Provisioned throughput in MB/s for the boot disk — the bandwidth
              # dial decoupled from disk size, honored by disk types that support
              # provisioned performance (hyperdisks).
              boot_disk_provisioned_throughput = optional(number)

              # Additional persistent disks attached to every node of this role,
              # beyond the boot disk — for HDFS data or shuffle spill that should not
              # share the boot volume, or for Hyperdisk performance tiers the boot
              # disk cannot use. Each entry is one disk on each node. Immutable: the
              # whole set is fixed at cluster creation (ForceNew). Not available on
              # auxiliary (driver) node groups.
              attached_disks = optional(list(object({
                # Size of the disk in GB. Leave unset for the API's default size.
                disk_size_gb = optional(number, 0)

                # Disk type: "pd-standard", "pd-ssd", "pd-balanced", or
                # "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
                # dials apply). Leave empty for the API's default.
                disk_type = optional(string, "")

                # Provisioned I/O operations per second, honored by disk types that
                # support provisioned performance (hyperdisks).
                provisioned_iops = optional(number)

                # Provisioned throughput in MB/s, honored by disk types that support
                # provisioned performance (hyperdisks).
                provisioned_throughput = optional(number)
              })), [])
            }))
          })), [])

          # Standard/spot capacity mix for the group.
          provisioning_model_mix = optional(object({
            # Number of secondary workers that must be standard (on-demand)
            # capacity before any spot VMs are used.
            standard_capacity_base = optional(number, 0)

            # Percentage of capacity above the base that should also be standard
            # (0-100). The remainder is provisioned as spot.
            standard_capacity_percent_above_base = optional(number, 0)
          }))
        }))
      }))

      # Secondary (preemptible/spot) worker node configuration.
      # If not specified, no secondary workers are created.
      secondary_worker_config = optional(object({
        # Number of secondary worker instances. Default: 0.
        # Updatable in place (manual scaling).
        num_instances = optional(number, 0)

        # Preemptibility of the secondary workers.
        # SPOT: modern spot VMs with dynamic pricing (recommended).
        # PREEMPTIBLE: legacy preemptible VMs with fixed pricing.
        # NON_PREEMPTIBLE: standard on-demand pricing (unusual for secondary workers).
        # If not specified, GCP defaults to PREEMPTIBLE.
        # Immutable after creation.
        preemptibility = optional(string, "")

        # Boot disk and local SSD configuration for secondary workers.
        # Secondary workers do not support custom machine types or accelerators
        # directly; they inherit machine configuration from the primary worker
        # config unless an instance_flexibility_policy overrides it.
        disk_config = optional(object({
          # Size of the boot disk in GB. Minimum 10 GB.
          # If not specified, GCP defaults to 500 GB for master and worker nodes.
          boot_disk_size_gb = optional(number, 0)

          # Boot disk type: "pd-standard" (GCP default), "pd-ssd", "pd-balanced",
          # or "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
          # dials apply). The API validates availability per image version and
          # machine family at deploy time.
          boot_disk_type = optional(string, "")

          # Number of local SSDs to attach. Each local SSD is 375 GB.
          # Default: 0 (no local SSDs).
          num_local_ssds = optional(number, 0)

          # Interface used to attach local SSDs: "scsi" (default) or "nvme".
          # NVMe offers higher throughput for shuffle-heavy Spark workloads but
          # requires an image that ships NVMe drivers (all current Dataproc
          # images do).
          local_ssd_interface = optional(string, "")

          # Provisioned I/O operations per second for the boot disk — the IOPS
          # dial decoupled from disk size, honored by disk types that support
          # provisioned performance (hyperdisks).
          boot_disk_provisioned_iops = optional(number)

          # Provisioned throughput in MB/s for the boot disk — the bandwidth
          # dial decoupled from disk size, honored by disk types that support
          # provisioned performance (hyperdisks).
          boot_disk_provisioned_throughput = optional(number)

          # Additional persistent disks attached to every node of this role,
          # beyond the boot disk — for HDFS data or shuffle spill that should not
          # share the boot volume, or for Hyperdisk performance tiers the boot
          # disk cannot use. Each entry is one disk on each node. Immutable: the
          # whole set is fixed at cluster creation (ForceNew). Not available on
          # auxiliary (driver) node groups.
          attached_disks = optional(list(object({
            # Size of the disk in GB. Leave unset for the API's default size.
            disk_size_gb = optional(number, 0)

            # Disk type: "pd-standard", "pd-ssd", "pd-balanced", or
            # "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
            # dials apply). Leave empty for the API's default.
            disk_type = optional(string, "")

            # Provisioned I/O operations per second, honored by disk types that
            # support provisioned performance (hyperdisks).
            provisioned_iops = optional(number)

            # Provisioned throughput in MB/s, honored by disk types that support
            # provisioned performance (hyperdisks).
            provisioned_throughput = optional(number)
          })), [])
        }))

        # Machine-type flexibility and standard/spot capacity mix for the
        # group — only secondary workers support flexible provisioning.
        instance_flexibility_policy = optional(object({
          # Ranked machine-type preferences. Dataproc provisions from the
          # lowest-rank entry with available capacity.
          instance_selection_list = optional(list(object({
            # Machine types to try for this selection entry (e.g.
            # ["n2-standard-8", "n2d-standard-8"]).
            machine_types = list(string)

            # Preference rank. Lower rank is preferred; Dataproc falls back to
            # higher ranks when capacity for the preferred types is unavailable.
            rank = optional(number, 0)

            # Disk shape for nodes provisioned from THIS selection entry —
            # overrides the role's disk_config so, for example, the fallback
            # machine types can carry a different boot disk or local SSD count
            # than the preferred ones. Leave unset to inherit the role's
            # disk_config. Immutable.
            disk_config = optional(object({
              # Size of the boot disk in GB. Minimum 10 GB.
              # If not specified, GCP defaults to 500 GB for master and worker nodes.
              boot_disk_size_gb = optional(number, 0)

              # Boot disk type: "pd-standard" (GCP default), "pd-ssd", "pd-balanced",
              # or "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
              # dials apply). The API validates availability per image version and
              # machine family at deploy time.
              boot_disk_type = optional(string, "")

              # Number of local SSDs to attach. Each local SSD is 375 GB.
              # Default: 0 (no local SSDs).
              num_local_ssds = optional(number, 0)

              # Interface used to attach local SSDs: "scsi" (default) or "nvme".
              # NVMe offers higher throughput for shuffle-heavy Spark workloads but
              # requires an image that ships NVMe drivers (all current Dataproc
              # images do).
              local_ssd_interface = optional(string, "")

              # Provisioned I/O operations per second for the boot disk — the IOPS
              # dial decoupled from disk size, honored by disk types that support
              # provisioned performance (hyperdisks).
              boot_disk_provisioned_iops = optional(number)

              # Provisioned throughput in MB/s for the boot disk — the bandwidth
              # dial decoupled from disk size, honored by disk types that support
              # provisioned performance (hyperdisks).
              boot_disk_provisioned_throughput = optional(number)

              # Additional persistent disks attached to every node of this role,
              # beyond the boot disk — for HDFS data or shuffle spill that should not
              # share the boot volume, or for Hyperdisk performance tiers the boot
              # disk cannot use. Each entry is one disk on each node. Immutable: the
              # whole set is fixed at cluster creation (ForceNew). Not available on
              # auxiliary (driver) node groups.
              attached_disks = optional(list(object({
                # Size of the disk in GB. Leave unset for the API's default size.
                disk_size_gb = optional(number, 0)

                # Disk type: "pd-standard", "pd-ssd", "pd-balanced", or
                # "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
                # dials apply). Leave empty for the API's default.
                disk_type = optional(string, "")

                # Provisioned I/O operations per second, honored by disk types that
                # support provisioned performance (hyperdisks).
                provisioned_iops = optional(number)

                # Provisioned throughput in MB/s, honored by disk types that support
                # provisioned performance (hyperdisks).
                provisioned_throughput = optional(number)
              })), [])
            }))
          })), [])

          # Standard/spot capacity mix for the group.
          provisioning_model_mix = optional(object({
            # Number of secondary workers that must be standard (on-demand)
            # capacity before any spot VMs are used.
            standard_capacity_base = optional(number, 0)

            # Percentage of capacity above the base that should also be standard
            # (0-100). The remainder is provisioned as spot.
            standard_capacity_percent_above_base = optional(number, 0)
          }))
        }))
      }))

      # Software configuration including Dataproc image version, optional
      # components, and framework property overrides.
      software_config = optional(object({
        # Dataproc image version (e.g., "2.2-debian12", "2.1-ubuntu20").
        # See https://cloud.google.com/dataproc/docs/concepts/versioning/dataproc-versions
        # If not specified, GCP uses the latest stable version.
        image_version = optional(string, "")

        # Optional components to install on the cluster. Common values:
        # JUPYTER, DOCKER, PRESTO, ZEPPELIN, HIVE_WEBHCAT, FLINK, TRINO.
        # See https://cloud.google.com/dataproc/docs/concepts/components/overview
        optional_components = optional(list(string), [])

        # Key-value pairs to override or set Hadoop, Spark, YARN, and other
        # framework properties. Keys use the format "prefix:property", e.g.:
        #   "spark:spark.executor.memory" = "4g"
        #   "hdfs:dfs.replication" = "2"
        #   "dataproc:dataproc.allow.zero.workers" = "true"  (single-node cluster)
        # See https://cloud.google.com/dataproc/docs/concepts/configuring-clusters/cluster-properties
        properties = optional(map(string), {})
      }))

      # Initialization actions (startup scripts) that run on all nodes
      # when the cluster is created.
      initialization_actions = optional(list(object({
        # GCS URI of the initialization script (must start with "gs://").
        script = string

        # Maximum time (in seconds) the script is allowed to run before
        # being forcefully terminated. Default: 300 (5 minutes).
        timeout_sec = optional(number, 0)
      })), [])

      # Autoscaling policy that governs worker scaling for this cluster.
      # Resolves to the policy's full resource name
      # (projects/{project}/locations/{location}/autoscalingPolicies/{policy}).
      # Attaching, swapping, or detaching the policy updates in place.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      autoscaling_policy_uri = optional(string, "")

      # Cloud KMS key for encrypting persistent disks attached to cluster
      # nodes (CMEK). Format: projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{key}
      # If not specified, disks are encrypted with Google-managed keys.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      encryption_kms_key_name = optional(string, "")

      # In-cluster authentication hardening: Kerberos or personal-cluster
      # identity mapping.
      security_config = optional(object({
        # Kerberos (Hadoop Secure Mode) configuration.
        kerberos_config = optional(object({
          # Flag to indicate whether to Kerberize the cluster.
          enable_kerberos = optional(bool, false)

          # Cloud Storage URI of a KMS-encrypted file containing the root
          # principal password. Required to enable Kerberos.
          root_principal_password_uri = string

          # URI of the KMS key used to decrypt the root password and the other
          # encrypted files below.
          # Format: projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          kms_key_uri = string

          # Custom Kerberos realm. If unset, Dataproc derives a realm from the
          # cluster's hostname.
          realm = optional(string, "")

          # Lifetime of ticket-granting tickets, in hours. Default: 10.
          tgt_lifetime_hours = optional(number, 0)

          # Cloud Storage URI of a KMS-encrypted file containing the master key
          # of the KDC database.
          kdc_db_key_uri = optional(string, "")

          # Cloud Storage URI of the keystore file (TLS). If unset, Dataproc
          # generates a self-signed certificate.
          keystore_uri = optional(string, "")

          # Cloud Storage URI of a KMS-encrypted file with the keystore password.
          keystore_password_uri = optional(string, "")

          # Cloud Storage URI of a KMS-encrypted file with the key password.
          key_password_uri = optional(string, "")

          # Cloud Storage URI of the truststore file (TLS).
          truststore_uri = optional(string, "")

          # Cloud Storage URI of a KMS-encrypted file with the truststore password.
          truststore_password_uri = optional(string, "")

          # Remote realm for cross-realm trust.
          cross_realm_trust_realm = optional(string, "")

          # KDC host of the remote trusted realm.
          cross_realm_trust_kdc = optional(string, "")

          # Admin server host of the remote trusted realm.
          cross_realm_trust_admin_server = optional(string, "")

          # Cloud Storage URI of a KMS-encrypted file with the shared password
          # between the on-cluster KDC and the remote trusted realm.
          cross_realm_trust_shared_password_uri = optional(string, "")
        }))

        # Personal cluster authentication (user-to-service-account mapping).
        identity_config = optional(object({
          # Map of user accounts to the service accounts they operate as
          # (e.g. "bob@example.com" -> "bob-sa@project.iam.gserviceaccount.com").
          user_service_account_mapping = optional(map(string), {})
        }))
      }))

      # Component Gateway configuration for authenticated web UI access.
      endpoint_config = optional(object({
        # Whether to enable the Dataproc Component Gateway for web UI access.
        # When enabled, GCP creates authenticated HTTPS endpoints for each
        # cluster component. Requires the cluster to have external IP access
        # or appropriate Private Google Access configuration.
        enable_http_port_access = optional(bool, false)
      }))

      # Lifecycle configuration for automatic cluster deletion.
      # Critical for cost management of ephemeral batch processing clusters.
      lifecycle_config = optional(object({
        # Duration of inactivity after which the cluster is automatically
        # deleted. Format: duration in seconds with 's' suffix (e.g., "1800s"
        # for 30 minutes). Valid range: 10 minutes to 14 days.
        #
        # A cluster is considered idle when no jobs are running and no
        # interactive sessions are active.
        idle_delete_ttl = optional(string, "")

        # RFC3339 timestamp at which the cluster is automatically deleted,
        # regardless of activity (e.g., "2026-03-01T00:00:00Z").
        # Useful for time-boxed clusters with a known end date.
        auto_delete_time = optional(string, "")

        # Duration of inactivity after which the cluster is automatically
        # STOPPED (not deleted) — VMs shut down, storage and configuration
        # retained, restartable later. Same format and range as
        # idle_delete_ttl (e.g., "1800s"; 10 minutes to 14 days).
        idle_stop_ttl = optional(string, "")

        # RFC3339 timestamp at which the cluster is automatically STOPPED
        # (not deleted), regardless of activity.
        auto_stop_time = optional(string, "")
      }))

      # Attach the cluster to a persistent Dataproc Metastore service.
      metastore_config = optional(object({
        # Resource name of an existing Dataproc Metastore service.
        # Format: projects/{project}/locations/{location}/services/{service}
        # Accepts a literal resource name today; references attach when a
        # metastore-service kind lands in the catalog.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        dataproc_metastore_service = string
      }))

      # OSS metric collection into Cloud Monitoring.
      dataproc_metric_config = optional(object({
        # Metric sources to collect.
        metrics = list(object({
          # Source to collect metrics from.
          metric_source = string

          # Specific metrics to collect, overriding the source's default set
          # (e.g. "yarn:ResourceManager:QueueMetrics:AppsCompleted").
          metric_overrides = optional(list(string), [])
        }))
      }))

      # Dedicated DRIVER node groups, separating Spark driver capacity
      # from the master's control-plane duties.
      auxiliary_node_groups = optional(list(object({
        # Roles assigned to the group. The API currently supports DRIVER.
        roles = list(string)

        # VM sizing for the group.
        node_group_config = optional(object({
          # Number of instances in the group.
          num_instances = optional(number, 0)

          # Compute Engine machine type for the group's instances.
          machine_type = optional(string, "")

          # Minimum CPU platform for the group's instances.
          min_cpu_platform = optional(string, "")

          # Boot disk and local SSD configuration.
          disk_config = optional(object({
            # Size of the boot disk in GB. Minimum 10 GB.
            # If not specified, GCP defaults to 500 GB for master and worker nodes.
            boot_disk_size_gb = optional(number, 0)

            # Boot disk type: "pd-standard" (GCP default), "pd-ssd", "pd-balanced",
            # or "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
            # dials apply). The API validates availability per image version and
            # machine family at deploy time.
            boot_disk_type = optional(string, "")

            # Number of local SSDs to attach. Each local SSD is 375 GB.
            # Default: 0 (no local SSDs).
            num_local_ssds = optional(number, 0)

            # Interface used to attach local SSDs: "scsi" (default) or "nvme".
            # NVMe offers higher throughput for shuffle-heavy Spark workloads but
            # requires an image that ships NVMe drivers (all current Dataproc
            # images do).
            local_ssd_interface = optional(string, "")

            # Provisioned I/O operations per second for the boot disk — the IOPS
            # dial decoupled from disk size, honored by disk types that support
            # provisioned performance (hyperdisks).
            boot_disk_provisioned_iops = optional(number)

            # Provisioned throughput in MB/s for the boot disk — the bandwidth
            # dial decoupled from disk size, honored by disk types that support
            # provisioned performance (hyperdisks).
            boot_disk_provisioned_throughput = optional(number)

            # Additional persistent disks attached to every node of this role,
            # beyond the boot disk — for HDFS data or shuffle spill that should not
            # share the boot volume, or for Hyperdisk performance tiers the boot
            # disk cannot use. Each entry is one disk on each node. Immutable: the
            # whole set is fixed at cluster creation (ForceNew). Not available on
            # auxiliary (driver) node groups.
            attached_disks = optional(list(object({
              # Size of the disk in GB. Leave unset for the API's default size.
              disk_size_gb = optional(number, 0)

              # Disk type: "pd-standard", "pd-ssd", "pd-balanced", or
              # "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
              # dials apply). Leave empty for the API's default.
              disk_type = optional(string, "")

              # Provisioned I/O operations per second, honored by disk types that
              # support provisioned performance (hyperdisks).
              provisioned_iops = optional(number)

              # Provisioned throughput in MB/s, honored by disk types that support
              # provisioned performance (hyperdisks).
              provisioned_throughput = optional(number)
            })), [])
          }))

          # GPU accelerators attached to the group's instances.
          accelerators = optional(list(object({
            # Full accelerator type name (e.g., "nvidia-tesla-t4", "nvidia-tesla-v100").
            # See https://cloud.google.com/compute/docs/gpus for available types.
            accelerator_type = string

            # Number of accelerators to attach. Must be at least 1.
            accelerator_count = optional(number, 0)
          })), [])
        }))

        # Optional stable identifier for the group (3-33 characters).
        # If unset, Dataproc generates one.
        node_group_id = optional(string, "")
      })), [])

      # The cluster's structural type. STANDARD (default) has masters and
      # workers; SINGLE_NODE runs everything on one VM (the modern
      # alternative to the "dataproc:dataproc.allow.zero.workers" property);
      # ZERO_SCALE keeps only the control plane warm and provisions workers
      # on demand. Immutable after creation.
      cluster_type = optional(string, "")

      # The execution engine. DEFAULT runs open-source Spark as-is;
      # LIGHTNING enables the Lightning Engine (Google's accelerated Spark
      # runtime, premium tier). Immutable after creation.
      engine = optional(string, "")
    }))

    # Dataproc-on-GKE arm: run Spark as pods on an existing GKE cluster.
    # Mutually exclusive with cluster_config.
    virtual_cluster_config = optional(object({
      # Cloud Storage bucket for staging job dependencies. If not
      # specified, GCP auto-creates a staging bucket.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      staging_bucket = optional(string, "")

      # The Kubernetes-side configuration: target GKE cluster, node-pool
      # role mapping, namespace, and component versions.
      kubernetes_cluster_config = object({
        # Kubernetes namespace Dataproc deploys workloads into. If unset,
        # Dataproc uses a namespace named after the virtual cluster.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kubernetes_namespace = optional(string, "")

        # The target GKE cluster and node-pool role mapping.
        gke_cluster_config = object({
          # The target GKE cluster. Resolves to the cluster's fully qualified
          # resource name (projects/{project}/locations/{location}/clusters/{cluster}).
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          gke_cluster_target = string

          # Node pools Dataproc schedules roles onto. When omitted, Dataproc
          # creates and manages a default pool on the target cluster.
          node_pool_target = optional(list(object({
            # The target GKE node pool. Resolves to the pool's fully qualified
            # resource name
            # (projects/{project}/locations/{location}/clusters/{cluster}/nodePools/{pool}).
            # The pool may pre-exist (composed via GcpGkeNodePool) or be created
            # by Dataproc using node_pool_config.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            node_pool = string

            # Dataproc roles scheduled onto this pool.
            # DEFAULT: catch-all for roles without a dedicated pool.
            # CONTROLLER: the Dataproc control plane pods.
            # SPARK_DRIVER: Spark driver pods.
            # SPARK_EXECUTOR: Spark executor pods.
            roles = list(string)

            # Sizing for the pool when Dataproc creates it. Ignored for
            # pre-existing pools.
            node_pool_config = optional(object({
              # Compute Engine zones where the pool's nodes are placed
              # (e.g. ["us-central1-a"]). Required when Dataproc creates the pool.
              locations = list(string)

              # GKE autoscaling bounds for the pool.
              autoscaling = optional(object({
                # Minimum number of nodes per zone. May be 0 for scale-to-zero pools.
                min_node_count = optional(number, 0)

                # Maximum number of nodes per zone.
                max_node_count = optional(number, 0)
              }))

              # Compute Engine machine type for the pool's nodes.
              machine_type = optional(string, "")

              # Number of local SSDs attached to each node.
              local_ssd_count = optional(number, 0)

              # Minimum CPU platform for the pool's nodes.
              min_cpu_platform = optional(string, "")

              # Use legacy preemptible VMs for the pool. Mutually exclusive with spot.
              preemptible = optional(bool, false)

              # Use modern spot VMs for the pool (recommended over preemptible).
              spot = optional(bool, false)
            }))
          })), [])
        })

        # Component versions (SPARK is required) and framework properties.
        kubernetes_software_config = object({
          # Component-to-version map. The SPARK component is required, e.g.
          # {"SPARK": "3.5-dataproc-17"}.
          # See https://cloud.google.com/dataproc/docs/guides/dpgke/dataproc-gke-version-compatibility
          component_version = optional(map(string), {})

          # Framework properties (e.g. "spark:spark.eventLog.enabled" = "true").
          properties = optional(map(string), {})
        })
      })

      # Shared persistent services (metastore, Spark History Server).
      auxiliary_services_config = optional(object({
        # Persistent Hive metastore for the virtual cluster's jobs.
        metastore_config = optional(object({
          # Resource name of an existing Dataproc Metastore service.
          # Format: projects/{project}/locations/{location}/services/{service}
          # Accepts a literal resource name today; references attach when a
          # metastore-service kind lands in the catalog.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          dataproc_metastore_service = string
        }))

        # Persistent Spark History Server hosted on another Dataproc cluster.
        spark_history_server_config = optional(object({
          # The Dataproc cluster hosting the Spark History Server. Resolves to
          # the cluster's fully qualified resource name
          # (projects/{project}/regions/{region}/clusters/{cluster}).
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          dataproc_cluster = optional(string, "")
        }))
      }))
    }))

    # Timeout for graceful YARN decommissioning when reducing the number
    # of workers. During this period, YARN waits for running tasks to
    # complete before shutting down nodes. Without this, scaling down
    # can terminate running jobs. Only meaningful for the GCE arm.
    # Format: duration in seconds with 's' suffix (e.g., "3600s").
    # Default: "0s" (immediate decommission).
    graceful_decommission_timeout = optional(string, "")

    # User-defined labels applied to the cluster (and propagated to its
    # VMs). Merged beneath Planton's platform attribution labels
    # (platform keys win on conflict). Not supported by the API for the
    # virtual (GKE-based) arm.
    labels = optional(map(string), {})

    # Engine-side teardown behavior. "DELETE" (default) destroys the
    # cluster; "PREVENT" fails any plan that would destroy it; "ABANDON"
    # removes it from IaC management while leaving it running in GCP.
    deletion_policy = optional(string, "")
  })
}
