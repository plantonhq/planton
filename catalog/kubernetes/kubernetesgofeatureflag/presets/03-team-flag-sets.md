# Team Flag Sets

One relay serving several isolated flag sets: each team owns its own flag file, and the API key a caller presents selects which set it evaluates. Flag sets share nothing - a key for one set can never read another's flags.

## When to Use

- Several teams or products sharing one relay without seeing each other's flags.
- Separate flag sets for web and mobile clients with their own keys.
- Consolidating several small relays into one.

## Key Configuration Choices

- `flagSets.items` replaces `flagSource`: the relay runs in exactly one of the two modes.
- Each set carries at least one API key in `apiKeys`, and set names are unique and never `default`; a key shared between sets fails the deploy.
- Each set reads its own KubernetesGoFeatureFlagFlagFile; the module grants the relay `get` on every ConfigMap the sets name.
- `startWithRetrieverError: true` per set lets the relay start before the flag files exist.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<namespace>` | Namespace for the relay | Your cluster's namespace plan |
| `<web-flag-file-name>` | Flag file of the web set | The web team's KubernetesGoFeatureFlagFlagFile |
| `<mobile-flag-file-name>` | Flag file of the mobile set | The mobile team's KubernetesGoFeatureFlagFlagFile |
| `<web-team-key>` | Managed secret holding the web set's key | Create it in your organization's secrets |
| `<mobile-team-key>` | Managed secret holding the mobile set's key | Create it in your organization's secrets |

## Related Presets

- **Flag File Relay** - one flag set served to every caller.
- **Flags From GitHub** - flags read from a repository with change notifications.
