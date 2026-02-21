# OpenFGA Cross-References and Enhanced Spec Fields

**Date**: January 17, 2026
**Type**: Feature Enhancement
**Components**: OpenFgaAuthorizationModel, OpenFgaRelationshipTuple, Proto Specs, Terraform Modules

## Summary

Enhanced OpenFGA deployment components with cross-reference support, DSL format for authorization models, and structured user/object fields for relationship tuples. These changes make it easier to work with OpenFGA resources by enabling name-based references and more intuitive field structures.

## Problem Statement / Motivation

The initial OpenFGA implementation required users to:

1. **Manually copy IDs**: Store IDs and model IDs had to be copied between resources
2. **Use verbose JSON**: Authorization models required JSON format, which is hard to read
3. **Construct colon-separated strings**: User and object fields required `type:id` format

These patterns were error-prone and didn't leverage OpenMCF's foreign key system.

## Solution / What's New

### 1. Cross-Reference Support

All `store_id` and `authorization_model_id` fields now support the `StringValueOrRef` pattern:

```yaml
# Before: Manual ID copy
spec:
  storeId: "01HXYZ..."

# After: Reference by name
spec:
  storeId:
    valueFrom:
      name: production-authz  # References OpenFgaStore
```

**Cross-references added:**
- `OpenFgaAuthorizationModel.store_id` → `OpenFgaStore.status.outputs.id`
- `OpenFgaRelationshipTuple.store_id` → `OpenFgaStore.status.outputs.id`
- `OpenFgaRelationshipTuple.authorization_model_id` → `OpenFgaAuthorizationModel.status.outputs.id`

### 2. DSL Format for Authorization Models

Authorization models can now be specified in DSL format (recommended) or JSON format:

```yaml
# DSL format (recommended - human-readable)
spec:
  modelDsl: |
    model
      schema 1.1
    
    type user
    
    type document
      relations
        define viewer: [user]
        define editor: [user]
        define owner: [user]

# JSON format (still supported)
spec:
  modelJson: |
    {"schema_version": "1.1", "type_definitions": [...]}
```

The Terraform module uses `openfga_authorization_model_document` data source to convert DSL to JSON.

### 3. Structured User and Object Fields

Relationship tuple `user` and `object` fields are now structured messages:

```yaml
# Before: String format
spec:
  user: "user:anne"
  object: "document:budget-2024"

# After: Structured format
spec:
  user:
    type: user
    id: anne
  object:
    type: document
    id: budget-2024

# Userset support
spec:
  user:
    type: group
    id: engineering
    relation: member  # Creates "group:engineering#member"
```

**New messages:**
- `OpenFgaRelationshipTupleUser`: `type`, `id`, optional `relation`
- `OpenFgaRelationshipTupleObject`: `type`, `id`

## Implementation Details

### Proto Changes

**OpenFgaAuthorizationModel spec.proto:**
```protobuf
message OpenFgaAuthorizationModelSpec {
  StringValueOrRef store_id = 1 [
    (buf.validate.field).required = true,
    (default_kind) = OpenFgaStore,
    (default_kind_field_path) = "status.outputs.id"
  ];
  string model_dsl = 2;  // NEW: DSL format (recommended)
  string model_json = 3; // Existing JSON format
}
```

**OpenFgaRelationshipTuple spec.proto:**
```protobuf
message OpenFgaRelationshipTupleSpec {
  StringValueOrRef store_id = 1 [...];
  StringValueOrRef authorization_model_id = 2 [...];  // Now optional StringValueOrRef
  OpenFgaRelationshipTupleUser user = 3;   // NEW: Structured user
  string relation = 4;
  OpenFgaRelationshipTupleObject object = 5;  // NEW: Structured object
  OpenFgaRelationshipTupleCondition condition = 6;
}

message OpenFgaRelationshipTupleUser {
  string type = 1;      // e.g., "user", "group"
  string id = 2;        // e.g., "anne", "engineering"
  string relation = 3;  // optional, for usersets
}

message OpenFgaRelationshipTupleObject {
  string type = 1;  // e.g., "document", "folder"
  string id = 2;    // e.g., "budget-2024", "reports"
}
```

### Terraform Module Changes

**OpenFgaAuthorizationModel:**
- Added `openfga_authorization_model_document` data source for DSL-to-JSON conversion
- Updated variables to accept `store_id` as `object({ value = string })`
- Added validation: exactly one of `model_dsl` or `model_json` required

**OpenFgaRelationshipTuple:**
- Updated variables to accept structured `user` and `object` objects
- Added locals to construct `type:id` and `type:id#relation` formats
- Updated `store_id` and `authorization_model_id` to `object({ value = string })`

## Files Changed

| Category | Files |
|----------|-------|
| Proto API | `openfgaauthorizationmodel/v1/spec.proto`, `openfgarelationshiptuple/v1/spec.proto` |
| Generated | `*.pb.go`, TypeScript types |
| Terraform | `variables.tf`, `locals.tf`, `main.tf` for both components |
| Pulumi | `locals.go` for both components (updated to handle new types) |
| Documentation | `examples.md` for both components |

Note: The Pulumi modules are pass-through placeholders (OpenFGA has no Pulumi provider), but their `locals.go` files needed updating to compile with the new proto types.

## Benefits

### For Users
- **Simpler Configuration**: Reference resources by name instead of copying IDs
- **Readable Models**: DSL format is much easier to read and write than JSON
- **Structured Fields**: User and object fields are self-documenting and validated

### For Developers
- **Better IDE Support**: Structured fields enable autocomplete and validation
- **Reduced Errors**: No more malformed `type:id` strings
- **Consistent Patterns**: Follows Auth0 cross-reference patterns

### For Operations
- **Dependency Tracking**: Platform can track resource dependencies
- **Validation**: References validated before deployment
- **Audit Trail**: Clear relationships between resources

## Breaking Change Notice

This is a **breaking change** for existing manifests.

### store_id and authorization_model_id

```yaml
# Old format
spec:
  storeId: "01HXYZ..."

# New format
spec:
  storeId:
    value: "01HXYZ..."
  # OR
  storeId:
    valueFrom:
      name: my-store
```

### user and object

```yaml
# Old format
spec:
  user: "user:anne"
  object: "document:budget-2024"

# New format
spec:
  user:
    type: user
    id: anne
  object:
    type: document
    id: budget-2024
```

### model_json

```yaml
# Old format (required)
spec:
  modelJson: |
    {...}

# New format (one of model_dsl or model_json required)
spec:
  modelDsl: |
    model
      schema 1.1
    ...
```

## Usage Example

Complete workflow with cross-references:

```yaml
# 1. Store
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaStore
metadata:
  name: production-authz
spec:
  name: production-authorization-store

---
# 2. Model (references store)
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaAuthorizationModel
metadata:
  name: document-authz-v1
spec:
  storeId:
    valueFrom:
      name: production-authz
  modelDsl: |
    model
      schema 1.1
    type user
    type document
      relations
        define viewer: [user]

---
# 3. Tuple (references store and model)
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: anne-views-budget
spec:
  storeId:
    valueFrom:
      name: production-authz
  authorizationModelId:
    valueFrom:
      name: document-authz-v1
  user:
    type: user
    id: anne
  relation: viewer
  object:
    type: document
    id: budget-2024
```

## Related Work

- OpenFgaStore: `2026-01-17-085733-openfgastore-deployment-component.md`
- OpenFgaAuthorizationModel: `2026-01-17-090928-openfgaauthorizationmodel-deployment-component.md`
- OpenFgaRelationshipTuple: `2026-01-17-095002-openfgarelationshiptuple-deployment-component.md`
- Auth0Client Cross-References: `2026-01-10-185920-auth0-client-cross-references.md`

---

**Status**: ✅ Production Ready
**Build**: Protos generated, Terraform validates
