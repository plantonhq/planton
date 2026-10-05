# A webhook destination for notification policies. Cloudflare infers the
# destination type (slack, datadog, generic, ...) from the URL and reports
# it as an output. A plain CRUD resource (real create/update/delete; only
# the account forces replacement).
#
# `secret` is WRITE-ONLY at the API: sent on create/update, never returned
# by any read -- Cloudflare cannot drift-detect it and an imported webhook
# has no secret in state.
resource "cloudflare_notification_policy_webhooks" "main" {
  account_id = var.spec.account_id
  name       = var.spec.name
  url        = var.spec.url

  secret = try(var.spec.secret, "") != "" ? var.spec.secret : null
}

# Cloudflare's create response returns ONLY the id (measured 2026-08-27:
# POST answers {"result":{"id":...}} while the GET returns the full body),
# so the resource's computed `type` is null in state after the first apply
# and the first refresh backfills it -- an output riding the resource
# attribute would be empty on first deploy and flip on the next plan. This
# read-after-create performs the GET the create response omitted; the
# `type` output rides it instead of the resource.
data "cloudflare_notification_policy_webhooks" "main" {
  account_id = var.spec.account_id
  webhook_id = cloudflare_notification_policy_webhooks.main.id
}
