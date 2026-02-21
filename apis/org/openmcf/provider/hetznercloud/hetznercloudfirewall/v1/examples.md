# HetznerCloudFirewall Examples

## Minimal SSH-Only

The simplest useful firewall: allow SSH from anywhere, block everything else inbound. Outbound traffic is unrestricted by default.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudFirewall
metadata:
  name: ssh-only
spec:
  rules:
    - direction: in
      protocol: tcp
      port: "22"
      sourceIps:
        - "0.0.0.0/0"
        - "::/0"
      description: "allow SSH from anywhere"
```

---

## Web Server

A typical firewall for a public-facing web server: SSH for administration, HTTP and HTTPS for application traffic, and ICMP for diagnostics (ping, PMTU discovery).

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudFirewall
metadata:
  name: web-server
  org: my-org
  env: production
spec:
  rules:
    - direction: in
      protocol: tcp
      port: "22"
      sourceIps:
        - "0.0.0.0/0"
        - "::/0"
      description: "allow SSH"
    - direction: in
      protocol: tcp
      port: "80"
      sourceIps:
        - "0.0.0.0/0"
        - "::/0"
      description: "allow HTTP"
    - direction: in
      protocol: tcp
      port: "443"
      sourceIps:
        - "0.0.0.0/0"
        - "::/0"
      description: "allow HTTPS"
    - direction: in
      protocol: icmp
      sourceIps:
        - "0.0.0.0/0"
        - "::/0"
      description: "allow ping"
```

---

## Database with Restricted Access

A firewall for a database server that should only accept connections from a private subnet. No public SSH — administration is done via a bastion host on the same network.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudFirewall
metadata:
  name: db-restricted
  org: acme-corp
  env: production
  labels:
    role: database
    tier: data
spec:
  rules:
    - direction: in
      protocol: tcp
      port: "5432"
      sourceIps:
        - "10.0.1.0/24"
      description: "PostgreSQL from app subnet"
    - direction: in
      protocol: tcp
      port: "22"
      sourceIps:
        - "10.0.0.0/24"
      description: "SSH from bastion subnet only"
    - direction: in
      protocol: icmp
      sourceIps:
        - "10.0.0.0/16"
      description: "ping from private network"
```

---

## Full Production with Outbound Restrictions

A hardened firewall for a high-security environment. Inbound is limited to specific services; outbound is restricted to prevent a compromised server from reaching arbitrary destinations.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudFirewall
metadata:
  name: hardened-app
  org: acme-corp
  env: production
  labels:
    compliance: soc2
    tier: critical
spec:
  rules:
    - direction: in
      protocol: tcp
      port: "443"
      sourceIps:
        - "0.0.0.0/0"
        - "::/0"
      description: "HTTPS from anywhere"
    - direction: in
      protocol: tcp
      port: "22"
      sourceIps:
        - "203.0.113.0/24"
      description: "SSH from office IP range only"
    - direction: in
      protocol: icmp
      sourceIps:
        - "0.0.0.0/0"
        - "::/0"
      description: "ICMP for diagnostics"
    - direction: out
      protocol: tcp
      port: "443"
      destinationIps:
        - "0.0.0.0/0"
        - "::/0"
      description: "outbound HTTPS (package updates, API calls)"
    - direction: out
      protocol: tcp
      port: "53"
      destinationIps:
        - "0.0.0.0/0"
        - "::/0"
      description: "outbound DNS over TCP"
    - direction: out
      protocol: udp
      port: "53"
      destinationIps:
        - "0.0.0.0/0"
        - "::/0"
      description: "outbound DNS over UDP"
    - direction: out
      protocol: udp
      port: "123"
      destinationIps:
        - "0.0.0.0/0"
        - "::/0"
      description: "outbound NTP"
```

---

## InfraChart Composition with valueFrom

In an infra chart, a `HetznerCloudServer` references the firewall via `valueFrom` so the firewall ID is resolved from stack outputs. This eliminates hardcoded IDs and establishes a dependency edge in the DAG.

Firewall manifest:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudFirewall
metadata:
  name: web-firewall
  org: acme-corp
  env: production
spec:
  rules:
    - direction: in
      protocol: tcp
      port: "22"
      sourceIps:
        - "0.0.0.0/0"
        - "::/0"
      description: "allow SSH"
    - direction: in
      protocol: tcp
      port: "80"
      sourceIps:
        - "0.0.0.0/0"
        - "::/0"
      description: "allow HTTP"
    - direction: in
      protocol: tcp
      port: "443"
      sourceIps:
        - "0.0.0.0/0"
        - "::/0"
      description: "allow HTTPS"
```

Server manifest referencing the firewall output:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: web-01
  org: acme-corp
  env: production
spec:
  serverType: cx22
  image: ubuntu-24.04
  location: fsn1
  firewallIds:
    - valueFrom:
        kind: HetznerCloudFirewall
        name: web-firewall
        fieldPath: status.outputs.firewall_id
```

The `valueFrom` reference ensures that:
1. The firewall is created before the server
2. The correct numeric ID is passed to the server without manual lookup
3. Changes to the firewall (adding/removing rules) do not require updating the server manifest — only the firewall_id reference matters
