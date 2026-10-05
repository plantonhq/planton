# Kubernetes Karapace

## When NOT to Use This

**TLS serving and multiple replicas do not mix.** With more than one
replica, followers forward writes to the leader at its advertised POD
IP — and a certificate issued for a DNS name does not cover pod IPs.
Pair `server_tls` with `replicas: 1`, or run multiple plain-HTTP
replicas behind TLS terminated at an Ingress/Gateway (the spec
carries this caveat on the field).

Also not the right kind when:

- **You need Confluent Schema Registry itself** — Karapace is Aiven's
  Apache-2.0 registry, API-compatible with Confluent SR (existing SR
  clients work unchanged); if you need Confluent-proprietary
  extensions beyond that API surface, this kind does not provide
  them.
- **You want a database-backed registry** — there is no database:
  schemas live in a compacted Kafka topic (`_schemas` by convention)
  on the connected cluster, the exact architecture Confluent SR uses.
  No Kafka cluster, no registry.
- **You want external exposure baked in** — the registry is
  in-cluster plumbing at the exported `endpoint`; compose a
  first-class exposure kind against the exported service name.
- **You expect a Helm chart or operator underneath** — Karapace has
  neither upstream; the module OWNS the deployment manifests
  (Deployments, Services, Secret wiring) and types every meaningful
  configuration surface. Knobs beyond the spec (statsd/telemetry,
  name-strategy tuning) are not reachable through an escape hatch —
  see the research doc.

## Overview

**KubernetesKarapace** declares a Karapace schema registry — the
Apache-2.0, Confluent-API-compatible schema registry from Aiven.
Producers and consumers register and fetch Avro, JSON Schema and
Protobuf schemas through the standard Schema Registry REST API, and
the registry enforces compatibility between schema versions so a
producer cannot silently break its consumers.

**Storage is Kafka-native**: schemas live in a compacted Kafka topic
on the connected cluster (`registry.topic_name`, default `_schemas`;
created by the registry on first start). Multiple replicas coordinate
leadership through a consumer group; followers forward writes to the
leader — replicas are an availability measure, not write scaling.

**The optional REST-proxy role** (`rest_proxy`) deploys the same
engine as a second, independently-sized Deployment
(`<name>-rest`) serving Kafka's REST proxy API — produce/consume/admin
over HTTP, wired to this registry for schema-aware payloads.

**Key design points:**

- **Module-owned manifests** — a Deployment and Service per role,
  configured entirely through `KARAPACE_*` environment variables on
  upstream's own container image (`ghcr.io/aiven-open/karapace`,
  pinned 6.2.1; `image` overrides for air-gapped registries).
- **Per-pod advertised identity** — each replica advertises its POD
  IP (via the downward API), never a shared Service name: the leader
  publishes its advertised address through the consumer group and
  followers forward writes to it — a shared name would make followers
  forward to themselves.
- **The Kafka connection is typed end-to-end** — `security_protocol`
  (PLAINTEXT / SSL / SASL_PLAINTEXT / SASL_SSL) with spec-enforced
  pairing: SSL forms require the `tls` block, SASL forms require
  `sasl`, and `sasl` requires the protocol EXPLICITLY set to a SASL_*
  value (the default PLAINTEXT would silently ignore credentials).
- **Passwords never ride the pod spec** — a referenced
  `password_secret` wires straight into a secretKeyRef; a literal
  `password` is materialized by the module into the `<name>-sasl`
  Secret first (exactly one of the two, spec-enforced).
- **HTTP-layer authentication is opt-in** — `basic` (a Karapace
  authfile JSON mounted from a Secret, hot-reloaded on change) XOR
  `oidc` (JWT validation against your IdP's HTTPS JWKS endpoint).
  Omitted = the API is open to anyone who can reach the Service.
- **Compatibility is the registry's job** — `registry.compatibility`
  sets the default mode for new subjects (BACKWARD default; the
  _TRANSITIVE variants check ALL prior versions); per-subject
  overrides ride the standard SR config API.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace for the registry — literal or a
  KubernetesNamespace reference
- **`spec.kafka.bootstrap_servers`**: the Kafka cluster that stores
  the schemas — literal `host:port` or a KubernetesKafka reference

### Common

- **`spec.replicas`**: registry replicas (default 1; 2 is the
  production norm — leader election + follower forwarding)
- **`spec.kafka.security_protocol` / `tls` / `sasl`**: the connection
  posture — reference the KubernetesKafka cluster CA and a
  KubernetesKafkaUser credential Secret for Strimzi-managed clusters
- **`spec.registry.topic_name`**: the schemas topic (default
  `_schemas` — the convention existing tooling expects)
- **`spec.registry.replication_factor`**: schemas-topic RF AT
  CREATION — upstream default 1 is a data-loss risk in production;
  set 3 on multi-broker clusters (changing it later means Kafka
  topic reassignment, not editing this field)
- **`spec.registry.compatibility`**: default subject compatibility
  (BACKWARD default)
- **`spec.registry.group_id`**: leader-election consumer group
  (default `metadata.name`) — unique per registry installation
  sharing a Kafka cluster
- **`spec.rest_proxy`**: enable + replicas/port/resources for the
  REST-proxy role
- **`spec.server_tls`**: serve the registry API over TLS from a
  certificate Secret (cert-manager seam) — replicas: 1 caveat above
- **`spec.http_authentication`**: `basic` XOR `oidc`
- **`spec.port` / `spec.log_level` / `spec.image` /
  `spec.resources` / `spec.node_selector` / `spec.tolerations`**

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the registry runs in |
| `service_name` | The registry Service (`<name>`) |
| `endpoint` | In-cluster registry endpoint (`http(s)://<name>.<namespace>.svc.cluster.local:<port>`) — the `schema.registry.url` value for producers, consumers, Connect converters and consoles |
| `rest_proxy_endpoint` | In-cluster REST-proxy endpoint (`<name>-rest`); empty when the role is not enabled |
| `schemas_topic` | The Kafka topic storing the schemas |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace); **`kafka.bootstrap_servers`** references a
  KubernetesKafka (`status.outputs.internal_bootstrap_endpoint`);
  **`kafka.tls.ca_secret_name`** references a KubernetesKafka
  (`status.outputs.cluster_ca_cert_secret_name`);
  **`kafka.tls.client_cert_secret_name`** and
  **`kafka.sasl.password_secret.secret_name`** reference a
  KubernetesKafkaUser (`status.outputs.secret_name`);
  **`server_tls.secret_name`** references a KubernetesCertificate
  (`status.outputs.secret_name`) — the cert-manager seam.
- **Applications consume `endpoint`** as `schema.registry.url`; a
  KubernetesKafkaUi console wires the same output as a cluster's
  `schema_registry.url`.
- **Confluent-SR clients work unchanged** — Karapace is API-
  compatible, which is what makes it the registry half of a
  Confluent exit (MirrorMaker 2 moves the topics; re-register
  subjects here).

## Examples

### Minimal (dev)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarapace
metadata:
  name: dev-registry
spec:
  namespace:
    value: dev-kafka
  kafka:
    bootstrap_servers:
      value: dev-kafka-kafka-bootstrap.dev-kafka.svc.cluster.local:9092
```

### Production (2 replicas, SASL_SSL, RF 3)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarapace
metadata:
  name: schema-registry
spec:
  namespace:
    value: kafka
  replicas: 2
  kafka:
    bootstrap_servers:
      value: events-kafka-bootstrap.kafka.svc.cluster.local:9094
    security_protocol: SASL_SSL
    tls:
      ca_secret_name:
        value: events-cluster-ca-cert
    sasl:
      mechanism: SCRAM-SHA-512
      username: karapace
      password_secret:
        secret_name:
          value: karapace
  registry:
    replication_factor: 3
    compatibility: BACKWARD
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: "1"
      memory: 512Mi
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
