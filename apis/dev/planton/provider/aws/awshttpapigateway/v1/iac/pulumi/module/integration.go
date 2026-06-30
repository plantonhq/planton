package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// integrations creates deduplicated API Gateway integrations. Routes that share the
// same integration_type, integration_uri, and payload_format_version point to a single
// underlying Integration resource. Returns a map of integration dedup key to the
// created Integration resource.
func integrations(
	ctx *pulumi.Context,
	locals *Locals,
	createdApi *apigatewayv2.Api,
	provider *aws.Provider,
) (map[string]*apigatewayv2.Integration, error) {
	result := make(map[string]*apigatewayv2.Integration)
	counter := 0

	for _, route := range locals.Spec.Routes {
		key := integrationKey(route.Integration)
		if _, exists := result[key]; exists {
			continue // Already created an integration for this backend
		}

		counter++
		resourceName := fmt.Sprintf("%s-integration-%d", locals.ApiName, counter)
		integration := route.Integration

		// Default payload format version to 2.0 for HTTP APIs.
		payloadVersion := "2.0"
		if integration.PayloadFormatVersion != "" {
			payloadVersion = integration.PayloadFormatVersion
		}

		args := &apigatewayv2.IntegrationArgs{
			ApiId:                createdApi.ID(),
			IntegrationType:      pulumi.String(integration.IntegrationType),
			IntegrationUri:       pulumi.String(integration.IntegrationUri.GetValue()),
			PayloadFormatVersion: pulumi.StringPtr(payloadVersion),
		}

		// Integration method (defaults to POST for Lambda, route method for HTTP).
		if integration.IntegrationMethod != "" {
			args.IntegrationMethod = pulumi.StringPtr(integration.IntegrationMethod)
		}

		// Timeout
		if integration.TimeoutMilliseconds > 0 {
			args.TimeoutMilliseconds = pulumi.IntPtr(int(integration.TimeoutMilliseconds))
		}

		created, err := apigatewayv2.NewIntegration(ctx, resourceName, args, pulumi.Provider(provider))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create integration %d", counter)
		}

		result[key] = created
	}

	return result, nil
}
