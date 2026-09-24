# KubernetesNeo4j Terraform Module

Terraform/OpenTofu module for the KubernetesNeo4j component: installs
Neo4j — the graph database behind knowledge graphs, GraphRAG and
agent-memory architectures — from the official Neo4j Helm chart (`neo4j`
at https://helm.neo4j.com/neo4j).

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
- **The typed spec renders into chart values** (`locals.helm_values`), and
  the spec's `helm_values` escape hatch is passed as a SECOND values
  document that the provider merges over the first with Helm `-f`
  semantics — the exact semantic twin of the Pulumi module.

### The auth Secret contract (created BEFORE the release)

The chart's credential interface is `neo4j.passwordFromSecret`: a Secret
carrying key `NEO4J_AUTH` with value `neo4j/<password>`, which the chart
LOOKS UP AT TEMPLATE TIME — the install fails if the Secret is missing or
lacks the key. The module therefore wires an explicit dependency so the
Secret always exists first:

- auth absent (the default) → `random_password.admin` generates the
  password (24 letters and digits; the generation shape is
  `ignore_changes` so an imported credential never regenerates), and the
  module materializes `kubernetes_secret_v1.auth` named
  `<metadata.name>-auth` (key `NEO4J_AUTH` = `neo4j/<password>` plus a
  bare `password` key, both wrapped in `sensitive()`) before
  `helm_release.neo4j`, rendering `passwordFromSecret` = that name.
- `auth.password` arm → the same Secret, carrying the declared value
  instead of a generated one. The password itself NEVER appears in
  rendered chart values on either arm.
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

| Spec field | Chart value |
|---|---|
| `cluster_name` (else `metadata.name`) | `neo4j.name` |
| `edition` | `neo4j.edition` |
| `accept_license_agreement` | `neo4j.acceptLicenseAgreement: "yes"` (the chart's string shape; rendered only when true) |
| `auth.*` | `neo4j.passwordFromSecret` (see the contract above) |
| `resources.requests` | `neo4j.resources.{cpu,memory}` (the chart's flat shape, applied to requests) |
| `resources.limits` | `neo4j.resources.limits.{cpu,memory}` (the chart's full-format limits) |
| `data_volume.storage_class` set | `volumes.data.mode: dynamic` + `volumes.data.dynamic.{storageClassName, accessModes [ReadWriteOnce], requests.storage}` |
| `data_volume.storage_class` empty | `volumes.data.mode: defaultStorageClass` + `volumes.data.defaultStorageClass.{accessModes, requests.storage}` |
| `memory.heap_initial/heap_max/page_cache` | `config."server.memory.heap.initial_size"/"server.memory.heap.max_size"/"server.memory.pagecache.size"` — merged over `spec.config`, TYPED KEYS WIN on collision |
| `config` | `config` (free-form neo4j.conf entries) |
| `apoc_config` | `apoc_config` |
| `additional_jvm_arguments` | `jvm.additionalJvmArguments` |
| `use_default_jvm_arguments` | `jvm.useNeo4jDefaultJvmArguments` (rendered only when explicitly declared) |
| `service.type` (default ClusterIP) | `services.neo4j.spec.type` |
| `service.annotations` | `services.neo4j.annotations` |
| `ssl.bolt/https.secret` | `ssl.<scope>.privateKey.secretName` AND `ssl.<scope>.publicCertificate.secretName` (one Secret, both roles) |
| `scheduling.node_selector` | `nodeSelector` (top-level chart key) |
| `scheduling.tolerations` | `podSpec.tolerations` |
| `scheduling.pod_anti_affinity` | `podSpec.podAntiAffinity` (rendered only when explicitly declared) |
| `scheduling.priority_class_name` | `podSpec.priorityClassName` |
| `service_monitor_enabled` | `serviceMonitor.enabled` |
| `image.registry/repository/tag` | `image.{registry,repository,tag}` — repository resolves to `neo4j` when empty (the chart fails on separated fields without a repository) |
| `helm_values` | merged LAST over everything above |

The chart REJECTS resources below its floor (500m CPU / 2Gi memory); the
module never defaults below it — when `spec.resources` is empty nothing
renders and the chart's own defaults (1000m/2Gi) apply. The install waits
for the workload to become Ready (`wait`/`atomic`/`cleanup_on_fail`, 600s
budget — Neo4j recovers/upgrades store files on startup).

## Resources

| Resource | Condition |
|---|---|
| `kubernetes_namespace_v1.neo4j` | `spec.create_namespace` |
| `random_password.admin` | `spec.auth` absent |
| `kubernetes_secret_v1.auth` | every arm but `spec.auth.existing_secret` |
| `helm_release.neo4j` | always |

## Outputs

| Output | Meaning |
|---|---|
| `namespace` | Namespace the server runs in |
| `release_name` | Helm release name (= `metadata.name`) |
| `service_name` | The main Neo4j Service (= the release name) |
| `bolt_endpoint` | `neo4j://<svc>.<ns>.svc.cluster.local:7687` |
| `http_endpoint` | `http://<svc>.<ns>.svc.cluster.local:7474` |
| `auth_secret_name` | `<name>-auth` (declared or generated password), or the existing Secret name; always set |
| `port_forward_command` | kubectl one-liner for reaching bolt from a workstation |
| `password_secret` | `{name, key}` of the bare password in `<name>-auth` (key `password`); unset for the `existing_secret` arm |

## Parity

Kept in lockstep with the Pulumi module (`iac/pulumi/module/`): same chart
identity, same values rendering (resolved defaults, the always-rendered
`neo4j.name` and data volume, the ClusterIP override), same auth Secret
name, key, and contents, same outputs. Conditional objects use the
null-prune idiom throughout so numbers and booleans keep their types in
the rendered values.
