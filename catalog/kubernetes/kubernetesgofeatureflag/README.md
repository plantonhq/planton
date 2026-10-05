# Kubernetes GO Feature Flag

## When NOT to Use This

**One resource is ONE GO Feature Flag relay proxy** -- the OpenFeature-native feature flag server, installed from the official `relay-proxy` chart (chart 1.56.x = relay v1.56) at `charts.gofeatureflag.org`. The relay reads flag files from retrievers, evaluates flags for any OpenFeature SDK, announces flag changes and exports evaluation events.

Not the right kind when:

- **You want the flags themselves** -- flags are data with their own lifecycle. Declare them as a `KubernetesGoFeatureFlagFlagFile` (typed, validated, rendered into a ConfigMap) and point a `configMap` retriever at it. This kind deploys the ENGINE.
- **You prefer the CNCF OpenFeature reference daemon** -- `KubernetesFlagd` runs flagd with JSONLogic targeting and a gRPC sync stream; choose it when CNCF governance and flagd's flag format matter more than GO Feature Flag's readable query rules, notifiers and exporters.
- **You want a hosted flag service** -- point your SDKs at it and deploy nothing here.

## Flags come from retrievers

`flagSource.retrievers` lists where the flag files live: a ConfigMap (read through the Kubernetes API on every poll -- the way to serve a `KubernetesGoFeatureFlagFlagFile`), HTTP, GitHub, GitLab, Bitbucket, S3, Google Cloud Storage, Azure Blob Storage, MongoDB, Redis or PostgreSQL. When several are listed the relay merges them and a later retriever wins on a flag both define. `pollingIntervalMs` (relay default 60000) is the flip latency: an edited flag file is served within one interval, with no restart.

A `configMap` retriever's `key` takes a literal or a reference to the flag file's `status.outputs.key`. For every `configMap` retriever the module creates `<name>-flag-reader`, a Role and RoleBinding in the ConfigMap's namespace granting the relay's ServiceAccount `get` on exactly the named ConfigMaps (the chart ships no RBAC). `startWithRetrieverError: true` lets the relay start before a flag file exists, so the relay and its flag files deploy in any order.

## Two modes

`flagSource` serves one flag source to every caller. `flagSets` serves several isolated flag sets, each selected by the API key the caller presents, each with its own retrievers, notifiers and exporters. The relay accepts exactly one mode, and the spec makes it a required choice. Flag-set names are unique and never `default` (the relay reserves it), checked at validation; flag-set keys must be unique across sets, which the module checks at deploy time.

## Secrets never reach the configuration

The relay configuration renders into the chart's ConfigMap and carries no secret. Every token, password, webhook URL, credential header and API key is written into the module-owned `<name>-env` Secret and reaches the relay as an environment variable -- the relay reads list entries from indexed variables (`RETRIEVERS_<i>_TOKEN`, `FLAGSETS_<i>_RETRIEVERS_<j>_TOKEN`) and key lists from comma-separated ones (`AUTHORIZEDKEYS_EVALUATION`). Two consequences: a key may not contain a comma, and a sensitive header name may not contain `_`; the module checks both at deploy time. Every variable carries `runtime.envVariablePrefix` (default `GOFFRELAY_`), which keeps the relay from reading the service-link variables Kubernetes injects for every Service in the namespace.

## Authentication

Without `authorizedKeys` (or flag sets, which always carry keys) the API is OPEN: anyone who can reach the Service evaluates every flag and reads the full flag configuration. `authorizedKeys.evaluation` keys are what SDKs and in-process providers present; `authorizedKeys.admin` keys guard the admin endpoints.

## Notifiers and exporters

`notifiers` (Slack, Microsoft Teams, Discord, webhook) are told about every flag created, changed or removed. `exporters` (webhook, log, S3, Google Cloud Storage, Azure Blob Storage, SQS, Kinesis, Pub/Sub, BigQuery, Kafka, OpenTelemetry) receive evaluation events in batches. Cloud destinations take credentials from workload identity (`serviceAccount.annotations`) or from `extraEnv` / `extraEnvFromSecret`.

## What the chart cannot carry

The chart mounts no volume besides its configuration, so the persistent flag configuration file, the `file` retriever, the `file` exporter and the HTTP retriever's client-certificate paths are not modeled; a Kubernetes Service reaches only an all-interfaces HTTP listener, so the module pins `server.mode: http`. The chart always runs the configuration through Helm's `tpl`, so the module escapes every `{{` (the exporters' filename, CSV and log templates) and it reaches the relay unchanged. Exposure composes from Gateway API kinds against the `service` output. `helmValues` merges last (Helm `-f` semantics) for chart keys the spec does not type, and `fullnameOverride` is re-pinned after the merge.

## Traces

`telemetry` sends the relay's traces to an OTLP collector (`otlpEndpoint`, `otlpProtocol` `grpc` or `http/protobuf`). `tracesSampler` takes an OpenTelemetry sampler name (`always_on`, `always_off`, `traceidratio`, `parentbased_always_on`, `parentbased_always_off`, `parentbased_traceidratio`) or `jaeger_remote`; `tracesSamplerArg` (a fraction from 0 to 1) applies only to the two ratio samplers, and `jaegerSampler` only to `jaeger_remote`, with `managerHostPort` an http(s) URL of the sampling manager.

## Outputs

| Output | Meaning |
|---|---|
| `namespace` | Namespace the relay runs in |
| `service` | The relay Service (`metadata.name`) |
| `api_endpoint` | Evaluation endpoint -- the base URL for the GO Feature Flag OpenFeature providers, OFREP clients and the REST API |
| `monitoring_endpoint` | `/health`, `/info` and `/metrics` |
| `port_forward_command` | Workstation access to the evaluation API |

## Name budget

`metadata.name` is at most 63 characters: the chart names the relay Service exactly after the resource, and a Service name is a 63-character DNS label. Both engines refuse a longer name before creating anything.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
