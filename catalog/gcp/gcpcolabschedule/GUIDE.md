# GcpColabSchedule Guide

The judgment this guide protects: a schedule keeps launching runs that bill until it is paused, bounded, or destroyed, and it runs them as whoever you name. Give unattended runs a service account, bound schedules that should end, and decide concurrency before a slow run piles up behind itself.

## One API, two kinds of run

Google's Schedules API launches either a Colab Enterprise notebook execution or a Vertex AI Pipelines run; the provider files it under Colab, which is why the kind is named for it. Exactly one of `notebookExecutionJob` and `pipelineJob` is set:

- **Notebook runs** execute a notebook top to bottom and write the executed copy (with its outputs) to `gcsOutputUri`. The notebook comes from Cloud Storage (`gcsNotebookSource`, optionally pinned to an object generation) or a Dataform repository (`dataformRepositorySource`, optionally pinned to a commit). The machine comes from a runtime template (`notebookRuntimeTemplateResourceName`, the usual choice) or a `customEnvironmentSpec` (machine, GPUs or TPUs, reservation, network, disk). The notebook arm is immutable: a change replaces the schedule.
- **Pipeline runs** launch a compiled Kubeflow pipeline -- inline JSON (`pipelineSpec`) or an Artifact Registry template (`templateUri`) -- with `runtimeConfig.parameterValues` and an output root. The pipeline arm updates in place.

## Identity

A notebook run needs exactly one of `serviceAccount` and `executionUser`. Use a service account for anything unattended: a user-run schedule breaks when the person leaves, and it acts with their access. Whoever applies the schedule needs `iam.serviceAccounts.actAs` on the account. Pipeline runs use `pipelineJob.serviceAccount` (the Compute Engine default when unset).

## Timing and concurrency

`cron` accepts a `TZ=Area/City` prefix; without it, times are UTC. `startTime`, `endTime`, and `maxRunCount` bound the schedule -- it completes when either limit is reached. `maxConcurrentRunCount` caps runs started at once; a run over the cap is skipped unless `allowQueueing` is true. For pipelines, `maxConcurrentActiveRunCount` caps runs in flight.

## Pausing

`desiredState` is enforced on every apply: `PAUSED` stops launching (resume with `ACTIVE`). A schedule paused by hand in the console resumes on the next apply if the manifest says `ACTIVE`.

## Networking and encryption

Custom notebook environments can join a VPC (`networkSpec`); pipeline runs peer with a network whose path Google wants with the project number -- reference a `GcpVpcNetwork` and the modules resolve it -- or reach private services through PSC interface (`pscInterfaceConfig`). Both arms take `kmsKeyName` for customer-managed encryption.

## Destroy

Destroying the schedule stops future runs. Runs already launched, their executed notebooks, and their pipeline artifacts stay. `ABANDON` leaves the schedule launching runs outside management -- rarely what you want.
