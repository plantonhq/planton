# Governed Canvas Workspace

A SageMaker Domain configured for business analysts using Canvas (no-code ML),
with governance guardrails: models flow through the model registry instead of
being deployed straight to endpoints, and the code-first Studio surfaces are
hidden from the UI.

## When to Use

- Business or analytics teams building models without writing code
- Organizations that require model review/approval before anything serves traffic
- Mixed-audience domains where analysts should not see notebooks or terminals

## Configuration Highlights

- **Direct deploy disabled**: Canvas users cannot create billable real-time
  endpoints; registering to the SageMaker Model Registry is the only path out
- **Forecasting enabled**: time-series forecasting via Amazon Forecast under a
  dedicated IAM role
- **Pinned workspace**: Canvas artifacts (datasets, intermediate results,
  models) live at a known S3 location for lifecycle and access control
- **UI governance**: classic Jupyter Server and Code Editor hidden from the
  Studio launcher

## Cost Estimate

- Canvas sessions: billed hourly per active user (Canvas workspace pricing) — the dominant line
- Model building: billed per training cell/AutoML job
- Domain infrastructure: EFS storage, billed per GB-month

The verified figure for this preset lives in the kind's generated estimate at `catalog/_pricing/estimates/awssagemakerdomain.yaml` — computed from the pinned price book, never hand-typed here.

## Customization

- Add `identityProviderOauthSettings` to connect Snowflake or Salesforce Data
  Cloud through OAuth (client credentials in a Secrets Manager secret)
- Add `generativeAiBedrockRoleArn` to enable Bedrock generative-AI features
- Add `emrServerlessSettings` for large-scale data preparation
- Set `kendraSettingsStatus: ENABLED` for document querying against Kendra
