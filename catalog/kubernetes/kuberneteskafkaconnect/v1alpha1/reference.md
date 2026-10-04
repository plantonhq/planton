# KubernetesKafkaConnect

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesKafkaConnectSpec** declares a Kafka Connect cluster on
the Strimzi `KafkaConnect` custom resource — the pluggable
integration engine that streams data between Kafka and external
systems (databases via Debezium CDC, object stores, search indexes,
SaaS APIs). The Strimzi cluster operator
(KubernetesStrimziKafkaOperator) reconciles it into Connect worker
pods; individual pipes are declared as KubernetesKafkaConnector
resources against this cluster.

CONNECTOR MANAGEMENT: the module always sets the
`strimzi.io/use-connector-resources: "true"` annotation, so
connectors are managed DECLARATIVELY through KubernetesKafkaConnector
resources. Do not create or modify connectors through the Connect
REST API on this cluster — the operator reverts changes it does not
own.

CONNECTOR PLUGINS reach the workers four ways, from simplest to
most self-contained:
1. The STOCK image — carries ONLY the MirrorMaker 2 connectors
   (MirrorSource/MirrorCheckpoint/MirrorHeartbeat — verified live
   against the workers' own plugin listing; Kafka's FileStream
   example connectors are NOT on the distribution's classpath, so
   every real integration needs one of the arms below).
2. `image` — run a prebuilt Connect image that already carries your
   plugins (the fastest path when a vendor publishes one).
3. `plugins` — mount plugins from OCI artifacts as Kubernetes image
   volumes (no image build; requires the cluster's ImageVolume
   feature — see the field comment).
4. `build` — have the operator BUILD a custom image from declared
   artifacts (Kaniko/Buildah on Kubernetes) and push it to your
   registry.
When `build` is configured the operator runs the built image and
`image` is not used — set one or the other.

The GROUP IDENTITY fields (`group_id` and the three storage topics)
default from metadata.name and MUST be unique per Connect cluster
sharing a Kafka cluster — two Connect clusters sharing a group.id
or a storage topic corrupt each other's state.

## Example

```yaml
# Full-surface development manifest: exercises every typed arm so the
# offline plan proofs cover surfaces the live lanes exclude (TLS trust
# with both certificate and pattern selection, SASL SCRAM authentication,
# OCI image-volume plugins, an operator-driven image build with maven and
# jar artifacts, rack awareness, scheduling knobs). Not a runnable-on-kind
# shape — the image-volume plugins need the ImageVolume feature, the
# build needs a pushable registry, and the zone-keyed rack needs labeled
# nodes. The group identity fields (groupId, storage topics) are LEFT
# UNSET on purpose: the proofs verify their metadata.name-derived
# defaults.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafkaConnect
metadata:
  name: connect-hack
spec:
  namespace:
    value: connect-hack
  createNamespace: true
  version: 4.3.0
  replicas: 3
  bootstrapServers:
    value: kafka-hack-kafka-bootstrap.kafka-hack.svc.cluster.local:9094
  tls:
    trustedCertificates:
      - secretName:
          value: kafka-hack-cluster-ca-cert
        certificate: ca.crt
      - secretName:
          value: extra-trust-bundle
        pattern: "*.crt"
  authentication:
    type: scram-sha-512
    username: connect-user
    passwordSecret:
      secretName:
        value: connect-user
      password: password
  config:
    key.converter: org.apache.kafka.connect.json.JsonConverter
    value.converter: org.apache.kafka.connect.json.JsonConverter
    key.converter.schemas.enable: "false"
    value.converter.schemas.enable: "false"
    config.storage.replication.factor: "3"
    offset.storage.replication.factor: "3"
    status.storage.replication.factor: "3"
  plugins:
    - name: debezium-postgres
      artifacts:
        - reference: quay.io/example/debezium-postgres:3.1.0
          pullPolicy: IfNotPresent
  build:
    output:
      type: docker
      image: registry.example.com/team/connect-hack:latest
      pushSecret: registry-push-creds
      additionalBuildOptions:
        - "--verbosity=info"
      additionalPushOptions:
        - "--tls-verify=true"
    plugins:
      - name: debezium-connector-postgres
        artifacts:
          - type: maven
            repository: https://repo1.maven.org/maven2
            group: io.debezium
            artifact: debezium-connector-postgres
            version: 3.1.0.Final
      - name: camel-timer
        artifacts:
          - type: jar
            url: https://repo.example.com/connectors/camel-timer-kafka-connector-4.8.0.jar
            sha512sum: 8b1d2a4f8c9e3b5a7d6f0c2e4a8b6d9f1c3e5a7b9d0f2c4e6a8b0d2f4c6e8a0b1d3f5c7e9a1b3d5f7c9e1a3b5d7f9c1e3a5b7d9f1c3e5a7b9d1f3c5e7a9b1d3c5
            insecure: true
  resources:
    requests:
      cpu: 250m
      memory: 1Gi
    limits:
      cpu: "1"
      memory: 2Gi
  jvm:
    xms: 1g
    xmx: 1g
  rack:
    topologyKey: topology.kubernetes.io/zone
  metrics:
    enabled: true
  nodeSelector:
    workload: connect
  tolerations:
    - key: dedicated
      operator: Equal
      value: connect
      effect: NoSchedule
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.version` | `string` |  |  |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.bootstrapServers` | `string \| valueFrom` | yes |  | KubernetesKafka (`status.outputs.internal_bootstrap_endpoint`) |
| `spec.tls` | `StrimziKafkaClientTls` |  |  |  |
| `spec.tls.trustedCertificates` | `[]StrimziKafkaClientTrustedCertificate` | yes |  |  |
| `spec.tls.trustedCertificates[].secretName` | `string \| valueFrom` | yes |  | KubernetesKafka (`status.outputs.cluster_ca_cert_secret_name`) |
| `spec.tls.trustedCertificates[].certificate` | `string` |  |  |  |
| `spec.tls.trustedCertificates[].pattern` | `string` |  |  |  |
| `spec.authentication` | `StrimziKafkaClientAuthentication` |  |  |  |
| `spec.authentication.type` | `string` | yes |  |  |
| `spec.authentication.certificateAndKey` | `StrimziKafkaClientCertificateAndKey` |  |  |  |
| `spec.authentication.certificateAndKey.secretName` | `string \| valueFrom` | yes |  | KubernetesKafkaUser (`status.outputs.secret_name`) |
| `spec.authentication.certificateAndKey.certificate` | `string` |  | `user.crt` |  |
| `spec.authentication.certificateAndKey.key` | `string` |  | `user.key` |  |
| `spec.authentication.username` | `string` |  |  |  |
| `spec.authentication.passwordSecret` | `StrimziKafkaClientPasswordSecret` |  |  |  |
| `spec.authentication.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesKafkaUser (`status.outputs.secret_name`) |
| `spec.authentication.passwordSecret.password` | `string` |  | `password` |  |
| `spec.authentication.sasl` | `bool` |  |  |  |
| `spec.authentication.config` | `map<string, string>` |  |  |  |
| `spec.groupId` | `string` |  |  |  |
| `spec.configStorageTopic` | `string` |  |  |  |
| `spec.statusStorageTopic` | `string` |  |  |  |
| `spec.offsetStorageTopic` | `string` |  |  |  |
| `spec.config` | `map<string, string>` |  |  |  |
| `spec.image` | `string` |  |  |  |
| `spec.plugins` | `[]KubernetesKafkaConnectOciPlugin` |  |  |  |
| `spec.plugins[].name` | `string` | yes |  |  |
| `spec.plugins[].artifacts` | `[]KubernetesKafkaConnectOciArtifact` | yes |  |  |
| `spec.plugins[].artifacts[].reference` | `string` | yes |  |  |
| `spec.plugins[].artifacts[].pullPolicy` | `string` |  |  |  |
| `spec.build` | `KubernetesKafkaConnectBuild` |  |  |  |
| `spec.build.output` | `KubernetesKafkaConnectBuildOutput` | yes |  |  |
| `spec.build.output.type` | `string` |  | `docker` |  |
| `spec.build.output.image` | `string` | yes |  |  |
| `spec.build.output.pushSecret` | `string` |  |  |  |
| `spec.build.output.additionalBuildOptions` | `[]string` |  |  |  |
| `spec.build.output.additionalPushOptions` | `[]string` |  |  |  |
| `spec.build.plugins` | `[]KubernetesKafkaConnectBuildPlugin` | yes |  |  |
| `spec.build.plugins[].name` | `string` | yes |  |  |
| `spec.build.plugins[].artifacts` | `[]KubernetesKafkaConnectBuildArtifact` | yes |  |  |
| `spec.build.plugins[].artifacts[].type` | `string` | yes |  |  |
| `spec.build.plugins[].artifacts[].url` | `string` |  |  |  |
| `spec.build.plugins[].artifacts[].sha512sum` | `string` |  |  |  |
| `spec.build.plugins[].artifacts[].insecure` | `bool` |  |  |  |
| `spec.build.plugins[].artifacts[].fileName` | `string` |  |  |  |
| `spec.build.plugins[].artifacts[].repository` | `string` |  |  |  |
| `spec.build.plugins[].artifacts[].group` | `string` |  |  |  |
| `spec.build.plugins[].artifacts[].artifact` | `string` |  |  |  |
| `spec.build.plugins[].artifacts[].version` | `string` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.jvm` | `KubernetesKafkaConnectJvm` |  |  |  |
| `spec.jvm.xms` | `string` |  |  |  |
| `spec.jvm.xmx` | `string` |  |  |  |
| `spec.rack` | `KubernetesKafkaConnectRack` |  |  |  |
| `spec.rack.topologyKey` | `string` | yes |  |  |
| `spec.metrics` | `KubernetesKafkaConnectMetrics` |  |  |  |
| `spec.metrics.enabled` | `bool` |  |  |  |
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

Namespace for the Connect cluster. Accepts a literal namespace
name or a reference to a KubernetesNamespace resource. The
namespace must be watched by a Strimzi operator installation, and
KubernetesKafkaConnector declarations for this cluster must live
in THIS namespace.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the
resource. When false, the namespace must already exist.

### spec.version

`string`

Kafka version the Connect workers run (e.g. "4.3.0"). Empty =
the operator's default version (the newest the pinned Strimzi
release supports). Keep it aligned with the target cluster's
version during upgrades.

### spec.replicas

`int32` · optional (explicit presence)

Number of Connect worker pods. Workers share connector tasks
through the Connect group protocol — scale for task throughput.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.bootstrapServers

`string | valueFrom` · required

Bootstrap address of the Kafka cluster this Connect cluster
reads from and writes to, as host:port. Accepts a literal
address (an external cluster, Confluent, MSK) or a reference to
a KubernetesKafka resource, which resolves to its in-cluster
bootstrap endpoint.

containment_exempt: Connect TALKS TO this cluster; it is not
deployed inside it (its own home is its namespace).

- references: KubernetesKafka (`status.outputs.internal_bootstrap_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKafka, name: <that resource's name>, fieldPath: status.outputs.internal_bootstrap_endpoint}} -- a bare string does not parse

### spec.tls

`StrimziKafkaClientTls`

TLS trust for the Kafka connection. Set when the bootstrap
address is a TLS listener; for a Strimzi-managed cluster,
reference the KubernetesKafka resource to trust its cluster CA.
Omitted = plaintext connection.

### spec.tls.trustedCertificates

`[]StrimziKafkaClientTrustedCertificate` · required

Certificates to trust. For a Strimzi-managed cluster, reference
the cluster's CA certificate Secret (the KubernetesKafka
resource's `cluster_ca_cert_secret_name` output — the default
wiring below); for external clusters, name any Secret in the
consumer's namespace holding the PEM certificate(s).

- rule: {"repeated":{"minItems":"1"}}
- rule: set exactly one of certificate (a single file name in the Secret, e.g. "ca.crt") or pattern (a glob over the Secret's files, e.g. "*.crt")

### spec.tls.trustedCertificates[].secretName

`string | valueFrom` · required

Name of the Secret (in the consuming resource's namespace)
holding the certificate. Accepts a literal Secret name or a
reference to a KubernetesKafka resource, which resolves to that
cluster's CA certificate Secret — the common wiring when the
target is a Strimzi-managed cluster in the same namespace.

containment_exempt: trust material fetched FROM the cluster —
access, never placement.

- references: KubernetesKafka (`status.outputs.cluster_ca_cert_secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKafka, name: <that resource's name>, fieldPath: status.outputs.cluster_ca_cert_secret_name}} -- a bare string does not parse

### spec.tls.trustedCertificates[].certificate

`string`

The certificate file name within the Secret. Strimzi cluster CA
Secrets carry "ca.crt". Exactly one of certificate or pattern
must be set.

### spec.tls.trustedCertificates[].pattern

`string`

Glob pattern selecting certificate files within the Secret (e.g.
"*.crt") — the multi-certificate alternative to naming one file.
Exactly one of certificate or pattern must be set.

### spec.authentication

`StrimziKafkaClientAuthentication`

How the Connect workers authenticate to Kafka. Must match the
target listener's authentication type. Omitted = unauthenticated
(only for unauthenticated listeners).

- rule: tls authentication requires certificate_and_key (the client certificate the workload presents — reference a KubernetesKafkaUser with tls authentication)
- rule: scram-sha-512, scram-sha-256 and plain authentication require username and password_secret (reference a KubernetesKafkaUser with scram-sha-512 authentication for Strimzi-managed clusters)
- rule: certificate_and_key is only used with tls authentication

### spec.authentication.type

`string` · required

Authentication type:
"tls" (mutual TLS — the client presents the certificate in
`certificate_and_key`; pairs with tls-auth listeners),
"scram-sha-512" / "scram-sha-256" (SASL username/password from
`username` + `password_secret`; Strimzi-managed clusters use
scram-sha-512),
"plain" (SASL PLAIN — username/password in the clear inside the
TLS session; for external clusters that only offer PLAIN), or
"custom" (bring-your-own SASL mechanism via `sasl` + `config`).

- rule: authentication type must be one of tls, scram-sha-512, scram-sha-256, plain, custom
- rule: {"required":true}

### spec.authentication.certificateAndKey

`StrimziKafkaClientCertificateAndKey`

tls type only: the client certificate and key the workload
presents. Reference a KubernetesKafkaUser resource (tls
authentication) to use its operator-generated credential Secret
— the default wiring below.

### spec.authentication.certificateAndKey.secretName

`string | valueFrom` · required

Name of the Secret (in the consuming resource's namespace)
holding the client certificate and key. Accepts a literal Secret
name or a reference to a KubernetesKafkaUser resource, which
resolves to that user's operator-generated credential Secret.

- references: KubernetesKafkaUser (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKafkaUser, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.authentication.certificateAndKey.certificate

`string` · optional (explicit presence)

Certificate file name within the Secret. KubernetesKafkaUser
credential Secrets carry "user.crt"; cert-manager Secrets carry
"tls.crt".

- default: `user.crt`

### spec.authentication.certificateAndKey.key

`string` · optional (explicit presence)

Private-key file name within the Secret. KubernetesKafkaUser
credential Secrets carry "user.key"; cert-manager Secrets carry
"tls.key".

- default: `user.key`

### spec.authentication.username

`string`

SASL username (scram-sha-512, scram-sha-256, plain).

### spec.authentication.passwordSecret

`StrimziKafkaClientPasswordSecret`

SASL password source (scram-sha-512, scram-sha-256, plain) — a
key within a Secret. Reference a KubernetesKafkaUser resource
(scram-sha-512 authentication) to use its operator-generated
Secret, whose password lives under the "password" key.

### spec.authentication.passwordSecret.secretName

`string | valueFrom` · required

Name of the Secret (in the consuming resource's namespace)
holding the password. Accepts a literal Secret name or a
reference to a KubernetesKafkaUser resource, which resolves to
that user's operator-generated credential Secret.

- references: KubernetesKafkaUser (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKafkaUser, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.authentication.passwordSecret.password

`string` · optional (explicit presence)

The key within the Secret whose value is the password.
KubernetesKafkaUser credential Secrets carry it under
"password".

- default: `password`

### spec.authentication.sasl

`bool`

custom type only: enable SASL for the custom mechanism.

### spec.authentication.config

`map<string, string>`

custom type only: the mechanism's client configuration entries
(sasl.mechanism, sasl.jaas.config references, callback handlers).
Values are Kafka configuration strings — write numbers and
booleans as strings.

### spec.groupId

`string`

Connect group ID — the identity workers share. Empty = the
resource's metadata.name. MUST be unique among Connect clusters
(including MirrorMaker 2 instances) sharing a Kafka cluster.

### spec.configStorageTopic

`string`

Topic storing connector configurations. Empty =
"<metadata.name>-connect-configs". Must be unique per Connect
cluster sharing a Kafka cluster.

### spec.statusStorageTopic

`string`

Topic storing connector/task status. Empty =
"<metadata.name>-connect-status".

### spec.offsetStorageTopic

`string`

Topic storing source-connector offsets. Empty =
"<metadata.name>-connect-offsets".

### spec.config

`map<string, string>`

Connect worker configuration (connect-distributed.properties
entries), e.g. "key.converter", "value.converter",
"config.storage.replication.factor". Values are configuration
strings — write numbers and booleans as strings ("3", "false");
the operator serializes every value into Java properties form.

The operator OWNS connection, identity and listener
configuration: entries with the prefixes group.id,
config.storage.topic, offset.storage.topic, status.storage.topic,
ssl., sasl., security., listeners, plugin.path, rest.,
bootstrap.servers, consumer.interceptor.classes,
producer.interceptor.classes and prometheus.metrics.reporter.
are IGNORED with an operator log warning (the
ssl.endpoint.identification.algorithm / ssl.cipher.suites /
ssl.protocol / ssl.enabled.protocols client-tuning keys are the
exception) — configure those concerns through their typed fields.

- rule: connector.plugin.version is not accepted in Connect worker config on this Strimzi line — declare the desired plugin version on each KubernetesKafkaConnector's version field instead

### spec.image

`string`

Run a PREBUILT Connect image that already carries your connector
plugins (e.g. a published Debezium Connect image or the output
of a previous `build`). Empty = the operator's stock image for
`version`. Not used when `build` is configured — the operator
runs the image it builds.

### spec.plugins

`[]KubernetesKafkaConnectOciPlugin`

Mount connector plugins from OCI ARTIFACTS as Kubernetes image
volumes — plugins ship as container images and mount directly
into the workers, no image build or registry push. REQUIRES the
Kubernetes ImageVolume feature to be enabled AND supported by
the cluster's container runtime (a cluster-level capability —
verified live: workers fail to schedule with an image-volume
admission error on clusters without it; check before relying on
this arm).

- rule: plugin names must be unique within the Connect cluster

### spec.plugins[].name

`string` · required

Unique plugin name — becomes the plugin's mount path inside the
workers. Lowercase alphanumerics, '-' and '_', starting and
ending with an alphanumeric.

- rule: plugin name must be lowercase alphanumerics, '-' and '_', starting and ending with an alphanumeric
- rule: {"required":true}

### spec.plugins[].artifacts

`[]KubernetesKafkaConnectOciArtifact` · required

The OCI artifacts carrying this plugin's binaries. Each is
mounted as an image volume.

- rule: {"repeated":{"minItems":"1"}}

### spec.plugins[].artifacts[].reference

`string` · required

Reference to the container image (OCI artifact) containing the
plugin binaries, e.g. "quay.io/example/debezium-postgres:3.1.0".

- rule: {"required":true}

### spec.plugins[].artifacts[].pullPolicy

`string` · optional (explicit presence)

Image pull policy: "Always", "Never", or "IfNotPresent". Empty =
the Kubernetes default (Always for :latest references,
IfNotPresent otherwise).

- rule: pull_policy must be Always, Never, or IfNotPresent

### spec.build

`KubernetesKafkaConnectBuild`

Have the OPERATOR build a custom Connect image containing the
declared plugin artifacts (Kaniko or Buildah on Kubernetes) and
push it to `output.image`. The workers then run the built image.

### spec.build.output

`KubernetesKafkaConnectBuildOutput` · required

Where the built image goes.

- rule: {"required":true}

### spec.build.output.type

`string` · optional (explicit presence)

Output type: "docker" (push to any Docker-compatible registry —
the Kubernetes path) or "imagestream" (push to an OpenShift
ImageStream; only meaningful on OpenShift clusters).

- default: `docker`
- rule: build output type must be docker (any Docker-compatible registry) or imagestream (OpenShift only)

### spec.build.output.image

`string` · required

Full name of the image to build and push, including the
registry, e.g. "registry.example.com/team/my-connect:latest".

- rule: {"required":true}

### spec.build.output.pushSecret

`string`

Name of a docker-registry Secret (in the Connect namespace) with
push credentials. This is the SECRET'S NAME, not a credential
value. Empty for registries the build pod can push to
anonymously or via ambient identity.

### spec.build.output.additionalBuildOptions

`[]string`

Additional options passed to the Kaniko/Buildah build command
(e.g. "--insecure" for plain-HTTP in-cluster registries — the
allowed option set is validated by the operator).

### spec.build.output.additionalPushOptions

`[]string`

Additional options passed to the Buildah push command (ignored
when Kaniko performs the build).

### spec.build.plugins

`[]KubernetesKafkaConnectBuildPlugin` · required

The plugins to bake into the image.

- rule: build plugin names must be unique
- rule: {"repeated":{"minItems":"1"}}

### spec.build.plugins[].name

`string` · required

Unique plugin name — becomes the plugin's directory inside the
image.

- rule: plugin name must be lowercase alphanumerics, '-' and '_', starting and ending with an alphanumeric
- rule: {"required":true}

### spec.build.plugins[].artifacts

`[]KubernetesKafkaConnectBuildArtifact` · required

The artifacts that make up this plugin.

- rule: {"repeated":{"minItems":"1"}}
- rule: jar, tgz, zip and other artifacts require a url to download from
- rule: maven artifacts require group, artifact and version (repository is optional — Maven Central is the default)
- rule: maven artifacts are resolved from their coordinates — url is only for jar, tgz, zip and other types

### spec.build.plugins[].artifacts[].type

`string` · required

Artifact type: "jar" (a single JAR from a URL), "tgz" / "zip"
(an archive from a URL, unpacked into the plugin directory),
"maven" (a Maven coordinate resolved at build time), or "other"
(any file from a URL, stored as file_name).

- rule: artifact type must be one of jar, tgz, zip, maven, other
- rule: {"required":true}

### spec.build.plugins[].artifacts[].url

`string`

Download URL (jar, tgz, zip, other types).

### spec.build.plugins[].artifacts[].sha512sum

`string`

SHA-512 checksum of the downloaded artifact — strongly
recommended for URL artifacts so a tampered download fails the
build instead of running in the workers.

### spec.build.plugins[].artifacts[].insecure

`bool`

Allow insecure (plain-HTTP / unverified-TLS) download of this
artifact. Leave false outside isolated dev environments.

### spec.build.plugins[].artifacts[].fileName

`string`

other type only: the file name to store the artifact under.

### spec.build.plugins[].artifacts[].repository

`string`

maven type only: repository URL. Empty = Maven Central.

### spec.build.plugins[].artifacts[].group

`string`

maven type only: group ID, e.g. "io.debezium".

### spec.build.plugins[].artifacts[].artifact

`string`

maven type only: artifact ID, e.g.
"debezium-connector-postgres".

### spec.build.plugins[].artifacts[].version

`string`

maven type only: version, e.g. "3.1.0.Final".

### spec.resources

`ContainerResources`

CPU/memory for each worker pod. Empty = no requests/limits
(fine for kind/dev; always set for production — the JVM heap
default derives from the memory limit).

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

### spec.jvm

`KubernetesKafkaConnectJvm`

JVM heap for the workers (rendered as -Xms/-Xmx). Empty =
Strimzi's dynamic default. Set both to the SAME value for
production.

### spec.jvm.xms

`string`

Initial heap (-Xms), e.g. "1g". Set equal to xmx in production.

### spec.jvm.xmx

`string`

Maximum heap (-Xmx), e.g. "1g".

### spec.rack

`KubernetesKafkaConnectRack`

Rack awareness: injects each worker's failure domain from its
node's label value (e.g. "topology.kubernetes.io/zone") so
consumers fetch from the closest replica. Requires nodes labeled
with the key.

### spec.rack.topologyKey

`string` · required

Node label whose value identifies a node's failure domain,
e.g. "topology.kubernetes.io/zone".

- rule: {"required":true}

### spec.metrics

`KubernetesKafkaConnectMetrics`

JMX Prometheus metrics. When enabled, the module renders the
canonical Strimzi connect-metrics ConfigMap and wires it as the
cluster's metricsConfig (port 9404 inside the pods).

### spec.metrics.enabled

`bool`

Render the canonical Strimzi JMX exporter rules ConfigMap and
enable the metrics endpoint on every worker.

### spec.nodeSelector

`map<string, string>`

Node selector for the worker pods.

### spec.tolerations

`[]WorkloadToleration`

Tolerations for the worker pods.

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

## Validation Rules

- `spec.image_xor_build`: image and build cannot both be set — when build is configured the operator deploys the image IT builds and a declared image is silently overridden (verified in the operator source); set image for a prebuilt-plugins image OR build to have one built

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesKafkaConnect, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the Connect cluster runs in. |
| `status.outputs.connect_name` | `string` | The Connect cluster's name (metadata.name) — the value KubernetesKafkaConnector resources bind to via their connect_cluster field (rendered as the strimzi.io/cluster label). |
| `status.outputs.rest_api_service_name` | `string` | Name of the Connect REST API Service (`<name>-connect-api`). |
| `status.outputs.rest_api_endpoint` | `string` | In-cluster Connect REST API endpoint (`http://<name>-connect-api.<namespace>.svc.cluster.local:8083`) — read-only inspection (connector status, plugin listing); connector management is declarative through KubernetesKafkaConnector. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.bootstrapServers` | KubernetesKafka | `status.outputs.internal_bootstrap_endpoint` |
| `spec.tls.trustedCertificates[].secretName` | KubernetesKafka | `status.outputs.cluster_ca_cert_secret_name` |
| `spec.authentication.certificateAndKey.secretName` | KubernetesKafkaUser | `status.outputs.secret_name` |
| `spec.authentication.passwordSecret.secretName` | KubernetesKafkaUser | `status.outputs.secret_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesKafkaConnector | `spec.connectCluster` | `status.outputs.connect_name` |
| KubernetesKafkaUi | `spec.clusters[].kafkaConnect[].address` | `status.outputs.rest_api_endpoint` |

## See Also

- [Overview](../README.md)
