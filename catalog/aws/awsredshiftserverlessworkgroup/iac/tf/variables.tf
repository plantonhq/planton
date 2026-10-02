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
  description = "AwsRedshiftServerlessWorkgroup specification"
  type = object({
    # The AWS region the workgroup is created in. Must match the region
    # of the namespace and of the subnets and security groups it
    # references.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # The namespace this workgroup serves -- the data plane behind this
    # compute. Create-time only: moving a workgroup to another namespace
    # replaces it. Reference an AwsRedshiftServerlessNamespace
    # namespace_name output or pass a literal namespace name.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    namespace_name = string

    # The baseline compute the workgroup starts every query with, in
    # Redshift Processing Units. 0 keeps the AWS default (128 RPU).
    # AWS currently accepts 4-1024 RPU (values above 512 in units of 8;
    # small values in the 4/8-RPU steps AWS publishes) -- the exact
    # increments have changed over time, so they are validated by AWS at
    # deploy rather than frozen here. Higher base = faster first query,
    # higher floor cost per second of execution. Mutually exclusive with
    # an enabled price_performance_target, where AWS picks the baseline.
    base_capacity = optional(number, 0)

    # A hard ceiling on the compute the workgroup may scale to, in RPUs.
    # 0 leaves scaling uncapped (the AWS default) -- the workgroup grows
    # to whatever the query mix demands. Set it to bound worst-case
    # spend; must be at least base_capacity when both are set.
    max_capacity = optional(number, 0)

    # Let AWS choose capacity against a price-performance dial instead of
    # a fixed base: level 1 leans cheapest, 100 leans fastest, 50 is
    # balanced. When enabled, base_capacity must stay unset (AWS owns the
    # baseline); max_capacity still caps spend.
    price_performance_target = optional(object({
      # Turn price-performance targeting on. While enabled, AWS owns the
      # capacity baseline and base_capacity must stay unset.
      enabled = optional(bool, false)

      # Where on the dial AWS should aim: 1 (cheapest), 25, 50 (balanced),
      # 75, or 100 (fastest). 0 keeps the AWS default (50).
      level = optional(number, 0)
    }))

    # Subnets the workgroup places its compute (and managed VPC endpoint)
    # in. AWS requires at least THREE subnets spanning three distinct
    # availability zones; the free-IP requirement per subnet scales with
    # base capacity. Empty lets AWS use the account's default VPC (only
    # meaningful in accounts that still have one). Reference AwsSubnet
    # subnet_id outputs or pass literal subnet IDs.
    #
    # Containment-exempt: on a diagram the workgroup lives inside the
    # namespace it serves (the box AWS's own model draws around its
    # workgroups); the subnets are where its network interfaces land, and
    # the workgroup reaches into that VPC by lines -- the same verdict an
    # ECS service carries for its subnets while its cluster places it.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = optional(list(string), [])

    # Security groups attached to the workgroup's endpoint. Empty uses
    # the VPC's default security group (the AWS default). Reference
    # AwsSecurityGroup security_group_id outputs or pass literal SG IDs
    # -- warehouse ingress rules (e.g. port 5439 from your BI tooling)
    # belong on the referenced AwsSecurityGroup node, never inside this
    # workgroup.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # Force all COPY and UNLOAD traffic between the workgroup and data
    # repositories (S3, DynamoDB, ...) through the VPC instead of the
    # public internet -- enabling VPC flow logs, endpoints, and other
    # network controls to see and govern warehouse data movement.
    enhanced_vpc_routing = optional(bool, false)

    # Give the workgroup's endpoint a public IP so it can be reached from
    # outside the VPC. Off by default -- serverless warehouses almost
    # always stay private behind VPC routing (and Query Editor / private
    # BI reach them fine).
    publicly_accessible = optional(bool, false)

    # The port the workgroup accepts connections on. 0 keeps the AWS
    # default (5439). Redshift Serverless only accepts ports within
    # 5431-5455 or 8191-8215.
    port = optional(number, 0)

    # Query-level configuration parameters (e.g. "require_ssl",
    # "max_query_execution_time", "search_path"), applied directly to the
    # workgroup -- serverless has no parameter groups, so there is
    # nothing to fold or reference. The name list mirrors what the
    # Redshift Serverless API accepts.
    config_parameters = optional(list(object({
      # The parameter name. The Redshift Serverless API accepts exactly
      # this set; workload-limit parameters (max_query_*, max_scan_*,
      # max_*_row_count) implement query monitoring rules.
      name = string

      # The parameter value. Required.
      value = string
    })), [])

    # The release track the workgroup follows: "current" (the AWS
    # default -- the latest certified release) or "trailing" (one
    # certified release behind); AWS also accepts named preview tracks.
    # Empty keeps the AWS default.
    track_name = optional(string, "")

    # A custom DNS name for the workgroup's endpoint, fronted by an ACM
    # certificate (one custom domain per workgroup -- AWS's own model).
    # You own the CNAME record pointing the domain at the workgroup
    # endpoint; AWS serves TLS for it from the certificate.
    custom_domain = optional(object({
      # The custom domain name (e.g. "warehouse.example.com"), 1-253
      # characters. Changing it replaces the association.
      domain_name = string

      # The ACM certificate that serves TLS for the domain -- it must
      # cover the domain name and live in the workgroup's region.
      # Reference an AwsCertManagerCert cert_arn output or pass a literal
      # certificate ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      certificate_arn = string
    }))

    # VPC endpoints exposing this workgroup inside other subnets --
    # same-account cross-VPC access without peering. Each entry keys one
    # managed endpoint; its private address and port are exported per
    # endpoint on the outputs contract.
    endpoint_accesses = optional(list(object({
      # The endpoint's name: 1-30 characters, unique within the workgroup.
      # Changing it replaces the endpoint.
      endpoint_name = string

      # Subnets the endpoint's network interfaces land in -- typically the
      # CONSUMING VPC's subnets. Empty reuses the workgroup's own
      # subnet_ids (which must then be set -- CEL-enforced). Reference
      # AwsSubnet subnet_id outputs or pass literal subnet IDs. Changing
      # the list replaces the endpoint.
      #
      # Containment-exempt: these are the consuming VPC's subnets, an access
      # path INTO the workgroup; the workgroup is not deployed into them.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnet_ids = optional(list(string), [])

      # Security groups attached to the endpoint's network interfaces.
      # Empty uses the VPC's default security group. Reference
      # AwsSecurityGroup security_group_id outputs or pass literal SG IDs.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      vpc_security_group_ids = optional(list(string), [])
    })), [])

    # Usage limits capping this workgroup's consumption -- RPU-hours of
    # serverless compute or terabytes of cross-region datasharing
    # transfer -- each with a breach action from logging to deactivation.
    usage_limits = optional(list(object({
      # What the limit measures: "serverless-compute" (RPU-hours of
      # compute) or "cross-region-datasharing" (terabytes transferred to
      # consumers in other regions). Required.
      usage_type = string

      # The limit amount: RPU-hours for "serverless-compute", terabytes
      # for "cross-region-datasharing". Must be positive.
      amount = optional(number, 0)

      # The period the amount applies to: "daily", "weekly", or "monthly".
      # Empty keeps the AWS default (monthly). A weekly period begins on
      # Sunday. Changing it replaces the limit.
      period = optional(string, "")

      # What Redshift Serverless does when the limit is breached: "log"
      # writes an event to the system table, "emit-metric" additionally
      # publishes a CloudWatch metric, "deactivate" turns queries off
      # until the period resets. Empty keeps the AWS default (log). Note
      # the serverless action vocabulary differs from provisioned
      # clusters' ("deactivate", not "disable").
      breach_action = optional(string, "")
    })), [])
  })
}
