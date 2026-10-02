# GcpComputeImage Guide

The judgment this guide protects: an image is a frozen build -- you version it, never edit it -- and consumers should boot its family, not its name.

## Versioning through families

Everything except labels replaces the image, so treat each build as a new image: a fresh `imageName` (`web-base-20261001`) in a stable `family` (`web-base`). A disk or instance that boots `projects/{project}/global/images/family/web-base` always gets the newest non-deprecated image of the family, so rolling a new build forward is declaring one more image block; retiring an old build is destroying its block. Images bill storage until deleted. A consumer that references this image with `valueFrom` (`status.outputs.self_link`, or `status.outputs.image_id` for a GKE secondary boot disk) pins this exact build and orders after it; a consumer that writes the family path as a literal follows the newest build instead.

## Choosing the source

Exactly one source: `sourceDisk` (the classic golden-image path -- configure a VM's boot disk, stop the VM, image the disk), `sourceImage` (copy or re-key an image, including a public family such as Debian's), `sourceSnapshot`, or `rawDisk` (a `disk.raw` tarball in Cloud Storage, the import and Packer path). Guest OS features, licenses, and storage locations are inherited from the source when unset; set `guestOsFeatures` explicitly for an imported disk (`UEFI_COMPATIBLE`, `GVNIC`, `SECURE_BOOT`, ...).

## Encryption

`kmsKey` encrypts the image with a Cloud KMS key or an Autokey key handle (Autokey serves Compute Engine images); the Compute Engine service agent needs encrypt and decrypt on it. A source that is itself CMEK-encrypted needs its key in the matching `source*Encryption` block. Customer-supplied raw keys are deliberately not modeled -- raw key material does not belong in manifests or state.
