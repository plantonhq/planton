package pulumiociprovider

import (
	"fmt"

	"github.com/pkg/errors"
	ociprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Get builds a pulumi-oci Provider from the supplied OciProviderConfig.
// It maps the authentication_type and associated credential fields to the
// appropriate Pulumi provider arguments. If the config is nil, Pulumi's
// provider will fall back to OCI_* environment variables.
func Get(
	ctx *pulumi.Context,
	ociProviderConfig *ociprovider.OciProviderConfig,
	nameSuffixes ...string,
) (*oci.Provider, error) {

	providerArgs := &oci.ProviderArgs{}

	if ociProviderConfig != nil {
		if ociProviderConfig.Region != "" {
			providerArgs.Region = pulumi.StringPtr(ociProviderConfig.Region)
		}

		switch ociProviderConfig.AuthenticationType {
		case ociprovider.AuthenticationType_api_key:
			providerArgs.Auth = pulumi.StringPtr("ApiKey")
			if apiKey := ociProviderConfig.GetApiKey(); apiKey != nil {
				providerArgs.TenancyOcid = pulumi.StringPtr(apiKey.TenancyOcid)
				providerArgs.UserOcid = pulumi.StringPtr(apiKey.UserOcid)
				providerArgs.Fingerprint = pulumi.StringPtr(apiKey.Fingerprint)
				if apiKey.PrivateKey != "" {
					providerArgs.PrivateKey = pulumi.StringPtr(apiKey.PrivateKey)
				}
				if apiKey.PrivateKeyPath != "" {
					providerArgs.PrivateKeyPath = pulumi.StringPtr(apiKey.PrivateKeyPath)
				}
				if apiKey.PrivateKeyPassword != "" {
					providerArgs.PrivateKeyPassword = pulumi.StringPtr(apiKey.PrivateKeyPassword)
				}
			}

		case ociprovider.AuthenticationType_instance_principal:
			providerArgs.Auth = pulumi.StringPtr("InstancePrincipal")

		case ociprovider.AuthenticationType_security_token:
			providerArgs.Auth = pulumi.StringPtr("SecurityToken")
			if secToken := ociProviderConfig.GetSecurityToken(); secToken != nil {
				providerArgs.ConfigFileProfile = pulumi.StringPtr(secToken.ConfigFileProfile)
				if secToken.PrivateKeyPassword != "" {
					providerArgs.PrivateKeyPassword = pulumi.StringPtr(secToken.PrivateKeyPassword)
				}
			}

		case ociprovider.AuthenticationType_resource_principal:
			providerArgs.Auth = pulumi.StringPtr("ResourcePrincipal")

		case ociprovider.AuthenticationType_oke_workload_identity:
			providerArgs.Auth = pulumi.StringPtr("OKEWorkloadIdentity")
		}
	}

	name := "oci-provider"
	for _, s := range nameSuffixes {
		name = fmt.Sprintf("%s-%s", name, s)
	}

	provider, err := oci.NewProvider(ctx, name, providerArgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create oci provider")
	}

	return provider, nil
}
