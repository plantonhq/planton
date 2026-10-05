# KubernetesIngress

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesIngressSpec** declares HTTP(S) exposure for in-cluster services as a
first-class networking/v1 Ingress object: host rules and path matches routing to
Service backends, with optional TLS termination from certificate Secrets.

Exposure in Planton is composed, never embedded: a workload kind exports its
Service name (`service` output), this kind routes a hostname to that Service,
and a certificate kind materializes the TLS Secret the `tls` block names. Every
piece of the exposure path is a visible, independently-managed node in the
resource graph.

An Ingress object is inert until an ingress controller (e.g. ingress-nginx,
cloud L7 controllers) runs in the cluster and claims it via `ingress_class_name`
or a default IngressClass. Creating the Ingress before the controller exists is
valid — its load-balancer status stays empty until a controller reconciles it.
IngressClass objects themselves ship with their controllers, which is why the
class is a plain name here rather than a reference to a Planton kind.

The spec covers the complete networking/v1 IngressSpec surface. The single
deliberate omission is the `resource` backend variant (an ObjectRef to an
arbitrary same-namespace object, e.g. a static-asset bucket CRD) — it is
controller-specific and rarely implemented; Service backends cover the real
exposure paths. Controller-specific behavior (rewrites, timeouts, body sizes,
auth) is configured through `annotations`, which is the upstream contract.

## Example

```yaml
# Full-surface test manifest for the offline plan proofs. Exercises the
# ingress class, a default backend, TLS blocks (named-secret and SNI-only),
# host and wildcard-host rules, all three path types, and both backend port
# forms (number and name).
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesIngress
metadata:
  name: test-ingress
spec:
  namespace:
    value: default
  name: test-ingress
  labels:
    team: platform-engineering
  annotations:
    nginx.ingress.kubernetes.io/proxy-body-size: 50m
    cert-manager.io/cluster-issuer: letsencrypt-prod
  ingress_class_name: nginx
  default_backend:
    service_name:
      value: fallback-svc
    port_number: 80
  tls:
    - hosts:
        - app.example.com
      secret_name:
        value: app-example-com-tls
    - hosts:
        - "*.preview.example.com"
  rules:
    - host: app.example.com
      paths:
        - path: /
          path_type: prefix
          backend:
            service_name:
              value: app-svc
            port_number: 8080
        - path: /api
          path_type: exact
          backend:
            service_name:
              value: api-svc
            port_name: http
    - host: "*.preview.example.com"
      paths:
        - path: /
          path_type: implementation_specific
          backend:
            service_name:
              value: preview-router
            port_number: 80
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` |  |  | KubernetesNamespace (`spec.name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.ingressClassName` | `string` |  |  |  |
| `spec.defaultBackend` | `KubernetesIngressBackend` |  |  |  |
| `spec.defaultBackend.serviceName` | `string \| valueFrom` | yes |  | KubernetesService (`status.outputs.service_name`) |
| `spec.defaultBackend.portNumber` | `int32` |  |  |  |
| `spec.defaultBackend.portName` | `string` |  |  |  |
| `spec.tls` | `[]KubernetesIngressTls` |  |  |  |
| `spec.tls[].hosts` | `[]string` |  |  |  |
| `spec.tls[].secretName` | `string \| valueFrom` |  |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.rules` | `[]KubernetesIngressRule` |  |  |  |
| `spec.rules[].host` | `string` |  |  |  |
| `spec.rules[].paths` | `[]KubernetesIngressHttpPath` | yes |  |  |
| `spec.rules[].paths[].path` | `string` |  |  |  |
| `spec.rules[].paths[].pathType` | `enum` |  |  |  |
| `spec.rules[].paths[].backend` | `KubernetesIngressBackend` | yes |  |  |
| `spec.rules[].paths[].backend.serviceName` | `string \| valueFrom` | yes |  | KubernetesService (`status.outputs.service_name`) |
| `spec.rules[].paths[].backend.portNumber` | `int32` |  |  |  |
| `spec.rules[].paths[].backend.portName` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom`

The namespace to create the Ingress in — it must be the namespace of the
backend Services, because Ingress backends can only reference Services in
their own namespace (a Kubernetes API constraint). Accepts a literal
namespace name or a reference to a KubernetesNamespace resource. When
omitted, the Ingress lands in the cluster's `default` namespace.

- references: KubernetesNamespace (`spec.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.name

`string` · required

The name of the Ingress (its `metadata.name` in the cluster).
Must be a valid DNS subdomain: lowercase alphanumeric characters, hyphens,
and dots, at most 253 characters.

- rule: Name must be a valid DNS subdomain (lowercase alphanumeric, hyphens, and dots, no leading/trailing dots or hyphens)
- rule: {"string":{"minLen":"1","maxLen":"253"}}

### spec.labels

`map<string, string>`

Additional labels to apply to the Ingress object.
Merged with the standard Planton governance labels.

### spec.annotations

`map<string, string>`

Annotations on the Ingress object — the upstream contract for
controller-specific behavior. Common ingress-nginx examples:
- `nginx.ingress.kubernetes.io/rewrite-target: /$2` — path rewriting
- `nginx.ingress.kubernetes.io/proxy-body-size: 50m` — upload size limit
- `nginx.ingress.kubernetes.io/ssl-redirect: "false"` — disable HTTPS redirect
- `cert-manager.io/cluster-issuer: letsencrypt-prod` — cert-manager issues the
  TLS certificate into the Secret named by the `tls` block
- `external-dns.alpha.kubernetes.io/hostname: app.example.com` — DNS record

### spec.ingressClassName

`string`

The IngressClass that selects which controller serves this Ingress (e.g.
"nginx", "alb", "gce"). Classes are cluster-scoped objects installed with
their controllers — `kubectl get ingressclass` lists what the cluster
offers. Omit to use the cluster's default class (the IngressClass annotated
`ingressclass.kubernetes.io/is-default-class: "true"`); on clusters without
a default, an Ingress with no class is not served by any controller.

- rule: ingress_class_name must be a valid DNS subdomain (lowercase alphanumeric, hyphens, and dots)
- rule: {"string":{"maxLen":"253"}}

### spec.defaultBackend

`KubernetesIngressBackend`

The backend that handles requests no rule matches. Also the way to expose a
single Service on ALL traffic the controller routes here: set only
default_backend and omit rules. When rules are present, most controllers
also fall back to their own global default backend if this is unset.

- rule: exactly one of port_number or port_name must be set on a backend

### spec.defaultBackend.serviceName

`string | valueFrom` · required

The name of the backend Service. Accepts a literal Service name or a
reference to a KubernetesService resource — and in an infra chart this is
where a workload's exported `service` output is wired in (via valueFrom), so
"deploy the app, then expose it" composes without copying names by hand.

- references: KubernetesService (`status.outputs.service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesService, name: <that resource's name>, fieldPath: status.outputs.service_name}} -- a bare string does not parse

### spec.defaultBackend.portNumber

`int32`

The Service port to route to, by number. Exactly one of `port_number` or
`port_name` is required.

- rule: port_number must be in the range 1-65535

### spec.defaultBackend.portName

`string`

The Service port to route to, by name (the `name` of a port on the Service,
e.g. "http"). Prefer names when the Service defines them — the reference
survives port-number changes. Exactly one of `port_number` or `port_name`
is required.

- rule: port_name must be a valid IANA service name: lowercase alphanumeric and hyphens, at most 15 characters, containing at least one letter (e.g. "http")

### spec.tls

`[]KubernetesIngressTls`

TLS termination configuration. Each entry names the hosts served under one
certificate Secret; the controller multiplexes them on port 443 via SNI.

### spec.tls[].hosts

`[]string`

The hosts served under this certificate — they must appear in the
certificate's SANs and should match hosts used in `rules`. The controller
multiplexes multiple TLS entries on port 443 via SNI. Wildcard entries
("*.example.com") follow the same single-label semantics as rule hosts.

- rule: {"repeated":{"items":{"cel":[{"id":"tls_host.precise_or_wildcard","message":"each TLS host must be a lowercase DNS name, optionally with a single leading wildcard label (e.g. \"app.example.com\" or \"*.example.com\")","expression":"this.matches('^(\\\\*\\\\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$')"}]}}}

### spec.tls[].secretName

`string | valueFrom`

The name of the `kubernetes.io/tls` Secret (in the Ingress's namespace)
holding the certificate and key. Accepts a literal Secret name or a
reference to a KubernetesSecret resource. With cert-manager, leave the
Secret non-existent and add the issuer annotation — cert-manager creates
this Secret for you, named exactly as written here. Optional: omitting it
asks the controller to serve these hosts by SNI with its default (or
separately-provisioned) certificate.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.rules

`[]KubernetesIngressRule`

Host rules routing requests to backends: a request is first matched on its
Host header, then on the rule's HTTP paths. Either `rules` or
`default_backend` must be present (an Ingress with neither routes nothing —
the API accepts it but it is always a mistake, so validation rejects it).

### spec.rules[].host

`string`

The fully qualified domain name this rule serves, matched against the
request's Host header. Either a precise host ("app.example.com") or a
wildcard whose FIRST label is "*" ("*.example.com" — matches exactly one
label, so it matches "a.example.com" but not "b.a.example.com" or
"example.com"). Omit to match every host reaching the controller. Ports and
IPs are not allowed.

- rule: host must be a lowercase DNS name, optionally with a single leading wildcard label (e.g. "app.example.com" or "*.example.com"); IPs and ports are not allowed
- rule: {"string":{"maxLen":"253"}}

### spec.rules[].paths

`[]KubernetesIngressHttpPath` · required

The HTTP paths served for this host. At least one path is required — a rule
with a host but no paths routes nothing (upstream treats a missing
IngressRuleValue as "controller-defined", which in practice means ignored).

- rule: {"repeated":{"minItems":"1"}}
- rule: path is required when path_type is prefix or exact

### spec.rules[].paths[].path

`string`

The URL path to match, beginning with "/". Required for Exact and Prefix
path types. With `prefix`, matching is per path element: "/api" matches
"/api" and "/api/users" but not "/apiary". With `exact`, the path must match
exactly and case-sensitively. May be omitted only for
implementation_specific paths (the controller decides).

- rule: path must begin with '/' (e.g. "/", "/api")
- rule: {"string":{"maxLen":"1024"}}

### spec.rules[].paths[].pathType

`enum` · optional (explicit presence)

How the path is matched.
Default: prefix

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `kubernetes_ingress_path_type_unspecified` -- Unspecified. Defaults to prefix — the match semantics users almost always want, and the one every controller must implement identically.
- `prefix` -- Prefix: match on URL path prefix, split by "/" per path element. The longest matching path wins when multiple paths match.
- `exact` -- Exact: match the URL path exactly, case-sensitively.
- `implementation_specific` -- ImplementationSpecific: matching semantics are delegated to the IngressClass — e.g. ingress-nginx treats these paths as regex candidates. Non-portable across controllers by definition.

### spec.rules[].paths[].backend

`KubernetesIngressBackend` · required

The Service backend receiving the matched traffic.

- rule: {"required":true}
- rule: exactly one of port_number or port_name must be set on a backend

### spec.rules[].paths[].backend.serviceName

`string | valueFrom` · required

The name of the backend Service. Accepts a literal Service name or a
reference to a KubernetesService resource — and in an infra chart this is
where a workload's exported `service` output is wired in (via valueFrom), so
"deploy the app, then expose it" composes without copying names by hand.

- references: KubernetesService (`status.outputs.service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesService, name: <that resource's name>, fieldPath: status.outputs.service_name}} -- a bare string does not parse

### spec.rules[].paths[].backend.portNumber

`int32`

The Service port to route to, by number. Exactly one of `port_number` or
`port_name` is required.

- rule: port_number must be in the range 1-65535

### spec.rules[].paths[].backend.portName

`string`

The Service port to route to, by name (the `name` of a port on the Service,
e.g. "http"). Prefer names when the Service defines them — the reference
survives port-number changes. Exactly one of `port_number` or `port_name`
is required.

- rule: port_name must be a valid IANA service name: lowercase alphanumeric and hyphens, at most 15 characters, containing at least one letter (e.g. "http")

## Validation Rules

- `rules_or_default_backend_required`: at least one rule or a default_backend must be specified — an Ingress with neither routes no traffic

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesIngress, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.ingress_name` | `string` | The name of the Ingress object as created in the cluster. |
| `status.outputs.namespace` | `string` | The namespace the Ingress was created in. |
| `status.outputs.load_balancer_ip` | `string` | The IP address the controller's load balancer exposes for this Ingress. Populated by IP-based controllers (GCE, ingress-nginx on most clouds); empty until a controller reconciles the Ingress. |
| `status.outputs.load_balancer_hostname` | `string` | The DNS hostname the controller's load balancer exposes for this Ingress. Populated by hostname-based controllers (AWS ALB/ELB); empty until a controller reconciles the Ingress. |
| `status.outputs.first_host` | `string` | The first host declared in the Ingress rules — the primary public FQDN this Ingress serves, ready for downstream references (dashboards, smoke tests, DNS records). Empty when the Ingress only declares host-less catch-all rules or only a default backend. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.defaultBackend.serviceName` | KubernetesService | `status.outputs.service_name` |
| `spec.tls[].secretName` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.rules[].paths[].backend.serviceName` | KubernetesService | `status.outputs.service_name` |

## See Also

- [Overview](../README.md)
