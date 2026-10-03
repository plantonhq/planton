# AzureDataFactoryPipeline Pulumi Module

## Overview

Creates one pipeline inside an Azure Data Factory -- an ordered set of activities that executes as a whole when triggered. The activities travel as raw JSON; Azure owns the activity schema.

## Resources Created

- `datafactory.Pipeline` -- the pipeline (activities JSON, parameters, variables, concurrency, folder, elapsed-time metric)

## Outputs

- `pipeline_id` -- the pipeline's ARM resource ID (`{factory_id}/pipelines/{name}`)
- `pipeline_name` -- the pipeline's name (what triggers and run-API calls reference)

## Behavior Notes

- **`activities_json` is the raw "activities" ARRAY** from the Data Factory Studio's Code view. Invalid JSON or unknown activity shapes fail at deploy time; JSON key ordering is normalized (never shows as drift).
- **Parameters and variables are String-typed on this surface** -- the provider sends and reads only string-typed entries. Declare other-typed parameters inside the activities JSON.
- **The bridged SDK preserves a historic field-name typo** (`MoniterMetricsAfterDuration`) for v5's `monitor_metrics_after_duration` -- a name difference only; both engines write the same ARM policy.
- **No tags**: pipelines are ARM sub-resources of the factory and expose none.

## Usage

The module is executed by the Planton platform with an IaC input containing the target `AzureDataFactoryPipeline` resource and an Azure provider configuration. For a manifest example, see `../../e2e/manifest.yaml`.
