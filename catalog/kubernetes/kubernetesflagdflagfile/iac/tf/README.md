# KubernetesFlagdFlagFile Terraform Module

Renders the flag definitions into one `kubernetes_config_map_v1.flag_file`, named after the resource, under the spec's data key (default `flags.flagd.json`).

## Module Behavior

- **The document is built in `locals.tf`** -- byte-identical to the Pulumi twin's rendering: `$schema`, `flags`, and `$evaluators` and `metadata` when declared. Flags and evaluators arrive untyped (they hold free-form JSONLogic), so every field is read with `try()`.
- **Nothing else is created** -- the file reserves no cluster capacity and has no cost of its own.

## Resources

| Resource | Condition |
|---|---|
| `kubernetes_config_map_v1.flag_file` | always |

## Usage

```shell
planton tofu init --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu plan --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu apply --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

State is kept by the default local backend.
