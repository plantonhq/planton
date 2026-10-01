# Computed values for the KubernetesKeycloakOperator module. Every
# resolution here has an exact twin in the Pulumi module — keep them in
# lockstep: same namespace stamping, same patched Deployment, same
# outputs.
#
# HCL DISCIPLINE: conditional keys are contributed as merge() of
# `cond ? { key = value } : {}` singleton maps — a ternary whose branches
# are differently-shaped objects fails plan-time type unification.
# Optional nested blocks are read with try(): HCL's && does NOT
# short-circuit.

locals {
  # Pinned keycloak/keycloak-k8s-resources release tag.
  #
  # MUST stay in sync with the Pulumi module's BundleRelease constant.
  # There is NO user-facing version field BY DESIGN: the
  # KubernetesKeycloak declaration kind's CR rendering is built against
  # the CRD schema this bundle installs — a selectable operator version
  # would drift the schema away from what the declaration kind renders.
  # The module pins the release; upgrades arrive as module updates.
  # Always an exact release TAG, never a branch — tag pinning keeps
  # installs reproducible.
  bundle_release = "26.7.0"

  # Raw-content root of the pinned tag — keycloak-k8s-resources
  # publishes NO single-file release asset; the tagged tree's
  # kubernetes/ directory IS the official distribution (Keycloak ships
  # no official Helm chart either; the operator is the first-party
  # Kubernetes distribution).
  bundle_base_url = "https://raw.githubusercontent.com/keycloak/keycloak-k8s-resources/${local.bundle_release}/kubernetes"

  # The 16-document operator bundle for the requested watch scope:
  # kubernetes.yml (the operator watches ONLY its own namespace —
  # JOSDK_WATCH_CURRENT) or cluster-wide/kubernetes.yml (per-controller
  # ClusterRoleBindings and JOSDK_ALL_NAMESPACES).
  bundle_url = (
    try(var.spec.cluster_wide, false) ?
    "${local.bundle_base_url}/cluster-wide/kubernetes.yml" :
    "${local.bundle_base_url}/kubernetes.yml"
  )

  # The four k8s.keycloak.org CRD files published beside the bundle
  # (SHARED by both watch-scope variants; each carries exactly one CRD
  # document plus a generator comment header).
  crd_files = [
    "keycloaks.k8s.keycloak.org-v1.yml",
    "keycloakrealmimports.k8s.keycloak.org-v1.yml",
    "keycloakoidcclients.k8s.keycloak.org-v1.yml",
    "keycloaksamlclients.k8s.keycloak.org-v1.yml",
  ]

  # Namespace the operator installs into (the value-or-ref foreign key
  # resolves to a literal before Terraform runs) — stamped onto every
  # namespaced bundle document below.
  namespace = var.spec.namespace

  # Fixed handles from the bundle (upstream's own `keycloak-operator`
  # names, not derived from this resource's name). Fixed names mean
  # exactly ONE operator install fits per namespace.
  deployment_name = "keycloak-operator"
  service_name    = "keycloak-operator"

  # Resource-identity labels stamped on the namespace this module
  # creates (never injected into the bundle's own documents — faithful
  # apply).
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesKeycloakOperator"
    },
    var.metadata.id != null && var.metadata.id != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    var.metadata.org != null && var.metadata.org != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # ---- manifest documents ------------------------------------------------------
  # The bundle splits into documents on `---` separator LINES — the
  # bundle's separators all sit at column 0 with nothing else on the
  # line, so the "\n---\n" boundary never matches inside a document.
  # The CRD files are fetched separately and are single-document (their
  # generator comment headers are YAML comments yamldecode ignores), so
  # they decode whole. Same document set as the Pulumi module's
  # fetchManifestPartitions.
  bundle_documents_raw = [
    for doc in split("\n---\n", data.http.keycloak_operator_bundle.response_body) :
    yamldecode(doc)
    if trimspace(doc) != "" && can(yamldecode(doc).kind)
  ]

  crd_documents_raw = [
    for file in local.crd_files :
    yamldecode(data.http.keycloak_operator_crds[file].response_body)
  ]

  all_documents_raw = concat(local.bundle_documents_raw, local.crd_documents_raw)

  # ---- namespace stamping (the module owns upstream's kustomize step) ----------
  # The bundle ships every document WITHOUT a namespace field (upstream
  # expects kustomize to set it). Two rewrites, mirroring the Pulumi
  # module's namespaceTransformation:
  #
  # 1. metadata.namespace = <spec namespace> on every NAMESPACED
  #    document. The namespaced kinds in THIS bundle are listed below;
  #    ClusterRole, ClusterRoleBinding and the CRDs are cluster-scoped
  #    and get none. The set is bundle-specific and pinned with the
  #    release.
  # 2. Every RoleBinding AND ClusterRoleBinding ServiceAccount subject
  #    gets namespace = <spec namespace>. Upstream bakes
  #    `namespace: keycloak` into exactly ONE ClusterRoleBinding subject
  #    and leaves the rest empty for kustomize — both wrong for a
  #    configurable namespace, so ALL subjects are rewritten.
  namespaced_kinds = ["ServiceAccount", "Role", "RoleBinding", "Service", "Deployment"]

  stamped_documents = [
    for doc in local.all_documents_raw :
    merge(
      doc,
      contains(local.namespaced_kinds, doc.kind) ? {
        metadata = merge(doc.metadata, { namespace = local.namespace })
      } : {},
      contains(["RoleBinding", "ClusterRoleBinding"], doc.kind) ? {
        subjects = [
          for subject in try(doc.subjects, []) :
          merge(subject, subject.kind == "ServiceAccount" ? { namespace = local.namespace } : {})
        ]
      } : {}
    )
  ]

  # Each document keyed by its COMPOSED IDENTITY
  # `apiVersion//kind//name[//namespace]` — the exact ID form the kubectl
  # importer takes, so state addresses stay stable across manifest
  # reorderings AND the address key feeds the composed import ID blind.
  # Names REPEAT across kinds in this bundle (ServiceAccount, Service
  # and Deployment are all named `keycloak-operator`), so the kind
  # segment is what keeps the keys unique. Keyed AFTER stamping so
  # namespaced documents carry their 4-part keys; cluster-scoped
  # documents render 3-part keys (the importer rejects a trailing
  # `//`).
  documents_by_id = {
    for doc in local.stamped_documents :
    join("//", concat(
      [doc.apiVersion, doc.kind, doc.metadata.name],
      try(doc.metadata.namespace, "") != "" ? [doc.metadata.namespace] : []
    )) => doc
  }

  # ---- typed overrides for the operator Deployment --------------------------------
  # operator_image is a full image reference (the tag must stay at the
  # pinned release or the CRD schema drifts); empty keeps the bundle's
  # quay.io/keycloak/keycloak-operator at the pin.
  operator_image = try(var.spec.operator_image, "")

  # default_keycloak_image rewrites the RELATED_IMAGE_KEYCLOAK env value
  # — the DEFAULT Keycloak server image the operator stamps into
  # Keycloak StatefulSets whose declaration sets no image; empty keeps
  # the bundle's quay.io/keycloak/keycloak at the pin.
  default_keycloak_image = try(var.spec.default_keycloak_image, "")

  # Operator container resources; null keeps the bundle's own values
  # (requests 300m/450Mi, limits 700m/450Mi — identical to the spec's
  # proto defaults, so an unset spec drifts nothing). Pulumi twin:
  # resourcesMap. The inner prune can leave an EMPTY object when the
  # converter emits a present-but-empty resources block — normalized
  # back to null so `resources: {}` never strips upstream's values
  # (Pulumi's resourcesMap returns nil on the same input).
  operator_resources_raw = try(var.spec.resources, null) == null ? null : {
    for k, v in {
      requests = try(var.spec.resources.requests, null) == null ? null : {
        for rk, rv in {
          cpu    = var.spec.resources.requests.cpu
          memory = var.spec.resources.requests.memory
        } : rk => rv if rv != null && rv != ""
      }
      limits = try(var.spec.resources.limits, null) == null ? null : {
        for rk, rv in {
          cpu    = var.spec.resources.limits.cpu
          memory = var.spec.resources.limits.memory
        } : rk => rv if rv != null && rv != ""
      }
    } : k => v if v != null && length(v) > 0
  }
  operator_resources = local.operator_resources_raw == null ? null : (length(local.operator_resources_raw) > 0 ? local.operator_resources_raw : null)

  # Pod scheduling constraints (spec.scheduling; upstream ships none).
  node_selector = try(var.spec.scheduling.node_selector, {})

  pod_tolerations = [
    for t in try(var.spec.scheduling.tolerations, []) : { for k, v in {
      key               = t.key != "" ? t.key : null
      operator          = t.operator != "" ? t.operator : null
      value             = t.value != "" ? t.value : null
      effect            = t.effect != "" ? t.effect : null
      tolerationSeconds = t.toleration_seconds
    } : k => v if v != null }
  ]

  pod_spec_overrides = merge(
    length(local.node_selector) > 0 ? { nodeSelector = local.node_selector } : {},
    length(local.pod_tolerations) > 0 ? { tolerations = local.pod_tolerations } : {}
  )

  # ---- the patched Deployment ------------------------------------------------------
  # The spec's typed overrides land on the single operator container
  # (image, resources, the RELATED_IMAGE_KEYCLOAK env value) and the pod
  # spec (nodeSelector / tolerations) — every other document applies
  # verbatim (faithful distribution). Pulumi twin:
  # deploymentTransformation.
  operator_deployment_id = "apps/v1//Deployment//${local.deployment_name}//${local.namespace}"

  original_operator_deployment = local.documents_by_id[local.operator_deployment_id]
  patched_operator_deployment = merge(
    local.original_operator_deployment,
    {
      spec = merge(
        local.original_operator_deployment.spec,
        {
          template = merge(
            local.original_operator_deployment.spec.template,
            {
              spec = merge(
                local.original_operator_deployment.spec.template.spec,
                {
                  containers = [
                    for c in local.original_operator_deployment.spec.template.spec.containers :
                    merge(
                      c,
                      local.operator_image != "" ? { image = local.operator_image } : {},
                      local.operator_resources != null ? { resources = local.operator_resources } : {},
                      local.default_keycloak_image != "" ? {
                        env = [
                          for e in try(c.env, []) :
                          merge(e, e.name == "RELATED_IMAGE_KEYCLOAK" ? { value = local.default_keycloak_image } : {})
                        ]
                      } : {}
                    )
                  ]
                },
                local.pod_spec_overrides
              )
            }
          )
        }
      )
    }
  )

  # The applied set: every document verbatim, with the operator
  # Deployment swapped for its patched form. merge() — never a
  # per-element ternary: a ternary whose branches are differently-shaped
  # objects fails HCL plan-time type unification.
  applied_documents = merge(
    local.documents_by_id,
    {
      (local.operator_deployment_id) = local.patched_operator_deployment
    }
  )

  # The bundle partitions into THREE ordered groups because for_each
  # instances apply — and destroy — in PARALLEL, and release-manifest
  # bundles have two structural ordering hazards:
  #
  # 1. CREATE: a namespaced document landing before the Namespace fails
  #    with "namespaces ... not found".
  # 2. DESTROY: CRD deletion cascade-deletes every Keycloak CR on the
  #    cluster, and any operator-processed finalizers on those CRs need
  #    the LIVE operator to drain — deleting the CRDs and the operator
  #    in one flat pass risks wedging the drain in Terminating until the
  #    provider's delete await times out (the tektonoperator exemplar
  #    caught this class live). The CRDs must therefore delete FIRST,
  #    while the operator still runs.
  #
  # Terraform destroy runs the reverse of create ordering, so the
  # dependency chain namespace ← workloads ← crds yields create
  # namespace→workloads→crds and destroy crds→workloads→namespace —
  # both hazards resolved structurally. The operator tolerates starting
  # before its CRDs exist (the JOSDK operator crash-loops until they
  # appear), and with no admission webhook nothing it defines is refused
  # while it starts, so nothing waits for it.
  #
  # The Namespace document is MODULE-AUTHORED (the bundle ships none):
  # rendered only when create_namespace is set, carrying the standard
  # Planton governance labels — the same document the Pulumi module's
  # namespaceDocumentYaml builds, up to each engine's fleet-conventional
  # label KEYS (TF planton.ai/resource-* vs Pulumi planton.ai/name|kind|
  # id; the program-wide divergence, not this module's). When
  # create_namespace is false the namespace must already exist and the
  # module never touches it.
  namespace_documents = try(var.spec.create_namespace, false) ? {
    "v1//Namespace//${local.namespace}" = {
      apiVersion = "v1"
      kind       = "Namespace"
      metadata = {
        name   = local.namespace
        labels = local.labels
      }
    }
  } : {}

  crd_documents = {
    for k, v in local.applied_documents : k => v
    if try(v.kind, "") == "CustomResourceDefinition"
  }
  workload_documents = {
    for k, v in local.applied_documents : k => v
    if try(v.kind, "") != "CustomResourceDefinition"
  }
}
