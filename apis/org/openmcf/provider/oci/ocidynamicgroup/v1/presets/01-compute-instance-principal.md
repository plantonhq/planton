# Compute Instance Principal

This preset creates a dynamic group that matches all compute instances in a specific compartment. Combined with an `OciIdentityPolicy`, this enables instance principal authentication -- OCI's mechanism for letting compute instances call OCI APIs without storing credentials. This is the OCI equivalent of AWS IAM roles for EC2 or GCP service accounts for Compute Engine, and is the most common dynamic group pattern.

## When to Use

- Compute instances that need to read secrets from OCI Vault, access Object Storage, or call any OCI API
- OKE worker nodes that use instance principal authentication for node-level OCI service access
- Automation or application workloads running on OCI compute that should authenticate without embedded API keys
- Any scenario where you need credential-less authentication for instances in a compartment

## Key Configuration Choices

- **Compartment-scoped matching rule** (`Any {instance.compartment.id = '...'}`) -- Matches all compute instances in the specified compartment. The `Any` keyword means a resource only needs to satisfy one condition to be included. This is the broadest and most common pattern. For tighter scoping (e.g., only instances with a specific tag), modify the rule after deploying.
- **Tenancy-level placement** (`compartmentId` is the tenancy OCID) -- Dynamic groups are tenancy-level IAM resources in OCI. They must be created in the tenancy root compartment, not in a child compartment. The `compartmentId` here is the tenancy OCID, while the matching rule references the target compartment where instances live.
- **Name falls back to metadata.name** -- The `name` field in the spec is not set, so it defaults to `metadata.name`. Dynamic group names must be unique across all groups (including user groups) in the tenancy and cannot be changed after creation.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<tenancy-ocid>` | OCID of the tenancy root compartment (dynamic groups are tenancy-level resources) | OCI Console > Tenancy Details page, or `oci iam tenancy get` CLI command |
| `<compartment-ocid>` | OCID of the compartment whose compute instances should be included in this dynamic group | OCI Console > Identity > Compartments, or `OciCompartment` status outputs (`compartmentId`) |

## Related Presets

- **02-functions-workload-identity** -- Use instead when grouping OCI Functions for serverless workload identity
