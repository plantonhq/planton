# OpenFGA Store

Deploys an OpenFGA store -- the top-level container for authorization models and relationship tuples, and the isolation boundary that keeps one environment's, application's, or tenant's authorization data invisible to every other. A store is always the first resource in an OpenFGA deployment: models and tuples are created inside it and every downstream operation -- creating models, writing tuples, running permission checks -- addresses it by the server-generated store ID. The spec is intentionally minimal, a single `name` field, because all authorization complexity lives in the model and tuples deployed into the store.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **OpenFGA Store** -- a single `openfga_store` resource named from `spec.name` on the connected OpenFGA server. The server generates the store ID at creation; that ID -- not the name -- is what authorization models, relationship tuples, and permission checks reference.

## Before You Deploy

### Planton Setup

- **OpenFGA Provider Connection** -- an active connection in the Connect module with the OpenFGA API URL and authentication credentials: an API token, or client credentials (client ID, client secret, token issuer, and audience). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline authentication.

### OpenFGA Server

- **A running OpenFGA instance** -- self-hosted or cloud-hosted, reachable from the Planton Runner or provisioner environment at the API URL configured in the Provider Connection.

## Deploy

### Console

Open the deployment store, find **OpenFGA Store**, and click **Deploy**. The creation wizard walks you through environment and connection configuration and the single spec field -- the store name. Start from the **Standard Authorization Store** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: openfga.planton.dev/v1alpha1
kind: OpenFgaStore
metadata:
  name: prod-authz
  org: acme-corp
  env: prod
spec:
  name: production-authz
```

```shell
planton apply -f openfga-store.yaml
```

This creates a store named `production-authz` on the connected OpenFGA server and surfaces the generated store ID in `status.outputs`. OpenFGA ships only a Terraform provider, so this kind runs on OpenTofu or Terraform: set `planton.dev/provisioner: tofu` (or `terraform`), or leave it unset and the CLI asks between the two, and Pulumi is refused before anything runs. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an OpenFGA store. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Store name** -- The `name` field is immutable: changing it replaces the store, and the replacement destroys every authorization model and relationship tuple inside it. Name the store for its long-term grain (`production-authz`, `billing-permissions`), not the first project that happens to use it.

**Isolation grain** -- The one real decision this component carries: what a store represents. Models and tuples in one store can never affect checks in another, so the store boundary IS the authorization blast radius. Choose per-environment, per-application, or per-tenant (see Common Patterns) before writing tuples -- there is no move operation between stores, so changing grain later means re-writing every tuple.

**No deletion protection** -- Neither the OpenFGA API nor the Terraform provider exposes a deletion guard for stores. A destroy proceeds the moment it is issued and removes the store together with all models and tuples inside it. Where protection matters, it has to be operational -- restrict who can delete the Infra Component.

**Where the data actually lives** -- The store is a logical container inside the OpenFGA server; its contents persist in the server's backing datastore (PostgreSQL, MySQL, or SQLite). Backups, encryption at rest, and availability are that datastore deployment's posture -- nothing in this spec configures them.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies -- it is the root of the OpenFGA dependency graph.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | Server-generated store identifier (a ULID) | `storeId` on OpenFGA Authorization Model and OpenFGA Relationship Tuple resources |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard authorization store** -- One store holding all authorization data for an application: create the store, deploy an authorization model into it, then write relationship tuples. Start from the **Standard Authorization Store** preset.

**Per-environment stores** -- Separate stores for development, staging, and production. Each environment gets its own models and tuples, so a model change tested in the dev store cannot alter a single production access decision. This is the minimum isolation any real deployment should run with.

**Per-application stores** -- When multiple applications share one OpenFGA server but their authorization models are unrelated, a store per application avoids type-definition conflicts and lets each application version its model independently. Share a store only when the applications genuinely share one model.

**Per-tenant stores** -- A store per tenant gives complete authorization data isolation with a clean compliance and data-residency story, at the cost of store count growing with tenants -- every model change must then be rolled out once per store.

## Works With

- [**OpenFGA Authorization Model**](/infra-catalog/openfga-authorization-model) -- the schema of types and relations deployed into the store, referenced through the store's `id` output
- [**OpenFGA Relationship Tuple**](/infra-catalog/openfga-relationship-tuple) -- the authorization data written into the store, one tuple per resource
