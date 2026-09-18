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
  description = "AwsFsxOntapStorageVirtualMachine specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # The ID of the FSx for ONTAP file system that this SVM belongs to. Required.
    # ForceNew — the SVM cannot be moved to a different file system.
    #
    # The file system provides the underlying storage, throughput, networking,
    # and HA infrastructure. Multiple SVMs can share a single file system for
    # multi-tenancy scenarios.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    file_system_id = string

    # The name of the SVM within the ONTAP file system. Required. ForceNew.
    #
    # This is the ONTAP SVM name (not the Planton metadata name). It must be
    # unique within the file system and is used as the SVM identity in ONTAP CLI
    # operations, SnapMirror relationships, and DNS endpoint names.
    #
    # Constraints: 1-47 characters, alphanumeric and underscore only (no hyphens,
    # no spaces, no special characters).
    name = string

    # The security style for the root volume and the default for all volumes
    # created under this SVM. ForceNew — cannot be changed after creation.
    #
    # - "UNIX": UNIX permissions (mode bits, uid/gid). Best for Linux/NFS
    #   workloads. This is the most common choice for NFS-only SVMs.
    # - "NTFS": Windows ACLs. Best for Windows/SMB workloads with Active
    #   Directory. Requires AD configuration for meaningful use.
    # - "MIXED": Both UNIX and NTFS permissions. The effective security style
    #   depends on which protocol last set permissions. Use for mixed NFS/SMB
    #   access patterns (advanced use case).
    #
    # Default: UNIX
    root_volume_security_style = optional(string)

    # Password for the SVM administrative user ("vsadmin"). Enables SSH access
    # to the SVM management endpoint for ONTAP CLI operations scoped to this SVM.
    #
    # The vsadmin account can manage volumes, LIFs, export policies, and other
    # SVM-scoped resources. Unlike the file system's fsxadmin account (which has
    # cluster-wide access), vsadmin is limited to this SVM.
    #
    # Length: 8-50 characters. Optional — omit if SVM CLI access is not needed.
    # This value is sensitive and will not be returned in read operations.
    svm_admin_password = optional(string, "")

    # Active Directory configuration for enabling SMB protocol access. Optional.
    #
    # When configured, the SVM joins the specified AD domain and an SMB endpoint
    # is automatically created. Without AD, only NFS and iSCSI endpoints are
    # available.
    #
    # ONTAP SVMs support only self-managed Active Directory (on-premises, on EC2,
    # or Azure AD DS). AWS Managed Microsoft AD is not supported for ONTAP SVMs.
    active_directory_configuration = optional(object({
      # NetBIOS name for the SVM's computer object in Active Directory.
      #
      # This is the short name (up to 15 characters) that identifies the SVM in
      # the AD domain. If omitted, AWS generates a name automatically.
      #
      # Constraints: 1-15 characters, alphanumeric only.
      netbios_name = optional(string, "")

      # Fully qualified domain name of the Active Directory directory.
      # Example: "corp.example.com"
      #
      # Constraints: 1-255 characters. Can be changed after creation (the SVM
      # re-joins the new domain).
      domain_name = string

      # IP addresses of the DNS servers for the AD domain. Required. These must be
      # reachable from the file system's subnets (typically in the same VPC CIDR or
      # connected via VPN/peering).
      #
      # Minimum 1, maximum 3 IPv4 addresses.
      dns_ips = list(string)

      # Service account username for AD domain join operations. Required.
      #
      # This account must have permissions to create computer objects in the
      # specified OU (or the default Computers container).
      #
      # Constraints: 1-256 characters.
      username = string

      # Service account password for AD domain join operations. Required.
      #
      # This value is sensitive and will not be returned in read operations.
      # For production workloads, consider injecting this value via CI/CD secrets
      # management rather than storing it in the resource manifest.
      #
      # Constraints: 1-256 characters.
      password = string

      # Name of the AD domain group whose members are granted administrative
      # privileges on the SVM for file share management.
      #
      # Constraints: 1-256 characters.
      # Default: Domain Admins
      file_system_administrators_group = optional(string)

      # Organizational Unit (OU) distinguished name within the AD directory where
      # the SVM's computer object is created.
      #
      # Example: "OU=FSx,DC=corp,DC=example,DC=com"
      #
      # Only the OU immediately above the computer object can be specified.
      # If not provided, the computer object is created in the default "Computers"
      # container in the AD domain.
      #
      # Constraints: up to 2000 characters.
      organizational_unit_distinguished_name = optional(string, "")
    }))
  })
}
