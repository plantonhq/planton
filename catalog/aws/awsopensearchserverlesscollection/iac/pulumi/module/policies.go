package module

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/opensearch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The policy documents rendered here follow AWS's OpenSearch Serverless
// policy schemas. Every rule is scoped to exactly this collection
// ("collection/<name>", "index/<name>/<pattern>") -- the account-wide
// pattern-matching form these policies also support is deliberately outside
// this kind's contract.

// encryptionPolicy renders the collection-scoped encryption security policy
// -- the one policy that must exist BEFORE the collection (AWS rejects
// CreateCollection without a matching encryption policy). Always created:
// AWS-owned key by default, the referenced customer-managed key otherwise.
func encryptionPolicy(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*opensearch.ServerlessSecurityPolicy, error) {
	spec := locals.Spec

	doc := map[string]any{
		"Rules": []any{
			map[string]any{
				"ResourceType": "collection",
				"Resource":     []any{fmt.Sprintf("collection/%s", locals.CollectionName)},
			},
		},
	}
	if spec.Encryption != nil && spec.Encryption.KmsKeyArn.GetValue() != "" {
		// Customer-managed key: AWSOwnedKey false + the key ARN under the
		// document's KmsARN member (the AWS-documented schema).
		doc["AWSOwnedKey"] = false
		doc["KmsARN"] = spec.Encryption.KmsKeyArn.GetValue()
	} else {
		doc["AWSOwnedKey"] = true
	}

	policyJSON, err := json.Marshal(doc)
	if err != nil {
		return nil, errors.Wrap(err, "marshal encryption policy document")
	}

	created, err := opensearch.NewServerlessSecurityPolicy(ctx, "encryption-policy", &opensearch.ServerlessSecurityPolicyArgs{
		// Policy names share the collection's name -- everything the module
		// owns is discoverable by one name. Types are separate namespaces,
		// so the encryption and network policies may share it.
		Name:        pulumi.String(locals.CollectionName),
		Type:        pulumi.String("encryption"),
		Description: pulumi.String(fmt.Sprintf("Encryption for collection %s", locals.CollectionName)),
		Policy:      pulumi.String(string(policyJSON)),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create encryption security policy")
	}
	return created, nil
}

// networkPolicy renders the collection-scoped network security policy. An
// omitted spec.network block still renders the PUBLIC posture (the AWS
// console's easy-create default) -- network "public" is reachability only;
// data access still requires SigV4 auth plus a data-access rule.
func networkPolicy(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	allowFromPublic := true
	var vpcEndpointIds []string
	includeDashboards := true
	if spec.Network != nil {
		if spec.Network.AllowFromPublic != nil {
			allowFromPublic = *spec.Network.AllowFromPublic
		}
		vpcEndpointIds = spec.Network.VpcEndpointIds
		if spec.Network.IncludeDashboards != nil {
			includeDashboards = *spec.Network.IncludeDashboards
		}
	}

	rules := []any{
		map[string]any{
			"ResourceType": "collection",
			"Resource":     []any{fmt.Sprintf("collection/%s", locals.CollectionName)},
		},
	}
	if includeDashboards {
		rules = append(rules, map[string]any{
			"ResourceType": "dashboard",
			"Resource":     []any{fmt.Sprintf("collection/%s", locals.CollectionName)},
		})
	}

	statement := map[string]any{"Rules": rules}
	if allowFromPublic {
		statement["AllowFromPublic"] = true
	} else {
		statement["AllowFromPublic"] = false
		vpces := make([]any, 0, len(vpcEndpointIds))
		for _, id := range vpcEndpointIds {
			vpces = append(vpces, id)
		}
		statement["SourceVPCEs"] = vpces
	}

	// Network policy documents are ARRAYS of statements.
	policyJSON, err := json.Marshal([]any{statement})
	if err != nil {
		return errors.Wrap(err, "marshal network policy document")
	}

	if _, err := opensearch.NewServerlessSecurityPolicy(ctx, "network-policy", &opensearch.ServerlessSecurityPolicyArgs{
		Name:        pulumi.String(locals.CollectionName),
		Type:        pulumi.String("network"),
		Description: pulumi.String(fmt.Sprintf("Network access for collection %s", locals.CollectionName)),
		Policy:      pulumi.String(string(policyJSON)),
	}, pulumi.Provider(provider)); err != nil {
		return errors.Wrap(err, "create network security policy")
	}
	return nil
}

// dataAccessPolicy renders the collection-scoped data access policy -- one
// statement per spec rule. Skipped entirely when the manifest declares no
// rules (the collection is then write-proof and read-proof by design; the
// spec comment says so loudly).
func dataAccessPolicy(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec
	if len(spec.DataAccess) == 0 {
		return nil
	}

	statements := make([]any, 0, len(spec.DataAccess))
	for _, rule := range spec.DataAccess {
		var ruleDocs []any
		if len(rule.CollectionPermissions) > 0 {
			perms := make([]any, 0, len(rule.CollectionPermissions))
			for _, p := range rule.CollectionPermissions {
				perms = append(perms, p)
			}
			ruleDocs = append(ruleDocs, map[string]any{
				"ResourceType": "collection",
				"Resource":     []any{fmt.Sprintf("collection/%s", locals.CollectionName)},
				"Permission":   perms,
			})
		}
		if len(rule.IndexPermissions) > 0 {
			patterns := rule.IndexPatterns
			if len(patterns) == 0 {
				// Default: all indexes of this collection.
				patterns = []string{"*"}
			}
			resources := make([]any, 0, len(patterns))
			for _, p := range patterns {
				resources = append(resources, fmt.Sprintf("index/%s/%s", locals.CollectionName, p))
			}
			perms := make([]any, 0, len(rule.IndexPermissions))
			for _, p := range rule.IndexPermissions {
				perms = append(perms, p)
			}
			ruleDocs = append(ruleDocs, map[string]any{
				"ResourceType": "index",
				"Resource":     resources,
				"Permission":   perms,
			})
		}

		principals := make([]any, 0, len(rule.Principals))
		for _, p := range rule.Principals {
			if p.GetValue() != "" {
				principals = append(principals, p.GetValue())
			}
		}

		statements = append(statements, map[string]any{
			"Rules":     ruleDocs,
			"Principal": principals,
		})
	}

	// Data access policy documents are ARRAYS of statements.
	policyJSON, err := json.Marshal(statements)
	if err != nil {
		return errors.Wrap(err, "marshal data access policy document")
	}

	if _, err := opensearch.NewServerlessAccessPolicy(ctx, "data-access-policy", &opensearch.ServerlessAccessPolicyArgs{
		Name:        pulumi.String(locals.CollectionName),
		Type:        pulumi.String("data"),
		Description: pulumi.String(fmt.Sprintf("Data access for collection %s", locals.CollectionName)),
		Policy:      pulumi.String(string(policyJSON)),
	}, pulumi.Provider(provider)); err != nil {
		return errors.Wrap(err, "create data access policy")
	}
	return nil
}

// lifecyclePolicy renders the collection-scoped index-retention lifecycle
// policy -- one rule per spec entry. Skipped when the manifest declares no
// retention rules (indexes are then retained indefinitely, AWS's default).
func lifecyclePolicy(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec
	if len(spec.RetentionRules) == 0 {
		return nil
	}

	rules := make([]any, 0, len(spec.RetentionRules))
	for _, r := range spec.RetentionRules {
		resources := make([]any, 0, len(r.IndexPatterns))
		for _, p := range r.IndexPatterns {
			resources = append(resources, fmt.Sprintf("index/%s/%s", locals.CollectionName, p))
		}
		rule := map[string]any{
			"ResourceType": "index",
			"Resource":     resources,
		}
		if r.Unlimited {
			rule["NoMinIndexRetention"] = true
		} else {
			rule["MinIndexRetention"] = r.MinIndexRetention
		}
		rules = append(rules, rule)
	}

	policyJSON, err := json.Marshal(map[string]any{"Rules": rules})
	if err != nil {
		return errors.Wrap(err, "marshal lifecycle policy document")
	}

	if _, err := opensearch.NewServerlessLifecyclePolicy(ctx, "retention-policy", &opensearch.ServerlessLifecyclePolicyArgs{
		Name:        pulumi.String(locals.CollectionName),
		Type:        pulumi.String("retention"),
		Description: pulumi.String(fmt.Sprintf("Index retention for collection %s", locals.CollectionName)),
		Policy:      pulumi.String(string(policyJSON)),
	}, pulumi.Provider(provider)); err != nil {
		return errors.Wrap(err, "create lifecycle policy")
	}
	return nil
}
