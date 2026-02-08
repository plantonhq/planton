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
};

