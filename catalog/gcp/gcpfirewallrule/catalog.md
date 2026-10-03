# GCP Firewall Rule

Deploys a Compute Engine firewall rule that controls inbound or outbound traffic to VM instances in a VPC network. Each rule allows or denies traffic matching protocol/port combinations, filtered by source or destination CIDR ranges and scoped by network tags or service accounts, with optional per-rule logging and a priority that decides which rule wins when several match.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Compute Engine Firewall Rule** -- a `compute.Firewall` in the specified project and VPC network, configured with the direction (INGRESS or EGRESS), action (ALLOW or DENY), protocol/port rules, and traffic source or destination filters
- **Traffic Matching Rules** -- one or more protocol/port blocks mapped to either allow or deny entries based on the `action` field
- **Logging Configuration** -- created only when `logConfig` is present; enables firewall logging with configurable metadata inclusion (INCLUDE_ALL_METADATA or EXCLUDE_ALL_METADATA)
- **Resource Manager tag bindings** -- created only when `resourceManagerTags` is set; bound at create time only, and changing them REPLACES the firewall rule

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the firewall rule will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **A VPC network** to attach the firewall rule to. Provide the network name or self-link directly, or reference a GcpVpcNetwork Infra Component via ValueFromRef.
- **Compute Engine API** (`compute.googleapis.com`) enabled in the target project.

## Deploy

### Console

Open the deployment store, find **GCP Firewall Rule**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Ingress Allow Web (HTTP/HTTPS)** preset in the [Presets](#presets) tab to pre-populate a standard HTTP/HTTPS ingress rule.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirewallRule
metadata:
  name: allow-web
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  network:
    value: "main-vpc"
  ruleName: "allow-http-https"
  direction: INGRESS
  action: ALLOW
  rules:
    - protocol: tcp
      ports: ["80", "443"]
  sourceRanges: ["0.0.0.0/0"]
```

```shell
planton apply -f gcp-firewall-rule.yaml
```

This creates an ingress rule allowing HTTP and HTTPS traffic from any source to all instances in the VPC. No target tags are set, so the rule applies network-wide. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the firewall rule to a GCP project and VPC deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: main-vpc
      fieldPath: status.outputs.network_self_link
```

The InfraPipeline resolves the dependency graph, deploys the project and VPC first, then provisions the firewall rule with the resolved values.

## Key Configuration

These are the most important decisions when configuring a firewall rule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Direction and action** -- Set `direction` to INGRESS for inbound rules or EGRESS for outbound rules, and `action` to ALLOW or DENY. INGRESS rules require at least one of `sourceRanges`, `sourceTags`, or `sourceServiceAccounts`. EGRESS rules default to matching all destinations if `destinationRanges` is omitted.

**Priority** -- `priority` defaults to 1000 (range 0-65535, lower is higher priority). At the same priority, DENY rules take precedence over ALLOW. Use priorities below 1000 for emergency overrides and above 1000 for baseline policies. The implied GCP allow-all egress rule sits at 65535.

**Target scoping** -- Without `targetTags` or `targetServiceAccounts`, the rule applies to all instances in the VPC. Add network tags to restrict which VMs are affected. Tag-based targeting (`sourceTags`, `targetTags`) and service-account-based targeting (`sourceServiceAccounts`, `targetServiceAccounts`) are mutually exclusive -- you cannot use both in the same rule.

**Logging** -- Omit `logConfig` to disable logging (the default). Set `logConfig.metadata` to `INCLUDE_ALL_METADATA` for full audit trails or `EXCLUDE_ALL_METADATA` for reduced log volume. Firewall logs are written to Cloud Logging and can be significant for high-traffic rules.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `network` | `status.outputs.network_self_link` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `firewall_self_link` | Full self-link URI of the firewall rule | Dependency ordering in InfraCharts |
| `firewall_name` | Name of the firewall rule as it exists in GCP | gcloud commands, log filters on `firewall_rule_name` |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Ingress allow web** -- Allows inbound HTTP (port 80) and HTTPS (port 443) from all IPv4 addresses. The standard rule for internet-facing web servers and load balancer backends. Start from the **Ingress Allow Web (HTTP/HTTPS)** preset.

**Ingress allow SSH via IAP** -- Allows SSH (port 22) exclusively from Google's Identity-Aware Proxy range (`35.235.240.0/20`). Provides authenticated, audited SSH access to VMs without exposing port 22 to the public internet. Scoped to VMs tagged with `allow-ssh`. Start from the **Ingress Allow SSH via IAP** preset.

**Egress deny all** -- Denies all outbound traffic at priority 65534, overriding GCP's implied allow-all egress rule. Establishes a deny-by-default egress posture for compliance-driven environments. Add higher-priority allow rules for approved destinations. Start from the **Egress Deny All** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the firewall rule is created
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- provides the VPC network that the firewall rule is attached to