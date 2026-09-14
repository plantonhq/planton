# KubernetesAuthorizationPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

KubernetesAuthorizationPolicySpec defines an Istio AuthorizationPolicy: a
namespaced policy that enforces access control on the workloads it selects.
The policy `action` (ALLOW, DENY, AUDIT, or CUSTOM) is applied to requests that
match its `rules`; CUSTOM delegates the decision to a named external authorizer.

100% fidelity with the upstream istio.io/api AuthorizationPolicy
(security/v1beta1/authorization_policy.proto, served as security.istio.io/v1),
pinned to the 1.30 line (tag 1.30.3). Upstream spec fields are flattened directly
after the Planton namespaced envelope (namespace); there is no
nested `authorization_policy` sub-message.

Scope semantics (upstream): if neither `selector` nor `target_refs` is set the
policy matches all workloads in its namespace (or, in the mesh root namespace,
the whole mesh). At most one of `selector` and `target_refs` may be set
(enforced below). An empty `rules` list means the match never occurs, which is a
deny-by-default when the action is ALLOW.

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAuthorizationPolicy
metadata:
  name: test-authorization-policy
spec:
  namespace:
    value: test-namespace
  selector:
    match_labels:
      app: httpbin
  action: ALLOW
  rules:
    - from:
        - source:
            request_principals:
              - "*"
        - source:
            namespaces:
              - test-namespace
      to:
        - operation:
            methods:
              - GET
              - POST
            paths:
              - /info*
              - /data
      when:
        - key: request.auth.claims[iss]
          values:
            - https://accounts.example.com
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
| `spec.rules` | `[]KubernetesAuthorizationPolicyRule` |  |  |  |
| `spec.rules[].from` | `[]KubernetesAuthorizationPolicyRuleFrom` |  |  |  |
| `spec.rules[].from[].source` | `KubernetesAuthorizationPolicySource` |  |  |  |
| `spec.rules[].from[].source.principals` | `[]string` |  |  |  |
| `spec.rules[].from[].source.notPrincipals` | `[]string` |  |  |  |
| `spec.rules[].from[].source.requestPrincipals` | `[]string` |  |  |  |
| `spec.rules[].from[].source.notRequestPrincipals` | `[]string` |  |  |  |
| `spec.rules[].from[].source.namespaces` | `[]string` |  |  |  |
| `spec.rules[].from[].source.notNamespaces` | `[]string` |  |  |  |
| `spec.rules[].from[].source.serviceAccounts` | `[]string` |  |  |  |
| `spec.rules[].from[].source.notServiceAccounts` | `[]string` |  |  |  |
| `spec.rules[].from[].source.ipBlocks` | `[]string` |  |  |  |
| `spec.rules[].from[].source.notIpBlocks` | `[]string` |  |  |  |
| `spec.rules[].from[].source.remoteIpBlocks` | `[]string` |  |  |  |
| `spec.rules[].from[].source.notRemoteIpBlocks` | `[]string` |  |  |  |
| `spec.rules[].from[].source.trustDomains` | `[]string` |  |  |  |
| `spec.rules[].from[].source.notTrustDomains` | `[]string` |  |  |  |
| `spec.rules[].to` | `[]KubernetesAuthorizationPolicyRuleTo` |  |  |  |
| `spec.rules[].to[].operation` | `KubernetesAuthorizationPolicyOperation` |  |  |  |
| `spec.rules[].to[].operation.hosts` | `[]string` |  |  |  |
| `spec.rules[].to[].operation.notHosts` | `[]string` |  |  |  |
| `spec.rules[].to[].operation.ports` | `[]string` |  |  |  |
| `spec.rules[].to[].operation.notPorts` | `[]string` |  |  |  |
| `spec.rules[].to[].operation.methods` | `[]string` |  |  |  |
| `spec.rules[].to[].operation.notMethods` | `[]string` |  |  |  |
| `spec.rules[].to[].operation.paths` | `[]string` |  |  |  |
| `spec.rules[].to[].operation.notPaths` | `[]string` |  |  |  |
| `spec.rules[].when` | `[]KubernetesAuthorizationPolicyCondition` |  |  |  |
| `spec.rules[].when[].key` | `string` | yes |  |  |
| `spec.rules[].when[].values` | `[]string` |  |  |  |
| `spec.rules[].when[].notValues` | `[]string` |  |  |  |
| `spec.rules[].matchAll` | `bool` |  |  |  |
| `spec.action` | `string` |  |  |  |
| `spec.provider` | `KubernetesAuthorizationPolicyExtensionProvider` |  |  |  |
| `spec.provider.name` | `string` | yes |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace in which the AuthorizationPolicy is created. The policy's scope is
this namespace (or mesh-wide if this is the Istio root namespace and no
selector or target_refs is set).

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.selector

`KubernetesIstioApiWorkloadSelector`

Selects the workloads (by pod/VM label) this policy applies to. When omitted
(and target_refs is also omitted), the policy applies to every workload in its
namespace (or, in the root namespace, the whole mesh). At most one of
`selector` and `target_refs` may be set (enforced above).

INFRA-CHART COMPOSABILITY: selector is a PLAIN label match, not an
Planton foreign key (StringValueOrRef). It is matched at runtime by istiod
against pod labels and creates NO automatic DAG edge to any workload resource.
To order this policy after the workload it protects in an infra chart, an
author MUST express the dependency via metadata.relationships, e.g.:
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
`target_refs` may be set (enforced above). Waypoint proxies require this field
(label `selector` policies are ignored by waypoints). Upstream allows up to 16.

INFRA-CHART COMPOSABILITY: a target reference is a PLAIN cross-resource
reference, not an Planton foreign key. istiod resolves it at runtime, creating
NO automatic DAG edge. Order this policy after the referenced resource via
metadata.relationships (`uses` -> KubernetesGateway / KubernetesService /
KubernetesServiceEntry). See the component's "Composing in Infra Charts" docs.

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

### spec.rules

`[]KubernetesAuthorizationPolicyRule`

The list of rules evaluated against each request. A request matches the policy
when at least one rule matches it. When empty, the match never occurs: with the
ALLOW action this denies all requests to the selected workloads; with DENY it
is a no-op. Upstream allows up to 512.

- rule: {"repeated":{"maxItems":"512"}}
- rule: match_all: true matches every request — drop from, to, and when, or drop match_all to match by them
- rule: a rule must say what it matches — set match_all: true to match every request, or name from, to, or when

### spec.rules[].from

`[]KubernetesAuthorizationPolicyRuleFrom`

Sources of the request. If empty, any source matches. Upstream allows up to 512.

- rule: {"repeated":{"maxItems":"512"}}

### spec.rules[].from[].source

`KubernetesAuthorizationPolicySource`

The source of a request.

- rule: service_accounts/not_service_accounts cannot be set together with principals, not_principals, namespaces, or not_namespaces

### spec.rules[].from[].source.principals

`[]string`

Peer identities (from the peer certificate), e.g.
`cluster.local/ns/default/sa/productpage`. Requires mTLS. If empty, any
principal matches.

### spec.rules[].from[].source.notPrincipals

`[]string`

Negative match of peer identities.

### spec.rules[].from[].source.requestPrincipals

`[]string`

Request identities (from the validated JWT), in the form `<ISS>/<SUB>`, e.g.
`example.com/sub-1`. Requires request authentication. If empty, any request
principal matches.

### spec.rules[].from[].source.notRequestPrincipals

`[]string`

Negative match of request identities.

### spec.rules[].from[].source.namespaces

`[]string`

Namespaces (from the peer certificate). Requires mTLS. If empty, any namespace
matches.

### spec.rules[].from[].source.notNamespaces

`[]string`

Negative match of namespaces.

### spec.rules[].from[].source.serviceAccounts

`[]string`

Kubernetes service accounts (from the peer certificate), in the form
`<namespace>/<serviceaccount>` (or `<serviceaccount>` for the policy's own
namespace). Requires mTLS. No wildcards. Cannot be combined with `principals`
or `namespaces` (enforced above). Upstream allows up to 16, each up to 320
characters.

- rule: {"repeated":{"maxItems":"16","items":{"string":{"maxLen":"320"}}}}

### spec.rules[].from[].source.notServiceAccounts

`[]string`

Negative match of Kubernetes service accounts. Same format and bounds as
`service_accounts`.

- rule: {"repeated":{"maxItems":"16","items":{"string":{"maxLen":"320"}}}}

### spec.rules[].from[].source.ipBlocks

`[]string`

IP blocks from the source address of the IP packet. Single IP (`203.0.113.4`)
or CIDR (`203.0.113.0/24`). If empty, any IP matches.

### spec.rules[].from[].source.notIpBlocks

`[]string`

Negative match of IP blocks.

### spec.rules[].from[].source.remoteIpBlocks

`[]string`

IP blocks from the `X-Forwarded-For` header or proxy protocol (requires
`numTrustedProxies` configured on the gateway topology). Single IP or CIDR. If
empty, any IP matches.

### spec.rules[].from[].source.notRemoteIpBlocks

`[]string`

Negative match of remote IP blocks.

### spec.rules[].from[].source.trustDomains

`[]string`

Trust domains of the peer identity (the mesh's identity root, e.g.
`cluster.local`) — matches any workload in the listed trust domains without
enumerating principals. Requires mTLS. If empty, any trust domain matches.

### spec.rules[].from[].source.notTrustDomains

`[]string`

Negative match of trust domains.

### spec.rules[].to

`[]KubernetesAuthorizationPolicyRuleTo`

Operations of the request. If empty, any operation matches.

### spec.rules[].to[].operation

`KubernetesAuthorizationPolicyOperation`

The operation of a request.

### spec.rules[].to[].operation.hosts

`[]string`

Hosts (the HTTP Host/Authority header), case-insensitive. HTTP only. If empty,
any host matches.

### spec.rules[].to[].operation.notHosts

`[]string`

Negative match of hosts.

### spec.rules[].to[].operation.ports

`[]string`

Ports of the connection (as strings, e.g. `"8080"`). If empty, any port
matches.

### spec.rules[].to[].operation.notPorts

`[]string`

Negative match of ports.

### spec.rules[].to[].operation.methods

`[]string`

HTTP methods (for gRPC this is always `POST`). HTTP only. If empty, any method
matches.

### spec.rules[].to[].operation.notMethods

`[]string`

Negative match of methods.

### spec.rules[].to[].operation.paths

`[]string`

HTTP request paths (for gRPC, `/package.service/method`). Supports the Envoy URI
template operators `{*}` and `{**}`. HTTP only. If empty, any path matches.

### spec.rules[].to[].operation.notPaths

`[]string`

Negative match of paths.

### spec.rules[].when

`[]KubernetesAuthorizationPolicyCondition`

Additional conditions of the request. If empty, any condition matches.

### spec.rules[].when[].key

`string` · required

The Istio attribute name, e.g. `request.auth.claims[iss]` or `source.ip`.
Required. See the Istio "supported conditions" reference for the full list.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].when[].values

`[]string`

Allowed values for the attribute. Upstream requires at least one of `values` or
`not_values` to be set; that coupling is not part of the CRD's declared
validation, so it is not enforced here (match the validated surface).

### spec.rules[].when[].notValues

`[]string`

Negated values for the attribute.

### spec.rules[].matchAll

`bool` · optional (explicit presence)

Match every request. The manifest's word for the empty rule on the Istio
wire -- it never reaches the cluster; the rule it describes is emitted with
no matchers. Cannot be combined with from, to, or when. With ALLOW this
admits every request to the selected workloads; with DENY it is a total
lockout.

### spec.action

`string` · optional (explicit presence)

The action to take on a matched request. When omitted, the upstream default
ALLOW is applied. One of:
  ALLOW  — allow the request only if it matches a rule (the default).
  DENY   — deny the request if it matches any rule.
  AUDIT  — mark the request for auditing; does not affect allow/deny.
  CUSTOM — delegate the decision to the external authorizer named in `provider`
           (evaluated before ALLOW/DENY); requires `provider` to be set.
ALLOW/DENY/AUDIT/CUSTOM // external standard exception -- Istio AuthorizationPolicy.Action enum

Modeled as a closed string set (not a proto enum). Left unset to inherit
the upstream default (ALLOW); no Planton default is imposed.

- rule: {"string":{"in":["ALLOW","DENY","AUDIT","CUSTOM"]}}

### spec.provider

`KubernetesAuthorizationPolicyExtensionProvider`

The external authorizer to delegate to, used only with the CUSTOM action. Names
an extension provider declared in the mesh's MeshConfig. istiod enforces the
"CUSTOM-only" coupling at runtime (it is not a CRD-level validation), so this
spec carries `provider` independently of `action` to remain faithful to the
CRD's accepted surface.

### spec.provider.name

`string` · required

The name of a MeshConfig extension provider. At most one provider per workload
is supported by istiod.

- rule: {"string":{"minLen":"1"}}

## Validation Rules

- `authorization_policy.selector_xor_target_refs`: at most one of selector or target_refs may be set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesAuthorizationPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.authorization_policy_name` | `string` | Name of the created AuthorizationPolicy (equals metadata.name). |
| `status.outputs.namespace` | `string` | Namespace the AuthorizationPolicy was created in (the resolved spec.namespace). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.targetRefs[].name` | KubernetesGateway | `status.outputs.gateway_name` |

## See Also

- [Overview](../README.md)
