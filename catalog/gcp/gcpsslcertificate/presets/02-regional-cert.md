# Regional ALB Certificate

A self-managed certificate scoped to one region — for regional external and internal Application Load Balancer proxies, which cannot reference global certificates.

## When to Use

- Regional external ALB frontends
- Internal ALBs serving private traffic with TLS (a very common self-managed use — managed certificates need public DNS, which internal ALBs do not have)
- Data-residency postures that keep TLS material regional

## Key Configuration Choices

- **`region`** — selects the regional API collection; the certificate must live in the proxy's own region
- **Internal ALB fit** — internal load balancers cannot complete Google-managed DNS validation, so self-managed certificates (often from a private CA) are the standard choice

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the certificate will live | GCP Console or `GcpProject` outputs |
| `us-central1` | The region of the proxy that will attach this certificate | Your regional ALB design |
| PEM bodies | Your real chain and unencrypted key | Your CA |

## Remix Notes

- The `region` output confirms scope — regional proxies reject certificates from other regions or global scope
- Clear `region` to create the global twin for global external HTTPS proxies

## Related Presets

- **01-imported-cert** — The global variant
- **03-rotation-versioned-name** — Version the GCP name to make rotations explicit
