# OCI Network Load Balancer Examples

This document provides practical examples for deploying Oracle Cloud Infrastructure Network Load Balancers using the OpenMCF API. Each example demonstrates different use cases from basic TCP distribution to advanced firewall appliance topologies.

## Table of Contents

- [Example 1: Basic TCP Load Balancer](#example-1-basic-tcp-load-balancer)
- [Example 2: Private Internal NLB](#example-2-private-internal-nlb)
- [Example 3: Firewall Appliance with Transparent Mode](#example-3-firewall-appliance-with-transparent-mode)
- [Example 4: Multi-Protocol with DNS Health Checks](#example-4-multi-protocol-with-dns-health-checks)
- [Example 5: Production NLB with Advanced Failover](#example-5-production-nlb-with-advanced-failover)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Example 1: Basic TCP Load Balancer

**Use Case:** Public-facing TCP load balancer for web servers. Five-tuple hashing distributes traffic evenly, and HTTP health checks verify backend application health.

**Configuration:**
- **Visibility:** Public
- **Protocol:** TCP on port 80
- **Backends:** Two web servers on port 8080
- **Health Check:** HTTP on /health

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkLoadBalancer
metadata:
  name: web-nlb
  org: my-org
  env: dev
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

**Deploy with OpenMCF CLI:**

```bash
openmcf apply -f web-nlb.yaml
```

**What happens:**
- A public NLB is created in the specified subnet with an auto-assigned public IP.
- One backend set with two backends is created using five-tuple hashing.
- HTTP health checks probe each backend's /health endpoint every 10 seconds (default).
- A TCP listener on port 80 routes all incoming traffic to the web-backends set.

---

## Example 2: Private Internal NLB

**Use Case:** Internal NLB for gRPC services behind a VCN. Only accessible from within the VCN or peered networks. Uses references to OpenMCF-managed infrastructure.

**Configuration:**
- **Visibility:** Private
- **Protocol:** TCP on port 9090
- **Health Check:** HTTP on /healthz
- **Failover:** Instant failover with TCP RST

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkLoadBalancer
metadata:
  name: grpc-nlb
  org: acme
  env: prod
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  subnetId:
    valueFrom:
      kind: OciSubnet
      name: app-subnet
      fieldPath: status.outputs.subnetId
  isPrivate: true
  displayName: "gRPC Internal NLB"
  networkSecurityGroupIds:
    - valueFrom:
        kind: OciSecurityGroup
        name: app-nsg
        fieldPath: status.outputs.networkSecurityGroupId
  backendSets:
    - name: "grpc-backends"
      policy: five_tuple
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

**Deploy:**

```bash
openmcf apply -f grpc-nlb.yaml
```

**What happens:**
- A private NLB is created with no public IP — accessible only within the VCN.
- The compartment, subnet, and NSG are all resolved from OpenMCF-managed resources via `valueFrom`.
- Two primary backends with weight 3 each share traffic; a backup backend activates only when both primaries are unhealthy.
- Instant failover immediately migrates existing connections to a healthy backend, sending TCP RST to clients for fast reconnection.
- The listener's TCP idle timeout is set to 3600 seconds (1 hour) to accommodate long-lived gRPC streams.

---

## Example 3: Firewall Appliance with Transparent Mode

**Use Case:** NLB as a bump-in-the-wire traffic inspection point for firewall VMs. Source and destination IP preservation ensures firewalls see the true packet headers. Symmetric hashing guarantees return traffic follows the same path through the same firewall.

**Configuration:**
- **Mode:** Transparent (source/destination preservation + symmetric hash)
- **Protocol:** ANY (all IP traffic)
- **Backends:** Two firewall VMs referenced by instance OCID

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkLoadBalancer
metadata:
  name: firewall-nlb
  org: acme
  env: prod
  labels:
    team: security
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1.iad.example"
  isPreserveSourceDestination: true
  isSymmetricHashEnabled: true
  displayName: "Firewall Inspection NLB"
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

**What happens:**
- The NLB operates in transparent mode — packets arrive at firewall VMs with the original source and destination IPs unchanged.
- Symmetric hashing ensures that forward and return traffic for the same flow are sent to the same firewall VM, eliminating the need for backends to perform SNAT.
- Fail-open is enabled so traffic continues flowing even if health checks fail, preventing a total network blackout.
- Backends reference compute instances by OCID (`targetId`), and OCI resolves IP addresses automatically.

---

## Example 4: Multi-Protocol with DNS Health Checks

**Use Case:** Internal DNS server load balancing. Both TCP and UDP listeners serve DNS queries on port 53, and DNS-based health checks verify that backends respond to actual DNS queries.

**Configuration:**
- **Visibility:** Private
- **Protocols:** UDP and TCP on port 53
- **Health Check:** DNS query for a sentinel domain

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkLoadBalancer
metadata:
  name: dns-nlb
  org: acme
  env: prod
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1.iad.example"
  isPrivate: true
  displayName: "Internal DNS NLB"
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

**What happens:**
- A private NLB with two listeners (UDP and TCP) on port 53, both routing to the same backend set.
- Three-tuple hashing (src IP, dst IP, protocol) provides session stickiness without binding to source ports.
- DNS health checks send an actual A-record query for `health.internal.example.com` and expect `NOERROR` responses. This validates that the DNS server is fully functional, not just listening on the port.

---

## Example 5: Production NLB with Advanced Failover

**Use Case:** Production TCP service with reserved IPs, weighted backends, active-standby topology, Proxy Protocol v2, and fine-tuned health checking.

**Configuration:**
- **Reserved IP:** Pre-allocated static public IP
- **Proxy Protocol:** v2 enabled for client metadata
- **Backends:** Weighted primary + standby with operationally-active preference

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkLoadBalancer
metadata:
  name: prod-nlb
  org: acme
  env: prod
  labels:
    team: platform
    cost-center: infrastructure
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1.iad.example"
  displayName: "Production API NLB"
  reservedIps:
    - id: "ocid1.publicip.oc1.iad.reserved1"
  networkSecurityGroupIds:
    - value: "ocid1.networksecuritygroup.oc1.iad.example"
  backendSets:
    - name: "api-backends"
      policy: five_tuple
      isFailOpen: true
      isInstantFailoverEnabled: true
      isInstantFailoverTcpResetEnabled: true
      areOperationallyActiveBackendsPreferred: true
      healthChecker:
        protocol: https
        port: 8443
        urlPath: "/ready"
        returnCode: 200
        responseBodyRegex: ".*OK.*"
        intervalInMillis: 3000
        timeoutInMillis: 2000
        retries: 2
      backends:
        - port: 8443
          ipAddress: "10.0.1.10"
          weight: 5
        - port: 8443
          ipAddress: "10.0.1.11"
          weight: 5
        - port: 8443
          ipAddress: "10.0.1.12"
          weight: 3
        - port: 8443
          ipAddress: "10.0.1.20"
          isBackup: true
  listeners:
    - name: "api-listener"
      port: 443
      protocol: tcp
      defaultBackendSetName: "api-backends"
      isPpv2Enabled: true
      tcpIdleTimeout: 900
```

**What happens:**
- The NLB uses a pre-allocated reserved public IP for stable DNS and firewall configurations.
- Three primary backends receive weighted traffic (5:5:3 ratio), and one backup backend activates only during total primary failure.
- HTTPS health checks probe `/ready` on port 8443, validating both the HTTP response code and a response body pattern.
- Proxy Protocol v2 prepends client metadata (source IP, port, protocol) to every TCP connection, allowing backends to extract the original client identity.
- Instant failover with TCP RST ensures sub-second recovery for existing connections when a backend fails.

---

## Common Operations

### Get NLB IP Addresses After Deployment

```bash
# Pulumi
pulumi stack output ip_addresses

# Terraform
terraform output ip_addresses
```

### Use NLB IP in DNS Records

The `ip_addresses` output is a comma-separated list. Use it to create DNS records pointing to the NLB:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDnsRecord
metadata:
  name: api-dns
spec:
  zoneId:
    value: "ocid1.dns-zone.oc1..example"
  domain: "api.example.com"
  rtype: "A"
  items:
    - rdata: "129.146.10.5"
```

### Drain a Backend

To gracefully remove a backend from rotation, set `isDrain: true` in the backend's spec. Existing connections complete, but no new connections are sent to the drained backend.

---

## Best Practices

### Choose the Right Load Balancing Policy

| Policy | Hash Fields | Best For |
|--------|------------|----------|
| **five_tuple** | src IP, src port, dst IP, dst port, protocol | General purpose. Provides the best distribution across backends. |
| **three_tuple** | src IP, dst IP, protocol | Connection affinity by client IP regardless of source port. Good for protocols that use multiple connections. |
| **two_tuple** | src IP, dst IP | Strongest affinity. All traffic from a client goes to the same backend regardless of protocol or port. |

### NLB vs Application Load Balancer

| Requirement | Use NLB | Use L7 LB |
|-------------|---------|-----------|
| Raw TCP/UDP distribution | Yes | — |
| Source IP preservation | Yes | — |
| SSL termination | — | Yes |
| HTTP path/header routing | — | Yes |
| Firewall/security appliances | Yes | — |
| DNS server load balancing | Yes | — |
| Elastic bandwidth (no shape) | Yes | — |
| gRPC with URL routing | — | Yes |

### Health Check Protocol Selection

- **HTTP/HTTPS** — best for application-level health. Validates that the backend application is responding, not just that the port is open.
- **TCP** — fastest and simplest. Validates port reachability. Use when backend health is managed by other means.
- **UDP** — for UDP-only backends. Sends a payload and checks the response.
- **DNS** — for DNS servers. Sends an actual DNS query and validates the response code. Far more reliable than TCP port checks for DNS.

### Production Recommendations

- Enable **instant failover** for latency-sensitive services. TCP RST signaling gives clients immediate reconnection signals.
- Enable **fail-open** for services where partial availability is better than total outage.
- Use **reserved IPs** for production NLBs. Ephemeral IPs change on recreation, breaking DNS records and firewall rules.
- Set **health check intervals** to 3000-5000ms for production, with 2-3 retries. Faster detection means faster failover.
- Use **backup backends** for active-standby topologies rather than relying solely on fail-open.
