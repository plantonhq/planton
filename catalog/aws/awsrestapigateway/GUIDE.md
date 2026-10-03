# AwsRestApiGateway — Kind Guide

Authored operational judgment for the REST API Gateway component: the
design decisions behind the spec's shape, and what to know before
running REST APIs in production.

## Design decisions

- **Typed routes XOR OpenAPI.** AWS accepts both; mixing them silently
  overwrites. The spec enforces exactly one so a stray `routes` block
  cannot clobber an imported document.
- **The resource tree is derived, not declared.** Paths like
  `/users/{id}/orders` become nested API Gateway resources (max five
  segments). Authors write the HTTP surface; the modules own the
  parent/child wiring.
- **Explicit deployment, hashed from the definition.** REST APIs do
  not auto-deploy. The modules create one deployment whose trigger is
  a hash of the full API definition, so every spec change redeploys —
  the declarative behavior a catalog kind owes its users.
- **One stage.** Planton resources are already environment-scoped.
  Canary traffic shifting needs two live deployments and is a deploy
  workflow, not a resource field.
- **Satellites are named and referenced.** Authorizers, models, and
  validators live on the spec and routes point at them by name, so
  the same TOKEN authorizer can guard many methods without duplication.
- **Account-level CloudWatch logging is a region singleton.** Stage
  access logs need no account role; method-level execution logging
  (`method_settings.logging_level`) does. That role is not modeled
  here.

## Running REST APIs in production

- **Prefer typed routes for day-to-day APIs.** OpenAPI import is the
  right tool when the contract already lives in a document; typed
  routes are the right tool when Planton is the source of truth.
- **MOCK integrations are first-class.** Use them for health checks
  and contract stubs — they cost nothing and prove the tree without a
  backend.
- **Keep TOKEN authorizers on a short TTL.** `result_ttl_seconds`
  caches the Lambda's decision; a stolen token lives that long.
- **Resource policies are the PRIVATE-endpoint contract.** An
  `endpoint_configuration.type` of PRIVATE without a policy that
  admits the VPC endpoint is an API nobody can call.
- **Client certificates are for backend trust, not caller auth.**
  Generate one when the HTTP backend must verify that traffic came
  through this API; distribute the PEM to the backend trust store.
- **Enabling the stage cache can block an apply for up to 90 minutes.**
  AWS provisions a dedicated cache cluster when `cache_cluster` turns
  on or resizes, and the apply waits for it. Plan cache changes for a
  window where a long apply is acceptable — nothing is wrong, AWS is
  just slow.
- **OpenAPI-body applies log a reconciliation pass — it is not drift.**
  With an `openapi.body`, AWS's import wipes settings configured
  outside the document (description, endpoint settings, policy), and
  the apply re-applies them in a third pass. The extra update in the
  apply log is expected.
- **Many-response APIs apply serially.** AWS rejects concurrent
  method-response writes on one API, so the engines serialize them
  (with retries). A large route/response surface makes applies slower,
  not flakier.
- **Compression cannot be turned off by removing the field.** Once
  `minimum_compression_size` is set, unsetting it keeps compression on
  (AWS treats absent as "no change"). Set it to `-1` to explicitly
  disable compression again.
- **Adopting an existing API re-deploys it once.** The deployment's redeploy-on-change trigger is engine-side metadata AWS never stores, so an import cannot recover it (upstream documents this). The first reconcile after adopting an existing REST API therefore mints a fresh deployment from the adopted definition and repoints the stage — a behavioral no-op when the definitions match. The prior deployment stays in the API's deployment history until the API is deleted.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
