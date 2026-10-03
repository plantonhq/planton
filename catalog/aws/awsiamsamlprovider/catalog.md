# AWS IAM SAML Provider

Creates an IAM SAML identity provider -- the account's trust anchor for SAML 2.0 federation, built from the metadata XML your identity provider (Okta, Entra ID, Google Workspace) publishes. IAM roles then trust it via `sts:AssumeRoleWithSAML` in their trust policies, replacing long-lived IAM user credentials with federated, short-lived role sessions. The provider's name is write-once at AWS; the metadata document updates in place, which is how certificate rotations ship.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SAML Provider** -- the IAM identity provider, named from `metadata.name`, created from the IdP's metadata document (issuer, SSO endpoints, and public signing certificates). IAM is global: the provider exists account-wide regardless of the endpoint region the stack ran against.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with IAM permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- The metadata XML from your IdP's metadata URL -- most IdPs publish it unauthenticated; it is a public trust document, not a secret.
- IAM roles that trust the provider come next; the provider must exist before their trust policies can name its ARN.

## Deploy

### Console

Open the deployment store, find **AWS IAM SAML Provider**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the metadata document. Start from the **Corporate IdP** preset in the [Presets](#presets) tab for the one-time federation setup.

### CLI

Create a manifest and apply it. The `samlMetadataDocument` below is a structural example -- paste your IdP's actual published metadata verbatim, because IAM parses the document at create (including decoding every certificate) and rejects placeholders:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsIamSamlProvider
metadata:
  name: okta
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  samlMetadataDocument: |
    <?xml version="1.0" encoding="UTF-8"?>
    <md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata"
        entityID="http://www.okta.com/exk1234example">
      <md:IDPSSODescriptor WantAuthnRequestsSigned="false"
          protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
        <md:KeyDescriptor use="signing">
          <ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
            <ds:X509Data>
              <ds:X509Certificate>your-idp-signing-certificate-goes-here</ds:X509Certificate>
            </ds:X509Data>
          </ds:KeyInfo>
        </md:KeyDescriptor>
        <md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified</md:NameIDFormat>
        <md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
            Location="https://acme-corp.okta.com/app/amazon_aws/exk1234example/sso/saml"/>
        <md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
            Location="https://acme-corp.okta.com/app/amazon_aws/exk1234example/sso/saml"/>
      </md:IDPSSODescriptor>
    </md:EntityDescriptor>
```

```shell
planton apply -f aws-iam-saml-provider.yaml
```

This creates the account's federation trust anchor from the pasted metadata, exporting the provider ARN for role trust policies and the certificate expiry date to rotate by. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a SAML provider. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The name is write-once** -- AWS has no rename for SAML providers: changing `metadata.name` replaces the provider, mints a NEW ARN, and silently breaks every role trust policy naming the old one. Treat a rename as a migration -- create the new provider, update every trust policy and the IdP's app settings, then retire the old.

**The metadata document is public, and IAM actually parses it** -- it carries the IdP's issuer, endpoints, and PUBLIC signing certificates, so it rides the spec as plain text, never a secret reference. IAM base64-decodes every certificate in it at create and rejects anything that is not a valid certificate -- a self-signed certificate is fine, a placeholder blob is not. It does NOT validate trust (no chain, no CA, no endpoint reachability).

**Rotate before `valid_until` or federation breaks** -- the output is derived from the metadata's certificate expiry, and an expired document fails every federated sign-in at once. IdPs rotate signing certificates on their own schedules (Okta does it routinely); re-pasting the fresh metadata is an ordinary in-place update, so put the date on a calendar.

**Both sides must agree** -- an IAM role trusts the provider with `Principal: {Federated: <provider_arn>}`, action `sts:AssumeRoleWithSAML`, and typically a `SAML:aud` condition pinned to AWS's SAML sign-in endpoint. The IdP's AWS app must carry the matching `role_arn,provider_arn` pairs, because SAML assertions name both ARNs. Half-configured federation fails at sign-in, not at apply.

**Session length lives on the role** -- federated console sessions cap at the trusting role's `maxSessionDuration`. Raise it there and align the IdP's session settings with it; nothing on this resource controls session length.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies -- the metadata document is pasted IdP content, not a reference to another Infra Component.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `provider_arn` | The provider's ARN (also the import ID) | The `Federated` principal in AWS IAM Role trust policies, and the provider half of the IdP app's `role_arn,provider_arn` pairs |
| `valid_until` | When the metadata's certificates expire | The rotation deadline -- re-paste fresh IdP metadata before this date |

The `saml_provider_uuid` output is a record value for audit, not a composition input.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Corporate IdP trust anchor** -- the one-time setup connecting any SAML 2.0 identity provider to the account: create the provider from the published metadata, then wire roles to it with `sts:AssumeRoleWithSAML` trust policies. Corporate sign-ins get short-lived role sessions; long-lived IAM user credentials retire. Start from the **Corporate IdP** preset.

**Okta federation** -- the most common pairing: Okta's AWS Account Federation app signs users into IAM roles, with per-group role mapping managed on the Okta side. Set the app's Identity Provider ARN to this provider's `provider_arn` output and map groups to `role_arn,provider_arn` pairs. Start from the **Okta Federation** preset.

**One provider, many roles** -- a single trust anchor serves every federated role in the account (admin, developer, read-only); the IdP's group-to-role mapping decides who may assume which. New roles add a trust policy naming the same `provider_arn` -- the provider itself never changes.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the roles federated users assume; their trust policies name this provider's ARN
- [**AWS IAM User**](/infra-catalog/aws-iam-user) -- what federation replaces: long-lived per-human credentials give way to short-lived role sessions
