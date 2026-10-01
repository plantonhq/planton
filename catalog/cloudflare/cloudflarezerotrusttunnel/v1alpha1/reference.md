# CloudflareZeroTrustTunnel

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustTunnelSpec configures a Cloudflare Tunnel (cloudflared): a secure,
outbound-only connection from a private network to Cloudflare's edge. A tunnel exposes
private HTTP/TCP/SSH/RDP services via public hostnames (ingress rules) and/or makes
private IP ranges reachable to WARP clients (via CloudflareZeroTrustTunnelRoute). The
connector (cloudflared) authenticates with the tunnel token exported in the stack
outputs, so no inbound firewall ports are ever opened.

When managed remotely (config_src = cloudflare, the default), the ingress rules are
configured here as desired state. When managed locally (config_src = local), ingress
lives in a YAML file on the origin machine and is not set here. The remote ingress
configuration is provisioned as its own provider resource, so editing ingress never
recreates the tunnel.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustTunnel
metadata:
  name: test-tunnel
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: prod-tunnel
  configSrc: cloudflare
  ingress:
    - hostname: app.example.com
      service: "http://localhost:8080"
      path: "/api/.*"
    - hostname: ssh.example.com
      service: "ssh://localhost:22"
    - service: "http_status:404"
  originRequest:
    connectTimeout: 30
    noHappyEyeballs: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.configSrc` | `enum` |  | `cloudflare` |  |
| `spec.tunnelSecret` | `string` (sensitive) |  |  |  |
| `spec.ingress` | `[]CloudflareZeroTrustTunnelIngressRule` |  |  |  |
| `spec.ingress[].hostname` | `string` |  |  |  |
| `spec.ingress[].service` | `string` | yes |  |  |
| `spec.ingress[].path` | `string` |  |  |  |
| `spec.ingress[].originRequest` | `CloudflareZeroTrustTunnelOriginRequest` |  |  |  |
| `spec.ingress[].originRequest.access` | `CloudflareZeroTrustTunnelAccessConfig` |  |  |  |
| `spec.ingress[].originRequest.access.audTag` | `[]string \| valueFrom` | yes |  | CloudflareZeroTrustAccessApplication (`status.outputs.aud`) |
| `spec.ingress[].originRequest.access.teamName` | `string` | yes |  |  |
| `spec.ingress[].originRequest.access.required` | `bool` |  |  |  |
| `spec.ingress[].originRequest.caPool` | `string` |  |  |  |
| `spec.ingress[].originRequest.connectTimeout` | `int64` |  |  |  |
| `spec.ingress[].originRequest.disableChunkedEncoding` | `bool` |  |  |  |
| `spec.ingress[].originRequest.http2Origin` | `bool` |  |  |  |
| `spec.ingress[].originRequest.httpHostHeader` | `string` |  |  |  |
| `spec.ingress[].originRequest.keepAliveConnections` | `int64` |  |  |  |
| `spec.ingress[].originRequest.keepAliveTimeout` | `int64` |  |  |  |
| `spec.ingress[].originRequest.matchSniToHost` | `bool` |  |  |  |
| `spec.ingress[].originRequest.noHappyEyeballs` | `bool` |  |  |  |
| `spec.ingress[].originRequest.noTlsVerify` | `bool` |  |  |  |
| `spec.ingress[].originRequest.originServerName` | `string` |  |  |  |
| `spec.ingress[].originRequest.proxyType` | `string` |  |  |  |
| `spec.ingress[].originRequest.tcpKeepAlive` | `int64` |  |  |  |
| `spec.ingress[].originRequest.tlsTimeout` | `int64` |  |  |  |
| `spec.originRequest` | `CloudflareZeroTrustTunnelOriginRequest` |  |  |  |
| `spec.originRequest.access` | `CloudflareZeroTrustTunnelAccessConfig` |  |  |  |
| `spec.originRequest.access.audTag` | `[]string \| valueFrom` | yes |  | CloudflareZeroTrustAccessApplication (`status.outputs.aud`) |
| `spec.originRequest.access.teamName` | `string` | yes |  |  |
| `spec.originRequest.access.required` | `bool` |  |  |  |
| `spec.originRequest.caPool` | `string` |  |  |  |
| `spec.originRequest.connectTimeout` | `int64` |  |  |  |
| `spec.originRequest.disableChunkedEncoding` | `bool` |  |  |  |
| `spec.originRequest.http2Origin` | `bool` |  |  |  |
| `spec.originRequest.httpHostHeader` | `string` |  |  |  |
| `spec.originRequest.keepAliveConnections` | `int64` |  |  |  |
| `spec.originRequest.keepAliveTimeout` | `int64` |  |  |  |
| `spec.originRequest.matchSniToHost` | `bool` |  |  |  |
| `spec.originRequest.noHappyEyeballs` | `bool` |  |  |  |
| `spec.originRequest.noTlsVerify` | `bool` |  |  |  |
| `spec.originRequest.originServerName` | `string` |  |  |  |
| `spec.originRequest.proxyType` | `string` |  |  |  |
| `spec.originRequest.tcpKeepAlive` | `int64` |  |  |  |
| `spec.originRequest.tlsTimeout` | `int64` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID that owns this tunnel.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.name

`string` · required

A user-friendly name for the tunnel (shown in the Zero Trust dashboard).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"100"}}

### spec.configSrc

`enum` · optional (explicit presence)

Where the tunnel's configuration is managed. `cloudflare` (the default) manages the
ingress rules below as desired state from the control plane; `local` leaves ingress
to a cloudflared YAML file on the origin and manages only the tunnel object + token.

- default: `cloudflare`

Allowed values (use exactly as shown):

- `cloudflare_zero_trust_tunnel_config_source_unspecified`
- `local` -- Ingress lives in a cloudflared YAML file on the origin machine.
- `cloudflare` -- Ingress is managed remotely as desired state (the default for this component).

### spec.tunnelSecret

`string` · sensitive

Optional secret used to run a locally-managed tunnel. The RAW value is a
base64 string encoding at least 32 bytes -- but no value-content rule is
enforced here, because this field is sensitive: consuming platforms store
a managed-secret REFERENCE in its place (plaintext is rejected at
create/update), and a content rule written for the raw value could never
match the stored reference. Cloudflare validates the decoded secret at
apply. Omit to let Cloudflare generate one (recommended); the run token
is always available in `status.outputs.tunnel_token` regardless.

### spec.ingress

`[]CloudflareZeroTrustTunnelIngressRule`

Public-hostname ingress rules, evaluated top to bottom. Each rule maps a hostname
(and optional path) to a local service. The final rule MUST be a catch-all (a
service with no hostname, for example `http_status:404`). Only valid when
config_src = cloudflare.

### spec.ingress[].hostname

`string`

Public hostname this rule matches (for example app.example.com). Leave empty ONLY
on the final catch-all rule. A DNS record for this hostname must CNAME to the
tunnel's `status.outputs.tunnel_cname`.

- rule: {"string":{"maxLen":"255"}}

### spec.ingress[].service

`string` · required

Protocol and address of the local service, for example http://localhost:8080,
tcp://10.0.0.5:22, ssh://10.0.0.5:22, or an HTTP status response like
http_status:404 (used for the catch-all rule).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.ingress[].path

`string`

Optional path filter: only requests whose path matches route to this service
(for example "/api/.*").

### spec.ingress[].originRequest

`CloudflareZeroTrustTunnelOriginRequest`

Optional per-rule origin connection settings, overriding the tunnel-level defaults.

### spec.ingress[].originRequest.access

`CloudflareZeroTrustTunnelAccessConfig`

Require Cloudflare Access authentication for L7 requests to the matched hostname(s).

### spec.ingress[].originRequest.access.audTag

`[]string | valueFrom` · required

Audience (AUD) tags of the Access applications allowed to reach this hostname. Each
is a literal AUD tag, or a reference to a CloudflareZeroTrustAccessApplication
resource (whose `aud` output is used) — so an Access app and the tunnel ingress it
protects compose as an explicit graph edge.

- references: CloudflareZeroTrustAccessApplication (`status.outputs.aud`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareZeroTrustAccessApplication, name: <that resource's name>, fieldPath: status.outputs.aud}} -- a bare string does not parse

### spec.ingress[].originRequest.access.teamName

`string` · required

The Zero Trust organization (team) name that owns the Access applications.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.ingress[].originRequest.access.required

`bool`

Deny traffic that has not satisfied Access authorization.

### spec.ingress[].originRequest.caPool

`string`

Path to a certificate authority (CA) bundle for the origin's certificate, used when
the origin certificate is not signed by Cloudflare.

### spec.ingress[].originRequest.connectTimeout

`int64`

Seconds to wait when establishing a new TCP connection to the origin (excludes TLS).

- rule: connect_timeout must be between 0 and 3600 seconds

### spec.ingress[].originRequest.disableChunkedEncoding

`bool`

Disable chunked transfer encoding (useful for some WSGI origins).

### spec.ingress[].originRequest.http2Origin

`bool`

Attempt to connect to the origin over HTTP/2 (origin must serve HTTPS).

### spec.ingress[].originRequest.httpHostHeader

`string`

Override the HTTP Host header sent to the local service.

### spec.ingress[].originRequest.keepAliveConnections

`int64`

Maximum number of idle keepalive connections to the origin (does not cap total
concurrent connections).

- rule: keep_alive_connections must be between 0 and 10000

### spec.ingress[].originRequest.keepAliveTimeout

`int64`

Seconds after which an idle keepalive connection to the origin is discarded.

- rule: keep_alive_timeout must be between 0 and 3600 seconds

### spec.ingress[].originRequest.matchSniToHost

`bool`

Auto-configure the expected origin server name from the request hostname.

### spec.ingress[].originRequest.noHappyEyeballs

`bool`

Disable the "happy eyeballs" IPv4/IPv6 fallback algorithm.

### spec.ingress[].originRequest.noTlsVerify

`bool`

Disable TLS verification of the origin certificate (accepts any origin cert).

### spec.ingress[].originRequest.originServerName

`string`

Hostname cloudflared should expect on the origin server's certificate.

### spec.ingress[].originRequest.proxyType

`string`

Proxy type for TCP-over-HTTP services: empty for the regular proxy, or "socks" for
a SOCKS5 proxy.

- rule: proxy_type must be empty or 'socks'

### spec.ingress[].originRequest.tcpKeepAlive

`int64`

Seconds between TCP keepalive packets on the connection to the origin.

- rule: tcp_keep_alive must be between 0 and 3600 seconds

### spec.ingress[].originRequest.tlsTimeout

`int64`

Seconds to complete a TLS handshake to an HTTPS origin.

- rule: tls_timeout must be between 0 and 3600 seconds

### spec.originRequest

`CloudflareZeroTrustTunnelOriginRequest`

Default origin connection settings applied to every ingress rule unless a rule
overrides them. Only meaningful when config_src = cloudflare.

### spec.originRequest.access

`CloudflareZeroTrustTunnelAccessConfig`

Require Cloudflare Access authentication for L7 requests to the matched hostname(s).

### spec.originRequest.access.audTag

`[]string | valueFrom` · required

Audience (AUD) tags of the Access applications allowed to reach this hostname. Each
is a literal AUD tag, or a reference to a CloudflareZeroTrustAccessApplication
resource (whose `aud` output is used) — so an Access app and the tunnel ingress it
protects compose as an explicit graph edge.

- references: CloudflareZeroTrustAccessApplication (`status.outputs.aud`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareZeroTrustAccessApplication, name: <that resource's name>, fieldPath: status.outputs.aud}} -- a bare string does not parse

### spec.originRequest.access.teamName

`string` · required

The Zero Trust organization (team) name that owns the Access applications.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.originRequest.access.required

`bool`

Deny traffic that has not satisfied Access authorization.

### spec.originRequest.caPool

`string`

Path to a certificate authority (CA) bundle for the origin's certificate, used when
the origin certificate is not signed by Cloudflare.

### spec.originRequest.connectTimeout

`int64`

Seconds to wait when establishing a new TCP connection to the origin (excludes TLS).

- rule: connect_timeout must be between 0 and 3600 seconds

### spec.originRequest.disableChunkedEncoding

`bool`

Disable chunked transfer encoding (useful for some WSGI origins).

### spec.originRequest.http2Origin

`bool`

Attempt to connect to the origin over HTTP/2 (origin must serve HTTPS).

### spec.originRequest.httpHostHeader

`string`

Override the HTTP Host header sent to the local service.

### spec.originRequest.keepAliveConnections

`int64`

Maximum number of idle keepalive connections to the origin (does not cap total
concurrent connections).

- rule: keep_alive_connections must be between 0 and 10000

### spec.originRequest.keepAliveTimeout

`int64`

Seconds after which an idle keepalive connection to the origin is discarded.

- rule: keep_alive_timeout must be between 0 and 3600 seconds

### spec.originRequest.matchSniToHost

`bool`

Auto-configure the expected origin server name from the request hostname.

### spec.originRequest.noHappyEyeballs

`bool`

Disable the "happy eyeballs" IPv4/IPv6 fallback algorithm.

### spec.originRequest.noTlsVerify

`bool`

Disable TLS verification of the origin certificate (accepts any origin cert).

### spec.originRequest.originServerName

`string`

Hostname cloudflared should expect on the origin server's certificate.

### spec.originRequest.proxyType

`string`

Proxy type for TCP-over-HTTP services: empty for the regular proxy, or "socks" for
a SOCKS5 proxy.

- rule: proxy_type must be empty or 'socks'

### spec.originRequest.tcpKeepAlive

`int64`

Seconds between TCP keepalive packets on the connection to the origin.

- rule: tcp_keep_alive must be between 0 and 3600 seconds

### spec.originRequest.tlsTimeout

`int64`

Seconds to complete a TLS handshake to an HTTPS origin.

- rule: tls_timeout must be between 0 and 3600 seconds

## Validation Rules

- `tunnel.ingress_requires_remote_config`: ingress rules require config_src 'cloudflare' (remote management); a 'local' tunnel is configured by a YAML file on the origin
- `tunnel.ingress_last_rule_is_catch_all`: the last ingress rule must be a catch-all (a service with no hostname, e.g. http_status:404)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustTunnel, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.tunnel_id` | `string` | The Cloudflare-assigned UUID of the tunnel. Referenced by routes (CloudflareZeroTrustTunnelRoute) and used to build the CNAME target. |
| `status.outputs.tunnel_cname` | `string` | The CNAME target for public hostnames served by this tunnel (<tunnel_id>.cfargotunnel.com). Point a CloudflareDnsRecord CNAME at this value to route a public hostname through the tunnel. |
| `status.outputs.tunnel_token` | `string` (sensitive) | The connector run token. cloudflared uses this to authenticate and establish the tunnel (for example `cloudflared tunnel run --token <token>`). Sensitive. |
| `status.outputs.tunnel_status` | `string` | The tunnel status: inactive, degraded, healthy, or down. |
| `status.outputs.account_tag` | `string` | The Cloudflare account tag the tunnel belongs to. |
| `status.outputs.created_on` | `string` | RFC3339 timestamp of when the tunnel was created. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.ingress[].originRequest.access.audTag` | CloudflareZeroTrustAccessApplication | `status.outputs.aud` |
| `spec.originRequest.access.audTag` | CloudflareZeroTrustAccessApplication | `status.outputs.aud` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareWorker | `spec.vpcNetworks[].tunnelId` | `status.outputs.tunnel_id` |
| CloudflareZeroTrustTunnelRoute | `spec.tunnelId` | `status.outputs.tunnel_id` |

## See Also

- [Overview](../README.md)
