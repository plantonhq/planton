'use client';

import React, { useState, useEffect, useCallback, useRef } from 'react';
import Image from 'next/image';
import { AwsCredential } from '../interfaces';
import { AwsCredentialFormElements } from './AwsCredentialFormElements';
import { AccordionSection } from './AccordionSection';
import { MetadataSection } from './MetadataSection';
import { AnimatedFormField } from './AnimatedFormField';
import { useAwsCredentialData } from '../../../hooks/useAwsCredentialData';
import { AutoFillProvider, useAutoFill } from '../../../contexts/AutoFillContext';
import { CURRENT_PRESET } from '../../../constants/animationConfig';

interface AwsCredentialFormProps {
  onSubmit: (data: AwsCredential) => void;
  onCancel: () => void;
  initialData?: Partial<AwsCredential>;
}

interface AwsCredentialFormContentProps extends AwsCredentialFormProps {
  onAnimationComplete?: () => void;
}

// Internal form component that uses the auto-fill context
const AwsCredentialFormContent = React.forwardRef<
  { handleAnimationComplete: () => void; handleFieldComplete: (fieldName: string) => void },
  AwsCredentialFormContentProps
>(({ onSubmit, onCancel, initialData, onAnimationComplete }, ref) => {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [authorizedEnvironments, setAuthorizedEnvironments] = useState<string[]>([]);
  const { data: jsonData } = useAwsCredentialData();

  const { startAnimation, isAnimating, getFieldValue, isFieldAnimating, isFieldCompleted } =
    useAutoFill();

  const appDevRef = React.useRef<HTMLDivElement>(null);
  const appProdRef = React.useRef<HTMLDivElement>(null);

  // Track if animation has already been started to prevent looping
  const hasAnimationStartedRef = useRef(false);
  const animationTimerRef = useRef<NodeJS.Timeout | null>(null);

  // Auto-start animation when data is loaded (only once)
  useEffect(() => {
    const isIOS =
      /iPad|iPhone|iPod/.test(navigator.userAgent) ||
      /iPad|iPhone|iPod/.test(navigator.platform) ||
      (navigator.maxTouchPoints &&
        navigator.maxTouchPoints > 2 &&
        /MacIntel/.test(navigator.platform));

    if (jsonData && !hasAnimationStartedRef.current) {
      hasAnimationStartedRef.current = true;

      // iOS-specific timing adjustments
      const baseDelay = Math.max(CURRENT_PRESET.autoStartDelay, 2000);
      const animationDelay = isIOS ? baseDelay + 2000 : baseDelay; // Extra 2 seconds for iOS

      // Clear any existing timer
      if (animationTimerRef.current) {
        clearTimeout(animationTimerRef.current);
        animationTimerRef.current = null;
      }

      animationTimerRef.current = setTimeout(() => {
        const fields = [
          { name: 'metadata.name', value: jsonData.metadata.name || '', order: 1 },
          { name: 'metadata.slug', value: jsonData.metadata.slug || '', order: 2 },
          { name: 'spec.accountId', value: jsonData.spec.accountId || '', order: 3 },
          { name: 'spec.region', value: jsonData.spec.region || '', order: 4 },
          { name: 'spec.accessKeyId', value: jsonData.spec.accessKeyId || '', order: 5 },
          { name: 'spec.secretAccessKey', value: jsonData.spec.secretAccessKey || '', order: 6 },
          { name: 'authorizedEnvironments.app-dev', value: true, order: 7 },
          { name: 'authorizedEnvironments.app-prod', value: true, order: 8 },
        ];

        try {
          startAnimation(fields);
        } catch (error) {
          console.error('Error calling startAnimation:', error);
        }

        // Clear the timer reference after it fires
        animationTimerRef.current = null;
      }, animationDelay);

      // Don't clean up the timer in the return function - let it fire
      return () => {
        // Timer will fire even if component unmounts
      };
    } else if (!jsonData && !hasAnimationStartedRef.current) {
      // Data not loaded yet, set up a retry mechanism
      const retryDelay = isIOS ? 1000 : 500; // Longer retry delay for iOS
      const retryTimer = setTimeout(() => {
        // This will trigger the useEffect again when jsonData becomes available
      }, retryDelay);

      return () => {
        clearTimeout(retryTimer);
      };
    }
  }, [jsonData, startAnimation]);

  const [formData, setFormData] = useState<AwsCredential>({
    metadata: {
      name: '',
      slug: '',
      id: '',
      org: '',
      env: '',
      labels: {},
      annotations: {},
      tags: [],
      ...initialData?.metadata,
    },
    spec: {
      accountId: '',
      region: '',
      accessKeyId: '',
      secretAccessKey: '',
      ...initialData?.spec,
    },
  });

  // Note: Removed automatic data filling to allow animation system to handle it
  // The animation system will fill the data when the user clicks "Auto Fill"

  const handleInputChange = (field: string, value: string) => {
    setFormData((prev) => {
      const newData = { ...prev };
      const keys = field.split('.');

      if (keys.length === 2) {
        // Handle nested fields like 'metadata.name' or 'spec.accountId'
        const [parent, child] = keys;
        if (parent === 'metadata') newData.metadata = { ...newData.metadata, [child]: value };
        else if (parent === 'spec') newData.spec = { ...newData.spec, [child]: value };
      }

      return newData;
    });
  };

  const handleMetadataNameChange = useCallback((value: string) => {
    setFormData((prev) => ({
      ...prev,
      metadata: {
        ...prev.metadata,
        name: value,
      },
    }));
  }, []);

  const handleMetadataSlugChange = useCallback((value: string) => {
    setFormData((prev) => ({
      ...prev,
      metadata: {
        ...prev.metadata,
        slug: value,
      },
    }));
  }, []);

  const handleEnvironmentToggle = (environment: string) => {
    setAuthorizedEnvironments((prev) => {
      if (prev.includes(environment)) {
        return prev.filter((env) => env !== environment);
      } else {
        return [...prev, environment];
      }
    });
  };

  const handleFormSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    setIsSubmitting(true);
    try {
      // Include authorized environments in the form data
      const formDataWithEnvironments = {
        ...formData,
        authorizedEnvironments,
      };
      await onSubmit(formDataWithEnvironments);
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle animation completion by updating the actual form data
  const handleAnimationCompleteInternal = useCallback(() => {
    if (jsonData) {
      setFormData((prev) => ({
        ...prev,
        metadata: {
          ...prev.metadata,
          name: jsonData.metadata.name || prev.metadata.name,
          slug: jsonData.metadata.slug || prev.metadata.slug,
        },
        spec: {
          ...prev.spec,
          accountId: jsonData.spec.accountId || prev.spec.accountId,
          region: jsonData.spec.region || prev.spec.region,
          accessKeyId: jsonData.spec.accessKeyId || prev.spec.accessKeyId,
          secretAccessKey: jsonData.spec.secretAccessKey || prev.spec.secretAccessKey,
        },
      }));

      // Set authorized environments
      setAuthorizedEnvironments(['app-dev', 'app-prod']);
    }
    onAnimationComplete?.();
  }, [jsonData, onAnimationComplete]);

  // Handle field completion during animation to update form data in real-time
  const handleFieldComplete = useCallback(
    (fieldName: string) => {
      if (jsonData) {
        const keys = fieldName.split('.');
        if (keys.length === 2) {
          const [parent, child] = keys;
          let value = '';

          if (parent === 'metadata') {
            value = (jsonData.metadata[child as keyof typeof jsonData.metadata] as string) || '';
          } else if (parent === 'spec') {
            value = (jsonData.spec[child as keyof typeof jsonData.spec] as string) || '';
          }

          if (value) {
            setFormData((prev) => {
              const newData = { ...prev };
              if (parent === 'metadata') newData.metadata = { ...newData.metadata, [child]: value };
              else if (parent === 'spec') newData.spec = { ...newData.spec, [child]: value };
              return newData;
            });
          }
        } else if (keys.length === 3 && keys[0] === 'authorizedEnvironments') {
          // Handle checkbox fields like 'authorizedEnvironments.app-dev'
          const environment = keys[2];
          setAuthorizedEnvironments((prev) => {
            if (!prev.includes(environment)) {
              return [...prev, environment];
            }
            return prev;
          });
        }
      }
    },
    [jsonData]
  );

  // Expose the completion handlers to parent
  React.useImperativeHandle(ref, () => ({
    handleAnimationComplete: handleAnimationCompleteInternal,
    handleFieldComplete: handleFieldComplete,
  }));

  const helpTextMapping = {
    'metadata.name': 'Enter a unique name for your AWS credential',
    'spec.accountId': 'Your 12-digit AWS account ID',
    'spec.region': 'Default AWS region for deployments',
    'spec.accessKeyId': 'Your AWS access key ID (starts with AKIA)',
    'spec.secretAccessKey': 'Your AWS secret access key',
  };

  return (
    <div className="mx-auto bg-gray-50 min-h-screen">
      <form onSubmit={handleFormSubmit} className="h-screen flex flex-col">
        {/* Fixed Header */}
        <div className="bg-white border-b border-gray-200 px-6 py-4 flex-shrink-0">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-6 h-6 bg-white rounded flex items-center justify-center shadow-sm">
                <Image src="/_site/images/resources/aws.svg" alt="AWS icon" width={16} height={16} className="w-4 h-4" />
              </div>
              <h1 className="text-xl font-semibold text-gray-900">Create AWS Credentials</h1>
            </div>
            <div className="flex items-center gap-3">
              <button
                type="button"
                onClick={onCancel}
                className="px-4 py-2 text-sm font-medium text-gray-800 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-white"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={isSubmitting}
                className="px-4 py-2 text-sm font-medium text-white bg-gray-800 rounded-lg hover:bg-gray-900 focus:outline-none focus:ring-2 focus:ring-white disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isSubmitting ? (
                  <>
                    <svg
                      className="animate-spin -ml-1 mr-2 h-4 w-4 text-white inline"
                      fill="none"
                      viewBox="0 0 24 24"
                    >
                      <circle
                        className="opacity-25"
                        cx="12"
                        cy="12"
                        r="10"
                        stroke="currentColor"
                        strokeWidth="4"
                      ></circle>
                      <path
                        className="opacity-75"
                        fill="currentColor"
                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                      ></path>
                    </svg>
                    Creating...
                  </>
                ) : (
                  'Submit'
                )}
              </button>
            </div>
          </div>
        </div>

        {/* Scrollable Content Area */}
        <div className="flex-1 overflow-y-auto">
          <div className="p-6 pb-12 space-y-6">
            {/* Metadata Section */}
            <AccordionSection title="Metadata">
              <MetadataSection
                name={formData.metadata.name}
                slug={formData.metadata.slug}
                onNameChange={handleMetadataNameChange}
                onSlugChange={handleMetadataSlugChange}
                nameHelpText={helpTextMapping['metadata.name']}
              />
            </AccordionSection>

            {/* AWS Account Details Section */}
            <AccordionSection title="AWS Account Details">
              <AwsCredentialFormElements
                formData={formData}
                onInputChange={handleInputChange}
                helpTextMapping={helpTextMapping}
              />
            </AccordionSection>

            {/* Authorized Environments Section */}
            <AccordionSection title="Authorized Environments">
              <div className="space-y-3">
                {['app-dev', 'app-prod'].map((environment) => (
                  <AnimatedFormField
                    key={environment}
                    ref={environment === 'app-dev' ? appDevRef : appProdRef}
                    label={environment}
                    value={authorizedEnvironments.includes(environment)}
                    onChange={(checked) => {
                      if (typeof checked === 'boolean') {
                        handleEnvironmentToggle(environment);
                      }
                    }}
                    type="checkbox"
                    fieldName={`authorizedEnvironments.${environment}`}
                    isAnimating={isAnimating}
                    isCurrentField={isFieldAnimating(`authorizedEnvironments.${environment}`)}
                    isCompleted={isFieldCompleted(`authorizedEnvironments.${environment}`)}
                    animatedValue={getFieldValue(`authorizedEnvironments.${environment}`)}
                  />
                ))}
              </div>
            </AccordionSection>
          </div>
        </div>
      </form>
    </div>
  );
});

AwsCredentialFormContent.displayName = 'AwsCredentialFormContent';

// Main exported component that wraps the form with AutoFillProvider
export const AwsCredentialForm: React.FC<AwsCredentialFormProps> = (props) => {
  const formContentRef = React.useRef<{
    handleAnimationComplete: () => void;
    handleFieldComplete: (fieldName: string) => void;
  } | null>(null);

  return (
    <AutoFillProvider
      delay={CURRENT_PRESET.delay}
      pauseBetweenFields={CURRENT_PRESET.pauseBetweenFields}
      onComplete={() => {
        formContentRef.current?.handleAnimationComplete();
      }}
      onFieldComplete={(fieldName) => {
        formContentRef.current?.handleFieldComplete(fieldName);
      }}
    >
      <AwsCredentialFormContent {...props} ref={formContentRef} />
    </AutoFillProvider>
  );
};
