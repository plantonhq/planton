# Azure Container App Environment Managed Certificate

Provisions a TLS certificate Azure issues and renews end to end for one custom domain -- free, domain-validated, and rotated before expiry. The managed certificate lives on the Container App Environment and attaches to the matching Azure Container App Custom Domain binding asynchronously once issued. It covers exactly one hostname -- no wildcards, no additional SANs (bring your own certificate for those). Deployment blocks on domain validation, so the required DNS records must resolve before this resource deploys.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Managed Certificate** -- on the referenced Container App Environment, issued for the subject hostname after Azure proves domain control (HTTP token or CNAME check)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureContainerAppEnvironment** to provision on. Reference its `environment_id` output via ValueFromRef.
- **Public DNS records that ALREADY resolve** -- the create operation polls until domain validation succeeds (or fails around the 30-minute mark), so publish these BEFORE deploying: a TXT record at `asuid.{subjectName}` carrying the app's `custom_domain_verification_id` output, and the routing record for the hostname itself (a CNAME to the app's `ingress_fqdn` for CNAME validation, or an HTTP-reachable A-record path for HTTP validation). With the zone on Azure DNS, both are AzureDnsRecord resources.
- **The domain binding, typically first** -- deploy the AzureContainerAppCustomDomain certificate-less (its managed flow), then this certificate for the same hostname; Azure attaches it to the binding once issued.

## Deploy

### Console

Open the deployment store, find **Azure Container App Environment Managed Certificate**, and click **Deploy**. The creation wizard leads hostname-first -- the subject and validation method with a live preview of the exact DNS records validation polls for -- then the placement (environment + a one-click dots-to-hyphens name suggestion) and tags. Start from the **CNAME-Validated Certificate** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureContainerAppEnvironmentManagedCertificate
metadata:
  name: app-example-com-managed-cert
  org: acme-corp
  env: prod
spec:
  certificateName: app-example-com
  containerAppEnvironmentId:
    value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.App/managedEnvironments/prod-apps-env"
  subjectName: app.example.com
  domainControlValidation: CNAME
```

```shell
planton apply -f managed-certificate.yaml
```

This issues a free, CNAME-validated certificate for `app.example.com` on the environment, provided the asuid TXT and CNAME records already resolve. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying the full managed-TLS story as one environment, compose the binding, the DNS records, and this certificate in the same InfraPipeline -- the binding deploys certificate-less first, and Azure attaches the issued certificate out of band:

```yaml
spec:
  subjectName: app.example.com
  containerAppEnvironmentId:
    valueFrom:
      kind: AzureContainerAppEnvironment
      name: apps-env
      fieldPath: status.outputs.environment_id
```

The InfraPipeline resolves the dependency graph, deploys the environment, binding, and DNS records first, then provisions the certificate against the published records.

## Key Configuration

These are the most important decisions when configuring a managed certificate. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Subject name** -- the ONE hostname the certificate covers, the same hostname the domain binding uses (attachment matches on it). No wildcards -- `*.example.com` is bring-your-own territory.

**Domain control validation** -- how Azure proves you control the domain before issuing. `CNAME` is the standard choice for subdomains (the domain already CNAMEs to the app's ingress FQDN for routing); `HTTP` serves a token on the domain and fits apex domains routed by A record. Left unspecified, Azure deploys its HTTP default.

**Certificate name** -- the resource's identity on the environment, conventionally the subject with dots as hyphens (`app-example-com`). Nothing references it by name in this composition -- attachment matches on the hostname. Only `tags` update in place; every other change re-issues the certificate (Azure re-issues rather than mutating managed certificates), with fresh domain validation against the published records.

**Keep the DNS records published** -- renewal re-validates domain control. Removing the asuid TXT or routing record turns the next automatic renewal into an outage.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureContainerAppEnvironment** | `containerAppEnvironmentId` | `status.outputs.environment_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `certificate_id` | Azure Resource Manager ID of the managed certificate | Operational tooling; the domain binding's `managed_certificate_id` output echoes it once Azure attaches |
| `validation_token` | The domain-validation token Azure generated | Relevant while validation is pending; informational once issued (the deployment waits for validation) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Free TLS for a subdomain** -- the everyday flow: bind `app.example.com` certificate-less, publish the TXT + CNAME records, deploy this certificate CNAME-validated. Start from the **CNAME-Validated Certificate** preset.

**Apex domain** -- `example.com` cannot CNAME: route it with an A record to the environment's static IP and validate over HTTP. Start from the **HTTP-Validated Certificate** preset.

**Many hostnames** -- one managed certificate per hostname, each free and self-renewing; a wildcard that covers them all is the bring-your-own kind's job.

## Works With

- [**Azure Container App Environment**](/infra-catalog/azure-container-app-environment) -- where the certificate is provisioned
- [**Azure Container App Custom Domain**](/infra-catalog/azure-container-app-custom-domain) -- the binding Azure attaches the issued certificate to
- [**Azure Container App**](/infra-catalog/azure-container-app) -- carries the `custom_domain_verification_id` and `ingress_fqdn` outputs the DNS records need
- [**Azure DNS Record**](/infra-catalog/azure-dns-record) -- publishes the validation TXT and routing records when the zone is on Azure DNS
- [**Azure Container App Environment Certificate**](/infra-catalog/azure-container-app-environment-certificate) -- the bring-your-own alternative for wildcards, SANs, and mandated CAs
