package helmcrds

import (
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// documentSeparator matches a YAML document separator on its OWN line. A bare
// "---" substring split corrupts CRDs: their schema description strings embed
// "---" mid-line (unix permission strings, TODO markers).
var documentSeparator = regexp.MustCompile(`(?m)^---[ \t]*$`)

// splitDocuments splits a multi-document YAML stream into documents, dropping
// empty and comment-only fragments (Helm emits "# Source: ..." headers;
// upstream bundles open with license comments).
func splitDocuments(stream string) []string {
	var documents []string
	for _, doc := range documentSeparator.Split(stream, -1) {
		if isBlankOrComments(doc) {
			continue
		}
		documents = append(documents, strings.TrimSpace(doc)+"\n")
	}
	return documents
}

func isBlankOrComments(doc string) bool {
	for _, line := range strings.Split(doc, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			return false
		}
	}
	return true
}

// objectIdentity is the minimum a document must carry to be a CRD candidate,
// plus the one annotation that decides who keeps it.
type objectIdentity struct {
	Kind     string `json:"kind"`
	Metadata struct {
		Name        string            `json:"name"`
		Annotations map[string]string `json:"annotations"`
	} `json:"metadata"`
}

// crdName returns metadata.name when the document is a CustomResourceDefinition,
// "" otherwise. Not every rendered fragment is an object (NOTES, empty maps).
func crdName(doc string) string {
	var id objectIdentity
	if err := yaml.Unmarshal([]byte(doc), &id); err != nil || id.Kind != "CustomResourceDefinition" {
		return ""
	}
	return id.Metadata.Name
}

// crdDocument is a CRD as the render produced it, with the two facts ownership
// turns on: which Helm surface it came from and whether the chart marked it to
// be kept.
type crdDocument struct {
	CRD
	// onDirectorySurface: the document came from the chart's crds/ directory
	// (Helm's install-once surface, which skip_crds governs).
	onDirectorySurface bool
	// keptByChart: a templated CRD carrying helm.sh/resource-policy: keep, so
	// the chart itself protects it from a cascading uninstall.
	keptByChart bool
}

// filterCRDs keeps only CustomResourceDefinition documents, keyed and sorted
// by metadata.name so callers register resources in a deterministic order.
// A chart that ships the same CRD on both of Helm's surfaces yields it once,
// attributed to the directory. directoryNames is nil for a bundle, whose
// documents have no Helm surface at all.
func filterCRDs(documents []string, directoryNames map[string]bool) ([]crdDocument, error) {
	byName := map[string]crdDocument{}
	for _, doc := range documents {
		var id objectIdentity
		if err := yaml.Unmarshal([]byte(doc), &id); err != nil {
			continue
		}
		if id.Kind != "CustomResourceDefinition" {
			continue
		}
		if id.Metadata.Name == "" {
			return nil, errors.New("a rendered CustomResourceDefinition carries no metadata.name")
		}
		byName[id.Metadata.Name] = crdDocument{
			CRD:                CRD{Name: id.Metadata.Name, YAML: doc},
			onDirectorySurface: directoryNames[id.Metadata.Name],
			keptByChart:        id.Metadata.Annotations[HelmKeepAnnotation] == "keep",
		}
	}
	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names)
	crds := make([]crdDocument, 0, len(names))
	for _, name := range names {
		crds = append(crds, byName[name])
	}
	return crds, nil
}

// stamp writes the source annotations and the selection label onto a CRD. The
// document round-trips through a generic map: key order may change, values
// and types do not (sigs.k8s.io/yaml goes through JSON, which preserves the
// YAML 1.2 core schema types the API server reads).
func stamp(crd CRD, src Source) (CRD, error) {
	var object map[string]interface{}
	if err := yaml.Unmarshal([]byte(crd.YAML), &object); err != nil {
		return CRD{}, errors.Wrapf(err, "CRD %s is not a YAML object", crd.Name)
	}
	metadata, _ := object["metadata"].(map[string]interface{})
	if metadata == nil {
		metadata = map[string]interface{}{}
		object["metadata"] = metadata
	}
	annotations, _ := metadata["annotations"].(map[string]interface{})
	if annotations == nil {
		annotations = map[string]interface{}{}
		metadata["annotations"] = annotations
	}
	labels, _ := metadata["labels"].(map[string]interface{})
	if labels == nil {
		labels = map[string]interface{}{}
		metadata["labels"] = labels
	}
	annotations[AnnotationSourceChart] = src.SourceDescription()
	annotations[AnnotationSourceVersion] = src.Version
	labels[LabelSource] = labelSourceValue(src)

	out, err := yaml.Marshal(object)
	if err != nil {
		return CRD{}, errors.Wrapf(err, "failed to re-encode CRD %s", crd.Name)
	}
	return CRD{Name: crd.Name, YAML: string(out)}, nil
}

// labelSourceValue is what LabelSource carries: the source's own name, never
// anything derived from where it is downloaded (one host serves many
// unrelated projects, and two of them must never read as one owner).
func labelSourceValue(src Source) string {
	return src.Name()
}
