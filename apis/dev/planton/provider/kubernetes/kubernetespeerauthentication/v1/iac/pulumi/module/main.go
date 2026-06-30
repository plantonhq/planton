package module

import (
	"strconv"

	"github.com/pkg/errors"
	kubernetespeerauthenticationv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetespeerauthentication/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	istiosecurityv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/istio/kubernetes/security/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetespeerauthenticationv1.KubernetesPeerAuthenticationStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createPeerAuthentication(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create peer authentication")
	}

	ctx.Export(OpPeerAuthenticationName, pulumi.String(locals.PeerAuthenticationName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return nil
}

// createPeerAuthentication creates the namespaced Istio PeerAuthentication using
// the typed crd2pulumi SDK (istiosecurityv1.NewPeerAuthentication), consistent
// with every other Planton Istio component. The typed approach catches
// field-name and structure errors at compile time rather than at deployment
// time. Each optional upstream block is only attached when present, so unset
// fields fall through to istiod's defaults (inheritance).
func createPeerAuthentication(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesPeerAuthentication.Spec

	// The typed resource's Spec field is a PtrInput satisfied by the Args value
	// itself (not the SpecPtr() wrapper, which marshals to the wrong element
	// type); assigned directly below, mirroring the Gateway component.
	peerAuthSpec := istiosecurityv1.PeerAuthenticationSpecArgs{}

	if mtls := spec.GetMtls(); mtls != nil {
		peerAuthSpec.Mtls = istiosecurityv1.PeerAuthenticationSpecMtlsArgs{
			Mode: pulumi.String(mtls.GetMode()),
		}
	}

	if selector := spec.GetSelector(); selector != nil && len(selector.GetMatchLabels()) > 0 {
		peerAuthSpec.Selector = istiosecurityv1.PeerAuthenticationSpecSelectorArgs{
			MatchLabels: pulumi.ToStringMap(selector.GetMatchLabels()),
		}
	}

	if portLevelMtls := spec.GetPortLevelMtls(); len(portLevelMtls) > 0 {
		// The upstream CRD keys port_level_mtls by port number, but JSON/CRD map
		// keys are strings, so the crd2pulumi SDK models it as a string-keyed map
		// of string maps ({"8080": {"mode": "STRICT"}}). Convert the proto's
		// uint32 keys to their decimal string form.
		portMap := pulumi.StringMapMap{}
		for port, portMtls := range portLevelMtls {
			portMap[strconv.FormatUint(uint64(port), 10)] = pulumi.StringMap{
				"mode": pulumi.String(portMtls.GetMode()),
			}
		}
		peerAuthSpec.PortLevelMtls = portMap
	}

	_, err := istiosecurityv1.NewPeerAuthentication(ctx, locals.PeerAuthenticationName,
		&istiosecurityv1.PeerAuthenticationArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.PeerAuthenticationName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: peerAuthSpec,
		},
		pulumi.Provider(kubeProvider))

	return err
}
