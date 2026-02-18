'use client';

import { useCallback } from 'react';
import { Divider, Typography, Stack } from '@mui/material';
import { SimpleInput } from '@/components/shared/simple-input';
import { SimpleSelect } from '@/components/shared/simple-select';
import { CredentialFormData } from '@/app/credentials/_components/forms/types';
import { UseFormRegister, UseFormSetValue, UseFormWatch } from 'react-hook-form';

interface OciCredentialFormProps {
  register: UseFormRegister<CredentialFormData>;
  setValue: UseFormSetValue<CredentialFormData>;
  watch: UseFormWatch<CredentialFormData>;
  disabled?: boolean;
}

const AUTH_METHOD_OPTIONS = [
  { label: 'API Key (tenancy/user OCID + signing key)', value: 'api_key' },
  { label: 'Instance Principal (OCI compute)', value: 'instance_principal' },
  { label: 'Security Token (OCI CLI session)', value: 'security_token' },
  { label: 'Resource Principal (OCI Functions)', value: 'resource_principal' },
  { label: 'OKE Workload Identity (Kubernetes)', value: 'oke_workload_identity' },
];

export function OciCredentialForm({
  register,
  setValue,
  watch,
  disabled,
}: OciCredentialFormProps) {
  const authMethod = watch('ociAuthMethod') || 'api_key';

  const onAuthMethodChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      const method = e.target.value as NonNullable<CredentialFormData['ociAuthMethod']>;
      setValue('ociAuthMethod', method);
      setValue('oci', {});
    },
    [setValue]
  );

  return (
    <>
      <SimpleSelect
        name="Authentication Method"
        value={authMethod}
        onChange={onAuthMethodChange}
        options={AUTH_METHOD_OPTIONS}
        disabled={disabled}
        fullWidth
      />

      {authMethod === 'api_key' && (
        <>
          <SimpleInput
            register={register}
            path="oci.tenancyOcid"
            name="Tenancy OCID"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="oci.userOcid"
            name="User OCID"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="oci.fingerprint"
            name="Fingerprint"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="oci.privateKey"
            name="Private Key (PEM)"
            type="password"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="oci.privateKeyPassword"
            name="Private Key Password"
            type="password"
            disabled={disabled}
          />
        </>
      )}

      {authMethod === 'security_token' && (
        <>
          <SimpleInput
            register={register}
            path="oci.configFileProfile"
            name="Config File Profile"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="oci.privateKeyPassword"
            name="Private Key Password"
            type="password"
            disabled={disabled}
          />
        </>
      )}

      <Divider sx={{ my: 1 }} />
      <Typography variant="caption" color="text.secondary" sx={{ mb: 0.5 }}>
        Common (optional, all methods)
      </Typography>
      <Stack spacing={1}>
        <SimpleInput
          register={register}
          path="oci.region"
          name="Region"
          disabled={disabled}
        />
      </Stack>
    </>
  );
}
