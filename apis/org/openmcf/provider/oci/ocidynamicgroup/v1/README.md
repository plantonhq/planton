# Overview

The **OCI Dynamic Group API Resource** provides a consistent and standardized interface for deploying and managing dynamic groups on Oracle Cloud Infrastructure. A dynamic group is OCI's workload identity mechanism — it allows compute instances, OKE pods, Functions, and other OCI resources to authenticate to OCI services without stored credentials by matching them to a group via a rule-based membership system.

## Purpose

This API resource streamlines the creation and management of OCI dynamic groups as the identity foundation for workload authentication. By offering a unified interface, it enables users to:

- **Enable Instance Principal Authentication**: Match compute instances in a compartment so they can call OCI APIs (read Vault secrets, use KMS keys, access Object Storage) without embedded API keys or credentials. This is the OCI equivalent of AWS IAM roles for EC2 or GCP service accounts for Compute Engine.
- **Enable Serverless Workload Identity**: Match OCI Functions by resource type and compartment, allowing serverless workloads to interact with OCI services during execution without credentials in function configuration or code.
- **Define Rule-Based Membership**: Use OCI's matching rule syntax (`Any {conditions}` or `All {conditions}`) to dynamically determine group membership based on compartment, resource type, or freeform tags. Members are automatically added or removed as resources are created or destroyed.
- **Feed the Two-Resource Identity Pattern**: Dynamic groups define *who* (which resources are members) while OciIdentityPolicy defines *what they can do* (which permissions are granted). Every dynamic group needs a companion policy to be useful.

## Key Features

- **Consistent Interface**: Aligns with the OpenMCF pattern for deploying cloud infrastructure across providers.
- **Flexible Matching Rules**: Support for `Any` (match any condition) and `All` (match all conditions) rule syntax. Match by compartment OCID, resource type, or freeform tag values. Rules can combine multiple conditions for precise membership.
- **Tenancy-Level Placement**: Dynamic groups are created in the tenancy root compartment, making them visible and referenceable across the entire compartment hierarchy. A dynamic group in the tenancy can match resources in any child compartment.
- **Automatic Tagging**: Standard OpenMCF freeform tags are applied to every dynamic group (resource kind, resource ID, organization, environment, and user-defined labels from metadata).
- **Foreign Key Composability**: The `compartmentId` field supports `valueFrom` references to OciCompartment resources. The `dynamic_group_id` output is available for reference by downstream automation and auditing tools.

## How OCI Dynamic Groups Differ from Other Providers

Understanding dynamic groups is essential when coming from AWS, GCP, or Azure:

- **Dynamic Groups vs AWS IAM Roles for EC2/EKS**: AWS uses IAM roles attached to EC2 instance profiles or EKS service accounts. The role is explicitly assigned to a specific instance or pod configuration. OCI dynamic groups use a matching rule to *implicitly* select instances based on their properties (compartment, tags, resource type). The matching approach means new instances in a compartment automatically become members without updating the group definition. AWS requires updating the instance profile or service account to change role assignments.
- **Dynamic Groups vs GCP Workload Identity**: GCP's workload identity binds a Kubernetes service account to a GCP IAM service account. This is an explicit binding — each service account must be individually configured. OCI dynamic groups achieve the same goal (credential-less API access) with a rule-based approach that automatically includes all matching resources.
- **Dynamic Groups vs Azure Managed Identities**: Azure uses system-assigned or user-assigned managed identities attached to specific resources. Like AWS, this is an explicit per-resource assignment. OCI's matching rule approach is more declarative — you define the criteria for membership rather than assigning identity resource-by-resource.
- **Shared Namespace with User Groups**: OCI dynamic group names share the same namespace as user group names within a tenancy. A dynamic group and a user group cannot have the same name. This is an OCI API constraint with no equivalent in other providers.
- **Tenancy-Level Only**: Dynamic groups must be created in the tenancy root compartment, unlike IAM policies which can be attached to any compartment. This is because dynamic groups need to match resources across the entire tenancy — a compartment-scoped dynamic group would not be able to see resources in sibling compartments.

## Critical Constraints

- **Tenancy-Level Creation Only**: Dynamic groups must be created in the tenancy root compartment. The `compartmentId` field must reference the tenancy OCID, not a child compartment OCID. The OCI API returns an error if you attempt to create a dynamic group in a child compartment.
- **Name Uniqueness**: Dynamic group names share a namespace with user group names. The name must be unique across all groups (dynamic and user) in the tenancy and cannot be changed after creation.
- **Matching Rule Required**: At least one matching rule must be provided. Rules use OCI's specific syntax — common patterns include `Any {instance.compartment.id = 'ocid1...'}` for compartment-scoped matching and `All {resource.type = 'fnfunc', resource.compartment.id = 'ocid1...'}` for type-and-compartment matching.
- **No Permissions Without a Policy**: A dynamic group by itself grants no permissions. It only defines group membership. To grant permissions, create a companion `OciIdentityPolicy` with statements referencing the dynamic group name (e.g., `Allow dynamic-group compute-workers to read secret-family in compartment production`).
- **Membership Is Eventually Consistent**: When a new compute instance is launched in a compartment matched by a dynamic group, there can be a short delay before the instance is recognized as a member. This is typically under a minute but should be accounted for in bootstrapping scripts.

## Use Cases

- **Compute Instance Principal**: Match all compute instances in a production compartment. Create a companion policy granting the dynamic group access to Vault secrets, KMS keys, and Object Storage. Instances authenticate via the instance metadata service without any credentials in the image or user data.
- **OKE Node-Level Access**: Match OKE worker nodes by compartment and tag to grant them access to OCI services at the node level. This is useful for kubelet-level operations (pulling images from Container Registry, accessing Object Storage for logs) that operate outside of Kubernetes RBAC.
- **Serverless Functions Identity**: Match all Functions in a compartment using `resource.type = 'fnfunc'`. Grant the dynamic group permission to read secrets, write to Object Storage, or push to Streaming during function execution.
- **Tag-Based Workload Isolation**: Use freeform tag conditions in matching rules to create fine-grained groups that span compartments or select subsets within a compartment. For example, only instances tagged with `workload-identity=enabled` become members, allowing opt-in rather than blanket membership.
- **Multi-Tier Identity Separation**: Create separate dynamic groups for web tier, application tier, and database tier instances. Each dynamic group gets its own companion policy with tier-appropriate permissions — the web tier can read configs, the app tier can access secrets and keys, and the database tier can manage backups.

## Production Features

This resource provides complete support for production-grade dynamic group management, including:

- **Rule-Based Membership**: Dynamic membership via matching rules that support compartment scoping, resource type filtering, and tag-based selection.
- **Freeform Tagging**: Standard OpenMCF labels applied as OCI freeform tags for group tracking, compliance, and organizational reporting.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations producing identical resource topology and outputs.
- **Proto Validation**: Required fields (compartmentId, description, matchingRule) are validated at the schema level with minimum-length constraints before deployment.
- **Foreign Key Composability**: Designed to reference OciCompartment resources via `valueFrom` for the tenancy OCID, and to be referenced by name in OciIdentityPolicy statements for the two-resource identity pattern.
