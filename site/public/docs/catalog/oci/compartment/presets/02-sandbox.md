---
title: "Sandbox Compartment"
description: "This preset creates an ephemeral compartment for development, testing, or proof-of-concept work. Unlike the project preset, this compartment is destroyed when the IaC resource is destroyed..."
type: "preset"
rank: "02"
presetSlug: "02-sandbox"
componentSlug: "compartment"
componentTitle: "Compartment"
provider: "oci"
icon: "package"
order: 2
---

# Sandbox Compartment

This preset creates an ephemeral compartment for development, testing, or proof-of-concept work. Unlike the project preset, this compartment is destroyed when the IaC resource is destroyed (`enableDelete: true`), making it safe for automated teardown workflows. Use this for any compartment that is temporary by design -- CI/CD pipelines, developer sandboxes, training environments, or short-lived experiments.

## When to Use

- Developer sandbox environments that are spun up and torn down regularly
- CI/CD pipelines that create temporary compartments for integration testing
- Proof-of-concept or demo environments with a defined expiration
- Training or workshop environments that need clean teardown after the event
- Any compartment where the resources inside are also ephemeral and fully IaC-managed

## Key Configuration Choices

- **Delete enabled** (`enableDelete: true`) -- The compartment is deleted when the IaC resource is destroyed. This is the critical difference from the project preset. OCI will refuse to delete a compartment that still contains active resources, so ensure all child resources are destroyed first (or are managed by IaC in the same stack). Without this flag, destroying the IaC resource leaves the compartment orphaned in OCI.
- **Description signals intent** -- The description explicitly marks this as an ephemeral sandbox. This is important for operators browsing the OCI Console: it signals that the compartment is temporary and can be cleaned up if its resources appear abandoned.
- **Same parent pattern** -- The parent `compartmentId` works identically to the project preset. Sandboxes are commonly nested under a dedicated "sandboxes" or "development" parent compartment to keep the hierarchy tidy.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<parent-compartment-or-tenancy-ocid>` | OCID of the parent compartment or tenancy root | OCI Console > Identity > Compartments (tenancy OCID is on the Tenancy Details page), or a parent `OciCompartment` status outputs |

## Related Presets

- **01-project** -- Use instead for long-lived compartments that should survive IaC destruction (production, staging, shared services)
