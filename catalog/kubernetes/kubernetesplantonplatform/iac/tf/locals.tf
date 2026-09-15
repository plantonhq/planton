# Computed values for the KubernetesPlantonPlatform module. Every
# resolution here has an exact twin in the Pulumi module's locals.go /
# platform_cr.go — keep them in lockstep.
#
# THE SPEC BODY RENDERS ONLY WHAT THE MANIFEST DECLARED: every block below
# is a null-pruned object (`key = cond ? value : null` inside one literal,
# pruned with `{ for k, v in {...} : k => v if v != null }`), so the
# operator's own defaulting stays authoritative for everything unset — the
# same posture as the `planton` Helm chart's verbatim pass-through.
# Three-state optionals (the default-true toggles, the defaulted scalars)
# render exactly when PRESENT: an explicit `enabled: true` is faithfully
# forwarded even though it matches the CRD default, because presence is
# the user's deliberate statement.
#
# HCL DISCIPLINE: `cond ? {...} : {}` ternaries fail plan-time type
# unification when branches carry different attributes, and merge() of
# primitive-only sibling objects silently unifies them into map(string) —
# the null-prune form preserves every value's type. Optional nested blocks
# are read with try(): HCL's && does NOT short-circuit.

locals {
  # The PlantonPlatform CR identity. The CR name is THIS resource's
  # metadata.name — the prefix of every object the operator creates for
  # the platform.
  api_version   = "planton.ai/v1"
  cr_kind       = "PlantonPlatform"
  platform_name = var.metadata.name

  namespace = var.spec.namespace

  # The operator's deterministic per-platform naming — the consumer
  # handles this module exports. Twin of the Pulumi module's vars.
  gateway_service   = "${local.platform_name}-gateway"
  setup_code_secret = "${local.platform_name}-identity-setup-code"

  gateway_local_port = coalesce(try(var.spec.gateway.local_port, null), 8080)

  port_forward_command = "kubectl port-forward -n ${local.namespace} svc/${local.gateway_service} ${local.gateway_local_port}:80"
  setup_code_command   = "kubectl -n ${local.namespace} get secret ${local.setup_code_secret} -o jsonpath='{.data.setup-code}' | base64 -d"

  # Resource-identity labels stamped on the module-created objects (the
  # namespace and the CR itself).
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesPlantonPlatform"
    },
    var.metadata.id != null && var.metadata.id != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    var.metadata.org != null && var.metadata.org != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # ---- license ---------------------------------------------------------------
  license_body = {
    for k, v in {
      key = try(var.spec.license.key, "") != "" ? var.spec.license.key : null
      secretKeyRef = try(var.spec.license.secret_key_ref, null) == null ? null : {
        name = var.spec.license.secret_key_ref.name
        key  = var.spec.license.secret_key_ref.key
      }
    } : k => v if v != null
  }

  # ---- storage ---------------------------------------------------------------
  storage_body = {
    for k, v in {
      storageClassName = try(var.spec.storage.storage_class_name, "") != "" ? var.spec.storage.storage_class_name : null
      size             = try(var.spec.storage.size, "") != "" ? var.spec.storage.size : null
    } : k => v if v != null
  }

  # ---- database: the backup and recovery object stores ------------------------
  # The spec declares credential VALUES (references already resolved); the
  # operator reads Secret NAMES. This module materializes the Secrets and the
  # CR names them — created before the CR so the database is born archiving.
  # Named after the operator's database ("<platform>-postgres"), so they read
  # as the database's own beside the ObjectStore and settings Secret the
  # operator creates under that name. Twin of the Pulumi module's
  # object_store_secrets.go and its vars.
  postgres_cluster_name      = "${local.platform_name}-postgres"
  backup_creds_secret_name   = "${local.postgres_cluster_name}-backup-creds"
  recovery_creds_secret_name = "${local.postgres_cluster_name}-recovery-creds"
  backup_endpoint_ca_name    = "${local.postgres_cluster_name}-backup-endpoint-ca"
  recovery_endpoint_ca_name  = "${local.postgres_cluster_name}-recovery-endpoint-ca"
  endpoint_ca_secret_key     = "ca.crt"

  postgresql_backup   = try(var.spec.database.postgresql.backup, null)
  postgresql_recovery = try(var.spec.database.postgresql.recover_from, null)

  # One rendering context per declared store, so backup and recovery share
  # every line below and differ only in which Secrets they name.
  object_store_contexts = merge(
    local.postgresql_backup != null ? {
      backup = {
        object_store      = local.postgresql_backup.object_store
        creds_secret_name = local.backup_creds_secret_name
        ca_secret_name    = local.backup_endpoint_ca_name
      }
    } : {},
    local.postgresql_recovery != null ? {
      recovery = {
        object_store      = local.postgresql_recovery.object_store
        creds_secret_name = local.recovery_creds_secret_name
        ca_secret_name    = local.recovery_endpoint_ca_name
      }
    } : {}
  )

  # The credentials Secret's data per store, under the keys the operator's
  # preflight reads for each backend, or null for a keyless posture (no
  # Secret exists; the CR names none; the operator reads the pods' own cloud
  # identity). R2 always has one — R2 has no keyless posture. All values are
  # strings, so the chained ternary unifies safely to map(string).
  object_store_creds_data = {
    for key, ctx in local.object_store_contexts : key => (
      try(ctx.object_store.s3.access_keys, null) != null ? {
        ACCESS_KEY_ID     = ctx.object_store.s3.access_keys.access_key_id
        SECRET_ACCESS_KEY = ctx.object_store.s3.access_keys.secret_access_key
        } : try(ctx.object_store.r2, null) != null ? {
        ACCESS_KEY_ID     = ctx.object_store.r2.credentials.access_key_id
        SECRET_ACCESS_KEY = ctx.object_store.r2.credentials.secret_access_key
        } : try(ctx.object_store.gcs.service_account_key_json, "") != "" ? {
        APPLICATION_CREDENTIALS = ctx.object_store.gcs.service_account_key_json
        } : try(ctx.object_store.azure_blob.connection_string, "") != "" ? {
        AZURE_STORAGE_CONNECTION_STRING = ctx.object_store.azure_blob.connection_string
      } : null
    )
  }

  # The endpoint-CA Secret's data per store: only an S3-compatible endpoint
  # with a private CA has one.
  object_store_ca_data = {
    for key, ctx in local.object_store_contexts : key => (
      try(ctx.object_store.s3.endpoint_ca_pem, "") != "" ? {
        (local.endpoint_ca_secret_key) = ctx.object_store.s3.endpoint_ca_pem
      } : null
    )
  }

  # The CR's objectStore per store: the destination path and exactly one
  # backend arm in the operator's vocabulary. credentialsSecretName renders
  # only when a Secret exists; endpointCASecretRef likewise. The r2 arm
  # passes account and jurisdiction through — composing the S3 endpoint is
  # the operator's job, so there is exactly one host table in the product.
  # An arm with nothing to say (keyless gcs) still renders as an empty
  # object: its presence is what selects the backend.
  object_store_body = {
    for key, ctx in local.object_store_contexts : key => {
      for k, v in {
        destinationPath = ctx.object_store.destination_path
        s3 = try(ctx.object_store.s3, null) == null ? null : {
          for k2, v2 in {
            endpointURL           = try(ctx.object_store.s3.endpoint_url, "") != "" ? ctx.object_store.s3.endpoint_url : null
            region                = try(ctx.object_store.s3.region, "") != "" ? ctx.object_store.s3.region : null
            credentialsSecretName = try(ctx.object_store.s3.access_keys, null) != null ? ctx.creds_secret_name : null
            endpointCASecretRef = local.object_store_ca_data[key] == null ? null : {
              name = ctx.ca_secret_name
              key  = local.endpoint_ca_secret_key
            }
          } : k2 => v2 if v2 != null
        }
        gcs = try(ctx.object_store.gcs, null) == null ? null : {
          for k2, v2 in {
            credentialsSecretName = try(ctx.object_store.gcs.service_account_key_json, "") != "" ? ctx.creds_secret_name : null
          } : k2 => v2 if v2 != null
        }
        azureBlob = try(ctx.object_store.azure_blob, null) == null ? null : {
          for k2, v2 in {
            storageAccount        = ctx.object_store.azure_blob.storage_account
            credentialsSecretName = try(ctx.object_store.azure_blob.connection_string, "") != "" ? ctx.creds_secret_name : null
          } : k2 => v2 if v2 != null
        }
        r2 = try(ctx.object_store.r2, null) == null ? null : {
          for k2, v2 in {
            accountId             = ctx.object_store.r2.account_id
            jurisdiction          = try(ctx.object_store.r2.jurisdiction, "") != "" ? ctx.object_store.r2.jurisdiction : null
            credentialsSecretName = ctx.creds_secret_name
          } : k2 => v2 if v2 != null
        }
      } : k => v if v != null
    }
  }

  postgresql_backup_body = local.postgresql_backup == null ? null : {
    for k, v in {
      objectStore               = lookup(local.object_store_body, "backup", null)
      retentionPolicy           = try(local.postgresql_backup.retention_policy, "") != "" ? local.postgresql_backup.retention_policy : null
      schedule                  = try(local.postgresql_backup.schedule, "") != "" ? local.postgresql_backup.schedule : null
      serviceAccountAnnotations = length(try(local.postgresql_backup.service_account_annotations, {})) > 0 ? local.postgresql_backup.service_account_annotations : null
    } : k => v if v != null
  }
  postgresql_recover_from_body = local.postgresql_recovery == null ? null : {
    for k, v in {
      objectStore = lookup(local.object_store_body, "recovery", null)
      serverName  = local.postgresql_recovery.server_name
      targetTime  = try(local.postgresql_recovery.target_time, "") != "" ? local.postgresql_recovery.target_time : null
    } : k => v if v != null
  }

  # ---- database --------------------------------------------------------------
  postgresql_body = {
    for k, v in {
      replicas         = try(var.spec.database.postgresql.replicas, null)
      storageSize      = try(var.spec.database.postgresql.storage_size, "") != "" ? var.spec.database.postgresql.storage_size : null
      storageClassName = try(var.spec.database.postgresql.storage_class_name, "") != "" ? var.spec.database.postgresql.storage_class_name : null
      backup           = local.postgresql_backup_body
      recoverFrom      = local.postgresql_recover_from_body
    } : k => v if v != null
  }
  redis_body = {
    for k, v in {
      storageSize      = try(var.spec.database.redis.storage_size, "") != "" ? var.spec.database.redis.storage_size : null
      storageClassName = try(var.spec.database.redis.storage_class_name, "") != "" ? var.spec.database.redis.storage_class_name : null
    } : k => v if v != null
  }
  database_body = {
    for k, v in {
      postgresql = length(local.postgresql_body) > 0 ? local.postgresql_body : null
      redis      = length(local.redis_body) > 0 ? local.redis_body : null
    } : k => v if v != null
  }

  # ---- ingress ---------------------------------------------------------------
  ingress_tls_issuer = try(var.spec.ingress.tls.issuer, null) == null ? null : {
    for k, v in {
      name = var.spec.ingress.tls.issuer.name
      kind = try(var.spec.ingress.tls.issuer.kind, "") != "" ? var.spec.ingress.tls.issuer.kind : null
    } : k => v if v != null
  }
  ingress_tls = try(var.spec.ingress.tls, null) == null ? null : {
    for k, v in {
      secretName = try(var.spec.ingress.tls.secret_name, "") != "" ? var.spec.ingress.tls.secret_name : null
      issuer     = local.ingress_tls_issuer
    } : k => v if v != null
  }
  # The Gateway API front door: the fork's other arm (the CRD refuses it
  # beside ingressClassName). Rendered only when the manifest named a
  # Gateway; the operator reads the Gateway's listeners for everything else.
  ingress_gateway_ref = try(var.spec.ingress.gateway_ref, null) == null ? null : {
    for k, v in {
      name        = var.spec.ingress.gateway_ref.name
      namespace   = try(var.spec.ingress.gateway_ref.namespace, "") != "" ? var.spec.ingress.gateway_ref.namespace : null
      sectionName = try(var.spec.ingress.gateway_ref.section_name, "") != "" ? var.spec.ingress.gateway_ref.section_name : null
    } : k => v if v != null
  }
  ingress_body = {
    for k, v in {
      enabled          = try(var.spec.ingress.enabled, false) ? true : null
      hostname         = try(var.spec.ingress.hostname, "") != "" ? var.spec.ingress.hostname : null
      ingressClassName = try(var.spec.ingress.ingress_class_name, "") != "" ? var.spec.ingress.ingress_class_name : null
      gatewayRef       = local.ingress_gateway_ref
      annotations      = length(try(var.spec.ingress.annotations, {})) > 0 ? var.spec.ingress.annotations : null
      tls              = local.ingress_tls
      # Rendered on presence, like every defaulted three-state string: an
      # omitted value is left to the CRD's own default (auto).
      reachability = try(var.spec.ingress.reachability, "") != "" ? var.spec.ingress.reachability : null
    } : k => v if v != null
  }

  # ---- gateway / identity ------------------------------------------------------
  gateway_body = {
    for k, v in {
      localPort = try(var.spec.gateway.local_port, null)
    } : k => v if v != null
  }
  identity_body = {
    for k, v in {
      realm      = try(var.spec.identity.realm, "") != "" ? var.spec.identity.realm : null
      adminEmail = try(var.spec.identity.admin_email, "") != "" ? var.spec.identity.admin_email : null
    } : k => v if v != null
  }

  # ---- bootstrap -------------------------------------------------------------
  bootstrap_org = {
    for k, v in {
      slug = try(var.spec.bootstrap.organization.slug, "") != "" ? var.spec.bootstrap.organization.slug : null
      name = try(var.spec.bootstrap.organization.name, "") != "" ? var.spec.bootstrap.organization.name : null
    } : k => v if v != null
  }
  bootstrap_env = {
    for k, v in {
      slug = try(var.spec.bootstrap.environment.slug, "") != "" ? var.spec.bootstrap.environment.slug : null
      name = try(var.spec.bootstrap.environment.name, "") != "" ? var.spec.bootstrap.environment.name : null
    } : k => v if v != null
  }
  bootstrap_secret_backend = try(var.spec.bootstrap.secret_backend, null) == null ? null : {
    for k, v in {
      type = var.spec.bootstrap.secret_backend.type
      awsSecretsManager = try(var.spec.bootstrap.secret_backend.aws_secrets_manager, null) == null ? null : {
        region    = var.spec.bootstrap.secret_backend.aws_secrets_manager.region
        kmsKeyArn = var.spec.bootstrap.secret_backend.aws_secrets_manager.kms_key_arn
      }
    } : k => v if v != null
  }
  bootstrap_body = {
    for k, v in {
      organization   = length(local.bootstrap_org) > 0 ? local.bootstrap_org : null
      environment    = length(local.bootstrap_env) > 0 ? local.bootstrap_env : null
      admins         = length(try(var.spec.bootstrap.admins, [])) > 0 ? var.spec.bootstrap.admins : null
      iacProvisioner = try(var.spec.bootstrap.iac_provisioner, "") != "" ? var.spec.bootstrap.iac_provisioner : null
      secretBackend  = local.bootstrap_secret_backend
    } : k => v if v != null
  }

  # ---- runner / build / vault ----------------------------------------------------
  runner_body = {
    for k, v in {
      enabled                    = try(var.spec.runner.enabled, null)
      storageSize                = try(var.spec.runner.storage_size, "") != "" ? var.spec.runner.storage_size : null
      storageClassName           = try(var.spec.runner.storage_class_name, "") != "" ? var.spec.runner.storage_class_name : null
      serviceAccountAnnotations  = length(try(var.spec.runner.service_account_annotations, {})) > 0 ? var.spec.runner.service_account_annotations : null
      cloudCredentialsSecretName = try(var.spec.runner.cloud_credentials_secret_name, "") != "" ? var.spec.runner.cloud_credentials_secret_name : null
    } : k => v if v != null
  }
  build_body = {
    for k, v in {
      enabled = try(var.spec.build.enabled, null)
    } : k => v if v != null
  }
  remote_runners_body = {
    for k, v in {
      enabled = try(var.spec.remote_runners.enabled, null)
    } : k => v if v != null
  }

  # ---- email -----------------------------------------------------------------
  # One declaration for both senders. Each nested object is its own local so
  # the parent stays a flat null-prune; the two provider arms render only
  # when declared (the spec's CEL already holds exactly one). Defaulted
  # scalars (port, security, from.name) render on presence only, so an
  # omitted value is left to the CRD's own default.
  email_from = {
    for k, v in {
      address = try(var.spec.email.from.address, "") != "" ? var.spec.email.from.address : null
      name    = try(var.spec.email.from.name, "") != "" ? var.spec.email.from.name : null
    } : k => v if v != null
  }
  email_smtp_oauth2 = try(var.spec.email.smtp.oauth2, null) == null ? null : {
    user     = var.spec.email.smtp.oauth2.user
    tokenUrl = var.spec.email.smtp.oauth2.token_url
    scope    = var.spec.email.smtp.oauth2.scope
    clientId = var.spec.email.smtp.oauth2.client_id
    clientSecretRef = {
      name = var.spec.email.smtp.oauth2.client_secret_ref.name
      key  = var.spec.email.smtp.oauth2.client_secret_ref.key
    }
  }
  email_smtp = try(var.spec.email.smtp, null) == null ? null : {
    for k, v in {
      host                  = var.spec.email.smtp.host
      port                  = try(var.spec.email.smtp.port, null)
      security              = try(var.spec.email.smtp.security, "") != "" ? var.spec.email.smtp.security : null
      credentialsSecretName = try(var.spec.email.smtp.credentials_secret_name, "") != "" ? var.spec.email.smtp.credentials_secret_name : null
      oauth2                = local.email_smtp_oauth2
      caBundleSecretRef = try(var.spec.email.smtp.ca_bundle_secret_ref, null) == null ? null : {
        name = var.spec.email.smtp.ca_bundle_secret_ref.name
        key  = var.spec.email.smtp.ca_bundle_secret_ref.key
      }
    } : k => v if v != null
  }
  email_resend = try(var.spec.email.resend, null) == null ? null : {
    apiKeySecretRef = {
      name = var.spec.email.resend.api_key_secret_ref.name
      key  = var.spec.email.resend.api_key_secret_ref.key
    }
  }
  email_body = {
    for k, v in {
      from    = length(local.email_from) > 0 ? local.email_from : null
      replyTo = try(var.spec.email.reply_to, "") != "" ? var.spec.email.reply_to : null
      smtp    = local.email_smtp
      resend  = local.email_resend
    } : k => v if v != null
  }
  # ---- vault: the seal and its credential ----------------------------------------
  # The seal follows the object-store discipline: the spec declares a
  # credential VALUE, this module materializes it as a Secret the CR names,
  # and a keyless arm names none. The Secret hangs off the operator's name for
  # the vault ("<platform>-openbao") and is keyed by the ENV VAR each seal
  # wrapper reads (the cloud SDKs' standard variables; transit's token
  # variable), so the operator hands every key to the vault's process as the
  # variable of the same name. Twin of the Pulumi module's seal_secret.go.
  openbao_release_name   = "${local.platform_name}-openbao"
  seal_creds_secret_name = "${local.openbao_release_name}-seal-creds"

  seal_aws     = try(var.spec.vault.auto_unseal.aws_kms, null)
  seal_gcp     = try(var.spec.vault.auto_unseal.gcp_kms, null)
  seal_azure   = try(var.spec.vault.auto_unseal.azure_key_vault, null)
  seal_transit = try(var.spec.vault.auto_unseal.transit, null)

  # Credential material only — public identifiers (an access key id, a client
  # id) ride the CR as plain fields. null when the declared arm is keyless, so
  # no Secret exists and the CR names none.
  seal_creds_data = (
    local.seal_aws != null && try(coalesce(local.seal_aws.secret_access_key), "") != "" ? {
      AWS_SECRET_ACCESS_KEY = local.seal_aws.secret_access_key
      } : local.seal_azure != null && try(coalesce(local.seal_azure.client_secret), "") != "" ? {
      AZURE_CLIENT_SECRET = local.seal_azure.client_secret
      } : local.seal_transit != null && try(coalesce(local.seal_transit.token), "") != "" ? {
      VAULT_TOKEN = local.seal_transit.token
    } : null
  )

  # The CR's autoUnseal: exactly one arm in the operator's vocabulary,
  # identifiers through, credentialsSecretName only when a Secret exists.
  # The GCP arm is keyless by construction; its workload identity rides the
  # ServiceAccount annotations below, never this body. mountPath renders on
  # presence only, so the operator's own default stands.
  vault_auto_unseal_body = (
    local.seal_aws != null ? {
      awsKms = {
        for k, v in {
          region                = local.seal_aws.region
          kmsKeyId              = local.seal_aws.kms_key_id
          accessKeyId           = try(coalesce(local.seal_aws.access_key_id), "") != "" ? local.seal_aws.access_key_id : null
          credentialsSecretName = local.seal_creds_data != null ? local.seal_creds_secret_name : null
        } : k => v if v != null
      }
      } : local.seal_gcp != null ? {
      gcpKms = {
        project   = local.seal_gcp.project
        region    = local.seal_gcp.region
        keyRing   = local.seal_gcp.key_ring
        cryptoKey = local.seal_gcp.crypto_key
      }
      } : local.seal_azure != null ? {
      azureKeyVault = {
        for k, v in {
          vaultName             = local.seal_azure.vault_name
          keyName               = local.seal_azure.key_name
          tenantId              = local.seal_azure.tenant_id
          clientId              = try(coalesce(local.seal_azure.client_id), "") != "" ? local.seal_azure.client_id : null
          credentialsSecretName = local.seal_creds_data != null ? local.seal_creds_secret_name : null
        } : k => v if v != null
      }
      } : local.seal_transit != null ? {
      transit = {
        for k, v in {
          address               = local.seal_transit.address
          keyName               = local.seal_transit.key_name
          mountPath             = try(coalesce(local.seal_transit.mount_path), "") != "" ? local.seal_transit.mount_path : null
          credentialsSecretName = local.seal_creds_data != null ? local.seal_creds_secret_name : null
        } : k => v if v != null
      }
    } : null
  )

  # The vault ServiceAccount's annotations, merged the one way both engines
  # agree on: the GCP arm's declared workload identity contributes the GKE
  # annotation (the annotation follows the identity by reference), and every
  # explicit vault.service_account_annotations entry is laid over it, so an
  # explicit value wins on conflict. The operator sees one map.
  vault_sa_annotations = merge(
    local.seal_gcp != null && try(coalesce(local.seal_gcp.workload_identity_service_account), "") != "" ? {
      "iam.gke.io/gcp-service-account" = local.seal_gcp.workload_identity_service_account
    } : {},
    try(var.spec.vault.service_account_annotations, null) != null ? var.spec.vault.service_account_annotations : {}
  )

  vault_body = {
    for k, v in {
      enabled                   = try(var.spec.vault.enabled, null)
      autoUnseal                = local.vault_auto_unseal_body
      initSecretName            = try(var.spec.vault.init_secret_name, "") != "" ? var.spec.vault.init_secret_name : null
      serviceAccountAnnotations = length(local.vault_sa_annotations) > 0 ? local.vault_sa_annotations : null
    } : k => v if v != null
  }

  # ---- components ------------------------------------------------------------
  components_graph = {
    for k, v in {
      enabled          = try(var.spec.components.graph.enabled, false) ? true : null
      storageSize      = try(var.spec.components.graph.storage_size, "") != "" ? var.spec.components.graph.storage_size : null
      storageClassName = try(var.spec.components.graph.storage_class_name, "") != "" ? var.spec.components.graph.storage_class_name : null
    } : k => v if v != null
  }
  components_body = {
    for k, v in {
      graph = length(local.components_graph) > 0 ? local.components_graph : null
    } : k => v if v != null
  }

  # ---- prerequisites ---------------------------------------------------------
  prerequisites_body = {
    for k, v in {
      postgresOperator     = try(var.spec.prerequisites.postgres_operator, "") != "" ? var.spec.prerequisites.postgres_operator : null
      tektonPipelines      = try(var.spec.prerequisites.tekton_pipelines, "") != "" ? var.spec.prerequisites.tekton_pipelines : null
      postgresBackupPlugin = try(var.spec.prerequisites.postgres_backup_plugin, "") != "" ? var.spec.prerequisites.postgres_backup_plugin : null
    } : k => v if v != null
  }

  # ---- controlPlane / console ----------------------------------------------------
  control_plane_image = {
    for k, v in {
      repository = try(var.spec.control_plane.image.repository, "") != "" ? var.spec.control_plane.image.repository : null
      tag        = try(var.spec.control_plane.image.tag, "") != "" ? var.spec.control_plane.image.tag : null
    } : k => v if v != null
  }
  control_plane_body = {
    for k, v in {
      image                     = length(local.control_plane_image) > 0 ? local.control_plane_image : null
      replicas                  = try(var.spec.control_plane.replicas, null)
      externalConfigSecretName  = try(var.spec.control_plane.external_config_secret_name, "") != "" ? var.spec.control_plane.external_config_secret_name : null
      serviceAccountAnnotations = length(try(var.spec.control_plane.service_account_annotations, {})) > 0 ? var.spec.control_plane.service_account_annotations : null
    } : k => v if v != null
  }
  console_image = {
    for k, v in {
      repository = try(var.spec.console.image.repository, "") != "" ? var.spec.console.image.repository : null
      tag        = try(var.spec.console.image.tag, "") != "" ? var.spec.console.image.tag : null
    } : k => v if v != null
  }
  console_body = {
    for k, v in {
      image                    = length(local.console_image) > 0 ? local.console_image : null
      replicas                 = try(var.spec.console.replicas, null)
      externalConfigSecretName = try(var.spec.console.external_config_secret_name, "") != "" ? var.spec.console.external_config_secret_name : null
    } : k => v if v != null
  }

  # ---- the CR spec (twin of the Pulumi module's platformSpecBody) -------------
  platform_spec = {
    for k, v in {
      version = var.spec.version

      license       = length(local.license_body) > 0 ? local.license_body : null
      storage       = length(local.storage_body) > 0 ? local.storage_body : null
      database      = length(local.database_body) > 0 ? local.database_body : null
      ingress       = length(local.ingress_body) > 0 ? local.ingress_body : null
      gateway       = length(local.gateway_body) > 0 ? local.gateway_body : null
      identity      = length(local.identity_body) > 0 ? local.identity_body : null
      bootstrap     = length(local.bootstrap_body) > 0 ? local.bootstrap_body : null
      runner        = length(local.runner_body) > 0 ? local.runner_body : null
      build         = length(local.build_body) > 0 ? local.build_body : null
      remoteRunners = length(local.remote_runners_body) > 0 ? local.remote_runners_body : null
      email         = length(local.email_body) > 0 ? local.email_body : null
      vault         = length(local.vault_body) > 0 ? local.vault_body : null
      components    = length(local.components_body) > 0 ? local.components_body : null
      prerequisites = length(local.prerequisites_body) > 0 ? local.prerequisites_body : null
      controlPlane  = length(local.control_plane_body) > 0 ? local.control_plane_body : null
      console       = length(local.console_body) > 0 ? local.console_body : null
    } : k => v if v != null
  }
}
