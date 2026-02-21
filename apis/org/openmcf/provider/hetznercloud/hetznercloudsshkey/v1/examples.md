# HetznerCloudSshKey Examples

## Minimal ED25519 Key

The simplest configuration: a single ED25519 public key. ED25519 is the recommended algorithm for new keys.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudSshKey
metadata:
  name: deploy-key
spec:
  publicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleKeyDataHere deploy@ci"
```

---

## RSA Key for Legacy Compatibility

An RSA 4096-bit key for environments that require RSA (older OpenSSH versions, hardware tokens, compliance requirements).

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudSshKey
metadata:
  name: legacy-deploy-key
  org: my-org
  env: production
spec:
  publicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDExampleRsaKeyDataHere deploy@legacy-system"
```

---

## Team Keys with Organizational Metadata

Register individual SSH keys per team member with org and environment context. This enables key lifecycle management — when someone leaves the team, delete their manifest and redeploy.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudSshKey
metadata:
  name: ops-alice
  org: acme-corp
  env: production
  labels:
    team: platform
    role: ops
spec:
  publicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAliceKeyData alice@acme.com"
```

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudSshKey
metadata:
  name: ops-bob
  org: acme-corp
  env: production
  labels:
    team: platform
    role: ops
spec:
  publicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBobKeyDataHere bob@acme.com"
```

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudSshKey
metadata:
  name: ci-runner
  org: acme-corp
  env: production
  labels:
    team: platform
    role: automation
spec:
  publicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAICiRunnerKeyData ci@acme-pipeline"
```

---

## InfraChart Composition with valueFrom

In an infra chart, a `HetznerCloudServer` references SSH keys via `valueFrom` so the key ID is resolved from the SSH key's stack outputs. This eliminates hardcoded IDs and establishes a dependency edge in the DAG.

SSH key manifest:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudSshKey
metadata:
  name: web-deploy-key
  org: my-org
  env: production
spec:
  publicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIWebDeployKey deploy@web"
```

Server manifest referencing the SSH key output:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: web-01
  org: my-org
  env: production
spec:
  serverType: cx22
  image: ubuntu-24.04
  location: fsn1
  sshKeyIds:
    - valueFrom:
        kind: HetznerCloudSshKey
        name: web-deploy-key
        fieldPath: status.outputs.ssh_key_id
```

The `valueFrom` reference ensures that:
1. The SSH key is created before the server
2. The correct numeric ID is passed to the server without manual lookup
3. Replacing the SSH key (changing `publicKey`) propagates to dependent servers on the next apply
