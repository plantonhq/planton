package module

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/memorydb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// subnetGroup creates a MemoryDB subnet group when subnet_ids are provided.
func subnetGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*memorydb.SubnetGroup, error) {
	spec := locals.Spec
	if spec == nil || len(spec.SubnetIds) == 0 {
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

	sanitizedName := sanitizeSubnetGroupName(locals.Target.Metadata.Id)

	sg, err := memorydb.NewSubnetGroup(ctx, "subnet-group", &memorydb.SubnetGroupArgs{
		Name:        pulumi.String(sanitizedName),
		Description: pulumi.Sprintf("MemoryDB subnet group for %s", locals.Target.Metadata.Id),
		SubnetIds:   subnetIds,
		Tags:        pulumi.ToStringMap(locals.AwsTags),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create subnet group")
	}

	ctx.Export(OpSubnetGroupName, sg.Name)
	return sg, nil
}

// sanitizeSubnetGroupName normalises a name for AWS MemoryDB subnet group naming:
// lowercase alphanumeric and hyphens only, max 255 characters.
func sanitizeSubnetGroupName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")

	re := regexp.MustCompile(`[^a-z0-9-]`)
	name = re.ReplaceAllString(name, "-")

	re = regexp.MustCompile(`-+`)
	name = re.ReplaceAllString(name, "-")

	name = strings.Trim(name, "-")

	if name == "" {
		name = "subnet-group"
	}
	if len(name) > 255 {
		name = name[:255]
		name = strings.Trim(name, "-")
	}
	return name
}
