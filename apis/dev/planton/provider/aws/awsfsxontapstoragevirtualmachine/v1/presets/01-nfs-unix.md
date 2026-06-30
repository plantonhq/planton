# NFS-Only UNIX SVM

Basic NFS-only Storage Virtual Machine with UNIX security style. No Active Directory, no SMB. The simplest and most common SVM configuration for Linux/NFS workloads.

## When to Use

- Linux workloads accessing shared storage via NFS
- Kubernetes persistent volumes backed by FSx ONTAP
- Data science and ML pipelines needing shared file access
- Development environments where NFS is sufficient
- Any workload that does not require Windows SMB access

## What It Configures

- **UNIX security style** — UNIX permissions (mode bits, uid/gid) for all volumes
- **No Active Directory** — NFS and iSCSI endpoints only, no SMB
- **No admin password** — SVM CLI access not enabled (manage via file system fsxadmin)

## What to Customize

- Replace placeholders: `name`, `id`, `org`, `env`, and `file_system_id`
- Change `name` to match your naming convention (alphanumeric + underscore only)
- Add `svm_admin_password` if you need SVM-scoped ONTAP CLI access via vsadmin
- Add `active_directory_configuration` if you later need SMB access
- Change `root_volume_security_style` to `MIXED` for dual-protocol access
