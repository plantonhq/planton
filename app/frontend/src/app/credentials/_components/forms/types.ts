'use client';

import { Credential_CredentialProvider } from '@/gen/org/openmcf/app/credential/v1/api_pb';
import { Auth0ProviderConfig } from '@/gen/org/openmcf/provider/auth0/provider_pb';
import { GcpProviderConfig } from '@/gen/org/openmcf/provider/gcp/provider_pb';
import { AwsProviderConfig } from '@/gen/org/openmcf/provider/aws/provider_pb';
import { AzureProviderConfig } from '@/gen/org/openmcf/provider/azure/provider_pb';

// Flattened form data for OpenStack credentials.
// The proto uses a oneof for credentials, which doesn't map cleanly to react-hook-form.
// This flat interface is converted to the proper proto structure in handleSave.
export interface OpenStackFormData {
  authUrl?: string;
  region?: string;
  // Password authentication
  userName?: string;
  password?: string;
  // Application credential authentication
  applicationCredentialId?: string;
  applicationCredentialName?: string;
  applicationCredentialSecret?: string;
  // Token authentication
  token?: string;
  // Project/tenant context
  tenantName?: string;
  tenantId?: string;
  // Domain context
  userDomainName?: string;
  userDomainId?: string;
  projectDomainName?: string;
  projectDomainId?: string;
  // TLS
  insecure?: boolean;
  cacertFile?: string;
  // Advanced
  endpointType?: string;
}

// Form data for Scaleway credentials.
// Flat structure matching ScalewayProviderConfig proto fields.
export interface ScalewayFormData {
  accessKey?: string;
  secretKey?: string;
  projectId?: string;
  organizationId?: string;
  region?: string;
  zone?: string;
}

// Flattened form data for Alibaba Cloud credentials.
// All auth method fields are combined into a single flat interface.
// The alicloudAuthMethod discriminator selects which fields are active.
export interface AlicloudFormData {
  // Common
  region?: string;
  accountId?: string;
  accountType?: string;
  // Static / STS / AssumeRole
  accessKey?: string;
  secretKey?: string;
  // STS
  securityToken?: string;
  // ECS role
  ecsRoleName?: string;
  // AssumeRole / OIDC shared
  roleArn?: string;
  sessionName?: string;
  policy?: string;
  externalId?: string;
  // OIDC
  oidcProviderArn?: string;
  oidcToken?: string;
  oidcTokenFile?: string;
  // Shared credentials
  credentialsFile?: string;
  profile?: string;
  // Sidecar
  credentialsUri?: string;
}

export type AlicloudAuthMethod =
  | 'static_credentials'
  | 'sts_token'
  | 'ecs_role'
  | 'assume_role'
  | 'assume_role_with_oidc'
  | 'shared_credentials'
  | 'sidecar_credentials';

// Flattened form data for OCI credentials.
// All auth method fields are combined into a single flat interface.
// The ociAuthMethod discriminator selects which fields are active.
export interface OciFormData {
  // Common
  region?: string;
  // API Key
  tenancyOcid?: string;
  userOcid?: string;
  fingerprint?: string;
  privateKey?: string;
  privateKeyPassword?: string;
  // Security Token
  configFileProfile?: string;
}

export type OciAuthMethod =
  | 'api_key'
  | 'instance_principal'
  | 'security_token'
  | 'resource_principal'
  | 'oke_workload_identity';

// Form-friendly type based on CreateCredentialRequest fields (without the protobuf Message wrapper)
export type CredentialFormData = {
  name: string;
  provider: Credential_CredentialProvider;
  auth0?: Partial<Auth0ProviderConfig>;
  gcp?: Partial<GcpProviderConfig>;
  aws?: Partial<AwsProviderConfig>;
  azure?: Partial<AzureProviderConfig>;
  openstack?: OpenStackFormData;
  // Auth method discriminator for the OpenStack credential form (not part of the proto)
  openstackAuthMethod?: 'password' | 'application_credential' | 'token';
  scaleway?: ScalewayFormData;
  alicloud?: AlicloudFormData;
  alicloudAuthMethod?: AlicloudAuthMethod;
  oci?: OciFormData;
  ociAuthMethod?: OciAuthMethod;
};

