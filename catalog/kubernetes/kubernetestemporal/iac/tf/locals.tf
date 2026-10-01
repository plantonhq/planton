# Computed values for the KubernetesTemporal module. Every resolution here
# has an exact twin in the Pulumi module's locals.go / values.go — keep
# them in lockstep.
#
# SECRET DISCIPLINE (load-bearing): nothing in this module transports
# credential material. Every database password rides the chart's
# existingSecret contract — the chart wires a secretKeyRef into the server
# and schema-Job pods and STRIPS the Helm-side keys before writing the
# server config; because existingSecret is always set, the chart's own
# per-store password Secret (which would embed an inline password) is
# never created.
#
# HCL DISCIPLINE (applies to every conditional object in this file):
# conditional entries are written as `key = cond ? value : null` inside ONE
# object literal, pruned with `{ for k, v in {...} : k => v if v != null }`.
# `cond ? {...} : {}` ternaries fail plan-time type unification when
# branches carry different attributes. The database oneof is normalized at
# the SCALAR level first (scalar ternaries always unify) and the store
# blocks are built ONCE from the normalized values — never a ternary
# between two differently-pruned arm objects. Optional nested blocks are
# read with try() (HCL's && does NOT short-circuit); optional scalars
# inside optional blocks with try(coalesce(x), null).

locals {
  # Chart identity — must stay byte-identical with the Pulumi module's
  # vars. The temporal chart is served from go.temporal.io/helm-charts;
  # chart 1.6.0 ships Temporal 1.31.2.
  helm_chart_name = "temporal"
  helm_chart_repo = "https://go.temporal.io/helm-charts"

  # Release name — metadata.name, NOT a fixed chart name: several
  # Temporal clusters can coexist in one Kubernetes cluster.
  # fullnameOverride below pins every chart child name to this.
  release_name = var.metadata.name

  # Chart version resolved to the pinned default when unset, so both
  # engines install the same chart whether or not the platform's
  # defaulting middleware ran.
  chart_version = coalesce(var.spec.chart_version, "1.6.0")

  namespace = var.spec.namespace

  # Resource-identity labels: stamped on the module-created satellites
  # (the namespace) and on every pod the chart runs, through the chart's
  # own pod-label values (server.podLabels, web.podLabels,
  # admintools.podLabels, schema.podLabels) -- so a log line, a metric or
  # an alert from any Temporal pod names its organization and environment.
  # The chart's selectors are its own fixed labels; these never reach them.
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesTemporal"
    },
    var.metadata.id != null && var.metadata.id != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    var.metadata.org != null && var.metadata.org != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # The frontend Service is `<fullname>-frontend`; the Web UI Service
  # `<fullname>-web` (the chart appends each component's name to the
  # fullname, which fullnameOverride pins to the resource name). Ports
  # are the chart's defaults. Feed the exported handles.
  web_ui_enabled        = try(coalesce(var.spec.web_ui.enabled), null) != null ? var.spec.web_ui.enabled : true
  frontend_service_name = "${local.release_name}-frontend"
  web_ui_service_name   = local.web_ui_enabled ? "${local.release_name}-web" : ""

  # ---- the database oneof, normalized at the scalar level -------------------
  # Exactly one backend is declared (proto oneof). The SQL arms
  # (postgres/mysql) normalize into one set of scalars so the sql store
  # block is built ONCE — scalar ternaries always type-unify; ternaries
  # between two pruned arm OBJECTS do not (the phantom branch is still
  # type-checked).
  pg_declared   = try(var.spec.database.postgres, null) != null
  my_declared   = try(var.spec.database.mysql, null) != null
  cass_declared = try(var.spec.database.cassandra, null) != null
  sql_declared  = local.pg_declared || local.my_declared

  # createDatabase/manageSchema are ALWAYS rendered explicitly — the
  # chart's getStore helper silently defaults BOTH to true when unset,
  # and an unintended create-database attempt fails against
  # least-privilege users.
  create_databases = var.spec.database.create_databases
  manage_schema    = !var.spec.database.skip_schema_setup

  database_name            = coalesce(var.spec.database.database_name, "temporal")
  visibility_database_name = coalesce(var.spec.database.visibility_database_name, "temporal_visibility")

  # The pluginName doubles as the driverName (the chart's documented
  # postgres12/mysql8 SQL plugin names).
  sql_plugin = local.pg_declared ? "postgres12" : "mysql8"
  sql_host   = local.pg_declared ? try(var.spec.database.postgres.host, "") : try(var.spec.database.mysql.host, "")
  sql_port = local.pg_declared ? (
    try(coalesce(var.spec.database.postgres.port), null) != null ? var.spec.database.postgres.port : 5432
    ) : (
    try(coalesce(var.spec.database.mysql.port), null) != null ? var.spec.database.mysql.port : 3306
  )
  sql_user           = local.pg_declared ? try(var.spec.database.postgres.username, "") : try(var.spec.database.mysql.username, "")
  sql_secret_name    = local.pg_declared ? try(var.spec.database.postgres.password_secret.secret_name, "") : try(var.spec.database.mysql.password_secret.secret_name, "")
  sql_secret_key     = local.pg_declared ? (try(coalesce(var.spec.database.postgres.password_secret.secret_key), null) != null ? var.spec.database.postgres.password_secret.secret_key : "password") : try(var.spec.database.mysql.password_secret.secret_key, "")
  sql_max_conns      = local.pg_declared ? (try(coalesce(var.spec.database.postgres.max_conns), null) != null ? var.spec.database.postgres.max_conns : 20) : (try(coalesce(var.spec.database.mysql.max_conns), null) != null ? var.spec.database.mysql.max_conns : 20)
  sql_max_idle_conns = local.pg_declared ? (try(coalesce(var.spec.database.postgres.max_idle_conns), null) != null ? var.spec.database.postgres.max_idle_conns : 20) : (try(coalesce(var.spec.database.mysql.max_idle_conns), null) != null ? var.spec.database.mysql.max_idle_conns : 20)
  sql_conn_lifetime  = local.pg_declared ? (try(coalesce(var.spec.database.postgres.max_conn_lifetime), null) != null ? var.spec.database.postgres.max_conn_lifetime : "1h") : (try(coalesce(var.spec.database.mysql.max_conn_lifetime), null) != null ? var.spec.database.mysql.max_conn_lifetime : "1h")
  # Both arms share the KubernetesTemporalDatabaseTls shape — the
  # ternary unifies (identical generated object types).
  sql_tls = local.pg_declared ? try(var.spec.database.postgres.tls, null) : try(var.spec.database.mysql.tls, null)

  # ---- store TLS blocks (shared shape; null when not declared) ---------------
  sql_tls_block = local.sql_tls == null ? null : {
    for k, v in {
      enabled                = local.sql_tls.enabled
      enableHostVerification = local.sql_tls.host_verification
      serverName             = local.sql_tls.server_name != "" ? local.sql_tls.server_name : null
    } : k => v if v != null
  }

  cass_tls = local.cass_declared ? try(var.spec.database.cassandra.tls, null) : null
  cass_tls_block = local.cass_tls == null ? null : {
    for k, v in {
      enabled                = local.cass_tls.enabled
      enableHostVerification = local.cass_tls.host_verification
      serverName             = local.cass_tls.server_name != "" ? local.cass_tls.server_name : null
    } : k => v if v != null
  }

  # ---- the default store ------------------------------------------------------
  # connectAddr carries host:port — the admintools/schema-Job env
  # template REQUIRES that form (it parses SQL_HOST and SQL_PORT out of
  # it). existingSecret/secretKey are Helm-side keys: the chart wires
  # the secretKeyRef and strips them from the rendered server config.
  sql_default_block = local.sql_declared ? {
    for k, v in {
      pluginName      = local.sql_plugin
      driverName      = local.sql_plugin
      databaseName    = local.database_name
      connectAddr     = "${local.sql_host}:${local.sql_port}"
      connectProtocol = "tcp"
      user            = local.sql_user
      existingSecret  = local.sql_secret_name
      secretKey       = local.sql_secret_key
      maxConns        = local.sql_max_conns
      maxIdleConns    = local.sql_max_idle_conns
      maxConnLifetime = local.sql_conn_lifetime
      createDatabase  = local.create_databases
      manageSchema    = local.manage_schema
      tls             = local.sql_tls_block
    } : k => v if v != null
  } : null

  # hosts is a comma-joined string (the chart's own documented form; its
  # env template takes the first for the schema tools).
  cassandra_block = local.cass_declared ? {
    for k, v in {
      hosts             = join(",", var.spec.database.cassandra.hosts)
      port              = try(coalesce(var.spec.database.cassandra.port), null) != null ? var.spec.database.cassandra.port : 9042
      keyspace          = local.database_name
      user              = var.spec.database.cassandra.username
      existingSecret    = var.spec.database.cassandra.password_secret.secret_name
      secretKey         = try(coalesce(var.spec.database.cassandra.password_secret.secret_key), null) != null ? var.spec.database.cassandra.password_secret.secret_key : "password"
      replicationFactor = try(coalesce(var.spec.database.cassandra.replication_factor), null) != null ? var.spec.database.cassandra.replication_factor : 3
      datacenter        = var.spec.database.cassandra.datacenter != "" ? var.spec.database.cassandra.datacenter : null
      createDatabase    = local.create_databases
      manageSchema      = local.manage_schema
      tls               = local.cass_tls_block
    } : k => v if v != null
  } : null

  datastore_default = {
    for k, v in {
      sql       = local.sql_default_block
      cassandra = local.cassandra_block
    } : k => v if v != null
  }

  # ---- the visibility store ------------------------------------------------------
  # The dedicated `visibility` block when declared (REQUIRED for the
  # cassandra arm — CEL-enforced); otherwise the default store's own
  # connection pointed at the visibility database. Normalized at the
  # scalar level exactly like the default store.
  vis_declared    = try(var.spec.database.visibility, null) != null
  vis_pg_declared = try(var.spec.database.visibility.postgres, null) != null

  vis_database_name = local.vis_declared ? (
    try(coalesce(var.spec.database.visibility.database_name), null) != null ? var.spec.database.visibility.database_name : local.visibility_database_name
  ) : local.visibility_database_name

  vis_plugin = local.vis_declared ? (local.vis_pg_declared ? "postgres12" : "mysql8") : local.sql_plugin
  vis_host = local.vis_declared ? (
    local.vis_pg_declared ? try(var.spec.database.visibility.postgres.host, "") : try(var.spec.database.visibility.mysql.host, "")
  ) : local.sql_host
  vis_port = local.vis_declared ? (
    local.vis_pg_declared ? (
      try(coalesce(var.spec.database.visibility.postgres.port), null) != null ? var.spec.database.visibility.postgres.port : 5432
      ) : (
      try(coalesce(var.spec.database.visibility.mysql.port), null) != null ? var.spec.database.visibility.mysql.port : 3306
    )
  ) : local.sql_port
  vis_user = local.vis_declared ? (
    local.vis_pg_declared ? try(var.spec.database.visibility.postgres.username, "") : try(var.spec.database.visibility.mysql.username, "")
  ) : local.sql_user
  vis_secret_name = local.vis_declared ? (
    local.vis_pg_declared ? try(var.spec.database.visibility.postgres.password_secret.secret_name, "") : try(var.spec.database.visibility.mysql.password_secret.secret_name, "")
  ) : local.sql_secret_name
  vis_secret_key = local.vis_declared ? (
    local.vis_pg_declared ? (
      try(coalesce(var.spec.database.visibility.postgres.password_secret.secret_key), null) != null ? var.spec.database.visibility.postgres.password_secret.secret_key : "password"
    ) : try(var.spec.database.visibility.mysql.password_secret.secret_key, "")
  ) : local.sql_secret_key
  vis_max_conns = local.vis_declared ? (
    local.vis_pg_declared ? (
      try(coalesce(var.spec.database.visibility.postgres.max_conns), null) != null ? var.spec.database.visibility.postgres.max_conns : 20
      ) : (
      try(coalesce(var.spec.database.visibility.mysql.max_conns), null) != null ? var.spec.database.visibility.mysql.max_conns : 20
    )
  ) : local.sql_max_conns
  vis_max_idle_conns = local.vis_declared ? (
    local.vis_pg_declared ? (
      try(coalesce(var.spec.database.visibility.postgres.max_idle_conns), null) != null ? var.spec.database.visibility.postgres.max_idle_conns : 20
      ) : (
      try(coalesce(var.spec.database.visibility.mysql.max_idle_conns), null) != null ? var.spec.database.visibility.mysql.max_idle_conns : 20
    )
  ) : local.sql_max_idle_conns
  vis_conn_lifetime = local.vis_declared ? (
    local.vis_pg_declared ? (
      try(coalesce(var.spec.database.visibility.postgres.max_conn_lifetime), null) != null ? var.spec.database.visibility.postgres.max_conn_lifetime : "1h"
      ) : (
      try(coalesce(var.spec.database.visibility.mysql.max_conn_lifetime), null) != null ? var.spec.database.visibility.mysql.max_conn_lifetime : "1h"
    )
  ) : local.sql_conn_lifetime
  vis_tls = local.vis_declared ? (
    local.vis_pg_declared ? try(var.spec.database.visibility.postgres.tls, null) : try(var.spec.database.visibility.mysql.tls, null)
  ) : local.sql_tls

  vis_tls_block = local.vis_tls == null ? null : {
    for k, v in {
      enabled                = local.vis_tls.enabled
      enableHostVerification = local.vis_tls.host_verification
      serverName             = local.vis_tls.server_name != "" ? local.vis_tls.server_name : null
    } : k => v if v != null
  }

  datastore_visibility = {
    sql = {
      for k, v in {
        pluginName      = local.vis_plugin
        driverName      = local.vis_plugin
        databaseName    = local.vis_database_name
        connectAddr     = "${local.vis_host}:${local.vis_port}"
        connectProtocol = "tcp"
        user            = local.vis_user
        existingSecret  = local.vis_secret_name
        secretKey       = local.vis_secret_key
        maxConns        = local.vis_max_conns
        maxIdleConns    = local.vis_max_idle_conns
        maxConnLifetime = local.vis_conn_lifetime
        createDatabase  = local.create_databases
        manageSchema    = local.manage_schema
        tls             = local.vis_tls_block
      } : k => v if v != null
    }
  }

  # ---- per-service sizing --------------------------------------------------------
  services_or_null = {
    frontend = try(var.spec.services.frontend, null)
    history  = try(var.spec.services.history, null)
    matching = try(var.spec.services.matching, null)
    worker   = try(var.spec.services.worker, null)
  }

  rendered_services = {
    for name, s in local.services_or_null : name => s == null ? null : {
      for k, v in {
        replicaCount = try(coalesce(s.replicas), null)
        resources = try(s.resources, null) == null ? null : {
          for rk, rv in {
            requests = try(s.resources.requests, null) == null ? null : {
              for qk, qv in {
                cpu    = s.resources.requests.cpu != "" ? s.resources.requests.cpu : null
                memory = s.resources.requests.memory != "" ? s.resources.requests.memory : null
              } : qk => qv if qv != null
            }
            limits = try(s.resources.limits, null) == null ? null : {
              for lk, lv in {
                cpu    = s.resources.limits.cpu != "" ? s.resources.limits.cpu : null
                memory = s.resources.limits.memory != "" ? s.resources.limits.memory : null
              } : lk => lv if lv != null
            }
          } : rk => rv if rv != null && rv != {}
        }
      } : k => v if v != null && v != {}
    }
  }

  # ---- web UI / admin-tools resources ----------------------------------------------
  web_resources = try(var.spec.web_ui.resources, null) == null ? null : {
    for rk, rv in {
      requests = try(var.spec.web_ui.resources.requests, null) == null ? null : {
        for qk, qv in {
          cpu    = var.spec.web_ui.resources.requests.cpu != "" ? var.spec.web_ui.resources.requests.cpu : null
          memory = var.spec.web_ui.resources.requests.memory != "" ? var.spec.web_ui.resources.requests.memory : null
        } : qk => qv if qv != null
      }
      limits = try(var.spec.web_ui.resources.limits, null) == null ? null : {
        for lk, lv in {
          cpu    = var.spec.web_ui.resources.limits.cpu != "" ? var.spec.web_ui.resources.limits.cpu : null
          memory = var.spec.web_ui.resources.limits.memory != "" ? var.spec.web_ui.resources.limits.memory : null
        } : lk => lv if lv != null
      }
    } : rk => rv if rv != null && rv != {}
  }

  # ---- declarative Temporal namespaces ----------------------------------------------
  temporal_namespaces_declared = length(try(var.spec.temporal_namespaces, [])) > 0
  temporal_namespaces_block = local.temporal_namespaces_declared ? {
    create = true
    namespace = [
      for ns in var.spec.temporal_namespaces : {
        name      = ns.name
        retention = try(coalesce(ns.retention), null) != null ? ns.retention : "3d"
      }
    ]
  } : null

  # ---- dynamic-config limits ----------------------------------------------------------
  # Keys verified against the server source at the pin
  # (common/dynamicconfig/constants.go). Each key takes a list of
  # {value, constraints} entries; empty constraints = global.
  dynamic_config_declared = try(var.spec.dynamic_config, null) != null
  dynamic_config_block = local.dynamic_config_declared ? {
    for k, v in {
      "limit.historySize.error"  = try(coalesce(var.spec.dynamic_config.history_size_limit_error), null)
      "limit.historySize.warn"   = try(coalesce(var.spec.dynamic_config.history_size_limit_warn), null)
      "limit.historyCount.error" = try(coalesce(var.spec.dynamic_config.history_count_limit_error), null)
      "limit.historyCount.warn"  = try(coalesce(var.spec.dynamic_config.history_count_limit_warn), null)
      "limit.blobSize.error"     = try(coalesce(var.spec.dynamic_config.blob_size_limit_error), null)
      "limit.blobSize.warn"      = try(coalesce(var.spec.dynamic_config.blob_size_limit_warn), null)
    } : k => [{ value = v, constraints = {} }] if v != null
  } : null

  # ---- archival --------------------------------------------------------------------------
  # The provider block enables the capability; the namespaceDefaults URIs
  # make every Temporal namespace archive by default. Cloud credentials
  # are ambient (IRSA / workload identity) — nothing credential-bearing
  # renders.
  archival_declared           = try(var.spec.archival, null) != null
  archival_s3_declared        = try(var.spec.archival.s3, null) != null
  archival_gcs_declared       = try(var.spec.archival.gcs, null) != null
  archival_filestore_declared = try(var.spec.archival.filestore, null) != null

  archival_provider = local.archival_declared ? {
    for k, v in {
      s3store  = local.archival_s3_declared ? { region = var.spec.archival.s3.region } : null
      gstorage = local.archival_gcs_declared ? {} : null
      filestore = local.archival_filestore_declared ? {
        fileMode = "0666"
        dirMode  = "0766"
      } : null
    } : k => v if v != null
  } : null

  archival_block = local.archival_declared ? {
    history = {
      state      = "enabled"
      enableRead = true
      provider   = local.archival_provider
    }
    visibility = {
      state      = "enabled"
      enableRead = true
      provider   = local.archival_provider
    }
  } : null

  namespace_defaults_block = local.archival_declared ? {
    archival = {
      history = {
        state = "enabled"
        URI   = var.spec.archival.history_uri
      }
      visibility = {
        state = "enabled"
        URI   = var.spec.archival.visibility_uri
      }
    }
  } : null

  # ---- scheduling (server-wide; web and admintools carry their own copies) ---
  scheduling_block = {
    for k, v in {
      nodeSelector = length(try(var.spec.scheduling.node_selector, {})) > 0 ? var.spec.scheduling.node_selector : null
      tolerations = length(try(var.spec.scheduling.tolerations, [])) > 0 ? [
        for t in var.spec.scheduling.tolerations : {
          for tk, tv in {
            key               = t.key != "" ? t.key : null
            operator          = t.operator != "" ? t.operator : null
            value             = t.value != "" ? t.value : null
            effect            = t.effect != "" ? t.effect : null
            tolerationSeconds = try(t.toleration_seconds, null)
          } : tk => tv if tv != null
        }
      ] : null
    } : k => v if v != null
  }

  # ---- image overrides (the chart's COMBINED image form per component) -------
  images_or_null = {
    server      = try(var.spec.images.server, null)
    web_ui      = try(var.spec.images.web_ui, null)
    admin_tools = try(var.spec.images.admin_tools, null)
  }

  rendered_images = {
    for name, img in local.images_or_null : name => img == null ? null : {
      for k, v in {
        repository = img.repo != "" ? img.repo : null
        tag        = img.tag != "" ? img.tag : null
      } : k => v if v != null
    }
  }

  image_pull_secrets = distinct([
    for img in values(local.images_or_null) : { name = img.pull_secret_name }
    if img != null && try(img.pull_secret_name, "") != ""
  ])

  # ---- the server block ---------------------------------------------------------------------
  server_block = merge(
    {
      for k, v in {
        config = {
          for ck, cv in {
            logLevel = coalesce(var.spec.log_level, "info")
            persistence = {
              defaultStore     = "default"
              visibilityStore  = "visibility"
              numHistoryShards = try(coalesce(var.spec.num_history_shards), null) != null ? var.spec.num_history_shards : 512
              datastores = {
                default    = local.datastore_default
                visibility = local.datastore_visibility
              }
            }
            namespaces = local.temporal_namespaces_block
          } : ck => cv if cv != null
        }
        frontend = local.rendered_services.frontend
        history  = local.rendered_services.history
        matching = local.rendered_services.matching
        worker   = local.rendered_services.worker
        # NOTE the chart key carries a dash.
        "internal-frontend" = var.spec.internal_frontend_enabled ? { enabled = true } : null
        dynamicConfig       = local.dynamic_config_block
        archival            = local.archival_block
        namespaceDefaults   = local.namespace_defaults_block
        metrics             = var.spec.service_monitor_enabled ? { serviceMonitor = { enabled = true } } : null
        image               = local.rendered_images.server
        podLabels           = local.labels # every server service's pods (frontend, history, matching, worker)
      } : k => v if v != null && v != {}
    },
    local.scheduling_block
  )

  # ---- the web UI block -----------------------------------------------------------------------
  # A disabled UI renders only enabled=false; scheduling keys fold in
  # per key (never `enabled ? scheduling : {}` — the type-unification
  # class).
  web_block = {
    for k, v in {
      enabled      = local.web_ui_enabled ? null : false
      replicaCount = local.web_ui_enabled ? try(coalesce(var.spec.web_ui.replicas), null) : null
      resources    = local.web_ui_enabled ? local.web_resources : null
      image        = local.web_ui_enabled ? local.rendered_images.web_ui : null
      nodeSelector = local.web_ui_enabled ? try(local.scheduling_block.nodeSelector, null) : null
      tolerations  = local.web_ui_enabled ? try(local.scheduling_block.tolerations, null) : null
      podLabels    = local.web_ui_enabled ? local.labels : null
    } : k => v if v != null && v != {}
  }

  # ---- admin tools ------------------------------------------------------------------------------
  # The image is needed even with the pod disabled — the schema and
  # namespace Jobs run it.
  admin_tools_enabled = try(coalesce(var.spec.admin_tools_enabled), null) != null ? var.spec.admin_tools_enabled : true
  admintools_block = merge(
    {
      for k, v in {
        enabled   = local.admin_tools_enabled ? null : false
        image     = local.rendered_images.admin_tools
        podLabels = local.admin_tools_enabled ? local.labels : null
      } : k => v if v != null
    },
    local.scheduling_block
  )

  # ---- typed chart values (twin of the Pulumi module's buildHelmValues) ----
  helm_values = {
    for k, v in {
      # fullnameOverride pins every child name (`<name>-frontend`,
      # `<name>-web`, ...) to the resource name; the exported outputs
      # are built from that contract.
      fullnameOverride = local.release_name

      server     = local.server_block
      web        = local.web_block
      admintools = local.admintools_block
      # The schema-setup and namespace Jobs' pods (the chart reads only
      # podLabels from this block here; the Jobs themselves are enabled by
      # the datastores' manageSchema).
      schema = { podLabels = local.labels }

      # 1.29-image compatibility shims: OFF at this pin (the chart
      # defaults both ON for Temporal 1.29 images; our pin runs 1.31+).
      shims = {
        dockerize         = false
        elasticsearchTool = false
      }

      imagePullSecrets = length(local.image_pull_secrets) > 0 ? local.image_pull_secrets : null
    } : k => v if v != null && v != {}
  }
}
