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
  description = "AwsMskServerlessCluster specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # vpc_configs are the VPC placements for the cluster -- AWS provisions
    # client-facing network interfaces in EACH declared VPC, so applications in
    # every listed VPC connect privately without peering or PrivateLink setup.
    # One entry (the common shape) places the cluster in a single VPC;
    # additional entries extend private access to more VPCs in the same
    # account and region.
    # ForceNew: adding, removing, or changing a placement forces cluster
    # replacement.
    vpc_configs = list(object({
      # subnet_ids are the VPC subnets where the cluster places its network
      # interfaces. Provide subnets in at least two Availability Zones for
      # production use; clients connect through these ENIs. All subnets in one
      # entry must belong to the same VPC.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnet_ids = list(string)

      # security_group_ids are the security groups ATTACHED to this placement's
      # network interfaces -- they define what can reach the brokers from this
      # VPC. The ingress rule for the SASL/IAM listener port (9098) belongs on
      # these referenced AwsSecurityGroup nodes. Maximum 5. If omitted, AWS
      # attaches the VPC's default security group.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      security_group_ids = optional(list(string), [])
    }))
  })
}
