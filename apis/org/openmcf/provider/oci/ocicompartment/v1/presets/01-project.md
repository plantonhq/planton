# Project Compartment

This preset creates a long-lived compartment for a project, team, or workload. The compartment is retained even if the IaC resource is destroyed (`enableDelete: false`), which is OCI's safety mechanism to prevent accidental deletion of compartments containing active resources. This is the standard compartment configuration for any organizational use case and should be the default choice for production, staging, and shared-services compartments.

## When to Use

- Creating a compartment for a project, application, or service that will persist indefinitely
- Organizing resources for a team or business unit within the tenancy hierarchy
- Setting up shared-services compartments (networking, security, logging) that other compartments depend on
- Any compartment where accidental deletion would cause data loss or outages

## Key Configuration Choices

- **Delete protection on** (`enableDelete: false`) -- The compartment survives IaC destruction. This is the OCI default and the correct choice for any compartment containing production resources. To actually delete the compartment, you must first set `enableDelete: true`, apply, then destroy. This two-step process is intentional friction against accidental deletion.
- **Description required** -- OCI requires a description on every compartment. Use it to document what the compartment is for and which team owns it. This appears in the OCI Console and is the first thing operators see when navigating the compartment hierarchy.
- **Name falls back to metadata.name** -- The `name` field in the spec is optional. If omitted, the compartment is named after `metadata.name`. Set it explicitly only if the OCI compartment name needs to differ from the resource identifier (e.g., to include spaces or match an existing naming convention).

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<parent-compartment-or-tenancy-ocid>` | OCID of the parent compartment or tenancy root | OCI Console > Identity > Compartments (tenancy OCID is on the Tenancy Details page), or a parent `OciCompartment` status outputs |

## Related Presets

- **02-sandbox** -- Use instead for ephemeral dev/test compartments that should be destroyed along with their IaC resource
