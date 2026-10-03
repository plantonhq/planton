# Infra Component Apply Command with Automatic Deployment

**Date**: December 11, 2025
**Type**: Feature Enhancement
**Components**: CLI, Backend API, Documentation

## Summary

Enhanced the `infra-component:apply` command to automatically trigger Pulumi deployments when creating or updating infra components. The command now uses the `ApplyInfraComponent` API for a simpler, more reliable upsert operation, and provides clear feedback about deployment status to users.

## Problem Statement

The `infra-component:apply` command was missing automatic deployment triggering. When users applied an infra component, the resource was created or updated in the database, but Pulumi deployment was not automatically triggered. This required users to manually call the deployment API separately, creating a poor user experience.

Additionally, the command implementation was using manual create/update logic instead of leveraging the existing `ApplyInfraComponent` API, which already handled the upsert operation correctly.

## Solution

### 1. Added Automatic Deployment Trigger to ApplyInfraComponent API

**File**: `app/backend/internal/service/infra_component_service.go`

Added automatic Pulumi deployment triggering to the `ApplyInfraComponent` method, matching the behavior of `CreateInfraComponent` and `UpdateInfraComponent`:

```go
// Trigger Pulumi deployment automatically (credentials will be resolved from database)
if s.stackUpdateService != nil {
    // Create a deployment request (no provider_config needed - will be resolved automatically)
    deployReq := &connect.Request[backendv1.DeployInfraComponentRequest]{
        Msg: &backendv1.DeployInfraComponentRequest{
            InfraComponentId: resultResource.ID.Hex(),
        },
    }

    // Trigger deployment asynchronously (don't wait for it)
    go func() {
        _, _ = s.stackUpdateService.DeployInfraComponent(context.Background(), deployReq)
    }()
}
```

**Impact**: Now all three operations (Create, Update, Apply) consistently trigger deployments automatically.

### 2. Refactored CLI Command to Use ApplyInfraComponent API

**File**: `cmd/planton/root/infra_component_apply.go`

**Before**: Manual implementation that:

- Listed all resources to find existing ones
- Called `CreateInfraComponent` or `UpdateInfraComponent` separately
- Required complex logic to determine create vs update

**After**: Simplified implementation that:

- Uses `ApplyInfraComponent` API directly (single call handles both create and update)
- Validates YAML manifest before making API call
- Provides better user feedback with progress messages
- Shows deployment status information

**Key Changes**:

1. **YAML Validation**: Added validation to check for required fields (`kind`, `metadata.name`) before API call
2. **Better User Feedback**:
   - Shows "Applying infra component: kind=X, name=Y" message
   - Displays "Created" or "Updated" action clearly
   - Shows deployment status message at the end
3. **Simplified Logic**: Removed ~50 lines of manual create/update detection code

**Example Output**:

```
Applying infra component: kind=GcpCloudSql, name=gcp-postgres-example
Checking if resource exists...
✅ infra component created successfully!

Action: Created
ID: 507f1f77bcf86cd799439011
Name: gcp-postgres-example
Kind: GcpCloudSql
Created At: 2025-12-11 08:50:26
Updated At: 2025-12-11 08:50:26

🚀 Pulumi deployment has been triggered automatically.
   Deployment is running in the background.
   Use 'planton stack-update:list' to check deployment status.
```

### 3. Updated CLI Documentation

**File**: `cmd/planton/CLI-HELP.md`

Updated the `infra-component:apply` documentation to:

- Reflect the actual command output format
- Document the automatic deployment behavior
- Update sample outputs to match real command output
- Explain how the `ApplyInfraComponent` API works internally

## Technical Details

### API Flow

1. **CLI Command** → Reads YAML manifest, validates it
2. **ApplyInfraComponent API** → Checks if resource exists (by `name` + `kind`)
   - If exists: Updates resource
   - If not exists: Creates resource
3. **Automatic Deployment** → Triggers `DeployInfraComponent` asynchronously
   - Credentials resolved from database based on provider
   - Infra job created with "in_progress" status
   - Pulumi deployment runs in background
4. **Response** → Returns resource with `created` flag

### Benefits

- **Consistency**: All three operations (Create, Update, Apply) now trigger deployments
- **Simplicity**: Single API call instead of manual create/update logic
- **User Experience**: Clear feedback about what's happening and deployment status
- **Reliability**: Uses the same proven upsert logic as the API

## Files Changed

1. `app/backend/internal/service/infra_component_service.go` (+15 lines)

   - Added deployment trigger to `ApplyInfraComponent` method

2. `cmd/planton/root/infra_component_apply.go` (+66 lines, -22 lines)

   - Refactored to use `ApplyInfraComponent` API
   - Added YAML validation
   - Improved user feedback

3. `cmd/planton/CLI-HELP.md` (+50 lines modified)
   - Updated documentation to match actual behavior
   - Added deployment status information

## Testing

The changes were tested with:

- Creating new GCP Cloud SQL resources
- Updating existing resources (storage size changes)
- Verifying automatic deployment triggers
- Confirming deployment status messages appear correctly

## Related Work

This enhancement builds on the database-driven credential management system implemented earlier, which enables automatic credential resolution during deployments. The `ApplyInfraComponent` API was already implemented but was missing the deployment trigger that existed in `CreateInfraComponent` and `UpdateInfraComponent`.

## Migration Notes

No migration required. This is a backward-compatible enhancement that adds functionality without breaking existing behavior.
