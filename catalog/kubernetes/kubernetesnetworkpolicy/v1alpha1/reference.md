# KubernetesNetworkPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesNetworkPolicySpec** declares which network traffic is allowed to and
from a set of pods — the in-cluster firewall. A NetworkPolicy selects pods with
`pod_selector` and then ALLOWS the traffic its rules describe; everything not
allowed by some policy is denied once the pod is selected by any policy in that
direction. Policies are additive: multiple policies selecting the same pods
combine by union, and there is no "deny" rule — isolation comes from selecting
pods while allowing little.

The two canonical shapes fall out of this model:
- Default-deny: an empty `pod_selector` (all pods in the namespace) with
  `policy_types: [ingress, egress]` and no rules — everything is denied.
- Targeted allow: select specific pods and enumerate the peers/ports allowed.

Selectors compose with Planton workloads through the documented label contract:
every workload kind stamps the `app` label (set to the workload's
`metadata.name`) on its pods as immutable selection identity and exports the
full set as its `selector_labels` output — so `match_labels: {app: <name>}`
targets one workload's pods.

IMPORTANT: NetworkPolicy objects are only ENFORCED by a CNI that implements
them (Calico, Cilium, and cloud CNIs with policy enforcement enabled). On a
cluster whose CNI ignores them (including default kind/kindnet clusters), the
object exists but all traffic still flows.

The spec covers the complete networking/v1 NetworkPolicySpec surface.

## Example

```yaml
# Full-surface test manifest for the offline plan proofs. Exercises explicit
# policy types, all three peer forms (pod selector, namespace selector with
# expressions, IP block with except), the AND'd pod+namespace peer, named and
# numeric ports, and a port range.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesNetworkPolicy
metadata:
  name: test-network-policy
spec:
  namespace:
    value: default
  name: test-network-policy
  labels:
    team: platform-engineering
  pod_selector:
    match_labels:
      app: test-app
  policy_types:
    - ingress
    - egress
  ingress_rules:
    - from:
        - pod_selector:
            match_labels:
              app: frontend
        - namespace_selector:
            match_labels:
              kubernetes.io/metadata.name: monitoring
          pod_selector:
            match_expressions:
              - key: app
                operator: In
                values:
                  - prometheus
                  - grafana
      ports:
        - protocol: TCP
          port: "8080"
        - protocol: TCP
          port: metrics
  egress_rules:
    - to:
        - namespace_selector: { match_all: true }
      ports:
        - protocol: UDP
          port: "53"
        - protocol: TCP
          port: "53"
    - to:
        - ip_block:
            cidr: 0.0.0.0/0
            except:
              - 169.254.169.254/32
      ports:
        - protocol: TCP
          port: "30000"
          end_port: 32767
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` |  |  | KubernetesNamespace (`spec.name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.podSelector` | `KubernetesNetworkPolicyLabelSelector` |  |  |  |
| `spec.podSelector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.podSelector.matchExpressions` | `[]KubernetesNetworkPolicyLabelSelectorRequirement` |  |  |  |
| `spec.podSelector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.podSelector.matchExpressions[].operator` | `string` |  |  |  |
| `spec.podSelector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.podSelector.matchAll` | `bool` |  |  |  |
| `spec.policyTypes` | `[]enum` |  |  |  |
| `spec.ingressRules` | `[]KubernetesNetworkPolicyIngressRule` |  |  |  |
| `spec.ingressRules[].from` | `[]KubernetesNetworkPolicyPeer` |  |  |  |
| `spec.ingressRules[].from[].podSelector` | `KubernetesNetworkPolicyLabelSelector` |  |  |  |
| `spec.ingressRules[].from[].podSelector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.ingressRules[].from[].podSelector.matchExpressions` | `[]KubernetesNetworkPolicyLabelSelectorRequirement` |  |  |  |
| `spec.ingressRules[].from[].podSelector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.ingressRules[].from[].podSelector.matchExpressions[].operator` | `string` |  |  |  |
| `spec.ingressRules[].from[].podSelector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.ingressRules[].from[].podSelector.matchAll` | `bool` |  |  |  |
| `spec.ingressRules[].from[].namespaceSelector` | `KubernetesNetworkPolicyLabelSelector` |  |  |  |
| `spec.ingressRules[].from[].namespaceSelector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.ingressRules[].from[].namespaceSelector.matchExpressions` | `[]KubernetesNetworkPolicyLabelSelectorRequirement` |  |  |  |
| `spec.ingressRules[].from[].namespaceSelector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.ingressRules[].from[].namespaceSelector.matchExpressions[].operator` | `string` |  |  |  |
| `spec.ingressRules[].from[].namespaceSelector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.ingressRules[].from[].namespaceSelector.matchAll` | `bool` |  |  |  |
| `spec.ingressRules[].from[].ipBlock` | `KubernetesNetworkPolicyIpBlock` |  |  |  |
| `spec.ingressRules[].from[].ipBlock.cidr` | `string` | yes |  |  |
| `spec.ingressRules[].from[].ipBlock.except` | `[]string` |  |  |  |
| `spec.ingressRules[].ports` | `[]KubernetesNetworkPolicyPort` |  |  |  |
| `spec.ingressRules[].ports[].protocol` | `enum` |  |  |  |
| `spec.ingressRules[].ports[].port` | `string` |  |  |  |
| `spec.ingressRules[].ports[].endPort` | `int32` |  |  |  |
| `spec.egressRules` | `[]KubernetesNetworkPolicyEgressRule` |  |  |  |
| `spec.egressRules[].to` | `[]KubernetesNetworkPolicyPeer` |  |  |  |
| `spec.egressRules[].to[].podSelector` | `KubernetesNetworkPolicyLabelSelector` |  |  |  |
| `spec.egressRules[].to[].podSelector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.egressRules[].to[].podSelector.matchExpressions` | `[]KubernetesNetworkPolicyLabelSelectorRequirement` |  |  |  |
| `spec.egressRules[].to[].podSelector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.egressRules[].to[].podSelector.matchExpressions[].operator` | `string` |  |  |  |
| `spec.egressRules[].to[].podSelector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.egressRules[].to[].podSelector.matchAll` | `bool` |  |  |  |
| `spec.egressRules[].to[].namespaceSelector` | `KubernetesNetworkPolicyLabelSelector` |  |  |  |
| `spec.egressRules[].to[].namespaceSelector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.egressRules[].to[].namespaceSelector.matchExpressions` | `[]KubernetesNetworkPolicyLabelSelectorRequirement` |  |  |  |
| `spec.egressRules[].to[].namespaceSelector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.egressRules[].to[].namespaceSelector.matchExpressions[].operator` | `string` |  |  |  |
| `spec.egressRules[].to[].namespaceSelector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.egressRules[].to[].namespaceSelector.matchAll` | `bool` |  |  |  |
| `spec.egressRules[].to[].ipBlock` | `KubernetesNetworkPolicyIpBlock` |  |  |  |
| `spec.egressRules[].to[].ipBlock.cidr` | `string` | yes |  |  |
| `spec.egressRules[].to[].ipBlock.except` | `[]string` |  |  |  |
| `spec.egressRules[].ports` | `[]KubernetesNetworkPolicyPort` |  |  |  |
| `spec.egressRules[].ports[].protocol` | `enum` |  |  |  |
| `spec.egressRules[].ports[].port` | `string` |  |  |  |
| `spec.egressRules[].ports[].endPort` | `int32` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom`

The namespace the policy lives in — a NetworkPolicy governs only pods in its
own namespace. Accepts a literal namespace name or a reference to a
KubernetesNamespace resource. When omitted, the policy lands in the
cluster's `default` namespace.

- references: KubernetesNamespace (`spec.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.name

`string` · required

The name of the NetworkPolicy (its `metadata.name` in the cluster).
Must be a valid DNS subdomain: lowercase alphanumeric characters, hyphens,
and dots, at most 253 characters.

- rule: Name must be a valid DNS subdomain (lowercase alphanumeric, hyphens, and dots, no leading/trailing dots or hyphens)
- rule: {"string":{"minLen":"1","maxLen":"253"}}

### spec.labels

`map<string, string>`

Additional labels to apply to the NetworkPolicy object.
Merged with the standard Planton governance labels.

### spec.annotations

`map<string, string>`

Annotations to apply to the NetworkPolicy object.

### spec.podSelector

`KubernetesNetworkPolicyLabelSelector`

Selects the pods this policy applies to, within the policy's namespace.
Omitting the selector, or writing `match_all: true`, selects ALL pods in
the namespace — the default-deny building block. To target one Planton
workload, match on its `app` label: `match_labels: {app: <workload-name>}`.

- rule: match_all: true selects everything — drop match_labels and match_expressions, or drop match_all to select by label
- rule: a selector must say what it selects — set match_all: true to select every pod (or namespace), or name match_labels / match_expressions

### spec.podSelector.matchLabels

`map<string, string>`

Exact-match label requirements: every key/value pair listed must be present,
e.g. {"app": "checkout"}.

### spec.podSelector.matchExpressions

`[]KubernetesNetworkPolicyLabelSelectorRequirement`

Set-based label requirements, for selections exact-match cannot express
(key existence, value-in-set).

- rule: values must be non-empty for In/NotIn and empty for Exists/DoesNotExist

### spec.podSelector.matchExpressions[].key

`string` · required

The label key the requirement applies to.

- rule: {"string":{"minLen":"1"}}

### spec.podSelector.matchExpressions[].operator

`string`

The relationship between the key and the values: "In" (value must be one of
`values`), "NotIn" (must not be), "Exists" (key present, `values` must be
empty), or "DoesNotExist" (key absent, `values` must be empty).

- rule: {"string":{"in":["In","NotIn","Exists","DoesNotExist"]}}

### spec.podSelector.matchExpressions[].values

`[]string`

The values compared against the label's value. Required (non-empty) for
In/NotIn; must be empty for Exists/DoesNotExist.

### spec.podSelector.matchAll

`bool` · optional (explicit presence)

Select everything: every pod in scope for a pod selector, every namespace
for a namespace selector. The manifest's word for the empty selector on
the Kubernetes wire -- it never reaches the cluster; the selector it
describes is emitted with no criteria. Cannot be combined with
match_labels or match_expressions.

### spec.policyTypes

`[]enum`

Which directions this policy governs. When omitted, Kubernetes infers:
ingress is always included, and egress is included only when egress rules
are present. Set explicitly whenever the intent is isolation — an
egress-only policy MUST say `[egress]` (or it also isolates ingress), and a
deny-all-egress policy MUST say `[egress]` with no egress rules (there is
no rule to infer the direction from).

- rule: {"repeated":{"maxItems":"2","items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `kubernetes_network_policy_type_unspecified` -- Unspecified.
- `ingress` -- Ingress: the policy governs traffic INTO the selected pods.
- `egress` -- Egress: the policy governs traffic OUT OF the selected pods.

### spec.ingressRules

`[]KubernetesNetworkPolicyIngressRule`

The ingress allow rules. Each rule is an independent OR: traffic is allowed
into the selected pods if it matches ANY rule's sources AND ports. An empty
list (with ingress in policy_types) denies all inbound traffic to the
selected pods. A rule with empty `from` allows from ALL sources; a rule
with empty `ports` allows on ALL ports.

### spec.ingressRules[].from

`[]KubernetesNetworkPolicyPeer`

The allowed sources. Empty means ALL sources (traffic not restricted by
origin — the rule then only restricts by port).

- rule: ip_block cannot be combined with pod_selector or namespace_selector in the same peer — use separate peers
- rule: a peer must specify at least one of pod_selector, namespace_selector, or ip_block

### spec.ingressRules[].from[].podSelector

`KubernetesNetworkPolicyLabelSelector`

Selects pods by label. `match_all: true` selects ALL pods (in the policy's
namespace, or in the namespaces selected by namespace_selector when both
are set).

- rule: match_all: true selects everything — drop match_labels and match_expressions, or drop match_all to select by label
- rule: a selector must say what it selects — set match_all: true to select every pod (or namespace), or name match_labels / match_expressions

### spec.ingressRules[].from[].podSelector.matchLabels

`map<string, string>`

Exact-match label requirements: every key/value pair listed must be present,
e.g. {"app": "checkout"}.

### spec.ingressRules[].from[].podSelector.matchExpressions

`[]KubernetesNetworkPolicyLabelSelectorRequirement`

Set-based label requirements, for selections exact-match cannot express
(key existence, value-in-set).

- rule: values must be non-empty for In/NotIn and empty for Exists/DoesNotExist

### spec.ingressRules[].from[].podSelector.matchExpressions[].key

`string` · required

The label key the requirement applies to.

- rule: {"string":{"minLen":"1"}}

### spec.ingressRules[].from[].podSelector.matchExpressions[].operator

`string`

The relationship between the key and the values: "In" (value must be one of
`values`), "NotIn" (must not be), "Exists" (key present, `values` must be
empty), or "DoesNotExist" (key absent, `values` must be empty).

- rule: {"string":{"in":["In","NotIn","Exists","DoesNotExist"]}}

### spec.ingressRules[].from[].podSelector.matchExpressions[].values

`[]string`

The values compared against the label's value. Required (non-empty) for
In/NotIn; must be empty for Exists/DoesNotExist.

### spec.ingressRules[].from[].podSelector.matchAll

`bool` · optional (explicit presence)

Select everything: every pod in scope for a pod selector, every namespace
for a namespace selector. The manifest's word for the empty selector on
the Kubernetes wire -- it never reaches the cluster; the selector it
describes is emitted with no criteria. Cannot be combined with
match_labels or match_expressions.

### spec.ingressRules[].from[].namespaceSelector

`KubernetesNetworkPolicyLabelSelector`

Selects namespaces by label (e.g. the automatic
`kubernetes.io/metadata.name: <name>` label every namespace carries).
`match_all: true` selects ALL namespaces — the cluster-wide-allow building
block.

- rule: match_all: true selects everything — drop match_labels and match_expressions, or drop match_all to select by label
- rule: a selector must say what it selects — set match_all: true to select every pod (or namespace), or name match_labels / match_expressions

### spec.ingressRules[].from[].namespaceSelector.matchLabels

`map<string, string>`

Exact-match label requirements: every key/value pair listed must be present,
e.g. {"app": "checkout"}.

### spec.ingressRules[].from[].namespaceSelector.matchExpressions

`[]KubernetesNetworkPolicyLabelSelectorRequirement`

Set-based label requirements, for selections exact-match cannot express
(key existence, value-in-set).

- rule: values must be non-empty for In/NotIn and empty for Exists/DoesNotExist

### spec.ingressRules[].from[].namespaceSelector.matchExpressions[].key

`string` · required

The label key the requirement applies to.

- rule: {"string":{"minLen":"1"}}

### spec.ingressRules[].from[].namespaceSelector.matchExpressions[].operator

`string`

The relationship between the key and the values: "In" (value must be one of
`values`), "NotIn" (must not be), "Exists" (key present, `values` must be
empty), or "DoesNotExist" (key absent, `values` must be empty).

- rule: {"string":{"in":["In","NotIn","Exists","DoesNotExist"]}}

### spec.ingressRules[].from[].namespaceSelector.matchExpressions[].values

`[]string`

The values compared against the label's value. Required (non-empty) for
In/NotIn; must be empty for Exists/DoesNotExist.

### spec.ingressRules[].from[].namespaceSelector.matchAll

`bool` · optional (explicit presence)

Select everything: every pod in scope for a pod selector, every namespace
for a namespace selector. The manifest's word for the empty selector on
the Kubernetes wire -- it never reaches the cluster; the selector it
describes is emitted with no criteria. Cannot be combined with
match_labels or match_expressions.

### spec.ingressRules[].from[].ipBlock

`KubernetesNetworkPolicyIpBlock`

Allows a CIDR range, with optional carve-outs. For traffic to/from
addresses OUTSIDE the cluster (external services, VPC ranges, the internet)
— cluster-internal pod IPs are ephemeral and should be matched with
selectors, never CIDRs.

### spec.ingressRules[].from[].ipBlock.cidr

`string` · required

The allowed CIDR, e.g. "10.100.0.0/16" or "2001:db8::/64". "0.0.0.0/0"
allows all IPv4 — pair it with `except` to allow "everything but".

- rule: cidr must be a valid CIDR (e.g. "10.100.0.0/16", "2001:db8::/64")
- rule: {"required":true}

### spec.ingressRules[].from[].ipBlock.except

`[]string`

CIDRs carved OUT of the allow — each must be a sub-range of `cidr` (the API
rejects out-of-range exceptions). E.g. allow "0.0.0.0/0" except
"169.254.169.254/32" (the cloud metadata endpoint).

- rule: {"repeated":{"items":{"cel":[{"id":"except.format","message":"each except entry must be a valid CIDR inside the allowed cidr range","expression":"this.isIpPrefix()"}]}}}

### spec.ingressRules[].ports

`[]KubernetesNetworkPolicyPort`

The allowed destination ports on the selected pods. Empty means ALL ports.

- rule: a numeric port must be in the range 1-65535
- rule: end_port requires port to be set to a NUMERIC port (a named port cannot anchor a range)
- rule: end_port must be greater than or equal to port

### spec.ingressRules[].ports[].protocol

`enum` · optional (explicit presence)

The protocol this rule matches.
Default: TCP

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `kubernetes_network_policy_protocol_unspecified` -- Unspecified. Defaults to TCP.
- `TCP` -- TCP — the default.
- `UDP` -- UDP — required for DNS allow rules (port 53).
- `SCTP` -- SCTP — requires cluster SCTP support.

### spec.ingressRules[].ports[].port

`string`

The destination port: a number ("5432") or a named container port
("metrics") as declared on the target pods. Omit to match ALL ports for
the protocol.

- rule: port must be a port number ("5432") or a named container port ("metrics" — a valid IANA service name)

### spec.ingressRules[].ports[].endPort

`int32`

The inclusive upper bound of a port RANGE starting at `port`. Requires
`port` to be numeric (a named port cannot anchor a range) and must be >=
that number. E.g. port "30000" + end_port 32767 allows the whole node-port
range.

- rule: end_port must be 0 (single port) or in the range 1-65535

### spec.egressRules

`[]KubernetesNetworkPolicyEgressRule`

The egress allow rules. Each rule is an independent OR: traffic is allowed
out of the selected pods if it matches ANY rule's destinations AND ports.
An empty list (with egress in policy_types) denies all outbound traffic
from the selected pods — including DNS, which is why deny-all-egress
policies are usually paired with an allow-DNS rule (UDP+TCP port 53 to the
cluster DNS pods).

### spec.egressRules[].to

`[]KubernetesNetworkPolicyPeer`

The allowed destinations. Empty means ALL destinations.

- rule: ip_block cannot be combined with pod_selector or namespace_selector in the same peer — use separate peers
- rule: a peer must specify at least one of pod_selector, namespace_selector, or ip_block

### spec.egressRules[].to[].podSelector

`KubernetesNetworkPolicyLabelSelector`

Selects pods by label. `match_all: true` selects ALL pods (in the policy's
namespace, or in the namespaces selected by namespace_selector when both
are set).

- rule: match_all: true selects everything — drop match_labels and match_expressions, or drop match_all to select by label
- rule: a selector must say what it selects — set match_all: true to select every pod (or namespace), or name match_labels / match_expressions

### spec.egressRules[].to[].podSelector.matchLabels

`map<string, string>`

Exact-match label requirements: every key/value pair listed must be present,
e.g. {"app": "checkout"}.

### spec.egressRules[].to[].podSelector.matchExpressions

`[]KubernetesNetworkPolicyLabelSelectorRequirement`

Set-based label requirements, for selections exact-match cannot express
(key existence, value-in-set).

- rule: values must be non-empty for In/NotIn and empty for Exists/DoesNotExist

### spec.egressRules[].to[].podSelector.matchExpressions[].key

`string` · required

The label key the requirement applies to.

- rule: {"string":{"minLen":"1"}}

### spec.egressRules[].to[].podSelector.matchExpressions[].operator

`string`

The relationship between the key and the values: "In" (value must be one of
`values`), "NotIn" (must not be), "Exists" (key present, `values` must be
empty), or "DoesNotExist" (key absent, `values` must be empty).

- rule: {"string":{"in":["In","NotIn","Exists","DoesNotExist"]}}

### spec.egressRules[].to[].podSelector.matchExpressions[].values

`[]string`

The values compared against the label's value. Required (non-empty) for
In/NotIn; must be empty for Exists/DoesNotExist.

### spec.egressRules[].to[].podSelector.matchAll

`bool` · optional (explicit presence)

Select everything: every pod in scope for a pod selector, every namespace
for a namespace selector. The manifest's word for the empty selector on
the Kubernetes wire -- it never reaches the cluster; the selector it
describes is emitted with no criteria. Cannot be combined with
match_labels or match_expressions.

### spec.egressRules[].to[].namespaceSelector

`KubernetesNetworkPolicyLabelSelector`

Selects namespaces by label (e.g. the automatic
`kubernetes.io/metadata.name: <name>` label every namespace carries).
`match_all: true` selects ALL namespaces — the cluster-wide-allow building
block.

- rule: match_all: true selects everything — drop match_labels and match_expressions, or drop match_all to select by label
- rule: a selector must say what it selects — set match_all: true to select every pod (or namespace), or name match_labels / match_expressions

### spec.egressRules[].to[].namespaceSelector.matchLabels

`map<string, string>`

Exact-match label requirements: every key/value pair listed must be present,
e.g. {"app": "checkout"}.

### spec.egressRules[].to[].namespaceSelector.matchExpressions

`[]KubernetesNetworkPolicyLabelSelectorRequirement`

Set-based label requirements, for selections exact-match cannot express
(key existence, value-in-set).

- rule: values must be non-empty for In/NotIn and empty for Exists/DoesNotExist

### spec.egressRules[].to[].namespaceSelector.matchExpressions[].key

`string` · required

The label key the requirement applies to.

- rule: {"string":{"minLen":"1"}}

### spec.egressRules[].to[].namespaceSelector.matchExpressions[].operator

`string`

The relationship between the key and the values: "In" (value must be one of
`values`), "NotIn" (must not be), "Exists" (key present, `values` must be
empty), or "DoesNotExist" (key absent, `values` must be empty).

- rule: {"string":{"in":["In","NotIn","Exists","DoesNotExist"]}}

### spec.egressRules[].to[].namespaceSelector.matchExpressions[].values

`[]string`

The values compared against the label's value. Required (non-empty) for
In/NotIn; must be empty for Exists/DoesNotExist.

### spec.egressRules[].to[].namespaceSelector.matchAll

`bool` · optional (explicit presence)

Select everything: every pod in scope for a pod selector, every namespace
for a namespace selector. The manifest's word for the empty selector on
the Kubernetes wire -- it never reaches the cluster; the selector it
describes is emitted with no criteria. Cannot be combined with
match_labels or match_expressions.

### spec.egressRules[].to[].ipBlock

`KubernetesNetworkPolicyIpBlock`

Allows a CIDR range, with optional carve-outs. For traffic to/from
addresses OUTSIDE the cluster (external services, VPC ranges, the internet)
— cluster-internal pod IPs are ephemeral and should be matched with
selectors, never CIDRs.

### spec.egressRules[].to[].ipBlock.cidr

`string` · required

The allowed CIDR, e.g. "10.100.0.0/16" or "2001:db8::/64". "0.0.0.0/0"
allows all IPv4 — pair it with `except` to allow "everything but".

- rule: cidr must be a valid CIDR (e.g. "10.100.0.0/16", "2001:db8::/64")
- rule: {"required":true}

### spec.egressRules[].to[].ipBlock.except

`[]string`

CIDRs carved OUT of the allow — each must be a sub-range of `cidr` (the API
rejects out-of-range exceptions). E.g. allow "0.0.0.0/0" except
"169.254.169.254/32" (the cloud metadata endpoint).

- rule: {"repeated":{"items":{"cel":[{"id":"except.format","message":"each except entry must be a valid CIDR inside the allowed cidr range","expression":"this.isIpPrefix()"}]}}}

### spec.egressRules[].ports

`[]KubernetesNetworkPolicyPort`

The allowed destination ports. Empty means ALL ports.

- rule: a numeric port must be in the range 1-65535
- rule: end_port requires port to be set to a NUMERIC port (a named port cannot anchor a range)
- rule: end_port must be greater than or equal to port

### spec.egressRules[].ports[].protocol

`enum` · optional (explicit presence)

The protocol this rule matches.
Default: TCP

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `kubernetes_network_policy_protocol_unspecified` -- Unspecified. Defaults to TCP.
- `TCP` -- TCP — the default.
- `UDP` -- UDP — required for DNS allow rules (port 53).
- `SCTP` -- SCTP — requires cluster SCTP support.

### spec.egressRules[].ports[].port

`string`

The destination port: a number ("5432") or a named container port
("metrics") as declared on the target pods. Omit to match ALL ports for
the protocol.

- rule: port must be a port number ("5432") or a named container port ("metrics" — a valid IANA service name)

### spec.egressRules[].ports[].endPort

`int32`

The inclusive upper bound of a port RANGE starting at `port`. Requires
`port` to be numeric (a named port cannot anchor a range) and must be >=
that number. E.g. port "30000" + end_port 32767 allows the whole node-port
range.

- rule: end_port must be 0 (single port) or in the range 1-65535

## Validation Rules

- `policy_types_distinct`: policy_types entries must be distinct
- `ingress_rules_require_ingress_type`: ingress rules are present but policy_types does not include ingress — the rules would be ignored
- `egress_rules_require_egress_type`: egress rules are present but policy_types does not include egress — the rules would be ignored

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesNetworkPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.network_policy_name` | `string` | The name of the NetworkPolicy object as created in the cluster. |
| `status.outputs.namespace` | `string` | The namespace the NetworkPolicy was created in. |
| `status.outputs.policy_types` | `string` | The directions this policy governs, as deployed — "Ingress", "Egress", or "Ingress,Egress". Includes inferred types when the spec omitted policy_types. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)
