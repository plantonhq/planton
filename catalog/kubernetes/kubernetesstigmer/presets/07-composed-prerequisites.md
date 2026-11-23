# Composed Prerequisites

Stigmer on a cluster where CloudNativePG and its Barman Cloud backup plugin come from the catalog, as their own resources, ordered first with `depends_on`. Stigmer's operator installs a prerequisite only when it is missing and leaves a copy someone else installed alone, so with these deployed first it uses them, and a destroy of Stigmer never removes them.

## When to Use

- Other teams on the cluster run their own databases on CloudNativePG
- You manage cluster-level operators as their own resources, with their own upgrades

## Key Configuration Choices

- **`metadata.relationships`** — `depends_on` the KubernetesCloudNativePgOperator and the KubernetesCnpgBarmanCloudPlugin, so the InfraPipeline deploys them first. Without the order, Stigmer's operator could install its own copies first, and the catalog's resources would then find someone else's install and refuse it
- **`prerequisites`** — left at `auto`: the operator finds the catalog's copies and never touches them. `skip` would also work, and refuses in the status if one is missing
- **The same shape** fits agent-sandbox (KubernetesAgentSandboxController) and Agent Substrate (KubernetesAgentSubstrate) when sessions run in sandboxes

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `cloudnative-pg`, `barman-cloud` | Your prerequisite resources | The resources in this environment |
| `storage` | Your KubernetesSeaweedFs | The SeaweedFS resource in this environment |

## Related Presets

- **03-bucket-and-backups** — the same install with the operator's own prerequisites
