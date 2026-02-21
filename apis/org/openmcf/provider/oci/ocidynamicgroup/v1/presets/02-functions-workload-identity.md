# Functions Workload Identity

This preset creates a dynamic group that matches all OCI Functions in a specific compartment. Combined with an `OciIdentityPolicy`, this enables serverless workload identity -- letting Functions call OCI APIs (read Vault secrets, write to Object Storage, use KMS keys, etc.) during execution without embedding credentials in the function code. This is the serverless equivalent of instance principal authentication for compute instances.

## When to Use

- OCI Functions that need to read secrets from OCI Vault at invocation time
- Serverless workloads that write to Object Storage, push to Streaming, or interact with any OCI service
- Functions that need to encrypt or decrypt data using OCI KMS keys
- Any OCI Functions application where credentials should not be stored in function configuration or code

## Key Configuration Choices

- **Functions-specific matching rule** (`All {resource.type = 'fnfunc', resource.compartment.id = '...'}`) -- The `All` keyword requires every condition to be satisfied: the resource must be of type `fnfunc` (an OCI Function) AND must reside in the specified compartment. This is more restrictive than `Any` because it narrows to a specific resource type, preventing compute instances or other resources in the same compartment from inheriting the permissions.
- **`fnfunc` resource type** -- This is OCI's internal resource type identifier for Functions. It matches individual function resources within a Functions Application. The matching rule does not need to reference the application OCID because compartment scoping is sufficient for most deployments.
- **Tenancy-level placement** (`compartmentId` is the tenancy OCID) -- Dynamic groups are tenancy-level IAM resources in OCI. They must be created in the tenancy root compartment, not in a child compartment. The `compartmentId` here is the tenancy OCID, while the matching rule references the target compartment where functions are deployed.
- **Name falls back to metadata.name** -- The `name` field in the spec is not set, so it defaults to `metadata.name`. Dynamic group names must be unique across all groups (including user groups) in the tenancy and cannot be changed after creation.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<tenancy-ocid>` | OCID of the tenancy root compartment (dynamic groups are tenancy-level resources) | OCI Console > Tenancy Details page, or `oci iam tenancy get` CLI command |
| `<compartment-ocid>` | OCID of the compartment whose functions should be included in this dynamic group | OCI Console > Identity > Compartments, or `OciCompartment` status outputs (`compartmentId`) |

## Related Presets

- **01-compute-instance-principal** -- Use instead when grouping compute instances for instance principal authentication
