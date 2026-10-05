# Azure Front Door Origin

Deploys a Front Door origin -- one backend inside an origin group, with hostname, ports, priority/weight traffic role, certificate name check, and optional Private Link. Origins are many-per-group with independent lifecycles: a regional stamp adds its backend to a shared group without touching other regions' origins, a blue/green cutover swaps origins one at a time, and each Private Link origin carries its own connection-approval workflow -- which is why the origin is a first-class kind referencing its group rather than a list folded into the group's spec.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Front Door Origin** -- a named child of an origin group pointing at a backend hostname, IPv4, or IPv6 address
- **Traffic role** -- HTTP/HTTPS ports (Azure defaults 80/443), priority (1-5, lower serves first), weight (1-1000 within a priority tier), and the enabled drain switch
- **Certificate name check** -- always sent by the module; validation of the origin certificate against the connect hostname stays on unless explicitly disabled
- **Private Link connection** (created only when `privateLink` is set; PREMIUM profiles only) -- a private-endpoint connection to an App Service, storage endpoint, Container Apps environment, Application Gateway, or Private Link Service

ARM does not support tags on origins, so no Azure tags are applied here.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **Planton Runner** -- required when using Runner-based credential delivery.

### Azure Subscription

- **An Azure Front Door Origin Group** the origin nests under (`originGroupId`).
- **A backend hostname** -- an App Service default hostname, a storage endpoint, or any reachable FQDN or IP (`hostName`).
- **A PREMIUM Front Door profile** (only for Private Link) -- Azure rejects Private Link origins on STANDARD at apply time; the sku lives on the profile, so the spec cannot check it statically.

## Deploy

### Console

Open the deployment store, find **Azure Front Door Origin**, and click **Deploy**. The wizard walks you through the parent group and name, the backend address (hostname and host header), the traffic role (ports, priority, weight), and optional Private Link. Start from the **App Service Origin** preset in the [Presets](#presets) tab for the common web-app pairing.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFrontDoorOrigin
metadata:
  name: my-app-origin
  org: acme-corp
  env: prod
spec:
  originGroupId:
    valueFrom:
      kind: AzureFrontDoorOriginGroup
      name: my-app-backends
      fieldPath: status.outputs.origin_group_id
  originName: primary-app
  hostName:
    value: "acme-app.azurewebsites.net"
```

```shell
planton apply -f front-door-origin.yaml
```

This creates a public origin at Azure's defaults -- ports 80/443, priority 1, weight 500, certificate name check on -- and leaving `originHostHeader` unset sends the origin's own hostname as the Host header, the correct pairing for multi-tenant Azure backends like App Service. An Infra Job tracks the provisioning in real time.

### InfraChart

When the backend deploys in the same InfraPipeline, reference its hostname output instead of a literal:

```yaml
spec:
  originGroupId:
    valueFrom:
      kind: AzureFrontDoorOriginGroup
      name: my-app-backends
      fieldPath: status.outputs.origin_group_id
  originName: primary-app
  hostName:
    valueFrom:
      kind: AzureLinuxWebApp
      name: my-web-app
      fieldPath: status.outputs.default_hostname
  priority: 1
  weight: 500
```

The InfraPipeline resolves the dependency graph, deploys the origin group and the web app first, then provisions the origin with the resolved ARM ID and hostname.

## Key Configuration

These are the most important decisions when configuring an origin. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Hostname** -- `hostName` is required: a DNS hostname, IPv4, or IPv6 address. Reference an Infra Component's hostname output (an AzureLinuxWebApp's `default_hostname`, an AzureStorageAccount's `primary_web_host`) when the backend is part of the same deployment, or pass a literal for anything outside it. No kind dominates origin backends, so references declare their kind explicitly.

**Host header** -- leaving `originHostHeader` unset sends the origin's own `hostName` -- exactly right for multi-tenant Azure services (App Service, Container Apps, Functions, Storage static sites), which route BY Host header. Override only when the backend expects the client-facing domain instead, and then make sure it can actually serve it.

**Certificate name check** -- `certificateNameCheckEnabled` defaults to true; keep it on. Disabling it accepts ANY valid certificate from the origin -- a man-in-the-middle door. Azure requires it true when `privateLink` is configured, and the spec enforces that pairing.

**Priority and weight** -- equal priorities load-balance; distinct priorities express active/passive failover (traffic only reaches priority-2 origins when every priority-1 origin is unhealthy). Weight (default 500) splits traffic among siblings at the SAME priority -- a 950/50 split is the classic canary. A canary accidentally placed at priority 2 receives nothing until the fleet fails.

**Enabled** -- `enabled: false` drains the origin (health probes stop, load balancing skips it) without deleting it -- the maintenance and cutover switch.

**Private Link** -- keeps origin traffic off the public internet so the backend can disable public access entirely. Requires the target's own region in `location` (private-link connections are regional even though Front Door is global), the target in `privateLinkTargetId` -- a literal ARM ID starting with `/subscriptions/`, or a reference to a Linux web app, function app, storage account, Container Apps environment, Application Gateway or Private Link Service in the same environment -- and a `targetType` for every target except a Private Link Service (whose ARM ID is itself the attachment point). After deploy, the target's owner must approve the pending private-endpoint connection before traffic flows.

**ForceNew fields** -- `originGroupId` and `originName` both fix the origin's ARM identity at creation; changing either replaces the origin.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureFrontDoorOriginGroup** | `originGroupId` | `status.outputs.origin_group_id` |
| **AzureLinuxWebApp** (or any backend kind) | `hostName` / `originHostHeader` | `status.outputs.default_hostname` (kind-specific hostname outputs) |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `origin_id` | ARM resource ID of the origin | AzureFrontDoorRoute's `originIds` list -- deploy sequencing so the route applies only after its backends exist |
| `origin_name` | The origin's name within its group | Operator tooling, portal cross-reference |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**App Service backend** -- the most common Front Door origin, and the shape where every default is exactly right: Host header falls back to the origin hostname (which App Service routes by), certificate check on, ports 80/443. Start from the **App Service Origin** preset.

**Weighted canary** -- a low-weight origin beside the main backend at the SAME priority (~5% of traffic at weight 50 vs 950). Ramping is a weight change -- an in-place update -- and rollback is deleting one resource, which restores 100% to the main origin without touching it. Start from the **Weighted Canary Origin** preset.

**Private Link origin** -- Front Door reaches the backend over Private Link and the backend disables public network access entirely: defense in depth with no direct-to-origin bypass of the WAF or edge policies. The trade: PREMIUM profile pricing plus a manual connection approval on the target before traffic flows. Start from the **Private Link Origin** preset.

**Active/passive failover** -- a same-region or cross-region standby at priority 2 behind the primary at priority 1; the standby receives nothing until every priority-1 origin fails health probes (probe behavior lives on the origin group).

## Works With

- [**Azure Front Door Origin Group**](/infra-catalog/azure-front-door-origin-group) -- the parent pool that owns load-balancing and health-probe policy
- [**Azure Front Door Route**](/infra-catalog/azure-front-door-route) -- lists this origin's `origin_id` so route deployment sequences after the backends exist
- [**Azure Linux Web App**](/infra-catalog/azure-linux-web-app) -- the most common backend; its `default_hostname` feeds `hostName`
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- static-website backends via `primary_web_host`
- [**Azure Container App Environment**](/infra-catalog/azure-container-app-environment) -- a Private Link target (`targetType: MANAGED_ENVIRONMENTS`)
- [**Azure Private Link Service**](/infra-catalog/azure-private-link-service) -- fronts internal load balancers as a Private Link target with no `targetType`
