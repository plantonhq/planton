variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "CloudflareZeroTrustDevicePostureRule specification"
  type = object({
    account_id = string
    name = string
    type = string
    description = optional(string, "")
    expiration = optional(string, "")
    schedule = optional(string, "")
    match = optional(list(object({
      platform = string
    })), [])
    input = optional(object({
      operating_system = optional(string, "")
      path = optional(string, "")
      exists = optional(bool)
      sha256 = optional(string, "")
      thumbprint = optional(string, "")
      id = optional(string, "")
      domain = optional(string, "")
      operator = optional(string, "")
      version = optional(string, "")
      os_distro_name = optional(string, "")
      os_distro_revision = optional(string, "")
      os_version_extra = optional(string, "")
      enabled = optional(bool)
      check_disks = optional(list(string), [])
      require_all = optional(bool)
      certificate_id = optional(string, "")
      cn = optional(string, "")
      check_private_key = optional(bool)
      extended_key_usage = optional(list(string), [])
      locations = optional(object({
        paths = optional(list(string), [])
        trust_stores = optional(list(string), [])
      }))
      subject_alternative_names = optional(list(string), [])
      update_window_days = optional(number)
      compliance_status = optional(string, "")
      connection_id = optional(string, "")
      last_seen = optional(string, "")
      os = optional(string, "")
      overall = optional(string, "")
      sensor_config = optional(string, "")
      state = optional(string, "")
      version_operator = optional(string, "")
      auth_state = optional(list(string), [])
      count_operator = optional(string, "")
      issue_count = optional(string, "")
      eid_last_seen = optional(string, "")
      risk_level = optional(string, "")
      score_operator = optional(string, "")
      total_score = optional(number)
      active_threats = optional(number)
      infected = optional(bool)
      is_active = optional(bool)
      network_status = optional(string, "")
      operational_state = optional(string, "")
      score = optional(number)
    }))
  })
}