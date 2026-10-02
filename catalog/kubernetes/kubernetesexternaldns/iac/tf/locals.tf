# Computed values for the KubernetesExternalDns module. Every resolution here
# has an exact twin in the Pulumi module's locals.go / values.go / secrets.go —
# keep them in lockstep.
#
# HCL DISCIPLINE (applies to every conditional object in this file):
# conditional entries are written as `key = cond ? value : null` inside ONE
# object literal, pruned with `{ for k, v in {...} : k => v if v != null }`.
# The two tempting alternatives are both broken: `cond ? {...} : {}`
# ternaries fail plan-time type unification when branches carry different
# attributes, and `merge(concat(cond ? [{...}] : [], ...)...)` silently
# UNIFIES primitive-only sibling objects into map(string) — numbers and
# booleans arrive in the chart values as strings. The null-prune form
# preserves every value's type. Lists of same-shaped elements assemble with
# concat(); lists of DIFFERENTLY-shaped elements (env entries, volume mounts)
# assemble as `[for e in [cond ? {...} : null, ...] : e if e != null]` so no
# two shapes ever meet in a ternary.
#
# Optional nested blocks are read with try(): HCL's && does NOT
# short-circuit, so chained null checks still dereference the null.

locals {
  # Chart identity — must stay byte-identical with the Pulumi module's vars:
  # cross-engine chart-name drift deploys two different products from one
  # manifest.
  helm_chart_name = "external-dns"
  helm_chart_repo = "https://kubernetes-sigs.github.io/external-dns/"

  # Release name — metadata.name, NOT a fixed chart name: multiple
  # ExternalDNS instances per cluster are a first-class pattern (one per DNS
  # provider / zone set, separated by TXT owner IDs), so each manifest gets
  # its own release.
  release_name = var.metadata.name

  # Chart version resolved to the pinned default when unset, so both engines
  # install the same chart whether or not the platform's defaulting
  # middleware ran — mirror of the Pulumi module's DefaultChartVersion.
  chart_version = coalesce(var.spec.chart_version, "1.21.1")

  namespace = var.spec.namespace

  # Controller ServiceAccount name. The chart creates the SA; the module
  # pins its name to metadata.name (serviceAccount.name) so cloud-side
  # identity bindings have a deterministic subject.
  service_account_name = var.metadata.name

  # Resource-identity labels stamped on the module-created satellites
  # (namespace, credential Secrets — never injected into the chart's own
  # resources; Helm owns those).
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesExternalDns"
    },
    var.metadata.id != null && var.metadata.id != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    var.metadata.org != null && var.metadata.org != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # Deterministic names for the credential satellites this module may
  # materialize (which ones actually exist depends on the selected provider
  # arm and whether static credentials were declared — see main.tf).
  cloudflare_secret_name = "${var.metadata.name}-cloudflare-credentials"
  aws_secret_name        = "${var.metadata.name}-aws-credentials"
  gcp_secret_name        = "${var.metadata.name}-gcp-credentials"
  azure_secret_name      = "${var.metadata.name}-azure-config"

  # ---- workload identity annotations -----------------------------------
  # The chart creates the controller ServiceAccount; the identity annotation
  # rides serviceAccount.annotations. AKS additionally needs the
  # azure.workload.identity/use pod label (the azure-workload-identity
  # webhook only injects the federated token volume into pods carrying that
  # label — the SA annotation alone does nothing).
  workload_identity_annotations = merge(
    try(var.spec.workload_identity.gke, null) != null ? {
      "iam.gke.io/gcp-service-account" = var.spec.workload_identity.gke.service_account_email
    } : {},
    try(var.spec.workload_identity.eks, null) != null ? {
      "eks.amazonaws.com/role-arn" = var.spec.workload_identity.eks.role_arn
    } : {},
    try(var.spec.workload_identity.aks, null) != null ? merge(
      {
        "azure.workload.identity/client-id" = var.spec.workload_identity.aks.client_id
      },
      try(var.spec.workload_identity.aks.tenant_id, null) != null ? {
        "azure.workload.identity/tenant-id" = var.spec.workload_identity.aks.tenant_id
      } : {}
    ) : {}
  )

  # ---- container resources (shared ContainerResources shape) ------------
  controller_resources = try(var.spec.resources, null) == null ? null : {
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

  webhook_resources = try(var.spec.webhook.resources, null) == null ? null : {
    for k, v in {
      limits = try(var.spec.webhook.resources.limits, null) == null ? null : {
        for lk, lv in {
          cpu    = try(var.spec.webhook.resources.limits.cpu, "") != "" ? var.spec.webhook.resources.limits.cpu : null
          memory = try(var.spec.webhook.resources.limits.memory, "") != "" ? var.spec.webhook.resources.limits.memory : null
        } : lk => lv if lv != null
      }
      requests = try(var.spec.webhook.resources.requests, null) == null ? null : {
        for rk, rv in {
          cpu    = try(var.spec.webhook.resources.requests.cpu, "") != "" ? var.spec.webhook.resources.requests.cpu : null
          memory = try(var.spec.webhook.resources.requests.memory, "") != "" ? var.spec.webhook.resources.requests.memory : null
        } : rk => rv if rv != null
      }
    } : k => v if v != null && length(v) > 0
  }

  # ---- provider selection (spec oneof: exactly one arm is set) -----------
  # The chart's provider.name per arm — Azure switches to the
  # "azure-private-dns" upstream provider for private zones.
  provider_name = (
    try(var.spec.aws_route53, null) != null ? "aws" :
    try(var.spec.google_cloud_dns, null) != null ? "google" :
    try(var.spec.azure_dns, null) != null ? (var.spec.azure_dns.private_zones ? "azure-private-dns" : "azure") :
    try(var.spec.cloudflare, null) != null ? "cloudflare" :
    try(var.spec.webhook, null) != null ? "webhook" :
    "inmemory"
  )

  # GCP mounts a key file only when a static key was declared; keyless
  # installs (workload identity / node SA) mount nothing.
  gcp_key_mounted = try(var.spec.google_cloud_dns, null) != null && try(var.spec.google_cloud_dns.service_account_key_json, "") != ""

  # Webhook arm: provider.name=webhook makes the chart run the provider
  # image as a sidecar next to the controller, talking over localhost.
  # Env entries render sorted by name — deterministic rendering across
  # engines for map-shaped spec fields that land in ordered lists.
  webhook_provider = try(var.spec.webhook, null) == null ? null : {
    for k, v in {
      image = {
        for ik, iv in {
          repository = var.spec.webhook.image_repository
          tag        = var.spec.webhook.image_tag != "" ? var.spec.webhook.image_tag : null
        } : ik => iv if iv != null
      }
      args = length(var.spec.webhook.args) > 0 ? var.spec.webhook.args : null
      env = length(var.spec.webhook.env) > 0 ? [
        for name in sort(keys(var.spec.webhook.env)) : {
          name  = name
          value = var.spec.webhook.env[name]
        }
      ] : null
      resources = local.webhook_resources
    } : k => v if v != null
  }

  provider_block = {
    for k, v in {
      name    = local.provider_name
      webhook = local.webhook_provider
    } : k => v if v != null
  }

  # ---- provider CLI flags (extraArgs) ------------------------------------
  # Assembled in a FIXED order — the SAME order as the Pulumi module's
  # applyProvider — so both engines render byte-identical lists: each arm's
  # zone-id filters first, then its arm-specific flags. The dynamodb registry
  # settings are AWS-scoped flags. Only one arm's branch is ever non-empty
  # (the spec oneof).
  provider_extra_args = concat(
    try(var.spec.aws_route53, null) == null ? [] : concat(
      [for z in try(var.spec.aws_route53.zone_id_filters, []) : "--zone-id-filter=${z}" if z != ""],
      compact([
        # zone_type is an OPTIONAL scalar: unset arrives as null (not ""),
        # and null passes a `try(x, "") != ""` guard straight into the
        # template. try(coalesce(x), "") is the null-safe read — coalesce
        # rejects a lone null/empty, and try turns that rejection into "".
        try(coalesce(var.spec.aws_route53.zone_type), "") != "" ? "--aws-zone-type=${var.spec.aws_route53.zone_type}" : "",
        try(var.spec.aws_route53.assume_role, "") != "" ? "--aws-assume-role=${var.spec.aws_route53.assume_role}" : "",
        try(var.spec.aws_route53.assume_role_external_id, "") != "" ? "--aws-assume-role-external-id=${var.spec.aws_route53.assume_role_external_id}" : "",
        var.spec.dynamodb_region != "" ? "--dynamodb-region=${var.spec.dynamodb_region}" : "",
        var.spec.dynamodb_table != "" ? "--dynamodb-table=${var.spec.dynamodb_table}" : "",
      ])
    ),
    try(var.spec.google_cloud_dns, null) == null ? [] : concat(
      ["--google-project=${var.spec.google_cloud_dns.project}"],
      [for z in try(var.spec.google_cloud_dns.zone_id_filters, []) : "--zone-id-filter=${z}" if z != ""],
      compact([
        # zone_visibility: same optional-scalar null trap as zone_type above.
        try(coalesce(var.spec.google_cloud_dns.zone_visibility), "") != "" ? "--google-zone-visibility=${var.spec.google_cloud_dns.zone_visibility}" : "",
      ])
    ),
    try(var.spec.azure_dns, null) == null ? [] : (
      [for z in try(var.spec.azure_dns.zone_id_filters, []) : "--zone-id-filter=${z}" if z != ""]
    ),
    try(var.spec.cloudflare, null) == null ? [] : concat(
      [for z in try(var.spec.cloudflare.zone_id_filters, []) : "--zone-id-filter=${z}" if z != ""],
      compact([
        try(var.spec.cloudflare.proxied, false) ? "--cloudflare-proxied" : "",
      ]),
      try(var.spec.cloudflare.dns_records_per_page, null) != null ? [
        "--cloudflare-dns-records-per-page=${var.spec.cloudflare.dns_records_per_page}"
      ] : []
    ),
    try(var.spec.in_memory, null) == null ? [] : (
      [for z in try(var.spec.in_memory.zones, []) : "--inmemory-zone=${z}"]
    )
  )

  # ---- provider credential env wiring -------------------------------------
  # Secret-backed entries reference the Secrets main.tf materializes — the
  # credential itself never appears in chart values or pod specs. Entry
  # shapes differ (plain value vs valueFrom), so this list is assembled with
  # the null-prune form, never a ternary between shapes.
  provider_env = [for e in [
    # AWS: the SDK requires a region even though Route 53 is global.
    try(var.spec.aws_route53, null) != null && try(var.spec.aws_route53.region, "") != "" ? {
      name  = "AWS_DEFAULT_REGION"
      value = var.spec.aws_route53.region
    } : null,
    # AWS static keys: declared as a pair (proto CEL rule) — key presence
    # gates both entries.
    try(var.spec.aws_route53.access_key_id, "") != "" ? {
      name      = "AWS_ACCESS_KEY_ID"
      valueFrom = { secretKeyRef = { name = local.aws_secret_name, key = "access-key-id" } }
    } : null,
    try(var.spec.aws_route53.access_key_id, "") != "" ? {
      name      = "AWS_SECRET_ACCESS_KEY"
      valueFrom = { secretKeyRef = { name = local.aws_secret_name, key = "secret-access-key" } }
    } : null,
    # GCP: ADC reads the key from a file — point
    # GOOGLE_APPLICATION_CREDENTIALS at the mounted Secret.
    local.gcp_key_mounted ? {
      name  = "GOOGLE_APPLICATION_CREDENTIALS"
      value = "/etc/kubernetes/gcp/credentials.json"
    } : null,
    # Cloudflare: always token-authenticated.
    try(var.spec.cloudflare, null) != null ? {
      name      = "CF_API_TOKEN"
      valueFrom = { secretKeyRef = { name = local.cloudflare_secret_name, key = "api-token" } }
    } : null,
  ] : e if e != null]

  # ---- credential file mounts ---------------------------------------------
  # GCP mounts the key directory; Azure mounts azure.json at the
  # controller's DEFAULT config path (so no --azure-config-file override is
  # needed) — the controller reads identity + subscription + resource group
  # from it. Mount shapes differ (subPath), hence the null-prune list form.
  extra_volumes = [for v in [
    local.gcp_key_mounted ? {
      name   = "gcp-credentials"
      secret = { secretName = local.gcp_secret_name }
    } : null,
    try(var.spec.azure_dns, null) != null ? {
      name   = "azure-config"
      secret = { secretName = local.azure_secret_name }
    } : null,
  ] : v if v != null]

  extra_volume_mounts = [for m in [
    local.gcp_key_mounted ? {
      name      = "gcp-credentials"
      mountPath = "/etc/kubernetes/gcp"
      readOnly  = true
    } : null,
    try(var.spec.azure_dns, null) != null ? {
      name      = "azure-config"
      mountPath = "/etc/kubernetes/azure.json"
      subPath   = "azure.json"
      readOnly  = true
    } : null,
  ] : m if m != null]

  # ---- azure.json ----------------------------------------------------------
  # Twin of the Pulumi module's renderAzureJson. Identity selection order
  # (documented on the spec): service principal when client_id/client_secret
  # are set; otherwise Workload Identity when workload_identity.aks is set;
  # otherwise managed identity (user-assigned when managed_identity_client_id
  # is set, else system-assigned). Both engines emit keys sorted
  # alphabetically (Go json.Marshal / HCL jsonencode) — byte-identical
  # documents.
  azure_config_json = try(var.spec.azure_dns, null) == null ? null : jsonencode({
    for k, v in {
      subscriptionId  = var.spec.azure_dns.subscription_id
      resourceGroup   = var.spec.azure_dns.resource_group
      tenantId        = try(var.spec.azure_dns.tenant_id, "") != "" ? var.spec.azure_dns.tenant_id : null
      aadClientId     = try(var.spec.azure_dns.client_id, "") != "" ? var.spec.azure_dns.client_id : null
      aadClientSecret = try(var.spec.azure_dns.client_id, "") != "" ? var.spec.azure_dns.client_secret : null
      useWorkloadIdentityExtension = (
        try(var.spec.azure_dns.client_id, "") == "" && try(var.spec.workload_identity.aks, null) != null
      ) ? true : null
      useManagedIdentityExtension = (
        try(var.spec.azure_dns.client_id, "") == "" && try(var.spec.workload_identity.aks, null) == null
      ) ? true : null
      userAssignedIdentityID = (
        try(var.spec.azure_dns.client_id, "") == "" &&
        try(var.spec.workload_identity.aks, null) == null &&
        try(var.spec.azure_dns.managed_identity_client_id, "") != ""
      ) ? var.spec.azure_dns.managed_identity_client_id : null
    } : k => v if v != null
  })

  # ---- typed chart values (twin of the Pulumi module's buildHelmValues) --
  helm_values = {
    for k, v in {
      # Pin the chart's fullname to the release name (= metadata.name): every
      # chart object (Deployment, Service, RBAC) then carries a
      # deterministic, manifest-derived name — what verification, imports,
      # and multi-instance coexistence all key off.
      fullnameOverride = local.release_name

      # The chart creates the controller ServiceAccount; the module pins its
      # name (deterministic identity subject) and rides workload-identity
      # annotations on it.
      serviceAccount = {
        for sk, sv in {
          name        = local.service_account_name
          annotations = length(local.workload_identity_annotations) > 0 ? local.workload_identity_annotations : null
        } : sk => sv if sv != null
      }
      podLabels = try(var.spec.workload_identity.aks, null) != null ? {
        "azure.workload.identity/use" = "true"
      } : null

      # ---- watching / sync behavior ------------------------------------
      sources                   = length(var.spec.sources) > 0 ? var.spec.sources : null
      enableGatewayListenerSets = var.spec.gateway_listener_sets ? true : null
      policy                    = try(var.spec.policy, null)
      registry                  = try(var.spec.registry, null)
      txtOwnerId                = var.spec.txt_owner_id != "" ? var.spec.txt_owner_id : null
      txtPrefix                 = var.spec.txt_prefix != "" ? var.spec.txt_prefix : null
      txtSuffix                 = var.spec.txt_suffix != "" ? var.spec.txt_suffix : null
      domainFilters             = length(var.spec.domain_filters) > 0 ? var.spec.domain_filters : null
      excludeDomains            = length(var.spec.exclude_domains) > 0 ? var.spec.exclude_domains : null
      annotationFilter          = var.spec.annotation_filter != "" ? var.spec.annotation_filter : null
      labelFilter               = var.spec.label_filter != "" ? var.spec.label_filter : null
      managedRecordTypes        = length(var.spec.managed_record_types) > 0 ? var.spec.managed_record_types : null
      interval                  = try(var.spec.interval, null)
      triggerLoopOnEvent        = var.spec.trigger_loop_on_event ? true : null
      namespaced                = var.spec.namespaced ? true : null
      logLevel                  = try(var.spec.log_level, null)
      logFormat                 = try(var.spec.log_format, null)

      # ---- pod placement / sizing ----------------------------------------
      resources    = local.controller_resources
      nodeSelector = length(var.spec.node_selector) > 0 ? var.spec.node_selector : null
      tolerations = length(var.spec.tolerations) > 0 ? [
        for t in var.spec.tolerations : {
          for tk, tv in {
            key               = t.key != "" ? t.key : null
            operator          = t.operator != "" ? t.operator : null
            value             = t.value != "" ? t.value : null
            effect            = t.effect != "" ? t.effect : null
            tolerationSeconds = try(t.toleration_seconds, null)
          } : tk => tv if tv != null
        }
      ] : null
      priorityClassName = var.spec.priority_class_name != "" ? var.spec.priority_class_name : null

      # ---- observability ---------------------------------------------------
      serviceMonitor = try(var.spec.prometheus.service_monitor, false) ? {
        for smk, smv in {
          enabled          = true
          interval         = try(var.spec.prometheus.service_monitor_interval, "") != "" ? var.spec.prometheus.service_monitor_interval : null
          additionalLabels = length(try(var.spec.prometheus.service_monitor_labels, {})) > 0 ? var.spec.prometheus.service_monitor_labels : null
        } : smk => smv if smv != null
      } : null

      # ---- image -------------------------------------------------------------
      image = (var.spec.image_repository != "" || var.spec.image_tag != "") ? {
        for ik, iv in {
          repository = var.spec.image_repository != "" ? var.spec.image_repository : null
          tag        = var.spec.image_tag != "" ? var.spec.image_tag : null
        } : ik => iv if iv != null
      } : null

      # ---- provider + its env/args/volumes -----------------------------------
      provider          = local.provider_block
      extraArgs         = length(local.provider_extra_args) > 0 ? local.provider_extra_args : null
      env               = length(local.provider_env) > 0 ? local.provider_env : null
      extraVolumes      = length(local.extra_volumes) > 0 ? local.extra_volumes : null
      extraVolumeMounts = length(local.extra_volume_mounts) > 0 ? local.extra_volume_mounts : null
    } : k => v if v != null && v != {}
  }
}
