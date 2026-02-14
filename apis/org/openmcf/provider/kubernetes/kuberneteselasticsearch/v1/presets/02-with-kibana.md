# Elasticsearch with Kibana

This preset deploys a single-node Elasticsearch instance with Kibana enabled and exposed via ingress. Kibana provides a web UI for searching, visualizing, and dashboarding Elasticsearch data.

## When to Use

- Log analysis and visualization (ELK/EFK stack)
- Teams that need Kibana dashboards for monitoring and observability
- Development or staging environments with a visual Elasticsearch frontend

## Key Configuration Choices

- **Kibana enabled** with ingress -- accessible at the specified hostname via the cluster's ingress controller
- **Single Elasticsearch node** -- suitable for moderate data volumes; scale up with preset 03-production-cluster
- **Kibana resources** -- lightweight defaults; Kibana is primarily a frontend application

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |
| `<your-kibana.example.com>` | Hostname for Kibana ingress access | Your DNS provider |

## Related Presets

- **01-single-node** -- Elasticsearch without Kibana
- **03-production-cluster** -- Multi-node cluster for production
