package module

import (
	"fmt"

	"github.com/pkg/errors"
	awslblistenerv1alpha1 "github.com/plantonhq/planton/catalog/aws/awslblistener/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// listener provisions the listener, its default-action chain, and any
// additional SNI certificates. The load balancer is create-only (moving a
// listener replaces it); port, protocol, certificates, and actions update in
// place.
func listener(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource) error {
	spec := locals.AwsLbListener.Spec
	listenerName := locals.AwsLbListener.Metadata.Name

	isTlsProtocol := spec.Protocol == "HTTPS" || spec.Protocol == "TLS"

	// The proto cannot enforce certificate requiredness per protocol
	// (message-level CEL on StringValueOrRef fields breaks protovalidate-java),
	// so the module is the enforcement point: fail fast with a clear message
	// instead of letting AWS reject the listener at apply time.
	if isTlsProtocol && spec.CertificateArn.GetValue() == "" {
		return errors.Errorf("certificate_arn is required when protocol is %q", spec.Protocol)
	}

	args := &lb.ListenerArgs{
		LoadBalancerArn: pulumi.String(spec.LoadBalancerArn.GetValue()),
		Port:            pulumi.IntPtr(int(spec.Port)),
		Protocol:        pulumi.StringPtr(spec.Protocol),
		Tags: convertstringmaps.ConvertGoStringMapToPulumiStringMap(
			stringmaps.AddEntry(locals.AwsTags, "Name", listenerName)),
	}

	if isTlsProtocol {
		args.CertificateArn = pulumi.StringPtr(spec.CertificateArn.GetValue())
	}
	if spec.SslPolicy != "" {
		args.SslPolicy = pulumi.StringPtr(spec.SslPolicy)
	}
	if spec.AlpnPolicy != "" {
		args.AlpnPolicy = pulumi.StringPtr(spec.AlpnPolicy)
	}
	if spec.TcpIdleTimeoutSeconds > 0 {
		args.TcpIdleTimeoutSeconds = pulumi.IntPtr(int(spec.TcpIdleTimeoutSeconds))
	}

	if spec.MutualAuthentication != nil {
		mtls := &lb.ListenerMutualAuthenticationArgs{
			Mode: pulumi.String(spec.MutualAuthentication.Mode),
		}
		if spec.MutualAuthentication.TrustStoreArn.GetValue() != "" {
			mtls.TrustStoreArn = pulumi.StringPtr(spec.MutualAuthentication.TrustStoreArn.GetValue())
		}
		if spec.MutualAuthentication.IgnoreClientCertificateExpiry {
			mtls.IgnoreClientCertificateExpiry = pulumi.BoolPtr(true)
		}
		if spec.MutualAuthentication.AdvertiseTrustStoreCaNames != "" {
			mtls.AdvertiseTrustStoreCaNames = pulumi.StringPtr(spec.MutualAuthentication.AdvertiseTrustStoreCaNames)
		}
		args.MutualAuthentication = mtls
	}

	applyHttpHeaders(args, spec.HttpHeaders)

	defaultActions, err := defaultActionArgs(spec.DefaultActions)
	if err != nil {
		return err
	}
	args.DefaultActions = defaultActions

	createdListener, err := lb.NewListener(ctx, listenerName, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create listener")
	}

	// Additional SNI certificates. AWS models each as a separate
	// listener-certificate attachment; they are folded into this kind because
	// an attachment is pure glue with no referenceable identity. Keyed by
	// index because a certificate may be a reference resolved at apply time.
	for i, certificate := range spec.AdditionalCertificateArns {
		if _, err := lb.NewListenerCertificate(ctx,
			fmt.Sprintf("%s-certificate-%d", listenerName, i),
			&lb.ListenerCertificateArgs{
				ListenerArn:    createdListener.Arn,
				CertificateArn: pulumi.String(certificate.GetValue()),
			}, pulumi.Provider(provider)); err != nil {
			return errors.Wrapf(err, "failed to attach additional certificate %d", i)
		}
	}

	ctx.Export(OpListenerArn, createdListener.Arn)

	return nil
}

// applyHttpHeaders maps the spec's HTTP header handling onto the provider's
// flat routing_http_* attributes. The proto groups them into request/response
// messages purely for readability; AWS models them as listener attributes.
func applyHttpHeaders(args *lb.ListenerArgs, headers *awslblistenerv1alpha1.AwsLbListenerHttpHeaders) {
	if headers == nil {
		return
	}

	if request := headers.Request; request != nil {
		if request.MtlsClientcertHeaderName != "" {
			args.RoutingHttpRequestXAmznMtlsClientcertHeaderName = pulumi.StringPtr(request.MtlsClientcertHeaderName)
		}
		if request.MtlsClientcertIssuerHeaderName != "" {
			args.RoutingHttpRequestXAmznMtlsClientcertIssuerHeaderName = pulumi.StringPtr(request.MtlsClientcertIssuerHeaderName)
		}
		if request.MtlsClientcertLeafHeaderName != "" {
			args.RoutingHttpRequestXAmznMtlsClientcertLeafHeaderName = pulumi.StringPtr(request.MtlsClientcertLeafHeaderName)
		}
		if request.MtlsClientcertSerialNumberHeaderName != "" {
			args.RoutingHttpRequestXAmznMtlsClientcertSerialNumberHeaderName = pulumi.StringPtr(request.MtlsClientcertSerialNumberHeaderName)
		}
		if request.MtlsClientcertSubjectHeaderName != "" {
			args.RoutingHttpRequestXAmznMtlsClientcertSubjectHeaderName = pulumi.StringPtr(request.MtlsClientcertSubjectHeaderName)
		}
		if request.MtlsClientcertValidityHeaderName != "" {
			args.RoutingHttpRequestXAmznMtlsClientcertValidityHeaderName = pulumi.StringPtr(request.MtlsClientcertValidityHeaderName)
		}
		if request.TlsCipherSuiteHeaderName != "" {
			args.RoutingHttpRequestXAmznTlsCipherSuiteHeaderName = pulumi.StringPtr(request.TlsCipherSuiteHeaderName)
		}
		if request.TlsVersionHeaderName != "" {
			args.RoutingHttpRequestXAmznTlsVersionHeaderName = pulumi.StringPtr(request.TlsVersionHeaderName)
		}
	}

	if response := headers.Response; response != nil {
		if response.AccessControlAllowCredentials != "" {
			args.RoutingHttpResponseAccessControlAllowCredentialsHeaderValue = pulumi.StringPtr(response.AccessControlAllowCredentials)
		}
		if response.AccessControlAllowHeaders != "" {
			args.RoutingHttpResponseAccessControlAllowHeadersHeaderValue = pulumi.StringPtr(response.AccessControlAllowHeaders)
		}
		if response.AccessControlAllowMethods != "" {
			args.RoutingHttpResponseAccessControlAllowMethodsHeaderValue = pulumi.StringPtr(response.AccessControlAllowMethods)
		}
		if response.AccessControlAllowOrigin != "" {
			args.RoutingHttpResponseAccessControlAllowOriginHeaderValue = pulumi.StringPtr(response.AccessControlAllowOrigin)
		}
		if response.AccessControlExposeHeaders != "" {
			args.RoutingHttpResponseAccessControlExposeHeadersHeaderValue = pulumi.StringPtr(response.AccessControlExposeHeaders)
		}
		if response.AccessControlMaxAge != "" {
			args.RoutingHttpResponseAccessControlMaxAgeHeaderValue = pulumi.StringPtr(response.AccessControlMaxAge)
		}
		if response.ContentSecurityPolicy != "" {
			args.RoutingHttpResponseContentSecurityPolicyHeaderValue = pulumi.StringPtr(response.ContentSecurityPolicy)
		}
		if response.ServerEnabled != nil {
			args.RoutingHttpResponseServerEnabled = pulumi.BoolPtr(response.GetServerEnabled())
		}
		if response.StrictTransportSecurity != "" {
			args.RoutingHttpResponseStrictTransportSecurityHeaderValue = pulumi.StringPtr(response.StrictTransportSecurity)
		}
		if response.XContentTypeOptions != "" {
			args.RoutingHttpResponseXContentTypeOptionsHeaderValue = pulumi.StringPtr(response.XContentTypeOptions)
		}
		if response.XFrameOptions != "" {
			args.RoutingHttpResponseXFrameOptionsHeaderValue = pulumi.StringPtr(response.XFrameOptions)
		}
	}
}

// defaultActionArgs maps the spec's action chain onto provider args. The spec
// already guarantees (via CEL) that exactly one configuration message matches
// each action's type, so this is a pure translation.
func defaultActionArgs(actions []*awslblistenerv1alpha1.AwsLbListenerAction) (lb.ListenerDefaultActionArray, error) {
	result := make(lb.ListenerDefaultActionArray, 0, len(actions))

	for i, action := range actions {
		args := &lb.ListenerDefaultActionArgs{
			Type: pulumi.String(action.Type),
		}
		if action.Order > 0 {
			args.Order = pulumi.IntPtr(int(action.Order))
		}

		switch action.Type {
		case "forward":
			forward := action.Forward
			// A single unweighted target group uses the simple target_group_arn
			// form; AWS treats the weighted forward block and the simple ARN as
			// different configurations, and the simple form avoids spurious
			// diffs on the common case.
			if len(forward.TargetGroups) == 1 && forward.Stickiness == nil && forward.TargetGroups[0].Weight == 0 {
				args.TargetGroupArn = pulumi.StringPtr(forward.TargetGroups[0].Arn.GetValue())
				break
			}
			forwardArgs := &lb.ListenerDefaultActionForwardArgs{}
			targetGroups := make(lb.ListenerDefaultActionForwardTargetGroupArray, 0, len(forward.TargetGroups))
			for _, targetGroup := range forward.TargetGroups {
				targetGroupArgs := &lb.ListenerDefaultActionForwardTargetGroupArgs{
					Arn: pulumi.String(targetGroup.Arn.GetValue()),
				}
				if targetGroup.Weight > 0 {
					targetGroupArgs.Weight = pulumi.IntPtr(int(targetGroup.Weight))
				}
				targetGroups = append(targetGroups, targetGroupArgs)
			}
			forwardArgs.TargetGroups = targetGroups
			// Stickiness is written only when enabled: AWS rejects a duration
			// of 0, and a forward block without stickiness is stickiness off.
			if forward.Stickiness.GetEnabled() {
				forwardArgs.Stickiness = &lb.ListenerDefaultActionForwardStickinessArgs{
					Enabled:  pulumi.BoolPtr(forward.Stickiness.Enabled),
					Duration: pulumi.Int(int(forward.Stickiness.DurationSeconds)),
				}
			}
			args.Forward = forwardArgs

		case "redirect":
			redirect := action.Redirect
			redirectArgs := &lb.ListenerDefaultActionRedirectArgs{
				StatusCode: pulumi.String(redirect.StatusCode),
			}
			if redirect.Protocol != "" {
				redirectArgs.Protocol = pulumi.StringPtr(redirect.Protocol)
			}
			if redirect.Port != "" {
				redirectArgs.Port = pulumi.StringPtr(redirect.Port)
			}
			if redirect.Host != "" {
				redirectArgs.Host = pulumi.StringPtr(redirect.Host)
			}
			if redirect.Path != "" {
				redirectArgs.Path = pulumi.StringPtr(redirect.Path)
			}
			if redirect.Query != "" {
				redirectArgs.Query = pulumi.StringPtr(redirect.Query)
			}
			args.Redirect = redirectArgs

		case "fixed-response":
			fixedResponse := action.FixedResponse
			fixedResponseArgs := &lb.ListenerDefaultActionFixedResponseArgs{
				ContentType: pulumi.String(fixedResponse.ContentType),
			}
			if fixedResponse.StatusCode != "" {
				fixedResponseArgs.StatusCode = pulumi.StringPtr(fixedResponse.StatusCode)
			}
			if fixedResponse.MessageBody != "" {
				fixedResponseArgs.MessageBody = pulumi.StringPtr(fixedResponse.MessageBody)
			}
			args.FixedResponse = fixedResponseArgs

		case "authenticate-cognito":
			cognito := action.AuthenticateCognito
			cognitoArgs := &lb.ListenerDefaultActionAuthenticateCognitoArgs{
				UserPoolArn:      pulumi.String(cognito.UserPoolArn.GetValue()),
				UserPoolClientId: pulumi.String(cognito.UserPoolClientId.GetValue()),
				UserPoolDomain:   pulumi.String(cognito.UserPoolDomain.GetValue()),
			}
			if len(cognito.AuthenticationRequestExtraParams) > 0 {
				cognitoArgs.AuthenticationRequestExtraParams = pulumi.ToStringMap(cognito.AuthenticationRequestExtraParams)
			}
			if cognito.OnUnauthenticatedRequest != "" {
				cognitoArgs.OnUnauthenticatedRequest = pulumi.StringPtr(cognito.OnUnauthenticatedRequest)
			}
			if cognito.Scope != "" {
				cognitoArgs.Scope = pulumi.StringPtr(cognito.Scope)
			}
			if cognito.SessionCookieName != "" {
				cognitoArgs.SessionCookieName = pulumi.StringPtr(cognito.SessionCookieName)
			}
			if cognito.SessionTimeoutSeconds > 0 {
				cognitoArgs.SessionTimeout = pulumi.IntPtr(int(cognito.SessionTimeoutSeconds))
			}
			args.AuthenticateCognito = cognitoArgs

		case "authenticate-oidc":
			oidc := action.AuthenticateOidc
			oidcArgs := &lb.ListenerDefaultActionAuthenticateOidcArgs{
				Issuer:                pulumi.String(oidc.Issuer),
				AuthorizationEndpoint: pulumi.String(oidc.AuthorizationEndpoint),
				TokenEndpoint:         pulumi.String(oidc.TokenEndpoint),
				UserInfoEndpoint:      pulumi.String(oidc.UserInfoEndpoint),
				ClientId:              pulumi.String(oidc.ClientId),
				ClientSecret:          pulumi.String(oidc.ClientSecret),
			}
			if len(oidc.AuthenticationRequestExtraParams) > 0 {
				oidcArgs.AuthenticationRequestExtraParams = pulumi.ToStringMap(oidc.AuthenticationRequestExtraParams)
			}
			if oidc.OnUnauthenticatedRequest != "" {
				oidcArgs.OnUnauthenticatedRequest = pulumi.StringPtr(oidc.OnUnauthenticatedRequest)
			}
			if oidc.Scope != "" {
				oidcArgs.Scope = pulumi.StringPtr(oidc.Scope)
			}
			if oidc.SessionCookieName != "" {
				oidcArgs.SessionCookieName = pulumi.StringPtr(oidc.SessionCookieName)
			}
			if oidc.SessionTimeoutSeconds > 0 {
				oidcArgs.SessionTimeout = pulumi.IntPtr(int(oidc.SessionTimeoutSeconds))
			}
			args.AuthenticateOidc = oidcArgs

		case "jwt-validation":
			jwt := action.JwtValidation
			jwtArgs := &lb.ListenerDefaultActionJwtValidationArgs{
				Issuer:       pulumi.String(jwt.Issuer),
				JwksEndpoint: pulumi.String(jwt.JwksEndpoint),
			}
			claims := make(lb.ListenerDefaultActionJwtValidationAdditionalClaimArray, 0, len(jwt.AdditionalClaims))
			for _, claim := range jwt.AdditionalClaims {
				claims = append(claims, &lb.ListenerDefaultActionJwtValidationAdditionalClaimArgs{
					Name:   pulumi.String(claim.Name),
					Format: pulumi.String(claim.Format),
					Values: pulumi.ToStringArray(claim.Values),
				})
			}
			if len(claims) > 0 {
				jwtArgs.AdditionalClaims = claims
			}
			args.JwtValidation = jwtArgs

		default:
			return nil, errors.Errorf("unsupported action type %q at index %d", action.Type, i)
		}

		result = append(result, args)
	}

	return result, nil
}
