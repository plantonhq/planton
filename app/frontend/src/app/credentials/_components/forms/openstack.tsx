'use client';

import { useCallback } from 'react';
import { SimpleInput } from '@/components/shared/simple-input';
import { SimpleSelect } from '@/components/shared/simple-select';
import { CredentialFormData } from '@/app/credentials/_components/forms/types';
import { UseFormRegister, UseFormSetValue, UseFormWatch } from 'react-hook-form';
import { Typography, Divider, Stack } from '@mui/material';

interface OpenStackCredentialFormProps {
  register: UseFormRegister<CredentialFormData>;
  setValue: UseFormSetValue<CredentialFormData>;
  watch: UseFormWatch<CredentialFormData>;
  disabled?: boolean;
}

const AUTH_METHOD_OPTIONS = [
  { label: 'Application Credential (recommended)', value: 'application_credential' },
  { label: 'Password', value: 'password' },
  { label: 'Token', value: 'token' },
];

export function OpenStackCredentialForm({
  register,
  setValue,
  watch,
  disabled,
}: OpenStackCredentialFormProps) {
  const authMethod = watch('openstackAuthMethod') || 'application_credential';

  const onAuthMethodChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      const method = e.target.value as CredentialFormData['openstackAuthMethod'];
      setValue('openstackAuthMethod', method);
      // Clear auth-method-specific fields when switching
      setValue('openstack', {
        ...watch('openstack'),
        userName: undefined,
        password: undefined,
        applicationCredentialId: undefined,
        applicationCredentialName: undefined,
        applicationCredentialSecret: undefined,
        token: undefined,
      });
    },
    [setValue, watch]
  );

  return (
    <>
      {/* Connection */}
      <SimpleInput
        register={register}
        path="openstack.authUrl"
        name="Auth URL"
        registerOptions={{ required: true }}
        disabled={disabled}
      />
      <SimpleInput
        register={register}
        path="openstack.region"
        name="Region"
        disabled={disabled}
      />

      {/* Authentication Method */}
      <Divider sx={{ my: 1 }} />
      <SimpleSelect
        name="Authentication Method"
        value={authMethod}
        onChange={onAuthMethodChange}
        options={AUTH_METHOD_OPTIONS}
        disabled={disabled}
        fullWidth
      />

      {/* Application Credential fields */}
      {authMethod === 'application_credential' && (
        <>
          <SimpleInput
            register={register}
            path="openstack.applicationCredentialId"
            name="Credential ID"
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="openstack.applicationCredentialName"
            name="Credential Name"
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="openstack.applicationCredentialSecret"
            name="Credential Secret"
            type="password"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
        </>
      )}

      {/* Password fields */}
      {authMethod === 'password' && (
        <>
          <SimpleInput
            register={register}
            path="openstack.userName"
            name="Username"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="openstack.password"
            name="Password"
            type="password"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="openstack.userDomainName"
            name="User Domain Name"
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="openstack.tenantName"
            name="Project / Tenant Name"
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="openstack.projectDomainName"
            name="Project Domain Name"
            disabled={disabled}
          />
        </>
      )}

      {/* Token fields */}
      {authMethod === 'token' && (
        <>
          <SimpleInput
            register={register}
            path="openstack.token"
            name="Token"
            type="password"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="openstack.tenantName"
            name="Project / Tenant Name"
            disabled={disabled}
          />
        </>
      )}

      {/* Advanced (shown for all methods) */}
      <Divider sx={{ my: 1 }} />
      <Typography variant="caption" color="text.secondary" sx={{ mb: 0.5 }}>
        Advanced (optional)
      </Typography>
      <Stack spacing={1}>
        <SimpleInput
          register={register}
          path="openstack.endpointType"
          name="Endpoint Type"
          disabled={disabled}
        />
        <SimpleInput
          register={register}
          path="openstack.cacertFile"
          name="CA Certificate File"
          disabled={disabled}
        />
      </Stack>
    </>
  );
}
