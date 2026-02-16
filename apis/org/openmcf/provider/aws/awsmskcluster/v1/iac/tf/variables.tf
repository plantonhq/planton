variable "provider_config" {
  description = "AWS provider configuration"
  type = object({
    region            = string
    access_key_id     = optional(string)
    secret_access_key = optional(string)
    session_token     = optional(string)
  })
}

variable "metadata" {
  description = "Resource metadata"
  type = object({
    id   = string
    name = string
    org  = optional(string, "")
    env  = optional(string, "")
  })
}

variable "spec" {
  description = "AwsMskClusterSpec - desired state of the MSK cluster"
  type = object({
    kafka_version          = string
    number_of_broker_nodes = number
    instance_type          = string

    subnet_ids                   = list(string)
    security_group_ids           = optional(list(string), [])
    allowed_cidr_blocks          = optional(list(string), [])
    associate_security_group_ids = optional(list(string), [])
    vpc_id                       = optional(string, "")

    ebs_volume_size_gib            = optional(number, null)
    provisioned_throughput_enabled  = optional(bool, false)
    provisioned_throughput_mbs      = optional(number, 0)
    storage_mode                   = optional(string, "")

    kms_key_arn              = optional(string, "")
    client_broker_encryption = optional(string, "TLS")
    in_cluster_encryption    = optional(bool, true)

    authentication = optional(object({
      sasl_iam_enabled                = optional(bool, false)
      sasl_scram_enabled              = optional(bool, false)
      tls_enabled                     = optional(bool, false)
      tls_certificate_authority_arns  = optional(list(string), [])
      unauthenticated                 = optional(bool, false)
    }), null)

    configuration_arn      = optional(string, "")
    configuration_revision = optional(number, 0)
    server_properties      = optional(map(string), {})

    logging = optional(object({
      cloudwatch_logs = optional(object({
        enabled   = bool
        log_group = optional(string, "")
      }), null)
      firehose = optional(object({
        enabled         = bool
        delivery_stream = optional(string, "")
      }), null)
      s3 = optional(object({
        enabled = bool
        bucket  = optional(string, "")
        prefix  = optional(string, "")
      }), null)
    }), null)

    enhanced_monitoring   = optional(string, "DEFAULT")
    jmx_exporter_enabled  = optional(bool, false)
    node_exporter_enabled = optional(bool, false)
    public_access_type    = optional(string, "")
  })
}
