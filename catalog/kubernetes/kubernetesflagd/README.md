# KubernetesFlagd

Runs flagd -- the OpenFeature project's flag evaluation daemon -- as a central service on Kubernetes: a Deployment and a Service the module owns end to end (flagd publishes no Helm chart; the image is `ghcr.io/open-feature/flagd`). flagd reads flag definitions from one or more sources and serves them over gRPC evaluation, OFREP, and a sync stream that in-process OpenFeature providers subscribe to, with health and Prometheus metrics on a separate management port.

This kind deploys the ENGINE. The flags are data with their own lifecycle: declare them as a **KubernetesFlagdFlagFile** (typed, validated flags rendered into a ConfigMap) and point a `config_map` source at it, or let flagd read flags over HTTP, gRPC, from object storage, or from the OpenFeature Operator's FeatureFlag resources.

Choose **KubernetesGoFeatureFlag** instead for readable YAML flag files read through the Kubernetes API (a flip in about one polling interval instead of a kubelet sync), change notifications, evaluation export, and API-key-isolated flag sets; choose flagd for JSONLogic targeting, a CNCF-governed project, and the in-process sync protocol.

## What Gets Created

| Object | Name | When |
|---|---|---|
| Namespace | `spec.namespace` | only when `createNamespace` is true |
| Secret | `<name>-sources` | always -- holds flagd's SourceConfig array under the data key `sources`, read into `FLAGD_SOURCES` |
| ServiceAccount | `<name>` | unless `serviceAccount.existingName` names an existing account |
| Role + RoleBinding | `<name>-flag-reader` | one pair per namespace a `featureFlag` source reads, granting get/list/watch on `featureflags.core.openfeature.dev` |
| Deployment | `<name>` | always -- the flagd pods |
| Service | `<name>` | always -- ports `evaluation` (8013), `management` (8014), `sync` (8015), `ofrep` (8016) |
| HorizontalPodAutoscaler | `<name>` | only when `hpa.enabled` is true |
| PodDisruptionBudget | `<name>` | only when `pdb.enabled` is true |
| ServiceMonitor | `<name>-metrics` | only when `metrics.serviceMonitorEnabled` is true (requires the Prometheus Operator CRDs) |

## Spec Walkthrough

- **`namespace`** (required) -- a literal or a reference to a KubernetesNamespace. ConfigMap sources and every referenced Secret must live in this same namespace: pod volumes cannot cross namespaces.
- **`image`** -- repository `ghcr.io/open-feature/flagd`, tag `v0.17.0`, `fips` appends `-fips` to the tag, `pullPolicy`, and `pullSecretNames` for a private mirror.
- **`replicas`**, **`resources`**, **`hpa`**, **`pdb`** -- flagd is stateless; every replica reads the same sources. The default resources are 50m CPU / 64Mi requests and 500m / 256Mi limits.
- **`sources`** (at least one) -- a list where each entry is exactly one of:
  - `configMap` -- a ConfigMap (typically a KubernetesFlagdFlagFile's) mounted as a DIRECTORY at `/etc/flagd/sources/<index>`; `key` (a literal, or a reference to the flag file's `status.outputs.key`) must end in `.json`, `.yaml` or `.yml`; `watcher` picks `fsnotify` or `fileinfo`.
  - `http` -- an http(s) URL polled every `intervalSeconds` (default 5), with an optional request `timeoutSeconds`, plain or credential `headers`, and one way to authenticate: an `authHeader`, or `oauth` client credentials (`clientId`, a sensitive `clientSecret`, `tokenUrl`) that flagd exchanges for a token.
  - `grpc` -- a gRPC sync server, optionally over TLS with a CA read from a Secret, with a `selector` (`flagSetId=<id>` or `source=<name>`) the server filters the synced flags by.
  - `featureFlag` -- an OpenFeature Operator FeatureFlag resource.
  - `googleStorage`, `azureBlob`, `s3` -- an object polled from a bucket or container.

  When several sources define the same flag key, the source listed later wins.
- **`server`** -- the four ports, `tlsSecretName` (a kubernetes.io/tls Secret served on the evaluation and sync listeners; the OFREP and management listeners stay plain HTTP) and `serviceType`.
- **`evaluation`** -- static `contextValues`, `contextFromHeader` (request header to context key) and browser `corsOrigins` -- an empty list allows ANY origin, so list the origins to restrict browser access.
- **`ofrepSse`**, **`sync`**, **`log`**, **`telemetry`**, **`limits`** -- the remaining `flagd start` settings: the OFREP change stream, the sync service's HTTP document and metadata, log format and debug, the Prometheus or OpenTelemetry metrics exporter with collector TLS, and request size limits.
- **`scheduling`**, **`serviceAccount`**, **`podAnnotations`**, **`podLabels`**, **`podSecurityContext`**, **`containerSecurityContext`**, **`extraEnv`**, **`extraEnvFromSecret`**, **`metrics`** -- the pod's placement, identity, security and environment, and the optional ServiceMonitor. `FLAGD_*` variables are refused in the extra environment (the typed fields own them), and a variable may appear in only one of the two maps.

## Outputs

| Output | Description |
|---|---|
| `namespace` | Namespace flagd runs in |
| `service` | The flagd Service name |
| `evaluation_endpoint` | In-cluster gRPC evaluation endpoint `host:8013` for remote flagd providers |
| `sync_endpoint` | In-cluster sync endpoint `host:8015` for in-process flagd providers |
| `ofrep_endpoint` | In-cluster OFREP base URL `http://host:8016` |
| `management_endpoint` | In-cluster `/healthz`, `/readyz` and `/metrics` endpoint `http://host:8014` |
| `port_forward_command` | `kubectl port-forward` command for the OFREP port |

## How the Module Works

- **Sources live in a Secret.** HTTP authorization headers and credential headers are part of flagd's SourceConfig array, so the whole array renders into `<name>-sources` and reaches flagd as `FLAGD_SOURCES`. Every other setting is a `flagd start` argument. A SHA-256 of the sources document is stamped on the pod template (`checksum/sources`), so changing a source rolls the pods.
- **ConfigMaps mount as directories, never `subPath`.** A `subPath` mount never sees an edit; a directory mount follows the kubelet's volume sync, so a flag flip reaches flagd in one to two minutes with no restart.
- **Readiness is honest.** The readiness probe is `/readyz` on the management port, which flagd answers only after every source has synced once. A source that cannot be read keeps the pods unready and the apply waiting.
- **Both engines render the same thing.** The Pulumi module and the OpenTofu module produce identical `flagd start` arguments and a byte-identical sources document.

## Official Documentation

- flagd: https://flagd.dev
- Sync configuration: https://flagd.dev/reference/sync-configuration/
- Flag definitions: https://flagd.dev/reference/flag-definitions/
- OpenFeature Operator: https://github.com/open-feature/open-feature-operator

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
