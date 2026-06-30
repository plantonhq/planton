package e2ediscover

import (
	"fmt"
	"io"
	"strings"

	componentv1 "github.com/plantonhq/planton/apis/dev/planton/qa/componente2eprofile/v1"
	sharedpb "github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/pkg/e2e/profile"
)

// RenderTable writes a plain-text table to w. No ANSI codes, suitable for piping.
func RenderTable(w io.Writer, result *profile.DiscoverResult) error {
	fmt.Fprintf(w, "%-5s %-40s %-10s %-9s %s\n", "TIER", "COMPONENT", "STATUS", "PROV", "TIMEOUT")
	fmt.Fprintf(w, "%s\n", strings.Repeat("─", 80))

	var currentTier int32
	for _, ce := range result.Components {
		spec := ce.Profile.Spec
		if spec == nil {
			continue
		}

		if spec.Tier != currentTier {
			if currentTier != 0 {
				fmt.Fprintln(w)
			}
			currentTier = spec.Tier
		}

		statusStr := statusName(spec.Status)
		provStr := provisionerShorthand(spec.ValidatedProvisioners)
		timeout := fmt.Sprintf("%dm", spec.TimeoutMinutes)

		fmt.Fprintf(w, "%-5d %-40s %-10s %-9s %s\n",
			spec.Tier, ce.Name, statusStr, provStr, timeout)
	}

	fmt.Fprintln(w)
	counts := profile.CountByStatus(result)
	fmt.Fprintf(w, "Summary: %d GREEN, %d DEFERRED, %d SKIP, %d STUB (%d total)\n",
		counts.Green, counts.Deferred, counts.Skip, counts.Stub, counts.Total)

	return nil
}

func statusName(s componentv1.ComponentE2EProfileSpec_Status) string {
	switch s {
	case componentv1.ComponentE2EProfileSpec_green:
		return "GREEN"
	case componentv1.ComponentE2EProfileSpec_deferred:
		return "DEFERRED"
	case componentv1.ComponentE2EProfileSpec_skip:
		return "SKIP"
	case componentv1.ComponentE2EProfileSpec_stub:
		return "STUB"
	default:
		return "UNKNOWN"
	}
}

func provisionerShorthand(provisioners []sharedpb.IacProvisioner) string {
	var parts []string
	for _, p := range provisioners {
		switch p {
		case sharedpb.IacProvisioner_pulumi:
			parts = append(parts, "P")
		case sharedpb.IacProvisioner_terraform, sharedpb.IacProvisioner_tofu:
			parts = append(parts, "T")
		}
	}
	return strings.Join(parts, " ")
}
