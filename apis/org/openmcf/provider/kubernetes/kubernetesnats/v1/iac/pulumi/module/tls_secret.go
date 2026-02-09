package module

import (
	"time"

	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// tlsSecret creates a self-signed certificate and stores it in a Kubernetes
// Secret named "tls-<namespace>".  We export <name> + "tls.crt" key via
// OpTlsSecretName / OpTlsSecretKey so downstream jobs can mount or copy it.
//
// Simplicity > PKI perfection: the cert is valid for 5 years and uses a
// single CN = "<namespace>.svc".  If the user needs full ACME / cert-manager
// workflow, they can fork the module later.
func tlsSecret(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	if !locals.KubernetesNats.Spec.TlsEnabled {
		return nil
	}

	// --------------------- private key --------------------------------------
	createdPrivateKey, err := tls.NewPrivateKey(ctx,
		"tls-key",
		&tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(2048),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create private key")
	}

	// -------------------- self-signed cert ----------------------------------
	subject := tls.SelfSignedCertSubjectArgs{
		CommonName:   pulumi.String(locals.Namespace + ".svc"),
		Organization: pulumi.String("OpenMCF"),
	}

	createdCert, err := tls.NewSelfSignedCert(ctx,
		"tls-cert",
		&tls.SelfSignedCertArgs{
			KeyAlgorithm:        createdPrivateKey.Algorithm,
			PrivateKeyPem:       createdPrivateKey.PrivateKeyPem,
			Subject:             subject,
			ValidityPeriodHours: pulumi.Int(int(time.Hour*24*365*5) / int(time.Hour)),
			EarlyRenewalHours:   pulumi.Int(720), // 30 days before expiry
			IsCaCertificate:     pulumi.Bool(false),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create self-signed certificate")
	}

	// ----------------------- Kubernetes Secret ------------------------------
	tlsOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err = kubernetescorev1.NewSecret(ctx,
		"tls-secret",
		&kubernetescorev1.SecretArgs{
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(locals.TlsSecretName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			StringData: pulumi.StringMap{
				"tls.crt": createdCert.CertPem,
				"tls.key": createdPrivateKey.PrivateKeyPem,
			},
			Type: pulumi.String("kubernetes.io/tls"),
		}, tlsOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create TLS secret")
	}

	return nil
}
