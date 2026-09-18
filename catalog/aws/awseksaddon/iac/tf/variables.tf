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
  description = "AwsEksAddon specification"
  type = object({
    # The AWS region the add-on's cluster lives in. Must match the
    # cluster's region. Example: "us-west-2", "eu-west-1".
    region = string

    # The EKS cluster the add-on installs on. Reference an AwsEksCluster's
    # name output or pass a literal cluster name for a cluster managed
    # outside Planton. Create-only in AWS.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    cluster_name = string

    # The add-on to install, by its EKS catalog name: AWS-built add-ons
    # ("vpc-cni", "coredns", "kube-proxy", "aws-ebs-csi-driver",
    # "aws-efs-csi-driver", "eks-pod-identity-agent", "snapshot-controller",
    # "amazon-cloudwatch-observability", ...) or a marketplace add-on's
    # vendor-prefixed name. `aws eks describe-addon-versions` lists what the
    # cluster's Kubernetes version supports. Create-only in AWS -- changing
    # the name replaces the add-on.
    addon_name = string

    # The add-on version to run, e.g. "v1.18.1-eksbuild.3". Empty installs
    # the AWS default version for the cluster's Kubernetes version -- the
    # never-goes-stale choice; pin a version for byte-identical clusters
    # and controlled upgrades. Version changes update in place (the add-on
    # rolls its own pods).
    addon_version = optional(string, "")

    # How to resolve conflicts when the add-on's Kubernetes resources
    # already exist on the cluster at install time -- typically because
    # the cluster bootstrapped self-managed copies of vpc-cni / coredns /
    # kube-proxy at creation. "OVERWRITE" adopts and overwrites the
    # existing install (the standard way to migrate a self-managed add-on
    # to a managed one); "NONE" (the default when empty) fails the install
    # with a conflict instead. AWS accepts only these two at create time
    # -- "PRESERVE" exists only for updates.
    resolve_conflicts_on_create = optional(string, "")

    # How to resolve drift when an update finds fields changed out-of-band
    # (e.g. someone kubectl-edited the add-on's config). "OVERWRITE"
    # restores the managed values, "PRESERVE" keeps the hand-made changes,
    # "NONE" (the default when empty) fails the update on conflict.
    resolve_conflicts_on_update = optional(string, "")

    # Custom configuration for the add-on as a single JSON document (e.g.
    # '{"replicaCount":3}' for coredns). Each add-on publishes its own
    # schema -- `aws eks describe-addon-configuration` shows what is
    # configurable. Empty keeps every add-on default. Updates in place.
    configuration_values = optional(string, "")

    # The IAM role the add-on's service account assumes via IRSA. Requires
    # an IAM OIDC provider for the cluster (AwsIamOidcProvider on the
    # cluster's oidc_issuer_url output); without it AWS rejects the
    # install. Empty means the add-on's pods fall back to the node role's
    # permissions. Prefer pod_identity_associations on new clusters --
    # Pod Identity needs no per-cluster OIDC provider. Reference an
    # AwsIamRole's role_arn output or pass a literal ARN. Updates in place.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    service_account_role_arn = optional(string, "")

    # EKS Pod Identity wiring: bind the add-on's service account(s) to IAM
    # roles through the Pod Identity agent (the modern, no-OIDC-provider
    # alternative to IRSA). The role must trust
    # "pods.eks.amazonaws.com"; the eks-pod-identity-agent add-on must be
    # installed on the cluster. Updates in place.
    pod_identity_associations = optional(list(object({
      # The IAM role the service account assumes. It must trust
      # "pods.eks.amazonaws.com" and carry the permissions the add-on's
      # documentation requires (e.g. the EBS CSI driver policy). Reference
      # an AwsIamRole's role_arn output or pass a literal ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = string

      # The Kubernetes service account (in the add-on's namespace) the role
      # binds to -- each add-on documents its service account name(s), e.g.
      # "ebs-csi-controller-sa" for the EBS CSI driver.
      service_account = string
    })), [])

    # Keep the add-on's Kubernetes resources on the cluster when this
    # resource is deleted -- AWS stops managing them but leaves them
    # running (they become self-managed). Off by default: deletion removes
    # the software. Turn it on when handing an add-on's lifecycle back to
    # cluster operators without an outage.
    preserve = optional(bool, false)

    # Install the add-on into a custom namespace instead of its default.
    # Create-only in AWS: changing the namespace requires removing and
    # re-creating the add-on. Only some add-ons support this.
    namespace_config = optional(object({
      # The Kubernetes namespace to install the add-on into. Must be a valid
      # RFC 1123 DNS label. Create-only in AWS.
      namespace = string
    }))
  })
}
