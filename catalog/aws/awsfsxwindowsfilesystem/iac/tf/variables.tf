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
  description = "AwsFsxWindowsFileSystem specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Deployment type controlling availability and performance characteristics.
    # ForceNew — cannot be changed after creation.
    #
    # - "SINGLE_AZ_1": first-generation single-AZ. Limited throughput tiers.
    # - "SINGLE_AZ_2": latest single-AZ with higher throughput ceiling and HDD
    #   support. Recommended for most workloads.
    # - "MULTI_AZ_1": multi-AZ with automatic failover across two AZs. Requires
    #   two subnets and a preferred_subnet_id. Supports HDD storage.
    #
    # Default: SINGLE_AZ_2
    deployment_type = optional(string)

    # Storage capacity in GiB.
    #
    # Valid ranges depend on storage type:
    # - SSD: 32–65536 GiB
    # - HDD: 2000–65536 GiB
    #
    # Storage can be increased after creation but never decreased. Leave unset
    # only when restoring from a backup (`backup_id` — capacity comes from the
    # backup).
    storage_capacity_gib = optional(number)

    # Storage media type. ForceNew — cannot be changed after creation.
    #
    # - "SSD": solid-state drives. Sub-millisecond latency. Required for
    #   SINGLE_AZ_1. Recommended for most workloads.
    # - "HDD": hard disk drives. Lower cost, higher latency. Only available for
    #   SINGLE_AZ_2 and MULTI_AZ_1 deployment types. Requires minimum 2000 GiB
    #   storage capacity.
    #
    # The provider's shared storage-type enum also accepts INTELLIGENT_TIERING,
    # but AWS's CreateFileSystem contract scopes Intelligent-Tiering to OpenZFS
    # multi-AZ and Lustre PERSISTENT_2 file systems only — it is not valid for
    # Windows. The two-value domain here is deliberate (spec mirrors AWS where
    # the provider is loose).
    #
    # Default: SSD
    storage_type = optional(string)

    # Throughput capacity in MB/s. Required.
    #
    # Valid values: 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4608, 6144, 9216, 12288.
    # The maximum available throughput depends on the deployment type, and the
    # 4608+ tiers are available only in a handful of regions (per the FSx for
    # Windows performance documentation).
    #
    # Known provider defect at the pinned release: the provider's value list
    # carries a digit-transposition typo — it validates 12228 where AWS's
    # documented top tier is 12288 — so the 12288 tier fails the provider's
    # plan-time validation on both engines until a provider release fixes the
    # list. The spec mirrors AWS's contract, never the typo.
    #
    # Throughput can be changed after creation to scale performance up or down.
    throughput_capacity = optional(number, 0)

    # Subnet IDs for the file system's network interfaces. Required. ForceNew.
    #
    # - SINGLE_AZ_1 / SINGLE_AZ_2: exactly one subnet.
    # - MULTI_AZ_1: exactly two subnets in different availability zones.
    #
    # All compute resources mounting this file system must have SMB network
    # connectivity to these subnets (TCP port 445).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = list(string)

    # Preferred subnet for the active file server in a MULTI_AZ_1 deployment.
    # ForceNew. Required when deployment_type is MULTI_AZ_1. Must be one of the
    # subnets specified in subnet_ids.
    #
    # In a failover event, the standby file server in the other subnet takes over.
    # Ignored for SINGLE_AZ deployments.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    preferred_subnet_id = optional(string, "")

    # Security groups for the file system's network interfaces. ForceNew.
    #
    # Must allow SMB traffic between the file system and its clients:
    # - TCP port 445 (SMB)
    # - TCP port 5985 (WinRM for PowerShell remote administration)
    #
    # Additionally, for Active Directory communication:
    # - TCP/UDP port 53 (DNS), TCP/UDP port 88 (Kerberos),
    #   TCP port 389 (LDAP), TCP port 636 (LDAPS)
    #
    # Up to 50 security groups. When empty, AWS attaches the VPC's default
    # security group.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # Customer-managed KMS key ARN for encryption at rest. ForceNew — the KMS key
    # cannot be changed after creation. When omitted, the file system uses the
    # AWS-managed FSx key. All Windows file systems are encrypted at rest by
    # default; this field upgrades to a customer-managed key.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # ID of an FSx backup to restore this file system from ("backup-...").
    # ForceNew. When set, storage capacity and most settings come from the
    # backup; leave storage_capacity_gib unset.
    backup_id = optional(string, "")

    # ID of an existing AWS Managed Microsoft AD (Directory Service) to join.
    # ForceNew. Mutually exclusive with `self_managed_active_directory`.
    #
    # Use this when you have an AWS Directory Service managed AD already
    # provisioned. The file system joins the domain automatically.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    active_directory_id = optional(string, "")

    # Self-managed Active Directory configuration for joining an on-premises or
    # EC2-hosted AD domain. Mutually exclusive with `active_directory_id`.
    #
    # Use this when your AD domain controller runs outside AWS Directory Service
    # (e.g., on-premises AD, AD on EC2, or Azure AD DS).
    self_managed_active_directory = optional(object({
      # Fully qualified domain name of the self-managed AD directory.
      # Example: "corp.example.com"
      domain_name = string

      # IP addresses of the DNS servers for the AD domain. Required.
      # Must be reachable from the file system's subnets (same VPC CIDR or
      # RFC 1918 private ranges). Minimum 1, maximum 2 IP addresses.
      dns_ips = list(string)

      # Service account username for domain join operations. Mutually exclusive
      # with `domain_join_service_account_secret_arn`. Length: 1-256 characters.
      username = optional(string, "")

      # Service account password for domain join operations. Mutually exclusive
      # with `domain_join_service_account_secret_arn`. Length: 1-256 characters.
      #
      # For production workloads, prefer `domain_join_service_account_secret_arn`
      # to avoid storing credentials in the resource manifest.
      password = optional(string, "")

      # ARN of an AWS Secrets Manager secret containing the service account
      # credentials for domain join. Mutually exclusive with `username`/`password`.
      #
      # The secret must contain a JSON object with "username" and "password" keys.
      # This is the recommended approach for production deployments.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      domain_join_service_account_secret_arn = optional(string, "")

      # Name of the AD group whose members are granted administrative privileges
      # on the file system. Members can administer the file system from a remote
      # PowerShell endpoint using the FSx Remote PowerShell.
      #
      # Default: Domain Admins
      file_system_administrators_group = optional(string)

      # Organizational Unit (OU) distinguished name within the AD directory where
      # the file system's computer object is created.
      #
      # Example: "OU=FSx,DC=corp,DC=example,DC=com"
      #
      # Only the OU immediately above the computer object can be specified.
      # If not provided, the computer object is created in the default "Computers"
      # container in the AD domain. Length: 1-2000 characters.
      organizational_unit_distinguished_name = optional(string, "")
    }))

    # DNS alias names to associate with the file system. Up to 50 aliases.
    #
    # Aliases allow the file system to be accessed via custom DNS names
    # (e.g., "finance.corp.example.com") in addition to the default DNS name.
    # Useful for DFS namespace integration, migration from on-premises filers,
    # or providing user-friendly mount points.
    #
    # Each alias must be a valid DNS name (4-253 characters). You must create a
    # DNS CNAME record pointing each alias to the file system's DNS name.
    aliases = optional(list(string), [])

    # Audit log configuration for tracking file access and file share access
    # events. Logs are sent to CloudWatch Logs for compliance and security
    # monitoring. When omitted, audit logging is disabled.
    audit_log_configuration = optional(object({
      # Logging level for individual file access events (open, read, write, delete,
      # rename, change permissions on files and folders).
      #
      # - "DISABLED": no file access logging.
      # - "SUCCESS_ONLY": log successful access events only.
      # - "FAILURE_ONLY": log failed access attempts only (e.g., access denied).
      # - "SUCCESS_AND_FAILURE": log all access events.
      #
      # Default: DISABLED
      file_access_audit_log_level = optional(string)

      # Logging level for file share access events (connect to share, disconnect,
      # change share permissions).
      #
      # - "DISABLED": no file share access logging.
      # - "SUCCESS_ONLY": log successful share access events only.
      # - "FAILURE_ONLY": log failed share access attempts only.
      # - "SUCCESS_AND_FAILURE": log all share access events.
      #
      # Default: DISABLED
      file_share_access_audit_log_level = optional(string)

      # CloudWatch Logs log group ARN to receive audit events. The log group must
      # start with "/aws/fsx/" as required by AWS. If not set when audit levels are
      # enabled, FSx creates a default log stream in the "/aws/fsx/windows" group.
      #
      # Only valid when at least one audit log level is not DISABLED.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      audit_log_destination = optional(string, "")
    }))

    # SSD IOPS configuration for the file system. Controls the total provisioned
    # IOPS. When omitted, AWS uses AUTOMATIC mode which scales IOPS with storage.
    # Only applicable to SSD storage type.
    disk_iops_configuration = optional(object({
      # IOPS provisioning mode.
      #
      # - "AUTOMATIC": IOPS scale automatically based on storage capacity.
      #   Provides 3 IOPS per GiB of storage.
      # - "USER_PROVISIONED": you specify the exact IOPS via the `iops` field.
      #   Allows higher performance independent of storage size but at extra cost.
      #
      # Default: AUTOMATIC
      mode = optional(string)

      # Total SSD IOPS provisioned. Only valid when mode is "USER_PROVISIONED".
      #
      # Valid range: 0–350000.
      iops = optional(number)
    }))

    # Number of days to retain automatic backups. Range: 0-90. Set to 0 to
    # disable automatic backups.
    #
    # Default: 7
    automatic_backup_retention_days = optional(number)

    # Daily UTC time to start automatic backups, in HH:MM format (e.g., "01:00").
    # If not specified and backups are enabled, AWS chooses a default window.
    daily_automatic_backup_start_time = optional(string, "")

    # Copy tags from the file system to backups. ForceNew.
    copy_tags_to_backups = optional(bool, false)

    # Skip creating a final backup when the file system is deleted.
    #
    # Default: true — deletion is clean by default; set to false to keep a
    # last-resort restore point (the final backup outlives the file system and
    # keeps billing until deleted).
    skip_final_backup = optional(bool)

    # Tags applied to the final backup taken on deletion. Only meaningful when
    # skip_final_backup is false.
    final_backup_tags = optional(map(string), {})

    # Weekly UTC maintenance window in the format "d:HH:MM" where d is the day of
    # the week (1=Monday, 7=Sunday). Example: "7:02:00" for Sunday at 02:00 UTC.
    # If not specified, AWS chooses a default window.
    weekly_maintenance_start_time = optional(string, "")
  })
}
