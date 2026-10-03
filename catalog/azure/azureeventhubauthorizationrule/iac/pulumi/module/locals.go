package module

import (
	"regexp"

	"github.com/pkg/errors"
	azureeventhubauthorizationrulev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventhubauthorizationrule/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureEventHubAuthorizationRule *azureeventhubauthorizationrulev1alpha1.AzureEventHubAuthorizationRule

	// Exactly one scope is set (spec-enforced XOR); the parsed parent
	// names drive whichever azurerm resource materializes.
	IsNamespaceScoped bool
	IsHubScoped       bool

	// Parsed from the resolved parent ARM id -- azurerm still addresses
	// Event Hub authorization rules by discrete names, so the module
	// derives them from the spec's single parent reference to keep the
	// spec on the catalog's ARM-id grain.
	ResourceGroupName string
	NamespaceName     string
	EventHubName      string
}

// namespaceIdPattern / hubIdPattern accept the same anchored shapes as the
// Terraform module's regexes -- a malformed id fails loudly on both
// engines instead of creating a rule somewhere unintended.
var (
	namespaceIdPattern = regexp.MustCompile(`/resourceGroups/([^/]+)/providers/Microsoft\.EventHub/namespaces/([^/]+)$`)
	hubIdPattern       = regexp.MustCompile(`/resourceGroups/([^/]+)/providers/Microsoft\.EventHub/namespaces/([^/]+)/eventhubs/([^/]+)$`)
)

func initializeLocals(ctx *pulumi.Context, iacInput *azureeventhubauthorizationrulev1alpha1.AzureEventHubAuthorizationRuleIacInput) (*Locals, error) {
	locals := &Locals{}
	locals.AzureEventHubAuthorizationRule = iacInput.Target
	spec := iacInput.Target.Spec

	// Authorization rules carry no Azure tags: ARM does not support tags
	// on Event Hubs entities, so the platform's identity tags live on the
	// parent namespace.

	namespaceId := spec.NamespaceId.GetValue()
	eventHubId := spec.EventHubId.GetValue()

	switch {
	case namespaceId != "":
		matches := namespaceIdPattern.FindStringSubmatch(namespaceId)
		if matches == nil {
			return nil, errors.Errorf("namespace_id %q is not an Event Hubs namespace ARM id", namespaceId)
		}
		locals.IsNamespaceScoped = true
		locals.ResourceGroupName = matches[1]
		locals.NamespaceName = matches[2]

	case eventHubId != "":
		matches := hubIdPattern.FindStringSubmatch(eventHubId)
		if matches == nil {
			return nil, errors.Errorf("event_hub_id %q is not an event hub ARM id", eventHubId)
		}
		locals.IsHubScoped = true
		locals.ResourceGroupName = matches[1]
		locals.NamespaceName = matches[2]
		locals.EventHubName = matches[3]

	default:
		// Unreachable behind the spec's exactly-one-scope CEL; guards a
		// IaC input that bypassed validation.
		return nil, errors.New("exactly one of namespace_id or event_hub_id must be set")
	}

	return locals, nil
}

// presenceGuardedBool returns the field's value when set and false
// otherwise -- azurerm defaults all three rights to false, and the spec's
// at-least-one-right CEL guarantees a usable rule.
func presenceGuardedBool(field *bool) bool {
	if field == nil {
		return false
	}
	return *field
}
