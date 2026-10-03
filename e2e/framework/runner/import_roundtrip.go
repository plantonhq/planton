package runner

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/gruntwork-io/terratest/modules/logger"
	tt "github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/iac/importmap"
	"github.com/plantonhq/planton/pkg/protobufyaml"
)

// ImportRoundTripEnvVar opts the import round-trip phase in. It stays opt-in
// because it roughly doubles a kind's E2E wall time (every resource is
// re-imported and re-planned) -- the scheduled matrix enables it per lane.
const ImportRoundTripEnvVar = "PLANTON_E2E_IMPORT_ROUNDTRIP"

// ImportRoundTripSkipAnnotation opts ONE scenario out of the round-trip
// phase; its value MUST state the reason (an empty value does not skip).
// It exists for configs that STRUCTURALLY cannot round-trip to zero
// changes no matter how correct the recipes are -- the first user: a
// scenario carrying write-only ForceNew SECRETS (a container group's
// secure environment variables), where a blind import necessarily lacks
// the secret, the config supplies it, and the only plan is a replace --
// which the oracle rightly refuses, and which ignore_changes must never
// paper over because rotating such a secret is SUPPOSED to replace the
// resource. The kind's recipes still enroll and still prove on its
// secret-free scenarios; the skip is per scenario, reason-carrying, and
// printed into the lane log so the record stays honest.
const ImportRoundTripSkipAnnotation = "planton.dev/e2e-import-roundtrip-skip"

// importRoundTripEnabled gates the phase: opted in, terraform engine (the
// pulumi arm rides the same recipes once its lane lands), and the kind
// actually ships an import map. File presence is the recipes' single
// enrollment signal everywhere -- this gate, the offline conformance guard,
// and the platform's catalog bundler all key off the same import-map.yaml,
// so a map cannot ship while dodging its checks. A scenario-declared,
// reason-carrying skip annotation (above) excludes ONE scenario.
func importRoundTripEnabled(tc *provider.KindTestContext) bool {
	if os.Getenv(ImportRoundTripEnvVar) != "1" ||
		tc.Engine != "terraform" ||
		!importmap.HasCatalogKindImportMap(tc.RepoRoot, tc.Provider, tc.Kind) {
		return false
	}
	if reason, err := ManifestAnnotation(tc.ManifestPath, ImportRoundTripSkipAnnotation); err == nil && reason != "" {
		fmt.Printf("  [import-rt] SKIP (scenario-declared): %s\n", reason)
		return false
	}
	return true
}

// runImportRoundTrip is the machine proof that a kind's import recipes
// are CORRECT, not just well-formed: with the deployed fixture's state set
// aside, every resource is re-imported "blind" -- addresses from the real
// state, IDs derived purely through the recipes (spec/metadata/outputs, no
// human) -- and a plan against the re-imported state must show zero changes.
// A recipe that imports the wrong resource, or misses one, cannot pass.
//
// The freshly-imported state replaces the deployed one for the DESTROY phase
// that follows, which is itself part of the proof: destroy tearing the
// fixture down cleanly shows the re-imported state fully owns the resources.
func runImportRoundTrip(tc *provider.KindTestContext) error {
	opts, ok := tc.TerraformOpts.(*tt.Options)
	if !ok || opts == nil {
		return errors.New("terraform options not initialized (runValidate must run first)")
	}
	workDir := tc.TerraformWorkDir

	// 1. Addresses come from the REAL deployed state -- never authored.
	stateList, err := tt.RunTerraformCommandE(tc.T, opts, "state", "list")
	if err != nil {
		return errors.Wrap(err, "listing deployed state")
	}
	var addresses []string
	for _, line := range strings.Split(stateList, "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "data.") {
			addresses = append(addresses, line)
		}
	}
	if len(addresses) == 0 {
		return errors.New("deployed state lists no resources")
	}

	// 2. Set the deployed state aside -- the re-import starts blind.
	statePath := filepath.Join(workDir, "terraform.tfstate")
	asidePath := statePath + ".pre-roundtrip"
	if err := os.Rename(statePath, asidePath); err != nil {
		return errors.Wrap(err, "setting deployed state aside")
	}
	// A stale backup would let tofu offer state recovery paths that mask a
	// broken import; the round-trip must stand on imports alone.
	_ = os.Remove(statePath + ".backup")

	// 3. Resolve every import ID through the recipes -- no human values.
	catalog, err := importmap.LoadProviderCatalog(tc.RepoRoot, tc.Provider)
	if err != nil {
		return err
	}
	kindMap, err := importmap.LoadCatalogKindImportMap(tc.RepoRoot, tc.Provider, tc.Kind)
	if err != nil {
		return err
	}
	// Both declared tolerance classes merge into one allowlist for the plan
	// oracle: config-only attributes never exist cloud-side, write-normalized
	// attributes read back in a provider-normalized form -- either way an
	// in-place update touching only them is the documented post-import shape,
	// not a wrong import.
	// A declared tolerance is either a top-level attribute name ("force_destroy")
	// or a dotted sub-path ("spec.update_strategy") for the narrower class where
	// a provider's importer fails to read back ONE nested block while the rest
	// of the attribute must stay under the oracle. Dotted entries are collected
	// separately: the changed attribute is then re-compared with only those
	// sub-paths pruned, so an undeclared sibling drift still fails.
	formats := map[string]string{}
	notImportableReasons := map[string]string{}
	toleratedAttributes := map[string]map[string]bool{}
	toleratedSubPaths := map[string][][]string{}
	for _, rt := range catalog.GetSpec().GetResourceTypes() {
		formats[rt.GetTerraformType()] = rt.GetIdFormat()
		if reason := rt.GetNotImportableUpstreamReason(); reason != "" {
			notImportableReasons[rt.GetTerraformType()] = reason
		}
		declared := append(append([]string{}, rt.GetConfigOnlyAttributes()...), rt.GetWriteNormalizedAttributes()...)
		for _, attr := range declared {
			if strings.Contains(attr, ".") {
				toleratedSubPaths[rt.GetTerraformType()] = append(
					toleratedSubPaths[rt.GetTerraformType()], importmap.SplitAttributePath(attr))
				continue
			}
			if toleratedAttributes[rt.GetTerraformType()] == nil {
				toleratedAttributes[rt.GetTerraformType()] = map[string]bool{}
			}
			toleratedAttributes[rt.GetTerraformType()][attr] = true
		}
	}

	// The kind map may additionally declare IMPORT-NORMALIZED sub-paths
	// scoped to ONE of its own resources -- values that cannot round-trip by
	// provider construction (a salted-hash computed attribute re-salts on
	// import) where the first post-adoption apply is functionally a no-op.
	// Collected per logical resource name so the tolerance never leaks to a
	// sibling resource of the same type.
	// Paths split through the shared grammar: dot-separated segments,
	// with bracket-quoted segments for keys that contain literal dots
	// (`data["password.db"]` — a plain dotted split would walk
	// data→password→db and never match the real key).
	normalizedSubPaths := map[string][][]string{}
	for _, nr := range kindMap.GetSpec().GetImportNormalized() {
		for _, sp := range nr.GetSubPaths() {
			normalizedSubPaths[nr.GetTofuResourceName()] = append(
				normalizedSubPaths[nr.GetTofuResourceName()], importmap.SplitAttributePath(sp.GetPath()))
		}
	}

	metadataName, spec, err := loadManifestMetadataAndSpec(tc.Kind, tc.ManifestPath)
	if err != nil {
		return err
	}

	// The round-trip has no user-pasted ARN, but the ACCOUNT-LEVEL ARN parts
	// (account_id, region) are properties of the deployed account itself, not
	// of any one resource -- the platform always knows them from the
	// connection, and here every ARN-shaped output carries the same
	// pair. Recipes for IDs that embed the account id (e.g. DynamoDB
	// contributor insights) stay blind-derivable. Per-resource parts
	// (resource_id/resource_name/arn) are deliberately NOT filled: those
	// genuinely require the user's ARN, and faking them from an unrelated
	// output would let a wrong recipe pass.
	accountArnParts := accountLevelArnParts(tc.FlatOutputs)

	skippedNotImportable := 0
	for _, address := range addresses {
		resourceType, logicalName, instanceKey, parsed := importmap.ParseTofuAddress(address)
		if !parsed {
			return errors.Errorf("address %q is not mappable (module-nested?)", address)
		}
		// Types the upstream provider ships no importer for cannot enter the
		// re-imported state. Their proof is the adopter's contract instead:
		// the oracle below expects the plan to propose re-CREATING exactly
		// these addresses, and the reconcile-apply executes it -- honest only
		// because the catalog admits the reason solely for upsert-convergent
		// create paths.
		if reason, declared := notImportableReasons[resourceType]; declared {
			fmt.Printf("  [import-rt] skipping %s: %s not importable upstream (%s)\n", address, resourceType, reason)
			skippedNotImportable++
			continue
		}
		idFormat, mapped := formats[resourceType]
		if !mapped {
			return errors.Errorf("no id_format for %s (address %s) -- the conformance guard should have caught this", resourceType, address)
		}
		resolved, unresolved := importmap.ResolveValues(kindMap, importmap.Placeholders(idFormat), importmap.ResolveContext{
			MetadataName: metadataName,
			Spec:         spec,
			Outputs:      tc.FlatOutputs,
			AddressKey:   instanceKey,
			LogicalName:  logicalName,
			ArnParts:     accountArnParts,
			// The round-trip holds cluster credentials (the harness
			// exports KUBECONFIG), so it can serve secret-material
			// import IDs the way the human recipe does: by reading the
			// materialized Secret. The value feeds `tofu import` only
			// and is never printed.
			ReadClusterSecret: func(secretName, key string) (string, error) {
				return readClusterSecretKey(tc.FlatOutputs["namespace"], secretName, key)
			},
		})
		// Only required placeholders abort the blind import -- optional
		// ("{name?}") segments are provider-documented as legitimately empty
		// for some variants and render as "".
		if missing := requiredUnresolved(idFormat, unresolved); len(missing) > 0 {
			return errors.Errorf("address %s: values %v not derivable from spec/metadata/outputs -- blind import impossible", address, missing)
		}
		importID, err := importmap.RenderID(idFormat, resolved)
		if err != nil {
			return errors.Wrapf(err, "address %s", address)
		}

		// An ID carrying a cluster-Secret-derived value IS secret
		// material: redact it from the progress line, silence
		// terratest's command echo (it prints full argv), and keep it
		// out of error wraps.
		displayID := importID
		importOpts := opts
		if idContainsSecret(idFormat, importmap.SecretDerivedNames(kindMap)) {
			displayID = "[redacted secret-material import ID]"
			silenced := *opts
			silenced.Logger = logger.Discard
			importOpts = &silenced
		}
		fmt.Printf("  [import-rt] tofu import %s %s\n", address, displayID)
		args := []string{"import", "-input=false"}
		for _, varFile := range opts.VarFiles {
			args = append(args, "-var-file="+varFile)
		}
		args = append(args, address, importID)
		if _, err := tt.RunTerraformCommandE(tc.T, importOpts, args...); err != nil {
			return errors.Wrapf(err, "importing %s as %q", address, displayID)
		}
	}

	// 4. The truth oracle: a plan over the re-imported state must propose no
	// real change. "Real" is precise, not exit-code-blunt: a bare
	// -detailed-exitcode would fail every kind whose config sets a
	// CONFIG-ONLY attribute (aws_s3_bucket.force_destroy is engine delete
	// behavior with no cloud-side existence -- no import can ever carry it,
	// and it legitimately plans as an in-place update after any import) or a
	// WRITE-NORMALIZED one (policy documents read back provider-normalized).
	// So the plan JSON is inspected structurally: creates, destroys, and
	// replaces always fail; in-place updates pass only when every changed
	// attribute is declared in the provider catalog's tolerance lists.
	// Shallow-copy the options: the show-with-struct helper requires a plan
	// file path, and the shared opts must stay pristine for the DESTROY phase.
	planOpts := *opts
	planOpts.PlanFilePath = filepath.Join(workDir, "roundtrip.tfplan")
	planStruct, err := tt.InitAndPlanAndShowWithStructE(tc.T, &planOpts)
	if err != nil {
		return errors.Wrap(err, "plan after re-import")
	}
	tolerated := 0
	recreated := 0
	for address, rc := range planStruct.ResourceChangesMap {
		if rc.Change == nil || rc.Change.Actions.NoOp() || rc.Change.Actions.Read() || rc.Mode == "data" {
			continue
		}
		// A not-importable type's address was deliberately never imported, so
		// the ONLY honest plan for it is a create (the adopter's first apply
		// re-puts the resource). Anything else on such an address -- and a
		// create on any other type -- still fails.
		if _, declared := notImportableReasons[rc.Type]; declared && rc.Change.Actions.Create() {
			fmt.Printf("  [import-rt] tolerating re-create of %s: %s not importable upstream; the reconcile-apply takes ownership\n",
				address, rc.Type)
			recreated++
			continue
		}
		if !rc.Change.Actions.Update() {
			return errors.Errorf(
				"plan after blind re-import proposes %v on %s -- the recipes imported the wrong resource or missed one; deployed state kept at %s",
				rc.Change.Actions, address, asidePath)
		}
		changed := changedTopLevelAttributes(rc.Change.Before, rc.Change.After)
		_, changeLogicalName, _, _ := importmap.ParseTofuAddress(address)
		// Plugin-framework providers mark COMPUTED attributes (id, metadata,
		// status objects) as unknown on every in-place update -- the plan
		// cannot state their post-apply values, so they surface here as
		// "known before, absent after". That is a property of how the
		// framework plans recomputation, never evidence of a wrong import,
		// and no catalog declaration can cover it (computed attributes are
		// not config-only). Prune exactly the attributes the plan itself
		// marks after_unknown as a WHOLE; partially-unknown objects and all
		// known drift stay under the oracle.
		unknownAfter := wholeAttributeAfterUnknown(rc.Change.AfterUnknown)
		for _, attribute := range changed {
			if unknownAfter[attribute] {
				continue
			}
			if toleratedAttributes[rc.Type][attribute] {
				continue
			}
			if changeCoveredBySubPaths(rc.Change.Before, rc.Change.After, attribute, toleratedSubPaths[rc.Type]) {
				continue
			}
			// The kind map's own resource-scoped declarations (see the
			// import_normalized vocabulary): tolerated exactly like catalog
			// sub-paths, but only for THIS logical resource.
			if changeCoveredBySubPaths(rc.Change.Before, rc.Change.After, attribute, normalizedSubPaths[changeLogicalName]) {
				continue
			}
			return errors.Errorf(
				"plan after blind re-import updates %s.%s (changed: %v) -- not a declared config-only/write-normalized attribute of %s nor an import-normalized sub-path of the kind map; deployed state kept at %s",
				address, attribute, changed, rc.Type, asidePath)
		}
		fmt.Printf("  [import-rt] tolerating declared update on %s: %v (config-only/write-normalized in the %s catalog, or import-normalized in the kind map)\n",
			address, changed, tc.Provider)
		tolerated++
	}

	fmt.Printf("  [import-rt] %d resources re-imported blind; plan proposes no real change (%d config-only updates tolerated, %d not-importable-upstream re-creates via reconcile-apply)\n",
		len(addresses)-skippedNotImportable, tolerated, recreated)

	// 5. Reconcile the imported state the way a real adopter does: import →
	// APPLY → operate. An import cannot read CONFIG-ONLY attributes back
	// from the cluster, so the freshly-imported state carries them
	// null/false — and since this state replaces the deployed one for the
	// DESTROY phase, any delete semantics riding a config-only attribute
	// silently degrade (verified live: a kubectl_manifest `wait = true`
	// delete returned instantly from imported state and wedged an
	// operator-finalizer drain, while the same config blocked correctly
	// from applied state). Applying the just-proven plan writes those
	// attributes into state; the oracle above already guarantees the apply
	// changes nothing real.
	if _, err := tt.RunTerraformCommandE(tc.T, &planOpts, "apply", "-input=false", "-auto-approve", planOpts.PlanFilePath); err != nil {
		return errors.Wrap(err, "reconcile-apply after the re-import (import → apply → operate)")
	}

	return nil
}

// accountLevelArnParts extracts the account_id and region from the first
// ARN-shaped output. Within one deployment every ARN carries the same
// account (and, for regional services, the same region), so these two parts
// are deployment-level facts -- the same facts the platform derives from the
// provider connection in the real import flow.
func accountLevelArnParts(flatOutputs map[string]string) map[string]string {
	parts := map[string]string{}
	for _, value := range flatOutputs {
		if !strings.HasPrefix(value, "arn:") {
			continue
		}
		// arn:partition:service:region:account-id:resource
		segments := strings.SplitN(value, ":", 6)
		if len(segments) < 6 {
			continue
		}
		// Global-service ARNs (S3, IAM) leave region -- and S3 even the
		// account -- empty, so each part fills from the first ARN that
		// carries it.
		if parts["account_id"] == "" && segments[4] != "" {
			parts["account_id"] = segments[4]
		}
		if parts["region"] == "" && segments[3] != "" {
			parts["region"] = segments[3]
		}
		if parts["account_id"] != "" && parts["region"] != "" {
			break
		}
	}
	if len(parts) == 0 {
		return nil
	}
	return parts
}

// requiredUnresolved filters the unresolved placeholder names down to the
// ones the id_format marks required -- the only ones that block a blind
// import.
func requiredUnresolved(idFormat string, unresolved []string) []string {
	required := map[string]bool{}
	for _, name := range importmap.RequiredPlaceholders(idFormat) {
		required[name] = true
	}
	var missing []string
	for _, name := range unresolved {
		if required[name] {
			missing = append(missing, name)
		}
	}
	return missing
}

// changeCoveredBySubPaths reports whether the drift on ONE changed top-level
// attribute is fully explained by the declared dotted sub-paths: both sides
// are re-compared with those sub-paths pruned, and only if the remainder is
// identical is the change tolerated. Sub-path segments walk JSON objects and
// apply element-wise through arrays (Terraform plan JSON renders blocks as
// arrays), so "spec.update_strategy" prunes spec[i].update_strategy on both
// sides. Any structural surprise fails closed (returns false).
func changeCoveredBySubPaths(before, after interface{}, attribute string, subPaths [][]string) bool {
	if len(subPaths) == 0 {
		return false
	}
	var relevant [][]string
	for _, p := range subPaths {
		if p[0] == attribute {
			relevant = append(relevant, p[1:])
		}
	}
	if len(relevant) == 0 {
		return false
	}
	beforeMap, beforeOK := before.(map[string]interface{})
	afterMap, afterOK := after.(map[string]interface{})
	if !beforeOK || !afterOK {
		return false
	}
	prunedBefore := pruneSubPaths(beforeMap[attribute], relevant)
	prunedAfter := pruneSubPaths(afterMap[attribute], relevant)
	return reflect.DeepEqual(prunedBefore, prunedAfter)
}

// pruneSubPaths returns a deep copy of value with every declared sub-path
// removed. Arrays are traversed element-wise; scalar leaves along a declared
// path are dropped wherever the path bottoms out.
func pruneSubPaths(value interface{}, paths [][]string) interface{} {
	if len(paths) == 0 {
		return value
	}
	switch typed := value.(type) {
	case map[string]interface{}:
		copied := make(map[string]interface{}, len(typed))
		for key, child := range typed {
			var childPaths [][]string
			drop := false
			for _, p := range paths {
				if len(p) > 0 && p[0] == key {
					if len(p) == 1 {
						drop = true
						break
					}
					childPaths = append(childPaths, p[1:])
				}
			}
			if drop {
				continue
			}
			copied[key] = pruneSubPaths(child, childPaths)
		}
		return copied
	case []interface{}:
		copied := make([]interface{}, len(typed))
		for i, element := range typed {
			copied[i] = pruneSubPaths(element, paths)
		}
		return copied
	default:
		return value
	}
}

// wholeAttributeAfterUnknown returns the top-level attributes the plan marks
// entirely unknown post-apply (after_unknown value is literal true). Only a
// WHOLLY unknown attribute is returned: an object with some unknown members
// (a map value in after_unknown) is not pruned, so known sibling drift
// inside it still fails the oracle.
func wholeAttributeAfterUnknown(afterUnknown interface{}) map[string]bool {
	result := map[string]bool{}
	unknownMap, ok := afterUnknown.(map[string]interface{})
	if !ok {
		return result
	}
	for k, v := range unknownMap {
		if b, isBool := v.(bool); isBool && b {
			result[k] = true
		}
	}
	return result
}

// changedTopLevelAttributes diffs a plan change's before/after objects and
// returns the top-level attribute names whose values differ. Both sides are
// JSON objects for managed resources; anything non-object degrades to a
// sentinel so the caller fails closed rather than tolerating blindly.
func changedTopLevelAttributes(before, after interface{}) []string {
	beforeMap, beforeOK := before.(map[string]interface{})
	afterMap, afterOK := after.(map[string]interface{})
	if !beforeOK || !afterOK {
		return []string{"<non-object change>"}
	}
	keys := map[string]bool{}
	for k := range beforeMap {
		keys[k] = true
	}
	for k := range afterMap {
		keys[k] = true
	}
	var changed []string
	for k := range keys {
		if !reflect.DeepEqual(beforeMap[k], afterMap[k]) {
			changed = append(changed, k)
		}
	}
	sort.Strings(changed)
	return changed
}

// loadManifestMetadataAndSpec loads the scenario manifest into the kind's api
// message and returns the metadata.name plus the spec message the recipe
// derivations read.
func loadManifestMetadataAndSpec(kindDir, manifestPath string) (string, proto.Message, error) {
	kind := catalogkindreflect.KindFromString(kindDir)
	apiMessage, err := catalogkindreflect.NewInstance(kind)
	if err != nil {
		return "", nil, errors.Wrapf(err, "no proto instance for kind %s", kindDir)
	}
	if err := protobufyaml.Load(manifestPath, apiMessage); err != nil {
		return "", nil, errors.Wrapf(err, "loading manifest %s", manifestPath)
	}

	reflected := apiMessage.ProtoReflect()
	metadataField := reflected.Descriptor().Fields().ByName("metadata")
	specField := reflected.Descriptor().Fields().ByName("spec")
	if metadataField == nil || specField == nil {
		return "", nil, errors.Errorf("%s is not a KRM envelope", reflected.Descriptor().FullName())
	}
	nameField := metadataField.Message().Fields().ByName("name")
	metadataName := reflected.Get(metadataField).Message().Get(nameField).String()
	spec := reflected.Get(specField).Message().Interface()
	return metadataName, spec, nil
}

// idContainsSecret reports whether any placeholder of the id_format is a
// secret-derived value (per the kind map's declarations).
func idContainsSecret(idFormat string, secretNames map[string]bool) bool {
	for _, name := range importmap.Placeholders(idFormat) {
		if secretNames[name] {
			return true
		}
	}
	return false
}

// readClusterSecretKey reads one data key of a Kubernetes Secret through
// the ambient KUBECONFIG (the harness exports it) -- the cluster-connected
// implementation behind the from_cluster_secret_key derivation arm. The
// returned value is secret material and must never be logged.
func readClusterSecretKey(namespace, secretName, key string) (string, error) {
	if namespace == "" {
		return "", errors.New("no namespace in the flattened outputs to scope the Secret read")
	}
	out, err := exec.Command("kubectl", "get", "secret", secretName,
		"-n", namespace, "-o", fmt.Sprintf("jsonpath={.data.%s}", key)).Output()
	if err != nil {
		return "", errors.Wrapf(err, "reading secret %s/%s", namespace, secretName)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(out)))
	if err != nil {
		return "", errors.Wrapf(err, "decoding secret %s/%s key %s", namespace, secretName, key)
	}
	return string(decoded), nil
}
