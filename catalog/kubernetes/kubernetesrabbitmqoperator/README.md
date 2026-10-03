# Kubernetes RabbitMQ Operator

## When NOT to Use This

**This component installs the ENGINE, not a message broker.** The
RabbitMQ Cluster Operator reconciles `RabbitmqCluster` custom
resources into running RabbitMQ clusters (StatefulSet, Services,
generated credentials, rolling upgrades); those clusters are declared
with KubernetesRabbitMq — one resource per cluster. Install the
operator once per Kubernetes cluster, then declare RabbitMQ clusters
against it.

Also not the right kind when:

- **You want a RabbitMQ cluster** — that is KubernetesRabbitMq; this
  component is the controller that reconciles it.
- **You want declarative queues, exchanges, users, vhosts, or
  policies** — those are served by RabbitMQ's
  messaging-topology-operator, a separate upstream product this
  component deliberately does NOT install. This is the CLUSTER
  operator only.
- **You want a single-container dev broker** — a throwaway
  RabbitMQ-in-a-pod does not need an operator; use
  KubernetesDeployment.
- **You want Kafka-style event streaming** — that is
  KubernetesKafka; RabbitMQ is a message broker, not a distributed
  log.

## Overview

**KubernetesRabbitMqOperator** installs the RabbitMQ Cluster
Operator — the MPL-2.0 operator maintained by the RabbitMQ team
(rabbitmq/cluster-operator) — from its released single-file manifest.
The operator reconciles `RabbitmqCluster` custom resources (declared
with KubernetesRabbitMq) into running RabbitMQ clusters.

**Key design points:**

- **No Helm chart — the release manifest IS the distribution.** The
  operator's official distribution is the single-file
  `cluster-operator.yml` attached to each GitHub release; the modules
  fetch it from the pinned release tag (v2.22.3) and apply it per
  document. The released asset pins the operator image
  `ghcr.io/rabbitmq/cluster-operator:2.22.3`.
- **No version field — deliberately.** The installed operator and its
  `RabbitmqCluster` CRD schema are pinned to the release this
  catalog's typed SDK was generated against; a user-selectable
  version would silently drift the CRD schema away from the typed
  resources built on it.
- **The namespace is FIXED at `rabbitmq-system`.** The name is baked
  into the manifest's own cross-references (webhook client
  configuration, Certificate DNS names, CA-injection annotations,
  cluster-role binding subjects). Exactly ONE install per cluster —
  the admission webhooks are cluster-scoped singletons with fixed
  names, so a second install cannot coexist.
- **cert-manager is a HARD prerequisite** (declared as a registry
  prerequisite of this kind). The manifest ships mutating and
  validating admission webhooks for `RabbitmqCluster` with
  `failurePolicy: Fail`, whose serving certificate is a cert-manager
  `Certificate` (self-signed `Issuer`) with CA injection. Without a
  running cert-manager the certificate never issues and every
  RabbitmqCluster admission fails. Compose with
  KubernetesCertManager.
- **CRD LIFECYCLE WARNING: the CRD deletes with the resource.** The
  `RabbitmqCluster` CRD is one document of the applied manifest —
  destroying the operator deletes the CRD, which cascade-deletes
  every RabbitmqCluster on the cluster. Never destroy the operator
  while KubernetesRabbitMq resources exist.
- **Empty `watch_namespaces` watches EVERYTHING.** The upstream
  default is cluster-wide — the opposite default from chart-scoped
  operators like the Altinity operator. Setting entries fences the
  watch (rendered as the operator's `OPERATOR_SCOPE_NAMESPACE`
  environment variable, comma-separated); every namespace that will
  hold KubernetesRabbitMq resources must then be covered.

## Essential Configuration Fields

### Required

None — `spec: {}` is a valid, production-standard install: the
release manifest's own defaults, watching all namespaces. There is no
namespace field (the manifest's `rabbitmq-system` is fixed) and no
version field (pinned by design).

### Common

- **`spec.watch_namespaces`**: namespaces the operator watches for
  RabbitmqCluster resources; empty = ALL namespaces (the upstream
  default), entries fence the watch on multi-tenant clusters
- **`spec.default_rabbitmq_image`**: fleet-wide default RabbitMQ
  server image for every RabbitmqCluster that does not pin its own
  (empty = the compiled-in default `rabbitmq:4.2.6-management` — the
  `-management` variant is required); set for air-gapped clusters
- **`spec.default_user_updater_image`**: default image for the
  credential-updater sidecar Vault-backed clusters run (empty = the
  compiled-in default
  `ghcr.io/rabbitmq/default-user-credential-updater:1.0.14`)
- **`spec.operator_image`**: override the operator image itself
  (air-gap / private-mirror path; empty = the release manifest's
  pinned `ghcr.io/rabbitmq/cluster-operator` at the release tag)
- **`spec.resources`**: operator container resources — empty = the
  release manifest's defaults (200m CPU / 500Mi memory for both
  requests and limits)
- **`spec.node_selector` / `spec.tolerations`**: operator pod
  scheduling
- **`spec.image_pull_secrets`**: names of image-pull secrets (in the
  `rabbitmq-system` namespace) for pulling the operator image from a
  private mirror

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the operator is installed into (always `rabbitmq-system` — the manifest's fixed namespace) |
| `deployment_name` | The operator Deployment (`rabbitmq-cluster-operator` — the manifest's fixed name) |
| `metrics_endpoint` | In-cluster Prometheus metrics endpoint (`http://rabbitmq-cluster-operator-metrics-service.rabbitmq-system.svc.cluster.local:8080/metrics`) |
| `crd_name` | The RabbitmqCluster CustomResourceDefinition (`rabbitmqclusters.rabbitmq.com`) — deleted with this resource |

## Composing in Infra Charts

- **KubernetesCertManager comes first**: cert-manager must be running
  before this operator installs — the admission webhooks' serving
  certificate is a cert-manager Certificate, and without it every
  RabbitmqCluster admission fails.
- **KubernetesRabbitMq resources depend on this component**: the
  operator must be running and watching their namespace before their
  RabbitmqCluster resources reconcile. With the default (empty) watch
  scope, one install serves the whole cluster; with a fenced
  `watch_namespaces` list, keep clusters inside the watched
  namespaces.
- **Destroy ordering matters**: the CRD deletes with this resource,
  cascade-deleting every RabbitmqCluster. Destroy all
  KubernetesRabbitMq resources before ever destroying the operator.

## Examples

The smallest declarable install is also the production standard —
every field has a working default:

### Standard cluster-wide install

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRabbitMqOperator
metadata:
  name: rabbitmq-operator
spec: {}
```

### Namespace-fenced install on a multi-tenant cluster

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRabbitMqOperator
metadata:
  name: rabbitmq-operator
spec:
  watch_namespaces:
    - messaging
    - integrations
```

### Air-gapped mirror

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRabbitMqOperator
metadata:
  name: rabbitmq-operator
spec:
  operator_image:
    repo: registry.example.internal/rabbitmq/cluster-operator
    tag: 2.22.3
    pull_secret_name: mirror-pull
  default_rabbitmq_image: registry.example.internal/rabbitmq:4.2.6-management
  default_user_updater_image: registry.example.internal/default-user-credential-updater:1.0.14
  image_pull_secrets:
    - mirror-pull
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
