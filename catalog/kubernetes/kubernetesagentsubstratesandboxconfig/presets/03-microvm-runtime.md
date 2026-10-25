# MicroVM Runtime

The microVM sandbox class's runtime, as Agent Substrate v0.6.0 defines it: cloud-hypervisor and virtiofsd with Kata's guest kernel and root filesystem, at the checksums upstream pins for each architecture, staged in your own bucket under a prefix per architecture. It is the class default, so a microVM worker pool that names no config resolves it. Each actor then runs in its own lightweight virtual machine, a harder boundary than gVisor's, on nodes that expose KVM.

## When to Use

- Your agents run code you trust least, and each should get a hardware-virtualized boundary.
- Your workloads need what gVisor does not offer: a full Linux kernel, or a root filesystem snapshot (`ROOTFS` fidelity is microVM-only).

## How It Works

The files are not baked into the worker image: the node agent fetches each one through the control plane's object-store plugin, verifies its SHA-256 and caches it by hash. The checksums here are the asset sets upstream's `hack/microvm-assets/assemble.sh` produces at v0.6.0, one per architecture: run it with `ARCH=amd64` and with `ARCH=arm64` (each on a Linux host of that architecture), and upload each run's four files (`cloud-hypervisor`, `virtiofsd`, `vmlinux`, `rootfs.img`) under `kata-assets/<arch>/`. The prefix per architecture is what lets one bucket serve a cluster with both kinds of nodes; upstream's own manifest puts one architecture's files at a single path per cluster.

## Key Configuration Choices

- **Your bucket, outside the snapshot location.** The plugin serves only its own backend's scheme (gs:// on GCS, s3:// on S3) and refuses runtime files under the control plane's snapshot location; use a separate prefix or bucket.
- **The class default.** The annotation makes this config what microVM pools resolve when they name none. Drop it and name the config in each pool if you run several microVM configs.
- **Not upstream's install script.** `hack/install-microvm-deps.sh --install` stages one architecture's files at upstream's paths and also applies upstream's own `microvm` SandboxConfig, carrying the same class-default annotation; whichever of the two is created last becomes the default. Stage the files yourself as above, and let this config be the only microVM default.
- **Nodes with KVM.** Workers of a microVM pool land only on nodes whose node agent advertises KVM (nested virtualization on cloud VMs).

## Prerequisites

- **Agent Substrate** on the cluster, with its storage on the bucket that holds these files (or on the same provider).
- Each architecture's asset set staged in the bucket under `kata-assets/amd64/` and `kata-assets/arm64/` (one of them is enough when every node shares an architecture; drop the other block).
- A **worker pool** of class `microvm` on KVM-capable nodes.

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<bucket-url>` | The bucket holding the staged files, with its scheme: `gs://<bucket>` on GCS, `s3://<bucket>` on S3 (the control plane's storage provider). |
