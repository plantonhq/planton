# Kubernetes GO Feature Flag File

## When NOT to Use This

**One resource is ONE GO Feature Flag flag file** -- typed, validated flags rendered into a ConfigMap named after the resource, which a `KubernetesGoFeatureFlag` relay reads through its `configMap` retriever.

Not the right kind when:

- **Your flags live in a repository, a bucket or a database** -- point the relay's GitHub, GitLab, Bitbucket, S3, GCS, Azure Blob, MongoDB, Redis or PostgreSQL retriever at them instead.
- **You run flagd** -- use `KubernetesFlagdFlagFile`, which renders flagd's own flag format.
- **The flags are secrets** -- a flag file is configuration in a ConfigMap; nothing in it is confidential.

## Why the flags are their own resource

Flags change far more often than the relay, are many per relay, and are often owned by a different team. Keeping them here means a flag flip touches this resource only: the relay is never re-applied, write access to the flags can be granted without access to the relay, and the relay serves the change on its next poll (its `pollingIntervalMs`, 60 seconds by default) with no restart. Several flag files can feed one relay; the retriever listed later wins on a flag both define.

## What is validated before it ships

Every rule GO Feature Flag enforces on a flag is checked at plan time, except query syntax, which the relay checks when it loads the file (a query that does not parse drops the flag):

- at least one variation, and every variation of a flag holding the same type (boolean, string, number, object or list);
- a default rule that resolves to a value (a variation, a percentage split or a progressive rollout), carries no query and is never disabled;
- every enabled targeting rule carrying a query and resolving to exactly one outcome;
- rules and progressive rollouts naming only variations the flag defines;
- percentage shares that are never negative and give some variation a share (shares are relative weights);
- progressive rollouts that ramp forward -- an end share (empty or 0 means 100) at least the initial share, and an end date after the initial date;
- unique rule names (scheduled steps update rules by name);
- real RFC 3339 dates on rollouts, experimentation windows and scheduled steps -- a date that does not exist, such as February 30, is refused.

## Scheduled steps

At its date a `scheduledRollout` step is merged into the flag: variations are added or replaced by name, targeting rules are merged by name field by field (a rule with an unknown name is added), and the default rule is merged field by field. Fields a step leaves empty leave the flag as it is. A step sets variations, targeting, the default rule, `trackEvents`, `disable`, `version` and `experimentation`; the bucketing key and metadata belong to the flag itself.

## Rendering

The module renders the flags as one JSON object keyed by flag name, in GO Feature Flag's flag format (`variations`, `targeting`, `defaultRule`, `percentage`, `progressiveRollout`, `scheduledRollout`, `experimentation`, `metadata`). JSON is YAML, so the relay's default `yaml` file format reads it under the default key `flags.goff.yaml`; both engines produce the file byte for byte. A ConfigMap holds up to 1 MiB.

## Outputs

| Output | Meaning |
|---|---|
| `config_map_name` | The rendered ConfigMap -- the relay retriever's `configMapName` |
| `key` | The data key holding the flag file -- the retriever's `key` |
| `namespace` | The ConfigMap's namespace -- the relay's own, or one its retriever names |

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
