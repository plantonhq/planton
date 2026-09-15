# DigitalOcean Project

Deploys a DigitalOcean project -- the account-level container that organizes droplets, load balancers, domains, buckets, and most other resources into named groups with a purpose and an environment. Membership is declared on the project itself: the `resources` list wires members by reference to their `urn` outputs or by literal URN, and removing a member relocates it to the account's default project -- membership changes never destroy anything. Leave the list empty and membership stays unmanaged entirely.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Project** -- one `digitalocean_project` resource: the named container with description, purpose, and environment
- **Membership assignments** -- configured only when `resources` is set: each listed URN is moved into the project, and removing one from the list moves it back to the account's default project. An empty list leaves membership unmanaged, so console assignments and the members' own project selections are left untouched.

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Member resources** (only for managed membership) -- the Cloud Resources joining by reference must exist or deploy in the same InfraPipeline; pre-existing resources join by literal URN (`do:<type>:<id>`).

### DigitalOcean Account

- **Nothing** -- a project is an organizational container; creating one provisions no billable infrastructure.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Project**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Production Environment Project** preset in the [Presets](#presets) tab for a per-environment container with unmanaged membership.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanProject
metadata:
  name: web-production
  org: acme-corp
  env: prod
spec:
  projectName: web-production
  description: Production web workloads
  purpose: Web Application
  environment: production
```

```shell
planton apply -f do-project.yaml
```

This creates a production-labeled project with membership left unmanaged -- resources join from the console or by their own project selections without the manifest fighting them. A Stack Job tracks the provisioning in real time.

### InfraChart

When the project owns its membership, wire members by reference to their `urn` outputs. The list is polymorphic across kinds -- no single default kind applies -- so each reference names its own `kind`:

```yaml
spec:
  projectName: orders-service
  description: Everything the orders service runs on
  purpose: Service or API
  environment: staging
  resources:
    - valueFrom:
        kind: DigitalOceanDroplet
        name: orders-worker
        fieldPath: status.outputs.urn
    - value: do:space:orders-assets
```

The InfraPipeline resolves the dependency graph, deploys the droplet first, then creates the project with the resolved URN; the pre-existing bucket joins by literal URN.

## Key Configuration

These are the most important decisions when configuring a project. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Manage membership -- or leave it alone** -- an empty `resources` list means "do not manage membership": resources assigned from the console or by their own project fields are left untouched. A declared list owns the FULL membership list. Use the empty form when teams assign resources manually and the manifest should only own the container.

**One project per resource** -- a resource belongs to exactly one project. Declaring it in this project's `resources` list MOVES it, including out of another project that also claims it -- two projects listing the same resource will fight forever. Give each resource one home, and prefer wiring membership by reference so the graph is visible in code.

**Destroy relocates, never deletes** -- DigitalOcean requires a project to be empty before deletion, so both provisioners relocate every member to the account's default project first and retry the delete while the asynchronous moves settle. In an environment teardown, destroy the member resources first -- otherwise they keep running, and billing, from the default project.

**purpose round-trips -- with one trap** -- DigitalOcean recognizes standard purposes ("Web Application", "Website or blog", "Service or API") and stores anything else prefixed as `Other: <text>`, stripping the prefix on read, so free text converges cleanly. The one value that can never converge is text that itself starts with `Other:` -- DigitalOcean keeps exactly one prefix and re-capitalizes the rest (`Other: probe` is stored as `Other: Probe`), the provider strips the prefix on read, and the result can never equal what you wrote -- so validation rejects it up front.

**environment is lowercase** -- declare `development`, `staging`, or `production`; DigitalOcean reports the value back capitalized and the provisioners absorb the difference. Do not "fix" it by writing the capitalized form -- validation rejects it to keep one canonical spelling.

**isDefault: almost always leave unset** -- the account can have only one default project, so out-of-band changes to the default show up here as drift, and a project marked default refuses deletion. Set it only if the account's default is genuinely meant to be managed as code.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **Any DigitalOcean kind** (polymorphic membership) | `resources[]` | `status.outputs.urn` |

The list carries no default kind -- each `valueFrom` names its own `kind` (droplets, load balancers, buckets, domains, ...), or the entry is a literal URN.

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `project_id` | UUID of the project -- the API identity and the import id | `projectId` on droplet autoscale pools, so pool members land in this project |

The remaining outputs: `owner_uuid` and `owner_id` identify the account or team that owns the project (account facts with no downstream ValueFromRef story), and `resource_urns` lists, sorted, the members DigitalOcean reports after apply -- the managed membership when `resources` is set, or whatever the account has assigned out of band when it is not.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**One project per environment** -- a production/staging/development container with membership left unmanaged, giving the console a per-environment view of resources and billing while resources join by their own selections. Start from the **Production Environment Project** preset.

**Per-service project that owns its membership** -- the project and its members deploy together, mixing chart-managed members (by reference to `urn` outputs) with pre-existing ones (by literal URN), so the console and billing views match the architecture. Start from the **Service Project with Managed Membership** preset.

## Works With

- [**DigitalOcean Droplet**](/cloud-catalog/digital-ocean-droplet) -- joins the project by reference to its `urn` output
- [**DigitalOcean Droplet Autoscale Pool**](/cloud-catalog/digital-ocean-droplet-autoscale-pool) -- consumes this project's `project_id` so pool members are created here
- [**DigitalOcean DNS Zone**](/cloud-catalog/digital-ocean-dns-zone) -- joins by reference to its `urn` output (`do:domain:example.com`)
- [**DigitalOcean Spaces Bucket**](/cloud-catalog/digital-ocean-bucket) -- joins by literal Spaces URN (`do:space:<name>`)
