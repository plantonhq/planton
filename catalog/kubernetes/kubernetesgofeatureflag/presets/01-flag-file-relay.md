# Flag File Relay

The production GO Feature Flag shape: two relay replicas behind a disruption budget, reading their flags from a KubernetesGoFeatureFlagFlagFile, with admin and evaluation keys guarding every endpoint. The flags live in their own resource, so flipping a flag edits only the flag file - the relay is never re-applied and picks the change up within its polling interval, with no restart.

## When to Use

- Release switches for your own services, evaluated by OpenFeature SDKs inside the cluster.
- Teams that want flag changes reviewed as code and kept separate from the relay's operations.
- Any install where the flags should be typed and validated before they reach the relay.

## Key Configuration Choices

- `flagSource.retrievers[0].configMap` references a KubernetesGoFeatureFlagFlagFile by name and reads its `flags.goff.yaml` key; the module grants the relay `get` on exactly that ConfigMap.
- `startWithRetrieverError: true` lets the relay start before the flag file exists, so the two resources can deploy in any order.
- `pollingIntervalMs: 30000` sets how quickly a flip takes effect (the relay default is 60000).
- `authorizedKeys` turns authentication on; without it anyone who can reach the Service can read every flag. Both keys are managed-secret references that reach the relay through the module-owned env Secret, never through its configuration.
- `replicas: 2` with `pdb.minAvailable: "1"` keeps evaluations answering through node drains.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<namespace>` | Namespace for the relay | Your cluster's namespace plan |
| `<flag-file-name>` | Name of the KubernetesGoFeatureFlagFlagFile holding the flags | The flag file resource you deploy alongside |
| `<relay-admin-key>` | Managed secret holding the admin API key | Create it in your organization's secrets |
| `<relay-evaluation-key>` | Managed secret holding the key SDKs present | Create it in your organization's secrets |

## Related Presets

- **Flags From GitHub** - the same relay reading a flag file from a repository, with change notifications.
- **Team Flag Sets** - isolated flag sets per team, each selected by its own API key.
