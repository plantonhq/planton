# Minimal

Stigmer inside the cluster, everything bundled: the operator runs Postgres (through CloudNativePG, which it installs) and Temporal for you, files live on a volume the server and the shared runner share, and there is no front door. You reach it with the `port_forward_command` output, and every caller acts as the owner, which is safe while the install is reachable only from inside the cluster.

## When to Use

- A first install, an evaluation, or a team that reaches Stigmer from inside the cluster
- The 30-second choice: one model key, one owner, nothing else to decide

## Key Configuration Choices

- **`operator.create_namespace: true`** — the operator gets `stigmer-system`, created and owned by this resource
- **`namespace`** — the Stigmer's own namespace must exist; it holds the encryption keys, so this resource never creates or deletes it
- **`runner.model.keys.anthropic_api_key`** — a Planton secret reference; this resource writes it into a Secret the runner reads
- **No `database`, `temporal` or `object_store`** — the operator runs them, and the database has no backups until you add a bucket

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `version` | The Stigmer release | Stigmer's releases page; the chart and the server share it |
| `owner.email` | The person who runs the install | Your team |
| `$secret/anthropic-api-key` | Your model key | A Planton secret holding your Anthropic API key |

## Related Presets

- **02-bring-your-own-database** — your own Postgres and Temporal
- **03-bucket-and-backups** — files and backups in a bucket
- **04-exposed-with-sign-in** — a hostname with the bundled sign-in
