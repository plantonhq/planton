# AWS Cognito Identity Provider Resource Kind

**Date**: February 15, 2026
**Type**: Feature
**Components**: API Definitions, AWS Provider, Pulumi CLI Integration, Terraform Module

## Summary

Added AwsCognitoIdentityProvider (R12a) as a new AWS resource kind for managing external identity providers (Google, Facebook, Login with Amazon, Sign in with Apple, OIDC, SAML) federated into Amazon Cognito User Pools. This is the fifteenth new AWS resource kind in the cloud provider expansion project, and the first Planton component to use a proto enum for provider-specific type safety and strongly typed oneof configuration messages.

## Problem Statement / Motivation

The AwsCognitoUserPool (R12) was completed earlier in this session, but identity providers were deliberately deferred to a separate component because they have independent lifecycles (ForceNew on name and type, can be added/removed without affecting the pool). Users configuring federated sign-in need a dedicated resource to manage the external IdP registration, attribute mapping, and provider-specific OAuth/OIDC/SAML configuration.

### Pain Points

- No way to configure social login (Google, Facebook) or enterprise SSO (OIDC, SAML) through Planton
- The Terraform `provider_details` map is untyped -- keys vary by provider type with no compile-time validation
- SAML keys use PascalCase (MetadataFile) while OAuth keys use snake_case (client_id), creating confusion
- No guidance on which fields are required vs optional for each provider type

## Solution / What's New

A complete deployment component with strongly typed provider configuration instead of the flat map used by Terraform/Pulumi. Each provider type has its own proto message with required/optional field annotations, giving users auto-complete, validation, and self-documenting APIs.

### Key Features

- **Proto enum `AwsCognitoIdentityProviderType`** with 6 values matching the AWS API exactly (Google, Facebook, LoginWithAmazon, SignInWithApple, OIDC, SAML)
- **6 typed configuration messages** in a proto `oneof` -- each provider type has its own strongly typed config with appropriate required/optional fields
- **CEL match validation** -- ensures `provider_type` enum matches the set `oneof` branch, eliminating impossible states
- **SAML mutual exclusion** -- CEL enforces exactly one of `metadata_file` or `metadata_url`
- **IaC mapping layer** -- Pulumi/TF modules convert typed messages to the flat `map[string]string` the AWS API expects, handling the snake_case vs PascalCase divergence transparently

## Implementation Details

### Proto Design

```
AwsCognitoIdentityProviderSpec
  user_pool_id     (StringValueOrRef -> AwsCognitoUserPool)
  provider_name    (string, 1-32 chars, ForceNew)
  provider_type    (AwsCognitoIdentityProviderType enum, ForceNew)
  oneof provider_config:
    google               (3 required fields)
    facebook             (3 required + 1 optional)
    login_with_amazon    (3 required)
    sign_in_with_apple   (5 required)
    oidc                 (2 required + 7 optional)
    saml                 (1 required XOR + 5 optional)
  attribute_mapping (map<string, string>, optional)
  idp_identifiers   (repeated string, max 50)
```

### Provider Details Mapping

The Pulumi module's `buildProviderDetails()` function converts each typed message to the flat map:
- OAuth/OIDC providers: snake_case keys (`client_id`, `authorize_scopes`)
- SAML: PascalCase keys (`MetadataFile`, `MetadataURL`, `IDPSignout`)
- SAML booleans: converted to string `"true"`/`"false"`

### Validation Coverage

30 spec tests covering:
- 11 happy path (all 6 provider types + attribute mapping + idp identifiers + valueFrom)
- 19 failure scenarios (type/config mismatch, SAML mutual exclusion, nested required fields, constraints)

## Benefits

- **Type-safe configuration**: Developers get auto-complete and compile-time validation per provider type instead of guessing map keys
- **Self-documenting API**: The proto messages clearly show what each provider type requires
- **Impossible states eliminated**: CEL validation prevents provider_type=Google with SAML config
- **Transparent complexity handling**: The snake_case vs PascalCase key divergence is hidden from users inside the IaC mapping layer

## Impact

- **Users**: Can now configure social login and enterprise SSO through Planton with type-safe manifests
- **Infra charts**: Identity providers can be composed with User Pools using StringValueOrRef for the user_pool_id
- **Pattern**: First component to use a proto enum for provider-specific values, establishing a precedent for future typed-config patterns

## Related Work

- `2026-02-15-150832-aws-cognito-user-pool-resource-kind.md` -- Parent component (R12)
- AwsEventBridgeRule -- Child resource pattern reference
- Cloud provider expansion project: 15 of ~32 AWS resource kinds complete

---

**Status**: Production Ready
**Timeline**: ~2 hours (design review + implementation)
