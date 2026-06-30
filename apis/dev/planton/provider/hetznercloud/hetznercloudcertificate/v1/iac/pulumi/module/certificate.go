package module

import (
	"fmt"

	"github.com/pkg/errors"
	hetznercloudcertificatev1 "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud/hetznercloudcertificate/v1"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func certificate(
	ctx *pulumi.Context,
	locals *Locals,
	hcloudProvider *hcloud.Provider,
) error {
	spec := locals.HetznerCloudCertificate.Spec
	name := locals.HetznerCloudCertificate.Metadata.Name

	switch cert := spec.Certificate.(type) {
	case *hetznercloudcertificatev1.HetznerCloudCertificateSpec_Uploaded:
		return uploadedCertificate(ctx, name, cert.Uploaded, locals, hcloudProvider)
	case *hetznercloudcertificatev1.HetznerCloudCertificateSpec_Managed:
		return managedCertificate(ctx, name, cert.Managed, locals, hcloudProvider)
	default:
		return fmt.Errorf("certificate oneof is not set; exactly one of uploaded or managed must be provided")
	}
}

func uploadedCertificate(
	ctx *pulumi.Context,
	name string,
	config *hetznercloudcertificatev1.UploadedCertificateConfig,
	locals *Locals,
	hcloudProvider *hcloud.Provider,
) error {
	created, err := hcloud.NewUploadedCertificate(
		ctx,
		"certificate",
		&hcloud.UploadedCertificateArgs{
			Name:        pulumi.String(name),
			Certificate: pulumi.String(config.Certificate),
			PrivateKey:  pulumi.ToSecret(pulumi.String(config.PrivateKey)).(pulumi.StringInput),
			Labels:      pulumi.ToStringMap(locals.Labels),
		},
		pulumi.Provider(hcloudProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create uploaded certificate")
	}

	ctx.Export(OpCertificateId, created.ID())
	ctx.Export(OpType, created.Type)
	ctx.Export(OpFingerprint, created.Fingerprint)
	ctx.Export(OpNotValidBefore, created.NotValidBefore)
	ctx.Export(OpNotValidAfter, created.NotValidAfter)

	return nil
}

func managedCertificate(
	ctx *pulumi.Context,
	name string,
	config *hetznercloudcertificatev1.ManagedCertificateConfig,
	locals *Locals,
	hcloudProvider *hcloud.Provider,
) error {
	domainNames := pulumi.ToStringArray(config.DomainNames)

	created, err := hcloud.NewManagedCertificate(
		ctx,
		"certificate",
		&hcloud.ManagedCertificateArgs{
			Name:        pulumi.String(name),
			DomainNames: domainNames,
			Labels:      pulumi.ToStringMap(locals.Labels),
		},
		pulumi.Provider(hcloudProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create managed certificate")
	}

	ctx.Export(OpCertificateId, created.ID())
	ctx.Export(OpType, created.Type)
	ctx.Export(OpFingerprint, created.Fingerprint)
	ctx.Export(OpNotValidBefore, created.NotValidBefore)
	ctx.Export(OpNotValidAfter, created.NotValidAfter)

	return nil
}
