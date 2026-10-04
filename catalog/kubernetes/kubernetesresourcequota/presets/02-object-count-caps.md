# Object Count Caps

This preset caps how MANY objects a namespace may hold — pods, Services, and PersistentVolumeClaims — without touching compute. It is the safest quota to introduce on a live namespace: object counts constrain nothing that pods must declare, so naive pod creation keeps working and no LimitRange companion is needed (none is created — `limit_defaults` is absent).

## When to Use

- The first quota on any shared namespace: contain sprawl before governing compute
- Capping cost-bearing objects — PVCs claim storage, and Services of type LoadBalancer provision provider resources (add `services.loadbalancers` to cap those specifically)
- Namespaces hosting operators or CI systems that can create objects in unbounded loops

## Key Configuration Choices

- **Counts only, no compute caps** — this is what makes the preset safe to apply anywhere: pods that omit requests/limits are NOT rejected, because no `requests.*`/`limits.*` entry exists to require them
- **No `limit_defaults`** — only the ResourceQuota is created; the `limit_range_name` output is empty
- **`pods: "100"`** — counts all non-terminal pods; note a Deployment scale-up that hits the cap stalls silently in its ReplicaSet events, so watch `kubectl describe resourcequota` (used vs hard) after rollout
- **Extendable vocabulary** — the same map takes `secrets`, `configmaps`, `services.nodeports`, `services.loadbalancers`, and the generic `count/<resource>.<group>` form for any countable object, including CRDs

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | The namespace whose object counts to cap | Your namespace management |

## Related Presets

- **01-team-namespace-governed** — add compute caps with the safe LimitRange pairing once counts are in place
- **03-besteffort-guard** — a scoped pod cap targeting only naive (BestEffort) pods
