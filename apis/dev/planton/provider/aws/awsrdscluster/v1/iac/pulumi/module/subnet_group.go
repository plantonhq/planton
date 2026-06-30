package module

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// subnetGroup creates a DB Subnet Group when subnetIds are provided and dbSubnetGroupName is not set.
func subnetGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*rds.SubnetGroup, error) {
	spec := locals.AwsRdsCluster.Spec
	if spec == nil {
		return nil, nil
	}

	if (spec.DbSubnetGroupName != nil && spec.DbSubnetGroupName.GetValue() != "") || len(spec.SubnetIds) == 0 {
		return nil, nil
	}

	var subnetIds pulumi.StringArray
	for _, s := range spec.SubnetIds {
		if s.GetValue() != "" {
			subnetIds = append(subnetIds, pulumi.String(s.GetValue()))
		}
	}
	if len(subnetIds) == 0 {
		return nil, nil
	}

	// Sanitize the subnet group name to meet AWS requirements:
	// Only lowercase alphanumeric characters, hyphens, underscores, periods, and spaces are allowed
	sanitizedName := sanitizeSubnetGroupName(locals.AwsRdsCluster.Metadata.Id)

	sg, err := rds.NewSubnetGroup(ctx, "cluster-subnet-group", &rds.SubnetGroupArgs{
		Name:      pulumi.String(sanitizedName),
		SubnetIds: subnetIds,
		Tags:      pulumi.ToStringMap(locals.Labels),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create subnet group")
	}
	return sg, nil
}

// sanitizeSubnetGroupName sanitizes a name to meet AWS RDS subnet group naming requirements:
// Only lowercase alphanumeric characters, hyphens, underscores, and periods are allowed.
// AWS documentation says spaces are allowed, but in practice they often cause issues, so we replace them with hyphens.
func sanitizeSubnetGroupName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace spaces with hyphens (AWS allows spaces but it's safer to use hyphens)
	name = strings.ReplaceAll(name, " ", "-")

	// Replace any character that's not lowercase alphanumeric, hyphen, underscore, or period with a hyphen
	// Regex: [^a-z0-9._-] matches anything that's not allowed
	re := regexp.MustCompile(`[^a-z0-9._-]`)
	name = re.ReplaceAllString(name, "-")

	// Replace multiple consecutive hyphens with a single hyphen
	re = regexp.MustCompile(`-+`)
	name = re.ReplaceAllString(name, "-")

	// Remove leading/trailing hyphens and periods
	name = strings.Trim(name, "-.")

	// Ensure the name is not empty (fallback to a default if needed)
	if name == "" {
		name = "subnet-group"
	}

	// AWS RDS subnet group names have a max length of 255 characters, but we'll limit to 255 to be safe
	if len(name) > 255 {
		name = name[:255]
		// Trim any trailing hyphens/periods after truncation
		name = strings.Trim(name, "-.")
	}

	return name
}
