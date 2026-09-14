package module

import (
	"strconv"

	"github.com/pkg/errors"
	kubernetesnetworkpolicyv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesnetworkpolicy/v1alpha1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	kubernetesnetworkingv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/networking/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createNetworkPolicy creates the networking/v1 NetworkPolicy resource.
//
// policyTypes is ALWAYS sent explicitly (using the resolved/inferred set from
// locals) rather than left to API-server inference — both engines then submit
// byte-identical direction sets for the same manifest, and the deployed object
// never depends on which engine applied it.
func createNetworkPolicy(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource) (*kubernetesnetworkingv1.NetworkPolicy, error) {
	spec := locals.Spec

	policySpecArgs := &kubernetesnetworkingv1.NetworkPolicySpecArgs{
		// An absent pod_selector means "all pods in the namespace" — the empty
		// selector is the correct wire form for that.
		PodSelector: buildLabelSelector(spec.GetPodSelector()),
		PolicyTypes: pulumi.ToStringArray(locals.PolicyTypes),
	}

	if len(spec.GetIngressRules()) > 0 {
		ingressArray := kubernetesnetworkingv1.NetworkPolicyIngressRuleArray{}
		for _, rule := range spec.GetIngressRules() {
			ingressArray = append(ingressArray, &kubernetesnetworkingv1.NetworkPolicyIngressRuleArgs{
				From:  buildPeers(rule.GetFrom()),
				Ports: buildPorts(rule.GetPorts()),
			})
		}
		policySpecArgs.Ingress = ingressArray
	}

	if len(spec.GetEgressRules()) > 0 {
		egressArray := kubernetesnetworkingv1.NetworkPolicyEgressRuleArray{}
		for _, rule := range spec.GetEgressRules() {
			egressArray = append(egressArray, &kubernetesnetworkingv1.NetworkPolicyEgressRuleArgs{
				To:    buildPeers(rule.GetTo()),
				Ports: buildPorts(rule.GetPorts()),
			})
		}
		policySpecArgs.Egress = egressArray
	}

	networkPolicy, err := kubernetesnetworkingv1.NewNetworkPolicy(
		ctx,
		locals.Name,
		&kubernetesnetworkingv1.NetworkPolicyArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:        pulumi.String(locals.Name),
				Namespace:   pulumi.String(locals.Namespace),
				Labels:      pulumi.ToStringMap(locals.Labels),
				Annotations: pulumi.ToStringMap(locals.Annotations),
			},
			Spec: policySpecArgs,
		},
		pulumi.Provider(provider),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create network policy %s/%s", locals.Namespace, locals.Name)
	}

	return networkPolicy, nil
}

// buildPeers converts proto peers into Pulumi peer args. The three peer forms
// (pod selector, namespace selector, IP block) pass through with their exact
// combination semantics — pod+namespace selectors in ONE peer are an AND,
// which is why each proto peer maps to exactly one API peer.
func buildPeers(peers []*kubernetesnetworkpolicyv1alpha1.KubernetesNetworkPolicyPeer) kubernetesnetworkingv1.NetworkPolicyPeerArray {
	var peerArray kubernetesnetworkingv1.NetworkPolicyPeerArray
	for _, p := range peers {
		peerArgs := &kubernetesnetworkingv1.NetworkPolicyPeerArgs{}
		if p.GetPodSelector() != nil {
			peerArgs.PodSelector = buildLabelSelector(p.GetPodSelector())
		}
		if p.GetNamespaceSelector() != nil {
			peerArgs.NamespaceSelector = buildLabelSelector(p.GetNamespaceSelector())
		}
		if p.GetIpBlock() != nil {
			ipBlockArgs := &kubernetesnetworkingv1.IPBlockArgs{
				Cidr: pulumi.String(p.GetIpBlock().GetCidr()),
			}
			if len(p.GetIpBlock().GetExcept()) > 0 {
				ipBlockArgs.Except = pulumi.ToStringArray(p.GetIpBlock().GetExcept())
			}
			peerArgs.IpBlock = ipBlockArgs
		}
		peerArray = append(peerArray, peerArgs)
	}
	return peerArray
}

// buildPorts converts proto port rules into Pulumi port args. An omitted port
// matches all ports for the protocol; a numeric port with end_port expresses a
// contiguous range.
func buildPorts(ports []*kubernetesnetworkpolicyv1alpha1.KubernetesNetworkPolicyPort) kubernetesnetworkingv1.NetworkPolicyPortArray {
	var portArray kubernetesnetworkingv1.NetworkPolicyPortArray
	for _, p := range ports {
		portArgs := &kubernetesnetworkingv1.NetworkPolicyPortArgs{
			Protocol: pulumi.String(resolveProtocol(p.GetProtocol())),
		}
		// port is an IntOrString upstream: a numeric string matches a port
		// number, anything else a named container port on the target pods.
		if p.GetPort() != "" {
			if num, err := strconv.Atoi(p.GetPort()); err == nil {
				portArgs.Port = pulumi.Int(num)
			} else {
				portArgs.Port = pulumi.String(p.GetPort())
			}
		}
		if p.GetEndPort() != 0 {
			portArgs.EndPort = pulumi.Int(int(p.GetEndPort()))
		}
		portArray = append(portArray, portArgs)
	}
	return portArray
}

// buildLabelSelector converts the proto label selector into Pulumi args. A nil
// selector, or one that says match_all, renders as the EMPTY selector — "match
// everything" — which is load-bearing for default-deny policies (a pod_selector
// selecting every pod in the namespace). match_all itself never reaches the
// wire: the empty selector IS its Kubernetes spelling.
func buildLabelSelector(s *kubernetesnetworkpolicyv1alpha1.KubernetesNetworkPolicyLabelSelector) *metav1.LabelSelectorArgs {
	selectorArgs := &metav1.LabelSelectorArgs{}
	if s == nil {
		return selectorArgs
	}
	if len(s.GetMatchLabels()) > 0 {
		selectorArgs.MatchLabels = pulumi.ToStringMap(s.GetMatchLabels())
	}
	if len(s.GetMatchExpressions()) > 0 {
		var exprArray metav1.LabelSelectorRequirementArray
		for _, e := range s.GetMatchExpressions() {
			exprArgs := &metav1.LabelSelectorRequirementArgs{
				Key:      pulumi.String(e.GetKey()),
				Operator: pulumi.String(e.GetOperator()),
			}
			if len(e.GetValues()) > 0 {
				exprArgs.Values = pulumi.ToStringArray(e.GetValues())
			}
			exprArray = append(exprArray, exprArgs)
		}
		selectorArgs.MatchExpressions = exprArray
	}
	return selectorArgs
}

// resolveProtocol maps the protobuf protocol enum to the Kubernetes API string.
func resolveProtocol(p kubernetesnetworkpolicyv1alpha1.KubernetesNetworkPolicyPort_KubernetesNetworkPolicyProtocol) string {
	switch p {
	case kubernetesnetworkpolicyv1alpha1.KubernetesNetworkPolicyPort_UDP:
		return "UDP"
	case kubernetesnetworkpolicyv1alpha1.KubernetesNetworkPolicyPort_SCTP:
		return "SCTP"
	default:
		return "TCP"
	}
}
