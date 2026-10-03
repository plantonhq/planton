# KubernetesServiceMonitor Guide

The judgment this guide carries: a ServiceMonitor that applies cleanly can still scrape nothing. Which Prometheus reads it is decided by labels on the object, which targets it yields is decided by labels on Services, and when one endpoint breaks one of the operator's rules the operator skips the whole object without a word. Get the fence, the selector and the port right, keep credentials as references, and bound what a target may cost.

## ServiceMonitor or PodMonitor

| Situation | Kind | Why |
|---|---|---|
| The workload has a Service that names its metrics port | ServiceMonitor | Targets carry the Service's name; `target_labels` copies the Service's labels onto every series. |
| No Service exposes the metrics port (a database operator's instances, a DaemonSet's exporters, a sidecar) | [KubernetesPodMonitor](../kubernetespodmonitor/GUIDE.md) | Adding a Service only to be scraped is a resource with no other purpose. |
| Every replica must be scraped even while it is not ready (a pod failing readiness still has metrics worth reading) | KubernetesPodMonitor | A Service drops unready pods from its endpoints, so a ServiceMonitor stops seeing the pod exactly when it is in trouble. |
| The Service fronts pods from several workloads | ServiceMonitor with a narrower `selector`, or a PodMonitor per workload | One job then mixes series from unrelated pods. |

## Which Prometheus reads it

A ServiceMonitor is configuration a Prometheus selects; nothing on the object names a Prometheus. A [KubernetesKubePrometheusStack](../kuberneteskubeprometheusstack/GUIDE.md) on its default discovery (`all_monitors`) loads every ServiceMonitor in every namespace, so on a cluster with one stack you set no label.

On a cluster with a second stack (a receiver that only stores remote-written series runs `release_managed_only`), that stack loads only objects labelled `release: <its release name>` (its `release_name` output). Scraping belongs to the agent stack on each cluster, which holds complete data; the label is the exception, and it deserves a comment in the manifest where it is used.

`labels` and `annotations` here are the object's own metadata. They never reach the series. Labels on the series come from the Service (`job_label`, `target_labels`), the pod (`pod_target_labels`), and relabelings.

## The selector, the namespace and the port

- **`selector` matches Service labels, not pod labels.** Read the Service's labels (`kubectl get svc -n <ns> --show-labels`); a selector copied from the Deployment matches nothing when the Service is labelled differently. `{}` matches every Service in the searched namespaces, which is rarely what a component wants.
- **Services are searched in the monitor's own namespace** unless `namespace_selector` says otherwise. Put the monitor beside the workload. `namespace_selector.match_names` is for a monitor owned by a platform team watching several namespaces, and every Secret the monitor reads must still live in the monitor's namespace.
- **`port` names the Service port** (`spec.ports[].name`). Name the metrics port on the Service and use the name: a port name survives a number change. `target_port` is for a Service that exposes the port without naming it, by container port number or name. A number in `target_port` reaches the object as a number, which is what the operator matches against `containerPort`.
- **`job_label`** sets the `job` every dashboard and alert groups by. Unset, `job` is the Service's name, which differs between environments when Service names carry a prefix. Choose a label whose value is the same everywhere, such as `app.kubernetes.io/name`.

## What makes the operator skip the monitor

The operator validates each monitor when it renders the scrape configuration. When a check fails it skips the whole object and records the reason only in its own log and in the `prometheus_operator_rejected_resources` metric. The API server accepted the object and the deploy succeeded, so nothing else says anything is wrong. The spec refuses these shapes before the apply:

- more than one of `authorization`, `basic_auth`, `oauth2` and `bearer_token_secret` on one endpoint, and an `authorization.type` of `Basic`;
- a client certificate without its key (or the reverse), and a CA, certificate or key taken from both a Secret and a file;
- `proxy_connect_header` without a proxy, `proxy_from_environment` beside `proxy_url` or `no_proxy`, and `no_proxy` without `proxy_url`;
- relabeling steps that break their action's rules;
- a `target_port` that is neither a port number nor a port name.

What the spec can't check:

- **A referenced Secret or key that is missing.** The operator skips the monitor unless the selector is `optional`.
- **A `scrape_timeout` longer than the `interval`.** Durations allow units CEL can't compare.
- **A `scrape_class` the Prometheus doesn't declare.**
- **A relabel regex Prometheus can't compile.**

After applying, `kubectl -n <prometheus namespace> port-forward svc/<prometheus> 9090` and open `/targets`. The monitor's job (`serviceMonitor/<namespace>/<name>/<endpoint index>`) shows each target and its last error.

## Credentials are references

Every Secret and ConfigMap an endpoint reads is a selector whose `name` defaults to a reference to a KubernetesSecret or KubernetesConfigMap. In an infra chart, reference the Secret resource rather than typing its name: the graph then shows the dependency and creates the Secret before the monitor. A monitor applied before its Secret exists is skipped by the operator until the next reconcile, which hides a misordering as a flaky scrape.

Prefer `authorization` (a bearer token read from a Secret) or `oauth2` over `basic_auth`, and over the deprecated `bearer_token_secret` and `bearer_token_file`. The file forms (`bearer_token_file`, `tls_config`'s `ca_file`, `cert_file` and `key_file`) read paths inside the Prometheus container. They exist for scraping the Kubernetes API with the pod's own service-account token, and a Prometheus that sets `arbitraryFSAccessThroughSMs.deny` refuses every ServiceMonitor that names one.

## Bound what one target can cost

Every series a target exposes is stored at every scrape. One release that adds a label carrying a request ID multiplies the Prometheus's memory, and with remote write, the bill. Three settings bound the damage:

- **`sample_limit`**: a scrape returning more samples than this fails as a whole (`up` reads 0), which is loud and cheap.
- **`metric_relabelings` with `action: drop`** on `__name__`: stops expensive series you never read before they are stored.
- **`interval`**: 30s is enough for nearly every service; halving it doubles the samples.

## On the diagram

The monitor draws edges to its namespace, to the namespaces it watches (access-style, so it never nests inside them), and to every Secret and ConfigMap it reads. It draws no edge to the Services it scrapes or to the Prometheus that reads it, because both are label matches, not references. A reviewer checks those two matches deliberately. The registry prerequisite orders the monitor after the stack that installs its CRDs.

## Design rationale

The spec mirrors the upstream ServiceMonitor field for field under the upstream keys, so a monitor written for any prometheus-operator reads the same here. The types it shares with KubernetesPodMonitor (TLS, authorization, OAuth2, relabeling, selectors) live once in `catalog/kubernetes/prometheus_operator_api.proto`. The kind adds four things:

- **The envelope.** `namespace` is a reference so composition can draw it. `labels` and `annotations` are routed to the object's metadata, because a monitor's labels are configuration.
- **References for credentials.** Every Secret and ConfigMap selector's `name` is a `StringValueOrRef`. The projection writes the resolved name, so the object is upstream's exactly.
- **The operator's checks as validation.** They are the rules the operator applies when it selects a monitor, moved forward to before the apply.
- **Two upstream shapes with no direct proto form.** `target_port` is an IntOrString: one string field, marked so the projection writes an all-digit value as a number. `params` and `proxy_connect_header` are maps whose values are lists: a manifest writes `{name: {values: [...]}}`, and the object receives the bare list.

## Parity accounting

- **Pinned:** prometheus-operator v0.94.1 (`monitoring.coreos.com/v1` ServiceMonitor), as shipped in kube-prometheus-stack chart 91.8.2.
- **Coverage:** all 129 upstream spec leaves are spec fields under their upstream keys (envelope excluded), checked leaf by leaf against the pinned CRD.
- **Mapped:**
  - **64-bit integers.** The limits and `modulus` are 32-bit unsigned fields with upstream's minimum of 0, because protojson writes a 64-bit integer as a string the API server refuses.
  - **`convertClassicHistogramsToNHCB`.** It keeps its exact key through an explicit JSON name.
  - **Secret and ConfigMap names, and `namespace_selector.match_names`.** They are references that resolve to the plain name upstream reads.
  - **`target_port`.** It is one string, written to the object as a number when it is all digits.
  - **`params` and `proxy_connect_header`.** Their `{values: [...]}` wrappers are written as bare lists.
  - **`endpoints`.** It requires at least one entry: upstream requires the key, and an empty list can't be written.
- **Deprecated, modelled:** `bearer_token_secret` and `bearer_token_file` (use `authorization`). Each comment names the replacement.
- **Composed:** nothing; the kind applies one object.
- **Excluded:** nothing. The object's `status` is written by the operator, not configured.

## Pairs well with

- [KubernetesKubePrometheusStack](../kuberneteskubeprometheusstack/GUIDE.md): installs the CRDs and the Prometheus that scrapes through the monitor.
- [KubernetesPodMonitor](../kubernetespodmonitor/GUIDE.md): scraping pods without a Service.
- [KubernetesPrometheusRule](../kubernetesprometheusrule/GUIDE.md): the alerts and recording rules over what is scraped.
- [The observability stack pattern](../../_patterns/observability-stack.md): where monitors sit in an agent-and-hub layout.
