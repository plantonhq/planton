---
title: "Writing Manifests"
description: "Practical guide to writing, validating, and managing OpenMCF manifests for infrastructure deployments"
icon: "book"
order: 20
---

# Writing Manifests

This guide walks you through writing OpenMCF manifests from scratch. For the conceptual foundation — KRM structure, field semantics, label conventions, and manifest sources — see [Manifests](../concepts/manifests).

## Finding the Right Kind and Spec Fields

Every manifest starts with two decisions: which `kind` to use and what to put in `spec`. Here is how to find both.

### Step 1: Identify the Kind

Browse the [Deployment Component Catalog](/docs/catalog) to find the component that matches what you want to deploy. Each catalog entry shows the `kind` name, provider, and a description of what the component deploys.

If you know the provider and resource type, the kind name follows a predictable pattern:

| Provider | Pattern | Examples |
|----------|---------|----------|
| AWS | `Aws` + resource | `AwsRdsInstance`, `AwsS3Bucket`, `AwsEksCluster` |
| GCP | `Gcp` + resource | `GcpCloudSql`, `GcpGkeCluster`, `GcpGcsBucket` |
| Azure | `Azure` + resource | `AzureAksCluster`, `AzureResourceGroup` |
| Kubernetes | `Kubernetes` + resource | `KubernetesPostgres`, `KubernetesRedis` |

For the complete taxonomy, see [Cloud Resource Kinds](../concepts/cloud-resource-kinds).

### Step 2: Find the Spec Fields

Once you know the kind, find the available spec fields using one of these methods:

**Catalog page** — Each component's catalog page lists the spec fields with descriptions and types. Start here.

**Protocol Buffer definition** — The canonical source of truth. Every component's spec is defined at:

```
apis/org/openmcf/provider/{provider}/{component}/v1/spec.proto
```

The proto file shows every field, its type, whether it is required, and any validation rules.

**Buf Schema Registry** — Browse the API definitions at [buf.build/openmcf/openmcf](https://buf.build/openmcf/openmcf) for generated documentation of every component's spec.

## Writing a Manifest Step by Step

### 1. Start with the Envelope

Every manifest has the same outer structure. Fill in `apiVersion` and `kind` from the catalog:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsInstance
metadata:
  name: my-database
spec:
  # Fields go here
```

The `apiVersion` follows the pattern `{provider}.openmcf.org/v1`. The `name` in metadata must be lowercase alphanumeric with hyphens, 63 characters or fewer.

### 2. Add Required Spec Fields

Check the component's spec definition for required fields. For `AwsRdsInstance`, the spec includes engine configuration, instance sizing, and network settings:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsInstance
metadata:
  name: my-database
spec:
  subnetIds:
    - value: subnet-abc123
    - value: subnet-def456
  securityGroupIds:
    - value: sg-xyz789
  engine: postgres
  engineVersion: "15.4"
  instanceClass: db.t3.micro
  allocatedStorageGb: 20
  username: postgres
  password: my-secure-password
  port: 5432
```

Field names in YAML use `camelCase` and match the proto field names exactly. If the proto defines `allocated_storage_gb`, the YAML field is `allocatedStorageGb`.

### 3. Add Provisioner and State Labels

Labels tell OpenMCF which IaC engine to use and where to store state:

```yaml
metadata:
  name: my-database
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/stack.name: prod.AwsRdsInstance.my-database
```

For OpenTofu or Terraform, use backend labels instead:

```yaml
metadata:
  name: my-database
  labels:
    openmcf.org/provisioner: tofu
    tofu.openmcf.org/backend.type: s3
    tofu.openmcf.org/backend.bucket: my-state-bucket
    tofu.openmcf.org/backend.key: rds/my-database.tfstate
    tofu.openmcf.org/backend.region: us-west-2
```

For details on provisioner selection, see [Dual IaC Engines](../concepts/dual-iac-engines). For backend configuration, see [State Backends](./state-backends).

### 4. Validate Before Deploying

Run validation to catch errors before any cloud API call:

```bash
openmcf validate -f my-database.yaml
```

Validation checks the manifest against the proto schema, enforces field constraints (type, range, pattern), and verifies required fields are present. If everything passes, the command exits silently with code 0. If there are errors, you get specific field-level messages.

### 5. Inspect Defaults

Many components define default values for optional fields. Use `load-manifest` to see the effective manifest with all defaults applied:

```bash
openmcf load-manifest -f my-database.yaml
```

The output shows the complete manifest as OpenMCF will interpret it, including any default values filled in for fields you omitted.

## Real-World Examples

These examples are from the OpenMCF repository and follow the actual proto schemas.

### PostgreSQL on Kubernetes

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: app-database
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/stack.name: prod.KubernetesPostgres.app-database
spec:
  namespace:
    value: app-database
  container:
    replicas: 1
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 2000m
        memory: 2Gi
    diskSize: 1Gi
  ingress:
    enabled: false
```

### PostgreSQL on GCP Cloud SQL

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSql
metadata:
  name: analytics-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/stack.name: prod.GcpCloudSql.analytics-db
spec:
  databaseEngine: POSTGRESQL
  databaseVersion: POSTGRES_15
  projectId: my-gcp-project
  region: us-central1
  tier: db-f1-micro
  storageGb: 10
  rootPassword: my-secure-password
  network:
    authorizedNetworks:
      - 10.0.0.0/8
```

### PostgreSQL on AWS RDS

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsInstance
metadata:
  name: orders-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/stack.name: prod.AwsRdsInstance.orders-db
spec:
  subnetIds:
    - value: subnet-abc123
    - value: subnet-def456
  securityGroupIds:
    - value: sg-xyz789
  engine: postgres
  engineVersion: "15.4"
  instanceClass: db.t3.micro
  allocatedStorageGb: 20
  storageEncrypted: true
  username: postgres
  password: my-secure-password
  port: 5432
  publiclyAccessible: false
  multiAz: false
```

## Common Patterns

### Minimal vs. Explicit

Components with defaults let you write concise manifests. You only need to specify fields where the default does not match your needs:

```yaml
# Minimal — relies on component defaults
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: dev-database
spec:
  container:
    replicas: 1
    diskSize: 1Gi
```

For production, be explicit about resource limits, replicas, and security settings:

```yaml
# Explicit — production configuration
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: prod-database
spec:
  container:
    replicas: 3
    resources:
      requests:
        cpu: 1000m
        memory: 2Gi
      limits:
        cpu: 2000m
        memory: 4Gi
    diskSize: 100Gi
```

### One Resource Per File

Keep each manifest in its own file. This makes it straightforward to deploy, version, and manage resources independently:

```
ops/
  database.yaml
  cache.yaml
  storage-bucket.yaml
```

Deploy each resource separately:

```bash
openmcf pulumi up -f ops/database.yaml
openmcf pulumi up -f ops/cache.yaml
```

### Runtime Overrides

Use `--set` to override individual fields without editing the file. This is useful for CI/CD pipelines, testing, and temporary changes:

```bash
openmcf pulumi up -f ops/api.yaml \
  --set spec.container.replicas=5
```

Overrides apply after all other manifest resolution (file, Kustomize overlay). For details, see [Advanced Usage](./advanced-usage).

## Best Practices

**Use version control.** Track manifests in Git alongside application code. This gives you change history, code review, and rollback capability.

**Validate before deploying.** Run `openmcf validate -f manifest.yaml` before every deployment. Validation catches field-level errors in seconds that would otherwise surface minutes into a cloud API call.

**Use meaningful names.** The `metadata.name` appears in state files, cloud resources, and logs. Make it descriptive: `prod-api-postgres` instead of `db1`.

**Keep secrets out of manifests.** Do not put passwords, API keys, or tokens directly in manifest files that will be committed to Git. Use environment variables, secret managers, or credential references instead.

**Document non-obvious choices.** Add YAML comments explaining why you chose specific values — instance sizes, region selections, non-default settings. The next person reading the manifest will thank you.

## What's Next

- [Credentials](./credentials) — Set up cloud provider authentication
- [Kustomize Integration](./kustomize) — Manage multi-environment deployments
- [State Backends](./state-backends) — Configure where deployment state is stored
- [CLI Reference](/docs/cli/cli-reference) — Complete command and flag reference
