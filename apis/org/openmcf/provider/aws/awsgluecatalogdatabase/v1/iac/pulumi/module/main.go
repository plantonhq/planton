package module

import (
	"github.com/pkg/errors"
	awsgluecatalogdatabase "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsgluecatalogdatabase/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates Glue Data Catalog database creation and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsgluecatalogdatabase.AwsGlueCatalogDatabaseStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(awsProviderConfig.GetRegion()),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	if err := catalogDatabase(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "glue catalog database")
	}

	return nil
}
