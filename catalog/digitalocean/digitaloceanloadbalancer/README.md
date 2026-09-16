# DigitalOcean Load Balancer

A DigitalOcean load balancer described once in a Planton manifest: regional and global types, sizing by slug or scaling units, forwarding rules with TLS termination or passthrough, health checks with full threshold tuning, cookie-based sticky sessions, backend targeting by Droplet references or tag, VPC and subnet placement, an LB-level firewall, HTTPS redirect, PROXY protocol, backend keepalive, idle-timeout, TLS cipher policy, project placement, bring-your-own-IP, and the global balancer's domains, target balancers, CDN, and regional failover.

## What this component models

The spec maps one-to-one onto DigitalOcean's load balancer:

| Spec field | What it controls |
|---|---|
| `loadBalancerName` | The balancer's name in DigitalOcean (unique per account) |
| `region` | Data-center region for REGIONAL / REGIONAL_NETWORK; forbidden for GLOBAL |
| `type` | `REGIONAL` (default), `REGIONAL_NETWORK`, or `GLOBAL` |
| `size` / `sizeUnit` | Capacity: three slugs (`lb-small`/`lb-medium`/`lb-large`) or 1–200 scaling units; mutually exclusive |
| `vpc` | Optional VPC placement — a literal UUID or a reference to a `DigitalOceanVpc`; omit to use the region's default |
| `forwardingRules` | Regional routing: entry/target port and protocol, optional TLS passthrough, certificate by name |
| `glbSettings` | Global routing: target protocol/port, region priorities, failover threshold, CDN; mutually exclusive with `forwardingRules` |
| `domains` | Domains that ingress a GLOBAL balancer |
| `targetLoadBalancerIds` | Regional balancers a GLOBAL balancer routes to |
| `healthCheck` | Backend probes (http/https/tcp) with interval, timeout, and thresholds |
| `stickySessions` | Cookie affinity (`cookies` + name + TTL, or `none`) |
| `dropletIds` / `dropletTag` | Backend membership; mutually exclusive |
| `firewall` | Source allow/deny rules (`ip:` / `cidr:`) |
| `redirectHttpToHttps` | Redirect port 80 to 443 |
| `enableProxyProtocol` / `enableBackendKeepalive` | Connection behavior toward backends |
| `httpIdleTimeoutSeconds` | Idle HTTP connection lifetime (0 = DigitalOcean's 60s default) |
| `tlsCipherPolicy` | `DEFAULT` or `STRONG` |
| `network` / `networkStack` | `EXTERNAL`/`INTERNAL` visibility and `IPV4`/`DUALSTACK`; create-only, never read back |
| `projectId` | DigitalOcean project UUID (literal until the Project kind is forged) |
| `subnetUuid` | DigitalOcean-managed subnet placement; requires `vpc` |
| `ip` | Bring-your-own IP at create time |

## Quick start

A regional HTTP balancer in an existing VPC, targeting every Droplet tagged `web`:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanLoadBalancer
metadata:
  name: web-lb
  org: acme-corp
  env: prod
spec:
  loadBalancerName: web-lb
  region: nyc3
  vpc:
    valueFrom:
      kind: DigitalOceanVpc
      name: app-network
      fieldPath: status.outputs.vpc_id
  forwardingRules:
    - entryPort: 80
      entryProtocol: http
      targetPort: 8080
      targetProtocol: http
  healthCheck:
    port: 8080
    protocol: http
    path: /health
  dropletTag: web
```

```shell
planton apply -f web-lb.yaml
```

## Outputs

Both provisioners export the identical output set:

| Output | Description |
|---|---|
| `load_balancer_id` | The balancer UUID (import id for `digitalocean_loadbalancer`) |
| `ip` | Public IPv4 address |
| `urn` | `do:loadbalancer:<id>` |
| `ipv6` | IPv6 address when `networkStack` is `DUALSTACK` |

## Behavior worth knowing

- **Certificates are identified by name**, not UUID. Let's Encrypt renewals rotate the UUID; `DigitalOceanCertificate.status.outputs.certificate_id` carries the stable name.
- **`size` and `sizeUnit` are the cost knob.** `lb-small` equals 1 unit; `lb-medium` is 3; `lb-large` is 6 — you pay per unit. Past that, only `sizeUnit` (up to 200) applies.
- **GLOBAL balancers have no region and no forwarding rules.** They route through `glbSettings`, `domains`, and `targetLoadBalancerIds`.
- **A `dropletTag` does not have to exist yet.** DigitalOcean stores it as a selector and attaches Droplets as they pick up the tag; unlike a firewall, it never rejects an unknown tag. Load-balancer names are unique per account.
- **`network`, `networkStack`, and `tlsCipherPolicy` are write-only.** The API never reports them back, so import leaves them empty and drift on them is invisible.
- **The pinned Pulumi SDK (v4.53.0) cannot express `subnetUuid` or BYOIP `ip`.** The Pulumi module fails loudly if they are set; Terraform wires them. See the [GUIDE](GUIDE.md).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
