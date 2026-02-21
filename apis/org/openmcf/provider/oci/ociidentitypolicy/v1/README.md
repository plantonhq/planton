# Overview

The **OCI Identity Policy API Resource** provides a consistent and standardized interface for deploying and managing IAM policies on Oracle Cloud Infrastructure. A policy is OCI's authorization mechanism — it contains one or more human-readable statements that grant a group or dynamic group specific permissions on resources within a compartment.

## Purpose

This API resource streamlines the deployment and management of OCI IAM policies. By offering a unified interface, it enables users to:

- **Grant Compartment-Scoped Access**: Attach policies to compartments to grant groups specific permissions within that compartment and all of its children. Policies are the only way to grant access in OCI — without a policy, no group has any permissions.
- **Support Workload Identity**: Grant dynamic groups (compute instances, OKE pods, Functions) access to OCI services by referencing them as subjects in policy statements (`Allow dynamic-group ...`). This is the foundation of credential-less authentication in OCI.
- **Express Fine-Grained Authorization**: OCI's policy language supports four verbs (`inspect`, `read`, `use`, `manage`), resource type families (e.g., `virtual-network-family`, `object-family`), compartment scoping, and `where` conditions. A single policy resource can encode sophisticated access patterns in readable English statements.
- **Enable Infra-Chart Composability**: Export the policy OCID as a stack output for reference and auditing. Downstream automation can verify that required policies exist before deploying workloads.

## Key Features

- **Consistent Interface**: Aligns with the OpenMCF pattern for deploying cloud infrastructure across providers.
- **Human-Readable Statements**: Policy statements are written in OCI's English-like syntax (`Allow group Developers to use instances in compartment dev`), making policies auditable without tooling.
- **Compartment and Tenancy Scoping**: Policies can be attached to any compartment for scoped access, or to the tenancy root for organization-wide rules. The `compartmentId` field determines the attachment point.
- **Dynamic Group Support**: Statements can reference dynamic groups as subjects, enabling the workload identity pattern where compute instances and serverless functions authenticate to OCI services via instance principal.
- **Version Date Pinning**: The optional `versionDate` field locks policy evaluation to the service behavior on a specific date. This prevents unexpected permission changes when OCI updates service behavior.
- **Automatic Tagging**: Standard OpenMCF freeform tags are applied to every policy (resource kind, resource ID, organization, environment, and user-defined labels from metadata).
- **Foreign Key Composability**: The `compartmentId` field supports `valueFrom` references to OciCompartment resources, enabling declarative policy-to-compartment bindings without hardcoded OCIDs.

## How OCI Policies Differ from Other Providers

Understanding OCI's policy model is essential when coming from AWS, GCP, or Azure:

- **Policies vs AWS IAM Policies**: AWS IAM policies are JSON documents attached to users, groups, or roles. They define permissions inline with `Effect/Action/Resource` tuples. OCI policies are separate resources attached to compartments, written in English-like syntax. The key structural difference: AWS policies are identity-attached (who gets the permission), while OCI policies are location-attached (where the permission applies). In OCI, a single policy statement combines the subject (`Allow group X`), the permission (`to manage instances`), and the scope (`in compartment Y`) in one line.
- **Policies vs GCP IAM Bindings**: GCP IAM bindings associate a role with a set of members on a specific resource. OCI policies combine these concepts — a statement is simultaneously a binding (subject + permission) and a scope (compartment). GCP roles are predefined or custom; OCI uses verb + resource-type pairs (`manage instances`, `read secret-family`) as inline permission definitions.
- **Policies vs Azure RBAC**: Azure RBAC assigns built-in or custom roles to security principals at a scope (management group, subscription, resource group, or resource). OCI's model is simpler in that there are no role definitions — the verb + resource-type in the statement IS the permission. Azure's role assignment model is more structured; OCI's policy language is more readable and flexible.
- **English-Like Syntax**: OCI is unique among major cloud providers in using a human-readable policy language rather than JSON or structured bindings. Statements read as English sentences, making policies auditable by non-engineers (e.g., compliance officers, security reviewers).
- **Compartment Inheritance**: OCI policies grant access within the attachment compartment and all child compartments. A policy attached to the tenancy root applies everywhere. A policy attached to a child compartment applies only within that subtree. This inheritance model eliminates the need for AWS-style policy propagation or GCP-style hierarchical bindings.

## Critical Constraints

- **Name Uniqueness**: Policy names must be unique across all policies in the tenancy (not just within the compartment). The name cannot be changed after creation.
- **Statement Syntax**: Statements must follow OCI's policy language syntax exactly. Invalid statements cause the OCI API to reject the entire policy. The format is: `Allow <subject> to <verb> <resource-type> in <location> [where <conditions>]`.
- **No Deny Rules**: OCI policies only grant access (`Allow`). There is no explicit deny mechanism. Access is implicitly denied unless a policy grants it. This simplifies policy evaluation but means you cannot override a broad `Allow` with a more specific `Deny`.
- **Compartment Scoping**: A policy can only grant access within the compartment it is attached to and that compartment's children. To grant cross-compartment access, the policy must be attached to a common ancestor or the tenancy root.
- **Statements Are Ordered but Evaluated as a Set**: The order of statements in a policy has no effect on evaluation. OCI evaluates all statements across all policies simultaneously — if any statement grants access, the request is allowed.

## Use Cases

- **Team Access Management**: Create a policy per team granting the team's group access to their compartment. The platform team gets `manage all-resources` on the infrastructure compartment; the application team gets `use instances` and `manage object-family` on their workload compartment.
- **Workload Identity**: Pair with OciDynamicGroup to enable credential-less authentication. Create a dynamic group matching compute instances, then create a policy granting that dynamic group access to Vault secrets, KMS keys, or Object Storage.
- **Security Auditing**: Create a tenancy-level policy granting auditors `inspect all-resources` and `read audit-events` across the entire tenancy. The `inspect` verb provides metadata visibility without data access.
- **Delegated Administration**: Attach a policy to a compartment granting a group `manage` on specific resource types within that subtree. The group administers resources within their scope without seeing or affecting resources in sibling compartments.
- **Service-Level Isolation**: Use resource type families (`database-family`, `virtual-network-family`, `object-family`) to grant access to specific service categories rather than `all-resources`. This follows the principle of least privilege.

## Production Features

This resource provides complete support for production-grade IAM policy management, including:

- **Multiple Statement Support**: A single policy can contain multiple statements, each granting different permissions to different subjects. This keeps related access rules together in one auditable resource.
- **Version Date Pinning**: Lock policy evaluation to a specific date, preventing unexpected permission changes when OCI updates service behavior.
- **Freeform Tagging**: Standard OpenMCF labels applied as OCI freeform tags for policy tracking, compliance, and organizational reporting.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations producing identical resource topology and outputs.
- **Proto Validation**: Required fields, minimum statement count, and description constraints are validated at the schema level before deployment.
- **Foreign Key Composability**: Designed to reference OciCompartment resources via `valueFrom` for declarative compartment-to-policy bindings.
