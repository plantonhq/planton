# Automation Runbook Window

This preset runs YOUR automation runbook — wired by reference to an
[AWS SSM Document](/infra-catalog/aws-ssm-document) component — once
per nightly window, untargeted (the runbook manages its own scope).

## When to Use

- Nightly operational chores expressed as Automation documents:
  snapshot rotation, resource cleanup, drift remediation
- Runbooks that should execute inside a bounded window with a hard
  cutoff, not on an open-ended schedule

## What You Get

- A daily 2-hour window whose single untargeted task runs the
  referenced runbook's `$DEFAULT` version once per execution
- Chart wiring: releasing a new default document version changes what
  runs with no window edit

## Customize

- Add a registered target plus `WindowTargetIds` task targeting (with
  rate controls) to fan a runbook across instances
- Add LAMBDA or STEP_FUNCTIONS tasks for glue that is code, not
  documents — each type's invocation arm is CEL-matched to its task
  type
- `enabled: false` stages the window dark until you flip it on
