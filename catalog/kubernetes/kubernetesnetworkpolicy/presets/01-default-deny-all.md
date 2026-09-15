# Default Deny All

This preset creates the namespace-wide security baseline: it denies ALL inbound and outbound traffic for every pod in the namespace. NetworkPolicies are additive allows with no deny rule, so "deny everything" is expressed by selecting all pods (`pod_selector: { match_all: true }`) while allowing nothing (no rules). Every subsequent policy in the namespace is then a targeted exception that only widens.

## When to Use

- As the first policy applied to any namespace you intend to segment — deny by default, then allow per flow
- Meeting compliance requirements that mandate default-deny network posture
- Before onboarding workloads into a shared cluster, so nothing is reachable until explicitly allowed

## Key Configuration Choices

- **`pod_selector: { match_all: true }`** — selects ALL pods in the namespace; this is what makes the policy namespace-wide
- **Explicit `policy_types: [ingress, egress]`** — with no rules present, Kubernetes would infer ingress only; egress isolation must be declared explicitly because there is no egress rule to infer it from
- **No rules at all** — selected in both directions with nothing allowed means all traffic to and from these pods is denied
- **Denying egress also blocks DNS** — pair this preset with **04-allow-dns-egress** or nearly everything in the namespace will fail name resolution

> Enforcement requires a NetworkPolicy-implementing CNI (Calico, Cilium, cloud CNIs with enforcement enabled). On clusters whose CNI ignores NetworkPolicies — including default kind clusters — this object exists but all traffic still flows.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | The namespace to lock down — a NetworkPolicy governs only pods in its own namespace | Your namespace management |

## Related Presets

- **02-allow-same-namespace** — re-allow pod-to-pod traffic within the namespace
- **03-allow-from-namespace** — allow specific cross-namespace callers
- **04-allow-dns-egress** — the DNS allow this preset almost always needs alongside it
