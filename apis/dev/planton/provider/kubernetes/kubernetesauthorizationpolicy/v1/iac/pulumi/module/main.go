package module

import (
	"github.com/pkg/errors"
	kubernetesauthorizationpolicyv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesauthorizationpolicy/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	istiosecurityv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/istio/kubernetes/security/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesauthorizationpolicyv1.KubernetesAuthorizationPolicyStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createAuthorizationPolicy(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create authorization policy")
	}

	ctx.Export(OpAuthorizationPolicyName, pulumi.String(locals.AuthorizationPolicyName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return nil
}

// createAuthorizationPolicy creates the namespaced Istio AuthorizationPolicy using
// the typed crd2pulumi SDK (istiosecurityv1.NewAuthorizationPolicy), consistent with
// every other Planton Istio component. The typed approach catches field-name and
// structure errors at compile time. Each optional upstream block is only attached
// when present, so unset fields fall through to istiod's defaults (e.g. an absent
// `action` becomes the upstream default ALLOW).
func createAuthorizationPolicy(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesAuthorizationPolicy.Spec

	// The typed resource's Spec field is a PtrInput satisfied by the Args value
	// itself (not the SpecPtr() wrapper, which marshals to the wrong element
	// type); assigned directly below, mirroring the sibling Istio components.
	apSpec := istiosecurityv1.AuthorizationPolicySpecArgs{}

	if selector := spec.GetSelector(); selector != nil && len(selector.GetMatchLabels()) > 0 {
		apSpec.Selector = istiosecurityv1.AuthorizationPolicySpecSelectorArgs{
			MatchLabels: pulumi.ToStringMap(selector.GetMatchLabels()),
		}
	}

	if targetRefs := spec.GetTargetRefs(); len(targetRefs) > 0 {
		refs := istiosecurityv1.AuthorizationPolicySpecTargetRefsArray{}
		for _, ref := range targetRefs {
			refArgs := istiosecurityv1.AuthorizationPolicySpecTargetRefsArgs{
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
		apSpec.TargetRefs = refs
	}

	if rules := spec.GetRules(); len(rules) > 0 {
		ruleArgs := istiosecurityv1.AuthorizationPolicySpecRulesArray{}
		for _, r := range rules {
			ruleArgs = append(ruleArgs, buildRuleArgs(r))
		}
		apSpec.Rules = ruleArgs
	}

	if spec.Action != nil {
		apSpec.Action = pulumi.String(spec.GetAction())
	}

	if provider := spec.GetProvider(); provider != nil {
		apSpec.Provider = istiosecurityv1.AuthorizationPolicySpecProviderArgs{
			Name: pulumi.String(provider.GetName()),
		}
	}

	_, err := istiosecurityv1.NewAuthorizationPolicy(ctx, locals.AuthorizationPolicyName,
		&istiosecurityv1.AuthorizationPolicyArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.AuthorizationPolicyName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: apSpec,
		},
		pulumi.Provider(kubeProvider))

	return err
}

// buildRuleArgs maps one Planton rule to the typed SDK args, attaching the from/to/
// when blocks only when present so empty lists are omitted from the CR.
func buildRuleArgs(r *kubernetesauthorizationpolicyv1.KubernetesAuthorizationPolicyRule) istiosecurityv1.AuthorizationPolicySpecRulesArgs {
	args := istiosecurityv1.AuthorizationPolicySpecRulesArgs{}

	if from := r.GetFrom(); len(from) > 0 {
		fromArgs := istiosecurityv1.AuthorizationPolicySpecRulesFromArray{}
		for _, f := range from {
			fromArgs = append(fromArgs, istiosecurityv1.AuthorizationPolicySpecRulesFromArgs{
				Source: buildSourceArgs(f.GetSource()),
			})
		}
		args.From = fromArgs
	}

	if to := r.GetTo(); len(to) > 0 {
		toArgs := istiosecurityv1.AuthorizationPolicySpecRulesToArray{}
		for _, t := range to {
			toArgs = append(toArgs, istiosecurityv1.AuthorizationPolicySpecRulesToArgs{
				Operation: buildOperationArgs(t.GetOperation()),
			})
		}
		args.To = toArgs
	}

	if when := r.GetWhen(); len(when) > 0 {
		whenArgs := istiosecurityv1.AuthorizationPolicySpecRulesWhenArray{}
		for _, c := range when {
			condArgs := istiosecurityv1.AuthorizationPolicySpecRulesWhenArgs{
				Key: pulumi.String(c.GetKey()),
			}
			if len(c.GetValues()) > 0 {
				condArgs.Values = pulumi.ToStringArray(c.GetValues())
			}
			if len(c.GetNotValues()) > 0 {
				condArgs.NotValues = pulumi.ToStringArray(c.GetNotValues())
			}
			whenArgs = append(whenArgs, condArgs)
		}
		args.When = whenArgs
	}

	return args
}

// buildSourceArgs maps an Planton source to the typed SDK args. Each identity list
// is attached only when non-empty, so the CR carries exactly what the user set.
func buildSourceArgs(source *kubernetesauthorizationpolicyv1.KubernetesAuthorizationPolicySource) istiosecurityv1.AuthorizationPolicySpecRulesFromSourceArgs {
	args := istiosecurityv1.AuthorizationPolicySpecRulesFromSourceArgs{}
	if source == nil {
		return args
	}
	if len(source.GetPrincipals()) > 0 {
		args.Principals = pulumi.ToStringArray(source.GetPrincipals())
	}
	if len(source.GetNotPrincipals()) > 0 {
		args.NotPrincipals = pulumi.ToStringArray(source.GetNotPrincipals())
	}
	if len(source.GetRequestPrincipals()) > 0 {
		args.RequestPrincipals = pulumi.ToStringArray(source.GetRequestPrincipals())
	}
	if len(source.GetNotRequestPrincipals()) > 0 {
		args.NotRequestPrincipals = pulumi.ToStringArray(source.GetNotRequestPrincipals())
	}
	if len(source.GetNamespaces()) > 0 {
		args.Namespaces = pulumi.ToStringArray(source.GetNamespaces())
	}
	if len(source.GetNotNamespaces()) > 0 {
		args.NotNamespaces = pulumi.ToStringArray(source.GetNotNamespaces())
	}
	if len(source.GetServiceAccounts()) > 0 {
		args.ServiceAccounts = pulumi.ToStringArray(source.GetServiceAccounts())
	}
	if len(source.GetNotServiceAccounts()) > 0 {
		args.NotServiceAccounts = pulumi.ToStringArray(source.GetNotServiceAccounts())
	}
	if len(source.GetIpBlocks()) > 0 {
		args.IpBlocks = pulumi.ToStringArray(source.GetIpBlocks())
	}
	if len(source.GetNotIpBlocks()) > 0 {
		args.NotIpBlocks = pulumi.ToStringArray(source.GetNotIpBlocks())
	}
	if len(source.GetRemoteIpBlocks()) > 0 {
		args.RemoteIpBlocks = pulumi.ToStringArray(source.GetRemoteIpBlocks())
	}
	if len(source.GetNotRemoteIpBlocks()) > 0 {
		args.NotRemoteIpBlocks = pulumi.ToStringArray(source.GetNotRemoteIpBlocks())
	}
	return args
}

// buildOperationArgs maps an Planton operation to the typed SDK args. Each match
// list is attached only when non-empty.
func buildOperationArgs(operation *kubernetesauthorizationpolicyv1.KubernetesAuthorizationPolicyOperation) istiosecurityv1.AuthorizationPolicySpecRulesToOperationArgs {
	args := istiosecurityv1.AuthorizationPolicySpecRulesToOperationArgs{}
	if operation == nil {
		return args
	}
	if len(operation.GetHosts()) > 0 {
		args.Hosts = pulumi.ToStringArray(operation.GetHosts())
	}
	if len(operation.GetNotHosts()) > 0 {
		args.NotHosts = pulumi.ToStringArray(operation.GetNotHosts())
	}
	if len(operation.GetPorts()) > 0 {
		args.Ports = pulumi.ToStringArray(operation.GetPorts())
	}
	if len(operation.GetNotPorts()) > 0 {
		args.NotPorts = pulumi.ToStringArray(operation.GetNotPorts())
	}
	if len(operation.GetMethods()) > 0 {
		args.Methods = pulumi.ToStringArray(operation.GetMethods())
	}
	if len(operation.GetNotMethods()) > 0 {
		args.NotMethods = pulumi.ToStringArray(operation.GetNotMethods())
	}
	if len(operation.GetPaths()) > 0 {
		args.Paths = pulumi.ToStringArray(operation.GetPaths())
	}
	if len(operation.GetNotPaths()) > 0 {
		args.NotPaths = pulumi.ToStringArray(operation.GetNotPaths())
	}
	return args
}
