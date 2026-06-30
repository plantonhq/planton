package module

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/docdb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// subnetGroup creates a DB Subnet Group when subnets are provided and dbSubnetGroup is not set.
func subnetGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*docdb.SubnetGroup, error) {
	spec := locals.AwsDocumentDb.Spec
	if spec == nil {
		return nil, nil
	}

	if (spec.DbSubnetGroup != nil && spec.DbSubnetGroup.GetValue() != "") || len(spec.Subnets) == 0 {
		return nil, nil
	}

	var subnetIds pulumi.StringArray
	for _, s := range spec.Subnets {
		if s.GetValue() != "" {
			subnetIds = append(subnetIds, pulumi.String(s.GetValue()))
		}
	}
	if len(subnetIds) == 0 {
		return nil, nil
	}

	// Sanitize the subnet group name to meet AWS requirements
	sanitizedName := sanitizeSubnetGroupName(locals.AwsDocumentDb.Metadata.Id)

	sg, err := docdb.NewSubnetGroup(ctx, "cluster-subnet-group", &docdb.SubnetGroupArgs{
		Name:        pulumi.String(sanitizedName),
		SubnetIds:   subnetIds,
		Description: pulumi.String("DocumentDB subnet group for " + locals.AwsDocumentDb.Metadata.Id),
		Tags:        pulumi.ToStringMap(locals.Labels),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create subnet group")
	}
	return sg, nil
}

// sanitizeSubnetGroupName sanitizes a name to meet AWS DocumentDB subnet group naming requirements.
func sanitizeSubnetGroupName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")

	// Replace any character that's not lowercase alphanumeric, hyphen, underscore, or period with a hyphen
	re := regexp.MustCompile(`[^a-z0-9._-]`)
	name = re.ReplaceAllString(name, "-")

	// Replace multiple consecutive hyphens with a single hyphen
	re = regexp.MustCompile(`-+`)
	name = re.ReplaceAllString(name, "-")

	// Remove leading/trailing hyphens and periods
	name = strings.Trim(name, "-.")

	// Ensure the name is not empty
	if name == "" {
		name = "subnet-group"
	}

	// AWS subnet group names have a max length of 255 characters
	if len(name) > 255 {
		name = name[:255]
		name = strings.Trim(name, "-.")
	}

	return name
}
