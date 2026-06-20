// Package verify checks that AWS resources created by an E2E scenario exist after
// DEPLOY and are gone after DESTROY. Each component family has its own verifier
// because AWS verification is service-specific (HeadBucket for S3,
// DescribeSubnets for a subnet, ...) -- unlike the single Management-API path a
// SaaS provider uses. All verifiers run against the same ambient credential
// chain the deploy used, so a verification failure reflects real cloud state.
package verify

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pkg/errors"
)

// Verifier checks a single component's AWS resource for existence/absence.
type Verifier interface {
	// IDOutputKey is the stack-output key carrying the identifier used to verify
	// the resource (e.g. "bucket_id").
	IDOutputKey() string
	// VerifyExists returns an error unless the resource exists.
	VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error
	// VerifyAbsent returns an error unless the resource is gone.
	VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error
}

// verifiers maps a component name to its verifier. New AWS components register
// here as they are forged; today it carries the S3 walking-skeleton only.
var verifiers = map[string]Verifier{
	"awss3bucket":        &s3Verifier{},
	"awssubnet":          &subnetVerifier{},
	"awsvpc":             &vpcVerifier{},
	"awsinternetgateway": &internetGatewayVerifier{},
	"awsnatgateway":      &natGatewayVerifier{},
	"awselasticip":       &elasticIpVerifier{},
}

// GetVerifier returns the verifier for a component, or an error if none is registered.
func GetVerifier(component string) (Verifier, error) {
	v, ok := verifiers[component]
	if !ok {
		return nil, errors.Errorf("no AWS verifier registered for component %q", component)
	}
	return v, nil
}
