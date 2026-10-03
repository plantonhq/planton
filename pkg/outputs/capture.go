package outputs

import (
	"strings"

	"google.golang.org/protobuf/proto"
)

// CaptureResult carries a stack's outputs captured after a successful apply,
// in every shape a consumer needs: the engine-decoded raw map, the dotted-key
// flattening, the kind's typed Outputs proto, and which outputs the
// kind's schema declares secrets.
//
// Captured values include the real secrets, because resolving a downstream
// reference on this machine needs them, so rendering is the leak boundary:
// every renderer consults IsSensitive before printing a value.
type CaptureResult struct {
	// Raw is the engine-decoded output map: output name -> JSON-decoded value.
	Raw map[string]interface{}

	// Flat is Raw flattened to dotted string keys (see Flatten).
	Flat map[string]string

	// Typed is the kind's Outputs proto populated from Raw, honoring
	// module-shipped transform overrides. Nil when the kind declares no
	// outputs message or the transform was skipped.
	Typed proto.Message

	// Secrets is the kind's schema marks (see SecretOutputs): top-level
	// outputs field name -> whether it is a secret.
	Secrets map[string]bool
}

// IsSensitive reports whether a flat (dotted) key must never render. A dotted
// key inherits its top-level output's mark, and a key the schema does not
// declare (a customized module's extra output) is sensitive too: nothing says
// it is safe to print.
func (r *CaptureResult) IsSensitive(flatKey string) bool {
	if r == nil {
		return true
	}
	root, _, _ := strings.Cut(flatKey, ".")
	secret, declared := r.Secrets[strings.ReplaceAll(root, "-", "_")]
	return secret || !declared
}
