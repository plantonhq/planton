# OpenStackApplicationCredential Examples

## Minimal Credential (Auto-generated Secret)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackApplicationCredential
metadata:
  name: ci-pipeline-cred
spec:
  description: CI/CD pipeline automation credential
```

## Credential with Scoped Roles

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackApplicationCredential
metadata:
  name: deploy-cred
spec:
  description: Deployment credential with limited roles
  roles:
    - member
    - reader
```

## Credential with Access Rules (Fine-grained)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackApplicationCredential
metadata:
  name: compute-readonly-cred
spec:
  description: Read-only access to compute API
  access_rules:
    - path: /v2.1/servers
      method: GET
      service: compute
    - path: /v2.1/servers/*
      method: GET
      service: compute
    - path: /v2.1/flavors
      method: GET
      service: compute
```

## Credential with Expiration

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackApplicationCredential
metadata:
  name: temp-access-cred
spec:
  description: Temporary credential for contractor access
  expires_at: "2027-03-01T00:00:00Z"
  roles:
    - reader
```

## Credential with User-provided Secret

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackApplicationCredential
metadata:
  name: known-secret-cred
spec:
  description: Credential with a pre-agreed secret
  secret: my-secure-shared-secret
```

## Unrestricted Credential (Use with Caution)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackApplicationCredential
metadata:
  name: admin-automation-cred
spec:
  description: Admin automation credential (can create sub-credentials)
  unrestricted: true
```
