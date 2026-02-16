---
title: "Preset: SAML Federation"
description: "Configures a SAML 2.0 identity provider for enterprise federation. Works with Azure AD, ADFS, Salesforce, Okta (SAML mode), and any SAML 2.0 compliant IdP."
type: "preset"
rank: "03"
presetSlug: "03-saml-federation"
componentSlug: "cognito-identity-provider"
componentTitle: "Cognito Identity Provider"
provider: "aws"
icon: "package"
order: 3
---

# Preset: SAML Federation

Configures a SAML 2.0 identity provider for enterprise federation. Works with
Azure AD, ADFS, Salesforce, Okta (SAML mode), and any SAML 2.0 compliant IdP.

## What This Creates

- An identity provider registration for SAML 2.0 federation
- Attribute mapping using standard SAML claim URIs
- Single logout (SLO) enabled

## Variables to Replace

- `${USER_POOL_ID}` -- Cognito User Pool ID (e.g., `us-east-1_Ab1Cd2EfG`)
- `${SAML_METADATA_URL}` -- URL to the IdP's SAML metadata document

## After Deployment

Add the provider name (`"CorpAD"`) to the `supportedIdentityProviders` list
in your User Pool Client configuration.

## Metadata Source

This preset uses `metadataUrl` which fetches the metadata document from the
IdP. For air-gapped environments, replace `metadataUrl` with `metadataFile`
containing the inline XML metadata content.

## Attribute Mapping

The SAML claim URIs used in `attributeMapping` follow the standard
WS-Federation claim types. Adjust these to match your IdP's SAML assertion
attribute names.
