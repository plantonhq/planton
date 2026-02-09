# OpenStackKeypair Examples

## Import an ED25519 Key

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackKeypair
metadata:
  name: ed25519-key
spec:
  public_key: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleKey user@workstation"
```

## Import an RSA Key

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackKeypair
metadata:
  name: rsa-key
spec:
  public_key: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDxyz... user@workstation"
```

## Generate a Keypair (No Public Key)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackKeypair
metadata:
  name: generated-key
spec: {}
```

Retrieve the generated private key:
```bash
# Pulumi
pulumi stack output private_key --show-secrets

# Terraform
terraform output -raw private_key
```

## Keypair with Region Override

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackKeypair
metadata:
  name: regional-key
spec:
  public_key: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleKey user@workstation"
  region: "RegionTwo"
```

## Keypair with Organization Metadata

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackKeypair
metadata:
  name: platform-deploy-key
  org: acme-corp
  env: production
  labels:
    team: platform
    purpose: deployment
spec:
  public_key: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleKey deploy@acme"
```

## CLI Usage

```bash
# Deploy with provider config file
openmcf apply --manifest keypair.yaml -p openstack-creds.yaml

# Deploy with stored credentials (auto-resolved)
openmcf apply --manifest keypair.yaml

# Preview changes
openmcf plan --manifest keypair.yaml -p openstack-creds.yaml

# Destroy
openmcf destroy --manifest keypair.yaml -p openstack-creds.yaml
```
