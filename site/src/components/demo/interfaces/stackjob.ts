/**
 * TypeScript interfaces for StackJob and related types
 * Extracted from planton-cloud protobuf definitions
 * Refactored to use common interfaces
 */

import {
  Timestamp,
  ApiResourceMetadata,
  ApiResourceLifecycle,
  ApiResourceAudit,
  ApiResourceKindAndNameAndId,
} from './common';

// Enums
export enum StackJobOperationType {
  STACK_JOB_OPERATION_TYPE_UNSPECIFIED = 0,
  INIT = 1,
  REFRESH = 2,
  UPDATE_PREVIEW = 3,
  UPDATE = 4,
  DESTROY_PREVIEW = 5,
  DESTROY = 6,
}

export enum StackJobProgressEventType {
  STACK_JOB_PROGRESS_EVENT_TYPE_UNSPECIFIED = 0,
  STACK_JOB_STATUS_CHANGED_EVENT = 1,
  IAC_MODULE_REPO_SETUP_EVENT = 2,
  IAC_OPERATION_STATUS_EVENT = 3,
  IAC_STACK_RESOURCES_EVENT = 4,
  IAC_STACK_OUTPUTS_EVENT = 5,
  TERRAFORM_ENGINE_EVENT = 6,
  PULUMI_ENGINE_EVENT = 7,
}

export enum IacResourceProgressStatusLabel {
  IAC_RESOURCE_PROGRESS_STATUS_LABEL_UNSPECIFIED = 0,
  REFRESHING = 1,
  UNCHANGED = 2,
  CREATE = 3,
  IAC_RESOURCE_PROGRESS_STATUS_LABEL_UPDATE = 4,
  DELETE = 5,
  REPLACE = 6,
  CREATING = 7,
  UPDATING = 8,
  DELETING = 9,
  REPLACING = 10,
  REFRESHED = 11,
  IAC_RESOURCE_PROGRESS_STATUS_LABEL_CREATED = 12,
  IAC_RESOURCE_PROGRESS_STATUS_LABEL_UPDATED = 13,
  IAC_RESOURCE_PROGRESS_STATUS_LABEL_DELETED = 14,
  REPLACED = 15,
  REFRESH_FAILED = 16,
  CREATE_FAILED = 17,
  UPDATE_FAILED = 18,
  DELETE_FAILED = 19,
  REPLACE_FAILED = 20,
}

export enum IacDiagnosticEventSeverityType {
  IAC_DIAGNOSTIC_EVENT_SEVERITY_TYPE_UNSPECIFIED = 0,
  INFO = 1,
  WARNING = 2,
  ERROR = 3,
}

// Workflow enums (from commons.workflow)
export enum WorkflowExecutionStatus {
  WORKFLOW_EXECUTION_STATUS_UNSPECIFIED = 0,
  RUNNING = 1,
  COMPLETED = 2,
  FAILED = 3,
  CANCELLED = 4,
  TERMINATED = 5,
  CONTINUED_AS_NEW = 6,
  TIMED_OUT = 7,
}

export enum WorkflowExecutionResult {
  WORKFLOW_EXECUTION_RESULT_UNSPECIFIED = 0,
  SUCCEEDED = 1,
  FAILED = 2,
  CANCELLED = 3,
  TERMINATED = 4,
  CONTINUED_AS_NEW = 5,
  TIMED_OUT = 6,
}

// StackJob specific interfaces
export interface IacProvisioner {
  type: string;
}

export interface IacOperationDiagnosticMessage {
  severity: IacDiagnosticEventSeverityType;
  resourceUrn: string;
  message: string;
}

export interface IacOperationResourceChangesSnapshotRow {
  resourceKind: string;
  resourceName: string;
  status?: string;
  isDone: boolean;
  isFailed?: boolean;
  elapsedDurationSeconds?: number | string;
  info?: string;
}

export interface IacOperationSummarySnapshot {
  durationSeconds?: number;
  unchangedResourceCount?: number;
  createResourceCount?: number;
  updateResourceCount?: number;
  deleteResourceCount?: number;
  replaceResourceCount?: number;
}

export interface StackJobProgressIacOperationSnapshot {
  preludeMessages?: string[];
  resourceRowMap: { [key: string]: IacOperationResourceChangesSnapshotRow };
  resourceDiffs?: string[];
  diagnosticMessages?: IacOperationDiagnosticMessage[];
  summary?: IacOperationSummarySnapshot;
  outputs?: { [key: string]: string };
}

export interface StackJobProgressIacOperationState {
  id?: string;
  isRequired: boolean;
  isApproved: boolean;
  status: WorkflowExecutionStatus | string;
  result: WorkflowExecutionResult | string;
  startTime: Timestamp | string;
  endTime: Timestamp | string;
  errors?: string[];
  snapshot?: StackJobProgressIacOperationSnapshot;
}

export interface StackJobStatusIacOperationsState {
  init?: StackJobProgressIacOperationState;
  refresh?: StackJobProgressIacOperationState;
  updatePreview?: StackJobProgressIacOperationState;
  destroyPreview?: StackJobProgressIacOperationState;
  update?: StackJobProgressIacOperationState;
  destroy?: StackJobProgressIacOperationState;
}

export interface IacResource {
  address: string;
  resourceType: string;
  provider: string;
  logicalName: string;
  resourceExternalId: string;
}

export interface StackJobProgressStatus {
  startTime: Timestamp | string;
  endTime: Timestamp | string;
  status: WorkflowExecutionStatus | string;
  result: WorkflowExecutionResult | string;
  errors?: string[];
}

export interface StackJobStatus {
  lifecycle?: ApiResourceLifecycle;
  audit: ApiResourceAudit;
  progress: StackJobProgressStatus;
  iacOperations: StackJobStatusIacOperationsState;
  iacResources: IacResource[];
}

export interface StackJobEssentials {
  stackJobRunnerInfo: ApiResourceKindAndNameAndId;
  iacModuleInfo: ApiResourceKindAndNameAndId;
  provisioner: IacProvisioner | string;
  backendCredentialInfo: ApiResourceKindAndNameAndId;
  flowControlPolicyInfo: ApiResourceKindAndNameAndId;
  providerCredentialInfo: ApiResourceKindAndNameAndId;
}

// CloudResource interface (simplified for StackJob usage)
export interface CloudResource {
  metadata: ApiResourceMetadata;
  spec: Record<string, unknown>; // This would be the full CloudResourceSpec
  status?: Record<string, unknown>; // This would be the full CloudResourceStatus
}

export interface CloudResourceMutator {
  // This would contain the mutator fields
  [key: string]: unknown;
}

export interface StackJobSpec {
  stackJobOperation: StackJobOperationType | string;
  essentials: StackJobEssentials;
  cloudResource: CloudResource;
  cloudResourceMutator: CloudResourceMutator;
}

export interface StackJob {
  apiVersion: string;
  kind: string;
  metadata: ApiResourceMetadata;
  spec: StackJobSpec;
  status: StackJobStatus;
}

// Progress event interfaces
export interface StackJobProgressEvent {
  type: StackJobProgressEventType;
  stackJobId: string;
  timestamp: Timestamp;
  data?: unknown;
}

// Input/Output interfaces
export interface StackJobId {
  stackJobId: string;
}

export interface StackJobList {
  entries: StackJob[];
  totalCount: number;
}

export interface NewStackJobInput {
  stackJobSpec: StackJobSpec;
  cloudResourceId: string;
}

// Utility types for form handling
export interface StackJobFormData {
  operationType: StackJobOperationType;
  essentials: Partial<StackJobEssentials>;
  cloudResource?: Partial<CloudResource>;
  cloudResourceMutator?: Partial<CloudResourceMutator>;
}

// Status display helpers
export interface StackJobStatusDisplay {
  status: string;
  result: string;
  isRunning: boolean;
  isCompleted: boolean;
  isFailed: boolean;
  isCancelled: boolean;
  progressPercentage?: number;
  estimatedTimeRemaining?: number;
}

// Operation status helpers
export interface IacOperationStatusDisplay {
  operation: string;
  status: string;
  result: string;
  isRequired: boolean;
  isApproved: boolean;
  isRunning: boolean;
  isCompleted: boolean;
  isFailed: boolean;
  startTime?: string;
  endTime?: string;
  duration?: number;
  errors: string[];
}
