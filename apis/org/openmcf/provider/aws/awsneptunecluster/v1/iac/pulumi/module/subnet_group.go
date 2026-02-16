package module

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/neptune"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// subnetGroup creates a Neptune subnet group when subnet_ids are provided
// and neptune_subnet_group_name is not set.
func subnetGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*neptune.SubnetGroup, error) {
	spec := locals.AwsNeptuneCluster.Spec
	if spec == nil {
		return nil, nil
	}

	if (spec.NeptuneSubnetGroupName != nil && spec.NeptuneSubnetGroupName.GetValue() != "") || len(spec.SubnetIds) == 0 {
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

	sanitizedName := sanitizeSubnetGroupName(locals.AwsNeptuneCluster.Metadata.Id)

	sg, err := neptune.NewSubnetGroup(ctx, "neptune-subnet-group", &neptune.SubnetGroupArgs{
		Name:        pulumi.String(sanitizedName),
		Description: pulumi.String("Neptune subnet group for " + locals.AwsNeptuneCluster.Metadata.Id),
		SubnetIds:   subnetIds,
		Tags:        pulumi.ToStringMap(locals.Labels),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create neptune subnet group")
	}
	return sg, nil
}

// sanitizeSubnetGroupName ensures the name meets AWS Neptune subnet group naming requirements:
// lowercase alphanumeric characters, hyphens, underscores, and periods; max 255 chars.
func sanitizeSubnetGroupName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")

	re := regexp.MustCompile(`[^a-z0-9._-]`)
	name = re.ReplaceAllString(name, "-")

	re = regexp.MustCompile(`-+`)
	name = re.ReplaceAllString(name, "-")

	name = strings.Trim(name, "-.")

	if name == "" {
		name = "neptune-subnet-group"
	}

	if len(name) > 255 {
		name = name[:255]
		name = strings.Trim(name, "-.")
	}

	return name
}
