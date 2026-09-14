# Allow Same Namespace

This preset isolates a namespace from the rest of the cluster while keeping it open internally: all pods in the namespace accept inbound traffic from all other pods in the SAME namespace, and nothing else. It is the default-deny-ingress shape (`pod_selector: { match_all: true }`, ingress governed) plus exactly one allow — a peer whose `pod_selector: { match_all: true }` means "every pod in this namespace".

## When to Use

- Namespace-per-team or namespace-per-environment clusters where traffic inside a namespace is trusted but cross-namespace traffic is not
- A softer starting point than full default-deny: intra-namespace flows keep working while you enumerate external callers
- Any namespace whose services call each other freely but should not be reachable from other tenants

## Key Configuration Choices

- **Top-level `pod_selector: { match_all: true }`** — the policy governs ALL pods in the namespace
- **Peer `pod_selector: { match_all: true }` with no `namespace_selector`** — a pod selector alone always scopes to the policy's OWN namespace, so selecting every pod here means "all pods in this namespace" (not "all pods in the cluster")
- **No `ports`** — traffic from allowed sources is permitted on all ports; add a `ports` list to narrow
- **`policy_types: [ingress]`** — egress is untouched; pods can still initiate outbound traffic anywhere. Add the default-deny + DNS presets to govern egress too
- **Additive with other policies** — cross-namespace callers can be granted later with separate policies (e.g. 03-allow-from-namespace) without editing this one

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | The namespace to isolate — intra-namespace traffic stays allowed, everything else inbound is denied | Your namespace management |

## Related Presets

- **01-default-deny-all** — the stricter baseline: deny even intra-namespace traffic
- **03-allow-from-namespace** — grant specific cross-namespace callers on top of this
