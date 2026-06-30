---
title: "Network Load Balancer"
description: "Network Load Balancer deployment documentation"
icon: "package"
order: 100
componentName: "ocinetworkloadbalancer"
---

# OCI Network Load Balancer

Deploys an Oracle Cloud Infrastructure Network Load Balancer (Layer 4) with backend sets, backends, and listeners in a single deployment unit. The NLB provides high-performance, low-latency traffic distribution for TCP, UDP, and mixed-protocol workloads with fully elastic bandwidth and source IP preservation.

## What Gets Created

When you deploy an OciNetworkLoadBalancer resource, Planton provisions:

- **Network Load Balancer** — an `oci_network_load_balancer_network_load_balancer` resource in the specified compartment and subnet. Supports public or private deployment, optional reserved IP addresses, source IP preservation, and symmetric hashing. Standard Planton freeform tags are applied for resource tracking.
- **Backend Sets** — one `oci_network_load_balancer_backend_set` per entry in `backendSets`. Each set defines a load balancing policy (tuple-based hashing), health checker configuration, and failover behavior. Backend sets are created as children of the NLB.
- **Backends** — one `oci_network_load_balancer_backend` per entry in each backend set's `backends` list. Backends are identified by IP address, compute instance OCID, or both. Created as children of their backend set.
- **Listeners** — one `oci_network_load_balancer_listener` per entry in `listeners`. Each listener binds a port and Layer 4 protocol to a default backend set. Listeners are created after all backend sets, ensuring referential integrity.

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the NLB will be created — literal value or reference to an OciCompartment resource
- **A subnet OCID** for the NLB — literal value or reference to an OciSubnet resource. The subnet determines whether the NLB gets public or private IP addresses.
- **Backend server IP addresses or compute instance OCIDs** for traffic targets

## Quick Start

Create a file `nlb.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciNetworkLoadBalancer
metadata:
  name: my-nlb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciNetworkLoadBalancer.my-nlb
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1.iad.example"
  backendSets:
    - name: "tcp-backends"
      policy: five_tuple
      healthChecker:
        protocol: tcp
      backends:
        - port: 8080
          ipAddress: "10.0.1.10"
        - port: 8080
          ipAddress: "10.0.1.11"
  listeners:
    - name: "tcp-listener"
      port: 80
      protocol: tcp
      defaultBackendSetName: "tcp-backends"
```

Deploy:

```shell
planton apply -f nlb.yaml
```

This creates a public NLB listening on port 80 (TCP) and distributing traffic to two backends on port 8080 using five-tuple hashing. TCP health checks verify backend availability. The NLB ID and assigned IP addresses are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the NLB will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `subnetId` | `StringValueOrRef` | OCID of the subnet where the NLB will be deployed. The NLB is deployed into a single subnet (unlike the L7 load balancer). Changing this after creation forces recreation. Can reference an OciSubnet resource via `valueFrom`. | Required |
| `backendSets` | `BackendSet[]` | Backend sets defining groups of backend servers with load balancing policies and health checking. See [backendSet fields](#backendset-fields). | Minimum 1 item |
| `listeners` | `Listener[]` | Listeners defining ports and protocols on which the NLB accepts connections. See [listener fields](#listener-fields). | Minimum 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Human-readable name shown in the OCI Console. |
| `isPrivate` | `bool` | `false` | When `true`, creates a private NLB accessible only within the VCN. Private NLBs receive only private IP addresses. Changing this after creation forces recreation. |
| `isPreserveSourceDestination` | `bool` | `false` | When `true`, preserves the original source and destination IP addresses in packet headers. Automatically enables skip-source-dest-check on the NLB's VNIC. Essential for firewalls and security appliances. |
| `isSymmetricHashEnabled` | `bool` | `false` | When `true`, enables symmetric hashing. Only valid when `isPreserveSourceDestination` is also `true`. Removes the need for backends to perform SNAT. |
| `networkSecurityGroupIds` | `StringValueOrRef[]` | — | OCIDs of network security groups applied to the NLB. Can reference OciSecurityGroup resources via `valueFrom`. |
| `nlbIpVersion` | `string` | `"IPV4"` | IP version for the NLB. Values: `"IPV4"`, `"IPV6"`, `"IPV4_AND_IPV6"`. |
| `reservedIps` | `ReservedIp[]` | — | Pre-created reserved public IPs to assign. When omitted, OCI assigns ephemeral public IPs (for public NLBs). Each entry has an `id` field (OCID of the reserved IP). |
| `assignedIpv6` | `string` | — | IPv6 address to assign to the NLB. Must be part of one of the subnet's IPv6 prefixes. |
| `assignedPrivateIpv4` | `string` | — | Private IPv4 address within the subnet's CIDR range. When omitted, OCI assigns one automatically. Changing this after creation forces recreation. |
| `subnetIpv6cidr` | `string` | — | IPv6 subnet prefix selection. When provided, the NLB IPv6 address is assigned within this CIDR block. |

### backendSet Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Unique name for this backend set. Listeners reference backend sets by this name. | Minimum 1 character |
| `policy` | `enum` | Load balancing policy. Values: `five_tuple` (src IP, src port, dst IP, dst port, protocol), `three_tuple` (src IP, dst IP, protocol), `two_tuple` (src IP, dst IP). | Required (cannot be unspecified) |
| `healthChecker` | `HealthChecker` | Health checker configuration. See [healthChecker fields](#healthchecker-fields). | Required |
| `backends` | `Backend[]` | Backend servers in this set. When omitted, the set is created empty. See [backend fields](#backend-fields). | Optional |
| `isPreserveSource` | `bool` | When `true`, preserves the source IP when forwarding to backends in this set. | Optional |
| `isFailOpen` | `bool` | When `true`, continues distributing traffic to all backends even when all are marked unhealthy. | Optional |
| `isInstantFailoverEnabled` | `bool` | When `true`, immediately forwards existing connections to a healthy backend when the current one becomes unhealthy. | Optional |
| `isInstantFailoverTcpResetEnabled` | `bool` | When `true` (and instant failover is enabled), sends a TCP RST to clients instead of silently failing over. | Optional |
| `areOperationallyActiveBackendsPreferred` | `bool` | When `true`, preferentially routes traffic to operationally active backends (active-standby support). | Optional |
| `ipVersion` | `string` | IP version for this backend set. When omitted, inherits from the NLB's `nlbIpVersion`. | Optional |

### healthChecker Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `protocol` | `enum` | Health check protocol. Values: `http`, `https`, `tcp`, `udp`, `dns`. | Required (cannot be unspecified) |
| `port` | `int32` | Port to probe. When omitted, uses the backend's traffic port. | Optional |
| `urlPath` | `string` | URL path for HTTP/HTTPS health checks (e.g., `"/health"`). Required when protocol is `http` or `https`. | Optional |
| `returnCode` | `int32` | Expected HTTP status code (e.g., `200`). Used with `http` and `https` protocols. | Optional |
| `responseBodyRegex` | `string` | Regex pattern to match against the response body. | Optional |
| `intervalInMillis` | `int32` | Interval between health checks in milliseconds. Default: `10000`. | Optional |
| `timeoutInMillis` | `int32` | Maximum wait time for a response in milliseconds. Default: `3000`. | Optional |
| `retries` | `int32` | Consecutive failures before marking a backend unhealthy. Default: `3`. | Optional |
| `requestData` | `string` | Base64-encoded probe payload for TCP/UDP health checks. | Optional |
| `responseData` | `string` | Base64-encoded expected response for TCP/UDP health checks. | Optional |
| `dnsHealthCheck` | `DnsHealthCheck` | DNS health check configuration. Required when protocol is `dns`. See [dnsHealthCheck fields](#dnshealthcheck-fields). | Optional |

### dnsHealthCheck Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `domainName` | `string` | Fully qualified domain name to query. | Minimum 1 character |
| `queryClass` | `string` | DNS query class. Values: `"IN"`, `"CH"`. Default: `"IN"`. | Optional |
| `queryType` | `string` | DNS query type. Values: `"A"`, `"AAAA"`, `"TXT"`. Default: `"A"`. | Optional |
| `rcodes` | `string[]` | Acceptable DNS response codes (e.g., `["NOERROR"]`). Default: `["NOERROR"]`. | Optional |
| `transportProtocol` | `string` | Transport protocol for DNS queries. Values: `"UDP"`, `"TCP"`. Default: `"UDP"`. | Optional |

### backend Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `port` | `int32` | Port on which the backend server listens for traffic. | 1–65535 |
| `ipAddress` | `string` | IP address of the backend server. When omitted, `targetId` must be provided. | Optional |
| `targetId` | `string` | OCID of a compute instance or private IP. OCI resolves the IP address automatically. When omitted, `ipAddress` must be provided. | Optional |
| `weight` | `int32` | Relative weight for traffic distribution. Higher weights receive more traffic. Default: `1`. | Optional |
| `isBackup` | `bool` | When `true`, this backend only receives traffic when all non-backup backends are unhealthy. | Optional |
| `isDrain` | `bool` | When `true`, existing connections complete but no new connections are routed to this backend. | Optional |
| `isOffline` | `bool` | When `true`, no traffic is sent to this backend. | Optional |
| `name` | `string` | Unique name for the backend within the set. When omitted, OCI generates `"IP:port"`. | Optional |

### listener Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Unique name for this listener. | Minimum 1 character |
| `port` | `int32` | Port on which the listener accepts connections. | 1–65535 |
| `protocol` | `enum` | Layer 4 protocol. Values: `tcp`, `udp`, `tcp_and_udp`, `any`. | Required (cannot be unspecified) |
| `defaultBackendSetName` | `string` | Name of the backend set that receives traffic. Must match a name in `backendSets`. | Minimum 1 character |
| `ipVersion` | `string` | IP version for this listener. When omitted, inherits from the NLB's `nlbIpVersion`. | Optional |
| `isPpv2Enabled` | `bool` | When `true`, enables Proxy Protocol v2. PPv2 prepends connection metadata (source IP, port, protocol) to the TCP stream. | Optional |
| `tcpIdleTimeout` | `int32` | TCP idle timeout in seconds. Connections idle longer than this are closed. | Optional |
| `udpIdleTimeout` | `int32` | UDP idle timeout in seconds. | Optional |
| `l3ipIdleTimeout` | `int32` | L3IP idle timeout in seconds. Applies when protocol is `any`. | Optional |

## Examples

### Basic TCP Load Balancer

A public NLB distributing TCP traffic across two backends with five-tuple hashing and TCP health checks:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciNetworkLoadBalancer
metadata:
  name: web-nlb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciNetworkLoadBalancer.web-nlb
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1.iad.example"
  backendSets:
    - name: "web-backends"
      policy: five_tuple
      healthChecker:
        protocol: http
        port: 8080
        urlPath: "/health"
        returnCode: 200
      backends:
        - port: 8080
          ipAddress: "10.0.1.10"
        - port: 8080
          ipAddress: "10.0.1.11"
  listeners:
    - name: "http-listener"
      port: 80
      protocol: tcp
      defaultBackendSetName: "web-backends"
```

### Private NLB with Failover

A private NLB for internal services with instant failover and fail-open enabled. When a backend becomes unhealthy, existing connections immediately move to a healthy backend with a TCP RST signal:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciNetworkLoadBalancer
metadata:
  name: internal-nlb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: platform
    pulumi.planton.dev/stack.name: prod.OciNetworkLoadBalancer.internal-nlb
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  subnetId:
    valueFrom:
      kind: OciSubnet
      name: private-subnet
      fieldPath: status.outputs.subnetId
  isPrivate: true
  displayName: "Internal Service NLB"
  networkSecurityGroupIds:
    - valueFrom:
        kind: OciSecurityGroup
        name: nlb-nsg
        fieldPath: status.outputs.networkSecurityGroupId
  backendSets:
    - name: "grpc-backends"
      policy: five_tuple
      isFailOpen: true
      isInstantFailoverEnabled: true
      isInstantFailoverTcpResetEnabled: true
      healthChecker:
        protocol: http
        port: 8081
        urlPath: "/healthz"
        returnCode: 200
        intervalInMillis: 5000
        timeoutInMillis: 2000
        retries: 2
      backends:
        - port: 9090
          ipAddress: "10.0.2.10"
          weight: 3
        - port: 9090
          ipAddress: "10.0.2.11"
          weight: 3
        - port: 9090
          ipAddress: "10.0.2.12"
          isBackup: true
  listeners:
    - name: "grpc-listener"
      port: 9090
      protocol: tcp
      defaultBackendSetName: "grpc-backends"
      tcpIdleTimeout: 3600
```

### Firewall Appliance with Source IP Preservation

An NLB in transparent mode forwarding all traffic to firewall appliances while preserving the original source and destination IPs. Symmetric hashing ensures return traffic follows the same path:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciNetworkLoadBalancer
metadata:
  name: fw-nlb
  org: acme
  env: prod
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: security
    pulumi.planton.dev/stack.name: prod.OciNetworkLoadBalancer.fw-nlb
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1.iad.example"
  isPreserveSourceDestination: true
  isSymmetricHashEnabled: true
  displayName: "Firewall NLB"
  backendSets:
    - name: "firewall-backends"
      policy: five_tuple
      isPreserveSource: true
      isFailOpen: true
      healthChecker:
        protocol: tcp
        port: 443
        intervalInMillis: 5000
        timeoutInMillis: 3000
        retries: 3
      backends:
        - port: 0
          targetId: "ocid1.instance.oc1.iad.firewallvm1"
        - port: 0
          targetId: "ocid1.instance.oc1.iad.firewallvm2"
  listeners:
    - name: "all-traffic"
      port: 0
      protocol: any
      defaultBackendSetName: "firewall-backends"
```

### Multi-Protocol NLB with DNS Health Checks

An NLB serving both TCP and UDP traffic on separate listeners, using DNS-based health checks for a DNS server backend set:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciNetworkLoadBalancer
metadata:
  name: dns-nlb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: platform
    pulumi.planton.dev/stack.name: prod.OciNetworkLoadBalancer.dns-nlb
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1.iad.example"
  isPrivate: true
  backendSets:
    - name: "dns-backends"
      policy: three_tuple
      healthChecker:
        protocol: dns
        dnsHealthCheck:
          domainName: "health.internal.example.com"
          queryType: "A"
          rcodes:
            - "NOERROR"
          transportProtocol: "UDP"
      backends:
        - port: 53
          ipAddress: "10.0.3.10"
        - port: 53
          ipAddress: "10.0.3.11"
  listeners:
    - name: "dns-udp"
      port: 53
      protocol: udp
      defaultBackendSetName: "dns-backends"
      udpIdleTimeout: 120
    - name: "dns-tcp"
      port: 53
      protocol: tcp
      defaultBackendSetName: "dns-backends"
      tcpIdleTimeout: 300
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `network_load_balancer_id` | `string` | OCID of the created network load balancer. |
| `ip_addresses` | `string` | Comma-separated IP addresses assigned to the NLB (includes both public and private IPs as applicable). |

## Related Components

- [OciCompartment](/docs/catalog/oci/compartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciSubnet](/docs/catalog/oci/subnet) — provides the subnet referenced by `subnetId` via `valueFrom`
- [OciSecurityGroup](/docs/catalog/oci/network-security-group) — manages network security rules for the NLB via `networkSecurityGroupIds`
- [OciApplicationLoadBalancer](/docs/catalog/oci/application-load-balancer) — the Layer 7 counterpart for HTTP/HTTPS workloads with SSL termination, hostname routing, and rule sets
- [OciPublicIp](/docs/catalog/oci/public-ip) — provides reserved public IPs that can be assigned to the NLB via `reservedIps`
