---
title: "Load Balancer"
description: "Load Balancer deployment documentation"
icon: "package"
order: 100
componentName: "digitaloceanloadbalancer"
---

# DigitalOcean Load Balancer

Deploys a managed regional load balancer on DigitalOcean with configurable forwarding rules, health checks, sticky sessions, and backend targeting via Droplet IDs or tags. The component provisions the load balancer inside a VPC for private-network communication with backend Droplets.

## What Gets Created

When you deploy a DigitalOceanLoadBalancer resource, OpenMCF provisions:

- **Load Balancer** -- a `digitalocean_loadbalancer` resource in the specified region and VPC, with one or more forwarding rules that define how traffic is routed from the load balancer to backend Droplets
- **Forwarding Rules** -- each rule maps an entry port/protocol on the load balancer to a target port/protocol on the backend, with optional TLS certificate for HTTPS termination
- **Health Check** -- created only when `healthCheck` is specified, probes backend Droplets at a configurable interval to determine availability
- **Sticky Sessions** -- created only when `enableStickySessions` is `true`, configures cookie-based session affinity so repeated requests from the same client reach the same Droplet

## Prerequisites

- **DigitalOcean credentials** configured via environment variables or OpenMCF provider config
- **A DigitalOcean VPC** in the target region (can reference a DigitalOceanVpc resource via `valueFrom`)
- **At least one backend target** -- either a list of Droplet IDs or a Droplet tag that matches running Droplets

## Quick Start

Create a file `lb.yaml`:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanLoadBalancer
metadata:
  name: my-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.DigitalOceanLoadBalancer.my-lb
spec:
  loadBalancerName: my-lb
  region: nyc3
  vpc:
    value: "vpc-uuid-here"
  forwardingRules:
    - entryPort: 80
      entryProtocol: http
      targetPort: 80
      targetProtocol: http
  dropletTag: web-dev
```

Deploy:

```shell
openmcf apply -f lb.yaml
```

This creates an HTTP load balancer in the NYC3 region that routes port 80 traffic to all Droplets tagged `web-dev` within the specified VPC.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `loadBalancerName` | `string` | Name of the load balancer in DigitalOcean. Must be unique per account. | Required, 1-64 characters, lowercase alphanumeric and hyphens (`^[a-z0-9-]+$`) |
| `region` | `enum` | DigitalOcean region for the load balancer. Valid values: `nyc3`, `sfo3`, `fra1`, `sgp1`, `lon1`, `tor1`, `blr1`, `ams3`. | Required |
| `vpc` | `StringValueOrRef` | UUID of the VPC in which to place the load balancer. Can reference a DigitalOceanVpc resource via `valueFrom`. Resolves `status.outputs.vpc_id` from the referenced resource. | Required |
| `forwardingRules` | `ForwardingRule[]` | One or more rules that define how inbound traffic is routed to backend Droplets. | Required, minimum 1 rule |

### Forwarding Rule Fields

Each entry in `forwardingRules` contains:

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `entryPort` | `uint32` | Port on the load balancer that listens for incoming traffic. | Required, 1-65535 |
| `entryProtocol` | `enum` | Protocol for incoming traffic. Valid values: `http`, `https`, `tcp`. | Required |
| `targetPort` | `uint32` | Port on the backend Droplet that receives forwarded traffic. | Required, 1-65535 |
| `targetProtocol` | `enum` | Protocol for traffic between the load balancer and the Droplet. Valid values: `http`, `https`, `tcp`. | Required |
| `certificateName` | `string` | Name of a TLS certificate uploaded to DigitalOcean. Required when `entryProtocol` is `https`. Use the certificate name (not ID) to avoid breaking IaC state when Let's Encrypt auto-renews. | Optional, 1-255 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `healthCheck` | `HealthCheck` | -- | Health check configuration for backend Droplets. See Health Check Fields below. |
| `dropletIds` | `StringValueOrRef[]` | `[]` | Specific Droplet IDs to attach to the load balancer. Can reference DigitalOceanDroplet resources via `valueFrom`. Mutually exclusive with `dropletTag`. |
| `dropletTag` | `string` | -- | A Droplet tag name. All Droplets with this tag in the VPC are automatically attached. Mutually exclusive with `dropletIds`. 1-255 characters. |
| `enableStickySessions` | `bool` | `false` | When `true`, enables cookie-based sticky sessions so repeated requests from the same client are directed to the same Droplet. |

### Health Check Fields

When `healthCheck` is specified:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `port` | `uint32` | -- | Port on the Droplet to probe. | Required, 1-65535 |
| `protocol` | `enum` | -- | Protocol for health checks. Valid values: `http`, `https`, `tcp`. | Required |
| `path` | `string` | -- | Request path for HTTP/HTTPS health checks (e.g., `/health`). Ignored for TCP. |
| `checkIntervalSec` | `uint32` | `10` | Interval in seconds between health check probes. |

## Examples

### HTTP Load Balancer

A basic HTTP load balancer for development or testing, using tag-based backend targeting:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanLoadBalancer
metadata:
  name: dev-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.DigitalOceanLoadBalancer.dev-lb
spec:
  loadBalancerName: dev-lb
  region: nyc3
  vpc:
    value: "vpc-dev-uuid"
  forwardingRules:
    - entryPort: 80
      entryProtocol: http
      targetPort: 80
      targetProtocol: http
  dropletTag: web-dev
```

### HTTPS with TLS Certificate

A production load balancer that terminates TLS at the load balancer and forwards HTTP to backend Droplets. The `certificateName` field references a TLS certificate already uploaded to DigitalOcean (use the certificate name, not its ID):

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanLoadBalancer
metadata:
  name: prod-web-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanLoadBalancer.prod-web-lb
spec:
  loadBalancerName: prod-web-lb
  region: sfo3
  vpc:
    value: "vpc-prod-uuid"
  forwardingRules:
    - entryPort: 443
      entryProtocol: https
      targetPort: 80
      targetProtocol: http
      certificateName: my-le-cert
  dropletTag: web-prod
  healthCheck:
    port: 80
    protocol: http
    path: "/healthz"
```

### Full-Featured with Health Check, Sticky Sessions, and VPC Reference

Production configuration using a VPC foreign key reference, explicit health check tuning, sticky sessions, and multiple forwarding rules:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanLoadBalancer
metadata:
  name: full-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanLoadBalancer.full-lb
spec:
  loadBalancerName: full-lb
  region: fra1
  vpc:
    valueFrom:
      kind: DigitalOceanVpc
      name: prod-vpc
      field: status.outputs.vpc_id
  forwardingRules:
    - entryPort: 443
      entryProtocol: https
      targetPort: 8080
      targetProtocol: http
      certificateName: prod-cert
    - entryPort: 80
      entryProtocol: http
      targetPort: 8080
      targetProtocol: http
  healthCheck:
    port: 8080
    protocol: http
    path: "/health"
    checkIntervalSec: 15
  dropletTag: app-prod
  enableStickySessions: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `load_balancer_id` | `string` | UUID of the created DigitalOcean load balancer |
| `ip` | `string` | Public IP address assigned to the load balancer |
| `dns_name` | `string` | DNS name for the load balancer. DigitalOcean does not expose an explicit DNS field; the load balancer name is exported as a placeholder. |

## Related Components

- [DigitalOceanVpc](/docs/catalog/digitalocean/vpc) -- provides the VPC for load balancer placement
- [DigitalOceanDroplet](/docs/catalog/digitalocean/droplet) -- backend compute instances that receive traffic from the load balancer
- [DigitalOceanKubernetesCluster](/docs/catalog/digitalocean/kubernetes-cluster) -- managed Kubernetes cluster whose services can be exposed through load balancers
- [DigitalOceanFirewall](/docs/catalog/digitalocean/firewall) -- controls network access to backend Droplets
