# AwsSagemakerEndpoint — Kind Guide

Authored operational judgment for the SageMaker endpoint kind:
the design decisions behind the spec's shape, and what to know before
running real-time inference in production.

## Design decisions

- **The endpoint's AWS name derives from `metadata.name`.** Clients
  invoke the endpoint by that name; it never changes across
  configuration rolls.
- **The endpoint configuration is folded into the endpoint.** Upstream
  they are two resources, but the configuration is immutable (every
  argument ForceNew) while the endpoint's pointer to it updates in
  place. Keeping them apart would push AWS's roll-and-repoint
  choreography onto every author; folding them lets the modules own it
  — configurations are name-suffixed (`<name>-cfg-…`) and created
  before the old one is destroyed, UpdateEndpoint repoints, and only
  then does the predecessor disappear. The endpoint never references a
  deleted configuration.
- **Variant names default deterministically per position**
  (`variant-0`, `variant-1`, …; `shadow-variant-0` on the shadow
  side). The provider would otherwise mint a random name per plan,
  forcing a configuration roll on every apply.
- **Serverless and instance settings are exclusive per variant.** AWS's
  ServerlessConfig contract — a serverless variant rejects
  `instance_type` and every instance-tuning field; the spec enforces
  the split at manifest time.
- **Shadow testing is exactly one variant on each side.** AWS's own
  shape — the spec rejects other combinations before the API does.
- **Exactly one deployment strategy.** When `deployment` is set, it is
  `blue_green` or `rolling`, never both; canary and linear step sizes
  pair with their routing types (AWS's rules, spec-enforced).

## Running endpoints in production

- **Start serverless, graduate to instances.** A serverless variant
  costs nothing idle and needs no capacity math; move to
  instance-backed variants when latency floors, GPU needs, or
  provisioned-concurrency costs say so.
- **Request instance quotas BEFORE the first instance-backed deploy.**
  On a fresh AWS account, the per-type "for endpoint usage" Service
  Quota defaults to ZERO for nearly every instance family — ml.m5.large
  and ml.c6i.large included — and `CreateEndpoint` fails with
  ResourceLimitExceeded until Service Quotas grants an increase
  (minutes to days). The entry-level exceptions with a default of 2 are
  ml.t2.* (x86) and ml.m6g.large (Graviton — your container image must
  be arm64). Serverless variants need no per-type quota (defaults: 5
  serverless endpoints, 10 total concurrency per region) — which is
  half the reason to start serverless. Probe first:
  `aws service-quotas list-service-quotas --service-code sagemaker`.
  Size the quota for rollouts, not steady state: a rolling deploy
  transiently runs old + new batches together, and blue/green doubles
  the whole fleet.
- **Shape the crossing before you need it.** Configuration rolls are
  routine (every capacity change is one) — give production endpoints a
  `deployment` policy with `auto_rollback_alarm_names` so a bad model
  version backs itself out instead of paging you.
- **Weight, don't flip.** New model versions ride a second variant at
  a small `initial_variant_weight`; a weight of 0 keeps a variant
  deployed but takes no traffic — an instant rollback target.
- **Mind the KMS caveat.** `kms_key_arn` encrypts ML storage volumes,
  but nitro-local-storage families (ml.g5/g6/p4d/p5 and similar)
  encrypt locally by default and reject a custom key — as do
  serverless-only endpoints.
- **Capture from day one.** `data_capture` is the Model Monitor feed;
  flipping it on later is itself a configuration roll, so sample a low
  percentage early rather than retrofitting under incident pressure.
- **Budget bake time.** Blue/green provisions a full parallel fleet
  and holds it through `wait_interval_seconds` per step (plus
  `termination_wait_seconds` after the shift); rolling replaces
  batches in place with no parallel-fleet cost — pick by budget and
  blast radius.
- **Rolling needs a fleet.** AWS rejects a rolling policy on a
  single-instance endpoint at create AND update ("Cannot update
  endpoint with single instance using RollingUpdatePolicy" —
  live-verified; the spec front-loads the single-variant case). A
  one-instance endpoint uses blue/green or omits `deployment`.
- **An endpoint is only as healthy as its model's artifacts.** A model
  whose container has nothing to load (no `model_data_url` /
  `model_data_source`) passes CreateModel but can never answer the
  endpoint's ping health checks — the endpoint parks at `Failed`
  ("Unable to successfully stand up your model within the allotted
  180 second timeout", live-verified on a serverless variant). And a
  FAILED create can strand the endpoint object outside your IaC state:
  delete it explicitly before retrying the same name.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
