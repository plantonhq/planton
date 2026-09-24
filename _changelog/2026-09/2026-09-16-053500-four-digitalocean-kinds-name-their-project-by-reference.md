# Four DigitalOcean kinds name their project by reference

## What changed

- **`project_id` on `DigitalOceanApp`, `DigitalOceanFunction`, `DigitalOceanLoadBalancer`, and `DigitalOceanDatabaseCluster` is a typed reference to `DigitalOceanProject`** (default wiring: the project's `project_id` output), each returning at a new field number with the old one reserved. All four carried the same comment -- "a typed reference will land when the Project kind is forged" -- and the kind was forged; the autoscale pool already referenced it.
- Each Pulumi module reads the literal through the reference; the database cluster's e2e manifest carries its project under `value:`; the four catalog pages' Consumes tables name the project (the functions page gains the InfraChart arm the table requires), and the App guide's project section reads in the present.

## Why

A project is DigitalOcean's account-level grouping, and these four kinds are the ones DigitalOcean's API places in a project at create time. Naming it by reference gives the console a picker, the set lane an order, and the diagram a line. The droplet, the Kubernetes cluster, the volume, the zone, the bucket, the reserved IP, and the VPC carry no `project_id` at all -- their provider resources have none, and membership for them is a separate project-resources assignment in both engines -- so they follow in their own change, and the project becomes a container only when its whole membership can name it.

## How to check

```bash
for k in digitaloceanapp digitaloceanfunction digitaloceanloadbalancer digitaloceandatabasecluster; do go test ./catalog/digitalocean/$k/v1alpha1/; done
grep -n 'DigitalOceanProject,' catalog/digitalocean/{digitaloceanapp,digitaloceanfunction,digitaloceanloadbalancer,digitaloceandatabasecluster}/v1alpha1/spec.proto
```
