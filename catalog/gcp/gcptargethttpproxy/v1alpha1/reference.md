# GcpTargetHttpProxy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpTargetHttpProxySpec defines a Compute Engine target HTTP proxy — the
plaintext-HTTP frontend adapter of an Application Load Balancer (and of
Traffic Director meshes). A target HTTP proxy binds a forwarding rule (the
VIP) to a URL map (the routing brain): the forwarding rule delivers client
connections to the proxy, and the proxy consults the URL map to pick the
backend service or bucket for each request.

One kind, two scopes. With region empty the proxy is GLOBAL (the global
external ALB, the cross-region internal ALB, Traffic Director); with
region set it is REGIONAL (the regional external ALB and the regional
internal ALB), and every link in its chain must be regional too: a
regional URL map in front of regional backend services, and a regional
forwarding rule pointing at it. A proxy cannot move between scopes.

The proxy itself is deliberately thin — TLS termination lives on the
target HTTPS proxy, routing lives on the URL map, and traffic policy lives
on the backend service. The standard production pattern is a PAIR of
proxies sharing one frontend story: this HTTP proxy points at a
redirect-only URL map (http→https 301) while the HTTPS proxy serves the
real application.

The only mutable field is url_map (GCP updates it in place via a dedicated
setUrlMap call); everything else is immutable and forces
destroy-and-recreate.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTargetHttpProxy
metadata:
  name: my-sample-http-proxy
spec:
  # GCP project that owns the proxy.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Cloud-side name; omit to default to metadata.name.
  proxyName: web-http-frontend

  description: Port-80 frontend serving the http-to-https redirect URL map

  # The URL map the proxy routes through (reference a GcpUrlMap or provide a
  # self-link). The standard pattern points the HTTP proxy at a
  # redirect-only URL map while the HTTPS proxy serves the application.
  urlMap:
    value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/urlMaps/http-redirect

  # Idle client keep-alive (EXTERNAL_MANAGED only); omit for GCP's default.
  httpKeepAliveTimeoutSec: 610

  # DELETE keeps destroys real; PREVENT belongs on production frontends.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.proxyName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.region` | `string` |  |  |  |
| `spec.urlMap` | `string \| valueFrom` | yes |  | GcpUrlMap (`status.outputs.self_link`) |
| `spec.httpKeepAliveTimeoutSec` | `int32` |  |  |  |
| `spec.proxyBind` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the target HTTP proxy.
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
for the operator tracing a request path later. Immutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.region

`string`

The scope selector. Empty builds a GLOBAL target HTTP proxy (the global
external ALB, the cross-region internal ALB, Traffic Director); a region
name such as us-central1 builds a REGIONAL one (the regional external
ALB and the regional internal ALB). The URL map it references must live
in the same scope — and, for a regional proxy, the same region — and so
must the forwarding rule in front of it. Immutable: a proxy cannot move
between scopes or regions.

- rule: region must be a valid GCP region name such as us-central1, or empty for a global target HTTP proxy

### spec.urlMap

`string | valueFrom` · required

The URL map that decides where each request goes — the proxy's single
routing dependency. Reference a GcpUrlMap resource or provide a URL map
self-link directly. Required. A regional proxy can only point at a
regional URL map in its own region (a GcpUrlMap declared with the same
region). Mutable: GCP swaps it in place (a dedicated setUrlMap call),
so repointing a live frontend at a new routing table causes no
downtime.

- references: GcpUrlMap (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpUrlMap, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.httpKeepAliveTimeoutSec

`int32`

Seconds an idle client connection is kept open after a response while no
matching traffic flows (5-1200). Only honored by load balancers with the
EXTERNAL_MANAGED scheme (the envoy-based external ALBs, global and
regional), where the GCP default is 610; the classic EXTERNAL ALB
ignores it. Raise it above your clients' own keep-alive to avoid the
load balancer closing connections first. 0 means unset (GCP applies its
default). Immutable on both scopes: changing it destroys and recreates
the proxy.

- rule: http_keep_alive_timeout_sec must be between 5 and 1200 seconds (or 0 to let GCP apply its default)

### spec.proxyBind

`bool`

Bind the proxy to the private IPs of the Traffic Director mesh instead
of Google's edge. Only meaningful when the forwarding rule that
references this proxy uses the INTERNAL_SELF_MANAGED scheme (Traffic
Director); leave false for internet-facing load balancers. Global
proxies only — Traffic Director has no regional proxy, and the regional
resource carries no such argument. Immutable.

### spec.deletionPolicy

`string`

Deletion policy for the proxy — what happens when this resource is
destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the proxy is deleted (GCP refuses while a forwarding
               rule still references it, so a frontend cannot be torn
               down out from under its VIP)
  "PREVENT" -- destroy FAILS; protects a production frontend from
               accidental teardown
  "ABANDON" -- the proxy is removed from management but keeps
               serving in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `proxy_bind_global_only`: proxy_bind is a global-proxy (Traffic Director) lever — a regional target HTTP proxy has no mesh to bind to; clear region or remove proxy_bind

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpTargetHttpProxy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.self_link` | `string` | Self-link URI of the target HTTP proxy. This is the value a forwarding rule references as its target — the composition handle that puts a VIP in front of this proxy. Format: https://www.googleapis.com/compute/v1/projects/{project}/global/targetHttpProxies/{name} (a regional proxy's link carries regions/{region} in place of global). |
| `status.outputs.proxy_name` | `string` | Name of the proxy as it exists in GCP. |
| `status.outputs.proxy_id` | `string` | Server-assigned numeric ID of the proxy. |
| `status.outputs.fingerprint` | `string` | Server-computed fingerprint for optimistic concurrency control. |
| `status.outputs.region` | `string` | Region of a regional target HTTP proxy; empty for a global one. Downstream blocks read it to confirm scope compatibility (a regional link must point at a regional target in the same region), and the E2E verifier picks the regional or global API by it. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.urlMap` | GcpUrlMap | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)
