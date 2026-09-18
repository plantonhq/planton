'use client';

import React, { useRef } from 'react';
import { AwsAlb } from '../interfaces';
import { AnimatedFormField } from './AnimatedFormField';
import { FieldArrayWithTabs } from './FieldArrayWithTabs';
import { useAutoFill } from '../../../contexts/AutoFillContext';

interface AwsAlbFormElementsProps {
  formData: AwsAlb;
  onInputChange: (field: string, value: unknown) => void;
  helpTextMapping?: { [key: string]: string };
  section?: 'subnets' | 'securityGroups' | 'loadBalancer' | 'dns' | 'ssl';
}

export const AwsAlbFormElements: React.FC<AwsAlbFormElementsProps> = ({
  formData,
  onInputChange,
  helpTextMapping,
  section,
}) => {
  const { isAnimating, getFieldValue, isFieldAnimating, isFieldCompleted } = useAutoFill();

  // Refs for animated fields
  const internalRef = useRef<HTMLDivElement>(null);
  const deleteProtectionRef = useRef<HTMLDivElement>(null);
  const idleTimeoutRef = useRef<HTMLDivElement>(null);
  const dnsEnabledRef = useRef<HTMLDivElement>(null);
  const route53ZoneIdRef = useRef<HTMLDivElement>(null);
  const hostnameRef = useRef<HTMLDivElement>(null);
  const sslEnabledRef = useRef<HTMLDivElement>(null);
  const certificateArnRef = useRef<HTMLDivElement>(null);

  // Render specific section based on section prop
  if (section === 'subnets') {
    return (
      <FieldArrayWithTabs
        items={formData.spec?.subnets || []}
        onItemsChange={(items) => onInputChange('spec.subnets', items)}
        fieldName="spec.subnets"
        helpText={helpTextMapping?.['spec.subnets']}
      />
    );
  }

  if (section === 'securityGroups') {
    return (
      <FieldArrayWithTabs
        items={formData.spec?.securityGroups || []}
        onItemsChange={(items) => onInputChange('spec.securityGroups', items)}
        fieldName="spec.securityGroups"
        helpText={helpTextMapping?.['spec.securityGroups']}
      />
    );
  }

  if (section === 'loadBalancer') {
    return (
      <div>
        <AnimatedFormField
          ref={idleTimeoutRef}
          label="Idle Timeout (seconds)"
          value={formData.spec?.idleTimeoutSeconds?.toString() || '60'}
          onChange={(value) =>
            onInputChange('spec.idleTimeoutSeconds', parseInt(value as string) || 60)
          }
          type="number"
          helpText={helpTextMapping?.['spec.idleTimeoutSeconds']}
          fieldName="spec.idleTimeoutSeconds"
          isAnimating={isAnimating}
          isCurrentField={isFieldAnimating('spec.idleTimeoutSeconds')}
          isCompleted={isFieldCompleted('spec.idleTimeoutSeconds')}
          animatedValue={getFieldValue('spec.idleTimeoutSeconds')}
        />

        <div className="mt-4 space-y-3">
          <AnimatedFormField
            ref={internalRef}
            label="Internal Load Balancer"
            value={formData.spec?.internal || false}
            onChange={(checked) => onInputChange('spec.internal', checked)}
            type="checkbox"
            fieldName="spec.internal"
            isAnimating={isAnimating}
            isCurrentField={isFieldAnimating('spec.internal')}
            isCompleted={isFieldCompleted('spec.internal')}
            animatedValue={getFieldValue('spec.internal')}
            helpText={helpTextMapping?.['spec.internal']}
          />

          <AnimatedFormField
            ref={deleteProtectionRef}
            label="Deletion Protection"
            value={formData.spec?.deleteProtectionEnabled || false}
            onChange={(checked) => onInputChange('spec.deleteProtectionEnabled', checked)}
            type="checkbox"
            fieldName="spec.deleteProtectionEnabled"
            isAnimating={isAnimating}
            isCurrentField={isFieldAnimating('spec.deleteProtectionEnabled')}
            isCompleted={isFieldCompleted('spec.deleteProtectionEnabled')}
            animatedValue={getFieldValue('spec.deleteProtectionEnabled')}
            helpText={helpTextMapping?.['spec.deleteProtectionEnabled']}
          />
        </div>
      </div>
    );
  }

  if (section === 'dns') {
    return (
      <div className="space-y-4">
        <AnimatedFormField
          ref={dnsEnabledRef}
          label="Enable DNS"
          value={formData.spec?.dns?.enabled || false}
          onChange={(checked) => onInputChange('spec.dns.enabled', checked)}
          type="checkbox"
          fieldName="spec.dns.enabled"
          isAnimating={isAnimating}
          isCurrentField={isFieldAnimating('spec.dns.enabled')}
          isCompleted={isFieldCompleted('spec.dns.enabled')}
          animatedValue={getFieldValue('spec.dns.enabled')}
          helpText={helpTextMapping?.['spec.dns.enabled']}
        />

        {formData.spec?.dns?.enabled && (
          <>
            <AnimatedFormField
              ref={route53ZoneIdRef}
              label="Route53 Zone ID"
              value={formData.spec?.dns?.route53ZoneId?.value || ''}
              onChange={(value) => onInputChange('spec.dns.route53ZoneId.value', value)}
              type="text"
              fieldName="spec.dns.route53ZoneId"
              isAnimating={isAnimating}
              isCurrentField={isFieldAnimating('spec.dns.route53ZoneId')}
              isCompleted={isFieldCompleted('spec.dns.route53ZoneId')}
              animatedValue={getFieldValue('spec.dns.route53ZoneId')}
              helpText={helpTextMapping?.['spec.dns.route53ZoneId']}
            />

            <AnimatedFormField
              ref={hostnameRef}
              label="Hostname"
              value={formData.spec?.dns?.hostnames?.[0] || ''}
              onChange={(value) => onInputChange('spec.dns.hostnames.0', value)}
              type="text"
              fieldName="spec.dns.hostnames.0"
              isAnimating={isAnimating}
              isCurrentField={isFieldAnimating('spec.dns.hostnames.0')}
              isCompleted={isFieldCompleted('spec.dns.hostnames.0')}
              animatedValue={getFieldValue('spec.dns.hostnames.0')}
              helpText={helpTextMapping?.['spec.dns.hostnames']}
            />
          </>
        )}
      </div>
    );
  }

  if (section === 'ssl') {
    return (
      <div className="space-y-4">
        <AnimatedFormField
          ref={sslEnabledRef}
          label="Enable SSL"
          value={formData.spec?.ssl?.enabled || false}
          onChange={(checked) => onInputChange('spec.ssl.enabled', checked)}
          type="checkbox"
          fieldName="spec.ssl.enabled"
          isAnimating={isAnimating}
          isCurrentField={isFieldAnimating('spec.ssl.enabled')}
          isCompleted={isFieldCompleted('spec.ssl.enabled')}
          animatedValue={getFieldValue('spec.ssl.enabled')}
          helpText={helpTextMapping?.['spec.ssl.enabled']}
        />

        {formData.spec?.ssl?.enabled && (
          <AnimatedFormField
            ref={certificateArnRef}
            label="Certificate ARN"
            value={formData.spec?.ssl?.certificateArn?.value || ''}
            onChange={(value) => onInputChange('spec.ssl.certificateArn.value', value)}
            type="text"
            fieldName="spec.ssl.certificateArn"
            isAnimating={isAnimating}
            isCurrentField={isFieldAnimating('spec.ssl.certificateArn')}
            isCompleted={isFieldCompleted('spec.ssl.certificateArn')}
            animatedValue={getFieldValue('spec.ssl.certificateArn')}
            helpText={helpTextMapping?.['spec.ssl.certificateArn']}
          />
        )}
      </div>
    );
  }

  // Default: render all sections (for backward compatibility)
  return <div className="space-y-6">{/* All sections as before */}</div>;
};
