---
title: "Load Balancer"
description: "Load Balancer deployment documentation"
icon: "package"
order: 100
componentName: "scalewayloadbalancer"
---

# Scaleway Load Balancer

Deploys a managed Scaleway Load Balancer that bundles a Flexible IP, the LB appliance, backend server pools, frontend listeners, and optional TLS certificates into a single declarative resource. The Load Balancer distributes incoming Layer 4/7 traffic across backend servers based on configurable forwarding rules, health checks, and load-balancing algorithms.

## What Gets Created

When you deploy a ScalewayLoadBalancer resource, Planton provisions:

- **Flexible IP** — a dedicated `loadbalancers.Ip` public IPv4 address with independent lifecycle that survives LB replacement
- **Load Balancer** — a `loadbalancers.LoadBalancer` appliance of the specified type, with optional Private Network attachment
- **Backend(s)** — one or more `loadbalancers.Backend` server pools with health checks and load-balancing configuration
- **Frontend(s)** — one or more `loadbalancers.Frontend` listeners that route inbound traffic to backends
- **Certificate(s)** — zero or more `loadbalancers.Certificate` resources (Let's Encrypt or custom PEM) for HTTPS frontends
- **Scaleway Tags** — resource kind, name, organization, and environment labels applied as flat `key=value` tags

## Prerequisites

- **Scaleway credentials** configured via environment variables or Planton provider config
- **A target zone** — Load Balancers are zonal resources (e.g., `fr-par-1`, `nl-ams-1`, `pl-waw-1`)
- **Backend server IPs** — at least one server IP per backend, reachable from the LB (public IPs or private IPs if attached to a Private Network)
- **(Optional) A Private Network** — either a literal Private Network UUID or an Planton-managed ScalewayPrivateNetwork resource whose output can be referenced via `valueFrom`

## Quick Start

Create a file `load-balancer.yaml`:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayLoadBalancer
metadata:
  name: my-lb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayLoadBalancer.my-lb
spec:
  zone: fr-par-1
  type: LB-S
  backends:
    - name: web
      serverIps:
        - "10.0.1.10"
      forwardPort: 80
      forwardProtocol: http
  frontends:
    - name: http
      inboundPort: 80
      backendName: web
```

Deploy:

```shell
planton apply -f load-balancer.yaml
```

This creates a small Load Balancer in `fr-par-1` with a single HTTP backend and frontend on port 80. The public IP is available in stack outputs as `lb_ip_address`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `zone` | `string` | Scaleway zone where the Load Balancer is created (e.g., `"fr-par-1"`, `"nl-ams-1"`). Must be within the same region as any attached Private Network. Cannot be changed after creation. | Required |
| `type` | `string` | Load Balancer type that determines bandwidth and pricing tier. Options: `"LB-S"` (up to 400 Mbps), `"LB-GP-M"` (up to 4 Gbps), `"LB-GP-L"` (up to 8 Gbps), `"LB-GP-XL"` (up to 10 Gbps). Recommended default: `"LB-S"`. Can be changed after creation. | Required |
| `backends` | `Backend[]` | One or more backend server pools. Each backend defines a named set of servers, a forwarding port and protocol, health check rules, and load-balancing configuration. | Required, min 1 item |
| `backends[].name` | `string` | Unique name identifying this backend. Frontends reference backends by this name. | Required |
| `backends[].serverIps` | `string[]` | IP addresses of backend servers. Use private IPs when the LB is attached to a Private Network; use public IPs otherwise. | Required, min 1 item |
| `backends[].forwardPort` | `int` | Port on backend servers that receives forwarded traffic (e.g., `80`, `443`, `8080`). | Required |
| `backends[].forwardProtocol` | `string` | Protocol for LB-to-backend communication. Options: `"http"`, `"https"`, `"tcp"`. Recommended default: `"http"`. | Required |
| `frontends` | `Frontend[]` | One or more frontend listeners. Each frontend listens on a port and routes traffic to a named backend. | Required, min 1 item |
| `frontends[].name` | `string` | Unique name identifying this frontend. | Required |
| `frontends[].inboundPort` | `int` | TCP port to listen on for incoming connections (e.g., `80`, `443`). Must be unique across frontends. | Required |
| `frontends[].backendName` | `string` | Name of the backend to route traffic to. Must match a backend's `name` in `spec.backends`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `privateNetworkId` | `StringValueOrRef` | none | Private Network to attach the LB to. When set, the LB receives a private IP and can reach backend servers on private addresses. Accepts a literal UUID or a `valueFrom` reference to a ScalewayPrivateNetwork's `status.outputs.private_network_id`. |
| `description` | `string` | `""` | Human-readable description displayed in the Scaleway console. |
| `sslCompatibilityLevel` | `string` | `"ssl_compatibility_level_intermediate"` | Minimum TLS version for HTTPS frontends. Options: `"ssl_compatibility_level_intermediate"` (TLS 1.2+), `"ssl_compatibility_level_modern"` (TLS 1.3 only). |
| `certificates` | `Certificate[]` | `[]` | TLS certificates for HTTPS frontends. Each certificate is either Let's Encrypt auto-provisioned or a custom PEM chain. Frontends reference certificates by name. |
| `certificates[].name` | `string` | — | Unique name for the certificate. Frontends reference this in their `certificateNames` field. Required when a certificate entry is present. |
| `certificates[].letsencrypt.commonName` | `string` | — | Primary domain for a Let's Encrypt certificate. Domain must resolve to the LB's public IP. Exactly one of `letsencrypt` or `customCertificate` must be set. |
| `certificates[].letsencrypt.subjectAlternativeNames` | `string[]` | `[]` | Additional domains covered by the Let's Encrypt certificate. All SANs must resolve to the LB's IP. |
| `certificates[].customCertificate.certificateChain` | `string` | — | Full PEM certificate chain (server cert + intermediates). Exactly one of `letsencrypt` or `customCertificate` must be set. |
| `backends[].forwardPortAlgorithm` | `string` | `"roundrobin"` | Load-balancing algorithm. Options: `"roundrobin"`, `"leastconn"`, `"first"`. |
| `backends[].stickySessions` | `string` | `"none"` | Sticky session type. Options: `"none"`, `"cookie"`, `"table"`. |
| `backends[].stickySessionsCookieName` | `string` | `""` | Cookie name for sticky sessions. Required when `stickySessions` is `"cookie"`. |
| `backends[].healthCheck.type` | `string` | `"tcp"` | Health check protocol. Options: `"tcp"`, `"http"`, `"https"`. |
| `backends[].healthCheck.uri` | `string` | `"/"` | URI path for HTTP/HTTPS health checks. Ignored for TCP checks. |
| `backends[].healthCheck.expectedCode` | `int` | `200` | Expected HTTP status code for a healthy response. Ignored for TCP checks. |
| `backends[].healthCheck.checkDelay` | `string` | `"5s"` | Interval between health check probes. |
| `backends[].healthCheck.checkTimeout` | `string` | `"3s"` | Maximum time to wait for a health check response. Must be less than `checkDelay`. |
| `backends[].healthCheck.checkMaxRetries` | `int` | `3` | Consecutive failed checks before marking a server as unhealthy. |
| `backends[].healthCheck.port` | `int` | same as `forwardPort` | Port to send health check probes to. Set when health checks use a dedicated monitoring port. |
| `backends[].timeoutConnect` | `string` | Scaleway default | Maximum time to wait for a connection to a backend server (e.g., `"5s"`). |
| `backends[].timeoutServer` | `string` | Scaleway default | Maximum idle time for a backend connection before it is closed (e.g., `"30s"`). |
| `backends[].onMarkedDownAction` | `string` | `"none"` | Action when a server is marked down. Options: `"none"`, `"shutdown_sessions"`. |
| `backends[].sslBridging` | `bool` | `false` | When `true`, the LB re-encrypts traffic to backend servers (TLS between LB and backends). |
| `backends[].proxyProtocol` | `string` | `"none"` | PROXY protocol version for passing client metadata to backends. Options: `"none"`, `"v1"`, `"v2"`, `"v2_ssl"`, `"v2_ssl_cn"`. |
| `frontends[].certificateNames` | `string[]` | `[]` | Names of TLS certificates to attach. Must match entries in `spec.certificates`. Required for HTTPS frontends. |
| `frontends[].timeoutClient` | `string` | Scaleway default | Maximum idle time for a client connection before it is closed (e.g., `"30s"`). |
| `frontends[].enableHttp3` | `bool` | `false` | When `true`, the frontend accepts HTTP/3 (QUIC) connections in addition to HTTP/1.1 and HTTP/2. Requires a TLS certificate. |

## Examples

### Minimal HTTP Load Balancer

A single-backend, single-frontend Load Balancer for development or internal services:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayLoadBalancer
metadata:
  name: dev-lb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayLoadBalancer.dev-lb
spec:
  zone: fr-par-1
  type: LB-S
  backends:
    - name: web
      serverIps:
        - "10.0.1.10"
      forwardPort: 8080
      forwardProtocol: http
      healthCheck:
        type: http
        uri: /health
        expectedCode: 200
  frontends:
    - name: http
      inboundPort: 80
      backendName: web
```

### HTTPS Load Balancer with Let's Encrypt and Private Network

A production-grade Load Balancer with automatic TLS, HTTP-to-HTTPS frontend pairing, a Private Network attachment, and HTTP health checks:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayLoadBalancer
metadata:
  name: prod-lb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ScalewayLoadBalancer.prod-lb
spec:
  zone: fr-par-1
  type: LB-GP-M
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: app-network
      fieldPath: status.outputs.private_network_id
  description: "Production web load balancer"
  sslCompatibilityLevel: ssl_compatibility_level_intermediate
  certificates:
    - name: app-cert
      letsencrypt:
        commonName: app.example.com
        subjectAlternativeNames:
          - www.example.com
  backends:
    - name: web
      serverIps:
        - "10.0.1.10"
        - "10.0.1.11"
        - "10.0.1.12"
      forwardPort: 8080
      forwardProtocol: http
      forwardPortAlgorithm: leastconn
      stickySessions: cookie
      stickySessionsCookieName: SERVERID
      healthCheck:
        type: http
        uri: /ready
        expectedCode: 200
        checkDelay: "10s"
        checkTimeout: "5s"
        checkMaxRetries: 3
      timeoutConnect: "5s"
      timeoutServer: "60s"
      onMarkedDownAction: shutdown_sessions
  frontends:
    - name: http
      inboundPort: 80
      backendName: web
    - name: https
      inboundPort: 443
      backendName: web
      certificateNames:
        - app-cert
      enableHttp3: true
```

### Multi-Backend TCP Load Balancer with PROXY Protocol

A Load Balancer routing traffic to separate backend pools (API and gRPC), using TCP forwarding and PROXY protocol to preserve client IPs:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayLoadBalancer
metadata:
  name: gateway-lb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ScalewayLoadBalancer.gateway-lb
spec:
  zone: nl-ams-1
  type: LB-GP-L
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: infra-network
      fieldPath: status.outputs.private_network_id
  backends:
    - name: api
      serverIps:
        - "10.0.2.10"
        - "10.0.2.11"
      forwardPort: 3000
      forwardProtocol: tcp
      forwardPortAlgorithm: leastconn
      proxyProtocol: v2
      healthCheck:
        type: tcp
        checkDelay: "5s"
        checkTimeout: "3s"
        checkMaxRetries: 3
      timeoutConnect: "5s"
      timeoutServer: "300s"
    - name: grpc
      serverIps:
        - "10.0.2.20"
        - "10.0.2.21"
      forwardPort: 50051
      forwardProtocol: tcp
      proxyProtocol: v2
      sslBridging: true
      healthCheck:
        type: tcp
        checkDelay: "10s"
        checkTimeout: "5s"
        checkMaxRetries: 5
  frontends:
    - name: api
      inboundPort: 443
      backendName: api
      timeoutClient: "60s"
    - name: grpc
      inboundPort: 50051
      backendName: grpc
      timeoutClient: "300s"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `lb_id` | `string` | Zoned ID of the created Load Balancer (e.g., `"fr-par-1/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"`). Can be used to reference the LB in Scaleway APIs. |
| `lb_ip_address` | `string` | Public IPv4 address assigned to the Load Balancer's Flexible IP. This is the address clients connect to. Use for DNS A records (via ScalewayDnsRecord), firewall allowlists, and monitoring. |
| `lb_ip_id` | `string` | Zoned ID of the Flexible IP resource. The IP has independent lifecycle and survives LB replacement. |

## Related Components

- [ScalewayPrivateNetwork](/docs/catalog/scaleway/private-network) — the Private Network that the Load Balancer attaches to for private backend connectivity
- [ScalewayVpc](/docs/catalog/scaleway/vpc) — the parent VPC containing the Private Network
- [ScalewayInstance](/docs/catalog/scaleway/instance) — compute instances that serve as backend servers behind the Load Balancer
- [ScalewayDnsRecord](/docs/catalog/scaleway/dns-record) — DNS records that can reference `status.outputs.lb_ip_address` to point a domain to the Load Balancer
- [ScalewayInstanceSecurityGroup](/docs/catalog/scaleway/instance-security-group) — controls network access for backend instances
