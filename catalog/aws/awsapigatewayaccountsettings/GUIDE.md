# AwsApiGatewayAccountSettings — Kind Guide

Authored operational judgment for the API Gateway account-settings
singleton: the design decisions behind the spec's shape, and what to
know before operating region-wide API Gateway logging.

## Design decisions

- **A settings singleton, not a gateway field.** Multiple REST APIs
  share the ONE account object per region — folding the role into
  AwsRestApiGateway would make every gateway instance fight over it.
  This is the same reasoning as the Bedrock invocation-logging and
  SES account-settings singletons.
- **The region is the identity.** metadata.name never reaches AWS;
  deploy at most one instance per region. Two instances targeting the
  same region will each try to own the same setting.
- **Unset role is a real posture.** An instance without
  `cloudwatch_role_arn` explicitly manages "no API Gateway logging"
  — applying it clears any role previously set by anyone.

## Operating region-wide logging in production

- **Set this BEFORE enabling stage logging.** Stage-level
  execution/access logging on REST APIs fails silently without the
  account role — the stages deploy fine and log nothing.
- **One role serves every API in the region.** Use AWS's managed
  `AmazonAPIGatewayPushToCloudWatchLogs` policy; custom policies that
  miss `logs:CreateLogGroup` fail AWS's apply-time validation.
- **IAM propagation lag is retried, not fatal.** A role created
  seconds before this setting can race IAM propagation; both engines'
  providers retry through "The role ARN does not have required
  permissions".
- **Destroy stops logging region-wide.** Destroying this component
  resets the role, which silences execution/access logs for EVERY
  REST API in the region — drain expectations accordingly.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
