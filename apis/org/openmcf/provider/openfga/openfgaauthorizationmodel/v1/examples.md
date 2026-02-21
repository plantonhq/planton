# OpenFGA Authorization Model Examples

## Basic Document Authorization (DSL Format - Recommended)

A simple model for document access control with viewer, editor, and owner roles.
The DSL format is more human-readable than JSON:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaAuthorizationModel
metadata:
  name: document-authz-v1
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz  # References an OpenFgaStore by name
  modelDsl: |
    model
      schema 1.1

    type user

    type document
      relations
        define viewer: [user]
        define editor: [user]
        define owner: [user]
```

## Using Direct Store ID

If you have an existing store ID, you can provide it directly:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaAuthorizationModel
metadata:
  name: document-authz-v1
  org: my-org
  env: production
spec:
  storeId:
    value: "01HXYZ..."  # Direct store ID value
  modelDsl: |
    model
      schema 1.1

    type user

    type document
      relations
        define viewer: [user]
        define editor: [user]
        define owner: [user]
```

## Google Drive-like Model

A more complex model with folders, documents, and inheritance using DSL:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaAuthorizationModel
metadata:
  name: drive-authz-v1
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  modelDsl: |
    model
      schema 1.1

    type user

    type folder
      relations
        define viewer: [user] or viewer from parent
        define editor: [user]
        define owner: [user]
        define parent: [folder]

    type document
      relations
        define viewer: [user] or viewer from parent
        define editor: [user]
        define owner: [user]
        define parent: [folder]
```

## Model with Groups (Usersets)

A model that supports group-based access:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaAuthorizationModel
metadata:
  name: team-authz-v1
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  modelDsl: |
    model
      schema 1.1

    type user

    type group
      relations
        define member: [user]

    type document
      relations
        define viewer: [user, group#member]
        define editor: [user, group#member]
        define owner: [user]

    type folder
      relations
        define viewer: [user, group#member]
        define editor: [user, group#member]
```

## Multi-Tenant SaaS Model

A model for multi-tenant applications with organizations, teams, and projects:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaAuthorizationModel
metadata:
  name: saas-authz-v1
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  modelDsl: |
    model
      schema 1.1

    type user

    type organization
      relations
        define admin: [user]
        define member: [user]

    type team
      relations
        define member: [user]
        define organization: [organization]

    type project
      relations
        define viewer: [user, team#member]
        define editor: [user, team#member]
        define admin: [user]
        define team: [team]
        define organization: [organization]
```

## Model with Conditions

A model with dynamic conditions for time-based or context-based access:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaAuthorizationModel
metadata:
  name: conditional-authz-v1
  org: my-org
  env: production
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
        define viewer: [user with in_allowed_ip_range]
        define editor: [user]
        define owner: [user]

    condition in_allowed_ip_range(user_ip: ipaddress, allowed_ranges: list<string>) {
      user_ip.in_cidr(allowed_ranges)
    }
```

## Using JSON Format (Alternative)

If you prefer JSON format or are migrating from existing JSON models:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaAuthorizationModel
metadata:
  name: document-authz-v1
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  modelJson: |
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
      ]
    }
```

## Deployment

All examples require Terraform/Tofu as the provisioner:

```bash
# Create OpenFGA credentials
cat > openfga-creds.yaml << EOF
apiUrl: https://api.fga.example.com
apiToken: your-api-token
EOF

# Deploy the store first
openmcf apply --manifest store.yaml \
  --openfga-provider-config openfga-creds.yaml \
  --provisioner tofu

# Then deploy the authorization model
openmcf apply --manifest model.yaml \
  --openfga-provider-config openfga-creds.yaml \
  --provisioner tofu
```

## Versioning Strategy

Since authorization models are immutable, use metadata names that include versions:

```yaml
metadata:
  name: document-authz-v2  # Increment version for new model
```

This makes it easy to:
- Track model evolution over time
- Roll back to previous versions if needed
- A/B test different authorization models
