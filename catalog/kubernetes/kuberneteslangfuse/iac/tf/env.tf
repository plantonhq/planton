# The environment the chart's additionalEnv channels carry (twin: env.go
# buildEnvironment): every typed setting the chart has no value for.
# Variables the chart renders itself (the stores, the salt, encryption key
# and NextAuth secret, telemetry, sign-up, the license key, logging) are
# never repeated here: the chart refuses a variable set both ways.
#
# Each channel is assembled from three maps - plain values (an empty value
# sets nothing, so Langfuse's own default applies), keys of the instance
# Secret, and the spec's extra variables (last: the spec's own entries
# win) - and rendered sorted by name.

locals {
  enum_value = {
    # A proto enum left unset arrives as "" or as its *_unspecified name.
    clickhouse_delete_mode = endswith(try(var.spec.clickhouse.delete_mode, ""), "unspecified") ? "" : try(var.spec.clickhouse.delete_mode, "")
    checksum_algorithm     = endswith(try(var.spec.blob_storage.delete_objects_checksum_algorithm, ""), "unspecified") ? "" : upper(try(var.spec.blob_storage.delete_objects_checksum_algorithm, ""))
    org_role               = endswith(try(var.spec.auth.new_user_access.organization_role, ""), "unspecified") ? "" : try(var.spec.auth.new_user_access.organization_role, "")
    project_role           = endswith(try(var.spec.auth.new_user_access.project_role, ""), "unspecified") ? "" : try(var.spec.auth.new_user_access.project_role, "")
    ai_provider            = endswith(try(var.spec.ai_features.provider, ""), "unspecified") ? "" : try(var.spec.ai_features.provider, "")
    media_transport        = endswith(try(var.spec.evaluation.media_transport, ""), "unspecified") ? "" : try(var.spec.evaluation.media_transport, "")
    model_adapter          = endswith(try(var.spec.enterprise.ui.default_model_adapter, ""), "unspecified") ? "" : try(var.spec.enterprise.ui.default_model_adapter, "")
  }

  # Server-side encryption per upload type.
  sse = { for uc, s in {
    EVENT_UPLOAD = try(var.spec.blob_storage.event_upload.sse, null)
    MEDIA_UPLOAD = try(var.spec.blob_storage.media_upload.sse, null)
    BATCH_EXPORT = try(var.spec.blob_storage.batch_export.sse, null)
  } : uc => s if s != null }

  code_evaluators = try(var.spec.evaluation.code_evaluators, null)
  assistant       = try(var.spec.ai_features.assistant, null)
  sandbox         = try(local.assistant.sandbox, null)
  masking         = try(var.spec.enterprise.ingestion_masking, null)
  instance_links  = try(var.spec.enterprise.ui.instance_links, [])

  # ---- plain values ----------------------------------------------------------
  common_plain = merge(
    {
      # clickhouse
      CLICKHOUSE_CLUSTER_NAME                 = try(var.spec.clickhouse.cluster.name, "")
      CLICKHOUSE_READ_ONLY_URL                = try(var.spec.clickhouse.read_only_url, "")
      CLICKHOUSE_LIGHTWEIGHT_DELETE_MODE      = local.enum_value.clickhouse_delete_mode
      CLICKHOUSE_USE_LIGHTWEIGHT_UPDATE       = try(var.spec.clickhouse.use_lightweight_update, false) ? "true" : ""
      LANGFUSE_CLICKHOUSE_DELETION_TIMEOUT_MS = try(var.spec.clickhouse.deletion_timeout_ms, null) == null ? "" : tostring(var.spec.clickhouse.deletion_timeout_ms)
      # redis
      REDIS_KEY_PREFIX                    = try(var.spec.redis.key_prefix, "")
      REDIS_SOCKET_TIMEOUT_MS             = try(var.spec.redis.socket_timeout_ms, null) == null ? "" : tostring(var.spec.redis.socket_timeout_ms)
      REDIS_CLUSTER_SLOTS_REFRESH_TIMEOUT = try(var.spec.redis.cluster.slots_refresh_timeout_ms, null) == null ? "" : tostring(var.spec.redis.cluster.slots_refresh_timeout_ms)
      REDIS_SENTINEL_TLS_ENABLED          = try(var.spec.redis.sentinel.tls_enabled, false) ? "true" : ""
      REDIS_TLS_SERVERNAME                = try(local.redis_tls.server_name, "")
      REDIS_TLS_REJECT_UNAUTHORIZED       = try(local.redis_tls.reject_unauthorized, null) == null ? "" : tostring(local.redis_tls.reject_unauthorized)
      REDIS_TLS_CHECK_SERVER_IDENTITY     = try(local.redis_tls.check_server_identity, null) == null ? "" : tostring(local.redis_tls.check_server_identity)
      REDIS_TLS_CIPHERS                   = try(local.redis_tls.ciphers, "")
      REDIS_TLS_SECURE_PROTOCOL           = try(local.redis_tls.secure_protocol, "")
      REDIS_TLS_HONOR_CIPHER_ORDER        = try(local.redis_tls.honor_cipher_order, null) == null ? "" : tostring(local.redis_tls.honor_cipher_order)
      # blob storage
      LANGFUSE_S3_MEDIA_UPLOAD_INTERNAL_ENDPOINT    = try(var.spec.blob_storage.media_upload.public_endpoint, "") != "" ? local.media_target.endpoint : ""
      LANGFUSE_S3_BATCH_EXPORT_EXTERNAL_ENDPOINT    = try(var.spec.blob_storage.batch_export.public_endpoint, "")
      LANGFUSE_S3_DELETE_OBJECTS_CHECKSUM_ALGORITHM = local.enum_value.checksum_algorithm
      # caching
      LANGFUSE_CACHE_PROMPT_ENABLED     = try(var.spec.caching.prompt_cache_enabled, null) == null ? "" : tostring(var.spec.caching.prompt_cache_enabled)
      LANGFUSE_CACHE_PROMPT_TTL_SECONDS = try(var.spec.caching.prompt_cache_ttl_seconds, null) == null ? "" : tostring(var.spec.caching.prompt_cache_ttl_seconds)
      # ingestion (shard counts reach both workloads: the web server enqueues
      # to the shards the worker consumes)
      LANGFUSE_INGESTION_QUEUE_SHARD_COUNT    = try(var.spec.ingestion.queue_shard_count, null) == null ? "" : tostring(var.spec.ingestion.queue_shard_count)
      LANGFUSE_TRACE_UPSERT_QUEUE_SHARD_COUNT = try(var.spec.ingestion.trace_upsert_queue_shard_count, null) == null ? "" : tostring(var.spec.ingestion.trace_upsert_queue_shard_count)
      # langfuse ai (rendered here rather than through the chart's aiFeatures
      # values: the chart's provider list predates Vertex)
      LANGFUSE_AI_PROVIDER            = local.enum_value.ai_provider
      LANGFUSE_AI_MODEL               = try(var.spec.ai_features.model, "")
      LANGFUSE_AI_SMALL_MODEL         = try(var.spec.ai_features.small_model, "")
      LANGFUSE_AI_BASE_URL            = try(var.spec.ai_features.base_url, "")
      LANGFUSE_AI_EXTRA_HEADERS       = length(try(var.spec.ai_features.extra_headers, {})) > 0 ? jsonencode(var.spec.ai_features.extra_headers) : ""
      LANGFUSE_AI_USE_RESPONSES_API   = try(var.spec.ai_features.use_responses_api, false) ? "true" : ""
      LANGFUSE_AI_AWS_BEDROCK_REGION  = try(var.spec.ai_features.bedrock_region, "")
      LANGFUSE_AI_VERTEX_LOCATION     = try(var.spec.ai_features.vertex_location, "")
      LANGFUSE_AI_FEATURES_PROJECT_ID = try(var.spec.ai_features.trace_project_id, "")
      LANGFUSE_IN_APP_AGENT_ENABLED   = try(local.assistant.enabled, false) ? "true" : ""
      # evaluation
      LANGFUSE_EVALUATOR_MEDIA_TRANSPORT        = local.enum_value.media_transport
      LANGFUSE_EVALUATOR_MEDIA_INLINE_MAX_BYTES = try(var.spec.evaluation.media_inline_max_bytes, null) == null ? "" : tostring(var.spec.evaluation.media_inline_max_bytes)
      # smtp
      EMAIL_FROM_ADDRESS = try(var.spec.smtp.from_address, "")
      # slack
      SLACK_CLIENT_ID = try(var.spec.slack.client_id, "")
      # observability
      LANGFUSE_LOG_PROPAGATED_HEADERS = join(",", try(var.spec.observability.log_propagated_headers, []))
      OTEL_EXPORTER_OTLP_ENDPOINT     = try(var.spec.observability.otlp_endpoint, "")
      # network
      HTTPS_PROXY         = try(var.spec.network.https_proxy, "")
      NO_PROXY            = join(",", try(var.spec.network.no_proxy, []))
      NODE_EXTRA_CA_CERTS = local.trusted_ca_path
    },
    merge([for uc, s in local.sse : {
      "LANGFUSE_S3_${uc}_SSE"            = try(s.algorithm, "") == "aws_kms" ? "aws:kms" : "AES256"
      "LANGFUSE_S3_${uc}_SSE_KMS_KEY_ID" = try(s.kms_key_id, "")
    }]...),
    local.code_evaluators == null ? {} : merge(
      try(local.code_evaluators.aws_lambda, null) != null ? {
        LANGFUSE_CODE_EVAL_DISPATCHER                      = "aws-lambda"
        LANGFUSE_CODE_EVAL_AWS_LAMBDA_NODE_FUNCTION_NAME   = try(local.code_evaluators.aws_lambda.node_function_name, "")
        LANGFUSE_CODE_EVAL_AWS_LAMBDA_PYTHON_FUNCTION_NAME = try(local.code_evaluators.aws_lambda.python_function_name, "")
        LANGFUSE_CODE_EVAL_AWS_LAMBDA_ENDPOINT             = try(local.code_evaluators.aws_lambda.endpoint, "")
        } : {
        LANGFUSE_CODE_EVAL_DISPATCHER       = "insecure-local"
        LANGFUSE_CODE_EVAL_LOCAL_TIMEOUT_MS = try(local.code_evaluators.insecure_local.timeout_ms, null) == null ? "" : tostring(local.code_evaluators.insecure_local.timeout_ms)
      },
      { LANGFUSE_CODE_EVAL_EXECUTION_QUEUE_SHARD_COUNT = try(local.code_evaluators.queue_shard_count, null) == null ? "" : tostring(local.code_evaluators.queue_shard_count) },
    ),
    local.masking == null ? {} : {
      LANGFUSE_INGESTION_MASKING_CALLBACK_URL         = local.masking.callback_url
      LANGFUSE_INGESTION_MASKING_CALLBACK_TIMEOUT_MS  = try(local.masking.timeout_ms, null) == null ? "" : tostring(local.masking.timeout_ms)
      LANGFUSE_INGESTION_MASKING_CALLBACK_FAIL_CLOSED = try(local.masking.fail_closed, false) ? "true" : ""
      LANGFUSE_INGESTION_MASKING_MAX_RETRIES          = try(local.masking.max_retries, null) == null ? "" : tostring(local.masking.max_retries)
      LANGFUSE_INGESTION_MASKING_PROPAGATED_HEADERS   = join(",", try(local.masking.propagated_headers, []))
    },
    merge([for prefix, a in {
      LANGFUSE_WEBHOOK                         = try(var.spec.network.allowlists.webhooks, null)
      LANGFUSE_LLM_CONNECTION                  = try(var.spec.network.allowlists.llm_connections, null)
      LANGFUSE_BLOB_STORAGE_ENDPOINT           = try(var.spec.network.allowlists.blob_storage_endpoints, null)
      LANGFUSE_EXTERNAL_MEDIA_STORAGE_ENDPOINT = try(var.spec.network.allowlists.external_media_storage_endpoints, null)
      LANGFUSE_SSO_DISCOVERY                   = try(var.spec.network.allowlists.sso_discovery, null)
      } : {
      "${prefix}_WHITELISTED_HOST"        = join(",", try(a.hosts, []))
      "${prefix}_WHITELISTED_IPS"         = join(",", try(a.ips, []))
      "${prefix}_WHITELISTED_IP_SEGMENTS" = join(",", try(a.ip_segments, []))
    }]...),
  )

  web_plain = merge(
    local.init_enabled ? {
      LANGFUSE_INIT_ORG_ID            = local.organization_id
      LANGFUSE_INIT_ORG_NAME          = coalesce(try(var.spec.init.organization.name, null), "Default Organization")
      LANGFUSE_INIT_PROJECT_ID        = local.project_id
      LANGFUSE_INIT_PROJECT_NAME      = coalesce(try(var.spec.init.project.name, null), "Default Project")
      LANGFUSE_INIT_PROJECT_RETENTION = try(var.spec.init.project.retention_days, null) == null ? "" : tostring(var.spec.init.project.retention_days)
    } : {},
    local.admin_seeded ? {
      LANGFUSE_INIT_USER_EMAIL = local.admin_email
      LANGFUSE_INIT_USER_NAME  = coalesce(try(var.spec.init.admin.name, null), "Admin")
    } : {},
    {
      # auth
      AUTH_EMAIL_VERIFICATION_REQUIRED  = try(var.spec.auth.email_verification_required, false) ? "true" : ""
      AUTH_SESSION_MAX_AGE              = try(var.spec.auth.session_max_age_minutes, null) == null ? "" : tostring(var.spec.auth.session_max_age_minutes)
      AUTH_DOMAINS_WITH_SSO_ENFORCEMENT = join(",", try(var.spec.auth.sso_enforced_domains, []))
      AUTH_IGNORE_ACCOUNT_FIELDS        = join(",", try(var.spec.auth.ignore_account_fields, []))
      AUTH_HTTPS_PROXY                  = try(var.spec.auth.https_proxy, "")
      AUTH_HTTP_PROXY                   = try(var.spec.auth.http_proxy, "")
      AUTH_SSO_TIMEOUT                  = try(var.spec.auth.sso_timeout_ms, null) == null ? "" : tostring(var.spec.auth.sso_timeout_ms)
      LANGFUSE_DEFAULT_ORG_ID           = try(var.spec.auth.new_user_access.organization_id, "")
      LANGFUSE_DEFAULT_ORG_ROLE         = local.enum_value.org_role
      LANGFUSE_DEFAULT_PROJECT_ID       = join(",", try(var.spec.auth.new_user_access.project_ids, []))
      LANGFUSE_DEFAULT_PROJECT_ROLE     = local.enum_value.project_role
      LANGFUSE_CUSTOM_SSO_EMAIL_CLAIM   = try(local.sso.custom.claims.email, "")
      LANGFUSE_CUSTOM_SSO_NAME_CLAIM    = try(local.sso.custom.claims.name, "")
      LANGFUSE_CUSTOM_SSO_SUB_CLAIM     = try(local.sso.custom.claims.subject, "")
      LANGFUSE_CUSTOM_SSO_IMAGE_CLAIM   = try(local.sso.custom.claims.image, "")
      # features (the chart's own experimental-features value sets a
      # variable Langfuse never reads; the module sets the one it does)
      LANGFUSE_ENABLE_EXPERIMENTAL_FEATURES    = try(var.spec.features.experimental_features_enabled, false) ? "true" : ""
      LANGFUSE_CSP_ENFORCE_HTTPS               = try(var.spec.features.csp_enforce_https, false) ? "true" : ""
      LANGFUSE_MARKDOWN_RENDER_CHARACTER_LIMIT = try(var.spec.features.markdown_render_character_limit, null) == null ? "" : tostring(var.spec.features.markdown_render_character_limit)
      # caching
      LANGFUSE_CACHE_API_KEY_ENABLED     = try(var.spec.caching.api_key_cache_enabled, null) == null ? "" : tostring(var.spec.caching.api_key_cache_enabled)
      LANGFUSE_CACHE_API_KEY_TTL_SECONDS = try(var.spec.caching.api_key_cache_ttl_seconds, null) == null ? "" : tostring(var.spec.caching.api_key_cache_ttl_seconds)
      # assistant
      LANGFUSE_IN_APP_AGENT_MAX_ACTIVE_RUNS_PER_USER = try(local.assistant.max_active_runs_per_user, null) == null ? "" : tostring(local.assistant.max_active_runs_per_user)
      LANGFUSE_IN_APP_AGENT_MAX_ACTIVE_RUNS_PER_ORG  = try(local.assistant.max_active_runs_per_org, null) == null ? "" : tostring(local.assistant.max_active_runs_per_org)
      LANGFUSE_MCP_ALLOWED_HOSTS                     = try(local.assistant.use_internal_web_url, false) ? local.web_service_name : ""
      # ui customization
      LANGFUSE_UI_DOCUMENTATION_HREF         = try(var.spec.enterprise.ui.documentation_href, "")
      LANGFUSE_UI_SUPPORT_HREF               = try(var.spec.enterprise.ui.support_href, "")
      LANGFUSE_UI_FEEDBACK_HREF              = try(var.spec.enterprise.ui.feedback_href, "")
      LANGFUSE_UI_API_HOST                   = try(var.spec.enterprise.ui.api_host, "")
      LANGFUSE_UI_LOGO_LIGHT_MODE_HREF       = try(var.spec.enterprise.ui.logo_light_mode_href, "")
      LANGFUSE_UI_LOGO_DARK_MODE_HREF        = try(var.spec.enterprise.ui.logo_dark_mode_href, "")
      LANGFUSE_UI_DEFAULT_MODEL_ADAPTER      = local.enum_value.model_adapter
      LANGFUSE_UI_DEFAULT_BASE_URL_OPENAI    = try(var.spec.enterprise.ui.default_base_url_openai, "")
      LANGFUSE_UI_DEFAULT_BASE_URL_ANTHROPIC = try(var.spec.enterprise.ui.default_base_url_anthropic, "")
      LANGFUSE_UI_DEFAULT_BASE_URL_AZURE     = try(var.spec.enterprise.ui.default_base_url_azure, "")
      LANGFUSE_UI_VISIBLE_PRODUCT_MODULES    = join(",", try(var.spec.enterprise.ui.visible_product_modules, []))
      LANGFUSE_UI_HIDDEN_PRODUCT_MODULES     = join(",", try(var.spec.enterprise.ui.hidden_product_modules, []))
      LANGFUSE_UI_INSTANCE_LINKS             = length(local.instance_links) > 0 ? jsonencode([for l in local.instance_links : { name = l.name, url = l.url }]) : ""
      # observability
      OTEL_TRACE_SAMPLING_RATIO = try(var.spec.observability.trace_sampling_ratio, null) == null ? "" : tostring(var.spec.observability.trace_sampling_ratio)
    },
    # single sign-on providers (twin: ssoProvider.env)
    merge([for p in local.sso_providers : merge(
      {
        "AUTH_${upper(p.id)}_CLIENT_ID"                    = p.client.client_id
        "AUTH_${upper(p.id)}_ALLOW_ACCOUNT_LINKING"        = try(p.client.allow_account_linking, false) ? "true" : ""
        "AUTH_${upper(p.id)}_CLIENT_AUTH_METHOD"           = endswith(try(p.client.client_auth_method, ""), "unspecified") ? "" : try(p.client.client_auth_method, "")
        "AUTH_${upper(p.id)}_CHECKS"                       = join(",", [for c in try(p.client.checks, []) : c == "no_checks" ? "none" : c])
        "AUTH_${upper(p.id)}_ID_TOKEN_SIGNED_RESPONSE_ALG" = endswith(try(p.client.id_token_signed_response_alg, ""), "unspecified") ? "" : try(p.client.id_token_signed_response_alg, "")
      },
      { for suffix, value in p.extra : "AUTH_${upper(p.id)}_${suffix}" => value },
    )]...),
  )

  worker_plain = merge(
    {
      BATCH_EXPORT_PAGE_SIZE                                          = try(var.spec.blob_storage.batch_export.page_size, null) == null ? "" : tostring(var.spec.blob_storage.batch_export.page_size)
      BATCH_EXPORT_ROW_LIMIT                                          = try(var.spec.blob_storage.batch_export.row_limit, null) == null ? "" : tostring(var.spec.blob_storage.batch_export.row_limit)
      BATCH_EXPORT_DOWNLOAD_LINK_EXPIRATION_HOURS                     = try(var.spec.blob_storage.batch_export.download_link_expiration_hours, null) == null ? "" : tostring(var.spec.blob_storage.batch_export.download_link_expiration_hours)
      LANGFUSE_INGESTION_QUEUE_PROCESSING_CONCURRENCY                 = try(var.spec.ingestion.queue_processing_concurrency, null) == null ? "" : tostring(var.spec.ingestion.queue_processing_concurrency)
      LANGFUSE_TRACE_UPSERT_WORKER_CONCURRENCY                        = try(var.spec.ingestion.trace_upsert_worker_concurrency, null) == null ? "" : tostring(var.spec.ingestion.trace_upsert_worker_concurrency)
      LANGFUSE_SKIP_INGESTION_CLICKHOUSE_READ_MIN_PROJECT_CREATE_DATE = try(var.spec.ingestion.skip_clickhouse_read_min_project_create_date, "")
      LANGFUSE_IN_APP_AGENT_RUN_QUEUE_PROCESSING_CONCURRENCY          = try(local.assistant.run_concurrency, null) == null ? "" : tostring(local.assistant.run_concurrency)
      LANGFUSE_MCP_BASE_URL                                           = try(local.assistant.use_internal_web_url, false) ? "http://${local.web_service_name}:${local.web_service_port}" : ""
      LANGFUSE_QUEUE_CONSUMPTION_STUCK_THRESHOLD_MINUTES              = try(var.spec.worker.queue_consumption_stuck_threshold_minutes, null) == null ? "" : tostring(var.spec.worker.queue_consumption_stuck_threshold_minutes)
    },
    local.sandbox == null ? {} : {
      LANGFUSE_IN_APP_AGENT_SANDBOX_PROVIDER                                        = "lambda-microvm"
      LANGFUSE_IN_APP_AGENT_SANDBOX_AWS_LAMBDA_MICROVM_IMAGE_IDENTIFIER             = local.sandbox.image_identifier
      LANGFUSE_IN_APP_AGENT_SANDBOX_AWS_LAMBDA_MICROVM_EXECUTION_ROLE_ARN           = local.sandbox.execution_role_arn
      LANGFUSE_IN_APP_AGENT_SANDBOX_AWS_LAMBDA_MICROVM_REGION                       = local.sandbox.region
      LANGFUSE_IN_APP_AGENT_SANDBOX_AWS_LAMBDA_MICROVM_EGRESS_NETWORK_CONNECTOR_ARN = local.sandbox.egress_network_connector_arn
    },
    local.code_evaluators == null ? {} : {
      LANGFUSE_CODE_EVAL_EXECUTION_WORKER_CONCURRENCY     = try(local.code_evaluators.concurrency, null) == null ? "" : tostring(local.code_evaluators.concurrency)
      QUEUE_CONSUMER_CODE_EVAL_EXECUTION_QUEUE_IS_ENABLED = "true"
    },
  )

  # ---- keys of the instance Secret -------------------------------------------
  common_instance = merge(
    contains(keys(local.declared_secrets), "postgres-direct-url") ? { DIRECT_URL = "postgres-direct-url" } : {},
    contains(keys(local.declared_secrets), "postgres-shadow-database-url") ? { SHADOW_DATABASE_URL = "postgres-shadow-database-url" } : {},
    contains(keys(local.declared_secrets), "redis-tls-key-passphrase") ? { REDIS_TLS_KEY_PASSPHRASE = "redis-tls-key-passphrase" } : {},
    contains(keys(local.declared_secrets), "ai-api-key") ? { LANGFUSE_AI_API_KEY = "ai-api-key" } : {},
    try(var.spec.smtp, null) != null ? { SMTP_CONNECTION_URL = "smtp-connection-url" } : {},
    try(var.spec.slack, null) != null ? { SLACK_CLIENT_SECRET = "slack-client-secret", SLACK_STATE_SECRET = "slack-state-secret" } : {},
  )
  web_instance = merge(
    local.init_enabled ? {
      LANGFUSE_INIT_PROJECT_PUBLIC_KEY = "init-project-public-key"
      LANGFUSE_INIT_PROJECT_SECRET_KEY = "init-project-secret-key"
    } : {},
    local.admin_seeded ? { LANGFUSE_INIT_USER_PASSWORD = "init-user-password" } : {},
    contains(keys(local.declared_secrets), "admin-api-key") ? { ADMIN_API_KEY = "admin-api-key" } : {},
    { for p in local.sso_providers : "AUTH_${upper(p.id)}_CLIENT_SECRET" => "sso-${p.id}-client-secret" },
  )

  # ---- assembled channels ----------------------------------------------------
  common_entries = merge(
    { for k, v in local.common_plain : k => { name = k, value = v, valueFrom = null } if v != "" },
    { for k, v in local.common_instance : k => { name = k, value = null, valueFrom = { secretKeyRef = { name = local.instance_secret_name, key = v } } } },
    { for k, v in try(var.spec.extra_env, {}) : k => { name = k, value = v, valueFrom = null } if v != "" },
    { for k, v in try(var.spec.extra_env_from_secret, {}) : k => { name = k, value = null, valueFrom = { secretKeyRef = { name = v.name, key = v.key } } } },
  )
  web_entries = merge(
    { for k, v in local.web_plain : k => { name = k, value = v, valueFrom = null } if v != "" },
    { for k, v in local.web_instance : k => { name = k, value = null, valueFrom = { secretKeyRef = { name = local.instance_secret_name, key = v } } } },
    { for k, v in try(var.spec.web.extra_env, {}) : k => { name = k, value = v, valueFrom = null } if v != "" },
    { for k, v in try(var.spec.web.extra_env_from_secret, {}) : k => { name = k, value = null, valueFrom = { secretKeyRef = { name = v.name, key = v.key } } } },
  )
  worker_entries = merge(
    { for k, v in local.worker_plain : k => { name = k, value = v, valueFrom = null } if v != "" },
    { for k, v in try(var.spec.worker.extra_env, {}) : k => { name = k, value = v, valueFrom = null } if v != "" },
    { for k, v in try(var.spec.worker.extra_env_from_secret, {}) : k => { name = k, value = null, valueFrom = { secretKeyRef = { name = v.name, key = v.key } } } },
  )

  common_env = [for k in sort(keys(local.common_entries)) : { for kk, vv in local.common_entries[k] : kk => vv if vv != null }]
  web_env    = [for k in sort(keys(local.web_entries)) : { for kk, vv in local.web_entries[k] : kk => vv if vv != null }]
  worker_env = [for k in sort(keys(local.worker_entries)) : { for kk, vv in local.worker_entries[k] : kk => vv if vv != null }]
}
