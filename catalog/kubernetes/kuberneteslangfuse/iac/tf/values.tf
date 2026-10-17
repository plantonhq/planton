# The typed spec rendered into the chart's values (twin: values.go
# buildTypedValues). Unset keys are pruned, never rendered null: a null in a
# Helm values document DELETES the chart's default for that key. The one
# deliberate null is redis.auth.username (see redis_values).
#
# CHART FACTS (verified at chart 2.1.5) the mapping relies on:
#   - the four bundled stores are switched off (`deploy: false`) and every
#     store is wired by host and existingSecret;
#   - each s3.<uploadType> key falls back to the top-level s3.* key through
#     Helm's `default`, which reads false as unset: the top-level region and
#     endpoint are pinned empty and forcePathStyle false, and every upload
#     type renders its own values explicitly;
#   - media uploads have no off switch: the chart always renders the media
#     bucket;
#   - salt, encryptionKey, nextauth.secret and licenseKey take
#     {secretKeyRef}; given one, the chart's `<release>-app` Secret holds no
#     generated value;
#   - additionalEnv lists are concatenated (global first, then the
#     component's own) into each container's env.

locals {
  instance_ref = { for k in ["salt", "encryption-key", "nextauth-secret", "license-key", "gcs-credentials"] :
    k => { secretKeyRef = { name = local.instance_secret_name, key = k } }
  }

  # ---- shared renderers (twins: bodies.go) ------------------------------------
  pod_security_context = { for k, v in {
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

  container_security_context = { for k, v in {
    privileged               = try(var.spec.container_security_context.privileged, false) == true ? true : null
    runAsUser                = try(var.spec.container_security_context.run_as_user, null)
    runAsGroup               = try(var.spec.container_security_context.run_as_group, null)
    runAsNonRoot             = try(var.spec.container_security_context.run_as_non_root, null)
    readOnlyRootFilesystem   = try(var.spec.container_security_context.read_only_root_filesystem, null)
    allowPrivilegeEscalation = try(var.spec.container_security_context.allow_privilege_escalation, null)
    capabilities = (length(try(var.spec.container_security_context.capabilities.add, [])) + length(try(var.spec.container_security_context.capabilities.drop, []))) == 0 ? null : { for ck, cv in {
      add  = length(try(var.spec.container_security_context.capabilities.add, [])) > 0 ? var.spec.container_security_context.capabilities.add : null
      drop = length(try(var.spec.container_security_context.capabilities.drop, [])) > 0 ? var.spec.container_security_context.capabilities.drop : null
    } : ck => cv if cv != null }
    seccompProfile = try(var.spec.container_security_context.seccomp_profile, null) == null ? null : { for sk, sv in {
      type             = try(var.spec.container_security_context.seccomp_profile.type, "") != "" ? var.spec.container_security_context.seccomp_profile.type : null
      localhostProfile = try(var.spec.container_security_context.seccomp_profile.localhost_profile, "") != "" ? var.spec.container_security_context.seccomp_profile.localhost_profile : null
    } : sk => sv if sv != null }
  } : k => v if v != null && v != {} }

  dns_config = try(var.spec.dns_config, null) == null ? {} : { for k, v in {
    nameservers = length(try(var.spec.dns_config.nameservers, [])) > 0 ? var.spec.dns_config.nameservers : null
    searches    = length(try(var.spec.dns_config.searches, [])) > 0 ? var.spec.dns_config.searches : null
    options = length(try(var.spec.dns_config.options, [])) > 0 ? [for o in var.spec.dns_config.options : { for ok, ov in {
      name  = o.name
      value = try(o.value, "") != "" ? o.value : null
    } : ok => ov if ov != null }] : null
  } : k => v if v != null }

  # Workload identity rendered as the ServiceAccount annotation each cloud
  # expects (twin: serviceAccountValues).
  service_account = try(var.spec.service_account, null)
  sa_annotations = merge(
    try(local.service_account.annotations, {}),
    try(local.service_account.workload_identity.gke.service_account_email, "") != "" ? { "iam.gke.io/gcp-service-account" = local.service_account.workload_identity.gke.service_account_email } : {},
    try(local.service_account.workload_identity.eks.role_arn, "") != "" ? { "eks.amazonaws.com/role-arn" = local.service_account.workload_identity.eks.role_arn } : {},
    try(local.service_account.workload_identity.aks.client_id, "") != "" ? { "azure.workload.identity/client-id" = local.service_account.workload_identity.aks.client_id } : {},
    try(local.service_account.workload_identity.aks.tenant_id, "") != "" ? { "azure.workload.identity/tenant-id" = local.service_account.workload_identity.aks.tenant_id } : {},
  )
  automount_token = try(local.service_account.automount_token, null) == null ? true : local.service_account.automount_token
  # One object filtered by null, never a choice between two object shapes
  # (a conditional's branches must share one type).
  existing_service_account = try(local.service_account.existing_name, "")
  service_account_values = { for k, v in {
    automountServiceAccountToken = local.automount_token
    create                       = local.existing_service_account != "" ? false : null
    name                         = local.existing_service_account != "" ? local.existing_service_account : null
    annotations                  = local.existing_service_account == "" && length(local.sa_annotations) > 0 ? local.sa_annotations : null
  } : k => v if v != null }

  # ---- per-component rendering (twin: componentValues) ------------------------
  components = {
    web    = try(var.spec.web, null)
    worker = try(var.spec.worker, null)
  }
  component_values = { for name, c in local.components : name => { for k, v in {
    replicas = try(c.replicas, null) == null ? 1 : c.replicas
    pdb = { for pk, pv in {
      create         = try(c.pdb.enabled, null) == null ? true : c.pdb.enabled
      minAvailable   = try(c.pdb.min_available, "") != "" ? c.pdb.min_available : null
      maxUnavailable = try(c.pdb.max_unavailable, "") != "" ? c.pdb.max_unavailable : null
    } : pk => pv if pv != null }
    image = length({ for ik, iv in {
      repository  = try(c.image.repository, "") != "" ? c.image.repository : null
      tag         = try(c.image.tag, "") != "" ? c.image.tag : null
      pullPolicy  = try(c.image.pull_policy, "") != "" ? c.image.pull_policy : null
      pullSecrets = length(try(c.image.pull_secret_names, [])) > 0 ? [for n in c.image.pull_secret_names : { name = n }] : null
      } : ik => iv if iv != null }) == 0 ? null : { for ik, iv in {
      repository  = try(c.image.repository, "") != "" ? c.image.repository : null
      tag         = try(c.image.tag, "") != "" ? c.image.tag : null
      pullPolicy  = try(c.image.pull_policy, "") != "" ? c.image.pull_policy : null
      pullSecrets = length(try(c.image.pull_secret_names, [])) > 0 ? [for n in c.image.pull_secret_names : { name = n }] : null
    } : ik => iv if iv != null }
    resources = length({ for rk, rv in {
      requests = try(c.resources.requests, null) == null ? null : { for qk, qv in { cpu = try(c.resources.requests.cpu, ""), memory = try(c.resources.requests.memory, "") } : qk => qv if qv != "" }
      limits   = try(c.resources.limits, null) == null ? null : { for qk, qv in { cpu = try(c.resources.limits.cpu, ""), memory = try(c.resources.limits.memory, "") } : qk => qv if qv != "" }
      } : rk => rv if rv != null && rv != {} }) == 0 ? null : { for rk, rv in {
      requests = try(c.resources.requests, null) == null ? null : { for qk, qv in { cpu = try(c.resources.requests.cpu, ""), memory = try(c.resources.requests.memory, "") } : qk => qv if qv != "" }
      limits   = try(c.resources.limits, null) == null ? null : { for qk, qv in { cpu = try(c.resources.limits.cpu, ""), memory = try(c.resources.limits.memory, "") } : qk => qv if qv != "" }
    } : rk => rv if rv != null && rv != {} }
    hpa = try(c.autoscaling.hpa, null) == null ? null : { for hk, hv in {
      enabled                           = true
      minReplicas                       = coalesce(try(c.autoscaling.hpa.min_replicas, null), 1)
      maxReplicas                       = coalesce(try(c.autoscaling.hpa.max_replicas, null), 2)
      targetCPUUtilizationPercentage    = coalesce(try(c.autoscaling.hpa.target_cpu_utilization_percent, null), 50)
      targetMemoryUtilizationPercentage = try(c.autoscaling.hpa.target_memory_utilization_percent, null)
      annotations                       = length(try(c.autoscaling.hpa.annotations, {})) > 0 ? c.autoscaling.hpa.annotations : null
    } : hk => hv if hv != null }
    keda = try(c.autoscaling.keda, null) == null ? null : { for kk, kv in {
      enabled         = true
      minReplicas     = try(c.autoscaling.keda.min_replicas, null) == null ? 1 : c.autoscaling.keda.min_replicas
      maxReplicas     = coalesce(try(c.autoscaling.keda.max_replicas, null), 2)
      pollingInterval = coalesce(try(c.autoscaling.keda.polling_interval_seconds, null), 30)
      triggerType     = try(c.autoscaling.keda.trigger, "") == "memory" ? "memory" : "cpu"
      metricType      = try(c.autoscaling.keda.metric_type, "") == "AverageValue" ? "AverageValue" : "Utilization"
      value           = coalesce(try(c.autoscaling.keda.target_value, null), "50")
      containerName   = try(c.autoscaling.keda.container_name, "") != "" ? c.autoscaling.keda.container_name : null
    } : kk => kv if kv != null }
    vpa = try(c.autoscaling.vpa, null) == null ? null : { for vk, vv in {
      enabled             = true
      updatePolicy        = { updateMode = coalesce(try(c.autoscaling.vpa.update_mode, null), "Auto") }
      controlledResources = length(try(c.autoscaling.vpa.controlled_resources, [])) > 0 ? c.autoscaling.vpa.controlled_resources : null
      maxAllowed          = try(c.autoscaling.vpa.max_allowed, null) == null ? null : (length({ for qk, qv in { cpu = try(c.autoscaling.vpa.max_allowed.cpu, ""), memory = try(c.autoscaling.vpa.max_allowed.memory, "") } : qk => qv if qv != "" }) == 0 ? null : { for qk, qv in { cpu = try(c.autoscaling.vpa.max_allowed.cpu, ""), memory = try(c.autoscaling.vpa.max_allowed.memory, "") } : qk => qv if qv != "" })
      minAllowed          = try(c.autoscaling.vpa.min_allowed, null) == null ? null : (length({ for qk, qv in { cpu = try(c.autoscaling.vpa.min_allowed.cpu, ""), memory = try(c.autoscaling.vpa.min_allowed.memory, "") } : qk => qv if qv != "" }) == 0 ? null : { for qk, qv in { cpu = try(c.autoscaling.vpa.min_allowed.cpu, ""), memory = try(c.autoscaling.vpa.min_allowed.memory, "") } : qk => qv if qv != "" })
    } : vk => vv if vv != null }
    pod = { for pk, pv in {
      # The component's own labels and annotations, the Planton identity
      # labels (they win), and the instance Secret checksum.
      labels       = merge(try(c.workload.pod_labels, {}), local.labels)
      annotations  = merge(try(c.workload.pod_annotations, {}), { "checksum/instance-secret" = local.secret_checksum })
      nodeSelector = length(try(c.scheduling.node_selector, {})) > 0 ? c.scheduling.node_selector : null
      tolerations = length(try(c.scheduling.tolerations, [])) > 0 ? [for t in c.scheduling.tolerations : { for tk, tv in {
        key               = try(t.key, "") != "" ? t.key : null
        operator          = try(t.operator, "") != "" ? t.operator : null
        value             = try(t.value, "") != "" ? t.value : null
        effect            = try(t.effect, "") != "" ? t.effect : null
        tolerationSeconds = try(t.toleration_seconds, null)
      } : tk => tv if tv != null }] : null
      affinity = local.affinity[name]
      topologySpreadConstraints = length(try(c.scheduling.topology_spread_constraints, [])) > 0 ? [for t in c.scheduling.topology_spread_constraints : {
        maxSkew           = try(t.max_skew, 0) == 0 ? 1 : t.max_skew
        topologyKey       = t.topology_key
        whenUnsatisfiable = t.when_unsatisfiable
        labelSelector = { matchLabels = length(try(t.match_labels, {})) > 0 ? t.match_labels : {
          "app.kubernetes.io/name"     = local.helm_chart_name
          "app.kubernetes.io/instance" = local.release_name
          "app"                        = name
        } }
      }] : null
      priorityClassName = try(c.scheduling.priority_class_name, "") != "" ? c.scheduling.priority_class_name : null
      additionalEnv     = length(local.component_env[name]) > 0 ? local.component_env[name] : null
    } : pk => pv if pv != null }
    deployment = length({ for dk, dv in {
      additionalLabels = length(try(c.workload.deployment_labels, {})) > 0 ? c.workload.deployment_labels : null
      annotations      = length(try(c.workload.deployment_annotations, {})) > 0 ? c.workload.deployment_annotations : null
      strategy         = local.strategy[name]
      } : dk => dv if dv != null }) == 0 ? null : { for dk, dv in {
      additionalLabels = length(try(c.workload.deployment_labels, {})) > 0 ? c.workload.deployment_labels : null
      annotations      = length(try(c.workload.deployment_annotations, {})) > 0 ? c.workload.deployment_annotations : null
      strategy         = local.strategy[name]
    } : dk => dv if dv != null }
    revisionHistoryLimit = try(c.workload.revision_history_limit, null)
    hostAliases          = length(try(c.workload.host_aliases, [])) > 0 ? [for a in c.workload.host_aliases : { ip = a.ip, hostnames = a.hostnames }] : null
  } : k => v if v != null } }

  # Keyed by component so each one reads its own list: a conditional that
  # picked between the two lists would need both to share one type, and
  # their entries differ.
  component_env = { web = local.web_env, worker = local.worker_env }

  affinity = { for name, c in local.components : name => (
    length(local.affinity_parts[name]) == 0 ? null : local.affinity_parts[name]
  ) }
  affinity_parts = { for name, c in local.components : name => { for k, v in {
    nodeAffinity = (length(try(c.scheduling.node_affinity.required, [])) + length(try(c.scheduling.node_affinity.preferred, []))) == 0 ? null : { for nk, nv in {
      requiredDuringSchedulingIgnoredDuringExecution = length(try(c.scheduling.node_affinity.required, [])) > 0 ? {
        nodeSelectorTerms = [for t in c.scheduling.node_affinity.required : {
          matchExpressions = [for e in try(t.match_expressions, []) : { for ek, ev in {
            key      = e.key
            operator = e.operator
            values   = length(try(e.values, [])) > 0 ? e.values : null
          } : ek => ev if ev != null }]
        }]
      } : null
      preferredDuringSchedulingIgnoredDuringExecution = length(try(c.scheduling.node_affinity.preferred, [])) > 0 ? [for p in c.scheduling.node_affinity.preferred : {
        weight = p.weight
        preference = {
          matchExpressions = [for e in try(p.term.match_expressions, []) : { for ek, ev in {
            key      = e.key
            operator = e.operator
            values   = length(try(e.values, [])) > 0 ? e.values : null
          } : ek => ev if ev != null }]
        }
      }] : null
    } : nk => nv if nv != null }
    podAffinity     = local.pod_affinity_terms[name]["pod_affinity"]
    podAntiAffinity = local.pod_affinity_terms[name]["pod_anti_affinity"]
  } : k => v if v != null } }
  pod_affinity_terms = { for name, c in local.components : name => { for kind in ["pod_affinity", "pod_anti_affinity"] : kind => (
    (length(try(c.scheduling[kind].required, [])) + length(try(c.scheduling[kind].preferred, []))) == 0 ? null : { for k, v in {
      requiredDuringSchedulingIgnoredDuringExecution = length(try(c.scheduling[kind].required, [])) > 0 ? [for t in c.scheduling[kind].required : { for tk, tv in {
        labelSelector = { matchLabels = try(t.match_labels, {}) }
        topologyKey   = t.topology_key
        namespaces    = length(try(t.namespaces, [])) > 0 ? t.namespaces : null
      } : tk => tv if tv != null }] : null
      preferredDuringSchedulingIgnoredDuringExecution = length(try(c.scheduling[kind].preferred, [])) > 0 ? [for w in c.scheduling[kind].preferred : {
        weight = w.weight
        podAffinityTerm = { for tk, tv in {
          labelSelector = { matchLabels = try(w.term.match_labels, {}) }
          topologyKey   = w.term.topology_key
          namespaces    = length(try(w.term.namespaces, [])) > 0 ? w.term.namespaces : null
        } : tk => tv if tv != null }
      }] : null
    } : k => v if v != null }
  ) } }

  # Rollout strategy; an integer-looking bound renders as a number (a string
  # "1" would read as an invalid percentage) - twin: strategyBody.
  # An integer-looking bound renders as a number (Kubernetes reads a string
  # only as a percentage). try() keeps the number: a conditional would unify
  # both branches to string.
  strategy = { for name, c in local.components : name => (
    (try(c.workload.strategy.type, "") == "" && try(c.workload.strategy.max_surge, "") == "" && try(c.workload.strategy.max_unavailable, "") == "") ? null : { for k, v in {
      type = try(c.workload.strategy.type, "") != "" ? c.workload.strategy.type : null
      rollingUpdate = (try(c.workload.strategy.max_surge, "") == "" && try(c.workload.strategy.max_unavailable, "") == "") ? null : { for rk, rv in {
        maxSurge       = try(c.workload.strategy.max_surge, "") == "" ? null : try(tonumber(c.workload.strategy.max_surge), c.workload.strategy.max_surge)
        maxUnavailable = try(c.workload.strategy.max_unavailable, "") == "" ? null : try(tonumber(c.workload.strategy.max_unavailable), c.workload.strategy.max_unavailable)
      } : rk => rv if rv != null }
    } : k => v if v != null }
  ) }

  probes = {
    web_liveness  = { for k, v in local.probe_fields : k => try(var.spec.web.liveness_probe[v], null) if try(var.spec.web.liveness_probe[v], null) != null }
    web_readiness = { for k, v in local.probe_fields : k => try(var.spec.web.readiness_probe[v], null) if try(var.spec.web.readiness_probe[v], null) != null }
    worker_liveness = merge(
      { for k, v in local.probe_fields : k => try(var.spec.worker.liveness_probe[v], null) if try(var.spec.worker.liveness_probe[v], null) != null },
      try(var.spec.worker.fail_liveness_if_queue_consumption_stuck, false) ? { path = "/api/health?failIfQueueConsumptionStuck=true" } : {},
    )
  }
  probe_fields = {
    initialDelaySeconds = "initial_delay_seconds"
    periodSeconds       = "period_seconds"
    timeoutSeconds      = "timeout_seconds"
    failureThreshold    = "failure_threshold"
    successThreshold    = "success_threshold"
  }

  web_values = merge(
    local.component_values.web,
    length(local.probes.web_liveness) > 0 ? { livenessProbe = local.probes.web_liveness } : {},
    length(local.probes.web_readiness) > 0 ? { readinessProbe = local.probes.web_readiness } : {},
    {
      service = { for k, v in {
        type             = coalesce(try(var.spec.web.service.type, null), "ClusterIP")
        port             = local.web_port
        externalPort     = try(var.spec.web.service.external_port, null)
        nodePort         = try(var.spec.web.service.node_port, null)
        annotations      = length(try(var.spec.web.service.annotations, {})) > 0 ? var.spec.web.service.annotations : null
        additionalLabels = length(try(var.spec.web.service.labels, {})) > 0 ? var.spec.web.service.labels : null
      } : k => v if v != null }
    },
  )
  worker_values = merge(
    local.component_values.worker,
    length(local.probes.worker_liveness) > 0 ? { livenessProbe = local.probes.worker_liveness } : {},
  )

  # ---- the langfuse block -------------------------------------------------------
  log_level  = endswith(try(var.spec.observability.log_level, ""), "unspecified") ? "" : try(var.spec.observability.log_level, "")
  log_format = endswith(try(var.spec.observability.log_format, ""), "unspecified") ? "" : try(var.spec.observability.log_format, "")
  langfuse_values = { for k, v in {
    salt          = local.instance_ref["salt"]
    encryptionKey = local.instance_ref["encryption-key"]
    nextauth = {
      url    = local.external_url
      secret = local.instance_ref["nextauth-secret"]
    }
    features = {
      telemetryEnabled = try(var.spec.features.telemetry_enabled, null) == null ? true : var.spec.features.telemetry_enabled
      signUpDisabled   = try(var.spec.auth.disable_sign_up, false)
    }
    serviceAccount              = local.service_account_values
    web                         = local.web_values
    worker                      = local.worker_values
    logging                     = (local.log_level == "" && local.log_format == "") ? null : { for lk, lv in { level = local.log_level, format = local.log_format } : lk => lv if lv != "" }
    licenseKey                  = contains(keys(local.declared_secrets), "license-key") ? local.instance_ref["license-key"] : null
    allowedOrganizationCreators = length(try(var.spec.features.allowed_organization_creators, [])) > 0 ? var.spec.features.allowed_organization_creators : null
    auth                        = try(var.spec.auth.disable_password_sign_in, false) ? { disableUsernamePassword = true } : null
    podSecurityContext          = length(local.pod_security_context) > 0 ? local.pod_security_context : null
    securityContext             = length(local.container_security_context) > 0 ? local.container_security_context : null
    dnsConfig                   = length(local.dns_config) > 0 ? local.dns_config : null
    extraVolumes                = length(local.volumes) > 0 ? local.volumes : null
    extraVolumeMounts           = length(local.volumes) > 0 ? local.volume_mounts : null
    additionalEnv               = length(local.common_env) > 0 ? local.common_env : null
  } : k => v if v != null }

  # ---- the stores -------------------------------------------------------------
  postgres_values = {
    deploy = false
    host   = var.spec.postgres.host
    port   = coalesce(try(var.spec.postgres.port, null), 5432)
    auth = {
      username       = var.spec.postgres.username
      existingSecret = var.spec.postgres.password_secret.secret_name
      secretKeys     = { userPasswordKey = coalesce(try(var.spec.postgres.password_secret.secret_key, null), "password") }
      database       = coalesce(try(var.spec.postgres.database, null), "langfuse")
    }
    migration = { autoMigrate = try(var.spec.postgres.auto_migrate, null) == null ? true : var.spec.postgres.auto_migrate }
  }
  postgres_values_with_args = merge(local.postgres_values, try(var.spec.postgres.args, "") != "" ? { args = var.spec.postgres.args } : {})

  # The chart selects TLS for queries from an https:// prefix on the host and
  # strips it again for the migration URL.
  clickhouse_values = {
    deploy     = false
    host       = try(var.spec.clickhouse.secure, false) ? "https://${var.spec.clickhouse.host}" : var.spec.clickhouse.host
    httpPort   = coalesce(try(var.spec.clickhouse.http_port, null), 8123)
    nativePort = coalesce(try(var.spec.clickhouse.native_port, null), 9000)
    database   = coalesce(try(var.spec.clickhouse.database, null), "default")
    auth = {
      username          = var.spec.clickhouse.username
      existingSecret    = var.spec.clickhouse.password_secret.secret_name
      existingSecretKey = coalesce(try(var.spec.clickhouse.password_secret.secret_key, ""), var.spec.clickhouse.username)
    }
    migration = {
      ssl         = try(var.spec.clickhouse.secure, false)
      autoMigrate = try(var.spec.clickhouse.auto_migrate, null) == null ? true : var.spec.clickhouse.auto_migrate
    }
    cluster = { enabled = try(var.spec.clickhouse.cluster.enabled, null) == null ? true : var.spec.clickhouse.cluster.enabled }
  }

  redis_values = merge(
    {
      deploy = false
      auth = {
        # The chart builds the connection string with `<username>:` only
        # when a username is set; null removes the chart's "default".
        username                  = try(var.spec.redis.username, "") != "" ? var.spec.redis.username : null
        existingSecret            = try(var.spec.redis.password_secret.secret_name, "")
        existingSecretPasswordKey = try(var.spec.redis.password_secret.secret_key, "")
        database                  = try(var.spec.redis.standalone.database, 0)
      }
      tls = { for k, v in {
        enabled  = try(local.redis_tls.enabled, false)
        caPath   = local.redis_tls_ca_path != "" ? local.redis_tls_ca_path : null
        certPath = local.redis_tls_cert_path != "" ? local.redis_tls_cert_path : null
        keyPath  = local.redis_tls_key_path != "" ? local.redis_tls_key_path : null
      } : k => v if v != null }
      cluster = {
        enabled = try(var.spec.redis.cluster, null) != null
        nodes   = try(var.spec.redis.cluster.nodes, [])
      }
      sentinel = try(var.spec.redis.sentinel, null) == null ? { enabled = false } : { for k, v in {
        enabled                   = true
        masterName                = var.spec.redis.sentinel.master_name
        nodes                     = join(",", var.spec.redis.sentinel.nodes)
        username                  = try(var.spec.redis.sentinel.username, "") != "" ? var.spec.redis.sentinel.username : null
        existingSecret            = try(var.spec.redis.sentinel.password_secret, null) != null ? var.spec.redis.sentinel.password_secret.name : null
        existingSecretPasswordKey = try(var.spec.redis.sentinel.password_secret, null) != null ? var.spec.redis.sentinel.password_secret.key : null
      } : k => v if v != null }
    },
    try(var.spec.redis.standalone, null) != null ? {
      host = var.spec.redis.standalone.host
      port = coalesce(try(var.spec.redis.standalone.port, null), 6379)
    } : {},
  )

  # One upload type's values block (twin: blobTarget.values).
  blob_upload_values = { for uc, pair in {
    eventUpload = { t = local.event_target, prefix = coalesce(try(var.spec.blob_storage.event_upload.prefix, null), "events/") }
    mediaUpload = { t = local.media_target, prefix = coalesce(try(var.spec.blob_storage.media_upload.prefix, null), "media/") }
    batchExport = { t = local.batch_target, prefix = coalesce(try(var.spec.blob_storage.batch_export.prefix, null), "exports/") }
    } : uc => { for k, v in {
      bucket          = pair.t.bucket
      prefix          = pair.prefix
      forcePathStyle  = pair.t.force_path_style
      region          = pair.t.region != "" ? pair.t.region : null
      endpoint        = pair.t.endpoint != "" ? pair.t.endpoint : null
      accessKeyId     = pair.t.access_key_id == null ? null : { for ak, av in pair.t.access_key_id : ak => av if av != null }
      secretAccessKey = pair.t.secret_access_key == null ? null : { for sk, sv in pair.t.secret_access_key : sk => sv if sv != null }
  } : k => v if v != null } }

  blob_values = merge(
    {
      deploy          = false
      storageProvider = local.event_target.family
      bucket          = local.event_target.bucket
      region          = ""
      endpoint        = ""
      forcePathStyle  = false
      concurrency = {
        reads  = coalesce(try(var.spec.blob_storage.concurrent_reads, null), 50)
        writes = coalesce(try(var.spec.blob_storage.concurrent_writes, null), 50)
      }
      eventUpload = local.blob_upload_values.eventUpload
      # Browsers and SDKs get presigned URLs for the public endpoint; Langfuse
      # reads media through the target's own endpoint
      # (LANGFUSE_S3_MEDIA_UPLOAD_INTERNAL_ENDPOINT in the environment).
      mediaUpload = merge(
        local.blob_upload_values.mediaUpload,
        try(var.spec.blob_storage.media_upload.public_endpoint, "") != "" ? { endpoint = var.spec.blob_storage.media_upload.public_endpoint } : {},
        {
          maxContentLength         = coalesce(try(var.spec.blob_storage.media_upload.max_content_length_bytes, null), 1000000000)
          downloadUrlExpirySeconds = coalesce(try(var.spec.blob_storage.media_upload.download_url_expiry_seconds, null), 3600)
        },
      )
      batchExport = merge(
        local.blob_upload_values.batchExport,
        { enabled = try(var.spec.blob_storage.batch_export.enabled, null) == null ? true : var.spec.blob_storage.batch_export.enabled },
      )
    },
    contains(keys(local.declared_secrets), "gcs-credentials") ? { gcs = { credentials = local.instance_ref["gcs-credentials"] } } : {},
  )

  typed_helm_values = {
    langfuse   = local.langfuse_values
    postgresql = local.postgres_values_with_args
    redis      = local.redis_values
    clickhouse = local.clickhouse_values
    s3         = local.blob_values
  }

  # Merged LAST (twin: rePinnedValues): the fullname every output derives
  # from, and the four bundled stores held off, so an override can never
  # deploy a private database beside the composed ones.
  re_pinned_values = {
    fullnameOverride = local.release_name
    postgresql       = { deploy = false }
    redis            = { deploy = false }
    clickhouse       = { deploy = false }
    s3               = { deploy = false }
  }
}
