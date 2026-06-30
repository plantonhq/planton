package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	ociloggroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ociloggroup/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/logging"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func logResources(ctx *pulumi.Context, locals *Locals, provider *oci.Provider, logGroup *logging.LogGroup) error {
	spec := locals.OciLogGroup.Spec

	for _, l := range spec.Logs {
		resourceName := fmt.Sprintf("%s-%s", locals.GroupName, l.DisplayName)

		args := &logging.LogArgs{
			DisplayName:  pulumi.String(l.DisplayName),
			LogGroupId:   logGroup.ID(),
			LogType:      pulumi.String(strings.ToUpper(l.LogType.String())),
			FreeformTags: pulumi.ToStringMap(locals.FreeformTags),
		}

		if l.IsEnabled != nil {
			args.IsEnabled = pulumi.Bool(*l.IsEnabled)
		}

		if l.RetentionDuration != nil {
			args.RetentionDuration = pulumi.Int(int(*l.RetentionDuration))
		}

		if l.Configuration != nil {
			args.Configuration = buildLogConfiguration(l.Configuration)
		}

		_, err := logging.NewLog(ctx, resourceName, args,
			pulumiOciOpt(provider), pulumi.DependsOn([]pulumi.Resource{logGroup}))
		if err != nil {
			return errors.Wrapf(err, "failed to create log %s", l.DisplayName)
		}
	}

	return nil
}

func buildLogConfiguration(
	cfg *ociloggroupv1.OciLogGroupSpec_Log_ServiceLogConfiguration,
) logging.LogConfigurationPtrInput {
	sourceArgs := &logging.LogConfigurationSourceArgs{
		SourceType: pulumi.String("OCISERVICE"),
		Service:    pulumi.String(cfg.Service),
		Resource:   pulumi.String(cfg.Resource.GetValue()),
		Category:   pulumi.String(cfg.Category),
	}

	if len(cfg.Parameters) > 0 {
		sourceArgs.Parameters = pulumi.ToStringMap(cfg.Parameters)
	}

	configArgs := &logging.LogConfigurationArgs{
		Source: sourceArgs,
	}

	if cfg.CompartmentId != nil {
		configArgs.CompartmentId = pulumi.String(cfg.CompartmentId.GetValue())
	}

	return configArgs
}
