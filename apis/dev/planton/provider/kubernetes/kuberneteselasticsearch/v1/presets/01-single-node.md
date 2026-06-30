# Single Node Elasticsearch

This preset deploys a single-node Elasticsearch instance with persistence. No Kibana. Suitable for development, testing, or small log/search workloads.

## When to Use

- Development or testing Elasticsearch queries and mappings
- Small search indexes or log volumes
- Environments where Kibana visualization is not needed

## Key Configuration Choices

- **Single node** -- no clustering or sharding; all data on one node
- **Persistence enabled** with 10Gi disk -- index data survives pod restarts
- **Higher memory** (`512Mi` request, `4Gi` limit) -- Elasticsearch JVM heap and filesystem cache benefit from generous memory
- **No Kibana** -- deploy Kibana separately or use preset 02-with-kibana

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |

## Related Presets

- **02-with-kibana** -- Single-node Elasticsearch with Kibana for visualization
- **03-production-cluster** -- Multi-node cluster for production
