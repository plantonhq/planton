# AwsSsmMaintenanceWindow — Kind Guide

Authored operational judgment for the maintenance window component:
the design decisions behind the spec's shape, and what to know before
operating windows in production.

## Design decisions

- **Targets and tasks fold into the window.** Both are true satellites
  — their window binding forces replacement at the provider and they
  cannot outlive the window — so both modules create and destroy them
  with the window, keyed by name (the for_each address, the
  target_ids/task_ids output key, and half of each
  `{window_id}/{id}` import composite).
- **`enabled` is a tri-state** because the provider default is TRUE:
  unset means enabled, and only an explicit `false` creates the window
  paused — the modules render it only on an explicit choice.
- **The invocation union is CEL-enforced twice**: exactly one arm, and
  the arm must match `taskType` — the provider sends whatever arms are
  present without checking, and AWS rejects mismatches server-side at
  apply time. The spec rejects them at validate time.
- **Rate controls render only on targeted tasks**: AWS rejects
  `max_concurrency`/`max_errors` on untargeted tasks, so both modules
  omit them when a task has no targets.
- **`taskArn` is a value-or-reference without a default kind** — what
  it names follows `taskType`: a document name (AWS-managed literal or
  an AwsSsmDocument's `document_name` output) for
  RUN_COMMAND/AUTOMATION, an AwsLambda `function_arn` for LAMBDA, an
  AwsStepFunction `state_machine_arn` for STEP_FUNCTIONS.

## Operating windows in production

- **Cutoff arithmetic**: duration is 1–24 hours, cutoff 0–23 and
  strictly less than duration — the cutoff is when NEW task starts
  stop; running tasks continue or cancel per each task's
  `cutoffBehavior` (default CONTINUE_TASK).
- **Priority 0 is highest and also AWS's default** — an unset priority
  competes at the front. Tasks sharing a priority run in parallel.
- **Untargeted tasks run once per window execution** regardless of
  fleet size — pair them with runbooks that manage their own scope. Untargeted is legal ONLY for Automation, Lambda, and Step Functions tasks: AWS rejects an untargeted RUN_COMMAND task at registration ("you must specify at least one resource as the target" — live-caught), so every run-command task must name instance IDs or window target IDs.
- **Targeting registered targets from a task uses the
  `WindowTargetIds` key** — inside one spec, just write the target
  entry's NAME as the value and the modules resolve it to the
  cloud-generated registration ID (the name-based join);
  cross-component consumers read the `target_ids` output, never guess.
- **A target's name, description, and resource type force
  re-registration** (a new target ID) — task references via
  WindowTargetIds must follow, which the outputs do automatically.
- **`allowUnassociatedTargets` is create-relevant honesty**: leave it
  false so every task must name a registered target; setting it true
  lets tasks target ad-hoc instances the window never registered.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
