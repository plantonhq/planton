locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base labels
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "kubernetes_openbao"
  }

  # Organization label only if var.metadata.org is non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  # Environment label only if var.metadata.env is non-empty
  env_label = (
    var.metadata.env != null &&
    try(var.metadata.env, "") != ""
  ) ? { "environment" = var.metadata.env } : {}

  # Merge base, org, and environment labels
  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  # Get namespace from spec
  namespace = var.spec.namespace

  # Helm chart configuration
  helm_chart_name    = "openbao"
  helm_chart_repo    = "https://openbao.github.io/openbao-helm"
  helm_chart_version = coalesce(var.spec.helm_chart_version, "0.23.3")

  # OpenBao ports
  openbao_port         = 8200
  openbao_cluster_port = 8201

  # Service name (uses release name)
  kube_service_name = var.metadata.name

  # Internal DNS name for the service
  kube_service_fqdn = "${local.kube_service_name}.${local.namespace}.svc.cluster.local:${local.openbao_port}"

  # API address
  api_address = "http://${local.kube_service_name}.${local.namespace}.svc.cluster.local:${local.openbao_port}"

  # Cluster address (HA mode)
  cluster_address = "https://${local.kube_service_name}-0.${local.kube_service_name}-internal.${local.namespace}.svc.cluster.local:${local.openbao_cluster_port}"

  # Port-forward command
  kube_port_forward_command = "kubectl port-forward -n ${local.namespace} service/${local.kube_service_name} ${local.openbao_port}:${local.openbao_port}"

  # HA configuration
  ha_enabled  = try(var.spec.high_availability.enabled, false)
  ha_replicas = local.ha_enabled ? coalesce(try(var.spec.high_availability.replicas, null), 3) : 1

  # Server replicas
  server_replicas = var.spec.server_container.replicas

  # UI enabled (default true)
  ui_enabled = coalesce(var.spec.ui_enabled, true)

  # TLS configuration
  tls_enabled = coalesce(var.spec.tls_enabled, false)

  # Injector configuration
  injector_enabled  = try(var.spec.injector.enabled, false)
  injector_replicas = local.injector_enabled ? coalesce(try(var.spec.injector.replicas, null), 1) : 0

  # Ingress configuration
  ingress_is_enabled        = try(var.spec.ingress.enabled, false)
  ingress_external_hostname = try(var.spec.ingress.hostname, null)

  # Extract domain from hostname for certificate issuer
  # Example: "openbao.example.com" -> "example.com"
  ingress_cert_cluster_issuer_name = local.ingress_external_hostname != null ? (
    join(".", slice(split(".", local.ingress_external_hostname), 1,
    length(split(".", local.ingress_external_hostname))))
  ) : null

  # Auto-unseal seal HCL stanza
  auto_unseal = try(var.spec.auto_unseal, null)

  seal_hcl = (
    local.auto_unseal != null && try(local.auto_unseal.gcp_kms, null) != null ? join("", [
      "\nseal \"gcpckms\" {\n",
      "  project    = \"${local.auto_unseal.gcp_kms.project}\"\n",
      "  region     = \"${local.auto_unseal.gcp_kms.region}\"\n",
      "  key_ring   = \"${local.auto_unseal.gcp_kms.key_ring}\"\n",
      "  crypto_key = \"${local.auto_unseal.gcp_kms.crypto_key}\"\n",
      "}\n",
    ]) :
    local.auto_unseal != null && try(local.auto_unseal.aws_kms, null) != null ? join("", [
      "\nseal \"awskms\" {\n",
      "  region     = \"${local.auto_unseal.aws_kms.region}\"\n",
      "  kms_key_id = \"${local.auto_unseal.aws_kms.kms_key_id}\"\n",
      "}\n",
    ]) :
    local.auto_unseal != null && try(local.auto_unseal.azure_key_vault, null) != null ? join("", [
      "\nseal \"azurekeyvault\" {\n",
      "  vault_name = \"${local.auto_unseal.azure_key_vault.vault_name}\"\n",
      "  key_name   = \"${local.auto_unseal.azure_key_vault.key_name}\"\n",
      "  tenant_id  = \"${local.auto_unseal.azure_key_vault.tenant_id}\"\n",
      "}\n",
    ]) :
    local.auto_unseal != null && try(local.auto_unseal.transit, null) != null ? join("", [
      "\nseal \"transit\" {\n",
      "  address    = \"${local.auto_unseal.transit.address}\"\n",
      "  key_name   = \"${local.auto_unseal.transit.key_name}\"\n",
      "  mount_path = \"${coalesce(try(local.auto_unseal.transit.mount_path, null), "transit/")}\"\n",
      "}\n",
    ]) :
    ""
  )

  # Workload Identity service account email (GCP KMS only)
  workload_identity_sa = try(local.auto_unseal.gcp_kms.workload_identity_service_account, "")

  # Standalone mode server config HCL
  standalone_config = <<-EOT
ui = true

listener "tcp" {
  tls_disable = 1
  address = "[::]:8200"
  cluster_address = "[::]:8201"
}

storage "file" {
  path = "/openbao/data"
}
${local.seal_hcl}
EOT

  # HA mode server config HCL
  ha_raft_config = <<-EOT
ui = true

listener "tcp" {
  tls_disable = 1
  address = "[::]:8200"
  cluster_address = "[::]:8201"
}

storage "raft" {
  path = "/openbao/data"
}

service_registration "kubernetes" {}
${local.seal_hcl}
EOT

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  ingress_cert_secret_name         = "${var.metadata.name}-tls"
  ingress_certificate_name         = "${var.metadata.name}-certificate"
  ingress_gateway_name             = "${var.metadata.name}-external"
  ingress_http_redirect_route_name = "${var.metadata.name}-http-redirect"
  ingress_https_route_name         = "${var.metadata.name}-https"

  # OpenBao Helm chart values. Mirrors the Pulumi module's helm values map
  # (iac/pulumi/module/helm_chart.go) so both engines render the same chart config;
  # consumed by helm_release via `values = [yamlencode(local.helm_values)]`. helm
  # provider v3 dropped the `set {}` block, so conditional fragments are appended as
  # single-element lists (list branches unify where bare-object branches do not).
  helm_values = {
    fullnameOverride = var.metadata.name
    global = {
      enabled    = true
      tlsDisable = !local.tls_enabled
    }
    server = merge(concat(
      [
        {
          dataStorage = {
            enabled = true
            size    = var.spec.server_container.data_storage_size
          }
          resources = {
            requests = {
              cpu    = var.spec.server_container.resources.requests.cpu
              memory = var.spec.server_container.resources.requests.memory
            }
            limits = {
              cpu    = var.spec.server_container.resources.limits.cpu
              memory = var.spec.server_container.resources.limits.memory
            }
          }
          ha = merge(concat(
            [{ enabled = local.ha_enabled }],
            local.ha_enabled ? [{
              replicas = local.ha_replicas
              raft = {
                enabled   = true
                setNodeId = true
                config    = local.ha_raft_config
              }
            }] : [],
          )...)
          standalone = merge(concat(
            [{ enabled = !local.ha_enabled }],
            local.ha_enabled ? [] : [{ config = local.standalone_config }],
          )...)
        },
      ],
      local.workload_identity_sa != "" ? [{
        serviceAccount = {
          annotations = {
            "iam.gke.io/gcp-service-account" = local.workload_identity_sa
          }
        }
      }] : [],
    )...)
    ui = {
      enabled = local.ui_enabled
    }
    injector = merge(concat(
      [{ enabled = local.injector_enabled }],
      local.injector_enabled ? [{ replicas = local.injector_replicas }] : [],
    )...)
  }
}
