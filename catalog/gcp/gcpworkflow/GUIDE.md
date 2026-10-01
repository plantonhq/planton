# GcpWorkflow Guide

Operational judgment for running Cloud Workflows as code — the things
the spec reference cannot tell you.

## Revisions roll forward, executions do not

Every source / env-var / service-account change mints a NEW revision, and
only NEW executions use it. A long-running execution started before the
deploy finishes on the OLD revision — a bug fix does not reach executions
already in flight. If an in-flight execution must die with the old logic,
cancel it explicitly (`gcloud workflows executions cancel`).

## The service account is the blast radius

Every HTTP call with OIDC/OAuth auth and every connector call carries the
workflow's service account. The default compute account makes a workflow
quietly over-privileged. Mint a dedicated GcpServiceAccount per workflow
and grant it exactly the target services' invoke/read roles — the
workflow is then an auditable principal like any other service.

## Secrets are pointers, not values

Cloud Workflows has no native secret injection, so never put a secret in `userEnvVars`. Put it in `secretEnvVars` instead. The module stores each value in its own Secret Manager secret and grants the workflow's service account `secretmanager.secretAccessor` on that secret alone. The workflow's environment variable then holds only the version's resource name (`projects/<p>/secrets/<id>/versions/<n>`). The source reads the value when it needs it, with `googleapis.secretmanager.v1.projects.secrets.versions.access` and `text.decode(base64.decode(...))`. Secrets need `region`. The grant goes to `serviceAccount`, which must then be an email, or to the default compute account when it is empty. The 20-variable limit counts both maps together.

## Execution history dies with the workflow

Deleting a workflow deletes its execution history — there is no undelete
and no export-after-the-fact. `deletionProtection` ships ON for exactly
this reason. If the history matters for audit, route it out continuously:
`callLogLevel: LOG_ALL_CALLS` sends call logs to Cloud Logging, where a
GcpLoggingSink can archive them independently of the workflow's life.

## Logging levels are a cost lever, not a toggle

`LOG_ALL_CALLS` on a high-frequency workflow (one execution per Pub/Sub
message) writes a log entry per step per execution — that is a real
Logging bill. Develop loud (`LOG_ALL_CALLS` +
`EXECUTION_HISTORY_DETAILED`), run quiet (`LOG_ERRORS_ONLY` + basic
history), and rely on the trigger/bus platform logs for delivery-level
visibility.

## Source size is a design smell before it is a limit

The 128KB cap is generous; a workflow approaching it is doing too much
inline. Push per-item logic into the services the workflow calls, keep
the workflow as the coordination skeleton — smaller sources also diff
reviewably, which is the point of workflows-as-code.

## CMEK must be granted before the first deploy

With `cryptoKey` set, the Workflows service agent
(service-{project_number}@gcp-sa-workflows.iam.gserviceaccount.com) needs
roles/cloudkms.cryptoKeyEncrypterDecrypter on the key BEFORE the apply —
the deploy fails otherwise, and the agent only exists after the API is
first enabled in the project. Grant through the GcpKmsKey's iam_members
in the same chart and order the workflow after it.

## Teardown discipline

`deletionProtection: false` plus `deletionPolicy: DELETE` removes the
workflow, cancels running executions, and erases history. `ABANDON`
keeps the workflow running unmanaged — the right escape when handing a
workflow to another team's IaC. `PREVENT` is defense in depth on top of
deletion_protection for compliance-critical orchestrations.

The guard's exact behavior, live-verified: with `deletionProtection`
ON, a destroy fails with "cannot destroy workflow without setting
deletion_protection=false and running `terraform apply`" and the
workflow is untouched. Disarming is a TWO-step operation by design —
apply the `false` first, then destroy; there is no single-shot
override.
