# Dev single instance preset

The smallest declarable Neo4j: community edition, the admin password
minted by the module (materialized as the `<name>-auth` Kubernetes
Secret, never in rendered values; the `password_secret` output names
it), a 10Gi data volume on the cluster's default StorageClass,
and resources at the chart's own floor — 500m CPU / 2Gi memory, below
which the chart refuses to install at all. For developers who need a
real Cypher endpoint, and for the first iteration of a knowledge
graph or agent-memory experiment.

Do not shrink this preset to save resources — there is no smaller
Neo4j; the floor is the chart's, not this preset's. The credential
needs no attention: it is generated, stable across re-applies, and
read by clients from the Secret the outputs name. In-cluster clients connect via the default Service in the
stack outputs (bolt 7687); nothing is exposed outside the cluster.

Change first: `data_volume.size` if
your graph will outgrow 10Gi — growing a PVC later depends on the
StorageClass allowing expansion.

See [01-dev-single-instance.yaml](./01-dev-single-instance.yaml) for
the manifest.
