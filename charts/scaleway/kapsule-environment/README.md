# Scaleway Kapsule Environment

The **Scaleway Kapsule Environment** InfraChart provisions a complete, production-ready
Kubernetes environment on Scaleway Cloud.
It supports **conditional resource generation via Jinja templates** -- for example, you can
choose whether to create a DNS zone, container registry, additional node pool, and which
Kubernetes add-ons to install.

Chart resources and configuration parameters are defined in the [`templates`](templates)
directory and documented in [`values.yaml`](values.yaml).

---

## Architecture

The chart creates a layered infrastructure stack with the following dependency flow:

```
ScalewayVpc
  └── ScalewayPrivateNetwork
        └── ScalewayKapsuleCluster (with embedded default node pool)
              ├── ScalewayKapsulePool (optional additional pool)
              └── Kubernetes Add-ons (optional, 9 available)

ScalewayDnsZone (optional, independent)
ScalewayContainerRegistry (optional, independent)
```

All resources in the stack communicate over the Private Network. The Kapsule cluster
requires a Private Network (Scaleway mandate), so VPC and Private Network are always
created.

---

## Included Cloud Resources (conditional)

| Resource | Always created | Controlled by boolean flag |
|----------|----------------|----------------------------|
| **Scaleway VPC** | Yes | -- |
| **Scaleway Private Network** | Yes | -- |
| **Scaleway Kapsule Cluster** (with default node pool) | Yes | -- |
| **Scaleway Kapsule Pool** (additional) | No | `create_additional_pool` |
| **Scaleway DNS Zone** | No | `create_dns_zone` |
| **Scaleway Container Registry** | No | `create_container_registry` |
| **Kubernetes Add-ons** (Cert-Manager, Istio, etc.) | No | Individual `*Enabled` flags |

---

## Kubernetes Add-ons (toggleable)

Each add-on has its own boolean switch (default **true** for backward compatibility):

| Flag | Add-on |
|------|--------|
| `certManagerEnabled` | Cert-Manager |
| `elasticOperatorEnabled` | Elastic Operator |
| `externalDnsEnabled` | External-DNS |
| `externalSecretsEnabled` | External-Secrets |
| `ingressNginxEnabled` | Ingress-NGINX |
| `istioEnabled` | Istio (Ingress Gateway) |
| `kafkaOperatorEnabled` | Kafka Operator (Strimzi) |
| `postgresOperatorEnabled` | PostgreSQL Operator (Zalando) |
| `solrOperatorEnabled` | Solr Operator |

Setting a flag to `false` omits the corresponding manifest from the final render.

---

## Chart Input Values

| Parameter | Description | Example / Options | Required / Default |
|-----------|-------------|-------------------|-------------------|
| **region** | Scaleway region | `fr-par`, `nl-ams`, `pl-waw` | Default `fr-par` |
| **cluster_name** | Kubernetes cluster name | `kapsule-demo` | Required |
| **kubernetes_version** | Kubernetes version | `1.32` | Required |
| **cni** | Container Network Interface | `cilium`, `calico` | Default `cilium` |
| **delete_additional_resources** | Delete K8s-created LBs/volumes on destroy | `true` / `false` | Default `true` |
| **node_type** | Default pool node instance type | `DEV1-M`, `GP1-XS`, `PRO2-S` | Required |
| **node_count** | Default pool node count | `2` | Required |
| **node_auto_scale** | Enable autoscaling (default pool) | `true` / `false` | Default `false` |
| **min_nodes** | Autoscaler minimum (default pool) | `1` | Default `1` |
| **max_nodes** | Autoscaler maximum (default pool) | `6` | Default `6` |
| **create_additional_pool** | Create additional node pool | `true` / `false` | Default `false` |
| **additional_pool_name** | Additional pool name | `workload` | Default `workload` |
| **additional_pool_node_type** | Additional pool instance type | `GP1-XS` | Default `GP1-XS` |
| **additional_pool_node_count** | Additional pool node count | `3` | Default `3` |
| **additional_pool_auto_scale** | Enable autoscaling (additional pool) | `true` / `false` | Default `false` |
| **additional_pool_min_nodes** | Autoscaler minimum (additional pool) | `1` | Default `1` |
| **additional_pool_max_nodes** | Autoscaler maximum (additional pool) | `10` | Default `10` |
| **create_dns_zone** | Create Scaleway DNS zone | `true` / `false` | Default `false` |
| **domain_name** | Parent domain for DNS zone | `example.com` | Required if `create_dns_zone=true` |
| **subdomain** | Subdomain prefix | `staging` | Optional |
| **create_container_registry** | Create container registry | `true` / `false` | Default `false` |
| **registry_name** | Registry namespace name | `my-registry` | Required if `create_container_registry=true` |
| **certManagerEnabled** | Install Cert-Manager | `true` / `false` | Default `true` |
| **elasticOperatorEnabled** | Install Elastic Operator | `true` / `false` | Default `true` |
| **externalDnsEnabled** | Install External-DNS | `true` / `false` | Default `true` |
| **externalSecretsEnabled** | Install External-Secrets | `true` / `false` | Default `true` |
| **ingressNginxEnabled** | Install Ingress-NGINX | `true` / `false` | Default `true` |
| **istioEnabled** | Install Istio (Ingress) | `true` / `false` | Default `true` |
| **kafkaOperatorEnabled** | Install Kafka Operator | `true` / `false` | Default `true` |
| **postgresOperatorEnabled** | Install PostgreSQL Operator | `true` / `false` | Default `true` |
| **solrOperatorEnabled** | Install Solr Operator | `true` / `false` | Default `true` |

> **Tip:** Set any of the `*Enabled` flags to `false` to skip that add-on entirely.

---

## Customization & Management

* Enable or disable add-ons per environment simply by overriding their boolean flags
  in a higher-priority values file.
* Use `create_additional_pool` to add a dedicated workload pool with a different instance
  type alongside the default system pool.
* Resource references (`valueFrom` vs `value`) are automatically wired by the templates --
  no manual edits needed.
* The container registry, when created, is private by default for security.
* Autohealing is enabled on all node pools for production resilience.

---

## Important Notes

* **Kapsule uses Kubernetes (not K3s)** -- Scaleway Kapsule runs full upstream Kubernetes,
  unlike some lightweight providers.
* **CNI is immutable** -- The CNI plugin (Cilium or Calico) cannot be changed after cluster
  creation. Choose Cilium for eBPF-based networking with Hubble observability, or Calico
  for teams already familiar with its network policy model.
* **Private Network is required** -- Scaleway mandates a Private Network for all Kapsule
  clusters. The chart always creates a VPC and Private Network.
* **No Security Group needed** -- Kapsule auto-manages firewall rules for the cluster.
  Unlike Civo, no explicit firewall resource is required.
* Ensure the Kubernetes version specified is supported by Scaleway (check Scaleway
  documentation for available versions per region).
* Node instance types must be eligible for Kubernetes (instances with insufficient memory
  like DEV1-S and STARDUST are not supported).
* When creating a DNS zone, ensure the domain is available and you have access to
  configure nameservers at your domain registrar.

---

(c) 2025 Planton. All rights reserved.
