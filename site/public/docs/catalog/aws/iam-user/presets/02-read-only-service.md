---
title: "Read-Only Service User"
description: "This preset creates an IAM user with broad read-only access and no access keys. This is suitable for monitoring integrations, audit tools, or identity-only users where credentials are managed..."
type: "preset"
rank: "02"
presetSlug: "02-read-only-service"
componentSlug: "iam-user"
componentTitle: "IAM User"
provider: "aws"
icon: "package"
order: 2
---

# Read-Only Service User

This preset creates an IAM user with broad read-only access and no access keys. This is suitable for monitoring integrations, audit tools, or identity-only users where credentials are managed externally (e.g., federated access or console-only users).

## When to Use

- Third-party monitoring tools (Datadog, New Relic) that need read-only AWS access via federated credentials
- Audit or compliance users that should only observe resources without making changes
- Identity-only users where access keys are not needed

## Key Configuration Choices

- **Read-only access** (`ReadOnlyAccess`) -- AWS-managed policy granting read permissions across all services; no create, update, or delete
- **Access keys disabled** (`disableAccessKeys: true`) -- No programmatic credentials are created; access is through other means (console, federation)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<service-user-name>` | IAM user name (e.g., `datadog-readonly`); must match `[a-zA-Z0-9+=,.@_-]{1,64}` | Your team's IAM naming convention |

## Related Presets

- **01-ci-cd-pipeline** -- Use instead for CI/CD systems that need write access and programmatic credentials
