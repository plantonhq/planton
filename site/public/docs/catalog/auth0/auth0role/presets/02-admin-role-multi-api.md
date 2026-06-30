---
title: "Preset: Administrator Role (Multiple APIs)"
description: "A privileged role that aggregates permissions across more than one resource server (API). Use this when an access tier needs to span several APIs — for example, an administrator who can manage both..."
type: "preset"
rank: "02"
presetSlug: "02-admin-role-multi-api"
componentSlug: "auth0role"
componentTitle: "Auth0Role"
provider: "auth0"
icon: "package"
order: 2
---

# Preset: Administrator Role (Multiple APIs)

## Pattern

A privileged role that aggregates permissions across more than one resource server (API). Use this when an access tier needs to span several APIs — for example, an administrator who can manage both orders and billing.

## What It Does

- Creates an `Administrator` role.
- Grants a broad set of scopes drawn from two different resource servers (`/orders` and `/billing`), demonstrating that a single role's permissions can reference multiple APIs.
- Manages the full permission set authoritatively.

## When to Use

- You need a cross-API administrative or power-user tier.
- Multiple resource servers expose scopes that a single role should aggregate.
- You want admin access expressed as explicit, auditable scopes rather than an opaque "superuser" flag.

## Customization

- Adjust the scope list to grant least privilege — include only what the tier genuinely needs.
- Point each `resource_server_identifier` at your real API identifiers.
- Split into multiple narrower roles if different teams should own different APIs.

## Placeholders to Replace

| Placeholder | Description |
|---|---|
| `metadata.org` | Your Planton organization |
| `spec.permissions[].resource_server_identifier` | The identifiers (audiences) of your APIs |
| `spec.permissions[].name` | The scope names defined on those APIs |
