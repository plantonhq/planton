'use client';

import { SimpleInput } from '@/components/shared/simple-input';
import { CredentialFormData } from '@/app/credentials/_components/forms/types';
import { UseFormRegister } from 'react-hook-form';

interface ScalewayCredentialFormProps {
  register: UseFormRegister<CredentialFormData>;
  disabled?: boolean;
}

export function ScalewayCredentialForm({
  register,
  disabled,
}: ScalewayCredentialFormProps) {
  return (
    <>
      {/* Authentication (required) */}
      <SimpleInput
        register={register}
        path="scaleway.accessKey"
        name="Access Key"
        registerOptions={{ required: true }}
        disabled={disabled}
      />
      <SimpleInput
        register={register}
        path="scaleway.secretKey"
        name="Secret Key"
        type="password"
        registerOptions={{ required: true }}
        disabled={disabled}
      />

      {/* Project scope (recommended) */}
      <SimpleInput
        register={register}
        path="scaleway.projectId"
        name="Project ID"
        disabled={disabled}
      />

      {/* Optional */}
      <SimpleInput
        register={register}
        path="scaleway.organizationId"
        name="Organization ID"
        disabled={disabled}
      />
      <SimpleInput
        register={register}
        path="scaleway.region"
        name="Region"
        disabled={disabled}
      />
      <SimpleInput
        register={register}
        path="scaleway.zone"
        name="Zone"
        disabled={disabled}
      />
    </>
  );
}
