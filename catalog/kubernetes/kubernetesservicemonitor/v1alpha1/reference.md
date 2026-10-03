# KubernetesServiceMonitor

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

KubernetesServiceMonitorSpec defines a prometheus-operator ServiceMonitor: a
namespaced object that tells every Prometheus selecting it which Services'
endpoints to scrape, and how. The operator turns it into scrape
configuration: each Service matching `selector` in the namespaces
`namespace_selector` names contributes one target per ready endpoint
address, scraped on each of `endpoints`.

100% fidelity with the upstream ServiceMonitor custom resource
(monitoring.coreos.com/v1), pinned to prometheus-operator v0.94.1 -- the
operator the kube-prometheus-stack chart 91.8.2 ships, which is what
KubernetesKubePrometheusStack installs. The upstream spec follows the
Planton envelope directly; there is no nested `service_monitor` sub-message.
Types shared with KubernetesPodMonitor live in
catalog/kubernetes/prometheus_operator_api.proto.

The envelope: `namespace` places the object, and `labels` and `annotations`
are the object's own metadata. The labels are configuration here, not
decoration: a Prometheus picks monitors up through its ServiceMonitor
selector. KubernetesKubePrometheusStack's default discovery (all_monitors)
loads every ServiceMonitor in the cluster; its release_managed_only
discovery loads only objects labelled `release: <the stack's release name>`
(its release_name output).

ServiceMonitor or PodMonitor: scrape through a Service (this kind) when the
workload already has one that names its metrics port -- the targets then
carry the Service's name and labels, and `target_labels` copies them onto
the series. Scrape pods directly (KubernetesPodMonitor) when there is no
such Service, as for a database operator's instance pods or a DaemonSet's
exporters.

What the operator refuses silently: it skips a whole ServiceMonitor -- every
endpoint of it scrapes nothing, and only its own log says why -- when a
referenced Secret or key is missing, when an endpoint sets two
authentication methods, when a client certificate lacks its key, or when a
relabeling step breaks its action's rules. Each of those is a validation
rule here instead.

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesServiceMonitor
metadata:
  name: api
spec:
  namespace:
    value: api
  labels:
    release: kube-prometheus-stack
  job_label: app.kubernetes.io/name
  target_labels:
    - team
  sample_limit: 50000
  selector:
    match_labels:
      app.kubernetes.io/name: api
  endpoints:
    - port: http-metrics
      interval: 30s
      scrape_timeout: 10s
      metric_relabelings:
        - source_labels:
            - __name__
          regex: go_gc_duration_seconds.*
          action: drop
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.jobLabel` | `string` |  |  |  |
| `spec.targetLabels` | `[]string` |  |  |  |
| `spec.podTargetLabels` | `[]string` |  |  |  |
| `spec.endpoints` | `[]KubernetesServiceMonitorEndpoint` | yes |  |  |
| `spec.endpoints[].port` | `string` |  |  |  |
| `spec.endpoints[].targetPort` | `string` |  |  |  |
| `spec.endpoints[].path` | `string` |  |  |  |
| `spec.endpoints[].scheme` | `string` |  |  |  |
| `spec.endpoints[].params` | `map<string, KubernetesPrometheusOperatorApiStringList>` |  |  |  |
| `spec.endpoints[].params.*.values` | `[]string` |  |  |  |
| `spec.endpoints[].interval` | `string` |  |  |  |
| `spec.endpoints[].scrapeTimeout` | `string` |  |  |  |
| `spec.endpoints[].tlsConfig` | `KubernetesPrometheusOperatorApiTlsConfig` |  |  |  |
| `spec.endpoints[].tlsConfig.ca` | `KubernetesPrometheusOperatorApiSecretOrConfigMap` |  |  |  |
| `spec.endpoints[].tlsConfig.ca.secret` | `KubernetesPrometheusOperatorApiSecretKeySelector` |  |  |  |
| `spec.endpoints[].tlsConfig.ca.secret.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.endpoints[].tlsConfig.ca.secret.key` | `string` | yes |  |  |
| `spec.endpoints[].tlsConfig.ca.secret.optional` | `bool` |  |  |  |
| `spec.endpoints[].tlsConfig.ca.configMap` | `KubernetesPrometheusOperatorApiConfigMapKeySelector` |  |  |  |
| `spec.endpoints[].tlsConfig.ca.configMap.name` | `string \| valueFrom` | yes |  | KubernetesConfigMap (`status.outputs.configmap_name`) |
| `spec.endpoints[].tlsConfig.ca.configMap.key` | `string` | yes |  |  |
| `spec.endpoints[].tlsConfig.ca.configMap.optional` | `bool` |  |  |  |
| `spec.endpoints[].tlsConfig.cert` | `KubernetesPrometheusOperatorApiSecretOrConfigMap` |  |  |  |
| `spec.endpoints[].tlsConfig.cert.secret` | `KubernetesPrometheusOperatorApiSecretKeySelector` |  |  |  |
| `spec.endpoints[].tlsConfig.cert.secret.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.endpoints[].tlsConfig.cert.secret.key` | `string` | yes |  |  |
| `spec.endpoints[].tlsConfig.cert.secret.optional` | `bool` |  |  |  |
| `spec.endpoints[].tlsConfig.cert.configMap` | `KubernetesPrometheusOperatorApiConfigMapKeySelector` |  |  |  |
| `spec.endpoints[].tlsConfig.cert.configMap.name` | `string \| valueFrom` | yes |  | KubernetesConfigMap (`status.outputs.configmap_name`) |
| `spec.endpoints[].tlsConfig.cert.configMap.key` | `string` | yes |  |  |
| `spec.endpoints[].tlsConfig.cert.configMap.optional` | `bool` |  |  |  |
| `spec.endpoints[].tlsConfig.keySecret` | `KubernetesPrometheusOperatorApiSecretKeySelector` |  |  |  |
| `spec.endpoints[].tlsConfig.keySecret.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.endpoints[].tlsConfig.keySecret.key` | `string` | yes |  |  |
| `spec.endpoints[].tlsConfig.keySecret.optional` | `bool` |  |  |  |
| `spec.endpoints[].tlsConfig.serverName` | `string` |  |  |  |
| `spec.endpoints[].tlsConfig.insecureSkipVerify` | `bool` |  |  |  |
| `spec.endpoints[].tlsConfig.minVersion` | `string` |  |  |  |
| `spec.endpoints[].tlsConfig.maxVersion` | `string` |  |  |  |
| `spec.endpoints[].tlsConfig.caFile` | `string` |  |  |  |
| `spec.endpoints[].tlsConfig.certFile` | `string` |  |  |  |
| `spec.endpoints[].tlsConfig.keyFile` | `string` |  |  |  |
| `spec.endpoints[].bearerTokenFile` | `string` |  |  |  |
| `spec.endpoints[].bearerTokenSecret` | `KubernetesPrometheusOperatorApiSecretKeySelector` |  |  |  |
| `spec.endpoints[].bearerTokenSecret.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.endpoints[].bearerTokenSecret.key` | `string` | yes |  |  |
| `spec.endpoints[].bearerTokenSecret.optional` | `bool` |  |  |  |
| `spec.endpoints[].authorization` | `KubernetesPrometheusOperatorApiSafeAuthorization` |  |  |  |
| `spec.endpoints[].authorization.type` | `string` |  |  |  |
| `spec.endpoints[].authorization.credentials` | `KubernetesPrometheusOperatorApiSecretKeySelector` | yes |  |  |
| `spec.endpoints[].authorization.credentials.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.endpoints[].authorization.credentials.key` | `string` | yes |  |  |
| `spec.endpoints[].authorization.credentials.optional` | `bool` |  |  |  |
| `spec.endpoints[].basicAuth` | `KubernetesPrometheusOperatorApiBasicAuth` |  |  |  |
| `spec.endpoints[].basicAuth.username` | `KubernetesPrometheusOperatorApiSecretKeySelector` |  |  |  |
| `spec.endpoints[].basicAuth.username.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.endpoints[].basicAuth.username.key` | `string` | yes |  |  |
| `spec.endpoints[].basicAuth.username.optional` | `bool` |  |  |  |
| `spec.endpoints[].basicAuth.password` | `KubernetesPrometheusOperatorApiSecretKeySelector` |  |  |  |
| `spec.endpoints[].basicAuth.password.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.endpoints[].basicAuth.password.key` | `string` | yes |  |  |
| `spec.endpoints[].basicAuth.password.optional` | `bool` |  |  |  |
| `spec.endpoints[].oauth2` | `KubernetesPrometheusOperatorApiOAuth2` |  |  |  |
| `spec.endpoints[].oauth2.clientId` | `KubernetesPrometheusOperatorApiSecretOrConfigMap` | yes |  |  |
| `spec.endpoints[].oauth2.clientId.secret` | `KubernetesPrometheusOperatorApiSecretKeySelector` |  |  |  |
| `spec.endpoints[].oauth2.clientId.secret.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.endpoints[].oauth2.clientId.secret.key` | `string` | yes |  |  |
| `spec.endpoints[].oauth2.clientId.secret.optional` | `bool` |  |  |  |
| `spec.endpoints[].oauth2.clientId.configMap` | `KubernetesPrometheusOperatorApiConfigMapKeySelector` |  |  |  |
| `spec.endpoints[].oauth2.clientId.configMap.name` | `string \| valueFrom` | yes |  | KubernetesConfigMap (`status.outputs.configmap_name`) |
| `spec.endpoints[].oauth2.clientId.configMap.key` | `string` | yes |  |  |
| `spec.endpoints[].oauth2.clientId.configMap.optional` | `bool` |  |  |  |
| `spec.endpoints[].oauth2.clientSecret` | `KubernetesPrometheusOperatorApiSecretKeySelector` | yes |  |  |
| `spec.endpoints[].oauth2.clientSecret.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.endpoints[].oauth2.clientSecret.key` | `string` | yes |  |  |
| `spec.endpoints[].oauth2.clientSecret.optional` | `bool` |  |  |  |
| `spec.endpoints[].oauth2.tokenUrl` | `string` | yes |  |  |
| `spec.endpoints[].oauth2.scopes` | `[]string` |  |  |  |
| `spec.endpoints[].oauth2.endpointParams` | `map<string, string>` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig` | `KubernetesPrometheusOperatorApiSafeTlsConfig` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig.ca` | `KubernetesPrometheusOperatorApiSecretOrConfigMap` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig.ca.secret` | `KubernetesPrometheusOperatorApiSecretKeySelector` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig.ca.secret.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.endpoints[].oauth2.tlsConfig.ca.secret.key` | `string` | yes |  |  |
| `spec.endpoints[].oauth2.tlsConfig.ca.secret.optional` | `bool` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig.ca.configMap` | `KubernetesPrometheusOperatorApiConfigMapKeySelector` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig.ca.configMap.name` | `string \| valueFrom` | yes |  | KubernetesConfigMap (`status.outputs.configmap_name`) |
| `spec.endpoints[].oauth2.tlsConfig.ca.configMap.key` | `string` | yes |  |  |
| `spec.endpoints[].oauth2.tlsConfig.ca.configMap.optional` | `bool` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig.cert` | `KubernetesPrometheusOperatorApiSecretOrConfigMap` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig.cert.secret` | `KubernetesPrometheusOperatorApiSecretKeySelector` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig.cert.secret.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.endpoints[].oauth2.tlsConfig.cert.secret.key` | `string` | yes |  |  |
| `spec.endpoints[].oauth2.tlsConfig.cert.secret.optional` | `bool` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig.cert.configMap` | `KubernetesPrometheusOperatorApiConfigMapKeySelector` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig.cert.configMap.name` | `string \| valueFrom` | yes |  | KubernetesConfigMap (`status.outputs.configmap_name`) |
| `spec.endpoints[].oauth2.tlsConfig.cert.configMap.key` | `string` | yes |  |  |
| `spec.endpoints[].oauth2.tlsConfig.cert.configMap.optional` | `bool` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig.keySecret` | `KubernetesPrometheusOperatorApiSecretKeySelector` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig.keySecret.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.endpoints[].oauth2.tlsConfig.keySecret.key` | `string` | yes |  |  |
| `spec.endpoints[].oauth2.tlsConfig.keySecret.optional` | `bool` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig.serverName` | `string` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig.insecureSkipVerify` | `bool` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig.minVersion` | `string` |  |  |  |
| `spec.endpoints[].oauth2.tlsConfig.maxVersion` | `string` |  |  |  |
| `spec.endpoints[].oauth2.proxyUrl` | `string` |  |  |  |
| `spec.endpoints[].oauth2.noProxy` | `string` |  |  |  |
| `spec.endpoints[].oauth2.proxyFromEnvironment` | `bool` |  |  |  |
| `spec.endpoints[].oauth2.proxyConnectHeader` | `map<string, KubernetesPrometheusOperatorApiSecretKeySelectorList>` |  |  |  |
| `spec.endpoints[].oauth2.proxyConnectHeader.*.values` | `[]KubernetesPrometheusOperatorApiSecretKeySelector` | yes |  |  |
| `spec.endpoints[].oauth2.proxyConnectHeader.*.values[].name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.endpoints[].oauth2.proxyConnectHeader.*.values[].key` | `string` | yes |  |  |
| `spec.endpoints[].oauth2.proxyConnectHeader.*.values[].optional` | `bool` |  |  |  |
| `spec.endpoints[].honorLabels` | `bool` |  |  |  |
| `spec.endpoints[].honorTimestamps` | `bool` |  |  |  |
| `spec.endpoints[].trackTimestampsStaleness` | `bool` |  |  |  |
| `spec.endpoints[].metricRelabelings` | `[]KubernetesPrometheusOperatorApiRelabelConfig` |  |  |  |
| `spec.endpoints[].metricRelabelings[].sourceLabels` | `[]string` |  |  |  |
| `spec.endpoints[].metricRelabelings[].separator` | `string` |  |  |  |
| `spec.endpoints[].metricRelabelings[].targetLabel` | `string` |  |  |  |
| `spec.endpoints[].metricRelabelings[].regex` | `string` |  |  |  |
| `spec.endpoints[].metricRelabelings[].modulus` | `uint32` |  |  |  |
| `spec.endpoints[].metricRelabelings[].replacement` | `string` |  |  |  |
| `spec.endpoints[].metricRelabelings[].action` | `string` |  |  |  |
| `spec.endpoints[].relabelings` | `[]KubernetesPrometheusOperatorApiRelabelConfig` |  |  |  |
| `spec.endpoints[].relabelings[].sourceLabels` | `[]string` |  |  |  |
| `spec.endpoints[].relabelings[].separator` | `string` |  |  |  |
| `spec.endpoints[].relabelings[].targetLabel` | `string` |  |  |  |
| `spec.endpoints[].relabelings[].regex` | `string` |  |  |  |
| `spec.endpoints[].relabelings[].modulus` | `uint32` |  |  |  |
| `spec.endpoints[].relabelings[].replacement` | `string` |  |  |  |
| `spec.endpoints[].relabelings[].action` | `string` |  |  |  |
| `spec.endpoints[].proxyUrl` | `string` |  |  |  |
| `spec.endpoints[].noProxy` | `string` |  |  |  |
| `spec.endpoints[].proxyFromEnvironment` | `bool` |  |  |  |
| `spec.endpoints[].proxyConnectHeader` | `map<string, KubernetesPrometheusOperatorApiSecretKeySelectorList>` |  |  |  |
| `spec.endpoints[].proxyConnectHeader.*.values` | `[]KubernetesPrometheusOperatorApiSecretKeySelector` | yes |  |  |
| `spec.endpoints[].proxyConnectHeader.*.values[].name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.endpoints[].proxyConnectHeader.*.values[].key` | `string` | yes |  |  |
| `spec.endpoints[].proxyConnectHeader.*.values[].optional` | `bool` |  |  |  |
| `spec.endpoints[].followRedirects` | `bool` |  |  |  |
| `spec.endpoints[].enableHttp2` | `bool` |  |  |  |
| `spec.endpoints[].filterRunning` | `bool` |  |  |  |
| `spec.selector` | `KubernetesPrometheusOperatorApiLabelSelector` | yes |  |  |
| `spec.selector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.selector.matchExpressions` | `[]KubernetesPrometheusOperatorApiLabelSelectorRequirement` |  |  |  |
| `spec.selector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.selector.matchExpressions[].operator` | `string` |  |  |  |
| `spec.selector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.selectorMechanism` | `string` |  |  |  |
| `spec.namespaceSelector` | `KubernetesPrometheusOperatorApiNamespaceSelector` |  |  |  |
| `spec.namespaceSelector.any` | `bool` |  |  |  |
| `spec.namespaceSelector.matchNames` | `[]string \| valueFrom` |  |  | KubernetesNamespace (`spec.name`) |
| `spec.sampleLimit` | `uint32` |  |  |  |
| `spec.targetLimit` | `uint32` |  |  |  |
| `spec.scrapeProtocols` | `[]string` |  |  |  |
| `spec.fallbackScrapeProtocol` | `string` |  |  |  |
| `spec.labelLimit` | `uint32` |  |  |  |
| `spec.labelNameLengthLimit` | `uint32` |  |  |  |
| `spec.labelValueLengthLimit` | `uint32` |  |  |  |
| `spec.scrapeNativeHistograms` | `bool` |  |  |  |
| `spec.scrapeClassicHistograms` | `bool` |  |  |  |
| `spec.nativeHistogramBucketLimit` | `uint32` |  |  |  |
| `spec.nativeHistogramMinBucketFactor` | `string` |  |  |  |
| `spec.convertClassicHistogramsToNHCB` | `bool` |  |  |  |
| `spec.keepDroppedTargets` | `uint32` |  |  |  |
| `spec.attachMetadata` | `KubernetesPrometheusOperatorApiAttachMetadata` |  |  |  |
| `spec.attachMetadata.node` | `bool` |  |  |  |
| `spec.scrapeClass` | `string` | yes |  |  |
| `spec.bodySizeLimit` | `string` |  |  |  |
| `spec.serviceDiscoveryRole` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Kubernetes namespace the ServiceMonitor is created in. Typically a
reference to a KubernetesNamespace resource's `spec.name`.

It is also where the Services are searched unless `namespace_selector`
says otherwise, and where every Secret and ConfigMap the endpoints
reference must live. Put the monitor beside the workload it watches.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.labels

`map<string, string>`

Labels on the ServiceMonitor object itself (its metadata.labels), merged
under Planton's identity labels (a key in both keeps the identity value).

These are how a Prometheus with a ServiceMonitor selector decides the
monitor is its own: a KubernetesKubePrometheusStack on
release_managed_only discovery loads only objects labelled
`release: <its release_name output>`, and a Prometheus installed any
other way selects by whatever its serviceMonitorSelector names. Under the
stack's default all_monitors discovery no label is needed.

These are NOT labels added to the scraped series; for that, use
`target_labels`, `pod_target_labels` or a relabeling.

- rule: {"map":{"keys":{"string":{"minLen":"1","maxLen":"317","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$"}},"values":{"string":{"maxLen":"63","pattern":"^([A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?)?$"}}}}

### spec.annotations

`map<string, string>`

Annotations on the ServiceMonitor object itself (its
metadata.annotations): notes for people and tools reading the object,
such as an owning team. They do not reach Prometheus or the series.

### spec.jobLabel

`string`

The Service label whose value becomes every series' `job` label. With
`job_label: app.kubernetes.io/name` and a Service labelled
`app.kubernetes.io/name: api`, the series carry `job="api"`. Empty, or when
the Service lacks the label, `job` is the Service's name.

`job` is what dashboards and alerts group by, so pick a label whose value
stays the same across environments and releases.

### spec.targetLabels

`[]string`

Service labels copied onto every scraped series, each under its own name
(a Service labelled `team: payments` adds `team="payments"`).

### spec.podTargetLabels

`[]string`

Labels of the pod behind each endpoint copied onto every scraped series
(`app.kubernetes.io/version` to tell a canary's series from the rest).

### spec.endpoints

`[]KubernetesServiceMonitorEndpoint` · required

How each matching Service is scraped: one entry per port, each its own
set of targets. At least one is required; a monitor with none scrapes
nothing (upstream requires the key and protojson cannot write an empty
list).

- rule: {"repeated":{"minItems":"1"}}
- rule: Set at most one of authorization, basic_auth, oauth2 and bearer_token_secret: the operator skips a monitor whose endpoint has two
- rule: proxy_connect_header needs proxy_url or proxy_from_environment
- rule: proxy_from_environment takes the proxy from the environment, so proxy_url and no_proxy must be unset
- rule: no_proxy needs proxy_url

### spec.endpoints[].port

`string`

The Service port to scrape, by its name (`spec.ports[].name` on the
Service). Takes precedence over target_port. Name the metrics port on
the Service ("metrics", "http-metrics") and use it here.

### spec.endpoints[].targetPort

`string` · optional (explicit presence)

The container port to scrape on the pods behind the Service, by number
("9090") or by container port name ("metrics"). An all-digit value is
written to the object as a number, anything else as a name. Used when
`port` is unset -- for a Service that exposes the port without naming it.

- rule: target_port is a port number from 1 to 65535, or a container port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)

### spec.endpoints[].path

`string`

The HTTP path scraped. Empty, "/metrics".

### spec.endpoints[].scheme

`string` · optional (explicit presence)

The scheme scraped: "http" or "https" (either case). Unset, "http"; set
"https" together with tls_config for a TLS-only metrics port.

- rule: {"string":{"in":["http","https","HTTP","HTTPS"]}}

### spec.endpoints[].params

`map<string, KubernetesPrometheusOperatorApiStringList>`

URL query parameters sent with every scrape, each with one or more
values: `{format: {values: [prometheus]}}` in the manifest,
`{format: [prometheus]}` on the object. An exporter that serves several
modules from one port (a blackbox or SNMP exporter's `module`) is told
which through these.

### spec.endpoints[].params.*.values

`[]string`

The parameter's values, each sent as its own `name=value` pair.

### spec.endpoints[].interval

`string`

How often the targets are scraped, as a Prometheus duration ("30s",
"1m"). Empty, the Prometheus's scrape interval. A shorter interval costs
samples in proportion; most services need nothing finer than 30s.

- rule: {"string":{"pattern":"^(0|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$"}}

### spec.endpoints[].scrapeTimeout

`string`

How long a scrape may take before it counts as failed, as a Prometheus
duration. Empty, the Prometheus's scrape timeout, capped at `interval`.
It may not exceed `interval`: the operator skips the monitor otherwise.

- rule: {"string":{"pattern":"^(0|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$"}}

### spec.endpoints[].tlsConfig

`KubernetesPrometheusOperatorApiTlsConfig`

TLS for the scrape connection (with `scheme: https`).

- rule: Take the CA from ca or from ca_file, not both
- rule: Take the client certificate from cert or from cert_file, not both
- rule: Take the client key from key_secret or from key_file, not both
- rule: A client certificate (cert or cert_file) and its private key (key_secret or key_file) are set together or not at all
- rule: max_version must be at least min_version

### spec.endpoints[].tlsConfig.ca

`KubernetesPrometheusOperatorApiSecretOrConfigMap`

The certificate authority used to verify the target's certificate, from a
Secret or ConfigMap. Excludes ca_file.

- rule: Take the value from a Secret or from a ConfigMap, not both

### spec.endpoints[].tlsConfig.ca.secret

`KubernetesPrometheusOperatorApiSecretKeySelector`

The Secret key holding the value.

### spec.endpoints[].tlsConfig.ca.secret.name

`string | valueFrom` · required

Name of the Secret, in the monitor's namespace. Defaults to a
KubernetesSecret foreign key (its secret_name output); pass the literal
name with `value:` for a Secret Planton does not manage. Upstream allows an
empty name for backwards compatibility only and calls it "almost
certainly wrong", so a name is required here.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.endpoints[].tlsConfig.ca.secret.key

`string` · required

The key within the Secret's data whose value is used ("token",
"password", "ca.crt").

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].tlsConfig.ca.secret.optional

`bool`

When true, a missing Secret or key is tolerated instead of making the
operator skip the monitor. Leave it false for credentials: a scrape
without its credentials fails anyway, and false makes the operator say so.

### spec.endpoints[].tlsConfig.ca.configMap

`KubernetesPrometheusOperatorApiConfigMapKeySelector`

The ConfigMap key holding the value.

### spec.endpoints[].tlsConfig.ca.configMap.name

`string | valueFrom` · required

Name of the ConfigMap, in the monitor's namespace. Defaults to a
KubernetesConfigMap foreign key (its configmap_name output); pass the
literal name with `value:` for a ConfigMap Planton does not manage.

- references: KubernetesConfigMap (`status.outputs.configmap_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesConfigMap, name: <that resource's name>, fieldPath: status.outputs.configmap_name}} -- a bare string does not parse

### spec.endpoints[].tlsConfig.ca.configMap.key

`string` · required

The key within the ConfigMap's `data` whose value is used. Keys under
`binaryData` are not read.

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].tlsConfig.ca.configMap.optional

`bool`

When true, a missing ConfigMap or key is tolerated instead of making the
operator skip the monitor.

### spec.endpoints[].tlsConfig.cert

`KubernetesPrometheusOperatorApiSecretOrConfigMap`

The client certificate presented for mutual TLS, from a Secret or
ConfigMap. Requires a key (key_secret or key_file); excludes cert_file.

- rule: Take the value from a Secret or from a ConfigMap, not both

### spec.endpoints[].tlsConfig.cert.secret

`KubernetesPrometheusOperatorApiSecretKeySelector`

The Secret key holding the value.

### spec.endpoints[].tlsConfig.cert.secret.name

`string | valueFrom` · required

Name of the Secret, in the monitor's namespace. Defaults to a
KubernetesSecret foreign key (its secret_name output); pass the literal
name with `value:` for a Secret Planton does not manage. Upstream allows an
empty name for backwards compatibility only and calls it "almost
certainly wrong", so a name is required here.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.endpoints[].tlsConfig.cert.secret.key

`string` · required

The key within the Secret's data whose value is used ("token",
"password", "ca.crt").

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].tlsConfig.cert.secret.optional

`bool`

When true, a missing Secret or key is tolerated instead of making the
operator skip the monitor. Leave it false for credentials: a scrape
without its credentials fails anyway, and false makes the operator say so.

### spec.endpoints[].tlsConfig.cert.configMap

`KubernetesPrometheusOperatorApiConfigMapKeySelector`

The ConfigMap key holding the value.

### spec.endpoints[].tlsConfig.cert.configMap.name

`string | valueFrom` · required

Name of the ConfigMap, in the monitor's namespace. Defaults to a
KubernetesConfigMap foreign key (its configmap_name output); pass the
literal name with `value:` for a ConfigMap Planton does not manage.

- references: KubernetesConfigMap (`status.outputs.configmap_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesConfigMap, name: <that resource's name>, fieldPath: status.outputs.configmap_name}} -- a bare string does not parse

### spec.endpoints[].tlsConfig.cert.configMap.key

`string` · required

The key within the ConfigMap's `data` whose value is used. Keys under
`binaryData` are not read.

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].tlsConfig.cert.configMap.optional

`bool`

When true, a missing ConfigMap or key is tolerated instead of making the
operator skip the monitor.

### spec.endpoints[].tlsConfig.keySecret

`KubernetesPrometheusOperatorApiSecretKeySelector`

The Secret key holding the client certificate's private key. Requires a
certificate (cert or cert_file); excludes key_file.

### spec.endpoints[].tlsConfig.keySecret.name

`string | valueFrom` · required

Name of the Secret, in the monitor's namespace. Defaults to a
KubernetesSecret foreign key (its secret_name output); pass the literal
name with `value:` for a Secret Planton does not manage. Upstream allows an
empty name for backwards compatibility only and calls it "almost
certainly wrong", so a name is required here.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.endpoints[].tlsConfig.keySecret.key

`string` · required

The key within the Secret's data whose value is used ("token",
"password", "ca.crt").

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].tlsConfig.keySecret.optional

`bool`

When true, a missing Secret or key is tolerated instead of making the
operator skip the monitor. Leave it false for credentials: a scrape
without its credentials fails anyway, and false makes the operator say so.

### spec.endpoints[].tlsConfig.serverName

`string` · optional (explicit presence)

The hostname the target's certificate must carry, when it differs from
the address Prometheus dials.

### spec.endpoints[].tlsConfig.insecureSkipVerify

`bool` · optional (explicit presence)

Skips verification of the target's certificate. Prefer a CA.

### spec.endpoints[].tlsConfig.minVersion

`string` · optional (explicit presence)

The minimum TLS version accepted: "TLS10", "TLS11", "TLS12" or "TLS13".
Requires Prometheus >= 2.35.

- rule: {"string":{"in":["TLS10","TLS11","TLS12","TLS13"]}}

### spec.endpoints[].tlsConfig.maxVersion

`string` · optional (explicit presence)

The maximum TLS version accepted, at least min_version. Requires
Prometheus >= 2.41.

- rule: {"string":{"in":["TLS10","TLS11","TLS12","TLS13"]}}

### spec.endpoints[].tlsConfig.caFile

`string`

Path of a CA certificate file inside the Prometheus container. Excludes
ca. The in-cluster API server's CA is
/var/run/secrets/kubernetes.io/serviceaccount/ca.crt.

### spec.endpoints[].tlsConfig.certFile

`string`

Path of a client certificate file inside the Prometheus container.
Requires a key; excludes cert.

### spec.endpoints[].tlsConfig.keyFile

`string`

Path of the client certificate's private key file inside the Prometheus
container. Requires a certificate; excludes key_secret.

### spec.endpoints[].bearerTokenFile

`string`

Deprecated upstream: use `authorization` with a Secret reference instead.
A path inside the Prometheus container whose contents are sent as a
bearer token -- typically the pod's own service-account token
(/var/run/secrets/kubernetes.io/serviceaccount/token) when scraping the
Kubernetes API itself. A Prometheus whose
`arbitraryFSAccessThroughSMs.deny` is set refuses a ServiceMonitor that
names one.

### spec.endpoints[].bearerTokenSecret

`KubernetesPrometheusOperatorApiSecretKeySelector`

Deprecated upstream: use `authorization` instead, which reads the same
Secret key. The Secret key holding a bearer token sent with every scrape.

### spec.endpoints[].bearerTokenSecret.name

`string | valueFrom` · required

Name of the Secret, in the monitor's namespace. Defaults to a
KubernetesSecret foreign key (its secret_name output); pass the literal
name with `value:` for a Secret Planton does not manage. Upstream allows an
empty name for backwards compatibility only and calls it "almost
certainly wrong", so a name is required here.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.endpoints[].bearerTokenSecret.key

`string` · required

The key within the Secret's data whose value is used ("token",
"password", "ca.crt").

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].bearerTokenSecret.optional

`bool`

When true, a missing Secret or key is tolerated instead of making the
operator skip the monitor. Leave it false for credentials: a scrape
without its credentials fails anyway, and false makes the operator say so.

### spec.endpoints[].authorization

`KubernetesPrometheusOperatorApiSafeAuthorization`

The Authorization header sent with every scrape (`Bearer <token>` by
default), the token read from a Secret key. Excludes the other
authentication methods.

### spec.endpoints[].authorization.type

`string` · optional (explicit presence)

The authentication scheme written before the credentials,
case-insensitive. Unset, Prometheus sends "Bearer". "Basic" is refused
upstream: use basic_auth instead.

- rule: Basic authentication is basic_auth, not an authorization type

### spec.endpoints[].authorization.credentials

`KubernetesPrometheusOperatorApiSecretKeySelector` · required

The Secret key holding the credentials (the token itself). Required: the
operator skips a monitor whose authorization has none.

- rule: {"required":true}

### spec.endpoints[].authorization.credentials.name

`string | valueFrom` · required

Name of the Secret, in the monitor's namespace. Defaults to a
KubernetesSecret foreign key (its secret_name output); pass the literal
name with `value:` for a Secret Planton does not manage. Upstream allows an
empty name for backwards compatibility only and calls it "almost
certainly wrong", so a name is required here.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.endpoints[].authorization.credentials.key

`string` · required

The key within the Secret's data whose value is used ("token",
"password", "ca.crt").

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].authorization.credentials.optional

`bool`

When true, a missing Secret or key is tolerated instead of making the
operator skip the monitor. Leave it false for credentials: a scrape
without its credentials fails anyway, and false makes the operator say so.

### spec.endpoints[].basicAuth

`KubernetesPrometheusOperatorApiBasicAuth`

HTTP Basic authentication, both halves read from Secret keys. Excludes
the other authentication methods.

### spec.endpoints[].basicAuth.username

`KubernetesPrometheusOperatorApiSecretKeySelector`

The Secret key holding the username.

### spec.endpoints[].basicAuth.username.name

`string | valueFrom` · required

Name of the Secret, in the monitor's namespace. Defaults to a
KubernetesSecret foreign key (its secret_name output); pass the literal
name with `value:` for a Secret Planton does not manage. Upstream allows an
empty name for backwards compatibility only and calls it "almost
certainly wrong", so a name is required here.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.endpoints[].basicAuth.username.key

`string` · required

The key within the Secret's data whose value is used ("token",
"password", "ca.crt").

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].basicAuth.username.optional

`bool`

When true, a missing Secret or key is tolerated instead of making the
operator skip the monitor. Leave it false for credentials: a scrape
without its credentials fails anyway, and false makes the operator say so.

### spec.endpoints[].basicAuth.password

`KubernetesPrometheusOperatorApiSecretKeySelector`

The Secret key holding the password.

### spec.endpoints[].basicAuth.password.name

`string | valueFrom` · required

Name of the Secret, in the monitor's namespace. Defaults to a
KubernetesSecret foreign key (its secret_name output); pass the literal
name with `value:` for a Secret Planton does not manage. Upstream allows an
empty name for backwards compatibility only and calls it "almost
certainly wrong", so a name is required here.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.endpoints[].basicAuth.password.key

`string` · required

The key within the Secret's data whose value is used ("token",
"password", "ca.crt").

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].basicAuth.password.optional

`bool`

When true, a missing Secret or key is tolerated instead of making the
operator skip the monitor. Leave it false for credentials: a scrape
without its credentials fails anyway, and false makes the operator say so.

### spec.endpoints[].oauth2

`KubernetesPrometheusOperatorApiOAuth2`

OAuth2 client-credentials authentication. Excludes the other
authentication methods. Requires Prometheus >= 2.27.

- rule: client_id takes its value from a Secret or a ConfigMap
- rule: proxy_connect_header needs proxy_url or proxy_from_environment
- rule: proxy_from_environment takes the proxy from the environment, so proxy_url and no_proxy must be unset
- rule: no_proxy needs proxy_url

### spec.endpoints[].oauth2.clientId

`KubernetesPrometheusOperatorApiSecretOrConfigMap` · required

The OAuth2 client's id, from a Secret or (as it is not secret) a
ConfigMap.

- rule: {"required":true}
- rule: Take the value from a Secret or from a ConfigMap, not both

### spec.endpoints[].oauth2.clientId.secret

`KubernetesPrometheusOperatorApiSecretKeySelector`

The Secret key holding the value.

### spec.endpoints[].oauth2.clientId.secret.name

`string | valueFrom` · required

Name of the Secret, in the monitor's namespace. Defaults to a
KubernetesSecret foreign key (its secret_name output); pass the literal
name with `value:` for a Secret Planton does not manage. Upstream allows an
empty name for backwards compatibility only and calls it "almost
certainly wrong", so a name is required here.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.endpoints[].oauth2.clientId.secret.key

`string` · required

The key within the Secret's data whose value is used ("token",
"password", "ca.crt").

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].oauth2.clientId.secret.optional

`bool`

When true, a missing Secret or key is tolerated instead of making the
operator skip the monitor. Leave it false for credentials: a scrape
without its credentials fails anyway, and false makes the operator say so.

### spec.endpoints[].oauth2.clientId.configMap

`KubernetesPrometheusOperatorApiConfigMapKeySelector`

The ConfigMap key holding the value.

### spec.endpoints[].oauth2.clientId.configMap.name

`string | valueFrom` · required

Name of the ConfigMap, in the monitor's namespace. Defaults to a
KubernetesConfigMap foreign key (its configmap_name output); pass the
literal name with `value:` for a ConfigMap Planton does not manage.

- references: KubernetesConfigMap (`status.outputs.configmap_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesConfigMap, name: <that resource's name>, fieldPath: status.outputs.configmap_name}} -- a bare string does not parse

### spec.endpoints[].oauth2.clientId.configMap.key

`string` · required

The key within the ConfigMap's `data` whose value is used. Keys under
`binaryData` are not read.

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].oauth2.clientId.configMap.optional

`bool`

When true, a missing ConfigMap or key is tolerated instead of making the
operator skip the monitor.

### spec.endpoints[].oauth2.clientSecret

`KubernetesPrometheusOperatorApiSecretKeySelector` · required

The Secret key holding the OAuth2 client's secret.

- rule: {"required":true}

### spec.endpoints[].oauth2.clientSecret.name

`string | valueFrom` · required

Name of the Secret, in the monitor's namespace. Defaults to a
KubernetesSecret foreign key (its secret_name output); pass the literal
name with `value:` for a Secret Planton does not manage. Upstream allows an
empty name for backwards compatibility only and calls it "almost
certainly wrong", so a name is required here.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.endpoints[].oauth2.clientSecret.key

`string` · required

The key within the Secret's data whose value is used ("token",
"password", "ca.crt").

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].oauth2.clientSecret.optional

`bool`

When true, a missing Secret or key is tolerated instead of making the
operator skip the monitor. Leave it false for credentials: a scrape
without its credentials fails anyway, and false makes the operator say so.

### spec.endpoints[].oauth2.tokenUrl

`string` · required

The token endpoint, "https://<issuer>/oauth2/token".

- rule: {"string":{"minLen":"1","pattern":"^(http|https)://.+$"}}

### spec.endpoints[].oauth2.scopes

`[]string`

The scopes requested with the token.

### spec.endpoints[].oauth2.endpointParams

`map<string, string>`

Extra form parameters sent to the token endpoint (an `audience`, a
`resource`).

### spec.endpoints[].oauth2.tlsConfig

`KubernetesPrometheusOperatorApiSafeTlsConfig`

TLS for the connection to the token endpoint (not to the target).
Requires Prometheus >= 2.43.

- rule: A client certificate (cert) and its private key (key_secret) are set together or not at all
- rule: max_version must be at least min_version

### spec.endpoints[].oauth2.tlsConfig.ca

`KubernetesPrometheusOperatorApiSecretOrConfigMap`

The certificate authority used to verify the target's certificate. Unset,
Prometheus trusts its container's system roots, which a cluster-internal
CA is not among.

- rule: Take the value from a Secret or from a ConfigMap, not both

### spec.endpoints[].oauth2.tlsConfig.ca.secret

`KubernetesPrometheusOperatorApiSecretKeySelector`

The Secret key holding the value.

### spec.endpoints[].oauth2.tlsConfig.ca.secret.name

`string | valueFrom` · required

Name of the Secret, in the monitor's namespace. Defaults to a
KubernetesSecret foreign key (its secret_name output); pass the literal
name with `value:` for a Secret Planton does not manage. Upstream allows an
empty name for backwards compatibility only and calls it "almost
certainly wrong", so a name is required here.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.endpoints[].oauth2.tlsConfig.ca.secret.key

`string` · required

The key within the Secret's data whose value is used ("token",
"password", "ca.crt").

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].oauth2.tlsConfig.ca.secret.optional

`bool`

When true, a missing Secret or key is tolerated instead of making the
operator skip the monitor. Leave it false for credentials: a scrape
without its credentials fails anyway, and false makes the operator say so.

### spec.endpoints[].oauth2.tlsConfig.ca.configMap

`KubernetesPrometheusOperatorApiConfigMapKeySelector`

The ConfigMap key holding the value.

### spec.endpoints[].oauth2.tlsConfig.ca.configMap.name

`string | valueFrom` · required

Name of the ConfigMap, in the monitor's namespace. Defaults to a
KubernetesConfigMap foreign key (its configmap_name output); pass the
literal name with `value:` for a ConfigMap Planton does not manage.

- references: KubernetesConfigMap (`status.outputs.configmap_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesConfigMap, name: <that resource's name>, fieldPath: status.outputs.configmap_name}} -- a bare string does not parse

### spec.endpoints[].oauth2.tlsConfig.ca.configMap.key

`string` · required

The key within the ConfigMap's `data` whose value is used. Keys under
`binaryData` are not read.

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].oauth2.tlsConfig.ca.configMap.optional

`bool`

When true, a missing ConfigMap or key is tolerated instead of making the
operator skip the monitor.

### spec.endpoints[].oauth2.tlsConfig.cert

`KubernetesPrometheusOperatorApiSecretOrConfigMap`

The client certificate presented for mutual TLS. Requires key_secret.

- rule: Take the value from a Secret or from a ConfigMap, not both

### spec.endpoints[].oauth2.tlsConfig.cert.secret

`KubernetesPrometheusOperatorApiSecretKeySelector`

The Secret key holding the value.

### spec.endpoints[].oauth2.tlsConfig.cert.secret.name

`string | valueFrom` · required

Name of the Secret, in the monitor's namespace. Defaults to a
KubernetesSecret foreign key (its secret_name output); pass the literal
name with `value:` for a Secret Planton does not manage. Upstream allows an
empty name for backwards compatibility only and calls it "almost
certainly wrong", so a name is required here.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.endpoints[].oauth2.tlsConfig.cert.secret.key

`string` · required

The key within the Secret's data whose value is used ("token",
"password", "ca.crt").

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].oauth2.tlsConfig.cert.secret.optional

`bool`

When true, a missing Secret or key is tolerated instead of making the
operator skip the monitor. Leave it false for credentials: a scrape
without its credentials fails anyway, and false makes the operator say so.

### spec.endpoints[].oauth2.tlsConfig.cert.configMap

`KubernetesPrometheusOperatorApiConfigMapKeySelector`

The ConfigMap key holding the value.

### spec.endpoints[].oauth2.tlsConfig.cert.configMap.name

`string | valueFrom` · required

Name of the ConfigMap, in the monitor's namespace. Defaults to a
KubernetesConfigMap foreign key (its configmap_name output); pass the
literal name with `value:` for a ConfigMap Planton does not manage.

- references: KubernetesConfigMap (`status.outputs.configmap_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesConfigMap, name: <that resource's name>, fieldPath: status.outputs.configmap_name}} -- a bare string does not parse

### spec.endpoints[].oauth2.tlsConfig.cert.configMap.key

`string` · required

The key within the ConfigMap's `data` whose value is used. Keys under
`binaryData` are not read.

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].oauth2.tlsConfig.cert.configMap.optional

`bool`

When true, a missing ConfigMap or key is tolerated instead of making the
operator skip the monitor.

### spec.endpoints[].oauth2.tlsConfig.keySecret

`KubernetesPrometheusOperatorApiSecretKeySelector`

The Secret key holding the client certificate's private key. Requires
cert.

### spec.endpoints[].oauth2.tlsConfig.keySecret.name

`string | valueFrom` · required

Name of the Secret, in the monitor's namespace. Defaults to a
KubernetesSecret foreign key (its secret_name output); pass the literal
name with `value:` for a Secret Planton does not manage. Upstream allows an
empty name for backwards compatibility only and calls it "almost
certainly wrong", so a name is required here.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.endpoints[].oauth2.tlsConfig.keySecret.key

`string` · required

The key within the Secret's data whose value is used ("token",
"password", "ca.crt").

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].oauth2.tlsConfig.keySecret.optional

`bool`

When true, a missing Secret or key is tolerated instead of making the
operator skip the monitor. Leave it false for credentials: a scrape
without its credentials fails anyway, and false makes the operator say so.

### spec.endpoints[].oauth2.tlsConfig.serverName

`string` · optional (explicit presence)

The hostname the target's certificate must carry, when it differs from
the address Prometheus dials (a pod IP never matches a certificate's
names).

### spec.endpoints[].oauth2.tlsConfig.insecureSkipVerify

`bool` · optional (explicit presence)

Skips verification of the target's certificate. The connection is still
encrypted, but anyone on the path can impersonate the target; prefer `ca`
with the issuing CA.

### spec.endpoints[].oauth2.tlsConfig.minVersion

`string` · optional (explicit presence)

The minimum TLS version accepted: "TLS10", "TLS11", "TLS12" or "TLS13".
Requires Prometheus >= 2.35.

- rule: {"string":{"in":["TLS10","TLS11","TLS12","TLS13"]}}

### spec.endpoints[].oauth2.tlsConfig.maxVersion

`string` · optional (explicit presence)

The maximum TLS version accepted, at least min_version: "TLS10",
"TLS11", "TLS12" or "TLS13". Requires Prometheus >= 2.41.

- rule: {"string":{"in":["TLS10","TLS11","TLS12","TLS13"]}}

### spec.endpoints[].oauth2.proxyUrl

`string` · optional (explicit presence)

The proxy that token requests go through, "http://", "https://" or
"socks5://". Excludes proxy_from_environment. Requires Prometheus >= 2.43.

- rule: {"string":{"pattern":"^(http|https|socks5)://.+$"}}

### spec.endpoints[].oauth2.noProxy

`string` · optional (explicit presence)

Comma-separated hosts, domains, IPs or CIDRs (ports allowed) that bypass
proxy_url. Requires proxy_url. Requires Prometheus >= 2.43.

### spec.endpoints[].oauth2.proxyFromEnvironment

`bool` · optional (explicit presence)

Take the proxy from the Prometheus container's HTTP_PROXY, HTTPS_PROXY and
NO_PROXY environment variables. Excludes proxy_url and no_proxy. Requires
Prometheus >= 2.43.

### spec.endpoints[].oauth2.proxyConnectHeader

`map<string, KubernetesPrometheusOperatorApiSecretKeySelectorList>`

Headers sent to the proxy on CONNECT, each value read from Secret keys
(a header may repeat): `{Proxy-Authorization: {values: [{name: ...,
key: ...}]}}` in the manifest, `{Proxy-Authorization: [...]}` on the
object. Requires proxy_url or proxy_from_environment. Requires Prometheus
>= 2.43.

### spec.endpoints[].oauth2.proxyConnectHeader.*.values

`[]KubernetesPrometheusOperatorApiSecretKeySelector` · required

The Secret keys whose values are sent, one header line each. Upstream
refuses an empty list.

- rule: {"repeated":{"minItems":"1"}}

### spec.endpoints[].oauth2.proxyConnectHeader.*.values[].name

`string | valueFrom` · required

Name of the Secret, in the monitor's namespace. Defaults to a
KubernetesSecret foreign key (its secret_name output); pass the literal
name with `value:` for a Secret Planton does not manage. Upstream allows an
empty name for backwards compatibility only and calls it "almost
certainly wrong", so a name is required here.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.endpoints[].oauth2.proxyConnectHeader.*.values[].key

`string` · required

The key within the Secret's data whose value is used ("token",
"password", "ca.crt").

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].oauth2.proxyConnectHeader.*.values[].optional

`bool`

When true, a missing Secret or key is tolerated instead of making the
operator skip the monitor. Leave it false for credentials: a scrape
without its credentials fails anyway, and false makes the operator say so.

### spec.endpoints[].honorLabels

`bool`

When a scraped series carries a label the target also sets (`job`,
`instance`, a target label), keep the series' value instead of renaming
it to `exported_<label>`. For federation and pushgateway-style targets
that report on behalf of others.

### spec.endpoints[].honorTimestamps

`bool` · optional (explicit presence)

Keep the timestamps a target exposes instead of the scrape time. Unset,
Prometheus keeps them.

### spec.endpoints[].trackTimestampsStaleness

`bool` · optional (explicit presence)

Mark a series stale when it disappears from a target even though its
samples carried explicit timestamps. No effect when honor_timestamps is
false. Requires Prometheus >= 2.48.

### spec.endpoints[].metricRelabelings

`[]KubernetesPrometheusOperatorApiRelabelConfig`

Relabeling applied to each scraped sample before it is stored: drop
expensive series (`action: drop` on `__name__`) or rename labels.

- rule: replace (the default action), hashmod, lowercase, uppercase, keepequal and dropequal write a label, so they need target_label
- rule: hashmod needs a non-zero modulus
- rule: lowercase, uppercase, keepequal and dropequal take no replacement
- rule: keepequal and dropequal take only source_labels and target_label
- rule: labeldrop and labelkeep take only regex

### spec.endpoints[].metricRelabelings[].sourceLabels

`[]string`

The labels whose values are joined (with `separator`) and matched against
`regex`. Target-side relabelings read discovery labels such as
`__meta_kubernetes_pod_label_app`; metric relabelings read the scraped
series' labels, including `__name__`.

### spec.endpoints[].metricRelabelings[].separator

`string` · optional (explicit presence)

The string joining the source labels' values. Unset, ";".

### spec.endpoints[].metricRelabelings[].targetLabel

`string`

The label a replace, hashmod, lowercase, uppercase, keepequal or dropequal
step writes or compares. Regex capture groups are available ("${1}").

### spec.endpoints[].metricRelabelings[].regex

`string`

The RE2 expression the joined source value must match (anchored at both
ends). Empty, "(.*)".

### spec.endpoints[].metricRelabelings[].modulus

`uint32` · optional (explicit presence)

The modulus a hashmod step takes of the hash of the source value; with a
`keep` step on the result it shards targets across Prometheus replicas.
hashmod only, and required there.

### spec.endpoints[].metricRelabelings[].replacement

`string` · optional (explicit presence)

The value a replace step writes, with regex capture groups ("$1",
"${1}-suffix"); a labelmap step's label-name template. Unset, "$1".

### spec.endpoints[].metricRelabelings[].action

`string`

What the step does, case-insensitive: "replace" (the default), "keep",
"drop", "hashmod", "labelmap", "labeldrop", "labelkeep", "lowercase",
"uppercase" (Prometheus >= 2.36), "keepequal" or "dropequal" (Prometheus
>= 2.41). keep and drop filter targets or series; labeldrop and labelkeep
filter label names by regex.

- rule: {"string":{"in":["","replace","Replace","keep","Keep","drop","Drop","hashmod","HashMod","labelmap","LabelMap","labeldrop","LabelDrop","labelkeep","LabelKeep","lowercase","Lowercase","uppercase","Uppercase","keepequal","KeepEqual","dropequal","DropEqual"]}}

### spec.endpoints[].relabelings

`[]KubernetesPrometheusOperatorApiRelabelConfig`

Relabeling applied to each discovered target before it is scraped:
filter targets, or copy discovery labels (`__meta_kubernetes_*`) onto the
series. The operator adds its own steps first; the original job name is
in `__tmp_prometheus_job_name`.

- rule: replace (the default action), hashmod, lowercase, uppercase, keepequal and dropequal write a label, so they need target_label
- rule: hashmod needs a non-zero modulus
- rule: lowercase, uppercase, keepequal and dropequal take no replacement
- rule: keepequal and dropequal take only source_labels and target_label
- rule: labeldrop and labelkeep take only regex

### spec.endpoints[].relabelings[].sourceLabels

`[]string`

The labels whose values are joined (with `separator`) and matched against
`regex`. Target-side relabelings read discovery labels such as
`__meta_kubernetes_pod_label_app`; metric relabelings read the scraped
series' labels, including `__name__`.

### spec.endpoints[].relabelings[].separator

`string` · optional (explicit presence)

The string joining the source labels' values. Unset, ";".

### spec.endpoints[].relabelings[].targetLabel

`string`

The label a replace, hashmod, lowercase, uppercase, keepequal or dropequal
step writes or compares. Regex capture groups are available ("${1}").

### spec.endpoints[].relabelings[].regex

`string`

The RE2 expression the joined source value must match (anchored at both
ends). Empty, "(.*)".

### spec.endpoints[].relabelings[].modulus

`uint32` · optional (explicit presence)

The modulus a hashmod step takes of the hash of the source value; with a
`keep` step on the result it shards targets across Prometheus replicas.
hashmod only, and required there.

### spec.endpoints[].relabelings[].replacement

`string` · optional (explicit presence)

The value a replace step writes, with regex capture groups ("$1",
"${1}-suffix"); a labelmap step's label-name template. Unset, "$1".

### spec.endpoints[].relabelings[].action

`string`

What the step does, case-insensitive: "replace" (the default), "keep",
"drop", "hashmod", "labelmap", "labeldrop", "labelkeep", "lowercase",
"uppercase" (Prometheus >= 2.36), "keepequal" or "dropequal" (Prometheus
>= 2.41). keep and drop filter targets or series; labeldrop and labelkeep
filter label names by regex.

- rule: {"string":{"in":["","replace","Replace","keep","Keep","drop","Drop","hashmod","HashMod","labelmap","LabelMap","labeldrop","LabelDrop","labelkeep","LabelKeep","lowercase","Lowercase","uppercase","Uppercase","keepequal","KeepEqual","dropequal","DropEqual"]}}

### spec.endpoints[].proxyUrl

`string` · optional (explicit presence)

The proxy scrapes go through, "http://", "https://" or "socks5://".
Excludes proxy_from_environment.

- rule: {"string":{"pattern":"^(http|https|socks5)://.+$"}}

### spec.endpoints[].noProxy

`string` · optional (explicit presence)

Comma-separated hosts, domains, IPs or CIDRs (ports allowed) that bypass
proxy_url. Requires proxy_url. Requires Prometheus >= 2.43.

### spec.endpoints[].proxyFromEnvironment

`bool` · optional (explicit presence)

Take the proxy from the Prometheus container's HTTP_PROXY, HTTPS_PROXY and
NO_PROXY environment variables. Excludes proxy_url and no_proxy. Requires
Prometheus >= 2.43.

### spec.endpoints[].proxyConnectHeader

`map<string, KubernetesPrometheusOperatorApiSecretKeySelectorList>`

Headers sent to the proxy on CONNECT, each value read from Secret keys:
`{Proxy-Authorization: {values: [{name: ..., key: ...}]}}` in the
manifest, `{Proxy-Authorization: [...]}` on the object. Requires
proxy_url or proxy_from_environment. Requires Prometheus >= 2.43.

### spec.endpoints[].proxyConnectHeader.*.values

`[]KubernetesPrometheusOperatorApiSecretKeySelector` · required

The Secret keys whose values are sent, one header line each. Upstream
refuses an empty list.

- rule: {"repeated":{"minItems":"1"}}

### spec.endpoints[].proxyConnectHeader.*.values[].name

`string | valueFrom` · required

Name of the Secret, in the monitor's namespace. Defaults to a
KubernetesSecret foreign key (its secret_name output); pass the literal
name with `value:` for a Secret Planton does not manage. Upstream allows an
empty name for backwards compatibility only and calls it "almost
certainly wrong", so a name is required here.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.endpoints[].proxyConnectHeader.*.values[].key

`string` · required

The key within the Secret's data whose value is used ("token",
"password", "ca.crt").

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[-._a-zA-Z0-9]+$"}}

### spec.endpoints[].proxyConnectHeader.*.values[].optional

`bool`

When true, a missing Secret or key is tolerated instead of making the
operator skip the monitor. Leave it false for credentials: a scrape
without its credentials fails anyway, and false makes the operator say so.

### spec.endpoints[].followRedirects

`bool` · optional (explicit presence)

Follow HTTP 3xx redirects. Unset, Prometheus follows them.

### spec.endpoints[].enableHttp2

`bool` · optional (explicit presence)

Allow HTTP/2. Set false for a target whose HTTP/2 support is broken.

### spec.endpoints[].filterRunning

`bool` · optional (explicit presence)

Drop targets whose pod is not running (Failed or Succeeded), so a
finished Job's pod does not read as a down target. Unset, they are
dropped.

### spec.selector

`KubernetesPrometheusOperatorApiLabelSelector` · required

Which Services are scraped, by label, among those in the namespaces
searched. Required. `{}` (match_labels and match_expressions both empty)
selects every Service there -- usually too broad; match the labels the
workload's Service carries.

- rule: {"required":true}

### spec.selector.matchLabels

`map<string, string>`

Labels an object must carry with exactly these values.

### spec.selector.matchExpressions

`[]KubernetesPrometheusOperatorApiLabelSelectorRequirement`

Set-based requirements on an object's labels.

- rule: In and NotIn need at least one value; Exists and DoesNotExist take none

### spec.selector.matchExpressions[].key

`string` · required

The label key the requirement applies to.

- rule: {"string":{"minLen":"1"}}

### spec.selector.matchExpressions[].operator

`string`

How the key relates to the values: "In", "NotIn", "Exists" or
"DoesNotExist".

- rule: {"string":{"in":["In","NotIn","Exists","DoesNotExist"]}}

### spec.selector.matchExpressions[].values

`[]string`

The values for In and NotIn; empty for Exists and DoesNotExist.

### spec.selectorMechanism

`string` · optional (explicit presence)

How Prometheus narrows discovery to the selected Services:
"RelabelConfig" (the default: discover the namespace's endpoints and drop
the unselected by relabeling) or "RoleSelector" (pass the selector to the
Kubernetes API, which is cheaper in large clusters). Requires Prometheus
>= 2.17.

- rule: {"string":{"in":["RelabelConfig","RoleSelector"]}}

### spec.namespaceSelector

`KubernetesPrometheusOperatorApiNamespaceSelector`

The namespaces searched for Services. Unset, only the monitor's own
namespace.

### spec.namespaceSelector.any

`bool`

Search every namespace. Takes precedence over match_names.

### spec.namespaceSelector.matchNames

`[]string | valueFrom`

Search exactly these namespaces. Each is typically a reference to a
KubernetesNamespace resource's `spec.name`; pass a literal with `value:`
for a namespace Planton does not manage.

containment_exempt: these are the namespaces the monitor WATCHES, not the
namespace it lives in (that is the kind's own `namespace`), so a diagram
never nests the monitor inside them.

- references: KubernetesNamespace (`spec.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.sampleLimit

`uint32` · optional (explicit presence)

The most samples one scrape may return; a scrape over it is discarded
whole and the target reads as failed (`up` 0). The guard against a target
that suddenly exports millions of series. Unset, the Prometheus's own
enforced limit applies; 0 means no limit.

### spec.targetLimit

`uint32` · optional (explicit presence)

The most targets this monitor may produce after relabeling; over it,
every target of the monitor fails. Unset or 0, no limit.

### spec.scrapeProtocols

`[]string`

The exposition formats offered to targets, most preferred first:
"PrometheusProto", "OpenMetricsText0.0.1", "OpenMetricsText1.0.0",
"PrometheusText0.0.4", "PrometheusText1.0.0". Native histograms need
"PrometheusProto" first. Unset, Prometheus's default order. Requires
Prometheus >= 2.49.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["PrometheusProto","OpenMetricsText0.0.1","OpenMetricsText1.0.0","PrometheusText0.0.4","PrometheusText1.0.0"]}}}}

### spec.fallbackScrapeProtocol

`string` · optional (explicit presence)

The format assumed when a target answers with a missing or unparseable
Content-Type, one of the scrape_protocols values. Requires Prometheus >=
3.0, which otherwise fails such a scrape.

- rule: {"string":{"in":["PrometheusProto","OpenMetricsText0.0.1","OpenMetricsText1.0.0","PrometheusText0.0.4","PrometheusText1.0.0"]}}

### spec.labelLimit

`uint32` · optional (explicit presence)

The most labels a sample may carry; a scrape with one over it fails.
Requires Prometheus >= 2.27.

### spec.labelNameLengthLimit

`uint32` · optional (explicit presence)

The longest label name a sample may carry. Requires Prometheus >= 2.27.

### spec.labelValueLengthLimit

`uint32` · optional (explicit presence)

The longest label value a sample may carry. Requires Prometheus >= 2.27.

### spec.scrapeNativeHistograms

`bool` · optional (explicit presence)

Scrape native histograms. Requires Prometheus >= 3.8 and
"PrometheusProto" (or OpenMetrics 1.0 with native histograms) among
scrape_protocols.

### spec.scrapeClassicHistograms

`bool` · optional (explicit presence)

Keep scraping the classic (bucketed) form of a histogram that is also
exposed as a native histogram (Prometheus's
`always_scrape_classic_histograms`). Requires Prometheus >= 2.45.

### spec.nativeHistogramBucketLimit

`uint32` · optional (explicit presence)

The most buckets a native histogram may keep; above it, neighbouring
buckets merge. Requires Prometheus >= 2.45.

### spec.nativeHistogramMinBucketFactor

`string` · optional (explicit presence)

The smallest growth factor between neighbouring native-histogram buckets;
buckets closer than this merge. A decimal or Kubernetes quantity ("1.1",
"2"). Requires Prometheus >= 2.50.

- rule: {"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}

### spec.convertClassicHistogramsToNHCB

`bool` · optional (explicit presence)

Convert scraped classic histograms into native histograms with custom
buckets. Requires Prometheus >= 3.0.

### spec.keepDroppedTargets

`uint32` · optional (explicit presence)

How many targets dropped by relabeling the Prometheus keeps in memory for
its targets page. 0 means no limit. Requires Prometheus >= 2.47.

### spec.attachMetadata

`KubernetesPrometheusOperatorApiAttachMetadata`

Metadata of related objects attached to the targets as discovery labels.

### spec.attachMetadata.node

`bool` · optional (explicit presence)

Attach the node's metadata, as `__meta_kubernetes_node_*` labels that
relabelings can copy onto the series (the node's zone or instance type).
Prometheus's service account needs list and watch on Nodes.

### spec.scrapeClass

`string` · required · optional (explicit presence)

The Prometheus scrape class whose defaults (TLS, authorization,
relabelings) this monitor inherits, by name. The class is declared on the
Prometheus; the operator skips a monitor that names a class its
Prometheus lacks. Unset, the Prometheus's default class (if any).

- rule: {"string":{"minLen":"1"}}

### spec.bodySizeLimit

`string` · optional (explicit presence)

The largest uncompressed response body a scrape accepts ("10MiB",
"512KB"; "0" for no limit); a larger body fails the scrape. Requires
Prometheus >= 2.28.

- rule: {"string":{"pattern":"(^0|([0-9]*[.])?[0-9]+((K|M|G|T|E|P)i?)?B)$"}}

### spec.serviceDiscoveryRole

`string` · optional (explicit presence)

The Kubernetes API Prometheus discovers the endpoints through:
"Endpoints" or "EndpointSlice" (which scales to large Services and is the
API Kubernetes keeps). Unset, the Prometheus's own setting.

- rule: {"string":{"in":["Endpoints","EndpointSlice"]}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesServiceMonitor, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.service_monitor_name` | `string` | Name of the created ServiceMonitor (equals metadata.name). |
| `status.outputs.namespace` | `string` | Namespace the ServiceMonitor was created in (the resolved spec.namespace). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.endpoints[].tlsConfig.ca.secret.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.endpoints[].tlsConfig.ca.configMap.name` | KubernetesConfigMap | `status.outputs.configmap_name` |
| `spec.endpoints[].tlsConfig.cert.secret.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.endpoints[].tlsConfig.cert.configMap.name` | KubernetesConfigMap | `status.outputs.configmap_name` |
| `spec.endpoints[].tlsConfig.keySecret.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.endpoints[].bearerTokenSecret.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.endpoints[].authorization.credentials.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.endpoints[].basicAuth.username.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.endpoints[].basicAuth.password.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.endpoints[].oauth2.clientId.secret.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.endpoints[].oauth2.clientId.configMap.name` | KubernetesConfigMap | `status.outputs.configmap_name` |
| `spec.endpoints[].oauth2.clientSecret.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.endpoints[].oauth2.tlsConfig.ca.secret.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.endpoints[].oauth2.tlsConfig.ca.configMap.name` | KubernetesConfigMap | `status.outputs.configmap_name` |
| `spec.endpoints[].oauth2.tlsConfig.cert.secret.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.endpoints[].oauth2.tlsConfig.cert.configMap.name` | KubernetesConfigMap | `status.outputs.configmap_name` |
| `spec.endpoints[].oauth2.tlsConfig.keySecret.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.endpoints[].oauth2.proxyConnectHeader.*.values[].name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.endpoints[].proxyConnectHeader.*.values[].name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.namespaceSelector.matchNames` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)
