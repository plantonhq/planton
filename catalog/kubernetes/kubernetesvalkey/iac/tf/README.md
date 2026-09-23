# KubernetesValkey Terraform Module

Terraform/OpenTofu module for the KubernetesValkey component: installs
Valkey — the Redis-compatible in-memory data store — from the official
Valkey Helm chart (`valkey` at https://valkey.io/valkey-helm/), in either
standalone or primary/replica topology.

## Module Behavior

- **One Helm release named after `metadata.name`** — several Valkey
  instances coexist in one cluster, so nothing here is a fixed chart name.
  The chart's fullname is pinned to the release name too, so the rendered
  Services carry deterministic, manifest-derived names: `<name>` (write),
  `<name>-headless` (pod discovery, replication mode), `<name>-read`
  (reads, replication mode), `<name>-metrics` (exporter, when metrics are
  enabled).
- **The typed spec renders into chart values** (`locals.helm_values`), and
  the spec's `helm_values` escape hatch is passed as a SECOND values
  document that the provider merges over the first with Helm `-f`
  semantics — the exact semantic twin of the Pulumi module.
- **ACL passwords materialize as the `<name>-auth` Secret** (one key per
  username), which the chart consumes via `auth.usersExistingSecret` — its
  init script reads each user's password from the Secret key named after
  the user. A user declared without a password gets one from
  `random_password.user` (keyed by username, so a spec reorder never swaps
  credentials; the generation shape is `ignore_changes` so an imported
  credential never regenerates); a declared password is used as given.
  The rendered `aclUsers` carry permissions only, never passwords, so
  credentials never appear in chart values.
- **The module owns `valkey.conf` rendering**: the typed `config` block
  becomes the chart's single `valkeyConfig` string, deterministically
  ordered (`appendonly`, `save` points or the disable directive,
  `maxmemory`, `maxmemory-policy`, then `extra_directives` verbatim) and
  byte-identical across both engines.
- **TLS pins the chart's Secret key names to the kubernetes.io/tls layout**
  (`tls.crt`/`tls.key`/`ca.crt`) — the chart's own defaults
  (`server.crt`/`server.key`) predate that convention, and the spec's
  certificate seam is cert-manager, which emits kubernetes.io/tls Secrets.
- **The PodDisruptionBudget renders only in replication mode** — the
  chart's PDB template is gated on `replica.enabled`, so a standalone
  declaration would be a silent no-op in the release; the module omits it
  instead of rendering dead values.
- **The install waits for the workload to become Ready**
  (`wait`/`atomic`/`cleanup_on_fail`, 600s budget — replication starts
  pods one at a time and each replica full-syncs before Ready).

## Resources

| Resource | Condition |
|---|---|
| `kubernetes_namespace_v1.valkey` | `spec.create_namespace` |
| `random_password.user` | one per ACL user declared without a password |
| `kubernetes_secret_v1.auth` | `spec.auth` declared |
| `helm_release.valkey` | always |

## Outputs

| Output | Meaning |
|---|---|
| `namespace` | Namespace the instance runs in |
| `service` | Write Service name (= `metadata.name`) |
| `read_service` | Read Service name — replication mode with the read service enabled, empty otherwise |
| `headless_service` | Headless Service name — replication mode only, empty standalone (the chart renders no headless Service for the standalone Deployment) |
| `kube_endpoint` | In-cluster endpoint of the write Service |
| `port_forward_command` | kubectl one-liner for workstation access |
| `username` | `default` when auth is declared, empty otherwise |
| `password_secret` | `{name, key}` handle into the `<name>-auth` Secret (key = `default`), the same whether that password was declared or generated; unset when auth is off |

## Parity

Kept in lockstep with the Pulumi module (`iac/pulumi/module/`): same chart
identity, same values rendering (byte-identical `valkeyConfig` line order,
resolved defaults, and TLS key pinning), same auth Secret name, keys, and
contents, same outputs. Conditional objects use the null-prune idiom
throughout so numbers and booleans keep their types in the rendered
values.
