package module

import (
	"github.com/pkg/errors"
	kubernetesrequestauthenticationv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesrequestauthentication/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	istiosecurityv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/istio/kubernetes/security/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesrequestauthenticationv1.KubernetesRequestAuthenticationStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createRequestAuthentication(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create request authentication")
	}

	ctx.Export(OpRequestAuthenticationName, pulumi.String(locals.RequestAuthenticationName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return nil
}

// createRequestAuthentication creates the namespaced Istio RequestAuthentication
// using the typed crd2pulumi SDK (istiosecurityv1.NewRequestAuthentication),
// consistent with every other Planton Istio component. The typed approach catches
// field-name and structure errors at compile time. Each optional upstream block is
// only attached when present, so unset fields fall through to istiod's defaults.
func createRequestAuthentication(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesRequestAuthentication.Spec

	// The typed resource's Spec field is a PtrInput satisfied by the Args value
	// itself (not the SpecPtr() wrapper, which marshals to the wrong element
	// type); assigned directly below, mirroring the PeerAuthentication component.
	raSpec := istiosecurityv1.RequestAuthenticationSpecArgs{}

	if selector := spec.GetSelector(); selector != nil && len(selector.GetMatchLabels()) > 0 {
		raSpec.Selector = istiosecurityv1.RequestAuthenticationSpecSelectorArgs{
			MatchLabels: pulumi.ToStringMap(selector.GetMatchLabels()),
		}
	}

	if targetRefs := spec.GetTargetRefs(); len(targetRefs) > 0 {
		refs := istiosecurityv1.RequestAuthenticationSpecTargetRefsArray{}
		for _, ref := range targetRefs {
			refArgs := istiosecurityv1.RequestAuthenticationSpecTargetRefsArgs{
				Kind: pulumi.String(ref.GetKind()),
				Name: pulumi.String(ref.GetName()),
			}
			if ref.GetGroup() != "" {
				refArgs.Group = pulumi.String(ref.GetGroup())
			}
			if ref.GetNamespace() != "" {
				refArgs.Namespace = pulumi.String(ref.GetNamespace())
			}
			refs = append(refs, refArgs)
		}
		raSpec.TargetRefs = refs
	}

	if jwtRules := spec.GetJwtRules(); len(jwtRules) > 0 {
		rules := istiosecurityv1.RequestAuthenticationSpecJwtRulesArray{}
		for _, rule := range jwtRules {
			rules = append(rules, buildJwtRuleArgs(rule))
		}
		raSpec.JwtRules = rules
	}

	_, err := istiosecurityv1.NewRequestAuthentication(ctx, locals.RequestAuthenticationName,
		&istiosecurityv1.RequestAuthenticationArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.RequestAuthenticationName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: raSpec,
		},
		pulumi.Provider(kubeProvider))

	return err
}

// buildJwtRuleArgs maps one Planton JWT rule to the typed SDK args. Optional scalar
// fields are attached only when present in the proto (the proto3 `optional` pointer
// distinguishes unset from empty), so unset fields are omitted from the CR and
// upstream defaults apply.
func buildJwtRuleArgs(rule *kubernetesrequestauthenticationv1.KubernetesRequestAuthenticationJwtRule) istiosecurityv1.RequestAuthenticationSpecJwtRulesArgs {
	args := istiosecurityv1.RequestAuthenticationSpecJwtRulesArgs{
		Issuer: pulumi.String(rule.GetIssuer()),
	}

	if len(rule.GetAudiences()) > 0 {
		args.Audiences = pulumi.ToStringArray(rule.GetAudiences())
	}
	if rule.JwksUri != nil {
		args.JwksUri = pulumi.String(rule.GetJwksUri())
	}
	if rule.Jwks != nil {
		args.Jwks = pulumi.String(rule.GetJwks())
	}
	if len(rule.GetFromHeaders()) > 0 {
		headers := istiosecurityv1.RequestAuthenticationSpecJwtRulesFromHeadersArray{}
		for _, header := range rule.GetFromHeaders() {
			headerArgs := istiosecurityv1.RequestAuthenticationSpecJwtRulesFromHeadersArgs{
				Name: pulumi.String(header.GetName()),
			}
			if header.Prefix != nil {
				headerArgs.Prefix = pulumi.String(header.GetPrefix())
			}
			headers = append(headers, headerArgs)
		}
		args.FromHeaders = headers
	}
	if len(rule.GetFromParams()) > 0 {
		args.FromParams = pulumi.ToStringArray(rule.GetFromParams())
	}
	if len(rule.GetFromCookies()) > 0 {
		args.FromCookies = pulumi.ToStringArray(rule.GetFromCookies())
	}
	if rule.OutputPayloadToHeader != nil {
		args.OutputPayloadToHeader = pulumi.String(rule.GetOutputPayloadToHeader())
	}
	if rule.ForwardOriginalToken != nil {
		args.ForwardOriginalToken = pulumi.Bool(rule.GetForwardOriginalToken())
	}
	if len(rule.GetOutputClaimToHeaders()) > 0 {
		claims := istiosecurityv1.RequestAuthenticationSpecJwtRulesOutputClaimToHeadersArray{}
		for _, claim := range rule.GetOutputClaimToHeaders() {
			claims = append(claims, istiosecurityv1.RequestAuthenticationSpecJwtRulesOutputClaimToHeadersArgs{
				Header: pulumi.String(claim.GetHeader()),
				Claim:  pulumi.String(claim.GetClaim()),
			})
		}
		args.OutputClaimToHeaders = claims
	}
	if rule.Timeout != nil {
		args.Timeout = pulumi.String(rule.GetTimeout())
	}

	return args
}
