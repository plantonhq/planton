import { Credential_CredentialProvider } from '@/gen/org/openmcf/app/credential/v1/api_pb';
import { ICON_NAMES } from '@/components/shared/icon';

export interface ProviderConfig {
  label: string;
  description: string;
  icon?: ICON_NAMES;
}

export const providerConfig: Record<Credential_CredentialProvider, ProviderConfig> = {
  [Credential_CredentialProvider.CREDENTIAL_PROVIDER_UNSPECIFIED]: {
    label: 'Unspecified',
    description: '',
    icon: undefined,
  },
  [Credential_CredentialProvider.GCP]: {
    label: 'Google Cloud Platform',
    description: 'Link your GCP Organization to start deploying infrastructure',
    icon: ICON_NAMES.GCP,
  },
  [Credential_CredentialProvider.AWS]: {
    label: 'Amazon Web Services',
    description: 'Link your AWS Account to start deploying infrastructure',
    icon: ICON_NAMES.AWS,
  },
  [Credential_CredentialProvider.AZURE]: {
    label: 'Microsoft Azure',
    description: 'Link your Azure Subscription to start deploying infrastructure',
    icon: ICON_NAMES.AZURE,
  },
  [Credential_CredentialProvider.AUTH0]: {
    label: 'Auth0',
    description: 'Link your Auth0 tenant to manage identity resources',
    icon: undefined,
  },
  [Credential_CredentialProvider.OPEN_FGA]: {
    label: 'OpenFGA',
    description: 'Link your OpenFGA server to manage authorization resources',
    icon: undefined,
  },
  [Credential_CredentialProvider.OPENSTACK]: {
    label: 'OpenStack',
    description: 'Link your OpenStack cloud to deploy compute, networking, and storage resources',
    icon: undefined,
  },
  [Credential_CredentialProvider.SCALEWAY]: {
    label: 'Scaleway',
    description: 'Link your Scaleway account to deploy cloud infrastructure',
    icon: undefined,
  },
  [Credential_CredentialProvider.ALICLOUD]: {
    label: 'Alibaba Cloud',
    description: 'Link your Alibaba Cloud account to deploy cloud infrastructure',
    icon: undefined,
  },
  [Credential_CredentialProvider.OCI]: {
    label: 'OCI',
    description: 'Link your Oracle Cloud Infrastructure account to deploy cloud resources',
    icon: undefined,
  },
  [Credential_CredentialProvider.HETZNER_CLOUD]: {
    label: 'Hetzner Cloud',
    description: 'Link your Hetzner Cloud account to deploy cloud infrastructure',
    icon: undefined,
  },
};

