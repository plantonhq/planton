# Dev single node preset

The smallest declarable Qdrant: one node, no authentication (the
upstream default), a 5Gi data volume on the cluster's default
StorageClass, and the chart's own resource defaults. For developers
who need a real vector endpoint — the first iteration of a RAG
pipeline, a semantic-search prototype, an embedding experiment —
without any production ceremony.

Know what "one node" means here: distributed mode is always on, so
this is a one-member Raft cluster, not a special standalone mode.
Growing later is a `replicas` change with no migration — the new pods
join the existing consensus. The unauthenticated posture is the part
that must NOT outlive dev: the listeners accept any request, so this
preset belongs strictly inside a private cluster's namespace
boundary. In-cluster clients connect via the outputs — gRPC
6334 (what SDKs default to) or REST 6333; nothing is exposed outside
the cluster.

Change first: declare `api_key: {generate: true}` the moment anyone
but you can reach the namespace, then `storage.size` if your
embeddings will outgrow 5Gi — growing a PVC later depends on the
StorageClass allowing expansion. Memory is the real capacity bound
(Qdrant keeps hot segments and indexes in RAM); add a `resources`
block when the corpus stops being a toy.

See [01-dev-single-node.yaml](./01-dev-single-node.yaml) for the
manifest.
