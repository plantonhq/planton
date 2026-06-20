---
title: "Preset: Post-Login Custom Claims"
description: "Enrich ID and access tokens with custom claims after successful authentication. This is the most common Auth0 Action pattern — nearly every production Auth0 tenant needs custom claims for role-based..."
type: "preset"
rank: "01"
presetSlug: "01-post-login-custom-claims"
componentSlug: "auth0action"
componentTitle: "Auth0Action"
provider: "auth0"
icon: "package"
order: 1
---

# Preset: Post-Login Custom Claims

## Pattern

Enrich ID and access tokens with custom claims after successful authentication. This is the most common Auth0 Action pattern — nearly every production Auth0 tenant needs custom claims for role-based access control.

## What It Does

- Adds user roles from Auth0's authorization context to both ID and access tokens.
- Adds the user's email to access tokens so downstream APIs can identify the user without an extra lookup.
- Adds organization context (ID and name) when Auth0 Organizations are enabled.

## When to Use

- Your application or API needs to know the user's roles, email, or organization from the token itself.
- You want to avoid extra API calls from your backend to Auth0 to fetch user details.
- You use Auth0 RBAC and need roles propagated to your SPA or API.

## Customization

- Change the `namespace` to match your application's domain (Auth0 requires namespaced custom claims).
- Add or remove claims based on what your downstream services need.
- Add conditional logic (e.g., only enrich for certain connections or client IDs).
