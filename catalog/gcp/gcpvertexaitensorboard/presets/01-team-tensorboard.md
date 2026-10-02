# Team TensorBoard

## Use Case

One shared TensorBoard for a team's training work: every Vertex AI training job, pipeline step, and SDK session names it, creates its own experiments and runs, and the history stays comparable after the jobs end.

## When to Use

- The first managed TensorBoard in a project
- Teams that let training code decide experiment and run names
- Any time "where did last month's loss curve go?" has been asked

## What This Creates

- A TensorBoard in `us-central1` displayed as "Training metrics", with no declared experiments

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | The region your training jobs run in (fixed at creation). |
| `experiments` | none | Declare experiments and runs something must find before the first job logs. |
| `deletionPolicy` | `DELETE` | `PREVENT` once the TensorBoard holds history the team compares against. |
