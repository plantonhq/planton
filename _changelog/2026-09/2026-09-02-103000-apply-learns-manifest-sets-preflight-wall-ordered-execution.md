# `planton apply` learns manifest sets: the preflight wall, dependency-ordered execution, output-fed references

## What changed

- **`planton apply` deploys manifest SETS.** When the input holds more than
  one manifest — a multi-document file, a directory of manifests
  (`planton apply -f manifests/`, new), or a kustomize overlay that renders
  several resources — apply becomes the orchestrated lane: one preflight
  report, then the resources deploy sequentially in dependency order derived
  from their own `valueFrom` references, `metadata.relationships`, and
  literal namespace placement, with each node's captured outputs
  resolving the next nodes' references to literals before handoff. One
  manifest behaves exactly as before, byte for byte. This replaces the
  loud multi-document refusal that previously guarded the silent-truncation
  bug — multi-document input is now the feature, not the error.

- **The preflight wall verifies everything verifiable BEFORE the first IaC
  handoff, as one honest report — never fail-at-first.** Eight checks, in
  report order: every document loads as a known kind and passes schema
  validation (violations reported per field, across ALL documents at once);
  no two documents claim one `(kind, name, env)` identity; every `valueFrom`
  reference resolves to a node in the set (an external target refuses naming
  the field, the target, and the honest reason — no backend exists to
  discover it — while external `metadata.relationships` targets are stated
  assumptions, since they carry no value need); no value anywhere carries a
  `$var/` or `$secret/` prefix (these resolve only through a Planton
  backend; the refusal points at provider-native secret references);
  dependencies form an order (a cycle names its chain); the tofu/terraform
  AND pulumi binaries are on PATH (the pulumi check is new — it previously
  failed mid-run) and a catalog module exists for every kind, with published
  artifacts HEAD-probed so version skew degrades LOUDLY to a source-checkout
  warning instead of silently; every remote state backend is complete,
  collision-free, and REACHABLE — probed with the credentials in hand
  (bucket-level probes for s3/S3-compatible/gcs; pulumi object-store URLs
  through the same probes); and every required cloud provider's ambient
  credentials actually AUTHENTICATE (live identity calls for aws/gcp/azure;
  kube contexts verified against the kubeconfig; providers without a probe
  are stated assumptions, never silence). The report renders every check
  line-itemed with a one-line verdict, and the refusal sentences follow one
  grammatical shape: what failed, where, why, what fixes it.

- **Distinct exit codes tell CI the truth**: 0 success, 1 deploy failure
  (state may have advanced; re-running the same command re-applies completed
  nodes as no-ops and continues), 2 refused at preflight (nothing was handed
  to an IaC engine). On a node failure the run stops and reports every
  node's status — succeeded (outputs captured), failed (the engine's error
  verbatim, never paraphrased), never started — and the honest recovery
  sentence.

- **Every tofu/terraform node executes in its own stable, identity-keyed
  workspace** (`~/.planton/setdeploy/<env>/<kind>/<slug>/` — a per-node copy
  of the module). Module cache directories are shared per kind, so running
  set nodes in them directly would collide same-kind local state files and
  leak one node's backend config into the next. The identity-keyed copy
  makes local state SAFE for sets (it persists per node across runs, so
  re-running is a true recovery story — the report states each node's state
  location, with a notice recommending a remote backend for CI), gives
  remote-backend nodes clean per-node backend configs, and shares one
  provider plugin cache so fresh workspaces hardlink providers instead of
  re-downloading. Pulumi nodes need no copy: their state lives behind the
  backend URL, addressed per node by the `pulumi.planton.dev/stack.fqdn`
  annotation (two nodes sharing a stack, or two tofu nodes sharing a state
  key, is a preflight refusal naming both).

- **A sensitive output resolved into a downstream manifest warns before the
  value leaves our hands**: the resolved literal becomes plain config the
  engine may echo in its diff — in CI, that is the log — so the executor
  names the field and points at provider-native secret references. Resolved
  manifests are written 0600 and always removed after handoff.

- **Single-resource flags refuse set input with the fix named** (`--set`,
  `--module-dir`, `--local-module`, `--stack`, `--backend-key`,
  `--iac-input`, `--clipboard`); set-wide flags stay legal
  (`--backend-type/bucket/region/endpoint`, `--backend-url`,
  `--module-version`, `--kube-context`, `--auto-approve`/`--yes`). Approval
  is ONE decision per set: `--auto-approve` (or `--yes`) in CI, one
  interactive question otherwise — never a hidden per-node prompt that hangs
  a runner. `plan` and `destroy` still take one document at a time; their
  refusal now says that apply deploys sets.

- **New package `pkg/setdeploy`** owns the wall and the loop, composing
  `pkg/manifestgraph` (semantics) with the engines' single-manifest run
  paths (execution). The library never prints and never exits: the wall
  returns a structured report, execution streams through an events seam, and
  every environment probe is an injected interface — so the logic is
  exhaustively tested with fakes and other consumers can render their own
  surface. The report renderer lives in `internal/cli/ui/preflightreport`,
  with the rendered text pinned by golden files
  (`PLANTON_REGEN_PREFLIGHT_GOLDENS=1` regenerates; a wording change is a
  deliberate diff, never an accident).

- **Ride-alongs**: `internal/manifest` gains `ViolationLines` (the same
  protovalidate run as `ValidateLoaded`, returning per-field lines instead
  of one styled banner, so N documents' failures aggregate into one report);
  the wall's missing-backend-field refusals name the annotation prefix of
  the node's OWN provisioner (`tofu.planton.dev/...` for tofu nodes — the
  generic validator hint always said `terraform.`).

## Why

The engine deployed one manifest at a time while real environments are sets
of resources that reference each other. Deploying a set by hand means
ordering by intuition and copying outputs between manifests — and a mistake
surfaces as an IaC stack trace twenty minutes in, in exactly the CI
environments where nobody is watching. The wall makes that defect class
structurally impossible for everything verifiable up front, and the executor
makes the manifests' own references do the ordering and wiring work they
already declare.

## The report, on a real run

A two-manifest set with an external reference and a `$secret/` value refuses
like this (abridged; exit code 2):

```
🧱 Preflight
  ✔ Manifests load and validate
     ✔ 2 of 2 documents load as known kinds and pass schema validation
  ✖ References resolve inside this set
     ✖ manifests/02-service.yaml: spec.annotated_ref: references TestCatalogKindGeneric "missing-resource" outside this set; the set does not deploy it — the value must come from a resource that already exists — no backend exists here to discover it; add its manifest to this set, or deploy connected
  ✖ No values require a Planton backend
     ✖ manifests/01-cache.yaml: spec.sensitive_string carries a $secret reference, which resolves only through a Planton backend — for runtime secrets use provider-native secret references (`planton secret snippet`), or deploy connected
  ✔ Dependencies form a deployable order
     ✔ deploy order: TestCatalogKindGeneric/cache@prod -> TestCatalogKindGeneric/service@prod

✖ preflight refused the deploy: 2 problems named above — nothing was handed to an IaC engine
```

The same set, fixed, deploys both resources through real tofu with the
producer's captured `id` output resolved into the consumer's config
(`annotated_ref = "tcrg-cache"` in the plan diff) and ends with masked
per-node output summaries and exit code 0.

## Verification

- `go build ./...` and `go vet ./cmd/... ./internal/... ./pkg/...` clean.
- `pkg/setdeploy` suite: the wall check by check with fake probes (schema
  aggregation, duplicate identity, external-reference refusal, `$secret`
  refusal, cycle chain, missing backend key naming the right annotation,
  state-key and stack collisions, local-state statement + CI notice, pulumi
  stack identity, probe refusal surfacing, module-fallback-as-warning, the
  `_test` provider needing no credentials); the executor with a fake
  deployer (order, output-fed resolution proven in the handoff manifest
  bytes, missing-output failure naming the field, stop-on-failure statuses,
  sensitive-resolution warning, refused-plan rejection); the backend-ref
  detector across singular/list/map/literal-arm containers with no
  mid-string false positives.
- **The offline-policy corpus test**: all fourteen `pkg/manifestgraph`
  behavior-corpus scenarios driven through the wall, pinning which ones a
  backendless deploy refuses and which deploy with stated assumptions.
- **The zero-cloud live proof**: a two-node `_test`-kind set through the
  REAL deploy path — live probes, identity-keyed workspaces, actual
  `tofu init`/`apply`, real `tofu output -json` capture, the producer's id
  resolved into the consumer — plus the re-run-as-no-ops recovery story and
  the persisted per-node state file, all asserted; requires only tofu on
  PATH.
- Rendered-report goldens committed (`internal/cli/ui/preflightreport/testdata/`).
- The full refusal and deploy journeys exercised against the built binary
  (the captures above), including exit codes 2 and 0.
- `internal/manifest` suite green after the `ViolationLines` seam; gazelle
  BUILD regeneration included.
