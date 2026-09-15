# Allow DNS Egress

This preset allows all pods in the namespace to resolve DNS: egress on UDP and TCP port 53 to the cluster DNS pods in `kube-system`. It is the mandatory companion to any deny-all-egress posture — a policy that denies egress also blocks DNS, and workloads then fail with confusing timeouts and `no such host` errors that don't obviously point at the network policy.

## When to Use

- Alongside **01-default-deny-all** (or any egress-denying policy) in the same namespace — apply the two together
- As the first allow when building an egress allowlist: DNS first, then each real destination as its own policy

## Key Configuration Choices

- **Explicit `policy_types: [egress]`** — an egress-only policy MUST declare its direction. When `policy_types` is omitted, Kubernetes infers ingress ALWAYS (and egress only when egress rules exist); leaving it off here would additionally select every pod for ingress with no ingress rules — accidental deny-all-inbound
- **`pod_selector: { match_all: true }`** — the allow applies to ALL pods in the namespace; DNS is a namespace-wide need
- **`namespace_selector` on `kube-system`** — uses the automatic `kubernetes.io/metadata.name` label; the peer allows all pods in that namespace, which covers the cluster DNS deployment wherever it schedules
- **Both UDP and TCP 53** — DNS is UDP first, but falls back to TCP for large responses (and some resolvers prefer it); allowing only UDP causes rare, maddening failures
- **Additive** — this policy grants DNS only; it does not weaken the deny-all it accompanies

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | The egress-restricted namespace whose pods need DNS | Your namespace management |

## Related Presets

- **01-default-deny-all** — the deny-all-egress baseline this preset exists to accompany
- **03-allow-from-namespace** — the same namespace-selector technique, applied to ingress
