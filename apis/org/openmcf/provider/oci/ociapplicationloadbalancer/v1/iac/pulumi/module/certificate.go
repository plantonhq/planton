package module

import (
	"fmt"

	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/loadbalancer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createCertificates(ctx *pulumi.Context, locals *Locals, provider *oci.Provider, lb *loadbalancer.LoadBalancer) ([]*loadbalancer.Certificate, error) {
	spec := locals.OciApplicationLoadBalancer.Spec
	var created []*loadbalancer.Certificate

	for _, certSpec := range spec.Certificates {
		args := &loadbalancer.CertificateArgs{
			LoadBalancerId:  lb.ID(),
			CertificateName: pulumi.String(certSpec.CertificateName),
		}

		if certSpec.CaCertificate != "" {
			args.CaCertificate = pulumi.StringPtr(certSpec.CaCertificate)
		}
		if certSpec.PublicCertificate != "" {
			args.PublicCertificate = pulumi.StringPtr(certSpec.PublicCertificate)
		}
		if certSpec.PrivateKey != "" {
			args.PrivateKey = pulumi.StringPtr(certSpec.PrivateKey)
		}
		if certSpec.Passphrase != "" {
			args.Passphrase = pulumi.StringPtr(certSpec.Passphrase)
		}

		cert, err := loadbalancer.NewCertificate(ctx, certSpec.CertificateName, args,
			pulumiOciOpt(provider), pulumi.Parent(lb))
		if err != nil {
			return nil, fmt.Errorf("failed to create certificate %s: %w", certSpec.CertificateName, err)
		}
		created = append(created, cert)
	}

	return created, nil
}
