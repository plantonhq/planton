# OpenStackApplicationCredential Pulumi Module Overview

## Architecture

Single-resource module: creates one `identity.ApplicationCredential` from the spec.

## Notes

- All fields are ForceNew -- any spec change recreates the credential
- Secret output is sensitive -- Pulumi marks it in state
- Access rules use `identity.ApplicationCredentialAccessRuleArgs`
- Roles are passed as `pulumi.StringArray`
