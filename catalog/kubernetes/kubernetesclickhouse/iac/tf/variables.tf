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
  description = "KubernetesClickHouse specification"
  type = object({
    # Namespace to deploy the cluster into. Accepts a literal namespace
    # name or a reference to a KubernetesNamespace resource. The
    # ClickHouseInstallation is namespaced; the Altinity operator must
    # be watching this namespace (see the operator kind's
    # `watch_namespaces`).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    namespace = string

    # When true, the namespace is created (with the standard Planton
    # governance labels) before deploying and deleted with the resource.
    # When false, the namespace must already exist.
    create_namespace = optional(bool, false)

    # ClickHouse server version tag, e.g. "25.3" (an LTS line) or
    # "24.8". Resolves to the `clickhouse/clickhouse-server` image
    # unless `image` overrides the repository. Always pin a version:
    # the operator's built-in fallback is the `latest` tag, which makes
    # cluster upgrades happen implicitly on pod restarts.
    version = string

    # Override the ClickHouse server image (air-gap / private-mirror
    # path). `repo` empty = "clickhouse/clickhouse-server"; `tag` empty
    # = `version`.
    image = optional(object({
      # The repository of the image (e.g., "gcr.io/project/image").
      repo = optional(string, "")

      # The tag of the image (e.g., "latest" or "1.0.0").
      tag = optional(string, "")

      # The name of the image pull secret for private image repositories.
      pull_secret_name = optional(string, "")
    }))

    # Logical ClickHouse cluster name — the entry in `remote_servers`
    # that Distributed tables and `ON CLUSTER '<name>'` DDL statements
    # target. Also a segment of every generated child name (see the
    # naming contract above), which is why the operator's CRD caps it at
    # 15 characters. Default: "main".
    #
    # Verified live: the in-server cluster definition propagates through
    # mounted config and can LAG the installation reaching Completed —
    # `ON CLUSTER` DDL initiated in that window silently executes on the
    # subset of hosts the initiator can see (and still returns success).
    # Before running distributed DDL right after a deploy or topology
    # change, confirm `SELECT count() FROM system.clusters WHERE cluster
    # = '<name>'` reports every declared host.
    cluster_name = optional(string)

    # Number of shards. Each shard holds a disjoint slice of the data;
    # queries against Distributed-engine tables run on all shards in
    # parallel. 1 (the default) means all data on every host — scale
    # this only for datasets or write rates a single shard cannot carry.
    shards = optional(number)

    # Replicas per shard. Each replica is a full copy of its shard's
    # data, kept in sync through ReplicatedMergeTree — which requires
    # `coordination`. 1 (the default) means no redundancy: a lost volume
    # loses that shard's data. Production: 2–3.
    replicas = optional(number)

    # CPU and memory for each ClickHouse host container. Empty = no
    # requests/limits (schedulable anywhere, no protection). ClickHouse
    # is memory-hungry under analytical load: give production hosts at
    # least 4Gi and set `max_server_memory_usage_to_ram_ratio` (via
    # `settings`) if the limit is tight.
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

    # Size of the persistent data volume for EACH host (e.g. "100Gi").
    # Rendered as the data VolumeClaimTemplate mounted at
    # /var/lib/clickhouse. Kubernetes cannot shrink PVCs, and expanding
    # requires a storage class that allows it — plan for growth.
    disk_size = string

    # Storage class for the data volumes. Accepts a literal name or a
    # reference to a KubernetesStorageClass resource. Empty = the
    # cluster's default storage class.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    storage_class = optional(string, "")

    # Size of a SEPARATE persistent volume for server logs (e.g.
    # "10Gi"), mounted at /var/log/clickhouse-server. Empty (default) =
    # logs live on the container filesystem and vanish with the pod.
    # Useful when log retention matters and no log shipper is in place.
    log_disk_size = optional(string, "")

    # Keep the data volumes when the resource (or a host) is deleted.
    # Maps to the operator's PVC reclaim policy: false (the operator
    # default) deletes PVCs with their StatefulSet; true retains them,
    # so a re-created cluster with the same name re-attaches the data.
    # Retained PVCs are never garbage-collected — deleting them becomes
    # a manual operation.
    retain_volumes_on_delete = optional(bool, false)

    # Coordination service for replication and distributed DDL.
    # UNSET (recommended): the module deploys a managed ClickHouse
    # Keeper (3 nodes) automatically whenever the topology needs one
    # (replicas > 1 or shards > 1) and none otherwise. Set explicitly to
    # size the managed Keeper, point at external coordination, or opt
    # out entirely.
    coordination = optional(object({
      # Coordination flavor. See the enum values for semantics.
      type = optional(string, "")

      # Managed Keeper sizing. Only read when the effective type is
      # managed_keeper; empty = 3 replicas with operator-default
      # resources and a 10Gi volume.
      keeper = optional(object({
        # Keeper ensemble size. Raft quorum needs an odd count: 1 (dev — no
        # fault tolerance), 3 (production — survives one node loss), or 5
        # (survives two). Default 3.
        replicas = optional(number)

        # CPU and memory per Keeper pod. Keeper is deliberately light
        # (C++, no JVM): 100m/256Mi requests with 500m/1Gi limits carry most
        # clusters. Empty = operator defaults.
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

        # Persistent volume per Keeper pod for the coordination log and
        # snapshots (metadata only, not table data). Default "10Gi".
        disk_size = optional(string)

        # Storage class for the Keeper volumes. Empty = the cluster's
        # default storage class.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        storage_class = optional(string, "")
      }))

      # External coordination endpoints. Required when type is
      # external_keeper or external_zookeeper.
      external = optional(object({
        # Ensemble nodes. List every member for client-side failover.
        nodes = optional(list(object({
          # DNS name or IP of the node, e.g.
          # "zk-0.zk-headless.zoo.svc.cluster.local".
          host = string

          # Client port. Default 2181 — the standard for both ZooKeeper and
          # ClickHouse Keeper.
          port = optional(number)
        })), [])

        # Optional root znode path under which this cluster stores its
        # replication and DDL metadata — set it when several ClickHouse
        # clusters share one ensemble (e.g. "/clickhouse/prod-analytics").
        root = optional(string, "")

        # Optional digest-auth credentials for the ensemble in
        # "user:password" form (ZooKeeper ACL identity). Delivered into the
        # generated config verbatim.
        identity = optional(string, "")
      }))
    }))

    # ClickHouse users to provision, each with a password (declared, or
    # generated by the module when omitted) delivered through a
    # Kubernetes Secret (never plaintext in the CHI). The
    # built-in `default` user stays operator-managed: passwordless but
    # network-restricted to the cluster's own pods (the operator
    # generates the IP allowlist) — create named users for every real
    # client. KNOW THIS (upstream-documented): secret-sourced passwords
    # reach ClickHouse through pod environment variables, so rotating
    # the Secret alone does not re-render config — bump any spec field
    # (the operator re-reconciles on CHI change) to roll a rotation out.
    users = optional(list(object({
      # User name (becomes the XML section name and the Secret key).
      name = string

      # Password. Accepts a managed-secret reference or a reference to
      # another resource's output. Empty = the module generates one (32
      # letters and digits, generated once and kept across applies) and
      # writes it to the auth Secret under the user's name, where every
      # client reads it (`auth_secret_name` output, key = this user's
      # name). Every named user authenticates with a password: the kind
      # never provisions a passwordless named user.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      password = optional(string, "")

      # Settings profile this user runs under (a name from `profiles`, or
      # a profile defined via `files`/upstream defaults such as
      # "default").
      profile = optional(string, "")

      # Quota applied to this user (a name from `quotas`).
      quota = optional(string, "")

      # Networks the user may connect from, as IPs/CIDRs (e.g.
      # "10.0.0.0/8") or "::/0" for anywhere. KNOW THIS (verified live):
      # empty is NOT ClickHouse's own any-network default — the operator
      # normalizes a user declared without networks to a restrictive
      # fence (the cluster's own pods by host_regexp, plus ::1 and
      # 127.0.0.1), and ClickHouse reports the network rejection as
      # "Authentication failed: password is incorrect, or there is no
      # user with such name" — indistinguishable from a wrong password.
      # Port-forwarded connections arrive as localhost and slip through
      # the fence, which is why a smoke test can pass while every
      # in-cluster client fails. Declare networks explicitly for every
      # user a workload connects as (e.g. "0.0.0.0/0" and "::/0" when the
      # password is the gate).
      networks = optional(list(string), [])

      # SQL GRANT statements executed for the user at config render, e.g.
      # "GRANT SELECT ON analytics.*". The declarative alternative to
      # running GRANTs by hand; requires `access_management` on an admin
      # user only for runtime-managed grants, not for these.
      #
      # Two access behaviors verified live (ClickHouse 25.3): a user
      # declared with NO grants receives ClickHouse's default,
      # UNRESTRICTED access (config-file user semantics) — declare grants
      # to constrain a user, never to widen one. And once grants are
      # declared, distributed DDL (`... ON CLUSTER`) additionally requires
      # "GRANT CLUSTER ON *.*" — CREATE on the target database alone is
      # rejected with ACCESS_DENIED for ON CLUSTER statements.
      grants = optional(list(string), [])

      # Allow this user to manage users and grants at runtime via SQL
      # (CREATE USER / GRANT). Reserve for administrative users.
      access_management = optional(bool, false)

      # Per-user setting overrides. Keys are `/`-separated paths within
      # the user's XML section, e.g. "max_memory_usage" = "10000000000".
      settings = optional(map(string), {})
    })), [])

    # Settings profiles: named bundles of query-level settings users
    # reference by profile name (e.g. a "readonly" profile with
    # `readonly: "1"`). Keys are `/`-separated setting paths within the
    # profile, exactly as the upstream CRD's `profiles` section takes
    # them.
    profiles = optional(list(object({
      # Bundle name (profile name or quota name) referenced from `users`.
      name = string

      # Path-keyed values within the bundle, exactly as the upstream CRD
      # takes them (profiles: setting paths; quotas: "interval/…" paths).
      settings = optional(map(string), {})
    })), [])

    # Quotas: named resource-consumption limits users reference by quota
    # name. Keys are `/`-separated paths within the quota, e.g.
    # "interval/duration" = "3600", "interval/queries" = "10000".
    quotas = optional(list(object({
      # Bundle name (profile name or quota name) referenced from `users`.
      name = string

      # Path-keyed values within the bundle, exactly as the upstream CRD
      # takes them (profiles: setting paths; quotas: "interval/…" paths).
      settings = optional(map(string), {})
    })), [])

    # Server-level settings rendered into config.d for every host. Keys
    # are `/`-separated XML paths exactly as the upstream CRD's
    # `settings` section takes them, e.g.
    # "compression/case/method" = "zstd",
    # "merge_tree/max_suspicious_broken_parts" = "5",
    # "max_concurrent_queries" = "200".
    settings = optional(map(string), {})

    # Raw configuration file drop-ins: file name → full file content.
    # Rendered into the generated ConfigMaps alongside the operator's
    # own files. A file name may carry the upstream's placement prefix
    # ({common}, {users}, {hosts}) to choose between config.d, users.d
    # and conf.d — unprefixed names land in config.d.
    files = optional(map(string), {})

    # Secure distributed queries between this cluster's own hosts with
    # an operator-generated shared secret (the CHI `secret.auto`
    # mechanism). Default true; rendered only when the topology has more
    # than one host. Disable only for ClickHouse versions below 20.10,
    # which predate the mechanism.
    auto_inter_node_secret = optional(bool)

    # Never schedule two replicas of the same shard on the same
    # Kubernetes node (the operator's ShardAntiAffinity pod
    # distribution). Off by default so single-node dev clusters
    # schedule; turn on in production — co-located replicas make
    # replication pointless against node loss.
    spread_replicas_across_nodes = optional(bool, false)

    # How many of this cluster's pods may be voluntarily evicted at
    # once (PodDisruptionBudget maxUnavailable, one PDB per cluster,
    # operator-managed). Default 1.
    pdb_max_unavailable = optional(number)

    # Node selector for every ClickHouse host pod.
    node_selector = optional(map(string), {})

    # Tolerations for every ClickHouse host pod.
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

    # Annotations for the cluster-wide client Service `clickhouse-<name>`
    # (internal LB or service-mesh annotations). The Service stays
    # ClusterIP — compose external exposure from first-class kinds.
    service_annotations = optional(map(string), {})

    # Stop the cluster without losing data (the CHI `stop` verb): true
    # scales every host StatefulSet to zero — pods and Services go away,
    # every PVC stays — and false brings the same data back. The
    # declarative pause switch for expensive dev/staging clusters.
    stopped = optional(bool, false)

    # Names of image-pull secrets (in the deployment namespace) for
    # pulling images from a private mirror.
    image_pull_secrets = optional(list(string), [])
  })
}
