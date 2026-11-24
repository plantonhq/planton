# Sized for One Team

The minimal install with every size written out: the operator, the server, the shared runner, Postgres, Temporal's four services and every volume, at the sizes Stigmer's operator applies for 0.3.2 when they are left out. Nothing here changes what runs; it shows what one install reserves from the cluster and gives you each number to change.

## When to Use

- You need to know what Stigmer will reserve before you deploy it: the capacity estimate counts every size declared here
- You are sizing for more than one team, or for a smaller cluster, and want to start from the numbers the operator would pick

## Key Configuration Choices

- **Sized for one team** — the operator's own defaults: the server requests 250m CPU and 1Gi of memory with a 2Gi limit, the runner 500m and 1Gi with 4Gi, Postgres and each Temporal service 100m and 256Mi with 1Gi
- **CPU is requested, never limited** — a burst (a search-index rebuild at boot) is not throttled; memory is requested and limited, so a leak restarts a pod instead of starving a node
- **Memory floors** — the operator refuses a size it cannot boot with: below 512Mi for the server and the runner, 128Mi for Postgres and Temporal, or a request above its own limit
- **`temporal.bundled.resources`** — one size for each of Temporal's four services; the estimate counts it four times
- **Volumes** — 5Gi for the server, 20Gi for artifacts (only without a bucket), 10Gi for the runner and 10Gi for each Postgres instance, on the cluster's default storage class
- **A size you leave out** of a block (say, a memory limit) is filled from the operator's default for that version, and the estimate does not count it

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `version` | The Stigmer release | Stigmer's releases page; the chart and the server share it |
| `owner.email` | The person who runs the install | Your team |
| `$secret/anthropic-api-key` | Your model key | A Planton secret holding your Anthropic API key |

## Related Presets

- **01-minimal** — the same install, with the operator choosing every size
- **03-bucket-and-backups** — files and backups in a bucket, and a standby instance
- **05-per-session-sandboxes** — each session in a sandbox of its own, with no shared runner to size
