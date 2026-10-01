// Package artifactslot is the ONE authored home for injecting a built
// container image into a workload manifest's image slot — the offline
// twin of the hosted control plane's per-kind injectors, driven entirely
// by the catalog's own annotations so the two lanes can never disagree
// about WHERE an image lands.
//
// The slots are declared in the protos (shared/options/options.proto):
// `artifact_image_slot` on a spec's container-carrying field names the
// dotted subpath from each element to the image; `artifact_version_slot`
// marks the field deployment pipelines stamp from the git branch. The
// injection semantics derive from the annotated field's SHAPE and are
// implemented here exactly once:
//
//   - a REPEATED annotated field is blank-fill: elements whose image is
//     empty receive the built reference, authored images (sidecars) are
//     untouched — leaving the image blank IS the authoring contract;
//   - a SINGULAR annotated field is unconditional: the kind models one
//     injectable container;
//   - a subpath resolving to a repo+tag MESSAGE receives the reference
//     split -- repo, tag, and the digest into its own leaf; a string subpath
//     receives it whole.
//
// The hosted control plane keeps its authored injectors; an agreement
// test there holds them to these same annotations. This package never
// prints — callers render their own surface from what Inject reports.
package artifactslot

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	sharedoptions "github.com/plantonhq/planton/shared/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Reference is a parsed container image reference: `host/path:tag`,
// `host/path@digest`, or both -- `host/path:tag@digest`, the form a build
// Planton ran stamps, the tag its readable name and the digest what runs. A
// bare repository parses with an empty tag.
type Reference struct {
	Repo   string
	Tag    string
	Digest string
	// Raw is the reference exactly as given — what string slots receive.
	Raw string
}

// ParseReference splits an image reference: its digest after the '@' first,
// then the tag on the last ':' AFTER the last '/' of what remains (a colon in
// the registry host's port is not a tag separator).
func ParseReference(reference string) Reference {
	name, digest := reference, ""
	if at := strings.IndexByte(reference, '@'); at >= 0 {
		name, digest = reference[:at], reference[at+1:]
	}
	repo, tag := name, ""
	if lastColon := strings.LastIndexByte(name, ':'); lastColon > strings.LastIndexByte(name, '/') {
		repo, tag = name[:lastColon], name[lastColon+1:]
	}
	return Reference{Repo: repo, Tag: tag, Digest: digest, Raw: reference}
}

// versionMaxLength is the version slots' shared grammar ceiling (the
// annotated fields carry the same constraint in their own validation).
const versionMaxLength = 30

var versionIllegal = regexp.MustCompile(`[^a-z0-9]+`)
var versionEdgeHyphens = regexp.MustCompile(`^-+|-+$`)

// SanitizeVersion folds a git branch name into the version slots' grammar
// (lowercase letters, numbers, hyphens; never a hyphen at either edge;
// max 30) — e.g. `feature/PR_42` becomes `feature-pr-42`.
func SanitizeVersion(branch string) string {
	sanitized := strings.ToLower(branch)
	sanitized = versionIllegal.ReplaceAllString(sanitized, "-")
	sanitized = versionEdgeHyphens.ReplaceAllString(sanitized, "")
	if len(sanitized) > versionMaxLength {
		sanitized = strings.TrimRight(sanitized[:versionMaxLength], "-")
	}
	return sanitized
}

// Injection reports one write Inject performed, in the manifest's own
// field-path vocabulary, for the caller's rendering.
type Injection struct {
	// FieldPath is the dotted path written, from the message root
	// (e.g. "spec.containers[0].image", "spec.version").
	FieldPath string
	// Value is what was written (the reference, or the stamped version).
	Value string
}

// Inject writes the image reference into every annotated image slot of the
// manifest (per the semantics above) and, when branch is non-empty, stamps
// every annotated version slot with the sanitized branch. The manifest is
// mutated in place (pass a Clone when the original must survive). A
// manifest whose kind declares NO image slot returns zero injections and
// no error — the caller decides whether that refuses (the offline deploy
// arm does, naming the fact honestly).
func Inject(manifest proto.Message, imageReference, branch string) ([]Injection, error) {
	ref := ParseReference(imageReference)
	root := manifest.ProtoReflect()
	specField := root.Descriptor().Fields().ByName("spec")
	if specField == nil || specField.Message() == nil {
		return nil, nil
	}
	spec := root.Mutable(specField).Message()

	var injections []Injection
	fields := spec.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		opts := fd.Options()
		if opts == nil {
			continue
		}
		if subpath, _ := proto.GetExtension(opts, sharedoptions.E_ArtifactImageSlot).(string); subpath != "" {
			injected, err := injectImageSlot(spec, fd, subpath, ref)
			if err != nil {
				return nil, err
			}
			injections = append(injections, injected...)
		}
		if isVersionSlot, _ := proto.GetExtension(opts, sharedoptions.E_ArtifactVersionSlot).(bool); isVersionSlot && branch != "" {
			version := SanitizeVersion(branch)
			if fd.Kind() != protoreflect.StringKind {
				return nil, errors.Errorf("version slot spec.%s is not a string field", fd.Name())
			}
			spec.Set(fd, protoreflect.ValueOfString(version))
			injections = append(injections, Injection{FieldPath: "spec." + string(fd.Name()), Value: version})
		}
	}
	return injections, nil
}

// injectImageSlot applies the shape-derived semantics to one annotated
// container-carrying field.
func injectImageSlot(spec protoreflect.Message, fd protoreflect.FieldDescriptor, subpath string, ref Reference) ([]Injection, error) {
	basePath := "spec." + string(fd.Name())
	if fd.IsList() {
		// Repeated container list: blank-fill each element at the subpath.
		var injections []Injection
		list := spec.Mutable(fd).List()
		for i := 0; i < list.Len(); i++ {
			element := list.Get(i).Message()
			wrote, err := setImageAtSubpath(element, subpath, ref, true)
			if err != nil {
				return nil, errors.Wrapf(err, "at %s[%d]", basePath, i)
			}
			if wrote {
				injections = append(injections, Injection{
					FieldPath: basePath + "[" + strconv.Itoa(i) + "]." + subpath,
					Value:     ref.Raw,
				})
			}
		}
		return injections, nil
	}
	// Singular container group: unconditional at the subpath.
	if fd.Message() == nil {
		return nil, errors.Errorf("image slot %s is neither a message nor a list of messages", basePath)
	}
	wrote, err := setImageAtSubpath(spec.Mutable(fd).Message(), subpath, ref, false)
	if err != nil {
		return nil, errors.Wrapf(err, "at %s", basePath)
	}
	if !wrote {
		return nil, nil
	}
	return []Injection{{FieldPath: basePath + "." + subpath, Value: ref.Raw}}, nil
}

// setImageAtSubpath walks the dotted subpath from a container message to
// its image and writes the reference — whole into a string leaf, split
// into a repo+tag message leaf. With blankFillOnly, an already-authored
// image is left untouched and the write is reported as skipped.
func setImageAtSubpath(container protoreflect.Message, subpath string, ref Reference, blankFillOnly bool) (bool, error) {
	segments := strings.Split(subpath, ".")
	current := container
	for _, segment := range segments[:len(segments)-1] {
		fd := current.Descriptor().Fields().ByName(protoreflect.Name(segment))
		if fd == nil || fd.Message() == nil {
			return false, errors.Errorf("subpath segment %q does not name a message field on %s", segment, current.Descriptor().FullName())
		}
		current = current.Mutable(fd).Message()
	}
	leafName := segments[len(segments)-1]
	leaf := current.Descriptor().Fields().ByName(protoreflect.Name(leafName))
	if leaf == nil {
		return false, errors.Errorf("subpath leaf %q does not exist on %s", leafName, current.Descriptor().FullName())
	}

	switch {
	case leaf.Kind() == protoreflect.StringKind && !leaf.IsList():
		if blankFillOnly && current.Get(leaf).String() != "" {
			return false, nil
		}
		current.Set(leaf, protoreflect.ValueOfString(ref.Raw))
		return true, nil
	case leaf.Message() != nil:
		// A repo+tag message (the Kubernetes workload image shape): the
		// reference is written SPLIT. Blank-fill checks the repo half.
		image := current.Mutable(leaf).Message()
		repoField := image.Descriptor().Fields().ByName("repo")
		tagField := image.Descriptor().Fields().ByName("tag")
		if repoField == nil || tagField == nil {
			return false, errors.Errorf("image leaf %s carries no repo+tag shape", image.Descriptor().FullName())
		}
		if blankFillOnly && image.Get(repoField).String() != "" {
			return false, nil
		}
		digestField := image.Descriptor().Fields().ByName("digest")
		if digestField == nil && ref.Digest != "" {
			return false, errors.Errorf("image leaf %s has no digest field, so %q cannot be written without losing its digest",
				image.Descriptor().FullName(), ref.Raw)
		}
		image.Set(repoField, protoreflect.ValueOfString(ref.Repo))
		image.Set(tagField, protoreflect.ValueOfString(ref.Tag))
		if digestField != nil {
			// Written even when empty: a new build's reference replaces an authored pin.
			image.Set(digestField, protoreflect.ValueOfString(ref.Digest))
		}
		return true, nil
	default:
		return false, errors.Errorf("image leaf %q on %s is neither a string nor a repo+tag message", leafName, current.Descriptor().FullName())
	}
}
