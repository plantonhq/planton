package verify

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	pkgerrors "github.com/pkg/errors"
)

// iamOidcProviderVerifier verifies an AwsIamOidcProvider via
// GetOpenIDConnectProvider, keyed on the provider ARN (the AWS API for OIDC
// providers takes the ARN). IAM is a global service, so the region parameter
// is ignored. A deleted provider returns the typed NoSuchEntity error, which
// is the "absent" signal; any other error is a genuine failure and must
// surface.
type iamOidcProviderVerifier struct{}

func (*iamOidcProviderVerifier) IDOutputKey() string { return "provider_arn" }

func (*iamOidcProviderVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, _ string) error {
	exists, err := iamOidcProviderExists(ctx, cfg, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsiamoidcprovider verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awsiamoidcprovider %q not found after deploy", id)
	}
	return nil
}

func (*iamOidcProviderVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, _ string) error {
	exists, err := iamOidcProviderExists(ctx, cfg, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsiamoidcprovider verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awsiamoidcprovider %q still exists after destroy", id)
	}
	return nil
}

func iamOidcProviderExists(ctx context.Context, cfg aws.Config, providerArn string) (bool, error) {
	client := iam.NewFromConfig(cfg)
	_, err := client.GetOpenIDConnectProvider(ctx, &iam.GetOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(providerArn),
	})
	if err != nil {
		if isIamNoSuchEntity(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
