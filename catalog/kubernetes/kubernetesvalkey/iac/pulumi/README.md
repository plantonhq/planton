# KubernetesValkey Pulumi Module

Pulumi (Go) module for the KubernetesValkey component: installs Valkey —
the Redis-compatible in-memory store — from the official Helm chart as a
real Helm release (`helm/v3.Release`), one release named after
`metadata.name` so several instances coexist in one cluster.

## Module Behavior

- **One Helm release named after `metadata.name`**, with the chart's
  `fullnameOverride` pinned to the same value so every derived Service and
  StatefulSet (`<name>`, `<name>-headless`, and in replication mode
  `<name>-read`) is deterministic.
- **The typed spec renders into chart values** (`values.go`), and the
  spec's `helm_values` escape hatch deep-merges over them with Helm `-f`
  semantics (`helpers.go` mergeMaps) — the exact semantic twin of the
  Terraform module's two-document `values` list.
- **ACL passwords materialize as the `<name>-auth` Secret** (`secrets.go`,
  one key per username), which the chart consumes via
  `auth.usersExistingSecret` — its init script reads each user's password
  from the Secret key named after the user. A user declared without a
  password gets one from a `random.RandomPassword` whose logical name is
  keyed by username (so a spec reorder never swaps credentials; the
  generation shape is `IgnoreChanges` so an imported credential never
  regenerates); a declared password is used as given. The rendered
  `aclUsers` carry permissions only, never passwords, so credentials never
  appear in chart values.
- **The module owns `valkey.conf` rendering**: the typed `config` block
  becomes the chart's single `valkeyConfig` string, deterministically
  ordered and byte-identical with the Terraform module's.
- **TLS pins the chart's Secret key names to the kubernetes.io/tls layout**
  (`tls.crt`/`tls.key`/`ca.crt`), the layout cert-manager emits.
- **The PodDisruptionBudget renders only in replication mode** — the
  chart's PDB template is gated on `replica.enabled`.
- **The install waits for the workload to become Ready** (`SkipAwait:
  false`, `Atomic`, `CleanupOnFail`, 600s budget — the Terraform twin's
  `wait`/`atomic`/`cleanup_on_fail`).

## Resources

| Resource | Condition |
|---|---|
| `core/v1.Namespace` | `spec.create_namespace` |
| `random.RandomPassword` (`auth-password-<username>`) | one per ACL user declared without a password |
| `core/v1.Secret` (`<name>-auth`) | `spec.auth` declared |
| `helm/v3.Release` | always |

## Outputs

| Output | Meaning |
|---|---|
| `namespace` | Namespace the instance runs in |
| `service` | Write Service name (= `metadata.name`) |
| `read_service` | Read Service name — replication mode with the read service enabled, empty otherwise |
| `headless_service` | Headless Service name — replication mode only, empty standalone |
| `kube_endpoint` | In-cluster endpoint of the write Service |
| `port_forward_command` | kubectl one-liner for workstation access |
| `username` | `default` when auth is declared, empty otherwise |
| `password_secret` | `{name, key}` handle into the `<name>-auth` Secret (key = `default`), the same whether that password was declared or generated; unset when auth is off |

## Parity

Kept in lockstep with the Terraform module (`iac/tf/`): same chart
identity, same values rendering, same auth Secret name, keys, and
contents (including which users get a generated password), same outputs.
A behavior added to one engine lands in the other in the same change.
