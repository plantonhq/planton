package verify

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	pkgerrors "github.com/pkg/errors"
)

// iamRoleVerifier verifies an AwsIamRole via GetRole, keyed on the role name.
// IAM is a global service, so the region parameter is ignored: the role is the
// same object from every regional endpoint. A deleted role returns the typed
// NoSuchEntity error, which is the "absent" signal; any other error is a
// genuine failure and must surface.
type iamRoleVerifier struct{}

func (*iamRoleVerifier) IDOutputKey() string { return "role_name" }

func (*iamRoleVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, _ string) error {
	exists, err := iamRoleExists(ctx, cfg, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsiamrole verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awsiamrole %q not found after deploy", id)
	}
	return nil
}

func (*iamRoleVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, _ string) error {
	exists, err := iamRoleExists(ctx, cfg, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsiamrole verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awsiamrole %q still exists after destroy", id)
	}
	return nil
}

func iamRoleExists(ctx context.Context, cfg aws.Config, roleName string) (bool, error) {
	client := iam.NewFromConfig(cfg)
	_, err := client.GetRole(ctx, &iam.GetRoleInput{RoleName: aws.String(roleName)})
	if err != nil {
		if isIamNoSuchEntity(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
