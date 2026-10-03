# Infra Component Commands and APIs

**Date**: November 28, 2025

## Summary

Added CLI commands and Connect-RPC APIs for infra component management in Planton, enabling developers to create and list infra components from the command line. This extends the CLI-backend integration pattern established with catalog kinds, providing a complete interface for managing cloud infrastructure resources through the Planton backend service.

## Problem Statement

The Planton CLI and backend service lacked infra component management capabilities. While catalog kinds could be listed and queried, there was no way to:

- Create infra components from YAML manifests via CLI
- List existing infra components with filtering
- Integrate infra component operations into CLI workflows
- Manage infra components through the backend service

Without these capabilities, users had to manage infra components manually or through other interfaces, creating workflow friction and preventing CLI-based automation.

## Solution

Implemented comprehensive infra component management through two CLI commands and a complete backend service with Connect-RPC integration, following the same patterns established for catalog kind management.

## CLI Commands Added

### 1. `infra-component:create` - Create Infra Component from YAML

**Purpose**: Create a new infra component by providing a YAML manifest file containing the resource specification.

**Usage**:
```bash
planton infra-component:create --arg=path/to/manifest.yaml
```

**Features**:
- Reads YAML manifest from file
- Validates manifest format and required fields (`kind`, `metadata.name`)
- Creates resource in backend database
- Displays created resource details (ID, name, kind, created timestamp)
- Error handling for invalid manifests, duplicate names, and connection issues
- 30-second timeout for backend operations

**Implementation**:
- Reads YAML file content
- Validates manifest structure (kind, metadata.name required)
- Calls backend `CreateInfraComponent` RPC via Connect-RPC
- Handles validation errors, duplicate resource errors, and connection failures
- Displays success message with resource details

**Use cases**:
- Creating infra components from local YAML files
- Scripting infra component creation in CI/CD pipelines
- Bulk resource creation from manifest templates
- Testing infra component creation workflows

**Example**:
```bash
$ planton infra-component:create --arg=my-vpc.yaml
✅ infra component created successfully!

ID: 507f1f77bcf86cd799439011
Name: my-vpc
Kind: CivoVpc
Created At: 2025-11-28 13:14:12
```

### 2. `infra-component:list` - List Infra Components

**Purpose**: List all infra components stored in the backend, with optional filtering by resource kind.

**Usage**:
```bash
# List all infra components
planton infra-component:list

# Filter by kind
planton infra-component:list --kind CivoVpc
planton infra-component:list -k AwsRdsInstance
```

**Features**:
- Lists all infra components in tabular format
- Optional `--kind` / `-k` flag for filtering by resource kind
- Displays ID, name, kind, and created timestamp
- Shows total count with filter information
- Handles empty results gracefully
- Connection error handling with helpful messages

**Output Format**:
```
ID                     NAME      KIND       CREATED
507f1f77bcf86cd799439011  my-vpc   CivoVpc   2025-11-28 13:14:12
507f1f77bcf86cd799439012  my-db    AwsRdsInstance  2025-11-28 13:15:00

Total: 2 infra component(s)
```

**Implementation**:
- Calls backend `ListInfraComponents` RPC via Connect-RPC
- Applies optional kind filter if provided
- Formats results in tabular output using `text/tabwriter`
- Handles empty results and connection errors
- Displays summary with total count

**Use cases**:
- Discovering existing infra components
- Filtering resources by type for specific operations
- Auditing infra component inventory
- Integration with automation scripts

**Example with filtering**:
```bash
$ planton infra-component:list --kind CivoVpc
ID                     NAME      KIND     CREATED
507f1f77bcf86cd799439011  my-vpc   CivoVpc  2025-11-28 13:14:12

Total: 1 infra component(s) (filtered by kind: CivoVpc)
```

## Backend APIs

### InfraComponentService

Service for managing infra components through Connect-RPC, providing create and list operations.

**Service Definition** (`app/backend/apis/proto/infra_component_service.proto`):

```protobuf
service InfraComponentService {
  rpc CreateInfraComponent(CreateInfraComponentRequest) returns (CreateInfraComponentResponse);
  rpc ListInfraComponents(ListInfraComponentsRequest) returns (ListInfraComponentsResponse);
}
```

### RPC Methods

#### 1. `CreateInfraComponent`

**Purpose**: Create a new infra component from a YAML manifest.

**Request**:
```protobuf
message CreateInfraComponentRequest {
  string manifest = 1;  // YAML manifest content
}
```

**Response**:
```protobuf
message CreateInfraComponentResponse {
  InfraComponent resource = 1;
}
```

**Validation**:
- Manifest cannot be empty
- Must contain valid YAML format
- Must include `kind` field
- Must include `metadata.name` field
- Resource name must be unique (no duplicates)

**Error Handling**:
- `CodeInvalidArgument`: Invalid YAML or missing required fields
- `CodeAlreadyExists`: Resource with same name already exists
- `CodeInternal`: Database or service errors

**Implementation** (`app/backend/internal/service/infra_component_service.go`):
- Parses YAML manifest to extract kind and name
- Validates manifest structure
- Checks for duplicate resource names
- Creates resource in database via repository
- Returns created resource with generated ID and timestamps

#### 2. `ListInfraComponents`

**Purpose**: Retrieve all infra components, optionally filtered by kind.

**Request**:
```protobuf
message ListInfraComponentsRequest {
  optional string kind = 1;  // Optional filter by resource kind
}
```

**Response**:
```protobuf
message ListInfraComponentsResponse {
  repeated InfraComponent resources = 1;
}
```

**Features**:
- Returns all resources if no filter provided
- Filters by kind if `kind` field is set
- Returns empty list if no resources match

**Implementation**:
- Applies optional kind filter to database query
- Retrieves all matching resources from database
- Converts domain models to protobuf messages
- Returns list of resources with timestamps

### Data Model

**InfraComponent Message**:
```protobuf
message InfraComponent {
  string id = 1;                    // Unique identifier (MongoDB ObjectID)
  string name = 2;                  // Resource name (from metadata.name)
  string kind = 3;                  // Resource kind (e.g., "CivoVpc", "AwsRdsInstance")
  string manifest = 4;              // Full YAML manifest content
  google.protobuf.Timestamp created_at = 5;
  google.protobuf.Timestamp updated_at = 6;
}
```

**Domain Model** (`app/backend/pkg/models/infra_component.go`):
- Uses MongoDB ObjectID for unique identifiers
- Stores full YAML manifest for complete resource representation
- Tracks creation and update timestamps
- Repository pattern for database operations

## Technical Implementation

### File Structure

**CLI Commands**:
- `cmd/planton/root/infra_component_create.go` - Create command (95 lines)
- `cmd/planton/root/infra_component_list.go` - List command (110 lines)
- `cmd/planton/root.go` - Command registration

**Backend Service**:
- `app/backend/apis/proto/infra_component_service.proto` - Service definition
- `app/backend/internal/service/infra_component_service.go` - Service implementation (157 lines)
- `app/backend/internal/database/infra_component_repo.go` - Repository implementation
- `app/backend/pkg/models/infra_component.go` - Domain model

**Generated Code**:
- `internal/backend/proto/infra_component_service.pb.go` - Protobuf generated code
- `internal/backend/proto/backendv1connect/infra_component_service.connect.go` - Connect-RPC generated code

### Connect-RPC Integration

Following the same pattern as catalog kind commands:

```go
// Create Connect-RPC client
client := backendv1connect.NewInfraComponentServiceClient(
    http.DefaultClient,
    backendURL,
)

// Prepare request
req := &backendv1.CreateInfraComponentRequest{
    Manifest: string(manifestContent),
}

// Execute with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := client.CreateInfraComponent(ctx, connect.NewRequest(req))
```

**Key Features**:
- Uses backend URL from configuration (same as catalog kinds)
- 30-second timeout for all operations
- Consistent error handling across commands
- Connect-RPC binary transport for efficiency

### Database Integration

**Repository Pattern**:
- `InfraComponentRepository` handles all database operations
- MongoDB for persistence
- Methods: `Create`, `List`, `FindByName`
- Supports filtering by kind in list operations

**Database Schema**:
- Collection: `infra_components`
- Fields: `_id`, `name`, `kind`, `manifest`, `created_at`, `updated_at`
- Indexes on `name` for uniqueness and `kind` for filtering

### Error Handling

**CLI Error Messages**:
```bash
# Missing manifest
Error: --arg flag is required. Provide path to YAML manifest file

# Invalid YAML
Error: Invalid manifest - invalid YAML format: yaml: line 2: found character that cannot start any token

# Duplicate resource
Error: Invalid manifest - infra component with name 'my-vpc' already exists

# Connection issues
Error: Cannot connect to backend service at http://localhost:50051. Please check:
  1. The backend service is running
  2. The backend URL is correct
  3. Network connectivity
```

**Backend Error Codes**:
- `CodeInvalidArgument`: Validation failures (empty manifest, missing fields, invalid YAML)
- `CodeAlreadyExists`: Duplicate resource name
- `CodeInternal`: Database or service errors
- `CodeUnavailable`: Backend service not reachable

## Benefits

### For CLI Users

- **Unified Workflow**: Manage infra components directly from CLI without context switching
- **Automation-Friendly**: Scriptable commands enable CI/CD integration
- **Resource Discovery**: List and filter resources to understand current infrastructure state
- **Consistent Experience**: Same configuration system and error handling as catalog kind commands

### For Development Teams

- **Backend Integration**: Establishes pattern for future CLI-backend integrations
- **Code Reuse**: Backend service serves both CLI and web frontend
- **Testing**: CLI provides direct API testing capabilities
- **Documentation**: Commands include help text and clear error messages

### System Architecture

- **Unified Backend**: Single backend service supports multiple client types
- **Connect-RPC Standardization**: Consistent RPC layer across all services
- **Repository Pattern**: Clean separation between service and data layers
- **Error Handling**: Comprehensive error codes and user-friendly messages

## Usage Examples

### Creating an Infra Component

```bash
# Prepare YAML manifest
cat > my-vpc.yaml <<EOF
kind: CivoVpc
metadata:
  name: my-vpc
spec:
  region: NYC1
  cidr: 10.0.0.0/16
EOF

# Create resource
planton infra-component:create --arg=my-vpc.yaml
✅ infra component created successfully!

ID: 507f1f77bcf86cd799439011
Name: my-vpc
Kind: CivoVpc
Created At: 2025-11-28 13:14:12
```

### Listing Infra Components

```bash
# List all resources
$ planton infra-component:list
ID                     NAME      KIND            CREATED
507f1f77bcf86cd799439011  my-vpc   CivoVpc        2025-11-28 13:14:12
507f1f77bcf86cd799439012  my-db    AwsRdsInstance  2025-11-28 13:15:00

Total: 2 infra component(s)

# Filter by kind
$ planton infra-component:list --kind CivoVpc
ID                     NAME      KIND     CREATED
507f1f77bcf86cd799439011  my-vpc   CivoVpc  2025-11-28 13:14:12

Total: 1 infra component(s) (filtered by kind: CivoVpc)
```

### Integration with Configuration

```bash
# Set backend URL (if not already configured)
planton config set backend-url http://localhost:50051

# Use infra component commands
planton infra-component:create --arg=resource.yaml
planton infra-component:list
```

## Impact

### Command Coverage

- **2 CLI commands** for infra component management
- **2 RPC methods** for backend operations
- **Complete CRUD foundation** (Create and Read operations)

### Developer Experience

- **Reduced friction**: Create and list resources with simple commands
- **Better visibility**: List command provides inventory of all resources
- **Automation**: Scriptable commands enable CI/CD integration
- **Consistency**: Same patterns as catalog kind commands

### System Capabilities

- **Backend persistence**: Resources stored in MongoDB with full manifest
- **Resource discovery**: List and filter capabilities for resource inventory
- **Validation**: YAML parsing and field validation before creation
- **Error handling**: Comprehensive error codes and user-friendly messages

## Related Work

### Foundation

This work builds on:
- **CLI Configuration System** - Uses same backend URL configuration
- **Catalog Kind Commands** - Follows same Connect-RPC integration patterns
- **Backend Service Architecture** - Extends existing service patterns
- **Connect-RPC Framework** - Uses same RPC layer as other services

### Complements

This work complements:
- **Catalog Kind Management** - Provides resource management alongside kind management
- **Web Frontend** - Backend service serves both CLI and web clients
- **Planton CLI** - Extends CLI capabilities for infrastructure management

### Future Extensions

This work enables:
- **Update and Delete Operations**: Extend to full CRUD capabilities
- **Resource Status Queries**: Add status tracking and querying
- **Bulk Operations**: Support creating multiple resources from directory
- **Resource Validation**: Pre-creation validation against schemas

## Files Created/Modified

### CLI Commands

- `cmd/planton/root/infra_component_create.go` - Create command implementation
- `cmd/planton/root/infra_component_list.go` - List command implementation
- `cmd/planton/root.go` - Command registration

### Backend Service

- `app/backend/apis/proto/infra_component_service.proto` - Service definition
- `app/backend/internal/service/infra_component_service.go` - Service implementation
- `app/backend/internal/database/infra_component_repo.go` - Repository implementation
- `app/backend/pkg/models/infra_component.go` - Domain model

### Generated Code

- `internal/backend/proto/infra_component_service.pb.go` - Protobuf generated
- `internal/backend/proto/backendv1connect/infra_component_service.connect.go` - Connect-RPC generated

## Next Steps

**Immediate**:
- Add update and delete operations for complete CRUD
- Add resource validation against proto schemas
- Enhance error messages with field-level validation details

**Short-term**:
- Add bulk create from directory of YAML files
- Add resource status tracking and queries
- Add JSON output format option for scripting

**Long-term**:
- Integrate with Planton IaC operations
- Add resource diff and preview capabilities
- Support resource templates and parameterization

---

**Status**: ✅ Complete  
**Component**: CLI Commands and Backend APIs  
**Commands Added**: 2 commands (`infra-component:create`, `infra-component:list`)  
**APIs Added**: 2 RPCs (`CreateInfraComponent`, `ListInfraComponents`)  
**Location**: `cmd/planton/root/` and `app/backend/`

