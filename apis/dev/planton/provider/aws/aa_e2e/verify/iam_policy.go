package verify

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/smithy-go"
	pkgerrors "github.com/pkg/errors"
)

// iamPolicyVerifier verifies an AwsIamPolicy via GetPolicy, keyed on the
// policy ARN (the AWS API for managed policies takes the ARN, not the name).
// IAM is a global service, so the region parameter is ignored. A deleted
// policy returns the typed NoSuchEntity error, which is the "absent" signal;
// any other error is a genuine failure and must surface.
type iamPolicyVerifier struct{}

func (*iamPolicyVerifier) IDOutputKey() string { return "policy_arn" }

func (*iamPolicyVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, _ string) error {
	exists, err := iamPolicyExists(ctx, cfg, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsiampolicy verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awsiampolicy %q not found after deploy", id)
	}
	return nil
}

func (*iamPolicyVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, _ string) error {
	exists, err := iamPolicyExists(ctx, cfg, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsiampolicy verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awsiampolicy %q still exists after destroy", id)
	}
	return nil
}

func iamPolicyExists(ctx context.Context, cfg aws.Config, policyArn string) (bool, error) {
	client := iam.NewFromConfig(cfg)
	_, err := client.GetPolicy(ctx, &iam.GetPolicyInput{PolicyArn: aws.String(policyArn)})
	if err != nil {
		if isIamNoSuchEntity(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// isIamNoSuchEntity reports whether err is IAM's typed "the entity does not
// exist" error -- the shared absent-signal for every IAM verifier.
func isIamNoSuchEntity(err error) bool {
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && apiErr.ErrorCode() == "NoSuchEntity"
}
