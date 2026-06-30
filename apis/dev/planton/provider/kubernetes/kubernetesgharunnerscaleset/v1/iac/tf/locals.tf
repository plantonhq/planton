##############################################
# locals.tf
#
# Local variables for KubernetesGhaRunnerScaleSet
##############################################

locals {
  # Chart configuration
  # For OCI charts, the full URL must be passed as the chart parameter
  # (repository doesn't work with OCI registries in Terraform helm_release)
  chart_oci     = "oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set"
  chart_version = var.spec.helm_chart_version

  # Derived values
  namespace       = var.spec.namespace.value
  release_name    = var.metadata.name
  create_namespace = var.spec.create_namespace

  # Runner scale set name (defaults to release name)
  runner_scale_set_name = coalesce(var.spec.runner_scale_set_name, var.metadata.name)

  # Labels for Kubernetes resources
  labels = merge(
    {
      "planton.dev/resource"      = "true"
      "planton.dev/resource-name" = var.metadata.name
      "planton.dev/resource-kind" = "KubernetesGhaRunnerScaleSet"
    },
    var.metadata.id != "" ? { "planton.dev/resource-id" = var.metadata.id } : {},
    var.metadata.org != "" ? { "planton.dev/organization" = var.metadata.org } : {},
    var.metadata.env != "" ? { "planton.dev/environment" = var.metadata.env } : {},
    var.spec.labels
  )

  # Container mode mapping
  container_mode_type = lookup({
    "DIND"                = "dind"
    "KUBERNETES"          = "kubernetes"
    "KUBERNETES_NO_VOLUME" = "kubernetes-novolume"
    "DEFAULT"             = ""
  }, var.spec.container_mode.type, "")

  # GitHub authentication type
  use_pat_token       = var.spec.github.pat_token != null
  use_github_app      = var.spec.github.github_app != null
  use_existing_secret = var.spec.github.existing_secret_name != null && var.spec.github.existing_secret_name != ""

  github_secret_name = local.use_existing_secret ? var.spec.github.existing_secret_name : "${local.release_name}-github-secret"

  # Scaling configuration
  min_runners = try(var.spec.scaling.min_runners, 0)
  max_runners = try(var.spec.scaling.max_runners, 5)

  # Build Helm values
  helm_values = {
    githubConfigUrl = var.spec.github.config_url

    # GitHub secret configuration
    githubConfigSecret = local.use_existing_secret ? local.github_secret_name : (
      local.use_pat_token ? {
        github_token = var.spec.github.pat_token.token
      } : (
        local.use_github_app ? {
          github_app_id              = var.spec.github.github_app.app_id
          github_app_installation_id = var.spec.github.github_app.installation_id
          github_app_private_key     = base64decode(var.spec.github.github_app.private_key_base64)
        } : null
      )
    )

    # Scaling
    minRunners = local.min_runners
    maxRunners = local.max_runners

    # Runner group and name
    runnerGroup        = var.spec.runner_group != "" ? var.spec.runner_group : null
    runnerScaleSetName = local.runner_scale_set_name

    # Labels and annotations
    labels      = length(var.spec.labels) > 0 ? var.spec.labels : null
    annotations = length(var.spec.annotations) > 0 ? var.spec.annotations : null
  }

  # Container mode values
  container_mode_values = local.container_mode_type != "" ? {
    containerMode = merge(
      { type = local.container_mode_type },
      local.container_mode_type == "kubernetes" && var.spec.container_mode.work_volume_claim != null ? {
        kubernetesModeWorkVolumeClaim = {
          accessModes = var.spec.container_mode.work_volume_claim.access_modes
          storageClassName = var.spec.container_mode.work_volume_claim.storage_class != "" ? var.spec.container_mode.work_volume_claim.storage_class : null
          resources = {
            requests = {
              storage = var.spec.container_mode.work_volume_claim.size
            }
          }
        }
      } : {}
    )
  } : {}

  # Controller service account values
  controller_sa_values = try(var.spec.controller_service_account.name, "") != "" || try(var.spec.controller_service_account.namespace, "") != "" ? {
    controllerServiceAccount = {
      name      = try(var.spec.controller_service_account.name, null)
      namespace = try(var.spec.controller_service_account.namespace, null)
    }
  } : {}

  # Template spec for runner container
  runner_spec = var.spec.runner != null ? {
    template = {
      spec = merge(
        # Containers
        {
          containers = [
            merge(
              {
                name    = "runner"
                command = ["/home/runner/run.sh"]
              },
              try(var.spec.runner.image.repository, "") != "" ? {
                image = "${var.spec.runner.image.repository}${try(var.spec.runner.image.tag, "") != "" ? ":${var.spec.runner.image.tag}" : ""}"
              } : {},
              try(var.spec.runner.image.pull_policy, "") != "" ? {
                imagePullPolicy = var.spec.runner.image.pull_policy
              } : {},
              try(var.spec.runner.resources, null) != null ? {
                resources = {
                  requests = try(var.spec.runner.resources.requests, null) != null ? {
                    cpu    = try(var.spec.runner.resources.requests.cpu, null)
                    memory = try(var.spec.runner.resources.requests.memory, null)
                  } : null
                  limits = try(var.spec.runner.resources.limits, null) != null ? {
                    cpu    = try(var.spec.runner.resources.limits.cpu, null)
                    memory = try(var.spec.runner.resources.limits.memory, null)
                  } : null
                }
              } : {},
              length(try(var.spec.runner.env, [])) > 0 ? {
                env = [for e in var.spec.runner.env : {
                  name  = e.name
                  value = e.value
                }]
              } : {},
              length(var.spec.persistent_volumes) > 0 || length(try(var.spec.runner.volume_mounts, [])) > 0 ? {
                volumeMounts = concat(
                  [for pv in var.spec.persistent_volumes : {
                    name      = pv.name
                    mountPath = pv.mount_path
                    readOnly  = pv.read_only
                  }],
                  [for vm in try(var.spec.runner.volume_mounts, []) : {
                    name      = vm.name
                    mountPath = vm.mount_path
                    readOnly  = vm.read_only
                    subPath   = vm.sub_path != "" ? vm.sub_path : null
                  }]
                )
              } : {}
            )
          ]
        },
        # Volumes for persistent volumes
        length(var.spec.persistent_volumes) > 0 ? {
          volumes = [for pv in var.spec.persistent_volumes : {
            name = pv.name
            persistentVolumeClaim = {
              claimName = "${local.release_name}-${pv.name}"
            }
          }]
        } : {},
        # Image pull secrets
        length(var.spec.image_pull_secrets) > 0 ? {
          imagePullSecrets = [for s in var.spec.image_pull_secrets : { name = s }]
        } : {}
      )
    }
  } : {}

  # Final Helm values
  helm_values_final = merge(
    local.helm_values,
    local.container_mode_values,
    local.controller_sa_values,
    local.runner_spec
  )

  # PVC names for output
  pvc_names = [for pv in var.spec.persistent_volumes : "${local.release_name}-${pv.name}"]
}

