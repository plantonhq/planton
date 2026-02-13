---
title: "Deploy Your First Kubernetes Resource"
description: "Deploy a production-oriented PostgreSQL database on Kubernetes with custom databases, users, and resource tuning"
icon: "tutorial"
order: 30
---

# Deploy Your First Kubernetes Resource

In this tutorial, you will deploy a PostgreSQL database to a Kubernetes cluster using OpenMCF. Unlike the [Getting Started](../getting-started) guide, which covers a minimal deployment, this tutorial builds a production-oriented configuration with custom databases, named users, tuned resource limits, and persistent storage.

By the end, you will know how to configure a Kubernetes component with real-world settings, connect to the deployed database, modify the deployment through manifest changes and runtime overrides, and tear it down cleanly.

## What You Will Build

A PostgreSQL deployment on Kubernetes with:

- A dedicated namespace
- Custom resource limits and persistent storage
- Two named databases (`appdb` and `analyticsdb`)
- A named user (`appuser`) with ownership of both databases
- Port-forwarding for local access

## Prerequisites

Before starting, ensure you have:

- **OpenMCF CLI** installed (`openmcf version`). See [Getting Started](../getting-started) for installation.
- **A Kubernetes cluster** running and accessible via `kubectl`. A local cluster works fine — [Getting Started](../getting-started) shows how to create one with Kind.
- **Pulumi CLI** installed (`brew install pulumi`) with a backend configured (`pulumi login --local` for local state).
- **psql** (PostgreSQL client) installed for verification (`brew install libpq`). Optional but recommended.

## Step 1: Write the Manifest

Create a file named `postgres.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: tutorial-postgres
  labels:
    openmcf.org/provisioner: pulumi
spec:
  namespace:
    value: tutorial-postgres
  createNamespace: true
  container:
    replicas: 1
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 1000m
        memory: 1Gi
    diskSize: 5Gi
  databases:
    - name: appdb
      ownerRole: appuser
    - name: analyticsdb
      ownerRole: appuser
  users:
    - name: appuser
      flags:
        - login
        - createdb
  ingress:
    enabled: false
```

Here is what each section configures:

### Namespace

```yaml
namespace:
  value: tutorial-postgres
createNamespace: true
```

The `namespace` field uses a `StringValueOrRef` type — you provide the namespace name inside a `value` wrapper. Setting `createNamespace: true` tells the component to create the namespace if it does not exist.

### Container Resources

```yaml
container:
  replicas: 1
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: 1000m
      memory: 1Gi
  diskSize: 5Gi
```

This mirrors standard Kubernetes resource specifications. The `diskSize` field sets the PersistentVolumeClaim size for PostgreSQL data. It must be a valid Kubernetes quantity (e.g., `1Gi`, `500Mi`, `10Gi`).

### Databases and Users

```yaml
databases:
  - name: appdb
    ownerRole: appuser
  - name: analyticsdb
    ownerRole: appuser
users:
  - name: appuser
    flags:
      - login
      - createdb
```

The component creates these databases and users during initialization. The `ownerRole` field references a user declared in the `users` list. If you omit `databases` and `users`, only the default `postgres` database and superuser are available.

Available user flags: `login`, `createdb`, `superuser`, `createrole`, `inherit`, `replication`.

### Ingress

```yaml
ingress:
  enabled: false
```

When `enabled` is `false`, PostgreSQL is only accessible within the cluster (or via port-forwarding). Setting it to `true` requires a `hostname` field and creates a LoadBalancer service with external-dns annotations.

## Step 2: Preview the Deployment

Preview what OpenMCF will create:

```bash
openmcf plan -f postgres.yaml
```

The plan shows the Kubernetes resources that will be created: namespace, StatefulSet, PersistentVolumeClaim, Services, Secrets, and the PostgreSQL databases and users. Review the output to confirm it matches your expectations.

## Step 3: Deploy

Apply the manifest:

```bash
openmcf apply -f postgres.yaml
```

OpenMCF resolves the `KubernetesPostgres` component module, sets up the deployment environment with your manifest as input, and delegates to Pulumi. The deployment creates the namespace, installs the PostgreSQL operator resources, and initializes your databases and users.

## Step 4: Verify

Check that the pods are running:

```bash
kubectl get pods -n tutorial-postgres
```

You should see a PostgreSQL pod in `Running` state.

Check that the namespace and services exist:

```bash
kubectl get svc -n tutorial-postgres
```

The deployment outputs include connection information:

| Output | Description |
|--------|-------------|
| `namespace` | The Kubernetes namespace where PostgreSQL was deployed |
| `service` | The name of the Kubernetes Service |
| `port_forward_command` | Ready-to-use command for local access via port-forwarding |
| `kube_endpoint` | In-cluster endpoint for applications in the same cluster |
| `username_secret` | Kubernetes Secret name and key containing the username |
| `password_secret` | Kubernetes Secret name and key containing the password |

### Connect via Port-Forwarding

Set up port-forwarding to access PostgreSQL locally. Use the `port_forward_command` from the deployment outputs, or construct it manually:

```bash
kubectl port-forward svc/<service-name> -n tutorial-postgres 5432:5432
```

Retrieve the credentials from the Kubernetes Secret:

```bash
# Get the password (secret name and key from deployment outputs)
kubectl get secret <secret-name> -n tutorial-postgres -o jsonpath='{.data.<key>}' | base64 -d
```

Connect with `psql`:

```bash
psql -h localhost -p 5432 -U appuser -d appdb
```

Verify your databases exist:

```sql
\l
```

You should see `appdb` and `analyticsdb` in the database list, both owned by `appuser`.

## Step 5: Modify the Deployment

### Option A: Update the Manifest

Add a third database by editing `postgres.yaml`:

```yaml
  databases:
    - name: appdb
      ownerRole: appuser
    - name: analyticsdb
      ownerRole: appuser
    - name: stagingdb
      ownerRole: appuser
```

Then re-apply:

```bash
openmcf apply -f postgres.yaml
```

OpenMCF computes the diff and applies only the change — the new database is created without affecting the existing ones.

### Option B: Use Runtime Overrides

Scale the replicas without editing the manifest file:

```bash
openmcf apply -f postgres.yaml --set spec.container.replicas=2
```

The `--set` flag overrides manifest values at deploy time. This is useful for CI/CD pipelines where environment-specific values differ from the base manifest. See [CLI Reference](../cli/cli-reference) for full details on the `--set` flag.

## Step 6: Clean Up

Destroy the deployment:

```bash
openmcf destroy -f postgres.yaml
```

This removes all Kubernetes resources created by the component — the StatefulSet, Services, PersistentVolumeClaims, Secrets, and the namespace (if it was created by the component).

## What You Learned

- How to configure a Kubernetes deployment component with custom databases, users, and resource limits
- The `StringValueOrRef` pattern for fields that can reference other resources (like `namespace`)
- How deployment outputs provide connection information including Secrets for credentials
- Two ways to modify deployments: updating the manifest file, or using `--set` for runtime overrides
- The declarative lifecycle: `plan` to preview, `apply` to deploy, `destroy` to clean up

## What's Next

- [Multi-Environment Deployments](./multi-environment) — deploy this same PostgreSQL across dev, staging, and prod using Kustomize overlays
- [Deploy Across Providers](./multi-provider) — deploy object storage on both AWS and GCP
- [Manifests](../concepts/manifests) — deep dive into the KRM model and label-based configuration
- [Validation](../concepts/validation) — understand how OpenMCF validates manifests before deployment
