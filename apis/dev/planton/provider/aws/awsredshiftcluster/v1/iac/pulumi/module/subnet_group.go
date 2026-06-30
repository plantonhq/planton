package module

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/redshift"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// subnetGroup creates a Redshift Subnet Group when subnetIds are provided and cluster_subnet_group_name is not set.
func subnetGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*redshift.SubnetGroup, error) {
	spec := locals.AwsRedshiftCluster.Spec
	if spec == nil {
		return nil, nil
	}

	if (spec.ClusterSubnetGroupName != nil && spec.ClusterSubnetGroupName.GetValue() != "") || len(spec.SubnetIds) == 0 {
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

	sanitizedName := sanitizeSubnetGroupName(locals.AwsRedshiftCluster.Metadata.Id)

	sg, err := redshift.NewSubnetGroup(ctx, "cluster-subnet-group", &redshift.SubnetGroupArgs{
		Name:      pulumi.String(sanitizedName),
		SubnetIds: subnetIds,
		Tags:      pulumi.ToStringMap(locals.Labels),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create subnet group")
	}
	return sg, nil
}

// sanitizeSubnetGroupName sanitizes a name to meet AWS Redshift subnet group naming requirements.
func sanitizeSubnetGroupName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")

	re := regexp.MustCompile(`[^a-z0-9._-]`)
	name = re.ReplaceAllString(name, "-")

	re = regexp.MustCompile(`-+`)
	name = re.ReplaceAllString(name, "-")

	name = strings.Trim(name, "-.")

	if name == "" {
		name = "subnet-group"
	}

	if len(name) > 255 {
		name = name[:255]
		name = strings.Trim(name, "-.")
	}

	return name
}
