# KubernetesGateway

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

KubernetesGatewaySpec defines a Kubernetes Gateway API Gateway: a
namespaced instance of traffic-handling infrastructure that binds a set of
Listeners (logical endpoints with a port, protocol, and optional TLS) to
addresses, managed by the controller behind a GatewayClass.

100% fidelity with the upstream Gateway API v1.6.1 GatewaySpec
(kubernetes-sigs/gateway-api apis/v1/gateway_types.go), standard channel.
Upstream spec fields are flattened after the Planton namespaced envelope
(namespace). The experimental `defaultScope` field is
intentionally excluded: it is absent from the standard-channel CRD and from
the typed Pulumi resource Planton provisions with.

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGateway
metadata:
  name: test-gateway
spec:
  namespace:
    value: test-namespace
  gatewayClassName:
    value: istio
  listeners:
    - name: http
      port: 80
      protocol: HTTP
      allowedRoutes:
        namespaces:
          from: All
        kinds:
          - kind: HTTPRoute
    - name: https
      hostname: app.example.com
      port: 443
      protocol: HTTPS
      tls:
        mode: Terminate
        certificateRefs:
          - name:
              value: app-tls-cert
        options:
          example.com/minimum-tls-version: "1.3"
      allowedRoutes:
        namespaces:
          from: Selector
          selector:
            matchLabels:
              team: web
            matchExpressions:
              - key: environment
                operator: In
                values:
                  - staging
                  - production
    - name: tls-passthrough
      hostname: db.example.com
      port: 8443
      protocol: TLS
      tls:
        mode: Passthrough
  addresses:
    - type: IPAddress
      value: 203.0.113.10
    - type: Hostname
      value: gw.example.com
  infrastructure:
    labels:
      example.com/cost-center: platform
    annotations:
      service.beta.kubernetes.io/aws-load-balancer-type: nlb
    parametersRef:
      group: example.com
      kind: GatewayParams
      name: test-gateway-params
  allowedListeners:
    namespaces:
      from: Same
  tls:
    backend:
      clientCertificateRef:
        name:
          value: backend-client-cert
    frontend:
      default:
        validation:
          caCertificateRefs:
            - group: ""
              kind: ConfigMap
              name:
                value: client-ca-bundle
          mode: AllowValidOnly
      perPort:
        - port: 443
          tls:
            validation:
              caCertificateRefs:
                - group: ""
                  kind: ConfigMap
                  name:
                    value: strict-ca-bundle
              mode: AllowValidOnly
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.gatewayClassName` | `string \| valueFrom` | yes |  | KubernetesGatewayClass (`status.outputs.gateway_class_name`) |
| `spec.listeners` | `[]KubernetesGatewayListener` | yes |  |  |
| `spec.listeners[].name` | `string` | yes |  |  |
| `spec.listeners[].hostname` | `string` | yes |  |  |
| `spec.listeners[].port` | `int32` | yes |  |  |
| `spec.listeners[].protocol` | `string` | yes |  |  |
| `spec.listeners[].tls` | `KubernetesGatewayApiListenerTlsConfig` |  |  |  |
| `spec.listeners[].tls.mode` | `string` |  |  |  |
| `spec.listeners[].tls.certificateRefs` | `[]KubernetesGatewayApiSecretObjectReference` |  |  |  |
| `spec.listeners[].tls.certificateRefs[].group` | `string` |  |  |  |
| `spec.listeners[].tls.certificateRefs[].kind` | `string` | yes |  |  |
| `spec.listeners[].tls.certificateRefs[].name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.listeners[].tls.certificateRefs[].namespace` | `string` |  |  |  |
| `spec.listeners[].tls.options` | `map<string, string>` |  |  |  |
| `spec.listeners[].allowedRoutes` | `KubernetesGatewayApiAllowedRoutes` |  |  |  |
| `spec.listeners[].allowedRoutes.namespaces` | `KubernetesGatewayApiRouteNamespaces` |  |  |  |
| `spec.listeners[].allowedRoutes.namespaces.from` | `string` |  |  |  |
| `spec.listeners[].allowedRoutes.namespaces.selector` | `KubernetesGatewayApiLabelSelector` |  |  |  |
| `spec.listeners[].allowedRoutes.namespaces.selector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.listeners[].allowedRoutes.namespaces.selector.matchExpressions` | `[]KubernetesGatewayApiLabelSelectorRequirement` |  |  |  |
| `spec.listeners[].allowedRoutes.namespaces.selector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.listeners[].allowedRoutes.namespaces.selector.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.listeners[].allowedRoutes.namespaces.selector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.listeners[].allowedRoutes.kinds` | `[]KubernetesGatewayApiRouteGroupKind` |  |  |  |
| `spec.listeners[].allowedRoutes.kinds[].group` | `string` |  |  |  |
| `spec.listeners[].allowedRoutes.kinds[].kind` | `string` | yes |  |  |
| `spec.addresses` | `[]KubernetesGatewayAddress` |  |  |  |
| `spec.addresses[].type` | `string` | yes |  |  |
| `spec.addresses[].value` | `string` |  |  |  |
| `spec.infrastructure` | `KubernetesGatewayInfrastructure` |  |  |  |
| `spec.infrastructure.labels` | `map<string, string>` |  |  |  |
| `spec.infrastructure.annotations` | `map<string, string>` |  |  |  |
| `spec.infrastructure.parametersRef` | `KubernetesGatewayLocalParametersReference` |  |  |  |
| `spec.infrastructure.parametersRef.group` | `string` | yes |  |  |
| `spec.infrastructure.parametersRef.kind` | `string` | yes |  |  |
| `spec.infrastructure.parametersRef.name` | `string` | yes |  |  |
| `spec.allowedListeners` | `KubernetesGatewayAllowedListeners` |  |  |  |
| `spec.allowedListeners.namespaces` | `KubernetesGatewayListenerNamespaces` |  |  |  |
| `spec.allowedListeners.namespaces.from` | `string` |  |  |  |
| `spec.allowedListeners.namespaces.selector` | `KubernetesGatewayApiLabelSelector` |  |  |  |
| `spec.allowedListeners.namespaces.selector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.allowedListeners.namespaces.selector.matchExpressions` | `[]KubernetesGatewayApiLabelSelectorRequirement` |  |  |  |
| `spec.allowedListeners.namespaces.selector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.allowedListeners.namespaces.selector.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.allowedListeners.namespaces.selector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.tls` | `KubernetesGatewayTlsConfig` |  |  |  |
| `spec.tls.backend` | `KubernetesGatewayBackendTls` |  |  |  |
| `spec.tls.backend.clientCertificateRef` | `KubernetesGatewayApiSecretObjectReference` |  |  |  |
| `spec.tls.backend.clientCertificateRef.group` | `string` |  |  |  |
| `spec.tls.backend.clientCertificateRef.kind` | `string` | yes |  |  |
| `spec.tls.backend.clientCertificateRef.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.tls.backend.clientCertificateRef.namespace` | `string` |  |  |  |
| `spec.tls.frontend` | `KubernetesGatewayFrontendTlsConfig` |  |  |  |
| `spec.tls.frontend.default` | `KubernetesGatewayFrontendTlsValidationConfig` | yes |  |  |
| `spec.tls.frontend.default.validation` | `KubernetesGatewayFrontendTlsValidation` |  |  |  |
| `spec.tls.frontend.default.validation.caCertificateRefs` | `[]KubernetesGatewayApiObjectReference` | yes |  |  |
| `spec.tls.frontend.default.validation.caCertificateRefs[].group` | `string` | yes |  |  |
| `spec.tls.frontend.default.validation.caCertificateRefs[].kind` | `string` | yes |  |  |
| `spec.tls.frontend.default.validation.caCertificateRefs[].name` | `string \| valueFrom` | yes |  | KubernetesConfigMap (`status.outputs.configmap_name`) |
| `spec.tls.frontend.default.validation.caCertificateRefs[].namespace` | `string` |  |  |  |
| `spec.tls.frontend.default.validation.mode` | `string` |  |  |  |
| `spec.tls.frontend.perPort` | `[]KubernetesGatewayTlsPortConfig` |  |  |  |
| `spec.tls.frontend.perPort[].port` | `int32` | yes |  |  |
| `spec.tls.frontend.perPort[].tls` | `KubernetesGatewayFrontendTlsValidationConfig` | yes |  |  |
| `spec.tls.frontend.perPort[].tls.validation` | `KubernetesGatewayFrontendTlsValidation` |  |  |  |
| `spec.tls.frontend.perPort[].tls.validation.caCertificateRefs` | `[]KubernetesGatewayApiObjectReference` | yes |  |  |
| `spec.tls.frontend.perPort[].tls.validation.caCertificateRefs[].group` | `string` | yes |  |  |
| `spec.tls.frontend.perPort[].tls.validation.caCertificateRefs[].kind` | `string` | yes |  |  |
| `spec.tls.frontend.perPort[].tls.validation.caCertificateRefs[].name` | `string \| valueFrom` | yes |  | KubernetesConfigMap (`status.outputs.configmap_name`) |
| `spec.tls.frontend.perPort[].tls.validation.caCertificateRefs[].namespace` | `string` |  |  |  |
| `spec.tls.frontend.perPort[].tls.validation.mode` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace in which the Gateway is created. Routes that attach to this
Gateway, and the TLS Secrets its listeners reference, are subject to the
usual same-namespace / ReferenceGrant rules.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.gatewayClassName

`string | valueFrom` · required

Name of the GatewayClass this Gateway belongs to. The GatewayClass selects
the controller (Istio, Envoy Gateway, NGINX, ...) that programs this
Gateway. As a foreign key it points at a KubernetesGatewayClass output so
InfraChart DAGs deploy the class before the Gateway.

- references: KubernetesGatewayClass (`status.outputs.gateway_class_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesGatewayClass, name: <that resource's name>, fieldPath: status.outputs.gateway_class_name}} -- a bare string does not parse

### spec.listeners

`[]KubernetesGatewayListener` · required

Listeners are the logical endpoints bound on this Gateway's addresses.
Each listener defines a port, protocol, and optional TLS and hostname.
At least one listener is required.

- rule: {"repeated":{"minItems":"1","maxItems":"64"}}
- rule: tls must not be set when protocol is HTTP, TCP, or UDP
- rule: tls mode must be Terminate (or left unset) when protocol is HTTPS
- rule: tls and tls.mode must be set when protocol is TLS
- rule: hostname must not be set when protocol is TCP or UDP

### spec.listeners[].name

`string` · required

Name of the listener; must be unique within the Gateway. Routes attach to
a specific listener by this name (parentRef.sectionName).

Upstream SectionName constraints: 1-253 chars; lowercase RFC 1123
subdomain. Pattern: ^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$

- rule: {"required":true,"string":{"minLen":"1","maxLen":"253","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"}}

### spec.listeners[].hostname

`string` · required · optional (explicit presence)

Virtual hostname to match for protocols that support it (HTTP, HTTPS, TLS).
When unset, all hostnames match. A leading wildcard label (for example
"*.example.com") is a suffix match. Ignored for TCP and UDP.

Upstream Hostname constraints: 1-253 chars. Pattern (wildcard prefix
allowed): ^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^(\\*\\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"}}

### spec.listeners[].port

`int32` · required

Network port this listener binds. Multiple listeners may share a port
subject to the distinctness rules above.

- rule: {"required":true,"int32":{"lte":65535,"gte":1}}

### spec.listeners[].protocol

`string` · required

Application protocol the listener accepts. Core values are HTTP, HTTPS,
TLS, TCP, and UDP. Implementation-specific protocols use a domain-prefixed
path such as "example.com/my-protocol".

Upstream ProtocolType is an open set validated by pattern (not a closed
enum), so custom domain-prefixed protocols remain valid.
Pattern: ^[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?$|<domain>/<path>

- rule: {"required":true,"string":{"minLen":"1","maxLen":"255","pattern":"^[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?$|[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\\/[A-Za-z0-9]+$"}}

### spec.listeners[].tls

`KubernetesGatewayApiListenerTlsConfig`

TLS configuration for this listener. Required for HTTPS and TLS protocols;
must be absent for HTTP, TCP, and UDP (enforced above). Certificate
references default to KubernetesSecret foreign keys — wire valueFrom
against a KubernetesCertificate's status.outputs.secret_name to terminate
with cert-manager-issued material. Shared Gateway API type (also used by
KubernetesListenerSet listener entries).

- rule: certificate_refs or options must be provided when tls mode is Terminate (the default)

### spec.listeners[].tls.mode

`string` · optional (explicit presence)

How the listener handles the client's TLS session. "Terminate" (default)
decrypts at the Gateway and requires a certificate. "Passthrough" forwards
the encrypted stream untouched (TLSRoute only) and ignores certificate_refs.

Upstream default: "Terminate". Closed enum: Terminate | Passthrough.

- rule: tls mode must be either 'Terminate' or 'Passthrough'

### spec.listeners[].tls.certificateRefs

`[]KubernetesGatewayApiSecretObjectReference`

References to Kubernetes Secrets holding the TLS certificate/key used to
terminate the listener. A single reference to a kubernetes.io/tls Secret
has Core support; multiple references are implementation-specific.

Each reference's name is a KubernetesSecret foreign key; the Secret is
typically produced by a KubernetesCertificate (cert-manager) — wire
valueFrom against the Certificate's status.outputs.secret_name so the
listener terminates with the issued certificate. See
KubernetesGatewayApiSecretObjectReference.

- rule: {"repeated":{"maxItems":"64"}}

### spec.listeners[].tls.certificateRefs[].group

`string` · optional (explicit presence)

Group of the referent. Empty string infers the core API group.

Upstream default: "" (core API group)
Group pattern: empty or an RFC 1123 subdomain (max 253).

- rule: {"string":{"maxLen":"253","pattern":"^$|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"}}

### spec.listeners[].tls.certificateRefs[].kind

`string` · required · optional (explicit presence)

Kind of the referent.

Upstream default: "Secret"
Kind pattern: 1-63 chars, ^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$

- rule: {"string":{"minLen":"1","maxLen":"63","pattern":"^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$"}}

### spec.listeners[].tls.certificateRefs[].name

`string | valueFrom` · required

Name of the referent. Defaults to a KubernetesSecret foreign key. The
Secret is typically produced by cert-manager: reference a
KubernetesCertificate's status.outputs.secret_name with valueFrom to wire
a certificate the moment it is issued, or a KubernetesSecret directly for
externally provisioned material. Pass the literal name with `value:` when
the Secret is not Planton-managed.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.listeners[].tls.certificateRefs[].namespace

`string` · optional (explicit presence)

Namespace of the referenced object. When unspecified, the local namespace
is inferred. Cross-namespace references require a ReferenceGrant.

### spec.listeners[].tls.options

`map<string, string>`

Implementation-specific TLS options (for example minimum TLS version or
cipher suites). Keys should be domain-prefixed to avoid ambiguity.

Upstream key/value: Gateway API AnnotationKey (1-253) / AnnotationValue
(0-4096).

- rule: {"map":{"maxPairs":"16","keys":{"string":{"minLen":"1","maxLen":"253","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?([A-Za-z0-9][-A-Za-z0-9_.]{0,61})?[A-Za-z0-9]$"}},"values":{"string":{"maxLen":"4096"}}}}

### spec.listeners[].allowedRoutes

`KubernetesGatewayApiAllowedRoutes`

Restricts which Routes (by kind and namespace) may attach to this
listener. When unset, the controller selects routes by listener protocol
and the same namespace as the Gateway. Shared Gateway API type (also used
by KubernetesListenerSet listener entries).

### spec.listeners[].allowedRoutes.namespaces

`KubernetesGatewayApiRouteNamespaces`

Namespaces from which Routes may attach. Defaults to the Gateway's own
namespace.

### spec.listeners[].allowedRoutes.namespaces.from

`string` · optional (explicit presence)

Where Routes are selected from: "All" (any namespace), "Selector"
(namespaces matching the selector), or "Same" (the Gateway's namespace).

Upstream default: "Same". Closed enum: All | Selector | Same.

- rule: from must be one of 'All', 'Selector', or 'Same'

### spec.listeners[].allowedRoutes.namespaces.selector

`KubernetesGatewayApiLabelSelector`

Namespace label selector. Required (and only honored) when `from` is
"Selector".

### spec.listeners[].allowedRoutes.namespaces.selector.matchLabels

`map<string, string>`

Map of {key,value} pairs. A namespace matches if it carries all of these
labels.

### spec.listeners[].allowedRoutes.namespaces.selector.matchExpressions

`[]KubernetesGatewayApiLabelSelectorRequirement`

List of label selector requirements, ANDed together with match_labels.

### spec.listeners[].allowedRoutes.namespaces.selector.matchExpressions[].key

`string` · required

Label key the requirement applies to.

- rule: {"required":true}

### spec.listeners[].allowedRoutes.namespaces.selector.matchExpressions[].operator

`string` · required

Relationship between the key and values: "In", "NotIn", "Exists", or
"DoesNotExist". For "Exists"/"DoesNotExist", values must be empty.

- rule: operator must be one of 'In', 'NotIn', 'Exists', or 'DoesNotExist'
- rule: {"required":true}

### spec.listeners[].allowedRoutes.namespaces.selector.matchExpressions[].values

`[]string`

Values for the requirement. Must be non-empty for "In"/"NotIn" and empty
for "Exists"/"DoesNotExist".

### spec.listeners[].allowedRoutes.kinds

`[]KubernetesGatewayApiRouteGroupKind`

Route kinds allowed to bind to this listener. When empty, allowed kinds
are inferred from the listener protocol.

- rule: {"repeated":{"maxItems":"8"}}

### spec.listeners[].allowedRoutes.kinds[].group

`string` · optional (explicit presence)

API group of the Route. Empty string selects the core API group.

Upstream default: "gateway.networking.k8s.io".
Group pattern: empty or an RFC 1123 subdomain (max 253).

- rule: {"string":{"maxLen":"253","pattern":"^$|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"}}

### spec.listeners[].allowedRoutes.kinds[].kind

`string` · required

Kind of the Route (for example "HTTPRoute", "GRPCRoute", "TLSRoute").

Kind pattern: 1-63 chars, ^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$

- rule: {"required":true,"string":{"minLen":"1","maxLen":"63","pattern":"^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$"}}

### spec.addresses

`[]KubernetesGatewayAddress`

Addresses requested for this Gateway (for example a specific static IP).
Optional; when omitted the controller assigns addresses. Behavior is
implementation-specific. Upstream support level: Extended.

- rule: {"repeated":{"maxItems":"16"}}
- rule: a Hostname address value must be a valid lowercase DNS hostname

### spec.addresses[].type

`string` · required · optional (explicit presence)

How the address value is interpreted. Core values are "IPAddress"
(default) and "Hostname"; "NamedAddress" is deprecated. Custom types use a
domain-prefixed string.

Upstream default: "IPAddress". Open set validated by pattern.

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^Hostname|IPAddress|NamedAddress|[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\\/[A-Za-z0-9\\/\\-._~%!$&'()*+,;=:]+$"}}

### spec.addresses[].value

`string`

The address value (for example "1.2.3.4", "128::1", or "my-ip-address").
When empty, the implementation assigns a matching address if it can.

- rule: {"string":{"maxLen":"253"}}

### spec.infrastructure

`KubernetesGatewayInfrastructure`

Infrastructure-level attributes (labels, annotations, and a per-Gateway
parameters reference) applied to resources the controller creates for this
Gateway. Upstream support level: Extended.

### spec.infrastructure.labels

`map<string, string>`

Labels applied to created resources. Up to 8 entries.

Upstream key/value: Gateway API LabelKey (1-253) / LabelValue (0-63).

- rule: {"map":{"maxPairs":"8","keys":{"string":{"minLen":"1","maxLen":"253","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?([A-Za-z0-9][-A-Za-z0-9_.]{0,61})?[A-Za-z0-9]$"}},"values":{"string":{"maxLen":"63"}}}}

### spec.infrastructure.annotations

`map<string, string>`

Annotations applied to created resources. Up to 16 entries. This is the
Gateway's per-cloud injection surface: the controller propagates these
annotations to the LoadBalancer Service (or equivalent) it creates, so
cloud LB behavior (internal vs internet-facing, NLB attributes, static
IPs) is selected exactly as it is on a Service of type LoadBalancer.

Upstream key/value: Gateway API AnnotationKey (1-253) / AnnotationValue
(0-4096).

- rule: {"map":{"maxPairs":"16","keys":{"string":{"minLen":"1","maxLen":"253","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?([A-Za-z0-9][-A-Za-z0-9_.]{0,61})?[A-Za-z0-9]$"}},"values":{"string":{"maxLen":"4096"}}}}

### spec.infrastructure.parametersRef

`KubernetesGatewayLocalParametersReference`

Reference to a controller-specific parameters resource for this Gateway,
mirroring GatewayClass.parametersRef but per-Gateway. Optional.

### spec.infrastructure.parametersRef.group

`string` · required · optional (explicit presence)

API group of the referent. Empty string infers the core API group
(e.g. for ConfigMap parameters).

Upstream requires the KEY to be present but allows the empty value --
so this is a proto3 `optional` string with a presence `required` rule:
it must be SET (and is therefore emitted to the rendered CR, whose CRD
rejects a missing key) but may be empty. The `optional` is what keeps
the projection faithful: protojson omits unset proto3 scalars, so a
non-optional empty-string group would be dropped from the manifest and
rejected by the API server.

Group pattern: empty or an RFC 1123 subdomain (max 253).

- rule: {"required":true,"string":{"maxLen":"253","pattern":"^$|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"}}

### spec.infrastructure.parametersRef.kind

`string` · required

Kind of the referent.

- rule: {"required":true}

### spec.infrastructure.parametersRef.name

`string` · required

Name of the referent (1-253 chars).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"253"}}

### spec.allowedListeners

`KubernetesGatewayAllowedListeners`

Controls which ListenerSets may attach to this Gateway. Defaults to
allowing none — a Gateway must opt in before any KubernetesListenerSet
can merge additional listeners into it.

### spec.allowedListeners.namespaces

`KubernetesGatewayListenerNamespaces`

Namespaces from which ListenerSets may attach. Defaults to allowing none.

### spec.allowedListeners.namespaces.from

`string` · optional (explicit presence)

Where ListenerSets may attach from: "All", "Selector", "Same", or "None".

Upstream default: "None". Closed enum: All | Selector | Same | None.

- rule: from must be one of 'All', 'Selector', 'Same', or 'None'

### spec.allowedListeners.namespaces.selector

`KubernetesGatewayApiLabelSelector`

Namespace label selector. Required (and only honored) when `from` is
"Selector".

### spec.allowedListeners.namespaces.selector.matchLabels

`map<string, string>`

Map of {key,value} pairs. A namespace matches if it carries all of these
labels.

### spec.allowedListeners.namespaces.selector.matchExpressions

`[]KubernetesGatewayApiLabelSelectorRequirement`

List of label selector requirements, ANDed together with match_labels.

### spec.allowedListeners.namespaces.selector.matchExpressions[].key

`string` · required

Label key the requirement applies to.

- rule: {"required":true}

### spec.allowedListeners.namespaces.selector.matchExpressions[].operator

`string` · required

Relationship between the key and values: "In", "NotIn", "Exists", or
"DoesNotExist". For "Exists"/"DoesNotExist", values must be empty.

- rule: operator must be one of 'In', 'NotIn', 'Exists', or 'DoesNotExist'
- rule: {"required":true}

### spec.allowedListeners.namespaces.selector.matchExpressions[].values

`[]string`

Values for the requirement. Must be non-empty for "In"/"NotIn" and empty
for "Exists"/"DoesNotExist".

### spec.tls

`KubernetesGatewayTlsConfig`

Gateway-wide TLS configuration: client-certificate validation for inbound
(frontend) HTTPS traffic and client-certificate material for outbound
(backend) connections. Per-listener TLS termination is configured on each
listener instead. Upstream support level: Extended.

### spec.tls.backend

`KubernetesGatewayBackendTls`

TLS configuration the Gateway uses as a client when connecting to backends.

### spec.tls.backend.clientCertificateRef

`KubernetesGatewayApiSecretObjectReference`

Reference to a Secret holding a client certificate and key the Gateway
presents to backends. A kubernetes.io/tls Secret has Core support.

### spec.tls.backend.clientCertificateRef.group

`string` · optional (explicit presence)

Group of the referent. Empty string infers the core API group.

Upstream default: "" (core API group)
Group pattern: empty or an RFC 1123 subdomain (max 253).

- rule: {"string":{"maxLen":"253","pattern":"^$|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"}}

### spec.tls.backend.clientCertificateRef.kind

`string` · required · optional (explicit presence)

Kind of the referent.

Upstream default: "Secret"
Kind pattern: 1-63 chars, ^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$

- rule: {"string":{"minLen":"1","maxLen":"63","pattern":"^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$"}}

### spec.tls.backend.clientCertificateRef.name

`string | valueFrom` · required

Name of the referent. Defaults to a KubernetesSecret foreign key. The
Secret is typically produced by cert-manager: reference a
KubernetesCertificate's status.outputs.secret_name with valueFrom to wire
a certificate the moment it is issued, or a KubernetesSecret directly for
externally provisioned material. Pass the literal name with `value:` when
the Secret is not Planton-managed.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.tls.backend.clientCertificateRef.namespace

`string` · optional (explicit presence)

Namespace of the referenced object. When unspecified, the local namespace
is inferred. Cross-namespace references require a ReferenceGrant.

### spec.tls.frontend

`KubernetesGatewayFrontendTlsConfig`

Client-certificate validation for inbound (frontend) HTTPS connections.

- rule: each per-port frontend TLS configuration must target a unique port

### spec.tls.frontend.default

`KubernetesGatewayFrontendTlsValidationConfig` · required

Default client-certificate validation applied to every HTTPS listener
unless overridden per port. Required.

- rule: {"required":true}

### spec.tls.frontend.default.validation

`KubernetesGatewayFrontendTlsValidation`

Client (frontend) certificate validation settings. Setting this enables
mutual TLS for connections to the Gateway.

### spec.tls.frontend.default.validation.caCertificateRefs

`[]KubernetesGatewayApiObjectReference` · required

References to ConfigMaps holding PEM-encoded CA bundles used as trust
anchors for client certificates. At least one, up to sixteen. Each
reference's name defaults to a KubernetesConfigMap foreign key.

- rule: {"repeated":{"minItems":"1","maxItems":"16"}}

### spec.tls.frontend.default.validation.caCertificateRefs[].group

`string` · required · optional (explicit presence)

Group of the referent. Empty string infers the core API group
(e.g. for ConfigMap CA bundles).

Upstream requires the KEY to be present, but its Group type explicitly
allows the empty value -- so this is a proto3 `optional` string with a
presence `required` rule: it must be SET (and is therefore emitted to
the rendered CR, whose CRD rejects a missing key) but may be empty.
The `optional` is what keeps the projection faithful: protojson omits
unset proto3 scalars, so a non-optional empty-string group would be
dropped from the manifest and rejected by the API server.

Group pattern: empty or an RFC 1123 subdomain (max 253).

- rule: {"required":true,"string":{"maxLen":"253","pattern":"^$|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"}}

### spec.tls.frontend.default.validation.caCertificateRefs[].kind

`string` · required

Kind of the referent (e.g., "ConfigMap" or "Service").

Upstream models Kind as a required value type.
Kind pattern: 1-63 chars, ^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$

- rule: {"required":true,"string":{"minLen":"1","maxLen":"63","pattern":"^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$"}}

### spec.tls.frontend.default.validation.caCertificateRefs[].name

`string | valueFrom` · required

Name of the referent. Defaults to a KubernetesConfigMap foreign key (this
reference's primary use is CA-bundle ConfigMaps for client-certificate
validation): wire it with valueFrom against the ConfigMap resource, or
pass the literal name with `value:` when the object is a different kind
or not Planton-managed.

- references: KubernetesConfigMap (`status.outputs.configmap_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesConfigMap, name: <that resource's name>, fieldPath: status.outputs.configmap_name}} -- a bare string does not parse

### spec.tls.frontend.default.validation.caCertificateRefs[].namespace

`string` · optional (explicit presence)

Namespace of the referenced object. When unspecified, the local namespace
is inferred. Cross-namespace references require a ReferenceGrant.

### spec.tls.frontend.default.validation.mode

`string` · optional (explicit presence)

How strictly client certificates are validated. "AllowValidOnly" (default)
accepts only connections presenting a valid client certificate.
"AllowInsecureFallback" accepts connections even without a valid
certificate and carries significant security risk.

Upstream default: "AllowValidOnly". Closed enum.

- rule: mode must be either 'AllowValidOnly' or 'AllowInsecureFallback'

### spec.tls.frontend.perPort

`[]KubernetesGatewayTlsPortConfig`

Per-port overrides of the default validation. Each entry targets a unique
port.

- rule: {"repeated":{"maxItems":"64"}}

### spec.tls.frontend.perPort[].port

`int32` · required

Port the configuration applies to.

- rule: {"required":true,"int32":{"lte":65535,"gte":1}}

### spec.tls.frontend.perPort[].tls

`KubernetesGatewayFrontendTlsValidationConfig` · required

TLS validation configuration applied to HTTPS listeners on this port.

- rule: {"required":true}

### spec.tls.frontend.perPort[].tls.validation

`KubernetesGatewayFrontendTlsValidation`

Client (frontend) certificate validation settings. Setting this enables
mutual TLS for connections to the Gateway.

### spec.tls.frontend.perPort[].tls.validation.caCertificateRefs

`[]KubernetesGatewayApiObjectReference` · required

References to ConfigMaps holding PEM-encoded CA bundles used as trust
anchors for client certificates. At least one, up to sixteen. Each
reference's name defaults to a KubernetesConfigMap foreign key.

- rule: {"repeated":{"minItems":"1","maxItems":"16"}}

### spec.tls.frontend.perPort[].tls.validation.caCertificateRefs[].group

`string` · required · optional (explicit presence)

Group of the referent. Empty string infers the core API group
(e.g. for ConfigMap CA bundles).

Upstream requires the KEY to be present, but its Group type explicitly
allows the empty value -- so this is a proto3 `optional` string with a
presence `required` rule: it must be SET (and is therefore emitted to
the rendered CR, whose CRD rejects a missing key) but may be empty.
The `optional` is what keeps the projection faithful: protojson omits
unset proto3 scalars, so a non-optional empty-string group would be
dropped from the manifest and rejected by the API server.

Group pattern: empty or an RFC 1123 subdomain (max 253).

- rule: {"required":true,"string":{"maxLen":"253","pattern":"^$|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"}}

### spec.tls.frontend.perPort[].tls.validation.caCertificateRefs[].kind

`string` · required

Kind of the referent (e.g., "ConfigMap" or "Service").

Upstream models Kind as a required value type.
Kind pattern: 1-63 chars, ^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$

- rule: {"required":true,"string":{"minLen":"1","maxLen":"63","pattern":"^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$"}}

### spec.tls.frontend.perPort[].tls.validation.caCertificateRefs[].name

`string | valueFrom` · required

Name of the referent. Defaults to a KubernetesConfigMap foreign key (this
reference's primary use is CA-bundle ConfigMaps for client-certificate
validation): wire it with valueFrom against the ConfigMap resource, or
pass the literal name with `value:` when the object is a different kind
or not Planton-managed.

- references: KubernetesConfigMap (`status.outputs.configmap_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesConfigMap, name: <that resource's name>, fieldPath: status.outputs.configmap_name}} -- a bare string does not parse

### spec.tls.frontend.perPort[].tls.validation.caCertificateRefs[].namespace

`string` · optional (explicit presence)

Namespace of the referenced object. When unspecified, the local namespace
is inferred. Cross-namespace references require a ReferenceGrant.

### spec.tls.frontend.perPort[].tls.validation.mode

`string` · optional (explicit presence)

How strictly client certificates are validated. "AllowValidOnly" (default)
accepts only connections presenting a valid client certificate.
"AllowInsecureFallback" accepts connections even without a valid
certificate and carries significant security risk.

Upstream default: "AllowValidOnly". Closed enum.

- rule: mode must be either 'AllowValidOnly' or 'AllowInsecureFallback'

## Validation Rules

- `gateway.listener_names_unique`: each listener name must be unique within the Gateway
- `gateway.listener_port_protocol_hostname_unique`: each listener must have a unique combination of port, protocol, and hostname
- `gateway.ip_addresses_unique`: each requested IPAddress value must be unique
- `gateway.hostname_addresses_unique`: each requested Hostname address value must be unique

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesGateway, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.gateway_name` | `string` | Name of the created Gateway (equals metadata.name). Routes attach to this Gateway by referencing this name in their parentRefs; in InfraCharts it orders the Gateway before the Routes that target it. |
| `status.outputs.namespace` | `string` | Namespace the Gateway was created in (the resolved spec.namespace). Routes and TLS Secrets that reference this Gateway are subject to same-namespace and ReferenceGrant rules relative to this value. |
| `status.outputs.gateway_class_name` | `string` | Name of the GatewayClass this Gateway belongs to (the resolved spec.gateway_class_name), exposed for observability and to confirm which controller implementation programs this Gateway. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.gatewayClassName` | KubernetesGatewayClass | `status.outputs.gateway_class_name` |
| `spec.listeners[].tls.certificateRefs[].name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.tls.backend.clientCertificateRef.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.tls.frontend.default.validation.caCertificateRefs[].name` | KubernetesConfigMap | `status.outputs.configmap_name` |
| `spec.tls.frontend.perPort[].tls.validation.caCertificateRefs[].name` | KubernetesConfigMap | `status.outputs.configmap_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesAuthorizationPolicy | `spec.targetRefs[].name` | `status.outputs.gateway_name` |
| KubernetesEnvoyFilter | `spec.targetRefs[].name` | `status.outputs.gateway_name` |
| KubernetesGrpcRoute | `spec.parentRefs[].name` | `status.outputs.gateway_name` |
| KubernetesHttpRoute | `spec.parentRefs[].name` | `status.outputs.gateway_name` |
| KubernetesListenerSet | `spec.parentRef.name` | `status.outputs.gateway_name` |
| KubernetesPlantonPlatform | `spec.ingress.gatewayRef.name` | `status.outputs.gateway_name` |
| KubernetesPlantonPlatform | `spec.ingress.gatewayRef.namespace` | `status.outputs.namespace` |
| KubernetesRequestAuthentication | `spec.targetRefs[].name` | `status.outputs.gateway_name` |
| KubernetesTcpRoute | `spec.parentRefs[].name` | `status.outputs.gateway_name` |
| KubernetesTelemetry | `spec.targetRefs[].name` | `status.outputs.gateway_name` |
| KubernetesTlsRoute | `spec.parentRefs[].name` | `status.outputs.gateway_name` |
| KubernetesUdpRoute | `spec.parentRefs[].name` | `status.outputs.gateway_name` |

## See Also

- [Overview](../README.md)
