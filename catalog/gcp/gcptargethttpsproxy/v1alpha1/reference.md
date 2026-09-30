# GcpTargetHttpsProxy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpTargetHttpsProxySpec defines a global Compute Engine target HTTPS proxy
— the TLS-termination node of a global external Application Load Balancer
(and of Traffic Director meshes). A target HTTPS proxy binds a global
forwarding rule (the VIP) to a URL map (the routing brain) and owns
everything about the client-facing TLS handshake: which certificates are
presented, which TLS policy constrains ciphers and versions, whether QUIC
(HTTP/3) is negotiated, and whether TLS 1.3 0-RTT early data is accepted.

Certificates attach through exactly one of three mechanisms:
  - ssl_certificates: classic Compute Engine SSL certificates (the
    Google-managed GcpManagedSslCertificate kind, or self-managed compute
    certificates), up to 15 per proxy.
  - certificate_manager_certificates: Certificate Manager certificates —
    only honored by the cross-region internal ALB (INTERNAL_MANAGED).
  - certificate_map: a Certificate Manager certificate map that selects
    the certificate by SNI hostname at scale — only honored by external
    ALBs (EXTERNAL / EXTERNAL_MANAGED); required beyond ~15 certs.

Traffic Director proxies (INTERNAL_SELF_MANAGED) skip client certificates
entirely and drive TLS through server_tls_policy instead.

url_map, certificates, certificate_map, ssl_policy, server_tls_policy, and
quic_override all update in place via dedicated API calls; name,
description, keep-alive, early data, and proxy_bind are immutable.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTargetHttpsProxy
metadata:
  name: my-sample-https-proxy
spec:
  # GCP project that owns the proxy.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Cloud-side name; omit to default to metadata.name.
  proxyName: web-https-frontend

  description: Port-443 frontend terminating TLS for www.example.com

  # The URL map the proxy routes through (reference a GcpUrlMap or provide a
  # self-link).
  urlMap:
    value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/urlMaps/web-routing

  # Exactly one certificate mechanism. Here: classic compute SSL
  # certificates (reference GcpManagedSslCertificate resources or provide
  # self-links).
  sslCertificates:
    - value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/sslCertificates/web-cert

  # Enforce modern TLS instead of GCP's permissive default policy.
  sslPolicy:
    value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/sslPolicies/modern-tls

  # QUIC (HTTP/3) negotiation: NONE lets Google decide.
  quicOverride: NONE

  # DELETE keeps destroys real; PREVENT belongs on production TLS
  # frontends.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.proxyName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.urlMap` | `string \| valueFrom` | yes |  | GcpUrlMap (`status.outputs.self_link`) |
| `spec.sslCertificates` | `[]string \| valueFrom` |  |  | GcpManagedSslCertificate (`status.outputs.self_link`), GcpSslCertificate (`status.outputs.self_link`) |
| `spec.certificateManagerCertificates` | `[]string \| valueFrom` |  |  | GcpCertManagerCert (`status.outputs.certificate_name`) |
| `spec.certificateMap` | `string` |  |  |  |
| `spec.sslPolicy` | `string \| valueFrom` |  |  | GcpSslPolicy (`status.outputs.self_link`) |
| `spec.serverTlsPolicy` | `string \| valueFrom` |  |  |  |
| `spec.quicOverride` | `string` |  |  |  |
| `spec.tlsEarlyData` | `string` |  |  |  |
| `spec.httpKeepAliveTimeoutSec` | `int32` |  |  |  |
| `spec.proxyBind` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the target HTTPS proxy.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing it destroys and recreates the proxy.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.proxyName

`string`

Name of the proxy in GCP. Must be 1-63 characters: lowercase letters,
digits, and hyphens; must start with a letter and end with a letter or
digit. If not specified, defaults to metadata.name.
Immutable: changing it destroys and recreates the proxy, briefly
breaking every forwarding rule that references the old self_link.

- rule: proxy_name must be RFC1035-compliant: 1-63 lowercase letters, digits, or hyphens; must start with a letter and end with a letter or digit

### spec.description

`string`

What this proxy fronts and which forwarding rule points at it — write it
for the operator tracing a TLS incident later. Immutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.urlMap

`string | valueFrom` · required

The URL map that decides where each decrypted request goes — the proxy's
single routing dependency. Reference a GcpUrlMap resource or provide a
URL map self-link directly. Required. Mutable: GCP swaps it in place (a
dedicated setUrlMap call), so repointing a live frontend at a new
routing table causes no downtime.

- references: GcpUrlMap (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpUrlMap, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.sslCertificates

`[]string | valueFrom`

Compute Engine SSL certificates presented to clients (1-15). Reference
GcpManagedSslCertificate resources (the default kind), self-managed
GcpSslCertificate resources via an explicit valueFrom.kind (both are
declared candidates, so the self_link output fills in), or provide
SSL certificate self-links directly — both certificate kinds share one
API collection and attach identically. The load balancer picks the
certificate matching the client's SNI hostname.
Not honored by Traffic Director (INTERNAL_SELF_MANAGED) proxies — use
server_tls_policy there. Mutually exclusive with
certificate_manager_certificates and certificate_map. Mutable: GCP swaps
the list in place (setSslCertificates), which is how zero-downtime
certificate rotation works — attach the replacement before detaching the
old one.

- references: GcpManagedSslCertificate (`status.outputs.self_link`), GcpSslCertificate (`status.outputs.self_link`)
- rule: {"repeated":{"maxItems":"15"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpManagedSslCertificate, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.certificateManagerCertificates

`[]string | valueFrom`

Certificate Manager certificates presented to clients — only honored by
the cross-region internal ALB (INTERNAL_MANAGED); external ALBs use
certificate_map instead. Reference GcpCertManagerCert resources or
provide certificate resource names directly, in the form
projects/{project}/locations/{location}/certificates/{name} (a
//certificatemanager.googleapis.com/ prefix is also accepted). Mutually
exclusive with ssl_certificates and certificate_map. Mutable.

- references: GcpCertManagerCert (`status.outputs.certificate_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCertManagerCert, name: <that resource's name>, fieldPath: status.outputs.certificate_name}} -- a bare string does not parse

### spec.certificateMap

`string`

A Certificate Manager certificate map that selects the served
certificate by SNI hostname — the mechanism for serving many domains
(SaaS custom domains) beyond the 15-certificate list limit. Only honored
by external ALBs (EXTERNAL / EXTERNAL_MANAGED). Format:
//certificatemanager.googleapis.com/projects/{project}/locations/{location}/certificateMaps/{name}.
Mutually exclusive with ssl_certificates and
certificate_manager_certificates. Mutable.

- rule: {"string":{"maxLen":"1024"}}

### spec.sslPolicy

`string | valueFrom`

The SSL policy constraining TLS versions and cipher suites for client
handshakes. Reference a GcpSslPolicy resource or provide an SSL policy
self-link directly (e.g.
https://www.googleapis.com/compute/v1/projects/{project}/global/sslPolicies/{name}).
If not set, GCP applies its permissive default policy (min TLS 1.0,
COMPATIBLE profile) — set one to enforce modern TLS for compliance.
Mutable: GCP swaps it in place (setSslPolicy).

- references: GcpSslPolicy (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSslPolicy, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.serverTlsPolicy

`string | valueFrom`

A network security ServerTlsPolicy resource that configures server-side
TLS — the mTLS mechanism: it can demand and validate client
certificates. Applies to global proxies behind EXTERNAL /
EXTERNAL_MANAGED / INTERNAL_SELF_MANAGED forwarding rules; for Traffic
Director this is the ONLY TLS lever (ssl_certificates are ignored).
Format: projects/{project}/locations/global/serverTlsPolicies/{name}.
If left blank, no server-side TLS policy applies. Mutable — and
clearable: removing it PATCHes the proxy back to no policy.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.quicOverride

`string` · optional (explicit presence)

QUIC (HTTP/3) negotiation policy. NONE lets Google decide (currently
enables QUIC), ENABLE forces QUIC negotiation on, DISABLE turns it off.
GCP default: NONE. Mutable.

- rule: quic_override must be one of NONE, ENABLE, or DISABLE

### spec.tlsEarlyData

`string`

TLS 1.3 0-RTT "early data" policy — lets a resuming client send the
first HTTP request inside the TLS handshake itself (zero effective round
trips, over TCP and QUIC/HTTP-3). Early data is replayable by design, so
the modes trade latency against replay safety: STRICT accepts it only
for safe methods (GET/HEAD) with no query parameters, PERMISSIVE for all
requests, UNRESTRICTED additionally skips rejecting non-idempotent
replays (only for services that tolerate replays), DISABLED turns it
off. Empty lets GCP apply its default (DISABLED). Immutable: changing it
destroys and recreates the proxy.

- rule: tls_early_data must be one of STRICT, PERMISSIVE, UNRESTRICTED, or DISABLED

### spec.httpKeepAliveTimeoutSec

`int32`

Seconds an idle client connection is kept open after a response while no
matching traffic flows (5-1200). Only honored by load balancers with the
EXTERNAL_MANAGED scheme (the envoy-based global external ALB), where the
GCP default is 610; the classic EXTERNAL ALB ignores it. Raise it above
your clients' own keep-alive to avoid the load balancer closing
connections first. 0 means unset (GCP applies its default). Immutable:
changing it destroys and recreates the proxy.

- rule: http_keep_alive_timeout_sec must be between 5 and 1200 seconds (or 0 to let GCP apply its default)

### spec.proxyBind

`bool`

Bind the proxy to the private IPs of the Traffic Director mesh instead
of Google's edge. Only meaningful when the forwarding rule that
references this proxy uses the INTERNAL_SELF_MANAGED scheme (Traffic
Director); leave false for internet-facing load balancers. Immutable.

### spec.deletionPolicy

`string`

Deletion policy for the proxy — what happens when this resource is
destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the proxy is deleted (GCP refuses while a forwarding
               rule still references it, so a TLS frontend cannot be
               torn down out from under its VIP)
  "PREVENT" -- destroy FAILS; protects a production TLS frontend from
               accidental teardown
  "ABANDON" -- the proxy is removed from management but keeps
               terminating TLS in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `single_certificate_source`: choose one certificate mechanism: ssl_certificates (classic compute certificates), certificate_manager_certificates (cross-region internal ALB), or certificate_map (SNI-scale external ALB) — GCP rejects combinations

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpTargetHttpsProxy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.self_link` | `string` | Self-link URI of the target HTTPS proxy. This is the value a global forwarding rule references as its target — the composition handle that puts a VIP in front of this proxy. Format: https://www.googleapis.com/compute/v1/projects/{project}/global/targetHttpsProxies/{name} |
| `status.outputs.proxy_name` | `string` | Name of the proxy as it exists in GCP. |
| `status.outputs.proxy_id` | `string` | Server-assigned numeric ID of the proxy. |
| `status.outputs.fingerprint` | `string` | Server-computed fingerprint for optimistic concurrency control. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.urlMap` | GcpUrlMap | `status.outputs.self_link` |
| `spec.sslCertificates` | GcpManagedSslCertificate | `status.outputs.self_link` |
| `spec.sslCertificates` | GcpSslCertificate | `status.outputs.self_link` |
| `spec.certificateManagerCertificates` | GcpCertManagerCert | `status.outputs.certificate_name` |
| `spec.sslPolicy` | GcpSslPolicy | `status.outputs.self_link` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpGlobalForwardingRule | `spec.target` | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)
