# Tagged Patch Scan

This preset binds AWS's own `AWS-RunPatchBaseline` document — no
custom document needed — to every instance tagged `env: prod`,
scanning nightly and reporting missing patches as HIGH-severity
compliance findings.

## When to Use

- The first patch-visibility step in any account: know what's missing
  before automating installs
- Fleets targeted by tag so new instances join coverage with no plan
  edits

## What You Get

- A nightly 02:00 scan (only at the interval — no immediate first run)
  that never installs anything (`Operation: Scan`)
- Compliance findings per instance in the Systems Manager console
- Rate controls keeping large fleets civilized (10% at a time, stop at
  5% failures)

## Customize

- Switch `Operation: Install` to remediate (pair with a maintenance
  window and the baseline you govern via
  [AWS SSM Patch Baseline](/infra-catalog/aws-ssm-patch-baseline))
- Add `outputLocation` to keep command output in S3
- Gate runs on a Change Calendar with `calendarNames`
