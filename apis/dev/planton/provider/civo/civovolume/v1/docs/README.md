# Deploying Civo Block Storage Volumes: A Production Guide

## Introduction

Block storage is one of those cloud primitives that's simultaneously straightforward and subtle. On the surface, it's simple: you get a virtual disk that acts like a physical hard drive. You partition it, format it, mount it, and use it to store data that needs to outlive any particular VM. Yet the way you provision and manage that storage can mean the difference between a resilient production system and a brittle one plagued by data loss, orphaned resources, and unpredictable costs.

**Civo Volumes** are Civo's answer to block storage—the equivalent of AWS Elastic Block Store (EBS), Google Cloud Persistent Disks, or DigitalOcean Volumes. They're SSD-backed, network-attached volumes that you can create independently and attach to Civo Compute instances or consume dynamically within Civo Kubernetes clusters via the CSI driver. Unlike ephemeral instance disks that vanish when a VM is deleted, Civo Volumes persist. You can detach them from one instance and reattach them to another, making them ideal for databases, application data, logs, and any workload where durability matters.

What makes Civo's block storage compelling is its simplicity and cost-effectiveness. Volumes range from 1 GiB to 16 TiB, priced at roughly $0.10 per GB-month—competitive with the major clouds but without the complexity of performance tiers or provisioned IOPS. There's one storage class, backed by SSDs delivered over a 10 Gbps network. No "gp3 vs io2" decisions. No knob-turning for throughput. You get reliable, performant storage with straightforward pricing.

But simplicity doesn't mean limitation. Civo Volumes integrate seamlessly with both Civo Compute (as attachable disks) and Civo Kubernetes (as PersistentVolumes via the `civo-volume` StorageClass). They're region-scoped, meaning a volume in LON1 stays in LON1, and they can only attach to one instance at a time—no multi-attach scenarios. They also lack native snapshot capabilities on the public cloud, which means backup strategies require a bit more thought than simply clicking "Create Snapshot."

This guide explains the landscape of deployment methods for Civo Volumes, from manual console clicks to production-grade Infrastructure-as-Code, and shows how Planton abstracts these choices into a clean, protobuf-defined API.

---

## The Deployment Spectrum: From Manual to Production

Not all approaches to managing block storage are created equal. Here's how the methods stack up, from what to avoid to what works at scale:

### Level 0: The Manual Console (Anti-Pattern for Production)

**What it is:** Using Civo's web dashboard to click through volume creation, attachment, and deletion.

**What it solves:** Nothing that can't be solved better another way. The console is fine for learning or one-off testing, but as a production practice, it's a recipe for human error. You'll forget to attach volumes, misconfigure regions, or—worse—orphan volumes after deleting instances, leaving them to quietly rack up charges.

**What it doesn't solve:** Repeatability, auditability, version control, automation. If you can't codify it, you can't reliably reproduce it across environments or hand it off to another engineer.

**Verdict:** Use it to explore Civo's interface and understand the workflow, but never for production or even staging environments that matter.

---

### Level 1: CLI Scripting (Better, But Still Brittle)

**What it is:** Using the `civo` CLI to create and manage volumes in shell scripts:

```bash
civo volume create db-data-volume --size 50 --region LON1
civo volume attach db-data-volume my-instance
```

**What it solves:** Automation. You can script provisioning, integrate it into CI/CD, and version-control your scripts. The CLI is synchronous, returns structured output (JSON mode available), and supports all volume operations: create, attach, detach, delete.

**What it doesn't solve:** State management. Scripts don't track what was created or what changed. If a script runs twice, you might create duplicate volumes or fail because resources already exist. Cleanup on failure is manual. There's no declarative model—just imperative commands.

**Verdict:** Acceptable for throwaway dev environments or integration tests where you create and destroy everything in one pass. Not suitable for production, where you need state tracking, idempotency, and rollback capabilities.

---

### Level 2: Direct API Integration (Flexible but High-Maintenance)

**What it is:** Calling Civo's REST API directly from custom tooling or configuration management systems (like Ansible via the `uri` module).

**What it solves:** Maximum flexibility. You can integrate volume management into any HTTP-capable tool. The API exposes create, attach, detach, and delete endpoints with clear parameters: `volume_name`, `size_gb`, `network_id`, `region`.

**What it doesn't solve:** Abstraction. You're managing HTTP calls, handling authentication (API keys in headers), sequencing operations (create before attach), and implementing idempotency yourself. There's no built-in state file to tell you what exists. You're essentially building your own IaC layer.

**Verdict:** Useful if you're building a custom provisioning system or integrating Civo into a broader orchestration framework. But for most teams, higher-level tools (Terraform, Pulumi) handle the API calls and state management for you.

---

### Level 3: Infrastructure-as-Code (Production-Ready)

**What it is:** Using Terraform or Pulumi with Civo's official provider to declaratively define volumes and their lifecycle.

**Terraform example:**

```hcl
provider "civo" {
  token  = var.civo_api_key
  region = "LON1"
}

resource "civo_volume" "db_data" {
  name       = "db-data-volume"
  size_gb    = 50
  network_id = data.civo_network.default.id
}

resource "civo_volume_attachment" "db_attach" {
  instance_id = civo_instance.myserver.id
  volume_id   = civo_volume.db_data.id
}
```

**Pulumi example (Python):**

```python
import pulumi_civo as civo

volume = civo.Volume("db-data",
    name="db-data-volume",
    size_gb=50,
    network_id=default_net_id)
```

**What it solves:** Everything. You get:
- **Declarative configuration**: State what you want, not how to get there
- **State management**: Terraform/Pulumi track what exists and what changed
- **Idempotency**: Running the same config twice produces the same result
- **Plan/preview**: See changes before applying them
- **Version control**: Treat infrastructure as code, with diffs, reviews, and rollbacks
- **Multi-environment support**: Reuse configs for dev/staging/prod with different parameters

**What it doesn't solve:** The underlying limitations of Civo Volumes (no snapshots on public cloud, region-scoped, single-attach only). But it makes managing those constraints predictable and reproducible.

**Verdict:** This is the production standard. Terraform has broader adoption and maturity. Pulumi offers more programming flexibility if your team prefers TypeScript/Python/Go over HCL. Both are solid choices.

---

### Level 4: Kubernetes CSI Dynamic Provisioning (Cloud-Native Integration)

**What it is:** On Civo Kubernetes clusters, the built-in `civo-volume` StorageClass uses the Civo CSI driver to provision volumes on-demand when you create PersistentVolumeClaims.

**Example PVC:**

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysql-data
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: civo-volume
  resources:
    requests:
      storage: 20Gi
```

The CSI driver automatically creates a 20 GiB Civo Volume, attaches it to the node running the pod, formats it (ext4 by default), and mounts it into your container.

**What it solves:** Complete automation for Kubernetes workloads. No manual volume creation. No worrying about node placement or attachment. Just declare what you need in a PVC, and Kubernetes handles the rest.

**What it doesn't solve:** Management outside Kubernetes. If you need to pre-create volumes, manage them independently, or use them with non-K8s instances, you need a lower-level approach (Terraform/Pulumi).

**Verdict:** Perfect for stateful Kubernetes workloads (databases, StatefulSets, etc.). The CSI driver is production-ready and comes with Civo K3s clusters out of the box.

---

## IaC Tool Comparison: Terraform vs. Pulumi

Both Terraform and Pulumi support Civo Volumes in production. Here's how they compare:

### Terraform: The Battle-Tested Standard

**Maturity:** Terraform has been the IaC standard for years. The Civo provider is officially maintained, stable, and well-documented. It covers all essential volume operations: create, attach, detach, delete.

**Configuration Model:** Declarative HCL. You define resources and their relationships, and Terraform figures out the dependency graph and execution order.

**State Management:** Local or remote backends (S3, Terraform Cloud, etc.). State tracks resource IDs, making updates and deletions predictable.

**Strengths:**
- Broad ecosystem and community support
- Familiarity across ops teams
- Straightforward for standard use cases (create volume, attach to instance)
- Clear plan/apply workflow

**Limitations:**
- HCL is less expressive than a full programming language (no complex loops, limited conditionals)
- Complex provisioning logic (e.g., "if prod, create extra volumes") can be verbose

**Verdict:** The default choice for teams already using Terraform or prioritizing stability and ecosystem maturity.

---

### Pulumi: The Programmer's IaC

**Maturity:** Newer than Terraform, but production-ready. The Pulumi Civo provider is built on the Terraform provider (via a bridge), so it has equivalent resource coverage.

**Configuration Model:** Real programming languages (TypeScript, Python, Go). You write infrastructure logic as code, with loops, conditionals, and unit tests.

**State Management:** Pulumi Cloud or self-managed backends (S3, Azure Blob). Similar state tracking to Terraform.

**Strengths:**
- Full programming language expressiveness (easier to build dynamic configs)
- Better for complex provisioning logic or integration with application code
- Native testing frameworks

**Limitations:**
- Smaller community than Terraform
- Requires a runtime (Node.js, Python, etc.)
- Bridged provider means any quirks in Terraform's Civo provider carry over

**Verdict:** Great if your team prefers coding infrastructure in familiar languages or needs complex orchestration logic. Slightly more overhead than Terraform for simple use cases.

---

### Which Should You Choose?

- **Default to Terraform** if you want the most mature, widely-adopted solution with straightforward HCL configs.
- **Choose Pulumi** if you prefer writing infrastructure in TypeScript/Python/Go and need advanced logic (dynamic resource generation, complex conditionals).
- **Both work equally well** for standard volume provisioning. The choice is more about team preference and existing tooling than capability.

---

## Production Essentials: Lifecycle, Backups, and Performance

### Volume Lifecycle: Create → Attach → Detach → Delete

Managing Civo Volumes in production requires understanding the full lifecycle:

1. **Create:** Provision a volume with a name, size (GiB), region, and network ID. The volume starts unattached and unformatted—a raw block device.

2. **Attach:** Bind the volume to a running instance. The OS sees a new device (e.g., `/dev/sda`). You must partition, format (ext4 or xfs), and mount it before use.

3. **Detach:** Safely remove the volume from an instance. **Always unmount the filesystem first** to avoid data corruption. Then call the Civo detach API.

4. **Delete:** Permanently destroy the volume and its data. Only possible when detached. Ensure backups exist before deletion.

**Best Practice:** Automate attachment and formatting via cloud-init scripts or configuration management (Ansible) so volumes are ready for use immediately after provisioning.

---

### Data Protection: The Snapshot Gap

Unlike AWS EBS or DigitalOcean Volumes, **Civo Volumes do not currently offer native snapshots on the public cloud**. This is the most significant operational gap.

**Backup Strategies:**

- **Application-Level Backups:** Use database backup tools (e.g., `pg_dump`, `mysqldump`) to export data to object storage or another volume.
- **Filesystem Backups:** Use `rsync`, `tar`, or similar tools to copy volume data to external storage on a schedule.
- **Volume Cloning:** Manually create a new volume and use `dd` or `rsync` to copy data from the original to the new volume (attach both to a helper instance).
- **Instance Snapshots:** On private Civo installations (CivoStack), instance snapshots can include attached volumes. This feature is not yet available on public cloud.

**Key Takeaway:** You must implement backups at the application or OS level. Budget time for designing and testing backup/restore workflows.

---

### Performance Characteristics

- **Type:** SSD-backed, network-attached storage over 10 Gbps
- **Performance Tier:** Single tier (no gp3 vs io2 complexity). All volumes use the same SSD class.
- **IOPS/Throughput:** Civo doesn't publish specific IOPS limits, but expect performance comparable to DigitalOcean Volumes or AWS gp3 (baseline suitable for most workloads)
- **Scaling:** Larger volumes may have higher throughput ceilings (common in cloud storage). For extreme IOPS, consider striping multiple volumes with LVM or RAID0.

**When to Use Volumes vs. Instance Storage:**
- **Volumes:** Persistent data (databases, user uploads, logs). Data survives instance deletion.
- **Instance Storage:** Ephemeral data (OS, caches, temp files). Faster if local, but lost on instance termination.

---

### Resizing: Expansion Only

Civo supports **offline volume expansion**:
- Detach the volume (or stop pods using it in Kubernetes)
- Resize via API/CLI/Terraform
- Reattach and expand the filesystem (`resize2fs` for ext4, `xfs_growfs` for xfs)

**No shrinking.** Size increases only. Plan capacity carefully, but err on the small side—you can always grow.

---

### Cost Optimization

- **Right-size:** Start with the minimum size you need. You can expand later.
- **Delete orphaned volumes:** If you delete an instance, its attached volumes remain (and keep billing). Regularly audit for unattached volumes.
- **Use IaC for tracking:** Terraform/Pulumi state files help you remember what exists and why.
- **Pricing:** ~$0.10/GB-month (competitive with major clouds). A 500 GB volume costs ~$50/month.

---

## Planton's Approach: Abstraction with Pragmatism

Planton abstracts Civo Volume provisioning behind a clean, protobuf-defined API (`CivoVolume`). This provides a consistent interface across clouds while respecting Civo's unique characteristics.

### What We Abstract

**The `CivoVolumeSpec` includes:**

- **`volume_name`**: Human-readable identifier (lowercase alphanumeric + hyphens, 1-64 chars)
- **`region`**: Civo region (must match instance region for attachment)
- **`size_gib`**: Volume size (1-16,000 GiB)
- **`filesystem_type`**: Optional pre-formatting (`NONE`, `EXT4`, `XFS`). Default: `NONE` (user formats manually)
- **`snapshot_id`**: Optional snapshot reference (for future snapshot support or multi-cloud compatibility)
- **`tags`**: Organizational metadata (for cost allocation, filtering)

This follows the **80/20 principle**: 80% of users need only these fields. Advanced use cases (custom mount options, multi-volume RAID) happen at the OS or application layer, not in the volume spec.

### Default Choices

- **Filesystem Type:** We default to `NONE` to avoid assumptions. Users can specify `EXT4` (safe default for Linux) or `XFS` (better for very large volumes or high-concurrency workloads).
- **No Network ID exposure:** We infer the network from the region or instance context, simplifying the API.
- **Region Matching:** We enforce that volumes and instances must be in the same region (Civo's constraint).

### Under the Hood: Pulumi or Terraform?

Planton currently uses **Pulumi (Go)** for Civo Volume provisioning. Why?

- **Language Flexibility:** Pulumi's Go SDK fits naturally into our broader multi-cloud orchestration (which is also Go-based).
- **Equivalent Coverage:** Pulumi's Civo provider (bridged from Terraform) supports all volume operations we need.
- **Future-Proofing:** Pulumi's programming model makes it easier to add conditional logic, multi-volume strategies, or custom integrations.

That said, Terraform would work equally well. The choice is implementation detail—the protobuf API remains the same.

---

## Configuration Examples: Dev, Staging, Production

### Development: Minimal Volume

**Use Case:** Small test database for a developer's sandbox.

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoVolume
metadata:
  name: dev-db-data
spec:
  volume_name: dev-db-data
  region: LON1
  size_gib: 10
  filesystem_type: EXT4
  tags:
    - env:dev
    - project:myapp
```

**Rationale:**
- 10 GiB is enough for testing
- `EXT4` for simplicity (automatic formatting)
- Tags for organization
- No snapshot ID (ephemeral dev data)

---

### Staging: Medium Volume with Backup Strategy

**Use Case:** Staging database with periodic backups to object storage.

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoVolume
metadata:
  name: staging-db-data
spec:
  volume_name: staging-db-data
  region: NYC1
  size_gib: 100
  filesystem_type: EXT4
  tags:
    - env:staging
    - project:myapp
    - backup:enabled
```

**Rationale:**
- 100 GiB balances cost and capacity
- `backup:enabled` tag signals automated backup jobs (handled externally, since no native snapshots)
- Same region as staging compute instances

---

### Production: Large Volume, XFS, Tagged

**Use Case:** Production database with 1 TiB storage and application-level backups.

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoVolume
metadata:
  name: prod-db-data
spec:
  volume_name: prod-db-data
  region: FRA1
  size_gib: 1000
  filesystem_type: XFS
  tags:
    - env:prod
    - project:myapp
    - criticality:high
    - backup:daily
```

**Rationale:**
- 1 TiB for production scale
- `XFS` for better performance with large files and parallel I/O
- Tags for cost allocation, alerting, and backup automation
- No `snapshot_id` yet (not supported), but tag indicates daily backups via external tooling

---

## Key Takeaways

1. **Civo Volumes are simple, reliable, and cost-effective** block storage—equivalent to EBS, GCP PD, or DigitalOcean Volumes. One tier, SSD-backed, region-scoped.

2. **Manual management is an anti-pattern.** Use IaC (Terraform or Pulumi) for production. The Civo provider is mature and supports all essential operations.

3. **Snapshots are a gap.** Implement backups at the application or OS level. Budget time for designing backup/restore workflows.

4. **The 80/20 config is name, region, size, and filesystem type.** Advanced features (multi-volume RAID, custom mount options) happen at the OS layer.

5. **For Kubernetes, the CSI driver is king.** Use PersistentVolumeClaims and let the `civo-volume` StorageClass handle provisioning automatically.

6. **Planton abstracts the API** into a clean protobuf spec, making multi-cloud deployments consistent while respecting Civo's unique characteristics.

---

## Further Reading

- **Civo Volumes Documentation:** [Civo Docs - Volumes](https://www.civo.com/docs/compute/instance-volumes)
- **Terraform Civo Provider:** [GitHub - civo/terraform-provider-civo](https://github.com/civo/terraform-provider-civo)
- **Civo CSI Driver:** [GitHub - civo/civo-csi](https://github.com/civo/civo-csi)
- **Kubernetes Storage with Civo:** [Civo Docs - Kubernetes Volumes](https://www.civo.com/docs/kubernetes/config/kubernetes-volumes)
- **Civo API Reference:** [Civo API - Volumes](https://www.civo.com/api/volumes)

---

**Bottom Line:** Civo Volumes give you persistent, SSD-backed block storage with straightforward pricing and no performance tier complexity. Manage them with Terraform or Pulumi, implement backups at the application layer, and leverage the CSI driver for Kubernetes. Planton makes this simple with a protobuf API that hides the complexity while exposing the essential configuration you actually need.

