# What every file of the module reads, resolved once, and the documents the
# module applies: the chart's values, the Secrets, and the Stigmer. The Pulumi
# twin is module/vars.go, locals.go and render.go, and the cross-engine test
# (iac/pulumi/module/parity_test.go) holds the rendered documents together.
#
# RULE: a key renders only when the spec declares it, so the chart's and the
# operator's own defaults stay authoritative for everything else; an empty map
# or list is never written.
#
# PATTERN: every object is built with null for an undeclared key, then
# filtered with { for k, v in RAW : k => v if v != null }. HCL refuses a
# conditional whose two branches are objects of different shapes, so a key is
# never added by `cond ? { k = v } : {}`; it is set to null instead.

locals {
  # ---- fixed names (vars.go) ------------------------------------------------
  chart_repository                  = "oci://ghcr.io/stigmer/charts"
  chart_name                        = "stigmer-operator"
  operator_release                  = "stigmer-operator"
  stigmer_api_version               = "stigmer.ai/v1"
  stigmer_kind                      = "Stigmer"
  stigmer_name                      = "stigmer"
  stigmer_definition                = "stigmers.stigmer.ai"
  upgrade_without_backup_annotation = "stigmer.ai/upgrade-without-backup"
  keys_secret                       = "stigmer-keys"
  server_port                       = 7234
  artifacts_port                    = 7235
  model_secret                      = "stigmer-model-key"
  object_store_secret               = "stigmer-object-store-key"
  database_secret                   = "stigmer-database-password"
  temporal_api_key_secret           = "stigmer-temporal-api-key"
  temporal_tls_secret               = "stigmer-temporal-tls"
  smtp_secret                       = "stigmer-smtp-login"
  license_secret                    = "stigmer-license"
  server_extra_env_secret           = "stigmer-server-extra-env"
  runner_extra_env_secret           = "stigmer-runner-extra-env"
  database_password_key             = "password"
  temporal_api_key_key              = "api-key"
  temporal_ca_key                   = "ca.crt"
  temporal_cert_key                 = "tls.crt"
  temporal_private_key_key          = "tls.key"
  license_key                       = "STIGMER_LICENSE_KEY"
  keys_gate                         = "stigmer-keys-check"
  cleanup_wait_seconds              = 600

  # ---- identity (locals.go) -------------------------------------------------
  kind_name          = "KubernetesStigmer"
  owner_label        = "planton.ai/kind"
  namespace          = var.spec.namespace
  operator_namespace = var.spec.operator.namespace
  create_namespace   = try(var.spec.operator.create_namespace, false) == true

  # Labels on every object the module writes itself. Their planton.ai/kind is
  # also the ownership mark the preflight reads.
  labels = { for k, v in {
    "planton.ai/resource"     = "true"
    "planton.ai/name"         = var.metadata.name
    "planton.ai/kind"         = local.kind_name
    "planton.ai/id"           = try(var.metadata.id, "")
    "planton.ai/organization" = try(var.metadata.org, "")
    "planton.ai/environment"  = try(var.metadata.env, "")
  } : k => v if v != "" }
}

# ---- the chart's values (render.go helmValues) -------------------------------
locals {
  op = var.spec.operator

  chart_image = { for k, v in {
    repository  = try(local.op.image_repository, "") == "" ? null : local.op.image_repository
    pullPolicy  = try(local.op.image_pull_policy, "") == "" ? null : local.op.image_pull_policy
    pullSecrets = length(try(local.op.image_pull_secrets, [])) == 0 ? null : [for n in local.op.image_pull_secrets : { name = n }]
  } : k => v if v != null }

  chart_resources = { for k, v in {
    requests = length(try(local.op.resources.requests, {})) == 0 ? null : local.op.resources.requests
    limits   = length(try(local.op.resources.limits, {})) == 0 ? null : local.op.resources.limits
  } : k => v if v != null }

  chart_values = { for k, v in {
    image        = length(local.chart_image) == 0 ? null : local.chart_image
    replicas     = try(local.op.replicas, null)
    resources    = length(local.chart_resources) == 0 ? null : local.chart_resources
    logLevel     = try(local.op.log_level, "") == "" ? null : local.op.log_level
    nodeSelector = length(try(local.op.node_selector, {})) == 0 ? null : local.op.node_selector
    tolerations = length(try(local.op.tolerations, [])) == 0 ? null : [for t in local.op.tolerations : { for tk, tv in {
      key               = try(t.key, "") == "" ? null : t.key
      operator          = try(t.operator, "") == "" ? null : t.operator
      value             = try(t.value, "") == "" ? null : t.value
      effect            = try(t.effect, "") == "" ? null : t.effect
      tolerationSeconds = try(t.toleration_seconds, null)
    } : tk => tv if tv != null }]
  } : k => v if v != null }
}

# ---- the Secrets (render.go materializedSecrets) -----------------------------
locals {
  temporal_ext = try(var.spec.temporal.external, null)
  temporal_pem = try(local.temporal_ext.tls.pem, null)

  # Each Secret holds only the keys the spec gives values for; a Secret with
  # none is not written.
  secrets_raw = {
    (local.model_secret) = { for k, v in {
      ANTHROPIC_API_KEY         = try(var.spec.runner.model.keys.anthropic_api_key, "")
      ANTHROPIC_BASE_URL        = try(var.spec.runner.model.keys.anthropic_base_url, "")
      OPENAI_API_KEY            = try(var.spec.runner.model.keys.openai_api_key, "")
      ANTHROPIC_FOUNDRY_API_KEY = try(var.spec.runner.model.keys.anthropic_foundry_api_key, "")
      AWS_BEARER_TOKEN_BEDROCK  = try(var.spec.runner.model.keys.aws_bearer_token_bedrock, "")
    } : k => v if v != "" && v != null }
    (local.object_store_secret) = { for k, v in {
      ACCESS_KEY_ID     = try(var.spec.object_store.s3.credentials.access_key.access_key_id, "")
      SECRET_ACCESS_KEY = try(var.spec.object_store.s3.credentials.access_key.secret_access_key, "")
    } : k => v if v != "" && v != null }
    (local.database_secret) = { for k, v in {
      (local.database_password_key) = try(var.spec.database.external.password.value, "")
    } : k => v if v != "" && v != null }
    (local.temporal_api_key_secret) = { for k, v in {
      (local.temporal_api_key_key) = try(local.temporal_ext.api_key.value, "")
    } : k => v if v != "" && v != null }
    (local.temporal_tls_secret) = { for k, v in {
      (local.temporal_ca_key)          = try(local.temporal_pem.ca_certificate, "")
      (local.temporal_cert_key)        = try(local.temporal_pem.client_certificate, "")
      (local.temporal_private_key_key) = try(local.temporal_pem.client_private_key, "")
    } : k => v if v != "" && v != null }
    (local.smtp_secret) = { for k, v in {
      username = try(var.spec.identity.bundled.smtp.credentials.login.username, "")
      password = try(var.spec.identity.bundled.smtp.credentials.login.password, "")
    } : k => v if v != "" && v != null }
    (local.license_secret) = { for k, v in {
      (local.license_key) = try(var.spec.enterprise.license.value, "")
    } : k => v if v != "" && v != null }
    (local.server_extra_env_secret) = { for s in try(var.spec.server.extra_env.secrets, []) : s.name => s.value if try(s.value, null) != null && try(s.value, "") != "" }
    (local.runner_extra_env_secret) = { for s in try(var.spec.runner.extra_env.secrets, []) : s.name => s.value if try(s.value, null) != null && try(s.value, "") != "" }
  }
  secrets = { for name, data in local.secrets_raw : name => data if length(data) > 0 }
}

# ---- the Stigmer (render.go stigmerSpec) --------------------------------------
locals {
  # The definition's ResourceRequirements and volumes, for each block of the
  # spec that carries one (HCL has no functions of its own, so each shared
  # shape is rendered once for every place it appears).
  resource_blocks = {
    server    = try(var.spec.server.resources, null)
    runner    = try(var.spec.runner.resources, null)
    database  = try(var.spec.database.bundled.resources, null)
    temporal  = try(var.spec.temporal.bundled.resources, null)
    substrate = try(var.spec.sandboxes.substrate.worker_resources, null)
  }
  resources = { for name, r in local.resource_blocks : name => { for k, v in {
    requests = length(try(r.requests, {})) == 0 ? null : r.requests
    limits   = length(try(r.limits, {})) == 0 ? null : r.limits
    claims = length(try(r.claims, [])) == 0 ? null : [for c in r.claims : { for ck, cv in {
      name    = c.name
      request = try(c.request, "") == "" ? null : c.request
    } : ck => cv if cv != null }]
  } : k => v if v != null } }

  volume_blocks = {
    server_storage   = try(var.spec.server.storage, null)
    server_artifacts = try(var.spec.server.artifacts, null)
    runner_storage   = try(var.spec.runner.storage, null)
    database_storage = try(var.spec.database.bundled.storage, null)
  }
  volumes = { for name, b in local.volume_blocks : name => { for k, v in {
    size             = try(b.size, "") == "" ? null : b.size
    storageClassName = try(b.storage_class_name, "") == "" ? null : b.storage_class_name
  } : k => v if v != null } }

  # extraEnv (render.go extraEnv): plain values and Kubernetes sources as
  # written, every secret by reference (the module's Secret for a value, the
  # user's own for a reference).
  env_blocks = {
    server = { env = try(var.spec.server.extra_env, null), secret = local.server_extra_env_secret }
    runner = { env = try(var.spec.runner.extra_env, null), secret = local.runner_extra_env_secret }
  }
  extra_env = { for name, b in local.env_blocks : name => concat(
    [for v in try(b.env.variables, []) : { for k, val in {
      name  = v.name
      value = try(v.value, null)
      valueFrom = try(v.value, null) != null ? null : { for sk, sv in {
        configMapKeyRef = try(v.config_map_key_ref, null) == null ? null : { for rk, rv in {
          name     = v.config_map_key_ref.name
          key      = v.config_map_key_ref.key
          optional = try(v.config_map_key_ref.optional, false) ? true : null
        } : rk => rv if rv != null }
        fieldRef = try(v.field_ref, null) == null ? null : { for rk, rv in {
          fieldPath  = v.field_ref.field_path
          apiVersion = try(v.field_ref.api_version, "") == "" ? null : v.field_ref.api_version
        } : rk => rv if rv != null }
        resourceFieldRef = try(v.resource_field_ref, null) == null ? null : { for rk, rv in {
          resource      = v.resource_field_ref.resource
          containerName = try(v.resource_field_ref.container_name, "") == "" ? null : v.resource_field_ref.container_name
          divisor       = try(v.resource_field_ref.divisor, "") == "" ? null : v.resource_field_ref.divisor
        } : rk => rv if rv != null }
      } : sk => sv if sv != null }
      } : k => val if val != null } if try(v.value, null) != null || try(v.config_map_key_ref, null) != null || try(v.field_ref, null) != null || try(v.resource_field_ref, null) != null
    ],
    [for s in try(b.env.secrets, []) : {
      name = s.name
      valueFrom = { secretKeyRef = { for k, v in {
        name     = try(s.secret_ref, null) != null ? s.secret_ref.name : b.secret
        key      = try(s.secret_ref, null) != null ? s.secret_ref.key : s.name
        optional = try(s.secret_ref.optional, false) ? true : null
      } : k => v if v != null } }
    } if try(s.secret_ref, null) != null || (try(s.value, null) != null && try(s.value, "") != "")],
  ) }

  substrate_tolerations = [for t in try(var.spec.sandboxes.substrate.worker_tolerations, []) : { for k, v in {
    key               = try(t.key, "") == "" ? null : t.key
    operator          = try(t.operator, "") == "" ? null : t.operator
    value             = try(t.value, "") == "" ? null : t.value
    effect            = try(t.effect, "") == "" ? null : t.effect
    tolerationSeconds = try(t.toleration_seconds, null)
  } : k => v if v != null }]

  # ---- spec.image, owner, server, runner --------------------------------------
  image = { for k, v in {
    registry    = try(var.spec.image.registry, "") == "" ? null : var.spec.image.registry
    pullPolicy  = try(var.spec.image.pull_policy, "") == "" ? null : var.spec.image.pull_policy
    pullSecrets = length(try(var.spec.image.pull_secrets, [])) == 0 ? null : [for n in var.spec.image.pull_secrets : { name = n }]
  } : k => v if v != null }

  owner = try(var.spec.owner, null) == null ? null : { for k, v in {
    email = var.spec.owner.email
    name  = try(var.spec.owner.name, "") == "" ? null : var.spec.owner.name
  } : k => v if v != null }

  server = { for k, v in {
    resources = length(local.resources.server) == 0 ? null : local.resources.server
    storage   = length(local.volumes.server_storage) == 0 ? null : local.volumes.server_storage
    artifacts = length(local.volumes.server_artifacts) == 0 ? null : local.volumes.server_artifacts
    help = try(var.spec.server.help, null) == null ? null : { for hk, hv in {
      url   = var.spec.server.help.url
      label = try(var.spec.server.help.label, "") == "" ? null : var.spec.server.help.label
    } : hk => hv if hv != null }
    internalAccounts             = length(try(var.spec.server.internal_accounts, [])) == 0 ? null : var.spec.server.internal_accounts
    allowUnauthenticatedExposure = try(var.spec.server.allow_unauthenticated_exposure, false) ? true : null
    egress = try(var.spec.server.egress, null) == null ? null : { for ek, ev in {
      fenced     = try(var.spec.server.egress.fenced, false) ? true : null
      allowCIDRs = length(try(var.spec.server.egress.allow_cidrs, [])) == 0 ? null : var.spec.server.egress.allow_cidrs
    } : ek => ev if ev != null }
    extraEnv = length(local.extra_env.server) == 0 ? null : local.extra_env.server
  } : k => v if v != null }

  runner = { for k, v in {
    llm = try(var.spec.runner.model, null) == null ? null : {
      secretName = try(var.spec.runner.model.keys, null) != null ? local.model_secret : var.spec.runner.model.existing_secret_name
    }
    resources = length(local.resources.runner) == 0 ? null : local.resources.runner
    storage   = length(local.volumes.runner_storage) == 0 ? null : local.volumes.runner_storage
    extraEnv  = length(local.extra_env.runner) == 0 ? null : local.extra_env.runner
  } : k => v if v != null }

  # ---- spec.database -------------------------------------------------------------
  db_external = try(var.spec.database.external, null)
  db_bundled  = try(var.spec.database.bundled, null)
  db_password = try(local.db_external.password.existing_secret, null)
  database = try(var.spec.database, null) == null ? null : {
    postgresql = { for k, v in {
      external = local.db_external == null ? null : { for ek, ev in {
        host = local.db_external.host
        user = local.db_external.user
        passwordSecret = {
          name = local.db_password != null ? local.db_password.name : local.database_secret
          key  = local.db_password != null ? coalesce(try(local.db_password.key, null), local.database_password_key) : local.database_password_key
        }
        port     = try(local.db_external.port, null)
        database = try(local.db_external.database, "") == "" ? null : local.db_external.database
        sslMode  = try(local.db_external.ssl_mode, "") == "" ? null : local.db_external.ssl_mode
      } : ek => ev if ev != null }
      bundled = local.db_bundled == null ? null : { for bk, bv in {
        instances = try(local.db_bundled.instances, null)
        storage   = length(local.volumes.database_storage) == 0 ? null : local.volumes.database_storage
        resources = length(local.resources.database) == 0 ? null : local.resources.database
        backup = try(local.db_bundled.backup, null) == null ? null : { for xk, xv in {
          schedule  = try(local.db_bundled.backup.schedule, "") == "" ? null : local.db_bundled.backup.schedule
          retention = try(local.db_bundled.backup.retention, "") == "" ? null : local.db_bundled.backup.retention
        } : xk => xv if xv != null }
        recoverFrom = try(local.db_bundled.recover_from, null) == null ? null : { for rk, rv in {
          serverName = local.db_bundled.recover_from.server_name
          targetTime = try(local.db_bundled.recover_from.target_time, "") == "" ? null : local.db_bundled.recover_from.target_time
        } : rk => rv if rv != null }
      } : bk => bv if bv != null }
    } : k => v if v != null }
  }

  # ---- spec.temporal -------------------------------------------------------------
  temporal_tls      = try(local.temporal_ext.tls, null)
  temporal_tls_from = try(local.temporal_tls.existing_secret, null)
  temporal_api_key  = try(local.temporal_ext.api_key.existing_secret, null)
  temporal = try(var.spec.temporal, null) == null ? null : { for k, v in {
    bundled = try(var.spec.temporal.bundled, null) == null ? null : { for bk, bv in {
      resources = length(local.resources.temporal) == 0 ? null : local.resources.temporal
    } : bk => bv if bv != null }
    external = local.temporal_ext == null ? null : { for ek, ev in {
      hostPort  = local.temporal_ext.host_port
      namespace = try(local.temporal_ext.namespace, "") == "" ? null : local.temporal_ext.namespace
      tls = local.temporal_tls == null ? null : { for tk, tv in {
        serverName = try(local.temporal_tls.server_name, "") == "" ? null : local.temporal_tls.server_name
        secretName = local.temporal_pem != null ? local.temporal_tls_secret : (local.temporal_tls_from != null ? local.temporal_tls_from.name : null)
        caKey = (local.temporal_pem != null ? (try(local.temporal_pem.ca_certificate, "") == "" ? null : local.temporal_ca_key)
        : (try(local.temporal_tls_from.ca_key, "") == "" ? null : local.temporal_tls_from.ca_key))
        certKey = (local.temporal_pem != null ? (try(local.temporal_pem.client_certificate, "") == "" ? null : local.temporal_cert_key)
        : (try(local.temporal_tls_from.certificate_key, "") == "" ? null : local.temporal_tls_from.certificate_key))
        keyKey = (local.temporal_pem != null ? (try(local.temporal_pem.client_certificate, "") == "" ? null : local.temporal_private_key_key)
        : (try(local.temporal_tls_from.private_key_key, "") == "" ? null : local.temporal_tls_from.private_key_key))
      } : tk => tv if tv != null }
      apiKey = try(local.temporal_ext.api_key, null) == null ? null : {
        name = local.temporal_api_key != null ? local.temporal_api_key.name : local.temporal_api_key_secret
        key  = local.temporal_api_key != null ? local.temporal_api_key.key : local.temporal_api_key_key
      }
    } : ek => ev if ev != null }
  } : k => v if v != null }

  # ---- spec.objectStore, ingress, identity, enterprise ---------------------------
  s3 = try(var.spec.object_store.s3, null)
  object_store = local.s3 == null ? null : {
    s3 = { for k, v in {
      bucket = local.s3.bucket
      credentialsSecret = {
        name = try(local.s3.credentials.access_key, null) != null ? local.object_store_secret : local.s3.credentials.existing_secret_name
      }
      region         = try(local.s3.region, "") == "" ? null : local.s3.region
      endpoint       = try(local.s3.endpoint, "") == "" ? null : local.s3.endpoint
      runnerEndpoint = try(local.s3.runner_endpoint, "") == "" ? null : local.s3.runner_endpoint
    } : k => v if v != null }
  }

  ing = try(var.spec.ingress, null)
  ingress = local.ing == null ? null : { for k, v in {
    hostname           = local.ing.hostname
    className          = try(local.ing.class_name, "") == "" ? null : local.ing.class_name
    annotations        = length(try(local.ing.annotations, {})) == 0 ? null : local.ing.annotations
    serviceAnnotations = length(try(local.ing.service_annotations, {})) == 0 ? null : local.ing.service_annotations
    tls = try(local.ing.tls, null) == null ? null : { for tk, tv in {
      secretName = try(local.ing.tls.secret_name, "") == "" ? null : local.ing.tls.secret_name
      issuer = try(local.ing.tls.issuer, null) == null ? null : { for ik, iv in {
        name = local.ing.tls.issuer.name
        kind = try(local.ing.tls.issuer.kind, "") == "" ? null : local.ing.tls.issuer.kind
      } : ik => iv if iv != null }
    } : tk => tv if tv != null }
    gatewayRef = try(local.ing.gateway, null) == null ? null : { for gk, gv in {
      name        = local.ing.gateway.name
      namespace   = try(local.ing.gateway.namespace, "") == "" ? null : local.ing.gateway.namespace
      sectionName = try(local.ing.gateway.section_name, "") == "" ? null : local.ing.gateway.section_name
    } : gk => gv if gv != null }
  } : k => v if v != null }

  idn  = try(var.spec.identity, null)
  smtp = try(local.idn.bundled.smtp, null)
  identity = local.idn == null ? null : { for k, v in {
    oidc = try(local.idn.oidc, null) == null ? null : { for ok, ov in {
      issuer          = local.idn.oidc.issuer
      audience        = local.idn.oidc.audience
      consoleClientId = try(local.idn.oidc.console_client_id, "") == "" ? null : local.idn.oidc.console_client_id
      nativeClientId  = try(local.idn.oidc.native_client_id, "") == "" ? null : local.idn.oidc.native_client_id
      discoveryUrl    = try(local.idn.oidc.discovery_url, "") == "" ? null : local.idn.oidc.discovery_url
    } : ok => ov if ov != null }
    bundled = try(local.idn.bundled, null) == null ? null : { for bk, bv in {
      signUp = length(try(local.idn.bundled.sign_up_email_domains, [])) == 0 ? null : { emailDomains = local.idn.bundled.sign_up_email_domains }
      smtp = local.smtp == null ? null : { for sk, sv in {
        host     = local.smtp.host
        from     = local.smtp.from
        port     = try(local.smtp.port, null)
        starttls = try(local.smtp.starttls, false) ? true : null
        secretName = try(local.smtp.credentials, null) == null ? null : (
          try(local.smtp.credentials.login, null) != null ? local.smtp_secret : local.smtp.credentials.existing_secret_name
        )
      } : sk => sv if sv != null }
    } : bk => bv if bv != null }
  } : k => v if v != null }

  ent         = try(var.spec.enterprise, null)
  ent_license = try(local.ent.license.existing_secret, null)
  enterprise = local.ent == null ? null : { for k, v in {
    license = {
      name = local.ent_license != null ? local.ent_license.name : local.license_secret
      key  = local.ent_license != null ? local.ent_license.key : local.license_key
    }
    organizationCreators       = length(try(local.ent.organization_creators, [])) == 0 ? null : local.ent.organization_creators
    accessHistoryRetentionDays = try(local.ent.access_history_retention_days, null)
    outboundEgress             = try(local.ent.outbound_egress, "") == "" ? null : local.ent.outbound_egress
  } : k => v if v != null }

  # ---- spec.sandboxes and prerequisites ------------------------------------------
  sbx = try(var.spec.sandboxes, null)
  sub = try(local.sbx.substrate, null)
  sandboxes = local.sbx == null ? null : { for k, v in {
    driver      = try(local.sbx.driver, "") == "" ? null : local.sbx.driver
    idleSeconds = try(local.sbx.idle_seconds, null)
    agentSandbox = try(local.sbx.agent_sandbox, null) == null ? null : { for ak, av in {
      egress = length(try(local.sbx.agent_sandbox.egress_allow_cidrs, [])) == 0 ? null : { allowCIDRs = local.sbx.agent_sandbox.egress_allow_cidrs }
    } : ak => av if av != null }
    substrate = local.sub == null ? null : { for sk, sv in {
      workers = try(local.sub.workers, null)
      worker = { for wk, wv in {
        resources    = length(local.resources.substrate) == 0 ? null : local.resources.substrate
        nodeSelector = length(try(local.sub.worker_node_selector, {})) == 0 ? null : local.sub.worker_node_selector
        tolerations  = length(local.substrate_tolerations) == 0 ? null : local.substrate_tolerations
      } : wk => wv if wv != null }
      sandbox = { for xk, xv in {
        cpuMillicores = try(local.sub.sandbox_cpu_millicores, null)
        memoryMiB     = try(local.sub.sandbox_memory_mib, null)
      } : xk => xv if xv != null }
      egress = length(try(local.sub.egress_http, [])) == 0 ? null : { http = local.sub.egress_http }
    } : sk => sv if sv != null && sv != {} }
  } : k => v if v != null }

  prerequisites = { for k, v in {
    postgresOperator     = try(var.spec.prerequisites.postgres_operator, "")
    postgresBackupPlugin = try(var.spec.prerequisites.postgres_backup_plugin, "")
    agentSandbox         = try(var.spec.prerequisites.agent_sandbox, "")
    substrate            = try(var.spec.prerequisites.substrate, "")
  } : k => v if v != "" && v != null }

  stigmer_spec = { for k, v in {
    image         = length(local.image) == 0 ? null : local.image
    owner         = local.owner
    server        = length(local.server) == 0 ? null : local.server
    runner        = length(local.runner) == 0 ? null : local.runner
    database      = local.database
    temporal      = local.temporal
    objectStore   = local.object_store
    ingress       = local.ingress
    identity      = local.identity
    enterprise    = local.enterprise
    sandboxes     = local.sandboxes
    prerequisites = length(local.prerequisites) == 0 ? null : local.prerequisites
  } : k => v if v != null }

  stigmer_metadata = { for k, v in {
    name      = local.stigmer_name
    namespace = local.namespace
    labels    = local.labels
    annotations = try(var.spec.upgrade_without_backup_to, "") == "" ? null : {
      (local.upgrade_without_backup_annotation) = "v${var.spec.upgrade_without_backup_to}"
    }
  } : k => v if v != null }

  stigmer_document = {
    apiVersion = local.stigmer_api_version
    kind       = local.stigmer_kind
    metadata   = local.stigmer_metadata
    spec       = local.stigmer_spec
  }

  # Stigmer's own rule for when the bundled sign-in runs (outputs.go).
  bundled_sign_in = try(var.spec.identity, null) != null ? try(var.spec.identity.bundled, null) != null : (
    try(var.spec.ingress, null) != null && !try(var.spec.server.allow_unauthenticated_exposure, false)
  )
}
