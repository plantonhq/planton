# GCP Cloud Run

Deploys a Google Cloud Run v2 service with configurable container resources, autoscaling, VPC egress, and optional custom domain mapping with DNS verification. The component supports environment variables, Secret Manager references, and configurable ingress controls.

## What Gets Created

When you deploy a GcpCloudRun resource, OpenMCF provisions:

- **Cloud Run v2 Service** — a `google_cloud_run_v2_service` with the specified container image, resource limits, scaling configuration, and ingress settings
- **Domain Mapping** — created only when DNS is enabled, binds the first hostname in `dns.hostnames` to the Cloud Run service
- **DNS TXT Record** — created only when DNS is enabled, a Cloud DNS record set in the specified managed zone for Google's domain ownership verification

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **A GCP project** with the Cloud Run API enabled
- **A container image** pushed to a registry accessible from the project (e.g., Artifact Registry, Container Registry)
- **A Cloud DNS managed zone** if enabling custom domain mapping
- **A VPC network and subnet** if configuring Direct VPC Egress

## Quick Start

Create a file `cloudrun.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudRun
metadata:
  name: my-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpCloudRun.my-api
spec:
  projectId: my-gcp-project
  region: us-central1
  container:
    image:
      repo: us-docker.pkg.dev/my-gcp-project/registry/my-api
      tag: "1.0.0"
    cpu: 1
    memory: 512
    replicas:
      min: 0
      max: 3
```

Deploy:

```shell
openmcf apply -f cloudrun.yaml
```

This creates a publicly accessible Cloud Run service with 1 vCPU, 512 MiB memory, scaling from 0 to 3 instances, serving on port 8080.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `string` | GCP project ID where the service is created. Can reference a GcpProject resource via `valueFrom`. | Required |
| `region` | `string` | GCP region for the service (e.g., `us-central1`). | Pattern: `^[a-z]+-[a-z]+[0-9]$` |
| `container.image.repo` | `string` | Container image repository (e.g., `us-docker.pkg.dev/prj/registry/app`). | Minimum length: 1 |
| `container.image.tag` | `string` | Container image tag (e.g., `1.0.0`). | Minimum length: 1 |
| `container.cpu` | `int32` | vCPU units per instance. | Allowed values: `1`, `2`, `4`. Recommended default: `1` |
| `container.memory` | `int32` | Memory in MiB per instance. | 128–32768. Recommended default: `512` |
| `container.replicas` | `object` | Scaling bounds for the service. | Required |
| `container.replicas.min` | `int32` | Minimum warm instances. Set to `0` for scale-to-zero. | >= 0. Recommended default: `0` |
| `container.replicas.max` | `int32` | Maximum instances Cloud Run may scale to. | >= 0 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `serviceName` | `string` | `metadata.name` | Cloud Run service name on GCP. Must be lowercase alphanumeric with hyphens, max 63 characters. |
| `serviceAccount` | `string` | Default Compute Engine SA | Service account email the service runs as. |
| `container.port` | `int32` | `8080` | Container port that receives HTTP traffic. Range: 1–65535. |
| `container.env.variables` | `map<string, string>` | `{}` | Plain environment variables injected as KEY=VALUE pairs. |
| `container.env.secrets` | `map<string, string>` | `{}` | Secret Manager references injected as KEY=`projects/*/secrets/*:version`. |
| `maxConcurrency` | `int32` | `80` | Maximum concurrent requests handled by one instance. Range: 1–1000. |
| `timeoutSeconds` | `int32` | `300` | Request timeout in seconds. Range: 1–3600. |
| `ingress` | `enum` | `INGRESS_TRAFFIC_ALL` | Ingress setting. `INGRESS_TRAFFIC_ALL`: public internet. `INGRESS_TRAFFIC_INTERNAL_ONLY`: internal only. `INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER`: internal + Cloud Load Balancing. |
| `allowUnauthenticated` | `bool` | `true` | When `true`, the service is publicly invokable without IAM authentication. |
| `executionEnvironment` | `enum` | `EXECUTION_ENVIRONMENT_GEN2` | Execution environment. `EXECUTION_ENVIRONMENT_GEN1`: first generation. `EXECUTION_ENVIRONMENT_GEN2`: full Linux compatibility, slower cold starts. |
| `deleteProtection` | `bool` | `false` | Prevents accidental deletion of the service when enabled. |
| `vpcAccess.network` | `string` | — | VPC network name for Direct VPC Egress. Can reference a GcpVpc resource via `valueFrom`. |
| `vpcAccess.subnet` | `string` | — | Subnet name for Direct VPC Egress. Can reference a GcpSubnetwork resource via `valueFrom`. |
| `vpcAccess.egress` | `string` | — | Egress routing: `ALL_TRAFFIC` routes all egress through VPC, `PRIVATE_RANGES_ONLY` routes only private IP traffic. |
| `dns.enabled` | `bool` | `false` | Enables custom domain mapping for the service. |
| `dns.hostnames` | `string[]` | `[]` | Fully-qualified hostnames routed to the service. Must be unique. Required when `dns.enabled` is `true`. |
| `dns.managedZone` | `string` | — | Cloud DNS managed zone for domain verification records. Required when `dns.enabled` is `true`. Can reference a GcpDnsZone resource via `valueFrom`. |

## Examples

### Internal Microservice

A service accessible only from within GCP, with scale-to-zero disabled for low latency:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudRun
metadata:
  name: order-svc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpCloudRun.order-svc
spec:
  projectId: my-gcp-project
  region: us-central1
  container:
    image:
      repo: us-docker.pkg.dev/my-gcp-project/registry/order-svc
      tag: "2.1.0"
    cpu: 1
    memory: 256
    port: 3000
    replicas:
      min: 1
      max: 5
  ingress: INGRESS_TRAFFIC_INTERNAL_ONLY
  allowUnauthenticated: false
```

### Service with Environment Variables and Secrets

A service that connects to a database using environment variables and Secret Manager references:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudRun
metadata:
  name: web-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.GcpCloudRun.web-app
spec:
  projectId: my-gcp-project
  region: us-east1
  container:
    image:
      repo: us-docker.pkg.dev/my-gcp-project/registry/web-app
      tag: "3.0.0"
    cpu: 2
    memory: 1024
    replicas:
      min: 1
      max: 10
    env:
      variables:
        NODE_ENV: production
        LOG_LEVEL: info
      secrets:
        DATABASE_URL: projects/my-gcp-project/secrets/db-url:latest
        API_KEY: projects/my-gcp-project/secrets/api-key:latest
  maxConcurrency: 100
  timeoutSeconds: 60
```

### Production Service with VPC Egress and Custom Domain

Full-featured production deployment with VPC connectivity, custom DNS, and deletion protection:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudRun
metadata:
  name: prod-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCloudRun.prod-api
spec:
  projectId: my-gcp-project
  region: us-central1
  serviceName: prod-api
  serviceAccount: prod-api@my-gcp-project.iam.gserviceaccount.com
  container:
    image:
      repo: us-docker.pkg.dev/my-gcp-project/registry/prod-api
      tag: "5.2.1"
    cpu: 4
    memory: 4096
    port: 8080
    replicas:
      min: 2
      max: 20
    env:
      variables:
        ENVIRONMENT: production
      secrets:
        DB_PASSWORD: projects/my-gcp-project/secrets/db-password:latest
  maxConcurrency: 200
  timeoutSeconds: 120
  executionEnvironment: EXECUTION_ENVIRONMENT_GEN2
  deleteProtection: true
  vpcAccess:
    network: my-vpc
    subnet: my-subnet
    egress: PRIVATE_RANGES_ONLY
  dns:
    enabled: true
    hostnames:
      - api.example.com
    managedZone: example-com-zone
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding values:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudRun
metadata:
  name: ref-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCloudRun.ref-api
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      field: status.outputs.project_id
  region: us-central1
  container:
    image:
      repo: us-docker.pkg.dev/my-gcp-project/registry/ref-api
      tag: "1.0.0"
    cpu: 1
    memory: 512
    replicas:
      min: 0
      max: 5
  vpcAccess:
    network:
      valueFrom:
        kind: GcpVpc
        name: my-vpc
        field: status.outputs.network_name
    subnet:
      valueFrom:
        kind: GcpSubnetwork
        name: my-subnet
        field: status.outputs.subnetwork_name
    egress: PRIVATE_RANGES_ONLY
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `url` | `string` | Public or internal URL of the Cloud Run service (e.g., `https://my-api-abc123-uc.a.run.app`) |
| `service_name` | `string` | Name of the Cloud Run service as it appears in GCP |
| `revision` | `string` | Name of the latest ready revision deployed |

## Related Components

- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project where the service is created
- [GcpVpc](/docs/catalog/gcp/gcpvpc) — provides the VPC network for Direct VPC Egress
- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — provides the subnet for Direct VPC Egress
- [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) — commonly co-deployed as the database backend accessed via VPC
