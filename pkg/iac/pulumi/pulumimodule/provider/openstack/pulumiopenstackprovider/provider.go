package pulumiopenstackprovider

import (
	"fmt"

	"github.com/pkg/errors"
	openstackprovider "github.com/plantonhq/planton/apis/dev/planton/provider/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Get builds a pulumi-openstack Provider using the supplied credential.
// It maps the OpenStackProviderConfig oneof credentials to the appropriate
// Pulumi provider arguments. If the credential is nil, Pulumi's provider will
// fall back to OS_* environment variables.
func Get(
	ctx *pulumi.Context,
	openstackProviderConfig *openstackprovider.OpenStackProviderConfig,
	nameSuffixes ...string,
) (*openstack.Provider, error) {

	providerArgs := &openstack.ProviderArgs{}

	// Map credential fields when present; leave them nil to defer to env-vars.
	if openstackProviderConfig != nil {
		if openstackProviderConfig.AuthUrl != "" {
			providerArgs.AuthUrl = pulumi.StringPtr(openstackProviderConfig.AuthUrl)
		}
		if openstackProviderConfig.Region != "" {
			providerArgs.Region = pulumi.StringPtr(openstackProviderConfig.Region)
		}

		// Project/tenant context
		if openstackProviderConfig.TenantName != "" {
			providerArgs.TenantName = pulumi.StringPtr(openstackProviderConfig.TenantName)
		}
		if openstackProviderConfig.TenantId != "" {
			providerArgs.TenantId = pulumi.StringPtr(openstackProviderConfig.TenantId)
		}

		// Domain context
		if openstackProviderConfig.UserDomainName != "" {
			providerArgs.UserDomainName = pulumi.StringPtr(openstackProviderConfig.UserDomainName)
		}
		if openstackProviderConfig.UserDomainId != "" {
			providerArgs.UserDomainId = pulumi.StringPtr(openstackProviderConfig.UserDomainId)
		}
		if openstackProviderConfig.ProjectDomainName != "" {
			providerArgs.ProjectDomainName = pulumi.StringPtr(openstackProviderConfig.ProjectDomainName)
		}
		if openstackProviderConfig.ProjectDomainId != "" {
			providerArgs.ProjectDomainId = pulumi.StringPtr(openstackProviderConfig.ProjectDomainId)
		}

		// TLS
		if openstackProviderConfig.Insecure {
			providerArgs.Insecure = pulumi.BoolPtr(openstackProviderConfig.Insecure)
		}
		if openstackProviderConfig.CacertFile != "" {
			providerArgs.CacertFile = pulumi.StringPtr(openstackProviderConfig.CacertFile)
		}

		// Advanced
		if openstackProviderConfig.EndpointType != "" {
			providerArgs.EndpointType = pulumi.StringPtr(openstackProviderConfig.EndpointType)
		}

		// Authentication method (oneof credentials)
		switch creds := openstackProviderConfig.Credentials.(type) {
		case *openstackprovider.OpenStackProviderConfig_Password:
			if creds.Password != nil {
				if creds.Password.UserName != "" {
					providerArgs.UserName = pulumi.StringPtr(creds.Password.UserName)
				}
				if creds.Password.Password != "" {
					providerArgs.Password = pulumi.StringPtr(creds.Password.Password)
				}
			}
		case *openstackprovider.OpenStackProviderConfig_ApplicationCredential:
			if creds.ApplicationCredential != nil {
				if creds.ApplicationCredential.Id != "" {
					providerArgs.ApplicationCredentialId = pulumi.StringPtr(creds.ApplicationCredential.Id)
				}
				if creds.ApplicationCredential.Name != "" {
					providerArgs.ApplicationCredentialName = pulumi.StringPtr(creds.ApplicationCredential.Name)
				}
				if creds.ApplicationCredential.Secret != "" {
					providerArgs.ApplicationCredentialSecret = pulumi.StringPtr(creds.ApplicationCredential.Secret)
				}
			}
		case *openstackprovider.OpenStackProviderConfig_Token:
			if creds.Token != nil {
				if creds.Token.Token != "" {
					providerArgs.Token = pulumi.StringPtr(creds.Token.Token)
				}
			}
		}
	}

	provider, err := openstack.NewProvider(
		ctx,
		ProviderResourceName(nameSuffixes),
		providerArgs,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create openstack provider")
	}

	return provider, nil
}

// ProviderResourceName builds a deterministic Pulumi resource name such as
// "openstack-primary".
func ProviderResourceName(suffixes []string) string {
	name := "openstack"
	for _, s := range suffixes {
		name = fmt.Sprintf("%s-%s", name, s)
	}
	return name
}
