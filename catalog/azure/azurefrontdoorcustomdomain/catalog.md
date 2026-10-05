# Azure Front Door Custom Domain

Deploys a custom domain inside an Azure Front Door (Standard/Premium) profile -- your own hostname (www.example.com, or the wildcard *.example.com) served at Microsoft's edge instead of the generated `*.azurefd.net` endpoint hostname. TLS is always on: Azure manages the certificate end to end by default, or the domain serves a bring-your-own certificate wrapped by an Azure Front Door Secret. The domain deploys immediately in a pending-validation state and goes live in two moves -- publish the exported validation token as a TXT record at `_dnsauth.<hostname>`, then CNAME the hostname to the endpoint so traffic arrives.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Front Door Custom Domain** -- a named child of the profile, created in a pending-validation state (deployment does not block on DNS proof) and exporting a `validation_token` for the DNS challenge
- **TLS configuration** -- an Azure-managed DV certificate (free, auto-rotated) or a customer certificate served from a Front Door secret, optionally with a pinned cipher-suite policy
- **DNS zone binding** -- created only when `dnsZoneId` references an AzureDnsZone; Front Door then watches that zone for the validation records

The validation TXT record and the traffic CNAME are not created by this module: they are AzureDnsRecord resources when the zone lives in Azure DNS, or records you publish at your external provider.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **Planton Runner** -- required when using Runner-based credential delivery.

### Azure Subscription

- **An Azure Front Door Profile** the domain nests under.
- **Control of the hostname's DNS** -- validation publishes a TXT record under it.
- **For wildcards or hostnames over 64 characters**: an AzureFrontDoorSecret wrapping a customer certificate (Azure's managed certificates cannot cover them).

## Deploy

### Console

Open the deployment store, find **Azure Front Door Custom Domain**, and click **Deploy**. The wizard walks you through the hostname (wildcards flagged live when they force a customer certificate), the dot-free ARM resource name, the TLS decision, and the DNS zone binding with a live preview of the exact validation records. Start from the **Managed-Certificate Domain** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFrontDoorCustomDomain
metadata:
  name: www-example-com
  org: acme-corp
  env: prod
spec:
  profileId:
    valueFrom:
      kind: AzureFrontDoorProfile
      name: cdn-profile
      fieldPath: status.outputs.profile_id
  domainName: www-example-com
  hostName: www.example.com
  dnsZoneId:
    valueFrom:
      kind: AzureDnsZone
      name: example-com-zone
      fieldPath: status.outputs.zone_id
  tls: {}
```

```shell
planton apply -f front-door-custom-domain.yaml
```

The empty `tls` block deploys Azure's managed certificate, creating the domain in a pending-validation state ready for the TXT-record challenge at `_dnsauth.www.example.com`. An Infra Job tracks the provisioning in real time.

### InfraChart

In an InfraChart, wire the customer-certificate composition through `valueFrom`:

```yaml
spec:
  profileId:
    valueFrom:
      kind: AzureFrontDoorProfile
      name: cdn-profile
      fieldPath: status.outputs.profile_id
  domainName: www-example-com
  hostName: www.example.com
  tls:
    certificateType: CUSTOMER_CERTIFICATE
    secretId:
      valueFrom:
        kind: AzureFrontDoorSecret
        name: wildcard-cert-secret
        fieldPath: status.outputs.secret_id
```

The InfraPipeline resolves the dependency graph, provisioning the profile and the secret before the domain that references them.

## Key Configuration

These are the most important decisions when configuring a Front Door custom domain. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Hostname** -- the FQDN the domain serves; only the FIRST label may be the wildcard `*`. ForceNew: the hostname IS the domain's identity. With an Azure-managed certificate the hostname caps at 64 characters and cannot be a wildcard.

**Domain resource name** -- the ARM label, NOT the hostname (dots forbidden); convention is the hostname with dots as hyphens. ForceNew: renaming replaces the domain and restarts validation.

**Certificate type** -- unspecified deploys MANAGED_CERTIFICATE (Azure's default). CUSTOMER_CERTIFICATE requires `tls.secretId` and unlocks wildcards, EV/OV, and org-mandated CAs; the spec enforces the pairing in both directions (a managed domain must not reference a secret).

**Cipher suite** -- unset serves Azure's default suite set. Pin TLS12_2022/TLS12_2023 or hand-pick suites (CUSTOMIZED: at least one TLS 1.2 suite; TLS 1.3 suites both-or-none). The minimum TLS version is not configurable -- every domain floors at TLS 1.2.

**DNS zone binding** -- set `dnsZoneId` when the hostname's DNS lives in Azure DNS so Front Door watches the zone for the validation TXT record; leave it unset with external DNS and publish the token at your provider instead. The binding updates in place, so migrating DNS into Azure later does not replace the domain.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureFrontDoorProfile** | `profileId` | `status.outputs.profile_id` |
| **AzureDnsZone** (optional) | `dnsZoneId` | `status.outputs.zone_id` |
| **AzureFrontDoorSecret** (customer certificate) | `tls.secretId` | `status.outputs.secret_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `custom_domain_id` | ARM resource ID of the domain | AzureFrontDoorRoute.`customDomainIds`; AzureFrontDoorSecurityPolicy.`domainIds` |
| `host_name` | The hostname the domain serves | The CNAME source at your DNS provider |
| `validation_token` | The DNS challenge (public by design) | The `_dnsauth.<host>` TXT record's value |
| `expiration_date` | When the current token expires (RFC-3339) | Re-validation scheduling |

## Common Patterns

**Single hostname on a managed certificate** -- the right default for most sites: an empty `tls` block gives you a free, auto-rotated DV certificate with zero certificate operations (no vault, no secret, no renewal calendar). The trade: no wildcards, no EV/OV, hostname capped at 64 characters. Start from the **Managed-Certificate Domain** preset.

**Wildcard domain for multi-tenant platforms** -- when every tenant gets a subdomain, a wildcard hostname with CUSTOMER_CERTIFICATE is mandatory (managed certificates do not cover wildcards). One wildcard certificate serves many domains: point each domain's `secretId` at the same AzureFrontDoorSecret, and rotation becomes a single Key Vault operation. Start from the **Wildcard Domain with Bring-Your-Own Certificate** preset.

**Pinned ciphers for compliance baselines** -- when PCI-DSS, FedRAMP, or an internal crypto baseline names allowed suites, pin them. Prefer the predefined TLS12_2023 set when it satisfies the baseline (predefined sets receive Microsoft's updates automatically); reach for CUSTOMIZED only when the auditor's list is exact -- a pinned list is yours to maintain. Start from the **Hardened Cipher Policy** preset.

**Validate, point, then attach** -- the domain goes live in a fixed sequence: publish the `validation_token` output as a TXT record at `_dnsauth.<hostname>` (an AzureDnsRecord when the zone lives in Azure DNS), CNAME the hostname to the endpoint's `host_name` once Azure flips the domain to approved, then list the domain in an AzureFrontDoorRoute's `customDomainIds` -- the route side owns the attachment. A domain left unvalidated past `expiration_date` needs a fresh token; Azure regenerates it on re-validation.

## Works With

- [**Azure Front Door Profile**](/infra-catalog/azure-front-door-profile) -- the parent container the domain nests under via `profileId`
- [**Azure Front Door Secret**](/infra-catalog/azure-front-door-secret) -- wraps the bring-your-own certificate referenced by `tls.secretId` with CUSTOMER_CERTIFICATE
- [**Azure DNS Zone**](/infra-catalog/azure-dns-zone) -- the zone Front Door watches for validation when `dnsZoneId` is set
- [**Azure DNS Record**](/infra-catalog/azure-dns-record) -- hosts the validation TXT record and the traffic CNAME in Azure DNS
- [**Azure Front Door Route**](/infra-catalog/azure-front-door-route) -- attaches this domain to serving paths through its `customDomainIds`
- [**Azure Front Door Security Policy**](/infra-catalog/azure-front-door-security-policy) -- scopes a WAF to this domain through the same `custom_domain_id`
