# CloudflareEmailRoutingZone — Research & Design Notes

## What it is

The anchor of the Email Routing family. Enabling Email Routing on a zone:

- turns the feature on (the `cloudflare_email_routing_settings` resource — its
  existence means "enabled"; destroying it disables), and
- causes Cloudflare to provision the zone's required MX / SPF / DKIM records.

## Fold vs split

- `email_routing_settings` (enable) and the single `email_routing_catch_all` are
  strictly 1:1 with the zone and meaningless on their own, so they are folded
  into this one kind (the catch-all as `spec.catch_all`).
- `email_routing_dns` (record locking) is folded behind `lock_dns_records`.
- Per-address routing rules (`email_routing_rule`, N per zone) and destination
  addresses (`email_routing_address`, account-scoped) have independent lifecycles
  and are separate kinds.

## Typed catch-all action

The provider models a catch-all action as a generic `{type, value[]}`. The spec
exposes a typed action instead — `type` plus `forward_to` (FK →
`CloudflareEmailRoutingAddress`) and `worker` (FK → `CloudflareWorker`) — so the
graph edges are real. The modules map it back to `{type, value[]}` (forward →
addresses; worker → the single script name; drop → no values). The matcher is
always `{type: "all"}` (the only valid catch-all matcher), so it is not exposed.

## Engine parity

Both engines create `email_routing_settings`, conditionally
`email_routing_catch_all` (when `catch_all` is set) and `email_routing_dns` (when
`lock_dns_records`), with the catch-all ordered after settings. Outputs map to
`zone_id` / `enabled` / `status` / `name`.

## Provider quirk: the catch-all is not destroyable

`cloudflare_email_routing_catch_all` cannot be destroyed via the provider — once
created it persists in the API until manually changed (there is no delete
endpoint; a catch-all always exists for a zone, defaulting to drop). Both engines
inherit this. Destroying this kind removes the enablement and (when locked) the
DNS records, but the catch-all simply remains at its last-applied configuration.
This is upstream provider behavior, not a module defect.

## Live-validation note

Enabling Email Routing rewrites a real zone's MX/SPF/DKIM records, so live
apply must target a disposable zone. Destination-address verification and real
end-to-end mail delivery require receiving Cloudflare's verification email and so
are out of scope for automated validation.
