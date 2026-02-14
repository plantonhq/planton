# Single Instance Neo4j

This preset deploys a single-replica Neo4j graph database with persistence. Neo4j is memory-intensive, so this preset allocates higher memory than typical database defaults.

## When to Use

- Development, testing, or small production graph workloads
- Applications using Cypher queries on graph data
- Knowledge graphs, recommendation engines, or network analysis

## Key Configuration Choices

- **Single replica** -- standalone Neo4j instance without clustering
- **Higher memory** (`512Mi` request, `4Gi` limit) -- Neo4j performs best with data cached in memory
- **10Gi disk** -- persistent storage for the graph database; increase for larger graphs

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |
