# KubernetesKarapace

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesKarapaceSpec** declares a Karapace schema registry —
the Apache-2.0, Confluent-API-compatible schema registry from
Aiven. Producers and consumers register and fetch Avro, JSON
Schema and Protobuf schemas through the standard Schema Registry
REST API (existing Confluent SR clients work unchanged), and the
registry enforces compatibility between schema versions so a
producer cannot silently break its consumers.

STORAGE IS KAFKA-NATIVE: schemas live in a compacted Kafka topic
(`_schemas` by convention) on the connected cluster — the exact
architecture Confluent SR uses. No database. Multiple replicas
coordinate leadership through a consumer group; followers forward
writes to the leader.

Karapace has no official Helm chart or operator — the module owns
the deployment manifests (Deployment + Service, configuration via
KARAPACE_* environment variables, upstream's own container image),
with every meaningful configuration surface typed below.

The same engine can optionally serve the Kafka REST PROXY role
(produce/consume over HTTP); `rest_proxy` deploys it as a second,
independently-sized Deployment wired to this registry.

## Example

```yaml
# Full-surface manifest for offline module proofs (planton validate,
# tofu plan, pulumi preview). Exercises every typed arm the two engines
# must render identically: SASL_SSL Kafka (TLS CA trust + SCRAM
# credentials from an existing Secret), the full registry behavior
# block, the REST-proxy role, registry-side server TLS, HTTP Basic
# authentication, resources, and scheduling. Not a runnable-on-kind
# shape — the referenced Kafka cluster and Secrets must exist.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarapace
metadata:
  name: karapace-hack
spec:
  namespace:
    value: karapace-hack
  createNamespace: true
  replicas: 2
  kafka:
    bootstrapServers:
      value: kafka-hack-kafka-bootstrap.kafka-hack.svc.cluster.local:9094
    securityProtocol: SASL_SSL
    tls:
      caSecretName:
        value: kafka-hack-cluster-ca-cert
      caCertificate: ca.crt
    sasl:
      mechanism: SCRAM-SHA-512
      username: karapace
      passwordSecret:
        secretName:
          value: karapace-kafka-user
        key: password
  registry:
    topicName: _schemas
    replicationFactor: 3
    compatibility: FULL_TRANSITIVE
    groupId: karapace-hack-registry
    masterElectionStrategy: highest
  restProxy:
    enabled: true
    replicas: 2
    port: 8082
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 500m
        memory: 512Mi
  serverTls:
    secretName:
      value: karapace-hack-server-tls
    certificate: tls.crt
    key: tls.key
  httpAuthentication:
    basic:
      secretName:
        value: karapace-hack-authfile
      key: authfile.json
  port: 8081
  logLevel: INFO
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: "1"
      memory: 512Mi
  nodeSelector:
    workload: karapace
  tolerations:
    - key: dedicated
      operator: Equal
      value: karapace
      effect: NoSchedule
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.kafka` | `KubernetesKarapaceKafka` | yes |  |  |
| `spec.kafka.bootstrapServers` | `string \| valueFrom` | yes |  | KubernetesKafka (`status.outputs.internal_bootstrap_endpoint`) |
| `spec.kafka.securityProtocol` | `string` |  | `PLAINTEXT` |  |
| `spec.kafka.tls` | `KubernetesKarapaceKafkaTls` |  |  |  |
| `spec.kafka.tls.caSecretName` | `string \| valueFrom` | yes |  | KubernetesKafka (`status.outputs.cluster_ca_cert_secret_name`) |
| `spec.kafka.tls.caCertificate` | `string` |  | `ca.crt` |  |
| `spec.kafka.tls.clientCertSecretName` | `string \| valueFrom` |  |  | KubernetesKafkaUser (`status.outputs.secret_name`) |
| `spec.kafka.tls.clientCertificate` | `string` |  | `user.crt` |  |
| `spec.kafka.tls.clientKey` | `string` |  | `user.key` |  |
| `spec.kafka.sasl` | `KubernetesKarapaceKafkaSasl` |  |  |  |
| `spec.kafka.sasl.mechanism` | `string` | yes |  |  |
| `spec.kafka.sasl.username` | `string` | yes |  |  |
| `spec.kafka.sasl.passwordSecret` | `KubernetesKarapacePasswordSecret` |  |  |  |
| `spec.kafka.sasl.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesKafkaUser (`status.outputs.secret_name`) |
| `spec.kafka.sasl.passwordSecret.key` | `string` |  | `password` |  |
| `spec.kafka.sasl.password` | `string` (sensitive) |  |  |  |
| `spec.registry` | `KubernetesKarapaceRegistry` |  |  |  |
| `spec.registry.topicName` | `string` |  | `_schemas` |  |
| `spec.registry.replicationFactor` | `int32` |  |  |  |
| `spec.registry.compatibility` | `string` |  | `BACKWARD` |  |
| `spec.registry.groupId` | `string` |  |  |  |
| `spec.registry.masterElectionStrategy` | `string` |  | `lowest` |  |
| `spec.restProxy` | `KubernetesKarapaceRestProxy` |  |  |  |
| `spec.restProxy.enabled` | `bool` |  |  |  |
| `spec.restProxy.replicas` | `int32` |  | `1` |  |
| `spec.restProxy.port` | `int32` |  | `8082` |  |
| `spec.restProxy.resources` | `ContainerResources` |  |  |  |
| `spec.restProxy.resources.limits` | `CpuMemory` |  |  |  |
| `spec.restProxy.resources.limits.cpu` | `string` |  |  |  |
| `spec.restProxy.resources.limits.memory` | `string` |  |  |  |
| `spec.restProxy.resources.requests` | `CpuMemory` |  |  |  |
| `spec.restProxy.resources.requests.cpu` | `string` |  |  |  |
| `spec.restProxy.resources.requests.memory` | `string` |  |  |  |
| `spec.serverTls` | `KubernetesKarapaceServerTls` |  |  |  |
| `spec.serverTls.secretName` | `string \| valueFrom` | yes |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.serverTls.certificate` | `string` |  | `tls.crt` |  |
| `spec.serverTls.key` | `string` |  | `tls.key` |  |
| `spec.httpAuthentication` | `KubernetesKarapaceHttpAuthentication` |  |  |  |
| `spec.httpAuthentication.basic` | `KubernetesKarapaceBasicAuth` |  |  |  |
| `spec.httpAuthentication.basic.secretName` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.httpAuthentication.basic.key` | `string` |  | `authfile.json` |  |
| `spec.httpAuthentication.oidc` | `KubernetesKarapaceOidc` |  |  |  |
| `spec.httpAuthentication.oidc.jwksEndpointUrl` | `string` | yes |  |  |
| `spec.httpAuthentication.oidc.expectedIssuer` | `string` |  |  |  |
| `spec.httpAuthentication.oidc.expectedAudience` | `string` |  |  |  |
| `spec.port` | `int32` |  | `8081` |  |
| `spec.logLevel` | `string` |  | `INFO` |  |
| `spec.image` | `string` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.tolerations[].key` | `string` |  |  |  |
| `spec.tolerations[].operator` | `string` |  |  |  |
| `spec.tolerations[].value` | `string` |  |  |  |
| `spec.tolerations[].effect` | `string` |  |  |  |
| `spec.tolerations[].tolerationSeconds` | `int64` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace for the registry. Accepts a literal namespace name or
a reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the
resource. When false, the namespace must already exist.

### spec.replicas

`int32` · optional (explicit presence)

Number of registry replicas. Replicas elect a leader through
the registry's consumer group; followers forward writes — more
than one replica is an availability measure, not a write-scaling
one. 2 is the production norm.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.kafka

`KubernetesKarapaceKafka` · required

The Kafka cluster that stores the schemas.

- rule: {"required":true}
- rule: SSL and SASL_SSL security protocols require the tls block (at minimum the CA to trust)
- rule: SASL_PLAINTEXT and SASL_SSL security protocols require the sasl block (mechanism and credentials)
- rule: the sasl block requires security_protocol to be EXPLICITLY set to SASL_PLAINTEXT or SASL_SSL — the protocol defaults to PLAINTEXT when unset, which would silently ignore the credentials

### spec.kafka.bootstrapServers

`string | valueFrom` · required

Bootstrap address of the Kafka cluster, as host:port. Accepts a
literal address or a reference to a KubernetesKafka resource,
which resolves to its in-cluster bootstrap endpoint.

containment_exempt: the registry SERVES SCHEMAS FOR this
cluster; it is not deployed inside it.

- references: KubernetesKafka (`status.outputs.internal_bootstrap_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKafka, name: <that resource's name>, fieldPath: status.outputs.internal_bootstrap_endpoint}} -- a bare string does not parse

### spec.kafka.securityProtocol

`string` · optional (explicit presence)

Kafka security protocol: "PLAINTEXT" (default), "SSL",
"SASL_PLAINTEXT", or "SASL_SSL". Must match the target
listener's posture — SSL forms require `tls`; SASL forms
require `sasl`.

- default: `PLAINTEXT`
- rule: security_protocol must be one of PLAINTEXT, SSL, SASL_PLAINTEXT, SASL_SSL

### spec.kafka.tls

`KubernetesKarapaceKafkaTls`

TLS trust and (optionally) mutual-TLS client identity for the
Kafka connection (SSL / SASL_SSL protocols).

### spec.kafka.tls.caSecretName

`string | valueFrom` · required

Name of the Secret holding the CA certificate to trust. Accepts
a literal Secret name or a reference to a KubernetesKafka
resource, which resolves to that cluster's CA certificate
Secret.

containment_exempt: trust material fetched FROM the cluster —
access, never placement.

- references: KubernetesKafka (`status.outputs.cluster_ca_cert_secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKafka, name: <that resource's name>, fieldPath: status.outputs.cluster_ca_cert_secret_name}} -- a bare string does not parse

### spec.kafka.tls.caCertificate

`string` · optional (explicit presence)

The CA certificate's key within the Secret. Strimzi cluster CA
Secrets carry "ca.crt".

- default: `ca.crt`

### spec.kafka.tls.clientCertSecretName

`string | valueFrom`

Mutual-TLS client identity (tls-authenticated listeners): name
of the Secret holding the client certificate and key. Accepts a
literal Secret name or a reference to a KubernetesKafkaUser
resource (tls authentication), which resolves to its
operator-generated credential Secret.

- references: KubernetesKafkaUser (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKafkaUser, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.kafka.tls.clientCertificate

`string` · optional (explicit presence)

The client certificate's key within the Secret.
KubernetesKafkaUser credential Secrets carry "user.crt".

- default: `user.crt`

### spec.kafka.tls.clientKey

`string` · optional (explicit presence)

The client private key's key within the Secret.
KubernetesKafkaUser credential Secrets carry "user.key".

- default: `user.key`

### spec.kafka.sasl

`KubernetesKarapaceKafkaSasl`

SASL credentials (SASL_PLAINTEXT / SASL_SSL protocols).

- rule: set exactly one of password_secret (a key in an existing Secret — reference a KubernetesKafkaUser) or password (a declared value the module materializes into a Secret)

### spec.kafka.sasl.mechanism

`string` · required

SASL mechanism: "PLAIN", "SCRAM-SHA-256", or "SCRAM-SHA-512"
(Strimzi scram-sha-512 listeners pair with SCRAM-SHA-512).

- rule: mechanism must be PLAIN, SCRAM-SHA-256, or SCRAM-SHA-512
- rule: {"required":true}

### spec.kafka.sasl.username

`string` · required

SASL username.

- rule: {"required":true}

### spec.kafka.sasl.passwordSecret

`KubernetesKarapacePasswordSecret`

SASL password, from a key within a Secret — reference a
KubernetesKafkaUser resource (scram-sha-512 authentication) to
use its operator-generated Secret. Exactly one of
password_secret or password must be set.

### spec.kafka.sasl.passwordSecret.secretName

`string | valueFrom` · required

Name of the Secret. Accepts a literal Secret name or a reference
to a KubernetesKafkaUser resource, which resolves to its
credential Secret.

- references: KubernetesKafkaUser (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKafkaUser, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.kafka.sasl.passwordSecret.key

`string` · optional (explicit presence)

The key within the Secret whose value is the password.
KubernetesKafkaUser credential Secrets carry it under
"password".

- default: `password`

### spec.kafka.sasl.password

`string` · sensitive

SASL password as a declared value (external clusters where no
Secret exists yet) — the module materializes it into a Secret
and mounts it; it never lands in the pod spec as plaintext.
Exactly one of password_secret or password must be set.

### spec.registry

`KubernetesKarapaceRegistry`

Schema-registry behavior knobs.

### spec.registry.topicName

`string` · optional (explicit presence)

The compacted Kafka topic storing schemas. The Confluent SR
convention — and the value existing tooling expects — is
"_schemas". The registry creates it on first start.

- default: `_schemas`

### spec.registry.replicationFactor

`int32` · optional (explicit presence)

Replication factor for the schemas topic AT CREATION. The
upstream default is 1 — fine for dev, a data-loss risk in
production: set 3 on multi-broker clusters (cannot exceed the
broker count). Changing it later means reassigning the existing
topic with Kafka tooling, not editing this field.

- rule: {"int32":{"lte":32767,"gte":1}}

### spec.registry.compatibility

`string` · optional (explicit presence)

Default compatibility mode for new subjects: BACKWARD (default —
consumers on the new schema read old records), FORWARD, FULL,
their _TRANSITIVE variants (checked against ALL prior versions,
not just the latest), or NONE. Per-subject overrides ride the
standard SR config API.

- default: `BACKWARD`
- rule: compatibility must be one of BACKWARD, BACKWARD_TRANSITIVE, FORWARD, FORWARD_TRANSITIVE, FULL, FULL_TRANSITIVE, NONE

### spec.registry.groupId

`string`

Consumer group the replicas coordinate leadership through.
Empty = "<metadata.name>". Must be unique per registry
installation sharing a Kafka cluster — two registries sharing a
group id fight over leadership.

### spec.registry.masterElectionStrategy

`string` · optional (explicit presence)

Leader election strategy among replicas: "lowest" (default) or
"highest" (member ordering).

- default: `lowest`
- rule: master_election_strategy must be lowest or highest

### spec.restProxy

`KubernetesKarapaceRestProxy`

Optional Kafka REST proxy role — produce/consume/admin over
HTTP, served by a second Deployment of the same engine.

### spec.restProxy.enabled

`bool`

Deploy the Kafka REST proxy (a second Deployment of the same
engine, wired to this registry for schema-aware
produce/consume).

### spec.restProxy.replicas

`int32` · optional (explicit presence)

Number of REST-proxy replicas.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.restProxy.port

`int32` · optional (explicit presence)

HTTP port the proxy serves on.

- default: `8082`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.restProxy.resources

`ContainerResources`

CPU/memory for the proxy pods.

### spec.restProxy.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.restProxy.resources.limits.cpu

`string`

### spec.restProxy.resources.limits.memory

`string`

### spec.restProxy.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.restProxy.resources.requests.cpu

`string`

### spec.restProxy.resources.requests.memory

`string`

### spec.serverTls

`KubernetesKarapaceServerTls`

Serve the registry API over TLS (the cert-manager seam).
Omitted = plain HTTP (fine inside the cluster; put TLS at an
Ingress/Gateway for external exposure, or terminate here).

WITH MULTIPLE REPLICAS: followers forward writes to the leader at
its advertised POD IP, so the serving certificate would need to
cover pod IPs — which certificates for a DNS name do not. Pair
server_tls with replicas: 1, or run multiple plain-HTTP replicas
behind TLS terminated at an Ingress/Gateway.

### spec.serverTls.secretName

`string | valueFrom` · required

Name of the Secret holding the server certificate and key.
Accepts a literal Secret name or a reference to a
KubernetesCertificate resource's output secret.

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.serverTls.certificate

`string` · optional (explicit presence)

Certificate key within the Secret. cert-manager writes
"tls.crt".

- default: `tls.crt`

### spec.serverTls.key

`string` · optional (explicit presence)

Private-key key within the Secret. cert-manager writes
"tls.key".

- default: `tls.key`

### spec.httpAuthentication

`KubernetesKarapaceHttpAuthentication`

Authentication for the registry's own HTTP API. Omitted = the
API is open to anyone who can reach the Service.

- rule: configure exactly one of basic (authfile-backed HTTP Basic) or oidc (JWT bearer tokens)

### spec.httpAuthentication.basic

`KubernetesKarapaceBasicAuth`

HTTP Basic authentication: name of a Secret holding a Karapace
authfile (JSON users/permissions document) under the given key.
The registry hot-reloads the file on change.

### spec.httpAuthentication.basic.secretName

`string | valueFrom` · required

Name of the Secret holding the authfile. Accepts a literal
Secret name or a reference to a KubernetesSecret resource.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.httpAuthentication.basic.key

`string` · optional (explicit presence)

The key within the Secret whose value is the authfile JSON.

- default: `authfile.json`

### spec.httpAuthentication.oidc

`KubernetesKarapaceOidc`

OIDC bearer-token authentication — validates JWTs against your
identity provider's JWKS endpoint.

### spec.httpAuthentication.oidc.jwksEndpointUrl

`string` · required

JWKS endpoint of the identity provider (the signing keys used
to verify tokens), e.g.
"https://idp.example.com/.well-known/jwks.json". Must be HTTPS —
a plain-HTTP JWKS source lets an in-path attacker forge tokens
(the upstream engine refuses it outside dev overrides).

- rule: jwks_endpoint_url must be an https:// URL — plain HTTP would let an attacker substitute signing keys
- rule: {"required":true}

### spec.httpAuthentication.oidc.expectedIssuer

`string`

Expected `iss` claim. Empty = not checked.

### spec.httpAuthentication.oidc.expectedAudience

`string`

Expected `aud` claim. Empty = not checked.

### spec.port

`int32` · optional (explicit presence)

HTTP port the registry serves on.

- default: `8081`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.logLevel

`string` · optional (explicit presence)

Log level: DEBUG, INFO, WARNING, or ERROR. The upstream default
is DEBUG — noisy for production; INFO is the sensible steady
state.

- default: `INFO`
- rule: log_level must be DEBUG, INFO, WARNING, or ERROR

### spec.image

`string`

Container image. Empty = the pinned upstream release image
(ghcr.io/aiven-open/karapace at the module's pinned tag).
Override for air-gapped registries or version pinning.

### spec.resources

`ContainerResources`

CPU/memory for each registry pod. Karapace is deliberately
lightweight — modest requests go far.

### spec.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.resources.limits.cpu

`string`

### spec.resources.limits.memory

`string`

### spec.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.resources.requests.cpu

`string`

### spec.resources.requests.memory

`string`

### spec.nodeSelector

`map<string, string>`

Node selector for the registry pods.

### spec.tolerations

`[]WorkloadToleration`

Tolerations for the registry pods.

### spec.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesKarapace, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the registry runs in. |
| `status.outputs.service_name` | `string` | Name of the registry Service. |
| `status.outputs.endpoint` | `string` | In-cluster registry endpoint (`http(s)://<name>.<namespace>.svc.cluster.local:<port>`) — the schema.registry.url value for clients. |
| `status.outputs.rest_proxy_endpoint` | `string` | In-cluster REST-proxy endpoint (empty when the rest_proxy role is not enabled). |
| `status.outputs.schemas_topic` | `string` | The Kafka topic storing the schemas. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.kafka.bootstrapServers` | KubernetesKafka | `status.outputs.internal_bootstrap_endpoint` |
| `spec.kafka.tls.caSecretName` | KubernetesKafka | `status.outputs.cluster_ca_cert_secret_name` |
| `spec.kafka.tls.clientCertSecretName` | KubernetesKafkaUser | `status.outputs.secret_name` |
| `spec.kafka.sasl.passwordSecret.secretName` | KubernetesKafkaUser | `status.outputs.secret_name` |
| `spec.serverTls.secretName` | KubernetesCertificate | `status.outputs.secret_name` |
| `spec.httpAuthentication.basic.secretName` | KubernetesSecret | `status.outputs.secret_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesKafkaUi | `spec.clusters[].schemaRegistry.url` | `status.outputs.endpoint` |

## See Also

- [Overview](../README.md)
