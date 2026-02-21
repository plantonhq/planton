# Compartment Admin Policy

This preset creates an IAM policy granting a group full administrative access to all resources within a compartment. This is the most common OCI policy pattern -- the first thing every team creates after setting up a new compartment. The policy is attached to the target compartment itself, following OCI's recommended practice of keeping policies close to the resources they govern.

## When to Use

- A team or project just received a new compartment and needs full administrative control over it
- You want a single group to manage all resource types (compute, networking, storage, databases, etc.) within a compartment
- Setting up initial access before creating more fine-grained policies for specific roles
- Any scenario where one group should have unrestricted control over a compartment's resources

## Key Configuration Choices

- **`manage all-resources`** -- The broadest OCI permission grant. The `manage` verb includes `inspect`, `read`, `use`, and all create/update/delete operations. `all-resources` covers every resource type in the compartment. This is appropriate for compartment administrators who need to create and operate any type of infrastructure.
- **Policy attached to the compartment** (`compartmentId` points to the target compartment) -- OCI policies grant access within the compartment they are attached to, plus all child compartments. By attaching the policy to the same compartment it governs, permissions are scoped exactly where needed and follow OCI's recommended practice.
- **Single statement** -- Keeps the policy focused on one concern: admin access. If additional groups need different access levels (e.g., read-only for developers), create a separate `OciIdentityPolicy` resource rather than adding statements here. This keeps policies auditable and independently manageable.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment this policy is attached to and governs | OCI Console > Identity > Compartments, or `OciCompartment` status outputs (`compartmentId`) |
| `<admin-group-name>` | Name of the IAM group receiving admin access | OCI Console > Identity > Groups |
| `<compartment-name>` | Display name of the target compartment (must match exactly) | OCI Console > Identity > Compartments, or the `name`/`metadata.name` of the `OciCompartment` resource |

## Related Presets

- **02-service-access** -- Use instead when granting OCI service access to compute instances or OKE pods via a dynamic group (workload identity pattern)
- **03-read-only-auditor** -- Use instead when granting inspect-level visibility for compliance or security audit purposes
