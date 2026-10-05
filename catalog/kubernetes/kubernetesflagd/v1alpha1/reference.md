# KubernetesFlagd

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesFlagdSpec** runs flagd - the CNCF OpenFeature project's flag
evaluation daemon - as a central service: a Deployment and Service the
module owns end to end (flagd publishes no Helm chart; image
ghcr.io/open-feature/flagd). flagd reads flag definitions from one or
more sources and serves them over gRPC evaluation (8013), OFREP (8016)
and a sync stream in-process providers subscribe to (8015), with health
and Prometheus metrics on the management port (8014).

This component deploys the ENGINE. The flags themselves are DATA with
their own lifecycle: declare them as a KubernetesFlagdFlagFile (typed,
validated flags rendered into a ConfigMap) and add a `config_map` source,
or let flagd read flags over HTTP, gRPC, from object storage, or from the
OpenFeature Operator's FeatureFlag resources. A ConfigMap source is
mounted as a DIRECTORY (never a subPath mount, which would never see an
edit), so a flag flip reaches flagd when the kubelet syncs the volume -
typically one to two minutes - with no restart.

READINESS: flagd reports ready only after every source has synced once,
so a source that cannot be read keeps the pods unready and the apply
waiting.

SECRETS: HTTP authorization headers and credential-bearing headers are
written by the module into its own Secret, which holds flagd's source
configuration file; nothing credential-bearing renders into the
Deployment.

NOT MODELED, BY DESIGN (verified at flagd v0.17.0): the unix socket
listeners (`--socket-path`, `--sync-socket-path`) serve only a sidecar in
the same pod, never a Service; `--keep-alive-min-time` and
`--keep-alive-permit-without-stream` are no-ops kept for compatibility; the
HTTP source's OAuth `folder` and `reloadDelayS` re-read client credentials
from mounted files, where the module passes them directly;
the `file`, `fsnotify` and `fileinfo` providers are reached through
`config_map` sources (choose the watcher there); exposure is a Gateway API
route composed against the `service` output.

## Example

```yaml
# Full-surface development manifest - every source kind (a ConfigMap mounted
# as a directory, HTTP with an authorization header, gRPC with TLS and a CA,
# an OpenFeature Operator FeatureFlag resource, and the three object stores),
# server TLS, evaluation context and CORS, the OFREP stream, sync settings,
# OpenTelemetry with TLS, request limits, HPA, PDB, the full scheduling block,
# security contexts, extra environment and the ServiceMonitor. Every secret
# below is a development literal; a real manifest holds managed-secret
# references ($secret/<slug>) in these fields.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesFlagd
metadata:
  name: flagd-dev
spec:
  namespace:
    value: flagd-dev
  createNamespace: true
  image:
    repository: ghcr.io/open-feature/flagd
    tag: v0.17.0
    pullPolicy: IfNotPresent
    pullSecretNames:
      - registry-mirror
  replicas: 2
  resources:
    limits:
      cpu: 500m
      memory: 256Mi
    requests:
      cpu: 50m
      memory: 64Mi
  hpa:
    enabled: true
    minReplicas: 2
    maxReplicas: 6
    targetCpuUtilizationPercent: 70
    targetMemoryUtilizationPercent: 80
  pdb:
    enabled: true
    maxUnavailable: "1"
  sources:
    - configMap:
        configMapName:
          value: flagd-dev-flags
        key:
          value: flags.flagd.json
        watcher: fsnotify
    - http:
        url: https://flags.example.com/flags.json
        authHeader: Bearer dev-http-token
        headers:
          Accept: application/json
        sensitiveHeaders:
          X-Api-Key: dev-http-api-key
        intervalSeconds: 30
        intervalSeed: flagd-dev
        timeoutSeconds: 10
    - http:
        url: https://flags.example.com/team-flags.json
        oauth:
          clientId: flagd-dev
          clientSecret: dev-oauth-client-secret
          tokenUrl: https://idp.example.com/oauth/token
    - grpc:
        target: flag-service.flags.svc.cluster.local:8015
        tls: true
        caCertSecret:
          name: flag-service-ca
          key: ca.crt
        providerId: flagd-dev
        maxMsgSize: 5242880
        headers:
          x-env: dev
        sensitiveHeaders:
          authorization: Bearer dev-grpc-token
        selector: flagSetId=web
    - featureFlag:
        namespace: apps
        name: web-flags
    - googleStorage:
        bucket: acme-flags
        object: flags.json
        intervalSeconds: 60
    - azureBlob:
        bucket: flags
        object: flags.json
    - s3:
        bucket: acme-flags
        object: flags.yaml
        intervalSeed: flagd-dev
  server:
    port: 8013
    managementPort: 8014
    syncPort: 8015
    ofrepPort: 8016
    tlsSecretName: flagd-dev-tls
    serviceType: ClusterIP
  evaluation:
    contextValues:
      env: dev
    contextFromHeader:
      X-Tenant: tenant
    corsOrigins:
      - https://app.example.com
  ofrepSse:
    enabled: true
    inactivityDelaySeconds: 120
    publicUrl: https://flags.example.com
  sync:
    httpEnabled: true
    streamDeadline: 1h30m
  log:
    format: json
  telemetry:
    otelCollectorUri: otel-collector.observability:4317
    otelCaCertSecret:
      name: otel-ca
      key: ca.crt
    otelClientTlsSecretName: otel-client-tls
    otelReloadInterval: 30m
  limits:
    maxRequestBodyBytes: 1000000
    maxRequestHeaderBytes: 1000000
  scheduling:
    nodeSelector:
      kubernetes.io/os: linux
    tolerations:
      - key: dedicated
        operator: Equal
        value: platform
        effect: NoSchedule
    topologySpreadConstraints:
      - maxSkew: 1
        topologyKey: topology.kubernetes.io/zone
        whenUnsatisfiable: ScheduleAnyway
  serviceAccount:
    annotations:
      eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/flagd-dev
  podAnnotations:
    sidecar.istio.io/inject: "false"
  podLabels:
    team: platform
  podSecurityContext:
    runAsNonRoot: true
  containerSecurityContext:
    readOnlyRootFilesystem: true
    allowPrivilegeEscalation: false
    capabilities:
      drop:
        - ALL
  extraEnv:
    AWS_REGION: us-east-1
  extraEnvFromSecret:
    AZURE_STORAGE_KEY:
      name: azure-storage
      key: key
  metrics:
    serviceMonitorEnabled: true
    serviceMonitorLabels:
      release: kube-prometheus-stack
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.image` | `KubernetesFlagdImage` |  |  |  |
| `spec.image.repository` | `string` |  | `ghcr.io/open-feature/flagd` |  |
| `spec.image.tag` | `string` |  | `v0.17.0` |  |
| `spec.image.fips` | `bool` |  |  |  |
| `spec.image.pullPolicy` | `string` |  | `IfNotPresent` |  |
| `spec.image.pullSecretNames` | `[]string` |  |  |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.hpa` | `KubernetesFlagdHpa` |  |  |  |
| `spec.hpa.enabled` | `bool` |  |  |  |
| `spec.hpa.minReplicas` | `int32` |  | `1` |  |
| `spec.hpa.maxReplicas` | `int32` |  | `10` |  |
| `spec.hpa.targetCpuUtilizationPercent` | `int32` |  | `80` |  |
| `spec.hpa.targetMemoryUtilizationPercent` | `int32` |  |  |  |
| `spec.pdb` | `KubernetesFlagdPdb` |  |  |  |
| `spec.pdb.enabled` | `bool` |  |  |  |
| `spec.pdb.minAvailable` | `string` |  |  |  |
| `spec.pdb.maxUnavailable` | `string` |  |  |  |
| `spec.sources` | `[]KubernetesFlagdSource` | yes |  |  |
| `spec.sources[].configMap` | `KubernetesFlagdConfigMapSource` |  |  |  |
| `spec.sources[].configMap.configMapName` | `string \| valueFrom` | yes |  | KubernetesFlagdFlagFile (`status.outputs.config_map_name`) |
| `spec.sources[].configMap.key` | `string \| valueFrom` | yes |  | KubernetesFlagdFlagFile (`status.outputs.key`) |
| `spec.sources[].configMap.watcher` | `string` |  |  |  |
| `spec.sources[].http` | `KubernetesFlagdHttpSource` |  |  |  |
| `spec.sources[].http.url` | `string` | yes |  |  |
| `spec.sources[].http.authHeader` | `string` (sensitive) |  |  |  |
| `spec.sources[].http.headers` | `map<string, string>` |  |  |  |
| `spec.sources[].http.sensitiveHeaders` | `map<string, string>` (sensitive) |  |  |  |
| `spec.sources[].http.intervalSeconds` | `int32` |  |  |  |
| `spec.sources[].http.intervalSeed` | `string` |  |  |  |
| `spec.sources[].http.timeoutSeconds` | `int32` |  |  |  |
| `spec.sources[].http.oauth` | `KubernetesFlagdHttpOAuth` |  |  |  |
| `spec.sources[].http.oauth.clientId` | `string` | yes |  |  |
| `spec.sources[].http.oauth.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.sources[].http.oauth.tokenUrl` | `string` | yes |  |  |
| `spec.sources[].grpc` | `KubernetesFlagdGrpcSource` |  |  |  |
| `spec.sources[].grpc.target` | `string` | yes |  |  |
| `spec.sources[].grpc.tls` | `bool` |  |  |  |
| `spec.sources[].grpc.caCertSecret` | `KubernetesSecretKey` |  |  |  |
| `spec.sources[].grpc.caCertSecret.name` | `string` |  |  |  |
| `spec.sources[].grpc.caCertSecret.key` | `string` |  |  |  |
| `spec.sources[].grpc.providerId` | `string` |  |  |  |
| `spec.sources[].grpc.maxMsgSize` | `int32` |  |  |  |
| `spec.sources[].grpc.incrementalUpdates` | `bool` |  |  |  |
| `spec.sources[].grpc.headers` | `map<string, string>` |  |  |  |
| `spec.sources[].grpc.sensitiveHeaders` | `map<string, string>` (sensitive) |  |  |  |
| `spec.sources[].grpc.selector` | `string` |  |  |  |
| `spec.sources[].featureFlag` | `KubernetesFlagdFeatureFlagSource` |  |  |  |
| `spec.sources[].featureFlag.namespace` | `string` |  |  |  |
| `spec.sources[].featureFlag.name` | `string` | yes |  |  |
| `spec.sources[].googleStorage` | `KubernetesFlagdObjectSource` |  |  |  |
| `spec.sources[].googleStorage.bucket` | `string` | yes |  |  |
| `spec.sources[].googleStorage.object` | `string` | yes |  |  |
| `spec.sources[].googleStorage.intervalSeconds` | `int32` |  |  |  |
| `spec.sources[].googleStorage.intervalSeed` | `string` |  |  |  |
| `spec.sources[].azureBlob` | `KubernetesFlagdObjectSource` |  |  |  |
| `spec.sources[].azureBlob.bucket` | `string` | yes |  |  |
| `spec.sources[].azureBlob.object` | `string` | yes |  |  |
| `spec.sources[].azureBlob.intervalSeconds` | `int32` |  |  |  |
| `spec.sources[].azureBlob.intervalSeed` | `string` |  |  |  |
| `spec.sources[].s3` | `KubernetesFlagdObjectSource` |  |  |  |
| `spec.sources[].s3.bucket` | `string` | yes |  |  |
| `spec.sources[].s3.object` | `string` | yes |  |  |
| `spec.sources[].s3.intervalSeconds` | `int32` |  |  |  |
| `spec.sources[].s3.intervalSeed` | `string` |  |  |  |
| `spec.server` | `KubernetesFlagdServer` |  |  |  |
| `spec.server.port` | `int32` |  | `8013` |  |
| `spec.server.managementPort` | `int32` |  | `8014` |  |
| `spec.server.syncPort` | `int32` |  | `8015` |  |
| `spec.server.ofrepPort` | `int32` |  | `8016` |  |
| `spec.server.tlsSecretName` | `string` |  |  |  |
| `spec.server.serviceType` | `string` |  | `ClusterIP` |  |
| `spec.evaluation` | `KubernetesFlagdEvaluation` |  |  |  |
| `spec.evaluation.contextValues` | `map<string, string>` |  |  |  |
| `spec.evaluation.contextFromHeader` | `map<string, string>` |  |  |  |
| `spec.evaluation.corsOrigins` | `[]string` |  |  |  |
| `spec.ofrepSse` | `KubernetesFlagdOfrepSse` |  |  |  |
| `spec.ofrepSse.enabled` | `bool` |  | `true` |  |
| `spec.ofrepSse.inactivityDelaySeconds` | `int32` |  | `120` |  |
| `spec.ofrepSse.publicUrl` | `string` |  |  |  |
| `spec.sync` | `KubernetesFlagdSync` |  |  |  |
| `spec.sync.httpEnabled` | `bool` |  | `true` |  |
| `spec.sync.disableMetadata` | `bool` |  |  |  |
| `spec.sync.streamDeadline` | `string` |  |  |  |
| `spec.log` | `KubernetesFlagdLog` |  |  |  |
| `spec.log.format` | `string` |  | `json` |  |
| `spec.log.debug` | `bool` |  |  |  |
| `spec.telemetry` | `KubernetesFlagdTelemetry` |  |  |  |
| `spec.telemetry.metricsExporter` | `string` |  |  |  |
| `spec.telemetry.otelCollectorUri` | `string` |  |  |  |
| `spec.telemetry.otelCaCertSecret` | `KubernetesSecretKey` |  |  |  |
| `spec.telemetry.otelCaCertSecret.name` | `string` |  |  |  |
| `spec.telemetry.otelCaCertSecret.key` | `string` |  |  |  |
| `spec.telemetry.otelClientTlsSecretName` | `string` |  |  |  |
| `spec.telemetry.otelReloadInterval` | `string` |  |  |  |
| `spec.limits` | `KubernetesFlagdLimits` |  |  |  |
| `spec.limits.maxRequestBodyBytes` | `int64` |  | `1000000` |  |
| `spec.limits.maxRequestHeaderBytes` | `int64` |  | `1000000` |  |
| `spec.scheduling` | `WorkloadScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.scheduling.nodeAffinity` | `WorkloadNodeAffinity` |  |  |  |
| `spec.scheduling.nodeAffinity.required` | `[]WorkloadNodeSelectorTerm` |  |  |  |
| `spec.scheduling.nodeAffinity.required[].matchExpressions` | `[]WorkloadNodeSelectorRequirement` | yes |  |  |
| `spec.scheduling.nodeAffinity.required[].matchExpressions[].key` | `string` | yes |  |  |
| `spec.scheduling.nodeAffinity.required[].matchExpressions[].operator` | `string` | yes |  |  |
| `spec.scheduling.nodeAffinity.required[].matchExpressions[].values` | `[]string` |  |  |  |
| `spec.scheduling.nodeAffinity.preferred` | `[]WorkloadPreferredNodeSelectorTerm` |  |  |  |
| `spec.scheduling.nodeAffinity.preferred[].weight` | `int32` |  |  |  |
| `spec.scheduling.nodeAffinity.preferred[].term` | `WorkloadNodeSelectorTerm` | yes |  |  |
| `spec.scheduling.nodeAffinity.preferred[].term.matchExpressions` | `[]WorkloadNodeSelectorRequirement` | yes |  |  |
| `spec.scheduling.nodeAffinity.preferred[].term.matchExpressions[].key` | `string` | yes |  |  |
| `spec.scheduling.nodeAffinity.preferred[].term.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.scheduling.nodeAffinity.preferred[].term.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.scheduling.podAffinity` | `WorkloadPodAffinity` |  |  |  |
| `spec.scheduling.podAffinity.required` | `[]WorkloadPodAffinityTerm` |  |  |  |
| `spec.scheduling.podAffinity.required[].matchLabels` | `map<string, string>` | yes |  |  |
| `spec.scheduling.podAffinity.required[].topologyKey` | `string` | yes |  |  |
| `spec.scheduling.podAffinity.required[].namespaces` | `[]string` |  |  |  |
| `spec.scheduling.podAffinity.preferred` | `[]WorkloadWeightedPodAffinityTerm` |  |  |  |
| `spec.scheduling.podAffinity.preferred[].weight` | `int32` |  |  |  |
| `spec.scheduling.podAffinity.preferred[].term` | `WorkloadPodAffinityTerm` | yes |  |  |
| `spec.scheduling.podAffinity.preferred[].term.matchLabels` | `map<string, string>` | yes |  |  |
| `spec.scheduling.podAffinity.preferred[].term.topologyKey` | `string` | yes |  |  |
| `spec.scheduling.podAffinity.preferred[].term.namespaces` | `[]string` |  |  |  |
| `spec.scheduling.podAntiAffinity` | `WorkloadPodAffinity` |  |  |  |
| `spec.scheduling.podAntiAffinity.required` | `[]WorkloadPodAffinityTerm` |  |  |  |
| `spec.scheduling.podAntiAffinity.required[].matchLabels` | `map<string, string>` | yes |  |  |
| `spec.scheduling.podAntiAffinity.required[].topologyKey` | `string` | yes |  |  |
| `spec.scheduling.podAntiAffinity.required[].namespaces` | `[]string` |  |  |  |
| `spec.scheduling.podAntiAffinity.preferred` | `[]WorkloadWeightedPodAffinityTerm` |  |  |  |
| `spec.scheduling.podAntiAffinity.preferred[].weight` | `int32` |  |  |  |
| `spec.scheduling.podAntiAffinity.preferred[].term` | `WorkloadPodAffinityTerm` | yes |  |  |
| `spec.scheduling.podAntiAffinity.preferred[].term.matchLabels` | `map<string, string>` | yes |  |  |
| `spec.scheduling.podAntiAffinity.preferred[].term.topologyKey` | `string` | yes |  |  |
| `spec.scheduling.podAntiAffinity.preferred[].term.namespaces` | `[]string` |  |  |  |
| `spec.scheduling.topologySpreadConstraints` | `[]WorkloadTopologySpreadConstraint` |  |  |  |
| `spec.scheduling.topologySpreadConstraints[].maxSkew` | `int32` |  |  |  |
| `spec.scheduling.topologySpreadConstraints[].topologyKey` | `string` | yes |  |  |
| `spec.scheduling.topologySpreadConstraints[].whenUnsatisfiable` | `string` | yes |  |  |
| `spec.scheduling.topologySpreadConstraints[].matchLabels` | `map<string, string>` |  |  |  |
| `spec.scheduling.schedulerName` | `string` |  |  |  |
| `spec.serviceAccount` | `KubernetesFlagdServiceAccount` |  |  |  |
| `spec.serviceAccount.annotations` | `map<string, string>` |  |  |  |
| `spec.serviceAccount.existingName` | `string` |  |  |  |
| `spec.podAnnotations` | `map<string, string>` |  |  |  |
| `spec.podLabels` | `map<string, string>` |  |  |  |
| `spec.podSecurityContext` | `WorkloadPodSecurityContext` |  |  |  |
| `spec.podSecurityContext.runAsUser` | `int64` |  |  |  |
| `spec.podSecurityContext.runAsGroup` | `int64` |  |  |  |
| `spec.podSecurityContext.runAsNonRoot` | `bool` |  |  |  |
| `spec.podSecurityContext.fsGroup` | `int64` |  |  |  |
| `spec.podSecurityContext.fsGroupChangePolicy` | `string` |  |  |  |
| `spec.podSecurityContext.supplementalGroups` | `[]int64` |  |  |  |
| `spec.podSecurityContext.sysctls` | `[]WorkloadSysctl` |  |  |  |
| `spec.podSecurityContext.sysctls[].name` | `string` | yes |  |  |
| `spec.podSecurityContext.sysctls[].value` | `string` | yes |  |  |
| `spec.podSecurityContext.seccompProfile` | `WorkloadSeccompProfile` |  |  |  |
| `spec.podSecurityContext.seccompProfile.type` | `string` | yes |  |  |
| `spec.podSecurityContext.seccompProfile.localhostProfile` | `string` |  |  |  |
| `spec.containerSecurityContext` | `WorkloadContainerSecurityContext` |  |  |  |
| `spec.containerSecurityContext.privileged` | `bool` |  |  |  |
| `spec.containerSecurityContext.runAsUser` | `int64` |  |  |  |
| `spec.containerSecurityContext.runAsGroup` | `int64` |  |  |  |
| `spec.containerSecurityContext.runAsNonRoot` | `bool` |  |  |  |
| `spec.containerSecurityContext.readOnlyRootFilesystem` | `bool` |  |  |  |
| `spec.containerSecurityContext.allowPrivilegeEscalation` | `bool` |  |  |  |
| `spec.containerSecurityContext.capabilities` | `WorkloadCapabilities` |  |  |  |
| `spec.containerSecurityContext.capabilities.add` | `[]string` |  |  |  |
| `spec.containerSecurityContext.capabilities.drop` | `[]string` |  |  |  |
| `spec.containerSecurityContext.seccompProfile` | `WorkloadSeccompProfile` |  |  |  |
| `spec.containerSecurityContext.seccompProfile.type` | `string` | yes |  |  |
| `spec.containerSecurityContext.seccompProfile.localhostProfile` | `string` |  |  |  |
| `spec.extraEnv` | `map<string, string>` |  |  |  |
| `spec.extraEnvFromSecret` | `map<string, KubernetesSecretKey>` |  |  |  |
| `spec.extraEnvFromSecret.*.name` | `string` |  |  |  |
| `spec.extraEnvFromSecret.*.key` | `string` |  |  |  |
| `spec.metrics` | `KubernetesFlagdMetrics` |  |  |  |
| `spec.metrics.serviceMonitorEnabled` | `bool` |  |  |  |
| `spec.metrics.serviceMonitorLabels` | `map<string, string>` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install into. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource. ConfigMap sources and
referenced Secrets must live in this same namespace (pod volumes cannot
cross namespaces).

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the resource.
When false, the namespace must already exist.

### spec.image

`KubernetesFlagdImage`

flagd container image.

### spec.image.repository

`string` · optional (explicit presence)

Image repository.

- default: `ghcr.io/open-feature/flagd`

### spec.image.tag

`string` · optional (explicit presence)

Image tag (a flagd release, e.g. "v0.17.0").

- default: `v0.17.0`

### spec.image.fips

`bool`

Run the FIPS 140-3 build (the module appends `-fips` to the tag).

### spec.image.pullPolicy

`string` · optional (explicit presence)

Image pull policy: Always, IfNotPresent or Never.

- default: `IfNotPresent`
- rule: Image pull policy must be one of: Always, IfNotPresent, Never.

### spec.image.pullSecretNames

`[]string`

Names of existing image pull Secrets in flagd's namespace.

### spec.replicas

`int32` · optional (explicit presence)

flagd replicas. flagd is stateless (every replica reads the same
sources), so replicas add availability and evaluation throughput.
Ignored when hpa is enabled.

- default: `1`
- rule: {"int32":{"lte":50,"gte":1}}

### spec.resources

`ContainerResources`

CPU and memory for the flagd container. flagd holds every flag in
memory; these defaults suit hundreds of flags at moderate traffic.

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

### spec.hpa

`KubernetesFlagdHpa`

Horizontal pod autoscaling on CPU and memory utilization.

- rule: The autoscaler's minimum replica count cannot exceed its maximum.

### spec.hpa.enabled

`bool`

Enable the HorizontalPodAutoscaler (replicas then belongs to the
autoscaler).

### spec.hpa.minReplicas

`int32` · optional (explicit presence)

Minimum replicas.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.hpa.maxReplicas

`int32` · optional (explicit presence)

Maximum replicas.

- default: `10`
- rule: {"int32":{"gte":1}}

### spec.hpa.targetCpuUtilizationPercent

`int32` · optional (explicit presence)

Target average CPU utilization percent.

- default: `80`
- rule: {"int32":{"lte":100,"gte":1}}

### spec.hpa.targetMemoryUtilizationPercent

`int32` · optional (explicit presence)

Target average memory utilization percent. Empty = memory does not
drive scaling.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.pdb

`KubernetesFlagdPdb`

PodDisruptionBudget for voluntary disruptions.

- rule: A PodDisruptionBudget takes either min_available or max_unavailable, not both.

### spec.pdb.enabled

`bool`

Create the PodDisruptionBudget. It only protects availability when
more than one replica runs.

### spec.pdb.minAvailable

`string`

Minimum pods that must stay available: an integer ("1") or a
percentage ("50%"). Empty with max_unavailable also empty = 1 - which
with a single replica blocks every voluntary eviction (node drains
wait), so pair the budget with two or more replicas.

- rule: min_available must be a non-negative integer or a percentage up to 100%, such as 50%.

### spec.pdb.maxUnavailable

`string`

Maximum pods that may be unavailable: an integer or a percentage.
Mutually exclusive with min_available.

- rule: max_unavailable must be a non-negative integer or a percentage up to 100%, such as 25%.

### spec.sources

`[]KubernetesFlagdSource` · required

Where flag definitions come from. When several sources define the same
flag key, the source listed later wins.

- rule: {"repeated":{"minItems":"1"}}

### spec.sources[].configMap

`KubernetesFlagdConfigMapSource`

A key of a ConfigMap in flagd's namespace, mounted as a directory -
the way to serve a KubernetesFlagdFlagFile.

### spec.sources[].configMap.configMapName

`string | valueFrom` · required

ConfigMap name, in flagd's namespace. Accepts a literal or a reference
to a KubernetesFlagdFlagFile (its rendered ConfigMap). The ConfigMap
must exist: a missing one keeps the pods from starting.

- references: KubernetesFlagdFlagFile (`status.outputs.config_map_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesFlagdFlagFile, name: <that resource's name>, fieldPath: status.outputs.config_map_name}} -- a bare string does not parse

### spec.sources[].configMap.key

`string | valueFrom` · required

ConfigMap data key holding the flag definitions; it must end in .json,
.yaml or .yml (flagd picks the parser from the extension). Accepts a
literal or a reference to the same KubernetesFlagdFlagFile (its
rendered key, `flags.flagd.json` unless the flag file says otherwise).

- references: KubernetesFlagdFlagFile (`status.outputs.key`)
- rule: The key must end in .json, .yaml or .yml.
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesFlagdFlagFile, name: <that resource's name>, fieldPath: status.outputs.key}} -- a bare string does not parse

### spec.sources[].configMap.watcher

`string`

How flagd notices edits: fsnotify (file events - flagd's choice inside
Kubernetes) or fileinfo (polls the file's metadata). Empty = flagd
picks (fsnotify in Kubernetes).

- rule: Watcher must be fsnotify or fileinfo.

### spec.sources[].http

`KubernetesFlagdHttpSource`

A flag definition document served over HTTP(S), polled.

- rule: Authenticate an HTTP source one way: auth_header or oauth.

### spec.sources[].http.url

`string` · required

URL of the flag definition document.

- rule: The URL must be http(s).
- rule: {"required":true,"string":{"uri":true}}

### spec.sources[].http.authHeader

`string` · sensitive

The full Authorization header value (e.g. "Bearer <token>" or
"Basic <base64>").

### spec.sources[].http.headers

`map<string, string>`

Plain request headers.

### spec.sources[].http.sensitiveHeaders

`map<string, string>` · sensitive

Request headers carrying credentials. They may not repeat a plain
header.

### spec.sources[].http.intervalSeconds

`int32` · optional (explicit presence)

Poll interval in seconds (maximum 86400). Empty = 5.

- rule: {"int32":{"lte":86400,"gt":0}}

### spec.sources[].http.intervalSeed

`string`

Seed that offsets this source's poll schedule within the interval, so
this flagd does not poll in lockstep with other flagd deployments
reading the same document (every replica of one deployment shares the
seed and polls together). Empty = no offset.

### spec.sources[].http.timeoutSeconds

`int32` · optional (explicit presence)

Request timeout in seconds. Empty = flagd's default.

- rule: {"int32":{"gt":0}}

### spec.sources[].http.oauth

`KubernetesFlagdHttpOAuth`

OAuth 2.0 client credentials: flagd fetches a token from token_url and
sends it as the Authorization header. Mutually exclusive with
auth_header.

### spec.sources[].http.oauth.clientId

`string` · required

OAuth client id.

- rule: {"required":true}

### spec.sources[].http.oauth.clientSecret

`string` · required · sensitive

OAuth client secret.

- rule: {"required":true}

### spec.sources[].http.oauth.tokenUrl

`string` · required

Token endpoint URL.

- rule: token_url must be an absolute http(s) URL.
- rule: {"required":true}

### spec.sources[].grpc

`KubernetesFlagdGrpcSource`

A gRPC server implementing flagd's sync protocol (another flagd, or
a custom flag service).

- rule: A CA certificate only applies to a TLS connection - set tls: true.

### spec.sources[].grpc.target

`string` · required

Target address: host:port, or an envoy://, dns://, uds:// or xds://
target.

- rule: {"required":true}

### spec.sources[].grpc.tls

`bool`

Connect with TLS.

### spec.sources[].grpc.caCertSecret

`KubernetesSecretKey`

CA certificate for the TLS connection, read from an existing Secret in
flagd's namespace (mounted into the pod). Empty = the system trust
store.

### spec.sources[].grpc.caCertSecret.name

`string`

The name of the Kubernetes Secret.

### spec.sources[].grpc.caCertSecret.key

`string`

The key within the Kubernetes Secret.

### spec.sources[].grpc.providerId

`string`

Provider id sent to the server (servers may use it to identify the
connecting flagd).

### spec.sources[].grpc.maxMsgSize

`int32` · optional (explicit presence)

Maximum message size in bytes. Empty = 4194304 (4 MB).

- rule: {"int32":{"gt":0}}

### spec.sources[].grpc.incrementalUpdates

`bool`

Experimental: each update replaces only the flag sets in its payload,
so flags from other flag sets accumulate.

### spec.sources[].grpc.headers

`map<string, string>`

Plain gRPC metadata.

### spec.sources[].grpc.sensitiveHeaders

`map<string, string>` · sensitive

gRPC metadata carrying credentials. It may not repeat a plain header.

### spec.sources[].grpc.selector

`string`

Selector the server filters the synced flags by (`flagSetId=<id>` or
`source=<name>`). Empty = every flag the server offers.

### spec.sources[].featureFlag

`KubernetesFlagdFeatureFlagSource`

A FeatureFlag custom resource of the OpenFeature Operator (the
operator's CRDs must be installed). The module grants flagd read
access to FeatureFlag resources in the named namespace.

### spec.sources[].featureFlag.namespace

`string`

Namespace of the FeatureFlag resource. Empty = flagd's namespace.

### spec.sources[].featureFlag.name

`string` · required

Name of the FeatureFlag resource.

- rule: {"required":true}

### spec.sources[].googleStorage

`KubernetesFlagdObjectSource`

An object in a Google Cloud Storage bucket, polled (credentials from
workload identity or GOOGLE_APPLICATION_CREDENTIALS).

### spec.sources[].googleStorage.bucket

`string` · required

Bucket (GCS, S3) or container (Azure Blob).

- rule: {"required":true}

### spec.sources[].googleStorage.object

`string` · required

Object or blob name of the flag definitions. flagd picks the parser by
the extension (.json, .yaml or .yml, any case); a name without an
extension is parsed by the object's content type (application/json or
application/yaml).

- rule: An object's extension must be .json, .yaml or .yml (or none, parsed by its content type) - flagd picks the parser by it.
- rule: {"required":true}

### spec.sources[].googleStorage.intervalSeconds

`int32` · optional (explicit presence)

Poll interval in seconds (maximum 86400). Empty = 5.

- rule: {"int32":{"lte":86400,"gt":0}}

### spec.sources[].googleStorage.intervalSeed

`string`

Seed that offsets this source's poll schedule within the interval,
against other flagd deployments reading the same object. Empty = no
offset.

### spec.sources[].azureBlob

`KubernetesFlagdObjectSource`

A blob in Azure Blob Storage, polled (account and credentials from
the AZURE_STORAGE_* environment).

### spec.sources[].azureBlob.bucket

`string` · required

Bucket (GCS, S3) or container (Azure Blob).

- rule: {"required":true}

### spec.sources[].azureBlob.object

`string` · required

Object or blob name of the flag definitions. flagd picks the parser by
the extension (.json, .yaml or .yml, any case); a name without an
extension is parsed by the object's content type (application/json or
application/yaml).

- rule: An object's extension must be .json, .yaml or .yml (or none, parsed by its content type) - flagd picks the parser by it.
- rule: {"required":true}

### spec.sources[].azureBlob.intervalSeconds

`int32` · optional (explicit presence)

Poll interval in seconds (maximum 86400). Empty = 5.

- rule: {"int32":{"lte":86400,"gt":0}}

### spec.sources[].azureBlob.intervalSeed

`string`

Seed that offsets this source's poll schedule within the interval,
against other flagd deployments reading the same object. Empty = no
offset.

### spec.sources[].s3

`KubernetesFlagdObjectSource`

An object in an S3 bucket, polled (credentials from the standard AWS
SDK chain).

### spec.sources[].s3.bucket

`string` · required

Bucket (GCS, S3) or container (Azure Blob).

- rule: {"required":true}

### spec.sources[].s3.object

`string` · required

Object or blob name of the flag definitions. flagd picks the parser by
the extension (.json, .yaml or .yml, any case); a name without an
extension is parsed by the object's content type (application/json or
application/yaml).

- rule: An object's extension must be .json, .yaml or .yml (or none, parsed by its content type) - flagd picks the parser by it.
- rule: {"required":true}

### spec.sources[].s3.intervalSeconds

`int32` · optional (explicit presence)

Poll interval in seconds (maximum 86400). Empty = 5.

- rule: {"int32":{"lte":86400,"gt":0}}

### spec.sources[].s3.intervalSeed

`string`

Seed that offsets this source's poll schedule within the interval,
against other flagd deployments reading the same object. Empty = no
offset.

### spec.server

`KubernetesFlagdServer`

Listener ports, TLS and the Service.

- rule: flagd's four ports must all differ.

### spec.server.port

`int32` · optional (explicit presence)

gRPC evaluation port (also serves Connect over HTTP).

- default: `8013`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.server.managementPort

`int32` · optional (explicit presence)

Management port: /healthz, /readyz and /metrics.

- default: `8014`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.server.syncPort

`int32` · optional (explicit presence)

Flag sync service port (gRPC, gRPC-Web and Connect), for in-process
providers.

- default: `8015`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.server.ofrepPort

`int32` · optional (explicit presence)

OFREP port (OpenFeature Remote Evaluation Protocol over HTTP).

- default: `8016`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.server.tlsSecretName

`string`

Serve the evaluation and sync listeners over TLS, from an existing
kubernetes.io/tls Secret in flagd's namespace (keys tls.crt and
tls.key; e.g. one a cert-manager Certificate maintains). The OFREP and
management listeners stay plain HTTP - terminate TLS for them at a
gateway.

### spec.server.serviceType

`string` · optional (explicit presence)

Service type: ClusterIP, NodePort or LoadBalancer.

- default: `ClusterIP`
- rule: Service type must be one of: ClusterIP, NodePort, LoadBalancer.

### spec.evaluation

`KubernetesFlagdEvaluation`

Evaluation behavior: static context, header-to-context mapping, CORS.

### spec.evaluation.contextValues

`map<string, string>`

Static key-value pairs added to every evaluation context (e.g. the
environment name, for targeting).

- rule: Context keys and values may not contain ',', '=' or '"' (flagd parses the pairs as comma-separated key=value text).

### spec.evaluation.contextFromHeader

`map<string, string>`

Request headers copied into the evaluation context: header name ->
context key (e.g. X-Tenant -> tenant).

- rule: Header names and context keys may not contain ',', '=' or '"' (flagd parses the pairs as comma-separated key=value text).

### spec.evaluation.corsOrigins

`[]string`

Origins allowed to call flagd from a browser (CORS). Empty = ANY origin
(flagd always installs its CORS handler, and an empty list allows
every origin) - list the origins to restrict browser access.

### spec.ofrepSse

`KubernetesFlagdOfrepSse`

The OFREP flag-change event stream (server-sent events on the OFREP
port).

### spec.ofrepSse.enabled

`bool` · optional (explicit presence)

Serve the change-notification stream at /ofrep/v1/sse/{channel}.

- default: `true`

### spec.ofrepSse.inactivityDelaySeconds

`int32` · optional (explicit presence)

Seconds after which clients close an idle stream.

- default: `120`
- rule: {"int32":{"gt":0}}

### spec.ofrepSse.publicUrl

`string`

Public origin (scheme://host) advertised for the stream, when clients
reach flagd through a proxy or gateway. Empty = clients resolve it
against the OFREP base URL.

- rule: public_url must be an absolute http(s) origin.

### spec.sync

`KubernetesFlagdSync`

The flag sync service in-process providers subscribe to.

### spec.sync.httpEnabled

`bool` · optional (explicit presence)

Serve the flag configuration document over HTTP at /v1/flags on the
sync port.

- default: `true`

### spec.sync.disableMetadata

`bool`

Disable the sync service's metadata endpoint.

### spec.sync.streamDeadline

`string`

Server-side deadline for sync and event streams (Go duration, e.g.
"1h"). Empty = no deadline.

- rule: stream_deadline must be a Go duration such as 30m or 1h.

### spec.log

`KubernetesFlagdLog`

Logging.

### spec.log.format

`string` · optional (explicit presence)

Log format: json or console.

- default: `json`
- rule: Log format must be json or console.

### spec.log.debug

`bool`

Verbose debug logging.

### spec.telemetry

`KubernetesFlagdTelemetry`

Metrics and traces.

- rule: The otel metrics exporter pushes to a collector - set otel_collector_uri.

### spec.telemetry.metricsExporter

`string`

Metrics exporter: prometheus (served on the management port) or otel
(pushed to otel_collector_uri). Empty = prometheus.

- rule: metrics_exporter must be prometheus or otel.

### spec.telemetry.otelCollectorUri

`string`

gRPC URI of an OpenTelemetry collector for traces (and metrics when
metrics_exporter is otel). Empty = traces are not exported.

### spec.telemetry.otelCaCertSecret

`KubernetesSecretKey`

CA certificate the collector's TLS certificate is verified against,
read from an existing Secret in flagd's namespace (mounted into the
pod). Empty = the system trust store.

### spec.telemetry.otelCaCertSecret.name

`string`

The name of the Kubernetes Secret.

### spec.telemetry.otelCaCertSecret.key

`string`

The key within the Kubernetes Secret.

### spec.telemetry.otelClientTlsSecretName

`string`

Client certificate for mutual TLS with the collector: an existing
kubernetes.io/tls Secret in flagd's namespace (keys tls.crt and
tls.key).

### spec.telemetry.otelReloadInterval

`string`

How often the collector TLS files are reloaded (Go duration). Empty =
1h.

- rule: otel_reload_interval must be a Go duration such as 30m or 1h.

### spec.limits

`KubernetesFlagdLimits`

Request size limits.

### spec.limits.maxRequestBodyBytes

`int64` · optional (explicit presence)

Maximum request body in bytes (HTTP 413 / 429 above it). 0 disables
the limit, which exposes flagd to memory exhaustion.

- default: `1000000`
- rule: {"int64":{"gte":"0"}}

### spec.limits.maxRequestHeaderBytes

`int64` · optional (explicit presence)

Maximum request header size in bytes (HTTP 431 above it). 0 = Go's
built-in default (1 MiB).

- default: `1000000`
- rule: {"int64":{"gte":"0"}}

### spec.scheduling

`WorkloadScheduling`

Pod scheduling: node selection, tolerations, affinity and topology
spread.

### spec.scheduling.nodeSelector

`map<string, string>`

Simple hard node filter: every listed label must match the node. The right tool
for "run on the GPU pool" — reach for node_affinity only when you need operators
(In/NotIn/Exists) or soft preferences.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Taint tolerations. A toleration does not attract pods to tainted nodes — it only
permits scheduling there; pair with node_selector or affinity to target them.

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

### spec.scheduling.nodeAffinity

`WorkloadNodeAffinity`

Expressive node selection: hard requirements and weighted soft preferences over
node labels.

### spec.scheduling.nodeAffinity.required

`[]WorkloadNodeSelectorTerm`

Hard requirement. The outer list ORs its terms; expressions within one term AND.

### spec.scheduling.nodeAffinity.required[].matchExpressions

`[]WorkloadNodeSelectorRequirement` · required

- rule: {"repeated":{"minItems":"1"}}
- rule: In/NotIn require at least one value, Gt/Lt exactly one, Exists/DoesNotExist none

### spec.scheduling.nodeAffinity.required[].matchExpressions[].key

`string` · required

Node label key, e.g. "topology.kubernetes.io/zone".

- rule: {"required":true}

### spec.scheduling.nodeAffinity.required[].matchExpressions[].operator

`string` · required

Operator: "In"/"NotIn" (value set), "Exists"/"DoesNotExist" (key presence), or
"Gt"/"Lt" (single integer value, as strings — the Kubernetes API convention).

- rule: Operator must be one of "In", "NotIn", "Exists", "DoesNotExist", "Gt", or "Lt"
- rule: {"required":true}

### spec.scheduling.nodeAffinity.required[].matchExpressions[].values

`[]string`

Values for the operator: required non-empty for In/NotIn, exactly one integer
string for Gt/Lt, and must be empty for Exists/DoesNotExist.

### spec.scheduling.nodeAffinity.preferred

`[]WorkloadPreferredNodeSelectorTerm`

Weighted soft preferences.

### spec.scheduling.nodeAffinity.preferred[].weight

`int32`

Preference weight, 1–100. Higher weights dominate placement scoring.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.scheduling.nodeAffinity.preferred[].term

`WorkloadNodeSelectorTerm` · required

- rule: {"required":true}

### spec.scheduling.nodeAffinity.preferred[].term.matchExpressions

`[]WorkloadNodeSelectorRequirement` · required

- rule: {"repeated":{"minItems":"1"}}
- rule: In/NotIn require at least one value, Gt/Lt exactly one, Exists/DoesNotExist none

### spec.scheduling.nodeAffinity.preferred[].term.matchExpressions[].key

`string` · required

Node label key, e.g. "topology.kubernetes.io/zone".

- rule: {"required":true}

### spec.scheduling.nodeAffinity.preferred[].term.matchExpressions[].operator

`string` · required

Operator: "In"/"NotIn" (value set), "Exists"/"DoesNotExist" (key presence), or
"Gt"/"Lt" (single integer value, as strings — the Kubernetes API convention).

- rule: Operator must be one of "In", "NotIn", "Exists", "DoesNotExist", "Gt", or "Lt"
- rule: {"required":true}

### spec.scheduling.nodeAffinity.preferred[].term.matchExpressions[].values

`[]string`

Values for the operator: required non-empty for In/NotIn, exactly one integer
string for Gt/Lt, and must be empty for Exists/DoesNotExist.

### spec.scheduling.podAffinity

`WorkloadPodAffinity`

Attract pods toward nodes/zones already running matching pods (co-location with
a cache, for example).

### spec.scheduling.podAffinity.required

`[]WorkloadPodAffinityTerm`

Hard rules — unschedulable until satisfied. Use sparingly; they can deadlock rollouts.

### spec.scheduling.podAffinity.required[].matchLabels

`map<string, string>` · required

Labels of the pods to match against — for self-anti-affinity, the workload's own
selector labels (exported as the `selector_labels` output).

- rule: {"map":{"minPairs":"1"}}

### spec.scheduling.podAffinity.required[].topologyKey

`string` · required

Node label defining the domain: "kubernetes.io/hostname" separates by node,
"topology.kubernetes.io/zone" by zone.

- rule: {"required":true}

### spec.scheduling.podAffinity.required[].namespaces

`[]string`

Namespaces whose pods are considered. Empty means the workload's own namespace.

### spec.scheduling.podAffinity.preferred

`[]WorkloadWeightedPodAffinityTerm`

Weighted soft rules — the scheduler's tiebreakers.

### spec.scheduling.podAffinity.preferred[].weight

`int32`

Preference weight, 1–100.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.scheduling.podAffinity.preferred[].term

`WorkloadPodAffinityTerm` · required

- rule: {"required":true}

### spec.scheduling.podAffinity.preferred[].term.matchLabels

`map<string, string>` · required

Labels of the pods to match against — for self-anti-affinity, the workload's own
selector labels (exported as the `selector_labels` output).

- rule: {"map":{"minPairs":"1"}}

### spec.scheduling.podAffinity.preferred[].term.topologyKey

`string` · required

Node label defining the domain: "kubernetes.io/hostname" separates by node,
"topology.kubernetes.io/zone" by zone.

- rule: {"required":true}

### spec.scheduling.podAffinity.preferred[].term.namespaces

`[]string`

Namespaces whose pods are considered. Empty means the workload's own namespace.

### spec.scheduling.podAntiAffinity

`WorkloadPodAffinity`

Repel pods from nodes/zones already running matching pods — the classic
high-availability pattern is anti-affinity on the workload's own labels across
`kubernetes.io/hostname`.

### spec.scheduling.podAntiAffinity.required

`[]WorkloadPodAffinityTerm`

Hard rules — unschedulable until satisfied. Use sparingly; they can deadlock rollouts.

### spec.scheduling.podAntiAffinity.required[].matchLabels

`map<string, string>` · required

Labels of the pods to match against — for self-anti-affinity, the workload's own
selector labels (exported as the `selector_labels` output).

- rule: {"map":{"minPairs":"1"}}

### spec.scheduling.podAntiAffinity.required[].topologyKey

`string` · required

Node label defining the domain: "kubernetes.io/hostname" separates by node,
"topology.kubernetes.io/zone" by zone.

- rule: {"required":true}

### spec.scheduling.podAntiAffinity.required[].namespaces

`[]string`

Namespaces whose pods are considered. Empty means the workload's own namespace.

### spec.scheduling.podAntiAffinity.preferred

`[]WorkloadWeightedPodAffinityTerm`

Weighted soft rules — the scheduler's tiebreakers.

### spec.scheduling.podAntiAffinity.preferred[].weight

`int32`

Preference weight, 1–100.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.scheduling.podAntiAffinity.preferred[].term

`WorkloadPodAffinityTerm` · required

- rule: {"required":true}

### spec.scheduling.podAntiAffinity.preferred[].term.matchLabels

`map<string, string>` · required

Labels of the pods to match against — for self-anti-affinity, the workload's own
selector labels (exported as the `selector_labels` output).

- rule: {"map":{"minPairs":"1"}}

### spec.scheduling.podAntiAffinity.preferred[].term.topologyKey

`string` · required

Node label defining the domain: "kubernetes.io/hostname" separates by node,
"topology.kubernetes.io/zone" by zone.

- rule: {"required":true}

### spec.scheduling.podAntiAffinity.preferred[].term.namespaces

`[]string`

Namespaces whose pods are considered. Empty means the workload's own namespace.

### spec.scheduling.topologySpreadConstraints

`[]WorkloadTopologySpreadConstraint`

Even distribution of replicas across topology domains (zones, hosts). Preferred
over hostname anti-affinity for large replica counts because skew is bounded
rather than binary.

### spec.scheduling.topologySpreadConstraints[].maxSkew

`int32`

Maximum allowed difference in matching-pod counts between any two domains.
1 is the strictest even spread.

- rule: {"int32":{"gte":1}}

### spec.scheduling.topologySpreadConstraints[].topologyKey

`string` · required

Node label defining the domains to spread across (e.g.
"topology.kubernetes.io/zone").

- rule: {"required":true}

### spec.scheduling.topologySpreadConstraints[].whenUnsatisfiable

`string` · required

What happens when the constraint cannot be met: "DoNotSchedule" (hard — pod
stays Pending) or "ScheduleAnyway" (soft — scheduler minimizes skew).

- rule: whenUnsatisfiable must be either "DoNotSchedule" or "ScheduleAnyway"
- rule: {"required":true}

### spec.scheduling.topologySpreadConstraints[].matchLabels

`map<string, string>`

Labels selecting the pods counted per domain. Omit to have the module default to
the workload's own selector labels — self-spreading, the overwhelmingly common
intent.

### spec.scheduling.schedulerName

`string`

Hand pods to a non-default scheduler installed in the cluster. Leave empty for
the standard scheduler.

### spec.serviceAccount

`KubernetesFlagdServiceAccount`

flagd's ServiceAccount.

### spec.serviceAccount.annotations

`map<string, string>`

Annotations on the ServiceAccount the module creates - the cloud
workload-identity seam for the GCS, S3 and Azure Blob sources.

### spec.serviceAccount.existingName

`string`

Run as an existing ServiceAccount instead of creating one (its
annotations are then yours to manage). The module binds FeatureFlag
reads to whichever account runs flagd.

### spec.podAnnotations

`map<string, string>`

Annotations added to the flagd pods.

### spec.podLabels

`map<string, string>`

Labels added to the flagd pods.

### spec.podSecurityContext

`WorkloadPodSecurityContext`

Pod-level security context. The flagd image already runs as the
non-root user 65532.

### spec.podSecurityContext.runAsUser

`int64` · optional (explicit presence)

UID all container processes run as unless overridden per container.

### spec.podSecurityContext.runAsGroup

`int64` · optional (explicit presence)

Primary GID all container processes run as unless overridden per container.

### spec.podSecurityContext.runAsNonRoot

`bool` · optional (explicit presence)

Refuse to start any container whose effective user is root.

### spec.podSecurityContext.fsGroup

`int64` · optional (explicit presence)

GID that owns mounted volumes and is added to every container's supplemental
groups — the standard fix for "permission denied" on persistent volumes written
by non-root apps.

### spec.podSecurityContext.fsGroupChangePolicy

`string`

When volume ownership is re-chowned to fs_group: "Always" (default) or
"OnRootMismatch" (skip the recursive chown when the root already matches —
dramatically faster pod starts on large volumes).

- rule: fsGroupChangePolicy must be either "Always" or "OnRootMismatch"

### spec.podSecurityContext.supplementalGroups

`[]int64`

Additional group IDs applied to all container processes.

### spec.podSecurityContext.sysctls

`[]WorkloadSysctl`

Kernel parameters set for the pod. Only safe sysctls (or those the cluster
administrator has allow-listed on the kubelet) are admitted.

### spec.podSecurityContext.sysctls[].name

`string` · required

Sysctl name, e.g. "net.core.somaxconn".

- rule: {"required":true}

### spec.podSecurityContext.sysctls[].value

`string` · required

Sysctl value, e.g. "1024".

- rule: {"required":true}

### spec.podSecurityContext.seccompProfile

`WorkloadSeccompProfile`

Pod-wide seccomp profile; containers may override with their own.

- rule: localhost_profile is required when type is "Localhost" and must be empty otherwise

### spec.podSecurityContext.seccompProfile.type

`string` · required

Profile type: "RuntimeDefault" (the container runtime's default filter — the
recommended baseline), "Unconfined" (no filtering), or "Localhost" (a profile
file installed on the node, named via localhost_profile).

- rule: Seccomp profile type must be one of "RuntimeDefault", "Unconfined", or "Localhost"
- rule: {"required":true}

### spec.podSecurityContext.seccompProfile.localhostProfile

`string`

Path of the profile file relative to the node's seccomp profile root. Required
when (and only meaningful when) type is "Localhost".

### spec.containerSecurityContext

`WorkloadContainerSecurityContext`

Container-level security context for the flagd container.

### spec.containerSecurityContext.privileged

`bool`

Runs the container with full host access — equivalent to root on the node.
Required by some node-level agents (device managers, network plugins). Never
combine with untrusted images.

### spec.containerSecurityContext.runAsUser

`int64` · optional (explicit presence)

UID the container process runs as. Overrides the image's USER directive.

### spec.containerSecurityContext.runAsGroup

`int64` · optional (explicit presence)

Primary GID the container process runs as.

### spec.containerSecurityContext.runAsNonRoot

`bool` · optional (explicit presence)

Refuses to start the container if its effective user is root. The standard
baseline hardening — it catches images that silently default to UID 0.

### spec.containerSecurityContext.readOnlyRootFilesystem

`bool` · optional (explicit presence)

Mounts the container's root filesystem read-only. Pair with EmptyDir mounts for
paths the app must write (e.g. /tmp).

### spec.containerSecurityContext.allowPrivilegeEscalation

`bool` · optional (explicit presence)

Whether the process can gain more privileges than its parent (setuid binaries,
file capabilities). The restricted Pod Security Standard requires this to be
false. Always true when `privileged` is set, so leave it unset in that case.

### spec.containerSecurityContext.capabilities

`WorkloadCapabilities`

Linux capabilities to add or drop. The restricted profile drops ALL and adds
back only NET_BIND_SERVICE when needed. Capability names are uppercase without
the CAP_ prefix (e.g. "NET_ADMIN", "SYS_TIME").

### spec.containerSecurityContext.capabilities.add

`[]string`

Capabilities to add (e.g. "NET_BIND_SERVICE").

### spec.containerSecurityContext.capabilities.drop

`[]string`

Capabilities to drop. Use ["ALL"] as the hardened baseline.

### spec.containerSecurityContext.seccompProfile

`WorkloadSeccompProfile`

Seccomp syscall filter for the container. "RuntimeDefault" is the hardened
baseline; "Localhost" selects a node-local profile file via `localhost_profile`.

- rule: localhost_profile is required when type is "Localhost" and must be empty otherwise

### spec.containerSecurityContext.seccompProfile.type

`string` · required

Profile type: "RuntimeDefault" (the container runtime's default filter — the
recommended baseline), "Unconfined" (no filtering), or "Localhost" (a profile
file installed on the node, named via localhost_profile).

- rule: Seccomp profile type must be one of "RuntimeDefault", "Unconfined", or "Localhost"
- rule: {"required":true}

### spec.containerSecurityContext.seccompProfile.localhostProfile

`string`

Path of the profile file relative to the node's seccomp profile root. Required
when (and only meaningful when) type is "Localhost".

### spec.extraEnv

`map<string, string>`

Extra plain environment variables - e.g. AWS_REGION for an S3 source,
or AZURE_STORAGE_ACCOUNT for an Azure Blob source (an Azure Blob source
needs the account set here or flagd fails to start). Variables named
FLAGD_* override flagd settings and are refused here (use the typed
fields).

- rule: FLAGD_* variables override flagd's own settings; use the typed fields instead.

### spec.extraEnvFromSecret

`map<string, KubernetesSecretKey>`

Extra environment variables read from existing Secrets in flagd's
namespace, keyed by variable name - e.g. AWS_ACCESS_KEY_ID and
AWS_SECRET_ACCESS_KEY, or AZURE_STORAGE_KEY.

- rule: FLAGD_* variables override flagd's own settings; use the typed fields instead.

### spec.extraEnvFromSecret.*.name

`string`

The name of the Kubernetes Secret.

### spec.extraEnvFromSecret.*.key

`string`

The key within the Kubernetes Secret.

### spec.metrics

`KubernetesFlagdMetrics`

Prometheus scraping of the management port.

### spec.metrics.serviceMonitorEnabled

`bool`

Create a ServiceMonitor for the management port (requires the
Prometheus Operator CRDs on the cluster; the apply FAILS without
them).

### spec.metrics.serviceMonitorLabels

`map<string, string>`

Labels on the ServiceMonitor, to match a Prometheus instance's
serviceMonitorSelector.

## Validation Rules

- `spec.extra_env.unique_names`: A variable may be declared in extra_env or extra_env_from_secret, not both.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesFlagd, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace flagd runs in. |
| `status.outputs.service` | `string` | The flagd Service name. |
| `status.outputs.evaluation_endpoint` | `string` | In-cluster gRPC evaluation endpoint host:port (e.g. "flagd.feature-flags.svc.cluster.local:8013"). |
| `status.outputs.sync_endpoint` | `string` | In-cluster sync endpoint host:port for in-process providers (e.g. "flagd.feature-flags.svc.cluster.local:8015"). |
| `status.outputs.ofrep_endpoint` | `string` | In-cluster OFREP base URL (e.g. "http://flagd.feature-flags.svc.cluster.local:8016"). |
| `status.outputs.management_endpoint` | `string` | In-cluster management endpoint serving /healthz, /readyz and /metrics (e.g. "http://flagd.feature-flags.svc.cluster.local:8014"). |
| `status.outputs.port_forward_command` | `string` | Copy-paste command for reaching the OFREP endpoint from a workstation. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.sources[].configMap.configMapName` | KubernetesFlagdFlagFile | `status.outputs.config_map_name` |
| `spec.sources[].configMap.key` | KubernetesFlagdFlagFile | `status.outputs.key` |

## See Also

- [Overview](../README.md)
