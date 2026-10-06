# GO Feature Flag

Deploys a GO Feature Flag relay proxy -- an OpenFeature-native feature flag server -- from the official `relay-proxy` chart at `charts.gofeatureflag.org`. The relay reads flag files from one or more retrievers (a ConfigMap, Git hosting, HTTP, object storage or a database), evaluates flags for any OpenFeature SDK over REST, OFREP and the flag-configuration endpoint in-process providers sync from, notifies on every flag change and exports evaluation events.

Know the grain before you deploy: this component deploys the ENGINE, never the FLAGS. Declare flags as a GO Feature Flag File and point a `configMap` retriever at it, so a flag flip edits only the flag file and the relay serves it on its next poll with no restart. And know the default posture: without `authorizedKeys` the API is OPEN -- anyone who can reach the Service evaluates every flag and reads the full flag configuration.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Helm Release** -- the `relay-proxy` chart, creating:
  - Deployment with the configured replicas (default 1), probed on `/health` at the monitoring port
  - Kubernetes Service on the evaluation port (default 1031) and the monitoring port (default 1032)
  - ConfigMap holding the relay configuration -- secret-free; the module renders it as a `server.monitoringPort` line and one JSON document
  - ServiceAccount carrying your workload-identity annotations -- skipped when `serviceAccount.existingName` names your own
  - PodDisruptionBudget -- only when `pdb.enabled`; the module adds the pod label the chart's budget selects, so it protects the relay
  - HorizontalPodAutoscaler -- only when `hpa.enabled`
- **Env Secret** (`<name>-env`) -- created only when the spec holds a secret value; every key, token, password, webhook URL and credential header reaches the relay from here as an environment variable, never through the configuration
- **Flag-reader Role and RoleBinding** (`<name>-flag-reader`) -- one pair per namespace a `configMap` retriever reads, granting the relay `get` on exactly the named ConfigMaps (the chart ships no RBAC)
- **ServiceMonitor** (`<name>-metrics`) -- created only when `metrics.serviceMonitorEnabled` is `true`; requires the Prometheus Operator CRDs (the apply FAILS without them)
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

Deliberately absent: no ingress -- exposure composes from Gateway API kinds -- and no volume: the chart mounts only its configuration, so the persistent flag configuration file, the `file` retriever and the `file` exporter are not offered.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A flag source** -- a GO Feature Flag File (or any ConfigMap holding a GO Feature Flag flag file), or a repository, bucket, URL or database the relay can reach. With `startWithRetrieverError: true` the source may arrive after the relay.
- **Permission to grant ConfigMap reads** -- the runner must itself hold `get` on configmaps in every namespace a `configMap` retriever names; Kubernetes refuses to create a Role granting more than its creator holds.
- **Cloud identity** (only for S3, Google Cloud Storage, Azure Blob Storage, SQS, Kinesis, Pub/Sub or BigQuery) -- workload identity through `serviceAccount.annotations`, or credentials through `extraEnv` / `extraEnvFromSecret`.
- **Prometheus Operator CRDs** (only when enabling the ServiceMonitor).

## Deploy

### Console

Open the deployment store, find **GO Feature Flag**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Flag File Relay** preset in the [Presets](#presets) tab for the production shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGoFeatureFlag
metadata:
  name: flags
  org: acme-corp
  env: prod
spec:
  namespace:
    value: feature-flags
  createNamespace: false
  replicas: 2
  pdb:
    enabled: true
  authorizedKeys:
    evaluation:
      - $secret/flags-evaluation-key
  flagSource:
    retrievers:
      - configMap:
          configMapName:
            value: release-flags
          key:
            value: flags.goff.yaml
    startWithRetrieverError: true
    pollingIntervalMs: 30000
```

```shell
planton apply -f flags.yaml
```

This deploys two relay replicas behind a disruption budget, reading the `release-flags` ConfigMap every 30 seconds through a Role scoped to that ConfigMap, with evaluation guarded by a key delivered through the module-owned env Secret. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the relay to a flag file managed alongside it:

```yaml
spec:
  namespace:
    value: feature-flags
  flagSource:
    retrievers:
      - configMap:
          configMapName:
            valueFrom:
              kind: KubernetesGoFeatureFlagFlagFile
              name: release-flags
              fieldPath: status.outputs.config_map_name
          key:
            valueFrom:
              kind: KubernetesGoFeatureFlagFlagFile
              name: release-flags
              fieldPath: status.outputs.key
    startWithRetrieverError: true
```

The InfraPipeline resolves the flag file's ConfigMap name and data key and provisions the relay against it; with `startWithRetrieverError` the two can deploy in either order.

## Key Configuration

These are the most important decisions when configuring GO Feature Flag. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One flag source or flag sets** -- `flagSource` and `flagSets` are one required choice. `flagSource` serves the same flags to every caller. `flagSets` gives each team or client its own flags behind its own API keys -- a key selects exactly one set, sets share nothing, set names are unique and never `default`, and a key shared between sets fails the deploy. The relay ignores top-level sources beside flag sets, which is why the spec makes the choice exclusive.

**Where the flags live** -- A `configMap` retriever pointed at a GO Feature Flag File keeps flags typed, validated at plan time and separate from the relay's lifecycle. Git retrievers make every flag change a reviewed commit; set a `token` to lift the provider's anonymous rate limit, which a short polling interval otherwise exhausts. Later retrievers win on a flag several define.

**Flip latency** -- `pollingIntervalMs` (relay default 60000, minimum 1000) is how long a flag change takes to reach evaluations. Shorter means faster flips and more retriever reads per replica; `enablePollingJitter` spreads replicas' reads apart. A negative value disables polling -- flags then load once at startup.

**Authentication** -- Without `authorizedKeys` the API is open. `evaluation` keys are what SDKs present; `admin` keys guard the admin endpoints. A key may not contain a comma -- the relay reads key lists from comma-separated environment variables, and the module refuses one at deploy time.

**Notifications and export** -- `notifiers` post every flag change to Slack, Teams, Discord or a signed webhook; `exporters` batch evaluation events to a webhook, the log, object storage, queues and streams, BigQuery or OpenTelemetry. Every credential is a managed secret delivered as an environment variable. The log, filename and CSV templates use Go template syntax; the chart always runs the configuration through Helm's `tpl`, and the module escapes every `{{` so the templates reach the relay unchanged.

**Traces** -- `telemetry.tracesSampler` picks an OpenTelemetry sampler or `jaeger_remote`. `tracesSamplerArg` sets the sampled fraction for `traceidratio` and `parentbased_traceidratio` only, and `jaegerSampler` (an http(s) `managerHostPort`) is read only with `jaeger_remote`.

**Environment prefix** -- `runtime.envVariablePrefix` (default `GOFFRELAY_`) scopes the variables the relay reads as configuration. Without one, a Service named `server` in the namespace would inject `SERVER_PORT` and silently move the relay's listener. Keep the default unless you have a reason.

**The escape hatch** -- `helmValues` merges LAST over everything the typed fields render (Helm `-f` semantics) for chart keys the spec does not type, such as `extraManifests`. Overriding `relayproxy.config` replaces the whole rendered configuration and its secret wiring. The module re-pins `fullnameOverride` after the merge so the Service name stays stable.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesGoFeatureFlagFlagFile** (optional) | `flagSource.retrievers[].configMap.configMapName` | `status.outputs.config_map_name` |
| **KubernetesGoFeatureFlagFlagFile** (optional) | `flagSource.retrievers[].configMap.key` | `status.outputs.key` |
| **KubernetesNamespace** (optional) | `flagSource.retrievers[].configMap.namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `api_endpoint` | In-cluster evaluation endpoint (`http://<service>.<namespace>.svc.cluster.local:1031`) | The base URL for GO Feature Flag OpenFeature providers (remote and in-process), OFREP clients and the REST API in your services' configuration |
| `monitoring_endpoint` | In-cluster `/health`, `/info` and `/metrics` | Health checks and scrape targets outside the ServiceMonitor |
| `service` | The relay Service name | Gateway API routes that expose the relay |
| `namespace` | Namespace the relay runs in | Locating the install for diagnostics |
| `port_forward_command` | Ready-to-run `kubectl port-forward` command | Workstation access to the evaluation API |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Release switches from a flag file** -- Two replicas, a disruption budget, evaluation keys, and one flag file per environment that an operator or an agent edits to turn a feature on for one more organization. Start from the **Flag File Relay** preset.

**Flags as reviewed code** -- The relay reads its flag file from a GitHub repository and posts every change to Slack, so the flag history is the repository history. Start from the **Flags From GitHub** preset.

**One relay, many teams** -- Flag sets per team or client, each reading its own flag file behind its own key. Start from the **Team Flag Sets** preset.

**Evaluation analytics** -- Add an exporter to S3, BigQuery or Kafka to see which variation every caller received, for experiment analysis and rollout monitoring.

## Works With

- [**GO Feature Flag File**](/infra-catalog/kubernetes-go-feature-flag-flag-file) -- the typed flags this relay reads through its ConfigMap retriever
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the relay install
- [**Kubernetes ConfigMap**](/infra-catalog/kubernetes-config-map) -- any ConfigMap holding a GO Feature Flag flag file works as a source
- [**flagd**](/infra-catalog/kubernetes-flagd) -- the CNCF OpenFeature alternative engine, for teams that prefer flagd's flag format and gRPC sync
- [**kube-prometheus-stack**](/infra-catalog/kubernetes-kube-prometheus-stack) -- provides the Prometheus Operator CRDs the ServiceMonitor needs
- [**OpenTelemetry Collector**](/infra-catalog/kubernetes-otel-collector) -- the OTLP target for the relay's traces
