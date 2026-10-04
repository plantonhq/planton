# One Word per Idea Across the Catalog

**Date**: October 4, 2026
**Type**: Breaking Change (protos, module input contract, CLI, operator boot contract, docs)
**Components**: `shared/catalogkind`, every catalog kind's `IacInput`/`Outputs`, `pkg/iac/pulumi/pulumimodule/iacinput`, `pkg/catalogkindreflect`, `pkg/catalogbundle`, `pkg/vocabulary`, the CLI, the operator, the skills, the docs site

## Summary

A platform engineer, a developer and an agent now meet one word per idea, the same in the API, the CLI, the docs and the code, and the words are the ones the industry already uses:

- **catalog kind** (`CatalogKind`) — a type the catalog offers (`AwsS3Bucket`, `GcpGkeCluster`). Its schema is a proto message; its enum value names are permanent, because IaC state keys hang on them.
- **catalog object** — one typed document of a kind (`apiVersion`, `kind`, `metadata`, `spec`), carrying `CatalogObjectMetadata`. A preset is a `CatalogKindPreset`; one rendered object is a `CatalogObjectManifest`.
- **Infra Chart** — a parameterized template of catalog objects. **Infra Stack** — one chart installed into one environment. **Infra Component** — one deployed instance of a catalog kind. **Infra Pipeline** — one deploy or undeploy run over a stack, which runs one **Infra Job** per component.
- **provider resources** — what a component's IaC module creates in the cloud account; one component usually creates several.
- **Infra Catalog** — the catalog of kinds and charts an organization can deploy. A kind belongs to a **CatalogProvider** and its **CatalogProviderServiceGroup**.

"Component" names a deployed Infra Component and nothing else; a kind is always a catalog kind.

## What changes for a module and its callers

- Every kind's module receives `<Kind>IacInput` and returns `<Kind>Outputs` (held in `status.outputs`).
- A Pulumi module reads its input under the config key `planton:iac-input`, or from `IAC_INPUT_YAML` / `IAC_INPUT_YAML_FILE`, through `iacinput.LoadIacInput`.
- The CLI's flag is `--iac-input`.
- The operator hands the control plane its Temporal queues as `TEMPORAL_TASK_QUEUE_INFRA_JOB` (`infra-job`), `TEMPORAL_TASK_QUEUE_INFRA_JOB_IAC_OPERATION`, `TEMPORAL_TASK_QUEUE_INFRA_COMPONENT_PURGE` and `TEMPORAL_TASK_QUEUE_INFRA_STACK_PURGE`.
- The catalog bundle's entries link each kind's `iacInput` and `outputs`; the bundle declares format `2`, and a consumer that understands another format refuses it with both versions named.
- IDs carry `ic_` (an Infra Component, with its kind segment), `ij_` (an Infra Job), `infstk_` (an Infra Stack) and `infpipe_` (an Infra Pipeline).
- A folder checked out from a deployed stack carries `.planton/stack.yaml` (`stackId`, `stackName`, `org`, `env`).

## Keeping it that way

`pkg/vocabulary` holds the words with their meanings (`vocabulary.yaml`) and a whole-tree test that fails on any spelling the vocabulary retires, naming the word to write instead; the `lint.vocabulary` lane runs it on every pull request. Vendors' own words — Google Cloud's Resource Manager, Pulumi's stacks and projects, BigQuery's own connection types — are allowed by name, never by blanket exception.

## Docs and skills

The architecture docs, the catalog-kind rules (`_rules/catalog-kind/`), both skills, the Planton Assistant's instructions, the docs site (`/docs/infrastructure/infra-components`, `catalog-kinds`, `infra-stacks`, `infra-jobs`) and every kind's README, catalog page and guide speak these words. The site serves its own 404 for any address it no longer has.
