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
  description = "AwsEfsAccessPoint specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # The Elastic File System this access point enters. ForceNew — an access
    # point cannot be moved between file systems.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    file_system_id = string

    # POSIX user and group identity enforced for ALL file operations through
    # this access point. When set, the NFS client's identity is overridden —
    # regardless of what UID/GID the client claims, all operations use these
    # values. ForceNew.
    #
    # Omit to let clients keep their own POSIX identity (rare — enforcing the
    # identity is most of the point of an access point).
    posix_user = optional(object({
      # POSIX user ID (0–4294967295). All file system operations through this
      # access point use this UID as the file owner. Required inside posix_user;
      # 0 is root.
      uid = number

      # POSIX primary group ID (0–4294967295). All file system operations through
      # this access point use this GID as the file group. Required inside
      # posix_user; 0 is root.
      gid = number

      # Secondary POSIX group IDs supplementing the primary GID for group
      # permission checks. Maximum 16 (the NFS AUTH_SYS group limit).
      secondary_gids = optional(list(number), [])
    }))

    # Root directory exposed as "/" when mounting through this access point.
    # ForceNew. Omit to expose the entire file system.
    root_directory = optional(object({
      # Path on the file system to expose as the root directory. Must be an
      # absolute path (starts with "/"), up to 4 subdirectories deep and at most
      # 100 characters (AWS limits, enforced server-side).
      #
      # If this path does not exist yet, provide `creation_info` — EFS creates the
      # directory with that ownership and permissions on first mount. Without it,
      # mounting an access point whose path does not exist fails.
      path = string

      # POSIX ownership and permissions applied when EFS creates the root
      # directory. Required in practice whenever `path` does not already exist on
      # the file system (AWS validates existence at mount time, not create time).
      creation_info = optional(object({
        # POSIX user ID for the directory owner (0–4294967295).
        owner_uid = optional(number, 0)

        # POSIX group ID for the directory owner (0–4294967295).
        owner_gid = optional(number, 0)

        # POSIX permissions for the directory, as 3–4 octal digits (e.g., "755",
        # "0755", "0750").
        permissions = string
      }))
    }))
  })
}
