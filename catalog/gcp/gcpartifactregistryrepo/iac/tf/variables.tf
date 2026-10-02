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
  description = "GcpArtifactRegistryRepo specification"
  type = object({
    # The GCP project where the repository is created.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The last segment of the repository's resource name — the ID that appears
    # in registry URLs (e.g. "us-docker.pkg.dev/{project}/{repository_id}").
    # Must start with a letter, contain only lowercase letters, numbers, and
    # hyphens, and be at most 63 characters. If omitted, metadata.name is used.
    # Immutable after creation.
    repository_id = optional(string, "")

    # The location for the repository: a region (e.g. "us-central1",
    # "asia-south1") or a multi-region ("us", "europe", "asia"). Regional
    # repositories co-located with your build/runtime infrastructure minimize
    # pull latency and egress cost; multi-region maximizes availability for
    # globally consumed artifacts. Immutable after creation.
    location = string

    # The package format stored in this repository. Common values:
    #   "DOCKER"  -- container images (Docker/OCI); the format behind
    #                us-docker.pkg.dev URLs and GKE/Cloud Run image pulls
    #   "MAVEN"   -- Java/JVM packages
    #   "NPM"     -- Node.js packages
    #   "PYTHON"  -- Python packages (pip/PyPI layout)
    #   "GO"      -- Go module proxy (remote mode caches proxy.golang.org)
    #   "APT"     -- Debian/Ubuntu OS packages
    #   "YUM"     -- RHEL/CentOS/Rocky OS packages
    #   "GENERIC" -- arbitrary versioned files
    #   "KFP"     -- Kubeflow Pipelines templates
    #
    # Artifact Registry adds formats over time, so this field deliberately
    # accepts any string (matched case-insensitively by GCP) and lets the API
    # validate — see the repository format reference for the authoritative
    # list: https://cloud.google.com/artifact-registry/docs/supported-formats
    # Immutable after creation.
    format = string

    # The serving mode of the repository. Defaults to STANDARD_REPOSITORY.
    #
    #   "STANDARD_REPOSITORY" -- artifacts are pushed directly to this repo
    #   "REMOTE_REPOSITORY"   -- pull-through cache of one upstream source
    #                            (requires remote_repository_config)
    #   "VIRTUAL_REPOSITORY"  -- priority-ordered aggregation of other
    #                            Artifact Registry repositories (requires
    #                            virtual_repository_config)
    #
    # Immutable after creation.
    mode = optional(string, "")

    # Human-readable description of the repository's purpose, shown in the
    # console and API listings. Mutable in place.
    description = optional(string, "")

    # User-defined labels attached to the repository, for cost attribution
    # and fleet queries. Merged with Planton's platform labels (which win on
    # key conflicts). Mutable in place.
    labels = optional(map(string), {})

    # Customer-managed encryption key (CMEK) protecting artifacts in this
    # repository. Accepts the fully qualified crypto key path
    #   projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
    # or a reference to a GcpKmsKey resource. The Artifact Registry service
    # agent needs roles/cloudkms.cryptoKeyEncrypterDecrypter on the key.
    # If omitted, artifacts are encrypted with Google-managed keys.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # Docker-format-specific configuration. Only meaningful when format is
    # DOCKER. Mutable in place.
    docker_config = optional(object({
      # If true, image tags cannot be modified, moved, or deleted once created
      # — a pushed tag permanently identifies one digest. New tags can still be
      # created. Immutable tags make deployments reproducible ("v1.2.3 today is
      # v1.2.3 forever") at the cost of losing mutable convenience tags like
      # "latest". Mutable in place.
      immutable_tags = optional(bool, false)
    }))

    # Maven-format-specific configuration. Only meaningful when format is
    # MAVEN. Effectively immutable: GCP rejects changing the version policy
    # or snapshot-overwrite behavior on an existing repository.
    maven_config = optional(object({
      # The Maven version policy for this repository:
      #   ""         -- accept both release and snapshot versions (default)
      #   "RELEASE"  -- accept only release versions
      #   "SNAPSHOT" -- accept only snapshot versions
      #   "VERSION_POLICY_UNSPECIFIED" -- the API's explicit mixed-versions
      #                sentinel (equivalent to leaving this unset)
      # The conventional Maven setup is a RELEASE repository and a SNAPSHOT
      # repository, with builds publishing to each as appropriate.
      version_policy = optional(string, "")

      # If true, re-publishing a non-snapshot artifact at an existing version
      # overwrites it. Defaults to false — GCP rejects duplicate uploads,
      # preserving release immutability (the safer posture for RELEASE repos).
      allow_snapshot_overwrites = optional(bool, false)
    }))

    # Cleanup policies that automatically delete or protect artifact versions
    # based on age, tag state, or name prefixes. Policies with action DELETE
    # remove matching versions; policies with action KEEP protect matching
    # versions from every DELETE policy (KEEP always wins on overlap). Without
    # cleanup policies, CI-pushed repositories grow without bound — pair every
    # busy standard repository with at least a delete-untagged policy.
    # Mutable in place.
    cleanup_policies = optional(list(object({
      # Unique identifier for this policy within the repository.
      id = string

      # What to do with versions matching this policy:
      #   "DELETE" -- delete matching versions
      #   "KEEP"   -- protect matching versions from all DELETE policies
      # When a version matches both a DELETE and a KEEP policy, KEEP wins.
      action = string

      # Selects versions by age, tag state, and name prefixes. All specified
      # criteria must match (logical AND).
      condition = optional(object({
        # Match versions newer than this duration since upload, in seconds with
        # an "s" suffix (e.g. "2592000s" for 30 days). Typically used on KEEP
        # policies ("protect everything pushed in the last 30 days").
        newer_than = optional(string, "")

        # Match versions older than this duration since upload, in seconds with
        # an "s" suffix (e.g. "7776000s" for 90 days). The workhorse of DELETE
        # policies ("delete anything older than 90 days").
        older_than = optional(string, "")

        # Match versions whose package name starts with any of these prefixes.
        package_name_prefixes = optional(list(string), [])

        # Match versions with a tag starting with any of these prefixes
        # (e.g. "release-" to select release-tagged images).
        tag_prefixes = optional(list(string), [])

        # Match versions by tag status:
        #   "ANY"      -- tagged or untagged (default)
        #   "TAGGED"   -- only versions with at least one tag
        #   "UNTAGGED" -- only versions with no tags (superseded digests in
        #                 Docker repos — the classic cleanup target)
        tag_state = optional(string, "")

        # Match versions whose version name starts with any of these prefixes.
        version_name_prefixes = optional(list(string), [])
      }))

      # Selects the N most recent versions (per package) to protect. Only valid
      # with action KEEP — the standard "always keep the last 10 builds" guard
      # paired with an age-based DELETE policy.
      most_recent_versions = optional(object({
        # Minimum number of the most recent versions to keep per package.
        keep_count = optional(number, 0)

        # Restrict the keep-count protection to packages whose name starts with
        # any of these prefixes. If empty, applies to all packages in the repo.
        package_name_prefixes = optional(list(string), [])
      }))
    })), [])

    # If true, cleanup policies run in dry-run mode: matches are logged
    # (visible in Cloud Audit Logs) but nothing is deleted. Use this to
    # validate new policies against real traffic before letting them delete.
    # Mutable in place.
    cleanup_policy_dry_run = optional(bool, false)

    # Upstream source configuration for REMOTE_REPOSITORY mode — which
    # registry this repository caches. Required when (and only valid when)
    # mode is REMOTE_REPOSITORY. The block is immutable after creation EXCEPT
    # upstream_credentials and disable_upstream_validation, which update in
    # place (credential rotation never recreates the cache).
    remote_repository_config = optional(object({
      # Human-readable description of the upstream source. Immutable.
      description = optional(string, "")

      # Well-known public Docker upstream. The only supported value is
      # "DOCKER_HUB". For any other Docker registry (ghcr.io, quay.io, another
      # AR repository), use common_repository instead. Immutable.
      docker_public_repository = optional(string, "")

      # Well-known public Maven upstream. The only supported value is
      # "MAVEN_CENTRAL". For other Maven registries use common_repository.
      # Immutable.
      maven_public_repository = optional(string, "")

      # Well-known public npm upstream. The only supported value is "NPMJS".
      # For other npm registries use common_repository. Immutable.
      npm_public_repository = optional(string, "")

      # Well-known public Python upstream. The only supported value is "PYPI".
      # For other Python registries use common_repository. Immutable.
      python_public_repository = optional(string, "")

      # Public Apt upstream (Debian/Ubuntu mirror trees). Immutable.
      apt_repository = optional(object({
        # The mirror tree to cache from:
        #   "DEBIAN"          -- deb.debian.org
        #   "UBUNTU"          -- archive.ubuntu.com
        #   "DEBIAN_SNAPSHOT" -- snapshot.debian.org (point-in-time Debian
        #                        archives, for reproducible builds)
        repository_base = string

        # The specific repository path within the base, e.g. "debian/dists/bookworm".
        repository_path = string
      }))

      # Public Yum upstream (RHEL-family mirror trees). Immutable.
      yum_repository = optional(object({
        # The mirror tree to cache from:
        #   "CENTOS", "CENTOS_DEBUG", "CENTOS_VAULT", "CENTOS_STREAM",
        #   "ROCKY", "EPEL"
        repository_base = string

        # The specific repository path within the base,
        # e.g. "pub/rocky/9/BaseOS/x86_64/os".
        repository_path = string
      }))

      # Custom upstream: another Artifact Registry repository
      # ("projects/{p}/locations/{l}/repositories/{r}") or a registry URI
      # (e.g. "https://ghcr.io", "https://registry.company.com"). The upstream
      # must serve the same format as this repository. Immutable.
      common_repository = optional(object({
        # Another Artifact Registry repository
        # ("projects/{p}/locations/{l}/repositories/{r}") or a registry URI
        # ("https://registry.company.com"). Accepts a literal or a reference to
        # a GcpArtifactRegistryRepo resource (its repository_path output is
        # exactly the AR form of this value).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        uri = string
      }))

      # Credentials for authenticating to the upstream (private registries, or
      # Docker Hub authenticated pulls for higher rate limits). Mutable in
      # place: rotating credentials never recreates the cache.
      upstream_credentials = optional(object({
        # The username to authenticate with.
        username = string

        # Secret Manager secret version holding the password, in the form
        #   projects/{project}/secrets/{secret}/versions/{version}
        # (use ".../versions/latest" to track rotation automatically). The
        # Artifact Registry service agent needs
        # roles/secretmanager.secretAccessor on the secret.
        password_secret_version = string
      }))

      # If true, skip validating the upstream URL and credentials at
      # create/update time. Useful when the upstream is temporarily unreachable
      # from the control plane but will be reachable at pull time. Mutable in
      # place.
      disable_upstream_validation = optional(bool, false)
    }))

    # Upstream aggregation configuration for VIRTUAL_REPOSITORY mode — which
    # Artifact Registry repositories this endpoint serves from and in what
    # priority order. Required when (and only valid when) mode is
    # VIRTUAL_REPOSITORY. Mutable in place: upstreams can be added, removed,
    # and re-prioritized without recreating the virtual endpoint.
    virtual_repository_config = optional(object({
      # The Artifact Registry repositories this virtual endpoint serves from.
      # When the same package exists in multiple upstreams, the highest
      # priority wins. All upstreams must serve this repository's format.
      upstream_policies = list(object({
        # User-chosen identifier for this policy entry, unique within the
        # virtual repository.
        id = string

        # The upstream Artifact Registry repository, as the full resource path
        #   projects/{project}/locations/{location}/repositories/{repository}
        # Accepts a literal or a reference to a GcpArtifactRegistryRepo resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        repository = string

        # Serving priority — entries with higher values are tried first.
        priority = optional(number, 0)
      }))
    }))

    # Whether Artifact Analysis automatically scans artifacts in this
    # repository for vulnerabilities.
    #
    #   ""          -- inherit the project-level setting (default)
    #   "INHERITED" -- same as empty: follow the project-level setting
    #   "DISABLED"  -- never scan this repository, regardless of project config
    #
    # Scanning requires the Container Scanning API to be enabled on the
    # project and incurs per-artifact scan charges. Mutable in place.
    vulnerability_scanning_enablement = optional(string, "")

    # Additive IAM grants on this repository. Each entry grants one role to
    # one member and composes safely with grants made by other tools or
    # charts — removal subtracts only that exact (role, member) pair.
    #
    # Common roles:
    #   roles/artifactregistry.reader -- pull artifacts (grant to runtime SAs)
    #   roles/artifactregistry.writer -- push and pull (grant to CI SAs)
    #   roles/artifactregistry.repoAdmin -- manage artifacts and settings
    #
    # Public access: grant roles/artifactregistry.reader to the special
    # member "allUsers" (requires the project to allow public access).
    iam_members = optional(list(object({
      # The role to grant, e.g. "roles/artifactregistry.reader",
      # "roles/artifactregistry.writer", "roles/artifactregistry.repoAdmin",
      # or a custom role's fully-qualified name.
      role = string

      # The identity receiving the grant, in GCP IAM member format:
      #   serviceAccount:<email>  -- a service account (the most common in IaC;
      #                              reference a GcpServiceAccount resource —
      #                              its `member` output is exactly this value)
      #   user:<email> / group:<email> / domain:<domain>
      #   allUsers / allAuthenticatedUsers -- public access (grant with care)
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      member = string

      # Optional IAM Condition restricting when this grant applies. The
      # condition is part of the grant's identity: the same role with and
      # without a condition are two independent grants.
      condition = optional(object({
        # Short human-readable title identifying the condition's intent,
        # e.g. "expires-2026-12-31".
        title = string

        # The CEL condition expression, e.g.
        # request.time < timestamp("2027-01-01T00:00:00Z").
        expression = string

        # Optional longer explanation of what the condition does.
        description = optional(string, "")
      }))
    })), [])

    # What destroying this resource does to the repository (IAM grants are
    # plain bindings — they are always removed with the destroy):
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the repository and every artifact stored in it are
    #                deleted; images/packages become unpullable immediately
    #   "PREVENT" -- destroy FAILS; protects a registry that running
    #                workloads still pull from
    #   "ABANDON" -- the repository is removed from management but keeps
    #                serving artifacts in GCP
    deletion_policy = optional(string, "")
  })
}
