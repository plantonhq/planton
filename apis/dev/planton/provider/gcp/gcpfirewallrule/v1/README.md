# GCP Firewall Rule

Deploys a GCP compute firewall rule (`google_compute_firewall`) with an `action` + `rules` abstraction, supporting INGRESS/EGRESS directions, source/destination CIDR ranges, network tags, service account targeting, priority, logging, and a disabled toggle.

## What Gets Created

When you deploy a GcpFirewallRule resource, Planton provisions:

- **Firewall Rule** — a `google_compute_firewall` resource in the specified VPC network with allow or deny blocks generated from the `action` and `rules` fields

No additional supporting resources (API enablement, labels on the firewall, etc.) are created. The module assumes the Compute Engine API is already enabled on the target project.

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

This creates an INGRESS ALLOW rule permitting TCP ports 80 and 443 from any source to instances tagged `web-server`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project ID where the firewall rule is created. Can reference a GcpProject resource. | Required |
| `network` | `StringValueOrRef` | VPC network name or self-link URL. Can reference a GcpVpc resource. | Required |
| `ruleName` | `string` | Name of the firewall rule in GCP. | 1-63 chars, lowercase letters/numbers/hyphens, must start with a letter and end with a letter or number |
| `direction` | `string` | Traffic direction: `INGRESS` or `EGRESS`. | Required, must be `INGRESS` or `EGRESS` |
| `action` | `string` | Action when traffic matches: `ALLOW` or `DENY`. | Required, must be `ALLOW` or `DENY` |
| `rules` | `GcpFirewallProtocolPort[]` | Protocol and port combinations to match. At least one entry required. | Min 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `priority` | `int32` | `1000` | Rule priority (0-65535). Lower = higher priority. At the same priority, DENY beats ALLOW. |
| `description` | `string` | `""` | Human-readable description for the rule. |
| `sourceRanges` | `string[]` | — | Source CIDRs for INGRESS rules. E.g., `["0.0.0.0/0"]` for all traffic. Required for INGRESS unless source_tags or source_service_accounts are set. |
| `destinationRanges` | `string[]` | — | Destination CIDRs for EGRESS rules. GCP defaults to `["0.0.0.0/0"]` if omitted. |
| `sourceTags` | `string[]` | — | Source instance network tags for INGRESS rules. Mutually exclusive with service account fields. |
| `targetTags` | `string[]` | — | Target instance network tags. If omitted, rule applies to all instances. Mutually exclusive with service account fields. |
| `sourceServiceAccounts` | `string[]` | — | Source service accounts for INGRESS rules (max 10). Mutually exclusive with tag fields. |
| `targetServiceAccounts` | `string[]` | — | Target service accounts (max 10). Mutually exclusive with tag fields. |
| `disabled` | `bool` | `false` | When `true`, the rule exists but is not enforced. |
| `logConfig.metadata` | `string` | — | When present, enables firewall logging. Must be `EXCLUDE_ALL_METADATA` or `INCLUDE_ALL_METADATA`. |

### GcpFirewallProtocolPort

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `protocol` | `string` | Yes | IP protocol: `tcp`, `udp`, `icmp`, `esp`, `ah`, `sctp`, `ipip`, `all`, or an IANA protocol number. |
| `ports` | `string[]` | No | Ports or port ranges (e.g., `"80"`, `"8000-9000"`). Only for `tcp` and `udp`. Omit for other protocols. |

### Validation Rules

- **INGRESS rules** must specify at least one of `sourceRanges`, `sourceTags`, or `sourceServiceAccounts`.
- **Tag-based** targeting (`sourceTags`, `targetTags`) and **service-account-based** targeting (`sourceServiceAccounts`, `targetServiceAccounts`) are mutually exclusive. You cannot use both in the same rule.

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
| `firewallSelfLink` | `string` | Full self-link URI of the firewall rule (e.g., `projects/my-project/global/firewalls/allow-http-https`) |
| `firewallName` | `string` | Name of the firewall rule as it exists in GCP |
| `creationTimestamp` | `string` | RFC 3339 timestamp of when the firewall rule was created |

## Deployment Methods

Planton supports two deployment methods:

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Examples

For more comprehensive examples, see [`examples.md`](examples.md), including:

- Basic web traffic INGRESS ALLOW
- SSH via IAP
- Restrictive EGRESS DENY baseline
- Service account targeting
- Firewall logging
- Internal traffic rules

## Related Components

- [GcpVpc](/docs/catalog/gcp/gcpvpc) — provides the VPC network that firewall rules are attached to
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project and enables the Compute Engine API
- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — creates subnets within the VPC
- [GcpGkeCluster](/docs/catalog/gcp/gcpgkecluster) — deploys GKE clusters that benefit from firewall rules for node traffic control

## Additional Resources

- [GCP Firewall Rules Documentation](https://cloud.google.com/vpc/docs/firewalls)
- [VPC Firewall Rules Overview](https://cloud.google.com/vpc/docs/firewalls#firewall_rule_components)
- [IAP TCP Forwarding](https://cloud.google.com/iap/docs/using-tcp-forwarding)
- [Firewall Rules Logging](https://cloud.google.com/vpc/docs/firewall-rules-logging)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.
