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
  description = "GcpComputeDisk specification"
  type = object({
    # The GCP project that owns the disk.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the disk. 1-63 characters, lowercase letters, numbers, and
    # hyphens; must start with a letter and cannot end with a hyphen. When
    # omitted, metadata.name is used. Immutable after creation.
    disk_name = optional(string, "")

    # Zone the disk lives in, e.g. "us-central1-a". A disk attaches only to
    # instances in the same zone. Immutable after creation.
    zone = string

    # Human-readable description of the disk.
    description = optional(string, "")

    # Disk type: "pd-standard" (HDD), "pd-balanced" (GCP's default and the
    # sensible general choice), "pd-ssd" (high IOPS), "pd-extreme"
    # (provisioned-IOPS pd), or a hyperdisk type ("hyperdisk-balanced",
    # "hyperdisk-extreme", "hyperdisk-throughput", "hyperdisk-ml") on
    # supported machine families. Immutable after creation.
    type = optional(string, "")

    # Size in GB. Required for empty disks; optional with a source (the
    # source's size is used, and a larger value grows the disk). Grows in
    # place; shrinking is impossible.
    size_gb = optional(number, 0)

    # Source image to initialize the disk from — makes the disk bootable.
    # Accepts an image family ("debian-cloud/debian-12") or a specific
    # image self link. Create-time only.
    image = optional(string, "")

    # Source snapshot to restore the disk from (name or self link).
    # Create-time only.
    source_snapshot = optional(string, "")

    # Existing disk to clone, referenced as another GcpComputeDisk (or a
    # literal self link). Create-time only.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    source_disk = optional(string, "")

    # Customer-managed encryption key (CMEK), referenced as a GcpKmsKey.
    # The Compute Engine service agent
    # (service-<project-number>@compute-system.iam.gserviceaccount.com)
    # must hold roles/cloudkms.cryptoKeyEncrypterDecrypter on the key.
    # When omitted, Google-managed encryption is used. Immutable after
    # creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key = optional(string, "")

    # Provisioned IOPS. Required for pd-extreme and hyperdisk-extreme;
    # tunable on hyperdisk-balanced. Hyperdisk types update in place
    # (at most every 4 hours); other types cannot set this.
    provisioned_iops = optional(number)

    # Provisioned throughput in MB/s. Tunable on hyperdisk-throughput and
    # hyperdisk-balanced; updates in place (at most every 4 hours).
    provisioned_throughput = optional(number)

    # Access mode of the disk:
    #   ""                     -- same as "READ_WRITE_SINGLE" (GCP default)
    #   "READ_WRITE_SINGLE"    -- one instance read-write
    #   "READ_WRITE_MANY"      -- many instances read-write (hyperdisk-ml
    #                             and supported multi-writer types)
    #   "READ_ONLY_MANY"       -- many instances read-only
    access_mode = optional(string, "")

    # CPU architecture the disk's contents target: "X86_64" or "ARM64".
    # Relevant for bootable disks; normally inferred from the image.
    # Immutable after creation.
    architecture = optional(string, "")

    # Create the disk in confidential-compute mode (hyperdisk SKUs only;
    # requires kms_key).
    enable_confidential_compute = optional(bool, false)

    # Physical block size in bytes: 4096 (default) or 16384.
    physical_block_size_bytes = optional(number)

    # Take a snapshot of the disk immediately before it is destroyed — a
    # last-resort recovery net for precious data volumes. The snapshot is
    # named "<disk-name>-YYYYMMDD-HHmmss" unless
    # snapshot_before_destroy_prefix overrides the prefix.
    create_snapshot_before_destroy = optional(bool, false)

    # Custom name prefix for the snapshot taken by
    # create_snapshot_before_destroy.
    snapshot_before_destroy_prefix = optional(string, "")

    # URL or name of the storage pool to create the disk in (hyperdisk
    # storage pools).
    storage_pool = optional(string, "")

    # User labels merged with Planton attribution labels (which win on key
    # conflicts).
    labels = optional(map(string), {})

    # Resource Manager tags bound to the disk for org-policy and IAM
    # conditions. Keys in the form "tagKeys/{id}", values "tagValues/{id}".
    # Create-time only.
    resource_manager_tags = optional(map(string), {})

    # Guest OS features to enable when this disk is used as a boot disk,
    # e.g. ["UEFI_COMPATIBLE", "SECURE_BOOT", "GVNIC", "MULTI_IP_SUBNET",
    # "WINDOWS"]. The accepted set evolves with GCP — see "Enabling guest
    # operating system features" in the Compute Engine docs. Create-time
    # only.
    guest_os_features = optional(list(string), [])

    # License URIs applicable to this disk, e.g. a Windows Server or
    # SQL Server license
    # ("https://www.googleapis.com/compute/v1/projects/windows-cloud/global/licenses/windows-server-core").
    # Normally inherited from the source image; set explicitly when
    # importing raw disks that need bring-your-own-license attribution.
    # Create-time only.
    licenses = optional(list(string), [])

    # Source INSTANT snapshot to create the disk from (name, partial or
    # full URL). Instant snapshots are near-instant, same-region-only
    # point-in-time copies — the fast restore path, vs. standard
    # snapshots' cross-region durability. Create-time only.
    source_instant_snapshot = optional(string, "")

    # Full Google Cloud Storage URI of a raw disk image to create the
    # disk from (e.g. "https://storage.googleapis.com/bucket/image.vmdk"
    # or a gs:// URI) — the import-a-disk-file path, skipping the
    # intermediate compute image. Create-time only.
    source_storage_object = optional(string, "")

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the disk (and its data) is deleted
    #   "PREVENT" -- destroy FAILS; a guard rail for precious data volumes
    #                (create_snapshot_before_destroy is the softer net)
    #   "ABANDON" -- the disk is removed from management but left running
    #                in GCP
    deletion_policy = optional(string, "")

    # Service account used for the encryption request of kms_key (CMEK).
    # When omitted, the Compute Engine default service agent is used.
    # Only meaningful together with kms_key. Immutable after creation.
    kms_key_service_account = optional(string, "")

    # Decrypts the source image when it is itself CMEK-encrypted. Only
    # valid together with image.
    source_image_encryption = optional(object({
      # The KMS key the source was encrypted with, referenced as a GcpKmsKey
      # or a literal self link. The service agent performing the read needs
      # roles/cloudkms.cryptoKeyEncrypterDecrypter on it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key = string

      # Service account used for the decryption request. When omitted, the
      # Compute Engine default service agent is used.
      kms_key_service_account = optional(string, "")
    }))

    # Decrypts the source snapshot when it is itself CMEK-encrypted. Only
    # valid together with source_snapshot.
    source_snapshot_encryption = optional(object({
      # The KMS key the source was encrypted with, referenced as a GcpKmsKey
      # or a literal self link. The service agent performing the read needs
      # roles/cloudkms.cryptoKeyEncrypterDecrypter on it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key = string

      # Service account used for the decryption request. When omitted, the
      # Compute Engine default service agent is used.
      kms_key_service_account = optional(string, "")
    }))

    # Makes this disk an ASYNC REPLICATION SECONDARY of the referenced
    # primary disk (another GcpComputeDisk, or a literal disk self link in
    # another region). Replication starts when the pair is activated on
    # the primary; the secondary must match the primary's size and type.
    # Create-time only.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    async_primary_disk = optional(string, "")
  })
}
