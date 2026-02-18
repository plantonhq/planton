'use client';

import { SimpleInput } from '@/components/shared/simple-input';
import { CredentialFormData } from '@/app/credentials/_components/forms/types';
import { UseFormRegister } from 'react-hook-form';

interface HetznercloudCredentialFormProps {
  register: UseFormRegister<CredentialFormData>;
  disabled?: boolean;
}

export function HetznercloudCredentialForm({
  register,
  disabled,
}: HetznercloudCredentialFormProps) {
  return (
    <>
      {/* Authentication (required) */}
      <SimpleInput
        register={register}
        path="hetznercloud.token"
        name="API Token"
        type="password"
        registerOptions={{ required: true }}
        disabled={disabled}
      />

      {/* Optional */}
      <SimpleInput
        register={register}
        path="hetznercloud.endpoint"
        name="Cloud API Endpoint"
        disabled={disabled}
      />
      <SimpleInput
        register={register}
        path="hetznercloud.endpointHetzner"
        name="Hetzner API Endpoint"
        disabled={disabled}
      />
      <SimpleInput
        register={register}
        path="hetznercloud.pollInterval"
        name="Poll Interval"
        disabled={disabled}
      />
      <SimpleInput
        register={register}
        path="hetznercloud.pollFunction"
        name="Poll Function"
        disabled={disabled}
      />
    </>
  );
}
