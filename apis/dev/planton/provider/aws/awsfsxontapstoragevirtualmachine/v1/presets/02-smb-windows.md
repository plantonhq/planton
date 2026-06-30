# SMB Windows SVM with Active Directory

Windows-focused Storage Virtual Machine with NTFS security style and Active Directory domain join. Enables SMB/CIFS file share access with Windows ACLs and identity-based access control.

## When to Use

- Windows workloads requiring SMB file shares (home directories, department shares)
- .NET applications needing Windows-native file access
- SQL Server databases using SMB for shared storage
- Enterprise content management with Windows ACLs
- Any workload requiring Active Directory integration for file access

## What It Configures

- **NTFS security style** — Windows ACLs for all volumes (identity-based permissions)
- **Active Directory** — Self-managed AD with domain join, NetBIOS name, and OU placement
- **SVM admin password** — Enables vsadmin SSH access for SVM-scoped ONTAP CLI operations
- **SMB endpoint** — Automatically created when AD is configured

## What to Customize

- Replace all `<REPLACE>` placeholders with actual values
- **Critical**: Replace password placeholders with real credentials. For production, inject via CI/CD secrets
- Adjust `netbios_name` (1-15 characters, must be unique in the AD domain)
- Adjust `dns_ips` to point to your actual AD DNS servers
- Adjust `organizational_unit_distinguished_name` to your AD OU structure
- Remove `file_system_administrators_group` to use default "Domain Admins"
- Remove `svm_admin_password` if SVM CLI access is not needed
