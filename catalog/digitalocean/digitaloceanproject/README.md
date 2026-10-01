# DigitalOcean Project

Built for 100% parity with the Terraform DigitalOcean provider's `digitalocean_project` resource at the pinned provider version.

## What this component models

A DigitalOcean project -- the account-level container that organizes droplets, load balancers, domains, buckets, and most other resources into named groups with a purpose and an environment. Membership is carried here on the project itself as resource URNs; DigitalOcean's standalone partial-ownership membership resource is deliberately not modeled (one project object owns its full membership list, which is also how the API reports it back).

The component covers the provider's full argument surface:

- `project_name` -- the display name (1-175 characters)
- `description` -- optional free text (up to 255 characters)
- `purpose` -- optional; DigitalOcean recognizes standard purposes ("Web Application", "Service or API", ...) and stores anything else as `Other: <text>`, stripping the prefix on read so free text round-trips cleanly; values that themselves start with `Other:` are rejected at validation (they would never converge)
- `environment` -- optional; `development`, `staging`, or `production` (lowercase canonical; DigitalOcean reports it back capitalized)
- `is_default` -- optional; make this the account's default project (semi-supported upstream -- see the GUIDE)
- `resources` -- optional membership list: literal URNs (`do:droplet:123456`) or references to producing kinds' `urn` outputs with an explicit `valueFrom.kind` (the list is polymorphic across kinds)

## Quick start

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanProject
metadata:
  name: web-production
spec:
  projectName: web-production
  environment: production
  purpose: Web Application
```

Deploy with either provisioner; both produce identical resources and outputs.

## Outputs

| Output | Description |
|---|---|
| `project_id` | UUID of the project (the API identity, and the import id) |
| `owner_uuid` | UUID of the owning account or team |
| `owner_id` | Numeric id of the owning account or team |
| `resource_urns` | Sorted URNs of the members DigitalOcean reports after apply -- the managed membership when `resources` is set, whatever the account assigned out of band when it is not |

## Behavior worth knowing

- **Destroy evacuates, never destroys.** Deleting the project relocates every member resource to the account's default project and retries while the asynchronous moves settle. Nothing inside is ever destroyed. The moves can take minutes (measured: seconds on most destroys, over three minutes once), so both provisioners give the delete ten minutes instead of the provider's three; a destroy that still reports `cannot delete a project with resources` has already moved the members -- run it again.
- **Membership moves resources.** A resource belongs to exactly one project: listing it here moves it from wherever it was; removing it from the list moves it to the default project. An empty list means membership is not managed at all.
- **The `Other:` purpose trap is unrepresentable.** DigitalOcean prefixes non-standard purposes with `Other: ` (keeping exactly one prefix and re-capitalizing the rest) and the provider strips it on read; a user-supplied value already carrying the prefix would drift forever, so validation rejects it.
- **Never adopt the account's default project.** It reads back `is_default: true` against a manifest that leaves the flag unset, so the first apply would try to un-default it.
- **The account's default project cannot be deleted**, and DigitalOcean documents that a managed project should not be MADE the default (see the GUIDE).

## Module layout

- `iac/tf/` -- OpenTofu/Terraform module (provider pinned `~> 2.99`)
- `iac/pulumi/` -- Pulumi module (Go, pulumi-digitalocean SDK)
- Both engines wire the same spec fields and export the same outputs; behavioral parity is the contract.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
