# Computed values for the KubernetesGoFeatureFlag module.
# Every resolution here has an exact twin in the Pulumi module - keep them in
# lockstep: the same relay configuration document byte for byte, the same
# secret environment, the same grants, the same chart values, the same
# outputs.
#
# HCL DISCIPLINE: conditional keys are contributed with the null-prune idiom
# - `key = cond ? value : null` inside one for-comprehension that drops
# nulls. Never `cond ? {} : {...}` ternaries (differently shaped object
# branches fail plan-time type unification). Fields of list elements are read
# with try(), because lists holding free-form values (exporters, flag sets)
# arrive untyped.

locals {
  helm_chart_name       = "relay-proxy"
  helm_chart_repo       = "https://charts.gofeatureflag.org"
  default_chart_version = "1.56.0"
  chart_version         = try(var.spec.chart_version, "") != "" && var.spec.chart_version != null ? var.spec.chart_version : local.default_chart_version

  namespace    = var.spec.namespace
  release_name = var.metadata.name

  service_account_name = try(var.spec.service_account.existing_name, "") != "" ? var.spec.service_account.existing_name : var.metadata.name

  port            = try(var.spec.server.port, null) != null ? var.spec.server.port : 1031
  monitoring_port = try(var.spec.server.monitoring_port, null) != null ? var.spec.server.monitoring_port : 1032

  # Without a prefix the relay would read EVERY environment variable as
  # configuration - including the service-link variables Kubernetes injects
  # for each Service in the namespace.
  env_prefix = try(var.spec.runtime.env_variable_prefix, "") != "" && try(var.spec.runtime.env_variable_prefix, null) != null ? var.spec.runtime.env_variable_prefix : "GOFFRELAY_"

  flag_reader_name     = "${var.metadata.name}-flag-reader"
  service_monitor_name = "${var.metadata.name}-metrics"

  labels = merge(
    {
      "planton.ai/resource" = "true"
      "planton.ai/name"     = var.metadata.name
      "planton.ai/kind"     = "KubernetesGoFeatureFlag"
    },
    try(var.metadata.id, "") != "" ? { "planton.ai/id" = var.metadata.id } : {},
    try(var.metadata.org, "") != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    try(var.metadata.env, "") != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # ---- flag sources ----------------------------------------------------------
  # flag_source mode renders one source into the top level; flag_sets mode
  # renders one per flag set. try() lets one list carry either shape.
  mode_flag_sets = try(var.spec.flag_sets, null) != null
  source_count   = local.mode_flag_sets ? length(var.spec.flag_sets.items) : 1
  source_inputs  = [for i in range(local.source_count) : try(var.spec.flag_sets.items[i].source, var.spec.flag_source)]
  source_env     = [for i in range(local.source_count) : local.mode_flag_sets ? "FLAGSETS_${i}_" : ""]

  rendered_sources = [for i, s in local.source_inputs : { for k, v in {
    retrievers = [for j, r in s.retrievers : { for rk, rv in {
      kind = (
        try(r.config_map, null) != null ? "configmap" :
        try(r.http, null) != null ? "http" :
        try(r.github, null) != null ? "github" :
        try(r.gitlab, null) != null ? "gitlab" :
        try(r.bitbucket, null) != null ? "bitbucket" :
        try(r.s3, null) != null ? "s3" :
        try(r.google_storage, null) != null ? "googleStorage" :
        try(r.azure_blob_storage, null) != null ? "azureBlobStorage" :
        try(r.mongodb, null) != null ? "mongodb" :
        try(r.redis, null) != null ? "redis" :
        try(r.postgresql, null) != null ? "postgresql" : null
      )
      namespace      = try(r.config_map, null) != null ? (try(r.config_map.namespace, "") != "" && try(r.config_map.namespace, null) != null ? r.config_map.namespace : local.namespace) : null
      configmap      = try(r.config_map.config_map_name, null)
      key            = try(r.config_map.key, null)
      url            = try(r.http.url, null)
      method         = try(r.http.method, "") != "" ? try(r.http.method, null) : null
      body           = try(r.http.body, "") != "" ? try(r.http.body, null) : null
      timeout        = try(tonumber(try(r.http.timeout_ms, r.github.timeout_ms, r.gitlab.timeout_ms, r.bitbucket.timeout_ms, null)), null)
      headers        = length(try(r.http.headers, {})) > 0 ? { for hk, hv in r.http.headers : hk => [hv] } : null
      repositorySlug = try(r.github.repository_slug, r.gitlab.repository_slug, r.bitbucket.repository_slug, null)
      path           = try(r.github.path, r.gitlab.path, r.bitbucket.path, null)
      branch         = try(r.github.branch, r.gitlab.branch, r.bitbucket.branch, "") != "" ? try(r.github.branch, r.gitlab.branch, r.bitbucket.branch, null) : null
      baseUrl        = try(r.github.base_url, r.gitlab.base_url, r.bitbucket.base_url, "") != "" ? try(r.github.base_url, r.gitlab.base_url, r.bitbucket.base_url, null) : null
      bucket         = try(r.s3.bucket, r.google_storage.bucket, null)
      item           = try(r.s3.item, null)
      object         = try(r.google_storage.object, r.azure_blob_storage.object, null)
      accountName    = try(r.azure_blob_storage.account_name, null)
      container      = try(r.azure_blob_storage.container, null)
      database       = try(r.mongodb.database, null)
      collection     = try(r.mongodb.collection, null)
      redisoptions = try(r.redis, null) == null ? null : { for ok, ov in {
        addr                  = r.redis.options.addr
        network               = try(r.redis.options.network, "") != "" ? try(r.redis.options.network, null) : null
        username              = try(r.redis.options.username, "") != "" ? try(r.redis.options.username, null) : null
        db                    = try(tonumber(r.redis.options.db), 0) != 0 ? tonumber(r.redis.options.db) : null
        tlsEnabled            = try(r.redis.options.tls_enabled, false) == true ? true : null
        protocol              = try(tonumber(r.redis.options.protocol), null)
        clientName            = try(r.redis.options.client_name, "") != "" ? try(r.redis.options.client_name, null) : null
        identitySuffix        = try(r.redis.options.identity_suffix, "") != "" ? try(r.redis.options.identity_suffix, null) : null
        disableIdentity       = try(r.redis.options.disable_identity, false) == true ? true : null
        maxRetries            = try(tonumber(r.redis.options.max_retries), null)
        minRetryBackoff       = try(tonumber(r.redis.options.min_retry_backoff_ms), null)
        maxRetryBackoff       = try(tonumber(r.redis.options.max_retry_backoff_ms), null)
        dialTimeout           = try(tonumber(r.redis.options.dial_timeout_ms), null)
        readTimeout           = try(tonumber(r.redis.options.read_timeout_ms), null)
        writeTimeout          = try(tonumber(r.redis.options.write_timeout_ms), null)
        contextTimeoutEnabled = try(r.redis.options.context_timeout_enabled, false) == true ? true : null
        poolFIFO              = try(r.redis.options.pool_fifo, false) == true ? true : null
        poolSize              = try(tonumber(r.redis.options.pool_size), null)
        poolTimeout           = try(tonumber(r.redis.options.pool_timeout_ms), null)
        minIdleConns          = try(tonumber(r.redis.options.min_idle_conns), 0) != 0 ? tonumber(r.redis.options.min_idle_conns) : null
        maxIdleConns          = try(tonumber(r.redis.options.max_idle_conns), 0) != 0 ? tonumber(r.redis.options.max_idle_conns) : null
        connMaxIdleTime       = try(tonumber(r.redis.options.conn_max_idle_time_ms), null)
        connMaxLifetime       = try(tonumber(r.redis.options.conn_max_lifetime_ms), null)
      } : ok => ov if ov != null }
      redisPrefix = try(r.redis.prefix, "") != "" ? try(r.redis.prefix, null) : null
      table       = try(r.postgresql.table, null)
      columns     = length(try(r.postgresql.columns, {})) > 0 ? r.postgresql.columns : null
    } : rk => rv if rv != null }]

    notifiers = length(try(s.notifiers, [])) == 0 ? null : [for n in s.notifiers : { for nk, nv in {
      kind = (
        try(n.slack, null) != null ? "slack" :
        try(n.microsoft_teams, null) != null ? "microsoftteams" :
        try(n.discord, null) != null ? "discord" :
        try(n.webhook, null) != null ? "webhook" : null
      )
      endpointUrl = try(n.webhook.endpoint_url, null)
      meta        = length(try(n.webhook.meta, {})) > 0 ? n.webhook.meta : null
      headers     = length(try(n.webhook.headers, {})) > 0 ? { for hk, hv in n.webhook.headers : hk => [hv] } : null
    } : nk => nv if nv != null }]

    exporters = length(try(s.exporters, [])) == 0 ? null : [for e in s.exporters : { for ek, ev in {
      kind = (
        try(e.webhook, null) != null ? "webhook" :
        try(e.log, null) != null ? "log" :
        try(e.s3, null) != null ? "s3" :
        try(e.google_storage, null) != null ? "googleStorage" :
        try(e.azure_blob_storage, null) != null ? "azureBlobStorage" :
        try(e.sqs, null) != null ? "sqs" :
        try(e.kinesis, null) != null ? "kinesis" :
        try(e.pubsub, null) != null ? "pubsub" :
        try(e.bigquery, null) != null ? "bigquery" :
        try(e.kafka, null) != null ? "kafka" :
        try(e.opentelemetry, null) != null ? "opentelemetry" : null
      )
      endpointUrl             = try(e.webhook.endpoint_url, null)
      meta                    = length(try(e.webhook.meta, {})) > 0 ? e.webhook.meta : null
      headers                 = length(try(e.webhook.headers, {})) > 0 ? { for hk, hv in e.webhook.headers : hk => [hv] } : null
      logFormat               = try(e.log.log_format, "") != "" ? try(e.log.log_format, null) : null
      bucket                  = try(e.s3.bucket, e.google_storage.bucket, null)
      accountName             = try(e.azure_blob_storage.account_name, null)
      container               = try(e.azure_blob_storage.container, null)
      path                    = try(e.s3.path, e.google_storage.path, e.azure_blob_storage.path, "") != "" ? try(e.s3.path, e.google_storage.path, e.azure_blob_storage.path, null) : null
      format                  = try(e.s3.file.format, e.google_storage.file.format, e.azure_blob_storage.file.format, e.kinesis.format, "") != "" ? try(e.s3.file.format, e.google_storage.file.format, e.azure_blob_storage.file.format, e.kinesis.format, null) : null
      filename                = try(e.s3.file.filename, e.google_storage.file.filename, e.azure_blob_storage.file.filename, "") != "" ? try(e.s3.file.filename, e.google_storage.file.filename, e.azure_blob_storage.file.filename, null) : null
      csvTemplate             = try(e.s3.file.csv_template, e.google_storage.file.csv_template, e.azure_blob_storage.file.csv_template, "") != "" ? try(e.s3.file.csv_template, e.google_storage.file.csv_template, e.azure_blob_storage.file.csv_template, null) : null
      parquetCompressionCodec = try(e.s3.file.parquet_compression_codec, e.google_storage.file.parquet_compression_codec, e.azure_blob_storage.file.parquet_compression_codec, "") != "" ? try(e.s3.file.parquet_compression_codec, e.google_storage.file.parquet_compression_codec, e.azure_blob_storage.file.parquet_compression_codec, null) : null
      queueUrl                = try(e.sqs.queue_url, null)
      streamArn               = try(e.kinesis.stream_arn, "") != "" ? try(e.kinesis.stream_arn, null) : null
      streamName              = try(e.kinesis.stream_name, "") != "" ? try(e.kinesis.stream_name, null) : null
      projectID               = try(e.pubsub.project_id, e.bigquery.project_id, null)
      topic                   = try(e.pubsub.topic, null)
      datasetID               = try(e.bigquery.dataset_id, null)
      tableName               = try(e.bigquery.table_name, "") != "" ? try(e.bigquery.table_name, null) : null
      autoMigrate             = try(e.bigquery.auto_migrate, false) == true ? true : null
      kafka = try(e.kafka, null) == null ? null : { for kk, kv in {
        topic     = e.kafka.topic
        addresses = e.kafka.addresses
        # Keys lowercased so the SASL password variable merges into the
        # same keys. PARITY-EXCEPTION: lowercased to a depth of six, keys
        # inside lists left as written (HCL has no recursion; the Pulumi
        # twin lowerKeys is unbounded) - a Sarama configuration nests four
        # levels deep and holds no lists of objects.
        config = length(try(keys(e.kafka.config), [])) > 0 ? try(keys(e.kafka.config) == null ? null : { for k6, v6 in e.kafka.config : lower(k6) => try(keys(v6) == null ? null : { for k5, v5 in v6 : lower(k5) => try(keys(v5) == null ? null : { for k4, v4 in v5 : lower(k4) => try(keys(v4) == null ? null : { for k3, v3 in v4 : lower(k3) => try(keys(v3) == null ? null : { for k2, v2 in v3 : lower(k2) => try(keys(v2) == null ? null : { for k1, v1 in v2 : lower(k1) => v1 }, v2) }, v3) }, v4) }, v5) }, v6) }, e.kafka.config) : null
      } : kk => kv if kv != null }
      tracerName       = try(e.opentelemetry.tracer_name, "") != "" ? try(e.opentelemetry.tracer_name, null) : null
      flushInterval    = try(tonumber(e.flush_interval_ms), null)
      maxEventInMemory = try(tonumber(e.max_event_in_memory), null)
      eventType        = try(e.event_type, "") != "" ? try(e.event_type, null) : null
    } : ek => ev if ev != null }]

    fileFormat                  = try(s.file_format, "") != "" ? try(s.file_format, null) : null
    pollingInterval             = try(tonumber(s.polling_interval_ms), null)
    startWithRetrieverError     = try(s.start_with_retriever_error, null)
    enablePollingJitter         = try(s.enable_polling_jitter, false) == true ? true : null
    disableNotifierOnInit       = try(s.disable_notifier_on_init, false) == true ? true : null
    evaluationContextEnrichment = length(try(keys(s.evaluation_context_enrichment), [])) > 0 ? s.evaluation_context_enrichment : null
  } : k => v if v != null }]

  # ---- secret environment ------------------------------------------------------
  # The relay reads list entries from indexed variables and key lists from
  # comma-separated ones (Pulumi twin: renderer.secret / keyList / headers).
  secret_env_entries = concat(
    [{ name = "AUTHORIZEDKEYS_ADMIN", value = join(",", try(var.spec.authorized_keys.admin, [])) }],
    [{ name = "AUTHORIZEDKEYS_EVALUATION", value = join(",", try(var.spec.authorized_keys.evaluation, [])) }],
    [for i in range(local.mode_flag_sets ? local.source_count : 0) : { name = "FLAGSETS_${i}_APIKEYS", value = join(",", try(var.spec.flag_sets.items[i].api_keys, [])) }],
    flatten([for i, s in local.source_inputs : [
      [for j, r in s.retrievers : concat(
        [
          { name = "${local.source_env[i]}RETRIEVERS_${j}_TOKEN", value = try(r.github.token, r.gitlab.token, r.bitbucket.token, "") },
          { name = "${local.source_env[i]}RETRIEVERS_${j}_ACCOUNTKEY", value = try(r.azure_blob_storage.account_key, "") },
          { name = "${local.source_env[i]}RETRIEVERS_${j}_URI", value = try(r.mongodb.uri, r.postgresql.uri, "") },
          { name = "${local.source_env[i]}RETRIEVERS_${j}_REDISOPTIONS_PASSWORD", value = try(r.redis.options.password, "") },
        ],
        [for hk, hv in try(r.http.sensitive_headers, {}) : { name = "${local.source_env[i]}RETRIEVERS_${j}_HEADERS_${upper(hk)}", value = hv }]
      )],
      [for j, n in try(s.notifiers, []) : concat(
        [
          { name = "${local.source_env[i]}NOTIFIERS_${j}_WEBHOOKURL", value = try(n.slack.webhook_url, n.microsoft_teams.webhook_url, n.discord.webhook_url, "") },
          { name = "${local.source_env[i]}NOTIFIERS_${j}_SECRET", value = try(n.webhook.secret, "") },
        ],
        [for hk, hv in try(n.webhook.sensitive_headers, {}) : { name = "${local.source_env[i]}NOTIFIERS_${j}_HEADERS_${upper(hk)}", value = hv }]
      )],
      [for j, e in try(s.exporters, []) : concat(
        [
          { name = "${local.source_env[i]}EXPORTERS_${j}_SECRET", value = try(e.webhook.secret, "") },
          { name = "${local.source_env[i]}EXPORTERS_${j}_ACCOUNTKEY", value = try(e.azure_blob_storage.account_key, "") },
          { name = "${local.source_env[i]}EXPORTERS_${j}_GOOGLECREDENTIALS", value = try(e.bigquery.google_credentials, "") },
          { name = "${local.source_env[i]}EXPORTERS_${j}_KAFKA_CONFIG_NET_SASL_PASSWORD", value = try(e.kafka.sasl_password, "") },
        ],
        [for hk, hv in try(e.webhook.sensitive_headers, {}) : { name = "${local.source_env[i]}EXPORTERS_${j}_HEADERS_${upper(hk)}", value = hv }]
      )],
    ]])
  )
  secret_env      = { for e in local.secret_env_entries : "${local.env_prefix}${e.name}" => e.value if e.value != null && e.value != "" }
  env_secret_name = length(local.secret_env) > 0 ? "${var.metadata.name}-env" : ""

  # ---- deploy-time checks (Pulumi twin: errors from buildRelayConfig) -----------
  flag_set_keys = flatten([for i in range(local.mode_flag_sets ? local.source_count : 0) : try(var.spec.flag_sets.items[i].api_keys, [])])
  all_keys      = concat(try(var.spec.authorized_keys.admin, []), try(var.spec.authorized_keys.evaluation, []), local.flag_set_keys)
  header_pairs = flatten([for s in local.source_inputs : concat(
    [for r in s.retrievers : { plain = keys(try(r.http.headers, {})), sensitive = keys(try(r.http.sensitive_headers, {})) }],
    [for n in try(s.notifiers, []) : { plain = keys(try(n.webhook.headers, {})), sensitive = keys(try(n.webhook.sensitive_headers, {})) }],
    [for e in try(s.exporters, []) : { plain = keys(try(e.webhook.headers, {})), sensitive = keys(try(e.webhook.sensitive_headers, {})) }]
  )])
  sensitive_header_names = flatten([for p in local.header_pairs : p.sensitive])
  header_conflicts       = flatten([for p in local.header_pairs : [for h in p.sensitive : h if contains([for x in p.plain : lower(x)], lower(h))]])

  # ---- ConfigMap grants (Pulumi twin: RelayConfig.ConfigMapGrants) ---------------
  config_map_refs = flatten([for s in local.source_inputs : [for r in s.retrievers : {
    namespace = try(r.config_map.namespace, "") != "" && try(r.config_map.namespace, null) != null ? r.config_map.namespace : local.namespace
    name      = r.config_map.config_map_name
  } if try(r.config_map, null) != null]])
  config_map_grants = { for ns in distinct([for c in local.config_map_refs : c.namespace]) : ns => sort(distinct([for c in local.config_map_refs : c.name if c.namespace == ns])) }

  # ---- the relay configuration document (Pulumi twin: buildRelayConfig) ----------
  relay_top = { for k, v in {
    server = {
      mode           = "http"
      port           = local.port
      monitoringPort = local.monitoring_port
    }
    envVariablePrefix = local.env_prefix
    logLevel          = try(var.spec.log.level, "") != "" && try(var.spec.log.level, null) != null ? var.spec.log.level : "info"
    logFormat         = try(var.spec.log.format, "") != "" && try(var.spec.log.format, null) != null ? var.spec.log.format : "json"
    flagsets = local.mode_flag_sets ? [for i, fs in var.spec.flag_sets.items : merge(
      local.rendered_sources[i],
      try(fs.name, "") != "" ? { name = fs.name } : {}
    )] : null
    ofrepEventStream = (try(var.spec.ofrep_event_stream.base_url, "") != "" || try(var.spec.ofrep_event_stream.inactivity_delay_sec, 0) != 0) ? { for ok, ov in {
      baseUrl            = try(var.spec.ofrep_event_stream.base_url, "") != "" ? var.spec.ofrep_event_stream.base_url : null
      inactivityDelaySec = try(var.spec.ofrep_event_stream.inactivity_delay_sec, 0) != 0 ? var.spec.ofrep_event_stream.inactivity_delay_sec : null
    } : ok => ov if ov != null } : null
    otel   = length(local.otel_block) > 0 ? local.otel_block : null
    jaeger = length(local.jaeger_sampler_block) > 0 ? { sampler = local.jaeger_sampler_block } : null
    swagger = (try(var.spec.swagger.enabled, false) == true || try(var.spec.swagger.host, "") != "") ? { for sk, sv in {
      enabled = try(var.spec.swagger.enabled, false) == true
      host    = try(var.spec.swagger.host, "") != "" ? var.spec.swagger.host : null
    } : sk => sv if sv != null } : null
    hideBanner                 = try(var.spec.runtime.hide_banner, false) == true ? true : null
    enablePprof                = try(var.spec.runtime.enable_pprof, false) == true ? true : null
    disableVersionHeader       = try(var.spec.runtime.disable_version_header, false) == true ? true : null
    enableBulkMetricFlagNames  = try(var.spec.runtime.enable_bulk_metric_flag_names, false) == true ? true : null
    disableFlagDetailsInStream = try(var.spec.runtime.disable_flag_details_in_stream, false) == true ? true : null
    exporterCleanQueueInterval = try(var.spec.runtime.exporter_clean_queue_interval, "") != "" ? var.spec.runtime.exporter_clean_queue_interval : null
  } : k => v if v != null }

  otel_block = { for k, v in {
    exporter = try(var.spec.telemetry.otlp_endpoint, "") != "" ? { otlp = { endpoint = var.spec.telemetry.otlp_endpoint } } : null
    sdk      = try(var.spec.telemetry.sdk_disabled, false) == true ? { disabled = true } : null
    service  = try(var.spec.telemetry.service_name, "") != "" ? { name = var.spec.telemetry.service_name } : null
    traces   = try(var.spec.telemetry.traces_sampler, "") != "" ? { sampler = var.spec.telemetry.traces_sampler } : null
    resource = length(try(var.spec.telemetry.resource_attributes, {})) > 0 ? { attributes = var.spec.telemetry.resource_attributes } : null
  } : k => v if v != null }

  jaeger_sampler_block = { for k, v in {
    manager = try(var.spec.telemetry.jaeger_sampler.manager_host_port, "") != "" ? { host = { port = var.spec.telemetry.jaeger_sampler.manager_host_port } } : null
    refresh = try(var.spec.telemetry.jaeger_sampler.refresh_interval, "") != "" ? { interval = var.spec.telemetry.jaeger_sampler.refresh_interval } : null
    max     = try(var.spec.telemetry.jaeger_sampler.max_operations, 0) != 0 ? { operations = var.spec.telemetry.jaeger_sampler.max_operations } : null
  } : k => v if v != null }

  # flag_source mode merges the one rendered source into the top level.
  flag_source_part = [for i in(local.mode_flag_sets ? [] : [0]) : local.rendered_sources[i]]
  relay_doc        = merge(concat([local.relay_top], local.flag_source_part)...)

  # The chart discovers the monitoring port by scanning the config string for
  # "server.monitoringPort: <port>", and runs the string through Helm's tpl.
  relay_config = "# server.monitoringPort: ${local.monitoring_port}\n${replace(jsonencode(local.relay_doc), "{{", "{{\"{{\"}}")}"

  # ---- chart values (Pulumi twin: buildHelmValues) ----------------------------
  # Environment: extra variables as given, secret values by secretKeyRef into
  # the module-owned Secret, and the OpenTelemetry protocol and sampler
  # variables, which the relay's exporter and sampler read from the
  # environment only.
  extra_env_values = merge(
    { for name, value in try(var.spec.extra_env, {}) : name => { value = value } },
    { for name, ref in try(var.spec.extra_env_from_secret, {}) : name => { valueFrom = { secretKeyRef = { name = ref.name, key = ref.key } } } },
    try(var.spec.telemetry.otlp_protocol, "") != "" ? { OTEL_EXPORTER_OTLP_PROTOCOL = { value = var.spec.telemetry.otlp_protocol } } : {},
    # Every sampler but jaeger_remote is built by the OpenTelemetry SDK from
    # its own environment.
    try(var.spec.telemetry.traces_sampler, "") != "" && try(var.spec.telemetry.traces_sampler, "") != "jaeger_remote" ? { OTEL_TRACES_SAMPLER = { value = var.spec.telemetry.traces_sampler } } : {},
    try(var.spec.telemetry.traces_sampler_arg, "") != "" ? { OTEL_TRACES_SAMPLER_ARG = { value = var.spec.telemetry.traces_sampler_arg } } : {}
  )
  env_values = merge(
    local.extra_env_values,
    { for name, _ in local.secret_env : name => { valueFrom = { secretKeyRef = { name = local.env_secret_name, key = name } } } }
  )

  # The chart's PodDisruptionBudget selects `name: <fullname>`, a label the
  # chart never puts on pods - podLabels carries it so the budget protects
  # the relay.
  pod_labels = merge(try(var.spec.pod_labels, {}), local.labels, { name = var.metadata.name })

  image_block = { for k, v in {
    repository = try(var.spec.image.repository, "") != "" && try(var.spec.image.repository, null) != null ? var.spec.image.repository : null
    tag        = try(var.spec.image.tag, "") != "" ? var.spec.image.tag : null
    pullPolicy = try(var.spec.image.pull_policy, "") != "" && try(var.spec.image.pull_policy, null) != null ? var.spec.image.pull_policy : null
    fips       = try(var.spec.image.fips, false) == true ? true : null
  } : k => v if v != null }

  # The chart renders a memory target unless it is explicitly null, so the
  # HPA block always states it.
  autoscaling_block = try(var.spec.hpa.enabled, false) == true ? merge({ for k, v in {
    enabled                        = true
    minReplicas                    = try(var.spec.hpa.min_replicas, null)
    maxReplicas                    = try(var.spec.hpa.max_replicas, null)
    targetCPUUtilizationPercentage = try(var.spec.hpa.target_cpu_utilization_percent, null)
  } : k => v if v != null }, { targetMemoryUtilizationPercentage = try(var.spec.hpa.target_memory_utilization_percent, null) }) : null

  pdb_block = try(var.spec.pdb.enabled, false) == true ? { for k, v in {
    enable         = true
    maxUnavailable = try(var.spec.pdb.max_unavailable, "") != "" ? try(tonumber(regex("^[0-9]+$", var.spec.pdb.max_unavailable)), var.spec.pdb.max_unavailable) : null
    minAvailable   = try(var.spec.pdb.max_unavailable, "") == "" && try(var.spec.pdb.min_available, "") != "" ? try(tonumber(regex("^[0-9]+$", var.spec.pdb.min_available)), var.spec.pdb.min_available) : null
  } : k => v if v != null } : null

  resources_block = { for k, v in {
    requests = try(var.spec.resources.requests, null) == null ? null : { for rk, rv in {
      cpu    = var.spec.resources.requests.cpu
      memory = var.spec.resources.requests.memory
    } : rk => rv if rv != null && rv != "" }
    limits = try(var.spec.resources.limits, null) == null ? null : { for rk, rv in {
      cpu    = var.spec.resources.limits.cpu
      memory = var.spec.resources.limits.memory
    } : rk => rv if rv != null && rv != "" }
  } : k => v if v != null && length(v) > 0 }

  tolerations = [for t in try(var.spec.scheduling.tolerations, []) : { for k, v in {
    key               = t.key != "" ? t.key : null
    operator          = t.operator != "" ? t.operator : null
    value             = t.value != "" ? t.value : null
    effect            = t.effect != "" ? t.effect : null
    tolerationSeconds = t.toleration_seconds
  } : k => v if v != null }]

  node_affinity = try(var.spec.scheduling.node_affinity, null) == null ? null : { for k, v in {
    requiredDuringSchedulingIgnoredDuringExecution = length(var.spec.scheduling.node_affinity.required) > 0 ? {
      nodeSelectorTerms = [for t in var.spec.scheduling.node_affinity.required : {
        matchExpressions = [for e in t.match_expressions : { for ek, ev in { key = e.key, operator = e.operator, values = length(e.values) > 0 ? e.values : null } : ek => ev if ev != null }]
      }]
    } : null
    preferredDuringSchedulingIgnoredDuringExecution = length(var.spec.scheduling.node_affinity.preferred) > 0 ? [for p in var.spec.scheduling.node_affinity.preferred : {
      weight = p.weight
      preference = {
        matchExpressions = [for e in p.term.match_expressions : { for ek, ev in { key = e.key, operator = e.operator, values = length(e.values) > 0 ? e.values : null } : ek => ev if ev != null }]
      }
    }] : null
  } : k => v if v != null }

  pod_affinity_terms = { for kind in ["pod_affinity", "pod_anti_affinity"] : kind => try(var.spec.scheduling[kind], null) == null ? null : { for k, v in {
    requiredDuringSchedulingIgnoredDuringExecution = length(var.spec.scheduling[kind].required) > 0 ? [for t in var.spec.scheduling[kind].required : { for tk, tv in {
      labelSelector = { matchLabels = t.match_labels }
      topologyKey   = t.topology_key
      namespaces    = length(t.namespaces) > 0 ? t.namespaces : null
    } : tk => tv if tv != null }] : null
    preferredDuringSchedulingIgnoredDuringExecution = length(var.spec.scheduling[kind].preferred) > 0 ? [for w in var.spec.scheduling[kind].preferred : {
      weight = w.weight
      podAffinityTerm = { for tk, tv in {
        labelSelector = { matchLabels = w.term.match_labels }
        topologyKey   = w.term.topology_key
        namespaces    = length(w.term.namespaces) > 0 ? w.term.namespaces : null
      } : tk => tv if tv != null }
    }] : null
  } : k => v if v != null } }

  affinity_block = { for k, v in {
    nodeAffinity    = local.node_affinity != null && length(coalesce(local.node_affinity, {})) > 0 ? local.node_affinity : null
    podAffinity     = local.pod_affinity_terms["pod_affinity"] != null && length(coalesce(local.pod_affinity_terms["pod_affinity"], {})) > 0 ? local.pod_affinity_terms["pod_affinity"] : null
    podAntiAffinity = local.pod_affinity_terms["pod_anti_affinity"] != null && length(coalesce(local.pod_affinity_terms["pod_anti_affinity"], {})) > 0 ? local.pod_affinity_terms["pod_anti_affinity"] : null
  } : k => v if v != null }

  pod_security_context = try(var.spec.pod_security_context, null) == null ? null : { for k, v in {
    runAsUser           = try(var.spec.pod_security_context.run_as_user, null)
    runAsGroup          = try(var.spec.pod_security_context.run_as_group, null)
    runAsNonRoot        = try(var.spec.pod_security_context.run_as_non_root, null)
    fsGroup             = try(var.spec.pod_security_context.fs_group, null)
    fsGroupChangePolicy = try(var.spec.pod_security_context.fs_group_change_policy, "") != "" ? var.spec.pod_security_context.fs_group_change_policy : null
    supplementalGroups  = length(try(var.spec.pod_security_context.supplemental_groups, [])) > 0 ? var.spec.pod_security_context.supplemental_groups : null
    sysctls             = length(try(var.spec.pod_security_context.sysctls, [])) > 0 ? [for s in var.spec.pod_security_context.sysctls : { name = s.name, value = s.value }] : null
    seccompProfile = try(var.spec.pod_security_context.seccomp_profile, null) == null ? null : { for sk, sv in {
      type             = try(var.spec.pod_security_context.seccomp_profile.type, "") != "" ? var.spec.pod_security_context.seccomp_profile.type : null
      localhostProfile = try(var.spec.pod_security_context.seccomp_profile.localhost_profile, "") != "" ? var.spec.pod_security_context.seccomp_profile.localhost_profile : null
    } : sk => sv if sv != null }
  } : k => v if v != null && v != {} }

  container_security_context = try(var.spec.container_security_context, null) == null ? null : { for k, v in {
    privileged               = try(var.spec.container_security_context.privileged, false) == true ? true : null
    runAsUser                = try(var.spec.container_security_context.run_as_user, null)
    runAsGroup               = try(var.spec.container_security_context.run_as_group, null)
    runAsNonRoot             = try(var.spec.container_security_context.run_as_non_root, null)
    readOnlyRootFilesystem   = try(var.spec.container_security_context.read_only_root_filesystem, null)
    allowPrivilegeEscalation = try(var.spec.container_security_context.allow_privilege_escalation, null)
    capabilities = (length(try(var.spec.container_security_context.capabilities.add, [])) + length(try(var.spec.container_security_context.capabilities.drop, []))) > 0 ? { for ck, cv in {
      add  = length(try(var.spec.container_security_context.capabilities.add, [])) > 0 ? var.spec.container_security_context.capabilities.add : null
      drop = length(try(var.spec.container_security_context.capabilities.drop, [])) > 0 ? var.spec.container_security_context.capabilities.drop : null
    } : ck => cv if cv != null } : null
    seccompProfile = try(var.spec.container_security_context.seccomp_profile, null) == null ? null : { for sk, sv in {
      type             = try(var.spec.container_security_context.seccomp_profile.type, "") != "" ? var.spec.container_security_context.seccomp_profile.type : null
      localhostProfile = try(var.spec.container_security_context.seccomp_profile.localhost_profile, "") != "" ? var.spec.container_security_context.seccomp_profile.localhost_profile : null
    } : sk => sv if sv != null }
  } : k => v if v != null && v != {} }

  service_account_block = { for k, v in {
    create      = try(var.spec.service_account.existing_name, "") != "" ? false : null
    name        = try(var.spec.service_account.existing_name, "") != "" ? var.spec.service_account.existing_name : null
    annotations = try(var.spec.service_account.existing_name, "") == "" && length(try(var.spec.service_account.annotations, {})) > 0 ? var.spec.service_account.annotations : null
  } : k => v if v != null }

  # The env Secret's checksum rolls the pods when a secret value changes: env
  # is read once at process start (Pulumi twin: secretEnvChecksum).
  pod_annotations = merge(
    try(var.spec.pod_annotations, {}),
    length(local.secret_env) > 0 ? { "checksum/env-secret" = sha256(jsonencode(local.secret_env)) } : {},
  )

  typed_helm_values = { for k, v in {
    fullnameOverride = var.metadata.name
    relayproxy       = { config = local.relay_config }
    service = { for sk, sv in {
      port = local.port
      type = try(var.spec.server.service_type, "") != "" && try(var.spec.server.service_type, null) != null ? var.spec.server.service_type : null
    } : sk => sv if sv != null }
    image              = length(local.image_block) > 0 ? local.image_block : null
    imagePullSecrets   = length(try(var.spec.image.pull_secret_names, [])) > 0 ? [for n in var.spec.image.pull_secret_names : { name = n }] : null
    autoscaling        = local.autoscaling_block
    replicaCount       = try(var.spec.hpa.enabled, false) == true ? null : try(var.spec.replicas, null)
    pdb                = local.pdb_block
    podLabels          = local.pod_labels
    podAnnotations     = length(local.pod_annotations) > 0 ? local.pod_annotations : null
    commonLabels       = length(try(var.spec.common_labels, {})) > 0 ? var.spec.common_labels : null
    podSecurityContext = local.pod_security_context != null && try(length(local.pod_security_context), 0) > 0 ? local.pod_security_context : null
    securityContext    = local.container_security_context != null && try(length(local.container_security_context), 0) > 0 ? local.container_security_context : null
    resources          = length(local.resources_block) > 0 ? local.resources_block : null
    env                = length(local.env_values) > 0 ? local.env_values : null
    nodeSelector       = length(try(var.spec.scheduling.node_selector, {})) > 0 ? var.spec.scheduling.node_selector : null
    tolerations        = length(local.tolerations) > 0 ? local.tolerations : null
    affinity           = length(local.affinity_block) > 0 ? local.affinity_block : null
    serviceAccount     = length(local.service_account_block) > 0 ? local.service_account_block : null
  } : k => v if v != null }
}
