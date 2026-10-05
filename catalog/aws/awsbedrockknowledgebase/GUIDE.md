# AwsBedrockKnowledgeBase — Kind Guide

Authored operational judgment for the Bedrock knowledge base kind:
the design decisions behind the spec's shape, and what to know before
running knowledge bases in production.

## Design decisions

- **The type discriminator is derived, never asked for.** The provider
  carries `knowledge_base_configuration.type` plus one config block per
  type; the spec carries exactly one of `vector`/`managed`/`kendra`/`sql`
  (CEL-guarded) and the modules derive AWS's discriminator. A leaf that
  must agree with structure is drift surface, not configuration — the
  same rule derives the storage backend, data-source connector types, and
  the managed type's `embedding_model_type` (CUSTOM exactly when an
  embedding-model ARN is brought, MANAGED otherwise). That last leaf is
  also always *sent*: the provider marks it Optional+Computed, and
  leaving it unknown lets AWS create the knowledge base then fails the
  Pulumi apply with "unexpected unknown property value", stranding the
  resource outside engine state (live-caught 2026-08-14).
- **`storage` pairs with `vector` and nothing else.** The CEL guard
  `has(vector) == has(storage)` encodes AWS's real contract: managed
  brings AWS's store, Kendra and SQL delegate storage entirely.
- **Data sources are folded, name-keyed satellites.** One lifecycle, one
  component: each entry's `name` is the for_each key on both engines,
  the key in the `data_source_ids` output map, AND the AWS-side data
  source name (renames replace — upstream marks the name ForceNew).
- **Fixed wrappers flatten to leaves.** The supplemental-data location
  (`storage_location { type = "S3", s3_location { uri } }`) is one URI
  leaf; the crawler filter tree (`filter_configuration { type = "PATTERN",
  pattern_object_filter { filters } }`) is one `filters` list per SaaS
  connector. Every flattening is recorded in the parity manifest.
- **Two-value states are bools.** The managed connector's
  ENABLED/DISABLED extraction and deletion-protection states read as
  `audio: true`, `enabled: true` — the modules send the constants.
- **Chunking keeps an explicit `strategy` leaf** (unlike the derived
  discriminators) because `NONE` — treat each document as one chunk — is
  a real strategy with no configuration block to derive from.

## Running knowledge bases in production

- **Pick the store by operational appetite.** S3 Vectors is pay-per-use
  with no infrastructure (the cheapest start); OpenSearch Serverless
  gives sub-second recall at OCU cost but requires index management
  (create the vector index before the knowledge base — AWS will not);
  managed hands the decision to AWS. The SaaS backends (Pinecone, Atlas,
  Redis) carry their credentials as Secrets Manager references — never
  paste keys into specs.
- **Dimensions are a contract.** Titan Text Embeddings V2 supports
  256/512/1024; whatever you pick must equal the vector index's
  dimension. The failure surfaces at first ingestion sync.
- **Ingestion is a separate step.** Deploying the component creates the
  knowledge base and data sources; document syncs
  (`StartIngestionJob`) run from the console, CLI, or your pipeline.
  Live-proven 2026-08-14 on the VECTOR + S3 Vectors 256-dim pairing: a
  one-object `StartIngestionJob` reached COMPLETE (1 scanned, 1 indexed,
  0 failed) and the DELETE data-deletion policy still let destroy
  finish cleanly.
- **The role fails at create, not at query.** AWS validates assume-role
  and store access at CreateKnowledgeBase with a propagation retry
  window (~2 minutes) — a missing s3vectors/aoss permission is a deploy
  failure with the reason in `failure_reasons`.
- **Replacement means re-ingestion.** Budget sync time when changing
  chunking or storage: the old store's vectors are dropped (DELETE
  policy) and every source re-syncs into the new shape.

## Cost model

Creating and deleting knowledge bases is free. Ingestion bills embedding
tokens per sync; queries bill retrieval (and generation when a model
generates over the results). The managed type additionally bills for the
AWS-run store; S3 Vectors bills storage and per-query request units.
