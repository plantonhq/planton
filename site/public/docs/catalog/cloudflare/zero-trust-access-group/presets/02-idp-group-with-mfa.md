---
title: "Preset: IdP group with MFA login method"
description: "An account-scoped group that matches an Okta group (federated through a configured identity provider) and additionally requires a hardware-key authentication method."
type: "preset"
rank: "02"
presetSlug: "02-idp-group-with-mfa"
componentSlug: "zero-trust-access-group"
componentTitle: "Zero Trust Access Group"
provider: "cloudflare"
icon: "package"
order: 2
---

# Preset: IdP group with MFA login method

An account-scoped group that matches an Okta group (federated through a configured
identity provider) and additionally requires a hardware-key authentication method.

## When to use

- Membership is driven by an IdP group and you want an extra authentication-method
  requirement layered on.

## Key choices

- `include.okta`: the Okta group name plus the Cloudflare identity-provider ID.
- `require.authMethod`: an AMR value such as `hwk` (hardware key) or `mfa`.

## Placeholders

| Placeholder | Description |
|---|---|
| `REPLACE_WITH_ACCOUNT_ID` | 32-character Cloudflare account ID |
| `REPLACE_WITH_IDP_ID` | The Cloudflare identity-provider ID for the Okta connection |
