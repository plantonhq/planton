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
  description = "AwsEksAccessEntry specification"
  type = object({
    # The AWS region the entry's cluster lives in. Must match the
    # cluster's region. Example: "us-west-2", "eu-west-1".
    region = string

    # The EKS cluster the principal gets access to. Reference an
    # AwsEksCluster's name output or pass a literal cluster name for a
    # cluster managed outside Planton. Create-only in AWS.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    cluster_name = string

    # The IAM principal being granted access. Reference an AwsIamRole's
    # role_arn output, or pass a literal role or user ARN (IAM users work
    # as literals -- roles are the norm for team and workload access).
    # Create-only in AWS: one entry per principal per cluster.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    principal_arn = string

    # The entry type. Empty or "STANDARD" is the human/workload entry this
    # kind exists for. The node types -- "EC2", "EC2_LINUX", "EC2_WINDOWS",
    # "FARGATE_LINUX", "HYBRID_LINUX" -- exist for infrastructure roles
    # (EKS creates them automatically for managed node groups and Fargate
    # profiles; set one only when registering self-managed or hybrid
    # nodes). AWS forbids groups, username, and policy associations on
    # non-STANDARD entries (enforced below). Create-only in AWS.
    type = optional(string, "")

    # Kubernetes groups the principal is mapped onto -- your own RBAC
    # (Cluster)RoleBindings reference these group names; nothing is
    # created in-cluster for you. Groups may not start with the reserved
    # "system:" prefix. STANDARD entries only. Updates in place.
    kubernetes_groups = optional(list(string), [])

    # The Kubernetes username the principal authenticates as, visible in
    # audit logs and usable in RBAC bindings. Empty lets AWS default it
    # (the principal ARN for users; a session-templated name for roles --
    # the right choice for audit trails, since it preserves the session
    # name). STANDARD entries only. Updates in place.
    user_name = optional(string, "")

    # AWS-managed EKS access policies attached to the principal,
    # cluster-scoped or namespace-scoped -- authorization without any
    # in-cluster RBAC objects. STANDARD entries only. Associations add,
    # change scope, and remove in place.
    policy_associations = optional(list(object({
      # The access policy's ARN. These are AWS-managed and account-less --
      # "arn:aws:eks::aws:cluster-access-policy/AmazonEKSViewPolicy",
      # ...EditPolicy, ...AdminPolicy (namespace-level admin),
      # ...ClusterAdminPolicy (full cluster admin), plus service-specific
      # policies. `aws eks list-access-policies` shows the catalog; custom
      # policies do not exist.
      policy_arn = string

      # What the policy's permissions apply to: the whole cluster or a set
      # of namespaces.
      access_scope = object({
        # "cluster" applies the policy cluster-wide; "namespace" restricts it
        # to the listed namespaces.
        type = string

        # The namespaces the policy applies to when type is "namespace" --
        # e.g. give a team AmazonEKSAdminPolicy inside its own namespaces
        # only. Trailing "*" wildcards are allowed ("team-*").
        namespaces = optional(list(string), [])
      })
    })), [])
  })
}
