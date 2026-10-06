# Computed values for the KubernetesGrafana module. Every resolution here
# has an exact twin in the Pulumi module's locals.go / values.go — keep
# them in lockstep.
#
# SECRET DISCIPLINE (load-bearing): the chart renders grafana.ini AND the
# datasource provisioning file into a ConfigMap, and its own
# assertNoLeakedSecrets helper REFUSES to render known secret paths into
# it. Every credential below therefore travels as an environment variable
# sourced from a Secret (envValueFrom / the chart's smtp existingSecret
# wiring), with the provisioning file carrying only a $__env{VAR}
# placeholder that Grafana expands at runtime.
#
# HCL DISCIPLINE (applies to every conditional object in this file):
# conditional entries are written as `key = cond ? value : null` inside ONE
# object literal, pruned with `{ for k, v in {...} : k => v if v != null }`.
# `cond ? {...} : {}` ternaries fail plan-time type unification when
# branches carry different attributes, and merge() over primitive-only
# sibling objects silently UNIFIES them into map(string). Optional nested
# blocks are read with try() (HCL's && does NOT short-circuit); optional
# scalars inside optional blocks with try(coalesce(x), null).

locals {
  # Chart identity — must stay byte-identical with the Pulumi module's
  # vars. KNOW THIS about the repo URL: the grafana chart's canonical
  # home is the grafana-community index — the old
  # https://grafana.github.io/helm-charts stopped serving new versions at
  # chart 10.5.x, and kube-prometheus-stack's own dependency block points
  # at the community repo. Never "fix" this back.
  helm_chart_name = "grafana"
  helm_chart_repo = "https://grafana-community.github.io/helm-charts"

  # Release name — metadata.name, NOT a fixed chart name: several Grafana
  # instances coexist in one Kubernetes cluster. fullnameOverride below
  # pins every chart child name to this.
  release_name = var.metadata.name

  # Chart version resolved to the pinned default when unset, so both
  # engines install the same chart whether or not the platform's
  # defaulting middleware ran. Chart 12.8.0 ships Grafana 13.1.1.
  chart_version = coalesce(var.spec.chart_version, "12.8.0")

  namespace = var.spec.namespace

  # Resource-identity labels stamped on the module-created satellites
  # (the namespace — never injected into the chart's own resources; Helm
  # owns those).
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesGrafana"
    },
    var.metadata.id != null && var.metadata.id != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    var.metadata.org != null && var.metadata.org != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # The chart's Service is grafana.fullname — pinned to the resource name
  # via fullnameOverride. Feeds the endpoint output.
  service_name = local.release_name

  # Name of the Secret carrying the admin credentials: the referenced
  # existing Secret when declared, else the chart-owned `<name>` Secret
  # the chart generates ONCE at first install (stable across upgrades via
  # its lookup).
  admin_secret_name = try(var.spec.admin_secret.name, null) != null ? var.spec.admin_secret.name : local.release_name

  # In-cluster endpoint (the chart serves plain HTTP on the Service; TLS
  # terminates at the composed exposure layer). Feeds the endpoint output
  # and the agent reader's Job.
  endpoint = "http://${local.service_name}.${local.namespace}.svc.cluster.local"

  # ---- image registry/repository split -----------------------------------
  # The spec's repository carries the registry
  # ("my.registry.com/grafana/grafana") but the chart keeps them as
  # SEPARATE values composed {registry}/{repository}:{tag} with registry
  # defaulting to docker.io — mapped verbatim onto image.repository a
  # mirror override would render docker.io/my.registry.com/... and
  # ImagePullBackOff. The first path segment is a registry exactly when
  # it looks like a host (a dot, a port colon, or the literal localhost)
  # — the container runtimes' own rule. Twin: splitImageRepository.
  image_repo_declared = try(coalesce(var.spec.image.repository), null)
  image_repo_parts    = local.image_repo_declared != null ? split("/", local.image_repo_declared) : []
  image_has_registry = (
    length(local.image_repo_parts) > 1 &&
    (can(regex("[.:]", local.image_repo_parts[0])) || local.image_repo_parts[0] == "localhost")
  )
  image_registry   = local.image_has_registry ? local.image_repo_parts[0] : null
  image_repository = local.image_has_registry ? join("/", slice(local.image_repo_parts, 1, length(local.image_repo_parts))) : local.image_repo_declared

  # ---- database (GF_DATABASE_* env config) -------------------------------
  # Env-based configuration keeps the password out of grafana.ini (which
  # lands in a ConfigMap). The password rides envValueFrom — a
  # secretKeyRef the kubelet resolves — so it never appears in rendered
  # values or any config text.
  database_declared = try(var.spec.database, null) != null
  database_env = local.database_declared ? {
    for k, v in {
      GF_DATABASE_TYPE     = var.spec.database.engine == "mysql" ? "mysql" : "postgres"
      GF_DATABASE_HOST     = var.spec.database.host
      GF_DATABASE_NAME     = var.spec.database.name
      GF_DATABASE_USER     = var.spec.database.user
      GF_DATABASE_SSL_MODE = var.spec.database.ssl_mode != "" ? var.spec.database.ssl_mode : null
    } : k => v if v != null
  } : {}

  # ---- datasource credential env vars --------------------------------------
  # One deterministic env var per basic-auth datasource
  # (GF_DS_<NAME>_PASSWORD, non-alphanumerics collapsed to _) — the name
  # appears in the provisioning file AND the envValueFrom key, and both
  # engines derive it identically (twin: datasourcePasswordEnvVar).
  datasource_password_env = {
    for ds in var.spec.datasources : ds.name =>
    "GF_DS_${trim(replace(upper(ds.name), "/[^A-Z0-9]+/", "_"), "_")}_PASSWORD"
    if try(ds.basic_auth, null) != null
  }

  # ---- sign-in (auth.google / auth.generic_oauth) --------------------------
  # Twin of the Pulumi module's signin.go. Non-secret provider settings
  # join grafana.ini; each client secret goes into the module-owned
  # `<name>-sso` Secret (main.tf) and reaches Grafana as
  # GF_AUTH_<PROVIDER>_CLIENT_SECRET, which overrides the ini key, so the
  # secret never enters the ConfigMap the chart renders grafana.ini into.
  sign_in_google  = try(var.spec.auth.google, null)
  sign_in_generic = try(var.spec.auth.generic_oauth, null)
  sign_in_enabled = local.sign_in_google != null || local.sign_in_generic != null
  sso_secret_name = "${local.release_name}-sso"

  sso_secret_data = merge(
    local.sign_in_google != null ? { "google-client-secret" = local.sign_in_google.client_secret } : {},
    local.sign_in_generic != null ? { "generic-oauth-client-secret" = local.sign_in_generic.client_secret } : {}
  )

  sign_in_env_value_from = merge(
    local.sign_in_google != null ? {
      GF_AUTH_GOOGLE_CLIENT_SECRET = {
        secretKeyRef = { name = local.sso_secret_name, key = "google-client-secret" }
      }
    } : {},
    local.sign_in_generic != null ? {
      GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET = {
        secretKeyRef = { name = local.sso_secret_name, key = "generic-oauth-client-secret" }
      }
    } : {}
  )

  sign_in_ini = {
    for k, v in {
      "auth.google" = local.sign_in_google == null ? null : {
        for gk, gv in {
          enabled             = true
          client_id           = local.sign_in_google.client_id
          allow_sign_up       = local.sign_in_google.allow_sign_up == null ? true : local.sign_in_google.allow_sign_up
          allowed_domains     = length(local.sign_in_google.allowed_domains) > 0 ? join(" ", local.sign_in_google.allowed_domains) : null
          hosted_domain       = local.sign_in_google.hosted_domain != "" ? local.sign_in_google.hosted_domain : null
          auto_login          = local.sign_in_google.auto_login ? true : null
          role_attribute_path = local.sign_in_google.role_attribute_path != "" ? local.sign_in_google.role_attribute_path : null
          # Grafana ships [auth.google] with skip_org_role_sync = true, which
          # silently ignores role_attribute_path; a declared mapping applies.
          skip_org_role_sync    = local.sign_in_google.role_attribute_path != "" ? false : null
          role_attribute_strict = local.sign_in_google.role_attribute_strict ? true : null
        } : gk => gv if gv != null
      }
      "auth.generic_oauth" = local.sign_in_generic == null ? null : {
        for ok, ov in {
          enabled               = true
          name                  = try(coalesce(local.sign_in_generic.name), "OAuth")
          client_id             = local.sign_in_generic.client_id
          auth_url              = local.sign_in_generic.auth_url
          token_url             = local.sign_in_generic.token_url
          api_url               = local.sign_in_generic.api_url
          scopes                = join(" ", length(local.sign_in_generic.scopes) > 0 ? local.sign_in_generic.scopes : ["openid", "email", "profile"])
          allow_sign_up         = local.sign_in_generic.allow_sign_up == null ? true : local.sign_in_generic.allow_sign_up
          use_pkce              = local.sign_in_generic.use_pkce == null ? true : local.sign_in_generic.use_pkce
          email_attribute_path  = local.sign_in_generic.email_attribute_path != "" ? local.sign_in_generic.email_attribute_path : null
          login_attribute_path  = local.sign_in_generic.login_attribute_path != "" ? local.sign_in_generic.login_attribute_path : null
          name_attribute_path   = local.sign_in_generic.name_attribute_path != "" ? local.sign_in_generic.name_attribute_path : null
          role_attribute_path   = local.sign_in_generic.role_attribute_path != "" ? local.sign_in_generic.role_attribute_path : null
          skip_org_role_sync    = local.sign_in_generic.role_attribute_path != "" ? false : null
          groups_attribute_path = local.sign_in_generic.groups_attribute_path != "" ? local.sign_in_generic.groups_attribute_path : null
          allowed_groups        = length(local.sign_in_generic.allowed_groups) > 0 ? join(" ", local.sign_in_generic.allowed_groups) : null
          allowed_domains       = length(local.sign_in_generic.allowed_domains) > 0 ? join(" ", local.sign_in_generic.allowed_domains) : null
          role_attribute_strict = local.sign_in_generic.role_attribute_strict ? true : null
          auto_login            = local.sign_in_generic.auto_login ? true : null
        } : ok => ov if ov != null
      }
      # Once the manifest declares sign-in it owns sign-in: settings saved
      # through Administration > Authentication live in Grafana's database
      # and override grafana.ini and the environment. Grafana skips empty
      # values when it layers custom configuration over its defaults, so
      # the list that replaces the default names no provider.
      sso_settings = local.sign_in_enabled ? { configurable_providers = "none" } : null
    } : k => v if v != null
  }

  env_value_from = merge(
    local.sign_in_env_value_from,
    local.database_declared ? {
      GF_DATABASE_PASSWORD = {
        secretKeyRef = {
          name = var.spec.database.password_secret.name
          key  = var.spec.database.password_secret.key
        }
      }
    } : {},
    {
      for ds in var.spec.datasources : local.datasource_password_env[ds.name] => {
        secretKeyRef = {
          name = ds.basic_auth.password_secret.name
          key  = ds.basic_auth.password_secret.key
        }
      } if try(ds.basic_auth, null) != null
    }
  )

  # ---- datasource provisioning entries ---------------------------------------
  datasource_entries = [
    for ds in var.spec.datasources : {
      for k, v in {
        name      = ds.name
        type      = try(coalesce(ds.type), null) != null ? ds.type : "prometheus"
        access    = "proxy"
        url       = ds.url
        isDefault = ds.is_default ? true : null
        uid       = ds.uid != "" ? ds.uid : null
        jsonData  = ds.json_data != "" ? yamldecode(ds.json_data) : null
        # Basic-auth passwords ride $__env{VAR} placeholders expanded by
        # Grafana at runtime — the provisioning file stays
        # credential-free (it lands in the chart's ConfigMap).
        basicAuth     = try(ds.basic_auth, null) != null ? true : null
        basicAuthUser = try(ds.basic_auth.username, null)
        secureJsonData = try(ds.basic_auth, null) != null ? {
          basicAuthPassword = "$__env{${local.datasource_password_env[ds.name]}}"
        } : null
      } : k => v if v != null
    }
  ]

  # ---- community dashboards ------------------------------------------------------
  community_dashboards_declared = length(var.spec.community_dashboards) > 0
  community_dashboard_entries = {
    for cd in var.spec.community_dashboards : "gnet-${cd.gnet_id}" => {
      for k, v in {
        gnetId     = cd.gnet_id
        revision   = cd.revision > 0 ? cd.revision : null
        datasource = cd.datasource
      } : k => v if v != null
    }
  }

  # ---- grafana.ini (non-secret settings ONLY) ---------------------------------------
  grafana_ini = {
    for k, v in {
      server = try(coalesce(var.spec.server.root_url), null) != null ? {
        root_url = var.spec.server.root_url
      } : null
      "auth.anonymous" = try(var.spec.auth.anonymous_enabled, false) ? {
        enabled  = true
        org_role = try(coalesce(var.spec.auth.anonymous_org_role), null) != null ? var.spec.auth.anonymous_org_role : "Viewer"
      } : null
      auth = try(var.spec.auth.disable_login_form, false) ? {
        disable_login_form = true
      } : null
      smtp = try(var.spec.smtp, null) != null ? {
        for sk, sv in {
          enabled      = true
          host         = var.spec.smtp.host
          from_address = var.spec.smtp.from_address != "" ? var.spec.smtp.from_address : null
          from_name    = var.spec.smtp.from_name != "" ? var.spec.smtp.from_name : null
          skip_verify  = var.spec.smtp.skip_verify ? true : null
        } : sk => sv if sv != null
      } : null
    } : k => v if v != null
  }

  # ---- typed chart values (twin of the Pulumi module's buildHelmValues) --
  helm_values = {
    for k, v in {
      # fullnameOverride pins grafana.fullname to the resource name: the
      # Service, the chart-generated admin Secret, the ConfigMap and the
      # ServiceAccount all derive deterministically, and the exported
      # outputs are built from that contract.
      fullnameOverride = local.release_name

      # replicas > 1 is CEL-fenced to require the external database: the
      # embedded SQLite state cannot be shared between pods.
      replicas = try(coalesce(var.spec.replicas), null) != null ? var.spec.replicas : 1

      resources = try(var.spec.resources, null) == null ? null : {
        for rk, rv in {
          requests = try(var.spec.resources.requests, null) == null ? null : {
            for qk, qv in {
              cpu    = var.spec.resources.requests.cpu != "" ? var.spec.resources.requests.cpu : null
              memory = var.spec.resources.requests.memory != "" ? var.spec.resources.requests.memory : null
            } : qk => qv if qv != null
          }
          limits = try(var.spec.resources.limits, null) == null ? null : {
            for lk, lv in {
              cpu    = var.spec.resources.limits.cpu != "" ? var.spec.resources.limits.cpu : null
              memory = var.spec.resources.limits.memory != "" ? var.spec.resources.limits.memory : null
            } : lk => lv if lv != null
          }
        } : rk => rv if rv != null && rv != {}
      }

      # Existing-secret arm: the chart wires the referenced Secret's keys
      # into env at pod start — it must exist BEFORE the apply. Generate
      # arm (admin_secret absent): the chart creates its own `<fullname>`
      # Secret ONCE (lookup-stable across upgrades), so nothing renders
      # here. Credential material never lands in chart values either way.
      admin = try(var.spec.admin_secret, null) != null ? {
        existingSecret = var.spec.admin_secret.name
        userKey        = try(coalesce(var.spec.admin_secret.user_key), null) != null ? var.spec.admin_secret.user_key : "admin-user"
        passwordKey    = try(coalesce(var.spec.admin_secret.password_key), null) != null ? var.spec.admin_secret.password_key : "admin-password"
      } : null

      # Rendered only when declared: the chart default is ephemeral,
      # which the spec documents honestly.
      persistence = try(var.spec.storage, null) != null ? {
        for pk, pv in {
          enabled          = true
          size             = try(coalesce(var.spec.storage.size), null) != null ? var.spec.storage.size : "10Gi"
          storageClassName = var.spec.storage.storage_class != "" ? var.spec.storage.storage_class : null
        } : pk => pv if pv != null
      } : null

      datasources = length(local.datasource_entries) > 0 ? {
        "datasources.yaml" = {
          apiVersion  = 1
          datasources = local.datasource_entries
        }
      } : null

      # The composition contract: any ConfigMap labeled
      # grafana_dashboard: "1" in ANY namespace becomes a dashboard here.
      # Default ON (proto default true).
      sidecar = (try(coalesce(var.spec.dashboard_sidecar_enabled), null) != null ? var.spec.dashboard_sidecar_enabled : true) ? {
        dashboards = {
          enabled         = true
          searchNamespace = "ALL"
        }
      } : null

      # gnetId imports need a file dashboard provider; entries key under
      # the provider's name.
      dashboardProviders = local.community_dashboards_declared ? {
        "dashboardproviders.yaml" = {
          apiVersion = 1
          providers = [
            {
              name            = "default"
              orgId           = 1
              folder          = ""
              type            = "file"
              disableDeletion = false
              options         = { path = "/var/lib/grafana/dashboards/default" }
            }
          ]
        }
      } : null

      dashboards = local.community_dashboards_declared ? {
        default = local.community_dashboard_entries
      } : null

      # Grafana 13 ships its once-core datasource plugins (elasticsearch,
      # cloudwatch) outside the image, and the image's bundled-plugin
      # directory is read-only — shadowBundledPlugins empties it into an
      # emptyDir so listed plugins install cleanly. Rendered together so
      # a plugin list can never hit the read-only-directory failure.
      plugins              = length(var.spec.plugins) > 0 ? var.spec.plugins : null
      shadowBundledPlugins = length(var.spec.plugins) > 0 ? true : null

      "grafana.ini" = length(merge(local.grafana_ini, local.sign_in_ini)) > 0 ? merge(local.grafana_ini, local.sign_in_ini) : null

      # Grafana reads the sign-in Secret only at start: the fingerprint of
      # its data changes the pod template when a secret rotates, so the
      # next apply rolls Grafana onto it. sha256(jsonencode(map)) is what
      # the Pulumi twin's credentialsChecksum computes, byte for byte.
      podAnnotations = local.sign_in_enabled ? {
        "checksum/credentials" = sha256(jsonencode(local.sso_secret_data))
      } : null

      # Credentials ride the chart's own existingSecret wiring — it
      # injects GF_SMTP_USER / GF_SMTP_PASSWORD from the referenced
      # Secret (keys user / password), overriding the ini section at
      # runtime. Nothing secret lands in the ini text.
      smtp = try(coalesce(var.spec.smtp.credentials_secret_name), null) != null ? {
        existingSecret = var.spec.smtp.credentials_secret_name
      } : null

      serviceMonitor = var.spec.service_monitor_enabled ? { enabled = true } : null

      # The declared repository is split into the chart's separate
      # registry/repository values (see the image split locals above) —
      # the chart composes {registry}/{repository}:{tag} with registry
      # defaulting to docker.io. pullSecrets is the private-mirror
      # credential fold — the Pulumi twin renders the same split and
      # list (a silent single-engine ImagePullBackOff class otherwise).
      image = (try(var.spec.image, null) != null && (try(var.spec.image.repository, "") != "" || try(var.spec.image.tag, "") != "" || try(var.spec.image.pull_secret_name, "") != "")) ? {
        for ik, iv in {
          registry    = local.image_registry
          repository  = local.image_repository
          tag         = var.spec.image.tag != "" ? var.spec.image.tag : null
          pullSecrets = var.spec.image.pull_secret_name != "" ? [var.spec.image.pull_secret_name] : null
        } : ik => iv if iv != null
      } : null

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

      priorityClassName = try(var.spec.scheduling.priority_class_name, "") != "" ? var.spec.scheduling.priority_class_name : null

      env          = length(local.database_env) > 0 ? local.database_env : null
      envValueFrom = length(local.env_value_from) > 0 ? local.env_value_from : null
    } : k => v if v != null && v != {}
  }

  # ---- agent teammates' way in (spec.agent_reader) -------------------------
  # Twin of buildAgentReader in the Pulumi module's agentreader.go. Grafana
  # cannot provision service accounts from files, so a Job run after the
  # release is Ready keeps a Viewer account and its one token
  # (agent_reader.tf, scripts.tf).
  agent_reader          = try(var.spec.agent_reader, null)
  agent_reader_declared = local.agent_reader != null

  # `<name>-agent-reader`: the ServiceAccount the Job runs as, its Role and
  # RoleBinding, and the token Secret the Job writes and the ServiceAccount
  # owns.
  agent_reader_name        = "${local.release_name}-agent-reader"
  agent_reader_script_name = "${local.release_name}-agent-reader-script"

  # The account and the generation, defaulted as the proto declares.
  agent_reader_service_account  = try(coalesce(local.agent_reader.service_account_name), "agent-reader")
  agent_reader_token_generation = try(coalesce(local.agent_reader.token_generation), 1)
  agent_reader_disabled         = try(local.agent_reader.disabled, false) == true

  # The Job's image: a POSIX shell with curl, jq and kubectl. A repo or tag
  # left empty in the override keeps the default's, so a mirror never
  # floats to `latest`.
  agent_reader_image            = "${try(coalesce(local.agent_reader.image.repo), "docker.io/alpine/k8s")}:${try(coalesce(local.agent_reader.image.tag), "1.35.8")}"
  agent_reader_pull_secret_name = try(coalesce(local.agent_reader.image.pull_secret_name), "")

  # The admin keys follow the credential arm: the declared Secret's keys
  # (defaulting to the chart's names) or the chart-generated Secret's.
  agent_reader_admin_user_key     = try(coalesce(var.spec.admin_secret.user_key), "admin-user")
  agent_reader_admin_password_key = try(coalesce(var.spec.admin_secret.password_key), "admin-password")

  # `<name>-agent-reader-<8 hex>`: the hex hashes what the Job reconciles --
  # the account, the generation, whether it is disabled, and the script's
  # own text -- so an unchanged declaration is a no-op on every apply and a
  # changed one is a new run. Twin of agentReaderJobName in agentreader.go;
  # the canonical string is identical on both engines.
  agent_reader_job_name = local.agent_reader_declared ? "${local.agent_reader_name}-${substr(sha256(join("|", [
    local.agent_reader_service_account,
    tostring(local.agent_reader_token_generation),
    local.agent_reader_disabled ? "true" : "false",
    sha256(local.agent_reader_script),
  ])), 0, 8)}" : ""

  # The script's plain environment (scripts.go lists the contract), in the
  # Pulumi module's order. SECRET_LABELS is the resource's labels as
  # key-sorted JSON, the labels the script stamps on the token Secret.
  agent_reader_env = [
    ["GRAFANA_URL", local.endpoint],
    ["ADMIN_SECRET", local.admin_secret_name],
    ["NAMESPACE", local.namespace],
    ["RELEASE_NAME", local.release_name],
    ["SERVICE_ACCOUNT", local.agent_reader_service_account],
    ["TOKEN_GENERATION", tostring(local.agent_reader_token_generation)],
    ["DISABLED", local.agent_reader_disabled ? "true" : "false"],
    ["TOKEN_SECRET", local.agent_reader_name],
    ["OWNER_SERVICE_ACCOUNT", local.agent_reader_name],
    ["SECRET_LABELS", jsonencode(local.labels)],
    ["HOME", "/tmp"],
  ]
}
