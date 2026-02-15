# ---------------------------------------------------------------------------
# AWS EFS File System
# ---------------------------------------------------------------------------
# Core file system with encryption, throughput, and lifecycle policies.
# ForceNew attributes: encrypted, kms_key_id, performance_mode, availability_zone_name.
# ---------------------------------------------------------------------------

resource "aws_efs_file_system" "this" {
  creation_token = local.resource_name

  # Encryption at rest (ForceNew)
  encrypted  = var.encrypted
  kms_key_id = local.kms_key_id

  # Performance and throughput (ForceNew for performance_mode)
  performance_mode                = var.performance_mode
  throughput_mode                 = var.throughput_mode
  provisioned_throughput_in_mibps  = local.provisioned_throughput_in_mibps

  # One Zone storage (ForceNew)
  availability_zone_name = local.availability_zone_name

  # Lifecycle policies — dynamic blocks for IA, Archive, Primary transitions
  dynamic "lifecycle_policy" {
    for_each = local.lifecycle_policies
    content {
      transition_to_ia                    = try(lifecycle_policy.value.transition_to_ia, null)
      transition_to_archive               = try(lifecycle_policy.value.transition_to_archive, null)
      transition_to_primary_storage_class = try(lifecycle_policy.value.transition_to_primary_storage_class, null)
    }
  }

  tags = local.tags
}

# ---------------------------------------------------------------------------
# Mount Targets — one per subnet
# ---------------------------------------------------------------------------
# Each subnet gets a mount target. Security groups must allow NFS (TCP 2049).
# ---------------------------------------------------------------------------

resource "aws_efs_mount_target" "this" {
  for_each = local.subnet_ids_set

  file_system_id  = aws_efs_file_system.this.id
  subnet_id       = each.key
  security_groups = length(var.security_group_ids) > 0 ? var.security_group_ids : null
}

# ---------------------------------------------------------------------------
# Access Points — one per entry, keyed by name
# ---------------------------------------------------------------------------
# Application-specific entry points with POSIX user and root directory.
# ---------------------------------------------------------------------------

resource "aws_efs_access_point" "this" {
  for_each = local.access_point_map

  file_system_id = aws_efs_file_system.this.id

  # POSIX user identity enforcement
  dynamic "posix_user" {
    for_each = try(each.value.posix_user, null) != null ? [each.value.posix_user] : []
    content {
      uid            = posix_user.value.uid
      gid            = posix_user.value.gid
      secondary_gids = try(posix_user.value.secondary_gids, null)
    }
  }

  # Root directory restriction
  dynamic "root_directory" {
    for_each = try(each.value.root_directory, null) != null ? [each.value.root_directory] : []
    content {
      path = root_directory.value.path

      dynamic "creation_info" {
        for_each = try(root_directory.value.creation_info, null) != null ? [root_directory.value.creation_info] : []
        content {
          owner_uid   = creation_info.value.owner_uid
          owner_gid   = creation_info.value.owner_gid
          permissions = creation_info.value.permissions
        }
      }
    }
  }

  tags = {
    Name = each.key
  }
}

# ---------------------------------------------------------------------------
# Backup Policy — automatic daily backups
# ---------------------------------------------------------------------------

resource "aws_efs_backup_policy" "this" {
  count = var.backup_enabled ? 1 : 0

  file_system_id = aws_efs_file_system.this.id

  backup_policy {
    status = "ENABLED"
  }
}

# ---------------------------------------------------------------------------
# File System Policy — IAM resource policy
# ---------------------------------------------------------------------------
# JSON policy for access control (encryption in transit, principal restrictions).
# ---------------------------------------------------------------------------

resource "aws_efs_file_system_policy" "this" {
  count = var.policy != "" ? 1 : 0

  file_system_id = aws_efs_file_system.this.id
  policy         = var.policy
}
