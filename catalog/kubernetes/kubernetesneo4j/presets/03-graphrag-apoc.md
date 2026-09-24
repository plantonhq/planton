# GraphRAG APOC preset

A Neo4j tuned for knowledge-graph and agent-memory workloads: the
APOC procedure library activated at startup (it ships inside the
official image — no download), apoc.conf arms enabled for the
import/export and trigger procedures GraphRAG pipelines lean on, the
procedure sandbox opened exactly to `apoc.*`, a page-cache-weighted
memory split for traversal-heavy reads, and heap dumps on
out-of-memory for diagnosing runaway queries.

Do not widen the procedure allowlist beyond `apoc.*`, and think twice
before enabling the file arms on a multi-tenant server — they let
procedures read and write the server's filesystem, which is exactly
why they are off by default; `use_neo4j_config` keeps that confined
to the pod's import directory. This preset is also still a dev-grade
posture on sizing (its credential is module-minted, the same as
production's default): for a production GraphRAG store, take the
production preset's storage class and sizing, and carry this preset's
`apoc_config`, `config` and `helm_values` blocks over.

Change first: the memory split if your
graph outgrows the 4Gi container — keep the page cache large relative
to heap; traversals live or die on graph data staying in memory.

See [03-graphrag-apoc.yaml](./03-graphrag-apoc.yaml) for the
manifest.
