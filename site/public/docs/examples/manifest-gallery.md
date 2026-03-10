---
title: "Manifest Gallery"
description: "Curated, proto-verified manifest examples organized by cloud provider for deploying infrastructure with OpenMCF"
icon: "package"
order: 20
---

# Manifest Gallery

Copy-paste-ready manifests for deploying infrastructure with OpenMCF. Every example on this page has been verified against the component's Protocol Buffer schema — field names, nesting, value types, and enum values are accurate.

## How to Use These Manifests

1. Copy the YAML into a file (e.g., `my-resource.yaml`)
2. Replace placeholder values (`my-org`, `my-project`, subnet IDs, credentials, etc.) with your actual values
3. Adjust the provisioner labels for your IaC engine and state backend
4. Run `openmcf plan -f my-resource.yaml` to preview, then `openmcf apply -f my-resource.yaml` to deploy

## Metadata Pattern

Every OpenMCF manifest follows the Kubernetes Resource Model. The `metadata` block is the same structure across all components:

```yaml
metadata:
  name: my-resource-name
  labels:
    # Choose your provisioner: "pulumi" or "tofu"
    openmcf.org/provisioner: pulumi
    # Pulumi state backend labels
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsS3Bucket.my-resource-name
```

For OpenTofu, replace the provisioner and backend labels:

```yaml
metadata:
  name: my-resource-name
  labels:
    openmcf.org/provisioner: tofu
    # OpenTofu state backend labels (example: S3 backend)
    tf.openmcf.org/backend.type: s3
    tf.openmcf.org/backend.s3.bucket: my-tf-state-bucket
    tf.openmcf.org/backend.s3.region: us-east-1
```

See [State Management](/docs/concepts/state-management) for all backend options.

The examples below use Pulumi labels. Swap the metadata labels to use OpenTofu instead — the `spec` section is identical regardless of provisioner.

---

## AWS

### S3 Bucket

Object storage with versioning, encryption, and lifecycle management.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: my-app-assets
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsS3Bucket.my-app-assets
spec:
  awsRegion: us-east-1
  isPublic: false
  versioningEnabled: true
  encryptionType: ENCRYPTION_TYPE_SSE_S3
  tags:
    environment: production
    team: platform
  lifecycleRules:
    - id: archive-old-objects
      enabled: true
      prefix: "logs/"
      transitionDays: 90
      transitionStorageClass: STORAGE_CLASS_GLACIER_FLEXIBLE_RETRIEVAL
      expirationDays: 365
```

[Full field reference: AwsS3Bucket](/docs/catalog/aws/awss3bucket)

### RDS Instance

Managed relational database (PostgreSQL, MySQL, MariaDB, Oracle, SQL Server).

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsInstance
metadata:
  name: my-app-database
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsRdsInstance.my-app-database
spec:
  subnetIds:
    - value: subnet-abc123    # Private subnet in AZ-a
    - value: subnet-def456    # Private subnet in AZ-b
  securityGroupIds:
    - value: sg-xyz789
  engine: postgres
  engineVersion: "15.4"
  instanceClass: db.t3.micro
  allocatedStorageGb: 20
  storageEncrypted: true
  username: postgres
  password: replace-with-secure-password
  port: 5432
  publiclyAccessible: false
  multiAz: false
```

[Full field reference: AwsRdsInstance](/docs/catalog/aws/awsrdsinstance)

### VPC

Virtual Private Cloud with subnets across availability zones.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: my-network
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsVpc.my-network
spec:
  vpcCidr: "10.0.0.0/16"
  availabilityZones:
    - us-west-2a
    - us-west-2b
  subnetsPerAvailabilityZone: 1
  subnetSize: 1
  isNatGatewayEnabled: true
  isDnsHostnamesEnabled: true
  isDnsSupportEnabled: true
```

[Full field reference: AwsVpc](/docs/catalog/aws/awsvpc)

---

## GCP

### Cloud SQL

Managed relational database (PostgreSQL or MySQL) on Google Cloud.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSql
metadata:
  name: my-gcp-database
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCloudSql.my-gcp-database
spec:
  projectId:
    value: my-gcp-project-id
  region: us-central1
  databaseEngine: POSTGRESQL
  databaseVersion: POSTGRES_15
  tier: db-n1-standard-1
  storageGb: 10
  rootPassword: replace-with-secure-password
  backup:
    enabled: true
    startTime: "03:00"
    retentionDays: 7
```

[Full field reference: GcpCloudSql](/docs/catalog/gcp/gcpcloudsql)

### GKE Cluster

Google Kubernetes Engine cluster with private nodes and VPC-native networking.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGkeCluster
metadata:
  name: my-gke-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGkeCluster.my-gke-cluster
spec:
  projectId:
    value: my-gcp-project-id
  networkSelfLink:
    value: projects/my-gcp-project-id/global/networks/my-vpc
  location: us-central1
  subnetworkSelfLink:
    value: projects/my-gcp-project-id/regions/us-central1/subnetworks/my-subnet
  clusterSecondaryRangeName:
    value: pods-range
  servicesSecondaryRangeName:
    value: services-range
  masterIpv4CidrBlock: "172.16.0.0/28"
  routerNatName:
    value: my-cloud-nat
  clusterName: my-gke-cluster
```

[Full field reference: GcpGkeCluster](/docs/catalog/gcp/gcpgkecluster)

---

## Azure

### AKS Cluster

Azure Kubernetes Service cluster with system and user node pools.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureAksCluster
metadata:
  name: my-aks-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureAksCluster.my-aks-cluster
spec:
  region: eastus
  resourceGroup:
    value: my-resource-group
  vnetSubnetId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/aks-subnet
  kubernetesVersion: "1.30"
  systemNodePool:
    vmSize: Standard_D4s_v5
    autoscaling:
      minCount: 3
      maxCount: 5
    availabilityZones:
      - "1"
      - "2"
      - "3"
  userNodePools:
    - name: general
      vmSize: Standard_D8s_v5
      autoscaling:
        minCount: 2
        maxCount: 10
      availabilityZones:
        - "1"
        - "2"
        - "3"
      spotEnabled: false
```

[Full field reference: AzureAksCluster](/docs/catalog/azure/azureakscluster)

---

## Kubernetes

### Deployment

Microservice deployment with container configuration, ingress, and autoscaling.

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDeployment
metadata:
  name: my-api-service
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesDeployment.my-api-service
spec:
  namespace:
    value: my-app-namespace
  createNamespace: true
  version: main
  container:
    app:
      image:
        repo: nginx
        tag: latest
      resources:
        requests:
          cpu: 50m
          memory: 100Mi
        limits:
          cpu: 1000m
          memory: 1Gi
      env:
        variables:
          LOG_LEVEL:
            value: info
          DATABASE_PORT:
            value: "5432"
      ports:
        - name: http
          containerPort: 80
          networkProtocol: TCP
          appProtocol: http
          servicePort: 80
          isIngressPort: true
  ingress:
    enabled: true
    hostname: my-api.example.com
  availability:
    minReplicas: 2
```

[Full field reference: KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment)

### PostgreSQL

PostgreSQL database on Kubernetes with custom databases, users, and resource tuning.

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: my-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesPostgres.my-postgres
spec:
  namespace:
    value: my-postgres-namespace
  createNamespace: true
  container:
    replicas: 2
    resources:
      requests:
        cpu: 250m
        memory: 512Mi
      limits:
        cpu: 2000m
        memory: 4Gi
    diskSize: 10Gi
  databases:
    - name: app_database
      ownerRole: app_user
  users:
    - name: app_user
      flags:
        - login
  ingress:
    enabled: false
```

[Full field reference: KubernetesPostgres](/docs/catalog/kubernetes/kubernetespostgres)

---

## Cloudflare

### Worker

Serverless function deployed to Cloudflare's edge network.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: my-edge-function
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareWorker.my-edge-function
spec:
  accountId: a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4
  workerName: my-edge-function
  scriptBundle:
    bucket: my-workers-bucket
    path: builds/my-edge-function-v1.0.0.js
  compatibilityDate: "2025-01-15"
  dns:
    enabled: true
    zoneId: z1y2x3w4v5u6z1y2x3w4v5u6z1y2x3w4
    hostname: api.example.com
```

[Full field reference: CloudflareWorker](/docs/catalog/cloudflare/cloudflareworker)

---

## Civo

### VPC

Private network on the Civo cloud platform.

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoVpc
metadata:
  name: my-civo-network
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CivoVpc.my-civo-network
spec:
  civoCredentialId: your-civo-credential-id
  networkName: my-civo-network
  region: LON1
  ipRangeCidr: "10.10.1.0/24"
  description: Production network
```

[Full field reference: CivoVpc](/docs/catalog/civo/civovpc)

---

## What's Next

- **[Tutorials](/docs/tutorials)** — Guided walkthroughs for deploying resources step-by-step
- **[Catalog](/docs/catalog)** — Full field reference for all 360+ deployment components
- **[CLI Reference](/docs/cli/cli-reference)** — Complete command and flag reference
- **[Guides](/docs/guides)** — How-to guides for credentials, state backends, Kustomize, and more
