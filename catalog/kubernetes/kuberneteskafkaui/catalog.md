# Kafka UI

Deploys kafbat UI — the Apache-2.0 web console for Kafka — from the served `kafka-ui` Helm chart. One installation observes and manages **many** clusters: browse topics and live messages, inspect consumer groups and lag, view schemas through a connected registry, and monitor Connect pipes. Each `clusters` entry wires one Kafka cluster (plus its optional schema registry and Connect clusters) into the console; the foreign keys compose directly with KubernetesKafka, KubernetesKarapace, and KubernetesKafkaConnect siblings.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Helm Release** (`<metadata.name>`) — the kafbat UI chart with the Service pinned to the resource name (`fullnameOverride`), so outputs stay deterministic and several consoles can coexist in one cluster
- **Console Service** — ClusterIP by default; compose KubernetesIngress or Gateway routes against the exported `service_name` / `endpoint` handles for shared access
- **Module-owned Secrets** — the console login password (when declared) materializes into `<name>-secrets`; referenced cluster SASL and basic-auth passwords mount from their source Secrets — nothing sensitive lands in rendered chart values
- **Namespace** (optional) — created with standard governance labels when `createNamespace` is true

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- At least one **Apache Kafka** cluster to observe (bootstrap endpoint wired by reference or literal)
- Optional **Karapace Schema Registry** for schema browsing and **Kafka Connect** for connector monitoring — wired per cluster entry
- Optional **Kafka User** for SASL credentials — reference its operator-generated Secret; never type passwords into the console manifest

## Deploy

### Console

Open the deployment store, find **Kafka UI**, and click **Deploy**. The creation wizard walks you through the console's namespace, the served-chart pin, the cluster composition step (where the whole Kafka family's outputs meet), console login, Service exposure, sizing, scheduling, and the Helm-values escape hatch. Start from the **Single cluster readonly preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafkaUi
metadata:
  name: kafka-console
  org: acme-corp
  env: prod
spec:
  namespace:
    value: kafka-console
  createNamespace: true
  clusters:
    - name: production
      bootstrapServers:
        value: events-kafka-bootstrap.kafka.svc.cluster.local:9092
      readOnly: true
```

```shell
planton apply -f kafka-console.yaml
```

This creates an observe-only console for one cluster; reach it through the exported port-forward command until you compose exposure, and add `login_form` before anything shared can reach the Service. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire sibling outputs so the family cannot drift:

```yaml
spec:
  namespace:
    value: kafka-console
  clusters:
    - name: events
      bootstrapServers:
        valueFrom:
          kind: KubernetesKafka
          name: event-bus
          fieldPath: status.outputs.internal_bootstrap_endpoint
      tls:
        caSecretName:
          valueFrom:
            kind: KubernetesKafka
            name: event-bus
            fieldPath: status.outputs.cluster_ca_cert_secret_name
      sasl:
        mechanism: SCRAM-SHA-512
        username: kafka-ui
        passwordSecret:
          secretName:
            valueFrom:
              kind: KubernetesKafkaUser
              name: kafka-ui-user
              fieldPath: status.outputs.secret_name
      schemaRegistry:
        url:
          valueFrom:
            kind: KubernetesKarapace
            name: schema-registry
            fieldPath: status.outputs.endpoint
      kafkaConnect:
        - name: cdc
          address:
            valueFrom:
              kind: KubernetesKafkaConnect
              name: cdc-connect
              fieldPath: status.outputs.rest_api_endpoint
```

The InfraPipeline deploys the Kafka family first, then wires the console against every resolved endpoint.

## Key Configuration

These are the most important decisions when configuring the console. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Open vs login-gated console** — omitting `auth` means NO authentication: anyone who can reach the Service has full console access on every wired cluster. Acceptable only for cluster-internal evaluation paired with ClusterIP and no composed route. The typed login is exactly ONE `login_form` account; OAuth2/OIDC/LDAP compose through `helmValues`.

**Per-cluster `readOnly`** — an app-side switch that hides every mutating action (topic create/delete, message produce, config edits) for that cluster. The right posture for production clusters on a shared console — set it where the risk lives, not globally.

**Exposure composes, never embeds** — prefer ClusterIP + a composed Ingress/Gateway against the exported handles. NodePort/LoadBalancer exist as Service knobs, not a hostname story.

**Credentials never land in rendered configuration** — SASL and basic-auth passwords are Secret references; the console login password is the one literal the module materializes into a Secret-backed environment variable.

**`helmValues` is the escape hatch** — YAML merged LAST over everything the typed fields render: probes, security contexts, OAuth2/LDAP login. Never for secrets.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesKafka** | `clusters[].bootstrapServers` | `status.outputs.internal_bootstrap_endpoint` |
| **KubernetesKafka** | `clusters[].tls.caSecretName` | `status.outputs.cluster_ca_cert_secret_name` |
| **KubernetesKafkaUser** | `clusters[].sasl.passwordSecret.secretName` | `status.outputs.secret_name` |
| **KubernetesKarapace** | `clusters[].schemaRegistry.url` | `status.outputs.endpoint` |
| **KubernetesKafkaConnect** | `clusters[].kafkaConnect[].address` | `status.outputs.rest_api_endpoint` |

### What This Kind Provides

After provisioning, `status.outputs` contains values downstream Infra Components can consume:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the console runs in | Composed route placement context |
| `service_name` | Name of the console Service | KubernetesIngress / Gateway backend target |
| `endpoint` | In-cluster HTTP endpoint | Internal tooling links |
| `port_forward_command` | Copyable kubectl port-forward | Quick look with no exposure |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single cluster readonly** — one cluster wired with `readOnly` on, no login, no exposure: the safe first console for observing production. Start from the **Single cluster readonly preset**.

**Full stack console** — TLS + SCRAM cluster connection, schema registry, Connect monitoring, and a login gate. Start from the **Full stack console preset**.

**Multi-cluster org console** — staging with full powers and production locked to observe-only on one shared pane. Start from the **Multi cluster preset**.

## Works With

- [**Apache Kafka**](/infra-catalog/kubernetes-kafka) — bootstrap endpoint and cluster CA for each wired cluster
- [**Kafka User**](/infra-catalog/kubernetes-kafka-user) — SASL credential Secrets for secured listeners
- [**Karapace Schema Registry**](/infra-catalog/kubernetes-karapace) — schema registry endpoint for schema-aware browsing
- [**Kafka Connect**](/infra-catalog/kubernetes-kafka-connect) — REST endpoints for connector monitoring
- [**Kubernetes Ingress**](/infra-catalog/kubernetes-ingress) — shared exposure composed against the exported Service handles
- [**Kubernetes HTTPRoute**](/infra-catalog/kubernetes-http-route) — the Gateway API alternative for the same exposure seam
