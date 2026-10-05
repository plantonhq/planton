# KubernetesGoFeatureFlag Guide

The judgment this guide carries: the relay is the engine and the flags are
data -- compose them as two resources, keep secrets out of the relay's
configuration, and know which upstream features the chart cannot carry
before an architecture depends on them.

Substitutes for: hosted feature flag services such as LaunchDarkly or
Unleash, for teams that want flags evaluated in-cluster through OpenFeature
and kept as reviewable code.

Pinned: the `relay-proxy` chart 1.56.0, which installs GO Feature Flag
relay v1.56.0 (the chart version and the relay version move together).
Coverage: every relay setting at that version is a typed field except
these, left out by design -- the chart's ingress block (exposure composes
from Gateway API kinds); the `lambda` and `unixsocket` server modes and the
listen host (a Service reaches only an all-interfaces HTTP listener); the
persistent flag configuration file, the `file` retriever, the `file`
exporter and HTTP-retriever client certificates (each needs a mounted
file, and the chart mounts none); the flag-set `environment` key (parsed,
never read); and the singular `retriever` / `exporter` keys,
`openTelemetryOtlpEndpoint` and the deprecated keys, which the lists and
current keys cover.

## When to use it (and when not)

Choose this kind over KubernetesFlagd when you want readable flag files
with query-language targeting (`org in ["acme"]`), change notifications to
Slack, Teams, Discord or a webhook, evaluation export to object storage,
streams or BigQuery, and team flag sets behind their own keys. Choose
KubernetesFlagd when CNCF governance matters more, when JSONLogic
targeting or flagd's gRPC sync stream fits your providers better, or when
you need TLS on the evaluation listener -- this relay serves plain HTTP
inside the cluster.

The engine reads flags; it never owns them. Compose a
KubernetesGoFeatureFlagFlagFile per owner and point a `configMap`
retriever at it. A flag flip then edits only the flag file -- the relay is
never re-applied, flip rights can be granted without rights over the relay,
and the change is served within `pollingIntervalMs` (relay default 60
seconds) with no restart. Inline flags inside the relay spec do not exist
by design.

## Conventions and gotchas

- **Exactly one mode.** `flagSource` and `flagSets` are a required oneof
  because the relay silently ignores top-level retrievers, notifiers and
  exporters whenever flag sets exist. In flag-sets mode every set needs
  at least one API key and set names are unique and never `default` (the
  relay reserves it). A key shared between sets fails the deploy -- the
  module checks it, and the relay refuses to start with one.
- **Secrets travel as environment variables.** The relay's configuration
  ConfigMap is secret-free. Every sensitive field is written into the
  module-owned `<name>-env` Secret and read through indexed variables
  (`RETRIEVERS_<i>_TOKEN`, `FLAGSETS_<i>_NOTIFIERS_<j>_SECRET`,
  `AUTHORIZEDKEYS_EVALUATION`). The relay splits variable names on `_` and
  key lists on `,`, so a sensitive header name may not contain `_` and a
  key may not contain `,` -- both refused at deploy time, not at
  validation, because secret values are only known then.
- **Keep the environment prefix.** `runtime.envVariablePrefix` defaults to
  `GOFFRELAY_`. Without a prefix the relay reads every variable in its pod
  as configuration, including the service-link variables Kubernetes injects
  for each Service in the namespace -- a Service named `server` injects
  `SERVER_PORT=tcp://...` and the relay fails to bind.
- **Install order never matters** with `startWithRetrieverError: true`: the
  relay starts serving and loads flags on the first successful poll. Leave
  it false only when a relay without flags must not start.
- **The ConfigMap grant follows the retriever.** For every `configMap`
  retriever the module creates `<name>-flag-reader` in the ConfigMap's
  namespace, limited by `resourceNames` to the named ConfigMaps. The
  runner must itself hold `get` on configmaps there -- Kubernetes refuses
  to create a Role granting more than its creator holds. Two relays with
  the same name in different namespaces that read ConfigMaps from one
  shared namespace would both claim `<name>-flag-reader` there: give them
  distinct names.
- **A changed secret rolls the pods.** Every secret value reaches the relay
  as an environment variable, read once at start; the pod template carries
  a checksum of the env Secret, so rotating a key or token restarts the
  relay with the new value.
- **The disruption budget selects the relay.** The chart's
  PodDisruptionBudget selects a `name` label the chart's pods do not carry
  on their own; the module sets `name: <fullname>` through `podLabels`, so
  `pdb.enabled` protects the relay.
- **Name budget: 63 characters.** The chart names the relay Service
  exactly after the resource, and a Service name is a 63-character DNS
  label; both engines refuse a longer `metadata.name` before creating
  anything.
- **What the chart cannot carry.** The chart mounts no volume besides its
  configuration and runs one HTTP listener, so the persistent flag
  configuration file, the `file` retriever and exporter, HTTP-retriever
  client certificates, and the `lambda` and `unixsocket` server modes are
  not modeled. Need one of these? It is the KubernetesFlagd decision, or a
  module-owned workload.
- **Templates survive.** The chart always runs the configuration through
  Helm's `tpl`. Exporter filename, CSV and log templates use Go template
  syntax; the module escapes every `{{` so they reach the relay intact.
- **Samplers and their settings pair up.** `telemetry.tracesSamplerArg`
  applies only to `traceidratio` and `parentbased_traceidratio`, and
  `telemetry.jaegerSampler` only to `jaeger_remote`, its
  `managerHostPort` an http(s) URL; validation refuses any other pairing.

## On the diagram

The relay renders as one node with reference edges from each flag file it
reads -- the flags are visible, separately owned nodes. A relay reading
from a repository or a bucket shows no flag node at all; that is the trade
for keeping flags outside the cluster.

## Pairs well with

- KubernetesGoFeatureFlagFlagFile -- the typed flags; wire its
  `config_map_name` into `flagSource.retrievers[].configMap.configMapName`
  and its `key` into `flagSource.retrievers[].configMap.key`.
- KubernetesNamespace -- the relay's namespace, and optionally a separate
  namespace for flag files owned by another team.
- KubernetesKubePrometheusStack -- the Prometheus Operator CRDs the
  ServiceMonitor needs.
- Gateway API route kinds -- when clients outside the cluster evaluate
  through OFREP; point the route at the `service` output and set
  `ofrepEventStream.baseUrl` to the public URL.

Presets (Flag File Relay, Flags From GitHub, Team Flag Sets) ship in the
release's `presets.zip` and in the repository. The full field reference is
[reference.md](v1alpha1/reference.md).
