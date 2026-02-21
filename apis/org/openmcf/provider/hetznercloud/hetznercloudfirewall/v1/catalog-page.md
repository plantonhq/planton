# Hetzner Cloud Firewall

Creates a firewall in Hetzner Cloud with inline rules that control inbound and outbound network traffic for servers. The firewall enforces deny-by-default inbound (all packets not matching a rule are dropped) and allow-by-default outbound. Rules support TCP, UDP, ICMP, ESP, and GRE protocols with IPv4 and IPv6 CIDR blocks.

## What Gets Created

- **Firewall** — an `hcloud_firewall` resource containing inline rules, a name derived from `metadata.name`, and standard labels computed from resource metadata. Rules define allowed traffic by direction, protocol, port (for TCP/UDP), and source or destination CIDR blocks.

## Prerequisites

- **Hetzner Cloud API token** configured via environment variable (`HCLOUD_TOKEN`) or OpenMCF provider config

## Quick Start

Create a file `firewall.yaml`:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFirewall
metadata:
  name: ssh-only
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudFirewall.ssh-only
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

Deploy:

```shell
openmcf apply -f firewall.yaml
```

This creates a firewall that allows inbound SSH from any IP and blocks all other inbound traffic when applied to a server.

## Configuration Reference

### Required Fields

The `spec` itself is required, but `rules` is optional — an empty rules list creates a valid firewall that blocks all inbound traffic.

When rules are provided, each rule requires:

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `rules[].direction` | `enum` (`in`, `out`) | Traffic direction this rule applies to | Required, must be a defined enum value |
| `rules[].protocol` | `enum` (`icmp`, `tcp`, `udp`, `esp`, `gre`) | IP protocol this rule matches | Required, must be a defined enum value |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `rules` | `repeated Rule` | empty (block all inbound) | Firewall rules defining allowed traffic. Up to 50 rules per firewall. |
| `rules[].port` | `string` | — | Port (`"80"`), range (`"80-443"`), or `"any"`. Required when protocol is `tcp` or `udp`. Must not be set for `icmp`, `esp`, or `gre`. |
| `rules[].sourceIps` | `repeated string` | — | CIDR blocks for inbound traffic sources. Required when direction is `in`. Use `["0.0.0.0/0", "::/0"]` for all traffic. |
| `rules[].destinationIps` | `repeated string` | — | CIDR blocks for outbound traffic destinations. Required when direction is `out`. Use `["0.0.0.0/0", "::/0"]` for all traffic. |
| `rules[].description` | `string` | — | Human-readable description of the rule's purpose. |

### Cross-Field Validation (CEL Rules)

| Constraint | Message |
|-----------|---------|
| `port` required for TCP/UDP | "port is required when protocol is tcp or udp" |
| `sourceIps` required for inbound | "source_ips is required when direction is in" |
| `destinationIps` required for outbound | "destination_ips is required when direction is out" |

## Examples

### Minimal SSH-Only Firewall

A single inbound rule allowing SSH from anywhere.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFirewall
metadata:
  name: ssh-only
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudFirewall.ssh-only
spec:
  rules:
    - direction: in
      protocol: tcp
      port: "22"
      sourceIps:
        - "0.0.0.0/0"
        - "::/0"
```

### Web Server Firewall

SSH, HTTP, HTTPS, and ICMP for a public-facing web server.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFirewall
metadata:
  name: web-server
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudFirewall.web-server
    role: web
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

### Database with Restricted Access

PostgreSQL access restricted to a private subnet, SSH from bastion only.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFirewall
metadata:
  name: db-restricted
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudFirewall.db-restricted
    role: database
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
      description: "SSH from bastion subnet"
```

### Server Composition via valueFrom

A firewall referenced by a HetznerCloudServer using `valueFrom`. The server receives the firewall's numeric ID from the firewall's stack outputs, establishing a dependency edge in the deployment DAG.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFirewall
metadata:
  name: web-firewall
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudFirewall.web-firewall
spec:
  rules:
    - direction: in
      protocol: tcp
      port: "22"
      sourceIps:
        - "0.0.0.0/0"
        - "::/0"
    - direction: in
      protocol: tcp
      port: "443"
      sourceIps:
        - "0.0.0.0/0"
        - "::/0"
```

The server references this firewall:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: web-01
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudServer.web-01
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

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `firewall_id` | `string` | Hetzner Cloud numeric ID of the created firewall. Referenced by HetznerCloudServer via `firewallIds`. |

## Related Components

- [HetznerCloudServer](/docs/catalog/hetznercloud/hetznercloudserver) — References firewall IDs to apply traffic rules at server creation
- [HetznerCloudSshKey](/docs/catalog/hetznercloud/hetznercloudsshkey) — Commonly deployed alongside firewalls as a foundation for server provisioning
- [HetznerCloudNetwork](/docs/catalog/hetznercloud/hetznercloudnetwork) — Private networking for restricting firewall source CIDRs to internal subnets
