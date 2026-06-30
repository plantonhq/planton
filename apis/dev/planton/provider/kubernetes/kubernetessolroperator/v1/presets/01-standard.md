# Standard Apache Solr Operator

This preset deploys the Apache Solr Operator with recommended default resources. The operator manages the lifecycle of SolrCloud clusters on Kubernetes, enabling declarative provisioning, scaling, and configuration of Solr search infrastructure.

## When to Use

- You need to run Apache Solr on Kubernetes
- You want operator-managed SolrCloud cluster lifecycle (create, scale, configure, upgrade)
- Standard resource allocation is sufficient for the operator control plane

## Key Configuration Choices

- **Namespace** (`solr-system`) -- dedicated namespace isolates the operator from Solr workloads it manages
- **Create namespace** (`true`) -- namespace is created automatically if it does not exist
- **Resource requests** (`50m` CPU, `100Mi` memory) -- lightweight baseline for the operator pod
- **Resource limits** (`1000m` CPU, `1Gi` memory) -- sufficient headroom for managing multiple SolrCloud clusters

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.
