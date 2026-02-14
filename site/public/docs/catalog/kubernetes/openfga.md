---
title: "OpenFGA"
description: "OpenFGA deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesopenfga"
---

# Kubernetes OpenFGA

Deploys OpenFGA on Kubernetes using the official OpenFGA Helm chart. Supports configurable replicas, resource tuning, PostgreSQL or MySQL backends with secure password handling via Kubernetes Secrets, and optional external access through Istio Gateway ingress with automatic TLS.

## What Gets Created

When you deploy a KubernetesOpenFga resource, OpenMCF provisions:

- **Kubernetes Namespace** — created if `createNamespace` is `true`
- **OpenFGA Helm Release** — installs the upstream [openfga/openfga](https://github.com/openfga/helm-charts) Helm chart (v0.2.12), which creates:
  - Deployment with the configured number of replicas
  - Kubernetes Service for cluster-internal access (port 8080)
  - Datastore connection configured for PostgreSQL or MySQL
- **Istio Ingress Resources** (when ingress is enabled):
  - cert-manager Certificate for TLS
  - Gateway API Gateway with HTTPS and HTTP listeners
  - HTTPRoute for HTTPS traffic to the OpenFGA service
  - HTTPRoute for HTTP-to-HTTPS redirect (301)

## Prerequisites

- **A Kubernetes cluster** with kubectl configured
- **A running PostgreSQL or MySQL database** accessible from the cluster
- **Istio** and **cert-manager** installed on the cluster (only if using ingress)
- **A ClusterIssuer** matching the ingress domain (only if using ingress)

## Quick Start

Create a file `openfga.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesOpenFga
metadata:
  name: my-openfga
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesOpenFga.my-openfga
spec:
  namespace:
    value: openfga-dev
  createNamespace: true
  datastore:
    engine: postgres
    host: postgres.database.svc.cluster.local
    database: openfga
    username: openfga
    password:
      value: changeme
```

Deploy:

```shell
openmcf apply -f openfga.yaml
```

This creates a single-replica OpenFGA instance connected to a PostgreSQL database in the `openfga-dev` namespace.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for the OpenFGA deployment. Use `value` for a direct string or `valueFrom` to reference a KubernetesNamespace resource. | Required |
| `datastore.engine` | `string` | Database engine type. | Must be `"postgres"` or `"mysql"` |
| `datastore.host` | `string` | Hostname or endpoint of the database server. | Required |
| `datastore.database` | `string` | Name of the database to connect to. | Required |
| `datastore.username` | `string` | Username for database authentication. | Required |
| `datastore.password` | `KubernetesSensitiveValue` | Database password. Provide as `value` (plain string) or `secretRef` (reference to an existing Kubernetes Secret with `name` and `key`). | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `createNamespace` | `bool` | `false` | Create the namespace if it does not exist. |
| `container.replicas` | `int` | `1` | Number of OpenFGA pod replicas. |
| `container.resources.limits.cpu` | `string` | `"1000m"` | CPU limit per pod. |
| `container.resources.limits.memory` | `string` | `"1Gi"` | Memory limit per pod. |
| `container.resources.requests.cpu` | `string` | `"50m"` | CPU request per pod. |
| `container.resources.requests.memory` | `string` | `"100Mi"` | Memory request per pod. |
| `datastore.port` | `int` | `5432` (postgres) / `3306` (mysql) | Port number of the database server. Must be between 1 and 65535. |
| `datastore.isSecure` | `bool` | `false` | Enable SSL/TLS for the database connection. Adds `sslmode=require` for PostgreSQL or `tls=true` for MySQL. |
| `ingress.enabled` | `bool` | `false` | Expose OpenFGA externally via Istio Gateway with TLS. |
| `ingress.hostname` | `string` | — | Hostname for external access (e.g., `openfga.example.com`). Required when `ingress.enabled` is `true`. |

## Examples

### Development Setup with Inline Password

A minimal deployment for local development using a plain-text password:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesOpenFga
metadata:
  name: dev-openfga
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesOpenFga.dev-openfga
spec:
  namespace:
    value: openfga-dev
  createNamespace: true
  container:
    replicas: 1
    resources:
      requests:
        cpu: "50m"
        memory: "100Mi"
      limits:
        cpu: "500m"
        memory: "512Mi"
  datastore:
    engine: postgres
    host: postgres.database.svc.cluster.local
    database: openfga
    username: openfga
    password:
      value: dev-password
```

### Production with Secret Reference and MySQL

A production deployment using MySQL with the password sourced from a Kubernetes Secret and SSL enabled:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesOpenFga
metadata:
  name: prod-openfga
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesOpenFga.prod-openfga
spec:
  namespace:
    value: openfga-prod
  createNamespace: true
  container:
    replicas: 3
    resources:
      requests:
        cpu: "250m"
        memory: "512Mi"
      limits:
        cpu: "2000m"
        memory: "2Gi"
  datastore:
    engine: mysql
    host: mysql-primary.databases.svc.cluster.local
    port: 3306
    database: openfga
    username: openfga_app
    password:
      secretRef:
        name: openfga-db-credentials
        key: password
    isSecure: true
```

### Full-Featured with Ingress and PostgreSQL

External access via Istio Gateway with TLS, backed by a secure PostgreSQL connection:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesOpenFga
metadata:
  name: openfga-main
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesOpenFga.openfga-main
spec:
  namespace:
    value: authorization
  createNamespace: true
  container:
    replicas: 2
    resources:
      requests:
        cpu: "200m"
        memory: "256Mi"
      limits:
        cpu: "1000m"
        memory: "1Gi"
  datastore:
    engine: postgres
    host: postgres-primary.databases.svc.cluster.local
    port: 5432
    database: openfga
    username: openfga_app
    password:
      secretRef:
        name: openfga-db-credentials
        key: password
    isSecure: true
  ingress:
    enabled: true
    hostname: openfga.example.com
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where OpenFGA was created |
| `service` | `string` | Name of the Kubernetes service for OpenFGA |
| `port_forward_command` | `string` | Ready-to-run `kubectl port-forward` command for local access on port 8080 |
| `kube_endpoint` | `string` | Cluster-internal endpoint (e.g., `my-openfga.namespace.svc.cluster.local`) |
| `external_hostname` | `string` | External hostname when ingress is enabled |
| `internal_hostname` | `string` | Internal hostname for access from within the cluster network |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — pre-create a namespace to reference via `valueFrom`
- [KubernetesPostgres](/docs/catalog/kubernetes/postgres) — deploy a PostgreSQL cluster as the OpenFGA datastore backend
- [KubernetesSecret](/docs/catalog/kubernetes/secret) — manage Kubernetes Secrets for database credentials
