# pkg/manifestgraph: one home for treating a manifest set as a dependency graph

## What changed

- **A new package, `pkg/manifestgraph`, owns everything about treating a
  SET of catalog object manifests as a dependency graph**: node identity
  (`(kind, slug, env)` — explicit `metadata.slug` passes through, else the
  slug derives from the name by the platform's slug rules, ported and pinned
  by test), reference collection (one walker across singular, repeated,
  nested, AND map-typed containers, with an in-place substitution seam),
  the strict reference rules (effective kind from `default_kind`, annotated
  field path only when kinds match, composition-key override protection,
  proto-surface path validity), edge derivation from THREE sources
  (`valueFrom` references by their EFFECTIVE kind — annotation defaults
  materialize before ordering; explicit `metadata.relationships`; literal
  namespace placement minting derived targets), classification
  (in-set / derived / external-valueFrom / env-external / external-
  relationship — severity is deliberately the consumer's policy),
  deterministic topological ordering (Kahn's, cycle named as a chain), and
  output-fed resolution rewriting in-set references to literals.

- **Three existing implementations converge on it.** Chart validation
  (`pkg/infrachart`) delegates its walker, reference rules, and cycle
  detection (its `refs.go` and `dag.go` die); the E2E harness's deploy-time
  resolver and offline fixture-integrity checker both migrate off their
  private spec-tree walker — whose map-typed-field blind spot dies with it —
  while the harness's deliberate leniencies (the sole-instance name
  fallback, unresolved-target tolerance) stay quarantined in the harness
  callers as documented policy, never in the shared rules.

- **The behavior corpus lands** (`pkg/manifestgraph/testdata/corpus/`):
  fourteen scenarios, each pinning one ordering/classification semantics
  point with a committed golden — annotation-riding references order like
  explicit ones, map-typed references form edges, relationships order
  without any `valueFrom`, literal namespaces edge when deployed and derive
  when not, names with spaces join their nodes through the one slug
  derivation, cycles name their chain, and every external class is
  distinguished. The goldens are a contract for every lane that orders
  manifests; regeneration is a reviewed semantics change
  (`PLANTON_REGEN_MANIFESTGRAPH_GOLDENS=1`).

- **Ride-alongs**: the E2E prerequisite chain's path fixtures now split
  multi-document files before loading (the loader's single-document
  contract; documents must agree on kind), and `pkg/infrachart`'s test
  fixtures catch up with the `_test` kind's v1alpha2 migration (five tests
  had been failing at HEAD since the fixture kind moved; verified
  pre-existing against a pristine checkout).

## Why

Ordering and resolving a set of manifests had three divergent
implementations (chart validation's strict rules, the E2E harness's lenient
mirror, and none at all for maps). One tested home means the semantics can
never drift between the lanes that validate, test, and deploy the same
trees — and the committed corpus turns any future drift into a failing test
instead of a production surprise.

## Verification

- New suites: the fourteen-scenario corpus goldens, slug-derivation pinning
  (platform-rule parity cases), substitution across singular/list/map
  containers, unresolved-vs-missing-output classification, topological-order
  properties (every edge respected, deterministic tie-break), cycle naming.
- `go build`/`go vet`/`go test` green across `pkg/manifestgraph`,
  `pkg/infrachart` (its full suite, now green again), and
  `e2e/framework/...` (fixture-integrity gate: zero baseline churn under
  the map-safe walker). Gazelle BUILD regeneration included.
