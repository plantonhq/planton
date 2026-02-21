---
title: "Read-Only Auditor Policy"
description: "This preset creates a tenancy-level IAM policy granting a group read-only visibility across all compartments. The `inspect` verb allows listing and viewing resource metadata without accessing data..."
type: "preset"
rank: "03"
presetSlug: "03-read-only-auditor"
componentSlug: "identity-policy"
componentTitle: "Identity Policy"
provider: "oci"
icon: "package"
order: 3
---

# Read-Only Auditor Policy

This preset creates a tenancy-level IAM policy granting a group read-only visibility across all compartments. The `inspect` verb allows listing and viewing resource metadata without accessing data contents or making changes. A supplementary `read audit-events` statement enables access to the OCI Audit service logs. This is the standard pattern for compliance teams, security reviewers, and finance teams that need cross-tenancy visibility.

## When to Use

- Security teams that need to review resource configurations and access patterns across the tenancy
- Compliance auditors who need to verify that infrastructure meets regulatory requirements
- Finance or cost management teams that need to inventory resources across compartments
- Operations teams that need read-only dashboards and monitoring across all environments
- Any group that needs broad visibility without the ability to create, modify, or delete resources

## Key Configuration Choices

- **`inspect all-resources`** -- The `inspect` verb is the lowest permission level in OCI. It allows listing resources and viewing their metadata (names, OCIDs, compartments, tags, lifecycle state) but does not allow reading data contents. For example, an auditor can see that an Object Storage bucket exists and view its configuration, but cannot read the objects inside it. This is intentional -- auditors need visibility into what exists, not access to the data itself.
- **`read audit-events`** -- Supplements `inspect` with access to OCI Audit service logs, which record all API calls made in the tenancy. This is essential for compliance investigations and security incident reviews. `read` is required instead of `inspect` because audit event contents need to be viewed, not just listed.
- **Tenancy-level scope** (`in tenancy`) -- Auditor policies are attached to the tenancy root so they apply across all compartments. A compartment-scoped auditor policy would leave blind spots. The `compartmentId` placeholder uses `<tenancy-ocid>` (the tenancy OCID) to make this explicit.
- **No `read` on `all-resources`** -- The preset intentionally uses `inspect` (not `read`) as the broad verb. `read` would grant access to resource data contents (e.g., reading secret values, downloading objects), which exceeds what auditors typically need. `read` is granted only for `audit-events` where content access is specifically required.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<tenancy-ocid>` | OCID of the tenancy root (this is a tenancy-level policy) | OCI Console > Tenancy Details page, or `oci iam tenancy get` CLI command |
| `<auditor-group-name>` | Name of the IAM group receiving read-only access | OCI Console > Identity > Groups |

## Related Presets

- **01-compartment-admin** -- Use instead when granting a group full administrative access to a specific compartment
- **02-service-access** -- Use instead when granting OCI service access to compute instances or OKE pods via a dynamic group
