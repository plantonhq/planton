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
  description = "CloudflareBotManagement specification"
  type = object({
    zone_id = string
    fight_mode = optional(bool)
    sbfm_definitely_automated = optional(string)
    sbfm_likely_automated = optional(string)
    sbfm_verified_bots = optional(string)
    sbfm_static_resource_protection = optional(bool)
    optimize_wordpress = optional(bool)
    auto_update_model = optional(bool)
    suppress_session_score = optional(bool)
    enable_js = optional(bool)
    bm_cookie_enabled = optional(bool)
    ai_bots_protection = optional(string)
    crawler_protection = optional(string)
    content_bots_protection = optional(string)
    cf_robots_variant = optional(string)
    is_robots_txt_managed = optional(bool)
  })
}