# DigitalOcean Firewall

Deploys a stateful, network-edge Cloud Firewall on DigitalOcean that enforces a default-deny security model for Droplets. The component supports inbound and outbound rules with IP-based, tag-based, Load Balancer, and Kubernetes cluster source/destination targeting, and can be applied to Droplets by ID or tag.

## What Gets Created

When you deploy a DigitalOceanFirewall resource, Planton provisions:

- **DigitalOcean Firewall** — a `digitalocean_firewall` resource with the specified name, inbound rules, outbound rules, and Droplet targeting (by ID or tag)

## Prerequisites

- **DigitalOcean credentials** configured via environment variables or Planton provider config
- **Existing Droplets or Droplet tags** to which the firewall will be applied
- **Knowledge of the network ports and protocols** your services require

## Quick Start

Create a file `firewall.yaml`:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanFirewall
metadata:
  name: my-firewall
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.DigitalOceanFirewall.my-firewall
spec:
  name: my-firewall
  tags:
    - web
  inboundRules:
    - protocol: tcp
      portRange: "443"
      sourceAddresses:
        - "0.0.0.0/0"
        - "::/0"
  outboundRules:
    - protocol: tcp
      portRange: "1-65535"
      destinationAddresses:
        - "0.0.0.0/0"
        - "::/0"
    - protocol: udp
      portRange: "1-65535"
      destinationAddresses:
        - "0.0.0.0/0"
        - "::/0"
```

Deploy:

```shell
planton apply -f firewall.yaml
```

This creates a firewall that allows inbound HTTPS from any address and permits all outbound traffic, applied to every Droplet tagged `web`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `spec.name` | `string` | Name of the firewall in DigitalOcean. Must be unique per account. | Min length 1, max length 255 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `spec.inboundRules` | `InboundRule[]` | `[]` | Rules defining traffic allowed to reach the targeted Droplets. See Inbound Rule fields below. |
| `spec.outboundRules` | `OutboundRule[]` | `[]` | Rules defining traffic allowed to leave the targeted Droplets. See Outbound Rule fields below. |
| `spec.dropletIds` | `int64[]` | `[]` | Specific Droplet IDs to which this firewall is applied (max 10). Use tags for production workloads. |
| `spec.tags` | `string[]` | `[]` | Droplet tag names to which this firewall is applied. Any Droplet with a matching tag is protected by these rules. |

#### Inbound Rule Fields

Each entry in `inboundRules` accepts the following fields:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `protocol` | `string` | — | Required. One of `tcp`, `udp`, or `icmp`. |
| `portRange` | `string` | `""` | Port or range to allow (e.g., `"80"`, `"8000-9000"`, `"1-65535"`). Leave empty for ICMP. |
| `sourceAddresses` | `string[]` | `[]` | IPv4/IPv6 addresses or CIDR ranges (e.g., `"0.0.0.0/0"`, `"10.0.0.0/8"`). |
| `sourceDropletIds` | `int64[]` | `[]` | IDs of Droplets from which traffic is allowed. |
| `sourceTags` | `string[]` | `[]` | Droplet tag names; any Droplet with these tags is allowed as a source. |
| `sourceKubernetesIds` | `string[]` | `[]` | IDs of DigitalOcean Kubernetes clusters from which traffic is allowed. |
| `sourceLoadBalancerUids` | `string[]` | `[]` | IDs of DigitalOcean Load Balancers from which traffic is allowed. |

#### Outbound Rule Fields

Each entry in `outboundRules` accepts the following fields:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `protocol` | `string` | — | Required. One of `tcp`, `udp`, or `icmp`. |
| `portRange` | `string` | `""` | Port or range to allow (e.g., `"5432"`, `"1-65535"`). Leave empty for ICMP. |
| `destinationAddresses` | `string[]` | `[]` | IPv4/IPv6 addresses or CIDR ranges to which traffic is allowed. |
| `destinationDropletIds` | `int64[]` | `[]` | IDs of Droplets to which traffic is allowed. |
| `destinationTags` | `string[]` | `[]` | Droplet tag names whose members are allowed destinations. |
| `destinationKubernetesIds` | `string[]` | `[]` | IDs of DigitalOcean Kubernetes clusters to which traffic is allowed. |
| `destinationLoadBalancerUids` | `string[]` | `[]` | IDs of DigitalOcean Load Balancers which are allowed as destinations. |

## Examples

### Basic Web Server

A firewall for a web server that accepts HTTP and HTTPS from any address and allows all outbound traffic:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanFirewall
metadata:
  name: web-server-fw
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.DigitalOceanFirewall.web-server-fw
spec:
  name: web-server-firewall
  tags:
    - web
  inboundRules:
    - protocol: tcp
      portRange: "80"
      sourceAddresses:
        - "0.0.0.0/0"
        - "::/0"
    - protocol: tcp
      portRange: "443"
      sourceAddresses:
        - "0.0.0.0/0"
        - "::/0"
    - protocol: tcp
      portRange: "22"
      sourceAddresses:
        - "203.0.113.0/24"
  outboundRules:
    - protocol: tcp
      portRange: "1-65535"
      destinationAddresses:
        - "0.0.0.0/0"
        - "::/0"
    - protocol: udp
      portRange: "53"
      destinationAddresses:
        - "0.0.0.0/0"
        - "::/0"
```

### Multi-Tier Application

Separate firewalls for a web tier and a database tier. The web tier accepts HTTPS only from a Load Balancer and connects to the database tier via tags. The database tier accepts PostgreSQL connections only from the web tier:

**Web tier firewall:**

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanFirewall
metadata:
  name: app-web-fw
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.DigitalOceanFirewall.app-web-fw
spec:
  name: app-web-firewall
  tags:
    - web-tier
  inboundRules:
    - protocol: tcp
      portRange: "443"
      sourceLoadBalancerUids:
        - "lb-abc-123"
    - protocol: tcp
      portRange: "80"
      sourceLoadBalancerUids:
        - "lb-abc-123"
    - protocol: tcp
      portRange: "22"
      sourceAddresses:
        - "203.0.113.10/32"
  outboundRules:
    - protocol: tcp
      portRange: "5432"
      destinationTags:
        - db-tier
    - protocol: tcp
      portRange: "443"
      destinationAddresses:
        - "0.0.0.0/0"
        - "::/0"
    - protocol: udp
      portRange: "53"
      destinationAddresses:
        - "0.0.0.0/0"
        - "::/0"
```

**Database tier firewall:**

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanFirewall
metadata:
  name: app-db-fw
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.DigitalOceanFirewall.app-db-fw
spec:
  name: app-db-firewall
  tags:
    - db-tier
  inboundRules:
    - protocol: tcp
      portRange: "5432"
      sourceTags:
        - web-tier
    - protocol: tcp
      portRange: "22"
      sourceAddresses:
        - "203.0.113.10/32"
  outboundRules:
    - protocol: udp
      portRange: "53"
      destinationAddresses:
        - "0.0.0.0/0"
        - "::/0"
    - protocol: tcp
      portRange: "443"
      destinationAddresses:
        - "91.189.88.0/21"
```

### Multiple Source Types with Kubernetes Integration

A firewall that combines CIDR-based, tag-based, and Kubernetes cluster-based source rules, applied to specific Droplets by ID:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanFirewall
metadata:
  name: mixed-sources-fw
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.DigitalOceanFirewall.mixed-sources-fw
spec:
  name: mixed-sources-firewall
  dropletIds:
    - 386734086
    - 386734087
  tags:
    - backend
  inboundRules:
    - protocol: tcp
      portRange: "8080"
      sourceAddresses:
        - "10.0.0.0/8"
      sourceTags:
        - frontend
      sourceKubernetesIds:
        - "cluster-abc-123"
    - protocol: tcp
      portRange: "22"
      sourceAddresses:
        - "203.0.113.0/24"
    - protocol: icmp
      sourceAddresses:
        - "0.0.0.0/0"
        - "::/0"
  outboundRules:
    - protocol: tcp
      portRange: "6379"
      destinationTags:
        - cache-tier
    - protocol: tcp
      portRange: "5432"
      destinationTags:
        - db-tier
    - protocol: tcp
      portRange: "443"
      destinationAddresses:
        - "0.0.0.0/0"
        - "::/0"
    - protocol: udp
      portRange: "53"
      destinationAddresses:
        - "0.0.0.0/0"
        - "::/0"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `firewallId` | `string` | The unique identifier (UUID) of the created DigitalOcean firewall |

## Related Components

- [DigitalOceanDroplet](/docs/catalog/digitalocean/digitaloceandroplet) — provisions the Droplets that the firewall protects
- [DigitalOceanLoadBalancer](/docs/catalog/digitalocean/digitaloceanloadbalancer) — can be referenced as a source or destination in firewall rules via Load Balancer UIDs
- [DigitalOceanKubernetesCluster](/docs/catalog/digitalocean/digitaloceankubernetescluster) — can be referenced as a source or destination in firewall rules via Kubernetes cluster IDs
- [DigitalOceanVpc](/docs/catalog/digitalocean/digitaloceanvpc) — provides the network layer for Droplets protected by this firewall
- [DigitalOceanDatabaseCluster](/docs/catalog/digitalocean/digitaloceandatabasecluster) — commonly co-deployed with firewall rules restricting database port access
