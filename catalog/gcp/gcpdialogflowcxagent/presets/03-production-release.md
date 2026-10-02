# Production Release

## Use Case

A regional support agent run like a product: security settings redact and expire conversation data, a Cloud Run webhook does fulfillment, and a frozen release of the start flow serves production and staging while the draft keeps changing.

## When to Use

- Agents serving real customers, with data residency and privacy requirements
- Teams that release conversation changes deliberately instead of editing live
- Anywhere an accidental destroy must fail

## What This Creates

- A `us-central1` agent in English and Spanish, attached to `GcpDialogflowCxSecuritySettings`
- Interaction logging and Cloud Logging
- A webhook to a Cloud Run service, authenticated with an ID token the Dialogflow service agent mints
- A version of the start flow and two environments pinning it
- `PREVENT` on destroy

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `versions` | one start-flow release | Add a version per release; flows authored in the console are named by `flowId`, and an environment needs a version for every flow its start flow can reach. |
| `environments[].versionConfigs` | the declared release | Point staging at a newer version first; `versionId` pins a version made outside this block. |
| `webhooks[].genericWebService.uri` | a Cloud Run URL | Your fulfillment endpoint (https only). Grant the Dialogflow service agent `roles/run.invoker` on a private service. |
| `securitySettings` | the `pii-redaction` block | Settings in the agent's own location. |
