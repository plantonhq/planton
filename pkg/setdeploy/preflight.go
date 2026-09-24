package setdeploy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	pulumibackend "github.com/plantonhq/planton/pkg/iac/pulumi/backendconfig"
	"github.com/plantonhq/planton/pkg/iac/tofu/backendconfig"
	"github.com/plantonhq/planton/pkg/kubernetes/kubecontext"
	"github.com/plantonhq/planton/pkg/manifestgraph"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"google.golang.org/protobuf/proto"
)

// Preflight runs the whole wall over a set of documents and returns the
// preflighted plan. Every check contributes to the report even when an
// earlier check refused — a user in CI fixes the full list in one commit, not
// one problem per run — with the one structural exception that documents
// which failed to LOAD cannot take part in graph checks (their remainder
// still does).
//
// The check order is the report order: load-and-schema, identity, references,
// backend-resolved values, cycles, engine-and-modules, state-backend,
// provider-credentials. Everything verifiable is verified here; what cannot
// be verified is stated as an assumption. After a passing wall, the first IaC
// handoff is the first unverified act.
func Preflight(docs []Doc, flags Flags, probes Probes) *Plan {
	report := &Report{}

	// ── Check 1: load and schema ────────────────────────────────────────────
	loadCheck := Check{Name: "load-and-schema", Title: "Manifests load and validate"}
	var items []manifestgraph.Item
	var loaded []proto.Message
	var sources []string
	for _, doc := range docs {
		msg, err := manifest.LoadManifestBytes(doc.Bytes, doc.Source)
		if err != nil {
			loadCheck.Entries = append(loadCheck.Entries, Entry{
				Severity: SeverityRefusal, Source: doc.Source,
				Message: err.Error(),
			})
			continue
		}
		lines, err := manifest.ViolationLines(msg)
		if err != nil {
			loadCheck.Entries = append(loadCheck.Entries, Entry{
				Severity: SeverityRefusal, Source: doc.Source,
				Message: fmt.Sprintf("validation could not run: %v", err),
			})
			continue
		}
		if len(lines) > 0 {
			for _, line := range lines {
				loadCheck.Entries = append(loadCheck.Entries, Entry{
					Severity: SeverityRefusal, Source: doc.Source, Message: line,
				})
			}
			continue
		}
		items = append(items, manifestgraph.Item{Msg: msg, Source: doc.Source})
		loaded = append(loaded, msg)
		sources = append(sources, doc.Source)
	}
	loadCheck.Verified = append(loadCheck.Verified,
		fmt.Sprintf("%d of %d documents load as known kinds and pass schema validation", len(items), len(docs)))
	report.add(loadCheck)

	// ── Check 2: identity ───────────────────────────────────────────────────
	set, identityFindings := manifestgraph.NewSet(items)
	identityCheck := Check{Name: "identity", Title: "Every resource has one identity"}
	for _, f := range identityFindings {
		identityCheck.Entries = append(identityCheck.Entries, entryFromFinding(f, SeverityRefusal))
	}
	if len(identityFindings) == 0 {
		identityCheck.Verified = append(identityCheck.Verified,
			fmt.Sprintf("%d distinct (kind, name, env) identities", len(set.Nodes)))
	}
	report.add(identityCheck)

	// ── Check 3: references ─────────────────────────────────────────────────
	// The offline consumer policy over the shared finding vocabulary: a
	// reference that NEEDS a value from outside the set refuses (there is no
	// backend to discover the target); a relationship to the outside carries
	// no value need and is a stated assumption; a derived namespace is a
	// placement fact the target's module verifies at apply.
	graph := manifestgraph.BuildGraph(set)
	refCheck := Check{Name: "references", Title: "References resolve inside this set"}
	inSetEdges := 0
	for _, deps := range graph.DependsOn {
		inSetEdges += len(deps)
	}
	for _, f := range graph.Findings {
		switch f.Class {
		case manifestgraph.FindingRefRule:
			refCheck.Entries = append(refCheck.Entries, entryFromFinding(f, SeverityRefusal))
		case manifestgraph.FindingExternalValueFrom, manifestgraph.FindingEnvExternalValueFrom:
			e := entryFromFinding(f, SeverityRefusal)
			e.Message += " — no backend exists here to discover it; add its manifest to this set, or deploy connected"
			refCheck.Entries = append(refCheck.Entries, e)
		case manifestgraph.FindingExternalRelationship:
			refCheck.Entries = append(refCheck.Entries, entryFromFinding(f, SeverityAssumption))
		case manifestgraph.FindingDerivedEdgeDropped:
			// An inference yielded to the author's own order; the set still
			// deploys, and the report says why that edge is not in the order.
			refCheck.Entries = append(refCheck.Entries, entryFromFinding(f, SeverityAssumption))
		default:
			// Resolution-time classes cannot occur at graph build; anything
			// new fails loud rather than passing silent.
			refCheck.Entries = append(refCheck.Entries, entryFromFinding(f, SeverityRefusal))
		}
	}
	for _, derived := range graph.Derived {
		refCheck.Entries = append(refCheck.Entries, Entry{
			Severity: SeverityAssumption,
			Message:  fmt.Sprintf("%s is implied by literal namespace placement and not deployed by this set — its existence is verified by the module at apply", derived),
		})
	}
	if refCheck.refusals() == 0 {
		refCheck.Verified = append(refCheck.Verified,
			fmt.Sprintf("%d cross-resource references resolve to nodes in this set", inSetEdges))
	}
	report.add(refCheck)

	// ── Check 4: backend-resolved values ────────────────────────────────────
	backendRefCheck := Check{Name: "backend-resolved-values", Title: "No values require a Planton backend"}
	for i := range set.Nodes {
		node := &set.Nodes[i]
		for _, use := range CollectBackendRefs(node.Msg) {
			backendRefCheck.Entries = append(backendRefCheck.Entries, Entry{
				Severity: SeverityRefusal, Source: node.Source, FieldPath: use.FieldPath,
				Message: fmt.Sprintf("%s carries a %s reference, which resolves only through a Planton backend — for runtime secrets use provider-native secret references (`planton secret snippet`), or deploy connected", use.FieldPath, strings.TrimSuffix(use.Prefix, "/")),
			})
		}
	}
	if backendRefCheck.refusals() == 0 {
		backendRefCheck.Verified = append(backendRefCheck.Verified, "no $var/ or $secret/ references anywhere in the set")
	}
	report.add(backendRefCheck)

	// ── Check 5: cycles ─────────────────────────────────────────────────────
	order, cycleFinding := graph.TopoOrder()
	cycleCheck := Check{Name: "cycles", Title: "Dependencies form a deployable order"}
	if cycleFinding != nil {
		cycleCheck.Entries = append(cycleCheck.Entries, entryFromFinding(*cycleFinding, SeverityRefusal))
	} else if len(set.Nodes) > 0 {
		cycleCheck.Verified = append(cycleCheck.Verified, "deploy order: "+orderSummary(set, order))
	}
	report.add(cycleCheck)

	// ── Per-node execution facts (feeding checks 6–8) ───────────────────────
	plan := &Plan{Set: set, Graph: graph, Order: order, Report: report}
	engineCheck := Check{Name: "engine-and-modules", Title: "Engines and modules are available"}
	backendCheck := Check{Name: "state-backend", Title: "State backends are configured and reachable"}
	credsCheck := Check{Name: "provider-credentials", Title: "Provider credentials authenticate"}

	plan.Nodes = make([]NodePlan, len(set.Nodes))
	for i := range set.Nodes {
		node := &set.Nodes[i]
		np := NodePlan{Index: i, Identity: node.Identity, Source: node.Source}

		kindName, err := crkreflect.ExtractKindFromProto(node.Msg)
		if err != nil {
			engineCheck.Entries = append(engineCheck.Entries, Entry{
				Severity: SeverityRefusal, Source: node.Source,
				Message: fmt.Sprintf("kind could not be resolved: %v", err),
			})
			plan.Nodes[i] = np
			continue
		}
		np.KindName = kindName
		np.Provider = crkreflect.GetProvider(node.Identity.Kind)

		provType, err := provisioner.ExtractFromManifest(node.Msg)
		if err != nil {
			engineCheck.Entries = append(engineCheck.Entries, Entry{
				Severity: SeverityRefusal, Source: node.Source,
				Message: fmt.Sprintf("invalid planton.dev/provisioner label: %v", err),
			})
			plan.Nodes[i] = np
			continue
		}
		if provType == provisioner.ProvisionerTypeUnspecified {
			// A set deploy is one decision, not N interviews: the engine's
			// default provisioner applies and the report says so.
			provType = provisioner.ProvisionerTypeTofu
			np.ProvisionerDefault = true
		}
		np.Provisioner = provType

		np.KubeContext = flags.KubeContext
		if np.KubeContext == "" {
			np.KubeContext = kubecontext.ExtractFromManifest(node.Msg)
		}

		switch provType {
		case provisioner.ProvisionerTypePulumi:
			resolvePulumiState(&np, node, flags, &backendCheck)
		default:
			resolveTofuState(&np, node, flags, &backendCheck)
		}
		plan.Nodes[i] = np
	}

	// ── Check 6: engine binaries and modules ────────────────────────────────
	runChecks6(plan, flags, probes, &engineCheck)
	report.add(engineCheck)

	// ── Check 7: state backend (collisions, then reachability) ─────────────
	runCheck7(plan, probes, &backendCheck)
	report.add(backendCheck)

	// ── Check 8: provider credentials ───────────────────────────────────────
	runCheck8(plan, probes, &credsCheck)
	report.add(credsCheck)

	return plan
}

// entryFromFinding maps a shared graph finding to a report entry with the
// offline policy's severity. Location details ride along so the renderer can
// group by document.
func entryFromFinding(f manifestgraph.Finding, severity Severity) Entry {
	return Entry{
		Severity:  severity,
		Source:    f.Source,
		FieldPath: f.FieldPath,
		Message:   f.Message,
	}
}

func orderSummary(set *manifestgraph.Set, order []int) string {
	parts := make([]string, 0, len(order))
	for _, idx := range order {
		parts = append(parts, set.Nodes[idx].Identity.String())
	}
	return strings.Join(parts, " -> ")
}

// resolveTofuState merges the node's backend annotations with the set-wide
// flags (flags win, matching the single-manifest lane) and records the
// completeness refusals. The state KEY is per-node by nature and comes only
// from the manifest; its absence names the exact annotation to add.
func resolveTofuState(np *NodePlan, node *manifestgraph.Node, flags Flags, backendCheck *Check) {
	binaryName := "tofu"
	if np.Provisioner == provisioner.ProvisionerTypeTerraform {
		binaryName = "terraform"
	}
	cfg, err := backendconfig.BuildBackendConfig(node.Msg, binaryName, backendconfig.CLIBackendFlags{
		BackendType:     flags.BackendType,
		BackendBucket:   flags.BackendBucket,
		BackendRegion:   flags.BackendRegion,
		BackendEndpoint: flags.BackendEndpoint,
	})
	if err != nil {
		backendCheck.Entries = append(backendCheck.Entries, Entry{
			Severity: SeverityRefusal, Source: node.Source,
			Message: fmt.Sprintf("state backend configuration failed to build: %v", err),
		})
		return
	}
	np.TofuBackend = cfg

	if cfg.BackendType == "" || cfg.BackendType == "local" {
		// Legal and machine-local: execution gives the node an identity-keyed
		// workspace where local state persists across runs. The set-level
		// CI notice is added once in runCheck7.
		return
	}

	validation := backendconfig.Validate(cfg)
	for _, missing := range validation.MissingFields {
		if !missing.Required {
			continue
		}
		// The annotation prefix follows the node's OWN provisioner: a tofu
		// node reads tofu.planton.dev/* annotations, and telling its author
		// to set terraform.planton.dev/* (the validator's generic hint) would
		// be a fix that does not fix.
		fix := fmt.Sprintf("annotation %s.planton.dev/backend.%s", binaryName, missing.Name)
		if missing.AnnotationName == "" {
			fix = missing.FlagName
		}
		backendCheck.Entries = append(backendCheck.Entries, Entry{
			Severity: SeverityRefusal, Source: node.Source,
			Message: fmt.Sprintf("state backend %s is missing %s (%s) — set %s (example: %s)",
				cfg.BackendType, missing.Name, missing.Description, fix, missing.Example),
		})
	}
}

// resolvePulumiState resolves the node's stack identity and backend URL. A
// pulumi node without a stack identity cannot deploy — the refusal names the
// exact annotation.
func resolvePulumiState(np *NodePlan, node *manifestgraph.Node, flags Flags, backendCheck *Check) {
	cfg, err := pulumibackend.ExtractFromManifest(node.Msg)
	if err != nil || cfg.StackFqdn == "" {
		backendCheck.Entries = append(backendCheck.Entries, Entry{
			Severity: SeverityRefusal, Source: node.Source,
			Message: "pulumi node has no stack identity — add annotation pulumi.planton.dev/stack.fqdn: <organization/project/stack>",
		})
		return
	}
	np.PulumiStackFqdn = cfg.StackFqdn
	url, _ := pulumibackend.ResolveBackendURL(node.Msg, flags.PulumiBackendURL)
	np.PulumiBackendURL = url
}

// runChecks6 verifies binaries once per engine and module availability once
// per (kind, provisioner). Module probe assumptions render as warnings: a
// missing published artifact silently degrades to a slow source checkout, and
// loud beats slow-and-silent.
func runChecks6(plan *Plan, flags Flags, probes Probes, check *Check) {
	binariesProbed := map[string]bool{}
	modulesProbed := map[string]bool{}
	for i := range plan.Nodes {
		np := &plan.Nodes[i]
		if np.KindName == "" {
			continue
		}
		var binaryKey string
		switch np.Provisioner {
		case provisioner.ProvisionerTypePulumi:
			binaryKey = "pulumi"
		case provisioner.ProvisionerTypeTerraform:
			binaryKey = "terraform"
		default:
			binaryKey = "tofu"
		}
		if !binariesProbed[binaryKey] {
			binariesProbed[binaryKey] = true
			var res ProbeResult
			if binaryKey == "pulumi" {
				res = probes.PulumiBinary()
			} else {
				res = probes.HclBinary(binaryKey)
			}
			recordProbe(check, res, "", SeverityAssumption)
		}

		moduleKey := np.KindName + "/" + binaryKey
		if !modulesProbed[moduleKey] {
			modulesProbed[moduleKey] = true
			res := probes.ModulePublished(np.KindName, np.Provisioner, flags.ModuleVersion)
			recordProbe(check, res, np.Source, SeverityWarning)
		}

		if np.ProvisionerDefault {
			check.Verified = append(check.Verified,
				fmt.Sprintf("%s: provisioner defaulted to tofu (no planton.dev/provisioner label)", np.Identity))
		}
	}
}

// runCheck7 enforces the state law and probes reachability once per distinct
// backend target. The law: every remote node names its own state key, no two
// nodes share one (a shared key is two resources overwriting each other's
// state — the silent-corruption class), and local state is stated with its
// machine-local consequence.
func runCheck7(plan *Plan, probes Probes, check *Check) {
	type keyOwner struct{ source string }
	tofuKeys := map[string]keyOwner{}
	pulumiStacks := map[string]keyOwner{}
	probed := map[string]bool{}
	localNodes := 0

	for i := range plan.Nodes {
		np := &plan.Nodes[i]
		switch {
		case np.TofuBackend != nil:
			cfg := np.TofuBackend
			if cfg.BackendType == "" || cfg.BackendType == "local" {
				localNodes++
				check.Verified = append(check.Verified,
					fmt.Sprintf("%s: local state in this node's own workspace", np.Identity))
				continue
			}
			if cfg.BackendKey != "" {
				keyID := cfg.BackendType + "://" + cfg.BackendBucket + "/" + cfg.BackendKey
				if prev, dup := tofuKeys[keyID]; dup {
					check.Entries = append(check.Entries, Entry{
						Severity: SeverityRefusal, Source: np.Source,
						Message: fmt.Sprintf("state key %s is also used by %s — two resources writing one state file overwrite each other; give each manifest its own %s.planton.dev/backend.key", keyID, prev.source, backendAnnotationPrefix(np.Provisioner)),
					})
					continue
				}
				tofuKeys[keyID] = keyOwner{source: np.Source}
			}
			probeID := "tofu:" + cfg.BackendType + "/" + cfg.BackendBucket + "/" + cfg.BackendRegion + "/" + cfg.BackendEndpoint
			if !probed[probeID] {
				probed[probeID] = true
				recordProbe(check, probes.TofuBackend(cfg), np.Source, SeverityAssumption)
			}
		case np.PulumiStackFqdn != "":
			if prev, dup := pulumiStacks[np.PulumiStackFqdn]; dup {
				check.Entries = append(check.Entries, Entry{
					Severity: SeverityRefusal, Source: np.Source,
					Message: fmt.Sprintf("pulumi stack %s is also used by %s — two resources sharing one stack overwrite each other; give each manifest its own pulumi.planton.dev/stack.fqdn", np.PulumiStackFqdn, prev.source),
				})
				continue
			}
			pulumiStacks[np.PulumiStackFqdn] = keyOwner{source: np.Source}
			probeID := "pulumi:" + np.PulumiBackendURL
			if !probed[probeID] {
				probed[probeID] = true
				recordProbe(check, probes.PulumiBackend(np.PulumiBackendURL), np.Source, SeverityAssumption)
			}
		}
	}

	if localNodes > 0 {
		check.Entries = append(check.Entries, Entry{
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("%d node(s) keep state on this machine only — re-runs must happen here; name a remote backend (s3/gcs/azurerm) for CI", localNodes),
		})
	}
}

func backendAnnotationPrefix(prov provisioner.ProvisionerType) string {
	if prov == provisioner.ProvisionerTypeTerraform {
		return "terraform"
	}
	return "tofu"
}

// runCheck8 probes each distinct provider's ambient credentials once, and
// each distinct kube context once. Deterministic iteration keeps the report
// stable for golden pinning.
func runCheck8(plan *Plan, probes Probes, check *Check) {
	providers := map[cloudresourcekind.CloudResourceProvider]bool{}
	kubeContexts := map[string]bool{}
	for i := range plan.Nodes {
		np := &plan.Nodes[i]
		if np.KindName == "" {
			continue
		}
		if requiresProviderCredentials(np.Provider) {
			providers[np.Provider] = true
		}
		if np.KubeContext != "" {
			kubeContexts[np.KubeContext] = true
		}
	}

	providerNames := make([]string, 0, len(providers))
	providerByName := map[string]cloudresourcekind.CloudResourceProvider{}
	for p := range providers {
		providerNames = append(providerNames, p.String())
		providerByName[p.String()] = p
	}
	sort.Strings(providerNames)
	for _, name := range providerNames {
		recordProbe(check, probes.ProviderCredentials(providerByName[name]), "", SeverityAssumption)
	}

	contextNames := make([]string, 0, len(kubeContexts))
	for c := range kubeContexts {
		contextNames = append(contextNames, c)
	}
	sort.Strings(contextNames)
	for _, name := range contextNames {
		recordProbe(check, probes.KubeContext(name), "", SeverityAssumption)
	}
}

// requiresProviderCredentials mirrors the provider-detection rule: only the
// unspecified and _test providers deploy without credentials.
func requiresProviderCredentials(p cloudresourcekind.CloudResourceProvider) bool {
	return p != cloudresourcekind.CloudResourceProvider_cloud_resource_provider_unspecified &&
		p != cloudresourcekind.CloudResourceProvider__test
}

// recordProbe folds a probe outcome into a check: Verified joins the pass
// lines, Refused is always a refusal, and Assumed takes the check's chosen
// severity (assumption for genuinely unverifiable facts; warning where silent
// degradation would follow).
func recordProbe(check *Check, res ProbeResult, source string, assumedSeverity Severity) {
	switch {
	case res.Verified != "":
		check.Verified = append(check.Verified, res.Verified)
	case res.Refused != "":
		check.Entries = append(check.Entries, Entry{Severity: SeverityRefusal, Source: source, Message: res.Refused})
	case res.Assumed != "":
		check.Entries = append(check.Entries, Entry{Severity: assumedSeverity, Source: source, Message: res.Assumed})
	}
}
