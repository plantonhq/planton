package catalogkindreflect

import (
	"github.com/pkg/errors"
	"strings"
)

// IdPrefixFromId extracts the id prefix from an InfraComponent id
// For the new pattern: ic_<prefix>_<ulid>, it extracts <prefix>
// For legacy pattern: <prefix>_<rest>, it extracts <prefix>
func IdPrefixFromId(resourceId string) (string, error) {
	// Check if it follows the new InfraComponent pattern
	if strings.HasPrefix(resourceId, "ic_") {
		// Pattern: ic_<prefix>_<ulid>
		parts := strings.SplitN(resourceId, "_", 3)
		if len(parts) < 3 || parts[1] == "" {
			return "", errors.Errorf("invalid infra component id format: %s", resourceId)
		}
		return parts[1], nil
	}

	// Legacy pattern: <prefix>_<rest>
	parts := strings.SplitN(resourceId, "_", 2)
	if len(parts) < 2 || parts[0] == "" {
		return "", errors.Errorf("failed to extract resource-id prefix from resource id: %s", resourceId)
	}
	return parts[0], nil
}
