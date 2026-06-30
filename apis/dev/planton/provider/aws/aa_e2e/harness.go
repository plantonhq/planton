// Package aa_e2e implements the E2E provider harness for AWS. Unlike Kubernetes
// (a local kind cluster), AWS is a real cloud account: Setup validates that the
// ambient AWS credential chain can reach the account, and resource verification
// runs through the AWS SDK.
//
// Credentials are intentionally NOT plumbed through the stack input. The E2E
// framework builds every stack input with a nil provider config, so the IaC
// modules resolve credentials from the SDK's ambient chain. That chain is
// populated keylessly -- a short-lived AWS SSO session locally, or a GitHub
// Actions OIDC role in CI -- so no static secret is ever stored on disk or in CI.
package aa_e2e

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/provider/aws/aa_e2e/verify"
	"github.com/plantonhq/planton/e2e/framework/provider"
)

// defaultRegion is used for credential validation and as the fallback when a
// scenario manifest does not pin spec.region. It matches the region of the
// preconfigured local AWS SSO profile; override with E2E_AWS_REGION.
const defaultRegion = "us-west-2"

// Harness manages the AWS E2E test lifecycle.
type Harness struct {
	cfg aws.Config

	// mu guards deployed, written by VerifyDeployed and read by VerifyDestroyed.
	mu       sync.Mutex
	deployed map[string]deployedResource
}

// deployedResource records what VerifyDeployed observed so VerifyDestroyed can
// re-probe the same resource in the same region.
type deployedResource struct {
	id     string
	region string
}

// NewHarness creates an AWS test harness. Credentials come from the ambient chain
// (see the package doc); none are passed here.
func NewHarness() *Harness {
	return &Harness{deployed: make(map[string]deployedResource)}
}

// Setup loads AWS config from the ambient credential chain and confirms it
// resolves to a usable identity via sts:GetCallerIdentity (zero IAM permission,
// side-effect-free). It also exports AWS_REGION so the IaC modules and the SDK
// agree on a default region when a scenario omits spec.region.
func (h *Harness) Setup(ctx context.Context) error {
	region := firstNonEmpty(os.Getenv("E2E_AWS_REGION"), os.Getenv("AWS_REGION"), defaultRegion)
	if err := os.Setenv("AWS_REGION", region); err != nil {
		return errors.Wrap(err, "failed to export AWS_REGION")
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return errors.Wrap(err, "failed to load AWS config from the ambient credential chain "+
			"(locally: `aws sso login --sso-session planton-aws-e2e`; in CI: assume the OIDC role)")
	}

	ident, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return errors.Wrap(err, "AWS credential validation failed (sts:GetCallerIdentity); "+
			"no usable credentials in the ambient chain")
	}

	fmt.Printf("  [aws] authenticated as %s (account %s, region %s)\n",
		aws.ToString(ident.Arn), aws.ToString(ident.Account), region)

	h.cfg = cfg
	return nil
}

// Teardown is a no-op. Each scenario destroys its own resources in the DESTROY
// phase and confirms removal in VERIFY-CLN; cross-run orphans are reclaimed by
// the scheduled cloud-nuke janitor (CI), not here.
func (h *Harness) Teardown(ctx context.Context) error {
	return nil
}

// VerifyDeployed confirms the component's resource exists via its registered
// verifier, using the resource id and region carried in the stack outputs.
func (h *Harness) VerifyDeployed(ctx context.Context, component string, outputs map[string]interface{}) error {
	v, err := verify.GetVerifier(component)
	if err != nil {
		return err
	}

	id := stringOutput(outputs, v.IDOutputKey())
	if id == "" {
		return errors.Errorf("no %q in outputs for %s -- cannot verify", v.IDOutputKey(), component)
	}
	region := stringOutput(outputs, "region")
	if region == "" {
		region = h.cfg.Region
	}

	h.mu.Lock()
	h.deployed[componentKey(ctx, component)] = deployedResource{id: id, region: region}
	h.mu.Unlock()

	return v.VerifyExists(ctx, h.cfg, id, region)
}

// VerifyDestroyed confirms the previously deployed resource no longer exists.
func (h *Harness) VerifyDestroyed(ctx context.Context, component string) error {
	v, err := verify.GetVerifier(component)
	if err != nil {
		return err
	}

	h.mu.Lock()
	res := h.deployed[componentKey(ctx, component)]
	h.mu.Unlock()

	if res.id == "" {
		return errors.Errorf("no stored resource id for %s -- VerifyDeployed may not have run", component)
	}
	return v.VerifyAbsent(ctx, h.cfg, res.id, res.region)
}

// stringOutput reads a string-valued stack output, tolerating non-string scalars.
func stringOutput(outputs map[string]interface{}, key string) string {
	if outputs == nil {
		return ""
	}
	if v, ok := outputs[key]; ok {
		if s, isStr := v.(string); isStr {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// componentKey combines the manifest path (from context) with the component name
// so concurrent scenarios of the same component type do not collide in the map.
func componentKey(ctx context.Context, component string) string {
	if mp, ok := ctx.Value(provider.ManifestPathKey{}).(string); ok && mp != "" {
		return mp + "::" + component
	}
	return component
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
