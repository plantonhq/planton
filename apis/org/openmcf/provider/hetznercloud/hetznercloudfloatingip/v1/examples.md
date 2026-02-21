# HetznerCloudFloatingIp Examples

## Minimal IPv4 Floating IP

The simplest configuration: allocate a single IPv4 address in Falkenstein. No server assignment, no reverse DNS, no delete protection. Suitable for reserving an IP address before a server is provisioned.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFloatingIp
metadata:
  name: dev-ip
spec:
  type: ipv4
  homeLocation: fsn1
```

---

## IPv4 with Description and Reverse DNS

An IPv4 address with a human-readable description and a reverse DNS pointer — the standard setup for a mail server. The `dnsPtr` value should match the forward DNS A record for the mail server's hostname. Without matching forward and reverse DNS, outbound email will be rejected by most receiving servers.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFloatingIp
metadata:
  name: mail-ip
  org: acme-corp
  env: production
spec:
  type: ipv4
  homeLocation: fsn1
  description: Production mail server failover IP
  dnsPtr: mail.example.com
  deleteProtection: true
```

---

## IPv6 with Delete Protection

An IPv6 /64 network block in Helsinki with delete protection enabled. The allocated block provides approximately 18 quintillion addresses. Useful for services that need multiple IPv6 addresses or when IPv4 cost is a concern.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFloatingIp
metadata:
  name: web-ipv6
  org: acme-corp
  env: production
  labels:
    purpose: web-frontend
spec:
  type: ipv6
  homeLocation: hel1
  description: IPv6 block for web frontends
  deleteProtection: true
```

---

## Assigned to a Server (Literal ID)

A Floating IP assigned to a specific server at creation time using a literal server ID. The server must be in the same location (`fsn1`) as the Floating IP's `homeLocation`. After assignment, the server's operating system must be configured with an IP alias to accept traffic on the Floating IP address.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFloatingIp
metadata:
  name: app-failover
  org: acme-corp
  env: production
spec:
  type: ipv4
  homeLocation: fsn1
  description: Application failover IP
  serverId:
    value: "12345678"
  dnsPtr: app.example.com
  deleteProtection: true
```

---

## InfraChart Composition with valueFrom

In an infra chart, the Floating IP references a `HetznerCloudServer` via `valueFrom` so the server's numeric ID is resolved from stack outputs. This eliminates hardcoded IDs and establishes a dependency edge in the DAG — the Floating IP waits for the server to be created before attempting assignment.

Server manifest:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: web-01
  org: acme-corp
  env: production
spec:
  serverType: cx22
  image: ubuntu-24.04
  location: fsn1
```

Floating IP manifest referencing the server output:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFloatingIp
metadata:
  name: web-failover
  org: acme-corp
  env: production
spec:
  type: ipv4
  homeLocation: fsn1
  description: Web frontend failover IP
  serverId:
    valueFrom:
      kind: HetznerCloudServer
      name: web-01
      fieldPath: status.outputs.server_id
  dnsPtr: web.example.com
  deleteProtection: true
```

The `valueFrom` reference ensures that:
1. The server is created before the Floating IP attempts assignment
2. The correct numeric ID is passed without manual lookup
3. Replacing the server automatically updates the Floating IP's assignment on the next apply
