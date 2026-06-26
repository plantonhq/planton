# CloudflareZeroTrustAccessGroup — design notes

## What this component is

A first-class, reusable Cloudflare Zero Trust **Access group**: a named collection
of access rules (`include` / `exclude` / `require`) that other Access policies and
groups reference by ID. It maps to the provider's `cloudflare_zero_trust_access_group`
resource.

## The access-rule model

The heart of the component is the `CloudflareAccessRule` — a `oneof` of every rule
variant Cloudflare supports. Exactly one variant is set per rule (enforced by the
proto oneof + `buf.validate.oneof.required`), mirroring the provider's
"one populated key per rule object" model.

Rule evaluation:

- `include` — OR. A user matches if they satisfy any include rule.
- `exclude` — NOT. A user is removed if they satisfy any exclude rule.
- `require` — AND. A user must satisfy every require rule.

### Variant catalog

Identity: `email`, `email_domain`, `email_list`, `everyone`, `group`, `azure_ad`,
`github_organization`, `gsuite`, `okta`, `saml`, `oidc`, `auth_context`,
`login_method`, `cloudflare_account_member`. Network/device: `ip`, `ip_list`,
`geo`, `device_posture`, `certificate`, `common_name`. Service: `service_token`,
`any_valid_service_token`, `linked_app_token`. Risk/external: `user_risk_score`,
`external_evaluation`, `auth_method`.

## Composability

Cross-resource IDs are modeled as `StringValueOrRef` so they participate in the
resource graph:

- `group.id` defaults to a `CloudflareZeroTrustAccessGroup` reference
  (`status.outputs.group_id`) — enabling group-of-groups composition.
- Identity-provider IDs, list IDs, service-token IDs, device-posture integration
  UIDs, and linked-app UIDs are `StringValueOrRef` (no fixed default kind yet),
  ready to become foreign keys when those first-class kinds are forged.

## Scope

`account_id` XOR `zone_id` (CEL-enforced). Account scope is the common, reusable
case; zone scope mirrors the provider's optional zone scoping.

## Engine parity

Both engines provision the full rule surface with one exception: the
`cloudflare_account_member` rule variant is not yet exposed by the Pulumi Cloudflare
SDK (v6.17.0). The proto models it and the Terraform module provisions it; the
Pulumi module logs a warning and skips it. See the Pulumi module README. When a
newer SDK adds the field, restore it in the Pulumi module and remove the note.

## Outputs

- `group_id` — the Access group ID.
