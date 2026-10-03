# GCP Region Network Endpoint Group

Deploys a regional network endpoint group (NEG) — the bridge that lets a load balancer's backend service send traffic to something other than Compute Engine VMs. A serverless NEG (the default) fronts a Cloud Run service, a Cloud Functions function, or an App Engine app — which is how serverless workloads get custom domains, Cloud CDN, Cloud Armor, and IAP in front of them. Other endpoint types front Private Service Connect endpoints, internet origins (by IP or FQDN), or PSC port-mapped VM targets. Every field is immutable in GCP, so reshaping a live NEG is a create-before-destroy operation — the replacement must exist before the original can be released.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Compute Engine API enablement** -- the module enables `compute.googleapis.com` in the target project first, so a fresh project works on the first deploy (never disabled on destroy)
- **Regional Network Endpoint Group** -- scoped to one region, holding endpoints of the selected type
- **Target wiring** -- the serverless target block (Cloud Run / Cloud Function / App Engine), or the PSC/internet target and its VPC facts

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the NEG will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef. The module enables the Compute Engine API itself — no manual API setup is needed.
- **The fronted workload's region** -- a serverless NEG must live in the SAME region as the Cloud Run/Functions/App Engine workload it fronts (the workload itself need not exist yet — GCP resolves endpoints at serving time).

## Deploy

### Console

Open the deployment store, find **GCP Region Network Endpoint Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Cloud Run behind a Load Balancer** preset in the [Presets](#presets) tab to pre-populate the dominant serverless shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpRegionNetworkEndpointGroup
metadata:
  name: orders-run-neg
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  region: us-central1
  cloudRun:
    service:
      value: "orders-api"
```

```shell
planton apply -f region-neg.yaml
```

This creates a serverless NEG fronting the `orders-api` Cloud Run service — ready to be referenced by a backend service. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the NEG to the workload it fronts:

```yaml
spec:
  region: us-central1
  cloudRun:
    service:
      valueFrom:
        kind: GcpCloudRun
        name: orders-api
        fieldPath: status.outputs.service_name
```

The InfraPipeline resolves the dependency graph, deploys the Cloud Run service first, then the NEG — and a downstream GcpBackendService references this NEG's `self_link` in its `backends[].group`.

## Key Configuration

These are the most important decisions when configuring a regional NEG. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Endpoint type** -- `SERVERLESS` (the default when unset) fronts Cloud Run/Functions/App Engine; `PRIVATE_SERVICE_CONNECT` fronts a PSC endpoint or Google API bundle; `INTERNET_IP_PORT` / `INTERNET_FQDN_PORT` front external origins; `GCE_VM_IP_PORTMAP` does PSC port mapping. Immutable — the type decides which other fields apply.

**Serverless target** -- Exactly one of `cloudRun`, `cloudFunction`, or `appEngine` on a serverless NEG. Set the workload by reference (or name), or a `urlMask` that parses the target out of each request URL — host/path fan-out to many workloads from ONE NEG. A Cloud Run `tag` routes to one tagged revision (LB-level canaries). An empty `appEngine` block routes to the default application.

**PSC / internet wiring** -- A PSC NEG requires `pscTargetService` (the producer's service-attachment URL or a Google API bundle) plus the VPC network/subnetwork the endpoint lives in; internet NEGs name their origin in the same field.

**Immutability** -- EVERY field is ForceNew: any change destroys and recreates the NEG, and GCP refuses to delete a NEG a backend service still uses. Reshape with create-before-destroy: create the replacement, repoint the backend service, delete the original.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** (PSC/internet/portmap NEGs) | `network` | `status.outputs.network_self_link` |
| **GcpSubnetwork** (PSC/portmap NEGs) | `subnetwork` | `status.outputs.subnetwork_self_link` |
| **GcpCloudRun** (serverless NEGs) | `cloudRun.service` | `status.outputs.service_name` |
| **GcpCloudFunction** (serverless NEGs) | `cloudFunction.function` | `status.outputs.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `self_link` | Self-link URI of the NEG | GcpBackendService `backends[].group` — the composition handle that puts this NEG behind a load balancer |

The remaining outputs (`network_endpoint_group_name`, `network_endpoint_type`, `region`) echo the resolved inputs back for tooling that traces the serving chain; nothing downstream composes on them.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Cloud Run NEG** -- The dominant shape: front one Cloud Run service for a backend service. Start from the **Cloud Run behind a Load Balancer** preset.

**Private Service Connect NEG** -- Put a load balancer in front of a producer's published service or a Google API. Start from the **Private Service Connect Backend** preset.

**Internet FQDN NEG** -- Front an on-prem or third-party origin by hostname behind Google's edge. Start from the **External Internet Origin** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the NEG is created
- [**GCP Cloud Run**](/infra-catalog/gcp-cloud-run) -- the service a serverless NEG fronts
- [**GCP Cloud Function**](/infra-catalog/gcp-cloud-function) -- the function a serverless NEG fronts
- [**GCP Backend Service**](/infra-catalog/gcp-backend-service) -- consumes the NEG's `self_link` in its `backends[].group`
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- the network for PSC, internet, and portmap NEGs
- [**GCP Subnetwork**](/infra-catalog/gcp-subnetwork) -- the subnetwork for PSC and portmap NEGs
