# KubernetesFlagdFlagFile

Declares a flagd flag definition file -- typed flags with their variants, default variant and JSONLogic targeting, plus shared evaluators and flag-set metadata -- and renders it as JSON, with the flagd `$schema`, into one ConfigMap named after the resource. A **KubernetesFlagd** `config_map` source mounts that ConfigMap as a directory and serves the flags.

Flags change far more often than the daemon that serves them, are often many per daemon, and are often owned by a different team. Keeping them in their own resource means a flag flip touches this resource only: flagd is never re-applied, and it serves the change once the kubelet syncs the mounted volume -- typically one to two minutes -- with no restart.

## What Gets Created

| Object | Name | When |
|---|---|---|
| ConfigMap | `<metadata.name>` | always -- the rendered definitions under the data key `key` (default `flags.flagd.json`) |

## Spec Walkthrough

- **`namespace`** (required) -- a literal or a reference to a KubernetesNamespace. Put it in the KubernetesFlagd's namespace: pod volumes cannot cross namespaces.
- **`key`** -- the ConfigMap data key, default `flags.flagd.json`. It must end in `.json`: flagd picks its parser from the extension.
- **`flags`** -- keyed by flag key (the key OpenFeature SDKs evaluate). Each flag carries:
  - `state` -- `ENABLED` serves the flag; `DISABLED` makes every evaluation return the caller's default value with reason `DISABLED`.
  - `variants` -- the values the flag can return, each exactly one of `boolValue`, `stringValue`, `numberValue` or `objectValue`; every variant of a flag holds the same type.
  - `defaultVariant` -- the variant returned when targeting does not choose one. Empty means the caller's code default (reason `DEFAULT`, no value).
  - `targeting` -- JSONLogic returning a variant name, with flagd's `fractional`, `sem_ver`, `starts_with` and `ends_with` operations. A `null` result falls back to `defaultVariant`; a name the flag does not define is an evaluation error.
  - `metadata` -- information returned with evaluations (a description, an owner): scalar values only, with `flagSetId` and `version` strings.
- **`evaluators`** -- shared JSONLogic fragments rendered as `$evaluators`, reused from any flag's targeting with `{"$ref": "<name>"}`.
- **`metadata`** -- flag-set metadata merged into every flag's metadata, scalar values only (for example `flagSetId`, which gRPC sync selectors and SDKs filter on; `flagSetId` and `version` are strings).

Validation runs at plan time: a flag needs at least one variant, variant names are non-empty, every variant must hold the same value type, `defaultVariant` must name a defined variant, flag keys may not contain whitespace, no evaluator is empty, and metadata holds scalars only.

## Outputs

| Output | Description |
|---|---|
| `config_map_name` | Name of the rendered ConfigMap -- the value a KubernetesFlagd `config_map` source references |
| `key` | ConfigMap data key holding the definitions |
| `namespace` | Namespace of the rendered ConfigMap |

## How the Module Works

The Pulumi module and the OpenTofu module render the definitions byte for byte the same way (keys sorted, so `$evaluators` and the flagd `$schema` lead the document, ahead of `flags` and `metadata`) and apply one `ConfigMap`. Nothing else is created; the file has no cost of its own.

## Official Documentation

- Flag definitions: https://flagd.dev/reference/flag-definitions/
- Flag definition schema: https://flagd.dev/reference/schema/
- Custom JSONLogic operations: https://flagd.dev/reference/custom-operations/fractional-operation/

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
