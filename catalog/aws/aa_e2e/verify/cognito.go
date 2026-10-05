package verify

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	pkgerrors "github.com/pkg/errors"
)

// The Cognito family verifies through the cognito-idp control plane. The user
// pool is keyed by its own id; the three pool-scoped kinds (app client,
// identity provider, resource server) have COMPOSITE AWS identities -- (pool
// id, X) -- and no ARN, so their verifiers read the pool id from the
// outputs (every kind echoes its resolved user_pool_id) via the
// OutputsVerifier path. Cognito deletes are synchronous: gone means gone, no
// DELETING states to special-case.

// cognitoUserPoolVerifier verifies an AwsCognitoUserPool via DescribeUserPool.
type cognitoUserPoolVerifier struct{}

func (*cognitoUserPoolVerifier) IDOutputKey() string { return "user_pool_id" }

func (*cognitoUserPoolVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := cognitoUserPoolExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awscognitouserpool verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awscognitouserpool %q not found after deploy", id)
	}
	return nil
}

func (*cognitoUserPoolVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := cognitoUserPoolExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awscognitouserpool verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awscognitouserpool %q still exists after destroy", id)
	}
	return nil
}

func cognitoUserPoolExists(ctx context.Context, cfg aws.Config, poolId, region string) (bool, error) {
	client := cognitoIdpClient(cfg, region)
	_, err := client.DescribeUserPool(ctx, &cognitoidentityprovider.DescribeUserPoolInput{
		UserPoolId: aws.String(poolId),
	})
	return cognitoFound(err)
}

// cognitoUserPoolClientVerifier verifies an AwsCognitoUserPoolClient via
// DescribeUserPoolClient, keyed by the (user_pool_id, client_id) output pair.
type cognitoUserPoolClientVerifier struct{}

func (*cognitoUserPoolClientVerifier) IDOutputKey() string { return "client_id" }

func (*cognitoUserPoolClientVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	return pkgerrors.New("awscognitouserpoolclient verify-exists requires full outputs (user_pool_id); use OutputsVerifier path")
}

func (*cognitoUserPoolClientVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	return pkgerrors.New("awscognitouserpoolclient verify-absent requires full outputs (user_pool_id); use OutputsVerifier path")
}

func (*cognitoUserPoolClientVerifier) VerifyExistsFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	poolId, clientId, err := cognitoPoolScopedLookup(outputs, "client_id")
	if err != nil {
		return pkgerrors.Wrap(err, "awscognitouserpoolclient verify-exists")
	}
	exists, err := cognitoUserPoolClientExists(ctx, cfg, poolId, clientId, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awscognitouserpoolclient verify-exists failed for pool %q client %q", poolId, clientId)
	}
	if !exists {
		return pkgerrors.Errorf("awscognitouserpoolclient %q on pool %q not found after deploy", clientId, poolId)
	}
	return nil
}

func (*cognitoUserPoolClientVerifier) VerifyAbsentFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	poolId, clientId, err := cognitoPoolScopedLookup(outputs, "client_id")
	if err != nil {
		return pkgerrors.Wrap(err, "awscognitouserpoolclient verify-absent")
	}
	// Destroying the client's POOL also removes the client; either absence is
	// a clean destroy.
	poolExists, err := cognitoUserPoolExists(ctx, cfg, poolId, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awscognitouserpoolclient verify-absent failed checking pool %q", poolId)
	}
	if !poolExists {
		return nil
	}
	exists, err := cognitoUserPoolClientExists(ctx, cfg, poolId, clientId, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awscognitouserpoolclient verify-absent failed for pool %q client %q", poolId, clientId)
	}
	if exists {
		return pkgerrors.Errorf("awscognitouserpoolclient %q on pool %q still exists after destroy", clientId, poolId)
	}
	return nil
}

func cognitoUserPoolClientExists(ctx context.Context, cfg aws.Config, poolId, clientId, region string) (bool, error) {
	client := cognitoIdpClient(cfg, region)
	_, err := client.DescribeUserPoolClient(ctx, &cognitoidentityprovider.DescribeUserPoolClientInput{
		UserPoolId: aws.String(poolId),
		ClientId:   aws.String(clientId),
	})
	return cognitoFound(err)
}

// cognitoIdentityProviderVerifier verifies an AwsCognitoIdentityProvider via
// DescribeIdentityProvider, keyed by the (user_pool_id, provider_name) pair.
type cognitoIdentityProviderVerifier struct{}

func (*cognitoIdentityProviderVerifier) IDOutputKey() string { return "provider_name" }

func (*cognitoIdentityProviderVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	return pkgerrors.New("awscognitoidentityprovider verify-exists requires full outputs (user_pool_id); use OutputsVerifier path")
}

func (*cognitoIdentityProviderVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	return pkgerrors.New("awscognitoidentityprovider verify-absent requires full outputs (user_pool_id); use OutputsVerifier path")
}

func (*cognitoIdentityProviderVerifier) VerifyExistsFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	poolId, providerName, err := cognitoPoolScopedLookup(outputs, "provider_name")
	if err != nil {
		return pkgerrors.Wrap(err, "awscognitoidentityprovider verify-exists")
	}
	exists, err := cognitoIdentityProviderExists(ctx, cfg, poolId, providerName, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awscognitoidentityprovider verify-exists failed for pool %q provider %q", poolId, providerName)
	}
	if !exists {
		return pkgerrors.Errorf("awscognitoidentityprovider %q on pool %q not found after deploy", providerName, poolId)
	}
	return nil
}

func (*cognitoIdentityProviderVerifier) VerifyAbsentFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	poolId, providerName, err := cognitoPoolScopedLookup(outputs, "provider_name")
	if err != nil {
		return pkgerrors.Wrap(err, "awscognitoidentityprovider verify-absent")
	}
	poolExists, err := cognitoUserPoolExists(ctx, cfg, poolId, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awscognitoidentityprovider verify-absent failed checking pool %q", poolId)
	}
	if !poolExists {
		return nil
	}
	exists, err := cognitoIdentityProviderExists(ctx, cfg, poolId, providerName, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awscognitoidentityprovider verify-absent failed for pool %q provider %q", poolId, providerName)
	}
	if exists {
		return pkgerrors.Errorf("awscognitoidentityprovider %q on pool %q still exists after destroy", providerName, poolId)
	}
	return nil
}

func cognitoIdentityProviderExists(ctx context.Context, cfg aws.Config, poolId, providerName, region string) (bool, error) {
	client := cognitoIdpClient(cfg, region)
	_, err := client.DescribeIdentityProvider(ctx, &cognitoidentityprovider.DescribeIdentityProviderInput{
		UserPoolId:   aws.String(poolId),
		ProviderName: aws.String(providerName),
	})
	return cognitoFound(err)
}

// cognitoResourceServerVerifier verifies an AwsCognitoResourceServer via
// DescribeResourceServer, keyed by the (user_pool_id, identifier) pair.
type cognitoResourceServerVerifier struct{}

func (*cognitoResourceServerVerifier) IDOutputKey() string { return "resource_server_identifier" }

func (*cognitoResourceServerVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	return pkgerrors.New("awscognitoresourceserver verify-exists requires full outputs (user_pool_id); use OutputsVerifier path")
}

func (*cognitoResourceServerVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	return pkgerrors.New("awscognitoresourceserver verify-absent requires full outputs (user_pool_id); use OutputsVerifier path")
}

func (*cognitoResourceServerVerifier) VerifyExistsFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	poolId, identifier, err := cognitoPoolScopedLookup(outputs, "resource_server_identifier")
	if err != nil {
		return pkgerrors.Wrap(err, "awscognitoresourceserver verify-exists")
	}
	exists, err := cognitoResourceServerExists(ctx, cfg, poolId, identifier, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awscognitoresourceserver verify-exists failed for pool %q identifier %q", poolId, identifier)
	}
	if !exists {
		return pkgerrors.Errorf("awscognitoresourceserver %q on pool %q not found after deploy", identifier, poolId)
	}
	return nil
}

func (*cognitoResourceServerVerifier) VerifyAbsentFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	poolId, identifier, err := cognitoPoolScopedLookup(outputs, "resource_server_identifier")
	if err != nil {
		return pkgerrors.Wrap(err, "awscognitoresourceserver verify-absent")
	}
	poolExists, err := cognitoUserPoolExists(ctx, cfg, poolId, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awscognitoresourceserver verify-absent failed checking pool %q", poolId)
	}
	if !poolExists {
		return nil
	}
	exists, err := cognitoResourceServerExists(ctx, cfg, poolId, identifier, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awscognitoresourceserver verify-absent failed for pool %q identifier %q", poolId, identifier)
	}
	if exists {
		return pkgerrors.Errorf("awscognitoresourceserver %q on pool %q still exists after destroy", identifier, poolId)
	}
	return nil
}

func cognitoResourceServerExists(ctx context.Context, cfg aws.Config, poolId, identifier, region string) (bool, error) {
	client := cognitoIdpClient(cfg, region)
	_, err := client.DescribeResourceServer(ctx, &cognitoidentityprovider.DescribeResourceServerInput{
		UserPoolId: aws.String(poolId),
		Identifier: aws.String(identifier),
	})
	return cognitoFound(err)
}

// cognitoPoolScopedLookup reads the composite (user_pool_id, <key>) identity a
// pool-scoped Cognito resource is keyed by.
func cognitoPoolScopedLookup(outputs map[string]interface{}, key string) (poolId, id string, err error) {
	poolId = stringOutputMap(outputs, "user_pool_id")
	if poolId == "" {
		return "", "", pkgerrors.New("no user_pool_id in outputs -- cannot verify")
	}
	id = stringOutputMap(outputs, key)
	if id == "" {
		return "", "", pkgerrors.Errorf("no %s in outputs -- cannot verify", key)
	}
	return poolId, id, nil
}

// cognitoFound normalizes a cognito-idp Describe error into (exists, err).
func cognitoFound(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	var notFound *types.ResourceNotFoundException
	if errors.As(err, &notFound) {
		return false, nil
	}
	return false, err
}

func cognitoIdpClient(cfg aws.Config, region string) *cognitoidentityprovider.Client {
	return cognitoidentityprovider.NewFromConfig(cfg, func(o *cognitoidentityprovider.Options) {
		if region != "" {
			o.Region = region
		}
	})
}
