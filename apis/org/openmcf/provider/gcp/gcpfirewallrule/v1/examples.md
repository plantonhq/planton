# GCP Firewall Rule Examples

This document provides comprehensive examples for deploying GCP firewall rules using OpenMCF. Each example includes the manifest YAML and explains the use case and key configuration choices.

## Table of Contents

- [Example 1: Basic INGRESS ALLOW for Web Traffic](#example-1-basic-ingress-allow-for-web-traffic)
- [Example 2: INGRESS ALLOW for SSH via IAP](#example-2-ingress-allow-for-ssh-via-iap)
- [Example 3: EGRESS DENY All (Restrictive Baseline)](#example-3-egress-deny-all-restrictive-baseline)
- [Example 4: INGRESS ALLOW with Service Account Targeting](#example-4-ingress-allow-with-service-account-targeting)
- [Example 5: INGRESS with Logging Enabled](#example-5-ingress-with-logging-enabled)
- [Example 6: INGRESS ALLOW Internal Traffic](#example-6-ingress-allow-internal-traffic)

---

## Example 1: Basic INGRESS ALLOW for Web Traffic

### Use Case

The most common firewall rule — allow inbound HTTP and HTTPS traffic from the internet to instances tagged as web servers. This is the starting point for any public-facing application.

### Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFirewallRule
metadata:
  name: allow-web
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpFirewallRule.allow-web
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
  priority: 1000
  description: Allow HTTP and HTTPS traffic from the internet to web servers
```

Deploy:

```shell
openmcf apply -f allow-web.yaml
```

### Key Choices

- **`sourceRanges: ["0.0.0.0/0"]`** — matches all IPv4 sources. For production, consider narrowing this to known CDN or load balancer ranges if applicable.
- **`targetTags: ["web-server"]`** — the rule only applies to VM instances that carry this network tag. Instances without the tag are unaffected.
- **`priority: 1000`** — the GCP default. Higher-priority DENY rules (lower number) will still block traffic if they match.
- **Two ports in one rule** — both `80` and `443` are handled by a single `GcpFirewallProtocolPort` entry because they share the same protocol (`tcp`).

---

## Example 2: INGRESS ALLOW for SSH via IAP

### Use Case

Allow SSH access to instances exclusively through Google Identity-Aware Proxy (IAP) TCP forwarding. This eliminates the need for public IP addresses on VMs while still enabling SSH management. Google's IAP tunnel originates from the `35.235.240.0/20` range.

### Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFirewallRule
metadata:
  name: allow-iap-ssh
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpFirewallRule.allow-iap-ssh
spec:
  projectId:
    value: my-gcp-project-123
  network:
    value: prod-vpc
  ruleName: allow-iap-ssh
  direction: INGRESS
  action: ALLOW
  rules:
    - protocol: tcp
      ports:
        - "22"
  sourceRanges:
    - "35.235.240.0/20"
  priority: 900
  description: Allow SSH from Google IAP tunnel IP range
```

Deploy:

```shell
openmcf apply -f allow-iap-ssh.yaml
```

### Key Choices

- **`sourceRanges: ["35.235.240.0/20"]`** — this is the well-known CIDR range used by Google's IAP TCP forwarding service. Only traffic originating from IAP is allowed; direct SSH from the internet is blocked.
- **No `targetTags`** — the rule applies to all instances in the network. To restrict to specific instances, add `targetTags`.
- **`priority: 900`** — slightly higher than the default `1000`, ensuring IAP SSH access works even if a default-deny rule exists at priority `1000`.
- **Single port** — only port `22` is needed for SSH.

### Usage

After deploying, connect via IAP:

```bash
gcloud compute ssh <instance-name> --tunnel-through-iap --project=my-gcp-project-123
```

---

## Example 3: EGRESS DENY All (Restrictive Baseline)

### Use Case

A "deny-all" egress rule at the lowest possible priority, establishing a restrictive baseline. All outbound traffic is blocked unless a higher-priority ALLOW rule explicitly permits it. This is a common pattern for compliance-driven environments (PCI-DSS, HIPAA) where egress filtering is required.

### Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFirewallRule
metadata:
  name: deny-all-egress
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpFirewallRule.deny-all-egress
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
  description: Deny all outbound traffic by default — use higher-priority ALLOW rules for exceptions
```

Deploy:

```shell
openmcf apply -f deny-all-egress.yaml
```

### Key Choices

- **`action: DENY`** — this is a deny rule. The `rules` field still populates `deny` blocks in the underlying GCP resource.
- **`protocol: all`** — matches every IP protocol (TCP, UDP, ICMP, etc.).
- **`priority: 65534`** — the lowest practical priority. GCP reserves `65535` for implied rules, so `65534` is the lowest user-configurable priority.
- **`destinationRanges: ["0.0.0.0/0"]`** — blocks traffic to all destinations. In practice, GCP defaults EGRESS rules to `0.0.0.0/0` when omitted, but being explicit improves readability.
- **No `targetTags`** — applies to all instances in the VPC. To scope it, add `targetTags`.

### Companion Allow Rule

Pair this with selective egress ALLOW rules at higher priority:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFirewallRule
metadata:
  name: allow-google-apis-egress
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpFirewallRule.allow-google-apis-egress
spec:
  projectId:
    value: my-prod-project-456
  network:
    value: prod-vpc
  ruleName: allow-google-apis-egress
  direction: EGRESS
  action: ALLOW
  rules:
    - protocol: tcp
      ports:
        - "443"
  priority: 1000
  destinationRanges:
    - "199.36.153.4/30"
  description: Allow HTTPS egress to Google restricted API VIPs
```

---

## Example 4: INGRESS ALLOW with Service Account Targeting

### Use Case

Instead of network tags, target instances by their attached service account. This is preferred in environments where VM creation is automated and service accounts are managed declaratively. A common pattern for GKE node pools and managed instance groups.

### Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFirewallRule
metadata:
  name: allow-backend-traffic
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpFirewallRule.allow-backend-traffic
spec:
  projectId:
    value: my-gcp-project-123
  network:
    value: prod-vpc
  ruleName: allow-backend-http
  direction: INGRESS
  action: ALLOW
  rules:
    - protocol: tcp
      ports:
        - "8080"
        - "8443"
  sourceServiceAccounts:
    - frontend@my-gcp-project-123.iam.gserviceaccount.com
  targetServiceAccounts:
    - backend@my-gcp-project-123.iam.gserviceaccount.com
  priority: 1000
  description: Allow frontend service to reach backend on application ports
```

Deploy:

```shell
openmcf apply -f allow-backend-traffic.yaml
```

### Key Choices

- **Service account targeting** — uses `sourceServiceAccounts` and `targetServiceAccounts` instead of tags. This means only VMs running as the `frontend@...` service account can initiate traffic, and only VMs running as `backend@...` receive it.
- **Mutual exclusion** — tag fields (`sourceTags`, `targetTags`) cannot be set when service account fields are used. GCP enforces this constraint.
- **Max 10 entries** — GCP limits both `sourceServiceAccounts` and `targetServiceAccounts` to 10 entries each.
- **No `sourceRanges`** — source filtering is entirely by service account identity, not IP range. The `sourceServiceAccounts` field satisfies the INGRESS source requirement.

---

## Example 5: INGRESS with Logging Enabled

### Use Case

Enable firewall logging to capture metadata about matched connections. Useful for security auditing, troubleshooting connectivity issues, and feeding data into SIEM systems via Cloud Logging exports.

### Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFirewallRule
metadata:
  name: allow-db-with-logging
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpFirewallRule.allow-db-with-logging
spec:
  projectId:
    value: my-gcp-project-123
  network:
    value: prod-vpc
  ruleName: allow-db-access-logged
  direction: INGRESS
  action: ALLOW
  rules:
    - protocol: tcp
      ports:
        - "5432"
  sourceRanges:
    - "10.10.0.0/16"
  targetTags:
    - postgres-server
  priority: 1000
  description: Allow PostgreSQL access from application subnet with logging
  logConfig:
    metadata: INCLUDE_ALL_METADATA
```

Deploy:

```shell
openmcf apply -f allow-db-with-logging.yaml
```

### Key Choices

- **`logConfig.metadata: INCLUDE_ALL_METADATA`** — logs include full metadata (source/destination IP, port, protocol, instance details). Use `EXCLUDE_ALL_METADATA` to reduce log volume while still capturing connection events.
- **`sourceRanges: ["10.10.0.0/16"]`** — only traffic from the application subnet is allowed, not the entire VPC.
- **`targetTags: ["postgres-server"]`** — only instances tagged as database servers are affected.
- **Cost consideration** — firewall logging generates Cloud Logging entries that incur charges at volume. Use `EXCLUDE_ALL_METADATA` or limit logging to critical rules in high-throughput environments.

### Viewing Logs

After deployment, query firewall logs:

```bash
gcloud logging read 'resource.type="gce_subnetwork" AND jsonPayload.rule_details.reference="network:prod-vpc/firewall:allow-db-access-logged"' \
  --project=my-gcp-project-123 \
  --limit=10
```

---

## Example 6: INGRESS ALLOW Internal Traffic

### Use Case

Allow all internal traffic within the RFC 1918 `10.0.0.0/8` range. This is a common foundational rule for VPCs where internal communication between application tiers, monitoring agents, and service meshes should be unrestricted.

### Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFirewallRule
metadata:
  name: allow-internal
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpFirewallRule.allow-internal
spec:
  projectId:
    value: my-gcp-project-123
  network:
    value: prod-vpc
  ruleName: allow-internal-all
  direction: INGRESS
  action: ALLOW
  rules:
    - protocol: tcp
      ports:
        - "0-65535"
    - protocol: udp
      ports:
        - "0-65535"
    - protocol: icmp
  sourceRanges:
    - "10.0.0.0/8"
  priority: 1000
  description: Allow all internal TCP, UDP, and ICMP traffic within 10.0.0.0/8
```

Deploy:

```shell
openmcf apply -f allow-internal.yaml
```

### Key Choices

- **Three protocol entries** — GCP firewall rules require separate `allow` blocks per protocol. TCP and UDP each get a full port range (`0-65535`), and ICMP has no ports.
- **`sourceRanges: ["10.0.0.0/8"]`** — covers the entire `10.x.x.x` private address space. Narrow this to your actual VPC CIDR (e.g., `10.10.0.0/16`) for tighter security.
- **No `targetTags`** — applies to all instances in the network. This is intentional for a broad "allow internal" rule.
- **`protocol: icmp`** — included so internal health checks and `ping` diagnostics work between instances.

### Tighter Alternative

For environments where you want internal communication only within specific subnets:

```yaml
spec:
  # ... same as above, but with narrower source ranges
  sourceRanges:
    - "10.10.0.0/16"
    - "10.20.0.0/16"
  rules:
    - protocol: tcp
      ports:
        - "0-65535"
    - protocol: udp
      ports:
        - "0-65535"
    - protocol: icmp
```

---

## Common Patterns

### Pattern: Defense in Depth

Combine multiple rules with different priorities to build layered security:

1. **Priority 100**: Allow health check probes from Google ranges
2. **Priority 900**: Allow IAP SSH from `35.235.240.0/20`
3. **Priority 1000**: Allow application-specific traffic (HTTP, gRPC, database)
4. **Priority 65534**: Deny all remaining INGRESS/EGRESS

### Pattern: Tag vs. Service Account

| Approach | When to Use |
|----------|-------------|
| **Network tags** | Simple setups, manual VM creation, quick prototyping |
| **Service accounts** | Automated environments, GKE, Terraform/Pulumi-managed VMs, stronger identity guarantees |

Tags can be added/removed by anyone with `compute.instances.setTags` permission. Service accounts provide stronger identity because only project IAM admins can assign them.

### Pattern: Using Foreign Key References

Reference a GcpVpc resource instead of hardcoding the network:

```yaml
spec:
  network:
    ref:
      kind: GcpVpc
      name: my-vpc
  # ...
```

OpenMCF resolves the VPC's `network_self_link` output and passes it as the `network` argument.

---

## Deployment Commands

### Using the OpenMCF CLI

```bash
openmcf apply -f <filename>.yaml
```

### Using Pulumi Directly

```bash
cd iac/pulumi
pulumi stack init dev
pulumi up
```

### Using Terraform Directly

```bash
cd iac/tf
terraform init
terraform plan
terraform apply
```

---

## Additional Resources

- [GCP Firewall Rules Overview](https://cloud.google.com/vpc/docs/firewalls)
- [IAP TCP Forwarding](https://cloud.google.com/iap/docs/using-tcp-forwarding)
- [Firewall Rules Logging](https://cloud.google.com/vpc/docs/firewall-rules-logging)
- [VPC Best Practices](https://cloud.google.com/architecture/best-practices-vpc-design)

---

For more details, see the [main README](README.md).
