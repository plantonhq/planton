'use client';

import React, { useState, useEffect, useCallback, useRef } from 'react';
import Image from 'next/image';
import { AwsAlb } from '../interfaces';
import { AwsAlbFormElements } from './AwsAlbFormElements';
import { AccordionSection } from './AccordionSection';
import { MetadataSection } from './MetadataSection';
import { useAwsAlbData } from '../../../hooks/useAwsAlbData';
import { AutoFillProvider, useAutoFill } from '../../../contexts/AutoFillContext';
import { CURRENT_PRESET } from '../../../constants/animationConfig';
import { AnimatedFormField } from './AnimatedFormField';

interface AwsAlbFormProps {
  onSubmit: (data: AwsAlb) => void;
  onCancel: () => void;
  initialData?: Partial<AwsAlb>;
}

interface AwsAlbFormContentProps extends AwsAlbFormProps {
  onAnimationComplete?: () => void;
}

// Internal form component that uses the auto-fill context
const AwsAlbFormContent = React.forwardRef<
  { handleAnimationComplete: () => void; handleFieldComplete: (fieldName: string) => void },
  AwsAlbFormContentProps
>(({ onSubmit, onCancel, initialData, onAnimationComplete }, ref) => {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { data: jsonData } = useAwsAlbData();

  const { startAnimation, isAnimating, getFieldValue, isFieldAnimating, isFieldCompleted } =
    useAutoFill();

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
          { name: 'metadata.env', value: jsonData.metadata?.env || 'dev', order: 1 },
          { name: 'metadata.name', value: jsonData.metadata?.name || '', order: 2 },
          { name: 'metadata.slug', value: jsonData.metadata?.slug || '', order: 3 },
          { name: 'spec.subnets.0', value: jsonData.spec?.subnets?.[0]?.value || '', order: 4 },
          { name: 'spec.subnets.1', value: jsonData.spec?.subnets?.[1]?.value || '', order: 5 },
          {
            name: 'spec.securityGroups.0',
            value: jsonData.spec?.securityGroups?.[0]?.value || '',
            order: 6,
          },
          { name: 'spec.internal', value: jsonData.spec?.internal || false, order: 7 },
          {
            name: 'spec.deleteProtectionEnabled',
            value: jsonData.spec?.deleteProtectionEnabled || false,
            order: 8,
          },
          {
            name: 'spec.idleTimeoutSeconds',
            value: jsonData.spec?.idleTimeoutSeconds?.toString() || '60',
            order: 9,
          },
          { name: 'spec.dns.enabled', value: jsonData.spec?.dns?.enabled || false, order: 10 },
          {
            name: 'spec.dns.route53ZoneId',
            value: jsonData.spec?.dns?.route53ZoneId?.value || '',
            order: 11,
          },
          {
            name: 'spec.dns.hostnames.0',
            value: jsonData.spec?.dns?.hostnames?.[0] || '',
            order: 12,
          },
          { name: 'spec.ssl.enabled', value: jsonData.spec?.ssl?.enabled || false, order: 13 },
          {
            name: 'spec.ssl.certificateArn',
            value: jsonData.spec?.ssl?.certificateArn?.value || '',
            order: 14,
          },
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

  const [formData, setFormData] = useState<AwsAlb>({
    apiVersion: 'aws.planton.dev/v1alpha1',
    kind: 'AwsAlb',
    metadata: {
      name: initialData?.metadata?.name || '',
      slug: initialData?.metadata?.slug || '',
      id: initialData?.metadata?.id || '',
      org: initialData?.metadata?.org || '',
      env: initialData?.metadata?.env || '',
      labels: initialData?.metadata?.labels || {},
      annotations: initialData?.metadata?.annotations || {},
      tags: initialData?.metadata?.tags || [],
      version: initialData?.metadata?.version,
    },
    spec: {
      subnets: initialData?.spec?.subnets || [],
      securityGroups: initialData?.spec?.securityGroups || [],
      internal: initialData?.spec?.internal || false,
      deleteProtectionEnabled: initialData?.spec?.deleteProtectionEnabled || false,
      idleTimeoutSeconds: initialData?.spec?.idleTimeoutSeconds || 60,
      dns: initialData?.spec?.dns,
      ssl: initialData?.spec?.ssl,
    },
  });

  // Handle animation completion by updating the actual form data
  const handleAnimationCompleteInternal = useCallback(() => {
    if (jsonData) {
      setFormData((prev) => ({
        ...prev,
        metadata: {
          name: jsonData.metadata?.name || prev.metadata.name,
          slug: jsonData.metadata?.slug || prev.metadata.slug,
          id: jsonData.metadata?.id || prev.metadata.id,
          org: jsonData.metadata?.org || prev.metadata.org,
          env: jsonData.metadata?.env || prev.metadata.env,
          labels: jsonData.metadata?.labels || prev.metadata.labels,
          annotations: jsonData.metadata?.annotations || prev.metadata.annotations,
          tags: jsonData.metadata?.tags || prev.metadata.tags,
          version: jsonData.metadata?.version || prev.metadata.version,
        },
        spec: {
          subnets: jsonData.spec?.subnets || prev.spec.subnets,
          securityGroups: jsonData.spec?.securityGroups || prev.spec.securityGroups,
          internal: jsonData.spec?.internal ?? prev.spec.internal,
          deleteProtectionEnabled:
            jsonData.spec?.deleteProtectionEnabled ?? prev.spec.deleteProtectionEnabled,
          idleTimeoutSeconds: jsonData.spec?.idleTimeoutSeconds || prev.spec.idleTimeoutSeconds,
          dns: jsonData.spec?.dns || prev.spec.dns,
          ssl: jsonData.spec?.ssl || prev.spec.ssl,
        },
      }));
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
            const metadataValue = jsonData.metadata?.[child as keyof typeof jsonData.metadata];
            value = typeof metadataValue === 'string' ? metadataValue : '';
          } else if (parent === 'spec') {
            const specValue = jsonData.spec?.[child as keyof typeof jsonData.spec];
            value = typeof specValue === 'string' ? specValue : '';
          }

          if (value) {
            setFormData((prev) => {
              const newData = { ...prev };
              if (parent === 'metadata') {
                newData.metadata = {
                  ...newData.metadata,
                  [child]: value,
                };
              } else if (parent === 'spec') {
                newData.spec = {
                  ...newData.spec,
                  [child]: value,
                };
              }
              return newData;
            });
          }
        } else if (keys.length === 3) {
          // Handle nested fields like 'spec.subnets.0' or 'spec.dns.enabled'
          const [parent, child, grandchild] = keys;
          if (parent === 'spec' && child === 'subnets') {
            const index = parseInt(grandchild);
            const subnetValue = jsonData.spec?.subnets?.[index]?.value;
            if (subnetValue) {
              setFormData((prev) => {
                const newSubnets = [...(prev.spec.subnets || [])];
                newSubnets[index] = { value: subnetValue };
                return {
                  ...prev,
                  spec: {
                    ...prev.spec,
                    subnets: newSubnets,
                  },
                };
              });
            }
          } else if (parent === 'spec' && child === 'securityGroups') {
            const index = parseInt(grandchild);
            const sgValue = jsonData.spec?.securityGroups?.[index]?.value;
            if (sgValue) {
              setFormData((prev) => {
                const newSecurityGroups = [...(prev.spec.securityGroups || [])];
                newSecurityGroups[index] = { value: sgValue };
                return {
                  ...prev,
                  spec: {
                    ...prev.spec,
                    securityGroups: newSecurityGroups,
                  },
                };
              });
            }
          } else if (parent === 'spec' && child === 'dns') {
            if (grandchild === 'enabled') {
              setFormData((prev) => ({
                ...prev,
                spec: {
                  ...prev.spec,
                  dns: {
                    enabled: jsonData.spec?.dns?.enabled || false,
                    route53ZoneId: prev.spec.dns?.route53ZoneId,
                    hostnames: prev.spec.dns?.hostnames || [],
                  },
                },
              }));
            } else if (grandchild === 'route53ZoneId') {
              setFormData((prev) => ({
                ...prev,
                spec: {
                  ...prev.spec,
                  dns: {
                    enabled: prev.spec.dns?.enabled || false,
                    route53ZoneId:
                      jsonData.spec?.dns?.route53ZoneId || prev.spec.dns?.route53ZoneId,
                    hostnames: prev.spec.dns?.hostnames || [],
                  },
                },
              }));
            }
          } else if (parent === 'spec' && child === 'ssl') {
            if (grandchild === 'enabled') {
              setFormData((prev) => ({
                ...prev,
                spec: {
                  ...prev.spec,
                  ssl: {
                    enabled: jsonData.spec?.ssl?.enabled || false,
                    certificateArn: prev.spec.ssl?.certificateArn,
                  },
                },
              }));
            } else if (grandchild === 'certificateArn') {
              setFormData((prev) => ({
                ...prev,
                spec: {
                  ...prev.spec,
                  ssl: {
                    enabled: prev.spec.ssl?.enabled || false,
                    certificateArn:
                      jsonData.spec?.ssl?.certificateArn || prev.spec.ssl?.certificateArn,
                  },
                },
              }));
            }
          }
        } else if (keys.length === 4) {
          // Handle very deeply nested fields like 'spec.dns.hostnames.0'
          const [parent, child, grandchild, greatGrandchild] = keys;
          if (parent === 'spec' && child === 'dns' && grandchild === 'hostnames') {
            const index = parseInt(greatGrandchild);
            const hostnameValue = jsonData.spec?.dns?.hostnames?.[index];
            if (hostnameValue) {
              setFormData((prev) => {
                const newHostnames = [...(prev.spec.dns?.hostnames || [])];
                newHostnames[index] = hostnameValue;
                return {
                  ...prev,
                  spec: {
                    ...prev.spec,
                    dns: {
                      enabled: prev.spec.dns?.enabled || false,
                      route53ZoneId: prev.spec.dns?.route53ZoneId,
                      hostnames: newHostnames,
                    },
                  },
                };
              });
            }
          }
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

  const handleInputChange = (field: string, value: unknown) => {
    setFormData((prev) => {
      const newData = { ...prev };
      const keys = field.split('.');

      if (keys.length === 2) {
        // Handle nested fields like 'metadata.name' or 'spec.internal'
        const [parent, child] = keys;
        if (parent === 'metadata') {
          newData.metadata = {
            name: newData.metadata.name,
            slug: newData.metadata.slug,
            id: newData.metadata.id,
            org: newData.metadata.org,
            env: newData.metadata.env,
            labels: newData.metadata.labels,
            annotations: newData.metadata.annotations,
            tags: newData.metadata.tags,
            version: newData.metadata.version,
            [child]: value,
          };
        } else if (parent === 'spec') {
          newData.spec = {
            subnets: newData.spec.subnets,
            securityGroups: newData.spec.securityGroups,
            internal: newData.spec.internal,
            deleteProtectionEnabled: newData.spec.deleteProtectionEnabled,
            idleTimeoutSeconds: newData.spec.idleTimeoutSeconds,
            dns: newData.spec.dns,
            ssl: newData.spec.ssl,
            [child]: value,
          };
        }
      } else if (keys.length === 3) {
        // Handle deeply nested fields like 'spec.dns.enabled'
        const [parent, child, grandchild] = keys;
        if (parent === 'spec' && child === 'dns') {
          newData.spec = {
            ...newData.spec,
            dns: {
              enabled: newData.spec.dns?.enabled || false,
              route53ZoneId: newData.spec.dns?.route53ZoneId,
              hostnames: newData.spec.dns?.hostnames || [],
              [grandchild]: value,
            },
          };
        } else if (parent === 'spec' && child === 'ssl') {
          newData.spec = {
            ...newData.spec,
            ssl: {
              enabled: newData.spec.ssl?.enabled || false,
              certificateArn: newData.spec.ssl?.certificateArn,
              [grandchild]: value,
            },
          };
        }
      } else if (keys.length === 4) {
        // Handle very deeply nested fields like 'spec.dns.route53ZoneId.value'
        const [parent, child, grandchild, greatGrandchild] = keys;
        if (parent === 'spec' && child === 'dns' && grandchild === 'route53ZoneId') {
          newData.spec = {
            ...newData.spec,
            dns: {
              enabled: newData.spec.dns?.enabled || false,
              route53ZoneId: {
                ...newData.spec?.dns?.route53ZoneId,
                [greatGrandchild]: value,
              },
              hostnames: newData.spec.dns?.hostnames || [],
            },
          };
        } else if (parent === 'spec' && child === 'ssl' && grandchild === 'certificateArn') {
          newData.spec = {
            ...newData.spec,
            ssl: {
              enabled: newData.spec.ssl?.enabled || false,
              certificateArn: {
                ...newData.spec?.ssl?.certificateArn,
                [greatGrandchild]: value,
              },
            },
          };
        }
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

  const handleFormSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    setIsSubmitting(true);
    try {
      await onSubmit(formData);
    } finally {
      setIsSubmitting(false);
    }
  };

  const helpTextMapping = {
    'metadata.name': 'Enter a unique name for your AWS ALB',
    'spec.subnets': 'List of subnet IDs where the ALB will be deployed (minimum 2 required)',
    'spec.securityGroups': 'Security groups to attach to the ALB',
    'spec.internal': 'Whether the ALB should be internal (not internet-facing)',
    'spec.deleteProtectionEnabled': 'Enable deletion protection to prevent accidental deletion',
    'spec.idleTimeoutSeconds': 'Connection idle timeout in seconds (default: 60)',
    'spec.dns.enabled': 'Enable Route53 DNS configuration',
    'spec.dns.route53ZoneId': 'Route53 hosted zone ID for DNS records',
    'spec.dns.hostnames': 'Domain names that will point to this ALB',
    'spec.ssl.enabled': 'Enable SSL/TLS termination',
    'spec.ssl.certificateArn': 'ARN of the SSL certificate from AWS Certificate Manager',
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
              <h1 className="text-xl font-semibold text-gray-900">Deploy AWS ALB</h1>
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
                  'Deploy ALB'
                )}
              </button>
            </div>
          </div>
        </div>

        {/* Scrollable Content Area */}
        <div className="flex-1 overflow-y-auto">
          <div className="p-6 pb-12">
            {/* Accordion Sections */}
            <div className="space-y-6">
              {/* Metadata Section */}
              <AccordionSection title="Metadata">
                <div className="space-y-4">
                  {/* Environment Dropdown */}
                  <AnimatedFormField
                    label="Environment"
                    value={formData.metadata?.env || 'dev'}
                    onChange={(value) => handleInputChange('metadata.env', value)}
                    type="select"
                    fieldName="metadata.env"
                    isAnimating={isAnimating}
                    isCurrentField={isFieldAnimating('metadata.env')}
                    isCompleted={isFieldCompleted('metadata.env')}
                    animatedValue={getFieldValue('metadata.env')}
                    options={[
                      { value: 'dev', label: 'dev' },
                      { value: 'staging', label: 'staging' },
                      { value: 'prod', label: 'prod' },
                    ]}
                  />

                  {/* Name and Slug using MetadataSection */}
                  <MetadataSection
                    name={formData.metadata?.name || ''}
                    slug={formData.metadata?.slug || ''}
                    onNameChange={handleMetadataNameChange}
                    onSlugChange={handleMetadataSlugChange}
                    nameHelpText={helpTextMapping['metadata.name']}
                  />
                </div>
              </AccordionSection>

              {/* Subnets Section */}
              <AccordionSection title="Subnets">
                <AwsAlbFormElements
                  formData={formData}
                  onInputChange={handleInputChange}
                  helpTextMapping={helpTextMapping}
                  section="subnets"
                />
              </AccordionSection>

              {/* Security Groups Section */}
              <AccordionSection title="Security Groups">
                <AwsAlbFormElements
                  formData={formData}
                  onInputChange={handleInputChange}
                  helpTextMapping={helpTextMapping}
                  section="securityGroups"
                />
              </AccordionSection>

              {/* Load Balancer Settings Section */}
              <AccordionSection title="Load Balancer Settings">
                <AwsAlbFormElements
                  formData={formData}
                  onInputChange={handleInputChange}
                  helpTextMapping={helpTextMapping}
                  section="loadBalancer"
                />
              </AccordionSection>

              {/* DNS Section */}
              <AccordionSection title="DNS">
                <AwsAlbFormElements
                  formData={formData}
                  onInputChange={handleInputChange}
                  helpTextMapping={helpTextMapping}
                  section="dns"
                />
              </AccordionSection>

              {/* SSL Section */}
              <AccordionSection title="SSL">
                <AwsAlbFormElements
                  formData={formData}
                  onInputChange={handleInputChange}
                  helpTextMapping={helpTextMapping}
                  section="ssl"
                />
              </AccordionSection>
            </div>
          </div>
        </div>
      </form>
    </div>
  );
});

AwsAlbFormContent.displayName = 'AwsAlbFormContent';

// Main exported component that wraps the form with AutoFillProvider
export const AwsAlbForm: React.FC<AwsAlbFormProps> = (props) => {
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
      <AwsAlbFormContent {...props} ref={formContentRef} />
    </AutoFillProvider>
  );
};
