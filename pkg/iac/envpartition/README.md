# envpartition — deterministic environment partitioning

Assigns discovered cloud account resources to environments before any
grouping or mapping happens. Environment is part of a Planton resource's
identity (`org/env/kind/name`) and `value_from` references are env-scoped,
so the partition must be settled first — a wrong split poisons every
downstream reference and the resource graph itself.

## The model

A **rule** (`EnvironmentPartitionRule`, `iac.planton.dev/v1`) declares the
vocabulary; the **engine** applies it to resources with a strict precedence
ladder; the output is a per-resource **assignment with signal attribution**
plus an honest **unassigned** remainder. The engine never guesses: a
resource with no signal — or with contradictory signals — stays unassigned
with the contradiction flagged, and a human review gate owns the remainder.

```text
taught override > authoritative tag > name token > containment > unassigned
```

1. **Taught overrides** (tier 1) pin one resource to one environment — they
   are what a review gate's corrections become. The human's direct word
   beats everything, including a stale tag.
2. **Authoritative tags** (tier 2): when a resource carries a declared tag
   key, the tag's *value is the environment* — normalized through the
   vocabulary when it matches a declared token (`prd` → prod), accepted
   literally otherwise (`env: ops` names an "ops" environment). Explicit
   keys are trusted wholesale; the inference tiers never run for a tagged
   resource. The default vocabulary includes `planton.ai/environment` — the
   tag Planton itself writes on every deploy — so a Planton-deployed
   resource self-identifies on re-scan, and a user can deterministically
   force any resource into an environment by stamping that tag in their
   cloud console.
3. **Name tokens** (tier 3): whole tokens of *human-authored* name surfaces
   (a `Name` tag, a chosen queue/bucket/role name), matched against the
   declared vocabulary only. Opaque cloud identifiers are never tokenized
   (a random id can embed `prodfix`), matching is never substring, and an
   undeclared token never spawns an environment.
4. **Containment inheritance** (tier 4): a resource with no signal of its
   own inherits its container's environment (a NAT gateway through its
   subnet through its VPC — the fixpoint handles multi-hop chains). A
   resource whose own signal disagrees with its container keeps its own
   signal, flagged.

## The untaught and taught states differ only in data

`DefaultRule()` is an embedded `EnvironmentPartitionRule` document
([default-rule.yaml](default-rule.yaml)) — the kind dogfoods itself. When a
human confirms or corrects a proposed split at a review gate, the confirmed
rule (edited definitions + per-resource overrides) is persisted and replayed
deterministically on every re-scan and future discovery run: the user
teaches the rule **once**, never per resource.

## Two load-bearing properties

- **Determinism.** Same rule + same resources ⇒ same result, byte for byte.
  Re-scans must reproduce a confirmed split exactly; review surfaces diff
  against prior runs. Nothing here may iterate a map without imposing order
  (tags are scanned in the rule's declared key order).
- **Resources are partitioned; manifests aggregate.** The engine assigns
  environments to *resources*. A proposed kind manifest inherits the
  unique environment its claimed resources agree on; disagreement is a
  flagged conflict and an honest non-answer, never a guess.

## Consumers

- `pkg/iac/mappingeval/baseline` — the eval harness's deterministic
  reference proposer stamps `metadata.env` on proposed manifests from this
  engine's assignments (untaught default rule), and the harness grades the
  result as the partition scoring axis on suites that declare it.
- The account-import journey's scan stage (the same adapter, the taught
  rule) — the engine is product code; eval machinery depends on it, never
  the reverse.

The `awsscan` subpackage owns all AWS property-model knowledge (which
properties are name surfaces, which are containers), declared once so every
consumer reads the same signals the same way.
