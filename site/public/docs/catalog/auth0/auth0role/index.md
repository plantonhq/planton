---
title: "Auth0Role"
description: "Auth0Role deployment documentation"
icon: "package"
order: 100
componentName: "auth0role"
---

---
title: Auth0Role
kind: Auth0Role
provider: auth0
api_version: auth0.planton.dev/v1
id_prefix: a0role
description: Manage Auth0 Roles — named collections of API permissions that implement role-based access control (RBAC) for assignment to users.
---

# Auth0Role

Manage Auth0 Roles — named collections of API permissions (scopes) that implement Auth0's role-based access control (RBAC). A role groups scopes defined on one or more Auth0 Resource Servers and can be assigned to users, giving them the role's permissions in their access tokens.

## Provider

Auth0

## Category

Identity & Access Management

## Use Cases

- Define standard access tiers (Administrator, Editor, Viewer) for an application
- Group API scopes into reusable roles for assignment to users
- Aggregate permissions across multiple resource servers (APIs) into one role
- Manage role-to-permission mappings as version-controlled infrastructure

## Related Components

- [Auth0ResourceServer](/docs/catalog/auth0/resource-server) — defines the APIs and scopes (permissions) that roles grant
- [Auth0Client](/docs/catalog/auth0/client) — applications whose users are assigned roles
- [Auth0Action](/docs/catalog/auth0/auth0action) — post-login actions can read a user's roles to enrich tokens
