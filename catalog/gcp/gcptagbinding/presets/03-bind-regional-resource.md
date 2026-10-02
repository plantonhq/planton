# Bind Regional Resource

Tags a Cloud SQL instance by its full resource name. The instance is
regional, so `location` names its region and the module uses the
location-scoped binding Google serves from that region.

## What it configures

- `parent.resourceName` — the `//{service}.googleapis.com/...` full resource
  name of any taggable resource.
- `location` — the resource's region (or zone for a zonal resource such as
  a Compute instance); required for regional and zonal resources, forbidden
  for organizations, folders, and projects.

## Adjust before deploying

- **`resourceName`** and **`location`** — your instance and its region.
- **`tagValue`** — the `tagValues/{id}` literal or a `GcpTagValue` reference.

## When to choose something else

Projects and folders are global and take the **Bind Project** or **Bind
Folder** preset with no `location`.
