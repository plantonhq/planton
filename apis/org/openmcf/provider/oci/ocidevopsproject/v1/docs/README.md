# OciDevopsProject — Design Notes

## Design Rationale

OciDevopsProject provisions a single DevOps project resource. The component is intentionally minimal — it creates the organizational container, not the CI/CD resources within it.

### Why not bundle build pipelines or repositories?

Build pipelines, deploy pipelines, code repositories, and connections have fundamentally different lifecycles from the project. The project is infrastructure that changes infrequently. Pipelines and repositories are created and modified by development teams as part of their workflow. Bundling them would create a single manifest that needs updating every time a team adds a pipeline — defeating the separation of infrastructure and application concerns.

### Why flatten notification_config.topic_id?

The OCI provider nests the notification topic inside a `notification_config` block with a single field (`topic_id`). Single-field wrapper blocks add YAML depth without adding meaning. Flattening to `notificationTopicId` at the spec level makes manifests more readable and matches the pattern used elsewhere in OpenMCF for single-field provider blocks.

### Why is the project name derived from metadata.name?

OCI DevOps project names must be unique within a compartment and are immutable after creation. Using `metadata.name` as the project name keeps the naming consistent with the OpenMCF resource identity and avoids a separate `displayName` field that could diverge from the resource name.

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| Exclude pipelines and repos | Clean separation of infra and CI/CD lifecycle | Pipelines managed separately |
| Flatten notification_config | Simpler YAML; one level less nesting | Deviates from raw provider schema |
| metadata.name as project name | Consistent naming; no divergence | Cannot set a different display name |
| notificationTopicId as StringValueOrRef | Composable with ONS topic resources | ONS topics are not yet an OpenMCF component |

## Resource Graph

```
OciDevopsProject
└── oci_devops_project (always)
    ├── notification_config.topic_id (from spec.notificationTopicId)
    ├── description (optional)
    └── outputs: project_id, namespace
```

## Deferred from v1

- **Build pipelines, deploy pipelines, repositories, connections** — different lifecycle; managed by development teams.
- **defined_tags / system_tags** — managed by platform. `freeform_tags` are auto-populated from `metadata.labels`.

## Freeform Tags

The module automatically populates freeform tags from metadata:

| Tag Key | Source |
|---------|--------|
| `resource` | `"true"` (constant) |
| `resource_kind` | `OciDevopsProject` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| All `metadata.labels` | Copied as-is |
