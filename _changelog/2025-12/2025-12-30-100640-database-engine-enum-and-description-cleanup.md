# Database Engine Enum Support and Description Cleanup

**Date**: December 30, 2025
**Type**: Enhancement
**Provider**: Multi-Provider
**Chart(s)**: gcp/cloud-run-environment, aws/ecs-environment, aws/eks-environment, azure/aks-environment, gcp/gke-environment, digital-ocean/app-platform-environment

## Summary

Added `string_enum` parameter type for database engine selection in the GCP Cloud Run Environment chart, enabling users to choose between PostgreSQL and MySQL via a dropdown. Also cleaned up boolean parameter descriptions across all charts by removing the redundant "true →" prefix pattern.

## Problem Statement / Motivation

Two issues addressed in this change:

### 1. Database Engine Selection

The Cloud Run Environment chart was hardcoded to PostgreSQL only. Users who needed MySQL had no option to switch engines, limiting the chart's flexibility for diverse database requirements.

### 2. Description Redundancy

Boolean parameter descriptions across all charts used a "true → action" pattern (e.g., `"true → create storage bucket"`). This was redundant since the parameter type already indicates it's a boolean, and the arrow notation added visual noise without value.

### Pain Points

- No database engine choice in Cloud Run Environment chart
- PostgreSQL-specific parameter names (`postgres_*`) even though Cloud SQL supports both engines
- Inconsistent, verbose description patterns across values.yaml files
- Poor readability with "true →" prefix cluttering descriptions

## Solution / What's New

### Database Engine as String Enum

Added `database_engine` as a `string_enum` parameter with `["POSTGRESQL", "MYSQL"]` options:

```yaml
- name: database_engine
  description: Database engine type (POSTGRESQL or MYSQL)
  type: string_enum
  enum_values: ["POSTGRESQL", "MYSQL"]
  value: POSTGRESQL
```

This renders as a dropdown in the web console, matching the proto definition from `GcpCloudSqlDatabaseEngine`.

### Renamed Parameters

Renamed all PostgreSQL-specific parameters to generic database parameters:

| Before | After |
|--------|-------|
| `postgresEnabled` | `databaseEnabled` |
| `postgres_instance_name` | `database_instance_name` |
| `postgres_tier` | `database_tier` |
| `postgres_storage_gb` | `database_storage_gb` |
| `postgres_version` | `database_version` |
| `postgres_root_password` | `database_root_password` |
| `postgres_authorized_networks` | `database_authorized_networks` |

### Description Cleanup

Transformed descriptions from:
```yaml
description: "true → create storage bucket"
```

To:
```yaml
description: Create storage bucket
```

## Implementation Details

### Template Update

Renamed `postgres.yaml` to `database.yaml` and updated to use dynamic engine:

```yaml
spec:
  databaseEngine: {{ values.database_engine }}
  databaseVersion: "{{ values.database_version }}"
```

### Files Updated

| File | Changes |
|------|---------|
| `gcp/cloud-run-environment/values.yaml` | Added `database_engine` enum, renamed params, cleaned 7 descriptions |
| `gcp/cloud-run-environment/templates/database.yaml` | Dynamic engine, renamed from postgres.yaml |
| `gcp/cloud-run-environment/templates/backend-service.yaml` | Updated relationship references |
| `gcp/cloud-run-environment/README.md` | Updated documentation for database engine |
| `gcp/cloud-run-environment/Chart.yaml` | Updated description |
| `aws/ecs-environment/values.yaml` | Cleaned 1 description |
| `aws/eks-environment/values.yaml` | Cleaned 3 descriptions |
| `azure/aks-environment/values.yaml` | Cleaned 5 descriptions |
| `gcp/gke-environment/values.yaml` | Cleaned 1 description |
| `digital-ocean/app-platform-environment/values.yaml` | Cleaned 1 description |

## Benefits

- **Database flexibility**: Users can now choose MySQL or PostgreSQL for Cloud Run environments
- **Dropdown UX**: Web console renders enum as a select dropdown instead of free-text input
- **Cleaner descriptions**: Boolean parameters have concise, action-oriented descriptions
- **Consistent naming**: Generic `database_*` naming supports both engines
- **Validation**: Backend validates enum values against allowed options before chart rendering

## Impact

- **Chart users**: Can now select database engine when deploying Cloud Run environments
- **All charts**: Cleaner, more readable values.yaml files
- **Web console**: Improved UX with dropdown for database engine selection

## Usage Example

```yaml
params:
  - name: databaseEnabled
    description: Create Cloud SQL database instance
    type: bool
    value: true

  - name: database_engine
    description: Database engine type (POSTGRESQL or MYSQL)
    type: string_enum
    enum_values: ["POSTGRESQL", "MYSQL"]
    value: MYSQL

  - name: database_version
    description: Database version (e.g., POSTGRES_15 or MYSQL_8_0)
    value: MYSQL_8_0
```

## Related Work

- [2025-12-30] Planton: InfraChart string_enum parameter type support
- [2025-12-29] StringValueOrRef migration for Cloud Run Environment

---

**Status**: ✅ Production Ready
**Timeline**: ~30 minutes

