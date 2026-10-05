# Karapace Schema Registry

Deploys a Karapace schema registry — the Apache-2.0, Confluent Schema Registry API-compatible engine from Aiven. Producers and consumers register and fetch Avro, JSON Schema, and Protobuf schemas through the standard REST API; existing Confluent SR client libraries, Connect converters, and tooling work unchanged.

Schemas live in a compacted Kafka topic (`_schemas` by convention) on the connected cluster — the same Kafka-native architecture Confluent SR uses. Multiple replicas coordinate leadership through a consumer group; followers forward writes to the leader.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Deployment + Service** — the registry engine (Karapace has no official Helm chart; the module owns the manifests)
- **Optional second Deployment** — the Kafka REST-proxy role (`<name>-rest`) when `restProxy.enabled` is true
- **Schemas topic** — created on first start at the replication factor you declare (applies at creation only)

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- A **Kafka cluster** reachable at the bootstrap address you declare — reference an **Apache Kafka** resource for in-cluster wiring.
- A **listener whose security posture matches** your declared `securityProtocol` — SSL forms need TLS trust material; SASL forms need credentials. The SASL password is exactly one of a Secret reference (recommended — reference a **Kafka User**) or a declared value the module materializes into a Secret.

## Deploy

### Console

Open the deployment store, find **Karapace Schema Registry**, and click **Deploy**. The creation wizard walks you through namespace placement, the Kafka connection (with the SASL password XOR), replica leadership, registry behavior, optional REST proxy, server TLS, HTTP authentication, runtime, and pod sizing. Start from the **Minimal preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarapace
metadata:
  name: dev-registry
  org: acme-corp
  env: dev
spec:
  namespace:
    value: dev-registry
  createNamespace: true
  kafka:
    bootstrapServers:
      value: dev-kafka-kafka-bootstrap.dev-kafka.svc.cluster.local:9092
```

```shell
planton apply -f karapace-registry.yaml
```

This creates a single-replica registry storing schemas in the connected Kafka cluster's `_schemas` topic, reachable in-cluster at the exported endpoint. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the bootstrap endpoint from the Kafka cluster itself:

```yaml
spec:
  namespace:
    value: kafka
  kafka:
    bootstrapServers:
      valueFrom:
        kind: KubernetesKafka
        name: event-bus
        fieldPath: status.outputs.internal_bootstrap_endpoint
```

The InfraPipeline deploys the Kafka cluster first, then provisions the registry against its resolved endpoint.

## Key Configuration

These are the most important decisions when configuring a Karapace registry. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Namespace placement** — unlike a KafkaTopic, the registry has no same-namespace contract with the Kafka cluster; it connects over the network. The namespace locks after creation (moving a Deployment is a delete-and-recreate).

**Replication factor is a one-shot door** — it applies when the registry creates the `_schemas` topic. Start production registries at RF 3; graduating later means Kafka topic reassignment, not a field edit.

**Server TLS + multiple replicas** — followers forward writes to the leader at its pod IP, which DNS-name certificates do not cover. Pair server TLS with one replica, or terminate TLS at an Ingress/Gateway in front of plain-HTTP replicas.

**HTTP authentication** — absent means the API is open to anyone who can reach the Service. Configure Basic (authfile Secret) or OIDC before shared exposure.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesKafka** | `kafka.bootstrapServers` | `status.outputs.internal_bootstrap_endpoint` |
| **KubernetesKafka** | `kafka.tls.caSecretName` | `status.outputs.cluster_ca_cert_secret_name` |
| **KubernetesKafkaUser** | `kafka.sasl.passwordSecret.secretName` | `status.outputs.secret_name` |
| **KubernetesCertificate** | `serverTls.secretName` | `status.outputs.secret_name` |
| **KubernetesSecret** | `httpAuthentication.basic.secretName` | `status.outputs.secret_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `endpoint` | In-cluster registry URL — the `schema.registry.url` for clients | Producers, consumers, Connect converters, and the Kafka UI console |
| `rest_proxy_endpoint` | REST-proxy URL (empty when the role is disabled) | HTTP produce/consume clients |
| `schemas_topic` | The Kafka topic storing schemas | Monitoring, backup planning |
| `service_name` | Name of the registry Service | Ingress/Gateway backend target |
| `namespace` | Namespace the registry runs in | Composed route placement context |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Minimal** — one replica, plaintext bootstrap, everything else on defaults. Start from the **Minimal preset**.

**Production HA** — two replicas, SASL_SSL, schemas topic at RF 3. Start from the **Production HA preset**.

**REST proxy and TLS** — single replica with server TLS, HTTP Basic auth, and the REST-proxy role enabled. Start from the **REST proxy and TLS preset**.

## Works With

- [**Apache Kafka**](/infra-catalog/kubernetes-kafka) — the cluster storing schemas; wire `bootstrapServers` from its outputs
- [**Kafka User**](/infra-catalog/kubernetes-kafka-user) — operator-generated SASL credentials via `passwordSecret`
- [**Cert Manager Certificate**](/infra-catalog/kubernetes-certificate) — cert-manager-issued TLS for `serverTls`
- [**Kafka Connect**](/infra-catalog/kubernetes-kafka-connect) — Connect converters point `schema.registry.url` at the `endpoint` output
- [**Kafka UI**](/infra-catalog/kubernetes-kafka-ui) — schema-aware browsing wired through the same endpoint
