# Remove Automatic Service Account Creation from GCP Artifact Registry Repo Module

**Date**: November 4, 2025
**Type**: Refactoring
**Components**: GCP Provider, Pulumi Integration, IAC Execution

## Summary

Simplified the GCP Artifact Registry Repository Pulumi module by removing automatic service account creation and management. The module now focuses solely on creating the repository and managing public access, preparing the codebase for a future enhancement where service account emails will be provided via the API spec rather than created automatically.

## Problem Statement / Motivation

The GCP Artifact Registry module was automatically creating service accounts (reader and writer) and managing their credentials as part of repository creation. This tight coupling made the module more complex and less flexible. The planned enhancement involves accepting service account emails through the API specification, allowing users to provide pre-existing service accounts rather than having the module create them automatically.

### Pain Points

- **Tight Coupling**: Repository creation was bundled with service account lifecycle management
- **Reduced Flexibility**: Users couldn't use existing service accounts; new ones were always created
- **Complexity**: Module handled both infrastructure (repository) and identity (service accounts) concerns
- **Preparation Needed**: Clean separation required before implementing spec-based service account input

## Solution / What's New

Removed all automatic service account creation, credential management, and IAM role assignment logic. The module now creates only the GCP Artifact Registry repository and optionally grants public access via the `allUsers` principal when `EnablePublicAccess` is enabled.

### Key Changes

1. **Deleted**: `service_accounts.go` - entire service account creation module
2. **Simplified**: `main.go` - removed service account creation flow
3. **Streamlined**: `repo.go` - removed IAM role grants for service accounts
4. **Cleaned**: `outputs.go` - removed service account output constants

## Removed Functionality Details

This section documents the removed functionality for future reference when implementing service account email-based permissions.

### Service Account Creation (Removed)

The module previously created two Google Cloud service accounts:

**Reader Service Account**:
- **Naming Pattern**: `{repo-name}-{random-6-chars}-ro`
- **Purpose**: Read-only access to artifact registry
- **Resources Created**:
  - Service account with display name
  - JSON credentials key (TYPE_X509_PEM_FILE)
- **Outputs Exported**:
  - `reader_service_account_email`: Service account email address
  - `reader_service_account_key_base64`: Base64-encoded private key

**Writer Service Account**:
- **Naming Pattern**: `{repo-name}-{random-6-chars}-rw`
- **Purpose**: Read-write-admin access to artifact registry
- **Resources Created**:
  - Service account with display name
  - JSON credentials key (TYPE_X509_PEM_FILE)
- **Outputs Exported**:
  - `writer_service_account_email`: Service account email address
  - `writer_service_account_key_base64`: Base64-encoded private key

**Random Suffix Generation**:
- Used `pulumi-random` provider to generate 6-character alphanumeric suffix
- Configuration: lowercase letters and numbers only, no special characters
- Purpose: Avoid service account ID conflicts (GCP requires unique IDs within project)

### IAM Role Grants (Removed)

The module previously granted the following IAM roles on the created repository:

**When Public Access Disabled** (default):
1. **Reader Role for Reader Service Account**:
   - Resource: `artifactregistry.RepositoryIamMember`
   - Role: `roles/artifactregistry.reader`
   - Member: `serviceAccount:{reader-sa-email}`
   - Purpose: Allow pulling artifacts using reader credentials

2. **Writer Role for Writer Service Account**:
   - Resource: `artifactregistry.RepositoryIamMember`
   - Role: `roles/artifactregistry.writer`
   - Member: `serviceAccount:{writer-sa-email}`
   - Purpose: Allow pushing artifacts using writer credentials

3. **Admin Role for Writer Service Account**:
   - Resource: `artifactregistry.RepositoryIamMember`
   - Role: `roles/artifactregistry.repoAdmin`
   - Member: `serviceAccount:{writer-sa-email}`
   - Purpose: Allow repository management using writer credentials

**When Public Access Enabled**:
- Only `allUsers` reader grant was applied (reader/writer service account grants skipped)

### Code Reference

The removed service account creation logic (from `service_accounts.go`):

```go
func serviceAccounts(ctx *pulumi.Context, locals *Locals, gcpProvider *pulumigcp.Provider) (
    createdReaderServiceAccount, createdWriterServiceAccount *serviceaccount.Account, err error) {
    
    // Generate random suffix for unique service account IDs
    createdServiceAccountSuffixRandomString, err := random.NewRandomString(ctx, "service-account-suffix",
        &random.RandomStringArgs{
            Special: pulumi.Bool(false),
            Lower:   pulumi.Bool(true),
            Upper:   pulumi.Bool(false),
            Number:  pulumi.Bool(true),
            Length:  pulumi.Int(6),
        })
    
    // Create reader service account
    readerServiceAccountName := pulumi.Sprintf("%s-%s-ro", 
        locals.GcpArtifactRegistryRepo.Metadata.Name,
        createdServiceAccountSuffixRandomString.Result)
    
    createdReaderServiceAccount, err = serviceaccount.NewAccount(ctx,
        "reader-service-account",
        &serviceaccount.AccountArgs{
            Project:     pulumi.String(locals.GcpArtifactRegistryRepo.Spec.ProjectId),
            AccountId:   readerServiceAccountName,
            DisplayName: readerServiceAccountName,
        }, pulumi.Provider(gcpProvider))
    
    // Create reader service account key
    createdReaderServiceAccountKey, err := serviceaccount.NewKey(ctx,
        "reader-service-account",
        &serviceaccount.KeyArgs{
            ServiceAccountId: createdReaderServiceAccount.Name,
            PublicKeyType:    pulumi.String("TYPE_X509_PEM_FILE"),
        }, pulumi.Parent(createdReaderServiceAccount))
    
    // Export reader outputs
    ctx.Export(OpReaderServiceAccountEmail, createdReaderServiceAccount.Email)
    ctx.Export(OpReaderServiceAccountKeyBase64, createdReaderServiceAccountKey.PrivateKey)
    
    // Similar logic for writer service account...
}
```

The removed IAM grant logic (from `repo.go`):

```go
// When EnablePublicAccess was false
if !gcpArtifactRegistryRepo.Spec.EnablePublicAccess {
    // Grant reader role to reader service account
    _, err = artifactregistry.NewRepositoryIamMember(ctx,
        fmt.Sprintf("%s-reader", repoName),
        &artifactregistry.RepositoryIamMemberArgs{
            Project:    pulumi.String(gcpArtifactRegistryRepo.Spec.ProjectId),
            Location:   pulumi.String(gcpArtifactRegistryRepo.Spec.Region),
            Repository: createdRepo.RepositoryId,
            Role:       pulumi.String("roles/artifactregistry.reader"),
            Member:     pulumi.Sprintf("serviceAccount:%s", readerServiceAccount.Email),
        }, pulumi.Provider(gcpProvider))
}

// Grant writer role to writer service account
_, err = artifactregistry.NewRepositoryIamMember(ctx, 
    fmt.Sprintf("%s-writer", repoName), 
    &artifactregistry.RepositoryIamMemberArgs{
        Project:    pulumi.String(gcpArtifactRegistryRepo.Spec.ProjectId),
        Location:   pulumi.String(gcpArtifactRegistryRepo.Spec.Region),
        Repository: createdRepo.RepositoryId,
        Role:       pulumi.String("roles/artifactregistry.writer"),
        Member:     pulumi.Sprintf("serviceAccount:%s", writerServiceAccount.Email),
    }, pulumi.Provider(gcpProvider))

// Grant admin role to writer service account
_, err = artifactregistry.NewRepositoryIamMember(ctx, 
    fmt.Sprintf("%s-admin", repoName), 
    &artifactregistry.RepositoryIamMemberArgs{
        Project:    pulumi.String(gcpArtifactRegistryRepo.Spec.ProjectId),
        Location:   pulumi.String(gcpArtifactRegistryRepo.Spec.Region),
        Repository: createdRepo.RepositoryId,
        Role:       pulumi.String("roles/artifactregistry.repoAdmin"),
        Member:     pulumi.Sprintf("serviceAccount:%s", writerServiceAccount.Email),
    }, pulumi.Provider(gcpProvider))
```

## Current State

The module now provides a focused implementation:

**Repository Creation**:
- Creates GCP Artifact Registry repository with specified format (Docker, Maven, npm, etc.)
- Applies GCP labels for resource tracking
- Exports repository metadata: hostname, name, and URL

**Public Access Management**:
- When `Spec.EnablePublicAccess` is `true`:
  - Grants `roles/artifactregistry.reader` to `allUsers` principal
  - Allows unauthenticated pulls from the repository
- When `false`: No IAM grants applied

**Module Outputs**:
```go
const (
    OpRepoHostname = "repo_hostname"  // e.g., "us-central1-docker.pkg.dev"
    OpRepoName     = "repo_name"      // e.g., "my-repo"
    OpRepoUrl      = "repo_url"       // e.g., "us-central1-docker.pkg.dev/project/repo"
)
```

## Implementation Details

### Files Modified

**1. `service_accounts.go`** - DELETED
- Removed entire file (101 lines)
- Contained all service account and credential creation logic

**2. `main.go`** - Simplified
```go
// Before: Created service accounts, then passed them to repo()
addedReaderServiceAccount, addedWriterServiceAccount, err := serviceAccounts(ctx, locals, googleProvider)
if err := repo(ctx, locals, googleProvider, addedReaderServiceAccount, addedWriterServiceAccount); err != nil

// After: Direct repository creation
if err := repo(ctx, locals, googleProvider); err != nil {
    return errors.Wrap(err, "failed to create docker repo")
}
```

**3. `repo.go`** - Streamlined
- Removed `serviceaccount` import
- Removed service account parameters from function signature
- Removed IAM member grants for service accounts
- Kept only repository creation and public access logic
- Updated function documentation

**4. `outputs.go`** - Cleaned
- Removed 4 service account-related output constants
- Retained 3 repository-related output constants

### Dependency Changes

The `BUILD.bazel` file will be auto-regenerated by Gazelle to remove unused dependencies:
- `@com_github_pulumi_pulumi_gcp_sdk_v8//go/gcp/serviceaccount` - No longer needed
- `@com_github_pulumi_pulumi_random_sdk_v4//go/random` - No longer needed

## Benefits

- **Simplified Module**: Focused responsibility (repository creation only)
- **Reduced Complexity**: ~200 lines of code removed
- **Cleaner Architecture**: Separation of infrastructure and identity management
- **Flexible Future**: Prepared for spec-based service account input
- **Fewer Dependencies**: Removed `pulumi-random` dependency
- **Easier Testing**: Simpler module with fewer moving parts

## Future Enhancements

The next phase will implement service account email-based IAM grants:

### Planned Spec Changes

Update `spec.proto` to accept service account emails:

```protobuf
message GcpArtifactRegistryRepoSpec {
    // ... existing fields ...
    
    // Optional: Email of service account to grant reader role
    string reader_service_account_email = X;
    
    // Optional: Email of service account to grant writer role  
    string writer_service_account_email = Y;
}
```

### Planned Implementation

When service account emails are provided:
1. Grant `roles/artifactregistry.reader` to reader service account email (if provided)
2. Grant `roles/artifactregistry.writer` to writer service account email (if provided)
3. Grant `roles/artifactregistry.repoAdmin` to writer service account email (if provided)

**Benefits of Future Approach**:
- Users can use existing service accounts (no creation needed)
- Service account lifecycle managed separately from repository
- More flexible permission model
- Supports organization-wide service accounts
- Better alignment with GCP IAM best practices

### Implementation Reference

Use the removed IAM grant code as reference, but apply to emails from spec:

```go
// Future implementation pattern
if gcpArtifactRegistryRepo.Spec.ReaderServiceAccountEmail != "" {
    _, err = artifactregistry.NewRepositoryIamMember(ctx,
        fmt.Sprintf("%s-reader", repoName),
        &artifactregistry.RepositoryIamMemberArgs{
            Repository: createdRepo.RepositoryId,
            Role:       pulumi.String("roles/artifactregistry.reader"),
            Member:     pulumi.Sprintf("serviceAccount:%s", 
                gcpArtifactRegistryRepo.Spec.ReaderServiceAccountEmail),
        }, pulumi.Provider(gcpProvider))
}

// Similar logic for writer service account email...
```

## Impact

**Developers**:
- Cleaner, more maintainable module code
- Easier to understand module purpose and behavior
- Clear foundation for upcoming service account feature

**Users**:
- Current functionality unchanged for public repositories
- Private repositories now require external IAM configuration (temporary)
- Future enhancement will provide better service account flexibility

**Operations**:
- No automatic service account creation (reduces resource sprawl)
- Service accounts must be managed separately (until future enhancement)

## Related Work

This refactoring prepares for:
- GCP service account management best practices
- Centralized identity management
- Organization-wide service account reuse
- Improved security model with explicit service account permissions

## Testing Notes

**Verify**:
- Repository creation succeeds without service accounts
- Public access grant works when `EnablePublicAccess: true`
- No IAM grants applied when `EnablePublicAccess: false`
- Module outputs contain only repository metadata
- No service account-related outputs

**Test Commands**:
```bash
# Create public artifact registry repository
planton pulumi up \
  --manifest gcp-artifact-registry-repo.yaml \
  --module-dir ${MODULE}

# Verify outputs
planton stack-outputs \
  --manifest gcp-artifact-registry-repo.yaml
```

## Follow-up Fix: GCP Label Validation Compliance

After removing service account creation, deployment testing revealed a systemic GCP label validation issue affecting **12 out of 14** GCP Pulumi modules.

### Issue

GCP labels have strict validation requirements:
- Label values must start with a lowercase letter
- Can only contain lowercase letters, numbers, hyphens, and underscores
- Values like `GcpArtifactRegistryRepo` with uppercase letters cause deployment failures

**Error Example**:
```
Error 400: Invalid repository labels: value "GcpArtifactRegistryRepo" contains invalid character 'G' at index 0 for key "planton-ai_kind"
```

### Root Cause

The `ResourceKind` label value was using the raw protobuf enum string (e.g., `GcpArtifactRegistryRepo`) without converting to lowercase, violating GCP's label value constraints.

### Solution Applied

Fixed all 12 affected GCP modules by:
1. Adding `"strings"` import to each `locals.go` file
2. Wrapping `cloudresourcekind.CloudResourceKind_*.String()` calls with `strings.ToLower()`

**Before**:
```go
gcplabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_GcpArtifactRegistryRepo.String()
```

**After**:
```go
gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpArtifactRegistryRepo.String())
```

### Affected Modules Fixed

All 12 modules updated:
1. `gcpartifactregistryrepo/v1/iac/pulumi/module/locals.go`
2. `gcpvpc/v1/iac/pulumi/module/locals.go`
3. `gcpsubnetwork/v1/iac/pulumi/module/locals.go`
4. `gcprouternat/v1/iac/pulumi/module/locals.go`
5. `gcpgkeworkloadidentitybinding/v1/iac/pulumi/module/locals.go`
6. `gcpgkecluster/v1/iac/pulumi/module/locals.go`
7. `gcpgkeaddonbundle/v1/iac/pulumi/module/locals.go`
8. `gcpsecretsmanager/v1/iac/pulumi/module/locals.go`
9. `gcpgcsbucket/v1/iac/pulumi/module/locals.go`
10. `gcpdnszone/v1/iac/pulumi/module/locals.go`
11. `gcpgkenodepool/v1/iac/pulumi/module/locals.go`
12. `gcpproject/v1/iac/pulumi/module/locals.go`

**Note**: Two modules (`gcpcloudrun` and `gcpcloudsql`) already had the correct implementation.

### Result

Label values now comply with GCP validation:
- Before: `GcpArtifactRegistryRepo` ❌
- After: `gcpartifactregistryrepo` ✅

This fix ensures all GCP resource deployments succeed without label validation errors.

---

**Status**: ✅ Production Ready
**Timeline**: Completed November 4, 2025
**Next Steps**: Implement service account email-based permissions via spec.proto enhancement

