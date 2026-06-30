package module

import (
	"fmt"

	"github.com/plantonhq/planton/internal/valuefrom"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// alb creates an AWS Application Load Balancer with listeners.
// If SSL is enabled: HTTP (80) redirects to HTTPS (443) with certificate.
// If SSL is disabled: HTTP (80) listener only.
func alb(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*lb.LoadBalancer, error) {
	createdLoadBalancer, err := lb.NewLoadBalancer(ctx, locals.AwsAlb.Metadata.Name, &lb.LoadBalancerArgs{
		Name:                     pulumi.String(locals.AwsAlb.Metadata.Name),
		LoadBalancerType:         pulumi.String("application"),
		SecurityGroups:           pulumi.ToStringArray(valuefrom.ToStringArray(locals.AwsAlb.Spec.SecurityGroups)),
		Subnets:                  pulumi.ToStringArray(valuefrom.ToStringArray(locals.AwsAlb.Spec.Subnets)),
		Internal:                 pulumi.Bool(locals.AwsAlb.Spec.Internal),
		IpAddressType:            pulumi.String("ipv4"),
		EnableDeletionProtection: pulumi.Bool(locals.AwsAlb.Spec.DeleteProtectionEnabled),
		IdleTimeout:              pulumi.Int(int(locals.AwsAlb.Spec.IdleTimeoutSeconds)),
		Tags:                     pulumi.ToStringMap(locals.AwsTags),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create AWS ALB")
	}

	// Create listeners based on SSL configuration
	if locals.AwsAlb.Spec.Ssl != nil && locals.AwsAlb.Spec.Ssl.Enabled {
		// SSL is enabled - create HTTP->HTTPS redirect and HTTPS listener
		if locals.AwsAlb.Spec.Ssl.CertificateArn == nil || locals.AwsAlb.Spec.Ssl.CertificateArn.GetValue() == "" {
			return nil, fmt.Errorf("ssl.enabled is true, but ssl.certificate_arn is not provided")
		}

		// HTTP listener on port 80 - redirect to HTTPS
		httpListenerName := fmt.Sprintf("%s-listener-80", locals.AwsAlb.Metadata.Name)
		_, err = lb.NewListener(ctx, httpListenerName, &lb.ListenerArgs{
			LoadBalancerArn: createdLoadBalancer.Arn,
			Port:            pulumi.Int(80),
			Protocol:        pulumi.String("HTTP"),
			DefaultActions: lb.ListenerDefaultActionArray{
				&lb.ListenerDefaultActionArgs{
					Type: pulumi.String("redirect"),
					Redirect: &lb.ListenerDefaultActionRedirectArgs{
						Port:       pulumi.String("443"),
						Protocol:   pulumi.String("HTTPS"),
						StatusCode: pulumi.String("HTTP_301"),
					},
				},
			},
		}, pulumi.Provider(provider))
		if err != nil {
			return nil, errors.Wrap(err, "unable to create HTTP->HTTPS redirect listener")
		}

		// HTTPS listener on port 443 with certificate
		httpsListenerName := fmt.Sprintf("%s-listener-443", locals.AwsAlb.Metadata.Name)
		_, err = lb.NewListener(ctx, httpsListenerName, &lb.ListenerArgs{
			LoadBalancerArn: createdLoadBalancer.Arn,
			Port:            pulumi.Int(443),
			Protocol:        pulumi.String("HTTPS"),
			CertificateArn:  pulumi.String(locals.AwsAlb.Spec.Ssl.CertificateArn.GetValue()),
			SslPolicy:       pulumi.String("ELBSecurityPolicy-2016-08"),
			DefaultActions: lb.ListenerDefaultActionArray{
				&lb.ListenerDefaultActionArgs{
					Type: pulumi.String("fixed-response"),
					FixedResponse: &lb.ListenerDefaultActionFixedResponseArgs{
						ContentType: pulumi.String("text/plain"),
						StatusCode:  pulumi.String("200"),
						MessageBody: pulumi.String("OK"),
					},
				},
			},
		}, pulumi.Provider(provider))
		if err != nil {
			return nil, errors.Wrap(err, "unable to create HTTPS listener")
		}
	} else {
		// SSL is not enabled - create simple HTTP listener on port 80
		httpListenerName := fmt.Sprintf("%s-listener-80", locals.AwsAlb.Metadata.Name)
		_, err = lb.NewListener(ctx, httpListenerName, &lb.ListenerArgs{
			LoadBalancerArn: createdLoadBalancer.Arn,
			Port:            pulumi.Int(80),
			Protocol:        pulumi.String("HTTP"),
			DefaultActions: lb.ListenerDefaultActionArray{
				&lb.ListenerDefaultActionArgs{
					Type: pulumi.String("fixed-response"),
					FixedResponse: &lb.ListenerDefaultActionFixedResponseArgs{
						ContentType: pulumi.String("text/plain"),
						StatusCode:  pulumi.String("200"),
						MessageBody: pulumi.String("OK"),
					},
				},
			},
		}, pulumi.Provider(provider))
		if err != nil {
			return nil, errors.Wrap(err, "unable to create HTTP listener")
		}
	}

	// Export key ALB outputs
	ctx.Export(OpAlbArn, createdLoadBalancer.Arn)
	ctx.Export(OpAlbName, createdLoadBalancer.Name)
	ctx.Export(OpAlbDnsName, createdLoadBalancer.DnsName)
	ctx.Export(OpAlbHostedZoneId, createdLoadBalancer.ZoneId)

	return createdLoadBalancer, nil
}
