package importmap

import (
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	kindv1 "github.com/plantonhq/planton/iac/catalogkindimportmap/v1"
)

// ResolveContext carries everything a kind's derivations may draw from
// when resolving import-ID placeholder values for ONE enumerated address.
type ResolveContext struct {
	// The Planton resource's metadata.name (from_metadata_name).
	MetadataName string
	// The kind's spec message (from_spec_field). May be nil.
	Spec proto.Message
	// Flattened outputs, key -> value (from_output). May be nil.
	Outputs map[string]string
	// The enumerated address's instance key, e.g. "archive" from
	// `...intelligent_tiering_configuration.this["archive"]`
	// (from_address_key). Empty for non-repeated resources.
	AddressKey string
	// The enumerated address's Terraform logical resource name (the `istiod`
	// in `helm_release.istiod`) -- selects tofu_resource_name-scoped value
	// declarations when a module has several resources of one type whose
	// placeholders carry different values. Empty disables scoped selection.
	LogicalName string
	// Parsed parts of a user-pasted ARN, keyed "resource_id" |
	// "resource_name" | "account_id" | "region" | "arn" (the full ARN as
	// pasted) for from_arn_part. May be nil or partial: deterministic
	// contexts without a pasted ARN (the E2E round-trip) still populate the
	// ACCOUNT-LEVEL parts ("account_id", "region") from facts they hold,
	// because those are properties of the connected account, not of any one
	// resource -- but never the per-resource parts.
	ArnParts map[string]string
	// Reads one data key of a named Kubernetes Secret in the deployed
	// resource's namespace (from_cluster_secret_key). Bound ONLY by
	// cluster-connected contexts -- the E2E round-trip binds it through
	// the cluster credentials it already holds; disconnected contexts
	// leave it nil, so the arm resolves empty and the value falls back
	// to being asked with where_to_find. The returned value is secret
	// material: callers use it for the import operation and never log
	// or persist it. May be nil.
	ReadClusterSecret func(secretName, key string) (string, error)
}

// ResolveValues resolves the named placeholders through the kind map's
// ordered derivations. The first derivation that yields a non-empty value
// wins. A declaration scoped to the address's logical resource name
// (tofu_resource_name) wins over the unscoped declaration of the same
// placeholder; scoped declarations for OTHER resources are never consulted.
// Returns the resolved values and the names nothing could derive -- the
// caller decides whether unresolved names are user inputs (the wizard) or a
// failure (the round-trip proof).
func ResolveValues(
	m *kindv1.CatalogKindImportMap,
	names []string,
	rctx ResolveContext,
) (resolved map[string]string, unresolved []string) {
	unscopedByName := make(map[string]*kindv1.ImportValue)
	scopedByName := make(map[string]*kindv1.ImportValue)
	for _, v := range m.GetSpec().GetValues() {
		switch v.GetTofuResourceName() {
		case "":
			unscopedByName[v.GetName()] = v
		case rctx.LogicalName:
			scopedByName[v.GetName()] = v
		}
	}

	resolved = make(map[string]string)
	for _, name := range names {
		importValue, declared := scopedByName[name]
		if !declared {
			importValue, declared = unscopedByName[name]
		}
		value := ""
		if declared {
			for _, derivation := range importValue.GetDerivations() {
				value = resolveDerivation(derivation, rctx)
				if value != "" {
					break
				}
			}
		}
		if value == "" {
			unresolved = append(unresolved, name)
			continue
		}
		resolved[name] = value
	}
	return resolved, unresolved
}

// SecretDerivedNames returns the placeholder names whose declared
// derivations include a cluster-Secret read — values that are secret
// material when resolved. Callers use this to redact import IDs from
// logs and command traces (the values exist to feed the import
// operation, never to be displayed).
func SecretDerivedNames(m *kindv1.CatalogKindImportMap) map[string]bool {
	names := make(map[string]bool)
	for _, v := range m.GetSpec().GetValues() {
		for _, d := range v.GetDerivations() {
			if _, ok := d.GetSource().(*kindv1.ImportValueDerivation_FromClusterSecretKey); ok {
				names[v.GetName()] = true
			}
		}
	}
	return names
}

func resolveDerivation(d *kindv1.ImportValueDerivation, rctx ResolveContext) string {
	switch source := d.GetSource().(type) {
	case *kindv1.ImportValueDerivation_FromMetadataName:
		if source.FromMetadataName {
			return rctx.MetadataName
		}
	case *kindv1.ImportValueDerivation_FromSpecField:
		return specFieldValue(rctx.Spec, source.FromSpecField)
	case *kindv1.ImportValueDerivation_FromOutput:
		return rctx.Outputs[source.FromOutput]
	case *kindv1.ImportValueDerivation_FromArnPart:
		return rctx.ArnParts[source.FromArnPart]
	case *kindv1.ImportValueDerivation_FromAddressKey:
		if source.FromAddressKey {
			return rctx.AddressKey
		}
	case *kindv1.ImportValueDerivation_FromMetadataNameSuffix:
		if rctx.MetadataName != "" && source.FromMetadataNameSuffix != "" {
			return rctx.MetadataName + source.FromMetadataNameSuffix
		}
	case *kindv1.ImportValueDerivation_FromMetadataNamePrefix:
		if rctx.MetadataName != "" && source.FromMetadataNamePrefix != "" {
			return source.FromMetadataNamePrefix + rctx.MetadataName
		}
	case *kindv1.ImportValueDerivation_Literal:
		return source.Literal
	case *kindv1.ImportValueDerivation_FromClusterSecretKey:
		// Only cluster-connected contexts bind the reader; anywhere else
		// the arm resolves empty and the caller's ask-the-user fallback
		// (where_to_find) carries the recipe. A read failure is treated
		// the same as absence -- the Secret may legitimately not exist
		// for variants that reference user-provided credentials.
		if rctx.ReadClusterSecret == nil || rctx.MetadataName == "" {
			return ""
		}
		ref := source.FromClusterSecretKey
		// The Secret key is either declared statically or taken from the
		// address's own instance key (keyed resource collections whose
		// per-instance credentials live under manifest-driven keys --
		// e.g. one random_password per declared user, keyed by username).
		key := ref.GetKey()
		if ref.GetKeyFromAddressKey() && rctx.AddressKey != "" {
			key = rctx.AddressKey
		}
		if key == "" {
			return ""
		}
		value, err := rctx.ReadClusterSecret(rctx.MetadataName+ref.GetNameSuffix(), key)
		if err != nil {
			return ""
		}
		return value
	case *kindv1.ImportValueDerivation_FromOutputKeyedByAddress:
		// The module exports a map keyed by the SAME key as the resource's
		// for_each instances, so the enumerated address selects the entry.
		// The lookup composes the flattened dot-path key directly (map keys
		// may themselves contain dots -- a CIDR like "10.1.0.0/16" -- which
		// is safe because the flattened map is consulted by exact string,
		// never re-parsed). Addresses without an instance key resolve empty.
		if rctx.AddressKey == "" {
			return ""
		}
		return rctx.Outputs[source.FromOutputKeyedByAddress+"."+rctx.AddressKey]
	case *kindv1.ImportValueDerivation_FromAddressKeySegment:
		// The delimiter is fixed to "//" — the kubectl composed-ID form
		// this arm exists for (see the proto comment). An out-of-range
		// index resolves to "" so an optional trailing segment (the
		// cluster-scoped 3-part key) drops out of the ID's bracketed group.
		if rctx.AddressKey == "" {
			return ""
		}
		segments := strings.Split(rctx.AddressKey, "//")
		index := int(source.FromAddressKeySegment)
		if index < 0 || index >= len(segments) {
			return ""
		}
		return segments[index]
	}
	return ""
}

// specFieldValue walks a dot path of proto field names through the spec
// message and returns the scalar leaf as a string. Any miss along the path
// (unknown field, nil message, non-scalar leaf, repeated segment) resolves to
// "" -- derivations are best-effort by contract; validation of the PATH
// itself is the conformance guard's job.
func specFieldValue(spec proto.Message, dotPath string) string {
	if spec == nil || dotPath == "" {
		return ""
	}
	current := spec.ProtoReflect()
	segments := strings.Split(dotPath, ".")
	for i, segment := range segments {
		field := current.Descriptor().Fields().ByName(protoreflect.Name(segment))
		if field == nil || field.IsList() || field.IsMap() {
			return ""
		}
		value := current.Get(field)
		if i == len(segments)-1 {
			if field.Kind() == protoreflect.MessageKind || field.Kind() == protoreflect.GroupKind {
				return ""
			}
			return fmt.Sprintf("%v", value.Interface())
		}
		if field.Kind() != protoreflect.MessageKind {
			return ""
		}
		current = value.Message()
	}
	return ""
}

// tofuAddressPattern splits a Terraform/OpenTofu resource address into
// type, logical name, and optional instance key:
//
//	aws_s3_bucket.this                     -> (aws_s3_bucket, this, "")
//	aws_s3_bucket_intelligent_tiering_configuration.this["archive"]
//	                                       -> (..., this, archive)
//	aws_vpc_ipv4_cidr_block_association.secondary["10.1.0.0/16"]
//	                                       -> (..., secondary, 10.1.0.0/16)
//	count instances use [0]                -> key "" (see below)
var tofuAddressPattern = regexp.MustCompile(`^([a-z0-9_]+)\.([A-Za-z0-9_-]+)(?:\[(?:"([^\]"]*)"|\d+)\])?$`)

// ParseTofuAddress decomposes an enumerated OpenTofu address. Returns ok=false
// for shapes this package does not map (module-nested addresses, data sources).
//
// Only quoted for_each keys surface as the instance key: modules key
// for_each instances by exactly the discriminator an import ID needs (an
// alias name, a policy ARN), which is what from_address_key exists for. A
// numeric count index is positional, never identity -- letting it leak into
// an import ID would import a resource literally named "0" -- so count
// instances parse with an empty key (their conditional resources derive
// their ID from elsewhere, or the segment is optional).
func ParseTofuAddress(address string) (resourceType, logicalName, instanceKey string, ok bool) {
	if strings.HasPrefix(address, "module.") || strings.HasPrefix(address, "data.") {
		return "", "", "", false
	}
	m := tofuAddressPattern.FindStringSubmatch(address)
	if m == nil {
		return "", "", "", false
	}
	return m[1], m[2], m[3], true
}
