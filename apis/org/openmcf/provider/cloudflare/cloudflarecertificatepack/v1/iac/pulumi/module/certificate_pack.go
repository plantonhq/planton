package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// certificatePack orders an advanced edge certificate for a zone. The `type`
// default ("advanced") is coalesced here so a standalone module run matches the
// control-plane middleware and the Terraform module byte-for-byte.
func certificatePack(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) error {
	spec := locals.CloudflareCertificatePack.Spec

	certType := spec.GetType()
	if certType == "" {
		certType = "advanced"
	}

	hosts := make(pulumi.StringArray, 0, len(spec.Hosts))
	for _, h := range spec.Hosts {
		hosts = append(hosts, pulumi.String(h))
	}

	args := &cloudflare.CertificatePackArgs{
		ZoneId:               pulumi.String(spec.ZoneId.GetValue()),
		CertificateAuthority: pulumi.String(spec.CertificateAuthority),
		Type:                 pulumi.String(certType),
		ValidationMethod:     pulumi.String(spec.ValidationMethod),
		ValidityDays:         pulumi.Int(int(spec.ValidityDays)),
		Hosts:                hosts,
	}
	if spec.CloudflareBranding {
		args.CloudflareBranding = pulumi.BoolPtr(true)
	}

	created, err := cloudflare.NewCertificatePack(
		ctx,
		"certificate-pack",
		args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudflare certificate pack")
	}

	ctx.Export(OpCertificatePackId, created.ID())
	ctx.Export(OpStatus, created.Status)
	ctx.Export(OpPrimaryCertificate, created.PrimaryCertificate)

	return nil
}
