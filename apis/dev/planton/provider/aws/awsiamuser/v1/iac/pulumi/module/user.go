package module

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
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

func iamUser(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*IamUserResults, error) {
	spec := locals.AwsIamUser.Spec
	userName := spec.UserName

	usr, err := iam.NewUser(ctx, userName, &iam.UserArgs{
		Name: pulumi.String(userName),
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create iam user")
	}

	// Attach managed policies
	for idx, policyArn := range spec.ManagedPolicyArns {
		attachName := fmt.Sprintf("%s-attach-%d", userName, idx)
		_, err := iam.NewUserPolicyAttachment(ctx, attachName, &iam.UserPolicyAttachmentArgs{
			User:      usr.Name,
			PolicyArn: pulumi.String(policyArn),
		}, pulumi.Provider(provider))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to attach policy arn %s", policyArn)
		}
	}

	// Inline policies
	for policyName, inlineStruct := range spec.InlinePolicies {
		inlinePolicyString, err := structToJSONString(inlineStruct)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to marshal inline policy for %s", policyName)
		}
		inlineName := fmt.Sprintf("%s-inline-%s", userName, policyName)
		_, err = iam.NewUserPolicy(ctx, inlineName, &iam.UserPolicyArgs{
			User:   usr.Name,
			Policy: pulumi.String(inlinePolicyString),
		}, pulumi.Provider(provider))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create inline policy %s", policyName)
		}
	}

	// Optionally create access keys
	var accessKey *iam.AccessKey
	if !spec.DisableAccessKeys {
		akName := fmt.Sprintf("%s-ak", userName)
		var akErr error
		accessKey, akErr = iam.NewAccessKey(ctx, akName, &iam.AccessKeyArgs{
			User: usr.Name,
		}, pulumi.Provider(provider))
		if akErr != nil {
			return nil, errors.Wrap(akErr, "failed to create access key")
		}
	}

	// Build console URL: https://signin.aws.amazon.com/console
	consoleUrl := pulumi.Sprintf("%s", "https://signin.aws.amazon.com/console")

	var accessKeyId pulumi.StringPtrOutput
	var secretAccessKey pulumi.StringPtrOutput
	if accessKey != nil {
		accessKeyId = accessKey.ID().ApplyT(func(id string) *string {
			v := id
			return &v
		}).(pulumi.StringPtrOutput)
		// Secret is already sensitive; optionally base64-encode for uniform format
		secretAccessKey = accessKey.Secret.ApplyT(func(s string) *string {
			enc := base64.StdEncoding.EncodeToString([]byte(s))
			return &enc
		}).(pulumi.StringPtrOutput)
	}

	return &IamUserResults{
		UserArn:         usr.Arn,
		UserName:        usr.Name,
		UserId:          usr.UniqueId,
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
	m := s.AsMap()
	bytes, err := jsonMarshal(m)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// jsonMarshal is a tiny wrapper to permit testing/mocking if needed.
var jsonMarshal = func(v any) ([]byte, error) { return json.Marshal(v) }
