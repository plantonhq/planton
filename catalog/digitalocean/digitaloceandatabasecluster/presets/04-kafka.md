# Kafka Cluster

This preset creates a three-node Apache Kafka cluster inside a VPC -- the minimum topology DigitalOcean accepts for Kafka, sized for modest event-streaming workloads.

## When to Use

- Event streaming between services (change feeds, activity streams, log pipelines)
- Decoupling producers from consumers with durable, replayable topics
- Workloads that need ordered, at-least-once delivery

## Key Configuration Choices

- **Three nodes** (`nodeCount: 3`) -- Kafka's replication model requires at least 3 nodes on DigitalOcean; this is the floor, not a tuning choice.
- **Kafka 4.2** (`engine: kafka`, `engineVersion: "4.2"`) -- DigitalOcean's default Kafka line at the time this preset was verified (2026-09-16; `3.9` was the other offered version). DigitalOcean retires versions on a schedule and rejects a version it no longer offers at create (`422 invalid cluster engine version`), so check `GET /v2/databases/options` if a deploy fails there.
- **VPC placement** (`vpc.valueFrom`) -- references a `DigitalOceanVpc` resource named `my-vpc`; rename it to your VPC resource, or replace the block with `value: <uuid>` for an unmanaged VPC. Brokers should never be publicly reachable.
- **Node size** (`sizeSlug: db-s-2vcpu-2gb`) -- entry sizing; Kafka throughput scales primarily with disk and network, so move up sizes before adding partitions.

Topics and schema-registry entries are managed outside this cluster resource.

## Related Presets

- **01-postgresql-ha** -- Use for the relational system of record beside the stream
- **05-opensearch** -- A common downstream consumer for log/search pipelines
