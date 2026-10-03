# A Keyless Connection Deploys Through Either Engine, and Never as the Machine

**Date**: September 28, 2026
**Type**: Fix
**Components**: Provider environment (`pkg/iac/iacinput/providerenvvars`), Google Cloud keyless exchange (`pkg/iac/provider/gcp/gcpwebidentity`, new), OpenTofu boundary (`pkg/iac/tofu/tofumodule`), E2E harness (`e2e/framework/runner`)

## Summary

A keyless Google Cloud connection couldn't deploy anything. The provider-environment loader refused its web identity before it knew which engine it served, so Pulumi, which exchanges the token in the plugin, was refused along with OpenTofu, which had no keyless Google path at all. A keyless Azure connection on OpenTofu was worse: the loader ignored its token, so azurerm signed in as whatever identity the runner's machine held. On a runner with an Azure managed identity, the deploy would have run as the runner, not the connection.

Keyless now works on every engine. On OpenTofu and Terraform, Google Cloud's token is exchanged once (STS, then impersonation of the connection's service account) for an access token in `GOOGLE_OAUTH_ACCESS_TOKEN`, and Azure's token goes to azurerm as `ARM_USE_OIDC` and `ARM_OIDC_TOKEN`. On Pulumi the builders keep the exchange, and the loader emits no credential.

The engine is now a required choice, not a flag. `Options.ResolveAwsWebIdentity bool` is gone, because its default meant "Pulumi", and an OpenTofu caller that forgot it (the E2E harness did) left a keyless AWS configuration with no credential. `Options.Engine` has no default. A keyless configuration whose caller names no engine is refused with a sentence that says what to set.

## What Changed

- **`providerenvvars.Engine`:** `EngineReadsEnvironment` (OpenTofu, Terraform) and `EngineBuildsProviders` (Pulumi). An unset engine refuses keyless configurations only. Stored-key and runner-mode configurations read the same on every engine and are unchanged.
- **Google Cloud (`gcp.go`):** keyless on OpenTofu emits exactly `GOOGLE_OAUTH_ACCESS_TOKEN`, through the access-token arm. Keyless on Pulumi emits nothing.
  - Why an exchange rather than an external-account credentials file: the file would put the minted token on disk, which the provider config promises never happens, and every engine process would re-exchange a token minted to live minutes.
  - One exchange before the engine runs gives the job an hour, the same contract AWS keyless has.
- **`gcpwebidentity` (new):** the mirror of `awswebidentity`, with `Validate`, a `TokenResolver` seam and `ResolveAccessToken`. It is built on `golang.org/x/oauth2/google/externalaccount`, with an in-memory subject token and the impersonation URL.
- **Azure (`azure.go`):** keyless on OpenTofu emits `ARM_CLIENT_ID`, `ARM_TENANT_ID`, `ARM_SUBSCRIPTION_ID`, `ARM_USE_OIDC=true` and `ARM_OIDC_TOKEN`, never a secret. Keyless on Pulumi emits the coordinates only.
- **AWS (`aws.go`):** the same exchange as before, now keyed on the engine. The timeout is shared by both exchanges.
- **Callers:** `tofumodule.GetProviderConfigEnvVars` and the E2E harness's `terraform_input.go` name `EngineReadsEnvironment`.
- **The gate:** `TestEveryKeylessProvider_NeverFallsBackToAmbient` derives its providers from the registered provider configs carrying `web_identity`. A provider that gains keyless later is therefore held to the rule, and fails until its arm and its row exist. It runs each provider through both engines and an unset one.
- **`gcpplantonrunner/iac/tf/provider.tf`:** its comment now names the variable the runtime injects.

## Verification

- `go test ./pkg/iac/iacinput/providerenvvars/ ./pkg/iac/tofu/tofumodule/ ./pkg/iac/provider/...` passes.
- The four Bazel test targets pass.
- The `gcpwebidentity` tests run the exchange against an `httptest` fake of STS and IAM Credentials. They pin the audience sent, the subject-token type, the impersonated account and the returned token, and check that a refused exchange's error names the account and the pool provider.
- **Red proofs:**
  - Azure's previous body, restored with only the new signature, fails the gate on all three Azure subtests.
  - Google's previous refusal, restored, fails the gate and four Google tests.
- `go build ./e2e/framework/... ./pkg/iac/...` is clean.
- The live proof is the platform's keyless connection checks, which now deploy through both engines.
