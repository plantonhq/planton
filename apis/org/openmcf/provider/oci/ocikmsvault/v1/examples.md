# OciKmsVault Examples

## Shared HSM Vault

A default vault with shared HSM partition — the most cost-effective option for standard encryption:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciKmsVault
metadata:
  name: dev-vault
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciKmsVault.dev-vault
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vaultType: default_vault
```

## Dedicated HSM Vault with Custom Display Name

A virtual private vault with a dedicated HSM partition for high-throughput production workloads. Uses `valueFrom` to reference a compartment managed by OpenMCF:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciKmsVault
metadata:
  name: prod-vault
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciKmsVault.prod-vault
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-security
      fieldPath: status.outputs.compartmentId
  displayName: "prod-encryption-vault"
  vaultType: virtual_private
```

## External Key Manager Vault

A BYOK/EKMS vault for organizations that must retain key material in a customer-controlled HSM. Connects to the external HSM via IDCS OAuth and a KMS private endpoint:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciKmsVault
metadata:
  name: byok-vault
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciKmsVault.byok-vault
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vaultType: external
  externalKeyManagerMetadata:
    externalVaultEndpointUrl: "https://ekm.corp.example.com/vault/prod"
    oauthMetadata:
      clientAppId: "abcdef1234567890"
      clientAppSecret: "secret-value-here"
      idcsAccountNameUrl: "https://idcs-abc123.identity.oraclecloud.com"
    privateEndpointId: "ocid1.kmsendpoint.oc1..example"
```

## Common Operations

### Retrieve the management endpoint

After deploying a vault, retrieve the management endpoint from stack outputs. OciKmsKey resources use this endpoint to create keys within the vault:

```shell
openmcf outputs -f vault.yaml --output management_endpoint
```

### Move a vault to a different compartment

Update `compartmentId` to the new compartment OCID and re-apply. The vault is moved within OCI without recreation.

## Best Practices

1. **Use `default_vault` unless you need dedicated throughput** — shared HSM is sufficient for most workloads and costs less.
2. **Use `virtual_private` for high-volume crypto operations** — provides isolated throughput limits that are not shared with other tenants.
3. **Use `external` only when regulatory requirements mandate BYOK** — adds operational complexity (IDCS OAuth, private endpoint, third-party HSM management).
4. **Use `valueFrom` references** for `compartmentId` to avoid hardcoding OCIDs and maintain dependency ordering.
5. **Create one vault per security boundary** — production and development workloads should use separate vaults.
