# DigitalOcean Volume -- Operational Guide

Judgment calls that matter when you run block storage on DigitalOcean.

## Size is a one-way ratchet

Volumes only expand. A manifest that lowers `sizeGib` fails at plan time — by design, since DigitalOcean has no shrink operation. Growing the volume is safe and online, but the filesystem inside does not grow itself: after an expansion, resize the filesystem from the Droplet (`resize2fs` for ext4, `xfs_growfs` for xfs). Budget sizes with headroom; "grow later" is easy, "grow now under pressure" is an incident.

## Format at creation or format yourself — never both

`filesystemType` formats the volume exactly once, at creation. DigitalOcean never reports the setting back, and both provisioners deliberately ignore later edits to it: changing `filesystemType` on an existing volume neither reformats nor replaces it (the provider alone would replace the volume, destroying the data — the modules take that footgun away). If you leave it unset, the volume arrives raw and you own `mkfs` from the Droplet. Pick one path per volume and stay on it: a volume formatted by hand but declared `ext4` later in the manifest is simply left as it is, and the manifest is then lying about the disk. To reformat, create a new volume.

## Adopting an existing volume is safe, by design

When you import a volume you already own, the three creation-time arguments (`filesystemType`, `initialFilesystemLabel`, `snapshotId`) come back empty — DigitalOcean never reports them. Keep them in the manifest anyway if that is how the volume was made: because the modules ignore changes to them after creation, the first apply after an import is a no-op. This is the difference between "the manifest describes the volume" and "the manifest destroys the volume"; the provider on its own would plan a destroy-and-recreate for exactly this case.

## The label is for mount automation

`initialFilesystemLabel` exists so cloud-init and fstab entries can mount by label (`LABEL=pgdata`) instead of by device path, which shifts across reboots and attach order. If you format at creation, set the label too — retrofitting one later means hand-running `e2label`/`xfs_admin` on the Droplet (editing the manifest field after creation is ignored, like `filesystemType`).

## Description is not editable — yet

The provider marks `description` create-only ("update-ability coming soon" in its own schema), so editing it in the manifest plans a full volume replacement. The plan will say so honestly — read plans on volume changes carefully before approving; a replaced data volume is data loss unless you snapshot first.

## Snapshots restore into new volumes, not in place

`snapshotId` seeds a NEW volume from a point-in-time capture — there is no in-place restore. The new volume inherits the snapshot's region and minimum size. Like the formatting arguments, it acts at creation only and later edits are ignored: restoring a different snapshot means declaring a new volume, never editing this field on an existing one. Snapshots also outlive their source volume: deleting a volume does not delete its snapshots, which keep billing until removed (`doctl compute snapshot list`).

## Attachment lives on the Droplet

This kind creates storage; the Droplet's `volumeIds` list attaches it. That split is deliberate — the volume's lifecycle (usually long) stays independent of any one Droplet's lifecycle (often short). Move a volume between Droplets by editing the Droplets' manifests, never by recreating the volume. Both must be in the same region.
