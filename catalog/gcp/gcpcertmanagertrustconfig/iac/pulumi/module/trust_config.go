package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/certificatemanager"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// trustConfig provisions one Certificate Manager trust config: the CAs (and
// individually allowlisted certificates) a load balancer validates client or
// backend certificates against. Updates are in place -- rotating a CA never
// replaces the config.
func trustConfig(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, projectService pulumi.Resource) error {
	spec := locals.GcpCertManagerTrustConfig.Spec

	args := &certificatemanager.TrustConfigArgs{
		Name:     pulumi.String(locals.TrustConfigName),
		Location: pulumi.String(locals.Location),
		Labels:   pulumi.ToStringMap(locals.GcpLabels),
	}
	if locals.ProjectId != "" {
		args.Project = pulumi.StringPtr(locals.ProjectId)
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	// Unset defers to the provider default (DELETE).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	// Each PEM is its own one-field block in the provider; the spec lifts
	// those wrappers into plain string lists. The certificates are public, but
	// the provider masks trust-store certificates, so Pulumi state matches.
	trustStores := certificatemanager.TrustConfigTrustStoreArray{}
	for _, store := range spec.TrustStores {
		anchors := certificatemanager.TrustConfigTrustStoreTrustAnchorArray{}
		for _, pem := range store.TrustAnchors {
			anchors = append(anchors, &certificatemanager.TrustConfigTrustStoreTrustAnchorArgs{
				PemCertificate: pulumi.ToSecret(pulumi.String(pem)).(pulumi.StringOutput),
			})
		}
		intermediates := certificatemanager.TrustConfigTrustStoreIntermediateCaArray{}
		for _, pem := range store.IntermediateCas {
			intermediates = append(intermediates, &certificatemanager.TrustConfigTrustStoreIntermediateCaArgs{
				PemCertificate: pulumi.ToSecret(pulumi.String(pem)).(pulumi.StringOutput),
			})
		}
		trustStores = append(trustStores, &certificatemanager.TrustConfigTrustStoreArgs{
			TrustAnchors:    anchors,
			IntermediateCas: intermediates,
		})
	}
	if len(trustStores) > 0 {
		args.TrustStores = trustStores
	}

	allowlisted := certificatemanager.TrustConfigAllowlistedCertificateArray{}
	for _, pem := range spec.AllowlistedCertificates {
		allowlisted = append(allowlisted, &certificatemanager.TrustConfigAllowlistedCertificateArgs{
			PemCertificate: pulumi.String(pem),
		})
	}
	if len(allowlisted) > 0 {
		args.AllowlistedCertificates = allowlisted
	}

	createdTrustConfig, err := certificatemanager.NewTrustConfig(ctx,
		locals.GcpCertManagerTrustConfig.Metadata.Name,
		args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{projectService}))
	if err != nil {
		return errors.Wrap(err, "failed to create trust config")
	}

	ctx.Export(OpTrustConfigId, createdTrustConfig.ID())
	ctx.Export(OpTrustConfigName, createdTrustConfig.Name)
	ctx.Export(OpLocation, pulumi.String(locals.Location))

	return nil
}
