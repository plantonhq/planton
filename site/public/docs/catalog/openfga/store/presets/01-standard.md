---
title: "Standard Authorization Store"
description: "This preset creates an OpenFGA store, the logical container for all authorization data (models and relationship tuples). Every OpenFGA deployment starts with a store. The store name is immutable --..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "store"
componentTitle: "Store"
provider: "openfga"
icon: "package"
order: 1
---

# Standard Authorization Store

This preset creates an OpenFGA store, the logical container for all authorization data (models and relationship tuples). Every OpenFGA deployment starts with a store. The store name is immutable -- changing it replaces the store and all its data.

## When to Use

- Any project that needs fine-grained authorization via OpenFGA
- First step in setting up an OpenFGA authorization system (store -> model -> tuples)
- Separate stores per environment (dev/staging/production) or per tenant

## Key Configuration Choices

- **Single required field** (`name`) -- the store is intentionally minimal; all complexity lives in the authorization model and tuples
- **Terraform/Tofu only** -- OpenFGA has no Pulumi provider; this component uses Terraform as the provisioner

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `my-authz-store` | Descriptive store name (e.g., `production-authz`, `dev-permissions`) | Your naming convention |
