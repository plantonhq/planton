package module

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/valuefrom"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

type IamUserResults struct {
	UserArn         pulumi.StringOutput
	UserName        pulumi.StringOutput
	UserId          pulumi.StringOutput
	ConsoleUrl      pulumi.StringOutput
	AccessKeyId     pulumi.StringPtrOutput
	SecretAccessKey pulumi.StringPtrOutput
}

// iamUser provisions the user and its policy wiring. An IAM user is a
// long-lived identity with permanent credentials -- prefer roles wherever
// temporary credentials work. The user name is mutable (AWS renames in place)
// and so is the path; the permissions boundary caps the maximum the user can
// ever do, which matters most on principals whose credentials do not expire.
func iamUser(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*IamUserResults, error) {
	spec := locals.AwsIamUser.Spec
	userName := spec.UserName

	userArgs := &iam.UserArgs{
		Name: pulumi.String(userName),
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	if spec.Path != "" {
		userArgs.Path = pulumi.StringPtr(spec.Path)
	}
	// The boundary is a ceiling, not a grant: effective permissions are the
	// intersection of this policy and the user's permission policies. A
	// valueFrom reference is resolved to the AwsIamPolicy's policy_arn before
	// the module runs.
	if spec.PermissionsBoundary.GetValue() != "" {
		userArgs.PermissionsBoundary = pulumi.StringPtr(spec.PermissionsBoundary.GetValue())
	}
	// When enabled, deletion also removes credentials created OUTSIDE this
	// resource (login profile, extra access keys, MFA devices, SSH keys,
	// signing certs) instead of failing on them.
	if spec.ForceDestroy {
		userArgs.ForceDestroy = pulumi.BoolPtr(true)
	}

	createdUser, err := iam.NewUser(ctx, userName, userArgs, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create iam user")
	}

	// Each managed-policy attachment is its own resource so attachments
	// reconcile individually: adding or removing an entry attaches or detaches
	// just that policy, and attachments made outside this resource are left
	// alone. valueFrom references were resolved to policy ARNs before the
	// module ran.
	for idx, policyArn := range valuefrom.ToStringArray(spec.ManagedPolicyArns) {
		attachName := fmt.Sprintf("%s-attach-%d", userName, idx)
		_, err := iam.NewUserPolicyAttachment(ctx, attachName, &iam.UserPolicyAttachmentArgs{
			User:      createdUser.Name,
			PolicyArn: pulumi.String(policyArn),
		}, pulumi.Provider(provider), pulumi.Parent(createdUser))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to attach policy arn %s", policyArn)
		}
	}

	// Inline policies live and die with the user -- permissions unique to this
	// user that would be noise as standalone AwsIamPolicy resources.
	for policyName, inlineStruct := range spec.InlinePolicies {
		inlinePolicyString, err := structToJSONString(inlineStruct)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to marshal inline policy for %s", policyName)
		}
		inlineName := fmt.Sprintf("%s-inline-%s", userName, policyName)
		_, err = iam.NewUserPolicy(ctx, inlineName, &iam.UserPolicyArgs{
			Name:   pulumi.String(policyName),
			User:   createdUser.Name,
			Policy: pulumi.String(inlinePolicyString),
		}, pulumi.Provider(provider), pulumi.Parent(createdUser))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create inline policy %s", policyName)
		}
	}

	// One active access key by default -- programmatic access is the usual
	// reason an IAM user exists. No PGP key is used because the platform
	// delivers outputs through its own secret-handling channel.
	var accessKey *iam.AccessKey
	if !spec.DisableAccessKeys {
		akName := fmt.Sprintf("%s-ak", userName)
		var akErr error
		accessKey, akErr = iam.NewAccessKey(ctx, akName, &iam.AccessKeyArgs{
			User: createdUser.Name,
		}, pulumi.Provider(provider), pulumi.Parent(createdUser))
		if akErr != nil {
			return nil, errors.Wrap(akErr, "failed to create access key")
		}
	}

	consoleUrl := pulumi.Sprintf("%s", "https://signin.aws.amazon.com/console")

	var accessKeyId pulumi.StringPtrOutput
	var secretAccessKey pulumi.StringPtrOutput
	if accessKey != nil {
		accessKeyId = accessKey.ID().ApplyT(func(id string) *string {
			v := id
			return &v
		}).(pulumi.StringPtrOutput)
		// Base64-encoded to match the stack-outputs contract (the proto
		// documents the secret as base64), keeping both engines' outputs
		// byte-identical. Pulumi already tracks the value as a secret.
		secretAccessKey = accessKey.Secret.ApplyT(func(s string) *string {
			enc := base64.StdEncoding.EncodeToString([]byte(s))
			return &enc
		}).(pulumi.StringPtrOutput)
	}

	return &IamUserResults{
		UserArn:         createdUser.Arn,
		UserName:        createdUser.Name,
		UserId:          createdUser.UniqueId,
		ConsoleUrl:      consoleUrl,
		AccessKeyId:     accessKeyId,
		SecretAccessKey: secretAccessKey,
	}, nil
}

// structToJSONString converts a google.protobuf.Struct to a raw JSON string.
func structToJSONString(s *structpb.Struct) (string, error) {
	if s == nil {
		return "{}", nil
	}
	bytes, err := json.Marshal(s.AsMap())
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
