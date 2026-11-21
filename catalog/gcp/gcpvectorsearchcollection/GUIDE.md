# GcpVectorSearchCollection Guide

The judgment this guide protects: a collection is a schema, and its
indexes are frozen decisions. Get the vector fields and the index settings
right before data arrives, because the schema shapes every object written
and an index changes only by being rebuilt.

## The collection declares, applications write

This block creates the store and its indexes; it never writes an object.
Applications write data objects (and, for fields without an embedding
config, their vectors) through the Vector Search API or its SDKs, and run
similarity searches against an index. Field names in `dataSchema` and
`vectorSchemas` are the names those objects carry, so change them as you
would a database schema -- deliberately, with the writers. Vector fields
update in place: emptying `vectorSchemas` asks Google to drop every vector
field, and every index's `indexField` must name a declared field.

## Who computes the vectors

A dense field with `vertexEmbeddingConfig` is embedded by Vertex AI on
every write: the `textTemplate` assembles text from the object's fields
(`"Title: {title} ---- Body: {body}"`), the named model embeds it, and
`dimensions` must match what that model produces. This is the simplest
path when you have text. Without an embedding config, your pipeline writes
the vectors itself -- the path for models Vertex does not host or vectors
you already have. A `sparseVector: true` field holds lexical vectors
(SPLADE, BM25-style) for hybrid search; Google's sparse field has no
settings, so the declaration is a flag.

## Indexes are mostly immutable

An index's display name, description, labels, and replica bounds change in
place. Every other index setting -- the field, the metric, the ScaNN norm,
the filter and store fields, the infrastructure mode -- is fixed at
creation. Changing one replaces the index, and the replacement is rebuilt
from the collection's data while searches against the old one are
interrupted. Removing a metric, norm, or dedicated block sends nothing:
Google keeps what the index has, nodes included. The collection's labels
and the platform attribution labels land on every index and win on a key
conflict. Choose `COSINE_DISTANCE` with `UNIT_L2_NORM` for embeddings
that are not already normalized; `DOT_PRODUCT` (Google's default) for
ones that are. Google builds one index at a time per collection and refuses a
second operation while one runs (409, "unable to queue the operation"):
the Pulumi module creates the indexes one after another, while Terraform
starts them together and Google refuses all but one, so a Terraform apply
of a collection with several indexes needs a second apply for the rest. A `STORAGE_OPTIMIZED` index takes only `DOT_PRODUCT` and
no ScaNN options such as the norm (Google refuses both): unit-normalize the
vectors before ingestion there to get cosine similarity. Push into `filterFields` the fields searches filter on and
into `storeFields` the ones results should return inline, so a hit needs
no second lookup.

## Pooled or dedicated

By default an index serves from Google's pooled infrastructure and bills
per query and per GiB indexed. `dedicatedInfrastructure` gives it its own
nodes -- `PERFORMANCE_OPTIMIZED` for latency, `STORAGE_OPTIMIZED` for cost
per vector -- autoscaling between `minReplicaCount` and `maxReplicaCount`
and billing per node-hour from creation, queries or not. The minimum is
the committed spend, and both bounds change in place. Set
`maxReplicaCount`: unset, its default is the greater of the minimum and 2
under Terraform (Google's v1 API) but 5 under Pulumi (the v1beta API).

## Encryption and destroy

`kmsKeyName` encrypts the collection and its indexes under your key; it
is immutable, so decide before the first object. `deletionPolicy` fans to
every index: `DELETE` removes indexes then the collection, data included;
`PREVENT` guards a collection holding data you cannot regenerate.
