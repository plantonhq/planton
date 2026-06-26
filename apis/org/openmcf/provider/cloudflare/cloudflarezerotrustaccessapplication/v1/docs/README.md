# CloudflareZeroTrustAccessApplication — design notes

## What this component is

The protected resource guarded by Cloudflare Access. It maps to the provider's
`cloudflare_zero_trust_access_application` resource and binds reusable Access
policies (by reference) to a self-hosted app, SaaS app, SSH/VNC/RDP target, app
launcher, infrastructure target, or MCP endpoint.

## Composition model

The application is the leaf of a three-resource graph:

```
CloudflareZeroTrustAccessGroup ──▶ CloudflareZeroTrustAccessPolicy ──▶ CloudflareZeroTrustAccessApplication
```

- `policies[].policy` is a `StringValueOrRef` defaulting to a
  `CloudflareZeroTrustAccessPolicy` (`status.outputs.policy_id`). The application
  never embeds inline rules — authorization lives on the standalone policy.
- `allowed_idps`, `scim_config.idp_uid`, and SaaS `source.name_by_idp[].idp_id` are
  `StringValueOrRef`, ready to become foreign keys when a first-class identity-provider
  kind is forged.

## Scope

`account_id` XOR `zone_id` (CEL-enforced). `zone_id` defaults to a
`CloudflareDnsZone` reference.

## Type-conditional surface

A message-level CEL requires `domain` for `self_hosted`/`ssh`/`vnc`/`rdp`. Other
fields (app-launcher visuals, SaaS, target criteria, CORS, ...) apply to their
respective types; the provider enforces the finer type compatibility, and the dense
field comments document which type each field applies to.

## SaaS depth

`saas_app` models the full SAML and OIDC surface (auth type, NameID format, custom
attributes with per-IdP sources, custom claims, grant types, scopes, hybrid/implicit
options, refresh-token options). The IdP-issued material (`client_id`,
`client_secret`, `public_key`, `sso_endpoint`, `idp_entity_id`) is computed by the
provider and surfaced as stack outputs.

## Outputs and JWT validation

`aud` is exported so downstream Workers / origins can validate the Cloudflare Access
JWT for requests to this application — a key composability hook.

## Notable modeling choices

- `app_launcher_visible` and `http_only_cookie_attribute` are `optional bool` so the
  provider's computed defaults apply when unset (rather than being forced to false).
- `destinations` (not the deprecated `self_hosted_domains`) is the modeled surface.
- `target_attributes` is `[{name, values}]` in the proto (proto maps can't carry
  repeated values) and rebuilt into the provider's `{name => values}` map by the
  modules.

## Engine parity

Both engines provision the full surface with no deferrals.

## Outputs

`application_id`, `aud`, `domain`, `saas_client_id`, `saas_client_secret`,
`saas_public_key`, `saas_sso_endpoint`, `saas_idp_entity_id`.
