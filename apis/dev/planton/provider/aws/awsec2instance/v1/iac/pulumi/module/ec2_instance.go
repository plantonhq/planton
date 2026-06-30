package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/valuefrom"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	awsec2instancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsec2instance/v1"
)

// ec2Instance creates a single AWS EC2 instance in a private subnet.
// It keeps logic linear and “Terraform‑like” for readability.
func ec2Instance(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	// --------‑‑‑ Validate / satisfy connection‑method requirements ------
	var (
		keyPair      *ec2.KeyPair    // AWS side key‑pair
		generatedKey *tls.PrivateKey // Locally generated key (if any)
		err          error
	)

	switch locals.AwsEc2Instance.Spec.GetConnectionMethod() {
	case awsec2instancev1.AwsEc2InstanceConnectionMethod_SSM:
		// SSM needs an instance profile
		if locals.AwsEc2Instance.Spec.IamInstanceProfileArn == nil ||
			locals.AwsEc2Instance.Spec.IamInstanceProfileArn.GetValue() == "" {
			return errors.New("iam_instance_profile_arn is required when connection_method = SSM")
		}

	case awsec2instancev1.AwsEc2InstanceConnectionMethod_BASTION,
		awsec2instancev1.AwsEc2InstanceConnectionMethod_INSTANCE_CONNECT:
		// If the user did not supply a key‑name, autogenerate one
		if locals.AwsEc2Instance.Spec.KeyName == "" {
			// 1) Generate a new RSA‑4096 key pair locally
			generatedKey, err = tls.NewPrivateKey(ctx,
				fmt.Sprintf("%s-autokeygen", locals.AwsEc2Instance.Metadata.Name),
				&tls.PrivateKeyArgs{
					Algorithm: pulumi.String("RSA"),
					RsaBits:   pulumi.Int(4096),
				})
			if err != nil {
				return errors.Wrap(err, "generate ssh private key")
			}

			// 2) Register the public key with AWS
			keyPair, err = ec2.NewKeyPair(ctx,
				fmt.Sprintf("%s-autokey", locals.AwsEc2Instance.Metadata.Name),
				&ec2.KeyPairArgs{
					KeyName:   pulumi.String(fmt.Sprintf("%s-autokey", locals.AwsEc2Instance.Metadata.Name)),
					PublicKey: generatedKey.PublicKeyOpenssh,
					Tags:      pulumi.ToStringMap(locals.AwsTags),
				},
				pulumi.Provider(provider))
			if err != nil {
				return errors.Wrap(err, "create aws key pair")
			}
		}
	}

	// --------‑‑‑ Convert security‑group refs ---------------------------
	sgIDs := pulumi.ToStringArray(valuefrom.ToStringArray(locals.AwsEc2Instance.Spec.SecurityGroupIds))

	// --------‑‑‑ Root block device (always include – defaults to 30 GiB) -
	rootBlock := &ec2.InstanceRootBlockDeviceArgs{
		VolumeSize: pulumi.Int(int(locals.AwsEc2Instance.Spec.GetRootVolumeSizeGb())),
	}

	// --------‑‑‑ Assemble EC2 arguments --------------------------------
	instanceArgs := &ec2.InstanceArgs{
		Ami:                   pulumi.String(locals.AwsEc2Instance.Spec.AmiId),
		InstanceType:          pulumi.String(locals.AwsEc2Instance.Spec.InstanceType),
		SubnetId:              pulumi.String(locals.AwsEc2Instance.Spec.SubnetId.GetValue()),
		VpcSecurityGroupIds:   sgIDs,
		EbsOptimized:          pulumi.Bool(locals.AwsEc2Instance.Spec.EbsOptimized),
		DisableApiTermination: pulumi.Bool(locals.AwsEc2Instance.Spec.DisableApiTermination),
		RootBlockDevice:       rootBlock,
		Tags:                  pulumi.ToStringMap(locals.AwsTags),
	}

	// Wire‑up the key name (user‑supplied or generated)
	if locals.AwsEc2Instance.Spec.KeyName != "" {
		instanceArgs.KeyName = pulumi.String(locals.AwsEc2Instance.Spec.KeyName)
	} else if keyPair != nil {
		instanceArgs.KeyName = keyPair.KeyName
	}

	// Optional fields
	if locals.AwsEc2Instance.Spec.IamInstanceProfileArn != nil &&
		locals.AwsEc2Instance.Spec.IamInstanceProfileArn.GetValue() != "" {
		instanceArgs.IamInstanceProfile = pulumi.String(locals.AwsEc2Instance.Spec.IamInstanceProfileArn.GetValue())
	}
	if locals.AwsEc2Instance.Spec.UserData != "" {
		instanceArgs.UserData = pulumi.String(locals.AwsEc2Instance.Spec.UserData)
	}

	// --------‑‑‑ Create the instance -----------------------------------
	createdInstance, err := ec2.NewInstance(ctx,
		locals.AwsEc2Instance.Metadata.Name,
		instanceArgs,
		pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "create aws ec2 instance")
	}

	// --------‑‑‑ Export stack outputs ----------------------------------
	ctx.Export(OpInstanceId, createdInstance.ID())
	ctx.Export(OpPrivateIp, createdInstance.PrivateIp)
	ctx.Export(OpPrivateDnsName, createdInstance.PrivateDns)
	ctx.Export(OpAvailabilityZone, createdInstance.AvailabilityZone)
	ctx.Export(OpInstanceProfileArn, createdInstance.IamInstanceProfile)

	// If we generated a key, expose it
	if generatedKey != nil {
		ctx.Export(OpSshPrivateKey, generatedKey.PrivateKeyPem)
		ctx.Export(OpSshPublicKey, generatedKey.PublicKeyOpenssh)
	}

	return nil
}
