# DigitalOceanVolume

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanVolumeSpec defines the specification required to create a DigitalOcean block
storage volume, modeling the provider's full argument surface. A block storage volume provides
expandable storage that is attached to Droplets through the Droplet kind's `volume_ids` list;
attachment is a property of the Droplet, never of the volume.

## Example

```yaml
# Example DigitalOceanVolume manifests. Deploy with:
#   planton apply -f manifest.yaml
#
# Document 1 -- the smallest real volume: region + size, left unformatted
# (format it yourself from the Droplet).
#
# Document 2 -- a production-shaped volume: xfs-formatted with a filesystem
# label, described, and tagged. Attach volumes through the Droplet kind's
# volumeIds list; attachment is a property of the Droplet, never the volume.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanVolume
metadata:
  name: example-dovol-minimal
spec:
  volumeName: scratch-data
  region: nyc3
  sizeGib: 50
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanVolume
metadata:
  name: example-dovol-full
spec:
  volumeName: postgres-data
  description: PostgreSQL data volume for production
  region: nyc3
  sizeGib: 500
  filesystemType: xfs
  initialFilesystemLabel: pgdata
  tags:
    - env:prod
    - service:postgres
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.volumeName` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.region` | `enum` | yes |  |  |
| `spec.sizeGib` | `uint32` | yes |  |  |
| `spec.filesystemType` | `enum` |  |  |  |
| `spec.initialFilesystemLabel` | `string` |  |  |  |
| `spec.snapshotId` | `string` |  |  |  |
| `spec.tags` | `[]string` |  |  |  |

## Field Details

### spec.volumeName

`string` · required

The name of the volume. Must be lowercase letters, numbers, and hyphens only,
starting with a letter and ending with a letter or number. Maximum 64 characters.
The name cannot be changed after creation.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"64","pattern":"^[a-z]([a-z0-9-]*[a-z0-9])?$"}}

### spec.description

`string`

(Optional) A free-form description for the volume.
Create-only at the current provider pin: changing the description REPLACES the volume.

### spec.region

`enum` · required

The DigitalOcean region where the volume will be created.
Must match the region of any Droplet that will attach to this volume.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_region_unspecified` -- 0: default / unspecified region
- `nyc3` -- new york 3
- `sfo3` -- san francisco 3
- `fra1` -- frankfurt 1
- `sgp1` -- singapore 1
- `lon1` -- london 1
- `tor1` -- toronto 1
- `blr1` -- bangalore 1
- `ams3` -- amsterdam 3
- `nyc1` -- new york 1
- `nyc2` -- new york 2
- `sfo2` -- san francisco 2
- `syd1` -- sydney 1
- `atl1` -- atlanta 1

### spec.sizeGib

`uint32` · required

The size of the volume in GiB. Volumes can only be EXPANDED after creation: the provider
rejects a shrink at plan time, so lowering this value fails before anything is applied.
DigitalOcean caps volume size at 16 TiB (larger requests fail at the API).

- rule: {"required":true,"uint32":{"gte":1}}

### spec.filesystemType

`enum`

(Optional) The initial filesystem to format the volume with at creation time.
Applied at creation ONLY: DigitalOcean formats the volume once and never reports this
argument back (the resulting filesystem is observable through the separate computed
attributes), and both IaC modules ignore later edits -- changing it on an existing volume
neither reformats nor replaces the volume, which is what keeps an adopted (imported) volume's
data safe when the manifest still describes how it was formatted. To reformat, create a new
volume. Leave unset (unformatted) to format the volume yourself from the Droplet.

Allowed values (use exactly as shown):

- `unformatted`
- `ext4`
- `xfs`

### spec.initialFilesystemLabel

`string`

(Optional) The filesystem label applied when the volume is formatted at creation time
(e.g. "data"). Only meaningful together with `filesystem_type`. Applied at creation ONLY
and never reported back by the API; later edits are ignored, exactly like `filesystem_type`.

### spec.snapshotId

`string`

(Optional) A volume snapshot ID to create this volume from. The new volume inherits the
snapshot's region and minimum size. Applied at creation ONLY and never reported back by the
API; later edits are ignored (a different snapshot means a different volume -- create one).

### spec.tags

`[]string`

(Optional) Tags applied to the volume. Both provisioners apply the union of these tags and
the standard Planton labels. Tags may contain letters, numbers, colons, dashes, and
underscores, up to 255 characters each.

- rule: {"repeated":{"unique":true,"items":{"string":{"pattern":"^[a-zA-Z0-9:\\-_]{1,255}$"}}}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanVolume, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.volume_id` | `string` | The unique identifier (UUID) of the created DigitalOcean volume. The Droplet kind's volume_ids list consumes this output to attach the volume. |
| `status.outputs.urn` | `string` | The uniform resource name (URN) of the volume, e.g. "do:volume:<uuid>". |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| DigitalOceanDroplet | `spec.volumeIds` | `status.outputs.volume_id` |

## See Also

- [Overview](../README.md)
