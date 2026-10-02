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
  description = "AwsFsxDataRepositoryAssociation specification"
  type = object({
    # The AWS region where the association will be created — the file system's
    # region. Example: "us-west-2", "eu-west-1"
    region = string

    # The FSx for Lustre file system the association attaches to. Required.
    # ForceNew. The file system must not use the legacy in-spec S3 link
    # (import_path) — AWS forbids mixing the two generations.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    file_system_id = string

    # The path on the file system that maps to the data repository, beginning
    # with "/" (e.g., "/datasets/2026" or "/" for the whole namespace).
    # 1-4096 characters. ForceNew.
    #
    # Paths must not overlap between associations on the same file system —
    # each directory subtree belongs to at most one repository.
    file_system_path = string

    # The S3 data repository URI, with an optional prefix (e.g.,
    # "s3://training-data" or "s3://training-data/2026/"). 3-900 characters.
    # ForceNew.
    data_repository_path = string

    # S3 events that automatically IMPORT metadata into the Lustre namespace —
    # how the file system tracks the bucket after creation.
    #
    # - "NEW": objects added to the bucket appear as files.
    # - "CHANGED": changed objects refresh their file metadata.
    # - "DELETED": deleted objects remove their files.
    #
    # Empty means no automatic import (the namespace reflects the bucket only
    # as of creation, or via manual import tasks). Updates in place.
    auto_import_events = optional(list(string), [])

    # File-system events that automatically EXPORT back to S3 — how the bucket
    # tracks the file system.
    #
    # - "NEW": new files are written to the bucket.
    # - "CHANGED": changed file contents/metadata update their objects.
    # - "DELETED": deleted files remove their objects.
    #
    # Empty means no automatic export (write results back with manual export
    # tasks). Updates in place.
    auto_export_events = optional(list(string), [])

    # Stripe configuration for imported files: the maximum amount of data per
    # file (in MiB) stored on a single physical disk. Range: 1-512000; AWS
    # defaults to 1024. Updates in place.
    imported_file_chunk_size = optional(number)

    # Run a batch import of the existing S3 metadata into the file-system path
    # as soon as the association is created — without it, only objects that
    # change AFTER creation (per auto_import_events) appear in the namespace.
    # Create-time behavior.
    batch_import_meta_data_on_create = optional(bool, false)

    # Delete the data in the file-system path when the association is deleted.
    # By default the files remain on the file system (only the S3 link goes
    # away). Delete-time behavior; use deliberately.
    delete_data_in_filesystem = optional(bool, false)
  })
}
