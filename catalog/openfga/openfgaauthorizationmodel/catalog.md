# OpenFGA Authorization Model

Deploys an authorization model into an existing OpenFGA store -- the schema of types, relations, and conditions that determines what relationship tuples mean and how permission checks are computed. Models are immutable: every change mints a new model version with a new ID, previous versions are retained, and existing tuples remain valid across versions as long as the types and relations they reference are still defined. Write the model in the OpenFGA DSL (recommended) or raw JSON -- exactly one of the two.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Authorization Model** -- an `openfga_authorization_model` resource in the target store, containing the type definitions, relations, and conditions. Because models are immutable, each apply that changes the definition creates a new model version (new ID) rather than updating in place.
- **DSL-to-JSON conversion** -- created only when `modelDsl` is set: an `openfga_authorization_model_document` data source converts the human-readable DSL into the JSON the OpenFGA API accepts.

## Before You Deploy

### Planton Setup

- **OpenFGA Provider Connection** -- an active connection in the Connect module with the OpenFGA API URL and authentication credentials (API token or client credentials). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline authentication.

### OpenFGA Server

- **A running OpenFGA instance** -- self-hosted or cloud-hosted, reachable from the Planton Runner or provisioner environment.
- **An existing OpenFGA store** -- provide the store ID directly in `storeId` or reference an OpenFgaStore Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **OpenFGA Authorization Model**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the target store, and the model definition. Start from the **RBAC Authorization Model (DSL)** preset in the [Presets](#presets) tab for the proven viewer/editor/owner-with-groups pattern.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: openfga.planton.dev/v1alpha1
kind: OpenFgaAuthorizationModel
metadata:
  name: rbac-model
  org: acme-corp
  env: prod
spec:
  storeId:
    value: "01HXYZABCDEF"
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

```shell
planton apply -f openfga-authz-model.yaml
```

This creates an authorization model with a `user` type and a `document` type carrying viewer, editor, and owner relations, and surfaces the new model version's ID in `status.outputs`. OpenFGA ships only a Terraform provider, so this component provisions with Terraform/OpenTofu. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the model to a store deployed in the same InfraPipeline:

```yaml
spec:
  storeId:
    valueFrom:
      kind: OpenFgaStore
      name: prod-authz
      fieldPath: status.outputs.id
```

The InfraPipeline resolves the dependency graph, deploys the store first, then creates the authorization model in it.

## Key Configuration

These are the most important decisions when configuring an OpenFGA authorization model. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Model format** -- Exactly one of `modelDsl` and `modelJson` must be set. Prefer the DSL: it is the format OpenFGA's own documentation and tooling use, and the IaC module converts it to JSON automatically at deploy time. Reach for `modelJson` only when migrating existing JSON models.

**Model immutability** -- Every change to the definition creates a new model version with a new ID; the previous version is retained on the server. Existing relationship tuples remain valid across versions as long as the types and relations they reference are still defined -- so additive changes (new types, new relations) are safe, while removing a type or relation that live tuples reference changes what those tuples mean at check time.

**Store binding** -- `storeId` is immutable; changing it replaces the model resource. Provide a direct value or a ValueFromRef to an OpenFgaStore -- the reference resolves from the store's `status.outputs.id`.

**Type design** -- Define types that map to your application's entities and give each the relations that control access. Usersets like `group#member` grant a relation to every member of a group; `viewer from parent` cascades permissions down a folder hierarchy. Both patterns are implemented by the shipped presets, which are the fastest way to start from a known-good model.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **OpenFgaStore** | `storeId` | `status.outputs.id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | Version-specific model identifier (a new one per model change) | `authorizationModelId` on OpenFGA Relationship Tuple resources, pinning tuples to a known model version |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**RBAC with groups** -- User, group, and document types where documents grant viewer/editor to `group#member` usersets: assign permissions to groups, and membership tuples do the rest. The most common starting model for applications that need role-based access control. Start from the **RBAC Authorization Model (DSL)** preset.

**Hierarchical document access** -- Folder and document types with `parent` relations, where `viewer from parent` and `editor from parent` cascade access down the tree (Google Drive-style). Ownership deliberately does not inherit -- `owner` is always a direct grant. Start from the **Hierarchical Document Access Model (DSL)** preset.

## Works With

- [**OpenFGA Store**](/infra-catalog/openfga-store) -- the store the model is created in, wired through `storeId`
- [**OpenFGA Relationship Tuple**](/infra-catalog/openfga-relationship-tuple) -- the authorization data evaluated against this model; tuples can pin to a specific model version via its `id` output
