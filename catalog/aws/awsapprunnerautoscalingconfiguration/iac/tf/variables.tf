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
  description = "AwsAppRunnerAutoScalingConfiguration specification"
  type = object({
    # The AWS region where the auto scaling configuration will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Maximum number of concurrent requests routed to a single instance
    # before App Runner launches additional instances. Lower values give each
    # instance more headroom (better tail latency) at higher cost -- more
    # instances serve the same traffic. 1 effectively dedicates an instance
    # per request, the serverless-function posture.
    max_concurrency = optional(number)

    # Maximum number of instances App Runner scales out to during traffic
    # spikes -- the cost ceiling for the services using this configuration.
    # AWS caps this at 25 by default (a service-quota increase can raise it).
    max_size = optional(number)

    # Minimum number of instances App Runner keeps provisioned at all times.
    # These warm instances serve traffic without cold-start latency and are
    # billed for memory only (CPU is billed only while actively serving).
    # Raise above 1 for latency-sensitive services worth the standing memory
    # charge.
    min_size = optional(number)

    # Claim this configuration as the ACCOUNT-WIDE default for its region:
    # App Runner services created WITHOUT an explicit auto scaling
    # configuration use the account default. One default exists per account
    # per region -- claiming it silently displaces whichever configuration
    # held the designation (AWS marks it non-default), and the claim only
    # affects services created afterwards; existing services keep their
    # current associations. One-way at AWS: un-setting this flag (or
    # deleting this resource) does NOT restore the previous default -- AWS
    # has no restore API, so the designation stays with this configuration
    # until another one claims it. Coordinate so at most one configuration
    # per account/region sets this flag.
    set_as_account_default = optional(bool, false)
  })
}
