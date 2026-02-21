# HetznerCloudPrimaryIp Examples

## Minimal IPv4 Primary IP

The simplest configuration: allocate a single IPv4 address in Falkenstein. No reverse DNS, no delete protection. Suitable for development servers or non-critical services.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudPrimaryIp
metadata:
  name: dev-ip
spec:
  type: ipv4
  location: fsn1
```

---

## IPv4 with Reverse DNS

An IPv4 address with a reverse DNS pointer — the standard setup for a mail server. The `dnsPtr` value should match the forward DNS A record for the mail server's hostname. Without matching forward and reverse DNS, outbound email will be rejected by most receiving servers.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudPrimaryIp
metadata:
  name: mail-ip
  org: acme-corp
  env: production
spec:
  type: ipv4
  location: fsn1
  dnsPtr: mail.example.com
  deleteProtection: true
```

---

## IPv6 with Delete Protection

An IPv6 /64 network block in Helsinki with delete protection enabled. The allocated block provides approximately 18 quintillion addresses. Useful for services that need multiple IPv6 addresses or when IPv4 cost is a concern.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudPrimaryIp
metadata:
  name: web-ipv6
  org: acme-corp
  env: production
  labels:
    purpose: web-frontend
spec:
  type: ipv6
  location: hel1
  deleteProtection: true
```

---

## InfraChart Composition with valueFrom

In an infra chart, a `HetznerCloudServer` references the Primary IP via `valueFrom` so the IP's numeric ID is resolved from stack outputs. This eliminates hardcoded IDs and establishes a dependency edge in the DAG — the server waits for the Primary IP to be allocated before attempting creation.

Primary IP manifest:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudPrimaryIp
metadata:
  name: app-ip
  org: acme-corp
  env: production
spec:
  type: ipv4
  location: fsn1
  dnsPtr: app.example.com
  deleteProtection: true
```

Server manifest referencing the Primary IP output:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: app-01
  org: acme-corp
  env: production
spec:
  serverType: cx22
  image: ubuntu-24.04
  location: fsn1
  primaryIpId:
    valueFrom:
      kind: HetznerCloudPrimaryIp
      name: app-ip
      fieldPath: status.outputs.primary_ip_id
```

The `valueFrom` reference ensures that:
1. The Primary IP is allocated before the server is created
2. The correct numeric ID is passed without manual lookup
3. Replacing the server does not affect the Primary IP — the IP persists and the new server can use the same reference
