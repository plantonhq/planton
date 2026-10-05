# Kubernetes Kafka User

## When NOT to Use This

**If the identity authenticates through a custom listener mechanism and needs no ACLs, there is nothing to declare.** A KubernetesKafkaUser exists to have the USER OPERATOR generate credentials and/or apply ACLs. An `authentication`-less user is valid for exactly one purpose: carrying ACLs for a principal authenticated OUTSIDE the user operator (custom listener mechanisms) — no credentials are generated for it. And ACLs are only enforced when the cluster runs `simple` authorization; on a cluster without authorization, an ACL-only user declares nothing enforceable.

## Overview

**KubernetesKafkaUser** declares ONE Kafka client identity on the Strimzi `KafkaUser` custom resource. The target cluster's USER OPERATOR (enabled by default on KubernetesKafka) reconciles it into a real principal: it GENERATES the credentials into a Secret named after the user (exported as `secret_name`) and — when the cluster runs `simple` authorization — applies the declared ACLs. The credentials are operator-born; no module or manifest ever carries secret material.

The contract worth internalizing before the first apply:

- **Placement** — the KafkaUser must live in the SAME NAMESPACE as its Kafka cluster and binds to it through the `strimzi.io/cluster` label (rendered from `kafka_cluster`). The user operator watches only that namespace; a user anywhere else is accepted by the API server and then silently never reconciled
- **Match the listener** — a user's authentication type must match a listener's authentication type on the target cluster: a `scram-sha-512` user cannot authenticate on a tls-auth listener and vice versa
- **What lands in the Secret** — `password` plus a ready `sasl.jaas.config` for `scram-sha-512` users; `user.crt` / `user.key` (plus `user.p12` / `user.password` keystore forms) for `tls` users. `tls-external` users bring certificates issued OUTSIDE the cluster — NO Secret is generated, and the principal is the certificate's subject (`CN=<name>`)
- **ACLs need cluster-side authorization** — effective only when the cluster's `authorization` is `simple`. Declared against a cluster without it, the rules are rejected at reconcile (the resource reports NotReady with a teaching message) — and a cluster without authorization enforces nothing anyway (every authenticated client can do everything)

## Essential Configuration Fields

### Required

- **`spec.namespace`**: MUST be the Kafka cluster's own namespace (the placement contract above)
- **`spec.kafka_cluster`**: the cluster this user belongs to — a literal cluster name or a reference to a KubernetesKafka resource; rendered as the `strimzi.io/cluster` label

### Common

- **`spec.authentication.type`**: `scram-sha-512`, `tls`, or `tls-external`
- **`spec.authorization.acls`**: the rules granting access — each names a resource (`topic`, `group`, `cluster`, `transactionalId`; `literal` or `prefix` matching) and the operations granted. A typical producer needs Write + Describe (+ IdempotentWrite for idempotent producers); a consumer needs Read + Describe on the topic and Read on its consumer group
- **`spec.quotas`**: broker-enforced client caps — producer/consumer byte rates, request-time percentage, partition-mutation rate

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Where the KafkaUser resource lives (the Kafka cluster's namespace) |
| `username` | Kafka principal name (`metadata.name`) — `User:<name>` for scram-sha-512, `User:CN=<name>` for tls |
| `secret_name` | The operator-generated credentials Secret (equal to the user name; EMPTY for tls-external users — no Secret is generated) |

## Composing in Infra Charts

`KubernetesKafka → KubernetesKafkaUser → workload` deploys in one chart run: the user references the cluster's `cluster_name` output, and workloads mount or env-reference `status.outputs.secret_name` for credentials, combining them with the cluster's `internal_bootstrap_endpoint` output. The bootstrap endpoint deliberately does not live here — it belongs to the cluster.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
