# Temporal on Kubernetes Pulumi Module

This Planton component simplifies deploying [Temporal](https://temporal.io/) onto Kubernetes clusters using
Pulumi and the official Temporal Helm chart.

- Aligns with Terraform module conventions for familiarity.
- Organizes resources via clear, noun-based helper functions (`namespace()`, `helmChart()`, `frontendIngress()`, etc.).
- Self-contained in a single folder for straightforward integration into your DevOps workflows.
- Precisely mirrors the fields defined by `TemporalKubernetesSpec`, ensuring simplicity and clarity.

---

## 📋 Prerequisites

| Dependency                             | Purpose                                             |
|----------------------------------------|-----------------------------------------------------|
| **Kubernetes v1.24+**                  | Kubernetes cluster to deploy Temporal               |
| **Pulumi v3 CLI**                      | To deploy and manage Pulumi stacks                  |
| **cert-manager** *(optional)*          | Automates TLS certificates when ingress is enabled  |
| **Istio / Gateway-API** *(optional)*   | Provides secure ingress routing for Temporal Web UI |
| **kubectl** & **Helm** *(recommended)* | Useful for debugging and direct interaction         |

> **Note**: By default, the module automatically provisions a built-in database (Cassandra, MySQL, PostgreSQL). To use
> an external database, set `spec.database.externalDatabase`.

---

## 🚀 Quick Start

Create a manifest file `temporal.yaml` adhering to the protobuf schema:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: TemporalKubernetes
metadata:
  name: my-temporal
spec:
  database:
    backend: cassandra
```

### Deploying with Planton CLI

```bash
planton pulumi up --manifest temporal.yaml --stack acme/dev/my-temporal
```

The CLI performs:

- Schema validation
- Conversion to Pulumi input
- Deployment and output export

---

## 📥 Input Variables

| Input                                  | Description                                                                                |
|----------------------------------------|--------------------------------------------------------------------------------------------|
| `spec.database.backend`                | Select database: `cassandra`, `postgresql`, or `mysql`.                                    |
| `spec.database.externalDatabase.*`     | External DB connection (host, port, user, password); password creates a Kubernetes Secret. |
| `spec.database.disableAutoSchemaSetup` | Disables auto schema initialization.                                                       |
| `spec.disableWebUi`                    | Option to disable Temporal Web UI deployment.                                              |
| `spec.ingress.frontend.enabled`        | Enables external ingress for frontend (gRPC).                                              |
| `spec.ingress.frontend.hostname`       | Full hostname for frontend access (e.g., "temporal-frontend.example.com").                 |
| `spec.ingress.webUi.enabled`           | Enables external ingress for web UI.                                                       |
| `spec.ingress.webUi.hostname`          | Full hostname for web UI access (e.g., "temporal-ui.example.com").                         |
| `spec.externalElasticsearch.*`         | Connects Temporal to an external Elasticsearch cluster for observability.                  |

---

## 📤 Outputs

| Pulumi Export                                               | Description                                          |
|-------------------------------------------------------------|------------------------------------------------------|
| `namespace`                                                 | Kubernetes namespace for Temporal resources          |
| `frontend_service`                                          | gRPC frontend service name                           |
| `ui_service`                                                | Web UI service name                                  |
| `frontend_endpoint` / `ui_endpoint`                         | Fully qualified in-cluster service addresses         |
| `port_forward_frontend_command` / `port_forward_ui_command` | Quick commands to port-forward services locally      |
| `external_frontend_hostname`                                | DNS hostname for external gRPC frontend              |
| `external_ui_hostname`                                      | DNS hostname for external Web UI (via Istio Gateway) |

---

## ⚙️ How It Works

### Database Management

- Auto-provisions a database if external details are not provided.
- Generates a Kubernetes Secret (`temporal-db-password`) for external database credentials only when provided.

### Ingress Setup

- **Frontend (gRPC)**: Exposed via LoadBalancer service on TCP port `7233`.
- **Web UI**: Managed through Istio/Gateway-API with automatic HTTPS and HTTP-to-HTTPS redirection.

### Observability

- Integrates with an external Elasticsearch cluster if specified; otherwise, built-in observability remains inactive.

---

## 📌 Differences from Terraform Module

The Pulumi implementation closely mirrors the Terraform equivalent:

- Maintains consistent file structure and naming conventions.
- Utilizes Go maps and structs instead of Terraform’s `locals` and `variable` blocks.
- Adopts Istio/Gateway-API ingress patterns consistent across other Planton modules.

---

## 🔧 Extending and Customizing

To modify or enhance the module:

- Adjust Helm chart settings (`variables.go`)
- Add custom Helm values (`helm_chart.go`)
- Switch ingress methods (`web_ui_ingress.go`)

---

Enjoy deploying with Planton! 🌿🚀
