# Terraform module for Altinity ClickHouse Operator
# This is a placeholder - the operator is primarily deployed via Helm/Pulumi

resource "kubernetes_namespace" "kubernetes_altinity_operator" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name = local.namespace
  }
}

resource "helm_release" "kubernetes_altinity_operator" {
  name       = local.helm_release_name
  repository = "https://docs.altinity.com/clickhouse-operator/"
  chart      = "altinity-clickhouse-operator"
  version    = "0.25.4"
  namespace  = local.namespace

  # Chart values mirror the prior --set keys. helm provider v3 replaced the
  # `set {}` block with the `values`/`set` list attributes; we use the house
  # values=[yamlencode(...)] idiom. `--set watchNamespaces={}` parsed to an empty
  # list, preserved here as []. PARITY-NOTE: the Pulumi module instead sets
  # configs.files."config.yaml".watch.namespaces=[".*"]; reconcile via the parity sweep.
  values = [
    yamlencode({
      operator = {
        createCRD = true
        resources = {
          limits = {
            cpu    = var.spec.container.resources.limits.cpu
            memory = var.spec.container.resources.limits.memory
          }
          requests = {
            cpu    = var.spec.container.resources.requests.cpu
            memory = var.spec.container.resources.requests.memory
          }
        }
      }
      watchNamespaces = []
    })
  ]

  timeout         = 300
  atomic          = true
  cleanup_on_fail = true
  wait            = true
}

