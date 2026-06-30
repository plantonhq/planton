package backendconfig

import (
	"fmt"
	"strings"

	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumilabels"
	"github.com/plantonhq/planton/pkg/reflection/metadatareflect"
	"google.golang.org/protobuf/proto"
)

// PulumiBackendConfig represents the Pulumi backend configuration
type PulumiBackendConfig struct {
	// StackFqdn is the fully qualified stack name (org/project/stack)
	StackFqdn string
	// Organization is the Pulumi organization name
	Organization string
	// Project is the Pulumi project name
	Project string
	// StackName is the Pulumi stack name
	StackName string
}

// ExtractFromManifest extracts Pulumi backend configuration from manifest labels
// Priority: stack.fqdn > (organization + project + stack.name)
func ExtractFromManifest(manifest proto.Message) (*PulumiBackendConfig, error) {
	labels := metadatareflect.ExtractLabels(manifest)
	if labels == nil {
		return nil, fmt.Errorf("no labels found in manifest")
	}

	config := &PulumiBackendConfig{}

	// First priority: Check for stack.fqdn
	if stackFqdn, ok := labels[pulumilabels.StackFqdnLabelKey]; ok && stackFqdn != "" {
		config.StackFqdn = stackFqdn

		// Parse the FQDN to extract components
		org, project, stack, err := parseStackFqdn(stackFqdn)
		if err != nil {
			return nil, fmt.Errorf("invalid stack.fqdn format: %w", err)
		}

		config.Organization = org
		config.Project = project
		config.StackName = stack

		return config, nil
	}

	// Second priority: Check for individual components
	org, hasOrg := labels[pulumilabels.OrganizationLabelKey]
	project, hasProject := labels[pulumilabels.ProjectLabelKey]
	stack, hasStack := labels[pulumilabels.StackNameLabelKey]

	if !hasOrg || !hasProject || !hasStack {
		return nil, fmt.Errorf("missing required Pulumi backend labels: need either %s or all of (%s, %s, %s)",
			pulumilabels.StackFqdnLabelKey,
			pulumilabels.OrganizationLabelKey,
			pulumilabels.ProjectLabelKey,
			pulumilabels.StackNameLabelKey)
	}

	if org == "" || project == "" || stack == "" {
		return nil, fmt.Errorf("Pulumi backend labels cannot be empty")
	}

	config.Organization = org
	config.Project = project
	config.StackName = stack
	config.StackFqdn = fmt.Sprintf("%s/%s/%s", org, project, stack)

	return config, nil
}

// parseStackFqdn splits "org/project/stack" into components
func parseStackFqdn(fqdn string) (org, project, stack string, err error) {
	parts := strings.Split(fqdn, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("stack FQDN must be in format 'organization/project/stack', got: %s", fqdn)
	}

	org = strings.TrimSpace(parts[0])
	project = strings.TrimSpace(parts[1])
	stack = strings.TrimSpace(parts[2])

	if org == "" || project == "" || stack == "" {
		return "", "", "", fmt.Errorf("stack FQDN components cannot be empty")
	}

	return org, project, stack, nil
}
