# OpenFGA Authorization Model

Deploy OpenFGA authorization models declaratively through OpenMCF.

## Overview

An **authorization model** in OpenFGA defines the types, relations, and access rules that govern fine-grained authorization decisions. Each model specifies:

- **Types**: Object types in your system (user, document, folder, organization, etc.)
- **Relations**: Relationships between types (viewer, editor, owner, member, parent, etc.)
- **Rewrites**: How relations are computed (direct assignment, computed via other relations, union, intersection)
- **Conditions** (optional): Dynamic rules for access decisions based on context

## Important Notes

**Terraform-Only**: OpenFGA only has a Terraform provider. There is no Pulumi provider available. You must use `--provisioner tofu` when deploying.

**Immutable Models**: Authorization models in OpenFGA are immutable. Each change to `modelJson` creates a new model version with a new ID. The previous model version is retained.

## Quick Start

### 1. Create a Store First

Authorization models require an existing OpenFGA store. Deploy an `OpenFgaStore` first:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaStore
metadata:
  name: my-store
  org: my-org
  env: production
spec:
  name: my-authorization-store
```

### 2. Deploy the Model

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaAuthorizationModel
metadata:
  name: my-model-v1
  org: my-org
  env: production
spec:
  storeId: "01HXYZ..."  # ID from the OpenFgaStore deployment
  modelJson: |
    {
      "schema_version": "1.1",
      "type_definitions": [
        {"type": "user", "relations": {}},
        {
          "type": "document",
          "relations": {
            "viewer": {"this": {}},
            "editor": {"this": {}},
            "owner": {"this": {}}
          },
          "metadata": {
            "relations": {
              "viewer": {"directly_related_user_types": [{"type": "user"}]},
              "editor": {"directly_related_user_types": [{"type": "user"}]},
              "owner": {"directly_related_user_types": [{"type": "user"}]}
            }
          }
        }
      ]
    }
```

### 3. Deploy with CLI

```bash
# Create credentials file
cat > openfga-creds.yaml << EOF
apiUrl: http://localhost:8080
apiToken: your-api-token
EOF

# Deploy using Terraform/Tofu (required - no Pulumi provider)
openmcf apply --manifest authorization-model.yaml \
  --openfga-provider-config openfga-creds.yaml \
  --provisioner tofu
```

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `storeId` | string | Yes | The ID of the OpenFGA store where this model will be created |
| `modelJson` | string | Yes | The authorization model definition in JSON format |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `id` | The unique identifier of the created authorization model |

## Model JSON Format

The `modelJson` field must conform to the OpenFGA authorization model schema:

```json
{
  "schema_version": "1.1",
  "type_definitions": [
    {
      "type": "user",
      "relations": {}
    },
    {
      "type": "document",
      "relations": {
        "viewer": {"this": {}},
        "editor": {"this": {}},
        "owner": {"this": {}}
      },
      "metadata": {
        "relations": {
          "viewer": {"directly_related_user_types": [{"type": "user"}]},
          "editor": {"directly_related_user_types": [{"type": "user"}]},
          "owner": {"directly_related_user_types": [{"type": "user"}]}
        }
      }
    }
  ],
  "conditions": {}
}
```

### Converting from DSL

If you have a model in OpenFGA DSL format, you can use the `openfga_authorization_model_document` Terraform data source to convert it to JSON. Alternatively, use the OpenFGA CLI:

```bash
# Validate and convert DSL to JSON
fga model transform --file model.fga --output-format json
```

## References

- [Terraform Provider Documentation](https://registry.terraform.io/providers/openfga/openfga/latest/docs/resources/authorization_model)
- [OpenFGA Concepts: Authorization Models](https://openfga.dev/docs/concepts#what-is-an-authorization-model)
- [OpenFGA Modeling Guide](https://openfga.dev/docs/modeling)
- [OpenFGA Schema Version 1.1](https://openfga.dev/docs/modeling/migrating/migrating-schema-1-1)
