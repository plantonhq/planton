package verify

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	pkgerrors "github.com/pkg/errors"
)

// iamInstanceProfileVerifier verifies an AwsIamInstanceProfile via
// GetInstanceProfile, keyed on the profile name. IAM is a global service, so
// the region parameter is ignored. A deleted profile returns the typed
// NoSuchEntity error, which is the "absent" signal; any other error is a
// genuine failure and must surface.
type iamInstanceProfileVerifier struct{}

func (*iamInstanceProfileVerifier) IDOutputKey() string { return "instance_profile_name" }

func (*iamInstanceProfileVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, _ string) error {
	exists, err := iamInstanceProfileExists(ctx, cfg, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsiaminstanceprofile verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awsiaminstanceprofile %q not found after deploy", id)
	}
	return nil
}

func (*iamInstanceProfileVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, _ string) error {
	exists, err := iamInstanceProfileExists(ctx, cfg, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsiaminstanceprofile verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awsiaminstanceprofile %q still exists after destroy", id)
	}
	return nil
}

func iamInstanceProfileExists(ctx context.Context, cfg aws.Config, profileName string) (bool, error) {
	client := iam.NewFromConfig(cfg)
	_, err := client.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{InstanceProfileName: aws.String(profileName)})
	if err != nil {
		if isIamNoSuchEntity(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
