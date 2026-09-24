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
  description = "GcpBigQueryConnection specification"
  type = object({
    # The GCP project the connection lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Where the connection lives, and so which datasets can use it: a
    # BigQuery multi-region (US, EU) or a region (us-central1,
    # europe-west1). Each arm has its own rule -- Cloud SQL must match (with
    # us-central1 -> US and europe-west1 -> EU allowed), Spanner matches the
    # instance's region, AWS uses aws-{region} (e.g. aws-us-east-1) and
    # Azure azure-{region} (e.g. azure-eastus2). Empty leaves Google's
    # default. Immutable.
    location = optional(string, "")

    # The connection's ID. Defaults to metadata.name. Immutable.
    connection_id = optional(string, "")

    # A human-readable name shown in the console.
    friendly_name = optional(string, "")

    # A free-text description.
    description = optional(string, "")

    # A Cloud KMS key encrypting the connection's stored credential (CMEK):
    # a GcpKmsKey reference or a literal
    # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}.
    # Empty uses Google-managed keys.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # AWS through BigQuery Omni.
    aws = optional(object({
      # The ARN of the AWS IAM role BigQuery assumes, e.g.
      # arn:aws:iam::123456789012:role/bigquery-omni. The role must trust the
      # Google-owned identity the connection outputs.
      iam_role_id = string
    }))

    # Azure through BigQuery Omni.
    azure = optional(object({
      # Your Azure AD (Entra ID) tenant ID -- the directory that holds the
      # data.
      customer_tenant_id = string

      # The client ID of YOUR Azure application that hosts a federated
      # credential for the connection's Google identity (the recommended,
      # secretless setup). Empty makes Google create a multi-tenant Azure
      # application you grant consent to (see the azure_* outputs).
      federated_application_client_id = optional(string, "")
    }))

    # A Google-managed service account BigQuery acts as when it reads Cloud
    # Storage (BigLake and object tables), calls Vertex AI (remote models),
    # or invokes Cloud Run functions (remote functions). Google's arm has no
    # settings: true declares it. Grant the cloud_resource_service_account_id
    # output access to what it reads.
    cloud_resource = optional(bool, false)

    # A Cloud Spanner database.
    cloud_spanner = optional(object({
      # The database, as project/instance/database (slashes, no
      # "projects/" prefix).
      database = string

      # A Spanner fine-grained access control role the queries run as (it
      # must start with a letter; letters, digits, underscores). Empty reads
      # with the caller's database-level permissions.
      database_role = optional(string, "")

      # Read with Spanner parallelism. Required by use_data_boost and
      # max_parallelism.
      use_parallelism = optional(bool, false)

      # Run the queries on Spanner Data Boost -- independent compute that
      # leaves the instance's serving capacity untouched (billed separately
      # by Spanner). Requires use_parallelism.
      use_data_boost = optional(bool, false)

      # The most parallel reads per query on Data Boost. Requires both
      # use_parallelism and use_data_boost. Empty lets Spanner pick from the
      # instance configuration.
      max_parallelism = optional(number, 0)
    }))

    # A Cloud SQL database.
    cloud_sql = optional(object({
      # The Cloud SQL instance, as project:region:instance -- a GcpCloudSql
      # reference (its connection_name output) or a literal in that form.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      instance_id = string

      # The database name inside the instance.
      database = string

      # The engine: POSTGRES or MYSQL.
      type = string

      # The database user and password.
      credential = object({
        # The database username.
        username = string

        # The database user's password. Stored by BigQuery and never returned.
        password = string
      })
    }))

    # The BigQuery Connector framework.
    configuration = optional(object({
      # The connector, e.g. google-alloydb, google-cloudsql-mysql,
      # google-cloudsql-postgres. Immutable.
      connector_id = string

      # The data source the connector reads.
      asset = object({
        # The database name.
        database = optional(string, "")

        # The full resource name of the Google Cloud resource -- for AlloyDB,
        # //alloydb.googleapis.com/projects/{project}/locations/{region}/clusters/{cluster}/instances/{instance}.
        google_cloud_resource = optional(string, "")
      })

      # Database-user authentication. The connection outputs the service
      # account the connector uses (connector_service_account).
      username_password = optional(object({
        # The database username.
        username = string

        # The database user's password, sent as Google's plaintext secret form
        # and stored by BigQuery; never returned.
        password = string
      }))

      # The endpoint to connect to, as host:port, when the connector needs one
      # spelled out.
      host_port = optional(string, "")

      # A Private Service Connect network attachment the connector reaches a
      # private endpoint through, as
      # projects/{project}/regions/{region}/networkAttachments/{attachment}.
      network_attachment = optional(string, "")
    }))

    # Stored procedures for Apache Spark.
    spark = optional(object({
      # An existing Dataproc Metastore service the Spark code uses as its Hive
      # metastore, as projects/{project}/locations/{region}/services/{service}.
      metastore_service = optional(string, "")

      # An existing Dataproc cluster that serves as the Spark History Server
      # for the procedures' runs, as
      # projects/{project}/regions/{region}/clusters/{cluster}.
      history_server_dataproc_cluster = optional(string, "")
    }))

    # What happens to the connection when this resource is destroyed:
    #   "" / "DELETE" -- deleted (tables and routines using it stop working)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- it leaves management and stays in GCP
    deletion_policy = optional(string, "")
  })
}
