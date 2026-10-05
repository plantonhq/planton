# Production Environment Project

This preset creates a production-labeled project as the organizational home for one environment's resources. Membership is left unmanaged, so resources join by their own project selections (or from the console) without the manifest fighting them.

## When to Use

- One project per environment (production/staging/development) as the account's organizing convention
- Giving DigitalOcean's console a per-environment view of resources and billing
- As the anchor other manifests reference through the `project_id` output

## Key Configuration Choices

- **`environment: production`** -- lowercase canonical; DigitalOcean displays it capitalized.
- **`purpose: Web Application`** -- one of DigitalOcean's standard purposes; any free text also round-trips cleanly.
- **No `resources` list** -- membership stays unmanaged; declare the list only when this manifest should own it.

## What You Get

A project visible in the DigitalOcean control panel with its `project_id`, `owner_uuid`, and `owner_id` exported as outputs.
