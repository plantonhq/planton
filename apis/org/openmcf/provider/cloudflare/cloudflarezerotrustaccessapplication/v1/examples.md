# CloudflareZeroTrustAccessApplication examples

## Self-hosted web app with a referenced policy

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: internal-dashboard
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: internal-dashboard
  type: self_hosted
  domain: dashboard.example.com
  sessionDuration: 24h
  policies:
    - policy:
        valueFrom:
          kind: CloudflareZeroTrustAccessPolicy
          name: allow-staff
          fieldPath: status.outputs.policy_id
      precedence: 1
```

## SaaS app over OIDC

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: grafana
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: grafana
  type: saas
  saasApp:
    authType: oidc
    redirectUris: [https://grafana.example.com/login/generic_oauth]
    grantTypes: [authorization_code]
    scopes: [openid, email, profile]
  policies:
    - policy:
        valueFrom:
          kind: CloudflareZeroTrustAccessPolicy
          name: allow-staff
          fieldPath: status.outputs.policy_id
```

## App launcher

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: launcher
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: company-launcher
  type: app_launcher
  landingPageDesign:
    title: "Welcome"
    message: "Choose an application to continue."
```

## Self-hosted app with CORS and private destinations

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: api
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: internal-api
  type: self_hosted
  domain: api.example.com
  corsHeaders:
    allowedMethods: [GET, POST, OPTIONS]
    allowedOrigins: [https://app.example.com]
    allowCredentials: true
    maxAge: 600
  destinations:
    - type: public
      uri: https://api.example.com
    - type: private
      cidr: 10.0.0.0/24
      l4Protocol: tcp
      portRange: "8080"
  policies:
    - policy:
        valueFrom:
          kind: CloudflareZeroTrustAccessPolicy
          name: allow-staff
          fieldPath: status.outputs.policy_id
```
