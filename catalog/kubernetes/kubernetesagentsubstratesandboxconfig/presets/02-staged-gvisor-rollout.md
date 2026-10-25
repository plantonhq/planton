# Staged gVisor Rollout

A new gVisor release as the default, with the previous release kept Enabled beside it, so moving forward is one change and moving back is the same change reversed. Every new actor boots the new release; an actor suspended under the previous one restores its memory on it, and an actor running at the moment keeps its sandbox until it stops.

## When to Use

- You are moving your agents to a newer gVisor release and want a rollback that takes one apply.
- You want the previous release one field away while the new one proves itself under real agents.

## How It Works

`defaultVersion` decides what every new actor on the pools running this config boots, and every node downloads the default ahead of time. To roll back, point `defaultVersion` at `gvisor-2026-10-17` and apply: nodes that ran it as the default still hold its files, so the rollback does not wait on a download there. If the new release proves bad, set its `state` to `Disabled` after rolling back, so no one makes it the default again; once you no longer want the previous release as a rollback target and no actor you care about is suspended on it, remove it: an actor suspended on a removed version resumes from its files without its memory.

## Key Configuration Choices

- **Both versions listed and Enabled during the rollout.** Only an Enabled version can be the default, so keeping the previous one listed and Enabled is what makes the rollback one field, and what lets actors suspended on it keep their memory.
- **Real checksums.** Replace the zero checksums with the SHA-256 of each new tarball (gVisor publishes one beside every release file); a wrong checksum refuses the download on every node instead of booting unknown bytes.
- **Named versions.** Name the new version after the release it pins (`gvisor-2026-11-14`), the way the previous one is named.

## Prerequisites

- **Agent Substrate** on the cluster.
- Worker nodes that can download both releases' tarballs.

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `gvisor-rollout` (name) | The config's name in the cluster, which worker pools name; rename it freely. |
| `gvisor-new-release` (version and default) | The new version's name, a DNS label named after the release it pins, such as `gvisor-2026-11-14`; change it in both places. |
| `<new-amd64-tarball-url>` | The new release's `gvisor.tar.zstd` for amd64, such as `gs://gvisor/releases/nightly/<date>/x86_64/gvisor.tar.zstd`. |
| `<new-arm64-tarball-url>` | The new release's `gvisor.tar.zstd` for arm64, such as `gs://gvisor/releases/nightly/<date>/aarch64/gvisor.tar.zstd`. |
| `0000…0000` (both checksums) | The SHA-256 of each new tarball, which gVisor publishes beside every release file. |
