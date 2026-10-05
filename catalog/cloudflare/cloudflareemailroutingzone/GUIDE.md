# CloudflareEmailRoutingZone guide

Operational judgment for running Email Routing well. The README covers what each field is; this covers how the pieces interact.

## Known upstream blocker: deploys currently fail on both engines

The Cloudflare Terraform provider shipped a defect in v5.23.0 that is still present in v5.24.0 (the latest release as of 2026-08-26): the `cloudflare_email_routing_settings` resource fails EVERY create and refresh with "Value Conversion Error ... support_subaddress" because the provider's internal model and schema disagree ([issue #7301](https://github.com/cloudflare/terraform-provider-cloudflare/issues/7301), fix pending in [PR #7302](https://github.com/cloudflare/terraform-provider-cloudflare/pull/7302)). The Pulumi provider bridges the same broken code. Nothing in your manifest causes or avoids this, and a plan/preview looks clean — the failure only appears at apply. Until a provider release carries the fix, this kind cannot deploy; the manifests you write today are valid and will work unchanged once the fix ships.

## Enabling rewrites the zone's mail records

Creating this resource is the enable: Cloudflare provisions the zone's MX/SPF/DKIM records and starts accepting inbound mail for routing. If the domain already receives mail elsewhere (Google Workspace, O365), enabling Email Routing replaces that delivery path — never enable it on a zone whose mail another system must keep handling.

## Destroy semantics differ per folded resource

- The settings resource is a toggle: destroy DISABLES routing (it does not "delete" an object).
- The catch-all's provider Delete is a genuine no-op — destroying the resource (or removing `catchAll` from the spec and re-applying) leaves the zone's last catch-all configuration in place on Cloudflare's side. To actually neutralize a catch-all without disabling routing, set it to a disabled drop (`enabled: false`, `actions: [{type: drop}]`) rather than deleting it.
- The managed-DNS resource's destroy removes the routing DNS records it manages.

## Catch-all actions are a list

One catch-all can forward to mailboxes AND hand a copy to an Email Worker — actions apply in order. `drop` stands alone by nature. Every `forwardTo` destination must be a VERIFIED `CloudflareEmailRoutingAddress`; Cloudflare rejects rules that forward to unverified addresses, and verification only happens through the emailed link.

## Subdomain routing rides the managed DNS records

`dnsName` targets the managed records at a subdomain (`mail.example.com`) instead of the apex — that is what makes Cloudflare accept and route mail addressed to `*@mail.example.com`. It requires `lockDnsRecords: true` because the subdomain choice lives on the managed-DNS resource; without explicit management there is nothing to carry it.

## Order of operations for a full mail setup

1. `CloudflareEmailRoutingAddress` for each real mailbox (owners click the verification links).
2. This kind, to enable routing (with the catch-all as the fallback policy).
3. `CloudflareEmailRoutingRule` per address you route explicitly — rules always run before the catch-all regardless of creation order.
