'use client';

import React from 'react';
import { CheckCircle } from 'lucide-react';
import {
  StackJob,
  StackJobStatus,
  Timestamp,
  WorkflowExecutionStatus,
  WorkflowExecutionResult,
} from '../interfaces';

interface StackJobHeaderProps {
  stackJob: StackJob;
  stackJobStatus: StackJobStatus;
}

const StackJobHeader: React.FC<StackJobHeaderProps> = ({ stackJob, stackJobStatus }) => {
  // Calculate duration
  const getDuration = () => {
    if (stackJobStatus.progress.startTime && stackJobStatus.progress.endTime) {
      const start = typeof stackJobStatus.progress.startTime === 'string' 
        ? new Date(stackJobStatus.progress.startTime).getTime()
        : stackJobStatus.progress.startTime.seconds * 1000;
      const end = typeof stackJobStatus.progress.endTime === 'string'
        ? new Date(stackJobStatus.progress.endTime).getTime()
        : stackJobStatus.progress.endTime.seconds * 1000;
      const durationMs = end - start;
      const minutes = Math.floor(durationMs / 60000);
      const seconds = Math.floor((durationMs % 60000) / 1000);
      return `${minutes}m ${seconds}s`;
    }
    return '0s';
  };

  // Get operation time
  const getOperationTime = (operation: { startTime?: Timestamp | string; endTime?: Timestamp | string } | undefined) => {
    if (operation?.endTime && operation?.startTime) {
      const start = typeof operation.startTime === 'string' 
        ? new Date(operation.startTime).getTime()
        : operation.startTime.seconds * 1000;
      const end = typeof operation.endTime === 'string'
        ? new Date(operation.endTime).getTime()
        : operation.endTime.seconds * 1000;
      const durationMs = end - start;
      const minutes = Math.floor(durationMs / 60000);
      const seconds = Math.floor((durationMs % 60000) / 1000);
      if (minutes > 0) {
        return `${minutes}m ${seconds}s`;
      }
      return `${seconds}s`;
    }
    return '';
  };

  // Get triggered time ago
  const getTriggeredTime = () => {
    if (stackJob.status.audit.specAudit?.createdAt) {
      const createdAt = stackJob.status.audit.specAudit.createdAt;
      const created = new Date(createdAt.seconds * 1000);
      const now = new Date();
      const diffMs = now.getTime() - created.getTime();
      const hours = Math.floor(diffMs / 3600000);
      if (hours > 0) {
        return `${hours} hours ago`;
      }
      const minutes = Math.floor(diffMs / 60000);
      if (minutes > 0) {
        return `${minutes} minutes ago`;
      }
      return 'just now';
    }
    return '';
  };

  // Get operations in display order
  const operations = [
    { key: 'init', operation: stackJobStatus.iacOperations.init, title: 'init' },
    { key: 'refresh', operation: stackJobStatus.iacOperations.refresh, title: 'refresh' },
    { key: 'updatePreview', operation: stackJobStatus.iacOperations.updatePreview, title: 'preview' },
    { key: 'update', operation: stackJobStatus.iacOperations.update, title: 'up' },
  ].filter(({ operation }) => operation && operation.isRequired);

  return (
    <div className="p-6">
      <div className="max-w-6xl mx-auto">
        <div className="bg-gray-50 rounded-lg border border-gray-200 p-4">
          {/* Top Section */}
          <div className="flex items-start justify-between mb-4">
            <div className="flex-1">
              {/* Title */}
              <h2 className="text-base font-medium text-gray-900 mb-2">
                {stackJob.metadata.version?.message || 'Stack Job'}
              </h2>
              
              {/* Version Info Line */}
              <div className="flex items-center gap-2 text-sm">
                <span className="bg-[#1a1a1a] text-gray-800 px-2 py-0.5 rounded font-mono text-xs">
                  {stackJob.spec?.cloudResource?.metadata?.version?.id?.slice(-7) || 'j5g7m6k'}
                </span>
                <span className="text-gray-500">
                  {typeof stackJob.spec.essentials.provisioner === 'string' 
                    ? stackJob.spec.essentials.provisioner 
                    : stackJob.spec.essentials.provisioner.type} {stackJob.spec.stackJobOperation} by
                </span>
                <div className="flex items-center gap-1">
                  {stackJob.status.audit.specAudit?.createdBy?.avatar && (
                    // eslint-disable-next-line @next/next/no-img-element -- the avatar is the run record's own URL, dimensions unknown; images are unoptimized on this static export
                    <img 
                      src={stackJob.status.audit.specAudit.createdBy.avatar} 
                      alt={stackJob.status.audit.specAudit.createdBy.id} 
                      className="w-4 h-4 rounded-full"
                    />
                  )}
                  <span className="text-gray-900 font-medium">
                    {stackJob.status.audit.specAudit?.createdBy?.id?.split('@')[0] || 'Unknown'}
                  </span>
                </div>
              </div>
            </div>
            
            {/* Triggered Section */}
            <div className="flex items-center gap-2 text-sm">
              <span className="text-gray-500">Triggered</span>
              <span className="bg-gray-100 text-gray-700 px-2 py-0.5 rounded">
                {stackJob.spec.stackJobOperation}
              </span>
              <span className="text-gray-500">{getTriggeredTime()}</span>
            </div>
          </div>
          
          {/* Divider */}
          <div className="border-t border-gray-200 my-3"></div>
          
          {/* Bottom Section */}
          <div className="flex items-center justify-between">
            {/* Operations Status */}
            <div className="flex items-center gap-6">
              {operations.map(({ key, title, operation }) => {
                const result = typeof operation?.result === 'string' ? operation.result : WorkflowExecutionResult[operation?.result as WorkflowExecutionResult];
                const status = typeof operation?.status === 'string' ? operation.status : WorkflowExecutionStatus[operation?.status as WorkflowExecutionStatus];
                
                return (
                  <div key={key} className="flex items-center gap-2">
                    <CheckCircle className={`w-4 h-4 ${
                      result === 'succeeded' ? 'text-[#10b981]' :
                      result === 'failed' ? 'text-red-600' :
                      status === 'running' ? 'text-white' :
                      'text-gray-400'
                    }`} />
                    <span className="text-sm text-gray-700">
                      {typeof stackJob.spec.essentials.provisioner === 'string' 
                        ? stackJob.spec.essentials.provisioner 
                        : 'pulumi'} {title}
                    </span>
                    <span className="text-sm text-gray-500">
                      {getOperationTime(operation)}
                    </span>
                  </div>
                );
              })}
              
              {/* Status Badge */}
              <div className="flex items-center gap-2 ml-4">
                <CheckCircle className={`w-5 h-5 ${
                  (typeof stackJobStatus.progress.result === 'string' ? stackJobStatus.progress.result : WorkflowExecutionResult[stackJobStatus.progress.result as WorkflowExecutionResult]) === 'succeeded'
                    ? 'text-[#10b981]'
                    : (typeof stackJobStatus.progress.result === 'string' ? stackJobStatus.progress.result : WorkflowExecutionResult[stackJobStatus.progress.result as WorkflowExecutionResult]) === 'failed'
                    ? 'text-red-600'
                    : 'text-white'
                }`} />
                <span className={`text-sm font-medium ${
                  (typeof stackJobStatus.progress.result === 'string' ? stackJobStatus.progress.result : WorkflowExecutionResult[stackJobStatus.progress.result as WorkflowExecutionResult]) === 'succeeded'
                    ? 'text-[#10b981]'
                    : (typeof stackJobStatus.progress.result === 'string' ? stackJobStatus.progress.result : WorkflowExecutionResult[stackJobStatus.progress.result as WorkflowExecutionResult]) === 'failed'
                    ? 'text-red-700'
                    : 'text-white'
                }`}>
                  {(typeof stackJobStatus.progress.result === 'string' ? stackJobStatus.progress.result : WorkflowExecutionResult[stackJobStatus.progress.result as WorkflowExecutionResult]) || 
                   (typeof stackJobStatus.progress.status === 'string' ? stackJobStatus.progress.status : WorkflowExecutionStatus[stackJobStatus.progress.status as WorkflowExecutionStatus])}
                </span>
              </div>
            </div>
            
            {/* Total Duration */}
            <div className="text-right">
              <div className="text-xs text-gray-500 mb-1">Total duration</div>
              <div className="text-base font-medium text-gray-900">{getDuration()}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default StackJobHeader;
