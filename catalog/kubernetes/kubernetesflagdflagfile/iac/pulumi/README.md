# KubernetesFlagdFlagFile Pulumi Module

Renders the flag definitions into one ConfigMap named after the resource, under the spec's data key (default `flags.flagd.json`). A KubernetesFlagd `config_map` source mounts it as a directory.

## What the Module Creates

1. **ConfigMap `<metadata.name>`** -- the flagd JSON document (`module/main.go`), stamped with the standard governance labels

## Rendering Notes

- **One renderer** -- `flag_file.go` writes `$schema`, `flags`, and `$evaluators` and `metadata` when declared; a flag renders `state`, `variants`, and `defaultVariant`, `targeting` and `metadata` only when set. The Terraform twin renders the byte-identical document (Go `json.Marshal` and OpenTofu `jsonencode` sort keys and escape alike).
- **Variant values keep their type** -- booleans, strings, numbers and objects render as native JSON values.

## Inputs

`KubernetesFlagdFlagFileIacInput`: `target` and `provider_config`.

## Outputs

`config_map_name`, `key`, `namespace`.

## Local Development

```shell
planton pulumi preview --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```
