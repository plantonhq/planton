# GcpColabRuntimeTemplate Guide

The judgment this guide protects: a template is a policy for every notebook machine your team will start, and almost all of it is fixed at creation. Decide machine size, network posture, and idle shutdown up front, publish one template per size, and never let a GPU template ship without idle shutdown.

## What a template governs

Every Colab Enterprise runtime is created from a template. The template fixes the machine (`machineSpec`), the data disk (`dataPersistentDiskSpec`), the network (`networkSpec`), idle shutdown (`idleTimeout`), end-user-credential access (`eucDisabled`), Shielded VM Secure Boot (`enableSecureBoot`), network tags, and labels. It also carries the software -- environment variables, a post-startup script, and the Colab image release -- and the encryption key, which are the only things (with the display name) that change in place.

## Fixed at creation, labels included

Google fixes a template's labels at creation: changing any label replaces the template. So does changing the machine, disk, network, idle, EUC, Secure Boot, tags, or description. Runtimes already created keep running; new ones use the new template. Plan a template per machine size rather than editing one.

## Sizing and GPUs

`machineSpec.machineType` picks the Compute Engine machine; GPUs (`acceleratorType` + `acceleratorCount`, set together) attach to N1 machines and the accelerator-optimized families. GPU availability varies by region.

## Idle shutdown is the cost control

A runtime bills while it runs, idle or not. `idleTimeout` (seconds ending in `s`; between 600s and 86400s, or `0s` to disable) stops a runtime nobody is using. Every GPU template should set it.

## Network posture

With no `networkSpec`, runtimes join the default network with internet access. For private runtimes, reference a `GcpVpcNetwork` and `GcpSubnetwork` and set `enableInternetAccess: false`; the subnetwork then needs Private Google Access (or Cloud NAT) for Google APIs and package installs. `networkTags` let VPC firewall rules target notebook machines.

## Credentials and encryption

`eucDisabled: true` keeps users' own Google credentials out of the runtime, so notebooks act only as the runtime's service identity -- the posture for shared or regulated data. `kmsKeyName` encrypts runtime disks under your key (the Compute Engine and Vertex AI service agents need `roles/cloudkms.cryptoKeyEncrypterDecrypter`). Environment values in `softwareConfig.env` are stored in plain text: never put secrets there.
