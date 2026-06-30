# A Cloudflare Turnstile widget. Optional flags are sent only when set (false
# booleans and empty strings become null) so the provider applies its own
# defaults, matching the Pulumi module byte-for-byte.
resource "cloudflare_turnstile_widget" "main" {
  account_id      = var.spec.account_id
  name            = var.spec.name
  domains         = var.spec.domains
  mode            = var.spec.mode
  clearance_level = try(var.spec.clearance_level, "") != "" ? var.spec.clearance_level : null
  bot_fight_mode  = try(var.spec.bot_fight_mode, false) ? true : null
  ephemeral_id    = try(var.spec.ephemeral_id, false) ? true : null
  offlabel        = try(var.spec.offlabel, false) ? true : null
  region          = try(var.spec.region, "") != "" ? var.spec.region : null
}
