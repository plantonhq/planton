# Computed values for the KubernetesKubePrometheusStack module. Every
# resolution here has an exact twin in the Pulumi module's locals.go /
# values.go — keep them in lockstep.
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
  # vars: cross-engine chart-name drift deploys two different products
  # from one manifest.
  helm_chart_name = "kube-prometheus-stack"
  helm_chart_repo = "https://prometheus-community.github.io/helm-charts"

  # Release name — metadata.name. fullnameOverride below pins every chart
  # child name to it: `<name>-prometheus`, `<name>-alertmanager`,
  # `<name>-grafana`, `<name>-operator`, the exporter subcharts. The
  # 26-character fullname budget is enforced by the helm_release
  # precondition in main.tf — the chart would otherwise TRUNCATE silently.
  release_name = var.metadata.name

  # Chart version resolved to the pinned default when unset, so both
  # engines install the same chart whether or not the platform's
  # defaulting middleware ran. Chart 91.8.2 = Prometheus Operator
  # v0.94.1.
  chart_version = coalesce(var.spec.chart_version, "91.8.2")

  namespace = var.spec.namespace

  # Resource-identity labels stamped on the module-created satellites
  # (the namespace and the remote-write auth Secret — never injected into
  # the chart's own resources; Helm owns those).
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesKubePrometheusStack"
    },
    var.metadata.id != null && var.metadata.id != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    var.metadata.org != null && var.metadata.org != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # Whether the alertmanager / bundled-grafana halves deploy (proto
  # optional-bool defaults resolve to true).
  alertmanager_enabled = try(coalesce(var.spec.alertmanager.enabled), null) != null ? var.spec.alertmanager.enabled : true
  grafana_enabled      = try(coalesce(var.spec.grafana.enabled), null) != null ? var.spec.grafana.enabled : true

  # Child service names derived from the pinned fullname.
  prometheus_service   = "${local.release_name}-prometheus"
  alertmanager_service = local.alertmanager_enabled ? "${local.release_name}-alertmanager" : ""
  grafana_service      = local.grafana_enabled ? "${local.release_name}-grafana" : ""

  # Name of the Secret holding the bundled Grafana's admin credentials:
  # the referenced existing Secret when declared, else the grafana
  # subchart's own `<name>-grafana` Secret (generated once, lookup-stable
  # across upgrades). Empty when grafana is disabled.
  grafana_admin_secret_name = !local.grafana_enabled ? "" : (
    try(var.spec.grafana.admin_secret.name, null) != null ? var.spec.grafana.admin_secret.name : local.grafana_service
  )

  # ---- remote write ---------------------------------------------------------
  remote_write_entries = try(var.spec.prometheus.remote_write, [])

  # The Prometheus CRD reads BOTH basic-auth halves from Secrets; declared
  # plain-string usernames ride the module-owned `<name>-remote-write-auth`
  # Secret (key `username-<i>` — twin: remoteWriteUsernameKey), passwords
  # stay in the user's Secrets. Materialized only when any entry declares
  # basic auth.
  remote_write_username_data = {
    for i, rw in local.remote_write_entries :
    "username-${i}" => rw.basic_auth.username if try(rw.basic_auth, null) != null
  }
  remote_write_auth_secret_name = length(local.remote_write_username_data) > 0 ? "${local.release_name}-remote-write-auth" : ""

  remote_write_values = [
    for i, rw in local.remote_write_entries : {
      for k, v in {
        url  = rw.url
        name = rw.name != "" ? rw.name : null
        basicAuth = try(rw.basic_auth, null) != null ? {
          username = {
            name = local.remote_write_auth_secret_name
            key  = "username-${i}"
          }
          password = {
            name = rw.basic_auth.password_secret.name
            key  = rw.basic_auth.password_secret.key
          }
        } : null
        authorization = try(rw.bearer_token_secret, null) != null ? {
          type = "Bearer"
          credentials = {
            name = rw.bearer_token_secret.name
            key  = rw.bearer_token_secret.key
          }
        } : null
        sigv4 = try(rw.sigv4, null) != null ? {
          for sk, sv in {
            region  = rw.sigv4.region
            roleArn = rw.sigv4.role_arn != "" ? rw.sigv4.role_arn : null
            accessKey = try(rw.sigv4.access_key_secret, null) != null ? {
              name = rw.sigv4.access_key_secret.name
              key  = rw.sigv4.access_key_secret.key
            } : null
            secretKey = try(rw.sigv4.secret_key_secret, null) != null ? {
              name = rw.sigv4.secret_key_secret.name
              key  = rw.sigv4.secret_key_secret.key
            } : null
          } : sk => sv if sv != null
        } : null
        azureAd = try(rw.azure_ad, null) != null ? {
          for ak, av in {
            managedIdentity = { clientId = rw.azure_ad.managed_identity_client_id }
            cloud           = rw.azure_ad.cloud != "" ? rw.azure_ad.cloud : null
          } : ak => av if av != null
        } : null
      } : k => v if v != null
    }
  ]

  # ---- prometheus spec ----------------------------------------------------------
  # Discovery: the component default is cluster-wide (`all_monitors`) —
  # every catalog kind's service_monitor toggle and any user-authored
  # monitor lights up without extra wiring. `release_managed_only`
  # restores the chart's release-fenced default by rendering nothing.
  discovery_all = try(coalesce(var.spec.prometheus.discovery), "all_monitors") != "release_managed_only"

  prometheus_ephemeral = try(var.spec.prometheus.ephemeral, false)

  prometheus_spec = {
    for k, v in {
      replicas      = try(coalesce(var.spec.prometheus.replicas), null) != null ? var.spec.prometheus.replicas : 1
      retention     = try(coalesce(var.spec.prometheus.retention), null)
      retentionSize = try(var.spec.prometheus.retention_size, "") != "" ? var.spec.prometheus.retention_size : null

      # Storage: a PVC per replica BY DEFAULT — the chart's own default
      # is an emptyDir that loses every metric on restart, which the
      # spec deliberately inverts. The ephemeral arm restores the chart
      # default for throwaway clusters.
      storageSpec = local.prometheus_ephemeral ? null : {
        volumeClaimTemplate = {
          spec = {
            for pk, pv in {
              accessModes      = ["ReadWriteOnce"]
              storageClassName = try(var.spec.prometheus.storage_class, "") != "" ? var.spec.prometheus.storage_class : null
              resources = {
                requests = {
                  storage = try(coalesce(var.spec.prometheus.disk_size), null) != null ? var.spec.prometheus.disk_size : "50Gi"
                }
              }
            } : pk => pv if pv != null
          }
        }
      }

      resources = try(var.spec.prometheus.resources, null) == null ? null : {
        for rk, rv in {
          requests = try(var.spec.prometheus.resources.requests, null) == null ? null : {
            for qk, qv in {
              cpu    = var.spec.prometheus.resources.requests.cpu != "" ? var.spec.prometheus.resources.requests.cpu : null
              memory = var.spec.prometheus.resources.requests.memory != "" ? var.spec.prometheus.resources.requests.memory : null
            } : qk => qv if qv != null
          }
          limits = try(var.spec.prometheus.resources.limits, null) == null ? null : {
            for lk, lv in {
              cpu    = var.spec.prometheus.resources.limits.cpu != "" ? var.spec.prometheus.resources.limits.cpu : null
              memory = var.spec.prometheus.resources.limits.memory != "" ? var.spec.prometheus.resources.limits.memory : null
            } : lk => lv if lv != null
          }
        } : rk => rv if rv != null && rv != {}
      }

      externalLabels     = length(try(var.spec.prometheus.external_labels, {})) > 0 ? var.spec.prometheus.external_labels : null
      scrapeInterval     = try(var.spec.prometheus.scrape_interval, "") != "" ? var.spec.prometheus.scrape_interval : null
      evaluationInterval = try(var.spec.prometheus.evaluation_interval, "") != "" ? var.spec.prometheus.evaluation_interval : null

      enableRemoteWriteReceiver = try(var.spec.prometheus.enable_remote_write_receiver, false) ? true : null

      # The five NilUsesHelmValues switches are the chart's documented
      # mechanism for "select every monitor object in the cluster".
      serviceMonitorSelectorNilUsesHelmValues = local.discovery_all ? false : null
      podMonitorSelectorNilUsesHelmValues     = local.discovery_all ? false : null
      ruleSelectorNilUsesHelmValues           = local.discovery_all ? false : null
      probeSelectorNilUsesHelmValues          = local.discovery_all ? false : null
      scrapeConfigSelectorNilUsesHelmValues   = local.discovery_all ? false : null

      remoteWrite = length(local.remote_write_values) > 0 ? local.remote_write_values : null

      # Raw scrape_config seam: an inline LIST of scrape_config entries,
      # outside the operator's validation (the spec comment carries the
      # warning).
      additionalScrapeConfigs = try(var.spec.prometheus.additional_scrape_configs, "") != "" ? yamldecode(var.spec.prometheus.additional_scrape_configs) : null

      nodeSelector = length(try(var.spec.prometheus.scheduling.node_selector, {})) > 0 ? var.spec.prometheus.scheduling.node_selector : null
      tolerations = length(try(var.spec.prometheus.scheduling.tolerations, [])) > 0 ? [
        for t in var.spec.prometheus.scheduling.tolerations : {
          for tk, tv in {
            key               = t.key != "" ? t.key : null
            operator          = t.operator != "" ? t.operator : null
            value             = t.value != "" ? t.value : null
            effect            = t.effect != "" ? t.effect : null
            tolerationSeconds = try(t.toleration_seconds, null)
          } : tk => tv if tv != null
        }
      ] : null
      priorityClassName = try(var.spec.prometheus.scheduling.priority_class_name, "") != "" ? var.spec.prometheus.scheduling.priority_class_name : null
    } : k => v if v != null
  }

  # ---- alertmanager ---------------------------------------------------------------
  alertmanager_ephemeral = try(var.spec.alertmanager.ephemeral, false)

  alertmanager_spec = {
    for k, v in {
      replicas  = try(coalesce(var.spec.alertmanager.replicas), null) != null ? var.spec.alertmanager.replicas : 1
      retention = try(coalesce(var.spec.alertmanager.retention), null)

      # A small PVC per replica BY DEFAULT (silences and the notification
      # log survive restarts); the ephemeral arm restores the chart's
      # emptyDir default.
      storage = local.alertmanager_ephemeral ? null : {
        volumeClaimTemplate = {
          spec = {
            for pk, pv in {
              accessModes      = ["ReadWriteOnce"]
              storageClassName = try(var.spec.alertmanager.storage_class, "") != "" ? var.spec.alertmanager.storage_class : null
              resources = {
                requests = {
                  storage = try(coalesce(var.spec.alertmanager.disk_size), null) != null ? var.spec.alertmanager.disk_size : "2Gi"
                }
              }
            } : pk => pv if pv != null
          }
        }
      }

      resources = try(var.spec.alertmanager.resources, null) == null ? null : {
        for rk, rv in {
          requests = try(var.spec.alertmanager.resources.requests, null) == null ? null : {
            for qk, qv in {
              cpu    = var.spec.alertmanager.resources.requests.cpu != "" ? var.spec.alertmanager.resources.requests.cpu : null
              memory = var.spec.alertmanager.resources.requests.memory != "" ? var.spec.alertmanager.resources.requests.memory : null
            } : qk => qv if qv != null
          }
          limits = try(var.spec.alertmanager.resources.limits, null) == null ? null : {
            for lk, lv in {
              cpu    = var.spec.alertmanager.resources.limits.cpu != "" ? var.spec.alertmanager.resources.limits.cpu : null
              memory = var.spec.alertmanager.resources.limits.memory != "" ? var.spec.alertmanager.resources.limits.memory : null
            } : lk => lv if lv != null
          }
        } : rk => rv if rv != null && rv != {}
      }

      nodeSelector = length(try(var.spec.alertmanager.scheduling.node_selector, {})) > 0 ? var.spec.alertmanager.scheduling.node_selector : null
      tolerations = length(try(var.spec.alertmanager.scheduling.tolerations, [])) > 0 ? [
        for t in var.spec.alertmanager.scheduling.tolerations : {
          for tk, tv in {
            key               = t.key != "" ? t.key : null
            operator          = t.operator != "" ? t.operator : null
            value             = t.value != "" ? t.value : null
            effect            = t.effect != "" ? t.effect : null
            tolerationSeconds = try(t.toleration_seconds, null)
          } : tk => tv if tv != null
        }
      ] : null
      priorityClassName = try(var.spec.alertmanager.scheduling.priority_class_name, "") != "" ? var.spec.alertmanager.scheduling.priority_class_name : null

      # The notification credentials Secret, mounted by the operator at
      # /etc/alertmanager/secrets/<secret>/ for the receivers' `_file`
      # fields (typed notifications below).
      secrets = length(local.notifications_secret_data) > 0 ? [local.notifications_secret_name] : null
    } : k => v if v != null
  }

  # ---- typed notifications (twin of the Pulumi module's notifications.go) ---
  # Three pieces the chart already carries: the alertmanager.config map
  # (routing tree and receivers; Helm merges it over the chart's default
  # config, so the chart's inhibit rules, resolve timeout and template glob
  # stay the chart's), alertmanager.templateFiles (the one message
  # template), and the module-owned credentials Secret, listed in
  # alertmanagerSpec.secrets so the operator mounts it at
  # /etc/alertmanager/secrets/<secret>/ for the receivers' `_file` fields.
  # No credential lands in chart values or in the rendered configuration.
  notifications = local.alertmanager_enabled ? try(var.spec.alertmanager.notifications, null) : null

  notifications_secret_name = local.notifications == null ? "" : "${local.release_name}-alertmanager-notifications"
  notifications_secret_dir  = "/etc/alertmanager/secrets/${local.notifications_secret_name}"
  notifications_heartbeat   = local.notifications == null ? false : try(local.notifications.heartbeat, null) != null
  notifications_interval    = local.notifications_heartbeat ? coalesce(local.notifications.heartbeat.interval, "1m") : ""

  notification_env_label       = local.notifications == null ? "" : try(coalesce(local.notifications.message.environment_label), "environment")
  notification_component_label = local.notifications == null ? "" : try(coalesce(local.notifications.message.component_label), "component")

  # Deterministic Secret keys: <receiver>-<integration>-<index>-<field>,
  # plus heartbeat-bearer-token (twin: the Pulumi module's key format).
  notifications_secret_data = local.notifications == null ? {} : merge(
    merge(concat([{}], [for r in local.notifications.receivers : merge(
      { for j, d in r.discord : "${r.name}-discord-${j}-webhook-url" => d.webhook_url },
      merge(concat([{}], [for j, p in r.pushover : {
        "${r.name}-pushover-${j}-token"    = p.token
        "${r.name}-pushover-${j}-user-key" = p.user_key
      }])...),
      { for j, w in r.webhook : "${r.name}-webhook-${j}-bearer-token" => w.bearer_token if w.bearer_token != "" }
    )])...),
    local.notifications_heartbeat && try(local.notifications.heartbeat.bearer_token, "") != "" ? {
      "heartbeat-bearer-token" = local.notifications.heartbeat.bearer_token
    } : {}
  )

  notification_title = "{{ template \"planton.title\" . }}"
  notification_text  = "{{ template \"planton.text\" . }}"

  # The one message shape every Discord and Pushover notification carries
  # (twin: notificationTemplate, byte for byte, held equal by the Pulumi
  # module's template_twin_test.go). The title leads with environment and
  # component (else the scrape job, else "cluster": the alert is about the
  # whole cluster); the body is the first alert's customer_impact (else
  # summary, else name), a count when the group holds more, and the runbook.
  # It renders ONLY those labels and annotations: never description,
  # namespace or pod, because on a shared cluster a namespace can name a
  # customer and upstream rule descriptions interpolate it.
  notification_template = <<EOT
{{ define "planton.title" }}{{ if eq .Status "resolved" }}[RESOLVED] {{ end }}[{{ or (index .CommonLabels "${local.notification_env_label}") "unknown" }}] {{ or (index .CommonLabels "${local.notification_component_label}") .CommonLabels.job "cluster" }}: {{ .CommonLabels.alertname }}{{ end }}
{{ define "planton.text" }}{{ with index .Alerts 0 }}{{ or .Annotations.customer_impact .Annotations.summary .Labels.alertname }}{{ end }}{{ if gt (len .Alerts) 1 }} ({{ len .Alerts }} alerts){{ end }}{{ with (index .Alerts 0).Annotations.runbook_url }}
Runbook: {{ . }}{{ end }}{{ end }}
EOT

  # A firing alert pushes at the declared priority, the resolved follow-up
  # at normal, so a recovery never pages (twin: pushoverPriority).
  pushover_priority = {
    normal    = "0"
    high      = "{{ if eq .Status \"firing\" }}1{{ else }}0{{ end }}"
    emergency = "{{ if eq .Status \"firing\" }}2{{ else }}0{{ end }}"
  }
  matcher_operator = {
    equals            = "="
    not_equals        = "!="
    matches_regex     = "=~"
    not_matches_regex = "!~"
  }

  # null (never []) when nothing is declared: the receivers and routes are
  # tuples of differently-shaped objects, which a conditional cannot unify
  # with an empty tuple. Both are read only inside notification_config.
  notification_receivers = local.notifications == null ? null : concat(
    [{ name = "discard" }],
    local.notifications_heartbeat ? [{
      name = "heartbeat"
      webhook_configs = [{
        for k, v in {
          url           = local.notifications.heartbeat.url
          send_resolved = false
          max_alerts    = 1
          http_config = try(local.notifications.heartbeat.bearer_token, "") != "" ? {
            authorization = { type = "Bearer", credentials_file = "${local.notifications_secret_dir}/heartbeat-bearer-token" }
          } : null
        } : k => v if v != null
      }]
    }] : [],
    [for r in local.notifications.receivers : {
      for k, v in {
        name = r.name
        discord_configs = length(r.discord) > 0 ? [for j, d in r.discord : {
          webhook_url_file = "${local.notifications_secret_dir}/${r.name}-discord-${j}-webhook-url"
          title            = local.notification_title
          message          = local.notification_text
        }] : null
        pushover_configs = length(r.pushover) > 0 ? [for j, p in r.pushover : {
          token_file    = "${local.notifications_secret_dir}/${r.name}-pushover-${j}-token"
          user_key_file = "${local.notifications_secret_dir}/${r.name}-pushover-${j}-user-key"
          title         = local.notification_title
          message       = local.notification_text
          url           = "{{ (index .Alerts 0).Annotations.runbook_url }}"
          url_title     = "Runbook"
          priority      = lookup(local.pushover_priority, try(coalesce(p.priority), "emergency"), local.pushover_priority.emergency)
        }] : null
        webhook_configs = length(r.webhook) > 0 ? [for j, w in r.webhook : {
          for wk, wv in {
            url = w.url
            http_config = w.bearer_token != "" ? {
              authorization = { type = "Bearer", credentials_file = "${local.notifications_secret_dir}/${r.name}-webhook-${j}-bearer-token" }
            } : null
          } : wk => wv if wv != null
        }] : null
      } : k => v if v != null
    }]
  )

  # The module-owned Watchdog and InfoInhibitor routes come first, so
  # declared routes never see either; an always-firing alert is re-sent
  # only every repeat_interval, so the heartbeat route repeats at its own
  # interval, well inside the outside monitor's staleness window.
  notification_routes = local.notifications == null ? null : concat(
    [
      {
        for k, v in {
          receiver        = local.notifications_heartbeat ? "heartbeat" : "discard"
          matchers        = ["alertname=\"Watchdog\""]
          group_wait      = local.notifications_heartbeat ? "0s" : null
          group_interval  = local.notifications_heartbeat ? local.notifications_interval : null
          repeat_interval = local.notifications_heartbeat ? local.notifications_interval : null
        } : k => v if v != null
      },
      { receiver = "discard", matchers = ["alertname=\"InfoInhibitor\""] },
    ],
    [for c in local.notifications.route.routes : {
      for k, v in {
        receiver = c.receiver
        matchers = [for m in c.matchers : format("%s%s\"%s\"",
          m.label,
          lookup(local.matcher_operator, try(coalesce(m.operator), "equals"), "="),
          replace(replace(m.value, "\\", "\\\\"), "\"", "\\\"")
        )]
        continue        = c.continue_matching ? true : null
        repeat_interval = c.repeat_interval != "" ? c.repeat_interval : null
      } : k => v if v != null
    }]
  )

  notification_config = local.notifications == null ? null : {
    route = {
      for k, v in {
        receiver        = local.notifications.route.receiver
        group_by        = length(local.notifications.route.group_by) > 0 ? local.notifications.route.group_by : ["alertname", local.notification_env_label, local.notification_component_label]
        group_wait      = local.notifications.route.group_wait != "" ? local.notifications.route.group_wait : null
        group_interval  = local.notifications.route.group_interval != "" ? local.notifications.route.group_interval : null
        repeat_interval = local.notifications.route.repeat_interval != "" ? local.notifications.route.repeat_interval : null
        routes          = local.notification_routes
      } : k => v if v != null
    }
    receivers = local.notification_receivers
  }

  # The raw config document and the typed notifications are mutually
  # exclusive (a spec rule); one() picks whichever is declared without
  # forcing their two object types to unify in a conditional.
  alertmanager_config_yaml = local.alertmanager_enabled && try(var.spec.alertmanager.config_yaml, "") != "" ? yamldecode(var.spec.alertmanager.config_yaml) : null

  # The alerting configuration (route/receivers). The chart value is a MAP
  # (rendered into the Alertmanager Secret); neither declared = the chart's
  # null-receiver + Watchdog default. Credential discipline lives on the
  # spec field comments.
  alertmanager_values = {
    for k, v in {
      enabled          = local.alertmanager_enabled
      alertmanagerSpec = local.alertmanager_enabled ? local.alertmanager_spec : null
      config           = one([for c in [local.alertmanager_config_yaml, local.notification_config] : c if c != null])
      templateFiles    = local.notification_config != null ? { "planton-notifications.tmpl" = local.notification_template } : null
    } : k => v if v != null
  }

  # ---- bundled grafana --------------------------------------------------------------
  grafana_values = {
    for k, v in {
      enabled = local.grafana_enabled

      defaultDashboardsEnabled = local.grafana_enabled && try(var.spec.grafana.default_dashboards_enabled, null) == false ? false : null

      # Existing-secret arm: the subchart wires the referenced Secret's
      # keys into env at pod start. Generate arm (admin_secret absent):
      # the subchart creates its own `<name>-grafana` Secret ONCE
      # (lookup-stable across upgrades). Credential material never lands
      # in chart values either way.
      admin = local.grafana_enabled && try(var.spec.grafana.admin_secret, null) != null ? {
        existingSecret = var.spec.grafana.admin_secret.name
        userKey        = try(coalesce(var.spec.grafana.admin_secret.user_key), null) != null ? var.spec.grafana.admin_secret.user_key : "admin-user"
        passwordKey    = try(coalesce(var.spec.grafana.admin_secret.password_key), null) != null ? var.spec.grafana.admin_secret.password_key : "admin-password"
      } : null

      # Rendered only when declared: the subchart default is ephemeral —
      # honest for dashboards-as-code (the spec comment carries the
      # trade).
      persistence = local.grafana_enabled && try(var.spec.grafana.storage, null) != null ? {
        for pk, pv in {
          enabled          = true
          size             = try(coalesce(var.spec.grafana.storage.size), null) != null ? var.spec.grafana.storage.size : "10Gi"
          storageClassName = var.spec.grafana.storage.storage_class != "" ? var.spec.grafana.storage.storage_class : null
        } : pk => pv if pv != null
      } : null

      resources = local.grafana_enabled && try(var.spec.grafana.resources, null) != null ? {
        for rk, rv in {
          requests = try(var.spec.grafana.resources.requests, null) == null ? null : {
            for qk, qv in {
              cpu    = var.spec.grafana.resources.requests.cpu != "" ? var.spec.grafana.resources.requests.cpu : null
              memory = var.spec.grafana.resources.requests.memory != "" ? var.spec.grafana.resources.requests.memory : null
            } : qk => qv if qv != null
          }
          limits = try(var.spec.grafana.resources.limits, null) == null ? null : {
            for lk, lv in {
              cpu    = var.spec.grafana.resources.limits.cpu != "" ? var.spec.grafana.resources.limits.cpu : null
              memory = var.spec.grafana.resources.limits.memory != "" ? var.spec.grafana.resources.limits.memory : null
            } : lk => lv if lv != null
          }
        } : rk => rv if rv != null && rv != {}
      } : null
    } : k => v if v != null
  }

  # ---- operator -----------------------------------------------------------------------
  # Admission webhooks: chart-default ON with the self-contained certgen
  # hook Job. The cert-manager arm swaps the certificate machinery
  # (requires KubernetesCertManager); the disabled arm turns validation
  # off entirely (rules then fail at config-reload time, not admission).
  admission_webhooks_disabled     = try(var.spec.operator.admission_webhooks.disabled, false)
  admission_webhooks_cert_manager = try(var.spec.operator.admission_webhooks.cert_manager, false)

  operator_values = {
    for k, v in {
      resources = try(var.spec.operator.resources, null) == null ? null : {
        for rk, rv in {
          requests = try(var.spec.operator.resources.requests, null) == null ? null : {
            for qk, qv in {
              cpu    = var.spec.operator.resources.requests.cpu != "" ? var.spec.operator.resources.requests.cpu : null
              memory = var.spec.operator.resources.requests.memory != "" ? var.spec.operator.resources.requests.memory : null
            } : qk => qv if qv != null
          }
          limits = try(var.spec.operator.resources.limits, null) == null ? null : {
            for lk, lv in {
              cpu    = var.spec.operator.resources.limits.cpu != "" ? var.spec.operator.resources.limits.cpu : null
              memory = var.spec.operator.resources.limits.memory != "" ? var.spec.operator.resources.limits.memory : null
            } : lk => lv if lv != null
          }
        } : rk => rv if rv != null && rv != {}
      }

      admissionWebhooks = (local.admission_webhooks_disabled || local.admission_webhooks_cert_manager) ? {
        for ak, av in {
          enabled     = local.admission_webhooks_disabled ? false : null
          patch       = local.admission_webhooks_disabled ? { enabled = false } : null
          certManager = local.admission_webhooks_cert_manager ? { enabled = true } : null
        } : ak => av if av != null
      } : null

      nodeSelector = length(try(var.spec.operator.scheduling.node_selector, {})) > 0 ? var.spec.operator.scheduling.node_selector : null
      tolerations = length(try(var.spec.operator.scheduling.tolerations, [])) > 0 ? [
        for t in var.spec.operator.scheduling.tolerations : {
          for tk, tv in {
            key               = t.key != "" ? t.key : null
            operator          = t.operator != "" ? t.operator : null
            value             = t.value != "" ? t.value : null
            effect            = t.effect != "" ? t.effect : null
            tolerationSeconds = try(t.toleration_seconds, null)
          } : tk => tv if tv != null
        }
      ] : null
      priorityClassName = try(var.spec.operator.scheduling.priority_class_name, "") != "" ? var.spec.operator.scheduling.priority_class_name : null
    } : k => v if v != null
  }

  # ---- exporters --------------------------------------------------------------------------
  kube_state_metrics_enabled = try(coalesce(var.spec.exporters.kube_state_metrics_enabled), null) != null ? var.spec.exporters.kube_state_metrics_enabled : true
  node_exporter_enabled      = try(coalesce(var.spec.exporters.node_exporter_enabled), null) != null ? var.spec.exporters.node_exporter_enabled : true

  kube_state_metrics_resources = try(var.spec.exporters.kube_state_metrics_resources, null) == null ? null : {
    for rk, rv in {
      requests = try(var.spec.exporters.kube_state_metrics_resources.requests, null) == null ? null : {
        for qk, qv in {
          cpu    = var.spec.exporters.kube_state_metrics_resources.requests.cpu != "" ? var.spec.exporters.kube_state_metrics_resources.requests.cpu : null
          memory = var.spec.exporters.kube_state_metrics_resources.requests.memory != "" ? var.spec.exporters.kube_state_metrics_resources.requests.memory : null
        } : qk => qv if qv != null
      }
      limits = try(var.spec.exporters.kube_state_metrics_resources.limits, null) == null ? null : {
        for lk, lv in {
          cpu    = var.spec.exporters.kube_state_metrics_resources.limits.cpu != "" ? var.spec.exporters.kube_state_metrics_resources.limits.cpu : null
          memory = var.spec.exporters.kube_state_metrics_resources.limits.memory != "" ? var.spec.exporters.kube_state_metrics_resources.limits.memory : null
        } : lk => lv if lv != null
      }
    } : rk => rv if rv != null && rv != {}
  }

  node_exporter_resources = try(var.spec.exporters.node_exporter_resources, null) == null ? null : {
    for rk, rv in {
      requests = try(var.spec.exporters.node_exporter_resources.requests, null) == null ? null : {
        for qk, qv in {
          cpu    = var.spec.exporters.node_exporter_resources.requests.cpu != "" ? var.spec.exporters.node_exporter_resources.requests.cpu : null
          memory = var.spec.exporters.node_exporter_resources.requests.memory != "" ? var.spec.exporters.node_exporter_resources.requests.memory : null
        } : qk => qv if qv != null
      }
      limits = try(var.spec.exporters.node_exporter_resources.limits, null) == null ? null : {
        for lk, lv in {
          cpu    = var.spec.exporters.node_exporter_resources.limits.cpu != "" ? var.spec.exporters.node_exporter_resources.limits.cpu : null
          memory = var.spec.exporters.node_exporter_resources.limits.memory != "" ? var.spec.exporters.node_exporter_resources.limits.memory : null
        } : lk => lv if lv != null
      }
    } : rk => rv if rv != null && rv != {}
  }

  # ---- control-plane scrapers ------------------------------------------------------------------
  # Rendered only when the manifest carries an explicit decision (the
  # managed-cloud posture — see the spec's MANAGED CLOUDS note and the
  # per-cloud presets); chart defaults stay untouched otherwise.
  scraper_toggles = {
    for k, v in {
      kubeApiServer         = try(coalesce(var.spec.control_plane_scrapers.kube_api_server), null)
      kubelet               = try(coalesce(var.spec.control_plane_scrapers.kubelet), null)
      kubeControllerManager = try(coalesce(var.spec.control_plane_scrapers.kube_controller_manager), null)
      coreDns               = try(coalesce(var.spec.control_plane_scrapers.core_dns), null)
      kubeEtcd              = try(coalesce(var.spec.control_plane_scrapers.kube_etcd), null)
      kubeScheduler         = try(coalesce(var.spec.control_plane_scrapers.kube_scheduler), null)
      kubeProxy             = try(coalesce(var.spec.control_plane_scrapers.kube_proxy), null)
    } : k => { enabled = v } if v != null
  }

  # ---- default rules ------------------------------------------------------------------------------
  default_rules_values = {
    for k, v in {
      create = try(var.spec.default_rules.enabled, null) == false ? false : null
      rules = length(try(var.spec.default_rules.disabled_groups, [])) > 0 ? {
        for group in var.spec.default_rules.disabled_groups : group => false
      } : null
      # The chart gates every curated alert on defaultRules.disabled.<alert>.
      disabled = length(try(var.spec.default_rules.disabled_alerts, [])) > 0 ? {
        for alert in var.spec.default_rules.disabled_alerts : alert => true
      } : null
    } : k => v if v != null
  }

  # The chart reads a curated alert's hold and severity from the top-level
  # customRules map; only the keys the manifest sets are rendered, so an
  # override of one keeps the chart's own value for the other. ("for" is
  # quoted: bare, HCL reads it as a for expression.)
  custom_rules_values = {
    for o in try(var.spec.default_rules.alert_overrides, []) : o.alert => {
      for k, v in {
        "for"    = o.for_duration != "" ? o.for_duration : null
        severity = o.severity != "" ? o.severity : null
      } : k => v if v != null
    }
  }

  # ---- crds subchart ---------------------------------------------------------------------------------
  # skip_crds and the upgradeJob render DIFFERENT object shapes under the
  # same `crds` key — a single ternary cannot unify them (plan-time
  # type-unification), so the value is built with the null-prune idiom
  # and merged only when non-empty.
  crds_values = {
    for k, v in {
      enabled    = var.spec.skip_crds ? false : null
      upgradeJob = (!var.spec.skip_crds && var.spec.crd_upgrade_job) ? { enabled = true } : null
    } : k => v if v != null
  }

  # ---- typed chart values (twin of the Pulumi module's buildHelmValues) --
  helm_values = merge(
    {
      # fullnameOverride pins the fullname to the resource name: every
      # child name derives deterministically, and the exported outputs
      # are built from that contract.
      fullnameOverride = local.release_name

      prometheus = { prometheusSpec = local.prometheus_spec }

      alertmanager = local.alertmanager_values
      grafana      = local.grafana_values

      # The toggle keys gate the subcharts; the subchart config keys
      # carry their resources.
      kubeStateMetrics = { enabled = local.kube_state_metrics_enabled }
      nodeExporter     = { enabled = local.node_exporter_enabled }
    },
    # The CRDs ship in the chart's local `crds` subchart whose crds/
    # directory Helm installs ONCE: upgrades never touch them, uninstall
    # keeps them. skip_crds is the bring-your-own-CRDs arm; the
    # upgradeJob is the chart's own pre-upgrade hook that
    # server-side-applies the bundle across operator versions. (Rendered
    # via the pruned crds_values local — the two arms carry different
    # object shapes a ternary cannot unify.)
    length(local.crds_values) > 0 ? { crds = local.crds_values } : {},
    # imageRegistry replaces the registry of EVERY image the stack pulls
    # (the air-gap path); imagePullSecrets must reach every workload —
    # which is exactly what the chart's global block does. The Pulumi
    # twin renders the same shape (a silent single-engine
    # ImagePullBackOff class otherwise).
    (var.spec.image_registry != "" || length(var.spec.image_pull_secrets) > 0) ? {
      global = {
        for gk, gv in {
          imageRegistry    = var.spec.image_registry != "" ? var.spec.image_registry : null
          imagePullSecrets = length(var.spec.image_pull_secrets) > 0 ? [for s in var.spec.image_pull_secrets : { name = s }] : null
        } : gk => gv if gv != null
      }
    } : {},
    length(local.operator_values) > 0 ? { prometheusOperator = local.operator_values } : {},
    local.kube_state_metrics_resources != null ? { "kube-state-metrics" = { resources = local.kube_state_metrics_resources } } : {},
    local.node_exporter_resources != null ? { "prometheus-node-exporter" = { resources = local.node_exporter_resources } } : {},
    local.scraper_toggles,
    length(local.default_rules_values) > 0 ? { defaultRules = local.default_rules_values } : {},
    length(local.custom_rules_values) > 0 ? { customRules = local.custom_rules_values } : {}
  )
}
