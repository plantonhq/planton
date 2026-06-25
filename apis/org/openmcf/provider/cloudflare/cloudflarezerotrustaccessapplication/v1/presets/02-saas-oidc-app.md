# Preset: SaaS application (OIDC)

Federate a SaaS application into Cloudflare Access over OIDC. Cloudflare acts as the
identity provider; it issues the OAuth `client_id` / `client_secret` (exported as
stack outputs) that you paste into the SaaS provider's SSO settings.

## When to use

- A third-party SaaS app supports OIDC SSO and you want Cloudflare Access to broker
  identity.

## Key choices

- `saasApp.authType: oidc` (use `saml` for SAML apps and the SAML fields instead).
- `redirectUris`: the SaaS provider's OAuth callback URLs.
- `grantTypes` / `scopes`: match what the SaaS app expects.
- `policies`: reference the policy that decides who may sign in.

## Outputs to wire into the SaaS provider

| Output | Use |
|---|---|
| `saas_client_id` | OAuth client ID |
| `saas_client_secret` | OAuth client secret (sensitive) |

(For SAML apps, use `saas_public_key`, `saas_sso_endpoint`, and `saas_idp_entity_id`.)

## Placeholders

| Placeholder | Description |
|---|---|
| `REPLACE_WITH_ACCOUNT_ID` | 32-character Cloudflare account ID |
