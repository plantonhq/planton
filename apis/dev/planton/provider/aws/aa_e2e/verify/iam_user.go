package verify

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	pkgerrors "github.com/pkg/errors"
)

// iamUserVerifier verifies an AwsIamUser via GetUser, keyed on the user name.
// IAM is a global service, so the region parameter is ignored. A deleted user
// returns the typed NoSuchEntity error, which is the "absent" signal; any
// other error is a genuine failure and must surface.
type iamUserVerifier struct{}

func (*iamUserVerifier) IDOutputKey() string { return "user_name" }

func (*iamUserVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, _ string) error {
	exists, err := iamUserExists(ctx, cfg, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsiamuser verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awsiamuser %q not found after deploy", id)
	}
	return nil
}

func (*iamUserVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, _ string) error {
	exists, err := iamUserExists(ctx, cfg, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsiamuser verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awsiamuser %q still exists after destroy", id)
	}
	return nil
}

func iamUserExists(ctx context.Context, cfg aws.Config, userName string) (bool, error) {
	client := iam.NewFromConfig(cfg)
	_, err := client.GetUser(ctx, &iam.GetUserInput{UserName: aws.String(userName)})
	if err != nil {
		if isIamNoSuchEntity(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
