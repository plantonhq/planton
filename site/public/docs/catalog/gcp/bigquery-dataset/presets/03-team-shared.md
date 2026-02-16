---
title: "Preset: Team-Shared Dataset"
description: "Use this preset when a dataset needs explicit, team-level access control -- separating data engineers (who create and modify tables) from data analysts (who only read). This is the standard pattern..."
type: "preset"
rank: "03"
presetSlug: "03-team-shared"
componentSlug: "bigquery-dataset"
componentTitle: "BigQuery Dataset"
provider: "gcp"
icon: "package"
order: 3
---

# Preset: Team-Shared Dataset

## When to Use

Use this preset when a dataset needs explicit, team-level access control --
separating data engineers (who create and modify tables) from data analysts
(who only read). This is the standard pattern for shared analytics datasets
in organizations with defined data team roles.

## What It Creates

- A BigQuery dataset with explicit access control
- Project owners retain OWNER access
- A specified group gets WRITER access (can create/modify tables)
- A specified group gets READER access (read-only)
- Google-managed encryption (default)

## Configuration

| Field | Value | Notes |
|-------|-------|-------|
| Location | US | Multi-regional; change for data residency |
| Encryption | Google-managed | Default |
| Access | Explicit | Project owners + editor group + viewer group |

## Important: Authoritative Access

The `access` field is **authoritative**. BigQuery will remove any access entries
not listed in the spec. This means:

- You **must** explicitly include `projectOwners` as OWNER if you want project
  owners to retain access
- Any access granted through the GCP console that is not in the spec will be
  removed on the next apply
- This provides a single source of truth for dataset access

## How to Use

1. Replace `<project-id>` with your GCP project ID
2. Replace `<your_dataset_id>` with a descriptive name
3. Replace `<editors-group>@example.com` with your data engineering team's Google Group
4. Replace `<viewers-group>@example.com` with your data analyst team's Google Group
5. Add `friendlyName` and `description` for discoverability

## Extending Access

To add more access entries:

```yaml
access:
  # ... existing entries ...
  - role: READER
    userByEmail: "external-consultant@partner.com"
  - role: READER
    domain: "trusted-partner.com"
  - role: roles/bigquery.dataViewer
    iamMember: "serviceAccount:pipeline@project.iam.gserviceaccount.com"
```
