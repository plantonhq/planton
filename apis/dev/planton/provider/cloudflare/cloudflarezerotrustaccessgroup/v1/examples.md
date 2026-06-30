# CloudflareZeroTrustAccessGroup examples

## Corporate email domain

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareZeroTrustAccessGroup
metadata:
  name: staff
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: staff
  include:
    - emailDomain:
        domain: example.com
```

## IdP group with a country requirement

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareZeroTrustAccessGroup
metadata:
  name: platform-admins
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: platform-admins
  include:
    - okta:
        name: platform-admins
        identityProviderId: 99999999-aaaa-bbbb-cccc-123456789012
  require:
    - geo:
        countryCode: US
```

## Group of groups (composition)

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareZeroTrustAccessGroup
metadata:
  name: all-engineering
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: all-engineering
  include:
    # Reference other groups by their output ID.
    - group:
        id:
          valueFrom:
            kind: CloudflareZeroTrustAccessGroup
            name: backend-team
            fieldPath: status.outputs.group_id
    - group:
        id:
          valueFrom:
            kind: CloudflareZeroTrustAccessGroup
            name: frontend-team
            fieldPath: status.outputs.group_id
```

## Risk-based exclusion

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareZeroTrustAccessGroup
metadata:
  name: low-risk-staff
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: low-risk-staff
  include:
    - emailDomain:
        domain: example.com
  exclude:
    - userRiskScore:
        userRiskScore:
          - high
```
