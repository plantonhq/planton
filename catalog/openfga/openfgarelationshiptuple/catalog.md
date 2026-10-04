# OpenFGA Relationship Tuple

Deploys a single relationship tuple into an existing OpenFGA store -- the fundamental unit of authorization data, stating that a user or userset holds a specific relation to an object (`user:anne` is `viewer` of `document:budget-2024`). Every field is immutable: changing any of them deletes the old tuple and writes a new one, which the IaC module handles in a single apply. The user and object are structured messages (`type` plus `id`, with an optional userset `relation`), so identifiers can be wired from other resources via ValueFromRef instead of hand-assembling `type:id` strings.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Relationship Tuple** -- an `openfga_relationship_tuple` resource that writes one (user, relation, object) row into the target store. The module assembles OpenFGA's string forms from the structured spec: `user.type` + `user.id` become `user:anne`, adding `user.relation` produces a userset like `group:engineering#member`, and `object.type` + `object.id` become `document:budget-2024`.

## Before You Deploy

### Planton Setup

- **OpenFGA Provider Connection** -- an active connection in the Connect module with the OpenFGA API URL and authentication credentials (API token or client credentials). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline authentication.

### OpenFGA Server

- **A running OpenFGA instance** -- self-hosted or cloud-hosted, reachable from the Planton Runner or provisioner environment.
- **An existing store with an authorization model** that defines the types and relations the tuple uses -- a tuple naming an undefined type or relation is rejected at write time.

## Deploy

### Console

Open the deployment store, find **OpenFGA Relationship Tuple**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the target store, and the structured user, relation, and object fields. Start from the **User-Document Access Tuple** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: openfga.planton.dev/v1alpha1
kind: OpenFgaRelationshipTuple
metadata:
  name: anne-views-budget
  org: acme-corp
  env: prod
spec:
  storeId:
    value: "01HXYZABCDEF"
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

```shell
planton apply -f openfga-tuple.yaml
```

This writes one tuple granting `user:anne` the `viewer` relation on `document:budget-2024`, validated against the store's latest authorization model. OpenFGA ships only a Terraform provider, so this kind provisions with Terraform/OpenTofu. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the tuple to a store and model deployed in the same InfraPipeline:

```yaml
spec:
  storeId:
    valueFrom:
      kind: OpenFgaStore
      name: prod-authz
      fieldPath: status.outputs.id
  authorizationModelId:
    valueFrom:
      kind: OpenFgaAuthorizationModel
      name: rbac-model
      fieldPath: status.outputs.id
```

The InfraPipeline resolves the dependency graph, deploys the store and model first, then writes the relationship tuple.

## Key Configuration

These are the most important decisions when configuring an OpenFGA relationship tuple. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**User structure and usersets** -- The `user` object carries `type`, `id`, and an optional `relation`. Without `relation`, the subject is a single entity (`user:anne`); with it, the subject becomes a userset -- `group:engineering` with relation `member` yields `group:engineering#member`, granting the tuple's relation to every current and future member of that group. An `id` of `*` is the wildcard: `user:*` grants the relation to all users of the type, the mechanism for public access.

**Relation** -- The `relation` must be defined in the authorization model for the object's type, and the user's type must be an allowed subject for it -- the model is the contract every tuple is checked against.

**Model pinning** -- `authorizationModelId` is optional. Omitted, the tuple is validated against the store's latest model at creation time; set, it pins validation to that specific model version. Pin in production so tuple writes cannot silently start validating against a newer model than the one you tested.

**Conditions** -- The optional `condition` block makes the tuple dynamic: it only participates in a check when the named condition (which must be declared in the authorization model) evaluates true. `contextJson` supplies partial context stored with the tuple -- for example an allowed IP range -- which is merged with the context supplied at check time before evaluation.

**Tuple granularity** -- Each Infra Component manages exactly one tuple. That fits long-lived structural grants -- group memberships, organization hierarchy, service-to-service access -- where declarative history matters. High-churn per-user grants created as users interact with the application are better written by the application through the OpenFGA API.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **OpenFgaStore** | `storeId` | `status.outputs.id` |
| **OpenFgaAuthorizationModel** (optional) | `authorizationModelId` | `status.outputs.id` |

### What This Kind Provides

Relationship tuples have no server-side identifier -- a tuple is identified by the combination of store, user, relation, and object. `status.outputs` echoes `user`, `relation`, and `object` back as confirmation that the tuple was written; there is nothing here for downstream Infra Components to consume via ValueFromRef.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Direct user access** -- One user, one relation, one resource: `user:anne` is `viewer` of `document:budget-2024`. The fundamental grant, right for individual assignments outside any group structure. Start from the **User-Document Access Tuple** preset.

**Group membership** -- `user:anne` is `member` of `group:engineering`. On its own this grants nothing; it activates every tuple and model rule that references `group:engineering#member`, which is what makes groups the scalable alternative to per-user grants. Start from the **Group Membership Tuple** preset.

**Public access** -- The wildcard subject `user:*` (user `id` of `*`) as `viewer` of an object makes it readable by all users of the type -- deliberate, auditable public access rather than a policy loophole.

**Conditional access** -- A tuple carrying a `condition` participates in checks only when runtime context satisfies it (IP ranges, business hours). The condition must be declared in the authorization model first; the tuple contributes its stored `contextJson` to the evaluation.

## Works With

- [**OpenFGA Store**](/infra-catalog/openfga-store) -- the store the tuple is written into, wired through `storeId`
- [**OpenFGA Authorization Model**](/infra-catalog/openfga-authorization-model) -- defines the types, relations, and conditions the tuple must conform to; pin a version via `authorizationModelId`
