# Azure Compute Gallery

Deploys an Azure Compute Gallery -- the shared library an organization keeps its approved VM images in. Image definitions (AzureComputeGalleryImage) live inside the gallery, each with published, region-replicated versions that VMs and scale sets deploy from. The gallery is free at rest (image versions bill for their replica storage), private by default with RBAC-granted access, and its entire sharing configuration is fixed at creation -- changing it replaces the gallery.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Compute gallery** -- the gallery itself, with its description and optional sharing profile (Private, Groups, or Community with its public publishing identity)
- **Azure Tags** -- your governance tags merged over the Planton-derived resource tags (organization, environment, resource ID); a user tag with the same key wins

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** -- referenced through `resourceGroup` as a literal name or an AzureResourceGroup ValueFromRef.
- **A publisher identity** (only for Community sharing) -- an EULA, a 5-16 character public name prefix, a publisher email, and a publisher URI, all shown to community consumers and all fixed at creation.

## Deploy

### Console

Open the deployment store, find **Azure Compute Gallery**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Private Gallery** or **Community Gallery** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureComputeGallery
metadata:
  name: platform-images
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "acme-prod-rg"
  name: platform.images
  region: eastus
  description: "Approved golden images for platform workloads"
```

```shell
planton apply -f compute-gallery.yaml
```

This creates a private, RBAC-only gallery -- free at rest, ready for image definitions and published versions. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the resource group:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      fieldPath: status.outputs.resource_group_name
  name: platform.images
  region: eastus
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the gallery with the resolved group name.

## Key Configuration

These are the most important decisions when configuring a compute gallery. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Decide the sharing posture before you create** -- The whole sharing tree -- Private, Groups, or Community, and every community publishing detail -- is create-only: changing any of it replaces the gallery, and replacing a gallery means recreating every definition and republishing every version inside it. Almost every gallery should start (and stay) Private, with RBAC granting consuming subscriptions read access; reach for Community only when you genuinely publish images to the world.

**Name galleries like packages, not like resources** -- Gallery names forbid dashes (unlike most Azure names) and allow dots; the convention that works is a package-style name (`platform.images`, `data.golden`) that reads well in the image ARM IDs every consumer sees. The name is ForceNew and image references embed it everywhere, so a rename is an estate-wide migration -- choose once.

**One gallery per audience, not per image** -- The gallery is the ACL boundary and the publishing boundary: definitions inside it share its visibility. Split galleries by who consumes them (platform-wide, team-private, public), never by OS or application -- that is what image definitions are for. One well-named gallery with many definitions is easier to govern than many galleries with one definition each.

**Community publishing is a storefront, not a switch** -- Community sharing attaches your EULA, publisher email, and URI to a public identity Azure generates from your prefix (read it back from the `community_gallery_name` output). Everything published in that gallery becomes publicly deployable. Keep community galleries separate from internal ones, review what gets published into them, and remember the prefix is create-only -- the public name cannot be rebranded in place.

**The region holds metadata, not replicas** -- The gallery's region is where definition metadata lives; image versions replicate to their own target regions independently. Put the gallery in your primary region and let versions decide their own replication footprint.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `gallery_name` | The gallery's name | AzureComputeGalleryImage's `galleryName` -- how image definitions attach to the gallery |
| `gallery_id` | The gallery's ARM ID | RBAC scope for granting consuming subscriptions read access |
| `unique_name` | The globally unique name Azure assigns | Cross-tenant and community image addressing |
| `community_gallery_name` | The generated public community name | Public image references for community consumers (empty unless Community-shared) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Organization golden-image library** -- one private gallery per platform team, RBAC read access granted to consuming subscriptions, image definitions per OS/role inside it, and a build pipeline publishing versions. Start from the **Private Gallery** preset.

**Public image distribution** -- a dedicated Community-shared gallery for a vendor or open-source project distributing prebuilt VM images under a public identity. Keep it strictly separate from internal galleries -- everything published in it is publicly deployable. Start from the **Community Gallery** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the gallery lives in
- [**Azure Compute Gallery Image**](/infra-catalog/azure-compute-gallery-image) -- the image definitions created inside the gallery, referencing its `gallery_name` output
- [**Azure Virtual Machine**](/infra-catalog/azure-virtual-machine) -- deploys from the published, region-replicated image versions
