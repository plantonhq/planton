# flagd

Deploys flagd, the OpenFeature project's flag evaluation daemon, as a central service on Kubernetes that serves feature flags over gRPC evaluation, OFREP, and a sync stream for in-process OpenFeature providers. The module owns every object -- flagd publishes no Helm chart -- and reads flags from typed flag files, HTTP, gRPC, object storage, or OpenFeature Operator FeatureFlag resources. Flags are data with their own lifecycle: a flag flip edits a flag file and never re-applies the daemon, and flagd reports ready only after every source has synced.

## What Gets Created

- **flagd Deployment** -- the daemon pods, with every setting passed as a `flagd start` argument and the source list read from `FLAGD_SOURCES`. ConfigMap sources mount as directories, never `subPath`, so a flag edit reaches flagd when the kubelet syncs the volume. Liveness is `/healthz` and readiness `/readyz` on the management port.
- **Service** -- named after the resource, exposing `evaluation` (8013), `management` (8014), `sync` (8015) and `ofrep` (8016).
- **Sources Secret** -- `<name>-sources`, holding the source list. HTTP authorization headers and credential headers live only here, never in the Deployment. A checksum of the document on the pod template rolls the pods when a source changes.
- **ServiceAccount** -- named after the resource, created unless `serviceAccount.existingName` names an existing account. Its annotations are the cloud workload-identity seam for bucket sources.
- **FeatureFlag reader Role and RoleBinding** -- `<name>-flag-reader`, created only for `featureFlag` sources, one pair per namespace they read, granting get/list/watch on `featureflags.core.openfeature.dev`.
- **Namespace** -- created only when `createNamespace` is true.
- **HorizontalPodAutoscaler** -- created only when `hpa.enabled` is true.
- **PodDisruptionBudget** -- created only when `pdb.enabled` is true.
- **ServiceMonitor** -- `<name>-metrics`, created only when `metrics.serviceMonitorEnabled` is true; scrapes `/metrics` on the management port.

## Before You Deploy

### Planton Setup

A Kubernetes provider connection in the Connect module targeting the cluster. The flag files flagd serves are separate KubernetesFlagdFlagFile resources, usually deployed in the same InfraChart.

### Kubernetes Cluster

- Every ConfigMap source and every Secret the spec references (`server.tlsSecretName`, gRPC and OpenTelemetry CA and client certificates, `extraEnvFromSecret`) must exist in flagd's namespace: pod volumes and secretKeyRefs cannot cross namespaces.
- The OpenFeature Operator's CRDs (only for `featureFlag` sources).
- The Prometheus Operator CRDs (only for `metrics.serviceMonitorEnabled`).
- Cloud credentials for bucket sources (only for `googleStorage`, `azureBlob`, `s3`): workload identity through `serviceAccount.annotations`, or variables through `extraEnv` / `extraEnvFromSecret`.

## Deploy

### Console

Open the deployment store, find **flagd**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Flag File Source** preset in the [Presets](#presets) tab for the standard shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesFlagd
metadata:
  name: flagd
  org: acme-corp
  env: prod
spec:
  namespace:
    value: feature-flags
  createNamespace: false
  replicas: 2
  pdb:
    enabled: true
  sources:
    - configMap:
        configMapName:
          value: release-flags
        key:
          value: flags.flagd.json
  log:
    format: json
```

```shell
planton apply -f flagd.yaml
```

This deploys two flagd replicas behind a disruption budget, serving the flags in the `release-flags` ConfigMap's `flags.flagd.json` key from a directory mount. The pods turn ready once that ConfigMap has been read. An Infra Job tracks the provisioning in real time.

### InfraChart

When the flag file is deployed alongside flagd, wire the ConfigMap by reference:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: feature-flags
      fieldPath: spec.name
  sources:
    - configMap:
        configMapName:
          valueFrom:
            kind: KubernetesFlagdFlagFile
            name: release-flags
            fieldPath: status.outputs.config_map_name
        key:
          valueFrom:
            kind: KubernetesFlagdFlagFile
            name: release-flags
            fieldPath: status.outputs.key
```

The InfraPipeline renders the flag file's ConfigMap first, so flagd's first sync finds it and its pods turn ready on the first rollout.

## Key Configuration

These are the most important decisions when configuring flagd. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Flag file or remote source.** A `configMap` source backed by a KubernetesFlagdFlagFile keeps flags as reviewed manifests next to the services that read them, and a flip lands once the kubelet syncs the volume -- typically one to two minutes. An `http`, `grpc` or bucket source suits flags published by another system; it polls every `intervalSeconds` (default 5), which is faster but makes that system part of flagd's availability.

**A source that cannot be read blocks the rollout.** flagd answers `/readyz` only after every source has synced once, so a missing ConfigMap or an unreachable URL keeps the pods unready and the apply waiting. A missing ConfigMap goes further: the volume never mounts and the pods never start. Deploy flag files before or alongside flagd.

**Source order decides duplicates.** When two sources define the same flag key, the source listed later wins. Put the authoritative source last. A `grpc` source's `selector` (`flagSetId=<id>`) takes only one flag set from a shared sync server.

**Where credentials go.** `http.authHeader`, `http.oauth.clientSecret` and the `sensitiveHeaders` maps are sensitive fields: they render only into the `<name>-sources` Secret. An HTTP source authenticates one way -- `authHeader` or `oauth`, never both. A header declared both plain and sensitive fails the deploy. Bucket sources authenticate through workload identity on the ServiceAccount, or through `extraEnvFromSecret`; `FLAGD_*` variables are refused there because the typed fields own them.

**Remote or in-process evaluation.** Services can evaluate remotely against `evaluation_endpoint` (gRPC) or `ofrep_endpoint`, paying a network hop per evaluation, or sync the whole flag set in-process from `sync_endpoint` and evaluate locally. `sync.httpEnabled` also serves the full document at `/v1/flags` on the sync port -- anyone who reaches that port reads every flag.

**TLS is opt-in, and covers evaluation and sync.** The listeners speak plaintext inside the cluster unless `server.tlsSecretName` names a kubernetes.io/tls Secret (a cert-manager Certificate's, for example); TLS then covers the evaluation and sync listeners, while the OFREP and management listeners stay plain HTTP -- terminate TLS for OFREP at a gateway. Exposure outside the cluster composes through Gateway API routes against the `service` output; the Service type stays ClusterIP by default.

**Browser origins.** An empty `evaluation.corsOrigins` allows ANY origin -- flagd always installs its CORS handler. List the origins whenever browsers reach flagd.

**Availability.** flagd is stateless, so `replicas` and `hpa` add throughput and resilience directly. Pair two or more replicas with `pdb`, and spread them with `scheduling.topologySpreadConstraints` -- an empty `matchLabels` spreads flagd's own pods.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesFlagdFlagFile** (optional) | `sources.configMap.configMapName` | `status.outputs.config_map_name` |
| **KubernetesFlagdFlagFile** (optional) | `sources.configMap.key` | `status.outputs.key` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `evaluation_endpoint` | In-cluster gRPC evaluation endpoint `host:8013` | The flagd OpenFeature provider in RPC mode |
| `sync_endpoint` | In-cluster sync endpoint `host:8015` | The flagd OpenFeature provider in in-process mode |
| `ofrep_endpoint` | In-cluster OFREP base URL on port 8016 | OFREP providers in any language, and Gateway API routes for browser clients |
| `management_endpoint` | In-cluster `/healthz`, `/readyz` and `/metrics` URL on port 8014 | Monitoring and probes |
| `service` | The flagd Service name | Gateway API backend references |
| `port_forward_command` | Ready-to-run `kubectl port-forward` command for the OFREP port | Workstation access |

## Common Patterns

**Flags declared in Planton.** One KubernetesFlagdFlagFile per owning team, each mounted as a `configMap` source, and two flagd replicas behind a disruption budget. Teams flip their own flags without touching the daemon. Start from the **Flag File Source** preset.

**Flags published elsewhere.** An `http` source polling a protected document, with the Authorization header held as a managed secret and an `intervalSeed` so this flagd does not poll in lockstep with other flagd deployments reading the same document. Start from the **HTTP Source** preset.

**Heavy evaluation traffic.** Three to ten replicas autoscaled on CPU, spread across zones, with a disruption budget keeping two pods serving. Start from the **Autoscaled Zone Spread** preset.

**Operator-managed flags.** A `featureFlag` source reads an OpenFeature Operator FeatureFlag resource, and the module grants flagd watch on FeatureFlags in that namespace. Use it when the operator already manages flags for sidecar-injected workloads and a central flagd should serve the same definitions.

## Works With

- [**flagd Flag File**](/infra-catalog/kubernetes-flagd-flag-file) -- the typed flag definitions flagd serves through a `configMap` source
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the namespace flagd and its flag files share
- [**GO Feature Flag**](/infra-catalog/kubernetes-go-feature-flag) -- the alternative engine: choose it for readable YAML flag files read through the Kubernetes API (a flip in about one polling interval instead of a kubelet sync), change notifications, evaluation export, and API-key-isolated flag sets; choose flagd for JSONLogic targeting, a CNCF-governed project, and the in-process sync protocol
- [**GO Feature Flag File**](/infra-catalog/kubernetes-go-feature-flag-flag-file) -- the flag file format of the alternative engine
