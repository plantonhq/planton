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
  description = "AwsEksFargateProfile specification"
  type = object({
    # The AWS region the profile's cluster lives in. Must match the
    # cluster's region. Example: "us-west-2", "eu-west-1".
    region = string

    # The EKS cluster the profile attaches to. Reference an
    # AwsEksCluster's name output or pass a literal cluster name for a
    # cluster managed outside Planton. Create-only in AWS.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    cluster_name = string

    # The IAM role Fargate uses to run the matched pods -- pulling images,
    # writing logs. It must trust "eks-fargate-pods.amazonaws.com" and
    # carry AmazonEKSFargatePodExecutionRolePolicy -- attach it on the
    # AwsIamRole itself; this component never modifies a role it merely
    # references. Reference an AwsIamRole's role_arn output or pass a
    # literal ARN. Create-only in AWS.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    pod_execution_role_arn = string

    # The subnets Fargate launches the matched pods into. PRIVATE subnets
    # only -- AWS rejects subnets whose route table carries an internet
    # gateway route; give the pods outbound internet through a NAT
    # gateway. Reference AwsSubnet subnet_id outputs or pass literal
    # subnet IDs. Create-only in AWS. A Fargate profile is a member of
    # its cluster and lives there on a diagram; the subnets are where its
    # pods land, so the reference is access, not placement -- otherwise a
    # profile on subnets its cluster does not name would be drawn outside
    # the cluster it belongs to.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = list(string)

    # Which pods run on Fargate: a pod matches the profile when it matches
    # ANY selector (namespace, plus every label when labels are given).
    # AWS allows at most 5 selectors per profile. Create-only in AWS.
    selectors = list(object({
      # The Kubernetes namespace to match. Wildcards are allowed -- "*"
      # matches any sequence, "?" any single character (e.g. "prod-*").
      namespace = string

      # Labels a pod must ALL carry to match (AND semantics within a
      # selector). Values may use the same "*" / "?" wildcards. Empty
      # matches every pod in the namespace. AWS allows at most 5 label pairs
      # per selector.
      labels = optional(map(string), {})
    }))
  })
}
