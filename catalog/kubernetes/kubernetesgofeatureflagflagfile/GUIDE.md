# KubernetesGoFeatureFlagFlagFile Guide

The judgment this guide carries: flags are data with their own owners and
lifecycle -- keep them in flag files, not in the relay, and write targeting
that stays true as organizations come and go.

The rendered file is GO Feature Flag's flag format as relay v1.56.0
reads it; every flag field of that format is a typed field, except a
scheduled step's `bucketingKey` and `metadata`, which the relay's step
merge ignores.

## When to use it (and when not)

Use one flag file per owner. A KubernetesGoFeatureFlag relay reads any
number of them through `configMap` retrievers, and a flag flip touches
only the flag file: the relay is never re-applied and serves the change
within its polling interval, with no restart. Choose a repository, bucket
or database retriever on the relay instead when the flags must live
outside the cluster or be shared across clusters -- you then lose the
plan-time validation this kind gives. For flagd, the equivalent is
KubernetesFlagdFlagFile; the two flag formats are not interchangeable.

## Conventions and gotchas

- **Structurally broken flags never ship.** The spec enforces what the relay would
  otherwise reject at load time (one value type across variations, a
  default rule that resolves, carries no query and is never disabled, a
  query on every enabled targeting rule, rules naming defined variations,
  rollouts that ramp forward, unique rule names, real dates). The relay drops a flag that fails those
  checks and only logs an error, so every evaluation of it silently returns
  the caller's code default -- easy to miss, which is why the checks live
  here. Query syntax is the exception: the relay parses each query on load,
  so test a new query against the relay (an OFREP evaluation with a context
  that should match) before relying on it.
- **A progressive rollout moves callers from one variation to another.**
  Before the initial date everyone gets the initial variation; between the
  dates the end `percentage` share gets the end variation; an empty or zero
  end share means 100. Name different variations at the two ends -- the
  same variation at both serves it to everyone from day one.
- **Target stable identifiers.** Rules read the evaluation context the
  caller sends. Target attributes whose values are never reused (an
  organization identifier that is never reissued) so a rule written once
  cannot later match someone else.
- **Rule names matter when you schedule.** A `scheduledRollout` step is
  merged into the flag at its date: targeting rules merge by `name`, field
  by field (a step naming an unknown rule adds a new one), the default rule
  merges field by field, and variations are added or replaced by name.
  Fields a step leaves empty keep their value. Steps carry no bucketing key
  or metadata -- those belong to the flag.
- **Rendered as JSON.** The ConfigMap holds one JSON document under the
  default key `flags.goff.yaml`; JSON is YAML, so the relay's default file
  format reads it. If you change `key`, change the relay retriever's `key`
  to match.
- **1 MiB ceiling.** A ConfigMap holds at most 1 MiB; split large flag sets
  across flag files by owner.
- **Retire in order.** Remove the code that evaluates a flag, then delete
  the flag from the file. Deleting first makes every evaluation return the
  caller's code default -- usually the off state, which is safe for a
  release flag but not for a configuration flag.

## On the diagram

Each flag file renders as its own node with a reference edge into the
relay that reads it -- ownership of switches is visible on the
architecture, separate from the engine.

## Pairs well with

- KubernetesGoFeatureFlag -- the relay; its `configMap` retriever takes
  this resource's `config_map_name` and `key` outputs.
- KubernetesNamespace -- the namespace the ConfigMap renders into.

Presets (Release Flags, Progressive Rollout) ship in the release's
`presets.zip` and in the repository. The full field reference is
[reference.md](v1alpha1/reference.md).
