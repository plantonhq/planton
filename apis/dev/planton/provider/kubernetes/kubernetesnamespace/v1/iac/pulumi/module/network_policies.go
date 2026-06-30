package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	kubernetesnetworkingv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/networking/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createNetworkPolicies creates NetworkPolicy resources for ingress and egress control
func createNetworkPolicies(ctx *pulumi.Context, locals *Locals, namespace *kubernetescorev1.Namespace, provider pulumi.ProviderResource) error {
	// Create ingress isolation policy if enabled
	if locals.NetworkPolicy.IsolateIngress {
		if err := createIngressPolicy(ctx, locals, namespace, provider); err != nil {
			return errors.Wrap(err, "failed to create ingress network policy")
		}
	}

	// Create egress restriction policy if enabled
	if locals.NetworkPolicy.RestrictEgress {
		if err := createEgressPolicy(ctx, locals, namespace, provider); err != nil {
			return errors.Wrap(err, "failed to create egress network policy")
		}
	}

	return nil
}

// createIngressPolicy creates a NetworkPolicy for ingress isolation
func createIngressPolicy(ctx *pulumi.Context, locals *Locals, namespace *kubernetescorev1.Namespace, provider pulumi.ProviderResource) error {
	policyName := fmt.Sprintf("%s-ingress-policy", locals.NamespaceName)

	ingressRules := kubernetesnetworkingv1.NetworkPolicyIngressRuleArray{}

	// Allow ingress from specified namespaces
	if len(locals.NetworkPolicy.AllowedIngressNamespaces) > 0 {
		for _, allowedNs := range locals.NetworkPolicy.AllowedIngressNamespaces {
			ingressRules = append(ingressRules, &kubernetesnetworkingv1.NetworkPolicyIngressRuleArgs{
				From: kubernetesnetworkingv1.NetworkPolicyPeerArray{
					&kubernetesnetworkingv1.NetworkPolicyPeerArgs{
						NamespaceSelector: &metav1.LabelSelectorArgs{
							MatchLabels: pulumi.StringMap{
								"kubernetes.io/metadata.name": pulumi.String(allowedNs),
							},
						},
					},
				},
			})
		}
	}

	// Allow ingress from within the same namespace
	ingressRules = append(ingressRules, &kubernetesnetworkingv1.NetworkPolicyIngressRuleArgs{
		From: kubernetesnetworkingv1.NetworkPolicyPeerArray{
			&kubernetesnetworkingv1.NetworkPolicyPeerArgs{
				PodSelector: &metav1.LabelSelectorArgs{},
			},
		},
	})

	_, err := kubernetesnetworkingv1.NewNetworkPolicy(
		ctx,
		policyName,
		&kubernetesnetworkingv1.NetworkPolicyArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(policyName),
				Namespace: namespace.Metadata.Name(),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: &kubernetesnetworkingv1.NetworkPolicySpecArgs{
				PodSelector: &metav1.LabelSelectorArgs{},
				PolicyTypes: pulumi.StringArray{
					pulumi.String("Ingress"),
				},
				Ingress: ingressRules,
			},
		},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{namespace}),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create ingress network policy in namespace %s", locals.NamespaceName)
	}

	return nil
}

// createEgressPolicy creates a NetworkPolicy for egress restriction
func createEgressPolicy(ctx *pulumi.Context, locals *Locals, namespace *kubernetescorev1.Namespace, provider pulumi.ProviderResource) error {
	policyName := fmt.Sprintf("%s-egress-policy", locals.NamespaceName)

	egressRules := kubernetesnetworkingv1.NetworkPolicyEgressRuleArray{}

	// Always allow DNS (kube-system namespace)
	egressRules = append(egressRules, &kubernetesnetworkingv1.NetworkPolicyEgressRuleArgs{
		To: kubernetesnetworkingv1.NetworkPolicyPeerArray{
			&kubernetesnetworkingv1.NetworkPolicyPeerArgs{
				NamespaceSelector: &metav1.LabelSelectorArgs{
					MatchLabels: pulumi.StringMap{
						"kubernetes.io/metadata.name": pulumi.String("kube-system"),
					},
				},
			},
		},
		Ports: kubernetesnetworkingv1.NetworkPolicyPortArray{
			&kubernetesnetworkingv1.NetworkPolicyPortArgs{
				Protocol: pulumi.String("UDP"),
				Port:     pulumi.Int(53),
			},
			&kubernetesnetworkingv1.NetworkPolicyPortArgs{
				Protocol: pulumi.String("TCP"),
				Port:     pulumi.Int(53),
			},
		},
	})

	// Allow egress to specified CIDR blocks
	if len(locals.NetworkPolicy.AllowedEgressCIDRs) > 0 {
		for _, cidr := range locals.NetworkPolicy.AllowedEgressCIDRs {
			egressRules = append(egressRules, &kubernetesnetworkingv1.NetworkPolicyEgressRuleArgs{
				To: kubernetesnetworkingv1.NetworkPolicyPeerArray{
					&kubernetesnetworkingv1.NetworkPolicyPeerArgs{
						IpBlock: &kubernetesnetworkingv1.IPBlockArgs{
							Cidr: pulumi.String(cidr),
						},
					},
				},
			})
		}
	}

	// Allow egress within the same namespace
	egressRules = append(egressRules, &kubernetesnetworkingv1.NetworkPolicyEgressRuleArgs{
		To: kubernetesnetworkingv1.NetworkPolicyPeerArray{
			&kubernetesnetworkingv1.NetworkPolicyPeerArgs{
				PodSelector: &metav1.LabelSelectorArgs{},
			},
		},
	})

	_, err := kubernetesnetworkingv1.NewNetworkPolicy(
		ctx,
		policyName,
		&kubernetesnetworkingv1.NetworkPolicyArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(policyName),
				Namespace: namespace.Metadata.Name(),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: &kubernetesnetworkingv1.NetworkPolicySpecArgs{
				PodSelector: &metav1.LabelSelectorArgs{},
				PolicyTypes: pulumi.StringArray{
					pulumi.String("Egress"),
				},
				Egress: egressRules,
			},
		},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{namespace}),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create egress network policy in namespace %s", locals.NamespaceName)
	}

	return nil
}
