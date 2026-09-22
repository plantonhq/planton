# GcpVertexAiRagEngineConfig Guide

The judgment this guide protects: this block does not create anything. It
sets the tier of a database Google already owns for your project and
location, and the one way it can hurt you is by turning that database off
-- which deletes every corpus in it. Choose the tier deliberately and
choose the destroy behavior more deliberately still.

## A singleton, updated in place

Google creates one RAG Engine configuration per project per location the
first time RAG Engine is used there. There is no "create": applying this
block PATCHes the tier, so a manifest applied over an already-configured
location changes the tier rather than failing, a re-run cannot hit
"already exists", and two blocks for one location overwrite each other.
Declare exactly one per location.

## Which tier

`BASIC` is the experimentation tier -- small corpora, latency-insensitive
retrieval, prototypes -- and it is also all you need when your corpora
live in an external vector database (Vector Search, Pinecone, Weaviate),
because then RAG Engine's managed database holds nothing. `SCALED` buys
production performance and autoscaling for large or hot corpora; raising
`BASIC` to `SCALED` is a live upgrade. `UNPROVISIONED` disables the
managed database and DELETES its data; billing stops, and the data cannot
be recovered.

## What destroy means

Because the configuration cannot be deleted, the provider's destroy
PATCHes the location to `UNPROVISIONED` -- data loss for every corpus in
that location. The default `deletionPolicy` (`DELETE`) does exactly that.
For any location whose corpora matter, set `ABANDON` (the block stops
managing the tier; the tier and data stay) or `PREVENT` (destroy fails
until you change your mind). The Basic preset ships with `ABANDON` and the
Scaled preset with `PREVENT` for this reason.

## Where this sits in a chart

Put this block before the agents and applications that ingest into RAG
corpora in the location. Nothing references its outputs; it just has to
exist, at the right tier, before the first corpus is created there.
