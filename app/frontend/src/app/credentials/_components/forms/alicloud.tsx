'use client';

import { useCallback } from 'react';
import { SimpleInput } from '@/components/shared/simple-input';
import { SimpleSelect } from '@/components/shared/simple-select';
import { CredentialFormData } from '@/app/credentials/_components/forms/types';
import { UseFormRegister, UseFormSetValue, UseFormWatch } from 'react-hook-form';
import { Typography, Divider, Stack } from '@mui/material';

interface AlicloudCredentialFormProps {
  register: UseFormRegister<CredentialFormData>;
  setValue: UseFormSetValue<CredentialFormData>;
  watch: UseFormWatch<CredentialFormData>;
  disabled?: boolean;
}

const AUTH_METHOD_OPTIONS = [
  { label: 'Static Credentials (access key pair)', value: 'static_credentials' },
  { label: 'STS Token (temporary credentials)', value: 'sts_token' },
  { label: 'ECS Instance Role', value: 'ecs_role' },
  { label: 'Assume RAM Role', value: 'assume_role' },
  { label: 'Assume Role with OIDC', value: 'assume_role_with_oidc' },
  { label: 'Shared Credentials File', value: 'shared_credentials' },
  { label: 'Sidecar Credentials', value: 'sidecar_credentials' },
];

export function AlicloudCredentialForm({
  register,
  setValue,
  watch,
  disabled,
}: AlicloudCredentialFormProps) {
  const authMethod = watch('alicloudAuthMethod') || 'static_credentials';

  const onAuthMethodChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      const method = e.target.value as NonNullable<CredentialFormData['alicloudAuthMethod']>;
      setValue('alicloudAuthMethod', method);
      setValue('alicloud', {});
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

      {authMethod === 'static_credentials' && (
        <>
          <SimpleInput
            register={register}
            path="alicloud.accessKey"
            name="Access Key ID"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="alicloud.secretKey"
            name="Access Key Secret"
            type="password"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
        </>
      )}

      {authMethod === 'sts_token' && (
        <>
          <SimpleInput
            register={register}
            path="alicloud.accessKey"
            name="Access Key ID"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="alicloud.secretKey"
            name="Access Key Secret"
            type="password"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="alicloud.securityToken"
            name="Security Token"
            type="password"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
        </>
      )}

      {authMethod === 'ecs_role' && (
        <SimpleInput
          register={register}
          path="alicloud.ecsRoleName"
          name="ECS Role Name"
          registerOptions={{ required: true }}
          disabled={disabled}
        />
      )}

      {authMethod === 'assume_role' && (
        <>
          <SimpleInput
            register={register}
            path="alicloud.accessKey"
            name="Access Key ID"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="alicloud.secretKey"
            name="Access Key Secret"
            type="password"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="alicloud.roleArn"
            name="Role ARN"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="alicloud.sessionName"
            name="Session Name"
            disabled={disabled}
          />
          <Divider sx={{ my: 1 }} />
          <Typography variant="caption" color="text.secondary" sx={{ mb: 0.5 }}>
            Advanced (optional)
          </Typography>
          <Stack spacing={1}>
            <SimpleInput
              register={register}
              path="alicloud.policy"
              name="Policy (JSON)"
              disabled={disabled}
            />
            <SimpleInput
              register={register}
              path="alicloud.externalId"
              name="External ID"
              disabled={disabled}
            />
          </Stack>
        </>
      )}

      {authMethod === 'assume_role_with_oidc' && (
        <>
          <SimpleInput
            register={register}
            path="alicloud.oidcProviderArn"
            name="OIDC Provider ARN"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="alicloud.roleArn"
            name="Role ARN"
            registerOptions={{ required: true }}
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="alicloud.oidcToken"
            name="OIDC Token"
            type="password"
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="alicloud.oidcTokenFile"
            name="OIDC Token File"
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="alicloud.sessionName"
            name="Session Name"
            disabled={disabled}
          />
        </>
      )}

      {authMethod === 'shared_credentials' && (
        <>
          <SimpleInput
            register={register}
            path="alicloud.credentialsFile"
            name="Credentials File"
            disabled={disabled}
          />
          <SimpleInput
            register={register}
            path="alicloud.profile"
            name="Profile"
            disabled={disabled}
          />
        </>
      )}

      {authMethod === 'sidecar_credentials' && (
        <SimpleInput
          register={register}
          path="alicloud.credentialsUri"
          name="Credentials URI"
          registerOptions={{ required: true }}
          disabled={disabled}
        />
      )}

      <Divider sx={{ my: 1 }} />
      <Typography variant="caption" color="text.secondary" sx={{ mb: 0.5 }}>
        Common (optional, all methods)
      </Typography>
      <Stack spacing={1}>
        <SimpleInput
          register={register}
          path="alicloud.region"
          name="Region"
          disabled={disabled}
        />
        <SimpleInput
          register={register}
          path="alicloud.accountId"
          name="Account ID"
          disabled={disabled}
        />
        <SimpleInput
          register={register}
          path="alicloud.accountType"
          name="Account Type"
          disabled={disabled}
        />
      </Stack>
    </>
  );
}
