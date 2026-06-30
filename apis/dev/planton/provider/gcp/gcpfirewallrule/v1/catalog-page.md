# GCP Firewall Rule

Deploys a GCP compute firewall rule (`google_compute_firewall`) with an `action` + `rules` abstraction for allow/deny traffic control, supporting INGRESS/EGRESS directions, CIDR-based source/destination filtering, network tags, service account targeting, priority, logging, and a disabled toggle.

## What Gets Created

When you deploy a GcpFirewallRule resource, Planton provisions:

- **Firewall Rule** — a `google_compute_firewall` resource in the specified VPC network with allow or deny blocks generated from the `action` and `rules` fields

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId`
- **An existing VPC network** — referenced via `network` (name or self-link)
- **Compute Engine API enabled** (`compute.googleapis.com`) on the target project
- **IAM permissions** — `roles/compute.securityAdmin` or `roles/compute.networkAdmin` on the target project

## Quick Start

Create a file `firewall.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpFirewallRule
metadata:
  name: allow-web
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpFirewallRule.allow-web
spec:
  projectId:
    value: my-gcp-project-123
  network:
    value: my-vpc
  ruleName: allow-http-https
  direction: INGRESS
  action: ALLOW
  rules:
    - protocol: tcp
      ports:
        - "80"
        - "443"
  sourceRanges:
    - "0.0.0.0/0"
  targetTags:
    - web-server
```

Deploy:

```shell
planton apply -f firewall.yaml
```

This creates an INGRESS firewall rule that allows HTTP and HTTPS traffic from any source to instances tagged `web-server` in the `my-vpc` network.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project ID where the firewall rule is created. Can reference a GcpProject resource. | Required |
| `network` | `StringValueOrRef` | VPC network name or self-link URL. Can reference a GcpVpc resource. | Required |
| `ruleName` | `string` | Name of the firewall rule in GCP. | 1-63 chars, lowercase letters/numbers/hyphens |
| `direction` | `string` | Traffic direction: `INGRESS` or `EGRESS`. | Required |
| `action` | `string` | Action when traffic matches: `ALLOW` or `DENY`. | Required |
| `rules` | `GcpFirewallProtocolPort[]` | Protocol and port combinations to match. | Min 1 item |
| `rules[].protocol` | `string` | IP protocol: `tcp`, `udp`, `icmp`, `esp`, `ah`, `sctp`, `ipip`, `all`, or an IANA protocol number. | Required |
| `rules[].ports` | `string[]` | Ports or port ranges (e.g., `"80"`, `"8000-9000"`). Only applicable for `tcp` or `udp`. Omit for other protocols. | — |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `priority` | `int32` | `1000` | Rule priority (0-65535). Lower = higher priority. |
| `description` | `string` | `""` | Human-readable description. |
| `sourceRanges` | `string[]` | — | Source CIDRs for INGRESS rules. |
| `destinationRanges` | `string[]` | — | Destination CIDRs for EGRESS rules. |
| `sourceTags` | `string[]` | — | Source instance network tags (INGRESS). Mutually exclusive with service account fields. |
| `targetTags` | `string[]` | — | Target instance network tags. Mutually exclusive with service account fields. |
| `sourceServiceAccounts` | `string[]` | — | Source service accounts for INGRESS (max 10). Mutually exclusive with tag fields. |
| `targetServiceAccounts` | `string[]` | — | Target service accounts (max 10). Mutually exclusive with tag fields. |
| `disabled` | `bool` | `false` | When `true`, rule exists but is not enforced. |
| `logConfig.metadata` | `string` | — | Enables logging. Must be `EXCLUDE_ALL_METADATA` or `INCLUDE_ALL_METADATA`. |

## Examples

### Basic INGRESS ALLOW for Web Traffic

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpFirewallRule
metadata:
  name: allow-web
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpFirewallRule.allow-web
spec:
  projectId:
    value: my-gcp-project-123
  network:
    value: my-vpc
  ruleName: allow-http-https
  direction: INGRESS
  action: ALLOW
  rules:
    - protocol: tcp
      ports:
        - "80"
        - "443"
  sourceRanges:
    - "0.0.0.0/0"
  targetTags:
    - web-server
```

### INGRESS ALLOW for SSH via IAP

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpFirewallRule
metadata:
  name: allow-iap-ssh
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpFirewallRule.allow-iap-ssh
spec:
  projectId:
    value: my-gcp-project-123
  network:
    value: my-vpc
  ruleName: allow-iap-ssh
  direction: INGRESS
  action: ALLOW
  rules:
    - protocol: tcp
      ports:
        - "22"
  sourceRanges:
    - "35.235.240.0/20"
  description: Allow SSH from Google IAP tunnel IP range
```

### EGRESS DENY All (Restrictive Baseline)

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpFirewallRule
metadata:
  name: deny-all-egress
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpFirewallRule.deny-all-egress
spec:
  projectId:
    value: my-prod-project-456
  network:
    value: prod-vpc
  ruleName: deny-all-egress
  direction: EGRESS
  action: DENY
  rules:
    - protocol: all
  priority: 65534
  destinationRanges:
    - "0.0.0.0/0"
  description: Deny all outbound traffic by default
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `firewall_self_link` | `string` | Full self-link URI of the firewall rule (e.g., `projects/my-project/global/firewalls/allow-http-https`) |
| `firewall_name` | `string` | Name of the firewall rule as it exists in GCP |
| `creation_timestamp` | `string` | RFC 3339 timestamp of when the firewall rule was created |

## Related Components

- [GcpVpc](/docs/catalog/gcp/gcpvpc) — provides the VPC network that firewall rules are attached to
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project and enables the Compute Engine API
- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — creates subnets within the VPC
- [GcpGkeCluster](/docs/catalog/gcp/gcpgkecluster) — deploys GKE clusters that benefit from firewall rules for node traffic control
