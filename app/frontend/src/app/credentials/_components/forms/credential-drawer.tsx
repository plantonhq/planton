'use client';

import { useEffect, useCallback, useMemo } from 'react';
import { useForm, useWatch } from 'react-hook-form';
import { Drawer } from '@/components/shared/drawer';
import { Stack, Button } from '@mui/material';
import {
  DrawerContainer,
  DrawerContentArea,
  DrawerFooter,
} from '@/app/credentials/_components/styled';
import { SimpleInput } from '@/components/shared/simple-input';
import { SimpleSelect } from '@/components/shared/simple-select';
import {
  CredentialFormData,
  Auth0CredentialForm,
  GcpCredentialForm,
  AwsCredentialForm,
  AzureCredentialForm,
  OpenStackCredentialForm,
  ScalewayCredentialForm,
  AlicloudCredentialForm,
  OciCredentialForm,
} from '@/app/credentials/_components/forms';
import { useCredentialCommand } from '@/app/credentials/_services';
import {
  Credential_CredentialProvider,
  Credential,
  CredentialProviderConfigSchema,
} from '@/gen/org/openmcf/app/credential/v1/api_pb';
import { CreateCredentialRequest } from '@/gen/org/openmcf/app/credential/v1/io_pb';
import { Auth0ProviderConfig, Auth0ProviderConfigSchema } from '@/gen/org/openmcf/provider/auth0/provider_pb';
import { GcpProviderConfig, GcpProviderConfigSchema } from '@/gen/org/openmcf/provider/gcp/provider_pb';
import { AwsProviderConfig, AwsProviderConfigSchema } from '@/gen/org/openmcf/provider/aws/provider_pb';
import { AzureProviderConfig, AzureProviderConfigSchema } from '@/gen/org/openmcf/provider/azure/provider_pb';
import {
  OpenStackProviderConfigSchema,
  OpenStackPasswordCredentialsSchema,
  OpenStackApplicationCredentialsSchema,
  OpenStackTokenCredentialsSchema,
} from '@/gen/org/openmcf/provider/openstack/provider_pb';
import {
  ScalewayProviderConfigSchema,
} from '@/gen/org/openmcf/provider/scaleway/provider_pb';
import {
  AliCloudProviderConfigSchema,
  AlicloudStaticCredentialsSchema,
  AlicloudStsTokenCredentialsSchema,
  AlicloudEcsRoleCredentialsSchema,
  AlicloudAssumeRoleCredentialsSchema,
  AlicloudAssumeRoleWithOidcCredentialsSchema,
  AlicloudSharedCredentialsSchema,
  AlicloudSidecarCredentialsSchema,
  AuthenticationType,
} from '@/gen/org/openmcf/provider/alicloud/provider_pb';
import {
  OciProviderConfigSchema,
  OciApiKeyAuthSchema,
  OciSecurityTokenAuthSchema,
  AuthenticationType as OciAuthenticationType,
} from '@/gen/org/openmcf/provider/oci/provider_pb';
import type { OpenStackFormData, ScalewayFormData, AlicloudFormData, AlicloudAuthMethod, OciFormData, OciAuthMethod } from '@/app/credentials/_components/forms/types';
import { create } from '@bufbuild/protobuf';
import { providerConfig } from '@/app/credentials/_components/utils';

export type DrawerMode = 'view' | 'edit' | 'create' | null;

interface CredentialDrawerProps {
  open: boolean;
  mode: DrawerMode;
  onClose: () => void;
  onSaveSuccess: () => void;
  selectedCredential?: Credential | null;
  initialProvider?: Credential_CredentialProvider;
}

export function CredentialDrawer({
  open,
  mode,
  onClose,
  onSaveSuccess,
  selectedCredential,
  initialProvider,
}: CredentialDrawerProps) {
  const { command } = useCredentialCommand();
  const isView = mode === 'view';
  const submitLabel = mode === 'edit' ? 'Update' : 'Create';

  const { register, handleSubmit, reset, setValue, control, watch } = useForm<CredentialFormData>({
    defaultValues: {
      name: '',
      provider: Credential_CredentialProvider.CREDENTIAL_PROVIDER_UNSPECIFIED,
      auth0: {},
      gcp: {},
      aws: {},
      azure: {},
      openstack: {},
      openstackAuthMethod: 'application_credential',
      scaleway: {},
      alicloud: {},
      alicloudAuthMethod: 'static_credentials',
      oci: {},
      ociAuthMethod: 'api_key',
    },
  });

  useEffect(() => {
    if (initialProvider) {
      setValue('provider', initialProvider);
    }
  }, [initialProvider, setValue]);

  const formProvider = useWatch({ control, name: 'provider' });

  const providerOptions = useMemo(() => {
    return (Object.keys(providerConfig) as unknown as Array<Credential_CredentialProvider>)
      .filter((provider) => {
        // Filter out UNSPECIFIED (value 0) by comparing numeric enum values
        return Number(provider) !== Credential_CredentialProvider.CREDENTIAL_PROVIDER_UNSPECIFIED;
      })
      .map((provider) => ({
        label: providerConfig[provider].label,
        value: provider,
      }));
  }, []);

  // Populate form when selectedCredential changes
  useEffect(() => {
    if (selectedCredential && (mode === 'view' || mode === 'edit')) {
      const providerConfigData = selectedCredential.providerConfig;
      const formData: CredentialFormData = {
        name: selectedCredential.name,
        provider: selectedCredential.provider,
        auth0: {},
        gcp: {},
        aws: {},
        azure: {},
        openstack: {},
        openstackAuthMethod: 'application_credential',
        scaleway: {},
        alicloud: {},
        alicloudAuthMethod: 'static_credentials',
        oci: {},
        ociAuthMethod: 'api_key',
      };
      if (providerConfigData?.data?.case === 'auth0') {
        formData.auth0 = {
          domain: providerConfigData.data.value.domain,
          clientId: providerConfigData.data.value.clientId,
          clientSecret: providerConfigData.data.value.clientSecret,
        };
      } else if (providerConfigData?.data?.case === 'gcp') {
        formData.gcp = {
          serviceAccountKeyBase64: providerConfigData.data.value.serviceAccountKeyBase64,
        };
      } else if (providerConfigData?.data?.case === 'aws') {
        formData.aws = {
          accountId: providerConfigData.data.value.accountId,
          accessKeyId: providerConfigData.data.value.accessKeyId,
          secretAccessKey: providerConfigData.data.value.secretAccessKey,
          region: providerConfigData.data.value.region,
          sessionToken: providerConfigData.data.value.sessionToken,
        };
      } else if (providerConfigData?.data?.case === 'azure') {
        formData.azure = {
          clientId: providerConfigData.data.value.clientId,
          clientSecret: providerConfigData.data.value.clientSecret,
          tenantId: providerConfigData.data.value.tenantId,
          subscriptionId: providerConfigData.data.value.subscriptionId,
        };
      } else if (providerConfigData?.data?.case === 'openstack') {
        const os = providerConfigData.data.value;
        const osData: OpenStackFormData = {
          authUrl: os.authUrl,
          region: os.region,
          tenantName: os.tenantName,
          tenantId: os.tenantId,
          userDomainName: os.userDomainName,
          userDomainId: os.userDomainId,
          projectDomainName: os.projectDomainName,
          projectDomainId: os.projectDomainId,
          insecure: os.insecure,
          cacertFile: os.cacertFile,
          endpointType: os.endpointType,
        };
        // Determine auth method from the credentials oneof
        if (os.credentials?.case === 'password') {
          formData.openstackAuthMethod = 'password';
          osData.userName = os.credentials.value.userName;
          osData.password = os.credentials.value.password;
        } else if (os.credentials?.case === 'applicationCredential') {
          formData.openstackAuthMethod = 'application_credential';
          osData.applicationCredentialId = os.credentials.value.id;
          osData.applicationCredentialName = os.credentials.value.name;
          osData.applicationCredentialSecret = os.credentials.value.secret;
        } else if (os.credentials?.case === 'token') {
          formData.openstackAuthMethod = 'token';
          osData.token = os.credentials.value.token;
        }
        formData.openstack = osData;
      } else if (providerConfigData?.data?.case === 'scaleway') {
        const scw = providerConfigData.data.value;
        formData.scaleway = {
          accessKey: scw.accessKey,
          secretKey: scw.secretKey,
          projectId: scw.projectId,
          organizationId: scw.organizationId,
          region: scw.region,
          zone: scw.zone,
        };
      } else if (providerConfigData?.data?.case === 'alicloud') {
        const ali = providerConfigData.data.value;
        const aliData: AlicloudFormData = {
          region: ali.region,
          accountId: ali.accountId,
          accountType: ali.accountType,
        };
        const authTypeMap: Record<number, AlicloudAuthMethod> = {
          [AuthenticationType.static_credentials]: 'static_credentials',
          [AuthenticationType.sts_token]: 'sts_token',
          [AuthenticationType.ecs_role]: 'ecs_role',
          [AuthenticationType.assume_role]: 'assume_role',
          [AuthenticationType.assume_role_with_oidc]: 'assume_role_with_oidc',
          [AuthenticationType.shared_credentials]: 'shared_credentials',
          [AuthenticationType.sidecar_credentials]: 'sidecar_credentials',
        };
        formData.alicloudAuthMethod = authTypeMap[ali.authenticationType] || 'static_credentials';
        if (ali.staticCredentials) {
          aliData.accessKey = ali.staticCredentials.accessKey;
          aliData.secretKey = ali.staticCredentials.secretKey;
        }
        if (ali.stsToken) {
          aliData.accessKey = ali.stsToken.accessKey;
          aliData.secretKey = ali.stsToken.secretKey;
          aliData.securityToken = ali.stsToken.securityToken;
        }
        if (ali.ecsRole) {
          aliData.ecsRoleName = ali.ecsRole.ecsRoleName;
        }
        if (ali.assumeRole) {
          aliData.accessKey = ali.assumeRole.accessKey;
          aliData.secretKey = ali.assumeRole.secretKey;
          aliData.roleArn = ali.assumeRole.roleArn;
          aliData.sessionName = ali.assumeRole.sessionName;
          aliData.policy = ali.assumeRole.policy;
          aliData.externalId = ali.assumeRole.externalId;
        }
        if (ali.assumeRoleWithOidc) {
          aliData.oidcProviderArn = ali.assumeRoleWithOidc.oidcProviderArn;
          aliData.roleArn = ali.assumeRoleWithOidc.roleArn;
          aliData.oidcToken = ali.assumeRoleWithOidc.oidcToken;
          aliData.oidcTokenFile = ali.assumeRoleWithOidc.oidcTokenFile;
          aliData.sessionName = ali.assumeRoleWithOidc.sessionName;
          aliData.policy = ali.assumeRoleWithOidc.policy;
        }
        if (ali.sharedCredentials) {
          aliData.credentialsFile = ali.sharedCredentials.credentialsFile;
          aliData.profile = ali.sharedCredentials.profile;
        }
        if (ali.sidecarCredentials) {
          aliData.credentialsUri = ali.sidecarCredentials.credentialsUri;
        }
        formData.alicloud = aliData;
      } else if (providerConfigData?.data?.case === 'oci') {
        const ociData: OciFormData = {};
        const ociConfig = providerConfigData.data.value;
        ociData.region = ociConfig.region;

        const ociAuthTypeMap: Record<number, OciAuthMethod> = {
          [OciAuthenticationType.api_key]: 'api_key',
          [OciAuthenticationType.instance_principal]: 'instance_principal',
          [OciAuthenticationType.security_token]: 'security_token',
          [OciAuthenticationType.resource_principal]: 'resource_principal',
          [OciAuthenticationType.oke_workload_identity]: 'oke_workload_identity',
        };
        formData.ociAuthMethod = ociAuthTypeMap[ociConfig.authenticationType] || 'api_key';

        if (ociConfig.apiKey) {
          ociData.tenancyOcid = ociConfig.apiKey.tenancyOcid;
          ociData.userOcid = ociConfig.apiKey.userOcid;
          ociData.fingerprint = ociConfig.apiKey.fingerprint;
          ociData.privateKey = ociConfig.apiKey.privateKey;
          ociData.privateKeyPassword = ociConfig.apiKey.privateKeyPassword;
        }
        if (ociConfig.securityToken) {
          ociData.configFileProfile = ociConfig.securityToken.configFileProfile;
          ociData.privateKeyPassword = ociConfig.securityToken.privateKeyPassword;
        }
        formData.oci = ociData;
      }
      reset(formData);
    } else if (mode === 'create') {
      reset({
        name: '',
        provider: initialProvider || Credential_CredentialProvider.CREDENTIAL_PROVIDER_UNSPECIFIED,
        auth0: {},
        gcp: {},
        aws: {},
        azure: {},
        openstack: {},
        openstackAuthMethod: 'application_credential',
        scaleway: {},
        alicloud: {},
        alicloudAuthMethod: 'static_credentials',
        oci: {},
        ociAuthMethod: 'api_key',
      });
    }
  }, [selectedCredential, mode, initialProvider, reset]);

  const handleSave = useCallback(
    (formData: CredentialFormData) => {
      if (!command) return;

      let providerConfig: CreateCredentialRequest['providerConfig'];

      if (
        formData.provider == Credential_CredentialProvider.AUTH0 &&
        formData.auth0?.domain &&
        formData.auth0?.clientId &&
        formData.auth0?.clientSecret
      ) {
        providerConfig = create(CredentialProviderConfigSchema, {
          data: {
            case: 'auth0',
            value: create(Auth0ProviderConfigSchema, formData.auth0 as Auth0ProviderConfig),
          },
        });
      } else if (
        formData.provider == Credential_CredentialProvider.GCP &&
        formData.gcp?.serviceAccountKeyBase64
      ) {
        providerConfig = create(CredentialProviderConfigSchema, {
          data: {
            case: 'gcp',
            value: create(GcpProviderConfigSchema, formData.gcp as GcpProviderConfig),
          },
        });
      } else if (
        formData.provider == Credential_CredentialProvider.AWS &&
        formData.aws?.accountId &&
        formData.aws?.accessKeyId &&
        formData.aws?.secretAccessKey
      ) {
        providerConfig = create(CredentialProviderConfigSchema, {
          data: {
            case: 'aws',
            value: create(AwsProviderConfigSchema, formData.aws as AwsProviderConfig),
          },
        });
      } else if (
        formData.provider == Credential_CredentialProvider.AZURE &&
        formData.azure?.clientId &&
        formData.azure?.clientSecret &&
        formData.azure?.tenantId &&
        formData.azure?.subscriptionId
      ) {
        providerConfig = create(CredentialProviderConfigSchema, {
          data: {
            case: 'azure',
            value: create(AzureProviderConfigSchema, formData.azure as AzureProviderConfig),
          },
        });
      } else if (
        formData.provider == Credential_CredentialProvider.OPENSTACK &&
        formData.openstack?.authUrl
      ) {
        // Build the credentials oneof based on the selected auth method
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        let credentials: any = {};
        const method = formData.openstackAuthMethod || 'application_credential';
        if (method === 'password' && formData.openstack.userName && formData.openstack.password) {
          credentials = {
            case: 'password' as const,
            value: create(OpenStackPasswordCredentialsSchema, {
              userName: formData.openstack.userName,
              password: formData.openstack.password,
            }),
          };
        } else if (method === 'application_credential' && formData.openstack.applicationCredentialSecret) {
          credentials = {
            case: 'applicationCredential' as const,
            value: create(OpenStackApplicationCredentialsSchema, {
              id: formData.openstack.applicationCredentialId || '',
              name: formData.openstack.applicationCredentialName || '',
              secret: formData.openstack.applicationCredentialSecret,
            }),
          };
        } else if (method === 'token' && formData.openstack.token) {
          credentials = {
            case: 'token' as const,
            value: create(OpenStackTokenCredentialsSchema, {
              token: formData.openstack.token,
            }),
          };
        } else {
          return; // Required credential fields missing
        }

        providerConfig = create(CredentialProviderConfigSchema, {
          data: {
            case: 'openstack',
            value: create(OpenStackProviderConfigSchema, {
              authUrl: formData.openstack.authUrl,
              region: formData.openstack.region || '',
              credentials,
              tenantName: formData.openstack.tenantName || '',
              tenantId: formData.openstack.tenantId || '',
              userDomainName: formData.openstack.userDomainName || '',
              userDomainId: formData.openstack.userDomainId || '',
              projectDomainName: formData.openstack.projectDomainName || '',
              projectDomainId: formData.openstack.projectDomainId || '',
              insecure: formData.openstack.insecure || false,
              cacertFile: formData.openstack.cacertFile || '',
              endpointType: formData.openstack.endpointType || '',
            }),
          },
        });
      } else if (
        formData.provider == Credential_CredentialProvider.SCALEWAY &&
        formData.scaleway?.accessKey &&
        formData.scaleway?.secretKey
      ) {
        providerConfig = create(CredentialProviderConfigSchema, {
          data: {
            case: 'scaleway',
            value: create(ScalewayProviderConfigSchema, {
              accessKey: formData.scaleway.accessKey,
              secretKey: formData.scaleway.secretKey,
              projectId: formData.scaleway.projectId || '',
              organizationId: formData.scaleway.organizationId || '',
              region: formData.scaleway.region || '',
              zone: formData.scaleway.zone || '',
            }),
          },
        });
      } else if (
        formData.provider == Credential_CredentialProvider.ALICLOUD &&
        formData.alicloud
      ) {
        const method = formData.alicloudAuthMethod || 'static_credentials';
        const ali = formData.alicloud;

        const authTypeEnumMap: Record<string, AuthenticationType> = {
          static_credentials: AuthenticationType.static_credentials,
          sts_token: AuthenticationType.sts_token,
          ecs_role: AuthenticationType.ecs_role,
          assume_role: AuthenticationType.assume_role,
          assume_role_with_oidc: AuthenticationType.assume_role_with_oidc,
          shared_credentials: AuthenticationType.shared_credentials,
          sidecar_credentials: AuthenticationType.sidecar_credentials,
        };

        // Build method-specific sub-message
        const configFields: Record<string, unknown> = {
          authenticationType: authTypeEnumMap[method],
          region: ali.region || '',
          accountId: ali.accountId || '',
          accountType: ali.accountType || '',
        };

        if (method === 'static_credentials' && ali.accessKey && ali.secretKey) {
          configFields.staticCredentials = create(AlicloudStaticCredentialsSchema, {
            accessKey: ali.accessKey,
            secretKey: ali.secretKey,
          });
        } else if (method === 'sts_token' && ali.accessKey && ali.secretKey && ali.securityToken) {
          configFields.stsToken = create(AlicloudStsTokenCredentialsSchema, {
            accessKey: ali.accessKey,
            secretKey: ali.secretKey,
            securityToken: ali.securityToken,
          });
        } else if (method === 'ecs_role' && ali.ecsRoleName) {
          configFields.ecsRole = create(AlicloudEcsRoleCredentialsSchema, {
            ecsRoleName: ali.ecsRoleName,
          });
        } else if (method === 'assume_role' && ali.accessKey && ali.secretKey && ali.roleArn) {
          configFields.assumeRole = create(AlicloudAssumeRoleCredentialsSchema, {
            accessKey: ali.accessKey,
            secretKey: ali.secretKey,
            roleArn: ali.roleArn,
            sessionName: ali.sessionName || '',
            policy: ali.policy || '',
            externalId: ali.externalId || '',
          });
        } else if (method === 'assume_role_with_oidc' && ali.oidcProviderArn && ali.roleArn) {
          configFields.assumeRoleWithOidc = create(AlicloudAssumeRoleWithOidcCredentialsSchema, {
            oidcProviderArn: ali.oidcProviderArn,
            roleArn: ali.roleArn,
            oidcToken: ali.oidcToken || '',
            oidcTokenFile: ali.oidcTokenFile || '',
            sessionName: ali.sessionName || '',
            policy: ali.policy || '',
          });
        } else if (method === 'shared_credentials') {
          configFields.sharedCredentials = create(AlicloudSharedCredentialsSchema, {
            credentialsFile: ali.credentialsFile || '',
            profile: ali.profile || '',
          });
        } else if (method === 'sidecar_credentials' && ali.credentialsUri) {
          configFields.sidecarCredentials = create(AlicloudSidecarCredentialsSchema, {
            credentialsUri: ali.credentialsUri,
          });
        } else {
          return;
        }

        providerConfig = create(CredentialProviderConfigSchema, {
          data: {
            case: 'alicloud',
            value: create(AliCloudProviderConfigSchema, configFields),
          },
        });
      } else if (
        formData.provider == Credential_CredentialProvider.OCI &&
        formData.oci
      ) {
        const method = formData.ociAuthMethod || 'api_key';
        const ociFormData = formData.oci;

        const ociAuthTypeEnumMap: Record<string, OciAuthenticationType> = {
          api_key: OciAuthenticationType.api_key,
          instance_principal: OciAuthenticationType.instance_principal,
          security_token: OciAuthenticationType.security_token,
          resource_principal: OciAuthenticationType.resource_principal,
          oke_workload_identity: OciAuthenticationType.oke_workload_identity,
        };

        const configFields: Record<string, unknown> = {
          authenticationType: ociAuthTypeEnumMap[method],
          region: ociFormData.region || '',
        };

        if (method === 'api_key' && ociFormData.tenancyOcid && ociFormData.userOcid && ociFormData.fingerprint && ociFormData.privateKey) {
          configFields.apiKey = create(OciApiKeyAuthSchema, {
            tenancyOcid: ociFormData.tenancyOcid,
            userOcid: ociFormData.userOcid,
            fingerprint: ociFormData.fingerprint,
            privateKey: ociFormData.privateKey,
            privateKeyPassword: ociFormData.privateKeyPassword || '',
          });
        } else if (method === 'security_token' && ociFormData.configFileProfile) {
          configFields.securityToken = create(OciSecurityTokenAuthSchema, {
            configFileProfile: ociFormData.configFileProfile,
            privateKeyPassword: ociFormData.privateKeyPassword || '',
          });
        } else if (method === 'instance_principal' || method === 'resource_principal' || method === 'oke_workload_identity') {
          // Ambient methods have no additional fields
        } else {
          return;
        }

        providerConfig = create(CredentialProviderConfigSchema, {
          data: {
            case: 'oci',
            value: create(OciProviderConfigSchema, configFields),
          },
        });
      } else {
        return;
      }

      if (mode === 'create') {
        command.create(formData.name, formData.provider, providerConfig).then(() => {
          onSaveSuccess();
        });
      } else if (mode === 'edit' && selectedCredential) {
        command
          .update(selectedCredential.id, formData.name, formData.provider, providerConfig)
          .then(() => {
            onSaveSuccess();
          });
      }
    },
    [command, mode, selectedCredential, onSaveSuccess]
  );

  const handleClose = () => {
    reset({
      name: '',
      provider: initialProvider || Credential_CredentialProvider.CREDENTIAL_PROVIDER_UNSPECIFIED,
      auth0: {},
      gcp: {},
      aws: {},
      azure: {},
      openstack: {},
      openstackAuthMethod: 'application_credential',
      scaleway: {},
      alicloud: {},
      alicloudAuthMethod: 'static_credentials',
      oci: {},
      ociAuthMethod: 'api_key',
    });
    onClose();
  };

  const onProviderChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      if (isView || initialProvider) return;
      const newProvider = parseInt(e.target.value, 10) as Credential_CredentialProvider;
      setValue('provider', newProvider);
      setValue('auth0', {});
      setValue('gcp', {});
      setValue('aws', {});
      setValue('azure', {});
      setValue('openstack', {});
      setValue('openstackAuthMethod', 'application_credential');
      setValue('scaleway', {});
      setValue('alicloud', {});
      setValue('alicloudAuthMethod', 'static_credentials');
      setValue('oci', {});
      setValue('ociAuthMethod', 'api_key');
    },
    [setValue, isView, initialProvider]
  );

  const title =
    mode === 'view'
      ? 'View Credential'
      : mode === 'edit'
        ? 'Edit Credential'
        : initialProvider
          ? `Create ${providerConfig[initialProvider].label} Credential`
          : 'Create Credential';

  return (
    <Drawer open={open} onClose={handleClose} title={title} width={600}>
      <DrawerContainer>
        <DrawerContentArea $hasFooter={!isView}>
          <Stack spacing={3}>
            <Stack>
              <SimpleSelect
                name="Provider"
                value={formProvider}
                required
                disabled={isView || !!initialProvider}
                onChange={onProviderChange}
                options={providerOptions}
                sx={{ minWidth: 250 }}
              />
              {!!formProvider && (
                <SimpleInput
                  register={register}
                  path="name"
                  name="Name"
                  registerOptions={{ required: true }}
                  disabled={isView}
                />
              )}
              {formProvider == Credential_CredentialProvider.AUTH0 && (
                <Auth0CredentialForm register={register} disabled={isView} />
              )}
              {formProvider == Credential_CredentialProvider.GCP && (
                <GcpCredentialForm
                  setValue={setValue}
                  watch={watch}
                  disabled={isView}
                  credentialName={selectedCredential?.name}
                />
              )}
              {formProvider == Credential_CredentialProvider.AWS && (
                <AwsCredentialForm register={register} disabled={isView} />
              )}
              {formProvider == Credential_CredentialProvider.AZURE && (
                <AzureCredentialForm register={register} disabled={isView} />
              )}
              {formProvider == Credential_CredentialProvider.OPENSTACK && (
                <OpenStackCredentialForm
                  register={register}
                  setValue={setValue}
                  watch={watch}
                  disabled={isView}
                />
              )}
              {formProvider == Credential_CredentialProvider.SCALEWAY && (
                <ScalewayCredentialForm register={register} disabled={isView} />
              )}
              {formProvider == Credential_CredentialProvider.ALICLOUD && (
                <AlicloudCredentialForm
                  register={register}
                  setValue={setValue}
                  watch={watch}
                  disabled={isView}
                />
              )}
              {formProvider == Credential_CredentialProvider.OCI && (
                <OciCredentialForm
                  register={register}
                  setValue={setValue}
                  watch={watch}
                  disabled={isView}
                />
              )}
            </Stack>
          </Stack>
        </DrawerContentArea>
        {!isView && (
          <DrawerFooter>
            <Button variant="contained" color="secondary" onClick={handleClose}>
              Cancel
            </Button>
            <Button variant="contained" color="primary" onClick={handleSubmit(handleSave)}>
              {submitLabel}
            </Button>
          </DrawerFooter>
        )}
      </DrawerContainer>
    </Drawer>
  );
}
