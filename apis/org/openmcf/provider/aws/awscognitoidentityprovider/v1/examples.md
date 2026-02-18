# AwsCognitoIdentityProvider Examples

Progressive examples for federating external identity providers into an Amazon Cognito User Pool.

---

## 1. Google OAuth (Minimal)

The most common social IdP. Cognito auto-discovers authorize, token, and OIDC endpoints from Google's well-known configuration.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCognitoIdentityProvider
metadata:
  name: google-idp
  labels:
    openmcf.org/provisioner: pulumi
    app.kubernetes.io/component: auth
spec:
  region: us-west-2
  # User Pool ID — pool must exist before creating this IdP
  userPoolId:
    value: us-east-1_Ab1Cd2EfG
  providerName: Google
  providerType: Google
  google:
    clientId: "123456789-abc.apps.googleusercontent.com"
    clientSecret: "GOCSPX-xxxxxxxxxxxxxxxxxxxxxxxx"
    authorizeScopes: "email profile openid"
  # Optional: AWS applies defaults if omitted
  attributeMapping:
    email: email
    username: sub
```

---

## 2. Facebook Login with api_version

Facebook uses comma-separated scopes (unlike space-separated for other providers). Pin a specific Graph API version for stability.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCognitoIdentityProvider
metadata:
  name: facebook-idp
  labels:
    openmcf.org/provisioner: pulumi
    app.kubernetes.io/component: auth
spec:
  region: us-west-2
  userPoolId:
    value: us-east-1_Ab1Cd2EfG
  providerName: Facebook
  providerType: Facebook
  facebook:
    clientId: "1234567890123456"
    clientSecret: "abcdef1234567890abcdef1234567890"
    # Facebook uses comma-separated scopes
    authorizeScopes: "email,public_profile"
    # Pin Graph API version; omit to use Cognito's default
    apiVersion: "v17.0"
  attributeMapping:
    email: email
    username: id
```

---

## 3. Sign in with Apple (All 5 Required Fields)

Apple uses a private-key-based authentication model. All five fields are required.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCognitoIdentityProvider
metadata:
  name: apple-idp
  labels:
    openmcf.org/provisioner: pulumi
    app.kubernetes.io/component: auth
spec:
  region: us-west-2
  userPoolId:
    value: us-east-1_Ab1Cd2EfG
  providerName: SignInWithApple
  providerType: SignInWithApple
  signInWithApple:
    # Apple Services ID from Apple Developer Portal
    clientId: "com.example.app.service"
    # 10-character alphanumeric Team ID
    teamId: "ABCD123456"
    # Key ID for the private key
    keyId: "XYZ789KEY"
    # PEM-format private key (use secret ref in production)
    privateKey: |
      -----BEGIN PRIVATE KEY-----
      MIGTAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBHkwdwIBAQQg...
      -----END PRIVATE KEY-----
    authorizeScopes: "email name"
  attributeMapping:
    email: email
    username: sub
```

---

## 4. OIDC Minimal (Okta)

For OIDC providers, only `client_id` and `oidc_issuer` are required. Cognito auto-discovers all endpoints from the issuer's `.well-known/openid-configuration`.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCognitoIdentityProvider
metadata:
  name: okta-idp
  labels:
    openmcf.org/provisioner: pulumi
    app.kubernetes.io/component: auth
spec:
  region: us-west-2
  userPoolId:
    value: us-east-1_Ab1Cd2EfG
  providerName: CorpOkta
  providerType: OIDC
  oidc:
    clientId: "0oa1bc2d3ef4gh5ij6kl"
    # Cognito fetches authorize, token, userinfo, JWKS from this URL
    oidcIssuer: "https://dev-123456.okta.com/oauth2/default"
    # Optional: add scopes for richer profile
    authorizeScopes: "openid email profile"
  attributeMapping:
    email: email
    username: sub
```

---

## 5. OIDC Full (Azure AD)

All optional URL overrides and `attributes_request_method` for IdPs that require POST for the userinfo endpoint.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCognitoIdentityProvider
metadata:
  name: azure-ad-idp
  labels:
    openmcf.org/provisioner: pulumi
    app.kubernetes.io/component: auth
spec:
  region: us-west-2
  userPoolId:
    value: us-east-1_Ab1Cd2EfG
  providerName: AzureAD
  providerType: OIDC
  oidc:
    clientId: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    clientSecret: "secret~value"
    oidcIssuer: "https://login.microsoftonline.com/tenant-id-here/v2.0"
    authorizeScopes: "openid email profile"
    # Some IdPs require POST for userinfo
    attributesRequestMethod: "POST"
    # Optional: override auto-discovered URLs if needed
    authorizeUrl: "https://login.microsoftonline.com/tenant-id/v2.0/authorize"
    tokenUrl: "https://login.microsoftonline.com/tenant-id/v2.0/token"
    attributesUrl: "https://graph.microsoft.com/oidc/userinfo"
    jwksUri: "https://login.microsoftonline.com/tenant-id/discovery/v2.0/keys"
  attributeMapping:
    email: email
    username: sub
    given_name: given_name
    family_name: family_name
```

---

## 6. SAML with metadata_url and SLO

Enterprise federation with SAML 2.0. Uses `metadata_url` for IdP metadata and enables single logout (SLO).

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCognitoIdentityProvider
metadata:
  name: corp-saml-idp
  labels:
    openmcf.org/provisioner: pulumi
    app.kubernetes.io/component: auth
spec:
  region: us-west-2
  userPoolId:
    value: us-east-1_Ab1Cd2EfG
  providerName: CorpAD
  providerType: SAML
  saml:
    # Cognito fetches metadata from this URL
    metadataUrl: "https://idp.corp.example.com/metadata/saml"
    # Sign user out of IdP when they sign out of Cognito
    idpSignOut: true
    # Optional: enable IdP-initiated SSO
    idpInit: false
  # SAML claims use URIs; map to Cognito attributes
  attributeMapping:
    email: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
    username: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/nameidentifier"
    given_name: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"
    family_name: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"
```

---

## 7. Google with valueFrom (Infra Chart Pattern)

Reference the User Pool from another OpenMCF resource instead of hardcoding. Use this when composing resources in an infra chart.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCognitoIdentityProvider
metadata:
  name: google-idp
  labels:
    openmcf.org/provisioner: pulumi
    app.kubernetes.io/component: auth
spec:
  region: us-west-2
  # Reference AwsCognitoUserPool outputs — resolves at deploy time
  userPoolId:
    valueFrom:
      kind: AwsCognitoUserPool
      name: my-user-pool
      fieldPath: status.outputs.user_pool_id
  providerName: Google
  providerType: Google
  google:
    clientId: "123456789-abc.apps.googleusercontent.com"
    clientSecret: "GOCSPX-xxxxxxxxxxxxxxxxxxxxxxxx"
    authorizeScopes: "email profile openid"
  attributeMapping:
    email: email
    username: sub
```

---

## Enabling Federated Sign-In

After creating an identity provider, add its `provider_name` to the User Pool Client's `supported_identity_providers` list:

```yaml
# In your AwsCognitoUserPoolClient or User Pool app client config:
supportedIdentityProviders:
  - COGNITO        # native pool auth
  - Google        # matches providerName from example 1
  - CorpOkta      # matches providerName from example 4
```
