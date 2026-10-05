# KubernetesGoFeatureFlag

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesGoFeatureFlagSpec** installs the GO Feature Flag relay proxy -
an OpenFeature-native feature flag server - from the official `relay-proxy`
chart (https://charts.gofeatureflag.org, chart 1.56.x = GO Feature Flag
v1.56). The relay reads flag files from one or more retrievers, evaluates
flags for any OpenFeature SDK (REST, OFREP, and the flag-configuration
endpoint in-process providers sync from), notifies on every flag change,
and exports evaluation events.

This component deploys the ENGINE. The flags themselves are DATA with
their own lifecycle: declare them as a KubernetesGoFeatureFlagFlagFile
(typed, validated flags rendered into a ConfigMap) and point a
`config_map` retriever at it, or let the relay read a flag file from Git,
HTTP, object storage, or a database. A flag flip then edits the flag file
only - the relay is never re-applied, and the relay picks the change up on
its next poll with no restart.

TWO MODES (the relay accepts exactly one): `flag_source` serves one flag
file set to every caller; `flag_sets` serves several isolated flag sets,
each selected by the API key the caller presents.

SECRETS: every token, password, webhook URL and API key is written by the
module into its own Secret and reaches the relay as an environment
variable. None is ever rendered into the chart's configuration ConfigMap.

NOT MODELED, BY DESIGN (verified at chart 1.56.0 / relay v1.56.0):
- exposure: compose a Gateway API route against the `service` output
  instead of the chart's ingress block;
- server modes `lambda` and `unixsocket` and the listen host: a
  Kubernetes Service reaches only an all-interfaces HTTP listener, so the
  module pins `server.mode: http` on 0.0.0.0;
- `persistentFlagConfigurationFile`, the `file` retriever, the `file`
  exporter and the HTTP retriever's client-certificate paths: each needs a
  file mounted into the container, and the chart mounts no volume besides
  its own configuration;
- the flag-set `environment` key, which the relay parses but never reads;
- the singular `retriever`/`exporter` keys, `openTelemetryOtlpEndpoint`
  (an alias of `otel.exporter.otlp.endpoint`) and the deprecated keys
  (`listen`, top-level `monitoringPort`, `enableSwagger`, `host`,
  `apiKeys`, `startAsAwsLambda`, `notifier`, `githubToken`,
  `slackWebhookUrl`): the lists and current keys below cover them.

## Example

```yaml
# Full-surface development manifest - exercises every module-rendered arm so
# the offline plan/preview proofs cover what the kind-cluster lanes exclude:
# every retriever, notifier and exporter kind (their secrets riding the
# module-owned env Secret), authorized keys, telemetry, Swagger, the runtime
# switches, HPA, PDB, scheduling, security contexts, extra environment, the
# ServiceMonitor, and the helm_values escape hatch with its fullnameOverride
# re-pin. Every secret below is a development literal; a real manifest holds
# managed-secret references ($secret/<slug>) in these fields.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGoFeatureFlag
metadata:
  name: goff-dev
spec:
  namespace:
    value: goff-dev
  createNamespace: true
  chartVersion: 1.56.0
  image:
    repository: gofeatureflag/go-feature-flag
    tag: v1.56.0
    pullPolicy: IfNotPresent
    pullSecretNames:
      - registry-mirror
  replicas: 2
  resources:
    limits:
      cpu: 1000m
      memory: 512Mi
    requests:
      cpu: 100m
      memory: 128Mi
  hpa:
    enabled: true
    minReplicas: 2
    maxReplicas: 6
    targetCpuUtilizationPercent: 70
  pdb:
    enabled: true
    minAvailable: "1"
  server:
    port: 1031
    monitoringPort: 1032
    serviceType: ClusterIP
  log:
    level: info
    format: json
  authorizedKeys:
    admin:
      - dev-admin-key
    evaluation:
      - dev-evaluation-key
  flagSource:
    retrievers:
      - configMap:
          configMapName:
            value: goff-dev-flags
          key:
            value: flags.goff.yaml
      - http:
          url: https://flags.example.com/flags.yaml
          method: GET
          headers:
            Accept: application/yaml
          sensitiveHeaders:
            Authorization: Bearer dev-http-token
          timeoutMs: 5000
      - github:
          repositorySlug: acme/flags
          path: flags/release.goff.yaml
          branch: main
          token: dev-github-token
      - gitlab:
          repositorySlug: acme/platform/flags
          path: flags.goff.yaml
          baseUrl: https://gitlab.example.com
          token: dev-gitlab-token
      - bitbucket:
          repositorySlug: acme/flags
          path: flags.goff.yaml
          token: dev-bitbucket-token
      - s3:
          bucket: acme-flags
          item: flags.goff.yaml
      - googleStorage:
          bucket: acme-flags
          object: flags.goff.yaml
      - azureBlobStorage:
          accountName: acmeflags
          accountKey: dev-azure-key
          container: flags
          object: flags.goff.yaml
      - mongodb:
          uri: mongodb://flags:dev-mongo-password@mongo:27017
          database: flags
          collection: flags
      - redis:
          options:
            addr: redis:6379
            password: dev-redis-password
            db: 2
            tlsEnabled: true
            poolSize: 20
          prefix: "goff:"
      - postgresql:
          uri: postgres://flags:dev-pg-password@postgres:5432/flags?sslmode=require
          table: flags
          columns:
            flag_name: name
    notifiers:
      - slack:
          webhookUrl: https://hooks.slack.com/services/dev/slack
      - microsoftTeams:
          webhookUrl: https://acme.webhook.office.com/dev/teams
      - discord:
          webhookUrl: https://discord.com/api/webhooks/dev/discord
      - webhook:
          endpointUrl: https://ops.example.com/flag-changes
          secret: dev-webhook-signing-secret
          meta:
            environment: dev
          headers:
            X-Source: goff-dev
          sensitiveHeaders:
            X-Api-Key: dev-webhook-api-key
    exporters:
      - log:
          logFormat: '[{{ .FormattedDate}}] {{ .Key}}={{ .Value}}'
        eventType: feature
      - webhook:
          endpointUrl: https://events.example.com/goff
          secret: dev-exporter-signing-secret
        flushIntervalMs: 10000
        maxEventInMemory: 5000
      - s3:
          bucket: acme-flag-events
          path: goff/
          file:
            format: Parquet
            parquetCompressionCodec: ZSTD
      - googleStorage:
          bucket: acme-flag-events
          file:
            format: CSV
            csvTemplate: "{{ .Key}};{{ .Value}}\n"
      - azureBlobStorage:
          accountName: acmeevents
          accountKey: dev-azure-events-key
          container: events
      - sqs:
          queueUrl: https://sqs.us-east-1.amazonaws.com/123456789012/flag-events
      - kinesis:
          streamName: flag-events
      - pubsub:
          projectId: acme-dev
          topic: flag-events
      - bigquery:
          projectId: acme-dev
          datasetId: flags
          tableName: events
          googleCredentials: '{"type":"service_account"}'
          autoMigrate: true
      - kafka:
          topic: flag-events
          addresses:
            - kafka:9092
          config:
            Net:
              SASL:
                Enable: true
                Mechanism: SCRAM-SHA-512
                User: goff
          saslPassword: dev-kafka-password
      - opentelemetry:
          tracerName: goff-dev
    fileFormat: yaml
    pollingIntervalMs: 15000
    startWithRetrieverError: true
    enablePollingJitter: true
    disableNotifierOnInit: true
    evaluationContextEnrichment:
      region: us-east-1
      tier: 2
  ofrepEventStream:
    baseUrl: https://flags.example.com
    inactivityDelaySec: 120
  telemetry:
    otlpEndpoint: http://otel-collector:4318
    otlpProtocol: http/protobuf
    serviceName: goff-dev
    tracesSampler: jaeger_remote
    resourceAttributes:
      deployment.environment: dev
    jaegerSampler:
      managerHostPort: http://jaeger:5778
      refreshInterval: 1m
      maxOperations: 100
  swagger:
    enabled: true
    host: flags.example.com
  runtime:
    hideBanner: true
    enableBulkMetricFlagNames: true
    exporterCleanQueueInterval: 1m
    envVariablePrefix: GOFFRELAY_
  metrics:
    serviceMonitorEnabled: true
    serviceMonitorLabels:
      release: kube-prometheus-stack
  scheduling:
    nodeSelector:
      kubernetes.io/os: linux
    tolerations:
      - key: dedicated
        operator: Equal
        value: platform
        effect: NoSchedule
    podAntiAffinity:
      preferred:
        - weight: 100
          term:
            matchLabels:
              app.kubernetes.io/name: relay-proxy
            topologyKey: kubernetes.io/hostname
  serviceAccount:
    annotations:
      iam.gke.io/gcp-service-account: goff-dev@acme-dev.iam.gserviceaccount.com
  podAnnotations:
    sidecar.istio.io/inject: "false"
  podLabels:
    team: platform
  commonLabels:
    app.kubernetes.io/part-of: feature-flags
  podSecurityContext:
    runAsNonRoot: true
    runAsUser: 65532
  containerSecurityContext:
    readOnlyRootFilesystem: true
    allowPrivilegeEscalation: false
    capabilities:
      drop:
        - ALL
  extraEnv:
    AWS_REGION: us-east-1
  extraEnvFromSecret:
    AWS_ACCESS_KEY_ID:
      name: aws-credentials
      key: access-key-id
  helmValues: |
    extraManifests: []
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `1.56.0` |  |
| `spec.image` | `KubernetesGoFeatureFlagImage` |  |  |  |
| `spec.image.repository` | `string` |  | `gofeatureflag/go-feature-flag` |  |
| `spec.image.tag` | `string` |  |  |  |
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
| `spec.hpa` | `KubernetesGoFeatureFlagHpa` |  |  |  |
| `spec.hpa.enabled` | `bool` |  |  |  |
| `spec.hpa.minReplicas` | `int32` |  | `1` |  |
| `spec.hpa.maxReplicas` | `int32` |  | `10` |  |
| `spec.hpa.targetCpuUtilizationPercent` | `int32` |  | `80` |  |
| `spec.hpa.targetMemoryUtilizationPercent` | `int32` |  |  |  |
| `spec.pdb` | `KubernetesGoFeatureFlagPdb` |  |  |  |
| `spec.pdb.enabled` | `bool` |  |  |  |
| `spec.pdb.minAvailable` | `string` |  |  |  |
| `spec.pdb.maxUnavailable` | `string` |  |  |  |
| `spec.server` | `KubernetesGoFeatureFlagServer` |  |  |  |
| `spec.server.port` | `int32` |  | `1031` |  |
| `spec.server.monitoringPort` | `int32` |  | `1032` |  |
| `spec.server.serviceType` | `string` |  | `ClusterIP` |  |
| `spec.log` | `KubernetesGoFeatureFlagLog` |  |  |  |
| `spec.log.level` | `string` |  | `info` |  |
| `spec.log.format` | `string` |  | `json` |  |
| `spec.authorizedKeys` | `KubernetesGoFeatureFlagAuthorizedKeys` |  |  |  |
| `spec.authorizedKeys.admin` | `[]string` (sensitive) |  |  |  |
| `spec.authorizedKeys.evaluation` | `[]string` (sensitive) |  |  |  |
| `spec.flagSource` | `KubernetesGoFeatureFlagFlagSource` |  |  |  |
| `spec.flagSource.retrievers` | `[]KubernetesGoFeatureFlagRetriever` | yes |  |  |
| `spec.flagSource.retrievers[].configMap` | `KubernetesGoFeatureFlagConfigMapRetriever` |  |  |  |
| `spec.flagSource.retrievers[].configMap.configMapName` | `string \| valueFrom` | yes |  | KubernetesGoFeatureFlagFlagFile (`status.outputs.config_map_name`) |
| `spec.flagSource.retrievers[].configMap.key` | `string \| valueFrom` | yes |  | KubernetesGoFeatureFlagFlagFile (`status.outputs.key`) |
| `spec.flagSource.retrievers[].configMap.namespace` | `string \| valueFrom` |  |  | KubernetesNamespace (`spec.name`) |
| `spec.flagSource.retrievers[].http` | `KubernetesGoFeatureFlagHttpRetriever` |  |  |  |
| `spec.flagSource.retrievers[].http.url` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].http.method` | `string` |  |  |  |
| `spec.flagSource.retrievers[].http.body` | `string` |  |  |  |
| `spec.flagSource.retrievers[].http.headers` | `map<string, string>` |  |  |  |
| `spec.flagSource.retrievers[].http.sensitiveHeaders` | `map<string, string>` (sensitive) |  |  |  |
| `spec.flagSource.retrievers[].http.timeoutMs` | `int32` |  |  |  |
| `spec.flagSource.retrievers[].github` | `KubernetesGoFeatureFlagGitRetriever` |  |  |  |
| `spec.flagSource.retrievers[].github.repositorySlug` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].github.path` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].github.branch` | `string` |  |  |  |
| `spec.flagSource.retrievers[].github.token` | `string` (sensitive) |  |  |  |
| `spec.flagSource.retrievers[].github.baseUrl` | `string` |  |  |  |
| `spec.flagSource.retrievers[].github.timeoutMs` | `int32` |  |  |  |
| `spec.flagSource.retrievers[].gitlab` | `KubernetesGoFeatureFlagGitRetriever` |  |  |  |
| `spec.flagSource.retrievers[].gitlab.repositorySlug` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].gitlab.path` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].gitlab.branch` | `string` |  |  |  |
| `spec.flagSource.retrievers[].gitlab.token` | `string` (sensitive) |  |  |  |
| `spec.flagSource.retrievers[].gitlab.baseUrl` | `string` |  |  |  |
| `spec.flagSource.retrievers[].gitlab.timeoutMs` | `int32` |  |  |  |
| `spec.flagSource.retrievers[].bitbucket` | `KubernetesGoFeatureFlagGitRetriever` |  |  |  |
| `spec.flagSource.retrievers[].bitbucket.repositorySlug` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].bitbucket.path` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].bitbucket.branch` | `string` |  |  |  |
| `spec.flagSource.retrievers[].bitbucket.token` | `string` (sensitive) |  |  |  |
| `spec.flagSource.retrievers[].bitbucket.baseUrl` | `string` |  |  |  |
| `spec.flagSource.retrievers[].bitbucket.timeoutMs` | `int32` |  |  |  |
| `spec.flagSource.retrievers[].s3` | `KubernetesGoFeatureFlagS3Retriever` |  |  |  |
| `spec.flagSource.retrievers[].s3.bucket` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].s3.item` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].googleStorage` | `KubernetesGoFeatureFlagGcsRetriever` |  |  |  |
| `spec.flagSource.retrievers[].googleStorage.bucket` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].googleStorage.object` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].azureBlobStorage` | `KubernetesGoFeatureFlagAzureBlobRetriever` |  |  |  |
| `spec.flagSource.retrievers[].azureBlobStorage.accountName` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].azureBlobStorage.accountKey` | `string` (sensitive) |  |  |  |
| `spec.flagSource.retrievers[].azureBlobStorage.container` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].azureBlobStorage.object` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].mongodb` | `KubernetesGoFeatureFlagMongoDbRetriever` |  |  |  |
| `spec.flagSource.retrievers[].mongodb.uri` | `string` (sensitive) | yes |  |  |
| `spec.flagSource.retrievers[].mongodb.database` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].mongodb.collection` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].redis` | `KubernetesGoFeatureFlagRedisRetriever` |  |  |  |
| `spec.flagSource.retrievers[].redis.options` | `KubernetesGoFeatureFlagRedisOptions` | yes |  |  |
| `spec.flagSource.retrievers[].redis.options.addr` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].redis.options.network` | `string` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.username` | `string` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.password` | `string` (sensitive) |  |  |  |
| `spec.flagSource.retrievers[].redis.options.db` | `int32` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.tlsEnabled` | `bool` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.protocol` | `int32` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.clientName` | `string` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.identitySuffix` | `string` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.disableIdentity` | `bool` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.maxRetries` | `int32` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.minRetryBackoffMs` | `int64` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.maxRetryBackoffMs` | `int64` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.dialTimeoutMs` | `int64` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.readTimeoutMs` | `int64` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.writeTimeoutMs` | `int64` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.contextTimeoutEnabled` | `bool` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.poolFifo` | `bool` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.poolSize` | `int32` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.poolTimeoutMs` | `int64` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.minIdleConns` | `int32` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.maxIdleConns` | `int32` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.connMaxIdleTimeMs` | `int64` |  |  |  |
| `spec.flagSource.retrievers[].redis.options.connMaxLifetimeMs` | `int64` |  |  |  |
| `spec.flagSource.retrievers[].redis.prefix` | `string` |  |  |  |
| `spec.flagSource.retrievers[].postgresql` | `KubernetesGoFeatureFlagPostgresRetriever` |  |  |  |
| `spec.flagSource.retrievers[].postgresql.uri` | `string` (sensitive) | yes |  |  |
| `spec.flagSource.retrievers[].postgresql.table` | `string` | yes |  |  |
| `spec.flagSource.retrievers[].postgresql.columns` | `map<string, string>` |  |  |  |
| `spec.flagSource.notifiers` | `[]KubernetesGoFeatureFlagNotifier` |  |  |  |
| `spec.flagSource.notifiers[].slack` | `KubernetesGoFeatureFlagWebhookUrlNotifier` |  |  |  |
| `spec.flagSource.notifiers[].slack.webhookUrl` | `string` (sensitive) | yes |  |  |
| `spec.flagSource.notifiers[].microsoftTeams` | `KubernetesGoFeatureFlagWebhookUrlNotifier` |  |  |  |
| `spec.flagSource.notifiers[].microsoftTeams.webhookUrl` | `string` (sensitive) | yes |  |  |
| `spec.flagSource.notifiers[].discord` | `KubernetesGoFeatureFlagWebhookUrlNotifier` |  |  |  |
| `spec.flagSource.notifiers[].discord.webhookUrl` | `string` (sensitive) | yes |  |  |
| `spec.flagSource.notifiers[].webhook` | `KubernetesGoFeatureFlagWebhookNotifier` |  |  |  |
| `spec.flagSource.notifiers[].webhook.endpointUrl` | `string` | yes |  |  |
| `spec.flagSource.notifiers[].webhook.secret` | `string` (sensitive) |  |  |  |
| `spec.flagSource.notifiers[].webhook.meta` | `map<string, string>` |  |  |  |
| `spec.flagSource.notifiers[].webhook.headers` | `map<string, string>` |  |  |  |
| `spec.flagSource.notifiers[].webhook.sensitiveHeaders` | `map<string, string>` (sensitive) |  |  |  |
| `spec.flagSource.exporters` | `[]KubernetesGoFeatureFlagExporter` |  |  |  |
| `spec.flagSource.exporters[].webhook` | `KubernetesGoFeatureFlagWebhookExporter` |  |  |  |
| `spec.flagSource.exporters[].webhook.endpointUrl` | `string` | yes |  |  |
| `spec.flagSource.exporters[].webhook.secret` | `string` (sensitive) |  |  |  |
| `spec.flagSource.exporters[].webhook.meta` | `map<string, string>` |  |  |  |
| `spec.flagSource.exporters[].webhook.headers` | `map<string, string>` |  |  |  |
| `spec.flagSource.exporters[].webhook.sensitiveHeaders` | `map<string, string>` (sensitive) |  |  |  |
| `spec.flagSource.exporters[].log` | `KubernetesGoFeatureFlagLogExporter` |  |  |  |
| `spec.flagSource.exporters[].log.logFormat` | `string` |  |  |  |
| `spec.flagSource.exporters[].s3` | `KubernetesGoFeatureFlagS3Exporter` |  |  |  |
| `spec.flagSource.exporters[].s3.bucket` | `string` | yes |  |  |
| `spec.flagSource.exporters[].s3.path` | `string` |  |  |  |
| `spec.flagSource.exporters[].s3.file` | `KubernetesGoFeatureFlagEventFileFormat` |  |  |  |
| `spec.flagSource.exporters[].s3.file.format` | `string` |  |  |  |
| `spec.flagSource.exporters[].s3.file.filename` | `string` |  |  |  |
| `spec.flagSource.exporters[].s3.file.csvTemplate` | `string` |  |  |  |
| `spec.flagSource.exporters[].s3.file.parquetCompressionCodec` | `string` |  |  |  |
| `spec.flagSource.exporters[].googleStorage` | `KubernetesGoFeatureFlagGcsExporter` |  |  |  |
| `spec.flagSource.exporters[].googleStorage.bucket` | `string` | yes |  |  |
| `spec.flagSource.exporters[].googleStorage.path` | `string` |  |  |  |
| `spec.flagSource.exporters[].googleStorage.file` | `KubernetesGoFeatureFlagEventFileFormat` |  |  |  |
| `spec.flagSource.exporters[].googleStorage.file.format` | `string` |  |  |  |
| `spec.flagSource.exporters[].googleStorage.file.filename` | `string` |  |  |  |
| `spec.flagSource.exporters[].googleStorage.file.csvTemplate` | `string` |  |  |  |
| `spec.flagSource.exporters[].googleStorage.file.parquetCompressionCodec` | `string` |  |  |  |
| `spec.flagSource.exporters[].azureBlobStorage` | `KubernetesGoFeatureFlagAzureBlobExporter` |  |  |  |
| `spec.flagSource.exporters[].azureBlobStorage.accountName` | `string` | yes |  |  |
| `spec.flagSource.exporters[].azureBlobStorage.accountKey` | `string` (sensitive) |  |  |  |
| `spec.flagSource.exporters[].azureBlobStorage.container` | `string` | yes |  |  |
| `spec.flagSource.exporters[].azureBlobStorage.path` | `string` |  |  |  |
| `spec.flagSource.exporters[].azureBlobStorage.file` | `KubernetesGoFeatureFlagEventFileFormat` |  |  |  |
| `spec.flagSource.exporters[].azureBlobStorage.file.format` | `string` |  |  |  |
| `spec.flagSource.exporters[].azureBlobStorage.file.filename` | `string` |  |  |  |
| `spec.flagSource.exporters[].azureBlobStorage.file.csvTemplate` | `string` |  |  |  |
| `spec.flagSource.exporters[].azureBlobStorage.file.parquetCompressionCodec` | `string` |  |  |  |
| `spec.flagSource.exporters[].sqs` | `KubernetesGoFeatureFlagSqsExporter` |  |  |  |
| `spec.flagSource.exporters[].sqs.queueUrl` | `string` | yes |  |  |
| `spec.flagSource.exporters[].kinesis` | `KubernetesGoFeatureFlagKinesisExporter` |  |  |  |
| `spec.flagSource.exporters[].kinesis.streamArn` | `string` |  |  |  |
| `spec.flagSource.exporters[].kinesis.streamName` | `string` |  |  |  |
| `spec.flagSource.exporters[].kinesis.format` | `string` |  |  |  |
| `spec.flagSource.exporters[].pubsub` | `KubernetesGoFeatureFlagPubSubExporter` |  |  |  |
| `spec.flagSource.exporters[].pubsub.projectId` | `string` | yes |  |  |
| `spec.flagSource.exporters[].pubsub.topic` | `string` | yes |  |  |
| `spec.flagSource.exporters[].bigquery` | `KubernetesGoFeatureFlagBigQueryExporter` |  |  |  |
| `spec.flagSource.exporters[].bigquery.projectId` | `string` | yes |  |  |
| `spec.flagSource.exporters[].bigquery.datasetId` | `string` | yes |  |  |
| `spec.flagSource.exporters[].bigquery.tableName` | `string` |  |  |  |
| `spec.flagSource.exporters[].bigquery.googleCredentials` | `string` (sensitive) |  |  |  |
| `spec.flagSource.exporters[].bigquery.autoMigrate` | `bool` |  |  |  |
| `spec.flagSource.exporters[].kafka` | `KubernetesGoFeatureFlagKafkaExporter` |  |  |  |
| `spec.flagSource.exporters[].kafka.topic` | `string` | yes |  |  |
| `spec.flagSource.exporters[].kafka.addresses` | `[]string` | yes |  |  |
| `spec.flagSource.exporters[].kafka.config` | `object` |  |  |  |
| `spec.flagSource.exporters[].kafka.saslPassword` | `string` (sensitive) |  |  |  |
| `spec.flagSource.exporters[].opentelemetry` | `KubernetesGoFeatureFlagOpenTelemetryExporter` |  |  |  |
| `spec.flagSource.exporters[].opentelemetry.tracerName` | `string` |  |  |  |
| `spec.flagSource.exporters[].flushIntervalMs` | `int64` |  |  |  |
| `spec.flagSource.exporters[].maxEventInMemory` | `int64` |  |  |  |
| `spec.flagSource.exporters[].eventType` | `string` |  |  |  |
| `spec.flagSource.fileFormat` | `string` |  |  |  |
| `spec.flagSource.pollingIntervalMs` | `int32` |  |  |  |
| `spec.flagSource.startWithRetrieverError` | `bool` |  |  |  |
| `spec.flagSource.enablePollingJitter` | `bool` |  |  |  |
| `spec.flagSource.disableNotifierOnInit` | `bool` |  |  |  |
| `spec.flagSource.evaluationContextEnrichment` | `object` |  |  |  |
| `spec.flagSets` | `KubernetesGoFeatureFlagFlagSets` |  |  |  |
| `spec.flagSets.items` | `[]KubernetesGoFeatureFlagFlagSet` | yes |  |  |
| `spec.flagSets.items[].name` | `string` |  |  |  |
| `spec.flagSets.items[].apiKeys` | `[]string` (sensitive) | yes |  |  |
| `spec.flagSets.items[].source` | `KubernetesGoFeatureFlagFlagSource` | yes |  |  |
| `spec.flagSets.items[].source.retrievers` | `[]KubernetesGoFeatureFlagRetriever` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].configMap` | `KubernetesGoFeatureFlagConfigMapRetriever` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].configMap.configMapName` | `string \| valueFrom` | yes |  | KubernetesGoFeatureFlagFlagFile (`status.outputs.config_map_name`) |
| `spec.flagSets.items[].source.retrievers[].configMap.key` | `string \| valueFrom` | yes |  | KubernetesGoFeatureFlagFlagFile (`status.outputs.key`) |
| `spec.flagSets.items[].source.retrievers[].configMap.namespace` | `string \| valueFrom` |  |  | KubernetesNamespace (`spec.name`) |
| `spec.flagSets.items[].source.retrievers[].http` | `KubernetesGoFeatureFlagHttpRetriever` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].http.url` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].http.method` | `string` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].http.body` | `string` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].http.headers` | `map<string, string>` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].http.sensitiveHeaders` | `map<string, string>` (sensitive) |  |  |  |
| `spec.flagSets.items[].source.retrievers[].http.timeoutMs` | `int32` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].github` | `KubernetesGoFeatureFlagGitRetriever` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].github.repositorySlug` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].github.path` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].github.branch` | `string` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].github.token` | `string` (sensitive) |  |  |  |
| `spec.flagSets.items[].source.retrievers[].github.baseUrl` | `string` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].github.timeoutMs` | `int32` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].gitlab` | `KubernetesGoFeatureFlagGitRetriever` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].gitlab.repositorySlug` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].gitlab.path` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].gitlab.branch` | `string` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].gitlab.token` | `string` (sensitive) |  |  |  |
| `spec.flagSets.items[].source.retrievers[].gitlab.baseUrl` | `string` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].gitlab.timeoutMs` | `int32` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].bitbucket` | `KubernetesGoFeatureFlagGitRetriever` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].bitbucket.repositorySlug` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].bitbucket.path` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].bitbucket.branch` | `string` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].bitbucket.token` | `string` (sensitive) |  |  |  |
| `spec.flagSets.items[].source.retrievers[].bitbucket.baseUrl` | `string` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].bitbucket.timeoutMs` | `int32` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].s3` | `KubernetesGoFeatureFlagS3Retriever` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].s3.bucket` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].s3.item` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].googleStorage` | `KubernetesGoFeatureFlagGcsRetriever` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].googleStorage.bucket` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].googleStorage.object` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].azureBlobStorage` | `KubernetesGoFeatureFlagAzureBlobRetriever` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].azureBlobStorage.accountName` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].azureBlobStorage.accountKey` | `string` (sensitive) |  |  |  |
| `spec.flagSets.items[].source.retrievers[].azureBlobStorage.container` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].azureBlobStorage.object` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].mongodb` | `KubernetesGoFeatureFlagMongoDbRetriever` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].mongodb.uri` | `string` (sensitive) | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].mongodb.database` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].mongodb.collection` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].redis` | `KubernetesGoFeatureFlagRedisRetriever` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options` | `KubernetesGoFeatureFlagRedisOptions` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.addr` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.network` | `string` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.username` | `string` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.password` | `string` (sensitive) |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.db` | `int32` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.tlsEnabled` | `bool` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.protocol` | `int32` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.clientName` | `string` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.identitySuffix` | `string` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.disableIdentity` | `bool` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.maxRetries` | `int32` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.minRetryBackoffMs` | `int64` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.maxRetryBackoffMs` | `int64` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.dialTimeoutMs` | `int64` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.readTimeoutMs` | `int64` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.writeTimeoutMs` | `int64` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.contextTimeoutEnabled` | `bool` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.poolFifo` | `bool` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.poolSize` | `int32` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.poolTimeoutMs` | `int64` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.minIdleConns` | `int32` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.maxIdleConns` | `int32` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.connMaxIdleTimeMs` | `int64` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.options.connMaxLifetimeMs` | `int64` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].redis.prefix` | `string` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].postgresql` | `KubernetesGoFeatureFlagPostgresRetriever` |  |  |  |
| `spec.flagSets.items[].source.retrievers[].postgresql.uri` | `string` (sensitive) | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].postgresql.table` | `string` | yes |  |  |
| `spec.flagSets.items[].source.retrievers[].postgresql.columns` | `map<string, string>` |  |  |  |
| `spec.flagSets.items[].source.notifiers` | `[]KubernetesGoFeatureFlagNotifier` |  |  |  |
| `spec.flagSets.items[].source.notifiers[].slack` | `KubernetesGoFeatureFlagWebhookUrlNotifier` |  |  |  |
| `spec.flagSets.items[].source.notifiers[].slack.webhookUrl` | `string` (sensitive) | yes |  |  |
| `spec.flagSets.items[].source.notifiers[].microsoftTeams` | `KubernetesGoFeatureFlagWebhookUrlNotifier` |  |  |  |
| `spec.flagSets.items[].source.notifiers[].microsoftTeams.webhookUrl` | `string` (sensitive) | yes |  |  |
| `spec.flagSets.items[].source.notifiers[].discord` | `KubernetesGoFeatureFlagWebhookUrlNotifier` |  |  |  |
| `spec.flagSets.items[].source.notifiers[].discord.webhookUrl` | `string` (sensitive) | yes |  |  |
| `spec.flagSets.items[].source.notifiers[].webhook` | `KubernetesGoFeatureFlagWebhookNotifier` |  |  |  |
| `spec.flagSets.items[].source.notifiers[].webhook.endpointUrl` | `string` | yes |  |  |
| `spec.flagSets.items[].source.notifiers[].webhook.secret` | `string` (sensitive) |  |  |  |
| `spec.flagSets.items[].source.notifiers[].webhook.meta` | `map<string, string>` |  |  |  |
| `spec.flagSets.items[].source.notifiers[].webhook.headers` | `map<string, string>` |  |  |  |
| `spec.flagSets.items[].source.notifiers[].webhook.sensitiveHeaders` | `map<string, string>` (sensitive) |  |  |  |
| `spec.flagSets.items[].source.exporters` | `[]KubernetesGoFeatureFlagExporter` |  |  |  |
| `spec.flagSets.items[].source.exporters[].webhook` | `KubernetesGoFeatureFlagWebhookExporter` |  |  |  |
| `spec.flagSets.items[].source.exporters[].webhook.endpointUrl` | `string` | yes |  |  |
| `spec.flagSets.items[].source.exporters[].webhook.secret` | `string` (sensitive) |  |  |  |
| `spec.flagSets.items[].source.exporters[].webhook.meta` | `map<string, string>` |  |  |  |
| `spec.flagSets.items[].source.exporters[].webhook.headers` | `map<string, string>` |  |  |  |
| `spec.flagSets.items[].source.exporters[].webhook.sensitiveHeaders` | `map<string, string>` (sensitive) |  |  |  |
| `spec.flagSets.items[].source.exporters[].log` | `KubernetesGoFeatureFlagLogExporter` |  |  |  |
| `spec.flagSets.items[].source.exporters[].log.logFormat` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].s3` | `KubernetesGoFeatureFlagS3Exporter` |  |  |  |
| `spec.flagSets.items[].source.exporters[].s3.bucket` | `string` | yes |  |  |
| `spec.flagSets.items[].source.exporters[].s3.path` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].s3.file` | `KubernetesGoFeatureFlagEventFileFormat` |  |  |  |
| `spec.flagSets.items[].source.exporters[].s3.file.format` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].s3.file.filename` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].s3.file.csvTemplate` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].s3.file.parquetCompressionCodec` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].googleStorage` | `KubernetesGoFeatureFlagGcsExporter` |  |  |  |
| `spec.flagSets.items[].source.exporters[].googleStorage.bucket` | `string` | yes |  |  |
| `spec.flagSets.items[].source.exporters[].googleStorage.path` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].googleStorage.file` | `KubernetesGoFeatureFlagEventFileFormat` |  |  |  |
| `spec.flagSets.items[].source.exporters[].googleStorage.file.format` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].googleStorage.file.filename` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].googleStorage.file.csvTemplate` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].googleStorage.file.parquetCompressionCodec` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].azureBlobStorage` | `KubernetesGoFeatureFlagAzureBlobExporter` |  |  |  |
| `spec.flagSets.items[].source.exporters[].azureBlobStorage.accountName` | `string` | yes |  |  |
| `spec.flagSets.items[].source.exporters[].azureBlobStorage.accountKey` | `string` (sensitive) |  |  |  |
| `spec.flagSets.items[].source.exporters[].azureBlobStorage.container` | `string` | yes |  |  |
| `spec.flagSets.items[].source.exporters[].azureBlobStorage.path` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].azureBlobStorage.file` | `KubernetesGoFeatureFlagEventFileFormat` |  |  |  |
| `spec.flagSets.items[].source.exporters[].azureBlobStorage.file.format` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].azureBlobStorage.file.filename` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].azureBlobStorage.file.csvTemplate` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].azureBlobStorage.file.parquetCompressionCodec` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].sqs` | `KubernetesGoFeatureFlagSqsExporter` |  |  |  |
| `spec.flagSets.items[].source.exporters[].sqs.queueUrl` | `string` | yes |  |  |
| `spec.flagSets.items[].source.exporters[].kinesis` | `KubernetesGoFeatureFlagKinesisExporter` |  |  |  |
| `spec.flagSets.items[].source.exporters[].kinesis.streamArn` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].kinesis.streamName` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].kinesis.format` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].pubsub` | `KubernetesGoFeatureFlagPubSubExporter` |  |  |  |
| `spec.flagSets.items[].source.exporters[].pubsub.projectId` | `string` | yes |  |  |
| `spec.flagSets.items[].source.exporters[].pubsub.topic` | `string` | yes |  |  |
| `spec.flagSets.items[].source.exporters[].bigquery` | `KubernetesGoFeatureFlagBigQueryExporter` |  |  |  |
| `spec.flagSets.items[].source.exporters[].bigquery.projectId` | `string` | yes |  |  |
| `spec.flagSets.items[].source.exporters[].bigquery.datasetId` | `string` | yes |  |  |
| `spec.flagSets.items[].source.exporters[].bigquery.tableName` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].bigquery.googleCredentials` | `string` (sensitive) |  |  |  |
| `spec.flagSets.items[].source.exporters[].bigquery.autoMigrate` | `bool` |  |  |  |
| `spec.flagSets.items[].source.exporters[].kafka` | `KubernetesGoFeatureFlagKafkaExporter` |  |  |  |
| `spec.flagSets.items[].source.exporters[].kafka.topic` | `string` | yes |  |  |
| `spec.flagSets.items[].source.exporters[].kafka.addresses` | `[]string` | yes |  |  |
| `spec.flagSets.items[].source.exporters[].kafka.config` | `object` |  |  |  |
| `spec.flagSets.items[].source.exporters[].kafka.saslPassword` | `string` (sensitive) |  |  |  |
| `spec.flagSets.items[].source.exporters[].opentelemetry` | `KubernetesGoFeatureFlagOpenTelemetryExporter` |  |  |  |
| `spec.flagSets.items[].source.exporters[].opentelemetry.tracerName` | `string` |  |  |  |
| `spec.flagSets.items[].source.exporters[].flushIntervalMs` | `int64` |  |  |  |
| `spec.flagSets.items[].source.exporters[].maxEventInMemory` | `int64` |  |  |  |
| `spec.flagSets.items[].source.exporters[].eventType` | `string` |  |  |  |
| `spec.flagSets.items[].source.fileFormat` | `string` |  |  |  |
| `spec.flagSets.items[].source.pollingIntervalMs` | `int32` |  |  |  |
| `spec.flagSets.items[].source.startWithRetrieverError` | `bool` |  |  |  |
| `spec.flagSets.items[].source.enablePollingJitter` | `bool` |  |  |  |
| `spec.flagSets.items[].source.disableNotifierOnInit` | `bool` |  |  |  |
| `spec.flagSets.items[].source.evaluationContextEnrichment` | `object` |  |  |  |
| `spec.ofrepEventStream` | `KubernetesGoFeatureFlagOfrepEventStream` |  |  |  |
| `spec.ofrepEventStream.baseUrl` | `string` |  |  |  |
| `spec.ofrepEventStream.inactivityDelaySec` | `int32` |  |  |  |
| `spec.telemetry` | `KubernetesGoFeatureFlagTelemetry` |  |  |  |
| `spec.telemetry.otlpEndpoint` | `string` |  |  |  |
| `spec.telemetry.otlpProtocol` | `string` |  |  |  |
| `spec.telemetry.sdkDisabled` | `bool` |  |  |  |
| `spec.telemetry.serviceName` | `string` |  |  |  |
| `spec.telemetry.tracesSampler` | `string` |  |  |  |
| `spec.telemetry.resourceAttributes` | `map<string, string>` |  |  |  |
| `spec.telemetry.jaegerSampler` | `KubernetesGoFeatureFlagJaegerSampler` |  |  |  |
| `spec.telemetry.jaegerSampler.managerHostPort` | `string` |  |  |  |
| `spec.telemetry.jaegerSampler.refreshInterval` | `string` |  |  |  |
| `spec.telemetry.jaegerSampler.maxOperations` | `int32` |  |  |  |
| `spec.telemetry.tracesSamplerArg` | `string` |  |  |  |
| `spec.swagger` | `KubernetesGoFeatureFlagSwagger` |  |  |  |
| `spec.swagger.enabled` | `bool` |  |  |  |
| `spec.swagger.host` | `string` |  |  |  |
| `spec.runtime` | `KubernetesGoFeatureFlagRuntime` |  |  |  |
| `spec.runtime.hideBanner` | `bool` |  |  |  |
| `spec.runtime.enablePprof` | `bool` |  |  |  |
| `spec.runtime.disableVersionHeader` | `bool` |  |  |  |
| `spec.runtime.enableBulkMetricFlagNames` | `bool` |  |  |  |
| `spec.runtime.disableFlagDetailsInStream` | `bool` |  |  |  |
| `spec.runtime.exporterCleanQueueInterval` | `string` |  |  |  |
| `spec.runtime.envVariablePrefix` | `string` |  | `GOFFRELAY_` |  |
| `spec.metrics` | `KubernetesGoFeatureFlagMetrics` |  |  |  |
| `spec.metrics.serviceMonitorEnabled` | `bool` |  |  |  |
| `spec.metrics.serviceMonitorLabels` | `map<string, string>` |  |  |  |
| `spec.scheduling` | `KubernetesGoFeatureFlagScheduling` |  |  |  |
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
| `spec.serviceAccount` | `KubernetesGoFeatureFlagServiceAccount` |  |  |  |
| `spec.serviceAccount.annotations` | `map<string, string>` |  |  |  |
| `spec.serviceAccount.existingName` | `string` |  |  |  |
| `spec.podAnnotations` | `map<string, string>` |  |  |  |
| `spec.podLabels` | `map<string, string>` |  |  |  |
| `spec.commonLabels` | `map<string, string>` |  |  |  |
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
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install into. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource. Secrets the relay reads
through `extra_env_from_secret` must live in this same namespace.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the resource.
When false, the namespace must already exist.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "1.56.0" = GO Feature Flag
v1.56.0; the chart version and the relay version move together).
Versions must exist in the SERVED index at
https://charts.gofeatureflag.org.

- default: `1.56.0`

### spec.image

`KubernetesGoFeatureFlagImage`

Relay container image. Empty fields keep the chart's image, at the tag
matching chart_version.

### spec.image.repository

`string` · optional (explicit presence)

Image repository. The same images are mirrored at
ghcr.io/go-feature-flag/go-feature-flag (tags after v1.55.1).

- default: `gofeatureflag/go-feature-flag`

### spec.image.tag

`string`

Image tag. Empty = the tag matching chart_version (the chart's
appVersion). Set it only to run a relay version other than the
chart's.

### spec.image.fips

`bool`

Run the FIPS 140 build (the module appends `-fips` to the tag).

### spec.image.pullPolicy

`string` · optional (explicit presence)

Image pull policy: Always, IfNotPresent or Never.

- default: `IfNotPresent`
- rule: Image pull policy must be one of: Always, IfNotPresent, Never.

### spec.image.pullSecretNames

`[]string`

Names of existing image pull Secrets in the relay's namespace (for a
private mirror).

### spec.replicas

`int32` · optional (explicit presence)

Relay replicas. The relay is stateless (every replica polls the same
retrievers), so replicas add availability and evaluation throughput.
Ignored when hpa is enabled (the autoscaler owns the count).

- default: `1`
- rule: {"int32":{"lte":50,"gte":1}}

### spec.resources

`ContainerResources`

CPU and memory for the relay container. The relay holds every flag in
memory and evaluates in-process; these defaults suit hundreds of flags
at moderate traffic - raise memory for very large flag files and CPU
for heavy evaluation volume.

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

`KubernetesGoFeatureFlagHpa`

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

Maximum replicas. The platform default is 10 (the chart's own default
is 100).

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
drive scaling (CPU alone does).

- rule: {"int32":{"lte":100,"gte":1}}

### spec.pdb

`KubernetesGoFeatureFlagPdb`

PodDisruptionBudget for voluntary disruptions (node drains, upgrades).

- rule: A PodDisruptionBudget takes either min_available or max_unavailable, not both.

### spec.pdb.enabled

`bool`

Create the PodDisruptionBudget. It only protects availability when
more than one replica runs.

### spec.pdb.minAvailable

`string`

Minimum pods that must stay available during a voluntary disruption:
an integer ("1") or a percentage ("50%"). Empty with max_unavailable
also empty = 1.

- rule: min_available must be a positive integer or a percentage up to 100%, such as 50% (the chart reads 0 as unset and renders 1).

### spec.pdb.maxUnavailable

`string`

Maximum pods that may be unavailable during a voluntary disruption: an
integer or a percentage. Mutually exclusive with min_available.

- rule: max_unavailable must be a non-negative integer or a percentage up to 100%, such as 25%.

### spec.server

`KubernetesGoFeatureFlagServer`

HTTP listener, monitoring listener and Service shape.

- rule: The evaluation port and the monitoring port must differ.

### spec.server.port

`int32` · optional (explicit presence)

Port of the evaluation API: REST evaluation, OFREP, and the flag
configuration endpoint in-process providers sync from. Also the
Service port.

- default: `1031`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.server.monitoringPort

`int32` · optional (explicit presence)

Separate port for /health, /info and /metrics, so probes and scraping
never share the evaluation listener. The Service exposes it as the
`monitoring` port.

- default: `1032`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.server.serviceType

`string` · optional (explicit presence)

Service type: ClusterIP, NodePort or LoadBalancer. Keep ClusterIP and
compose a Gateway API route for external access.

- default: `ClusterIP`
- rule: Service type must be one of: ClusterIP, NodePort, LoadBalancer.

### spec.log

`KubernetesGoFeatureFlagLog`

Relay log settings.

### spec.log.level

`string` · optional (explicit presence)

Log level: debug, info, warn, error, dpanic, panic or fatal.

- default: `info`
- rule: Log level must be one of: debug, info, warn, error, dpanic, panic, fatal.

### spec.log.format

`string` · optional (explicit presence)

Log format: json or logfmt.

- default: `json`
- rule: Log format must be either "json" or "logfmt".

### spec.authorizedKeys

`KubernetesGoFeatureFlagAuthorizedKeys`

API keys guarding the relay's endpoints. Unset = NO authentication -
anyone who can reach the Service can evaluate every flag and read the
full flag configuration. Fine on a lab cluster, never in production.
In `flag_sets` mode each flag set carries its own evaluation keys;
`admin` keys here still guard the admin endpoints.

### spec.authorizedKeys.admin

`[]string` · sensitive

Keys allowed to call the admin endpoints (forcing a flag refresh,
reading the full configuration). Admin keys can also evaluate.

### spec.authorizedKeys.evaluation

`[]string` · sensitive

Keys allowed to evaluate flags and read the flag configuration (what
OpenFeature SDKs and in-process providers present).

### spec.flagSource

`KubernetesGoFeatureFlagFlagSource`

One flag source served to every caller: its retrievers, change
notifiers and evaluation exporters.

### spec.flagSource.retrievers

`[]KubernetesGoFeatureFlagRetriever` · required

Retrievers the flag files are read from. When several are listed, the
relay merges their flags and a later retriever wins on a flag both
define.

- rule: {"repeated":{"minItems":"1"}}

### spec.flagSource.retrievers[].configMap

`KubernetesGoFeatureFlagConfigMapRetriever`

A key of a ConfigMap, read through the Kubernetes API on every poll -
the way to serve a KubernetesGoFeatureFlagFlagFile. The module grants
the relay `get` on exactly this ConfigMap.

### spec.flagSource.retrievers[].configMap.configMapName

`string | valueFrom` · required

ConfigMap name. Accepts a literal or a reference to a
KubernetesGoFeatureFlagFlagFile (its rendered ConfigMap); any ConfigMap
holding a GO Feature Flag flag file works.

- references: KubernetesGoFeatureFlagFlagFile (`status.outputs.config_map_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesGoFeatureFlagFlagFile, name: <that resource's name>, fieldPath: status.outputs.config_map_name}} -- a bare string does not parse

### spec.flagSource.retrievers[].configMap.key

`string | valueFrom` · required

Key within the ConfigMap holding the flag file. Accepts a literal or a
reference to the same KubernetesGoFeatureFlagFlagFile (its rendered
key, `flags.goff.yaml` unless the flag file says otherwise).

- references: KubernetesGoFeatureFlagFlagFile (`status.outputs.key`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesGoFeatureFlagFlagFile, name: <that resource's name>, fieldPath: status.outputs.key}} -- a bare string does not parse

### spec.flagSource.retrievers[].configMap.namespace

`string | valueFrom`

Namespace of the ConfigMap. Empty = the relay's own namespace. The
module grants the read in whichever namespace this names.

- references: KubernetesNamespace (`spec.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.flagSource.retrievers[].http

`KubernetesGoFeatureFlagHttpRetriever`

A flag file served over HTTP(S).

### spec.flagSource.retrievers[].http.url

`string` · required

URL of the flag file.

- rule: {"required":true,"string":{"uri":true}}

### spec.flagSource.retrievers[].http.method

`string`

HTTP method, upper case (e.g. GET or POST). Empty = GET.

- rule: The HTTP method must be an upper-case token such as GET or POST.

### spec.flagSource.retrievers[].http.body

`string`

Request body (for POST/PUT/PATCH endpoints).

### spec.flagSource.retrievers[].http.headers

`map<string, string>`

Plain request headers. Repeat a header by joining its values with
", " (equivalent on the wire).

### spec.flagSource.retrievers[].http.sensitiveHeaders

`map<string, string>` · sensitive

Request headers carrying credentials (e.g. Authorization). Header names
may not contain "_" and may not repeat a plain header; the module
checks both at deploy time.

### spec.flagSource.retrievers[].http.timeoutMs

`int32` · optional (explicit presence)

Request timeout in milliseconds. Empty = 10000.

- rule: {"int32":{"gt":0}}

### spec.flagSource.retrievers[].github

`KubernetesGoFeatureFlagGitRetriever`

A flag file in a GitHub repository.

### spec.flagSource.retrievers[].github.repositorySlug

`string` · required

Repository as `owner/name` (GitLab: the project path, groups
included).

- rule: {"required":true}

### spec.flagSource.retrievers[].github.path

`string` · required

Path of the flag file in the repository.

- rule: {"required":true}

### spec.flagSource.retrievers[].github.branch

`string`

Branch to read. Empty = main.

### spec.flagSource.retrievers[].github.token

`string` · sensitive

Access token. Required for private repositories; also lifts the
provider's anonymous rate limit, which a short polling interval
otherwise exhausts.

### spec.flagSource.retrievers[].github.baseUrl

`string`

API base URL for a self-hosted server (GitHub Enterprise, a GitLab or
Bitbucket instance). Empty = the public service.

- rule: base_url must be an absolute http(s) URL.

### spec.flagSource.retrievers[].github.timeoutMs

`int32` · optional (explicit presence)

Request timeout in milliseconds. Empty = 10000.

- rule: {"int32":{"gt":0}}

### spec.flagSource.retrievers[].gitlab

`KubernetesGoFeatureFlagGitRetriever`

A flag file in a GitLab repository.

### spec.flagSource.retrievers[].gitlab.repositorySlug

`string` · required

Repository as `owner/name` (GitLab: the project path, groups
included).

- rule: {"required":true}

### spec.flagSource.retrievers[].gitlab.path

`string` · required

Path of the flag file in the repository.

- rule: {"required":true}

### spec.flagSource.retrievers[].gitlab.branch

`string`

Branch to read. Empty = main.

### spec.flagSource.retrievers[].gitlab.token

`string` · sensitive

Access token. Required for private repositories; also lifts the
provider's anonymous rate limit, which a short polling interval
otherwise exhausts.

### spec.flagSource.retrievers[].gitlab.baseUrl

`string`

API base URL for a self-hosted server (GitHub Enterprise, a GitLab or
Bitbucket instance). Empty = the public service.

- rule: base_url must be an absolute http(s) URL.

### spec.flagSource.retrievers[].gitlab.timeoutMs

`int32` · optional (explicit presence)

Request timeout in milliseconds. Empty = 10000.

- rule: {"int32":{"gt":0}}

### spec.flagSource.retrievers[].bitbucket

`KubernetesGoFeatureFlagGitRetriever`

A flag file in a Bitbucket repository.

### spec.flagSource.retrievers[].bitbucket.repositorySlug

`string` · required

Repository as `owner/name` (GitLab: the project path, groups
included).

- rule: {"required":true}

### spec.flagSource.retrievers[].bitbucket.path

`string` · required

Path of the flag file in the repository.

- rule: {"required":true}

### spec.flagSource.retrievers[].bitbucket.branch

`string`

Branch to read. Empty = main.

### spec.flagSource.retrievers[].bitbucket.token

`string` · sensitive

Access token. Required for private repositories; also lifts the
provider's anonymous rate limit, which a short polling interval
otherwise exhausts.

### spec.flagSource.retrievers[].bitbucket.baseUrl

`string`

API base URL for a self-hosted server (GitHub Enterprise, a GitLab or
Bitbucket instance). Empty = the public service.

- rule: base_url must be an absolute http(s) URL.

### spec.flagSource.retrievers[].bitbucket.timeoutMs

`int32` · optional (explicit presence)

Request timeout in milliseconds. Empty = 10000.

- rule: {"int32":{"gt":0}}

### spec.flagSource.retrievers[].s3

`KubernetesGoFeatureFlagS3Retriever`

A flag file in an S3 bucket (credentials from the standard AWS SDK
chain: workload identity, or AWS_* variables).

### spec.flagSource.retrievers[].s3.bucket

`string` · required

Bucket name.

- rule: {"required":true}

### spec.flagSource.retrievers[].s3.item

`string` · required

Object key of the flag file.

- rule: {"required":true}

### spec.flagSource.retrievers[].googleStorage

`KubernetesGoFeatureFlagGcsRetriever`

A flag file in a Google Cloud Storage bucket (credentials from
workload identity or GOOGLE_APPLICATION_CREDENTIALS).

### spec.flagSource.retrievers[].googleStorage.bucket

`string` · required

Bucket name.

- rule: {"required":true}

### spec.flagSource.retrievers[].googleStorage.object

`string` · required

Object name of the flag file.

- rule: {"required":true}

### spec.flagSource.retrievers[].azureBlobStorage

`KubernetesGoFeatureFlagAzureBlobRetriever`

A flag file in Azure Blob Storage.

### spec.flagSource.retrievers[].azureBlobStorage.accountName

`string` · required

Storage account name.

- rule: {"required":true}

### spec.flagSource.retrievers[].azureBlobStorage.accountKey

`string` · sensitive

Storage account key. Empty = the Azure default credential chain
(workload identity or a managed identity).

### spec.flagSource.retrievers[].azureBlobStorage.container

`string` · required

Blob container.

- rule: {"required":true}

### spec.flagSource.retrievers[].azureBlobStorage.object

`string` · required

Blob name of the flag file.

- rule: {"required":true}

### spec.flagSource.retrievers[].mongodb

`KubernetesGoFeatureFlagMongoDbRetriever`

Flags stored as documents in a MongoDB collection.

### spec.flagSource.retrievers[].mongodb.uri

`string` · required · sensitive

Connection URI (credentials included).

- rule: {"required":true}

### spec.flagSource.retrievers[].mongodb.database

`string` · required

Database holding the flags collection.

- rule: {"required":true}

### spec.flagSource.retrievers[].mongodb.collection

`string` · required

Collection with one document per flag (the document's `flag` field
names the flag).

- rule: {"required":true}

### spec.flagSource.retrievers[].redis

`KubernetesGoFeatureFlagRedisRetriever`

Flags stored as Redis keys.

### spec.flagSource.retrievers[].redis.options

`KubernetesGoFeatureFlagRedisOptions` · required

Redis connection.

- rule: {"required":true}

### spec.flagSource.retrievers[].redis.options.addr

`string` · required

Server address as host:port.

- rule: {"required":true}

### spec.flagSource.retrievers[].redis.options.network

`string`

Network: tcp or unix. Empty = tcp.

- rule: Network must be tcp or unix.

### spec.flagSource.retrievers[].redis.options.username

`string`

Username (Redis 6 ACL).

### spec.flagSource.retrievers[].redis.options.password

`string` · sensitive

Password.

### spec.flagSource.retrievers[].redis.options.db

`int32`

Database number selected after connecting.

- rule: {"int32":{"gte":0}}

### spec.flagSource.retrievers[].redis.options.tlsEnabled

`bool`

Connect with TLS (minimum TLS 1.2).

### spec.flagSource.retrievers[].redis.options.protocol

`int32` · optional (explicit presence)

RESP protocol version: 2 or 3. Empty = 3.

- rule: RESP protocol must be 2 or 3.

### spec.flagSource.retrievers[].redis.options.clientName

`string`

Name set on every connection with CLIENT SETNAME.

### spec.flagSource.retrievers[].redis.options.identitySuffix

`string`

Suffix appended to the client library identity.

### spec.flagSource.retrievers[].redis.options.disableIdentity

`bool`

Do not send the client library identity (CLIENT SETINFO) on connect.

### spec.flagSource.retrievers[].redis.options.maxRetries

`int32` · optional (explicit presence)

Retries before giving up. -1 disables retries. Empty = 3.

- rule: {"int32":{"gte":-1}}

### spec.flagSource.retrievers[].redis.options.minRetryBackoffMs

`int64` · optional (explicit presence)

Minimum backoff between retries. -1 disables backoff. Empty = 8.

- rule: {"int64":{"gte":"-1"}}

### spec.flagSource.retrievers[].redis.options.maxRetryBackoffMs

`int64` · optional (explicit presence)

Maximum backoff between retries. -1 disables backoff. Empty = 512.

- rule: {"int64":{"gte":"-1"}}

### spec.flagSource.retrievers[].redis.options.dialTimeoutMs

`int64` · optional (explicit presence)

Timeout for establishing a connection. Empty = 5000.

- rule: {"int64":{"gt":"0"}}

### spec.flagSource.retrievers[].redis.options.readTimeoutMs

`int64` · optional (explicit presence)

Socket read timeout. -1 = no timeout, -2 = never set a read deadline.
Empty = 3000.

- rule: {"int64":{"gte":"-2"}}

### spec.flagSource.retrievers[].redis.options.writeTimeoutMs

`int64` · optional (explicit presence)

Socket write timeout. Empty = the read timeout.

- rule: {"int64":{"gte":"-2"}}

### spec.flagSource.retrievers[].redis.options.contextTimeoutEnabled

`bool`

Honor context deadlines on commands.

### spec.flagSource.retrievers[].redis.options.poolFifo

`bool`

Use FIFO instead of LIFO for the connection pool.

### spec.flagSource.retrievers[].redis.options.poolSize

`int32` · optional (explicit presence)

Maximum socket connections. Empty = 10 per available CPU.

- rule: {"int32":{"gt":0}}

### spec.flagSource.retrievers[].redis.options.poolTimeoutMs

`int64` · optional (explicit presence)

How long to wait for a pooled connection when all are busy. Empty =
the read timeout plus one second.

- rule: {"int64":{"gt":"0"}}

### spec.flagSource.retrievers[].redis.options.minIdleConns

`int32`

Minimum idle connections kept open.

- rule: {"int32":{"gte":0}}

### spec.flagSource.retrievers[].redis.options.maxIdleConns

`int32`

Maximum idle connections. 0 = unlimited.

- rule: {"int32":{"gte":0}}

### spec.flagSource.retrievers[].redis.options.connMaxIdleTimeMs

`int64` · optional (explicit presence)

Maximum idle time of a connection. -1 disables the check. Empty =
1800000 (30 minutes).

- rule: {"int64":{"gte":"-1"}}

### spec.flagSource.retrievers[].redis.options.connMaxLifetimeMs

`int64` · optional (explicit presence)

Maximum lifetime of a connection. Empty = connections are never
retired for age.

- rule: {"int64":{"gt":"0"}}

### spec.flagSource.retrievers[].redis.prefix

`string`

Key prefix: every key starting with it holds one flag (the rest of the
key is the flag name). Empty = every key.

### spec.flagSource.retrievers[].postgresql

`KubernetesGoFeatureFlagPostgresRetriever`

Flags stored as rows in a PostgreSQL table.

### spec.flagSource.retrievers[].postgresql.uri

`string` · required · sensitive

Connection URI (credentials included), e.g.
postgres://user:password@host:5432/db?sslmode=require.

- rule: {"required":true}

### spec.flagSource.retrievers[].postgresql.table

`string` · required

Table holding one row per flag.

- rule: {"required":true}

### spec.flagSource.retrievers[].postgresql.columns

`map<string, string>`

Column mapping, from the relay's field name (`flag_name`, `flagset`,
`config`) to the column that holds it. Empty = those same names.

- rule: Column mappings may only name the relay's fields: flag_name, flagset, config.

### spec.flagSource.notifiers

`[]KubernetesGoFeatureFlagNotifier`

Notifiers told about every flag change (created, updated, deleted).

### spec.flagSource.notifiers[].slack

`KubernetesGoFeatureFlagWebhookUrlNotifier`

A Slack incoming webhook.

### spec.flagSource.notifiers[].slack.webhookUrl

`string` · required · sensitive

The incoming-webhook URL (it embeds the channel's credential).

- rule: {"required":true}

### spec.flagSource.notifiers[].microsoftTeams

`KubernetesGoFeatureFlagWebhookUrlNotifier`

A Microsoft Teams incoming webhook.

### spec.flagSource.notifiers[].microsoftTeams.webhookUrl

`string` · required · sensitive

The incoming-webhook URL (it embeds the channel's credential).

- rule: {"required":true}

### spec.flagSource.notifiers[].discord

`KubernetesGoFeatureFlagWebhookUrlNotifier`

A Discord webhook.

### spec.flagSource.notifiers[].discord.webhookUrl

`string` · required · sensitive

The incoming-webhook URL (it embeds the channel's credential).

- rule: {"required":true}

### spec.flagSource.notifiers[].webhook

`KubernetesGoFeatureFlagWebhookNotifier`

Any HTTP endpoint, with an optional HMAC signature.

### spec.flagSource.notifiers[].webhook.endpointUrl

`string` · required

Endpoint receiving a POST with the change payload.

- rule: {"required":true,"string":{"uri":true}}

### spec.flagSource.notifiers[].webhook.secret

`string` · sensitive

Secret used to sign each payload (HMAC-SHA256 in the
X-Hub-Signature-256 header), so the receiver can verify it.

### spec.flagSource.notifiers[].webhook.meta

`map<string, string>`

Static fields added to every payload's `meta` (e.g. the environment).

### spec.flagSource.notifiers[].webhook.headers

`map<string, string>`

Plain request headers.

### spec.flagSource.notifiers[].webhook.sensitiveHeaders

`map<string, string>` · sensitive

Request headers carrying credentials. Names may not contain "_" and
may not repeat a plain header (checked at deploy time).

### spec.flagSource.exporters

`[]KubernetesGoFeatureFlagExporter`

Exporters that receive evaluation events.

### spec.flagSource.exporters[].webhook

`KubernetesGoFeatureFlagWebhookExporter`

POST batches of events to an HTTP endpoint.

### spec.flagSource.exporters[].webhook.endpointUrl

`string` · required

Endpoint receiving a POST per batch.

- rule: {"required":true,"string":{"uri":true}}

### spec.flagSource.exporters[].webhook.secret

`string` · sensitive

Secret used to sign each batch (HMAC-SHA256 in X-Hub-Signature-256).

### spec.flagSource.exporters[].webhook.meta

`map<string, string>`

Static fields added to every batch's `meta`.

### spec.flagSource.exporters[].webhook.headers

`map<string, string>`

Plain request headers.

### spec.flagSource.exporters[].webhook.sensitiveHeaders

`map<string, string>` · sensitive

Request headers carrying credentials. Names may not contain "_" and
may not repeat a plain header (checked at deploy time).

### spec.flagSource.exporters[].log

`KubernetesGoFeatureFlagLogExporter`

Write one log line per event to the relay's log.

### spec.flagSource.exporters[].log.logFormat

`string`

Go template for each line. Empty =
`[{{ .FormattedDate}}] user="{{ .UserKey}}", flag="{{ .Key}}", value="{{ .Value}}"`.

### spec.flagSource.exporters[].s3

`KubernetesGoFeatureFlagS3Exporter`

Write event files to an S3 bucket.

### spec.flagSource.exporters[].s3.bucket

`string` · required

Bucket name.

- rule: {"required":true}

### spec.flagSource.exporters[].s3.path

`string`

Key prefix (folder) for the files. Empty = the bucket root.

### spec.flagSource.exporters[].s3.file

`KubernetesGoFeatureFlagEventFileFormat`

File format and naming.

### spec.flagSource.exporters[].s3.file.format

`string`

File format: JSON, CSV or Parquet. Empty = JSON.

- rule: File format must be JSON, CSV or Parquet (any case).

### spec.flagSource.exporters[].s3.file.filename

`string`

Go template for each file name. Empty =
`flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`.

### spec.flagSource.exporters[].s3.file.csvTemplate

`string`

Go template for each CSV line (CSV format only). Empty = every event
field, separated by ";".

### spec.flagSource.exporters[].s3.file.parquetCompressionCodec

`string`

Parquet compression codec (Parquet format only): UNCOMPRESSED, SNAPPY,
GZIP, LZO, BROTLI, LZ4, ZSTD or LZ4_RAW. Empty = SNAPPY.

- rule: Parquet codec must be one of: UNCOMPRESSED, SNAPPY, GZIP, LZO, BROTLI, LZ4, ZSTD, LZ4_RAW.

### spec.flagSource.exporters[].googleStorage

`KubernetesGoFeatureFlagGcsExporter`

Write event files to a Google Cloud Storage bucket.

### spec.flagSource.exporters[].googleStorage.bucket

`string` · required

Bucket name.

- rule: {"required":true}

### spec.flagSource.exporters[].googleStorage.path

`string`

Object prefix (folder) for the files. Empty = the bucket root.

### spec.flagSource.exporters[].googleStorage.file

`KubernetesGoFeatureFlagEventFileFormat`

File format and naming.

### spec.flagSource.exporters[].googleStorage.file.format

`string`

File format: JSON, CSV or Parquet. Empty = JSON.

- rule: File format must be JSON, CSV or Parquet (any case).

### spec.flagSource.exporters[].googleStorage.file.filename

`string`

Go template for each file name. Empty =
`flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`.

### spec.flagSource.exporters[].googleStorage.file.csvTemplate

`string`

Go template for each CSV line (CSV format only). Empty = every event
field, separated by ";".

### spec.flagSource.exporters[].googleStorage.file.parquetCompressionCodec

`string`

Parquet compression codec (Parquet format only): UNCOMPRESSED, SNAPPY,
GZIP, LZO, BROTLI, LZ4, ZSTD or LZ4_RAW. Empty = SNAPPY.

- rule: Parquet codec must be one of: UNCOMPRESSED, SNAPPY, GZIP, LZO, BROTLI, LZ4, ZSTD, LZ4_RAW.

### spec.flagSource.exporters[].azureBlobStorage

`KubernetesGoFeatureFlagAzureBlobExporter`

Write event files to Azure Blob Storage.

### spec.flagSource.exporters[].azureBlobStorage.accountName

`string` · required

Storage account name.

- rule: {"required":true}

### spec.flagSource.exporters[].azureBlobStorage.accountKey

`string` · sensitive

Storage account key. Empty = the Azure default credential chain.

### spec.flagSource.exporters[].azureBlobStorage.container

`string` · required

Blob container.

- rule: {"required":true}

### spec.flagSource.exporters[].azureBlobStorage.path

`string`

Blob prefix (folder) for the files. Empty = the container root.

### spec.flagSource.exporters[].azureBlobStorage.file

`KubernetesGoFeatureFlagEventFileFormat`

File format and naming.

### spec.flagSource.exporters[].azureBlobStorage.file.format

`string`

File format: JSON, CSV or Parquet. Empty = JSON.

- rule: File format must be JSON, CSV or Parquet (any case).

### spec.flagSource.exporters[].azureBlobStorage.file.filename

`string`

Go template for each file name. Empty =
`flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`.

### spec.flagSource.exporters[].azureBlobStorage.file.csvTemplate

`string`

Go template for each CSV line (CSV format only). Empty = every event
field, separated by ";".

### spec.flagSource.exporters[].azureBlobStorage.file.parquetCompressionCodec

`string`

Parquet compression codec (Parquet format only): UNCOMPRESSED, SNAPPY,
GZIP, LZO, BROTLI, LZ4, ZSTD or LZ4_RAW. Empty = SNAPPY.

- rule: Parquet codec must be one of: UNCOMPRESSED, SNAPPY, GZIP, LZO, BROTLI, LZ4, ZSTD, LZ4_RAW.

### spec.flagSource.exporters[].sqs

`KubernetesGoFeatureFlagSqsExporter`

Send events to an SQS queue.

### spec.flagSource.exporters[].sqs.queueUrl

`string` · required

Queue URL.

- rule: {"required":true,"string":{"uri":true}}

### spec.flagSource.exporters[].kinesis

`KubernetesGoFeatureFlagKinesisExporter`

Send events to a Kinesis data stream.

- rule: Name the Kinesis stream by stream_arn or stream_name.

### spec.flagSource.exporters[].kinesis.streamArn

`string`

Stream ARN. Give the ARN or the name.

### spec.flagSource.exporters[].kinesis.streamName

`string`

Stream name. Give the ARN or the name.

### spec.flagSource.exporters[].kinesis.format

`string`

Record format. Empty = JSON (the only format the relay produces for
streams).

### spec.flagSource.exporters[].pubsub

`KubernetesGoFeatureFlagPubSubExporter`

Publish events to a Google Cloud Pub/Sub topic.

### spec.flagSource.exporters[].pubsub.projectId

`string` · required

Google Cloud project holding the topic.

- rule: {"required":true}

### spec.flagSource.exporters[].pubsub.topic

`string` · required

Topic name.

- rule: {"required":true}

### spec.flagSource.exporters[].bigquery

`KubernetesGoFeatureFlagBigQueryExporter`

Stream events into a BigQuery table.

### spec.flagSource.exporters[].bigquery.projectId

`string` · required

Google Cloud project holding the dataset.

- rule: {"required":true}

### spec.flagSource.exporters[].bigquery.datasetId

`string` · required

Dataset holding the table.

- rule: {"required":true}

### spec.flagSource.exporters[].bigquery.tableName

`string`

Table name. Empty = the relay's default table name.

### spec.flagSource.exporters[].bigquery.googleCredentials

`string` · sensitive

Service account key JSON. Empty = workload identity or
GOOGLE_APPLICATION_CREDENTIALS.

### spec.flagSource.exporters[].bigquery.autoMigrate

`bool`

Create or update the table schema automatically.

### spec.flagSource.exporters[].kafka

`KubernetesGoFeatureFlagKafkaExporter`

Produce events to a Kafka topic.

### spec.flagSource.exporters[].kafka.topic

`string` · required

Topic events are produced to.

- rule: {"required":true}

### spec.flagSource.exporters[].kafka.addresses

`[]string` · required

Bootstrap broker addresses (host:port).

- rule: {"repeated":{"minItems":"1"}}

### spec.flagSource.exporters[].kafka.config

`object`

Producer tuning: a Sarama client configuration object
(https://pkg.go.dev/github.com/IBM/sarama#Config), merged over Sarama's
defaults - e.g. {"net": {"sasl": {"enable": true, "mechanism":
"SCRAM-SHA-512", "user": "events"}, "tls": {"enable": true}}}. Keys are
matched case-insensitively. Put the SASL password in sasl_password,
never here.

### spec.flagSource.exporters[].kafka.saslPassword

`string` · sensitive

SASL password, merged into config as net.sasl.password.

### spec.flagSource.exporters[].opentelemetry

`KubernetesGoFeatureFlagOpenTelemetryExporter`

Record events as OpenTelemetry span events on the relay's traces.

### spec.flagSource.exporters[].opentelemetry.tracerName

`string`

Tracer name the span events are recorded under. Empty = the relay's
default tracer.

### spec.flagSource.exporters[].flushIntervalMs

`int64` · optional (explicit presence)

How often a batch is flushed, in milliseconds. Empty = 60000.

- rule: {"int64":{"gt":"0"}}

### spec.flagSource.exporters[].maxEventInMemory

`int64` · optional (explicit presence)

Events buffered before a forced flush. Empty = 100000.

- rule: {"int64":{"gt":"0"}}

### spec.flagSource.exporters[].eventType

`string`

Which events this exporter receives: feature (flag evaluations) or
tracking (custom tracking events). Empty = feature.

- rule: event_type must be feature or tracking.

### spec.flagSource.fileFormat

`string`

Format of the flag files: yaml, json or toml. Empty = yaml (what a
KubernetesGoFeatureFlagFlagFile renders).

- rule: File format must be one of: yaml, json, toml.

### spec.flagSource.pollingIntervalMs

`int32` · optional (explicit presence)

How often the retrievers are re-read, in milliseconds (minimum 1000).
A negative value disables polling (flags load once at startup). Empty =
60000. A flag flip reaches evaluations within one interval.

- rule: The polling interval must be at least 1000 ms, or negative to disable polling.

### spec.flagSource.startWithRetrieverError

`bool` · optional (explicit presence)

Start serving even when a retriever fails at startup (flags then load
on the first successful poll). Empty = false: a retriever that cannot
be read keeps the relay from starting. Set true when the flag file is
deployed alongside the relay, so install order never matters.

### spec.flagSource.enablePollingJitter

`bool`

Add up to +-10% random jitter to every poll, so many replicas never hit
the retrievers at the same instant.

### spec.flagSource.disableNotifierOnInit

`bool`

Do not notify for the initial load at startup (only for later changes).

### spec.flagSource.evaluationContextEnrichment

`object`

Attributes merged into every evaluation context (a field here
overrides the same field sent by the caller) - e.g. the environment
name or a region, so targeting rules can use them.

### spec.flagSets

`KubernetesGoFeatureFlagFlagSets`

Several isolated flag sets, each selected by the API key the caller
presents, each with its own retrievers, notifiers and exporters
(there is no inheritance between flag sets).

- rule: Flag set names must be unique (the relay keys flag sets by name, so a second set with the same name takes over the first set's keys).
- rule: "default" is reserved by the relay; name the flag set something else.

### spec.flagSets.items

`[]KubernetesGoFeatureFlagFlagSet` · required

The flag sets. Every caller's API key selects exactly one of them.

- rule: {"repeated":{"minItems":"1"}}

### spec.flagSets.items[].name

`string`

Flag set name, shown in notifications and exported events. Unique across
the flag sets; "default" is reserved. Empty = the relay generates one.

### spec.flagSets.items[].apiKeys

`[]string` · required · sensitive

API keys that select this flag set. At least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.flagSets.items[].source

`KubernetesGoFeatureFlagFlagSource` · required

This flag set's retrievers, notifiers and exporters.

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers

`[]KubernetesGoFeatureFlagRetriever` · required

Retrievers the flag files are read from. When several are listed, the
relay merges their flags and a later retriever wins on a flag both
define.

- rule: {"repeated":{"minItems":"1"}}

### spec.flagSets.items[].source.retrievers[].configMap

`KubernetesGoFeatureFlagConfigMapRetriever`

A key of a ConfigMap, read through the Kubernetes API on every poll -
the way to serve a KubernetesGoFeatureFlagFlagFile. The module grants
the relay `get` on exactly this ConfigMap.

### spec.flagSets.items[].source.retrievers[].configMap.configMapName

`string | valueFrom` · required

ConfigMap name. Accepts a literal or a reference to a
KubernetesGoFeatureFlagFlagFile (its rendered ConfigMap); any ConfigMap
holding a GO Feature Flag flag file works.

- references: KubernetesGoFeatureFlagFlagFile (`status.outputs.config_map_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesGoFeatureFlagFlagFile, name: <that resource's name>, fieldPath: status.outputs.config_map_name}} -- a bare string does not parse

### spec.flagSets.items[].source.retrievers[].configMap.key

`string | valueFrom` · required

Key within the ConfigMap holding the flag file. Accepts a literal or a
reference to the same KubernetesGoFeatureFlagFlagFile (its rendered
key, `flags.goff.yaml` unless the flag file says otherwise).

- references: KubernetesGoFeatureFlagFlagFile (`status.outputs.key`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesGoFeatureFlagFlagFile, name: <that resource's name>, fieldPath: status.outputs.key}} -- a bare string does not parse

### spec.flagSets.items[].source.retrievers[].configMap.namespace

`string | valueFrom`

Namespace of the ConfigMap. Empty = the relay's own namespace. The
module grants the read in whichever namespace this names.

- references: KubernetesNamespace (`spec.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.flagSets.items[].source.retrievers[].http

`KubernetesGoFeatureFlagHttpRetriever`

A flag file served over HTTP(S).

### spec.flagSets.items[].source.retrievers[].http.url

`string` · required

URL of the flag file.

- rule: {"required":true,"string":{"uri":true}}

### spec.flagSets.items[].source.retrievers[].http.method

`string`

HTTP method, upper case (e.g. GET or POST). Empty = GET.

- rule: The HTTP method must be an upper-case token such as GET or POST.

### spec.flagSets.items[].source.retrievers[].http.body

`string`

Request body (for POST/PUT/PATCH endpoints).

### spec.flagSets.items[].source.retrievers[].http.headers

`map<string, string>`

Plain request headers. Repeat a header by joining its values with
", " (equivalent on the wire).

### spec.flagSets.items[].source.retrievers[].http.sensitiveHeaders

`map<string, string>` · sensitive

Request headers carrying credentials (e.g. Authorization). Header names
may not contain "_" and may not repeat a plain header; the module
checks both at deploy time.

### spec.flagSets.items[].source.retrievers[].http.timeoutMs

`int32` · optional (explicit presence)

Request timeout in milliseconds. Empty = 10000.

- rule: {"int32":{"gt":0}}

### spec.flagSets.items[].source.retrievers[].github

`KubernetesGoFeatureFlagGitRetriever`

A flag file in a GitHub repository.

### spec.flagSets.items[].source.retrievers[].github.repositorySlug

`string` · required

Repository as `owner/name` (GitLab: the project path, groups
included).

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].github.path

`string` · required

Path of the flag file in the repository.

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].github.branch

`string`

Branch to read. Empty = main.

### spec.flagSets.items[].source.retrievers[].github.token

`string` · sensitive

Access token. Required for private repositories; also lifts the
provider's anonymous rate limit, which a short polling interval
otherwise exhausts.

### spec.flagSets.items[].source.retrievers[].github.baseUrl

`string`

API base URL for a self-hosted server (GitHub Enterprise, a GitLab or
Bitbucket instance). Empty = the public service.

- rule: base_url must be an absolute http(s) URL.

### spec.flagSets.items[].source.retrievers[].github.timeoutMs

`int32` · optional (explicit presence)

Request timeout in milliseconds. Empty = 10000.

- rule: {"int32":{"gt":0}}

### spec.flagSets.items[].source.retrievers[].gitlab

`KubernetesGoFeatureFlagGitRetriever`

A flag file in a GitLab repository.

### spec.flagSets.items[].source.retrievers[].gitlab.repositorySlug

`string` · required

Repository as `owner/name` (GitLab: the project path, groups
included).

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].gitlab.path

`string` · required

Path of the flag file in the repository.

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].gitlab.branch

`string`

Branch to read. Empty = main.

### spec.flagSets.items[].source.retrievers[].gitlab.token

`string` · sensitive

Access token. Required for private repositories; also lifts the
provider's anonymous rate limit, which a short polling interval
otherwise exhausts.

### spec.flagSets.items[].source.retrievers[].gitlab.baseUrl

`string`

API base URL for a self-hosted server (GitHub Enterprise, a GitLab or
Bitbucket instance). Empty = the public service.

- rule: base_url must be an absolute http(s) URL.

### spec.flagSets.items[].source.retrievers[].gitlab.timeoutMs

`int32` · optional (explicit presence)

Request timeout in milliseconds. Empty = 10000.

- rule: {"int32":{"gt":0}}

### spec.flagSets.items[].source.retrievers[].bitbucket

`KubernetesGoFeatureFlagGitRetriever`

A flag file in a Bitbucket repository.

### spec.flagSets.items[].source.retrievers[].bitbucket.repositorySlug

`string` · required

Repository as `owner/name` (GitLab: the project path, groups
included).

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].bitbucket.path

`string` · required

Path of the flag file in the repository.

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].bitbucket.branch

`string`

Branch to read. Empty = main.

### spec.flagSets.items[].source.retrievers[].bitbucket.token

`string` · sensitive

Access token. Required for private repositories; also lifts the
provider's anonymous rate limit, which a short polling interval
otherwise exhausts.

### spec.flagSets.items[].source.retrievers[].bitbucket.baseUrl

`string`

API base URL for a self-hosted server (GitHub Enterprise, a GitLab or
Bitbucket instance). Empty = the public service.

- rule: base_url must be an absolute http(s) URL.

### spec.flagSets.items[].source.retrievers[].bitbucket.timeoutMs

`int32` · optional (explicit presence)

Request timeout in milliseconds. Empty = 10000.

- rule: {"int32":{"gt":0}}

### spec.flagSets.items[].source.retrievers[].s3

`KubernetesGoFeatureFlagS3Retriever`

A flag file in an S3 bucket (credentials from the standard AWS SDK
chain: workload identity, or AWS_* variables).

### spec.flagSets.items[].source.retrievers[].s3.bucket

`string` · required

Bucket name.

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].s3.item

`string` · required

Object key of the flag file.

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].googleStorage

`KubernetesGoFeatureFlagGcsRetriever`

A flag file in a Google Cloud Storage bucket (credentials from
workload identity or GOOGLE_APPLICATION_CREDENTIALS).

### spec.flagSets.items[].source.retrievers[].googleStorage.bucket

`string` · required

Bucket name.

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].googleStorage.object

`string` · required

Object name of the flag file.

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].azureBlobStorage

`KubernetesGoFeatureFlagAzureBlobRetriever`

A flag file in Azure Blob Storage.

### spec.flagSets.items[].source.retrievers[].azureBlobStorage.accountName

`string` · required

Storage account name.

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].azureBlobStorage.accountKey

`string` · sensitive

Storage account key. Empty = the Azure default credential chain
(workload identity or a managed identity).

### spec.flagSets.items[].source.retrievers[].azureBlobStorage.container

`string` · required

Blob container.

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].azureBlobStorage.object

`string` · required

Blob name of the flag file.

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].mongodb

`KubernetesGoFeatureFlagMongoDbRetriever`

Flags stored as documents in a MongoDB collection.

### spec.flagSets.items[].source.retrievers[].mongodb.uri

`string` · required · sensitive

Connection URI (credentials included).

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].mongodb.database

`string` · required

Database holding the flags collection.

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].mongodb.collection

`string` · required

Collection with one document per flag (the document's `flag` field
names the flag).

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].redis

`KubernetesGoFeatureFlagRedisRetriever`

Flags stored as Redis keys.

### spec.flagSets.items[].source.retrievers[].redis.options

`KubernetesGoFeatureFlagRedisOptions` · required

Redis connection.

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].redis.options.addr

`string` · required

Server address as host:port.

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].redis.options.network

`string`

Network: tcp or unix. Empty = tcp.

- rule: Network must be tcp or unix.

### spec.flagSets.items[].source.retrievers[].redis.options.username

`string`

Username (Redis 6 ACL).

### spec.flagSets.items[].source.retrievers[].redis.options.password

`string` · sensitive

Password.

### spec.flagSets.items[].source.retrievers[].redis.options.db

`int32`

Database number selected after connecting.

- rule: {"int32":{"gte":0}}

### spec.flagSets.items[].source.retrievers[].redis.options.tlsEnabled

`bool`

Connect with TLS (minimum TLS 1.2).

### spec.flagSets.items[].source.retrievers[].redis.options.protocol

`int32` · optional (explicit presence)

RESP protocol version: 2 or 3. Empty = 3.

- rule: RESP protocol must be 2 or 3.

### spec.flagSets.items[].source.retrievers[].redis.options.clientName

`string`

Name set on every connection with CLIENT SETNAME.

### spec.flagSets.items[].source.retrievers[].redis.options.identitySuffix

`string`

Suffix appended to the client library identity.

### spec.flagSets.items[].source.retrievers[].redis.options.disableIdentity

`bool`

Do not send the client library identity (CLIENT SETINFO) on connect.

### spec.flagSets.items[].source.retrievers[].redis.options.maxRetries

`int32` · optional (explicit presence)

Retries before giving up. -1 disables retries. Empty = 3.

- rule: {"int32":{"gte":-1}}

### spec.flagSets.items[].source.retrievers[].redis.options.minRetryBackoffMs

`int64` · optional (explicit presence)

Minimum backoff between retries. -1 disables backoff. Empty = 8.

- rule: {"int64":{"gte":"-1"}}

### spec.flagSets.items[].source.retrievers[].redis.options.maxRetryBackoffMs

`int64` · optional (explicit presence)

Maximum backoff between retries. -1 disables backoff. Empty = 512.

- rule: {"int64":{"gte":"-1"}}

### spec.flagSets.items[].source.retrievers[].redis.options.dialTimeoutMs

`int64` · optional (explicit presence)

Timeout for establishing a connection. Empty = 5000.

- rule: {"int64":{"gt":"0"}}

### spec.flagSets.items[].source.retrievers[].redis.options.readTimeoutMs

`int64` · optional (explicit presence)

Socket read timeout. -1 = no timeout, -2 = never set a read deadline.
Empty = 3000.

- rule: {"int64":{"gte":"-2"}}

### spec.flagSets.items[].source.retrievers[].redis.options.writeTimeoutMs

`int64` · optional (explicit presence)

Socket write timeout. Empty = the read timeout.

- rule: {"int64":{"gte":"-2"}}

### spec.flagSets.items[].source.retrievers[].redis.options.contextTimeoutEnabled

`bool`

Honor context deadlines on commands.

### spec.flagSets.items[].source.retrievers[].redis.options.poolFifo

`bool`

Use FIFO instead of LIFO for the connection pool.

### spec.flagSets.items[].source.retrievers[].redis.options.poolSize

`int32` · optional (explicit presence)

Maximum socket connections. Empty = 10 per available CPU.

- rule: {"int32":{"gt":0}}

### spec.flagSets.items[].source.retrievers[].redis.options.poolTimeoutMs

`int64` · optional (explicit presence)

How long to wait for a pooled connection when all are busy. Empty =
the read timeout plus one second.

- rule: {"int64":{"gt":"0"}}

### spec.flagSets.items[].source.retrievers[].redis.options.minIdleConns

`int32`

Minimum idle connections kept open.

- rule: {"int32":{"gte":0}}

### spec.flagSets.items[].source.retrievers[].redis.options.maxIdleConns

`int32`

Maximum idle connections. 0 = unlimited.

- rule: {"int32":{"gte":0}}

### spec.flagSets.items[].source.retrievers[].redis.options.connMaxIdleTimeMs

`int64` · optional (explicit presence)

Maximum idle time of a connection. -1 disables the check. Empty =
1800000 (30 minutes).

- rule: {"int64":{"gte":"-1"}}

### spec.flagSets.items[].source.retrievers[].redis.options.connMaxLifetimeMs

`int64` · optional (explicit presence)

Maximum lifetime of a connection. Empty = connections are never
retired for age.

- rule: {"int64":{"gt":"0"}}

### spec.flagSets.items[].source.retrievers[].redis.prefix

`string`

Key prefix: every key starting with it holds one flag (the rest of the
key is the flag name). Empty = every key.

### spec.flagSets.items[].source.retrievers[].postgresql

`KubernetesGoFeatureFlagPostgresRetriever`

Flags stored as rows in a PostgreSQL table.

### spec.flagSets.items[].source.retrievers[].postgresql.uri

`string` · required · sensitive

Connection URI (credentials included), e.g.
postgres://user:password@host:5432/db?sslmode=require.

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].postgresql.table

`string` · required

Table holding one row per flag.

- rule: {"required":true}

### spec.flagSets.items[].source.retrievers[].postgresql.columns

`map<string, string>`

Column mapping, from the relay's field name (`flag_name`, `flagset`,
`config`) to the column that holds it. Empty = those same names.

- rule: Column mappings may only name the relay's fields: flag_name, flagset, config.

### spec.flagSets.items[].source.notifiers

`[]KubernetesGoFeatureFlagNotifier`

Notifiers told about every flag change (created, updated, deleted).

### spec.flagSets.items[].source.notifiers[].slack

`KubernetesGoFeatureFlagWebhookUrlNotifier`

A Slack incoming webhook.

### spec.flagSets.items[].source.notifiers[].slack.webhookUrl

`string` · required · sensitive

The incoming-webhook URL (it embeds the channel's credential).

- rule: {"required":true}

### spec.flagSets.items[].source.notifiers[].microsoftTeams

`KubernetesGoFeatureFlagWebhookUrlNotifier`

A Microsoft Teams incoming webhook.

### spec.flagSets.items[].source.notifiers[].microsoftTeams.webhookUrl

`string` · required · sensitive

The incoming-webhook URL (it embeds the channel's credential).

- rule: {"required":true}

### spec.flagSets.items[].source.notifiers[].discord

`KubernetesGoFeatureFlagWebhookUrlNotifier`

A Discord webhook.

### spec.flagSets.items[].source.notifiers[].discord.webhookUrl

`string` · required · sensitive

The incoming-webhook URL (it embeds the channel's credential).

- rule: {"required":true}

### spec.flagSets.items[].source.notifiers[].webhook

`KubernetesGoFeatureFlagWebhookNotifier`

Any HTTP endpoint, with an optional HMAC signature.

### spec.flagSets.items[].source.notifiers[].webhook.endpointUrl

`string` · required

Endpoint receiving a POST with the change payload.

- rule: {"required":true,"string":{"uri":true}}

### spec.flagSets.items[].source.notifiers[].webhook.secret

`string` · sensitive

Secret used to sign each payload (HMAC-SHA256 in the
X-Hub-Signature-256 header), so the receiver can verify it.

### spec.flagSets.items[].source.notifiers[].webhook.meta

`map<string, string>`

Static fields added to every payload's `meta` (e.g. the environment).

### spec.flagSets.items[].source.notifiers[].webhook.headers

`map<string, string>`

Plain request headers.

### spec.flagSets.items[].source.notifiers[].webhook.sensitiveHeaders

`map<string, string>` · sensitive

Request headers carrying credentials. Names may not contain "_" and
may not repeat a plain header (checked at deploy time).

### spec.flagSets.items[].source.exporters

`[]KubernetesGoFeatureFlagExporter`

Exporters that receive evaluation events.

### spec.flagSets.items[].source.exporters[].webhook

`KubernetesGoFeatureFlagWebhookExporter`

POST batches of events to an HTTP endpoint.

### spec.flagSets.items[].source.exporters[].webhook.endpointUrl

`string` · required

Endpoint receiving a POST per batch.

- rule: {"required":true,"string":{"uri":true}}

### spec.flagSets.items[].source.exporters[].webhook.secret

`string` · sensitive

Secret used to sign each batch (HMAC-SHA256 in X-Hub-Signature-256).

### spec.flagSets.items[].source.exporters[].webhook.meta

`map<string, string>`

Static fields added to every batch's `meta`.

### spec.flagSets.items[].source.exporters[].webhook.headers

`map<string, string>`

Plain request headers.

### spec.flagSets.items[].source.exporters[].webhook.sensitiveHeaders

`map<string, string>` · sensitive

Request headers carrying credentials. Names may not contain "_" and
may not repeat a plain header (checked at deploy time).

### spec.flagSets.items[].source.exporters[].log

`KubernetesGoFeatureFlagLogExporter`

Write one log line per event to the relay's log.

### spec.flagSets.items[].source.exporters[].log.logFormat

`string`

Go template for each line. Empty =
`[{{ .FormattedDate}}] user="{{ .UserKey}}", flag="{{ .Key}}", value="{{ .Value}}"`.

### spec.flagSets.items[].source.exporters[].s3

`KubernetesGoFeatureFlagS3Exporter`

Write event files to an S3 bucket.

### spec.flagSets.items[].source.exporters[].s3.bucket

`string` · required

Bucket name.

- rule: {"required":true}

### spec.flagSets.items[].source.exporters[].s3.path

`string`

Key prefix (folder) for the files. Empty = the bucket root.

### spec.flagSets.items[].source.exporters[].s3.file

`KubernetesGoFeatureFlagEventFileFormat`

File format and naming.

### spec.flagSets.items[].source.exporters[].s3.file.format

`string`

File format: JSON, CSV or Parquet. Empty = JSON.

- rule: File format must be JSON, CSV or Parquet (any case).

### spec.flagSets.items[].source.exporters[].s3.file.filename

`string`

Go template for each file name. Empty =
`flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`.

### spec.flagSets.items[].source.exporters[].s3.file.csvTemplate

`string`

Go template for each CSV line (CSV format only). Empty = every event
field, separated by ";".

### spec.flagSets.items[].source.exporters[].s3.file.parquetCompressionCodec

`string`

Parquet compression codec (Parquet format only): UNCOMPRESSED, SNAPPY,
GZIP, LZO, BROTLI, LZ4, ZSTD or LZ4_RAW. Empty = SNAPPY.

- rule: Parquet codec must be one of: UNCOMPRESSED, SNAPPY, GZIP, LZO, BROTLI, LZ4, ZSTD, LZ4_RAW.

### spec.flagSets.items[].source.exporters[].googleStorage

`KubernetesGoFeatureFlagGcsExporter`

Write event files to a Google Cloud Storage bucket.

### spec.flagSets.items[].source.exporters[].googleStorage.bucket

`string` · required

Bucket name.

- rule: {"required":true}

### spec.flagSets.items[].source.exporters[].googleStorage.path

`string`

Object prefix (folder) for the files. Empty = the bucket root.

### spec.flagSets.items[].source.exporters[].googleStorage.file

`KubernetesGoFeatureFlagEventFileFormat`

File format and naming.

### spec.flagSets.items[].source.exporters[].googleStorage.file.format

`string`

File format: JSON, CSV or Parquet. Empty = JSON.

- rule: File format must be JSON, CSV or Parquet (any case).

### spec.flagSets.items[].source.exporters[].googleStorage.file.filename

`string`

Go template for each file name. Empty =
`flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`.

### spec.flagSets.items[].source.exporters[].googleStorage.file.csvTemplate

`string`

Go template for each CSV line (CSV format only). Empty = every event
field, separated by ";".

### spec.flagSets.items[].source.exporters[].googleStorage.file.parquetCompressionCodec

`string`

Parquet compression codec (Parquet format only): UNCOMPRESSED, SNAPPY,
GZIP, LZO, BROTLI, LZ4, ZSTD or LZ4_RAW. Empty = SNAPPY.

- rule: Parquet codec must be one of: UNCOMPRESSED, SNAPPY, GZIP, LZO, BROTLI, LZ4, ZSTD, LZ4_RAW.

### spec.flagSets.items[].source.exporters[].azureBlobStorage

`KubernetesGoFeatureFlagAzureBlobExporter`

Write event files to Azure Blob Storage.

### spec.flagSets.items[].source.exporters[].azureBlobStorage.accountName

`string` · required

Storage account name.

- rule: {"required":true}

### spec.flagSets.items[].source.exporters[].azureBlobStorage.accountKey

`string` · sensitive

Storage account key. Empty = the Azure default credential chain.

### spec.flagSets.items[].source.exporters[].azureBlobStorage.container

`string` · required

Blob container.

- rule: {"required":true}

### spec.flagSets.items[].source.exporters[].azureBlobStorage.path

`string`

Blob prefix (folder) for the files. Empty = the container root.

### spec.flagSets.items[].source.exporters[].azureBlobStorage.file

`KubernetesGoFeatureFlagEventFileFormat`

File format and naming.

### spec.flagSets.items[].source.exporters[].azureBlobStorage.file.format

`string`

File format: JSON, CSV or Parquet. Empty = JSON.

- rule: File format must be JSON, CSV or Parquet (any case).

### spec.flagSets.items[].source.exporters[].azureBlobStorage.file.filename

`string`

Go template for each file name. Empty =
`flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`.

### spec.flagSets.items[].source.exporters[].azureBlobStorage.file.csvTemplate

`string`

Go template for each CSV line (CSV format only). Empty = every event
field, separated by ";".

### spec.flagSets.items[].source.exporters[].azureBlobStorage.file.parquetCompressionCodec

`string`

Parquet compression codec (Parquet format only): UNCOMPRESSED, SNAPPY,
GZIP, LZO, BROTLI, LZ4, ZSTD or LZ4_RAW. Empty = SNAPPY.

- rule: Parquet codec must be one of: UNCOMPRESSED, SNAPPY, GZIP, LZO, BROTLI, LZ4, ZSTD, LZ4_RAW.

### spec.flagSets.items[].source.exporters[].sqs

`KubernetesGoFeatureFlagSqsExporter`

Send events to an SQS queue.

### spec.flagSets.items[].source.exporters[].sqs.queueUrl

`string` · required

Queue URL.

- rule: {"required":true,"string":{"uri":true}}

### spec.flagSets.items[].source.exporters[].kinesis

`KubernetesGoFeatureFlagKinesisExporter`

Send events to a Kinesis data stream.

- rule: Name the Kinesis stream by stream_arn or stream_name.

### spec.flagSets.items[].source.exporters[].kinesis.streamArn

`string`

Stream ARN. Give the ARN or the name.

### spec.flagSets.items[].source.exporters[].kinesis.streamName

`string`

Stream name. Give the ARN or the name.

### spec.flagSets.items[].source.exporters[].kinesis.format

`string`

Record format. Empty = JSON (the only format the relay produces for
streams).

### spec.flagSets.items[].source.exporters[].pubsub

`KubernetesGoFeatureFlagPubSubExporter`

Publish events to a Google Cloud Pub/Sub topic.

### spec.flagSets.items[].source.exporters[].pubsub.projectId

`string` · required

Google Cloud project holding the topic.

- rule: {"required":true}

### spec.flagSets.items[].source.exporters[].pubsub.topic

`string` · required

Topic name.

- rule: {"required":true}

### spec.flagSets.items[].source.exporters[].bigquery

`KubernetesGoFeatureFlagBigQueryExporter`

Stream events into a BigQuery table.

### spec.flagSets.items[].source.exporters[].bigquery.projectId

`string` · required

Google Cloud project holding the dataset.

- rule: {"required":true}

### spec.flagSets.items[].source.exporters[].bigquery.datasetId

`string` · required

Dataset holding the table.

- rule: {"required":true}

### spec.flagSets.items[].source.exporters[].bigquery.tableName

`string`

Table name. Empty = the relay's default table name.

### spec.flagSets.items[].source.exporters[].bigquery.googleCredentials

`string` · sensitive

Service account key JSON. Empty = workload identity or
GOOGLE_APPLICATION_CREDENTIALS.

### spec.flagSets.items[].source.exporters[].bigquery.autoMigrate

`bool`

Create or update the table schema automatically.

### spec.flagSets.items[].source.exporters[].kafka

`KubernetesGoFeatureFlagKafkaExporter`

Produce events to a Kafka topic.

### spec.flagSets.items[].source.exporters[].kafka.topic

`string` · required

Topic events are produced to.

- rule: {"required":true}

### spec.flagSets.items[].source.exporters[].kafka.addresses

`[]string` · required

Bootstrap broker addresses (host:port).

- rule: {"repeated":{"minItems":"1"}}

### spec.flagSets.items[].source.exporters[].kafka.config

`object`

Producer tuning: a Sarama client configuration object
(https://pkg.go.dev/github.com/IBM/sarama#Config), merged over Sarama's
defaults - e.g. {"net": {"sasl": {"enable": true, "mechanism":
"SCRAM-SHA-512", "user": "events"}, "tls": {"enable": true}}}. Keys are
matched case-insensitively. Put the SASL password in sasl_password,
never here.

### spec.flagSets.items[].source.exporters[].kafka.saslPassword

`string` · sensitive

SASL password, merged into config as net.sasl.password.

### spec.flagSets.items[].source.exporters[].opentelemetry

`KubernetesGoFeatureFlagOpenTelemetryExporter`

Record events as OpenTelemetry span events on the relay's traces.

### spec.flagSets.items[].source.exporters[].opentelemetry.tracerName

`string`

Tracer name the span events are recorded under. Empty = the relay's
default tracer.

### spec.flagSets.items[].source.exporters[].flushIntervalMs

`int64` · optional (explicit presence)

How often a batch is flushed, in milliseconds. Empty = 60000.

- rule: {"int64":{"gt":"0"}}

### spec.flagSets.items[].source.exporters[].maxEventInMemory

`int64` · optional (explicit presence)

Events buffered before a forced flush. Empty = 100000.

- rule: {"int64":{"gt":"0"}}

### spec.flagSets.items[].source.exporters[].eventType

`string`

Which events this exporter receives: feature (flag evaluations) or
tracking (custom tracking events). Empty = feature.

- rule: event_type must be feature or tracking.

### spec.flagSets.items[].source.fileFormat

`string`

Format of the flag files: yaml, json or toml. Empty = yaml (what a
KubernetesGoFeatureFlagFlagFile renders).

- rule: File format must be one of: yaml, json, toml.

### spec.flagSets.items[].source.pollingIntervalMs

`int32` · optional (explicit presence)

How often the retrievers are re-read, in milliseconds (minimum 1000).
A negative value disables polling (flags load once at startup). Empty =
60000. A flag flip reaches evaluations within one interval.

- rule: The polling interval must be at least 1000 ms, or negative to disable polling.

### spec.flagSets.items[].source.startWithRetrieverError

`bool` · optional (explicit presence)

Start serving even when a retriever fails at startup (flags then load
on the first successful poll). Empty = false: a retriever that cannot
be read keeps the relay from starting. Set true when the flag file is
deployed alongside the relay, so install order never matters.

### spec.flagSets.items[].source.enablePollingJitter

`bool`

Add up to +-10% random jitter to every poll, so many replicas never hit
the retrievers at the same instant.

### spec.flagSets.items[].source.disableNotifierOnInit

`bool`

Do not notify for the initial load at startup (only for later changes).

### spec.flagSets.items[].source.evaluationContextEnrichment

`object`

Attributes merged into every evaluation context (a field here
overrides the same field sent by the caller) - e.g. the environment
name or a region, so targeting rules can use them.

### spec.ofrepEventStream

`KubernetesGoFeatureFlagOfrepEventStream`

The OFREP flag-change event stream advertised to clients.

### spec.ofrepEventStream.baseUrl

`string`

Public base URL clients use to reach the relay (e.g.
https://flags.example.com); the stream path is appended. Empty = no
event stream is advertised.

- rule: base_url must be an absolute http(s) URL.

### spec.ofrepEventStream.inactivityDelaySec

`int32`

Seconds after which a client treats a silent stream as inactive. Empty
= not advertised.

- rule: {"int32":{"gte":0}}

### spec.telemetry

`KubernetesGoFeatureFlagTelemetry`

OpenTelemetry traces and Jaeger remote sampling.

- rule: jaeger_sampler is read only when traces_sampler is jaeger_remote.
- rule: traces_sampler_arg applies to the traceidratio and parentbased_traceidratio samplers.

### spec.telemetry.otlpEndpoint

`string`

OTLP collector endpoint for the relay's traces (e.g.
http://otel-collector:4318).

### spec.telemetry.otlpProtocol

`string`

OTLP protocol: grpc or http/protobuf (the two the relay's exporter
supports). Empty = http/protobuf.

- rule: OTLP protocol must be grpc or http/protobuf.

### spec.telemetry.sdkDisabled

`bool`

Disable the OpenTelemetry SDK entirely.

### spec.telemetry.serviceName

`string`

Service name reported on traces. Empty = the relay's own name.

### spec.telemetry.tracesSampler

`string`

Trace sampler: an OpenTelemetry sampler name (always_on, always_off,
traceidratio, parentbased_always_on, parentbased_always_off,
parentbased_traceidratio) or jaeger_remote (sampling strategies served
by a Jaeger sampling manager - configure jaeger_sampler). Empty = every
trace is sampled.

- rule: traces_sampler must be one of: always_on, always_off, traceidratio, parentbased_always_on, parentbased_always_off, parentbased_traceidratio, jaeger_remote.

### spec.telemetry.resourceAttributes

`map<string, string>`

Resource attributes added to every span.

### spec.telemetry.jaegerSampler

`KubernetesGoFeatureFlagJaegerSampler`

Jaeger remote sampling (only with traces_sampler jaeger_remote).

### spec.telemetry.jaegerSampler.managerHostPort

`string`

Sampling manager URL (e.g. http://jaeger:5778/sampling).

- rule: manager_host_port must be an absolute http(s) URL of the sampling manager.

### spec.telemetry.jaegerSampler.refreshInterval

`string`

How often the sampling strategy is refreshed (Go duration, e.g. "1m").

- rule: refresh_interval must be a Go duration such as 30s or 1m.

### spec.telemetry.jaegerSampler.maxOperations

`int32`

Maximum operations the sampler tracks.

- rule: {"int32":{"gte":0}}

### spec.telemetry.tracesSamplerArg

`string`

Argument of the ratio samplers: the fraction of traces sampled, 0 to 1
(e.g. "0.25").

- rule: traces_sampler_arg must be a number between 0 and 1.

### spec.swagger

`KubernetesGoFeatureFlagSwagger`

The relay's built-in Swagger UI.

### spec.swagger.enabled

`bool`

Serve the Swagger UI and OpenAPI document on the evaluation port.

### spec.swagger.host

`string`

Host name the Swagger document advertises (the public host clients
use). Empty = localhost.

### spec.runtime

`KubernetesGoFeatureFlagRuntime`

Less common relay switches.

### spec.runtime.hideBanner

`bool`

Do not print the startup banner.

### spec.runtime.enablePprof

`bool`

Serve Go pprof profiling endpoints on the monitoring port.

### spec.runtime.disableVersionHeader

`bool`

Omit the relay version response header.

### spec.runtime.enableBulkMetricFlagNames

`bool`

Label bulk-evaluation metrics with each flag name (higher cardinality).

### spec.runtime.disableFlagDetailsInStream

`bool`

Send only flag names, not flag definitions, in WebSocket change
messages.

### spec.runtime.exporterCleanQueueInterval

`string`

How often exporter queues are cleaned (Go duration, e.g. "1m"). Empty =
the relay default.

- rule: exporter_clean_queue_interval must be a Go duration such as 30s or 1m.

### spec.runtime.envVariablePrefix

`string` · optional (explicit presence)

Prefix the relay requires on the environment variables it reads as
configuration. A prefix keeps the relay from reading the service-link
variables Kubernetes injects for every Service in the namespace (a
Service named `server` injects SERVER_PORT=tcp://..., which an
unprefixed relay reads as its own port and then fails to start). The
module prefixes the variables it generates for secret values to match.

- default: `GOFFRELAY_`
- rule: env_variable_prefix may contain only upper-case letters, digits and underscores.

### spec.metrics

`KubernetesGoFeatureFlagMetrics`

Prometheus scraping. The relay always serves /metrics on the
monitoring port.

### spec.metrics.serviceMonitorEnabled

`bool`

Create a ServiceMonitor for the monitoring port (requires the
Prometheus Operator CRDs on the cluster; the apply FAILS without
them).

### spec.metrics.serviceMonitorLabels

`map<string, string>`

Labels on the ServiceMonitor, to match a Prometheus instance's
serviceMonitorSelector (e.g. release: kube-prometheus-stack).

### spec.scheduling

`KubernetesGoFeatureFlagScheduling`

Pod scheduling constraints.

### spec.scheduling.nodeSelector

`map<string, string>`

Schedule onto nodes carrying these labels.

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

### spec.scheduling.nodeAffinity

`WorkloadNodeAffinity`

Node affinity: hard requirements and weighted preferences.

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

Pod affinity toward pods already running elsewhere.

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

Pod anti-affinity - spread replicas across nodes or zones.

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

### spec.serviceAccount

`KubernetesGoFeatureFlagServiceAccount`

The relay's ServiceAccount.

### spec.serviceAccount.annotations

`map<string, string>`

Annotations on the ServiceAccount the chart creates - the cloud
workload-identity seam (e.g. eks.amazonaws.com/role-arn,
iam.gke.io/gcp-service-account) for the S3, GCS, SQS, Kinesis,
Pub/Sub and BigQuery integrations.

### spec.serviceAccount.existingName

`string`

Run as an existing ServiceAccount instead of creating one (its
annotations are then yours to manage, and `annotations` is ignored).
The module binds the ConfigMap reads to whichever account runs the
relay.

### spec.podAnnotations

`map<string, string>`

Annotations added to the relay pods (e.g. a service mesh or a log
shipper's settings). Values may not contain Helm template syntax.

- rule: Pod annotation values may not contain Helm template syntax ({{).

### spec.podLabels

`map<string, string>`

Labels added to the relay pods only.

### spec.commonLabels

`map<string, string>`

Labels added to every object the chart renders.

### spec.podSecurityContext

`WorkloadPodSecurityContext`

Pod-level security context.

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

Container-level security context for the relay container.

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

Extra plain environment variables for the relay container - e.g.
AWS_REGION for the S3, SQS and Kinesis integrations, which read the
standard AWS SDK environment. Names may not collide with the
variables the module generates for secret values.

### spec.extraEnvFromSecret

`map<string, KubernetesSecretKey>`

Extra environment variables read from existing Secrets in the relay's
namespace, keyed by variable name - e.g. AWS_ACCESS_KEY_ID and
AWS_SECRET_ACCESS_KEY when the cluster has no workload identity.

### spec.extraEnvFromSecret.*.name

`string`

The name of the Kubernetes Secret.

### spec.extraEnvFromSecret.*.key

`string`

The key within the Kubernetes Secret.

### spec.helmValues

`string`

Advanced escape hatch: raw Helm values merged LAST (Helm `-f`
semantics) over everything this spec renders - later keys win. Use it
for chart keys this spec does not type (`extraManifests`). The module
re-pins `fullnameOverride` after the merge. Overriding
`relayproxy.config` here replaces the whole rendered relay
configuration. YAML document as a string.

## Validation Rules

- `spec.extra_env.unique_names`: A variable may be declared in extra_env or extra_env_from_secret, not both.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesGoFeatureFlag, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the relay runs in. |
| `status.outputs.service` | `string` | The relay Service name. |
| `status.outputs.api_endpoint` | `string` | In-cluster evaluation endpoint (e.g. "http://flags.feature-flags.svc.cluster.local:1031") - the base URL for the GO Feature Flag OpenFeature providers (remote and in-process), OFREP clients (`/ofrep/v1/...`) and the REST API. |
| `status.outputs.monitoring_endpoint` | `string` | In-cluster monitoring endpoint serving /health, /info and /metrics (e.g. "http://flags.feature-flags.svc.cluster.local:1032"). |
| `status.outputs.port_forward_command` | `string` | Copy-paste command for reaching the evaluation API from a workstation. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.flagSource.retrievers[].configMap.configMapName` | KubernetesGoFeatureFlagFlagFile | `status.outputs.config_map_name` |
| `spec.flagSource.retrievers[].configMap.key` | KubernetesGoFeatureFlagFlagFile | `status.outputs.key` |
| `spec.flagSource.retrievers[].configMap.namespace` | KubernetesNamespace | `spec.name` |
| `spec.flagSets.items[].source.retrievers[].configMap.configMapName` | KubernetesGoFeatureFlagFlagFile | `status.outputs.config_map_name` |
| `spec.flagSets.items[].source.retrievers[].configMap.key` | KubernetesGoFeatureFlagFlagFile | `status.outputs.key` |
| `spec.flagSets.items[].source.retrievers[].configMap.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)
