package verify

import (
	"context"
	"fmt"
)

// GenericVerifier is a fallback that always passes (for components without specific verifiers yet).
type GenericVerifier struct {
	Component string
}

func (v *GenericVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] No specific verifier for %s -- skipping resource verification\n", v.Component)
	return nil
}

func (v *GenericVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] No specific verifier for %s -- skipping cleanup verification\n", v.Component)
	return nil
}
