# Overview

The **OCI Compartment API Resource** provides a consistent and standardized interface for deploying and managing compartments on Oracle Cloud Infrastructure. A compartment is OCI's fundamental organizational primitive — a logical container for resources that enables hierarchical isolation, fine-grained IAM policy scoping, and cost tracking.

## Purpose

This API resource streamlines the creation and management of OCI compartments as the foundation of resource organization. By offering a unified interface, it enables users to:

- **Establish Resource Hierarchies**: Create nested compartment trees that mirror organizational structures (tenancy > business unit > team > environment > workload). Each OciCompartment references its parent via `compartmentId`, and children reference it via `valueFrom`.
- **Scope IAM Policies**: Compartments are the primary boundary for OCI IAM policy statements. A policy attached to a compartment governs access to all resources within it and its descendants. Creating the right compartment structure determines the IAM policy model for the entire tenancy.
- **Track Costs**: OCI cost analysis is compartment-aware. Isolating workloads in separate compartments provides built-in cost attribution without additional tagging gymnastics.
- **Feed Every Downstream Resource**: The `compartmentId` output is consumed by every other OCI component in Planton — OciVcn, OciSubnet, OciComputeInstance, OciAutonomousDatabase, and all others. This makes OciCompartment the root of the dependency graph for OCI deployments.

## Key Features

- **Consistent Interface**: Aligns with the Planton pattern for deploying cloud infrastructure across providers.
- **Nested Hierarchies via Chaining**: Build arbitrarily deep compartment trees by referencing parent compartments through `valueFrom`. A platform team creates top-level compartments; application teams create child compartments within them.
- **Delete Protection by Default**: The `enableDelete` field defaults to `false`, meaning the compartment survives IaC resource destruction. This matches OCI's philosophy of preventing accidental deletion of compartments that may contain active resources. Ephemeral compartments (dev, CI) can opt in to deletion with `enableDelete: true`.
- **Automatic Tagging**: Standard Planton freeform tags are applied to every compartment (resource kind, resource ID, organization, environment, and user-defined labels from metadata).
- **Foreign Key Composability**: Exports the `compartmentId` output as a `StringValueOrRef` target. Every downstream OCI component can reference this output directly, creating a declarative dependency chain without hardcoded OCIDs.

## How OCI Compartments Differ from Other Providers

Understanding compartments is essential when coming from AWS, GCP, or Azure:

- **Compartments vs AWS Accounts**: AWS uses accounts as the primary isolation boundary. Multi-account architectures (via AWS Organizations) are the standard enterprise pattern. OCI uses compartments within a single tenancy, providing similar isolation without the operational overhead of managing separate accounts. Compartments can be nested up to six levels deep, while AWS accounts are flat (OU hierarchy provides grouping but not nested resource scoping).
- **Compartments vs GCP Folders/Projects**: GCP uses a three-level hierarchy: organization > folders > projects. OCI compartments serve the combined role of both folders (organizational grouping) and projects (resource containment). A single OCI tenancy with a well-designed compartment tree replaces the GCP pattern of many separate projects.
- **Compartments vs Azure Resource Groups**: Azure resource groups are flat containers within a subscription. OCI compartments are hierarchical, support IAM policy inheritance, and can be nested. The closest Azure equivalent to the full OCI compartment model is Management Groups + Subscriptions + Resource Groups combined.
- **First-Class IAM Boundary**: In OCI, IAM policies are always scoped to a compartment. The statement `Allow group NetworkAdmins to manage virtual-network-family in compartment networking` grants access only within the `networking` compartment and its children. This is why `compartmentId` is field 1 on every OCI component — the compartment determines who can access the resource.

## Critical Constraints

- **Name Uniqueness**: Compartment names must be unique among siblings within the same parent compartment. Different parent compartments can have children with the same name.
- **Description Required**: The OCI API requires a non-empty description on every compartment. This is enforced via `buf.validate` with `min_len = 1`.
- **Delete Protection**: When `enableDelete` is `false` (the default), destroying the IaC resource leaves the compartment intact in OCI. To actually delete the compartment, you must first set `enableDelete: true`, apply the change, then destroy. This two-step process is intentional friction.
- **Compartment Limits**: OCI tenancies have a default limit of 100 compartments. This limit can be increased via a service limit request but should inform hierarchy design — avoid creating a compartment per microservice if you have hundreds of services.
- **Move and Rename**: Compartments can be renamed and moved to a different parent, but these operations are not currently exposed through this component. They are rare, administrative-level operations typically performed through the OCI Console.

## Use Cases

- **Project Isolation**: One compartment per project, containing all of the project's VCNs, compute instances, databases, and other resources. IAM policies grant project teams access only to their compartment.
- **Team and Business Unit Hierarchies**: Top-level compartments for business units, nested compartments for teams within each BU, and further nesting for environments (dev, staging, prod) within each team.
- **Environment Separation**: Separate compartments for development, staging, and production. Each environment gets its own IAM policies, cost tracking, and resource quotas.
- **Shared Services**: A dedicated compartment for shared infrastructure (networking, security, logging) that other compartments consume via cross-compartment references.

## Production Features

This resource provides complete support for production-grade compartment management, including:

- **Delete Protection**: Prevents accidental compartment deletion by default, requiring explicit opt-in before destruction.
- **Freeform Tagging**: Standard Planton labels applied as OCI freeform tags for resource management, cost tracking, and compliance.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations producing identical outputs.
- **Foreign Key Composability**: Designed as the Layer 0 foundation that every downstream OCI component references via `StringValueOrRef` for the `compartmentId` output.
