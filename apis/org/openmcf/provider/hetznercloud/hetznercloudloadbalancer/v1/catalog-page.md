# Hetzner Cloud Load Balancer

Deploys a Hetzner Cloud load balancer with configurable services (HTTP, HTTPS, TCP listeners), backend targets (servers, label selectors, external IPs), health checks, TLS termination, and optional private network attachment. This is the most feature-rich component in the Hetzner Cloud catalog — it bundles four provider resources into a single manifest that defines how traffic reaches your servers.

## What Gets Created

- **Load Balancer** — an `hcloud_load_balancer` resource provisioning a load balancer with the specified type, location, algorithm, labels, and delete protection.
- **Load Balancer Service** (one per entry in `services`) — an `hcloud_load_balancer_service` resource configuring a listener with protocol, ports, HTTP settings (sticky sessions, TLS certificates, HTTP-to-HTTPS redirect), and health check.
- **Load Balancer Target** (one per entry across `serverTargets`, `labelSelectorTargets`, and `ipTargets`) — an `hcloud_load_balancer_target` resource adding a backend. Server targets reference a Hetzner Cloud server by ID, label selector targets dynamically match servers by labels, and IP targets route to external addresses.
- **Load Balancer Network** (created only when `network` is set) — an `hcloud_load_balancer_network` resource attaching the load balancer to a private network with an optional fixed IP and public interface control.

## Prerequisites

- **Hetzner Cloud API token** configured via environment variable (`HCLOUD_TOKEN`) or OpenMCF provider config
- **At least one backend** — server targets require pre-existing servers (or `HetznerCloudServer` resources), label selector targets require servers with matching labels, IP targets require reachable external IPs
- **A private network** if using the `network` attachment — either pre-existing or a `HetznerCloudNetwork` resource with a subnet in the same network zone as the load balancer's location
- **TLS certificates** if configuring HTTPS services — either pre-existing certificate IDs or `HetznerCloudCertificate` resources referenced via `valueFrom`

## Quick Start

Create a file `load-balancer.yaml`:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudLoadBalancer
metadata:
  name: my-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudLoadBalancer.my-lb
spec:
  loadBalancerType: lb11
  location: fsn1
  services:
    - protocol: http
  serverTargets:
    - serverId:
        value: "12345"
```

Deploy:

```shell
openmcf apply -f load-balancer.yaml
```

This provisions an lb11 load balancer in Falkenstein with a single HTTP listener on port 80 forwarding to one server target.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `loadBalancerType` | `string` | Load balancer size. Available types: `lb11` (25 targets, 10k conn/s), `lb21` (75 targets, 20k conn/s), `lb31` (150 targets, 40k conn/s). Can be changed after creation (in-place resize). | `min_len: 1` |
| `location` | `string` | Hetzner Cloud location for the load balancer. Known locations: `fsn1`, `nbg1`, `hel1`, `ash`, `hil`, `sin`. Changing this forces replacement. | `min_len: 1` |
| `services` | `Service[]` | Listeners that the load balancer exposes. Each service binds to a listen port and forwards traffic to targets. | `min_items: 1` |
| `services[].protocol` | `enum` | Listener protocol: `http`, `https`, or `tcp`. Changing forces service replacement. | `required`, `defined_only` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `algorithm` | `enum` | `round_robin` | Traffic distribution algorithm. Values: `round_robin`, `least_connections`. |
| `deleteProtection` | `bool` | `false` | Prevent accidental deletion via the Hetzner Cloud API. |
| `services[].listenPort` | `int` | `80` (http), `443` (https) | Port the load balancer listens on. Must be unique across services. Required for `tcp`. Changing forces replacement. |
| `services[].destinationPort` | `int` | same as `listenPort` | Port on target servers that receives forwarded traffic. Required for `tcp`. |
| `services[].proxyprotocol` | `bool` | `false` | Enable PROXY protocol v1 when forwarding to targets. Target must support PROXY protocol. |
| `services[].http` | `object` | unset | HTTP-level configuration. Only applicable for `http` and `https` protocols. |
| `services[].http.stickySessions` | `bool` | `false` | Enable cookie-based session affinity. |
| `services[].http.cookieName` | `string` | `HCLBSTICKY` | Sticky session cookie name. Only used when `stickySessions` is `true`. |
| `services[].http.cookieLifetime` | `int` | `300` | Sticky session cookie lifetime in seconds. Only used when `stickySessions` is `true`. |
| `services[].http.certificateIds` | `StringValueOrRef[]` | empty | TLS certificates for HTTPS termination. Can reference `HetznerCloudCertificate` resources via `valueFrom`. Only used when protocol is `https`. |
| `services[].http.redirectHttp` | `bool` | `false` | Redirect HTTP traffic on port 80 to HTTPS. Only valid for `https` services on port 443. |
| `services[].healthCheck` | `object` | provider default | Custom health check. When omitted, the provider creates a default matching the service protocol and destination port. |
| `services[].healthCheck.protocol` | `enum` | matches service | Health check protocol: `http`, `https`, or `tcp`. HTTPS services default to `http` health checks. |
| `services[].healthCheck.port` | `int` | matches `destinationPort` | Port to health-check on the target. |
| `services[].healthCheck.interval` | `int` | `15` | Seconds between health checks. |
| `services[].healthCheck.timeout` | `int` | `10` | Seconds to wait for a response. Must be less than `interval`. |
| `services[].healthCheck.retries` | `int` | `3` | Consecutive failures before marking target unhealthy. |
| `services[].healthCheck.http` | `object` | unset | HTTP-specific health check settings. Only used when health check protocol is `http` or `https`. |
| `services[].healthCheck.http.domain` | `string` | target IP | HTTP Host header for the health check request. |
| `services[].healthCheck.http.path` | `string` | `/` | URL path for the health check request. |
| `services[].healthCheck.http.response` | `string` | empty | Expected response body substring. |
| `services[].healthCheck.http.tls` | `bool` | `false` | Verify target TLS certificate during health check. Only meaningful for `https` protocol. |
| `services[].healthCheck.http.statusCodes` | `string[]` | `["2??", "3??"]` | Expected HTTP status codes. Uses wildcard notation: `2??` matches any 2xx status. |
| `serverTargets` | `ServerTarget[]` | empty | Static server backends. |
| `serverTargets[].serverId` | `StringValueOrRef` | — | Server to add as a target. Can reference `HetznerCloudServer` resources via `valueFrom`. Required within each entry. Changing forces replacement. |
| `serverTargets[].usePrivateIp` | `bool` | `false` | Route traffic over private network instead of public IP. Requires network attachment. |
| `labelSelectorTargets` | `LabelSelectorTarget[]` | empty | Dynamic server backends selected by labels. |
| `labelSelectorTargets[].selector` | `string` | — | Hetzner Cloud label selector expression (e.g., `env=production,role=web`). Required within each entry. Changing forces replacement. |
| `labelSelectorTargets[].usePrivateIp` | `bool` | `false` | Route traffic over private network. Requires network attachment. |
| `ipTargets` | `IpTarget[]` | empty | External IP backends for targets outside Hetzner Cloud. |
| `ipTargets[].ip` | `string` | — | IP address of the external backend. Required within each entry. Changing forces replacement. |
| `network` | `object` | unset | Private network attachment. At most one network per load balancer. |
| `network.networkId` | `StringValueOrRef` | — | Network to attach to. Can reference `HetznerCloudNetwork` resources via `valueFrom`. Required when `network` is set. Changing forces replacement. |
| `network.ip` | `string` | auto-assigned | Fixed private IP within the network's subnet range. |
| `network.enablePublicInterface` | `bool` | `true` | Enable the load balancer's public interface. When `false`, the LB is only reachable via its private IP. |

## Examples

### Minimal HTTP Load Balancer

An lb11 load balancer with an HTTP listener on port 80 forwarding to a single server target.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudLoadBalancer
metadata:
  name: web-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudLoadBalancer.web-lb
spec:
  loadBalancerType: lb11
  location: fsn1
  services:
    - protocol: http
  serverTargets:
    - serverId:
        value: "12345"
```

### HTTPS with TLS Termination and Health Check

An HTTPS load balancer using a certificate from a `HetznerCloudCertificate` resource. HTTP traffic is redirected to HTTPS. A custom health check verifies the backend on `/health`.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudLoadBalancer
metadata:
  name: web-https-lb
  org: acme-corp
  env: staging
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: web-platform
    pulumi.openmcf.org/stack.name: staging.HetznerCloudLoadBalancer.web-https-lb
spec:
  loadBalancerType: lb11
  location: fsn1
  services:
    - protocol: https
      destinationPort: 8080
      http:
        certificateIds:
          - valueFrom:
              kind: HetznerCloudCertificate
              name: web-cert
              fieldPath: status.outputs.certificate_id
        redirectHttp: true
      healthCheck:
        protocol: http
        port: 8080
        interval: 10
        timeout: 5
        retries: 3
        http:
          path: /health
  serverTargets:
    - serverId:
        valueFrom:
          kind: HetznerCloudServer
          name: web-01
          fieldPath: status.outputs.server_id
    - serverId:
        valueFrom:
          kind: HetznerCloudServer
          name: web-02
          fieldPath: status.outputs.server_id
```

### Private Network Load Balancer with Label Selector Targets

A load balancer on a private network using label selectors to dynamically discover backend servers. All traffic flows over the private network.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudLoadBalancer
metadata:
  name: internal-lb
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudLoadBalancer.internal-lb
spec:
  loadBalancerType: lb21
  location: fsn1
  algorithm: least_connections
  services:
    - protocol: http
      destinationPort: 8080
      healthCheck:
        protocol: http
        port: 8080
        http:
          path: /ready
  labelSelectorTargets:
    - selector: "env=production,role=web"
      usePrivateIp: true
  network:
    networkId:
      valueFrom:
        kind: HetznerCloudNetwork
        name: main-vpc
        fieldPath: status.outputs.network_id
    ip: "10.0.1.100"
    enablePublicInterface: true
```

### Full-Featured Production Load Balancer

A production load balancer with HTTPS, sticky sessions, multiple target types, private networking, custom health checks, and delete protection.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudLoadBalancer
metadata:
  name: prod-lb
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudLoadBalancer.prod-lb
spec:
  loadBalancerType: lb31
  location: fsn1
  algorithm: least_connections
  deleteProtection: true
  services:
    - protocol: https
      destinationPort: 8080
      http:
        certificateIds:
          - valueFrom:
              kind: HetznerCloudCertificate
              name: prod-cert
              fieldPath: status.outputs.certificate_id
        redirectHttp: true
        stickySessions: true
        cookieName: PRODSESSION
        cookieLifetime: 1800
      healthCheck:
        protocol: http
        port: 8080
        interval: 10
        timeout: 5
        retries: 3
        http:
          path: /health
          statusCodes:
            - "200"
    - protocol: tcp
      listenPort: 6379
      destinationPort: 6379
      healthCheck:
        protocol: tcp
        port: 6379
        interval: 10
        timeout: 5
        retries: 2
  serverTargets:
    - serverId:
        valueFrom:
          kind: HetznerCloudServer
          name: web-01
          fieldPath: status.outputs.server_id
      usePrivateIp: true
    - serverId:
        valueFrom:
          kind: HetznerCloudServer
          name: web-02
          fieldPath: status.outputs.server_id
      usePrivateIp: true
  labelSelectorTargets:
    - selector: "env=production,role=web"
      usePrivateIp: true
  ipTargets:
    - ip: "203.0.113.50"
  network:
    networkId:
      valueFrom:
        kind: HetznerCloudNetwork
        name: main-vpc
        fieldPath: status.outputs.network_id
    ip: "10.0.1.200"
    enablePublicInterface: true
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `load_balancer_id` | `string` | Hetzner Cloud numeric ID of the created load balancer. Can be referenced by other components via `StringValueOrRef`. |
| `ipv4_address` | `string` | The public IPv4 address assigned to the load balancer. Empty if the public interface is disabled via `network.enablePublicInterface = false`. |
| `ipv6_address` | `string` | The public IPv6 address assigned to the load balancer. Empty if the public interface is disabled. |

## Related Components

- [HetznerCloudServer](/docs/catalog/hetznercloud/hetznercloudserver) — Servers added as load balancer backends, referenced via `serverTargets[].serverId`
- [HetznerCloudNetwork](/docs/catalog/hetznercloud/hetznercloudnetwork) — Private network for internal traffic routing, referenced via `network.networkId`
- [HetznerCloudCertificate](/docs/catalog/hetznercloud/hetznercloudcertificate) — TLS certificates for HTTPS termination, referenced via `services[].http.certificateIds`
- [HetznerCloudFirewall](/docs/catalog/hetznercloud/hetznercloudfirewall) — Network security rules applied to backend servers
- [HetznerCloudSshKey](/docs/catalog/hetznercloud/hetznercloudsshkey) — SSH keys for accessing backend servers
