# Flag File Source

The standard flagd shape: two stateless replicas serving the flags declared in a KubernetesFlagdFlagFile, with a PodDisruptionBudget so a node drain never takes evaluation down. The flag file renders a ConfigMap that this daemon mounts as a directory, so a flag flip edits only the flag file and reaches flagd within one to two minutes, with no restart.

## When to Use

- Flags are declared in Planton next to the services that evaluate them
- A team owns its flags separately from the platform team that runs flagd
- You want every flag change reviewed as a manifest change

## Key Configuration Choices

- `sources[0].configMap.configMapName` references a KubernetesFlagdFlagFile through `status.outputs.config_map_name`; the flag file must live in this namespace (pod volumes cannot cross namespaces)
- `sources[0].configMap.key` references the same flag file's `status.outputs.key` (a literal `value: flags.flagd.json` works too)
- `createNamespace: false`: the namespace exists before the flag file, which renders its ConfigMap there first
- `replicas: 2` with `pdb.enabled` keeps one pod serving through voluntary disruptions
- `log.format: json` for log pipelines

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<namespace>` | Namespace for flagd and its flag file | Your cluster's namespace plan |
| `my-flags` | Name of the KubernetesFlagdFlagFile to serve | The flag file's `metadata.name` |

## Related Presets

- **HTTP Source** -- flags served from a remote URL instead of a flag file
- **Autoscaled Zone Spread** -- the production shape for heavy evaluation traffic
