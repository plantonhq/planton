---
title: "Standard Keycloak"
description: "This preset deploys Keycloak with ingress for external access. Keycloak is an open-source identity and access management solution providing SSO, user federation, OIDC, and SAML support."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "keycloak"
componentTitle: "Keycloak"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard Keycloak

This preset deploys Keycloak with ingress for external access. Keycloak is an open-source identity and access management solution providing SSO, user federation, OIDC, and SAML support.

## When to Use

- You need a self-hosted identity provider for SSO, OIDC, or SAML
- You want user federation, social login, or multi-factor authentication
- You need the Keycloak admin console accessible via a hostname

## Key Configuration Choices

- **Ingress enabled** -- exposes the Keycloak admin console and auth endpoints at the specified hostname
- **Higher memory** (`512Mi` request, `2Gi` limit) -- Keycloak is a Java application with significant JVM memory requirements

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-keycloak.example.com>` | Hostname for Keycloak (admin console and auth endpoints) | Your DNS provider |
