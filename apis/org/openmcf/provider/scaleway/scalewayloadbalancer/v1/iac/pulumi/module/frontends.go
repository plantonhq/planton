package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/loadbalancers"
)

// frontends creates all frontend listeners defined in the spec.
//
// Each frontend listens on a port and routes traffic to a backend (resolved
// by name from backendMap) with optional TLS certificates (resolved by name
// from certMap).
func frontends(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scaleway.Provider,
	lb *loadbalancers.LoadBalancer,
	backendMap map[string]*loadbalancers.Backend,
	certMap map[string]*loadbalancers.Certificate,
) error {
	for _, frontendSpec := range locals.ScalewayLoadBalancer.Spec.Frontends {
		// Resolve backend reference by name.
		backend, ok := backendMap[frontendSpec.BackendName]
		if !ok {
			return fmt.Errorf(
				"frontend %q references backend %q which does not exist in spec.backends",
				frontendSpec.Name, frontendSpec.BackendName,
			)
		}

		args := &loadbalancers.FrontendArgs{
			LbId:        lb.ID(),
			Name:        pulumi.String(frontendSpec.Name),
			InboundPort: pulumi.Int(int(frontendSpec.InboundPort)),
			BackendId:   backend.ID(),
		}

		// Resolve certificate references by name.
		if len(frontendSpec.CertificateNames) > 0 {
			certIds := make(pulumi.StringArray, 0, len(frontendSpec.CertificateNames))
			for _, certName := range frontendSpec.CertificateNames {
				cert, certOk := certMap[certName]
				if !certOk {
					return fmt.Errorf(
						"frontend %q references certificate %q which does not exist in spec.certificates",
						frontendSpec.Name, certName,
					)
				}
				certIds = append(certIds, cert.ID())
			}
			args.CertificateIds = certIds
		}

		// Client timeout.
		if frontendSpec.TimeoutClient != "" {
			args.TimeoutClient = pulumi.String(frontendSpec.TimeoutClient)
		}

		// HTTP/3 support.
		if frontendSpec.EnableHttp3 {
			args.EnableHttp3 = pulumi.Bool(true)
		}

		resourceName := fmt.Sprintf("frontend-%s", frontendSpec.Name)
		_, err := loadbalancers.NewFrontend(
			ctx,
			resourceName,
			args,
			pulumi.Provider(scalewayProvider),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create frontend %q", frontendSpec.Name)
		}
	}

	return nil
}

// certificates creates all TLS certificates defined in the spec.
//
// Each certificate is either a Let's Encrypt auto-provisioned cert or a
// custom PEM chain. The returned map is keyed by certificate name for
// frontend→certificate resolution.
func certificates(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scaleway.Provider,
	lb *loadbalancers.LoadBalancer,
) (map[string]*loadbalancers.Certificate, error) {
	certMap := make(map[string]*loadbalancers.Certificate, len(locals.ScalewayLoadBalancer.Spec.Certificates))

	for _, certSpec := range locals.ScalewayLoadBalancer.Spec.Certificates {
		args := &loadbalancers.CertificateArgs{
			LbId: lb.ID(),
			Name: pulumi.String(certSpec.Name),
		}

		if certSpec.Letsencrypt != nil {
			leArgs := &loadbalancers.CertificateLetsencryptArgs{
				CommonName: pulumi.String(certSpec.Letsencrypt.CommonName),
			}
			if len(certSpec.Letsencrypt.SubjectAlternativeNames) > 0 {
				leArgs.SubjectAlternativeNames = pulumi.ToStringArray(certSpec.Letsencrypt.SubjectAlternativeNames)
			}
			args.Letsencrypt = leArgs
		} else if certSpec.CustomCertificate != nil {
			args.CustomCertificate = &loadbalancers.CertificateCustomCertificateArgs{
				CertificateChain: pulumi.String(certSpec.CustomCertificate.CertificateChain),
			}
		}

		resourceName := fmt.Sprintf("cert-%s", certSpec.Name)
		createdCert, err := loadbalancers.NewCertificate(
			ctx,
			resourceName,
			args,
			pulumi.Provider(scalewayProvider),
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create certificate %q", certSpec.Name)
		}

		certMap[certSpec.Name] = createdCert
	}

	return certMap, nil
}
