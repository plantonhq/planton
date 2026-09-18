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
  description = "GcpBigQueryDataset specification"
  type = object({
    # The GCP project the dataset is created in. Accepts a literal project ID
    # or a reference to a GcpProject resource. If omitted, the provider's
    # default project is used.
    # Immutable: changing the project destroys and recreates the dataset.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Unique identifier for the dataset within the project. Immutable.
    # Must contain only letters (upper/lower), numbers, and underscores;
    # maximum 1024 characters. This is the value SQL queries and downstream
    # tables reference (e.g. SELECT * FROM `project.dataset.table`).
    # Example: "analytics_prod", "raw_events"
    dataset_id = string

    # Geographic location where the dataset and its tables reside. Immutable
    # — moving data requires a new dataset and a copy job. Use multi-regional
    # locations ("US", "EU") for maximum availability, or regional locations
    # ("us-central1", "europe-west1") for data-residency requirements. Every
    # table a query joins must live in the same location.
    location = string

    # User-friendly display name for the dataset.
    friendly_name = optional(string, "")

    # Description of the dataset's contents or purpose.
    description = optional(string, "")

    # Labels applied to the dataset for cost attribution and organization.
    # Merged with Planton's platform labels (which win on key conflicts).
    labels = optional(map(string), {})

    # Resource Manager tags bound to the dataset, as
    # "tagKeys-namespaced-name" -> "tagValue short name" pairs (e.g.
    # "123456789012/environment" -> "production"). Unlike labels, tags
    # participate in IAM conditions and organization policy.
    resource_tags = optional(map(string), {})

    # Default lifetime for all NEW tables created in the dataset, in
    # milliseconds (existing tables keep their expiration). Each table is
    # deleted this long after its creation time. Minimum: 3600000 (1 hour).
    # If not set (0), tables do not automatically expire.
    default_table_expiration_ms = optional(number, 0)

    # Default expiration for partitions in NEW partitioned tables, in
    # milliseconds. If not set (0), partitions do not automatically expire.
    default_partition_expiration_ms = optional(number, 0)

    # Maximum hours of time travel for the dataset (point-in-time recovery
    # window). Range: 48 to 168 hours (2 to 7 days); GCP defaults to 168
    # when unset. Lower values reduce storage cost on PHYSICAL billing;
    # higher values give longer accidental-delete recovery.
    max_time_travel_hours = optional(number, 0)

    # Whether dataset and table names are treated as case-insensitive
    # ("MyTable" and "mytable" become the same table). Immutable.
    # Default: false (case-sensitive).
    is_case_insensitive = optional(bool, false)

    # Default collation specification for string columns in new tables.
    # "und:ci" makes string comparison case-insensitive; empty keeps the
    # default case-sensitive collation.
    default_collation = optional(string, "")

    # Storage billing model for the dataset.
    # "LOGICAL" (GCP default) bills on uncompressed bytes; "PHYSICAL" bills
    # on compressed bytes on disk — often 60-80% cheaper for compressible
    # data, but time-travel storage is billed too. Switching to PHYSICAL is
    # allowed once every 14 days.
    storage_billing_model = optional(string, "")

    # If true, destroying the dataset also deletes every table in it. If
    # false (default), destroy fails while the dataset contains tables —
    # the guard against deleting data with its container.
    delete_contents_on_destroy = optional(bool, false)

    # Cloud KMS key for default table encryption (CMEK). All NEW tables are
    # encrypted with this key unless they specify their own key. The
    # BigQuery service agent must hold roles/cloudkms.cryptoKeyEncrypterDecrypter
    # on the key before the first table is written. Format:
    # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
    # If not set, tables use Google-managed encryption.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # Access control entries for the dataset.
    #
    # AUTHORITATIVE: the entries listed here become the dataset's complete
    # ACL; BigQuery removes anything not listed. If omitted entirely,
    # BigQuery applies default access (project owners = OWNER, editors =
    # WRITER, viewers = READER) — list those explicitly to keep them
    # alongside custom grants.
    #
    # Each entry is either a principal grant (role + one identity) or a
    # resource authorization (view / routine / dataset, no role).
    access = optional(list(object({
      # The role to grant to a principal. Use basic dataset roles (OWNER,
      # WRITER, READER) or predefined IAM roles (roles/bigquery.dataOwner,
      # roles/bigquery.dataEditor, roles/bigquery.dataViewer). BigQuery stores
      # basic and predefined forms interchangeably (OWNER and
      # roles/bigquery.dataOwner are the same grant).
      #
      # Required for principal grants; must be omitted for view/routine/dataset
      # authorizations (those carry implicit read access).
      role = optional(string, "")

      # An email address of a Google Account to grant access to.
      user_by_email = optional(string, "")

      # An email address of a Google Group to grant access to.
      group_by_email = optional(string, "")

      # A domain to grant access to. All users signed in with the domain's
      # account will be granted access (e.g., "example.com").
      domain = optional(string, "")

      # A special group to grant access to. Valid values:
      #   "projectOwners"  -- all project owners
      #   "projectReaders" -- all project viewers
      #   "projectWriters" -- all project editors
      #   "allAuthenticatedUsers" -- all authenticated Google accounts
      special_group = optional(string, "")

      # An IAM member expression to grant access to.
      # Examples: "allUsers", "serviceAccount:sa@project.iam.gserviceaccount.com"
      iam_member = optional(string, "")

      # A view that is authorized to read the dataset's data (no role).
      view = optional(object({
        # The GCP project that contains the view's dataset.
        project_id = string

        # The dataset that contains the view.
        dataset_id = string

        # The ID of the authorized view (a table resource whose type is VIEW).
        table_id = string
      }))

      # A routine (UDF / stored procedure) that is authorized to read the
      # dataset's data (no role).
      routine = optional(object({
        # The GCP project that contains the routine's dataset.
        project_id = string

        # The dataset that contains the routine.
        dataset_id = string

        # The ID of the authorized routine.
        routine_id = string
      }))

      # Another dataset whose resources are authorized to read this dataset's
      # data (no role).
      dataset = optional(object({
        # The GCP project that contains the grantee dataset.
        project_id = string

        # The grantee dataset whose resources are authorized to read this
        # dataset's data.
        dataset_id = string

        # Which resource types in the grantee dataset receive access.
        # Currently the BigQuery API supports "VIEWS".
        target_types = list(string)
      }))

      # Optional IAM condition gating a principal grant (for example, a
      # time-bounded grant). Applies to principal grants only.
      condition = optional(object({
        # The condition expression in IAM Common Expression Language, e.g.
        # request.time < timestamp("2030-01-01T00:00:00Z").
        expression = string

        # Optional short title summarizing the condition's purpose.
        title = optional(string, "")

        # Optional longer description of the condition.
        description = optional(string, "")

        # Optional string indicating the location of the expression for error
        # reporting (for example, a file and position in that file).
        location = optional(string, "")
      }))
    })), [])

    # Makes this dataset a read-only projection of an external source
    # (e.g. an AWS Glue database) through a BigQuery Omni connection,
    # instead of a container for BigQuery-managed tables. Immutable.
    external_dataset_reference = optional(object({
      # The external source this dataset mirrors, e.g.
      # aws-glue://arn:aws:glue:us-east1:1234567:database/database1
      external_source = string

      # The connection (with credentials for the external source) used to read
      # it. Format: projects/{project}/locations/{location}/connections/{id}
      connection = string
    }))

    # Open-source catalog (Hive Metastore compatibility) metadata, letting
    # engines like Spark address this dataset as a Hive database.
    external_catalog_options = optional(object({
      # Default storage location for tables in this catalog database, e.g.
      # gs://bucket/path. Maximum 1024 characters.
      default_storage_location_uri = optional(string, "")

      # Hive-database-style key/value parameters (the whole map is limited to
      # 30 KB by the API).
      parameters = optional(map(string), {})
    }))

    # What destroying this resource does to the dataset. Works alongside
    # delete_contents_on_destroy (which decides whether a NON-EMPTY dataset
    # may be removed); this switch decides whether removal is attempted at
    # all:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the dataset is deleted (GCP refuses while tables remain
    #                unless delete_contents_on_destroy is true)
    #   "PREVENT" -- destroy FAILS; protects a dataset other projects'
    #                queries and authorized views depend on
    #   "ABANDON" -- the dataset is removed from management but keeps
    #                serving queries in GCP
    deletion_policy = optional(string, "")
  })
}
