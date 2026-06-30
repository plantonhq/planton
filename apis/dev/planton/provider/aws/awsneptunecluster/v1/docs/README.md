# Amazon Neptune: Graph Database Architecture, Use Cases, and Query Languages

## Introduction

Amazon Neptune is a fully managed graph database service from AWS designed for applications that need to store and query highly connected data. Unlike relational databases that excel at tabular data with well-defined schemas, graph databases model data as nodes (vertices) and relationships (edges), making them ideal for traversing complex relationships across billions of connections in milliseconds.

Neptune supports two industry-standard graph models and query languages:

1. **Property Graph** with **Apache TinkerPop Gremlin** — vertices and edges with key-value properties; traversal-oriented queries
2. **RDF (Resource Description Framework)** with **SPARQL** — triples (subject-predicate-object); W3C-standard semantic queries

Neptune also supports **openCypher** as an additional query option for property graph workloads. This flexibility allows teams to choose the model and language that best fits their domain—whether building recommendation engines, fraud detection systems, knowledge graphs, or social networks.

## What Is a Graph Database?

A graph database stores data in structures optimized for relationship traversal. The core concepts:

- **Vertices (nodes)**: Entities such as users, products, accounts, or locations
- **Edges (relationships)**: Connections between vertices with direction and optional properties
- **Properties**: Key-value attributes on vertices and edges

Example: In a social network, a "User" vertex might connect to another "User" via a "FOLLOWS" edge. Querying "friends of friends" becomes a natural graph traversal rather than complex JOINs in SQL.

### Property Graph vs. RDF

| Aspect | Property Graph (Gremlin) | RDF (SPARQL) |
|--------|--------------------------|--------------|
| **Model** | Vertices, edges, properties | Triples (subject-predicate-object) |
| **Use case** | Traversal, path finding, recommendations | Semantic web, knowledge graphs, linked data |
| **Query style** | Step-by-step traversal | Declarative SELECT/WHERE |
| **Standard** | Apache TinkerPop | W3C SPARQL 1.1 |

## Neptune Architecture

### High-Level Design

Neptune clusters consist of:

- **Cluster**: The logical container; stores graph data in a distributed, replicated storage layer
- **Instances**: Compute nodes that process queries; one primary (writer) and up to 15 read replicas
- **Storage**: Automatically replicated across multiple Availability Zones; scales up to 64 TiB per cluster
- **Subnet group**: VPC subnets where instances are deployed (minimum 2 subnets in different AZs)
- **Parameter group**: Engine configuration (e.g., query timeout, audit logging)

### Storage Types

- **Standard**: Default; pay-per-I/O model; suitable for most workloads
- **I/O-Optimized (iopt1)**: Higher throughput, predictable pricing; ideal for read-heavy or high-I/O workloads

### Neptune Serverless

Neptune Serverless uses the `db.serverless` instance class with **Neptune Capacity Units (NCUs)**. Capacity scales automatically between 1.0 and 128.0 NCUs based on workload demand. Benefits:

- No capacity planning; pay for what you use
- Instant scaling for spiky or variable traffic
- Same Gremlin/SPARQL APIs as provisioned clusters

## Authentication and Security

**Neptune does not use master username/password.** Access is controlled by:

1. **IAM database authentication**: IAM users and roles authenticate using temporary credentials (SigV4 signing)
2. **Network security**: VPC, security groups, and private subnets restrict who can reach the cluster
3. **TLS**: All connections are encrypted in transit

Default port: **8182** (Gremlin) or **8182** (SPARQL over HTTP).

### IAM Roles for S3 Bulk Loading

Neptune can load graph data from S3 using the `LOAD` command (Gremlin) or `LOAD` (SPARQL). IAM roles associated with the cluster grant Neptune permission to read from S3 buckets. This is essential for bulk ingestion pipelines.

## Query Languages

### Gremlin (Property Graph)

Gremlin is a traversal language: you walk the graph step by step.

```groovy
// Find all people that the user "alice" follows
g.V().has('name', 'alice').out('follows').values('name')

// Friends of friends (2-hop traversal)
g.V().has('name', 'alice').out('follows').out('follows').values('name')

// Shortest path between two users
g.V().has('name', 'alice').repeat(out().simplePath()).until(has('name', 'bob')).path()
```

### SPARQL (RDF)

SPARQL uses SELECT and WHERE clauses to query triples.

```sparql
# Find all people Alice follows
SELECT ?friend WHERE {
  :alice :follows ?friend .
}

# Find mutual connections
SELECT ?person WHERE {
  :alice :follows ?person .
  :bob :follows ?person .
}
```

### openCypher

openCypher provides a SQL-like syntax for property graphs:

```cypher
MATCH (a:Person {name: 'alice'})-[:FOLLOWS]->(f:Person)
RETURN f.name
```

## Use Cases

### 1. Recommendation Engines

- **Product recommendations**: "Users who bought X also bought Y" — traverse co-purchase edges
- **Content recommendations**: Similar users, similar content — multi-hop traversals
- **Personalization**: User preferences, item attributes — property filtering

### 2. Fraud Detection

- **Transaction networks**: Identify rings of connected accounts
- **Pattern detection**: Unusual paths (e.g., money laundering flows)
- **Real-time scoring**: Graph embeddings for ML-based fraud models

### 3. Knowledge Graphs

- **Enterprise knowledge**: Documents, entities, relationships
- **Semantic search**: RDF/SPARQL for ontology-based queries
- **Data lineage**: Track data flow and dependencies

### 4. Social Networks

- **Friend suggestions**: Friends of friends, common connections
- **Feed generation**: Traverse follow graph for relevance
- **Influence analysis**: Centrality, community detection

### 5. Network and IT Operations

- **Infrastructure topology**: Servers, networks, dependencies
- **Impact analysis**: "If this service fails, what is affected?"
- **Root cause analysis**: Trace failures through dependency graph

## CloudWatch Logs

Neptune exports two log types to CloudWatch:

- **audit**: Tracks database activity for compliance and security auditing
- **slowquery**: Logs slow queries for performance tuning

## Best Practices

### High Availability

- Deploy at least 2 instances (1 writer + 1 reader) in different Availability Zones
- Use the reader endpoint for read-only queries to distribute load
- Enable deletion protection for production clusters

### Security

- Use IAM database authentication instead of storing credentials
- Place clusters in private subnets; never enable public access
- Enable storage encryption at rest (default in Planton)
- Associate IAM roles only when S3 bulk loading is required

### Performance

- Use I/O-Optimized storage (`storageType: iopt1`) for read-heavy workloads
- Add read replicas to scale read throughput
- Export slowquery logs and optimize traversal patterns
- Consider Neptune Serverless for variable or unpredictable traffic

### Backup and Recovery

- Configure `backupRetentionPeriod` (1–35 days) for point-in-time recovery
- Set `finalSnapshotIdentifier` when `skipFinalSnapshot` is false
- Use `copyTagsToSnapshot` for consistent tagging

## References

- [Amazon Neptune User Guide](https://docs.aws.amazon.com/neptune/latest/userguide/what-is-neptune.html)
- [Neptune Engine Releases](https://docs.aws.amazon.com/neptune/latest/userguide/engine-releases.html)
- [Neptune Instance Classes](https://docs.aws.amazon.com/neptune/latest/userguide/instance-classes.html)
- [Accessing Neptune with Gremlin](https://docs.aws.amazon.com/neptune/latest/userguide/access-graph-gremlin.html)
- [Accessing Neptune with SPARQL](https://docs.aws.amazon.com/neptune/latest/userguide/access-graph-sparql.html)
- [Apache TinkerPop Gremlin](https://tinkerpop.apache.org/gremlin.html)
- [W3C SPARQL 1.1](https://www.w3.org/TR/sparql11-overview/)
- [Neptune Best Practices](https://docs.aws.amazon.com/neptune/latest/userguide/best-practices.html)
