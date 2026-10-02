# Computed values for the KubernetesOtelCollector module. Every resolution
# here has an exact twin in the Pulumi module's locals.go /
# otel_collector_cr.go — keep them in byte lockstep.
#
# HCL DISCIPLINE: conditional entries are written as
# `key = cond ? value : null` inside ONE object literal, pruned with
# `{ for k, v in {...} : k => v if v != null }` — never `cond ? {...} : {}`
# ternaries (plan-time type unification) and never merge() over
# conditional lists (silent primitive unification). Optional nested blocks
# are read with try() — HCL's && does not short-circuit.

locals {
  api_version = "opentelemetry.io/v1beta1"
  kind        = "OpenTelemetryCollector"

  resource_name = var.metadata.name
  namespace     = var.spec.namespace

  # Resource-identity labels stamped on the CR and the module-created
  # namespace.
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesOtelCollector"
    },
    var.metadata.id != null && var.metadata.id != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    var.metadata.org != null && var.metadata.org != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # The mode the collector actually runs in. The CRD defaults an unset
  # mode to deployment — the resolution here exists only for the
  # mode-dependent rendering rules below, the CR carries mode only when
  # the spec declares one.
  effective_mode = try(var.spec.mode, "") != "" ? var.spec.mode : "deployment"

  # replicas/autoscaler apply only to the workload modes. In
  # daemonset/sidecar modes the (possibly middleware-defaulted) replicas
  # value is IGNORED by design — the spec CEL tolerates the stamped 1 so
  # those manifests stay expressible.
  is_workload_mode = contains(["deployment", "statefulset"], local.effective_mode)

  # THE PIPELINE IS THE PRODUCT: config_yaml carries the collector's own
  # configuration document on its own open contract. The v1beta1 CR's
  # `config` is a STRUCTURED object (not a string — that was v1alpha1),
  # so the document is parsed here and embedded as an object. An
  # unparseable document fails AT PLAN, loudly. Twin: the Pulumi module's
  # yaml.Unmarshal in otel_collector_cr.go.
  collector_config = yamldecode(var.spec.config_yaml)

  # Plain env vars, sorted by name — deterministic CR bodies on both
  # engines (Go maps iterate randomly; HCL maps sort — the sort() makes
  # the contract explicit rather than incidental).
  env_list = [
    for k in sort(keys(try(var.spec.env, {}))) : {
      name  = k
      value = var.spec.env[k]
    }
  ]

  # Secrets loaded whole as environment variables (the credential path —
  # referenced in config as $${env:VAR_NAME}; nothing secret-bearing ever
  # lands in the rendered config document).
  env_from_list = [
    for s in try(var.spec.env_from_secrets, []) : {
      secretRef = { name = s }
    }
  ]

  # ---- volumes (the shared VolumeMount message carries BOTH halves) ---------
  # The CR splits what the spec models as one entry: `volumes` (the pod
  # volume sources) and `volumeMounts` (the container mounts). Exactly one
  # source block per entry (spec-documented); single-key config_map/secret
  # entries project just that key via items.
  cr_volumes = [
    for v in try(var.spec.volumes, []) : {
      for vk, vv in {
        name = v.name
        configMap = try(v.config_map, null) == null ? null : {
          for ck, cv in {
            name        = v.config_map.name
            defaultMode = try(v.config_map.default_mode, null) != null && try(v.config_map.default_mode, 0) != 0 ? v.config_map.default_mode : null
            items = try(v.config_map.key, "") != "" ? [{
              key  = v.config_map.key
              path = try(v.config_map.path, "") != "" ? v.config_map.path : v.config_map.key
            }] : null
          } : ck => cv if cv != null
        }
        secret = try(v.secret, null) == null ? null : {
          for sk, sv in {
            secretName  = v.secret.name
            defaultMode = try(v.secret.default_mode, null) != null && try(v.secret.default_mode, 0) != 0 ? v.secret.default_mode : null
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

  cr_volume_mounts = [
    for v in try(var.spec.volumes, []) : {
      for mk, mv in {
        name      = v.name
        mountPath = v.mount_path
        readOnly  = try(v.read_only, false) ? true : null
        subPath   = try(v.sub_path, "") != "" ? v.sub_path : null
      } : mk => mv if mv != null
    }
  ]

  # ---- extra Service ports ----------------------------------------------------
  # Only for receivers the operator cannot infer — it derives the standard
  # components' ports from the config itself.
  cr_ports = [
    for p in try(var.spec.additional_ports, []) : {
      for pk, pv in {
        name     = p.name
        port     = p.port
        protocol = try(p.protocol, "") != "" && try(p.protocol, "") != null ? p.protocol : null
      } : pk => pv if pv != null
    }
  ]

  # ---- operator-managed HPA ---------------------------------------------------
  autoscaler = try(var.spec.autoscaler, null) == null ? null : {
    for ak, av in {
      minReplicas             = try(var.spec.autoscaler.min_replicas, null)
      maxReplicas             = var.spec.autoscaler.max_replicas
      targetCPUUtilization    = try(var.spec.autoscaler.target_cpu_utilization, null)
      targetMemoryUtilization = try(var.spec.autoscaler.target_memory_utilization, null)
    } : ak => av if av != null
  }

  # ---- container resources ------------------------------------------------------
  resources = try(var.spec.resources, null) == null ? null : {
    for k, v in {
      limits = try(var.spec.resources.limits, null) == null ? null : {
        for lk, lv in {
          cpu    = try(var.spec.resources.limits.cpu, "") != "" ? var.spec.resources.limits.cpu : null
          memory = try(var.spec.resources.limits.memory, "") != "" ? var.spec.resources.limits.memory : null
        } : lk => lv if lv != null
      }
      requests = try(var.spec.resources.requests, null) == null ? null : {
        for rk, rv in {
          cpu    = try(var.spec.resources.requests.cpu, "") != "" ? var.spec.resources.requests.cpu : null
          memory = try(var.spec.resources.requests.memory, "") != "" ? var.spec.resources.requests.memory : null
        } : rk => rv if rv != null
      }
    } : k => v if v != null && length(v) > 0
  }

  # ---- pod security context ----------------------------------------------------
  # The daemonset log-collection pattern typically needs runAsUser 0 —
  # container runtimes write pod log files readable only by root (taught
  # on the spec field). Rendered per key, null-pruned.
  pod_security_context = try(var.spec.pod_security_context, null) == null ? null : {
    for pk, pv in {
      runAsUser           = try(var.spec.pod_security_context.run_as_user, null)
      runAsGroup          = try(var.spec.pod_security_context.run_as_group, null)
      runAsNonRoot        = try(var.spec.pod_security_context.run_as_non_root, null)
      fsGroup             = try(var.spec.pod_security_context.fs_group, null)
      fsGroupChangePolicy = try(var.spec.pod_security_context.fs_group_change_policy, "") != "" ? var.spec.pod_security_context.fs_group_change_policy : null
      supplementalGroups  = length(try(var.spec.pod_security_context.supplemental_groups, [])) > 0 ? var.spec.pod_security_context.supplemental_groups : null
      sysctls = length(try(var.spec.pod_security_context.sysctls, [])) > 0 ? [
        for s in var.spec.pod_security_context.sysctls : { name = s.name, value = s.value }
      ] : null
      seccompProfile = try(var.spec.pod_security_context.seccomp_profile, null) == null ? null : {
        for sk, sv in {
          type             = try(var.spec.pod_security_context.seccomp_profile.type, "") != "" ? var.spec.pod_security_context.seccomp_profile.type : null
          localhostProfile = try(var.spec.pod_security_context.seccomp_profile.localhost_profile, "") != "" ? var.spec.pod_security_context.seccomp_profile.localhost_profile : null
        } : sk => sv if sv != null
      }
    } : pk => pv if pv != null
  }

  # ---- the CR spec body (twin of the Pulumi module's collectorSpecBody) --------
  # Field names are the CRD's own JSON keys (verified against the pinned
  # v1beta1 API types at operator 0.156.0). Every key renders ONLY when
  # the spec declares it, so the operator's defaulting stays authoritative
  # for everything the manifest leaves unsaid.
  collector_spec = {
    for k, v in {
      # The CRD defaults an absent mode to deployment; rendered on
      # declaration (an explicit "deployment" re-states the default
      # harmlessly).
      mode = try(var.spec.mode, "") != "" ? var.spec.mode : null

      config = local.collector_config

      # Workload modes only; never alongside the autoscaler (it manages
      # the count); the middleware-defaulted 1 in daemonset/sidecar modes
      # is deliberately ignored (the spec CEL's expressibility tolerance).
      replicas = local.is_workload_mode && local.autoscaler == null && try(var.spec.replicas, null) != null ? var.spec.replicas : null

      autoscaler = local.is_workload_mode ? local.autoscaler : null

      # Empty = the operator injects its default collector image
      # (fleet-wide override on the operator kind's
      # default_collector_image).
      image = try(var.spec.image, "") != "" ? var.spec.image : null

      # Empty = the operator creates a default ServiceAccount. Set when
      # the pipeline reads cluster state (see the spec's PERMISSIONS
      # note).
      serviceAccount = try(var.spec.service_account, "") != "" ? var.spec.service_account : null

      env     = length(local.env_list) > 0 ? local.env_list : null
      envFrom = length(local.env_from_list) > 0 ? local.env_from_list : null

      volumes      = length(local.cr_volumes) > 0 ? local.cr_volumes : null
      volumeMounts = length(local.cr_volume_mounts) > 0 ? local.cr_volume_mounts : null

      ports = length(local.cr_ports) > 0 ? local.cr_ports : null

      resources = local.resources != null && length(local.resources) > 0 ? local.resources : null

      nodeSelector = length(try(var.spec.scheduling.node_selector, {})) > 0 ? var.spec.scheduling.node_selector : null
      tolerations = length(try(var.spec.scheduling.tolerations, [])) > 0 ? [
        for t in var.spec.scheduling.tolerations : {
          for tk, tv in {
            key               = try(t.key, "") != "" ? t.key : null
            operator          = try(t.operator, "") != "" ? t.operator : null
            value             = try(t.value, "") != "" ? t.value : null
            effect            = try(t.effect, "") != "" ? t.effect : null
            tolerationSeconds = try(t.toleration_seconds, null)
          } : tk => tv if tv != null
        }
      ] : null
      priorityClassName = try(var.spec.scheduling.priority_class_name, "") != "" ? var.spec.scheduling.priority_class_name : null

      podSecurityContext = local.pod_security_context != null && length(local.pod_security_context) > 0 ? local.pod_security_context : null

      # The operator then creates the monitor itself (a ServiceMonitor, or a PodMonitor in sidecar
      # mode); absent, it creates none.
      observability = try(var.spec.service_monitor_enabled, false) ? { metrics = { enableMetrics = true } } : null
    } : k => v if v != null
  }
}
