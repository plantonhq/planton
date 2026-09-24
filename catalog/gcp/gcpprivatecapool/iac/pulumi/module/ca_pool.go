package module

import (
	"strconv"

	"github.com/pkg/errors"
	gcpprivatecapoolv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpprivatecapool/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/certificateauthority"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// caPool creates the CA pool: the tier (immutable), the issuance policy, the
// publishing options, and the at-rest encryption key.
func caPool(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpPrivateCaPool.Spec
	resourceName := locals.GcpPrivateCaPool.Metadata.Name

	// Enable the Certificate Authority Service API first so a fresh project
	// works on the first deploy. DisableOnDestroy stays false: tearing down
	// one pool must never disable the API for every pool, template, and
	// certificate in the project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("privateca.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcppca-privateca.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable privateca.googleapis.com api")
	}

	args := &certificateauthority.CaPoolArgs{
		Location: pulumi.String(spec.Location),
		Name:     pulumi.String(locals.CaPoolId),
		Tier:     pulumi.String(spec.Tier),
		Labels:   pulumi.ToStringMap(locals.GcpLabels),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.IssuancePolicy != nil {
		args.IssuancePolicy = issuancePolicy(spec.IssuancePolicy)
	}
	if publishing := spec.PublishingOptions; publishing != nil {
		args.PublishingOptions = &certificateauthority.CaPoolPublishingOptionsArgs{
			PublishCaCert:  pulumi.Bool(publishing.PublishCaCert),
			PublishCrl:     pulumi.Bool(publishing.PublishCrl),
			EncodingFormat: stringPtr(publishing.EncodingFormat),
		}
	}
	if spec.KmsKeyName.GetValue() != "" {
		args.EncryptionSpec = &certificateauthority.CaPoolEncryptionSpecArgs{
			CloudKmsKey: pulumi.String(spec.KmsKeyName.GetValue()),
		}
	}

	// Engine-side destroy stance: DELETE (the provider's default), PREVENT,
	// or ABANDON. Sent only when set so the provider default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := certificateauthority.NewCaPool(ctx, resourceName, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create certificate authority service ca pool")
	}

	ctx.Export(OpName, created.ID())
	ctx.Export(OpCaPoolId, created.Name)
	ctx.Export(OpLocation, created.Location)
	return nil
}

// issuancePolicy maps the policy every issued certificate follows. Each
// lever is sent only when the spec sets it; a message that is set sends
// both of its flags.
func issuancePolicy(policy *gcpprivatecapoolv1alpha1.GcpPrivateCaPoolIssuancePolicy) *certificateauthority.CaPoolIssuancePolicyArgs {
	args := &certificateauthority.CaPoolIssuancePolicyArgs{
		MaximumLifetime:  stringPtr(policy.MaximumLifetime),
		BackdateDuration: stringPtr(policy.BackdateDuration),
	}
	if len(policy.AllowedKeyTypes) > 0 {
		keyTypes := certificateauthority.CaPoolIssuancePolicyAllowedKeyTypeArray{}
		for _, keyType := range policy.AllowedKeyTypes {
			keyTypeArgs := &certificateauthority.CaPoolIssuancePolicyAllowedKeyTypeArgs{}
			if rsa := keyType.GetRsa(); rsa != nil {
				keyTypeArgs.Rsa = &certificateauthority.CaPoolIssuancePolicyAllowedKeyTypeRsaArgs{
					MinModulusSize: modulusSize(rsa.MinModulusSize),
					MaxModulusSize: modulusSize(rsa.MaxModulusSize),
				}
			}
			if algorithm := keyType.GetEllipticCurveSignatureAlgorithm(); algorithm != "" {
				keyTypeArgs.EllipticCurve = &certificateauthority.CaPoolIssuancePolicyAllowedKeyTypeEllipticCurveArgs{
					SignatureAlgorithm: pulumi.String(algorithm),
				}
			}
			keyTypes = append(keyTypes, keyTypeArgs)
		}
		args.AllowedKeyTypes = keyTypes
	}
	if modes := policy.AllowedIssuanceModes; modes != nil {
		args.AllowedIssuanceModes = &certificateauthority.CaPoolIssuancePolicyAllowedIssuanceModesArgs{
			AllowCsrBasedIssuance:    pulumi.Bool(modes.AllowCsrBasedIssuance),
			AllowConfigBasedIssuance: pulumi.Bool(modes.AllowConfigBasedIssuance),
		}
	}
	if identity := policy.IdentityConstraints; identity != nil {
		identityArgs := &certificateauthority.CaPoolIssuancePolicyIdentityConstraintsArgs{
			AllowSubjectPassthrough:         pulumi.Bool(identity.AllowSubjectPassthrough),
			AllowSubjectAltNamesPassthrough: pulumi.Bool(identity.AllowSubjectAltNamesPassthrough),
		}
		if expression := identity.CelExpression; expression != nil {
			identityArgs.CelExpression = &certificateauthority.CaPoolIssuancePolicyIdentityConstraintsCelExpressionArgs{
				Expression:  pulumi.String(expression.Expression),
				Title:       stringPtr(expression.Title),
				Description: stringPtr(expression.Description),
				Location:    stringPtr(expression.Location),
			}
		}
		args.IdentityConstraints = identityArgs
	}
	if policy.BaselineValues != nil {
		args.BaselineValues = x509Parameters(policy.BaselineValues)
	}
	return args
}

// modulusSize sends an RSA bound as the decimal string the provider takes,
// and leaves 0 unset (Google's "no explicit bound").
func modulusSize(bits int64) pulumi.StringPtrInput {
	if bits == 0 {
		return nil
	}
	return pulumi.String(strconv.FormatInt(bits, 10))
}
