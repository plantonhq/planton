variable "metadata" {
  description = "metadata"
  type = object({

    # name of the resource
    name = string

    # id of the resource
    id = string

    # id of the organization to which the api-resource belongs to
    org = string

    # environment to which the resource belongs to
    env = string

    # labels for the resource
    labels = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # annotations for the resource
    annotations = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # tags for the resource
    tags = list(string)
  })
}

variable "spec" {
  description = "spec"
  type = object({

    # The AWS region where the resource will be created.
    region = string

    # repositoryName is the name of the ECR repository. Must be unique within
    # the AWS account and region. Commonly includes the microservice or project name.
    # Example: "github.com/team-blue/my-microservice"
    repository_name = string

    # indicates whether image tags can be overwritten (MUTABLE) or not (IMMUTABLE).
    image_immutable = bool

    # encryptionType determines how ECR encrypts images at rest. Default is AES256,
    # using AWS-managed encryption. Use KMS to specify your own KMS key for compliance.
    # EncryptionType determines how images are encrypted at rest in ECR.
    # By default, AWS uses AES-256 (service-managed keys). Choose KMS to use
    # a customer-managed key (CMK).
    encryption_type = string

    # kmsKeyId is the ARN or ID of a KMS key used when encryption_type = KMS.
    # If omitted, AWS uses the default service-managed key for ECR.
    # Ignored if encryption_type = AES256 (StringValueOrRef).
    kms_key_id = optional(object({ value = string }))

    # forceDelete, if true, allows deleting the repository even when it contains
    # images (all images get removed on delete). By default, it is false, preventing
    # accidental data loss.
    force_delete = bool
  })
}