package module

import (
	"github.com/pkg/errors"
	kubernetesnatsv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesnats/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the single entry-point consumed by the OpenMCF
// runtime.  It wires together noun-style helpers in a Terraform-like
// top-down order so the flow is easy for DevOps engineers to follow.
//
// Deployment order (per ChatGPT guidance for avoiding race conditions):
// 1. NATS Helm release
// 2. NACK CRDs (explicit step, not via Helm)
// 3. NACK controller Helm release
// 4. Stream/Consumer custom resources
func Resources(ctx *pulumi.Context,
	stackInput *kubernetesnatsv1.KubernetesNatsStackInput) error {

	// ----------------------------- locals ---------------------------------
	locals := initializeLocals(ctx, stackInput)

	// ------------------------- kubernetes provider ------------------------
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Build conditional namespace dependency (Pulumi equivalent of Terraform depends_on).
	// When create_namespace is false, createdNamespace is nil and namespaceDeps is empty.
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ----------------------------- secrets --------------------------------
	if err := tlsSecret(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create TLS secret")
	}

	// ------------------------------ NATS helm -----------------------------
	// Step 1: Deploy NATS server (with JetStream enabled)
	// Returns auth password for NACK controller (if basic_auth enabled)
	natsHelmChart, authPassword, err := helmChart(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to deploy NATS Helm chart")
	}

	// ----------------------------- ingress --------------------------------
	if err := ingress(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create external ingress")
	}

	// ----------------------------- NACK CRDs ------------------------------
	// Step 2: Deploy NACK CRDs (explicit step for better control)
	// CRDs must be registered before the controller can watch them
	nackCrdsResource, err := nackCrds(ctx, locals, kubernetesProvider, natsHelmChart, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to deploy NACK CRDs")
	}

	// --------------------------- NACK controller --------------------------
	// Step 3: Deploy NACK controller (watches CRDs and reconciles to NATS)
	// Pass auth password so NACK can connect to authenticated NATS server
	nackControllerResource, err := nackController(ctx, locals, kubernetesProvider, nackCrdsResource, authPassword, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to deploy NACK controller")
	}

	// -------------------------- streams/consumers -------------------------
	// Step 4: Create Stream/Consumer custom resources
	// These depend on both CRDs (for schema) and controller (for reconciliation)
	if err := streams(ctx, locals, kubernetesProvider, nackControllerResource, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create JetStream streams/consumers")
	}

	return nil
}
