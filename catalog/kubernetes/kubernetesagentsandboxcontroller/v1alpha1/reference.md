# KubernetesAgentSandboxController

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesAgentSandboxControllerSpec** installs agent-sandbox, the
Kubernetes SIG Apps controller that runs isolated, stateful, singleton
workloads for AI agents: a `Sandbox` is one pod with a stable identity,
optional persistent volumes and an optional headless Service, which the
controller creates, suspends (pod deleted, volumes kept) and expires.
With extensions on, it also serves `SandboxTemplate` (a reusable,
network-policed blueprint), `SandboxWarmPool` (pre-started sandboxes kept
ready) and `SandboxClaim` (an application's request for a sandbox, bound
to a warm member when one is ready). The optional router is the HTTP
entry point upstream's SDKs use to reach a sandbox by name, and the only
access path when sandboxes run under gVisor or Kata, where port-forward
does not work.

DISTRIBUTION AND VERSION: the modules install agent-sandbox v1.0.7 from
upstream's own release manifest (`sandbox.yaml`, or
`sandbox-with-extensions.yaml` when extensions are on) and the router from
upstream's deployment files at the same tag, and apply the typed settings
below onto upstream's own arguments. The spec deliberately has NO version
field: the catalog's agent sandbox kinds (KubernetesAgentSandbox and its
template, warm pool and claim) mirror these resource definitions field
for field, and an independently chosen version would let the definitions
drift from the typed resources built on them. Each upstream release is a
deliberate catalog update.

NAMESPACE IS FIXED: everything installs into `agent-sandbox-system`, and
exactly one install per cluster is the upstream contract. The namespace is
baked into upstream's manifests, and the default network policy every
SandboxTemplate gets admits traffic only from a router in this namespace,
which is why the router is part of this kind rather than its own. Who
owns the namespace is still a choice (`namespace`, `create_namespace`):
the Secrets this kind reads (router TLS, token keys, pull secrets) must be
in it before the controller starts.

RESOURCE DEFINITIONS ARE KEPT: the four definitions are applied as kept,
ownership-checked resources (`crds`), so destroying the controller never
deletes a sandbox, its pod or its volume. The kind refuses to install over
definitions or a controller it did not create.

GKE: GKE's managed Agent Sandbox add-on (`GcpGkeCluster` addon
`agent_sandbox_enabled`) installs the same controller and definitions in
the same namespace. Run the add-on or this kind on a cluster, never both;
this kind refuses to install over the add-on's definitions.

Built for 100% parity with every setting upstream's controller and router
offer; the guide accounts for each one.

## Example

```yaml
# Full-surface development manifest: every typed field set, so the offline
# plan proofs cover arms the live lanes leave out. Values differ from the
# defaults wherever a default exists, so every setting is visible in the
# rendered objects.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAgentSandboxController
metadata:
  name: agent-sandbox-hack
spec:
  namespace:
    value: agent-sandbox-system
  createNamespace: true
  crds:
    install: true
    keepOnUninstall: false
  clusterDomain: cluster.local
  imageRegistry: mirror.example.com/k8s
  imagePullSecrets:
    - mirror-pull
  extensions:
    allowedLabelDomains:
      - sandbox.users.io
      - example.com
    claim:
      concurrentWorkers: 80
      warmCandidateGracePeriod: 500ms
      eventsDisabled: true
      observabilityAnnotationsDisabled: true
    warmPool:
      concurrentWorkers: 4
      maxBatchSize: 100
      replenishDelay: 2s
      maxRefillRate: 20
      readinessGracePeriod: 3m
      unschedulableRecheckInterval: 30s
      evictionEnabled: false
    template:
      concurrentWorkers: 2
  controller:
    image:
      repo: mirror.example.com/k8s/agent-sandbox/agent-sandbox-controller
      tag: v1.0.7
    replicas: 2
    resources:
      requests:
        cpu: 200m
        memory: 256Mi
      limits:
        cpu: "1"
        memory: 1Gi
    nodeSelector:
      kubernetes.io/os: linux
    tolerations:
      - key: dedicated
        operator: Equal
        value: agents
        effect: NoSchedule
    leaderElection:
      namespace: agent-sandbox-system
    metrics:
      port: 8443
      restClientMetricsEnabled: false
      tls:
        certificateSecret:
          value: controller-metrics-tls
        minVersion: tls13
    healthProbePort: 8082
    sandbox:
      concurrentWorkers: 200
      eventsDisabled: true
      writeBehindWindow: 250ms
    cacheLabelSelectors: true
    kubeApi:
      qps: 80
      burst: 120
      connections: 2
      separateWatchConnection: true
    tracing:
      otlpEndpoint: http://otel-collector.observability:4317
      insecure: true
      headers:
        secret:
          value: otel-headers
        key: headers
    profiling:
      cpuEnabled: true
      debugEnabled: true
      blockProfileRate: 500000
      mutexProfileFraction: 20
    logging:
      development: true
      encoder: json
      level: debug
      stacktraceLevel: panic
      timeEncoding: rfc3339
  router:
    image:
      repo: mirror.example.com/k8s/agent-sandbox/sandbox-router-go
      tag: v1.0.7
    replicas: 3
    resources:
      requests:
        cpu: 500m
        memory: 512Mi
      limits:
        cpu: "2"
        memory: 1Gi
    nodeSelector:
      kubernetes.io/os: linux
    tolerations:
      - key: dedicated
        operator: Equal
        value: agents
        effect: NoSchedule
    minAvailable: 2
    authorization:
      scopedToken:
        v1Secret:
          secret:
            value: router-scoped-v1
          key: secret
        verificationKeys:
          secret:
            value: router-scoped-keys
          key: keys.json
        v1AcceptUntil: "2027-01-01T00:00:00Z"
      cookie:
        name: sandbox-session
        queryParam: token
        sameSite: strict
        allowedOrigins:
          - https://app.example.com
        trustForwardedProto: true
    tls:
      certificateSecret:
        value: sandbox-router-tls
      port: 9443
      httpEnabled: true
      mtlsMode: optional
      clientCa:
        secret:
          value: router-client-ca
        key: ca.crt
      minVersion: tls12
      cipherSuites:
        - TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
    proxy:
      timeout: 2m
      responseHeaderTimeout: 45s
      shutdownTimeout: 1m
      maxRequestBodyBytes: 10485760
      allowLoopbackPodIp: true
      pathRoutingPrefix: /sandboxes
      upstreamMaxRetries: 5
      upstreamRetryInitialDelay: 100ms
      upstreamRetryMaxDelay: 2s
    cache:
      namespace: agents
    telemetry:
      otlpEndpoint: http://otel-collector.observability:4317
      insecure: true
      tracesEnabled: true
      metricsEnabled: false
      headers:
        secret:
          value: otel-headers
        key: headers
    accessLogEnabled: false
    ports:
      http: 8090
      metrics: 9091
      healthProbe: 8083
    networkPolicy:
      clientNamespaceLabels:
        kubernetes.io/metadata.name: agents
      metricsNamespace: observability
      sandboxPorts:
        - 8888
        - 3000
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.crds` | `KubernetesAgentSandboxControllerCrds` |  |  |  |
| `spec.crds.install` | `bool` |  | `true` |  |
| `spec.crds.keepOnUninstall` | `bool` |  | `true` |  |
| `spec.extensions` | `KubernetesAgentSandboxControllerExtensions` |  |  |  |
| `spec.extensions.enabled` | `bool` |  | `true` |  |
| `spec.extensions.allowedLabelDomains` | `[]string` |  |  |  |
| `spec.extensions.claim` | `KubernetesAgentSandboxControllerClaimSettings` |  |  |  |
| `spec.extensions.claim.concurrentWorkers` | `int32` |  | `50` |  |
| `spec.extensions.claim.warmCandidateGracePeriod` | `string` |  | `2s` |  |
| `spec.extensions.claim.eventsDisabled` | `bool` |  |  |  |
| `spec.extensions.claim.observabilityAnnotationsDisabled` | `bool` |  |  |  |
| `spec.extensions.warmPool` | `KubernetesAgentSandboxControllerWarmPoolSettings` |  |  |  |
| `spec.extensions.warmPool.concurrentWorkers` | `int32` |  | `1` |  |
| `spec.extensions.warmPool.maxBatchSize` | `int32` |  | `300` |  |
| `spec.extensions.warmPool.replenishDelay` | `string` |  | `0s` |  |
| `spec.extensions.warmPool.maxRefillRate` | `double` |  | `0` |  |
| `spec.extensions.warmPool.readinessGracePeriod` | `string` |  | `5m` |  |
| `spec.extensions.warmPool.unschedulableRecheckInterval` | `string` |  | `1m` |  |
| `spec.extensions.warmPool.evictionEnabled` | `bool` |  | `true` |  |
| `spec.extensions.template` | `KubernetesAgentSandboxControllerTemplateSettings` |  |  |  |
| `spec.extensions.template.concurrentWorkers` | `int32` |  | `1` |  |
| `spec.controller` | `KubernetesAgentSandboxControllerManager` |  |  |  |
| `spec.controller.image` | `ContainerImage` |  |  |  |
| `spec.controller.image.repo` | `string` |  |  |  |
| `spec.controller.image.tag` | `string` |  |  |  |
| `spec.controller.image.pullSecretName` | `string` |  |  |  |
| `spec.controller.replicas` | `int32` |  | `1` |  |
| `spec.controller.resources` | `ContainerResources` |  |  |  |
| `spec.controller.resources.limits` | `CpuMemory` |  |  |  |
| `spec.controller.resources.limits.cpu` | `string` |  |  |  |
| `spec.controller.resources.limits.memory` | `string` |  |  |  |
| `spec.controller.resources.requests` | `CpuMemory` |  |  |  |
| `spec.controller.resources.requests.cpu` | `string` |  |  |  |
| `spec.controller.resources.requests.memory` | `string` |  |  |  |
| `spec.controller.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.controller.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.controller.tolerations[].key` | `string` |  |  |  |
| `spec.controller.tolerations[].operator` | `string` |  |  |  |
| `spec.controller.tolerations[].value` | `string` |  |  |  |
| `spec.controller.tolerations[].effect` | `string` |  |  |  |
| `spec.controller.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.controller.leaderElection` | `KubernetesAgentSandboxControllerLeaderElection` |  |  |  |
| `spec.controller.leaderElection.enabled` | `bool` |  | `true` |  |
| `spec.controller.leaderElection.namespace` | `string` |  |  |  |
| `spec.controller.metrics` | `KubernetesAgentSandboxControllerMetrics` |  |  |  |
| `spec.controller.metrics.port` | `int32` |  | `8080` |  |
| `spec.controller.metrics.restClientMetricsEnabled` | `bool` |  | `true` |  |
| `spec.controller.metrics.tls` | `KubernetesAgentSandboxControllerMetricsTls` |  |  |  |
| `spec.controller.metrics.tls.enabled` | `bool` |  | `true` |  |
| `spec.controller.metrics.tls.certificateSecret` | `string \| valueFrom` |  |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.controller.metrics.tls.minVersion` | `enum` |  |  |  |
| `spec.controller.metrics.tls.cipherSuites` | `[]string` |  |  |  |
| `spec.controller.healthProbePort` | `int32` |  | `8081` |  |
| `spec.controller.sandbox` | `KubernetesAgentSandboxControllerSandboxSettings` |  |  |  |
| `spec.controller.sandbox.concurrentWorkers` | `int32` |  | `100` |  |
| `spec.controller.sandbox.eventsDisabled` | `bool` |  |  |  |
| `spec.controller.sandbox.writeBehindWindow` | `string` |  | `0s` |  |
| `spec.controller.cacheLabelSelectors` | `bool` |  |  |  |
| `spec.controller.kubeApi` | `KubernetesAgentSandboxControllerKubeApi` |  |  |  |
| `spec.controller.kubeApi.qps` | `double` |  |  |  |
| `spec.controller.kubeApi.burst` | `int32` |  | `10` |  |
| `spec.controller.kubeApi.connections` | `int32` |  | `1` |  |
| `spec.controller.kubeApi.separateWatchConnection` | `bool` |  |  |  |
| `spec.controller.tracing` | `KubernetesAgentSandboxControllerTracing` |  |  |  |
| `spec.controller.tracing.otlpEndpoint` | `string` | yes |  |  |
| `spec.controller.tracing.insecure` | `bool` |  |  |  |
| `spec.controller.tracing.headers` | `KubernetesAgentSandboxControllerSecretKey` |  |  |  |
| `spec.controller.tracing.headers.secret` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.controller.tracing.headers.key` | `string` | yes |  |  |
| `spec.controller.profiling` | `KubernetesAgentSandboxControllerProfiling` |  |  |  |
| `spec.controller.profiling.cpuEnabled` | `bool` |  |  |  |
| `spec.controller.profiling.debugEnabled` | `bool` |  |  |  |
| `spec.controller.profiling.blockProfileRate` | `int32` |  | `1000000` |  |
| `spec.controller.profiling.mutexProfileFraction` | `int32` |  | `10` |  |
| `spec.controller.logging` | `KubernetesAgentSandboxControllerLogging` |  |  |  |
| `spec.controller.logging.development` | `bool` |  |  |  |
| `spec.controller.logging.encoder` | `enum` |  |  |  |
| `spec.controller.logging.level` | `string` |  |  |  |
| `spec.controller.logging.stacktraceLevel` | `enum` |  |  |  |
| `spec.controller.logging.timeEncoding` | `enum` |  |  |  |
| `spec.router` | `KubernetesAgentSandboxControllerRouter` |  |  |  |
| `spec.router.enabled` | `bool` |  | `true` |  |
| `spec.router.image` | `ContainerImage` |  |  |  |
| `spec.router.image.repo` | `string` |  |  |  |
| `spec.router.image.tag` | `string` |  |  |  |
| `spec.router.image.pullSecretName` | `string` |  |  |  |
| `spec.router.replicas` | `int32` |  | `2` |  |
| `spec.router.resources` | `ContainerResources` |  |  |  |
| `spec.router.resources.limits` | `CpuMemory` |  |  |  |
| `spec.router.resources.limits.cpu` | `string` |  |  |  |
| `spec.router.resources.limits.memory` | `string` |  |  |  |
| `spec.router.resources.requests` | `CpuMemory` |  |  |  |
| `spec.router.resources.requests.cpu` | `string` |  |  |  |
| `spec.router.resources.requests.memory` | `string` |  |  |  |
| `spec.router.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.router.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.router.tolerations[].key` | `string` |  |  |  |
| `spec.router.tolerations[].operator` | `string` |  |  |  |
| `spec.router.tolerations[].value` | `string` |  |  |  |
| `spec.router.tolerations[].effect` | `string` |  |  |  |
| `spec.router.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.router.minAvailable` | `int32` |  | `1` |  |
| `spec.router.authorization` | `KubernetesAgentSandboxControllerRouterAuthorization` | yes |  |  |
| `spec.router.authorization.allowAll` | `KubernetesAgentSandboxControllerRouterAllowAll` |  |  |  |
| `spec.router.authorization.tokenReview` | `KubernetesAgentSandboxControllerRouterTokenReview` |  |  |  |
| `spec.router.authorization.tokenReview.ttl` | `string` |  | `30s` |  |
| `spec.router.authorization.tokenReview.cacheSize` | `int32` |  | `2048` |  |
| `spec.router.authorization.tokenReview.requireToken` | `bool` |  |  |  |
| `spec.router.authorization.tokenReview.audiences` | `[]string` |  |  |  |
| `spec.router.authorization.scopedToken` | `KubernetesAgentSandboxControllerRouterScopedToken` |  |  |  |
| `spec.router.authorization.scopedToken.v1Secret` | `KubernetesAgentSandboxControllerSecretKey` |  |  |  |
| `spec.router.authorization.scopedToken.v1Secret.secret` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.router.authorization.scopedToken.v1Secret.key` | `string` | yes |  |  |
| `spec.router.authorization.scopedToken.verificationKeys` | `KubernetesAgentSandboxControllerSecretKey` |  |  |  |
| `spec.router.authorization.scopedToken.verificationKeys.secret` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.router.authorization.scopedToken.verificationKeys.key` | `string` | yes |  |  |
| `spec.router.authorization.scopedToken.v1AcceptUntil` | `string` |  |  |  |
| `spec.router.authorization.cookie` | `KubernetesAgentSandboxControllerRouterCookie` |  |  |  |
| `spec.router.authorization.cookie.name` | `string` | yes |  |  |
| `spec.router.authorization.cookie.queryParam` | `string` |  |  |  |
| `spec.router.authorization.cookie.sameSite` | `enum` |  |  |  |
| `spec.router.authorization.cookie.insecure` | `bool` |  |  |  |
| `spec.router.authorization.cookie.allowedOrigins` | `[]string` |  |  |  |
| `spec.router.authorization.cookie.trustForwardedProto` | `bool` |  |  |  |
| `spec.router.tls` | `KubernetesAgentSandboxControllerRouterTls` |  |  |  |
| `spec.router.tls.enabled` | `bool` |  | `true` |  |
| `spec.router.tls.certificateSecret` | `string \| valueFrom` | yes |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.router.tls.port` | `int32` |  | `8443` |  |
| `spec.router.tls.httpEnabled` | `bool` |  | `true` |  |
| `spec.router.tls.mtlsMode` | `enum` |  |  |  |
| `spec.router.tls.clientCa` | `KubernetesAgentSandboxControllerSecretKey` |  |  |  |
| `spec.router.tls.clientCa.secret` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.router.tls.clientCa.key` | `string` | yes |  |  |
| `spec.router.tls.minVersion` | `enum` |  |  |  |
| `spec.router.tls.cipherSuites` | `[]string` |  |  |  |
| `spec.router.proxy` | `KubernetesAgentSandboxControllerRouterProxy` |  |  |  |
| `spec.router.proxy.timeout` | `string` |  | `180s` |  |
| `spec.router.proxy.responseHeaderTimeout` | `string` |  | `30s` |  |
| `spec.router.proxy.shutdownTimeout` | `string` |  | `30s` |  |
| `spec.router.proxy.maxRequestBodyBytes` | `int64` |  |  |  |
| `spec.router.proxy.allowLoopbackPodIp` | `bool` |  |  |  |
| `spec.router.proxy.pathRoutingPrefix` | `string` |  |  |  |
| `spec.router.proxy.upstreamMaxRetries` | `int32` |  | `3` |  |
| `spec.router.proxy.upstreamRetryInitialDelay` | `string` |  | `200ms` |  |
| `spec.router.proxy.upstreamRetryMaxDelay` | `string` |  | `800ms` |  |
| `spec.router.cache` | `KubernetesAgentSandboxControllerRouterCache` |  |  |  |
| `spec.router.cache.enabled` | `bool` |  | `true` |  |
| `spec.router.cache.namespace` | `string` |  |  |  |
| `spec.router.telemetry` | `KubernetesAgentSandboxControllerRouterTelemetry` |  |  |  |
| `spec.router.telemetry.otlpEndpoint` | `string` | yes |  |  |
| `spec.router.telemetry.insecure` | `bool` |  |  |  |
| `spec.router.telemetry.tracesEnabled` | `bool` |  | `true` |  |
| `spec.router.telemetry.metricsEnabled` | `bool` |  | `true` |  |
| `spec.router.telemetry.headers` | `KubernetesAgentSandboxControllerSecretKey` |  |  |  |
| `spec.router.telemetry.headers.secret` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.router.telemetry.headers.key` | `string` | yes |  |  |
| `spec.router.accessLogEnabled` | `bool` |  | `true` |  |
| `spec.router.ports` | `KubernetesAgentSandboxControllerRouterPorts` |  |  |  |
| `spec.router.ports.http` | `int32` |  | `8080` |  |
| `spec.router.ports.metrics` | `int32` |  | `9090` |  |
| `spec.router.ports.healthProbe` | `int32` |  | `8081` |  |
| `spec.router.networkPolicy` | `KubernetesAgentSandboxControllerRouterNetworkPolicy` |  |  |  |
| `spec.router.networkPolicy.enabled` | `bool` |  | `true` |  |
| `spec.router.networkPolicy.clientNamespaceLabels` | `map<string, string>` |  |  |  |
| `spec.router.networkPolicy.metricsNamespace` | `string` |  | `monitoring` |  |
| `spec.router.networkPolicy.sandboxPorts` | `[]int32` |  |  |  |
| `spec.clusterDomain` | `string` |  | `cluster.local` |  |
| `spec.imageRegistry` | `string` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

The namespace everything installs into. It is always
"agent-sandbox-system": upstream's manifests, RBAC and the default
sandbox network policy name it. Reference a KubernetesNamespace of that
name to let it own the namespace (its quotas and policies apply, and
the Secrets this kind reads can be placed there first), or set the
literal value with create_namespace.

- references: KubernetesNamespace (`spec.name`)
- rule: namespace must be "agent-sandbox-system": upstream's manifests, RBAC and the default sandbox network policy name it, so agent-sandbox runs nowhere else
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

Create the namespace (with the standard Planton governance labels)
and delete it with this resource, which deletes everything in it. Leave
it false when a KubernetesNamespace owns the namespace; then the
namespace must exist before the deploy.

### spec.crds

`KubernetesAgentSandboxControllerCrds`

The lifecycle of the four resource definitions (Sandbox, and with
extensions SandboxTemplate, SandboxWarmPool and SandboxClaim). Empty =
install them from the pinned release and keep them on destroy.

### spec.crds.install

`bool` · optional (explicit presence)

Apply the definitions from the pinned release. Set false only when
another tool manages them (a GitOps repository applying upstream's
definitions) and no other installation of agent-sandbox runs on the
cluster; the controller still installs and needs them.

- default: `true`

### spec.crds.keepOnUninstall

`bool` · optional (explicit presence)

Leave the definitions on the cluster when this resource is destroyed.
Deleting a definition deletes every object of that type, so false
deletes every Sandbox, SandboxTemplate, SandboxWarmPool and SandboxClaim
on the cluster, with their pods; their volumes follow their own
reclaim policy. Keep it true outside throwaway clusters.

- default: `true`

### spec.extensions

`KubernetesAgentSandboxControllerExtensions`

The extensions: SandboxTemplate, SandboxWarmPool and SandboxClaim, and
the controllers that serve them. Absent = core only (Sandbox), which is
upstream's default and all a product that creates Sandboxes directly
needs. Present = on; turn it on for warm pools and claims, which is what
upstream's SDKs use.

### spec.extensions.enabled

`bool` · optional (explicit presence)

Run the extension controllers and install their definitions. False keeps
the settings below while turning the extensions off. Their definitions
stay on the cluster while crds.keep_on_uninstall is true, so existing
templates, warm pools and claims survive and resume when the extensions
return; with it false they are deleted, with every object of their kind.

- default: `true`

### spec.extensions.allowedLabelDomains

`[]string`

Label domains a SandboxClaim may set on its sandbox's pod through
`additionalPodMetadata.labels`. A label is allowed when its prefix is a
listed domain or a subdomain of one; every other prefixed label, and
every unprefixed one, refuses the claim. The list replaces upstream's
default, so include "sandbox.users.io" to keep it. The controller reads
the list at startup; a change restarts it. Empty = "sandbox.users.io".

- rule: {"repeated":{"unique":true,"items":{"cel":[{"id":"allowed_label_domains_format","message":"each allowed label domain must be a lowercase DNS domain such as \"sandbox.users.io\" or \"example.com\"","expression":"this.matches('^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$')"}]}}}

### spec.extensions.claim

`KubernetesAgentSandboxControllerClaimSettings`

The SandboxClaim controller.

### spec.extensions.claim.concurrentWorkers

`int32` · optional (explicit presence)

Claims reconciled at once.

- default: `50`
- rule: {"int32":{"gt":0}}

### spec.extensions.claim.warmCandidateGracePeriod

`string` · optional (explicit presence)

How long a new claim waits for a warm pool member to report a pod IP
before it creates a cold sandbox instead (Go duration such as "2s").
Raise it when warm members report late and claims start cold.

- default: `2s`
- rule: warm_candidate_grace_period must be a positive Go duration such as "2s" or "500ms"

### spec.extensions.claim.eventsDisabled

`bool`

Stop the claim controller emitting Kubernetes Events, saving API writes
during large claim bursts.

### spec.extensions.claim.observabilityAnnotationsDisabled

`bool`

Stop persisting a claim's observability annotations (when the controller
first saw it, its trace context), saving one API write per claim.
Startup-latency metrics and trace propagation keep working inside the
running controller; the on-object breadcrumbs are lost, and so is the
latency metric for claims first seen before a controller restart.

### spec.extensions.warmPool

`KubernetesAgentSandboxControllerWarmPoolSettings`

The SandboxWarmPool controller.

### spec.extensions.warmPool.concurrentWorkers

`int32` · optional (explicit presence)

Pools reconciled at once.

- default: `1`
- rule: {"int32":{"gt":0}}

### spec.extensions.warmPool.maxBatchSize

`int32` · optional (explicit presence)

Sandboxes created or deleted in one batch. A pool fills in about
replicas / max_batch_size round trips to the API server; raising it
trades round trips for burst size.

- default: `300`
- rule: {"int32":{"gt":0}}

### spec.extensions.warmPool.replenishDelay

`string` · optional (explicit presence)

How long refill waits after members leave the pool (claims adopting
them), so a claim burst gets the API server first (Go duration). The
wait restarts while members keep leaving. "0s" refills at once.

- default: `0s`
- rule: replenish_delay must be a Go duration such as "0s", "500ms" or "10s"

### spec.extensions.warmPool.maxRefillRate

`double` · optional (explicit presence)

Most replacement sandboxes created per second per pool, which turns
refill into a steady stream instead of whole-deficit bursts. Combines
with replenish_delay: the delay decides when refill starts, the rate how
fast it flows. 0 leaves refill unpaced.

- default: `0`
- rule: max_refill_rate must be a finite number of sandboxes per second, 0 or more (0 leaves refill unpaced)

### spec.extensions.warmPool.readinessGracePeriod

`string` · optional (explicit presence)

How long a pool member may stay not Ready, counted from when its pod was
scheduled, before the controller replaces it, or holds it when its pod
cannot be scheduled (Go duration). Raise it for images that take long to
initialize.

- default: `5m`
- rule: readiness_grace_period must be a positive Go duration such as "5m" or "90s"

### spec.extensions.warmPool.unschedulableRecheckInterval

`string` · optional (explicit presence)

How often a pool holding unschedulable members past their grace period
is checked again (Go duration).

- default: `1m`
- rule: unschedulable_recheck_interval must be a positive Go duration such as "1m" or "30s"

### spec.extensions.warmPool.evictionEnabled

`bool` · optional (explicit presence)

Mark pool members' pods safe for the cluster autoscaler to evict, so an
idle warm pool never holds a node up. Set false to keep warm members
where they are.

- default: `true`

### spec.extensions.template

`KubernetesAgentSandboxControllerTemplateSettings`

The SandboxTemplate controller.

### spec.extensions.template.concurrentWorkers

`int32` · optional (explicit presence)

Templates reconciled at once.

- default: `1`
- rule: {"int32":{"gt":0}}

### spec.controller

`KubernetesAgentSandboxControllerManager`

The controller Deployment (`agent-sandbox-controller`): replicas,
resources, scheduling and every controller setting. Empty = upstream's
release manifest as published.

- rule: more than one controller replica needs leader election: without it every replica reconciles at once; keep leader_election.enabled true or set replicas to 1
- rule: the controller's metrics port and health probe port must differ

### spec.controller.image

`ContainerImage`

Override the controller image (a patched build, or a private mirror's
path). Empty = registry.k8s.io/agent-sandbox/agent-sandbox-controller at
the pinned release (moved to image_registry when that is set).

### spec.controller.image.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.controller.image.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.controller.image.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.controller.replicas

`int32` · optional (explicit presence)

Controller pods. One leads and the rest wait as hot standbys, so more
than one needs leader election (on by default).

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.controller.resources

`ContainerResources`

Resources for the controller container. Empty = none set (upstream's
manifest sets none); set requests on clusters with resource quotas.

### spec.controller.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.controller.resources.limits.cpu

`string`

### spec.controller.resources.limits.memory

`string`

### spec.controller.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.controller.resources.requests.cpu

`string`

### spec.controller.resources.requests.memory

`string`

### spec.controller.nodeSelector

`map<string, string>`

Node selector for the controller pods.

### spec.controller.tolerations

`[]WorkloadToleration`

Tolerations for the controller pods.

### spec.controller.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.controller.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.controller.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.controller.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.controller.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.controller.leaderElection

`KubernetesAgentSandboxControllerLeaderElection`

Leader election among controller pods.

### spec.controller.leaderElection.enabled

`bool` · optional (explicit presence)

Elect one active controller among the replicas. Turn it off only for a
single replica that must start without waiting for a lease.

- default: `true`

### spec.controller.leaderElection.namespace

`string`

Namespace that holds the election lease. Empty = the controller's own
namespace, `agent-sandbox-system`.

- rule: leader_election.namespace must be a Kubernetes namespace name: lowercase letters, digits and hyphens, starting and ending with a letter or digit
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63"}}

### spec.controller.metrics

`KubernetesAgentSandboxControllerMetrics`

The metrics endpoint the controller serves, and its TLS.

### spec.controller.metrics.port

`int32` · optional (explicit presence)

Port the metrics endpoint listens on, and the port of the
`agent-sandbox-controller` Service in front of it. The conventional
port for HTTPS metrics is 8443.

- default: `8080`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.controller.metrics.restClientMetricsEnabled

`bool` · optional (explicit presence)

Publish the Kubernetes API client's own metrics (request latency, DNS
resolution, request and response sizes, rate limiting, retries). Set
false if their histograms cost too much.

- default: `true`

### spec.controller.metrics.tls

`KubernetesAgentSandboxControllerMetricsTls`

Serve metrics over HTTPS. Absent = plain HTTP.

- rule: cipher_suites cannot be set with min_version tls13: TLS 1.3 cipher suites are not configurable, so the list would be ignored

### spec.controller.metrics.tls.enabled

`bool` · optional (explicit presence)

Serve metrics over HTTPS. False keeps the settings below while serving
plain HTTP.

- default: `true`

### spec.controller.metrics.tls.certificateSecret

`string | valueFrom`

Secret in `agent-sandbox-system` holding the server certificate as
`tls.crt` and `tls.key`, typically written by a cert-manager
Certificate. Empty = the controller generates a self-signed certificate
at startup, which scrapers must then be told to trust or skip.

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.controller.metrics.tls.minVersion

`enum`

Oldest TLS version the metrics endpoint accepts. Empty = TLS 1.2 (Go's
default).

Allowed values (use exactly as shown):

- `tls_version_unspecified`
- `tls10`
- `tls11`
- `tls12`
- `tls13`

### spec.controller.metrics.tls.cipherSuites

`[]string`

Cipher suites the metrics endpoint offers, by Go name (for example
"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"). TLS 1.3 suites are not
configurable. Empty = Go's defaults.

- rule: {"repeated":{"unique":true,"items":{"cel":[{"id":"metrics_cipher_suites_format","message":"each cipher suite must be a Go cipher-suite name such as \"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256\"","expression":"this.matches('^TLS_[A-Z0-9_]+$')"}]}}}

### spec.controller.healthProbePort

`int32` · optional (explicit presence)

Port of the controller's /healthz and /readyz endpoints.

- default: `8081`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.controller.sandbox

`KubernetesAgentSandboxControllerSandboxSettings`

The Sandbox controller.

### spec.controller.sandbox.concurrentWorkers

`int32` · optional (explicit presence)

Sandboxes reconciled at once.

- default: `100`
- rule: {"int32":{"gt":0}}

### spec.controller.sandbox.eventsDisabled

`bool`

Stop the Sandbox controller emitting Kubernetes Events, saving API
writes during large warm-pool fills.

### spec.controller.sandbox.writeBehindWindow

`string` · optional (explicit presence)

Window over which the Sandbox controller coalesces its recoverable,
metadata-only writes into one (Go duration). "0s" writes each at once.

- default: `0s`
- rule: write_behind_window must be a Go duration such as "0s" or "200ms"

### spec.controller.cacheLabelSelectors

`bool`

Limit the controller's Pod and Service caches to objects carrying the
sandbox tracking label, cutting watch traffic, CPU and memory from the
whole cluster's pods to the sandboxes' own. Turn it on for shared or
busy clusters. Pods and Services created outside the controller and
adopted through the `agents.x-k8s.io/adoptable` label must then also
carry the tracking label, or the controller cannot see them.

### spec.controller.kubeApi

`KubernetesAgentSandboxControllerKubeApi`

How the controller talks to the Kubernetes API server. Touch these only
when a large fleet makes the controller wait on its own API client.

### spec.controller.kubeApi.qps

`double` · optional (explicit presence)

Client-side limit on requests per second. Empty = no client-side limit
(the API server's own flow control still applies).

- rule: kube_api.qps must be a positive number of requests per second; leave it empty for no client-side limit

### spec.controller.kubeApi.burst

`int32` · optional (explicit presence)

Requests allowed above qps in a burst.

- default: `10`
- rule: {"int32":{"gt":0}}

### spec.controller.kubeApi.connections

`int32` · optional (explicit presence)

Separate HTTP/2 connections to the API server for requests other than
watches. The API server caps concurrent requests per connection (100 by
default), so one connection caps the controller there whatever its
workers and qps; more connections raise the ceiling about N times.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.controller.kubeApi.separateWatchConnection

`bool`

Give the controller's watch streams their own connection, so watch
events never queue behind bursts of writes on a shared one.

### spec.controller.tracing

`KubernetesAgentSandboxControllerTracing`

Send traces to an OpenTelemetry collector over OTLP gRPC. Absent = no
tracing.

### spec.controller.tracing.otlpEndpoint

`string` · required

The collector's OTLP gRPC endpoint, for example
"http://otel-collector.observability.svc.cluster.local:4317". The
controller exits at startup when it cannot set up the exporter.

- rule: tracing.otlp_endpoint must be a URL such as "http://otel-collector.observability:4317"
- rule: {"required":true}

### spec.controller.tracing.insecure

`bool`

Connect to the collector without TLS.

### spec.controller.tracing.headers

`KubernetesAgentSandboxControllerSecretKey`

Headers sent with every export, read from a key of a Secret in
`agent-sandbox-system` in OpenTelemetry's form
("name1=value1,name2=value2"), for example a hosted collector's API key
("x-honeycomb-team=..."). The other OpenTelemetry exporter settings
(certificates, compression, timeouts, per-signal endpoints) are not
modeled; the guide says why.

### spec.controller.tracing.headers.secret

`string | valueFrom` · required

The Secret's name.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.controller.tracing.headers.key

`string` · required

The key within the Secret.

- rule: key must be a Secret key: letters, digits, "-", "_" and "."
- rule: {"required":true}

### spec.controller.profiling

`KubernetesAgentSandboxControllerProfiling`

Profiling endpoints on the metrics server, for diagnosing a live
controller. Absent = off.

- rule: block_profile_rate and mutex_profile_fraction apply only with debug_enabled: turn debug_enabled on, or leave them at their defaults

### spec.controller.profiling.cpuEnabled

`bool`

Serve the CPU profile at /debug/pprof/profile.

### spec.controller.profiling.debugEnabled

`bool`

Serve every pprof endpoint (heap, goroutine, block, mutex, trace,
cmdline, symbol) and fgprof, the CPU profile included. These can expose
sensitive process details and cost performance; use them while
diagnosing, then turn them off.

### spec.controller.profiling.blockProfileRate

`int32` · optional (explicit presence)

Block-profile sampling rate when debug_enabled is on: 0 or less
disables it, 1 samples every blocking event, N samples one event per N
nanoseconds blocked (1000000 is about 1ms).

- default: `1000000`

### spec.controller.profiling.mutexProfileFraction

`int32` · optional (explicit presence)

Mutex-contention sampling when debug_enabled is on: 0 or less disables
it, 1 samples every event, N samples about one in N.

- default: `10`

### spec.controller.logging

`KubernetesAgentSandboxControllerLogging`

Log format and verbosity. Empty = JSON at info level.

### spec.controller.logging.development

`bool`

Development logging: human-readable console output, debug level and
stack traces from warnings up, unless the settings below say otherwise.

### spec.controller.logging.encoder

`enum`

Log encoding. Empty = JSON (console when development is on).

Allowed values (use exactly as shown):

- `encoder_unspecified`
- `json`
- `console`

### spec.controller.logging.level

`string`

Verbosity: "debug", "info", "error" or "panic", or a whole number above
0 for increasingly verbose debug levels. Empty = "info" ("debug" when
development is on).

- rule: logging.level must be "debug", "info", "error" or "panic", or a whole number above 0
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.controller.logging.stacktraceLevel

`enum`

Level from which log lines carry a stack trace. Empty = error (warn when
development is on).

Allowed values (use exactly as shown):

- `stacktrace_level_unspecified`
- `info`
- `error`
- `panic`

### spec.controller.logging.timeEncoding

`enum`

How timestamps are written. Empty = RFC 3339.

Allowed values (use exactly as shown):

- `time_encoding_unspecified`
- `epoch`
- `millis`
- `nano`
- `iso8601`
- `rfc3339`
- `rfc3339nano`

### spec.router

`KubernetesAgentSandboxControllerRouter`

The router (`sandbox-router`): the HTTP entry point that forwards a
request to a sandbox named in its `X-Sandbox-ID` header (or in the
path). Absent = no router. Present = installed; its authorization mode
must be chosen explicitly, because upstream's default lets every
request through.

- rule: router.min_available must be lower than router.replicas, or no router pod can ever be drained; lower it, raise replicas, or set it to 0 to remove the budget
- rule: router.authorization.cookie needs router.proxy.path_routing_prefix: a browser reaches a sandbox only by path, so without a prefix there is no route for the cookie to protect
- rule: scoped_token.verification_keys need the pod-IP cache: v2 tokens are bound to a sandbox's UID, which only the cache can route; leave router.cache.enabled on
- rule: the router needs a listener: keep router.tls.http_enabled true, or the router serves nothing without TLS
- rule: the router's HTTP, HTTPS, metrics and health ports must all differ

### spec.router.enabled

`bool` · optional (explicit presence)

Install the router. False keeps the settings below while removing it.

- default: `true`

### spec.router.image

`ContainerImage`

Override the router image. Empty =
registry.k8s.io/agent-sandbox/sandbox-router-go at the pinned release
(moved to image_registry when that is set).

### spec.router.image.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.router.image.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.router.image.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.router.replicas

`int32` · optional (explicit presence)

Router pods, spread across zones and nodes. Two is the minimum for
availability, not a capacity figure; the router is CPU-bound at high
request rates.

- default: `2`
- rule: {"int32":{"gte":1}}

### spec.router.resources

`ContainerResources`

Resources for the router container. Empty = upstream's: requests 250m
CPU and 256Mi memory, limits 2 CPU and 1Gi.

### spec.router.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.router.resources.limits.cpu

`string`

### spec.router.resources.limits.memory

`string`

### spec.router.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.router.resources.requests.cpu

`string`

### spec.router.resources.requests.memory

`string`

### spec.router.nodeSelector

`map<string, string>`

Node selector for the router pods.

### spec.router.tolerations

`[]WorkloadToleration`

Tolerations for the router pods.

### spec.router.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.router.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.router.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.router.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.router.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.router.minAvailable

`int32` · optional (explicit presence)

Router pods that must stay up during voluntary disruptions such as node
drains (the router's PodDisruptionBudget). It must be lower than
replicas, or a drain waits forever. 0 removes the budget.

- default: `1`
- rule: {"int32":{"gte":0}}

### spec.router.authorization

`KubernetesAgentSandboxControllerRouterAuthorization` · required

Who may reach a sandbox through the router. Required: upstream's default
lets every request through, so an open router is chosen, never
inherited.

- rule: {"required":true}
- rule: authorization.cookie has no effect with allow_all, which authorizes every request regardless; choose token_review or scoped_token, or remove the cookie

### spec.router.authorization.allowAll

`KubernetesAgentSandboxControllerRouterAllowAll`

Every request is let through. Anything that can reach the router can
reach every sandbox, so pair it with a NetworkPolicy, a Gateway that
authenticates, or a cluster you fully trust.

### spec.router.authorization.tokenReview

`KubernetesAgentSandboxControllerRouterTokenReview`

Each request carries a Kubernetes bearer token
(`Authorization: Bearer ...`), checked through the TokenReview API.
Fits callers that run in the cluster with a ServiceAccount.

### spec.router.authorization.tokenReview.ttl

`string` · optional (explicit presence)

How long a review's answer is cached (Go duration). Shorter catches a
revoked token sooner at the cost of more API server load.

- default: `30s`
- rule: token_review.ttl must be a positive Go duration such as "30s" or "2m"

### spec.router.authorization.tokenReview.cacheSize

`int32` · optional (explicit presence)

Cached review answers kept before the oldest is evicted.

- default: `2048`
- rule: {"int32":{"gt":0}}

### spec.router.authorization.tokenReview.requireToken

`bool`

Refuse requests without a bearer token (401). Off lets tokenless
requests through, which only suits a rollout while callers start
sending tokens.

### spec.router.authorization.tokenReview.audiences

`[]string`

Audiences a token's `aud` claim must include. Required for projected
ServiceAccount tokens minted with an audience. Empty = no audience
check.

- rule: {"repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.router.authorization.scopedToken

`KubernetesAgentSandboxControllerRouterScopedToken`

Each request carries a token scoped to one sandbox, minted out of band
by the product that created it: v1 tokens are HMAC-signed with a
shared secret, v2 tokens are Ed25519-signed and bound to the sandbox's
UID.

- rule: scoped_token needs v1_secret, verification_keys, or both: otherwise no token can ever be verified
- rule: scoped_token.v1_accept_until is required while v1_secret and verification_keys are both set: it is when v1 tokens stop being accepted
- rule: scoped_token.v1_accept_until applies only to v1 tokens; set v1_secret or remove the cutoff

### spec.router.authorization.scopedToken.v1Secret

`KubernetesAgentSandboxControllerSecretKey`

The shared HMAC-SHA256 secret that verifies v1 tokens, as a key of a
Secret in `agent-sandbox-system`.

### spec.router.authorization.scopedToken.v1Secret.secret

`string | valueFrom` · required

The Secret's name.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.router.authorization.scopedToken.v1Secret.key

`string` · required

The key within the Secret.

- rule: key must be a Secret key: letters, digits, "-", "_" and "."
- rule: {"required":true}

### spec.router.authorization.scopedToken.verificationKeys

`KubernetesAgentSandboxControllerSecretKey`

The JSON key set of Ed25519 public keys that verifies v2 tokens, as a
key of a Secret in `agent-sandbox-system`. Needs the pod-IP cache.

### spec.router.authorization.scopedToken.verificationKeys.secret

`string | valueFrom` · required

The Secret's name.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.router.authorization.scopedToken.verificationKeys.key

`string` · required

The key within the Secret.

- rule: key must be a Secret key: letters, digits, "-", "_" and "."
- rule: {"required":true}

### spec.router.authorization.scopedToken.v1AcceptUntil

`string`

The moment v1 tokens stop being accepted while v1 and v2 run side by
side (RFC 3339, for example "2026-12-31T00:00:00Z", exclusive).
Required when both are set.

- rule: scoped_token.v1_accept_until must be an RFC 3339 time such as "2026-12-31T00:00:00Z"
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.router.authorization.cookie

`KubernetesAgentSandboxControllerRouterCookie`

Also accept the credential from a cookie, the one channel a browser
attaches to every request, a WebSocket handshake included. Needs
token_review or scoped_token, and proxy.path_routing_prefix.

- rule: cookie.same_site none needs a Secure cookie, so it cannot be combined with cookie.insecure

### spec.router.authorization.cookie.name

`string` · required

The cookie's name: letters, digits, "-", "_", "." and "~".

- rule: cookie.name must be a cookie-name token: letters, digits, "-", "_", "." and "~"
- rule: {"required":true}

### spec.router.authorization.cookie.queryParam

`string`

A URL query parameter that sets the cookie: a GET or HEAD request that
presents a valid credential here gets the cookie, scoped to exactly the
sandbox it authorized, and is redirected to the same URL without the
parameter, so the credential spends the least time in URLs, history and
Referer headers. Empty = the cookie must be set some other way.

- rule: cookie.query_param must be a token: letters, digits, "-", "_", "." and "~"
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.router.authorization.cookie.sameSite

`enum`

The cookie's SameSite attribute. Empty = lax, which covers same-site
embedding without HTTPS. strict is never sent cross-site, including the
navigation that sets it. none is for genuinely cross-site embedding,
needs a Secure cookie, and some browsers block it regardless.

Allowed values (use exactly as shown):

- `same_site_unspecified`
- `lax`
- `strict`
- `none`

### spec.router.authorization.cookie.insecure

`bool`

Drop the cookie's Secure attribute. Only for local development over
plain HTTP; not allowed with same_site none.

### spec.router.authorization.cookie.allowedOrigins

`[]string`

Origins (scheme://host[:port]) whose pages may make cookie-authorized
requests. Same-origin requests and requests without an Origin header
are always let through; only cookie-borne credentials are checked.

- rule: {"repeated":{"unique":true,"items":{"cel":[{"id":"cookie_allowed_origins_format","message":"each allowed origin must be scheme://host[:port], such as \"https://ide.example.com\", with no path, query or fragment","expression":"this.matches('^[a-z][a-z0-9+.-]*://[^/?#\\\\s]+$')"}]}}}

### spec.router.authorization.cookie.trustForwardedProto

`bool`

Read the request's scheme from X-Forwarded-Proto for the same-origin
check. Needed behind a TLS-terminating load balancer or Gateway; turn it
on only when the router is reachable solely through a proxy that sets
the header itself and strips any the client sent.

### spec.router.tls

`KubernetesAgentSandboxControllerRouterTls`

Serve the proxy over TLS. Absent = plain HTTP only.

- rule: router.tls.mtls_mode optional or required needs client_ca, the CA bundle that verifies client certificates
- rule: cipher_suites cannot be set with min_version tls13: TLS 1.3 cipher suites are not configurable, so the list would be ignored

### spec.router.tls.enabled

`bool` · optional (explicit presence)

Serve the proxy over TLS. False keeps the settings below while serving
plain HTTP only.

- default: `true`

### spec.router.tls.certificateSecret

`string | valueFrom` · required

Secret in `agent-sandbox-system` holding the server certificate as
`tls.crt` and `tls.key`, typically written by a cert-manager
Certificate.

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.router.tls.port

`int32` · optional (explicit presence)

Port of the TLS listener, and of the router Service's `proxy-tls` port.

- default: `8443`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.router.tls.httpEnabled

`bool` · optional (explicit presence)

Keep the plain-HTTP listener beside TLS. Set false to serve TLS only.

- default: `true`

### spec.router.tls.mtlsMode

`enum`

Client-certificate verification. Empty = off. optional verifies a
certificate when a client presents one; required refuses clients
without one. Both need client_ca.

Allowed values (use exactly as shown):

- `mtls_mode_unspecified`
- `off`
- `optional`
- `required`

### spec.router.tls.clientCa

`KubernetesAgentSandboxControllerSecretKey`

The CA bundle that verifies client certificates, as a key of a Secret in
`agent-sandbox-system` (a cert-manager Certificate's Secret carries it
as `ca.crt`).

### spec.router.tls.clientCa.secret

`string | valueFrom` · required

The Secret's name.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.router.tls.clientCa.key

`string` · required

The key within the Secret.

- rule: key must be a Secret key: letters, digits, "-", "_" and "."
- rule: {"required":true}

### spec.router.tls.minVersion

`enum`

Oldest TLS version the listener accepts. Empty = TLS 1.2.

Allowed values (use exactly as shown):

- `tls_version_unspecified`
- `tls10`
- `tls11`
- `tls12`
- `tls13`

### spec.router.tls.cipherSuites

`[]string`

Cipher suites the listener offers, by Go name. TLS 1.3 suites are not
configurable. Empty = Go's defaults.

- rule: {"repeated":{"unique":true,"items":{"cel":[{"id":"router_cipher_suites_format","message":"each cipher suite must be a Go cipher-suite name such as \"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256\"","expression":"this.matches('^TLS_[A-Z0-9_]+$')"}]}}}

### spec.router.proxy

`KubernetesAgentSandboxControllerRouterProxy`

How requests are forwarded to sandboxes.

### spec.router.proxy.timeout

`string` · optional (explicit presence)

Total time budget for proxying one request (Go duration). Long-running
requests and streams need more.

- default: `180s`
- rule: proxy.timeout must be a positive Go duration such as "180s" or "10m"

### spec.router.proxy.responseHeaderTimeout

`string` · optional (explicit presence)

Longest wait for a sandbox's response headers (Go duration).

- default: `30s`
- rule: proxy.response_header_timeout must be a positive Go duration such as "30s"

### spec.router.proxy.shutdownTimeout

`string` · optional (explicit presence)

Time to drain in-flight requests when a router pod stops (Go duration).
The pod's termination grace period is set 15 seconds longer.

- default: `30s`
- rule: proxy.shutdown_timeout must be a Go duration such as "30s" ("0s" stops without draining)

### spec.router.proxy.maxRequestBodyBytes

`int64`

Largest request body accepted, in bytes. 0 = no limit.

- rule: {"int64":{"gte":"0"}}

### spec.router.proxy.allowLoopbackPodIp

`bool`

Accept a loopback address (127.0.0.0/8, ::1) in X-Sandbox-Pod-IP. Only
for a router running as a sidecar beside its sandbox, or tests against
localhost; link-local, multicast and unspecified addresses stay refused.

### spec.router.proxy.pathRoutingPrefix

`string`

A path prefix that also lets a caller address a sandbox by URL:
<prefix>/<namespace>/<id>/<port>/<rest>. Browsers need it: a WebSocket
handshake cannot carry custom headers, so a web IDE's terminal or a dev
server's live reload inside a sandbox is reachable only by path. Every
path under the prefix is claimed for routing, so pick one no sandbox
route uses. Empty = headers only.

- rule: proxy.path_routing_prefix must start with "/", must not end with "/", and must not contain spaces, such as "/sandboxes"
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.router.proxy.upstreamMaxRetries

`int32` · optional (explicit presence)

Extra connection attempts when a sandbox cannot be dialed (DNS failure,
connection refused) before the router answers 502. Smooths requests to
a sandbox that is still starting. 0 = no retries.

- default: `3`
- rule: {"int32":{"gte":0}}

### spec.router.proxy.upstreamRetryInitialDelay

`string` · optional (explicit presence)

Wait before the first retry; each later wait doubles, up to
upstream_retry_max_delay (Go duration).

- default: `200ms`
- rule: proxy.upstream_retry_initial_delay must be a Go duration such as "200ms"

### spec.router.proxy.upstreamRetryMaxDelay

`string` · optional (explicit presence)

Longest wait between retries (Go duration).

- default: `800ms`
- rule: proxy.upstream_retry_max_delay must be a Go duration such as "800ms" or "2s"

### spec.router.cache

`KubernetesAgentSandboxControllerRouterCache`

The router's pod-IP cache.

### spec.router.cache.enabled

`bool` · optional (explicit presence)

Watch sandbox pods and dial a cached pod IP for requests that carry
X-Sandbox-UID, skipping DNS. Needs read access to pods, which the router
gets cluster-wide (it filters system pods out itself). Off = DNS-only
routing, and no pod read access.

- default: `true`

### spec.router.cache.namespace

`string`

Watch only this namespace's pods. Empty = every namespace.

- rule: router.cache.namespace must be a Kubernetes namespace name: lowercase letters, digits and hyphens, starting and ending with a letter or digit
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63"}}

### spec.router.telemetry

`KubernetesAgentSandboxControllerRouterTelemetry`

Send traces and metrics to an OpenTelemetry collector over OTLP gRPC.
Absent = none (the Prometheus /metrics endpoint is always served).

### spec.router.telemetry.otlpEndpoint

`string` · required

The collector's OTLP gRPC endpoint, for example
"http://otel-collector.observability.svc.cluster.local:4317".

- rule: router.telemetry.otlp_endpoint must be a URL such as "http://otel-collector.observability:4317"
- rule: {"required":true}

### spec.router.telemetry.insecure

`bool`

Connect to the collector without TLS.

### spec.router.telemetry.tracesEnabled

`bool` · optional (explicit presence)

Send traces.

- default: `true`

### spec.router.telemetry.metricsEnabled

`bool` · optional (explicit presence)

Also push metrics (the Prometheus /metrics endpoint stays served).

- default: `true`

### spec.router.telemetry.headers

`KubernetesAgentSandboxControllerSecretKey`

Headers sent with every export, read from a key of a Secret in
`agent-sandbox-system` in OpenTelemetry's form
("name1=value1,name2=value2"), for example a hosted collector's API key.

### spec.router.telemetry.headers.secret

`string | valueFrom` · required

The Secret's name.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.router.telemetry.headers.key

`string` · required

The key within the Secret.

- rule: key must be a Secret key: letters, digits, "-", "_" and "."
- rule: {"required":true}

### spec.router.accessLogEnabled

`bool` · optional (explicit presence)

Log one structured line per proxied request (health and metrics
requests are skipped).

- default: `true`

### spec.router.ports

`KubernetesAgentSandboxControllerRouterPorts`

The router's plain-HTTP, metrics and health ports.

### spec.router.ports.http

`int32` · optional (explicit presence)

Port of the plain-HTTP proxy listener, and of the `sandbox-router-svc`
Service's `proxy` port that upstream's SDKs dial.

- default: `8080`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.router.ports.metrics

`int32` · optional (explicit presence)

Port of the Prometheus /metrics endpoint.

- default: `9090`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.router.ports.healthProbe

`int32` · optional (explicit presence)

Port of the /healthz and /readyz endpoints the pod's probes use.

- default: `8081`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.router.networkPolicy

`KubernetesAgentSandboxControllerRouterNetworkPolicy`

The router's NetworkPolicy.

### spec.router.networkPolicy.enabled

`bool` · optional (explicit presence)

Install the router's NetworkPolicy. False leaves the router's traffic
unrestricted.

- default: `true`

### spec.router.networkPolicy.clientNamespaceLabels

`map<string, string>`

Labels of the namespaces whose pods may send requests to the router,
for example {"kubernetes.io/metadata.name": "gateway"} for a Gateway's
namespace. Empty = every namespace, which is upstream's starting point;
narrow it to your callers.

### spec.router.networkPolicy.metricsNamespace

`string` · optional (explicit presence)

Namespace whose pods may scrape the router's metrics.

- default: `monitoring`
- rule: router.network_policy.metrics_namespace must be a Kubernetes namespace name: lowercase letters, digits and hyphens, starting and ending with a letter or digit
- rule: {"string":{"maxLen":"63"}}

### spec.router.networkPolicy.sandboxPorts

`[]int32`

Ports the router may open to sandboxes. A sandbox serving on any other
port is unreachable through the router behind this policy, so list every
port your sandboxes serve. Empty = 8888, upstream's default sandbox
port.

- rule: {"repeated":{"unique":true,"items":{"int32":{"lte":65535,"gte":1}}}}

### spec.clusterDomain

`string` · optional (explicit presence)

The cluster's DNS suffix. The controller builds each sandbox Service's
fully qualified name with it, and the router resolves sandboxes with it.
Change it only on a cluster whose DNS domain is not the default.

- default: `cluster.local`
- rule: cluster_domain must be a DNS domain such as "cluster.local": lowercase letters, digits, hyphens and dots

### spec.imageRegistry

`string`

Registry that replaces the registry part of every image this kind runs
(the controller and the router, both published on registry.k8s.io) -
the air-gap and private-mirror path. Each image keeps its path and tag,
so "mirror.example.com/k8s" pulls
registry.k8s.io/agent-sandbox/agent-sandbox-controller:v1.0.7 as
mirror.example.com/k8s/agent-sandbox/agent-sandbox-controller:v1.0.7.
An image override on the controller or router wins over this for that
image. Sandbox images are your own and never move. Empty =
registry.k8s.io.

- rule: image_registry is a registry root such as "mirror.example.com/k8s": no scheme and no trailing slash
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.imagePullSecrets

`[]string`

Names of image-pull Secrets in `agent-sandbox-system` for pulling the
controller and router images from a private registry.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesAgentSandboxController, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the controller and router run in: always `agent-sandbox-system`. Grant a product's ServiceAccount its rights on sandboxes and claims in its own namespaces; this one holds only the controller and the router. |
| `status.outputs.metrics_service_name` | `string` | Name of the Service in front of the controller's metrics endpoint (`agent-sandbox-controller`), the target of a ServiceMonitor. |
| `status.outputs.metrics_port` | `int32` | Port of the controller's metrics Service. |
| `status.outputs.router_service_name` | `string` | Name of the router's Service (`sandbox-router-svc`, the name upstream's SDKs assume). Empty when no router is installed. |
| `status.outputs.router_endpoint` | `string` | The URL a product or SDK dials to reach a sandbox through the router, for example "http://sandbox-router-svc.agent-sandbox-system.svc.cluster.local:8080", or the HTTPS listener's when the router serves TLS only. Empty when no router is installed. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.controller.metrics.tls.certificateSecret` | KubernetesCertificate | `status.outputs.secret_name` |
| `spec.controller.tracing.headers.secret` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.router.authorization.scopedToken.v1Secret.secret` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.router.authorization.scopedToken.verificationKeys.secret` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.router.tls.certificateSecret` | KubernetesCertificate | `status.outputs.secret_name` |
| `spec.router.tls.clientCa.secret` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.router.telemetry.headers.secret` | KubernetesSecret | `status.outputs.secret_name` |

## See Also

- [Overview](../README.md)
