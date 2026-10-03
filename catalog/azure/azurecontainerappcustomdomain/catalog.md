# Azure Container App Custom Domain

Binds a custom domain to a Container App -- your own hostname (`app.example.com`) serving the app instead of the generated `*.azurecontainerapps.io` address. Azure models the binding as an entry in the app's ingress configuration: every field replaces it when changed, and TLS arrives exactly one of two ways -- an Azure-managed certificate attached asynchronously (the common path, certificate-less by design) or a bring-your-own certificate stored on the app's environment.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Custom Domain Binding** -- on the referenced Container App's ingress, after Azure validates domain ownership against your published DNS records. In the managed flow the deployment deliberately ignores Azure's out-of-band certificate attachment so it never reads as drift; in the bring-your-own flow the referenced certificate serves TLS immediately

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureContainerApp with ingress ENABLED** -- the binding is an ingress entry; there is nothing to attach to otherwise. Reference its `container_app_id` output via ValueFromRef.
- **Public DNS records that ALREADY resolve** -- Azure validates ownership DURING the create: a TXT record at `asuid.{domainName}` carrying the app's `custom_domain_verification_id` output, and a CNAME from the hostname to the app's `ingress_fqdn` (or an A record to the environment's static IP for apex domains). With the zone on Azure DNS, both are AzureDnsRecord resources.
- **For bring-your-own TLS**: an AzureContainerAppEnvironmentCertificate on the app's OWN environment. The managed flow needs no certificate resource up front -- deploy the AzureContainerAppEnvironmentManagedCertificate for the same hostname after this binding.

## Deploy

### Console

Open the deployment store, find **Azure Container App Custom Domain**, and click **Deploy**. The creation wizard leads hostname-first -- the domain and the app, with a live preview of the exact DNS records ownership validation requires -- then the TLS choice as a two-card selector (Azure-managed, certificate-less by design, or bring-your-own with SNI pre-selected). Start from the **Managed-Certificate Domain** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureContainerAppCustomDomain
metadata:
  name: app-example-com-binding
  org: acme-corp
  env: prod
spec:
  domainName: app.example.com
  containerAppId:
    valueFrom:
      kind: AzureContainerApp
      name: my-app
      fieldPath: status.outputs.container_app_id
```

```shell
planton apply -f custom-domain.yaml
```

This creates the certificate-less managed-flow binding: Azure validates ownership against the published DNS records during the create, and the managed certificate you deploy for the same hostname attaches out of band. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, compose the app, the DNS records, this binding, and the certificate in one InfraPipeline:

```yaml
spec:
  domainName: app.example.com
  containerAppId:
    valueFrom:
      kind: AzureContainerApp
      name: my-app
      fieldPath: status.outputs.container_app_id
  containerAppEnvironmentCertificateId:
    valueFrom:
      kind: AzureContainerAppEnvironmentCertificate
      name: app-example-com-cert
      fieldPath: status.outputs.certificate_id
  certificateBindingType: SNI_ENABLED
```

## Key Configuration

These are the most important decisions when configuring a binding. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Domain name** -- one CONCRETE hostname per binding, no wildcard label (the environment's custom DNS suffix is the wildcard mechanism). The hostname is also what a managed certificate matches on when Azure attaches it.

**TLS flow** -- the certificate and its binding type travel together, or neither (Azure rejects a half-pair). Leave BOTH unset for the managed flow: the binding deploys certificate-less and Azure attaches the issued managed certificate asynchronously. Set BOTH for bring-your-own: `containerAppEnvironmentCertificateId` referencing a certificate on the app's own environment, with `SNI_ENABLED` the standard binding type.

**Binding type** -- `SNI_ENABLED` serves TLS via Server Name Indication (the standard choice); `DISABLED` attaches the domain without serving TLS from the certificate (transitional -- browsers reject HTTPS until a certificate binds); `AUTO` lets Azure pick.

**DNS first** -- ownership validation happens during the create. A failing first deploy usually means the asuid TXT or routing record has not propagated yet.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureContainerApp** | `containerAppId` | `status.outputs.container_app_id` |
| **AzureContainerAppEnvironmentCertificate** | `containerAppEnvironmentCertificateId` (bring-your-own) | `status.outputs.certificate_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `custom_domain_id` | The binding's synthetic identifier -- `{container-app-id}/customDomainName/{domain}` (Azure has no standalone ARM resource for the binding) | Operational tooling and audit trails |
| `managed_certificate_id` | ARM ID of the Azure-managed certificate once one attaches | Empty for bring-your-own bindings, and empty on the managed flow until Azure attaches asynchronously -- a liveness signal for the managed TLS story |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Managed TLS end to end** -- bind the hostname certificate-less, publish the DNS records, then deploy a managed certificate for the same hostname; Azure wires them together. Start from the **Managed-Certificate Domain** preset.

**Wildcard-backed hostnames** -- several bindings (`app.example.com`, `api.example.com`) each referencing the SAME wildcard environment certificate. Start from the **BYO-Certificate Domain** preset.

**Apex domain** -- `example.com` routes with an A record to the environment's static IP (it cannot CNAME); the TXT proof and the TLS story are unchanged.

## Works With

- [**Azure Container App**](/infra-catalog/azure-container-app) -- the app whose ingress the domain binds to; carries the verification-ID and FQDN outputs the DNS records need
- [**Azure Container App Environment Certificate**](/infra-catalog/azure-container-app-environment-certificate) -- the bring-your-own certificate the binding serves
- [**Azure Container App Environment Managed Certificate**](/infra-catalog/azure-container-app-environment-managed-certificate) -- the free certificate Azure attaches to this binding out of band
- [**Azure DNS Record**](/infra-catalog/azure-dns-record) -- publishes the ownership TXT and routing records when the zone is on Azure DNS
- [**Azure Container App Environment**](/infra-catalog/azure-container-app-environment) -- its custom DNS suffix is the wildcard mechanism per-app bindings deliberately exclude
