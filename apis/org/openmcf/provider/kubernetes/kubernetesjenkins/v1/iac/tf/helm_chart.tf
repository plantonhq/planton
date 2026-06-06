resource "helm_release" "jenkins" {
  # Name the Helm release using the resource_id from locals
  name       = local.resource_id
  repository = local.jenkins_chart_repo
  chart      = local.jenkins_chart_name
  version    = local.jenkins_chart_version

  namespace = local.namespace

  # Merge container_resources into the user-defined helm_values:
  # "controller.resources" is how the Jenkins chart configures pod resource requests/limits.
  values = [
    yamlencode(
      merge(
        {
          "fullnameOverride" = local.resource_id,
          "controller" = {
            "image" = {
              "tag" = "latest" # Or your chosen tag
            },
            "admin" = {
              "existingSecret" = local.admin_credentials_secret_name,
              "passwordKey"    = local.jenkins_admin_password_secret_key
            },
            "resources" = {
              # CPU/memory from your container_resources
              "limits" = {
                "cpu"    = var.spec.container_resources.limits.cpu
                "memory" = var.spec.container_resources.limits.memory
              },
              "requests" = {
                "cpu"    = var.spec.container_resources.requests.cpu
                "memory" = var.spec.container_resources.requests.memory
              }
            }
          }
        },
        # Merge any user-provided overrides
        var.spec.helm_values != null ? var.spec.helm_values : {}
      )
    )
  ]
}
