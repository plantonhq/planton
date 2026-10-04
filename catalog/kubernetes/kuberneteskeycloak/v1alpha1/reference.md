# KubernetesKeycloak

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesKeycloakSpec** declares a Keycloak server — the
Kubernetes CR (`keycloaks.k8s.keycloak.org/v2beta1`) that the
official Keycloak Operator reconciles into a StatefulSet.
PREREQUISITE: a KubernetesKeycloakOperator watching this
namespace (with the default namespaced watch, the operator and its
Keycloak declarations live in the SAME namespace).

TLS-OR-HTTP, DECIDED UP FRONT (verified at Keycloak 26.7): the
server REFUSES TO START with neither `http.tls_secret_name` nor
`http.http_enabled: true` — upstream surfaces this only as a
CrashLoopBackOff. This spec makes the choice a validation rule.
Likewise HOSTNAME: with strict hostname resolution (the server
default) a hostname is mandatory; disable strict only behind a
reverse proxy that rewrites Host headers (and then set
proxy_headers, or browsers fail CORS with 403s on computed-origin
mismatches — the classic "login randomly breaks" misconfiguration).

WHAT THE OPERATOR CREATES from this declaration: the StatefulSet
(named exactly after this resource), `<name>-service` (https 8443 /
http 8080 + management 9000), `<name>-discovery` (headless, JGroups
clustering), a NetworkPolicy by default, and — unless you bring
your own bootstrap-admin Secret — the one-time
`<name>-initial-admin` credential Secret (create-once, never
rotated by the operator; exported as the credential handle). The
operator's own default Ingress is ALWAYS disabled by this module
(exposure composes from Gateway API kinds referencing the exported
service handles).

REALM IMPORTS, CLIENTS: the operator's KeycloakRealmImport CR is a
ONE-SHOT import Job (edits after a successful import are silently
ignored upstream) and the OIDC/SAML client CRs are alpha,
experimental-gated surfaces — none are modeled here; manage realms
and clients through Keycloak's admin API/console, or declare the
CRs via KubernetesManifest at your own risk.

NAME BUDGET: keep this resource's name at 48 characters or fewer —
the operator derives child names by suffixing (`-network-policy`
is the longest at 15) and StatefulSet pod hostnames must stay
DNS-legal. Both modules fail loudly past the budget.

## Example

```yaml
# Full-surface hack manifest for the offline plan/preview proofs: every
# module-rendered arm of the Keycloak CR expressed at once — postgres
# vendor with Secret-selector credentials, TLS + plain HTTP listeners,
# hostname posture, features, additional options (inline AND secret
# arms), Explicit update, tracing, scheduling and probes — while keeping
# the spec's validation rules satisfied (TLS-or-HTTP, hostname-when-
# strict, real-vendor connection details, Explicit-needs-revision).
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKeycloak
metadata:
  name: keycloak-full
spec:
  namespace:
    value: identity
  createNamespace: true
  instances: 2
  image: quay.io/keycloak/keycloak:26.7.0
  startOptimized: true
  db:
    vendor: postgres
    host:
      value: identity-pg-rw
    port: 5432
    database: keycloak
    schema: public
    usernameSecret:
      name:
        value: identity-pg-app
      key: username
    passwordSecret:
      name:
        value: identity-pg-app
      key: password
    poolMinSize: 5
    poolMaxSize: 50
  http:
    tlsSecretName:
      value: keycloak-full-tls
    httpEnabled: true
    httpPort: 8080
    httpsPort: 8443
  hostname:
    hostname: https://auth.example.com
    admin: https://auth-admin.example.com
    strict: true
    backchannelDynamic: true
  proxyHeaders: xforwarded
  features:
    enabled:
      - token-exchange
      - recovery-codes
    disabled:
      - impersonation
  transactionXaEnabled: true
  cacheConfig:
    configMapName: keycloak-cache-ispn
    key: cache-ispn.xml
  truststoreSecretNames:
    - private-ca-bundle
    - ldap-ca
  additionalOptions:
    - name: log-level
      value: info
    - name: https-client-auth
      secret:
        name:
          value: keycloak-tuning
        key: clientAuth
  bootstrapAdminSecretName: keycloak-bootstrap-admin
  resources:
    requests:
      cpu: 250m
      memory: 768Mi
    limits:
      cpu: 1000m
      memory: 1Gi
  scheduling:
    nodeSelector:
      workload: identity
    tolerations:
      - key: identity
        operator: Equal
        value: "true"
        effect: NoSchedule
      - key: maintenance
        operator: Exists
        effect: NoExecute
        tolerationSeconds: 300
    priorityClassName: platform-critical
  probes:
    livenessFailureThreshold: 5
    livenessPeriodSeconds: 15
    readinessFailureThreshold: 5
    readinessPeriodSeconds: 15
    startupFailureThreshold: 900
    startupPeriodSeconds: 1
  httpManagementPort: 9000
  networkPolicyEnabled: false
  serviceMonitorEnabled: false
  update:
    strategy: Explicit
    revision: "1"
  tracing:
    enabled: true
    endpoint: http://otel-collector.observability:4317
    protocol: grpc
    samplerRatio: "0.25"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.instances` | `int32` |  | `1` |  |
| `spec.image` | `string` |  |  |  |
| `spec.startOptimized` | `bool` |  |  |  |
| `spec.db` | `KubernetesKeycloakDb` | yes |  |  |
| `spec.db.vendor` | `string` | yes |  |  |
| `spec.db.host` | `string \| valueFrom` |  |  | KubernetesPostgres (`status.outputs.rw_service`) |
| `spec.db.port` | `int32` |  |  |  |
| `spec.db.database` | `string` |  |  |  |
| `spec.db.usernameSecret` | `KubernetesKeycloakSecretSelector` |  |  |  |
| `spec.db.usernameSecret.name` | `string \| valueFrom` | yes |  | KubernetesPostgres (`status.outputs.password_secret.name`) |
| `spec.db.usernameSecret.key` | `string` | yes |  |  |
| `spec.db.passwordSecret` | `KubernetesKeycloakSecretSelector` |  |  |  |
| `spec.db.passwordSecret.name` | `string \| valueFrom` | yes |  | KubernetesPostgres (`status.outputs.password_secret.name`) |
| `spec.db.passwordSecret.key` | `string` | yes |  |  |
| `spec.db.schema` | `string` |  |  |  |
| `spec.db.jdbcUrl` | `string` |  |  |  |
| `spec.db.poolMinSize` | `int32` |  |  |  |
| `spec.db.poolMaxSize` | `int32` |  |  |  |
| `spec.http` | `KubernetesKeycloakHttp` | yes |  |  |
| `spec.http.tlsSecretName` | `string \| valueFrom` |  |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.http.httpEnabled` | `bool` |  |  |  |
| `spec.http.httpPort` | `int32` |  | `8080` |  |
| `spec.http.httpsPort` | `int32` |  | `8443` |  |
| `spec.hostname` | `KubernetesKeycloakHostname` | yes |  |  |
| `spec.hostname.hostname` | `string` |  |  |  |
| `spec.hostname.admin` | `string` |  |  |  |
| `spec.hostname.strict` | `bool` |  | `true` |  |
| `spec.hostname.backchannelDynamic` | `bool` |  |  |  |
| `spec.proxyHeaders` | `string` |  |  |  |
| `spec.features` | `KubernetesKeycloakFeatures` |  |  |  |
| `spec.features.enabled` | `[]string` |  |  |  |
| `spec.features.disabled` | `[]string` |  |  |  |
| `spec.transactionXaEnabled` | `bool` |  |  |  |
| `spec.cacheConfig` | `KubernetesKeycloakCacheConfig` |  |  |  |
| `spec.cacheConfig.configMapName` | `string` | yes |  |  |
| `spec.cacheConfig.key` | `string` |  | `cache-ispn.xml` |  |
| `spec.truststoreSecretNames` | `[]string` |  |  |  |
| `spec.additionalOptions` | `[]KubernetesKeycloakAdditionalOption` |  |  |  |
| `spec.additionalOptions[].name` | `string` | yes |  |  |
| `spec.additionalOptions[].value` | `string` |  |  |  |
| `spec.additionalOptions[].secret` | `KubernetesKeycloakOptionSecretSelector` |  |  |  |
| `spec.additionalOptions[].secret.name` | `string \| valueFrom` | yes |  |  |
| `spec.additionalOptions[].secret.key` | `string` | yes |  |  |
| `spec.bootstrapAdminSecretName` | `string` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.scheduling` | `KubernetesKeycloakScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.scheduling.priorityClassName` | `string` |  |  |  |
| `spec.probes` | `KubernetesKeycloakProbes` |  |  |  |
| `spec.probes.livenessFailureThreshold` | `int32` |  |  |  |
| `spec.probes.livenessPeriodSeconds` | `int32` |  |  |  |
| `spec.probes.readinessFailureThreshold` | `int32` |  |  |  |
| `spec.probes.readinessPeriodSeconds` | `int32` |  |  |  |
| `spec.probes.startupFailureThreshold` | `int32` |  |  |  |
| `spec.probes.startupPeriodSeconds` | `int32` |  |  |  |
| `spec.httpManagementPort` | `int32` |  | `9000` |  |
| `spec.networkPolicyEnabled` | `bool` |  | `true` |  |
| `spec.serviceMonitorEnabled` | `bool` |  | `true` |  |
| `spec.update` | `KubernetesKeycloakUpdate` |  |  |  |
| `spec.update.strategy` | `string` |  | `RecreateOnImageChange` |  |
| `spec.update.revision` | `string` |  |  |  |
| `spec.tracing` | `KubernetesKeycloakTracing` |  |  |  |
| `spec.tracing.enabled` | `bool` |  |  |  |
| `spec.tracing.endpoint` | `string` |  |  |  |
| `spec.tracing.protocol` | `string` |  | `grpc` |  |
| `spec.tracing.samplerRatio` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to deploy into. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource. Must be watched by
the KubernetesKeycloakOperator (the operator's own namespace
under the default namespaced watch). NOTE the db credential
Secrets ride secretKeyRefs, readable only from this same
namespace — co-locate Keycloak with its database or replicate
the Secret.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before the CR and deleted with the resource.
When false, the namespace must already exist.

### spec.instances

`int32` · optional (explicit presence)

Server replicas. All state lives in the database, so instances
scale horizontally (JGroups/Infinispan clusters the caches
through the discovery Service). Requires a real database vendor
— the dev-file/dev-mem sandbox vendors force 1 (each pod would
hold its own divergent world on ephemeral storage).

- default: `1`
- rule: {"int32":{"lte":20,"gte":1}}

### spec.image

`string`

Custom Keycloak server image (empty = the operator's pinned
default, quay.io/keycloak/keycloak at the operator release).
Pre-augmented images pair with start_optimized. KNOW THIS: with
the default update strategy, CHANGING the image triggers a full
scale-to-zero recreate (two Keycloak versions cannot share one
cache cluster/schema) — an outage window by design.

### spec.startOptimized

`bool`

Start with `--optimized`, skipping the build phase — requires a
pre-augmented image whose build-time options (db vendor, enabled
features) match this spec exactly.

### spec.db

`KubernetesKeycloakDb` · required

The database — REQUIRED. Keycloak without an explicit database
silently runs embedded H2 on ephemeral pod storage and loses
everything on restart; this spec forces the decision instead.

- rule: {"required":true}
- rule: A real database vendor needs the connection details: host (or a KubernetesPostgres reference), database, username_secret and password_secret — or a full jdbc_url.

### spec.db.vendor

`string` · required

Database vendor. "postgres" is the recommended production path
(a KubernetesPostgres composes naturally); mysql/mariadb/tidb/
mssql/oracle connect to databases you operate; "dev-file" /
"dev-mem" are Keycloak's embedded H2 SANDBOX modes — data on
ephemeral pod storage (or memory), lost on restart, single
instance only. NEVER production.

- rule: Database vendor must be one of: postgres, mysql, mariadb, tidb, mssql, oracle — or the never-production sandbox vendors dev-file / dev-mem.
- rule: {"required":true}

### spec.db.host

`string | valueFrom`

Database host — a Service name (same namespace) or full FQDN.
Accepts a literal or a reference to a KubernetesPostgres
resource (its read-write Service — always the current primary).
Required for real vendors; ignored by the sandbox vendors.

- references: KubernetesPostgres (`status.outputs.rw_service`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.rw_service}} -- a bare string does not parse

### spec.db.port

`int32` · optional (explicit presence)

Database port. Empty = the vendor's default (5432 postgres,
3306 mysql/mariadb/tidb, 1433 mssql, 1521 oracle).

- rule: {"int32":{"lte":65535,"gt":0}}

### spec.db.database

`string`

Database name (on a KubernetesPostgres: declare it at bootstrap
via initdb, e.g. "keycloak").

### spec.db.usernameSecret

`KubernetesKeycloakSecretSelector`

Username, read from a Secret (on a KubernetesPostgres: the app
credential Secret's `username` key).

### spec.db.usernameSecret.name

`string | valueFrom` · required

Secret name. Accepts a literal or a reference to a
KubernetesPostgres resource (its `<cluster>-app` credential
Secret, maintained by the operator across failovers).

- references: KubernetesPostgres (`status.outputs.password_secret.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.password_secret.name}} -- a bare string does not parse

### spec.db.usernameSecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.db.passwordSecret

`KubernetesKeycloakSecretSelector`

Password, read from a Secret (on a KubernetesPostgres: the app
credential Secret's `password` key).

### spec.db.passwordSecret.name

`string | valueFrom` · required

Secret name. Accepts a literal or a reference to a
KubernetesPostgres resource (its `<cluster>-app` credential
Secret, maintained by the operator across failovers).

- references: KubernetesPostgres (`status.outputs.password_secret.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.password_secret.name}} -- a bare string does not parse

### spec.db.passwordSecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.db.schema

`string`

Database schema to use (empty = the vendor default, e.g.
"public" on postgres).

### spec.db.jdbcUrl

`string`

Full JDBC URL override (advanced — vendor-specific parameters,
multi-host failover URLs). When set, host/port/database are
ignored by the server.

### spec.db.poolMinSize

`int32` · optional (explicit presence)

Connection pool floor. Empty = the server default.

- rule: {"int32":{"gte":0}}

### spec.db.poolMaxSize

`int32` · optional (explicit presence)

Connection pool ceiling. Empty = the server default (100).

- rule: {"int32":{"gt":0}}

### spec.http

`KubernetesKeycloakHttp` · required

Listener/TLS posture — REQUIRED (see the spec-level TLS-or-HTTP
rule).

- rule: {"required":true}

### spec.http.tlsSecretName

`string | valueFrom`

Name of a kubernetes.io/tls Secret (keys tls.crt/tls.key) for
the HTTPS listener. Accepts a literal name or a reference to a
KubernetesCertificate resource (cert-manager) — the recommended
posture. The operator mounts it as REQUIRED: a missing Secret
blocks pod scheduling.

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.http.httpEnabled

`bool`

Serve plain HTTP (port 8080). Off by upstream default. The
legitimate use is behind a TLS-terminating proxy — pair it with
proxy_headers.

### spec.http.httpPort

`int32` · optional (explicit presence)

HTTP listener port. Empty = 8080.

- default: `8080`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.http.httpsPort

`int32` · optional (explicit presence)

HTTPS listener port. Empty = 8443.

- default: `8443`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.hostname

`KubernetesKeycloakHostname` · required

Hostname posture — REQUIRED (see the spec-level rule).

- rule: {"required":true}
- rule: backchannel_dynamic requires hostname to be a FULL URL ("https://..." or "http://...") — the server refuses to start otherwise (verified live: a bare hostname or no hostname crash-loops with "hostname-backchannel-dynamic must be set to false").
- rule: Setting hostname.admin requires hostname to be a FULL URL ("https://..." or "http://...") — the server refuses to start otherwise ("hostname must be set to a URL when hostname-admin is set").

### spec.hostname.hostname

`string`

The server's public base URL or hostname
(e.g. "https://auth.example.com" — full URLs pin scheme and
path). What tokens, redirects, and the OIDC discovery document
advertise. NOTE two server rules require the FULL-URL form
(scheme included), not a bare hostname: backchannel_dynamic and
admin (both validated here — the server otherwise refuses to
start, which on Kubernetes surfaces only as CrashLoopBackOff).

### spec.hostname.admin

`string`

Separate base URL for the admin console (empty = same as
hostname). Lets the admin surface live on an internal name
while the user-facing hostname stays public. Requires hostname
to be a full URL (validated here — a server-startup rule).

### spec.hostname.strict

`bool` · optional (explicit presence)

Strict hostname resolution. Empty = true (the server default):
the server uses only the declared hostname. Set false ONLY
behind a trusted reverse proxy that rewrites Host headers — and
then set proxy_headers too. When hostname is set, the server
ignores this flag entirely.

- default: `true`

### spec.hostname.backchannelDynamic

`bool`

Resolve BACKCHANNEL (server-to-server) URLs dynamically from
request headers while the public hostname stays fixed — for
clusters where internal clients reach Keycloak through the
Service instead of the public URL. Requires hostname as a FULL
URL (validated here): the server refuses to start with a bare
or absent hostname, because a partly-dynamic frontend could
leak backend request parts into frontend URLs (verified live —
the failure mode is a CrashLoopBackOff, not an error message).

### spec.proxyHeaders

`string`

Which proxy headers to trust when running behind a reverse
proxy/ingress: "xforwarded" (X-Forwarded-*) or "forwarded"
(RFC 7239). Empty = trust none. REQUIRED in practice whenever
TLS terminates in front of Keycloak — without it the server
computes wrong origins and browsers fail CORS.

- rule: proxy_headers must be "xforwarded", "forwarded", or empty (no proxy).

### spec.features

`KubernetesKeycloakFeatures`

Feature flags: entries for `--features` (enable) and
`--features-disabled`. Names as the Keycloak release documents
them (e.g. "token-exchange", "docker").

### spec.features.enabled

`[]string`

Features to enable (`--features`), e.g. "token-exchange",
"recovery-codes". Versioned syntax ("feature:v2") is accepted.

### spec.features.disabled

`[]string`

Features to disable (`--features-disabled`).

### spec.transactionXaEnabled

`bool`

Enable XA distributed transactions (two-phase commit against the
database).

### spec.cacheConfig

`KubernetesKeycloakCacheConfig`

Custom Infinispan cache topology from a ConfigMap (advanced —
the operator's default cache config suits most installs).

### spec.cacheConfig.configMapName

`string` · required

ConfigMap name carrying the Infinispan XML.

- rule: {"required":true}

### spec.cacheConfig.key

`string` · optional (explicit presence)

Key within the ConfigMap. Empty = "cache-ispn.xml".

- default: `cache-ispn.xml`

### spec.truststoreSecretNames

`[]string`

Names of Secrets carrying additional CA certificates the server
should trust (each mounted into the server truststore) — for
private-CA databases, LDAP servers, or identity providers.

### spec.additionalOptions

`[]KubernetesKeycloakAdditionalOption`

Any Keycloak server option not modeled above, by its
configuration key (e.g. "log-level", "spi-*"), with the value
inline or from a Secret. This is the config escape hatch — the
full `kc.sh` option surface.

- rule: Give the option's value exactly one way — inline OR from a Secret, not both.

### spec.additionalOptions[].name

`string` · required

The Keycloak option key (as documented for kc.sh / KC_* env,
e.g. "log-level").

- rule: {"required":true}

### spec.additionalOptions[].value

`string`

Inline value. Use secret for credential-bearing options.

### spec.additionalOptions[].secret

`KubernetesKeycloakOptionSecretSelector`

Value from a Secret (mutually exclusive with value).

### spec.additionalOptions[].secret.name

`string | valueFrom` · required

Secret name. Same-namespace constraint (a Kubernetes rule, not
an operator one): the Secret must live in the namespace Keycloak
installs into.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.additionalOptions[].secret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.bootstrapAdminSecretName

`string`

Bring-your-own bootstrap admin: name of an existing
kubernetes.io/basic-auth Secret (keys `username`/`password`).
Empty = the operator generates `<name>-initial-admin` (username
"temp-admin") — exported as the credential handle either way.
"Bootstrap" is literal: it seeds the first admin at FIRST start
and changes have no effect afterwards; manage real admin users
inside Keycloak.

### spec.resources

`ContainerResources`

CPU and memory for the server container. Keycloak is a JVM:
give it real memory (these defaults suit labs; production
typically runs 1-2Gi).

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

### spec.scheduling

`KubernetesKeycloakScheduling`

Pod scheduling constraints. When unset, the operator injects
soft zone/hostname topology-spread constraints on its own —
a sensible default worth knowing about.

### spec.scheduling.nodeSelector

`map<string, string>`

Schedule onto nodes carrying these labels (translated to
required node affinity — the Keycloak CR models affinity, not
nodeSelector).

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for tainted nodes.

### spec.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.scheduling.priorityClassName

`string`

PriorityClass for the server pods.

### spec.probes

`KubernetesKeycloakProbes`

Probe tuning. The operator's startup default gives Keycloak a
full 10 minutes (600 × 1s) before the first kill — first boots
run schema migrations; do not tighten this reflexively.

### spec.probes.livenessFailureThreshold

`int32` · optional (explicit presence)

Liveness: failures before restart. Empty = 3.

- rule: {"int32":{"gt":0}}

### spec.probes.livenessPeriodSeconds

`int32` · optional (explicit presence)

Liveness probe period seconds. Empty = 10.

- rule: {"int32":{"gt":0}}

### spec.probes.readinessFailureThreshold

`int32` · optional (explicit presence)

Readiness: failures before unready. Empty = 3.

- rule: {"int32":{"gt":0}}

### spec.probes.readinessPeriodSeconds

`int32` · optional (explicit presence)

Readiness probe period seconds. Empty = 10.

- rule: {"int32":{"gt":0}}

### spec.probes.startupFailureThreshold

`int32` · optional (explicit presence)

Startup: failures before the first kill. Empty = 600 (with the
1s period: a 10-minute first-boot budget — schema migrations
are slow; tighten with care).

- rule: {"int32":{"gt":0}}

### spec.probes.startupPeriodSeconds

`int32` · optional (explicit presence)

Startup probe period seconds. Empty = 1.

- rule: {"int32":{"gt":0}}

### spec.httpManagementPort

`int32` · optional (explicit presence)

Management (health/metrics) port. Empty = 9000.

- default: `9000`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.networkPolicyEnabled

`bool` · optional (explicit presence)

The operator's NetworkPolicy for the server pods. Empty = true —
the OPERATOR default; disable when the cluster manages network
policy through first-class KubernetesNetworkPolicy resources.

- default: `true`

### spec.serviceMonitorEnabled

`bool` · optional (explicit presence)

The operator's ServiceMonitor for the management endpoint.
Empty = true (the operator default). Without the Prometheus
Operator CRDs on the cluster the operator records a warning on
the CR and carries on — safe either way.

- default: `true`

### spec.update

`KubernetesKeycloakUpdate`

How spec changes roll out.

- rule: The Explicit update strategy is driven by the revision field — set one (any string; change it to trigger the recreate path).

### spec.update.strategy

`string` · optional (explicit presence)

"RecreateOnImageChange" (empty default): image changes take a
full scale-to-zero outage (safe — two versions never share a
cluster), everything else rolls. "Auto": the operator runs a
compatibility-check Job per change and recreates only when
needed (adds up to ~5 minutes per spec change). "Explicit": YOU
decide — bumping `revision` forces the recreate.

- default: `RecreateOnImageChange`
- rule: Update strategy must be one of: RecreateOnImageChange, Auto, Explicit.

### spec.update.revision

`string`

Revision marker for the Explicit strategy — change it to signal
"this update needs the recreate path".

### spec.tracing

`KubernetesKeycloakTracing`

OpenTelemetry tracing export.

- rule: Tracing needs somewhere to send spans — set endpoint (an OTLP collector URL) when tracing is enabled.

### spec.tracing.enabled

`bool`

Enable trace export.

### spec.tracing.endpoint

`string`

OTLP endpoint URL (e.g. "http://collector.observability:4317").
Required when enabled.

### spec.tracing.protocol

`string` · optional (explicit presence)

OTLP protocol: "grpc" (empty default) or "http/protobuf".

- default: `grpc`
- rule: Tracing protocol must be "grpc" or "http/protobuf".

### spec.tracing.samplerRatio

`string` · optional (explicit presence)

Head sampling ratio 0.0–1.0. Empty = the server default.

- rule: {"string":{"pattern":"^(0(\\.\\d+)?|1(\\.0+)?)$"}}

## Validation Rules

- `spec.http.tls_or_http`: Keycloak refuses to start with neither TLS nor HTTP: set http.tls_secret_name (a kubernetes.io/tls Secret or KubernetesCertificate reference — recommended) or opt into plain HTTP with http.http_enabled (only behind a TLS-terminating proxy).
- `spec.hostname.required_when_strict`: With strict hostname resolution (the default), Keycloak needs its public hostname: set hostname.hostname (e.g. "https://auth.example.com"), or set hostname.strict to false when a trusted reverse proxy rewrites Host headers.
- `spec.db.dev_vendors_single_instance`: The dev-file/dev-mem sandbox databases live on each pod's own ephemeral storage — more than one instance means divergent servers behind one Service. Use a real database vendor to scale out.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesKeycloak, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the server runs in. |
| `status.outputs.stateful_set` | `string` | The StatefulSet name (exactly this resource's name). |
| `status.outputs.service` | `string` | The main Service name (`<name>-service`) — https 8443 and/or http 8080, plus the management port. |
| `status.outputs.discovery_service` | `string` | The headless discovery Service (`<name>-discovery`) — JGroups cluster formation between instances. |
| `status.outputs.api_endpoint` | `string` | In-cluster API endpoint, scheme included (e.g. "https://sso-service.identity.svc.cluster.local:8443", or http/8080 when only plain HTTP is enabled) — where OIDC clients inside the cluster reach the server. |
| `status.outputs.management_endpoint` | `string` | The management endpoint (e.g. "https://sso-service.identity.svc.cluster.local:9000") — health probes and metrics. |
| `status.outputs.initial_admin_secret_name` | `string` | Name of the bootstrap-admin credential Secret: the operator-generated `<name>-initial-admin` (kubernetes.io/basic-auth, keys username/password, username "temp-admin"), or the user-provided Secret when bootstrap_admin_secret_name is set. Seeds the FIRST admin login; create durable admin users inside Keycloak and treat this as break-glass material. |
| `status.outputs.port_forward_command` | `string` | Copy-paste command for reaching the server from a workstation. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.db.host` | KubernetesPostgres | `status.outputs.rw_service` |
| `spec.db.usernameSecret.name` | KubernetesPostgres | `status.outputs.password_secret.name` |
| `spec.db.passwordSecret.name` | KubernetesPostgres | `status.outputs.password_secret.name` |
| `spec.http.tlsSecretName` | KubernetesCertificate | `status.outputs.secret_name` |

## See Also

- [Overview](../README.md)
