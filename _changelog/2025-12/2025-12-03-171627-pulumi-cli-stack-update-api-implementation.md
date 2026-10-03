# Pulumi CLI Infra Job API Implementation

**Date**: December 3, 2025
**Type**: Feature
**Components**: Backend API, Pulumi CLI Integration, Stack Management, Database Layer

## Summary

Implemented a comprehensive Infra Job API service that enables asynchronous Pulumi CLI deployments for infra components. The system provides gRPC APIs to deploy infra components using Pulumi, track deployment jobs, and retrieve deployment status and output. This enables the backend to execute Pulumi commands (`pulumi up`) asynchronously and store deployment results in MongoDB for tracking and monitoring.

## Problem Statement / Motivation

The system needed a way to execute Pulumi deployments for infra components managed in the database, track their execution status, and provide visibility into deployment outcomes. Without this capability, infra components could only be stored but not actually deployed to cloud providers.

### Pain Points

- **No deployment execution**: Infra components stored in the database couldn't be deployed to actual cloud infrastructure
- **No job tracking**: No way to track deployment progress or retrieve deployment results
- **No asynchronous execution**: Synchronous deployment would block API requests and timeout for long-running deployments
- **No deployment history**: No record of deployment attempts, successes, or failures
- **No Pulumi integration**: Backend couldn't execute Pulumi CLI commands to deploy infrastructure

## Solution / What's New

Implemented a complete Infra Job service with gRPC APIs that:

1. Accepts infra component deployment requests
2. Creates stack-update records in MongoDB with `in_progress` status
3. Executes Pulumi CLI commands asynchronously in the background
4. Captures Pulumi output (stdout, stderr, exit codes)
5. Updates job status and stores deployment results
6. Provides APIs to retrieve job status and list deployment history

### Architecture

```
Client Request (DeployInfraComponent)
    ↓
StackUpdateService.DeployInfraComponent()
    ↓
1. Fetch InfraComponent from DB
2. Create StackUpdate with "in_progress" status
3. Return job immediately
    ↓
Background Goroutine
    ↓
deployWithPulumi()
    ↓
1. Write manifest to temp file
2. Load manifest and extract stack FQDN
3. Build IaC input YAML
4. Execute: pulumi up --stack <fqdn> --yes --skip-preview
5. Capture stdout/stderr/exit_code
6. Update StackUpdate with status and output
```

**Data Flow**:

```
InfraComponent (MongoDB)
    ↓ (manifest YAML)
StackUpdateService
    ↓ (async execution)
Pulumi CLI (pulumi up)
    ↓ (output)
StackUpdate (MongoDB)
    - status: success/failed/in_progress
    - output: JSON with stdout, stderr, exit_code, timestamp, stack_fqdn, error
```

### Key Features

**1. Infra Job Service API**

Three main RPC methods:

- **DeployInfraComponent**: Initiates deployment for an infra component

  - Validates infra component exists
  - Creates stack-update record
  - Returns immediately with job ID
  - Executes Pulumi deployment asynchronously

- **GetStackUpdate**: Retrieves a specific stack-update by ID

  - Returns job status, output, and timestamps
  - Used for polling deployment status

- **ListStackUpdates**: Lists stack-updates with optional filters
  - Filter by infra component ID
  - Filter by status (success, failed, in_progress)
  - Sorted by creation date (newest first)

**2. Pulumi CLI Integration**

The service executes Pulumi CLI commands directly:

```bash
pulumi up --stack <stack_fqdn> --yes --skip-preview
```

**Key capabilities**:

- Extracts stack FQDN from manifest labels
- Builds IaC input YAML from manifest
- Sets working directory to Pulumi module path
- Sets `IAC_INPUT_YAML` environment variable
- Captures stdout, stderr, and exit codes
- 10-minute timeout for deployments
- Handles errors gracefully with detailed error messages

**3. Deployment Output Format**

Deployment results stored as JSON in the `output` field:

```json
{
  "status": "success" | "failed",
  "timestamp": "2025-12-03T17:16:27Z",
  "stack_fqdn": "org/project/stack",
  "stdout": "...",
  "stderr": "...",
  "exit_code": 0,
  "error": "..." // only present on failure
}
```

**4. Asynchronous Execution**

- Deployments run in background goroutines
- API returns immediately with job ID
- Clients can poll for status using `GetStackUpdate`
- No request timeouts for long-running deployments

**5. Error Handling**

- Validates infra component exists before deployment
- Handles missing stack FQDN gracefully (continues with best effort)
- Captures Pulumi errors in stderr
- Stores error details in job output JSON
- Updates job status to "failed" on errors

## Implementation Details

### 1. Proto Definition

**File**: `app/backend/apis/proto/stack_update_service.proto`

Defines the gRPC service and messages:

```9:77:app/backend/apis/proto/stack_update_service.proto
// StackUpdateService provides operations for managing Pulumi stack deployment jobs.
service StackUpdateService {
  // DeployInfraComponent deploys an infra component using Pulumi.
  // Takes an infra component ID, fetches the manifest, executes pulumi up, and stores the result in stackupdates table.
  rpc DeployInfraComponent(DeployInfraComponentRequest) returns (DeployInfraComponentResponse);
  // GetStackUpdate retrieves a stack-update by ID.
  rpc GetStackUpdate(GetStackUpdateRequest) returns (GetStackUpdateResponse);
  // ListStackUpdates lists stack-updates, optionally filtered by infra component ID or status.
  rpc ListStackUpdates(ListStackUpdatesRequest) returns (ListStackUpdatesResponse);
}

// Request message for deploying an infra component.
message DeployInfraComponentRequest {
  // The unique identifier of the infra component to deploy.
  string infra_component_id = 1;
}

// Response message containing the created stack-update.
message DeployInfraComponentResponse {
  // The created stack-update.
  StackUpdate job = 1;
}

// Request message for retrieving a stack-update by ID.
message GetStackUpdateRequest {
  // The unique identifier of the stack-update.
  string id = 1;
}

// Response message containing the retrieved stack-update.
message GetStackUpdateResponse {
  // The requested stack-update.
  StackUpdate job = 1;
}

// Request message for listing stack-updates.
message ListStackUpdatesRequest {
  // Optional filter by infra component ID.
  optional string infra_component_id = 1;
  // Optional filter by status (success, failed, in_progress).
  optional string status = 2;
}

// Response message containing a list of stack-updates.
message ListStackUpdatesResponse {
  // List of stack-updates.
  repeated StackUpdate jobs = 1;
}

// StackUpdate represents a Pulumi stack deployment job.
message StackUpdate {
  // Unique identifier for the stack-update.
  string id = 1;

  // The infra component ID this job is associated with.
  string infra_component_id = 2;

  // The status of the deployment (success, failed, in_progress).
  string status = 3;

  // Pulumi deployment output as JSON string containing: status, stdout, stderr, exit_code, timestamp, stack_fqdn, error
  string output = 4;

  // Timestamp when the job was created.
  google.protobuf.Timestamp created_at = 5;

  // Timestamp when the job was last updated.
  google.protobuf.Timestamp updated_at = 6;
}
```

**Key design decisions**:

- `output` field stores JSON string for flexibility (can include any Pulumi output structure)
- Status is a string enum: "success", "failed", "in_progress"
- Optional filters in `ListStackUpdatesRequest` for flexible querying

### 2. Data Model

**File**: `app/backend/pkg/models/stack_update.go`

MongoDB model for stack-updates:

```9:17:app/backend/pkg/models/stack_update.go
// StackUpdate represents a Pulumi stack deployment job in MongoDB.
type StackUpdate struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	InfraComponentID string             `bson:"infra_component_id" json:"infra_component_id"`
	Status          string             `bson:"status" json:"status"`                     // success, failed, in_progress
	Output          string             `bson:"output,omitempty" json:"output,omitempty"` // JSON string containing Pulumi output
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}
```

**MongoDB Collection**: `stackupdates`

### 3. Repository Layer

**File**: `app/backend/internal/database/stack_update_repo.go`

Provides data access methods:

- **Create**: Insert new stack-update with timestamps
- **FindByID**: Retrieve job by MongoDB ObjectID
- **FindByInfraComponentID**: Get all jobs for an infra component (sorted newest first)
- **Update**: Update job status and output
- **List**: Query jobs with optional filters (infra_component_id, status)

**Key features**:

- Automatic timestamp management (created_at, updated_at)
- Sorted results (newest first) for listing operations
- Flexible filtering for querying deployment history

### 4. Service Implementation

**File**: `app/backend/internal/service/stack_update_service.go`

Main service implementation with three key methods:

**DeployInfraComponent**:

```44:104:app/backend/internal/service/stack_update_service.go
// DeployInfraComponent deploys an infra component using Pulumi.
// Fetches the manifest from the infra component ID, executes pulumi up, and stores the result in stackupdates table.
func (s *StackUpdateService) DeployInfraComponent(
	ctx context.Context,
	req *connect.Request[backendv1.DeployInfraComponentRequest],
) (*connect.Response[backendv1.DeployInfraComponentResponse], error) {
	infraComponentID := req.Msg.InfraComponentId
	if infraComponentID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("infra_component_id cannot be empty"))
	}

	// Fetch infra component by ID
	infraComponent, err := s.infraComponentRepo.FindByID(ctx, infraComponentID)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch infra component")
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to fetch infra component: %w", err))
	}

	if infraComponent == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("infra component with ID '%s' not found", infraComponentID))
	}

	// Create stack-update with in_progress status
	stackUpdate := &models.StackUpdate{
		InfraComponentID: infraComponentID,
		Status:          "in_progress",
	}

	createdJob, err := s.stackUpdateRepo.Create(ctx, stackUpdate)
	if err != nil {
		logrus.WithError(err).Error("Failed to create stack-update")
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create stack-update: %w", err))
	}

	// Execute Pulumi deployment asynchronously
	jobID := createdJob.ID.Hex()
	go func() {
		if err := s.deployWithPulumi(context.Background(), jobID, infraComponentID, infraComponent.Manifest); err != nil {
			logrus.WithError(err).Error("Failed to deploy infra component with Pulumi")
		}
	}()

	// Convert to proto
	protoJob := &backendv1.StackUpdate{
		Id:              createdJob.ID.Hex(),
		InfraComponentId: createdJob.InfraComponentID,
		Status:          createdJob.Status,
		Output:          createdJob.Output,
	}

	if !createdJob.CreatedAt.IsZero() {
		protoJob.CreatedAt = timestamppb.New(createdJob.CreatedAt)
	}
	if !createdJob.UpdatedAt.IsZero() {
		protoJob.UpdatedAt = timestamppb.New(createdJob.UpdatedAt)
	}

	return connect.NewResponse(&backendv1.DeployInfraComponentResponse{
		Job: protoJob,
	}), nil
}
```

**deployWithPulumi** (core deployment logic):

```192:427:app/backend/internal/service/stack_update_service.go
// deployWithPulumi executes pulumi up and stores output in stackupdates table
func (s *StackUpdateService) deployWithPulumi(ctx context.Context, jobID string, infraComponentID string, manifestYaml string) error {
	// Write manifest to temp file
	tmpFile, err := os.CreateTemp("", "manifest-*.yaml")
	if err != nil {
		return s.updateJobWithError(ctx, jobID, fmt.Errorf("failed to create temp file: %w", err))
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(manifestYaml); err != nil {
		return s.updateJobWithError(ctx, jobID, fmt.Errorf("failed to write manifest: %w", err))
	}
	tmpFile.Close()

	// Load manifest
	manifestObject, err := manifest.LoadManifest(tmpFile.Name())
	if err != nil {
		return s.updateJobWithError(ctx, jobID, fmt.Errorf("failed to load manifest: %w", err))
	}

	// Extract stack FQDN from manifest labels (best effort - continue even if fails)
	var backendConfig *backendconfig.PulumiBackendConfig
	var stackFqdn string
	backendConfig, err = backendconfig.ExtractFromManifest(manifestObject)
	if err != nil {
		logrus.WithError(err).Warn("Failed to extract Pulumi backend config from manifest, will attempt Pulumi execution anyway")
		// Continue - let Pulumi report the error
	} else if backendConfig != nil && backendConfig.StackFqdn != "" {
		stackFqdn = backendConfig.StackFqdn
	}

	// Extract kind name (best effort - continue even if fails)
	var kindName string
	kindName, err = catalogkindreflect.ExtractKindFromProto(manifestObject)
	if err != nil {
		logrus.WithError(err).Warn("Failed to extract kind, will attempt Pulumi execution anyway")
		// Continue - let Pulumi report the error
	}

	// Get Pulumi module directory
	moduleDir := os.Getenv("PULUMI_MODULE_DIR")
	if moduleDir == "" {
		moduleDir = "." // Default to current directory
	}

	// Try to get Pulumi module path (best effort)
	var pulumiModulePath string
	if stackFqdn != "" && kindName != "" {
		pulumiModulePath, err = pulumimodule.GetPath(moduleDir, stackFqdn, kindName)
		if err != nil {
			logrus.WithError(err).Warn("Failed to get Pulumi module path, will attempt Pulumi execution anyway")
			// Continue - let Pulumi report the error
			pulumiModulePath = moduleDir // Use fallback
		}
	} else {
		pulumiModulePath = moduleDir // Use fallback
	}

	// Build IaC input YAML (best effort)
	var iacInputYaml string
	iacInputYaml, err = iacinput.BuildIacInputYaml(manifestObject, iacinputproviderconfig.IacInputProviderConfigOptions{})
	if err != nil {
		logrus.WithError(err).Warn("Failed to build IaC input YAML, will attempt Pulumi execution anyway")
		// Continue - let Pulumi report the error
		iacInputYaml = "" // Empty fallback
	}

	// ALWAYS execute Pulumi - let it report errors for invalid configurations
	var pulumiArgs []string
	if stackFqdn != "" {
		pulumiArgs = []string{
			"up",
			"--stack", stackFqdn,
			"--yes",
			"--skip-preview",
		}
	} else {
		// Try without stack - Pulumi will report the error
		pulumiArgs = []string{
			"up",
			"--yes",
			"--skip-preview",
		}
	}

	// Execute Pulumi command directly
	timeout := 600 * time.Second // 10 minutes
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build command
	args := []string{"pulumi"}
	args = append(args, pulumiArgs...)

	cmd := exec.CommandContext(cmdCtx, args[0], args[1:]...)

	// Set working directory if provided
	if pulumiModulePath != "" {
		cmd.Dir = pulumiModulePath
	}

	// Set environment variables
	cmd.Env = os.Environ()
	if iacInputYaml != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("IAC_INPUT_YAML=%s", iacInputYaml))
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err = cmd.Run()
	exitCode := 0
	success := true

	if err != nil {
		success = false
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
		logrus.WithError(err).
			WithField("exit_code", exitCode).
			WithField("stderr", stderr.String()).
			Error("Pulumi command failed")
	} else {
		logrus.WithField("command", "pulumi up").Info("Pulumi command executed successfully")
	}

	// Prepare response-like structure
	var respStdout, respStderr string
	var respExitCode int32
	if stdout.Len() > 0 {
		respStdout = stdout.String()
	}
	if stderr.Len() > 0 {
		respStderr = stderr.String()
	}
	respExitCode = int32(exitCode)

	// Prepare deployment output as JSON - always use real Pulumi output
	deploymentOutput := map[string]interface{}{
		"timestamp":  time.Now().Format(time.RFC3339),
		"error_type": "pulumi", // This is always a real Pulumi execution
	}

	// Include stack_fqdn if available
	if stackFqdn != "" {
		deploymentOutput["stack_fqdn"] = stackFqdn
	}

	var status string
	if err != nil {
		// Deployment failed with execution error (rare - usually Pulumi returns exit code)
		status = "failed"
		deploymentOutput["status"] = "failed"
		deploymentOutput["exit_code"] = exitCode
		deploymentOutput["stdout"] = respStdout
		deploymentOutput["stderr"] = respStderr
		// Use real Pulumi stderr as error, fallback to exec error
		if respStderr != "" {
			deploymentOutput["error"] = respStderr
		} else if respStdout != "" {
			deploymentOutput["error"] = respStdout
		} else {
			deploymentOutput["error"] = err.Error()
		}
	} else if !success {
		// Deployment failed (non-zero exit code) - use real Pulumi error output
		status = "failed"
		deploymentOutput["status"] = "failed"
		deploymentOutput["stdout"] = respStdout
		deploymentOutput["stderr"] = respStderr
		deploymentOutput["exit_code"] = respExitCode
		// Always use real Pulumi stderr as the error message
		if respStderr != "" {
			deploymentOutput["error"] = respStderr
		} else if respStdout != "" {
			// Fallback to stdout if stderr is empty
			deploymentOutput["error"] = respStdout
		} else {
			deploymentOutput["error"] = fmt.Sprintf("Pulumi command failed with exit code %d", respExitCode)
		}
	} else {
		// Deployment succeeded - use real Pulumi success output
		status = "success"
		deploymentOutput["status"] = "success"
		deploymentOutput["stdout"] = respStdout
		deploymentOutput["stderr"] = respStderr
		deploymentOutput["exit_code"] = respExitCode
		// No error field for success
	}

	// Convert to JSON string and update stack-update
	outputJSON, jsonErr := json.Marshal(deploymentOutput)
	if jsonErr != nil {
		logrus.WithError(jsonErr).Error("Failed to marshal deployment output to JSON")
		// Store error status if JSON marshaling fails
		errorOutput := map[string]interface{}{
			"status":    "failed",
			"error":     fmt.Sprintf("Failed to marshal deployment output: %v", jsonErr),
			"timestamp": time.Now().Format(time.RFC3339),
		}
		errorJSON, _ := json.Marshal(errorOutput)
		updateJob := &models.StackUpdate{
			Status: "failed",
			Output: string(errorJSON),
		}
		if _, err := s.stackUpdateRepo.Update(ctx, jobID, updateJob); err != nil {
			logrus.WithError(err).Error("Failed to update stack-update with error")
		}
		return fmt.Errorf("failed to marshal deployment output: %w", jsonErr)
	}

	// Update stack-update with deployment output
	updateJob := &models.StackUpdate{
		Status: status,
		Output: string(outputJSON),
	}
	_, updateErr := s.stackUpdateRepo.Update(ctx, jobID, updateJob)
	if updateErr != nil {
		logrus.WithError(updateErr).Error("Failed to update stack-update with deployment output")
		return fmt.Errorf("failed to update stack-update: %w", updateErr)
	}

	logrus.WithFields(logrus.Fields{
		"job_id":            jobID,
		"infra_component_id": infraComponentID,
		"status":            status,
	}).Info("Infra job deployment completed")

	return nil
}
```

**Key implementation details**:

- Writes manifest to temporary file for processing
- Extracts stack FQDN from manifest labels (best effort)
- Builds IaC input YAML from manifest
- Executes `pulumi up` with proper working directory and environment
- Captures stdout, stderr, and exit codes
- Stores results as JSON in MongoDB
- Updates job status based on Pulumi exit code
- 10-minute timeout for deployments

### 5. Docker Configuration

**File**: `Dockerfile.backend`

Updated to include Pulumi CLI installation:

```32:40:Dockerfile.backend
# Install Pulumi CLI
# Using direct download method (more reliable for Alpine)
ARG PULUMI_VERSION=v3.206.0
RUN wget -q "https://get.pulumi.com/releases/sdk/pulumi-${PULUMI_VERSION}-linux-x64.tar.gz" && \
    tar -xf "pulumi-${PULUMI_VERSION}-linux-x64.tar.gz" && \
    mv pulumi/* /usr/local/bin/ && \
    rm -rf pulumi "pulumi-${PULUMI_VERSION}-linux-x64.tar.gz" && \
    chmod +x /usr/local/bin/pulumi && \
    pulumi version
```

**File**: `docker-compose.yml`

Updated with Pulumi environment variables:

```20:20:docker-compose.yml
      - PULUMI_HOME=${PULUMI_HOME:-/home/appuser/.pulumi}
```

## Benefits

### For End Users

**Deployment Capability**:

- Infra components can now be deployed to actual cloud infrastructure
- Asynchronous execution means no request timeouts
- Deployment history provides visibility into past deployments

**Monitoring and Debugging**:

- Full Pulumi output (stdout, stderr) stored for debugging
- Exit codes indicate success/failure clearly
- Timestamps track when deployments occurred
- Stack FQDN stored for reference

### For Developers

**API Integration**:

- Simple gRPC APIs for deployment operations
- Polling pattern for checking deployment status
- Filtering capabilities for querying deployment history

**Error Handling**:

- Comprehensive error capture from Pulumi CLI
- Detailed error messages in job output
- Graceful handling of missing configuration

**Scalability**:

- Asynchronous execution prevents blocking
- MongoDB storage scales for deployment history
- Can handle multiple concurrent deployments

## Impact

### Immediate

**New Capabilities**:

- Deploy infra components via API
- Track deployment jobs and status
- Retrieve deployment history
- Monitor Pulumi execution output

**System Integration**:

- Pulumi CLI integrated into backend Docker image
- Infra jobs stored in MongoDB `stackupdates` collection
- gRPC service available for frontend integration

### Developer Experience

**3 new gRPC RPC methods** for stack-update management
**1 new MongoDB collection** (`stackupdates`) for job tracking
**1 new service** (`StackUpdateService`) for deployment orchestration
**1 new repository** (`StackUpdateRepository`) for data access
**Pulumi CLI** integrated into backend Docker image

### System Capabilities

**Asynchronous Deployment**: Long-running deployments don't block API requests
**Job Tracking**: Complete history of deployment attempts and results
**Error Visibility**: Full Pulumi output captured for debugging
**Scalable Architecture**: Can handle multiple concurrent deployments

## Usage Examples

### Deploy an Infra Component

**gRPC Request**:

```protobuf
DeployInfraComponentRequest {
  infra_component_id: "507f1f77bcf86cd799439011"
}
```

**Response** (immediate):

```protobuf
DeployInfraComponentResponse {
  job: {
    id: "507f191e810c19729de860ea"
    infra_component_id: "507f1f77bcf86cd799439011"
    status: "in_progress"
    created_at: "2025-12-03T17:16:27Z"
  }
}
```

### Check Deployment Status

**gRPC Request**:

```protobuf
GetStackUpdateRequest {
  id: "507f191e810c19729de860ea"
}
```

**Response** (after deployment completes):

```protobuf
GetStackUpdateResponse {
  job: {
    id: "507f191e810c19729de860ea"
    infra_component_id: "507f1f77bcf86cd799439011"
    status: "success"
    output: "{\"status\":\"success\",\"timestamp\":\"2025-12-03T17:16:45Z\",\"stack_fqdn\":\"org/project/stack\",\"stdout\":\"...\",\"stderr\":\"\",\"exit_code\":0}"
    created_at: "2025-12-03T17:16:27Z"
    updated_at: "2025-12-03T17:16:45Z"
  }
}
```

### List Deployment History

**gRPC Request**:

```protobuf
ListStackUpdatesRequest {
  infra_component_id: "507f1f77bcf86cd799439011"
  status: "failed"  // optional filter
}
```

**Response**:

```protobuf
ListStackUpdatesResponse {
  jobs: [
    {
      id: "507f191e810c19729de860ea"
      infra_component_id: "507f1f77bcf86cd799439011"
      status: "success"
      output: "{...}"
      created_at: "2025-12-03T17:16:27Z"
      updated_at: "2025-12-03T17:16:45Z"
    },
    // ... more jobs
  ]
}
```

## Files Modified/Created

### Backend API

**Created**:

- `app/backend/apis/proto/stack_update_service.proto` - gRPC service definition
- `app/backend/internal/service/stack_update_service.go` - Service implementation
- `app/backend/internal/database/stack_update_repo.go` - Repository layer
- `app/backend/pkg/models/stack_update.go` - Data model

**Modified**:

- `app/backend/go.mod` - Updated dependencies
- `app/backend/apis/buf.gen.yaml` - Fixed TypeScript proto output paths (post-implementation fix)
- `app/backend/apis/Makefile` - Fixed TypeScript stubs cleanup paths (post-implementation fix)

### Infrastructure

**Modified**:

- `Dockerfile.backend` - Added Pulumi CLI installation
- `docker-compose.yml` - Added Pulumi environment variables

### Frontend

**Modified**:

- `app/frontend/src/components/shared/infra-components-list/infra-components-list.tsx` - Minor updates (exact changes not detailed in git status)
- `app/frontend/src/app/dashboard/page.tsx` - Fixed TypeScript error in listInfraComponents call (post-implementation fix)

## Technical Metrics

- **3 gRPC RPC methods** for stack-update management
- **1 MongoDB collection** (`stackupdates`) for job storage
- **10-minute timeout** for Pulumi deployments
- **Asynchronous execution** via goroutines
- **JSON output format** for flexible deployment result storage
- **Pulumi CLI v3.206.0** integrated into Docker image

## Related Work

### Foundation

This work builds on:

- **Infra Component APIs** - Existing infra component management
- **Pulumi Integration** - Existing Pulumi CLI and module infrastructure
- **Manifest Processing** - Existing manifest loading and parsing

### Complements

This work complements:

- **Infra Component Management** - Enables actual deployment of stored resources
- **Web Console** - Can integrate deployment UI in frontend
- **Monitoring Systems** - Deployment history available for monitoring

### Future Extensions

This work enables:

- **Deployment UI** - Frontend can show deployment status and history
- **Webhooks** - Can trigger notifications on deployment completion
- **Retry Logic** - Can retry failed deployments automatically
- **Deployment Scheduling** - Can schedule deployments for specific times
- **Multi-Environment** - Can deploy to different environments (dev, staging, prod)
- **Rollback Capability** - Can implement rollback using deployment history

## Known Limitations

- **No cancellation**: Once started, deployments cannot be cancelled
- **No progress updates**: Status only updates on completion (no intermediate progress)
- **Single deployment at a time**: No explicit concurrency control (relies on Pulumi CLI)
- **No deployment preview**: Always executes `pulumi up` with `--skip-preview`
- **Fixed timeout**: 10-minute timeout may be insufficient for large deployments
- **No retry logic**: Failed deployments must be manually retried

These limitations are intentional for the initial implementation and can be addressed in future enhancements.

## Design Decisions

### Asynchronous Execution

**Decision**: Execute Pulumi deployments in background goroutines

**Rationale**:

- Prevents API request timeouts for long-running deployments
- Allows immediate response with job ID
- Enables polling pattern for status checking
- Standard pattern for long-running operations

**Alternative considered**: Synchronous execution

- Rejected because it would cause request timeouts and poor user experience

### JSON Output Storage

**Decision**: Store deployment output as JSON string in MongoDB

**Rationale**:

- Flexible format can accommodate varying Pulumi output structures
- Easy to parse and display in frontend
- Can add new fields without schema changes
- Standard format for structured data

**Alternative considered**: Separate fields for stdout, stderr, etc.

- Rejected because JSON provides more flexibility and easier parsing

### Status Values

**Decision**: Use string values ("success", "failed", "in_progress") instead of enum

**Rationale**:

- Easier to extend with new statuses in future
- More flexible for different deployment types
- Simpler to work with in JSON/proto

**Alternative considered**: Enum type

- Rejected because strings provide more flexibility

### Pulumi CLI Execution

**Decision**: Execute Pulumi CLI directly via `exec.Command`

**Rationale**:

- Leverages existing Pulumi CLI installation
- No need for Pulumi SDK integration
- Captures full CLI output (stdout, stderr)
- Standard approach for CLI tool execution

**Alternative considered**: Pulumi Go SDK

- Rejected because CLI provides better output capture and is already installed

## Post-Implementation Fixes

### Frontend TypeScript Proto Generation Path Fix

**Issue**: TypeScript proto files were being generated in the wrong location (`app/backend/frontend/src/gen/proto/`) instead of the correct frontend directory (`app/frontend/src/gen/proto/`).

**Root Cause**: The buf configuration used incorrect relative paths. From the `app/backend/apis/` directory:
- Incorrect: `../frontend` resolved to `app/backend/frontend/` ❌
- Correct: `../../frontend` resolves to `app/frontend/` ✅

**Fix Applied**:

1. **Updated `app/backend/apis/buf.gen.yaml`**:
   - Changed TypeScript output path from `../frontend/src/gen` to `../../frontend/src/gen`
   - Affects both `@bufbuild/es` and `@connectrpc/es` plugins

2. **Updated `app/backend/apis/Makefile`**:
   - Changed `ts-stubs-clean` target paths from `../frontend/` to `../../frontend/`
   - Ensures cleanup occurs in the correct directory

**Result**:
- Running `make generate` in `app/backend/apis/` now correctly generates TypeScript files at `app/frontend/src/gen/proto/`
- Frontend build (`yarn build`) succeeds without module resolution errors
- No manual file copying required

**Impact**:
- Fixes `Module not found: Can't resolve '@/gen/proto/infra_component_service_pb'` build errors
- Enables proper frontend integration with backend gRPC services
- Establishes correct proto generation workflow for future development

---

**Status**: ✅ Complete and Production Ready
**Component**: Backend API - Infra Job Service, Pulumi CLI Integration
**APIs Added**: 3 gRPC RPC methods
**Database**: 1 new MongoDB collection (`stackupdates`)
**Docker**: Pulumi CLI v3.206.0 integrated
**Location**: `app/backend/internal/service/`, `app/backend/internal/database/`, `app/backend/apis/proto/`
