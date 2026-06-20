---
title: "Presets"
description: "Ready-to-deploy configuration presets for Auth0Role"
type: "preset-list"
componentSlug: "auth0role"
componentTitle: "Auth0Role"
provider: "auth0"
icon: "package"
order: 200
presets:
  - slug: "01-role-with-permissions"
    rank: "01"
    title: "Preset: Role with Permissions"
    excerpt: "A role that grants a focused set of API permissions (scopes). This is the most common Auth0 Role pattern — a named access tier (Editor, Viewer, Manager) backed by the scopes an application's API..."
  - slug: "02-admin-role-multi-api"
    rank: "02"
    title: "Preset: Administrator Role (Multiple APIs)"
    excerpt: "A privileged role that aggregates permissions across more than one resource server (API). Use this when an access tier needs to span several APIs — for example, an administrator who can manage both..."
  - slug: "03-role-without-permissions"
    rank: "03"
    title: "Preset: Role Without Permissions"
    excerpt: "A bare role with only a name and description. Use this when you want the role to exist as a stable, assignable identity but intend to manage its permissions elsewhere."
---

# Auth0Role Presets

Ready-to-deploy configuration presets for Auth0Role. Each preset is a complete manifest you can copy, customize, and deploy.
