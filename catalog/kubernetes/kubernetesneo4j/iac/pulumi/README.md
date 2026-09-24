# KubernetesNeo4j Pulumi Module

Pulumi (Go) module for the KubernetesNeo4j component: installs Neo4j — the
graph database behind knowledge graphs, GraphRAG and agent-memory
architectures — from the official Neo4j Helm chart (`neo4j` at
https://helm.neo4j.com/neo4j) as a real Helm release
(`helm/v3.Release`).

## Module Behavior

- **One Helm release named after `metadata.name`** — several Neo4j servers
  coexist in one cluster, and Enterprise cluster members are each their own
  release. The chart names its always-created ClusterIP Service after the
  release (`neo4j.fullname` = the release name when no name overrides are
  set), so the exported `service_name` is deterministic.
- **`neo4j.name` always renders** — the chart REQUIRES it (its `neo4j.name`
  helper fails the install when empty; nothing defaults it to the release
  name): `cluster_name` when set (Enterprise members sharing it form one
  cluster), else `metadata.name`.
- **The typed spec renders into chart values** (`values.go`), and the
  spec's `helm_values` escape hatch deep-merges over them with Helm `-f`
  semantics (`helpers.go` mergeMaps) — the exact semantic twin of the
  Terraform module's two-document `values` list.

### The auth Secret contract (created BEFORE the release)

The chart's credential interface is `neo4j.passwordFromSecret`: a Secret
carrying key `NEO4J_AUTH` with value `neo4j/<password>`, which the chart
LOOKS UP AT TEMPLATE TIME — the install fails if the Secret is missing or
lacks the key. The module therefore wires an explicit `DependsOn` so the
Secret always exists first:

- auth absent (the default) → a `random.RandomPassword` generates the
  password (24 letters and digits; the generation shape is `IgnoreChanges`
  so an imported credential never regenerates), and the module
  materializes the `<metadata.name>-auth` Secret (`secrets.go`; key
  `NEO4J_AUTH` = `neo4j/<password>` plus a bare `password` key, both
  secret in state) before the Helm release, rendering `passwordFromSecret`
  = that name.
- `auth.password` arm → the same Secret, carrying the declared value
  (wrapped with `pulumi.ToSecret`) instead of a generated one. The
  password itself NEVER appears in rendered chart values on either arm.
- `auth.existing_secret` arm → `passwordFromSecret` = the given name; no
  Secret is created (it must already exist and carry the contract), and
  `password_secret` is unset because that Secret's layout is the owner's.

### The ClusterIP override (deliberate)

The chart ships `services.neo4j.spec.type: LoadBalancer`, which would
provision a cloud load balancer (or hang Pending) on every install. This
module pins it to **ClusterIP** unless `spec.service.type` says otherwise —
exposure composes from first-class kinds (KubernetesIngress, Gateway API
kinds) over the exported service handle instead. Note the chart names this
extra service `<neo4j.name>-lb-neo4j`; the always-created default ClusterIP
Service (= the release name) is the composition handle.

### The SSL key-name bridge (cert-manager Secrets)

The chart mounts `private.key` and `public.crt` from each `ssl.<scope>`
Secret (its `subPath` defaults). cert-manager Certificates store their
material as `tls.key`/`tls.crt` — the module does NOT silently rewrite key
names, so a cert-manager Secret needs a key bridge before use: either copy
the Secret with renamed keys, or set the chart's
`ssl.<scope>.privateKey.subPath: tls.key` and
`ssl.<scope>.publicCertificate.subPath: tls.crt` via `helm_values`.

## Values Mapping

See the Terraform module README (`../tf/README.md`) for the full
field-by-field table — both engines render byte-identical values. Notable
mechanics: `neo4j.resources` renders requests into the chart's flat
`{cpu, memory}` shape plus declared limits into the full-format `limits`
sub-map; `volumes.data` ALWAYS renders (the chart requires a mode) —
`dynamic` with the declared StorageClass, else `defaultStorageClass`; the
typed `memory` block merges over `spec.config` with TYPED KEYS WINNING on
collision; the chart REJECTS resources below its floor (500m CPU / 2Gi
memory) and the module never defaults below it.

## Outputs

| Output | Meaning |
|---|---|
| `namespace` | Namespace the server runs in |
| `release_name` | Helm release name (= `metadata.name`) |
| `service_name` | The main Neo4j Service (= the release name) |
| `bolt_endpoint` | `neo4j://<svc>.<ns>.svc.cluster.local:7687` |
| `http_endpoint` | `http://<svc>.<ns>.svc.cluster.local:7474` |
| `auth_secret_name` | `<name>-auth` (declared or generated password), or the existing Secret name; always set |
| `password_secret` | `{name, key}` of the bare password in `<name>-auth` (key `password`); unset for the `existing_secret` arm |
| `port_forward_command` | kubectl one-liner for reaching bolt from a workstation |

## Parity

Kept in lockstep with the Terraform module (`../tf/`): same chart identity
(`vars.go` ↔ `locals.tf`), same values rendering, same auth Secret name,
key, and contents, same outputs.
