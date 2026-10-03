# AwsBedrockFlow — Kind Guide

Authored operational judgment for the Bedrock flow component: the design
decisions behind the spec's shape, and what to know before running flows
in production.

## Design decisions

- **The node union keeps an explicit `type` plus derived members.** Nine
  node classes carry configuration (agent, prompt, knowledge_base,
  lambda_function, lex, condition, inline_code, retrieval, storage) and
  the spec pairs each with exactly its arm (CEL-guarded). The structural
  classes (Input, Output, Iterator, Collector) carry no arm — the
  modules render their EMPTY AWS union members from `type`, because AWS
  requires the member even when it holds nothing.
- **The Loop family is typed but not configurable at the pin.** AWS
  accepts Loop/LoopInput/LoopController nodes, but the pinned provider
  cannot express their configuration (an upstream gap its own source
  marks TODO). The spec admits the types; loop bodies wait on the
  provider.
- **Connection types are derived.** A connection sets `data` XOR
  `conditional`; the modules derive AWS's Data/Conditional discriminator.
- **The inline prompt tree mirrors AwsBedrockPrompt.** Upstream shares
  the same Go models between the prompt resource and the flow's inline
  prompt node; the two components' specs, modules, and parity manifests
  mirror each other divergence-for-divergence — change them together.
- **One-value vocabularies are module constants**: inline-code language
  (Python_3), cache-point type (default), and the retrieval/storage
  service (S3).

## Running flows in production

- **Build minimal, extend validated.** AWS's server-side graph
  validation is strict and its error classes are precise — start from
  the Input→Prompt→Output skeleton, deploy, then grow the graph a node
  at a time rather than debugging a twenty-node graph's first create.
- **The execution role aggregates every node's permissions.** A flow
  invoking a model, an agent alias, a knowledge base, and a Lambda needs
  all four invoke permissions on ONE role — audit it when adding node
  classes.
- **Flows do not auto-prepare.** The module creates the DRAFT
  definition; preparing (compiling for serving) and versioning happen at
  invocation setup. A NotPrepared status after deploy is healthy.
- **Socket types are contracts.** A String output feeding an
  Object-typed input fails validation at create — align the node IO
  types before reaching for expressions.
- **Importing an existing flow shows a one-time inference-float
  reconcile.** Bedrock stores node `temperature`/`top_p` as 32-bit
  floats, so a value that is not float32-exact (0.2, 0.9) reads back
  slightly widened on import, and the first plan proposes an in-place
  update back to your manifest's value. Applying it is a server-side
  no-op and the plan is clean thereafter. Normal deploys never see this
  — state keeps the manifest's value. Same class as AwsBedrockPrompt.

## Cost model

Creating, updating, and deleting flows is free. Costs are the nodes':
model tokens on prompt/knowledge-base nodes, agent invocations, Lambda
duration. An idle flow costs nothing. Verified per-preset figures live
in the generated estimate at
`catalog/_pricing/estimates/awsbedrockflow.yaml`, computed from the
pinned, source-dated price book.
