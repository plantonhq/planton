# KubernetesKafka Terraform Module

Deploys one Strimzi-operator-managed KRaft Kafka cluster: the optional
namespace, the optional JMX metrics rules ConfigMap, one
`kafka.strimzi.io/v1` `KafkaNodePool` per `node_pools` entry, and the
`kafka.strimzi.io/v1` `Kafka` CR itself. The typed spec renders into
the CR bodies in `locals.tf` — the exact twin of the Pulumi module's
`kafka.go` / `nodepools.go` (same keys rendered and omitted, numbers
as numbers, booleans as booleans).

## Module Behavior

- **The cluster name is `metadata.name`** — the Kafka CR name and the
  value of the `strimzi.io/cluster` label binding every KafkaNodePool
  (and the KafkaTopic/KafkaUser resources declared through their own
  kinds) to this cluster.
- **CRs apply through `kubectl_manifest` (alekc/kubectl)** — unlike
  the hashicorp provider's `kubernetes_manifest` it needs no cluster
  connection at plan time, so the cluster can be PLANNED before the
  Strimzi operator's CRDs exist — an Infra Chart can deploy the
  operator and its Kafka clusters in one run.
- **Node pools apply BEFORE the Kafka CR** — Strimzi tolerates either
  order, but a Kafka CR with no matching pools reports a transient
  warning state. Pools are keyed by their OWN NAME (never a positional
  index), so state addresses survive pool-list reorderings.
- **No wait_for block, deliberately** — cluster readiness depends on
  the operator (image pulls, KRaft quorum formation, listener
  provisioning) that is not part of applying the resources; the
  never-block-on-a-controller posture of every operator-CR kind.
- **The module (not the operator) owns the metrics ConfigMap** —
  `metrics.enabled` renders the canonical Strimzi JMX Prometheus rules
  as `<name>-kafka-metrics` (`kubernetes_config_map_v1`); the Kafka
  CR's `metricsConfig` only points at it.
- **The module owns namespace creation** — `create_namespace` drives a
  `kubernetes_namespace_v1` resource carrying the standard governance
  labels.

## Rendering Quirks

- **Every listener carries the CRD-required quartet**
  (name/port/type/tls, type coalesced to `internal`); authentication
  and configuration render only when declared. Custom-authentication
  knobs (`sasl`, `listenerConfig`) render only on the custom arm, and
  custom-authorizer knobs (`authorizerClass`, `supportsAdminApi`) only
  on the custom authorization arm.
- **Storage is one literal with per-arm gated keys** — ephemeral
  renders type only; persistent-claim renders size/class/deleteClaim;
  jbod renders per-volume bodies (each persistent-claim with its own
  size/class/deleteClaim, at most one carrying
  `kraftMetadata: shared`).
- **`node_selector` translates to node affinity** — the Strimzi pod
  template carries affinity and tolerations but NO nodeSelector; the
  module renders a required node affinity with one matchExpressions
  entry per label, sorted by key for determinism (the Pulumi module
  renders the same translation).
- **The entity operator omits cleanly** — each sub-operator renders
  when enabled (spec defaults both true); with BOTH disabled the block
  is omitted entirely and KafkaTopic/KafkaUser declarations for this
  cluster become inert.
- **cert-manager key fallbacks** — `broker_cert_chain_and_key` renders
  `certificate`/`key` coalesced to `tls.crt`/`tls.key`, the names
  cert-manager writes.
- **Null-prune idiom throughout** — conditional entries are written as
  `key = cond ? value : null` inside one object literal and pruned, so
  numbers and booleans keep their types in the rendered CR (server-side
  apply rejects strings where the CRD wants numbers).

## Resources

| Resource | Condition |
|---|---|
| `kubernetes_namespace_v1.namespace` | `spec.create_namespace` |
| `kubernetes_config_map_v1.kafka_metrics` | `spec.metrics.enabled` |
| `kubectl_manifest.node_pools` (for_each by pool name) | always |
| `kubectl_manifest.kafka` | always |

## Usage

```bash
planton tofu apply --manifest kafka.yaml
```

## Local Development

```bash
terraform init
terraform validate
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Inputs

See `variables.tf` for the full variable specification (generated from
the spec proto). The spec arrives from the proto→tfvars converter in
snake_case with the `namespace` foreign key (KubernetesNamespace) and
the storage-class / certificate-secret references resolved to literal
strings before Terraform runs.

## Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Namespace the cluster deploys into |
| `cluster_name` | Kafka cluster name (`metadata.name`) — the `strimzi.io/cluster` binding value |
| `bootstrap_service_name` | Internal bootstrap Service (`<cluster>-kafka-bootstrap`) |
| `internal_bootstrap_endpoint` | In-cluster bootstrap address for the first internal listener (empty when none is declared) |
| `cluster_ca_cert_secret_name` | Cluster CA Secret (`<cluster>-cluster-ca-cert`, key `ca.crt`) |

## Parity

Kept in lockstep with the Pulumi module (`../pulumi/module/`): same
`kafka.strimzi.io/v1` apiVersion, same CR bodies (listener quartet,
storage arms, affinity translation, entity-operator omission), same
module-owned metrics ConfigMap and naming contracts, same
pools-before-cluster ordering, same outputs.
