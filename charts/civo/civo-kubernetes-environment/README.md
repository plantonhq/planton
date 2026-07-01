# Civo Kubernetes Environment

The **Civo Kubernetes Environment** InfraChart provisions a complete, production-ready Kubernetes environment on Civo Cloud.
It supports **conditional resource generation via Jinja templates**—for example, you can choose whether to create a DNS zone and which Kubernetes add-ons to install.

Chart resources and configuration parameters are defined in the [`templates`](templates) directory and documented in [`values.yaml`](values.yaml).

---

## Included Cloud Resources (conditional)

| Resource | Always created | Controlled by boolean flag |
|----------|----------------|----------------------------|
| **Civo VPC (Network)** | Yes | — |
| **Civo Firewall** | Yes | — |
| **Civo Kubernetes Cluster** | Yes | — |
| **Civo DNS Zone** | No | `create_dns_zone` |
| **Optional Kubernetes Add-ons** (Cert-Manager, Istio, etc.) | No | Individual `*Enabled` flags |

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
| `kafkaOperatorEnabled` | Kafka Operator |
| `postgresOperatorEnabled` | PostgreSQL Operator |
| `solrOperatorEnabled` | Solr Operator |

Setting a flag to `false` omits the corresponding manifest from the final render.

---

## Chart Input Values

| Parameter | Description | Example / Options | Required / Default |
|-----------|-------------|-------------------|-------------------|
| **region** | Civo region | `fra1`, `lon1`, `nyc1`, `phx1`, `mum1` | Required |
| **cluster_name** | Kubernetes cluster name | `civo-demo` | Required |
| **kubernetes_version** | Kubernetes version | `1.26.3` | Required |
| **highly_available** | Enable HA control plane | `true` / `false` | Default `false` |
| **auto_upgrade** | Enable automatic upgrades | `true` / `false` | Default `false` |
| **disable_surge_upgrade** | Disable surge upgrades | `true` / `false` | Default `false` |
| **node_size** | Node instance size | `g4s.kube.medium` | Required |
| **node_count** | Number of nodes | `3` | Required |
| **allow_all_egress** | Allow all outbound traffic | `true` / `false` | Default `true` |
| **create_dns_zone** | Create DNS zone | `true` / `false` | Default `true` |
| **domain_name** | Domain for DNS zone | `example.com` | Required if `create_dns_zone=true` |
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

* Enable or disable add-ons per environment simply by overriding their boolean flags in a higher-priority values file.
* Resource references (`valueFrom` vs `value`) are automatically wired by the templates—no manual edits needed.
* The firewall includes sensible defaults for Kubernetes cluster access (HTTP, HTTPS, K8s API).

---

## Important Notes

* Civo Kubernetes clusters run K3s, a lightweight Kubernetes distribution.
* Ensure the Kubernetes version specified is supported by Civo (check Civo documentation).
* The default node pool size should match available Civo instance types (`g4s.kube.small`, `g4s.kube.medium`, `g4s.kube.large`, etc.).
* When creating a DNS zone, ensure the domain is available and you have access to configure nameservers.

---

© 2025 Planton. All rights reserved.

