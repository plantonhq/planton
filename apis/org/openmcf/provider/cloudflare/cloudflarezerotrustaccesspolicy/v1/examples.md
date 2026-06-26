# CloudflareZeroTrustAccessPolicy examples

## Allow a corporate email domain

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessPolicy
metadata:
  name: allow-staff
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: allow-staff
  decision: allow
  include:
    - emailDomain:
        domain: example.com
```

## Reference a reusable group, with approval and MFA

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessPolicy
metadata:
  name: admins
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: admins
  decision: allow
  sessionDuration: 1h
  include:
    - group:
        id:
          valueFrom:
            kind: CloudflareZeroTrustAccessGroup
            name: platform-admins
            fieldPath: status.outputs.group_id
  approvalRequired: true
  approvalGroups:
    - approvalsNeeded: 1
      emailAddresses: [security-lead@example.com]
  mfaConfig:
    allowedAuthenticators: [security_key]
```

## Service-token-only access (non_identity)

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessPolicy
metadata:
  name: ci-access
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: ci-access
  decision: non_identity
  include:
    - anyValidServiceToken: {}
```

## Bypass Access for health checks

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessPolicy
metadata:
  name: bypass-healthcheck
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: bypass-healthcheck
  decision: bypass
  include:
    - ip:
        ip: 203.0.113.0/24
```
