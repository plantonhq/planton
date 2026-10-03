variable "metadata" {
  description = "Catalog object metadata"
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
  description = "GcpComputeImage specification"
  type = object({
    # The GCP project that owns the image: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the image, unique in the project: 1-63 lowercase letters,
    # digits, and hyphens, starting with a letter and not ending with a
    # hyphen. Version it ("web-base-20261001") and let family carry the
    # stable name. Defaults to metadata.name. Immutable.
    image_name = optional(string, "")

    # A human-readable description of the image. Immutable.
    description = optional(string, "")

    # The image family: the stable name consumers boot from
    # ("projects/{project}/global/images/family/{family}" resolves to the
    # newest non-deprecated image in it). Same naming rules as image_name.
    # Immutable.
    family = optional(string, "")

    # Source: a disk to image, referenced as a GcpComputeDisk (its
    # self_link) or a literal disk self link. Stop or detach the VM using
    # the disk first so the image is consistent. Exactly one source.
    # Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    source_disk = optional(string, "")

    # Source: an image to copy, referenced as another GcpComputeImage (its
    # self_link) or a literal image self link (e.g.
    # "projects/debian-cloud/global/images/family/debian-12" for a public
    # image). Exactly one source. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    source_image = optional(string, "")

    # Source: a snapshot to image (name or self link). Exactly one source.
    # Immutable.
    source_snapshot = optional(string, "")

    # Source: a raw disk tarball in Cloud Storage. Exactly one source.
    # Immutable.
    raw_disk = optional(object({
      # The tarball's full Cloud Storage URL, e.g.
      # "https://storage.googleapis.com/my-bucket/images/web-base.tar.gz".
      # The archive holds one file named disk.raw. Required.
      source = string

      # The archive's SHA-1 checksum, base64-encoded; Google verifies it when
      # set.
      sha1 = optional(string, "")

      # The archive format. "TAR" is the only format Google accepts and the
      # default.
      container_type = optional(string, "")
    }))

    # Customer-managed encryption key (CMEK) for the image, referenced as a
    # GcpKmsKey or a GcpKmsKeyHandle (Autokey serves images). The Compute
    # Engine service agent
    # (service-<project-number>@compute-system.iam.gserviceaccount.com)
    # must hold roles/cloudkms.cryptoKeyEncrypterDecrypter on the key. When
    # omitted, Google-managed encryption is used. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key = optional(string, "")

    # Service account used for the encryption request of kms_key. When
    # omitted, the Compute Engine service agent is used. Only meaningful
    # together with kms_key. Immutable.
    kms_key_service_account = optional(string, "")

    # Decrypts the source disk when it is itself CMEK-encrypted. Only valid
    # together with source_disk.
    source_disk_encryption = optional(object({
      # The KMS key the source was encrypted with, referenced as a GcpKmsKey
      # or a literal self link. The service agent performing the read needs
      # roles/cloudkms.cryptoKeyEncrypterDecrypter on it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key = string

      # Service account used for the decryption request. When omitted, the
      # Compute Engine service agent is used.
      kms_key_service_account = optional(string, "")
    }))

    # Decrypts the source image when it is itself CMEK-encrypted. Only valid
    # together with source_image.
    source_image_encryption = optional(object({
      # The KMS key the source was encrypted with, referenced as a GcpKmsKey
      # or a literal self link. The service agent performing the read needs
      # roles/cloudkms.cryptoKeyEncrypterDecrypter on it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key = string

      # Service account used for the decryption request. When omitted, the
      # Compute Engine service agent is used.
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
      # Compute Engine service agent is used.
      kms_key_service_account = optional(string, "")
    }))

    # The image's size in GB, at least the source's size. Unset takes the
    # source's size. Immutable.
    disk_size_gb = optional(number)

    # Guest OS features VMs booted from this image may use, e.g.
    # ["UEFI_COMPATIBLE", "SECURE_BOOT", "GVNIC", "VIRTIO_SCSI_MULTIQUEUE",
    # "SEV_CAPABLE", "SEV_SNP_CAPABLE", "TDX_CAPABLE", "IDPF",
    # "MULTI_IP_SUBNET", "SUSPEND_RESUME_COMPATIBLE", "WINDOWS"]. The
    # accepted set evolves with GCP -- see "Enabling guest operating system
    # features" in the Compute Engine docs. Unset inherits the source's
    # features. Immutable.
    guest_os_features = optional(list(string), [])

    # License URIs the image carries, e.g. a Windows Server or SQL Server
    # license. Unset inherits the source's licenses; set explicitly when
    # importing a raw disk that needs bring-your-own-license attribution.
    # Immutable.
    licenses = optional(list(string), [])

    # Where the image's data is stored: one multi-region ("us", "eu",
    # "asia") or one region. Unset takes the multi-region nearest the
    # source. Immutable.
    storage_locations = optional(list(string), [])

    # The UEFI Secure Boot keys VMs booted from this image start with. Unset
    # takes Google's default certificates. Immutable.
    shielded_instance_initial_state = optional(object({
      # The Platform Key, which authorizes changes to the key exchange keys.
      pk = optional(object({
        # The raw content, base64-encoded. Public certificates, not secrets.
        # Required.
        content = string

        # "X509" (a certificate) or "BIN" (raw bytes). Empty lets Google infer.
        file_type = optional(string, "")
      }))

      # Key Exchange Keys, which authorize changes to db and dbx.
      keks = optional(list(object({
        # The raw content, base64-encoded. Public certificates, not secrets.
        # Required.
        content = string

        # "X509" (a certificate) or "BIN" (raw bytes). Empty lets Google infer.
        file_type = optional(string, "")
      })), [])

      # The allowed signature database: certificates of boot loaders and
      # drivers allowed to run.
      dbs = optional(list(object({
        # The raw content, base64-encoded. Public certificates, not secrets.
        # Required.
        content = string

        # "X509" (a certificate) or "BIN" (raw bytes). Empty lets Google infer.
        file_type = optional(string, "")
      })), [])

      # The forbidden signature database: revoked certificates and hashes.
      dbxs = optional(list(object({
        # The raw content, base64-encoded. Public certificates, not secrets.
        # Required.
        content = string

        # "X509" (a certificate) or "BIN" (raw bytes). Empty lets Google infer.
        file_type = optional(string, "")
      })), [])
    }))

    # Resource Manager tags bound to the image for org-policy and IAM
    # conditions. Keys in the form "tagKeys/{id}", values "tagValues/{id}".
    # Immutable.
    resource_manager_tags = optional(map(string), {})

    # User labels merged with Planton attribution labels (which win on key
    # conflicts). The one setting that updates in place.
    labels = optional(map(string), {})

    # Deletion policy -- what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the image is deleted
    #   "PREVENT" -- destroy FAILS; a guard for an image fleets boot from
    #   "ABANDON" -- the image is removed from management but kept in GCP
    deletion_policy = optional(string, "")
  })
}
