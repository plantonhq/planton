# Standard Elastic Operator (ECK)

This preset deploys the Elastic Cloud on Kubernetes (ECK) operator with recommended default resources. ECK manages the full lifecycle of Elasticsearch, Kibana, APM Server, Fleet, and other Elastic Stack components on Kubernetes.

## When to Use

- You need to run Elasticsearch or other Elastic Stack components on Kubernetes
- You want operator-managed deployment, scaling, and upgrades of Elastic resources
- Standard resource allocation is sufficient for the operator control plane

## Key Configuration Choices

- **Namespace** (`elastic-system`) -- the conventional namespace for ECK, isolates the operator from managed resources
- **Create namespace** (`true`) -- namespace is created automatically if it does not exist
- **Resource requests** (`50m` CPU, `100Mi` memory) -- lightweight baseline; the operator itself is not resource-intensive
- **Resource limits** (`1000m` CPU, `1Gi` memory) -- headroom for reconciliation of multiple Elasticsearch clusters

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.
