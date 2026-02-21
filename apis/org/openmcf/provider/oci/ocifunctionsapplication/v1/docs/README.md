# OciFunctionsApplication — Design Notes

## Design Rationale

OciFunctionsApplication provisions a single Functions application resource. The component manages the execution environment for serverless functions but intentionally excludes the functions themselves.

### Why not bundle individual functions?

Functions have a fundamentally different deployment lifecycle from the application. The application is infrastructure (networking, architecture, policies) that changes infrequently. Functions are code artifacts deployed via `fn deploy` or CI/CD pipelines on every commit. Bundling functions in IaC would require re-running Pulumi on every code change, which defeats the purpose of serverless.

### Why is shape immutable?

OCI Functions binds the processor architecture to the application at creation time. All functions in the application must match the application's shape. Changing the shape would require redeploying all function images for the new architecture — a breaking change best handled by creating a new application.

### Why is image policy config inline?

Image signature verification is a property of the application in the OCI API. It governs which images can be deployed to any function in the application. Managing it separately would create a misleading abstraction — the policy doesn't exist independently of the application.

### Why support multiple subnets?

OCI Functions can be configured with multiple subnets for high availability across fault domains. The repeated `subnetIds` field reflects this capability. All subnets must be in the same VCN.

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| Exclude individual functions | Clean IaC/code separation; no Pulumi on code changes | Functions managed separately via fn CLI or CI/CD |
| Immutable shape | Prevents architecture mismatch with deployed images | Must create new application to change architecture |
| Image policy inline | Matches OCI API; atomic updates with app config | All image policy changes require application update |
| Multiple subnet support | High availability across fault domains | More complex subnet configuration |

## Resource Graph

```
OciFunctionsApplication
└── oci_functions_application (always)
    ├── subnet_ids (1..N)
    ├── network_security_group_ids (0..N)
    ├── config (0..N key-value pairs)
    ├── image_policy_config (optional)
    │   └── key_details (1..N if enabled)
    ├── trace_config (optional)
    └── outputs: application_id
```

## Deferred from v1

- **Individual functions** — code artifacts with different lifecycle; managed via `fn deploy` or CI/CD.
- **defined_tags / system_tags** — managed by platform. `freeform_tags` are auto-populated from `metadata.labels`.
- **security_attributes** — Oracle ZPR (Zero-Trust Packet Routing) attributes; very low adoption.

## Freeform Tags

The module automatically populates freeform tags from metadata:

| Tag Key | Source |
|---------|--------|
| `resource` | `"true"` (constant) |
| `resource_kind` | `OciFunctionsApplication` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| All `metadata.labels` | Copied as-is |
