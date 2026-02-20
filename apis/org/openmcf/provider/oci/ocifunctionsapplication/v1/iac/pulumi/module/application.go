package module

import (
	"github.com/pkg/errors"
	ocifunctionsapplicationv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocifunctionsapplication/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/functions"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func applicationResource(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciFunctionsApplication.Spec

	subnetIds := make(pulumi.StringArray, len(spec.SubnetIds))
	for i, s := range spec.SubnetIds {
		subnetIds[i] = pulumi.String(s.GetValue())
	}

	args := &functions.ApplicationArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		DisplayName:   pulumi.String(locals.DisplayName),
		SubnetIds:     subnetIds,
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}

	if shape, ok := shapeMap[spec.Shape]; ok {
		args.Shape = pulumi.String(shape)
	}

	if len(spec.Config) > 0 {
		args.Config = pulumi.ToStringMap(spec.Config)
	}

	if len(spec.NetworkSecurityGroupIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(spec.NetworkSecurityGroupIds))
		for i, n := range spec.NetworkSecurityGroupIds {
			nsgIds[i] = pulumi.String(n.GetValue())
		}
		args.NetworkSecurityGroupIds = nsgIds
	}

	if spec.SyslogUrl != "" {
		args.SyslogUrl = pulumi.String(spec.SyslogUrl)
	}

	if spec.ImagePolicyConfig != nil {
		args.ImagePolicyConfig = buildImagePolicyConfig(spec.ImagePolicyConfig)
	}

	if spec.TraceConfig != nil {
		args.TraceConfig = buildTraceConfig(spec.TraceConfig)
	}

	createdApp, err := functions.NewApplication(ctx, locals.DisplayName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create functions application")
	}

	ctx.Export(OpApplicationId, createdApp.ID())

	return nil
}

func buildImagePolicyConfig(
	ipc *ocifunctionsapplicationv1.OciFunctionsApplicationSpec_ImagePolicyConfig,
) *functions.ApplicationImagePolicyConfigArgs {
	keyDetails := make(functions.ApplicationImagePolicyConfigKeyDetailArray, len(ipc.KeyDetails))
	for i, kd := range ipc.KeyDetails {
		keyDetails[i] = &functions.ApplicationImagePolicyConfigKeyDetailArgs{
			KmsKeyId: pulumi.String(kd.KmsKeyId.GetValue()),
		}
	}

	return &functions.ApplicationImagePolicyConfigArgs{
		IsPolicyEnabled: pulumi.Bool(ipc.IsPolicyEnabled),
		KeyDetails:      keyDetails,
	}
}

func buildTraceConfig(
	tc *ocifunctionsapplicationv1.OciFunctionsApplicationSpec_TraceConfig,
) *functions.ApplicationTraceConfigArgs {
	traceArgs := &functions.ApplicationTraceConfigArgs{}

	if tc.IsEnabled != nil {
		traceArgs.IsEnabled = pulumi.Bool(*tc.IsEnabled)
	}

	if tc.DomainId != "" {
		traceArgs.DomainId = pulumi.String(tc.DomainId)
	}

	return traceArgs
}
