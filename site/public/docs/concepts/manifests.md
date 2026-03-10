---
title: "Manifests"
description: "How OpenMCF uses the Kubernetes Resource Model (KRM) to declare infrastructure as structured YAML manifests with typed fields, validation, and provider-specific configuration"
icon: "book"
order: 25
---

# Manifests

Every interaction with OpenMCF starts with a manifest -- a YAML file that declares what you want to deploy. OpenMCF manifests follow the Kubernetes Resource Model (KRM), which means if you have written a Kubernetes YAML file before, the structure will be immediately familiar: `apiVersion`, `kind`, `metadata`, `spec`.

The difference is what backs the manifest. In Kubernetes, the schema is defined by Go structs and validated by the API server at admission time. In OpenMCF, the schema is defined by Protocol Buffer definitions with field-level validation rules that are enforced before any cloud API is ever called.

## Manifest Structure

Every OpenMCF manifest has five top-level fields:

```yaml
apiVersion: <provider>.openmcf.org/v1
kind: <ComponentKind>
metadata:
  name: <resource-name>
  org: <organization>
  env: <environment>
  labels: { ... }
spec:
  # Provider-specific configuration
status:
  # Read-only, populated after deployment
```

### apiVersion

The `apiVersion` identifies which provider and API version this manifest targets. It follows the pattern `{provider}.openmcf.org/v1`:

| Provider | apiVersion |
|----------|-----------|
| AWS | `aws.openmcf.org/v1` |
| GCP | `gcp.openmcf.org/v1` |
| Azure | `azure.openmcf.org/v1` |
| Kubernetes | `kubernetes.openmcf.org/v1` |
| DigitalOcean | `digital-ocean.openmcf.org/v1` |
| Civo | `civo.openmcf.org/v1` |
| Cloudflare | `cloudflare.openmcf.org/v1` |
| OpenStack | `openstack.openmcf.org/v1` |

The `apiVersion` value is enforced as a constant in the component's Protocol Buffer definition. If you set the wrong `apiVersion`, validation fails immediately -- not after a network call to a cloud provider.

### kind

The `kind` identifies which deployment component this manifest represents. It is the exact name from the `CloudResourceKind` enum -- for example, `KubernetesPostgres`, `AwsS3Bucket`, `GcpCloudSql`, or `AwsRdsInstance`.

Like `apiVersion`, the `kind` value is enforced as a constant in the protobuf definition:

```protobuf
string kind = 2 [(buf.validate.field).string.const = 'KubernetesPostgres'];
```

The combination of `apiVersion` and `kind` uniquely identifies the deployment component. The CLI uses these two fields to resolve the correct IaC module, load the right validation rules, and construct the stack input.

### metadata

The `metadata` section identifies the resource and carries organizational context. It is defined by the shared `CloudResourceMetadata` message, which is the same across all 362 components.

```yaml
metadata:
  name: session-store
  org: acme
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: production.KubernetesPostgres.session-store
  tags:
    - database
    - backend
```

The metadata fields:

| Field | Purpose |
|-------|---------|
| `name` | Resource name. Used for identification and often as the basis for cloud resource naming. |
| `org` | Organization identifier. Groups resources by team or business unit. |
| `env` | Environment identifier (e.g., `production`, `staging`, `dev`). |
| `labels` | Key-value pairs. Some labels have special meaning to the CLI (see below). |
| `annotations` | Key-value pairs for arbitrary metadata. Not used by the CLI. |
| `tags` | String list for categorization and filtering. |
| `relationships` | Explicit dependencies between this resource and other resources. |
| `group` | Visual grouping identifier (e.g., `app/services`, `infrastructure/networking`). |

### Labels with Special Meaning

Certain label keys are read by the CLI to configure IaC engine behavior. These labels control which provisioner runs, how state is stored, and where the state backend lives.

**Provisioner selection:**

| Label | Values | Purpose |
|-------|--------|---------|
| `openmcf.org/provisioner` | `pulumi`, `tofu`, `terraform` | Tells unified commands (`apply`, `plan`, `destroy`) which engine to use |

**Pulumi state configuration:**

| Label | Example | Purpose |
|-------|---------|---------|
| `pulumi.openmcf.org/organization` | `acme` | Pulumi Cloud organization or backend namespace |
| `pulumi.openmcf.org/project` | `platform` | Pulumi project name |
| `pulumi.openmcf.org/stack.name` | `prod.KubernetesPostgres.db` | Pulumi stack name |

**OpenTofu/Terraform state configuration:**

| Label | Example | Purpose |
|-------|---------|---------|
| `tofu.openmcf.org/backend.type` | `s3`, `gcs`, `azurerm`, `local` | State backend type |
| `tofu.openmcf.org/backend.bucket` | `my-tfstate-bucket` | Bucket name for remote state |
| `tofu.openmcf.org/backend.key` | `prod/postgres/terraform.tfstate` | State file path within the bucket |
| `tofu.openmcf.org/backend.region` | `us-east-1` | AWS region for S3 backend |
| `tofu.openmcf.org/backend.endpoint` | `https://acct.r2.cloudflarestorage.com` | Custom endpoint for S3-compatible backends (R2, MinIO) |

See [State Management](state-management) for the full list of backend labels and their validation rules.

### spec

The `spec` section contains the provider-specific configuration for the resource. Every field in the spec is defined by the component's `spec.proto` file, with types, validation rules, and documentation.

There is no universal spec structure -- each component defines what it needs. A `KubernetesPostgres` spec has container replicas, disk size, and ingress configuration. An `AwsS3Bucket` spec has region, encryption type, lifecycle rules, and CORS settings. An `AwsRdsInstance` spec has subnet IDs, engine version, instance class, and storage.

This is the provider-specific design in action. The spec for each component exposes the full native capability of its target platform.

### status

The `status` section is read-only. It is populated by the IaC module after deployment and contains the stack outputs -- connection strings, endpoints, ARNs, secret references, and other values produced by the deployment.

You do not set `status` in your manifest. It exists so that after deployment, the full resource state (what you asked for in `spec` plus what was produced in `status`) is captured in a single document.

## Real Manifest Examples

### PostgreSQL on Kubernetes

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: kubernetes-postgres-example
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: organization
    pulumi.openmcf.org/project: openmcf-examples
    pulumi.openmcf.org/stack.name: example-env.KubernetesPostgres.kubernetes-postgres-example
spec:
  namespace:
    value: kubernetes-postgres-example
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

### PostgreSQL on AWS RDS

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsInstance
metadata:
  name: aws-postgres-example
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: organization
    pulumi.openmcf.org/project: openmcf-examples
    pulumi.openmcf.org/stack.name: example-env.AwsRdsInstance.aws-postgres-example
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

### PostgreSQL on GCP Cloud SQL

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSql
metadata:
  name: gcp-postgres-example
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: organization
    pulumi.openmcf.org/project: openmcf-examples
    pulumi.openmcf.org/stack.name: example-env.GcpCloudSql.gcp-postgres-example
spec:
  databaseEngine: POSTGRESQL
  databaseVersion: POSTGRES_15
  network:
    authorizedNetworks:
    - 0.0.0.0/0
  projectId: openmcf-demo
  region: asia-south1
  rootPassword: my-secure-password
  storageGb: 10
  tier: db-f1-micro
```

Notice the pattern. The envelope is identical -- same `apiVersion` structure, same `kind` field, same `metadata` with the same label keys. The `spec` is entirely different because each platform requires different configuration. This is the consistency that KRM provides without sacrificing provider-specific capability.

## Manifest Sources

The CLI supports multiple ways to provide a manifest. These are controlled by flags on any command that accepts a manifest.

| Flag | Short | Description |
|------|-------|-------------|
| `--manifest` | `-f` | Path to a manifest YAML file |
| `--clipboard` | `-c` | Read manifest from system clipboard |
| `--stack-input` | `-i` | Path to a stack input YAML file (extracts manifest from `target` field) |
| `--input-dir` | -- | Directory containing `target.yaml` and credential files |
| `--kustomize-dir` | -- | Directory containing Kustomize configuration |
| `--overlay` | -- | Kustomize overlay to apply (used with `--kustomize-dir`) |

When multiple sources are provided, the CLI follows a priority order: `--clipboard` > `--stack-input` > `--manifest` > `--input-dir` > `--kustomize-dir` + `--overlay`.

The most common usage is `--manifest` / `-f`:

```bash
openmcf pulumi up -f postgres.yaml --stack my-org/my-project/production
```

The clipboard source is useful during development when iterating on a manifest:

```bash
# Copy manifest to clipboard, then:
openmcf validate --clipboard
```

The Kustomize source enables multi-environment workflows with overlays:

```bash
openmcf pulumi up --kustomize-dir ./k8s --overlay production
```

## Runtime Overrides

The `--set` flag allows overriding individual manifest values at execution time without modifying the YAML file. It accepts `key=value` pairs where the key is a dot-delimited path into the manifest:

```bash
openmcf pulumi up -f postgres.yaml \
  --set spec.container.replicas=3 \
  --set spec.container.diskSize=10Gi \
  --stack my-org/my-project/production
```

This is particularly useful in CI/CD pipelines where environment-specific values (replica counts, resource limits, regions) differ between deployments but the base manifest stays the same.

## What's Next

- **[Deployment Components](deployment-components)** -- The anatomy of what manifests target
- **[Cloud Resource Kinds](cloud-resource-kinds)** -- The full taxonomy of valid `kind` values
- **[Validation](validation)** -- How manifests are validated before deployment
- **[State Management](state-management)** -- How manifest labels configure state backends
