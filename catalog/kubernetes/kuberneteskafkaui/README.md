# Kubernetes Kafka UI

## When NOT to Use This

**Anyone who can reach the Service can act with the console's
permissions.** Omitting `auth` means NO authentication — acceptable
only for cluster-internal evaluation. Enable the login form (or keep
the Service internal) before any shared exposure, and set
`read_only` on clusters the console should observe but never mutate.

Also not the right kind when:

- **You need multiple console accounts, OAuth2/OIDC, or LDAP** — the
  typed `auth` models exactly ONE login-form user, because the app's
  form login authenticates against Spring's single default security
  user (it has no multi-user store — verified in the app source).
  OAuth2/OIDC, LDAP and RBAC ride the chart's own Spring
  configuration through `helm_values` — deliberately not typed.
- **You want Kafka, a registry, or Connect themselves** — the console
  observes them; the workloads are KubernetesKafka,
  KubernetesKarapace and KubernetesKafkaConnect siblings, wired in by
  reference.
- **You want external exposure baked in** — no ingress block exists.
  The console is in-cluster plumbing at the exported `endpoint`;
  compose a first-class exposure kind (KubernetesIngress / Gateway
  API) against the exported service handles. (`service_type`
  NodePort/LoadBalancer exists as a Service knob, not a
  hostname/DNS story.)
- **You need console-side authorization (per-user permissions)** —
  kafbat's RBAC configuration is Spring config; route it through
  `helm_values`. The typed surface's authorization story is
  per-cluster `read_only`.

## Overview

**KubernetesKafkaUi** deploys a kafbat UI installation — the
Apache-2.0 web console for Kafka — from the SERVED `kafka-ui` Helm
chart (https://ui.charts.kafbat.io, pinned 1.6.4). One installation
observes and manages MANY clusters: browse topics and live messages,
inspect consumer groups and lag, view and register schemas through a
connected schema registry, and monitor Connect pipes — the console
teams coming off Confluent expect.

**Each `clusters` entry wires one Kafka cluster** (plus its optional
schema registry and Connect clusters) into the console. The foreign
keys compose directly with siblings: `bootstrap_servers` from a
KubernetesKafka, `schema_registry.url` from a KubernetesKarapace,
`kafka_connect[].address` from a KubernetesKafkaConnect.

**The naming contract**: the Helm release and the chart fullname are
pinned to `metadata.name` (`fullnameOverride`), so the Service IS the
resource name, outputs stay deterministic, and several consoles
coexist in one cluster.

**Key design points:**

- **Credentials never land in rendered configuration** — the chart
  writes the app config into a ConfigMap, so the module renders every
  password position as a `${ENV_VAR}` placeholder and wires each env
  var to a Secret key through the chart's `envs.secretMappings`
  (Spring resolves the placeholders inside the running container).
  Referenced credentials (cluster SASL, registry and Connect
  passwords) point at their source Secrets; the one literal — the
  console login password — is materialized into the module-owned
  `<name>-secrets` Secret.
- **TLS trust is a mounted PEM file** — Kafka clients accept
  `ssl.truststore.type=PEM` pointing at a plain certificate file, so
  the module mounts the CA Secret as-is and a Strimzi cluster-CA
  Secret works directly: no JKS conversion, no truststore password.
- **`read_only` is an app-side switch** — hides every mutating action
  (topic create/delete, message produce, config edits) for that
  cluster; not a Kafka ACL. The right posture for production clusters
  on a shared console.
- **The deploy waits for readiness** — a console that never starts
  (bad image, unresolvable cluster config) fails THE DEPLOY, not the
  first browser hit (Helm wait + atomic on both engines).
- **`helm_values` is the escape hatch** — a YAML document merged LAST
  over everything the typed fields render (Helm `-f` semantics,
  identical on both engines): probes, security contexts, extra
  volumes, OAuth2/LDAP login. Never for secrets.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace for the console — literal or a
  KubernetesNamespace reference
- **`spec.clusters`**: at least one — unique `name` (the console's
  display/API identifier) + `bootstrap_servers` (literal or a
  KubernetesKafka reference)

### Common

- **`clusters[].read_only`**: observe-only posture per cluster
- **`clusters[].tls.ca_secret_name`**: CA Secret for the cluster
  connection (a KubernetesKafka's cluster CA by reference)
- **`clusters[].sasl`**: mechanism (PLAIN / SCRAM-SHA-256 /
  SCRAM-SHA-512) + username + `password_secret` (a
  KubernetesKafkaUser Secret by reference)
- **`clusters[].schema_registry`**: registry URL (a
  KubernetesKarapace reference) + optional HTTP Basic credentials
- **`clusters[].kafka_connect`**: Connect clusters (name + address —
  a KubernetesKafkaConnect reference + optional HTTP Basic
  credentials)
- **`clusters[].properties`**: extra Kafka client properties
  (timeouts); security properties are owned by the typed tls/sasl
  blocks — never put credentials here
- **`spec.auth`**: `type: login_form` + the SINGLE `user`
  (username + password — the password is materialized into a Secret,
  never plaintext in chart values)
- **`spec.replicas`** (stateless — availability only),
  **`spec.resources`**, **`spec.service_type` / `service_port`**,
  **`spec.node_selector` / `tolerations`**,
  **`spec.image_registry`** (air-gapped mirrors; default ghcr.io),
  **`spec.chart_version`** (default 1.6.4), **`spec.helm_values`**

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the console runs in |
| `service_name` | The console Service (`<name>` — pinned via fullnameOverride) |
| `endpoint` | In-cluster endpoint (`http://<name>.<namespace>.svc.cluster.local:<port>`) |
| `port_forward_command` | Port-forward one-liner for workstation access without any exposure |

## Composing in Infra Charts

- **The full-stack wiring**: `bootstrap_servers` references a
  KubernetesKafka (`status.outputs.internal_bootstrap_endpoint`),
  `tls.ca_secret_name` its cluster CA
  (`status.outputs.cluster_ca_cert_secret_name`),
  `sasl.password_secret.secret_name` a KubernetesKafkaUser
  (`status.outputs.secret_name`, key `password`),
  `schema_registry.url` a KubernetesKarapace
  (`status.outputs.endpoint`), and `kafka_connect[].address` a
  KubernetesKafkaConnect (`status.outputs.rest_api_endpoint`).
- **Exposure composes, never embeds**: attach a KubernetesIngress or
  Gateway API route to the `service_name` output — with `auth`
  enabled first.
- **One console, many environments**: add one `clusters` entry per
  environment and mark production `read_only`; the console is
  stateless, so replicas are purely availability.

## Examples

### Single cluster, observe-only

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafkaUi
metadata:
  name: kafka-console
spec:
  namespace:
    value: kafka-console
  create_namespace: true
  clusters:
    - name: production
      bootstrap_servers:
        value: events-kafka-bootstrap.kafka.svc.cluster.local:9092
      read_only: true
```

### Full stack with login

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafkaUi
metadata:
  name: platform-console
spec:
  namespace:
    value: kafka-console
  create_namespace: true
  clusters:
    - name: events
      bootstrap_servers:
        value: events-kafka-bootstrap.kafka.svc.cluster.local:9093
      tls:
        ca_secret_name:
          value: events-cluster-ca-cert
      sasl:
        mechanism: SCRAM-SHA-512
        username: kafka-ui
        password_secret:
          secret_name:
            value: kafka-ui
      schema_registry:
        url:
          value: http://schema-registry.kafka.svc.cluster.local:8081
      kafka_connect:
        - name: cdc
          address:
            value: http://cdc-connect-connect-api.kafka.svc.cluster.local:8083
  auth:
    type: login_form
    user:
      username: admin
      password: <set-a-strong-password>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
