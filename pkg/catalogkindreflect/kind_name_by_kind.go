package catalogkindreflect

import (
	"github.com/plantonhq/planton/shared/catalogkind"
	log "github.com/sirupsen/logrus"
)

func ExtractKindNameByKind(kind catalogkind.CatalogKind) string {
	kindMeta, err := KindMeta(kind)
	if err != nil {
		log.Errorf("failed to extract kind meta by kind %s", kind.String())
		return ""
	}
	// Fall back to enum string if Name is not set in proto
	if kindMeta.Name == "" {
		return kind.String()
	}
	return kindMeta.Name
}
