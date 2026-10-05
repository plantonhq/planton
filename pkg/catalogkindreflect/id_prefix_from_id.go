package catalogkindreflect

import (
	"github.com/pkg/errors"
	"strings"
)

// IdPrefixFromId extracts the kind's id prefix from an InfraComponent ID,
// which is always ic_<kind id prefix>_<ulid>.
func IdPrefixFromId(resourceId string) (string, error) {
	parts := strings.SplitN(resourceId, "_", 3)
	if len(parts) < 3 || parts[0] != "ic" || parts[1] == "" || parts[2] == "" {
		return "", errors.Errorf("%q is not an InfraComponent ID: an InfraComponent ID is ic_<kind id prefix>_<ulid>", resourceId)
	}
	return parts[1], nil
}
