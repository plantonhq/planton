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
  description = "GcpCloudBuildTrigger specification"
  type = object({
    # The project the trigger lives in: a literal project ID or a GcpProject
    # reference. Empty means the provider's default project. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Cloud Build location the trigger lives in: "global" (the default
    # when empty) or a region such as "us-central1". A trigger on a
    # repository linked through a regional connection, or whose builds use
    # a private pool, must be in that region. Immutable.
    location = optional(string, "")

    # The trigger's name, unique in the project: 1-64 letters, digits, and
    # dashes, starting and ending with a letter or digit (Google's rule).
    # Defaults to metadata.name. Renames in place.
    trigger_name = optional(string, "")

    # A human-readable description of the trigger.
    description = optional(string, "")

    # Turn the trigger off: it never starts a build until turned back on.
    disabled = optional(bool, false)

    # Tags on the trigger, for filtering in the console and API. (Triggers
    # have no labels.)
    tags = optional(list(string), [])

    # User substitutions every build the trigger starts receives, referenced
    # in the build as $_NAME or ${_NAME}. Keys start with an underscore and
    # use only uppercase letters, digits, and underscores (Google's rule),
    # e.g. "_DEPLOY_ENV".
    substitutions = optional(map(string), {})

    # The service account builds run as, and that Google uses for every
    # user-controlled operation on the trigger, as
    # projects/{project}/serviceAccounts/{email}: a GcpServiceAccount
    # reference (its name output) or a literal. Empty means the project's
    # default Cloud Build service account. The deploying identity needs
    # iam.serviceAccounts.actAs on it; the account needs
    # roles/logging.logWriter (and whatever the build itself touches), and
    # the build must set its log destination (build.logs_bucket, or
    # build.options.logging CLOUD_LOGGING_ONLY or NONE).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    service_account = optional(string, "")

    # Require a person with the Cloud Build Approver role to approve each
    # build before it runs.
    approval_config = optional(object({
      # Builds wait in a pending state until someone with the Cloud Build
      # Approver role (roles/cloudbuild.builds.approver) approves them.
      approval_required = optional(bool, false)
    }))

    # Fire on pushes or pull requests of a repository linked through a Cloud
    # Build connection (GcpCloudBuildRepository).
    repository_event_config = optional(object({
      # The linked repository, as
      # projects/{project}/locations/{location}/connections/{connection}/repositories/{repository}:
      # a GcpCloudBuildRepository reference (its name output) or a literal.
      # The trigger must be in the repository's region.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      repository = optional(string, "")

      # Build pull requests. Exactly one of pull_request or push.
      pull_request = optional(object({
        # An RE2 regular expression of the pull request's BASE branch, e.g.
        # "^main$". Required on github and bitbucket_server_trigger_config.
        branch = optional(string, "")

        # Whether a build waits for a "/gcbrun" comment:
        #   "COMMENTS_DISABLED" -- every pull request builds
        #   "COMMENTS_ENABLED"  -- builds start only on a "/gcbrun" comment
        #                          from an owner or collaborator
        #   "COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY" -- collaborators'
        #                          pull requests build; others wait for the
        #                          comment
        # Empty leaves Google's default (COMMENTS_DISABLED).
        comment_control = optional(string, "")

        # Build pull requests whose base branch does NOT match branch.
        invert_regex = optional(bool, false)
      }))

      # Build pushes. Exactly one of pull_request or push.
      push = optional(object({
        # An RE2 regular expression of the pushed branch, e.g. "^main$" or
        # "^release/.*". Exactly one of branch or tag.
        branch = optional(string, "")

        # An RE2 regular expression of the pushed tag, e.g. "^v[0-9]+\\.". Exactly
        # one of branch or tag.
        tag = optional(string, "")

        # Build pushes whose ref does NOT match.
        invert_regex = optional(bool, false)
      }))
    }))

    # Fire on GitHub pushes or pull requests through the first-generation
    # Cloud Build GitHub App or a GitHub Enterprise config. Mutually
    # exclusive with trigger_template.
    github = optional(object({
      # The repository's owner: the user or organization, e.g.
      # "googlecloudplatform" for github.com/googlecloudplatform/cloud-builders.
      owner = optional(string, "")

      # The repository's name, e.g. "cloud-builders".
      name = optional(string, "")

      # The GitHub Enterprise config the installation uses, as
      # projects/{project}/locations/{location}/githubEnterpriseConfigs/{id}.
      # Empty means github.com.
      enterprise_config_resource_name = optional(string, "")

      # Build pull requests (branch is required). Exactly one of pull_request
      # or push.
      pull_request = optional(object({
        # An RE2 regular expression of the pull request's BASE branch, e.g.
        # "^main$". Required on github and bitbucket_server_trigger_config.
        branch = optional(string, "")

        # Whether a build waits for a "/gcbrun" comment:
        #   "COMMENTS_DISABLED" -- every pull request builds
        #   "COMMENTS_ENABLED"  -- builds start only on a "/gcbrun" comment
        #                          from an owner or collaborator
        #   "COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY" -- collaborators'
        #                          pull requests build; others wait for the
        #                          comment
        # Empty leaves Google's default (COMMENTS_DISABLED).
        comment_control = optional(string, "")

        # Build pull requests whose base branch does NOT match branch.
        invert_regex = optional(bool, false)
      }))

      # Build pushes. Exactly one of pull_request or push.
      push = optional(object({
        # An RE2 regular expression of the pushed branch, e.g. "^main$" or
        # "^release/.*". Exactly one of branch or tag.
        branch = optional(string, "")

        # An RE2 regular expression of the pushed tag, e.g. "^v[0-9]+\\.". Exactly
        # one of branch or tag.
        tag = optional(string, "")

        # Build pushes whose ref does NOT match.
        invert_regex = optional(bool, false)
      }))
    }))

    # Fire on Bitbucket Server pushes or pull requests through a
    # first-generation Bitbucket Server config.
    bitbucket_server_trigger_config = optional(object({
      # The Bitbucket Server config the trigger uses, as
      # projects/{project}/locations/{location}/bitbucketServerConfigs/{id}.
      # Required.
      bitbucket_server_config_resource = string

      # The key of the Bitbucket project the repository is in, e.g. "TEST" for
      # https://mybitbucket.server/projects/TEST/repos/test-repo. Required.
      project_key = string

      # The repository's slug (its URL form), e.g. "test-repo". Required.
      repo_slug = string

      # Build pull requests (branch is required). Exactly one of pull_request
      # or push.
      pull_request = optional(object({
        # An RE2 regular expression of the pull request's BASE branch, e.g.
        # "^main$". Required on github and bitbucket_server_trigger_config.
        branch = optional(string, "")

        # Whether a build waits for a "/gcbrun" comment:
        #   "COMMENTS_DISABLED" -- every pull request builds
        #   "COMMENTS_ENABLED"  -- builds start only on a "/gcbrun" comment
        #                          from an owner or collaborator
        #   "COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY" -- collaborators'
        #                          pull requests build; others wait for the
        #                          comment
        # Empty leaves Google's default (COMMENTS_DISABLED).
        comment_control = optional(string, "")

        # Build pull requests whose base branch does NOT match branch.
        invert_regex = optional(bool, false)
      }))

      # Build pushes. Exactly one of pull_request or push.
      push = optional(object({
        # An RE2 regular expression of the pushed branch, e.g. "^main$" or
        # "^release/.*". Exactly one of branch or tag.
        branch = optional(string, "")

        # An RE2 regular expression of the pushed tag, e.g. "^v[0-9]+\\.". Exactly
        # one of branch or tag.
        tag = optional(string, "")

        # Build pushes whose ref does NOT match.
        invert_regex = optional(bool, false)
      }))
    }))

    # Fire on pushes or pull requests of a Developer Connect git repository
    # link.
    developer_connect_event_config = optional(object({
      # The Developer Connect git repository link, as
      # projects/{project}/locations/{location}/connections/{connection}/gitRepositoryLinks/{link}.
      # Required.
      git_repository_link = string

      # Build pull requests. Exactly one of pull_request or push.
      pull_request = optional(object({
        # An RE2 regular expression of the pull request's BASE branch, e.g.
        # "^main$". Required on github and bitbucket_server_trigger_config.
        branch = optional(string, "")

        # Whether a build waits for a "/gcbrun" comment:
        #   "COMMENTS_DISABLED" -- every pull request builds
        #   "COMMENTS_ENABLED"  -- builds start only on a "/gcbrun" comment
        #                          from an owner or collaborator
        #   "COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY" -- collaborators'
        #                          pull requests build; others wait for the
        #                          comment
        # Empty leaves Google's default (COMMENTS_DISABLED).
        comment_control = optional(string, "")

        # Build pull requests whose base branch does NOT match branch.
        invert_regex = optional(bool, false)
      }))

      # Build pushes. Exactly one of pull_request or push.
      push = optional(object({
        # An RE2 regular expression of the pushed branch, e.g. "^main$" or
        # "^release/.*". Exactly one of branch or tag.
        branch = optional(string, "")

        # An RE2 regular expression of the pushed tag, e.g. "^v[0-9]+\\.". Exactly
        # one of branch or tag.
        tag = optional(string, "")

        # Build pushes whose ref does NOT match.
        invert_regex = optional(bool, false)
      }))
    }))

    # Fire on pushes to a Cloud Source Repository. Mutually exclusive with
    # github.
    trigger_template = optional(object({
      # The project that owns the repository. Empty means the trigger's
      # project.
      project_id = optional(string, "")

      # The Cloud Source Repository's name. Empty means "default".
      repo_name = optional(string, "")

      # The directory, relative to the repository root, builds run in.
      dir = optional(string, "")

      # An RE2 regular expression of the branches to build. Exactly one of
      # branch_name, tag_name, or commit_sha.
      branch_name = optional(string, "")

      # An RE2 regular expression of the tags to build. Exactly one of
      # branch_name, tag_name, or commit_sha.
      tag_name = optional(string, "")

      # An explicit commit SHA to build. Exactly one of branch_name,
      # tag_name, or commit_sha.
      commit_sha = optional(string, "")

      # Build revisions that do NOT match the branch or tag expression.
      invert_regex = optional(bool, false)
    }))

    # Fire on each message published to a Pub/Sub topic. Cloud Build creates
    # and owns the push subscription.
    pubsub_config = optional(object({
      # The topic, as projects/{project}/topics/{topic}: a GcpPubSubTopic
      # reference (its topic_id output) or a literal. Required.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      topic = string

      # The service account email the push subscription authenticates as: a
      # GcpServiceAccount reference (its email output) or a literal. Empty
      # lets Google choose.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      service_account_email = optional(string, "")
    }))

    # Fire on each HTTP POST to the trigger's webhook URL, authenticated by
    # a secret key.
    webhook_config = optional(object({
      # The Secret Manager secret version holding the webhook's key, as
      # projects/{project}/secrets/{secret}/versions/{version}: a
      # GcpSecretManagerSecret reference (its latest_version_name output) or a
      # literal. Callers pass the key as the URL's "secret" parameter. Google's
      # Cloud Build service agent
      # (service-{PROJECT_NUMBER}@gcp-sa-cloudbuild.iam.gserviceaccount.com)
      # reads it, so grant that agent roles/secretmanager.secretAccessor on
      # the secret (GcpSecretManagerSecret.iam_members). Required.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      secret = string
    }))

    # The repository and ref a Pub/Sub, webhook, or manual trigger checks
    # out (triggers that respond to repository events build the commit that
    # caused the event and ignore this).
    source_to_build = optional(object({
      # A repository linked through a Cloud Build connection, as
      # projects/{project}/locations/{location}/connections/{connection}/repositories/{repository}:
      # a GcpCloudBuildRepository reference (its name output) or a literal.
      # Exactly one of repository or uri.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      repository = optional(string, "")

      # The repository's URI, e.g. "https://github.com/acme/orders" (for the
      # first-generation GitHub App, Bitbucket Server, or a Cloud Source
      # Repository). Exactly one of repository or uri.
      uri = optional(string, "")

      # The branch or tag to build, as a full ref: "refs/heads/main" or
      # "refs/tags/v1.0.0". Required.
      ref = string

      # The kind of repository: "GITHUB", "BITBUCKET_SERVER",
      # "CLOUD_SOURCE_REPOSITORIES", or "UNKNOWN" (the values the provider
      # accepts; a connected GitLab or Bitbucket Cloud repository is still
      # named through repository). Required.
      repo_type = string

      # The GitHub Enterprise config the repository is reached through, as
      # projects/{project}/locations/{location}/githubEnterpriseConfigs/{id}.
      github_enterprise_config = optional(string, "")

      # The Bitbucket Server config the repository is reached through, as
      # projects/{project}/locations/{location}/bitbucketServerConfigs/{id}.
      bitbucket_server_config = optional(string, "")
    }))

    # A Common Expression Language filter on the incoming event; the trigger
    # fires only when it evaluates true. Used with Pub/Sub and webhook
    # triggers, over the substitutions they bind from the payload, e.g.
    # "_ACTION == 'INSERT'".
    filter = optional(string, "")

    # File globs (Go filepath.Match plus "**") of changes to ignore: a
    # commit whose changed files all match is not built. Repository-event
    # triggers only.
    ignored_files = optional(list(string), [])

    # File globs of changes that matter: when set, a commit is built only if
    # at least one changed file (not ignored) matches. Repository-event
    # triggers only.
    included_files = optional(list(string), [])

    # Show the build log link on the GitHub check run when the build
    # finishes: "INCLUDE_BUILD_LOGS_WITH_STATUS". Google accepts it only on
    # GitHub triggers (github, or repository_event_config on a GitHub
    # repository) and refuses it with INVALID_ARGUMENT otherwise. Empty
    # leaves logs off GitHub.
    include_build_logs = optional(string, "")

    # The build, declared inline. Exactly one of build, filename, or
    # git_file_source.
    build = optional(object({
      # The build's steps, run in order (or as wait_for allows). At least one.
      # The sum of the steps' timeouts may not exceed the build's timeout.
      steps = list(object({
        # The image the step runs, e.g. "gcr.io/cloud-builders/docker",
        # "gcr.io/google.com/cloudsdktool/cloud-sdk", or "ubuntu". Required.
        name = string

        # The step's ID, for other steps' wait_for.
        id = optional(string, "")

        # Arguments to the image's entrypoint (or, with no entrypoint, the
        # command and its arguments). Not with script.
        args = optional(list(string), [])

        # An entrypoint replacing the image's own. Not with script.
        entrypoint = optional(string, "")

        # A shell script run in the step, instead of entrypoint and args.
        script = optional(string, "")

        # The working directory, relative to the build's workspace (an absolute
        # path is outside it and not kept between steps unless it is a volume).
        dir = optional(string, "")

        # Environment variables, each "KEY=VALUE".
        env = optional(list(string), [])

        # Names of secret environment variables the step receives, each defined
        # in the build's available_secrets.secret_manager (env) or secrets
        # (secret_env keys).
        secret_env = optional(list(string), [])

        # The IDs of the steps this one waits for. Empty waits for every
        # earlier step; ["-"] starts it at once.
        wait_for = optional(list(string), [])

        # How long the step may run, e.g. "300s". Empty means until the build
        # times out.
        timeout = optional(string, "")

        # Let the step fail without failing the build.
        allow_failure = optional(bool, false)

        # Let the step fail without failing the build only with one of these
        # exit codes (takes precedence over allow_failure).
        allow_exit_codes = optional(list(number), [])

        # Volumes mounted into the step. A named volume must be used by at
        # least two steps.
        volumes = optional(list(object({
          # The volume's name (a valid Docker volume name, unique in the step).
          # Required.
          name = string

          # The absolute path the volume mounts at. Required.
          path = string
        })), [])
      }))

      # How long the build may run, in seconds with an "s" suffix, e.g.
      # "1200s". Empty means 600s. Must be at least the sum of the steps'
      # timeouts.
      timeout = optional(string, "")

      # How long the build may wait in the queue before it expires, e.g.
      # "3600s". Empty means no limit.
      queue_ttl = optional(string, "")

      # Images pushed after every step succeeds, e.g.
      # "us-docker.pkg.dev/my-project/apps/orders:$SHORT_SHA". Their digests
      # land in the build's results.
      images = optional(list(string), [])

      # The Cloud Storage location build logs are written to, as
      # gs://{bucket} or gs://{bucket}/{path}: a GcpGcsBucket reference (its
      # url output) or a literal. Google writes {logs_bucket}/log-{build_id}.txt.
      # Empty means Cloud Build's default logs bucket. A build running as a
      # user-specified service account needs this or options.logging set.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      logs_bucket = optional(string, "")

      # Build-level substitutions, merged with the trigger's.
      substitutions = optional(map(string), {})

      # Tags on each build, for filtering the build history. (Not Docker tags.)
      tags = optional(list(string), [])

      # Non-image artifacts uploaded after every step succeeds.
      artifacts = optional(object({
        # Images pushed after every step succeeds (the same as the build's
        # images).
        images = optional(list(string), [])

        # Workspace files uploaded to Cloud Storage.
        objects = optional(object({
          # The destination, as gs://{bucket}/{path}/.
          location = optional(string, "")

          # Globs of workspace files to upload.
          paths = optional(list(string), [])
        }))

        # Maven artifacts uploaded to Artifact Registry.
        maven_artifacts = optional(list(object({
          # The repository, as https://{region}-maven.pkg.dev/{project}/{repository}.
          repository = optional(string, "")

          # The artifact's path in the workspace, e.g.
          # "my-app/target/my-app-1.0.SNAPSHOT.jar".
          path = optional(string, "")

          # The Maven artifactId.
          artifact_id = optional(string, "")

          # The Maven groupId.
          group_id = optional(string, "")

          # The Maven version.
          version = optional(string, "")
        })), [])

        # npm packages uploaded to Artifact Registry.
        npm_packages = optional(list(object({
          # The repository, as https://{region}-npm.pkg.dev/{project}/{repository}.
          repository = optional(string, "")

          # The directory holding the package's package.json.
          package_path = optional(string, "")
        })), [])

        # Python packages uploaded to Artifact Registry.
        python_packages = optional(list(object({
          # The repository, as https://{region}-python.pkg.dev/{project}/{repository}.
          repository = optional(string, "")

          # Globs of the files to upload, usually "dist/*".
          paths = optional(list(string), [])
        })), [])
      }))

      # Machine, disk, logging, and private-pool options.
      options = optional(object({
        # The machine type builds run on in Google's default pool: "E2_MEDIUM",
        # "E2_HIGHCPU_8", "E2_HIGHCPU_32", "N1_HIGHCPU_8", or "N1_HIGHCPU_32".
        # Empty means the standard machine. A private pool's machine is set on
        # the pool.
        machine_type = optional(string, "")

        # The minimum disk size in GB for the build's VM (some of it is used by
        # the system). 0 means the standard size.
        disk_size_gb = optional(number, 0)

        # A private pool the builds run on, as
        # projects/{project}/locations/{location}/workerPools/{pool}: a
        # GcpCloudBuildWorkerPool reference (its name output) or a literal. The
        # trigger must be in the pool's region. Google's API marks this field
        # (BuildOptions.workerPool) deprecated in favor of options.pool.name,
        # which the provider does not expose; a filename or git_file_source
        # build sets options.pool.name in its cloudbuild.yaml instead.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        worker_pool = optional(string, "")

        # Where build logs go:
        #   "CLOUD_LOGGING_ONLY" -- Cloud Logging only
        #   "GCS_ONLY"           -- the logs bucket only
        #   "LEGACY"             -- both (the old default)
        #   "STACKDRIVER_ONLY"   -- Cloud Logging only (older name)
        #   "NONE"               -- nowhere
        #   "LOGGING_UNSPECIFIED"-- Google's default
        # A build running as a user-specified service account needs
        # CLOUD_LOGGING_ONLY, NONE, or a logs_bucket.
        logging = optional(string, "")

        # Whether logs stream to the Cloud Storage bucket while the build runs:
        # "STREAM_ON", "STREAM_OFF", or "STREAM_DEFAULT".
        log_streaming_option = optional(string, "")

        # Request verifiable build provenance: "VERIFIED" or "NOT_VERIFIED".
        requested_verify_option = optional(string, "")

        # Hashes of the source recorded in the build's provenance: "SHA256",
        # "MD5", or "NONE".
        source_provenance_hash = optional(list(string), [])

        # Environment variables for every step, each "KEY=VALUE"; a step's own
        # env wins on a collision.
        env = optional(list(string), [])

        # Names of secret environment variables every step receives, each
        # defined in the build's secrets.
        secret_env = optional(list(string), [])

        # Volumes mounted into every step (names and paths may not collide with
        # a step's own volumes). Not valid on a one-step build.
        volumes = optional(list(object({
          # The volume's name (a valid Docker volume name).
          name = optional(string, "")

          # The absolute path the volume mounts at.
          path = optional(string, "")
        })), [])
      }))

      # An explicit source for the build. Usually empty: a triggered build
      # builds the source of its trigger.
      source = optional(object({
        # A Cloud Source Repository revision.
        repo_source = optional(object({
          # The project that owns the repository. Empty means the build's
          # project.
          project_id = optional(string, "")

          # The repository's name. Required.
          repo_name = string

          # The directory, relative to the repository root, the build runs in.
          dir = optional(string, "")

          # An RE2 regular expression of the branch to build. Exactly one of
          # branch_name, tag_name, or commit_sha.
          branch_name = optional(string, "")

          # An RE2 regular expression of the tag to build. Exactly one of
          # branch_name, tag_name, or commit_sha.
          tag_name = optional(string, "")

          # An explicit commit SHA. Exactly one of branch_name, tag_name, or
          # commit_sha.
          commit_sha = optional(string, "")

          # Build revisions that do NOT match the branch or tag expression.
          invert_regex = optional(bool, false)

          # Substitutions for a triggered build of this source.
          substitutions = optional(map(string), {})
        }))

        # A gzipped tarball in Cloud Storage.
        storage_source = optional(object({
          # The bucket's name. Required.
          bucket = string

          # The object's name: a .tar.gz of the source. Required.
          object = string

          # The object's generation. Empty means the latest.
          generation = optional(string, "")
        }))
      }))

      # Secret Manager secrets exposed to the steps as environment variables
      # (listed in a step's secret_env).
      available_secrets = optional(object({
        # One entry per secret environment variable. At least one.
        secret_manager = list(object({
          # The environment variable's name; list it in each step's secret_env
          # that reads it. Unique across the build's secrets. Required.
          env = string

          # The secret version, as
          # projects/{project}/secrets/{secret}/versions/{version}: a
          # GcpSecretManagerSecret reference (its latest_version_name output) or a
          # literal ("versions/latest" follows rotation). The build's service
          # account needs roles/secretmanager.secretAccessor on the secret.
          # Required.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          version_name = string
        }))
      }))

      # Cloud KMS-encrypted values exposed to the steps as environment
      # variables (listed in a step's secret_env). Prefer available_secrets.
      secrets = optional(list(object({
        # The key that decrypts the values, as
        # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}:
        # a GcpKmsKey reference (its key_id output) or a literal. The build's
        # service account needs roles/cloudkms.cryptoKeyDecrypter on it.
        # Required.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key_name = string

        # Environment variable name to its base64 ciphertext, encrypted with
        # kms_key_name (at most 64 KB each, 100 across the build).
        secret_env = optional(map(string), {})
      })), [])
    }))

    # The path, from the repository root, of the build configuration file
    # (e.g. "cloudbuild.yaml") in the triggering repository. For
    # repository_event_config, github, bitbucket_server_trigger_config,
    # developer_connect_event_config, and trigger_template triggers; Pub/Sub,
    # webhook, and manual triggers use git_file_source. Exactly one of
    # build, filename, or git_file_source.
    filename = optional(string, "")

    # The build configuration file read from a named repository and
    # revision -- what Pub/Sub, webhook, and manual triggers use. Exactly
    # one of build, filename, or git_file_source.
    git_file_source = optional(object({
      # The file's path from the repository root, e.g. "cloudbuild.yaml".
      # Required.
      path = string

      # The kind of repository: "GITHUB", "BITBUCKET_SERVER",
      # "CLOUD_SOURCE_REPOSITORIES", or "UNKNOWN". Required.
      repo_type = string

      # A repository linked through a Cloud Build connection: a
      # GcpCloudBuildRepository reference (its name output) or a literal
      # projects/{project}/locations/{location}/connections/{connection}/repositories/{repository}.
      # At most one of repository or uri; with neither, the file is read from
      # the repository the event came from.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      repository = optional(string, "")

      # The repository's URI, e.g. "https://github.com/acme/orders". At most
      # one of repository or uri.
      uri = optional(string, "")

      # The branch, tag, ref, or SHA to read the file at (git revision
      # syntax), e.g. "refs/heads/main". Empty means the revision that
      # triggered the build.
      revision = optional(string, "")

      # The GitHub Enterprise config the repository is reached through, as
      # projects/{project}/locations/{location}/githubEnterpriseConfigs/{id}.
      github_enterprise_config = optional(string, "")

      # The Bitbucket Server config the repository is reached through, as
      # projects/{project}/locations/{location}/bitbucketServerConfigs/{id}.
      bitbucket_server_config = optional(string, "")
    }))

    # What destroy does:
    #   "" / "DELETE" -- the trigger is deleted
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the trigger leaves management and stays in Google
    deletion_policy = optional(string, "")
  })
}
