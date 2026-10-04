# KubernetesPodMonitor Guide

The judgment this guide carries: a PodMonitor scrapes pods directly, which is right exactly when there is no Service to scrape through, or when a Service would hide the pods you most need to see. It shares every failure mode of a ServiceMonitor (the fence, the selector, the operator skipping the whole object in silence), adds the pod's declared ports as the thing that decides whether a target exists, and gives every replica its own target.

## PodMonitor or ServiceMonitor

Choose a PodMonitor when:

- **no Service names the metrics port**: a database operator's instances (CloudNativePG serves `metrics` on 9187), a DaemonSet's exporters, a sidecar's metrics port. Creating a Service only to scrape it is a resource with no other purpose;
- **every replica must be scraped even while it is not ready**: a Service drops unready pods from its endpoints, so a ServiceMonitor loses sight of a pod exactly when it is failing readiness;
- **the pods behind one Service belong to different workloads** and each should be its own job.

Choose a [KubernetesServiceMonitor](../kubernetesservicemonitor/GUIDE.md) when a Service already exposes the named metrics port and its labels are what the series should carry (`target_labels` exists only there).

## Which Prometheus reads it

Exactly as for a ServiceMonitor: a [KubernetesKubePrometheusStack](../kuberneteskubeprometheusstack/GUIDE.md) on `all_monitors` loads every PodMonitor; one on `release_managed_only` loads only objects labelled `release: <its release name>`. `labels` and `annotations` are the object's own metadata and never reach the series.

## One monitor for a class of components

A PodMonitor with `namespace_selector: {any: true}` and a selector every instance carries reads every instance on the cluster: every CloudNativePG instance (`cnpg.io/podRole: instance`), every Istio gateway (`gateway.networking.k8s.io/gateway-class-name: istio`, the `istio-gateways` preset). That is how a component installed **before** the agent stack is watched: its own composition cannot carry a monitor, because the monitor's CRD arrives with the agent, so the class monitor lives in the agent's composition and reads the instance an environment adds tomorrow with no change. Where an instance's namespace has a NetworkPolicy, it must admit the Prometheus pods on the metrics port, or the target reads down with a dial timeout.

## The selector, the namespace and the port

- **`selector` matches pod labels** (the workload's pod template labels, `kubectl get pods -n <ns> --show-labels`). Match labels the workload controls and keeps stable: `cnpg.io/cluster: <name>` for a CloudNativePG cluster's instances, `app.kubernetes.io/name` for most charts. Never match `pod-template-hash`, which changes on every rollout.
- **Pods are searched in the monitor's own namespace** unless `namespace_selector` says otherwise, and every Secret the monitor reads lives in the monitor's namespace.
- **The port must be declared on the pod.** `port` names a container port (`ports[].name`), and `port_number` gives its number. A pod that does not declare the port yields no target at all. The operator does not scrape undeclared ports, short of rewriting `__address__` with a relabeling. `target_port` is deprecated upstream in favour of the two; a number written in it reaches the object as a number.
- **One target per replica.** A three-instance database is three targets in one job. Dashboards and alerts aggregate across them (`sum by (job)`); `pod_target_labels` copies the role label (primary or replica) onto the series, so a query can tell them apart.
- **`job_label`** names the pod label whose value becomes `job`. Unset, `job` is `<monitor namespace>/<monitor name>`, which is stable but long. Set it to the label naming the workload.

## What makes the operator skip the monitor

The operator validates each monitor when it renders the scrape configuration. When a check fails it skips the whole object, and only its log and the `prometheus_operator_rejected_resources` metric say why. The spec refuses the shapes it can see before the apply:

- two authentication methods on one endpoint, and an authorization of type `Basic`;
- a client certificate without its key;
- the proxy combinations upstream rejects;
- relabeling steps that break their action's rules;
- a `port_number` outside 1-65535, and a `target_port` that is neither a port number nor a port name.

It can't check a missing Secret or key, a `scrape_timeout` longer than the `interval`, or a scrape class the Prometheus doesn't declare. After applying, the Prometheus's `/targets` page lists the job `podMonitor/<namespace>/<name>/<endpoint index>` with each pod's last error.

A PodMonitor endpoint reads its TLS material only from Secrets and ConfigMaps. Unlike a ServiceMonitor, it has no file fields, so a Prometheus that denies filesystem access to monitors accepts every PodMonitor.

## Credentials are references

Every Secret and ConfigMap an endpoint reads is a selector whose `name` defaults to a reference to a KubernetesSecret or KubernetesConfigMap. In an infra chart, reference the Secret resource: the graph shows the dependency and creates the Secret before the monitor. Prefer `authorization` or `oauth2` over `basic_auth`, and over the deprecated `bearer_token_secret`.

## Bound what one target can cost

Each replica is a target, so a scale-out multiplies the series. `sample_limit` fails a scrape that suddenly returns too many samples. `metric_relabelings` with `action: drop` on `__name__` stops series you never read, and an exporter's most verbose families (`pg_stat_statements_*`, `pg_settings_*`) are the usual candidates. A 30s `interval` is enough for nearly every exporter.

## On the diagram

The monitor draws edges to its namespace, to the namespaces it watches (access-style, never nesting), and to every Secret and ConfigMap it reads. It draws no edge to the pods or to the Prometheus, because both are label matches. The registry prerequisite orders the monitor after the stack that installs its CRDs.

## Design rationale

The spec mirrors the upstream PodMonitor field for field under the upstream keys. It shares its TLS, authorization, OAuth2, relabeling and selector types with KubernetesServiceMonitor through `catalog/kubernetes/prometheus_operator_api.proto`. It keeps its own endpoint message: a pod endpoint has `port_number`, no file fields, and a TLS configuration without files, and those differences stay visible in the spec. The envelope, references for credentials, the operator's checks as validation, and the two upstream shapes with no direct proto form (`target_port` as an IntOrString, `params` and `proxy_connect_header` as list-valued maps) are as the [ServiceMonitor guide](../kubernetesservicemonitor/GUIDE.md) describes.

## Parity accounting

- **Pinned:** prometheus-operator v0.94.1 (`monitoring.coreos.com/v1` PodMonitor), as shipped in kube-prometheus-stack chart 91.8.2.
- **Coverage:** all 124 upstream spec leaves are spec fields under their upstream keys (envelope excluded), checked leaf by leaf against the pinned CRD.
- **Mapped:**
  - **64-bit integers.** The limits and `modulus` are 32-bit unsigned fields with upstream's minimum of 0, because protojson writes a 64-bit integer as a string the API server refuses.
  - **`convertClassicHistogramsToNHCB`.** It keeps its exact key through an explicit JSON name.
  - **Secret and ConfigMap names, and `namespace_selector.match_names`.** They are references that resolve to the plain name upstream reads.
  - **`target_port`.** It is one string, written to the object as a number when it is all digits.
  - **`params` and `proxy_connect_header`.** Their `{values: [...]}` wrappers are written as bare lists.
- **Deprecated, modelled:** `bearer_token_secret` (use `authorization`) and `target_port` (use `port` or `port_number`). Each comment names the replacement.
- **Composed:** nothing; the kind applies one object.
- **Excluded:** nothing. The object's `status` is written by the operator, not configured.

## Pairs well with

- [KubernetesKubePrometheusStack](../kuberneteskubeprometheusstack/GUIDE.md): installs the CRDs and the Prometheus that scrapes through the monitor.
- [KubernetesServiceMonitor](../kubernetesservicemonitor/GUIDE.md): scraping through a Service.
- [KubernetesPrometheusRule](../kubernetesprometheusrule/GUIDE.md): the alerts and recording rules over what is scraped.
- [The observability stack pattern](../../_patterns/observability-stack.md): where monitors sit in an agent-and-hub layout.
