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
  description = "GcpDatastreamConnectionProfile specification"
  type = object({
    # The GCP project the profile lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Datastream region, e.g. "us-central1". Streams using the profile
    # live in the same region. Immutable.
    location = string

    # The profile's ID. Defaults to metadata.name. Immutable.
    connection_profile_id = optional(string, "")

    # The name shown in the console. Defaults to metadata.name.
    display_name = optional(string, "")

    # Labels on the profile. The platform attribution labels are added on
    # top and win on a key conflict.
    labels = optional(map(string), {})

    # Create the profile without Google's connectivity test. Useful when the
    # database or its network is not reachable yet; a wrong host or
    # password then surfaces when a stream starts. Immutable.
    create_without_validation = optional(bool, false)

    # A BigQuery destination. Google's block has no settings -- the stream
    # decides datasets -- so true declares it. Datastream's service agent
    # writes with BigQuery Data Editor on the target datasets.
    bigquery_profile = optional(bool, false)

    # A Cloud Storage destination.
    gcs_profile = optional(object({
      # The bucket name -- a GcpGcsBucket reference (its bucket_name output) or
      # a literal. Datastream's service agent needs object write access on it
      # (grant roles/storage.objectAdmin through the bucket's iam_members).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      bucket = string

      # The folder inside the bucket every stream using this profile writes
      # under, e.g. "/datastream". A stream's own path nests below it.
      root_path = optional(string, "")
    }))

    # A MySQL source.
    mysql_profile = optional(object({
      # The server's address. Defaults to a GcpCloudSql reference's public_ip
      # -- Google's direct path to Cloud SQL, with Datastream's regional IPs
      # allowlisted in the instance's authorized networks. For a private
      # instance, point at a proxy VM in your VPC reached through a private
      # connection (a reference with another fieldPath, or a literal).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      hostname = string

      # The port. Empty uses 3306.
      port = optional(number, 0)

      # The replication user.
      username = string

      # The user's password, stored by Google and never returned. Mutually
      # exclusive with secret_manager_stored_password.
      password = optional(string, "")

      # The Secret Manager secret VERSION holding the password, as
      # projects/{project}/secrets/{secret}/versions/{version} -- a
      # GcpSecretManagerSecret reference (its latest_version_name output, set
      # when the secret declares an initial version) or a literal version
      # name. Datastream's service agent reads it
      # (roles/secretmanager.secretAccessor on the secret). Mutually exclusive
      # with password.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      secret_manager_stored_password = optional(string, "")

      # TLS for the session.
      ssl_config = optional(object({
        # PEM certificate of the CA that signed the server's certificate.
        ca_certificate = optional(string, "")

        # PEM certificate Datastream presents to the server. Google requires
        # client_key and ca_certificate with it.
        client_certificate = optional(string, "")

        # PEM private key of client_certificate. Google requires
        # client_certificate and ca_certificate with it.
        client_key = optional(string, "")
      }))
    }))

    # A PostgreSQL source.
    postgresql_profile = optional(object({
      # The server's address. Defaults to a GcpCloudSql reference's public_ip
      # -- Google's direct path to Cloud SQL, with Datastream's regional IPs
      # allowlisted. AlloyDB: reference a GcpAlloydbInstance with fieldPath
      # status.outputs.ip_address, reached through a private connection and a
      # proxy VM. A literal takes any hostname or IP.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      hostname = string

      # The port. Empty uses 5432.
      port = optional(number, 0)

      # The replication user (it needs the REPLICATION attribute, or
      # cloudsqlsuperuser's replication role on Cloud SQL).
      username = string

      # The user's password, stored by Google and never returned. Mutually
      # exclusive with secret_manager_stored_password.
      password = optional(string, "")

      # The Secret Manager secret VERSION holding the password -- a
      # GcpSecretManagerSecret reference (its latest_version_name output) or a
      # literal projects/{project}/secrets/{secret}/versions/{version}.
      # Datastream's service agent reads it. Mutually exclusive with password.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      secret_manager_stored_password = optional(string, "")

      # The database to replicate from.
      database = string

      # TLS for the session.
      ssl_config = optional(object({
        # The server proves its identity.
        server_verification = optional(object({
          # PEM root CA certificate of the server.
          ca_certificate = string
        }))

        # Both the server and Datastream prove their identity.
        server_and_client_verification = optional(object({
          # PEM root CA certificate of the server.
          ca_certificate = string

          # PEM certificate Datastream presents, signed by a CA the server trusts.
          client_certificate = string

          # PEM private key of client_certificate.
          client_key = string
        }))
      }))
    }))

    # An Oracle source.
    oracle_profile = optional(object({
      # The server's hostname or IP.
      hostname = string

      # The port. Empty uses 1521.
      port = optional(number, 0)

      # The database user.
      username = string

      # The user's password, stored by Google and never returned. Mutually
      # exclusive with secret_manager_stored_password.
      password = optional(string, "")

      # The Secret Manager secret VERSION holding the password -- a
      # GcpSecretManagerSecret reference (its latest_version_name output) or a
      # literal version name. Mutually exclusive with password.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      secret_manager_stored_password = optional(string, "")

      # The database service (the Oracle service name) to connect to, e.g.
      # "ORCL".
      database_service = string

      # Extra Oracle connection-string attributes, as key/value pairs.
      connection_attributes = optional(map(string), {})
    }))

    # A SQL Server source.
    sql_server_profile = optional(object({
      # The server's address. Defaults to a GcpCloudSql reference's public_ip,
      # with Datastream's regional IPs allowlisted; a literal takes any
      # hostname or IP.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      hostname = string

      # The port. Empty uses 1433.
      port = optional(number, 0)

      # The database user.
      username = string

      # The user's password, stored by Google and never returned. Mutually
      # exclusive with secret_manager_stored_password.
      password = optional(string, "")

      # The Secret Manager secret VERSION holding the password -- a
      # GcpSecretManagerSecret reference (its latest_version_name output) or a
      # literal version name. Mutually exclusive with password.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      secret_manager_stored_password = optional(string, "")

      # The database to replicate from.
      database = string
    }))

    # A MongoDB source.
    mongodb_profile = optional(object({
      # The hosts to connect to -- every member of a replica set for the
      # standard format, the one SRV domain for the SRV format.
      host_addresses = list(object({
        # The host's name or IP. For the SRV format, the SRV record's domain.
        hostname = string

        # The host's port. Leave empty for the SRV format (the record carries
        # it).
        port = optional(number, 0)
      }))

      # The database user.
      username = string

      # The user's password, stored by Google and never returned. Mutually
      # exclusive with secret_manager_stored_password.
      password = optional(string, "")

      # The Secret Manager secret VERSION holding the password -- a
      # GcpSecretManagerSecret reference (its latest_version_name output) or a
      # literal version name. Mutually exclusive with password.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      secret_manager_stored_password = optional(string, "")

      # The replica set's name -- needed for a self-hosted replica set with the
      # standard format; must be empty with the SRV format.
      replica_set = optional(string, "")

      # Connect through a DNS SRV seed list (mongodb+srv://). Google's SRV
      # block has no settings: true declares it.
      srv_connection_format = optional(bool, false)

      # Connect through a standard mongodb:// URI to the listed hosts.
      standard_connection_format = optional(object({
        # Connect straight to the one listed host instead of discovering the
        # replica set. Google now recommends additional_options
        # {"directConnection": "true"} instead.
        direct_connection = optional(bool, false)
      }))

      # TLS for the session.
      ssl_config = optional(object({
        # PEM certificate of the CA that signed the server's certificate.
        ca_certificate = optional(string, "")

        # PEM certificate Datastream presents to the server.
        client_certificate = optional(string, "")

        # PEM private key of client_certificate. Mutually exclusive with
        # secret_manager_stored_client_key.
        client_key = optional(string, "")

        # The Secret Manager secret VERSION holding the client key -- a
        # GcpSecretManagerSecret reference (its latest_version_name output) or a
        # literal version name. Mutually exclusive with client_key.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        secret_manager_stored_client_key = optional(string, "")
      }))

      # Extra MongoDB connection-string options, keyed exactly as MongoDB
      # names them, e.g. {"serverSelectionTimeoutMS": "10000"}.
      additional_options = optional(map(string), {})
    }))

    # Reach the source through a private connection -- a
    # GcpDatastreamPrivateConnection reference (its name output) or a literal
    # projects/{project}/locations/{location}/privateConnections/{id} in the
    # profile's location. Mutually exclusive with forward_ssh_connectivity.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    private_connection = optional(string, "")

    # Reach the source through an SSH bastion. Mutually exclusive with
    # private_connection.
    forward_ssh_connectivity = optional(object({
      # The bastion's hostname or IP.
      hostname = string

      # The bastion's SSH port. Empty uses 22.
      port = optional(number, 0)

      # The SSH user.
      username = string

      # The SSH password. Mutually exclusive with private_key. Immutable.
      password = optional(string, "")

      # The SSH private key, PEM. Mutually exclusive with password; the
      # stronger choice.
      private_key = optional(string, "")
    }))

    # What happens to the profile when this resource is destroyed:
    #   "" / "DELETE" -- deleted (Google refuses while a stream uses it)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- it leaves management and stays in GCP
    deletion_policy = optional(string, "")
  })
}
