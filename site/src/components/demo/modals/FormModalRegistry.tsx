'use client';

import React from 'react';
import { AwsCredentialModal } from './AwsCredentialModal';
import { AwsAlbModal } from './AwsAlbModal';
import { FormModalFactory } from './FormModalFactory';
import type { AwsAlb, AwsCredential } from '../interfaces';

interface FormModalRegistryProps {
  providerId: string;
  isOpen: boolean;
  onClose: () => void;
  /** The record the chosen provider's form produces; the placeholder for a provider with no form submits none. */
  onSubmit: (data?: AwsCredential | AwsAlb) => void;
}

// Registry of available form modals
const FORM_MODALS = {
  aws: AwsCredentialModal,
  'aws-alb': AwsAlbModal,
  // Add more providers here as they are implemented
} as const;

type SupportedProvider = keyof typeof FORM_MODALS;

export const FormModalRegistry: React.FC<FormModalRegistryProps> = ({
  providerId,
  isOpen,
  onClose,
  onSubmit,
}) => {
  // Check if provider is supported
  if (!(providerId in FORM_MODALS)) {
    // Fallback to generic modal with placeholder content
    return (
      <FormModalFactory
        isOpen={isOpen}
        onClose={onClose}
        onSubmit={onSubmit}
        FormComponent={({ onSubmit, onCancel }) => (
          <div className="h-full flex flex-col">
            <div className="bg-white border-b border-gray-200 px-6 py-4 flex-shrink-0">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-6 h-6 bg-gray-600 rounded flex items-center justify-center">
                    <svg className="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 20 20">
                      <path
                        fillRule="evenodd"
                        d="M3 3a1 1 0 000 2v8a2 2 0 002 2h2.586l-1.293 1.293a1 1 0 101.414 1.414L10 15.414l2.293 2.293a1 1 0 001.414-1.414L12.414 15H15a2 2 0 002-2V5a1 1 0 100-2H3zm11.707 4.707a1 1 0 00-1.414-1.414L10 9.586 8.707 8.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                        clipRule="evenodd"
                      />
                    </svg>
                  </div>
                  <h1 className="text-xl font-semibold text-gray-900">
                    Create {providerId} Credentials
                  </h1>
                </div>
                <div className="flex items-center gap-3">
                  <button
                    type="button"
                    onClick={onCancel}
                    className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-white"
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    onClick={() => onSubmit(undefined)}
                    className="px-4 py-2 text-sm font-medium text-white bg-[#111] rounded-lg hover:bg-[#111] focus:outline-none focus:ring-2 focus:ring-white"
                  >
                    Submit
                  </button>
                </div>
              </div>
            </div>
            <div className="flex-1 overflow-y-auto p-6">
              <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
                <div className="flex items-center gap-2">
                  <svg className="w-5 h-5 text-gray-600" fill="currentColor" viewBox="0 0 20 20">
                    <path
                      fillRule="evenodd"
                      d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                      clipRule="evenodd"
                    />
                  </svg>
                  <p className="text-gray-800 font-medium">{providerId} Credential Form</p>
                </div>
                <p className="text-gray-700 text-sm mt-1">
                  This provider is not yet implemented. The credential form will be added here.
                </p>
              </div>
            </div>
          </div>
        )}
      />
    );
  }

  // Get the appropriate modal component
  const ModalComponent = FORM_MODALS[providerId as SupportedProvider];

  return (
    <ModalComponent
      isOpen={isOpen}
      onClose={onClose}
      onSubmit={onSubmit}
    />
  );
};

// Export the registry and supported providers
export { FORM_MODALS };
export type { SupportedProvider };
