variable "metadata" {
  description = "Catalog object metadata"
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
  description = "Auth0PromptScreenPartials specification"
  type = object({
    prompt_type = string
    screen_partials = list(object({
      screen_name = string
      insertion_points = object({
        form_content            = optional(string, "")
        form_content_start      = optional(string, "")
        form_content_end        = optional(string, "")
        form_footer_start       = optional(string, "")
        form_footer_end         = optional(string, "")
        secondary_actions_start = optional(string, "")
        secondary_actions_end   = optional(string, "")
      })
    }))
  })
}
