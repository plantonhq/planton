variable "metadata" {
  description = "Catalog object metadata"
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
  description = "KubernetesLangfuse specification"
  type = object({
    # Namespace to install into. Accepts a literal namespace name or a
    # reference to a KubernetesNamespace resource. Every Secret the spec
    # references by name must live in this namespace.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    namespace = string

    # When true, the namespace is created (with the standard Planton
    # governance labels) before installing and deleted with the resource.
    # When false, the namespace must already exist.
    create_namespace = optional(bool, false)

    # Helm chart version to install (e.g. "2.1.5" = Langfuse 4.56.0; each
    # chart release pins one Langfuse version as its appVersion). Versions
    # must exist in the SERVED index at https://langfuse.github.io/langfuse-k8s.
    # Chart 2.x needs Kubernetes 1.28 or newer.
    chart_version = optional(string)

    # The public URL people open Langfuse at (`NEXTAUTH_URL`), e.g.
    # "https://langfuse.example.com": the base of sign-in redirects and of
    # every link Langfuse writes into emails, Slack messages and integration
    # settings. Scheme, host and optional port only - Langfuse serves from
    # the root of its host. The default suits a port-forward
    # (`port_forward_command` output); set the URL your KubernetesHttpRoute
    # or KubernetesIngress serves before enabling single sign-on.
    external_url = optional(string)

    # The Postgres database: users, organizations, projects, prompts, API
    # keys and settings.
    postgres = object({
      # Postgres host. Defaults compose a KubernetesPostgres resource's
      # read-write Service (always the current primary); any reachable
      # PostgreSQL 15+ host works as a literal (RDS, Aurora, Cloud SQL,
      # Azure Database for PostgreSQL).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      host = string

      # Postgres port.
      port = optional(number)

      # Database name. Create it before Langfuse starts (a KubernetesPostgres
      # bootstrap `database` with this user as owner does).
      database = optional(string)

      # The user Langfuse connects as (a KubernetesPostgres bootstrap owner,
      # for example).
      username = string

      # The Secret holding that user's password, in Langfuse's namespace.
      # Defaults compose a KubernetesPostgres resource's application
      # credential Secret, which the operator keeps current across failovers.
      password_secret = object({
        # Name of the Secret. Defaults compose a KubernetesPostgres resource's
        # application credential Secret (`<name>-app`).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        secret_name = string

        # Key within the Secret.
        secret_key = optional(string)
      })

      # Extra connection parameters appended to the connection URL (e.g.
      # "sslmode=require&connection_limit=20").
      args = optional(string, "")

      # A separate connection URL for schema migrations (`DIRECT_URL`), when
      # the main connection goes through a pooler such as PgBouncer or uses a
      # user without DDL rights. Contains the credentials, so it is written to
      # the instance Secret. Give this user long statement timeouts: large
      # migrations take a while.
      direct_url = optional(string, "")

      # Connection URL of a shadow database (`SHADOW_DATABASE_URL`), needed
      # only when the user lacks CREATE DATABASE (common on managed
      # services) and migrations must diff against a scratch database.
      # Contains the credentials, so it is written to the instance Secret.
      shadow_database_url = optional(string, "")

      # Run Postgres migrations automatically when the web server starts.
      # Turn off only when migrations run out of band; Langfuse refuses to
      # serve a schema it does not expect.
      auto_migrate = optional(bool)
    })

    # The ClickHouse database: traces, observations, scores and sessions.
    clickhouse = object({
      # ClickHouse host. Defaults compose a KubernetesClickHouse resource's
      # client Service; ClickHouse Cloud or any reachable server works as a
      # literal host name (no scheme - `secure` selects TLS).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      host = string

      # HTTP interface port (queries). 8443 is the usual TLS port.
      http_port = optional(number)

      # Native protocol port (migrations). 9440 is the usual TLS port.
      native_port = optional(number)

      # Connect over TLS on both interfaces (https for queries, TLS for
      # migrations) - ClickHouse Cloud requires it. Set the TLS ports too.
      secure = optional(bool, false)

      # Database name.
      database = optional(string)

      # The user Langfuse connects as (a KubernetesClickHouse user, for
      # example). On a KubernetesClickHouse, declare that user's `networks`
      # (e.g. "0.0.0.0/0" and "::/0"): a user declared without networks is
      # fenced to ClickHouse's own pods and Langfuse is refused with what
      # looks like a wrong-password error.
      username = string

      # The Secret holding that user's password, in Langfuse's namespace.
      password_secret = object({
        # Name of the Secret. Defaults compose a KubernetesClickHouse
        # resource's auth Secret (`<name>-clickhouse-auth`, one key per user).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        secret_name = string

        # Key within the Secret. Empty = `clickhouse.username`, the key a
        # KubernetesClickHouse auth Secret stores each user's password under.
        secret_key = optional(string, "")
      })

      # Cluster mode: whether Langfuse runs its DDL `ON CLUSTER` with
      # replicated table engines.
      cluster = optional(object({
        # Run DDL `ON CLUSTER` with replicated table engines
        # (`CLICKHOUSE_CLUSTER_ENABLED`). Needs ClickHouse Keeper (or ZooKeeper)
        # coordination - a KubernetesClickHouse with replicas, or with a
        # managed Keeper. Turn off for a single server without coordination:
        # cluster-mode migrations fail there. Decide before first start;
        # switching later needs a data migration.
        enabled = optional(bool)

        # The cluster's name (`CLICKHOUSE_CLUSTER_NAME`). Defaults compose a
        # KubernetesClickHouse resource's logical cluster name; empty =
        # "default". Langfuse renders its migrations for any name.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        name = optional(string, "")
      }))

      # A read-only endpoint for UI and public-API reads
      # (`CLICKHOUSE_READ_ONLY_URL`, e.g. "https://reader.example.com:8443"):
      # a separate compute group on ClickHouse Cloud or BYOC. Reuses the user,
      # password and database. No benefit on a single server.
      read_only_url = optional(string, "")

      # Run ClickHouse migrations automatically when the web server starts.
      auto_migrate = optional(bool)

      # How Langfuse issues deletes (`CLICKHOUSE_LIGHTWEIGHT_DELETE_MODE`).
      # Unspecified = Langfuse's default, ALTER TABLE ... DELETE mutations.
      # The lightweight modes need ClickHouse 25.7+ and cut background merge
      # work on deletion-heavy workloads (retention, project deletes).
      delete_mode = optional(string, "")

      # Route tracing-table updates through native UPDATE statements instead
      # of ALTER TABLE ... UPDATE mutations (`CLICKHOUSE_USE_LIGHTWEIGHT_UPDATE`,
      # ClickHouse 25.7+).
      use_lightweight_update = optional(bool, false)

      # Client-side timeout for delete operations in milliseconds
      # (`LANGFUSE_CLICKHOUSE_DELETION_TIMEOUT_MS`). Empty = Langfuse's
      # default (600000, ten minutes). Raise it when retention jobs or
      # project deletes time out.
      deletion_timeout_ms = optional(number)
    })

    # The Redis or Valkey store behind the ingestion queues and caches.
    redis = object({
      # One primary (with or without replicas behind it) - a
      # KubernetesValkey, ElastiCache without cluster mode, Memorystore.
      standalone = optional(object({
        # Host. Defaults compose a KubernetesValkey resource's primary Service;
        # any reachable Redis-protocol host works as a literal.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        host = string

        # Port.
        port = optional(number)

        # Logical database number.
        database = optional(number, 0)
      }))

      # Redis Cluster - ElastiCache with cluster mode, Azure Managed Redis
      # with the OSS clustering policy.
      cluster = optional(object({
        # Seed nodes as "host:port" (e.g. "redis-0.redis:6379"); the client
        # discovers the rest of the cluster from them.
        nodes = list(string)

        # Timeout in milliseconds for refreshing the cluster's slot map
        # (`REDIS_CLUSTER_SLOTS_REFRESH_TIMEOUT`). Empty = Langfuse's default
        # (5000). Raise it on timeouts during topology changes.
        slots_refresh_timeout_ms = optional(number)
      }))

      # Redis Sentinel - a primary discovered through Sentinel nodes.
      sentinel = optional(object({
        # Name of the monitored primary, exactly as the Sentinels know it.
        master_name = string

        # Sentinel nodes as "host:port" (e.g. "sentinel-0.sentinel:26379").
        nodes = list(string)

        # Username the Sentinels require, when they authenticate.
        username = optional(string, "")

        # The Secret holding the Sentinels' password, in Langfuse's namespace.
        # Unset = auth-less Sentinels.
        password_secret = optional(object({
          # The name of the Kubernetes Secret.
          name = optional(string, "")

          # The key within the Kubernetes Secret.
          key = optional(string, "")
        }))

        # Also use TLS to the Sentinels, not just to the primary and replicas.
        # Requires `tls.enabled`.
        tls_enabled = optional(bool, false)
      }))

      # ACL username, when the store uses named users. Defaults compose a
      # KubernetesValkey resource's ACL user. Empty = password-only AUTH.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      username = optional(string, "")

      # The Secret holding the password, in Langfuse's namespace. Unset = an
      # auth-less store.
      password_secret = optional(object({
        # Name of the Secret. Defaults compose a KubernetesValkey resource's
        # auth Secret.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        secret_name = string

        # Key within the Secret. Defaults compose a KubernetesValkey resource's
        # password key (its auth Secret stores each user's password under the
        # user's name).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        secret_key = string
      }))

      # TLS to Redis (and, for Sentinel, optionally to the Sentinels).
      tls = optional(object({
        # Connect over TLS (`rediss://`).
        enabled = optional(bool, false)

        # The CA that signed the server certificate, when it is not a public
        # CA. Mounted into both workloads.
        ca = optional(object({
          # Name of a Secret holding the bundle.
          secret_name = optional(string)

          # Name of a ConfigMap holding the bundle.
          config_map_name = optional(string)

          # Key holding the PEM bundle.
          key = optional(string)
        }))

        # Client certificate for mutual TLS. Empty key = "tls.crt".
        client_certificate = optional(object({
          # Name of the Secret. A KubernetesCertificate's Secret composes
          # naturally.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          secret_name = string

          # Key within the Secret. Empty = the conventional key for the file
          # (named on the field that holds this message).
          key = optional(string, "")
        }))

        # Client private key for mutual TLS. Empty key = "tls.key".
        client_key = optional(object({
          # Name of the Secret. A KubernetesCertificate's Secret composes
          # naturally.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          secret_name = string

          # Key within the Secret. Empty = the conventional key for the file
          # (named on the field that holds this message).
          key = optional(string, "")
        }))

        # Passphrase of an encrypted client private key.
        client_key_passphrase = optional(string, "")

        # Server name for SNI and certificate verification, when it differs
        # from the host (`REDIS_TLS_SERVERNAME`).
        server_name = optional(string, "")

        # Reject a server certificate that does not chain to a trusted CA
        # (`REDIS_TLS_REJECT_UNAUTHORIZED`). Empty = Node.js's default, true.
        # Turning it off accepts any certificate - never in production.
        reject_unauthorized = optional(bool)

        # Check that the server certificate names the host
        # (`REDIS_TLS_CHECK_SERVER_IDENTITY`). Empty = Node.js's default, true.
        # Some managed services present certificates for another name.
        check_server_identity = optional(bool)

        # Allowed TLS cipher suites, OpenSSL format (`REDIS_TLS_CIPHERS`).
        ciphers = optional(string, "")

        # TLS protocol method, e.g. "TLSv1_2_method" (`REDIS_TLS_SECURE_PROTOCOL`).
        secure_protocol = optional(string, "")

        # Prefer the client's cipher order (`REDIS_TLS_HONOR_CIPHER_ORDER`).
        honor_cipher_order = optional(bool)
      }))

      # Prefix for every key Langfuse writes (`REDIS_KEY_PREFIX`), to share
      # one store between several Langfuse instances. Changing it on a live
      # instance orphans every queued job.
      key_prefix = optional(string, "")

      # Socket timeout in milliseconds for every Redis connection
      # (`REDIS_SOCKET_TIMEOUT_MS`). Empty = Langfuse's default (30000). 0
      # disables it; a positive value must be at least 10000, because queue
      # workers block on the socket between jobs.
      socket_timeout_ms = optional(number)
    })

    # Object storage for ingested events, media attachments and batch
    # exports.
    blob_storage = object({
      # Where events are stored - and media and exports, unless they declare
      # a target of their own.
      target = object({
        # Any S3-compatible store: a KubernetesSeaweedFs, MinIO, OCI Object
        # Storage, Backblaze B2.
        s3_compatible = optional(object({
          # The S3 API endpoint (e.g.
          # "http://seaweedfs-s3.storage.svc.cluster.local:8333"). Defaults
          # compose a KubernetesSeaweedFs resource's S3 endpoint.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          endpoint = string

          # Bucket name. It must exist before Langfuse starts (a
          # KubernetesSeaweedFs creates the buckets its spec declares).
          bucket = string

          # Region the store expects in request signatures.
          region = optional(string)

          # Address buckets by path (`endpoint/bucket/key`) rather than by
          # virtual host - required by SeaweedFS and MinIO.
          force_path_style = optional(bool)

          # An existing Secret in Langfuse's namespace holding the key pair.
          # Defaults compose a KubernetesSeaweedFs resource's S3 credentials
          # Secret.
          credentials_secret = optional(object({
            # Name of the Secret.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            secret_name = string

            # Key holding the access key id.
            access_key_id_key = optional(string)

            # Key holding the secret access key.
            secret_access_key_key = optional(string)
          }))

          # The key pair itself, written to the instance Secret.
          access_key = optional(object({
            # Access key id.
            access_key_id = string

            # Secret access key.
            secret_access_key = string
          }))
        }))

        # AWS S3.
        s3 = optional(object({
          # Bucket name. Defaults compose an AwsS3Bucket resource.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          bucket = string

          # The bucket's region. Defaults compose an AwsS3Bucket resource's
          # region.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          region = string

          # Endpoint override - a FIPS or VPC interface endpoint. Empty = the
          # regional S3 endpoint.
          endpoint = optional(string, "")

          # A static key pair. Empty = the AWS credential chain: an IAM role
          # through `service_account.workload_identity` (EKS IRSA) - the
          # recommended keyless posture - or the node's instance profile.
          access_key = optional(object({
            # Access key id.
            access_key_id = string

            # Secret access key.
            secret_access_key = string
          }))
        }))

        # Cloudflare R2.
        r2 = optional(object({
          # The bucket's S3 API endpoint (it encodes the account and the
          # jurisdiction). Defaults compose a CloudflareR2Bucket resource's
          # endpoint.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          endpoint = string

          # Bucket name. Defaults compose a CloudflareR2Bucket resource.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          bucket = string

          # The key pair of an R2 API token scoped to the bucket (Object Read and
          # Write).
          credentials = object({
            # The S3 access key id: the API token's id. By reference to a
            # CloudflareAccountApiToken's `r2_access_key_id` output, or a `$secret/`
            # reference for a token minted in the dashboard.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            access_key_id = string

            # The S3 secret access key: the SHA-256 of the API token's value.
            # Reference-only.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            secret_access_key = string
          })
        }))

        # Google Cloud Storage, through Langfuse's native GCS driver (more
        # reliable than GCS's S3 interoperability mode, which mishandles
        # deletes).
        gcs = optional(object({
          # Bucket name. Defaults compose a GcpGcsBucket resource.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          bucket = string

          # A service account JSON key. Empty = application default credentials:
          # a Google service account through `service_account.workload_identity`
          # (GKE Workload Identity) - the recommended keyless posture. Every GCS
          # target of one Langfuse shares one credential.
          credentials_json = optional(string, "")
        }))

        # Azure Blob Storage, through Langfuse's native Azure driver.
        azure_blob = optional(object({
          # Storage account name. Defaults compose an AzureStorageAccount
          # resource.
          #
          # containment_exempt: Langfuse reads and writes the account - access,
          # never placement.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          storage_account = string

          # Container name. Defaults compose an AzureStorageContainer resource.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          container = string

          # The storage account's access key.
          account_key = string

          # Blob endpoint. Defaults compose an AzureStorageAccount resource's
          # primary blob endpoint; empty = https://<storage_account>.blob.core.windows.net.
          #
          # containment_exempt: an address Langfuse calls - access, never
          # placement.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          endpoint = optional(string, "")
        }))
      })

      # Event upload settings.
      event_upload = optional(object({
        # Key prefix for events (must end with "/"). A separate prefix lets a
        # lifecycle rule expire event files (about 30 days is safe once they
        # are in ClickHouse) without touching media.
        prefix = optional(string)

        # Server-side encryption (S3 targets).
        sse = optional(object({
          # Encryption algorithm.
          algorithm = optional(string, "")

          # KMS key for `aws_kms`. Defaults compose an AwsKmsKey resource's ARN;
          # empty = the bucket's default AWS-managed key.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          kms_key_id = optional(string, "")
        }))
      }))

      # Media attachments (images, audio, documents traced with LLM calls).
      media_upload = optional(object({
        # A target of its own (a publicly reachable bucket, for example). Unset
        # = the blob storage target.
        target = optional(object({
          # Any S3-compatible store: a KubernetesSeaweedFs, MinIO, OCI Object
          # Storage, Backblaze B2.
          s3_compatible = optional(object({
            # The S3 API endpoint (e.g.
            # "http://seaweedfs-s3.storage.svc.cluster.local:8333"). Defaults
            # compose a KubernetesSeaweedFs resource's S3 endpoint.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            endpoint = string

            # Bucket name. It must exist before Langfuse starts (a
            # KubernetesSeaweedFs creates the buckets its spec declares).
            bucket = string

            # Region the store expects in request signatures.
            region = optional(string)

            # Address buckets by path (`endpoint/bucket/key`) rather than by
            # virtual host - required by SeaweedFS and MinIO.
            force_path_style = optional(bool)

            # An existing Secret in Langfuse's namespace holding the key pair.
            # Defaults compose a KubernetesSeaweedFs resource's S3 credentials
            # Secret.
            credentials_secret = optional(object({
              # Name of the Secret.
              # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
              secret_name = string

              # Key holding the access key id.
              access_key_id_key = optional(string)

              # Key holding the secret access key.
              secret_access_key_key = optional(string)
            }))

            # The key pair itself, written to the instance Secret.
            access_key = optional(object({
              # Access key id.
              access_key_id = string

              # Secret access key.
              secret_access_key = string
            }))
          }))

          # AWS S3.
          s3 = optional(object({
            # Bucket name. Defaults compose an AwsS3Bucket resource.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            bucket = string

            # The bucket's region. Defaults compose an AwsS3Bucket resource's
            # region.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            region = string

            # Endpoint override - a FIPS or VPC interface endpoint. Empty = the
            # regional S3 endpoint.
            endpoint = optional(string, "")

            # A static key pair. Empty = the AWS credential chain: an IAM role
            # through `service_account.workload_identity` (EKS IRSA) - the
            # recommended keyless posture - or the node's instance profile.
            access_key = optional(object({
              # Access key id.
              access_key_id = string

              # Secret access key.
              secret_access_key = string
            }))
          }))

          # Cloudflare R2.
          r2 = optional(object({
            # The bucket's S3 API endpoint (it encodes the account and the
            # jurisdiction). Defaults compose a CloudflareR2Bucket resource's
            # endpoint.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            endpoint = string

            # Bucket name. Defaults compose a CloudflareR2Bucket resource.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            bucket = string

            # The key pair of an R2 API token scoped to the bucket (Object Read and
            # Write).
            credentials = object({
              # The S3 access key id: the API token's id. By reference to a
              # CloudflareAccountApiToken's `r2_access_key_id` output, or a `$secret/`
              # reference for a token minted in the dashboard.
              # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
              access_key_id = string

              # The S3 secret access key: the SHA-256 of the API token's value.
              # Reference-only.
              # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
              secret_access_key = string
            })
          }))

          # Google Cloud Storage, through Langfuse's native GCS driver (more
          # reliable than GCS's S3 interoperability mode, which mishandles
          # deletes).
          gcs = optional(object({
            # Bucket name. Defaults compose a GcpGcsBucket resource.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            bucket = string

            # A service account JSON key. Empty = application default credentials:
            # a Google service account through `service_account.workload_identity`
            # (GKE Workload Identity) - the recommended keyless posture. Every GCS
            # target of one Langfuse shares one credential.
            credentials_json = optional(string, "")
          }))

          # Azure Blob Storage, through Langfuse's native Azure driver.
          azure_blob = optional(object({
            # Storage account name. Defaults compose an AzureStorageAccount
            # resource.
            #
            # containment_exempt: Langfuse reads and writes the account - access,
            # never placement.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            storage_account = string

            # Container name. Defaults compose an AzureStorageContainer resource.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            container = string

            # The storage account's access key.
            account_key = string

            # Blob endpoint. Defaults compose an AzureStorageAccount resource's
            # primary blob endpoint; empty = https://<storage_account>.blob.core.windows.net.
            #
            # containment_exempt: an address Langfuse calls - access, never
            # placement.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            endpoint = optional(string, "")
          }))
        }))

        # Key prefix for media (must end with "/"). Put no expiry lifecycle rule
        # on it: traces keep pointing at their media.
        prefix = optional(string)

        # The endpoint browsers and SDKs reach the store at, when the target's
        # endpoint is cluster-internal (e.g. "https://s3.langfuse.example.com"
        # for an in-cluster SeaweedFS behind a route). Langfuse signs URLs for
        # this endpoint and reads media itself through the target's endpoint.
        # Empty = the target's endpoint serves both.
        public_endpoint = optional(string, "")

        # Largest file accepted, in bytes (`LANGFUSE_S3_MEDIA_MAX_CONTENT_LENGTH`).
        max_content_length_bytes = optional(number)

        # Lifetime of presigned download URLs in seconds.
        download_url_expiry_seconds = optional(number)

        # Server-side encryption (S3 targets).
        sse = optional(object({
          # Encryption algorithm.
          algorithm = optional(string, "")

          # KMS key for `aws_kms`. Defaults compose an AwsKmsKey resource's ARN;
          # empty = the bucket's default AWS-managed key.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          kms_key_id = optional(string, "")
        }))
      }))

      # Batch exports (CSV and JSON downloads of traces, observations and
      # scores).
      batch_export = optional(object({
        # Offer batch exports.
        enabled = optional(bool)

        # A target of its own. Unset = the blob storage target.
        target = optional(object({
          # Any S3-compatible store: a KubernetesSeaweedFs, MinIO, OCI Object
          # Storage, Backblaze B2.
          s3_compatible = optional(object({
            # The S3 API endpoint (e.g.
            # "http://seaweedfs-s3.storage.svc.cluster.local:8333"). Defaults
            # compose a KubernetesSeaweedFs resource's S3 endpoint.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            endpoint = string

            # Bucket name. It must exist before Langfuse starts (a
            # KubernetesSeaweedFs creates the buckets its spec declares).
            bucket = string

            # Region the store expects in request signatures.
            region = optional(string)

            # Address buckets by path (`endpoint/bucket/key`) rather than by
            # virtual host - required by SeaweedFS and MinIO.
            force_path_style = optional(bool)

            # An existing Secret in Langfuse's namespace holding the key pair.
            # Defaults compose a KubernetesSeaweedFs resource's S3 credentials
            # Secret.
            credentials_secret = optional(object({
              # Name of the Secret.
              # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
              secret_name = string

              # Key holding the access key id.
              access_key_id_key = optional(string)

              # Key holding the secret access key.
              secret_access_key_key = optional(string)
            }))

            # The key pair itself, written to the instance Secret.
            access_key = optional(object({
              # Access key id.
              access_key_id = string

              # Secret access key.
              secret_access_key = string
            }))
          }))

          # AWS S3.
          s3 = optional(object({
            # Bucket name. Defaults compose an AwsS3Bucket resource.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            bucket = string

            # The bucket's region. Defaults compose an AwsS3Bucket resource's
            # region.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            region = string

            # Endpoint override - a FIPS or VPC interface endpoint. Empty = the
            # regional S3 endpoint.
            endpoint = optional(string, "")

            # A static key pair. Empty = the AWS credential chain: an IAM role
            # through `service_account.workload_identity` (EKS IRSA) - the
            # recommended keyless posture - or the node's instance profile.
            access_key = optional(object({
              # Access key id.
              access_key_id = string

              # Secret access key.
              secret_access_key = string
            }))
          }))

          # Cloudflare R2.
          r2 = optional(object({
            # The bucket's S3 API endpoint (it encodes the account and the
            # jurisdiction). Defaults compose a CloudflareR2Bucket resource's
            # endpoint.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            endpoint = string

            # Bucket name. Defaults compose a CloudflareR2Bucket resource.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            bucket = string

            # The key pair of an R2 API token scoped to the bucket (Object Read and
            # Write).
            credentials = object({
              # The S3 access key id: the API token's id. By reference to a
              # CloudflareAccountApiToken's `r2_access_key_id` output, or a `$secret/`
              # reference for a token minted in the dashboard.
              # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
              access_key_id = string

              # The S3 secret access key: the SHA-256 of the API token's value.
              # Reference-only.
              # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
              secret_access_key = string
            })
          }))

          # Google Cloud Storage, through Langfuse's native GCS driver (more
          # reliable than GCS's S3 interoperability mode, which mishandles
          # deletes).
          gcs = optional(object({
            # Bucket name. Defaults compose a GcpGcsBucket resource.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            bucket = string

            # A service account JSON key. Empty = application default credentials:
            # a Google service account through `service_account.workload_identity`
            # (GKE Workload Identity) - the recommended keyless posture. Every GCS
            # target of one Langfuse shares one credential.
            credentials_json = optional(string, "")
          }))

          # Azure Blob Storage, through Langfuse's native Azure driver.
          azure_blob = optional(object({
            # Storage account name. Defaults compose an AzureStorageAccount
            # resource.
            #
            # containment_exempt: Langfuse reads and writes the account - access,
            # never placement.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            storage_account = string

            # Container name. Defaults compose an AzureStorageContainer resource.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            container = string

            # The storage account's access key.
            account_key = string

            # Blob endpoint. Defaults compose an AzureStorageAccount resource's
            # primary blob endpoint; empty = https://<storage_account>.blob.core.windows.net.
            #
            # containment_exempt: an address Langfuse calls - access, never
            # placement.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            endpoint = optional(string, "")
          }))
        }))

        # Key prefix for exports (must end with "/"). Exports are disposable:
        # an expiry lifecycle rule here is safe.
        prefix = optional(string)

        # The endpoint download links point at, when the target's endpoint is
        # cluster-internal (`LANGFUSE_S3_BATCH_EXPORT_EXTERNAL_ENDPOINT`). Empty
        # = the target's endpoint.
        public_endpoint = optional(string, "")

        # Rows streamed per page while exporting (`BATCH_EXPORT_PAGE_SIZE`).
        # Empty = Langfuse's default (500). Lower it if exports exhaust worker
        # memory.
        page_size = optional(number)

        # Most rows one export may contain (`BATCH_EXPORT_ROW_LIMIT`). Empty =
        # Langfuse's default (1500000).
        row_limit = optional(number)

        # Hours a download link stays valid
        # (`BATCH_EXPORT_DOWNLOAD_LINK_EXPIRATION_HOURS`). Empty = Langfuse's
        # default (24).
        download_link_expiration_hours = optional(number)

        # Server-side encryption (S3 targets).
        sse = optional(object({
          # Encryption algorithm.
          algorithm = optional(string, "")

          # KMS key for `aws_kms`. Defaults compose an AwsKmsKey resource's ARN;
          # empty = the bucket's default AWS-managed key.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          kms_key_id = optional(string, "")
        }))
      }))

      # Maximum concurrent reads from object storage
      # (`LANGFUSE_S3_CONCURRENT_READS`). Raise it on "socket usage at
      # capacity" warnings.
      concurrent_reads = optional(number)

      # Maximum concurrent writes to object storage
      # (`LANGFUSE_S3_CONCURRENT_WRITES`).
      concurrent_writes = optional(number)

      # Checksum algorithm for S3 DeleteObjects requests
      # (`LANGFUSE_S3_DELETE_OBJECTS_CHECKSUM_ALGORITHM`). Unspecified = the
      # SDK's default (CRC32). Some S3-compatible stores reject the default
      # and need MD5.
      delete_objects_checksum_algorithm = optional(string, "")
    })

    # The instance's own secrets: the API-key salt, the encryption key and
    # the session secret. Generated when not declared.
    secrets = optional(object({
      # Salt for hashing API keys (`SALT`), at least 256 bits of entropy
      # (`openssl rand -base64 32`). Empty = generated.
      salt = optional(string, "")

      # Key for encrypting stored credentials (`ENCRYPTION_KEY`): exactly 64
      # hex characters, 256 bits (`openssl rand -hex 32`). Langfuse refuses to
      # start on any other shape. Empty = generated.
      encryption_key = optional(string, "")

      # Secret for signing session cookies (`NEXTAUTH_SECRET`). Empty =
      # generated. Changing it signs everyone out, nothing more.
      nextauth_secret = optional(string, "")
    }))

    # The organization, project, project key pair and admin user Langfuse
    # creates at first start (headless initialization). Seeded by default.
    init = optional(object({
      # Seed the organization and project at first start. Turn off to start
      # with an empty instance where the first sign-up creates everything.
      enabled = optional(bool)

      # The seeded organization.
      organization = optional(object({
        # Organization ID. Empty = metadata.name. Fixed at first start.
        id = optional(string, "")

        # Organization display name.
        name = optional(string)
      }))

      # The seeded project and its API key pair.
      project = optional(object({
        # Project ID. Empty = "<metadata.name>-project". Fixed at first start.
        id = optional(string, "")

        # Project display name.
        name = optional(string)

        # Data retention in days (3 or more). Applied only with an Enterprise
        # Edition license that includes data retention; otherwise data is kept
        # indefinitely. Empty = keep indefinitely.
        retention_days = optional(number)

        # Public key of the project's API key pair ("pk-lf-..."). Empty =
        # generated ("pk-lf-" + 32 random letters and digits).
        public_key = optional(string, "")

        # Secret key of the project's API key pair. It must start with
        # "sk-lf-", the prefix Langfuse's SDKs expect. Empty = generated
        # ("sk-lf-" + 32 random letters and digits).
        secret_key = optional(string, "")
      }))

      # The seeded admin user, owner of the organization. Unset = no user is
      # seeded (sign up, or sign in with single sign-on, instead).
      admin = optional(object({
        # Email address, the sign-in name.
        email = string

        # Display name.
        name = optional(string)

        # Password. Empty = generated (32 letters and digits) and readable from
        # the `admin_password_secret` output.
        password = optional(string, "")
      }))
    }))

    # Sign-in: email and password, sign-up, sessions, the access new users
    # receive, and single sign-on providers.
    auth = optional(object({
      # Disable email and password sign-in (`AUTH_DISABLE_USERNAME_PASSWORD`)
      # - single sign-on only. Keep at least one SSO provider configured.
      disable_password_sign_in = optional(bool, false)

      # Disable sign-up for new users, invited ones included
      # (`AUTH_DISABLE_SIGNUP`). Recommended for any exposed instance once
      # the admin exists.
      disable_sign_up = optional(bool, false)

      # Require new email-and-password users to verify their address
      # (`AUTH_EMAIL_VERIFICATION_REQUIRED`). Needs `smtp`.
      email_verification_required = optional(bool, false)

      # Session lifetime in minutes (`AUTH_SESSION_MAX_AGE`). Empty =
      # Langfuse's default (20160, fourteen days). Must exceed 5: sessions
      # refresh every five minutes.
      session_max_age_minutes = optional(number)

      # Email domains that may only sign in through single sign-on
      # (`AUTH_DOMAINS_WITH_SSO_ENFORCEMENT`), e.g. "example.com".
      sso_enforced_domains = optional(list(string), [])

      # Account fields to ignore from the identity provider when creating an
      # account (`AUTH_IGNORE_ACCOUNT_FIELDS`) - a fix for providers that
      # return fields Langfuse cannot store.
      ignore_account_fields = optional(list(string), [])

      # Proxy for requests to identity providers (`AUTH_HTTPS_PROXY`).
      https_proxy = optional(string, "")

      # Proxy for plain-HTTP requests to identity providers (`AUTH_HTTP_PROXY`).
      http_proxy = optional(string, "")

      # Organization and project memberships every new user receives.
      new_user_access = optional(object({
        # Organization every new user joins (`LANGFUSE_DEFAULT_ORG_ID`) - the
        # seeded organization's ID, for example.
        organization_id = optional(string, "")

        # Their role there (`LANGFUSE_DEFAULT_ORG_ROLE`). Unspecified =
        # Langfuse's default, VIEWER.
        organization_role = optional(string, "")

        # Projects every new user joins (`LANGFUSE_DEFAULT_PROJECT_ID`).
        project_ids = optional(list(string), [])

        # Their role there (`LANGFUSE_DEFAULT_PROJECT_ROLE`). Unspecified =
        # Langfuse's default, VIEWER. NONE is not a project role.
        project_role = optional(string, "")
      }))

      # Single sign-on providers. Each one configured becomes a sign-in
      # button; its redirect URI is `<external_url>/api/auth/callback/<id>`.
      sso = optional(object({
        # Google.
        google = optional(object({
          # The OAuth client.
          client = object({
            # Client ID.
            client_id = string

            # Client secret.
            client_secret = string

            # Merge a sign-in with an existing account of the same email address
            # (`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
            # password to SSO. Safe only when the provider verifies emails.
            allow_account_linking = optional(bool, false)

            # Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
            # Unspecified = client_secret_basic. Use `none` for PKCE.
            client_auth_method = optional(string, "")

            # Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
            # Empty = the provider's defaults.
            checks = optional(list(string), [])

            # Algorithm the provider signs ID tokens with
            # (`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.
            id_token_signed_response_alg = optional(string, "")
          })

          # Google Workspace domains allowed to sign in (`AUTH_GOOGLE_ALLOWED_DOMAINS`).
          # Empty = any Google account.
          allowed_domains = optional(list(string), [])
        }))

        # GitHub.
        github = optional(object({
          # The OAuth app.
          client = object({
            # Client ID.
            client_id = string

            # Client secret.
            client_secret = string

            # Merge a sign-in with an existing account of the same email address
            # (`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
            # password to SSO. Safe only when the provider verifies emails.
            allow_account_linking = optional(bool, false)

            # Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
            # Unspecified = client_secret_basic. Use `none` for PKCE.
            client_auth_method = optional(string, "")

            # Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
            # Empty = the provider's defaults.
            checks = optional(list(string), [])

            # Algorithm the provider signs ID tokens with
            # (`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.
            id_token_signed_response_alg = optional(string, "")
          })
        }))

        # GitHub Enterprise Server.
        github_enterprise = optional(object({
          # The OAuth app.
          client = object({
            # Client ID.
            client_id = string

            # Client secret.
            client_secret = string

            # Merge a sign-in with an existing account of the same email address
            # (`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
            # password to SSO. Safe only when the provider verifies emails.
            allow_account_linking = optional(bool, false)

            # Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
            # Unspecified = client_secret_basic. Use `none` for PKCE.
            client_auth_method = optional(string, "")

            # Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
            # Empty = the provider's defaults.
            checks = optional(list(string), [])

            # Algorithm the provider signs ID tokens with
            # (`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.
            id_token_signed_response_alg = optional(string, "")
          })

          # Base URL of the server (`AUTH_GITHUB_ENTERPRISE_BASE_URL`), e.g.
          # "https://github.example.com".
          base_url = optional(string, "")
        }))

        # GitLab (gitlab.com or self-managed).
        gitlab = optional(object({
          # The OAuth application.
          client = object({
            # Client ID.
            client_id = string

            # Client secret.
            client_secret = string

            # Merge a sign-in with an existing account of the same email address
            # (`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
            # password to SSO. Safe only when the provider verifies emails.
            allow_account_linking = optional(bool, false)

            # Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
            # Unspecified = client_secret_basic. Use `none` for PKCE.
            client_auth_method = optional(string, "")

            # Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
            # Empty = the provider's defaults.
            checks = optional(list(string), [])

            # Algorithm the provider signs ID tokens with
            # (`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.
            id_token_signed_response_alg = optional(string, "")
          })

          # Base URL of a self-managed GitLab (`AUTH_GITLAB_URL`). Empty =
          # https://gitlab.com.
          url = optional(string, "")

          # OIDC issuer, when it differs from `url` (`AUTH_GITLAB_ISSUER`).
          issuer = optional(string, "")
        }))

        # Microsoft Entra ID (Azure AD).
        azure_ad = optional(object({
          # The app registration.
          client = object({
            # Client ID.
            client_id = string

            # Client secret.
            client_secret = string

            # Merge a sign-in with an existing account of the same email address
            # (`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
            # password to SSO. Safe only when the provider verifies emails.
            allow_account_linking = optional(bool, false)

            # Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
            # Unspecified = client_secret_basic. Use `none` for PKCE.
            client_auth_method = optional(string, "")

            # Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
            # Empty = the provider's defaults.
            checks = optional(list(string), [])

            # Algorithm the provider signs ID tokens with
            # (`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.
            id_token_signed_response_alg = optional(string, "")
          })

          # Directory (tenant) ID (`AUTH_AZURE_AD_TENANT_ID`).
          tenant_id = string
        }))

        # Okta.
        okta = optional(object({
          # The OAuth client.
          client = object({
            # Client ID.
            client_id = string

            # Client secret.
            client_secret = string

            # Merge a sign-in with an existing account of the same email address
            # (`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
            # password to SSO. Safe only when the provider verifies emails.
            allow_account_linking = optional(bool, false)

            # Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
            # Unspecified = client_secret_basic. Use `none` for PKCE.
            client_auth_method = optional(string, "")

            # Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
            # Empty = the provider's defaults.
            checks = optional(list(string), [])

            # Algorithm the provider signs ID tokens with
            # (`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.
            id_token_signed_response_alg = optional(string, "")
          })

          # Issuer URL, e.g. "https://example.okta.com" or, for Cognito,
          # "https://cognito-idp.<region>.amazonaws.com/<user-pool-id>".
          issuer = string
        }))

        # Authentik.
        authentik = optional(object({
          # The OAuth provider.
          client = object({
            # Client ID.
            client_id = string

            # Client secret.
            client_secret = string

            # Merge a sign-in with an existing account of the same email address
            # (`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
            # password to SSO. Safe only when the provider verifies emails.
            allow_account_linking = optional(bool, false)

            # Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
            # Unspecified = client_secret_basic. Use `none` for PKCE.
            client_auth_method = optional(string, "")

            # Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
            # Empty = the provider's defaults.
            checks = optional(list(string), [])

            # Algorithm the provider signs ID tokens with
            # (`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.
            id_token_signed_response_alg = optional(string, "")
          })

          # Issuer, "https://<domain>/application/o/<slug>" (no trailing path
          # beyond the slug).
          issuer = string

          # Authorization URL, when it differs from the issuer's
          # (`AUTH_AUTHENTIK_AUTHORIZATION_URL`).
          authorization_url = optional(string, "")
        }))

        # OneLogin.
        onelogin = optional(object({
          # The OAuth client.
          client = object({
            # Client ID.
            client_id = string

            # Client secret.
            client_secret = string

            # Merge a sign-in with an existing account of the same email address
            # (`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
            # password to SSO. Safe only when the provider verifies emails.
            allow_account_linking = optional(bool, false)

            # Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
            # Unspecified = client_secret_basic. Use `none` for PKCE.
            client_auth_method = optional(string, "")

            # Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
            # Empty = the provider's defaults.
            checks = optional(list(string), [])

            # Algorithm the provider signs ID tokens with
            # (`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.
            id_token_signed_response_alg = optional(string, "")
          })

          # Issuer URL, e.g. "https://example.okta.com" or, for Cognito,
          # "https://cognito-idp.<region>.amazonaws.com/<user-pool-id>".
          issuer = string
        }))

        # Auth0.
        auth0 = optional(object({
          # The OAuth client.
          client = object({
            # Client ID.
            client_id = string

            # Client secret.
            client_secret = string

            # Merge a sign-in with an existing account of the same email address
            # (`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
            # password to SSO. Safe only when the provider verifies emails.
            allow_account_linking = optional(bool, false)

            # Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
            # Unspecified = client_secret_basic. Use `none` for PKCE.
            client_auth_method = optional(string, "")

            # Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
            # Empty = the provider's defaults.
            checks = optional(list(string), [])

            # Algorithm the provider signs ID tokens with
            # (`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.
            id_token_signed_response_alg = optional(string, "")
          })

          # Issuer URL, e.g. "https://example.okta.com" or, for Cognito,
          # "https://cognito-idp.<region>.amazonaws.com/<user-pool-id>".
          issuer = string
        }))

        # Amazon Cognito.
        cognito = optional(object({
          # The OAuth client.
          client = object({
            # Client ID.
            client_id = string

            # Client secret.
            client_secret = string

            # Merge a sign-in with an existing account of the same email address
            # (`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
            # password to SSO. Safe only when the provider verifies emails.
            allow_account_linking = optional(bool, false)

            # Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
            # Unspecified = client_secret_basic. Use `none` for PKCE.
            client_auth_method = optional(string, "")

            # Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
            # Empty = the provider's defaults.
            checks = optional(list(string), [])

            # Algorithm the provider signs ID tokens with
            # (`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.
            id_token_signed_response_alg = optional(string, "")
          })

          # Issuer URL, e.g. "https://example.okta.com" or, for Cognito,
          # "https://cognito-idp.<region>.amazonaws.com/<user-pool-id>".
          issuer = string
        }))

        # Keycloak.
        keycloak = optional(object({
          # The client.
          client = object({
            # Client ID.
            client_id = string

            # Client secret.
            client_secret = string

            # Merge a sign-in with an existing account of the same email address
            # (`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
            # password to SSO. Safe only when the provider verifies emails.
            allow_account_linking = optional(bool, false)

            # Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
            # Unspecified = client_secret_basic. Use `none` for PKCE.
            client_auth_method = optional(string, "")

            # Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
            # Empty = the provider's defaults.
            checks = optional(list(string), [])

            # Algorithm the provider signs ID tokens with
            # (`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.
            id_token_signed_response_alg = optional(string, "")
          })

          # Realm issuer, e.g. "https://keycloak.example.com/realms/langfuse".
          issuer = string

          # Button label (`AUTH_KEYCLOAK_NAME`). Empty = "Keycloak".
          name = optional(string, "")

          # OAuth scopes (`AUTH_KEYCLOAK_SCOPE`). Empty = "openid email profile".
          scope = optional(string, "")

          # Read the user from the ID token (`AUTH_KEYCLOAK_ID_TOKEN`). Empty =
          # true; false fetches the userinfo endpoint instead.
          use_id_token = optional(bool)
        }))

        # JumpCloud.
        jumpcloud = optional(object({
          # The SSO application.
          client = object({
            # Client ID.
            client_id = string

            # Client secret.
            client_secret = string

            # Merge a sign-in with an existing account of the same email address
            # (`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
            # password to SSO. Safe only when the provider verifies emails.
            allow_account_linking = optional(bool, false)

            # Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
            # Unspecified = client_secret_basic. Use `none` for PKCE.
            client_auth_method = optional(string, "")

            # Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
            # Empty = the provider's defaults.
            checks = optional(list(string), [])

            # Algorithm the provider signs ID tokens with
            # (`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.
            id_token_signed_response_alg = optional(string, "")
          })

          # Issuer URL.
          issuer = string

          # OAuth scopes (`AUTH_JUMPCLOUD_SCOPE`).
          scope = optional(string, "")
        }))

        # Any OpenID Connect provider.
        custom = optional(object({
          # The OAuth client.
          client = object({
            # Client ID.
            client_id = string

            # Client secret.
            client_secret = string

            # Merge a sign-in with an existing account of the same email address
            # (`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
            # password to SSO. Safe only when the provider verifies emails.
            allow_account_linking = optional(bool, false)

            # Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
            # Unspecified = client_secret_basic. Use `none` for PKCE.
            client_auth_method = optional(string, "")

            # Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
            # Empty = the provider's defaults.
            checks = optional(list(string), [])

            # Algorithm the provider signs ID tokens with
            # (`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.
            id_token_signed_response_alg = optional(string, "")
          })

          # Issuer URL (its /.well-known/openid-configuration is discovered).
          issuer = string

          # Button label (`AUTH_CUSTOM_NAME`), e.g. "Company SSO".
          name = string

          # OAuth scopes (`AUTH_CUSTOM_SCOPE`). Empty = "openid email profile".
          scope = optional(string, "")

          # Read the user from the ID token (`AUTH_CUSTOM_ID_TOKEN`). Empty =
          # true.
          use_id_token = optional(bool)

          # Also fetch the userinfo endpoint (`AUTH_CUSTOM_FETCH_USERINFO`).
          fetch_userinfo = optional(bool, false)

          # Claim names, when the provider's differ from the standard ones.
          claims = optional(object({
            # Email claim (`LANGFUSE_CUSTOM_SSO_EMAIL_CLAIM`). Empty = "email".
            email = optional(string, "")

            # Name claim (`LANGFUSE_CUSTOM_SSO_NAME_CLAIM`). Empty = "name".
            name = optional(string, "")

            # Subject claim (`LANGFUSE_CUSTOM_SSO_SUB_CLAIM`). Empty = "sub".
            subject = optional(string, "")

            # Picture claim (`LANGFUSE_CUSTOM_SSO_IMAGE_CLAIM`). Empty = "picture".
            image = optional(string, "")
          }))
        }))

        # WorkOS.
        workos = optional(object({
          # The WorkOS client.
          client = object({
            # Client ID.
            client_id = string

            # Client secret.
            client_secret = string

            # Merge a sign-in with an existing account of the same email address
            # (`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
            # password to SSO. Safe only when the provider verifies emails.
            allow_account_linking = optional(bool, false)

            # Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
            # Unspecified = client_secret_basic. Use `none` for PKCE.
            client_auth_method = optional(string, "")

            # Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
            # Empty = the provider's defaults.
            checks = optional(list(string), [])

            # Algorithm the provider signs ID tokens with
            # (`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.
            id_token_signed_response_alg = optional(string, "")
          })

          # Route sign-in to this organization's connection
          # (`AUTH_WORKOS_ORGANIZATION_ID`).
          organization_id = optional(string, "")

          # Route sign-in to this connection (`AUTH_WORKOS_CONNECTION_ID`).
          connection_id = optional(string, "")
        }))

        # WordPress.com.
        wordpress = optional(object({
          # Client ID.
          client_id = string

          # Client secret.
          client_secret = string

          # Merge a sign-in with an existing account of the same email address
          # (`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
          # password to SSO. Safe only when the provider verifies emails.
          allow_account_linking = optional(bool, false)

          # Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
          # Unspecified = client_secret_basic. Use `none` for PKCE.
          client_auth_method = optional(string, "")

          # Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
          # Empty = the provider's defaults.
          checks = optional(list(string), [])

          # Algorithm the provider signs ID tokens with
          # (`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.
          id_token_signed_response_alg = optional(string, "")
        }))
      }))

      # Requests to identity providers time out after this many milliseconds
      # (`AUTH_SSO_TIMEOUT`). Empty = the auth library's default.
      sso_timeout_ms = optional(number)
    }))

    # Transactional email (invitations, password resets, notifications).
    # Unset = Langfuse sends no email.
    smtp = optional(object({
      # Connection URL (`SMTP_CONNECTION_URL`):
      # "smtp://user:password@host:587", "smtps://user:password@host:465",
      # or "ses://<region>" to send through Amazon SES with the AWS credential
      # chain (an IAM role through workload identity).
      connection_url = string

      # Sender address (`EMAIL_FROM_ADDRESS`).
      from_address = string
    }))

    # Product switches: telemetry, experimental features, organization
    # creation, content security policy.
    features = optional(object({
      # Send anonymous usage statistics to Langfuse (`TELEMETRY_ENABLED`) -
      # counts, never trace content.
      telemetry_enabled = optional(bool)

      # Turn on features still in development
      # (`LANGFUSE_ENABLE_EXPERIMENTAL_FEATURES`).
      experimental_features_enabled = optional(bool, false)

      # Emails allowed to create organizations
      # (`LANGFUSE_ALLOWED_ORGANIZATION_CREATORS`, Enterprise Edition). Empty
      # = every user may.
      allowed_organization_creators = optional(list(string), [])

      # Restrict the Content-Security-Policy to HTTPS connections
      # (`LANGFUSE_CSP_ENFORCE_HTTPS`). Turn on once `external_url` is https.
      csp_enforce_https = optional(bool, false)

      # Inputs and outputs longer than this many characters render as plain
      # text instead of markdown (`LANGFUSE_MARKDOWN_RENDER_CHARACTER_LIMIT`).
      # Empty = Langfuse's default (150000).
      markdown_render_character_limit = optional(number)
    }))

    # API-key and prompt caching in Redis.
    caching = optional(object({
      # Cache API key lookups (`LANGFUSE_CACHE_API_KEY_ENABLED`).
      api_key_cache_enabled = optional(bool)

      # API key cache lifetime in seconds (`LANGFUSE_CACHE_API_KEY_TTL_SECONDS`)
      # - also the longest a revoked key keeps working. Empty = Langfuse's
      # default (60).
      api_key_cache_ttl_seconds = optional(number)

      # Cache prompts (`LANGFUSE_CACHE_PROMPT_ENABLED`).
      prompt_cache_enabled = optional(bool)

      # Prompt cache lifetime in seconds (`LANGFUSE_CACHE_PROMPT_TTL_SECONDS`).
      # Empty = Langfuse's default (3600).
      prompt_cache_ttl_seconds = optional(number)
    }))

    # Queue sharding and worker concurrency for high ingestion volume.
    ingestion = optional(object({
      # Shards of the ingestion queue (`LANGFUSE_INGESTION_QUEUE_SHARD_COUNT`),
      # about 2-3 per Redis Cluster shard. Never reduce it on a live instance:
      # jobs on removed shards are stranded. Empty = 1.
      queue_shard_count = optional(number)

      # Ingestion jobs each worker processes concurrently, per shard
      # (`LANGFUSE_INGESTION_QUEUE_PROCESSING_CONCURRENCY`). Empty = Langfuse's
      # default (20). With ten shards, 2 keeps the same total.
      queue_processing_concurrency = optional(number)

      # Shards of the trace-upsert queue (`LANGFUSE_TRACE_UPSERT_QUEUE_SHARD_COUNT`).
      # Never reduce it on a live instance. Empty = 1.
      trace_upsert_queue_shard_count = optional(number)

      # Trace-upsert jobs each worker processes concurrently, per shard
      # (`LANGFUSE_TRACE_UPSERT_WORKER_CONCURRENCY`). Empty = Langfuse's
      # default (25).
      trace_upsert_worker_concurrency = optional(number)

      # Skip the ClickHouse read before each ingestion write for projects
      # created on or after this date, "YYYY-MM-DD"
      # (`LANGFUSE_SKIP_INGESTION_CLICKHOUSE_READ_MIN_PROJECT_CREATE_DATE`).
      # Safe for any instance whose full event history is in blob storage -
      # a fresh instance: set a date before the first project.
      skip_clickhouse_read_min_project_create_date = optional(string, "")
    }))

    # Langfuse AI: the instance-wide model behind the Assistant and Ask AI.
    ai_features = optional(object({
      # Model provider (`LANGFUSE_AI_PROVIDER`). Unspecified = Langfuse AI
      # unconfigured.
      provider = optional(string, "")

      # Model ID for the provider (`LANGFUSE_AI_MODEL`), e.g. a Bedrock geo
      # inference profile such as "eu.anthropic.claude-opus-5", or
      # "claude-opus-5" on Vertex.
      model = optional(string, "")

      # A smaller model for conversation titles and Ask AI
      # (`LANGFUSE_AI_SMALL_MODEL`). Empty = `model`.
      small_model = optional(string, "")

      # API key (`LANGFUSE_AI_API_KEY`) - required for anthropic and openai,
      # refused for bedrock and vertex (they use the cloud's credential
      # chain: an IAM role or Google service account through workload
      # identity).
      api_key = optional(string, "")

      # API base URL including "/v1" (`LANGFUSE_AI_BASE_URL`), for a gateway
      # or an OpenAI-compatible endpoint.
      base_url = optional(string, "")

      # Extra request headers (`LANGFUSE_AI_EXTRA_HEADERS`), anthropic and
      # openai only.
      extra_headers = optional(map(string), {})

      # Call the OpenAI Responses API instead of Chat Completions
      # (`LANGFUSE_AI_USE_RESPONSES_API`), openai only.
      use_responses_api = optional(bool, false)

      # Bedrock region (`LANGFUSE_AI_AWS_BEDROCK_REGION`). Empty = the
      # workload's region.
      bedrock_region = optional(string, "")

      # Vertex location (`LANGFUSE_AI_VERTEX_LOCATION`). Empty = "global".
      vertex_location = optional(string, "")

      # Project on this instance that receives traces of Assistant and Ask AI
      # runs (`LANGFUSE_AI_FEATURES_PROJECT_ID`). Empty = nothing is traced.
      trace_project_id = optional(string, "")

      # The Assistant, the in-app agent.
      assistant = optional(object({
        # Show the Assistant (`LANGFUSE_IN_APP_AGENT_ENABLED`, on web and
        # worker; the worker then consumes the run queue).
        enabled = optional(bool, false)

        # Concurrent Assistant runs per worker
        # (`LANGFUSE_IN_APP_AGENT_RUN_QUEUE_PROCESSING_CONCURRENCY`). Empty =
        # Langfuse's default (5).
        run_concurrency = optional(number)

        # Concurrent runs admitted per user
        # (`LANGFUSE_IN_APP_AGENT_MAX_ACTIVE_RUNS_PER_USER`). Empty = 5.
        max_active_runs_per_user = optional(number)

        # Concurrent runs admitted per organization
        # (`LANGFUSE_IN_APP_AGENT_MAX_ACTIVE_RUNS_PER_ORG`). Empty = 20.
        max_active_runs_per_org = optional(number)

        # Point the worker's MCP calls at the in-cluster web Service
        # (`LANGFUSE_MCP_BASE_URL`, and the Service host added to the web's
        # `LANGFUSE_MCP_ALLOWED_HOSTS`) instead of `external_url` - for an
        # external URL the worker cannot reach.
        use_internal_web_url = optional(bool, false)

        # The AWS Lambda MicroVM sandbox the Assistant runs code in. Unset =
        # Langfuse tools only, no code execution.
        sandbox = optional(object({
          # MicroVM image ARN.
          image_identifier = string

          # Execution role ARN. Defaults compose an AwsIamRole resource.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          execution_role_arn = string

          # AWS region of the MicroVM.
          region = string

          # Egress network connector ARN - a VPC connector whose security group
          # denies egress. Without it AWS attaches INTERNET_EGRESS and sandboxed
          # code reaches the public internet.
          egress_network_connector_arn = string
        }))
      }))
    }))

    # Evaluation runtime: how LLM-as-a-judge evaluators receive media, and
    # where code evaluators execute.
    evaluation = optional(object({
      # How LLM-as-a-judge evaluators send media to the model
      # (`LANGFUSE_EVALUATOR_MEDIA_TRANSPORT`). Unspecified = inline.
      media_transport = optional(string, "")

      # Largest inline attachment in bytes
      # (`LANGFUSE_EVALUATOR_MEDIA_INLINE_MAX_BYTES`). Empty = Langfuse's
      # default (20000000); evaluations with larger attachments fail.
      media_inline_max_bytes = optional(number)

      # Code evaluators: where evaluator code runs. Unset = code evaluators
      # are unavailable.
      code_evaluators = optional(object({
        # AWS Lambda functions - the production dispatcher.
        aws_lambda = optional(object({
          # Function for TypeScript and JavaScript evaluators. Empty =
          # "code-based-eval-executor-node".
          node_function_name = optional(string, "")

          # Function for Python evaluators. Empty =
          # "code-based-eval-executor-python".
          python_function_name = optional(string, "")

          # Custom Lambda endpoint (`LANGFUSE_CODE_EVAL_AWS_LAMBDA_ENDPOINT`).
          endpoint = optional(string, "")
        }))

        # Inside the worker process (`insecure-local`): only when every
        # evaluator's code is fully trusted.
        insecure_local = optional(object({
          # Per-evaluation timeout in milliseconds
          # (`LANGFUSE_CODE_EVAL_LOCAL_TIMEOUT_MS`). Empty = 2000.
          timeout_ms = optional(number)
        }))

        # Shards of the code-evaluation queue
        # (`LANGFUSE_CODE_EVAL_EXECUTION_QUEUE_SHARD_COUNT`). Empty = 1.
        queue_shard_count = optional(number)

        # Code evaluations each worker runs concurrently
        # (`LANGFUSE_CODE_EVAL_EXECUTION_WORKER_CONCURRENCY`). Empty = 5.
        concurrency = optional(number)
      }))
    }))

    # The Slack app that lets projects send notifications to Slack.
    slack = optional(object({
      # Slack app client ID.
      client_id = string

      # Slack app client secret.
      client_secret = string

      # Secret signing the OAuth state parameter. Empty = generated.
      state_secret = optional(string, "")
    }))

    # Enterprise Edition features: license key, Instance Management API, UI
    # customization and ingestion masking.
    enterprise = optional(object({
      # License key (`LANGFUSE_EE_LICENSE_KEY`).
      license_key = optional(string, "")

      # Key for the Instance Management API (`ADMIN_API_KEY`), which manages
      # organizations and their API keys - what the Langfuse Terraform
      # provider uses. Empty = the API is off.
      admin_api_key = optional(string, "")

      # UI customization.
      ui = optional(object({
        # Documentation link (`LANGFUSE_UI_DOCUMENTATION_HREF`).
        documentation_href = optional(string, "")

        # Support link (`LANGFUSE_UI_SUPPORT_HREF`).
        support_href = optional(string, "")

        # Feature request and bug report link (`LANGFUSE_UI_FEEDBACK_HREF`).
        feedback_href = optional(string, "")

        # Host shown in project settings for API requests
        # (`LANGFUSE_UI_API_HOST`), when API traffic uses another host than
        # the UI.
        api_host = optional(string, "")

        # Logo for light mode (`LANGFUSE_UI_LOGO_LIGHT_MODE_HREF`).
        logo_light_mode_href = optional(string, "")

        # Logo for dark mode (`LANGFUSE_UI_LOGO_DARK_MODE_HREF`).
        logo_dark_mode_href = optional(string, "")

        # Default model adapter in the playground and evals
        # (`LANGFUSE_UI_DEFAULT_MODEL_ADAPTER`).
        default_model_adapter = optional(string, "")

        # Default OpenAI base URL (`LANGFUSE_UI_DEFAULT_BASE_URL_OPENAI`).
        default_base_url_openai = optional(string, "")

        # Default Anthropic base URL (`LANGFUSE_UI_DEFAULT_BASE_URL_ANTHROPIC`).
        default_base_url_anthropic = optional(string, "")

        # Default Azure OpenAI base URL (`LANGFUSE_UI_DEFAULT_BASE_URL_AZURE`).
        default_base_url_azure = optional(string, "")

        # Show only these product modules (`LANGFUSE_UI_VISIBLE_PRODUCT_MODULES`),
        # e.g. "tracing", "evaluation", "prompt-management". Wins over
        # `hidden_product_modules`.
        visible_product_modules = optional(list(string), [])

        # Hide these product modules (`LANGFUSE_UI_HIDDEN_PRODUCT_MODULES`), e.g.
        # "playground", "datasets".
        hidden_product_modules = optional(list(string), [])

        # Links in the instance switcher (`LANGFUSE_UI_INSTANCE_LINKS`).
        instance_links = optional(list(object({
          # Label, unique among the links.
          name = string

          # URL of the other instance.
          url = string
        })), [])
      }))

      # Ingestion masking: a callback that redacts events before they are
      # stored.
      ingestion_masking = optional(object({
        # URL of the masking callback (`LANGFUSE_INGESTION_MASKING_CALLBACK_URL`).
        callback_url = string

        # Callback timeout in milliseconds. Empty = 500.
        timeout_ms = optional(number)

        # Drop events when the callback fails (fail closed) instead of storing
        # them unmasked (fail open, the default).
        fail_closed = optional(bool, false)

        # Retries of a failed callback. Empty = 1.
        max_retries = optional(number)

        # Headers of the ingestion request forwarded to the callback.
        propagated_headers = optional(list(string), [])
      }))
    }))

    # Langfuse's own logs and OpenTelemetry traces.
    observability = optional(object({
      # Log level.
      log_level = optional(string, "")

      # Log format: text, or json for log pipelines.
      log_format = optional(string, "")

      # Request headers carried into log context through OpenTelemetry
      # baggage (`LANGFUSE_LOG_PROPAGATED_HEADERS`), e.g. "x-request-id".
      log_propagated_headers = optional(list(string), [])

      # OTLP endpoint Langfuse sends its own traces to
      # (`OTEL_EXPORTER_OTLP_ENDPOINT`). Defaults compose a
      # KubernetesOtelCollector's OTLP/HTTP endpoint. Empty = no export.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      otlp_endpoint = optional(string, "")

      # Fraction of requests traced, above 0 and at most 1
      # (`OTEL_TRACE_SAMPLING_RATIO`). Empty = 1.
      trace_sampling_ratio = optional(number)
    }))

    # Outbound networking: proxies, private certificate authorities and the
    # host allowlists for user-configured endpoints.
    network = optional(object({
      # Proxy for outbound HTTPS requests (`HTTPS_PROXY`).
      https_proxy = optional(string, "")

      # Hosts reached directly, bypassing the proxy (`NO_PROXY`), e.g.
      # ".svc.cluster.local".
      no_proxy = optional(list(string), [])

      # Extra certificate authorities Langfuse trusts for outbound TLS
      # (`NODE_EXTRA_CA_CERTS`): private CAs of ClickHouse, SMTP, identity
      # providers or LLM endpoints.
      trusted_ca = optional(object({
        # Name of a Secret holding the bundle.
        secret_name = optional(string)

        # Name of a ConfigMap holding the bundle.
        config_map_name = optional(string)

        # Key holding the PEM bundle.
        key = optional(string)
      }))

      # Allowlists for endpoints users configure inside Langfuse. By default
      # Langfuse refuses private and internal addresses for them; an entry
      # here admits one.
      allowlists = optional(object({
        # Webhook targets (`LANGFUSE_WEBHOOK_WHITELISTED_*`).
        webhooks = optional(object({
          # Host names admitted.
          hosts = optional(list(string), [])

          # IP addresses admitted.
          ips = optional(list(string), [])

          # CIDR ranges admitted.
          ip_segments = optional(list(string), [])
        }))

        # LLM connection base URLs (`LANGFUSE_LLM_CONNECTION_WHITELISTED_*`).
        llm_connections = optional(object({
          # Host names admitted.
          hosts = optional(list(string), [])

          # IP addresses admitted.
          ips = optional(list(string), [])

          # CIDR ranges admitted.
          ip_segments = optional(list(string), [])
        }))

        # Blob storage export endpoints (`LANGFUSE_BLOB_STORAGE_ENDPOINT_WHITELISTED_*`).
        blob_storage_endpoints = optional(object({
          # Host names admitted.
          hosts = optional(list(string), [])

          # IP addresses admitted.
          ips = optional(list(string), [])

          # CIDR ranges admitted.
          ip_segments = optional(list(string), [])
        }))

        # External media storage endpoints
        # (`LANGFUSE_EXTERNAL_MEDIA_STORAGE_ENDPOINT_WHITELISTED_*`).
        external_media_storage_endpoints = optional(object({
          # Host names admitted.
          hosts = optional(list(string), [])

          # IP addresses admitted.
          ips = optional(list(string), [])

          # CIDR ranges admitted.
          ip_segments = optional(list(string), [])
        }))

        # SSO discovery URLs (`LANGFUSE_SSO_DISCOVERY_WHITELISTED_*`).
        sso_discovery = optional(object({
          # Host names admitted.
          hosts = optional(list(string), [])

          # IP addresses admitted.
          ips = optional(list(string), [])

          # CIDR ranges admitted.
          ip_segments = optional(list(string), [])
        }))
      }))
    }))

    # The web server: UI, public API and ingestion endpoints.
    web = optional(object({
      # Web image.
      image = optional(object({
        # Image repository. Empty = the chart's
        # (docker.langfuse.com/langfuse/langfuse for web,
        # docker.langfuse.com/langfuse/langfuse-worker for the worker).
        repository = optional(string, "")

        # Image tag. Empty = the chart's appVersion for `chart_version`. Keep
        # web and worker on the same version.
        tag = optional(string, "")

        # Pull policy: Always, IfNotPresent or Never. Empty = the chart's
        # default, Always.
        pull_policy = optional(string, "")

        # Names of existing image pull Secrets in Langfuse's namespace (for a
        # private mirror).
        pull_secret_names = optional(list(string), [])
      }))

      # Replicas. Two or more for availability. Ignored while an autoscaler
      # owns the count.
      replicas = optional(number)

      # CPU and memory. Langfuse recommends 2 CPU and 4 GiB per web pod for
      # production; the defaults suit evaluation.
      resources = optional(object({
        # The resource limits for the container.
        # Specify the maximum amount of CPU and memory that the container can use.
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))

        # The resource requests for the container.
        # Specify the minimum amount of CPU and memory that the container is guaranteed.
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))

      # Autoscaling. One mechanism at most.
      autoscaling = optional(object({
        # HorizontalPodAutoscaler on CPU.
        hpa = optional(object({
          # Minimum replicas.
          min_replicas = optional(number)

          # Maximum replicas.
          max_replicas = optional(number)

          # Target average CPU utilization percent.
          target_cpu_utilization_percent = optional(number)

          # Target average memory utilization percent. Empty = memory does not
          # drive scaling.
          target_memory_utilization_percent = optional(number)

          # Annotations on the HorizontalPodAutoscaler.
          annotations = optional(map(string), {})
        }))

        # KEDA ScaledObject (needs KEDA on the cluster - a KubernetesKeda).
        keda = optional(object({
          # Minimum replicas.
          min_replicas = optional(number)

          # Maximum replicas.
          max_replicas = optional(number)

          # Seconds between metric checks.
          polling_interval_seconds = optional(number)

          # Resource the trigger scales on.
          trigger = optional(string, "")

          # How `target_value` reads: a utilization percent or an absolute
          # average value.
          metric_type = optional(string, "")

          # Target value, e.g. "50" (percent) or "1500m" (average CPU).
          target_value = optional(string)

          # Container whose metrics count. Empty = every container in the pod.
          container_name = optional(string, "")
        }))

        # VerticalPodAutoscaler. The chart renders it only on a cluster that
        # serves the autoscaling.k8s.io/v1 API (the VPA CRDs); elsewhere it is
        # skipped silently.
        vpa = optional(object({
          # Resources the VPA controls, e.g. "cpu", "memory". Empty = both.
          controlled_resources = optional(list(string), [])

          # Upper bounds for recommendations.
          max_allowed = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))

          # Lower bounds for recommendations.
          min_allowed = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))

          # How recommendations are applied: Off, Initial, Recreate or Auto.
          update_mode = optional(string)
        }))
      }))

      # PodDisruptionBudget.
      pdb = optional(object({
        # Create the PodDisruptionBudget.
        enabled = optional(bool)

        # Minimum pods available during a voluntary disruption: an integer or a
        # percentage.
        min_available = optional(string, "")

        # Maximum pods unavailable during a voluntary disruption. Empty with
        # min_available also empty = 1.
        max_unavailable = optional(string, "")
      }))

      # Liveness probe on `/api/public/health`.
      liveness_probe = optional(object({
        # Seconds after start before the first probe. Empty = 20 (the chart
        # reads 0 as unset).
        initial_delay_seconds = optional(number)

        # Seconds between probes. Empty = 10.
        period_seconds = optional(number)

        # Seconds before a probe times out. Empty = 5.
        timeout_seconds = optional(number)

        # Consecutive failures before the probe fails. Empty = 3.
        failure_threshold = optional(number)

        # Consecutive successes before the probe passes. Empty = 1.
        success_threshold = optional(number)
      }))

      # Readiness probe on `/api/public/ready`.
      readiness_probe = optional(object({
        # Seconds after start before the first probe. Empty = 20 (the chart
        # reads 0 as unset).
        initial_delay_seconds = optional(number)

        # Seconds between probes. Empty = 10.
        period_seconds = optional(number)

        # Seconds before a probe times out. Empty = 5.
        timeout_seconds = optional(number)

        # Consecutive failures before the probe fails. Empty = 3.
        failure_threshold = optional(number)

        # Consecutive successes before the probe passes. Empty = 1.
        success_threshold = optional(number)
      }))

      # The web Service.
      service = optional(object({
        # Service type: ClusterIP, NodePort or LoadBalancer. Keep ClusterIP and
        # compose a route or ingress for external access.
        type = optional(string)

        # Port the web server listens on, and the Service's target port.
        port = optional(number)

        # Port the Service exposes, when it differs from `port`.
        external_port = optional(number)

        # Node port, for type NodePort.
        node_port = optional(number)

        # Annotations on the Service (internal load balancer recipes, for
        # example).
        annotations = optional(map(string), {})

        # Extra labels on the Service.
        labels = optional(map(string), {})
      }))

      # Pod placement.
      scheduling = optional(object({
        # Schedule onto nodes carrying these labels.
        node_selector = optional(map(string), {})

        # Tolerations for tainted nodes.
        tolerations = optional(list(object({
          # Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.
          key = optional(string, "")

          # How key/value match: "Equal" (default — value must match too) or "Exists"
          # (key presence alone matches).
          operator = optional(string, "")

          # Taint value to match when operator is "Equal".
          value = optional(string, "")

          # Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
          # "NoExecute". Empty tolerates all effects for the key.
          effect = optional(string, "")

          # For "NoExecute" taints only: how many seconds already-running pods stay bound
          # after the taint appears. Unset means tolerate forever.
          toleration_seconds = optional(number)
        })), [])

        # Node affinity: hard requirements and weighted preferences.
        node_affinity = optional(object({
          # Hard requirement. The outer list ORs its terms; expressions within one term AND.
          required = optional(list(object({
            match_expressions = list(object({
              # Node label key, e.g. "topology.kubernetes.io/zone".
              key = string

              # Operator: "In"/"NotIn" (value set), "Exists"/"DoesNotExist" (key presence), or
              # "Gt"/"Lt" (single integer value, as strings — the Kubernetes API convention).
              operator = string

              # Values for the operator: required non-empty for In/NotIn, exactly one integer
              # string for Gt/Lt, and must be empty for Exists/DoesNotExist.
              values = optional(list(string), [])
            }))
          })), [])

          # Weighted soft preferences.
          preferred = optional(list(object({
            # Preference weight, 1–100. Higher weights dominate placement scoring.
            weight = optional(number, 0)

            term = object({
              match_expressions = list(object({
                # Node label key, e.g. "topology.kubernetes.io/zone".
                key = string

                # Operator: "In"/"NotIn" (value set), "Exists"/"DoesNotExist" (key presence), or
                # "Gt"/"Lt" (single integer value, as strings — the Kubernetes API convention).
                operator = string

                # Values for the operator: required non-empty for In/NotIn, exactly one integer
                # string for Gt/Lt, and must be empty for Exists/DoesNotExist.
                values = optional(list(string), [])
              }))
            })
          })), [])
        }))

        # Pod affinity toward pods already running elsewhere.
        pod_affinity = optional(object({
          # Hard rules — unschedulable until satisfied. Use sparingly; they can deadlock rollouts.
          required = optional(list(object({
            # Labels of the pods to match against — for self-anti-affinity, the workload's own
            # selector labels (exported as the `selector_labels` output).
            match_labels = optional(map(string), {})

            # Node label defining the domain: "kubernetes.io/hostname" separates by node,
            # "topology.kubernetes.io/zone" by zone.
            topology_key = string

            # Namespaces whose pods are considered. Empty means the workload's own namespace.
            namespaces = optional(list(string), [])
          })), [])

          # Weighted soft rules — the scheduler's tiebreakers.
          preferred = optional(list(object({
            # Preference weight, 1–100.
            weight = optional(number, 0)

            term = object({
              # Labels of the pods to match against — for self-anti-affinity, the workload's own
              # selector labels (exported as the `selector_labels` output).
              match_labels = optional(map(string), {})

              # Node label defining the domain: "kubernetes.io/hostname" separates by node,
              # "topology.kubernetes.io/zone" by zone.
              topology_key = string

              # Namespaces whose pods are considered. Empty means the workload's own namespace.
              namespaces = optional(list(string), [])
            })
          })), [])
        }))

        # Pod anti-affinity - spread replicas across nodes or zones.
        pod_anti_affinity = optional(object({
          # Hard rules — unschedulable until satisfied. Use sparingly; they can deadlock rollouts.
          required = optional(list(object({
            # Labels of the pods to match against — for self-anti-affinity, the workload's own
            # selector labels (exported as the `selector_labels` output).
            match_labels = optional(map(string), {})

            # Node label defining the domain: "kubernetes.io/hostname" separates by node,
            # "topology.kubernetes.io/zone" by zone.
            topology_key = string

            # Namespaces whose pods are considered. Empty means the workload's own namespace.
            namespaces = optional(list(string), [])
          })), [])

          # Weighted soft rules — the scheduler's tiebreakers.
          preferred = optional(list(object({
            # Preference weight, 1–100.
            weight = optional(number, 0)

            term = object({
              # Labels of the pods to match against — for self-anti-affinity, the workload's own
              # selector labels (exported as the `selector_labels` output).
              match_labels = optional(map(string), {})

              # Node label defining the domain: "kubernetes.io/hostname" separates by node,
              # "topology.kubernetes.io/zone" by zone.
              topology_key = string

              # Namespaces whose pods are considered. Empty means the workload's own namespace.
              namespaces = optional(list(string), [])
            })
          })), [])
        }))

        # Topology spread constraints.
        topology_spread_constraints = optional(list(object({
          # Maximum allowed difference in matching-pod counts between any two domains.
          # 1 is the strictest even spread.
          max_skew = optional(number, 0)

          # Node label defining the domains to spread across (e.g.
          # "topology.kubernetes.io/zone").
          topology_key = string

          # What happens when the constraint cannot be met: "DoNotSchedule" (hard — pod
          # stays Pending) or "ScheduleAnyway" (soft — scheduler minimizes skew).
          when_unsatisfiable = string

          # Labels selecting the pods counted per domain. Omit to have the module default to
          # the workload's own selector labels — self-spreading, the overwhelmingly common
          # intent.
          match_labels = optional(map(string), {})
        })), [])

        # PriorityClass for the pods.
        priority_class_name = optional(string, "")
      }))

      # Workload metadata and rollout.
      workload = optional(object({
        # Annotations on the pods.
        pod_annotations = optional(map(string), {})

        # Labels on the pods.
        pod_labels = optional(map(string), {})

        # Annotations on the Deployment.
        deployment_annotations = optional(map(string), {})

        # Extra labels on the Deployment.
        deployment_labels = optional(map(string), {})

        # Rollout strategy.
        strategy = optional(object({
          # Strategy type: RollingUpdate or Recreate. Empty = RollingUpdate.
          type = optional(string, "")

          # Extra pods during a rolling update: an integer or a percentage.
          max_surge = optional(string, "")

          # Pods unavailable during a rolling update: an integer or a percentage.
          max_unavailable = optional(string, "")
        }))

        # Old ReplicaSets kept for rollback. Empty = 10.
        revision_history_limit = optional(number)

        # Extra /etc/hosts entries.
        host_aliases = optional(list(object({
          # The IP the hostnames resolve to.
          ip = string

          # Hostnames mapped to the IP.
          hostnames = list(string)
        })), [])
      }))

      # Extra plain environment variables for the web server only.
      extra_env = optional(map(string), {})

      # Extra environment variables for the web server read from existing
      # Secrets, keyed by variable name.
      extra_env_from_secret = optional(map(object({
        # The name of the Kubernetes Secret.
        name = optional(string, "")

        # The key within the Kubernetes Secret.
        key = optional(string, "")
      })), {})
    }))

    # The worker: queue consumers, evaluations, exports and retention.
    worker = optional(object({
      # Worker image.
      image = optional(object({
        # Image repository. Empty = the chart's
        # (docker.langfuse.com/langfuse/langfuse for web,
        # docker.langfuse.com/langfuse/langfuse-worker for the worker).
        repository = optional(string, "")

        # Image tag. Empty = the chart's appVersion for `chart_version`. Keep
        # web and worker on the same version.
        tag = optional(string, "")

        # Pull policy: Always, IfNotPresent or Never. Empty = the chart's
        # default, Always.
        pull_policy = optional(string, "")

        # Names of existing image pull Secrets in Langfuse's namespace (for a
        # private mirror).
        pull_secret_names = optional(list(string), [])
      }))

      # Replicas. Scale on CPU: a worker above 50% of 2 CPU is saturated.
      # Ignored while an autoscaler owns the count.
      replicas = optional(number)

      # CPU and memory. Langfuse recommends 2 CPU and 4 GiB per worker for
      # production; the defaults suit evaluation.
      resources = optional(object({
        # The resource limits for the container.
        # Specify the maximum amount of CPU and memory that the container can use.
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))

        # The resource requests for the container.
        # Specify the minimum amount of CPU and memory that the container is guaranteed.
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))

      # Autoscaling. One mechanism at most.
      autoscaling = optional(object({
        # HorizontalPodAutoscaler on CPU.
        hpa = optional(object({
          # Minimum replicas.
          min_replicas = optional(number)

          # Maximum replicas.
          max_replicas = optional(number)

          # Target average CPU utilization percent.
          target_cpu_utilization_percent = optional(number)

          # Target average memory utilization percent. Empty = memory does not
          # drive scaling.
          target_memory_utilization_percent = optional(number)

          # Annotations on the HorizontalPodAutoscaler.
          annotations = optional(map(string), {})
        }))

        # KEDA ScaledObject (needs KEDA on the cluster - a KubernetesKeda).
        keda = optional(object({
          # Minimum replicas.
          min_replicas = optional(number)

          # Maximum replicas.
          max_replicas = optional(number)

          # Seconds between metric checks.
          polling_interval_seconds = optional(number)

          # Resource the trigger scales on.
          trigger = optional(string, "")

          # How `target_value` reads: a utilization percent or an absolute
          # average value.
          metric_type = optional(string, "")

          # Target value, e.g. "50" (percent) or "1500m" (average CPU).
          target_value = optional(string)

          # Container whose metrics count. Empty = every container in the pod.
          container_name = optional(string, "")
        }))

        # VerticalPodAutoscaler. The chart renders it only on a cluster that
        # serves the autoscaling.k8s.io/v1 API (the VPA CRDs); elsewhere it is
        # skipped silently.
        vpa = optional(object({
          # Resources the VPA controls, e.g. "cpu", "memory". Empty = both.
          controlled_resources = optional(list(string), [])

          # Upper bounds for recommendations.
          max_allowed = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))

          # Lower bounds for recommendations.
          min_allowed = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))

          # How recommendations are applied: Off, Initial, Recreate or Auto.
          update_mode = optional(string)
        }))
      }))

      # PodDisruptionBudget.
      pdb = optional(object({
        # Create the PodDisruptionBudget.
        enabled = optional(bool)

        # Minimum pods available during a voluntary disruption: an integer or a
        # percentage.
        min_available = optional(string, "")

        # Maximum pods unavailable during a voluntary disruption. Empty with
        # min_available also empty = 1.
        max_unavailable = optional(string, "")
      }))

      # Liveness probe on `/api/health`.
      liveness_probe = optional(object({
        # Seconds after start before the first probe. Empty = 20 (the chart
        # reads 0 as unset).
        initial_delay_seconds = optional(number)

        # Seconds between probes. Empty = 10.
        period_seconds = optional(number)

        # Seconds before a probe times out. Empty = 5.
        timeout_seconds = optional(number)

        # Consecutive failures before the probe fails. Empty = 3.
        failure_threshold = optional(number)

        # Consecutive successes before the probe passes. Empty = 1.
        success_threshold = optional(number)
      }))

      # Fail the liveness probe when the worker has consumed no job on any
      # queue for longer than `queue_consumption_stuck_threshold_minutes`, so
      # a wedged worker restarts itself. Scheduled jobs keep a healthy worker
      # busy at least hourly.
      fail_liveness_if_queue_consumption_stuck = optional(bool, false)

      # Minutes without consumption before the probe above fails
      # (`LANGFUSE_QUEUE_CONSUMPTION_STUCK_THRESHOLD_MINUTES`). Empty =
      # Langfuse's default (60).
      queue_consumption_stuck_threshold_minutes = optional(number)

      # Pod placement.
      scheduling = optional(object({
        # Schedule onto nodes carrying these labels.
        node_selector = optional(map(string), {})

        # Tolerations for tainted nodes.
        tolerations = optional(list(object({
          # Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.
          key = optional(string, "")

          # How key/value match: "Equal" (default — value must match too) or "Exists"
          # (key presence alone matches).
          operator = optional(string, "")

          # Taint value to match when operator is "Equal".
          value = optional(string, "")

          # Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
          # "NoExecute". Empty tolerates all effects for the key.
          effect = optional(string, "")

          # For "NoExecute" taints only: how many seconds already-running pods stay bound
          # after the taint appears. Unset means tolerate forever.
          toleration_seconds = optional(number)
        })), [])

        # Node affinity: hard requirements and weighted preferences.
        node_affinity = optional(object({
          # Hard requirement. The outer list ORs its terms; expressions within one term AND.
          required = optional(list(object({
            match_expressions = list(object({
              # Node label key, e.g. "topology.kubernetes.io/zone".
              key = string

              # Operator: "In"/"NotIn" (value set), "Exists"/"DoesNotExist" (key presence), or
              # "Gt"/"Lt" (single integer value, as strings — the Kubernetes API convention).
              operator = string

              # Values for the operator: required non-empty for In/NotIn, exactly one integer
              # string for Gt/Lt, and must be empty for Exists/DoesNotExist.
              values = optional(list(string), [])
            }))
          })), [])

          # Weighted soft preferences.
          preferred = optional(list(object({
            # Preference weight, 1–100. Higher weights dominate placement scoring.
            weight = optional(number, 0)

            term = object({
              match_expressions = list(object({
                # Node label key, e.g. "topology.kubernetes.io/zone".
                key = string

                # Operator: "In"/"NotIn" (value set), "Exists"/"DoesNotExist" (key presence), or
                # "Gt"/"Lt" (single integer value, as strings — the Kubernetes API convention).
                operator = string

                # Values for the operator: required non-empty for In/NotIn, exactly one integer
                # string for Gt/Lt, and must be empty for Exists/DoesNotExist.
                values = optional(list(string), [])
              }))
            })
          })), [])
        }))

        # Pod affinity toward pods already running elsewhere.
        pod_affinity = optional(object({
          # Hard rules — unschedulable until satisfied. Use sparingly; they can deadlock rollouts.
          required = optional(list(object({
            # Labels of the pods to match against — for self-anti-affinity, the workload's own
            # selector labels (exported as the `selector_labels` output).
            match_labels = optional(map(string), {})

            # Node label defining the domain: "kubernetes.io/hostname" separates by node,
            # "topology.kubernetes.io/zone" by zone.
            topology_key = string

            # Namespaces whose pods are considered. Empty means the workload's own namespace.
            namespaces = optional(list(string), [])
          })), [])

          # Weighted soft rules — the scheduler's tiebreakers.
          preferred = optional(list(object({
            # Preference weight, 1–100.
            weight = optional(number, 0)

            term = object({
              # Labels of the pods to match against — for self-anti-affinity, the workload's own
              # selector labels (exported as the `selector_labels` output).
              match_labels = optional(map(string), {})

              # Node label defining the domain: "kubernetes.io/hostname" separates by node,
              # "topology.kubernetes.io/zone" by zone.
              topology_key = string

              # Namespaces whose pods are considered. Empty means the workload's own namespace.
              namespaces = optional(list(string), [])
            })
          })), [])
        }))

        # Pod anti-affinity - spread replicas across nodes or zones.
        pod_anti_affinity = optional(object({
          # Hard rules — unschedulable until satisfied. Use sparingly; they can deadlock rollouts.
          required = optional(list(object({
            # Labels of the pods to match against — for self-anti-affinity, the workload's own
            # selector labels (exported as the `selector_labels` output).
            match_labels = optional(map(string), {})

            # Node label defining the domain: "kubernetes.io/hostname" separates by node,
            # "topology.kubernetes.io/zone" by zone.
            topology_key = string

            # Namespaces whose pods are considered. Empty means the workload's own namespace.
            namespaces = optional(list(string), [])
          })), [])

          # Weighted soft rules — the scheduler's tiebreakers.
          preferred = optional(list(object({
            # Preference weight, 1–100.
            weight = optional(number, 0)

            term = object({
              # Labels of the pods to match against — for self-anti-affinity, the workload's own
              # selector labels (exported as the `selector_labels` output).
              match_labels = optional(map(string), {})

              # Node label defining the domain: "kubernetes.io/hostname" separates by node,
              # "topology.kubernetes.io/zone" by zone.
              topology_key = string

              # Namespaces whose pods are considered. Empty means the workload's own namespace.
              namespaces = optional(list(string), [])
            })
          })), [])
        }))

        # Topology spread constraints.
        topology_spread_constraints = optional(list(object({
          # Maximum allowed difference in matching-pod counts between any two domains.
          # 1 is the strictest even spread.
          max_skew = optional(number, 0)

          # Node label defining the domains to spread across (e.g.
          # "topology.kubernetes.io/zone").
          topology_key = string

          # What happens when the constraint cannot be met: "DoNotSchedule" (hard — pod
          # stays Pending) or "ScheduleAnyway" (soft — scheduler minimizes skew).
          when_unsatisfiable = string

          # Labels selecting the pods counted per domain. Omit to have the module default to
          # the workload's own selector labels — self-spreading, the overwhelmingly common
          # intent.
          match_labels = optional(map(string), {})
        })), [])

        # PriorityClass for the pods.
        priority_class_name = optional(string, "")
      }))

      # Workload metadata and rollout.
      workload = optional(object({
        # Annotations on the pods.
        pod_annotations = optional(map(string), {})

        # Labels on the pods.
        pod_labels = optional(map(string), {})

        # Annotations on the Deployment.
        deployment_annotations = optional(map(string), {})

        # Extra labels on the Deployment.
        deployment_labels = optional(map(string), {})

        # Rollout strategy.
        strategy = optional(object({
          # Strategy type: RollingUpdate or Recreate. Empty = RollingUpdate.
          type = optional(string, "")

          # Extra pods during a rolling update: an integer or a percentage.
          max_surge = optional(string, "")

          # Pods unavailable during a rolling update: an integer or a percentage.
          max_unavailable = optional(string, "")
        }))

        # Old ReplicaSets kept for rollback. Empty = 10.
        revision_history_limit = optional(number)

        # Extra /etc/hosts entries.
        host_aliases = optional(list(object({
          # The IP the hostnames resolve to.
          ip = string

          # Hostnames mapped to the IP.
          hostnames = list(string)
        })), [])
      }))

      # Extra plain environment variables for the worker only (queue consumer
      # switches such as QUEUE_CONSUMER_BATCH_EXPORT_QUEUE_IS_ENABLED, for
      # example).
      extra_env = optional(map(string), {})

      # Extra environment variables for the worker read from existing
      # Secrets, keyed by variable name.
      extra_env_from_secret = optional(map(object({
        # The name of the Kubernetes Secret.
        name = optional(string, "")

        # The key within the Kubernetes Secret.
        key = optional(string, "")
      })), {})
    }))

    # The ServiceAccount both workloads run as, and its cloud identity.
    service_account = optional(object({
      # Run as an existing ServiceAccount instead of the one the chart
      # creates (`<metadata.name>`). Its annotations are then yours to
      # manage; `annotations` and `workload_identity` are refused.
      existing_name = optional(string, "")

      # Annotations on the ServiceAccount the chart creates.
      annotations = optional(map(string), {})

      # The cloud identity the ServiceAccount federates with - keyless access
      # to S3, GCS, Bedrock, Vertex and SES. Rendered as the ServiceAccount
      # annotation each cloud expects. On AKS, also label the pods
      # `azure.workload.identity/use: "true"` through `workload.pod_labels`.
      workload_identity = optional(object({
        # GKE Workload Identity: annotate the ServiceAccount with a GCP service account email.
        gke = optional(object({
          # GCP service account email, e.g. "dns-manager@my-project.iam.gserviceaccount.com".
          # Applied as the `iam.gke.io/gcp-service-account` annotation. Accepts a literal
          # email or a reference to a GcpServiceAccount resource's output.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          service_account_email = string
        }))

        # EKS IRSA: annotate the ServiceAccount with an AWS IAM role ARN.
        eks = optional(object({
          # AWS IAM role ARN, e.g. "arn:aws:iam::123456789012:role/dns-manager".
          # Applied as the `eks.amazonaws.com/role-arn` annotation. Accepts a literal ARN
          # or a reference to an AwsIamRole resource's output.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          role_arn = string
        }))

        # Azure AD Workload Identity: annotate the ServiceAccount with an Entra application
        # (or user-assigned managed identity) client ID.
        aks = optional(object({
          # Client ID (GUID) of the user-assigned managed identity or Entra application.
          # Applied as the `azure.workload.identity/client-id` annotation. Accepts a literal
          # GUID or a reference to an AzureUserAssignedIdentity resource's output.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          client_id = string

          # Entra tenant ID (GUID). Optional: only needed for cross-tenant scenarios; when
          # omitted the azure-workload-identity webhook uses its default tenant. Applied as
          # the `azure.workload.identity/tenant-id` annotation when set.
          tenant_id = optional(string)
        }))
      }))

      # Mount the ServiceAccount token into the pods. Workload identity on
      # EKS and GKE needs it.
      automount_token = optional(bool)
    }))

    # Pod-level security context for both workloads.
    pod_security_context = optional(object({
      # UID all container processes run as unless overridden per container.
      run_as_user = optional(number)

      # Primary GID all container processes run as unless overridden per container.
      run_as_group = optional(number)

      # Refuse to start any container whose effective user is root.
      run_as_non_root = optional(bool)

      # GID that owns mounted volumes and is added to every container's supplemental
      # groups — the standard fix for "permission denied" on persistent volumes written
      # by non-root apps.
      fs_group = optional(number)

      # When volume ownership is re-chowned to fs_group: "Always" (default) or
      # "OnRootMismatch" (skip the recursive chown when the root already matches —
      # dramatically faster pod starts on large volumes).
      fs_group_change_policy = optional(string, "")

      # Additional group IDs applied to all container processes.
      supplemental_groups = optional(list(number), [])

      # Kernel parameters set for the pod. Only safe sysctls (or those the cluster
      # administrator has allow-listed on the kubelet) are admitted.
      sysctls = optional(list(object({
        # Sysctl name, e.g. "net.core.somaxconn".
        name = string

        # Sysctl value, e.g. "1024".
        value = string
      })), [])

      # Pod-wide seccomp profile; containers may override with their own.
      seccomp_profile = optional(object({
        # Profile type: "RuntimeDefault" (the container runtime's default filter — the
        # recommended baseline), "Unconfined" (no filtering), or "Localhost" (a profile
        # file installed on the node, named via localhost_profile).
        type = string

        # Path of the profile file relative to the node's seccomp profile root. Required
        # when (and only meaningful when) type is "Localhost".
        localhost_profile = optional(string, "")
      }))
    }))

    # Container-level security context for both workloads.
    container_security_context = optional(object({
      # Runs the container with full host access — equivalent to root on the node.
      # Required by some node-level agents (device managers, network plugins). Never
      # combine with untrusted images.
      privileged = optional(bool, false)

      # UID the container process runs as. Overrides the image's USER directive.
      run_as_user = optional(number)

      # Primary GID the container process runs as.
      run_as_group = optional(number)

      # Refuses to start the container if its effective user is root. The standard
      # baseline hardening — it catches images that silently default to UID 0.
      run_as_non_root = optional(bool)

      # Mounts the container's root filesystem read-only. Pair with EmptyDir mounts for
      # paths the app must write (e.g. /tmp).
      read_only_root_filesystem = optional(bool)

      # Whether the process can gain more privileges than its parent (setuid binaries,
      # file capabilities). The restricted Pod Security Standard requires this to be
      # false. Always true when `privileged` is set, so leave it unset in that case.
      allow_privilege_escalation = optional(bool)

      # Linux capabilities to add or drop. The restricted profile drops ALL and adds
      # back only NET_BIND_SERVICE when needed. Capability names are uppercase without
      # the CAP_ prefix (e.g. "NET_ADMIN", "SYS_TIME").
      capabilities = optional(object({
        # Capabilities to add (e.g. "NET_BIND_SERVICE").
        add = optional(list(string), [])

        # Capabilities to drop. Use ["ALL"] as the hardened baseline.
        drop = optional(list(string), [])
      }))

      # Seccomp syscall filter for the container. "RuntimeDefault" is the hardened
      # baseline; "Localhost" selects a node-local profile file via `localhost_profile`.
      seccomp_profile = optional(object({
        # Profile type: "RuntimeDefault" (the container runtime's default filter — the
        # recommended baseline), "Unconfined" (no filtering), or "Localhost" (a profile
        # file installed on the node, named via localhost_profile).
        type = string

        # Path of the profile file relative to the node's seccomp profile root. Required
        # when (and only meaningful when) type is "Localhost".
        localhost_profile = optional(string, "")
      }))
    }))

    # DNS settings for both workloads' pods.
    dns_config = optional(object({
      # Nameserver IPs (max 3 total after merging with the policy's servers).
      nameservers = optional(list(string), [])

      # Search domains for hostname lookup.
      searches = optional(list(string), [])

      # resolv.conf options, e.g. name "ndots" value "2". Value may be empty for flags.
      options = optional(list(object({
        name  = string
        value = optional(string, "")
      })), [])
    }))

    # Extra volumes mounted into both workloads - a file a variable in
    # `extra_env` points at, for example.
    extra_volumes = optional(list(object({
      # Name of the volume mount. Must be unique within the container.
      # Used to correlate with the volume definition.
      name = string

      # Path within the container at which the volume should be mounted.
      # Must be an absolute path.
      mount_path = string

      # Whether the volume should be mounted read-only.
      # Default is false.
      read_only = optional(bool, false)

      # Path within the volume from which the container's volume should be mounted.
      # Defaults to "" (volume's root).
      # Useful for mounting a subdirectory of a volume.
      sub_path = optional(string, "")

      # ConfigMap volume source.
      # Use this to mount a ConfigMap as a file or directory.
      config_map = optional(object({
        # Name of the ConfigMap to mount.
        # Can reference a ConfigMap defined in spec.config_maps or an existing one in the namespace.
        name = string

        # Specific key from the ConfigMap to mount as a single file.
        # If not specified, all keys are mounted as files in the directory.
        key = optional(string, "")

        # If key is specified, this is the filename to use for the mounted file.
        # Defaults to the key name if not specified.
        # Example: key="config" path="app.yaml" mounts the "config" key as "app.yaml"
        path = optional(string, "")

        # Mode bits to use on created files. Must be a value between 0 and 0777.
        # Defaults to 0644.
        # Use 0755 (493 in decimal) for executable scripts.
        default_mode = optional(number, 0)
      }))

      # Secret volume source.
      # Use this to mount a Secret as a file or directory.
      secret = optional(object({
        # Name of the Secret to mount.
        name = string

        # Specific key from the Secret to mount as a single file.
        # If not specified, all keys are mounted as files in the directory.
        key = optional(string, "")

        # If key is specified, this is the filename to use for the mounted file.
        # Defaults to the key name if not specified.
        path = optional(string, "")

        # Mode bits to use on created files. Must be a value between 0 and 0777.
        # Defaults to 0644.
        default_mode = optional(number, 0)
      }))

      # HostPath volume source.
      # Use this to mount a file or directory from the host node's filesystem.
      # Common for DaemonSets that need access to node-level resources.
      host_path = optional(object({
        # Path on the host to mount.
        path = string

        # Type of the host path.
        # Valid values:
        #   "" - Empty string (default) means no check is performed before mounting
        #   "DirectoryOrCreate" - Create directory if it doesn't exist
        #   "Directory" - Directory must exist
        #   "FileOrCreate" - Create file if it doesn't exist
        #   "File" - File must exist
        #   "Socket" - UNIX socket must exist
        #   "CharDevice" - Character device must exist
        #   "BlockDevice" - Block device must exist
        type = optional(string, "")
      }))

      # EmptyDir volume source.
      # Use this for temporary storage that is erased when the pod is removed.
      # Useful for scratch space, caching, or sharing data between containers.
      empty_dir = optional(object({
        # Medium for the empty directory.
        # "" (default) uses the node's default medium (typically disk).
        # "Memory" uses a tmpfs (RAM-backed filesystem).
        #
        # Memory-backed volumes are faster but:
        # - Count against container memory limits
        # - Are lost on node restart
        # - Should have sizeLimit set to prevent OOM
        medium = optional(string, "")

        # Size limit for the empty directory.
        # Format: Kubernetes quantity (e.g., "1Gi", "500Mi").
        # Only strictly enforced when medium is "Memory".
        # For disk-backed volumes, this is a best-effort limit.
        size_limit = optional(string, "")
      }))

      # PersistentVolumeClaim volume source.
      # Use this to mount an existing PVC.
      # For StatefulSets, this can reference a volumeClaimTemplate.
      pvc = optional(object({
        # Name of the PersistentVolumeClaim to mount.
        # For StatefulSets, this can be the name of a volumeClaimTemplate.
        claim_name = string

        # Whether the PVC should be mounted read-only.
        # Default is false.
        read_only = optional(bool, false)
      }))

      # Projected ServiceAccount token volume source.
      # Use this to mount a short-lived, audience-bound identity token that the
      # kubelet issues for the pod's ServiceAccount and rotates automatically.
      service_account_token = optional(object({
        # Intended audience of the token. The receiving service must identify
        # itself with this audience when verifying the token; a token minted for a
        # different audience is rejected. Required: an audience-less token would be
        # replayable against any service in the cluster.
        audience = string

        # Requested lifetime of the token in seconds. The kubelet starts rotating
        # the token when it passes 80% of its lifetime or 24 hours, whichever is
        # shorter. Defaults to 3600 (1 hour). The Kubernetes API enforces a
        # minimum of 600 (10 minutes).
        expiration_seconds = optional(number, 0)

        # Filename for the token relative to the mount path.
        # Defaults to "token".
        path = optional(string, "")
      }))
    })), [])

    # Extra plain environment variables for both workloads: any Langfuse
    # variable this spec does not type, set verbatim (e.g.
    # LANGFUSE_INGESTION_QUEUE_DELAY_MS). A variable the spec types is set
    # through its field; naming it here too conflicts with the chart's own
    # validation. Secret values belong in `extra_env_from_secret`.
    extra_env = optional(map(string), {})

    # Extra environment variables for both workloads read from existing
    # Secrets in Langfuse's namespace, keyed by variable name.
    extra_env_from_secret = optional(map(object({
      # The name of the Kubernetes Secret.
      name = optional(string, "")

      # The key within the Kubernetes Secret.
      key = optional(string, "")
    })), {})

    # Advanced escape hatch: raw Helm values merged LAST (Helm `-f`
    # semantics) over everything this spec renders - later keys win. Use it
    # for chart keys this spec does not type (sidecars, init containers,
    # lifecycle hooks, `extraManifests`). The module re-pins
    # `fullnameOverride` and keeps the bundled stores off after the merge.
    # Never put secret material here: it lands in the Helm release in plain
    # text. YAML document as a string.
    helm_values = optional(string, "")
  })
}
