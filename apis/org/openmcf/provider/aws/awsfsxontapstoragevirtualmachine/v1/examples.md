# Examples

## 1. Minimal NFS-Only SVM

The simplest SVM for Linux/NFS workloads. UNIX security style, no Active Directory.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapStorageVirtualMachine
metadata:
  name: nfs-dev
  id: awsfxosvm-nfs-dev
  org: engineering
  env: dev
spec:
  region: us-east-1
  file_system_id:
    value: fs-0123456789abcdef0
  name: svm_dev
  root_volume_security_style: UNIX
```

## 2. NFS SVM with Admin Password

NFS-only SVM with vsadmin access enabled for ONTAP CLI operations.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapStorageVirtualMachine
metadata:
  name: nfs-prod
  id: awsfxosvm-nfs-prod
  org: engineering
  env: prod
spec:
  region: us-east-1
  file_system_id:
    value: fs-0123456789abcdef0
  name: svm_prod
  root_volume_security_style: UNIX
  svm_admin_password: VsAdminProd2024!
```

## 3. SMB SVM with Active Directory

Windows-focused SVM with NTFS security style and AD domain join. Enables
SMB file shares with Windows ACLs and identity-based access.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapStorageVirtualMachine
metadata:
  name: smb-prod
  id: awsfxosvm-smb-prod
  org: engineering
  env: prod
spec:
  region: us-east-1
  file_system_id:
    value: fs-0123456789abcdef0
  name: svm_windows
  root_volume_security_style: NTFS
  svm_admin_password: VsAdmin2024!
  active_directory_configuration:
    netbios_name: SVMWIN
    domain_name: corp.example.com
    dns_ips:
      - "10.0.0.1"
      - "10.0.0.2"
    username: svc_fsx_join
    password: ADJoinP@ssw0rd!
    organizational_unit_distinguished_name: "OU=FSx,DC=corp,DC=example,DC=com"
```

## 4. Multiprotocol SVM (NFS + SMB)

Dual-protocol SVM with MIXED security style for environments where both
Linux and Windows clients need access to the same data.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapStorageVirtualMachine
metadata:
  name: multi-shared
  id: awsfxosvm-multi-shared
  org: platform
  env: prod
spec:
  region: us-east-1
  file_system_id:
    value: fs-0123456789abcdef0
  name: svm_shared
  root_volume_security_style: MIXED
  svm_admin_password: SharedAdmin2024!
  active_directory_configuration:
    netbios_name: SVMSHARED
    domain_name: corp.example.com
    dns_ips:
      - "10.0.0.1"
      - "10.0.0.2"
    username: svc_fsx_join
    password: ADJoinP@ssw0rd!
    file_system_administrators_group: FSx Admins
    organizational_unit_distinguished_name: "OU=FSx,DC=corp,DC=example,DC=com"
```

## 5. Cross-Resource Reference (valueFrom)

SVM that references its parent file system via `valueFrom`, enabling
dependency wiring between OpenMCF resources.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapStorageVirtualMachine
metadata:
  name: linked-svm
  id: awsfxosvm-linked-svm
  org: data-team
  env: prod
spec:
  region: us-east-1
  file_system_id:
    valueFrom:
      kind: AwsFsxOntapFileSystem
      name: my-ontap-fs
      fieldPath: status.outputs.file_system_id
  name: svm_linked
  root_volume_security_style: UNIX
```
