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
  description = "AwsGlueCatalogDatabase specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Human-readable description of the database. Helps teams understand the
    # purpose and contents of this catalog namespace.
    #
    # Maximum 2048 characters (enforced by the AWS API).
    #
    # Examples:
    # - "Sales analytics data lake — raw and curated tables from the sales pipeline"
    # - "Clickstream events from web and mobile applications"
    description = optional(string, "")

    # ID of the Data Catalog in which to create the database. Defaults to the
    # AWS account ID of the deploying account; set it only to create the
    # database inside ANOTHER account's catalog (a cross-account governance
    # pattern that requires a matching catalog resource policy on that
    # account). Fixed at creation time (changing it replaces the database).
    catalog_id = optional(string, "")

    # Default S3 URI for tables created in this database. When a Glue Crawler or
    # CREATE TABLE statement does not specify a location, this path is used as the
    # base directory.
    #
    # Format: "s3://bucket-name/optional-prefix/"
    #
    # When omitted, each table must specify its own storage location explicitly.
    # Setting this is recommended for organized data lakes where all tables in a
    # database share a common S3 prefix.
    #
    # One-way in practice: once applied, REMOVING this field does not clear the
    # location at AWS -- the provider keeps the last-known value. Point it at a
    # new prefix to change it.
    #
    # This is a plain string (not StringValueOrRef) because it is an S3 URI with
    # a user-defined path prefix, not a direct resource identifier.
    location_uri = optional(string, "")

    # Free-form key-value properties attached to the database. Consumed by
    # engines and governance tooling that read catalog metadata -- for example
    # classification hints, team ownership labels, or engine-specific switches.
    # These are catalog metadata, NOT AWS resource tags (identity tags derive
    # from metadata.name/org/env automatically).
    parameters = optional(map(string), {})

    # Default Lake Formation permissions granted on tables CREATED in this
    # database. When omitted, AWS applies its default grant -- ALL permissions
    # to the virtual group IAM_ALLOWED_PRINCIPALS -- which keeps plain
    # IAM-policy access working (the compatibility mode most accounts run in).
    #
    # Lake Formation-governed data lakes typically override this: grant to
    # specific principals, or supply an entry with an empty permission list to
    # stop granting IAM_ALLOWED_PRINCIPALS on new tables entirely (the
    # recommended hardening step when migrating to Lake Formation permissions).
    create_table_default_permissions = optional(list(object({
      # Lake Formation permissions to grant. Valid values: "ALL", "SELECT",
      # "ALTER", "DROP", "DELETE", "INSERT", "CREATE_DATABASE", "CREATE_TABLE",
      # "DATA_LOCATION_ACCESS". An entry with an EMPTY list (and no principal)
      # is meaningful: it disables the default IAM_ALLOWED_PRINCIPALS grant on
      # newly created tables.
      permissions = optional(list(string), [])

      # The principal receiving the grant: an IAM user/role ARN, an AWS account
      # ID, or the virtual group "IAM_ALLOWED_PRINCIPALS" (grants to any
      # principal whose IAM policy allows the action -- Lake Formation's
      # IAM-compatibility mode). 1-255 characters when set.
      principal = optional(string, "")
    })), [])

    # Creates this database as a RESOURCE LINK: a local pointer to a database
    # that lives in another account or region and was shared via AWS RAM /
    # Lake Formation. Queries against the link (from Athena, Redshift Spectrum,
    # EMR) resolve to the target database's tables, subject to the permissions
    # granted on the share.
    #
    # A resource link carries no storage or schema of its own, so combining it
    # with location_uri or create-table permissions has no effect; leave those
    # unset. Fixed at creation time (changing it replaces the database).
    target_database = optional(object({
      # ID of the Data Catalog that owns the target database -- the sharing
      # account's AWS account ID.
      catalog_id = string

      # Name of the target database inside the owning catalog.
      database_name = string

      # Region of the target database, e.g. "us-east-1". Set for cross-REGION
      # links; omit when the target lives in the same region as this database.
      region = optional(string, "")
    }))

    # Creates this database as a FEDERATED database: a projection of an
    # external data source into the catalog through a Glue connection. The
    # primary use today is querying Amazon Redshift datashares from Athena and
    # other catalog consumers without copying data.
    federated_database = optional(object({
      # Unique identifier of the federated source, e.g. the Redshift datashare
      # ARN being projected into the catalog.
      identifier = optional(string, "")

      # Name of the Glue connection that carries the federation, e.g.
      # "aws:redshift" for the AWS-managed Redshift federation connection.
      connection_name = optional(string, "")
    }))
  })
}
