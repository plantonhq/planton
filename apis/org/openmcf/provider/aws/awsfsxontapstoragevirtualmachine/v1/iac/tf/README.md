# Terraform Module: AwsFsxOntapStorageVirtualMachine

Terraform HCL module that provisions an Amazon FSx for NetApp ONTAP Storage Virtual Machine (SVM) with optional Active Directory integration.

## Module Structure

```
tf/
├── main.tf       — SVM resource with dynamic AD configuration block
├── variables.tf  — Input variables (provider config, spec fields)
├── outputs.tf    — 12 output values matching stack_outputs.proto
└── provider.tf   — AWS provider configuration
```

## Resources Created

| Resource | Type | Description |
|----------|------|-------------|
| `this` | `aws_fsx_ontap_storage_virtual_machine` | The ONTAP SVM |

## Usage

```bash
terraform init
terraform plan -var="region=us-east-1" -var="file_system_id=fs-abc123" -var="svm_name=svm_test"
terraform apply
```

## Active Directory

Pass the `active_directory_configuration` variable as an object to enable SMB:

```hcl
active_directory_configuration = {
  netbios_name = "MYSERVER"
  domain_name  = "corp.example.com"
  dns_ips      = ["10.0.0.1", "10.0.0.2"]
  username     = "admin"
  password     = "secret"
}
```

Set to `null` (default) to create an NFS/iSCSI-only SVM.
