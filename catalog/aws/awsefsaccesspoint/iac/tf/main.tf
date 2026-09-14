# ---------------------------------------------------------------------------
# AWS EFS Access Point
# ---------------------------------------------------------------------------
# The ENTIRE access point is create-time immutable (only tags mutate):
# changing the file system, POSIX user, or root directory replaces it. AWS
# assigns the fsap- identity; the Name tag carries the human name.
# ---------------------------------------------------------------------------

resource "aws_efs_access_point" "this" {
  # The file system reference arrives pre-resolved as a plain ID string.
  file_system_id = var.spec.file_system_id

  # POSIX identity enforcement: when set, every file operation through this
  # access point uses this UID/GID regardless of what the NFS client claims
  # -- the core least-privilege mechanism of access points.
  dynamic "posix_user" {
    for_each = var.spec.posix_user != null ? [var.spec.posix_user] : []
    content {
      # Both IDs are required by the API inside the block, so neither is null
      # here; 0 is root and passes through as 0.
      uid            = posix_user.value.uid
      gid            = posix_user.value.gid
      secondary_gids = length(posix_user.value.secondary_gids) > 0 ? posix_user.value.secondary_gids : null
    }
  }

  # Root directory restriction: the path is exposed as "/" to clients.
  # creation_info lets EFS create a not-yet-existing path with the right
  # ownership on first mount -- without it, mounting a missing path fails.
  dynamic "root_directory" {
    for_each = var.spec.root_directory != null ? [var.spec.root_directory] : []
    content {
      path = root_directory.value.path

      dynamic "creation_info" {
        for_each = root_directory.value.creation_info != null ? [root_directory.value.creation_info] : []
        content {
          owner_uid   = creation_info.value.owner_uid
          owner_gid   = creation_info.value.owner_gid
          permissions = creation_info.value.permissions
        }
      }
    }
  }

  tags = local.aws_tags
}
