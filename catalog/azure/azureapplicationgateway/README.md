# AzureApplicationGateway

Azure Application Gateway is the Layer 7 (HTTP/HTTPS) load balancer and
reverse proxy: host- and path-based routing, TLS termination including
mutual TLS, request/response rewriting, redirects, custom health probes,
TCP/TLS layer-4 proxying, Private Link exposure, and Web Application
Firewall enforcement through a referenced policy.

The gateway bundles its sub-objects (frontends, ports, listeners, pools,
settings, rules, path maps, probes, certificates, profiles, redirects,
rewrites) because Azure configures them as one atomic resource wired
together by name. What other resources need to reach is exported as
name-keyed map outputs -- pool membership joins from the member side.

## When to Use

Use AzureApplicationGateway when you need:

- **L7 routing** -- one public host fanning out by host name or URI path
  to multiple backend pools
- **TLS termination with Key Vault certificates** that renew without
  touching the gateway, or end-to-end TLS with private-CA backend trust
- **WAF protection** (WAF_V2 SKU) via a composable
  `AzureWebApplicationFirewallPolicy` -- gateway-wide, per listener, or
  per route
- **Mutual TLS** -- client-certificate verification through SSL profiles
- **HTTP rewriting** -- header edits and URL rewrites driven by
  conditions
- **TCP/TLS proxying** -- the layer-4 listener/backend/rule trio for
  non-HTTP traffic
- **Private entry** -- internal-only frontends and Private Link exposure

For non-HTTP layer-4 load balancing at scale, `AzureLoadBalancer` is the
lighter, faster choice.

## The Traffic Path

```
frontend (public IP | private address) + port
  -> listener (protocol, host names, TLS cert, SSL profile, WAF override)
    -> request routing rule (priority)
      -> backend pool + backend HTTP settings     (BASIC_ROUTING)
      -> url path map -> per-path rules            (PATH_BASED_ROUTING)
      -> redirect configuration                    (redirects)
```

## Key Configuration

- **SKU**: `BASIC` (1-2 instances, reduced features), `STANDARD_V2` (the
  workhorse), `WAF_V2` (adds WAF policy enforcement). v1 SKUs are retired
  in Azure and not modeled.
- **Sizing**: `capacity` (fixed) XOR `autoscale` (bounds) -- exactly one.
- **Frontends**: public (an `AzurePublicIp` reference) or private (an
  address in the gateway's dedicated subnet).
- **Certificates**: Key Vault references (production -- renewals
  propagate) or inline PFX; Key Vault requires a user-assigned identity
  with GET on the vault's secrets.
- **Deploys run 15-25 minutes** -- Azure's slowest networking resource.

## Outputs

| Output | Description |
|--------|-------------|
| `application_gateway_id` | The gateway's ARM ID (AGIC and management seams) |
| `application_gateway_name` | The gateway's name |
| `backend_address_pool_ids` | Name-keyed pool ARM IDs -- what NICs and scale sets join |
| `frontend_ip_configuration_ids` | Name-keyed frontend ARM IDs |
| `private_ip_address` / `private_ip_addresses` | Private frontends' addresses (public addresses live on the referenced `AzurePublicIp`) |

## Related Resources

- **AzureSubnet** -- the dedicated gateway subnet
- **AzurePublicIp** -- public frontend addresses
- **AzureWebApplicationFirewallPolicy** -- the WAF rule set (WAF_V2)
- **AzureKeyVaultCertificate** -- TLS certificates that renew in place
- **AzureUserAssignedIdentity** -- the vault-access identity
- **AzureNetworkInterface** / **AzureVirtualMachineScaleSet** -- backend
  pool members (joined member-side)
- **AzureAksCluster** -- AGIC references the gateway by ID

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
