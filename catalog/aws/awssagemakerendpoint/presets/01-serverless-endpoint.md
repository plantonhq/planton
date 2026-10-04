# Serverless Endpoint

This preset serves one model on serverless compute — SageMaker scales
capacity with traffic and bills per inference, so an idle endpoint
incurs no charge at all. The start-cheap shape for new and spiky
workloads; the verified figure lives in the kind's generated
estimate at `catalog/_pricing/estimates/awssagemakerendpoint.yaml`.

## When to Use

- The first endpoint for a model, before traffic patterns are known
- Intermittent or spiky inference where paying for idle instances
  stings

## What You Get

- A single serverless variant (`AllTraffic` — AWS's console convention
  for one variant) processing up to 10 concurrent invocations in
  2 GB invocation environments (CPU scales with memory)
- No capacity math, no instance selection, no idle cost

## Customize

- Raise `maxConcurrency` (up to 200) and `memorySizeMb` (up to 6144)
  as traffic and model size grow
- Add `provisionedConcurrency` to hold pre-warmed capacity against
  cold starts — billed while provisioned, and never above
  `maxConcurrency`
- Graduating to dedicated instances means swapping `serverless` for
  `instanceType` and friends — a configuration roll the modules handle
  by minting a new configuration and repointing the endpoint
