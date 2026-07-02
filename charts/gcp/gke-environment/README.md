# GKE Environment

The **GKE Environment** InfraChart provisions a complete, production‑ready Kubernetes environment on Google
Cloud Platform.  
It now supports **conditional resource generation via Jinja templates**—for example, you can choose whether to create a
brand‑new GCP project and which Kubernetes add‑ons to install.

Chart resources and configuration parameters are defined in the [`templates`](templates) directory and documented in [
`values.yaml`](values.yaml).

---

## Included Cloud Resources (conditional)

| Resource                                                    | Always created | Controlled by boolean flag   |
|-------------------------------------------------------------|----------------|------------------------------|
| **GCP Project**                                             | _No_           | `create_project`             |
| **Custom VPC Network**                                      | Yes            | —                            |
| **Cloud NAT / Router**                                      | Yes            | —                            |
| **Private GKE Cluster**                                     | Yes            | —                            |
| **Autoscaled Node Pool**                                    | Yes            | —                            |
| **Optional Kubernetes Add‑ons** (Cert‑Manager, Istio, etc.) | _No_           | Individual `*$Enabled` flags |

### How the `create_project` flag works

* `create_project: true`→ A fresh **`GcpProject`** resource is rendered, and other resources reference the resulting
  `project_id` via `valueFrom`.
* `create_project: false`→ The chart reuses the supplied `project_id`, and all `projectId` fields are populated directly
  with that value.

---

## Kubernetes Add‑ons (toggleable)

Each add‑on has its own boolean switch (default **`true`** for backward compatibility):

| Flag                      | Add‑on                  |
|---------------------------|-------------------------|
| `certManagerEnabled`      | Cert‑Manager            |
| `elasticOperatorEnabled`  | Elastic Operator        |
| `externalDnsEnabled`      | External‑DNS            |
| `externalSecretsEnabled`  | External‑Secrets        |
| `ingressNginxEnabled`     | Ingress‑NGINX           |
| `istioEnabled`            | Istio (Ingress Gateway) |
| `kafkaOperatorEnabled`    | Kafka Operator          |
| `postgresOperatorEnabled` | PostgreSQL Operator     |
| `solrOperatorEnabled`     | Solr Operator           |

Setting a flag to `false` omits the corresponding manifest from the final render.

---

## Chart Input Values

The table below lists every configurable value.  
Booleans are shown as **unquoted YAML booleans** (`true`/`false`) to avoid string/boolean casting surprises.

| Parameter                         | Description                                               | Example / Options         | Required / Default                |
|-----------------------------------|-----------------------------------------------------------|---------------------------|-----------------------------------|
| **create_project**                | `true` → create new project, `false` → reuse `project_id` | `true` / `false`          | **Default:** `false`              |
| **project_id**                    | GCP Project ID (created or reused)                        | `my-gke-demo-123`         | Required                          |
| **parent_type**                   | Parent resource type                                      | `organization` / `folder` | Required if `create_project=true` |
| **parent_id**                     | Numeric parent ID                                         | `"123456789012"`          | Required if `create_project=true` |
| **billing_account_id**            | Billing account ID                                        | `0123AB-4567CD-89EFGH`    | Required if `create_project=true` |
| **vpc_name**                      | VPC name                                                  | `gke-vpc`                 | Required                          |
| **vpc_auto_create_subnetworks**   | Use auto‑mode VPC (not recommended)                       | `true` / `false`          | Default `false`                   |
| **region**                        | GCP region for subnet / NAT                               | `us-central1`             | Required                          |
| **subnet_name**                   | Subnetwork name                                           | `gke-subnet`              | Required                          |
| **subnet_primary_cidr**           | Primary subnet CIDR                                       | `10.0.0.0/17`             | Required                          |
| **pods_secondary_range_name**     | Secondary range name for Pods                             | `pods`                    | Required                          |
| **pods_secondary_cidr**           | Pods secondary CIDR                                       | `10.1.0.0/17`             | Required                          |
| **services_secondary_range_name** | Secondary range name for Services                         | `services`                | Required                          |
| **services_secondary_cidr**       | Services secondary CIDR                                   | `10.3.0.0/22`             | Required                          |
| **private_ip_google_access**      | Enable Private Google Access                              | `true` / `false`          | Default `true`                    |
| **router_nat_name**               | Cloud NAT / Router name                                   | `gke-nat`                 | Required                          |
| **cluster_name**                  | GKE cluster name                                          | `gke-demo`                | Required                          |
| **cluster_location**              | Cluster region/zone                                       | `us-central1`             | Required                          |
| **master_ipv4_cidr**              | `/28` CIDR for master                                     | `172.16.0.16/28`          | Required                          |
| **enable_public_nodes**           | Give nodes external IPs                                   | `true` / `false`          | Default `false`                   |
| **release_channel**               | RAPID / REGULAR / STABLE / NONE                           | `REGULAR`                 | Default `REGULAR`                 |
| **disable_network_policy**        | Disable Calico NetworkPolicy                              | `true` / `false`          | Default `false`                   |
| **disable_workload_identity**     | Disable Workload Identity                                 | `true` / `false`          | Default `false`                   |
| **node_pool_name**                | Node‑pool name                                            | `default-pool`            | Required                          |
| **node_pool_machine_type**        | Compute machine type                                      | `e2-medium`               | Required                          |
| **node_pool_disk_size_gb**        | Disk size (GB)                                            | `100`                     | Required                          |
| **node_pool_disk_type**           | `pd-standard` / `pd-ssd` / `pd-balanced`                  | `pd-standard`             | Default                           |
| **node_pool_image_type**          | `COS_CONTAINERD` / `COS` / `UBUNTU`                       | `COS_CONTAINERD`          | Default                           |
| **node_pool_spot**                | Use Spot instances                                        | `true` / `false`          | Default `false`                   |
| **node_pool_min_nodes**           | Autoscaler min                                            | `1`                       | Default                           |
| **node_pool_max_nodes**           | Autoscaler max                                            | `3`                       | Default                           |
| **dns_zone_name**                 | Cloud DNS zone name                                       | `demo-example-com`        | Required                          |
| **dns_domain_name**               | FQDN managed by zone (trailing dot)                       | `demo.example.com.`       | Required                          |
| **certManagerEnabled**            | Install Cert‑Manager                                      | `true` / `false`          | Default `true`                    |
| **elasticOperatorEnabled**        | Install Elastic Operator                                  | `true` / `false`          | Default `true`                    |
| **externalDnsEnabled**            | Install External‑DNS                                      | `true` / `false`          | Default `true`                    |
| **externalSecretsEnabled**        | Install External‑Secrets                                  | `true` / `false`          | Default `true`                    |
| **ingressNginxEnabled**           | Install Ingress‑NGINX                                     | `true` / `false`          | Default `true`                    |
| **istioEnabled**                  | Install Istio (Ingress)                                   | `true` / `false`          | Default `true`                    |
| **kafkaOperatorEnabled**          | Install Kafka Operator                                    | `true` / `false`          | Default `true`                    |
| **postgresOperatorEnabled**       | Install PostgreSQL Operator                               | `true` / `false`          | Default `true`                    |
| **solrOperatorEnabled**           | Install Solr Operator                                     | `true` / `false`          | Default `true`                    |

> **Tip:**Set any of the `*$Enabled` flags to `false` to skip that add‑on entirely.

---

## Customization & Management

* Flip `create_project` to quickly spin up isolated test environments or to reuse an existing billing‑configured
  project.
* Enable or disable add‑ons per environment simply by overriding their boolean flags in a higher‑priority values file.
* Resource references (`valueFrom` vs `value`) are automatically wired by the templates—no manual edits needed.

---

## Important Notes

* Ensure subnet CIDRs, NAT, and DNS settings comply with your organization’s network standards.
* When reusing an existing project (`create_project: false`), verify that required APIs (`compute`, `container`, `dns`,
  `iam`) are already enabled.
* Disabling Workload Identity (`disable_workload_identity: true`) affects IAM design; plan service‑account mapping
  accordingly.

---

© Planton. Licensed under [Apache-2.0](../../../LICENSE).
