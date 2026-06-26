# CloudflareZeroTrustAccessApplication

Provision a Cloudflare Zero Trust **Access application** — the protected resource
that Cloudflare Access guards. An application can be a self-hosted web app, a
federated SaaS app (SAML/OIDC), an SSH/VNC/RDP target, an app launcher, a WARP /
bookmark / dash-SSO entry, an infrastructure target, or an MCP endpoint. It binds
one or more standalone Access **policies** (by reference) to the resource and
configures how users reach and authenticate to it.

## Composable by design

In Cloudflare's v5 model, an application references reusable policies; each policy
references reusable groups. This component mirrors that: `policies[]` are foreign-key
references to `CloudflareZeroTrustAccessPolicy` resources, so the same policy can
guard many applications and authorization logic lives in one place.

```
CloudflareZeroTrustAccessGroup ──▶ CloudflareZeroTrustAccessPolicy ──▶ CloudflareZeroTrustAccessApplication
        (reusable rules)                  (decision + rules)                  (the protected resource)
```

## Scope

Account-scoped or zone-scoped (set exactly one of `accountId` or `zoneId`).
Account scope is the common case and can reuse account-level policies and groups.

## Requirements

- **API token**: requires **Account → Access: Apps and Policies → Edit** (and the
  matching zone permission for zone-scoped applications).

## Quick start

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: internal-dashboard
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: internal-dashboard
  type: self_hosted
  domain: dashboard.example.com
  policies:
    - policy:
        valueFrom:
          kind: CloudflareZeroTrustAccessPolicy
          name: allow-staff
          fieldPath: status.outputs.policy_id
      precedence: 1
```

## Application types

`self_hosted`, `saas`, `ssh`, `vnc`, `app_launcher`, `warp`, `biso`, `bookmark`,
`dash_sso`, `infrastructure`, `rdp`, `mcp`, `mcp_portal`, `proxy_endpoint`.
`domain` is required for `self_hosted`/`ssh`/`vnc`/`rdp`.

## Capability surface

- **Policies** — referenced by ID with precedence.
- **Destinations** — public URIs and private network targets (the modern
  replacement for self-hosted domain lists).
- **Access UX** — app-launcher visibility, logos/colors, landing page, footer links.
- **Self-hosted controls** — WARP auth, iframe, CORS, cookie attributes,
  interstitial, custom deny pages.
- **SaaS** — full SAML and OIDC federation (`saasApp`), with custom attributes /
  claims and signing/SSO outputs.
- **SCIM** — provisioning (`scimConfig`) with authentication and mapping rules.
- **Infrastructure** — `targetCriteria` for SSH/RDP targets.
- **MFA & MCP** — application-level MFA and OAuth authorization-server settings.

## Outputs

| Output | Description |
|---|---|
| `application_id` | The Access application ID |
| `aud` | The audience tag (validate Access JWTs with this) |
| `domain` | The protected domain |
| `saas_client_id` / `saas_client_secret` | OIDC client credentials (SaaS) |
| `saas_public_key` / `saas_sso_endpoint` / `saas_idp_entity_id` | SAML SSO material |

## Related components

- `CloudflareZeroTrustAccessPolicy` — the decisions attached here.
- `CloudflareZeroTrustAccessGroup` — reusable rule bundles referenced by policies.
