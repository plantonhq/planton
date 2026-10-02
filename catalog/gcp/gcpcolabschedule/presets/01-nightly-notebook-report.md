# Nightly Notebook Report

## Use Case

Run a reporting notebook every morning -- executed top to bottom on a standard machine, as a service account, with the executed copy (charts and tables included) saved to a bucket.

## When to Use

- Daily KPI reports built in a notebook
- Data-quality checks a team reviews each morning

## What This Creates

- A schedule running `nightly.ipynb` at 06:00 New York time on the `standard-cpu` template, one run at a time with a one-hour limit, as the `notebook-runner` service account, writing to the `notebook-runs` bucket

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `cron` | 06:00 New York daily | Your cadence and time zone. |
| `notebookExecutionJob.gcsNotebookSource` | Cloud Storage | `dataformRepositorySource` for notebooks under version control. |
| `notebookExecutionJob.executionTimeout` | `3600s` | Longer for heavy reports. |
| `desiredState` | active | `PAUSED` during a freeze. |
