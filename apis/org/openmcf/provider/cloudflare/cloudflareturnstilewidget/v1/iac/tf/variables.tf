variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "CloudflareTurnstileWidgetSpec defines a Turnstile widget"
  type = object({
    # (Required) The Cloudflare account ID that owns the widget.
    account_id = string

    # (Required) Human-readable widget name.
    name = string

    # (Required) Domains the widget may be served on (at least one).
    domains = list(string)

    # (Required) "non-interactive", "invisible", or "managed".
    mode = string

    # (Optional) "no_clearance", "jschallenge", "managed", or "interactive".
    clearance_level = optional(string, "")

    # (Optional) Enterprise-only flags.
    bot_fight_mode = optional(bool, false)
    ephemeral_id   = optional(bool, false)
    offlabel       = optional(bool, false)

    # (Optional) "world" (default) or "china". Immutable.
    region = optional(string, "")
  })
}
