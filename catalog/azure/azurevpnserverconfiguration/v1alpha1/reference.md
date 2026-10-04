# AzureVpnServerConfiguration

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureVpnServerConfigurationSpec** defines a VPN server
configuration -- the reusable "who may connect and how" policy for
point-to-site VPN: which authentication methods remote users sign in
with (Entra ID, certificate, RADIUS), which certificates are trusted
or revoked, the tunnel protocols offered, and optional per-user-group
policy groups. It deploys no gateway and costs nothing to keep; an
AzurePointToSiteVpnGateway is born pointing at ONE configuration
(fixed on the gateway at its creation) and many gateways may share
the same configuration.

**Authentication types drive the required blocks** (ARM's create
contract, enforced here so the error lands in seconds): "AAD" needs
the aad_authentication block, "Certificate" needs at least one
client_root_certificates entry, and "Radius" needs the radius block.
Multiple types may be enabled at once -- each brings its block.

**ForceNew fields**: `name`, `region`, and `resource_group` --
everything else updates in place (gateways using the configuration
pick the change up; users reconnect under the new policy).

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVpnServerConfiguration
metadata:
  name: test-vpn-server-configuration
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: remote-workforce
  vpnAuthenticationTypes:
    - Certificate
  clientRootCertificates:
    # A self-signed test root (public data only -- base64 X.509 body).
    # Generated with:
    #   openssl req -x509 -newkey rsa:2048 -keyout root.key -out root.pem \
    #     -days 3650 -nodes -subj "/CN=planton-oss-e2e-p2s-root"
    #   openssl x509 -in root.pem -outform der | base64
    - name: corp-root
      publicCertData: MIIDJzCCAg+gAwIBAgIUI4B0y9OsZnVyco5f2m1SK8vOOKIwDQYJKoZIhvcNAQELBQAwIzEhMB8GA1UEAwwYcGxhbnRvbi1vc3MtZTJlLXAycy1yb290MB4XDTI2MDgwOTIwMjAyM1oXDTM2MDgwNjIwMjAyM1owIzEhMB8GA1UEAwwYcGxhbnRvbi1vc3MtZTJlLXAycy1yb290MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAz0JLtVVB3WlaPydW1NBBRWamIFuvnNKQkSJGe2WzZ+yh8pL+oNaw9J2EyTpcrl0Wsvda7fLrqcWhE1Hw3KxjStsCh+fI37TBpV77oSTJjuwszN1EJRraU2MJuo1BafGBdfCRCufuuseJ8AQppLD/qgH3nrMQ2BnwrO+jaWxnkYdDl91vumDZPiWGctBcmTJp53ZzTICW39wCV1yq1/htK2wvBvvIf5nDQPHhTLfy8d2P5OksiTZWSY+HV3uUXTtrzr1QSLSP9A7OzuoBAvdBF5Fa3PtsZw27WmDDl2iOLM8cEQn3B4B4VWdQ9LCXkWE8MwMQI13yR2QNY+kt90rVywIDAQABo1MwUTAdBgNVHQ4EFgQU8ISSNCctYDUDNyo6vROk8V/LvfowHwYDVR0jBBgwFoAU8ISSNCctYDUDNyo6vROk8V/LvfowDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAK/pl9/Izolm4fGHI/xFvMJ6EPzFiAZ3kyZ0sg2mQZ1woF2Mv8IWjb+vRISVNyt3jPYpXGGNlmpZC1ZOjaK+BOlXsiw35L5Cuegkv+y/dwsH9EQe7lm1vOW2JNrpCITqN6dGm78+YJiLAD7RF29BxhNgh2yNx0RA+3eDduTTL6HlSUf2E7NeMNwfSRvUSUgbbynE+I6p3gq9Kit2hkHXn1uhXTRU93Qrk3GP9xbKr1pn2nOhdR5ucuspNuVS5oY+d0JIv5i+vLkON6gFXKSn5pbcv4j2Q8JdAfJfT6uTcBgIGv8ph32bte18K1MYiCrGpKHsRwOhiidCVFz8SNzpaGg==
  vpnProtocols:
    - OpenVPN
  policyGroups:
    - name: engineering
      isDefault: true
      priority: 0
      policies:
        - name: engineering-cn
          type: CertificateGroupId
          value: engineering
  tags:
    purpose: hack-test
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.vpnAuthenticationTypes` | `[]string` | yes |  |  |
| `spec.aadAuthentication` | `AzureVpnServerConfigurationAadAuthentication` |  |  |  |
| `spec.aadAuthentication.audience` | `string` | yes |  |  |
| `spec.aadAuthentication.issuer` | `string` | yes |  |  |
| `spec.aadAuthentication.tenant` | `string` | yes |  |  |
| `spec.clientRootCertificates` | `[]AzureVpnServerConfigurationClientRootCertificate` |  |  |  |
| `spec.clientRootCertificates[].name` | `string` | yes |  |  |
| `spec.clientRootCertificates[].publicCertData` | `string` | yes |  |  |
| `spec.clientRevokedCertificates` | `[]AzureVpnServerConfigurationClientRevokedCertificate` |  |  |  |
| `spec.clientRevokedCertificates[].name` | `string` | yes |  |  |
| `spec.clientRevokedCertificates[].thumbprint` | `string` | yes |  |  |
| `spec.ipsecPolicy` | `AzureVpnServerConfigurationIpsecPolicy` |  |  |  |
| `spec.ipsecPolicy.dhGroup` | `string` |  |  |  |
| `spec.ipsecPolicy.ikeEncryption` | `string` |  |  |  |
| `spec.ipsecPolicy.ikeIntegrity` | `string` |  |  |  |
| `spec.ipsecPolicy.ipsecEncryption` | `string` |  |  |  |
| `spec.ipsecPolicy.ipsecIntegrity` | `string` |  |  |  |
| `spec.ipsecPolicy.pfsGroup` | `string` |  |  |  |
| `spec.ipsecPolicy.saLifetimeSeconds` | `int32` |  |  |  |
| `spec.ipsecPolicy.saDataSizeKilobytes` | `int32` |  |  |  |
| `spec.radius` | `AzureVpnServerConfigurationRadius` |  |  |  |
| `spec.radius.servers` | `[]AzureVpnServerConfigurationRadiusServer` |  |  |  |
| `spec.radius.servers[].address` | `string` | yes |  |  |
| `spec.radius.servers[].secret` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.radius.servers[].score` | `int32` |  |  |  |
| `spec.radius.clientRootCertificates` | `[]AzureVpnServerConfigurationRadiusClientRootCertificate` |  |  |  |
| `spec.radius.clientRootCertificates[].name` | `string` | yes |  |  |
| `spec.radius.clientRootCertificates[].thumbprint` | `string` | yes |  |  |
| `spec.radius.serverRootCertificates` | `[]AzureVpnServerConfigurationRadiusServerRootCertificate` |  |  |  |
| `spec.radius.serverRootCertificates[].name` | `string` | yes |  |  |
| `spec.radius.serverRootCertificates[].publicCertData` | `string` | yes |  |  |
| `spec.vpnProtocols` | `[]string` |  |  |  |
| `spec.policyGroups` | `[]AzureVpnServerConfigurationPolicyGroup` |  |  |  |
| `spec.policyGroups[].name` | `string` | yes |  |  |
| `spec.policyGroups[].isDefault` | `bool` |  |  |  |
| `spec.policyGroups[].priority` | `int32` |  |  |  |
| `spec.policyGroups[].policies` | `[]AzureVpnServerConfigurationPolicyGroupPolicy` | yes |  |  |
| `spec.policyGroups[].policies[].name` | `string` | yes |  |  |
| `spec.policyGroups[].policies[].type` | `string` |  |  |  |
| `spec.policyGroups[].policies[].value` | `string` | yes |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the configuration object lives in, e.g.
"eastus". By convention the region of the hub whose point-to-site
gateway uses it. Changing the region replaces the object.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the configuration is created in. Can be
a literal resource-group name or a reference to an
AzureResourceGroup's name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The configuration's name, unique within the resource group. Name
it after the policy it expresses ("remote-workforce-entra",
"contractor-certs"). Changing the name replaces the object.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vpnAuthenticationTypes

`[]string` · required

The authentication methods remote users may sign in with -- at
least one of "AAD" (Entra ID), "Certificate", or "Radius" (the
wire values). Each enabled type requires its configuration block
(aad_authentication / client_root_certificates / radius).

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"in":["AAD","Certificate","Radius"]}}}}

### spec.aadAuthentication

`AzureVpnServerConfigurationAadAuthentication`

Entra ID (Azure AD) authentication parameters. Required when
vpn_authentication_types includes "AAD".

### spec.aadAuthentication.audience

`string` · required

The Entra ID application (client) ID VPN clients authenticate
against -- for the Microsoft Azure VPN Client this is the
published well-known application ID.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.aadAuthentication.issuer

`string` · required

The STS issuer URL, e.g. "https://sts.windows.net/{tenant-id}/".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.aadAuthentication.tenant

`string` · required

The tenant URL, e.g.
"https://login.microsoftonline.com/{tenant-id}".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.clientRootCertificates

`[]AzureVpnServerConfigurationClientRootCertificate`

Trusted root certificates for certificate authentication -- a
connecting client's certificate must chain to one of them. At
least one is required when vpn_authentication_types includes
"Certificate".

### spec.clientRootCertificates[].name

`string` · required

The certificate's name on the configuration (unique), e.g.
"corp-root-2026".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.clientRootCertificates[].publicCertData

`string` · required

The root certificate's public data: the base64-encoded X.509
certificate body (the PEM payload WITHOUT the BEGIN/END
CERTIFICATE lines and without line breaks). Public material only
-- the private key never leaves your certificate authority.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.clientRevokedCertificates

`[]AzureVpnServerConfigurationClientRevokedCertificate`

Individually revoked client certificates (by thumbprint) --
blocks a lost or compromised client without rotating the root.

### spec.clientRevokedCertificates[].name

`string` · required

The revocation entry's name on the configuration (unique).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.clientRevokedCertificates[].thumbprint

`string` · required

The revoked client certificate's SHA-1 thumbprint.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.ipsecPolicy

`AzureVpnServerConfigurationIpsecPolicy`

A pinned IPsec/IKE proposal for connecting clients. Leave unset
to offer Azure's default proposal set; pin it only when client
policy demands specific algorithms (every field of a configured
proposal is required -- no partial pinning).

### spec.ipsecPolicy.dhGroup

`string`

The IKE Phase 1 Diffie-Hellman group.

- rule: {"string":{"in":["DHGroup1","DHGroup14","DHGroup2","DHGroup2048","DHGroup24","ECP256","ECP384","None"]}}

### spec.ipsecPolicy.ikeEncryption

`string`

The IKE (Phase 1) encryption algorithm.

- rule: {"string":{"in":["AES128","AES192","AES256","DES","DES3","GCMAES128","GCMAES256"]}}

### spec.ipsecPolicy.ikeIntegrity

`string`

The IKE (Phase 1) integrity algorithm. With GCM IKE encryption,
Azure requires the MATCHING GCM integrity value.

- rule: {"string":{"in":["GCMAES128","GCMAES256","MD5","SHA1","SHA256","SHA384"]}}

### spec.ipsecPolicy.ipsecEncryption

`string`

The IPsec (Phase 2) encryption algorithm.

- rule: {"string":{"in":["AES128","AES192","AES256","DES","DES3","GCMAES128","GCMAES192","GCMAES256","None"]}}

### spec.ipsecPolicy.ipsecIntegrity

`string`

The IPsec (Phase 2) integrity algorithm.

- rule: {"string":{"in":["GCMAES128","GCMAES192","GCMAES256","MD5","SHA1","SHA256"]}}

### spec.ipsecPolicy.pfsGroup

`string`

The Perfect Forward Secrecy group.

- rule: {"string":{"in":["ECP256","ECP384","None","PFS1","PFS14","PFS2","PFS2048","PFS24","PFSMM"]}}

### spec.ipsecPolicy.saLifetimeSeconds

`int32`

The security association lifetime in seconds. The provider puts
no bounds on this resource (unlike the tunnel-level proposals of
a VPN gateway connection, whose provider enforces 300-172799) --
ARM validates the value at deploy time.

- rule: {"int32":{"gte":0}}

### spec.ipsecPolicy.saDataSizeKilobytes

`int32`

The security association size limit in kilobytes. Unbounded by
the provider on this resource; ARM validates at deploy time.

- rule: {"int32":{"gte":0}}

### spec.radius

`AzureVpnServerConfigurationRadius`

RADIUS authentication parameters (servers and their trust
anchors). Required when vpn_authentication_types includes
"Radius".

### spec.radius.servers

`[]AzureVpnServerConfigurationRadiusServer`

The RADIUS servers authentication is forwarded to. ARM expects at
least one server for RADIUS authentication to function (the
provider leaves the list optional; the deploy fails without one).

### spec.radius.servers[].address

`string` · required

The RADIUS server's address, reachable from the gateway.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.radius.servers[].secret

`string | valueFrom` · required · sensitive

The shared secret the gateway authenticates to the RADIUS server
with. Reference a secret rather than embedding the literal in
manifests. ARM never returns this value on reads.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.radius.servers[].score

`int32`

The server's priority score, 1-30 (higher is preferred).

- rule: {"int32":{"lte":30,"gte":1}}

### spec.radius.clientRootCertificates

`[]AzureVpnServerConfigurationRadiusClientRootCertificate`

Root certificates the RADIUS server uses to verify CLIENT
certificates (by thumbprint).

### spec.radius.clientRootCertificates[].name

`string` · required

The certificate's name on the configuration.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.radius.clientRootCertificates[].thumbprint

`string` · required

The root certificate's SHA-1 thumbprint.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.radius.serverRootCertificates

`[]AzureVpnServerConfigurationRadiusServerRootCertificate`

Root certificates clients use to verify the RADIUS SERVER (full
public certificate data).

### spec.radius.serverRootCertificates[].name

`string` · required

The certificate's name on the configuration.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.radius.serverRootCertificates[].publicCertData

`string` · required

The root certificate's public data: base64-encoded X.509 body
(no BEGIN/END lines, no line breaks).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vpnProtocols

`[]string`

The tunnel protocols offered to clients: "IkeV2" and/or "OpenVPN"
(the wire values). Leave empty for Azure's default selection.
Note: policy groups and multiple point-to-site connection
configurations require OpenVPN.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["IkeV2","OpenVPN"]}}}}

### spec.policyGroups

`[]AzureVpnServerConfigurationPolicyGroup`

Policy groups: named member-matching rules (Entra ID group ID,
certificate common name, or RADIUS Azure group ID) a
point-to-site gateway can map to different address pools. Each
group deploys as its own ARM child of the configuration; the
configuration publishes each group's ARM id in the
policy_group_ids output.

- rule: policy names must be unique within the policy group

### spec.policyGroups[].name

`string` · required

The group's name, unique on the configuration. The group's ARM id
surfaces in the configuration's policy_group_ids output under
this name. Changing the name replaces the group.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.policyGroups[].isDefault

`bool`

Mark this group as the default -- members matching no other group
land here. Fixed at creation (changing it replaces the group).

### spec.policyGroups[].priority

`int32`

The group's evaluation priority (lower evaluates first). 0 is the
provider's default.

- rule: {"int32":{"gte":0}}

### spec.policyGroups[].policies

`[]AzureVpnServerConfigurationPolicyGroupPolicy` · required

The member-matching rules of this group. At least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.policyGroups[].policies[].name

`string` · required

The rule's name, unique within the group.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.policyGroups[].policies[].type

`string`

What the rule matches on: "AADGroupId" (an Entra ID group's
object ID), "CertificateGroupId" (the client certificate's common
name), or "RadiusAzureGroupId" (the RADIUS-returned group ID) --
the wire values.

- rule: {"string":{"in":["AADGroupId","CertificateGroupId","RadiusAzureGroupId"]}}

### spec.policyGroups[].policies[].value

`string` · required

The value to match (the group object ID, certificate common
name, or RADIUS group ID -- per type).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.tags

`map<string, string>`

Free-form tags applied to the object, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins.

## Validation Rules

- `aad_block_required_for_aad_type`: vpn_authentication_types includes 'AAD' -- set the aad_authentication block (audience, issuer, tenant)
- `root_cert_required_for_certificate_type`: vpn_authentication_types includes 'Certificate' -- add at least one client_root_certificates entry
- `radius_block_required_for_radius_type`: vpn_authentication_types includes 'Radius' -- set the radius block
- `client_root_certificate_names_unique`: client_root_certificates names must be unique on the configuration
- `client_revoked_certificate_names_unique`: client_revoked_certificates names must be unique on the configuration
- `policy_group_names_unique`: policy_groups names must be unique on the configuration -- each is the key the policy_group_ids output uses

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureVpnServerConfiguration, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.vpn_server_configuration_id` | `string` | The Azure Resource Manager ID of the VPN server configuration -- what a point-to-site VPN gateway references as its vpn_server_configuration_id. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/vpnServerConfigurations/{name} |
| `status.outputs.vpn_server_configuration_name` | `string` | The name of the VPN server configuration. |
| `status.outputs.policy_group_ids` | `map<string, string>` | The ARM ID of each policy group on the configuration, keyed by the group's name from the spec. Example valueFrom fieldPath: status.outputs.policy_group_ids.engineering |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzurePointToSiteVpnGateway | `spec.vpnServerConfigurationId` | `status.outputs.vpn_server_configuration_id` |

## See Also

- [Overview](../README.md)
