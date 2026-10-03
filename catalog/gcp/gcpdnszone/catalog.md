# GCP DNS Zone

Deploys a Cloud DNS managed zone — the authoritative container for one domain. Public zones answer the internet once the domain is delegated to the assigned nameservers; private zones answer only inside the VPC networks and GKE clusters you attach, and can alternatively forward queries to upstream resolvers or peer with another VPC's Cloud DNS. The zone owns the shell only: DNS records are separate GcpDnsRecord Infra Components referencing this zone by name, so records deploy, change, and destroy independently of the zone.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cloud DNS API enablement** -- `dns.googleapis.com` is enabled in the target project (never disabled on destroy, so tearing down one zone cannot break the rest of the project)
- **Cloud DNS Managed Zone** -- public (internet-facing) or private (VPC-scoped), for the domain in `dnsName` (trailing dot required; when omitted, derived from the resource metadata name plus a dot)
- **Private visibility bindings** -- created only for private zones with `privateVisibilityConfig`; attaches the zone to the listed VPC networks and GKE clusters
- **Forwarding configuration** -- created only when `forwardingConfig` lists target name servers; queries for the domain forward to those upstream resolvers (hybrid/on-prem DNS)
- **DNS peering** -- created only when `peeringConfig` names a target network; queries resolve using that VPC's Cloud DNS configuration
- **DNSSEC signing** -- created only when `dnssecConfig` is set on a public zone; optional custom key specs and NSEC/NSEC3 denial of existence
- **Query logging** -- created only when `cloudLoggingConfig.enableLogging` is true; every answered query lands in Cloud Logging
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) merged with `labels` and applied for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the managed zone will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef; the module enables the Cloud DNS API itself.
- **Domain registrar access** (only for public zones) to update nameserver (NS) records for the domain after zone creation.
- **VPC networks or GKE clusters** (only for private zones) that the zone will be visible to — reference GcpVpcNetwork / GcpGkeCluster resources or supply their URLs.

## Deploy

### Console

Open the deployment store, find **GCP DNS Zone**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields — the visibility choice forks the flow into private resolution (private zones) or DNSSEC (public zones). Start from the **Public DNS Zone** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDnsZone
metadata:
  name: prod-example-zone
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  dnsName: "example.com."
  description: "Prod apex for example.com"
```

```shell
planton apply -f dns-zone.yaml
```

This creates a public managed zone for `example.com.` with GCP-assigned nameservers and no DNSSEC; update your domain registrar's NS records with the nameservers from the outputs to delegate the domain. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the DNS zone to a GCP project deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  dnsName: "internal.example.com."
  visibility: private
  privateVisibilityConfig:
    networks:
      - networkUrl:
          valueFrom:
            kind: GcpVpcNetwork
            name: prod-vpc
            fieldPath: status.outputs.network_self_link
```

The InfraPipeline resolves the dependency graph, deploys the project and network first, then provisions the zone with the resolved references.

## Key Configuration

These are the most important decisions when configuring a DNS zone. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**DNS name** -- The authoritative domain, in absolute form with the trailing dot (`example.com.`). Immutable: a different domain is a different zone with different nameservers. When omitted, the modules derive it from the resource metadata name plus a dot.

**Visibility** -- `public` answers the internet after registrar delegation; `private` answers only the attached networks and clusters. Immutable — split-horizon DNS (the same domain public AND private) is two zones with the same DNS name, one of each visibility. Untouched, GCP defaults to public.

**Private resolution mode** -- A standard private zone holds its own records and lists visibility targets in `privateVisibilityConfig`. A forwarding zone sends every query to `forwardingConfig.targetNameServers` (on-prem/hybrid DNS — each target needs an IPv4 address or domain name; forwarding and peering are mutually exclusive). A peering zone resolves through another VPC's Cloud DNS via `peeringConfig.targetNetwork` (the shared-services pattern).

**DNSSEC** -- Public zones only. Turning signing `on` is half the chain: publish the DS record at the registrar to activate validation, and never delete a signed zone without removing the DS record first. Key specs are optional — omitted, Cloud DNS generates modern defaults.

**Destroy levers** -- Unset, `forceDestroy` leaves the safety net in place: a zone destroy fails while non-default record sets exist. Armed, destroying the zone deletes every record set in it first — sane for ephemeral zones, dangerous for shared production domains. `deletionPolicy` decides the zone shell itself: `PREVENT` protects a zone that registrars and parent zones delegate to; `ABANDON` removes it from management while it keeps serving DNS.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** (private zones) | `privateVisibilityConfig.networks[].networkUrl` | `status.outputs.network_self_link` |
| **GcpGkeCluster** (private zones) | `privateVisibilityConfig.gkeClusters[].gkeClusterName` | `status.outputs.cluster_id` |
| **GcpVpcNetwork** (peering zones) | `peeringConfig.targetNetwork` | `status.outputs.network_self_link` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `zone_id` | Numeric ID the Cloud DNS API assigns to the zone | GCP console links, API references |
| `zone_name` | Managed zone resource name | GcpDnsRecord `managedZone` field via ValueFromRef |
| `nameservers` | Nameservers assigned to the zone | Domain registrar NS delegation (public zones) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Public zone with composed records** -- The zone owns the shell; each record is a standalone GcpDnsRecord referencing `zone_name`, so records deploy, change, and destroy independently. Start from the **Public DNS Zone** preset.

**Certificate validation chain** -- A GcpCertManagerDnsAuthorization exports a CNAME triple; a GcpDnsRecord serves it in this zone; the GcpCertManagerCert validates automatically — TLS issuance before any load balancer exists.

**Private service names** -- A private zone attached to the prod VPC holds internal service names that never leak to the internet; GKE-scoped attachment narrows visibility to a single cluster.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the managed zone is created
- [**GCP DNS Record**](/infra-catalog/gcp-dns-record) -- the record sets served from this zone, referencing its `zone_name` output
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- visibility targets for private zones and peering producers
- [**GCP GKE Cluster**](/infra-catalog/gcp-gke-cluster) -- narrows a private zone's visibility to a single cluster
- [**GCP Cert Manager DNS Authorization**](/infra-catalog/gcp-cert-manager-dns-authorization) -- exports the validation record a GcpDnsRecord serves in this zone
