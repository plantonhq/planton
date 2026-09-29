# KubernetesRequestAuthentication

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

KubernetesRequestAuthenticationSpec defines an Istio RequestAuthentication: a
namespaced policy that defines which JSON Web Tokens (JWTs) are accepted on the
workloads it selects. A valid token's identity is extracted and made available
to authorization policies; an invalid token causes the request to be rejected.
A request with no token is allowed by RequestAuthentication alone (pair it with
an AuthorizationPolicy to require a principal).

100% fidelity with the upstream istio.io/api RequestAuthentication
(security/v1beta1/request_authentication.proto, served as security.istio.io/v1),
pinned to the 1.30 line (tag 1.30.3). Upstream spec fields are flattened directly
after the Planton namespaced envelope (namespace); there is no
nested `request_authentication` sub-message.

Scope semantics (upstream): if neither `selector` nor `target_refs` is set the
policy matches all workloads in its namespace (or, in the mesh root namespace,
the whole mesh). At most one of `selector` and `target_refs` may be set
(enforced below).

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRequestAuthentication
metadata:
  name: test-request-authentication
spec:
  namespace:
    value: test-namespace
  selector:
    match_labels:
      app: finance
  jwt_rules:
    - issuer: https://accounts.example.com
      jwks_uri: https://accounts.example.com/.well-known/jwks.json
      audiences:
        - finance-api.example.com
      from_headers:
        - name: x-jwt-assertion
          prefix: "Bearer "
      output_claim_to_headers:
        - header: x-jwt-group
          claim: groups
      forward_original_token: true
      timeout: 5s
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.selector` | `KubernetesIstioApiWorkloadSelector` |  |  |  |
| `spec.selector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.targetRefs` | `[]KubernetesIstioApiPolicyTargetReference` |  |  |  |
| `spec.targetRefs[].group` | `string` |  |  |  |
| `spec.targetRefs[].kind` | `string` | yes |  |  |
| `spec.targetRefs[].name` | `string \| valueFrom` | yes |  | KubernetesGateway (`status.outputs.gateway_name`) |
| `spec.targetRefs[].namespace` | `string` |  |  |  |
| `spec.jwtRules` | `[]KubernetesRequestAuthenticationJwtRule` |  |  |  |
| `spec.jwtRules[].issuer` | `string` | yes |  |  |
| `spec.jwtRules[].audiences` | `[]string` |  |  |  |
| `spec.jwtRules[].jwksUri` | `string` |  |  |  |
| `spec.jwtRules[].jwks` | `string` |  |  |  |
| `spec.jwtRules[].fromHeaders` | `[]KubernetesRequestAuthenticationJwtHeader` |  |  |  |
| `spec.jwtRules[].fromHeaders[].name` | `string` | yes |  |  |
| `spec.jwtRules[].fromHeaders[].prefix` | `string` |  |  |  |
| `spec.jwtRules[].fromParams` | `[]string` |  |  |  |
| `spec.jwtRules[].fromCookies` | `[]string` |  |  |  |
| `spec.jwtRules[].outputPayloadToHeader` | `string` |  |  |  |
| `spec.jwtRules[].forwardOriginalToken` | `bool` |  |  |  |
| `spec.jwtRules[].outputClaimToHeaders` | `[]KubernetesRequestAuthenticationClaimToHeader` |  |  |  |
| `spec.jwtRules[].outputClaimToHeaders[].header` | `string` | yes |  |  |
| `spec.jwtRules[].outputClaimToHeaders[].claim` | `string` | yes |  |  |
| `spec.jwtRules[].timeout` | `string` |  |  |  |
| `spec.jwtRules[].spaceDelimitedClaims` | `[]string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace in which the RequestAuthentication is created. The policy's scope is
this namespace (or mesh-wide if this is the Istio root namespace and no selector
or target_refs is set).

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.selector

`KubernetesIstioApiWorkloadSelector`

Selects the workloads (by pod/VM label) this policy applies to. When omitted
(and target_refs is also omitted), the policy applies to every workload in its
namespace (or, in the root namespace, the whole mesh).

INFRA-CHART COMPOSABILITY: selector is a PLAIN label match, not an
Planton foreign key (StringValueOrRef). It is matched at runtime by istiod
against pod labels and creates NO automatic DAG edge to any workload resource.
To order this policy after the workload it protects in an infra chart, an author
MUST express the dependency via metadata.relationships, e.g.:
  metadata:
    relationships:
      - kind: KubernetesDeployment
        name: "{{ values.app }}"
        type: depends_on
See the component's "Composing in Infra Charts" docs for the full pattern.

### spec.selector.matchLabels

`map<string, string>`

One or more labels indicating the set of pods/VMs the policy applies to.
Faithful to istio.io/api `istio.type.v1beta1.WorkloadSelector.match_labels`,
whose upstream CRD constraints are: max 4096 entries; each value <= 63 chars;
label keys must be non-empty; and wildcards ('*') are not permitted in keys or
values. The size/length bounds are expressed via the standard `map` rule; the
non-empty-key and no-wildcard constraints map to upstream's CEL XValidation
rules and are expressed here as field-level CEL.

- rule: label selector keys must not be empty
- rule: wildcard ('*') is not allowed in label selector keys
- rule: wildcard ('*') is not allowed in label selector values
- rule: {"map":{"maxPairs":"4096","values":{"string":{"maxLen":"63"}}}}

### spec.targetRefs

`[]KubernetesIstioApiPolicyTargetReference`

Attaches the policy to specific resources (Gateway, Service, ServiceEntry)
instead of selecting workloads by label. At most one of `selector` and
`target_refs` may be set (enforced below). Waypoint proxies require this field
(label `selector` policies are ignored by waypoints). Upstream allows up to 16.

INFRA-CHART COMPOSABILITY: each target_ref's `name` is a foreign key that
defaults to a KubernetesGateway (its status.outputs.gateway_name): wiring it with
valueFrom orders this policy after the gateway it protects. A literal `value:`
covers a Service, a ServiceEntry, or anything created outside Planton; istiod
resolves group/kind/name at runtime, so order this policy after such a target
with metadata.relationships (`uses` -> KubernetesService / KubernetesServiceEntry).
See the component's "Composing in Infra Charts" docs.

- rule: {"repeated":{"maxItems":"16"}}

### spec.targetRefs[].group

`string`

Group of the target resource. Empty for the core API group (Services). Faithful
to the upstream pattern (empty, or a DNS-1123 subdomain).

- rule: {"string":{"maxLen":"253","pattern":"^$|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"}}

### spec.targetRefs[].kind

`string` · required

Kind of the target resource (e.g. Gateway, Service, ServiceEntry). Required.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"63","pattern":"^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$"}}

### spec.targetRefs[].name

`string | valueFrom` · required

Name of the target resource. Required. Defaults to a KubernetesGateway foreign
key (the policy attaches to a Gateway API Gateway) — in an infra chart, wire it
with valueFrom so the policy deploys after its gateway. For other target kinds,
pass the literal name with `value:`. Upstream bounds the name at 253 characters;
the API server enforces that at apply (a StringValueOrRef carries no bound).

- references: KubernetesGateway (`status.outputs.gateway_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesGateway, name: <that resource's name>, fieldPath: status.outputs.gateway_name}} -- a bare string does not parse

### spec.targetRefs[].namespace

`string`

Namespace of the target resource. Cross-namespace attachment is not supported
upstream in the 1.30 line, so this must be empty (the target is resolved in the
policy's own namespace). Mirrors the upstream XValidation rule
"cross namespace referencing is not currently supported".

- rule: cross-namespace target references are not supported; leave namespace empty

### spec.jwtRules

`[]KubernetesRequestAuthenticationJwtRule`

The set of JWT rules evaluated at the selected workloads' proxies. A token is
validated only when presented at a location a rule recognizes; if validation
fails the request is rejected. When empty, the policy installs no JWT
requirements (it is a no-op until rules are added). Upstream allows up to 4096.

- rule: {"repeated":{"maxItems":"4096"}}
- rule: at most one of jwks_uri or jwks may be set
- rule: at least one of issuer or jwks_uri must be set — without either, the verifier has no way to locate the signing keys

### spec.jwtRules[].issuer

`string` · required · optional (explicit presence)

The issuer that issued the JWT (the `iss` claim). A JWT with a different issuer
is rejected. Example: `https://foobar.auth0.com`. May be omitted only when
`jwks_uri` is set (at least one of the two is required — enforced below,
mirroring the istiod webhook's "issuer or jwksUri must be non-empty" rejection);
issuer-less rules validate any issuer against the fixed key set.

- rule: {"string":{"minLen":"1"}}

### spec.jwtRules[].audiences

`[]string`

The JWT audiences (`aud` claim) allowed to access. A JWT carrying any of these
audiences is accepted; when empty, the workload's service name is accepted.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.jwtRules[].jwksUri

`string` · optional (explicit presence)

URL of the provider's JSON Web Key Set used to validate the JWT signature.
Optional when the key set can be discovered from the issuer (OpenID Discovery)
or inferred from the issuer's email domain. Only one of `jwks_uri` and `jwks`
may be set (enforced below). Example:
`https://www.googleapis.com/oauth2/v1/certs`.

- rule: jwks_uri must use the http:// or https:// scheme
- rule: {"string":{"maxLen":"2048"}}

### spec.jwtRules[].jwks

`string` · optional (explicit presence)

The JSON Web Key Set (inline) used to validate the JWT signature. Only one of
`jwks_uri` and `jwks` may be set (enforced below).

### spec.jwtRules[].fromHeaders

`[]KubernetesRequestAuthenticationJwtHeader`

HTTP header locations to extract the JWT from. For example, a header
`x-jwt-assertion` with a `Bearer ` prefix. If no location is specified, the
Authorization Bearer header and the `access_token` query parameter are tried.

### spec.jwtRules[].fromHeaders[].name

`string` · required

The HTTP header name. Required.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.jwtRules[].fromHeaders[].prefix

`string` · optional (explicit presence)

The prefix stripped before decoding the token. For `Authorization: Bearer
<token>`, set prefix to `Bearer ` (with the trailing space). If the header does
not carry this exact prefix, it is treated as invalid.

### spec.jwtRules[].fromParams

`[]string`

Query parameter names to extract the JWT from (e.g. `?my_token=<JWT>`).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.jwtRules[].fromCookies

`[]string`

Cookie names to extract the JWT from.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.jwtRules[].outputPayloadToHeader

`string` · optional (explicit presence)

Header name to output the verified JWT payload to, as
`base64_encoded(jwt_payload_in_JSON)`. When unset, the payload is not emitted.

### spec.jwtRules[].forwardOriginalToken

`bool` · optional (explicit presence)

When true, the original token is forwarded to the upstream request. Upstream
default is false.

### spec.jwtRules[].outputClaimToHeaders

`[]KubernetesRequestAuthenticationClaimToHeader`

Operations copying individual verified claims to HTTP headers. Each header name
must be unique. An alternative to output_payload_to_header for emitting selected
claims rather than the whole payload.

### spec.jwtRules[].outputClaimToHeaders[].header

`string` · required

The header to create (overwriting any existing value). Required. Only
dash/underscore/alphanumeric characters are allowed.

- rule: {"required":true,"string":{"minLen":"1","pattern":"^[-_A-Za-z0-9]+$"}}

### spec.jwtRules[].outputClaimToHeaders[].claim

`string` · required

The claim to copy from. Required. Only string/int/bool claims are supported;
nested claims use dotted paths (e.g. `nested.key.group`).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.jwtRules[].timeout

`string` · optional (explicit presence)

Maximum time the resolver spends fetching the JWKS. A google.protobuf.Duration
string (e.g. "5s", "1500ms", "1m30s"); durations are modeled as strings.
Upstream default is 5s and minimum is 1ms.

- rule: timeout must be a valid duration of at least 1ms (e.g. "5s")

### spec.jwtRules[].spaceDelimitedClaims

`[]string`

JWT claim names whose string values should be treated as SPACE-DELIMITED lists
(the OAuth2 `scope` claim convention, e.g. "read write admin") when matched by
authorization policies and claim-to-header operations. Upstream allows up to 64.

- rule: {"repeated":{"maxItems":"64","items":{"string":{"minLen":"1"}}}}

## Validation Rules

- `request_authentication.selector_xor_target_refs`: at most one of selector or target_refs may be set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesRequestAuthentication, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.request_authentication_name` | `string` | Name of the created RequestAuthentication (equals metadata.name). |
| `status.outputs.namespace` | `string` | Namespace the RequestAuthentication was created in (the resolved spec.namespace). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.targetRefs[].name` | KubernetesGateway | `status.outputs.gateway_name` |

## See Also

- [Overview](../README.md)
