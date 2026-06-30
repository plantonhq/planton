package module

import (
	"github.com/pkg/errors"
	ocikmskeyv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocikmskey/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/kms"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func key(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciKmsKey.Spec

	keyArgs := &kms.KeyArgs{
		CompartmentId:      pulumi.String(spec.CompartmentId.GetValue()),
		DisplayName:        pulumi.String(locals.DisplayName),
		ManagementEndpoint: pulumi.String(spec.ManagementEndpoint.GetValue()),
		KeyShape:           buildKeyShape(spec.KeyShape),
		FreeformTags:       pulumi.ToStringMap(locals.FreeformTags),
	}

	if pm, ok := protectionModeMap[spec.ProtectionMode]; ok {
		keyArgs.ProtectionMode = pulumi.StringPtr(pm)
	}

	if spec.IsAutoRotationEnabled {
		keyArgs.IsAutoRotationEnabled = pulumi.BoolPtr(true)
	}

	if spec.AutoKeyRotationDetails != nil {
		keyArgs.AutoKeyRotationDetails = buildAutoKeyRotationDetails(spec.AutoKeyRotationDetails)
	}

	if spec.ExternalKeyReference != nil {
		keyArgs.ExternalKeyReference = buildExternalKeyReference(spec.ExternalKeyReference)
	}

	createdKey, err := kms.NewKey(ctx, locals.DisplayName, keyArgs, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create kms key")
	}

	ctx.Export(OpKeyId, createdKey.ID())
	ctx.Export(OpCurrentKeyVersion, createdKey.CurrentKeyVersion)

	return nil
}

func buildKeyShape(shape *ocikmskeyv1.OciKmsKeySpec_KeyShape) *kms.KeyKeyShapeArgs {
	args := &kms.KeyKeyShapeArgs{
		Algorithm: pulumi.String(algorithmMap[shape.Algorithm]),
		Length:    pulumi.Int(int(shape.Length)),
	}

	if curveStr, ok := curveIdMap[shape.CurveId]; ok {
		args.CurveId = pulumi.StringPtr(curveStr)
	}

	return args
}

func buildAutoKeyRotationDetails(
	details *ocikmskeyv1.OciKmsKeySpec_AutoKeyRotationDetails,
) *kms.KeyAutoKeyRotationDetailsArgs {
	args := &kms.KeyAutoKeyRotationDetailsArgs{}

	if details.RotationIntervalInDays > 0 {
		args.RotationIntervalInDays = pulumi.IntPtr(int(details.RotationIntervalInDays))
	}

	if details.TimeOfScheduleStart != "" {
		args.TimeOfScheduleStart = pulumi.StringPtr(details.TimeOfScheduleStart)
	}

	return args
}

func buildExternalKeyReference(
	ref *ocikmskeyv1.OciKmsKeySpec_ExternalKeyReference,
) *kms.KeyExternalKeyReferenceArgs {
	return &kms.KeyExternalKeyReferenceArgs{
		ExternalKeyId: pulumi.String(ref.ExternalKeyId),
	}
}
