---
title: "Migrating to OpenMCF"
description: "How to migrate from raw Terraform or Pulumi to OpenMCF — what changes, what stays the same, and step-by-step migration paths"
icon: "rocket"
order: 110
---

# Migrating to OpenMCF

If you are already using Terraform, OpenTofu, or Pulumi to manage infrastructure, OpenMCF does not replace your IaC engine. It wraps it with a consistent manifest layer, proto-based validation, and a unified CLI. This guide explains what changes when you adopt OpenMCF, what stays the same, and how to migrate existing infrastructure.

## What Changes, What Stays the Same

### What Changes

| Aspect | Before (Raw IaC) | After (OpenMCF) |
|--------|-------------------|-----------------|
| **Input format** | HCL `.tf` files or Go/TypeScript/Python code | YAML manifests following the KRM structure |
| **Validation** | `terraform validate` or language type checks | Proto-based validation with field-level constraints before any cloud API call |
| **Module source** | Terraform Registry, npm, pip, Go modules | OpenMCF module system (staging, binary, or local) |
| **Resource definition** | One project per resource/module | One manifest file per resource |
| **Multi-cloud consistency** | Different project structure per provider | Same manifest structure across all providers |

### What Stays the Same

| Aspect | Details |
|--------|---------|
| **Cloud provider APIs** | OpenMCF creates the same cloud resources. An `AwsRdsInstance` manifest produces the same RDS instance as raw Terraform or Pulumi. |
| **Credentials** | The same environment variables (`AWS_ACCESS_KEY_ID`, `GOOGLE_APPLICATION_CREDENTIALS`, `ARM_CLIENT_ID`, etc.) work with OpenMCF. |
| **State backends** | Pulumi Cloud, S3, GCS, Azure Blob, and local backends all work. OpenMCF configures them from manifest labels. |
| **IaC engine** | You choose Pulumi or OpenTofu/Terraform per resource. The same engine runs underneath. |
| **Provider plugins** | The same Terraform providers and Pulumi provider SDKs are used by OpenMCF modules. |

## From Raw Terraform to OpenMCF

### Before: Terraform Project

A typical Terraform project for an RDS instance:

```hcl
# variables.tf
variable "instance_class" {
  type    = string
  default = "db.t3.micro"
}

variable "engine_version" {
  type    = string
  default = "15.4"
}

# main.tf
resource "aws_db_instance" "main" {
  identifier           = "my-database"
  engine               = "postgres"
  engine_version       = var.engine_version
  instance_class       = var.instance_class
  allocated_storage    = 20
  username             = "postgres"
  password             = var.db_password
  publicly_accessible  = false
  storage_encrypted    = true

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.db.id]
}

# backend.tf
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "rds/my-database.tfstate"
    region = "us-west-2"
  }
}
```

### After: OpenMCF Manifest

The equivalent OpenMCF manifest:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsInstance
metadata:
  name: my-database
  labels:
    openmcf.org/provisioner: terraform
    terraform.openmcf.org/backend.type: s3
    terraform.openmcf.org/backend.bucket: my-terraform-state
    terraform.openmcf.org/backend.key: rds/my-database.tfstate
    terraform.openmcf.org/backend.region: us-west-2
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
  publiclyAccessible: false
  storageEncrypted: true
```

### Key Differences

**No HCL files to write.** The YAML manifest replaces `variables.tf`, `main.tf`, and `backend.tf`. OpenMCF's pre-built Terraform modules handle the resource creation.

**Backend in labels, not in a separate file.** The S3 backend configuration moves into manifest labels. OpenMCF generates the Terraform backend block from these labels.

**Spec fields match the proto, not Terraform resource arguments.** The manifest uses proto-defined fields (`allocatedStorageGb`, `engineVersion`) rather than Terraform argument names (`allocated_storage`, `engine_version`). Check the component's catalog page or `spec.proto` for the exact field names.

**Validation happens before Terraform runs.** `openmcf validate -f manifest.yaml` catches field-level errors (wrong types, missing required fields, constraint violations) before Terraform is even invoked.

### Migration Steps

1. **Identify the OpenMCF component** that matches your Terraform resource. Browse the [Catalog](/docs/catalog).
2. **Write the manifest** using the component's spec fields. Map your Terraform variable values to the corresponding spec fields.
3. **Add provisioner and backend labels** to use the same state backend you are currently using.
4. **Validate** with `openmcf validate -f manifest.yaml`.
5. **Import existing state** if the resource already exists. Use `openmcf terraform init` to initialize the workspace, then use standard Terraform import commands if needed.
6. **Deploy** with `openmcf apply -f manifest.yaml` or `openmcf terraform apply -f manifest.yaml`.

## From Raw Pulumi to OpenMCF

### Before: Pulumi Project

A typical Pulumi Go project for a Kubernetes PostgreSQL deployment:

```go
// main.go
func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        ns, _ := corev1.NewNamespace(ctx, "postgres-ns", &corev1.NamespaceArgs{
            Metadata: &metav1.ObjectMetaArgs{
                Name: pulumi.String("postgres"),
            },
        })

        // ... PostgreSQL operator CRD, deployment, service ...
        return nil
    })
}
```

```yaml
# Pulumi.yaml
name: kubernetes-postgres
runtime: go
```

### After: OpenMCF Manifest

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

### Key Differences

**No Go code to write.** The YAML manifest replaces `main.go` and `Pulumi.yaml`. OpenMCF's pre-built Pulumi modules (written in Go) handle the resource creation.

**Stack name in labels.** The Pulumi stack identity moves into the `pulumi.openmcf.org/stack.name` label.

**Spec fields are higher-level.** Instead of constructing Kubernetes resources in code (Namespace, StatefulSet, Service, PVC), you declare what you want (`replicas`, `diskSize`, `resources`) and the module creates the full resource graph.

### Migration Steps

1. **Identify the OpenMCF component** that matches your Pulumi program. Browse the [Catalog](/docs/catalog).
2. **Write the manifest** using the component's spec fields.
3. **Add provisioner and stack name labels.** Use the same Pulumi organization and project if you want to preserve state continuity.
4. **Validate** with `openmcf validate -f manifest.yaml`.
5. **Deploy** with `openmcf pulumi up -f manifest.yaml`.

If the Pulumi stack already exists with the same stack name and the underlying module creates the same resources, Pulumi will detect the existing state and perform a no-op or minimal update.

## What If My Resource Has No OpenMCF Component?

OpenMCF currently supports 198 deployment components across 14 providers. If the specific resource you need is not available as an OpenMCF component:

- **Continue using raw Terraform or Pulumi** for that resource. OpenMCF does not require all-or-nothing adoption.
- **Check the catalog** periodically — new components are added regularly.
- **Contribute a component** if you want to add support. See the contributing guide for how to add new deployment components.

OpenMCF is designed for incremental adoption. You can start with one resource, prove the workflow, and expand at your own pace.

## What's Next

- [Writing Manifests](./manifests) — How to write and validate manifests
- [Credentials](./credentials) — Configure cloud provider authentication
- [State Backends](./state-backends) — Configure state storage
- [Deployment Components](../concepts/deployment-components) — Anatomy of an OpenMCF component
- [Dual IaC Engines](../concepts/dual-iac-engines) — How Pulumi and OpenTofu/Terraform coexist
