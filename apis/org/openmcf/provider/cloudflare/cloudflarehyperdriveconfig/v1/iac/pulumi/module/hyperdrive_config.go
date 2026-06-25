package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// hyperdriveConfig provisions the Cloudflare Hyperdrive config and exports its outputs.
func hyperdriveConfig(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.HyperdriveConfig, error) {
	spec := locals.CloudflareHyperdriveConfig.Spec
	origin := spec.Origin

	// Sensitive values arrive resolved (literal) via StringValueOrRef.GetValue().
	originArgs := &cloudflare.HyperdriveConfigOriginArgs{
		Database: pulumi.String(origin.Database),
		Scheme:   pulumi.String(origin.Scheme.String()),
		User:     pulumi.String(origin.User),
		Host:     pulumi.String(origin.Host),
		Password: pulumi.String(origin.Password.GetValue()),
	}
	if origin.Port > 0 {
		originArgs.Port = pulumi.IntPtr(int(origin.Port))
	}
	if origin.AccessClientId != "" {
		originArgs.AccessClientId = pulumi.StringPtr(origin.AccessClientId)
	}
	if origin.AccessClientSecret != nil && origin.AccessClientSecret.GetValue() != "" {
		originArgs.AccessClientSecret = pulumi.StringPtr(origin.AccessClientSecret.GetValue())
	}

	args := &cloudflare.HyperdriveConfigArgs{
		AccountId: pulumi.String(spec.AccountId),
		Name:      pulumi.String(spec.Name),
		Origin:    originArgs,
	}

	if c := spec.Caching; c != nil {
		cachingArgs := &cloudflare.HyperdriveConfigCachingArgs{
			Disabled: pulumi.BoolPtr(c.Disabled),
		}
		if c.MaxAge > 0 {
			cachingArgs.MaxAge = pulumi.IntPtr(int(c.MaxAge))
		}
		if c.StaleWhileRevalidate > 0 {
			cachingArgs.StaleWhileRevalidate = pulumi.IntPtr(int(c.StaleWhileRevalidate))
		}
		args.Caching = cachingArgs
	}

	if m := spec.Mtls; m != nil && (m.CaCertificateId != "" || m.MtlsCertificateId != "" || m.Sslmode != "") {
		mtlsArgs := &cloudflare.HyperdriveConfigMtlsArgs{}
		if m.CaCertificateId != "" {
			mtlsArgs.CaCertificateId = pulumi.StringPtr(m.CaCertificateId)
		}
		if m.MtlsCertificateId != "" {
			mtlsArgs.MtlsCertificateId = pulumi.StringPtr(m.MtlsCertificateId)
		}
		if m.Sslmode != "" {
			mtlsArgs.Sslmode = pulumi.StringPtr(m.Sslmode)
		}
		args.Mtls = mtlsArgs
	}

	if spec.OriginConnectionLimit > 0 {
		args.OriginConnectionLimit = pulumi.IntPtr(int(spec.OriginConnectionLimit))
	}

	created, err := cloudflare.NewHyperdriveConfig(
		ctx,
		"hyperdrive-config",
		args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare hyperdrive config")
	}

	ctx.Export(OpHyperdriveId, created.ID())
	ctx.Export(OpName, created.Name)

	return created, nil
}
