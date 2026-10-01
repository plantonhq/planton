# GcpKmsAutokeyConfig Guide

The judgment this guide protects: Autokey is a switch on a scope, not a key. Choose the storage model deliberately, finish the one-time setup before the first key handle, and remember that destroying this block turns Autokey off for everything beneath it.

## Two storage models

- **Same-project storage (`RESOURCE_PROJECT`).** Each key lives in the project of the resource it protects. Works on a folder or a project. The only setup is the Cloud KMS API in each project -- this block enables it on a project configuration's project, and `GcpKmsKeyHandle` enables it on its own project -- and Google creates the Cloud KMS service agent the first time it is needed. Use it for most teams.
- **Dedicated-project storage (`DEDICATED_KEY_PROJECT`).** Every key for every project in a folder lives in one key project that only the key administrators can manage. Available on folders only, and it needs a one-time setup outside this block: create the key project's Cloud KMS service agent (`gcloud beta services identity create --service=cloudkms.googleapis.com --project=KEY_PROJECT`) and grant it `roles/cloudkms.admin` on the key project. This block enables the Cloud KMS API on the key project but does not grant an admin role from a configuration file -- that grant is a deliberate security decision.

## Scope and inheritance

A folder configuration is inherited by every project beneath it; a project configuration overrides its folder. To opt one project out of a folder's Autokey, give the project its own configuration with `keyProjectResolutionMode: DISABLED`. Empty `scope` means the connection's project.

## The lifecycle

Google keeps exactly one Autokey configuration per folder and per project. Applying this block takes over the existing configuration -- there is no import step -- so never declare two blocks for one scope. Under the default `deletionPolicy` (`DELETE`), destroy clears the configuration: Autokey is off for the scope until something configures it again. Keys Autokey already created are ordinary Cloud KMS keys; they stay and keep protecting their resources. Use `ABANDON` to retire the block while leaving Autokey on, or `PREVENT` to make destroy fail.

## What Autokey keys look like

HSM protection, AES-256-GCM, a one-year rotation period (an administrator may change it), created in a key ring named `autokey` in the resource's location, one key per resource or per location depending on the service. They bill as Cloud HSM key versions. Locations without Cloud HSM cannot use Autokey -- create those keys with `GcpKmsKey`.
