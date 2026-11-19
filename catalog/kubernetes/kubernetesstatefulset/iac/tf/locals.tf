# Computed values shared across the module's resources.
#
# The load-bearing decisions here:
#  - selector_labels are the IMMUTABLE pod-selection identity (metadata.name
#    based) — the StatefulSet selector, the governing Service selector, and the
#    PDB selector all use exactly this set. Selectors are immutable on apps/v1
#    StatefulSets, so nothing mutable may ever join it.
#  - all_containers normalizes the app container (default name "app") and the
#    sidecars into ONE uniform list, so statefulset.tf renders every container
#    through a single dynamic block — identical semantics for app and sidecars.
#  - pod_volumes derives the pod-level volume list from the union of every
#    container's mounts, EXCLUDING mounts bound to volume claim templates — the
#    StatefulSet controller stamps one PVC per replica from each template and
#    binds it as a pod volume itself; declaring it again at the pod level would
#    collide with the controller-injected volume.

locals {
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  selector_labels = {
    "app"           = var.metadata.name
    "resource_name" = var.metadata.name
  }

  base_labels = merge(local.selector_labels, {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "KubernetesStatefulSet"
  })

  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_label = (
    var.metadata.env != null && try(var.metadata.env, "") != ""
  ) ? { "environment" = var.metadata.env } : {}

  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  namespace = var.spec.namespace

  # Satellite resource names, prefixed with metadata.name so multiple
  # instances sharing a namespace never collide.
  env_secret_name        = "${var.metadata.name}-env-secrets"
  image_pull_secret_name = "${var.metadata.name}-image-pull"

  # ---------------------------------------------------------------------------
  # Container normalization: app first (named "app" unless the spec names it),
  # then sidecars in declared order. Kubernetes shows containers in declaration
  # order and tooling conventionally treats the first as primary.
  # ---------------------------------------------------------------------------
  app_container = merge(var.spec.container.app, {
    name = try(var.spec.container.app.name, "") != "" ? var.spec.container.app.name : "app"
  })

  all_containers = concat([local.app_container], try(var.spec.container.sidecars, []))

  init_containers = try(var.spec.pod.init_containers, [])

  # Literal secret env values across ALL containers (app, sidecars, init),
  # materialized into one workload-scoped Secret. secretRef entries are wired
  # directly as env references and never pass through this map.
  env_secret_data = merge([
    for c in concat(local.all_containers, local.init_containers) : {
      for s in try(c.env.secrets, []) : s.name => s.value
      if try(s.value, "") != "" && try(s.secret_ref, null) == null
    }
  ]...)

  # Names of the volume claim templates — PVC mounts whose claim name matches
  # one of these are bound by the StatefulSet controller (one PVC per replica),
  # so they must NOT produce a pod-level volume.
  volume_claim_template_names = toset([
    for t in try(var.spec.volume_claim_templates, []) : t.name
  ])

  # Pod-level volumes: union of every container's mounts, first declaration of
  # a name wins (two containers sharing an EmptyDir declare the same mount name
  # and source). Mounts referencing a volume claim template contribute no pod
  # volume — the controller injects those per pod under the template's name.
  volume_mounts_flat = flatten([
    for c in concat(local.all_containers, local.init_containers) : try(c.volume_mounts, [])
  ])

  pod_volumes = {
    for vm in local.volume_mounts_flat : vm.name => vm...
    if !(try(vm.pvc, null) != null && contains(local.volume_claim_template_names, vm.pvc.claim_name)) && (
      try(vm.config_map, null) != null || try(vm.secret, null) != null || try(vm.host_path, null) != null ||
      try(vm.empty_dir, null) != null || try(vm.pvc, null) != null || try(vm.service_account_token, null) != null ||
      try(vm.cluster_trust_bundle, null) != null || try(vm.pod_certificate, null) != null
    )
  }

  # PARITY-EXCEPTION: the hashicorp/kubernetes provider models neither the
  # clusterTrustBundle nor the podCertificate projection, so this engine cannot
  # write either; the workload resource's precondition refuses them by name.
  # The Pulumi engine writes both (pkg/iac/pulumi/pulumimodule/provider/
  # kubernetes/workloadpod).
  unsupported_projected_mounts = distinct([
    for vm in local.volume_mounts_flat : vm.name
    if try(vm.cluster_trust_bundle, null) != null || try(vm.pod_certificate, null) != null
  ])

  # Registry logins the workload declares on pod.image_registries, rendered as the
  # .dockerconfigjson document of the module-owned image-pull Secret: one `auths`
  # record per server, with the base64 "user:password" pair the kubelet reads.
  # The same shape the kubernetessecret module's locals.tf renders for its docker
  # arm (and the Go twin, pkg/kubernetes/dockerconfigjson). The password arrives
  # already resolved from its $secret/ reference. A server named twice is refused
  # by the spec's own validation rule before any plan runs.
  image_registries = try(var.spec.pod.image_registries, [])

  image_pull_docker_config_json = jsonencode({
    auths = {
      for r in local.image_registries : r.server => merge(
        {
          username = r.username
          password = r.password
          auth     = base64encode("${r.username}:${r.password}")
        },
        try(r.email, "") != "" ? { email = r.email } : {}
      )
    }
  })

  create_image_pull_secret = length(local.image_registries) > 0

  # Pod-level image pull secrets: Secrets declared beside the workload and named in
  # pod.image_pull_secrets, plus the module-owned Secret when the pod declares
  # registries. ServiceAccount-attached pull secrets need no entry here.
  image_pull_secret_names = concat(
    try(var.spec.pod.image_pull_secrets, []),
    local.create_image_pull_secret ? [local.image_pull_secret_name] : []
  )

  # Pod template labels: controller labels win over user pod labels so a user
  # label can never break pod selection.
  pod_template_labels = merge(try(var.spec.pod.labels, {}), local.final_labels)

  # ---------------------------------------------------------------------------
  # Governing Service wiring. The headless Service ALWAYS exists — StatefulSets
  # require one for per-pod DNS regardless of whether the app exposes ports —
  # so every service-derived output is unconditionally populated. The port list
  # comes from the app container's ports; service_port defaults to
  # container_port so the common "expose as-is" case needs no extra config.
  # ---------------------------------------------------------------------------
  app_ports = try(var.spec.container.app.ports, [])

  service_ports = [
    for p in local.app_ports : {
      name         = p.name
      protocol     = try(p.network_protocol, "") != "" ? p.network_protocol : "TCP"
      port         = try(p.service_port, 0) > 0 ? p.service_port : p.container_port
      target_port  = p.container_port
      app_protocol = try(p.app_protocol, "") != "" ? p.app_protocol : null
    }
  ]

  kube_service_name = var.metadata.name
  kube_service_fqdn = "${local.kube_service_name}.${local.namespace}.svc.cluster.local"

  kube_port_forward_command = "kubectl port-forward -n ${local.namespace} service/${local.kube_service_name} 8080:8080"

  # Per-replica DNS template: {ordinal} is a literal placeholder the consumer
  # substitutes with the replica index (e.g. "0") to address a specific
  # member — this is how clustered clients build their member lists.
  pod_dns_template = "${var.metadata.name}-{ordinal}.${local.kube_service_name}.${local.namespace}.svc.cluster.local"

  # Selector labels rendered as a deterministic sorted "k=v,k=v" string — the
  # exact syntax kubectl -l and NetworkPolicy tooling accept.
  selector_labels_string = join(",", [for k in sort(keys(local.selector_labels)) : "${k}=${local.selector_labels[k]}"])

  # Replica count. Scaling stateful members is application-aware work (data
  # sync, quorum changes), so there is deliberately no inline HPA on this kind.
  # Unset, Kubernetes or an external autoscaler owns the size: null keeps the
  # live count (replicas is optional and computed on kubernetes_stateful_set_v1),
  # so an apply never moves it. Pulumi twin: the workloadreplicas package.
  statefulset_replicas = try(var.spec.availability.replicas, null)

  update_strategy_type = try(var.spec.update_strategy.type, "") != "" ? var.spec.update_strategy.type : "RollingUpdate"

  pdb_enabled = try(var.spec.availability.pod_disruption_budget.enabled, false)
}
