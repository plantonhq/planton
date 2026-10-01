# Sandbox Project Noise

## Use Case

Mute one category of finding in a sandbox project where it is expected, so it stops reaching triage.

## When to Use

- A demo or sandbox project that is public by design
- A detector that is noisy for one project's workload

## What This Creates

- The Security Command Center API on the project
- A dynamic mute rule for public-bucket findings in the project

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `filter` | `PUBLIC_BUCKET_ACL` | Mute a different category. |
| `type` | `DYNAMIC` | `STATIC` only for permanent mutes. |
