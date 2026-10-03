# Kafka User

Declares ONE Kafka client identity on a Strimzi-managed cluster. The target cluster's user operator (part of its entity operator, enabled by default on Apache Kafka) reconciles the declaration into a real principal: it GENERATES the credentials into a Kubernetes Secret named after the user — a SCRAM password with a ready `sasl.jaas.config`, or a client certificate issued from the cluster's clients CA — and, when the cluster runs `simple` authorization, applies the declared ACLs. No manifest ever carries secret material, and nothing sensitive is typed into the console. Per-user broker quotas cap how hard the client may push the cluster.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **KafkaUser** -- the Strimzi custom resource, named after this resource, placed in the Kafka cluster's own namespace and bound to the cluster through the `strimzi.io/cluster` label (rendered from `kafkaCluster`)
- **Credentials Secret** (materialized by the cluster's user operator, not the module) -- named after the user: keys `password` + `sasl.jaas.config` for `scram-sha-512` users; `user.crt` / `user.key` plus `user.p12` / `user.password` for `tls` users. NO Secret for `tls-external` users -- their certificates are issued outside the cluster

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- A **Apache Kafka** cluster with its user operator enabled (the default).
- A **listener whose authentication type matches this user's** -- a scram-sha-512 user cannot authenticate on a tls-auth listener, and vice versa.
- **`simple` authorization on the cluster** when this user declares ACLs -- on a cluster without it, the user operator REJECTS ACL-bearing users outright (the KafkaUser reports NotReady and no credentials are generated).

## Deploy

### Console

Open the deployment store, find **Kafka User**, and click **Deploy**. The creation wizard walks you through the owning cluster (with a one-click fill of the cluster's own namespace -- the placement contract), the authentication type, the ACL rules with producer/consumer quick-picks, and the quota dials. Start from the **Producer (SCRAM)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafkaUser
metadata:
  name: order-service
  org: acme-corp
  env: prod
spec:
  namespace:
    value: kafka # MUST be the Kafka cluster's own namespace
  kafkaCluster:
    value: my-kafka
  authentication:
    type: scram-sha-512
  authorization:
    acls:
      - resource:
          type: topic
          name: order-events
        operations:
          - Write
          - Describe
```

```shell
planton apply -f kafka-user.yaml
```

This declares a SCRAM producer identity: the user operator generates the password into a Secret named `order-service` and grants Write + Describe on the `order-events` topic. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire both placement facts from the Kafka cluster itself so the pair can never drift apart:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesKafka
      name: event-bus
      fieldPath: status.outputs.namespace
  kafkaCluster:
    valueFrom:
      kind: KubernetesKafka
      name: event-bus
      fieldPath: status.outputs.cluster_name
```

The InfraPipeline deploys the cluster first, then declares the user against it.

## Key Configuration

These are the most important decisions when declaring a Kafka user. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The placement contract** -- the KafkaUser must live in the Kafka cluster's OWN namespace: the user operator watches only there, and a user declared anywhere else is accepted by the API server and then silently never reconciled -- no principal, no credentials, no error. `kafkaCluster` renders as the `strimzi.io/cluster` label that binds the user to its cluster.

**Authentication must match a listener** -- `scram-sha-512` (username/password; the operator mints the password), `tls` (the operator issues a client certificate from the cluster's clients CA; the principal becomes `CN=<name>`), or `tls-external` (certificates issued outside the cluster; no Secret is generated). Omit the block entirely for an ACL-only principal carried by a custom listener mechanism.

**ACLs only ever allow** -- Kafka's simple authorizer denies everything not granted. A producer needs `Write` + `Describe` on its topic; a consumer needs `Read` + `Describe` on the topic AND `Read` on its consumer group (forgetting the group half is the classic "authorized but cannot consume" failure). `pattern_type: prefix` turns a name into a family rule (`orders-` covers every current and future `orders-*` topic). Cluster-type rules are nameless. An empty `host` means any host (`*`).

**Quotas are ceilings, not guarantees** -- brokers enforce by throttling: `producerByteRate` / `consumerByteRate` in bytes/second, `requestPercentage` as a per-thread-group share (legally above 100), `controllerMutationRate` in partition mutations/second. Every dial left unset means unlimited; an explicit 0 is a real, enforced zero-throughput quota.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesKafka** | `kafkaCluster` | `status.outputs.cluster_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the KafkaUser lives in (the Kafka cluster's namespace) | Placing credential consumers / Secret syncs |
| `username` | Kafka principal name (`metadata.name`) -- `User:<name>` for SCRAM, `User:CN=<name>` for TLS | Super-user lists, cluster-side ACL tooling, client identity wiring |
| `secret_name` | The operator-generated credentials Secret (equal to the user name; empty for `tls-external` users) | Workload env `valueFrom.secretKeyRef` / volume mounts -- pair with the cluster's `internal_bootstrap_endpoint` output |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Producer** -- SCRAM identity with a prefix ACL (`orders-` + `Write`/`Describe`/`Create`) covering a whole topic family. Start from the **Producer (SCRAM)** preset.

**Consumer** -- SCRAM identity with the two-part grant every consumer needs: `Read`/`Describe` on the topic and `Read` on its consumer group. Start from the **Consumer (SCRAM)** preset.

**mTLS service with quotas** -- certificate-based identity producing to one topic and consuming another under its own group, with broker-enforced byte-rate and request-time ceilings. Start from the **mTLS Service (Producer + Consumer, Quotas)** preset.

## Works With

- [**Apache Kafka**](/infra-catalog/kubernetes-kafka) -- the owning cluster; its user operator reconciles this resource, and its `internal_bootstrap_endpoint` output completes the client wiring
- [**Kafka Topic**](/infra-catalog/kubernetes-kafka-topic) -- declare the topics these ACLs govern as code
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) -- mount or env-reference this resource's `secret_name` output instead of hardcoding credentials (the same wiring applies to StatefulSets and CronJobs)
