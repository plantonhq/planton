# GCP Cloud Run Environment: StringValueOrRef API Migration

**Date**: December 29, 2025
**Type**: Refactoring
**Provider**: GCP
**Chart(s)**: gcp/cloud-run-environment

## Summary

Updated all templates in the GCP Cloud Run Environment chart to use the `StringValueOrRef` pattern for project ID fields, ensuring compatibility with the recent API migrations in planton. This change wraps literal project ID values in a `value:` field structure required by the updated proto schemas.

## Problem Statement / Motivation

The planton APIs recently migrated several GCP component fields from plain `string` types to `StringValueOrRef` types. This enables cross-resource references where values can be dynamically resolved from other resources' outputs.

### Pain Points

- **API Incompatibility**: Templates using the old flat string format (`projectId: "my-project"`) no longer match the updated proto schemas
- **Deployment Failures**: Manifests rendered from the chart would fail validation against the new API
- **Inconsistency**: Some templates (like `network.yaml`) already used the correct pattern while others didn't

## Solution / What's New

Updated 7 template files to wrap project ID values in the `StringValueOrRef` pattern:

**Before:**
```yaml
spec:
  projectId: "{{ values.gcp_project_id }}"
```

**After:**
```yaml
spec:
  projectId:
    value: "{{ values.gcp_project_id }}"
```

### Templates Updated

| Template | Resource Kind | Field Updated |
|----------|---------------|---------------|
| `frontend-service.yaml` | GcpCloudRun | `projectId` |
| `backend-service.yaml` | GcpCloudRun | `projectId` |
| `dns.yaml` | GcpDnsZone | `projectId` |
| `docker-repo.yaml` | GcpArtifactRegistryRepo | `projectId` |
| `postgres.yaml` | GcpCloudSql | `projectId` |
| `service-account.yaml` | GcpServiceAccount | `projectId` |
| `storage-bucket.yaml` | GcpGcsBucket | `gcpProjectId` |

### Templates Already Compliant

- `network.yaml` (GcpVpc, GcpSubnetwork, GcpRouterNat) - Already using correct pattern with both `value:` and `valueFrom:` syntax

## Implementation Details

### GcpCloudRun Templates

Both frontend and backend services updated:

```yaml
# frontend-service.yaml / backend-service.yaml
spec:
  projectId:
    value: "{{ values.gcp_project_id }}"
  region: "{{ values.gcp_region }}"
```

### GcpDnsZone Template

```yaml
# dns.yaml
spec:
  projectId:
    value: "{{ values.gcp_project_id }}"
```

### GcpArtifactRegistryRepo Template

```yaml
# docker-repo.yaml
spec:
  projectId:
    value: "{{ values.gcp_project_id }}"
  region: "{{ values.gcp_region }}"
  repoFormat: DOCKER
```

### GcpCloudSql Template

```yaml
# postgres.yaml
spec:
  projectId:
    value: "{{ values.gcp_project_id }}"
  region: "{{ values.gcp_region }}"
```

Note: The `network.vpcId` field in postgres.yaml already correctly uses `valueFrom:` for cross-resource references.

### GcpServiceAccount Template

```yaml
# service-account.yaml
spec:
  projectId:
    value: "{{ values.gcp_project_id }}"
```

### GcpGcsBucket Template

```yaml
# storage-bucket.yaml
spec:
  gcpProjectId:
    value: "{{ values.gcp_project_id }}"
```

## Benefits

- **API Compatibility**: All templates now produce valid manifests that pass proto validation
- **Future-Ready**: The `StringValueOrRef` pattern enables future use of `valueFrom:` for cross-resource references
- **Consistency**: All templates in the chart now use the same pattern for project ID fields
- **No User Impact**: Chart users don't need to change their `values.yaml` - the template handles the conversion

## Impact

### Chart Users

- **No action required**: The `values.yaml` schema remains unchanged
- Rendered manifests will now be valid against the latest planton APIs

### Related API Migrations

This chart update aligns with the following planton API migrations:

- `GcpCloudRun` - `project_id` field migrated to `StringValueOrRef`
- `GcpDnsZone` - `project_id` field migrated to `StringValueOrRef`
- `GcpArtifactRegistryRepo` - `project_id` field migrated to `StringValueOrRef`
- `GcpCloudSql` - `project_id` field migrated to `StringValueOrRef`
- `GcpServiceAccount` - `project_id` field migrated to `StringValueOrRef`
- `GcpGcsBucket` - `gcp_project_id` field migrated to `StringValueOrRef`

## Related Work

- Planton changelog: `2025-12-26-185740-gcpcloudrun-stringvalueorref-migration.md`
- Planton changelog: `2025-12-26-184912-gcpdnszone-valuefrom-migration.md`
- Planton changelog: `2025-12-26-184920-gcpartifactregistry-valuefrom-migration.md`
- Planton changelog: `2025-12-27-044402-gcpcloudsql-valuefrom-migration.md`
- Planton changelog: `2025-12-26-185828-gcpserviceaccount-stringvalueorref-migration.md`
- Planton changelog: `2025-12-26-184919-gcpgcsbucket-valuefrom-migration.md`

---

**Status**: ✅ Production Ready
**Files Changed**: 7 templates

