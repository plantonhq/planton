variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name   = string
    id     = optional(string)
    org    = optional(string)
    env    = optional(string)
    labels = optional(map(string))
    tags   = optional(list(string))
  })
}

variable "spec" {
  description = "Alibaba Cloud KMS Key specification"
  type = object({
    region                           = string
    description                      = optional(string, "")
    key_spec                         = optional(string, "Aliyun_AES_256")
    key_usage                        = optional(string, "ENCRYPT/DECRYPT")
    protection_level                 = optional(string, "SOFTWARE")
    automatic_rotation               = optional(bool, false)
    rotation_interval                = optional(string, "")
    pending_window_in_days           = optional(number, 30)
    deletion_protection              = optional(bool, false)
    deletion_protection_description  = optional(string, "")
    tags                             = optional(map(string), {})
  })

  validation {
    condition = contains([
      "Aliyun_AES_256", "Aliyun_AES_128", "Aliyun_AES_192", "Aliyun_SM4",
      "RSA_2048", "RSA_3072", "EC_P256", "EC_P256K", "EC_SM2"
    ], var.spec.key_spec)
    error_message = "key_spec must be one of: Aliyun_AES_256, Aliyun_AES_128, Aliyun_AES_192, Aliyun_SM4, RSA_2048, RSA_3072, EC_P256, EC_P256K, EC_SM2."
  }

  validation {
    condition     = contains(["ENCRYPT/DECRYPT", "SIGN/VERIFY"], var.spec.key_usage)
    error_message = "key_usage must be one of: ENCRYPT/DECRYPT, SIGN/VERIFY."
  }

  validation {
    condition     = contains(["SOFTWARE", "HSM"], var.spec.protection_level)
    error_message = "protection_level must be one of: SOFTWARE, HSM."
  }

  validation {
    condition     = var.spec.pending_window_in_days >= 7 && var.spec.pending_window_in_days <= 366
    error_message = "pending_window_in_days must be between 7 and 366."
  }
}
