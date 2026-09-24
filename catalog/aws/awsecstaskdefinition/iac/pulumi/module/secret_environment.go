package module

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
	awsecstaskdefinitionv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsecstaskdefinition/v1alpha1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/secretsmanager"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// secretNameChars is what Secrets Manager accepts in a secret name.
var secretNameChars = regexp.MustCompile(`^[A-Za-z0-9/_+=.@-]+$`)

// storedSecretEnv is one secret_environment entry after its value is stored:
// the container and variable it belongs to, and the valueFrom the container
// definition reads -- the secret's ARN pinned to the stored version.
type storedSecretEnv struct {
	container string
	name      string
	valueFrom pulumi.StringOutput
}

// storeSecretEnvironment keeps every secret_environment value in its own
// Secrets Manager secret named "<family>/<container>/<name>", readable only
// by the execution role through the secret's resource policy (so the
// author's role needs no edit), and returns each entry's pinned valueFrom.
//
// Two choices a reader would otherwise rediscover:
//   - The recovery window is zero. The copy is derived -- whoever supplied
//     the value holds it -- and a 30-day scheduled deletion would reserve the
//     name, so redeploying the same task definition would fail to create it.
//   - valueFrom pins the version id ("<arn>:::<version-id>"). A changed value
//     stores a new version, the container definitions change, and a new
//     revision registers, so a referencing service rolls: rotation is a
//     deploy the platform sees.
func storeSecretEnvironment(
	ctx *pulumi.Context,
	family string,
	spec *awsecstaskdefinitionv1alpha1.AwsEcsTaskDefinitionSpec,
	tags map[string]string,
	provider pulumi.ProviderResource,
) ([]storedSecretEnv, error) {
	var stored []storedSecretEnv
	executionRole := spec.ExecutionRole.GetValue()
	for _, container := range spec.Containers {
		names := make([]string, 0, len(container.SecretEnvironment))
		for name := range container.SecretEnvironment {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			secretName := fmt.Sprintf("%s/%s/%s", family, container.Name, name)
			if !secretNameChars.MatchString(secretName) || len(secretName) > 512 {
				return nil, errors.Errorf(
					"secret_environment variable %q in container %q cannot be stored: Secrets Manager secret %q may hold only letters, digits and /_+=.@- and at most 512 characters -- rename the variable",
					name, container.Name, secretName)
			}
			policy, err := readableOnlyBy(executionRole)
			if err != nil {
				return nil, err
			}
			resourceName := "secret-environment-" + strings.ReplaceAll(secretName, "/", "-")
			secret, err := secretsmanager.NewSecret(ctx, resourceName, &secretsmanager.SecretArgs{
				Name:                 pulumi.String(secretName),
				Description:          pulumi.String(fmt.Sprintf("Value of %s for container %s of ECS task definition family %s", name, container.Name, family)),
				RecoveryWindowInDays: pulumi.Int(0),
				Policy:               pulumi.String(policy),
				Tags:                 pulumi.ToStringMap(tags),
			}, pulumi.Provider(provider))
			if err != nil {
				return nil, errors.Wrapf(err, "failed to create Secrets Manager secret %s", secretName)
			}
			version, err := secretsmanager.NewSecretVersion(ctx, resourceName, &secretsmanager.SecretVersionArgs{
				SecretId:     secret.ID(),
				SecretString: pulumi.ToSecret(pulumi.String(container.SecretEnvironment[name])).(pulumi.StringOutput),
			}, pulumi.Provider(provider), pulumi.Parent(secret))
			if err != nil {
				return nil, errors.Wrapf(err, "failed to store the value of %s", secretName)
			}
			stored = append(stored, storedSecretEnv{
				container: container.Name,
				name:      name,
				valueFrom: pulumi.Sprintf("%s:::%s", secret.Arn, version.VersionId),
			})
		}
	}
	return stored, nil
}

// readableOnlyBy is the resource policy that lets exactly one role -- the
// execution role the ECS agent fetches secrets as -- read the secret.
func readableOnlyBy(roleArn string) (string, error) {
	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{{
			"Sid":       "ExecutionRoleReadsAtTaskStart",
			"Effect":    "Allow",
			"Principal": map[string]string{"AWS": roleArn},
			"Action":    "secretsmanager:GetSecretValue",
			"Resource":  "*",
		}},
	}
	document, err := json.Marshal(policy)
	if err != nil {
		return "", errors.Wrap(err, "failed to render the secret's resource policy")
	}
	return string(document), nil
}
