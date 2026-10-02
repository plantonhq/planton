# GcpKmsKeyHandle Guide

The judgment this guide protects: a key handle is a request, not a key you manage. It must match its resource's type and location exactly, it outlives this block in Google, and the key it returns keeps billing until a key administrator destroys it.

## How a handle becomes a key

With Autokey on for the project (a `GcpKmsAutokeyConfig` on the project or a folder above it), creating a handle asks Autokey for a key for `resourceTypeSelector` in `location`. Autokey creates the `autokey` key ring in that location if needed, creates or reuses an HSM key at the service's granularity (one per resource, or one per location for services such as Cloud Run and Secret Manager), grants the service's per-project service agent encrypt and decrypt, and returns the key in `kms_key`. The protected resource then names that key. Under dedicated-project storage the key lives in the folder's key project; otherwise in the handle's project.

## Wiring it to a resource

Reference `status.outputs.kms_key` from the resource's key field; those fields list `GcpKmsKeyHandle` beside `GcpKmsKey`. The handle's `location` must equal the resource's, and its `resourceTypeSelector` must be the resource's type -- a bucket's handle cannot key a disk.

## Lifecycle

Every field is immutable. Google cannot delete a handle, so destroy only forgets it, and the key stays (still protecting the resource, still billing as an HSM key version). A handle with the same ID cannot be created again, so give a recreated handle a new `keyHandleName`. Retire keys through Cloud KMS when the resources they protect are gone.

## Permissions

Creating a handle needs only `cloudkms.keyHandles.create` (in `roles/cloudkms.autokeyUser`), which Google includes in the resource-creation roles -- developers never hold key-administration rights. This block also enables the Cloud KMS API on the handle's project, which same-project storage requires.
