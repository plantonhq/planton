# Weekly Pipeline Retraining

## Use Case

Retrain a model every week by launching a compiled Vertex AI pipeline from Artifact Registry with its parameters, on a peered private network, under your encryption key.

## When to Use

- Scheduled retraining of production models
- Any Kubeflow pipeline that should run on a cadence

## What This Creates

- A schedule launching the `churn-retraining` pipeline template every Monday at 02:00 UTC, one run in flight at a time, failing fast, as the `pipeline-runner` service account, peered with `ml-vpc` (the modules resolve its project number), encrypted with `ml-key`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `pipelineJob.templateUri` | a template version | Your pipeline template, or `pipelineSpec` for inline JSON. |
| `pipelineJob.runtimeConfig.parameterValues` | training window | Your pipeline's parameters. |
| `cron` | Monday 02:00 UTC | Your cadence. |
