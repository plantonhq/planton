# CloudflareWorker guide

A Worker is a script that runs on Cloudflare's edge, plus everything that hangs off it: bindings, routes, schedules, logs, and Durable Object migrations.

## Pick a source

- **Inline `content`** for a small ES module. The minimal and observability E2E scenarios use this. It needs nothing but the Cloudflare API token.
- **`r2Bundle`** for a CI-built artifact. `bucket` is a CloudflareR2Bucket reference (or a literal name); the module fetches the object and deploys it.
- **`assets`** alone for a static site. Combine it with a script for a full-stack app.

## An R2 bundle needs the connection's R2 keys

Both engines read the bundle through R2's S3-compatible API, and that API is signed with an R2 access key pair, never with the Cloudflare API token. Give the Cloudflare connection an R2 key pair with Object Read on the bundle bucket; Planton hands it to Pulumi through the provider config and to OpenTofu as `TF_VAR_r2_access_key_id` / `TF_VAR_r2_secret_access_key`. Binding a Worker to a bucket (`r2Buckets`) needs no keys at all — only building the script from one does.

The bundle's declared Content-Type does not matter: the modules read the object's raw bytes, so a file uploaded as `application/javascript` deploys exactly as one uploaded as `text/javascript`. An empty object stops the deploy with an error naming it, instead of shipping a Worker with no code.

`bodyPart` marks service-worker syntax. `mainModule` marks ES-module syntax. Do not set both.

Leave `mainModule` unset and both engines default it to `index.js`. That default is load-bearing: an upload with no main module is treated by Cloudflare as a legacy service-worker script, and ES-module code then fails deploy with `Uncaught SyntaxError: Unexpected token 'export'` (measured live).

## Bindings are grouped by type

The provider wants one discriminated `bindings` array. The spec uses the wrangler.toml grain — a typed list per kind — and both engines flatten. Cross-resource bindings take a literal or a `valueFrom` reference.

When Planton has no kind yet (Vectorize, Analytics Engine, Workflows, Pipelines, Secrets Store, mTLS, AI Search, Images, Media, VPC network id), the id is a literal. When a kind exists, the field is a foreign key: Durable Object `scriptName`, tail consumers, R2 bundle bucket, dispatch outbound worker, VPC `tunnelId`.

`tailConsumers` names Workers that consume *this* Worker's logs. `tailConsumerBindings` is a different thing — a `type=tail_consumer` binding the script can call.

## Migrations are one-shot

`migrations` creates, renames, transfers, or deletes Durable Object classes on this upload. Cloudflare rejects a second apply of the same `newTag`. Do not put a migration in an idempotent live scenario. Author it, plan it (`e2e/migrations-plan.yaml`), prove it later by hand if needed.

## Observability

`enabled` + `headSamplingRate` is the original pair. Nested `logs` and `traces` add destinations, invocation logs, persist, and (on traces) `propagationPolicy`. The observability-json E2E scenario covers the nested tree.

`propagationPolicy` is feature-gated: on an account without the trace-propagation feature, Cloudflare rejects the whole script upload with 403 code 100342 (measured live). Set it only when the account carries the feature.

## Importing an existing Worker

The provider's import restores only the Worker's identity and content — `mainModule`, bindings, compatibility flags, and tail consumers do not survive it. That is expected, not a broken import: run one apply after importing and those attributes converge in place from your configuration without touching the deployed code.

## What Pulumi cannot send yet

Pulumi's Cloudflare SDK v6.17.0 has no inputs for `cacheOptions`, `exports`, or `packageDependencies`. Tofu honors them. Pulumi logs a PARITY-EXCEPTION and skips them. Rate-limit `mitigationTimeout` and traces `propagationPolicy` have the same gap. Never `reserved` those fields — the proto stays complete; the lagging engine catches up on the next SDK bump.
