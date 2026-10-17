# Every resolution here has an exact twin in the Pulumi module (locals.go,
# blob.go, env.go, sso.go, values.go) - keep them in lockstep.

locals {
  # ---- identity -------------------------------------------------------------
  helm_chart_name = "langfuse"
  helm_chart_repo = "https://langfuse.github.io/langfuse-k8s"
  chart_version   = coalesce(try(var.spec.chart_version, null), "2.1.5")

  release_name = var.metadata.name
  namespace    = var.spec.namespace
  external_url = coalesce(try(var.spec.external_url, null), "http://localhost:3000")

  # Resource-identity labels: stamped on the module-created objects
  # (namespace, instance Secret) and on both workloads' pods.
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesLangfuse"
    },
    try(var.metadata.id, "") != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    try(var.metadata.org, "") != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    try(var.metadata.env, "") != "" ? { "planton.ai/environment" = var.metadata.env } : {},
  )

  web_port = coalesce(try(var.spec.web.service.port, null), 3000)
  # The Service exposes external_port when set, the container port
  # otherwise (the chart's `externalPort | default port`).
  web_service_port     = coalesce(try(var.spec.web.service.external_port, null), local.web_port)
  web_service_name     = "${local.release_name}-web"
  instance_secret_name = "${local.release_name}-secrets"

  endpoint             = "http://${local.web_service_name}.${local.namespace}.svc.cluster.local:${local.web_service_port}"
  otel_endpoint        = "${local.endpoint}/api/public/otel"
  port_forward_command = "kubectl port-forward -n ${local.namespace} svc/${local.web_service_name} ${local.web_service_port}:${local.web_service_port}"

  # ---- seeding (headless initialization) ------------------------------------
  init_enabled    = try(var.spec.init.enabled, null) == null ? true : var.spec.init.enabled
  organization_id = local.init_enabled ? coalesce(try(var.spec.init.organization.id, ""), local.release_name) : ""
  project_id      = local.init_enabled ? coalesce(try(var.spec.init.project.id, ""), "${local.release_name}-project") : ""
  declared_keys   = try(var.spec.init.project.public_key, "") != ""
  admin_seeded    = local.init_enabled && try(var.spec.init.admin, null) != null
  admin_email     = local.admin_seeded ? var.spec.init.admin.email : ""

  # ---- single sign-on providers (fixed order; twin: sso.go ssoProviders) ----
  sso = try(var.spec.auth.sso, null)
  sso_providers = [for p in [
    { id = "google", p = try(local.sso.google, null), extra = { ALLOWED_DOMAINS = join(",", try(local.sso.google.allowed_domains, [])) } },
    { id = "github", p = try(local.sso.github, null), extra = {} },
    { id = "github_enterprise", p = try(local.sso.github_enterprise, null), extra = { BASE_URL = try(local.sso.github_enterprise.base_url, "") } },
    { id = "gitlab", p = try(local.sso.gitlab, null), extra = { URL = try(local.sso.gitlab.url, ""), ISSUER = try(local.sso.gitlab.issuer, "") } },
    { id = "azure_ad", p = try(local.sso.azure_ad, null), extra = { TENANT_ID = try(local.sso.azure_ad.tenant_id, "") } },
    { id = "okta", p = try(local.sso.okta, null), extra = { ISSUER = try(local.sso.okta.issuer, "") } },
    { id = "authentik", p = try(local.sso.authentik, null), extra = { ISSUER = try(local.sso.authentik.issuer, ""), AUTHORIZATION_URL = try(local.sso.authentik.authorization_url, "") } },
    { id = "onelogin", p = try(local.sso.onelogin, null), extra = { ISSUER = try(local.sso.onelogin.issuer, "") } },
    { id = "auth0", p = try(local.sso.auth0, null), extra = { ISSUER = try(local.sso.auth0.issuer, "") } },
    { id = "cognito", p = try(local.sso.cognito, null), extra = { ISSUER = try(local.sso.cognito.issuer, "") } },
    { id = "keycloak", p = try(local.sso.keycloak, null), extra = {
      ISSUER   = try(local.sso.keycloak.issuer, "")
      NAME     = try(local.sso.keycloak.name, "")
      SCOPE    = try(local.sso.keycloak.scope, "")
      ID_TOKEN = try(local.sso.keycloak.use_id_token, null) == null ? "" : tostring(local.sso.keycloak.use_id_token)
    } },
    { id = "jumpcloud", p = try(local.sso.jumpcloud, null), extra = { ISSUER = try(local.sso.jumpcloud.issuer, ""), SCOPE = try(local.sso.jumpcloud.scope, "") } },
    { id = "custom", p = try(local.sso.custom, null), extra = {
      ISSUER         = try(local.sso.custom.issuer, "")
      NAME           = try(local.sso.custom.name, "")
      SCOPE          = try(local.sso.custom.scope, "")
      ID_TOKEN       = try(local.sso.custom.use_id_token, null) == null ? "" : tostring(local.sso.custom.use_id_token)
      FETCH_USERINFO = try(local.sso.custom.fetch_userinfo, false) ? "true" : ""
    } },
    { id = "workos", p = try(local.sso.workos, null), extra = { ORGANIZATION_ID = try(local.sso.workos.organization_id, ""), CONNECTION_ID = try(local.sso.workos.connection_id, "") } },
    # WordPress takes the shared client message itself.
    { id = "wordpress", p = try(local.sso.wordpress, null) == null ? null : { client = local.sso.wordpress }, extra = {} },
  ] : { id = p.id, client = p.p.client, extra = p.extra } if p.p != null]

  # ---- blob storage targets --------------------------------------------------
  # Each upload type's resolved target (twin: blob.go resolveBlobTarget).
  # The access key blocks are uniform {value, secretKeyRef} objects with the
  # unused half null, pruned when rendered.
  blob_inputs = {
    event = var.spec.blob_storage.target
    media = try(var.spec.blob_storage.media_upload.target, null)
    batch = try(var.spec.blob_storage.batch_export.target, null)
  }
  blob_resolved = { for uc, t in local.blob_inputs : uc => t == null ? null : (
    try(t.s3_compatible, null) != null ? {
      family           = "s3"
      bucket           = t.s3_compatible.bucket
      region           = coalesce(try(t.s3_compatible.region, null), "auto")
      endpoint         = t.s3_compatible.endpoint
      force_path_style = try(t.s3_compatible.force_path_style, null) == null ? true : t.s3_compatible.force_path_style
      access_key_id = try(t.s3_compatible.credentials_secret, null) != null ? {
        value        = null
        secretKeyRef = { name = t.s3_compatible.credentials_secret.secret_name, key = coalesce(try(t.s3_compatible.credentials_secret.access_key_id_key, null), "admin_access_key_id") }
        } : try(t.s3_compatible.access_key, null) != null ? {
        value        = t.s3_compatible.access_key.access_key_id
        secretKeyRef = null
      } : null
      secret_access_key = try(t.s3_compatible.credentials_secret, null) != null ? {
        value        = null
        secretKeyRef = { name = t.s3_compatible.credentials_secret.secret_name, key = coalesce(try(t.s3_compatible.credentials_secret.secret_access_key_key, null), "admin_secret_access_key") }
        } : try(t.s3_compatible.access_key, null) != null ? {
        value        = null
        secretKeyRef = { name = local.instance_secret_name, key = "blob-${uc}-secret-access-key" }
      } : null
      } : try(t.s3, null) != null ? {
      family           = "s3"
      bucket           = t.s3.bucket
      region           = t.s3.region
      endpoint         = try(t.s3.endpoint, "")
      force_path_style = false
      access_key_id    = try(t.s3.access_key, null) != null ? { value = t.s3.access_key.access_key_id, secretKeyRef = null } : null
      secret_access_key = try(t.s3.access_key, null) != null ? {
        value        = null
        secretKeyRef = { name = local.instance_secret_name, key = "blob-${uc}-secret-access-key" }
      } : null
      } : try(t.r2, null) != null ? {
      # R2 serves path-style and virtual-hosted requests alike; path style
      # needs no per-bucket DNS name.
      family            = "s3"
      bucket            = t.r2.bucket
      region            = "auto"
      endpoint          = t.r2.endpoint
      force_path_style  = true
      access_key_id     = { value = t.r2.credentials.access_key_id, secretKeyRef = null }
      secret_access_key = { value = null, secretKeyRef = { name = local.instance_secret_name, key = "blob-${uc}-secret-access-key" } }
      } : try(t.gcs, null) != null ? {
      family            = "gcs"
      bucket            = t.gcs.bucket
      region            = ""
      endpoint          = ""
      force_path_style  = false
      access_key_id     = null
      secret_access_key = null
      } : {
      # Langfuse's Azure driver reads the account name from the access key
      # id and the account key from the secret access key.
      family            = "azure"
      bucket            = t.azure_blob.container
      region            = ""
      endpoint          = try(coalesce(t.azure_blob.endpoint), "https://${t.azure_blob.storage_account}.blob.core.windows.net")
      force_path_style  = false
      access_key_id     = { value = t.azure_blob.storage_account, secretKeyRef = null }
      secret_access_key = { value = null, secretKeyRef = { name = local.instance_secret_name, key = "blob-${uc}-azure-account-key" } }
    }
  ) }
  event_target = local.blob_resolved.event
  media_target = local.blob_resolved.media != null ? local.blob_resolved.media : local.blob_resolved.event
  batch_target = local.blob_resolved.batch != null ? local.blob_resolved.batch : local.blob_resolved.event

  blob_declared = merge([for uc, t in local.blob_inputs : t == null ? {} : merge(
    try(t.s3_compatible.access_key.secret_access_key, "") != "" ? { "blob-${uc}-secret-access-key" = t.s3_compatible.access_key.secret_access_key } : {},
    try(t.s3.access_key.secret_access_key, "") != "" ? { "blob-${uc}-secret-access-key" = t.s3.access_key.secret_access_key } : {},
    try(t.r2.credentials.secret_access_key, "") != "" ? { "blob-${uc}-secret-access-key" = t.r2.credentials.secret_access_key } : {},
    try(t.azure_blob.account_key, "") != "" ? { "blob-${uc}-azure-account-key" = t.azure_blob.account_key } : {},
  )]...)

  # Every GCS target of one Langfuse shares one credential: the first one
  # declared.
  gcs_credentials = try([for uc in ["event", "media", "batch"] : local.blob_inputs[uc].gcs.credentials_json if try(local.blob_inputs[uc].gcs.credentials_json, "") != ""][0], "")

  # ---- declared secrets and the credentials to generate ----------------------
  declared_secrets = { for k, v in merge(
    {
      "salt"                         = try(var.spec.secrets.salt, "")
      "encryption-key"               = try(var.spec.secrets.encryption_key, "")
      "nextauth-secret"              = try(var.spec.secrets.nextauth_secret, "")
      "postgres-direct-url"          = try(var.spec.postgres.direct_url, "")
      "postgres-shadow-database-url" = try(var.spec.postgres.shadow_database_url, "")
      "redis-tls-key-passphrase"     = try(var.spec.redis.tls.client_key_passphrase, "")
      "smtp-connection-url"          = try(var.spec.smtp.connection_url, "")
      "ai-api-key"                   = try(var.spec.ai_features.api_key, "")
      "license-key"                  = try(var.spec.enterprise.license_key, "")
      "admin-api-key"                = try(var.spec.enterprise.admin_api_key, "")
      "slack-client-secret"          = try(var.spec.slack.client_secret, "")
      "slack-state-secret"           = try(var.spec.slack.state_secret, "")
      "gcs-credentials"              = local.gcs_credentials
      "init-project-public-key"      = local.init_enabled ? try(var.spec.init.project.public_key, "") : ""
      "init-project-secret-key"      = local.init_enabled ? try(var.spec.init.project.secret_key, "") : ""
      "init-user-password"           = local.admin_seeded ? try(var.spec.init.admin.password, "") : ""
    },
    { for p in local.sso_providers : "sso-${p.id}-client-secret" => p.client.client_secret },
    local.blob_declared,
  ) : k => v if v != null && v != "" }

  generate = {
    salt               = try(var.spec.secrets.salt, "") == ""
    encryption_key     = try(var.spec.secrets.encryption_key, "") == ""
    nextauth_secret    = try(var.spec.secrets.nextauth_secret, "") == ""
    slack_state_secret = try(var.spec.slack, null) != null && try(var.spec.slack.state_secret, "") == ""
    admin_password     = local.admin_seeded && try(var.spec.init.admin.password, "") == ""
    project_keys       = local.init_enabled && !local.declared_keys
  }

  # sha256 of the declared values, sorted-key JSON (twin: checksumOf) - the
  # pod annotation that rolls both workloads when a declared value changes.
  secret_checksum = sha256(jsonencode(local.declared_secrets))

  # ---- files projected into both workloads ----------------------------------
  # Order (both engines): Redis TLS CA, client certificate, client key, the
  # trusted CA bundle, then spec.extra_volumes. Each file is a directory
  # mount (never subPath), so a rotated certificate reaches the pod.
  redis_tls      = try(var.spec.redis.tls, null)
  redis_ca_key   = coalesce(try(local.redis_tls.ca.key, null), "ca.crt")
  redis_cert_key = try(coalesce(local.redis_tls.client_certificate.key), "tls.crt")
  redis_key_key  = try(coalesce(local.redis_tls.client_key.key), "tls.key")
  trusted_ca     = try(var.spec.network.trusted_ca, null)
  trusted_ca_key = coalesce(try(local.trusted_ca.key, null), "ca.crt")

  redis_tls_ca_path   = try(local.redis_tls.ca, null) != null ? "/etc/langfuse/redis-tls/ca/${local.redis_ca_key}" : ""
  redis_tls_cert_path = try(local.redis_tls.client_certificate, null) != null ? "/etc/langfuse/redis-tls/client-cert/${local.redis_cert_key}" : ""
  redis_tls_key_path  = try(local.redis_tls.client_key, null) != null ? "/etc/langfuse/redis-tls/client-key/${local.redis_key_key}" : ""
  trusted_ca_path     = local.trusted_ca != null ? "/etc/langfuse/trusted-ca/${local.trusted_ca_key}" : ""

  file_volumes = concat(
    try(local.redis_tls.ca, null) != null ? [try(coalesce(local.redis_tls.ca.secret_name), "") != "" ? {
      name   = "redis-tls-ca"
      secret = { secretName = local.redis_tls.ca.secret_name, items = [{ key = local.redis_ca_key, path = local.redis_ca_key }] }
      } : {
      name      = "redis-tls-ca"
      configMap = { name = local.redis_tls.ca.config_map_name, items = [{ key = local.redis_ca_key, path = local.redis_ca_key }] }
    }] : [],
    try(local.redis_tls.client_certificate, null) != null ? [{
      name   = "redis-tls-client-cert"
      secret = { secretName = local.redis_tls.client_certificate.secret_name, items = [{ key = local.redis_cert_key, path = local.redis_cert_key }] }
    }] : [],
    try(local.redis_tls.client_key, null) != null ? [{
      name   = "redis-tls-client-key"
      secret = { secretName = local.redis_tls.client_key.secret_name, items = [{ key = local.redis_key_key, path = local.redis_key_key }] }
    }] : [],
    local.trusted_ca != null ? [try(coalesce(local.trusted_ca.secret_name), "") != "" ? {
      name   = "trusted-ca"
      secret = { secretName = local.trusted_ca.secret_name, items = [{ key = local.trusted_ca_key, path = local.trusted_ca_key }] }
      } : {
      name      = "trusted-ca"
      configMap = { name = local.trusted_ca.config_map_name, items = [{ key = local.trusted_ca_key, path = local.trusted_ca_key }] }
    }] : [],
  )
  file_mounts = concat(
    try(local.redis_tls.ca, null) != null ? [{ name = "redis-tls-ca", mountPath = "/etc/langfuse/redis-tls/ca", readOnly = true }] : [],
    try(local.redis_tls.client_certificate, null) != null ? [{ name = "redis-tls-client-cert", mountPath = "/etc/langfuse/redis-tls/client-cert", readOnly = true }] : [],
    try(local.redis_tls.client_key, null) != null ? [{ name = "redis-tls-client-key", mountPath = "/etc/langfuse/redis-tls/client-key", readOnly = true }] : [],
    local.trusted_ca != null ? [{ name = "trusted-ca", mountPath = "/etc/langfuse/trusted-ca", readOnly = true }] : [],
  )

  # spec.extra_volumes: the shared VolumeMount message carries both halves
  # (twin: volumeBody / volumeMountBody).
  extra_volumes = [
    for v in try(var.spec.extra_volumes, []) : {
      for vk, vv in {
        name = v.name
        configMap = try(v.config_map, null) == null ? null : {
          for ck, cv in {
            name        = v.config_map.name
            defaultMode = try(v.config_map.default_mode, 0) != 0 ? v.config_map.default_mode : null
            items = try(v.config_map.key, "") != "" ? [{
              key  = v.config_map.key
              path = try(v.config_map.path, "") != "" ? v.config_map.path : v.config_map.key
            }] : null
          } : ck => cv if cv != null
        }
        secret = try(v.secret, null) == null ? null : {
          for sk, sv in {
            secretName  = v.secret.name
            defaultMode = try(v.secret.default_mode, 0) != 0 ? v.secret.default_mode : null
            items = try(v.secret.key, "") != "" ? [{
              key  = v.secret.key
              path = try(v.secret.path, "") != "" ? v.secret.path : v.secret.key
            }] : null
          } : sk => sv if sv != null
        }
        hostPath = try(v.host_path, null) == null ? null : {
          for hk, hv in {
            path = v.host_path.path
            type = try(v.host_path.type, "") != "" ? v.host_path.type : null
          } : hk => hv if hv != null
        }
        emptyDir = try(v.empty_dir, null) == null ? null : {
          for ek, ev in {
            medium    = try(v.empty_dir.medium, "") != "" ? v.empty_dir.medium : null
            sizeLimit = try(v.empty_dir.size_limit, "") != "" ? v.empty_dir.size_limit : null
          } : ek => ev if ev != null
        }
        persistentVolumeClaim = try(v.pvc, null) == null ? null : {
          for pk, pv in {
            claimName = v.pvc.claim_name
            readOnly  = try(v.pvc.read_only, false) ? true : null
          } : pk => pv if pv != null
        }
        # Kubelet-rotated, audience-bound ServiceAccount tokens ride the
        # "projected" volume type in the Kubernetes API.
        projected = try(v.service_account_token, null) == null ? null : {
          sources = [{
            serviceAccountToken = {
              for tk, tv in {
                audience          = v.service_account_token.audience
                expirationSeconds = try(v.service_account_token.expiration_seconds, 0) > 0 ? v.service_account_token.expiration_seconds : null
                path              = try(v.service_account_token.path, "") != "" ? v.service_account_token.path : "token"
              } : tk => tv if tv != null
            }
          }]
        }
      } : vk => vv if vv != null
    }
  ]
  extra_volume_mounts = [
    for v in try(var.spec.extra_volumes, []) : {
      for mk, mv in {
        name      = v.name
        mountPath = v.mount_path
        readOnly  = try(v.read_only, false) ? true : null
        subPath   = try(v.sub_path, "") != "" ? v.sub_path : null
      } : mk => mv if mv != null
    }
  ]
  volumes       = concat(local.file_volumes, local.extra_volumes)
  volume_mounts = concat(local.file_mounts, local.extra_volume_mounts)
}
