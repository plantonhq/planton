---
title: "Zero Trust Access Application"
description: "Zero Trust Access Application deployment documentation"
icon: "package"
order: 100
componentName: "cloudflarezerotrustaccessapplication"
---

# Cloudflare Zero Trust Access Application

Put Cloudflare Access in front of any resource — a web app, SaaS app, SSH/RDP
target, app launcher, or MCP endpoint.

A Cloudflare Zero Trust Access application is the protected resource Access guards.
It binds reusable Access policies (by reference) to the resource and configures how
users reach and authenticate to it — across self-hosted apps, federated SaaS
(SAML/OIDC), infrastructure targets, and the 2026 agent-world MCP types.

## Highlights

- **Full type surface** — self-hosted, SaaS, SSH/VNC/RDP, app launcher, WARP,
  bookmark, dash-SSO, infrastructure, MCP, and more.
- **Composable authorization** — references `CloudflareZeroTrustAccessPolicy`
  resources, which in turn reference `CloudflareZeroTrustAccessGroup`s.
- **Deep SaaS federation** — complete SAML and OIDC configuration with custom
  attributes/claims and exported signing/SSO material.
- **SCIM provisioning, MFA, CORS, destinations, and target criteria** — the full v5
  application surface.
- **JWT-ready** — exports the `aud` tag so Workers and origins can validate Access
  tokens.

## Typical use

Compose with `CloudflareZeroTrustAccessGroup` (reusable rules) and
`CloudflareZeroTrustAccessPolicy` (decisions) to model your access control as a
dependency-aware graph.
