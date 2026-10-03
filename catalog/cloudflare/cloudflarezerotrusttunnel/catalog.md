# Cloudflare Zero Trust Tunnel

Provisions a Cloudflare Tunnel (cloudflared): a secure, outbound-only connection from a private network to Cloudflare's edge. A tunnel exposes private HTTP/TCP/SSH/RDP services via public hostnames (ingress rules) and makes private IP ranges reachable to WARP clients (via `CloudflareZeroTrustTunnelRoute`) -- without opening a single inbound firewall port. The connector authenticates with the run token exported in the outputs.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Tunnel** -- the cloudflared tunnel object and its run token
- **Ingress configuration** -- created only when `configSrc` is `cloudflare`; provisioned as its own provider resource, so editing ingress never recreates the tunnel

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Cloudflare Tunnel edit access. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Zero Trust enabled** -- the account must have Cloudflare Zero Trust (Cloudflare One) set up.
- **A connector to run** -- after deploy, run `cloudflared` somewhere on the private network with the run token from the outputs.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust Tunnel**, and click **Deploy**. The creation wizard captures the tunnel identity (account, name, configuration source, and an optional managed-secret run secret), the public-hostname ingress rules, and the tunnel-level origin connection defaults (including a Cloudflare Access block). Start from the **Publish a private app on a public hostname** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustTunnel
metadata:
  name: prod-connector
  org: acme-corp
  env: prod
spec:
  accountId: a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4
  name: prod-connector
  configSrc: cloudflare
  ingress:
    - hostname: app.example.com
      service: http://localhost:8080
    - service: http_status:404
```

```shell
planton apply -f cloudflare-zero-trust-tunnel.yaml
```

This exposes `app.example.com` through the tunnel to a local service on port 8080, with a catch-all 404 for everything else. An Infra Job tracks the provisioning in real time.

### InfraChart

Protect an ingress hostname with an Access application deployed in the same composition, wiring the Access block with ValueFromRef:

```yaml
spec:
  originRequest:
    access:
      audTag:
        - valueFrom:
            kind: CloudflareZeroTrustAccessApplication
            name: internal-dashboard
            fieldPath: status.outputs.aud
      teamName: acme-corp
      required: true
```

The InfraPipeline resolves the dependency graph, provisions the Access application first, then configures the tunnel with the resolved AUD tag.

## Key Configuration

These are the most important decisions when configuring a tunnel. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Account (`accountId`)** -- The account that owns the tunnel. Immutable -- changing it replaces the tunnel.

**Configuration Source (`configSrc`)** -- `cloudflare` (default) manages ingress here as desired state; `local` leaves ingress to a cloudflared YAML file on the origin. Immutable -- changing it replaces the tunnel.

**Ingress (`ingress`)** -- Public-hostname rules, evaluated top to bottom. The final rule must be a catch-all (a service with no hostname, e.g. `http_status:404`). A DNS CNAME for each hostname must point at the tunnel's CNAME target.

**Origin Defaults (`originRequest`)** -- Connection settings applied to every ingress rule unless a rule overrides them. The Cloudflare Access sub-block (`originRequest.access`) requires an Access JWT on the matched hostnames, referencing `CloudflareZeroTrustAccessApplication` resources.

**Tunnel Secret (`tunnelSecret`)** -- An optional, reference-only managed secret for a locally-managed connector. Leave empty to let Cloudflare generate one.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareZeroTrustAccessApplication** (optional) | `originRequest.access.audTag[]` (also per-rule `ingress[].originRequest.access.audTag[]`) | `status.outputs.aud` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `tunnel_id` | The Cloudflare-assigned UUID of the tunnel | Referenced by a CloudflareZeroTrustTunnelRoute's `tunnelId`, or a Worker's `vpcNetworks` binding |
| `tunnel_cname` | The CNAME target (`<tunnel_id>.cfargotunnel.com`) | Point a CloudflareDnsRecord CNAME at it to route a public hostname through the tunnel |
| `tunnel_token` | The connector run token (sensitive) | `cloudflared tunnel run --token <token>` on the private network |

`status.outputs` also carries `tunnel_status` (inactive/degraded/healthy/down), `account_tag`, and `created_on`.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Public hostname** -- an ingress rule maps a public hostname to a local HTTP service, with a catch-all 404. Start from the **Publish a private app on a public hostname** preset.

**Access-protected** -- the origin Access block requires an Access application's JWT so only authorized users reach a hostname. Start from the **Access-protected admin hostname** preset.

**Private network connector** -- pair the tunnel with `CloudflareZeroTrustTunnelRoute` resources to make private CIDRs reachable to WARP clients; no ingress rules needed. Start from the **Private-network connector (for WARP access)** preset.

## Works With

- [**Cloudflare Zero Trust Tunnel Route**](/infra-catalog/cloudflare-zero-trust-tunnel-route) -- advertises private CIDRs through this tunnel
- [**Cloudflare Zero Trust Tunnel Virtual Network**](/infra-catalog/cloudflare-zero-trust-tunnel-virtual-network) -- the routing segment a route belongs to
- [**Cloudflare Zero Trust Access Application**](/infra-catalog/cloudflare-zero-trust-access-application) -- the Access apps an ingress Access block references
- [**Cloudflare DNS Record**](/infra-catalog/cloudflare-dns-record) -- the CNAME that points a public hostname at the tunnel
