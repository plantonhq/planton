# KubernetesFlagdFlagFile Guide

A flag file is where a feature flag's real decisions live -- its safe default, its targeting, who owns it. This guide is the judgment for writing flag files that flip safely and stay reviewable.

## When to use it (and when not)

Use **KubernetesFlagdFlagFile** for every flag you own and serve with a **KubernetesFlagd**. Keeping flags in their own resource means a flip edits only this file: the daemon is never re-applied, and permission to change flags can be granted without permission to change flagd. Reach for a remote source on the flagd spec (HTTP, gRPC, a bucket) only when another system publishes the flags.

The rendered file declares flagd's flag-definition schema v0 (`https://flagd.dev/schema/v0/flags.json`), the format flagd v0.17.0 reads; every field of that format is a typed field.

If the engine is **KubernetesGoFeatureFlag**, use **KubernetesGoFeatureFlagFlagFile** instead -- the two formats differ (JSONLogic here, a query language there) and are not interchangeable.

## Conventions and gotchas

- **The flip arrives on the kubelet's schedule.** flagd mounts this ConfigMap as a directory, so an edit reaches it when the kubelet syncs the volume -- typically one to two minutes -- with no restart. Plan rollouts in minutes, not seconds.
- **Same namespace as the daemon.** The ConfigMap must live in flagd's namespace: pod volumes cannot cross namespaces.
- **1 MiB ceiling.** A ConfigMap holds at most 1 MiB; split a large flag set into several flag files (one flagd source each) by owner.
- **`.json` keys only.** flagd picks its parser from the extension, and the module renders JSON, so `key` must end in `.json` (default `flags.flagd.json`). The flagd source's `key` must name the same value.
- **Plan-time validation mirrors flagd's own rules.** A flag needs at least one variant, all variants share one type, `defaultVariant` must name a defined variant, no evaluator is empty, and metadata holds scalars only (`flagSetId` and `version` as strings) -- a broken file fails at plan instead of reaching flagd, where flagd would refuse it and keep serving the last good definitions.
- **DISABLED versus off.** `state: DISABLED` returns each caller's own default value with reason `DISABLED`; targeting and `defaultVariant` are ignored. To switch a feature off uniformly, point `defaultVariant` (and targeting) at the off variant and keep the flag `ENABLED`.
- **Targeting returns a defined name or null.** A `null` result falls back to `defaultVariant`; a variant name the flag does not define is an evaluation error, not a fallback.
- **An empty default variant defers to callers.** With `defaultVariant` unset, evaluations that targeting does not resolve return the caller's code default with reason `DEFAULT`. Use it only when every caller ships the same fallback.
- **Evaluators are one level deep.** `evaluators` render as `$evaluators`; a flag references one with `{"$ref": "<name>"}`, and an evaluator cannot reference another.
- **Duplicates across files.** When two files mounted by one flagd define the same flag key, the source listed later in flagd's `sources` wins. Give each file its own flag keys and a `metadata.flagSetId`.

## On the diagram

Each flag file renders as its own node wired to the flagd that mounts it, so the diagram shows which files -- and which teams' flags -- feed each daemon. Splitting flags into one file per owning team makes that picture, and the review trail, match the organization.

## Pairs well with

- **KubernetesFlagd** -- the daemon that mounts and serves the file.
- **KubernetesNamespace** -- one namespace for the daemon and its files.
