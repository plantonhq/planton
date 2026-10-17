# Computed values for the KubernetesClickHouse module. Every resolution
# here has an exact twin in the Pulumi module — keep them in lockstep:
# same rendered CR bodies, same outputs.
#
# HCL DISCIPLINE (applies to every conditional object in this file):
# conditional keys are contributed as merge() of `cond ? { key = value } : {}`
# singleton maps, or written as `key = cond ? value : null` inside ONE object
# literal pruned with `{ for k, v in {...} : k => v if v != null }`. The
# tempting alternative — a ternary whose branches are differently-shaped
# objects — fails plan-time type unification. The singleton/null-prune forms
# preserve every value's type: counts render as YAML numbers, the CRD's
# StringBool fields ("true", "yes") as strings, access_management as the
# number 1.
#
# Optional nested blocks are read with try(): HCL's && does NOT
# short-circuit, so chained null checks still dereference the null.

locals {
  # The CHI name is metadata.name — the naming root the operator derives
  # every object from: host StatefulSets/Services
  # `chi-<name>-<cluster>-<shard>-<replica>`, the per-cluster Service
  # `cluster-<name>-<cluster>`, and the cluster-wide client Service
  # `clickhouse-<name>` the outputs point at.
  chi_name  = var.metadata.name
  namespace = var.spec.namespace

  # Resource-identity labels stamped on the module-created objects
  # (namespace, auth Secret, CHK, CHI). The operator derives ITS objects'
  # identity from the CHI/CHK names; these labels tie the family back to
  # the Planton resource.
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesClickHouse"
    },
    var.metadata.id != null && var.metadata.id != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    var.metadata.org != null && var.metadata.org != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  cluster_name = try(coalesce(var.spec.cluster_name), "main")
  shards       = try(coalesce(var.spec.shards), 1)
  replicas     = try(coalesce(var.spec.replicas), 1)
  total_hosts  = local.shards * local.replicas

  # The CHI podTemplate takes ONE image string — the shared ContainerImage
  # folds into `repo:tag`, defaulting to the official server image at the
  # spec's pinned version (never the operator's implicit `latest`).
  image_repo = try(var.spec.image.repo, "") != "" ? var.spec.image.repo : "clickhouse/clickhouse-server"
  image_tag  = try(var.spec.image.tag, "") != "" ? var.spec.image.tag : var.spec.version
  image      = "${local.image_repo}:${local.image_tag}"

  # image.pull_secret_name joins image_pull_secrets, deduplicated — a
  # private image override naturally travels with its own pull secret.
  image_pull_secret_names = distinct(concat(
    try(var.spec.image_pull_secrets, []),
    try(var.spec.image.pull_secret_name, "") != "" ? [var.spec.image.pull_secret_name] : []
  ))

  # ---- coordination resolution --------------------------------------------
  # UNSET/unspecified = auto: a managed Keeper whenever the topology needs
  # coordination (replicas > 1 or shards > 1 — replication AND `ON CLUSTER`
  # DDL both depend on it), none otherwise. Explicit managed_keeper always
  # deploys; external/none never do.
  coordination_type   = try(coalesce(var.spec.coordination.type), "")
  coordination_auto   = local.coordination_type == "" || local.coordination_type == "unspecified"
  coordination_needed = local.replicas > 1 || local.shards > 1
  keeper_managed = (
    local.coordination_type == "managed_keeper" ||
    (local.coordination_auto && local.coordination_needed)
  )
  coordination_external = contains(["external_keeper", "external_zookeeper"], local.coordination_type)

  # Operator naming contract for CHK: the client Service is
  # `keeper-<chk-name>`.
  keeper_name         = local.keeper_managed ? "${local.chi_name}-keeper" : ""
  keeper_service_name = local.keeper_managed ? "keeper-${local.chi_name}-keeper" : ""

  keeper_replicas      = try(coalesce(var.spec.coordination.keeper.replicas), 3)
  keeper_disk_size     = try(coalesce(var.spec.coordination.keeper.disk_size), "10Gi")
  keeper_storage_class = try(var.spec.coordination.keeper.storage_class, "")

  # ---- outputs-facing handles ---------------------------------------------
  # Operator naming contract for CHI: the cluster-wide client Service is
  # `clickhouse-<name>`, native protocol on 9000, HTTP interface on 8123.
  service_name         = "clickhouse-${local.chi_name}"
  tcp_endpoint         = "${local.service_name}.${local.namespace}.svc.cluster.local:9000"
  http_endpoint        = "http://${local.service_name}.${local.namespace}.svc.cluster.local:8123"
  port_forward_command = "kubectl port-forward svc/${local.service_name} -n ${local.namespace} 8123:8123"

  # Passwords never appear in the CHI: the module writes them into this
  # Secret (one key per user name) and the CHI references the keys.
  auth_secret_name = length(try(var.spec.users, [])) > 0 ? "${local.chi_name}-clickhouse-auth" : ""

  # Users split by where their password comes from: declared in the spec
  # (already resolved to a literal) or generated by the module.
  generated_password_users = {
    for u in try(var.spec.users, []) : u.name => u
    if try(coalesce(u.password), "") == ""
  }
  declared_passwords = {
    for u in try(var.spec.users, []) : u.name => u.password
    if try(coalesce(u.password), "") != ""
  }

  # ---- shared resources body (requests/limits) ----------------------------
  server_resources = try(var.spec.resources, null) == null ? null : {
    for k, v in {
      limits = try(var.spec.resources.limits, null) == null ? null : {
        for rk, rv in {
          cpu    = var.spec.resources.limits.cpu
          memory = var.spec.resources.limits.memory
        } : rk => rv if rv != null && rv != ""
      }
      requests = try(var.spec.resources.requests, null) == null ? null : {
        for rk, rv in {
          cpu    = var.spec.resources.requests.cpu
          memory = var.spec.resources.requests.memory
        } : rk => rv if rv != null && rv != ""
      }
    } : k => v if v != null
  }

  keeper_resources = try(var.spec.coordination.keeper.resources, null) == null ? null : {
    for k, v in {
      limits = try(var.spec.coordination.keeper.resources.limits, null) == null ? null : {
        for rk, rv in {
          cpu    = var.spec.coordination.keeper.resources.limits.cpu
          memory = var.spec.coordination.keeper.resources.limits.memory
        } : rk => rv if rv != null && rv != ""
      }
      requests = try(var.spec.coordination.keeper.resources.requests, null) == null ? null : {
        for rk, rv in {
          cpu    = var.spec.coordination.keeper.resources.requests.cpu
          memory = var.spec.coordination.keeper.resources.requests.memory
        } : rk => rv if rv != null && rv != ""
      }
    } : k => v if v != null
  }

  tolerations_body = [
    for t in try(var.spec.tolerations, []) : {
      for tk, tv in {
        key               = try(t.key, "") != "" ? t.key : null
        operator          = try(t.operator, "") != "" ? t.operator : null
        value             = try(t.value, "") != "" ? t.value : null
        effect            = try(t.effect, "") != "" ? t.effect : null
        tolerationSeconds = try(t.toleration_seconds, null)
      } : tk => tv if tv != null
    }
  ]

  # ---- configuration.zookeeper --------------------------------------------
  # Managed: the CHI's native keeper reference — the operator resolves the
  # CHK endpoints itself. External: explicit nodes plus optional root znode
  # and digest identity. None / auto-without-need: omitted entirely.
  zookeeper_singleton = merge(
    local.keeper_managed ? {
      zookeeper = { keeper = { name = local.keeper_name } }
    } : {},
    local.coordination_external ? {
      zookeeper = {
        for k, v in {
          nodes = [
            for n in try(var.spec.coordination.external.nodes, []) : {
              host = n.host
              port = try(coalesce(n.port), 2181)
            }
          ]
          root     = try(var.spec.coordination.external.root, "") != "" ? var.spec.coordination.external.root : null
          identity = try(var.spec.coordination.external.identity, "") != "" ? var.spec.coordination.external.identity : null
        } : k => v if v != null
      }
    } : {}
  )

  # ---- configuration.users --------------------------------------------------
  # The CHI users section is path-keyed ("<user>/password", "<user>/profile",
  # …). Passwords ride valueFrom.secretKeyRef into the auth Secret;
  # access_management renders as the number 1 (the section's XML flag).
  # The leading {} seeds merge(): Terraform evaluates BOTH arms of the
  # gating conditionals below, and merge() with zero arguments errors.
  users_body = merge({}, [
    for u in try(var.spec.users, []) : merge(
      {
        "${u.name}/password" = {
          valueFrom = {
            secretKeyRef = {
              name = local.auth_secret_name
              key  = u.name
            }
          }
        }
      },
      try(u.profile, "") != "" ? { "${u.name}/profile" = u.profile } : {},
      try(u.quota, "") != "" ? { "${u.name}/quota" = u.quota } : {},
      length(try(u.networks, [])) > 0 ? { "${u.name}/networks/ip" = u.networks } : {},
      length(try(u.grants, [])) > 0 ? { "${u.name}/grants/query" = u.grants } : {},
      try(u.access_management, false) ? { "${u.name}/access_management" = 1 } : {},
      { for sk, sv in try(u.settings, {}) : "${u.name}/${sk}" => sv }
    )
  ]...)

  # profiles/quotas flatten the named bundles into the CRD's path-keyed
  # form: "<bundle>/<path>" = value.
  profiles_body = merge({}, [
    for p in try(var.spec.profiles, []) : {
      for sk, sv in try(p.settings, {}) : "${p.name}/${sk}" => sv
    }
  ]...)

  quotas_body = merge({}, [
    for q in try(var.spec.quotas, []) : {
      for sk, sv in try(q.settings, {}) : "${q.name}/${sk}" => sv
    }
  ]...)

  # ---- configuration.clusters[0] --------------------------------------------
  # secret.auto (StringBool) renders only when the topology has more than
  # one host AND the spec keeps auto_inter_node_secret (default true).
  cluster_secret_enabled = try(coalesce(var.spec.auto_inter_node_secret), true) && local.total_hosts > 1

  cluster_body = merge(
    {
      name = local.cluster_name
      layout = {
        shardsCount   = local.shards
        replicasCount = local.replicas
      }
      pdbMaxUnavailable = try(coalesce(var.spec.pdb_max_unavailable), 1)
    },
    local.cluster_secret_enabled ? { secret = { auto = "true" } } : {}
  )

  configuration_body = merge(
    { clusters = [local.cluster_body] },
    local.zookeeper_singleton,
    length(try(var.spec.users, [])) > 0 ? { users = local.users_body } : {},
    length(try(var.spec.profiles, [])) > 0 ? { profiles = local.profiles_body } : {},
    length(try(var.spec.quotas, [])) > 0 ? { quotas = local.quotas_body } : {},
    length(try(var.spec.settings, {})) > 0 ? { settings = var.spec.settings } : {},
    length(try(var.spec.files, {})) > 0 ? { files = var.spec.files } : {}
  )

  # ---- spec.defaults ---------------------------------------------------------
  # reclaimPolicy "Retain" renders only on divergence from the operator
  # default (Delete). Template wiring points every host at the module's
  # named templates; the client serviceTemplate only exists (and is only
  # wired) when service_annotations has something to carry.
  client_service_enabled = length(try(var.spec.service_annotations, {})) > 0
  log_volume_enabled     = try(var.spec.log_disk_size, "") != ""

  defaults_body = merge(
    {
      templates = merge(
        {
          podTemplate             = "server"
          dataVolumeClaimTemplate = "data"
        },
        local.log_volume_enabled ? { logVolumeClaimTemplate = "logs" } : {},
        local.client_service_enabled ? { serviceTemplate = "client" } : {}
      )
    },
    try(var.spec.retain_volumes_on_delete, false) ? { storageManagement = { reclaimPolicy = "Retain" } } : {}
  )

  # ---- spec.templates ---------------------------------------------------------
  server_pod_spec = merge(
    {
      containers = [
        merge(
          {
            name  = "clickhouse"
            image = local.image
          },
          local.server_resources != null ? { resources = local.server_resources } : {}
        )
      ]
    },
    length(try(var.spec.node_selector, {})) > 0 ? { nodeSelector = var.spec.node_selector } : {},
    length(local.tolerations_body) > 0 ? { tolerations = local.tolerations_body } : {},
    length(local.image_pull_secret_names) > 0 ? {
      imagePullSecrets = [for s in local.image_pull_secret_names : { name = s }]
    } : {}
  )

  # ShardAntiAffinity keeps replicas of the same shard on distinct nodes —
  # rendered only when the spec asks (single-node dev clusters must
  # schedule).
  server_pod_template = merge(
    {
      name = "server"
      spec = local.server_pod_spec
    },
    try(var.spec.spread_replicas_across_nodes, false) ? {
      podDistribution = [{ type = "ShardAntiAffinity" }]
    } : {}
  )

  volume_claim_templates = concat(
    [
      {
        name = "data"
        spec = merge(
          {
            accessModes = ["ReadWriteOnce"]
            resources   = { requests = { storage = var.spec.disk_size } }
          },
          try(var.spec.storage_class, "") != "" ? { storageClassName = var.spec.storage_class } : {}
        )
      }
    ],
    local.log_volume_enabled ? [
      {
        name = "logs"
        spec = merge(
          {
            accessModes = ["ReadWriteOnce"]
            resources   = { requests = { storage = var.spec.log_disk_size } }
          },
          try(var.spec.storage_class, "") != "" ? { storageClassName = var.spec.storage_class } : {}
        )
      }
    ] : []
  )

  # generateName pins the operator's own default ("clickhouse-{chi}") so
  # annotating the Service never renames it — the service_name output and
  # every client depend on that name. Type stays ClusterIP by design.
  # The template's spec is copied VERBATIM by the operator — no port
  # defaulting — so the standard interface ports must be declared
  # explicitly: a ClusterIP Service with zero ports is rejected by the
  # API server and the client Service simply never appears (verified
  # live).
  service_templates = [
    {
      name         = "client"
      generateName = "clickhouse-{chi}"
      metadata     = { annotations = var.spec.service_annotations }
      spec = {
        type = "ClusterIP"
        ports = [
          { name = "http", port = 8123 },
          { name = "tcp", port = 9000 },
        ]
      }
    }
  ]

  templates_body = merge(
    {
      podTemplates         = [local.server_pod_template]
      volumeClaimTemplates = local.volume_claim_templates
    },
    local.client_service_enabled ? { serviceTemplates = local.service_templates } : {}
  )

  # ---- the ClickHouseInstallation CR ------------------------------------------
  # stop is the CRD's StringBool pause switch: "yes" scales every host
  # StatefulSet to zero while keeping every PVC.
  chi_manifest = {
    apiVersion = "clickhouse.altinity.com/v1"
    kind       = "ClickHouseInstallation"
    metadata = {
      name      = local.chi_name
      namespace = local.namespace
      labels    = local.labels
    }
    spec = merge(
      {
        configuration = local.configuration_body
        defaults      = local.defaults_body
        templates     = local.templates_body
      },
      try(var.spec.stopped, false) ? { stop = "yes" } : {}
    )
  }

  # ---- the ClickHouseKeeperInstallation CR -------------------------------------
  # Same operator, sibling CRD. The pod template ALWAYS renders, and its
  # container ALWAYS carries an explicit image: the Keeper container is
  # declared explicitly so the image pins to the resource's own version
  # line instead of the operator's fallback (`latest`) — Keeper images
  # are published in lockstep with server releases and the protocol is
  # compatible across them. An explicit container entry SUPPRESSES the
  # operator's default-image injection entirely — verified live: a pod
  # template carrying only `resources` produced a StatefulSet the API
  # server rejected with `containers[0].image: Required value`, and the
  # keeper never came up. The data volumeClaimTemplate always renders —
  # coordination logs and snapshots must survive pod restarts or the
  # ensemble loses state.
  keeper_container = merge(
    {
      name  = "clickhouse-keeper"
      image = "clickhouse/clickhouse-keeper:${var.spec.version}"
    },
    local.keeper_resources != null ? { resources = local.keeper_resources } : {}
  )

  keeper_templates_body = {
    volumeClaimTemplates = [
      {
        name = "data"
        spec = merge(
          {
            accessModes = ["ReadWriteOnce"]
            resources   = { requests = { storage = local.keeper_disk_size } }
          },
          local.keeper_storage_class != "" ? { storageClassName = local.keeper_storage_class } : {}
        )
      }
    ]
    podTemplates = [
      {
        name = "keeper"
        spec = {
          containers = [local.keeper_container]
        }
      }
    ]
  }

  keeper_manifest = {
    apiVersion = "clickhouse-keeper.altinity.com/v1"
    kind       = "ClickHouseKeeperInstallation"
    metadata = {
      name      = local.keeper_name
      namespace = local.namespace
      labels    = local.labels
    }
    spec = {
      configuration = {
        clusters = [
          {
            name = "keeper"
            layout = {
              replicasCount = local.keeper_replicas
            }
          }
        ]
      }
      defaults = {
        templates = {
          dataVolumeClaimTemplate = "data"
          podTemplate             = "keeper"
        }
      }
      templates = local.keeper_templates_body
    }
  }
}
