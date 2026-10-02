# GcpCloudBuildRepository Guide

The judgment this guide protects: a repository link belongs to the team that owns the repository, not to whoever owns the connection -- so it is its own block, referencing the connection by name.

## Placement

The link lives in its connection's project and region. Both modules parse those from `parentConnection`'s full name, so they can never disagree; reference the connection's `name` output rather than writing it by hand. Triggers that build from the repository must use the same region.

## Readiness

Google links a repository only through a connection whose installation is complete (`installation_stage` `COMPLETE`) and whose credentials can see the repository. A GitHub connection that is still waiting for its app installation refuses the link.

## Lifecycle

Every field is immutable: a new `remoteUri` or `repositoryId` replaces the link. Destroy removes the link from Cloud Build -- triggers that reference it stop working -- and never touches the repository on the host.
