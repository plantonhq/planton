# Token-Protected Endpoint

Scrapes a metrics endpoint that requires a bearer token and serves HTTPS. The token comes from a KubernetesSecret, referenced so the pipeline creates it before the monitor. The target's certificate is verified against the CA bundle that issued it. Use it for a service that exposes its metrics only to authenticated callers.

## When to Use

- The metrics endpoint answers 401 without a token, or is served only over TLS.
- The token is managed as a KubernetesSecret in the same namespace (created by the chart, or synced from a secret store).

## How It Works

`authorization` sends `Authorization: Bearer <token>` with every scrape, reading the token from the referenced Secret's `token` key. The Secret is a reference (`valueFrom`), so the Infra Chart creates it first and the diagram shows the dependency. With a typed name, a monitor applied before its Secret is skipped by the operator until the next reconcile. `scheme: https` with `tls_config.ca` verifies the target against the issuing CA; `server_name` is the name on the target's certificate, because Prometheus dials the pod IP, which no certificate carries.

## Key Configuration Choices

- **`authorization`, not `basic_auth` or `bearer_token_secret`.** It is upstream's recommended form; the spec refuses a second authentication method on the same endpoint.
- **The CA from a ConfigMap.** A CA bundle is public material; the Secret form is for anything private.
- **No `insecure_skip_verify`.** It keeps encryption but lets anyone on the path impersonate the target.

## Prerequisites

- A KubernetesSecret `<token_secret>` with a `token` key, and a ConfigMap `<ca_configmap>` with a `ca.crt` key, both in the monitor's namespace.
- A Service labelled `app.kubernetes.io/name: <service>` whose named port serves HTTPS.
- A Prometheus that selects this object (`KubernetesKubePrometheusStack`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<service>` | The service's `app.kubernetes.io/name`, also used as the monitor's name and in the certificate name. |
| `<namespace>` | Namespace of the monitor, the Service, the Secret and the ConfigMap. |
| `<metrics_port_name>` | The Service port's name. |
| `<token_secret>` | The KubernetesSecret resource holding the bearer token under `token`. |
| `<ca_configmap>` | The ConfigMap holding the issuing CA under `ca.crt`. |
