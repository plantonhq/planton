variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "GcpDeployCustomTargetType specification"
  type = object({
    # The project the type lives in -- the project of the targets that use
    # it: a literal project ID or a GcpProject reference. Empty means the
    # provider's default project. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The region the type lives in, e.g. "us-central1" -- the region of the
    # targets that use it. Required. Immutable.
    location = string

    # The type's ID, unique in the project and location: 1-63 lowercase
    # letters, digits, and hyphens, starting with a letter and not ending
    # with a hyphen. Defaults to metadata.name. Immutable.
    custom_target_type_id = optional(string, "")

    # A description of the type, up to 255 characters.
    description = optional(string, "")

    # Labels on the type. The platform attribution labels are added on top
    # and win on key conflicts.
    labels = optional(map(string), {})

    # Annotations on the type (user metadata Cloud Deploy never reads).
    # Only the keys declared here are managed.
    annotations = optional(map(string), {})

    # Render and deploy through Skaffold custom actions. Mutually exclusive
    # with tasks.
    custom_actions = optional(object({
      # The name of the Skaffold custom action that deploys, as defined in
      # the release's skaffold.yaml or an included module. Required.
      deploy_action = string

      # The name of the Skaffold custom action that renders. Empty means
      # Cloud Deploy renders with `skaffold render`.
      render_action = optional(string, "")

      # Remote Skaffold modules Cloud Deploy adds to the release's Skaffold
      # config, so the custom actions can live in a shared repository or
      # bucket instead of every application's skaffold.yaml.
      include_skaffold_modules = optional(list(object({
        # The Skaffold config modules (the `metadata.name` of each config) to
        # use from the source. Empty uses every config in the file.
        configs = optional(list(string), [])

        # A Git repository Cloud Deploy clones.
        git = optional(object({
          # The repository to clone, e.g. "https://github.com/acme/deploy-actions.git".
          # Required.
          repo = string

          # The Skaffold file's path from the repository root. Empty means
          # skaffold.yaml at the root.
          path = optional(string, "")

          # The branch or tag to clone. Empty means the default branch.
          ref = optional(string, "")
        }))

        # A repository linked through a Cloud Build connection.
        google_cloud_build_repo = optional(object({
          # The repository, by full name
          # (projects/{p}/locations/{l}/connections/{c}/repositories/{r}): a
          # GcpCloudBuildRepository reference or the literal name. Required.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          repository = string

          # The Skaffold file's path from the repository root. Empty means
          # skaffold.yaml at the root.
          path = optional(string, "")

          # The branch or tag to clone. Empty means the default branch.
          ref = optional(string, "")
        }))

        # A Cloud Storage location Cloud Deploy copies.
        google_cloud_storage = optional(object({
          # The objects to copy, recursively: "gs://my-bucket/dir/configs/*"
          # copies everything under dir/configs. Required.
          source = string

          # The Skaffold file's path relative to source. Empty means skaffold.yaml.
          path = optional(string, "")
        }))
      })), [])
    }))

    # Render and deploy through containers Cloud Deploy runs. Mutually
    # exclusive with custom_actions.
    tasks = optional(object({
      # The task that deploys. Required.
      deploy = object({
        # The container Cloud Deploy runs in Cloud Build for this task -- set it,
        # since it is the task's whole definition.
        container = optional(object({
          # The container image, e.g.
          # "us-docker.pkg.dev/acme/deploy/vendor-deployer:1.4". Required.
          image = string

          # The entrypoint, replacing the image's own.
          command = optional(list(string), [])

          # The arguments, replacing the image's default arguments.
          args = optional(list(string), [])

          # Environment variables set in the container, alongside the
          # CLOUD_DEPLOY_* variables Cloud Deploy sets.
          env = optional(map(string), {})
        }))
      })

      # The task that renders. Unset means Cloud Deploy's default rendering.
      render = optional(object({
        # The container Cloud Deploy runs in Cloud Build for this task -- set it,
        # since it is the task's whole definition.
        container = optional(object({
          # The container image, e.g.
          # "us-docker.pkg.dev/acme/deploy/vendor-deployer:1.4". Required.
          image = string

          # The entrypoint, replacing the image's own.
          command = optional(list(string), [])

          # The arguments, replacing the image's default arguments.
          args = optional(list(string), [])

          # Environment variables set in the container, alongside the
          # CLOUD_DEPLOY_* variables Cloud Deploy sets.
          env = optional(map(string), {})
        }))
      }))
    }))

    # What destroy does:
    #   "" / "DELETE" -- the type is deleted
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the type leaves management and stays in Google
    deletion_policy = optional(string, "")
  })
}
