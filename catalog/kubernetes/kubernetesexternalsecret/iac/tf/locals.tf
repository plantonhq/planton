# Computed values for the KubernetesExternalSecret module.
#
# The CR spec rendered here is the Terraform twin of the Pulumi module's
# spec_builder.go — keep field mappings in lockstep: same CRD field names,
# same emit-only-when-set conditions, same upstream-default omissions
# (decodingStrategy "None", template mergePolicy "Replace").
#
# Optional nested blocks are read with try(): HCL's && does NOT
# short-circuit, so chained null checks still dereference the null.

locals {
  # ExternalSecret metadata.name.
  external_secret_name = var.metadata.name

  # Namespace the ExternalSecret (and its materialized Secret) lives in
  # (resolved literal from the spec's value-or-ref).
  namespace = var.spec.namespace

  # Name of the Kubernetes Secret the operator materializes: target.name
  # when set, else metadata.name (upstream's own default). Exported —
  # workloads wire env/volume references to THIS name.
  secret_name = try(var.spec.target.name, "") != "" ? var.spec.target.name : var.metadata.name

  labels = merge(concat(
    [{
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesExternalSecret"
    }],
    (var.metadata.id != null && var.metadata.id != "") ? [{ "planton.ai/resource-id" = var.metadata.id }] : [],
    (var.metadata.org != null && var.metadata.org != "") ? [{ "planton.ai/organization" = var.metadata.org }] : [],
    (var.metadata.env != null && var.metadata.env != "") ? [{ "planton.ai/environment" = var.metadata.env }] : []
  )...)

  # ---- target Secret ------------------------------------------------------
  # The template renders only when it carries content — the twin of the Go
  # builder's buildTemplate returning nil when nothing is set.
  template_has_content = try(var.spec.target.template, null) != null && (
    try(var.spec.target.template.type, "") != "" ||
    (try(var.spec.target.template.merge_policy, null) != null && try(var.spec.target.template.merge_policy, "") != "Replace") ||
    length(try(var.spec.target.template.labels, {})) > 0 ||
    length(try(var.spec.target.template.annotations, {})) > 0 ||
    length(try(var.spec.target.template.data, {})) > 0
  )

  # Always rendered: the materialized Secret's name is pinned to the
  # resolved local.secret_name so the exported handle can never drift from
  # what the operator creates. Null-prune idiom throughout this file: one
  # object literal whose conditional entries are null when absent,
  # filtered by the for-expression — heterogeneous conditional merges
  # (`cond ? {...} : {}` / concat-spread) fail HCL type unification when
  # sibling entries infer as different object types.
  external_secret_target = {
    for k, v in {
      name           = local.secret_name
      creationPolicy = try(var.spec.target.creation_policy, null)
      deletionPolicy = try(var.spec.target.deletion_policy, null)
      # immutable is emitted only when true — false is the API default,
      # and omitting it keeps the applied object byte-identical with the
      # Pulumi engine's.
      immutable = try(var.spec.target.immutable, false) ? true : null
      template = !local.template_has_content ? null : {
        for tk, tv in {
          type = try(var.spec.target.template.type, "") != "" ? var.spec.target.template.type : null
          # mergePolicy "Replace" is the upstream default — omitted, like
          # the Go builder.
          mergePolicy = (try(var.spec.target.template.merge_policy, null) != null && try(var.spec.target.template.merge_policy, "") != "Replace") ? var.spec.target.template.merge_policy : null
          metadata = (length(try(var.spec.target.template.labels, {})) > 0 || length(try(var.spec.target.template.annotations, {})) > 0) ? {
            for mk, mv in {
              labels      = length(try(var.spec.target.template.labels, {})) > 0 ? var.spec.target.template.labels : null
              annotations = length(try(var.spec.target.template.annotations, {})) > 0 ? var.spec.target.template.annotations : null
            } : mk => mv if mv != null
          } : null
          data = length(try(var.spec.target.template.data, {})) > 0 ? var.spec.target.template.data : null
        } : tk => tv if tv != null
      }
    } : k => v if v != null
  }

  # ---- explicit entries ---------------------------------------------------
  # remoteRef rendering is the twin of the Go builder's buildRemoteRef:
  # decodingStrategy "None" is the upstream default — omitted.
  external_secret_data = [
    for entry in try(var.spec.data, []) : {
      secretKey = entry.secret_key
      remoteRef = {
        for rk, rv in {
          key              = entry.remote_ref.key
          property         = try(entry.remote_ref.property, "") != "" ? entry.remote_ref.property : null
          version          = try(entry.remote_ref.version, "") != "" ? entry.remote_ref.version : null
          decodingStrategy = (try(entry.remote_ref.decoding_strategy, null) != null && try(entry.remote_ref.decoding_strategy, "") != "None") ? entry.remote_ref.decoding_strategy : null
        } : rk => rv if rv != null
      }
    }
  ]

  # ---- bulk pulls -----------------------------------------------------------
  external_secret_data_from = [
    for pull in try(var.spec.data_from, []) : {
      for k, v in {
        extract = try(pull.extract, null) == null ? null : {
          for rk, rv in {
            key              = pull.extract.key
            property         = try(pull.extract.property, "") != "" ? pull.extract.property : null
            version          = try(pull.extract.version, "") != "" ? pull.extract.version : null
            decodingStrategy = (try(pull.extract.decoding_strategy, null) != null && try(pull.extract.decoding_strategy, "") != "None") ? pull.extract.decoding_strategy : null
          } : rk => rv if rv != null
        }
        find = try(pull.find, null) == null ? null : {
          for fk, fv in {
            path = try(pull.find.path, "") != "" ? pull.find.path : null
            name = try(pull.find.name_regexp, "") != "" ? { regexp = pull.find.name_regexp } : null
            tags = length(try(pull.find.tags, {})) > 0 ? pull.find.tags : null
          } : fk => fv if fv != null
        }
        rewrite = length(try(pull.rewrite, [])) > 0 ? [
          for rw in pull.rewrite : {
            regexp = {
              source = rw.source
              target = rw.target
            }
          }
        ] : null
      } : k => v if v != null
    }
  ]

  # ---- the CR spec ----------------------------------------------------------
  external_secret_spec = {
    for k, v in {
      # The arm is the grain: a cluster store renders its kind, a namespaced
      # store renders only its name (SecretStore is the CRD's default).
      secretStoreRef = {
        for rk, rv in {
          name = try(var.spec.store_ref.cluster_secret_store.name, var.spec.store_ref.secret_store.name)
          kind = try(var.spec.store_ref.cluster_secret_store.name, null) != null ? "ClusterSecretStore" : null
        } : rk => rv if rv != null
      }
      refreshInterval = try(var.spec.refresh_interval, null)
      # refreshPolicy renders only when non-empty — empty means upstream
      # behavior (periodic per refreshInterval), like the Go builder.
      refreshPolicy = (try(var.spec.refresh_policy, null) != null && try(var.spec.refresh_policy, "") != "") ? var.spec.refresh_policy : null
      target        = local.external_secret_target
      data          = length(try(var.spec.data, [])) > 0 ? local.external_secret_data : null
      dataFrom      = length(try(var.spec.data_from, [])) > 0 ? local.external_secret_data_from : null
    } : k => v if v != null
  }
}
