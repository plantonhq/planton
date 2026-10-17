# The patches applied onto upstream's own objects, and the document set the
# module applies. Each patch has an exact twin in the Pulumi module's
# workloads.go and compose.go; the parity test holds both engines' documents
# equal for the same specs.

locals {
  # ---- the controller Deployment (patchControllerDeployment) --------------------
  controller_container_patched = local.controller_upstream_container == null ? null : merge(
    local.controller_upstream_container,
    {
      image = local.controller_image
      args  = local.controller_args
      ports = [
        for p in try(local.controller_upstream_container.ports, []) :
        merge(p, p.name == "metrics" ? { containerPort = local.c_metrics_port } : {}, p.name == "healthz" ? { containerPort = local.c_health_port } : {})
      ]
      # Upstream's manifest wires no probes to the controller's own health
      # endpoints; without a readiness probe a finished rollout only means
      # the container started. The timings are controller-runtime's
      # scaffolding defaults.
      livenessProbe  = { httpGet = { path = "/healthz", port = "healthz" }, initialDelaySeconds = 15, periodSeconds = 20 }
      readinessProbe = { httpGet = { path = "/readyz", port = "healthz" }, initialDelaySeconds = 5, periodSeconds = 10 }
    },
    local.c_resources != null ? { resources = local.c_resources } : {},
    length(local.controller_env) > 0 ? { env = local.controller_env } : {},
    local.c_metrics_tls_cert != "" ? {
      volumeMounts = concat(try(local.controller_upstream_container.volumeMounts, []), [{ name = "metrics-tls", mountPath = local.metrics_tls_mount_path, readOnly = true }])
    } : {},
  )

  controller_deployment_patched = local.controller_upstream == null ? null : merge(local.controller_upstream, {
    metadata = merge(local.controller_upstream.metadata, {
      labels = merge(try(local.controller_upstream.metadata.labels, {}), { (local.owner_label) = local.kind_name })
    })
    spec = merge(local.controller_upstream.spec, {
      replicas = local.c_replicas
      template = merge(local.controller_upstream.spec.template,
        length(local.allowed_label_domains) > 0 ? {
          metadata = merge(try(local.controller_upstream.spec.template.metadata, {}), {
            annotations = merge(try(local.controller_upstream.spec.template.metadata.annotations, {}), { (local.config_checksum_annotation) = sha256(local.config_map_content) })
          })
        } : {},
        {
          spec = merge(
            local.controller_upstream.spec.template.spec,
            # Upstream's pod has the one container the module patches; any
            # other would keep its place before it (a conditional cannot
            # carry two differently typed containers).
            {
              containers = concat(
                [for c in local.controller_upstream.spec.template.spec.containers : c if c.name != local.controller_container_name],
                [local.controller_container_patched],
              )
            },
            local.c_metrics_tls_cert != "" ? {
              volumes = concat(try(local.controller_upstream.spec.template.spec.volumes, []), [{ name = "metrics-tls", secret = { secretName = local.c_metrics_tls_cert } }])
            } : {},
            length(local.c_node_selector) > 0 ? { nodeSelector = local.c_node_selector } : {},
            length(local.c_tolerations) > 0 ? { tolerations = local.c_tolerations } : {},
            length(local.controller_pull_secrets) > 0 ? { imagePullSecrets = local.controller_pull_secrets } : {},
          )
      })
    })
  })

  controller_service = try(local.documents[local.controller_service_id], null)
  controller_service_patched = local.controller_service == null ? null : merge(local.controller_service, {
    spec = merge(local.controller_service.spec, {
      ports = [for p in local.controller_service.spec.ports : p.name == "metrics" ? merge(p, { port = local.c_metrics_port }) : p]
    })
  })

  # The controller's runtime configuration (configMap), applied only when the
  # extensions run and the spec lists label domains.
  config_map = {
    apiVersion = "v1"
    kind       = "ConfigMap"
    metadata   = { name = local.config_map_name, namespace = local.namespace }
    data       = { (local.config_map_key) = local.config_map_content }
  }

  # ---- the router (patchRouterDeployment, patchRouterService, ...) --------------
  router_proxy_tls_port = { name = "proxy-tls", containerPort = local.r_tls_port, protocol = "TCP" }
  router_container_patched = local.router_upstream_container == null ? null : merge(
    local.router_upstream_container,
    {
      image = local.router_image
      args  = local.router_args
      ports = concat(
        local.r_http_enabled ? [for p in local.router_upstream_container.ports : merge(p, { containerPort = local.r_http_port }) if p.name == "proxy"] : [],
        local.r_tls_on ? [local.router_proxy_tls_port] : [],
        [for p in local.router_upstream_container.ports : merge(p, { containerPort = p.name == "metrics" ? local.r_metrics_port : local.r_health_port }) if contains(["metrics", "healthz"], p.name)],
        [for p in local.router_upstream_container.ports : p if !contains(["proxy", "metrics", "healthz"], p.name)],
      )
    },
    local.r_resources != null ? { resources = local.r_resources } : {},
    length(local.router_env) > 0 ? { env = local.router_env } : {},
    length(local.router_mounts) > 0 ? { volumeMounts = concat(try(local.router_upstream_container.volumeMounts, []), local.router_mounts) } : {},
  )
  router_secret_mounts = concat(
    local.r_tls_on ? [{ volume = "tls-certs", secret = local.r_tls_cert, path = local.router_tls_mount }] : [],
    local.r_client_ca != null ? [{ volume = "client-ca", secret = local.r_client_ca.secret, path = local.router_client_ca_mount }] : [],
    local.r_scoped_v1 != null ? [{ volume = "scoped-token-v1", secret = local.r_scoped_v1.secret, path = local.router_scoped_v1_mount }] : [],
    local.r_scoped_v2 != null ? [{ volume = "scoped-token-v2", secret = local.r_scoped_v2.secret, path = local.router_scoped_v2_mount }] : [],
  )
  router_mounts  = [for m in local.router_secret_mounts : { name = m.volume, mountPath = m.path, readOnly = true }]
  router_volumes = [for m in local.router_secret_mounts : { name = m.volume, secret = { secretName = m.secret } }]

  router_deployment_patched = local.router_upstream == null ? null : merge(local.router_upstream, {
    metadata = merge(local.router_upstream.metadata, {
      labels = merge(try(local.router_upstream.metadata.labels, {}), { (local.owner_label) = local.kind_name })
    })
    spec = merge(local.router_upstream.spec, {
      replicas = local.r_replicas
      template = merge(local.router_upstream.spec.template, {
        spec = merge(
          local.router_upstream.spec.template.spec,
          {
            containers = concat(
              [for c in local.router_upstream.spec.template.spec.containers : c if c.name != local.router_container_name],
              [local.router_container_patched],
            )
          },
          length(local.router_volumes) > 0 ? { volumes = concat(try(local.router_upstream.spec.template.spec.volumes, []), local.router_volumes) } : {},
          local.r_shutdown_timeout != "30s" ? { terminationGracePeriodSeconds = local.router_termination_grace } : {},
          length(local.r_node_selector) > 0 ? { nodeSelector = local.r_node_selector } : {},
          length(local.r_tolerations) > 0 ? { tolerations = local.r_tolerations } : {},
          length(local.router_pull_secrets) > 0 ? { imagePullSecrets = local.router_pull_secrets } : {},
        )
      })
    })
  })

  router_service = try(local.documents[local.router_service_id], null)
  router_service_patched = local.router_service == null ? null : merge(local.router_service, {
    spec = merge(local.router_service.spec, {
      ports = concat(
        [for p in local.router_service.spec.ports : p.name == "proxy" ? merge(p, { port = local.r_http_port }) : p if p.name != "proxy" || local.r_http_enabled],
        local.r_tls_on ? [{ name = "proxy-tls", port = local.r_tls_port, targetPort = "proxy-tls", protocol = "TCP" }] : [],
      )
    })
  })

  router_pdb = try(local.documents[local.router_pdb_id], null)
  router_pdb_patched = local.router_pdb == null ? null : merge(local.router_pdb, {
    spec = merge(local.router_pdb.spec, { minAvailable = local.r_min_available })
  })

  # The router's NetworkPolicy: who may call it, which namespace scrapes it,
  # which sandbox ports it may open. The DNS, collector and API-server rules
  # stay upstream's.
  router_netpol_patched = local.router_netpol_upstream == null ? null : merge(local.router_netpol_upstream, {
    spec = merge(local.router_netpol_upstream.spec, {
      ingress = [
        merge(local.router_netpol_upstream.spec.ingress[0], {
          from = [{ namespaceSelector = length(local.r_client_ns_labels) > 0 ? { matchLabels = local.r_client_ns_labels } : {} }]
          ports = concat(
            local.r_http_enabled ? [{ protocol = "TCP", port = local.r_http_port }] : [],
            local.r_tls_on ? [{ protocol = "TCP", port = local.r_tls_port }] : [],
          )
        }),
        merge(local.router_netpol_upstream.spec.ingress[1], {
          from  = [{ namespaceSelector = { matchLabels = { "kubernetes.io/metadata.name" = local.r_metrics_namespace } } }]
          ports = [{ protocol = "TCP", port = local.r_metrics_port }]
        }),
        merge(local.router_netpol_upstream.spec.ingress[2], {
          ports = [{ protocol = "TCP", port = local.r_health_port }]
        }),
      ]
      egress = [
        local.router_netpol_upstream.spec.egress[0],
        merge(local.router_netpol_upstream.spec.egress[1], {
          ports = [for p in local.r_sandbox_ports : { protocol = "TCP", port = p }]
        }),
        local.router_netpol_upstream.spec.egress[2],
        local.router_netpol_upstream.spec.egress[3],
      ]
    })
  })

  # ---- the applied set (compose) ------------------------------------------------
  # Every upstream document, the patched ones swapped in with merge(), never a
  # per-element ternary (differently shaped objects fail plan-time type
  # unification), plus the module's ConfigMap.
  applied_documents = merge(
    local.documents,
    local.controller_deployment_patched != null ? { (local.controller_deployment_id) = local.controller_deployment_patched } : {},
    local.controller_service_patched != null ? { (local.controller_service_id) = local.controller_service_patched } : {},
    local.extensions && length(local.allowed_label_domains) > 0 ? { (local.config_map_id) = local.config_map } : {},
    local.router_deployment_patched != null ? { (local.router_deployment_id) = local.router_deployment_patched } : {},
    local.router_service_patched != null ? { (local.router_service_id) = local.router_service_patched } : {},
    local.router_pdb_patched != null ? { (local.router_pdb_id) = local.router_pdb_patched } : {},
    local.router_netpol_patched != null ? { (local.router_netpol_id) = local.router_netpol_patched } : {},
  )
  # The Namespace only when this resource owns it, carrying the identity
  # labels; otherwise its owner (a KubernetesNamespace) created it already.
  namespace_upstream = try(local.documents[local.namespace_id], null)
  namespace_documents = local.create_namespace && local.namespace_upstream != null ? {
    (local.namespace_id) = merge(local.namespace_upstream, {
      metadata = merge(local.namespace_upstream.metadata, {
        labels = merge(try(local.namespace_upstream.metadata.labels, {}), local.identity_labels)
      })
    })
  } : {}
  workload_documents = { for k, v in local.applied_documents : k => v if v.kind != "Namespace" }

  # ---- the namespace check (workloadowner) ---------------------------------------
  # Owned elsewhere when this resource would create a namespace someone else
  # did; missing when it would not create one and nobody has; created here
  # when create_namespace was turned off on a live installation, which would
  # delete the namespace this resource holds in state.
  namespace_object       = try(data.kubernetes_resources.namespace.objects[0], null)
  namespace_ours         = try(local.namespace_object.metadata.labels[local.owner_label], "") == local.kind_name
  namespace_missing      = !local.create_namespace && local.namespace_object == null
  namespace_foreign      = local.create_namespace && local.namespace_object != null && !local.namespace_ours
  namespace_created_here = !local.create_namespace && local.namespace_object != null && local.namespace_ours
  namespace_managers     = try([for m in local.namespace_object.metadata.managedFields : m.manager], [])
  namespace_owner = (
    length(local.namespace_managers) > 0
    ? "last written by field manager(s) ${join(", ", local.namespace_managers)}"
    : "it carries no ownership marks"
  )

  # ---- the workload ownership check (workloadowner) ------------------------------
  # Deployments the module applies under upstream's fixed names, which no
  # other installation may already hold.
  owned_deployment_names = concat([local.controller_deployment_name], local.router_on ? [local.router_deployment_name] : [])
  foreign_deployments = {
    for name, read in data.kubernetes_resources.owned_deployment :
    name => (
      length(try(read.objects[0].metadata.managedFields, [])) > 0
      ? "last written by field manager(s) ${join(", ", [for m in read.objects[0].metadata.managedFields : m.manager])}"
      : "it carries no ownership marks"
    )
    if length(read.objects) > 0 && try(read.objects[0].metadata.labels[local.owner_label], "") != local.kind_name
  }

  # ---- the kept definitions (the generated helm_crds.tf) -------------------------
  # The bundle branch: the same release asset, CRDs only, kept and
  # ownership-checked; the module pins the release, so no spec field chooses
  # the version.
  helm_crds_args = {
    install           = try(var.spec.crds.install, null) == null ? true : var.spec.crds.install
    keep_on_uninstall = try(var.spec.crds.keep_on_uninstall, null) == null ? true : var.spec.crds.keep_on_uninstall
    expect_crds       = true
    version           = local.release
    version_field     = ""
    chart             = null
    bundle = {
      name = local.bundle_name
      url  = "${local.release_asset_base_url}{{version}}/${local.release_asset}"
    }
  }
}
