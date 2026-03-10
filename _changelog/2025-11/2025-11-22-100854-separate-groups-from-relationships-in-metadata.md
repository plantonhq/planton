# Separate Groups from Relationships in Cloud Resource Metadata

**Date**: November 22, 2025
**Type**: Breaking Change
**Components**: API Definitions, Protobuf Schemas, Resource Management

## Summary

Refactored cloud resource grouping by adding a dedicated `group` field to `CloudResourceMetadata` and removing the `group` field from `CloudResourceRelationship`. This separates visual resource grouping (an intrinsic resource property) from relationships (connections between resources), fixing a fundamental design flaw where resources were declaring groups for other resources they referenced.

## Problem Statement

Previously, resource grouping was embedded in the `relationships` array via a `group` field in each `CloudResourceRelationship` entry. This created a logical inconsistency: when resource A declared a relationship to resource B, it specified what group resource B belonged to, rather than resource B declaring its own group.

### Example of the Problem

```yaml
# service-hub-service.yaml (Resource A)
metadata:
  relationships:
    - kind: CloudflareR2Bucket
      name: planton-pipeline-logs
      type: uses
      group: app/dependencies/data/storage-buckets  # WRONG: declaring bucket's group
```

The service was declaring the bucket's group membership, but the bucket itself should declare which group it belongs to. This violated the principle of resources owning their own metadata.

### Pain Points

- **Ownership confusion**: Resources couldn't declare their own groups; instead, their consumers declared it for them
- **Duplicate declarations**: The same group was declared in multiple places wherever a resource was referenced
- **Inconsistent groups**: Different consumers could declare different groups for the same resource
- **Semantic mismatch**: Relationships describe connections between resources, not intrinsic resource properties

## Solution

Separated groups from relationships by making group membership a first-class metadata field that each resource declares for itself.

### Key Changes

1. **Added `group` field to `CloudResourceMetadata`** (field 10)
   - Simple string field for hierarchical group paths
   - Optional - resources can exist without a group
   - Examples: "app/services", "infrastructure/networking", "app/dependencies/data/storage-buckets"

2. **Removed `group` field from `CloudResourceRelationship`**
   - Relationships now only describe connections between resources
   - Cleaner semantic: type (depends_on, runs_on, uses, managed_by) describes the relationship

### Before and After

**Before** (relationships contained groups):

```protobuf
message CloudResourceRelationship {
  CloudResourceKind kind = 1;
  string env = 2;
  string name = 3;
  RelationshipType type = 4;
  string group = 5;  // ← Removed
}

message CloudResourceMetadata {
  string name = 1;
  // ... other fields ...
  repeated CloudResourceRelationship relationships = 9;
  // No group field
}
```

**After** (groups in metadata):

```protobuf
message CloudResourceMetadata {
  string name = 1;
  // ... other fields ...
  repeated CloudResourceRelationship relationships = 9;
  string group = 10;  // ← Added
}

message CloudResourceRelationship {
  CloudResourceKind kind = 1;
  string env = 2;
  string name = 3;
  RelationshipType type = 4;
  // group field removed
}
```

## Implementation Details

### Proto Changes

**File**: `apis/org/openmcf/shared/metadata.proto`

Added group field as the 10th field in `CloudResourceMetadata`:

```protobuf
message CloudResourceMetadata {
  string name = 1;
  string slug = 2;
  string id = 3;
  string org = 4;
  string env = 5;
  map<string, string> labels = 6;
  map<string, string> annotations = 7;
  repeated string tags = 8;
  repeated org.openmcf.shared.relationship.v1.CloudResourceRelationship relationships = 9;
  string group = 10;  // NEW: Group for visual organization in DAG
}
```

**File**: `apis/org/openmcf/shared/relationship/v1/relationship.proto`

Removed field 5 (group) from `CloudResourceRelationship`:

```protobuf
message CloudResourceRelationship {
  CloudResourceKind kind = 1;
  string env = 2;
  string name = 3;
  RelationshipType type = 4;
  // REMOVED: string group = 5;
}
```

### Ancillary Changes

**Commented out KubernetesNamespace enum** in `cloud_resource_kind.proto`:
- The KubernetesNamespace implementation directory was empty (only docs)
- Caused build failures during codegen
- Temporarily commented out until implementation exists

## Benefits

### 1. Correct Ownership Model

Each resource now declares its own group membership. This is semantically correct - metadata belongs to the resource itself, not to its relationships with other resources.

```yaml
# bucket declares its own group
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: planton-pipeline-logs
  group: app/dependencies/data/storage-buckets  # ← Resource owns this

# service just declares that it uses the bucket
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDeployment
metadata:
  name: service-hub
  group: app/services  # ← Service owns its own group
  relationships:
    - kind: CloudflareR2Bucket
      name: planton-pipeline-logs
      type: uses
      # no group field here
```

### 2. Single Source of Truth

Groups are declared once per resource, not duplicated across all consumers. This eliminates inconsistencies and reduces manifest verbosity.

### 3. Cleaner Semantics

Relationships now purely describe connections between resources (type: uses, depends_on, runs_on, managed_by), not visual organization hints.

### 4. Backward Compatible DAG Construction

The DAG mapper will read from `metadata.group` instead of scanning relationships. Edge construction remains unchanged - relationships still drive dependency resolution.

## Impact

### Breaking Changes

This is a **breaking change** for any code that:
- Reads the `group` field from `CloudResourceRelationship`
- Expects grouping information to be part of relationships
- Generates proto stubs from the old schema

### Migration Required

**For OpenMCF consumers**:
1. Upgrade to OpenMCF v0.2.237+ (this release)
2. Update code that reads grouping information to use `metadata.group`
3. Update YAML manifests to declare groups in `metadata.group` instead of in each relationship

**For Planton monorepo**:
1. Upgrade openmcf dependency to v0.2.237
2. Update `CloudResourceDagMapper` to read from `metadata.group`
3. Update all infra-chart manifests to use the new structure

### Code Locations

**OpenMCF**:
- `apis/org/openmcf/shared/metadata.proto` - Added group field
- `apis/org/openmcf/shared/relationship/v1/relationship.proto` - Removed group field
- `apis/org/openmcf/shared/cloudresourcekind/cloud_resource_kind.proto` - Commented out KubernetesNamespace

**Planton (pending)**:
- `backend/libs/java/domain/infra-hub-commons/src/main/java/ai/planton/infrahubcommons/cloudresource/dag/mapper/CloudResourceDagMapper.java` - Will read from `metadata.group`
- `ops/organizations/planton/infra-hub/infra-chart/planton-gcp-environment/templates/**/*.yaml` - Manifest updates

## Design Decisions

### Why Not Keep Both?

We considered keeping groups in relationships AND metadata, but decided against it because:
- **Duplication**: Requires maintaining the same information in two places
- **Confusion**: Which takes precedence if they differ?
- **Complexity**: More code to handle both cases

### Why Simple String Instead of Structured Hierarchy?

Group paths like "app/dependencies/data/storage-buckets" are treated as opaque strings, not parsed hierarchies. This keeps the API simple while allowing naming conventions for organization. Visualization tools can parse the slashes if they want hierarchical display.

### Why Optional?

Not all resources need grouping. Resources that stand alone (like a GCP project at the root) or resources where grouping doesn't add value can omit the field entirely.

## Testing Considerations

1. ✅ Proto build and stub generation completes successfully
2. ⏳ Verify DAG mapper reads from `metadata.group` (pending monorepo changes)
3. ⏳ Validate manifest files with new structure (pending monorepo changes)
4. ⏳ Test web console DAG visualization with new grouping (pending monorepo changes)
5. ⏳ Ensure backward compatibility for resources without groups (pending monorepo changes)

## Related Work

- **[2025-11-22] Remove Automatic Kubernetes Addon Grouping in DAG**: Recent change that decoupled automatic grouping from relationships, setting the stage for this refactoring
- **[2025-10-17] Explicit Relationships for Cloud Resources**: Introduced the relationships feature that this change now cleans up

## Next Steps

1. ✅ Release OpenMCF v0.2.237 with updated protos
2. ⏳ Update Planton monorepo to consume new version
3. ⏳ Update CloudResourceDagMapper to read `metadata.group`
4. ⏳ Migrate all infra-chart YAML manifests to new structure
5. ⏳ Test DAG construction and visualization
6. ⏳ Document the new structure in deployment component templates

---

**Status**: ✅ Released in OpenMCF
**Pending**: Planton monorepo integration
**Timeline**: Proto changes completed November 22, 2025

**Key Insight**: Metadata should belong to the resource itself, not be scattered across references from other resources. This change aligns the API with that principle, making groups a first-class resource property rather than a relationship attribute.

