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
  description = "GcpCloudBuildConnection specification"
  type = object({
    # The project the connection lives in: a literal project ID or a
    # GcpProject reference. Empty means the provider's default project.
    # Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The region the connection lives in, e.g. "us-central1". Its
    # repositories, and the triggers that build from them, use the same
    # region. Required. Immutable.
    location = string

    # The connection's ID, unique in the project and region: letters,
    # digits, and any of -._~%!$&'()*+,;=@ (Google's rule). Defaults to
    # metadata.name. Immutable.
    connection_id = optional(string, "")

    # Annotations on the connection (AIP-128 key/value metadata; a
    # connection has no labels). Only the keys declared here are managed.
    annotations = optional(map(string), {})

    # Turn the connection off: repository API calls and webhook processing
    # for every repository in it stop until it is turned back on.
    disabled = optional(bool, false)

    # A connection to github.com through Cloud Build's GitHub App.
    github_config = optional(object({
      # The installation ID of Cloud Build's GitHub App on the GitHub account
      # or organization (from the app's settings URL after installing it).
      # 0 leaves the connection waiting for the installation.
      app_installation_id = optional(number, 0)

      # The OAuth credential of the GitHub account that authorized Cloud
      # Build's GitHub App -- a robot account is recommended over a person.
      authorizer_credential = optional(object({
        # The Secret Manager secret version holding the OAuth token, as
        # projects/{project}/secrets/{secret}/versions/{version}: a
        # GcpSecretManagerSecret reference (its latest_version_name output, set
        # when the secret declares an initial version) or a literal version name
        # ("versions/latest" follows rotation). The token must be tied to Cloud
        # Build's GitHub App.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        oauth_token_secret_version = optional(string, "")
      }))
    }))

    # A connection to a GitHub Enterprise server through a GitHub App
    # created on that server.
    github_enterprise_config = optional(object({
      # The server's URI, e.g. "https://github.example.com". Required.
      host_uri = string

      # The GitHub App's ID.
      app_id = optional(number, 0)

      # The GitHub App's installation ID on the organization.
      app_installation_id = optional(number, 0)

      # The GitHub App's URL-friendly name.
      app_slug = optional(string, "")

      # The Secret Manager secret version holding the GitHub App's private key
      # (projects/{project}/secrets/{secret}/versions/{version}): a
      # GcpSecretManagerSecret reference or a literal version name.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      private_key_secret_version = optional(string, "")

      # The Secret Manager secret version holding the GitHub App's webhook
      # secret: a GcpSecretManagerSecret reference or a literal version name.
      # Updates in place.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      webhook_secret_secret_version = optional(string, "")

      # The PEM CA certificate Cloud Build trusts when calling the server, for
      # a server with a private certificate authority.
      ssl_ca = optional(string, "")

      # Reach an on-premises server privately through Service Directory
      # instead of over the internet.
      service_directory_config = optional(object({
        # The Service Directory service, as
        # projects/{project}/locations/{location}/namespaces/{namespace}/services/{service}.
        # Required.
        service = string
      }))
    }))

    # A connection to gitlab.com or a GitLab Enterprise server.
    gitlab_config = optional(object({
      # The GitLab server's URI. Empty means https://gitlab.com.
      host_uri = optional(string, "")

      # A personal access token with the "api" scope, which Cloud Build uses
      # to create webhooks and read repositories. Required.
      authorizer_credential = object({
        # The Secret Manager secret version holding the token, as
        # projects/{project}/secrets/{secret}/versions/{version}: a
        # GcpSecretManagerSecret reference (its latest_version_name output, set
        # when the secret declares an initial version) or a literal version name
        # ("versions/latest" follows rotation). Required.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        user_token_secret_version = string
      })

      # A personal access token with at least the "read_api" scope, which
      # Cloud Build uses for read-only calls. Required.
      read_authorizer_credential = object({
        # The Secret Manager secret version holding the token, as
        # projects/{project}/secrets/{secret}/versions/{version}: a
        # GcpSecretManagerSecret reference (its latest_version_name output, set
        # when the secret declares an initial version) or a literal version name
        # ("versions/latest" follows rotation). Required.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        user_token_secret_version = string
      })

      # The Secret Manager secret version holding the webhook secret of the
      # GitLab project: a GcpSecretManagerSecret reference or a literal version
      # name. Required. Immutable (a change replaces the connection).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      webhook_secret_secret_version = string

      # The PEM CA certificate Cloud Build trusts when calling a GitLab
      # Enterprise server with a private certificate authority.
      ssl_ca = optional(string, "")

      # Reach an on-premises GitLab Enterprise server privately through
      # Service Directory instead of over the internet.
      service_directory_config = optional(object({
        # The Service Directory service, as
        # projects/{project}/locations/{location}/namespaces/{namespace}/services/{service}.
        # Required.
        service = string
      }))
    }))

    # A connection to a Bitbucket Cloud workspace.
    bitbucket_cloud_config = optional(object({
      # The Bitbucket Cloud workspace ID to connect. Required.
      workspace = string

      # A workspace, project, or repository access token with the "webhook",
      # "repository", "repository:admin", and "pullrequest" scopes (a system
      # account is recommended). Required.
      authorizer_credential = object({
        # The Secret Manager secret version holding the token, as
        # projects/{project}/secrets/{secret}/versions/{version}: a
        # GcpSecretManagerSecret reference (its latest_version_name output, set
        # when the secret declares an initial version) or a literal version name
        # ("versions/latest" follows rotation). Required.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        user_token_secret_version = string
      })

      # An access token with "repository" access, for read-only calls.
      # Required.
      read_authorizer_credential = object({
        # The Secret Manager secret version holding the token, as
        # projects/{project}/secrets/{secret}/versions/{version}: a
        # GcpSecretManagerSecret reference (its latest_version_name output, set
        # when the secret declares an initial version) or a literal version name
        # ("versions/latest" follows rotation). Required.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        user_token_secret_version = string
      })

      # The Secret Manager secret version holding the webhook secret Cloud
      # Build verifies webhook events with: a GcpSecretManagerSecret reference
      # or a literal version name. Required. Immutable (a change replaces the
      # connection).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      webhook_secret_secret_version = string
    }))

    # A connection to a Bitbucket Data Center server.
    bitbucket_data_center_config = optional(object({
      # The Bitbucket Data Center server's URI. Required.
      host_uri = string

      # An HTTP access token with the "REPO_ADMIN" scope. Required.
      authorizer_credential = object({
        # The Secret Manager secret version holding the token, as
        # projects/{project}/secrets/{secret}/versions/{version}: a
        # GcpSecretManagerSecret reference (its latest_version_name output, set
        # when the secret declares an initial version) or a literal version name
        # ("versions/latest" follows rotation). Required.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        user_token_secret_version = string
      })

      # An HTTP access token with "REPO_READ" access, for read-only calls.
      # Required.
      read_authorizer_credential = object({
        # The Secret Manager secret version holding the token, as
        # projects/{project}/secrets/{secret}/versions/{version}: a
        # GcpSecretManagerSecret reference (its latest_version_name output, set
        # when the secret declares an initial version) or a literal version name
        # ("versions/latest" follows rotation). Required.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        user_token_secret_version = string
      })

      # The Secret Manager secret version holding the webhook secret: a
      # GcpSecretManagerSecret reference or a literal version name. Required.
      # Immutable (a change replaces the connection).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      webhook_secret_secret_version = string

      # The PEM CA certificate Cloud Build trusts when calling the server.
      ssl_ca = optional(string, "")

      # Reach an on-premises server privately through Service Directory
      # instead of over the internet.
      service_directory_config = optional(object({
        # The Service Directory service, as
        # projects/{project}/locations/{location}/namespaces/{namespace}/services/{service}.
        # Required.
        service = string
      }))
    }))

    # What destroy does:
    #   "" / "DELETE" -- the connection is deleted
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the connection leaves management and stays in Google
    deletion_policy = optional(string, "")
  })
}
