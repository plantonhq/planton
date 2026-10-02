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
  description = "AwsAppRunnerObservabilityConfiguration specification"
  type = object({
    # The AWS region where the observability configuration will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Distributed tracing configuration. When set, services referencing this
    # configuration send request traces to the configured vendor -- each
    # instance runs an OpenTelemetry collector sidecar that forwards spans.
    # When omitted, the configuration is registered without tracing (a valid
    # but inert configuration; set trace_configuration to get value from this
    # resource).
    trace_configuration = optional(object({
      # The tracing vendor. "AWSXRAY" (AWS X-Ray) is the only vendor App Runner
      # supports today; the application must be instrumented with the AWS
      # Distro for OpenTelemetry (ADOT) SDK to emit spans the collector can
      # forward.
      vendor = optional(string)
    }))
  })
}
