# OpenFGA Relationship Tuple Examples

## Basic Document Access

Grant a user viewer access to a specific document using structured fields:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: anne-views-budget
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz  # References an OpenFgaStore by name
  user:
    type: user
    id:
      value: anne
  relation: viewer
  object:
    type: document
    id:
      value: budget-2024
```

## Role-Based Access

Grant different roles to different users:

```yaml
# Owner has full control
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: bob-owns-project
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  user:
    type: user
    id:
      value: bob
  relation: owner
  object:
    type: project
    id:
      value: acme-corp
---
# Editor can modify
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: carol-edits-project
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  user:
    type: user
    id:
      value: carol
  relation: editor
  object:
    type: project
    id:
      value: acme-corp
---
# Viewer can only read
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: dave-views-project
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  user:
    type: user
    id:
      value: dave
  relation: viewer
  object:
    type: project
    id:
      value: acme-corp
```

## Group Membership

Add a user to a group:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: anne-in-engineering
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  user:
    type: user
    id:
      value: anne
  relation: member
  object:
    type: group
    id:
      value: engineering
```

## Userset Access

Grant access to all members of a group (userset). The `relation` field in the user
creates the userset format `group:engineering#member`:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: engineering-views-docs
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  user:
    type: group
    id:
      value: engineering
    relation: member  # Creates "group:engineering#member"
  relation: viewer
  object:
    type: folder
    id:
      value: engineering-docs
```

## Hierarchical Relationships

Create folder-document hierarchy (document in folder):

```yaml
# Document is in the folder
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: budget-in-reports
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  user:
    type: folder
    id:
      value: reports
  relation: parent
  object:
    type: document
    id:
      value: budget-2024
---
# User has access to folder (inherited by documents via model)
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: anne-views-reports
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  user:
    type: user
    id:
      value: anne
  relation: viewer
  object:
    type: folder
    id:
      value: reports
```

## Public Access (Wildcard)

Make a resource publicly accessible using the wildcard `*` for user ID:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: public-announcement
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  user:
    type: user
    id:
      value: "*"  # Wildcard - all users
  relation: viewer
  object:
    type: document
    id:
      value: company-announcement
```

## Conditional Access

Grant access with a condition (requires condition defined in authorization model):

```yaml
# Access only from allowed IP ranges
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: anne-views-sensitive
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  user:
    type: user
    id:
      value: anne
  relation: viewer
  object:
    type: document
    id:
      value: sensitive-data
  condition:
    name: in_allowed_ip_range
    contextJson: |
      {
        "allowed_ips": ["192.168.1.0/24", "10.0.0.0/8"]
      }
```

## Multi-Tenant Organization

Set up organization-level access:

```yaml
# User is admin of organization
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: alice-admin-acme
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  user:
    type: user
    id:
      value: alice
  relation: admin
  object:
    type: organization
    id:
      value: acme-corp
---
# User is member of organization
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: bob-member-acme
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  user:
    type: user
    id:
      value: bob
  relation: member
  object:
    type: organization
    id:
      value: acme-corp
---
# Project belongs to organization
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: project-in-acme
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  user:
    type: organization
    id:
      value: acme-corp
  relation: organization
  object:
    type: project
    id:
      value: internal-tools
```

## Specifying Authorization Model

Pin to a specific authorization model version using a reference:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: anne-views-budget-v2
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  authorizationModelId:
    valueFrom:
      name: document-authz-v2  # References an OpenFgaAuthorizationModel
  user:
    type: user
    id:
      value: anne
  relation: viewer
  object:
    type: document
    id:
      value: budget-2024
```

Or with a direct model ID:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: anne-views-budget-v2
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  authorizationModelId:
    value: "01HABC..."  # Direct model ID
  user:
    type: user
    id:
      value: anne
  relation: viewer
  object:
    type: document
    id:
      value: budget-2024
```

## Complete Workflow Example

Deploy a store, model, and tuples together:

```yaml
# 1. Create the store
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaStore
metadata:
  name: production-authz
  org: my-org
  env: production
spec:
  name: production-authorization-store
---
# 2. Create the authorization model (references store)
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
---
# 3. Create relationship tuples (references store and model)
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: anne-views-budget
  org: my-org
  env: production
spec:
  storeId:
    valueFrom:
      name: production-authz
  authorizationModelId:
    valueFrom:
      name: document-authz-v1
  user:
    type: user
    id:
      value: anne
  relation: viewer
  object:
    type: document
    id:
      value: budget-2024
```

## Deployment

All examples require Terraform/Tofu as the provisioner:

```bash
# Create OpenFGA credentials
cat > openfga-creds.yaml << EOF
apiUrl: https://api.fga.example.com
apiToken: your-api-token
EOF

# Deploy the complete workflow
openmcf apply --manifest workflow.yaml \
  --openfga-provider-config openfga-creds.yaml \
  --provisioner tofu
```

## Verification

After deploying tuples, verify access using the OpenFGA CLI:

```bash
# Check if user:anne can view document:budget-2024
fga query check user:anne viewer document:budget-2024 \
  --store-id $(openmcf get openfgastore production-authz -o json | jq -r '.status.outputs.id') \
  --api-url http://localhost:8080
```
