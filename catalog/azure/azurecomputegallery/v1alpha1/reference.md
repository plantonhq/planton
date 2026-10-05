# AzureComputeGallery

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureComputeGallerySpec** defines an Azure Compute Gallery -- the
shared library an organization keeps its approved VM images in.
Image definitions (AzureComputeGalleryImage) live inside a gallery,
each with published, region-replicated versions; VMs and scale sets
deploy from those versions.

A gallery is free at rest: it bills nothing itself (image versions
bill for the storage their replicas consume). By default a gallery
is private to the subscriptions it is shared with through RBAC; the
optional sharing block opens it to a community audience.

## Example

```yaml
# Deep-shape example for docs and offline validation: a
# Community-shared gallery exercising the whole sharing tree.
# References are literal values so the manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureComputeGallery
metadata:
  name: test-compute-gallery
  id: test-compute-gallery
  org: test-org
  env: test
spec:
  resourceGroup:
    value: test-rg
  name: platform.images
  region: eastus
  description: The platform team's approved golden images
  sharing:
    permission: Community
    communityGallery:
      eula: https://example.com/image-eula
      prefix: acmeimages
      publisherEmail: images@acme.example
      publisherUri: https://acme.example
  tags:
    costCenter: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.sharing` | `AzureComputeGallerySharing` |  |  |  |
| `spec.sharing.permission` | `string` | yes |  |  |
| `spec.sharing.communityGallery` | `AzureComputeGalleryCommunitySharing` |  |  |  |
| `spec.sharing.communityGallery.eula` | `string` | yes |  |  |
| `spec.sharing.communityGallery.prefix` | `string` | yes |  |  |
| `spec.sharing.communityGallery.publisherEmail` | `string` | yes |  |  |
| `spec.sharing.communityGallery.publisherUri` | `string` | yes |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group the gallery lives in. Can be a literal
string or a reference to an AzureResourceGroup output.

**ForceNew**: changing this destroys and recreates the gallery.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The gallery's name -- up to 80 characters of letters, numbers,
dots, and underscores. Dashes are NOT allowed (unlike most Azure
names); dots separate logical segments by convention
(e.g. "platform.images").

**ForceNew**: changing this destroys and recreates the gallery.

- rule: Gallery names allow up to 80 letters, numbers, dots, and underscores (no dashes)
- rule: {"required":true}

### spec.region

`string` · required

The Azure region the gallery is created in, e.g. "eastus". Image
versions replicate to their own target regions independently --
the gallery's region is where the definition metadata lives.

**ForceNew**: changing this destroys and recreates the gallery.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.description

`string`

A description of the gallery's purpose shown in the portal and
returned by the API. Updatable in place.

### spec.sharing

`AzureComputeGallerySharing`

How the gallery is shared beyond RBAC. Omitted means Private
(RBAC-only) -- the right posture for almost every gallery.

**ForceNew**: the ENTIRE sharing configuration is fixed at
creation -- changing any part of it destroys and recreates the
gallery.

- rule: community_gallery must be set when permission is Community

### spec.sharing.permission

`string` · required

The sharing mode: "Private" (RBAC only), "Groups" (shared to
subscriptions/tenants added as sharing groups), or "Community"
(published publicly under a community gallery name -- requires
the community_gallery block).

- rule: {"required":true,"string":{"in":["Community","Groups","Private"]}}

### spec.sharing.communityGallery

`AzureComputeGalleryCommunitySharing`

The community-gallery publishing details. Required when
permission is "Community" (the provider rejects Community sharing
without it); ignored by Azure for the other modes.

### spec.sharing.communityGallery.eula

`string` · required

The end-user license agreement URL or text shown to community
consumers before they deploy from the gallery.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.sharing.communityGallery.prefix

`string` · required

The public name prefix -- 5-16 letters and numbers. Azure appends
a generated suffix to form the gallery's public community name.

- rule: The community prefix must be 5-16 letters and numbers
- rule: {"required":true}

### spec.sharing.communityGallery.publisherEmail

`string` · required

The publisher's contact email shown to community consumers.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.sharing.communityGallery.publisherUri

`string` · required

The publisher's URI (homepage or documentation) shown to
community consumers.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.tags

`map<string, string>`

Tags to apply to the gallery, merged over the Planton-derived
metadata tags (user values win on key conflicts).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureComputeGallery, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.gallery_id` | `string` | The gallery's Azure Resource Manager ID. |
| `status.outputs.gallery_name` | `string` | The gallery's name -- what image definitions (AzureComputeGalleryImage) reference as their gallery. |
| `status.outputs.unique_name` | `string` | The globally unique name Azure assigns the gallery (distinct from the user-chosen name; used in cross-tenant and community addressing). |
| `status.outputs.community_gallery_name` | `string` | The public community-gallery name generated from the sharing prefix. Empty unless the gallery is Community-shared. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureComputeGalleryImage | `spec.galleryName` | `status.outputs.gallery_name` |

## See Also

- [Overview](../README.md)
