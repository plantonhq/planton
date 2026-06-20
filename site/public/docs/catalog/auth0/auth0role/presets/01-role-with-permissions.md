---
title: "Preset: Role with Permissions"
description: "A role that grants a focused set of API permissions (scopes). This is the most common Auth0 Role pattern — a named access tier (Editor, Viewer, Manager) backed by the scopes an application's API..."
type: "preset"
rank: "01"
presetSlug: "01-role-with-permissions"
componentSlug: "auth0role"
componentTitle: "Auth0Role"
provider: "auth0"
icon: "package"
order: 1
---

# Preset: Role with Permissions

## Pattern

A role that grants a focused set of API permissions (scopes). This is the most common Auth0 Role pattern — a named access tier (Editor, Viewer, Manager) backed by the scopes an application's API exposes.

## What It Does

- Creates a role with a human-readable name and description.
- Grants the role two scopes (`read:items`, `write:items`) on a single resource server (API).
- Sets the role's permission list authoritatively — applying again with a changed list reconciles the role to exactly that set.

## When to Use

- You use Auth0 RBAC and need reusable access tiers to assign to users.
- The scopes already exist on a resource server (created via `Auth0ResourceServer` or directly in Auth0).
- You want a small, well-scoped role rather than blanket admin access.

## Customization

- Change `spec.name` and `spec.description` to match your access tier.
- Replace `resource_server_identifier` with your API's identifier (audience).
- Add or remove `permissions` entries to match the scopes your tier should grant.

## Placeholders to Replace

| Placeholder | Description |
|---|---|
| `metadata.org` | Your OpenMCF organization |
| `spec.permissions[].resource_server_identifier` | The identifier (audience) of your API |
| `spec.permissions[].name` | The scope names defined on that API |
