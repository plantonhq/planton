package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// originCaCertificate issues a Cloudflare Origin CA certificate. When the spec
// omits a CSR, the module generates a private key (keyed to request_type) and a
// CSR for the requested hostnames, then exports the generated key as a sensitive
// output. When a CSR is supplied, the user's key never leaves their control and
// the private_key output is empty. Defaults are coalesced here so a standalone
// module run matches the control-plane middleware byte-for-byte.
func originCaCertificate(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) error {
	spec := locals.CloudflareOriginCaCertificate.Spec

	requestType := spec.GetRequestType()
	if requestType == "" {
		requestType = "origin-rsa"
	}
	requestedValidity := spec.GetRequestedValidity()
	if requestedValidity == 0 {
		requestedValidity = 5475
	}

	hostnames := make(pulumi.StringArray, 0, len(spec.Hostnames))
	for _, h := range spec.Hostnames {
		hostnames = append(hostnames, pulumi.String(h))
	}

	var csrPem pulumi.StringInput
	var generatedPrivateKey *tls.PrivateKey

	if spec.Csr != "" {
		csrPem = pulumi.String(spec.Csr)
	} else {
		keyArgs := &tls.PrivateKeyArgs{}
		if requestType == "origin-ecc" {
			keyArgs.Algorithm = pulumi.String("ECDSA")
			keyArgs.EcdsaCurve = pulumi.String("P256")
		} else {
			keyArgs.Algorithm = pulumi.String("RSA")
			keyArgs.RsaBits = pulumi.Int(2048)
		}

		key, err := tls.NewPrivateKey(ctx, "origin-key", keyArgs)
		if err != nil {
			return errors.Wrap(err, "failed to generate origin private key")
		}
		generatedPrivateKey = key

		csr, err := tls.NewCertRequest(ctx, "origin-csr", &tls.CertRequestArgs{
			PrivateKeyPem: key.PrivateKeyPem,
			Subject: &tls.CertRequestSubjectArgs{
				CommonName: pulumi.String(spec.Hostnames[0]),
			},
			DnsNames: hostnames,
		})
		if err != nil {
			return errors.Wrap(err, "failed to generate origin CSR")
		}
		csrPem = csr.CertRequestPem
	}

	created, err := cloudflare.NewOriginCaCertificate(
		ctx,
		"origin-ca-certificate",
		&cloudflare.OriginCaCertificateArgs{
			Csr:               csrPem,
			Hostnames:         hostnames,
			RequestType:       pulumi.String(requestType),
			RequestedValidity: pulumi.Float64Ptr(float64(requestedValidity)),
		},
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudflare origin ca certificate")
	}

	ctx.Export(OpCertificateId, created.ID())
	ctx.Export(OpCertificate, created.Certificate)
	ctx.Export(OpExpiresOn, created.ExpiresOn)
	if generatedPrivateKey != nil {
		ctx.Export(OpPrivateKey, pulumi.ToSecret(generatedPrivateKey.PrivateKeyPem))
	} else {
		ctx.Export(OpPrivateKey, pulumi.String(""))
	}

	return nil
}
