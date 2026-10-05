# Flags From GitHub

A relay that reads its flag file straight from a GitHub repository, so every flag change is a reviewed pull request, and posts every change to a Slack channel. The relay polls the repository once a minute and serves the new flags with no restart.

## When to Use

- Flags owned in a repository with pull-request review and history.
- Teams that want a channel notified on every flag created, changed or removed.
- Relays outside a cluster's ConfigMap workflow, or shared across clusters.

## Key Configuration Choices

- `flagSource.retrievers[0].github` names the repository, file path and branch; the `token` is required for a private repository and lifts the anonymous rate limit.
- `notifiers[0].slack.webhookUrl` posts each flag change; the URL embeds the channel's credential, so it is a managed secret.
- `pollingIntervalMs: 60000` keeps well inside GitHub's rate limits.
- `authorizedKeys.evaluation` guards evaluation; the token, the webhook URL and the key all reach the relay as environment variables from the module-owned Secret.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<namespace>` | Namespace for the relay | Your cluster's namespace plan |
| `<owner>/<repository>` | Repository holding the flag file | GitHub |
| `<path-to-flags.goff.yaml>` | Path of the flag file in the repository | The repository tree |
| `<github-token>` | Managed secret holding a read-only token | A fine-grained GitHub token with Contents: read |
| `<slack-webhook-url>` | Managed secret holding the incoming-webhook URL | Slack app settings |
| `<relay-evaluation-key>` | Managed secret holding the key SDKs present | Create it in your organization's secrets |

## Related Presets

- **Flag File Relay** - flags typed in a KubernetesGoFeatureFlagFlagFile instead of a repository.
- **Team Flag Sets** - isolated flag sets per team.
