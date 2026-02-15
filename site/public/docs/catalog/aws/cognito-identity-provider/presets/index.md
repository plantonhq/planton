---
title: "Presets"
description: "Ready-to-deploy configuration presets for Cognito Identity Provider"
type: "preset-list"
componentSlug: "cognito-identity-provider"
componentTitle: "Cognito Identity Provider"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-google-oauth"
    rank: "01"
    title: "Preset: Google OAuth"
    excerpt: "Configures Google as a social identity provider for a Cognito User Pool."
  - slug: "02-oidc-enterprise"
    rank: "02"
    title: "Preset: Enterprise OIDC"
    excerpt: "Configures a generic OIDC provider for enterprise single sign-on. Works with Okta, Auth0, Azure AD, Keycloak, and any OIDC-compliant identity provider."
  - slug: "03-saml-federation"
    rank: "03"
    title: "Preset: SAML Federation"
    excerpt: "Configures a SAML 2.0 identity provider for enterprise federation. Works with Azure AD, ADFS, Salesforce, Okta (SAML mode), and any SAML 2.0 compliant IdP."
---

# Cognito Identity Provider Presets

Ready-to-deploy configuration presets for Cognito Identity Provider. Each preset is a complete manifest you can copy, customize, and deploy.
