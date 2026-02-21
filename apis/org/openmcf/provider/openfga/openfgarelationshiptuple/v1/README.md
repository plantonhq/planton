# OpenFGA Relationship Tuple

Deploy OpenFGA relationship tuples declaratively through OpenMCF.

## Overview

A **relationship tuple** in OpenFGA is the fundamental unit of authorization data. It represents a relationship between a user (or userset) and an object through a specific relation. When combined with an authorization model, tuples determine whether access is granted.

Each tuple consists of:

- **User**: Who is being granted access (e.g., `user:anne`, `group:engineering#member`)
- **Relation**: The type of access (e.g., `viewer`, `editor`, `owner`)
- **Object**: What is being accessed (e.g., `document:budget-2024`, `folder:reports`)
- **Condition** (optional): Dynamic rules evaluated at check time

## Important Notes

**Terraform-Only**: OpenFGA only has a Terraform provider. There is no Pulumi provider available. You must use `--provisioner tofu` when deploying.

**Immutable Tuples**: Relationship tuples are immutable. Changing any field (user, relation, object, or condition) results in the old tuple being deleted and a new one being created.

## Quick Start

### 1. Deploy Prerequisites First

Relationship tuples require an existing store and authorization model:

```yaml
# 1. Create a store
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaStore
metadata:
  name: my-store
spec:
  name: my-authorization-store

# 2. Create an authorization model
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaAuthorizationModel
metadata:
  name: my-model
spec:
  storeId: "01HXYZ..."  # From store deployment
  modelJson: |
    {
      "schema_version": "1.1",
      "type_definitions": [
        {"type": "user", "relations": {}},
        {"type": "document", "relations": {"viewer": {"this": {}}}}
      ]
    }
```

### 2. Create Relationship Tuple

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: anne-views-budget
  org: my-org
  env: production
spec:
  storeId: "01HXYZ..."  # From store deployment
  user: "user:anne"
  relation: "viewer"
  object: "document:budget-2024"
```

### 3. Deploy with CLI

```bash
# Create credentials file
cat > openfga-creds.yaml << EOF
apiUrl: http://localhost:8080
apiToken: your-api-token
EOF

# Deploy using Terraform/Tofu (required - no Pulumi provider)
openmcf apply --manifest relationship-tuple.yaml \
  --openfga-provider-config openfga-creds.yaml \
  --provisioner tofu
```

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `storeId` | string | Yes | The ID of the OpenFGA store |
| `authorizationModelId` | string | No | The ID of the authorization model (uses latest if not specified) |
| `user` | string | Yes | The subject being granted access |
| `relation` | string | Yes | The relationship type |
| `object` | string | Yes | The resource being accessed |
| `condition` | object | No | Optional condition for dynamic access |

### Condition Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Name of the condition (must be defined in the model) |
| `contextJson` | string | No | Partial context in JSON format |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `user` | The user of the relationship tuple |
| `relation` | The relation of the relationship tuple |
| `object` | The object of the relationship tuple |

## User Format

The user field supports several formats:

| Format | Description | Example |
|--------|-------------|---------|
| `type:id` | Specific user | `user:anne`, `user:1234` |
| `type:id#relation` | Userset (all users with a relation) | `group:engineering#member` |
| `type:*` | Wildcard (all users of type) | `user:*` |

## Examples

### Basic User Access

Grant a user viewer access to a document:

```yaml
spec:
  storeId: "01HXYZ..."
  user: "user:anne"
  relation: "viewer"
  object: "document:budget-2024"
```

### Group Access

Grant all members of a group access:

```yaml
spec:
  storeId: "01HXYZ..."
  user: "group:engineering#member"
  relation: "editor"
  object: "project:acme-corp"
```

### Public Access

Grant all users access (wildcard):

```yaml
spec:
  storeId: "01HXYZ..."
  user: "user:*"
  relation: "viewer"
  object: "document:public-announcement"
```

### Conditional Access

Grant access with a condition:

```yaml
spec:
  storeId: "01HXYZ..."
  user: "user:anne"
  relation: "viewer"
  object: "document:sensitive-report"
  condition:
    name: "in_allowed_ip_range"
    contextJson: |
      {"allowed_ips": ["192.168.1.0/24", "10.0.0.0/8"]}
```

## References

- [Terraform Provider Documentation](https://registry.terraform.io/providers/openfga/openfga/latest/docs/resources/relationship_tuple)
- [OpenFGA Concepts: Relationship Tuples](https://openfga.dev/docs/concepts#what-is-a-relationship-tuple)
- [OpenFGA Conditions](https://openfga.dev/docs/modeling/conditions)
- [OpenFGA Modeling Guide](https://openfga.dev/docs/modeling)
