# Per-session Sandboxes

Each engineer's agent session in a pod of its own, through agent-sandbox, behind a network fence: a sandbox reaches the server, the bucket, DNS and the public internet, and only the private ranges you name. A quiet sandbox sleeps and wakes with its files on the session's next message.

## When to Use

- Agents run commands you would not run side by side on one shared runner
- Stigmer Enterprise, which requires a per-session driver

## Key Configuration Choices

- **`sandboxes.driver: agentSandbox`** — the operator installs agent-sandbox's controller when it is missing (`prerequisites.agent_sandbox: auto`, the default)
- **`object_store`** — required by both per-session drivers; with the bundled database it also turns the archive on, so cert-manager must run on the cluster (deploy a Cert Manager first)
- **`egress_allow_cidrs`** — the only private addresses a sandbox reaches, such as a model gateway or an MCP server on your network; a fence needs a CNI that enforces NetworkPolicy
- **`driver: substrate`** is the other per-session choice: Agent Substrate actors, gVisor sandboxes on a pool of workers, which needs Kubernetes 1.37 with the pod-certificate APIs

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `storage` | Your KubernetesSeaweedFs | The SeaweedFS resource in this environment |
| `egress_allow_cidrs` | Private ranges sandboxes may reach | Your network plan |

## Related Presets

- **03-bucket-and-backups** — the bucket this preset needs
- **07-composed-prerequisites** — prerequisites installed from the catalog
