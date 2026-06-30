package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// certManagerCert provisions an ACM certificate with DNS validation
// and creates corresponding CNAME records in Route53.
func certManagerCert(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.AwsCertManagerCert.Spec
	meta := locals.AwsCertManagerCert.Metadata

	// Create the ACM certificate resource
	cert, err := acm.NewCertificate(ctx, meta.Name+"-cert", &acm.CertificateArgs{
		DomainName:              pulumi.String(spec.PrimaryDomainName),
		SubjectAlternativeNames: pulumi.ToStringArray(spec.AlternateDomainNames),
		ValidationMethod:        pulumi.String("DNS"),
		Tags:                    pulumi.ToStringMap(locals.AwsTags),
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create ACM certificate")
	}

	// Create corresponding DNS records, returning a pulumi.StringArray of FQDNs
	recordFqdns := cert.DomainValidationOptions.ApplyT(func(dvos []acm.CertificateDomainValidationOption) pulumi.StringArray {
		// Keep track of distinct DNS record names to avoid duplicates
		distinct := make(map[string]bool)
		var fqdnOutputs pulumi.StringArray
		for i, dvo := range dvos {
			if dvo.ResourceRecordName == nil || dvo.ResourceRecordValue == nil {
				continue
			}
			if distinct[*dvo.ResourceRecordName] {
				continue
			}
			distinct[*dvo.ResourceRecordName] = true

			record, createErr := route53.NewRecord(ctx, fmt.Sprintf("%s-cname-%d", meta.Name, i), &route53.RecordArgs{
				// UPSERT instead of CREATE so a validation record left by a prior
				// partial apply (or shared across certs) is adopted, not collided with.
				AllowOverwrite: pulumi.Bool(true),
				Name:           pulumi.String(*dvo.ResourceRecordName),
				Records: pulumi.StringArray{
					pulumi.String(*dvo.ResourceRecordValue),
				},
				Ttl:    pulumi.Int(300),
				Type:   pulumi.String(*dvo.ResourceRecordType),
				ZoneId: pulumi.String(spec.Route53HostedZoneId.GetValue()),
			}, pulumi.Provider(provider))
			if createErr != nil {
				panic(createErr)
			}
			fqdnOutputs = append(fqdnOutputs, record.Fqdn)
		}
		return fqdnOutputs
	}).(pulumi.StringArrayOutput)

	// Validate the certificate by passing the dynamic FQDNs
	_, err = acm.NewCertificateValidation(ctx, meta.Name+"-validation", &acm.CertificateValidationArgs{
		CertificateArn:        cert.Arn,
		ValidationRecordFqdns: recordFqdns,
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to validate ACM certificate")
	}

	// Export the certificate ARN
	ctx.Export(OpCertArn, cert.Arn)
	ctx.Export(OpCertificateDomainName, pulumi.String(spec.PrimaryDomainName))
	return nil
}
