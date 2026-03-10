---
title: "Getting Started"
description: "Install OpenMCF and deploy your first resource to a local Kubernetes cluster"
icon: "rocket"
order: 2
---

# Getting Started

By the end of this page, you will have installed the OpenMCF CLI, deployed a PostgreSQL database to a local Kubernetes cluster, verified it is running, and torn it down. The entire process takes about 10 minutes.

<!-- AI-AGENT NOTE: The deployment component count below MUST be verified against the authoritative source:
     apis/org/openmcf/shared/cloudresourcekind/cloud_resource_kind.proto
     Count only non-test, non-unspecified enum values in the CloudResourceKind enum. -->

You will use KubernetesPostgres — one of OpenMCF's 360+ deployment components — as your first resource. It runs on a local Kind cluster, so you do not need cloud provider credentials or a paid account to get started.

## What You'll Need

Install these tools before proceeding:

| Tool | Purpose | Install |
|------|---------|---------|
| [Git](https://git-scm.com/) | Module resolution (clones IaC modules) | `brew install git` |
| [Kind](https://kind.sigs.k8s.io/) | Local Kubernetes cluster | `brew install kind` |
| [kubectl](https://kubernetes.io/docs/tasks/tools/) | Cluster verification | `brew install kubectl` |
| [Pulumi CLI](https://www.pulumi.com/docs/install/) | IaC engine (executes deployments) | `brew install pulumi` |

This guide uses Pulumi as the IaC engine. If you prefer OpenTofu or Terraform, see [Dual IaC Engines](/docs/concepts/dual-iac-engines) for setup instructions.

## Install OpenMCF

```bash
brew install plantonhq/tap/openmcf
```

Verify the installation:

```bash
openmcf version
```

## Create a Local Cluster

Create a Kubernetes cluster using Kind:

```bash
kind create cluster --name openmcf-quickstart
```

Confirm the cluster is running:

```bash
kubectl cluster-info --context kind-openmcf-quickstart
```

## Write Your Manifest

Create a file named `postgres.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: my-first-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: local
    pulumi.openmcf.org/project: getting-started
    pulumi.openmcf.org/stack.name: dev
spec:
  namespace:
    value: my-first-postgres
  createNamespace: true
  container:
    replicas: 1
    resources:
      requests:
        cpu: 50m
        memory: 100Mi
      limits:
        cpu: 500m
        memory: 512Mi
    diskSize: 1Gi
```

Every OpenMCF manifest follows the Kubernetes Resource Model (KRM) — the same `apiVersion`, `kind`, `metadata`, `spec` structure used by Kubernetes itself. Here is what each section does:

- **`apiVersion` and `kind`** identify this as a KubernetesPostgres resource. OpenMCF has [360+ component kinds](/docs/concepts/cloud-resource-kinds) across 17 cloud providers, each with its own apiVersion and kind.
- **`metadata.name`** names this resource. The name is used in state tracking, logging, and resource identification.
- **`metadata.labels`** control how OpenMCF processes the manifest:
  - `openmcf.org/provisioner: pulumi` tells the CLI to route this deployment through the Pulumi engine. The alternative is `tofu` for OpenTofu/Terraform.
  - The three `pulumi.openmcf.org/*` labels configure the Pulumi stack identity — where deployment state is stored. For local development, any values work.
- **`spec`** defines the desired state of the resource. Each component kind has its own spec fields, defined by Protocol Buffer schemas with built-in validation.

For a deeper explanation of the manifest model, see [Manifests](/docs/concepts/manifests).

## Validate the Manifest

```bash
openmcf validate -f postgres.yaml
```

Validation checks the manifest against the KubernetesPostgres Protocol Buffer schema. It catches structural errors — missing required fields, invalid field types, values outside allowed ranges — before you attempt a deployment.

If validation passes, the CLI prints a confirmation. If it fails, the error message identifies the exact field and constraint that was violated.

## Prepare for Deployment

Two setup steps before deploying:

**Configure a local Pulumi backend.** Pulumi needs a backend to store deployment state. For local development, use file-based storage:

```bash
pulumi login --local
```

This stores state in `~/.pulumi/` on your machine. For team or production use, OpenMCF supports Pulumi Cloud, S3, GCS, and Azure Blob backends. See [State Management](/docs/concepts/state-management) for details.

**Initialize the stack.** A Pulumi stack is a unit of deployment state — it tracks what resources exist and their current configuration. Create one for this deployment:

```bash
openmcf init -f postgres.yaml
```

This reads the stack labels from the manifest (`local/getting-started/dev`) and registers the stack with your configured backend. The command is idempotent — running it again on an existing stack is safe.

## Deploy

```bash
openmcf apply -f postgres.yaml
```

The CLI loads the manifest, resolves the Pulumi module for KubernetesPostgres from the OpenMCF repository, and executes the deployment. You will see Pulumi's output as it creates Kubernetes resources — a StatefulSet, Service, PersistentVolumeClaim, and supporting objects.

The first run takes longer because the CLI clones the IaC module from GitHub. Subsequent runs use the cached module at `~/.openmcf/modules/`.

## Verify

Check that the PostgreSQL pod is running:

```bash
kubectl get pods -n my-first-postgres
```

You should see a pod named `my-first-postgres-postgresql-0` with status `Running`. It may take a minute for the pod to pull the PostgreSQL image and start.

Check the service:

```bash
kubectl get svc -n my-first-postgres
```

## Clean Up

Destroy the deployed resources:

```bash
openmcf destroy -f postgres.yaml
```

This removes all Kubernetes resources that were created by the deployment.

Optionally, delete the Kind cluster:

```bash
kind delete cluster --name openmcf-quickstart
```

## What Just Happened

When you ran `openmcf apply`, the CLI executed this pipeline:

1. **Loaded** the manifest from `postgres.yaml` and applied Protocol Buffer validation
2. **Read** the `openmcf.org/provisioner: pulumi` label and routed execution to the Pulumi engine
3. **Resolved** the Pulumi module for KubernetesPostgres — a Go program that translates the spec into Kubernetes resources
4. **Built** a stack input from the manifest and passed it to the Pulumi program as configuration
5. **Executed** `pulumi up`, which created a StatefulSet, Service, PersistentVolumeClaim, and associated resources in your cluster

This is the same workflow for every deployment component in OpenMCF. Whether you deploy an AWS S3 bucket, a GCP Cloud SQL instance, or a Cloudflare Worker, the pattern is identical: write a manifest, validate, init, apply. The manifest fields change; the workflow does not.

## Next Steps

You have deployed your first resource with OpenMCF. Here is where to go next:

- **Understand the model.** Read [Core Concepts](/docs/concepts) to learn how deployment components, manifests, validation, and the dual IaC engine system work together.
- **Deploy to a cloud provider.** Follow the [AWS S3 Bucket tutorial](/docs/tutorials/first-aws-resource) or the [Multi-Provider tutorial](/docs/tutorials/multi-provider) to deploy real cloud infrastructure.
- **Go deeper with Kubernetes.** The [Kubernetes Postgres tutorial](/docs/tutorials/first-kubernetes-resource) builds on this guide with custom databases, named users, resource tuning, and runtime overrides.
- **Set up cloud credentials.** Configure [AWS](/docs/guides/aws-provider-setup), [GCP](/docs/guides/gcp-provider-setup), or [Azure](/docs/guides/azure-provider-setup) for production deployments.
- **Explore the catalog.** Browse all [360+ deployment components](/docs/catalog) across 17 cloud providers.
- **Troubleshoot issues.** Check the [Troubleshooting Guide](/docs/troubleshooting) if you run into problems.
